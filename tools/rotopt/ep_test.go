package main

import (
	"slices"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
)

// keepByEP keeps each slot's best options by EP and everything a linear weight can't judge.
func TestKeepByEP(t *testing.T) {
	statsOf := func(stat proto.Stat, v float64) []float64 {
		s := make([]float64, len(proto.Stat_name))
		s[stat] = v
		return s
	}
	add := func(id int32, item *proto.UIItem) gearOption {
		item.Id = id
		itemsByID[id] = item
		return gearOption{spec: &proto.ItemSpec{Id: id}, twoHand: item.HandType == proto.HandType_HandTypeTwoHand}
	}
	pool := make([][]gearOption, len(gearSlotNames))
	head := int(proto.ItemSlot_ItemSlotHead)
	// Ten heads of 10 to 100 stamina, the 10 one with a proc, the 20 one a set piece, the 30 one
	// the starting item, the 40 one included, and a random suffix roll whose suffix wins.
	for i := int32(1); i <= 10; i++ {
		item := &proto.UIItem{Stats: statsOf(proto.Stat_StatStamina, float64(10*i))}
		if i == 2 {
			item.SetName = "Set"
		}
		pool[head] = append(pool[head], add(9000+i, item))
	}
	roll := add(9100, &proto.UIItem{})
	roll.spec.RandomSuffix = 7
	roll.suffixStats = statsOf(proto.Stat_StatStamina, 500)
	pool[head] = append(pool[head], roll)
	// Forever's hit rating counts as hit: 20 rating is 2%, worth 2 x 30.
	pool[head] = append(pool[head], add(9101, &proto.UIItem{HitRating: 20}))

	mh, oh := int(proto.ItemSlot_ItemSlotMainHand), int(proto.ItemSlot_ItemSlotOffHand)
	// One handers beat the two handers on EP. One two hander has the most weapon DPS.
	for i := int32(1); i <= 4; i++ {
		pool[mh] = append(pool[mh], add(9200+i, &proto.UIItem{Stats: statsOf(proto.Stat_StatStamina, float64(100+i)),
			HandType: proto.HandType_HandTypeOneHand, WeaponSpeed: 2, WeaponDamageMin: 10, WeaponDamageMax: 10}))
		pool[mh] = append(pool[mh], add(9300+i, &proto.UIItem{Stats: statsOf(proto.Stat_StatStamina, float64(i)),
			HandType: proto.HandType_HandTypeTwoHand, WeaponSpeed: 3, WeaponDamageMin: 30, WeaponDamageMax: 30}))
	}
	pool[mh] = append(pool[mh], add(9310, &proto.UIItem{HandType: proto.HandType_HandTypeTwoHand, WeaponSpeed: 3, WeaponDamageMin: 300, WeaponDamageMax: 300}))
	// Shields and held items have no weapon DPS, each class keeps its own best.
	for i := int32(1); i <= 3; i++ {
		pool[oh] = append(pool[oh], add(9400+i, &proto.UIItem{Stats: statsOf(proto.Stat_StatStamina, float64(100+i)), WeaponType: proto.WeaponType_WeaponTypeShield}))
		pool[oh] = append(pool[oh], add(9500+i, &proto.UIItem{Stats: statsOf(proto.Stat_StatStamina, float64(i)), WeaponType: proto.WeaponType_WeaponTypeOffHand}))
	}

	weights := make([]float64, len(proto.Stat_name))
	weights[proto.Stat_StatStamina] = 1
	weights[proto.Stat_StatMeleeHit] = 30
	start := make([]*proto.ItemSpec, len(gearSlotNames))
	start[head] = &proto.ItemSpec{Id: 9003}
	effect := func(id int32) bool { return id == 9001 }
	got := keepByEP(pool, weights, 2, start, []int32{9004}, effect)

	ids := func(options []gearOption) []int32 {
		var out []int32
		for _, o := range options {
			out = append(out, o.spec.Id)
		}
		slices.Sort(out)
		return out
	}
	check := func(slot int, want []int32) {
		t.Helper()
		if g := ids(got[slot]); !slices.Equal(g, want) {
			t.Errorf("%s kept %v, want %v", gearSlotNames[slot], g, want)
		}
	}
	// Top 2 by EP (the roll and 100 stamina), the proc, the set piece, the start, the include.
	check(head, []int32{9001, 9002, 9003, 9004, 9010, 9100})
	// One handers: top 2 by EP (also overall) and by DPS (all tie, the first two). Two handers: top 2
	// by EP, and by DPS the 9310 and the first of the ties.
	check(mh, []int32{9201, 9202, 9203, 9204, 9301, 9303, 9304, 9310})
	check(oh, []int32{9402, 9403, 9502, 9503})
	if v := epValue(pool[head][11], weights); v != 60 {
		t.Errorf("20 hit rating is worth %g, want 60", v)
	}
}
