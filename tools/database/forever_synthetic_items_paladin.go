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
// Might, Band of Accuria, Quick Strike Ring and Spinal Reaper. Hand of Justice, Blackhand's Breadth and
// Libram of Fervor stay real items.
//
// Budget. Every class's synthetic set uses one item level table (the orchestrator's fairness
// ruling): head, shoulders, chest, hands, legs and feet at 66 (Molten Core tier), neck, back,
// wrist, waist and both rings at 71, and two-handers at 76. The budget standard, backed out of the
// enhancement items:
//
//	value = 1.2 * ilvl - 36.4,   budget = value * slot,   (sum over stats of (price * amount)^1.5)^(2/3) = budget
//
// with slot 1 for head, chest, legs and two-handers, 0.77 for shoulders, hands, waist and feet, and
// 0.5625 for neck, back, wrist and fingers. The prices per point are 1 for the five attributes, 0.55
// for spell power (Forever's tuned price, see forever_synthetic_items.go), 1 for hit rating (10 make
// 1%) and 1 for crit rating (14 make 1%). Armor is free. Per slot, budget and what the stats come to:
//
//	head     66  42.80  42.77      waist    71  37.58  37.63
//	neck     71  27.45  27.58      legs     66  42.80  42.81
//	shoulder 66  32.96  32.99      feet     66  32.96  32.99
//	back     71  27.45  27.41      finger   71  27.45  27.38
//	chest    66  42.80  42.84      finger2  71  27.45  27.48
//	wrist    71  27.45  27.43      two-hand 76  54.80  54.76
//	hands    66  32.96  33.04
//
// The head, for example: (1.2 * 66 - 36.4) * 1 = 42.80, and (12^1.5 + 11^1.5 + 9^1.5 + 9^1.5 +
// (29 * 0.55)^1.5 + 10^1.5 + 14^1.5)^(2/3) = 42.77.
//
// The stat mix is a hybrid one like the enhancement shaman's, since the paladin's seals, judgements,
// Consecration and Holy Strike all take spell power under Forever:
//   - Strength is about 19% of the set's budget, spell power 20%, Stamina 16%, Intellect 11%,
//     Agility 5%.
//   - Spirit is about 10% (79 points on 8 of the 13 pieces), because the Forever devs want Spirit to
//     matter on mana users.
//   - 90 hit rating on 9 pieces, 9%, the melee cap against a level 63 boss from gear alone, which
//     the Shock and sub Holy builds need since they have no Precision. The Level 60 build's
//     Precision 3 then only counts for its spells (the judgements and Consecration miss 16% at
//     the most). Hit is unified under Forever.
//   - 70 crit rating on 5 pieces (5%), a little under the Era set's.
//   - Not every piece has every stat: the neck, wrist and first ring have no Intellect, the shoulders,
//     hands and rings no Spirit, and the back, waist, feet and weapon carry the Agility.
//
// Armor is the Battlegear of Might piece's of the slot (plate, item level 66), scaled by item level
// for the wrist and waist (the cloak's from the enhancement cloak). The two-hander has Spinal Reaper's damage and speed. It is an axe like
// Spinal Reaper, so no paladin race gets a weapon racial from it.
//
// The head and chest form a set with the bonuses of Justice Battlegear, Forever's Tier 1
// retribution set (see sim/paladin/item_sets_pve.go). Every class's preset carries its tier set's
// 2-piece bonus only (the orchestrator's fairness ruling), so only those two pieces are in it. The generator
// script lived in the session scratchpad (synth_pal.py), the numbers below are its output.
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
			Stats:      stats.Stats{stats.Strength: 7, stats.Stamina: 11, stats.Intellect: 16, stats.SpellPower: 30, stats.Armor: 608}.ToFloatArray(),
			CritRating: 23,
			Ilvl:       66, SetName: set,
		}),
		// Neck, item level 71, in place of Onyxia Tooth Pendant.
		retribution(&proto.UIItem{
			Id: 992002, Name: "Retribution Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_09",
			Type:      proto.ItemType_ItemTypeNeck,
			Stats:     stats.Stats{stats.Strength: 7, stats.Stamina: 12, stats.SpellPower: 22}.ToFloatArray(),
			HitRating: 10,
			Ilvl:      71,
		}),
		// Shoulder, item level 66, in place of Truestrike Shoulders.
		retribution(&proto.UIItem{
			Id: 992003, Name: "Retribution Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_10",
			Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats:     stats.Stats{stats.Strength: 9, stats.Stamina: 7, stats.Intellect: 16, stats.SpellPower: 24, stats.Armor: 562}.ToFloatArray(),
			HitRating: 10,
			Ilvl:      66,
		}),
		// Back, item level 71, in place of Cape of the Black Baron.
		retribution(&proto.UIItem{
			Id: 992004, Name: "Retribution Synthetic Back (Phase 1)", Icon: "inv_misc_cape_20",
			Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
			Stats:     stats.Stats{stats.Agility: 6, stats.Stamina: 19, stats.Spirit: 5, stats.SpellPower: 10, stats.Armor: 58}.ToFloatArray(),
			HitRating: 10,
			Ilvl:      71,
		}),
		// Chest, item level 66, in place of the Savage Gladiator Chain.
		retribution(&proto.UIItem{
			Id: 992005, Name: "Retribution Synthetic Chest (Phase 1)", Icon: "inv_chest_plate16",
			Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats: stats.Stats{stats.Strength: 12, stats.Stamina: 20, stats.Intellect: 19, stats.Spirit: 6, stats.SpellPower: 30, stats.Armor: 749}.ToFloatArray(),
			Ilvl:  66, SetName: set,
		}),
		// Wrist, item level 71, in place of Wristguards of True Flight.
		retribution(&proto.UIItem{
			Id: 992006, Name: "Retribution Synthetic Wrist (Phase 1)", Icon: "inv_bracer_14",
			Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats:     stats.Stats{stats.Strength: 11, stats.Stamina: 8, stats.SpellPower: 22, stats.Armor: 353}.ToFloatArray(),
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
			Stats:      stats.Stats{stats.Strength: 16, stats.Agility: 18, stats.Stamina: 16, stats.Armor: 453}.ToFloatArray(),
			CritRating: 18,
			Ilvl:       71,
		}),
		// Legs, item level 66, in place of Legplates of Might.
		retribution(&proto.UIItem{
			Id: 992009, Name: "Retribution Synthetic Legs (Phase 1)", Icon: "inv_pants_04",
			Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats:     stats.Stats{stats.Strength: 25, stats.Stamina: 24, stats.Spirit: 25, stats.Armor: 655}.ToFloatArray(),
			HitRating: 10,
			Ilvl:      66,
		}),
		// Feet, item level 66, in place of Sabatons of Might.
		retribution(&proto.UIItem{
			Id: 992010, Name: "Retribution Synthetic Feet (Phase 1)", Icon: "inv_boots_plate_06",
			Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats:     stats.Stats{stats.Stamina: 6, stats.Spirit: 18, stats.SpellPower: 18, stats.AttackPower: 28, stats.RangedAttackPower: 28, stats.Armor: 515}.ToFloatArray(),
			HitRating: 10,
			Ilvl:      66,
		}),
		// Finger, item level 71, in place of Band of Accuria.
		retribution(&proto.UIItem{
			Id: 992011, Name: "Retribution Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_15",
			Type:      proto.ItemType_ItemTypeFinger,
			Stats:     stats.Stats{stats.Strength: 9, stats.Stamina: 13, stats.SpellPower: 19}.ToFloatArray(),
			HitRating: 10,
			Ilvl:      71,
		}),
		// Second finger, item level 71, in place of Quick Strike Ring. The caster ring of the two.
		retribution(&proto.UIItem{
			Id: 992013, Name: "Retribution Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_07",
			Type:       proto.ItemType_ItemTypeFinger,
			Stats:      stats.Stats{stats.Intellect: 11, stats.Spirit: 9, stats.SpellPower: 18}.ToFloatArray(),
			CritRating: 13,
			Ilvl:       71,
		}),
		// Two-Hand, item level 76, in place of Spinal Reaper.
		retribution(&proto.UIItem{
			Id: 992012, Name: "Retribution Synthetic Two-Hand (Phase 1)", Icon: "inv_axe_09",
			Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeAxe, HandType: proto.HandType_HandTypeTwoHand,
			Stats:     stats.Stats{stats.Strength: 11, stats.Stamina: 13, stats.Intellect: 12, stats.Spirit: 21, stats.SpellPower: 27}.ToFloatArray(),
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
