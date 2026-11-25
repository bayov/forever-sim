package core

import (
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
)

// Under Forever an extra attack swings at once and keeps the main hand's schedule. Between its
// proc and its swing the main hand's swingAt points at the extra attack, but the rotation's time to
// the next main hand swing (Auto Time To Next, MainHand, Melee and Any) reads the kept schedule:
// read as the next swing, a main hand that just swung looked about to swing again, and rotopt's
// wfGate (a special only 0.1 s or more after a swing) let a strike through in Windfury Totem's
// 100 ms after the Heroic Strike that procced it (1fed43cb).
func TestForeverAutoTimeToNextKeepsSchedule(t *testing.T) {
	sim, fw := setupWindfurySim(t)
	aa := &fw.AutoAttacks
	scheduled := aa.MainhandSwingAt()
	if scheduled <= sim.CurrentTime {
		t.Fatalf("the main hand's next swing is at %v, now is %v: want a swing still to come", scheduled, sim.CurrentTime)
	}
	aa.ExtraMHAttack(sim, 1, ActionID{SpellID: WindfuryTotemSpellId[3]}, ActionID{SpellID: 11567})
	if got := aa.MainhandSwingAt(); got != sim.CurrentTime {
		t.Fatalf("after the extra attack's proc the main hand swings at %v, want now (%v)", got, sim.CurrentTime)
	}
	if got := aa.MainhandScheduledSwingAt(); got != scheduled {
		t.Errorf("the main hand's scheduled swing is %v with an extra attack pending, want the kept %v", got, scheduled)
	}
	for _, kind := range []proto.APLValueAutoTimeToNext_AttackType{
		proto.APLValueAutoTimeToNext_MainHand, proto.APLValueAutoTimeToNext_Melee, proto.APLValueAutoTimeToNext_Any,
	} {
		v := &APLValueAutoTimeToNext{unit: &fw.Unit, autoType: kind}
		if got, want := v.GetDuration(sim), scheduled-sim.CurrentTime; got != want {
			t.Errorf("Auto Time To Next (%s) reads %v with an extra attack pending, want %v", kind, got, want)
		}
	}
}
