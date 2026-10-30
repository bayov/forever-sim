package main

import (
	"math"
	"slices"

	"github.com/wowsims/classic/sim/core/proto"
)

// Stat weights: a weight for each stat and pseudo stat, as UnitStats has them.
type weights struct {
	stats, pseudo []float64
}

func weightsFromProto(u *proto.UnitStats) weights {
	return weights{stats: u.GetStats(), pseudo: u.GetPseudoStats()}
}

func (w *weights) stat(s proto.Stat) float64 { return at(w.stats, int(s)) }

func (w *weights) pseudoStat(s proto.PseudoStat) float64 { return at(w.pseudo, int(s)) }

func at(values []float64, i int) float64 {
	if i < len(values) {
		return values[i]
	}
	return 0
}

// The leveling gear plan, the same as ui/core/leveling_gear.ts (docs/leveling_gear.md): what we
// wear at each level from 1 to 60 when we gear up as we level. It plays out the same random
// history from the same seed, so the UI shows the gear we sim. Any change here goes in
// leveling_gear.ts too, and gear_test.go checks the two give the same plan.

var findsPerLevel = []struct {
	itemType proto.ItemType
	rate     float64
}{
	// In the order of the item types, as the TS object's keys go.
	{proto.ItemType_ItemTypeHead, 0.25},
	{proto.ItemType_ItemTypeNeck, 0.15},
	{proto.ItemType_ItemTypeShoulder, 0.25},
	{proto.ItemType_ItemTypeBack, 0.25},
	{proto.ItemType_ItemTypeChest, 0.3},
	{proto.ItemType_ItemTypeWrist, 0.3},
	{proto.ItemType_ItemTypeHands, 0.3},
	{proto.ItemType_ItemTypeWaist, 0.3},
	{proto.ItemType_ItemTypeLegs, 0.3},
	{proto.ItemType_ItemTypeFeet, 0.3},
	{proto.ItemType_ItemTypeFinger, 0.3},
	{proto.ItemType_ItemTypeTrinket, 0.06},
	{proto.ItemType_ItemTypeWeapon, 0.4},
	{proto.ItemType_ItemTypeRanged, 0.1},
}

var qualityMix = []struct{ level, white, green, blue float64 }{
	{1, 0.85, 0.15, 0},
	{8, 0.6, 0.4, 0},
	{12, 0.4, 0.58, 0.02},
	{18, 0.15, 0.8, 0.05},
	{25, 0.08, 0.82, 0.1},
	{35, 0.05, 0.81, 0.14},
	{45, 0.04, 0.78, 0.18},
	{60, 0.03, 0.72, 0.25},
}

const (
	questShare         = 0.6
	questChoices       = 3
	dungeonRunChance   = 0.8
	dungeonWinChance   = 0.75
	ilvlBelow          = 3
	ilvlAbove          = 5
	gearMaxLevel       = 60
	plainMaxIlvl       = 65
	plainIDBase        = 970_000_000
	pickEpsilon        = 1e-9
	emptySlotPreferred = 0.01
)

var dungeonRunLevels = []int32{1}

var levelingDungeons = []struct {
	zoneID  int32
	faction proto.Faction
}{
	{2437, proto.Faction_Horde},     // Ragefire Chasm
	{16611, proto.Faction_Horde},    // Ruins of Lordaeron
	{16919, proto.Faction_Alliance}, // The Hall of Thanes
	{1581, proto.Faction_Alliance},  // The Deadmines
	{718, proto.Faction_Unknown},    // Wailing Caverns
	{209, proto.Faction_Unknown},    // Shadowfang Keep
	{717, proto.Faction_Alliance},   // The Stockade
	{719, proto.Faction_Unknown},    // Blackfathom Deeps
	{721, proto.Faction_Unknown},    // Gnomeregan
	{491, proto.Faction_Unknown},    // Razorfen Kraul
	{796, proto.Faction_Unknown},    // Scarlet Monastery
	{722, proto.Faction_Unknown},    // Razorfen Downs
	{1337, proto.Faction_Unknown},   // Uldaman
	{1176, proto.Faction_Unknown},   // Zul'Farrak
	{2100, proto.Faction_Unknown},   // Maraudon
	{1477, proto.Faction_Unknown},   // The Temple of Atal'Hakkar
	{1584, proto.Faction_Unknown},   // Blackrock Depths
	{1583, proto.Faction_Unknown},   // Blackrock Spire
	{2557, proto.Faction_Unknown},   // Dire Maul
	{2017, proto.Faction_Unknown},   // Stratholme
	{2057, proto.Faction_Unknown},   // Scholomance
}

