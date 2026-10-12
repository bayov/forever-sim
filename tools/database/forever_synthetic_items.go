package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// ForeverSyntheticItems are made up level 60 items for a Phase 1 enhancement shaman under Forever.
//
// Forever's level 60 client gear is months old and not tuned yet, and Molten Core and Onyxia have no
// Forever data at all. So the gear search on real items ends up with Classic items, and it can't tell us
// what an enhancement shaman with Forever's hybrid gear (Strength, Intellect and spell power together) does.
// These items fill that gap. They are not in the game, have no source, and the gear search leaves them out.
//
// The pieces take the place of Era Phase 1 items an enhancement shaman would wear (Crown of Destruction,
// the Scholomance Bloodmail set, Death's Clutch and Wristguards of True Flight). Each one gets the full epic
// budget of its slot. Since 2026-10-12 every class's synthetic set follows one standard
// (orchestration/standards.md on the coordination branch): an epic piece is worth 1.2 * item level -
// 36.4, times 1 for the head, chest, legs and two-hander, 0.77 for the shoulders, hands, waist and feet and
// 0.5625 for the neck, back, wrist and rings. The head, shoulders, chest, hands, legs and feet are item
// level 66, the neck, back, wrist, waist and rings 71, and two-handers 76. The stats on a piece add up as
// (sum of (weight * amount)^1.5)^(2/3), with Strength, Agility, Stamina, Intellect and Spirit at 1, attack
// power at 0.5 and spell power at Forever's tuned price. A point of spell power costs 0.55 of a Strength
// point there, which is 0.73 of what Era pays at the same item level (the ratio on Forever's tuned gear
// at item level 25 to 54). Hit and crit are rating at 10 and 14 per percent, the Forever ratio, at 1 a
// point. For the legs: 1.2 * 66 - 36.4 = 42.8, from 17 Strength, 13 Stamina, 13 Intellect, 10 Spirit, 25
// spell power and 10 hit rating. Armor is free.
//
// We then hand pick a stat mix per slot, so the pieces are not all the same:
//   - Spirit is on 9 of the 13 pieces, about 10% of the set's budget. The Forever devs said they want Spirit
//     to matter again, so we expect it on endgame gear, even though their level 10 to 54 mail barely has it.
//   - Spell power is about 22% of the budget. Some pieces lean on it (the first ring is half spell power)
//     and some on Strength (the hands) or Intellect and Spirit (the second ring).
//   - Not every piece has every stat. The neck and wrist have no Intellect, the wrist has attack power
//     instead of Strength, and the shoulders and hands carry the set's crit.
//   - The set as a whole has 90 hit rating, the 9% melee hit cap against a level 63 boss (it had the 10%
//     the Era Phase 1 BiS carries before 2026-10-12). It sits on 9 pieces.
//
// The hands and legs carry The Spiritcaller's Rage, Forever's level 60 enhancement set (see
// sim/shaman/item_sets_pve.go). The standard keeps only its 2 piece bonus live (1% attack and cast speed),
// so exactly two pieces are in the set. The generator script lived in the session scratchpad (synth.py,
// rebudgeted with synth_enh2.py), the numbers below are its output.
//
// The other specs' synthetic items live in forever_synthetic_items_<class>.go, one file per class,
// and add themselves to this list from an init function. Each class has its own block of ids, all
// at or above 990000 so rotopt's -gear-synthetic flag sees them:
//
//	990000-990999 Shaman      991000-991999 Warrior    992000-992999 Paladin
//	993000-993999 Hunter      994000-994999 Rogue      995000-995999 Priest
//	996000-996999 Mage        997000-997999 Warlock    998000-998999 Druid

