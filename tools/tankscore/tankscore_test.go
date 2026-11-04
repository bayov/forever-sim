package tankscore

import (
	"os"
	"testing"

	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	sim.RegisterAll()
}

func scoredRequest(t *testing.T, path string, hps float64) *proto.RaidSimRequest {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	rsr := &proto.RaidSimRequest{}
	if err := protojson.Unmarshal(data, rsr); err != nil {
		t.Fatal(err)
	}
	rsr.SimOptions.Iterations = 20
	rsr.SimOptions.RandomSeed = 7
	MakeTank(rsr, &proto.HealingModel{Hps: hps, CadenceSeconds: 2, CadenceVariation: 1}, &proto.TankScoring{})
	return rsr
}

func runScored(t *testing.T, rsr *proto.RaidSimRequest) *proto.UnitMetrics {
	t.Helper()
	res := core.RunRaidSim(rsr)
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	return res.RaidMetrics.Parties[0].Players[0]
}

// A scored tank faces the boss, whatever its request says: the boss parries and blocks its attacks
// (tank_mechanics-20261104-1).
func TestMakeTankInFront(t *testing.T) {
	rsr := scoredRequest(t, "../tankbench/req/warrior_protection_balanced.json", 600)
	rsr.Raid.Parties[0].Players[0].InFrontOfTarget = false
	MakeTank(rsr, rsr.Raid.Parties[0].Players[0].HealingModel, &proto.TankScoring{})
	if !rsr.Raid.Parties[0].Players[0].InFrontOfTarget {
		t.Fatal("MakeTank left the tank behind the boss")
	}
	var parries, blocks int32
	for _, action := range runScored(t, rsr).Actions {
		for _, target := range action.Targets {
			parries += target.Parries
			blocks += target.Blocks
		}
	}
	if parries == 0 {
		t.Errorf("the boss parried none of the scored tank's attacks")
	}
	t.Logf("boss parried %d and blocked %d of the tank's attacks", parries, blocks)
}

// Nothing in a scoring run reads the tank's health, so its threat and damage taken don't depend on
// the stand-in healer's HPS, even with a survival cooldown (Greater Stoneshield Potion) set to fire
// under 90% health (tank_mechanics-20261104-8).
func TestScoringRunReadsNoHealth(t *testing.T) {
	for _, path := range []string{
		"../tankbench/req/shaman_warden_balanced.json",
		"../tankbench/req/druid_bear_balanced.json",
		"../tankbench/req/paladin_protection_balanced.json",
	} {
		var dtps, safe []float64
		for _, hps := range []float64{200, 5000} {
			rsr := scoredRequest(t, path, hps)
			tank := rsr.Raid.Parties[0].Players[0]
			tank.Cooldowns = &proto.Cooldowns{HpPercentForDefensives: 0.9}
			tank.Consumes.DefaultPotion = proto.Potions_GreaterStoneshieldPotion
			m := runScored(t, rsr)
			dtps = append(dtps, m.Dtps.Avg)
			safe = append(safe, m.TankScore.SafeDps.Avg)
		}
		if dtps[0] != dtps[1] || safe[0] != safe[1] {
			t.Errorf("%s: DTPS %v and SafeDPS %v move with the HPS", path, dtps, safe)
		}
	}
}
