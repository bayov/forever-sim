package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// The shaman's heals, for the leveling sim (docs/leveling_sim.md): we heal ourselves
// between and during pulls. Each one lands on the shaman, whatever target the rotation
// casts it at, because a shaman soloing mobs has no one else to heal. Chain Heal finds no
// one to jump to.
//
// Every value is the Forever client's (1.60.1.70338, SpellEffect, SpellLevels, SpellMisc and
// SpellPower on wago.tools). A rank heals its base at the level it's learned, plus its growth
// per level for 5 levels after, and the variance spreads it evenly around that:
// base * (1 - variance / 2) to base * (1 + variance / 2). Lightning Bolt's ranks read the same
// way give the sim's Lightning Bolt table to the point. The spell power coefficients follow the
// cast time (1.5 / 3.5 and so on), and Forever's downranking takes them down for a rank far
// below our level (downrankCoefficient).
type healRank struct {
	spellID              int32
	base, perLevel       float64
	variance             float64
	learnLevel, maxLevel int32
	coefficient          float64
	castTime             time.Duration
	manaCost             float64
}

// At is the rank's lowest and highest heal at the character's level.
func (r healRank) At(level int32) (float64, float64) {
	value := r.base + r.perLevel*float64(min(level, r.maxLevel)-r.learnLevel)
	return value * (1 - r.variance/2), value * (1 + r.variance/2)
}

var HealingWaveRanks = []healRank{
	{},
	{331, 39, 0.7, 0.25641027093, 1, 5, 0.429, 1500 * time.Millisecond, 25},
	{332, 71, 1.0, 0.19718310237, 6, 11, 0.571, 2000 * time.Millisecond, 45},
	{547, 131, 1.5, 0.18309858441, 12, 17, 0.714, 2500 * time.Millisecond, 80},
	{913, 256, 2.3, 0.16438356042, 18, 23, 0.857, 3000 * time.Millisecond, 155},
	{939, 342, 2.7, 0.15686275065, 24, 29, 0.857, 3000 * time.Millisecond, 200},
	{959, 484, 3.3, 0.14853195846, 32, 37, 0.857, 3000 * time.Millisecond, 265},
	{8005, 685, 3.9, 0.14303638041, 40, 45, 0.857, 3000 * time.Millisecond, 340},
	{10395, 972, 4.7, 0.13736264408, 48, 53, 0.857, 3000 * time.Millisecond, 440},
	{10396, 1344, 5.5, 0.13251365721, 56, 61, 0.857, 3000 * time.Millisecond, 560},
	{25357, 1615, 5.5, 0.1325648427, 60, 65, 0.857, 3000 * time.Millisecond, 620},
}

var LesserHealingWaveRanks = []healRank{
	{},
	{8004, 153, 1.7, 0.13793103397, 20, 25, 0.429, 1500 * time.Millisecond, 105},
	{8008, 224, 2.1, 0.12878787518, 28, 33, 0.429, 1500 * time.Millisecond, 145},
	{8010, 304, 2.5, 0.12256267667, 36, 41, 0.429, 1500 * time.Millisecond, 185},
	{10466, 426, 3.0, 0.11522634327, 44, 49, 0.429, 1500 * time.Millisecond, 235},
	{10467, 608, 3.6, 0.11077844352, 52, 57, 0.429, 1500 * time.Millisecond, 305},
	{10468, 820, 4.2, 0.10909090936, 60, 65, 0.429, 1500 * time.Millisecond, 380},
}

var ChainHealRanks = []healRank{
	{},
	{1064, 251, 2.5, 0.13953489065, 40, 45, 0.714, 2500 * time.Millisecond, 260},
	{10622, 335, 2.8, 0.13793103397, 46, 51, 0.714, 2500 * time.Millisecond, 315},
	{10623, 490, 3.3, 0.13220338523, 54, 59, 0.714, 2500 * time.Millisecond, 405},
}

// Riptide (the 31 point Restoration talent, 408521): an instant heal of 480 at level 40 plus
// 3.7 a level up to 48, and 89 every 3 sec for 15 sec. 245 Mana, 6 sec cooldown. While it's on
// a target, Chain Heal cast at that target heals it for 25% more.
var RiptideRank = healRank{408521, 480, 3.7, 0.10000000149, 40, 48, 0.214, 0, 245}

const (
	RiptideTickHealing     = 89
	RiptideTickCoefficient = 0.1
	RiptideChainHealBonus  = 0.25
)

