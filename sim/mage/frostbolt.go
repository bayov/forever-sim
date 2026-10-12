package mage

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

const FrostboltRanks = 11

var FrostboltSpellId = [FrostboltRanks + 1]int32{0, 116, 205, 837, 7322, 8406, 8407, 8408, 10179, 10180, 10181, 25304}

// Damage at each rank's cap (or at 60), its growth per level up to there, and the spell power
// coefficient, from the Forever client (1.60.1.70338, wago.tools SpellEffect and SpellLevels).
// Rank 11 is 457 to 493 (Classic 515 to 555), and the low ranks have close to full coefficients.
var FrostboltBaseDamage = [FrostboltRanks + 1][]float64{{0, 0}, {18.76, 20.84}, {28.37, 32.03}, {45.33, 50.67}, {69.45, 76.95}, {104.24, 114.16}, {134.61, 146.99}, {181.03, 196.97}, {243.21, 263.19}, {305.42, 331.38}, {382.44, 412.76}, {457.24, 492.76}}
var FrostboltScaling = [FrostboltRanks + 1]core.RankScaling{{}, {8, 0.2}, {12, 0.3}, {18, 0.5}, {24, 0.8}, {30, 1.3}, {36, 1.7}, {42, 2}, {48, 2.3}, {54, 2.6}, {60, 2.9}, {}}
var FrostboltSpellCoeff = [FrostboltRanks + 1]float64{0, .407, .489, .597, .706, .814, .814, .814, .814, .814, .814, .814}
var FrostboltCastTime = [FrostboltRanks + 1]int32{0, 1500, 1800, 2200, 2600, 3000, 3000, 3000, 3000, 3000, 3000, 3000}
var FrostboltManaCost = [FrostboltRanks + 1]float64{0, 25, 35, 50, 65, 100, 130, 160, 195, 225, 260, 290}
var FrostboltLevel = [FrostboltRanks + 1]int{0, 4, 8, 14, 20, 26, 32, 38, 44, 50, 56, 60}

func (mage *Mage) registerFrostboltSpell() {
	mage.Frostbolt = make([]*core.Spell, FrostboltRanks+1)

	maxRank := core.TernaryInt(core.IncludeAQ, FrostboltRanks, FrostboltRanks-1)
	for rank := 1; rank <= maxRank; rank++ {
		config := mage.getFrostboltConfig(rank)

		if config.RequiredLevel <= int(mage.Level) {
			mage.Frostbolt[rank] = mage.GetOrRegisterSpell(config)
		}
	}
}

func (mage *Mage) getFrostboltConfig(rank int) core.SpellConfig {
	spellId := FrostboltSpellId[rank]
	baseDamageLow := FrostboltScaling[rank].At(FrostboltBaseDamage[rank][0], mage.Level)
	baseDamageHigh := FrostboltScaling[rank].At(FrostboltBaseDamage[rank][1], mage.Level)
	spellCoeff := FrostboltSpellCoeff[rank]
	castTime := FrostboltCastTime[rank]
	manaCost := FrostboltManaCost[rank]
	level := FrostboltLevel[rank]

	return core.SpellConfig{
		ActionID:     core.ActionID{SpellID: spellId},
		SpellCode:    SpellCode_MageFrostbolt,
		SpellSchool:  core.SpellSchoolFrost,
		DefenseType:  core.DefenseTypeMagic,
		ProcMask:     core.ProcMaskSpellDamage,
		Flags:        SpellFlagMage | SpellFlagChillSpell | core.SpellFlagBinary | core.SpellFlagAPL,
		MissileSpeed: 28,

		RequiredLevel: level,
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond*time.Duration(castTime) - time.Millisecond*100*time.Duration(mage.Talents.ImprovedFrostbolt),
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: spellCoeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
			result := mage.calcMageDamage(sim, target, baseDamage, spell)

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				if result.Landed() {
					spell.DealDamage(sim, result)
				}
			})
		},
	}
}
