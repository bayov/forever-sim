package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The hunter's synthetic Phase 1 items, made the way forever_synthetic_items.go describes, ids
// 993000 to 993999. Two sets:
//   - Ranged Hunter (993001 to 993014) for Beast Mastery, Marksmanship and Lone Wolf. They shoot
//     from range on the same stats: Agility (2 ranged attack power and crit), ranged attack power,
//     crit and hit. The pet takes 10% of the hunter's ranged attack power and all of its crit
//     (sim/hunter/pet.go), so the same stats feed it. The two-hander is a stat stick.
//   - Survival (993101 to 993115). Survival melees under Forever, so it wants melee attack power and
//     a pair of weapons (Predator's Edge adds 50% off hand damage), and some Strength.
//
// Each piece stands in for an Era Phase 1 hunter BiS item: the Giantstalker set, Onyxia Tooth
// Pendant, Wristguards of True Flight, Cape of the Black Baron, the Devilsaur hands and legs, Quick
// Strike Ring, Blackstone Ring, Barbarous Blade (a pair of Brutality Blades for Survival) and
// Rhok'delar. The weapons keep their damage and speed and weapon type.
//
// Item levels and budgets follow the cross-class standard (orchestration/standards.md, 2026-10-12)
// so every class's sets cost the same: head, shoulders, chest, hands, legs and feet at item level
// 66, neck, back, wrist, waist and both rings at 71, a two-hander at 76, each one-hander at 71 and
// the ranged weapon at 71. An epic item at level L is worth V = 1.2 * L - 36.4 (42.8 at 66, 48.8 at
// 71, 54.8 at 76), times the slot modifier: 1 for head, chest, legs and two-handers, 0.77 for
// shoulders, hands, waist and feet, 0.5625 for neck, wrist, back and rings, 0.42 for a one-hander,
// 0.32 for a ranged weapon. The stats on an item add up as (sum (points * price)^1.5)^(2/3), which
// equals the slot's value. A point of Strength, Agility, Stamina, Intellect or Spirit costs 1,
// attack power 0.5, ranged attack power only 0.4, hit and crit rating 1 at Forever's 10 rating per
// 1% hit and 14 per 1% crit. Armor is free.
//
// So a head at item level 66 has 42.8, or 280 to spend in the 1.5 power: the Ranged Hunter head's 17
// Agility, 13 Stamina, 10 Intellect, 10 Spirit, 15 ranged attack power (6 points), 10 hit and 14
// crit rating come to 279. A ring at 71 has 0.5625 * 48.8 = 27.5 (144), the two-hander at 76 has
// 54.8 (406), the bow 0.32 * 48.8 = 15.6 (62). Hit and crit go on first, the rest is split by hand
// picked shares per slot, rounded, and trimmed back to within 1% of the slot's value (the
// generator was a script in the session scratchpad).
//
// The stat mix:
//   - Spirit is about 10% of each set's points, on 7 pieces, as for every mana user.
//   - Agility is about 30%, Stamina 20%. Ranged attack power (or attack power and Strength for
//     Survival) leans on the slots the Era items carry it in: the back, hands, legs, a ring and the
//     two-hander. Intellect is 5 to 7%, enough for the mana and for Careful Aim to be worth a look.
//   - Hit up to the cap against a level 63 boss and no further: 90 rating (9%, the ranged cap) on 9
//     pieces for the ranged set. 60 (6%) for Survival, whose Surefooted gives the other 3% of the
//     9% yellow cap.
//   - Crit: 70 rating (5%) on 5 pieces, the Era BiS's crit pieces (the Giantstalker helm, chest and
//     belt, Onyxia Tooth Pendant).
//
// The head and chest of each set form a set with the bonuses of Wildstalker Armor, Forever's own
// Molten Core tier set for the hunter (set 2101, items 280893 to 280898, client 1.60.1.70338, see
// sim/hunter/item_sets_pve.go), whose client pieces carry no stats yet. Only two pieces, so only the
// 2 piece bonus (1% attack and cast speed) is live, as the standard has it for every class.
//
// Swords favor the Human (2% crit with swords, melee only) and give no other race a racial. The bow
// carries no racial under Forever.
func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, foreverSyntheticHunterItems...)
}

