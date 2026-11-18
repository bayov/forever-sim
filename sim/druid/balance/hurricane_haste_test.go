package balance

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
)

// Under Forever cast speed shortens Hurricane's ticks, rounded to the ms, and Hurricane keeps its 10 s,
// gaining the ticks that fit: wowsims/forever's rule (dot.go CalcTickPeriod, hurricane.go), the user's
// ruling of 2026-11-18, to verify against a log. 20% cast speed: 1 s / 1.2 = 833 ms, 10 s / 833 ms =
// 12 ticks over 9.996 s. Moonfire's dot, which theirs leaves unhasted, keeps its 3 s ticks, and under
// Classic rules Hurricane is unhasted.
func TestHurricaneCastSpeed(t *testing.T) {
	cases := []struct {
		name         string
		ruleset      proto.Ruleset
		castSpeed    float64
		wantPeriod   time.Duration
		wantTicks    int32
		wantDuration time.Duration
	}{
		{"Forever unhasted", proto.Ruleset_RulesetForever, 1, time.Second, 10, 10 * time.Second},
		{"Forever 20% cast speed", proto.Ruleset_RulesetForever, 1.2, 833 * time.Millisecond, 12, 9996 * time.Millisecond},
		{"Classic 20% cast speed", proto.Ruleset_RulesetClassic, 1.2, time.Second, 10, 10 * time.Second},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			player := &proto.Player{
				Race:      proto.Race_RaceTauren,
				Class:     proto.Class_ClassDruid,
				Level:     60,
				Equipment: &proto.EquipmentSpec{},
				Rotation:  &proto.APLRotation{},
				Spec:      &proto.Player_BalanceDruid{BalanceDruid: &proto.BalanceDruid{Options: &proto.BalanceDruid_Options{}}},
			}
			sim := core.NewSim(&proto.RaidSimRequest{
				Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
				Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
				SimOptions: &proto.SimOptions{Ruleset: c.ruleset, RandomSeed: 1},
			}, simsignals.CreateSignals())
			sim.Reset()
			sim.PrePull()
			moonkin := sim.Raid.Parties[0].Players[0].(*BalanceDruid)
			target := sim.GetTargetUnit(0)
			moonkin.MultiplyCastSpeed(c.castSpeed)

			moonfire := topRank(moonkin.Moonfire)
			moonfire.Cost.Multiplier = 0
			moonkin.SetGCDTimer(sim, sim.CurrentTime)
			if !moonfire.Cast(sim, target) {
				t.Fatalf("Moonfire not cast")
			}
			if got := moonfire.Dot(target).TickPeriod(); got != 3*time.Second {
				t.Errorf("Moonfire ticks every %v, want 3s", got)
			}

			hurricane := topRank(moonkin.Hurricane)
			hurricane.Cost.Multiplier = 0
			moonkin.SetGCDTimer(sim, sim.CurrentTime)
			if !hurricane.Cast(sim, target) {
				t.Fatalf("Hurricane not cast")
			}
			dot := hurricane.AOEDot()
			if got := dot.TickPeriod(); got != c.wantPeriod {
				t.Errorf("Hurricane ticks every %v, want %v", got, c.wantPeriod)
			}
			if got := dot.Duration; got != c.wantDuration {
				t.Errorf("Hurricane lasts %v, want %v", got, c.wantDuration)
			}
			until := sim.CurrentTime + 11*time.Second
			sim.AddPendingAction(&core.PendingAction{NextActionAt: until, OnAction: func(*core.Simulation) {}})
			for sim.CurrentTime < until && !sim.Step() {
			}
			metrics := hurricane.SpellMetrics[target.UnitIndex]
			if got := metrics.Ticks + metrics.Misses; got != c.wantTicks {
				t.Errorf("Hurricane rolled %d ticks and misses, want %d", got, c.wantTicks)
			}
		})
	}
}
