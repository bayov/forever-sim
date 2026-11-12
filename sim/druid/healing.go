package druid

import (
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// The druid's heals (docs/healing_druid.md): Healing Touch, Regrowth, Rejuvenation, Tranquility,
// Swiftmend, Wild Growth, Bloom and Nature's Swiftness. A heal cast at a player lands on that player
// (a healer's rotation aims it with Most Hurt, Lowest Health or a tank, docs/healing_sim.md), and one
// cast at a mob or at nothing lands on the druid.
//
// Every value is the Forever client's, build 1.60.1.70753 (the launch realms' 2026-10-30 hotfix
// written into the tables; wago.tools SpellEffect, SpellLevels, SpellMisc, SpellPower, SpellCastTimes,
// SpellDuration and SpellCooldowns). A rank heals its base at the level it's learned, plus its growth
// per level up to its max level, and the variance spreads a direct heal evenly around that, as the
// shaman's heals read the same tables (sim/shaman/healing.go). A HoT's tick has no variance. The spell
// power coefficients are the client's EffectBonusCoefficient. Forever's downranking takes them down for
// a rank far below our level in the engine (sim/core/downrank.go, from each rank's Rank and
// RequiredLevel, the direct part and the HoT alike).
//
// The talents are the client's spell modifiers, each on the spells its class mask names
// (healingTalentMultipliers). Percent modifiers of one kind multiply, as the cost ones do
// (shaman_audit.md 8.1).

const (
	SpellCode_DruidHealingTouch int32 = 100 + iota
	SpellCode_DruidRegrowth
	SpellCode_DruidRejuvenation
	SpellCode_DruidTranquility
	SpellCode_DruidSwiftmend
	SpellCode_DruidWildGrowth
)

// DruidHealing is what the druid's heals add to the druid.
type DruidHealing struct {
	HealingTouch []*DruidSpell
	Regrowth     []*DruidSpell
	Rejuvenation []*DruidSpell
	Tranquility  []*DruidSpell
	WildGrowth   []*DruidSpell
	Swiftmend    *DruidSpell
	Bloom        *core.Spell

	NaturesSwiftness     *DruidSpell
	NaturesSwiftnessAura *core.Aura

	wildGrowthCD *core.Timer
	// What items take off Swiftmend's cooldown (items_restoration.go).
	swiftmendCooldownCut time.Duration
}

type healRank struct {
	spellID              int32
	base, perLevel       float64
	variance             float64
	learnLevel, maxLevel int32
	coefficient          float64
	castTime             time.Duration
	manaCost             float64
	// A HoT's tick, and its coefficient a tick (Regrowth's second effect, Rejuvenation's, Wild
	// Growth's and Tranquility's only one).
	tick, tickCoefficient float64
}

// At is the rank's lowest and highest direct heal at the character's level.
func (r healRank) At(level int32) (float64, float64) {
	value := r.base + r.perLevel*float64(min(level, r.maxLevel)-r.learnLevel)
	return value * (1 - r.variance/2), value * (1 + r.variance/2)
}

// TickAt is the rank's tick at the character's level. A HoT's growth is in its base (perLevel)
// when it has no direct heal.
func (r healRank) TickAt(level int32) float64 {
	if r.base > 0 {
		return r.tick
	}
	return r.tick + r.perLevel*float64(min(level, r.maxLevel)-r.learnLevel)
}

// Healing Touch: 1.5 to 3.5 sec casts, coefficients by cast time.
var HealingTouchRanks = []healRank{
	{},
	{spellID: 5185, base: 44, perLevel: 0.8, variance: 0.31818181276, learnLevel: 1, maxLevel: 5, coefficient: 0.429, castTime: 1500 * time.Millisecond, manaCost: 25},
	{spellID: 5186, base: 100, perLevel: 1.3, variance: 0.23999999464, learnLevel: 8, maxLevel: 13, coefficient: 0.571, castTime: 2000 * time.Millisecond, manaCost: 55},
	{spellID: 5187, base: 201, perLevel: 1.9, variance: 0.21917808056, learnLevel: 14, maxLevel: 19, coefficient: 0.714, castTime: 2500 * time.Millisecond, manaCost: 110},
	{spellID: 5188, base: 362, perLevel: 2.7, variance: 0.20297029614, learnLevel: 20, maxLevel: 25, coefficient: 0.857, castTime: 3000 * time.Millisecond, manaCost: 190},
	{spellID: 5189, base: 546, perLevel: 3.5, variance: 0.19273301959, learnLevel: 26, maxLevel: 31, coefficient: 1, castTime: 3500 * time.Millisecond, manaCost: 280},
	{spellID: 6778, base: 707, perLevel: 4.0, variance: 0.18581907451, learnLevel: 32, maxLevel: 37, coefficient: 1, castTime: 3500 * time.Millisecond, manaCost: 350},
	{spellID: 8903, base: 898, perLevel: 4.5, variance: 0.17898832262, learnLevel: 38, maxLevel: 43, coefficient: 1, castTime: 3500 * time.Millisecond, manaCost: 425},
	{spellID: 9758, base: 1173, perLevel: 5.2, variance: 0.17364813387, learnLevel: 44, maxLevel: 49, coefficient: 1, castTime: 3500 * time.Millisecond, manaCost: 520},
	{spellID: 9888, base: 1516, perLevel: 5.9, variance: 0.16908212006, learnLevel: 50, maxLevel: 55, coefficient: 1, castTime: 3500 * time.Millisecond, manaCost: 630},
	{spellID: 9889, base: 1920, perLevel: 6.6, variance: 0.16504853964, learnLevel: 56, maxLevel: 61, coefficient: 1, castTime: 3500 * time.Millisecond, manaCost: 755},
	{spellID: 25297, base: 2332, perLevel: 7.3, variance: 0.16585759819, learnLevel: 60, maxLevel: 65, coefficient: 1, castTime: 3500 * time.Millisecond, manaCost: 840},
}

// Regrowth: a 2 sec cast, a direct heal at 0.286 and a HoT of 7 ticks every 3 sec at 0.071 a
// tick. Forever made it cheaper than 1.12 (rank 9 525 mana, 1.12 880) with a smaller HoT (142 a
// tick, 1.12 152).
var RegrowthRanks = []healRank{
	{},
	{spellID: 8936, base: 87, perLevel: 1.8, variance: 0.15384615958, learnLevel: 12, maxLevel: 17, manaCost: 70, tick: 13},
	{spellID: 8938, base: 164, perLevel: 2.5, variance: 0.13636364043, learnLevel: 18, maxLevel: 23, manaCost: 125, tick: 22},
	{spellID: 8939, base: 235, perLevel: 3.1, variance: 0.13229571283, learnLevel: 24, maxLevel: 29, manaCost: 170, tick: 32},
	{spellID: 8940, base: 311, perLevel: 3.6, variance: 0.1238937974, learnLevel: 30, maxLevel: 35, manaCost: 210, tick: 42},
	{spellID: 8941, base: 394, perLevel: 4.1, variance: 0.12064965069, learnLevel: 36, maxLevel: 41, manaCost: 250, tick: 52},
	{spellID: 9750, base: 504, perLevel: 4.7, variance: 0.11786372215, learnLevel: 42, maxLevel: 47, manaCost: 305, tick: 68},
	{spellID: 9856, base: 645, perLevel: 5.3, variance: 0.11386861652, learnLevel: 48, maxLevel: 53, manaCost: 370, tick: 88},
	{spellID: 9857, base: 817, perLevel: 6.0, variance: 0.11201866716, learnLevel: 54, maxLevel: 59, manaCost: 445, tick: 113},
	{spellID: 9858, base: 1021, perLevel: 6.8, variance: 0.10933081806, learnLevel: 60, maxLevel: 65, manaCost: 525, tick: 142},
}

const (
	RegrowthCastTime        = 2000 * time.Millisecond
	RegrowthCoefficient     = 0.286
	RegrowthTickCoefficient = 0.071
	RegrowthTicks           = 7
)

// Rejuvenation: an instant HoT of 4 ticks every 3 sec at 0.2 a tick. Forever cut it (rank 11 194 a
// tick, 1.12 222).
var RejuvenationRanks = []healRank{
	{},
	{spellID: 774, learnLevel: 4, maxLevel: 9, manaCost: 25, tick: 8},
	{spellID: 1058, learnLevel: 10, maxLevel: 15, manaCost: 40, tick: 12},
	{spellID: 1430, learnLevel: 16, maxLevel: 21, manaCost: 75, tick: 23},
	{spellID: 2090, learnLevel: 22, maxLevel: 27, manaCost: 105, tick: 32},
	{spellID: 2091, learnLevel: 28, maxLevel: 33, manaCost: 135, tick: 42},
	{spellID: 3627, learnLevel: 34, maxLevel: 39, manaCost: 160, tick: 51},
	{spellID: 8910, learnLevel: 40, maxLevel: 45, manaCost: 195, tick: 71},
	{spellID: 9839, learnLevel: 46, maxLevel: 51, manaCost: 235, tick: 94},
	{spellID: 9840, learnLevel: 52, maxLevel: 57, manaCost: 280, tick: 124},
	{spellID: 9841, learnLevel: 58, maxLevel: 63, manaCost: 335, tick: 161},
	{spellID: 25299, learnLevel: 60, maxLevel: 65, manaCost: 360, tick: 194},
}

const (
	RejuvenationTickCoefficient = 0.2
	RejuvenationTicks           = 4
)

// Tranquility: a 10 sec channel that heals every party member every 2 sec at 0.067 a tick, 5 min
// cooldown.
var TranquilityRanks = []healRank{
	{},
	{spellID: 740, perLevel: 0.6, learnLevel: 30, maxLevel: 36, manaCost: 375, tick: 87},
	{spellID: 8918, perLevel: 0.7, learnLevel: 40, maxLevel: 46, manaCost: 505, tick: 129},
	{spellID: 9862, perLevel: 0.9, learnLevel: 50, maxLevel: 56, manaCost: 695, tick: 196},
	{spellID: 9863, perLevel: 1.0, learnLevel: 60, maxLevel: 66, manaCost: 925, tick: 285},
}

const (
	TranquilityTickCoefficient = 0.067
	TranquilityTicks           = 5
	TranquilityCooldown        = 5 * time.Minute
)

// Wild Growth (the 31 point talent, rank 1; ranks 2 and 3 from the trainer): a 3 sec cast with a 6
// sec cooldown that puts a HoT on the target and its party, 5 players at most (its second effect),
// 7 ticks a second apart at 0.033 a tick.
var WildGrowthRanks = []healRank{
	{},
	{spellID: 408120, perLevel: 1.0, learnLevel: 40, maxLevel: 48, manaCost: 550, tick: 40},
	{spellID: 1238214, perLevel: 1.3, learnLevel: 50, maxLevel: 58, manaCost: 755, tick: 60},
	{spellID: 1238215, perLevel: 1.6, learnLevel: 60, maxLevel: 68, manaCost: 1050, tick: 97},
}

const (
	WildGrowthCastTime        = 3 * time.Second
	WildGrowthCooldown        = 6 * time.Second
	WildGrowthTickCoefficient = 0.033
	WildGrowthTicks           = 7
	WildGrowthTargets         = 5
)

// Wild Growth's heal "is applied quickly at first, and slows down" (the tooltip). The client leaves
// the shape to the server, so the sim takes the Wrath of the Lich King one (the first tick 6% over the
// average, each next one 2% less, the last 6% under), which keeps the total.
func wildGrowthTickMultiplier(tick int32) float64 {
	return 1 + (6-2*float64(tick-1))/100
}

// Swiftmend (18562): 20% of base mana, instant, 15 sec cooldown.
const (
	SwiftmendCost     = 0.2
	SwiftmendCooldown = 15 * time.Second
)

func init() {
	core.RegisterSpellRanks(healRankIDs(HealingTouchRanks)...)
	core.RegisterSpellRanks(healRankIDs(RegrowthRanks)...)
	core.RegisterSpellRanks(healRankIDs(RejuvenationRanks)...)
	core.RegisterSpellRanks(healRankIDs(TranquilityRanks)...)
	core.RegisterSpellRanks(healRankIDs(WildGrowthRanks)...)
}

func healRankIDs(ranks []healRank) []int32 {
	ids := make([]int32, 0, len(ranks)-1)
	for _, r := range ranks[1:] {
		ids = append(ids, r.spellID)
	}
	return ids
}

// RegisterHealingSpells registers the heals a Restoration druid casts.
func (druid *Druid) RegisterHealingSpells() {
	druid.HealingTouch = druid.registerHealRanks(HealingTouchRanks, druid.newHealingTouchConfig)
	druid.Regrowth = druid.registerHealRanks(RegrowthRanks, druid.newRegrowthConfig)
	druid.Rejuvenation = druid.registerHealRanks(RejuvenationRanks, druid.newRejuvenationConfig)
	druid.Tranquility = druid.registerHealRanks(TranquilityRanks, druid.newTranquilityConfig)
	if druid.Talents.WildGrowth {
		druid.WildGrowth = druid.registerHealRanks(WildGrowthRanks, druid.newWildGrowthConfig)
	}
	druid.registerSwiftmendSpell()
	druid.registerBloomSpell()
	druid.registerNaturesSwiftnessCD()
}

func (druid *Druid) registerHealRanks(ranks []healRank, newConfig func(r healRank, rank int) core.SpellConfig) []*DruidSpell {
	spells := make([]*DruidSpell, len(ranks))
	for rank := 1; rank < len(ranks); rank++ {
		if druid.KnowsRank(ranks[rank].spellID, ranks[rank].learnLevel) {
			spells[rank] = druid.RegisterSpell(Humanoid, newConfig(ranks[rank], rank))
		}
	}
	return spells
}

// healSpells are the heals the druid has registered but Swiftmend, the ones Nature's Grace's global
// cooldown cut names.
func (druid *Druid) healSpells() []*DruidSpell {
	return druid.balanceSpells(druid.HealingTouch, druid.Regrowth, druid.Rejuvenation, druid.Tranquility, druid.WildGrowth)
}

// healTarget is where a heal cast at target lands: the player it was cast at, or ourselves for a
// cast at a mob or at nothing.
func (druid *Druid) healTarget(target *core.Unit) *core.Unit {
	if target == nil || target.Type == core.EnemyUnit {
		return &druid.Unit
	}
	return target
}

// The healing talents, as multipliers on each heal (the spell modifiers' class masks):
//
//   - Gift of Nature (17104): 2% a point on Healing Touch, Regrowth, Tranquility and Swiftmend (its
//     first effect, all of the spell) and on the periodic part of Rejuvenation, Regrowth and Wild Growth
//     (its second). Each heal takes it once, Regrowth's HoT too.
//   - Genesis (1223081): 2% a point on the periodic part of Rejuvenation, Regrowth and Wild Growth, and
//     on Swiftmend (its second effect), not on Tranquility.
//   - Improved Rejuvenation (17111): 5% a point on Rejuvenation.
func (druid *Druid) giftOfNatureMultiplier() float64 {
	return 1 + 0.02*float64(druid.Talents.GiftOfNature)
}
func (druid *Druid) genesisHealingMultiplier() float64 {
	return 1 + 0.02*float64(druid.Talents.Genesis)
}

// newHotConfig is a HoT's dot: ticks of tick + coefficient * healing power at the cast, times our
// healing multipliers and periodic, its own. The healing power and the multipliers snapshot when it
// goes up; the crit rolls at each tick, as every Forever dot's does (ruleset.go). onTick runs after
// each tick's heal.
func (druid *Druid) newHotConfig(label string, ticks int32, tickLength time.Duration, tick, coefficient, periodic float64,
	onTick func(sim *core.Simulation, target *core.Unit, dot *core.Dot)) core.DotConfig {
	return core.DotConfig{
		Aura: core.Aura{
			Label: label,
		},
		NumberOfTicks:    ticks,
		TickLength:       tickLength,
		BonusCoefficient: coefficient,
		DamageMultiplier: periodic,
		OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
			dot.SnapshotBaseDamage = tick + dot.BonusCoefficient*dot.Spell.HealingPower(target)
			dot.SnapshotAttackerMultiplier = dot.Spell.CasterHealingMultiplier() * dot.DamageMultiplier
		},
		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
			if target.IsDead() {
				return
			}
			dot.CalcAndDealPeriodicSnapshotHealing(sim, target, dot.OutcomeTick)
			if onTick != nil {
				onTick(sim, target, dot)
			}
		},
	}
}

