package compsearch

import (
	"context"
	"fmt"
	"sort"

	"github.com/wowsims/classic/sim/core/proto"
)

// CompValues is a hand-built raid's member values (docs/raid_sim.md, Member value: "for a raid built by
// hand, a button runs the sims the cache lacks"), for the native server's /compValues. Its members must
// be report builds (each slot's report_id).
func CompValues(comp *proto.RaidComp) *proto.CompValuesResult {
	return CompValuesCtx(context.Background(), comp)
}

// CompValuesCtx is CompValues that stops when the context is done (the page stopped waiting).
func CompValuesCtx(ctx context.Context, comp *proto.RaidComp) *proto.CompValuesResult {
	pool, sim, err := serverSearch()
	if err == nil {
		var out *proto.CompValuesResult
		out, err = compValues(pool, sim, comp, func() bool { return ctx.Err() != nil })
		if err == nil {
			return out
		}
	}
	return &proto.CompValuesResult{Error: &proto.ErrorOutcome{Message: err.Error()}}
}

// compFromRaidComp is the search's view of a hand-built comp: its placeholders, its members as report
// builds, and the buff set they make, with the assignments the resolver settled.
func compFromRaidComp(pool *Pool, sc *Scenario, comp *proto.RaidComp) (*Comp, map[*Member]int32, error) {
	res := Resolve(comp)
	if res.Error != nil {
		return nil, nil, fmt.Errorf("%s", res.Error.Message)
	}
	slots := map[int32]*proto.CompSlot{}
	for _, s := range comp.Slots {
		slots[s.Id] = s
	}
	duty := map[int32]proto.Duty{}
	judge := map[int32]proto.PaladinJudgement{}
	air := map[int32]proto.RaidTotem{}
	spring := map[int32]bool{} // the shaman placeholders that drop Mana Spring
	pi, inn := map[int32]bool{}, map[int32]bool{}
	bless := map[proto.Class][]proto.Blessing{}
	for _, a := range res.Assignments {
		switch j := a.Job.(type) {
		case *proto.Assignment_Duty:
			duty[a.Holder] = j.Duty
		case *proto.Assignment_Judgement:
			judge[a.Holder] = j.Judgement
		case *proto.Assignment_Totem:
			if t := j.Totem.Totem; t == proto.RaidTotem_RaidTotemGraceOfAir || t == proto.RaidTotem_RaidTotemWindfury {
				air[a.Holder] = t
			} else if t == proto.RaidTotem_RaidTotemManaSpring {
				spring[a.Holder] = true
			}
		case *proto.Assignment_Cooldown:
			if j.Cooldown.PowerInfusion {
				pi[j.Cooldown.Target] = true
			} else {
				inn[j.Cooldown.Target] = true
			}
		case *proto.Assignment_Blessing:
			bless[j.Blessing.Class] = append(bless[j.Blessing.Class], j.Blessing.Blessing)
		}
	}
	ordered := append([]*proto.CompSlot{}, comp.Slots...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Id < ordered[j].Id })
	var combo []*Preset
	var phParty []int
	var phIDs []int32
	for _, s := range ordered {
		if x, ok := s.Member.(*proto.CompSlot_Placeholder); ok {
			combo = append(combo, PresetByID(x.Placeholder))
			phParty = append(phParty, int(s.Party))
			phIDs = append(phIDs, s.Id)
		}
	}
	r := NewRaid(int(comp.Size), combo, nil)
	p := Plan{}
	for i, ph := range combo {
		if ph.Paladin && judge[phIDs[i]] == proto.PaladinJudgement_PaladinJudgementCrusader {
			p.PHCrusader++
		}
	}
	// PHAir follows the raid's shaman placeholders, richest package first.
	for _, sh := range r.phShamans {
		for i, ph := range combo {
			if ph == sh {
				if t, ok := air[phIDs[i]]; ok {
					p.PHAir = append(p.PHAir, t)
					delete(air, phIDs[i])
					break
				}
			}
		}
	}
	c := &Comp{Raid: r, PHParty: phParty, Blessings: bless}
	ids := map[*Member]int32{}
	pal := r.phPaladins
	up := Gift(0)
	phPkg, dpsPkg := map[int32][]proto.RaidPackage{}, map[int32][]proto.RaidPackage{}
	for i, ph := range combo {
		if ph.Shaman && spring[phIDs[i]] {
			phPkg[int32(phParty[i])] = append(phPkg[int32(phParty[i])], ph.Package)
		}
	}
	for _, s := range ordered {
		if _, ok := s.Member.(*proto.CompSlot_Build); !ok {
			continue
		}
		b, ok := pool.Builds[s.ReportId]
		if !ok {
			return nil, nil, fmt.Errorf("slot %d isn't a report build (%q), so it has no single-player sims", s.Id, s.ReportId)
		}
		// A base build runs its encounter variant against this boss; a pet or encounter variant is its own.
		if b.IsBase() {
			b = b.ForBoss(sc.Fight.MobType)
		}
		if d := duty[s.Id]; d != proto.Duty_DutyNone {
			v := b.Base.WithDuty(d)
			if v == nil {
				return nil, nil, fmt.Errorf("slot %d (%s) has no %s build", s.Id, b.ID, dutyLabel(d))
			}
			b = v
		}
		m := &Member{B: b, Party: int(s.Party)}
		m.Seat.PI, m.Seat.Innervate = pi[s.Id], inn[s.Id]
		c.Members = append(c.Members, m)
		ids[m] = s.Id
		up |= b.Gives &^ GiftBlessing
		if b.Class == proto.Class_ClassPaladin {
			pal++
		}
		if pullSunderer(b) {
			c.Warriors++
		}
		if b.Class == proto.Class_ClassShaman {
			dpsPkg[s.Party] = append(dpsPkg[s.Party], b.Pkg)
		}
	}
	// Each party's package as the resolver gives it (a DPS shaman's Mana Spring reaches the rest of its
	// party, in place of a given one).
	partyPkg := map[int32]proto.RaidPackage{}
	for party := int32(0); party < int32(r.Parties); party++ {
		partyPkg[party] = partyPackage(phPkg[party], dpsPkg[party])
	}
	p.Paladins = min(3, pal)
	p.Up = r.phUp(p)&^GiftBlessing | up
	if p.Paladins > 0 {
		p.Up |= GiftBlessing
	}
	c.Plan = p
	for _, m := range c.Members {
		pkg := partyPkg[int32(m.Party)]
		if m.B.Class == proto.Class_ClassShaman {
			pkg = proto.RaidPackage_RaidPackageNone // a shaman casts its own Mana Spring, so it's never given one too
		}
		pi, in := m.Seat.PI, m.Seat.Innervate
		m.Seat = c.seat(m.B, pkg, sc)
		m.Seat.PI, m.Seat.Innervate = pi, in
	}
	return c, ids, nil
}

