package feral

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
)

func newRaceCatSim(t *testing.T, race proto.Race) (*core.Simulation, *FeralDruid) {
	t.Helper()
	player := &proto.Player{
		Race:          race,
		Class:         proto.Class_ClassDruid,
		Level:         60,
		Equipment:     &proto.EquipmentSpec{Items: []*proto.ItemSpec{{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {Id: earthshaker}, {}, {}}},
		TalentsString: catTalents,
		Rotation:      &proto.APLRotation{},
		Spec: &proto.Player_FeralDruid{FeralDruid: &proto.FeralDruid{Options: &proto.FeralDruid_Options{
			InnervateTarget: &proto.UnitReference{},
		}}},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*FeralDruid)
}

// Ravage (9867) needs stealth and the target's back. Prowl is out of combat only, so the sim never
// casts it (orchestration/standards.md item 13), and a Night Elf's Shadowmeld (20580) in combat is
// the way in: off the global cooldown, 2 min cooldown, and Ravage ends it. Ravage costs 60 Energy
// and gives a combo point. Other races have no Shadowmeld.
func TestForeverShadowmeldIntoRavage(t *testing.T) {
	_, tauren := newRaceCatSim(t, proto.Race_RaceTauren)
	if tauren.Shadowmeld != nil {
		t.Error("a Tauren has Shadowmeld")
	}

	sim, cat := newRaceCatSim(t, proto.Race_RaceNightElf)
	target := cat.CurrentTarget
	if cat.Ravage == nil || cat.Ravage.ActionID.SpellID != 9867 || cat.Shadowmeld == nil || cat.Prowl == nil {
		t.Fatal("a level 60 Night Elf cat has no Ravage rank 4, Prowl or Shadowmeld")
	}
	cat.SetGCDTimer(sim, sim.CurrentTime)
	if cat.Prowl.CanCast(sim, target) {
		t.Error("Prowl is castable in the fight")
	}
	if cat.Ravage.CanCast(sim, target) {
		t.Error("Ravage is castable out of stealth")
	}

	if !cat.Shadowmeld.Cast(sim, target) || !cat.Stealthed() {
		t.Fatal("Shadowmeld didn't cast")
	}
	if cat.Shadowmeld.CD.TimeToReady(sim) != 2*time.Minute || !cat.GCD.IsReady(sim) {
		t.Errorf("Shadowmeld cooldown %s and global cooldown ready %v, want 2m0s and true", cat.Shadowmeld.CD.TimeToReady(sim), cat.GCD.IsReady(sim))
	}
	cat.PseudoStats.InFrontOfTarget = true
	if cat.Ravage.CanCast(sim, target) {
		t.Error("Ravage is castable in front of the target")
	}
	cat.PseudoStats.InFrontOfTarget = false

	energy := cat.CurrentEnergy()
	if !cat.Ravage.Cast(sim, target) {
		t.Fatal("Ravage didn't cast from Shadowmeld")
	}
	if cat.Stealthed() {
		t.Error("Ravage left Shadowmeld up")
	}
	if got := energy - cat.CurrentEnergy(); !near(got, 60) && !near(got, 12) {
		t.Errorf("Ravage cost %.1f Energy, want 60 (12 on a miss)", got)
	}
	if result := cat.Ravage.SpellMetrics[target.UnitIndex]; result.Dodges+result.Parries+result.Blocks > 0 {
		t.Error("Ravage was dodged, parried or blocked")
	}
}

// Any other cast ends Shadowmeld, a Shred included.
func TestForeverShadowmeldBreaksOnCast(t *testing.T) {
	sim, cat := newRaceCatSim(t, proto.Race_RaceNightElf)
	target := cat.CurrentTarget
	cat.Shadowmeld.Cast(sim, target)
	cat.SetGCDTimer(sim, sim.CurrentTime)
	if !cat.Shred.Cast(sim, target) {
		t.Fatal("Shred didn't cast")
	}
	if cat.Stealthed() {
		t.Error("Shred left Shadowmeld up")
	}
}
