package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The synthetic Phase 1 elemental shaman items, ids 990100 to 990199. See forever_synthetic_items.go for
// what these items are and why we need them.
//
// Each piece takes the place of an Era Phase 1 item an elemental shaman would wear: the Earthfury set,
// Choker of the Fire Lord, Deep Earth Spaulders, Robe of Volatile Power, Mana Igniting Cord, Ring of Spell
// Power and Staff of Dominance. The cloak takes the place of the Phase 1 caster cloaks.
//
// The budget follows the cross-class standard (standards item 1b on the coordination branch, the
// Forever budget per archetype). The weights: Stamina, Intellect, Spirit, hit and crit rating 1 a point
// (10 hit rating is 1%, 14 crit rating 1%), spell power 0.86. Armor is free. The line is L(ilvl) = 1.125
// * ilvl + 2.0, and each armor piece carries slot mod * L * f in weighted linear stats, within 1%: slot
// mods 1 (head, chest, legs), 0.75 (shoulders, hands, waist, feet) and 0.5625 (neck, back, wrist,
// rings), f 1.053 / 1.026 / 0.995 for a pure caster. The head, shoulders, chest, hands, legs and feet are
// item level 66, the neck, back, wrist, waist and rings 71. Each piece also passes the shape check: its
// 1.7-norm value, (sum of (weight * amount)^1.7)^(1/1.7), within 6% of slot mod * (0.661 * ilvl + 2.92),
// which takes 4 stat lines a piece. The staff gets Forever's caster weapon rate, 2.21 * L(76) = 193.4. The
// totem has no stats (Burning Totem, its effect is its value).
//
// The mix follows standards 5a, the Elemental row of orchestration/gear_targets.md (after Season of
// Discovery level 60 elemental BiS): 16.1% Stamina, 19.4% Intellect, 10% Spirit, 37.8% spell power, 4.8%
// hit and 11.9% crit. The hit goes to the cap (item 4: 16% spell hit against a level 63 boss with Tidal
// Focus 5 and the 2 piece bonus's 1%, so 100 rating). Measured against the hit-adjusted target
// (gear_targets.md 1a6968081): Stamina and Spirit hold their shares and Intellect, spell power and crit
// share what the capped hit leaves in the table's proportions. A stat's share is its weight times its
// amount over the set's weighted total (902.6, the trinkets left out):
//
//	stat      amount   share   hit-adjusted target
//	Stamina      137   15.2%   16.1 +- 1.5
//	Intellect    164   18.2%   17.6 +- 3
//	Spirit        86    9.5%   10
//	SP           367   35.0%   34.4 +- 3
//	Crit         100   11.1%   10.8 +- 3
//	Hit          100   11.1%   at the cap
//
// Per piece (target, weighted linear total, 1.7-norm value / shape target, stats). The armor comes to 709.3
// and the staff to 193.3:
//
//	Head         80.3    80.0 (-0.4%)   47.0 /  46.5 (+0.9%)  13 Sta, 16 Int, 20 Crit, 36 SP
//	Neck         45.8    46.2 (+0.8%)   27.1 /  28.0 (-3.2%)  7 Sta, 9 Int, 13 Hit, 20 SP
//	Shoulder     58.7    58.5 (-0.3%)   34.1 /  34.9 (-2.4%)  10 Sta, 12 Int, 15 Spi, 25 SP
//	Back         45.8    46.2 (+0.8%)   27.0 /  28.0 (-3.9%)  8 Sta, 9 Int, 12 Spi, 20 SP
//	Chest        80.3    80.0 (-0.4%)   47.0 /  46.5 (+0.9%)  13 Sta, 16 Int, 20 Crit, 36 SP
//	Wrist        45.8    46.2 (+0.8%)   27.1 /  28.0 (-3.2%)  7 Sta, 9 Int, 13 Hit, 20 SP
//	Hands        58.7    58.4 (-0.5%)   34.4 /  34.9 (-1.5%)  9 Sta, 11 Int, 16 Hit, 26 SP
//	Waist        63.0    62.9 (-0.1%)   37.2 /  37.4 (-0.6%)  10 Sta, 12 Int, 16 Crit, 29 SP
//	Legs         80.3    80.1 (-0.2%)   47.0 /  46.5 (+1.0%)  13 Sta, 15 Int, 22 Hit, 35 SP
//	Feet         58.7    58.5 (-0.3%)   34.1 /  34.9 (-2.4%)  10 Sta, 12 Int, 15 Spi, 25 SP
//	Finger       45.8    46.2 (+0.8%)   27.0 /  28.0 (-3.9%)  8 Sta, 9 Int, 12 Crit, 20 SP
//	Finger II    45.8    46.2 (+0.8%)   27.0 /  28.0 (-3.9%)  8 Sta, 9 Int, 12 Spi, 20 SP
//	Staff       193.4   193.3 (-0.1%)  -              21 Sta, 25 Int, 32 Spi, 32 Crit, 36 Hit, 55 SP
//
// Two synthetic trinkets at item level 71 (standards 8a and 1b): a stat trinket with 37 spell power and 14
// crit rating (45.8 linear, 1.7-norm 36.2 against 36.8), and an on-use trinket of 140 spell power for 20 sec
// every 2 min (26.8 weighted points over 180 sec, sim/shaman/forever_synthetic_trinkets.go). Before this
// round (2026-10-13 morning) the set followed standards 1a, 389 spell power at 0.55 and no trinkets. The
// generator script lived in the session scratchpad (synth_1b.py), the numbers below are its output.
//
// The hands and legs carry The Spiritcaller's Storm, Forever's elemental Tier 1 set (item set 2111, see
// sim/shaman/item_sets_pve.go). The standard keeps only the 2 piece bonus live (+1% hit), so exactly two
// pieces are in the set. They are the slots of the crafted Spiritcaller Gloves and Kilt.
//
// The staff is a staff on purpose: Staff of Dominance is the Era Phase 1 weapon, and no shaman race has a
// staff racial, so the race ranking doesn't depend on the weapon type (Orc axes and Dwarf maces would
// give 1% crit).

