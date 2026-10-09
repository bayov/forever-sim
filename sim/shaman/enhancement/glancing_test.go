package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
)

// TestOrcShamanGlancing checks how often our white hits glance and how much a glancing blow
// deals, by the mob's level and our weapon skill.
//
// The chance is 10% plus 2% for each point the mob's defense (its level times 5) is above
// our level times 5, from 0 to 40%. Skill from gear doesn't lower it. A glancing blow deals
// a random part of the hit between a low end, 1.3 - 0.05 * (defense - skill) and at most
// 0.91, and a high end, 1.2 - 0.03 * (defense - skill) and at most 0.99. Here skill from
// gear counts. These are Beaza's formulas, which magey's 2019 Classic tests confirmed.
func TestOrcShamanGlancing(t *testing.T) {
	cases := []struct {
		name               string
		level, targetLevel int32
		weapon             int32
		chance, low, high  float64
	}{
		{"level 60 against a level 63 boss", 60, 63, darkEdge, 0.4, 0.55, 0.75},
		{"310 skill against the boss", 60, 63, thorium, 0.4, 0.91, 0.99},
		{"level 30 against a level 34", 30, 34, whirlwind, 0.4, 0.3, 0.6},
		{"152 skill against a level 33", 30, 33, treeChop, 0.4, 0.65, 0.81},
		{"level 30 against Vishas (level 32)", 30, 32, whirlwind, 0.3, 0.8, 0.9},
		{"a level 31", 30, 31, whirlwind, 0.2, 0.91, 0.99},
		{"the same level", 30, 30, whirlwind, 0.1, 0.91, 0.99},
		{"a level 29", 30, 29, whirlwind, 0, 0, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sim, enh := newTargetLevelSim(c.level, c.targetLevel, c.weapon, stormstrike)
			white := enh.AutoAttacks.MHAuto()

			// A normal hit always deals the same here, so each glancing blow's share of it is
			// its multiplier.
			const rolls = 20000
			var hit float64
			var glances []float64
			for range rolls {
				result := white.CalcDamage(sim, enh.CurrentTarget, 100, white.OutcomeMeleeWhite)
				switch {
				case result.Outcome.Matches(core.OutcomeGlance):
					glances = append(glances, result.Damage)
				case result.Outcome.Matches(core.OutcomeHit):
					hit = result.Damage
				}
			}
			if hit == 0 {
				t.Fatalf("no normal hits")
			}

			rate := float64(len(glances)) / rolls
			if c.chance == 0 {
				if len(glances) != 0 {
					t.Errorf("%d of %d white hits glanced, want none", len(glances), rolls)
				}
				return
			}
			if se := math.Sqrt(c.chance * (1 - c.chance) / rolls); math.Abs(rate-c.chance) > 4*se {
				t.Errorf("%.4f of white hits glanced, want %.4f +- %.4f", rate, c.chance, 4*se)
			}
			lowest, highest := 1.0, 0.0
			for _, damage := range glances {
				lowest, highest = min(lowest, damage/hit), max(highest, damage/hit)
			}
			// Thousands of rolls reach both ends of the range.
			if math.Abs(lowest-c.low) > 0.005 || math.Abs(highest-c.high) > 0.005 {
				t.Errorf("glancing blows dealt %.3f to %.3f of a hit, want %.3f to %.3f", lowest, highest, c.low, c.high)
			}
		})
	}
}
