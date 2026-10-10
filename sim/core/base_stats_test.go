package core

import (
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Naked characters with no talents on the Forever beta (2026-10-10, /bayov export). Forever's
// Undead have 3 more Strength, 1 more Agility and 5 less Spirit than 1.12's, so these rows
// check that shift and the Undead paladin row the sim builds from the Undead warrior.
func TestForeverUndeadBaseAttributes(t *testing.T) {
	cases := []struct {
		class proto.Class
		level int32
		attrs [5]float64 // Strength, Agility, Stamina, Intellect, Spirit
	}{
		{proto.Class_ClassPaladin, 1, [5]float64{24, 19, 23, 18, 21}},
		{proto.Class_ClassPaladin, 20, [5]float64{44, 30, 42, 30, 34}},
		{proto.Class_ClassRogue, 1, [5]float64{23, 22, 22, 18, 20}},
		{proto.Class_ClassPriest, 1, [5]float64{22, 19, 21, 20, 23}},
		{proto.Class_ClassMage, 1, [5]float64{22, 19, 21, 21, 22}},
	}
	for _, c := range cases {
		base := getBaseStatsCombo(proto.Race_RaceUndead, c.class, c.level)
		got := [5]float64{base[stats.Strength], base[stats.Agility], base[stats.Stamina], base[stats.Intellect], base[stats.Spirit]}
		if got != c.attrs {
			t.Errorf("level %d Undead %s: got %v, the beta has %v", c.level, c.class, got, c.attrs)
		}
	}
}
