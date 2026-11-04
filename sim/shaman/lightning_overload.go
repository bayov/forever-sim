package shaman

import (
	"github.com/wowsims/classic/sim/core"
)

const CastTagLightningOverload = 1

// Overloads are free instant copies of the spell that hit for half damage and generate no threat.
//
// A copy is a triggered spell without Not a Proc (the client's overload rows, 408477 for Lightning
// Bolt rank 10, carry no Attributes 3), so it reaches only listeners with Can Proc From Procs. Judgement
// of Wisdom (20355), Elemental Devastation and Water Shield lack it, so they don't hear the copy, as
// with Fire Nova's hits. wowsims/forever marks the copy a proc too (electric_spell.go).
func (shaman *Shaman) registerOverloadSpell(config core.SpellConfig) *core.Spell {
	if shaman.Talents.LightningOverload == 0 {
		return nil
	}

	config.ActionID.Tag = CastTagLightningOverload
	config.ProcMask = core.ProcMaskSpellDamageProc
	config.Flags = config.Flags&^core.SpellFlagAPL | core.SpellFlagPassiveSpell | core.SpellFlagNoOnCastComplete
	config.Cast = core.CastConfig{}
	config.ManaCost = core.ManaCostOptions{}
	config.DamageMultiplier *= .5
	config.ThreatMultiplier = 0

	return shaman.RegisterSpell(config)
}

// 10 / 20 / 30% since the November 2 development notes (post 9, for launch; 3 / 7 / 10% in the
// beta client, 1.60.1.70753 included). Classic has no Lightning Overload talent; outside Forever it
// keeps the client's 3 / 7 / 10% (standards 37: notes-only values are gated under Forever).
func (shaman *Shaman) lightningOverloadChance() float64 {
	if !shaman.Env.IsForever() {
		return []float64{0, .03, .07, .10}[shaman.Talents.LightningOverload]
	}
	return []float64{0, .10, .20, .30}[shaman.Talents.LightningOverload]
}

func (shaman *Shaman) tryLightningOverload(sim *core.Simulation, target *core.Unit, spell *core.Spell, overload *core.Spell) {
	if overload == nil || spell == overload {
		return
	}

	if sim.Proc(shaman.lightningOverloadChance(), "Lightning Overload") {
		overload.Cast(sim, target)
	}
}