// expectedHot is what a HoT heals over its ticks at target, before crits.
func expectedHot(spell *core.Spell, target *core.Unit, ticks int32, tick, periodic float64) float64 {
	coefficient := spell.Hot(target).BonusCoefficient
	total := float64(ticks) * (tick + coefficient*spell.HealingPower(target)) * periodic
	return spell.ExpectedHealing(target, total)
}

// applyHot puts this rank's HoT on target, and takes our other ranks of the same heal off it: one
// druid's Rejuvenation (or Regrowth) replaces its last. Another druid's stays (development notes,
// build 70425: high ranks no longer block other druids' Rejuvenation).
func applyHot(ranks []*DruidSpell, spell *core.Spell, target *core.Unit, sim *core.Simulation) {
	for _, other := range ranks {
		if other != nil && other.Spell != spell {
			if hot := other.Hot(target); hot != nil && hot.IsActive() {
				hot.Cancel(sim)
			}
		}
	}
	spell.Hot(target).Apply(sim)
}

// bloomOnLastTick heals with Bloom when a Rejuvenation or Regrowth expires on its own, after its last
// tick: not when it's recast, consumed by Swiftmend (outside Forever) or ends with its target.
func (druid *Druid) bloomOnLastTick(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	if druid.Bloom != nil && dot.TickCount >= dot.NumberOfTicks {
		druid.Bloom.Cast(sim, target)
	}
}

