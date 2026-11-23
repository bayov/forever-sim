//go:build with_db

package core

import (
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Forever's gear crit counts for every kind of attack, but the game limits two trinkets'
// crit to one kind: Blackhand's Breadth reads "with Melee and Ranged attacks" and Eye of
// the Beast "with Spells". Painweaver Band's crit rating ("by 1.0%") still counts for both.
func TestForeverKindLimitedCrit(t *testing.T) {
	const blackhandsBreadth, eyeOfTheBeast, painweaverBand = 13965, 13968, 13098
	for _, id := range []int32{blackhandsBreadth, eyeOfTheBeast} {
		if !ItemsByID[id].KindLimitedCrit {
			t.Fatalf("%s has no KindLimitedCrit in the database", ItemsByID[id].Name)
		}
	}
	if ItemsByID[painweaverBand].KindLimitedCrit {
		t.Fatalf("%s has KindLimitedCrit in the database", ItemsByID[painweaverBand].Name)
	}

	equipCrit := func(ids ...int32) (melee, spell float64) {
		character := foreverTestCharacter()
		for i := range character.itemStatMultipliers {
			character.itemStatMultipliers[i] = 1
		}
		slots := []proto.ItemSlot{proto.ItemSlot_ItemSlotTrinket1, proto.ItemSlot_ItemSlotTrinket2, proto.ItemSlot_ItemSlotFinger1}
		for i, id := range ids {
			character.Equipment[slots[i]] = ItemsByID[id]
		}
		s := character.unifyEquipHitAndCrit(character.addEquipRatings(character.EquipStats()))
		return s[stats.MeleeCrit], s[stats.SpellCrit]
	}
	cases := []struct {
		name               string
		ids                []int32
		wantMelee, wantSpl float64
	}{
		// A caster gets no spell crit from Blackhand's Breadth, a melee build no melee crit
		// from Eye of the Beast.
		{"Blackhand's Breadth", []int32{blackhandsBreadth}, 1, 0},
		{"Eye of the Beast", []int32{eyeOfTheBeast}, 0, 1},
		{"Painweaver Band", []int32{painweaverBand}, 1, 1},
		{"all three", []int32{blackhandsBreadth, eyeOfTheBeast, painweaverBand}, 2, 2},
	}
	for _, c := range cases {
		melee, spell := equipCrit(c.ids...)
		if melee != c.wantMelee || spell != c.wantSpl {
			t.Errorf("%s: melee crit %.2f and spell crit %.2f, want %.2f and %.2f", c.name, melee, spell, c.wantMelee, c.wantSpl)
		}
	}
}
