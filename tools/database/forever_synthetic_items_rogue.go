package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The rogue's synthetic Phase 1 items, built to the cross-class standard (orchestration/standards.md on the
// coordination branch, items 1b, 2, 4, 5a, 6 and 8a, 2026-10-13). Each piece takes the place of an Era Phase 1
// BiS item a rogue wears (the Nightslayer set, Onyxia Tooth Pendant, Cape of the Black Baron, the Devilsaur
// set, Quick Strike Ring, Don Julio's Band, Striker's Mark and the Molten Core and Onyxia weapons). Item
// levels: 66 for head, shoulders, chest, hands, legs and feet, 71 for neck, back, wrist, waist and both rings,
// 71 for the weapons, the ranged weapon and the two trinkets.
//
// Budget (1b, physical melee row). Weights: Strength, Agility and Stamina 1, attack power 0.5 (counted once
// with the same ranged attack power), hit and crit rating 1 (10 hit and 14 crit rating are 1% under
// Forever). Each armor piece's weighted linear total is slot mod x L(ilvl) x f within 1%, with L = 1.125 x
// ilvl + 2.0, slot mods 1 / 0.75 / 0.5625 and f 0.950 / 0.959 / 0.911: 72.4 for head, chest and legs, 54.8
// for shoulders, hands and feet, 58.9 for the waist and 42.0 for neck, back, wrist and rings. Its value on a
// 1.7-norm, (sum of (weight x amount)^1.7)^(1/1.7), is within 6% of slot mod x (0.661 x ilvl + 2.92), which
// a physical piece meets with three stat lines, as each piece here has. The comment above each piece has
// both. Each one-hander is 0.39 x L(71) = 31.9 and the bow 0.24 x L(71) = 19.7. A set comes to 650.5 of
// armor, 63.9 of weapons and 19.5 of ranged, 734.0 in all.
//
// Two sets, one for swords (Sinister Strike) and one for daggers (Backstab, Mutilate and Subtlety). Each
// carries the fair stat mix of standards 5a at the 1b weights, the proportions the fairness auditor took
// from Season of Discovery level 60 rogue BiS (orchestration/gear_targets.md). Hit is held at the cap (item 4):
// 50 rating, 5%, which with the set's 2 piece bonus (1%) and Precision 3/3 is the 9% a Sinister Strike or
// Backstab needs against a level 63 boss. The share SoD gives hit (14%) goes to the other stats in their
// proportions, each kept within 3 points of its target. Shares are linear, trinkets and enchants left out:
//
//	Set            Stamina  Strength  Agility  Attack power  Hit       Crit
//	Combat Swords  151      131       273      82            50 (5%)   88 (6.3%)
//	  share (L)    20.6     17.8      37.2     5.6           6.8       12.0
//	  target       20.4     15.6      34.5     4.6           13.7      11.2
//	Daggers        152      132       250      110           50 (5%)   95 (6.8%)
//	  share (L)    20.7     18.0      34.1     7.5           6.8       12.9
//	  target       20.6     15.7      31.3     6.5           14.2      11.6
//
// Agility is 46.8% (swords) and 43.0% (daggers) of the non-Stamina total, under the 55% guardrail.
//
// Trinkets (8a, 1b): both sets wear the same two synthetic ilvl 71 trinkets, a stat trinket (31 Agility and 15
// crit rating, linear 46.0 of 45.9, 1.7-norm 36.0 of 36.8) and an on-use trinket with no passive stats, +335
// attack power for 20 sec on a 2 min offensive trinket cooldown (sim/rogue/forever_synthetic_trinkets.go),
// worth 37.2 weighted points over 180 sec like Earthstrike. Forever's real Phase 1 trinkets are removed or not
// simulated.
//
// The weapons keep the damage per second of the Era weapons they replace, swords for Sinister Strike
// (Vis'kag the Bloodletter and Brutality Blade) and daggers for the rest (Perdition's Blade and Core Hound
// Tooth), with the slot's budget in stats on top and no proc. Their type and speed are the ones each
// build sims best with among the speeds of Era Phase 1 BiS weapons (standards 22). Combat Swords holds
// Deathbringer's 2.9 sec axe (Forever lets rogues use axes) and Bone Slicing Hatchet's 1.7 sec axe.
// Seal Fate Swords holds the same main hand and Core Hound Tooth's 1.6 sec dagger. The dagger builds
// keep Perdition's Blade and Core Hound Tooth. Mutilate holds Felstriker's 1.7 sec dagger and Dal'Rend's
// Tribal Guardian's 1.8 sec sword, because its off hand strike is normalised to 2.4 sec for any weapon
// but a dagger. A fast off hand applies more poison. The ranged slot is a bow with Striker's Mark's
// damage.
//
// The hands and legs of each set form a set with the bonuses of Grimstitch Armor, Forever's crafted level 60
// rogue set (see sim/rogue/items_sets_pve.go), so only its 2 piece bonus (1% hit) is live, as for every
// class. The generator lived in the session scratchpad (b1.py), the numbers below are its output.
func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, rogueSyntheticItems...)
}

