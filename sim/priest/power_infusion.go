package priest

import (
	"github.com/wowsims/classic/sim/core"
)

// Power Infusion on the priest itself: 20% more spell damage for 15 sec, 3 min cooldown,
// 20% of base mana (client 1.60.1.70338). It needs Penance under Forever.
func (priest *Priest) registerPowerInfusionCD() {
	if !priest.Talents.PowerInfusion {
		return
	}

	actionID := core.ActionID{SpellID: 10060, Tag: priest.Index}
	piAura := core.PowerInfusionAura(&priest.Unit, actionID.Tag)

	priest.PowerInfusion = priest.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    SpellFlagPriest | core.SpellFlagHelpful | core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			BaseCost: 0.2,
		},
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    priest.NewTimer(),
				Duration: core.PowerInfusionCD,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			piAura.Activate(sim)
		},
	})

	priest.AddMajorCooldown(core.MajorCooldown{
		Spell: priest.PowerInfusion,
		Type:  core.CooldownTypeDPS,
	})
}
