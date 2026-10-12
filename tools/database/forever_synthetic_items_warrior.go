package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Synthetic level 60 Phase 1 warrior items, made the same way as the shaman's in
// forever_synthetic_items.go (read its comment first). They are not in the game.
//
// Each piece takes the place of an item of the Era Phase 1 DPS warrior BiS (ui/warrior/gear_sets/
// phase_1.gear.json) at the shared item level table (orchestration/standards.md): head, shoulders,
// chest, hands, legs and feet at 66, neck, back, wrist, waist and both rings at 71, a two-hander at
// 76, each one-hander at 71 and the ranged slot at 71. The back replaces Cape of the Black Baron
// rather than the list's Puissant Cape, which drops from Azuregos in Phase 2. The budget is the
// standard all the classes' synthetic items share:
//
//	(sum over the stats of (value * weight)^1.5)^(2/3) = (1.2 * ilvl - 36.4) * slot
//
// with the slot modifiers 1 (head, chest, legs, two-hand), 0.77 (shoulder, hands, waist, feet), 0.5625
// (neck, back, wrist, finger), 0.42 (one-hand) and 0.32 (ranged). Strength, Agility and Stamina weigh 1,
// attack power 0.5, hit rating 1 (10 to the percent) and crit rating 1 (14 to the percent). Armor is
// free. The weapons' damage and speed are the Era weapons' and not charged. Below, each piece's budget
// and what its stats use:
//
//	Head       66  x1       42.8  42.7  str 8 agi 4 sta 5 hit 20 crit 28
//	Neck       71  x0.5625  27.4  27.3  agi 10 sta 9 hit 10 crit 14
//	Shoulder   66  x0.77    33.0  32.5  str 17 ap 26 sta 11 hit 10
//	Back       71  x0.5625  27.4  27.3  ap 31 agi 8 sta 9 hit 10
//	Chest      66  x1       42.8  42.5  str 21 agi 15 sta 17 crit 14
//	Wrist      71  x0.5625  27.4  27.2  str 23 sta 10
//	Hands      66  x0.77    33.0  32.8  str 13 agi 14 sta 11 crit 14
//	Waist      71  x0.77    37.6  36.8  str 22 sta 11 hit 10 crit 14
//	Legs       66  x1       42.8  42.4  str 20 agi 17 sta 16 crit 14
//	Feet       66  x0.77    33.0  32.9  str 15 agi 14 sta 13 hit 10
//	Finger     71  x0.5625  27.4  26.9  agi 15 sta 8 ap 18 hit 10
//	Finger II  71  x0.5625  27.4  27.2  ap 28 str 7 sta 7 crit 14
//	Ranged     71  x0.32    15.6  15.1  agi 9 hit 10
//	Main Hand  71  x0.42    20.5  20.0  str 6 agi 5 sta 5 crit 14
//	Off Hand   71  x0.42    20.5  20.0  str 6 agi 5 sta 5 crit 14
//	Two-Hand   76  x1       54.8  54.5  str 25 agi 14 sta 18 crit 28
//
// The stat mix: Strength first, then Agility, Stamina on most pieces and attack power on the
// shoulders, back and rings (the Era pieces there carry it). No Spirit, the warrior runs on Rage. The
// armor, rings and bow carry 90 hit rating and 112 crit rating, and the weapons no hit: 9% hit is the
// cap for special attacks against a level 63 boss (8% miss and 1% hit suppression), and no warrior
// talent adds main hand hit (Furious Precision is off hand only). Each Fury sword adds 14 crit, so the
// Fury set has 9% hit and 10% crit, and the two-hander adds 28 crit, so the two-hand set has 9% hit and
// 10% crit. The plate has Battlegear of Might's armor at item level 66, the waist Onslaught Girdle's.
//
// Arms and Fury share the armor, the rings and the bow, because their Era Phase 1 BiS lists are the same
// outside the weapons: both want Strength, crit and hit. They differ in the weapons only. Fury dual wields
// a main hand and an off hand sword in place of Vis'kag the Bloodletter and Brutality Blade, Arms (and a
// two-hand Fury) a two hand sword in place of Bonereaver's Edge, each with the damage and speed of the
// Era weapon. All three are swords like the Era ones, which is what Human's sword racial and
// Weaponmaster's extra attacks see.
//
// The head and legs form a set with the bonuses of Battlegear of Glory, Forever's Molten Core tier set
// for Arms and Fury (set 2102, items 280899 to 280904, client 70338, see sim/warrior/item_sets_pve.go).
// Only two pieces carry it, so only the 2-piece bonus (1% attack speed) is live, the same for every
// class. The tier pieces use its icons. The generator script lived in the session scratchpad
// (synth_warrior.py), the numbers above are its output.
func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, foreverSyntheticWarriorItems...)
}

