package warlock

import (
	"github.com/wowsims/classic/sim/core"
)

func (warlock *Warlock) getSearingPainBaseConfig(rank int) core.SpellConfig {
	data := searingPainRanks[rank]
	damageLow, damageHigh := data.rangeAt(warlock.Level)

	return core.SpellConfig{
		SpellCode:     SpellCode_WarlockSearingPain,
		ActionID:      core.ActionID{SpellID: data.SpellID},
		SpellSchool:   core.SpellSchoolFire,
		DefenseType:   core.DefenseTypeMagic,
		ProcMask:      core.ProcMaskSpellDamage,
		Flags:         core.SpellFlagAPL | core.SpellFlagResetAttackSwing | WarlockFlagDestruction,
		RequiredLevel: data.Level,
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: data.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: data.castTime(),
			},
		},

		DamageMultiplier: 1,
		// Searing Pain makes double threat, and Demonic Brand takes 17 / 33 / 50% off it (client
		// 1.60.1.70630, spell 1293695).
		ThreatMultiplier: 2 * []float64{1, 0.83, 0.67, 0.5}[warlock.Talents.DemonicBrand],
		BonusCoefficient: data.Coef,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealDamage(sim, target, sim.Roll(damageLow, damageHigh), spell.OutcomeMagicHitAndCrit)
			warlock.procDecimation(sim, result)
			if result.Landed() {
				warlock.applyDemonicBrandTo(sim, target)
			}
		},
	}
}

func (warlock *Warlock) registerSearingPainSpell() {
	warlock.SearingPain = make([]*core.Spell, 0)
	for rank := 1; rank < len(searingPainRanks); rank++ {
		config := warlock.getSearingPainBaseConfig(rank)

		if config.RequiredLevel <= int(warlock.Level) {
			warlock.SearingPain = append(warlock.SearingPain, warlock.GetOrRegisterSpell(config))
		}
	}
}
