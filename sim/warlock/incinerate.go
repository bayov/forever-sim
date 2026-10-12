package warlock

import (
	"github.com/wowsims/classic/sim/core"
)

// Incinerate is the Destruction capstone with trainer ranks behind it (client 412758,
// 1293812, 1293813, learned at 40, 50 and 60): a 2.5 sec Fire bolt with a 71.4% coefficient
// that deals 25% more when the warlock's Immolate is on the target.
func (warlock *Warlock) getIncinerateConfig(rank int) core.SpellConfig {
	data := incinerateRanks[rank]
	damageLow, damageHigh := data.rangeAt(warlock.Level)

	return core.SpellConfig{
		SpellCode:     SpellCode_WarlockIncinerate,
		ActionID:      core.ActionID{SpellID: data.SpellID},
		SpellSchool:   core.SpellSchoolFire,
		DefenseType:   core.DefenseTypeMagic,
		ProcMask:      core.ProcMaskSpellDamage,
		Flags:         core.SpellFlagAPL | core.SpellFlagResetAttackSwing | WarlockFlagDestruction,
		Rank:          rank,
		RequiredLevel: data.Level,
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
			// The client's second effect is a script with 25 in it. We take it as 25% more of
			// the whole hit, spell power included, like the tooltip reads.
			multiplier := 1.0
			if warlock.getActiveImmolateSpell(target) != nil {
				multiplier = 1.25
			}
			spell.DamageMultiplier *= multiplier
			result := spell.CalcDamage(sim, target, sim.Roll(damageLow, damageHigh), spell.OutcomeMagicHitAndCrit)
			spell.DamageMultiplier /= multiplier
			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	}
}

func (warlock *Warlock) registerIncinerateSpell() {
	if !warlock.Talents.Incinerate {
		return
	}

	warlock.Incinerate = make([]*core.Spell, 0)
	for rank := 1; rank < len(incinerateRanks); rank++ {
		config := warlock.getIncinerateConfig(rank)

		if config.RequiredLevel <= int(warlock.Level) {
			warlock.Incinerate = append(warlock.Incinerate, warlock.GetOrRegisterSpell(config))
		}
	}
}
