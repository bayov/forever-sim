package core

import (
	"math"
	"testing"
)

// The GetManaRegen readings from the Forever beta (2026-10-07), with the 0.001 floor the
// client adds taken off. Each one is mana a second while not casting.
func TestForeverSpiritManaRegen(t *testing.T) {
	readings := []struct {
		who    string
		spirit float64
		regen  float64
	}{
		{"level 1 Orc shaman", 22, 5.5},
		{"level 20 Undead paladin", 34, 8.5},
		{"level 20 Undead paladin", 42, 10.5},
		{"level 30 Orc shaman", 53, 12.875},
		{"level 30 Orc shaman", 60, 13.75},
		{"level 30 Orc shaman", 65, 14.375},
		{"level 30 Orc shaman", 69, 14.875},
	}
	for _, r := range readings {
		if got := ForeverSpiritManaRegenPerSecond(r.spirit); math.Abs(got-r.regen) > 1e-9 {
			t.Errorf("%s with %.0f Spirit: got %.3f, want %.3f", r.who, r.spirit, got, r.regen)
		}
	}
}
