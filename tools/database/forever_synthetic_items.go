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
// the Scholomance Bloodmail set, Death's Clutch and Wristguards of True Flight), with Spinal Reaper's damage
// and speed on the two-hander. The budget follows the cross-class standard (standards item 1b on the
// coordination branch, the Forever budget per archetype). The weights: Strength, Agility, Stamina,
// Intellect, Spirit, hit and crit rating 1 a point (10 hit rating is 1%, 14 crit rating 1%), attack power
// 0.5 and spell power 0.86. Armor is free. The line is L(ilvl) = 1.125 * ilvl + 2.0, and each armor piece
// carries slot mod * L * f in weighted linear stats, within 1%: slot mods 1 (head, chest, legs), 0.75
// (shoulders, hands, waist, feet) and 0.5625 (neck, back, wrist, rings), f 1.093 / 1.103 / 1.048 for a
// hybrid (physical * 1.15, as Forever's own low level hybrid pieces). The head, shoulders, chest, hands,
// legs and feet are item level 66, the neck, back, wrist, waist and rings 71. Each piece also passes the
// shape check: its 1.7-norm value, (sum of (weight * amount)^1.7)^(1/1.7), within 6% of slot mod * (0.661 *
// ilvl + 2.92), which takes 4 or 5 stat lines a piece. The two-hander gets 1.035 * L(76) = 90.6, the totem
// no stats.
//
// The mix follows standards 5a and 5b, the Enhancement row of orchestration/gear_targets.md: Season of
// Discovery's melee stats with this sim's Intellect and spell power kept (the hybrid ruling), 19.4%
// Stamina, 13.0% Strength, 2.9% Agility, 10.4% Intellect, 10% Spirit, 3.9% attack power, 30.7% spell power,
// 3.6% hit and 6.0% crit. The hit goes to the cap (90 rating, the 9% melee cap against a level 63 boss),
// Stamina keeps its own share, and the other stats keep the table's proportions in what is left. A stat's
// share is its weight times its amount over the set's weighted total (839.6, the trinkets left out):
//
//	stat      amount   share   target
//	Stamina      161   19.2%   19.4 +- 1.5
//	Strength     100   11.9%   13.0 +- 3
//	Agility       22    2.6%    2.9
//	Intellect     80    9.5%   10.4
//	Spirit        78    9.3%   10
//	AP            59    3.5%    3.9
//	SP           271   27.8%   30.7 +- 3
//	Crit          46    5.5%    6.0
//	Hit           90   10.7%   at the cap
//
// Per piece (target, weighted linear total, 1.7-norm value / shape target, stats). The armor comes to 748.6
// and the two-hander to 91.0:
//
//	Head         83.3    83.5 (+0.2%)   44.9 /  46.5 (-3.5%)  12 Str, 13 Sta, 11 Int, 26 Crit, 25 SP
//	Neck         48.3    47.9 (-0.7%)   28.7 /  28.0 (+2.2%)  12 Sta, 5 Agi, 12 Hit, 22 SP
//	Shoulder     63.1    63.3 (+0.4%)   34.2 /  34.9 (-2.1%)  9 Str, 10 Sta, 8 Int, 20 Crit, 19 SP
//	Back         48.3    48.3 (+0.1%)   28.0 /  28.0 (-0.1%)  10 Sta, 8 Int, 14 Spi, 19 SP
//	Chest        83.3    83.1 (-0.3%)   48.1 /  46.5 (+3.3%)  18 Str, 19 Sta, 16 Int, 35 SP
//	Wrist        48.3    48.5 (+0.4%)   27.9 /  28.0 (-0.6%)  28 AP, 9 Sta, 10 Hit, 18 SP
//	Hands        63.1    63.1 (-0.0%)   33.1 /  34.9 (-5.3%)  11 Str, 12 Sta, 10 Int, 12 Hit, 21 SP
//	Waist        67.7    68.0 (+0.4%)   36.4 /  37.4 (-2.6%)  13 Str, 6 Agi, 14 Sta, 20 Spi, 15 Hit
//	Legs         83.3    83.2 (-0.1%)   47.9 /  46.5 (+3.0%)  17 Str, 18 Sta, 19 Hit, 34 SP
//	Feet         63.1    62.9 (-0.2%)   33.9 /  34.9 (-2.9%)  6 Agi, 12 Sta, 10 Int, 16 Spi, 22 SP
//	Finger       48.3    48.3 (+0.1%)   27.7 /  28.0 (-1.2%)  10 Str, 11 Sta, 11 Hit, 19 SP
//	Finger II    48.3    48.3 (+0.1%)   28.0 /  28.0 (-0.1%)  10 Sta, 8 Int, 14 Spi, 19 SP
//	Two-Hand     90.6    91.0 (+0.4%)  -              10 Str, 5 Agi, 11 Sta, 9 Int, 14 Spi, 31 AP, 11 Hit, 18 SP
//
// Two synthetic trinkets at item level 71 (standards 8a and 1b): a stat trinket with 32 Strength and 14
// crit rating (46.0 linear, 1.7-norm 36.4 against 36.8), and an on-use trinket of 335 attack power for 20 sec
// every 2 min (37.2 weighted points over 180 sec, sim/shaman/forever_synthetic_trinkets.go). A spell power
// on-use (140) is there too and sims 3.6 behind at 180 sec. The generator script lived in the session
// scratchpad (synth_1b.py), the numbers below are its output.
//
// The hands and legs carry The Spiritcaller's Rage, Forever's level 60 enhancement set (see
// sim/shaman/item_sets_pve.go). The standard keeps only its 2 piece bonus live (1% attack and cast speed),
// so exactly two pieces are in the set.
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
		Stats:      stats.Stats{stats.Strength: 12, stats.Stamina: 13, stats.Intellect: 11, stats.SpellPower: 25, stats.Armor: 392}.ToFloatArray(),
		CritRating: 26,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Neck, item level 71.
	{
		Id: 990002, Name: "Enhancement Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_09",
		Type:      proto.ItemType_ItemTypeNeck,
		Stats:     stats.Stats{stats.Agility: 5, stats.Stamina: 12, stats.SpellPower: 22}.ToFloatArray(),
		HitRating: 12,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Shoulder, item level 66, in place of Death's Clutch (62, 14503).
	{
		Id: 990003, Name: "Enhancement Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_25",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:      stats.Stats{stats.Strength: 9, stats.Stamina: 10, stats.Intellect: 8, stats.SpellPower: 19, stats.Armor: 317}.ToFloatArray(),
		CritRating: 20,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Back, item level 71.
	{
		Id: 990004, Name: "Enhancement Synthetic Back (Phase 1)", Icon: "inv_misc_cape_05",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 10, stats.Intellect: 8, stats.Spirit: 14, stats.SpellPower: 19, stats.Armor: 54}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Chest, item level 66, in place of Bloodmail Hauberk (61, 14611).
	{
		Id: 990005, Name: "Enhancement Synthetic Chest (Phase 1)", Icon: "inv_chest_leather_05",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Strength: 18, stats.Stamina: 19, stats.Intellect: 16, stats.SpellPower: 35, stats.Armor: 422}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Wrist, item level 71, in place of Wristguards of True Flight (71, 18812).
	{
		Id: 990006, Name: "Enhancement Synthetic Wrist (Phase 1)", Icon: "inv_bracer_02",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Stamina: 9, stats.AttackPower: 28, stats.SpellPower: 18, stats.Armor: 198}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Hands, item level 66, in place of Bloodmail Gauntlets (61, 14615).
	{
		Id: 990007, Name: "Enhancement Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_26",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Strength: 11, stats.Stamina: 12, stats.Intellect: 10, stats.SpellPower: 21, stats.Armor: 264}.ToFloatArray(),
		HitRating: 12,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
		SetName:        "Enhancement Synthetic (Phase 1)",
	},
	// Waist, item level 71, in place of Bloodmail Belt (61, 14614).
	{
		Id: 990008, Name: "Enhancement Synthetic Waist (Phase 1)", Icon: "inv_belt_23",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Strength: 13, stats.Agility: 6, stats.Stamina: 14, stats.Spirit: 20, stats.Armor: 255}.ToFloatArray(),
		HitRating: 15,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Legs, item level 66, in place of Bloodmail Legguards (61, 14612).
	{
		Id: 990009, Name: "Enhancement Synthetic Legs (Phase 1)", Icon: "inv_pants_06",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Strength: 17, stats.Stamina: 18, stats.SpellPower: 34, stats.Armor: 369}.ToFloatArray(),
		HitRating: 19,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
		SetName:        "Enhancement Synthetic (Phase 1)",
	},
	// Feet, item level 66, in place of Bloodmail Boots (61, 14616).
	{
		Id: 990010, Name: "Enhancement Synthetic Feet (Phase 1)", Icon: "inv_boots_01",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Agility: 6, stats.Stamina: 12, stats.Intellect: 10, stats.Spirit: 16, stats.SpellPower: 22, stats.Armor: 290}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Finger, item level 71.
	{
		Id: 990011, Name: "Enhancement Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_07",
		Type:      proto.ItemType_ItemTypeFinger,
		Stats:     stats.Stats{stats.Strength: 10, stats.Stamina: 11, stats.SpellPower: 19}.ToFloatArray(),
		HitRating: 11,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Second finger, item level 71. This is the caster ring of the two.
	{
		Id: 990013, Name: "Enhancement Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_15",
		Type:  proto.ItemType_ItemTypeFinger,
		Stats: stats.Stats{stats.Stamina: 10, stats.Intellect: 8, stats.Spirit: 14, stats.SpellPower: 19}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Two-Hand, item level 76.
	{
		Id: 990012, Name: "Enhancement Synthetic Two-Hand (Phase 1)", Icon: "inv_axe_09",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeAxe, HandType: proto.HandType_HandTypeTwoHand,
		Stats:     stats.Stats{stats.Strength: 10, stats.Agility: 5, stats.Stamina: 11, stats.Intellect: 9, stats.Spirit: 14, stats.AttackPower: 31, stats.SpellPower: 18}.ToFloatArray(),
		HitRating: 11,
		// Spinal Reaper damage and speed.
		WeaponDamageMin: 203, WeaponDamageMax: 305, WeaponSpeed: 3.4,
		Ilvl: 76, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Stat trinket, item level 71 (standards 1b: 46.0 linear, 1.7-norm 36.4).
	{
		Id: 990014, Name: "Enhancement Synthetic Trinket (Phase 1)", Icon: "inv_jewelry_talisman_08",
		Type:       proto.ItemType_ItemTypeTrinket,
		Stats:      stats.Stats{stats.Strength: 32}.ToFloatArray(),
		CritRating: 14,
		Ilvl:       71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// On-use trinket, item level 71: 335 attack power for 20 sec every 2 min (sim/shaman/forever_synthetic_trinkets.go).
	{
		Id: 990015, Name: "Enhancement Synthetic On-Use Trinket (Phase 1)", Icon: "inv_misc_stonetablet_11",
		Type:  proto.ItemType_ItemTypeTrinket,
		Stats: stats.Stats{}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// On-use trinket, item level 71: 140 spell power for 20 sec every 2 min, the spell power option for the hybrid.
	{
		Id: 990016, Name: "Enhancement Synthetic Spell On-Use Trinket (Phase 1)", Icon: "inv_misc_stonetablet_11",
		Type:  proto.ItemType_ItemTypeTrinket,
		Stats: stats.Stats{}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
}
