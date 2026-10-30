package compsearch

import (
	"fmt"
	"strings"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
	googleProto "google.golang.org/protobuf/proto"
)

// Gift is a set of raid-wide effects (docs/raid_sim.md, The provider table).
type Gift uint32

const (
	GiftBattleShout Gift = 1 << iota
	GiftSunderTank       // a warrior tank keeps Sunder Armor up after the pull
	GiftMinorArmor       // the minor armor slot: Faerie Fire (or Curse of Recklessness, Dust Cloud)
	GiftCoE              // Curse of the Elements
	GiftCrusader         // Judgement of the Crusader
	GiftJoW              // Judgement of Wisdom
	GiftSoE              // Strength of Earth Totem
	GiftGoA              // Grace of Air Totem
	GiftWF               // Windfury Totem
	GiftFT               // Flametongue Totem
	GiftMoonkin
	GiftLotP
	GiftTrueshot
	GiftAB
	GiftGotW
	GiftPoS         // Prayer of Spirit (Divine Spirit)
	GiftFort        // Prayer of Fortitude
	GiftBlessing    // a paladin: one Greater Blessing for each class
	GiftFuriousHowl // a hunter's wolf: 83 melee attack power in Battle Shout's slot
	numGifts        = iota
)

func (g Gift) Has(o Gift) bool { return g&o == o }

func (g Gift) String() string {
	var parts []string
	for i := 0; i < numGifts; i++ {
		if g&(1<<i) != 0 {
			parts = append(parts, effectOf[1<<i].Name)
		}
	}
	if len(parts) == 0 {
		return "nothing"
	}
	return strings.Join(parts, ", ")
}

// GiftName is one gift bit's name.
func GiftName(g Gift) string {
	if e, ok := effectOf[g]; ok {
		return e.Name
	}
	return g.String()
}

// GiftByName is the gift a buff record names, like "Fortitude" (or an alias, like Salvation).
func GiftByName(name string) (Gift, bool) {
	for _, e := range Effects {
		if e.Name == name {
			return e.Gift, true
		}
	}
	return 0, false
}

// RequiredGifts is what a search request requires every comp to give. Its names were validated.
func RequiredGifts(req *proto.CompSearchRequest) Gift {
	var g Gift
	for _, name := range req.RequiredEffects {
		if x, ok := GiftByName(name); ok {
			g |= x
		}
	}
	return g
}

// DutyGift is what a duty gives the raid.
func DutyGift(d proto.Duty) Gift {
	if r := dutyRow(d); r != nil {
		return r.Gift
	}
	return 0
}

// buildGifts says what a build gives the raid without a duty, the party package a shaman brings, and the
// effects its own rotation casts on itself (which its own request then leaves out).
func buildGifts(b *Build) (gives Gift, pkg proto.RaidPackage, casts Gift) {
	return PlayerGifts(player0(b.Req), b.Duty)
}

// PlayerGifts is buildGifts for a player and the duty it holds. The held auras come from the class hooks,
// read from a one-player environment, and a shaman's Mana Tide from its talents.
func PlayerGifts(player *proto.Player, duty proto.Duty) (gives Gift, pkg proto.RaidPackage, casts Gift) {
	raid := &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{googleProto.Clone(player).(*proto.Player)}}}}
	env, _, _ := core.NewEnvironment(raid, &proto.Encounter{}, proto.Ruleset_RulesetForever, false)
	agent := env.Raid.Parties[0].Players[0]
	spec := specOfPlayer(player)
	gives |= spec.Gives
	casts |= spec.Casts
	if spec.Package {
		pkg = proto.RaidPackage_RaidPackageManaSpring
		// A shaman with Mana Tide Totem gives its party Mana Tide too.
		if sa, ok := agent.(shaman.ShamanAgent); ok && sa.GetShaman().Talents.ManaTideTotem {
			pkg = proto.RaidPackage_RaidPackageManaSpringTide
		}
	}
	// The auras a build holds by its talents, from the class hooks.
	rb := env.Raid.GetRaidBuffs(&proto.RaidBuffs{})
	if rb.MoonkinAura {
		gives |= GiftMoonkin
	}
	if rb.LeaderOfThePack {
		gives |= GiftLotP
	}
	if rb.TrueshotAura {
		gives |= GiftTrueshot
	}
	if rb.ArcaneBrilliance {
		gives |= GiftAB
	}
	if rb.DivineSpirit {
		gives |= GiftPoS
	}
	if d := DutyGift(duty); d != 0 {
		gives |= d
		casts |= d
	}
	// A hunter's pet: a Tallstrider's Dust Cloud fills the minor armor slot, a wolf's Furious Howl shares
	// Battle Shout's slot.
	if h := player.GetHunter(); h != nil && h.Options != nil {
		switch h.Options.PetType {
		case proto.Hunter_Options_Tallstrider:
			gives |= GiftMinorArmor
			casts |= GiftMinorArmor
		case proto.Hunter_Options_Wolf:
			gives |= GiftFuriousHowl
			casts |= GiftFuriousHowl
		}
	}
	return
}

