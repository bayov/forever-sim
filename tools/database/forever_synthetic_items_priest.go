package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Synthetic level 60 items for a Phase 1 priest under Forever, made the way the enhancement
// items in forever_synthetic_items.go are (read its comment), by the cross-class standard in
// orchestration/standards.md. Ids 995000-995999.
//
// The budget is standards 1b (Forever's budget for a pure caster). Weights: 1 for Stamina,
// Intellect and Spirit, 0.86 for spell power, 1 for each point of hit rating (10 a percent) and
// crit rating (14 a percent); armor is free. With L(ilvl) = 1.125 * ilvl + 2.0, each armor
// piece's weighted linear total is slot * L * f within 1%: slot 1 / 0.75 / 0.5625 and f 1.053 /
// 1.026 / 0.995 for head, chest and legs / shoulders, hands, waist and feet / neck, back, wrist
// and rings, so 80.3 at 66, 58.7 at 66, 63.0 for the waist at 71 and 45.8 at 71. Its shape is
// checked on the 1.7-norm, (sum (weight * amount)^1.7)^(1/1.7) within 6% of slot * (0.661 *
// ilvl + 2.92), which holds every piece to four stat lines, as real caster epics have. Item
// levels: 66 for head, shoulders, chest, hands, legs and feet, 71 for neck, back, wrist, waist,
// rings, main hand, off hand, wand and trinkets, 76 for the staff. Weapons follow Forever's
// caster weapon rate: the staff 2.21 x L(76) = 193.4, or a main hand 1.82 x L(71) = 149.1 and
// an off hand 0.53 x L(71) = 43.4; the wand 0.26 x L(71) = 21.3. Each comment below gives the
// piece's linear total, what its rounded stats spend and its 1.7-norm against the check.
// Weapon and wand damage and speed are the real Era items' and are not charged to the budget.
//
// The pieces take the place of the Era Phase 1 items a shadow priest wears (the Molten Core
// tier, Choker of the Fire Lord, Mana Igniting Cord, Azuresong Mageblade and Fire Runed
// Grimoire, the Molten Core rings, Bonecreeper Stylus). Two synthetic trinkets stand in for
// Forever's launch trinkets, which the sim does not have (standards 8a): a stat trinket (36
// spell power, 15 crit rating) and an on-use trinket (140 spell power for 20 sec every 2 min).
//
// Stat mix (standards 5a, orchestration/gear_targets.md, bundled round: Season of Discovery's
// level 60 shadow priest BiS at the 1b weights, SoD's shadow-only power counted as generic spell
// power). A stat's share is weight * amount over the set's total, enchants and trinkets
// excluded. The 100 hit rating Shadow needs for the cap goes in first (standards 4), and
// Intellect, spell power and crit keep the table's proportions in the room left, which puts
// spell power 6.4 points under the table's own 43.4. Shares of the shipped sets (twelve armor
// pieces, the staff and the wand), in percent, and the set's amounts:
//
//	Set                 Sta   Int   Spi   SP    Hit   Crit
//	Shadow target       16.6  17.8  10    43.4  0.2   12.0   (SoD; hit goes to the cap)
//	Shadow (staff)      16.4  15.4  10.0  37.0  10.8  10.4   152 / 143 / 93 / 398 / 100 / 96
//	Shadow (MH + OH)    16.3  15.4  10.1  37.0  10.8  10.4
//	Smite target        16.6  17.8  10    43.5  0     12.1
//	Smite (staff)       16.6  17.8  10.0  43.5  0     12.2   153 / 164 / 92 / 467 / 0 / 113
//	Smite (MH + OH)     16.6  17.8  10.0  43.4  0     12.2
//
// Totals (linear, what the rounded stats spend): Shadow armor 710.9, staff 194.2, wand 21.2;
// Smite armor 708.2, staff 194.0, wand 21.5; each trinket 46.0 (the on-use one averages 26.8).
//
// The Shadow set (995001-995015, 995028) carries 100 hit rating (10%): with Shadow Focus 5/5
// (5%) and the 2 piece bonus (1%) that is the 16% a caster needs against a level 63 boss, and
// no more. The main hand and off hand carry the staff's 22 between them.
//
// The Smite set (995016-995027, 995029-995032, with its own rings and main hand) carries no hit
// rating: Holy Precision 3/3 gives Smite, Holy Fire and Penance 18%, so hit is worth nothing to
// it.
//
// Only the head and shoulders carry a set name, so only the 2 piece bonus is live: Raiments of
// Conviction (Forever's Tier 1 shadow set, 1% hit) on the Shadow set and Vestments of
// Conviction (Tier 1 Discipline and Holy, 10 Spirit) on the Smite set (see
// sim/priest/item_sets_pve.go). The 4 and 5 piece bonuses are measured apart in the presets.
//
// The main hand is a sword, like Azuresong Mageblade, so a Human gets their 2% crit with
// swords. A staff for each set (995028, 995029) takes the place of the main hand and off hand,
// for the 2H enchants; the shipped presets use the staff. The generator lived in the session
// scratchpad (synth7: design.py, emit.py); the numbers below are its output.

