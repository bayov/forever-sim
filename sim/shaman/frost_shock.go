package shaman

import (
	"github.com/wowsims/classic/sim/core"
)

const FrostShockRanks = 4

var FrostShockSpellId = [FrostShockRanks + 1]int32{0, 8056, 8058, 10472, 10473}

// Damage at each rank's cap, its growth per level up to there, and the spell power
// coefficient, from the Forever client data that https://foreverchanges.pro reads
// (downrank calculator). Forever cut most rank damage well below 1.12, and it gave the
// low ranks close to full coefficients instead of the 1.12 penalty for spells learned
// below level 20.
var FrostShockBaseDamage = [FrostShockRanks + 1][]float64{{0}, {68.2, 72.8}, {126.21, 134.79}, {189.8, 201.2}, {278.58, 294.62}}
var FrostShockScaling = [FrostShockRanks + 1]core.RankScaling{{}, {25, 0.9}, {39, 1.3}, {51, 1.5}, {60, 1.8}}
var FrostShockSpellCoef = [FrostShockRanks + 1]float64{0, 0.386, 0.386, 0.386, 0.386}
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
