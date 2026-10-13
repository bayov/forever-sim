package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The druid's synthetic Phase 1 items, made like the enhancement shaman's (forever_synthetic_items.go,
// read its comment first), ids 998000 to 998999. One set for Balance and one for Feral cat, because
// the two share no stat but Stamina.
//
// Each piece takes the place of an Era Phase 1 BiS item the spec wears (named above the piece) and is
// priced by standards 1b (orchestration/standards.md on the shared branch), the budget Forever's own
// items and real Era epics follow. Item levels are one table for every class: 66 for the head,
// shoulders, chest, hands, legs and feet, 71 for the neck, back, wrist, waist and rings, 76 for the
// two-handers. Stats are weighted 1 for Strength, Agility, Stamina, Intellect and Spirit, 0.86 for
// spell power, 0.5 for attack power, and 1 for a point of hit or crit rating (10 hit and 14 crit
// rating to the percent); armor is free. Then:
//
//	L(item level) = 1.125 * item level + 2.0 (76.25 at 66, 81.875 at 71, 87.5 at 76)
//	weighted linear stats of an armor piece = slot mod * L * f, within 1%
//	slot mods: 1 head / chest / legs, 0.75 shoulder / hands / waist / feet, 0.5625 neck / back / wrist / rings
//	f, Balance (caster): 1.053 / 1.026 / 0.995; Feral (physical): 0.950 / 0.959 / 0.911
//	shape check: (sum (weight * amount)^1.7)^(1/1.7) within 6% of slot mod * (0.661 * item level + 2.92)
//
// The shape check pins the number of stat lines per piece to the archetype's real count: 4 on every
// Balance piece, 3 on every Feral piece. The staff is at Forever's caster weapon rate, 2.21 * L(76) =
// 193.4, and the feral two-hander at 0.90 * L(76) = 78.8 with no attack power. Per armor piece
// (head, neck, shoulders, back, chest, wrist, hands, waist, legs, feet, rings), the weighted linear
// total and the shape value against its target:
//
//	Balance  80.2  45.6  58.6  45.6  80.2  45.6  58.6  63.4  80.7  59.1  46.1  45.6   (armor 709.5 of 709.0)
//	  shape  -0.2  -1.4  +1.8  -1.3  -0.2  -1.2  +0.7  +1.1  +1.6  +0.0  +2.9  -1.2 %
//	Feral    72.0  42.0  55.0  42.0  72.0  42.0  55.0  59.0  72.0  55.0  42.0  42.0   (armor 650.0 of 650.5)
//	  shape  -0.9  -1.6  +0.5  -2.9  -0.9  -3.9  +0.5  +0.5  -1.6  +2.8  -4.3  -3.2 %
//
// The stat mix follows standards 5a, the bundled-round targets of orchestration/gear_targets.md
// (Season of Discovery level 60 BiS proportions at the 1b weights, each spec's own SoD Stamina share,
// Spirit at 10% for a mana user). Hit is set first (item 4) and the other stats keep the table's
// proportions in what is left. Linear share (L) = weight * amount / the set's weighted total,
// weapon included, enchants and trinkets excluded:
//
//	Balance  L %   Sta 15.1  Int 16.3  Spi  9.9  SP 35.7  Hit 12.2  Crit 10.9
//	  target       Sta 15.0  Int 17.9  Spi 10    SP 39.3  Hit  5.9  Crit 11.9
//	Feral    L %   Sta 19.6  Str 24.3  Agi 25.9  Int 6.2  AP 2.2  Hit 12.3  Crit  9.5
//	  target       Sta 19.7  Str 26.9  Agi 28.5  Int 6.7  AP 2.4  Hit  5.5  Crit 10.4
//
// Balance (leather, ids 998001 to 998013) totals 375 spell power, 147 Intellect, 136 Stamina, 110 hit
// rating, 98 crit rating (7.0%) and 89 Spirit. The hit takes the moonkin to the spell hit cap against
// a level 63 boss and no further: 16% with Nature's Reach (4%) and the Grovekeeper 2 piece (1%). Hit at
// the cap leaves spell power 3.6 points under its target (the SoD sets carry less hit). Moving hit into
// spell power alone would be 1.5 to 1.9 ahead, but a set with 50 less hit and the rest in the table's
// proportions is 8.7 behind (814.0 against 822.7 at 180 sec) and the Wrath-only build 5.9 behind, so
// the cap stays. Spell power is the all-school kind.
//
// Feral cat (leather, ids 998101 to 998113) totals 189 Agility, 177 Strength, 143 Stamina, 90 hit
// rating (9%, the cat's special attack hit cap against a level 63 boss; moving 10 of it into Agility
// ties, into crit is 1.2 behind), 69 crit rating (4.9%), 45 Intellect and 32 attack power (on a ring).
// The cat fights on Energy, so no Spirit (the cross-class rule for Energy and Rage users), but Shifting
// Power, which replaced powershifting, costs mana (55% of base mana for 40 Energy).
//
// Each spec also gets two synthetic item level 71 trinkets (standards 1b and 8a, in place of real
// Phase 1 trinkets, which Forever removed or the sim does not model): a stat trinket with a weighted
// linear total of 45.9 and a shape value of 36.8 (Balance 36 spell power and 15 crit rating: 45.96,
// -2.2%; Feral 29 Agility and 17 Strength: 46.0, -3.8%), and an on-use trinket with no stats whose
// effect (sim/druid/forever_synthetic_trinkets.go) is worth 27 (spell power) or 37 (physical) weighted
// points over a 3 minute fight: 140 spell power, or 165 Strength, for 20 sec every 2 min.
//
// The head and chest of each spec, and only those two, form a set with the bonuses of Grovekeeper
// Eclipse (Balance) or Grovekeeper Ferocity (Feral), Forever's crafted level 60 druid sets
// (sim/druid/item_sets_pve.go), so only the 2 piece bonus is live (1% hit, or 1% attack and casting
// speed), the same shape as every class's set. The weapons keep the damage and speed of the Era weapon
// they replace (a cat's weapon damage does nothing in Cat Form, only its stats count), and the armor is
// that of the Molten Core leather tier (Cenarion Raiment) for the slot. The relic slot takes a real
// Forever Phase 1 idol, Swarming Idol for Balance and Howling Idol for Feral (sim/druid/items.go), which counts as
// worth its slot through its effect, so no stats go elsewhere for it (standards 1a). The
// generator lived in the session scratchpad (druidsets5.py), the numbers below are its output.

func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, foreverSyntheticDruidItems...)
}