func (druid *Druid) newHealingTouchConfig(r healRank, rank int) core.SpellConfig {
	low, high := r.At(druid.Level)
	// Naturalist: 0.1 sec a point off the cast. Tranquil Spirit: 2% a point off the cost.
	castTime := r.castTime - 100*time.Millisecond*time.Duration(druid.Talents.Naturalist)

	var spell *core.Spell
	return core.SpellConfig{
		SpellCode:   SpellCode_DruidHealingTouch,
		ActionID:    core.ActionID{SpellID: r.spellID},
		SpellSchool: core.SpellSchoolNature,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellHealing,
		Flags:       SpellFlagOmen | core.SpellFlagHelpful | core.SpellFlagAPL,

		RequiredLevel: int(r.learnLevel),
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: r.manaCost,
			Factor:   1 - 0.02*float64(druid.Talents.TranquilSpirit),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: castTime,
			},
		},

		DamageMultiplier:  druid.giftOfNatureMultiplier(),
		ThreatMultiplier:  1,
		BonusCoefficient:  r.coefficient,
		PushbackReduction: druid.naturesFocusChance(),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			target = druid.healTarget(target)
			if target.IsDead() {
				return
			}
			spell.CalcAndDealHealing(sim, target, sim.Roll(low, high), spell.OutcomeHealingCrit)
		},
		ExpectedHealing: func(target *core.Unit) float64 {
			if spell == nil {
				spell = druid.GetSpell(core.ActionID{SpellID: r.spellID})
			}
			return spell.ExpectedHealing(druid.healTarget(target), (low+high)/2)
		},
	}
}

