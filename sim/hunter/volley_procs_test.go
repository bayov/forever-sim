package hunter

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
)

// Volley activates no procs (Blizzard hotfix 2026-11-11). The one proc that saw it before was the
// Undead hunter's Touch of the Grave, a CallbackOnCastComplete trigger on ProcMaskDirect, which
// takes in Volley's ProcMaskRangedSpecial. Twenty Volleys, each past the 15 sec cooldown, set off
// no drain; twenty Multi-Shots, on the same ranged-special mask, do.
func TestForeverVolleyActivatesNoProcs(t *testing.T) {
	drains := func(shot func(*Hunter) *core.Spell) int {
		player := &proto.Player{
			Class:              proto.Class_ClassHunter,
			Race:               proto.Race_RaceUndead,
			Level:              60,
			Equipment:          core.GetGearSet("../../ui/hunter/gear_sets", "level60_phase1_synthetic_ranged").GearSet,
			Spec:               PlayerOptionsNoPet,
			Rotation:           &proto.APLRotation{},
			DistanceFromTarget: 20,
		}
		sim := core.NewSim(&proto.RaidSimRequest{
			Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter:  &proto.Encounter{Duration: 600, Targets: []*proto.Target{core.NewDefaultTarget()}},
			SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
		}, simsignals.CreateSignals())
		sim.Reset()
		sim.PrePull()
		hunter := sim.Raid.Parties[0].Players[0].(HunterAgent).GetHunter()
		manaMetrics := hunter.NewManaMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionManaRegen})
		target := sim.GetTargetUnit(0)
		stepTo := func(at time.Duration) {
			sim.AddPendingAction(&core.PendingAction{NextActionAt: at, OnAction: func(*core.Simulation) {}})
			for i := 0; sim.CurrentTime < at; i++ {
				if i > 100000 {
					t.Fatalf("the sim did not reach %v", at)
				}
				sim.Step()
			}
		}
		stepTo(time.Second)
		for i := 0; i < 20; i++ {
			hunter.AddMana(sim, hunter.MaxMana(), manaMetrics)
			if !shot(hunter).Cast(sim, target) {
				t.Fatalf("%v did not cast at %v", shot(hunter).ActionID, sim.CurrentTime)
			}
			// No Auto Shots: Touch of the Grave also rolls on those.
			hunter.AutoAttacks.CancelAutoSwing(sim)
			stepTo(sim.CurrentTime + 16*time.Second)
		}
		drain := hunter.GetSpell(core.ActionID{SpellID: 1260189})
		if drain == nil {
			t.Fatal("the Undead hunter has no Touch of the Grave")
		}
		casts := 0
		for _, m := range drain.SpellMetrics {
			casts += int(m.Casts)
		}
		return casts
	}
	if got := drains(func(h *Hunter) *core.Spell { return h.Volley }); got != 0 {
		t.Errorf("Volley set off Touch of the Grave %d times, want 0", got)
	}
	if got := drains(func(h *Hunter) *core.Spell { return h.MultiShot }); got == 0 {
		t.Errorf("Multi-Shot never set off Touch of the Grave in 20 casts")
	}
}
