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
// Budget. Each piece gets the full epic budget of its slot at the Phase 1 item level: 66 (Molten Core
// tier), or the Era item's level where that is higher (Onyxia Tooth Pendant 74, Wristguards of True
// Flight 71, Onslaught Girdle and Band of Accuria 78, Quick Strike Ring 67, Spinal Reaper 76). This is
// the budget standard all the classes' synthetic sets use, backed out of the enhancement items:
//
//	value = 1.2 * ilvl - 36.4,   budget = value * slot,   (sum over stats of (price * amount)^1.5)^(2/3) = budget
//
// with slot 1 for head, chest, legs and two-handers, 0.77 for shoulders, hands, waist and feet, and
// 0.5625 for neck, back, wrist and fingers. The prices per point are 1 for the five attributes, 0.55
// for spell power (Forever's tuned price, see forever_synthetic_items.go), 1 for hit rating (10 make
// 1%) and 1 for crit rating (14 make 1%). Armor is free. Per slot, budget and what the stats come to:
//
//	head     66  42.80  42.77      waist    78  44.04  44.13
//	neck     74  29.47  29.55      legs     66  42.80  42.81
//	shoulder 66  32.96  32.99      feet     66  32.96  32.99
//	back     66  24.08  24.07      finger   78  32.17  32.05
//	chest    66  42.80  42.84      finger2  67  24.75  24.75
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
//   - 90 hit rating on 9 pieces, 9% hit, about what the Era Phase 1 retribution set carries (Lionheart
//     Helm, Truestrike Shoulders and Onyxia Tooth Pendant on Precision 3/3). Hit is unified under
//     Forever, so it covers the judgements and Consecration too.
//   - 70 crit rating on 5 pieces (5%), a little under the Era set's.
//   - Not every piece has every stat: the neck, wrist and first ring have no Intellect, the shoulders,
//     hands and rings no Spirit, and the back, waist, feet and weapon carry the Agility.
//
// Armor is the Battlegear of Might piece's of the slot (plate, item level 66), scaled by item level
// for the wrist and waist. The two-hander has Spinal Reaper's damage and speed. It is an axe like
// Spinal Reaper, so no paladin race gets a weapon racial from it.
//
// The head, shoulders, chest, hands, legs and feet form a set with the bonuses of Justice
// Battlegear, Forever's Tier 1 retribution set (see sim/paladin/item_sets_pve.go). The generator
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
			Stats:     stats.Stats{stats.Strength: 12, stats.Stamina: 11, stats.Intellect: 9, stats.Spirit: 9, stats.SpellPower: 29, stats.Armor: 608}.ToFloatArray(),
			HitRating: 10, CritRating: 14,
			Ilvl: 66, SetName: set,
		}),
		// Neck, item level 74, in place of Onyxia Tooth Pendant.
		retribution(&proto.UIItem{
			Id: 992002, Name: "Retribution Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_09",
			Type:      proto.ItemType_ItemTypeNeck,
			Stats:     stats.Stats{stats.Strength: 12, stats.Stamina: 9, stats.SpellPower: 28}.ToFloatArray(),
			HitRating: 10,
			Ilvl:      74,
		}),
		// Shoulder, item level 66, in place of Truestrike Shoulders.
		retribution(&proto.UIItem{
			Id: 992003, Name: "Retribution Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_10",
			Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats:      stats.Stats{stats.Strength: 11, stats.Stamina: 9, stats.Intellect: 8, stats.SpellPower: 25, stats.Armor: 562}.ToFloatArray(),
			CritRating: 14,
			Ilvl:       66, SetName: set,
		}),
		// Back, item level 66, in place of Cape of the Black Baron.
		retribution(&proto.UIItem{
			Id: 992004, Name: "Retribution Synthetic Back (Phase 1)", Icon: "inv_misc_cape_20",
			Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
			Stats: stats.Stats{stats.Agility: 6, stats.Stamina: 8, stats.Intellect: 8, stats.Spirit: 9, stats.SpellPower: 18, stats.Armor: 54}.ToFloatArray(),
			Ilvl:  66,
		}),
		// Chest, item level 66, in place of the Savage Gladiator Chain.
		retribution(&proto.UIItem{
			Id: 992005, Name: "Retribution Synthetic Chest (Phase 1)", Icon: "inv_chest_plate16",
			Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats:     stats.Stats{stats.Strength: 15, stats.Stamina: 13, stats.Intellect: 13, stats.Spirit: 11, stats.SpellPower: 28, stats.Armor: 749}.ToFloatArray(),
			HitRating: 10,
			Ilvl:      66, SetName: set,
		}),
		// Wrist, item level 71, in place of Wristguards of True Flight.
		retribution(&proto.UIItem{
			Id: 992006, Name: "Retribution Synthetic Wrist (Phase 1)", Icon: "inv_bracer_14",
			Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats:     stats.Stats{stats.Strength: 14, stats.Stamina: 6, stats.SpellPower: 23, stats.Armor: 353}.ToFloatArray(),
			HitRating: 10,
			Ilvl:      71,
		}),
		// Hands, item level 66, in place of Flameguard Gauntlets.
		retribution(&proto.UIItem{
			Id: 992007, Name: "Retribution Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_10",
			Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats:     stats.Stats{stats.Strength: 12, stats.Stamina: 7, stats.Intellect: 6, stats.SpellPower: 18, stats.Armor: 468}.ToFloatArray(),
			HitRating: 10, CritRating: 14,
			Ilvl: 66, SetName: set,
		}),
		// Waist, item level 78, in place of Onslaught Girdle.
		retribution(&proto.UIItem{
			Id: 992008, Name: "Retribution Synthetic Waist (Phase 1)", Icon: "inv_belt_29",
			Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats:     stats.Stats{stats.Strength: 15, stats.Agility: 11, stats.Stamina: 12, stats.Spirit: 11, stats.SpellPower: 20, stats.Armor: 498}.ToFloatArray(),
			HitRating: 10, CritRating: 14,
			Ilvl: 78,
		}),
		// Legs, item level 66, in place of Legplates of Might.
		retribution(&proto.UIItem{
			Id: 992009, Name: "Retribution Synthetic Legs (Phase 1)", Icon: "inv_pants_04",
			Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats:     stats.Stats{stats.Strength: 18, stats.Stamina: 13, stats.Intellect: 12, stats.Spirit: 12, stats.SpellPower: 22, stats.Armor: 655}.ToFloatArray(),
			HitRating: 10,
			Ilvl:      66, SetName: set,
		}),
		// Feet, item level 66, in place of Sabatons of Might.
		retribution(&proto.UIItem{
			Id: 992010, Name: "Retribution Synthetic Feet (Phase 1)", Icon: "inv_boots_plate_06",
			Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats: stats.Stats{stats.Strength: 10, stats.Agility: 10, stats.Stamina: 11, stats.Intellect: 9, stats.Spirit: 10, stats.SpellPower: 18, stats.Armor: 515}.ToFloatArray(),
			Ilvl:  66, SetName: set,
		}),
		// Finger, item level 78, in place of Band of Accuria.
		retribution(&proto.UIItem{
			Id: 992011, Name: "Retribution Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_15",
			Type:      proto.ItemType_ItemTypeFinger,
			Stats:     stats.Stats{stats.Strength: 13, stats.Stamina: 7, stats.SpellPower: 35}.ToFloatArray(),
			HitRating: 10,
			Ilvl:      78,
		}),
		// Second finger, item level 67, in place of Quick Strike Ring. The caster ring of the two.
		retribution(&proto.UIItem{
			Id: 992013, Name: "Retribution Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_07",
			Type:       proto.ItemType_ItemTypeFinger,
			Stats:      stats.Stats{stats.Stamina: 6, stats.Intellect: 8, stats.Spirit: 7, stats.SpellPower: 11}.ToFloatArray(),
			CritRating: 14,
			Ilvl:       67,
		}),
		// Two-Hand, item level 76, in place of Spinal Reaper.
		retribution(&proto.UIItem{
			Id: 992012, Name: "Retribution Synthetic Two-Hand (Phase 1)", Icon: "inv_axe_09",
			Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeAxe, HandType: proto.HandType_HandTypeTwoHand,
			Stats:     stats.Stats{stats.Strength: 24, stats.Agility: 12, stats.Stamina: 18, stats.Intellect: 14, stats.Spirit: 10, stats.SpellPower: 26}.ToFloatArray(),
			HitRating: 10,
			// Spinal Reaper damage and speed.
			WeaponDamageMin: 203, WeaponDamageMax: 305, WeaponSpeed: 3.4,
			Ilvl: 76,
		}),
	)
}
