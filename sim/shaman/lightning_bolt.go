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
var LightningBoltBaseDamage = [LightningBoltRanks + 1][]float64{{0}, {14.39, 16.61}, {25.54, 29.46}, {42.70, 50.30}, {58.75, 67.25}, {78.83, 90.17}, {108.09, 121.91}, {142.23, 159.77}, {157.51, 176.49}, {172.56, 193.44}, {189.92, 211.68}}
var LightningBoltScaling = [LightningBoltRanks + 1]core.RankScaling{{}, {6, 0.1}, {13, 0.1}, {19, 0.3}, {25, 0.4}, {31, 0.5}, {37, 0.8}, {43, 0.8}, {49, 1}, {55, 1}, {60, 1.2}}
var LightningBoltSpellCoef = [LightningBoltRanks + 1]float64{0, 0.429, 0.571, 0.714, 0.714, 0.714, 0.714, 0.714, 0.714, 0.714, 0.714}

// Forever caps the cast at 2.5 sec from rank 4 up and takes about 20% off the mana
// cost of those ranks (wowhead Forever spell data).
var LightningBoltCastTime = [LightningBoltRanks + 1]int32{0, 1500, 2000, 2500, 2500, 2500, 2500, 2500, 2500, 2500, 2500}
var LightningBoltManaCost = [LightningBoltRanks + 1]float64{0, 15, 30, 45, 60, 85, 110, 135, 160, 190, 220}
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