var foreverSyntheticDruidItems = []*proto.UIItem{
	// Head, item level 66, in place of Cenarion Helm (66, 16834).
	{
		Id: 998001, Name: "Balance Synthetic Head (Phase 1)", Icon: "inv_helmet_09",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Stamina: 17, stats.Intellect: 18, stats.SpellPower: 34, stats.Armor: 163}.ToFloatArray(),
		CritRating: 16,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
		SetName:        "Balance Synthetic (Phase 1)",
	},
	// Neck, item level 71, in place of Choker of the Fire Lord (78, 18814).
	{
		Id: 998002, Name: "Balance Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_17",
		Type:       proto.ItemType_ItemTypeNeck,
		Stats:      stats.Stats{stats.Stamina: 8, stats.Intellect: 9, stats.SpellPower: 24}.ToFloatArray(),
		CritRating: 8,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Shoulder, item level 66, in place of Wild Growth Spaulders (71, 18810).
	{
		Id: 998003, Name: "Balance Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_07",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Stamina: 9, stats.Spirit: 7, stats.SpellPower: 24, stats.Armor: 159}.ToFloatArray(),
		HitRating: 22,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Back, item level 71.
	{
		Id: 998004, Name: "Balance Synthetic Back (Phase 1)", Icon: "inv_misc_cape_16",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 9, stats.Intellect: 9, stats.Spirit: 7, stats.SpellPower: 24, stats.Armor: 55}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Chest, item level 66, in place of Robe of Volatile Power (66, 19145).
	{
		Id: 998005, Name: "Balance Synthetic Chest (Phase 1)", Icon: "inv_chest_cloth_06",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Stamina: 17, stats.Intellect: 18, stats.Spirit: 16, stats.SpellPower: 34, stats.Armor: 200}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
		SetName:        "Balance Synthetic (Phase 1)",
	},
	// Wrist, item level 71, in place of Cenarion Bracers (66, 16830).
	{
		Id: 998006, Name: "Balance Synthetic Wrist (Phase 1)", Icon: "inv_bracer_03",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Intellect: 10, stats.Spirit: 7, stats.SpellPower: 24, stats.Armor: 88}.ToFloatArray(),
		CritRating: 8,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Hands, item level 66, in place of Cenarion Gloves (66, 16831).
	{
		Id: 998007, Name: "Balance Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_08",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Intellect: 9, stats.SpellPower: 24, stats.Armor: 125}.ToFloatArray(),
		HitRating: 21, CritRating: 8,
		Ilvl: 66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Waist, item level 71, in place of Mana Igniting Cord (71, 19136).
	{
		Id: 998008, Name: "Balance Synthetic Waist (Phase 1)", Icon: "inv_belt_06",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Stamina: 10, stats.SpellPower: 26, stats.Armor: 113}.ToFloatArray(),
		HitRating: 22, CritRating: 9,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Legs, item level 66, in place of Cenarion Leggings (66, 16835).
	{
		Id: 998009, Name: "Balance Synthetic Legs (Phase 1)", Icon: "inv_pants_06",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Stamina: 13, stats.Intellect: 14, stats.SpellPower: 31, stats.Armor: 175}.ToFloatArray(),
		HitRating: 27,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Feet, item level 66, in place of Cenarion Boots (66, 16829).
	{
		Id: 998010, Name: "Balance Synthetic Feet (Phase 1)", Icon: "inv_boots_08",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Stamina: 12, stats.Intellect: 13, stats.Spirit: 10, stats.SpellPower: 28, stats.Armor: 138}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Finger, item level 71, in place of Ring of Spell Power (66, 19147).
	{
		Id: 998011, Name: "Balance Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_25",
		Type:      proto.ItemType_ItemTypeFinger,
		Stats:     stats.Stats{stats.Stamina: 5, stats.SpellPower: 21}.ToFloatArray(),
		HitRating: 18, CritRating: 5,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Finger II, item level 71.
	{
		Id: 998012, Name: "Balance Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_29",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Intellect: 10, stats.Spirit: 7, stats.SpellPower: 24}.ToFloatArray(),
		CritRating: 8,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Two-Hand, item level 76, in place of Staff of Dominance (70, 18842).
	{
		Id: 998013, Name: "Balance Synthetic Two-Hand (Phase 1)", Icon: "inv_staff_13",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand,
		Stats:      stats.Stats{stats.Stamina: 36, stats.Intellect: 37, stats.Spirit: 35, stats.SpellPower: 57}.ToFloatArray(),
		CritRating: 36,
		// Staff of Dominance damage and speed.
		WeaponDamageMin: 126, WeaponDamageMax: 205, WeaponSpeed: 2.9,
		Ilvl: 76, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Trinket, item level 71, stat trinket, standards 1b.
	{
		Id: 998014, Name: "Balance Synthetic Stat Trinket (Phase 1)", Icon: "inv_jewelry_talisman_04",
		Type:       proto.ItemType_ItemTypeTrinket,
		Stats:      stats.Stats{stats.SpellPower: 36}.ToFloatArray(),
		CritRating: 15,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Trinket, item level 71, on-use trinket, standards 1b: no passive stats, the effect is in sim/druid/forever_synthetic_trinkets.go.
	{
		Id: 998015, Name: "Balance Synthetic Spell Power Trinket (Phase 1)", Icon: "inv_misc_gem_pearl_04",
		Type: proto.ItemType_ItemTypeTrinket,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Head, item level 66, in place of Wolfshead Helm (45, 8345).
	{
		Id: 998101, Name: "Feral Synthetic Head (Phase 1)", Icon: "inv_helmet_04",
		Type: proto.ItemType_ItemTypeHead, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Strength: 22, stats.Agility: 21, stats.Stamina: 29, stats.Armor: 163}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
		SetName:        "Feral Synthetic (Phase 1)",
	},
	// Neck, item level 71, in place of Onyxia Tooth Pendant (74, 18404).
	{
		Id: 998102, Name: "Feral Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_09",
		Type:      proto.ItemType_ItemTypeNeck,
		Stats:     stats.Stats{stats.Agility: 12, stats.Stamina: 20}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Shoulder, item level 66, in place of Truestrike Shoulders (61, 12927).
	{
		Id: 998103, Name: "Feral Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_25",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Strength: 20, stats.Agility: 19, stats.Armor: 150}.ToFloatArray(),
		HitRating: 16,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Back, item level 71, in place of Cape of the Black Baron (63, 13340).
	{
		Id: 998104, Name: "Feral Synthetic Back (Phase 1)", Icon: "inv_misc_cape_08",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats:      stats.Stats{stats.Agility: 10, stats.Stamina: 18, stats.Armor: 55}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Chest, item level 66, in place of Cadaverous Armor (61, 14637).
	{
		Id: 998105, Name: "Feral Synthetic Chest (Phase 1)", Icon: "inv_chest_leather_08",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Strength: 22, stats.Agility: 21, stats.Stamina: 29, stats.Armor: 200}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
		SetName:        "Feral Synthetic (Phase 1)",
	},
	// Wrist, item level 71, in place of Wristguards of Stability (65, 19146).
	{
		Id: 998106, Name: "Feral Synthetic Wrist (Phase 1)", Icon: "inv_bracer_07",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Strength: 17, stats.Intellect: 12, stats.Armor: 88}.ToFloatArray(),
		HitRating: 13,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Hands, item level 66, in place of Devilsaur Gauntlets (58, 15063).
	{
		Id: 998107, Name: "Feral Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_23",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Strength: 20, stats.Agility: 19, stats.Armor: 125}.ToFloatArray(),
		HitRating: 16,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Waist, item level 71, in place of Mugger's Belt (62, 18505).
	{
		Id: 998108, Name: "Feral Synthetic Waist (Phase 1)", Icon: "inv_belt_16",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:      stats.Stats{stats.Strength: 20, stats.Agility: 18, stats.Armor: 113}.ToFloatArray(),
		CritRating: 21,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Legs, item level 66, in place of Devilsaur Leggings (60, 15062).
	{
		Id: 998109, Name: "Feral Synthetic Legs (Phase 1)", Icon: "inv_pants_wolf",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats:     stats.Stats{stats.Strength: 25, stats.Agility: 24, stats.Armor: 175}.ToFloatArray(),
		HitRating: 23,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Feet, item level 66.
	{
		Id: 998110, Name: "Feral Synthetic Feet (Phase 1)", Icon: "inv_boots_08",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeLeather,
		Stats: stats.Stats{stats.Agility: 17, stats.Stamina: 25, stats.Intellect: 13, stats.Armor: 138}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Finger, item level 71, in place of Quick Strike Ring (67, 18821).
	{
		Id: 998111, Name: "Feral Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_15",
		Type:      proto.ItemType_ItemTypeFinger,
		Stats:     stats.Stats{stats.Agility: 14, stats.AttackPower: 32}.ToFloatArray(),
		HitRating: 12,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Finger II, item level 71, in place of Don Julio's Band (65, 19325).
	{
		Id: 998112, Name: "Feral Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_09",
		Type:       proto.ItemType_ItemTypeFinger,
		Stats:      stats.Stats{stats.Strength: 15, stats.Intellect: 10}.ToFloatArray(),
		CritRating: 17,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Two-Hand, item level 76, in place of Earthshaker (66, 17073).
	{
		Id: 998113, Name: "Feral Synthetic Two-Hand (Phase 1)", Icon: "inv_mace_25",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeMace, HandType: proto.HandType_HandTypeTwoHand,
		Stats:      stats.Stats{stats.Strength: 16, stats.Agility: 14, stats.Stamina: 22, stats.Intellect: 10}.ToFloatArray(),
		CritRating: 17,
		// Earthshaker damage and speed.
		WeaponDamageMin: 175, WeaponDamageMax: 263, WeaponSpeed: 3.5,
		Ilvl: 76, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Trinket, item level 71, stat trinket, standards 1b.
	{
		Id: 998114, Name: "Feral Synthetic Stat Trinket (Phase 1)", Icon: "inv_misc_monsterclaw_04",
		Type:  proto.ItemType_ItemTypeTrinket,
		Stats: stats.Stats{stats.Strength: 17, stats.Agility: 29}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Trinket, item level 71, on-use trinket, standards 1b: no passive stats, the effect is in sim/druid/forever_synthetic_trinkets.go.
	{
		Id: 998115, Name: "Feral Synthetic Strength Trinket (Phase 1)", Icon: "inv_misc_rune_07",
		Type: proto.ItemType_ItemTypeTrinket,
		Ilvl: 71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
}
