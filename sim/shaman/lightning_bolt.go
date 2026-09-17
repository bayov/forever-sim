package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

const LightningBoltRanks = 10

var LightningBoltSpellId = [LightningBoltRanks + 1]int32{0, 403, 529, 548, 915, 943, 6041, 10391, 10392, 15207, 15208}

// Damage at each rank's cap, and its growth per level up to there (1.12 spell data).
var LightningBoltBaseDamage = [LightningBoltRanks + 1][]float64{{0}, {15, 17}, {28.5, 32.5}, {48.5, 56.5}, {88, 100}, {131, 149}, {179.5, 201.5}, {235.5, 263.5}, {291.5, 325.5}, {357.5, 399.5}, {428, 477}}
var LightningBoltScaling = [LightningBoltRanks + 1]core.RankScaling{{}, {6, .4}, {13, .5}, {19, .7}, {25, 1}, {31, 1.2}, {37, 1.5}, {43, 1.7}, {49, 1.9}, {55, 2.1}, {60, 2.4}}
var LightningBoltSpellCoef = [LightningBoltRanks + 1]float64{0, .1233, .314, .554, .857, .857, .857, .857, .857, .857, .857}
var LightningBoltCastTime = [LightningBoltRanks + 1]int32{0, 1500, 2000, 2500, 3000, 3000, 3000, 3000, 3000, 3000, 3000}
var LightningBoltManaCost = [LightningBoltRanks + 1]float64{0, 15, 30, 45, 75, 105, 135, 165, 195, 230, 265}
var LightningBoltLevel = [LightningBoltRanks + 1]int{0, 1, 8, 14, 20, 26, 32, 38, 44, 50, 56}

func (shaman *Shaman) registerLightningBoltSpell() {
	shaman.LightningBolt = make([]*core.Spell, LightningBoltRanks+1)
	shaman.LightningBoltOverload = make([]*core.Spell, LightningBoltRanks+1)

	for rank := 1; rank <= LightningBoltRanks; rank++ {
		// Only the ranks the level has learned, so nothing below (a totem buff aura) is
		// built for a rank the shaman cannot cast.
		if LightningBoltLevel[rank] <= int(shaman.Level) {
			config := shaman.newLightningBoltSpellConfig(rank)
			// The overload gets a config of its own so that the two casts share no state.
			shaman.LightningBoltOverload[rank] = shaman.registerOverloadSpell(shaman.newLightningBoltSpellConfig(rank))
			shaman.LightningBolt[rank] = shaman.RegisterSpell(config)
		}
	}
}

func (shaman *Shaman) newLightningBoltSpellConfig(rank int) core.SpellConfig {
	spellId := LightningBoltSpellId[rank]
	baseDamageLow := LightningBoltScaling[rank].At(LightningBoltBaseDamage[rank][0], shaman.Level)
	baseDamageHigh := LightningBoltScaling[rank].At(LightningBoltBaseDamage[rank][1], shaman.Level)
	spellCoeff := LightningBoltSpellCoef[rank]
	castTime := LightningBoltCastTime[rank]
	manaCost := LightningBoltManaCost[rank]
	level := LightningBoltLevel[rank]

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
		stormstrike := shaman.spendStormstrike(sim, target)
		spell.DamageMultiplier *= stormstrike
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		spell.DamageMultiplier /= stormstrike

		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)
		})

		shaman.tryLightningOverload(sim, target, spell, shaman.LightningBoltOverload[rank])
	}

	return spell
}
