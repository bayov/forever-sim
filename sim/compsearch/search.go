package compsearch

import (
	"fmt"
	"math"
	"runtime"
	"sort"
	"sync"
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

// Search runs a comp search (docs/raid_sim.md, Comp search): one scenario for each faction and camp
// setting.
func Search(pool *Pool, sim *Simmer, req *proto.CompSearchRequest, progress Progress) (*proto.CompSearchResult, error) {
	return SearchWithAbort(pool, sim, req, progress, nil)
}

// SearchWithAbort is Search that stops with ErrAborted once abort says so.
func SearchWithAbort(pool *Pool, sim *Simmer, req *proto.CompSearchRequest, progress Progress, abort func() bool) (*proto.CompSearchResult, error) {
	start := time.Now()
	Defaults(req)
	if req.Tanks+req.Healers >= req.Size {
		return nil, fmt.Errorf("%d tanks and %d healers leave no DPS slot in a raid of %d", req.Tanks, req.Healers, req.Size)
	}
	if len(req.FixedTanks) > int(req.Tanks) || len(req.FixedHealers) > int(req.Healers) {
		return nil, fmt.Errorf("more fixed tank or healer presets than slots")
	}
	var fixed []*Build
	for _, id := range req.FixedDps {
		b, ok := pool.Builds[id]
		if !ok || !b.IsBase() {
			return nil, fmt.Errorf("fixed DPS member %q is not a base build of the report", id)
		}
		fixed = append(fixed, b)
	}
	if len(fixed) > int(req.Size-req.Tanks-req.Healers) {
		return nil, fmt.Errorf("more fixed DPS members than DPS slots")
	}
	run0, hits0 := sim.Run.Load(), sim.Hits.Load()
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
				Seed: req.Seed, Assumptions: req.Assumptions, Abort: abort,
				Fight: Fight{Duration: req.Encounter.Duration, DurationVariation: req.Encounter.DurationVariation,
					MobType: req.Encounter.MobType, FlaskZone: req.Encounter.NaturalFlaskZone}}
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
	res.SimsRun = sim.Run.Load() - run0
	res.CacheHits = sim.Hits.Load() - hits0
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
	var out *proto.CompScenarioResult
	for pass := 0; pass < 4; pass++ {
		sc.HealthLength = L
		// The model ranks at the nearest grid point; the exact values interpolate.
		sc.Fight.Duration = math.Max(lengthGrid, math.Round(L/lengthGrid)*lengthGrid)
		sc.Fight.DurationVariation = 0
		var err error
		out, err = sc.run(req, fixed, progress)
		if err != nil {
			return nil, err
		}
		next := health / out.Comps[0].Dps
		out.FightSeconds = L
		out.Notes = append(out.Notes, fmt.Sprintf("health fight pass %d: %.1f s, the winner kills the boss in %.1f s", pass+1, L, next))
		if math.Abs(next-L) <= 1 {
			break
		}
		L = next
	}
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
	return fmt.Sprintf("warrior tank %v, feral tank %v, druid %v, paladins %d, shamans %d, priest %v",
		s.WarriorTank, s.FeralTank, s.Druid, s.Paladins, s.Shamans, s.Priest)
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
	lastPhase, lastRun := "", sc.Sim.Run.Load()
	phaseStart := time.Now()
	report := func(phase string, best *Comp) {
		if progress != nil {
			progress(phase, best, sc)
		}
		// Sims and seconds by phase, for the cost section's measurements.
		name := phase
		if len(name) > 9 && name[:9] == "signature" {
			name = "signatures"
		}
		if name != lastPhase {
			if lastPhase != "" {
				phases = append(phases, fmt.Sprintf("%s %d sims %.0f s", lastPhase, sc.Sim.Run.Load()-lastRun, time.Since(phaseStart).Seconds()))
			}
			lastPhase, lastRun, phaseStart = name, sc.Sim.Run.Load(), time.Now()
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
			for _, p := range plans {
				if c := sc.Fill(r, p, f, mv); c != nil {
					ranked = append(ranked, scored{plan: p, model: c.Score})
				}
			}
			sort.SliceStable(ranked, func(a, b int) bool { return ranked[a].model > ranked[b].model })
			results[i] = sigResult{s, ranked, len(plans)}
		}()
	}
	wg.Wait()
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
	for i := range results {
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
		report(fmt.Sprintf("signature %d of %d", i+1, len(results)), best)
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
		if b := bestOf(sr); b != nil && b.Score > best.Score {
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
		if b := bestOf(sr); b != nil && b.Score > best.Score {
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
	for _, s := range sigOrder {
		for _, r := range sigs[s] {
			if req.MaxCombos > 0 && nCombo >= int(req.MaxCombos) {
				break
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
		k := cb.comp.Plan.Key() + cb.comp.Describe() + comboKey(cb.raid.Combo)
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
	best = contenders[0]
	report("done", best)

	// The confirmation (docs/raid_sim.md, Confirmation): every distinct buff set near the top raid-simmed,
	// paired on one seed; the top 2 again at more iterations on a fresh seed range. The raid sims rank
	// the confirmed comps.
	type confirmed struct {
		dps, se float64
	}
	conf := map[*Comp]confirmed{}
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
			if c.Score >= best.Score-0.001*best.Score-2.5*math.Hypot(best.SE, c.SE) || len(sets) < 2 {
				sets = append(sets, c)
			}
		}
		for _, c := range sets {
			d, se, err := sc.confirm(c, req.ConfirmIterations, sc.Seed)
			if err != nil {
				return nil, err
			}
			conf[c] = confirmed{d, se}
		}
		sort.SliceStable(sets, func(a, b int) bool { return conf[sets[a]].dps > conf[sets[b]].dps })
		fresh := sc.Seed + int64(req.ConfirmIterations) + 1_000_003
		for i := 0; i < 2 && i < len(sets); i++ {
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
		cp := sc.compProto(c, best, model)
		if x, ok := conf[c]; ok {
			cp.Confirmed, cp.RaidSimDps, cp.RaidSimSe = true, x.dps, x.se
			if b, okb := conf[best]; okb {
				cp.Gap, cp.GapSe = b.dps-x.dps, math.Hypot(b.se, x.se)
			}
		}
		out.Comps = append(out.Comps, cp)
	}
	if len(conf) > 0 {
		n := 0
		b := conf[best]
		for _, c := range contenders {
			x, ok := conf[c]
			if !ok || b.dps-x.dps > math.Max(2*math.Hypot(b.se, x.se), 0.001*b.dps) {
				break
			}
			n++
		}
		out.TieSet = int32(max(1, n))
	} else {
		out.TieSet = int32(tieSet(contenders, best))
	}
	for i, cb := range cbs {
		cr := &proto.CompComboRank{Dps: cb.comp.Score, Gap: best.Score - cb.comp.Score, Signature: cb.sig.String()}
		for _, p := range cb.raid.Combo {
			cr.Presets = append(cr.Presets, p.ID)
		}
		out.Combos = append(out.Combos, cr)
		if i >= 200 {
			break
		}
	}
	out.PresetValues = presetValues(cbs[0].raid, cbs)
	report("extras", best)
	// The role table and member values at the screening iterations, with the best comp's members
	// re-valued there too so the differences are like for like.
	bestLo, err := sc.EvalExact(best.Raid, best.Plan, f, ev)
	if err != nil {
		return nil, err
	}
	if bestLo == nil {
		bestLo = best
	}
	extras, err := sc.extras(bestLo, f, ev, model, dominatedBy, cands)
	if err != nil {
		return nil, err
	}
	out.Roles = extras.roles
	out.Fillers = extras.fillers
	out.MemberValues = extras.members
	out.Adding = extras.adding
	report("end", best)
	out.Timings = phases
	out.Buffs = buffRecords(best)
	out.Misses = missList(best)
	if !req.SkipSensitivity {
		out.Sensitivity = sc.assumptionSensitivity(contenders, best)
	}
	return out, nil
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
	return out
}

// localMoves polishes a signature's best exact sets (docs/raid_sim.md, The search, step 5): a move stays
// only when it beats the set by 2 standard errors of the difference, until none does.
func (sc *Scenario) localMoves(r *Raid, sr *sigResult, f *Fillers, ev *exactValuer, model *Model) error {
	seen := map[string]bool{}
	for _, s := range sr.ranked {
		seen[s.plan.Key()] = true
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
			sr.ranked = append(sr.ranked, scored{plan: q, model: mc.Score, comp: c})
			if c.Score > cur.comp.Score+2*math.Hypot(c.SE, cur.comp.SE) {
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
