package shaman

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core"
)

const FlameShockRanks = 6

var FlameShockSpellId = [FlameShockRanks + 1]int32{0, 8050, 8052, 8053, 10447, 10448, 29228}

// Direct damage at each rank's cap, its growth per level up to there, and the spell power
// coefficients, from the Forever client data that https://foreverchanges.pro reads
// (downrank calculator). The DoT does not grow with level.
var FlameShockBaseDamage = [FlameShockRanks + 1]float64{0, 24, 38, 49, 89, 136.5, 166}
var FlameShockScaling = [FlameShockRanks + 1]core.RankScaling{{}, {15, 0.8}, {23, 1}, {33, 1.2}, {45, 1.4}, {57, 1.7}, {60, 2.1}}
var FlameShockBaseDotDamage = [FlameShockRanks + 1]float64{0, 28, 32, 56, 84, 136, 176}
var FlameShockBaseSpellCoef = [FlameShockRanks + 1]float64{0, 0.214, 0.214, 0.214, 0.214, 0.214, 0.214}
var FlameShockDotSpellCoef = [FlameShockRanks + 1]float64{0, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1}
var FlameShockManaCost = [FlameShockRanks + 1]float64{0, 55, 95, 160, 250, 345, 410}
var FlameShockLevel = [FlameShockRanks + 1]int{0, 10, 18, 28, 40, 52, 60}

func (shaman *Shaman) registerFlameShockSpell(shockTimer *core.Timer) {
	shaman.FlameShock = make([]*core.Spell, FlameShockRanks+1)

	for rank := 1; rank <= FlameShockRanks; rank++ {
		if FlameShockLevel[rank] <= int(shaman.Level) {
			shaman.FlameShock[rank] = shaman.RegisterSpell(shaman.newFlameShockSpell(rank, shockTimer))
		}
	}
}

func (shaman *Shaman) newFlameShockSpell(rank int, shockTimer *core.Timer) core.SpellConfig {
	numTicks := 4
	tickDuration := time.Second * 3

	spellId := FlameShockSpellId[rank]
	baseDamage := FlameShockScaling[rank].At(FlameShockBaseDamage[rank], shaman.Level)
	baseDotDamage := FlameShockBaseDotDamage[rank] / float64(numTicks)
	baseSpellCoeff := FlameShockBaseSpellCoef[rank]
	dotSpellCoeff := FlameShockDotSpellCoef[rank]
	manaCost := FlameShockManaCost[rank]
	level := FlameShockLevel[rank]

	spell := shaman.newShockSpellConfig(
		core.ActionID{SpellID: spellId},
		core.SpellSchoolFire,
		manaCost,
		shockTimer,
	)

	spell.SpellCode = SpellCode_ShamanFlameShock
	spell.RequiredLevel = level
	spell.Rank = rank

	spell.Cast.IgnoreHaste = true

	spell.BonusCoefficient = baseSpellCoeff

	spell.Dot = core.DotConfig{
		Aura: core.Aura{
			Label: fmt.Sprintf("Flame Shock (Rank %d)", rank),
		},

		NumberOfTicks:    int32(numTicks),
		TickLength:       tickDuration,
		BonusCoefficient: dotSpellCoeff,

		OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
			dot.Snapshot(target, baseDotDamage, isRollover)
		},

		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
			dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
		},
	}

	spell.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		if result.Landed() {
			spell.Dot(result.Target).Apply(sim)
		}
	}

	return spell
}
