package main

import (
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
)

func TestEnchantFitsItem(t *testing.T) {
	player := &proto.Player{Class: proto.Class_ClassShaman}
	twoHander := &proto.UIItem{Type: proto.ItemType_ItemTypeWeapon, HandType: proto.HandType_HandTypeTwoHand, WeaponType: proto.WeaponType_WeaponTypeAxe, Ilvl: 63}
	oneHander := &proto.UIItem{Type: proto.ItemType_ItemTypeWeapon, HandType: proto.HandType_HandTypeMainHand, WeaponType: proto.WeaponType_WeaponTypeAxe, Ilvl: 63}
	legs := &proto.UIItem{Type: proto.ItemType_ItemTypeLegs, Ilvl: 30}
	totem := &proto.UIItem{Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeTotem}
	twoHand := &proto.UIEnchant{EffectId: 2646, Type: proto.ItemType_ItemTypeWeapon, EnchantType: proto.EnchantType_EnchantTypeTwoHand}
	crusader := &proto.UIEnchant{EffectId: 1900, Type: proto.ItemType_ItemTypeWeapon}
	rugged := &proto.UIEnchant{EffectId: 1254965, Type: proto.ItemType_ItemTypeChest, ExtraTypes: []proto.ItemType{proto.ItemType_ItemTypeLegs}, EnchantType: proto.EnchantType_EnchantTypeKit}
	scope := &proto.UIEnchant{EffectId: 664, Type: proto.ItemType_ItemTypeRanged}
	zg := &proto.UIEnchant{EffectId: 2606, Type: proto.ItemType_ItemTypeShoulder, Phase: 4}
	shoulder := &proto.UIItem{Type: proto.ItemType_ItemTypeShoulder}
	enchantMaxPhase = 2
	defer func() { enchantMaxPhase = 0 }()
	for _, c := range []struct {
		name string
		e    *proto.UIEnchant
		item *proto.UIItem
		want bool
	}{
		{"2H enchant on a two hander", twoHand, twoHander, true},
		{"2H enchant on a one hander", twoHand, oneHander, false},
		{"Crusader on a two hander", crusader, twoHander, true},
		{"Rugged kit on item level 30 legs", rugged, legs, false},
		{"scope on a totem", scope, totem, false},
		{"Zul'Gurub shoulder at phase 2", zg, shoulder, false},
	} {
		if got := enchantFitsItem(c.e, c.item, player, 60); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestEnchantEffect(t *testing.T) {
	stats := func(kv map[proto.Stat]float64) []float64 {
		out := make([]float64, len(proto.Stat_name))
		for k, v := range kv {
			out[k] = v
		}
		return out
	}
	for _, c := range []struct {
		stats map[proto.Stat]float64
		want  string
	}{
		{map[proto.Stat]float64{proto.Stat_StatStrength: 8}, "+8 Strength"},
		{map[proto.Stat]float64{proto.Stat_StatStrength: 4, proto.Stat_StatAgility: 4, proto.Stat_StatStamina: 4, proto.Stat_StatIntellect: 4, proto.Stat_StatSpirit: 4}, "+4 all stats"},
		{map[proto.Stat]float64{proto.Stat_StatAttackPower: 10, proto.Stat_StatRangedAttackPower: 10, proto.Stat_StatBonusArmor: 40}, "+10 AP, +40 armor"},
		{map[proto.Stat]float64{proto.Stat_StatMeleeHaste: 3}, "+3% attack speed"},
		{nil, ""},
	} {
		if got := enchantEffect(&proto.UIEnchant{Stats: stats(c.stats)}); got != c.want {
			t.Errorf("got %q, want %q", got, c.want)
		}
	}
}
