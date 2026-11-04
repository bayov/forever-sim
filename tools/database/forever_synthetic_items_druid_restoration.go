package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The synthetic Phase 1 Restoration druid items, ids 990400 to 990499: the healer's set
// (docs/healing_druid.md). See forever_synthetic_items.go for what these items are and why we need them.
// The trinkets are the Balance set's (998014 and 998015), whose spell power heals as it damages.
//
// The budget is the Restoration shaman's (forever_synthetic_items_shaman_restoration.go), standards 1b for
// a pure caster: weights 1 a point for Stamina, Intellect, Spirit and crit rating (14 is 1%), spell power
// 0.86, healing on top of spell power 0.45, MP5 2.4, armor free. Each armor piece carries slot mod *
// L(ilvl) * f in weighted linear stats within 1% (L = 1.125 * ilvl + 2.0, f 1.053 / 1.026 / 0.995), at
// item level 66 for the head, shoulders, chest, hands, legs and feet and 71 for the neck, back, wrist,
// waist and rings, and passes the shape check (its 1.7-norm within 6% of slot mod * (0.661 * ilvl +
// 2.92)). The staff gets Forever's caster weapon rate, 2.21 * L(76) = 193.4, and carries what the armor
// leaves of each stat. The armor values are the Balance set's leather.
//
// The mix is the Restoration druid row of orchestration/gear_targets.md (Healers), from two Wowhead
// Restoration druid BiS lists (WoW Forever Phase 1 and Classic pre-raid;
// orchestration/tools/healers/druid_restoration_mix.py and druid_restoration_synth.py): Stamina 14.8% (the
// lists' own share), Spirit 10% (standards 5), no hit (heals don't miss), and Intellect 18.7%, spell power
// 17.5%, healing 35.2%, crit 1.0% and MP5 2.8% in the lists' proportions. The druid lists lean on healing
// and Spirit more than the shaman's and carry almost no crit: the Classic pre-raid one is 46% healing. A
// stat's share is its weight times its amount over the set's weighted total (the trinkets left out). The
// set:
//
//	stat      amount   share   target
//	Stamina      134   14.9%   14.8 +- 1.5
//	Intellect    167   18.5%   18.7 +- 3
//	Spirit        89    9.9%   10
//	SP           183   17.5%   17.5 +- 3
//	Healing      709   35.4%   35.2 +- 3
//	Crit           9    1.0%    1.0 +- 3
//	MP5           11    2.9%    2.8 +- 3
//
// Per piece (target, weighted linear total, 1.7-norm value / shape target, stats). The armor comes to 707.9
// and the staff to 193.9:
//
//	Head         80.3   80.1 (-0.3%)  46.2 /  46.5 (-0.8%)  14 Sta, 18 Int, 24 SP, 61 Heal
//	Neck         45.8   45.8 (-0.2%)  29.1 /  28.0 (+3.8%)  15 Sta, 15 Spi, 35 Heal
//	Shoulder     58.7   58.5 (-0.3%)  34.1 /  34.9 (-2.3%)  11 Sta, 10 Spi, 19 SP, 47 Heal
//	Back         45.8   45.9 (+0.3%)  27.6 /  28.0 (-1.4%)  5 Sta, 10 Int, 14 SP, 42 Heal
//	Chest        80.3   80.2 (-0.1%)  46.6 /  46.5 (+0.0%)  16 Sta, 20 Int, 15 Spi, 65 Heal
//	Wrist        45.8   45.8 (-0.1%)  27.6 /  28.0 (-1.5%)  9 Sta, 11 Int, 6 Spi, 44 Heal
//	Hands        58.7   58.5 (-0.2%)  34.4 /  34.9 (-1.4%)  13 Int, 8 Spi, 18 SP, 49 Heal
//	Waist        63.0   62.8 (-0.3%)  36.5 /  37.4 (-2.3%)  10 Sta, 14 Int, 19 SP, 50 Heal
//	Legs         80.3   80.3 (-0.0%)  47.1 /  46.5 (+1.2%)  20 Int, 27 SP, 61 Heal, 4 MP5
//	Feet         58.7   58.6 (-0.1%)  34.1 /  34.9 (-2.2%)  12 Sta, 15 Int, 10 Spi, 48 Heal
//	Finger       45.8   45.8 (-0.1%)  29.1 /  28.0 (+3.9%)  15 Int, 18 SP, 34 Heal
//	Finger II    45.8   45.6 (-0.5%)  29.0 /  28.0 (+3.5%)  15 Sta, 36 Heal, 6 MP5
//	Staff       193.4  193.9 (+0.3%)                          27 Sta, 31 Int, 25 Spi, 44 SP, 137 Heal, 9 Crit, 1 MP5
//
// Forever's Tier 1 Restoration druid set (Molten Core; 1300943, 1301006, 1301077 and 1301250 in the
// client) goes on the head and chest, so only its 2 piece bonus (+10 Spirit) is live (standards 6). Its 3
// piece cuts Barkskin's cooldown, its 4 piece is 26 healing, its 5 piece 3 sec off Swiftmend's cooldown.

func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, restorationDruidSyntheticItems...)
}

