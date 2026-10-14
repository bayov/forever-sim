package enhancement

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestOrcShamanRageOfTheFarseer checks that Rage of the Farseer gives 30% attack speed for
// 25 sec and no cast speed. The 2026-09-24 beta build (1.60.1.70009) took the cast speed
// away (its patch notes, as foreverchanges.pro lists them), and Blizzard's Mage and Shaman
// deep dive (2026-10-07) gives only the attack speed. Its cooldown is 3 min, as the client
// (70338) and build 70009's tooltip say (shaman_audit.md 8.34). It fails when the buff speeds
// up spellcasting again, or when its attack speed, duration or cooldown change.
func TestOrcShamanRageOfTheFarseer(t *testing.T) {
	sim, enh := newShamanSim(newOrcShaman(60, "-000000000000100051-00000000100001", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 200)
	rage := enh.GetSpell(core.ActionID{SpellID: 425336})
	if rage == nil {
		t.Fatal("no Rage of the Farseer")
	}

	var swingBefore, castBefore float64
	at(sim, 1, func(sim *core.Simulation) {
		swingBefore, castBefore = enh.SwingSpeed(), enh.CastSpeed
		if !rage.Cast(sim, enh.CurrentTarget) {
			t.Fatal("Rage of the Farseer didn't cast")
		}
		if got := enh.SwingSpeed() / swingBefore; math.Abs(got-1.3) > 1e-9 {
			t.Errorf("attack speed went up %.4fx, want 1.3x", got)
		}
		if enh.CastSpeed != castBefore {
			t.Errorf("cast speed went from %.4f to %.4f, want no change", castBefore, enh.CastSpeed)
		}
		if got := rage.CD.TimeToReady(sim); got != 3*time.Minute {
			t.Errorf("Rage of the Farseer is ready again in %s, want 3m", got)
		}
	})
	at(sim, 25.5, func(sim *core.Simulation) {
		if got := enh.SwingSpeed() / swingBefore; math.Abs(got-1.3) > 1e-9 {
			t.Errorf("attack speed is up %.4fx 24.5 sec after the cast, want 1.3x", got)
		}
	})
	at(sim, 26.5, func(sim *core.Simulation) {
		if math.Abs(enh.SwingSpeed()-swingBefore) > 1e-9 {
			t.Errorf("attack speed is %.4f after the buff ran out, want %.4f", enh.SwingSpeed(), swingBefore)
		}
	})
	at(sim, 180.9, func(sim *core.Simulation) {
		if rage.IsReady(sim) {
			t.Errorf("Rage of the Farseer is ready before 3 min")
		}
	})
	at(sim, 181.1, func(sim *core.Simulation) {
		if !rage.IsReady(sim) {
			t.Errorf("Rage of the Farseer isn't ready 3 min after the cast")
		}
	})
	runSim(sim)
}
