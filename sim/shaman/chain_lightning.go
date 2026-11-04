package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

const ChainLightningRanks = 4
const ChainLightningTargetCount = int32(3)

var ChainLightningSpellId = [ChainLightningRanks + 1]int32{0, 421, 930, 2860, 10605}

// Damage at each rank's cap, its growth per level up to there, and the spell power
// coefficient, from the Forever client data that https://foreverchanges.pro reads
// (downrank calculator). Forever cut most rank damage well below 1.12, and it gave the
// low ranks close to full coefficients instead of the 1.12 penalty for spells learned
// below level 20.
//
// Hotfix push 113380 (region 91, build 1.60.1.70755, 2026-11-04; https://wago.tools/db2/SpellEffect,
// hotfixes) sets every rank's base points, growth per level and coefficient for launch, after the
// November 2 development notes ("approximately 145% of comparable rank Lightning Bolt base
// damage"): rank 4 hits for 494 to 552 at 60 (base 510 at 56, 3.3 a level; was 119 to 133), and
// every rank takes 0.714 (was 0.571, and 0.517 for rank 3). Push 113341 sets the cast to 2.5 sec
// (SpellMisc cast time 19; 2 sec before).
var ChainLightningBaseDamage = [ChainLightningRanks + 1][]float64{{0}, {207.85, 236.15}, {338.24, 379.76}, {416.18, 466.82}, {494.13, 552.27}}
var ChainLightningScaling = [ChainLightningRanks + 1]core.RankScaling{{}, {37, 1.4}, {45, 2.2}, {53, 2.7}, {60, 3.3}}
var ChainLightningSpellCoef = [ChainLightningRanks + 1]float64{0, 0.714, 0.714, 0.714, 0.714}

// Client 1.60.1.70425 cut every rank by about 35% (225 / 305 / 390 / 485 before, 145 / 200 / 250 /
// 315 after). Hotfix push 113282 (region 91, build 70755; SpellPower) sets 245 / 360 / 425 / 495.
var ChainLightningManaCost = [ChainLightningRanks + 1]float64{0, 245, 360, 425, 495}
var ChainLightningLevel = [ChainLightningRanks + 1]int{0, 32, 40, 48, 56}

// Outside Forever, Classic's damage at each rank's cap and its mana. The cast is 2.5 sec in both.
var classicChainLightningBaseDamage = [ChainLightningRanks + 1][]float64{{0}, {200, 227}, {288, 323}, {383, 430}, {505, 564}}
var classicChainLightningManaCost = [ChainLightningRanks + 1]float64{0, 280, 380, 490, 605}

func (shaman *Shaman) registerChainLightningSpell() {
	shaman.ChainLightning = make([]*core.Spell, ChainLightningRanks+1)
	shaman.ChainLightningOverload = make([]*core.Spell, ChainLightningRanks+1)

	cdTimer := shaman.NewTimer()

	for rank := 1; rank <= ChainLightningRanks; rank++ {
		// Only the ranks the level has learned, so nothing below (a totem buff aura) is
		// built for a rank the shaman cannot cast.
		if shaman.KnowsRank(ChainLightningSpellId[rank], int32(ChainLightningLevel[rank])) {
			config := shaman.newChainLightningSpellConfig(rank, cdTimer)
			// The overload gets a config of its own so that the two casts share no bounce results.
			shaman.ChainLightningOverload[rank] = shaman.registerOverloadSpell(shaman.newChainLightningSpellConfig(rank, cdTimer))
			shaman.ChainLightning[rank] = shaman.RegisterSpell(config)
		}
	}
}

func (shaman *Shaman) newChainLightningSpellConfig(rank int, cdTimer *core.Timer) core.SpellConfig {
	spellId := ChainLightningSpellId[rank]
	baseDamageLow := ChainLightningScaling[rank].At(ChainLightningBaseDamage[rank][0], shaman.Level)
	baseDamageHigh := ChainLightningScaling[rank].At(ChainLightningBaseDamage[rank][1], shaman.Level)
	spellCoeff := ChainLightningSpellCoef[rank]
	manaCost := ChainLightningManaCost[rank]
	level := ChainLightningLevel[rank]
	if !shaman.Env.IsForever() {
		baseDamageLow = classicChainLightningBaseDamage[rank][0]
		baseDamageHigh = classicChainLightningBaseDamage[rank][1]
		manaCost = classicChainLightningManaCost[rank]
	}

	cooldown := time.Second * 6
	// 2.5 sec since hotfix push 113341 (2 sec before, Classic's 2.5).
	castTime := time.Millisecond * 2500

	shaman.ChainLightningBounceCoefficient = .70 // 30% reduction per bounce
	targetCount := ChainLightningTargetCount

	spell := shaman.newElectricSpellConfig(
		core.ActionID{SpellID: spellId},
		manaCost,
		castTime,
	)

	spell.SpellCode = SpellCode_ShamanChainLightning
	spell.RequiredLevel = level
	spell.Rank = rank
	spell.BonusCoefficient = spellCoeff
	spell.Cast.CD = core.Cooldown{
		Timer:    cdTimer,
		Duration: cooldown,
	}

	results := make([]*core.SpellResult, min(targetCount, shaman.Env.GetNumTargets()))

	spell.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		primaryTarget := target
		origMult := spell.DamageMultiplier
		// It jumps only to living mobs, never back to one it already hit.
		hits := results[:sim.Encounter.AoETargetCount(int32(len(results)))]
		for hitIndex := range hits {
			baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
			// Each target carries its own Stormstrike mark, so a bounce can spend one too.
			stormstrike := shaman.stormstrikeBonus(target)
			spell.DamageMultiplier *= stormstrike
			results[hitIndex] = spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			spell.DamageMultiplier /= stormstrike
			shaman.spendStormstrike(sim, target, results[hitIndex])
			target = sim.Environment.NextTargetUnit(target)
			spell.DamageMultiplier *= shaman.ChainLightningBounceCoefficient
		}

		for _, result := range hits {
			spell.DealDamage(sim, result)
		}

		spell.DamageMultiplier = origMult

		shaman.tryLightningOverload(sim, primaryTarget, spell, shaman.ChainLightningOverload[rank])
	}

	return spell
}
