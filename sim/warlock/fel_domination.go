package warlock

import (
	"slices"
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (warlock *Warlock) registerFelDominationCD() {
	if !warlock.Talents.FelDomination {
		return
	}

	actionID := core.ActionID{SpellID: 18708}

	aura := warlock.RegisterAura(core.Aura{
		ActionID: actionID,
		Label:    "Fel Domination",
		Duration: time.Second * 15,
		// The 50% is a percent cost row, its own factor beside Master Summoner's (wowsims/forever
		// fef2bcaebd): with Master Summoner 2/2 a summon costs 0.6 x 0.5 = 30% of its mana, not 10%.
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range warlock.SummonDemonSpells {
				spell.DefaultCast.CastTime -= time.Millisecond * 5500
				spell.Cost.Factor *= 0.5
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range warlock.SummonDemonSpells {
				spell.DefaultCast.CastTime += time.Millisecond * 5500
				spell.Cost.Factor /= 0.5
			}
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if slices.Contains(warlock.SummonDemonSpells, spell) {
				aura.Deactivate(sim)
			}
		},
	})

	spell := warlock.RegisterSpell(core.SpellConfig{
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolShadow,
		ProcMask:    core.ProcMaskEmpty,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer: warlock.NewTimer(),
				// 5 min in the Forever client (18708), 15 in 1.12.
				Duration: time.Minute * 5,
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			aura.Activate(sim)
		},
	})

	warlock.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		Type:  core.CooldownTypeUnknown,
	})
}
