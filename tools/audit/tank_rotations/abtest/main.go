// Command abtest is a paired A/B run of two tankbench requests (same tank, e.g. one APL line
// apart) on the tank encounters, scored the way tools/tankbench -encounter does.
//
//	abtest [-preset max_threat|balanced|max_survival] [-iters 1000] [-seed 1] A.json B.json [encounter.json ...]
//
// The preset is inferred from A's filename suffix when not given. Both runs use the same seed
// (-seed, spaced apart by tankscore.Seed so two seeds share no iterations), so iteration k of each has
// the same rolls, and tankscore's PairedDiff pairs them.
package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/tools/tankscore"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
)

// tankbench's flag defaults.
const (
	dpsStart      = 2.0
	dpsThreat     = 1.0
	threatMembers = 2.0
	dpsSlot       = 900.0
	deathTarget   = 0.05
	window        = 6
)

func main() {
	preset := flag.String("preset", "", "max_threat|balanced|max_survival; default: A's filename suffix")
	iters := flag.Int("iters", 1000, "iterations per run")
	seed := flag.Int64("seed", 1, "random seed, spaced a billion iterations apart (tankscore.Seed) so two seeds share no iterations")
	flag.Parse()
	args := flag.Args()
	if len(args) < 2 {
		fail(fmt.Errorf("usage: abtest [-preset p] [-iters n] [-seed s] A.json B.json [encounter.json ...]"))
	}
	pathA, pathB, encs := args[0], args[1], args[2:]
	if *preset == "" {
		base := strings.TrimSuffix(filepath.Base(pathA), ".json")
		for _, p := range []string{"max_threat", "balanced", "max_survival"} {
			if strings.HasSuffix(base, p) {
				*preset = p
			}
		}
		if *preset == "" {
			fail(fmt.Errorf("can't infer the preset from %s; pass -preset", pathA))
		}
	}
	weight, ok := tankscore.Presets[strings.ReplaceAll(*preset, "_", "-")]
	if !ok {
		fail(fmt.Errorf("unknown preset %q", *preset))
	}
	if len(encs) == 0 {
		var err error
		if encs, err = filepath.Glob("tools/encounters/*.json"); err != nil || len(encs) == 0 {
			fail(fmt.Errorf("no encounter files (run from the repo root): %v", err))
		}
	}
	sim.RegisterAll()

	fmt.Printf("A=%s B=%s preset=%s iters=%d\n", filepath.Base(pathA), filepath.Base(pathB), *preset, *iters)
	fmt.Printf("%-13s %8s %8s %7s %6s %8s %8s\n", "boss", "scoreA", "scoreB", "diff", "se", "safeDiff", "needDiff")
	var sum, sumSq float64
	var n int
	unpaired := false
	for _, ep := range encs {
		enc, err := tankscore.LoadEncounter(ep)
		if err != nil {
			fail(err)
		}
		boss := strings.TrimSuffix(filepath.Base(ep), ".json")
		each := 400.0
		if enc.Healing != nil && enc.Healing.HpsEach > 0 {
			each = enc.Healing.HpsEach
		}
		obj := tankscore.Objective{Weight: weight, ThreatMembers: threatMembers, DpsSlot: dpsSlot, HealerEach: each, Death: deathTarget}
		ma := runOne(pathA, enc, *iters, *seed)
		mb := runOne(pathB, enc, *iters, *seed)
		sa, sb := obj.Score(ma, int32(*iters)), obj.Score(mb, int32(*iters))
		diff, se, paired := obj.PairedDiff(ma.TankScore, mb.TankScore)
		note := ""
		if !paired {
			diff, se = sb.Value-sa.Value, math.Hypot(sa.StdErr, sb.StdErr)
			note = " (unpaired)"
			unpaired = true
		}
		need := fmt.Sprintf("%8.0f", sb.Needed-sa.Needed)
		if sa.Needed < 0 || sb.Needed < 0 {
			need = fmt.Sprintf("%8s", "none")
		}
		fmt.Printf("%-13s %8.0f %8.0f %7.1f %6.1f %8.1f %s%s\n", boss, sa.Value, sb.Value, diff, se, sb.SafeDps-sa.SafeDps, need, note)
		if boss != "ragnaros" {
			sum += diff
			sumSq += se * se
			n++
		}
	}
	if n > 0 {
		se := math.Sqrt(sumSq) / float64(n)
		if unpaired {
			// An unpaired boss's error is that of two absolute scores, and the bosses share the
			// seed, so their scores correlate: the mean's error is about 1.25 times the independent
			// one (tankscore.SharedSeedSD, finding tank_rotations-20261104-13). Paired diffs don't need it.
			se *= tankscore.SharedSeedSD
		}
		fmt.Printf("%-13s %8s %8s %7.1f %6.1f\n", "scoring_mean", "", "", sum/float64(n), se)
	}
}

// runOne sets the request up as tankbench -encounter does and returns the tank's metrics.
func runOne(path string, enc *tankscore.Encounter, iters int, seed int64) *proto.UnitMetrics {
	data, err := os.ReadFile(path)
	if err != nil {
		fail(err)
	}
	rsr := &proto.RaidSimRequest{}
	if err := protojson.Unmarshal(data, rsr); err != nil {
		fail(fmt.Errorf("%s: %w", path, err))
	}
	rsr.SimOptions.Iterations = int32(iters)
	rsr.SimOptions.RandomSeed = tankscore.Seed(seed)
	tankscore.UseEncounter(rsr, enc)
	h := enc.Healing
	var hm *proto.HealingModel
	if h != nil && h.Healers > 0 && h.HpsEach > 0 && h.CastSeconds > 0 {
		hm = h.Model()
	} else { // tankbench's flag defaults: 600 HPS every 2 +- 1 s
		hm = &proto.HealingModel{Hps: 600, CadenceSeconds: 2, CadenceVariation: 1, BurstWindow: window}
	}
	tankscore.MakeTank(rsr, hm, &proto.TankScoring{DpsStart: dpsStart, ThreatPerDamage: dpsThreat})
	res := core.RunRaidSimConcurrent(googleProto.Clone(rsr).(*proto.RaidSimRequest))
	if res.Error != nil {
		fail(fmt.Errorf("%s: %s", path, res.Error.Message))
	}
	m := res.RaidMetrics.Parties[0].Players[0]
	if m.TankScore == nil {
		fail(fmt.Errorf("%s: no tank score (the boss didn't attack the tank)", path))
	}
	return m
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
