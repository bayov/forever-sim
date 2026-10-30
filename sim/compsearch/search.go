package compsearch

import (
	"fmt"
	googleProto "google.golang.org/protobuf/proto"
	"math"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wowsims/classic/sim/core/proto"
)

// Progress reports a search's phase and the best comp so far.
type Progress func(phase string, best *Comp, sc *Scenario)

// Defaults fills in a request's defaults.
func Defaults(req *proto.CompSearchRequest) {
	if req.Size == 0 {
		req.Size = 20
	}
	if req.Tanks == 0 && req.Healers == 0 {
		switch req.Size {
		case 10:
			req.Tanks, req.Healers = 1, 2
		case 20:
			req.Tanks, req.Healers = 2, 4
		case 40:
			req.Tanks, req.Healers = 3, 8
		}
	}
	if req.Encounter == nil {
		req.Encounter = &proto.CompEncounter{}
	}
	e := req.Encounter
	if e.Name == "" {
		e.Name = "Generic"
	}
	if e.Duration == 0 {
		e.Duration = 180
		if e.DurationVariation == 0 {
			e.DurationVariation = 20
		}
	}
	if len(req.Factions) == 0 {
		req.Factions = []proto.Faction{proto.Faction_Horde, proto.Faction_Alliance}
	}
	if len(req.CampBuffs) == 0 {
		req.CampBuffs = []bool{false, true}
	}
	if req.Iterations == 0 {
		req.Iterations = 1000
	}
	if req.ContenderIterations == 0 {
		req.ContenderIterations = 10000
	}
	if req.Seed == 0 {
		req.Seed = 101
	}
	if req.ConfirmIterations == 0 {
		req.ConfirmIterations = 5000
	}
	if req.ConfirmTopIterations == 0 {
		req.ConfirmTopIterations = 20000
	}
}

// validate rejects a request the search can't run, before any sims.
func validate(req *proto.CompSearchRequest) error {
	if req.Size < 1 || req.Size > 40 {
		return fmt.Errorf("a raid has 1 to 40 members, not %d", req.Size)
	}
	if req.Tanks < 0 || req.Healers < 0 {
		return fmt.Errorf("negative tank or healer slots")
	}
	for _, f := range req.Factions {
		if f != proto.Faction_Horde && f != proto.Faction_Alliance {
			return fmt.Errorf("unknown faction %v", f)
		}
	}
	e := req.Encounter
	if e.Duration <= 0 || e.DurationVariation < 0 || e.DurationVariation >= e.Duration {
		return fmt.Errorf("a fight of %.0f ± %.0f s", e.Duration, e.DurationVariation)
	}
	if e.UseHealth && e.BossHealth <= 0 {
		return fmt.Errorf("a health fight needs the boss's health")
	}
	for _, name := range req.RequiredEffects {
		if _, ok := GiftByName(name); !ok {
			return fmt.Errorf("unknown required effect %q (the names are %s)", name, strings.Join(giftNames, ", "))
		}
	}
	if req.Iterations <= 0 || req.ContenderIterations <= 0 || req.ConfirmIterations <= 0 || req.ConfirmTopIterations <= 0 {
		return fmt.Errorf("iterations must be positive")
	}
	for _, fx := range []struct {
		ids  []proto.RaidPreset
		role proto.RaidRole
		name string
	}{{req.FixedTanks, proto.RaidRole_RaidRoleTank, "tank"}, {req.FixedHealers, proto.RaidRole_RaidRoleHealer, "healer"}} {
		for _, id := range fx.ids {
			p := PresetByID(id)
			if p == nil {
				return fmt.Errorf("unknown preset %v", id)
			}
			if p.Role != fx.role {
				return fmt.Errorf("%s isn't a %s preset", p.Name, fx.name)
			}
		}
	}
	return nil
}

// Search runs a comp search (docs/raid_sim.md, Comp search): one scenario for each faction and camp
// setting.
func Search(pool *Pool, sim *Simmer, req *proto.CompSearchRequest, progress Progress) (*proto.CompSearchResult, error) {
	return SearchWithAbort(pool, sim, req, progress, nil)
}

