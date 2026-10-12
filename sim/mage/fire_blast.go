package mage

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

const FireBlastRanks = 7

var FireBlastSpellId = [FireBlastRanks + 1]int32{0, 2136, 2137, 2138, 8412, 8413, 10197, 10199}

// Damage at each rank's cap and its growth per level up to there, from the Forever client
// (1.60.1.70338). Rank 7 is 415 to 491 (Classic 446 to 524), and every rank has the full 42.9%
// coefficient (Classic 20.4% and 33.2% on ranks 1 and 2).
var FireBlastBaseDamage = [FireBlastRanks + 1][]float64{{0}, {26.57, 35.43}, {56.11, 69.89}, {95.83, 118.17}, {156.19, 187.81}, {224.71, 269.29}, {314.62, 373.38}, {415.41, 490.59}}
var FireBlastScaling = [FireBlastRanks + 1]core.RankScaling{{}, {11, 0.6}, {19, 1}, {27, 1.4}, {35, 1.8}, {43, 2.2}, {51, 2.6}, {59, 3}}
var FireBlastSpellCoeff = [FireBlastRanks + 1]float64{0, .429, .429, .429, .429, .429, .429, .429}
var FireBlastManaCost = [FireBlastRanks + 1]float64{0, 40, 75, 115, 165, 220, 280, 340}
var FireBlastLevel = [FireBlastRanks + 1]int{0, 6, 14, 22, 30, 38, 46, 54}

func (mage *Mage) registerFireBlastSpell() {
	mage.FireBlast = make([]*core.Spell, FireBlastRanks+1)
	cdTimer := mage.NewTimer()

	for rank := 1; rank <= FireBlastRanks; rank++ {
		config := mage.newFireBlastSpellConfig(rank, cdTimer)

		if config.RequiredLevel <= int(mage.Level) {
			mage.FireBlast[rank] = mage.GetOrRegisterSpell(config)
		}
	}
}

func (mage *Mage) newFireBlastSpellConfig(rank int, cdTimer *core.Timer) core.SpellConfig {

	spellId := FireBlastSpellId[rank]
	baseDamageLow := FireBlastScaling[rank].At(FireBlastBaseDamage[rank][0], mage.Level)
	baseDamageHigh := FireBlastScaling[rank].At(FireBlastBaseDamage[rank][1], mage.Level)
	spellCoeff := FireBlastSpellCoeff[rank]
	manaCost := FireBlastManaCost[rank]
	level := FireBlastLevel[rank]

	cooldown := time.Second * 8
	flags := SpellFlagMage | core.SpellFlagAPL

	return core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellId},
		SpellCode:   SpellCode_MageFireBlast,
		SpellSchool: core.SpellSchoolFire,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       flags,

		Rank:          rank,
		RequiredLevel: level,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer: cdTimer,
				// Wake of Fire takes 1 sec off per point. Its bonus after a kill needs a target that dies.
				Duration: cooldown - time.Second*time.Duration(mage.Talents.WakeOfFire),
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: spellCoeff,

		ExpectedInitialDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
			baseDamageCalc := (baseDamageLow + baseDamageHigh) / 2
			return spell.CalcDamage(sim, target, baseDamageCalc, spell.OutcomeExpectedMagicHitAndCrit)
		},
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		},
	}
}
