package compsearch

import (
	"math"
	"math/bits"
	"slices"

	"github.com/wowsims/classic/sim/core/proto"
)

// The model (docs/raid_sim.md, The search, step 2) ranks buff sets without sims: each build's DPS under the
// full set, times what each missing role leaves of it. The role costs come from the search's own
// single-player sims, one per build with each role removed from the full set, at the run's length, race
// and camp setting. Two pairs interact too much to multiply, so they're simmed together: Sunder with the
// minor armor slot (armor), and the two air totems.

// modelRoles are the roles the model removes one at a time.
var modelRoles = []Gift{GiftBattleShout, GiftFuriousHowl, GiftSunderTank, GiftMinorArmor, GiftCoE, GiftCrusader, GiftJoW, GiftSoE,
	GiftGoA, GiftWF, GiftFT, GiftMoonkin, GiftLotP, GiftTrueshot, GiftAB, GiftGotW, GiftPoS}

// The mana sources don't multiply: a caster loses little from any one while the others suffice, and runs
// dry without them all. So a mana user is simmed under every mix of them, and the model reads the mix.
var manaGifts = []Gift{GiftJoW, GiftAB, GiftPoS}

const manaMixes = 1 << 4 * 4 // JoW, AB, PoS, Wisdom on or off, times the 4 packages

func manaIndex(up Gift, wisdom bool, pkg proto.RaidPackage) int {
	i := 0
	for k, g := range manaGifts {
		if up.Has(g) {
			i |= 1 << k
		}
	}
	if wisdom {
		i |= 1 << 3
	}
	return i*4 + int(pkg)
}

var packages = []proto.RaidPackage{proto.RaidPackage_RaidPackageNone, proto.RaidPackage_RaidPackageManaSpring,
	proto.RaidPackage_RaidPackageManaSpringTide, proto.RaidPackage_RaidPackageRestoration}

// fullGifts is every effect.
const fullGifts = Gift(1)<<numGifts - 1

// buildModel is one build's role sims.
type buildModel struct {
	full     float64
	se       float64
	ratio    [32]float64 // DPS without the role / full, by the role's bit (an array: V runs millions of times)
	armor    float64     // without Sunder upkeep and the minor armor slot / full
	air      float64     // without Grace of Air and Windfury / full
	shout    float64     // without Battle Shout and Furious Howl (one slot) / full
	blessing [4]float64  // without each blessing / full (index by proto.Blessing)
	pkg      [4]float64  // with each package / full (full has the Restoration package)
	pi       float64     // what Power Infusion adds
	inn      float64     // what Innervate adds
	mana     []float64   // a mana user's DPS under each mix of mana sources / full; nil for the others
}

// Model is every build's role sims for one scenario.
type Model struct {
	builds map[*Build]*buildModel
}

type modelAsk struct {
	b    *Build
	seat Seat
	set  func(bm *buildModel, v float64)
}

// fullSeat is the model's full set for a build.
func (sc *Scenario) fullSeat(b *Build) Seat {
	return Seat{Up: fullGifts, Blessings: allBlessings, Package: proto.RaidPackage_RaidPackageRestoration,
		Camps: sc.Camps, Own: b.Casts, Assumptions: sc.Assumptions}
}

// BuildModel sims the role costs of the builds.
func (sc *Scenario) BuildModel(builds []*Build) (*Model, error) {
	m := &Model{builds: map[*Build]*buildModel{}}
	var asks []modelAsk
	for _, b := range builds {
		bm := &buildModel{}
		m.builds[b] = bm
		full := sc.fullSeat(b)
		asks = append(asks, modelAsk{b, full, func(bm *buildModel, v float64) { bm.full = v }})
		for _, r := range modelRoles {
			r := r
			s := full
			s.Up &^= r
			asks = append(asks, modelAsk{b, s, func(bm *buildModel, v float64) { bm.ratio[giftBit(r)] = v }})
		}
		s := full
		s.Up &^= GiftSunderTank | GiftMinorArmor
		asks = append(asks, modelAsk{b, s, func(bm *buildModel, v float64) { bm.armor = v }})
		s = full
		s.Up &^= GiftGoA | GiftWF
		asks = append(asks, modelAsk{b, s, func(bm *buildModel, v float64) { bm.air = v }})
		s = full
		s.Up &^= GiftBattleShout | GiftFuriousHowl
		asks = append(asks, modelAsk{b, s, func(bm *buildModel, v float64) { bm.shout = v }})
		for i, bl := range allBlessings {
			s := full
			s.Blessings = append(append([]proto.Blessing{}, allBlessings[:i]...), allBlessings[i+1:]...)
			bl := bl
			asks = append(asks, modelAsk{b, s, func(bm *buildModel, v float64) { bm.blessing[bl] = v }})
		}
		for _, p := range packages[:3] {
			s := full
			s.Package = p
			p := p
			asks = append(asks, modelAsk{b, s, func(bm *buildModel, v float64) { bm.pkg[p] = v }})
		}
		s = full
		s.PI = true
		asks = append(asks, modelAsk{b, s, func(bm *buildModel, v float64) { bm.pi = v }})
		s = full
		s.Innervate = true
		asks = append(asks, modelAsk{b, s, func(bm *buildModel, v float64) { bm.inn = v }})
		if usesMana(b) {
			bm.mana = make([]float64, manaMixes)
			for i := 0; i < manaMixes; i++ {
				s := full
				mix, pkg := i/4, proto.RaidPackage(i%4)
				s.Package = pkg
				for k, g := range manaGifts {
					if mix&(1<<k) == 0 {
						s.Up &^= g
					}
				}
				if mix&(1<<3) == 0 {
					s.Blessings = []proto.Blessing{proto.Blessing_BlessingKings, proto.Blessing_BlessingMight}
				}
				i := i
				asks = append(asks, modelAsk{b, s, func(bm *buildModel, v float64) { bm.mana[i] = v }})
			}
		}
	}
	jobs := make([]*Job, len(asks))
	for i, a := range asks {
		jobs[i] = sc.job(a.b, a.seat, sc.Iterations)
	}
	if err := sc.sim(jobs); err != nil {
		return nil, err
	}
	for i, a := range asks {
		a.set(m.builds[a.b], jobs[i].Value.DPS)
		if i == 0 || asks[i-1].b != a.b {
			m.builds[a.b].se = jobs[i].Value.SE
		}
	}
	for _, bm := range m.builds {
		f := bm.full
		if f <= 0 {
			continue
		}
		for r, v := range bm.ratio {
			bm.ratio[r] = v / f
		}
		bm.armor /= f
		bm.air /= f
		bm.shout /= f
		for i := range bm.blessing {
			if bm.blessing[i] != 0 {
				bm.blessing[i] /= f
			}
		}
		bm.pkg[proto.RaidPackage_RaidPackageRestoration] = f
		for i := range bm.pkg {
			bm.pkg[i] /= f
		}
		bm.pi -= f
		bm.inn -= f
		for i := range bm.mana {
			bm.mana[i] /= f
		}
	}
	return m, nil
}

