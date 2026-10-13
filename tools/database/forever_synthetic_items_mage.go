package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The synthetic level 60 Phase 1 mage sets, made the same way as the enhancement set in
// forever_synthetic_items.go (read that comment first). There are two sets. The main set
// (996001 to 996016) serves Frostfire, Fire and Frost. The Arcane set (996017 to 996030) is the
// same slots with Arcane's own stat mix. The Era Phase 1 BiS was one set for all three specs
// (Arcanist, Robe of Volatile Power, Mana Igniting Cord, Netherwind Pants, Azuresong Mageblade,
// Fire Runed Grimoire), none of it school specific, so both sets carry plain spell power and no
// school power.
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
// like every other slot. The off hand is at 71, so the two weapon slots come to about 73. The
// staff takes both budgets, 73.7, and the presets use it with Forever's 2H Spellblasting (+65
// spell power). The main hand and off hand split the staff's stats between them. The wand is
// 0.32 of the value at 71, 15.6. Weapon damage and speed are the Era items' (Azuresong, Staff of
// Dominance, Bonecreeper Stylus) and are not charged to the budget.
//
// Per slot, budget and the points the stats below add up to (main set, then Arcane):
//   head 42.8 (42.9, 42.9), neck 27.5 (27.7, 27.5), shoulders 33.0 (33.0, 32.7), back 27.5
//   (27.6, 27.7), chest 42.8 (42.9, 42.9), wrist 27.5 (27.2, 27.2), hands 33.0 (33.0, 32.7),
//   waist 37.6 (37.6, 37.6), legs 42.8 (42.9, 42.9), feet 33.0 (32.9, 33.2), finger 27.5 (27.2,
//   27.5), second finger 27.5 (27.4, 27.7), staff 73.7 (73.6, 73.7), wand 15.6 (15.6, 15.4),
//   main hand 46.2 (46.3), off hand 27.5 (27.6).
// For example the main set's head: 12^1.5 + 10^1.5 + 10^1.5 + (26 * 0.55)^1.5 + 14^1.5 + 17^1.5
// = 42.9^1.5.
//
// The stat mix follows the standards' fair mix rule (item 5a): a realistic spread of stats in the
// proportions Season of Discovery's level 60 BiS mage sets carry (orchestration/gear_targets.md),
// with Stamina at 18% and Spirit at 10%, not everything in spell power. Shares are linear: a
// stat's budget weight times its amount over the set's total (the twelve armor slots, the staff
// and the wand; enchants left out). Main set targets are the averaged SoD fire and frost sets,
// Arcane's the SoD arcane sets, school power counted as spell power:
//
//   set      share     Sta   Int   Spi   SP    Hit   Crit
//   main     target    18    14.7  10    21.3  12.8  23.3
//   main     actual    18.2  14.8  10.0  21.6  12.1  23.3   (151, 123, 83, 325 SP, 100, 193)
//   Arcane   target    18    16.1  10    26.6  12.8  16.1
//   Arcane   actual    18.4  16.3  10.1  26.9  12.1  16.2   (152, 135, 84, 405 SP, 100, 134)
//
// Hit stays at the cap against a level 63 boss (16% spell miss) in both sets: 100 hit rating
// (10%), with 5% from Elemental Precision or Arcane Focus and Manaflare's 2 piece 1%.
//
// Only the head and chest carry the Manaflare Regalia set (Forever's level 60 mage set, see
// sim/mage/item_sets_pve.go), in both sets, so only its 2 piece bonus is live, the same for
// every class's preset. The generator script lived in the session scratchpad (rb/gen.py,
// rb/run.py), the numbers below are its output.

func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, foreverSyntheticMageItems...)
}

