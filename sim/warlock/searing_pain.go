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
		ThreatMultiplier: 2,
		BonusCoefficient: data.Coef,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealDamage(sim, target, sim.Roll(damageLow, damageHigh), spell.OutcomeMagicHitAndCrit)
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