// The dungeons above and the raids and battlegrounds, whose loot drops only there.
var instanceZones = func() map[int32]bool {
	zones := map[int32]bool{15475: true, 1977: true, 2159: true, 2677: true, 2717: true, 2597: true, 3277: true, 3358: true}
	for _, d := range levelingDungeons {
		zones[d.zoneID] = true
	}
	return zones
}()

// The slots the plan fills, in the ItemSlot order the equipment has.
var gearSlots = []proto.ItemSlot{
	proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotNeck, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotBack,
	proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotWrist, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotWaist,
	proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet, proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2,
	proto.ItemSlot_ItemSlotTrinket1, proto.ItemSlot_ItemSlotTrinket2, proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand,
	proto.ItemSlot_ItemSlotRanged,
}

var slotType = map[proto.ItemSlot]proto.ItemType{
	proto.ItemSlot_ItemSlotHead:     proto.ItemType_ItemTypeHead,
	proto.ItemSlot_ItemSlotNeck:     proto.ItemType_ItemTypeNeck,
	proto.ItemSlot_ItemSlotShoulder: proto.ItemType_ItemTypeShoulder,
	proto.ItemSlot_ItemSlotBack:     proto.ItemType_ItemTypeBack,
	proto.ItemSlot_ItemSlotChest:    proto.ItemType_ItemTypeChest,
	proto.ItemSlot_ItemSlotWrist:    proto.ItemType_ItemTypeWrist,
	proto.ItemSlot_ItemSlotHands:    proto.ItemType_ItemTypeHands,
	proto.ItemSlot_ItemSlotWaist:    proto.ItemType_ItemTypeWaist,
	proto.ItemSlot_ItemSlotLegs:     proto.ItemType_ItemTypeLegs,
	proto.ItemSlot_ItemSlotFeet:     proto.ItemType_ItemTypeFeet,
	proto.ItemSlot_ItemSlotFinger1:  proto.ItemType_ItemTypeFinger,
	proto.ItemSlot_ItemSlotFinger2:  proto.ItemType_ItemTypeFinger,
	proto.ItemSlot_ItemSlotTrinket1: proto.ItemType_ItemTypeTrinket,
	proto.ItemSlot_ItemSlotTrinket2: proto.ItemType_ItemTypeTrinket,
	proto.ItemSlot_ItemSlotMainHand: proto.ItemType_ItemTypeWeapon,
	proto.ItemSlot_ItemSlotOffHand:  proto.ItemType_ItemTypeWeapon,
	proto.ItemSlot_ItemSlotRanged:   proto.ItemType_ItemTypeRanged,
}

// A piece we wear: the item and the random suffix it rolled.
type gearPiece struct {
	item   *proto.UIItem
	suffix *proto.ItemRandomSuffix
	source string
	since  int32
}

// What we wear at one level.
type gearLevel struct {
	level   int32
	gear    map[proto.ItemSlot]gearPiece
	changed []proto.ItemSlot
}

type gearPlanInputs struct {
	items       []*proto.UIItem
	suffixes    map[int32]*proto.ItemRandomSuffix
	itemsByID   map[int32]*proto.UIItem
	playerClass proto.Class
	rules       *proto.LevelingGearClassRules
	faction     proto.Faction
	seed        int32
	pins        []*proto.LevelingGearPin
	// The spec's leveling route, or nil when the plan makes up the quest rewards and dungeon
	// runs.
	route     *proto.LevelingRoute
	weightsAt func(level int32) weights
}

type gearPlan struct {
	levels []gearLevel
	// The plain white items the plan made up, by id, for the sim's database.
	madeUp map[int32]*proto.UIItem
}

