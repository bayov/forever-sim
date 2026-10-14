package enhancement

import (
	"fmt"
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestOrcShamanImprovedStormstrike checks how often Improved Stormstrike gives its buff and
// resets Stormstrike (shaman_audit.md 8.32).
//
// Each point gives a 50% chance that a Stormstrike cast gives the regen buff, and a 50% chance
// that our dodge or parry of an enemy attack resets Stormstrike's cooldown. A mob dodging or
// parrying our own swing doesn't reset it. On the beta all 194 casts at 2/2 gave the buff, and
// 24 casts came early after a dodge or parry of ours. We cast Stormstrike 1000 times at 1/2 and
// 2/2, and dodge or parry a level 30 mob's swing 1000 times with Stormstrike on cooldown.
func TestOrcShamanImprovedStormstrike(t *testing.T) {
	const tries = 1000
	for _, points := range []int{1, 2} {
		want := .5 * float64(points)
		// Stormstrike and Improved Stormstrike points.
		sim, enh := newTankSim(newOrcShaman(30, fmt.Sprintf("-000000000000100%d", points), &proto.EnhancementShaman_Options{}), 30)
		if enh.Talents.ImprovedStormstrike != int32(points) {
			t.Fatalf("got Improved Stormstrike %d, want %d", enh.Talents.ImprovedStormstrike, points)
		}
		buff := enh.GetAura("Improved Stormstrike")
		target := enh.CurrentTarget
		check := func(what string, got int, want float64) {
			if rate := float64(got) / tries; math.Abs(rate-want) > .06 {
				t.Errorf("Improved Stormstrike %d: %s %.1f%% of the time, want %.0f%%", points, what, rate*100, want*100)
			}
		}

		at(sim, 1, func(sim *core.Simulation) {
			var buffs, resets, ownResets int
			for range tries {
				buff.Deactivate(sim)
				enh.Stormstrike.CD.Reset()
				enh.AddMana(sim, enh.MaxMana(), enh.GetManaNotCastingMetrics())
				castNow(t, sim, enh, enh.Stormstrike)
				if buff.IsActive() {
					buffs++
				}
			}
			swing := target.AutoAttacks.MHAuto()
			for i := range tries {
				enh.Stormstrike.CD.Use(sim)
				outcome := []core.HitOutcome{core.OutcomeDodge, core.OutcomeParry}[i%2]
				swing.CalcAndDealDamage(sim, &enh.Unit, 100, forcedOutcome(outcome))
				if enh.Stormstrike.CD.IsReady(sim) {
					resets++
				}

				enh.Stormstrike.CD.Use(sim)
				ours := enh.AutoAttacks.MHAuto()
				ours.CalcAndDealDamage(sim, target, 100, forcedOutcome(outcome))
				if enh.Stormstrike.CD.IsReady(sim) {
					ownResets++
				}
			}
			check("gave the buff on a cast", buffs, want)
			check("reset Stormstrike on our dodge or parry", resets, want)
			check("reset Stormstrike when the mob dodged or parried us", ownResets, 0)
		})
		runSim(sim)
	}
}
