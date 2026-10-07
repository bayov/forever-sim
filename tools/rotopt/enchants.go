package main

import (
	"fmt"
	"math"
	"slices"

	"github.com/wowsims/classic/sim/core/proto"
	goproto "google.golang.org/protobuf/proto"
)

// Enchant search. After each round of the item search, every enchantable slot tries each
// enchant it can take, the same way a slot tries items.
//
// The enchants are what an enchanter at Enchanting 225 can make (see
// tools/database/forever_enchants.go and the enchants https://foreverchanges.pro picks
// for its level 20 BiS lists). Artisan Enchanting, the training past 225, needs level 35,
// so the list is right for searches up to level 34. The armor kits carry their own
// character level (Heavy kits from 30), which the search checks. Enchants that only
// change speed or resistances are left out.
var (
	neckEnchants = []int32{
		1249059, // Enchant Necklace - Agility (+5)
		1249019, // Enchant Necklace - Strength (+5)
		1249057, // Enchant Necklace - Spell Power (+6)
	}
	backEnchants = []int32{
		13419, // Enchant Cloak - Minor Agility (+3 under Forever)
	}
	chestEnchants = []int32{
		866,   // Enchant Chest - Lesser Stats (+2)
		13858, // Enchant Chest - Superior Stamina (+8 under Forever, for -min-health)
	}
	wristEnchants = []int32{
		1248497, // Enchant Bracer - Agility (+5)
		13661,   // Enchant Bracer - Strength (+5)
		13622,   // Enchant Bracer - Lesser Intellect (+5)
		1248498, // Enchant Bracer - Lesser Healing Power (6 spell power)
		852,     // Enchant Bracer - Stamina (+5, for -min-health)
	}
	handsEnchants = []int32{
		13815, // Enchant Gloves - Agility (+7)
		13887, // Enchant Gloves - Strength (+7)
	}
	feetEnchants = []int32{
		13637, // Enchant Boots - Lesser Agility (+4)
		852,   // Enchant Boots - Stamina (+5, for -min-health)
	}
	// Lesser Strength (+15) and Lesser Agility (+15) for two-handers are not in the beta (user,
	// 2026-10-05), so they are out.
	oneHandEnchants = []int32{
		943,     // Enchant Weapon - Striking (+3 damage)
		21931,   // Enchant Weapon - Winter's Might (+15 frost spell power)
		1248805, // Enchant Weapon - Revelation (spell crit proc, -revelation-chance)
	}
	twoHandEnchants = []int32{
		13695,   // Enchant 2H Weapon - Impact (+6 damage under Forever)
		7793,    // Enchant 2H Weapon - Lesser Intellect (+5)
		21931,   // Enchant Weapon - Winter's Might
		1248805, // Enchant Weapon - Revelation
	}

	// Armor kits go on the chest, hands, legs and feet. A Heavy kit needs an item of level
	// 15 or above and a Thick kit one of level 25 or above. They also need a character
	// level (15 for the Medium kits, 30 for Heavy, 40 for Thick), which enchantLevels
	// holds from the database, so a level 20 only gets the Medium ones.
	//
	// The plain kits carry Stamina under Forever, and the plain Thick kit only needs level
	// 30. They never win on damage, but a search with a health floor (-min-health) wants
	// them.
	mediumKits = []int32{
		1255123, // Forceful Medium Armor Kit (4 attack power)
		1255124, // Mystic Medium Armor Kit (2 spell power)
		16,      // Medium Armor Kit (2 Stamina)
	}
	heavyKits = []int32{
		1255095, // Forceful Heavy Armor Kit (6 attack power)
		1255096, // Mystic Heavy Armor Kit (4 spell power)
		17,      // Heavy Armor Kit (3 Stamina)
	}
	thickKits = []int32{
		1255073, // Forceful Thick Armor Kit (8 attack power)
		1255074, // Mystic Thick Armor Kit (5 spell power)
		18,      // Thick Armor Kit (4 Stamina)
	}
)

// The next kit down in the same line, for an item or a character too low for a kit.
var smallerKit = map[int32]int32{
	1255073: 1255095, 1255095: 1255123,
	1255074: 1255096, 1255096: 1255124,
	18: 17, 17: 16,
}

// Every item and enchant in the UI database, for the enchant rules that depend on the
// item and for the names in the output.
var (
	itemsByID     = map[int32]*proto.UIItem{}
	enchantNames  = map[int32]string{}
	enchantLevels = map[int32]int32{}
)

func enchantName(id int32) string {
	if id == 0 {
		return "no enchant"
	}
	if name, ok := enchantNames[id]; ok {
		return name
	}
	return fmt.Sprint(id)
}

