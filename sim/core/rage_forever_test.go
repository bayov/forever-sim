package core

import (
	"math"
	"testing"
)

// TestForeverDamageTakenRageArmor checks the armor Forever's Rage from damage taken assumes, by
// the attacker's level (EV-0048, EV-0034, EV-0035, EV-0050, de550b7df9): 21.8% up to 38, 23.9% at
// 42, 31% at 44, 36% at 46, 37.7% at 47, 38% at 50, 38.1% at 55, 38% at 59 and 40% from 63,
// interpolated in between.
func TestForeverDamageTakenRageArmor(t *testing.T) {
	for _, c := range []struct {
		level int32
		want  float64
	}{
		{1, 0.218}, {16, 0.218}, {38, 0.218}, {40, 0.2285}, {42, 0.239}, {43, 0.2745}, {44, 0.31},
		{45, 0.335}, {46, 0.36}, {47, 0.377}, {48, 0.378}, {50, 0.38}, {55, 0.381}, {59, 0.38},
		{60, 0.385}, {63, 0.40}, {70, 0.40},
	} {
		if got := foreverDamageTakenRageArmor(c.level); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("a level %d attacker: %.4f assumed armor, want %.4f", c.level, got, c.want)
		}
	}

	// About 0.0062 Rage a point of damage from a level 63 boss, 0.0179 from a level 30 mob.
	for _, c := range []struct {
		level int32
		want  float64
	}{{63, 0.006157}, {30, 0.017897}} {
		if got := ForeverDamageTakenRageFactor * (1 - foreverDamageTakenRageArmor(c.level)) / foreverRageConversion(c.level); math.Abs(got-c.want) > 1e-6 {
			t.Errorf("a level %d attacker: %.6f Rage a point of damage, want %.6f", c.level, got, c.want)
		}
	}
}
