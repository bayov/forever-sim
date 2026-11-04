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
// below level 20. The shocks don't opt into Forever's downranking penalty (core.DownrankKind):
// the 2026-10-10 beta log hits at full coefficient (TestOrcShamanSpellPowerBeta).
//
// The November 2 development notes (post 9, for launch; not in client 1.60.1.70753) revert every
// shock's base damage to Classic Era's: the 1.15.9 client's base points, die and growth per
// level below. The spell power coefficients, costs and cooldowns stay Forever's.
// Rank 4 is 492 to 520 at 60 (was 279 to 295).
var FrostShockBaseDamage = [FrostShockRanks + 1][]float64{{0}, {95, 101}, {215.5, 229.5}, {345.5, 365.5}, {492, 520}}
var FrostShockScaling = [FrostShockRanks + 1]core.RankScaling{{}, {25, 1.2}, {39, 1.9}, {51, 2.5}, {60, 3}}
var FrostShockSpellCoef = [FrostShockRanks + 1]float64{0, 0.386, 0.386, 0.386, 0.386}
var FrostShockManaCost = [FrostShockRanks + 1]float64{0, 115, 225, 325, 430}
var FrostShockLevel = [FrostShockRanks + 1]int{0, 20, 34, 46, 58}

func (shaman *Shaman) registerFrostShockSpell(shockTimer *core.Timer) {
	shaman.FrostShock = make([]*core.Spell, FrostShockRanks+1)

	for rank := 1; rank <= FrostShockRanks; rank++ {
		// Only the ranks the level has learned, so nothing below (a totem buff aura) is
		// built for a rank the shaman cannot cast.
		if shaman.KnowsRank(FrostShockSpellId[rank], int32(FrostShockLevel[rank])) {
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

	// Under Forever Frost Shock is binary, like Earth Shock, Frostbolt and Cone of Cold: the
	// client's row (10473) carries the slow as well as the damage. A binary spell that lands
	// still rolls a partial resist under Forever (earth_shock.go).
	if shaman.Env.IsForever() {
		spell.Flags |= core.SpellFlagBinary
	}

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
