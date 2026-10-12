package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The rogue's synthetic Phase 1 items, made the same way as the enhancement shaman's (see the comment on
// ForeverSyntheticItems): each piece takes the place of an Era Phase 1 BiS item a rogue wears (the
// Nightslayer set, Onyxia Tooth Pendant, Cape of the Black Baron, the Devilsaur set, Quick Strike Ring,
// Don Julio's Band, Striker's Mark and the Molten Core and Onyxia weapons) at the Phase 1 epic budget of its
// slot, at item level 66 or the replaced item's level where that is higher. The gear sets leave the
// trinket slots empty, as every class's synthetic preset does.
//
// The budget is the one the shaman's items spend: an epic item value of 1.2 * item level - 36.4 (42.8 at 66,
// 52.4 at 74, 56.0 at 77), times the slot modifier (1 for head, chest and legs, 0.77 for shoulders, hands,
// waist and feet, 0.5625 for neck, back, wrist and fingers, 0.42 for a one-hand weapon, 0.32 for ranged).
// A piece is worth the 1.5 norm of its stats, (sum of (weight * amount)^1.5)^(2/3), with Strength, Agility
// and Stamina at 1, attack power at 0.5, and a point of hit or crit rating at 1 (10 hit and 14 crit rating
// are 1% under Forever). Each piece lands within half a point of its budget, the comment above it has both.
// Rogues use energy, so there is no Spirit or Intellect.
//
// Two sets, because the specs value the stats differently. Stat weights on the Era Phase 1 sets (rotopt
// -weights, 20000 iterations, 180 sec, in attack power): Sinister Strike with swords puts Agility at 1.92,
// 1% crit at 15 and 1% hit at 31, while Backstab and Mutilate with daggers put Agility at 2.74 and 2.79,
// crit at 24.6 and 26.2 and hit at 20.7 and 26.2. The dagger specs are within 7% of each other on every
// stat but hit, so Backstab, Mutilate and Subtlety share the dagger set.
//   - Combat Swords: 187 Agility, 77 Strength, 176 attack power, 147 Stamina, 80 hit and 98 crit rating
//     (7%). Attack power and Strength take more of the budget than on the dagger set.
//   - Daggers: 218 Agility, 20 Strength, 145 attack power, 152 Stamina, 80 hit and 126 crit rating (9%).
//
// Both carry 8% hit on the gear, and the set's 2 piece bonus adds 1%, the 9% the Era Phase 1 BiS carries
// (Onyxia Tooth Pendant, the Devilsaur set, Blackstone Ring, Striker's Mark). Not every piece has every stat:
// the wrists and feet carry no attack power, the swords' wrist and feet no hit or crit.
//
// The weapons keep the damage and speed of the Era weapons they replace, swords for Sinister Strike
// (Vis'kag the Bloodletter and Brutality Blade) and daggers for the rest (Perdition's Blade and Core Hound
// Tooth), with the weapon slot's budget in stats on top and no proc. The ranged slot is a bow with Striker's
// Mark's damage.
//
// The head, shoulders, chest, hands, legs and feet of each set form a set with the bonuses of Grimstitch
// Armor, Forever's crafted level 60 rogue set (see sim/rogue/items_sets_pve.go). The generator lived in the
// session scratchpad (synth.py), the numbers below are its output.
func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, rogueSyntheticItems...)
}

