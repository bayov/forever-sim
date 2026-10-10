package shaman

import (
	"github.com/wowsims/classic/sim/core"
)

const EarthShockRanks = 7

var EarthShockSpellId = [EarthShockRanks + 1]int32{0, 8042, 8044, 8045, 8046, 10412, 10413, 10414}

// Damage at each rank's cap, its growth per level up to there, and the spell power
// coefficient, from the Forever client data that https://foreverchanges.pro reads
// (downrank calculator). Forever cut most rank damage well below 1.12, and it gave the
// low ranks close to full coefficients instead of the 1.12 penalty for spells learned
// below level 20.
var EarthShockBaseDamage = [EarthShockRanks + 1][]float64{{0}, {19.36, 21.64}, {35.39, 37.61}, {51.77, 55.23}, {83.69, 89.31}, {134.32, 142.68}, {206.66, 219.34}, {293.07, 308.93}}
var EarthShockScaling = [EarthShockRanks + 1]core.RankScaling{{}, {9, 0.5}, {13, 0.7}, {19, 0.9}, {29, 1.1}, {41, 1.3}, {53, 1.6}, {60, 1.9}}
var EarthShockSpellCoef = [EarthShockRanks + 1]float64{0, 0.386, 0.386, 0.386, 0.386, 0.386, 0.386, 0.386}
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

	// Earth Shock is binary under Classic, so a target's resistance makes it miss in full more
	// often instead of taking part of its damage.
	//
	// Under Forever the beta's Elder Cloud Serpents took 30% off every Earth Shock that landed on
	// them, as they did with our other Nature spells (shaman_audit.md 7.8). So it can be partly
	// resisted like Lightning Bolt, and the level part of partial resists counts for it too (5.2).
	if !shaman.Env.IsForever() {
		spell.Flags |= core.SpellFlagBinary
	}

	spell.SpellCode = SpellCode_ShamanEarthShock
	spell.RequiredLevel = level
	spell.Rank = rank

	spell.ThreatMultiplier = 2
	spell.BonusCoefficient = spellCoeff

	spell.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
		stormstrike := shaman.spendStormstrike(sim, target)
		spell.DamageMultiplier *= stormstrike
		spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		spell.DamageMultiplier /= stormstrike
	}

	return spell
}
