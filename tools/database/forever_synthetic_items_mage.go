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
// Each piece gets the full epic budget of its slot under the cross-class standards (orchestration/
// standards.md on the orchestration branch): the epic item value V = 1.2 * item level - 36.4 times
// the slot modifier (1 for head, chest and legs, 0.77 for shoulders, hands, waist and feet, 0.5625
// for neck, back, wrist, finger and off hand, 0.32 for the wand). The points add up as
// (sum of (weight * amount)^1.5)^(2/3), with 1 a point of Intellect, Stamina and Spirit, 0.55 a
// point of spell power, and 1 a point of hit and crit rating (10 rating is 1% hit, 14 rating 1%
// crit). Every class takes one item level table: head, shoulders, chest, hands, legs and feet at
// 66 (V 42.8), neck, back, wrist, waist and both rings at 71 (V 48.8).
//
// The main hand's budget is Azuresong Mageblade's own stats at Era's prices (7 Stamina, 12
// Intellect, 40 spell power at Era's 0.857 a point and 1% crit), 46.2, spent at Forever's 0.55
// like every other slot. The off hand is at 71, so the two weapon slots come to about 73, what a
// caster's staff gets. The wand is 0.32 of the value at 71, 15.6. Weapon damage and speed are
// the Era items' (Azuresong, Bonecreeper Stylus) and are not charged to the budget.
//
// Per slot, budget and the points the stats below add up to:
//   head 42.8 (43.0), neck 27.5 (27.5), shoulders 33.0 (33.0), back 27.5 (27.4), chest 42.8
//   (42.7), wrist 27.5 (27.2), hands 33.0 (32.9), waist 37.6 (37.5), legs 42.8 (42.8), feet 33.0
//   (33.2), finger 27.5 (27.6), second finger 27.5 (28.1), main hand 46.2 (45.9), off hand 27.5
//   (27.2), wand 15.6 (15.6).
// For example the head: 13^1.5 + 10^1.5 + 10^1.5 + (44 * 0.55)^1.5 + 14^1.5 = 43.0^1.5.
//
// The stat mix differs by slot:
//   - Spirit is on 9 of the 15 pieces, 10% of the set's budget (79 Spirit), because Spirit should
//     matter again under Forever.
//   - Spell power is 36% of the budget (509). The first ring and the main hand lean on it, the
//     second ring and the off hand on Intellect and Spirit.
//   - Hit is at the cap against a level 63 boss (16% spell miss): 100 hit rating (10%) on nine
//     pieces, with 5% from Elemental Precision or Arcane Focus and Manaflare's 2 piece 1%.
//   - Crit is 84 rating (6%, on six pieces), about what the Era Phase 1 BiS carries.
//
// Only the head and chest carry the Manaflare Regalia set (Forever's level 60 mage set, see
// sim/mage/item_sets_pve.go), so only its 2 piece bonus is live, the same for every class's
// preset. The generator script lived in the session scratchpad (synth_mage3.py), the numbers
// below are its output.

func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, foreverSyntheticMageItems...)
}

