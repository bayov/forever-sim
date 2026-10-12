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
// the full epic budget of its slot under the cross-class standards (orchestration/standards.md on the
// shared branch): one item level table for every class, 66 for the head, shoulders, chest, hands, legs
// and feet and 71 for the neck, back, wrist, waist and rings, two-handers at 76. The budget:
//
//	item value = 1.2 * item level - 36.4 (42.8 at 66, 48.8 at 71, 54.8 at 76)
//	slot budget = item value * slot modifier
//	cost of the stats = (sum over stats (price * amount)^1.5)^(2/3)
//
// The slot modifiers are 1 for the head, chest, legs and two-hand weapon, 0.77 for the shoulders, hands,
// waist and feet, and 0.5625 for the neck, back, wrist and rings. The prices are 1 for Strength,
// Agility, Stamina, Intellect and Spirit, 0.55 for a point of spell power (Forever's tuned price, see
// forever_synthetic_items.go), 0.5 for attack power, and 1 for a point of hit or crit rating, at 10
// hit rating and 14 crit rating to the percent. Armor is free. So the twelve armor slots have 402.5 to
// spend for either spec, as for every class: 42.8 for a head, 33.0 for a shoulder, 37.6 for the belt
// and 27.5 for a ring. Every piece comes within a point of its budget and none goes over (the
// generator rounds each stat and then takes points off until it fits), for example the Balance head:
// (10^1.5 + 14^1.5 + 13^1.5 + 13^1.5 + 4^1.5 + (0.55 * 37)^1.5)^(2/3) = 42.5.
//
// Balance (leather, ids 998001 to 998013) totals 465 spell power, 141 Intellect, 132 Stamina, 77 Spirit,
// 110 hit rating and 56 crit rating (4%). The hit takes the moonkin to the spell hit cap against a
// level 63 boss and no further: 16% with Nature's Reach (4%) and the Grovekeeper 2 piece (1%). Spirit is
// 10% of the set's budget points, the share the Forever devs want it to have for a mana user. The staff
// gets the casters' weapon budget, 73.0, what a caster's main hand (Azuresong Mageblade's stats at Era
// prices, 46.2) and an item level 71 off-hand (26.8) have together. Spell power is the all-school kind.
//
// Feral cat (leather, ids 998101 to 998113) totals 178 Agility, 170 Strength, 151 Stamina, 76 attack
// power, 90 hit rating (9%, the cat's special attack hit cap against a level 63 boss), 56 crit rating
// (4%), 21 Intellect and 16 Spirit. The cat fights on Energy, but Shifting Power, which replaced
// powershifting, costs mana (55% of base mana for 40 Energy), so the set gives about 5% of its budget to
// Intellect and Spirit. Agility and Strength split the rest about evenly with Stamina next, and the
// back, the first ring and the weapon put part of theirs into attack power (plain attack power, priced
// as such). The two-hander is at 76 like every class's.
//
// The head and chest of each spec, and only those two, form a set with the bonuses of Grovekeeper
// Eclipse (Balance) or Grovekeeper Ferocity (Feral), Forever's crafted level 60 druid sets
// (sim/druid/item_sets_pve.go), so only the 2 piece bonus is live (1% hit, or 1% attack and casting
// speed), the same shape as every class's set. The weapons keep the damage and speed of the Era weapon
// they replace (a cat's weapon damage does nothing in Cat Form, only its stats count), and the armor is
// that of the Molten Core leather tier (Cenarion Raiment) for the slot. The relic slot takes a real
// Forever Phase 1 idol, Swarming Idol for Balance and Howling Idol for Feral (sim/druid/items.go). The
// generator lived in the session scratchpad (druidsets2.py), the numbers below are its output.

func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, foreverSyntheticDruidItems...)
}