// SearchWithAbort is Search that stops with ErrAborted once abort says so.
func SearchWithAbort(pool *Pool, sim *Simmer, req *proto.CompSearchRequest, progress Progress, abort func() bool) (*proto.CompSearchResult, error) {
	start := time.Now()
	Defaults(req)
	if err := validate(req); err != nil {
		return nil, err
	}
	if req.Tanks+req.Healers >= req.Size {
		return nil, fmt.Errorf("%d tanks and %d healers leave no DPS slot in a raid of %d", req.Tanks, req.Healers, req.Size)
	}
	if len(req.FixedTanks) > int(req.Tanks) || len(req.FixedHealers) > int(req.Healers) {
		return nil, fmt.Errorf("more fixed tank or healer presets than slots")
	}
	var fixed []*Build
	for _, id := range req.FixedDps {
		b, ok := pool.Builds[id]
		// A base build or a pet variant (a build of its own); a duty or encounter variant is the search's
		// to choose.
		if !ok || b.Duty != proto.Duty_DutyNone || b.MobType != proto.MobType_MobTypeUnknown {
			return nil, fmt.Errorf("fixed DPS member %q is not a base build or pet variant of the report", id)
		}
		fixed = append(fixed, b)
	}
	if len(fixed) > int(req.Size-req.Tanks-req.Healers) {
		return nil, fmt.Errorf("more fixed DPS members than DPS slots")
	}
	counter := &Counter{}
	res := &proto.CompSearchResult{Request: req, Version: sim.version[:16]}
	res.LeftOut = []string{
		"a DPS shaman's totem choice: Enhancement and Elemental drop their builds' fixed totems (the engine's assigned totems aren't in their rotations yet)",
		"Dark Inspiration as a party package (a Shadow priest's own Spirit is in its sims)",
		"the hunter pet variants' cat swing speed and Cobra Reflexes on the Tallstrider and wolf (an assumption)",
	}
	res.NotModelled = []string{"healing", "tanking", "threat", "pet deaths", "movement", "boss mechanics",
		"the tanks' and healers' own damage"}
	res.Assumptions = assumptionLines(req.Assumptions)
	for _, faction := range req.Factions {
		for _, camps := range req.CampBuffs {
			sc := &Scenario{Pool: pool, Sim: sim, Faction: faction, Camps: camps, Iterations: req.Iterations,
				Seed: req.Seed, Assumptions: req.Assumptions, Abort: abort, Counter: counter,
				Fight: Fight{Duration: req.Encounter.Duration, DurationVariation: req.Encounter.DurationVariation,
					MobType: req.Encounter.MobType, FlaskZone: req.Encounter.NaturalFlaskZone,
					SetRaidHits: req.Encounter.SetRaidDamage, RaidHits: req.Encounter.RaidDamageHitsPerMinute, NoFlask: req.Encounter.NoFlask}}
			var out *proto.CompScenarioResult
			var err error
			if !req.Encounter.UseHealth {
				out, err = sc.run(req, fixed, progress)
			} else {
				out, err = sc.runHealth(req, fixed, progress)
			}
			if err != nil {
				return nil, err
			}
			res.Scenarios = append(res.Scenarios, out)
		}
	}
	res.SimsRun = counter.Run.Load()
	res.CacheHits = counter.Hits.Load()
	res.Seconds = time.Since(start).Seconds()
	return res, nil
}

// runHealth is a health fight (docs/raid_sim.md, Fight length): the length is the boss's health over the
// raid's DPS, and the raid's DPS depends on the length. Each pass holds one length for every set, and the
// pass runs again while the winner's length moves by more than a second.
func (sc *Scenario) runHealth(req *proto.CompSearchRequest, fixed []*Build, progress Progress) (*proto.CompScenarioResult, error) {
	health := req.Encounter.BossHealth
	if health <= 0 {
		return nil, fmt.Errorf("a health fight needs the boss's health")
	}
	L := req.Encounter.Duration
	sc.BossHealth = health
	var out *proto.CompScenarioResult
	var passNotes []string
	// The passes find the length without the confirmation and the extras; the last pass, at the length
	// they settle on, runs everything.
	lean := googleProto.Clone(req).(*proto.CompSearchRequest)
	lean.Confirm, lean.SkipSensitivity = false, true
	settled := false
	for pass := 0; pass < 5; pass++ {
		sc.HealthLength = L
		// The model ranks at the nearest grid point; the exact values interpolate.
		sc.Fight.Duration = math.Max(lengthGrid, math.Round(L/lengthGrid)*lengthGrid)
		sc.Fight.DurationVariation = 0
		last := settled || pass == 4
		sc.lean = !last
		var err error
		if last {
			out, err = sc.run(req, fixed, progress)
		} else {
			out, err = sc.run(lean, fixed, progress)
		}
		if err != nil {
			return nil, err
		}
		if last {
			out.FightSeconds = L
			if !settled {
				passNotes = append(passNotes, fmt.Sprintf("the length hadn't settled after 4 passes; the result is at %.1f s", L))
			}
			break
		}
		// The top DPS (the comp shown first may be a simpler one tied with it).
		dps := 0.0
		for _, c := range out.Comps {
			if out.Comps[0].Confirmed {
				if c.Confirmed {
					dps = math.Max(dps, c.RaidSimDps)
				}
			} else {
				dps = math.Max(dps, c.Dps)
			}
		}
		next := health / dps
		out.FightSeconds = L
		passNotes = append(passNotes, fmt.Sprintf("health fight pass %d: %.1f s, the winner kills the boss in %.1f s", pass+1, L, next))
		if math.Abs(next-L) <= 1 {
			settled = true
			continue
		}
		L = next
	}
	out.Notes = append(passNotes, out.Notes...)
	return out, nil
}

func assumptionLines(a *proto.RaidAssumptions) []string {
	out := []string{
		"Each build at its best race in the faction, from the report's race tables.",
		"Mana Spring from two shamans doesn't stack, and a party has one package (shaman_audit 8.45, unverified).",
		"Furious Howl and the minor armor slot: the strongest counts, nothing stacks (until the launch checks).",
		"Hunter's Mark: each hunter marks for itself; the shared mark is left to the raid sims.",
		"Power Infusion fires as soon as it's ready; Innervate below a mana threshold.",
		"Everyone is in range of every raid-wide buff.",
	}
	if a != nil && a.CampChairSpellCritOnly {
		out = append(out, "Moonkin Aura takes only the Camp Chair's spell crit.")
	} else {
		out = append(out, "Moonkin Aura takes the whole Camp Chair (unverified).")
	}
	if a != nil && a.CampBuffsStackWithGroupBuffs {
		out = append(out, "Camp buffs stack with the group versions of their class buffs.")
	}
	return out
}

// Signature is what a combo's placeholders can give (docs/raid_sim.md, Tank and healer combos).
type Signature struct {
	WarriorTank, FeralTank, Druid, Priest bool
	Paladins, Shamans                     int
}

