package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The rogue's synthetic Phase 1 items, made the same way as the enhancement shaman's (see the comment on
// ForeverSyntheticItems) and to the cross-class standard (orchestration/standards.md on the coordination
// branch): each piece takes the place of an Era Phase 1 BiS item a rogue wears (the Nightslayer set, Onyxia
// Tooth Pendant, Cape of the Black Baron, the Devilsaur set, Quick Strike Ring, Don Julio's Band, Striker's
// Mark and the Molten Core and Onyxia weapons) at the Phase 1 epic budget of its slot. The item levels are
// the shared table: 66 for head, shoulders, chest, hands, legs and feet, 71 for neck, back, wrist, waist and
// both rings, 71 for each one-hand weapon and the ranged weapon. The gear sets leave the trinket slots empty,
// as every class's synthetic preset does.
//
// The budget is an epic item value of V = 1.2 * item level - 36.4 (42.8 at 66, 48.8 at 71), times the slot
// modifier: 1 for head, chest and legs (42.8), 0.77 for shoulders, hands and feet (33.0 at 66) and the waist
// (37.6 at 71), 0.5625 for neck, back, wrist and fingers (27.5 at 71), 0.42 for a one-hand weapon (20.5 at
// 71) and 0.32 for ranged (15.6 at 71). The twelve armor slots come to 402.1. Each piece is spent linearly
// (standards 1a): its weighted stat total, with Strength, Agility and Stamina at 1, attack power at 0.5 (and
// counted once with ranged attack power) and a point of hit or crit rating at 1 (10 hit and 14 crit rating
// are 1% under Forever), is 1.73 times its budget, within 1%. The comment above each piece has both. A whole
// set comes to 793.5. Rogues use energy, so there is no Spirit or Intellect.
//
// Two sets, one for swords (Sinister Strike) and one for daggers (Backstab, Mutilate and Subtlety, whose stat
// weights agree within 7% on every stat but hit). Each carries the fair stat mix of the cross-class standard
// (standards 5a, 2026-10-13): a realistic spread rather than an all-in on Agility, at the proportions the
// fairness auditor took from Season of Discovery level 60 rogue BiS (orchestration/gear_targets.md, rows
// Rogue swords and Rogue daggers). Every piece but the five that carry the hit has the same mix, so the shares
// hold for the whole set. Shares are linear: a stat's weight times its amount over the set's weighted total,
// enchants left out and attack power counted once. No stat is over 45% of a set's non-Stamina total
// (swords Agility 44.9%).
//
//	Set            Stamina  Strength  Agility  Attack power  Hit         Crit
//	Combat Swords  143      155       292      89            50 (5%)     109 (7.8%)
//	  share (L)    18.0     19.5      36.8     5.6           6.3         13.7
//	  target       18       18.9      36.9     5.6           7.2         13.5
//	Daggers        145      145       289      113           50 (5%)     108 (7.7%)
//	  share (L)    18.3     18.3      36.4     7.1           6.3         13.6
//	  target       18       18.0      35.9     7.4           7.3         13.3
//
// Before the rebudget the sets leaned on Stamina (23 to 24%) and attack power (13 to 16%), the daggers
// carried almost no Strength (2.9%) and the swords little Agility (28.5%).
//
// Both carry 5% hit on the gear: with the set's 2 piece bonus (1%) and Precision 3/3, which every searched
// build takes, that is the 9% a Sinister Strike or Backstab needs against a level 63 boss and no more.
//
// The weapons keep the damage and speed of the Era weapons they replace, swords for Sinister Strike
// (Vis'kag the Bloodletter and Brutality Blade) and daggers for the rest (Perdition's Blade and Core Hound
// Tooth), with the weapon slot's budget in stats on top and no proc. The ranged slot is a bow with Striker's
// Mark's damage.
//
// The hands and legs of each set form a set with the bonuses of Grimstitch Armor, Forever's crafted level 60
// rogue set (see sim/rogue/items_sets_pve.go), so only its 2 piece bonus (1% hit) is live, as for every
// class. The generator lived in the session scratchpad (synth.py and lin.py), the numbers below are its output.
func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, rogueSyntheticItems...)
}

