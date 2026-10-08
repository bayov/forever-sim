package enhancement

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestOrcShamanHaste checks how Flurry and Rage of the Farseer stack, and what they do to the
// swing in progress at level 60.
//
// Attack speed effects multiply, as in Classic. When one comes up or drops mid-swing, the
// rest of the swing scales with the new speed. Neither effect touches cast speed.
func TestOrcShamanHaste(t *testing.T) {
	// Flurry 5/5 and Rage of the Farseer.
	sim, enh := newShamanSim(newOrcShaman(60, "-000000000005000001", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 20)
	flurry := enh.GetAuraByID(core.ActionID{SpellID: 16280})
	farseer := enh.GetAuraByID(core.ActionID{SpellID: 425336})
	farseerSpell := enh.GetSpell(core.ActionID{SpellID: 425336})
	if flurry == nil || farseer == nil || farseerSpell == nil {
		t.Fatalf("missing Flurry or Rage of the Farseer")
	}

	at(sim, 5, func(sim *core.Simulation) {
		// We check halfway through a swing, after our crits are done giving Flurry.
		flurry.Deactivate(sim)
		swing := enh.AutoAttacks.MainhandSwingSpeed()
		checkAt := enh.AutoAttacks.MainhandSwingAt() - swing/2
		for checkAt <= sim.CurrentTime {
			checkAt += swing
		}
		sim.AddPendingAction(&core.PendingAction{NextActionAt: checkAt, OnAction: func(sim *core.Simulation) {
			flurry.Deactivate(sim)
			left := enh.AutoAttacks.MainhandSwingAt() - sim.CurrentTime
			check := func(label string, speed float64) {
				if got, want := enh.AutoAttacks.MainhandSwingSpeed(), time.Duration(float64(swing)/speed); (got - want).Abs() > time.Microsecond {
					t.Errorf("%s: the swing takes %s, want %s", label, got, want)
				}
				if got, want := enh.AutoAttacks.MainhandSwingAt()-sim.CurrentTime, time.Duration(float64(left)/speed); (got - want).Abs() > time.Microsecond {
					t.Errorf("%s: the swing in progress lands in %s, want %s", label, got, want)
				}
				if enh.CastSpeed != 1 {
					t.Errorf("%s: cast speed is %.3f, want 1", label, 1/enh.CastSpeed)
				}
			}

			flurry.Activate(sim)
			flurry.SetStacks(sim, 3)
			check("Flurry", 1.25)
			if !farseerSpell.Cast(sim, enh.CurrentTarget) {
				t.Fatalf("Rage of the Farseer didn't cast")
			}
			check("Flurry and Rage of the Farseer", 1.25*1.3)
			flurry.Deactivate(sim)
			check("Rage of the Farseer after Flurry drops", 1.3)
			farseer.Deactivate(sim)
			check("no haste", 1)
		}})
	})
	runSim(sim)
}
