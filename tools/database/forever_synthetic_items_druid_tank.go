package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The Feral tank's synthetic Phase 1 set (ids 990801 to 990815), made like the Feral cat's
// (forever_synthetic_items_druid.go, read its comment first) by standards 36: the bear row of the
// Tanks table in orchestration/gear_targets.md, priced by standards 1b with standards 36's tank prices
// (Defense 1.5 a point, dodge 12 a 1%). The bear is the physical archetype, so every armor piece has
// the cat's budget for its slot (physical f, 650.0 weighted for the armor) and 3 stat lines, and the
// two-hander is the feral 0.90 x L(76) = 78.8 (79.0 here); the set comes to 729.5 of standards 36's
// 729.3. Per piece, the weighted linear total against its budget and the 1.7-norm shape against its
// target (the two-hander has no shape check):
//
//	Head 72.5 / 72 -0.7%, Neck 42.0 / 42 -4.3% (with Agility for hit), Shoulder 55.0 / 55 +0.9%, Back 42.0 / 42 -4.3%,
//	Chest 71.5 / 72 -1.6%, Wrist 42.0 / 42 -4.3%, Hands 55.0 / 55 +0.7%, Waist 59.0 / 59 +0.6%,
//	Legs 72.5 / 72 -0.7%, Feet 55.0 / 55 +0.7%, Finger 42.0 / 42 -4.3%, Finger II 42.0 / 42 -4.6%,
//	Two-Hand 79.0 / 78.8
//
// The bear row's small shares (Intellect 2.3, crit 1.7, attack power 0.5, spell power 0.3, 4.8 points
// in all, each under 5a's 3) go to the other stats in proportion. Hit is set by item 4 against the tank
// objective: the bear's best flask is the Flask of Natural Accuracy (5% hit; 105 ahead of the Flask of
// the Titans, 35 ahead of Natural Aggression with the gear's hit at 8%, Balanced, the scoring
// encounters), so the set carries the 3% that takes it to the 8% special attack cap, and the neck's 16
// hit rating went to Agility (+43 over hit past the cap, +27 into Defense, +18 into Stamina). Linear
// shares, weapon included:
//
//	L %     Sta 35.1  Agi 22.5  Def 21.6  Str 16.7  Hit 4.1
//	target  Sta 35.1  Agi 20.1  Def 21.8  Str 16.8  Hit 6.2   (the row's Sta 33.4, Agi 19.1, Def 20.8,
//	                                                           Str 16.0, Hit 5.9 with the small shares)
//
// The set totals 256 Stamina, 164 Agility, 122 Strength, 105 Defense and 30 hit rating (3%). The
// armor is the cat's (Molten Core's leather tier), the two-hander has the cat's weapon damage
// (Earthshaker's, which a bear swings at its DPS), and the relic slot takes a real Forever idol,
// worth its slot through its effect (standards 1a). The head and chest form a set with Grovekeeper
// Ferocity's bonuses, as the cat's do.
//
// The trinkets (standards 36): the tank stat trinket, 23 Defense and 1% dodge (46.5 weighted, 1.3% over
// 45.9: no whole Defense and dodge reach the 1% band, and a bear can't parry or block; shape 37.7,
// +2.5%), the row being all Defense, and a defensive on-use trinket of the same worth as 1b's
// physical one, 1,665 bonus armor for 20 sec every 2 min (sim/druid/forever_synthetic_trinkets.go),
// used on a timer that reads no health. The generator is in the session scratchpad (bearset.py).

func init() {
	ForeverSyntheticItems = append(ForeverSyntheticItems, foreverSyntheticFeralTankItems...)
}

func feralTankPiece(id int32, name, icon string, itemType proto.ItemType, ilvl int32, s stats.Stats, hit float64) *proto.UIItem {
	item := &proto.UIItem{
		Id: id, Name: "Feral Tank Synthetic " + name + " (Phase 1)", Icon: icon,
		Type:      itemType,
		Stats:     s.ToFloatArray(),
		HitRating: hit,
		Ilvl:      ilvl, RequiredLevel: 60, Phase: 1, Quality: proto.ItemQuality_ItemQualityEpic, Expansion: proto.Expansion_ExpansionVanilla,
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	}
	switch itemType {
	case proto.ItemType_ItemTypeNeck, proto.ItemType_ItemTypeFinger, proto.ItemType_ItemTypeTrinket:
	case proto.ItemType_ItemTypeBack:
		item.ArmorType = proto.ArmorType_ArmorTypeCloth
	default:
		item.ArmorType = proto.ArmorType_ArmorTypeLeather
	}
	return item
}