var rogueSyntheticItems = []*proto.UIItem{
	// Head, item level 66, in place of Nightslayer Cover (66, 16821). Budget 42.8, linear total 74.0.
	{
		Id: 994001, Name: "Combat Swords Synthetic Head (Phase 1)", Icon: "inv_helmet_41",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Strength: 13, stats.Agility: 25, stats.Stamina: 13, stats.AttackPower: 8, stats.Armor: 163, stats.RangedAttackPower: 8}.ToFloatArray(),
		HitRating: 10, CritRating: 9,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Neck, item level 71, in place of Onyxia Tooth Pendant (65, 18404). Budget 27.5, linear total 47.5.
	{
		Id: 994002, Name: "Combat Swords Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_07",
		Type:      proto.ItemType_ItemTypeNeck,
		Stats:     stats.Stats{stats.Strength: 8, stats.Agility: 14, stats.Stamina: 7, stats.AttackPower: 5, stats.RangedAttackPower: 5}.ToFloatArray(),
		HitRating: 10, CritRating: 6,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Shoulder, item level 66, in place of Nightslayer Shoulder Pads (66, 16823). Budget 33.0, linear total 57.0.
	{
		Id: 994003, Name: "Combat Swords Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_23",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Strength: 12, stats.Agility: 23, stats.Stamina: 11, stats.AttackPower: 6, stats.Armor: 150, stats.RangedAttackPower: 6}.ToFloatArray(),
		CritRating: 8,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Back, item level 71, in place of Cape of the Black Baron (63, 13340). Budget 27.5, linear total 47.5.
	{
		Id: 994004, Name: "Combat Swords Synthetic Back (Phase 1)", Icon: "inv_misc_cape_18",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Strength: 10, stats.Agility: 19, stats.Stamina: 9, stats.AttackPower: 5, stats.Armor: 54, stats.RangedAttackPower: 5}.ToFloatArray(),
		CritRating: 7,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Chest, item level 66, in place of Nightslayer Chestpiece (66, 16820). Budget 42.8, linear total 74.0.
	{
		Id: 994005, Name: "Combat Swords Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_07",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Strength: 16, stats.Agility: 28, stats.Stamina: 14, stats.AttackPower: 10, stats.Armor: 200, stats.RangedAttackPower: 10}.ToFloatArray(),
		CritRating: 11,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Wrist, item level 71, in place of Nightslayer Bracelets (66, 16825). Budget 27.5, linear total 47.5.
	{
		Id: 994006, Name: "Combat Swords Synthetic Wrist (Phase 1)", Icon: "inv_bracer_02",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Strength: 10, stats.Agility: 19, stats.Stamina: 9, stats.AttackPower: 5, stats.Armor: 88, stats.RangedAttackPower: 5}.ToFloatArray(),
		CritRating: 7,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Hands, item level 66, in place of Devilsaur Gauntlets (58, 15063). Budget 33.0, linear total 57.0.
	{
		Id: 994007, Name: "Combat Swords Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_21",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Strength: 9, stats.Agility: 19, stats.Stamina: 9, stats.AttackPower: 6, stats.Armor: 125, stats.RangedAttackPower: 6}.ToFloatArray(),
		HitRating: 10, CritRating: 7,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
		SetName:        "Combat Swords Synthetic (Phase 1)",
	},
	// Waist, item level 71, in place of Nightslayer Belt (66, 16827). Budget 37.6, linear total 65.0.
	{
		Id: 994008, Name: "Combat Swords Synthetic Waist (Phase 1)", Icon: "inv_belt_23",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Strength: 11, stats.Agility: 22, stats.Stamina: 11, stats.AttackPower: 6, stats.Armor: 113, stats.RangedAttackPower: 6}.ToFloatArray(),
		HitRating: 10, CritRating: 8,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Legs, item level 66, in place of Devilsaur Leggings (60, 15062). Budget 42.8, linear total 74.0.
	{
		Id: 994009, Name: "Combat Swords Synthetic Legs (Phase 1)", Icon: "inv_pants_06",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Strength: 16, stats.Agility: 28, stats.Stamina: 14, stats.AttackPower: 10, stats.Armor: 175, stats.RangedAttackPower: 10}.ToFloatArray(),
		CritRating: 11,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
		SetName:        "Combat Swords Synthetic (Phase 1)",
	},
	// Feet, item level 66, in place of Nightslayer Boots (66, 16824). Budget 33.0, linear total 57.0.
	{
		Id: 994010, Name: "Combat Swords Synthetic Feet (Phase 1)", Icon: "inv_boots_08",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Strength: 13, stats.Agility: 22, stats.Stamina: 11, stats.AttackPower: 6, stats.Armor: 138, stats.RangedAttackPower: 6}.ToFloatArray(),
		CritRating: 8,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Finger, item level 71, in place of Quick Strike Ring (67, 18821). Budget 27.5, linear total 47.5.
	{
		Id: 994011, Name: "Combat Swords Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_15",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Strength: 10, stats.Agility: 19, stats.Stamina: 9, stats.AttackPower: 5, stats.RangedAttackPower: 5}.ToFloatArray(),
		CritRating: 7,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Finger II, item level 71, in place of Don Julio's Band (65, 19325). Budget 27.5, linear total 47.5.
	{
		Id: 994012, Name: "Combat Swords Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_03",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Strength: 10, stats.Agility: 19, stats.Stamina: 9, stats.AttackPower: 5, stats.RangedAttackPower: 5}.ToFloatArray(),
		CritRating: 7,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Main Hand, item level 71, in place of Vis'kag the Bloodletter (74, 17075). Budget 20.5, linear total 35.5.
	{
		Id: 994013, Name: "Combat Swords Synthetic Main Hand (Phase 1)", Icon: "inv_sword_39",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand,
		Stats:     stats.Stats{stats.Strength: 5, stats.Agility: 10, stats.Stamina: 5, stats.AttackPower: 3, stats.RangedAttackPower: 3}.ToFloatArray(),
		HitRating: 10, CritRating: 4,
		// Vis'kag the Bloodletter damage and speed.
		WeaponDamageMin: 100, WeaponDamageMax: 187, WeaponSpeed: 2.6,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Off Hand, item level 71, in place of Brutality Blade (70, 18832). Budget 20.5, linear total 35.5.
	{
		Id: 994014, Name: "Combat Swords Synthetic Off Hand (Phase 1)", Icon: "inv_sword_40",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand,
		Stats:      stats.Stats{stats.Strength: 7, stats.Agility: 14, stats.Stamina: 7, stats.AttackPower: 5, stats.RangedAttackPower: 5}.ToFloatArray(),
		CritRating: 5,
		// Brutality Blade damage and speed.
		WeaponDamageMin: 90, WeaponDamageMax: 168, WeaponSpeed: 2.5,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Ranged, item level 71, in place of Striker's Mark (69, 17069). Budget 15.6, linear total 27.0.
	{
		Id: 994015, Name: "Combat Swords Synthetic Ranged (Phase 1)", Icon: "inv_weapon_bow_08",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeBow,
		Stats:      stats.Stats{stats.Strength: 5, stats.Agility: 11, stats.Stamina: 5, stats.AttackPower: 4, stats.RangedAttackPower: 4}.ToFloatArray(),
		CritRating: 4,
		// Striker's Mark damage and speed.
		WeaponDamageMin: 69, WeaponDamageMax: 129, WeaponSpeed: 2.5,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},

	// Head, item level 66, in place of Nightslayer Cover (66, 16821). Budget 42.8, linear total 74.0.
	{
		Id: 994101, Name: "Daggers Synthetic Head (Phase 1)", Icon: "inv_helmet_41",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Strength: 13, stats.Agility: 24, stats.Stamina: 13, stats.AttackPower: 10, stats.Armor: 163, stats.RangedAttackPower: 10}.ToFloatArray(),
		HitRating: 10, CritRating: 9,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Neck, item level 71, in place of Onyxia Tooth Pendant (65, 18404). Budget 27.5, linear total 47.5.
	{
		Id: 994102, Name: "Daggers Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_07",
		Type:       proto.ItemType_ItemTypeNeck,
		Stats:      stats.Stats{stats.Strength: 9, stats.Agility: 19, stats.Stamina: 9, stats.AttackPower: 7, stats.RangedAttackPower: 7}.ToFloatArray(),
		CritRating: 7,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Shoulder, item level 66, in place of Nightslayer Shoulder Pads (66, 16823). Budget 33.0, linear total 57.0.
	{
		Id: 994103, Name: "Daggers Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_23",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Strength: 9, stats.Agility: 18, stats.Stamina: 9, stats.AttackPower: 8, stats.Armor: 150, stats.RangedAttackPower: 8}.ToFloatArray(),
		HitRating: 10, CritRating: 7,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Back, item level 71, in place of Cape of the Black Baron (63, 13340). Budget 27.5, linear total 47.5.
	{
		Id: 994104, Name: "Daggers Synthetic Back (Phase 1)", Icon: "inv_misc_cape_18",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Strength: 9, stats.Agility: 19, stats.Stamina: 9, stats.AttackPower: 7, stats.Armor: 54, stats.RangedAttackPower: 7}.ToFloatArray(),
		CritRating: 7,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Chest, item level 66, in place of Nightslayer Chestpiece (66, 16820). Budget 42.8, linear total 74.0.
	{
		Id: 994105, Name: "Daggers Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_07",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Strength: 15, stats.Agility: 28, stats.Stamina: 15, stats.AttackPower: 10, stats.Armor: 200, stats.RangedAttackPower: 10}.ToFloatArray(),
		CritRating: 11,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Wrist, item level 71, in place of Nightslayer Bracelets (66, 16825). Budget 27.5, linear total 47.5.
	{
		Id: 994106, Name: "Daggers Synthetic Wrist (Phase 1)", Icon: "inv_bracer_02",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Strength: 7, stats.Agility: 15, stats.Stamina: 7, stats.AttackPower: 7, stats.Armor: 88, stats.RangedAttackPower: 7}.ToFloatArray(),
		HitRating: 10, CritRating: 5,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Hands, item level 66, in place of Devilsaur Gauntlets (58, 15063). Budget 33.0, linear total 57.0.
	{
		Id: 994107, Name: "Daggers Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_21",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Strength: 11, stats.Agility: 22, stats.Stamina: 11, stats.AttackPower: 10, stats.Armor: 125, stats.RangedAttackPower: 10}.ToFloatArray(),
		CritRating: 8,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
		SetName:        "Daggers Synthetic (Phase 1)",
	},
	// Waist, item level 71, in place of Nightslayer Belt (66, 16827). Budget 37.6, linear total 65.0.
	{
		Id: 994108, Name: "Daggers Synthetic Waist (Phase 1)", Icon: "inv_belt_23",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Strength: 11, stats.Agility: 21, stats.Stamina: 11, stats.AttackPower: 8, stats.Armor: 113, stats.RangedAttackPower: 8}.ToFloatArray(),
		HitRating: 10, CritRating: 8,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Legs, item level 66, in place of Devilsaur Leggings (60, 15062). Budget 42.8, linear total 74.0.
	{
		Id: 994109, Name: "Daggers Synthetic Legs (Phase 1)", Icon: "inv_pants_06",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Strength: 15, stats.Agility: 28, stats.Stamina: 15, stats.AttackPower: 10, stats.Armor: 175, stats.RangedAttackPower: 10}.ToFloatArray(),
		CritRating: 11,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
		SetName:        "Daggers Synthetic (Phase 1)",
	},
	// Feet, item level 66, in place of Nightslayer Boots (66, 16824). Budget 33.0, linear total 57.0.
	{
		Id: 994110, Name: "Daggers Synthetic Feet (Phase 1)", Icon: "inv_boots_08",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Strength: 11, stats.Agility: 22, stats.Stamina: 11, stats.AttackPower: 10, stats.Armor: 138, stats.RangedAttackPower: 10}.ToFloatArray(),
		CritRating: 8,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Finger, item level 71, in place of Quick Strike Ring (67, 18821). Budget 27.5, linear total 47.5.
	{
		Id: 994111, Name: "Daggers Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_15",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Strength: 9, stats.Agility: 19, stats.Stamina: 9, stats.AttackPower: 7, stats.RangedAttackPower: 7}.ToFloatArray(),
		CritRating: 7,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Finger II, item level 71, in place of Don Julio's Band (65, 19325). Budget 27.5, linear total 47.5.
	{
		Id: 994112, Name: "Daggers Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_03",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Strength: 9, stats.Agility: 19, stats.Stamina: 9, stats.AttackPower: 7, stats.RangedAttackPower: 7}.ToFloatArray(),
		CritRating: 7,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Main Hand, item level 71, in place of Perdition's Blade (77, 18816). Budget 20.5, linear total 35.5.
	{
		Id: 994113, Name: "Daggers Synthetic Main Hand (Phase 1)", Icon: "inv_weapon_shortblade_16",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeDagger, HandType: proto.HandType_HandTypeOneHand,
		Stats:     stats.Stats{stats.Strength: 5, stats.Agility: 10, stats.Stamina: 5, stats.AttackPower: 3, stats.RangedAttackPower: 3}.ToFloatArray(),
		HitRating: 10, CritRating: 4,
		// Perdition's Blade damage and speed.
		WeaponDamageMin: 73, WeaponDamageMax: 137, WeaponSpeed: 1.8,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Off Hand, item level 71, in place of Core Hound Tooth (70, 18805). Budget 20.5, linear total 35.5.
	{
		Id: 994114, Name: "Daggers Synthetic Off Hand (Phase 1)", Icon: "inv_weapon_shortblade_11",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeDagger, HandType: proto.HandType_HandTypeOneHand,
		Stats:      stats.Stats{stats.Strength: 7, stats.Agility: 14, stats.Stamina: 7, stats.AttackPower: 5, stats.RangedAttackPower: 5}.ToFloatArray(),
		CritRating: 5,
		// Core Hound Tooth damage and speed.
		WeaponDamageMin: 57, WeaponDamageMax: 107, WeaponSpeed: 1.6,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Ranged, item level 71, in place of Striker's Mark (69, 17069). Budget 15.6, linear total 27.0.
	{
		Id: 994115, Name: "Daggers Synthetic Ranged (Phase 1)", Icon: "inv_weapon_bow_08",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeBow,
		Stats:      stats.Stats{stats.Strength: 5, stats.Agility: 11, stats.Stamina: 5, stats.AttackPower: 4, stats.RangedAttackPower: 4}.ToFloatArray(),
		CritRating: 4,
		// Striker's Mark damage and speed.
		WeaponDamageMin: 69, WeaponDamageMax: 129, WeaponSpeed: 2.5,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
}
