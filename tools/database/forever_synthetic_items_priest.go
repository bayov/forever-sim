package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Synthetic level 60 items for a Phase 1 shadow priest under Forever, made the way the
// enhancement items in forever_synthetic_items.go are (read its comment). Shadow is the only
// priest DPS spec, so there is one set. Ids 995000-995999.
//
// The budget is the standard all nine classes use (backed out of the enhancement items): the
// stats of a piece satisfy
//
//	sum((stat * cost)^1.5)^(2/3) = value * slot, value = 1.2 * ilvl - 36.4 for an epic
//
// (42.8 at 66, 48.8 at 71, 47.6 at 70, 57.2 at 78), with a cost of 1 for Stamina, Intellect
// and Spirit, 0.55 for spell power (Forever's price, all schools, no Shadow-only power on the
// set) and 1 for each point of hit rating (10 a percent) and crit rating (14 a percent). Armor
// is free. The slot modifiers are 1 for head, chest and legs, 0.77 for shoulders, hands, waist
// and feet, 0.5625 for neck, back, wrist, fingers and the off hand, 0.32 for the wand.
//
// The main hand follows the real Era weapon instead of the 0.42 one-hand modifier, because Era
// caster weapons carry far more spell power than that: Azuresong Mageblade's own stats (7
// Stamina, 12 Intellect, 40 spell power at Era's 0.86, 1% crit) come to 46.2, and that is the
// main hand's budget. Each comment below gives the piece's budget (value * slot) and what its
// stats use, a little under after rounding. The wand's damage and speed are Bonecreeper
// Stylus's and are not charged to the budget.
//
// The pieces take the place of the Era Phase 1 items a shadow priest wears (the Molten Core
// tier, Choker of the Fire Lord, Mana Igniting Cord, Azuresong Mageblade and Fire Runed
// Grimoire, the Molten Core rings, Bonecreeper Stylus) at item level 66 or the replaced item's
// level where that is higher. The trinkets stay real items (Talisman of Ephemeral Power,
// Briarwood Reed).
//
// The stat mix:
//   - Spirit is on 9 of the 15 pieces, 75 in all, 10.3% of the set's stat points. Spirit Tap
//     and Spiritual Guidance do nothing worth taking in a raid, so the set has no more than
//     the 10% every mana user gets.
//   - Spell power is most of each piece (561 in all), more on the hands, the first ring and the
//     main hand, less on the second ring and the off hand, which carry Intellect and Spirit.
//   - 30 hit rating (3%) on the neck, wrist and waist, and 56 crit rating (4%) on the head,
//     chest, waist and main hand. That is the 2 to 3% hit and 4% crit the Era Phase 1 BiS
//     carries (Mana Igniting Cord, Robe of Volatile Power, Azuresong Mageblade). With Shadow
//     Focus 5/5 and the set's 2 piece bonus it leaves 7% of the misses against a level 63 boss.
//   - The neck, wrist, hands, waist, first ring and main hand have no Spirit.
//
// The main hand is a sword, like Azuresong Mageblade, so a Human gets their 2% crit with swords.
//
// The head, shoulders, chest, hands, legs and feet form a set with the bonuses of Raiments of
// Conviction, Forever's Tier 1 shadow set (see sim/priest/item_sets_pve.go). The generator
// script lived in the session scratchpad (synth.py), the numbers below are its output.
//
// The Smite build wants the same stats, so it wears the same pieces, but its six armor pieces
// are copies (995016-995021) that form a set with the bonuses of Vestments of Conviction,
// Forever's Tier 1 Discipline and Holy set, which beat Raiments of Conviction for it.

func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, foreverSyntheticPriestItems...)
}