// usesMana is whether a build has a mana bar, so mana effects can move its DPS.
func usesMana(b *Build) bool { return b.SpecRow().Mana }

// Seat is everything one member's single-player request gets from the rest of the raid
// (docs/raid_sim.md, Writing a role into a request).
type Seat struct {
	Up          Gift // the raid-wide effects up, from anyone
	Blessings   []proto.Blessing
	PullSunders int // the other warriors' pull Sunders
	Package     proto.RaidPackage
	PI          bool
	Innervate   bool
	Camps       bool
	Own         Gift // effects this member casts itself, left out of its own request
	Assumptions *proto.RaidAssumptions
}

func (s Seat) key() string {
	return fmt.Sprintf("%x|%v|%d|%d|%v|%v|%v|%x|%v", uint32(s.Up), s.Blessings, s.PullSunders, s.Package, s.PI, s.Innervate, s.Camps, uint32(s.Own), s.Assumptions.String())
}

// Write fills a request's buffs and debuffs with what the seat gets, starting from nothing, so the
// report's own full set doesn't leak through.
func (s Seat) Write(req *proto.RaidSimRequest, b *Build) {
	up := s.Up &^ s.Own
	rb := &proto.RaidBuffs{}
	db := &proto.Debuffs{}
	pb := &proto.PartyBuffs{}
	ib := &proto.IndividualBuffs{}

	w := &buffWriter{rb: rb, db: db, spec: b.SpecRow()}
	for i := 0; i < numGifts; i++ {
		if e := effectOf[1<<i]; up.Has(e.Gift) && e.write != nil {
			e.write(w)
		}
	}
	if r := packageRow(s.Package); r != nil {
		r.write(rb, pb)
	}
	for _, bl := range s.Blessings {
		for _, r := range Blessings {
			if r.Blessing == bl {
				r.write(ib)
			}
		}
	}
	if s.PI {
		ib.PowerInfusions = 1
	}
	if s.Innervate {
		ib.Innervates = 1
	}

	// A member keeping Sunder up itself gets no tank's ramp, only the other warriors' pull Sunders.
	if up.Has(GiftSunderTank) || s.PullSunders > 0 {
		db.SunderRamp = &proto.SunderRamp{Tank: up.Has(GiftSunderTank), PullSunders: int32(s.PullSunders)}
	}

	if s.Camps {
		rb.CampBuffs = &proto.CampBuffs{FishBowl: true, Lodestone: true, SharpeningWheel: true, CampChair: true,
			IncenseCandle: true, FactionBanner: true, ManaWell: true, EnchantedLute: true, FirstAidKit: true}
		if a := s.Assumptions; a != nil {
			rb.CampBuffs.CampChairSpellCritOnly = a.CampChairSpellCritOnly
			rb.CampBuffs.StackWithGroupBuffs = a.CampBuffsStackWithGroupBuffs
		}
	}

	req.Raid.Buffs = rb
	req.Raid.Debuffs = db
	req.Raid.Parties[0].Buffs = pb
	req.Raid.Parties[0].Players[0].Buffs = ib
}

func cloneReq(r *proto.RaidSimRequest) *proto.RaidSimRequest {
	return googleProto.Clone(r).(*proto.RaidSimRequest)
}
