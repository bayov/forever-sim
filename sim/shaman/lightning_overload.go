package shaman

import (
	"github.com/wowsims/classic/sim/core"
)

const CastTagLightningOverload = 1

// Overloads are free instant copies of the spell that hit for half damage and generate no threat.
func (shaman *Shaman) registerOverloadSpell(config core.SpellConfig) *core.Spell {
	if shaman.Talents.LightningOverload == 0 {
		return nil
	}

	config.ActionID.Tag = CastTagLightningOverload
	config.Flags = config.Flags&^core.SpellFlagAPL | core.SpellFlagPassiveSpell | core.SpellFlagNoOnCastComplete
	config.Cast = core.CastConfig{}
	config.ManaCost = core.ManaCostOptions{}
	config.DamageMultiplier *= .5
	config.ThreatMultiplier = 0

	return shaman.RegisterSpell(config)
}

// TODO: Only rank 1 was seen, beta will confirm that the ranks go up in steps of 3%.
func (shaman *Shaman) lightningOverloadChance() float64 {
	return .03 * float64(shaman.Talents.LightningOverload)
}

func (shaman *Shaman) tryLightningOverload(sim *core.Simulation, target *core.Unit, spell *core.Spell, overload *core.Spell) {
	if overload == nil || spell == overload {
		return
	}

	if sim.Proc(shaman.lightningOverloadChance(), "Lightning Overload") {
		overload.Cast(sim, target)
	}
}
