package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Synthetic level 60 items for a Phase 1 priest under Forever, made the way the enhancement
// items in forever_synthetic_items.go are (read its comment), by the cross-class standard in
// orchestration/standards.md. Ids 995000-995999.
//
// The stats of a piece satisfy
//
//	sum((stat * cost)^1.5)^(2/3) = value * slot, value = 1.2 * ilvl - 36.4 for an epic
//
// (42.8 at 66, 48.8 at 71), with a cost of 1 for Stamina, Intellect and Spirit, 0.55 for spell
// power (Forever's price, all schools) and 1 for each point of hit rating (10 a percent) and
// crit rating (14 a percent). Armor is free. The slot modifiers are 1 for head, chest and legs,
// 0.77 for shoulders, hands, waist and feet, 0.5625 for neck, back, wrist, fingers and the off
// hand, 0.32 for the wand. Every class uses one item level table: 66 for head, shoulders,
// chest, hands, legs and feet, 71 for neck, back, wrist, waist and both rings, and 71 for the
// off hand and the wand.
//
// The main hand follows the real Era weapon instead of the 0.42 one-hand modifier, because Era
// caster weapons carry far more spell power than that: Azuresong Mageblade's own stats (7
// Stamina, 12 Intellect, 40 spell power at Era's 0.86, 1% crit) come to 46.2, and that is the
// main hand's budget. Each comment below gives the piece's budget (value * slot) and what its
// stats use, a little under after rounding. Weapon and wand damage and speed are the real Era
// items' and are not charged to the budget.
//
// The pieces take the place of the Era Phase 1 items a shadow priest wears (the Molten Core
// tier, Choker of the Fire Lord, Mana Igniting Cord, Azuresong Mageblade and Fire Runed
// Grimoire, the Molten Core rings, Bonecreeper Stylus). The trinkets stay real items (Talisman
// of Ephemeral Power, Briarwood Reed).
//
// The Shadow set (995001-995015):
//   - 100 hit rating (10%), spread over twelve pieces. With Shadow Focus 5/5 (5%) and the 2
//     piece bonus (1%) that is the 16% a caster needs against a level 63 boss, and no more.
//   - 56 crit rating (4%) on the head, chest, waist and main hand, the 4% crit the Era Phase 1
//     BiS carries (Robe of Volatile Power, Mana Igniting Cord, Azuresong Mageblade).
//   - 71 Spirit, about 9% of the stat points, about the 10% every mana user gets.
//   - Spell power is the rest (544).
//
// The Smite set (995016-995027) has the same pieces without hit rating: Holy Precision 3/3
// gives Smite, Holy Fire and Penance 18%, so hit is worth nothing to it, and its budget goes to
// spell power (589), Intellect and Spirit. The rings and the main hand carry no hit and are
// shared.
//
// Only the head and shoulders carry a set name, so only the 2 piece bonus is live: Raiments of
// Conviction (Forever's Tier 1 shadow set, 1% hit) on the Shadow set and Vestments of
// Conviction (Tier 1 Discipline and Holy, 10 Spirit) on the Smite set (see
// sim/priest/item_sets_pve.go). The 4 and 5 piece bonuses are measured apart in the presets.
//
// The main hand is a sword, like Azuresong Mageblade, so a Human gets their 2% crit with
// swords. A staff for each set (995028, 995029) takes the place of the main hand and off hand at
// the same weapon budget, for the 2H enchants. The generator lived in the session scratchpad (synth3.py); the numbers below are its
// output.

func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, foreverSyntheticPriestItems...)
}

