package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
	"google.golang.org/protobuf/encoding/protojson"
	goproto "google.golang.org/protobuf/proto"
)

// Gear search. Every slot in turn tries each item the player can equip at their level,
// with the other slots held fixed, and keeps the best one when it beats the current item
// by more than the noise. That repeats until a full round changes nothing, the same
// coordinate descent the knobs use. The rotation and talents are held fixed, so run this
// with a sensible build, then search talents and knobs again on the gear it finds.
//
// The item pool comes from the UI database (assets/database/db.json), which has the
// fields the sim's own item table drops: quality, class limits, faction, random suffixes.
// The level an item needs is not in either database, so it is read from the wago item
// export under assets/db_inputs.

// One thing that can go in a slot: an item, possibly with a random suffix.
type gearOption struct {
	spec *proto.ItemSpec
	name string
	// twoHand marks a main hand option that leaves no room for an off hand.
	twoHand bool
	// unsourced marks an item wowhead lists no way of getting at all. Forever is in beta
	// and half of what it added has no drop, quest or vendor on wowhead yet, so these are
	// worth searching over, but a set that picks one needs a look by hand.
	unsourced bool
	// once marks an item a character can only hold one of: a unique item, or a quest
	// reward with no other source (Silent Hunter from Call to Arms, say). The main and
	// off hand never both hold it.
	once bool
}

var gearSlotNames = []string{"head", "neck", "shoulder", "back", "chest", "wrist", "hands", "waist", "legs", "feet", "finger1", "finger2", "trinket1", "trinket2", "mainhand", "offhand", "ranged"}

// Armor the class can wear at the level, in order of preference when two items tie.
func armorTypesFor(class proto.Class, level int32) []proto.ArmorType {
	switch class {
	case proto.Class_ClassWarrior, proto.Class_ClassPaladin:
		if level >= 40 {
			return []proto.ArmorType{proto.ArmorType_ArmorTypePlate, proto.ArmorType_ArmorTypeMail, proto.ArmorType_ArmorTypeLeather, proto.ArmorType_ArmorTypeCloth}
		}
		return []proto.ArmorType{proto.ArmorType_ArmorTypeMail, proto.ArmorType_ArmorTypeLeather, proto.ArmorType_ArmorTypeCloth}
	case proto.Class_ClassHunter, proto.Class_ClassShaman:
		if level >= 40 {
			return []proto.ArmorType{proto.ArmorType_ArmorTypeMail, proto.ArmorType_ArmorTypeLeather, proto.ArmorType_ArmorTypeCloth}
		}
		return []proto.ArmorType{proto.ArmorType_ArmorTypeLeather, proto.ArmorType_ArmorTypeCloth}
	case proto.Class_ClassRogue, proto.Class_ClassDruid:
		return []proto.ArmorType{proto.ArmorType_ArmorTypeLeather, proto.ArmorType_ArmorTypeCloth}
	default:
		return []proto.ArmorType{proto.ArmorType_ArmorTypeCloth}
	}
}

var weaponTypesFor = map[proto.Class][]proto.WeaponType{
	// The rogue gets axes under Forever, its Hack and Slash reads "Axe/Sword".
	proto.Class_ClassRogue:   {proto.WeaponType_WeaponTypeDagger, proto.WeaponType_WeaponTypeSword, proto.WeaponType_WeaponTypeMace, proto.WeaponType_WeaponTypeFist, proto.WeaponType_WeaponTypeAxe},
	proto.Class_ClassWarrior: {proto.WeaponType_WeaponTypeDagger, proto.WeaponType_WeaponTypeSword, proto.WeaponType_WeaponTypeMace, proto.WeaponType_WeaponTypeFist, proto.WeaponType_WeaponTypeAxe, proto.WeaponType_WeaponTypePolearm, proto.WeaponType_WeaponTypeStaff},
	proto.Class_ClassShaman:  {proto.WeaponType_WeaponTypeDagger, proto.WeaponType_WeaponTypeMace, proto.WeaponType_WeaponTypeFist, proto.WeaponType_WeaponTypeAxe, proto.WeaponType_WeaponTypeStaff, proto.WeaponType_WeaponTypeShield, proto.WeaponType_WeaponTypeOffHand},
	proto.Class_ClassPaladin: {proto.WeaponType_WeaponTypeSword, proto.WeaponType_WeaponTypeMace, proto.WeaponType_WeaponTypeAxe, proto.WeaponType_WeaponTypePolearm, proto.WeaponType_WeaponTypeShield, proto.WeaponType_WeaponTypeOffHand},
}

