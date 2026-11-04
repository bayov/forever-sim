package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The synthetic Phase 1 healing priest items, ids 990500 to 990599: the Holy and Discipline priest's set
// (docs/healing_priest.md). See forever_synthetic_items.go for what these items are and why we need them.
// One set for both specs: Wowhead's Holy and Discipline guides rank the same stats (healing, Intellect,
// Spirit, MP5, then crit). The trinkets are the Shadow and Smite sets' (995033 and 995034), whose spell
// power heals as it damages.
//
// The budget is the Restoration shaman's, standards 1b for a pure caster: weights 1 a point for Stamina,
// Intellect, Spirit and crit rating (14 is 1%), spell power 0.86, healing on top of spell power 0.45, MP5
// 2.4, armor free. Each armor piece carries slot mod * L(ilvl) * f in weighted linear stats within 1%
// (L = 1.125 * ilvl + 2.0, f 1.053 / 1.026 / 0.995), at item level 66 for the head, shoulders, chest,
// hands, legs and feet and 71 for the neck, back, wrist, waist and rings, and passes the shape check (its
// 1.7-norm within 6% of slot mod * (0.661 * ilvl + 2.92)). The staff gets Forever's caster weapon rate,
// 2.21 * L(76) = 193.4, and carries what the armor leaves of each stat; the wand (21.3 by
// standards 1b) carries nothing here, as the other healers' slot 16 relics. The
// armor values are the Shadow set's cloth.
//
// The mix is the healing priest row of orchestration/gear_targets.md (Healers): Wowhead's two WoW Forever
// priest guides have no items in their slot tables yet, so the Classic pre-raid healing priest list is
// averaged with the one WoW Forever healer list that has items, the Restoration shaman's (Forever's healer
// itemization, healing carried as spell power). Stamina 12.1% (the lists' own share), Spirit 10%
// (standards 5), no hit, and Intellect 15.3%, spell power 27.5%, healing 21.1%, crit 5.9% and MP5 8.1% in
// the lists' proportions. The set (orchestration/tools/healers/priest_healing_synth.py):
//
//	stat      amount   share   target
//	Stamina      112   12.1%   12.1 +- 1.5
//	Intellect    141   15.3%   15.3 +- 3
//	Spirit        93   10.1%   10
//	SP           291   27.1%   27.5 +- 3
//	Healing      438   21.3%   21.1 +- 3
//	Crit          54    5.8%    5.9 +- 3
//	MP5           32    8.3%    8.1 +- 3
//
// Per piece (target, weighted linear total, 1.7-norm value / shape target, stats). The armor comes to 708.8,
// the staff to 194.0 and the wand to 0 (stat-less, like the other healers' relics):
//
//	Head         80.3   80.4 (+0.1%)  45.6 /  46.5 (-2.0%)  4 Sta, 14 Int, 36 SP, 41 Heal, 13 Crit
//	Neck         45.8   45.7 (-0.3%)  27.6 /  28.0 (-1.7%)  10 Sta, 12 Spi, 42 Heal, 2 MP5
//	Shoulder     58.7   58.6 (-0.1%)  33.8 /  34.9 (-3.2%)  12 Sta, 14 Int, 24 SP, 5 MP5
//	Back         45.8   46.1 (+0.5%)  27.7 /  28.0 (-1.2%)  5 Sta, 10 Int, 22 SP, 27 Heal
//	Chest        80.3   80.3 (-0.0%)  46.2 /  46.5 (-0.7%)  14 Sta, 18 Spi, 31 SP, 48 Heal
//	Wrist        45.8   45.6 (-0.4%)  27.5 /  28.0 (-2.1%)  5 Int, 10 Spi, 22 SP, 26 Heal
//	Hands        58.7   58.7 (+0.0%)  34.3 /  34.9 (-1.6%)  12 Int, 25 SP, 36 Heal, 9 Crit
//	Waist        63.0   63.0 (-0.1%)  36.7 /  37.4 (-1.7%)  13 Int, 26 SP, 40 Heal, 4 MP5
//	Legs         80.3   80.2 (-0.1%)  45.4 /  46.5 (-2.5%)  12 Sta, 14 Int, 36 SP, 41 Heal, 2 MP5
//	Feet         58.7   58.5 (-0.3%)  34.0 /  34.9 (-2.6%)  10 Sta, 13 Int, 14 Spi, 25 SP
//	Finger       45.8   46.1 (+0.5%)  27.7 /  28.0 (-1.2%)  10 Int, 22 SP, 27 Heal, 5 Crit
//	Finger II    45.8   45.7 (-0.3%)  27.6 /  28.0 (-1.7%)  10 Sta, 12 Spi, 42 Heal, 2 MP5
//	Staff       193.4  194.0 (+0.3%)                          35 Sta, 30 Int, 27 Spi, 12 SP, 53 Heal, 27 Crit, 17 MP5
//	Wand            -      0                                   no stats (healer_fairness-20261104-3)