func (druid *Druid) newRegrowthConfig(r healRank, rank int) core.SpellConfig {
	low, high := r.At(druid.Level)
	tick := r.TickAt(druid.Level)
	tickCoefficient := RegrowthTickCoefficient
	// Nature's Splendor: 6 sec more, 2 more ticks.
	ticks := int32(RegrowthTicks + core.TernaryInt32(druid.Talents.NaturesSplendor, 2, 0))
	periodic := druid.genesisHealingMultiplier()

	var spell *core.Spell
	return core.SpellConfig{
		SpellCode:   SpellCode_DruidRegrowth,
		ActionID:    core.ActionID{SpellID: r.spellID},
		SpellSchool: core.SpellSchoolNature,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellHealing,
		Flags:       SpellFlagOmen | core.SpellFlagHelpful | core.SpellFlagAPL,

		RequiredLevel: int(r.learnLevel),
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: r.manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: RegrowthCastTime,
			},
		},

		// Improved Regrowth: 10% crit a point (aura 107, misc 7, on Regrowth). The spell's crit chance
		// is what its ticks roll too.
		BonusCritRating: 10 * float64(druid.Talents.ImprovedRegrowth) * core.SpellCritRatingPerCritChance,

		DamageMultiplier:  druid.giftOfNatureMultiplier(),
		ThreatMultiplier:  1,
		BonusCoefficient:  RegrowthCoefficient,
		PushbackReduction: druid.naturesFocusChance(),

		Hot: druid.newHotConfig("Regrowth "+strconv.Itoa(rank), ticks, 3*time.Second, tick, tickCoefficient, periodic, druid.bloomOnLastTick),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			target = druid.healTarget(target)
			if target.IsDead() {
				return
			}
			spell.CalcAndDealHealing(sim, target, sim.Roll(low, high), spell.OutcomeHealingCrit)
			applyHot(druid.Regrowth, spell, target, sim)
		},
		// The direct heal: what lands when the cast ends, as HealComm counted it.
		ExpectedHealing: func(target *core.Unit) float64 {
			if spell == nil {
				spell = druid.GetSpell(core.ActionID{SpellID: r.spellID})
			}
			return spell.ExpectedHealing(druid.healTarget(target), (low+high)/2)
		},
	}
}

