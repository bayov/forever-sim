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
// The November 2 development notes (post 9, for launch; not in client 1.60.1.70753) set the
// first target's damage to "approximately 145% of comparable rank Lightning Bolt base damage",
// the mana cost to about 225% of its mana and the cast to 2.5 sec. We take the comparable rank
// to be the newest Lightning Bolt at Chain Lightning's learn level (ranks 6, 7, 8 and 10 at 32,
// 40, 48 and 56), at the new Lightning Bolt damage: 1.45 times its range at that level and
// 1.45 times its growth per level. Rank 4 hits for 485 to 539 at 60 (was 119 to 133).
var ChainLightningBaseDamage = [ChainLightningRanks + 1][]float64{{0}, {203.01, 227.90}, {270.21, 301.88}, {338.27, 376.72}, {484.75, 539.04}}
var ChainLightningScaling = [ChainLightningRanks + 1]core.RankScaling{{}, {37, 1.70}, {45, 1.93}, {53, 2.15}, {60, 2.71}}
var ChainLightningSpellCoef = [ChainLightningRanks + 1]float64{0, 0.571, 0.571, 0.517, 0.571}

// Client 1.60.1.70425 cut every rank by about 35% (225 / 305 / 390 / 485 before, 145 / 200 / 250 /
// 315 after). The November 2 notes make it about 225% of the comparable Lightning Bolt's mana
// (110, 135, 160 and 220).
var ChainLightningManaCost = [ChainLightningRanks + 1]float64{0, 248, 304, 360, 495}
var ChainLightningLevel = [ChainLightningRanks + 1]int{0, 32, 40, 48, 56}

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
	spellCoeff := shaman.downrankCoefficient(ChainLightningSpellCoef[rank], ChainLightningLevel[rank])
	manaCost := ChainLightningManaCost[rank]
	level := ChainLightningLevel[rank]

	cooldown := time.Second * 6
	// 2.5 sec since the November 2 notes (2 sec before, Classic's 2.5).
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
