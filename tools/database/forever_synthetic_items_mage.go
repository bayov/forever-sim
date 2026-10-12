package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The synthetic level 60 Phase 1 mage set, made the same way as the enhancement set in
// forever_synthetic_items.go (read that comment first). One set serves Arcane, Fire and Frost: the
// three specs wore the same Era Phase 1 BiS (Arcanist, Robe of Volatile Power, Mana Igniting Cord,
// Netherwind Pants, Azuresong Mageblade and Fire Runed Grimoire, none of it school specific), and
// they want the same stats, spell power first, then hit and crit, then Intellect and Spirit for the
// mana. So the set carries plain spell power and no school power.
//
// Each piece gets the full epic budget of its slot, (item level - 1.3) / 1.3 times the Classic slot
// modifier (1 for head, chest and legs, 0.77 for shoulders, hands, waist and feet, 0.56 for neck,
// back, wrist, finger and off hand). For the one-hand caster weapon we fit the modifier to
// Azuresong Mageblade itself (0.79, its Era stats at Era's 0.75 a point of spell power). The points
// add up the Classic way, (sum of points^1.5)^(2/3), with spell power at 0.55 a point, and hit and
// crit rating at 1 a point (10 rating is 1% hit, 14 rating 1% crit). The item level is the Era
// Phase 1 BiS item's, or Molten Core tier's 66 where that is higher. Era Phase 1 has no caster
// cloak above 63, so the cloak takes the tier's 66. For example the head: (66 - 1.3) / 1.3 = 49.8,
// and 16^1.5 + 12^1.5 + 11^1.5 + (56 * 0.55)^1.5 + 10^1.5 = 49.2^1.5.
//
// The stat mix differs by slot:
//   - Spirit is on 8 of the 14 pieces, 10.2% of the set's budget (83 Spirit), because Spirit should
//     matter again under Forever.
//   - Spell power is 40% of the budget (588). The first ring and the main hand lean on it, the
//     second ring and the off hand on Intellect and Spirit.
//   - The neck, waist, hands and the first ring have no Spirit, the wrist no Stamina.
//   - The set has 20 hit rating (2%, on the head and hands) and 98 crit rating (7%, on six pieces),
//     about what the Era Phase 1 BiS carries (Arcanist Crown's 1% hit, the robe's 2% crit and 1%
//     on the belt, legs, boots and sword). Manaflare's 2 piece bonus adds 1% hit.
//
// The head, shoulders, chest, hands, legs and feet form a set with the bonuses of the Manaflare
// Regalia, Forever's level 60 mage set (see sim/mage/item_sets_pve.go). The generator script lived
// in the session scratchpad (synth_mage.py), the numbers below are its output.

func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, foreverSyntheticMageItems...)
}

