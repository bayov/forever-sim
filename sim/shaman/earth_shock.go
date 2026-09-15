package shaman

import (
	"github.com/wowsims/classic/sim/core"
)

const EarthShockRanks = 7

var EarthShockSpellId = [EarthShockRanks + 1]int32{0, 8042, 8044, 8045, 8046, 10412, 10413, 10414}

// Damage at each rank's cap, and its growth per level up to there (1.12 spell data).
var EarthShockBaseDamage = [EarthShockRanks + 1][]float64{{0}, {19.5, 21.5}, {35.5, 37.5}, {65, 69}, {126, 134}, {235, 249}, {372, 394}, {517, 545}}
var EarthShockScaling = [EarthShockRanks + 1]core.RankScaling{{}, {9, .5}, {13, .7}, {19, 1}, {29, 1.4}, {41, 2}, {53, 2.6}, {60, 3.1}}
var EarthShockSpellCoef = [EarthShockRanks + 1]float64{0, .154, .212, .299, .386, .386, .386, .386}
var EarthShockManaCost = [EarthShockRanks + 1]float64{0, 30, 50, 85, 145, 240, 345, 450}
var EarthShockLevel = [EarthShockRanks + 1]int{0, 4, 8, 14, 24, 36, 48, 60}

func (shaman *Shaman) registerEarthShockSpell(shockTimer *core.Timer) {
	shaman.EarthShock = make([]*core.Spell, EarthShockRanks+1)

	for rank := 1; rank <= EarthShockRanks; rank++ {
		// Only the ranks the level has learned, so nothing below (a totem buff aura) is
		// built for a rank the shaman cannot cast.
		if EarthShockLevel[rank] <= int(shaman.Level) {
			config := shaman.newEarthShockSpellConfig(rank, shockTimer)
			shaman.EarthShock[rank] = shaman.RegisterSpell(config)
		}
	}
}

func (shaman *Shaman) newEarthShockSpellConfig(rank int, shockTimer *core.Timer) core.SpellConfig {
	spellId := EarthShockSpellId[rank]
	baseDamageLow := EarthShockScaling[rank].At(EarthShockBaseDamage[rank][0], shaman.Level)
	baseDamageHigh := EarthShockScaling[rank].At(EarthShockBaseDamage[rank][1], shaman.Level)
	spellCoeff := EarthShockSpellCoef[rank]
	manaCost := EarthShockManaCost[rank]
	level := EarthShockLevel[rank]

	spell := shaman.newShockSpellConfig(
		core.ActionID{SpellID: spellId},
		core.SpellSchoolNature,
		manaCost,
		shockTimer,
	)

	spell.Flags |= core.SpellFlagBinary

	spell.SpellCode = SpellCode_ShamanEarthShock
	spell.RequiredLevel = level
	spell.Rank = rank

	spell.ThreatMultiplier = 2
	spell.BonusCoefficient = spellCoeff

	spell.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
		spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
	}

	return spell
}