var foreverSyntheticPriestItems = []*proto.UIItem{
	// Head, item level 66, in place of Circlet of Prophecy (66, 16813). Budget 42.80, uses 42.53.
	{
		Id: 995001, Name: "Shadow Synthetic Head (Phase 1)", Icon: "inv_helmet_34",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 11, stats.Intellect: 13, stats.Spirit: 7, stats.SpellPower: 45, stats.Armor: 83}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
		SetName:        "Shadow Synthetic (Phase 1)",
	},
	// Neck, item level 78, in place of Choker of the Fire Lord (78, 18814). Budget 32.17, uses 31.86.
	{
		Id: 995002, Name: "Shadow Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_17",
		Type:      proto.ItemType_ItemTypeNeck,
		Stats:     stats.Stats{stats.Stamina: 6, stats.Intellect: 7, stats.SpellPower: 43}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      78, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Shoulder, item level 66, in place of Mantle of Prophecy (66, 16816). Budget 32.96, uses 32.58.
	{
		Id: 995003, Name: "Shadow Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_02",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 10, stats.Intellect: 11, stats.Spirit: 7, stats.SpellPower: 39, stats.Armor: 76}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
		SetName:        "Shadow Synthetic (Phase 1)",
	},
	// Back, item level 66, a Molten Core cloak. Budget 24.08, uses 23.79.
	{
		Id: 995004, Name: "Shadow Synthetic Back (Phase 1)", Icon: "inv_misc_cape_16",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 7, stats.Intellect: 7, stats.Spirit: 7, stats.SpellPower: 28, stats.Armor: 55}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Chest, item level 66, in place of Robe of Volatile Power (66, 19145). Budget 42.80, uses 42.57.
	{
		Id: 995005, Name: "Shadow Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_08",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 11, stats.Intellect: 13, stats.Spirit: 9, stats.SpellPower: 43, stats.Armor: 102}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
		SetName:        "Shadow Synthetic (Phase 1)",
	},
	// Wrist, item level 66, in place of Vambraces of Prophecy (66, 16819). Budget 24.08, uses 23.70.
	{
		Id: 995006, Name: "Shadow Synthetic Wrist (Phase 1)", Icon: "inv_bracer_09",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 5, stats.Intellect: 7, stats.SpellPower: 26, stats.Armor: 44}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Hands, item level 66, in place of Gloves of Prophecy (66, 16812). Budget 32.96, uses 32.55.
	{
		Id: 995007, Name: "Shadow Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_14",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 8, stats.Intellect: 10, stats.SpellPower: 47, stats.Armor: 63}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
		SetName:        "Shadow Synthetic (Phase 1)",
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
	// Legs, item level 66, in place of Pants of Prophecy (66, 16814). Budget 42.80, uses 42.68.
	{
		Id: 995009, Name: "Shadow Synthetic Legs (Phase 1)", Icon: "inv_pants_08",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 13, stats.Intellect: 13, stats.Spirit: 11, stats.SpellPower: 51, stats.Armor: 89}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
		SetName:        "Shadow Synthetic (Phase 1)",
	},
	// Feet, item level 66, in place of Boots of Prophecy (66, 16811). Budget 32.96, uses 32.80.
	{
		Id: 995010, Name: "Shadow Synthetic Feet (Phase 1)", Icon: "inv_boots_cloth_05",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 11, stats.Intellect: 10, stats.Spirit: 10, stats.SpellPower: 36, stats.Armor: 70}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
		SetName:        "Shadow Synthetic (Phase 1)",
	},
	// Finger, item level 66, in place of Ring of Spell Power (66, 19147). Budget 24.08, uses 23.66.
	{
		Id: 995011, Name: "Shadow Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_29",
		Type:  proto.ItemType_ItemTypeFinger,
		Stats: stats.Stats{stats.Stamina: 4, stats.SpellPower: 41}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
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
	// Off Hand, item level 70, in place of Fire Runed Grimoire (70, 19142). Budget 26.78, uses 26.41.
	{
		Id: 995014, Name: "Shadow Synthetic Off Hand (Phase 1)", Icon: "inv_misc_book_09",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeOffHand, HandType: proto.HandType_HandTypeOffHand,
		Stats: stats.Stats{stats.Stamina: 9, stats.Intellect: 13, stats.Spirit: 8, stats.SpellPower: 21}.ToFloatArray(),
		Ilvl:  70, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Wand, item level 66, in place of Bonecreeper Stylus (62, 13938). Budget 13.70, uses 13.60.
	{
		Id: 995015, Name: "Shadow Synthetic Wand (Phase 1)", Icon: "inv_wand_07",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeWand,
		Stats: stats.Stats{stats.Intellect: 5, stats.Spirit: 3, stats.SpellPower: 19}.ToFloatArray(),
		// Bonecreeper Stylus damage and speed.
		WeaponDamageMin: 83, WeaponDamageMax: 155, WeaponSpeed: 1.9,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Head for the Smite set: the same stats as 995001, with Vestments of Conviction.
	{
		Id: 995016, Name: "Smite Synthetic Head (Phase 1)", Icon: "inv_helmet_34",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 11, stats.Intellect: 13, stats.Spirit: 7, stats.SpellPower: 45, stats.Armor: 83}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
		SetName:        "Smite Synthetic (Phase 1)",
	},
	// Shoulder for the Smite set: the same stats as 995003, with Vestments of Conviction.
	{
		Id: 995017, Name: "Smite Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_02",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 10, stats.Intellect: 11, stats.Spirit: 7, stats.SpellPower: 39, stats.Armor: 76}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
		SetName:        "Smite Synthetic (Phase 1)",
	},
	// Chest for the Smite set: the same stats as 995005, with Vestments of Conviction.
	{
		Id: 995018, Name: "Smite Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_08",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 11, stats.Intellect: 13, stats.Spirit: 9, stats.SpellPower: 43, stats.Armor: 102}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
		SetName:        "Smite Synthetic (Phase 1)",
	},
	// Hands for the Smite set: the same stats as 995007, with Vestments of Conviction.
	{
		Id: 995019, Name: "Smite Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_14",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 8, stats.Intellect: 10, stats.SpellPower: 47, stats.Armor: 63}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
		SetName:        "Smite Synthetic (Phase 1)",
	},
	// Legs for the Smite set: the same stats as 995009, with Vestments of Conviction.
	{
		Id: 995020, Name: "Smite Synthetic Legs (Phase 1)", Icon: "inv_pants_08",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 13, stats.Intellect: 13, stats.Spirit: 11, stats.SpellPower: 51, stats.Armor: 89}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
		SetName:        "Smite Synthetic (Phase 1)",
	},
	// Feet for the Smite set: the same stats as 995010, with Vestments of Conviction.
	{
		Id: 995021, Name: "Smite Synthetic Feet (Phase 1)", Icon: "inv_boots_cloth_05",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 11, stats.Intellect: 10, stats.Spirit: 10, stats.SpellPower: 36, stats.Armor: 70}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
		SetName:        "Smite Synthetic (Phase 1)",
	},
}