// planLevelingGear plays out one leveling journey and says what we wear at each level.
func planLevelingGear(in *gearPlanInputs) *gearPlan {
	pool := newItemPool(in)
	gear := map[proto.ItemSlot]gearPiece{}
	var levels []gearLevel
	runs := planDungeonRuns(in, pool)

	// We start in white chest, legs and boots, with a white weapon.
	for _, slot := range []proto.ItemSlot{proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet, proto.ItemSlot_ItemSlotMainHand} {
		if item := pool.plainItem(slotType[slot], 1, 1, func() float64 { return 0 }); item != nil {
			gear[slot] = gearPiece{item: item, source: "start", since: 1}
		}
	}

	for level := int32(1); level <= gearMaxLevel; level++ {
		weights := in.weightsAt(level)
		var changed []proto.ItemSlot
		gainOf := func(p gearPiece) (proto.ItemSlot, bool, float64) { return bestSlotFor(p, gear, &weights, in.rules) }
		offer := func(p gearPiece) {
			slot, ok, _ := gainOf(p)
			if !ok {
				return
			}
			gear[slot] = p
			if slot == proto.ItemSlot_ItemSlotMainHand && p.item.HandType == proto.HandType_HandTypeTwoHand {
				delete(gear, proto.ItemSlot_ItemSlotOffHand)
			}
			if !slices.Contains(changed, slot) {
				changed = append(changed, slot)
			}
		}

		for _, q := range in.route.GetQuests() {
			if q.Level != level {
				continue
			}
			take := func(ids []int32) []gearPiece {
				var out []gearPiece
				for _, id := range ids {
					if item := in.itemsByID[id]; item != nil && pool.canTakeAt(item, level) {
						out = append(out, gearPiece{item: item, source: "quest", since: level})
					}
				}
				return out
			}
			pieces := take(q.RewardIds)
			if choices := take(q.ChoiceIds); len(choices) > 0 {
				best := choices[0]
				for _, c := range choices[1:] {
					_, _, gc := gainOf(c)
					_, _, gb := gainOf(best)
					if gc > gb {
						best = c
					}
				}
				pieces = append([]gearPiece{best}, pieces...)
			}
			for _, p := range pieces {
				offer(p)
			}
		}
		for _, f := range findsPerLevel {
			rand := mulberry32(hash(in.seed, level, int32(f.itemType)))
			count := poisson(f.rate, rand)
			for i := 0; i < count; i++ {
				quest := rand() < questShare
				if quest && in.route != nil {
					continue
				}
				var choices []gearPiece
				n := 1
				if quest {
					n = questChoices
				}
				for c := 0; c < n; c++ {
					if p, ok := pool.find(f.itemType, level, rand); ok {
						choices = append(choices, p)
					}
				}
				if len(choices) == 0 {
					continue
				}
				// A quest lets us choose, so we take the reward that helps us most.
				best := choices[0]
				for _, c := range choices[1:] {
					_, _, gc := gainOf(c)
					_, _, gb := gainOf(best)
					if gc > gb {
						best = c
					}
				}
				if quest {
					best.source = "quest"
				} else {
					best.source = "drop"
				}
				offer(best)
			}
		}
		for _, run := range runs {
			if run.level != level {
				continue
			}
			rand := mulberry32(hash(in.seed, level, 1000+run.zoneID))
			for _, loot := range run.bosses {
				drops := 1
				if rand() < 0.5 {
					drops = 2
				}
				for i := 0; i < drops; i++ {
					var usable []*proto.UIItem
					for _, item := range loot {
						if pool.canWearAt(item, level) {
							usable = append(usable, item)
						}
					}
					if len(usable) == 0 {
						continue
					}
					p := pool.withSuffix(usable[int(math.Floor(rand()*float64(len(usable))))], rand, "dungeon", level)
					if _, ok, _ := gainOf(p); !ok || rand() >= dungeonWinChance {
						continue
					}
					offer(p)
				}
			}
		}
		for _, pin := range in.pins {
			if pin.Level != level || pin.Item == nil {
				continue
			}
			item := in.itemsByID[pin.Item.Id]
			if item == nil {
				continue
			}
			offer(gearPiece{item: item, suffix: in.suffixes[pin.Item.RandomSuffix], source: "pinned", since: level})
		}

		snapshot := make(map[proto.ItemSlot]gearPiece, len(gear))
		for slot, p := range gear {
			snapshot[slot] = p
		}
		levels = append(levels, gearLevel{level: level, gear: snapshot, changed: changed})
	}
	return &gearPlan{levels: levels, madeUp: pool.usedPlain}
}

type dungeonRun struct {
	level  int32
	zoneID int32
	bosses [][]*proto.UIItem
}

func planDungeonRuns(in *gearPlanInputs, pool *itemPool) []dungeonRun {
	var runs []dungeonRun
	if in.route != nil {
		for _, d := range in.route.Dungeons {
			if bosses := pool.bossLoot(d.ZoneId); len(bosses) > 0 {
				runs = append(runs, dungeonRun{level: d.Level, zoneID: d.ZoneId, bosses: bosses})
			}
		}
		return runs
	}
	for _, d := range levelingDungeons {
		if d.faction != proto.Faction_Unknown && d.faction != in.faction {
			continue
		}
		bosses := pool.bossLoot(d.zoneID)
		start := int32(math.MaxInt32)
		for _, loot := range bosses {
			for _, item := range loot {
				if item.RequiredLevel > 0 {
					start = min(start, item.RequiredLevel)
				}
			}
		}
		if start == math.MaxInt32 {
			continue
		}
		rand := mulberry32(hash(in.seed, 0, 1000+d.zoneID))
		if rand() >= dungeonRunChance {
			continue
		}
		for _, offset := range dungeonRunLevels {
			runs = append(runs, dungeonRun{level: min(start+offset, gearMaxLevel), zoneID: d.zoneID, bosses: bosses})
		}
	}
	return runs
}

