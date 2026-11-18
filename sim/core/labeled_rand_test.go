package core

import (
	"strconv"
	"testing"
)

// labeledRandOf finds a label's source through a cache keyed by where the label's text is. Each
// label must still get the stream the plain map gave it: one source per label text, seeded from
// the sim's seed and the label, and seeded again at its first roll after a reseed.
func TestLabeledRandMatchesMap(t *testing.T) {
	sim := &Simulation{
		rand:         NewSplitMix(12345),
		roundRand:    NewSplitMix(1),
		labeledRands: true,
		testRands:    make(map[string]*labeledRand),
		labelCache:   new([labelCacheSize]labelCacheEntry),
	}

	// The stream the labels had before the cache: a map from the label's text.
	type ref struct {
		rng     *SplitMix64
		reseeds int
	}
	refs := map[string]*ref{}
	refRoll := func(label string) float64 {
		r, ok := refs[label]
		if !ok {
			r = &ref{rng: NewSplitMix(uint64(makeTestRandSeed(sim.rand.GetSeed(), label)))}
			refs[label] = r
		} else if r.reseeds != sim.reseeds {
			r.rng.Seed(makeTestRandSeed(sim.rand.GetSeed(), label))
		}
		r.reseeds = sim.reseeds
		return r.rng.NextFloat64()
	}

	// More labels than the cache has slots, so they collide, and each also built again at
	// another address, so the same text comes from different pointers.
	labels := []string{"Damage Roll", "", "Crit", "Hit"}
	for i := 0; i < 3*labelCacheSize; i++ {
		labels = append(labels, "Label "+strconv.Itoa(i))
	}
	pick := NewSplitMix(777)
	for i := 0; i < 200_000; i++ {
		label := labels[int(pick.NextFloat64()*float64(len(labels)))]
		if pick.NextFloat64() < 0.3 {
			label = string([]byte(label)) // same text, new address
		}
		if pick.NextFloat64() < 0.001 {
			sim.reseedWithin("Test", i)
		}
		want := refRoll(label)
		var got float64
		if i%2 == 0 {
			got = sim.RandomFloat(label)
		} else {
			got = sim.labelRand(label).NextFloat64()
		}
		if got != want {
			t.Fatalf("roll %d of %q: got %v, want %v", i, label, got, want)
		}
	}
	if len(sim.testRands) != len(refs) {
		t.Fatalf("%d sources, want %d", len(sim.testRands), len(refs))
	}
}
