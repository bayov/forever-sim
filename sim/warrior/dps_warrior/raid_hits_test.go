package dpswarrior

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestRaidDamageHits checks that the raid hits land on the warrior at the set rate and give
// Rage like any damage taken.
//
// A level 60 warrior with no gear, no stance and no rotation takes 10 hits a minute of 1500
// damage for 9 min, so about 90 hits a fight. Each one gives 2.5 × 1500 / the level 63 boss's
// Rage conversion, 15.4 Rage. With no hits there is no Rage from damage taken at all.
func TestRaidDamageHits(t *testing.T) {
	const minutes, iterations = 9, 200
	for _, hitsPerMinute := range []float64{0, 10} {
		result := core.RunRaidSim(&proto.RaidSimRequest{
			Raid: core.SinglePlayerRaidProto(&proto.Player{
				Race:      proto.Race_RaceOrc,
				Class:     proto.Class_ClassWarrior,
				Level:     60,
				Equipment: &proto.EquipmentSpec{},
				Rotation:  &proto.APLRotation{},
				Spec: &proto.Player_Warrior{Warrior: &proto.Warrior{
					Options: &proto.Warrior_Options{RaidDamageHitsPerMinute: hitsPerMinute, RaidDamagePerHit: 1500},
				}},
			}, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter:  &proto.Encounter{Duration: minutes * 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
			SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, Iterations: iterations, RandomSeed: 1},
		})
		if result.Error != nil {
			t.Fatal(result.Error.Message)
		}

		var events int32
		var gain float64
		for _, r := range result.RaidMetrics.Parties[0].Players[0].Resources {
			if r.Id.GetOtherId() == proto.OtherAction_OtherActionDamageTaken {
				events, gain = r.Events, r.Gain
			}
		}
		perFight := float64(events) / iterations
		if want := minutes * hitsPerMinute; math.Abs(perFight-want) > 0.1*want+0.01 {
			t.Errorf("%.0f hits a minute: %.1f hits a fight, want about %.0f", hitsPerMinute, perFight, want)
		}
		if events > 0 {
			if perHit, want := gain/float64(events), 1500*2.5/core.GetRageConversion(63); math.Abs(perHit-want) > 0.01 {
				t.Errorf("%.2f Rage a hit, want %.2f", perHit, want)
			}
		}
	}
}
