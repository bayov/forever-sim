package hunter

import (
	"cmp"
	"math"
	"slices"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// The family abilities Forever added or changed, from client 1.60.1.70532 (SpellEffect, SpellMisc,
// SpellCooldowns; the same rows as wowsims/forever's spell store). Each rank is learned at its
// level and grows by its per-level amount up to its max level. None of the new rows carries a bonus
// coefficient, so Pet Aggression's 8 a hit adds to Bite and Claw only.
type foreverPetRank struct {
	level    int32
	maxLevel int32
	spellID  int32
	base     float64
	perLevel float64
	variance float64
}

func (rank foreverPetRank) amount(level int32) float64 {
	return rank.base + rank.perLevel*float64(min(level, rank.maxLevel)-rank.level)
}

// The damage range is the base points times 1 -/+ half the row's variance, as for Bite and Claw.
func (rank foreverPetRank) roll(sim *core.Simulation, level int32) float64 {
	amount := rank.amount(level)
	return sim.Roll(amount*(1-rank.variance/2), amount*(1+rank.variance/2))
}

// The highest rank learned at the level, if any.
func foreverPetRankAt(level int32, ranks []foreverPetRank) (foreverPetRank, bool) {
	for i := len(ranks) - 1; i >= 0; i-- {
		if ranks[i].level <= level {
			return ranks[i], true
		}
	}
	return foreverPetRank{}, false
}

// Most of the new strikes' rows have variance 0.1389.
const foreverPetVariance = 0.13888889

var foreverPetRanks = map[PetAbilityType][]foreverPetRank{
	// Raptor: a bleed, 26 a tick every 3 sec for 18 sec at 60.
	SavageRend: {
		{12, 24, 1265065, 5, 0, 0}, {24, 36, 1265066, 9, 0, 0}, {36, 48, 1265067, 12, 0, 0},
		{48, 60, 1265068, 20, 0, 0}, {60, 60, 1265069, 26, 0, 0},
	},
	// Gorilla: Nature damage to up to 4 enemies within 8 yd of the pet.
	Thunderstomp: {
		{30, 34, 26090, 57, 0, 0.13888889}, {40, 44, 26187, 74, 0, 0.12903225},
		{50, 54, 26188, 99, 0, 0.14516129}, {60, 60, 1264455, 132, 0, 0.14516129},
	},
	// Bear: a hit that chains to 3 enemies.
	Swipe: {
		{12, 24, 1264494, 4, 0, foreverPetVariance}, {24, 36, 1264497, 9, 0, foreverPetVariance},
		{36, 48, 1264498, 12, 0, foreverPetVariance}, {48, 60, 1264501, 17, 0, foreverPetVariance},
		{60, 60, 1264502, 21, 0, foreverPetVariance},
	},
	// Crab: a hit and a 50% snare.
	Pinch: {
		{12, 24, 1264735, 20, 0, foreverPetVariance}, {24, 36, 1264736, 34, 0, foreverPetVariance},
		{36, 48, 1264739, 44, 0, foreverPetVariance}, {48, 60, 1264741, 73, 0, foreverPetVariance},
		{60, 60, 1264742, 95, 0, foreverPetVariance},
	},
	// Crocolisk: a hit and 25% less healing taken.
	Dismember: {
		{12, 24, 1264758, 12, 0, foreverPetVariance}, {24, 36, 1264927, 19, 0, foreverPetVariance},
		{36, 48, 1264929, 28, 0, foreverPetVariance}, {48, 60, 1264930, 42, 0, foreverPetVariance},
		{60, 60, 1264933, 54, 0, foreverPetVariance},
	},
	// Bird of Prey: a hit and a disarm.
	Mine: {
		{10, 20, 1265054, 11, 0, foreverPetVariance}, {20, 30, 1265055, 15, 0, foreverPetVariance},
		{30, 40, 1265056, 22, 0, foreverPetVariance}, {40, 50, 1265057, 35, 0, foreverPetVariance},
		{50, 60, 1265058, 44, 0, foreverPetVariance},
	},
	// Hyena: a bleed, 20 a tick every 3 sec for 9 sec at 60, and a 50% snare.
	TendonRip: {
		{12, 24, 1265038, 4, 0, 0}, {24, 36, 1265039, 7, 0, 0}, {36, 48, 1265040, 9, 0, 0},
		{48, 60, 1265041, 15, 0, 0}, {60, 60, 1265042, 20, 0, 0},
	},
	// Spider: Nature damage, 13 a sec for 4 sec at 60, and a root.
	Web: {
		{12, 24, 1265843, 3, 0, 0}, {24, 36, 1265878, 5, 0, 0}, {36, 48, 1265880, 7, 0, 0},
		{48, 60, 1265881, 10, 0, 0}, {60, 60, 1265883, 13, 0, 0},
	},
	// Tallstrider: armor off the enemy.
	DustCloud: {
		{12, 24, 1265899, 65, 0, 0}, {24, 36, 1265901, 175, 0, 0}, {36, 48, 1265902, 285, 0, 0},
		{48, 60, 1265903, 395, 0, 0}, {60, 60, 1265904, 505, 0, 0},
	},
	// Wolf: melee attack power for the pet and its allies.
	FuriousHowl: {
		{10, 24, 24604, 5, 0.6, 0}, {24, 40, 24605, 34, 0.02, 0}, {40, 56, 24603, 60, 0.5, 0},
		{56, 60, 24597, 82, 0.4, 0},
	},
	// Fox: 30% attack speed (and 50% dodge) for 12 sec.
	TrickstersDance: {
		{20, 20, 1310612, 30, 0, 0},
	},
	// Boar: the pet charges and its next melee hit carries the attack power.
	Charge: {
		{1, 12, 7371, 44, 2, 0}, {12, 24, 26177, 106, 2, 0}, {24, 36, 26178, 168, 3, 0},
		{36, 48, 26179, 246, 4, 0}, {48, 60, 26201, 370, 5, 0}, {60, 66, 27685, 490, 5, 0},
	},
	// Movement speed for 15 sec.
	Dash: {{30, 30, 23099, 40, 0, 0}, {40, 40, 23109, 60, 0, 0}, {50, 50, 23110, 80, 0, 0}},
	Dive: {{30, 30, 23145, 40, 0, 0}, {40, 40, 23147, 60, 0, 0}, {50, 50, 23148, 80, 0, 0}},
}

// A Forever pet ability and when the pet wants it: a buff or debuff it keeps up is skipped while it
// is.
type foreverPetAction struct {
	spell  *core.Spell
	wanted func(sim *core.Simulation, target *core.Unit) bool
}

// The family's abilities under Forever, by the rows above or the Classic ones' (pet_abilities.go).
// The pet casts them longest cooldown first, as wowsims/forever's does (pets.go, Initialize).
func (hp *HunterPet) initForeverAbilities() {
	for _, abilityType := range hp.config.ForeverAbilities {
		switch abilityType {
		case Bite, Claw, Screech, LightningBreath, ScorpidPoison:
			hp.addForeverAction(hp.NewPetAbility(abilityType, false), nil)
		case Charge:
			if spell := hp.newForeverCharge(); spell != nil {
				hp.charge = spell
				hp.closers = append(hp.closers, spell)
			}
		case Dash, Dive:
			if spell := hp.newForeverSprint(abilityType); spell != nil {
				hp.closers = append(hp.closers, spell)
			}
		case FuriousHowl:
			spell, wanted := hp.newForeverFuriousHowl()
			hp.addForeverAction(spell, wanted)
		case DustCloud:
			spell, wanted := hp.newForeverDustCloud()
			hp.addForeverAction(spell, wanted)
		case TrickstersDance:
			hp.addForeverAction(hp.newForeverTrickstersDance(), nil)
		case SavageRend, TendonRip, Web:
			hp.addForeverAction(hp.newForeverDot(abilityType), nil)
		default:
			hp.addForeverAction(hp.newForeverStrike(abilityType), nil)
		}
	}
	slices.SortStableFunc(hp.foreverActions, func(a, b foreverPetAction) int {
		return cmp.Compare(b.spell.CD.Duration, a.spell.CD.Duration)
	})
}

func (hp *HunterPet) addForeverAction(spell *core.Spell, wanted func(sim *core.Simulation, target *core.Unit) bool) {
	if spell != nil {
		hp.foreverActions = append(hp.foreverActions, foreverPetAction{spell: spell, wanted: wanted})
	}
}

// Under Forever the pet runs in first, casting its Charge or Dash (or Dive) on the way, and in melee
// range casts the first ability it can, longest cooldown first. A cooldown that is up but short of
// focus is waited for rather than spent on a filler, as wowsims/forever's ExecuteCustomRotation has
// it.
func (hp *HunterPet) foreverRotation(sim *core.Simulation) {
	if hp.DistanceFromTarget > core.MaxMeleeAttackDistance {
		hp.foreverCloseIn(sim)
		return
	}

	target := hp.CurrentTarget
	for _, action := range hp.foreverActions {
		if action.wanted != nil && !action.wanted(sim, target) {
			continue
		}
		if action.spell.CanCast(sim, target) {
			action.spell.Cast(sim, target)
			return
		}
		if action.spell.CD.Duration > 0 && action.spell.IsReady(sim) {
			return
		}
	}
}

// Charge reaches 8 to 25 yd (27685: MinRange 8, MaxRange 25).
const (
	foreverChargeMinRange = 8
	foreverChargeMaxRange = 25
)

// The boar runs to Charge's 25 yd when it is further and the charge is ready. Otherwise the pet casts
// its first closer it can and runs in.
func (hp *HunterPet) foreverCloseIn(sim *core.Simulation) {
	if hp.IsMoving() {
		return
	}
	if hp.charge != nil && hp.charge.IsReady(sim) && hp.DistanceFromTarget > foreverChargeMaxRange {
		hp.moveToward(sim, foreverChargeMaxRange)
		return
	}
	for _, closer := range hp.closers {
		if closer.CanCast(sim, hp.CurrentTarget) {
			closer.Cast(sim, hp.CurrentTarget)
			break
		}
	}
	if !hp.IsMoving() {
		hp.runIn(sim)
	}
}

// The ability's config as the pet casts it: the row's focus and the client's 1.5 sec global cooldown.
func (hp *HunterPet) foreverPetSpellConfig(rank foreverPetRank, school core.SpellSchool, focus float64, cooldown time.Duration) core.SpellConfig {
	config := core.SpellConfig{
		ActionID:    core.ActionID{SpellID: rank.spellID},
		SpellSchool: school,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics,

		FocusCost: core.FocusCostOptions{
			Cost: focus,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: ForeverPetGCD,
			},
			IgnoreHaste: true,
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
	}
	switch school {
	case 0:
		// A buff, or the boar's charge.
		config.DefenseType = core.DefenseTypeNone
		config.ProcMask = core.ProcMaskEmpty
		config.Flags = 0
		config.DamageMultiplier = 0
		config.ThreatMultiplier = 0
	case core.SpellSchoolPhysical:
	default:
		config.DefenseType = core.DefenseTypeMagic
		config.ProcMask = core.ProcMaskSpellDamage
		config.Flags = 0
	}
	if cooldown > 0 {
		config.Cast.CD = core.Cooldown{
			Timer:    hp.NewTimer(),
			Duration: cooldown,
		}
	}
	return config
}

// A direct hit: its school, focus, cooldown and how many enemies it reaches, from the rows.
var foreverPetStrikes = map[PetAbilityType]struct {
	school   core.SpellSchool
	focus    float64
	cooldown time.Duration
	targets  int32
}{
	Thunderstomp: {core.SpellSchoolNature, 60, time.Minute, 4},
	Swipe:        {core.SpellSchoolPhysical, 20, time.Second * 5, 3},
	Pinch:        {core.SpellSchoolPhysical, 50, time.Second * 30, 1},
	Dismember:    {core.SpellSchoolPhysical, 35, time.Second * 15, 1},
	Mine:         {core.SpellSchoolPhysical, 20, time.Second * 90, 1},
}

// Pinch's snare, Dismember's healing cut and Mine!'s disarm change nothing the sim measures.
func (hp *HunterPet) newForeverStrike(abilityType PetAbilityType) *core.Spell {
	rank, ok := foreverPetRankAt(hp.Owner.Level, foreverPetRanks[abilityType])
	if !ok {
		return nil
	}
	strike := foreverPetStrikes[abilityType]
	level := hp.Owner.Level

	config := hp.foreverPetSpellConfig(rank, strike.school, strike.focus, strike.cooldown)
	outcome := func(spell *core.Spell) core.OutcomeApplier {
		if spell.DefenseType == core.DefenseTypeMagic {
			return spell.OutcomeMagicHitAndCrit
		}
		return spell.OutcomeMeleeSpecialHitAndCrit
	}
	results := make([]*core.SpellResult, strike.targets)
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		if strike.targets == 1 {
			spell.CalcAndDealDamage(sim, target, rank.roll(sim, level), outcome(spell))
			return
		}
		hits := sim.Encounter.AoETargetCount(strike.targets)
		curTarget := target
		for i := int32(0); i < hits; i++ {
			results[i] = spell.CalcDamage(sim, curTarget, rank.roll(sim, level), outcome(spell))
			curTarget = sim.Environment.NextTargetUnit(curTarget)
		}
		for i := int32(0); i < hits; i++ {
			spell.DealDamage(sim, results[i])
		}
	}
	return hp.RegisterSpell(config)
}

