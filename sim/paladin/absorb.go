package paladin

import (
	"github.com/wowsims/classic/sim/core"
)

// absorbShield is a damage absorb on the paladin (aura 69, all schools): while its aura is up it
// takes damage off every hit the paladin takes until its strength runs out, then the aura drops.
// Seal of Fury's shield and Templar's Bulwark use it.
type absorbShield struct {
	aura     *core.Aura
	strength float64
	// Absorbed is the damage the shield took this iteration, for the tests and the logs.
	absorbed float64
}

// registerAbsorbShield registers the shield's aura from config and the damage taken modifier that
// does the absorbing. The caller sets the strength with set before it activates the aura.
func (paladin *Paladin) registerAbsorbShield(config core.Aura) *absorbShield {
	shield := &absorbShield{}
	onExpire := config.OnExpire
	config.OnExpire = func(aura *core.Aura, sim *core.Simulation) {
		shield.strength = 0
		if onExpire != nil {
			onExpire(aura, sim)
		}
	}
	onReset := config.OnReset
	config.OnReset = func(aura *core.Aura, sim *core.Simulation) {
		shield.absorbed = 0
		if onReset != nil {
			onReset(aura, sim)
		}
	}
	shield.aura = paladin.RegisterAura(config)

	paladin.AddDynamicDamageTakenModifier(func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
		if !shield.aura.IsActive() || result.Damage <= 0 || shield.strength <= 0 {
			return
		}
		absorbed := min(shield.strength, result.Damage)
		result.Damage -= absorbed
		shield.strength -= absorbed
		shield.absorbed += absorbed
		if sim.Log != nil {
			paladin.Log(sim, "%s absorbed %.1f damage, %.1f left", shield.aura.Label, absorbed, shield.strength)
		}
		if shield.strength <= 0 {
			shield.aura.Deactivate(sim)
		}
	})
	return shield
}

// set gives the shield its strength and puts it up, a fresh duration.
func (shield *absorbShield) set(sim *core.Simulation, strength float64) {
	if strength <= 0 {
		return
	}
	shield.aura.Activate(sim)
	shield.strength = strength
}
