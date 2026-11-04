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
// below level 20. The shocks don't opt into Forever's downranking penalty (core.DownrankKind):
// the 2026-10-10 beta log hits at full coefficient (TestOrcShamanSpellPowerBeta).
//
// The November 2 development notes (post 9, for launch; not in client 1.60.1.70753) revert every
// shock's base damage to Classic Era's: the 1.15.9 client's base points, die and growth per
// level below. The spell power coefficients, costs and cooldowns stay Forever's.
// Rank 7 is 517 to 545 (was 293 to 309).
var EarthShockBaseDamage = [EarthShockRanks + 1][]float64{{0}, {19.5, 21.5}, {35.5, 37.5}, {65, 69}, {126, 134}, {235, 249}, {372, 394}, {517, 545}}
var EarthShockScaling = [EarthShockRanks + 1]core.RankScaling{{}, {9, 0.5}, {13, 0.7}, {19, 1}, {29, 1.4}, {41, 2}, {53, 2.6}, {60, 3.1}}
var EarthShockSpellCoef = [EarthShockRanks + 1]float64{0, 0.386, 0.386, 0.386, 0.386, 0.386, 0.386, 0.386}
var EarthShockManaCost = [EarthShockRanks + 1]float64{0, 30, 50, 85, 145, 240, 345, 450}
var EarthShockLevel = [EarthShockRanks + 1]int{0, 4, 8, 14, 24, 36, 48, 60}

// Report assumption sham_es7_tuning (user ruling 2026-10-15) took rank 7's base damage x1.08,
// as the ranks from level 36 kept 0.55 to 0.57 of Era's, as if not retuned yet. The November 2
// notes put every rank back at Era's (above), so the scale is 1.
const EarthShockRank7DamageScale = 1.0

func (shaman *Shaman) registerEarthShockSpell(shockTimer *core.Timer) {
	shaman.EarthShock = make([]*core.Spell, EarthShockRanks+1)

	for rank := 1; rank <= EarthShockRanks; rank++ {
		// Only the ranks the level has learned, so nothing below (a totem buff aura) is
		// built for a rank the shaman cannot cast.
		if shaman.KnowsRank(EarthShockSpellId[rank], int32(EarthShockLevel[rank])) {
			config := shaman.newEarthShockSpellConfig(rank, shockTimer)
			shaman.EarthShock[rank] = shaman.RegisterSpell(config)
		}
	}
}

// EarthShockBaseDamageAt is a rank's damage range at a level, before spell power.
func EarthShockBaseDamageAt(rank int, level int32) (float64, float64) {
	low := EarthShockScaling[rank].At(EarthShockBaseDamage[rank][0], level)
	high := EarthShockScaling[rank].At(EarthShockBaseDamage[rank][1], level)
	if rank == 7 {
		low *= EarthShockRank7DamageScale
		high *= EarthShockRank7DamageScale
	}
	return low, high
}

func (shaman *Shaman) newEarthShockSpellConfig(rank int, shockTimer *core.Timer) core.SpellConfig {
	spellId := EarthShockSpellId[rank]
	baseDamageLow, baseDamageHigh := EarthShockBaseDamageAt(rank, shaman.Level)
	spellCoeff := EarthShockSpellCoef[rank]
	manaCost := EarthShockManaCost[rank]
	level := EarthShockLevel[rank]

	spell := shaman.newShockSpellConfig(
		core.ActionID{SpellID: spellId},
		core.SpellSchoolNature,
		manaCost,
		shockTimer,
	)

	// Earth Shock is binary, so a target's resistance makes it miss in full more often. Under
	// Classic it never loses part of its damage.
	//
	// Under Forever a binary spell that lands still rolls a partial resist. The beta's Elder Cloud
	// Serpents fully resisted 3 of 9 Earth Shocks (logged RESIST) and took 30% off every one that
	// landed, as they did with our other Nature spells, where no orb or Lightning Bolt was fully
	// resisted (shaman_audit.md 5.2 and 7.8). wowsims/forever's logs show the same (78adad9607).
	spell.Flags |= core.SpellFlagBinary

	spell.SpellCode = SpellCode_ShamanEarthShock
	spell.RequiredLevel = level
	spell.Rank = rank

	spell.ThreatMultiplier = 2
	spell.BonusCoefficient = spellCoeff

	spell.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
		stormstrike := shaman.stormstrikeBonus(target)
		spell.DamageMultiplier *= stormstrike
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		spell.DamageMultiplier /= stormstrike
		shaman.spendStormstrike(sim, target, result)
		spell.DealDamage(sim, result)
	}

	return spell
}
