package database

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Forever's enchants and armor kits up to Enchanting 225, the most a level 20 enchanter
// can learn. The values are read from the recipe pages on https://hyjal.cc/professions
// (beta client data), and the list is the one https://foreverchanges.pro builds its
// level 20 BiS enchants from.
//
// Forever buffed a lot of the low Classic enchants. Cloak and bracer Minor Agility went
// from +1 to +3, Gloves Agility from +5 to +7, and 2H Minor Impact from +2 to +4 damage.
// It also added necklace enchants, +15 Strength and Agility for two-handers, and
// Forceful (Attack Power) and Mystic (Spell Power) armor kits. Classic shares one effect
// ID between enchants with equal stats, e.g. 247 is +1 Agility on a cloak, bracer and
// boots. Now those enchants differ, so every enchant here that is new or changed gets its
// own effect ID, which is its spell ID. ReplaceForeverEnchants drops the Classic entry of
// the same spell, so the enchant picker shows each enchant once.
//
// The armor kits keep the Classic kits' IDs and only gain the Stamina that Forever gave
// them. Heavy kits need an item of level 15 or above and Thick kits level 25 or above,
// which the sim does not check.
var ForeverEnchants = []*proto.UIEnchant{
	// Armor kits.
	foreverKit(15, 2304, 2831, "Light Armor Kit", stats.Stats{stats.Stamina: 1, stats.BonusArmor: 8}),
	foreverKit(16, 2313, 2832, "Medium Armor Kit", stats.Stats{stats.Stamina: 2, stats.BonusArmor: 16}),
	foreverKit(17, 4265, 2833, "Heavy Armor Kit", stats.Stats{stats.Stamina: 3, stats.BonusArmor: 24}),
	foreverKit(18, 8173, 10344, "Thick Armor Kit", stats.Stats{stats.Stamina: 4, stats.BonusArmor: 32}),
	foreverKit(1255123, 0, 1255123, "Forceful Medium Armor Kit", stats.Stats{stats.AttackPower: 4, stats.RangedAttackPower: 4, stats.BonusArmor: 16}),
	foreverKit(1255124, 0, 1255124, "Mystic Medium Armor Kit", stats.Stats{stats.SpellPower: 2, stats.BonusArmor: 16}),
	foreverKit(1255095, 0, 1255095, "Forceful Heavy Armor Kit", stats.Stats{stats.AttackPower: 6, stats.RangedAttackPower: 6, stats.BonusArmor: 24}),
	foreverKit(1255096, 0, 1255096, "Mystic Heavy Armor Kit", stats.Stats{stats.SpellPower: 4, stats.BonusArmor: 24}),
	foreverKit(1255073, 0, 1255073, "Forceful Thick Armor Kit", stats.Stats{stats.AttackPower: 8, stats.RangedAttackPower: 8, stats.BonusArmor: 32}),
	foreverKit(1255074, 0, 1255074, "Mystic Thick Armor Kit", stats.Stats{stats.SpellPower: 5, stats.BonusArmor: 32}),

	// Necklace, new in Forever (Enchanting 210, formulas for Merchant's Favor).
	foreverEnchant(1249059, "Enchant Necklace - Agility", proto.ItemType_ItemTypeNeck, stats.Stats{stats.Agility: 5}),
	foreverEnchant(1249019, "Enchant Necklace - Strength", proto.ItemType_ItemTypeNeck, stats.Stats{stats.Strength: 5}),
	foreverEnchant(1249057, "Enchant Necklace - Spell Power", proto.ItemType_ItemTypeNeck, stats.Stats{stats.SpellPower: 6}),
	foreverEnchant(1249058, "Enchant Necklace - Healing Power", proto.ItemType_ItemTypeNeck, stats.Stats{stats.HealingPower: 11, stats.SpellDamage: 4}),

	// Cloak.
	foreverEnchant(13419, "Enchant Cloak - Minor Agility", proto.ItemType_ItemTypeBack, stats.Stats{stats.Agility: 3}),

	// Bracer.
	foreverEnchant(7779, "Enchant Bracer - Minor Agility", proto.ItemType_ItemTypeWrist, stats.Stats{stats.Agility: 3}),
	foreverEnchant(7782, "Enchant Bracer - Minor Strength", proto.ItemType_ItemTypeWrist, stats.Stats{stats.Strength: 3}),
	foreverEnchant(1248458, "Enchant Bracer - Minor Intellect", proto.ItemType_ItemTypeWrist, stats.Stats{stats.Intellect: 3}),
	foreverEnchant(1248459, "Enchant Bracer - Minor Healing Power", proto.ItemType_ItemTypeWrist, stats.Stats{stats.HealingPower: 8, stats.SpellDamage: 3}),
	foreverEnchant(1248460, "Enchant Bracer - Lesser Agility", proto.ItemType_ItemTypeWrist, stats.Stats{stats.Agility: 4}),
	foreverEnchant(13536, "Enchant Bracer - Lesser Strength", proto.ItemType_ItemTypeWrist, stats.Stats{stats.Strength: 4}),
	foreverEnchant(13622, "Enchant Bracer - Lesser Intellect", proto.ItemType_ItemTypeWrist, stats.Stats{stats.Intellect: 5}),
	foreverEnchant(13661, "Enchant Bracer - Strength", proto.ItemType_ItemTypeWrist, stats.Stats{stats.Strength: 5}),
	foreverEnchant(1248497, "Enchant Bracer - Agility", proto.ItemType_ItemTypeWrist, stats.Stats{stats.Agility: 5}),
	foreverEnchant(1248498, "Enchant Bracer - Lesser Healing Power", proto.ItemType_ItemTypeWrist, stats.Stats{stats.HealingPower: 16, stats.SpellDamage: 6}),

	// Gloves.
	foreverEnchant(13815, "Enchant Gloves - Agility", proto.ItemType_ItemTypeHands, stats.Stats{stats.Agility: 7}),
	foreverEnchant(13887, "Enchant Gloves - Strength", proto.ItemType_ItemTypeHands, stats.Stats{stats.Strength: 7}),

	// Boots.
	foreverEnchant(7867, "Enchant Boots - Minor Agility", proto.ItemType_ItemTypeFeet, stats.Stats{stats.Agility: 3}),
	foreverEnchant(13637, "Enchant Boots - Lesser Agility", proto.ItemType_ItemTypeFeet, stats.Stats{stats.Agility: 4}),

	// Weapon. Winter's Might went from +7 to +15 Frost spell damage.
	foreverEnchant(21931, "Enchant Weapon - Winter's Might", proto.ItemType_ItemTypeWeapon, stats.Stats{stats.FrostPower: 15}),

	// Two-hand weapon. The Impact enchants are weapon effects in sim/common.
	foreverTwoHand(7793, "Enchant 2H Weapon - Lesser Intellect", stats.Stats{stats.Intellect: 5}),
	foreverTwoHand(7745, "Enchant 2H Weapon - Minor Impact", stats.Stats{}),
	foreverTwoHand(13529, "Enchant 2H Weapon - Lesser Impact", stats.Stats{}),
	foreverTwoHand(13695, "Enchant 2H Weapon - Impact", stats.Stats{}),
	foreverTwoHand(1248510, "Enchant 2H Weapon - Lesser Agility", stats.Stats{stats.Agility: 15}),
	foreverTwoHand(1248511, "Enchant 2H Weapon - Lesser Strength", stats.Stats{stats.Strength: 15}),
}

