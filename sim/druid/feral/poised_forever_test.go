package feral

import (
	"testing"
)

// TestCatPoised checks Poised (client 1.60.1.70753, 1328827): a druid at full Energy at the pull
// has one charge that takes 10 off the next Shred, and none after it; one that came in short of
// full Energy has none.
func TestCatPoised(t *testing.T) {
	sim, cat := newCatSim(t, catTalents)
	for sim.NextEventTime() <= 0 && !sim.Step() {
	}
	if !cat.PoisedAura.IsActive() {
		t.Fatal("no Poised at the pull at full Energy")
	}
	before := cat.CurrentEnergy()
	cat.SetGCDTimer(sim, sim.CurrentTime)
	if !cat.Shred.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Shred not cast")
	}
	if spent := before - cat.CurrentEnergy(); !near(spent, 42-10) && !near(spent, 0.8*(42-10)) {
		t.Errorf("Shred under Poised spent %.2f Energy, want 32 (or 80%% of it on a miss)", spent)
	}
	if cat.PoisedAura.IsActive() {
		t.Error("Poised still up after the Shred")
	}
	if got := cat.Shred.Cost.GetCurrentCost(); !near(got, 42) {
		t.Errorf("Shred costs %.2f after Poised, want 42", got)
	}

	sim, cat = newCatSim(t, catTalents)
	cat.SpendEnergy(sim, 1, cat.Shred.EnergyMetrics())
	for sim.NextEventTime() <= 0 && !sim.Step() {
	}
	if cat.PoisedAura.IsActive() {
		t.Error("Poised at the pull short of full Energy")
	}
}