// Savage Rend and Tendon Rip are bleeds, which ignore armor, and Web is Nature damage that rolls the
// ranged table to land (1265883 DefenseType 3). Their ticks crit, as every Forever dot's do (the
// rows also carry the periodic crit attribute).
func (hp *HunterPet) newForeverDot(abilityType PetAbilityType) *core.Spell {
	rank, ok := foreverPetRankAt(hp.Owner.Level, foreverPetRanks[abilityType])
	if !ok {
		return nil
	}
	dot := map[PetAbilityType]struct {
		school   core.SpellSchool
		focus    float64
		cooldown time.Duration
		ticks    int32
		period   time.Duration
	}{
		SavageRend: {core.SpellSchoolPhysical, 50, time.Minute, 6, time.Second * 3},
		TendonRip:  {core.SpellSchoolPhysical, 25, time.Second * 30, 3, time.Second * 3},
		Web:        {core.SpellSchoolNature, 20, time.Second * 40, 4, time.Second},
	}[abilityType]
	tick := rank.amount(hp.Owner.Level)

	config := hp.foreverPetSpellConfig(rank, dot.school, dot.focus, dot.cooldown)
	config.Flags |= core.SpellFlagPureDot
	if abilityType == Web {
		config.DefenseType = core.DefenseTypeRanged
	}
	if abilityType == SavageRend {
		config.SpellCode = SpellCode_HunterPetSavageRend
	}
	config.Dot = core.DotConfig{
		Aura: core.Aura{
			Label: map[PetAbilityType]string{SavageRend: "Savage Rend", TendonRip: "Tendon Rip", Web: "Web"}[abilityType] + "-" + hp.Label,
		},
		NumberOfTicks: dot.ticks,
		TickLength:    dot.period,

		OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
			dot.Snapshot(target, tick, isRollover)
		},
		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
			dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
		},
	}
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		var outcome core.OutcomeApplier
		switch spell.DefenseType {
		case core.DefenseTypeRanged:
			outcome = spell.OutcomeRangedHit
		default:
			outcome = spell.OutcomeMeleeSpecialHit
		}
		result := spell.CalcOutcome(sim, target, outcome)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		}
		spell.DealOutcome(sim, result)
	}
	return hp.RegisterSpell(config)
}

