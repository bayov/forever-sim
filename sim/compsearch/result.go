package compsearch

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/wowsims/classic/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

type comboBest struct {
	raid *Raid
	comp *Comp
	sig  Signature
}

func (sc *Scenario) compProto(c *Comp, best *Comp, model *Model) *proto.CompResult {
	out := &proto.CompResult{Dps: c.Score, Se: c.SE, Gap: best.Score - c.Score, GapSe: math.Hypot(best.SE, c.SE),
		BuffSet: c.Plan.String()}
	if model != nil {
		out.ModelDps = sc.Fill(c.Raid, c.Plan, &Fillers{Fill: membersBuilds(c), Providers: membersBuilds(c)}, modelValuer{model}).scoreOrZero()
	}
	for i, ph := range c.Raid.Combo {
		out.Combo = append(out.Combo, ph.ID)
		m := &proto.CompResultMember{Party: int32(c.PHParty[i]) + 1, Role: ph.Role, BuildId: ph.Name, Preset: ph.ID}
		if ph.Paladin {
			m.Judgement = proto.PaladinJudgement_PaladinJudgementWisdom
			if paladinIndex(c.Raid, i) < c.Plan.PHCrusader {
				m.Judgement = proto.PaladinJudgement_PaladinJudgementCrusader
			}
		}
		if ph.Shaman {
			m.Totems = []proto.RaidTotem{proto.RaidTotem_RaidTotemStrengthOfEarth, proto.RaidTotem_RaidTotemManaSpring, proto.RaidTotem_RaidTotemFlametongue}
			if k := shamanIndex(c.Raid, i); k < len(c.Plan.PHAir) {
				m.Totems = append(m.Totems, c.Plan.PHAir[k])
			}
		}
		out.Members = append(out.Members, m)
	}
	for _, mb := range c.Members {
		m := &proto.CompResultMember{Party: int32(mb.Party) + 1, Role: proto.RaidRole_RaidRoleDps, BuildId: mb.B.ID,
			Dps: mb.DPS, Se: mb.SE, Blessings: mb.Seat.Blessings, PowerInfusion: mb.Seat.PI, Innervate: mb.Seat.Innervate,
			Package: mb.Seat.Package, Race: strings.TrimPrefix(mb.B.Race[sc.Faction].String(), "Race"), ReportDps: mb.B.Base.Published}
		if mb.B.Duty != proto.Duty_DutyNone {
			m.Duties = []proto.Duty{mb.B.Duty}
		}
		out.Members = append(out.Members, m)
	}
	sort.SliceStable(out.Members, func(a, b int) bool { return out.Members[a].Party < out.Members[b].Party })
	return out
}

func (c *Comp) scoreOrZero() float64 {
	if c == nil {
		return 0
	}
	return c.Score
}

func membersBuilds(c *Comp) []*Build {
	seen := map[*Build]bool{}
	var out []*Build
	for _, m := range c.Members {
		if !seen[m.B] {
			seen[m.B] = true
			out = append(out, m.B)
		}
	}
	return out
}

func paladinIndex(r *Raid, i int) int {
	n := 0
	for j := 0; j < i; j++ {
		if r.Combo[j].Paladin {
			n++
		}
	}
	return n
}

func shamanIndex(r *Raid, i int) int {
	for k, p := range r.phShamans {
		if p == r.Combo[i] {
			// Two of one preset: count which one this is.
			n := 0
			for j := 0; j < i; j++ {
				if r.Combo[j] == p {
					n++
				}
			}
			return k + n
		}
	}
	return 0
}

