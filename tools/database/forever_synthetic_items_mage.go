package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The synthetic level 60 Phase 1 mage sets, made the same way as the enhancement set in
// forever_synthetic_items.go (read that comment first). There are two sets. The main set
// (996001 to 996016) serves Frostfire, Fire and Frost. The Arcane set (996017 to 996030) is the
// same slots with Arcane's own stat mix. The Era Phase 1 BiS was one set for all three specs
// (Arcanist, Robe of Volatile Power, Mana Igniting Cord, Netherwind Pants, Azuresong Mageblade,
// Fire Runed Grimoire), none of it school specific, so both sets carry plain spell power and no
// school power.
//
// Each piece is priced on the cross-class standards' Forever budget (orchestration/standards.md
// item 1b on the orchestration branch). Weights: 1 a point of Stamina, Intellect and Spirit, 0.86 a
// point of spell power, 1 a point of hit and crit rating (10 rating is 1% hit, 14 rating 1% crit).
// Each armor piece's weighted linear total is slot mod x L(ilvl) x f, within 1%, with L = 1.125 x
// ilvl + 2.0, slot mods 1 (head, chest, legs), 0.75 (shoulders, hands, waist, feet) and 0.5625
// (neck, back, wrist, rings), and the caster f 1.053 / 1.026 / 0.995 for the three tiers. Every
// class takes one item level table: head, shoulders, chest, hands, legs and feet at 66, neck,
// back, wrist, waist and both rings at 71. Each armor piece also passes the shape check: its
// 1.7-norm, (sum (weight x amount)^1.7)^(1/1.7), is within 6% of slot mod x (0.661 x ilvl +
// 2.92), which takes the four stat lines a real caster epic carries. Weapons are Forever's caster
// weapon rate: the staff (item level 76) 193.4 (2.21 x L(76)), or a main hand 149.1 plus an off hand 43.4, and
// the wand 21.3.
//
// Per slot, the budget and what the stats add up to (main set, then Arcane), and the shape
// check's target and the pieces' 1.7-norms:
//   head 80.3 (79.9, 79.9), norm 46.5 (48.2, 48.2); neck 45.8 (45.6, 45.6), norm 28.0 (27.7,
//   27.7); shoulders 58.7 (58.7, 58.8), norm 34.9 (35.6, 35.4); back 45.8 (45.6, 45.6), norm
//   28.0 (27.7, 27.7); chest 80.3 (80.4, 80.1), norm 46.5 (48.2, 48.7); wrist 45.8 (45.8, 46.1),
//   norm 28.0 (27.5, 27.2); hands 58.7 (58.9, 58.9), norm 34.9 (35.9, 35.5); waist 63.0 (62.6,
//   62.6), norm 37.4 (38.0, 38.5); legs 80.3 (80.1, 80.0), norm 46.5 (48.5, 47.1); feet 58.7
//   (58.9, 58.4), norm 34.9 (35.2, 34.3); finger 45.8 (45.6, 45.6), norm 28.0 (27.9, 27.9);
//   second finger 45.8 (45.9, 45.9), norm 28.0 (27.2, 27.5); staff 193.4 (193.7, 193.8); wand
//   21.3 (21.2, 21.2); main hand 149.1 (149.6); off hand 43.4 (43.2).
// Armor 708.2 (main) and 707.6 (Arcane), weapons 214.8 and 215.0. For example the main set's
// head: 10 + 13 + 29 * 0.86 + 32 = 79.9 against 80.3, and (10^1.7 + 13^1.7 + 24.9^1.7 +
// 32^1.7)^(1/1.7) = 48.2 against 46.5.
//
// The presets use the staff with Forever's 2H Spellblasting (+65 spell power). The main hand and
// off hand split the main staff's stats between them. Weapon damage and speed are the Era items'
// (Azuresong, Staff of Dominance, Bonecreeper Stylus) and are not charged to the budget.
//
// The stat mix follows the standards' fair mix rule (item 5a): a realistic spread in the
// proportions Season of Discovery's level 60 BiS mage sets carry (orchestration/gear_targets.md),
// with each spec's own SoD Stamina share and Spirit at about 10%. Shares are linear: a stat's
// weighted amount over the set's total (the twelve armor slots, the staff and the wand; enchants
// and trinkets left out). Hit is at the cap (item 4), 10.8% of the set where SoD's sets carry
// about 3% (SoD gave mages hit from runes), so the other stats sit below their targets in about
// the table's proportions, spell power taking the most of the rest:
//
//   set      share     Sta   Int   Spi   SP    Hit   Crit
//   main     target    13.9  18.1  10    41.7  3.2   13.1
//   main     actual    12.9  16.4  9.5   39.2  10.8  11.2   (119, 151, 88, 421 SP, 100, 103)
//   Arcane   target    14.2  17.7  10    45.4  3.4   8.9    (and 0.4 MP5, left out)
//   Arcane   actual    13.1  15.9  9.6   42.9  10.8  7.6    (121, 147, 89, 460 SP, 100, 70)
//
// Arcane's staff carries more of the set's spell power (130), so the armor pieces keep four even
// enough lines for the shape check.
//
// Two synthetic trinkets (item 8a: Forever's real Phase 1 trinkets are gone or not simulated) at
// item level 71, the same for both sets: 41 spell power and 11 crit rating (46.3, 1.7-norm 38.1),
// and an on-use +140 spell power for 20 sec every 2 min.
//
// Hit stays at the cap against a level 63 boss (16% spell miss) in both sets: 100 hit rating
// (10%), with 5% from Elemental Precision or Arcane Focus and Manaflare's 2 piece 1%.
//
// Only the head and chest carry the Manaflare Regalia set (Forever's level 60 mage set, see
// sim/mage/item_sets_pve.go), in both sets, so only its 2 piece bonus is live, the same for
// every class's preset. The generator script lived in the session scratchpad (rb/b1.py), the
// numbers below are its output.

