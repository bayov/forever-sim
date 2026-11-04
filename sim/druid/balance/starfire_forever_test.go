package balance

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestStarfireNov2Damage checks Starfire rank 6 at level 60 against the November 2 development
// notes' 40% more base damage: 313.43 to 369.77 in the client (9876 at its cap), so 438.80 to
// 517.68, with no spell power, crits or resists in the way.
func TestStarfireNov2Damage(t *testing.T) {
	sim, moonkin := newMoonkinSim(t, "")
	target := moonkin.CurrentTarget
	moonkin.AddStatDynamic(sim, stats.SpellCrit, -1000*core.SpellCritRatingPerCritChance)
	starfire := topRank(moonkin.Starfire)
	// The damage multipliers on a base of 1000, to take them out of the hits.
	multiplier := starfire.CalcDamage(sim, target, 1000, starfire.OutcomeAlwaysHit).Damage / 1000

	low, high := 1.4*313.43, 1.4*369.77
	seen, lowest, highest := 0, 1e9, 0.0
	for i := 0; i < 400; i++ {
		metrics := &starfire.SpellMetrics[target.UnitIndex]
		before, hits := metrics.TotalDamage, metrics.Hits
		starfire.SpellMetrics[target.UnitIndex].Casts++
		starfire.ApplyEffects(sim, target, starfire.Spell)
		if metrics.Hits == hits || metrics.TotalDamage == before {
			continue
		}
		base := (metrics.TotalDamage - before) / multiplier
		if base > highest {
			highest = base
		}
		if base < lowest {
			lowest = base
		}
		seen++
	}
	if seen < 100 {
		t.Fatalf("only %d Starfire hits", seen)
	}
	// Damage is rounded to a whole point, and a partial resist can take a hit under the range.
	if highest > high+1 || highest < high-15 {
		t.Errorf("Starfire's highest base damage is %.2f, want about %.2f", highest, high)
	}
	if lowest > low+15 {
		t.Errorf("Starfire's lowest base damage is %.2f, want about %.2f or less", lowest, low)
	}
}
