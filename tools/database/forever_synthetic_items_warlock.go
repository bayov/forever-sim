package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The synthetic Phase 1 warlock items, ids 997001 to 997999. See forever_synthetic_items.go for what these
// items are and why we need them.
//
// One set serves Affliction, Destruction and Destruction fire, and Demonology swaps in four pieces with
// more hit. All are casters that want spell power, hit up to the cap, crit, Intellect and Spirit, and the
// Era Phase 1 BiS lists give them the same gear (the Felheart set, Robe of Volatile Power, Mana Igniting Cord, Choker of the Fire Lord,
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
// The budget is Forever's, per archetype (standards 1b, 2026-10-13). A stat's weight is 1 for Stamina,
// Intellect, Spirit and a point of hit or crit rating (10 hit rating is 1% hit, 14 crit rating 1% crit),
// 0.86 for a point of spell power; armor is free, it is the Era piece's. The line is L = 1.125 * item
// level + 2.0, and a pure caster's armor piece carries slot mod * L * f in linear weighted stats, within
// 1%: slot mods 1 (head, chest, legs), 0.75 (shoulders, hands, waist, feet) and 0.5625 (neck, back,
// wrist, fingers), f 1.053, 1.026 and 0.995 for the three. Each piece also has to pass the shape check:
// its 1.7-norm value, (sum of (weight * amount)^1.7)^(1/1.7), within 6% of slot mod * (0.661 * item level
// + 2.92), which holds a caster to about 4 stat lines a piece. Forever doubles caster weapons: the staff
// carries 193.4 (2.21 L at 76), a main hand 149.1 and a held off hand 43.4, a wand 21.3.
//
// The stat mix is the fair one of the cross-class standards (item 5a): the Season of Discovery level 60
// warlock BiS sets' proportions at these weights (orchestration/gear_targets.md, da0a4c79e), Stamina at
// the warlock's own share, Spirit 10%. Hit goes to the cap first (item 4); the room it takes past the
// table's share came evenly off Intellect, spell power and crit; the targets below are the hit-adjusted
// ones (gear_targets.md 1a6968081: the table's Intellect, spell power and crit re-scaled into the room
// the capped hit leaves). A stat's linear share (L) is its weight
// times its amount over the set's weighted total, weapons in, enchants and trinkets out. The generator
// script lived in the session scratchpad (r3/gen.py); it picks each piece's stat lines, fits the amounts
// to the set's totals, and the numbers below are its output, rounded, with spell power taking what
// rounding left. Per piece:
//
//	Piece          Linear (target)   1.7-norm (target)
//	head            80.5 ( 80.3)     47.8 (46.5, +2.7%)
//	chest           80.1 ( 80.3)     46.9 (46.5, +0.7%)
//	legs            80.1 ( 80.3)     46.9 (46.5, +0.7%)
//	shoulder        59.1 ( 58.7)     34.9 (34.9, +0.0%)
//	hands           58.5 ( 58.7)     34.0 (34.9, -2.5%)
//	waist           62.9 ( 63.0)     37.2 (37.4, -0.6%)
//	feet            58.4 ( 58.7)     34.3 (34.9, -1.8%)
//	neck            45.9 ( 45.8)     27.2 (28.0, -3.0%)
//	back            46.1 ( 45.8)     27.0 (28.0, -3.7%)
//	wrist           45.9 ( 45.8)     27.2 (28.0, -3.0%)
//	finger          46.2 ( 45.8)     27.0 (28.0, -3.9%)
//	finger II       45.9 ( 45.8)     27.2 (28.0, -3.0%)
//	chest (Demo)    79.8 ( 80.3)     48.1 (46.5, +3.3%)
//	legs (Demo)     80.7 ( 80.3)     46.5 (46.5, -0.1%)
//	waist (Demo)    63.2 ( 63.0)     36.9 (37.4, -1.4%)
//
// The shared set (Destruction, Affliction, Destruction fire; level60_phase1_synthetic): armor 709.6,
// staff 193.6, wand 21.3, 412 spell power, 148 Intellect, 133 Stamina, 93 Spirit, 96 crit rating (6.9%)
// and 100 hit rating (10%):
//
//	Stat     Amount  L share  Target (Warlock, hit-adjusted)
//	Stamina    133    14.4      14.2 +/- 1.5
//	Intellect  148    16.0      16.3 (table 17.4)
//	Spirit      93    10.1      10
//	Spell pwr  412    38.3      37.5 (table 40.0)
//	Hit        100    10.8      at the cap (SoD 6.4)
//	Crit        96    10.4      11.2 (table 12.0)
//
// Demonology's old set (level60_phase1_synthetic_demonology, for its one point of Suppression; no Level 60 preset
// wears it since Demonology took Suppression 5/5 on 2026-10-24 (standards 26); the four hit pieces at the end of the list in
// place of the chest, legs, waist and staff): armor 710.2, staff 193.6, wand 21.3, 392 spell power, 119
// Intellect, 125 Stamina, 92 Spirit, 112 crit rating (8.0%) and 140 hit rating (14%):
//
//	Stat     Amount  L share  Target (Warlock Demonology, hit-adjusted)
//	Stamina    125    13.5      13.5 +/- 1.5
//	Intellect  119    12.9      13.8 (table 15.5)
//	Spirit      92     9.9      10
//	Spell pwr  392    36.4      34.5 (table 38.8)
//	Hit        140    15.1      at the cap (SoD 7.4)
//	Crit       112    12.1      13.2 (table 14.8)
//
// Both sets wear the two synthetic trinkets at the end of the list (outside the shares).
//
// Earlier sets: 524 spell power, 124 Intellect, 128 Stamina, 76 Spirit, 5% crit (all spell power, to
// 2026-10-13), then the 1a set of 434 spell power, 161 Intellect, 7.9% crit (1.73 times the old budget).
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
// The presets wear the staff (997020, item level 76) with Forever's Enchant 2H Weapon - Spellblasting. On
// the first set it beat the sword (Azuresong Mageblade, which gave a Human warlock Sword Specialization)
// and off hand by 23 to 33 DPS on every build; both are at Forever's caster weapon rate now and stay for
// anyone who wants to compare.