func (shaman *Shaman) registerHealingSpells() {
	shaman.HealingWave = shaman.registerHealRanks(HealingWaveRanks, SpellCode_ShamanHealingWave)
	shaman.LesserHealingWave = shaman.registerHealRanks(LesserHealingWaveRanks, SpellCode_ShamanLesserHealingWave)
	shaman.ChainHeal = shaman.registerHealRanks(ChainHealRanks, SpellCode_ShamanChainHeal)
	shaman.registerRiptideSpell()
}

func (shaman *Shaman) registerHealRanks(ranks []healRank, spellCode int32) []*core.Spell {
	spells := make([]*core.Spell, len(ranks))
	for rank := 1; rank < len(ranks); rank++ {
		if ranks[rank].learnLevel <= shaman.Level {
			spells[rank] = shaman.RegisterSpell(shaman.newHealSpellConfig(ranks[rank], rank, spellCode))
		}
	}
	return spells
}

func (shaman *Shaman) newHealSpellConfig(r healRank, rank int, spellCode int32) core.SpellConfig {
	low, high := r.At(shaman.Level)

	castTime := r.castTime
	multiplier := shaman.purificationMultiplier()
	if spellCode == SpellCode_ShamanHealingWave {
		castTime -= shaman.improvedHealingWaveReduction()
		multiplier *= shaman.healingWayMultiplier()
	}

	return core.SpellConfig{
		SpellCode:   spellCode,
		ActionID:    core.ActionID{SpellID: r.spellID},
		SpellSchool: core.SpellSchoolNature,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellHealing,
		Flags:       SpellFlagShaman | core.SpellFlagHelpful | core.SpellFlagAPL,

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
		},

		DamageMultiplier: multiplier,
		ThreatMultiplier: 1,
		BonusCoefficient: shaman.downrankCoefficient(r.coefficient, int(r.learnLevel)),
		// Healing Focus: a chance for a hit not to push the heal back.
		PushbackReduction: shaman.healingFocusChance(),

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.DamageMultiplier *= shaman.riptideChainHealMultiplier(spell)
			spell.CalcAndDealHealing(sim, &shaman.Unit, sim.Roll(low, high), spell.OutcomeHealingCrit)
			spell.DamageMultiplier /= shaman.riptideChainHealMultiplier(spell)
		},
	}
}

func (shaman *Shaman) registerRiptideSpell() {
	if !shaman.Talents.Riptide || RiptideRank.learnLevel > shaman.Level {
		return
	}
	r := RiptideRank
	low, high := r.At(shaman.Level)

	shaman.Riptide = shaman.RegisterSpell(core.SpellConfig{
		SpellCode:   SpellCode_ShamanRiptide,
		ActionID:    core.ActionID{SpellID: r.spellID},
		SpellSchool: core.SpellSchoolNature,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellHealing,
		Flags:       SpellFlagShaman | core.SpellFlagHelpful | core.SpellFlagAPL,

		RequiredLevel: int(r.learnLevel),

		ManaCost: core.ManaCostOptions{
			FlatCost: r.manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    shaman.NewTimer(),
				Duration: time.Second * 6,
			},
		},

		DamageMultiplier: shaman.purificationMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: r.coefficient,

		Hot: core.DotConfig{
			SelfOnly: true,
			Aura: core.Aura{
				Label: "Riptide",
			},
			NumberOfTicks:    5,
			TickLength:       time.Second * 3,
			BonusCoefficient: RiptideTickCoefficient,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.SnapshotHeal(target, RiptideTickHealing, isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotHealing(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.CalcAndDealHealing(sim, &shaman.Unit, sim.Roll(low, high), spell.OutcomeHealingCrit)
			spell.SelfHot().Apply(sim)
		},
	})
}

// riptideChainHealMultiplier is Riptide's 25% more on a Chain Heal at a target that has it,
// which for us is ourselves.
func (shaman *Shaman) riptideChainHealMultiplier(spell *core.Spell) float64 {
	if spell.SpellCode != SpellCode_ShamanChainHeal || shaman.Riptide == nil || !shaman.Riptide.SelfHot().IsActive() {
		return 1
	}
	return 1 + RiptideChainHealBonus
}

// healRankIDs are the spell IDs of every rank of a heal, for its rank family.
func healRankIDs(ranks []healRank) []int32 {
	ids := make([]int32, 0, len(ranks)-1)
	for _, r := range ranks[1:] {
		ids = append(ids, r.spellID)
	}
	return ids
}

// TopHealRankID is the spell ID of a heal's top rank (HealingWaveRanks,
// LesserHealingWaveRanks) once a shaman of this level knows any rank of it, 0 before. A
// rotation that names it casts the highest rank the shaman knows (core.TopRankID).
func TopHealRankID(ranks []healRank, level int32) int32 {
	if ranks[1].learnLevel > level {
		return 0
	}
	return ranks[len(ranks)-1].spellID
}