var foreverSyntheticWarriorItems = []*proto.UIItem{
	// Head, item level 66, in place of Lionheart Helm (61, 12640).
	{
		Id: 991001, Name: "Warrior Synthetic Head (Phase 1)", Icon: "inv_helm_plate_raidwarriorhyjalc60_d_01",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypePlate,
		Stats:     stats.Stats{stats.Strength: 8, stats.Agility: 4, stats.Stamina: 5, stats.Armor: 608}.ToFloatArray(),
		HitRating: 20, CritRating: 28,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
		SetName:        "Warrior Synthetic (Phase 1)",
	},
	// Neck, item level 71, in place of Onyxia Tooth Pendant (74, 18404).
	{
		Id: 991002, Name: "Warrior Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_09",
		Type:      proto.ItemType_ItemTypeNeck,
		Stats:     stats.Stats{stats.Agility: 10, stats.Stamina: 9}.ToFloatArray(),
		HitRating: 10, CritRating: 14,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Shoulder, item level 66, in place of Truestrike Shoulders (61, 12927).
	{
		Id: 991003, Name: "Warrior Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_plate_raidwarriorhyjalc60_d_01",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypePlate,
		Stats:     stats.Stats{stats.Strength: 17, stats.AttackPower: 26, stats.Stamina: 11, stats.Armor: 562}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Back, item level 71, in place of Cape of the Black Baron (63, 13340).
	{
		Id: 991004, Name: "Warrior Synthetic Back (Phase 1)", Icon: "inv_misc_cape_20",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.AttackPower: 31, stats.Agility: 8, stats.Stamina: 9, stats.Armor: 45}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Chest, item level 66, in place of Savage Gladiator Chain (57, 11726).
	{
		Id: 991005, Name: "Warrior Synthetic Chest (Phase 1)", Icon: "inv_chest_plate_raidwarriorhyjalc60_d_01",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypePlate,
		Stats:      stats.Stats{stats.Strength: 21, stats.Agility: 15, stats.Stamina: 17, stats.Armor: 749}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Wrist, item level 71, in place of Wristguards of Stability (65, 19146).
	{
		Id: 991006, Name: "Warrior Synthetic Wrist (Phase 1)", Icon: "inv_bracer_19",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypePlate,
		Stats: stats.Stats{stats.Strength: 23, stats.Stamina: 10, stats.Armor: 328}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Hands, item level 66, in place of Edgemaster's Handguards (49, 14551).
	{
		Id: 991007, Name: "Warrior Synthetic Hands (Phase 1)", Icon: "inv_glove_plate_raidwarriorhyjalc60_d_01",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypePlate,
		Stats:      stats.Stats{stats.Strength: 13, stats.Agility: 14, stats.Stamina: 11, stats.Armor: 468}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Waist, item level 71, in place of Onslaught Girdle (78, 19137).
	{
		Id: 991008, Name: "Warrior Synthetic Waist (Phase 1)", Icon: "inv_belt_09",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypePlate,
		Stats:     stats.Stats{stats.Strength: 22, stats.Stamina: 11, stats.Armor: 494}.ToFloatArray(),
		HitRating: 10, CritRating: 14,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Legs, item level 66, in place of Cloudkeeper Legplates (62, 14554).
	{
		Id: 991009, Name: "Warrior Synthetic Legs (Phase 1)", Icon: "inv_pant_plate_raidwarriorhyjalc60_d_01",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypePlate,
		Stats:      stats.Stats{stats.Strength: 20, stats.Agility: 17, stats.Stamina: 16, stats.Armor: 655}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
		SetName:        "Warrior Synthetic (Phase 1)",
	},
	// Feet, item level 66, in place of Bloodmail Boots (61, 14616).
	{
		Id: 991010, Name: "Warrior Synthetic Feet (Phase 1)", Icon: "inv_boot_plate_raidwarriorhyjalc60_d_01",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypePlate,
		Stats:     stats.Stats{stats.Strength: 15, stats.Agility: 14, stats.Stamina: 13, stats.Armor: 515}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Finger, item level 71, in place of Band of Accuria (78, 17063).
	{
		Id: 991011, Name: "Warrior Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_15",
		Type:      proto.ItemType_ItemTypeFinger,
		Stats:     stats.Stats{stats.Agility: 15, stats.Stamina: 8, stats.AttackPower: 18}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Finger II, item level 71, in place of Quick Strike Ring (67, 18821).
	{
		Id: 991012, Name: "Warrior Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_16",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.AttackPower: 28, stats.Strength: 7, stats.Stamina: 7}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Ranged, item level 71, in place of Striker's Mark (69, 17069).
	{
		Id: 991013, Name: "Warrior Synthetic Ranged (Phase 1)", Icon: "inv_weapon_bow_08",
		Type:             proto.ItemType_ItemTypeRanged,
		RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeBow,
		// Striker's Mark damage and speed.
		WeaponDamageMin: 69, WeaponDamageMax: 129, WeaponSpeed: 2.5,
		Stats:     stats.Stats{stats.Agility: 9}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Main Hand, item level 71, in place of Vis'kag the Bloodletter (74, 17075).
	{
		Id: 991014, Name: "Fury Synthetic Main Hand (Phase 1)", Icon: "inv_sword_38",
		Type:       proto.ItemType_ItemTypeWeapon,
		WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand,
		// Vis'kag the Bloodletter damage and speed.
		WeaponDamageMin: 100, WeaponDamageMax: 187, WeaponSpeed: 2.6,
		Stats:      stats.Stats{stats.Strength: 6, stats.Agility: 5, stats.Stamina: 5}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Off Hand, item level 71, in place of Brutality Blade (70, 18832).
	{
		Id: 991015, Name: "Fury Synthetic Off Hand (Phase 1)", Icon: "inv_sword_44",
		Type:       proto.ItemType_ItemTypeWeapon,
		WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand,
		// Brutality Blade damage and speed.
		WeaponDamageMin: 90, WeaponDamageMax: 168, WeaponSpeed: 2.5,
		Stats:      stats.Stats{stats.Strength: 6, stats.Agility: 5, stats.Stamina: 5}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Two-Hand, item level 76, in place of Bonereaver's Edge (77, 17076).
	{
		Id: 991016, Name: "Arms Synthetic Two-Hand (Phase 1)", Icon: "inv_sword_12",
		Type:       proto.ItemType_ItemTypeWeapon,
		WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeTwoHand,
		// Bonereaver's Edge damage and speed.
		WeaponDamageMin: 206, WeaponDamageMax: 310, WeaponSpeed: 3.4,
		Stats:      stats.Stats{stats.Strength: 25, stats.Agility: 14, stats.Stamina: 18}.ToFloatArray(),
		CritRating: 28,
		Ilvl:       76, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
}