// The slot a new piece goes in and what it gains us there, or false when it beats nothing we
// wear. A ring or a trinket replaces the weaker of the two, and we don't wear two of one item.
func bestSlotFor(p gearPiece, gear map[proto.ItemSlot]gearPiece, weights *weights, rules *proto.LevelingGearClassRules) (proto.ItemSlot, bool, float64) {
	var best proto.ItemSlot
	found := false
	bestGain := 0.0
	for _, slot := range slotsFor(p.item, rules) {
		if other, paired := pairedSlot(slot); paired {
			if worn, ok := gear[other]; ok && worn.item.Id == p.item.Id {
				continue
			}
		}
		gain := scorePiece(p, slot, weights) - currentScore(slot, gear, weights, p)
		if gain > bestGain+pickEpsilon {
			best, found, bestGain = slot, true, gain
		}
	}
	return best, found, bestGain
}

func pairedSlot(slot proto.ItemSlot) (proto.ItemSlot, bool) {
	switch slot {
	case proto.ItemSlot_ItemSlotFinger1:
		return proto.ItemSlot_ItemSlotFinger2, true
	case proto.ItemSlot_ItemSlotFinger2:
		return proto.ItemSlot_ItemSlotFinger1, true
	case proto.ItemSlot_ItemSlotTrinket1:
		return proto.ItemSlot_ItemSlotTrinket2, true
	case proto.ItemSlot_ItemSlotTrinket2:
		return proto.ItemSlot_ItemSlotTrinket1, true
	}
	return 0, false
}

// What we give up in a slot to wear the piece there. A two-hander also takes the off hand,
// and an off hand item can't go next to a two-hander.
func currentScore(slot proto.ItemSlot, gear map[proto.ItemSlot]gearPiece, weights *weights, p gearPiece) float64 {
	score := 0.0
	if worn, ok := gear[slot]; ok {
		score = scorePiece(worn, slot, weights)
	}
	if slot == proto.ItemSlot_ItemSlotMainHand && p.item.HandType == proto.HandType_HandTypeTwoHand {
		if oh, ok := gear[proto.ItemSlot_ItemSlotOffHand]; ok {
			score += scorePiece(oh, proto.ItemSlot_ItemSlotOffHand, weights)
		}
	}
	if slot == proto.ItemSlot_ItemSlotOffHand {
		if mh, ok := gear[proto.ItemSlot_ItemSlotMainHand]; ok && mh.item.HandType == proto.HandType_HandTypeTwoHand {
			return inf
		}
	}
	return score
}

func slotsFor(item *proto.UIItem, rules *proto.LevelingGearClassRules) []proto.ItemSlot {
	switch item.Type {
	case proto.ItemType_ItemTypeFinger:
		return []proto.ItemSlot{proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2}
	case proto.ItemType_ItemTypeTrinket:
		return []proto.ItemSlot{proto.ItemSlot_ItemSlotTrinket1, proto.ItemSlot_ItemSlotTrinket2}
	case proto.ItemType_ItemTypeWeapon:
		if item.HandType == proto.HandType_HandTypeOffHand || item.WeaponType == proto.WeaponType_WeaponTypeShield || item.WeaponType == proto.WeaponType_WeaponTypeOffHand {
			return []proto.ItemSlot{proto.ItemSlot_ItemSlotOffHand}
		}
		if item.HandType == proto.HandType_HandTypeTwoHand || item.HandType == proto.HandType_HandTypeMainHand {
			return []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}
		}
		if rules.CanDualWield {
			return []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand}
		}
		return []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}
	case proto.ItemType_ItemTypeRanged:
		return []proto.ItemSlot{proto.ItemSlot_ItemSlotRanged}
	}
	for _, slot := range gearSlots {
		if slotType[slot] == item.Type {
			return []proto.ItemSlot{slot}
		}
	}
	return nil
}

func stat(values []float64, s proto.Stat) float64 { return at(values, int(s)) }

