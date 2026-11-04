package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Synthetic Phase 1 items for a Protection paladin tank under Forever, ids 990700-990799 (the tank
// workers' brief). See forever_synthetic_items.go for why these exist.
//
// Budget: standards 36 (the tank prices: Defense 1.5 a point, block chance 5 a 1%, block value 0.4
// a point, armor free) on standards 1b's hybrid archetype, the item level table of 2, and the
// Protection paladin row of gear_targets.md's Tanks section. Each armor piece's weighted linear
// total is slot mod x L(ilvl) x f within 2%, L(ilvl) = 1.125 ilvl + 2.0, with the hybrid f 1.093 /
// 1.103 / 1.048, and its 1.7-norm within 5% of slot mod x (0.661 ilvl + 2.92), which holds a hybrid
// to 4 or 5 stat lines. The one-hander is 0.45 x L(71) = 36.8, the shield 0.72 x L(71) = 58.9 (its
// block value and Defense included, its base armor free). Linear total and target, 1.7-norm and
// target:
//
//	head     66   83.0 /  83.3   47.5 / 46.5
//	neck     71   48.5 /  48.3   28.1 / 28.0
//	shoulder 66   62.8 /  63.1   35.6 / 34.9
//	back     71   48.0 /  48.3   27.6 / 28.0
//	chest    66   83.0 /  83.3   47.5 / 46.5
//	wrist    71   48.0 /  48.3   27.5 / 28.0
//	hands    66   63.5 /  63.1   33.2 / 34.9
//	waist    71   67.1 /  67.7   35.9 / 37.4
//	legs     66   83.0 /  83.3   47.0 / 46.5
//	feet     66   62.5 /  63.1   34.9 / 34.9
//	finger   71   48.0 /  48.3   28.7 / 28.0
//	finger2  71   48.5 /  48.3   27.8 / 28.0
//	main hand 71  37.0 /  36.8
//	shield   71   58.9 /  58.9
//
// Set 841.8 of 844.1. Shares (gear_targets.md's targets in brackets): Stamina 30.2% (32.4), Strength
// 16.4 (16.7), Agility 3.6 (3.5), Spirit 9.9 (10.0), spell power 1.1 (1.1), hit 5.9 (3.2), crit 3.1
// (3.1), Defense 19.4 (19.5; 109 Defense), block 3.0 (3.2; 5%), block value 7.2 (7.2; 152 with the
// shield's own). No resistances, dodge or parry (standards 36).
//
// Hit is sized to the 8% cap counting Precision 3/3 (standards 4, the tank workers' restart brief of
// 2026-11-04): 50 hit rating is 5%, 8% with Precision. The row's 3.2% (27 rating, 5.7% with
// Precision) was raised by 23 rating on the two rings and the main hand, out of their own Stamina
// and Strength, so each keeps its budget. 2.3% hit against the same 23 points spread over the set's
// other stats in its proportions: +29 at Balanced on the six scoring encounters (3,000 iterations,
// client 70753).
//
// Armor is the Lawbringer / Battlegear of Might piece's of the slot (plate, item level 66), the
// wrist's, waist's and cloak's as the retribution set has them, and Drillborer Disk's for the shield.
// The main hand keeps Brutality Blade's 51.6 damage a second (Era Phase 1's tank one-hander) at the
// speed the tank objective sims best with among Phase 1's (standards 22): Mirah's Song's 1.8 sec
// sword beats a 2.4 sec mace by 27 / 52 / 18 (Balanced, Max survival, Max threat) and a 2.9 sec axe
// by 5 / 24 / -1, more swings carrying more Seal of Fury hits and absorbs.
//
// The head and chest form a set with Justice Battleplate's 2-piece bonus (+7 Defense), Forever's
// Tier 1 protection set (sim/paladin/item_sets_pve.go). The generator is palgear.py in the session
// scratchpad; the numbers are its output.
//
// Trinkets (standards 36, 1b, 8a): the tank stat trinket of 24 Defense and 25 block value (46.0
// linear, 1.7-norm 38.6), in the row's tank-stat proportions with 2 lines, and an on-use trinket of
// 165 Strength for 20 sec every 2 min (sim/paladin/forever_synthetic_trinkets.go). It beats 1,665
// bonus armor for 20 sec every 2 min, the same value at the tank price, at every preset weight (+88
// Balanced, +32 Max survival, +107 Max threat, 3,000 iterations on the six scoring encounters).
func init() {
	tank := func(item *proto.UIItem) *proto.UIItem {
		item.RequiredLevel = 60
		item.Phase = 1
		item.Quality = proto.ItemQuality_ItemQualityEpic
		item.Expansion = proto.Expansion_ExpansionVanilla
		item.ClassAllowlist = []proto.Class{proto.Class_ClassPaladin}
		return item
	}
	const set = "Protection Paladin Synthetic (Phase 1)"
	plate := proto.ArmorType_ArmorTypePlate
	ForeverSyntheticItems = append(ForeverSyntheticItems,
		tank(&proto.UIItem{
			Id: 990701, Name: "Protection Paladin Synthetic Head (Phase 1)", Icon: "inv_helmet_05",
			Type: proto.ItemType_ItemTypeHead, ArmorType: plate,
			Stats: stats.Stats{stats.Stamina: 27, stats.Strength: 17, stats.Defense: 12, stats.Spirit: 21, stats.Armor: 608}.ToFloatArray(),
			Ilvl:  66, SetName: set,
		}),
		tank(&proto.UIItem{
			Id: 990702, Name: "Protection Paladin Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_10",
			Type:  proto.ItemType_ItemTypeNeck,
			Stats: stats.Stats{stats.Stamina: 17, stats.Strength: 8, stats.Defense: 9, stats.Block: 2}.ToFloatArray(),
			Ilvl:  71,
		}),
		tank(&proto.UIItem{
			Id: 990703, Name: "Protection Paladin Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_20",
			Type: proto.ItemType_ItemTypeShoulder, ArmorType: plate,
			Stats: stats.Stats{stats.Stamina: 18, stats.Strength: 15, stats.Defense: 10, stats.BlockValue: 37, stats.Armor: 562}.ToFloatArray(),
			Ilvl:  66,
		}),
		tank(&proto.UIItem{
			Id: 990704, Name: "Protection Paladin Synthetic Back (Phase 1)", Icon: "inv_misc_cape_14",
			Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
			Stats: stats.Stats{stats.Stamina: 16, stats.Agility: 12, stats.Defense: 8, stats.Spirit: 8, stats.Armor: 58}.ToFloatArray(),
			Ilvl:  71,
		}),
		tank(&proto.UIItem{
			Id: 990705, Name: "Protection Paladin Synthetic Chest (Phase 1)", Icon: "inv_chest_plate03",
			Type: proto.ItemType_ItemTypeChest, ArmorType: plate,
			Stats: stats.Stats{stats.Stamina: 27, stats.Strength: 17, stats.Defense: 12, stats.Spirit: 21, stats.Armor: 749}.ToFloatArray(),
			Ilvl:  66, SetName: set,
		}),
		tank(&proto.UIItem{
			Id: 990706, Name: "Protection Paladin Synthetic Wrist (Phase 1)", Icon: "inv_bracer_09",
			Type: proto.ItemType_ItemTypeWrist, ArmorType: plate,
			Stats: stats.Stats{stats.Stamina: 16, stats.Strength: 10, stats.Defense: 8, stats.BlockValue: 25, stats.Armor: 353}.ToFloatArray(),
			Ilvl:  71,
		}),
		tank(&proto.UIItem{
			Id: 990707, Name: "Protection Paladin Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_29",
			Type: proto.ItemType_ItemTypeHands, ArmorType: plate,
			Stats:      stats.Stats{stats.Stamina: 17, stats.Strength: 14, stats.Agility: 10, stats.SpellPower: 11, stats.Armor: 468}.ToFloatArray(),
			CritRating: 13,
			Ilvl:       66,
		}),
		tank(&proto.UIItem{
			Id: 990708, Name: "Protection Paladin Synthetic Waist (Phase 1)", Icon: "inv_belt_27",
			Type: proto.ItemType_ItemTypeWaist, ArmorType: plate,
			Stats:      stats.Stats{stats.Stamina: 19, stats.Strength: 16, stats.Defense: 9, stats.BlockValue: 14, stats.Armor: 453}.ToFloatArray(),
			CritRating: 13,
			Ilvl:       71,
		}),
		tank(&proto.UIItem{
			Id: 990709, Name: "Protection Paladin Synthetic Legs (Phase 1)", Icon: "inv_pants_04",
			Type: proto.ItemType_ItemTypeLegs, ArmorType: plate,
			Stats: stats.Stats{stats.Stamina: 24, stats.Strength: 18, stats.Defense: 12, stats.Spirit: 23, stats.Armor: 655}.ToFloatArray(),
			Ilvl:  66,
		}),
		tank(&proto.UIItem{
			Id: 990710, Name: "Protection Paladin Synthetic Feet (Phase 1)", Icon: "inv_boots_plate_03",
			Type: proto.ItemType_ItemTypeFeet, ArmorType: plate,
			Stats: stats.Stats{stats.Stamina: 24, stats.Strength: 6, stats.Agility: 8, stats.Defense: 9, stats.Spirit: 11, stats.Armor: 515}.ToFloatArray(),
			Ilvl:  66,
		}),
		tank(&proto.UIItem{
			Id: 990711, Name: "Protection Paladin Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_14",
			Type:      proto.ItemType_ItemTypeFinger,
			Stats:     stats.Stats{stats.Stamina: 16, stats.Strength: 10, stats.Block: 1}.ToFloatArray(),
			HitRating: 17,
			Ilvl:      71,
		}),
		tank(&proto.UIItem{
			Id: 990712, Name: "Protection Paladin Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_24",
			Type:      proto.ItemType_ItemTypeFinger,
			Stats:     stats.Stats{stats.Stamina: 15, stats.Strength: 2, stats.Defense: 7, stats.Block: 1}.ToFloatArray(),
			HitRating: 16,
			Ilvl:      71,
		}),
		// Main hand, item level 71: Brutality Blade's 51.6 damage a second at Mirah's Song's 1.8 sec
		// (a one-hand sword, its 0.7 to 1.3 spread).
		tank(&proto.UIItem{
			Id: 990713, Name: "Protection Paladin Synthetic Main Hand (Phase 1)", Icon: "inv_sword_36",
			Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand,
			Stats:           stats.Stats{stats.Strength: 6, stats.Stamina: 5, stats.Defense: 6}.ToFloatArray(),
			HitRating:       17,
			WeaponDamageMin: 65, WeaponDamageMax: 121, WeaponSpeed: 1.8,
			Ilvl: 71,
		}),
		// Shield, item level 71, Drillborer Disk's armor.
		tank(&proto.UIItem{
			Id: 990714, Name: "Protection Paladin Synthetic Shield (Phase 1)", Icon: "inv_shield_10",
			Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeShield, HandType: proto.HandType_HandTypeOffHand,
			Stats: stats.Stats{stats.Stamina: 13, stats.Defense: 7, stats.BlockValue: 76, stats.Block: 1, stats.Armor: 2539}.ToFloatArray(),
			Ilvl:  71,
		}),
		tank(&proto.UIItem{
			Id: 990715, Name: "Protection Paladin Synthetic Trinket (Phase 1)", Icon: "inv_jewelry_talisman_02",
			Type:  proto.ItemType_ItemTypeTrinket,
			Stats: stats.Stats{stats.Defense: 24, stats.BlockValue: 25}.ToFloatArray(),
			Ilvl:  71,
		}),
		tank(&proto.UIItem{
			Id: 990716, Name: "Protection Paladin Synthetic Strength On-Use Trinket (Phase 1)", Icon: "inv_misc_rune_08",
			Type: proto.ItemType_ItemTypeTrinket,
			Ilvl: 71,
		}),
	)
}
