package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Synthetic level 60 Phase 1 Protection warrior tank items (ids 990601 to 990617), made the same way as
// the DPS warrior's (forever_synthetic_items_warrior.go, read its comment first) under standards 36:
// they are not in the game.
//
// The budget is standards 1b, physical (the warrior's archetype): the same item level table and slot
// budgets as the DPS set, so armor 650.5, a one-hander 0.39 x L(71) = 31.9, the shield 0.72 x L(71) =
// 58.9 (block value and Defense included, its base armor free) and the ranged slot 0.24 x L(71) = 19.7.
// The tank prices are standards 36's: Defense 1.5 a point and block value 0.4 a point, with Strength,
// Agility and Stamina 1, attack power 0.5, hit and crit rating 1. Each armor piece is three stat lines
// (Stamina and two more) and passes 1b's 1.7-norm shape check. Weighted linear total and its target,
// then the 1.7-norm and its shape target:
//
//	Head       66  T1        72.0  72.4   46.7  46.5  sta 22 def 12 str 32
//	Neck       71  T3        42.0  42.0   27.1  28.0  sta 13 def 12 crit 11
//	Shoulder   66  T2        55.0  54.8   35.6  34.9  sta 17 def 16 crit 14
//	Back       71  T3        42.0  42.0   27.1  28.0  sta 13 def 12 agi 11
//	Chest      66  T1        72.0  72.4   46.7  46.5  sta 22 def 12 str 32
//	Wrist      71  T3        42.0  42.0   27.4  28.0  sta 13 str 19 ap 20
//	Hands      66  T2        55.0  54.8   35.4  34.9  sta 17 agi 15 hit 23
//	Waist      71  T2        58.6  58.9   37.9  37.4  sta 18 block value 64 hit 15
//	Legs       66  T1        72.0  72.4   48.1  46.5  sta 22 def 24 agi 14
//	Feet       66  T2        55.0  54.8   35.9  34.9  sta 17 hit 13 str 25
//	Finger     71  T3        42.0  42.0   26.7  28.0  sta 13 crit 14 ap 30
//	Finger II  71  T3        42.0  42.0   26.8  28.0  sta 13 hit 16 crit 13
//	Shield     71  shield    58.9  58.9      -     -  sta 21 def 13 block value 46
//	Main Hand  71  onehand   32.0  31.9      -     -  sta 9 str 9 def 6 crit 5
//	Ranged     71  ranged    19.5  19.7      -     -  sta 9 agi 5 hit 5 ap 1
//
// Totals: armor 650.1, the set 760.5 (target 761.0). The stat mix is gear_targets.md's Protection
// warrior row (Tanks, 2026-11-03: wowsims/sod's phase 4 tanky and damage sets at standards 36's
// prices), linear shares, enchants and trinkets left out:
//
//	Set              Sta   Str   Agi   AP    Hit   Crit  Def   Block value
//	Target (row)     31.4  15.5  6.0   3.4   9.4   7.7   20.9  5.8
//	These items      31.4  15.4  5.9   3.4   9.5   7.7   20.9  5.8
//
// That is 106 Defense (and 7 more from the 2-piece bonus), 110 block value with the shield's own, 72 hit
// rating (7.2%, under the 8% cap) and 239 Stamina. The sword keeps Mirah's Song's speed class: the
// Fury main hand's damage and 2.8 sec speed (standards 22's weapon damage of the Era Phase 1 BiS one-hander),
// so a slow Heroic Strike hand. The plate keeps the DPS set's armor, the shield an item level 71 Era epic
// shield's (2,680, between Jagged Obsidian Shield at 70 and Red Dragonscale Protector at 74).
//
// The head and legs carry Battleplate of Glory's bonuses, Forever's Tier 1 Protection set (set 2103,
// items 280905 to 280910, client 70753): its 2-piece bonus is +7 Defense (1300970). Only two pieces
// carry it, so only the 2-piece is live (standards 6; see sim/warrior/item_sets_pve.go).
//
// The trinkets (standards 36): the tank trinket is 1b's stat trinket of tank stats in the row's
// proportions, Defense and block value (24 Defense, 25 block value: 46.0 weighted linear, 1.7-norm
// 38.4 against 36.8); the second is a defensive on-use of 1b's average value at 0.1 a point of bonus
// armor, 1,665 armor for 20 sec every 2 min (sim/warrior/forever_synthetic_trinkets.go), which the tank
// objective picks over the DPS set's attack power on-use or not (docs/tank_warrior.md, Gear).
func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, foreverSyntheticWarriorTankItems...)
}

func warriorTankItem(id int32, name, icon string, slot proto.ItemType, armorType proto.ArmorType, ilvl int32, s stats.Stats, hit, crit float64, setName string) *proto.UIItem {
	return &proto.UIItem{
		Id: id, Name: name, Icon: icon, Type: slot, ArmorType: armorType,
		Stats: s.ToFloatArray(), HitRating: hit, CritRating: crit,
		Ilvl: ilvl, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
		SetName:        setName,
	}
}

const warriorTankSet = "Protection Warrior Synthetic (Phase 1)"

