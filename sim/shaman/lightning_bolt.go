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
// level before, 45 and 56 now) so every rank is an upgrade over the one before. Forever's
// downranking penalty (sim/core/downrank.go) is on heals only: our live logs have rank 1 at
// levels 33 to 41 at its full coefficient.
//
// The 2026-10-08 beta build (1.60.1.70291) set ranks 1 to 5 again, so they grow more
// evenly into rank 6. Rank 5, the top rank at level 30, went from 75.5 to 84.5 on average
// at level 31, and rank 2 fell from 31.5 to 27.5.
//
// Hotfix push 113380 (region 91, build 1.60.1.70755, 2026-11-04; https://wago.tools/db2/SpellEffect,
// hotfixes) sets every rank's base points and growth per level for launch, after the November 2
// development notes ("roughly 78% of the base damage as the Classic Era version"). Rank 10 deals
// 342 to 381 at 60 (base 353 at 56, 2.1 a level), ranks 1 to 3 rise a little (16, 30 and 52 at the
// learn level), and the client's variances and coefficients stay.
var LightningBoltBaseDamage = [LightningBoltRanks + 1][]float64{{0}, {15.32, 17.68}, {28.32, 32.68}, {49.13, 57.87}, {70.41, 80.59}, {104.94, 120.06}, {143.80, 162.20}, {187.44, 210.56}, {233.43, 261.57}, {285.71, 320.29}, {341.82, 380.98}}
var LightningBoltScaling = [LightningBoltRanks + 1]core.RankScaling{{}, {6, 0.1}, {13, 0.1}, {19, 0.3}, {25, 0.5}, {31, 0.7}, {37, 1}, {43, 1}, {49, 1.5}, {55, 1.6}, {60, 2.1}}

// Outside Forever, Era's damage (1.15.9 client) for ranks 4 to 10 and the Forever client's from
// before the hotfix for ranks 1 to 3, which dealt about as much as Era's.
var classicLightningBoltBaseDamage = [LightningBoltRanks + 1][]float64{{0}, {14.39, 16.61}, {25.54, 29.46}, {42.70, 50.30}, {88, 100}, {131, 149}, {179.5, 201.5}, {235.5, 263.5}, {291.5, 325.5}, {357.5, 399.5}, {428.6, 476.6}}
var classicLightningBoltScaling = [LightningBoltRanks + 1]core.RankScaling{{}, {6, 0.1}, {13, 0.1}, {19, 0.3}, {25, 1}, {31, 1.2}, {37, 1.5}, {43, 1.7}, {49, 1.9}, {55, 2.1}, {60, 2.4}}
var LightningBoltSpellCoef = [LightningBoltRanks + 1]float64{0, 0.429, 0.571, 0.714, 0.714, 0.714, 0.714, 0.714, 0.714, 0.714, 0.714}

// Forever caps the cast at 2.5 sec from rank 4 up and takes about 20% off the mana
// cost of those ranks (wowhead Forever spell data).
var LightningBoltCastTime = [LightningBoltRanks + 1]int32{0, 1500, 2000, 2500, 2500, 2500, 2500, 2500, 2500, 2500, 2500}
var LightningBoltManaCost = [LightningBoltRanks + 1]float64{0, 15, 30, 45, 60, 85, 110, 135, 160, 190, 220}
var LightningBoltLevel = [LightningBoltRanks + 1]int{0, 1, 8, 14, 20, 26, 32, 38, 44, 50, 56}

// Report assumption sham_lb_31plus (user ruling 2026-10-15) took rank 10's base damage x1.12,
// because the ranks past level 38 fell from 0.61 of Era's damage to 0.44 as if not retuned yet.
// The November 2 notes and hotfix push 113380 retuned them (above), so the scale is 1.
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

// LightningBoltBaseDamageAt is a rank's damage range at a level under Forever, before spell power.
func LightningBoltBaseDamageAt(rank int, level int32) (float64, float64) {
	low := LightningBoltScaling[rank].At(LightningBoltBaseDamage[rank][0], level)
	high := LightningBoltScaling[rank].At(LightningBoltBaseDamage[rank][1], level)
	if rank == 10 {
		low *= LightningBoltRank10DamageScale
		high *= LightningBoltRank10DamageScale
	}
	return low, high
}

func (shaman *Shaman) newLightningBoltSpellConfig(rank int) core.SpellConfig {
	spellId := LightningBoltSpellId[rank]
	baseDamageLow, baseDamageHigh := LightningBoltBaseDamageAt(rank, shaman.Level)
	if !shaman.Env.IsForever() {
		baseDamageLow = classicLightningBoltScaling[rank].At(classicLightningBoltBaseDamage[rank][0], shaman.Level)
		baseDamageHigh = classicLightningBoltScaling[rank].At(classicLightningBoltBaseDamage[rank][1], shaman.Level)
	}
	level := LightningBoltLevel[rank]
	spellCoeff := LightningBoltSpellCoef[rank]
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
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		shaman.applyStormstrike(spell, target, result)
		shaman.spendStormstrike(sim, target, result)

		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)
		})

		shaman.tryLightningOverload(sim, target, spell, shaman.LightningBoltOverload[rank])
	}

	return spell
}