var rangedTypesFor = map[proto.Class][]proto.RangedWeaponType{
	proto.Class_ClassRogue:   {proto.RangedWeaponType_RangedWeaponTypeBow, proto.RangedWeaponType_RangedWeaponTypeCrossbow, proto.RangedWeaponType_RangedWeaponTypeGun, proto.RangedWeaponType_RangedWeaponTypeThrown},
	proto.Class_ClassWarrior: {proto.RangedWeaponType_RangedWeaponTypeBow, proto.RangedWeaponType_RangedWeaponTypeCrossbow, proto.RangedWeaponType_RangedWeaponTypeGun, proto.RangedWeaponType_RangedWeaponTypeThrown},
	proto.Class_ClassShaman:  {proto.RangedWeaponType_RangedWeaponTypeTotem},
	proto.Class_ClassPaladin: {proto.RangedWeaponType_RangedWeaponTypeLibram},
}

func contains[T comparable](list []T, v T) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// The items that need a PvP rank, from the wago export. Only that file records it.
func loadPvpRankItems(path string) (map[int32]bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.LazyQuotes = true
	header, err := r.Read()
	if err != nil {
		return nil, err
	}
	idCol, rankCol := -1, -1
	for i, name := range header {
		switch name {
		case "ID":
			idCol = i
		case "RequiredPVPRank":
			rankCol = i
		}
	}
	if idCol < 0 || rankCol < 0 {
		return nil, fmt.Errorf("%s: no ID/RequiredPVPRank columns", path)
	}
	ranked := map[int32]bool{}
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if rank, _ := strconv.Atoi(row[rankCol]); rank > 0 {
			id, _ := strconv.Atoi(row[idCol])
			ranked[int32(id)] = true
		}
	}
	return ranked, nil
}

// Whether a random suffix does anything for the class: a physical stat for a melee, and
// for a shaman also the caster stats that feed shocks, Lightning Shield and the mana pool.
func usefulSuffix(class proto.Class, s *proto.ItemRandomSuffix) bool {
	st := stats.FromFloatArray(s.Stats)
	useful := st[stats.Agility] + st[stats.Strength] + st[stats.AttackPower] + st[stats.MeleeCrit] + st[stats.MeleeHit] + st[stats.Stamina]
	if class == proto.Class_ClassShaman || class == proto.Class_ClassPaladin {
		useful += st[stats.Intellect] + st[stats.Spirit] + st[stats.MP5] + st[stats.SpellPower] + st[stats.NaturePower] + st[stats.FirePower] + st[stats.SpellCrit] + st[stats.SpellHit]
	}
	return useful > 0
}

// The Forever items wowhead lists no way of getting: not a drop, a quest, a vendor or a
// recipe, not even an unnamed one.
//
// The database carries them anyway, because wowhead has no source for half of what
// Forever added and leaving them out cost us Wolfsbane, but a set that picks one is
// worth a look by hand. An item with a source of its own in the database, hand entered
// in forever_items.go, is not one of these however the listing files it.
func loadUnsourcedItems(path string) (map[int32]bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var listing struct {
		Items []struct {
			ID     int32   `json:"id"`
			Source []int32 `json:"source"`
		} `json:"items"`
	}
	if err := json.Unmarshal(data, &listing); err != nil {
		return nil, err
	}
	unsourced := map[int32]bool{}
	for _, item := range listing.Items {
		if len(item.Source) == 0 {
			unsourced[item.ID] = true
		}
	}
	return unsourced, nil
}

// The PvP rank titles that start the names of Forever's rank gear, with or without
// "Premier" in front.
var foreverPvpRankName = regexp.MustCompile(`^(Premier )?(Private|Scout|Corporal|Grunt|Sergeant|Senior Sergeant|Master Sergeant|First Sergeant|Sergeant Major|Stone Guard|Knight|Blood Guard|Knight-Lieutenant|Legionnaire|Knight-Captain|Centurion|Knight-Champion|Champion|Lieutenant Commander|Lieutenant General|Commander|General|Marshal|Warlord|Field Marshal|High Warlord|Grand Marshal)'s `)

