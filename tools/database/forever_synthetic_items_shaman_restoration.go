package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The synthetic Phase 1 Restoration shaman items, ids 990200 to 990299: the healer's set
// (docs/healing_sim.md, H2). See forever_synthetic_items.go for what these items are and why we need them.
// The trinkets are the Elemental set's (990114 and 990115), whose spell power heals as it damages.
//
// The budget is the Elemental set's, standards 1b for a pure caster: weights 1 a point for Stamina,
// Intellect, Spirit and crit rating (14 is 1%), spell power 0.86, healing on top of spell power 0.45, MP5
// 2.4, armor free. Each armor piece carries slot mod * L(ilvl) * f in weighted linear stats within 1%
// (L = 1.125 * ilvl + 2.0, f 1.053 / 1.026 / 0.995), at item level 66 for the head, shoulders, chest,
// hands, legs and feet and 71 for the neck, back, wrist, waist and rings, and passes the shape check (its
// 1.7-norm within 6% of slot mod * (0.661 * ilvl + 2.92)). The staff gets Forever's caster weapon rate,
// 2.21 * L(76) = 193.4, and carries what the armor leaves of each stat. Spell power and healing both heal
// (core.Spell.HealingPower), healing for about half the budget, so a healer set carries both.
//
// The mix is the Restoration shaman row of orchestration/gear_targets.md (2026-11-03, the user's ruling:
// "yes, healer synthetic gear", from Wowhead's lists since wowsims/sod has no healer BiS sets). Stamina 10.8%
// (the lists' own share), Spirit 10% (standards 5), no hit (heals don't miss), and Intellect 17.5%, spell
// power 22.2%, healing 27.1%, crit 7.0% and MP5 5.4% in the lists' proportions. A stat's share is its weight
// times its amount over the set's weighted total (the trinkets left out). The set:
//
//	stat      amount   share   target
//	Stamina       96   10.6%   10.8 +- 1.5
//	Intellect    158   17.5%   17.5 +- 3
//	Spirit        91   10.1%   10
//	SP           235   22.4%   22.2 +- 3
//	Healing      539   26.9%   27.1 +- 3
//	Crit          63    7.0%    7.0 +- 3
//	MP5           21    5.6%    5.4 +- 3
//
// Per piece (target, weighted linear total, 1.7-norm value / shape target, stats). The armor comes to 709.8
// and the staff to 193.2:
//
//	Head         80.3   80.3 (+0.1%)  46.5 /  46.5 (-0.2%)  14 Sta, 20 Int, 63 Heal, 18 Crit
//	Neck         45.8   45.9 (+0.1%)  27.3 /  28.0 (-2.5%)  5 Sta, 17 SP, 37 Heal, 4 MP5
//	Shoulder     58.7   58.6 (-0.1%)  34.0 /  34.9 (-2.5%)  13 Int, 9 Spi, 19 SP, 45 Heal
//	Back         45.8   45.9 (+0.2%)  27.4 /  28.0 (-2.1%)  10 Int, 5 Spi, 15 SP, 40 Heal
//	Chest        80.3   80.0 (-0.4%)  46.0 /  46.5 (-1.1%)  14 Sta, 20 Int, 19 Spi, 60 Heal
//	Wrist        45.8   45.7 (-0.4%)  27.3 /  28.0 (-2.8%)  10 Int, 16 SP, 38 Heal, 2 MP5
//	Hands        58.7   58.9 (+0.4%)  34.5 /  34.9 (-1.2%)  13 Int, 20 SP, 46 Heal, 8 Crit
//	Waist        63.0   63.2 (+0.3%)  36.7 /  37.4 (-1.9%)  9 Sta, 15 Int, 22 SP, 45 Heal
//	Legs         80.3   80.4 (+0.1%)  46.3 /  46.5 (-0.6%)  13 Sta, 19 Int, 28 SP, 54 Heal
//	Feet         58.7   58.9 (+0.3%)  34.4 /  34.9 (-1.4%)  8 Sta, 13 Spi, 21 SP, 44 Heal
//	Finger       45.8   45.9 (+0.2%)  27.4 /  28.0 (-2.1%)  10 Int, 15 SP, 40 Heal, 5 Crit
//	Finger II    45.8   46.2 (+0.9%)  28.3 /  28.0 (+0.7%)  5 Sta, 11 Spi, 24 SP, 4 MP5
//	Staff       193.4  193.2 (-0.1%)                          28 Sta, 28 Int, 34 Spi, 38 SP, 27 Heal, 32 Crit, 11 MP5