func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, foreverSyntheticMageItems...)
}

var foreverSyntheticMageItems = []*proto.UIItem{
	// Head, item level 66, in place of Arcanist Crown (66, 16795).
	{
		Id: 996001, Name: "Mage Synthetic Head (Phase 1)", Icon: "inv_helmet_53",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 10, stats.Intellect: 13, stats.SpellPower: 29, stats.Armor: 83}.ToFloatArray(),
		HitRating: 32,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
		SetName:        "Mage Synthetic (Phase 1)",
	},
	// Neck, item level 71, in place of Choker of the Fire Lord (78, 18814).
	{
		Id: 996002, Name: "Mage Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_17",
		Type:       proto.ItemType_ItemTypeNeck,
		Stats:      stats.Stats{stats.Stamina: 7, stats.Intellect: 9, stats.SpellPower: 24}.ToFloatArray(),
		CritRating: 9,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Shoulder, item level 66, in place of Arcanist Mantle (66, 16797).
	{
		Id: 996003, Name: "Mage Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_02",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 9, stats.Intellect: 12, stats.SpellPower: 31, stats.Armor: 76}.ToFloatArray(),
		CritRating: 11,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Back, item level 71.
	{
		Id: 996004, Name: "Mage Synthetic Back (Phase 1)", Icon: "inv_misc_cape_16",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 7, stats.Intellect: 9, stats.SpellPower: 24, stats.Armor: 54}.ToFloatArray(),
		CritRating: 9,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Chest, item level 66, in place of Robe of Volatile Power (66, 19145).
	{
		Id: 996005, Name: "Mage Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_03",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 12, stats.Intellect: 16, stats.Spirit: 18, stats.SpellPower: 40, stats.Armor: 102}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
		SetName:        "Mage Synthetic (Phase 1)",
	},
	// Wrist, item level 71, in place of Arcanist Bindings (66, 16799).
	{
		Id: 996006, Name: "Mage Synthetic Wrist (Phase 1)", Icon: "inv_bracer_09",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 7, stats.Intellect: 9, stats.Spirit: 10, stats.SpellPower: 23, stats.Armor: 44}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Hands, item level 66, in place of Arcanist Gloves (66, 16801).
	{
		Id: 996007, Name: "Mage Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_14",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 7, stats.Intellect: 9, stats.SpellPower: 22, stats.Armor: 63}.ToFloatArray(),
		HitRating: 24,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Waist, item level 71, in place of Mana Igniting Cord (71, 19136).
	{
		Id: 996008, Name: "Mage Synthetic Waist (Phase 1)", Icon: "inv_belt_22",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 8, stats.SpellPower: 24, stats.Armor: 61}.ToFloatArray(),
		HitRating:  25,
		CritRating: 9,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Legs, item level 66, in place of Netherwind Pants (76, 16915).
	{
		Id: 996009, Name: "Mage Synthetic Legs (Phase 1)", Icon: "inv_pants_08",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 13, stats.Intellect: 16, stats.SpellPower: 42, stats.Armor: 101}.ToFloatArray(),
		CritRating: 15,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Feet, item level 66, in place of Arcanist Boots (66, 16800).
	{
		Id: 996010, Name: "Mage Synthetic Feet (Phase 1)", Icon: "inv_boots_07",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 9, stats.Intellect: 12, stats.Spirit: 13, stats.SpellPower: 29, stats.Armor: 70}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Finger, item level 71, in place of Ring of Spell Power (66, 19147).
	{
		Id: 996011, Name: "Mage Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_38",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Stamina: 5, stats.SpellPower: 17}.ToFloatArray(),
		HitRating:  19,
		CritRating: 7,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Finger II, item level 71, in place of Cauterizing Band (71, 19140).
	{
		Id: 996012, Name: "Mage Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_39",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Intellect: 9, stats.Spirit: 10, stats.SpellPower: 22}.ToFloatArray(),
		CritRating: 8,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Main Hand, item level 71, in place of Azuresong Mageblade (71, 17103).
	{
		Id: 996013, Name: "Mage Synthetic Main Hand (Phase 1)", Icon: "inv_sword_39",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeMainHand,
		Stats:      stats.Stats{stats.Stamina: 25, stats.Intellect: 24, stats.Spirit: 27, stats.SpellPower: 60}.ToFloatArray(),
		CritRating: 22,
		// Azuresong Mageblade damage and speed.
		WeaponDamageMin: 64, WeaponDamageMax: 140, WeaponSpeed: 2.4,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Off Hand, item level 71, in place of Fire Runed Grimoire (70, 19142).
	{
		Id: 996014, Name: "Mage Synthetic Off Hand (Phase 1)", Icon: "inv_misc_book_09",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeOffHand, HandType: proto.HandType_HandTypeOffHand,
		Stats:      stats.Stats{stats.Intellect: 8, stats.Spirit: 10, stats.SpellPower: 20}.ToFloatArray(),
		CritRating: 8,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Wand, item level 71, in place of Bonecreeper Stylus (62, 13938).
	{
		Id: 996015, Name: "Mage Synthetic Wand (Phase 1)", Icon: "inv_wand_07",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeWand,
		Stats:      stats.Stats{stats.Intellect: 5, stats.SpellPower: 13}.ToFloatArray(),
		CritRating: 5,
		// Bonecreeper Stylus damage and speed.
		WeaponDamageMin: 83, WeaponDamageMax: 155, WeaponSpeed: 1.9,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Staff, item level 76, Forever's caster staff rate (193.4, 2.21 x L(76)), in place of Staff of
	// Dominance (70, 18842).
	{
		Id: 996016, Name: "Mage Synthetic Staff (Phase 1)", Icon: "inv_staff_13",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand,
		Stats:      stats.Stats{stats.Stamina: 25, stats.Intellect: 32, stats.Spirit: 37, stats.SpellPower: 81}.ToFloatArray(),
		CritRating: 30,
		// Staff of Dominance damage and speed.
		WeaponDamageMin: 126, WeaponDamageMax: 205, WeaponSpeed: 2.9,
		Ilvl: 76, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: head, item level 66, in place of Arcanist Crown (66, 16795).
	{
		Id: 996017, Name: "Mage Arcane Synthetic Head (Phase 1)", Icon: "inv_helmet_53",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 10, stats.Intellect: 13, stats.SpellPower: 29, stats.Armor: 83}.ToFloatArray(),
		HitRating: 32,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
		SetName:        "Mage Synthetic (Phase 1)",
	},
	// Arcane: neck, item level 71, in place of Choker of the Fire Lord (78, 18814).
	{
		Id: 996018, Name: "Mage Arcane Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_17",
		Type:       proto.ItemType_ItemTypeNeck,
		Stats:      stats.Stats{stats.Stamina: 8, stats.Intellect: 10, stats.SpellPower: 24}.ToFloatArray(),
		CritRating: 7,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: shoulder, item level 66, in place of Arcanist Mantle (66, 16797).
	{
		Id: 996019, Name: "Mage Arcane Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_02",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 11, stats.Intellect: 13, stats.SpellPower: 30, stats.Armor: 76}.ToFloatArray(),
		CritRating: 9,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: back, item level 71.
	{
		Id: 996020, Name: "Mage Arcane Synthetic Back (Phase 1)", Icon: "inv_misc_cape_16",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 8, stats.Intellect: 10, stats.SpellPower: 24, stats.Armor: 54}.ToFloatArray(),
		CritRating: 7,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: chest, item level 66, in place of Robe of Volatile Power (66, 19145).
	{
		Id: 996021, Name: "Mage Arcane Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_03",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 14, stats.Intellect: 18, stats.SpellPower: 42, stats.Armor: 102}.ToFloatArray(),
		CritRating: 12,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
		SetName:        "Mage Synthetic (Phase 1)",
	},
	// Arcane: wrist, item level 71, in place of Arcanist Bindings (66, 16799).
	{
		Id: 996022, Name: "Mage Arcane Synthetic Wrist (Phase 1)", Icon: "inv_bracer_09",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 7, stats.Intellect: 9, stats.Spirit: 12, stats.SpellPower: 21, stats.Armor: 44}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: hands, item level 66, in place of Arcanist Gloves (66, 16801).
	{
		Id: 996023, Name: "Mage Arcane Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_14",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:     stats.Stats{stats.Stamina: 8, stats.Intellect: 9, stats.SpellPower: 22, stats.Armor: 63}.ToFloatArray(),
		HitRating: 23,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: waist, item level 71, in place of Mana Igniting Cord (71, 19136).
	{
		Id: 996024, Name: "Mage Arcane Synthetic Waist (Phase 1)", Icon: "inv_belt_22",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Stamina: 9, stats.SpellPower: 24, stats.Armor: 61}.ToFloatArray(),
		HitRating:  26,
		CritRating: 7,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: legs, item level 66, in place of Netherwind Pants (76, 16915).
	{
		Id: 996025, Name: "Mage Arcane Synthetic Legs (Phase 1)", Icon: "inv_pants_08",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 12, stats.Intellect: 16, stats.Spirit: 21, stats.SpellPower: 36, stats.Armor: 101}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: feet, item level 66, in place of Arcanist Boots (66, 16800).
	{
		Id: 996026, Name: "Mage Arcane Synthetic Feet (Phase 1)", Icon: "inv_boots_07",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 9, stats.Intellect: 12, stats.Spirit: 15, stats.SpellPower: 26, stats.Armor: 70}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: finger, item level 71, in place of Ring of Spell Power (66, 19147).
	{
		Id: 996027, Name: "Mage Arcane Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_38",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Stamina: 7, stats.SpellPower: 17}.ToFloatArray(),
		HitRating:  19,
		CritRating: 5,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: finger II, item level 71, in place of Cauterizing Band (71, 19140).
	{
		Id: 996028, Name: "Mage Arcane Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_39",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Intellect: 9, stats.Spirit: 12, stats.SpellPower: 22}.ToFloatArray(),
		CritRating: 6,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: staff, item level 76, Forever's caster staff rate (193.4, 2.21 x L(76)), in place of
	// Staff of Dominance (70, 18842).
	{
		Id: 996029, Name: "Mage Arcane Synthetic Staff (Phase 1)", Icon: "inv_staff_13",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand,
		Stats:      stats.Stats{stats.Stamina: 18, stats.Intellect: 22, stats.Spirit: 29, stats.SpellPower: 130}.ToFloatArray(),
		CritRating: 13,
		// Staff of Dominance damage and speed.
		WeaponDamageMin: 126, WeaponDamageMax: 205, WeaponSpeed: 2.9,
		Ilvl: 76, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Arcane: wand, item level 71, in place of Bonecreeper Stylus (62, 13938).
	{
		Id: 996030, Name: "Mage Arcane Synthetic Wand (Phase 1)", Icon: "inv_wand_07",
		Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeWand,
		Stats:      stats.Stats{stats.Intellect: 6, stats.SpellPower: 13}.ToFloatArray(),
		CritRating: 4,
		// Bonecreeper Stylus damage and speed.
		WeaponDamageMin: 83, WeaponDamageMax: 155, WeaponSpeed: 1.9,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// Trinket, item level 71, synthetic (standards 1b and 8a): spell power and crit, linear total
	// 46.3 (0.56 of L(71), 45.9), 1.7-norm 38.1 (36.8).
	{
		Id: 996031, Name: "Mage Synthetic Trinket (Phase 1)", Icon: "inv_jewelry_talisman_09",
		Type:       proto.ItemType_ItemTypeTrinket,
		Stats:      stats.Stats{stats.SpellPower: 41}.ToFloatArray(),
		CritRating: 11,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	// On-use trinket, item level 71, synthetic (standards 1b and 8a): no passive stats, Use: +140 spell
	// power for 20 sec, 2 min cooldown (sim/mage/forever_synthetic_trinkets.go). Over a 180 sec fight
	// that is 140 * 0.86 * 40 / 180 = 26.8 weighted points (27, Talisman of Ephemeral Power's size).
	{
		Id: 996032, Name: "Mage Synthetic Talisman (Phase 1)", Icon: "inv_misc_gem_pearl_04",
		Type: proto.ItemType_ItemTypeTrinket,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
}
