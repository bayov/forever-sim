package druid

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core"
)

const MoonfireRanks = 10

var MoonfireSpellId = [MoonfireRanks + 1]int32{0, 8921, 8924, 8925, 8926, 8927, 8928, 8929, 9833, 9834, 9835}

// Damage at each rank's cap (or 60), its growth per level, and the damage per tick, from the
// Forever client (1.60.1.70338, wago.tools). Forever cut Moonfire's hit to about two thirds
// of 1.12 and its periodic part more (rank 10 hits for 129 to 150 and ticks 60 four times,
// 1.12 has 195 to 228 and 384 over 12 sec). Every rank has the 15% direct and 13% a tick
// coefficients now, and the periodic part does not grow with level.
var MoonfireBaseDamage = [MoonfireRanks + 1][]float64{{0}, {9.19, 11.81}, {14.73, 19.27}, {24.11, 29.89}, {35.91, 43.09}, {49.62, 59.38}, {60.98, 73.02}, {74.41, 88.59}, {95.67, 113.33}, {111.55, 131.45}, {128.70, 150.50}}
var MoonfireScaling = [MoonfireRanks + 1]core.RankScaling{{}, {9, 0.5}, {15, 0.8}, {21, 1}, {27, 1.5}, {33, 1.5}, {39, 1.6}, {45, 1.7}, {51, 2.7}, {57, 2.1}, {60, 2.3}}
var MoonfireTickDamage = [MoonfireRanks + 1]float64{0, 4, 6, 9, 13, 19, 23, 31, 39, 49, 60}
var MoonfireDotTicks = [MoonfireRanks + 1]int32{0, 3, 4, 4, 4, 4, 4, 4, 4, 4, 4}
var MoonfireManaCost = [MoonfireRanks + 1]float64{0, 25, 50, 75, 105, 150, 190, 235, 280, 325, 375}
var MoonfireLevel = [MoonfireRanks + 1]int{0, 4, 10, 16, 22, 28, 34, 40, 46, 52, 58}

func (druid *Druid) registerMoonfireSpell() {
	druid.Moonfire = make([]*DruidSpell, 0)

	for rank := 1; rank <= MoonfireRanks; rank++ {
		config := druid.getMoonfireBaseConfig(rank)

		if config.RequiredLevel <= int(druid.Level) {
			druid.Moonfire = append(druid.Moonfire, druid.RegisterSpell(Humanoid|Moonkin, config))
		}
	}
}

func (druid *Druid) getMoonfireBaseConfig(rank int) core.SpellConfig {
	ticks := MoonfireDotTicks[rank]
	// Nature's Splendor makes it 3 sec longer, one more tick.
	if druid.Talents.NaturesSplendor {
		ticks++
	}
	tickLength := time.Second * 3

	spellId := MoonfireSpellId[rank]
	spellCoeff := 0.15
	spellDotCoeff := 0.13
	baseDamageLow := MoonfireScaling[rank].At(MoonfireBaseDamage[rank][0], druid.Level)
	baseDamageHigh := MoonfireScaling[rank].At(MoonfireBaseDamage[rank][1], druid.Level)
	baseDotDamage := MoonfireTickDamage[rank]
	manaCost := MoonfireManaCost[rank]
	level := MoonfireLevel[rank]

	return core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellId},
		SpellCode:   SpellCode_DruidMoonfire,
		SpellSchool: core.SpellSchoolArcane,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagOmen | core.SpellFlagAPL | core.SpellFlagResetAttackSwing,

		RequiredLevel: level,
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: 0,
			},
		},
		Dot: core.DotConfig{
			Aura: core.Aura{
				Label:    fmt.Sprintf("Moonfire (Rank %d)", rank),
				ActionID: core.ActionID{SpellID: spellId},
			},
			NumberOfTicks:    ticks,
			TickLength:       tickLength,
			BonusCoefficient: spellDotCoeff,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, baseDotDamage, isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		BonusCoefficient: spellCoeff,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)

			if result.Landed() {
				dot := spell.Dot(target)
				dot.Apply(sim)
			}
		},
	}
}