func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, elementalSyntheticItems...)
}

var elementalSyntheticItems = []*proto.UIItem{
	// Head, item level 66, in place of Earthfury Helmet (66, 16842).
	{
		Id: 990101, Name: "Elemental Synthetic Head (Phase 1)", Icon: "inv_helmet_09",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:      stats.Stats{stats.Stamina: 13, stats.Intellect: 16, stats.SpellPower: 36, stats.Armor: 343}.ToFloatArray(),
		CritRating: 20,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Neck, item level 71, in place of Choker of the Fire Lord (78, 18814).
	{
		Id: 990102, Name: "Elemental Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_17",
		Type:      proto.ItemType_ItemTypeNeck,
		Stats:     stats.Stats{stats.Stamina: 7, stats.Intellect: 9, stats.SpellPower: 20}.ToFloatArray(),
		HitRating: 13,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Shoulder, item level 66, in place of Deep Earth Spaulders (71, 18829).
	{
		Id: 990103, Name: "Elemental Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_29",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Stamina: 10, stats.Intellect: 12, stats.Spirit: 15, stats.SpellPower: 25, stats.Armor: 399}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Back, item level 71.
	{
		Id: 990104, Name: "Elemental Synthetic Back (Phase 1)", Icon: "inv_misc_cape_16",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 8, stats.Intellect: 9, stats.Spirit: 12, stats.SpellPower: 20, stats.Armor: 54}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Chest, item level 66, in place of Robe of Volatile Power (66, 19145).
	{
		Id: 990105, Name: "Elemental Synthetic Chest (Phase 1)", Icon: "inv_chest_chain_11",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:      stats.Stats{stats.Stamina: 13, stats.Intellect: 16, stats.SpellPower: 36, stats.Armor: 422}.ToFloatArray(),
		CritRating: 20,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Wrist, item level 71, in place of Earthfury Bracers (66, 16840).
	{
		Id: 990106, Name: "Elemental Synthetic Wrist (Phase 1)", Icon: "inv_bracer_16",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Stamina: 7, stats.Intellect: 9, stats.SpellPower: 20, stats.Armor: 185}.ToFloatArray(),
		HitRating: 13,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Hands, item level 66, in place of Earthfury Gauntlets (66, 16839).
	{
		Id: 990107, Name: "Elemental Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_11",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Stamina: 9, stats.Intellect: 11, stats.SpellPower: 26, stats.Armor: 264}.ToFloatArray(),
		HitRating: 16,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
		SetName:        "Elemental Synthetic (Phase 1)",
	},
	// Waist, item level 71, in place of Mana Igniting Cord (71, 19136).
	{
		Id: 990108, Name: "Elemental Synthetic Waist (Phase 1)", Icon: "inv_belt_14",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:      stats.Stats{stats.Stamina: 10, stats.Intellect: 12, stats.SpellPower: 29, stats.Armor: 255}.ToFloatArray(),
		CritRating: 16,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Legs, item level 66, in place of Earthfury Legguards (66, 16843).
	{
		Id: 990109, Name: "Elemental Synthetic Legs (Phase 1)", Icon: "inv_pants_mail_15",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Stamina: 13, stats.Intellect: 15, stats.SpellPower: 35, stats.Armor: 369}.ToFloatArray(),
		HitRating: 22,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
		SetName:        "Elemental Synthetic (Phase 1)",
	},
	// Feet, item level 66, in place of Earthfury Boots (66, 16837).
	{
		Id: 990110, Name: "Elemental Synthetic Feet (Phase 1)", Icon: "inv_boots_plate_06",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Stamina: 10, stats.Intellect: 12, stats.Spirit: 15, stats.SpellPower: 25, stats.Armor: 290}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Finger, item level 71, in place of Ring of Spell Power (66, 19147).
	{
		Id: 990111, Name: "Elemental Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_38",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Stamina: 8, stats.Intellect: 9, stats.SpellPower: 20}.ToFloatArray(),
		CritRating: 12,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Second finger, item level 71, in place of the second Ring of Spell Power.
	{
		Id: 990112, Name: "Elemental Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_29",
		Type:  proto.ItemType_ItemTypeFinger,
		Stats: stats.Stats{stats.Stamina: 8, stats.Intellect: 9, stats.Spirit: 12, stats.SpellPower: 20}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Two-Hand, item level 76.
	{
		Id: 990113, Name: "Elemental Synthetic Staff (Phase 1)", Icon: "inv_staff_13",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand,
		Stats:     stats.Stats{stats.Stamina: 21, stats.Intellect: 25, stats.Spirit: 32, stats.SpellPower: 55}.ToFloatArray(),
		HitRating: 36, CritRating: 32,
		// Staff of Dominance (70, 18842) damage and speed.
		WeaponDamageMin: 126, WeaponDamageMax: 205, WeaponSpeed: 2.9,
		Ilvl: 76, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Stat trinket, item level 71 (standards 1b: 45.9 linear, 1.7-norm 36.2).
	{
		Id: 990114, Name: "Elemental Synthetic Trinket (Phase 1)", Icon: "inv_jewelry_talisman_04",
		Type:       proto.ItemType_ItemTypeTrinket,
		Stats:      stats.Stats{stats.SpellPower: 37}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// On-use trinket, item level 71: 140 spell power for 20 sec every 2 min (sim/shaman/forever_synthetic_trinkets.go).
	{
		Id: 990115, Name: "Elemental Synthetic On-Use Trinket (Phase 1)", Icon: "inv_misc_stonetablet_11",
		Type:  proto.ItemType_ItemTypeTrinket,
		Stats: stats.Stats{}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
}