func foreverEnchant(spellID int32, name string, itemType proto.ItemType, s stats.Stats) *proto.UIEnchant {
	return &proto.UIEnchant{EffectId: spellID, SpellId: spellID, Name: name, Quality: proto.ItemQuality_ItemQualityCommon, Stats: s.ToFloatArray(), Type: itemType}
}

func foreverTwoHand(spellID int32, name string, s stats.Stats) *proto.UIEnchant {
	enchant := foreverEnchant(spellID, name, proto.ItemType_ItemTypeWeapon, s)
	enchant.EnchantType = proto.EnchantType_EnchantTypeTwoHand
	return enchant
}

func foreverKit(effectID int32, itemID int32, spellID int32, name string, s stats.Stats) *proto.UIEnchant {
	return &proto.UIEnchant{
		EffectId: effectID, ItemId: itemID, SpellId: spellID, Name: name, Quality: proto.ItemQuality_ItemQualityCommon, Stats: s.ToFloatArray(),
		Type:        proto.ItemType_ItemTypeChest,
		ExtraTypes:  []proto.ItemType{proto.ItemType_ItemTypeLegs, proto.ItemType_ItemTypeHands, proto.ItemType_ItemTypeFeet},
		EnchantType: proto.EnchantType_EnchantTypeKit,
	}
}

// ReplaceForeverEnchants puts the Forever enchants in, in place of the Classic entries
// for the same spell.
func (db *WowDatabase) ReplaceForeverEnchants(arr []*proto.UIEnchant) {
	for _, enchant := range arr {
		for key, old := range db.Enchants {
			if old.SpellId == enchant.SpellId && key != EnchantToDBKey(enchant) {
				delete(db.Enchants, key)
			}
		}
		db.MergeEnchant(enchant)
	}
}
