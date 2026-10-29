package compsearch

import (
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"

	"github.com/wowsims/classic/sim/core/proto"
)

// Scenario is one run's fixed inputs: a faction, a camp setting and a fight.
type Scenario struct {
	Pool        *Pool
	Sim         *Simmer
	Faction     proto.Faction
	Camps       bool
	Fight       Fight
	Iterations  int32
	Seed        int64
	Assumptions *proto.RaidAssumptions
	// A health fight's length this pass (0 for a fixed fight): the boss's health over the raid's DPS.
	HealthLength float64
	Abort        func() bool
}

// sim runs jobs, unless the search was aborted.
func (sc *Scenario) sim(jobs []*Job) error {
	if sc.Abort != nil && sc.Abort() {
		return ErrAborted
	}
	return sc.Sim.SimAbortable(jobs, sc.Abort)
}

func (sc *Scenario) job(b *Build, s Seat, iters int32) *Job {
	return &Job{Build: b, Faction: sc.Faction, Seat: s, Fight: sc.Fight, Iterations: iters, Seed: sc.Seed}
}

// Raid is a raid's size and placeholders: one tank and healer combo, and any fixed DPS members.
type Raid struct {
	Size    int
	Combo   []*Preset
	Fixed   []*Build
	Parties int
	DPS     int // DPS slots, fixed members included

	phGifts    Gift
	phPaladins int
	phShamans  []*Preset // the shaman placeholders, richest package first
	phPI       int
	phInn      int
}

// NewRaid is a raid with a combo.
func NewRaid(size int, combo []*Preset, fixed []*Build) *Raid {
	r := &Raid{Size: size, Combo: combo, Fixed: fixed, Parties: (size + 4) / 5, DPS: size - len(combo)}
	for _, p := range combo {
		r.phGifts |= p.Gives
		if p.Paladin {
			r.phPaladins++
		}
		if p.Shaman {
			r.phShamans = append(r.phShamans, p)
			r.phGifts |= GiftSoE | GiftFT
		}
		if p.PI {
			r.phPI++
		}
		if p.Innervate {
			r.phInn++
		}
	}
	sort.SliceStable(r.phShamans, func(i, j int) bool { return r.phShamans[i].Package > r.phShamans[j].Package })
	return r
}

// Plan is a buff set (docs/raid_sim.md, Why it doesn't try combinations): what's up, and the
// placeholders' choices. The provider members follow from it in the fill.
type Plan struct {
	Up         Gift
	Paladins   int // paladins giving blessings, placeholders and Ret members, at most 3
	PHCrusader int // paladin placeholders judging the Crusader; the others judge Wisdom
	PHAir      []proto.RaidTotem
}

func (p Plan) Key() string {
	return fmt.Sprintf("%x/%d/%d/%v", uint32(p.Up), p.Paladins, p.PHCrusader, p.PHAir)
}

func (p Plan) String() string {
	s := p.Up.String()
	if p.Paladins > 0 {
		s += fmt.Sprintf("; %d blessings per class", p.Paladins)
	}
	return s
}

// phUp is what the placeholders give under a plan.
func (r *Raid) phUp(p Plan) Gift {
	g := r.phGifts
	if r.phPaladins > 0 {
		g |= GiftBlessing
		if p.PHCrusader > 0 {
			g |= GiftCrusader
		}
		if r.phPaladins-p.PHCrusader > 0 {
			g |= GiftJoW
		}
	}
	for _, t := range p.PHAir {
		switch t {
		case proto.RaidTotem_RaidTotemGraceOfAir:
			g |= GiftGoA
		case proto.RaidTotem_RaidTotemWindfury:
			g |= GiftWF
		}
	}
	return g
}

// Valuer gives a build's DPS in a seat.
type Valuer interface {
	V(b *Build, s Seat) (dps, se float64, key string)
}

// Member is one DPS member of a filled comp.
type Member struct {
	B        *Build // the build it runs: a duty variant when it holds a duty
	Party    int
	Seat     Seat
	DPS      float64
	SE       float64
	key      string
	Provider bool
	Fixed    bool
}

// Comp is a filled buff set.
type Comp struct {
	Raid      *Raid
	Plan      Plan
	Members   []*Member
	PHParty   []int // each placeholder's party
	Blessings map[proto.Class][]proto.Blessing
	Warriors  int
	Score     float64
	SE        float64
}

