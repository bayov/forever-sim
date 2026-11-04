package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The synthetic Phase 1 Warden shaman (shaman tank) items, ids 999200 to 999299 (999000 to 999099 collided with
// sim/core's test items). See forever_synthetic_items.go for what these items are and why we need them, and
// docs/tank_shaman.md for the set's numbers.
//
// The budget is standards 36 (orchestration/standards.md): a tank piece carries its class's 1b archetype budget,
// hybrid for the shaman (f 1.093 / 1.103 / 1.048, 4 stat lines a piece), with the tank prices on top of 1b's
// weights: Defense 1.5 a point and block value 0.4. Weights: Stamina, Strength, Intellect, Spirit, hit and crit
// rating 1 a point, attack power 0.5, spell power 0.86. Armor is free (the Enhancement set's base armor, slot for
// slot). Item levels as standards 2: head, shoulders, chest, hands, legs and feet 66, the rest 71. The shield is
// 0.72 x L(71) = 58.9, block value and Defense included, at an item level 71 shield's base armor (between Jagged
// Obsidian Shield, 70, and Red Dragonscale Protector, 74). The one-hander is the hybrid's 0.45 x L(71) = 36.8, at
// Deathbringer's speed, spread and 56.4 damage per second (standards 22: the slowest Phase 1 BiS one-hander, for
// Stormstrike; Rockbiter's threat is the same a second at any speed). The totem slot keeps Totem of the Storm: no
// Phase 1 totem does anything for a tank.
//
// The mix is the Warden shaman row of orchestration/gear_targets.md (Tanks, from wowsims/sod's
// phase_4_enh_tank): 23.9% Stamina, 14.3% Strength, 15.5% Intellect, 10% Spirit, 6.4% attack power, 7.2% spell
// power, 6.6% hit, 9.3% crit, 4.6% Defense and 2.4% block value. The set (845.1 weighted, trinkets left out):
//
//	stat        amount  share   row
//	Stamina        202  23.9%  23.9
//	Strength       120  14.2%  14.3
//	Intellect      131  15.5%  15.5
//	Spirit          85  10.1%  10.0
//	AP             108   6.4%   6.4
//	SP              71   7.2%   7.2
//	Hit             55   6.5%   6.6
//	Crit            78   9.2%   9.3
//	Defense         26   4.6%   4.6
//	Block value     50   2.4%   2.4
//
// Per piece (target, weighted linear total, 1.7-norm value / shape target):
//
//	Head       83.3   83.0 (-0.4%)  46.9 / 46.5 (+0.8%)
//	Neck       48.3   48.5 (+0.5%)  28.2 / 28.0 (+0.7%)
//	Shoulder   63.1   63.3 (+0.4%)  35.8 / 34.9 (+2.6%)
//	Back       48.3   48.5 (+0.5%)  27.5 / 28.0 (-2.0%)
//	Chest      83.3   83.0 (-0.4%)  46.9 / 46.5 (+0.8%)
//	Wrist      48.3   48.5 (+0.5%)  27.5 / 28.0 (-2.1%)
//	Hands      63.1   63.0 (-0.1%)  36.1 / 34.9 (+3.3%)
//	Waist      67.7   68.1 (+0.5%)  38.5 / 37.4 (+2.9%)
//	Legs       83.3   83.0 (-0.4%)  46.9 / 46.5 (+0.8%)
//	Feet       63.1   63.3 (+0.4%)  35.8 / 34.9 (+2.6%)
//	Finger     48.3   48.5 (+0.5%)  28.2 / 28.0 (+0.5%)
//	Finger II  48.3   48.3 (+0.1%)  28.3 / 28.0 (+1.0%)
//	Shield     58.9   59.0
//	One-hand   36.8   37.0
//
// Two synthetic trinkets at item level 71 (standards 36 and 1b): the tank trinket, 20 Defense and 40 block value
// (the row's tank stats, 66% Defense and 34% block value; 46.0 linear, 1.7-norm 35.6 against 36.8), and an on-use
// one: the Enhancement set's 335 attack power for 20 sec every 2 min (990015), or the defensive one here, 1,665
// bonus armor for 20 sec every 2 min (sim/shaman/forever_synthetic_trinkets.go), whichever the tank objective
// scores higher (docs/tank_shaman.md).
//
// The hands and legs carry The Spiritcaller's Rage, Forever's melee shaman set (sim/shaman/item_sets_pve.go), as
// the Enhancement set does: only the 2 piece bonus (1% attack and casting speed) is live (standards 6).

func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, wardenSyntheticItems...)
}

func wardenArmor(id int32, name, icon string, slot proto.ItemType, armorType proto.ArmorType, ilvl int32, s stats.Stats, hit, crit float64) *proto.UIItem {
	return &proto.UIItem{
		Id: id, Name: "Warden Synthetic " + name + " (Phase 1)", Icon: icon,
		Type: slot, ArmorType: armorType,
		Stats:     s.ToFloatArray(),
		HitRating: hit, CritRating: crit,
		Ilvl: ilvl, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	}
}