var ForeverSyntheticItems = []*proto.UIItem{
	// Head, item level 66, in place of Crown of Destruction (76, 18817).
	{
		Id: 990001, Name: "Enhancement Synthetic Head (Phase 1)", Icon: "inv_crown_02",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Strength: 11, stats.Stamina: 11, stats.Intellect: 9, stats.Spirit: 9, stats.SpellPower: 30, stats.Armor: 392}.ToFloatArray(),
		HitRating: 10, CritRating: 14,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Neck, item level 71.
	{
		Id: 990002, Name: "Enhancement Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_09",
		Type:      proto.ItemType_ItemTypeNeck,
		Stats:     stats.Stats{stats.Agility: 12, stats.Stamina: 8, stats.SpellPower: 24}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Shoulder, item level 66, in place of Death's Clutch (62, 14503).
	{
		Id: 990003, Name: "Enhancement Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_25",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:      stats.Stats{stats.Strength: 11, stats.Stamina: 9, stats.Intellect: 8, stats.SpellPower: 26, stats.Armor: 317}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Back, item level 71.
	{
		Id: 990004, Name: "Enhancement Synthetic Back (Phase 1)", Icon: "inv_misc_cape_05",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 11, stats.Intellect: 9, stats.Spirit: 10, stats.SpellPower: 24, stats.Armor: 54}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Chest, item level 66, in place of Bloodmail Hauberk (61, 14611).
	{
		Id: 990005, Name: "Enhancement Synthetic Chest (Phase 1)", Icon: "inv_chest_leather_05",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Strength: 14, stats.Stamina: 13, stats.Intellect: 13, stats.Spirit: 10, stats.SpellPower: 31, stats.Armor: 422}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Wrist, item level 71, in place of Wristguards of True Flight (71, 18812).
	{
		Id: 990006, Name: "Enhancement Synthetic Wrist (Phase 1)", Icon: "inv_bracer_02",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.AttackPower: 27, stats.Stamina: 6, stats.SpellPower: 24, stats.Armor: 198}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Hands, item level 66, in place of Bloodmail Gauntlets (61, 14615).
	{
		Id: 990007, Name: "Enhancement Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_26",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Strength: 11, stats.Stamina: 7, stats.Intellect: 5, stats.SpellPower: 22, stats.Armor: 264}.ToFloatArray(),
		HitRating: 10, CritRating: 14,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
		SetName:        "Enhancement Synthetic (Phase 1)",
	},
	// Waist, item level 71, in place of Bloodmail Belt (61, 14614).
	{
		Id: 990008, Name: "Enhancement Synthetic Waist (Phase 1)", Icon: "inv_belt_23",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Strength: 12, stats.Agility: 12, stats.Stamina: 12, stats.Spirit: 11, stats.SpellPower: 21, stats.Armor: 255}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Legs, item level 66, in place of Bloodmail Legguards (61, 14612).
	{
		Id: 990009, Name: "Enhancement Synthetic Legs (Phase 1)", Icon: "inv_pants_06",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Strength: 17, stats.Stamina: 13, stats.Intellect: 13, stats.Spirit: 10, stats.SpellPower: 25, stats.Armor: 369}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
		SetName:        "Enhancement Synthetic (Phase 1)",
	},
	// Feet, item level 66, in place of Bloodmail Boots (61, 14616).
	{
		Id: 990010, Name: "Enhancement Synthetic Feet (Phase 1)", Icon: "inv_boots_01",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Agility: 12, stats.Stamina: 14, stats.Intellect: 9, stats.Spirit: 11, stats.SpellPower: 20, stats.Armor: 290}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Finger, item level 71.
	{
		Id: 990011, Name: "Enhancement Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_07",
		Type:      proto.ItemType_ItemTypeFinger,
		Stats:     stats.Stats{stats.Strength: 10, stats.Stamina: 6, stats.SpellPower: 30}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Second finger, item level 71. This is the caster ring of the two.
	{
		Id: 990013, Name: "Enhancement Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_15",
		Type:  proto.ItemType_ItemTypeFinger,
		Stats: stats.Stats{stats.Stamina: 11, stats.Intellect: 13, stats.Spirit: 11, stats.SpellPower: 16}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Two-Hand, item level 76.
	{
		Id: 990012, Name: "Enhancement Synthetic Two-Hand (Phase 1)", Icon: "inv_axe_09",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeAxe, HandType: proto.HandType_HandTypeTwoHand,
		Stats:     stats.Stats{stats.Strength: 23, stats.Agility: 14, stats.Stamina: 19, stats.Intellect: 14, stats.Spirit: 9, stats.SpellPower: 26}.ToFloatArray(),
		HitRating: 10,
		// Spinal Reaper damage and speed.
		WeaponDamageMin: 203, WeaponDamageMax: 305, WeaponSpeed: 3.4,
		Ilvl: 76, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
}
