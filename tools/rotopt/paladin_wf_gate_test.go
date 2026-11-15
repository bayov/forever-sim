package main

import (
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

// paladinRetLogTimes runs the twist report request once on a fixed 1800 sec fight with the
// paladin_ret template built from its preset knobs plus extra, and returns the times of each log
// line matching each pattern.
func paladinRetLogTimes(t *testing.T, extra Knobs, patterns ...string) [][]float64 {
	t.Helper()
	data, err := os.ReadFile("../dps_report/req/paladin__level_60.json")
	if err != nil {
		t.Fatal(err)
	}
	rsr := &proto.RaidSimRequest{}
	if err := protojson.Unmarshal(data, rsr); err != nil {
		t.Fatal(err)
	}
	k := defaultKnobs(paladinRet{})
	for name, v := range map[string]float64{"jotc": 0, "consecration": 1, "consecrationFirst": 1, "consecrationMana": 20,
		"twist": 1, "twistStacks": 1, "twistBack": 3, "twistWindow": 25, "hammerOfWrathMana": 0, "cooldownsFirst": 1,
		"trinketExecute": 1, "hammerOfWrathFirst": 1, "swiftJudgement": 0, "holyStrikeFirst": 1} {
		k[name] = v
	}
	for name, v := range extra {
		k[name] = v
	}
	rsr.Raid.Parties[0].Players[0].Rotation = paladinRet{}.Build(k)
	rsr.Encounter.Duration = 1800
	rsr.Encounter.DurationVariation = 0
	rsr.SimOptions.Iterations = 1
	rsr.SimOptions.Debug = true
	res := core.RunRaidSim(rsr)
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	out := make([][]float64, len(patterns))
	for i, p := range patterns {
		re := regexp.MustCompile(`\[(-?[0-9.]+)\] \[[^]]*\] ` + regexp.QuoteMeta(p))
		for _, m := range re.FindAllStringSubmatch(res.Logs, -1) {
			v, _ := strconv.ParseFloat(m[1], 64)
			out[i] = append(out[i], v)
		}
	}
	return out
}

const (
	logHolyStrike    = "Casting {SpellID: 10333}"
	logWindfuryProc  = "Aura gained: {SpellID: 10610}"
	logMainHandSwing = "Casting {OtherID: 7, Tag: 1}"
)

// countNear counts the times in a that fall in [b+from, b+to] for some time b in bs.
func countNear(a, bs []float64, from, to float64) int {
	n := 0
	for _, x := range a {
		for _, b := range bs {
			if x >= b+from && x <= b+to {
				n++
				break
			}
		}
	}
	return n
}

// The Windfury Totem timing knobs of paladin_ret: wfIcdGate holds Holy Strike while the totem's
// 100 ms internal cooldown runs (it starts on a proc, the Windfury Attack buff 10610), and swingGap
// holds it while the next main hand swing is that many ms away. The log rounds times to 10 ms.
func TestPaladinRetWindfuryGates(t *testing.T) {
	sim.RegisterAll()
	off := paladinRetLogTimes(t, Knobs{}, logHolyStrike, logWindfuryProc, logMainHandSwing)
	if len(off[0]) < 150 || len(off[1]) < 60 {
		t.Fatalf("%d Holy Strikes and %d Windfury procs in 1800 sec, want more", len(off[0]), len(off[1]))
	}
	if n := countNear(off[0], off[2], -0.19, -0.01); n == 0 {
		t.Errorf("without swingGap no Holy Strike came within 200 ms before a swing; the test can't tell")
	}

	icd := paladinRetLogTimes(t, Knobs{"wfIcdGate": 1}, logHolyStrike, logWindfuryProc)
	if n := countNear(icd[0], icd[1], 0.001, 0.09); n > 0 {
		t.Errorf("wfIcdGate=1: %d Holy Strikes inside the totem's 100 ms internal cooldown", n)
	}

	gap := paladinRetLogTimes(t, Knobs{"swingGap": 200}, logHolyStrike, logMainHandSwing)
	if n := countNear(gap[0], gap[1], -0.18, -0.01); n > 0 {
		t.Errorf("swingGap=200: %d Holy Strikes less than 200 ms before a main hand swing", n)
	}
}
