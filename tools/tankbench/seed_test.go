package main

import (
	"os"
	"testing"

	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/tools/tankscore"
	"google.golang.org/protobuf/encoding/protojson"
)

// The tank tools' -seed is spaced apart (tankscore.Seed), so two seeds run disjoint iterations and give
// clearly different results, while one seed reproduces its own (finding tank_rotations-20261104-13).
func TestSeedsDisjoint(t *testing.T) {
	sim.RegisterAll()
	data, err := os.ReadFile("req/warrior_protection_balanced.json")
	if err != nil {
		t.Fatal(err)
	}
	enc, err := tankscore.LoadEncounter("../encounters/onyxia.json")
	if err != nil {
		t.Fatal(err)
	}
	const iters = 300
	run := func(seed int64) []float64 {
		rsr := &proto.RaidSimRequest{}
		if err := protojson.Unmarshal(data, rsr); err != nil {
			t.Fatal(err)
		}
		rsr.SimOptions.Iterations = iters
		rsr.SimOptions.RandomSeed = seed
		tankscore.UseEncounter(rsr, enc)
		tankscore.MakeTank(rsr, enc.Healing.Model(), &proto.TankScoring{DpsStart: 2, ThreatPerDamage: 1})
		res := core.RunRaidSimConcurrent(rsr)
		if res.Error != nil {
			t.Fatal(res.Error.Message)
		}
		return res.RaidMetrics.Parties[0].Players[0].TankScore.SafeDpsEach
	}
	// How many of b's iterations also appear in a, by value.
	shared := func(a, b []float64) int {
		seen := map[float64]bool{}
		for _, v := range a {
			seen[v] = true
		}
		n := 0
		for _, v := range b {
			if seen[v] {
				n++
			}
		}
		return n
	}
	h1, h1again, h2 := run(tankscore.Seed(1)), run(tankscore.Seed(1)), run(tankscore.Seed(2))
	if tankscore.Seed(1) != 1 {
		t.Errorf("the default seed 1 maps to %d, want 1 so earlier numbers stand", tankscore.Seed(1))
	}
	if len(h1) != iters {
		t.Fatalf("%d per-iteration SafeDPS values, want %d", len(h1), iters)
	}
	for i := range h1 {
		if h1[i] != h1again[i] {
			t.Fatalf("seed 1 isn't reproducible: iteration %d %g vs %g", i, h1[i], h1again[i])
		}
	}
	if n := shared(h1, h2); n > iters/20 {
		t.Errorf("spaced seeds 1 and 2 share %d of %d iterations", n, iters)
	}
	// The plain seeds overlap in all but one iteration, which tankscore.Seed is there to avoid.
	if n := shared(run(1), run(2)); n < iters-5 {
		t.Errorf("plain seeds 1 and 2 share only %d of %d iterations; is the sim still seeding iteration i with seed+i?", n, iters)
	}
}