// Furious Howl (24597 at 56 to 60: 82 and 0.4 a level, 83 at 60 rounded down as Battle Shout's are)
// gives melee attack power for 60 sec on a 30 sec cooldown to the pet and its allies within 15 yd.
// It targets the caster's raid (client 70685: implicit target 56), and the raid sim takes every
// player to be in reach. It has no players-only flag, but it doesn't reach other pets until a
// launch check says so (docs/raid_sim.md, Hunter pets). Under Forever it shares Battle Shout's
// slot. The pet howls again when its own buff runs out.
func (hp *HunterPet) newForeverFuriousHowl() (*core.Spell, func(*core.Simulation, *core.Unit) bool) {
	rank, ok := foreverPetRankAt(hp.Owner.Level, foreverPetRanks[FuriousHowl])
	if !ok {
		return nil, nil
	}
	ap := math.Floor(rank.amount(hp.Owner.Level))
	petAura := core.ForeverFuriousHowlAura(&hp.Unit, rank.spellID, ap)
	raidAuras := hp.hunterOwner.NewRaidPlayerAuraArray(func(unit *core.Unit) *core.Aura {
		return core.ForeverFuriousHowlAura(unit, rank.spellID, ap)
	})

	config := hp.foreverPetSpellConfig(rank, 0, 60, time.Second*30)
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		petAura.Activate(sim)
		for _, aura := range raidAuras {
			if aura != nil {
				aura.Activate(sim)
			}
		}
	}
	return hp.RegisterSpell(config), func(_ *core.Simulation, _ *core.Unit) bool {
		return !petAura.IsActive()
	}
}

