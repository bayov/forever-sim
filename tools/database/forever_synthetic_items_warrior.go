package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Synthetic level 60 Phase 1 warrior items, made the same way as the shaman's in
// forever_synthetic_items.go (read its comment first). They are not in the game.
//
// Each piece takes the place of an item of the Era Phase 1 DPS warrior BiS (ui/warrior/gear_sets/
// phase_1.gear.json) at the shared item level table (orchestration/standards.md 2): head, shoulders,
// chest, hands, legs and feet at 66, neck, back, wrist, waist and both rings at 71, a two-hander at
// 76, each one-hander and the bow at 71. The back replaces Cape of the Black Baron rather than the
// list's Puissant Cape, which drops from Azuregos in Phase 2.
//
// The budget is standards 1b, physical melee: L(ilvl) = 1.125 x ilvl + 2.0, and each armor piece's
// weighted linear stat total is slot mod x L x f (1 x 0.950 for head, chest and legs, 0.75 x 0.959
// for shoulders, hands, waist and feet, 0.5625 x 0.911 for neck, back, wrist and rings), within 1%.
// Each armor piece also passes the shape check: its 1.7-norm value, (sum (weight x amount)^1.7)^(1/1.7),
// is within 6% of slot mod x (0.661 x ilvl + 2.92), which means three stat lines a piece. The bow is
// 0.24 x L(71), each one-hander 0.39 x L(71) and the two-hander 0.90 x L(76). Strength, Agility and
// Stamina weigh 1, attack power 0.5, hit rating 1 (10 to the percent), crit rating 1 (14 to the
// percent). Armor is free, and the weapons' damage and speed are the Era weapons' and free. Below,
// each piece's weighted linear total and its target, then its 1.7-norm value and its shape target:
//
//	Head       66  T1        72.0  72.4   46.8  46.5  str 26 sta 16 hit 30
//	Neck       71  T3        42.0  42.0   27.4  28.0  str 15 sta 9 crit 18
//	Shoulder   66  T2        55.0  54.8   35.9  34.9  str 19 sta 12 crit 24
//	Back       71  T3        42.0  42.0   27.4  28.0  str 15 sta 9 crit 18
//	Chest      66  T1        72.0  72.4   48.1  46.5  str 22 sta 14 agi 36
//	Wrist      71  T3        42.0  42.0   27.1  28.0  str 16 sta 10 ap 32
//	Hands      66  T2        55.0  54.8   36.5  34.9  str 17 sta 11 agi 27
//	Waist      71  T2        59.0  58.9   38.7  37.4  str 21 sta 12 hit 26
//	Legs       66  T1        72.0  72.4   48.1  46.5  str 22 sta 14 agi 36
//	Feet       66  T2        55.0  54.8   35.9  34.9  str 19 sta 12 hit 24
//	Finger     71  T3        42.0  42.0   27.1  28.0  str 16 sta 10 ap 32
//	Finger II  71  T3        42.0  42.0   27.1  28.0  str 16 sta 10 ap 32
//	Ranged     71  ranged    19.5  19.6      -     -  agi 9 ap 1 hit 10
//	Main Hand  71  onehand   32.0  31.9      -     -  str 10 agi 4 sta 3 crit 15
//	Off Hand   71  onehand   32.0  31.9      -     -  str 10 agi 4 sta 3 crit 15
//	Two-Hand   76  twohand   79.0  78.8      -     -  str 30 agi 18 sta 10 ap 16 crit 13
//
// Totals: armor 650.5 (target 650.5), bow 19.5, swords 64.0, two-hander 79.0. The Fury set is
// 733.5, the Arms / Fury 2H set 748.5.
//
// The stat mix is the fair-mix standard (orchestration/standards.md 5a, gear_targets.md at
// da0a4c79e): the shares of Season of Discovery level 60 BiS warrior gear, Stamina at the spec's own
// SoD share. Hit is set at the cap (90 rating, 9% for special attacks against a level 63 boss,
// and no warrior talent adds main hand hit) and the other stats keep the table's proportions.
// Linear shares, enchants and trinkets left out:
//
//	Set                     Sta   Str   Agi   AP    Crit  Hit
//	Fury target             19.6  36.0  17.0  6.8   13.2  7.4 (SoD; the cap here)
//	Fury (these items)      19.8  33.3  15.8  6.6   12.3  12.3
//	Arms / Fury 2H target   19.7  36.6  18.0  7.7   10.4  7.7 (SoD; the cap here)
//	Arms / Fury 2H (items)  19.9  33.9  16.8  7.5   9.8   12.0
//
// Every armor piece carries Strength and Stamina and one more stat: hit on the head, waist and feet,
// Agility on the chest, legs and hands, crit on the shoulders, neck and back, attack power on the
// wrist and rings. Arms and Fury share the armor, rings and bow, and the swords and the two-hander
// carry the difference between the two targets. The plate has Battlegear of Might's armor at item
// level 66, the waist Onslaught Girdle's.
//
// Arms and Fury share the armor, the rings, the bow and the trinkets, because their Era Phase 1 BiS
// lists are the same outside the weapons. Fury dual wields a main hand and an off hand in place of
// Vis'kag the Bloodletter and Brutality Blade, Arms and two-hand Fury a two-hander each in place of
// Bonereaver's Edge.
//
// Each weapon keeps the damage per second of the Era weapon it replaces. Its type and speed are the
// ones the spec sims best with (standards 22), taken from a real Era Phase 1 weapon: Deathbringer's
// 2.9 sec axe and Lord Blackwood's Blade's 1.5 sec sword for Fury, The Blackrock Slicer's 4.0 sec axe
// for Arms and Dreadforge Retaliator's 3.7 sec axe for two-hand Fury. Under Forever's normalised Rage
// a slow main hand gives the same Rage and bigger Heroic Strikes and Mortal Strikes, and a fast off
// hand rolls Unbridled Wrath more often while Whirlwind's off hand strike stays normalised.
//
// The head and legs form a set with the bonuses of Battlegear of Glory, Forever's Molten Core tier set
// for Arms and Fury (set 2102, items 280899 to 280904, client 70338, see sim/warrior/item_sets_pve.go).
// Only two pieces carry it, so only the 2-piece bonus (1% attack speed) is live, the same for every
// class. The tier pieces use its icons.
//
// The two trinkets are synthetic too (standards 1b and 8a: Forever removed the old dungeon
// trinkets and the sim does not model its launch trinkets yet). The generator script lived in the
// session scratchpad (synth_warrior.py), the numbers above are its output.
func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, foreverSyntheticWarriorItems...)
}

