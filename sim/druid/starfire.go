package druid

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

const StarfireRanks = 7

var StarfireSpellId = [StarfireRanks + 1]int32{0, 2912, 8949, 8950, 8951, 9875, 9876, 25298}

// Damage at each rank's cap (or 60) and its growth per level, from the Forever client
// (1.60.1.70338, wago.tools). Forever cut Starfire to about 70% of 1.12 (rank 7 is 350 to
// 412, 1.12 has 496 to 584), with the same costs, 3.5 sec cast and full coefficient.
var StarfireBaseDamage = [StarfireRanks + 1][]float64{{0}, {78.66, 96.34}, {107.26, 130.74}, {139.70, 167.50}, {190.44, 227.16}, {255.87, 302.53}, {313.43, 369.77}, {349.96, 412.04}}
var StarfireScaling = [StarfireRanks + 1]core.RankScaling{{}, {25, 1.3}, {32, 1.5}, {40, 1.6}, {48, 1.8}, {56, 2.2}, {60, 2.3}, {60, 2.4}}
var StarfireManaCost = [StarfireRanks + 1]float64{0, 95, 135, 180, 230, 275, 315, 340}
var StarfireLevel = [StarfireRanks + 1]int{0, 20, 26, 34, 42, 50, 58, 60}

func (druid *Druid) registerStarfireSpell() {
	druid.Starfire = make([]*DruidSpell, StarfireRanks+1)

	maxRank := core.TernaryInt(core.IncludeAQ, StarfireRanks, StarfireRanks-1)
	for rank := 1; rank <= maxRank; rank++ {
		config := druid.newStarfireSpellConfig(rank)

		if config.RequiredLevel <= int(druid.Level) {
			druid.Starfire[rank] = druid.RegisterSpell(Humanoid|Moonkin, config)
		}
	}
}

func (druid *Druid) newStarfireSpellConfig(rank int) core.SpellConfig {
	spellId := StarfireSpellId[rank]
	baseDamageLow := StarfireScaling[rank].At(StarfireBaseDamage[rank][0], druid.Level)
	baseDamageHigh := StarfireScaling[rank].At(StarfireBaseDamage[rank][1], druid.Level)
	manaCost := StarfireManaCost[rank]
	level := StarfireLevel[rank]

	castTime := 3500

	return core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellId},
		SpellCode:   SpellCode_DruidStarfire,
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
				CastTime: time.Millisecond * time.Duration(castTime),
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		},
	}
}
