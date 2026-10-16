package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

// TestLevelingStatWeights checks that a leveling session's stat weights are by mobs killed.
//
// Attack power moves by 10 there, so its weight is the mobs a session +10 attack power adds
// over -10, per point. We sim both sides ourselves with the options the stat weights give their
// sims (half the iterations each, labeled random numbers, exact damage), and the weight has to
// come out the same.
func TestLevelingStatWeights(t *testing.T) {
	const iterations = 40
	req := levelingRequest(t, &proto.LevelingOptions{SessionLength: 300}, iterations)
	player := req.Raid.Parties[0].Players[0]
	res := core.StatWeights(&proto.StatWeightsRequest{
		Player:          googleProto.Clone(player).(*proto.Player),
		PartyBuffs:      &proto.PartyBuffs{},
		RaidBuffs:       &proto.RaidBuffs{},
		Debuffs:         &proto.Debuffs{},
		Tanks:           req.Raid.Tanks,
		Encounter:       req.Encounter,
		SimOptions:      &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1, Iterations: iterations},
		StatsToWeigh:    []proto.Stat{proto.Stat_StatAttackPower},
		EpReferenceStat: proto.Stat_StatAttackPower,
	})
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}

	kills := func(ap float64) float64 {
		r := googleProto.Clone(req).(*proto.RaidSimRequest)
		r.SimOptions.Iterations = iterations / 2
		r.SimOptions.UseLabeledRands = true
		r.SimOptions.ExactDamage = true
		bonus := make([]float64, len(proto.Stat_name))
		bonus[proto.Stat_StatAttackPower] = ap
		r.Raid.Parties[0].Players[0].BonusStats = &proto.UnitStats{Stats: bonus, PseudoStats: make([]float64, len(proto.PseudoStat_name))}
		result := core.RunRaidSim(r)
		if result.Error != nil {
			t.Fatal(result.Error.Message)
		}
		return result.EncounterMetrics.Leveling.MobsKilled
	}
	want := (kills(10) - kills(-10)) / 20
	got := res.Dps.Weights.Stats[proto.Stat_StatAttackPower]
	if want <= 0 || math.Abs(got-want) > 1e-9 {
		t.Errorf("attack power weight %.6f mobs a point, want %.6f (the kills at +10 and -10)", got, want)
	}
}