func (r *Raid) Signature() Signature {
	s := Signature{Paladins: min(3, r.phPaladins), Shamans: min(2, len(r.phShamans)), Priest: r.phGifts.Has(GiftPoS)}
	for _, p := range r.Combo {
		switch p.ID {
		case proto.RaidPreset_RaidPresetProtectionWarrior:
			s.WarriorTank = true
		case proto.RaidPreset_RaidPresetFeralTank:
			s.FeralTank = true
		}
		if p.Class == proto.Class_ClassDruid {
			s.Druid = true
		}
	}
	return s
}

func (s Signature) String() string {
	var parts []string
	if s.WarriorTank {
		parts = append(parts, "a warrior tank")
	}
	if s.FeralTank {
		parts = append(parts, "a Feral tank")
	} else if s.Druid {
		parts = append(parts, "a druid")
	}
	if s.Paladins > 0 {
		parts = append(parts, fmt.Sprintf("%d paladin%s", s.Paladins, map[bool]string{true: "s"}[s.Paladins > 1]))
	}
	if s.Shamans > 0 {
		parts = append(parts, fmt.Sprintf("%d shaman%s", s.Shamans, map[bool]string{true: "s"}[s.Shamans > 1]))
	}
	if s.Priest {
		parts = append(parts, "a priest")
	}
	if len(parts) == 0 {
		return "nothing the search uses"
	}
	return strings.Join(parts, ", ")
}

// Combos is every tank and healer combo that fits the slots, as multisets.
func Combos(req *proto.CompSearchRequest) [][]*Preset {
	tanks := multisets(PresetsFor(proto.RaidRole_RaidRoleTank), int(req.Tanks)-len(req.FixedTanks))
	healers := multisets(PresetsFor(proto.RaidRole_RaidRoleHealer), int(req.Healers)-len(req.FixedHealers))
	var fixed []*Preset
	for _, id := range append(append([]proto.RaidPreset{}, req.FixedTanks...), req.FixedHealers...) {
		fixed = append(fixed, PresetByID(id))
	}
	var out [][]*Preset
	for _, t := range tanks {
		for _, h := range healers {
			c := append(append(append([]*Preset{}, fixed...), t...), h...)
			out = append(out, c)
		}
	}
	return out
}

func multisets(from []*Preset, n int) [][]*Preset {
	if n <= 0 {
		return [][]*Preset{nil}
	}
	var out [][]*Preset
	var rec func(start int, cur []*Preset)
	rec = func(start int, cur []*Preset) {
		if len(cur) == n {
			out = append(out, append([]*Preset{}, cur...))
			return
		}
		for i := start; i < len(from); i++ {
			rec(i, append(cur, from[i]))
		}
	}
	rec(0, nil)
	return out
}

// toggles are the roles DPS members can bring, with what each one turns on.
var toggles = []Gift{GiftBattleShout, GiftFuriousHowl, GiftSunderTank, GiftMinorArmor, GiftCoE, GiftCrusader, GiftJoW, GiftMoonkin, GiftLotP, GiftTrueshot,
	GiftAB, GiftGotW, GiftPoS | GiftFort, GiftSoE | GiftWF}

// Plans is every buff set a raid can have (docs/raid_sim.md, The search, step 2).
func (r *Raid) Plans() []Plan {
	var crus []int
	if r.phPaladins > 0 {
		crus = []int{0, 1}
	} else {
		crus = []int{0}
	}
	var airs [][]proto.RaidTotem
	switch n := len(r.phShamans); {
	case n == 0:
		airs = [][]proto.RaidTotem{nil}
	case n == 1:
		airs = [][]proto.RaidTotem{{proto.RaidTotem_RaidTotemGraceOfAir}, {proto.RaidTotem_RaidTotemWindfury}}
	default:
		mk := func(a, b proto.RaidTotem) []proto.RaidTotem {
			out := []proto.RaidTotem{a, b}
			for i := 2; i < n; i++ {
				out = append(out, proto.RaidTotem_RaidTotemGraceOfAir)
			}
			return out
		}
		airs = [][]proto.RaidTotem{mk(proto.RaidTotem_RaidTotemGraceOfAir, proto.RaidTotem_RaidTotemWindfury),
			mk(proto.RaidTotem_RaidTotemGraceOfAir, proto.RaidTotem_RaidTotemGraceOfAir),
			mk(proto.RaidTotem_RaidTotemWindfury, proto.RaidTotem_RaidTotemWindfury)}
	}
	var pals []int
	for n := min(3, r.phPaladins); n <= 3; n++ {
		pals = append(pals, n)
	}
	seen := map[string]bool{}
	var out []Plan
	for _, pc := range crus {
		for _, air := range airs {
			for _, np := range pals {
				base := Plan{Paladins: np, PHCrusader: pc, PHAir: air}
				ph := r.phUp(base)
				var opts []Gift
				for _, t := range toggles {
					if t&^ph != 0 {
						opts = append(opts, t&^ph)
					}
				}
				for mask := 0; mask < 1<<len(opts); mask++ {
					up := ph &^ GiftBlessing
					for i, t := range opts {
						if mask&(1<<i) != 0 {
							up |= t
						}
					}
					if np > 0 {
						up |= GiftBlessing
					}
					p := base
					p.Up = up
					if !r.meets(p) {
						continue
					}
					if k := p.Key(); !seen[k] {
						seen[k] = true
						out = append(out, p)
					}
				}
			}
		}
	}
	return out
}

type scored struct {
	plan  Plan
	model float64
	comp  *Comp // exact
}

