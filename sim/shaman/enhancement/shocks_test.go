package enhancement

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestOrcShamanSharedShockCooldown checks that Earth, Flame and Frost Shock share one cooldown
// (shaman_audit.md 7.11).
//
// After any shock all three wait 6 sec, and 5 sec with Reverberation 5/5. The client gives the
// three shocks one 6 sec cooldown, and Reverberation takes 0.2 sec a point off it. On the beta
// two different shocks never came less than 5.9 sec apart, though our rotation casts a shock as
// soon as it can.
func TestOrcShamanSharedShockCooldown(t *testing.T) {
	for _, c := range []struct {
		talents  string
		cooldown time.Duration
	}{{"", 6 * time.Second}, {"0005", 5 * time.Second}} {
		sim, enh := newShamanSim(newOrcShaman(60, c.talents, &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 10)
		shocks := []*core.Spell{
			topRank(t, "Earth Shock", enh.EarthShock),
			topRank(t, "Flame Shock", enh.FlameShock),
			topRank(t, "Frost Shock", enh.FrostShock),
		}
		at(sim, 1, func(sim *core.Simulation) {
			for _, first := range shocks {
				for _, shock := range shocks {
					shock.CD.Reset()
					if shock.SharedCD.Timer != nil {
						shock.SharedCD.Reset()
					}
				}
				enh.AddMana(sim, enh.MaxMana()-enh.CurrentMana(), enh.GetManaNotCastingMetrics())
				castNow(t, sim, enh, first)
				for _, shock := range shocks {
					if readyAt := shock.ReadyAt(); readyAt != sim.CurrentTime+c.cooldown {
						t.Errorf("talents %q: after %s, %s is ready at %s, want %s", c.talents, first.ActionID, shock.ActionID, readyAt, sim.CurrentTime+c.cooldown)
					}
				}
			}
		})
		runSim(sim)
	}
}
