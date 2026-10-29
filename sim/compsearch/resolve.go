package compsearch

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

// The resolver (docs/raid_sim.md, M1b): a comp of DPS members and tank and healer placeholders into one
// raid's request. Placeholders aren't units; what they give is written into the members' settings. Each
// member gets its own raid buffs (Player.raid_buffs), so the class hooks are skipped and the resolver owns
// every buff: given for the long buffs, the held auras and everything a placeholder gives, cast for
// what a member's rotation keeps up itself (its totems, its duty), which then reaches its scope in the
// engine. Never both.

type member struct {
	slot   *proto.CompSlot
	preset *Preset
	player *proto.Player // the player it runs: its duty player when it holds a duty
	base   *proto.Player
	duties map[proto.Duty]*proto.Player
	duty   proto.Duty
	gives  Gift
	casts  Gift
	pkg    proto.RaidPackage
	class  proto.Class
	bless  map[proto.Class]proto.Blessing // a paladin's blessing for each class
	judge  proto.PaladinJudgement
	totems map[proto.RaidTotem]bool // a shaman placeholder's totems
	pi     bool
	inn    bool
}

func (m *member) name() string {
	if m.preset != nil {
		return m.preset.Name
	}
	if m.slot.Name != "" {
		return m.slot.Name
	}
	if m.slot.ReportId != "" {
		return m.slot.ReportId
	}
	return fmt.Sprintf("slot %d", m.slot.Id)
}

func isMelee(m *member) bool {
	switch m.class {
	case proto.Class_ClassWarrior, proto.Class_ClassRogue, proto.Class_ClassPaladin:
		return true
	case proto.Class_ClassShaman:
		return m.player.GetEnhancementShaman() != nil
	case proto.Class_ClassDruid:
		return m.player.GetFeralDruid() != nil
	}
	return false
}

// blessingOrder is what a class wants first from its paladins.
func blessingOrder(c proto.Class) []proto.Blessing {
	switch c {
	case proto.Class_ClassWarrior, proto.Class_ClassRogue:
		return []proto.Blessing{proto.Blessing_BlessingKings, proto.Blessing_BlessingMight, proto.Blessing_BlessingWisdom}
	case proto.Class_ClassPaladin, proto.Class_ClassShaman, proto.Class_ClassDruid:
		return []proto.Blessing{proto.Blessing_BlessingKings, proto.Blessing_BlessingMight, proto.Blessing_BlessingWisdom}
	}
	return []proto.Blessing{proto.Blessing_BlessingKings, proto.Blessing_BlessingWisdom, proto.Blessing_BlessingMight}
}

// Resolve resolves a comp (docs/raid_sim.md, Assignments): every assignment set by hand is kept, and
// every other job is given automatically, to a placeholder first since that costs no DPS.
func Resolve(comp *proto.RaidComp) *proto.RaidCompResolution {
	res, err := resolve(comp)
	if err != nil {
		return &proto.RaidCompResolution{Error: &proto.ErrorOutcome{Message: err.Error()}}
	}
	return res
}

