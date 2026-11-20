package main

import (
	"encoding/json"
	"os"
	"slices"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
)

func rankedOf(name string, id int32, dps, stderr float64) rankedGear {
	return rankedGear{option: gearOption{spec: &proto.ItemSpec{Id: id}, name: name}, result: result{dps: dps, stderr: stderr}}
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

// Two challengers that change the same slot can't both go in: the one with the bigger gain
// (first) stays. A one hand setup changes both hands, so a later off hand change clashes too.
func TestDisjointChanges(t *testing.T) {
	item := func(id int32) *proto.ItemSpec { return &proto.ItemSpec{Id: id} }
	best := []*proto.ItemSpec{item(1), item(2), item(3), nil}
	with := func(slot int, spec *proto.ItemSpec) []*proto.ItemSpec {
		out := slices.Clone(best)
		out[slot] = spec
		return out
	}
	wrist := with(1, item(20))
	wristAgain := with(1, item(21))
	ring := with(2, item(30))
	enchanted := with(0, &proto.ItemSpec{Id: 1, Enchant: 5})
	hands := with(3, item(40))
	hands[2] = item(31)
	got := disjointChanges(best, [][]*proto.ItemSpec{wrist, wristAgain, ring, enchanted, hands})
	if want := []int{0, 2, 3}; !slices.Equal(got, want) {
		t.Errorf("disjointChanges = %v, want %v", got, want)
	}
	if got := gearChanges(best, best); len(got) != 0 {
		t.Errorf("gearChanges of the set itself = %v, want none", got)
	}
	if got := gearChanges(best, hands); !slices.Equal(got, []int{2, 3}) {
		t.Errorf("gearChanges = %v, want [2 3]", got)
	}
}

// A conditional rank puts the item among the list's other entries by its loss with the condition
// met, so a rank no higher than the number of picks means BiS under the condition.
func TestBisRank(t *testing.T) {
	// Two trinket picks, then entries losing 1.3, 3.5 (the item itself) and 4.5.
	entries := []bisEntry{{Loss: 0, BiS: true}, {Loss: 0, BiS: true}, {Loss: 1.3}, {Loss: 3.5}, {Loss: 4.5}}
	cases := []struct {
		self int
		loss float64
		want int
	}{
		{3, -0.4, 1}, // Beats a pick: BiS.
		{3, 0, 3},    // Level with the picks: right after them.
		{3, 1.0, 3},  // Between the picks and the 1.3.
		{3, 1.3, 4},  // A tie goes after the entry it ties.
		{3, 4.0, 4},  // Its own 3.5 doesn't count.
		{-1, 4.0, 5}, // An unlisted item counts every entry.
		{-1, 9.0, 6},
	}
	for _, c := range cases {
		if got := bisRank(entries, c.self, c.loss); got != c.want {
			t.Errorf("bisRank(self %d, loss %g) = %d, want %d", c.self, c.loss, got, c.want)
		}
	}
}

// With -gear-rank-slots the ranked lists come from the new ranking and every other list from the
// old file, in the file's order. A ring list covers both ring slots' enchant lists.
func TestMergeLists(t *testing.T) {
	defer func(old string) { *gearRankSlots = old }(*gearRankSlots)
	*gearRankSlots = "trinket,finger"
	old := map[string][]json.RawMessage{
		"head":    {json.RawMessage(`{"id": 1}`)},
		"finger":  {json.RawMessage(`{"id":2}`)},
		"trinket": {json.RawMessage(`{"id":3}`)},
		"offhand": {},
	}
	fresh := []rawList{{name: "trinket", entries: []json.RawMessage{json.RawMessage(`{"id":4}`)}}}
	var got []string
	for _, l := range mergeLists(bisListNames, fresh, old) {
		got = append(got, l.name)
		for _, e := range l.entries {
			got = append(got, string(e))
		}
	}
	// The finger list was ranked and came out empty, so the old one doesn't come back.
	want := []string{"head", `{"id":1}`, "trinket", `{"id":4}`, "offhand"}
	if !slices.Equal(got, want) {
		t.Errorf("mergeLists = %q, want %q", got, want)
	}
	if !rankListSelected("finger2") || rankListSelected("head") {
		t.Errorf("rankListSelected: finger2 %t, head %t", rankListSelected("finger2"), rankListSelected("head"))
	}
}

// The effects file is BayovCore's contract (docs/bis_lists.md, Effect values): valid JSON, the
// items keyed by ID, and cooldown and conditional only where the item has them.
func TestWriteEffectsFile(t *testing.T) {
	path := t.TempDir() + "/effects.json"
	zero := 0.0
	items := map[int32]effectsItem{
		11815: {Name: "Hand of Justice", DPS: 8.2},
		10455: {Name: "Chained Essence of Eranikus", DPS: 29.5, Cooldown: &zero,
			Conditional: []effectsConditional{{Label: "vs Trolls", DPS: 59.3}}},
	}
	if err := writeEffectsFile(path, 180, 90, 1120.184, 10000, items); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Dur    float64                    `json:"dur"`
		UseDur float64                    `json:"use_dur"`
		DPS    float64                    `json:"dps"`
		Items  map[string]json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("%v:\n%s", err, data)
	}
	if got.Dur != 180 || got.UseDur != 90 || got.DPS != 1120.18 || len(got.Items) != 2 {
		t.Errorf("dur %v, use_dur %v, dps %v, %d items, want 180, 90, 1120.18 and 2", got.Dur, got.UseDur, got.DPS, len(got.Items))
	}
	for id, want := range map[string]string{
		"11815": `{"name":"Hand of Justice","dps":8.2}`,
		"10455": `{"name":"Chained Essence of Eranikus","dps":29.5,"cooldown":0,"conditional":[{"label":"vs Trolls","dps":59.3}]}`,
	} {
		if string(got.Items[id]) != want {
			t.Errorf("item %s: %s, want %s", id, got.Items[id], want)
		}
	}
}