var rogueSyntheticItems = []*proto.UIItem{
	// Head, item level 66, in place of Nightslayer Cover (66, 16821). Budget 42.8, value 43.0.
	{
		Id: 994001, Name: "Combat Swords Synthetic Head (Phase 1)", Icon: "inv_helmet_41",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Strength: 10, stats.Agility: 21, stats.Stamina: 17, stats.Armor: 163}.ToFloatArray(),
		HitRating: 10, CritRating: 14,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
		SetName:        "Combat Swords Synthetic (Phase 1)",
	},
	// Neck, item level 66, in place of Onyxia Tooth Pendant (65, 18404). Budget 24.1, value 24.2.
	{
		Id: 994002, Name: "Combat Swords Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_07",
		Type:      proto.ItemType_ItemTypeNeck,
		Stats:     stats.Stats{stats.Agility: 9, stats.Stamina: 7, stats.AttackPower: 24, stats.RangedAttackPower: 24}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Shoulder, item level 66, in place of Nightslayer Shoulder Pads (66, 16823). Budget 33.0, value 32.9.
	{
		Id: 994003, Name: "Combat Swords Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_23",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Agility: 17, stats.Stamina: 12, stats.AttackPower: 17, stats.Armor: 150, stats.RangedAttackPower: 17}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
		SetName:        "Combat Swords Synthetic (Phase 1)",
	},
	// Back, item level 66, in place of Cape of the Black Baron (63, 13340). Budget 24.1, value 24.2.
	{
		Id: 994004, Name: "Combat Swords Synthetic Back (Phase 1)", Icon: "inv_misc_cape_18",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Agility: 18, stats.Stamina: 10, stats.AttackPower: 10, stats.Armor: 48, stats.RangedAttackPower: 10}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Chest, item level 66, in place of Nightslayer Chestpiece (66, 16820). Budget 42.8, value 42.7.
	{
		Id: 994005, Name: "Combat Swords Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_07",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Strength: 13, stats.Agility: 22, stats.Stamina: 18, stats.Armor: 200}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
		SetName:        "Combat Swords Synthetic (Phase 1)",
	},
	// Wrist, item level 66, in place of Nightslayer Bracelets (66, 16825). Budget 24.1, value 24.4.
	{
		Id: 994006, Name: "Combat Swords Synthetic Wrist (Phase 1)", Icon: "inv_bracer_02",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Strength: 13, stats.Agility: 13, stats.Stamina: 9, stats.Armor: 88}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Hands, item level 66, in place of Devilsaur Gauntlets (58, 15063). Budget 33.0, value 33.1.
	{
		Id: 994007, Name: "Combat Swords Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_21",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Strength: 14, stats.Agility: 10, stats.Stamina: 8, stats.Armor: 125}.ToFloatArray(),
		HitRating: 10, CritRating: 14,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
		SetName:        "Combat Swords Synthetic (Phase 1)",
	},
	// Waist, item level 66, in place of Nightslayer Belt (66, 16827). Budget 33.0, value 32.7.
	{
		Id: 994008, Name: "Combat Swords Synthetic Waist (Phase 1)", Icon: "inv_belt_23",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Agility: 15, stats.Stamina: 11, stats.AttackPower: 31, stats.Armor: 113, stats.RangedAttackPower: 31}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Legs, item level 66, in place of Devilsaur Leggings (60, 15062). Budget 42.8, value 42.9.
	{
		Id: 994009, Name: "Combat Swords Synthetic Legs (Phase 1)", Icon: "inv_pants_06",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Agility: 15, stats.Stamina: 14, stats.AttackPower: 39, stats.Armor: 175, stats.RangedAttackPower: 39}.ToFloatArray(),
		HitRating: 10, CritRating: 14,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
		SetName:        "Combat Swords Synthetic (Phase 1)",
	},
	// Feet, item level 66, in place of Nightslayer Boots (66, 16824). Budget 33.0, value 32.5.
	{
		Id: 994010, Name: "Combat Swords Synthetic Feet (Phase 1)", Icon: "inv_boots_08",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Strength: 10, stats.Agility: 20, stats.Stamina: 16, stats.Armor: 138}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
		SetName:        "Combat Swords Synthetic (Phase 1)",
	},
	// Finger, item level 67, in place of Quick Strike Ring (67, 18821). Budget 24.7, value 24.8.
	{
		Id: 994011, Name: "Combat Swords Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_15",
		Type:      proto.ItemType_ItemTypeFinger,
		Stats:     stats.Stats{stats.Agility: 8, stats.Stamina: 8, stats.AttackPower: 26, stats.RangedAttackPower: 26}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      67, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Finger II, item level 66, in place of Don Julio's Band (65, 19325). Budget 24.1, value 24.0.
	{
		Id: 994012, Name: "Combat Swords Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_03",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Strength: 10, stats.Stamina: 7, stats.AttackPower: 12, stats.RangedAttackPower: 12}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Main Hand, item level 74, in place of Vis'kag the Bloodletter (74, 17075). Budget 22.0, value 22.2.
	{
		Id: 994013, Name: "Combat Swords Synthetic Main Hand (Phase 1)", Icon: "inv_sword_39",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand,
		Stats:     stats.Stats{stats.Agility: 12, stats.Stamina: 10}.ToFloatArray(),
		HitRating: 10,
		// Vis'kag the Bloodletter damage and speed.
		WeaponDamageMin: 100, WeaponDamageMax: 187, WeaponSpeed: 2.6,
		Ilvl: 74, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Off Hand, item level 70, in place of Brutality Blade (70, 18832). Budget 20.0, value 20.0.
	{
		Id: 994014, Name: "Combat Swords Synthetic Off Hand (Phase 1)", Icon: "inv_sword_40",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand,
		Stats:      stats.Stats{stats.Strength: 7, stats.Agility: 7}.ToFloatArray(),
		CritRating: 14,
		// Brutality Blade damage and speed.
		WeaponDamageMin: 90, WeaponDamageMax: 168, WeaponSpeed: 2.5,
		Ilvl: 70, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Ranged, item level 69, in place of Striker's Mark (69, 17069). Budget 14.8, value 14.7.
	{
		Id: 994015, Name: "Combat Swords Synthetic Ranged (Phase 1)", Icon: "inv_weapon_bow_08",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeBow,
		Stats:     stats.Stats{stats.AttackPower: 17, stats.RangedAttackPower: 17}.ToFloatArray(),
		HitRating: 10,
		// Striker's Mark damage and speed.
		WeaponDamageMin: 69, WeaponDamageMax: 129, WeaponSpeed: 2.5,
		Ilvl: 69, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},

	// Head, item level 66, in place of Nightslayer Cover (66, 16821). Budget 42.8, value 42.8.
	{
		Id: 994101, Name: "Daggers Synthetic Head (Phase 1)", Icon: "inv_helmet_41",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Agility: 24, stats.Stamina: 16, stats.AttackPower: 12, stats.Armor: 163, stats.RangedAttackPower: 12}.ToFloatArray(),
		HitRating: 10, CritRating: 14,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
		SetName:        "Daggers Synthetic (Phase 1)",
	},
	// Neck, item level 66, in place of Onyxia Tooth Pendant (65, 18404). Budget 24.1, value 23.7.
	{
		Id: 994102, Name: "Daggers Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_07",
		Type:       proto.ItemType_ItemTypeNeck,
		Stats:      stats.Stats{stats.Agility: 10, stats.Stamina: 6, stats.AttackPower: 13, stats.RangedAttackPower: 13}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Shoulder, item level 66, in place of Nightslayer Shoulder Pads (66, 16823). Budget 33.0, value 33.1.
	{
		Id: 994103, Name: "Daggers Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_23",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Agility: 17, stats.Stamina: 11, stats.Armor: 150}.ToFloatArray(),
		HitRating: 10, CritRating: 14,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
		SetName:        "Daggers Synthetic (Phase 1)",
	},
	// Back, item level 66, in place of Cape of the Black Baron (63, 13340). Budget 24.1, value 23.7.
	{
		Id: 994104, Name: "Daggers Synthetic Back (Phase 1)", Icon: "inv_misc_cape_18",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Agility: 17, stats.Stamina: 9, stats.AttackPower: 14, stats.Armor: 48, stats.RangedAttackPower: 14}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Chest, item level 66, in place of Nightslayer Chestpiece (66, 16820). Budget 42.8, value 43.0.
	{
		Id: 994105, Name: "Daggers Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_07",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Strength: 9, stats.Agility: 26, stats.Stamina: 17, stats.Armor: 200}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
		SetName:        "Daggers Synthetic (Phase 1)",
	},
	// Wrist, item level 66, in place of Nightslayer Bracelets (66, 16825). Budget 24.1, value 23.9.
	{
		Id: 994106, Name: "Daggers Synthetic Wrist (Phase 1)", Icon: "inv_bracer_02",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Agility: 15, stats.Stamina: 9, stats.Armor: 88}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Hands, item level 66, in place of Devilsaur Gauntlets (58, 15063). Budget 33.0, value 33.0.
	{
		Id: 994107, Name: "Daggers Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_21",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Strength: 11, stats.Agility: 16, stats.Stamina: 11, stats.Armor: 125}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
		SetName:        "Daggers Synthetic (Phase 1)",
	},
	// Waist, item level 66, in place of Nightslayer Belt (66, 16827). Budget 33.0, value 32.8.
	{
		Id: 994108, Name: "Daggers Synthetic Waist (Phase 1)", Icon: "inv_belt_23",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Agility: 19, stats.Stamina: 12, stats.AttackPower: 20, stats.Armor: 113, stats.RangedAttackPower: 20}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Legs, item level 66, in place of Devilsaur Leggings (60, 15062). Budget 42.8, value 42.8.
	{
		Id: 994109, Name: "Daggers Synthetic Legs (Phase 1)", Icon: "inv_pants_06",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Agility: 19, stats.Stamina: 14, stats.AttackPower: 31, stats.Armor: 175, stats.RangedAttackPower: 31}.ToFloatArray(),
		HitRating: 10, CritRating: 14,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
		SetName:        "Daggers Synthetic (Phase 1)",
	},
	// Feet, item level 66, in place of Nightslayer Boots (66, 16824). Budget 33.0, value 33.4.
	{
		Id: 994110, Name: "Daggers Synthetic Feet (Phase 1)", Icon: "inv_boots_08",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Agility: 19, stats.Stamina: 15, stats.Armor: 138}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
		SetName:        "Daggers Synthetic (Phase 1)",
	},
	// Finger, item level 67, in place of Quick Strike Ring (67, 18821). Budget 24.7, value 25.0.
	{
		Id: 994111, Name: "Daggers Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_15",
		Type:      proto.ItemType_ItemTypeFinger,
		Stats:     stats.Stats{stats.Agility: 10, stats.Stamina: 8, stats.AttackPower: 23, stats.RangedAttackPower: 23}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      67, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Finger II, item level 66, in place of Don Julio's Band (65, 19325). Budget 24.1, value 24.5.
	{
		Id: 994112, Name: "Daggers Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_03",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Agility: 12, stats.Stamina: 9}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Main Hand, item level 77, in place of Perdition's Blade (77, 18816). Budget 23.5, value 23.1.
	{
		Id: 994113, Name: "Daggers Synthetic Main Hand (Phase 1)", Icon: "inv_weapon_shortblade_16",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeDagger, HandType: proto.HandType_HandTypeOneHand,
		Stats:     stats.Stats{stats.Agility: 14, stats.Stamina: 9}.ToFloatArray(),
		HitRating: 10,
		// Perdition's Blade damage and speed.
		WeaponDamageMin: 73, WeaponDamageMax: 137, WeaponSpeed: 1.8,
		Ilvl: 77, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Off Hand, item level 70, in place of Core Hound Tooth (70, 18805). Budget 20.0, value 19.7.
	{
		Id: 994114, Name: "Daggers Synthetic Off Hand (Phase 1)", Icon: "inv_weapon_shortblade_11",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeDagger, HandType: proto.HandType_HandTypeOneHand,
		Stats:      stats.Stats{stats.Stamina: 6, stats.AttackPower: 15, stats.RangedAttackPower: 15}.ToFloatArray(),
		CritRating: 14,
		// Core Hound Tooth damage and speed.
		WeaponDamageMin: 57, WeaponDamageMax: 107, WeaponSpeed: 1.6,
		Ilvl: 70, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Ranged, item level 69, in place of Striker's Mark (69, 17069). Budget 14.8, value 14.7.
	{
		Id: 994115, Name: "Daggers Synthetic Ranged (Phase 1)", Icon: "inv_weapon_bow_08",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeBow,
		Stats:     stats.Stats{stats.AttackPower: 17, stats.RangedAttackPower: 17}.ToFloatArray(),
		HitRating: 10,
		// Striker's Mark damage and speed.
		WeaponDamageMin: 69, WeaponDamageMax: 129, WeaponSpeed: 2.5,
		Ilvl: 69, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
}
