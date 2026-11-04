package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

const LightningBoltRanks = 10

var LightningBoltSpellId = [LightningBoltRanks + 1]int32{0, 403, 529, 548, 915, 943, 6041, 10391, 10392, 15207, 15208}

// Damage at each rank's cap, its growth per level up to there, and the spell power
// coefficient, from the Forever client data that https://foreverchanges.pro reads
// (downrank calculator). Forever cut most rank damage well below 1.12, and it gave the
// low ranks close to full coefficients instead of the 1.12 penalty for spells learned
// below level 20. The 2026-09-24 beta build raised ranks 3 and 4 (35 and 50 at the learn
// level before, 45 and 56 now) so every rank is an upgrade over the one before.
//
// The 2026-10-08 beta build (1.60.1.70291) set ranks 1 to 5 again, so they grow more
// evenly into rank 6. Rank 5, the top rank at level 30, went from 75.5 to 84.5 on average
// at level 31, and rank 2 fell from 31.5 to 27.5.
//
// The November 2 development notes (post 9, for launch; not in client 1.60.1.70753, whose
// Lightning Bolt rows are 70707's) raise every rank: "Each rank's new base damage is roughly
// 78% of the base damage as the Classic Era version", the top rank by 80%. Ranks 4 to 10 are
// 0.78 of Era's (1.15.9 client: base points, die and growth per level), which puts rank 10 at
// 334 to 372 at 60 (was 190 to 212, +76%). Ranks 1 to 3 already deal more than 0.78 of Era's
// (1.07 / 0.96 / 0.92), so they keep the client's values; the notes give no number for them.
var LightningBoltBaseDamage = [LightningBoltRanks + 1][]float64{{0}, {14.39, 16.61}, {25.54, 29.46}, {42.70, 50.30}, {68.64, 78.00}, {102.18, 116.22}, {140.01, 157.17}, {183.69, 205.53}, {227.37, 253.89}, {278.85, 311.61}, {334.31, 371.75}}
var LightningBoltScaling = [LightningBoltRanks + 1]core.RankScaling{{}, {6, 0.1}, {13, 0.1}, {19, 0.3}, {25, 0.78}, {31, 0.94}, {37, 1.17}, {43, 1.33}, {49, 1.48}, {55, 1.64}, {60, 1.87}}
var LightningBoltSpellCoef = [LightningBoltRanks + 1]float64{0, 0.429, 0.571, 0.714, 0.714, 0.714, 0.714, 0.714, 0.714, 0.714, 0.714}

// Forever caps the cast at 2.5 sec from rank 4 up and takes about 20% off the mana
// cost of those ranks (wowhead Forever spell data).
var LightningBoltCastTime = [LightningBoltRanks + 1]int32{0, 1500, 2000, 2500, 2500, 2500, 2500, 2500, 2500, 2500, 2500}
var LightningBoltManaCost = [LightningBoltRanks + 1]float64{0, 15, 30, 45, 60, 85, 110, 135, 160, 190, 220}
var LightningBoltLevel = [LightningBoltRanks + 1]int{0, 1, 8, 14, 20, 26, 32, 38, 44, 50, 56}

// Report assumption sham_lb_31plus (user ruling 2026-10-15) took rank 10's base damage x1.12,
// because the ranks past level 38 fell from 0.61 of Era's damage to 0.44 as if not retuned yet.
// The November 2 notes retuned them (above), so the scale is 1.
const LightningBoltRank10DamageScale = 1.0

func (shaman *Shaman) registerLightningBoltSpell() {
	shaman.LightningBolt = make([]*core.Spell, LightningBoltRanks+1)
	shaman.LightningBoltOverload = make([]*core.Spell, LightningBoltRanks+1)

	for rank := 1; rank <= LightningBoltRanks; rank++ {
		// Only the ranks the level has learned, so nothing below (a totem buff aura) is
		// built for a rank the shaman cannot cast.
		if shaman.KnowsRank(LightningBoltSpellId[rank], int32(LightningBoltLevel[rank])) {
			config := shaman.newLightningBoltSpellConfig(rank)
			// The overload gets a config of its own so that the two casts share no state.
			shaman.LightningBoltOverload[rank] = shaman.registerOverloadSpell(shaman.newLightningBoltSpellConfig(rank))
			shaman.LightningBolt[rank] = shaman.RegisterSpell(config)
		}
	}
}

// LightningBoltBaseDamageAt is a rank's damage range at a level, before spell power.
func LightningBoltBaseDamageAt(rank int, level int32) (float64, float64) {
	low := LightningBoltScaling[rank].At(LightningBoltBaseDamage[rank][0], level)
	high := LightningBoltScaling[rank].At(LightningBoltBaseDamage[rank][1], level)
	if rank == 10 {
		low *= LightningBoltRank10DamageScale
		high *= LightningBoltRank10DamageScale
	}
	return low, high
}

// Ranks 4 to 10 take the notes' 0.78 of Era's damage under Forever only; outside Forever they deal
// Era's (Classic's) full damage (standards 37: notes-only values are gated under Forever).
const lightningBoltNotesShareOfEra = 0.78

func (shaman *Shaman) newLightningBoltSpellConfig(rank int) core.SpellConfig {
	spellId := LightningBoltSpellId[rank]
	baseDamageLow, baseDamageHigh := LightningBoltBaseDamageAt(rank, shaman.Level)
	if !shaman.Env.IsForever() && rank >= 4 {
		baseDamageLow /= lightningBoltNotesShareOfEra
		baseDamageHigh /= lightningBoltNotesShareOfEra
	}
	level := LightningBoltLevel[rank]
	spellCoeff := shaman.downrankCoefficient(LightningBoltSpellCoef[rank], level)
	castTime := LightningBoltCastTime[rank]
	manaCost := LightningBoltManaCost[rank]

	spell := shaman.newElectricSpellConfig(
		core.ActionID{SpellID: spellId},
		manaCost,
		time.Millisecond*time.Duration(castTime),
	)
	spell.SpellCode = SpellCode_ShamanLightningBolt
	spell.MissileSpeed = 20
	spell.RequiredLevel = level
	spell.Rank = rank
	spell.BonusCoefficient = spellCoeff

	spell.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
		stormstrike := shaman.stormstrikeBonus(target)
		spell.DamageMultiplier *= stormstrike
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		spell.DamageMultiplier /= stormstrike
		shaman.spendStormstrike(sim, target, result)

		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)
		})

		shaman.tryLightningOverload(sim, target, spell, shaman.LightningBoltOverload[rank])
	}

	return spell
}
