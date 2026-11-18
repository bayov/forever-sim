package hunter

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
)

// Under Forever cast speed shortens Volley's ticks, rounded to the ms, and the channel with them: the
// wave count stays 6. That is wowsims/forever's rule (dot.go CalcTickPeriod, volley.go with
// HasteReducesDuration), the user's ruling of 2026-11-18, to verify against a log. 20% cast speed:
// 1 s / 1.2 = 833 ms, 6 waves over 4.998 s, and Auto Shot waits only that long. Ranged haste does not
// shorten it, and under Classic rules Volley is unhasted.
func TestVolleyCastSpeed(t *testing.T) {
	cases := []struct {
		name         string
		ruleset      proto.Ruleset
		castSpeed    float64
		rangedSpeed  float64
		wantPeriod   time.Duration
		wantDuration time.Duration
	}{
		{"Forever unhasted", proto.Ruleset_RulesetForever, 1, 1, time.Second, 6 * time.Second},
		{"Forever 20% cast speed", proto.Ruleset_RulesetForever, 1.2, 1, 833 * time.Millisecond, 4998 * time.Millisecond},
		{"Forever 20% ranged haste", proto.Ruleset_RulesetForever, 1, 1.2, time.Second, 6 * time.Second},
		{"Classic 20% cast speed", proto.Ruleset_RulesetClassic, 1.2, 1, time.Second, 6 * time.Second},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			player := &proto.Player{
				Class:              proto.Class_ClassHunter,
				Race:               proto.Race_RaceDwarf,
				Level:              60,
				Equipment:          core.GetGearSet("../../ui/hunter/gear_sets", "level60_phase1_synthetic_ranged").GearSet,
				Spec:               PlayerOptionsNoPet,
				Rotation:           &proto.APLRotation{},
				DistanceFromTarget: 20,
			}
			sim := core.NewSim(&proto.RaidSimRequest{
				Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
				Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
				SimOptions: &proto.SimOptions{Ruleset: c.ruleset, RandomSeed: 1},
			}, simsignals.CreateSignals())
			sim.Reset()
			sim.PrePull()
			hunter := sim.Raid.Parties[0].Players[0].(HunterAgent).GetHunter()
			target := sim.GetTargetUnit(0)
			// The gear's haste rating already speeds casts up: scale to exactly c.castSpeed.
			hunter.MultiplyCastSpeed(c.castSpeed * hunter.CastSpeed)
			hunter.PseudoStats.RangedSpeedMultiplier *= c.rangedSpeed

			hunter.Volley.Cost.Multiplier = 0
			hunter.SetGCDTimer(sim, sim.CurrentTime)
			if !hunter.Volley.Cast(sim, target) {
				t.Fatalf("Volley not cast")
			}
			dot := hunter.Volley.AOEDot()
			if got := dot.TickPeriod(); got != c.wantPeriod {
				t.Errorf("Volley ticks every %v, want %v", got, c.wantPeriod)
			}
			if got := dot.Duration; got != c.wantDuration {
				t.Errorf("Volley lasts %v, want %v", got, c.wantDuration)
			}
			if got := hunter.AutoAttacks.NextRangedAttackAt(); got < sim.CurrentTime+c.wantDuration {
				t.Errorf("Auto Shot due at %v, before the channel ends at %v", got, sim.CurrentTime+c.wantDuration)
			}
			start := sim.CurrentTime
			until := start + 7*time.Second
			sim.AddPendingAction(&core.PendingAction{NextActionAt: until, OnAction: func(*core.Simulation) {}})
			for sim.CurrentTime < until && !sim.Step() {
			}
			if got := dot.TickCount; got != 6 {
				t.Errorf("Volley sent %d waves, want 6", got)
			}
		})
	}
}