// presetValues is each tank and healer preset's DPS value (docs/raid_sim.md, Tank and healer combos).
func presetValues(bestRaid *Raid, cbs []comboBest) []*proto.CompPresetValue {
	key := func(c []*Preset) string {
		ids := make([]int, len(c))
		for i, p := range c {
			ids[i] = int(p.ID)
		}
		sort.Ints(ids)
		return fmt.Sprint(ids)
	}
	score := map[string]float64{}
	for _, cb := range cbs {
		score[key(cb.raid.Combo)] = cb.comp.Score
	}
	best := cbs[0]
	var out []*proto.CompPresetValue
	for _, p := range Presets {
		v := &proto.CompPresetValue{Preset: p.ID}
		// At the best combo: the best combo with one of this preset's slots changed.
		if idx := indexOfPreset(best.raid.Combo, p); idx >= 0 {
			alt := math.Inf(-1)
			for _, q := range PresetsFor(p.Role) {
				if q == p {
					continue
				}
				c := append([]*Preset{}, best.raid.Combo...)
				c[idx] = q
				if s, ok := score[key(c)]; ok {
					alt = math.Max(alt, s)
				}
			}
			if !math.IsInf(alt, -1) {
				v.AtBest = best.comp.Score - alt
			}
		}
		with, without := math.Inf(-1), math.Inf(-1)
		for _, cb := range cbs {
			if indexOfPreset(cb.raid.Combo, p) >= 0 {
				with = math.Max(with, cb.comp.Score)
			} else {
				without = math.Max(without, cb.comp.Score)
			}
		}
		if !math.IsInf(with, -1) && !math.IsInf(without, -1) {
			v.BestWithMinusWithout = with - without
		}
		out = append(out, v)
	}
	return out
}

func comboIDs(c []*Preset) []proto.RaidPreset {
	out := make([]proto.RaidPreset, len(c))
	for i, p := range c {
		out[i] = p.ID
	}
	return out
}

func indexOfPreset(c []*Preset, p *Preset) int {
	for i, q := range c {
		if q == p {
			return i
		}
	}
	return -1
}

type extrasOut struct {
	roles   []*proto.CompRoleValue
	fillers []*proto.CompFillerRank
	members []*proto.CompMemberValue
	adding  []*proto.CompValueOfAdding
}

// extras is the role table, the filler ranking and the member values at the best comp (docs/raid_sim.md,
// The result and Member value).
func (sc *Scenario) extras(best *Comp, f *Fillers, ev *exactValuer, model *Model, dominatedBy map[*Build]*Build, cands []*Build) (*extrasOut, error) {
	out := &extrasOut{}
	r := best.Raid
	ph := r.phUp(best.Plan)
	// The role table: each role removed from the winning set and re-filled, or added when it's missing.
	for _, g := range toggles {
		name := GiftName(g &^ GiftFort)
		if g == GiftSoE|GiftWF {
			name = "Enhancement's totems (Strength of Earth, Windfury)"
		}
		q := best.Plan
		q.PHAir = append([]proto.RaidTotem{}, best.Plan.PHAir...)
		in := best.Plan.Up.Has(g &^ ph)
		if g&^ph == 0 {
			// The placeholders give it at no cost.
			if in {
				out.roles = append(out.roles, &proto.CompRoleValue{Role: name, Provider: "placeholders", InSet: true})
			}
			continue
		}
		if in {
			q.Up &^= g &^ ph
		} else {
			q.Up |= g
		}
		c, err := sc.EvalExact(r, q, f, ev)
		if err != nil {
			return nil, err
		}
		if c == nil {
			continue
		}
		rv := &proto.CompRoleValue{Role: name, InSet: in}
		if in {
			rv.Net = best.Score - c.Score
			prov, cost := providerCost(best, g, sc, ev, f)
			rv.Provider, rv.ProviderCost = prov, cost
			rv.Gain = rv.Net + cost
		} else {
			rv.Net = c.Score - best.Score
			rv.Provider = "value of adding"
		}
		out.roles = append(out.roles, rv)
	}
	// A paladin's blessings.
	if best.Plan.Paladins > min(3, r.phPaladins) {
		q := best.Plan
		q.Paladins--
		if q.Paladins == 0 {
			q.Up &^= GiftBlessing
		}
		if c, err := sc.EvalExact(r, q, f, ev); err != nil {
			return nil, err
		} else if c != nil {
			out.roles = append(out.roles, &proto.CompRoleValue{Role: "a Ret paladin's blessings", InSet: true, Net: best.Score - c.Score, Provider: "Ret paladin"})
		}
	}

	// The filler ranking under the winning set, without a package.
	shift := loadAssumptionShifts(sc.Pool.Root)
	type fr struct {
		b      *Build
		v, se  float64
		lo, hi float64
	}
	var frs []fr
	for _, b := range cands {
		if !allowed(b, best.Plan) || (b.Class == proto.Class_ClassPaladin && best.Plan.Paladins < 3) {
			continue
		}
		s := best.seat(b, 0, sc)
		v, se, _ := ev.V(b, s)
		frs = append(frs, fr{b: b, v: v, se: se})
	}
	if _, err := ev.flush(); err != nil {
		return nil, err
	}
	for i := range frs {
		frs[i].v, frs[i].se, _ = ev.V(frs[i].b, best.seat(frs[i].b, 0, sc))
		for _, d := range shift[frs[i].b.Base.ID] {
			frs[i].lo = math.Min(frs[i].lo, d)
			frs[i].hi = math.Max(frs[i].hi, d)
		}
	}
	sort.SliceStable(frs, func(a, b int) bool { return frs[a].v > frs[b].v })
	for _, x := range frs {
		out.fillers = append(out.fillers, &proto.CompFillerRank{BuildId: x.b.ID, Dps: x.v, Se: x.se,
			AssumptionLow: x.v + x.lo, AssumptionHigh: x.v + x.hi,
			Interchangeable: len(frs) > 0 && frs[0].v-x.v <= 2*math.Hypot(frs[0].se, x.se)})
	}

	// Member values, one per distinct member (identical members share one value).
	seen := map[string]bool{}
	for i, m := range best.Members {
		k := askKey(m.B, m.Seat)
		if seen[k] {
			continue
		}
		seen[k] = true
		mv, err := sc.memberValue(best, i, f, ev, cands)
		if err != nil {
			return nil, err
		}
		out.members = append(out.members, mv)
	}

	// The value of adding a build that isn't in the comp: swapped in for the member it beats most.
	in := map[*Build]bool{}
	for _, m := range best.Members {
		in[m.B.Base] = true
	}
	for _, b := range cands {
		if in[b.Base] || !b.IsBase() {
			continue
		}
		bestV := math.Inf(-1)
		tried := map[*Build]bool{}
		for i, m := range best.Members {
			if tried[m.B] {
				continue
			}
			tried[m.B] = true
			v, ok, err := sc.swapScore(best, i, b, ev)
			if err != nil {
				return nil, err
			}
			if ok {
				bestV = math.Max(bestV, v)
			}
		}
		if !math.IsInf(bestV, -1) {
			out.adding = append(out.adding, &proto.CompValueOfAdding{BuildId: b.ID, Value: bestV - best.Score})
		}
	}
	sort.SliceStable(out.adding, func(a, b int) bool { return out.adding[a].Value > out.adding[b].Value })
	return out, nil
}