func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, foreverSyntheticPriestItems...)
}

var foreverSyntheticPriestItems = []*proto.UIItem{
	// Head, item level 66, in place of Circlet of Prophecy (66, 16813). Linear 80.3, spends 80.7; 1.7-norm 47.7 against 46.5 (+2.5%).
	{
		Id: 995001, Name: "Shadow Synthetic Head (Phase 1)", Icon: "inv_helmet_34",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 16, stats.Intellect: 19, stats.SpellPower: 38, stats.Armor: 83}.ToFloatArray(),
		HitRating: 13,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
		SetName:        "Shadow Synthetic (Phase 1)",
	},
	// Neck, item level 71, in place of Choker of the Fire Lord (78, 18814). Linear 45.8, spends 46.2; 1.7-norm 26.8 against 28.0 (-4.4%).
	{
		Id: 995002, Name: "Shadow Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_17",
		Type:       proto.ItemType_ItemTypeNeck,
		Stats:      stats.Stats{stats.Stamina: 9, stats.Intellect: 10, stats.SpellPower: 20}.ToFloatArray(),
		CritRating: 10,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Shoulder, item level 66, in place of Mantle of Prophecy (66, 16816). Linear 58.7, spends 58.2; 1.7-norm 34.4 against 34.9 (-1.5%).
	{
		Id: 995003, Name: "Shadow Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_02",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 12, stats.Intellect: 14, stats.SpellPower: 27, stats.Armor: 76}.ToFloatArray(),
		HitRating: 9,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
		SetName:        "Shadow Synthetic (Phase 1)",
	},
	// Back, item level 71, a Molten Core cloak. Linear 45.8, spends 45.9; 1.7-norm 27.3 against 28.0 (-2.8%).
	{
		Id: 995004, Name: "Shadow Synthetic Back (Phase 1)", Icon: "inv_misc_cape_16",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 10, stats.Spirit: 10, stats.SpellPower: 22, stats.Armor: 55}.ToFloatArray(),
		HitRating: 7,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Chest, item level 66, in place of Robe of Volatile Power (66, 19145). Linear 80.3, spends 80.1; 1.7-norm 46.6 against 46.5 (+0.1%).
	{
		Id: 995005, Name: "Shadow Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_08",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 15, stats.Intellect: 18, stats.SpellPower: 35, stats.Armor: 102}.ToFloatArray(),
		CritRating: 17,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Wrist, item level 71, in place of Vambraces of Prophecy (66, 16819). Linear 45.8, spends 45.9; 1.7-norm 27.3 against 28.0 (-2.6%).
	{
		Id: 995006, Name: "Shadow Synthetic Wrist (Phase 1)", Icon: "inv_bracer_09",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 9, stats.Intellect: 11, stats.SpellPower: 22, stats.Armor: 44}.ToFloatArray(),
		HitRating: 7,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Hands, item level 66, in place of Gloves of Prophecy (66, 16812). Linear 58.7, spends 59.2; 1.7-norm 34.9 against 34.9 (-0.1%).
	{
		Id: 995007, Name: "Shadow Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_14",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Intellect: 14, stats.SpellPower: 27, stats.Armor: 63}.ToFloatArray(),
		HitRating: 9, CritRating: 13,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Waist, item level 71, in place of Mana Igniting Cord (71, 19136). Linear 63.0, spends 62.9; 1.7-norm 36.9 against 37.4 (-1.2%).
	{
		Id: 995008, Name: "Shadow Synthetic Waist (Phase 1)", Icon: "inv_belt_08",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 12, stats.Spirit: 13, stats.SpellPower: 29, stats.Armor: 61}.ToFloatArray(),
		CritRating: 13,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Legs, item level 66, in place of Pants of Prophecy (66, 16814). Linear 80.3, spends 80.5; 1.7-norm 47.9 against 46.5 (+2.8%).
	{
		Id: 995009, Name: "Shadow Synthetic Legs (Phase 1)", Icon: "inv_pants_08",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 17, stats.Spirit: 17, stats.SpellPower: 39, stats.Armor: 89}.ToFloatArray(),
		HitRating: 13,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Feet, item level 66, in place of Boots of Prophecy (66, 16811). Linear 58.7, spends 59.1; 1.7-norm 34.9 against 34.9 (+0.0%).
	{
		Id: 995010, Name: "Shadow Synthetic Feet (Phase 1)", Icon: "inv_boots_cloth_05",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 12, stats.Spirit: 13, stats.SpellPower: 28, stats.Armor: 70}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Finger, item level 71, in place of Ring of Spell Power (66, 19147). Linear 45.8, spends 46.1; 1.7-norm 27.1 against 28.0 (-3.3%).
	{
		Id: 995011, Name: "Shadow Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_29",
		Type:      proto.ItemType_ItemTypeFinger,
		Stats:     stats.Stats{stats.Intellect: 11, stats.SpellPower: 21}.ToFloatArray(),
		HitRating: 7, CritRating: 10,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Finger II, item level 71, in place of Cauterizing Band (71, 19140). Linear 45.8, spends 46.1; 1.7-norm 27.0 against 28.0 (-3.8%).
	{
		Id: 995012, Name: "Shadow Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_36",
		Type:  proto.ItemType_ItemTypeFinger,
		Stats: stats.Stats{stats.Stamina: 9, stats.Intellect: 10, stats.Spirit: 9, stats.SpellPower: 21}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Main Hand, item level 71, in place of Azuresong Mageblade (71, 17103). Linear 149.0, spends 149.1.
	{
		Id: 995013, Name: "Shadow Synthetic Main Hand (Phase 1)", Icon: "inv_sword_39",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeMainHand,
		Stats:     stats.Stats{stats.Stamina: 21, stats.Intellect: 24, stats.Spirit: 22, stats.SpellPower: 49}.ToFloatArray(),
		HitRating: 17, CritRating: 23,
		// Azuresong Mageblade damage and speed.
		WeaponDamageMin: 64, WeaponDamageMax: 140, WeaponSpeed: 2.4,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Off Hand, item level 71, in place of Fire Runed Grimoire (70, 19142). Linear 43.4, spends 43.0.
	{
		Id: 995014, Name: "Shadow Synthetic Off Hand (Phase 1)", Icon: "inv_misc_book_09",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeOffHand, HandType: proto.HandType_HandTypeOffHand,
		Stats:     stats.Stats{stats.Stamina: 6, stats.Intellect: 7, stats.Spirit: 6, stats.SpellPower: 14}.ToFloatArray(),
		HitRating: 5, CritRating: 7,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Wand, item level 71, in place of Bonecreeper Stylus (62, 13938). Linear 21.3, spends 21.2.
	{
		Id: 995015, Name: "Shadow Synthetic Wand (Phase 1)", Icon: "inv_wand_07",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeWand,
		Stats:     stats.Stats{stats.Stamina: 3, stats.Intellect: 4, stats.Spirit: 3, stats.SpellPower: 6}.ToFloatArray(),
		HitRating: 3, CritRating: 3,
		// Bonecreeper Stylus damage, speed and school.
		SpellSchool:     proto.SpellSchool_SpellSchoolArcane,
		WeaponDamageMin: 83, WeaponDamageMax: 155, WeaponSpeed: 1.9,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Head for the Smite set, item level 66: head with the Smite proportions, no hit. Linear 80.3, spends 80.4; 1.7-norm 48.2 against 46.5 (+3.6%).
	{
		Id: 995016, Name: "Smite Synthetic Head (Phase 1)", Icon: "inv_helmet_34",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 15, stats.Intellect: 19, stats.SpellPower: 40, stats.Armor: 83}.ToFloatArray(),
		CritRating: 12,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
		SetName:        "Smite Synthetic (Phase 1)",
	},
	// Neck for the Smite set, item level 71: neck with the Smite proportions, no hit. Linear 45.8, spends 45.8; 1.7-norm 27.5 against 28.0 (-1.8%).
	{
		Id: 995022, Name: "Smite Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_17",
		Type:       proto.ItemType_ItemTypeNeck,
		Stats:      stats.Stats{stats.Stamina: 8, stats.Intellect: 11, stats.SpellPower: 23}.ToFloatArray(),
		CritRating: 7,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Shoulder for the Smite set, item level 66: shoulder with the Smite proportions, no hit. Linear 58.7, spends 58.8; 1.7-norm 35.6 against 34.9 (+1.9%).
	{
		Id: 995017, Name: "Smite Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_02",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 11, stats.Intellect: 14, stats.Spirit: 8, stats.SpellPower: 30, stats.Armor: 76}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
		SetName:        "Smite Synthetic (Phase 1)",
	},
	// Back for the Smite set, item level 71: back with the Smite proportions, no hit. Linear 45.8, spends 45.5; 1.7-norm 28.1 against 28.0 (+0.1%).
	{
		Id: 995023, Name: "Smite Synthetic Back (Phase 1)", Icon: "inv_misc_cape_16",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 10, stats.Spirit: 6, stats.SpellPower: 25, stats.Armor: 55}.ToFloatArray(),
		CritRating: 8,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Chest for the Smite set, item level 66: chest with the Smite proportions, no hit. Linear 80.3, spends 80.4; 1.7-norm 48.2 against 46.5 (+3.6%).
	{
		Id: 995018, Name: "Smite Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_08",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 15, stats.Intellect: 19, stats.SpellPower: 40, stats.Armor: 102}.ToFloatArray(),
		CritRating: 12,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Wrist for the Smite set, item level 71: wrist with the Smite proportions, no hit. Linear 45.8, spends 45.8; 1.7-norm 27.6 against 28.0 (-1.5%).
	{
		Id: 995024, Name: "Smite Synthetic Wrist (Phase 1)", Icon: "inv_bracer_09",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 9, stats.Intellect: 11, stats.Spirit: 6, stats.SpellPower: 23, stats.Armor: 44}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Hands for the Smite set, item level 66: hands with the Smite proportions, no hit. Linear 58.7, spends 58.7; 1.7-norm 36.0 against 34.9 (+3.0%).
	{
		Id: 995019, Name: "Smite Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_14",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Intellect: 15, stats.Spirit: 8, stats.SpellPower: 31, stats.Armor: 63}.ToFloatArray(),
		CritRating: 9,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Waist for the Smite set, item level 71: waist with the Smite proportions, no hit. Linear 63.0, spends 63.1; 1.7-norm 38.9 against 37.4 (+4.1%).
	{
		Id: 995025, Name: "Smite Synthetic Waist (Phase 1)", Icon: "inv_belt_08",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 13, stats.Spirit: 9, stats.SpellPower: 35, stats.Armor: 61}.ToFloatArray(),
		CritRating: 11,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Legs for the Smite set, item level 66: legs with the Smite proportions, no hit. Linear 80.3, spends 80.3; 1.7-norm 48.5 against 46.5 (+4.3%).
	{
		Id: 995020, Name: "Smite Synthetic Legs (Phase 1)", Icon: "inv_pants_08",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 15, stats.Intellect: 19, stats.Spirit: 11, stats.SpellPower: 41, stats.Armor: 89}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Feet for the Smite set, item level 66: feet with the Smite proportions, no hit. Linear 58.7, spends 58.4; 1.7-norm 36.3 against 34.9 (+3.9%).
	{
		Id: 995021, Name: "Smite Synthetic Feet (Phase 1)", Icon: "inv_boots_cloth_05",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 12, stats.Spirit: 8, stats.SpellPower: 33, stats.Armor: 70}.ToFloatArray(),
		CritRating: 10,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Off Hand for the Smite set, item level 71: off hand with the Smite proportions, no hit. Linear 43.4, spends 43.3.
	{
		Id: 995026, Name: "Smite Synthetic Off Hand (Phase 1)", Icon: "inv_misc_book_09",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeOffHand, HandType: proto.HandType_HandTypeOffHand,
		Stats:      stats.Stats{stats.Stamina: 7, stats.Intellect: 9, stats.Spirit: 5, stats.SpellPower: 19}.ToFloatArray(),
		CritRating: 6,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Wand for the Smite set, item level 71: wand with the Smite proportions, no hit. Linear 21.3, spends 21.5.
	{
		Id: 995027, Name: "Smite Synthetic Wand (Phase 1)", Icon: "inv_wand_07",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeWand,
		Stats:      stats.Stats{stats.Stamina: 3, stats.Intellect: 4, stats.Spirit: 2, stats.SpellPower: 11}.ToFloatArray(),
		CritRating: 3,
		// Bonecreeper Stylus damage, speed and school.
		SpellSchool:     proto.SpellSchool_SpellSchoolArcane,
		WeaponDamageMin: 83, WeaponDamageMax: 155, WeaponSpeed: 1.9,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Finger for the Smite set, item level 71: finger with the Smite proportions. Linear 45.8, spends 45.6; 1.7-norm 28.0 against 28.0 (-0.2%).
	{
		Id: 995030, Name: "Smite Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_29",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Intellect: 12, stats.Spirit: 6, stats.SpellPower: 24}.ToFloatArray(),
		CritRating: 7,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Finger II for the Smite set, item level 71: finger with the Smite proportions. Linear 45.8, spends 45.5; 1.7-norm 28.1 against 28.0 (+0.1%).
	{
		Id: 995031, Name: "Smite Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_36",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Stamina: 10, stats.Spirit: 6, stats.SpellPower: 25}.ToFloatArray(),
		CritRating: 8,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Main Hand for the Smite set, item level 71: main hand with the Smite proportions. Linear 149.0, spends 149.8.
	{
		Id: 995032, Name: "Smite Synthetic Main Hand (Phase 1)", Icon: "inv_sword_39",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeMainHand,
		Stats:      stats.Stats{stats.Stamina: 25, stats.Intellect: 31, stats.Spirit: 17, stats.SpellPower: 66}.ToFloatArray(),
		CritRating: 20,
		// Azuresong Mageblade damage and speed.
		WeaponDamageMin: 64, WeaponDamageMax: 140, WeaponSpeed: 2.4,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Staff for the Shadow set, item level 76, in place of the main hand and off hand: Forever's
	// caster staff rate, 2.21 x L(76) = 193.4 (standards 1b). Linear 193.4, spends 194.2.
	{
		Id: 995028, Name: "Shadow Synthetic Staff (Phase 1)", Icon: "inv_staff_13",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand,
		Stats:     stats.Stats{stats.Stamina: 28, stats.Intellect: 32, stats.Spirit: 28, stats.SpellPower: 63}.ToFloatArray(),
		HitRating: 22, CritRating: 30,
		// A caster staff's damage and speed, which do nothing for a priest.
		WeaponDamageMin: 105, WeaponDamageMax: 197, WeaponSpeed: 2.9,
		Ilvl: 76, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Staff for the Smite set, item level 76, in place of the main hand and off hand: Forever's
	// caster staff rate, 2.21 x L(76) = 193.4 (standards 1b). Linear 193.4, spends 194.0.
	{
		Id: 995029, Name: "Smite Synthetic Staff (Phase 1)", Icon: "inv_staff_13",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand,
		Stats:      stats.Stats{stats.Stamina: 32, stats.Intellect: 40, stats.Spirit: 22, stats.SpellPower: 86}.ToFloatArray(),
		CritRating: 26,
		// A caster staff's damage and speed, which do nothing for a priest.
		WeaponDamageMin: 105, WeaponDamageMax: 197, WeaponSpeed: 2.9,
		Ilvl: 76, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// Stat trinket, item level 71, for both priest specs (standards 1b and 8a: Forever's launch
	// trinkets are not simulated, so every Level 60 preset carries two synthetic ones). Linear
	// 0.56 x L(71) = 45.9, spends 46.0; 1.7-norm 36.0 against 36.8 (-2.2%).
	{
		Id: 995033, Name: "Priest Synthetic Trinket (Phase 1)", Icon: "inv_jewelry_talisman_08",
		Type:       proto.ItemType_ItemTypeTrinket,
		Stats:      stats.Stats{stats.SpellPower: 36}.ToFloatArray(),
		CritRating: 15,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	// On-use trinket, item level 71, no passive stats: 140 spell power for 20 sec every 2 min on
	// the shared offensive trinket cooldown (sim/priest/forever_synthetic_trinkets.go), 140 x 0.86
	// x 40 / 180 = 26.8 weighted points over a 180 sec fight, the size of Talisman of Ephemeral
	// Power (standards 1b).
	{
		Id: 995034, Name: "Priest Synthetic On-Use Trinket (Phase 1)", Icon: "inv_misc_gem_pearl_04",
		Type: proto.ItemType_ItemTypeTrinket,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
}