// Forever's Tier 1 Restoration shaman set (Molten Core; 1300960, 1301059, 1301093 and 1301713 in the
// client) goes on the head and chest, so only its 2 piece bonus (+10 Spirit) is live (standards 6). Its 3
// piece widens the friendly totems' radius, its 4 piece is 26 healing, its 5 piece 1 sec off Riptide's
// cooldown (sim/shaman/item_sets_pve.go).

func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, restorationShamanSyntheticItems...)
}

const restorationShamanSetName = "Restoration Synthetic (Phase 1)"

var restorationShamanSyntheticItems = []*proto.UIItem{
	// Head, item level 66.
	{
		Id: 990201, Name: "Restoration Synthetic Head (Phase 1)", Icon: "inv_helmet_09",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:      stats.Stats{stats.Stamina: 14, stats.Intellect: 20, stats.HealingPower: 63, stats.Armor: 343}.ToFloatArray(),
		CritRating: 18,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		SetName:        restorationShamanSetName,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Neck, item level 71.
	{
		Id: 990202, Name: "Restoration Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_17",
		Type:  proto.ItemType_ItemTypeNeck,
		Stats: stats.Stats{stats.Stamina: 5, stats.SpellPower: 17, stats.HealingPower: 37, stats.MP5: 4}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Shoulder, item level 66.
	{
		Id: 990203, Name: "Restoration Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_29",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Intellect: 13, stats.Spirit: 9, stats.SpellPower: 19, stats.HealingPower: 45, stats.Armor: 399}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Back, item level 71.
	{
		Id: 990204, Name: "Restoration Synthetic Back (Phase 1)", Icon: "inv_misc_cape_16",
		Type:  proto.ItemType_ItemTypeBack,
		Stats: stats.Stats{stats.Intellect: 10, stats.Spirit: 5, stats.SpellPower: 15, stats.HealingPower: 40, stats.Armor: 54}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Chest, item level 66.
	{
		Id: 990205, Name: "Restoration Synthetic Chest (Phase 1)", Icon: "inv_chest_chain_11",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Stamina: 14, stats.Intellect: 20, stats.Spirit: 19, stats.HealingPower: 60, stats.Armor: 422}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		SetName:        restorationShamanSetName,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Wrist, item level 71.
	{
		Id: 990206, Name: "Restoration Synthetic Wrist (Phase 1)", Icon: "inv_bracer_16",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Intellect: 10, stats.SpellPower: 16, stats.HealingPower: 38, stats.MP5: 2, stats.Armor: 185}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Hands, item level 66.
	{
		Id: 990207, Name: "Restoration Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_11",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:      stats.Stats{stats.Intellect: 13, stats.SpellPower: 20, stats.HealingPower: 46, stats.Armor: 264}.ToFloatArray(),
		CritRating: 8,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Waist, item level 71.
	{
		Id: 990208, Name: "Restoration Synthetic Waist (Phase 1)", Icon: "inv_belt_14",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Stamina: 9, stats.Intellect: 15, stats.SpellPower: 22, stats.HealingPower: 45, stats.Armor: 255}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Legs, item level 66.
	{
		Id: 990209, Name: "Restoration Synthetic Legs (Phase 1)", Icon: "inv_pants_mail_15",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Stamina: 13, stats.Intellect: 19, stats.SpellPower: 28, stats.HealingPower: 54, stats.Armor: 369}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Feet, item level 66.
	{
		Id: 990210, Name: "Restoration Synthetic Feet (Phase 1)", Icon: "inv_boots_plate_06",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Stamina: 8, stats.Spirit: 13, stats.SpellPower: 21, stats.HealingPower: 44, stats.Armor: 290}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Finger, item level 71.
	{
		Id: 990211, Name: "Restoration Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_38",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Intellect: 10, stats.SpellPower: 15, stats.HealingPower: 40}.ToFloatArray(),
		CritRating: 5,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Finger II, item level 71.
	{
		Id: 990212, Name: "Restoration Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_29",
		Type:  proto.ItemType_ItemTypeFinger,
		Stats: stats.Stats{stats.Stamina: 5, stats.Spirit: 11, stats.SpellPower: 24, stats.MP5: 4}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Staff, item level 76, with Staff of Dominance's (70, 18842) damage and speed.
	{
		Id: 990213, Name: "Restoration Synthetic Staff (Phase 1)", Icon: "inv_staff_13",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand,
		Stats:           stats.Stats{stats.Stamina: 28, stats.Intellect: 28, stats.Spirit: 34, stats.SpellPower: 38, stats.HealingPower: 27, stats.MP5: 11}.ToFloatArray(),
		CritRating:      32,
		WeaponDamageMin: 126, WeaponDamageMax: 205, WeaponSpeed: 2.9,
		Ilvl: 76, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
}
