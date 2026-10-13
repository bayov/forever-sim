package warlock

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Conflagrate is a row 4 talent with trainer ranks behind it under Forever (client: learned at
// 25, 32, 40, 48, 54 and 60). It consumes the warlock's own Immolate, unless Shadow and Flame
// keeps it.
func (warlock *Warlock) getConflagrateConfig(rank int) core.SpellConfig {
	data := conflagrateRanks[rank]
	damageLow, damageHigh := data.rangeAt(warlock.Level)

	return core.SpellConfig{
		SpellCode:     SpellCode_WarlockConflagrate,
		ActionID:      core.ActionID{SpellID: data.SpellID},
		SpellSchool:   core.SpellSchoolFire,
		DefenseType:   core.DefenseTypeMagic,
		ProcMask:      core.ProcMaskSpellDamage,
		Flags:         core.SpellFlagAPL | WarlockFlagDestruction,
		Rank:          rank,
		RequiredLevel: data.Level,

		ManaCost: core.ManaCostOptions{
			FlatCost: data.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    warlock.GetOrInitTimer(&warlock.conflagrateCD),
				Duration: time.Second * 10,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return warlock.getActiveImmolateSpell(target) != nil
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: data.Coef,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, sim.Roll(damageLow, damageHigh), spell.OutcomeMagicHitAndCrit)

			keepChance := 0.2 * float64(warlock.Talents.ShadowAndFlame)
			if keepChance > 0 && sim.Proc(keepChance, "Shadow and Flame") {
				return
			}
			if immoSpell := warlock.getActiveImmolateSpell(target); immoSpell != nil {
				immoSpell.Dot(target).Deactivate(sim)
			}
		},
	}
}

func (warlock *Warlock) registerConflagrateSpell() {
	if !warlock.Talents.Conflagrate {
		return
	}

	warlock.Conflagrate = make([]*core.Spell, 0)
	for rank := 1; rank < len(conflagrateRanks); rank++ {
		config := warlock.getConflagrateConfig(rank)

		if config.RequiredLevel <= int(warlock.Level) {
			warlock.Conflagrate = append(warlock.Conflagrate, warlock.GetOrRegisterSpell(config))
		}
	}
}
