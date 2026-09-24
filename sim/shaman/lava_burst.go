package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

const LavaBurstFlameShockBonus = .2

// Lava Burst ranks from the Forever client data that https://foreverchanges.pro reads
// (downrank calculator): damage at each rank's cap, its growth per level up to there, and
// mana. The talent teaches rank 1 and the trainer the other two at 50 and 60.
var LavaBurstBaseDamage = [][]float64{{0}, {104.98, 135.42}, {164.02, 211.58}, {192.14, 247.86}}
var LavaBurstScaling = []core.RankScaling{{}, {48, 0.9}, {58, 1.1}, {60, 1.3}}
var LavaBurstManaCost = []float64{0, 165, 230, 265}
var LavaBurstLevel = []int32{0, 40, 50, 60}

func (shaman *Shaman) registerLavaBurstSpell() {
	if !shaman.Talents.LavaBurst {
		return
	}

	rank := max(core.HighestRankAt(shaman.Level, LavaBurstLevel), 1)
	baseDamageLow := LavaBurstScaling[rank].At(LavaBurstBaseDamage[rank][0], shaman.Level)
	baseDamageHigh := LavaBurstScaling[rank].At(LavaBurstBaseDamage[rank][1], shaman.Level)
	spellCoeff := 0.714
	castTime := time.Millisecond * 2500

	shaman.LavaBurst = shaman.RegisterSpell(core.SpellConfig{
		SpellCode:   SpellCode_ShamanLavaBurst,
		ActionID:    core.ActionID{SpellID: 51505},
		SpellSchool: core.SpellSchoolFire,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagShaman | core.SpellFlagAPL,

		MissileSpeed: 20,

		ManaCost: core.ManaCostOptions{
			FlatCost:   LavaBurstManaCost[rank],
			Multiplier: 100 - 2*shaman.Talents.Convection,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				CastTime: castTime - shaman.elementalAlacrityReduction(),
				GCD:      core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    shaman.NewTimer(),
				Duration: time.Second * 10,
			},
			ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
				castTime := shaman.ApplyCastSpeedForSpell(cast.CastTime, spell)
				shaman.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime+castTime, false)
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: spellCoeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			if shaman.hasActiveFlameShock(target) {
				spell.DamageMultiplier *= 1 + LavaBurstFlameShockBonus
				defer func() { spell.DamageMultiplier /= 1 + LavaBurstFlameShockBonus }()
			}

			baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	})
}

func (shaman *Shaman) hasActiveFlameShock(target *core.Unit) bool {
	for _, spell := range shaman.FlameShock {
		if spell != nil && spell.Dot(target).IsActive() {
			return true
		}
	}
	return false
}