var foreverSyntheticPriestItems = []*proto.UIItem{
	// Head, item level 66, in place of Circlet of Prophecy (66, 16813). Budget 42.80, uses 42.70.
	{
		Id: 995001, Name: "Shadow Synthetic Head (Phase 1)", Icon: "inv_helmet_34",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 10, stats.Intellect: 12, stats.Spirit: 6, stats.SpellPower: 41, stats.Armor: 83}.ToFloatArray(),
		HitRating: 10, CritRating: 14,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
		SetName:        "Shadow Synthetic (Phase 1)",
	},
	// Neck, item level 71, in place of Choker of the Fire Lord (78, 18814). Budget 27.45, uses 27.21.
	{
		Id: 995002, Name: "Shadow Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_17",
		Type:      proto.ItemType_ItemTypeNeck,
		Stats:     stats.Stats{stats.Stamina: 5, stats.Intellect: 6, stats.SpellPower: 35}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Shoulder, item level 66, in place of Mantle of Prophecy (66, 16816). Budget 32.96, uses 32.80.
	{
		Id: 995003, Name: "Shadow Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_02",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 9, stats.Intellect: 10, stats.Spirit: 7, stats.SpellPower: 36, stats.Armor: 76}.ToFloatArray(),
		HitRating: 8,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
		SetName:        "Shadow Synthetic (Phase 1)",
	},
	// Back, item level 71, a Molten Core cloak. Budget 27.45, uses 27.24.
	{
		Id: 995004, Name: "Shadow Synthetic Back (Phase 1)", Icon: "inv_misc_cape_16",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 8, stats.Intellect: 7, stats.Spirit: 8, stats.SpellPower: 29, stats.Armor: 55}.ToFloatArray(),
		HitRating: 6,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Chest, item level 66, in place of Robe of Volatile Power (66, 19145). Budget 42.80, uses 42.71.
	{
		Id: 995005, Name: "Shadow Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_08",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 10, stats.Intellect: 12, stats.Spirit: 8, stats.SpellPower: 39, stats.Armor: 102}.ToFloatArray(),
		HitRating: 10, CritRating: 14,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Wrist, item level 71, in place of Vambraces of Prophecy (66, 16819). Budget 27.45, uses 27.32.
	{
		Id: 995006, Name: "Shadow Synthetic Wrist (Phase 1)", Icon: "inv_bracer_09",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 6, stats.Intellect: 8, stats.SpellPower: 32, stats.Armor: 44}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Hands, item level 66, in place of Gloves of Prophecy (66, 16812). Budget 32.96, uses 32.81.
	{
		Id: 995007, Name: "Shadow Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_14",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 8, stats.Intellect: 10, stats.SpellPower: 42, stats.Armor: 63}.ToFloatArray(),
		HitRating: 8,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Waist, item level 71, in place of Mana Igniting Cord (71, 19136). Budget 37.58, uses 37.45.
	{
		Id: 995008, Name: "Shadow Synthetic Waist (Phase 1)", Icon: "inv_belt_08",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 8, stats.Intellect: 9, stats.SpellPower: 38, stats.Armor: 61}.ToFloatArray(),
		HitRating: 10, CritRating: 14,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Legs, item level 66, in place of Pants of Prophecy (66, 16814). Budget 42.80, uses 42.58.
	{
		Id: 995009, Name: "Shadow Synthetic Legs (Phase 1)", Icon: "inv_pants_08",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 12, stats.Intellect: 12, stats.Spirit: 10, stats.SpellPower: 47, stats.Armor: 89}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Feet, item level 66, in place of Boots of Prophecy (66, 16811). Budget 32.96, uses 32.53.
	{
		Id: 995010, Name: "Shadow Synthetic Feet (Phase 1)", Icon: "inv_boots_cloth_05",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 10, stats.Intellect: 9, stats.Spirit: 9, stats.SpellPower: 33, stats.Armor: 70}.ToFloatArray(),
		HitRating: 8,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Finger, item level 71, in place of Ring of Spell Power (66, 19147). Budget 27.45, uses 27.30.
	{
		Id: 995011, Name: "Shadow Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_29",
		Type:  proto.ItemType_ItemTypeFinger,
		Stats: stats.Stats{stats.Stamina: 5, stats.SpellPower: 47}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Finger II, item level 71, in place of Cauterizing Band (71, 19140). Budget 27.45, uses 27.28.
	{
		Id: 995012, Name: "Shadow Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_36",
		Type:  proto.ItemType_ItemTypeFinger,
		Stats: stats.Stats{stats.Stamina: 8, stats.Intellect: 11, stats.Spirit: 13, stats.SpellPower: 20}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Main Hand, item level 71, in place of Azuresong Mageblade (71, 17103). Budget 46.22, uses 46.04.
	{
		Id: 995013, Name: "Shadow Synthetic Main Hand (Phase 1)", Icon: "inv_sword_39",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeMainHand,
		Stats:      stats.Stats{stats.Stamina: 6, stats.Intellect: 11, stats.SpellPower: 64}.ToFloatArray(),
		CritRating: 14,
		// Azuresong Mageblade damage and speed.
		WeaponDamageMin: 64, WeaponDamageMax: 140, WeaponSpeed: 2.4,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Off Hand, item level 71, in place of Fire Runed Grimoire (70, 19142). Budget 27.45, uses 27.13.
	{
		Id: 995014, Name: "Shadow Synthetic Off Hand (Phase 1)", Icon: "inv_misc_book_09",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeOffHand, HandType: proto.HandType_HandTypeOffHand,
		Stats:     stats.Stats{stats.Stamina: 8, stats.Intellect: 13, stats.Spirit: 7, stats.SpellPower: 22}.ToFloatArray(),
		HitRating: 5,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Wand, item level 71, in place of Bonecreeper Stylus (62, 13938). Budget 15.62, uses 15.55.
	{
		Id: 995015, Name: "Shadow Synthetic Wand (Phase 1)", Icon: "inv_wand_07",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeWand,
		Stats:     stats.Stats{stats.Intellect: 5, stats.Spirit: 3, stats.SpellPower: 19}.ToFloatArray(),
		HitRating: 5,
		// Bonecreeper Stylus damage and speed.
		WeaponDamageMin: 83, WeaponDamageMax: 155, WeaponSpeed: 1.9,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Head for the Smite set, item level 66: head without hit rating. Budget 42.80, uses 42.53.
	{
		Id: 995016, Name: "Smite Synthetic Head (Phase 1)", Icon: "inv_helmet_34",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 11, stats.Intellect: 13, stats.Spirit: 7, stats.SpellPower: 45, stats.Armor: 83}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
		SetName:        "Smite Synthetic (Phase 1)",
	},
	// Neck for the Smite set, item level 71: neck without hit rating. Budget 27.45, uses 27.00.
	{
		Id: 995022, Name: "Smite Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_17",
		Type:  proto.ItemType_ItemTypeNeck,
		Stats: stats.Stats{stats.Stamina: 6, stats.Intellect: 7, stats.SpellPower: 41}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Shoulder for the Smite set, item level 66: shoulder without hit rating. Budget 32.96, uses 32.58.
	{
		Id: 995017, Name: "Smite Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_02",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 10, stats.Intellect: 11, stats.Spirit: 7, stats.SpellPower: 39, stats.Armor: 76}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
		SetName:        "Smite Synthetic (Phase 1)",
	},
	// Back for the Smite set, item level 71: back without hit rating. Budget 27.45, uses 26.65.
	{
		Id: 995023, Name: "Smite Synthetic Back (Phase 1)", Icon: "inv_misc_cape_16",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 8, stats.Intellect: 7, stats.Spirit: 8, stats.SpellPower: 32, stats.Armor: 55}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Chest for the Smite set, item level 66: chest without hit rating. Budget 42.80, uses 42.57.
	{
		Id: 995018, Name: "Smite Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_08",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 11, stats.Intellect: 13, stats.Spirit: 9, stats.SpellPower: 43, stats.Armor: 102}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Wrist for the Smite set, item level 71: wrist without hit rating. Budget 27.45, uses 27.21.
	{
		Id: 995024, Name: "Smite Synthetic Wrist (Phase 1)", Icon: "inv_bracer_09",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 7, stats.Intellect: 10, stats.SpellPower: 37, stats.Armor: 44}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Hands for the Smite set, item level 66: hands without hit rating. Budget 32.96, uses 32.55.
	{
		Id: 995019, Name: "Smite Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_14",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 8, stats.Intellect: 10, stats.SpellPower: 47, stats.Armor: 63}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Waist for the Smite set, item level 71: waist without hit rating. Budget 37.58, uses 37.20.
	{
		Id: 995025, Name: "Smite Synthetic Waist (Phase 1)", Icon: "inv_belt_08",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 9, stats.Intellect: 11, stats.SpellPower: 42, stats.Armor: 61}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Legs for the Smite set, item level 66: legs without hit rating. Budget 42.80, uses 42.68.
	{
		Id: 995020, Name: "Smite Synthetic Legs (Phase 1)", Icon: "inv_pants_08",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 13, stats.Intellect: 13, stats.Spirit: 11, stats.SpellPower: 51, stats.Armor: 89}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Feet for the Smite set, item level 66: feet without hit rating. Budget 32.96, uses 32.80.
	{
		Id: 995021, Name: "Smite Synthetic Feet (Phase 1)", Icon: "inv_boots_cloth_05",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 11, stats.Intellect: 10, stats.Spirit: 10, stats.SpellPower: 36, stats.Armor: 70}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Off Hand for the Smite set, item level 71: off hand without hit rating. Budget 27.45, uses 27.15.
	{
		Id: 995026, Name: "Smite Synthetic Off Hand (Phase 1)", Icon: "inv_misc_book_09",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeOffHand, HandType: proto.HandType_HandTypeOffHand,
		Stats: stats.Stats{stats.Stamina: 9, stats.Intellect: 13, stats.Spirit: 8, stats.SpellPower: 23}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Wand for the Smite set, item level 71: wand without hit rating. Budget 15.62, uses 15.54.
	{
		Id: 995027, Name: "Smite Synthetic Wand (Phase 1)", Icon: "inv_wand_07",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeWand,
		Stats: stats.Stats{stats.Intellect: 5, stats.Spirit: 4, stats.SpellPower: 22}.ToFloatArray(),
		// Bonecreeper Stylus damage and speed.
		WeaponDamageMin: 83, WeaponDamageMax: 155, WeaponSpeed: 1.9,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Staff for the Shadow set, item level 71, in place of the main hand and off hand: a caster's
	// weapon slots total the same budget either way (46.22 + 27.45 = 73.67), uses 73.25.
	{
		Id: 995028, Name: "Shadow Synthetic Staff (Phase 1)", Icon: "inv_staff_13",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand,
		Stats:     stats.Stats{stats.Stamina: 14, stats.Intellect: 23, stats.Spirit: 14, stats.SpellPower: 90}.ToFloatArray(),
		HitRating: 5, CritRating: 14,
		// A caster staff's damage and speed, which do nothing for a priest.
		WeaponDamageMin: 105, WeaponDamageMax: 197, WeaponSpeed: 2.9,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Staff for the Smite set, item level 71, in place of the main hand and off hand: a caster's
	// weapon slots total the same budget either way (46.22 + 27.45 = 73.67), uses 73.27.
	{
		Id: 995029, Name: "Smite Synthetic Staff (Phase 1)", Icon: "inv_staff_13",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand,
		Stats:      stats.Stats{stats.Stamina: 15, stats.Intellect: 23, stats.Spirit: 15, stats.SpellPower: 90}.ToFloatArray(),
		CritRating: 14,
		// A caster staff's damage and speed, which do nothing for a priest.
		WeaponDamageMin: 105, WeaponDamageMax: 197, WeaponSpeed: 2.9,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
}