var foreverSyntheticWarriorItems = []*proto.UIItem{
	// Head, item level 66, in place of Lionheart Helm (61, 12640).
	{
		Id: 991001, Name: "Warrior Synthetic Head (Phase 1)", Icon: "inv_helm_plate_raidwarriorhyjalc60_d_01",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypePlate,
		Stats:     stats.Stats{stats.Strength: 26, stats.Stamina: 16, stats.Armor: 608}.ToFloatArray(),
		HitRating: 30,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
		SetName:        "Warrior Synthetic (Phase 1)",
	},
	// Neck, item level 71, in place of Onyxia Tooth Pendant (74, 18404).
	{
		Id: 991002, Name: "Warrior Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_09",
		Type:       proto.ItemType_ItemTypeNeck,
		Stats:      stats.Stats{stats.Strength: 15, stats.Stamina: 9}.ToFloatArray(),
		CritRating: 18,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Shoulder, item level 66, in place of Truestrike Shoulders (61, 12927).
	{
		Id: 991003, Name: "Warrior Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_plate_raidwarriorhyjalc60_d_01",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypePlate,
		Stats:      stats.Stats{stats.Strength: 19, stats.Stamina: 12, stats.Armor: 562}.ToFloatArray(),
		CritRating: 24,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Back, item level 71, in place of Cape of the Black Baron (63, 13340).
	{
		Id: 991004, Name: "Warrior Synthetic Back (Phase 1)", Icon: "inv_misc_cape_20",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Strength: 15, stats.Stamina: 9, stats.Armor: 45}.ToFloatArray(),
		CritRating: 18,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Chest, item level 66, in place of Savage Gladiator Chain (57, 11726).
	{
		Id: 991005, Name: "Warrior Synthetic Chest (Phase 1)", Icon: "inv_chest_plate_raidwarriorhyjalc60_d_01",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypePlate,
		Stats: stats.Stats{stats.Strength: 22, stats.Stamina: 14, stats.Agility: 36, stats.Armor: 749}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Wrist, item level 71, in place of Wristguards of Stability (65, 19146).
	{
		Id: 991006, Name: "Warrior Synthetic Wrist (Phase 1)", Icon: "inv_bracer_19",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypePlate,
		Stats: stats.Stats{stats.Strength: 16, stats.Stamina: 10, stats.AttackPower: 32, stats.Armor: 328}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Hands, item level 66, in place of Edgemaster's Handguards (49, 14551).
	{
		Id: 991007, Name: "Warrior Synthetic Hands (Phase 1)", Icon: "inv_glove_plate_raidwarriorhyjalc60_d_01",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypePlate,
		Stats: stats.Stats{stats.Strength: 17, stats.Stamina: 11, stats.Agility: 27, stats.Armor: 468}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Waist, item level 71, in place of Onslaught Girdle (78, 19137).
	{
		Id: 991008, Name: "Warrior Synthetic Waist (Phase 1)", Icon: "inv_belt_09",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypePlate,
		Stats:     stats.Stats{stats.Strength: 21, stats.Stamina: 12, stats.Armor: 494}.ToFloatArray(),
		HitRating: 26,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Legs, item level 66, in place of Cloudkeeper Legplates (62, 14554).
	{
		Id: 991009, Name: "Warrior Synthetic Legs (Phase 1)", Icon: "inv_pant_plate_raidwarriorhyjalc60_d_01",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypePlate,
		Stats: stats.Stats{stats.Strength: 22, stats.Stamina: 14, stats.Agility: 36, stats.Armor: 655}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
		SetName:        "Warrior Synthetic (Phase 1)",
	},
	// Feet, item level 66, in place of Bloodmail Boots (61, 14616).
	{
		Id: 991010, Name: "Warrior Synthetic Feet (Phase 1)", Icon: "inv_boot_plate_raidwarriorhyjalc60_d_01",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypePlate,
		Stats:     stats.Stats{stats.Strength: 19, stats.Stamina: 12, stats.Armor: 515}.ToFloatArray(),
		HitRating: 24,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Finger, item level 71, in place of Band of Accuria (78, 17063).
	{
		Id: 991011, Name: "Warrior Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_15",
		Type:  proto.ItemType_ItemTypeFinger,
		Stats: stats.Stats{stats.Strength: 16, stats.Stamina: 10, stats.AttackPower: 32}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Finger II, item level 71, in place of Quick Strike Ring (67, 18821).
	{
		Id: 991012, Name: "Warrior Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_16",
		Type:  proto.ItemType_ItemTypeFinger,
		Stats: stats.Stats{stats.Strength: 16, stats.Stamina: 10, stats.AttackPower: 32}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Ranged, item level 71, in place of Striker's Mark (69, 17069).
	{
		Id: 991013, Name: "Warrior Synthetic Ranged (Phase 1)", Icon: "inv_weapon_bow_08",
		Type:             proto.ItemType_ItemTypeRanged,
		RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeBow,
		// Striker's Mark damage and speed.
		WeaponDamageMin: 69, WeaponDamageMax: 129, WeaponSpeed: 2.5,
		Stats:     stats.Stats{stats.Agility: 9, stats.AttackPower: 1}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Main Hand, item level 71, in place of Vis'kag the Bloodletter (74, 17075).
	{
		Id: 991014, Name: "Fury Synthetic Main Hand (Phase 1)", Icon: "inv_axe_09",
		Type:       proto.ItemType_ItemTypeWeapon,
		WeaponType: proto.WeaponType_WeaponTypeAxe, HandType: proto.HandType_HandTypeOneHand,
		// Deathbringer's speed and damage spread at Vis'kag the Bloodletter's 55.2 DPS
		// (standards 22: the type and speed that sim best).
		WeaponDamageMin: 111.6, WeaponDamageMax: 208.5, WeaponSpeed: 2.9,
		Stats:      stats.Stats{stats.Strength: 10, stats.Agility: 4, stats.Stamina: 3}.ToFloatArray(),
		CritRating: 15,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Off Hand, item level 71, in place of Brutality Blade (70, 18832).
	{
		Id: 991015, Name: "Fury Synthetic Off Hand (Phase 1)", Icon: "inv_sword_05",
		Type:       proto.ItemType_ItemTypeWeapon,
		WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand,
		// Lord Blackwood's Blade's speed and damage spread at Brutality Blade's 51.6 DPS
		// (standards 22: the type and speed that sim best).
		WeaponDamageMin: 53.3, WeaponDamageMax: 101.5, WeaponSpeed: 1.5,
		Stats:      stats.Stats{stats.Strength: 10, stats.Agility: 4, stats.Stamina: 3}.ToFloatArray(),
		CritRating: 15,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Two-Hand, item level 76, in place of Bonereaver's Edge (77, 17076).
	{
		Id: 991016, Name: "Arms Synthetic Two-Hand (Phase 1)", Icon: "inv_weapon_halberd_11",
		Type:       proto.ItemType_ItemTypeWeapon,
		WeaponType: proto.WeaponType_WeaponTypeAxe, HandType: proto.HandType_HandTypeTwoHand,
		// The Blackrock Slicer's speed and damage spread at Bonereaver's Edge's 75.9 DPS
		// (standards 22: the type and speed that sim best).
		WeaponDamageMin: 242.5, WeaponDamageMax: 364.5, WeaponSpeed: 4,
		Stats:      stats.Stats{stats.Strength: 30, stats.Agility: 18, stats.Stamina: 10, stats.AttackPower: 16}.ToFloatArray(),
		CritRating: 13,
		Ilvl:       76, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Two-Hand for two-hand Fury, item level 76: the Arms two-hander's stats at the speed two-hand Fury
	// sims best with.
	{
		Id: 991019, Name: "Fury Synthetic Two-Hand (Phase 1)", Icon: "inv_axe_22",
		Type:       proto.ItemType_ItemTypeWeapon,
		WeaponType: proto.WeaponType_WeaponTypeAxe, HandType: proto.HandType_HandTypeTwoHand,
		// Dreadforge Retaliator's speed and damage spread at Bonereaver's Edge's 75.9 DPS
		// (standards 22: the type and speed that sim best).
		WeaponDamageMin: 223.7, WeaponDamageMax: 337.8, WeaponSpeed: 3.7,
		Stats:      stats.Stats{stats.Strength: 30, stats.Agility: 18, stats.Stamina: 10, stats.AttackPower: 16}.ToFloatArray(),
		CritRating: 13,
		Ilvl:       76, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// Stat trinket, item level 71 (standards 1b: 0.56 x L(71) = 45.9 linear, 1.7-norm 35.7).
	{
		Id: 991017, Name: "Warrior Synthetic Trinket (Phase 1)", Icon: "inv_misc_armorkit_09",
		Type:       proto.ItemType_ItemTypeTrinket,
		Stats:      stats.Stats{stats.Strength: 30}.ToFloatArray(),
		CritRating: 16,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	// On-use trinket, item level 71, no passive stats: +335 attack power for 20 sec every 2 min
	// (sim/warrior/forever_synthetic_trinkets.go), on average 335 x 0.5 x 40 / 180 = 37.2 weighted
	// points over a 180 sec fight, Earthstrike's size.
	{
		Id: 991018, Name: "Warrior Synthetic Battle Trinket (Phase 1)", Icon: "inv_misc_head_dragon_bronze",
		Type: proto.ItemType_ItemTypeTrinket,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
}
