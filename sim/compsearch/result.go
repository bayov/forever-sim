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
			Dps: mb.DPS, Se: mb.SE, Tps: mb.TPS, Blessings: mb.Seat.Blessings, PowerInfusion: mb.Seat.PI, Innervate: mb.Seat.Innervate,
			Package: mb.Seat.Package, Race: strings.TrimPrefix(mb.B.Race[sc.Faction].String(), "Race"), ReportDps: reportBaseline(mb.B).Published}
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
	// At the winner's combo, which after the contenders and the confirmation needn't be the combo that
	// screened best.
	best := cbs[0]
	for _, cb := range cbs {
		if key(cb.raid.Combo) == key(bestRaid.Combo) {
			best = cb
			break
		}
	}
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
				v.AtBest, v.HasAtBest = best.comp.Score-alt, true
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
			v.BestWithMinusWithout, v.HasBestWithMinusWithout = with-without, true
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
	notes   []string
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
	// The role table compares re-fills, and a re-fill picks its builds on the screening's noise (six fire
	// mages trade for six frostfire on 45 DPS of it at 100 iterations). So the winning set is re-filled the
	// same way for the baseline, and the better of that and the winner's own members is the one to beat.
	roleBase := best.Score
	if ref, err := sc.EvalExact(r, best.Plan, f, ev); err != nil {
		return nil, err
	} else if ref != nil && ref.Score > roleBase {
		roleBase = ref.Score
	}
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
				out.roles = append(out.roles, &proto.CompRoleValue{Role: name, Provider: "placeholders (not measured: they give it at no cost)", InSet: true})
			}
			continue
		}
		if in {
			q.Up &^= g &^ ph
		} else {
			q.Up |= g
			// A placeholder paladin judging Wisdom can judge the Crusader instead, at no member's cost.
			if g == GiftCrusader && r.phPaladins > q.PHCrusader {
				q.PHCrusader++
				q.Up = r.phUp(q)&^GiftBlessing | (best.Plan.Up &^ ph)
				if q.Paladins > 0 {
					q.Up |= GiftBlessing
				}
			}
		}
		c, err := sc.EvalExact(r, q, f, ev)
		if err != nil {
			return nil, err
		}
		// A role whose givers all bring something else too can't be toggled alone: it goes or comes with
		// a giver's other roles (and the Crusader with one more paladin), the best of those.
		with := ""
		if c == nil {
			for _, b := range f.Providers {
				if b.Gives&g == 0 {
					continue
				}
				extra := b.Gives &^ GiftBlessing &^ g &^ ph
				q2 := q
				q2.PHAir = append([]proto.RaidTotem{}, q.PHAir...)
				if in {
					q2.Up &^= extra
				} else {
					q2.Up |= extra
					if b.Class == proto.Class_ClassPaladin {
						q2.Paladins = min(3, q2.Paladins+1)
						q2.Up |= GiftBlessing
					}
				}
				if q2.Key() == q.Key() {
					continue
				}
				c2, err := sc.EvalExact(r, q2, f, ev)
				if err != nil {
					return nil, err
				}
				if c2 != nil && (c == nil || c2.Score > c.Score) {
					c, with = c2, ""
					if extra != 0 {
						with = extra.String()
					}
					if b.Class == proto.Class_ClassPaladin && !in {
						with = strings.TrimPrefix(with+" and a paladin", " and ")
					}
				}
			}
		}
		if c == nil {
			out.roles = append(out.roles, &proto.CompRoleValue{Role: name, InSet: in,
				Provider: "not measured: no buff set the raid can make differs by it alone"})
			continue
		}
		if with != "" {
			name += " (with " + with + ")"
		}
		rv := &proto.CompRoleValue{Role: name, InSet: in, Measured: true}
		if in {
			rv.Net = roleBase - c.Score
			prov, other, cost, err := providerCost(best, g, sc, ev, f)
			if err != nil {
				return nil, err
			}
			rv.Provider, rv.ProviderCost = prov, cost
			if other != 0 {
				rv.Role += fmt.Sprintf(" (its giver also gives %s, and its cost counts in each)", other)
			}
			rv.Gain = rv.Net + cost
		} else {
			rv.Net = c.Score - roleBase
			rv.Provider = "value of adding"
		}
		out.roles = append(out.roles, rv)
	}
	// A shaman placeholder's air totem: the other one, re-filled.
	for i, t := range best.Plan.PHAir {
		if i >= 2 {
			break
		}
		q := best.Plan
		q.PHAir = append([]proto.RaidTotem{}, best.Plan.PHAir...)
		q.PHAir[i] = otherAir(t)
		q.Up = r.phUp(q)&^GiftBlessing | (best.Plan.Up &^ r.phUp(best.Plan))
		if q.Paladins > 0 {
			q.Up |= GiftBlessing
		}
		if c, err := sc.EvalExact(r, q, f, ev); err != nil {
			return nil, err
		} else if c != nil {
			net := roleBase - c.Score
			out.roles = append(out.roles, &proto.CompRoleValue{Role: fmt.Sprintf("a shaman placeholder's %s over %s", totemName(t), totemName(otherAir(t))),
				InSet: true, Measured: true, Gain: net, Net: net, Provider: "placeholders"})
		}
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
			out.roles = append(out.roles, &proto.CompRoleValue{Role: "a Ret paladin's blessings", InSet: true, Measured: true, Net: roleBase - c.Score, Provider: "Ret paladin"})
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
		// A replacement that beats the member beyond noise is a better comp the search missed, unless the
		// member is fixed (the search couldn't replace it).
		if mv.BestReplacement != "" && mv.ReplacementMargin < -2*best.SE && !m.Fixed {
			out.notes = append(out.notes, fmt.Sprintf("the search missed a better comp: %s in place of %s scores %.1f higher (beyond noise)",
				mv.BestReplacement, m.B.ID, -mv.ReplacementMargin))
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
			// A fixed member stays; a duty holder's seat takes the build's own variant on that duty too.
			if tried[m.B] || m.Fixed {
				continue
			}
			tried[m.B] = true
			try := []*Build{b}
			if v := b.Duties[m.B.Duty]; m.B.Duty != proto.Duty_DutyNone && v != nil {
				try = append(try, v)
			}
			for _, x := range try {
				v, ok, err := sc.swapScore(best, i, x, ev)
				if err != nil {
					return nil, err
				}
				if ok {
					bestV = math.Max(bestV, v)
				}
			}
		}
		if !math.IsInf(bestV, -1) {
			out.adding = append(out.adding, &proto.CompValueOfAdding{BuildId: b.ID, Value: bestV - best.Score})
		}
	}
	sort.SliceStable(out.adding, func(a, b int) bool { return out.adding[a].Value > out.adding[b].Value })
	return out, nil
}