func (druid *Druid) newRejuvenationConfig(r healRank, rank int) core.SpellConfig {
	tick := r.TickAt(druid.Level)
	tickCoefficient := RejuvenationTickCoefficient
	// Nature's Splendor: 3 sec more, one more tick.
	ticks := int32(RejuvenationTicks + core.TernaryInt32(druid.Talents.NaturesSplendor, 1, 0))
	periodic := druid.giftOfNatureMultiplier() * druid.genesisHealingMultiplier() * (1 + 0.05*float64(druid.Talents.ImprovedRejuvenation))

	var spell *core.Spell
	return core.SpellConfig{
		SpellCode:   SpellCode_DruidRejuvenation,
		ActionID:    core.ActionID{SpellID: r.spellID},
		SpellSchool: core.SpellSchoolNature,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellHealing,
		Flags:       SpellFlagOmen | core.SpellFlagHelpful | core.SpellFlagAPL,

		RequiredLevel: int(r.learnLevel),
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: r.manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Hot: druid.newHotConfig("Rejuvenation "+strconv.Itoa(rank), ticks, 3*time.Second, tick, tickCoefficient, periodic, druid.bloomOnLastTick),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			target = druid.healTarget(target)
			if target.IsDead() {
				return
			}
			applyHot(druid.Rejuvenation, spell, target, sim)
		},
		// The whole HoT.
		ExpectedHealing: func(target *core.Unit) float64 {
			if spell == nil {
				spell = druid.GetSpell(core.ActionID{SpellID: r.spellID})
			}
			return expectedHot(spell, druid.healTarget(target), ticks, tick, periodic)
		},
	}
}

