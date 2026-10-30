package compsearch

import (
	"math"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"google.golang.org/protobuf/encoding/protojson"
)

// subPool is the pool cut down to some base builds and their variants.
func subPool(t *testing.T, ids ...string) *Pool {
	t.Helper()
	full := loadTestPool(t)
	keep := map[string]bool{}
	for _, id := range ids {
		keep[id] = true
	}
	p := &Pool{Root: full.Root, Builds: map[string]*Build{}}
	for _, b := range full.Builds {
		if keep[b.Base.ID] {
			p.Builds[b.ID] = b
		}
	}
	for _, b := range full.Bases {
		if keep[b.ID] {
			p.Bases = append(p.Bases, b)
		}
	}
	return p
}

// Check 5 (docs/raid_sim.md, Checking it): on a small pool the search finds the best comp that trying
// every one finds, scored by the same exact single-player sims.
func TestSearchMatchesBruteForce(t *testing.T) {
	if testing.Short() {
		t.Skip("about a minute")
	}
	ids := []string{"mage__fire", "hunter__beast-mastery", "warlock__demonology", "druid__balance", "warrior__fury", "shaman__enhancement-level60"}
	p := subPool(t, ids...)
	sim, err := NewSimmer(p, "")
	if err != nil {
		t.Fatal(err)
	}
	// 5 players: any tank and 4 DPS. Physical builds lean on a warrior tank's Sunder and Battle Shout,
	// the casters on the warlock's curse and the druid: with and without the warrior, Sunder and the minor
	// armor slot don't add up.
	req := &proto.CompSearchRequest{Size: 5, Tanks: 1, Healers: 0,
		Factions: []proto.Faction{proto.Faction_Horde}, CampBuffs: []bool{false}, Iterations: 100, ContenderIterations: 100,
		Seed: 5, SkipSensitivity: true, Encounter: &proto.CompEncounter{Duration: 120, DurationVariation: 0, NaturalFlaskZone: true}}
	res, err := Search(p, sim, req, nil)
	if err != nil {
		t.Fatal(err)
	}
	got := res.Scenarios[0].Comps[0].Dps

	// Every multiset of 4 from the builds and their duty variants.
	var items []*Build
	for _, b := range p.Bases {
		items = append(items, b)
		items = append(items, b.Pets...)
		for _, d := range b.Duties {
			items = append(items, d)
		}
	}
	sc := &Scenario{Pool: p, Sim: sim, Faction: proto.Faction_Horde, Iterations: 100, Seed: 5,
		Fight: Fight{Duration: 120, FlaskZone: true}}
	var builds []*Build
	builds = append(builds, items...)
	model, err := sc.BuildModel(builds)
	if err != nil {
		t.Fatal(err)
	}
	ev := newExactValuer(sc, 100, model)
	best := math.Inf(-1)
	var bestComp *Comp
	var tank []*Preset
	var rec func(start int, cur []*Build)
	rec = func(start int, cur []*Build) {
		if len(cur) == 4 {
			r := NewRaid(5, tank, append([]*Build{}, cur...))
			// Every choice the placeholder has: a paladin's judgement, a shaman's air totem.
			for _, crus := range []int{0, 1} {
				for _, air := range [][]proto.RaidTotem{nil, {proto.RaidTotem_RaidTotemGraceOfAir}, {proto.RaidTotem_RaidTotemWindfury}} {
					if (crus == 1 && r.phPaladins == 0) || (air == nil) != (len(r.phShamans) == 0) {
						continue
					}
					plan := Plan{PHCrusader: crus, PHAir: air, Paladins: r.phPaladins}
					for _, b := range cur {
						if b.Class == proto.Class_ClassPaladin {
							plan.Paladins++
						}
					}
					plan.Paladins = min(3, plan.Paladins)
					plan.Up = r.phUp(plan) &^ GiftBlessing
					for _, b := range cur {
						plan.Up |= b.Gives &^ GiftBlessing
					}
					if plan.Paladins > 0 {
						plan.Up |= GiftBlessing
					}
					f := &Fillers{Fill: cur, Providers: cur}
					c, err := sc.EvalExact(r, plan, f, ev)
					if err != nil {
						t.Fatal(err)
					}
					if c != nil && c.Score > best {
						best, bestComp = c.Score, c
					}
				}
			}
			return
		}
		for i := start; i < len(items); i++ {
			rec(i, append(cur, items[i]))
		}
	}
	for _, t := range PresetsFor(proto.RaidRole_RaidRoleTank) {
		tank = []*Preset{t}
		rec(0, nil)
	}
	if bestComp == nil {
		t.Fatal("brute force found nothing")
	}
	if got < best-0.01 {
		t.Errorf("the search found %.2f; trying every comp found %.2f (%s)", got, best, bestComp.Describe())
	}
	if got > best+0.01 {
		t.Errorf("the search found %.2f, above every comp tried (%.2f): the two score differently", got, best)
	}
}