var rogueSyntheticItems = []*proto.UIItem{
	// Head, item level 66, in place of Nightslayer Cover (66, 16821). Linear 72.0 of 72.4, 1.7-norm 48.7 of 46.5.
	{
		Id: 994001, Name: "Combat Swords Synthetic Head (Phase 1)", Icon: "inv_helmet_41",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Agility: 34, stats.Stamina: 28, stats.Armor: 163}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Neck, item level 71, in place of Onyxia Tooth Pendant (65, 18404). Linear 42.0 of 42.0, 1.7-norm 26.7 of 28.0.
	{
		Id: 994002, Name: "Combat Swords Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_07",
		Type:       proto.ItemType_ItemTypeNeck,
		Stats:      stats.Stats{stats.Strength: 14, stats.Agility: 15}.ToFloatArray(),
		CritRating: 13,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Shoulder, item level 66, in place of Nightslayer Shoulder Pads (66, 16823). Linear 55.0 of 54.8, 1.7-norm 35.1 of 34.9.
	{
		Id: 994003, Name: "Combat Swords Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_23",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Agility: 21, stats.Stamina: 17, stats.Armor: 150}.ToFloatArray(),
		CritRating: 17,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Back, item level 71, in place of Cape of the Black Baron (63, 13340). Linear 42.0 of 42.0, 1.7-norm 27.4 of 28.0.
	{
		Id: 994004, Name: "Combat Swords Synthetic Back (Phase 1)", Icon: "inv_misc_cape_18",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Agility: 18, stats.Stamina: 15, stats.AttackPower: 18, stats.Armor: 54, stats.RangedAttackPower: 18}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Chest, item level 66, in place of Nightslayer Chestpiece (66, 16820). Linear 72.0 of 72.4, 1.7-norm 45.9 of 46.5.
	{
		Id: 994005, Name: "Combat Swords Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_07",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Strength: 24, stats.Agility: 26, stats.Stamina: 22, stats.Armor: 200}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Wrist, item level 71, in place of Nightslayer Bracelets (66, 16825). Linear 42.0 of 42.0, 1.7-norm 27.1 of 28.0.
	{
		Id: 994006, Name: "Combat Swords Synthetic Wrist (Phase 1)", Icon: "inv_bracer_02",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Strength: 17, stats.Stamina: 15, stats.Armor: 88}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Hands, item level 66, in place of Devilsaur Gauntlets (58, 15063). Linear 55.0 of 54.8, 1.7-norm 35.0 of 34.9.
	{
		Id: 994007, Name: "Combat Swords Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_21",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Strength: 18, stats.Agility: 20, stats.Armor: 125}.ToFloatArray(),
		CritRating: 17,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
		SetName:        "Combat Swords Synthetic (Phase 1)",
	},
	// Waist, item level 71, in place of Nightslayer Belt (66, 16827). Linear 59.0 of 58.9, 1.7-norm 39.3 of 37.4.
	{
		Id: 994008, Name: "Combat Swords Synthetic Waist (Phase 1)", Icon: "inv_belt_23",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Agility: 27, stats.Stamina: 22, stats.Armor: 113}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Legs, item level 66, in place of Devilsaur Leggings (60, 15062). Linear 72.5 of 72.4, 1.7-norm 47.4 of 46.5.
	{
		Id: 994009, Name: "Combat Swords Synthetic Legs (Phase 1)", Icon: "inv_pants_06",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Strength: 28, stats.Agility: 30, stats.AttackPower: 29, stats.Armor: 175, stats.RangedAttackPower: 29}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
		SetName:        "Combat Swords Synthetic (Phase 1)",
	},
	// Feet, item level 66, in place of Nightslayer Boots (66, 16824). Linear 55.0 of 54.8, 1.7-norm 35.1 of 34.9.
	{
		Id: 994010, Name: "Combat Swords Synthetic Feet (Phase 1)", Icon: "inv_boots_08",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Agility: 21, stats.Stamina: 17, stats.Armor: 138}.ToFloatArray(),
		CritRating: 17,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Finger, item level 71, in place of Quick Strike Ring (67, 18821). Linear 42.0 of 42.0, 1.7-norm 27.1 of 28.0.
	{
		Id: 994011, Name: "Combat Swords Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_15",
		Type:      proto.ItemType_ItemTypeFinger,
		Stats:     stats.Stats{stats.Strength: 15, stats.Agility: 17}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Finger II, item level 71, in place of Don Julio's Band (65, 19325). Linear 42.0 of 42.0, 1.7-norm 27.4 of 28.0.
	{
		Id: 994012, Name: "Combat Swords Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_03",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Agility: 18, stats.AttackPower: 18, stats.RangedAttackPower: 18}.ToFloatArray(),
		CritRating: 15,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Main Hand, item level 71, in place of Vis'kag the Bloodletter (74, 17075). Linear 32.0 of 31.9.
	{
		Id: 994013, Name: "Combat Swords Synthetic Main Hand (Phase 1)", Icon: "inv_axe_09",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeAxe, HandType: proto.HandType_HandTypeOneHand,
		Stats:     stats.Stats{stats.Strength: 7, stats.Agility: 8, stats.Stamina: 7}.ToFloatArray(),
		HitRating: 10,
		// Deathbringer's speed and damage spread at Vis'kag the Bloodletter's 55.2 DPS (standards 22:
		// a Phase 1 BiS speed, the one that sims best). Forever lets rogues use axes.
		WeaponDamageMin: 111.6, WeaponDamageMax: 208.5, WeaponSpeed: 2.9,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Off Hand, item level 71, in place of Brutality Blade (70, 18832). Linear 32.0 of 31.9.
	{
		Id: 994014, Name: "Combat Swords Synthetic Off Hand (Phase 1)", Icon: "inv_axe_11",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeAxe, HandType: proto.HandType_HandTypeOneHand,
		Stats:      stats.Stats{stats.Agility: 10, stats.Stamina: 8, stats.AttackPower: 10, stats.RangedAttackPower: 10}.ToFloatArray(),
		CritRating: 9,
		// Bone Slicing Hatchet's speed and damage spread at Brutality Blade's 51.6 DPS (standards 22:
		// a Phase 1 BiS speed, the one that sims best).
		WeaponDamageMin: 61, WeaponDamageMax: 114.4, WeaponSpeed: 1.7,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Ranged, item level 71, in place of Striker's Mark (69, 17069). Linear 19.5 of 19.6.
	{
		Id: 994015, Name: "Combat Swords Synthetic Ranged (Phase 1)", Icon: "inv_weapon_bow_08",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeBow,
		Stats: stats.Stats{stats.Strength: 8, stats.Agility: 8, stats.AttackPower: 7, stats.RangedAttackPower: 7}.ToFloatArray(),
		// Striker's Mark damage and speed.
		WeaponDamageMin: 69, WeaponDamageMax: 129, WeaponSpeed: 2.5,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Seal Fate Swords off hand: the Combat Swords off hand's stats as a dagger, at the speed Seal Fate sims best with.
	{
		Id: 994017, Name: "Seal Fate Swords Synthetic Off Hand (Phase 1)", Icon: "inv_weapon_shortblade_11",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeDagger, HandType: proto.HandType_HandTypeOneHand,
		Stats:      stats.Stats{stats.Agility: 10, stats.Stamina: 8, stats.AttackPower: 10, stats.RangedAttackPower: 10}.ToFloatArray(),
		CritRating: 9,
		// Core Hound Tooth's speed and damage spread at Brutality Blade's 51.6 DPS (standards 22:
		// a Phase 1 BiS speed, the one that sims best).
		WeaponDamageMin: 57.4, WeaponDamageMax: 107.7, WeaponSpeed: 1.6,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},

	// Head, item level 66, in place of Nightslayer Cover (66, 16821). Linear 72.0 of 72.4, 1.7-norm 48.6 of 46.5.
	{
		Id: 994101, Name: "Daggers Synthetic Head (Phase 1)", Icon: "inv_helmet_41",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Agility: 33, stats.Stamina: 29, stats.Armor: 163}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Neck, item level 71, in place of Onyxia Tooth Pendant (65, 18404). Linear 42.0 of 42.0, 1.7-norm 26.7 of 28.0.
	{
		Id: 994102, Name: "Daggers Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_07",
		Type:       proto.ItemType_ItemTypeNeck,
		Stats:      stats.Stats{stats.Strength: 14, stats.Agility: 14}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Shoulder, item level 66, in place of Nightslayer Shoulder Pads (66, 16823). Linear 55.0 of 54.8, 1.7-norm 35.0 of 34.9.
	{
		Id: 994103, Name: "Daggers Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_23",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Agility: 19, stats.Stamina: 17, stats.Armor: 150}.ToFloatArray(),
		CritRating: 19,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Back, item level 71, in place of Cape of the Black Baron (63, 13340). Linear 42.0 of 42.0, 1.7-norm 26.8 of 28.0.
	{
		Id: 994104, Name: "Daggers Synthetic Back (Phase 1)", Icon: "inv_misc_cape_18",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Agility: 16, stats.Stamina: 14, stats.AttackPower: 24, stats.Armor: 54, stats.RangedAttackPower: 24}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Chest, item level 66, in place of Nightslayer Chestpiece (66, 16820). Linear 72.0 of 72.4, 1.7-norm 45.9 of 46.5.
	{
		Id: 994105, Name: "Daggers Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_07",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Strength: 25, stats.Agility: 25, stats.Stamina: 22, stats.Armor: 200}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Wrist, item level 71, in place of Nightslayer Bracelets (66, 16825). Linear 42.0 of 42.0, 1.7-norm 27.1 of 28.0.
	{
		Id: 994106, Name: "Daggers Synthetic Wrist (Phase 1)", Icon: "inv_bracer_02",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Strength: 17, stats.Stamina: 15, stats.Armor: 88}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Hands, item level 66, in place of Devilsaur Gauntlets (58, 15063). Linear 55.0 of 54.8, 1.7-norm 35.0 of 34.9.
	{
		Id: 994107, Name: "Daggers Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_21",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Strength: 18, stats.Agility: 18, stats.Armor: 125}.ToFloatArray(),
		CritRating: 19,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
		SetName:        "Daggers Synthetic (Phase 1)",
	},
	// Waist, item level 71, in place of Nightslayer Belt (66, 16827). Linear 59.0 of 58.9, 1.7-norm 39.2 of 37.4.
	{
		Id: 994108, Name: "Daggers Synthetic Waist (Phase 1)", Icon: "inv_belt_23",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Agility: 26, stats.Stamina: 23, stats.Armor: 113}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Legs, item level 66, in place of Devilsaur Leggings (60, 15062). Linear 72.5 of 72.4, 1.7-norm 46.4 of 46.5.
	{
		Id: 994109, Name: "Daggers Synthetic Legs (Phase 1)", Icon: "inv_pants_06",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Strength: 27, stats.Agility: 26, stats.AttackPower: 39, stats.Armor: 175, stats.RangedAttackPower: 39}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
		SetName:        "Daggers Synthetic (Phase 1)",
	},
	// Feet, item level 66, in place of Nightslayer Boots (66, 16824). Linear 55.0 of 54.8, 1.7-norm 35.0 of 34.9.
	{
		Id: 994110, Name: "Daggers Synthetic Feet (Phase 1)", Icon: "inv_boots_08",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Agility: 19, stats.Stamina: 17, stats.Armor: 138}.ToFloatArray(),
		CritRating: 19,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Finger, item level 71, in place of Quick Strike Ring (67, 18821). Linear 42.0 of 42.0, 1.7-norm 27.1 of 28.0.
	{
		Id: 994111, Name: "Daggers Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_15",
		Type:      proto.ItemType_ItemTypeFinger,
		Stats:     stats.Stats{stats.Strength: 16, stats.Agility: 16}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Finger II, item level 71, in place of Don Julio's Band (65, 19325). Linear 42.0 of 42.0, 1.7-norm 26.8 of 28.0.
	{
		Id: 994112, Name: "Daggers Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_03",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Agility: 15, stats.AttackPower: 24, stats.RangedAttackPower: 24}.ToFloatArray(),
		CritRating: 15,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Main Hand, item level 71, in place of Perdition's Blade (77, 18816). Linear 32.0 of 31.9.
	{
		Id: 994113, Name: "Daggers Synthetic Main Hand (Phase 1)", Icon: "inv_weapon_shortblade_16",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeDagger, HandType: proto.HandType_HandTypeOneHand,
		Stats:     stats.Stats{stats.Strength: 8, stats.Agility: 7, stats.Stamina: 7}.ToFloatArray(),
		HitRating: 10,
		// Perdition's Blade damage and speed.
		WeaponDamageMin: 73, WeaponDamageMax: 137, WeaponSpeed: 1.8,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Off Hand, item level 71, in place of Core Hound Tooth (70, 18805). Linear 32.0 of 31.9.
	{
		Id: 994114, Name: "Daggers Synthetic Off Hand (Phase 1)", Icon: "inv_weapon_shortblade_11",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeDagger, HandType: proto.HandType_HandTypeOneHand,
		Stats:      stats.Stats{stats.Agility: 9, stats.Stamina: 8, stats.AttackPower: 12, stats.RangedAttackPower: 12}.ToFloatArray(),
		CritRating: 9,
		// Core Hound Tooth damage and speed.
		WeaponDamageMin: 57, WeaponDamageMax: 107, WeaponSpeed: 1.6,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Ranged, item level 71, in place of Striker's Mark (69, 17069). Linear 19.5 of 19.6.
	{
		Id: 994115, Name: "Daggers Synthetic Ranged (Phase 1)", Icon: "inv_weapon_bow_08",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeBow,
		Stats: stats.Stats{stats.Strength: 7, stats.Agility: 7, stats.AttackPower: 11, stats.RangedAttackPower: 11}.ToFloatArray(),
		// Striker's Mark damage and speed.
		WeaponDamageMin: 69, WeaponDamageMax: 129, WeaponSpeed: 2.5,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Mutilate off hand: the dagger off hand's stats as a sword, at the speed Mutilate sims best with.
	{
		Id: 994116, Name: "Mutilate Synthetic Off Hand (Phase 1)", Icon: "inv_sword_40",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand,
		Stats:      stats.Stats{stats.Agility: 9, stats.Stamina: 8, stats.AttackPower: 12, stats.RangedAttackPower: 12}.ToFloatArray(),
		CritRating: 9,
		// Dal'Rend's Tribal Guardian's speed and damage spread at Core Hound Tooth's 51.3 DPS (standards 22:
		// a Phase 1 BiS speed, the one that sims best).
		WeaponDamageMin: 64.4, WeaponDamageMax: 120.1, WeaponSpeed: 1.8,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// Mutilate main hand: the dagger main hand's stats at the speed Mutilate sims best with.
	{
		Id: 994117, Name: "Mutilate Synthetic Main Hand (Phase 1)", Icon: "inv_weapon_shortblade_25",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeDagger, HandType: proto.HandType_HandTypeOneHand,
		Stats:     stats.Stats{stats.Strength: 8, stats.Agility: 7, stats.Stamina: 7}.ToFloatArray(),
		HitRating: 10,
		// Felstriker's speed and damage spread at Perdition's Blade's 58.3 DPS (standards 22:
		// a Phase 1 BiS speed, the one that sims best).
		WeaponDamageMin: 69.1, WeaponDamageMax: 129.2, WeaponSpeed: 1.7,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},

	// Stat trinket, item level 71, shared by both sets (standards 1b). Linear 46.0 of 45.9, 1.7-norm 36.0 of 36.8.
	{
		Id: 994201, Name: "Rogue Synthetic Trinket (Phase 1)", Icon: "inv_misc_bone_08",
		Type:       proto.ItemType_ItemTypeTrinket,
		Stats:      stats.Stats{stats.Agility: 31}.ToFloatArray(),
		CritRating: 15,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	// On-use trinket, item level 71, shared by both sets (standards 1b): no passive stats, +335 attack power for 20 sec
	// every 2 min on the offensive trinket cooldown (sim/rogue/forever_synthetic_trinkets.go). Two uses in a 180 sec
	// fight average 335 x 0.5 x 40 / 180 = 37.2 weighted points, Earthstrike's size.
	{
		Id: 994202, Name: "Rogue Synthetic On-Use Trinket (Phase 1)", Icon: "inv_misc_armorkit_10",
		Type: proto.ItemType_ItemTypeTrinket,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
}