var foreverSyntheticDruidItems = []*proto.UIItem{
	// Head, item level 66, in place of Cenarion Helm (66, 16834).
	{
		Id: 998001, Name: "Balance Synthetic Head (Phase 1)", Icon: "inv_helmet_09",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Stamina: 13, stats.Intellect: 13, stats.Spirit: 4, stats.SpellPower: 37, stats.Armor: 163}.ToFloatArray(),
		HitRating: 10, CritRating: 14,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
		SetName:        "Balance Synthetic (Phase 1)",
	},
	// Neck, item level 71, in place of Choker of the Fire Lord (78, 18814).
	{
		Id: 998002, Name: "Balance Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_17",
		Type:      proto.ItemType_ItemTypeNeck,
		Stats:     stats.Stats{stats.Stamina: 5, stats.Intellect: 5, stats.SpellPower: 36}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Shoulder, item level 66, in place of Wild Growth Spaulders (71, 18810).
	{
		Id: 998003, Name: "Balance Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_07",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Stamina: 10, stats.Intellect: 10, stats.Spirit: 5, stats.SpellPower: 34, stats.Armor: 159}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Back, item level 71.
	{
		Id: 998004, Name: "Balance Synthetic Back (Phase 1)", Icon: "inv_misc_cape_16",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 9, stats.Intellect: 4, stats.Spirit: 9, stats.SpellPower: 24, stats.Armor: 55}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Chest, item level 66, in place of Robe of Volatile Power (66, 19145).
	{
		Id: 998005, Name: "Balance Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_06",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Stamina: 13, stats.Intellect: 13, stats.Spirit: 9, stats.SpellPower: 32, stats.Armor: 200}.ToFloatArray(),
		HitRating: 10, CritRating: 14,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
		SetName:        "Balance Synthetic (Phase 1)",
	},
	// Wrist, item level 71, in place of Cenarion Bracers (66, 16830).
	{
		Id: 998006, Name: "Balance Synthetic Wrist (Phase 1)", Icon: "inv_bracer_03",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Stamina: 5, stats.Intellect: 11, stats.SpellPower: 29, stats.Armor: 88}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
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
	},
	// Waist, item level 71, in place of Mana Igniting Cord (71, 19136).
	{
		Id: 998008, Name: "Balance Synthetic Waist (Phase 1)", Icon: "inv_belt_06",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Stamina: 11, stats.Intellect: 11, stats.Spirit: 5, stats.SpellPower: 28, stats.Armor: 113}.ToFloatArray(),
		HitRating: 10, CritRating: 14,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
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
	},
	// Feet, item level 66, in place of Cenarion Boots (66, 16829).
	{
		Id: 998010, Name: "Balance Synthetic Feet (Phase 1)", Icon: "inv_boots_08",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Stamina: 10, stats.Intellect: 10, stats.Spirit: 10, stats.SpellPower: 28, stats.Armor: 138}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Finger, item level 71, in place of Ring of Spell Power (66, 19147).
	{
		Id: 998011, Name: "Balance Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_25",
		Type:      proto.ItemType_ItemTypeFinger,
		Stats:     stats.Stats{stats.Stamina: 5, stats.SpellPower: 39}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Finger II, item level 71.
	{
		Id: 998012, Name: "Balance Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_29",
		Type:  proto.ItemType_ItemTypeFinger,
		Stats: stats.Stats{stats.Stamina: 6, stats.Intellect: 12, stats.Spirit: 12, stats.SpellPower: 22}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Two-Hand, item level 76, in place of Staff of Dominance (70, 18842), with the casters' weapon budget (73.0, see above).
	{
		Id: 998013, Name: "Balance Synthetic Two-Hand (Phase 1)", Icon: "inv_staff_13",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand,
		Stats:      stats.Stats{stats.Stamina: 21, stats.Intellect: 28, stats.Spirit: 14, stats.SpellPower: 76}.ToFloatArray(),
		CritRating: 14,
		// Staff of Dominance damage and speed.
		WeaponDamageMin: 126, WeaponDamageMax: 205, WeaponSpeed: 2.9,
		Ilvl: 76, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
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
	// Neck, item level 71, in place of Onyxia Tooth Pendant (74, 18404).
	{
		Id: 998102, Name: "Feral Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_09",
		Type:      proto.ItemType_ItemTypeNeck,
		Stats:     stats.Stats{stats.Agility: 17, stats.Stamina: 12}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Shoulder, item level 66, in place of Truestrike Shoulders (61, 12927).
	{
		Id: 998103, Name: "Feral Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_25",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Strength: 17, stats.Agility: 12, stats.Stamina: 12, stats.Armor: 150}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Back, item level 71, in place of Cape of the Black Baron (63, 13340).
	{
		Id: 998104, Name: "Feral Synthetic Back (Phase 1)", Icon: "inv_misc_cape_08",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Agility: 16, stats.Stamina: 5, stats.AttackPower: 21, stats.Armor: 55}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
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
	// Wrist, item level 71, in place of Wristguards of Stability (65, 19146).
	{
		Id: 998106, Name: "Feral Synthetic Wrist (Phase 1)", Icon: "inv_bracer_07",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Strength: 20, stats.Stamina: 7, stats.Armor: 88}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
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
	},
	// Waist, item level 71, in place of Mugger's Belt (62, 18505).
	{
		Id: 998108, Name: "Feral Synthetic Waist (Phase 1)", Icon: "inv_belt_16",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Strength: 13, stats.Agility: 19, stats.Stamina: 13, stats.Armor: 113}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Legs, item level 66, in place of Devilsaur Leggings (60, 15062).
	{
		Id: 998109, Name: "Feral Synthetic Legs (Phase 1)", Icon: "inv_pants_wolf",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Strength: 17, stats.Agility: 14, stats.Stamina: 14, stats.Intellect: 5, stats.Armor: 175}.ToFloatArray(),
		HitRating: 20,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Feet, item level 66.
	{
		Id: 998110, Name: "Feral Synthetic Feet (Phase 1)", Icon: "inv_boots_08",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Strength: 13, stats.Agility: 18, stats.Stamina: 13, stats.Spirit: 6, stats.Armor: 138}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Finger, item level 71, in place of Quick Strike Ring (67, 18821).
	{
		Id: 998111, Name: "Feral Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_15",
		Type:      proto.ItemType_ItemTypeFinger,
		Stats:     stats.Stats{stats.Agility: 12, stats.Stamina: 7, stats.AttackPower: 26}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Finger II, item level 71, in place of Don Julio's Band (65, 19325).
	{
		Id: 998112, Name: "Feral Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_09",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Strength: 11, stats.Stamina: 11, stats.Intellect: 6}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Two-Hand, item level 76, in place of Earthshaker (66, 17073).
	{
		Id: 998113, Name: "Feral Synthetic Two-Hand (Phase 1)", Icon: "inv_mace_25",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeMace, HandType: proto.HandType_HandTypeTwoHand,
		Stats:     stats.Stats{stats.Strength: 24, stats.Agility: 20, stats.Stamina: 20, stats.Intellect: 5, stats.Spirit: 5, stats.AttackPower: 29}.ToFloatArray(),
		HitRating: 10,
		// Earthshaker damage and speed.
		WeaponDamageMin: 175, WeaponDamageMax: 263, WeaponSpeed: 3.5,
		Ilvl: 76, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
}