// providerCost is the member who brings a role and what it costs them: their DPS below the best filler
// in their seat.
func providerCost(c *Comp, g Gift, sc *Scenario, ev *exactValuer, f *Fillers) (string, float64) {
	for _, m := range c.Members {
		if m.B.Gives&g == 0 {
			continue
		}
		bestV := math.Inf(-1)
		for _, b := range f.Fill {
			if allowed(b, c.Plan) && b.Class != proto.Class_ClassShaman && (b.Class != proto.Class_ClassPaladin || c.Plan.Paladins >= 3) {
				s := c.seat(b, m.Seat.Package, sc)
				v, _, _ := ev.V(b, s)
				bestV = math.Max(bestV, v)
			}
		}
		_, _ = ev.flush()
		return m.B.ID, math.Max(0, bestV-m.DPS)
	}
	return "", 0
}

// withMembers is the comp with its members changed: the set they and the placeholders give, each member
// re-valued under it.
func (sc *Scenario) withMembers(c *Comp, members []*Member, ev *exactValuer) (*Comp, error) {
	r := c.Raid
	p := c.Plan
	p.PHAir = append([]proto.RaidTotem{}, c.Plan.PHAir...)
	up := r.phUp(p) &^ GiftBlessing
	pal := r.phPaladins
	for _, m := range members {
		up |= m.B.Gives &^ GiftBlessing
		if m.B.Class == proto.Class_ClassPaladin {
			pal++
		}
	}
	p.Up = up
	p.Paladins = min(3, pal)
	if p.Paladins > 0 {
		p.Up |= GiftBlessing
	}
	nc := &Comp{Raid: r, Plan: p, PHParty: c.PHParty, Blessings: c.Blessings}
	if p.Paladins != c.Plan.Paladins {
		nc.Blessings = map[proto.Class][]proto.Blessing{}
		for k := range c.Blessings {
			ch := blessingChoices(p.Paladins)
			nc.Blessings[k] = ch[0]
			if len(c.Blessings[k]) > 0 && p.Paladins > 0 {
				nc.Blessings[k] = c.Blessings[k][:min(len(c.Blessings[k]), p.Paladins)]
			}
		}
	}
	for _, m := range members {
		if m.B.Class == proto.Class_ClassWarrior {
			nc.Warriors++
		}
	}
	// Each party's package from the givers still in the raid: its shaman placeholder, or a DPS shaman.
	pkg := map[int]proto.RaidPackage{}
	for i, ph := range r.Combo {
		if ph.Shaman && ph.Package > pkg[c.PHParty[i]] {
			pkg[c.PHParty[i]] = ph.Package
		}
	}
	for _, m := range members {
		if m.B.Class == proto.Class_ClassShaman && m.B.Pkg > pkg[m.Party] {
			pkg[m.Party] = m.B.Pkg
		}
	}
	for _, m := range members {
		nm := *m
		p := pkg[m.Party]
		if m.B.Class == proto.Class_ClassShaman {
			p = proto.RaidPackage_RaidPackageNone // a shaman casts its own Mana Spring, so it's never given one too
		}
		if !usesMana(m.B) {
			p = m.Seat.Package // unread: keep the seat as it was
		}
		nm.Seat = nc.seat(m.B, p, sc)
		nm.Seat.PI, nm.Seat.Innervate = m.Seat.PI, m.Seat.Innervate
		nc.Members = append(nc.Members, &nm)
	}
	for _, m := range nc.Members {
		ev.V(m.B, m.Seat)
	}
	if _, err := ev.flush(); err != nil {
		return nil, err
	}
	for _, m := range nc.Members {
		m.DPS, m.SE, m.key = ev.V(m.B, m.Seat)
	}
	nc.sum()
	return nc, nil
}

