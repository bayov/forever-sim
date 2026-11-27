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

// A conditional rank is the item's place among the list's other entries by its loss with the
// condition met. It doesn't say BiS: a -gear-no-pick item can sit above the picks.
func TestBisRank(t *testing.T) {
	// A no pick item 1.85 ahead, two ring picks, then entries losing 1.3, 3.5 (the item itself)
	// and 4.5.
	entries := []bisEntry{{Loss: -1.85, noPick: true}, {Loss: 0, BiS: true}, {Loss: 0, BiS: true}, {Loss: 1.3}, {Loss: 3.5}, {Loss: 4.5}}
	cases := []struct {
		self int
		loss float64
		want int
	}{
		{4, -2.0, 1}, // Beats everything.
		{4, -0.4, 2}, // Beats the picks, but not the no pick item.
		{4, 0, 4},    // Level with the picks: right after them.
		{4, 1.0, 4},  // Between the picks and the 1.3.
		{4, 1.3, 5},  // A tie goes after the entry it ties.
		{4, 4.0, 5},  // Its own 3.5 doesn't count.
		{-1, 4.0, 6}, // An unlisted item counts every entry.
		{-1, 9.0, 7},
	}
	for _, c := range cases {
		if got := bisRank(entries, c.self, c.loss); got != c.want {
			t.Errorf("bisRank(self %d, loss %g) = %d, want %d", c.self, c.loss, got, c.want)
		}
	}
}

// An item is BiS under a condition or on cooldown when it would be the pick: a pick that loses
// nothing, or an item we may pick that beats the set. A no pick item never is.
func TestBisWins(t *testing.T) {
	cases := []struct {
		entry bisEntry
		loss  float64
		want  bool
	}{
		{bisEntry{BiS: true}, 0, true},
		{bisEntry{BiS: true}, 16.96, false}, // A pick with its use on cooldown.
		{bisEntry{}, -0.4, true},
		{bisEntry{}, 0, false},
		{bisEntry{noPick: true}, -10, false},
	}
	for _, c := range cases {
		if got := bisWins(c.entry, c.loss); got != c.want {
			t.Errorf("bisWins(%+v, %g) = %t, want %t", c.entry, c.loss, got, c.want)
		}
	}
}

// A trinket that adds nothing isn't ranked: Barov Peasant Caller on cooldown, Rune of the Dawn
// except vs Undead, and one that never adds anything isn't listed.
func TestMarkZero(t *testing.T) {
	zero := map[int32]bool{10455: true, 14022: true, 19812: true, 99: true}
	slots := markZero([]bisSlot{{name: "trinket", entries: []bisEntry{
		{ID: 10455, BiS: true, use: 18},
		{ID: 14022, Loss: 5.65, use: 7.6},
		{ID: 1, Loss: 8},
		{ID: 19812, Loss: 13.5, Conditional: []bisConditional{{Label: "vs Undead", Rank: 1}}},
		{ID: 99, Loss: 13.5},
	}}, {name: "neck", entries: []bisEntry{{ID: 99}}}}, zero)
	var got []int32
	for _, e := range slots[0].entries {
		got = append(got, e.ID)
		if e.Zero != (e.ID == 19812) {
			t.Errorf("item %d: zero %t", e.ID, e.Zero)
		}
	}
	if want := []int32{10455, 14022, 1, 19812}; !slices.Equal(got, want) || len(slots[1].entries) != 1 {
		t.Errorf("listed %v, want %v, and %d necks", got, want, len(slots[1].entries))
	}
	data, _ := json.Marshal(bisCooldown{Loss: 13.3})
	if string(data) != `{"loss":13.3}` {
		t.Errorf("a cooldown without a rank: %s", data)
	}
}

