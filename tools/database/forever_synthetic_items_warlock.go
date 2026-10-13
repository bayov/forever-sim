package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The synthetic Phase 1 warlock items, ids 997001 to 997999. See forever_synthetic_items.go for what these
// items are and why we need them.
//
// One set serves Affliction, Destruction and Destruction fire, and Demonology swaps in four pieces with
// more hit. All are casters that want spell power, hit up to the cap, crit, Intellect and Spirit, and the Era Phase 1 BiS lists give the three
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
// (7^1.5 + 12^1.5 + 14^1.5 + (40 * 0.857)^1.5)^(2/3) = 46.2, and the synthetic sword spends that on the
// set's stat mix below. The off
// hand is at 71 (27.5), so the two weapon slots come to about 73, what a caster's staff gets, and the
// wand is 0.32 of the value at 71 (15.6).
//
// The stats on a piece add up as (sum of (weight * amount)^1.5)^(2/3), where a point of Stamina,
// Intellect or Spirit weighs 1, a point of spell power 0.55 (Forever's tuned price) and a point of hit or
// crit rating 1 (10 hit rating is 1% hit, 14 crit rating 1% crit). Armor is free, it is the Era piece's.
//
// The stat mix is the fair one of the cross-class standards (item 5a, 2026-10-13): not all spell power
// but the proportions real level 60 warlock gear has, the Season of Discovery level 60 BiS sets' mix
// with Stamina at 18% and Spirit at 10% (orchestration/gear_targets.md). A stat's linear share (L) is its
// weight times its amount over the set's weighted total, enchants left out. Every piece carries every stat
// in the target proportions, scaled to its budget, with hit scaled so the set lands on the cap; the
// generator script lived in the session scratchpad (rebud/gen.py), the numbers below are its output,
// rounded, with spell power filling what rounding left of each budget. Per slot, budget and what the
// stats add up to: head, chest and legs 42.8 (42.6 to 42.8), shoulders, hands and feet 33.0 (33.0), waist
// 37.6 (37.3 to 37.6), neck, back, wrist and fingers 27.5 (27.2), staff 73.7 (73.5 to 73.6), wand 15.6
// (15.3), main hand 46.2 (45.9), off hand 27.5 (27.2).
//
// The shared set (Destruction, Affliction, Destruction fire; level60_phase1_synthetic) comes to 236
// spell power, 163 Intellect, 161 Stamina, 90 Spirit, 114 crit rating (8.1%) and 100 hit rating (10%):
//
//	Stat     Amount  L share  Target (Warlock)
//	Stamina    161    18.6      18 +/- 1.5
//	Intellect  163    18.9      18.8
//	Spirit      90    10.4      10
//	Spell pwr  236    27.3      27.7
//	Hit        100    11.6      12.6
//	Crit       114    13.2      12.9
//
// Demonology's set (level60_phase1_synthetic_demonology, the four hit pieces at the end of the list in
// place of the chest, legs, waist and staff) comes to 221 spell power, 148 Intellect, 156 Stamina, 86
// Spirit, 115 crit rating (8.2%) and 140 hit rating (14%):
//
//	Stat     Amount  L share  Target (Warlock Demonology)
//	Stamina    156    18.0      18 +/- 1.5
//	Intellect  148    17.1      15.4
//	Spirit      86     9.9      10
//	Spell pwr  221    25.5      24.6
//	Hit        140    16.2      17.3
//	Crit       115    13.3      14.7
//
// Before the rebudget the shared set was 524 spell power, 124 Intellect, 128 Stamina, 76 Spirit, 70 crit
// rating and 100 hit rating (spell power 36.7%, crit 8.9%).
//
//   - Hit is at the cap against a level 63 boss (17% miss, 1% of it can't be removed, so 16% hit counts):
//     10% here, 5% from Suppression and 1% from the 2 piece. Every Level 60 build but Demonology takes
//     Suppression 5/5. Demonology keeps one point (its talent search puts the other four in Agonizing
//     Flames and Improved Corruption) and wears the four hit pieces, for 14% hit from gear.
//   - Spirit is about 10% of the set. Under Forever it also adds a point of mana to every Life Tap, on top
//     of the regen.

// Only the head and chest carry the set, the bonuses of Demonheart Raiment, Forever's warlock Tier 1 set
// (item set 2100, see sim/warlock/item_sets_pve.go), so only its 2 piece bonus (1% hit) is live, the same
// for every class's preset.
//
// The presets wear the staff at the end of the list (997020), the main hand and off hand budgets together,
// with Forever's Enchant 2H Weapon - Spellblasting; it beats the sword (Azuresong Mageblade, which gave a
// Human warlock Sword Specialization) and off hand by 23 to 33 DPS on every build. The sword and off hand
// stay for anyone who wants to compare.

