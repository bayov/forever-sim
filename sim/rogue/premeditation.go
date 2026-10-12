package rogue

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (rogue *Rogue) registerPremeditation() {
	if !rogue.Talents.Premeditation {
		return
	}

	// Forever's Premeditation no longer needs Stealth (the client spell 14183 drops the stealth only
	// attribute Ambush still has), and the points last 20 sec. 2 min cooldown, off the GCD.
	comboMetrics := rogue.NewComboPointMetrics(core.ActionID{SpellID: 14183})

	rogue.Premeditation = rogue.RegisterSpell(core.SpellConfig{
		ActionID: core.ActionID{SpellID: 14183},
		Flags:    core.SpellFlagAPL,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				Cost: 0,
				GCD:  0,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    rogue.NewTimer(),
				Duration: time.Minute * 2,
			},
		},
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.AddComboPoints(sim, 2, target, comboMetrics)
		},
	})

	rogue.AddMajorCooldown(core.MajorCooldown{
		Spell:    rogue.Premeditation,
		Type:     core.CooldownTypeDPS,
		Priority: core.CooldownPriorityLow,
		// The points go on top of what the rogue has, so hold it while a third and fourth would be lost.
		ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
			return rogue.ComboPoints() <= 3
		},
	})
}
