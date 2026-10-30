package compsearch

import (
	"math"
	"testing"

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
	// 5 players: a Protection warrior tank and 4 DPS. Physical builds lean on its Sunder, the casters on
	// the warlock's curse and the druid: Sunder and the minor armor slot don't add up for them.
	req := &proto.CompSearchRequest{Size: 5, Tanks: 1, Healers: 0, FixedTanks: []proto.RaidPreset{proto.RaidPreset_RaidPresetProtectionWarrior},
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
	tank := []*Preset{PresetByID(proto.RaidPreset_RaidPresetProtectionWarrior)}
	best := math.Inf(-1)
	var bestComp *Comp
	var rec func(start int, cur []*Build)
	rec = func(start int, cur []*Build) {
		if len(cur) == 4 {
			r := NewRaid(5, tank, append([]*Build{}, cur...))
			plan := Plan{Up: r.phUp(Plan{})}
			for _, b := range cur {
				plan.Up |= b.Gives &^ GiftBlessing
			}
			f := &Fillers{Fill: cur, Providers: cur}
			c, err := sc.EvalExact(r, plan, f, ev)
			if err != nil {
				t.Fatal(err)
			}
			if c != nil && c.Score > best {
				best, bestComp = c.Score, c
			}
			return
		}
		for i := start; i < len(items); i++ {
			rec(i, append(cur, items[i]))
		}
	}
	rec(0, nil)
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
	run := func(workers int) string {
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
			ok, err := unread(p, b, camps, drops, 30, 3)
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