// The search gives the same result whatever the core count, run after run.
func TestSearchDeterministic(t *testing.T) {
	if testing.Short() {
		t.Skip("about a minute")
	}
	p := subPool(t, "mage__fire", "hunter__beast-mastery", "warlock__demonology", "warrior__fury")
	// The core count sizes both the sim workers and the ranking's goroutines (GOMAXPROCS).
	run := func(workers int) string {
		defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(workers))
		sim, err := NewSimmer(p, "")
		if err != nil {
			t.Fatal(err)
		}
		sim.Workers = workers
		req := &proto.CompSearchRequest{Size: 5, Tanks: 1, Healers: 0, Factions: []proto.Faction{proto.Faction_Alliance},
			CampBuffs: []bool{true}, Iterations: 60, ContenderIterations: 60, Seed: 9, SkipSensitivity: true,
			Encounter: &proto.CompEncounter{Duration: 90, DurationVariation: 0, NaturalFlaskZone: true}}
		res, err := Search(p, sim, req, nil)
		if err != nil {
			t.Fatal(err)
		}
		res.Seconds = 0
		for _, sc := range res.Scenarios {
			sc.Timings = nil
		}
		data, _ := protojson.Marshal(res)
		return string(data)
	}
	a, b, c := run(1), run(4), run(4)
	if a != b || b != c {
		for i := 0; i < len(a) && i < len(b); i++ {
			if a[i] != b[i] || (i < len(c) && b[i] != c[i]) {
				lo := max(0, i-300)
				t.Errorf("the result changed between runs, at:\n1: %s\n4: %s\n4: %s", a[lo:min(len(a), i+200)], b[lo:min(len(b), i+200)], c[lo:min(len(c), i+200)])
				break
			}
		}
	}
}

// The cache's drop table (drops.json): every dropped field leaves the build bit-identical, all of them
// together, at another seed than the table was made at, at each length and in each faction the table
// covers (dropVariants: 120, 180, 300 and 420 s).
func TestDropsAreUnread(t *testing.T) {
	if testing.Short() {
		t.Skip("about a minute")
	}
	p := loadTestPool(t)
	for _, b := range p.Builds {
		for _, camps := range []bool{false, true} {
			drops := Drops[DropKey(b.ID, camps)]
			if len(drops) == 0 {
				continue
			}
			ok, err := unread(p, b, camps, drops, 30, 50_000)
			if err != nil {
				t.Fatal(err)
			}
			if !ok {
				t.Errorf("%s (camps %v): dropping %v changes its result", b.ID, camps, drops)
			}
		}
	}
}

