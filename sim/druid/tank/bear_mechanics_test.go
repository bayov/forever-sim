package tank

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestThickHideDefenseChange: Thick Hide's Defense armor (0.67 a point of Defense a talent point,
// scaled by the form's armor multiplier) is read when the form comes up and stays as it was while
// the tank is in the form. Leaving the form removes exactly what was added, and entering it again
// reads the Defense of that moment.
func TestThickHideDefenseChange(t *testing.T) {
	const points, defense, extra = 3, 10.0, 20.0
	sim, bear := newBear(t, bearOpts{
		talents: talents(map[string]int{"thickHide": points}, ""),
		bonus:   stats.Stats{stats.Defense: defense},
	})
	perDefense := 0.67 * points * bear.BearArmorMultiplier()
	inForm := bear.GetStat(stats.Armor)

	bear.AddStatDynamic(sim, stats.Defense, extra)
	if got := bear.GetStat(stats.Armor); !near(got, inForm, 0.01) {
		t.Errorf("a Defense change in the form moved armor from %.2f to %.2f, it is read when the form comes up", inForm, got)
	}

	bear.BearFormAura.Deactivate(sim)
	humanoid := bear.GetStat(stats.Armor)
	ready(bear, bear.BearForm)
	bear.BearForm.Cast(sim, bear.CurrentTarget)
	if !bear.BearFormAura.IsActive() {
		t.Fatal("Bear Form didn't come back up")
	}
	if got, want := bear.GetStat(stats.Armor)-inForm, perDefense*extra; !near(got, want, 0.01) {
		t.Errorf("entering the form with %.0f more Defense added %.2f armor, want %.2f (0.67 x %d x multiplier x %.0f)", extra, got, want, points, extra)
	}

	bear.BearFormAura.Deactivate(sim)
	if got := bear.GetStat(stats.Armor); !near(got, humanoid, 0.01) {
		t.Errorf("armor out of the form is %.2f the second time, %.2f the first: Thick Hide's Defense armor leaked", got, humanoid)
	}
	// Defense at the first entry was the 10 from gear, so the first form held this much of it.
	if got, want := inForm-humanoid, perDefense*defense; got < want-0.01 {
		t.Errorf("the form's armor over humanoid is %.2f, less than Thick Hide's Defense part %.2f", got, want)
	}
}

// TestBearRageGainThreat: under Forever a Rage gain makes 5 threat a Rage times the unit's threat
// multiplier at the time (Bear Form's 2.15 here), split over the targets, booked at once.
func TestBearRageGainThreat(t *testing.T) {
	for _, targets := range []int{1, 2} {
		sim, bear := newBear(t, bearOpts{targets: targets})
		spell := bear.GetSpell(core.ActionID{OtherID: proto.OtherAction_OtherActionRageGain})
		if spell == nil {
			t.Fatal("no Rage gain spell")
		}
		metrics := bear.NewRageMetrics(core.ActionID{SpellID: 2})
		const gain = 12.0
		var before []float64
		for _, m := range spell.SpellMetrics {
			before = append(before, m.TotalThreat)
		}
		bear.AddRage(sim, gain, metrics)
		want := core.ThreatPerRageGained * formThreat * gain / float64(targets)
		for i := 0; i < targets; i++ {
			if got := spell.SpellMetrics[i].TotalThreat - before[i]; !near(got, want, 1e-6) {
				t.Errorf("%d targets: %.0f Rage in Bear Form made %.3f threat on target %d, want %.3f", targets, gain, got, i, want)
			}
		}
	}
}
