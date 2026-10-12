package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The druid's synthetic Phase 1 items, made like the enhancement shaman's (forever_synthetic_items.go,
// read its comment first), ids 998000 to 998999. One set for Balance and one for Feral cat, because
// the two share no stat but Stamina.
//
// Each piece takes the place of an Era Phase 1 BiS item the spec wears (named above the piece) and gets
// the full epic budget of its slot at the Phase 1 item level, 66 (Molten Core tier) or the Era item's
// level where that is higher. The budget is the standard all nine classes share, the one the
// enhancement shaman's items reproduce:
//
//	item value = 1.2 * item level - 36.4 (42.8 at 66, 47.6 at 70, 48.8 at 71, 57.2 at 78)
//	slot budget = item value * slot modifier
//	cost of the stats = (sum over stats (price * amount)^1.5)^(2/3)
//
// The slot modifiers are 1 for the head, chest, legs and two-hand weapon, 0.77 for the shoulders, hands,
// waist and feet, and 0.5625 for the neck, back, wrist and rings. The prices are 1 for Strength,
// Agility, Stamina, Intellect and Spirit, 0.55 for a point of spell power (Forever's tuned price, see
// forever_synthetic_items.go), 0.5 for attack power, and 1 for a point of hit or crit rating, at 10
// hit rating and 14 crit rating to the percent. Armor is free. So an item level 66 head has 42.8 to
// spend, an item level 66 shoulder 33.0, an item level 66 ring 24.1, the item level 71 belt and
// shoulders 37.6, the item level 78 neck 32.2 and the item level 70 staff 47.6. Every piece comes within
// a point of its budget (the generator rounds each stat), for example the Balance head: (14^1.5 + 14^1.5
// + 14^1.5 + 5^1.5 + (0.55 * 42)^1.5)^(2/3) = 42.7.
//
// Balance (leather, ids 998001 to 998013) totals 474 spell power, 139 Intellect, 134 Stamina, 77 Spirit,
// 30 hit rating (3%) and 56 crit rating (4%). Spell power is 37% of the set's budget points, Intellect
// 20%, Stamina 19% and Spirit 11%, the share the Forever devs want Spirit to have for a mana user. The
// 3% hit is about what an Era Phase 1 moonkin carried, on top of the 4% from Nature's Reach. The pieces
// lean different ways: the neck and first ring are mostly spell power, the second ring and the feet
// carry the Spirit, the wrist has no Spirit and the first ring only Stamina and spell power. The staff is
// a two-hand weapon, so it gets a full slot's budget like the head. Spell power is the all-school kind.
//
// Feral cat (leather, ids 998101 to 998113) totals 166 Agility, 160 Strength, 138 Stamina, 63 attack
// power, 90 hit rating (9%, an Era Phase 1 cat's hit against a level 63 boss), 56 crit rating (4%), 19
// Intellect and 15 Spirit. The cat fights on Energy, but Shifting Power, which replaced powershifting,
// costs mana (55% of base mana for 40 Energy), so the set gives about 5% of its budget to Intellect and
// Spirit. Agility and Strength split the rest about evenly with Stamina next, and the back, the first
// ring and the weapon put part of theirs into attack power (plain attack power, priced as such).
//
// The head, shoulders, chest, hands, legs and feet of each spec form a set with the bonuses of
// Grovekeeper Eclipse (Balance) or Grovekeeper Ferocity (Feral), Forever's crafted level 60 druid sets
// with those six slots (sim/druid/item_sets_pve.go). The weapons keep the damage and speed of the Era
// weapon they replace (a cat's weapon damage does nothing in Cat Form, only its stats count), and the
// armor is that of the Molten Core leather tier (Cenarion Raiment) for the slot. The generator lived in
// the session scratchpad (druidsets.py), the numbers below are its output.

func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, foreverSyntheticDruidItems...)
}

