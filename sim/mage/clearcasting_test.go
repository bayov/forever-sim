package mage

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
)

// newLongMageSim is newMageSim with a fight long enough for thousands of casts.
func newLongMageSim(t *testing.T, talents string) (*core.Simulation, *Mage) {
	player := &proto.Player{
		Race:          proto.Race_RaceGnome,
		Class:         proto.Class_ClassMage,
		Level:         60,
		TalentsString: talents,
		Equipment:     &proto.EquipmentSpec{},
		Spec:          &proto.Player_Mage{Mage: &proto.Mage{Options: &proto.Mage_Options{}}},
		Rotation:      &proto.APLRotation{},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 20000, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*Mage)
}

func advanceMageSim(sim *core.Simulation, by time.Duration) {
	until := sim.CurrentTime + by
	sim.AddPendingAction(&core.PendingAction{NextActionAt: until, OnAction: func(*core.Simulation) {}})
	for sim.CurrentTime < until && !sim.Step() {
	}
}

// Arcane Concentration rolls its chance once per Arcane Missiles channel in all, split across the
// cast and every missile (Blizzard hotfixes of 2026-11-09), not the full chance on each missile.
func TestMageArcaneConcentrationSplitsAcrossMissiles(t *testing.T) {
	// Arcane Concentration 5.
	sim, mage := newLongMageSim(t, "000005")
	target := mage.CurrentTarget
	missiles := topRank(t, "Arcane Missiles", mage.ArcaneMissiles)
	missiles.Cost.Multiplier = 0

	const channels = 3000
	procced := 0
	for i := 0; i < channels; i++ {
		mage.ClearcastingAura.Deactivate(sim)
		mage.SetGCDTimer(sim, sim.CurrentTime)
		if !missiles.Cast(sim, target) {
			t.Fatalf("Arcane Missiles didn't cast (channel %d at %v)", i, sim.CurrentTime)
		}
		advanceMageSim(sim, 6*time.Second)
		if mage.ClearcastingAura.IsActive() {
			procced++
		}
	}
	// About 9.6% of channels (six rolls of 10%/6); the full 10% on each of five missiles was 41%.
	if rate := float64(procced) / channels; rate < 0.07 || rate > 0.125 {
		t.Errorf("Clearcasting after %.1f%% of Arcane Missiles channels, want about 9.6%%", 100*rate)
	}
}

// Flamestrike's burn doesn't use up Clearcasting (Blizzard hotfixes of 2026-11-09); only its cast does.
func TestMageFlamestrikeTicksKeepClearcasting(t *testing.T) {
	// Arcane Concentration 5.
	sim, mage := newMageSim(t, "000005")
	target := mage.CurrentTarget
	flamestrike := topRank(t, "Flamestrike", mage.Flamestrike)

	mage.SetGCDTimer(sim, sim.CurrentTime)
	if !flamestrike.Cast(sim, target) {
		t.Fatalf("Flamestrike didn't cast")
	}
	advanceMageSim(sim, 3500*time.Millisecond) // past the cast and into the burn
	mage.ClearcastingAura.Activate(sim)
	advanceMageSim(sim, 4*time.Second) // two more burn ticks
	if !mage.ClearcastingAura.IsActive() {
		t.Errorf("Flamestrike's burn used up Clearcasting")
	}
}