// partyOf is the living players of target's party, target first.
func (druid *Druid) partyOf(target *core.Unit) []*core.Unit {
	units := []*core.Unit{target}
	agent := druid.Env.Raid.GetPlayerFromUnit(target)
	if agent == nil {
		return units
	}
	for _, a := range agent.GetCharacter().Party.Players {
		if unit := &a.GetCharacter().Unit; unit != target && !unit.IsDead() {
			units = append(units, unit)
		}
	}
	return units
}

func (druid *Druid) newTranquilityConfig(r healRank, rank int) core.SpellConfig {
	tick := r.TickAt(druid.Level)
	// Improved Tranquility: 30% a point off the cooldown (its threat cut is threat only).
	cooldown := time.Duration(float64(TranquilityCooldown) * (1 - 0.3*float64(druid.Talents.ImprovedTranquility)))

	var spell *core.Spell
	return core.SpellConfig{
		SpellCode:   SpellCode_DruidTranquility,
		ActionID:    core.ActionID{SpellID: r.spellID},
		SpellSchool: core.SpellSchoolNature,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellHealing,
		Flags:       SpellFlagOmen | core.SpellFlagHelpful | core.SpellFlagAPL | core.SpellFlagChanneled,

		RequiredLevel: int(r.learnLevel),
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: r.manaCost,
			Factor:   1 - 0.02*float64(druid.Talents.TranquilSpirit),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: cooldown,
			},
		},

		DamageMultiplier: druid.giftOfNatureMultiplier(),
		ThreatMultiplier: 1 - 0.5*float64(druid.Talents.ImprovedTranquility),
		BonusCoefficient: TranquilityTickCoefficient,

		// The channel, on us: each tick heals every living member of our party (the area aura's
		// "all nearby party members"; everyone is in reach).
		Hot: core.DotConfig{
			IsAOE: true,
			Aura: core.Aura{
				Label: "Tranquility " + strconv.Itoa(rank),
			},
			NumberOfTicks: TranquilityTicks,
			TickLength:    2 * time.Second,
			OnTick: func(sim *core.Simulation, _ *core.Unit, dot *core.Dot) {
				for _, unit := range druid.partyOf(&druid.Unit) {
					if unit.IsDead() {
						continue
					}
					result := dot.Spell.CalcHealing(sim, unit, tick, dot.OutcomeTick)
					dot.Spell.DealPeriodicHealing(sim, result)
				}
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.AOEHot().Apply(sim)
		},
		// The whole channel at one player.
		ExpectedHealing: func(target *core.Unit) float64 {
			if spell == nil {
				spell = druid.GetSpell(core.ActionID{SpellID: r.spellID})
			}
			return TranquilityTicks * spell.ExpectedHealing(druid.healTarget(target), tick)
		},
	}
}