func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, healingPriestSyntheticItems...)
}

var healingPriestSyntheticItems = []*proto.UIItem{
	// Head, item level 66.
	{
		Id: 990501, Name: "Holy Synthetic Head (Phase 1)", Icon: "inv_helmet_34",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 4, stats.Intellect: 14, stats.SpellPower: 36, stats.HealingPower: 41, stats.Armor: 83}.ToFloatArray(),
		CritRating: 13,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Neck, item level 71.
	{
		Id: 990502, Name: "Holy Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_17",
		Type:  proto.ItemType_ItemTypeNeck,
		Stats: stats.Stats{stats.Stamina: 10, stats.Spirit: 12, stats.HealingPower: 42, stats.MP5: 2}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Shoulder, item level 66.
	{
		Id: 990503, Name: "Holy Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_02",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 12, stats.Intellect: 14, stats.SpellPower: 24, stats.MP5: 5, stats.Armor: 76}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Back, item level 71.
	{
		Id: 990504, Name: "Holy Synthetic Back (Phase 1)", Icon: "inv_misc_cape_16",
		Type:  proto.ItemType_ItemTypeBack,
		Stats: stats.Stats{stats.Stamina: 5, stats.Intellect: 10, stats.SpellPower: 22, stats.HealingPower: 27, stats.Armor: 55}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Chest, item level 66.
	{
		Id: 990505, Name: "Holy Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_08",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 14, stats.Spirit: 18, stats.SpellPower: 31, stats.HealingPower: 48, stats.Armor: 102}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Wrist, item level 71.
	{
		Id: 990506, Name: "Holy Synthetic Wrist (Phase 1)", Icon: "inv_bracer_09",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Intellect: 5, stats.Spirit: 10, stats.SpellPower: 22, stats.HealingPower: 26, stats.Armor: 44}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Hands, item level 66.
	{
		Id: 990507, Name: "Holy Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_14",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Intellect: 12, stats.SpellPower: 25, stats.HealingPower: 36, stats.Armor: 63}.ToFloatArray(),
		CritRating: 9,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Waist, item level 71.
	{
		Id: 990508, Name: "Holy Synthetic Waist (Phase 1)", Icon: "inv_belt_08",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Intellect: 13, stats.SpellPower: 26, stats.HealingPower: 40, stats.MP5: 4, stats.Armor: 61}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Legs, item level 66.
	{
		Id: 990509, Name: "Holy Synthetic Legs (Phase 1)", Icon: "inv_pants_08",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 12, stats.Intellect: 14, stats.SpellPower: 36, stats.HealingPower: 41, stats.MP5: 2, stats.Armor: 89}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Feet, item level 66.
	{
		Id: 990510, Name: "Holy Synthetic Feet (Phase 1)", Icon: "inv_boots_cloth_05",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 10, stats.Intellect: 13, stats.Spirit: 14, stats.SpellPower: 25, stats.Armor: 70}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Finger, item level 71.
	{
		Id: 990511, Name: "Holy Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_29",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Intellect: 10, stats.SpellPower: 22, stats.HealingPower: 27}.ToFloatArray(),
		CritRating: 5,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Finger II, item level 71.
	{
		Id: 990512, Name: "Holy Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_36",
		Type:  proto.ItemType_ItemTypeFinger,
		Stats: stats.Stats{stats.Stamina: 10, stats.Spirit: 12, stats.HealingPower: 42, stats.MP5: 2}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},

	// Staff, item level 76, with Redemption's (61, 22406) damage and speed.
	{
		Id: 990513, Name: "Holy Synthetic Staff (Phase 1)", Icon: "inv_staff_13",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand,
		Stats:           stats.Stats{stats.Stamina: 35, stats.Intellect: 30, stats.Spirit: 27, stats.SpellPower: 12, stats.HealingPower: 53, stats.MP5: 17}.ToFloatArray(),
		CritRating:      27,
		WeaponDamageMin: 87, WeaponDamageMax: 132, WeaponSpeed: 2.1,
		Ilvl: 76, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Wand, item level 71, with Wand of Eternal Light's (57, 22254) damage, speed and school. It carries
	// no stats (healer_fairness-20261104-3): the shaman, paladin and druid healers' slot 16 is a stat-less
	// relic, so a budgeted wand (6 Int, 10 SP, 15 Heal, 21.4 weighted) put the priest 2.4% over the
	// healers' 902 budget. A healer's wand does nothing for it either way.
	{
		Id: 990514, Name: "Holy Synthetic Wand (Phase 1)", Icon: "inv_wand_07",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeWand,
		SpellSchool:     proto.SpellSchool_SpellSchoolHoly,
		WeaponDamageMin: 58, WeaponDamageMax: 109, WeaponSpeed: 1.5,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
}