const restorationDruidSetName = "Restoration Druid Synthetic (Phase 1)"

var restorationDruidSyntheticItems = []*proto.UIItem{
	// Head, item level 66.
	{
		Id: 990401, Name: "Restoration Druid Synthetic Head (Phase 1)", Icon: "inv_helmet_09",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:          stats.Stats{stats.Stamina: 14, stats.Intellect: 18, stats.SpellPower: 24, stats.HealingPower: 61, stats.Armor: 163}.ToFloatArray(),
		Ilvl:           66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		SetName:        restorationDruidSetName,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Neck, item level 71.
	{
		Id: 990402, Name: "Restoration Druid Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_17",
		Type:           proto.ItemType_ItemTypeNeck,
		Stats:          stats.Stats{stats.Stamina: 15, stats.Spirit: 15, stats.HealingPower: 35}.ToFloatArray(),
		Ilvl:           71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Shoulder, item level 66.
	{
		Id: 990403, Name: "Restoration Druid Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_07",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:          stats.Stats{stats.Stamina: 11, stats.Spirit: 10, stats.SpellPower: 19, stats.HealingPower: 47, stats.Armor: 159}.ToFloatArray(),
		Ilvl:           66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Back, item level 71.
	{
		Id: 990404, Name: "Restoration Druid Synthetic Back (Phase 1)", Icon: "inv_misc_cape_16",
		Type:           proto.ItemType_ItemTypeBack,
		Stats:          stats.Stats{stats.Stamina: 5, stats.Intellect: 10, stats.SpellPower: 14, stats.HealingPower: 42, stats.Armor: 55}.ToFloatArray(),
		Ilvl:           71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Chest, item level 66.
	{
		Id: 990405, Name: "Restoration Druid Synthetic Chest (Phase 1)", Icon: "inv_chest_leather_06",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:          stats.Stats{stats.Stamina: 16, stats.Intellect: 20, stats.Spirit: 15, stats.HealingPower: 65, stats.Armor: 200}.ToFloatArray(),
		Ilvl:           66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		SetName:        restorationDruidSetName,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Wrist, item level 71.
	{
		Id: 990406, Name: "Restoration Druid Synthetic Wrist (Phase 1)", Icon: "inv_bracer_03",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:          stats.Stats{stats.Stamina: 9, stats.Intellect: 11, stats.Spirit: 6, stats.HealingPower: 44, stats.Armor: 88}.ToFloatArray(),
		Ilvl:           71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Hands, item level 66.
	{
		Id: 990407, Name: "Restoration Druid Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_08",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:          stats.Stats{stats.Intellect: 13, stats.Spirit: 8, stats.SpellPower: 18, stats.HealingPower: 49, stats.Armor: 125}.ToFloatArray(),
		Ilvl:           66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Waist, item level 71.
	{
		Id: 990408, Name: "Restoration Druid Synthetic Waist (Phase 1)", Icon: "inv_belt_06",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:          stats.Stats{stats.Stamina: 10, stats.Intellect: 14, stats.SpellPower: 19, stats.HealingPower: 50, stats.Armor: 113}.ToFloatArray(),
		Ilvl:           71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Legs, item level 66.
	{
		Id: 990409, Name: "Restoration Druid Synthetic Legs (Phase 1)", Icon: "inv_pants_06",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:          stats.Stats{stats.Intellect: 20, stats.SpellPower: 27, stats.HealingPower: 61, stats.MP5: 4, stats.Armor: 175}.ToFloatArray(),
		Ilvl:           66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Feet, item level 66.
	{
		Id: 990410, Name: "Restoration Druid Synthetic Feet (Phase 1)", Icon: "inv_boots_08",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:          stats.Stats{stats.Stamina: 12, stats.Intellect: 15, stats.Spirit: 10, stats.HealingPower: 48, stats.Armor: 138}.ToFloatArray(),
		Ilvl:           66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Finger, item level 71.
	{
		Id: 990411, Name: "Restoration Druid Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_38",
		Type:           proto.ItemType_ItemTypeFinger,
		Stats:          stats.Stats{stats.Intellect: 15, stats.SpellPower: 18, stats.HealingPower: 34}.ToFloatArray(),
		Ilvl:           71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Finger II, item level 71.
	{
		Id: 990412, Name: "Restoration Druid Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_29",
		Type:           proto.ItemType_ItemTypeFinger,
		Stats:          stats.Stats{stats.Stamina: 15, stats.HealingPower: 36, stats.MP5: 6}.ToFloatArray(),
		Ilvl:           71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Staff, item level 76, with Staff of Dominance's (70, 18842) damage and speed.
	{
		Id: 990413, Name: "Restoration Druid Synthetic Staff (Phase 1)", Icon: "inv_staff_13",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand,
		Stats:           stats.Stats{stats.Stamina: 27, stats.Intellect: 31, stats.Spirit: 25, stats.SpellPower: 44, stats.HealingPower: 137, stats.MP5: 1}.ToFloatArray(),
		CritRating:      9,
		WeaponDamageMin: 126, WeaponDamageMax: 205, WeaponSpeed: 2.9,
		Ilvl: 76, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
}