func resolve(comp *proto.RaidComp) (*proto.RaidCompResolution, error) {
	size := int(comp.Size)
	if size <= 0 || size > 40 {
		return nil, fmt.Errorf("a raid has 1 to 40 members, not %d", size)
	}
	parties := (size + 4) / 5
	if len(comp.Slots) > size {
		return nil, fmt.Errorf("%d slots in a raid of %d", len(comp.Slots), size)
	}
	taken := map[[2]int32]bool{}
	ids := map[int32]*member{}
	var ms []*member
	for _, s := range comp.Slots {
		if s.Party < 0 || int(s.Party) >= parties || s.Index < 0 || s.Index >= 5 {
			return nil, fmt.Errorf("slot %d is at party %d, index %d, outside the raid", s.Id, s.Party+1, s.Index+1)
		}
		if taken[[2]int32{s.Party, s.Index}] {
			return nil, fmt.Errorf("two slots at party %d, index %d", s.Party+1, s.Index+1)
		}
		taken[[2]int32{s.Party, s.Index}] = true
		if _, dup := ids[s.Id]; dup {
			return nil, fmt.Errorf("two slots with id %d", s.Id)
		}
		m := &member{slot: s, duties: map[proto.Duty]*proto.Player{}, bless: map[proto.Class]proto.Blessing{}, totems: map[proto.RaidTotem]bool{}}
		switch x := s.Member.(type) {
		case *proto.CompSlot_Placeholder:
			p := PresetByID(x.Placeholder)
			if p == nil {
				return nil, fmt.Errorf("slot %d: unknown preset %v", s.Id, x.Placeholder)
			}
			m.preset = p
			m.class = p.Class
			m.gives = p.Gives
			if p.Paladin {
				m.gives |= GiftBlessing
			}
			if p.Shaman {
				m.gives |= GiftSoE | GiftFT
				m.pkg = p.Package
			}
			m.inn = p.Innervate || (p.InnervateOptional && s.PresetInnervate)
		case *proto.CompSlot_Build:
			if x.Build == nil || x.Build.Base == nil {
				return nil, fmt.Errorf("slot %d has no player", s.Id)
			}
			m.base = x.Build.Base
			m.player = m.base
			m.class = m.base.Class
			for _, d := range x.Build.Duties {
				m.duties[d.Duty] = d.Player
			}
			m.gives, m.pkg, m.casts = PlayerGifts(m.base, proto.Duty_DutyNone)
		default:
			return nil, fmt.Errorf("slot %d is empty", s.Id)
		}
		ids[s.Id] = m
		ms = append(ms, m)
	}
	sort.SliceStable(ms, func(i, j int) bool {
		if ms[i].slot.Party != ms[j].slot.Party {
			return ms[i].slot.Party < ms[j].slot.Party
		}
		return ms[i].slot.Index < ms[j].slot.Index
	})
	var dps, phs, paladins, phShamans []*member
	for _, m := range ms {
		if m.preset != nil {
			phs = append(phs, m)
			if m.preset.Paladin {
				paladins = append(paladins, m)
			}
			if m.preset.Shaman {
				phShamans = append(phShamans, m)
			}
		} else {
			dps = append(dps, m)
			if m.class == proto.Class_ClassPaladin {
				paladins = append(paladins, m)
			}
		}
	}
	paladins = phPaladinsFirst(paladins)

	// Assignments set by hand first.
	var out []*proto.Assignment
	handDuty := map[proto.Duty]bool{}
	handJudge := map[int32]bool{}
	handTotem := map[int32]map[proto.RaidTotem]bool{}
	handBless := map[int32]map[proto.Class]bool{}
	handCD := map[bool]bool{}
	for _, a := range comp.Assignments {
		m, ok := ids[a.Holder]
		if !ok {
			return nil, fmt.Errorf("an assignment's holder %d isn't in the raid", a.Holder)
		}
		switch j := a.Job.(type) {
		case *proto.Assignment_Duty:
			if m.preset != nil {
				return nil, fmt.Errorf("%s can't hold a duty: placeholders give theirs without one", m.name())
			}
			p, ok := m.duties[j.Duty]
			if !ok {
				return nil, fmt.Errorf("%s has no %s build", m.name(), dutyLabel(j.Duty))
			}
			if m.duty != proto.Duty_DutyNone {
				return nil, fmt.Errorf("%s holds two duties", m.name())
			}
			m.duty, m.player = j.Duty, p
			handDuty[j.Duty] = true
		case *proto.Assignment_Judgement:
			if m.class != proto.Class_ClassPaladin {
				return nil, fmt.Errorf("%s isn't a paladin", m.name())
			}
			m.judge = j.Judgement
			handJudge[a.Holder] = true
		case *proto.Assignment_Blessing:
			if m.class != proto.Class_ClassPaladin {
				return nil, fmt.Errorf("%s isn't a paladin", m.name())
			}
			m.bless[j.Blessing.Class] = j.Blessing.Blessing
			if handBless[a.Holder] == nil {
				handBless[a.Holder] = map[proto.Class]bool{}
			}
			handBless[a.Holder][j.Blessing.Class] = true
		case *proto.Assignment_Totem:
			if m.preset == nil || !m.preset.Shaman {
				return nil, fmt.Errorf("%s isn't a shaman placeholder (a DPS shaman drops its build's totems)", m.name())
			}
			if handTotem[a.Holder] == nil {
				handTotem[a.Holder] = map[proto.RaidTotem]bool{}
			}
			m.totems[j.Totem.Totem] = true
			handTotem[a.Holder][j.Totem.Totem] = true
		case *proto.Assignment_Cooldown:
			t, ok := ids[j.Cooldown.Target]
			if !ok || t.preset != nil {
				return nil, fmt.Errorf("a cooldown's target %d isn't a DPS member", j.Cooldown.Target)
			}
			if j.Cooldown.PowerInfusion {
				t.pi = true
			} else {
				t.inn = true
			}
			handCD[j.Cooldown.PowerInfusion] = true
		}
		out = append(out, googleProto.Clone(a).(*proto.Assignment))
	}

	// What the placeholders give before choices.
	phUp := Gift(0)
	for _, m := range phs {
		phUp |= m.gives
	}
	// Duties no placeholder covers go to the first DPS member with that duty's build.
	for _, d := range []proto.Duty{proto.Duty_DutyBattleShout, proto.Duty_DutySunderUpkeep, proto.Duty_DutyFaerieFire,
		proto.Duty_DutyCurseOfTheElements, proto.Duty_DutyJudgementOfTheCrusader} {
		g := DutyGift(d)
		if handDuty[d] || phUp.Has(g) || (g == GiftCrusader && len(paladinPHs(paladins)) > 0) {
			continue
		}
		if d == proto.Duty_DutyFaerieFire && handDuty[proto.Duty_DutyCurseOfRecklessness] {
			continue
		}
		for _, m := range dps {
			if p, ok := m.duties[d]; ok && m.duty == proto.Duty_DutyNone {
				m.duty, m.player = d, p
				out = append(out, &proto.Assignment{Holder: m.slot.Id, Job: &proto.Assignment_Duty{Duty: d}})
				break
			}
		}
	}
	for _, m := range dps {
		if m.duty != proto.Duty_DutyNone {
			m.gives, m.pkg, m.casts = PlayerGifts(m.player, m.duty)
		}
	}

	// Judgements: the paladin placeholders judge Wisdom, the second one the Crusader, unless a Ret
	// member holds the Crusader duty.
	crusader := false
	for _, m := range dps {
		if m.duty == proto.Duty_DutyJudgementOfTheCrusader {
			crusader = true
		}
	}
	for _, m := range paladinPHs(paladins) {
		if m.judge == proto.PaladinJudgement_PaladinJudgementCrusader {
			crusader = true
		}
	}
	nth := 0
	for _, m := range paladinPHs(paladins) {
		if handJudge[m.slot.Id] {
			continue
		}
		m.judge = proto.PaladinJudgement_PaladinJudgementWisdom
		if nth == 1 && !crusader {
			m.judge = proto.PaladinJudgement_PaladinJudgementCrusader
			crusader = true
		}
		nth++
		out = append(out, &proto.Assignment{Holder: m.slot.Id, Job: &proto.Assignment_Judgement{Judgement: m.judge}})
	}

	// Blessings: each paladin gives one Greater Blessing to each class, in the order the class wants them.
	classes := map[proto.Class]bool{}
	for _, m := range dps {
		classes[m.class] = true
	}
	for _, cl := range sortedClasses(classes) {
		given := map[proto.Blessing]bool{}
		for _, p := range paladins {
			if b, ok := p.bless[cl]; ok {
				given[b] = true
			}
		}
		order := blessingOrder(cl)
		for _, p := range paladins {
			if handBless[p.slot.Id][cl] {
				continue
			}
			for _, b := range order {
				if !given[b] {
					given[b] = true
					p.bless[cl] = b
					out = append(out, &proto.Assignment{Holder: p.slot.Id, Job: &proto.Assignment_Blessing{Blessing: &proto.BlessingAssignment{Class: cl, Blessing: b}}})
					break
				}
			}
		}
	}

	// Totems: a shaman placeholder drops Strength of Earth, Mana Spring and Flametongue, and Windfury for
	// a raid of melee or Grace of Air for the rest; a second one takes the other air totem.
	melee := 0
	for _, m := range dps {
		if isMelee(m) {
			melee++
		}
	}
	air := proto.RaidTotem_RaidTotemGraceOfAir
	if 2*melee > len(dps) {
		air = proto.RaidTotem_RaidTotemWindfury
	}
	for _, m := range phShamans {
		if handTotem[m.slot.Id] != nil {
			continue
		}
		for _, t := range []proto.RaidTotem{proto.RaidTotem_RaidTotemStrengthOfEarth, proto.RaidTotem_RaidTotemManaSpring, proto.RaidTotem_RaidTotemFlametongue, air} {
			m.totems[t] = true
			out = append(out, &proto.Assignment{Holder: m.slot.Id, Job: &proto.Assignment_Totem{Totem: &proto.TotemAssignment{Totem: t}}})
		}
		if air == proto.RaidTotem_RaidTotemGraceOfAir {
			air = proto.RaidTotem_RaidTotemWindfury
		} else {
			air = proto.RaidTotem_RaidTotemGraceOfAir
		}
	}

	// Power Infusion and Innervate: the placeholders' go to the first casters and mana users.
	nPI, nInn := 0, 0
	for _, m := range phs {
		if m.preset.PI {
			nPI++
		}
		if m.inn {
			nInn++
		}
	}
	for _, cd := range []struct {
		pi bool
		n  int
	}{{true, nPI}, {false, nInn}} {
		if handCD[cd.pi] {
			continue
		}
		left := cd.n
		for _, m := range dps {
			if left == 0 {
				break
			}
			if !usesManaPlayer(m.player) || (cd.pi && isMelee(m)) || (cd.pi && m.pi) || (!cd.pi && m.inn) {
				continue
			}
			if cd.pi {
				m.pi = true
			} else {
				m.inn = true
			}
			left--
			out = append(out, &proto.Assignment{Holder: m.slot.Id, Job: &proto.Assignment_Cooldown{Cooldown: &proto.CooldownTarget{PowerInfusion: cd.pi, Target: m.slot.Id}}})
		}
	}

	// What's up: the placeholders' gifts and choices, and every member's.
	up := Gift(0)
	for _, m := range phs {
		up |= m.gives &^ GiftBlessing
		switch m.judge {
		case proto.PaladinJudgement_PaladinJudgementCrusader:
			up |= GiftCrusader
		case proto.PaladinJudgement_PaladinJudgementWisdom:
			up |= GiftJoW
		}
		if m.totems[proto.RaidTotem_RaidTotemGraceOfAir] {
			up |= GiftGoA
		}
		if m.totems[proto.RaidTotem_RaidTotemWindfury] {
			up |= GiftWF
		}
		if !m.totems[proto.RaidTotem_RaidTotemStrengthOfEarth] && m.preset.Shaman && handTotem[m.slot.Id] != nil {
			up &^= GiftSoE
		}
		if !m.totems[proto.RaidTotem_RaidTotemFlametongue] && m.preset.Shaman && handTotem[m.slot.Id] != nil {
			up &^= GiftFT
		}
	}
	cast := Gift(0) // effects some member casts in the fight, so nobody gets them given
	for _, m := range dps {
		up |= m.gives &^ GiftBlessing
		cast |= m.casts
	}

	// The raid's given debuffs: what the placeholders keep up. Members' duties cast theirs.
	debuffs := &proto.Debuffs{}
	if phUp.Has(GiftSunderTank) {
		debuffs.SunderRamp = &proto.SunderRamp{Tank: true}
	}
	debuffs.FaerieFire = up.Has(GiftMinorArmor) && !cast.Has(GiftMinorArmor)
	debuffs.CurseOfElements = up.Has(GiftCoE) && !cast.Has(GiftCoE)
	if up.Has(GiftCrusader) && !cast.Has(GiftCrusader) {
		debuffs.JudgementOfTheCrusader = proto.TristateEffect_TristateEffectRegular
	}
	debuffs.JudgementOfWisdom = up.Has(GiftJoW) && !cast.Has(GiftJoW)

	// Each party's package from its shaman placeholder (a DPS shaman's Mana Spring is cast).
	partyPkg := make([]proto.RaidPackage, parties)
	partyTide := make([]int32, parties)
	for _, m := range phShamans {
		if m.totems[proto.RaidTotem_RaidTotemManaSpring] {
			p := m.slot.Party
			if m.pkg > partyPkg[p] {
				partyPkg[p] = m.pkg
			}
			if m.pkg == proto.RaidPackage_RaidPackageRestoration {
				partyTide[p]++
			}
		}
	}

	raid := &proto.Raid{Buffs: &proto.RaidBuffs{}, Debuffs: debuffs, NumActiveParties: int32(parties)}
	for i := 0; i < parties; i++ {
		raid.Parties = append(raid.Parties, &proto.Party{Buffs: &proto.PartyBuffs{ManaTideTotems: partyTide[i]}})
	}
	res := &proto.RaidCompResolution{}
	blessings := map[proto.Class][]proto.Blessing{}
	for cl := range classes {
		for _, p := range paladins {
			if b, ok := p.bless[cl]; ok {
				blessings[cl] = append(blessings[cl], b)
			}
		}
	}
	for _, m := range dps {
		seat := Seat{Up: up &^ cast, Blessings: blessings[m.class], Package: partyPkg[m.slot.Party], PI: m.pi,
			Innervate: m.inn, Camps: comp.CampBuffs, Own: 0, Assumptions: comp.Assumptions}
		tmp := &proto.RaidSimRequest{Raid: &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{{}}}}}}
		seat.Write(tmp, &Build{Class: m.class, Spec: specOf(m.player)})
		p := googleProto.Clone(m.player).(*proto.Player)
		p.RaidBuffs = tmp.Raid.Buffs
		p.Buffs = tmp.Raid.Parties[0].Players[0].Buffs
		if m.slot.Name != "" {
			p.Name = m.slot.Name
		} else if p.Name == "" {
			p.Name = m.name()
		}
		raid.Parties[m.slot.Party].Players = append(raid.Parties[m.slot.Party].Players, p)
	}
	// slot_of_player follows the request's order: parties, then players.
	for pi := range raid.Parties {
		for _, m := range dps {
			if int(m.slot.Party) == pi {
				res.SlotOfPlayer = append(res.SlotOfPlayer, m.slot.Id)
			}
		}
	}
	enc := comp.Encounter
	if enc == nil {
		enc = &proto.Encounter{}
	}
	res.Request = &proto.RaidSimRequest{Raid: raid, Encounter: enc, SimOptions: comp.SimOptions}
	if res.Request.SimOptions == nil {
		res.Request.SimOptions = &proto.SimOptions{Iterations: 1000, RandomSeed: 1}
	}
	res.Assignments = out
	res.Buffs, res.Misses = resolutionRecords(up, cast, phs, dps, blessings, partyPkg)
	return res, nil
}