// The enchant a new item in the slot keeps from the item it replaces.
//
// A kit the item or the character is too low for steps down to the next smaller kit of
// the same line, or comes off. With the enchant search on, any enchant the new item
// cannot take comes off, and the enchant search puts the best one on. Without it the
// list above does not apply (a level 60 search has enchants it does not know), so only
// the kit rules hold.
func (s *searcher) carriedEnchant(slot int, itemID, enchant int32) int32 {
	item := itemsByID[itemID]
	for item != nil && enchant != 0 && !kitFits(item, enchant, s.level) {
		enchant = smallerKit[enchant]
	}
	if s.enchantSearch && !enchantFits(slot, itemID, enchant, s.level) {
		return 0
	}
	return enchant
}

// Whether a kit goes on the item for a character of the level. Anything that is not a
// kit does.
func kitFits(item *proto.UIItem, enchant int32, level int32) bool {
	switch {
	case slices.Contains(heavyKits, enchant) && item.Ilvl < 15:
		return false
	case slices.Contains(thickKits, enchant) && item.Ilvl < 25:
		return false
	}
	return enchantLevels[enchant] <= level
}

// The enchants an item in a slot can take at the level, without the ones in exclude.
func enchantOptions(slot int, itemID int32, exclude []int32, level int32) []int32 {
	item := itemsByID[itemID]
	if item == nil {
		return nil
	}
	var kits []int32
	for _, kit := range slices.Concat(mediumKits, heavyKits, thickKits) {
		if kitFits(item, kit, level) {
			kits = append(kits, kit)
		}
	}
	var options []int32
	switch proto.ItemSlot(slot) {
	case proto.ItemSlot_ItemSlotNeck:
		options = neckEnchants
	case proto.ItemSlot_ItemSlotBack:
		options = backEnchants
	case proto.ItemSlot_ItemSlotChest:
		options = append(slices.Clone(chestEnchants), kits...)
	case proto.ItemSlot_ItemSlotWrist:
		options = wristEnchants
	case proto.ItemSlot_ItemSlotHands:
		options = append(slices.Clone(handsEnchants), kits...)
	case proto.ItemSlot_ItemSlotLegs:
		options = kits
	case proto.ItemSlot_ItemSlotFeet:
		options = append(slices.Clone(feetEnchants), kits...)
	case proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand:
		// Shields and held in off hand items take no weapon enchant.
		if item.Type != proto.ItemType_ItemTypeWeapon || item.WeaponType == proto.WeaponType_WeaponTypeShield || item.WeaponType == proto.WeaponType_WeaponTypeOffHand {
			return nil
		}
		if item.HandType == proto.HandType_HandTypeTwoHand {
			options = twoHandEnchants
		} else {
			options = oneHandEnchants
		}
	}
	var out []int32
	for _, e := range options {
		if !slices.Contains(exclude, e) {
			out = append(out, e)
		}
	}
	return out
}

// Whether the enchant can go on the item in the slot. 0 (no enchant) always can.
func enchantFits(slot int, itemID int32, enchant int32, level int32) bool {
	return enchant == 0 || slices.Contains(enchantOptions(slot, itemID, nil, level), enchant)
}

// searchEnchants tries every enchant in every slot once, keeping the best one when it
// beats the current enchant by more than the noise. It reports whether anything changed.
func (s *searcher) searchEnchants(best []*proto.ItemSpec, bestScore *result, knobs Knobs, exclude []int32, round int) bool {
	improved := false
	for slot, item := range best {
		if item == nil {
			continue
		}
		options := append([]int32{0}, enchantOptions(slot, item.Id, exclude, s.level)...)
		if len(options) == 1 {
			continue
		}
		bestEnchant, bestGain := int32(-1), 0.0
		var bestEnchantScore result
		for _, e := range options {
			if e == item.Enchant {
				continue
			}
			candidate := append([]*proto.ItemSpec(nil), best...)
			candidate[slot] = goproto.Clone(item).(*proto.ItemSpec)
			candidate[slot].Enchant = e
			s.setGear(candidate)
			score := s.score(knobs)
			gain := s.objective(score) - s.objective(*bestScore)
			noise := math.Sqrt(score.stderr*score.stderr + bestScore.stderr*bestScore.stderr)
			if gain > s.confidence*noise && gain > bestGain {
				bestEnchant, bestGain, bestEnchantScore = e, gain, score
			}
		}
		if bestEnchant >= 0 {
			best[slot] = goproto.Clone(item).(*proto.ItemSpec)
			best[slot].Enchant = bestEnchant
			*bestScore = bestEnchantScore
			improved = true
			fmt.Printf("round %d: %s enchant = %s (%+.1f), now %s\n", round, gearSlotNames[slot], enchantName(bestEnchant), bestGain, bestEnchantScore)
		}
	}
	s.setGear(best)
	return improved
}
