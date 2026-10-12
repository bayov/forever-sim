package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Synthetic level 60 items for a Phase 1 shadow priest under Forever, made the way the
// enhancement items in forever_synthetic_items.go are (read its comment). Shadow is the only
// priest DPS spec, so there is one set. Ids 995000-995999.
//
// The budget is the Classic formula: the stats of a piece satisfy
//
//	sum((stat * cost)^1.5) = (value * slot)^1.5, value = (ilvl - 1.3) / 1.3 for an epic,
//
// with a cost of 1 for Stamina, Intellect and Spirit, 0.55 for spell power (Forever's price)
// and 1 for each point of hit rating (10 a percent) and crit rating (14 a percent). The slot
// modifiers are Classic's: 1 for head, chest and legs, 0.75 for shoulders, hands, waist and
// feet, 0.5625 for neck, back, wrist, fingers and the off hand, 0.32 for the wand. The caster
// main hand gets 0.86, what Azuresong Mageblade's own stats come to at its item level. The
// formula checks out on Era items: Robe of Volatile Power (66) uses 50.4 of its 49.8 and
// Choker of the Fire Lord (78) 33.6 of 33.2 with spell power at Era's 0.86. Each comment
// below gives the piece's budget (value * slot) and what its stats use, a little under after
// rounding.
//
// The pieces take the place of the Era Phase 1 items a shadow priest wears (the Molten Core
// tier, Choker of the Fire Lord, Mana Igniting Cord, Azuresong Mageblade and Fire Runed
// Grimoire, the Molten Core rings) at item level 66 or the replaced item's level where that
// is higher. The trinkets stay real items (Talisman of Ephemeral Power, Briarwood Reed).
//
// The stat mix:
//   - Spirit is on 9 of the 15 pieces, 85 in all, 10.4% of the set's stat points. Spirit Tap
//     and Spiritual Guidance do nothing worth taking in a raid, so the set has no more than
//     the 10% every mana user gets.
//   - Spell power is most of each piece (635 in all), more on the hands, the first ring and the
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

func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, foreverSyntheticPriestItems...)
}