var foreverSyntheticDruidItems = []*proto.UIItem{
	// Head, item level 66, in place of Cenarion Helm (66, 16834).
	{
		Id: 998001, Name: "Balance Synthetic Head (Phase 1)", Icon: "inv_helmet_09",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Stamina: 14, stats.Intellect: 14, stats.Spirit: 5, stats.SpellPower: 42, stats.Armor: 163}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
		SetName:        "Balance Synthetic (Phase 1)",
	},
	// Neck, item level 78, in place of Choker of the Fire Lord (78, 18814).
	{
		Id: 998002, Name: "Balance Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_17",
		Type:  proto.ItemType_ItemTypeNeck,
		Stats: stats.Stats{stats.Stamina: 7, stats.Intellect: 7, stats.SpellPower: 50}.ToFloatArray(),
		Ilvl:  78, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Shoulder, item level 71, in place of Wild Growth Spaulders (71, 18810).
	{
		Id: 998003, Name: "Balance Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_07",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Stamina: 11, stats.Intellect: 11, stats.Spirit: 6, stats.SpellPower: 41, stats.Armor: 150}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
		SetName:        "Balance Synthetic (Phase 1)",
	},
	// Back, item level 66.
	{
		Id: 998004, Name: "Balance Synthetic Back (Phase 1)", Icon: "inv_misc_cape_16",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 9, stats.Intellect: 5, stats.Spirit: 9, stats.SpellPower: 25, stats.Armor: 55}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Chest, item level 66, in place of Robe of Volatile Power (66, 19145).
	{
		Id: 998005, Name: "Balance Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_06",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Stamina: 15, stats.Intellect: 15, stats.Spirit: 10, stats.SpellPower: 35, stats.Armor: 200}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
		SetName:        "Balance Synthetic (Phase 1)",
	},
	// Wrist, item level 66, in place of Cenarion Bracers (66, 16830).
	{
		Id: 998006, Name: "Balance Synthetic Wrist (Phase 1)", Icon: "inv_bracer_03",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Stamina: 6, stats.Intellect: 11, stats.SpellPower: 30, stats.Armor: 88}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Hands, item level 66, in place of Cenarion Gloves (66, 16831).
	{
		Id: 998007, Name: "Balance Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_08",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Stamina: 10, stats.Intellect: 10, stats.SpellPower: 37, stats.Armor: 125}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
		SetName:        "Balance Synthetic (Phase 1)",
	},
	// Waist, item level 71, in place of Mana Igniting Cord (71, 19136).
	{
		Id: 998008, Name: "Balance Synthetic Waist (Phase 1)", Icon: "inv_belt_06",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Stamina: 12, stats.Intellect: 12, stats.Spirit: 6, stats.SpellPower: 33, stats.Armor: 113}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Legs, item level 66, in place of Cenarion Leggings (66, 16835).
	{
		Id: 998009, Name: "Balance Synthetic Legs (Phase 1)", Icon: "inv_pants_06",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Stamina: 14, stats.Intellect: 14, stats.Spirit: 9, stats.SpellPower: 43, stats.Armor: 175}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
		SetName:        "Balance Synthetic (Phase 1)",
	},
	// Feet, item level 66, in place of Cenarion Boots (66, 16829).
	{
		Id: 998010, Name: "Balance Synthetic Feet (Phase 1)", Icon: "inv_boots_08",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Stamina: 12, stats.Intellect: 12, stats.Spirit: 12, stats.SpellPower: 31, stats.Armor: 138}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
		SetName:        "Balance Synthetic (Phase 1)",
	},
	// Finger, item level 66, in place of Ring of Spell Power (66, 19147).
	{
		Id: 998011, Name: "Balance Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_25",
		Type:  proto.ItemType_ItemTypeFinger,
		Stats: stats.Stats{stats.Stamina: 6, stats.SpellPower: 40}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Finger II, item level 66.
	{
		Id: 998012, Name: "Balance Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_29",
		Type:  proto.ItemType_ItemTypeFinger,
		Stats: stats.Stats{stats.Stamina: 5, stats.Intellect: 11, stats.Spirit: 11, stats.SpellPower: 20}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Two-Hand, item level 70, in place of Staff of Dominance (70, 18842).
	{
		Id: 998013, Name: "Balance Synthetic Two-Hand (Phase 1)", Icon: "inv_staff_13",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand,
		Stats:      stats.Stats{stats.Stamina: 13, stats.Intellect: 17, stats.Spirit: 9, stats.SpellPower: 47}.ToFloatArray(),
		CritRating: 14,
		// Staff of Dominance damage and speed.
		WeaponDamageMin: 126, WeaponDamageMax: 205, WeaponSpeed: 2.9,
		Ilvl: 70, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Head, item level 66, in place of Wolfshead Helm (45, 8345).
	{
		Id: 998101, Name: "Feral Synthetic Head (Phase 1)", Icon: "inv_helmet_04",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Strength: 21, stats.Agility: 15, stats.Stamina: 15, stats.Intellect: 5, stats.Armor: 163}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
		SetName:        "Feral Synthetic (Phase 1)",
	},
	// Neck, item level 66, in place of Onyxia Tooth Pendant (65, 18404).
	{
		Id: 998102, Name: "Feral Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_09",
		Type:      proto.ItemType_ItemTypeNeck,
		Stats:     stats.Stats{stats.Agility: 15, stats.Stamina: 10}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Shoulder, item level 66, in place of Truestrike Shoulders (61, 12927).
	{
		Id: 998103, Name: "Feral Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_25",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Strength: 18, stats.Agility: 12, stats.Stamina: 12, stats.Armor: 150}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
		SetName:        "Feral Synthetic (Phase 1)",
	},
	// Back, item level 66, in place of Cape of the Black Baron (63, 13340).
	{
		Id: 998104, Name: "Feral Synthetic Back (Phase 1)", Icon: "inv_misc_cape_08",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Agility: 14, stats.Stamina: 5, stats.AttackPower: 18, stats.Armor: 55}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Chest, item level 66, in place of Cadaverous Armor (61, 14637).
	{
		Id: 998105, Name: "Feral Synthetic Chest (Phase 1)", Icon: "inv_chest_leather_08",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Strength: 21, stats.Agility: 15, stats.Stamina: 15, stats.Spirit: 5, stats.Armor: 200}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
		SetName:        "Feral Synthetic (Phase 1)",
	},
	// Wrist, item level 66, in place of Wristguards of Stability (65, 19146).
	{
		Id: 998106, Name: "Feral Synthetic Wrist (Phase 1)", Icon: "inv_bracer_07",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Strength: 17, stats.Stamina: 6, stats.Armor: 88}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Hands, item level 66, in place of Devilsaur Gauntlets (58, 15063).
	{
		Id: 998107, Name: "Feral Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_23",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Strength: 13, stats.Agility: 20, stats.Stamina: 7, stats.Armor: 125}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
		SetName:        "Feral Synthetic (Phase 1)",
	},
	// Waist, item level 66, in place of Mugger's Belt (62, 18505).
	{
		Id: 998108, Name: "Feral Synthetic Waist (Phase 1)", Icon: "inv_belt_16",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Strength: 11, stats.Agility: 16, stats.Stamina: 11, stats.Armor: 113}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Legs, item level 66, in place of Devilsaur Leggings (60, 15062).
	{
		Id: 998109, Name: "Feral Synthetic Legs (Phase 1)", Icon: "inv_pants_wolf",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Strength: 18, stats.Agility: 14, stats.Stamina: 14, stats.Intellect: 5, stats.Armor: 175}.ToFloatArray(),
		HitRating: 20,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
		SetName:        "Feral Synthetic (Phase 1)",
	},
	// Feet, item level 66.
	{
		Id: 998110, Name: "Feral Synthetic Feet (Phase 1)", Icon: "inv_boots_08",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Strength: 13, stats.Agility: 19, stats.Stamina: 13, stats.Spirit: 6, stats.Armor: 138}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
		SetName:        "Feral Synthetic (Phase 1)",
	},
	// Finger, item level 67, in place of Quick Strike Ring (67, 18821).
	{
		Id: 998111, Name: "Feral Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_15",
		Type:      proto.ItemType_ItemTypeFinger,
		Stats:     stats.Stats{stats.Agility: 11, stats.Stamina: 6, stats.AttackPower: 23}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      67, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Finger II, item level 66, in place of Don Julio's Band (65, 19325).
	{
		Id: 998112, Name: "Feral Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_09",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Strength: 9, stats.Stamina: 9, stats.Intellect: 5}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Two-Hand, item level 66, in place of Earthshaker (66, 17073).
	{
		Id: 998113, Name: "Feral Synthetic Two-Hand (Phase 1)", Icon: "inv_mace_25",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeMace, HandType: proto.HandType_HandTypeTwoHand,
		Stats:     stats.Stats{stats.Strength: 19, stats.Agility: 15, stats.Stamina: 15, stats.Intellect: 4, stats.Spirit: 4, stats.AttackPower: 22}.ToFloatArray(),
		HitRating: 10,
		// Earthshaker damage and speed.
		WeaponDamageMin: 175, WeaponDamageMax: 263, WeaponSpeed: 3.5,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
}
