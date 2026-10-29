package compsearch

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
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
		data, _ := protojson.Marshal(res)
		return string(data)
	}
	a, b, c := run(1), run(4), run(4)
	if a != b || b != c {
		t.Errorf("the result changed between runs")
	}
}

// The cache's drop table (drops.json): every dropped field leaves the build bit-identical, all of them
// together, at another seed and length than the table was made at.
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