// scorePiece is a piece's worth by the stat weights, like the gear picker's EP. Under Forever
// an item's hit and crit count for melee and spells both, so we price them once, at the
// larger weight. Bonus armor takes the armor price when it has none of its own.
func scorePiece(p gearPiece, slot proto.ItemSlot, weights *weights) float64 {
	item := p.item
	values := make([]float64, max(len(item.Stats), len(p.suffix.GetStats())))
	copy(values, item.Stats)
	for i, v := range p.suffix.GetStats() {
		values[i] += v
	}
	score := 0.0
	for s := 0; s < len(item.Stats); s++ {
		switch proto.Stat(s) {
		case proto.Stat_StatMeleeHit, proto.Stat_StatSpellHit, proto.Stat_StatMeleeCrit, proto.Stat_StatSpellCrit, proto.Stat_StatArmor, proto.Stat_StatBonusArmor:
			continue
		}
		score += values[s] * weights.stat(proto.Stat(s))
	}
	hit := stat(values, proto.Stat_StatMeleeHit) + stat(values, proto.Stat_StatSpellHit) + item.HitRating/10
	crit := stat(values, proto.Stat_StatMeleeCrit) + stat(values, proto.Stat_StatSpellCrit) + item.CritRating/14
	score += hit * max(weights.stat(proto.Stat_StatMeleeHit), weights.stat(proto.Stat_StatSpellHit))
	score += crit * max(weights.stat(proto.Stat_StatMeleeCrit), weights.stat(proto.Stat_StatSpellCrit))
	armorWeight := weights.stat(proto.Stat_StatArmor)
	bonusArmorWeight := weights.stat(proto.Stat_StatBonusArmor)
	if bonusArmorWeight == 0 {
		bonusArmorWeight = armorWeight
	}
	score += stat(values, proto.Stat_StatArmor)*armorWeight + stat(values, proto.Stat_StatBonusArmor)*bonusArmorWeight
	if item.WeaponSpeed > 0 {
		dps := weaponDPS(item)
		switch slot {
		case proto.ItemSlot_ItemSlotMainHand:
			score += dps * weights.pseudoStat(proto.PseudoStat_PseudoStatMainHandDps)
		case proto.ItemSlot_ItemSlotOffHand:
			score += dps * weights.pseudoStat(proto.PseudoStat_PseudoStatOffHandDps)
		case proto.ItemSlot_ItemSlotRanged:
			score += dps * weights.pseudoStat(proto.PseudoStat_PseudoStatRangedDps)
		}
	}
	// An item with nothing we price still beats an empty slot, a little.
	return score + emptySlotPreferred
}

func weaponDPS(item *proto.UIItem) float64 {
	speed := item.WeaponSpeed
	if speed == 0 {
		speed = 1
	}
	return (item.WeaponDamageMin + item.WeaponDamageMax) / 2 / speed
}

// The items we can come across, by kind of slot and quality.
type itemPool struct {
	in            *gearPlanInputs
	byTypeQuality map[[2]int32][]*proto.UIItem
	// Each dungeon's boss loot we could use, by zone, its bosses in the order we met them.
	bossOrder map[int32][]int32
	bossItems map[int32]map[int32][]*proto.UIItem
	plain     *plainItems
	usedPlain map[int32]*proto.UIItem
}

func newItemPool(in *gearPlanInputs) *itemPool {
	pool := &itemPool{
		in:            in,
		byTypeQuality: map[[2]int32][]*proto.UIItem{},
		bossOrder:     map[int32][]int32{},
		bossItems:     map[int32]map[int32][]*proto.UIItem{},
		plain:         newPlainItems(in.items),
		usedPlain:     map[int32]*proto.UIItem{},
	}
	for _, item := range in.items {
		if !pool.canFind(item) {
			continue
		}
		// A dungeon boss's or a raid's loot only drops there, so it's not a find out in the
		// world.
		instanceOnly := len(item.Sources) > 0
		for _, s := range item.Sources {
			if d := s.GetDrop(); d == nil || !instanceZones[d.ZoneId] {
				instanceOnly = false
			}
		}
		// With a route, every find is a drop, and a quest reward doesn't drop.
		questOnly := len(item.Sources) > 0
		for _, s := range item.Sources {
			if s.GetQuest() == nil {
				questOnly = false
			}
		}
		if !instanceOnly && !(questOnly && in.route != nil) {
			key := [2]int32{int32(item.Type), int32(item.Quality)}
			pool.byTypeQuality[key] = append(pool.byTypeQuality[key], item)
		}
		for _, s := range item.Sources {
			d := s.GetDrop()
			if d == nil || d.NpcId == 0 {
				continue
			}
			if pool.bossItems[d.ZoneId] == nil {
				pool.bossItems[d.ZoneId] = map[int32][]*proto.UIItem{}
			}
			loot := pool.bossItems[d.ZoneId]
			if _, seen := loot[d.NpcId]; !seen {
				pool.bossOrder[d.ZoneId] = append(pool.bossOrder[d.ZoneId], d.NpcId)
			}
			if !slices.Contains(loot[d.NpcId], item) {
				loot[d.NpcId] = append(loot[d.NpcId], item)
			}
		}
	}
	return pool
}