// The resolver's writing of a seat into a request (docs/raid_sim.md, Writing a role into a request).
func TestSeatWrite(t *testing.T) {
	p := loadTestPool(t)
	enh := p.Builds["shaman__enhancement-level60"]
	fury := p.Builds["warrior__fury"]
	lock := p.Builds["warlock__demonology-duty-curse"]
	seat := Seat{Up: GiftWF | GiftFT | GiftSoE | GiftCoE | GiftBattleShout, Blessings: []proto.Blessing{proto.Blessing_BlessingKings},
		Package: proto.RaidPackage_RaidPackageRestoration, PullSunders: 2, Camps: true}

	write := func(b *Build) *proto.RaidSimRequest {
		req := cloneReq(b.Req)
		s := seat
		s.Own = b.Casts
		s.Write(req, b)
		return req
	}
	// One weapon totem per receiver: Flametongue for Enhancement, Windfury for the others.
	if got := write(enh).Raid.Buffs.TotemWeaponBuff; got != proto.TotemWeaponBuff_TotemWeaponBuffFlametongue {
		t.Errorf("Enhancement's weapon totem: %v", got)
	}
	if got := write(fury).Raid.Buffs.TotemWeaponBuff; got != proto.TotemWeaponBuff_TotemWeaponBuffWindfury {
		t.Errorf("a warrior's weapon totem: %v", got)
	}
	// What a member casts itself is left out of its own request.
	if write(enh).Raid.Buffs.StrengthOfEarthTotem != proto.TristateEffect_TristateEffectMissing {
		t.Errorf("Enhancement drops its own Strength of Earth, so it isn't given")
	}
	if write(lock).Raid.Debuffs.CurseOfElements {
		t.Errorf("the curse warlock casts its own Curse of the Elements, so it isn't given")
	}
	if !write(fury).Raid.Debuffs.CurseOfElements {
		t.Errorf("the warrior gets the curse")
	}
	req := write(fury)
	if !req.Raid.Parties[0].Players[0].Buffs.BlessingOfKings || req.Raid.Parties[0].Players[0].Buffs.BlessingOfMight != 0 {
		t.Errorf("blessings: %v", req.Raid.Parties[0].Players[0].Buffs)
	}
	if req.Raid.Buffs.ManaSpringTotem != proto.TristateEffect_TristateEffectImproved || req.Raid.Parties[0].Buffs.ManaTideTotems != 1 {
		t.Errorf("the Restoration package: %v, %v", req.Raid.Buffs.ManaSpringTotem, req.Raid.Parties[0].Buffs.ManaTideTotems)
	}
	if r := req.Raid.Debuffs.SunderRamp; r == nil || r.PullSunders != 2 || r.Tank || req.Raid.Debuffs.SunderArmor {
		t.Errorf("Sunder: %v", req.Raid.Debuffs)
	}
	if c := req.Raid.Buffs.CampBuffs; c == nil || !c.FishBowl || !c.FirstAidKit {
		t.Errorf("camps: %v", c)
	}
	// Nothing of the report's own full set leaks through.
	if req.Raid.Buffs.Thorns != 0 || req.Raid.Debuffs.JudgementOfLight || req.Raid.Buffs.Salvation || req.Raid.Debuffs.FaerieFire {
		t.Errorf("the report's set leaked through")
	}
}

