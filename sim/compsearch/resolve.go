package compsearch

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
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
				m.pkg = p.Package // its totems come from its choices below
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
				if d.GetPlayer() == nil {
					return nil, fmt.Errorf("slot %d's %s build has no player", s.Id, dutyLabel(d.Duty))
				}
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
		// An assignment not set by hand is the resolver's to make again.
		if !a.SetByHand {
			continue
		}
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
			if m.preset == nil {
				// A Ret member judges through its duty build.
				d := proto.Duty_DutyJudgementOfWisdom
				if j.Judgement == proto.PaladinJudgement_PaladinJudgementCrusader {
					d = proto.Duty_DutyJudgementOfTheCrusader
				}
				p, ok := m.duties[d]
				if !ok {
					return nil, fmt.Errorf("%s has no %s build", m.name(), dutyLabel(d))
				}
				if m.duty != proto.Duty_DutyNone && m.duty != d {
					return nil, fmt.Errorf("%s holds two duties", m.name())
				}
				m.duty, m.player = d, p
				handDuty[d] = true
				break
			}
			m.judge = j.Judgement
			handJudge[a.Holder] = true
		case *proto.Assignment_Blessing:
			if j.Blessing == nil {
				return nil, fmt.Errorf("a blessing assignment for %s names no blessing", m.name())
			}
			if m.class != proto.Class_ClassPaladin {
				return nil, fmt.Errorf("%s isn't a paladin", m.name())
			}
			m.bless[j.Blessing.Class] = j.Blessing.Blessing
			if handBless[a.Holder] == nil {
				handBless[a.Holder] = map[proto.Class]bool{}
			}
			handBless[a.Holder][j.Blessing.Class] = true
		case *proto.Assignment_Totem:
			if j.Totem == nil {
				return nil, fmt.Errorf("a totem assignment for %s names no totem", m.name())
			}
			if m.preset == nil || !m.preset.Shaman {
				return nil, fmt.Errorf("%s isn't a shaman placeholder (a DPS shaman drops its build's totems)", m.name())
			}
			if handTotem[a.Holder] == nil {
				handTotem[a.Holder] = map[proto.RaidTotem]bool{}
			}
			m.totems[j.Totem.Totem] = true
			handTotem[a.Holder][j.Totem.Totem] = true
		case *proto.Assignment_Cooldown:
			if j.Cooldown == nil {
				return nil, fmt.Errorf("a cooldown assignment for %s names no target", m.name())
			}
			if m.preset == nil || (j.Cooldown.PowerInfusion && !m.preset.PI) || (!j.Cooldown.PowerInfusion && !m.inn) {
				return nil, fmt.Errorf("%s has no %s to give", m.name(), map[bool]string{true: "Power Infusion", false: "Innervate"}[j.Cooldown.PowerInfusion])
			}
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
	// What the DPS members give without a duty, and with the ones set by hand.
	dpsUp := Gift(0)
	for _, m := range dps {
		if m.duty != proto.Duty_DutyNone {
			m.gives, m.pkg, m.casts = PlayerGifts(m.player, m.duty)
		}
		dpsUp |= m.gives
	}
	// Duties nobody covers go to the first DPS member with that duty's build, the biggest first.
	for _, d := range []proto.Duty{proto.Duty_DutySunderUpkeep, proto.Duty_DutyBattleShout, proto.Duty_DutyFaerieFire,
		proto.Duty_DutyCurseOfTheElements, proto.Duty_DutyJudgementOfTheCrusader} {
		g := DutyGift(d)
		if handDuty[d] || (phUp | dpsUp).Has(g) || (g == GiftCrusader && len(paladinPHs(paladins)) > 0) {
			continue
		}
		if d == proto.Duty_DutyFaerieFire && handDuty[proto.Duty_DutyCurseOfRecklessness] {
			continue
		}
		for _, m := range dps {
			if p, ok := m.duties[d]; ok && m.duty == proto.Duty_DutyNone {
				m.duty, m.player = d, p
				m.gives, m.pkg, m.casts = PlayerGifts(m.player, m.duty)
				dpsUp |= m.gives
				out = append(out, &proto.Assignment{Holder: m.slot.Id, Job: &proto.Assignment_Duty{Duty: d}})
				break
			}
		}
	}
	// Judgements: each paladin placeholder on auto takes the judgement nobody keeps up yet, Wisdom first.
	crusader, wisdom := dpsUp.Has(GiftCrusader), dpsUp.Has(GiftJoW)
	for _, m := range paladinPHs(paladins) {
		switch m.judge {
		case proto.PaladinJudgement_PaladinJudgementCrusader:
			crusader = true
		case proto.PaladinJudgement_PaladinJudgementWisdom:
			wisdom = true
		}
	}
	for _, m := range paladinPHs(paladins) {
		if handJudge[m.slot.Id] {
			continue
		}
		switch {
		case !wisdom:
			m.judge, wisdom = proto.PaladinJudgement_PaladinJudgementWisdom, true
		case !crusader:
			m.judge, crusader = proto.PaladinJudgement_PaladinJudgementCrusader, true
		default:
			m.judge = proto.PaladinJudgement_PaladinJudgementWisdom
		}
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
		order := blessingOrderFor(cl, dps)
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
	if 2*melee > len(dps) && !dpsUp.Has(GiftWF) {
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
	var piGivers, innGivers []int32
	for _, m := range phs {
		if m.preset.PI {
			piGivers = append(piGivers, m.slot.Id)
		}
		if m.inn {
			innGivers = append(innGivers, m.slot.Id)
		}
	}
	for _, cd := range []struct {
		pi     bool
		givers []int32
	}{{true, piGivers}, {false, innGivers}} {
		if handCD[cd.pi] {
			continue
		}
		left := len(cd.givers)
		for _, m := range dps {
			if left == 0 {
				break
			}
			if !usesManaPlayer(m.player) || (cd.pi && !isCaster(m)) || (cd.pi && m.pi) || (!cd.pi && m.inn) {
				continue
			}
			if cd.pi {
				m.pi = true
			} else {
				m.inn = true
			}
			left--
			out = append(out, &proto.Assignment{Holder: cd.givers[left], Job: &proto.Assignment_Cooldown{Cooldown: &proto.CooldownTarget{PowerInfusion: cd.pi, Target: m.slot.Id}}})
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
		for t, g := range map[proto.RaidTotem]Gift{proto.RaidTotem_RaidTotemStrengthOfEarth: GiftSoE, proto.RaidTotem_RaidTotemGraceOfAir: GiftGoA,
			proto.RaidTotem_RaidTotemWindfury: GiftWF, proto.RaidTotem_RaidTotemFlametongue: GiftFT} {
			if m.totems[t] {
				up |= g
				m.gives |= g // for the records
			}
		}
	}
	cast := Gift(0) // effects some member casts in the fight, so nobody gets them given
	for _, m := range dps {
		up |= m.gives &^ GiftBlessing
		cast |= m.casts
	}

	// The raid's given debuffs: what the placeholders keep up. Members' duties cast theirs.
	debuffs := &proto.Debuffs{}
	if phUp.Has(GiftSunderTank) && !cast.Has(GiftSunderTank) {
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
		pkg := partyPkg[m.slot.Party]
		if m.class == proto.Class_ClassShaman {
			pkg = proto.RaidPackage_RaidPackageNone // it casts its own Mana Spring: never given and cast
		}
		seat := Seat{Up: up &^ cast, Blessings: blessings[m.class], Package: pkg, PI: m.pi,
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

// isCaster is a member Power Infusion's 20% magic damage is for.
func isCaster(m *member) bool {
	switch m.class {
	case proto.Class_ClassMage, proto.Class_ClassWarlock, proto.Class_ClassPriest:
		return true
	case proto.Class_ClassDruid:
		return m.player.GetBalanceDruid() != nil
	case proto.Class_ClassShaman:
		return m.player.GetElementalShaman() != nil
	}
	return false
}

// blessingOrderFor is what a class's members want first: Kings, then Might when most of them are melee,
// else Wisdom.
func blessingOrderFor(c proto.Class, dps []*member) []proto.Blessing {
	melee, n := 0, 0
	for _, m := range dps {
		if m.class == c {
			n++
			if isMelee(m) {
				melee++
			}
		}
	}
	if 2*melee > n {
		return []proto.Blessing{proto.Blessing_BlessingKings, proto.Blessing_BlessingMight, proto.Blessing_BlessingWisdom}
	}
	return []proto.Blessing{proto.Blessing_BlessingKings, proto.Blessing_BlessingWisdom, proto.Blessing_BlessingMight}
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
	return runComp(comp, nil)
}

// runComp is RunComp that stops once abort says so.
func runComp(comp *proto.RaidComp, abort func() bool) *proto.RaidCompRunResult {
	res := Resolve(comp)
	out := &proto.RaidCompRunResult{Resolution: res}
	if res.Error != nil {
		return out
	}
	signals := simsignals.CreateSignals()
	done := make(chan struct{})
	if abort != nil {
		go func() {
			t := time.NewTicker(200 * time.Millisecond)
			defer t.Stop()
			for {
				select {
				case <-done:
					return
				case <-t.C:
					if abort() {
						signals.Abort.Trigger()
						return
					}
				}
			}
		}()
	}
	out.Result = core.RunRaidSimConcurrentWithSignals(res.Request, signals)
	close(done)
	return out
}

// compOf is a search comp as a RaidComp, for its confirmation: every member at the faction's race,
// every choice set by hand to the search's.
func (sc *Scenario) compOf(c *Comp, iterations int32, seed int64) (*proto.RaidComp, error) {
	rc := &proto.RaidComp{Size: int32(c.Raid.Size), Faction: sc.Faction, CampBuffs: sc.Camps, Assumptions: sc.Assumptions,
		SimOptions: &proto.SimOptions{Iterations: iterations, RandomSeed: seed, UseLabeledRands: sc.Sim.Labeled}}
	enc := cloneReq(sc.Pool.Bases[0].Req).Encounter
	enc.Duration = sc.Fight.Duration
	if sc.HealthLength > 0 {
		// A health fight's confirmation runs the real fight: the boss dies at its health.
		enc.Duration = sc.HealthLength
		enc.UseHealth = true
		for _, t := range enc.Targets {
			for len(t.Stats) <= int(proto.Stat_StatHealth) {
				t.Stats = append(t.Stats, 0)
			}
			t.Stats[proto.Stat_StatHealth] = sc.BossHealth
		}
	}
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
	var pals, piGivers, innGivers []int32
	for i, ph := range c.Raid.Combo {
		party, idx, err := place(c.PHParty[i])
		if err != nil {
			return nil, err
		}
		rc.Slots = append(rc.Slots, &proto.CompSlot{Id: id, Party: party, Index: idx, Role: ph.Role, Member: &proto.CompSlot_Placeholder{Placeholder: ph.ID}})
		if ph.PI {
			piGivers = append(piGivers, id)
		}
		if ph.Innervate {
			innGivers = append(innGivers, id)
		}
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
		// The build it was scored as: a pet or encounter variant is its own build; a duty holder's base is
		// the build it was a variant of, against this boss.
		baseBuild := m.B
		if m.B.Duty != proto.Duty_DutyNone {
			baseBuild = m.B.Base.ForBoss(sc.Fight.MobType)
		}
		base, err := sc.Pool.Player(baseBuild, sc.Faction)
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
			Member: &proto.CompSlot_Build{Build: mb}, ReportId: baseBuild.ID, Name: fmt.Sprintf("%s %d", baseBuild.ID, id)})
		memberID[m] = id
		if m.Seat.PI && len(piGivers) > 0 {
			rc.Assignments = append(rc.Assignments, &proto.Assignment{Holder: piGivers[0], SetByHand: true, Job: &proto.Assignment_Cooldown{Cooldown: &proto.CooldownTarget{PowerInfusion: true, Target: id}}})
			piGivers = piGivers[1:]
		}
		if m.Seat.Innervate && len(innGivers) > 0 {
			rc.Assignments = append(rc.Assignments, &proto.Assignment{Holder: innGivers[0], SetByHand: true, Job: &proto.Assignment_Cooldown{Cooldown: &proto.CooldownTarget{Target: id}}})
			innGivers = innGivers[1:]
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
	run := runComp(rc, sc.Abort)
	if sc.Abort != nil && sc.Abort() {
		return 0, 0, ErrAborted
	}
	if run.Resolution.Error != nil {
		return 0, 0, fmt.Errorf("%s", run.Resolution.Error.Message)
	}
	if run.Result.Error != nil {
		return 0, 0, fmt.Errorf("%s", run.Result.Error.Message)
	}
	m := run.Result.RaidMetrics.Dps
	return m.Avg, m.Stdev / math.Sqrt(float64(iterations)), nil
}