var foreverSyntheticHunterItems = []*proto.UIItem{
	// Head, item level 66, in place of Giantstalker's Helmet (66, 16846).
	{
		Id: 993001, Name: "Ranged Hunter Synthetic Head (Phase 1)", Icon: "inv_helmet_05",
		SetName: "Hunter Synthetic Ranged (Phase 1)",
		Type:    proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Agility: 17, stats.Stamina: 13, stats.Intellect: 10, stats.Spirit: 10, stats.RangedAttackPower: 15, stats.Armor: 343}.ToFloatArray(),
		HitRating: 10, CritRating: 14,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Neck, item level 71, in place of Onyxia Tooth Pendant (74, 18404).
	{
		Id: 993002, Name: "Ranged Hunter Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_09",
		Type:      proto.ItemType_ItemTypeNeck,
		Stats:     stats.Stats{stats.Agility: 12, stats.Stamina: 7}.ToFloatArray(),
		HitRating: 10, CritRating: 14,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Shoulder, item level 66, in place of Giantstalker's Epaulets (66, 16848).
	{
		Id: 993003, Name: "Ranged Hunter Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_10",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Agility: 16, stats.Stamina: 11, stats.Intellect: 7, stats.RangedAttackPower: 27, stats.Armor: 317}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Back, item level 71, in place of Cape of the Black Baron (63, 13340).
	{
		Id: 993004, Name: "Ranged Hunter Synthetic Back (Phase 1)", Icon: "inv_misc_cape_18",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Agility: 14, stats.Stamina: 11, stats.Spirit: 7, stats.RangedAttackPower: 27, stats.Armor: 54}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Chest, item level 66, in place of Giantstalker's Breastplate (66, 16845).
	{
		Id: 993005, Name: "Ranged Hunter Synthetic Chest (Phase 1)", Icon: "inv_chest_chain_03",
		SetName: "Hunter Synthetic Ranged (Phase 1)",
		Type:    proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:      stats.Stats{stats.Agility: 21, stats.Stamina: 15, stats.Intellect: 11, stats.Spirit: 11, stats.Armor: 422}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Wrist, item level 71, in place of Wristguards of True Flight (71, 18812).
	{
		Id: 993006, Name: "Ranged Hunter Synthetic Wrist (Phase 1)", Icon: "inv_bracer_17",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Agility: 16, stats.Stamina: 10, stats.Intellect: 6, stats.Armor: 198}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Hands, item level 66, in place of Devilsaur Gauntlets (58, 15063).
	{
		Id: 993007, Name: "Ranged Hunter Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_10",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Agility: 17, stats.Stamina: 8, stats.RangedAttackPower: 41, stats.Armor: 264}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Waist, item level 71, in place of Giantstalker's Belt (66, 16851).
	{
		Id: 993008, Name: "Ranged Hunter Synthetic Waist (Phase 1)", Icon: "inv_belt_28",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:      stats.Stats{stats.Agility: 15, stats.Stamina: 13, stats.Intellect: 8, stats.Spirit: 13, stats.Armor: 237}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Legs, item level 66, in place of Devilsaur Leggings (60, 15062).
	{
		Id: 993009, Name: "Ranged Hunter Synthetic Legs (Phase 1)", Icon: "inv_pants_mail_03",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Agility: 21, stats.Stamina: 14, stats.Spirit: 9, stats.RangedAttackPower: 44, stats.Armor: 369}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Feet, item level 66, in place of Giantstalker's Boots (66, 16849).
	{
		Id: 993010, Name: "Ranged Hunter Synthetic Feet (Phase 1)", Icon: "inv_boots_chain_04",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Agility: 19, stats.Stamina: 12, stats.Intellect: 8, stats.Spirit: 12, stats.Armor: 290}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Finger, item level 71, in place of Quick Strike Ring (67, 18821).
	{
		Id: 993011, Name: "Ranged Hunter Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_15",
		Type:      proto.ItemType_ItemTypeFinger,
		Stats:     stats.Stats{stats.Agility: 12, stats.Stamina: 8, stats.RangedAttackPower: 32}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Finger II, item level 71, in place of Blackstone Ring (54, 17713).
	{
		Id: 993012, Name: "Ranged Hunter Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_15",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Agility: 7, stats.Stamina: 7, stats.Intellect: 7, stats.Spirit: 11}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Two-Hand, item level 76, in place of Barbarous Blade (63, 18520).
	{
		Id: 993013, Name: "Ranged Hunter Synthetic Two-Hand (Phase 1)", Icon: "inv_sword_50",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeTwoHand,
		Stats:     stats.Stats{stats.Agility: 27, stats.Stamina: 21, stats.RangedAttackPower: 67}.ToFloatArray(),
		HitRating: 10,
		// Barbarous Blade damage and speed.
		WeaponDamageMin: 138, WeaponDamageMax: 207, WeaponSpeed: 3.2,
		Ilvl: 76, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Ranged, item level 71, in place of Rhok'delar, Longbow of the Ancient Keepers (75, 18713).
	{
		Id: 993014, Name: "Ranged Hunter Synthetic Ranged (Phase 1)", Icon: "inv_weapon_bow_01",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeBow,
		Stats:     stats.Stats{stats.Agility: 8, stats.RangedAttackPower: 10}.ToFloatArray(),
		HitRating: 10,
		// Rhok'delar damage and speed.
		WeaponDamageMin: 89, WeaponDamageMax: 166, WeaponSpeed: 2.9,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},

	// Head, item level 66, in place of Giantstalker's Helmet (66, 16846).
	{
		Id: 993101, Name: "Survival Synthetic Head (Phase 1)", Icon: "inv_helmet_05",
		SetName: "Hunter Synthetic Survival (Phase 1)",
		Type:    proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Agility: 17, stats.Stamina: 13, stats.Intellect: 10, stats.Spirit: 10, stats.Strength: 6, stats.Armor: 343}.ToFloatArray(),
		HitRating: 10, CritRating: 14,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Neck, item level 71, in place of Onyxia Tooth Pendant (74, 18404).
	{
		Id: 993102, Name: "Survival Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_09",
		Type:      proto.ItemType_ItemTypeNeck,
		Stats:     stats.Stats{stats.Agility: 12, stats.Stamina: 7}.ToFloatArray(),
		HitRating: 10, CritRating: 14,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Shoulder, item level 66, in place of Giantstalker's Epaulets (66, 16848).
	{
		Id: 993103, Name: "Survival Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_10",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Agility: 19, stats.Stamina: 12, stats.Intellect: 8, stats.AttackPower: 24, stats.RangedAttackPower: 24, stats.Armor: 317}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Back, item level 71, in place of Cape of the Black Baron (63, 13340).
	{
		Id: 993104, Name: "Survival Synthetic Back (Phase 1)", Icon: "inv_misc_cape_18",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Agility: 14, stats.Stamina: 11, stats.Spirit: 7, stats.AttackPower: 22, stats.RangedAttackPower: 22, stats.Armor: 54}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Chest, item level 66, in place of Giantstalker's Breastplate (66, 16845).
	{
		Id: 993105, Name: "Survival Synthetic Chest (Phase 1)", Icon: "inv_chest_chain_03",
		SetName: "Hunter Synthetic Survival (Phase 1)",
		Type:    proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:      stats.Stats{stats.Agility: 21, stats.Stamina: 15, stats.Strength: 11, stats.Spirit: 11, stats.Armor: 422}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Wrist, item level 71, in place of Wristguards of True Flight (71, 18812).
	{
		Id: 993106, Name: "Survival Synthetic Wrist (Phase 1)", Icon: "inv_bracer_17",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Agility: 16, stats.Stamina: 10, stats.Intellect: 6, stats.Armor: 198}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Hands, item level 66, in place of Devilsaur Gauntlets (58, 15063).
	{
		Id: 993107, Name: "Survival Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_10",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Agility: 19, stats.Stamina: 9, stats.AttackPower: 37, stats.RangedAttackPower: 37, stats.Armor: 264}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Waist, item level 71, in place of Giantstalker's Belt (66, 16851).
	{
		Id: 993108, Name: "Survival Synthetic Waist (Phase 1)", Icon: "inv_belt_28",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:      stats.Stats{stats.Agility: 15, stats.Stamina: 13, stats.Strength: 8, stats.Spirit: 13, stats.Armor: 237}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Legs, item level 66, in place of Devilsaur Leggings (60, 15062).
	{
		Id: 993109, Name: "Survival Synthetic Legs (Phase 1)", Icon: "inv_pants_mail_03",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Agility: 21, stats.Stamina: 14, stats.Spirit: 9, stats.AttackPower: 35, stats.RangedAttackPower: 35, stats.Armor: 369}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Feet, item level 66, in place of Giantstalker's Boots (66, 16849).
	{
		Id: 993110, Name: "Survival Synthetic Feet (Phase 1)", Icon: "inv_boots_chain_04",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Agility: 19, stats.Stamina: 12, stats.Intellect: 8, stats.Spirit: 12, stats.Armor: 290}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Finger, item level 71, in place of Quick Strike Ring (67, 18821).
	{
		Id: 993111, Name: "Survival Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_15",
		Type:  proto.ItemType_ItemTypeFinger,
		Stats: stats.Stats{stats.Agility: 13, stats.Stamina: 8, stats.Strength: 8, stats.AttackPower: 26, stats.RangedAttackPower: 26}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Finger II, item level 71, in place of Blackstone Ring (54, 17713).
	{
		Id: 993112, Name: "Survival Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_15",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Agility: 7, stats.Stamina: 7, stats.Intellect: 7, stats.Spirit: 11}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Main Hand, item level 71, in place of Brutality Blade (70, 18832), main hand.
	{
		Id: 993113, Name: "Survival Synthetic Main Hand (Phase 1)", Icon: "inv_sword_39",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand,
		Stats:     stats.Stats{stats.Agility: 10, stats.Strength: 7, stats.Stamina: 5}.ToFloatArray(),
		HitRating: 10,
		// Brutality Blade damage and speed.
		WeaponDamageMin: 90, WeaponDamageMax: 168, WeaponSpeed: 2.5,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Off Hand, item level 71, in place of Brutality Blade (70, 18832), off hand.
	{
		Id: 993114, Name: "Survival Synthetic Off Hand (Phase 1)", Icon: "inv_sword_39",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand,
		Stats:     stats.Stats{stats.Agility: 10, stats.Strength: 7, stats.Stamina: 5}.ToFloatArray(),
		HitRating: 10,
		// Brutality Blade damage and speed.
		WeaponDamageMin: 90, WeaponDamageMax: 168, WeaponSpeed: 2.5,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Ranged, item level 71, in place of Rhok'delar, Longbow of the Ancient Keepers (75, 18713).
	{
		Id: 993115, Name: "Survival Synthetic Ranged (Phase 1)", Icon: "inv_weapon_bow_01",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeBow,
		Stats: stats.Stats{stats.Agility: 13, stats.Stamina: 6}.ToFloatArray(),
		// Rhok'delar damage and speed.
		WeaponDamageMin: 89, WeaponDamageMax: 166, WeaponSpeed: 2.9,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
}