func compValues(pool *Pool, sim *Simmer, comp *proto.RaidComp, abort func() bool) (*proto.CompValuesResult, error) {
	sc := &Scenario{Pool: pool, Sim: sim, Faction: comp.Faction, Camps: comp.CampBuffs, Assumptions: comp.Assumptions,
		Iterations: 1000, Seed: 101, Fight: Fight{Duration: 180, DurationVariation: 20, FlaskZone: true}, Abort: abort}
	if sc.Faction == proto.Faction_Unknown {
		sc.Faction = proto.Faction_Horde
	}
	// The search's screening iterations at most: a raid run's 3,000 or more would take many minutes over
	// the dozens of sims a member value needs. Fewer if the comp asks for fewer.
	if o := comp.SimOptions; o != nil && o.Iterations > 0 && o.Iterations < sc.Iterations {
		sc.Iterations = o.Iterations
	}
	if o := comp.SimOptions; o != nil && o.RandomSeed != 0 {
		sc.Seed = o.RandomSeed
	}
	if e := comp.Encounter; e != nil {
		sc.Fight.Duration, sc.Fight.DurationVariation, sc.Fight.FlaskZone = e.Duration, e.DurationVariation, e.NaturalFlaskZone
		if len(e.Targets) > 0 {
			sc.Fight.MobType = e.Targets[0].MobType
		}
		if sc.Fight.Duration == 0 {
			sc.Fight.Duration = 180
		}
	}
	c, ids, err := compFromRaidComp(pool, sc, comp)
	if err != nil {
		return nil, err
	}
	var cands []*Build
	for _, b := range pool.Bases {
		x := b.ForBoss(sc.Fight.MobType)
		cands = append(cands, x)
		cands = append(cands, b.Pets...)
	}
	// No model: member values only read exact sims, each after a flush, so a model's guesses before them
	// would cost thousands of sims for nothing.
	ev := newExactValuer(sc, sc.Iterations, nil)
	scored, err := sc.withMembers(c, c.Members, ev)
	if err != nil {
		return nil, err
	}
	out := &proto.CompValuesResult{Score: scored.Score, Se: scored.SE}
	byKey := map[string]*proto.CompMemberValue{}
	f := &Fillers{Fill: cands, Providers: cands}
	for i, m := range scored.Members {
		k := askKey(m.B, m.Seat)
		if mv, ok := byKey[k]; ok {
			mv.Slots = append(mv.Slots, ids[c.Members[i]])
			continue
		}
		mv, err := sc.memberValue(scored, i, f, ev, cands)
		if err != nil {
			return nil, err
		}
		mv.Slots = []int32{ids[c.Members[i]]}
		byKey[k] = mv
		out.MemberValues = append(out.MemberValues, mv)
	}
	return out, nil
}
