package rogue

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

func (rogue *Rogue) RegisterEvasionSpell() {
	rogue.EvasionAura = rogue.RegisterAura(core.Aura{
		Label:    "Evasion",
		ActionID: core.ActionID{SpellID: 5277},
		Duration: time.Second * 15,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			rogue.AddStatDynamic(sim, stats.Dodge, 50*core.DodgeRatingPerDodgeChance)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			rogue.AddStatDynamic(sim, stats.Dodge, -50*core.DodgeRatingPerDodgeChance)
		},
	})

	rogue.Evasion = rogue.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 5277},
		SpellSchool: core.SpellSchoolPhysical,
		Flags:       core.SpellFlagAPL,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{},
			CD: core.Cooldown{
				Timer: rogue.NewTimer(),
				// Endurance shortens Sprint and Evasion, Elusiveness moved to Vanish and Blind.
				Duration: time.Duration(float64(time.Minute*5) * (1 - 0.3*float64(rogue.Talents.Endurance))),
			},
			IgnoreHaste: true,
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// Activate aura
			rogue.EvasionAura.Activate(sim)
		},
	})

	rogue.AddMajorCooldown(core.MajorCooldown{
		Spell: rogue.Evasion,
		Type:  core.CooldownTypeSurvival,
	})
}
