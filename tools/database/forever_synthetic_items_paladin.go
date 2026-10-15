package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Synthetic Phase 1 items for a retribution paladin under Forever, the paladin's block of
// ForeverSyntheticItems (ids 992000-992999). See forever_synthetic_items.go for why these exist.
//
// The pieces take the place of the Era Phase 1 items a retribution paladin wears: Lionheart Helm,
// Onyxia Tooth Pendant, Truestrike Shoulders, Cape of the Black Baron, the Savage Gladiator Chain,
// Wristguards of True Flight, Flameguard Gauntlets, Onslaught Girdle, Legplates of Might, Sabatons of
// Might, Band of Accuria, Quick Strike Ring and Spinal Reaper, plus two synthetic trinkets (Forever's
// launch trinkets are not simulated). Libram of Law stays a real item.
//
// Budget: the orchestrator's standards 1b (Forever budget per archetype). Every class uses one item
// level table: head, shoulders, chest, hands, legs and feet at 66, neck, back, wrist, waist and both
// rings at 71, the two-hander at 76. Each armor piece's weighted linear total is
//
//	slot mod x L(ilvl) x f,   L(ilvl) = 1.125 ilvl + 2.0,
//
// within 1%, with slot mods 1 (head, chest, legs), 0.75 (shoulders, hands, waist, feet) and 0.5625
// (neck, back, wrist, rings), and the hybrid factors f 1.093 / 1.103 / 1.048. Each piece's 1.7-norm,
// (sum (weight x amount)^1.7)^(1/1.7), is within 6% of slot mod x (0.661 ilvl + 2.92), which holds a
// hybrid to 4 or 5 stat lines. The two-hander is 1.035 x L(76) = 90.6. Weights: 1 for the five
// attributes, hit rating and crit rating, 0.86 for spell power, 0.5 for attack power. Armor is free.
// Linear total and target, 1.7-norm and target:
//
//	head     66   83.66 /  83.34   45.5 /  46.5
//	neck     71   48.64 /  48.27   29.0 /  28.0
//	shoulder 66   63.48 /  63.08   34.0 /  34.9
//	back     71   47.92 /  48.27   27.0 /  28.0
//	chest    66   83.52 /  83.34   46.0 /  46.5
//	wrist    71   48.64 /  48.27   29.0 /  28.0
//	hands    66   63.50 /  63.08   36.7 /  34.9
//	waist    71   68.00 /  67.73   38.5 /  37.4
//	legs     66   83.00 /  83.34   48.4 /  46.5
//	feet     66   63.00 /  63.08   34.0 /  34.9
//	finger   71   48.64 /  48.27   29.0 /  28.0
//	finger2  71   48.64 /  48.27   29.0 /  28.0
//	twohand  76   91.08 /  90.56
//
// Armor 750.6, weapon 91.1. Stat mix (standards 5a and 5b, gear_targets.md at the hit-adjusted
// targets): Stamina 17.6% of the set (target 17.7), Strength 12.6 (12.3), Agility 2.5 (2.4),
// Intellect 10.0 (10.0), Spirit 10.0 (10.0), attack power 1.5 (1.5), spell power 28.3 (28.4), crit
// 6.9 (7.0), hit 10.7. Ret keeps its Intellect and spell power amounts; the physical stats share
// what hit leaves. 90 hit rating (9%) is the melee cap from gear for the sub Holy and Shock builds;
// the same set with 60 hit rating simmed 9 / 34 / 11 DPS lower on twist / sub Holy / Shock.
//
// Trinkets (standards 1b, 8a): a stat trinket at 0.56 x L(71) = 45.9 linear (spell power 32 and crit
// rating 18: 45.5, 1.7-norm 34.7), and an on-use trinket of 140 spell power for 20 sec every 2 min
// (sim/paladin/forever_synthetic_trinkets.go). Both sim best for both playstyles.
//
// Armor is the Battlegear of Might piece's of the slot (plate, item level 66), scaled by item level
// for the wrist and waist (the cloak's from the enhancement cloak). The two-hander has Spinal
// Reaper's damage and speed. It is an axe like Spinal Reaper, so no paladin race gets a weapon racial
// from it.
//
// The head and chest form a set with the bonuses of Justice Battlegear, Forever's Tier 1
// retribution set (see sim/paladin/item_sets_pve.go). Every class's preset carries its tier set's
// 2-piece bonus only, so only those two pieces are in it. The generator script lived in the session
// scratchpad (gear1b.py), the numbers below are its output.
func init() {
	retribution := func(item *proto.UIItem) *proto.UIItem {
		item.RequiredLevel = 60
		item.Phase = 1
		item.Quality = proto.ItemQuality_ItemQualityEpic
		item.Expansion = proto.Expansion_ExpansionVanilla
		item.ClassAllowlist = []proto.Class{proto.Class_ClassPaladin}
		return item
	}
	const set = "Retribution Synthetic (Phase 1)"
	ForeverSyntheticItems = append(ForeverSyntheticItems,
		// Head, item level 66, in place of Lionheart Helm.
		retribution(&proto.UIItem{
			Id: 992001, Name: "Retribution Synthetic Head (Phase 1)", Icon: "inv_helmet_29",
			Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats:      stats.Stats{stats.Strength: 6, stats.Stamina: 17, stats.Intellect: 19, stats.SpellPower: 31, stats.Armor: 608}.ToFloatArray(),
			CritRating: 15,
			Ilvl:       66, SetName: set,
		}),
		// Neck, item level 71, in place of Onyxia Tooth Pendant.
		retribution(&proto.UIItem{
			Id: 992002, Name: "Retribution Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_09",
			Type:      proto.ItemType_ItemTypeNeck,
			Stats:     stats.Stats{stats.Strength: 9, stats.Stamina: 9, stats.SpellPower: 24}.ToFloatArray(),
			HitRating: 10,
			Ilvl:      71,
		}),
		// Shoulder, item level 66, in place of Truestrike Shoulders.
		retribution(&proto.UIItem{
			Id: 992003, Name: "Retribution Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_10",
			Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats:     stats.Stats{stats.Strength: 6, stats.Stamina: 14, stats.Intellect: 18, stats.SpellPower: 18, stats.Armor: 562}.ToFloatArray(),
			HitRating: 10,
			Ilvl:      66,
		}),
		// Back, item level 71, in place of Cape of the Black Baron.
		retribution(&proto.UIItem{
			Id: 992004, Name: "Retribution Synthetic Back (Phase 1)", Icon: "inv_misc_cape_20",
			Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
			Stats:     stats.Stats{stats.Agility: 5, stats.Stamina: 9, stats.Spirit: 5, stats.SpellPower: 22, stats.Armor: 58}.ToFloatArray(),
			HitRating: 10,
			Ilvl:      71,
		}),
		// Chest, item level 66, in place of the Savage Gladiator Chain.
		retribution(&proto.UIItem{
			Id: 992005, Name: "Retribution Synthetic Chest (Phase 1)", Icon: "inv_chest_plate16",
			Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats: stats.Stats{stats.Strength: 6, stats.Stamina: 14, stats.Intellect: 22, stats.Spirit: 14, stats.SpellPower: 32, stats.Armor: 749}.ToFloatArray(),
			Ilvl:  66, SetName: set,
		}),
		// Wrist, item level 71, in place of Wristguards of True Flight.
		retribution(&proto.UIItem{
			Id: 992006, Name: "Retribution Synthetic Wrist (Phase 1)", Icon: "inv_bracer_14",
			Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats:     stats.Stats{stats.Strength: 9, stats.Stamina: 9, stats.SpellPower: 24, stats.Armor: 353}.ToFloatArray(),
			HitRating: 10,
			Ilvl:      71,
		}),
		// Hands, item level 66, in place of Flameguard Gauntlets.
		retribution(&proto.UIItem{
			Id: 992007, Name: "Retribution Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_10",
			Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats:     stats.Stats{stats.Strength: 16, stats.SpellPower: 25, stats.Armor: 468}.ToFloatArray(),
			HitRating: 10, CritRating: 16,
			Ilvl: 66,
		}),
		// Waist, item level 71, in place of Onslaught Girdle.
		retribution(&proto.UIItem{
			Id: 992008, Name: "Retribution Synthetic Waist (Phase 1)", Icon: "inv_belt_29",
			Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats:      stats.Stats{stats.Strength: 15, stats.Agility: 16, stats.Stamina: 19, stats.Armor: 453}.ToFloatArray(),
			CritRating: 18,
			Ilvl:       71,
		}),
		// Legs, item level 66, in place of Legplates of Might.
		retribution(&proto.UIItem{
			Id: 992009, Name: "Retribution Synthetic Legs (Phase 1)", Icon: "inv_pants_04",
			Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats:     stats.Stats{stats.Strength: 24, stats.Stamina: 24, stats.Spirit: 25, stats.Armor: 655}.ToFloatArray(),
			HitRating: 10,
			Ilvl:      66,
		}),
		// Feet, item level 66, in place of Sabatons of Might.
		retribution(&proto.UIItem{
			Id: 992010, Name: "Retribution Synthetic Feet (Phase 1)", Icon: "inv_boots_plate_06",
			Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats:     stats.Stats{stats.Stamina: 8, stats.Spirit: 11, stats.SpellPower: 25, stats.AttackPower: 25, stats.RangedAttackPower: 25, stats.Armor: 515}.ToFloatArray(),
			HitRating: 10,
			Ilvl:      66,
		}),
		// Finger, item level 71, in place of Band of Accuria.
		retribution(&proto.UIItem{
			Id: 992011, Name: "Retribution Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_15",
			Type:      proto.ItemType_ItemTypeFinger,
			Stats:     stats.Stats{stats.Strength: 9, stats.Stamina: 9, stats.SpellPower: 24}.ToFloatArray(),
			HitRating: 10,
			Ilvl:      71,
		}),
		// Second finger, item level 71, in place of Quick Strike Ring. The caster ring of the two.
		retribution(&proto.UIItem{
			Id: 992013, Name: "Retribution Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_07",
			Type:       proto.ItemType_ItemTypeFinger,
			Stats:      stats.Stats{stats.Intellect: 9, stats.Spirit: 10, stats.SpellPower: 24}.ToFloatArray(),
			CritRating: 9,
			Ilvl:       71,
		}),
		// Two-Hand, item level 76, in place of Spinal Reaper.
		retribution(&proto.UIItem{
			Id: 992012, Name: "Retribution Synthetic Two-Hand (Phase 1)", Icon: "inv_axe_09",
			Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeAxe, HandType: proto.HandType_HandTypeTwoHand,
			Stats:     stats.Stats{stats.Strength: 6, stats.Stamina: 16, stats.Intellect: 16, stats.Spirit: 19, stats.SpellPower: 28}.ToFloatArray(),
			HitRating: 10,
			// Spinal Reaper damage and speed.
			WeaponDamageMin: 203, WeaponDamageMax: 305, WeaponSpeed: 3.4,
			Ilvl: 76,
		}),
		// Trinkets, item level 71 (standards 1b and 8a: Forever's launch trinkets are not simulated).
		// A stat trinket at 0.56 x L(71) = 45.9 linear, and an on-use trinket with no passive
		// stats whose effect is in sim/paladin/forever_synthetic_trinkets.go.
		retribution(&proto.UIItem{
			Id: 992014, Name: "Retribution Synthetic Trinket (Phase 1)", Icon: "inv_jewelry_talisman_07",
			Type:       proto.ItemType_ItemTypeTrinket,
			Stats:      stats.Stats{stats.SpellPower: 32}.ToFloatArray(),
			CritRating: 18,
			Ilvl:       71,
		}),
		retribution(&proto.UIItem{
			Id: 992015, Name: "Retribution Synthetic Spell Power On-Use Trinket (Phase 1)", Icon: "inv_misc_stonetablet_11",
			Type: proto.ItemType_ItemTypeTrinket,
			Ilvl: 71,
		}),
	)
}