func (sc *Scenario) swapScore(c *Comp, i int, b *Build, ev *exactValuer) (float64, bool, error) {
	members := make([]*Member, len(c.Members))
	copy(members, c.Members)
	m := *c.Members[i]
	m.B = b
	m.Seat.PI, m.Seat.Innervate = false, false
	members[i] = &m
	nc, err := sc.withMembers(c, members, ev)
	if err != nil {
		return 0, false, err
	}
	return nc.Score, true, nil
}

// memberValue is a member's contribution and replacement margin (docs/raid_sim.md, Member value).
func (sc *Scenario) memberValue(c *Comp, i int, f *Fillers, ev *exactValuer, cands []*Build) (*proto.CompMemberValue, error) {
	m := c.Members[i]
	out := &proto.CompMemberValue{Member: int32(i), OwnDps: m.DPS, BuildId: m.B.ID}
	for _, x := range c.Members {
		if askKey(x.B, x.Seat) == askKey(m.B, m.Seat) {
			out.Count++
		}
	}
	// Without the member: the others under the set they still give.
	var rest []*Member
	for j, x := range c.Members {
		if j != i {
			rest = append(rest, x)
		}
	}
	nc, err := sc.withMembers(c, rest, ev)
	if err != nil {
		return nil, err
	}
	out.Contribution = c.Score - nc.Score
	// The best other build in that slot.
	bestV, bestB := math.Inf(-1), ""
	for _, b := range cands {
		if b == m.B || b == m.B.Base {
			continue
		}
		v, ok, err := sc.swapScore(c, i, b, ev)
		if err != nil {
			return nil, err
		}
		if ok && v > bestV {
			bestV, bestB = v, b.ID
		}
	}
	if bestB != "" {
		out.ReplacementMargin = c.Score - bestV
		out.BestReplacement = bestB
	}
	// The difference from the DPS Report, one change at a time.
	d, parts, err := sc.reportDifference(m, ev)
	if err != nil {
		return nil, err
	}
	out.ReportDifference = d
	out.ReportDifferenceParts = parts
	return out, nil
}