func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, warlockSyntheticItems...)
}

var warlockSyntheticItems = []*proto.UIItem{
	// Head, item level 66, in place of Felheart Horns (66, 16808).
	{
		Id: 997001, Name: "Warlock Synthetic Head (Phase 1)", Icon: "inv_helmet_58",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 14, stats.Intellect: 14, stats.Spirit: 8, stats.SpellPower: 37, stats.Armor: 83}.ToFloatArray(),
		HitRating:  9,
		CritRating: 10,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
		SetName:        "Warlock Synthetic (Phase 1)",
	},
	// Neck, item level 71, in place of Choker of the Fire Lord (78, 18814).
	{
		Id: 997002, Name: "Warlock Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_17",
		Type:       proto.ItemType_ItemTypeNeck,
		Stats:      stats.Stats{stats.Stamina: 9, stats.Intellect: 9, stats.Spirit: 5, stats.SpellPower: 25}.ToFloatArray(),
		HitRating:  5,
		CritRating: 6,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Shoulder, item level 66, in place of Felheart Shoulder Pads (66, 16807).
	{
		Id: 997003, Name: "Warlock Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_23",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 11, stats.Intellect: 11, stats.Spirit: 6, stats.SpellPower: 28, stats.Armor: 76}.ToFloatArray(),
		HitRating:  7,
		CritRating: 8,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Back, item level 71.
	{
		Id: 997004, Name: "Warlock Synthetic Back (Phase 1)", Icon: "inv_misc_cape_20",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 9, stats.Intellect: 9, stats.Spirit: 5, stats.SpellPower: 25, stats.Armor: 54}.ToFloatArray(),
		HitRating:  5,
		CritRating: 6,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Chest, item level 66, in place of Robe of Volatile Power (66, 19145).
	{
		Id: 997005, Name: "Warlock Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_09",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 14, stats.Intellect: 14, stats.Spirit: 8, stats.SpellPower: 37, stats.Armor: 102}.ToFloatArray(),
		HitRating:  9,
		CritRating: 10,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
		SetName:        "Warlock Synthetic (Phase 1)",
	},
	// Wrist, item level 71, in place of Felheart Bracers (66, 16804).
	{
		Id: 997006, Name: "Warlock Synthetic Wrist (Phase 1)", Icon: "inv_bracer_07",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 9, stats.Intellect: 9, stats.Spirit: 5, stats.SpellPower: 25, stats.Armor: 44}.ToFloatArray(),
		HitRating:  5,
		CritRating: 6,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Hands, item level 66, in place of Felheart Gloves (66, 16805).
	{
		Id: 997007, Name: "Warlock Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_19",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 11, stats.Intellect: 11, stats.Spirit: 6, stats.SpellPower: 28, stats.Armor: 63}.ToFloatArray(),
		HitRating:  7,
		CritRating: 8,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Waist, item level 71, in place of Mana Igniting Cord (71, 19136).
	{
		Id: 997008, Name: "Warlock Synthetic Waist (Phase 1)", Icon: "inv_belt_13",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 12, stats.Intellect: 13, stats.Spirit: 7, stats.SpellPower: 32, stats.Armor: 61}.ToFloatArray(),
		HitRating:  8,
		CritRating: 9,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Legs, item level 66, in place of Felheart Pants (66, 16810); Nemesis Leggings are Blackwing Lair's.
	{
		Id: 997009, Name: "Warlock Synthetic Legs (Phase 1)", Icon: "inv_pants_cloth_14",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 14, stats.Intellect: 14, stats.Spirit: 8, stats.SpellPower: 36, stats.Armor: 102}.ToFloatArray(),
		HitRating:  10,
		CritRating: 10,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Feet, item level 66, in place of Felheart Slippers (66, 16803).
	{
		Id: 997010, Name: "Warlock Synthetic Feet (Phase 1)", Icon: "inv_boots_cloth_05",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 11, stats.Intellect: 11, stats.Spirit: 6, stats.SpellPower: 28, stats.Armor: 70}.ToFloatArray(),
		HitRating:  7,
		CritRating: 8,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Finger, item level 71, in place of Ring of Spell Power (66, 19147).
	{
		Id: 997011, Name: "Warlock Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_38",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Stamina: 9, stats.Intellect: 9, stats.Spirit: 5, stats.SpellPower: 25}.ToFloatArray(),
		HitRating:  5,
		CritRating: 6,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Finger II, item level 71.
	{
		Id: 997012, Name: "Warlock Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_15",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Stamina: 9, stats.Intellect: 9, stats.Spirit: 5, stats.SpellPower: 25}.ToFloatArray(),
		HitRating:  5,
		CritRating: 6,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Main Hand, item level 71, in place of Azuresong Mageblade (71, 17103). Its budget is the real
	// sword's, see the file comment.
	{
		Id: 997013, Name: "Warlock Synthetic Main Hand (Phase 1)", Icon: "inv_sword_39",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeMainHand,
		Stats:      stats.Stats{stats.Stamina: 15, stats.Intellect: 16, stats.Spirit: 8, stats.SpellPower: 40}.ToFloatArray(),
		HitRating:  9,
		CritRating: 11,
		// Azuresong Mageblade damage and speed.
		WeaponDamageMin: 64, WeaponDamageMax: 140, WeaponSpeed: 2.4,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Off Hand, item level 71, in place of Fire Runed Grimoire (70, 19142).
	{
		Id: 997014, Name: "Warlock Synthetic Off Hand (Phase 1)", Icon: "inv_misc_book_09",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeOffHand, HandType: proto.HandType_HandTypeOffHand,
		Stats:      stats.Stats{stats.Stamina: 9, stats.Intellect: 9, stats.Spirit: 5, stats.SpellPower: 25}.ToFloatArray(),
		HitRating:  5,
		CritRating: 6,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Wand, item level 71, in place of Crimson Shocker (63, 17077).
	{
		Id: 997015, Name: "Warlock Synthetic Wand (Phase 1)", Icon: "inv_wand_09",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeWand,
		Stats:      stats.Stats{stats.Stamina: 5, stats.Intellect: 5, stats.Spirit: 3, stats.SpellPower: 13}.ToFloatArray(),
		HitRating:  3,
		CritRating: 4,
		// Crimson Shocker damage and speed.
		WeaponDamageMin: 102, WeaponDamageMax: 191, WeaponSpeed: 2,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},

	// Demonology's hit pieces. Demonology keeps one point of Suppression (Agonizing Flames and Improved
	// Corruption are worth more than its other four), so its 2% from talents and the 2 piece leave it 4%
	// under the cap on the set above. These four take the place of the chest, legs, waist and staff, the
	// biggest budgets, built to Demonology's own stat mix with 40 more hit rating between them.
	// Chest, item level 66, the Demonology hit variant of 997005.
	{
		Id: 997016, Name: "Warlock Synthetic Chest of Precision (Phase 1)", Icon: "inv_chest_cloth_09",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 13, stats.Intellect: 11, stats.Spirit: 7, stats.SpellPower: 31, stats.Armor: 102}.ToFloatArray(),
		HitRating:  18,
		CritRating: 10,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
		SetName:        "Warlock Synthetic (Phase 1)",
	},
	// Legs, item level 66, the Demonology hit variant of 997009.
	{
		Id: 997017, Name: "Warlock Synthetic Legs of Precision (Phase 1)", Icon: "inv_pants_cloth_14",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 13, stats.Intellect: 11, stats.Spirit: 7, stats.SpellPower: 31, stats.Armor: 102}.ToFloatArray(),
		HitRating:  18,
		CritRating: 10,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Waist, item level 71, the Demonology hit variant of 997008.
	{
		Id: 997018, Name: "Warlock Synthetic Waist of Precision (Phase 1)", Icon: "inv_belt_13",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 11, stats.Intellect: 10, stats.Spirit: 6, stats.SpellPower: 26, stats.Armor: 61}.ToFloatArray(),
		HitRating:  16,
		CritRating: 9,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Staff, the Demonology hit variant of 997020 (73.7).
	{
		Id: 997019, Name: "Warlock Synthetic Staff of Precision (Phase 1)", Icon: "inv_staff_13",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand,
		Stats:      stats.Stats{stats.Stamina: 22, stats.Intellect: 19, stats.Spirit: 12, stats.SpellPower: 54}.ToFloatArray(),
		HitRating:  30,
		CritRating: 18,
		// Staff of Dominance damage and speed.
		WeaponDamageMin: 126, WeaponDamageMax: 205, WeaponSpeed: 2.9,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Staff, the main hand and off hand budgets together (46.2 + 27.5 = 73.7), in place of Staff of
	// Dominance (70, 18842).
	{
		Id: 997020, Name: "Warlock Synthetic Staff (Phase 1)", Icon: "inv_staff_13",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand,
		Stats:      stats.Stats{stats.Stamina: 24, stats.Intellect: 25, stats.Spirit: 13, stats.SpellPower: 65}.ToFloatArray(),
		HitRating:  15,
		CritRating: 17,
		// Staff of Dominance damage and speed.
		WeaponDamageMin: 126, WeaponDamageMax: 205, WeaponSpeed: 2.9,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
}
