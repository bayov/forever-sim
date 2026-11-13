package main

import (
	"slices"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
)

func rankedOf(name string, id int32, dps, stderr float64) rankedGear {
	return rankedGear{gearOption{spec: &proto.ItemSpec{Id: id}, name: name}, result{dps: dps, stderr: stderr}}
}

func namesOf(options []gearOption) []string {
	var out []string
	for _, o := range options {
		out = append(out, o.name)
	}
	return out
}

func TestRankStages(t *testing.T) {
	for _, c := range []struct {
		screen, full int32
		want         []int32
	}{
		{100, 2000, []int32{100, 447, 2000}},
		{0, 2000, []int32{2000}},
		{2000, 2000, []int32{2000}},
		{1000, 2000, []int32{1000, 2000}},
	} {
		if got := rankStages(c.screen, c.full); !slices.Equal(got, c.want) {
			t.Errorf("rankStages(%d, %d) = %v, want %v", c.screen, c.full, got, c.want)
		}
	}
}

// The race keeps an item a noisy screen puts behind the keep-th best or the empty slot, as
// long as it is within the noise. Blackhand's Breadth was lost to a plain top N cut at 100
// iterations in the Phase 0 list of 2026-11-13.
func TestRaceCut(t *testing.T) {
	// Best first, as scoreOptions returns them. Each screen score is 3 DPS off.
	ranked := []rankedGear{
		rankedOf("Hand of Justice", 1, 975, 3),
		rankedOf("Behemoth Bones", 2, 972, 3),
		rankedOf("Junk", 3, 968, 3),
		rankedOf("Blackhand's Breadth", 4, 960, 3), // 1 behind the empty slot, within the noise
		rankedOf("Stamina Trinket", 5, 961, 3),     // exactly the empty slot: does nothing
		rankedOf("Clearly Worse", 6, 940, 3),       // 21 behind the empty slot
	}
	floor := result{dps: 961, stderr: 3}
	got := namesOf(raceCut(ranked, &floor, 25, 2))
	want := []string{"Hand of Justice", "Behemoth Bones", "Junk", "Blackhand's Breadth"}
	if !slices.Equal(got, want) {
		t.Errorf("raceCut with the empty slot = %v, want %v", got, want)
	}

	// With keep 2, the cut is behind Behemoth Bones (972) by more than 2 * 4.2 DPS.
	got = namesOf(raceCut(ranked, nil, 2, 2))
	want = []string{"Hand of Justice", "Behemoth Bones", "Junk"}
	if !slices.Equal(got, want) {
		t.Errorf("raceCut keeping 2 = %v, want %v", got, want)
	}
}

// Rolls of one item and suffix name count once, the way trimBis counts them.
func TestTopByName(t *testing.T) {
	ranked := []rankedGear{
		rankedOf("Necklace of the Tiger", 1, 970, 1),
		rankedOf("Necklace of the Tiger", 1, 969, 1), // a weaker roll
		rankedOf("Necklace of the Gorilla", 1, 968, 1),
		rankedOf("Locket", 2, 967, 1),
	}
	got := topByName(ranked, 2)
	if len(got) != 3 || got[2].option.name != "Necklace of the Gorilla" {
		t.Errorf("topByName keeping 2 = %d entries, want the two Tiger rolls and the Gorilla", len(got))
	}
	if got := topByName(ranked, 5); len(got) != 4 {
		t.Errorf("topByName keeping 5 = %d entries, want all 4", len(got))
	}
}

// The keep-th best counts names, so weaker rolls between them don't move the cut.
func TestRaceCutCountsNames(t *testing.T) {
	ranked := []rankedGear{
		rankedOf("Necklace of the Tiger", 1, 980, 1),
		rankedOf("Necklace of the Tiger", 1, 979, 1),
		rankedOf("Locket", 2, 978, 1),
		rankedOf("Locket", 2, 977, 1),
		rankedOf("Far Behind", 3, 960, 1),
	}
	got := namesOf(raceCut(ranked, nil, 2, 2))
	want := []string{"Necklace of the Tiger", "Necklace of the Tiger", "Locket", "Locket"}
	if !slices.Equal(got, want) {
		t.Errorf("raceCut keeping 2 names = %v, want %v", got, want)
	}
}

func TestBestRolls(t *testing.T) {
	roll := func(name string, id, suffix int32, stats ...float64) gearOption {
		return gearOption{spec: &proto.ItemSpec{Id: id, RandomSuffix: suffix}, name: name, suffixStats: stats}
	}
	options := []gearOption{
		roll("Necklace of the Tiger", 1, 10, 5, 5),
		roll("Necklace of the Tiger", 1, 11, 6, 6), // best Tiger roll
		roll("Necklace of the Tiger", 1, 12, 7, 4), // more of one stat, less of the other: stays
		roll("Necklace of the Tiger", 1, 13, 6, 6), // equal to 11: the first one stays
		roll("Necklace of the Gorilla", 1, 20, 3),  // another suffix name
		roll("Ring of the Tiger", 2, 10, 5, 5),     // another item
		{spec: &proto.ItemSpec{Id: 3}, name: "Locket"},
	}
	var got []int32
	for _, o := range bestRolls(options) {
		got = append(got, o.spec.Id*100+o.spec.RandomSuffix)
	}
	want := []int32{111, 112, 120, 210, 300}
	if !slices.Equal(got, want) {
		t.Errorf("bestRolls = %v, want %v", got, want)
	}
}

// With each iteration's DPS, two results pair: a gap that holds in every iteration is sure
// however much the iterations spread.
func TestGainNoisePaired(t *testing.T) {
	a := result{dps: 950, stderr: 20, values: []float64{900, 1000, 950, 950}}
	b := result{dps: 951, stderr: 20, values: []float64{901, 1001, 951, 951}}
	if noise := gainNoise(a, b); noise > 1e-9 {
		t.Errorf("paired noise of a steady 1 DPS gap = %g, want 0", noise)
	}
	b.values = nil
	if noise := gainNoise(a, b); noise < 28 {
		t.Errorf("unpaired noise = %g, want the two errors added up (28.3)", noise)
	}
}
