package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"

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
	proto.Class_ClassRogue:   {proto.WeaponType_WeaponTypeDagger, proto.WeaponType_WeaponTypeSword, proto.WeaponType_WeaponTypeMace, proto.WeaponType_WeaponTypeFist},
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

// Every option for every slot the player can use at their level.
func loadGearPool(dbPath, levelsPath string, player *proto.Player, level int32, minQuality proto.ItemQuality, maxPhase int32, exclude []int32) ([][]gearOption, error) {
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
		if len(item.RandomSuffixOptions) == 0 {
			pool[slot] = append(pool[slot], gearOption{&proto.ItemSpec{Id: item.Id}, item.Name, twoHand})
			return
		}
		for _, id := range item.RandomSuffixOptions {
			if s, ok := suffixes[id]; ok && usefulSuffix(class, s) {
				pool[slot] = append(pool[slot], gearOption{&proto.ItemSpec{Id: item.Id, RandomSuffix: id}, item.Name + " " + s.Name, twoHand})
			}
		}
	}

	for _, item := range db.Items {
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
		if contains(exclude, item.Id) {
			continue
		}
		// A few items have no level of their own (Olmann Sewar, the engineering goggles),
		// the quest or profession is the gate. A quest's reward sits a few item levels above
		// the quest level, so that stands in.
		if item.RequiredLevel == 0 && item.Ilvl > level+6 {
			continue
		}
		// PvP rank gear is a separate grind, it stays out of the pool.
		if pvpRank[item.Id] {
			continue
		}
		if len(item.ClassAllowlist) > 0 && !contains(item.ClassAllowlist, class) {
			continue
		}
		// Engineering goggles and Forever's bind on pickup crafted sets need the profession
		// on the character. The player's two professions come from the settings or the
		// -professions flag.
		if item.RequiredProfession != proto.Profession_ProfessionUnknown &&
			item.RequiredProfession != player.Profession1 && item.RequiredProfession != player.Profession2 {
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

// The candidate with the enchant the slot has now, so a swap is not also a lost enchant.
func withEnchantOf(spec, current *proto.ItemSpec) *proto.ItemSpec {
	if current == nil || current.Enchant == 0 {
		return spec
	}
	out := goproto.Clone(spec).(*proto.ItemSpec)
	out.Enchant = current.Enchant
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

func (s *searcher) searchGear(pool [][]gearOption, knobs Knobs) ([]*proto.ItemSpec, result) {
	best := equipmentSlots(s.setup.player().Equipment)
	names := make([]string, len(best))
	for slot, item := range best {
		if item != nil {
			for _, o := range pool[slot] {
				if o.spec.Id == item.Id && o.spec.RandomSuffix == item.RandomSuffix {
					names[slot] = o.name
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
			var bestOption *gearOption
			bestGain := 0.0
			for i := range options {
				o := options[i]
				if current != nil && o.spec.Id == current.Id && o.spec.RandomSuffix == current.RandomSuffix {
					continue
				}
				// Rings and trinkets are mostly unique, so the pair never holds the same item.
				if p := pairedSlot(slot); p >= 0 && best[p] != nil && best[p].Id == o.spec.Id {
					continue
				}
				// A two hander empties the off hand, and nothing goes in the off hand next
				// to one. The one hander plus shield answer is reached through the main
				// hand slot first.
				if proto.ItemSlot(slot) == proto.ItemSlot_ItemSlotOffHand && isTwoHand(pool, best[proto.ItemSlot_ItemSlotMainHand]) {
					continue
				}
				candidate := append([]*proto.ItemSpec(nil), best...)
				candidate[slot] = withEnchantOf(o.spec, current)
				if o.twoHand {
					candidate[proto.ItemSlot_ItemSlotOffHand] = nil
				} else if proto.ItemSlot(slot) == proto.ItemSlot_ItemSlotMainHand && candidate[proto.ItemSlot_ItemSlotOffHand] == nil {
					candidate[proto.ItemSlot_ItemSlotOffHand] = heldOffHand
				}
				s.setGear(candidate)
				score := s.score(knobs)
				gain := score.dps - bestScore.dps
				noise := math.Sqrt(score.stderr*score.stderr + bestScore.stderr*bestScore.stderr)
				if gain > s.confidence*noise && gain > bestGain {
					bestGain, bestOption = gain, &o
				}
			}
			if bestOption != nil {
				best[slot] = withEnchantOf(bestOption.spec, current)
				names[slot] = bestOption.name
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
			fmt.Println(")")
		}
	}
	return best, bestScore
}
