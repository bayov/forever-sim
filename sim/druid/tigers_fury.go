package druid

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Tiger's Fury under Forever (1.60.1.70338, wago.tools 5217) is one rank learned at 24: 15%
// more Physical damage for 6 sec on a 30 sec cooldown, free and off the global cooldown.
// 1.12's ranks cost 30 Energy for a flat 10 to 40 damage a hit and had a 1 sec cooldown.
// The 15% is a dummy aura the server applies, so the sim takes it as a multiplier on all
// Physical damage, white hits and bleeds included.
const TigersFuryMultiplier = 1.15

func (druid *Druid) registerTigersFurySpell() {
	if druid.Level < 24 {
		return
	}
	actionID := core.ActionID{SpellID: 5217}

	druid.TigersFuryAura = druid.RegisterAura(core.Aura{
		Label:    "Tiger's Fury Aura",
		ActionID: actionID,
		Duration: 6 * time.Second,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			druid.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] *= TigersFuryMultiplier
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			druid.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] /= TigersFuryMultiplier
		},
	})

	druid.TigersFury = druid.RegisterSpell(Cat, core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagAPL,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: time.Second * 30,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			druid.TigersFuryAura.Activate(sim)
		},
	})
}