// reportDifference breaks a member's DPS less its report build's published DPS into single-player sims:
// the report's own request, at the faction's race, under the raid's buff set, then on duty.
func (sc *Scenario) reportDifference(m *Member, ev *exactValuer) (float64, []string, error) {
	base := m.B.Base
	if base.Published == 0 {
		return 0, nil, nil
	}
	reportJob := func(race bool) (*Job, error) {
		req := googleProto.Clone(base.Req).(*proto.RaidSimRequest)
		if race {
			pl, err := sc.Pool.Player(base, sc.Faction)
			if err != nil {
				return nil, err
			}
			req.Raid.Parties[0].Players[0] = pl
		}
		req.SimOptions = &proto.SimOptions{Iterations: ev.iters, RandomSeed: sc.Seed, UseLabeledRands: sc.Sim.Labeled}
		j := &Job{Build: base, Faction: sc.Faction, Iterations: ev.iters, Seed: sc.Seed}
		j.req = req
		return j, nil
	}
	j0, err := reportJob(false)
	if err != nil {
		return 0, nil, err
	}
	j1, err := reportJob(true)
	if err != nil {
		return 0, nil, err
	}
	seat := m.Seat
	seat.Own = base.Casts
	j2 := sc.job(base, seat, ev.iters)
	j3 := sc.job(m.B, m.Seat, ev.iters)
	if err := sc.sim([]*Job{j0, j1, j2, j3}); err != nil {
		return 0, nil, err
	}
	parts := []string{
		fmt.Sprintf("the report's request in this sim: %+.1f", j0.Value.DPS-base.Published),
		fmt.Sprintf("race (%s): %+.1f", strings.TrimPrefix(base.Race[sc.Faction].String(), "Race"), j1.Value.DPS-j0.Value.DPS),
		fmt.Sprintf("the raid's buff set: %+.1f", j2.Value.DPS-j1.Value.DPS),
	}
	if m.B != base {
		parts = append(parts, fmt.Sprintf("its duty (%s): %+.1f", m.B.Duty, j3.Value.DPS-j2.Value.DPS))
	}
	return m.DPS - base.Published, parts, nil
}

// buffRecords is each effect up, who gives it and how many members get it.
func buffRecords(c *Comp) []*proto.CompBuffRecord {
	var out []*proto.CompBuffRecord
	ph := c.Raid.phUp(c.Plan)
	for i := 0; i < numGifts; i++ {
		g := Gift(1) << i
		if !c.Plan.Up.Has(g) || g == GiftBlessing {
			continue
		}
		rec := &proto.CompBuffRecord{Effect: GiftName(g), Scope: "raid", Receivers: int32(len(c.Members)), Given: true}
		var who []string
		if ph.Has(g) {
			for _, p := range c.Raid.Combo {
				if p.Gives.Has(g) || (p.Shaman && (g == GiftSoE || g == GiftFT || g == GiftGoA || g == GiftWF)) ||
					(p.Paladin && (g == GiftCrusader || g == GiftJoW)) {
					who = append(who, p.Name)
				}
			}
		}
		for _, m := range c.Members {
			if m.B.Gives.Has(g) {
				who = append(who, m.B.ID)
				if m.B.Casts.Has(g) {
					rec.Given = false
				}
			}
		}
		rec.Provider = strings.Join(uniq(who), ", ")
		switch g {
		case GiftSunderTank, GiftMinorArmor, GiftCoE, GiftCrusader, GiftJoW:
			rec.Scope = "the boss"
		}
		out = append(out, rec)
	}
	for _, p := range packages[1:] {
		n := 0
		for _, m := range c.Members {
			if m.Seat.Package == p {
				n++
			}
		}
		if n > 0 {
			var who []string
			for _, ph := range c.Raid.Combo {
				if ph.Shaman && ph.Package == p {
					who = append(who, ph.Name)
				}
			}
			for _, m := range c.Members {
				if m.B.Class == proto.Class_ClassShaman && m.B.Pkg == p {
					who = append(who, m.B.ID)
				}
			}
			out = append(out, &proto.CompBuffRecord{Effect: packageName(p), Provider: strings.Join(uniq(who), ", "),
				Scope: "party", Receivers: int32(n), Given: true})
		}
	}
	return out
}