// Dust Cloud (10 focus, no cooldown) lands on the melee table and lasts 30 sec. The pet casts it
// when the enemy has no armor cut as strong in the slot, rather than as a filler, so not at all
// beside the raid's Faerie Fire or Curse of Recklessness.
func (hp *HunterPet) newForeverDustCloud() (*core.Spell, func(*core.Simulation, *core.Unit) bool) {
	rank, ok := foreverPetRankAt(hp.Owner.Level, foreverPetRanks[DustCloud])
	if !ok {
		return nil, nil
	}
	auras := hp.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.ForeverDustCloudAura(target, rank.spellID, rank.base)
	})

	config := hp.foreverPetSpellConfig(rank, core.SpellSchoolPhysical, 10, 0)
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMeleeSpecialHit)
		if result.Landed() {
			auras.Get(target).Activate(sim)
		}
	}
	config.RelatedAuras = []core.AuraArray{auras}
	return hp.RegisterSpell(config), func(sim *core.Simulation, target *core.Unit) bool {
		return auras.Get(target).ShouldRefreshExclusiveEffects(sim, 0)
	}
}

// Trickster's Dance (1310612): 30% melee attack speed for 12 sec on a 3 min cooldown, 10 focus. Its
// 50% dodge changes nothing the sim measures.
func (hp *HunterPet) newForeverTrickstersDance() *core.Spell {
	rank, ok := foreverPetRankAt(hp.Owner.Level, foreverPetRanks[TrickstersDance])
	if !ok {
		return nil
	}
	speed := 1 + rank.base/100
	aura := hp.RegisterAura(core.Aura{
		Label:    "Trickster's Dance",
		ActionID: core.ActionID{SpellID: rank.spellID},
		Duration: time.Second * 12,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.MultiplyMeleeSpeed(sim, speed)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.MultiplyMeleeSpeed(sim, 1/speed)
		},
	})

	config := hp.foreverPetSpellConfig(rank, 0, 10, time.Minute*3)
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		aura.Activate(sim)
	}
	return hp.RegisterSpell(config)
}

