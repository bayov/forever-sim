package compsearch

import (
	"fmt"
	"sort"

	"github.com/wowsims/classic/sim/core/proto"
)

// CompValues is a hand-built raid's member values (docs/raid_sim.md, Member value: "for a raid built by
// hand, a button runs the sims the cache lacks"), for the native server's /compValues. Its members must
// be report builds (each slot's report_id).
func CompValues(comp *proto.RaidComp) *proto.CompValuesResult {
	pool, sim, err := serverSearch()
	if err == nil {
		var out *proto.CompValuesResult
		out, err = compValues(pool, sim, comp)
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
					air[phIDs[i]] = proto.RaidTotem_RaidTotemNone
					break
				}
			}
		}
	}
	c := &Comp{Raid: r, PHParty: phParty, Blessings: bless}
	ids := map[*Member]int32{}
	pal := r.phPaladins
	up := Gift(0)
	partyPkg := map[int32]proto.RaidPackage{}
	for i, ph := range combo {
		if ph.Shaman && ph.Package > partyPkg[int32(phParty[i])] {
			partyPkg[int32(phParty[i])] = ph.Package
		}
	}
	var giverShamans []*Member
	for _, s := range ordered {
		if _, ok := s.Member.(*proto.CompSlot_Build); !ok {
			continue
		}
		b, ok := pool.Builds[s.ReportId]
		if !ok {
			return nil, nil, fmt.Errorf("slot %d isn't a report build (%q), so it has no single-player sims", s.Id, s.ReportId)
		}
		b = b.Base.ForBoss(sc.Fight.MobType)
		if d := duty[s.Id]; d != proto.Duty_DutyNone {
			if v := b.Base.WithDuty(d); v != nil {
				b = v
			}
		}
		m := &Member{B: b, Party: int(s.Party)}
		m.Seat.PI, m.Seat.Innervate = pi[s.Id], inn[s.Id]
		c.Members = append(c.Members, m)
		ids[m] = s.Id
		up |= b.Gives &^ GiftBlessing
		if b.Class == proto.Class_ClassPaladin {
			pal++
		}
		if b.Class == proto.Class_ClassWarrior {
			c.Warriors++
		}
		if b.Class == proto.Class_ClassShaman && partyPkg[s.Party] == proto.RaidPackage_RaidPackageNone {
			giverShamans = append(giverShamans, m)
		}
	}
	// A DPS shaman's Mana Spring reaches the rest of its party.
	for _, m := range giverShamans {
		if partyPkg[int32(m.Party)] < m.B.Pkg {
			partyPkg[int32(m.Party)] = m.B.Pkg
		}
	}
	p.Paladins = min(3, pal)
	p.Up = r.phUp(p)&^GiftBlessing | up
	if p.Paladins > 0 {
		p.Up |= GiftBlessing
	}
	c.Plan = p
	for _, m := range c.Members {
		pkg := partyPkg[int32(m.Party)]
		if m.B.Class == proto.Class_ClassShaman && pkg == m.B.Pkg {
			pkg = proto.RaidPackage_RaidPackageNone // its own Mana Spring is its own cast
		}
		pi, in := m.Seat.PI, m.Seat.Innervate
		m.Seat = c.seat(m.B, pkg, sc)
		m.Seat.PI, m.Seat.Innervate = pi, in
	}
	return c, ids, nil
}

func compValues(pool *Pool, sim *Simmer, comp *proto.RaidComp) (*proto.CompValuesResult, error) {
	sc := &Scenario{Pool: pool, Sim: sim, Faction: comp.Faction, Camps: comp.CampBuffs, Assumptions: comp.Assumptions,
		Iterations: 1000, Seed: 101, Fight: Fight{Duration: 180, DurationVariation: 20, FlaskZone: true}}
	if sc.Faction == proto.Faction_Unknown {
		sc.Faction = proto.Faction_Horde
	}
	if o := comp.SimOptions; o != nil && o.Iterations > 0 {
		sc.Iterations = o.Iterations
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
	var builds []*Build
	seen := map[*Build]bool{}
	var cands []*Build
	for _, b := range pool.Bases {
		x := b.ForBoss(sc.Fight.MobType)
		cands = append(cands, x)
		cands = append(cands, b.Pets...)
	}
	for _, b := range cands {
		if !seen[b] {
			seen[b] = true
			builds = append(builds, b)
		}
	}
	for _, m := range c.Members {
		if !seen[m.B] {
			seen[m.B] = true
			builds = append(builds, m.B)
		}
	}
	model, err := sc.BuildModel(builds)
	if err != nil {
		return nil, err
	}
	ev := newExactValuer(sc, sc.Iterations, model)
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
		if m.B.Duty != proto.Duty_DutyNone {
			s := m.Seat
			s.Own = m.B.Base.Casts
			v, _, _ := ev.V(m.B.Base, s)
			if _, err := ev.flush(); err != nil {
				return nil, err
			}
			v, _, _ = ev.V(m.B.Base, s)
			mv.DutyCost = v - m.DPS
		}
		byKey[k] = mv
		out.MemberValues = append(out.MemberValues, mv)
	}
	return out, nil
}
