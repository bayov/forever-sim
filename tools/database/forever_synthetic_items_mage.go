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
// for neck, back, wrist, finger and off hand, 0.32 for the wand). Every class takes one item level
// table: head, shoulders, chest, hands, legs and feet at 66 (V 42.8), neck, back, wrist, waist
// and both rings at 71 (V 48.8). The weapon slots come to 73 together: the main hand's 46.2 is
// Azuresong Mageblade's own stats at Era's prices, the off hand is 26.8, and the staff takes both.
// The wand is 0.32 of V(71), 15.6.
//
// Each piece is spent linearly (standards item 1a): its weighted stat total is 1.73 times its
// budget, within 1%, with 1 a point of Intellect, Stamina and Spirit, 0.55 a point of spell power
// and 1 a point of hit and crit rating (10 rating is 1% hit, 14 rating 1% crit). For example the
// main set's head: 12 + 12 + 10 + 31 * 0.55 + 13 + 10 = 74.0 = 1.73 * 42.8.
//
// Per slot, 1.73 times the budget and the stats' total (main set, then Arcane):
//   head 74.0 (74.0, 74.2), neck 47.5 (47.4, 47.7), shoulders 57.0 (56.9, 57.0), back 47.5
//   (47.8, 47.4), chest 74.0 (74.0, 74.2), wrist 47.5 (47.5, 47.6), hands 57.0 (56.9, 57.0),
//   waist 65.0 (65.0, 64.8), legs 74.0 (74.0, 74.2), feet 57.0 (56.9, 57.0), finger 47.5 (47.4,
//   47.7), second finger 47.5 (47.8, 47.4), staff 126.3 (126.2, 126.2), wand 27.0 (27.0, 27.2),
//   main hand 79.9 (79.8), off hand 46.4 (46.5).
//
// The presets use the staff with Forever's 2H Spellblasting (+65 spell power). The main hand and
// off hand split the main staff's stats between them. Weapon damage and speed are the Era items'
// (Azuresong, Staff of Dominance, Bonecreeper Stylus) and are not charged to the budget.
//
// The stat mix follows the standards' fair mix rule (item 5a): a realistic spread of stats in the
// proportions Season of Discovery's level 60 BiS mage sets carry (orchestration/gear_targets.md),
// with Stamina at 18% and Spirit at 10%, not everything in spell power. Shares are linear: a
// stat's weighted amount over the set's total (the twelve armor slots, the staff and the wand;
// enchants left out). Main set targets are the averaged SoD fire and frost sets, Arcane's the SoD
// arcane sets, school power counted as spell power:
//
//   set      share     Sta   Int   Spi   SP    Hit   Crit
//   main     target    18    18.6  10    27.3  12.8  13.4
//   main     actual    18.0  18.9  10.0  27.5  11.8  13.8   (153, 160, 85, 425 SP, 100, 117)
//   Arcane   target    18    18.7  10    30.7  12.8  9.4
//   Arcane   actual    18.2  19.1  10.0  31.4  11.8  9.5    (155, 162, 85, 485 SP, 100, 81)
//
// Hit stays at the cap against a level 63 boss (16% spell miss) in both sets: 100 hit rating
// (10%), with 5% from Elemental Precision or Arcane Focus and Manaflare's 2 piece 1%.
//
// Only the head and chest carry the Manaflare Regalia set (Forever's level 60 mage set, see
// sim/mage/item_sets_pve.go), in both sets, so only its 2 piece bonus is live, the same for
// every class's preset. The generator script lived in the session scratchpad (rb/lin.py), the
// numbers below are its output.

func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, foreverSyntheticMageItems...)
}

