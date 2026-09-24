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
var ChainLightningBaseDamage = [ChainLightningRanks + 1][]float64{{0}, {85.2, 96.8}, {97.05, 108.95}, {108.88, 122.12}, {119.19, 133.21}}
var ChainLightningScaling = [ChainLightningRanks + 1]core.RankScaling{{}, {37, 0.6}, {45, 0.6}, {53, 0.7}, {60, 0.8}}
var ChainLightningSpellCoef = [ChainLightningRanks + 1]float64{0, 0.571, 0.571, 0.517, 0.571}

// Forever's costs, about 20% under Classic. The cast is 2 sec, down from 2.5.
var ChainLightningManaCost = [ChainLightningRanks + 1]float64{0, 225, 305, 390, 485}
var ChainLightningLevel = [ChainLightningRanks + 1]int{0, 32, 40, 48, 56}

func (shaman *Shaman) registerChainLightningSpell() {
	shaman.ChainLightning = make([]*core.Spell, ChainLightningRanks+1)
	shaman.ChainLightningOverload = make([]*core.Spell, ChainLightningRanks+1)

	cdTimer := shaman.NewTimer()

	for rank := 1; rank <= ChainLightningRanks; rank++ {
		// Only the ranks the level has learned, so nothing below (a totem buff aura) is
		// built for a rank the shaman cannot cast.
		if ChainLightningLevel[rank] <= int(shaman.Level) {
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

	cooldown := time.Second * 6
	castTime := time.Millisecond * 2000

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
		for hitIndex := range results {
			baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
			// Each target carries its own Stormstrike mark, so a bounce can spend one too.
			stormstrike := shaman.spendStormstrike(sim, target)
			spell.DamageMultiplier *= stormstrike
			results[hitIndex] = spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			spell.DamageMultiplier /= stormstrike
			target = sim.Environment.NextTargetUnit(target)
			spell.DamageMultiplier *= shaman.ChainLightningBounceCoefficient
		}

		for _, result := range results {
			spell.DealDamage(sim, result)
		}

		spell.DamageMultiplier = origMult

		shaman.tryLightningOverload(sim, primaryTarget, spell, shaman.ChainLightningOverload[rank])
	}

	return spell
}