// M1's done-when (docs/raid_sim.md): a second run makes zero sims, and changing the encounter or the
// length misses the cache. (A code or item database change is a new binary, so a new version.)
func TestCacheReuse(t *testing.T) {
	if testing.Short() {
		t.Skip("about 20 s")
	}
	p := subPool(t, "mage__fire", "warrior__fury")
	dir := t.TempDir()
	req := func() *proto.CompSearchRequest {
		return &proto.CompSearchRequest{Size: 5, Tanks: 1, Healers: 0, Factions: []proto.Faction{proto.Faction_Horde},
			CampBuffs: []bool{false}, Iterations: 40, ContenderIterations: 40, Seed: 2, SkipSensitivity: true,
			Encounter: &proto.CompEncounter{Duration: 60, DurationVariation: 0, NaturalFlaskZone: true}}
	}
	for i, want0 := range []bool{false, true} {
		sim, err := NewSimmer(p, dir)
		if err != nil {
			t.Fatal(err)
		}
		res, err := Search(p, sim, req(), nil)
		if err != nil {
			t.Fatal(err)
		}
		if want0 && res.SimsRun != 0 {
			t.Errorf("run %d made %d sims, want 0", i+1, res.SimsRun)
		}
		if !want0 && res.SimsRun == 0 {
			t.Errorf("the first run made no sims")
		}
	}
	sim, _ := NewSimmer(p, "")
	b := p.Builds["warrior__fury"]
	key := func(f Fight) string {
		k, err := sim.key(&Job{Build: b, Faction: proto.Faction_Horde, Seat: Seat{Up: GiftAB}, Fight: f, Iterations: 40, Seed: 2})
		if err != nil {
			t.Fatal(err)
		}
		return k
	}
	base := Fight{Duration: 60, FlaskZone: true}
	k0 := key(base)
	for _, f := range []Fight{{Duration: 65, FlaskZone: true}, {Duration: 60}, {Duration: 60, FlaskZone: true, MobType: proto.MobType_MobTypeDragonkin},
		{Duration: 60, DurationVariation: 5, FlaskZone: true}} {
		if key(f) == k0 {
			t.Errorf("%+v has the same key as %+v", f, base)
		}
	}
	// Arcane Brilliance is unread by a warrior, so it's the same sim with it or without.
	k1, _ := sim.key(&Job{Build: b, Faction: proto.Faction_Horde, Seat: Seat{}, Fight: base, Iterations: 40, Seed: 2})
	if k1 != k0 {
		t.Errorf("Arcane Brilliance moved a warrior's key")
	}
}

// A health fight's length settles where the winner kills the boss.
func TestHealthFight(t *testing.T) {
	if testing.Short() {
		t.Skip("about 20 s")
	}
	p := subPool(t, "mage__fire", "warrior__fury")
	sim, _ := NewSimmer(p, "")
	req := &proto.CompSearchRequest{Size: 5, Tanks: 1, Healers: 0, Factions: []proto.Faction{proto.Faction_Horde},
		CampBuffs: []bool{false}, Iterations: 40, ContenderIterations: 40, Seed: 2, SkipSensitivity: true,
		Encounter: &proto.CompEncounter{Duration: 60, UseHealth: true, BossHealth: 400000, NaturalFlaskZone: true}}
	res, err := Search(p, sim, req, nil)
	if err != nil {
		t.Fatal(err)
	}
	sc := res.Scenarios[0]
	if kill := 400000 / sc.Comps[0].Dps; math.Abs(kill-sc.FightSeconds) > 1.5 {
		t.Errorf("the fight is %.1f s, but the winner kills the boss in %.1f s", sc.FightSeconds, kill)
	}
}

// The native server's async search ends with the result, and Abort stops it.
func TestSearchAsync(t *testing.T) {
	if testing.Short() {
		t.Skip("about 20 s")
	}
	t.Setenv("FOREVER_SIM_ROOT", repoRoot())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	req := func() *proto.CompSearchRequest {
		return &proto.CompSearchRequest{Size: 5, Tanks: 1, Healers: 0, Factions: []proto.Faction{proto.Faction_Horde},
			CampBuffs: []bool{false}, Iterations: 30, ContenderIterations: 30, Seed: 4, SkipSensitivity: true,
			Encounter: &proto.CompEncounter{Duration: 60, NaturalFlaskZone: true}, MaxCombos: 3}
	}
	ch := make(chan *proto.ProgressMetrics, 100)
	SearchAsync(req(), ch, "test-async-1")
	for p := range ch {
		if p.FinalCompSearchResult != nil {
			if p.FinalCompSearchResult.Error != nil {
				t.Fatal(p.FinalCompSearchResult.Error.Message)
			}
			if len(p.FinalCompSearchResult.Scenarios) != 1 {
				t.Errorf("want one scenario")
			}
			break
		}
	}
	ch = make(chan *proto.ProgressMetrics, 100)
	SearchAsync(req(), ch, "test-async-2")
	simsignals.AbortById("test-async-2")
	for p := range ch {
		if p.FinalCompSearchResult != nil {
			if p.FinalCompSearchResult.Error == nil || p.FinalCompSearchResult.Error.Message != ErrAborted.Error() {
				t.Errorf("an aborted search should end with %q, got %v", ErrAborted, p.FinalCompSearchResult.Error)
			}
			break
		}
	}
}

