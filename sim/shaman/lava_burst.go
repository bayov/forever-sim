package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

const LavaBurstFlameShockBonus = .2

// TODO: Only the damage range and the Flame Shock bonus were on the tooltip. The cast time, cooldown,
// mana cost and coefficient are taken from the spell of the same name, beta will confirm them.
func (shaman *Shaman) registerLavaBurstSpell() {
	if !shaman.Talents.LavaBurst {
		return
	}

	baseDamageLow := 158.0
	baseDamageHigh := 187.0
	spellCoeff := .5714
	castTime := time.Second * 2

	shaman.LavaBurst = shaman.RegisterSpell(core.SpellConfig{
		SpellCode:   SpellCode_ShamanLavaBurst,
		ActionID:    core.ActionID{SpellID: 51505},
		SpellSchool: core.SpellSchoolFire,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagShaman | core.SpellFlagAPL,

		MissileSpeed: 20,

		ManaCost: core.ManaCostOptions{
			BaseCost:   .1,
			Multiplier: 100 - 2*shaman.Talents.Convection,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				CastTime: castTime - shaman.elementalAlacrityReduction(),
				GCD:      core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    shaman.NewTimer(),
				Duration: time.Second * 8,
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
