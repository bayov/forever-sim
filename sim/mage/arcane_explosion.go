package mage

import (
	"github.com/wowsims/classic/sim/core"
)

const ArcaneExplosionRanks = 6

var ArcaneExplosionSpellId = [ArcaneExplosionRanks + 1]int32{0, 1449, 8437, 8438, 8439, 10201, 10202}

// Damage at each rank's cap and its growth per level up to there, from the Forever client
// (1.60.1.70338). Rank 6 is 239 to 258 (Classic 249 to 270), and rank 1 has the full 14.3%.
var ArcaneExplosionBaseDamage = [ArcaneExplosionRanks + 1][]float64{{0}, {32, 36}, {55.1, 60.9}, {94.6, 102.4}, {133.73, 145.27}, {180.73, 196.27}, {238.68, 258.32}}
var ArcaneExplosionScaling = [ArcaneExplosionRanks + 1]core.RankScaling{{}, {19, 0.4}, {27, 0.6}, {35, 0.9}, {43, 0.9}, {51, 1.1}, {59, 1.3}}
var ArcaneExplosionSpellCoeff = [ArcaneExplosionRanks + 1]float64{0, .143, .143, .143, .143, .143, .143}
var ArcaneExplosionManaCost = [ArcaneExplosionRanks + 1]float64{0, 75, 120, 185, 250, 315, 390}
var ArcaneExplosionLevel = [ArcaneExplosionRanks + 1]int{0, 14, 22, 30, 38, 46, 54}

func (mage *Mage) registerArcaneExplosionSpell() {
	mage.ArcaneExplosion = make([]*core.Spell, ArcaneExplosionRanks+1)

	for rank := 1; rank <= ArcaneExplosionRanks; rank++ {
		config := mage.newArcaneExplosionSpellConfig(rank)

		if config.RequiredLevel <= int(mage.Level) {
			mage.ArcaneExplosion[rank] = mage.GetOrRegisterSpell(config)
		}
	}
}

func (mage *Mage) newArcaneExplosionSpellConfig(rank int) core.SpellConfig {
	spellId := ArcaneExplosionSpellId[rank]
	baseDamageLow := ArcaneExplosionScaling[rank].At(ArcaneExplosionBaseDamage[rank][0], mage.Level)
	baseDamageHigh := ArcaneExplosionScaling[rank].At(ArcaneExplosionBaseDamage[rank][1], mage.Level)
	spellCoeff := ArcaneExplosionSpellCoeff[rank]
	manaCost := ArcaneExplosionManaCost[rank]
	level := ArcaneExplosionLevel[rank]

	return core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellId},
		SpellCode:   SpellCode_MageArcaneExplosion,
		SpellSchool: core.SpellSchoolArcane,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagMage | core.SpellFlagAPL,

		RequiredLevel: level,
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: spellCoeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			for _, aoeTarget := range sim.Encounter.TargetUnits {
				damage := sim.Roll(baseDamageLow, baseDamageHigh)
				spell.CalcAndDealDamage(sim, aoeTarget, damage, spell.OutcomeMagicCrit)
			}
		},
	}
}
