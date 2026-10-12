package rogue

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (rogue *Rogue) registerStealthAura() {
	rogue.StealthAura = rogue.RegisterAura(core.Aura{
		Label:    "Stealth",
		ActionID: core.ActionID{SpellID: 1787},
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			// Stealth triggered auras
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
		},
		// Stealth breaks on damage taken (if not absorbed)
		// This may be desirable later, but not applicable currently
	})

	// Stealth itself, so a rotation can open from it with a prepull cast. It can't be cast
	// once the fight is on (the rogue is in combat), which is what Vanish is for.
	rogue.Stealth = rogue.RegisterSpell(core.SpellConfig{
		ActionID: core.ActionID{SpellID: 1787},
		Flags:    core.SpellFlagAPL,
		Cast: core.CastConfig{
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    rogue.NewTimer(),
				Duration: time.Second * 10,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return sim.CurrentTime < 0 && !rogue.StealthAura.IsActive()
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			rogue.AutoAttacks.CancelAutoSwing(sim)
			rogue.StealthAura.Activate(sim)
		},
	})
}