// A list is in order of loss, so a no pick item that beats the picks sits above them (user,
// 2026-11-20). The Phase 0 rings: Don Julio's Band and Warden's Signet need Exalted, so the picks
// are the next two. The picks stay even when the list is full, and a weaker or stronger roll of a
// pick's name goes.
func TestTrimBis(t *testing.T) {
	entries := []bisEntry{
		{ID: 1, Name: "Watcher's Signet of Brilliance", BiS: true},
		{ID: 2, Name: "Channeler's Ring", BiS: true},
		{ID: 3, Name: "Don Julio's Band", Loss: -1.85, noPick: true},
		{ID: 4, Name: "Warden's Signet of Brilliance", Loss: -0.64, noPick: true},
		{ID: 1, Name: "Watcher's Signet of Brilliance", Loss: -0.1},
		{ID: 5, Name: "Level Ring", Loss: 0},
		{ID: 6, Name: "Don Mauricio's Band of Domination", Loss: 1.28},
	}
	names := func(list []bisEntry) []string {
		var out []string
		for _, e := range list {
			out = append(out, e.Name)
		}
		return out
	}
	want := []string{"Don Julio's Band", "Warden's Signet of Brilliance", "Watcher's Signet of Brilliance", "Channeler's Ring", "Level Ring", "Don Mauricio's Band of Domination"}
	if got := names(trimBis(entries, 25)); !slices.Equal(got, want) {
		t.Errorf("trimBis = %v, want %v", got, want)
	}
	want = []string{"Don Julio's Band", "Watcher's Signet of Brilliance", "Channeler's Ring"}
	if got := names(trimBis(entries, 3)); !slices.Equal(got, want) {
		t.Errorf("trimBis keep 3 = %v, want %v", got, want)
	}
	if entries[0].Name != "Watcher's Signet of Brilliance" {
		t.Errorf("trimBis changed its input, now %s first", entries[0].Name)
	}
}

// The enchant lists are in order of loss too, with the pick kept.
func TestTrimEnchants(t *testing.T) {
	entries := []bisEnchant{
		{Name: "Lesser Arcanum of Voracity", BiS: true},
		{Name: "Arcanum of Rapidity", Loss: -1.13},
		{Name: "Lesser Arcanum of Focus", Loss: 2},
	}
	var got []string
	for _, e := range trimEnchants(entries, 2) {
		got = append(got, e.Name)
	}
	if want := []string{"Arcanum of Rapidity", "Lesser Arcanum of Voracity"}; !slices.Equal(got, want) {
		t.Errorf("trimEnchants = %v, want %v", got, want)
	}
}

