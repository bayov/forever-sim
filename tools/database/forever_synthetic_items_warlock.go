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
// tier), Choker of the Fire Lord, Robe of Volatile Power, Mana Igniting Cord, Ring of Spell Power,
// Azuresong Mageblade, Fire Runed Grimoire and Crimson Shocker. The item levels are the cross-class
// standard's (orchestration/standards.md on the orchestration branch), not the Era items': head,
// shoulders, chest, hands, legs and feet at 66, neck, back, wrist, waist and both rings at 71, so the
// twelve armor slots cost the same for every spec.
//
// The budget is the one shared by every class's synthetic set (see forever_synthetic_items_shaman.go). An
// epic piece is worth V = 1.2 * item level - 36.4 (42.8 at 66, 48.8 at 71), times the slot modifier: 1
// for the head, chest and legs, 0.77 for the shoulders, hands, waist and feet, 0.5625 for the neck, back,
// wrist, fingers and the held in off hand, and 0.32 for a wand. A caster one-hander keeps the real Era
// weapon's budget (the orchestrator's ruling for every caster), its stats priced at Era rates with spell
// power at 0.857: Azuresong Mageblade's 7 Stamina, 12 Intellect, 40 spell power and 1% crit come to
// (7^1.5 + 12^1.5 + 14^1.5 + (40 * 0.857)^1.5)^(2/3) = 46.2, and the synthetic sword spends that at
// Forever's 0.55 a spell power: the same Stamina, Intellect and crit and 62 spell power (46.0). The off
// hand is at 71 (27.5), so the two weapon slots come to about 73, what a caster's staff gets, and the
// wand is 0.32 of the value at 71 (15.6).
//
// The stats on a piece add up as (sum of (weight * amount)^1.5)^(2/3), where a point of Stamina,
// Intellect or Spirit weighs 1, a point of spell power 0.55 (Forever's tuned price) and a point of hit or
// crit rating 1 (10 hit rating is 1% hit, 14 crit rating 1% crit). Armor is free, it is the Era piece's.
// We pick the Stamina, Intellect, Spirit, hit and crit of each piece, and spell power fills the rest. For
// the head: 42.8, and 12 Stamina, 13 Intellect, 9 Spirit, 10 hit and 14 crit rating leave room for 34
// spell power: (12^1.5 + 13^1.5 + 9^1.5 + 10^1.5 + 14^1.5 + (34 * 0.55)^1.5)^(2/3) = 42.8. Per slot,
// budget and the points the stats below add up to: head 42.8 (42.8), neck 27.5 (27.2), shoulders 33.0
// (32.9), back 27.5 (27.3), chest 42.8 (42.8), wrist 27.5 (27.5), hands 33.0 (32.9), waist 37.6 (37.4),
// legs 42.8 (42.7), feet 33.0 (33.1), finger 27.5 (27.6), second finger 27.5 (27.3), main hand 46.2
// (46.0), off hand 27.5 (27.4), wand 15.6 (15.5). The generator script lived in the session scratchpad
// (synth.py), the numbers below are its output.
//
// The set comes to 524 spell power, 124 Intellect, 128 Stamina, 76 Spirit, 70 crit rating (5%) and 100
// hit rating (10%):
//   - Hit is at the cap against a level 63 boss (17% miss, 1% of it can't be removed, so 16% hit counts):
//     10% here, 5% from Suppression and 1% from the 2 piece. Every Level 60 build but Demonology takes
//     Suppression 5/5. Demonology keeps one point (its talent search on this set puts the other four
//     in Agonizing Flames and Improved Corruption) and wears the four hit pieces at the end of the list,
//     997016 to 997019, for 14% hit from gear (level60_phase1_synthetic_demonology).
//   - Spirit is on 10 of the 15 pieces, about 10% of the set. Under Forever it also adds a point of mana
//     to every Life Tap, on top of the regen.
//   - The crit sits where the Era pieces had it: 1% each on the head, back, chest (Robe of Volatile
//     Power), the first ring and the sword (Azuresong Mageblade).
//
// Only the head and chest carry the set, the bonuses of Demonheart Raiment, Forever's warlock Tier 1 set
// (item set 2100, see sim/warlock/item_sets_pve.go), so only its 2 piece bonus (1% hit) is live, the same
// for every class's preset.
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
		Stats:      stats.Stats{stats.Stamina: 12, stats.Intellect: 13, stats.Spirit: 9, stats.SpellPower: 34, stats.Armor: 83}.ToFloatArray(),
		HitRating:  10,
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
		SetName:        "Warlock Synthetic (Phase 1)",
	},
	// Neck, item level 71, in place of Choker of the Fire Lord (78, 18814).
	{
		Id: 997002, Name: "Warlock Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_17",
		Type:      proto.ItemType_ItemTypeNeck,
		Stats:     stats.Stats{stats.Stamina: 6, stats.Intellect: 7, stats.SpellPower: 33}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Shoulder, item level 66, in place of Felheart Shoulder Pads (66, 16807).
	{
		Id: 997003, Name: "Warlock Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_23",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 9, stats.Intellect: 9, stats.Spirit: 7, stats.SpellPower: 35, stats.Armor: 76}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Back, item level 71.
	{
		Id: 997004, Name: "Warlock Synthetic Back (Phase 1)", Icon: "inv_misc_cape_20",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 7, stats.Spirit: 6, stats.SpellPower: 27, stats.Armor: 54}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Chest, item level 66, in place of Robe of Volatile Power (66, 19145).
	{
		Id: 997005, Name: "Warlock Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_09",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 12, stats.Intellect: 14, stats.Spirit: 9, stats.SpellPower: 41, stats.Armor: 102}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
		SetName:        "Warlock Synthetic (Phase 1)",
	},
	// Wrist, item level 71, in place of Felheart Bracers (66, 16804).
	{
		Id: 997006, Name: "Warlock Synthetic Wrist (Phase 1)", Icon: "inv_bracer_07",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 8, stats.Intellect: 8, stats.SpellPower: 30, stats.Armor: 44}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Hands, item level 66, in place of Felheart Gloves (66, 16805).
	{
		Id: 997007, Name: "Warlock Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_19",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 9, stats.Intellect: 9, stats.Spirit: 8, stats.SpellPower: 34, stats.Armor: 63}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Waist, item level 71, in place of Mana Igniting Cord (71, 19136).
	{
		Id: 997008, Name: "Warlock Synthetic Waist (Phase 1)", Icon: "inv_belt_13",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 10, stats.Intellect: 10, stats.Spirit: 7, stats.SpellPower: 43, stats.Armor: 61}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Legs, item level 66, in place of Felheart Pants (66, 16810); Nemesis Leggings are Blackwing Lair's.
	{
		Id: 997009, Name: "Warlock Synthetic Legs (Phase 1)", Icon: "inv_pants_cloth_14",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 14, stats.Intellect: 14, stats.Spirit: 10, stats.SpellPower: 42, stats.Armor: 102}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Feet, item level 66, in place of Felheart Slippers (66, 16803).
	{
		Id: 997010, Name: "Warlock Synthetic Feet (Phase 1)", Icon: "inv_boots_cloth_05",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 10, stats.Intellect: 9, stats.Spirit: 8, stats.SpellPower: 33, stats.Armor: 70}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Finger, item level 71, in place of Ring of Spell Power (66, 19147).
	{
		Id: 997011, Name: "Warlock Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_38",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Stamina: 7, stats.SpellPower: 32}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Finger II, item level 71.
	{
		Id: 997012, Name: "Warlock Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_15",
		Type:      proto.ItemType_ItemTypeFinger,
		Stats:     stats.Stats{stats.Stamina: 7, stats.Intellect: 8, stats.Spirit: 7, stats.SpellPower: 25}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Main Hand, item level 71, in place of Azuresong Mageblade (71, 17103). Its budget is the real
	// sword's, see the file comment.
	{
		Id: 997013, Name: "Warlock Synthetic Main Hand (Phase 1)", Icon: "inv_sword_39",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeMainHand,
		Stats:      stats.Stats{stats.Stamina: 7, stats.Intellect: 12, stats.SpellPower: 62}.ToFloatArray(),
		CritRating: 14,
		// Azuresong Mageblade damage and speed.
		WeaponDamageMin: 64, WeaponDamageMax: 140, WeaponSpeed: 2.4,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Off Hand, item level 71, in place of Fire Runed Grimoire (70, 19142).
	{
		Id: 997014, Name: "Warlock Synthetic Off Hand (Phase 1)", Icon: "inv_misc_book_09",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeOffHand, HandType: proto.HandType_HandTypeOffHand,
		Stats:     stats.Stats{stats.Stamina: 6, stats.Intellect: 7, stats.Spirit: 5, stats.SpellPower: 30}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Wand, item level 71, in place of Crimson Shocker (63, 17077).
	{
		Id: 997015, Name: "Warlock Synthetic Wand (Phase 1)", Icon: "inv_wand_09",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeWand,
		Stats: stats.Stats{stats.Stamina: 4, stats.Intellect: 4, stats.SpellPower: 23}.ToFloatArray(),
		// Crimson Shocker damage and speed.
		WeaponDamageMin: 102, WeaponDamageMax: 191, WeaponSpeed: 2,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},

	// Demonology's hit pieces. Demonology keeps one point of Suppression (Agonizing Flames and Improved
	// Corruption are worth more than its other four), so its 2% from talents and the 2 piece leave it 4%
	// under the cap on the set above. These four take the place of the chest, back, first ring and wand
	// with 10 hit rating each and less spell power at the same budget.
	// Chest, item level 66, the Demonology hit variant of 997005 (42.7 of 42.8).
	{
		Id: 997016, Name: "Warlock Synthetic Chest of Precision (Phase 1)", Icon: "inv_chest_cloth_09",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 12, stats.Intellect: 14, stats.Spirit: 9, stats.SpellPower: 32, stats.Armor: 102}.ToFloatArray(),
		HitRating:  10,
		CritRating: 14,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
		SetName:        "Warlock Synthetic (Phase 1)",
	},
	// Back, item level 71, the Demonology hit variant of 997004 (27.4 of 27.5).
	{
		Id: 997017, Name: "Warlock Synthetic Back of Precision (Phase 1)", Icon: "inv_misc_cape_20",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 7, stats.Spirit: 6, stats.SpellPower: 16, stats.Armor: 54}.ToFloatArray(),
		HitRating:  10,
		CritRating: 14,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Finger, item level 71, the Demonology hit variant of 997011 (27.6 of 27.5).
	{
		Id: 997018, Name: "Warlock Synthetic Finger of Precision (Phase 1)", Icon: "inv_jewelry_ring_38",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Stamina: 7, stats.SpellPower: 22}.ToFloatArray(),
		HitRating:  10,
		CritRating: 14,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Wand, item level 71, the Demonology hit variant of 997015 (15.7 of 15.6).
	{
		Id: 997019, Name: "Warlock Synthetic Wand of Precision (Phase 1)", Icon: "inv_wand_09",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeWand,
		Stats:     stats.Stats{stats.Stamina: 4, stats.Intellect: 4, stats.SpellPower: 11}.ToFloatArray(),
		HitRating: 10,
		// Crimson Shocker damage and speed.
		WeaponDamageMin: 102, WeaponDamageMax: 191, WeaponSpeed: 2,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Staff, the main hand and off hand budgets together (46.2 + 27.5 = 73.7), in place of Staff of
	// Dominance (70, 18842), with the off hand's 10 hit rating so the set stays at the cap (73.7).
	{
		Id: 997020, Name: "Warlock Synthetic Staff (Phase 1)", Icon: "inv_staff_13",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand,
		Stats:      stats.Stats{stats.Stamina: 13, stats.Intellect: 19, stats.Spirit: 5, stats.SpellPower: 100}.ToFloatArray(),
		HitRating:  10,
		CritRating: 14,
		// Staff of Dominance damage and speed.
		WeaponDamageMin: 126, WeaponDamageMax: 205, WeaponSpeed: 2.9,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
}