var foreverSyntheticMageItems = []*proto.UIItem{
	// Head, item level 66, in place of Arcanist Crown (66, 16795).
	{
		Id: 996001, Name: "Mage Synthetic Head (Phase 1)", Icon: "inv_helmet_53",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 10, stats.Intellect: 13, stats.Spirit: 10, stats.SpellPower: 44, stats.Armor: 83}.ToFloatArray(),
		HitRating: 14,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
		SetName:        "Mage Synthetic (Phase 1)",
	},
	// Neck, item level 71, in place of Choker of the Fire Lord (78, 18814).
	{
		Id: 996002, Name: "Mage Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_17",
		Type:      proto.ItemType_ItemTypeNeck,
		Stats:     stats.Stats{stats.Stamina: 5, stats.Intellect: 7, stats.SpellPower: 37}.ToFloatArray(),
		HitRating: 8,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Shoulder, item level 66, in place of Arcanist Mantle (66, 16797).
	{
		Id: 996003, Name: "Mage Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_02",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 6, stats.Intellect: 9, stats.SpellPower: 29, stats.Armor: 76}.ToFloatArray(),
		HitRating:  10,
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Back, item level 71.
	{
		Id: 996004, Name: "Mage Synthetic Back (Phase 1)", Icon: "inv_misc_cape_16",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 6, stats.Intellect: 9, stats.Spirit: 10, stats.SpellPower: 31, stats.Armor: 54}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Chest, item level 66, in place of Robe of Volatile Power (66, 19145).
	{
		Id: 996005, Name: "Mage Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_03",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 9, stats.Intellect: 12, stats.Spirit: 10, stats.SpellPower: 32, stats.Armor: 102}.ToFloatArray(),
		HitRating:  14,
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
		SetName:        "Mage Synthetic (Phase 1)",
	},
	// Wrist, item level 71, in place of Arcanist Bindings (66, 16799).
	{
		Id: 996006, Name: "Mage Synthetic Wrist (Phase 1)", Icon: "inv_bracer_09",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Intellect: 11, stats.Spirit: 10, stats.SpellPower: 32, stats.Armor: 44}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Hands, item level 66, in place of Arcanist Gloves (66, 16801).
	{
		Id: 996007, Name: "Mage Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_14",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 7, stats.Intellect: 11, stats.SpellPower: 37, stats.Armor: 63}.ToFloatArray(),
		HitRating: 12,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Waist, item level 71, in place of Mana Igniting Cord (71, 19136).
	{
		Id: 996008, Name: "Mage Synthetic Waist (Phase 1)", Icon: "inv_belt_22",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 8, stats.Intellect: 10, stats.SpellPower: 37, stats.Armor: 61}.ToFloatArray(),
		HitRating:  10,
		CritRating: 14,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Legs, item level 66, in place of Netherwind Pants (76, 16915).
	{
		Id: 996009, Name: "Mage Synthetic Legs (Phase 1)", Icon: "inv_pants_08",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 9, stats.Intellect: 12, stats.Spirit: 8, stats.SpellPower: 35, stats.Armor: 101}.ToFloatArray(),
		HitRating:  14,
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Feet, item level 66, in place of Arcanist Boots (66, 16800).
	{
		Id: 996010, Name: "Mage Synthetic Feet (Phase 1)", Icon: "inv_boots_07",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 8, stats.Intellect: 8, stats.Spirit: 8, stats.SpellPower: 21, stats.Armor: 70}.ToFloatArray(),
		HitRating:  10,
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Finger, item level 71, in place of Ring of Spell Power (66, 19147).
	{
		Id: 996011, Name: "Mage Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_38",
		Type:      proto.ItemType_ItemTypeFinger,
		Stats:     stats.Stats{stats.Stamina: 5, stats.SpellPower: 42}.ToFloatArray(),
		HitRating: 8,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Finger II, item level 71, in place of Cauterizing Band (71, 19140).
	{
		Id: 996012, Name: "Mage Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_39",
		Type:  proto.ItemType_ItemTypeFinger,
		Stats: stats.Stats{stats.Stamina: 8, stats.Intellect: 11, stats.Spirit: 12, stats.SpellPower: 24}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Main Hand, item level 71, in place of Azuresong Mageblade (71, 17103).
	{
		Id: 996013, Name: "Mage Synthetic Main Hand (Phase 1)", Icon: "inv_sword_39",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeMainHand,
		Stats:      stats.Stats{stats.Stamina: 7, stats.Intellect: 10, stats.SpellPower: 64}.ToFloatArray(),
		CritRating: 14,
		// Azuresong Mageblade damage and speed.
		WeaponDamageMin: 64, WeaponDamageMax: 140, WeaponSpeed: 2.4,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Off Hand, item level 71, in place of Fire Runed Grimoire (70, 19142).
	{
		Id: 996014, Name: "Mage Synthetic Off Hand (Phase 1)", Icon: "inv_misc_book_09",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeOffHand, HandType: proto.HandType_HandTypeOffHand,
		Stats: stats.Stats{stats.Stamina: 8, stats.Intellect: 15, stats.Spirit: 9, stats.SpellPower: 19}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Wand, item level 71, in place of Bonecreeper Stylus (62, 13938).
	{
		Id: 996015, Name: "Mage Synthetic Wand (Phase 1)", Icon: "inv_wand_07",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeWand,
		Stats: stats.Stats{stats.Intellect: 4, stats.Spirit: 2, stats.SpellPower: 25}.ToFloatArray(),
		// Bonecreeper Stylus damage and speed.
		WeaponDamageMin: 83, WeaponDamageMax: 155, WeaponSpeed: 1.9,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Staff, the main hand and off hand budgets together (46.2 + 27.5), in place of Staff of
	// Dominance (70, 18842), for the presets that sim better with a two-hander.
	{
		Id: 996016, Name: "Mage Synthetic Staff (Phase 1)", Icon: "inv_staff_13",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand,
		Stats:      stats.Stats{stats.Stamina: 13, stats.Intellect: 20, stats.Spirit: 11, stats.SpellPower: 100}.ToFloatArray(),
		CritRating: 14,
		// Staff of Dominance damage and speed.
		WeaponDamageMin: 126, WeaponDamageMax: 205, WeaponSpeed: 2.9,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
}