// run is one scenario.
func (sc *Scenario) run(req *proto.CompSearchRequest, fixed []*Build, progress Progress) (*proto.CompScenarioResult, error) {
	var phases []string
	lastPhase, lastRun := "", sc.Counter.Done()
	phaseStart := time.Now()
	var reportMu sync.Mutex
	report := func(phase string, best *Comp) {
		reportMu.Lock()
		defer reportMu.Unlock()
		if progress != nil {
			progress(phase, best, sc)
		}
		// Sims and seconds by phase, for the cost section's measurements.
		name := phase
		if strings.HasPrefix(name, "simming signature") {
			name = "signatures"
		}
		for _, prefix := range []string{"ranking buff sets", "confirmation", "combos"} {
			if strings.HasPrefix(name, prefix) {
				name = prefix
			}
		}
		if name != lastPhase {
			if lastPhase != "" {
				phases = append(phases, fmt.Sprintf("%s %d sims %.0f s", lastPhase, sc.Counter.Done()-lastRun, time.Since(phaseStart).Seconds()))
			}
			lastPhase, lastRun, phaseStart = name, sc.Counter.Done(), time.Now()
		}
	}
	mob := sc.Fight.MobType
	// The pool: base builds, and the encounter variants for this boss's creature type.
	var cands []*Build
	for _, b := range sc.Pool.Bases {
		cands = append(cands, b)
		if v, ok := b.Vs[mob]; ok && mob != proto.MobType_MobTypeUnknown {
			cands = append(cands, v)
		}
		cands = append(cands, b.Pets...)
	}
	var modelBuilds []*Build
	modelBuilds = append(modelBuilds, cands...)
	for _, b := range sc.Pool.Bases {
		for _, d := range b.Duties {
			modelBuilds = append(modelBuilds, d)
		}
	}
	report("model", nil)
	model, err := sc.BuildModel(modelBuilds)
	if err != nil {
		return nil, err
	}
	// A fixed base build runs its encounter variant against this boss, as the page's raids do.
	fixed = append([]*Build{}, fixed...)
	for i, b := range fixed {
		if b.IsBase() {
			fixed[i] = b.ForBoss(mob)
		}
	}
	survivors, dominatedBy := model.Survivors(cands, 1.5)
	f := &Fillers{Fill: append([]*Build{}, survivors...)}
	// The best shaman of each package kind fills too, for its party's mana.
	bestSham := map[proto.RaidPackage]*Build{}
	for _, b := range cands {
		if b.Class == proto.Class_ClassShaman {
			if cur, ok := bestSham[b.Pkg]; !ok || model.builds[b].full > model.builds[cur].full {
				bestSham[b.Pkg] = b
			}
		}
	}
	for _, b := range bestSham {
		if !containsBuild(f.Fill, b) {
			f.Fill = append(f.Fill, b)
		}
	}
	for _, b := range fixed {
		if !containsBuild(f.Fill, b) {
			f.Fill = append(f.Fill, b)
		}
	}
	sort.SliceStable(f.Fill, func(a, b int) bool { return f.Fill[a].ID < f.Fill[b].ID })
	f.Providers = pickProviders(model, modelBuilds)

	out := &proto.CompScenarioResult{Faction: sc.Faction, CampBuffs: sc.Camps, FightSeconds: sc.Fight.Duration}
	out.Notes = append(out.Notes, fmt.Sprintf("%d of %d builds survive pruning: %s", len(survivors), len(cands), buildList(survivors)))

	// Signatures: one search each, then every combo re-fills its signature's best sets.
	combos := Combos(req)
	sigs := map[Signature][]*Raid{}
	var sigOrder []Signature
	for _, c := range combos {
		r := NewRaid(int(req.Size), c, fixed)
		r.Required = RequiredGifts(req)
		s := r.Signature()
		if _, ok := sigs[s]; !ok {
			sigOrder = append(sigOrder, s)
		}
		sigs[s] = append(sigs[s], r)
	}
	out.Notes = append(out.Notes, fmt.Sprintf("%d tank and healer combos in %d signatures", len(combos), len(sigOrder)))

	// The model ranks every plan of every signature, in parallel.
	report("ranking buff sets", nil)
	results := make([]sigResult, len(sigOrder))
	var wg sync.WaitGroup
	sem := make(chan struct{}, runtime.GOMAXPROCS(0))
	mv := modelValuer{model}
	var ranked0 atomic.Int64
	for i, s := range sigOrder {
		i, s := i, s
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			r := sigs[s][0]
			plans := r.Plans()
			var ranked []scored
			for k, p := range plans {
				if k%256 == 0 && sc.Abort != nil && sc.Abort() {
					return
				}
				if c := sc.Fill(r, p, f, mv); c != nil {
					ranked = append(ranked, scored{plan: p, model: c.Score})
				}
			}
			sort.SliceStable(ranked, func(a, b int) bool { return ranked[a].model > ranked[b].model })
			results[i] = sigResult{s, ranked, len(plans)}
			report(fmt.Sprintf("ranking buff sets: signature %d of %d", int(ranked0.Add(1)), len(sigOrder)), nil)
		}()
	}
	wg.Wait()
	if sc.Abort != nil && sc.Abort() {
		return nil, ErrAborted
	}
	sort.SliceStable(results, func(a, b int) bool {
		if len(results[a].ranked) == 0 || len(results[b].ranked) == 0 {
			return len(results[a].ranked) > len(results[b].ranked)
		}
		return results[a].ranked[0].model > results[b].ranked[0].model
	})
	check := &proto.CompModelCheck{}
	for _, r := range results {
		check.SetsScored += int32(r.n)
	}
	// With required effects, the raid may have no buff set that gives them all.
	if len(results) == 0 || len(results[0].ranked) == 0 {
		if need := RequiredGifts(req); need != 0 {
			return nil, fmt.Errorf("no comp of these tanks, healers and report builds gives all of %s", need)
		}
		return nil, fmt.Errorf("no buff set could be filled with the report's builds")
	}

	// Exact sims of the best sets (step 3): every signature's top few, then for the leading signatures
	// every set the model's worst miss so far can't rule out.
	ev := newExactValuer(sc, sc.Iterations, model)
	var misses []float64
	exact := func(r *Raid, s *scored) error {
		if s.comp != nil {
			return nil
		}
		c, err := sc.EvalExact(r, s.plan, f, ev)
		if err != nil {
			return err
		}
		if c == nil {
			s.comp = &Comp{Score: math.Inf(-1)}
			return nil
		}
		s.comp = c
		misses = append(misses, math.Abs(c.Score-s.model))
		return nil
	}
	var best *Comp
	bestOf := func(sr *sigResult) *Comp {
		var b *Comp
		for i := range sr.ranked {
			if c := sr.ranked[i].comp; c != nil && !math.IsInf(c.Score, -1) && (b == nil || c.Score > b.Score) {
				b = c
			}
		}
		return b
	}
	const perSig = 4
	// Each phase is reported as it starts, so its sims and seconds are booked under its own name.
	for i := range results {
		report(fmt.Sprintf("simming signature %d of %d", i+1, len(results)), best)
		sr := &results[i]
		r := sigs[sr.sig][0]
		for j := 0; j < perSig && j < len(sr.ranked); j++ {
			if err := exact(r, &sr.ranked[j]); err != nil {
				return nil, err
			}
		}
		if b := bestOf(sr); b != nil && (best == nil || b.Score > best.Score) {
			best = b
		}
	}
	maxMiss := func() float64 {
		m := 0.0
		for _, x := range misses {
			m = math.Max(m, x)
		}
		return m
	}
	// The leading signatures by exact score get the full treatment.
	sort.SliceStable(results, func(a, b int) bool {
		ba, bb := bestOf(&results[a]), bestOf(&results[b])
		if ba == nil || bb == nil {
			return ba != nil
		}
		return ba.Score > bb.Score
	})
	const leading = 3
	const maxExact = 200
	report("exact sets", best)
	for i := 0; i < leading && i < len(results); i++ {
		sr := &results[i]
		r := sigs[sr.sig][0]
		// The model's worst miss on this signature's simmed sets bounds its error near the top.
		localMiss := func() float64 {
			m := 0.0
			for _, x := range sr.ranked {
				if x.comp != nil && !math.IsInf(x.comp.Score, -1) {
					m = math.Max(m, math.Abs(x.comp.Score-x.model))
				}
			}
			return m
		}
		for j := 0; j < len(sr.ranked) && j < maxExact; j++ {
			b := bestOf(sr)
			if b != nil && sr.ranked[j].model < b.Score-localMiss() {
				break
			}
			if err := exact(r, &sr.ranked[j]); err != nil {
				return nil, err
			}
		}
		if b := bestOf(sr); b != nil && (best == nil || b.Score > best.Score) {
			best = b
		}
		report("exact sets", best)
	}

	// Local moves from the best exact sets (step 5).
	report("local moves", best)
	for i := 0; i < leading && i < len(results); i++ {
		sr := &results[i]
		r := sigs[sr.sig][0]
		if err := sc.localMoves(r, sr, f, ev, model); err != nil {
			return nil, err
		}
		if b := bestOf(sr); b != nil && (best == nil || b.Score > best.Score) {
			best = b
		}
	}
	check.SetsSimmed = int32(len(misses))
	check.MaxMiss = maxMiss()
	if len(misses) > 0 {
		s := 0.0
		for _, x := range misses {
			s += x
		}
		check.MeanMiss = s / float64(len(misses))
	}
	out.ModelCheck = check

	// Every combo re-fills its signature's best sets with its own placeholders.
	report("combos", best)
	var cbs []comboBest
	sigTop := map[Signature][]Plan{}
	for i := range results {
		sr := &results[i]
		var cs []*scored
		for j := range sr.ranked {
			if c := sr.ranked[j].comp; c != nil && !math.IsInf(c.Score, -1) {
				cs = append(cs, &sr.ranked[j])
			}
		}
		sort.SliceStable(cs, func(a, b int) bool { return cs[a].comp.Score > cs[b].comp.Score })
		for j := 0; j < len(cs) && j < 3; j++ {
			sigTop[sr.sig] = append(sigTop[sr.sig], cs[j].plan)
		}
	}
	nCombo := 0
	// The combos of the best signatures first, so a limit keeps them.
	ranked := make([]Signature, 0, len(results))
	for _, r := range results {
		ranked = append(ranked, r.sig)
	}
	totalCombos := 0
	for _, s := range ranked {
		totalCombos += len(sigs[s])
	}
	if req.MaxCombos > 0 {
		totalCombos = min(totalCombos, int(req.MaxCombos))
	}
	for _, s := range ranked {
		for _, r := range sigs[s] {
			if req.MaxCombos > 0 && nCombo >= int(req.MaxCombos) {
				break
			}
			// Warm, this loop runs no sims, so it checks abort itself.
			if sc.Abort != nil && sc.Abort() {
				return nil, ErrAborted
			}
			if nCombo%50 == 0 {
				report(fmt.Sprintf("combos: %d of %d", nCombo, totalCombos), best)
			}
			nCombo++
			var cb *Comp
			for _, p := range sigTop[s] {
				c, err := sc.EvalExact(r, p, f, ev)
				if err != nil {
					return nil, err
				}
				if c != nil && (cb == nil || c.Score > cb.Score) {
					cb = c
				}
			}
			if cb != nil {
				cbs = append(cbs, comboBest{r, cb, s})
			}
		}
	}
	sort.SliceStable(cbs, func(a, b int) bool { return cbs[a].comp.Score > cbs[b].comp.Score })
	if len(cbs) == 0 {
		return nil, fmt.Errorf("no comp fits these slots")
	}
	best = cbs[0].comp

	// The contenders, re-simmed at more iterations.
	report("contenders", best)
	contenders := []*Comp{}
	seenPlan := map[string]bool{}
	for _, cb := range cbs {
		// Comps that resolve to the same request are one (two paladin presets, another healer that adds
		// nothing): the same members in the same seats.
		k := cb.comp.resolvedKey()
		if seenPlan[k] {
			continue
		}
		seenPlan[k] = true
		if cb.comp.Score >= best.Score-math.Max(0.001*best.Score, 0)-2.5*math.Hypot(best.SE, cb.comp.SE) || len(contenders) < 5 {
			contenders = append(contenders, cb.comp)
		}
		if len(contenders) >= 12 {
			break
		}
	}
	hi := newExactValuer(sc, req.ContenderIterations, model)
	for i, c := range contenders {
		cc, err := sc.EvalExact(c.Raid, c.Plan, f, hi)
		if err != nil {
			return nil, err
		}
		if cc != nil {
			contenders[i] = cc
		}
	}
	sort.SliceStable(contenders, func(a, b int) bool { return contenders[a].Score > contenders[b].Score })
	// The windows and gaps are measured from the top score; the comp shown first may be a simpler one
	// tied with it.
	top := contenders[0]
	contenders = showSimplest(contenders)
	best = contenders[0]
	report("done", best)

	// The confirmation (docs/raid_sim.md, Confirmation): every distinct buff set near the top raid-simmed,
	// paired on one seed; the top 2 again at more iterations on a fresh seed range. The raid sims rank
	// the confirmed comps.
	type confirmed struct {
		dps, se float64
	}
	conf := map[*Comp]confirmed{}
	var topConf *Comp // the best raid sim
	if req.Confirm {
		report("confirmation", best)
		seenSet := map[string]bool{}
		var sets []*Comp
		for _, c := range contenders {
			k := c.Plan.Key() + comboKey(c.Raid.Combo)
			if seenSet[k] {
				continue
			}
			seenSet[k] = true
			if c.Score >= top.Score-0.001*top.Score-2.5*math.Hypot(top.SE, c.SE) || len(sets) < 2 {
				sets = append(sets, c)
			}
		}
		total := len(sets) + min(2, len(sets))
		for k, c := range sets {
			if sc.Abort != nil && sc.Abort() {
				return nil, ErrAborted
			}
			report(fmt.Sprintf("confirmation: raid sim %d of %d", k+1, total), best)
			d, se, err := sc.confirm(c, req.ConfirmIterations, sc.Seed)
			if err != nil {
				return nil, err
			}
			conf[c] = confirmed{d, se}
		}
		// Members who cast debuffs or hold duties make a score run high against its raid sim (0.2 to 0.45%
		// in the third audit), more than the window allows for, so a set just outside it could beat the
		// confirmed ones. When the confirmed sets' scores ran high beyond noise, the window widens by the
		// most any did, once.
		lean := 0.0
		for _, c := range sets {
			if d := c.Score - conf[c].dps; d > 3*math.Hypot(conf[c].se, c.SE) {
				lean = math.Max(lean, d)
			}
		}
		if lean > 0 {
			doneSet := map[string]bool{}
			for _, c := range sets {
				doneSet[c.Plan.Key()+comboKey(c.Raid.Combo)] = true
			}
			for _, c := range contenders {
				k := c.Plan.Key() + comboKey(c.Raid.Combo)
				if doneSet[k] {
					continue
				}
				if c.Score < top.Score-0.001*top.Score-2.5*math.Hypot(top.SE, c.SE)-lean {
					continue
				}
				doneSet[k] = true
				if sc.Abort != nil && sc.Abort() {
					return nil, ErrAborted
				}
				report("confirmation: a set the scores' lean brings in", best)
				d, se, err := sc.confirm(c, req.ConfirmIterations, sc.Seed)
				if err != nil {
					return nil, err
				}
				conf[c] = confirmed{d, se}
				sets = append(sets, c)
			}
			out.Notes = append(out.Notes, fmt.Sprintf("the confirmed sets' scores ran up to %.1f above their raid sims, so the confirmation took in the sets within that much more", lean))
		}
		sort.SliceStable(sets, func(a, b int) bool { return conf[sets[a]].dps > conf[sets[b]].dps })
		fresh := sc.Seed + int64(req.ConfirmIterations) + 1_000_003
		for i := 0; i < 2 && i < len(sets); i++ {
			report(fmt.Sprintf("confirmation: raid sim %d of %d", len(sets)+i+1, total), best)
			d, se, err := sc.confirm(sets[i], req.ConfirmTopIterations, fresh)
			if err != nil {
				return nil, err
			}
			conf[sets[i]] = confirmed{d, se}
		}
		sort.SliceStable(contenders, func(a, b int) bool {
			ca, oka := conf[contenders[a]]
			cb, okb := conf[contenders[b]]
			if oka != okb {
				return oka
			}
			if oka {
				return ca.dps > cb.dps
			}
			return contenders[a].Score > contenders[b].Score
		})
		topConf = contenders[0]
		contenders = showSimplest(contenders, func(c *Comp) (float64, float64, bool) {
			x, ok := conf[c]
			return x.dps, x.se, ok
		})
		best = contenders[0]
		for _, c := range sets {
			if gap := conf[c].dps - c.Score; math.Abs(gap) > 3*math.Hypot(conf[c].se, c.SE) {
				out.Notes = append(out.Notes, fmt.Sprintf("the raid sim of %s differs from its score by %+.1f, more than noise", c.Describe(), gap))
			}
		}
	}

	// The result.
	for i, c := range contenders {
		if i >= 5 {
			break
		}
		cp := sc.compProto(c, top, model)
		if x, ok := conf[c]; ok {
			cp.Confirmed, cp.RaidSimDps, cp.RaidSimSe = true, x.dps, x.se
			if b, okb := conf[topConf]; okb && c != topConf {
				cp.RaidSimGap, cp.RaidSimGapSe = b.dps-x.dps, math.Hypot(b.se, x.se)
			}
		}
		out.Comps = append(out.Comps, cp)
	}
	if len(conf) > 0 {
		n := 0
		b := conf[topConf]
		for _, c := range contenders {
			x, ok := conf[c]
			if !ok || b.dps-x.dps > math.Max(2*math.Hypot(b.se, x.se), 0.001*b.dps) {
				break
			}
			n++
		}
		out.TieSet = int32(max(1, n))
	} else {
		out.TieSet = int32(tieSet(contenders, top))
	}
	for i, cb := range cbs {
		// The combo scores are all screening fills, so their gaps are against the best of them.
		cr := &proto.CompComboRank{Dps: cb.comp.Score, Gap: cbs[0].comp.Score - cb.comp.Score, Signature: cb.sig.String()}
		for _, p := range cb.raid.Combo {
			cr.Presets = append(cr.Presets, p.ID)
		}
		out.Combos = append(out.Combos, cr)
		if i >= 200 {
			break
		}
	}
	out.PresetValues = presetValues(best.Raid, cbs)
	// A health fight's passes before its length settles need only the winner's DPS.
	if sc.lean {
		report("end", best)
		out.Timings = phases
		return out, nil
	}
	report("extras", best)
	// The role table and member values at the screening iterations, with the best comp's members
	// re-valued there too so the differences are like for like.
	bestLo, err := sc.withMembers(best, best.Members, ev)
	if err != nil {
		return nil, err
	}
	extras, err := sc.extras(bestLo, f, ev, model, dominatedBy, cands)
	if err != nil {
		return nil, err
	}
	// A member value's index is into the best comp's result members (placeholders first, by party).
	if len(out.Comps) > 0 {
		for _, mv := range extras.members {
			m := bestLo.Members[mv.Member]
			for j, rm := range out.Comps[0].Members {
				if rm.BuildId == m.B.ID && rm.Party == int32(m.Party)+1 {
					mv.Member = int32(j)
					break
				}
			}
		}
	}
	if len(out.Comps) > 0 {
		cost := map[string]float64{}
		for _, mv := range extras.members {
			cost[mv.BuildId] = mv.DutyCost
		}
		for _, m := range out.Comps[0].Members {
			if len(m.Duties) > 0 {
				m.DutyCost = cost[m.BuildId]
			}
		}
	}
	out.Notes = append(out.Notes, extras.notes...)
	out.Roles = extras.roles
	out.Fillers = extras.fillers
	out.MemberValues = extras.members
	out.Adding = extras.adding
	report("end", best)
	out.Timings = phases
	// The buff records and misses are the resolver's for the shown comp, the same the page shows for it.
	if rc, err := sc.compOf(best, sc.Iterations, sc.Seed); err != nil {
		return nil, err
	} else if res := Resolve(rc); res.Error != nil {
		return nil, fmt.Errorf("%s", res.Error.Message)
	} else {
		out.Buffs, out.Misses = res.Buffs, res.Misses
	}
	if sc.Camps {
		out.Buffs = append(out.Buffs, &proto.CompBuffRecord{Effect: "camp buffs (all 9, each against its class buff)",
			Provider: "the camp", Scope: "raid", Receivers: int32(len(best.Members)), Given: true})
	}
	if !req.SkipSensitivity {
		out.Sensitivity = sc.assumptionSensitivity(contenders, best, func(c *Comp) (float64, float64, bool) {
			if len(conf) > 0 {
				x, ok := conf[c]
				return x.dps, x.se, ok
			}
			return c.Score, c.SE, true
		})
	}
	return out, nil
}