// Whether we could come across the item while leveling: a green or a blue our faction and
// class can use, from a drop or a quest, or a world drop with no source.
func (pool *itemPool) canFind(item *proto.UIItem) bool {
	if item.Quality != proto.ItemQuality_ItemQualityUncommon && item.Quality != proto.ItemQuality_ItemQualityRare {
		return false
	}
	if len(item.Sources) > 0 && !slices.ContainsFunc(item.Sources, func(s *proto.UIItemSource) bool { return s.GetDrop() != nil || s.GetQuest() != nil }) {
		return false
	}
	return pool.canUse(item)
}

// Whether our class and faction can use the item, and our class can wield it.
func (pool *itemPool) canUse(item *proto.UIItem) bool {
	if len(item.ClassAllowlist) > 0 && !slices.Contains(item.ClassAllowlist, pool.in.playerClass) {
		return false
	}
	switch item.FactionRestriction {
	case proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY:
		if pool.in.faction == proto.Faction_Horde {
			return false
		}
	case proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY:
		if pool.in.faction == proto.Faction_Alliance {
			return false
		}
	}
	switch item.Type {
	case proto.ItemType_ItemTypeWeapon:
		return pool.canWield(item.WeaponType, item.HandType == proto.HandType_HandTypeTwoHand)
	case proto.ItemType_ItemTypeRanged:
		return slices.Contains(pool.in.rules.RangedWeaponTypes, item.RangedWeaponType)
	}
	return true
}

func (pool *itemPool) canWield(weaponType proto.WeaponType, twoHand bool) bool {
	if !slices.Contains(pool.in.rules.WeaponTypes, weaponType) {
		return false
	}
	return !twoHand || slices.Contains(pool.in.rules.TwoHandWeaponTypes, weaponType)
}

func (pool *itemPool) bossLoot(zoneID int32) [][]*proto.UIItem {
	var out [][]*proto.UIItem
	for _, npc := range pool.bossOrder[zoneID] {
		out = append(out, pool.bossItems[zoneID][npc])
	}
	return out
}

func (pool *itemPool) maxArmorAt(level int32) proto.ArmorType {
	rows := pool.in.rules.MaxArmorType
	if len(rows) == 0 {
		return proto.ArmorType_ArmorTypeUnknown
	}
	return rows[min(max(int(level), 1), len(rows))-1]
}

// Whether we can wear the item at this level: its required level and our armor type then.
func (pool *itemPool) canWearAt(item *proto.UIItem, level int32) bool {
	return item.RequiredLevel <= level && (item.ArmorType == proto.ArmorType_ArmorTypeUnknown || item.ArmorType <= pool.maxArmorAt(level))
}

// Whether we can take a route's quest reward at this level and wear it.
func (pool *itemPool) canTakeAt(item *proto.UIItem, level int32) bool {
	return pool.canUse(item) && pool.canWearAt(item, level)
}

// The item as a piece we have, with one of its random suffixes rolled.
func (pool *itemPool) withSuffix(item *proto.UIItem, rand func() float64, source string, level int32) gearPiece {
	p := gearPiece{item: item, source: source, since: level}
	if n := len(item.RandomSuffixOptions); n > 0 {
		p.suffix = pool.in.suffixes[item.RandomSuffixOptions[int(math.Floor(rand()*float64(n)))]]
	}
	return p
}

// An item we come across at this level, or false when there's none of the rolled kind.
func (pool *itemPool) find(itemType proto.ItemType, level int32, rand func() float64) (gearPiece, bool) {
	quality := rollQuality(level, rand())
	if quality == proto.ItemQuality_ItemQualityCommon {
		ilvl := max(1, level+int32(math.Floor(rand()*(ilvlAbove+1))))
		item := pool.plainItem(itemType, ilvl, level, rand)
		if item == nil {
			return gearPiece{}, false
		}
		return gearPiece{item: item, source: "drop", since: level}, true
	}
	maxArmor := pool.maxArmorAt(level)
	var candidates []*proto.UIItem
	for _, item := range pool.byTypeQuality[[2]int32{int32(itemType), int32(quality)}] {
		if item.RequiredLevel <= level && item.Ilvl >= level-ilvlBelow && item.Ilvl <= level+ilvlAbove &&
			(item.ArmorType == proto.ArmorType_ArmorTypeUnknown || item.ArmorType <= maxArmor) {
			candidates = append(candidates, item)
		}
	}
	if len(candidates) == 0 {
		return gearPiece{}, false
	}
	return pool.withSuffix(candidates[int(math.Floor(rand()*float64(len(candidates))))], rand, "drop", level), true
}