// The confirmation raid-sims the comps it scored: each confirmed comp's raid sim lands near its score
// (members add up within about 0.1% in the doc's measurements; this allows 1% and 4 standard errors at
// these few iterations).
func TestConfirmMatchesScore(t *testing.T) {
	if testing.Short() {
		t.Skip("about a minute")
	}
	p := subPool(t, "mage__fire", "warlock__demonology", "hunter__beast-mastery", "warrior__fury", "priest__shadow")
	sim, _ := NewSimmer(p, "")
	req := &proto.CompSearchRequest{Size: 5, Tanks: 1, Healers: 0, FixedTanks: []proto.RaidPreset{proto.RaidPreset_RaidPresetProtectionWarrior},
		Factions: []proto.Faction{proto.Faction_Horde}, CampBuffs: []bool{false}, Iterations: 400, ContenderIterations: 400,
		Seed: 13, SkipSensitivity: true, Confirm: true, ConfirmIterations: 1500, ConfirmTopIterations: 1500,
		Encounter: &proto.CompEncounter{Duration: 120, NaturalFlaskZone: true}}
	res, err := Search(p, sim, req, nil)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, c := range res.Scenarios[0].Comps {
		if !c.Confirmed {
			continue
		}
		n++
		if d := c.RaidSimDps - c.Dps; math.Abs(d) > math.Max(4*math.Hypot(c.Se, c.RaidSimSe), 0.01*c.Dps) {
			t.Errorf("raid sim %.1f, score %.1f (%+.1f): %s", c.RaidSimDps, c.Dps, d, c.BuffSet)
		}
	}
	if n == 0 {
		t.Errorf("nothing was confirmed")
	}
}

// A code or item database change is a new binary, so a new version: its keys and its cache directory are
// new, and an old version's directory goes once unused for a week.
func TestCacheVersion(t *testing.T) {
	p := subPool(t, "warrior__fury")
	dir := t.TempDir()
	a, err := newSimmerAt(p, dir, strings.Repeat("a", 64))
	if err != nil {
		t.Fatal(err)
	}
	job := func() *Job {
		return &Job{Build: p.Builds["warrior__fury"], Faction: proto.Faction_Horde, Fight: Fight{Duration: 60}, Iterations: 10, Seed: 1}
	}
	ka, _ := a.key(job())
	b, err := newSimmerAt(p, dir, strings.Repeat("b", 64))
	if err != nil {
		t.Fatal(err)
	}
	kb, _ := b.key(job())
	if ka == kb || a.Dir == b.Dir {
		t.Errorf("two versions share a key or a directory")
	}
	// One in use stays (the CLI and the server can run different binaries); one unused for a week goes.
	if _, err := os.Stat(a.Dir); err != nil {
		t.Errorf("a version used just now should stay: %v", err)
	}
	old := time.Now().Add(-8 * 24 * time.Hour)
	if err := os.Chtimes(a.Dir, old, old); err != nil {
		t.Fatal(err)
	}
	if _, err := newSimmerAt(p, dir, strings.Repeat("b", 64)); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(a.Dir); !os.IsNotExist(err) {
		t.Errorf("a version unused for a week should be pruned")
	}
}

