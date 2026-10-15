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
// the Scholomance Bloodmail set, Death's Clutch and Wristguards of True Flight). The two-hander keeps Spinal
// Reaper's damage per second at The Blackrock Slicer's 4.0 sec speed, the axe speed it sims best with
// (standards 22). The budget follows the cross-class standard (standards item 1b on the
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
// The mix follows standards 5a and 5b, the Enhancement row of orchestration/gear_targets.md (1a6968081):
// Season of Discovery's melee stats with this sim's Intellect and spell power kept (the hybrid ruling).
// Stamina (19.4%), Spirit (10%), Intellect (10.4%) and spell power (30.7%) hold their shares, the hit goes
// to the cap (90 rating, the 9% melee cap against a level 63 boss), and Strength, Agility, attack power and
// crit fill the rest in the table's proportions (13.0 / 2.9 / 3.9 / 6.0). A stat's share is its weight
// times its amount over the set's weighted total (839.3, the trinkets left out):
//
//	stat      amount   share   hit-adjusted target
//	Stamina      164   19.5%   19.4 +- 1.5
//	Strength      81    9.7%    9.5 +- 3
//	Agility       18    2.1%    2.1
//	Intellect     86   10.2%   10.4 (held)
//	Spirit        83    9.9%   10
//	AP            48    2.9%    2.8
//	SP           298   30.5%   30.7 (held)
//	Crit          37    4.4%    4.4
//	Hit           90   10.7%   at the cap
//
// Per piece (target, weighted linear total, 1.7-norm value / shape target, stats). The armor comes to 748.5
// and the two-hander to 90.8:
//
//	Head         83.3    83.7 (+0.4%)   45.2 /  46.5 (-2.9%)  11 Str, 15 Sta, 10 Int, 21 Crit, 31 SP
//	Neck         48.3    48.1 (-0.4%)   27.0 /  28.0 (-3.8%)  10 Sta, 3 Agi, 7 Int, 10 Hit, 21 SP
//	Shoulder     63.1    62.8 (-0.5%)   33.9 /  34.9 (-3.0%)  8 Str, 11 Sta, 8 Int, 16 Crit, 23 SP
//	Back         48.3    47.9 (-0.7%)   28.3 /  28.0 (+0.8%)  11 Sta, 7 Int, 11 Spi, 22 SP
//	Chest        83.3    83.5 (+0.2%)   44.8 /  46.5 (-3.7%)  12 Str, 16 Sta, 11 Int, 17 Spi, 32 SP
//	Wrist        48.3    48.2 (-0.1%)   27.8 /  28.0 (-0.8%)  22 AP, 10 Sta, 10 Hit, 20 SP
//	Hands        63.1    63.5 (+0.7%)   34.3 /  34.9 (-1.8%)  9 Str, 12 Sta, 8 Int, 13 Hit, 25 SP
//	Waist        67.7    68.0 (+0.4%)   36.3 /  37.4 (-3.0%)  12 Str, 6 Agi, 16 Sta, 17 Spi, 17 Hit
//	Legs         83.3    83.5 (+0.2%)   44.8 /  46.5 (-3.7%)  12 Str, 16 Sta, 11 Int, 17 Hit, 32 SP
//	Feet         63.1    63.4 (+0.5%)   35.0 /  34.9 (+0.2%)  5 Agi, 13 Sta, 9 Int, 14 Spi, 26 SP
//	Finger       48.3    48.1 (-0.4%)   28.0 /  28.0 (-0.0%)  8 Str, 11 Sta, 11 Hit, 21 SP
//	Finger II    48.3    47.9 (-0.7%)   28.3 /  28.0 (+0.8%)  11 Sta, 7 Int, 11 Spi, 22 SP
//	Two-Hand     90.6    90.8 (+0.2%)  -              9 Str, 4 Agi, 12 Sta, 8 Int, 13 Spi, 26 AP, 12 Hit, 23 SP
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
		Stats:      stats.Stats{stats.Strength: 11, stats.Stamina: 15, stats.Intellect: 10, stats.SpellPower: 31, stats.Armor: 392}.ToFloatArray(),
		CritRating: 21,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Neck, item level 71.
	{
		Id: 990002, Name: "Enhancement Synthetic Neck (Phase 1)", Icon: "inv_jewelry_necklace_09",
		Type:      proto.ItemType_ItemTypeNeck,
		Stats:     stats.Stats{stats.Agility: 3, stats.Stamina: 10, stats.Intellect: 7, stats.SpellPower: 21}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Shoulder, item level 66, in place of Death's Clutch (62, 14503).
	{
		Id: 990003, Name: "Enhancement Synthetic Shoulder (Phase 1)", Icon: "inv_shoulder_25",
		Type: proto.ItemType_ItemTypeShoulder, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:      stats.Stats{stats.Strength: 8, stats.Stamina: 11, stats.Intellect: 8, stats.SpellPower: 23, stats.Armor: 317}.ToFloatArray(),
		CritRating: 16,
		Ilvl:       66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Back, item level 71.
	{
		Id: 990004, Name: "Enhancement Synthetic Back (Phase 1)", Icon: "inv_misc_cape_05",
		Type: proto.ItemType_ItemTypeBack, ArmorType: proto.ArmorType_ArmorTypeCloth,
		Stats: stats.Stats{stats.Stamina: 11, stats.Intellect: 7, stats.Spirit: 11, stats.SpellPower: 22, stats.Armor: 54}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Chest, item level 66, in place of Bloodmail Hauberk (61, 14611).
	{
		Id: 990005, Name: "Enhancement Synthetic Chest (Phase 1)", Icon: "inv_chest_leather_05",
		Type: proto.ItemType_ItemTypeChest, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Strength: 12, stats.Stamina: 16, stats.Intellect: 11, stats.Spirit: 17, stats.SpellPower: 32, stats.Armor: 422}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Wrist, item level 71, in place of Wristguards of True Flight (71, 18812).
	{
		Id: 990006, Name: "Enhancement Synthetic Wrist (Phase 1)", Icon: "inv_bracer_02",
		Type: proto.ItemType_ItemTypeWrist, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Stamina: 10, stats.AttackPower: 22, stats.SpellPower: 20, stats.Armor: 198}.ToFloatArray(),
		HitRating: 10,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Hands, item level 66, in place of Bloodmail Gauntlets (61, 14615).
	{
		Id: 990007, Name: "Enhancement Synthetic Hands (Phase 1)", Icon: "inv_gauntlets_26",
		Type: proto.ItemType_ItemTypeHands, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Strength: 9, stats.Stamina: 12, stats.Intellect: 8, stats.SpellPower: 25, stats.Armor: 264}.ToFloatArray(),
		HitRating: 13,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
		SetName:        "Enhancement Synthetic (Phase 1)",
	},
	// Waist, item level 71, in place of Bloodmail Belt (61, 14614).
	{
		Id: 990008, Name: "Enhancement Synthetic Waist (Phase 1)", Icon: "inv_belt_23",
		Type: proto.ItemType_ItemTypeWaist, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Strength: 12, stats.Agility: 6, stats.Stamina: 16, stats.Spirit: 17, stats.Armor: 255}.ToFloatArray(),
		HitRating: 17,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Legs, item level 66, in place of Bloodmail Legguards (61, 14612).
	{
		Id: 990009, Name: "Enhancement Synthetic Legs (Phase 1)", Icon: "inv_pants_06",
		Type: proto.ItemType_ItemTypeLegs, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats:     stats.Stats{stats.Strength: 12, stats.Stamina: 16, stats.Intellect: 11, stats.SpellPower: 32, stats.Armor: 369}.ToFloatArray(),
		HitRating: 17,
		Ilvl:      66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
		SetName:        "Enhancement Synthetic (Phase 1)",
	},
	// Feet, item level 66, in place of Bloodmail Boots (61, 14616).
	{
		Id: 990010, Name: "Enhancement Synthetic Feet (Phase 1)", Icon: "inv_boots_01",
		Type: proto.ItemType_ItemTypeFeet, ArmorType: proto.ArmorType_ArmorTypeMail,
		Stats: stats.Stats{stats.Agility: 5, stats.Stamina: 13, stats.Intellect: 9, stats.Spirit: 14, stats.SpellPower: 26, stats.Armor: 290}.ToFloatArray(),
		Ilvl:  66, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Finger, item level 71.
	{
		Id: 990011, Name: "Enhancement Synthetic Finger (Phase 1)", Icon: "inv_jewelry_ring_07",
		Type:      proto.ItemType_ItemTypeFinger,
		Stats:     stats.Stats{stats.Strength: 8, stats.Stamina: 11, stats.SpellPower: 21}.ToFloatArray(),
		HitRating: 11,
		Ilvl:      71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Second finger, item level 71. This is the caster ring of the two.
	{
		Id: 990013, Name: "Enhancement Synthetic Finger II (Phase 1)", Icon: "inv_jewelry_ring_15",
		Type:  proto.ItemType_ItemTypeFinger,
		Stats: stats.Stats{stats.Stamina: 11, stats.Intellect: 7, stats.Spirit: 11, stats.SpellPower: 22}.ToFloatArray(),
		Ilvl:  71, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	// Two-Hand, item level 76.
	{
		Id: 990012, Name: "Enhancement Synthetic Two-Hand (Phase 1)", Icon: "inv_weapon_halberd_11",
		Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeAxe, HandType: proto.HandType_HandTypeTwoHand,
		Stats:     stats.Stats{stats.Strength: 9, stats.Agility: 4, stats.Stamina: 12, stats.Intellect: 8, stats.Spirit: 13, stats.AttackPower: 26, stats.SpellPower: 23}.ToFloatArray(),
		HitRating: 12,
		// The Blackrock Slicer's speed and damage spread at Spinal Reaper's 74.7 DPS
		// (standards 22: the type and speed that sim best).
		WeaponDamageMin: 238.8, WeaponDamageMax: 358.9, WeaponSpeed: 4,
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