// giftBit is a single gift's bit index.
func giftBit(g Gift) int { return bits.TrailingZeros32(uint32(g)) }

// V is the model's DPS for a build in a seat.
func (m *Model) V(b *Build, s Seat) float64 {
	bm := m.builds[b]
	if bm == nil {
		return 0
	}
	v := bm.full
	up := s.Up | b.Casts
	missing := fullGifts &^ up
	armorBoth := missing.Has(GiftSunderTank | GiftMinorArmor)
	airBoth := missing.Has(GiftGoA | GiftWF)
	shoutBoth := missing.Has(GiftBattleShout | GiftFuriousHowl)
	for _, r := range modelRoles {
		if !missing.Has(r) {
			continue
		}
		if bm.mana != nil && (r == GiftJoW || r == GiftAB || r == GiftPoS) {
			continue
		}
		if armorBoth && (r == GiftSunderTank || r == GiftMinorArmor) {
			continue
		}
		if airBoth && (r == GiftGoA || r == GiftWF) {
			continue
		}
		if shoutBoth && (r == GiftBattleShout || r == GiftFuriousHowl) {
			continue
		}
		v *= bm.ratio[giftBit(r)]
	}
	if armorBoth {
		v *= bm.armor
	}
	if airBoth {
		v *= bm.air
	}
	if shoutBoth {
		v *= bm.shout
	}
	wisdom := false
	for _, bl := range allBlessings {
		has := false
		for _, x := range s.Blessings {
			if x == bl {
				has = true
			}
		}
		if bl == proto.Blessing_BlessingWisdom {
			wisdom = has
			if bm.mana != nil {
				continue
			}
		}
		if !has {
			v *= bm.blessing[bl]
		}
	}
	if bm.mana != nil {
		v *= bm.mana[manaIndex(up, wisdom, s.Package)]
	} else {
		v *= bm.pkg[s.Package]
	}
	if s.PI {
		v += bm.pi
	}
	if s.Innervate {
		v += bm.inn
	}
	return v
}

// costs is a build's DPS lost to each missing role, for the dominance rule.
func (m *Model) costs(b *Build) map[string]float64 {
	bm := m.builds[b]
	out := map[string]float64{}
	if bm == nil {
		return out
	}
	for _, r := range modelRoles {
		out[GiftName(r)] = bm.full * (1 - bm.ratio[giftBit(r)])
	}
	out["armor pair"] = bm.full * (1 - bm.armor)
	out["air pair"] = bm.full * (1 - bm.air)
	out["shout pair"] = bm.full * (1 - bm.shout)
	for _, bl := range allBlessings {
		out[bl.String()] = bm.full * (1 - bm.blessing[bl])
	}
	for _, p := range packages[:3] {
		out[p.String()] = bm.full * (1 - bm.pkg[p])
	}
	out["Power Infusion"] = bm.pi
	out["Innervate"] = bm.inn
	return out
}

// Survivors applies the dominance rule (docs/raid_sim.md, The search, step 1): a dominates b when its DPS
// under the full set, less b's, covers every role where a loses more than b, and a can fill wherever b can.
// Power Infusion and Innervate count as roles that are up in the full set.
func (m *Model) Survivors(builds []*Build, margin float64) (alive []*Build, by map[*Build]*Build) {
	by = map[*Build]*Build{}
	full := func(b *Build) float64 {
		bm := m.builds[b]
		return bm.full + bm.pi + bm.inn
	}
	for _, b := range builds {
		cb := m.costs(b)
		for _, a := range builds {
			if a == b || a.Gives&^b.Gives != 0 {
				continue
			}
			ca := m.costs(a)
			need := margin
			// In a fixed order, so a decision on the edge comes out the same every run.
			keys := make([]string, 0, len(ca))
			for k := range ca {
				keys = append(keys, k)
			}
			slices.Sort(keys)
			for _, k := range keys {
				need += math.Max(0, ca[k]-cb[k])
			}
			if full(a)-full(b) >= need {
				by[b] = a
				break
			}
		}
		if by[b] == nil {
			alive = append(alive, b)
		}
	}
	return alive, by
}