var plainWeapons = []struct {
	weaponType proto.WeaponType
	twoHand    bool
}{
	{proto.WeaponType_WeaponTypeAxe, false},
	{proto.WeaponType_WeaponTypeAxe, true},
	{proto.WeaponType_WeaponTypeMace, false},
	{proto.WeaponType_WeaponTypeMace, true},
	{proto.WeaponType_WeaponTypeSword, false},
	{proto.WeaponType_WeaponTypeSword, true},
	{proto.WeaponType_WeaponTypeDagger, false},
	{proto.WeaponType_WeaponTypeFist, false},
	{proto.WeaponType_WeaponTypeStaff, true},
	{proto.WeaponType_WeaponTypePolearm, true},
}

var plainArmorTypes = []proto.ItemType{
	proto.ItemType_ItemTypeHead, proto.ItemType_ItemTypeShoulder, proto.ItemType_ItemTypeBack, proto.ItemType_ItemTypeChest,
	proto.ItemType_ItemTypeWrist, proto.ItemType_ItemTypeHands, proto.ItemType_ItemTypeWaist, proto.ItemType_ItemTypeLegs,
	proto.ItemType_ItemTypeFeet,
}

// A plain white item of this kind and item level: armor of the best type we can wear at this
// level, or a weapon of a kind we can use.
func (pool *itemPool) plainItem(itemType proto.ItemType, ilvl, level int32, rand func() float64) *proto.UIItem {
	var item *proto.UIItem
	if itemType == proto.ItemType_ItemTypeWeapon {
		var kinds []int
		for i, w := range plainWeapons {
			if pool.canWield(w.weaponType, w.twoHand) {
				kinds = append(kinds, i)
			}
		}
		if len(kinds) == 0 {
			return nil
		}
		w := plainWeapons[kinds[int(math.Floor(rand()*float64(len(kinds))))]]
		item = pool.plain.weapon(w.weaponType, w.twoHand, ilvl)
	} else {
		armorType := pool.maxArmorAt(level)
		if itemType == proto.ItemType_ItemTypeBack {
			armorType = proto.ArmorType_ArmorTypeCloth
		}
		item = pool.plain.armor(itemType, armorType, ilvl)
	}
	if item != nil {
		pool.usedPlain[item.Id] = item
	}
	return item
}

// plainItems are the white items the plan makes up, since the item database has none (the TS
// PlainItems). Their armor and weapon DPS go with item level, at 90% of a green piece's armor
// and 85% of a green weapon's DPS, by lines fit to the database.
type plainItems struct {
	armorFit  map[[2]int32][2]float64
	weaponFit map[bool][2]float64
}

func newPlainItems(items []*proto.UIItem) *plainItems {
	weaponSamples := map[bool][][2]float64{}
	armorSamples := map[[2]int32][][2]float64{}
	for _, item := range items {
		if item.Quality != proto.ItemQuality_ItemQualityUncommon {
			continue
		}
		twoHand := item.HandType == proto.HandType_HandTypeTwoHand
		if item.WeaponSpeed > 0 {
			weaponSamples[twoHand] = append(weaponSamples[twoHand], [2]float64{float64(item.Ilvl), weaponDPS(item)})
		} else if a := stat(item.Stats, proto.Stat_StatArmor); a > 0 {
			key := [2]int32{int32(item.Type), int32(item.ArmorType)}
			armorSamples[key] = append(armorSamples[key], [2]float64{float64(item.Ilvl), a})
		}
	}
	p := &plainItems{armorFit: map[[2]int32][2]float64{}, weaponFit: map[bool][2]float64{}}
	for k, s := range armorSamples {
		p.armorFit[k] = lineFit(s)
	}
	for k, s := range weaponSamples {
		p.weaponFit[k] = lineFit(s)
	}
	return p
}

func (p *plainItems) armor(itemType proto.ItemType, armorType proto.ArmorType, ilvl int32) *proto.UIItem {
	if !slices.Contains(plainArmorTypes, itemType) {
		return nil
	}
	fit := p.armorFit[[2]int32{int32(itemType), int32(armorType)}]
	item := madeUpItem(plainID(itemType, int32(armorType), ilvl), itemType, ilvl)
	item.ArmorType = armorType
	item.Stats[proto.Stat_StatArmor] = max(1, jsRound(0.9*(fit[0]+fit[1]*float64(ilvl))))
	return item
}

