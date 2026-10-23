package core

import (
	"math"
	"testing"
)

// TestForeverDamageTakenRageArmor checks the armor Forever's Rage from damage taken assumes, by
// the attacker's level (wowsims/forever 73b317c7cf and de550b7df9): 20% up to 42, 30% at 44, 36%
// at 46, 38% at 50 and 40% from 63, interpolated in between.
func TestForeverDamageTakenRageArmor(t *testing.T) {
	for _, c := range []struct {
		level int32
		want  float64
	}{
		{1, 0.20}, {16, 0.20}, {42, 0.20}, {43, 0.25}, {44, 0.30}, {45, 0.33}, {46, 0.36},
		{48, 0.37}, {50, 0.38}, {60, 0.38 + 0.02*10/13}, {63, 0.40}, {70, 0.40},
	} {
		if got := foreverDamageTakenRageArmor(c.level); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("a level %d attacker: %.4f assumed armor, want %.4f", c.level, got, c.want)
		}
	}

	// About 0.0062 Rage a point of damage from a level 63 boss, 0.0183 from a level 30 mob.
	for _, c := range []struct {
		level int32
		want  float64
	}{{63, 0.006157}, {30, 0.018309}} {
		if got := ForeverDamageTakenRageFactor * (1 - foreverDamageTakenRageArmor(c.level)) / foreverRageConversion(c.level); math.Abs(got-c.want) > 1e-6 {
			t.Errorf("a level %d attacker: %.6f Rage a point of damage, want %.6f", c.level, got, c.want)
		}
	}
}