func (c *Comp) seat(b *Build, pkg proto.RaidPackage, sc *Scenario) Seat {
	pull := c.Warriors
	if b.Class == proto.Class_ClassWarrior {
		pull--
	}
	return Seat{Up: c.Plan.Up, Blessings: c.Blessings[b.Class], PullSunders: max(0, pull), Package: pkg,
		Camps: sc.Camps, Own: b.Casts, Assumptions: sc.Assumptions}
}

// item is a provider candidate: a build and the roles it covers.
type item struct {
	b     *Build
	cover Gift
	ret   bool // a paladin: one more blessing per class
}

// Fillers is the search's view of the builds: the fillers that survived pruning, and the providers.
type Fillers struct {
	Fill      []*Build // base builds that may fill a slot
	Providers []*Build // builds and duty variants that may provide a role
}

func allowed(b *Build, p Plan) bool {
	if (b.Gives&^GiftBlessing)&^p.Up != 0 {
		return false
	}
	return true
}

// Fill fills a plan exactly under the valuer (docs/raid_sim.md, The search, step 4). It returns nil when
// the plan can't be made with this raid.
func (sc *Scenario) Fill(r *Raid, p Plan, f *Fillers, val Valuer) *Comp {
	phUp := r.phUp(p)
	if phUp&^GiftBlessing&^p.Up != 0 {
		return nil
	}
	fixedUp := Gift(0)
	fixedRet := 0
	for _, b := range r.Fixed {
		if !allowed(b, p) {
			return nil
		}
		fixedUp |= b.Gives
		if b.Class == proto.Class_ClassPaladin {
			fixedRet++
		}
	}
	retHave := r.phPaladins + fixedRet
	if p.Paladins < 3 && retHave > p.Paladins {
		return nil
	}
	if p.Paladins > 0 && !p.Up.Has(GiftBlessing) {
		return nil
	}
	retNeed := max(0, p.Paladins-retHave)
	need := (p.Up &^ GiftBlessing &^ GiftFort) &^ phUp &^ fixedUp
	slots := r.DPS - len(r.Fixed)
	if slots < 0 {
		return nil
	}

	// The blessings each class gets. With fewer than 3 paladins a class gets the ones its builds gain most
	// from, solved by turns (each class's choice is its own).
	nb := min(3, p.Paladins)
	choices := blessingChoices(nb)
	c := &Comp{Raid: r, Plan: p, Blessings: map[proto.Class][]proto.Blessing{}}
	classes := []proto.Class{proto.Class_ClassDruid, proto.Class_ClassHunter, proto.Class_ClassMage, proto.Class_ClassPaladin,
		proto.Class_ClassPriest, proto.Class_ClassRogue, proto.Class_ClassShaman, proto.Class_ClassWarlock, proto.Class_ClassWarrior}
	for _, cl := range classes {
		c.Blessings[cl] = choices[0]
	}
	if len(choices) > 1 {
		for _, cl := range classes {
			best, bestV := choices[0], math.Inf(-1)
			for _, ch := range choices {
				c.Blessings[cl] = ch
				v := math.Inf(-1)
				for _, b := range f.Fill {
					if b.Class == cl && allowed(b, p) {
						d, _, _ := val.V(b, c.seat(b, 0, sc))
						v = math.Max(v, d)
					}
				}
				if v > bestV {
					best, bestV = ch, v
				}
			}
			c.Blessings[cl] = best
		}
	}

	var best *Comp
	for round := 0; round < 3; round++ {
		comp := sc.fillOnce(r, p, f, val, c.Blessings, need, retNeed, slots)
		if comp == nil {
			return nil
		}
		// Each class takes the blessings its members gain most from, summed.
		changed := false
		if len(choices) > 1 {
			for _, cl := range classes {
				var ms []*Member
				for _, m := range comp.Members {
					if m.B.Class == cl {
						ms = append(ms, m)
					}
				}
				if len(ms) == 0 {
					continue
				}
				bestCh, bestV := comp.Blessings[cl], math.Inf(-1)
				for _, ch := range choices {
					tmp := *comp
					tmp.Blessings = copyBlessings(comp.Blessings)
					tmp.Blessings[cl] = ch
					v := 0.0
					for _, m := range ms {
						d, _, _ := val.V(m.B, tmp.seat(m.B, m.Seat.Package, sc))
						v += d
					}
					if v > bestV+1e-9 {
						bestCh, bestV = ch, v
					}
				}
				if !slices.Equal(bestCh, comp.Blessings[cl]) {
					comp.Blessings[cl] = bestCh
					changed = true
				}
			}
		}
		if best == nil || comp.Score > best.Score {
			best = comp
		}
		if !changed {
			break
		}
		c.Blessings = comp.Blessings
	}
	return best
}