var foreverSyntheticWarriorTankItems = []*proto.UIItem{
	warriorTankItem(990601, "Protection Synthetic Head (Phase 1)", "inv_helm_plate_raidwarriorhyjalc60_d_01", proto.ItemType_ItemTypeHead, proto.ArmorType_ArmorTypePlate, 66,
		stats.Stats{stats.Stamina: 22, stats.Defense: 12, stats.Strength: 32, stats.Armor: 608}, 0, 0, warriorTankSet),
	warriorTankItem(990602, "Protection Synthetic Neck (Phase 1)", "inv_jewelry_necklace_09", proto.ItemType_ItemTypeNeck, proto.ArmorType_ArmorTypeUnknown, 71,
		stats.Stats{stats.Stamina: 13, stats.Defense: 12}, 0, 11, ""),
	warriorTankItem(990603, "Protection Synthetic Shoulder (Phase 1)", "inv_shoulder_plate_raidwarriorhyjalc60_d_01", proto.ItemType_ItemTypeShoulder, proto.ArmorType_ArmorTypePlate, 66,
		stats.Stats{stats.Stamina: 17, stats.Defense: 16, stats.Armor: 562}, 0, 14, ""),
	warriorTankItem(990604, "Protection Synthetic Back (Phase 1)", "inv_misc_cape_20", proto.ItemType_ItemTypeBack, proto.ArmorType_ArmorTypeCloth, 71,
		stats.Stats{stats.Stamina: 13, stats.Defense: 12, stats.Agility: 11, stats.Armor: 45}, 0, 0, ""),
	warriorTankItem(990605, "Protection Synthetic Chest (Phase 1)", "inv_chest_plate_raidwarriorhyjalc60_d_01", proto.ItemType_ItemTypeChest, proto.ArmorType_ArmorTypePlate, 66,
		stats.Stats{stats.Stamina: 22, stats.Defense: 12, stats.Strength: 32, stats.Armor: 749}, 0, 0, ""),
	warriorTankItem(990606, "Protection Synthetic Wrist (Phase 1)", "inv_bracer_19", proto.ItemType_ItemTypeWrist, proto.ArmorType_ArmorTypePlate, 71,
		stats.Stats{stats.Stamina: 13, stats.Strength: 19, stats.AttackPower: 20, stats.Armor: 328}, 0, 0, ""),
	warriorTankItem(990607, "Protection Synthetic Hands (Phase 1)", "inv_glove_plate_raidwarriorhyjalc60_d_01", proto.ItemType_ItemTypeHands, proto.ArmorType_ArmorTypePlate, 66,
		stats.Stats{stats.Stamina: 17, stats.Agility: 15, stats.Armor: 468}, 23, 0, ""),
	warriorTankItem(990608, "Protection Synthetic Waist (Phase 1)", "inv_belt_09", proto.ItemType_ItemTypeWaist, proto.ArmorType_ArmorTypePlate, 71,
		stats.Stats{stats.Stamina: 18, stats.BlockValue: 64, stats.Armor: 494}, 15, 0, ""),
	warriorTankItem(990609, "Protection Synthetic Legs (Phase 1)", "inv_pant_plate_raidwarriorhyjalc60_d_01", proto.ItemType_ItemTypeLegs, proto.ArmorType_ArmorTypePlate, 66,
		stats.Stats{stats.Stamina: 22, stats.Defense: 24, stats.Agility: 14, stats.Armor: 655}, 0, 0, warriorTankSet),
	warriorTankItem(990610, "Protection Synthetic Feet (Phase 1)", "inv_boot_plate_raidwarriorhyjalc60_d_01", proto.ItemType_ItemTypeFeet, proto.ArmorType_ArmorTypePlate, 66,
		stats.Stats{stats.Stamina: 17, stats.Strength: 25, stats.Armor: 515}, 13, 0, ""),
	warriorTankItem(990611, "Protection Synthetic Finger (Phase 1)", "inv_jewelry_ring_15", proto.ItemType_ItemTypeFinger, proto.ArmorType_ArmorTypeUnknown, 71,
		stats.Stats{stats.Stamina: 13, stats.AttackPower: 30}, 0, 14, ""),
	warriorTankItem(990612, "Protection Synthetic Finger II (Phase 1)", "inv_jewelry_ring_16", proto.ItemType_ItemTypeFinger, proto.ArmorType_ArmorTypeUnknown, 71,
		stats.Stats{stats.Stamina: 13}, 16, 13, ""),
	{
		Id: 990613, Name: "Protection Synthetic Ranged (Phase 1)", Icon: "inv_weapon_rifle_06",
		Type:             proto.ItemType_ItemTypeRanged,
		RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeGun,
		WeaponDamageMin:  69, WeaponDamageMax: 129, WeaponSpeed: 2.5,
		Stats:     stats.Stats{stats.Stamina: 9, stats.Agility: 5, stats.AttackPower: 1}.ToFloatArray(),
		HitRating: 5,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	{
		Id: 990614, Name: "Protection Synthetic Main Hand (Phase 1)", Icon: "inv_sword_39",
		Type:       proto.ItemType_ItemTypeWeapon,
		WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand,
		WeaponDamageMin: 107.9, WeaponDamageMax: 201.2, WeaponSpeed: 2.8,
		Stats:      stats.Stats{stats.Stamina: 9, stats.Strength: 9, stats.Defense: 6}.ToFloatArray(),
		CritRating: 5,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	{
		Id: 990615, Name: "Protection Synthetic Shield (Phase 1)", Icon: "inv_shield_10",
		Type:       proto.ItemType_ItemTypeWeapon,
		WeaponType: proto.WeaponType_WeaponTypeShield, HandType: proto.HandType_HandTypeOffHand,
		Stats: stats.Stats{stats.Stamina: 21, stats.Defense: 13, stats.BlockValue: 46, stats.Armor: 2680}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	{
		Id: 990616, Name: "Protection Synthetic Trinket (Phase 1)", Icon: "inv_misc_gem_pearl_04",
		Type:  proto.ItemType_ItemTypeTrinket,
		Stats: stats.Stats{stats.Defense: 24, stats.BlockValue: 25}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	{
		Id: 990617, Name: "Protection Synthetic Bulwark Trinket (Phase 1)", Icon: "inv_misc_stonetablet_01",
		Type: proto.ItemType_ItemTypeTrinket,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
}
