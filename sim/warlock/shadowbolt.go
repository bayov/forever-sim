package warlock

import (
	"github.com/wowsims/classic/sim/core"
)

func (warlock *Warlock) getShadowBoltBaseConfig(rank int) core.SpellConfig {
	data := shadowBoltRanks[rank]
	damageLow, damageHigh := data.rangeAt(warlock.Level)

	return core.SpellConfig{
		SpellCode:     SpellCode_WarlockShadowBolt,
		ActionID:      core.ActionID{SpellID: data.SpellID},
		SpellSchool:   core.SpellSchoolShadow,
		DefenseType:   core.DefenseTypeMagic,
		ProcMask:      core.ProcMaskSpellDamage,
		Flags:         core.SpellFlagAPL | core.SpellFlagResetAttackSwing | WarlockFlagDestruction,
		RequiredLevel: data.Level,
		Rank:          rank,
		MissileSpeed:  20,

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
		ThreatMultiplier: 1,
		BonusCoefficient: data.Coef,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcDamage(sim, target, sim.Roll(damageLow, damageHigh), spell.OutcomeMagicHitAndCrit)
			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	}
}

func (warlock *Warlock) registerShadowBoltSpell() {
	warlock.ShadowBolt = make([]*core.Spell, 0)

	maxRank := core.TernaryInt(core.IncludeAQ, len(shadowBoltRanks)-1, len(shadowBoltRanks)-2)
	for rank := 1; rank <= maxRank; rank++ {
		config := warlock.getShadowBoltBaseConfig(rank)

		if config.RequiredLevel <= int(warlock.Level) {
			warlock.ShadowBolt = append(warlock.ShadowBolt, warlock.GetOrRegisterSpell(config))
		}
	}
}