// With -gear-rank-slots the ranked lists come from the new ranking and every other list from the
// old file, in the file's order, with each copied entry's dps measured again on the set's DPS now
// (its loss stays). A ring list covers both ring slots' enchant lists.
func TestMergeLists(t *testing.T) {
	defer func(old string) { *gearRankSlots = old }(*gearRankSlots)
	*gearRankSlots = "trinket,finger"
	old := map[string][]json.RawMessage{
		"head": {json.RawMessage(`{"id": 1}`),
			json.RawMessage(`{"id":5, "name":"Old Helm", "dps": 966.23, "loss": 0, "bis": true, "cooldown": {"loss": 1.5}}`),
			json.RawMessage(`{"id":6,"dps":964.23,"loss":2,"note":"dps"}`)},
		"finger":  {json.RawMessage(`{"id":2}`)},
		"trinket": {json.RawMessage(`{"id":3}`)},
		"offhand": {},
	}
	fresh := []rawList{{name: "trinket", entries: []json.RawMessage{json.RawMessage(`{"id":4}`)}}}
	var got []string
	merged, err := mergeLists(bisListNames, fresh, old, 953.534)
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range merged {
		got = append(got, l.name)
		for _, e := range l.entries {
			got = append(got, string(e))
		}
	}
	// The finger list was ranked and came out empty, so the old one doesn't come back.
	want := []string{"head", `{"id":1}`,
		`{"id":5,"name":"Old Helm","dps":953.53,"loss":0,"bis":true,"cooldown":{"loss":1.5}}`,
		`{"id":6,"dps":951.53,"loss":2,"note":"dps"}`,
		"trinket", `{"id":4}`, "offhand"}
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
	if err := writeEffectsFile(path, 180, 90, 1120.184, 10000, []int32{11815, 10455}, items); err != nil {
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
		Set    []int32                    `json:"set"`
		Items  map[string]json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("%v:\n%s", err, data)
	}
	if got.Dur != 180 || got.UseDur != 90 || got.DPS != 1120.18 || len(got.Items) != 2 {
		t.Errorf("dur %v, use_dur %v, dps %v, %d items, want 180, 90, 1120.18 and 2", got.Dur, got.UseDur, got.DPS, len(got.Items))
	}
	if !slices.Equal(got.Set, []int32{11815, 10455}) {
		t.Errorf("set %v, want [11815 10455]", got.Set)
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

// An effects file values each item on the set it was measured on, so it goes stale when a
// re-rank moves the spec's gear set and nobody reruns EFFECTS_ONLY (the rogue files after the
// second Phase 0 re-rank, 2026-11-27). The file records that set's item IDs ("set"), and they
// must be the gear set's. A file written before "set" existed isn't checked.
func TestEffectsFilesMatchGearSets(t *testing.T) {
	data, err := os.ReadFile("phase0_specs.json")
	if err != nil {
		t.Fatal(err)
	}
	var table struct {
		Specs map[string]struct{ Dir, GearOut, EffectsOut, Bench string }
	}
	if err := json.Unmarshal(data, &table); err != nil {
		t.Fatal(err)
	}
	checked := 0
	for name, s := range table.Specs {
		if s.Bench != "" || s.GearOut == "" {
			continue
		}
		path := s.EffectsOut
		if path == "" {
			path = "ui/" + s.Dir + "/bis/level60_phase0_" + name + ".effects.json"
		}
		data, err := os.ReadFile("../../" + path)
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			t.Fatal(err)
		}
		var effects struct {
			Set []int32 `json:"set"`
		}
		if err := json.Unmarshal(data, &effects); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if effects.Set == nil {
			continue
		}
		data, err = os.ReadFile("../../" + s.GearOut)
		if err != nil {
			t.Fatal(err)
		}
		var gear struct {
			Items []struct {
				Id int32 `json:"id"`
			} `json:"items"`
		}
		if err := json.Unmarshal(data, &gear); err != nil {
			t.Fatalf("%s: %v", s.GearOut, err)
		}
		var ids []int32
		for _, item := range gear.Items {
			if item.Id != 0 {
				ids = append(ids, item.Id)
			}
		}
		set := slices.Clone(effects.Set)
		slices.Sort(set)
		slices.Sort(ids)
		if !slices.Equal(set, ids) {
			t.Errorf("%s was measured on %v, but %s is %v: rerun EFFECTS_ONLY=1 SPEC=%s tools/rotopt/bis_phase0.sh",
				path, effects.Set, s.GearOut, ids, name)
		}
		checked++
	}
	if checked == 0 {
		t.Error("no effects file records its set")
	}
}

// A -gear-no-pick item stays in the ranking but never challenges the pick: the best item we
// may pick does. The Unstoppable Force needs Exalted with Frostwolf, so the Phase 0 list
// ranks it but picks the best two hander behind it.
func TestNoPick(t *testing.T) {
	pool := [][]gearOption{{
		{spec: &proto.ItemSpec{Id: 19323}, name: "The Unstoppable Force"},
		{spec: &proto.ItemSpec{Id: 1}, name: "Feasible Axe"},
	}}
	markNoPick(pool, []int32{19323})
	if !pool[0][0].noPick || pool[0][1].noPick {
		t.Fatalf("markNoPick = %v, %v, want true, false", pool[0][0].noPick, pool[0][1].noPick)
	}
	if got := namesOf(pickable(pool[0])); !slices.Equal(got, []string{"Feasible Axe"}) {
		t.Errorf("pickable = %v, want [Feasible Axe]", got)
	}
	if len(pool[0]) != 2 {
		t.Errorf("pickable changed the pool, now %d options", len(pool[0]))
	}
	ranked := []rankedGear{{option: pool[0][0], result: result{dps: 1010}}, {option: pool[0][1], result: result{dps: 1000}}}
	if top := firstPickable(ranked); top == nil || top.option.name != "Feasible Axe" {
		t.Errorf("firstPickable = %v, want Feasible Axe", top)
	}
	if top := firstPickable(ranked[:1]); top != nil {
		t.Errorf("firstPickable with only a no pick item = %v, want nil", top.option.name)
	}
}
