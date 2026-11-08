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
// (downrank calculator). The DoT does not grow with level. Forever's downranking penalty
// (sim/core/downrank.go) is on heals only, so every rank keeps its full coefficients.
//
// Hotfix push 113380 (region 91, build 1.60.1.70755, 2026-11-04; https://wago.tools/db2/SpellEffect,
// hotfixes) sets every rank's direct base points, growth per level and DoT for launch, close to
// Classic Era's as the November 2 development notes said. The spell power coefficients, costs and
// cooldowns stay the client's.
// Rank 6 deals 292 and 320 over 12 sec (was 166 and 176).
var FlameShockBaseDamage = [FlameShockRanks + 1]float64{0, 29, 56, 101, 171, 253.5, 292}
var FlameShockScaling = [FlameShockRanks + 1]core.RankScaling{{}, {15, 0.8}, {23, 1}, {33, 1.2}, {45, 1.4}, {57, 1.7}, {60, 3.8}}
var FlameShockBaseDotDamage = [FlameShockRanks + 1]float64{0, 28, 48, 96, 168, 256, 320}
var FlameShockBaseSpellCoef = [FlameShockRanks + 1]float64{0, 0.214, 0.214, 0.214, 0.214, 0.214, 0.214}
var FlameShockDotSpellCoef = [FlameShockRanks + 1]float64{0, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1}
var FlameShockManaCost = [FlameShockRanks + 1]float64{0, 55, 95, 160, 250, 345, 410}
var FlameShockLevel = [FlameShockRanks + 1]int{0, 10, 18, 28, 40, 52, 60}

func (shaman *Shaman) registerFlameShockSpell(shockTimer *core.Timer) {
	shaman.FlameShock = make([]*core.Spell, FlameShockRanks+1)

	for rank := 1; rank <= FlameShockRanks; rank++ {
		if shaman.KnowsRank(FlameShockSpellId[rank], int32(FlameShockLevel[rank])) {
			shaman.FlameShock[rank] = shaman.RegisterSpell(shaman.newFlameShockSpell(rank, shockTimer))
		}
	}
}

func (shaman *Shaman) newFlameShockSpell(rank int, shockTimer *core.Timer) core.SpellConfig {
	// Call of Flame lengthens the DoT by 3 sec, one tick, a point since client 1.60.1.70425.
	// The base tick keeps its damage, so the extra ticks add to the total.
	numTicks := 4 + int(shaman.Talents.CallOfFlame)
	tickDuration := time.Second * 3

	spellId := FlameShockSpellId[rank]
	baseDamage := FlameShockScaling[rank].At(FlameShockBaseDamage[rank], shaman.Level)
	baseDotDamage := FlameShockBaseDotDamage[rank] / 4
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
