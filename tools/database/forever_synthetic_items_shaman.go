package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The synthetic Phase 1 elemental shaman items, ids 990100 to 990199. See forever_synthetic_items.go for
// what these items are and why we need them.
//
// Each piece takes the place of an Era Phase 1 item an elemental shaman would wear: the Earthfury set
// (Molten Core tier, item level 66), Choker of the Fire Lord (78), Deep Earth Spaulders (71), Robe of
// Volatile Power (66), Mana Igniting Cord (71), Ring of Spell Power (66) and Staff of Dominance (70). The
// cloak takes Molten Core's 66. Each piece gets the full epic budget of its slot at that item level.
//
// The budget is the Classic formula. A slot is worth (0.77 * item level - 1) * slot modifier, with the
// modifier 1 for the head, chest, legs and two-hander, 0.75 for the shoulders, hands, waist and feet and
// 0.5625 for the neck, back, wrist and fingers. The stats on a piece add up as (sum of value^1.5)^(2/3),
// where a point of Stamina, Intellect or Spirit is worth 1, a point of spell power 0.55 (Forever's tuned
// price, see forever_synthetic_items.go) and a point of hit or crit rating 1 (10 hit rating is 1% hit,
// 14 crit rating 1% crit). We pick the Stamina, Intellect, Spirit, hit and crit of each piece, and spell
// power fills the rest. For the chest: (0.77 * 66 - 1) * 1 = 49.8, and 14 Stamina, 17 Intellect, 10
// Spirit and 28 crit rating leave room for 24 spell power: (14^1.5 + 17^1.5 + 10^1.5 + 28^1.5 +
// (24 * 0.55)^1.5)^(2/3) = 49.7. The generator script lived in the session scratchpad (synth_ele.py), the
// numbers below are its output.
//
// The set comes to 531 spell power, 149 Intellect, 145 Stamina, 91 Spirit, 84 crit rating (6%) and 30 hit
// rating (3%). As shares of the budget (a stat times its price, added up) that is 37% spell power, 19%
// Intellect, 18% Stamina, 11.5% Spirit, 11% crit and 4% hit:
//   - Spirit is on 9 of the 13 pieces, about 10% of the budget, as on the enhancement set.
//   - The crit sits where the Era pieces had it: 2% on the chest (Robe of Volatile Power), 1% each on the
//     head, waist (Mana Igniting Cord), staff (Staff of Dominance) and the first ring.
//   - 3% hit on the neck, hands and legs. The Era Phase 1 caster gear had little hit, the elemental
//     shaman took 3% from Nature's Guidance, which Forever removed (and since gear hit is unified, the
//     same rating would also count for melee). The set's 2 piece bonus adds 1% more.
//   - Not every piece has every stat: the wrist and neck have no Spirit, the first ring no Intellect, and
//     the hands no Spirit.
//
// The head, shoulders, chest, hands, legs and feet form a set with the bonuses of The Spiritcaller's
// Storm, Forever's elemental Tier 1 set (item set 2111, see sim/shaman/item_sets_pve.go).
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
		Stats:      stats.Stats{stats.Stamina: 15, stats.Intellect: 16, stats.Spirit: 11, stats.SpellPower: 49, stats.Armor: 343}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
		SetName:        "Elemental Synthetic (Phase 1)",
	},
	// Neck, item level 78, in place of Choker of the Fire Lord (78, 18814).
	{
		Id: 990102, Name: "Elemental Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_17",
		Type:      proto.ItemType_ItemTypeNeck,
		Stats:     stats.Stats{stats.Stamina: 9, stats.Intellect: 10, stats.SpellPower: 39}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      78, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Shoulder, item level 71, in place of Deep Earth Spaulders (71, 18829).
	{
		Id: 990103, Name: "Elemental Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_29",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Stamina: 12, stats.Intellect: 12, stats.Spirit: 9, stats.SpellPower: 50, stats.Armor: 335}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
		SetName:        "Elemental Synthetic (Phase 1)",
	},
	// Back, item level 66.
	{
		Id: 990104, Name: "Elemental Synthetic Back (Phase 1)", Icon: "inv_misc_cape_16",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 8, stats.Intellect: 7, stats.Spirit: 8, stats.SpellPower: 35, stats.Armor: 54}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Chest, item level 66, in place of Robe of Volatile Power (66, 19145).
	{
		Id: 990105, Name: "Elemental Synthetic Chest (Phase 1)", Icon: "inv_chest_chain_11",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:      stats.Stats{stats.Stamina: 14, stats.Intellect: 17, stats.Spirit: 10, stats.SpellPower: 24, stats.Armor: 422}.ToFloatArray(),
		CritRating: 28,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
		SetName:        "Elemental Synthetic (Phase 1)",
	},
	// Wrist, item level 66, in place of Earthfury Bracers (66, 16840).
	{
		Id: 990106, Name: "Elemental Synthetic Wrist (Phase 1)", Icon: "inv_bracer_16",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Stamina: 8, stats.Intellect: 10, stats.SpellPower: 38, stats.Armor: 185}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Hands, item level 66, in place of Earthfury Gauntlets (66, 16839).
	{
		Id: 990107, Name: "Elemental Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_11",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Stamina: 10, stats.Intellect: 10, stats.SpellPower: 48, stats.Armor: 264}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
		SetName:        "Elemental Synthetic (Phase 1)",
	},
	// Waist, item level 71, in place of Mana Igniting Cord (71, 19136).
	{
		Id: 990108, Name: "Elemental Synthetic Waist (Phase 1)", Icon: "inv_belt_14",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:      stats.Stats{stats.Stamina: 11, stats.Intellect: 12, stats.Spirit: 9, stats.SpellPower: 39, stats.Armor: 255}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Legs, item level 66, in place of Earthfury Legguards (66, 16843).
	{
		Id: 990109, Name: "Elemental Synthetic Legs (Phase 1)", Icon: "inv_pants_mail_15",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Stamina: 15, stats.Intellect: 16, stats.Spirit: 14, stats.SpellPower: 50, stats.Armor: 369}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
		SetName:        "Elemental Synthetic (Phase 1)",
	},
	// Feet, item level 66, in place of Earthfury Boots (66, 16837).
	{
		Id: 990110, Name: "Elemental Synthetic Feet (Phase 1)", Icon: "inv_boots_plate_06",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Stamina: 12, stats.Intellect: 10, stats.Spirit: 10, stats.SpellPower: 45, stats.Armor: 290}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
		SetName:        "Elemental Synthetic (Phase 1)",
	},
	// Finger, item level 66, in place of Ring of Spell Power (66, 19147).
	{
		Id: 990111, Name: "Elemental Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_38",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Stamina: 7, stats.SpellPower: 33}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Second finger, item level 66, in place of the second Ring of Spell Power.
	{
		Id: 990112, Name: "Elemental Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_29",
		Type:  proto.ItemType_ItemTypeFinger,
		Stats: stats.Stats{stats.Stamina: 8, stats.Intellect: 9, stats.Spirit: 8, stats.SpellPower: 33}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Two-Hand, item level 70.
	{
		Id: 990113, Name: "Elemental Synthetic Staff (Phase 1)", Icon: "inv_staff_13",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand,
		Stats:      stats.Stats{stats.Stamina: 16, stats.Intellect: 20, stats.Spirit: 12, stats.SpellPower: 48}.ToFloatArray(),
		CritRating: 14,
		// Staff of Dominance (70, 18842) damage and speed.
		WeaponDamageMin: 126, WeaponDamageMax: 205, WeaponSpeed: 2.9,
		Ilvl: 70, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
}