var foreverSyntheticMageItems = []*proto.UIItem{
	// Head, item level 66, in place of Arcanist Crown (66, 16795).
	{
		Id: 996001, Name: "Mage Synthetic Head (Phase 1)", Icon: "inv_helmet_53",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 12, stats.Intellect: 12, stats.Spirit: 10, stats.SpellPower: 31, stats.Armor: 83}.ToFloatArray(),
		HitRating:  13,
		CritRating: 10,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
		SetName:        "Mage Synthetic (Phase 1)",
	},
	// Neck, item level 71, in place of Choker of the Fire Lord (78, 18814).
	{
		Id: 996002, Name: "Mage Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_17",
		Type:       proto.ItemType_ItemTypeNeck,
		Stats:      stats.Stats{stats.Stamina: 11, stats.Intellect: 12, stats.SpellPower: 28}.ToFloatArray(),
		CritRating: 9,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Shoulder, item level 66, in place of Arcanist Mantle (66, 16797).
	{
		Id: 996003, Name: "Mage Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_02",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 10, stats.Intellect: 11, stats.SpellPower: 27, stats.Armor: 76}.ToFloatArray(),
		HitRating:  12,
		CritRating: 9,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Back, item level 71.
	{
		Id: 996004, Name: "Mage Synthetic Back (Phase 1)", Icon: "inv_misc_cape_16",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 9, stats.Intellect: 10, stats.Spirit: 8, stats.SpellPower: 25, stats.Armor: 54}.ToFloatArray(),
		CritRating: 7,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Chest, item level 66, in place of Robe of Volatile Power (66, 19145).
	{
		Id: 996005, Name: "Mage Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_03",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 12, stats.Intellect: 12, stats.Spirit: 10, stats.SpellPower: 31, stats.Armor: 102}.ToFloatArray(),
		HitRating:  13,
		CritRating: 10,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
		SetName:        "Mage Synthetic (Phase 1)",
	},
	// Wrist, item level 71, in place of Arcanist Bindings (66, 16799).
	{
		Id: 996006, Name: "Mage Synthetic Wrist (Phase 1)", Icon: "inv_bracer_09",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 11, stats.Intellect: 11, stats.Spirit: 9, stats.SpellPower: 30, stats.Armor: 44}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Hands, item level 66, in place of Arcanist Gloves (66, 16801).
	{
		Id: 996007, Name: "Mage Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_14",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 10, stats.Intellect: 11, stats.SpellPower: 27, stats.Armor: 63}.ToFloatArray(),
		HitRating:  12,
		CritRating: 9,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Waist, item level 71, in place of Mana Igniting Cord (71, 19136).
	{
		Id: 996008, Name: "Mage Synthetic Waist (Phase 1)", Icon: "inv_belt_22",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 12, stats.Intellect: 13, stats.SpellPower: 31, stats.Armor: 61}.ToFloatArray(),
		HitRating:  13,
		CritRating: 10,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Legs, item level 66, in place of Netherwind Pants (76, 16915).
	{
		Id: 996009, Name: "Mage Synthetic Legs (Phase 1)", Icon: "inv_pants_08",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 12, stats.Intellect: 12, stats.Spirit: 10, stats.SpellPower: 31, stats.Armor: 101}.ToFloatArray(),
		HitRating:  13,
		CritRating: 10,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Feet, item level 66, in place of Arcanist Boots (66, 16800).
	{
		Id: 996010, Name: "Mage Synthetic Feet (Phase 1)", Icon: "inv_boots_07",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 10, stats.Intellect: 11, stats.Spirit: 9, stats.SpellPower: 27, stats.Armor: 70}.ToFloatArray(),
		HitRating: 12,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Finger, item level 71, in place of Ring of Spell Power (66, 19147).
	{
		Id: 996011, Name: "Mage Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_38",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Stamina: 11, stats.SpellPower: 28}.ToFloatArray(),
		HitRating:  12,
		CritRating: 9,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Finger II, item level 71, in place of Cauterizing Band (71, 19140).
	{
		Id: 996012, Name: "Mage Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_39",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Stamina: 9, stats.Intellect: 10, stats.Spirit: 8, stats.SpellPower: 25}.ToFloatArray(),
		CritRating: 7,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Main Hand, item level 71, in place of Azuresong Mageblade (71, 17103).
	{
		Id: 996013, Name: "Mage Synthetic Main Hand (Phase 1)", Icon: "inv_sword_39",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeMainHand,
		Stats:      stats.Stats{stats.Stamina: 17, stats.Intellect: 18, stats.SpellPower: 45}.ToFloatArray(),
		CritRating: 20,
		// Azuresong Mageblade damage and speed.
		WeaponDamageMin: 64, WeaponDamageMax: 140, WeaponSpeed: 2.4,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Off Hand, item level 71, in place of Fire Runed Grimoire (70, 19142).
	{
		Id: 996014, Name: "Mage Synthetic Off Hand (Phase 1)", Icon: "inv_misc_book_09",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeOffHand, HandType: proto.HandType_HandTypeOffHand,
		Stats: stats.Stats{stats.Stamina: 7, stats.Intellect: 8, stats.Spirit: 21, stats.SpellPower: 19}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Wand, item level 71, in place of Bonecreeper Stylus (62, 13938).
	{
		Id: 996015, Name: "Mage Synthetic Wand (Phase 1)", Icon: "inv_wand_07",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeWand,
		Stats:      stats.Stats{stats.Intellect: 9, stats.SpellPower: 20}.ToFloatArray(),
		CritRating: 7,
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
		Stats:      stats.Stats{stats.Stamina: 24, stats.Intellect: 26, stats.Spirit: 21, stats.SpellPower: 64}.ToFloatArray(),
		CritRating: 20,
		// Staff of Dominance damage and speed.
		WeaponDamageMin: 126, WeaponDamageMax: 205, WeaponSpeed: 2.9,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: head, item level 66, in place of Arcanist Crown (66, 16795).
	{
		Id: 996017, Name: "Mage Arcane Synthetic Head (Phase 1)", Icon: "inv_helmet_53",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 12, stats.Intellect: 13, stats.Spirit: 10, stats.SpellPower: 35, stats.Armor: 83}.ToFloatArray(),
		HitRating:  13,
		CritRating: 7,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
		SetName:        "Mage Synthetic (Phase 1)",
	},
	// Arcane: neck, item level 71, in place of Choker of the Fire Lord (78, 18814).
	{
		Id: 996018, Name: "Mage Arcane Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_17",
		Type:       proto.ItemType_ItemTypeNeck,
		Stats:      stats.Stats{stats.Stamina: 11, stats.Intellect: 12, stats.SpellPower: 34}.ToFloatArray(),
		CritRating: 6,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: shoulder, item level 66, in place of Arcanist Mantle (66, 16797).
	{
		Id: 996019, Name: "Mage Arcane Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_02",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 11, stats.Intellect: 11, stats.SpellPower: 31, stats.Armor: 76}.ToFloatArray(),
		HitRating:  12,
		CritRating: 6,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: back, item level 71.
	{
		Id: 996020, Name: "Mage Arcane Synthetic Back (Phase 1)", Icon: "inv_misc_cape_16",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 9, stats.Intellect: 10, stats.Spirit: 8, stats.SpellPower: 28, stats.Armor: 54}.ToFloatArray(),
		CritRating: 5,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: chest, item level 66, in place of Robe of Volatile Power (66, 19145).
	{
		Id: 996021, Name: "Mage Arcane Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_03",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 12, stats.Intellect: 13, stats.Spirit: 10, stats.SpellPower: 35, stats.Armor: 102}.ToFloatArray(),
		HitRating:  13,
		CritRating: 7,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
		SetName:        "Mage Synthetic (Phase 1)",
	},
	// Arcane: wrist, item level 71, in place of Arcanist Bindings (66, 16799).
	{
		Id: 996022, Name: "Mage Arcane Synthetic Wrist (Phase 1)", Icon: "inv_bracer_09",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 10, stats.Intellect: 11, stats.Spirit: 9, stats.SpellPower: 32, stats.Armor: 44}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: hands, item level 66, in place of Arcanist Gloves (66, 16801).
	{
		Id: 996023, Name: "Mage Arcane Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_14",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 11, stats.Intellect: 11, stats.SpellPower: 31, stats.Armor: 63}.ToFloatArray(),
		HitRating:  12,
		CritRating: 6,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: waist, item level 71, in place of Mana Igniting Cord (71, 19136).
	{
		Id: 996024, Name: "Mage Arcane Synthetic Waist (Phase 1)", Icon: "inv_belt_22",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 12, stats.Intellect: 12, stats.SpellPower: 36, stats.Armor: 61}.ToFloatArray(),
		HitRating:  14,
		CritRating: 7,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: legs, item level 66, in place of Netherwind Pants (76, 16915).
	{
		Id: 996025, Name: "Mage Arcane Synthetic Legs (Phase 1)", Icon: "inv_pants_08",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 12, stats.Intellect: 13, stats.Spirit: 10, stats.SpellPower: 35, stats.Armor: 101}.ToFloatArray(),
		HitRating:  13,
		CritRating: 7,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: feet, item level 66, in place of Arcanist Boots (66, 16800).
	{
		Id: 996026, Name: "Mage Arcane Synthetic Feet (Phase 1)", Icon: "inv_boots_07",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 10, stats.Intellect: 11, stats.Spirit: 9, stats.SpellPower: 29, stats.Armor: 70}.ToFloatArray(),
		HitRating: 11,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: finger, item level 71, in place of Ring of Spell Power (66, 19147).
	{
		Id: 996027, Name: "Mage Arcane Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_38",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Stamina: 11, stats.SpellPower: 34}.ToFloatArray(),
		HitRating:  12,
		CritRating: 6,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: finger II, item level 71, in place of Cauterizing Band (71, 19140).
	{
		Id: 996028, Name: "Mage Arcane Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_39",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Stamina: 9, stats.Intellect: 10, stats.Spirit: 8, stats.SpellPower: 28}.ToFloatArray(),
		CritRating: 5,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: staff, the main hand and off hand budgets together (46.2 + 27.5), in place of Staff of
	// Dominance (70, 18842), for the presets that sim better with a two-hander.
	{
		Id: 996029, Name: "Mage Arcane Synthetic Staff (Phase 1)", Icon: "inv_staff_13",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand,
		Stats:      stats.Stats{stats.Stamina: 25, stats.Intellect: 26, stats.Spirit: 21, stats.SpellPower: 73}.ToFloatArray(),
		CritRating: 14,
		// Staff of Dominance damage and speed.
		WeaponDamageMin: 126, WeaponDamageMax: 205, WeaponSpeed: 2.9,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: wand, item level 71, in place of Bonecreeper Stylus (62, 13938).
	{
		Id: 996030, Name: "Mage Arcane Synthetic Wand (Phase 1)", Icon: "inv_wand_07",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeWand,
		Stats:      stats.Stats{stats.Intellect: 9, stats.SpellPower: 24}.ToFloatArray(),
		CritRating: 5,
		// Bonecreeper Stylus damage and speed.
		WeaponDamageMin: 83, WeaponDamageMax: 155, WeaponSpeed: 1.9,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
}