func copyBlessings(m map[proto.Class][]proto.Blessing) map[proto.Class][]proto.Blessing {
	out := make(map[proto.Class][]proto.Blessing, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func blessingChoices(n int) [][]proto.Blessing {
	switch n {
	case 0:
		return [][]proto.Blessing{nil}
	case 1:
		return [][]proto.Blessing{{proto.Blessing_BlessingKings}, {proto.Blessing_BlessingMight}, {proto.Blessing_BlessingWisdom}}
	case 2:
		return [][]proto.Blessing{
			{proto.Blessing_BlessingKings, proto.Blessing_BlessingMight},
			{proto.Blessing_BlessingKings, proto.Blessing_BlessingWisdom},
			{proto.Blessing_BlessingMight, proto.Blessing_BlessingWisdom}}
	}
	return [][]proto.Blessing{allBlessings}
}

// fillOnce fills a plan with fixed blessings, under each warrior count the fill can come to.
func (sc *Scenario) fillOnce(r *Raid, p Plan, f *Fillers, val Valuer, bless map[proto.Class][]proto.Blessing,
	need Gift, retNeed, slots int) *Comp {
	var best *Comp
	warriors := -1
	for pass := 0; pass < 2; pass++ {
		comp := sc.fillWithWarriors(r, p, f, val, bless, need, retNeed, slots, max(0, warriors))
		if comp == nil {
			return best
		}
		// Score it at the warrior count it came out with.
		w := 0
		for _, m := range comp.Members {
			if m.B.Class == proto.Class_ClassWarrior {
				w++
			}
		}
		if w != comp.Warriors {
			comp.Warriors = w
			sc.rescore(comp, val)
		}
		if best == nil || comp.Score > best.Score {
			best = comp
		}
		if warriors == w {
			break
		}
		warriors = w
	}
	return best
}

// rescore recomputes each member's seat and DPS after a change to the comp.
func (sc *Scenario) rescore(c *Comp, val Valuer) {
	for _, m := range c.Members {
		pi, inn := m.Seat.PI, m.Seat.Innervate
		m.Seat = c.seat(m.B, m.Seat.Package, sc)
		m.Seat.PI, m.Seat.Innervate = pi, inn
		m.DPS, m.SE, m.key = val.V(m.B, m.Seat)
	}
	c.sum()
}

// sum is the comp's score: the members' DPS, and its standard error. Copies of one build in one seat are
// one sim, so their errors add, not their variances.
func (c *Comp) sum() {
	c.Score = 0
	byKey := map[string]float64{}
	var ses []float64
	for _, m := range c.Members {
		c.Score += m.DPS
		if m.key == "" {
			ses = append(ses, m.SE)
			continue
		}
		byKey[m.key] += m.SE
	}
	v := 0.0
	for _, s := range byKey {
		v += s * s
	}
	for _, s := range ses {
		v += s * s
	}
	c.SE = math.Sqrt(v)
}

func (sc *Scenario) fillWithWarriors(r *Raid, p Plan, f *Fillers, val Valuer, bless map[proto.Class][]proto.Blessing,
	need Gift, retNeed, slots, warriors int) *Comp {
	base := &Comp{Raid: r, Plan: p, Blessings: bless, Warriors: warriors}
	v0 := func(b *Build) float64 {
		d, _, _ := val.V(b, base.seat(b, 0, sc))
		return d
	}
	// The best filler in a seat without a package.
	retFill := p.Paladins >= 3
	var fill0 *Build
	fill0V := math.Inf(-1)
	for _, b := range f.Fill {
		if !allowed(b, p) || (b.Class == proto.Class_ClassPaladin && !retFill) {
			continue
		}
		if v := v0(b); v > fill0V {
			fill0, fill0V = b, v
		}
	}
	if fill0 == nil && slots > 0 {
		return nil
	}

	// Cover the roles the placeholders don't with the cheapest providers (docs/raid_sim.md, step 4).
	var items []item
	for _, b := range f.Providers {
		if !allowed(b, p) {
			continue
		}
		cov := b.Gives & need
		isRet := b.Class == proto.Class_ClassPaladin
		if cov == 0 && !(isRet && retNeed > 0) {
			continue
		}
		items = append(items, item{b: b, cover: cov, ret: isRet})
	}
	// With exact values, sim only the providers the model makes cheapest for each role.
	if ev, ok := val.(*exactValuer); ok && ev.model != nil {
		items = pruneItems(items, func(b *Build) float64 {
			return ev.model.V(b, base.seat(b, 0, sc))
		})
	}
	cost := map[*Build]float64{}
	for _, it := range items {
		if _, ok := cost[it.b]; !ok {
			cost[it.b] = fill0V - v0(it.b)
		}
	}
	type state struct {
		mask Gift
		ret  int
	}
	memo := map[state]struct {
		c   float64
		sel []*Build
	}{}
	var solve func(mask Gift, ret int) (float64, []*Build)
	solve = func(mask Gift, ret int) (float64, []*Build) {
		if mask == 0 && ret <= 0 {
			return 0, nil
		}
		st := state{mask, ret}
		if m, ok := memo[st]; ok {
			return m.c, m.sel
		}
		bestC, bestSel := math.Inf(1), []*Build(nil)
		low := mask & -mask
		for _, it := range items {
			if low != 0 && it.cover&low == 0 {
				continue
			}
			if low == 0 && !it.ret {
				continue
			}
			nr := ret
			if it.ret {
				if ret <= 0 && !retFill {
					continue
				}
				nr--
			}
			c, sel := solve(mask&^it.cover, nr)
			if c += cost[it.b]; c < bestC && len(sel) < slots {
				bestC, bestSel = c, append([]*Build{it.b}, sel...)
			}
		}
		memo[st] = struct {
			c   float64
			sel []*Build
		}{bestC, bestSel}
		return bestC, bestSel
	}
	c, providers := solve(need, retNeed)
	if math.IsInf(c, 1) {
		return nil
	}

	// Which members bring a party package: shaman providers, and any shaman fillers tried below.
	allowedShamans := []*Build{}
	for _, b := range f.Fill {
		if b.Class == proto.Class_ClassShaman && allowed(b, p) {
			allowedShamans = append(allowedShamans, b)
		}
	}
	var best *Comp
	for extra := 0; extra <= r.Parties; extra++ {
		for _, sh := range shamanChoices(allowedShamans, extra, v0) {
			comp := sc.seatAndFill(r, p, f, val, base, append(append([]*Build{}, r.Fixed...), append(providers, sh...)...),
				len(r.Fixed), len(r.Fixed)+len(providers), slots+len(r.Fixed))
			if comp == nil {
				continue
			}
			if best == nil || comp.Score > best.Score {
				best = comp
			}
		}
		if extra > 0 && len(allowedShamans) == 0 {
			break
		}
	}
	return best
}

// pruneItems keeps, for each role an item covers, the 3 items the model values most (the cheapest to
// bring), and every paladin.
func pruneItems(items []item, model func(*Build) float64) []item {
	keep := make([]bool, len(items))
	for i := 0; i < numGifts; i++ {
		g := Gift(1) << i
		var idx []int
		for j, it := range items {
			if it.cover&g != 0 {
				idx = append(idx, j)
			}
		}
		sort.SliceStable(idx, func(a, b int) bool { return model(items[idx[a]].b) > model(items[idx[b]].b) })
		for k := 0; k < 3 && k < len(idx); k++ {
			keep[idx[k]] = true
		}
	}
	var out []item
	for j, it := range items {
		if keep[j] || it.ret {
			out = append(out, it)
		}
	}
	return out
}

// shamanChoices is the shaman fillers to try: n of the best shaman build of each package kind.
func shamanChoices(shamans []*Build, n int, v0 func(*Build) float64) [][]*Build {
	if n == 0 {
		return [][]*Build{nil}
	}
	bestOf := map[proto.RaidPackage]*Build{}
	for _, b := range shamans {
		if cur, ok := bestOf[b.Pkg]; !ok || v0(b) > v0(cur) {
			bestOf[b.Pkg] = b
		}
	}
	var out [][]*Build
	for _, b := range bestOf {
		sel := make([]*Build, n)
		for i := range sel {
			sel[i] = b
		}
		out = append(out, sel)
	}
	sort.Slice(out, func(i, j int) bool { return out[i][0].ID < out[j][0].ID })
	return out
}

// seatAndFill places the placeholders and the forced members in parties and fills the remaining seats
// with the best build under each seat's package (docs/raid_sim.md, Groups).
func (sc *Scenario) seatAndFill(r *Raid, p Plan, f *Fillers, val Valuer, base *Comp, forced []*Build, nFixed, nProv, slots int) *Comp {
	if len(forced) > slots {
		return nil
	}
	parties := make([]struct {
		pkg   proto.RaidPackage
		free  int
		giver bool
	}, r.Parties)
	for i := range parties {
		parties[i].free = 5
	}
	if r.Size%5 != 0 {
		parties[len(parties)-1].free = r.Size % 5
	}
	comp := &Comp{Raid: r, Plan: p, Blessings: base.Blessings, Warriors: base.Warriors, PHParty: make([]int, len(r.Combo))}
	next := 0
	placed := map[int]bool{}
	// The package givers one per party, the richest first: shaman placeholders, then shaman members.
	phOrder := make([]int, 0, len(r.Combo))
	for i, ph := range r.Combo {
		if ph.Shaman {
			phOrder = append(phOrder, i)
		}
	}
	sort.SliceStable(phOrder, func(a, b int) bool { return r.Combo[phOrder[a]].Package > r.Combo[phOrder[b]].Package })
	for _, i := range phOrder {
		if next < len(parties) {
			parties[next].pkg = r.Combo[i].Package
			parties[next].giver = true
			parties[next].free--
			comp.PHParty[i] = next
			placed[i] = true
			next++
		}
	}
	var memberParty []int = make([]int, len(forced))
	for i := range memberParty {
		memberParty[i] = -1
	}
	shamanIdx := []int{}
	for i, b := range forced {
		if b.Class == proto.Class_ClassShaman {
			shamanIdx = append(shamanIdx, i)
		}
	}
	sort.SliceStable(shamanIdx, func(a, b int) bool { return forced[shamanIdx[a]].Pkg > forced[shamanIdx[b]].Pkg })
	for _, i := range shamanIdx {
		if next < len(parties) {
			parties[next].pkg = forced[i].Pkg
			parties[next].giver = true
			parties[next].free--
			memberParty[i] = next
			next++
		}
	}
	// The other placeholders go to parties without a package first, from the last party.
	for i := range r.Combo {
		if placed[i] {
			continue
		}
		for j := len(parties) - 1; j >= 0; j-- {
			if !parties[j].giver && parties[j].pkg == 0 && parties[j].free > 0 {
				parties[j].free--
				comp.PHParty[i] = j
				placed[i] = true
				break
			}
		}
		if !placed[i] {
			for j := len(parties) - 1; j >= 0; j-- {
				if parties[j].free > 0 {
					parties[j].free--
					comp.PHParty[i] = j
					placed[i] = true
					break
				}
			}
		}
		if !placed[i] {
			return nil
		}
	}
	// The rest of the forced members and the fillers share the free seats.
	retFill := p.Paladins >= 3
	bestUnder := map[proto.RaidPackage]*Build{}
	bestV := map[proto.RaidPackage]float64{}
	for _, pk := range packages {
		bestV[pk] = math.Inf(-1)
	}
	// Only the packages the parties have (and none) need a best filler.
	present := map[proto.RaidPackage]bool{proto.RaidPackage_RaidPackageNone: true}
	for _, pt := range parties {
		if pt.free > 0 {
			present[pt.pkg] = true
		}
	}
	for _, b := range f.Fill {
		if !allowed(b, p) || (b.Class == proto.Class_ClassPaladin && !retFill) || b.Class == proto.Class_ClassShaman {
			continue
		}
		for _, pk := range packages {
			if !present[pk] {
				continue
			}
			d, _, _ := val.V(b, comp.seat(b, pk, sc))
			if d > bestV[pk] {
				bestUnder[pk], bestV[pk] = b, d
			}
		}
	}
	type seatT struct {
		party int
		pkg   proto.RaidPackage
	}
	var seats []seatT
	for j, pt := range parties {
		for k := 0; k < pt.free; k++ {
			seats = append(seats, seatT{j, pt.pkg})
		}
	}
	nonGiver := 0
	for i := range forced {
		if memberParty[i] < 0 {
			nonGiver++
		}
	}
	if nonGiver > len(seats) {
		return nil
	}
	sort.SliceStable(seats, func(a, b int) bool { return seats[a].pkg > seats[b].pkg })
	// Forced members take the seats where they gain most over the filler they displace.
	taken := make([]bool, len(seats))
	type cand struct {
		i, s int
		gain float64
	}
	var cands []cand
	for i, b := range forced {
		if memberParty[i] >= 0 {
			continue
		}
		v0, _, _ := val.V(b, comp.seat(b, 0, sc))
		for s, st := range seats {
			v, _, _ := val.V(b, comp.seat(b, st.pkg, sc))
			fillerDelta := 0.0
			if bestUnder[st.pkg] != nil && bestUnder[0] != nil {
				fillerDelta = bestV[st.pkg] - bestV[0]
			}
			cands = append(cands, cand{i, s, (v - v0) - fillerDelta})
		}
	}
	sort.SliceStable(cands, func(a, b int) bool { return cands[a].gain > cands[b].gain })
	for _, cd := range cands {
		if memberParty[cd.i] >= 0 || taken[cd.s] {
			continue
		}
		memberParty[cd.i] = seats[cd.s].party
		taken[cd.s] = true
	}
	for i := range forced {
		if memberParty[i] < 0 {
			return nil
		}
	}
	for i, b := range forced {
		pk := parties[memberParty[i]].pkg
		if b.Class == proto.Class_ClassShaman && parties[memberParty[i]].giver {
			pk = 0 // its own party's Mana Spring is its own cast
		}
		comp.Members = append(comp.Members, &Member{B: b, Party: memberParty[i], Seat: comp.seat(b, pk, sc),
			Provider: i >= nFixed && i < nProv, Fixed: i < nFixed})
	}
	for s, st := range seats {
		if taken[s] {
			continue
		}
		b := bestUnder[st.pkg]
		if b == nil {
			return nil
		}
		comp.Members = append(comp.Members, &Member{B: b, Party: st.party, Seat: comp.seat(b, st.pkg, sc)})
	}
	if len(comp.Members) != slots {
		return nil
	}
	sc.assignCooldowns(comp, val)
	for _, m := range comp.Members {
		m.DPS, m.SE, m.key = val.V(m.B, m.Seat)
	}
	comp.sum()
	return comp
}

// assignCooldowns gives the placeholders' Power Infusions and Innervates to the members who gain most.
func (sc *Scenario) assignCooldowns(c *Comp, val Valuer) {
	give := func(n int, set func(*Seat, bool)) {
		if n <= 0 {
			return
		}
		type g struct {
			m    *Member
			gain float64
		}
		var gs []g
		for _, m := range c.Members {
			if !usesMana(m.B) {
				continue
			}
			s := m.Seat
			set(&s, true)
			with, _, _ := val.V(m.B, s)
			without, _, _ := val.V(m.B, m.Seat)
			gs = append(gs, g{m, with - without})
		}
		sort.SliceStable(gs, func(i, j int) bool { return gs[i].gain > gs[j].gain })
		for i := 0; i < n && i < len(gs); i++ {
			if gs[i].gain > 0 {
				set(&gs[i].m.Seat, true)
			}
		}
	}
	give(c.Raid.phPI, func(s *Seat, v bool) { s.PI = v })
	give(c.Raid.phInn, func(s *Seat, v bool) { s.Innervate = v })
}

// Describe is a comp in one line per build.
func (c *Comp) Describe() string {
	count := map[string]int{}
	var order []string
	for _, m := range c.Members {
		id := m.B.ID
		if count[id] == 0 {
			order = append(order, id)
		}
		count[id]++
	}
	sort.SliceStable(order, func(i, j int) bool { return count[order[i]] > count[order[j]] })
	var parts []string
	for _, id := range order {
		parts = append(parts, fmt.Sprintf("%d %s", count[id], id))
	}
	return strings.Join(parts, ", ")
}