func (p *plainItems) weapon(weaponType proto.WeaponType, twoHand bool, ilvl int32) *proto.UIItem {
	speed := 2.4
	if twoHand {
		speed = 3.3
	}
	fit := p.weaponFit[twoHand]
	dps := max(1, 0.85*(fit[0]+fit[1]*float64(ilvl)))
	kind := int32(weaponType) * 10
	if twoHand {
		kind++
	}
	item := madeUpItem(plainID(proto.ItemType_ItemTypeWeapon, kind, ilvl), proto.ItemType_ItemTypeWeapon, ilvl)
	item.WeaponType = weaponType
	item.HandType = proto.HandType_HandTypeMainHand
	if twoHand {
		item.HandType = proto.HandType_HandTypeTwoHand
	}
	item.WeaponSpeed = speed
	item.WeaponDamageMin = jsRound(dps * speed * 0.8)
	item.WeaponDamageMax = jsRound(dps * speed * 1.2)
	return item
}

func plainID(itemType proto.ItemType, kind, ilvl int32) int32 {
	return plainIDBase + int32(itemType)*100_000 + kind*100 + ilvl
}

func madeUpItem(id int32, itemType proto.ItemType, ilvl int32) *proto.UIItem {
	return &proto.UIItem{
		Id:            id,
		Name:          "Plain item",
		Type:          itemType,
		Stats:         make([]float64, int(proto.Stat_StatFeralAttackPower)+1),
		Ilvl:          ilvl,
		RequiredLevel: max(1, ilvl-5),
		Quality:       proto.ItemQuality_ItemQualityCommon,
	}
}

func rollQuality(level int32, roll float64) proto.ItemQuality {
	white, green := qualityMixAt(level)
	switch {
	case roll < white:
		return proto.ItemQuality_ItemQualityCommon
	case roll < white+green:
		return proto.ItemQuality_ItemQualityUncommon
	}
	return proto.ItemQuality_ItemQualityRare
}

func qualityMixAt(level int32) (white, green float64) {
	l := float64(level)
	if l <= qualityMix[0].level {
		return qualityMix[0].white, qualityMix[0].green
	}
	for i := 1; i < len(qualityMix); i++ {
		if l <= qualityMix[i].level {
			a, b := qualityMix[i-1], qualityMix[i]
			t := (l - a.level) / (b.level - a.level)
			return a.white + t*(b.white-a.white), a.green + t*(b.green-a.green)
		}
	}
	last := qualityMix[len(qualityMix)-1]
	return last.white, last.green
}

// The least squares line through the points, as [intercept, slope], summed in the same order
// as the TS so the numbers come out the same.
func lineFit(points [][2]float64) [2]float64 {
	n := float64(len(points))
	if len(points) < 2 {
		if len(points) == 1 {
			return [2]float64{points[0][1], 0}
		}
		return [2]float64{0, 0}
	}
	var sx, sy float64
	for _, p := range points {
		sx += p[0]
	}
	for _, p := range points {
		sy += p[1]
	}
	mx, my := sx/n, sy/n
	var sxx, sxy float64
	for _, p := range points {
		sxx += (p[0] - mx) * (p[0] - mx)
	}
	for _, p := range points {
		sxy += (p[0] - mx) * (p[1] - my)
	}
	slope := 0.0
	if sxx > 0 {
		slope = sxy / sxx
	}
	return [2]float64{my - slope*mx, slope}
}

// Math.round: halves go up.
func jsRound(x float64) float64 {
	return math.Floor(x + 0.5)
}

func poisson(mean float64, rand func() float64) int {
	limit := math.Exp(-mean)
	k := 0
	p := rand()
	for p > limit {
		k++
		p *= rand()
	}
	return k
}

// The TS hash and mulberry32, in uint32 math, which has the same bits as JS's int32 math.
func hash(values ...int32) uint32 {
	h := uint32(2166136261)
	for _, v := range values {
		h ^= uint32(v)
		h *= 16777619
		h ^= h >> 13
	}
	return h
}

func mulberry32(seed uint32) func() float64 {
	a := seed
	return func() float64 {
		a += 0x6d2b79f5
		t := (a ^ (a >> 15)) * (1 | a)
		t = (t + (t^(t>>7))*(61|t)) ^ t
		return float64(t^(t>>14)) / 4294967296
	}
}