func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, warlockSyntheticItems...)
}

var warlockSyntheticItems = []*proto.UIItem{
	// Head, item level 66, in place of Felheart Horns (66, 16808).
	{
		Id: 997001, Name: "Warlock Synthetic Head (Phase 1)", Icon: "inv_helmet_58",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 14, stats.Intellect: 17, stats.SpellPower: 39, stats.Armor: 83}.ToFloatArray(),
		CritRating: 16,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
		SetName:        "Warlock Synthetic (Phase 1)",
	},
	// Neck, item level 71, in place of Choker of the Fire Lord (78, 18814).
	{
		Id: 997002, Name: "Warlock Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_17",
		Type:       proto.ItemType_ItemTypeNeck,
		Stats:      stats.Stats{stats.Stamina: 8, stats.Intellect: 10, stats.SpellPower: 22}.ToFloatArray(),
		CritRating: 9,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Shoulder, item level 66, in place of Felheart Shoulder Pads (66, 16807).
	{
		Id: 997003, Name: "Warlock Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_23",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 10, stats.Intellect: 13, stats.SpellPower: 28, stats.Armor: 76}.ToFloatArray(),
		CritRating: 12,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Back, item level 71.
	{
		Id: 997004, Name: "Warlock Synthetic Back (Phase 1)", Icon: "inv_misc_cape_20",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 8, stats.Intellect: 10, stats.Spirit: 10, stats.SpellPower: 21, stats.Armor: 54}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Chest, item level 66, in place of Robe of Volatile Power (66, 19145).
	{
		Id: 997005, Name: "Warlock Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_09",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 13, stats.Intellect: 16, stats.SpellPower: 35, stats.Armor: 102}.ToFloatArray(),
		HitRating: 21,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
		SetName:        "Warlock Synthetic (Phase 1)",
	},
	// Wrist, item level 71, in place of Felheart Bracers (66, 16804).
	{
		Id: 997006, Name: "Warlock Synthetic Wrist (Phase 1)", Icon: "inv_bracer_07",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 8, stats.Intellect: 10, stats.SpellPower: 22, stats.Armor: 44}.ToFloatArray(),
		CritRating: 9,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Hands, item level 66, in place of Felheart Gloves (66, 16805).
	{
		Id: 997007, Name: "Warlock Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_19",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Intellect: 11, stats.Spirit: 11, stats.SpellPower: 25, stats.Armor: 63}.ToFloatArray(),
		HitRating: 15,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Waist, item level 71, in place of Mana Igniting Cord (71, 19136).
	{
		Id: 997008, Name: "Warlock Synthetic Waist (Phase 1)", Icon: "inv_belt_13",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 10, stats.Spirit: 12, stats.SpellPower: 29, stats.Armor: 61}.ToFloatArray(),
		HitRating: 16,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Legs, item level 66, in place of Felheart Pants (66, 16810); Nemesis Leggings are Blackwing Lair's.
	{
		Id: 997009, Name: "Warlock Synthetic Legs (Phase 1)", Icon: "inv_pants_cloth_14",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 13, stats.Spirit: 16, stats.SpellPower: 35, stats.Armor: 102}.ToFloatArray(),
		HitRating: 21,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Feet, item level 66, in place of Felheart Slippers (66, 16803).
	{
		Id: 997010, Name: "Warlock Synthetic Feet (Phase 1)", Icon: "inv_boots_cloth_05",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 9, stats.Intellect: 12, stats.SpellPower: 26, stats.Armor: 70}.ToFloatArray(),
		HitRating: 15,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Finger, item level 71, in place of Ring of Spell Power (66, 19147).
	{
		Id: 997011, Name: "Warlock Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_38",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Intellect: 9, stats.SpellPower: 20}.ToFloatArray(),
		HitRating:  12,
		CritRating: 8,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Finger II, item level 71.
	{
		Id: 997012, Name: "Warlock Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_15",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Stamina: 8, stats.Spirit: 10, stats.SpellPower: 22}.ToFloatArray(),
		CritRating: 9,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Main Hand, item level 71, in place of Azuresong Mageblade (71, 17103), Forever's caster main hand
	// rate (149.1).
	{
		Id: 997013, Name: "Warlock Synthetic Main Hand (Phase 1)", Icon: "inv_sword_39",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeMainHand,
		Stats:      stats.Stats{stats.Stamina: 27, stats.Intellect: 30, stats.SpellPower: 84}.ToFloatArray(),
		CritRating: 20,
		// Azuresong Mageblade damage and speed.
		WeaponDamageMin: 64, WeaponDamageMax: 140, WeaponSpeed: 2.4,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Off Hand, item level 71, in place of Fire Runed Grimoire (70, 19142), Forever's held off hand rate
	// (43.4).
	{
		Id: 997014, Name: "Warlock Synthetic Off Hand (Phase 1)", Icon: "inv_misc_book_09",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeOffHand, HandType: proto.HandType_HandTypeOffHand,
		Stats: stats.Stats{stats.Intellect: 11, stats.Spirit: 7, stats.SpellPower: 30}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Wand, item level 71, in place of Crimson Shocker (63, 17077), 0.26 of the line (21.3).
	{
		Id: 997015, Name: "Warlock Synthetic Wand (Phase 1)", Icon: "inv_wand_09",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeWand,
		Stats: stats.Stats{stats.Stamina: 5, stats.Intellect: 6, stats.SpellPower: 12}.ToFloatArray(),
		// Crimson Shocker damage, speed and school.
		SpellSchool:     proto.SpellSchool_SpellSchoolFire,
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
		Stats:     stats.Stats{stats.Stamina: 17, stats.Intellect: 8, stats.SpellPower: 37, stats.Armor: 102}.ToFloatArray(),
		HitRating: 23,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
		SetName:        "Warlock Synthetic (Phase 1)",
	},
	// Legs, item level 66, the Demonology hit variant of 997009.
	{
		Id: 997017, Name: "Warlock Synthetic Legs of Precision (Phase 1)", Icon: "inv_pants_cloth_14",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Spirit: 21, stats.SpellPower: 31, stats.Armor: 102}.ToFloatArray(),
		HitRating:  20,
		CritRating: 13,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Waist, item level 71, the Demonology hit variant of 997008.
	{
		Id: 997018, Name: "Warlock Synthetic Waist of Precision (Phase 1)", Icon: "inv_belt_13",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 12, stats.SpellPower: 27, stats.Armor: 61}.ToFloatArray(),
		HitRating:  17,
		CritRating: 11,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Staff, item level 76, the Demonology hit variant of 997020 (193.4).
	{
		Id: 997019, Name: "Warlock Synthetic Staff of Precision (Phase 1)", Icon: "inv_staff_13",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand,
		Stats:      stats.Stats{stats.Stamina: 26, stats.Intellect: 13, stats.Spirit: 40, stats.SpellPower: 60}.ToFloatArray(),
		HitRating:  38,
		CritRating: 25,
		// Staff of Dominance damage and speed.
		WeaponDamageMin: 126, WeaponDamageMax: 205, WeaponSpeed: 2.9,
		Ilvl: 76, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Staff, item level 76, Forever's caster staff rate (193.4), in place of Staff of Dominance (70, 18842).
	{
		Id: 997020, Name: "Warlock Synthetic Staff (Phase 1)", Icon: "inv_staff_13",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand,
		Stats:      stats.Stats{stats.Stamina: 27, stats.Intellect: 34, stats.Spirit: 34, stats.SpellPower: 76}.ToFloatArray(),
		CritRating: 33,
		// Staff of Dominance damage and speed.
		WeaponDamageMin: 126, WeaponDamageMax: 205, WeaponSpeed: 2.9,
		Ilvl: 76, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	// Trinkets, item level 71, synthetic (standards 1b and 8a): Forever's launch trinkets are not in the sim
	// and the Era Phase 1 ones are raid drops or gone from the client. The stat trinket is 0.56 of the
	// line at 71 in linear stats (45.7 of 45.9) on two lines (1.7-norm 36.8). The on-use trinket has no
	// stats; its 140 spell power for 20 sec every 2 min (sim/warlock/forever_synthetic_trinkets.go, on
	// the shared offensive trinket cooldown) averages 26.8 weighted points over a 180 sec fight, the
	// size of Talisman of Ephemeral Power.
	{
		Id: 997021, Name: "Warlock Synthetic Trinket (Phase 1)", Icon: "inv_jewelry_talisman_08",
		Type:       proto.ItemType_ItemTypeTrinket,
		Stats:      stats.Stats{stats.SpellPower: 38}.ToFloatArray(),
		CritRating: 13,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	{
		Id: 997022, Name: "Warlock Synthetic Talisman (Phase 1)", Icon: "inv_misc_stonetablet_11",
		Type: proto.ItemType_ItemTypeTrinket,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
}