// resolvedKey is what a comp resolves to: its members in their seats, and its placeholders' gifts.
func (c *Comp) resolvedKey() string {
	var ks []string
	for _, m := range c.Members {
		ks = append(ks, askKey(m.B, m.Seat))
	}
	sort.Strings(ks)
	return fmt.Sprintf("%x|%d|%d|%d|", uint32(c.Raid.phUp(c.Plan)), c.Plan.Paladins, c.Raid.phPI, c.Raid.phInn) + strings.Join(ks, ";")
}

// showSimplest is the tie rule's choice of the comp to show (docs/raid_sim.md, Confirmation): among the
// comps tied with the best, the one with the fewest distinct builds, then the fewest members on duty,
// goes first. With a score function, ties are on its scores (the raid sims).
func showSimplest(cs []*Comp, score ...func(*Comp) (float64, float64, bool)) []*Comp {
	if len(cs) < 2 {
		return cs
	}
	val := func(c *Comp) (float64, float64, bool) { return c.Score, c.SE, true }
	if len(score) > 0 {
		val = score[0]
	}
	bv, bse, ok := val(cs[0])
	if !ok {
		return cs
	}
	tied := []int{0}
	for i := 1; i < len(cs); i++ {
		v, se, ok := val(cs[i])
		if !ok || bv-v > math.Max(2*math.Hypot(bse, se), 0.001*bv) {
			break
		}
		tied = append(tied, i)
	}
	simplest := func(c *Comp) (int, int) {
		ids := map[string]bool{}
		duty := 0
		for _, m := range c.Members {
			ids[m.B.ID] = true
			if m.B.Duty != proto.Duty_DutyNone {
				duty++
			}
		}
		return len(ids), duty
	}
	pick := tied[0]
	for _, i := range tied[1:] {
		a, b := simplest(cs[i])
		pa, pb := simplest(cs[pick])
		if a < pa || (a == pa && b < pb) {
			pick = i
		}
	}
	if pick != 0 {
		out := append([]*Comp{cs[pick]}, cs[:pick]...)
		cs = append(out, cs[pick+1:]...)
	}
	return cs
}

