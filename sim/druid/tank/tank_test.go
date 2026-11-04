package tank

import (
	"math"
	"os"
	"testing"

	_ "github.com/wowsims/classic/sim/common"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	RegisterFeralTankDruid()
}

// TestForeverBear runs the tank bench's bear (tools/tankbench/req/druid_bear_balanced.json) against its
// boss with a stand-in healer and checks that the Forever bear is put together: Dire Bear Form's
// health and armor, the rotation's abilities, its own Demoralizing Roar, and a tank score.
// runBench runs the tank bench's bear (tools/tankbench/req/druid_bear_balanced.json) for 200 iterations of
// 120 sec against its boss with a 900 HPS stand-in healer, its own Demoralizing Roar and a tank
// score, after mod.
func runBench(t *testing.T, mod func(*proto.RaidSimRequest)) (*proto.RaidSimResult, float64) {
	t.Helper()
	data, err := os.ReadFile("../../../tools/tankbench/req/druid_bear_balanced.json")
	if err != nil {
		t.Fatal(err)
	}
	rsr := &proto.RaidSimRequest{}
	if err := protojson.Unmarshal(data, rsr); err != nil {
		t.Fatal(err)
	}
	const iterations, seconds = 200, 120
	rsr.Raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
	rsr.Raid.Debuffs.DemoralizingRoar = proto.TristateEffect_TristateEffectMissing
	rsr.Raid.Debuffs.DemoralizingShout = proto.TristateEffect_TristateEffectMissing
	rsr.Encounter.Duration, rsr.Encounter.DurationVariation = seconds, 0
	rsr.SimOptions.Iterations = iterations
	rsr.SimOptions.RandomSeed = 7
	player := rsr.Raid.Parties[0].Players[0]
	player.HealingModel = &proto.HealingModel{Hps: 900, CadenceSeconds: 2, CadenceVariation: 1}
	player.TankScoring = &proto.TankScoring{}
	mod(rsr)
	res := core.RunRaidSimConcurrent(rsr)
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	return res, iterations
}

func TestForeverBear(t *testing.T) {
	const iterations, seconds = 200, 120
	res, _ := runBench(t, func(*proto.RaidSimRequest) {})
	m := res.RaidMetrics.Parties[0].Players[0]

	// Dire Bear Form: 1,240 more health at 60 and 4.6 times the armor from items.
	sheet := m.TankSheet
	if sheet == nil {
		t.Fatal("no tank sheet: the boss didn't attack the bear")
	}
	if sheet.Health < 8000 || sheet.Armor < 7000 {
		t.Errorf("health %.0f and armor %.0f in Dire Bear Form, want over 8,000 and 7,000 on this gear", sheet.Health, sheet.Armor)
	}
	if sheet.Block != 0 || sheet.Parry != 0 {
		t.Errorf("a bear can't block or parry, the sheet has %.3f and %.3f", sheet.Block, sheet.Parry)
	}

	casts := map[int32]float64{}
	for _, a := range m.Actions {
		for _, tgt := range a.Targets {
			casts[a.Id.GetSpellId()] += float64(tgt.Casts) / iterations
		}
	}
	for id, name := range map[int32]string{9881: "Maul", 1235827: "Lacerate", 1238070: "Primal Bite", 9898: "Demoralizing Roar"} {
		if casts[id] < 1 {
			t.Errorf("%s cast %.1f times a fight, want it in the rotation", name, casts[id])
		}
	}
	// Primal Bite's 6 sec cooldown caps it at 21 a 120 sec fight, plus one for each reset by the Idol of
	// the Ursine Twins (10% of the Lacerates that land).
	if casts[1238070] > seconds/6+1+0.1*casts[1235827] {
		t.Errorf("Primal Bite cast %.1f times in %d sec, over its cooldown", casts[1238070], seconds)
	}

	// The bear's own Demoralizing Roar is on the boss.
	uptime := 0.0
	for _, a := range res.EncounterMetrics.Targets[0].Auras {
		if a.Id.GetSpellId() == 9898 {
			uptime = a.UptimeSecondsAvg / seconds
		}
	}
	// Roar's line is below Lacerate's keep-up (rotopt's roarAfterLacerate), so the first Roar waits for
	// Lacerate's 5 stacks and an early refresh can wait a GCD or two.
	if uptime < 0.8 {
		t.Errorf("Demoralizing Roar up %.0f%% of the fight, want it kept up", 100*uptime)
	}

	if m.Threat.Avg <= 0 || m.Dps.Avg <= 0 {
		t.Errorf("TPS %.0f and DPS %.0f, want both above 0", m.Threat.Avg, m.Dps.Avg)
	}
	if got := len(m.TankScore.GetSurviveHps()); got != iterations {
		t.Errorf("the tank score replayed %d iterations, want %d", got, iterations)
	}
	if math.IsNaN(m.Dtps.Avg) || m.Dtps.Avg <= 0 {
		t.Errorf("DTPS %.0f, want the boss to hit the bear", m.Dtps.Avg)
	}
}

// TestForeverBearFormThreat checks the November 2 development notes' Bear Form threat bonus, 115%
// (was 50%; clients 70753 and 70755 unchanged, standards 37), under Forever, and the client's 50%
// under Classic.
func TestForeverBearFormThreat(t *testing.T) {
	data, err := os.ReadFile("../../../tools/tankbench/req/druid_bear_balanced.json")
	if err != nil {
		t.Fatal(err)
	}
	for ruleset, want := range map[proto.Ruleset]float64{proto.Ruleset_RulesetForever: 2.15, proto.Ruleset_RulesetClassic: 1.5} {
		rsr := &proto.RaidSimRequest{}
		if err := protojson.Unmarshal(data, rsr); err != nil {
			t.Fatal(err)
		}
		rsr.SimOptions.Ruleset = ruleset
		sim := core.NewSim(rsr, simsignals.CreateSignals())
		sim.Reset()
		bear := sim.Raid.Parties[0].Players[0].(*FeralTankDruid)
		if !bear.BearFormAura.IsActive() {
			t.Fatal("the bear isn't in Bear Form after Reset")
		}
		inForm := bear.PseudoStats.ThreatMultiplier
		bear.CancelShapeshift(sim)
		if got := inForm / bear.PseudoStats.ThreatMultiplier; math.Abs(got-want) > 1e-9 {
			t.Errorf("%v: Bear Form threat x%.4f, want x%.2f", ruleset, got, want)
		}
	}
}