var foreverSyntheticPriestItems = []*proto.UIItem{
	// Head, item level 66, in place of Circlet of Prophecy (66, 16813). Budget 49.77, uses 49.45.
	{
		Id: 995001, Name: "Shadow Synthetic Head (Phase 1)", Icon: "inv_helmet_34",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 13, stats.Intellect: 16, stats.Spirit: 8, stats.SpellPower: 54, stats.Armor: 83}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
		SetName:        "Shadow Synthetic (Phase 1)",
	},
	// Neck, item level 78, in place of Choker of the Fire Lord (78, 18814). Budget 33.19, uses 32.79.
	{
		Id: 995002, Name: "Shadow Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_17",
		Type:      proto.ItemType_ItemTypeNeck,
		Stats:     stats.Stats{stats.Stamina: 7, stats.Intellect: 8, stats.SpellPower: 43}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      78, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Shoulder, item level 66, in place of Mantle of Prophecy (66, 16816). Budget 37.33, uses 37.31.
	{
		Id: 995003, Name: "Shadow Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_02",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 11, stats.Intellect: 12, stats.Spirit: 8, stats.SpellPower: 46, stats.Armor: 76}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
		SetName:        "Shadow Synthetic (Phase 1)",
	},
	// Back, item level 66, a Molten Core cloak. Budget 28.00, uses 27.77.
	{
		Id: 995004, Name: "Shadow Synthetic Back (Phase 1)", Icon: "inv_misc_cape_16",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 9, stats.Intellect: 7, stats.Spirit: 9, stats.SpellPower: 32, stats.Armor: 55}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Chest, item level 66, in place of Robe of Volatile Power (66, 19145). Budget 49.77, uses 49.46.
	{
		Id: 995005, Name: "Shadow Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_08",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 13, stats.Intellect: 16, stats.Spirit: 10, stats.SpellPower: 52, stats.Armor: 102}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
		SetName:        "Shadow Synthetic (Phase 1)",
	},
	// Wrist, item level 66, in place of Vambraces of Prophecy (66, 16819). Budget 28.00, uses 27.87.
	{
		Id: 995006, Name: "Shadow Synthetic Wrist (Phase 1)", Icon: "inv_bracer_09",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 6, stats.Intellect: 9, stats.SpellPower: 32, stats.Armor: 44}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Hands, item level 66, in place of Gloves of Prophecy (66, 16812). Budget 37.33, uses 37.10.
	{
		Id: 995007, Name: "Shadow Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_14",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 9, stats.Intellect: 12, stats.SpellPower: 53, stats.Armor: 63}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
		SetName:        "Shadow Synthetic (Phase 1)",
	},
	// Waist, item level 71, in place of Mana Igniting Cord (71, 19136). Budget 40.21, uses 40.07.
	{
		Id: 995008, Name: "Shadow Synthetic Waist (Phase 1)", Icon: "inv_belt_08",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 9, stats.Intellect: 10, stats.SpellPower: 42, stats.Armor: 61}.ToFloatArray(),
		HitRating: 10, CritRating: 14,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Legs, item level 66, in place of Pants of Prophecy (66, 16814). Budget 49.77, uses 49.47.
	{
		Id: 995009, Name: "Shadow Synthetic Legs (Phase 1)", Icon: "inv_pants_08",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 15, stats.Intellect: 15, stats.Spirit: 13, stats.SpellPower: 59, stats.Armor: 89}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
		SetName:        "Shadow Synthetic (Phase 1)",
	},
	// Feet, item level 66, in place of Boots of Prophecy (66, 16811). Budget 37.33, uses 37.20.
	{
		Id: 995010, Name: "Shadow Synthetic Feet (Phase 1)", Icon: "inv_boots_cloth_05",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 13, stats.Intellect: 11, stats.Spirit: 11, stats.SpellPower: 41, stats.Armor: 70}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
		SetName:        "Shadow Synthetic (Phase 1)",
	},
	// Finger, item level 66, in place of Ring of Spell Power (66, 19147). Budget 28.00, uses 27.83.
	{
		Id: 995011, Name: "Shadow Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_29",
		Type:  proto.ItemType_ItemTypeFinger,
		Stats: stats.Stats{stats.Stamina: 5, stats.SpellPower: 48}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Finger II, item level 71, in place of Cauterizing Band (71, 19140). Budget 30.16, uses 30.14.
	{
		Id: 995012, Name: "Shadow Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_36",
		Type:  proto.ItemType_ItemTypeFinger,
		Stats: stats.Stats{stats.Stamina: 9, stats.Intellect: 13, stats.Spirit: 14, stats.SpellPower: 21}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Main Hand, item level 71, in place of Azuresong Mageblade (71, 17103). Budget 46.11, uses 46.04.
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
	// Off Hand, item level 70, in place of Fire Runed Grimoire (70, 19142). Budget 29.73, uses 29.56.
	{
		Id: 995014, Name: "Shadow Synthetic Off Hand (Phase 1)", Icon: "inv_misc_book_09",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeOffHand, HandType: proto.HandType_HandTypeOffHand,
		Stats: stats.Stats{stats.Stamina: 10, stats.Intellect: 14, stats.Spirit: 8, stats.SpellPower: 26}.ToFloatArray(),
		Ilvl:  70, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Wand, item level 66, in place of Bonecreeper Stylus (62, 13938). Budget 15.93, uses 15.54.
	{
		Id: 995015, Name: "Shadow Synthetic Wand (Phase 1)", Icon: "inv_wand_07",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeWand,
		Stats: stats.Stats{stats.Intellect: 5, stats.Spirit: 4, stats.SpellPower: 22}.ToFloatArray(),
		// Bonecreeper Stylus damage and speed.
		WeaponDamageMin: 83, WeaponDamageMax: 155, WeaponSpeed: 1.9,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
}
