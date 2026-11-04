package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The synthetic Phase 1 Holy paladin items, ids 990300 to 990399: the healer's set (docs/healing_paladin.md).
// See forever_synthetic_items.go for what these items are and why we need them. The trinkets are the
// Retribution set's (992014 and 992015), spell power that heals as it damages, and the libram a real one.
//
// The budget is standards 1b for a pure caster, as the Restoration shaman's: weights 1 a point for Stamina,
// Intellect, Spirit and crit rating (14 is 1%), spell power 0.86, healing on top of spell power 0.45, MP5 2.4,
// armor free. Each armor piece carries slot mod * L(ilvl) * f in weighted linear stats within 1% (L = 1.125 *
// ilvl + 2.0, f 1.053 / 1.026 / 0.995), at item level 66 for the head, shoulders, chest, hands, legs and feet
// and 71 for the neck, back, wrist, waist and rings, and passes the shape check (its 1.7-norm within 6% of
// slot mod * (0.661 * ilvl + 2.92)). Paladins have no staff, so the weapons are a caster main hand at 1.82 *
// L(71) = 149.1 and a held off-hand at 0.53 * L(71) = 43.4, which carry what the armor leaves of each stat.
// The armor is the Retribution set's plate of each slot.
//
// The mix is the Holy paladin row of orchestration/gear_targets.md (Healers): the Classic pre-raid Holy
// paladin list's shares (Wowhead's WoW Forever Holy paladin guide lists no item yet), with its spell power and
// healing split as the Restoration shaman row's, Spirit 10% (standards 5) and no hit. A stat's share is its
// weight times its amount over the set's weighted total (the trinkets left out). The set
// (orchestration/tools/healers/paladin_holy_synth.py):
//
//	stat      amount   share   target
//	Stamina      134   14.9%   14.7 +- 1.5
//	Intellect    169   18.8%   18.7 +- 3
//	Spirit        90   10.0%   10
//	SP           227   21.7%   21.9 +- 3
//	Healing      538   26.9%   26.8 +- 3
//	Crit          25    2.8%    2.8 +- 3
//	MP5           19    5.1%    5.1 +- 3
//
// Per piece (target, weighted linear total, 1.7-norm value / shape target, stats). The armor comes to 708.8
// and the weapons to 192.1:
//
//	Head         80.3   80.2 (-0.1%)  47.1 /  46.5 (+1.1%)  18 Sta, 22 Int, 65 Heal, 11 Crit
//	Neck         45.8   45.7 (-0.3%)  27.5 /  28.0 (-1.9%)  11 Sta, 11 Spi, 42 Heal, 2 MP5
//	Shoulder     58.7   58.6 (-0.1%)  34.1 /  34.9 (-2.4%)  10 Sta, 14 Int, 13 Spi, 48 Heal
//	Back         45.8   45.7 (-0.4%)  27.3 /  28.0 (-2.6%)  5 Sta, 10 Int, 21 SP, 28 Heal
//	Chest        80.3   80.5 (+0.3%)  46.1 /  46.5 (-1.0%)  17 Sta, 20 Int, 17 Spi, 59 Heal
//	Wrist        45.8   46.0 (+0.3%)  27.4 /  28.0 (-2.1%)  10 Int, 20 SP, 31 Heal, 2 MP5
//	Hands        58.7   58.7 (+0.1%)  34.3 /  34.9 (-1.8%)  9 Sta, 12 Int, 24 SP, 38 Heal
//	Waist        63.0   62.9 (-0.1%)  36.6 /  37.4 (-2.2%)  11 Sta, 12 Spi, 26 SP, 39 Heal
//	Legs         80.3   80.3 (-0.0%)  47.4 /  46.5 (+1.7%)  18 Int, 33 SP, 54 Heal, 4 MP5
//	Feet         58.7   58.5 (-0.3%)  33.9 /  34.9 (-2.8%)  11 Sta, 14 Int, 12 Spi, 25 SP
//	Finger       45.8   45.7 (-0.4%)  27.3 /  28.0 (-2.6%)  5 Sta, 10 Int, 21 SP, 28 Heal
//	Finger II    45.8   46.0 (+0.3%)  27.4 /  28.0 (-2.1%)  10 Int, 20 SP, 31 Heal, 2 MP5
//	Main hand   149.1  148.9 (-0.1%)                          28 Sta, 21 Int, 19 Spi, 37 SP, 30 Heal, 14 Crit, 9 MP5
//	Off-hand     43.4   43.2 (-0.4%)                          9 Sta, 8 Int, 6 Spi, 45 Heal
//
// The head and chest form a set with the bonuses of Justice Armor, Forever's Tier 1 Holy set (sim/paladin/
// item_sets_pve.go): every preset carries its tier set's 2-piece bonus only (standards 6), +10 Spirit here.
func init() {
	holy := func(item *proto.UIItem) *proto.UIItem {
		item.RequiredLevel = 60
		item.Phase = 1
		item.Quality = proto.ItemQuality_ItemQualityEpic
		item.Expansion = proto.Expansion_ExpansionVanilla
		item.ClassAllowlist = []proto.Class{proto.Class_ClassPaladin}
		return item
	}
	const set = "Holy Synthetic (Phase 1)"
	ForeverSyntheticItems = append(ForeverSyntheticItems,
		holy(&proto.UIItem{
			Id: 990301, Name: "Holy Synthetic Head (Phase 1)", Icon: "inv_helmet_74",
			Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats:      stats.Stats{stats.Stamina: 18, stats.Intellect: 22, stats.HealingPower: 65, stats.Armor: 608}.ToFloatArray(),
			CritRating: 11,
			Ilvl:       66, SetName: set,
		}),
		holy(&proto.UIItem{
			Id: 990302, Name: "Holy Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_10",
			Type:  proto.ItemType_ItemTypeNeck,
			Stats: stats.Stats{stats.Stamina: 11, stats.Spirit: 11, stats.HealingPower: 42, stats.MP5: 2}.ToFloatArray(),
			Ilvl:  71,
		}),
		holy(&proto.UIItem{
			Id: 990303, Name: "Holy Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_20",
			Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats: stats.Stats{stats.Stamina: 10, stats.Intellect: 14, stats.Spirit: 13, stats.HealingPower: 48, stats.Armor: 562}.ToFloatArray(),
			Ilvl:  66,
		}),
		holy(&proto.UIItem{
			Id: 990304, Name: "Holy Synthetic Back (Phase 1)", Icon: "inv_misc_cape_17",
			Type:  proto.ItemType_ItemTypeBack,
			Stats: stats.Stats{stats.Stamina: 5, stats.Intellect: 10, stats.SpellPower: 21, stats.HealingPower: 28, stats.Armor: 58}.ToFloatArray(),
			Ilvl:  71,
		}),
		holy(&proto.UIItem{
			Id: 990305, Name: "Holy Synthetic Chest (Phase 1)", Icon: "inv_chest_plate03",
			Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats: stats.Stats{stats.Stamina: 17, stats.Intellect: 20, stats.Spirit: 17, stats.HealingPower: 59, stats.Armor: 749}.ToFloatArray(),
			Ilvl:  66, SetName: set,
		}),
		holy(&proto.UIItem{
			Id: 990306, Name: "Holy Synthetic Wrist (Phase 1)", Icon: "inv_bracer_09",
			Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats: stats.Stats{stats.Intellect: 10, stats.SpellPower: 20, stats.HealingPower: 31, stats.MP5: 2, stats.Armor: 353}.ToFloatArray(),
			Ilvl:  71,
		}),
		holy(&proto.UIItem{
			Id: 990307, Name: "Holy Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_29",
			Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats: stats.Stats{stats.Stamina: 9, stats.Intellect: 12, stats.SpellPower: 24, stats.HealingPower: 38, stats.Armor: 468}.ToFloatArray(),
			Ilvl:  66,
		}),
		holy(&proto.UIItem{
			Id: 990308, Name: "Holy Synthetic Waist (Phase 1)", Icon: "inv_belt_27",
			Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats: stats.Stats{stats.Stamina: 11, stats.Spirit: 12, stats.SpellPower: 26, stats.HealingPower: 39, stats.Armor: 453}.ToFloatArray(),
			Ilvl:  71,
		}),
		holy(&proto.UIItem{
			Id: 990309, Name: "Holy Synthetic Legs (Phase 1)", Icon: "inv_pants_plate_06",
			Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats: stats.Stats{stats.Intellect: 18, stats.SpellPower: 33, stats.HealingPower: 54, stats.MP5: 4, stats.Armor: 655}.ToFloatArray(),
			Ilvl:  66,
		}),
		holy(&proto.UIItem{
			Id: 990310, Name: "Holy Synthetic Feet (Phase 1)", Icon: "inv_boots_plate_07",
			Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypePlate,
			Stats: stats.Stats{stats.Stamina: 11, stats.Intellect: 14, stats.Spirit: 12, stats.SpellPower: 25, stats.Armor: 515}.ToFloatArray(),
			Ilvl:  66,
		}),
		holy(&proto.UIItem{
			Id: 990311, Name: "Holy Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_42",
			Type:  proto.ItemType_ItemTypeFinger,
			Stats: stats.Stats{stats.Stamina: 5, stats.Intellect: 10, stats.SpellPower: 21, stats.HealingPower: 28}.ToFloatArray(),
			Ilvl:  71,
		}),
		holy(&proto.UIItem{
			Id: 990312, Name: "Holy Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_43",
			Type:  proto.ItemType_ItemTypeFinger,
			Stats: stats.Stats{stats.Intellect: 10, stats.SpellPower: 20, stats.HealingPower: 31, stats.MP5: 2}.ToFloatArray(),
			Ilvl:  71,
		}),
		// Main hand, item level 71, a one-hand mace with The Hammer of Grace's (11923) damage and speed.
		holy(&proto.UIItem{
			Id: 990313, Name: "Holy Synthetic Main Hand (Phase 1)", Icon: "inv_mace_13",
			Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeMace, HandType: proto.HandType_HandTypeMainHand,
			Stats:           stats.Stats{stats.Stamina: 28, stats.Intellect: 21, stats.Spirit: 19, stats.SpellPower: 37, stats.HealingPower: 30, stats.MP5: 9}.ToFloatArray(),
			CritRating:      14,
			WeaponDamageMin: 71, WeaponDamageMax: 132, WeaponSpeed: 2.7,
			Ilvl: 71,
		}),
		// Off-hand, item level 71, held in the off hand like Brightly Glowing Stone (18523).
		holy(&proto.UIItem{
			Id: 990314, Name: "Holy Synthetic Off-Hand (Phase 1)", Icon: "inv_misc_orb_03",
			Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeOffHand, HandType: proto.HandType_HandTypeOffHand,
			Stats: stats.Stats{stats.Stamina: 9, stats.Intellect: 8, stats.Spirit: 6, stats.HealingPower: 45}.ToFloatArray(),
			Ilvl:  71,
		}),
	)
}