func comboKey(c []*Preset) string {
	s := ""
	for _, p := range c {
		s += p.Name + ","
	}
	return s
}

func containsBuild(list []*Build, b *Build) bool {
	for _, x := range list {
		if x == b {
			return true
		}
	}
	return false
}

func buildList(bs []*Build) string {
	s := ""
	for i, b := range bs {
		if i > 0 {
			s += ", "
		}
		s += b.ID
	}
	return s
}

// pickProviders keeps, for each role, the builds that give it best by the model's full-set DPS: a
// provider's cost is its DPS below the filler it displaces, so the strongest givers are the cheapest.
func pickProviders(m *Model, builds []*Build) []*Build {
	keep := map[*Build]bool{}
	for i := 0; i < numGifts; i++ {
		g := Gift(1) << i
		var givers []*Build
		for _, b := range builds {
			if b.Gives.Has(g) {
				givers = append(givers, b)
			}
		}
		sort.SliceStable(givers, func(a, b int) bool { return m.builds[givers[a]].full > m.builds[givers[b]].full })
		for j := 0; j < 4 && j < len(givers); j++ {
			keep[givers[j]] = true
		}
	}
	var out []*Build
	for _, b := range builds {
		if keep[b] {
			out = append(out, b)
		}
	}
	return out
}