func packageName(p proto.RaidPackage) string {
	switch p {
	case proto.RaidPackage_RaidPackageManaSpring:
		return "Mana Spring"
	case proto.RaidPackage_RaidPackageManaSpringTide:
		return "Mana Spring and Mana Tide"
	case proto.RaidPackage_RaidPackageRestoration:
		return "Mana Spring ×1.25 and Mana Tide"
	}
	return "none"
}

func uniq(s []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, x := range s {
		if !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	return out
}

// missList is what the raid lacks: roles nobody gives, and blessings a class goes without.
func missList(c *Comp) []string {
	var out []string
	for _, g := range modelRoles {
		if !c.Plan.Up.Has(g) {
			out = append(out, "nobody gives "+GiftName(g))
		}
	}
	present := map[proto.Class]bool{}
	for _, m := range c.Members {
		present[m.B.Class] = true
	}
	for cl := range present {
		have := c.Blessings[cl]
		if len(have) < 3 {
			var miss []string
			for _, b := range allBlessings {
				found := false
				for _, h := range have {
					if h == b {
						found = true
					}
				}
				if !found {
					miss = append(miss, strings.TrimPrefix(b.String(), "Blessing"))
				}
			}
			out = append(out, fmt.Sprintf("%s get no %s", strings.TrimPrefix(cl.String(), "Class"), strings.Join(miss, " or ")))
		}
	}
	sort.Strings(out)
	return out
}

// loadAssumptionShifts is each build's DPS change under each of the report's open assumptions
// (assumption_dps.json), at 180 s.
func loadAssumptionShifts(root string) map[string]map[string]float64 {
	data, err := os.ReadFile(filepath.Join(root, "ui", "dps_report", "data", "assumption_dps.json"))
	if err != nil {
		return nil
	}
	var rows []struct {
		Build  string   `json:"build"`
		Combo  string   `json:"combo"`
		Kind   string   `json:"kind"`
		DPS180 *float64 `json:"dps180"`
	}
	if json.Unmarshal(data, &rows) != nil {
		return nil
	}
	base := map[string]float64{}
	for _, r := range rows {
		if r.Kind == "base" && r.DPS180 != nil {
			base[r.Build] = *r.DPS180
		}
	}
	out := map[string]map[string]float64{}
	for _, r := range rows {
		if r.Kind != "single" && r.Kind != "grid" || r.DPS180 == nil {
			continue
		}
		b, ok := base[r.Build]
		if !ok {
			continue
		}
		if out[r.Build] == nil {
			out[r.Build] = map[string]float64{}
		}
		out[r.Build][r.Combo] = *r.DPS180 - b
	}
	return out
}

// assumptionSensitivity shifts each contender's members by the report's open assumptions, one combo at a
// time, and says whether the winner holds (docs/raid_sim.md, Sensitivity: no sims).
func (sc *Scenario) assumptionSensitivity(contenders []*Comp, best *Comp) []*proto.CompSensitivity {
	shift := loadAssumptionShifts(sc.Pool.Root)
	combos := map[string]bool{}
	for _, m := range shift {
		for k := range m {
			combos[k] = true
		}
	}
	var names []string
	for k := range combos {
		names = append(names, k)
	}
	sort.Strings(names)
	var out []*proto.CompSensitivity
	for _, name := range names {
		var vs []shifted
		for _, c := range contenders {
			v := c.Score
			for _, m := range c.Members {
				v += shift[m.B.Base.ID][name]
			}
			vs = append(vs, shifted{c, v})
		}
		sort.SliceStable(vs, func(a, b int) bool { return vs[a].v > vs[b].v })
		res := &proto.CompSensitivity{Scenario: name, WinnerHolds: vs[0].c == best ||
			vs[0].v-valueOf(vs, best) <= math.Max(2*best.SE, 0.001*best.Score)}
		if !res.WinnerHolds {
			res.Winners = append(res.Winners, &proto.CompResult{Dps: vs[0].v, Gap: vs[0].v - valueOf(vs, best),
				BuffSet: vs[0].c.Describe() + "; " + comboText(comboIDs(vs[0].c.Raid.Combo))})
		}
		out = append(out, res)
	}
	return out
}

type shifted struct {
	c *Comp
	v float64
}

func valueOf(vs []shifted, c *Comp) float64 {
	for _, x := range vs {
		if x.c == c {
			return x.v
		}
	}
	return math.Inf(-1)
}