// Every option for every slot the player can use at their level.
func loadGearPool(dbPath, levelsPath, foreverPath string, player *proto.Player, level int32, minQuality proto.ItemQuality, maxPhase, maxIlvl int32, exclude []int32, hitOnly, synthetic bool) ([][]gearOption, error) {
	data, err := os.ReadFile(dbPath)
	if err != nil {
		return nil, err
	}
	db := &proto.UIDatabase{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(data, db); err != nil {
		return nil, err
	}
	pvpRank, err := loadPvpRankItems(levelsPath)
	if err != nil {
		return nil, err
	}
	unsourcedItems, err := loadUnsourcedItems(foreverPath)
	if err != nil {
		return nil, err
	}
	suffixes := map[int32]*proto.ItemRandomSuffix{}
	for _, s := range db.RandomSuffixes {
		suffixes[s.Id] = s
	}

	class := player.Class
	horde := contains([]proto.Race{proto.Race_RaceOrc, proto.Race_RaceTroll, proto.Race_RaceTauren, proto.Race_RaceUndead}, player.Race)
	armor := armorTypesFor(class, level)
	pool := make([][]gearOption, len(gearSlotNames))

	add := func(slot proto.ItemSlot, item *proto.UIItem) {
		twoHand := item.HandType == proto.HandType_HandTypeTwoHand
		unsourced := len(item.Sources) == 0 && unsourcedItems[item.Id]
		once := item.Unique || len(item.Sources) > 0 && !slices.ContainsFunc(item.Sources, func(src *proto.UIItemSource) bool {
			return src.GetQuest() == nil
		})
		itemHit := hasHit(item.Stats) || item.HitRating > 0 || slices.ContainsFunc(item.WeaponSkills, func(v float64) bool { return v > 0 })
		if len(item.RandomSuffixOptions) == 0 {
			if hitOnly && !itemHit {
				return
			}
			pool[slot] = append(pool[slot], gearOption{&proto.ItemSpec{Id: item.Id}, item.Name, twoHand, unsourced, once})
			return
		}
		for _, id := range item.RandomSuffixOptions {
			if s, ok := suffixes[id]; ok && usefulSuffix(class, s) && (!hitOnly || itemHit || hasHit(s.Stats)) {
				pool[slot] = append(pool[slot], gearOption{&proto.ItemSpec{Id: item.Id, RandomSuffix: id}, item.Name + " " + s.Name, twoHand, unsourced, once})
			}
		}
	}

	for _, enchant := range db.Enchants {
		// Classic shares an effect between enchants with the same stats on different
		// slots (852 is Stamina +5 on bracers, boots and shields). Those print without
		// the slot, so the boots don't read "Enchant Shield - Stamina".
		if old, ok := enchantNames[enchant.EffectId]; ok && old != enchant.Name {
			_, stat, _ := strings.Cut(enchant.Name, " - ")
			enchantNames[enchant.EffectId] = "Enchant - " + stat
		} else if !ok {
			enchantNames[enchant.EffectId] = enchant.Name
		}
		enchantLevels[enchant.EffectId] = enchant.RequiredLevel
	}
	for _, item := range db.Items {
		itemsByID[item.Id] = item
		if item.HandType == proto.HandType_HandTypeTwoHand {
			twoHandItems[item.Id] = true
		}
		if item.Quality < minQuality || item.Expansion > proto.Expansion_ExpansionVanilla || item.RequiredLevel > level {
			continue
		}
		// The phase is the raid tier the item drops in (1 MC/Onyxia, 2 Dire Maul, 3 BWL,
		// 4 ZG, 5 AQ, 6 Naxx), for a "phase 2 BiS" style search.
		if maxPhase > 0 && item.Phase > maxPhase {
			continue
		}
		// Forever's new items know no raid tier, so the item level is what keeps a
		// "phase 1" search to Phase 1 power.
		if maxIlvl > 0 && item.Ilvl > maxIlvl {
			continue
		}
		// The client carries Season of Discovery's items, with sources, but Forever does
		// not use them (the item level 98 Scarlet Enclave sets came up in a level 60 search).
		if isSodItem(item.Id) {
			continue
		}
		// The synthetic Phase 1 items are not in the game. They join only when we ask for them.
		if isSyntheticItem(item.Id) && !synthetic {
			continue
		}
		if contains(exclude, item.Id) {
			continue
		}
		// A few items have no level of their own (Olmann Sewar, the engineering goggles),
		// the quest or profession is the gate. A quest's reward sits a few item levels above
		// the quest level, so that stands in.
		if item.RequiredLevel == 0 && item.Ilvl > level+6 {
			continue
		}
		// PvP rank gear is a separate grind, it stays out of the pool. Forever's own rank
		// sets ("Premier High Warlord's Blade", "Warlord's Linked Armor") are not in the wago
		// export and have no source on wowhead, so their rank title gives them away.
		if pvpRank[item.Id] || item.Id >= foreverFirstItemID && foreverPvpRankName.MatchString(item.Name) {
			continue
		}
		if len(item.ClassAllowlist) > 0 && !contains(item.ClassAllowlist, class) {
			continue
		}
		// Engineering goggles need the profession to be worn, and Forever's bind on pickup
		// crafted sets need it to be made. The player's two professions come from the
		// settings or the -professions flag. Crafted gear from another profession is only
		// in when it binds on equip, because then it can be bought.
		hasProfession := func(prof proto.Profession) bool {
			return prof == proto.Profession_ProfessionUnknown || prof == player.Profession1 || prof == player.Profession2
		}
		if !hasProfession(item.RequiredProfession) || item.BindOnPickup && !hasProfession(craftedProfession(item)) {
			continue
		}
		if item.FactionRestriction == proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY && !horde ||
			item.FactionRestriction == proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY && horde {
			continue
		}
		switch item.Type {
		case proto.ItemType_ItemTypeWeapon:
			if !contains(weaponTypesFor[class], item.WeaponType) {
				continue
			}
			switch item.HandType {
			case proto.HandType_HandTypeMainHand:
				add(proto.ItemSlot_ItemSlotMainHand, item)
			case proto.HandType_HandTypeOffHand:
				add(proto.ItemSlot_ItemSlotOffHand, item)
			case proto.HandType_HandTypeOneHand:
				add(proto.ItemSlot_ItemSlotMainHand, item)
				add(proto.ItemSlot_ItemSlotOffHand, item)
			case proto.HandType_HandTypeTwoHand:
				if class != proto.Class_ClassRogue {
					add(proto.ItemSlot_ItemSlotMainHand, item)
				}
			}
		case proto.ItemType_ItemTypeRanged:
			if contains(rangedTypesFor[class], item.RangedWeaponType) {
				add(proto.ItemSlot_ItemSlotRanged, item)
			}
		case proto.ItemType_ItemTypeFinger:
			add(proto.ItemSlot_ItemSlotFinger1, item)
			add(proto.ItemSlot_ItemSlotFinger2, item)
		case proto.ItemType_ItemTypeTrinket:
			add(proto.ItemSlot_ItemSlotTrinket1, item)
			add(proto.ItemSlot_ItemSlotTrinket2, item)
		case proto.ItemType_ItemTypeNeck, proto.ItemType_ItemTypeBack:
			add(proto.ItemSlot(item.Type-1), item)
		default:
			if item.ArmorType != proto.ArmorType_ArmorTypeUnknown && !contains(armor, item.ArmorType) {
				continue
			}
			if item.Type >= proto.ItemType_ItemTypeHead && item.Type <= proto.ItemType_ItemTypeFeet {
				add(proto.ItemSlot(item.Type-1), item)
			}
		}
	}
	for _, options := range pool {
		sort.Slice(options, func(i, j int) bool { return options[i].name < options[j].name })
	}
	return pool, nil
}

// Whether the stats give melee or spell hit.
func hasHit(stats []float64) bool {
	for _, stat := range []proto.Stat{proto.Stat_StatMeleeHit, proto.Stat_StatSpellHit} {
		if int(stat) < len(stats) && stats[stat] > 0 {
			return true
		}
	}
	return false
}

// The candidate with the enchant the slot has now, so a swap is not also a lost enchant.
func (s *searcher) withEnchantOf(slot int, spec, current *proto.ItemSpec) *proto.ItemSpec {
	if current == nil || current.Enchant == 0 {
		return spec
	}
	out := goproto.Clone(spec).(*proto.ItemSpec)
	out.Enchant = s.carriedEnchant(slot, spec.Id, current.Enchant)
	return out
}

// The equipment as a slot-indexed list, padded to every slot.
func equipmentSlots(e *proto.EquipmentSpec) []*proto.ItemSpec {
	slots := make([]*proto.ItemSpec, len(gearSlotNames))
	if e != nil {
		for i, item := range e.Items {
			if i < len(slots) && item != nil && item.Id != 0 {
				slots[i] = goproto.Clone(item).(*proto.ItemSpec)
			}
		}
	}
	return slots
}

func (s *searcher) setGear(slots []*proto.ItemSpec) {
	e := &proto.EquipmentSpec{}
	for _, item := range slots {
		if item == nil {
			e.Items = append(e.Items, &proto.ItemSpec{})
		} else {
			e.Items = append(e.Items, item)
		}
	}
	s.setup.player().Equipment = e
	// Gear is part of what the cache key leaves out, so it is cleared per gear set.
	s.cache = map[string]result{}
}

// screenGear keeps the options of a slot worth scoring at full iterations.
//
// A slot can have hundreds of options (every random suffix of every green), and most of
// them lose to the item in the slot by a lot. We score them all at screenIterations
// first and keep the best screenKeep, so only those cost full iterations. The number
// kept is generous next to the screen's noise: at 1000 iterations a candidate is about
// 0.4 DPS off, and the options that decide a slot are within a few tenths of each other.
func (s *searcher) screenGear(options []gearOption, candidateWith func(gearOption) []*proto.ItemSpec, knobs Knobs) []gearOption {
	if s.screenIterations == 0 || len(options) <= s.screenKeep {
		return options
	}
	type screened struct {
		option gearOption
		dps    float64
	}
	var scored []screened
	full := s.iterations
	s.iterations = s.screenIterations
	for _, o := range options {
		if candidate := candidateWith(o); candidate != nil {
			s.setGear(candidate)
			scored = append(scored, screened{o, s.objective(s.score(knobs))})
		}
	}
	s.iterations = full
	sort.Slice(scored, func(i, j int) bool { return scored[i].dps > scored[j].dps })
	kept := make([]gearOption, 0, s.screenKeep)
	for i := 0; i < len(scored) && i < s.screenKeep; i++ {
		kept = append(kept, scored[i].option)
	}
	return kept
}

// Whether the item in the main hand slot is a two hander, from the pool it came from.
func isTwoHand(pool [][]gearOption, item *proto.ItemSpec) bool {
	return item != nil && twoHandItems[item.Id]
}

// Every two hander in the item database, so the main hand the player starts with is
// known as one even when the pool leaves it out (an excluded item, or one below the
// quality floor).
var twoHandItems = map[int32]bool{}

// Which paired slot shares an item pool with this one, or -1.
func pairedSlot(slot int) int {
	switch proto.ItemSlot(slot) {
	case proto.ItemSlot_ItemSlotFinger1:
		return int(proto.ItemSlot_ItemSlotFinger2)
	case proto.ItemSlot_ItemSlotFinger2:
		return int(proto.ItemSlot_ItemSlotFinger1)
	case proto.ItemSlot_ItemSlotTrinket1:
		return int(proto.ItemSlot_ItemSlotTrinket2)
	case proto.ItemSlot_ItemSlotTrinket2:
		return int(proto.ItemSlot_ItemSlotTrinket1)
	}
	return -1
}

// The other weapon slot, for the items a character can only hold one of.
func weaponPair(slot int) int {
	switch proto.ItemSlot(slot) {
	case proto.ItemSlot_ItemSlotMainHand:
		return int(proto.ItemSlot_ItemSlotOffHand)
	case proto.ItemSlot_ItemSlotOffHand:
		return int(proto.ItemSlot_ItemSlotMainHand)
	}
	return -1
}

func (s *searcher) searchGear(pool [][]gearOption, knobs Knobs) ([]*proto.ItemSpec, result) {
	best := equipmentSlots(s.setup.player().Equipment)
	// An enchant the starting gear cannot hold (a Thick kit on an item below level 25)
	// comes off the same way it would on a swap.
	for slot, item := range best {
		if item != nil && item.Enchant != 0 {
			if e := s.carriedEnchant(slot, item.Id, item.Enchant); e != item.Enchant {
				fmt.Printf("%s: %s is not usable on this item at this level, now %s\n", gearSlotNames[slot], enchantName(item.Enchant), enchantName(e))
				item.Enchant = e
			}
		}
	}
	names := make([]string, len(best))
	unsourced := make([]bool, len(best))
	for slot, item := range best {
		if item != nil {
			for _, o := range pool[slot] {
				if o.spec.Id == item.Id && o.spec.RandomSuffix == item.RandomSuffix {
					names[slot], unsourced[slot] = o.name, o.unsourced
				}
			}
		}
	}
	s.setGear(best)
	bestScore := s.score(knobs)
	fmt.Printf("start: %s\n", bestScore)
	// The off hand a two hander pushed out, so a one hander tried later is scored with it
	// back rather than with an empty off hand.
	var heldOffHand *proto.ItemSpec
	var heldOffHandName string

	for round := 1; ; round++ {
		improved := false
		for slot, options := range pool {
			if len(options) == 0 {
				continue
			}
			current := best[slot]
			// The gear with option o in this slot, or nil when o cannot go there.
			candidateWith := func(o gearOption) []*proto.ItemSpec {
				if current != nil && o.spec.Id == current.Id && o.spec.RandomSuffix == current.RandomSuffix {
					return nil
				}
				// Rings and trinkets are mostly unique, so the pair never holds the same item.
				if p := pairedSlot(slot); p >= 0 && best[p] != nil && best[p].Id == o.spec.Id {
					return nil
				}
				if p := weaponPair(slot); o.once && p >= 0 && best[p] != nil && best[p].Id == o.spec.Id {
					return nil
				}
				// A two hander empties the off hand, and nothing goes in the off hand next
				// to one. The one hander plus shield answer is reached through the main
				// hand slot first.
				if proto.ItemSlot(slot) == proto.ItemSlot_ItemSlotOffHand && isTwoHand(pool, best[proto.ItemSlot_ItemSlotMainHand]) {
					return nil
				}
				candidate := append([]*proto.ItemSpec(nil), best...)
				candidate[slot] = s.withEnchantOf(slot, o.spec, current)
				if o.twoHand {
					candidate[proto.ItemSlot_ItemSlotOffHand] = nil
				} else if proto.ItemSlot(slot) == proto.ItemSlot_ItemSlotMainHand && candidate[proto.ItemSlot_ItemSlotOffHand] == nil {
					candidate[proto.ItemSlot_ItemSlotOffHand] = heldOffHand
				}
				return candidate
			}
			options = s.screenGear(options, candidateWith, knobs)
			var bestOption *gearOption
			bestGain := 0.0
			for i := range options {
				o := options[i]
				candidate := candidateWith(o)
				if candidate == nil {
					continue
				}
				s.setGear(candidate)
				score := s.score(knobs)
				gain := s.objective(score) - s.objective(bestScore)
				noise := math.Sqrt(score.stderr*score.stderr + bestScore.stderr*bestScore.stderr)
				if gain > s.confidence*noise && gain > bestGain {
					bestGain, bestOption = gain, &o
				}
			}
			if bestOption != nil {
				best[slot] = s.withEnchantOf(slot, bestOption.spec, current)
				names[slot], unsourced[slot] = bestOption.name, bestOption.unsourced
				if bestOption.twoHand {
					if best[proto.ItemSlot_ItemSlotOffHand] != nil {
						heldOffHand, heldOffHandName = best[proto.ItemSlot_ItemSlotOffHand], names[proto.ItemSlot_ItemSlotOffHand]
					}
					best[proto.ItemSlot_ItemSlotOffHand] = nil
					names[proto.ItemSlot_ItemSlotOffHand] = ""
				} else if proto.ItemSlot(slot) == proto.ItemSlot_ItemSlotMainHand && best[proto.ItemSlot_ItemSlotOffHand] == nil {
					best[proto.ItemSlot_ItemSlotOffHand] = heldOffHand
					names[proto.ItemSlot_ItemSlotOffHand] = heldOffHandName
				}
				s.setGear(best)
				bestScore = s.score(knobs)
				improved = true
				fmt.Printf("round %d: %s = %s (%+.1f), now %s\n", round, gearSlotNames[slot], bestOption.name, bestGain, bestScore)
			}
		}
		if s.enchantSearch && s.searchEnchants(best, &bestScore, knobs, s.enchantExclude, round) {
			improved = true
		}
		if !improved {
			break
		}
	}
	s.setGear(best)
	fmt.Println("gear:")
	for slot, item := range best {
		if item != nil {
			fmt.Printf("  %-9s %s (%d", gearSlotNames[slot], names[slot], item.Id)
			if item.RandomSuffix != 0 {
				fmt.Printf(", suffix %d", item.RandomSuffix)
			}
			fmt.Print(")")
			if item.Enchant != 0 {
				fmt.Printf(", %s", enchantName(item.Enchant))
			}
			if unsourced[slot] {
				fmt.Print("  no source wowhead knows, check it by hand")
			}
			fmt.Println()
		}
	}
	return best, bestScore
}

// craftedProfession is the profession that makes the item, or ProfessionUnknown when no
// profession does.
func craftedProfession(item *proto.UIItem) proto.Profession {
	for _, source := range item.Sources {
		if crafted := source.GetCrafted(); crafted != nil {
			return crafted.Profession
		}
	}
	return proto.Profession_ProfessionUnknown
}