type sigResult struct {
	sig    Signature
	ranked []scored
	n      int
}

// neighbours are the plans one move from a plan: a role more or less, a paladin more or less, another
// placeholder choice.
func (r *Raid) neighbours(p Plan) []Plan {
	var out []Plan
	ph := r.phUp(p)
	for _, t := range toggles {
		t &^= ph
		if t == 0 {
			continue
		}
		q := p
		q.PHAir = append([]proto.RaidTotem{}, p.PHAir...)
		q.Up ^= t
		out = append(out, q)
	}
	for _, d := range []int{-1, 1} {
		q := p
		q.Paladins += d
		if q.Paladins < min(3, r.phPaladins) || q.Paladins > 3 {
			continue
		}
		if q.Paladins > 0 {
			q.Up |= GiftBlessing
		} else {
			q.Up &^= GiftBlessing
		}
		out = append(out, q)
	}
	// A placeholder's choice changes what the placeholders give; what the DPS members bring stays.
	dps := p.Up &^ ph
	rebase := func(q Plan) Plan {
		q.Up = r.phUp(q)&^GiftBlessing | dps
		if q.Paladins > 0 {
			q.Up |= GiftBlessing
		}
		return q
	}
	if r.phPaladins > 0 {
		q := p
		q.PHCrusader = 1 - min(1, p.PHCrusader)
		out = append(out, rebase(q))
	}
	for i := range p.PHAir {
		q := p
		q.PHAir = append([]proto.RaidTotem{}, p.PHAir...)
		if q.PHAir[i] == proto.RaidTotem_RaidTotemGraceOfAir {
			q.PHAir[i] = proto.RaidTotem_RaidTotemWindfury
		} else {
			q.PHAir[i] = proto.RaidTotem_RaidTotemGraceOfAir
		}
		out = append(out, rebase(q))
	}
	kept := out[:0]
	for _, q := range out {
		if r.meets(q) {
			kept = append(kept, q)
		}
	}
	return kept
}