var foreverSyntheticMageItems = []*proto.UIItem{
	// Head, item level 66, in place of Arcanist Crown (66, 16795).
	{
		Id: 996001, Name: "Mage Synthetic Head (Phase 1)", Icon: "inv_helmet_53",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 12, stats.Intellect: 10, stats.Spirit: 10, stats.SpellPower: 26, stats.Armor: 83}.ToFloatArray(),
		HitRating:  14,
		CritRating: 17,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
		SetName:        "Mage Synthetic (Phase 1)",
	},
	// Neck, item level 71, in place of Choker of the Fire Lord (78, 18814).
	{
		Id: 996002, Name: "Mage Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_17",
		Type:       proto.ItemType_ItemTypeNeck,
		Stats:      stats.Stats{stats.Stamina: 10, stats.Intellect: 8, stats.SpellPower: 21}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Shoulder, item level 66, in place of Arcanist Mantle (66, 16797).
	{
		Id: 996003, Name: "Mage Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_02",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 10, stats.Intellect: 9, stats.SpellPower: 22, stats.Armor: 76}.ToFloatArray(),
		HitRating:  11,
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Back, item level 71.
	{
		Id: 996004, Name: "Mage Synthetic Back (Phase 1)", Icon: "inv_misc_cape_16",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 9, stats.Intellect: 7, stats.Spirit: 7, stats.SpellPower: 19, stats.Armor: 54}.ToFloatArray(),
		CritRating: 13,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Chest, item level 66, in place of Robe of Volatile Power (66, 19145).
	{
		Id: 996005, Name: "Mage Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_03",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 12, stats.Intellect: 10, stats.Spirit: 10, stats.SpellPower: 26, stats.Armor: 102}.ToFloatArray(),
		HitRating:  14,
		CritRating: 17,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
		SetName:        "Mage Synthetic (Phase 1)",
	},
	// Wrist, item level 71, in place of Arcanist Bindings (66, 16799).
	{
		Id: 996006, Name: "Mage Synthetic Wrist (Phase 1)", Icon: "inv_bracer_09",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 11, stats.Intellect: 9, stats.Spirit: 9, stats.SpellPower: 25, stats.Armor: 44}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Hands, item level 66, in place of Arcanist Gloves (66, 16801).
	{
		Id: 996007, Name: "Mage Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_14",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 10, stats.Intellect: 9, stats.SpellPower: 22, stats.Armor: 63}.ToFloatArray(),
		HitRating:  11,
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Waist, item level 71, in place of Mana Igniting Cord (71, 19136).
	{
		Id: 996008, Name: "Mage Synthetic Waist (Phase 1)", Icon: "inv_belt_22",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 12, stats.Intellect: 9, stats.SpellPower: 25, stats.Armor: 61}.ToFloatArray(),
		HitRating:  13,
		CritRating: 16,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Legs, item level 66, in place of Netherwind Pants (76, 16915).
	{
		Id: 996009, Name: "Mage Synthetic Legs (Phase 1)", Icon: "inv_pants_08",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 12, stats.Intellect: 10, stats.Spirit: 10, stats.SpellPower: 26, stats.Armor: 101}.ToFloatArray(),
		HitRating:  14,
		CritRating: 17,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Feet, item level 66, in place of Arcanist Boots (66, 16800).
	{
		Id: 996010, Name: "Mage Synthetic Feet (Phase 1)", Icon: "inv_boots_07",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 11, stats.Intellect: 9, stats.Spirit: 9, stats.SpellPower: 25, stats.Armor: 70}.ToFloatArray(),
		HitRating: 13,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Finger, item level 71, in place of Ring of Spell Power (66, 19147).
	{
		Id: 996011, Name: "Mage Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_38",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Stamina: 9, stats.SpellPower: 20}.ToFloatArray(),
		HitRating:  10,
		CritRating: 13,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Finger II, item level 71, in place of Cauterizing Band (71, 19140).
	{
		Id: 996012, Name: "Mage Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_39",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Stamina: 10, stats.Intellect: 9, stats.Spirit: 9}.ToFloatArray(),
		CritRating: 15,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Main Hand, item level 71, in place of Azuresong Mageblade (71, 17103).
	{
		Id: 996013, Name: "Mage Synthetic Main Hand (Phase 1)", Icon: "inv_sword_39",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeMainHand,
		Stats:      stats.Stats{stats.Stamina: 13, stats.Intellect: 10, stats.SpellPower: 29}.ToFloatArray(),
		CritRating: 31,
		// Azuresong Mageblade damage and speed.
		WeaponDamageMin: 64, WeaponDamageMax: 140, WeaponSpeed: 2.4,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Off Hand, item level 71, in place of Fire Runed Grimoire (70, 19142).
	{
		Id: 996014, Name: "Mage Synthetic Off Hand (Phase 1)", Icon: "inv_misc_book_09",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeOffHand, HandType: proto.HandType_HandTypeOffHand,
		Stats: stats.Stats{stats.Stamina: 8, stats.Intellect: 7, stats.Spirit: 17, stats.SpellPower: 19}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Wand, item level 71, in place of Bonecreeper Stylus (62, 13938).
	{
		Id: 996015, Name: "Mage Synthetic Wand (Phase 1)", Icon: "inv_wand_07",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeWand,
		Stats:      stats.Stats{stats.Intellect: 5, stats.SpellPower: 15}.ToFloatArray(),
		CritRating: 9,
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
		Stats:      stats.Stats{stats.Stamina: 23, stats.Intellect: 19, stats.Spirit: 19, stats.SpellPower: 53}.ToFloatArray(),
		CritRating: 34,
		// Staff of Dominance damage and speed.
		WeaponDamageMin: 126, WeaponDamageMax: 205, WeaponSpeed: 2.9,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: head, item level 66, in place of Arcanist Crown (66, 16795).
	{
		Id: 996017, Name: "Mage Arcane Synthetic Head (Phase 1)", Icon: "inv_helmet_53",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 12, stats.Intellect: 11, stats.Spirit: 10, stats.SpellPower: 33, stats.Armor: 83}.ToFloatArray(),
		HitRating:  14,
		CritRating: 12,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
		SetName:        "Mage Synthetic (Phase 1)",
	},
	// Arcane: neck, item level 71, in place of Choker of the Fire Lord (78, 18814).
	{
		Id: 996018, Name: "Mage Arcane Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_17",
		Type:       proto.ItemType_ItemTypeNeck,
		Stats:      stats.Stats{stats.Stamina: 10, stats.Intellect: 9, stats.SpellPower: 26}.ToFloatArray(),
		CritRating: 10,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: shoulder, item level 66, in place of Arcanist Mantle (66, 16797).
	{
		Id: 996019, Name: "Mage Arcane Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_02",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 10, stats.Intellect: 9, stats.SpellPower: 28, stats.Armor: 76}.ToFloatArray(),
		HitRating:  11,
		CritRating: 10,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: back, item level 71.
	{
		Id: 996020, Name: "Mage Arcane Synthetic Back (Phase 1)", Icon: "inv_misc_cape_16",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 9, stats.Intellect: 8, stats.Spirit: 7, stats.SpellPower: 25, stats.Armor: 54}.ToFloatArray(),
		CritRating: 9,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: chest, item level 66, in place of Robe of Volatile Power (66, 19145).
	{
		Id: 996021, Name: "Mage Arcane Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_03",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 12, stats.Intellect: 11, stats.Spirit: 10, stats.SpellPower: 33, stats.Armor: 102}.ToFloatArray(),
		HitRating:  14,
		CritRating: 12,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
		SetName:        "Mage Synthetic (Phase 1)",
	},
	// Arcane: wrist, item level 71, in place of Arcanist Bindings (66, 16799).
	{
		Id: 996022, Name: "Mage Arcane Synthetic Wrist (Phase 1)", Icon: "inv_bracer_09",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 10, stats.Intellect: 9, stats.Spirit: 8, stats.SpellPower: 28, stats.Armor: 44}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: hands, item level 66, in place of Arcanist Gloves (66, 16801).
	{
		Id: 996023, Name: "Mage Arcane Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_14",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 10, stats.Intellect: 9, stats.SpellPower: 28, stats.Armor: 63}.ToFloatArray(),
		HitRating:  11,
		CritRating: 10,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: waist, item level 71, in place of Mana Igniting Cord (71, 19136).
	{
		Id: 996024, Name: "Mage Arcane Synthetic Waist (Phase 1)", Icon: "inv_belt_22",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 12, stats.Intellect: 10, stats.SpellPower: 32, stats.Armor: 61}.ToFloatArray(),
		HitRating:  13,
		CritRating: 11,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: legs, item level 66, in place of Netherwind Pants (76, 16915).
	{
		Id: 996025, Name: "Mage Arcane Synthetic Legs (Phase 1)", Icon: "inv_pants_08",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 12, stats.Intellect: 11, stats.Spirit: 10, stats.SpellPower: 33, stats.Armor: 101}.ToFloatArray(),
		HitRating:  14,
		CritRating: 12,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: feet, item level 66, in place of Arcanist Boots (66, 16800).
	{
		Id: 996026, Name: "Mage Arcane Synthetic Feet (Phase 1)", Icon: "inv_boots_07",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 10, stats.Intellect: 9, stats.Spirit: 9, stats.SpellPower: 29, stats.Armor: 70}.ToFloatArray(),
		HitRating: 12,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: finger, item level 71, in place of Ring of Spell Power (66, 19147).
	{
		Id: 996027, Name: "Mage Arcane Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_38",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Stamina: 9, stats.SpellPower: 26}.ToFloatArray(),
		HitRating:  11,
		CritRating: 9,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: finger II, item level 71, in place of Cauterizing Band (71, 19140).
	{
		Id: 996028, Name: "Mage Arcane Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_39",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Stamina: 12, stats.Intellect: 11, stats.Spirit: 10}.ToFloatArray(),
		CritRating: 11,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: staff, the main hand and off hand budgets together (46.2 + 27.5), in place of Staff of
	// Dominance (70, 18842), for the presets that sim better with a two-hander.
	{
		Id: 996029, Name: "Mage Arcane Synthetic Staff (Phase 1)", Icon: "inv_staff_13",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand,
		Stats:      stats.Stats{stats.Stamina: 24, stats.Intellect: 22, stats.Spirit: 20, stats.SpellPower: 66}.ToFloatArray(),
		CritRating: 22,
		// Staff of Dominance damage and speed.
		WeaponDamageMin: 126, WeaponDamageMax: 205, WeaponSpeed: 2.9,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: wand, item level 71, in place of Bonecreeper Stylus (62, 13938).
	{
		Id: 996030, Name: "Mage Arcane Synthetic Wand (Phase 1)", Icon: "inv_wand_07",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeWand,
		Stats:      stats.Stats{stats.Intellect: 6, stats.SpellPower: 18}.ToFloatArray(),
		CritRating: 6,
		// Bonecreeper Stylus damage and speed.
		WeaponDamageMin: 83, WeaponDamageMax: 155, WeaponSpeed: 1.9,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
}