// The sensitivity matrix re-runs the search and says, for each scenario, whether the winner holds.
func TestMatrix(t *testing.T) {
	if testing.Short() {
		t.Skip("about a minute")
	}
	p := subPool(t, "mage__fire", "warrior__fury", "hunter__beast-mastery")
	sim, _ := NewSimmer(p, "")
	req := &proto.CompSearchRequest{Size: 5, Tanks: 1, Healers: 0, FixedTanks: []proto.RaidPreset{proto.RaidPreset_RaidPresetProtectionWarrior},
		Factions: []proto.Faction{proto.Faction_Horde}, CampBuffs: []bool{true}, Iterations: 30, ContenderIterations: 30, Seed: 3,
		SkipSensitivity: true, Encounter: &proto.CompEncounter{Duration: 90, NaturalFlaskZone: true}}
	res, err := Search(p, sim, req, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := RunMatrix(p, sim, req, res, nil); err != nil {
		t.Fatal(err)
	}
	if got, want := len(res.Scenarios[0].Sensitivity), len(Matrix(req)); got != want {
		t.Errorf("%d matrix rows, want %d", got, want)
	}
}

// fakeValuer values a build by its id alone.
type fakeValuer map[string]float64

func (v fakeValuer) V(b *Build, s Seat) (float64, float64, string) { return v[b.ID], 0, b.ID }

// A fixed DPS member can hold a duty in its own slot, rather than the fill adding a second member of its
// class just to hold it.
func TestFixedMemberHoldsDuty(t *testing.T) {
	p := subPool(t, "warrior__fury", "mage__fire")
	fury, mage := p.Builds["warrior__fury"], p.Builds["mage__fire"]
	sunder := fury.Duties[proto.Duty_DutySunderUpkeep]
	if sunder == nil {
		t.Fatal("warrior__fury has no Sunder duty variant")
	}
	sc := &Scenario{Pool: p, Faction: proto.Faction_Horde, Fight: Fight{Duration: 180}}
	r := NewRaid(10, []*Preset{PresetByID(proto.RaidPreset_RaidPresetFeralTank), PresetByID(proto.RaidPreset_RaidPresetDisciplinePriest)}, []*Build{fury})
	plan := Plan{Up: (r.phUp(Plan{}) | GiftSunderTank | fury.Gives | mage.Gives) &^ GiftBlessing}
	f := &Fillers{Fill: []*Build{fury, mage}, Providers: []*Build{sunder, mage}}
	c := sc.Fill(r, plan, f, fakeValuer{fury.ID: 1000, sunder.ID: 950, mage.ID: 1001})
	if c == nil {
		t.Fatalf("the plan can't be filled: plan %s, ph %s, fury %s, sunder %s, mage %s", plan.Up, r.phUp(Plan{}), fury.Gives, sunder.Gives, mage.Gives)
	}
	warriors := 0
	for _, m := range c.Members {
		if m.B.Class == proto.Class_ClassWarrior {
			warriors++
			if m.B != sunder || !m.Fixed {
				t.Errorf("the warrior is %s (fixed %v), not the fixed member on Sunder", m.B.ID, m.Fixed)
			}
		}
	}
	if warriors != 1 {
		t.Errorf("%d warriors, want the fixed one alone: %s", warriors, c.Describe())
	}
}

// A request the search can't run fails before any sims.
func TestSearchRejectsBadRequests(t *testing.T) {
	p := subPool(t, "mage__fire")
	sim, err := NewSimmer(p, "")
	if err != nil {
		t.Fatal(err)
	}
	bad := map[string]*proto.CompSearchRequest{
		"size 50":        {Size: 50},
		"unknown preset": {Size: 10, Tanks: 1, Healers: 1, FixedTanks: []proto.RaidPreset{999}},
		"healer as tank": {Size: 10, Tanks: 1, Healers: 1, FixedTanks: []proto.RaidPreset{proto.RaidPreset_RaidPresetHolyPriest}},
		"faction 7":      {Size: 10, Factions: []proto.Faction{7}},
		"negative fight": {Size: 10, Encounter: &proto.CompEncounter{Duration: -5}},
	}
	for name, req := range bad {
		if _, err := Search(p, sim, req, nil); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
	if n := sim.Run.Load(); n != 0 {
		t.Errorf("%d sims run for bad requests", n)
	}
}