func phPaladinsFirst(pals []*member) []*member {
	sort.SliceStable(pals, func(i, j int) bool { return pals[i].preset != nil && pals[j].preset == nil })
	return pals
}

func paladinPHs(pals []*member) []*member {
	var out []*member
	for _, p := range pals {
		if p.preset != nil {
			out = append(out, p)
		}
	}
	return out
}

func sortedClasses(m map[proto.Class]bool) []proto.Class {
	var out []proto.Class
	for c := range m {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func usesManaPlayer(p *proto.Player) bool {
	switch p.Class {
	case proto.Class_ClassWarrior, proto.Class_ClassRogue:
		return false
	case proto.Class_ClassDruid:
		return p.GetFeralDruid() == nil
	}
	return true
}

func specOf(p *proto.Player) string {
	switch {
	case p.GetEnhancementShaman() != nil:
		return "Enhancement"
	case p.GetFeralDruid() != nil:
		return "Feral"
	}
	return ""
}

func dutyLabel(d proto.Duty) string {
	for k, v := range dutyNames {
		if v == d {
			return k
		}
	}
	return d.String()
}

// resolutionRecords is each effect up with who gives it, and what the raid goes without.
func resolutionRecords(up, cast Gift, phs, dps []*member, blessings map[proto.Class][]proto.Blessing, partyPkg []proto.RaidPackage) ([]*proto.CompBuffRecord, []string) {
	var recs []*proto.CompBuffRecord
	for i := 0; i < numGifts; i++ {
		g := Gift(1) << i
		if !up.Has(g) || g == GiftBlessing {
			continue
		}
		var who []string
		for _, m := range phs {
			if m.gives.Has(g) || (g == GiftCrusader && m.judge == proto.PaladinJudgement_PaladinJudgementCrusader) ||
				(g == GiftJoW && m.judge == proto.PaladinJudgement_PaladinJudgementWisdom) ||
				(g == GiftGoA && m.totems[proto.RaidTotem_RaidTotemGraceOfAir]) || (g == GiftWF && m.totems[proto.RaidTotem_RaidTotemWindfury]) {
				who = append(who, m.name())
			}
		}
		for _, m := range dps {
			if m.gives.Has(g) {
				who = append(who, m.name())
			}
		}
		scope := "raid"
		switch g {
		case GiftSunderTank, GiftMinorArmor, GiftCoE, GiftCrusader, GiftJoW:
			scope = "the boss"
		}
		recs = append(recs, &proto.CompBuffRecord{Effect: GiftName(g), Provider: strings.Join(uniq(who), ", "), Scope: scope,
			Receivers: int32(len(dps)), Given: !cast.Has(g)})
	}
	for p, pk := range partyPkg {
		if pk == proto.RaidPackage_RaidPackageNone {
			continue
		}
		n := 0
		for _, m := range dps {
			if int(m.slot.Party) == p {
				n++
			}
		}
		recs = append(recs, &proto.CompBuffRecord{Effect: packageName(pk), Provider: fmt.Sprintf("party %d's shaman", p+1), Scope: "party", Receivers: int32(n), Given: true})
	}
	var misses []string
	for _, g := range modelRoles {
		if !up.Has(g) {
			misses = append(misses, "nobody gives "+GiftName(g))
		}
	}
	for _, cl := range sortedClasses(func() map[proto.Class]bool {
		m := map[proto.Class]bool{}
		for _, x := range dps {
			m[x.class] = true
		}
		return m
	}()) {
		if n := len(blessings[cl]); n < 3 {
			misses = append(misses, fmt.Sprintf("%s get %d of the 3 blessings", strings.TrimPrefix(cl.String(), "Class"), n))
		}
	}
	return recs, misses
}

// RunComp resolves a comp and runs its raid sim.
func RunComp(comp *proto.RaidComp) *proto.RaidCompRunResult {
	res := Resolve(comp)
	out := &proto.RaidCompRunResult{Resolution: res}
	if res.Error != nil {
		return out
	}
	out.Result = core.RunRaidSimConcurrent(res.Request)
	return out
}

// compOf is a search comp as a RaidComp, for its confirmation: every member at the faction's race,
// every choice set by hand to the search's.
func (sc *Scenario) compOf(c *Comp, iterations int32, seed int64) (*proto.RaidComp, error) {
	rc := &proto.RaidComp{Size: int32(c.Raid.Size), Faction: sc.Faction, CampBuffs: sc.Camps, Assumptions: sc.Assumptions,
		SimOptions: &proto.SimOptions{Iterations: iterations, RandomSeed: seed}}
	enc := cloneReq(sc.Pool.Bases[0].Req).Encounter
	enc.Duration = sc.Fight.Duration
	enc.DurationVariation = sc.Fight.DurationVariation
	enc.NaturalFlaskZone = sc.Fight.FlaskZone
	if sc.Fight.MobType != proto.MobType_MobTypeUnknown {
		for _, t := range enc.Targets {
			t.MobType = sc.Fight.MobType
		}
	}
	rc.Encounter = enc
	next := make([]int32, c.Raid.Parties)
	id := int32(0)
	place := func(party int) (int32, int32, error) {
		if party < 0 || party >= len(next) || next[party] >= 5 {
			return 0, 0, fmt.Errorf("party %d is full", party+1)
		}
		next[party]++
		id++
		return int32(party), next[party] - 1, nil
	}
	var pals []int32
	for i, ph := range c.Raid.Combo {
		party, idx, err := place(c.PHParty[i])
		if err != nil {
			return nil, err
		}
		rc.Slots = append(rc.Slots, &proto.CompSlot{Id: id, Party: party, Index: idx, Role: ph.Role, Member: &proto.CompSlot_Placeholder{Placeholder: ph.ID}})
		if ph.Paladin {
			pals = append(pals, id)
			j := proto.PaladinJudgement_PaladinJudgementWisdom
			if paladinIndex(c.Raid, i) < c.Plan.PHCrusader {
				j = proto.PaladinJudgement_PaladinJudgementCrusader
			}
			rc.Assignments = append(rc.Assignments, &proto.Assignment{Holder: id, SetByHand: true, Job: &proto.Assignment_Judgement{Judgement: j}})
		}
		if ph.Shaman {
			totems := []proto.RaidTotem{proto.RaidTotem_RaidTotemStrengthOfEarth, proto.RaidTotem_RaidTotemManaSpring, proto.RaidTotem_RaidTotemFlametongue}
			if k := shamanIndex(c.Raid, i); k < len(c.Plan.PHAir) {
				totems = append(totems, c.Plan.PHAir[k])
			}
			for _, t := range totems {
				rc.Assignments = append(rc.Assignments, &proto.Assignment{Holder: id, SetByHand: true, Job: &proto.Assignment_Totem{Totem: &proto.TotemAssignment{Totem: t}}})
			}
		}
	}
	memberID := map[*Member]int32{}
	for _, m := range c.Members {
		party, idx, err := place(m.Party)
		if err != nil {
			return nil, err
		}
		base, err := sc.Pool.Player(m.B.Base, sc.Faction)
		if err != nil {
			return nil, err
		}
		mb := &proto.CompMemberBuild{Base: base}
		if m.B.Duty != proto.Duty_DutyNone {
			dp, err := sc.Pool.Player(m.B, sc.Faction)
			if err != nil {
				return nil, err
			}
			mb.Duties = append(mb.Duties, &proto.CompDutyPlayer{Duty: m.B.Duty, Player: dp})
			rc.Assignments = append(rc.Assignments, &proto.Assignment{Holder: id, SetByHand: true, Job: &proto.Assignment_Duty{Duty: m.B.Duty}})
		}
		if m.B.Class == proto.Class_ClassPaladin {
			pals = append(pals, id)
		}
		rc.Slots = append(rc.Slots, &proto.CompSlot{Id: id, Party: party, Index: idx, Role: proto.RaidRole_RaidRoleDps,
			Member: &proto.CompSlot_Build{Build: mb}, ReportId: m.B.Base.ID, Name: fmt.Sprintf("%s %d", m.B.Base.ID, id)})
		memberID[m] = id
		if m.Seat.PI {
			rc.Assignments = append(rc.Assignments, &proto.Assignment{Holder: id, SetByHand: true, Job: &proto.Assignment_Cooldown{Cooldown: &proto.CooldownTarget{PowerInfusion: true, Target: id}}})
		}
		if m.Seat.Innervate {
			rc.Assignments = append(rc.Assignments, &proto.Assignment{Holder: id, SetByHand: true, Job: &proto.Assignment_Cooldown{Cooldown: &proto.CooldownTarget{Target: id}}})
		}
	}
	// The blessings each class got, one per paladin.
	for cl, bl := range c.Blessings {
		for k, b := range bl {
			if k < len(pals) {
				rc.Assignments = append(rc.Assignments, &proto.Assignment{Holder: pals[k], SetByHand: true,
					Job: &proto.Assignment_Blessing{Blessing: &proto.BlessingAssignment{Class: cl, Blessing: b}}})
			}
		}
	}
	sort.SliceStable(rc.Assignments, func(i, j int) bool { return rc.Assignments[i].Holder < rc.Assignments[j].Holder })
	return rc, nil
}

// confirm raid-sims a comp: its raid DPS and standard error.
func (sc *Scenario) confirm(c *Comp, iterations int32, seed int64) (float64, float64, error) {
	rc, err := sc.compOf(c, iterations, seed)
	if err != nil {
		return 0, 0, err
	}
	run := RunComp(rc)
	if run.Resolution.Error != nil {
		return 0, 0, fmt.Errorf("%s", run.Resolution.Error.Message)
	}
	if run.Result.Error != nil {
		return 0, 0, fmt.Errorf("%s", run.Result.Error.Message)
	}
	m := run.Result.RaidMetrics.Dps
	return m.Avg, m.Stdev / math.Sqrt(float64(iterations)), nil
}