// Charge and Dash share a cooldown category (44): Charge 25 sec, Dash and Dive 30 sec.
func (hp *HunterPet) closerTimer() *core.Timer {
	if hp.closerCD == nil {
		hp.closerCD = hp.NewTimer()
	}
	return hp.closerCD
}

// Dash and Dive: 80% movement speed for 15 sec at 50 and up, 20 focus.
func (hp *HunterPet) newForeverSprint(abilityType PetAbilityType) *core.Spell {
	rank, ok := foreverPetRankAt(hp.Owner.Level, foreverPetRanks[abilityType])
	if !ok {
		return nil
	}
	actionID := core.ActionID{SpellID: rank.spellID}
	aura := hp.RegisterAura(core.Aura{
		Label:    map[PetAbilityType]string{Dash: "Dash", Dive: "Dive"}[abilityType],
		ActionID: actionID,
		Duration: time.Second * 15,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.AddMoveSpeedModifier(&aura.ActionID, 1+rank.base/100)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.RemoveMoveSpeedModifier(&aura.ActionID)
		},
	})

	config := hp.foreverPetSpellConfig(rank, 0, 20, 0)
	config.Cast.CD = core.Cooldown{
		Timer:    hp.closerTimer(),
		Duration: time.Second * 30,
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		aura.Activate(sim)
	}
	return hp.RegisterSpell(config)
}

// The charge's travel: the sim takes the server's charge speed, 42 yd a sec, 6 times a run.
const foreverChargeSpeedMultiplier = 6.0

// The boar's Charge (27685 at 60: 490 attack power, 5 a level after; 35 focus) from 8 to 25 yd: the
// pet closes in, and its next melee hit within 4 sec carries the attack power. The root it puts on
// the enemy changes nothing the sim measures.
func (hp *HunterPet) newForeverCharge() *core.Spell {
	rank, ok := foreverPetRankAt(hp.Owner.Level, foreverPetRanks[Charge])
	if !ok {
		return nil
	}
	actionID := core.ActionID{SpellID: rank.spellID}
	ap := rank.amount(hp.Owner.Level)
	aura := hp.RegisterAura(core.Aura{
		Label:    "Charge",
		ActionID: actionID,
		Duration: time.Second * 4,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.AddStatDynamic(sim, stats.AttackPower, ap)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.AddStatDynamic(sim, stats.AttackPower, -ap)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.ProcMask.Matches(core.ProcMaskMelee) && result.Landed() {
				aura.Deactivate(sim)
			}
		},
	})

	config := hp.foreverPetSpellConfig(rank, 0, 35, 0)
	config.Cast.CD = core.Cooldown{
		Timer:    hp.closerTimer(),
		Duration: time.Second * 25,
	}
	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		return hp.DistanceFromTarget >= foreverChargeMinRange && hp.DistanceFromTarget <= foreverChargeMaxRange
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		aura.Activate(sim)
		hp.AddMoveSpeedModifier(&actionID, foreverChargeSpeedMultiplier)
		hp.moveToward(sim, core.MaxMeleeAttackDistance)
		hp.RemoveMoveSpeedModifier(&actionID)
	}
	return hp.RegisterSpell(config)
}