func (druid *Druid) newWildGrowthConfig(r healRank, rank int) core.SpellConfig {
	tick := r.TickAt(druid.Level)
	tickCoefficient := WildGrowthTickCoefficient
	periodic := druid.giftOfNatureMultiplier() * druid.genesisHealingMultiplier()
	// Naturalist: 0.1 sec a point off the cast (since the 2026-10-30 hotfix).
	castTime := WildGrowthCastTime - 100*time.Millisecond*time.Duration(druid.Talents.Naturalist)

	hot := druid.newHotConfig("Wild Growth "+strconv.Itoa(rank), WildGrowthTicks, time.Second, tick, tickCoefficient, periodic, nil)
	hot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		if target.IsDead() {
			return
		}
		multiplier := wildGrowthTickMultiplier(dot.TickCount)
		dot.SnapshotAttackerMultiplier *= multiplier
		dot.CalcAndDealPeriodicSnapshotHealing(sim, target, dot.OutcomeTick)
		dot.SnapshotAttackerMultiplier /= multiplier
	}

	var spell *core.Spell
	return core.SpellConfig{
		SpellCode:   SpellCode_DruidWildGrowth,
		ActionID:    core.ActionID{SpellID: r.spellID},
		SpellSchool: core.SpellSchoolNature,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellHealing,
		Flags:       SpellFlagOmen | core.SpellFlagHelpful | core.SpellFlagAPL,

		RequiredLevel: int(r.learnLevel),
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: r.manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: castTime,
			},
			CD: core.Cooldown{
				Timer:    druid.wildGrowthTimer(),
				Duration: WildGrowthCooldown,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Hot: hot,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			target = druid.healTarget(target)
			if target.IsDead() {
				return
			}
			units := druid.partyOf(target)
			for _, unit := range units[:min(len(units), WildGrowthTargets)] {
				applyHot(druid.WildGrowth, spell, unit, sim)
			}
		},
		// The whole HoT on the target alone.
		ExpectedHealing: func(target *core.Unit) float64 {
			if spell == nil {
				spell = druid.GetSpell(core.ActionID{SpellID: r.spellID})
			}
			return expectedHot(spell, druid.healTarget(target), WildGrowthTicks, tick, periodic)
		},
	}
}

// Wild Growth's ranks share one cooldown.
func (druid *Druid) wildGrowthTimer() *core.Timer {
	if druid.wildGrowthCD == nil {
		druid.wildGrowthCD = druid.NewTimer()
	}
	return druid.wildGrowthCD
}

// swiftmendHot is the Rejuvenation or Regrowth of ours on target Swiftmend reads: the one with the
// least time left (the 1.12 and later rule; the client leaves it to the server).
func (druid *Druid) swiftmendHot(target *core.Unit) *core.Dot {
	var best *core.Dot
	for _, spell := range druid.balanceSpells(druid.Rejuvenation, druid.Regrowth) {
		if hot := spell.Hot(target); hot != nil && hot.IsActive() {
			if best == nil || hot.ExpiresAt() < best.ExpiresAt() {
				best = hot
			}
		}
	}
	return best
}

