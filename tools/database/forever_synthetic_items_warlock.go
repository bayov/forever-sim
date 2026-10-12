package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The synthetic Phase 1 warlock items, ids 997001 to 997999. See forever_synthetic_items.go for what these
// items are and why we need them.
//
// One set serves Affliction, Demonology and Destruction. All three are casters that want spell power
// first, then hit up to the cap, crit, Intellect and Spirit, and the Era Phase 1 BiS lists give the three
// the same gear (the Felheart set, Robe of Volatile Power, Mana Igniting Cord, Choker of the Fire Lord,
// Azuresong Mageblade). Shadow and Fire spell power would be the only thing to split them on, and Phase 1
// has next to none of it. Measured on the synthetic set, the three specs' stat weights come out in the
// same order (see the Level 60 section of forever-wiki/sim-status.md).
//
// Each piece takes the place of an Era Phase 1 item a warlock would wear: the Felheart set (Molten Core
// tier, item level 66), Choker of the Fire Lord (78), Robe of Volatile Power (66), Mana Igniting Cord (71),
// Nemesis Leggings (76, Ragnaros), Ring of Spell Power (66), Azuresong Mageblade (71), Fire Runed Grimoire
// (70) and Crimson Shocker (63, so Molten Core's 66). The cloak and the second ring take Molten Core's 66.
// Each piece gets the full epic budget of its slot at that item level.
//
// The budget is the one shared by every class's synthetic set (see forever_synthetic_items_shaman.go). An
// epic piece is worth V = 1.2 * item level - 36.4 (42.8 at 66, 48.8 at 71, 47.6 at 70, 54.8 at 76, 57.2
// at 78), times the slot modifier: 1 for the head, chest and legs, 0.77 for the shoulders, hands, waist
// and feet, 0.5625 for the neck, back, wrist, fingers and the held in off hand, and Classic's 0.42 for a
// one-hand weapon and 0.32 for a wand. The stats on a piece add up as (sum of (weight * amount)^1.5)^(2/3),
// where a point of Stamina, Intellect or Spirit weighs 1, a point of spell power 0.55 (Forever's tuned
// price) and a point of hit or crit rating 1 (10 hit rating is 1% hit, 14 crit rating 1% crit). Armor is
// free, it is the Era piece's. We pick the Stamina, Intellect, Spirit, hit and crit of each piece, and
// spell power fills the rest. For the chest: 42.8 * 1, and 12 Stamina, 14 Intellect, 9 Spirit and 14 crit
// rating leave room for 40 spell power: (12^1.5 + 14^1.5 + 9^1.5 + 14^1.5 + (40 * 0.55)^1.5)^(2/3) = 42.4.
// For the waist: 48.8 * 0.77 = 37.6, from 10 Stamina, 10 Intellect, 7 Spirit, 14 crit rating and 38 spell
// power. The generator script lived in the session scratchpad (synth_lock.py), the numbers below are its
// output.
//
// The set comes to 527 spell power, 118 Intellect, 129 Stamina, 76 Spirit, 70 crit rating (5%) and 30 hit
// rating (3%). As shares of the budget that is 41% spell power, 17% Intellect, 18% Stamina, 10.7% Spirit,
// 10% crit and 4% hit:
//   - Spirit is on 10 of the 15 pieces, about 10% of the budget. Under Forever it also adds a point of mana
//     to every Life Tap, on top of the regen.
//   - The crit sits where the Era pieces had it: 1% each on the head, chest (Robe of Volatile Power), waist
//     (Mana Igniting Cord), the first ring and the sword (Azuresong Mageblade).
//   - 3% hit on the neck, wrist and legs, about what the Era Phase 1 caster gear had. The Tier 1 bonus adds
//     1% more, and Affliction's Suppression up to 5%.
//   - Not every piece has every stat: the back and the first ring have no Intellect, the neck, wrist, hands
//     and rings no Spirit (but the second ring).
//
// The head, shoulders, chest, hands, legs and feet form a set with the bonuses of Demonheart Raiment,
// Forever's warlock Tier 1 set (item set 2100, see sim/warlock/item_sets_pve.go).
//
// The main hand is a sword on purpose: Azuresong Mageblade is the Era Phase 1 weapon, and that gives a
// Human warlock its Sword Specialization (2% crit). A staff or a dagger would leave no warlock race with a
// weapon racial.

func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, warlockSyntheticItems...)
}

