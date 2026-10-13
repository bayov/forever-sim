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
// The budget and the item levels follow the cross-class standard every synthetic set shares
// (orchestration/standards.md on the coordination branch). An epic piece is worth V = 1.2 * item level -
// 36.4, times the slot modifier: 1 for the head, chest and legs, 0.77 for the shoulders, hands, waist and
// feet, and 0.5625 for the neck, back, wrist and fingers. The head, shoulders, chest, hands, legs and
// feet are item level 66 (42.8), the neck, back, wrist, waist and rings 71 (48.8). A caster's weapon slots
// are worth about 73 in all (Azuresong Mageblade's 46.2 plus an off-hand at 71, 26.8), and the staff
// gets that whole 73 (item level 76, the standard's two-hander level).
//
// Each piece is spent linearly (standards item 1a): its weighted stat total, where a point of Stamina,
// Intellect, Spirit, hit rating or crit rating weighs 1 and a point of spell power 0.55 (Forever's tuned
// price, see forever_synthetic_items.go), is 1.73 times its budget, within 1%. Armor is free. For the
// chest: 1.73 * 42.8 = 74.0, from 15 Stamina, 17 Intellect, 8 Spirit, 11 crit rating and 42 spell power
// (15 + 17 + 8 + 11 + 42 * 0.55 = 74.1).
//
// The mix is the fair stat mix of the standard (item 5a, the Elemental row of
// orchestration/gear_targets.md at e3f2ae044, after Season of Discovery level 60 elemental BiS). On each
// piece the hit comes first (10 rating on 10 of the 13 pieces), then Stamina, Intellect, Spirit, crit and
// spell power in the target proportions. The hit is at the spell hit cap against a level 63 boss (16%,
// with Tidal Focus 5 and the 2 piece bonus's 1%), a little under its 12.9 target, so crit takes the rest.
// A stat's share is its weight times its amount over the set's weighted total (821.0):
//
//	stat      amount   share   target
//	Stamina      150   18.3%   18 +- 1.5
//	Intellect    169   20.6%   20.7
//	Spirit        78    9.5%   10
//	SP           389   26.1%   25.8
//	Crit         110   13.4%   12.6 (+ the hit's 0.7)
//	Hit          100   12.2%   12.9, at the cap
//
// That is 389 spell power, 169 Intellect, 150 Stamina, 78 Spirit, 110 crit rating (7.9%) and 100 hit
// rating (10%). The generator script lived in the session scratchpad (synth_ele5.py), the numbers below
// are its output. Before the fair mix (2026-10-12) the set was 504 spell power, 127 Intellect, 121
// Stamina, 79 Spirit, 70 crit and 100 hit rating, priced with the 1.5-norm.
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
		Stats:     stats.Stats{stats.Stamina: 13, stats.Intellect: 15, stats.Spirit: 7, stats.SpellPower: 35, stats.Armor: 343}.ToFloatArray(),
		HitRating: 10, CritRating: 10,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Neck, item level 71, in place of Choker of the Fire Lord (78, 18814).
	{
		Id: 990102, Name: "Elemental Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_17",
		Type:      proto.ItemType_ItemTypeNeck,
		Stats:     stats.Stats{stats.Stamina: 8, stats.Intellect: 9, stats.Spirit: 4, stats.SpellPower: 19}.ToFloatArray(),
		HitRating: 10, CritRating: 6,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Shoulder, item level 66, in place of Deep Earth Spaulders (71, 18829).
	{
		Id: 990103, Name: "Elemental Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_29",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Stamina: 10, stats.Intellect: 11, stats.Spirit: 5, stats.SpellPower: 25, stats.Armor: 399}.ToFloatArray(),
		HitRating: 10, CritRating: 7,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Back, item level 71.
	{
		Id: 990104, Name: "Elemental Synthetic Back (Phase 1)", Icon: "inv_misc_cape_16",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 10, stats.Intellect: 11, stats.Spirit: 5, stats.SpellPower: 26, stats.Armor: 54}.ToFloatArray(),
		CritRating: 7,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Chest, item level 66, in place of Robe of Volatile Power (66, 19145).
	{
		Id: 990105, Name: "Elemental Synthetic Chest (Phase 1)", Icon: "inv_chest_chain_11",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:      stats.Stats{stats.Stamina: 15, stats.Intellect: 17, stats.Spirit: 8, stats.SpellPower: 42, stats.Armor: 422}.ToFloatArray(),
		CritRating: 11,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Wrist, item level 71, in place of Earthfury Bracers (66, 16840).
	{
		Id: 990106, Name: "Elemental Synthetic Wrist (Phase 1)", Icon: "inv_bracer_16",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Stamina: 8, stats.Intellect: 9, stats.Spirit: 4, stats.SpellPower: 19, stats.Armor: 185}.ToFloatArray(),
		HitRating: 10, CritRating: 6,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Hands, item level 66, in place of Earthfury Gauntlets (66, 16839).
	{
		Id: 990107, Name: "Elemental Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_11",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Stamina: 10, stats.Intellect: 11, stats.Spirit: 5, stats.SpellPower: 25, stats.Armor: 264}.ToFloatArray(),
		HitRating: 10, CritRating: 7,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
		SetName:        "Elemental Synthetic (Phase 1)",
	},
	// Waist, item level 71, in place of Mana Igniting Cord (71, 19136).
	{
		Id: 990108, Name: "Elemental Synthetic Waist (Phase 1)", Icon: "inv_belt_14",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Stamina: 11, stats.Intellect: 13, stats.Spirit: 6, stats.SpellPower: 31, stats.Armor: 255}.ToFloatArray(),
		HitRating: 10, CritRating: 8,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Legs, item level 66, in place of Earthfury Legguards (66, 16843).
	{
		Id: 990109, Name: "Elemental Synthetic Legs (Phase 1)", Icon: "inv_pants_mail_15",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Stamina: 13, stats.Intellect: 15, stats.Spirit: 7, stats.SpellPower: 35, stats.Armor: 369}.ToFloatArray(),
		HitRating: 10, CritRating: 10,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
		SetName:        "Elemental Synthetic (Phase 1)",
	},
	// Feet, item level 66, in place of Earthfury Boots (66, 16837).
	{
		Id: 990110, Name: "Elemental Synthetic Feet (Phase 1)", Icon: "inv_boots_plate_06",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Stamina: 10, stats.Intellect: 11, stats.Spirit: 5, stats.SpellPower: 25, stats.Armor: 290}.ToFloatArray(),
		HitRating: 10, CritRating: 7,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Finger, item level 71, in place of Ring of Spell Power (66, 19147).
	{
		Id: 990111, Name: "Elemental Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_38",
		Type:      proto.ItemType_ItemTypeFinger,
		Stats:     stats.Stats{stats.Stamina: 8, stats.Intellect: 9, stats.Spirit: 4, stats.SpellPower: 19}.ToFloatArray(),
		HitRating: 10, CritRating: 6,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Second finger, item level 71, in place of the second Ring of Spell Power.
	{
		Id: 990112, Name: "Elemental Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_29",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Stamina: 10, stats.Intellect: 11, stats.Spirit: 5, stats.SpellPower: 26}.ToFloatArray(),
		CritRating: 7,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Two-Hand, item level 76.
	{
		Id: 990113, Name: "Elemental Synthetic Staff (Phase 1)", Icon: "inv_staff_13",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand,
		Stats:     stats.Stats{stats.Stamina: 24, stats.Intellect: 27, stats.Spirit: 13, stats.SpellPower: 62}.ToFloatArray(),
		HitRating: 10, CritRating: 18,
		// Staff of Dominance (70, 18842) damage and speed.
		WeaponDamageMin: 126, WeaponDamageMax: 205, WeaponSpeed: 2.9,
		Ilvl: 76, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
}