// Swiftmend heals "for an amount equal to the full duration of the periodic effect of the consumed
// spell" (the Forever tooltip): the HoT's tick as it snapshot (its base and healing power) times all
// its ticks, Nature's Splendor's too, with Swiftmend's own modifiers (Gift of Nature and Genesis name it
// in their class masks) in place of the HoT's. Under Forever it no longer consumes the HoT: "Redesigned
// to not consume the Rejuvenation or Regrowth effect on the target" (news 24301515; dev notes post 9,
// 2026-11-04), so the HoT ticks on and still Blooms when it ends. Client 70755's tooltip still says
// "consumed"; the notes win (standards 37). Outside Forever it is consumed, as in 1.12.
func (druid *Druid) registerSwiftmendSpell() {
	if !druid.Talents.Swiftmend {
		return
	}
	druid.Swiftmend = druid.RegisterSpell(Humanoid, core.SpellConfig{
		SpellCode:   SpellCode_DruidSwiftmend,
		ActionID:    core.ActionID{SpellID: 18562},
		SpellSchool: core.SpellSchoolNature,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellHealing,
		Flags:       SpellFlagOmen | core.SpellFlagHelpful | core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			BaseCost: SwiftmendCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: SwiftmendCooldown - druid.swiftmendCooldownCut,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return druid.swiftmendHot(druid.healTarget(target)) != nil
		},

		DamageMultiplier: druid.giftOfNatureMultiplier() * druid.genesisHealingMultiplier(),
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			target = druid.healTarget(target)
			hot := druid.swiftmendHot(target)
			if target.IsDead() || hot == nil {
				return
			}
			heal := hot.SnapshotBaseDamage * float64(hot.NumberOfTicks)
			if !druid.Env.IsForever() {
				hot.Cancel(sim)
			}
			spell.CalcAndDealHealing(sim, target, heal, spell.OutcomeHealingCrit)
		},
		ExpectedHealing: func(target *core.Unit) float64 {
			target = druid.healTarget(target)
			hot := druid.swiftmendHot(target)
			if hot == nil {
				return 0
			}
			return druid.Swiftmend.ExpectedHealing(target, hot.SnapshotBaseDamage*float64(hot.NumberOfTicks))
		},
	})
}

// Bloom (1328621, the 2026-10-30 hotfix's talent in Gift of the Earthmother's place): a Rejuvenation or
// Regrowth that expires on its own heals its target for 100% of our Spirit (1328622, no class mask, so
// no healing talent reaches it; it can crit, as any heal does). The 50% on a dispel doesn't come up.
func (druid *Druid) registerBloomSpell() {
	if !druid.Talents.Bloom {
		return
	}
	druid.Bloom = druid.Unit.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 1328622},
		SpellSchool: core.SpellSchoolNature,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellHealing,
		Flags:       core.SpellFlagHelpful | core.SpellFlagPassiveSpell | core.SpellFlagNoOnCastComplete,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			if target.IsDead() {
				return
			}
			spell.CalcAndDealHealing(sim, target, druid.GetStat(stats.Spirit), spell.OutcomeHealingCrit)
		},
	})
}

// Nature's Swiftness (17116): the next Healing Touch, Regrowth, Wild Growth or Wrath is instant (the
// other Nature spells its class mask names the sim doesn't cast), off the global cooldown,
// 3 min cooldown from when the buff is used. A rotation that casts it controls it; one that doesn't
// has it used with the other cooldowns, on the next of those casts.
func (druid *Druid) registerNaturesSwiftnessCD() {
	if !druid.Talents.NaturesSwiftness {
		return
	}
	actionID := core.ActionID{SpellID: 17116}
	cdTimer := druid.NewTimer()
	cd := 3 * time.Minute

	// Hotfix push 113885 (SpellEffect 695239's class mask) adds Wild Growth: "Nature's Swiftness
	// now also makes Wild Growth instant" (Blizzard hotfixes of 2026-11-10).
	affected := func(spell *core.Spell) bool {
		switch spell.SpellCode {
		case SpellCode_DruidHealingTouch, SpellCode_DruidRegrowth, SpellCode_DruidWrath, SpellCode_DruidWildGrowth:
			return true
		}
		return false
	}
	var affectedSpells []*core.Spell
	druid.NaturesSwiftnessAura = druid.RegisterAura(core.Aura{
		Label:    "Natures Swiftness",
		ActionID: actionID,
		Duration: core.NeverExpires,
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			affectedSpells = core.FilterSlice(druid.Spellbook, affected)
		},
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			core.Each(affectedSpells, func(spell *core.Spell) { spell.CastTimeMultiplier -= 1 })
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			core.Each(affectedSpells, func(spell *core.Spell) { spell.CastTimeMultiplier += 1 })
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if affected(spell) {
				aura.Deactivate(sim)
				cdTimer.Set(sim.CurrentTime + cd)
				druid.UpdateMajorCooldowns()
			}
		},
	})

	druid.NaturesSwiftness = druid.RegisterSpell(Any, core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagAPL,
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    cdTimer,
				Duration: cd,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			druid.NaturesSwiftnessAura.Activate(sim)
		},
	})

	druid.AddMajorCooldown(core.MajorCooldown{
		Spell: druid.NaturesSwiftness.Spell,
		Type:  core.CooldownTypeDPS,
	})
}