var foreverSyntheticFeralTankItems = func() []*proto.UIItem {
	head := feralTankPiece(990801, "Head", "inv_helmet_04", proto.ItemType_ItemTypeHead, 66,
		stats.Stats{stats.Stamina: 27, stats.Defense: 15, stats.Agility: 23, stats.Armor: 163}, 0)
	head.SetName = "Feral Tank Synthetic (Phase 1)"
	chest := feralTankPiece(990805, "Chest", "inv_chest_leather_08", proto.ItemType_ItemTypeChest, 66,
		stats.Stats{stats.Stamina: 24, stats.Defense: 13, stats.Strength: 28, stats.Armor: 200}, 0)
	chest.SetName = "Feral Tank Synthetic (Phase 1)"
	weapon := feralTankPiece(990813, "Two-Hand", "inv_mace_25", proto.ItemType_ItemTypeWeapon, 76,
		stats.Stats{stats.Stamina: 30, stats.Agility: 25, stats.Defense: 16}, 0)
	weapon.WeaponType = proto.WeaponType_WeaponTypeMace
	weapon.HandType = proto.HandType_HandTypeTwoHand
	weapon.ArmorType = proto.ArmorType_ArmorTypeUnknown
	// Earthshaker's damage and speed, as the cat's.
	weapon.WeaponDamageMin, weapon.WeaponDamageMax, weapon.WeaponSpeed = 175, 263, 3.5
	return []*proto.UIItem{
		head,
		feralTankPiece(990802, "Neck", "inv_jewelry_necklace_09", proto.ItemType_ItemTypeNeck, 71,
			stats.Stats{stats.Stamina: 14, stats.Defense: 8, stats.Agility: 16}, 0),
		feralTankPiece(990803, "Shoulder", "inv_shoulder_25", proto.ItemType_ItemTypeShoulder, 66,
			stats.Stats{stats.Stamina: 19, stats.Defense: 10, stats.Strength: 21, stats.Armor: 150}, 0),
		feralTankPiece(990804, "Back", "inv_misc_cape_08", proto.ItemType_ItemTypeBack, 71,
			stats.Stats{stats.Stamina: 16, stats.Defense: 8, stats.Agility: 14, stats.Armor: 55}, 0),
		chest,
		feralTankPiece(990806, "Wrist", "inv_bracer_07", proto.ItemType_ItemTypeWrist, 71,
			stats.Stats{stats.Stamina: 14, stats.Defense: 8, stats.Strength: 16, stats.Armor: 88}, 0),
		feralTankPiece(990807, "Hands", "inv_gauntlets_23", proto.ItemType_ItemTypeHands, 66,
			stats.Stats{stats.Stamina: 18, stats.Agility: 16, stats.Strength: 21, stats.Armor: 125}, 0),
		feralTankPiece(990808, "Waist", "inv_belt_16", proto.ItemType_ItemTypeWaist, 71,
			stats.Stats{stats.Stamina: 22, stats.Defense: 12, stats.Agility: 19, stats.Armor: 113}, 0),
		feralTankPiece(990809, "Legs", "inv_pants_wolf", proto.ItemType_ItemTypeLegs, 66,
			stats.Stats{stats.Stamina: 27, stats.Defense: 15, stats.Agility: 23, stats.Armor: 175}, 0),
		feralTankPiece(990810, "Feet", "inv_boots_08", proto.ItemType_ItemTypeFeet, 66,
			stats.Stats{stats.Stamina: 18, stats.Agility: 16, stats.Strength: 21, stats.Armor: 138}, 0),
		feralTankPiece(990811, "Finger", "inv_jewelry_ring_15", proto.ItemType_ItemTypeFinger, 71,
			stats.Stats{stats.Stamina: 14, stats.Agility: 12}, 16),
		feralTankPiece(990812, "Finger II", "inv_jewelry_ring_09", proto.ItemType_ItemTypeFinger, 71,
			stats.Stats{stats.Stamina: 13, stats.Strength: 15}, 14),
		weapon,
		feralTankPiece(990814, "Tank Trinket", "inv_misc_armorkit_09", proto.ItemType_ItemTypeTrinket, 71,
			stats.Stats{stats.Defense: 23, stats.Dodge: 1}, 0),
		// The on-use half has no passive stats; its effect is in sim/druid/forever_synthetic_trinkets.go.
		feralTankPiece(990815, "Armor Trinket", "inv_misc_rune_06", proto.ItemType_ItemTypeTrinket, 71,
			stats.Stats{}, 0),
	}
}()
