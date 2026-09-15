package shaman

import (
	"github.com/wowsims/classic/sim/core"
)

const FrostShockRanks = 4

var FrostShockSpellId = [FrostShockRanks + 1]int32{0, 8056, 8058, 10472, 10473}

// Damage at each rank's cap, and its growth per level up to there (1.12 spell data).
var FrostShockBaseDamage = [FrostShockRanks + 1][]float64{{0}, {95, 101}, {215.5, 229.5}, {345.5, 365.5}, {492, 520}}
var FrostShockScaling = [FrostShockRanks + 1]core.RankScaling{{}, {25, 1.2}, {39, 1.9}, {51, 2.5}, {60, 3}}
var FrostShockSpellCoef = [FrostShockRanks + 1]float64{0, .386, .386, .386, .386}
var FrostShockManaCost = [FrostShockRanks + 1]float64{0, 115, 225, 325, 430}
var FrostShockLevel = [FrostShockRanks + 1]int{0, 20, 34, 46, 58}

func (shaman *Shaman) registerFrostShockSpell(shockTimer *core.Timer) {
	shaman.FrostShock = make([]*core.Spell, FrostShockRanks+1)

	for rank := 1; rank <= FrostShockRanks; rank++ {
		// Only the ranks the level has learned, so nothing below (a totem buff aura) is
		// built for a rank the shaman cannot cast.
		if FrostShockLevel[rank] <= int(shaman.Level) {
			config := shaman.newFrostShockSpellConfig(rank, shockTimer)
			shaman.FrostShock[rank] = shaman.RegisterSpell(config)
		}
	}
}

func (shaman *Shaman) newFrostShockSpellConfig(rank int, shockTimer *core.Timer) core.SpellConfig {
	spellId := FrostShockSpellId[rank]
	baseDamageLow := FrostShockScaling[rank].At(FrostShockBaseDamage[rank][0], shaman.Level)
	baseDamageHigh := FrostShockScaling[rank].At(FrostShockBaseDamage[rank][1], shaman.Level)
	spellCoeff := FrostShockSpellCoef[rank]
	manaCost := FrostShockManaCost[rank]
	level := FrostShockLevel[rank]

	spell := shaman.newShockSpellConfig(
		core.ActionID{SpellID: spellId},
		core.SpellSchoolFrost,
		manaCost,
		shockTimer,
	)

	spell.SpellCode = SpellCode_ShamanFrostShock
	spell.RequiredLevel = level
	spell.Rank = rank
	spell.BonusCoefficient = spellCoeff

	spell.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
		spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
	}

	return spell
}
