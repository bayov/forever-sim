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

var giftNames = []string{"Battle Shout", "Sunder upkeep", "Minor armor slot", "Curse of the Elements",
	"Judgement of the Crusader", "Judgement of Wisdom", "Strength of Earth", "Grace of Air", "Windfury Totem",
	"Flametongue Totem", "Moonkin Aura", "Leader of the Pack", "Trueshot Aura", "Arcane Brilliance",
	"Gift of the Wild", "Prayer of Spirit", "Fortitude", "Blessings", "Furious Howl"}

func (g Gift) Has(o Gift) bool { return g&o == o }

func (g Gift) String() string {
	var parts []string
	for i := 0; i < numGifts; i++ {
		if g&(1<<i) != 0 {
			parts = append(parts, giftNames[i])
		}
	}
	if len(parts) == 0 {
		return "nothing"
	}
	return strings.Join(parts, ", ")
}

// GiftName is one gift bit's name.
func GiftName(g Gift) string {
	for i := 0; i < numGifts; i++ {
		if g == 1<<i {
			return giftNames[i]
		}
	}
	return g.String()
}

// GiftByName is the gift a buff record names, like "Fortitude".
//
// Salvation is any paladin's under Forever (spell 1038): it's no longer a Blessing and takes no Blessing
// slot, so requiring it is requiring a paladin, as Blessings does.
func GiftByName(name string) (Gift, bool) {
	if name == "Salvation" {
		return GiftBlessing, true
	}
	for i, n := range giftNames {
		if n == name {
			return Gift(1) << i, true
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

var dutyGift = map[proto.Duty]Gift{
	proto.Duty_DutyBattleShout:            GiftBattleShout,
	proto.Duty_DutySunderUpkeep:           GiftSunderTank,
	proto.Duty_DutyFaerieFire:             GiftMinorArmor,
	proto.Duty_DutyCurseOfRecklessness:    GiftMinorArmor,
	proto.Duty_DutyCurseOfTheElements:     GiftCoE,
	proto.Duty_DutyJudgementOfTheCrusader: GiftCrusader,
	proto.Duty_DutyJudgementOfWisdom:      GiftJoW,
}

// DutyGift is what a duty gives the raid.
func DutyGift(d proto.Duty) Gift { return dutyGift[d] }

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
	switch player.Class {
	case proto.Class_ClassDruid:
		gives |= GiftGotW
		// The hook gives Leader of the Pack only in Cat or Bear Form, which a Feral druid takes after
		// the environment is built.
		if player.GetFeralDruid() != nil {
			gives |= GiftLotP
		}
	case proto.Class_ClassPaladin:
		gives |= GiftBlessing
	case proto.Class_ClassMage:
		gives |= GiftAB
	case proto.Class_ClassPriest:
		gives |= GiftPoS | GiftFort
	case proto.Class_ClassShaman:
		pkg = proto.RaidPackage_RaidPackageManaSpring
		if sa, ok := agent.(shaman.ShamanAgent); ok && sa.GetShaman().Talents.ManaTideTotem {
			pkg = proto.RaidPackage_RaidPackageManaSpringTide
		}
		if player.GetEnhancementShaman() != nil {
			// Its rotation drops Strength of Earth, Windfury, Mana Spring and Magma before the pull.
			gives |= GiftSoE | GiftWF
			casts |= GiftSoE | GiftWF
		}
	}
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
func usesMana(b *Build) bool {
	switch b.Class {
	case proto.Class_ClassWarrior, proto.Class_ClassRogue:
		return false
	case proto.Class_ClassDruid:
		return b.Spec != "Feral"
	}
	return true
}

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
	regular := proto.TristateEffect_TristateEffectRegular

	if up.Has(GiftBattleShout) {
		rb.BattleShout = regular
	}
	rb.FuriousHowl = up.Has(GiftFuriousHowl)
	if up.Has(GiftSoE) {
		rb.StrengthOfEarthTotem = regular
	}
	if up.Has(GiftGoA) {
		rb.GraceOfAirTotem = regular
	}
	// One weapon totem per receiver: Flametongue for Enhancement, whose Windfury Weapon makes Windfury
	// Totem useless to it, and Windfury for everyone else.
	if b.Class == proto.Class_ClassShaman && b.Spec == "Enhancement" {
		if up.Has(GiftFT) {
			rb.TotemWeaponBuff = proto.TotemWeaponBuff_TotemWeaponBuffFlametongue
		}
	} else if up.Has(GiftWF) {
		rb.TotemWeaponBuff = proto.TotemWeaponBuff_TotemWeaponBuffWindfury
	}
	rb.MoonkinAura = up.Has(GiftMoonkin)
	rb.LeaderOfThePack = up.Has(GiftLotP)
	rb.TrueshotAura = up.Has(GiftTrueshot)
	rb.ArcaneBrilliance = up.Has(GiftAB)
	rb.DivineSpirit = up.Has(GiftPoS)
	if up.Has(GiftGotW) {
		rb.GiftOfTheWild = regular
	}
	if up.Has(GiftFort) {
		rb.PowerWordFortitude = regular
	}

	switch s.Package {
	case proto.RaidPackage_RaidPackageManaSpring:
		rb.ManaSpringTotem = regular
	case proto.RaidPackage_RaidPackageManaSpringTide:
		rb.ManaSpringTotem = regular
		pb.ManaTideTotems = 1
	case proto.RaidPackage_RaidPackageRestoration:
		rb.ManaSpringTotem = proto.TristateEffect_TristateEffectImproved
		pb.ManaTideTotems = 1
	}

	for _, bl := range s.Blessings {
		switch bl {
		case proto.Blessing_BlessingKings:
			ib.BlessingOfKings = true
		case proto.Blessing_BlessingMight:
			ib.BlessingOfMight = regular
		case proto.Blessing_BlessingWisdom:
			ib.BlessingOfWisdom = regular
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
	db.FaerieFire = up.Has(GiftMinorArmor)
	db.CurseOfElements = up.Has(GiftCoE)
	if up.Has(GiftCrusader) {
		db.JudgementOfTheCrusader = regular
	}
	db.JudgementOfWisdom = up.Has(GiftJoW)

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