var foreverSyntheticMageItems = []*proto.UIItem{
	// Head, item level 66, in place of Arcanist Crown (66, 16795).
	{
		Id: 996001, Name: "Mage Synthetic Head (Phase 1)", Icon: "inv_helmet_53",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 12, stats.Intellect: 16, stats.Spirit: 11, stats.SpellPower: 56, stats.Armor: 83}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
		SetName:        "Mage Synthetic (Phase 1)",
	},
	// Neck, item level 78, in place of Choker of the Fire Lord (78, 18814).
	{
		Id: 996002, Name: "Mage Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_17",
		Type:  proto.ItemType_ItemTypeNeck,
		Stats: stats.Stats{stats.Stamina: 7, stats.Intellect: 9, stats.SpellPower: 49}.ToFloatArray(),
		Ilvl:  78, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Shoulder, item level 66, in place of Arcanist Mantle (66, 16797).
	{
		Id: 996003, Name: "Mage Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_02",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 9, stats.Intellect: 13, stats.SpellPower: 42, stats.Armor: 76}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
		SetName:        "Mage Synthetic (Phase 1)",
	},
	// Back, item level 66.
	{
		Id: 996004, Name: "Mage Synthetic Back (Phase 1)", Icon: "inv_misc_cape_16",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 7, stats.Intellect: 9, stats.Spirit: 9, stats.SpellPower: 33, stats.Armor: 54}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Chest, item level 66, in place of Robe of Volatile Power (66, 19145).
	{
		Id: 996005, Name: "Mage Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_03",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 10, stats.Intellect: 14, stats.Spirit: 10, stats.SpellPower: 36, stats.Armor: 102}.ToFloatArray(),
		CritRating: 28,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
		SetName:        "Mage Synthetic (Phase 1)",
	},
	// Wrist, item level 66, in place of Arcanist Bindings (66, 16799).
	{
		Id: 996006, Name: "Mage Synthetic Wrist (Phase 1)", Icon: "inv_bracer_09",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Intellect: 12, stats.Spirit: 9, stats.SpellPower: 34, stats.Armor: 44}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Hands, item level 66, in place of Arcanist Gloves (66, 16801).
	{
		Id: 996007, Name: "Mage Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_14",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 9, stats.Intellect: 14, stats.SpellPower: 46, stats.Armor: 63}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
		SetName:        "Mage Synthetic (Phase 1)",
	},
	// Waist, item level 71, in place of Mana Igniting Cord (71, 19136).
	{
		Id: 996008, Name: "Mage Synthetic Waist (Phase 1)", Icon: "inv_belt_22",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 11, stats.Intellect: 13, stats.SpellPower: 47, stats.Armor: 61}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Legs, item level 76, in place of Netherwind Pants (76, 16915).
	{
		Id: 996009, Name: "Mage Synthetic Legs (Phase 1)", Icon: "inv_pants_08",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 15, stats.Intellect: 20, stats.Spirit: 13, stats.SpellPower: 60, stats.Armor: 101}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       76, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
		SetName:        "Mage Synthetic (Phase 1)",
	},
	// Feet, item level 66, in place of Arcanist Boots (66, 16800).
	{
		Id: 996010, Name: "Mage Synthetic Feet (Phase 1)", Icon: "inv_boots_07",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 12, stats.Intellect: 12, stats.Spirit: 10, stats.SpellPower: 32, stats.Armor: 70}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
		SetName:        "Mage Synthetic (Phase 1)",
	},
	// Finger, item level 66, in place of Ring of Spell Power (66, 19147).
	{
		Id: 996011, Name: "Mage Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_38",
		Type:  proto.ItemType_ItemTypeFinger,
		Stats: stats.Stats{stats.Stamina: 5, stats.SpellPower: 48}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Finger II, item level 71, in place of Cauterizing Band (71, 19140).
	{
		Id: 996012, Name: "Mage Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_39",
		Type:  proto.ItemType_ItemTypeFinger,
		Stats: stats.Stats{stats.Stamina: 8, stats.Intellect: 12, stats.Spirit: 12, stats.SpellPower: 26}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Main Hand, item level 71, in place of Azuresong Mageblade (71, 17103).
	{
		Id: 996013, Name: "Mage Synthetic Main Hand (Phase 1)", Icon: "inv_sword_39",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeMainHand,
		Stats:      stats.Stats{stats.Stamina: 7, stats.Intellect: 9, stats.SpellPower: 58}.ToFloatArray(),
		CritRating: 14,
		// Azuresong Mageblade damage and speed.
		WeaponDamageMin: 64, WeaponDamageMax: 140, WeaponSpeed: 2.4,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Off Hand, item level 70, in place of Fire Runed Grimoire (70, 19142).
	{
		Id: 996014, Name: "Mage Synthetic Off Hand (Phase 1)", Icon: "inv_misc_book_09",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeOffHand, HandType: proto.HandType_HandTypeOffHand,
		Stats: stats.Stats{stats.Stamina: 9, stats.Intellect: 16, stats.Spirit: 9, stats.SpellPower: 21}.ToFloatArray(),
		Ilvl:  70, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
}