var warlockSyntheticItems = []*proto.UIItem{
	// Head, item level 66, in place of Felheart Horns (66, 16808).
	{
		Id: 997001, Name: "Warlock Synthetic Head (Phase 1)", Icon: "inv_helmet_58",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 14, stats.Intellect: 13, stats.Spirit: 9, stats.SpellPower: 39, stats.Armor: 83}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
		SetName:        "Warlock Synthetic (Phase 1)",
	},
	// Neck, item level 78, in place of Choker of the Fire Lord (78, 18814).
	{
		Id: 997002, Name: "Warlock Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_17",
		Type:      proto.ItemType_ItemTypeNeck,
		Stats:     stats.Stats{stats.Stamina: 8, stats.Intellect: 8, stats.SpellPower: 40}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      78, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Shoulder, item level 66, in place of Felheart Shoulder Pads (66, 16807).
	{
		Id: 997003, Name: "Warlock Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_23",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 10, stats.Intellect: 9, stats.Spirit: 7, stats.SpellPower: 42, stats.Armor: 76}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
		SetName:        "Warlock Synthetic (Phase 1)",
	},
	// Back, item level 66, in place of Molten Core's 66.
	{
		Id: 997004, Name: "Warlock Synthetic Back (Phase 1)", Icon: "inv_misc_cape_20",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 7, stats.Spirit: 6, stats.SpellPower: 35, stats.Armor: 48}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Chest, item level 66, in place of Robe of Volatile Power (66, 19145).
	{
		Id: 997005, Name: "Warlock Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_09",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 12, stats.Intellect: 14, stats.Spirit: 9, stats.SpellPower: 40, stats.Armor: 102}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
		SetName:        "Warlock Synthetic (Phase 1)",
	},
	// Wrist, item level 66, in place of Felheart Bracers (66, 16804).
	{
		Id: 997006, Name: "Warlock Synthetic Wrist (Phase 1)", Icon: "inv_bracer_07",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 8, stats.Intellect: 7, stats.SpellPower: 23, stats.Armor: 44}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Hands, item level 66, in place of Felheart Gloves (66, 16805).
	{
		Id: 997007, Name: "Warlock Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_19",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 9, stats.Intellect: 9, stats.Spirit: 8, stats.SpellPower: 42, stats.Armor: 63}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
		SetName:        "Warlock Synthetic (Phase 1)",
	},
	// Waist, item level 71, in place of Mana Igniting Cord (71, 19136).
	{
		Id: 997008, Name: "Warlock Synthetic Waist (Phase 1)", Icon: "inv_belt_13",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 10, stats.Intellect: 10, stats.Spirit: 7, stats.SpellPower: 38, stats.Armor: 61}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Legs, item level 76, in place of Nemesis Leggings (76, 16930).
	{
		Id: 997009, Name: "Warlock Synthetic Legs (Phase 1)", Icon: "inv_pants_cloth_14",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 16, stats.Intellect: 15, stats.Spirit: 10, stats.SpellPower: 66, stats.Armor: 102}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      76, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
		SetName:        "Warlock Synthetic (Phase 1)",
	},
	// Feet, item level 66, in place of Felheart Slippers (66, 16803).
	{
		Id: 997010, Name: "Warlock Synthetic Feet (Phase 1)", Icon: "inv_boots_cloth_05",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 11, stats.Intellect: 9, stats.Spirit: 8, stats.SpellPower: 39, stats.Armor: 70}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
		SetName:        "Warlock Synthetic (Phase 1)",
	},
	// Finger, item level 66, in place of Ring of Spell Power (66, 19147).
	{
		Id: 997011, Name: "Warlock Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_38",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Stamina: 7, stats.SpellPower: 23}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Finger II, item level 66, in place of a second Molten Core ring (66).
	{
		Id: 997012, Name: "Warlock Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_15",
		Type:  proto.ItemType_ItemTypeFinger,
		Stats: stats.Stats{stats.Stamina: 7, stats.Intellect: 8, stats.Spirit: 7, stats.SpellPower: 27}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Main Hand, item level 71, in place of Azuresong Mageblade (71, 17103).
	{
		Id: 997013, Name: "Warlock Synthetic Main Hand (Phase 1)", Icon: "inv_sword_39",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeMainHand,
		Stats:      stats.Stats{stats.Intellect: 5, stats.SpellPower: 17}.ToFloatArray(),
		CritRating: 14,
		// Azuresong Mageblade damage and speed.
		WeaponDamageMin: 64, WeaponDamageMax: 140, WeaponSpeed: 2.4,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Off Hand, item level 70, in place of Fire Runed Grimoire (70, 19142).
	{
		Id: 997014, Name: "Warlock Synthetic Off Hand (Phase 1)", Icon: "inv_misc_book_09",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeOffHand, HandType: proto.HandType_HandTypeOffHand,
		Stats: stats.Stats{stats.Stamina: 6, stats.Intellect: 7, stats.Spirit: 5, stats.SpellPower: 37}.ToFloatArray(),
		Ilvl:  70, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Wand, item level 66, in place of Crimson Shocker (63, 17077).
	{
		Id: 997015, Name: "Warlock Synthetic Wand (Phase 1)", Icon: "inv_wand_09",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeWand,
		Stats: stats.Stats{stats.Stamina: 4, stats.Intellect: 4, stats.SpellPower: 19}.ToFloatArray(),
		// Crimson Shocker damage and speed.
		WeaponDamageMin: 102, WeaponDamageMax: 191, WeaponSpeed: 2,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
}
