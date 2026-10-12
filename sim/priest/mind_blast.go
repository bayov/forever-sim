package priest

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

const MindBlastRanks = 9

// Damage at each rank's cap (level 60 for rank 9), its growth per level and the
// coefficient, from the Forever client (1.60.1.70338, wago.tools SpellEffect and
// SpellLevels). Forever cut every rank by about 6% (rank 9: 477 to 503, Classic 508 to
// 537) and gave the low ranks the full 42.9% coefficient.
var MindBlastSpellId = [MindBlastRanks + 1]int32{0, 8092, 8102, 8103, 8104, 8105, 8106, 10945, 10946, 10947}
var MindBlastBaseDamage = [MindBlastRanks + 1][]float64{{0}, {39.95, 44.05}, {69.6, 75.4}, {102.83, 110.17}, {153.41, 162.59}, {197.63, 210.38}, {259.14, 275.86}, {325.1, 343.9}, {405.55, 428.45}, {476.93, 503.47}}
var MindBlastScaling = [MindBlastRanks + 1]core.RankScaling{{}, {15, 0.6}, {21, 0.9}, {27, 1.1}, {33, 1.4}, {39, 1.6}, {45, 1.9}, {51, 2.1}, {57, 2.4}, {60, 2.6}}
var MindBlastSpellCoef = [MindBlastRanks + 1]float64{0, .429, .429, .429, .429, .429, .429, .429, .429, .429}
var MindBlastManaCost = [MindBlastRanks + 1]float64{0, 50, 80, 110, 150, 185, 225, 265, 310, 350}
var MindBlastLevel = [MindBlastRanks + 1]int{0, 10, 16, 22, 28, 34, 40, 46, 52, 58}

func (priest *Priest) registerMindBlast() {
	priest.MindBlast = make([]*core.Spell, MindBlastRanks+1)
	cdTimer := priest.NewTimer()

	for rank := 1; rank <= MindBlastRanks; rank++ {
		config := priest.getMindBlastBaseConfig(rank, cdTimer)

		if config.RequiredLevel <= int(priest.Level) {
			priest.MindBlast[rank] = priest.GetOrRegisterSpell(config)
		}
	}
}

func (priest *Priest) getMindBlastBaseConfig(rank int, cdTimer *core.Timer) core.SpellConfig {
	spellId := MindBlastSpellId[rank]
	baseDamageLow := MindBlastScaling[rank].At(MindBlastBaseDamage[rank][0], priest.Level)
	baseDamageHigh := MindBlastScaling[rank].At(MindBlastBaseDamage[rank][1], priest.Level)
	spellCoeff := MindBlastSpellCoef[rank]
	castTime := time.Millisecond * 1500
	manaCost := MindBlastManaCost[rank]
	level := MindBlastLevel[rank]

	return core.SpellConfig{
		SpellCode:   SpellCode_PriestMindBlast,
		ActionID:    core.ActionID{SpellID: spellId},
		SpellSchool: core.SpellSchoolShadow,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagPriest | core.SpellFlagAPL,

		RequiredLevel: level,
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: castTime,
			},
			CD: core.Cooldown{
				Timer:    cdTimer,
				Duration: time.Second*8 - time.Millisecond*500*time.Duration(priest.Talents.ImprovedMindBlast),
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: spellCoeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcDamage(sim, target, sim.Roll(baseDamageLow, baseDamageHigh), spell.OutcomeMagicHitAndCrit)

			if result.Landed() {
				priest.AddShadowWeavingStack(sim, target)
			}
			spell.DealDamage(sim, result)
		},

		ExpectedInitialDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
			damage := (baseDamageLow + baseDamageHigh) / 2
			result := spell.CalcDamage(sim, target, damage, spell.OutcomeExpectedMagicHitAndCrit)
			return result
		},
	}
}