// providerCost is the member giving a role, the other roles it gives in the set, and its cost: the best
// filler in its seat less its DPS.
func providerCost(c *Comp, g Gift, sc *Scenario, ev *exactValuer, f *Fillers) (string, Gift, float64, error) {
	for _, m := range c.Members {
		if m.B.Gives&g == 0 {
			continue
		}
		other := m.B.Gives &^ g &^ GiftBlessing & c.Plan.Up
		// A fixed member is never displaced, so the role costs it only its duty: its base build in the same
		// seat less it (nothing when it gives the role without one).
		if m.Fixed {
			if m.B.Duty == proto.Duty_DutyNone {
				return m.B.ID, other, 0, nil
			}
			s := m.Seat
			s.Own = m.B.Base.Casts
			ev.V(m.B.Base, s)
			if _, err := ev.flush(); err != nil {
				return "", 0, 0, err
			}
			v, _, _ := ev.V(m.B.Base, s)
			return m.B.ID, other, math.Max(0, v-m.DPS), nil
		}
		bestV := math.Inf(-1)
		// The first pass records the sims the cache lacks; the second reads them.
		for pass := 0; pass < 2; pass++ {
			for _, b := range f.Fill {
				if allowed(b, c.Plan) && b.Class != proto.Class_ClassShaman && (b.Class != proto.Class_ClassPaladin || c.Plan.Paladins >= 3) {
					s := c.seat(b, m.Seat.Package, sc)
					v, _, _ := ev.V(b, s)
					bestV = math.Max(bestV, v)
				}
			}
			if pass == 0 {
				if _, err := ev.flush(); err != nil {
					return "", 0, 0, err
				}
				bestV = math.Inf(-1)
			}
		}
		return m.B.ID, other, math.Max(0, bestV-m.DPS), nil
	}
	return "", 0, 0, nil
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
	for _, m := range members {
		if pullSunderer(m.B) {
			nc.Warriors++
		}
	}
	// Each party's package from the givers still in the raid: its shaman placeholder, or a DPS shaman.
	phPkg, dpsPkg := map[int][]proto.RaidPackage{}, map[int][]proto.RaidPackage{}
	for i, ph := range r.Combo {
		if ph.Shaman {
			phPkg[c.PHParty[i]] = append(phPkg[c.PHParty[i]], ph.Package)
		}
	}
	for _, m := range members {
		if m.B.Class == proto.Class_ClassShaman {
			dpsPkg[m.Party] = append(dpsPkg[m.Party], m.B.Pkg)
		}
	}
	pkg := map[int]proto.RaidPackage{}
	for party := 0; party < r.Parties; party++ {
		pkg[party] = partyPackage(phPkg[party], dpsPkg[party])
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
	reseat := func() {
		for _, m := range nc.Members {
			pi, inn := m.Seat.PI, m.Seat.Innervate
			m.Seat = nc.seat(m.B, m.Seat.Package, sc)
			m.Seat.PI, m.Seat.Innervate = pi, inn
		}
	}
	// A change in the paladin count re-solves each class's blessings: the choice its members gain most from.
	if p.Paladins != c.Plan.Paladins {
		choices := blessingChoices(p.Paladins)
		nc.Blessings = map[proto.Class][]proto.Blessing{}
		for k := range c.Blessings {
			nc.Blessings[k] = choices[0]
		}
		for _, m := range nc.Members {
			nc.Blessings[m.B.Class] = choices[0]
		}
		if len(choices) > 1 {
			for _, ch := range choices {
				for _, m := range nc.Members {
					s := m.Seat
					tmp := *nc
					tmp.Blessings = copyBlessings(nc.Blessings)
					tmp.Blessings[m.B.Class] = ch
					s2 := tmp.seat(m.B, s.Package, sc)
					s2.PI, s2.Innervate = s.PI, s.Innervate
					ev.V(m.B, s2)
				}
			}
			if _, err := ev.flush(); err != nil {
				return nil, err
			}
			next := copyBlessings(nc.Blessings)
			for cl := range nc.Blessings {
				bestV := math.Inf(-1)
				for _, ch := range choices {
					tmp := *nc
					tmp.Blessings = copyBlessings(nc.Blessings)
					tmp.Blessings[cl] = ch
					v, any := 0.0, false
					for _, m := range nc.Members {
						if m.B.Class != cl {
							continue
						}
						s2 := tmp.seat(m.B, m.Seat.Package, sc)
						s2.PI, s2.Innervate = m.Seat.PI, m.Seat.Innervate
						d, _, _ := ev.V(m.B, s2)
						v, any = v+d, true
					}
					if any && v > bestV+1e-9 {
						next[cl], bestV = ch, v
					}
				}
			}
			nc.Blessings = next
		}
		reseat()
	}
	// The placeholders' Power Infusions and Innervates a removed or swapped member held go to whoever
	// gains most from them now (the members keeping theirs keep them).
	for _, m := range nc.Members {
		ev.V(m.B, m.Seat)
	}
	if _, err := ev.flush(); err != nil {
		return nil, err
	}
	saved := make([]Seat, len(nc.Members))
	for i, m := range nc.Members {
		saved[i] = m.Seat
	}
	sc.assignCooldowns(nc, ev)
	if n, err := ev.flush(); err != nil {
		return nil, err
	} else if n > 0 {
		for i, m := range nc.Members {
			m.Seat = saved[i]
		}
		sc.assignCooldowns(nc, ev)
		if _, err := ev.flush(); err != nil {
			return nil, err
		}
	}
	for _, m := range nc.Members {
		m.DPS, m.SE, m.key = ev.V(m.B, m.Seat)
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
	// The best other build in that slot. A duty holder is also compared with every other build that can
	// hold its duty.
	alts := cands
	if m.B.Duty != proto.Duty_DutyNone {
		alts = append([]*Build{}, cands...)
		for _, b := range sc.Pool.Bases {
			if v := b.Duties[m.B.Duty]; v != nil && v != m.B {
				alts = append(alts, v)
			}
		}
	}
	bestV, bestB := math.Inf(-1), ""
	for _, b := range alts {
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
	// A duty's cost: the base build in the same seat, less the member on duty.
	if m.B.Duty != proto.Duty_DutyNone {
		s := m.Seat
		s.Own = m.B.Base.Casts
		ev.V(m.B.Base, s)
		if _, err := ev.flush(); err != nil {
			return nil, err
		}
		v, _, _ := ev.V(m.B.Base, s)
		out.DutyCost = v - m.DPS
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

// reportBaseline is the report build a member is compared with: the build it was picked as (a pet or
// encounter variant is its own report build), or for a duty holder its base build.
func reportBaseline(b *Build) *Build {
	if b.Duty != proto.Duty_DutyNone {
		return b.Base
	}
	return b
}

// reportDifference breaks a member's DPS less its report build's published DPS into single-player sims:
// the report's own request, at the faction's race, in this fight (boss, length, flask zone), under the
// raid's buff set, then on duty or as its variant.
func (sc *Scenario) reportDifference(m *Member, ev *exactValuer) (float64, []string, error) {
	base := reportBaseline(m.B)
	if base.Published == 0 {
		return 0, nil, nil
	}
	reportJob := func(race, fight bool) (*Job, error) {
		req := googleProto.Clone(base.Req).(*proto.RaidSimRequest)
		if race {
			pl, err := sc.Pool.Player(base, sc.Faction)
			if err != nil {
				return nil, err
			}
			req.Raid.Parties[0].Players[0] = pl
		}
		if fight {
			sc.Fight.apply(req)
		}
		req.SimOptions = &proto.SimOptions{Iterations: ev.iters, RandomSeed: sc.Seed, UseLabeledRands: sc.Sim.Labeled}
		j := &Job{Build: base, Faction: sc.Faction, Iterations: ev.iters, Seed: sc.Seed}
		j.req = req
		return j, nil
	}
	j0, err := reportJob(false, false)
	if err != nil {
		return 0, nil, err
	}
	j1, err := reportJob(true, false)
	if err != nil {
		return 0, nil, err
	}
	jf, err := reportJob(true, true)
	if err != nil {
		return 0, nil, err
	}
	seat := m.Seat
	seat.Own = base.Casts
	j2 := sc.job(base, seat, ev.iters)
	j3 := sc.job(m.B, m.Seat, ev.iters)
	if err := sc.sim([]*Job{j0, j1, jf, j2, j3}); err != nil {
		return 0, nil, err
	}
	parts := []string{
		fmt.Sprintf("the report's request in this sim: %+.1f", j0.Value.DPS-base.Published),
		fmt.Sprintf("race (%s): %+.1f", strings.TrimPrefix(base.Race[sc.Faction].String(), "Race"), j1.Value.DPS-j0.Value.DPS),
		fmt.Sprintf("this fight (boss, length, flask zone): %+.1f", jf.Value.DPS-j1.Value.DPS),
		fmt.Sprintf("the raid's buff set: %+.1f", j2.Value.DPS-jf.Value.DPS),
	}
	// Only a duty holder is compared with another build (its base); a pet or encounter variant is its own.
	if m.B != base {
		label := "its duty (" + dutyLabel(m.B.Duty) + ")"
		// A duty build's own race table can pick another race, which this step then includes.
		if r := m.B.Race[sc.Faction]; r != base.Race[sc.Faction] {
			label += ", at its own best race (" + strings.TrimPrefix(r.String(), "Race") + ")"
		}
		parts = append(parts, fmt.Sprintf("%s: %+.1f", label, j3.Value.DPS-j2.Value.DPS))
	}
	return m.DPS - base.Published, parts, nil
}

func totemName(t proto.RaidTotem) string {
	switch t {
	case proto.RaidTotem_RaidTotemGraceOfAir:
		return "Grace of Air"
	case proto.RaidTotem_RaidTotemWindfury:
		return "Windfury Totem"
	}
	return t.String()
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
// base is each comp's DPS before the shifts and its standard error: the raid sims' for confirmed comps
// (then only the confirmed take part, since the winner is the raid sims' choice), else the score's.
func (sc *Scenario) assumptionSensitivity(contenders []*Comp, best *Comp, base func(*Comp) (float64, float64, bool)) []*proto.CompSensitivity {
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
	label := assumptionLabels(sc.Pool.Root)
	var out []*proto.CompSensitivity
	for _, name := range names {
		var vs []shifted
		for _, c := range contenders {
			v, _, ok := base(c)
			if !ok {
				continue
			}
			for _, m := range c.Members {
				v += shift[m.B.Base.ID][name]
			}
			vs = append(vs, shifted{c, v})
		}
		sort.SliceStable(vs, func(a, b int) bool { return vs[a].v > vs[b].v })
		bv, bse, ok := base(best)
		if len(vs) == 0 || !ok {
			continue
		}
		res := &proto.CompSensitivity{Scenario: label(name), WinnerHolds: vs[0].c == best ||
			vs[0].v-valueOf(vs, best) <= math.Max(2*bse, 0.001*bv)}
		if !res.WinnerHolds {
			// The whole comp, so a page can name its members; its DPS and gap are the shifted ones.
			w := sc.compProto(vs[0].c, best, nil)
			w.Dps, w.Gap = vs[0].v, vs[0].v-valueOf(vs, best)
			w.BuffSet = vs[0].c.Describe() + "; " + comboText(comboIDs(vs[0].c.Raid.Combo))
			res.Winners = append(res.Winners, w)
		}
		out = append(out, res)
	}
	return out
}

// assumptionLabels names an assumption scenario ("raid_hits=high|natural_flask_zone=low") from the
// report's assumptions.json: each question and the setting it takes. A key it lacks stays as it is.
func assumptionLabels(root string) func(string) string {
	var defs []struct {
		Key      string `json:"key"`
		Label    string `json:"label"`
		Settings []struct {
			ID    string `json:"id"`
			Label string `json:"label"`
		} `json:"settings"`
	}
	if data, err := os.ReadFile(filepath.Join(root, "ui", "dps_report", "data", "assumptions.json")); err == nil {
		_ = json.Unmarshal(data, &defs)
	}
	return func(name string) string {
		var parts []string
		for _, kv := range strings.Split(name, "|") {
			k, v, _ := strings.Cut(kv, "=")
			part := kv
			for _, d := range defs {
				if d.Key != k {
					continue
				}
				part = strings.TrimSuffix(d.Label, "?") + ": " + v
				for _, st := range d.Settings {
					if st.ID == v {
						part = strings.TrimSuffix(d.Label, "?") + ": " + st.Label
					}
				}
			}
			parts = append(parts, part)
		}
		return strings.Join(parts, "; ")
	}
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