// localMoves polishes a signature's best exact sets (docs/raid_sim.md, The search, step 5): a move stays
// only when it beats the set by 2 standard errors of the difference, until none does.
func (sc *Scenario) localMoves(r *Raid, sr *sigResult, f *Fillers, ev *exactValuer, model *Model) error {
	// Only the sets already simmed exactly are done; the others were only ranked by the model, and a
	// move to one of them is what this step is for.
	seen := map[string]bool{}
	for _, s := range sr.ranked {
		if s.comp != nil {
			seen[s.plan.Key()] = true
		}
	}
	mv := modelValuer{model}
	for step := 0; step < 6; step++ {
		var cur *scored
		for i := range sr.ranked {
			if c := sr.ranked[i].comp; c != nil && !math.IsInf(c.Score, -1) && (cur == nil || c.Score > cur.comp.Score) {
				cur = &sr.ranked[i]
			}
		}
		if cur == nil {
			return nil
		}
		improved := false
		for _, q := range r.neighbours(cur.plan) {
			if seen[q.Key()] {
				continue
			}
			seen[q.Key()] = true
			mc := sc.Fill(r, q, f, mv)
			if mc == nil {
				continue
			}
			c, err := sc.EvalExact(r, q, f, ev)
			if err != nil {
				return err
			}
			if c == nil {
				continue
			}
			// A move stays only when it beats the set by 2 standard errors, so one better by noise alone
			// can't become the answer.
			if c.Score > cur.comp.Score+2*math.Hypot(c.SE, cur.comp.SE) {
				sr.ranked = append(sr.ranked, scored{plan: q, model: mc.Score, comp: c})
				improved = true
			}
		}
		if !improved {
			return nil
		}
	}
	return nil
}

func tieSet(cs []*Comp, best *Comp) int {
	n := 0
	for _, c := range cs {
		gap := best.Score - c.Score
		if gap <= math.Max(2*math.Hypot(best.SE, c.SE), 0.001*best.Score) {
			n++
		} else {
			break
		}
	}
	return max(1, n)
}