var wardenSyntheticItems = func() []*proto.UIItem {
	mail, cloth, none := proto.ArmorType_ArmorTypeMail, proto.ArmorType_ArmorTypeCloth, proto.ArmorType_ArmorTypeUnknown
	items := []*proto.UIItem{
		wardenArmor(999201, "Head", "inv_helmet_09", proto.ItemType_ItemTypeHead, mail, 66,
			stats.Stats{stats.Stamina: 22, stats.Strength: 20, stats.Intellect: 21, stats.Armor: 392}, 0, 20),
		wardenArmor(999202, "Neck", "inv_jewelry_necklace_09", proto.ItemType_ItemTypeNeck, none, 71,
			stats.Stats{stats.Stamina: 10, stats.Strength: 8, stats.AttackPower: 35}, 13, 0),
		wardenArmor(999203, "Shoulder", "inv_shoulder_25", proto.ItemType_ItemTypeShoulder, mail, 66,
			stats.Stats{stats.Stamina: 16, stats.Intellect: 16, stats.Spirit: 15, stats.SpellPower: 19, stats.Armor: 317}, 0, 0),
		wardenArmor(999204, "Back", "inv_misc_cape_05", proto.ItemType_ItemTypeBack, cloth, 71,
			stats.Stats{stats.Stamina: 13, stats.Strength: 11, stats.Defense: 9, stats.Armor: 54}, 0, 11),
		wardenArmor(999205, "Chest", "inv_chest_chain_11", proto.ItemType_ItemTypeChest, mail, 66,
			stats.Stats{stats.Stamina: 21, stats.Strength: 20, stats.Intellect: 21, stats.Spirit: 21, stats.Armor: 422}, 0, 0),
		wardenArmor(999206, "Wrist", "inv_bracer_02", proto.ItemType_ItemTypeWrist, mail, 71,
			stats.Stats{stats.Stamina: 11, stats.Intellect: 12, stats.Spirit: 12, stats.Defense: 9, stats.Armor: 198}, 0, 0),
		wardenArmor(999207, "Hands", "inv_gauntlets_26", proto.ItemType_ItemTypeHands, mail, 66,
			stats.Stats{stats.Stamina: 15, stats.Strength: 13, stats.Armor: 264}, 21, 14),
		wardenArmor(999208, "Waist", "inv_belt_23", proto.ItemType_ItemTypeWaist, mail, 71,
			stats.Stats{stats.Stamina: 17, stats.Intellect: 17, stats.SpellPower: 21, stats.Armor: 255}, 0, 16),
		wardenArmor(999209, "Legs", "inv_pants_06", proto.ItemType_ItemTypeLegs, mail, 66,
			stats.Stats{stats.Stamina: 21, stats.Strength: 20, stats.Intellect: 21, stats.Spirit: 21, stats.Armor: 369}, 0, 0),
		wardenArmor(999210, "Feet", "inv_boots_01", proto.ItemType_ItemTypeFeet, mail, 66,
			stats.Stats{stats.Stamina: 16, stats.Strength: 15, stats.Spirit: 16, stats.SpellPower: 19, stats.Armor: 290}, 0, 0),
		wardenArmor(999211, "Finger", "inv_jewelry_ring_07", proto.ItemType_ItemTypeFinger, none, 71,
			stats.Stats{stats.Stamina: 9, stats.AttackPower: 35}, 13, 9),
		wardenArmor(999212, "Finger II", "inv_jewelry_ring_15", proto.ItemType_ItemTypeFinger, none, 71,
			stats.Stats{stats.Stamina: 9, stats.Intellect: 10, stats.AttackPower: 38, stats.SpellPower: 12}, 0, 0),
		{
			Id: 999213, Name: "Warden Synthetic Main Hand (Phase 1)", Icon: "inv_axe_09",
			Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeAxe, HandType: proto.HandType_HandTypeOneHand,
			Stats:     stats.Stats{stats.Strength: 13, stats.Stamina: 8}.ToFloatArray(),
			HitRating: 8, CritRating: 8,
			// Deathbringer's speed, spread and damage per second.
			WeaponDamageMin: 114, WeaponDamageMax: 213, WeaponSpeed: 2.9,
			Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
			ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
		},
		{
			Id: 999214, Name: "Warden Synthetic Shield (Phase 1)", Icon: "inv_shield_10",
			Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeShield, HandType: proto.HandType_HandTypeOffHand,
			Stats: stats.Stats{stats.Stamina: 14, stats.Intellect: 13, stats.Defense: 8, stats.BlockValue: 50, stats.Armor: 2681}.ToFloatArray(),
			Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
			ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
		},
		// The tank trinket, item level 71: the row's Defense and block value.
		{
			Id: 999215, Name: "Warden Synthetic Trinket (Phase 1)", Icon: "inv_jewelry_talisman_02",
			Type:  proto.ItemType_ItemTypeTrinket,
			Stats: stats.Stats{stats.Defense: 20, stats.BlockValue: 40}.ToFloatArray(),
			Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
			ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
		},
		// The defensive on-use trinket, item level 71: 1,665 bonus armor for 20 sec every 2 min
		// (sim/shaman/forever_synthetic_trinkets.go), on a timer in the rotation.
		{
			Id: 999216, Name: "Warden Synthetic On-Use Trinket (Phase 1)", Icon: "inv_misc_stonetablet_05",
			Type:  proto.ItemType_ItemTypeTrinket,
			Stats: stats.Stats{}.ToFloatArray(),
			Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
			ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
		},
	}
	// Two pieces of The Spiritcaller's Rage, so its 2 piece bonus is live (standards 6).
	items[6].SetName = "Warden Synthetic (Phase 1)"
	items[8].SetName = "Warden Synthetic (Phase 1)"
	return items
}()
