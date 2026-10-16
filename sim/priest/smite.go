package priest

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

const SmiteRanks = 8

var SmiteSpellId = [SmiteRanks + 1]int32{0, 585, 591, 598, 984, 1004, 6060, 10933, 10934}

// Damage at each rank's cap, growth per level and coefficient from the Forever client
// (1.60.1.70338, wago.tools). Forever cut Smite to less than half (rank 8: 167 to 186,
// Classic 384 to 429), gave the low ranks full coefficients and lets Shadowform cast it.
var SmiteBaseDamage = [SmiteRanks + 1][]float64{{0}, {13.43, 17.57}, {25.45, 31.55}, {44.22, 50.78}, {58.96, 68.04}, {80.62, 91.38}, {93.34, 105.66}, {123.74, 139.26}, {166.62, 186.38}}
var SmiteScaling = [SmiteRanks + 1]core.RankScaling{{}, {6, 0.1}, {11, 0.3}, {19, 0.5}, {27, 0.7}, {35, 1}, {43, 1.1}, {51, 1.1}, {59, 1.3}}
var SmiteSpellCoef = [SmiteRanks + 1]float64{0, 0.429, 0.571, 0.714, 0.714, 0.714, 0.714, 0.714, 0.714}
var SmiteCastTime = [SmiteRanks + 1]int{0, 1500, 2000, 2500, 2500, 2500, 2500, 2500, 2500}
var SmiteManaCost = [SmiteRanks + 1]float64{0, 20, 30, 60, 95, 140, 185, 230, 280}
var SmiteLevel = [SmiteRanks + 1]int{0, 1, 6, 14, 22, 30, 38, 46, 54}

// Report assumption priest_smite8_tuning, user ruling 2026-10-15: rank 8's base damage x1.09,
// half the raise that would bring it to the retuned low ranks' 0.51 of Era (it keeps 0.41 to 0.43).
const SmiteRank8DamageScale = 1.09

func (priest *Priest) registerSmiteSpell() {
	priest.Smite = make([]*core.Spell, SmiteRanks+1)

	for rank := 1; rank <= SmiteRanks; rank++ {
		config := priest.getSmiteBaseConfig(rank)

		if config.RequiredLevel <= int(priest.Level) {
			priest.Smite[rank] = priest.GetOrRegisterSpell(config)
		}
	}
}

// SmiteBaseDamageAt is a rank's damage range at a level, before spell power.
func SmiteBaseDamageAt(rank int, level int32) (float64, float64) {
	low := SmiteScaling[rank].At(SmiteBaseDamage[rank][0], level)
	high := SmiteScaling[rank].At(SmiteBaseDamage[rank][1], level)
	if rank == 8 {
		low *= SmiteRank8DamageScale
		high *= SmiteRank8DamageScale
	}
	return low, high
}

func (priest *Priest) getSmiteBaseConfig(rank int) core.SpellConfig {
	spellId := SmiteSpellId[rank]
	baseDamageLow, baseDamageHigh := SmiteBaseDamageAt(rank, priest.Level)
	spellCoeff := SmiteSpellCoef[rank]
	castTime := SmiteCastTime[rank]
	manaCost := SmiteManaCost[rank]
	level := SmiteLevel[rank]

	return core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellId},
		SpellCode:   SpellCode_PriestSmite,
		SpellSchool: core.SpellSchoolHoly,
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
				CastTime: time.Millisecond*time.Duration(castTime) - time.Millisecond*100*time.Duration(priest.Talents.DivineFury),
			},
		},

		BonusCoefficient: priest.downrankCoefficient(spellCoeff, level),

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
			multiplier := priest.PowerInLightMultiplier(target)
			spell.DamageMultiplier *= multiplier
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			spell.DamageMultiplier /= multiplier
		},
	}
}
