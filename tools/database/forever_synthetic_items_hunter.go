package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The hunter's synthetic Phase 1 items, made the way forever_synthetic_items.go describes, ids
// 993000 to 993999. Two sets, each with two synthetic trinkets:
//   - Ranged Hunter (993001 to 993016) for Beast Mastery, Marksmanship and Lone Wolf. They shoot
//     from range on the same stats: Agility (2 ranged attack power and crit), ranged attack power,
//     crit and hit. The pet takes 10% of the hunter's ranged attack power and all of its crit
//     (sim/hunter/pet.go), so the same stats feed it. The two-hander is a stat stick.
//   - Survival (993101 to 993117). Survival melees under Forever, so it wants melee attack power and
//     a pair of weapons (Predator's Edge adds 50% off hand damage), and some Strength.
//
// Each piece stands in for an Era Phase 1 hunter BiS item: the Giantstalker set, Onyxia Tooth
// Pendant, Wristguards of True Flight, Cape of the Black Baron, the Devilsaur hands and legs, Quick
// Strike Ring, Blackstone Ring, Barbarous Blade (a pair of Brutality Blades for Survival) and
// Rhok'delar. The weapons keep their damage per second. Survival's pair takes the type and speed it
// sims best with among the speeds of Era Phase 1 BiS weapons (standards 22): Deathbringer's 2.9 sec
// axe in the main hand and Core Hound Tooth's 1.6 sec dagger in the off hand. The others keep their
// weapon's type and speed.
//
// Budgets follow the cross-class standard 1b (orchestration/standards.md, 2026-10-13), the budget
// real level 60 epics follow, per archetype:
//   - Weights: Strength, Agility, Stamina, Intellect and Spirit 1, attack power 0.5, ranged attack
//     power 0.4 (attack power and ranged attack power on one item count once, at 0.5), hit and
//     crit rating 1 (10 per 1% hit, 14 per 1% crit). Armor is free.
//   - Each armor piece's weighted linear total is slot mod x L(ilvl) x f within 1%, with L(ilvl) =
//     1.125 ilvl + 2.0, slot mods 1 (head, chest, legs), 0.75 (shoulders, hands, waist, feet) and
//     0.5625 (neck, back, wrist, rings), and the hunter's f 1.028 / 1.056 / 0.977 per tier. Head,
//     shoulders, chest, hands, legs and feet are item level 66 (78.4 and 60.4), the rest 71 (waist
//     64.8, the others 45.0).
//   - Shape check: each armor piece's 1.7-norm value, (sum (weight x amount)^1.7)^(1/1.7), is within
//     6% of slot mod x (0.661 ilvl + 2.92): 46.5, 34.9, 37.4 and 28.0. That takes the hunter's real
//     4 / 4 / 3 stat lines per piece.
//   - Weapons: the hunter melee two-hander 0.98 x L(76) = 85.8, the bow 0.24 x L(71) = 19.7. The
//     standard has no hunter one-hander, so Survival's pair takes the physical one-hander's 0.39 x
//     L(71) times the hunter / physical ratio of the two-handers (0.98 / 0.90), 0.42 x L(71) = 34.4
//     each.
//   - Trinkets at item level 71: a stat trinket of 32 Agility and 14 crit rating (46.0 linear,
//     36.4 in the 1.7-norm), and an on-use trinket worth 37 weighted points over a 180 sec fight:
//     415 ranged attack power for the ranged set, 335 attack power and ranged attack power for
//     Survival, 20 sec every 2 min.
//
// The stat mix is the 5a target for the spec (orchestration/gear_targets.md), linear shares of the
// set's weighted total with the trinkets left out. Stamina is the spec's own share (17.2% ranged,
// 15.5% Survival) and Spirit 10%, as for every mana user. Hit goes to the cap against a level 63
// boss and no further: 90 rating (9%, the ranged cap) for the ranged set, 60 (6%) for Survival,
// whose Surefooted gives the other 3% of the 9% yellow cap. The other stats keep the table's
// proportions in the room left: ranged Agility 34.6%, crit 17.8%, ranged attack power 4.7%,
// Intellect 3.5%, Strength 1.1%; Survival Agility 31.1%, crit 14.7%, Strength 11.2%, attack power
// 9.9% and no Intellect (its 0.3% is 2 points). The amounts were fitted by a script in the session
// scratchpad: proportional fitting to both the piece totals and the set shares, then rounding with
// one line per piece taking up the rounding.
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
		Stats:      stats.Stats{stats.Agility: 23, stats.Stamina: 14, stats.Intellect: 15, stats.Armor: 343}.ToFloatArray(),
		CritRating: 26,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Neck, item level 71, in place of Onyxia Tooth Pendant (74, 18404).
	{
		Id: 993002, Name: "Ranged Hunter Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_09",
		Type:      proto.ItemType_ItemTypeNeck,
		Stats:     stats.Stats{stats.Agility: 15}.ToFloatArray(),
		HitRating: 13, CritRating: 17,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Shoulder, item level 66, in place of Giantstalker's Epaulets (66, 16848).
	{
		Id: 993003, Name: "Ranged Hunter Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_10",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Agility: 18, stats.Stamina: 11, stats.Spirit: 15, stats.Armor: 317}.ToFloatArray(),
		HitRating: 16,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Back, item level 71, in place of Cape of the Black Baron (63, 13340).
	{
		Id: 993004, Name: "Ranged Hunter Synthetic Back (Phase 1)", Icon: "inv_misc_cape_18",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Agility: 21, stats.Stamina: 12, stats.RangedAttackPower: 30, stats.Armor: 54}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Chest, item level 66, in place of Giantstalker's Breastplate (66, 16845).
	{
		Id: 993005, Name: "Ranged Hunter Synthetic Chest (Phase 1)", Icon: "inv_chest_chain_03",
		SetName: "Hunter Synthetic Ranged (Phase 1)",
		Type:    proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:      stats.Stats{stats.Agility: 23, stats.Stamina: 13, stats.Spirit: 17, stats.Armor: 422}.ToFloatArray(),
		CritRating: 25,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Wrist, item level 71, in place of Wristguards of True Flight (71, 18812).
	{
		Id: 993006, Name: "Ranged Hunter Synthetic Wrist (Phase 1)", Icon: "inv_bracer_17",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Agility: 19, stats.Stamina: 11, stats.Armor: 198}.ToFloatArray(),
		HitRating: 15,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Hands, item level 66, in place of Devilsaur Gauntlets (58, 15063).
	{
		Id: 993007, Name: "Ranged Hunter Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_10",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Agility: 17, stats.Stamina: 10, stats.Armor: 264}.ToFloatArray(),
		HitRating: 14, CritRating: 19,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Waist, item level 71, in place of Giantstalker's Belt (66, 16851).
	{
		Id: 993008, Name: "Ranged Hunter Synthetic Waist (Phase 1)", Icon: "inv_belt_28",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:      stats.Stats{stats.Agility: 19, stats.Stamina: 11, stats.Spirit: 14, stats.Armor: 237}.ToFloatArray(),
		CritRating: 21,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Legs, item level 66, in place of Devilsaur Leggings (60, 15062).
	{
		Id: 993009, Name: "Ranged Hunter Synthetic Legs (Phase 1)", Icon: "inv_pants_mail_03",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Agility: 26, stats.Stamina: 14, stats.Spirit: 19, stats.Armor: 369}.ToFloatArray(),
		HitRating: 19,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Feet, item level 66, in place of Giantstalker's Boots (66, 16849).
	{
		Id: 993010, Name: "Ranged Hunter Synthetic Feet (Phase 1)", Icon: "inv_boots_chain_04",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Agility: 19, stats.Stamina: 12, stats.Intellect: 13, stats.Spirit: 16, stats.Armor: 290}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Finger, item level 71, in place of Quick Strike Ring (67, 18821).
	{
		Id: 993011, Name: "Ranged Hunter Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_15",
		Type:      proto.ItemType_ItemTypeFinger,
		Stats:     stats.Stats{stats.Agility: 15}.ToFloatArray(),
		HitRating: 13, CritRating: 17,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Finger II, item level 71, in place of Blackstone Ring (54, 17713).
	{
		Id: 993012, Name: "Ranged Hunter Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_15",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Agility: 16, stats.Stamina: 10}.ToFloatArray(),
		CritRating: 19,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Two-Hand, item level 76, in place of Barbarous Blade (63, 18520).
	{
		Id: 993013, Name: "Ranged Hunter Synthetic Two-Hand (Phase 1)", Icon: "inv_sword_50",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeTwoHand,
		Stats: stats.Stats{stats.Agility: 36, stats.Strength: 9, stats.Stamina: 21, stats.RangedAttackPower: 49}.ToFloatArray(),
		// Barbarous Blade damage and speed.
		WeaponDamageMin: 138, WeaponDamageMax: 207, WeaponSpeed: 3.2,
		Ilvl: 76, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Ranged, item level 71, in place of Rhok'delar, Longbow of the Ancient Keepers (75, 18713).
	{
		Id: 993014, Name: "Ranged Hunter Synthetic Ranged (Phase 1)", Icon: "inv_weapon_bow_01",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeBow,
		Stats: stats.Stats{stats.Agility: 13, stats.RangedAttackPower: 17}.ToFloatArray(),
		// Rhok'delar damage and speed.
		WeaponDamageMin: 89, WeaponDamageMax: 166, WeaponSpeed: 2.9,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Trinket, item level 71: the synthetic stat trinket (standards 1b), 0.56 x L(71) = 45.9 linear.
	{
		Id: 993015, Name: "Ranged Hunter Synthetic Trinket (Phase 1)", Icon: "inv_jewelry_talisman_07",
		Type:       proto.ItemType_ItemTypeTrinket,
		Stats:      stats.Stats{stats.Agility: 32}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Trinket, item level 71: the synthetic on-use trinket (standards 1b), 415 ranged attack power
	// for 20 sec every 2 min (sim/hunter/forever_synthetic_trinkets.go). No passive stats.
	{
		Id: 993016, Name: "Ranged Hunter Synthetic On-Use Trinket (Phase 1)", Icon: "inv_misc_gem_bloodstone_01",
		Type: proto.ItemType_ItemTypeTrinket,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},

	// Head, item level 66, in place of Giantstalker's Helmet (66, 16846).
	{
		Id: 993101, Name: "Survival Synthetic Head (Phase 1)", Icon: "inv_helmet_05",
		SetName: "Hunter Synthetic Survival (Phase 1)",
		Type:    proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:      stats.Stats{stats.Agility: 20, stats.Strength: 16, stats.Stamina: 13, stats.Armor: 343}.ToFloatArray(),
		CritRating: 29,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Neck, item level 71, in place of Onyxia Tooth Pendant (74, 18404).
	{
		Id: 993102, Name: "Survival Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_09",
		Type:      proto.ItemType_ItemTypeNeck,
		Stats:     stats.Stats{stats.Agility: 14}.ToFloatArray(),
		HitRating: 12, CritRating: 19,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Shoulder, item level 66, in place of Giantstalker's Epaulets (66, 16848).
	{
		Id: 993103, Name: "Survival Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_10",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Agility: 19, stats.Stamina: 12, stats.Spirit: 16, stats.AttackPower: 27, stats.RangedAttackPower: 27, stats.Armor: 317}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Back, item level 71, in place of Cape of the Black Baron (63, 13340).
	{
		Id: 993104, Name: "Survival Synthetic Back (Phase 1)", Icon: "inv_misc_cape_18",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Agility: 18, stats.Stamina: 12, stats.Spirit: 15, stats.Armor: 54}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Chest, item level 66, in place of Giantstalker's Breastplate (66, 16845).
	{
		Id: 993105, Name: "Survival Synthetic Chest (Phase 1)", Icon: "inv_chest_chain_03",
		SetName: "Hunter Synthetic Survival (Phase 1)",
		Type:    proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Agility: 24, stats.Strength: 19, stats.Stamina: 15, stats.Spirit: 20, stats.Armor: 422}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Wrist, item level 71, in place of Wristguards of True Flight (71, 18812).
	{
		Id: 993106, Name: "Survival Synthetic Wrist (Phase 1)", Icon: "inv_bracer_17",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Agility: 18, stats.Stamina: 11, stats.Armor: 198}.ToFloatArray(),
		HitRating: 16,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Hands, item level 66, in place of Devilsaur Gauntlets (58, 15063).
	{
		Id: 993107, Name: "Survival Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_10",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:      stats.Stats{stats.Agility: 17, stats.Stamina: 10, stats.AttackPower: 21, stats.RangedAttackPower: 21, stats.Armor: 264}.ToFloatArray(),
		CritRating: 23,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Waist, item level 71, in place of Giantstalker's Belt (66, 16851).
	{
		Id: 993108, Name: "Survival Synthetic Waist (Phase 1)", Icon: "inv_belt_28",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:      stats.Stats{stats.Agility: 16, stats.Stamina: 11, stats.Spirit: 14, stats.Armor: 237}.ToFloatArray(),
		CritRating: 24,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Legs, item level 66, in place of Devilsaur Leggings (60, 15062).
	{
		Id: 993109, Name: "Survival Synthetic Legs (Phase 1)", Icon: "inv_pants_mail_03",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Agility: 25, stats.Strength: 20, stats.Stamina: 16, stats.AttackPower: 35, stats.RangedAttackPower: 35, stats.Armor: 369}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Feet, item level 66, in place of Giantstalker's Boots (66, 16849).
	{
		Id: 993110, Name: "Survival Synthetic Feet (Phase 1)", Icon: "inv_boots_chain_04",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Agility: 18, stats.Stamina: 11, stats.Spirit: 15, stats.Armor: 290}.ToFloatArray(),
		HitRating: 16,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Finger, item level 71, in place of Quick Strike Ring (67, 18821).
	{
		Id: 993111, Name: "Survival Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_15",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Strength: 12, stats.AttackPower: 22, stats.RangedAttackPower: 22}.ToFloatArray(),
		CritRating: 22,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Finger II, item level 71, in place of Blackstone Ring (54, 17713).
	{
		Id: 993112, Name: "Survival Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_15",
		Type:      proto.ItemType_ItemTypeFinger,
		Stats:     stats.Stats{stats.Agility: 18, stats.Stamina: 11}.ToFloatArray(),
		HitRating: 16,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Main Hand, item level 71, in place of Brutality Blade (70, 18832), main hand.
	{
		Id: 993113, Name: "Survival Synthetic Main Hand (Phase 1)", Icon: "inv_axe_09",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeAxe, HandType: proto.HandType_HandTypeOneHand,
		Stats: stats.Stats{stats.Agility: 14, stats.Strength: 11, stats.AttackPower: 19, stats.RangedAttackPower: 19}.ToFloatArray(),
		// Deathbringer's speed and damage spread at Brutality Blade's 51.6 DPS (standards 22:
		// a Phase 1 BiS speed, the one that sims best).
		WeaponDamageMin: 104.3, WeaponDamageMax: 194.9, WeaponSpeed: 2.9,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Off Hand, item level 71, in place of Brutality Blade (70, 18832), off hand.
	{
		Id: 993114, Name: "Survival Synthetic Off Hand (Phase 1)", Icon: "inv_weapon_shortblade_11",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeDagger, HandType: proto.HandType_HandTypeOneHand,
		Stats: stats.Stats{stats.Agility: 14, stats.Strength: 11, stats.AttackPower: 19, stats.RangedAttackPower: 19}.ToFloatArray(),
		// Core Hound Tooth's speed and damage spread at Brutality Blade's 51.6 DPS (standards 22:
		// a Phase 1 BiS speed, the one that sims best).
		WeaponDamageMin: 57.4, WeaponDamageMax: 107.7, WeaponSpeed: 1.6,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Ranged, item level 71, in place of Rhok'delar, Longbow of the Ancient Keepers (75, 18713).
	{
		Id: 993115, Name: "Survival Synthetic Ranged (Phase 1)", Icon: "inv_weapon_bow_01",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeBow,
		Stats: stats.Stats{stats.Agility: 12, stats.AttackPower: 15, stats.RangedAttackPower: 15}.ToFloatArray(),
		// Rhok'delar damage and speed.
		WeaponDamageMin: 89, WeaponDamageMax: 166, WeaponSpeed: 2.9,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Trinket, item level 71: the synthetic stat trinket (standards 1b), 0.56 x L(71) = 45.9 linear.
	{
		Id: 993116, Name: "Survival Synthetic Trinket (Phase 1)", Icon: "inv_jewelry_talisman_07",
		Type:       proto.ItemType_ItemTypeTrinket,
		Stats:      stats.Stats{stats.Agility: 32}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	// Trinket, item level 71: the synthetic on-use trinket (standards 1b), 335 attack power and ranged attack power
	// for 20 sec every 2 min (sim/hunter/forever_synthetic_trinkets.go). No passive stats.
	{
		Id: 993117, Name: "Survival Synthetic On-Use Trinket (Phase 1)", Icon: "inv_misc_gem_bloodstone_01",
		Type: proto.ItemType_ItemTypeTrinket,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
}
