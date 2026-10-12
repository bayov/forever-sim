package paladin

import (
	"github.com/wowsims/classic/sim/core"
)

// Twist of Light (31 point Retribution talent): replacing Seal of Command or Seal of
// Righteousness with a different seal leaves an Echo of the old seal, and the next melee
// attack applies the old seal's effect on top of the new one's, which uses the Echo up.
// The talent's 20% off the seals is in sealCostMultiplier.
//
// The client (1.60.1.70338) has Echo of Command (1311703: "a chance to activate Seal of
// Command") and Echo of Righteousness (1311704), each with one charge, no duration and
// proc type 4 (the paladin's auto attacks). So an Echo of Command rolls the seal's usual
// chance and cooldown, and an Echo of Righteousness always fires. We assume a swing that
// misses, or is dodged or parried, keeps the Echo, as a seal proc needs a hit. Seal of
// Fury and Seal of Justice echoes are not modelled, the sim has neither seal.
func (paladin *Paladin) registerTwistOfLight() {
	if !paladin.Talents.TwistOfLight {
		return
	}
	paladin.echoOfCommand = paladin.registerEchoAura("Echo of Command", 1311703)
	paladin.echoOfRighteousness = paladin.registerEchoAura("Echo of Righteousness", 1311704)
	paladin.sealEchoes = map[*core.Aura]sealEcho{}
	paladin.echoApply = map[*core.Aura]func(sim *core.Simulation, target *core.Unit){}
}

// sealEcho is the Echo a seal leaves when it is replaced, and the seal's effect on a hit.
type sealEcho struct {
	aura  *core.Aura
	apply func(sim *core.Simulation, target *core.Unit)
}

func (paladin *Paladin) registerEchoAura(label string, spellID int32) *core.Aura {
	return paladin.RegisterAura(core.Aura{
		Label:    label,
		ActionID: core.ActionID{SpellID: spellID},
		Duration: core.NeverExpires,
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || !spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) {
				return
			}
			aura.Deactivate(sim)
			paladin.echoApply[aura](sim, result.Target)
		},
	})
}

// registerSealEcho ties a seal's aura to its Echo. All ranks of a seal share the Echo
// aura, and the Echo applies the rank that was replaced.
func (paladin *Paladin) registerSealEcho(sealAura *core.Aura, echoAura *core.Aura, apply func(sim *core.Simulation, target *core.Unit)) {
	if echoAura == nil {
		return
	}
	paladin.sealEchoes[sealAura] = sealEcho{aura: echoAura, apply: apply}
}

// leaveEcho is called when newSeal replaces the current seal.
func (paladin *Paladin) leaveEcho(sim *core.Simulation, newSeal *core.Aura) {
	if paladin.sealEchoes == nil || paladin.currentSeal == nil || !paladin.currentSeal.IsActive() {
		return
	}
	old, ok := paladin.sealEchoes[paladin.currentSeal]
	if !ok {
		return
	}
	if next, ok := paladin.sealEchoes[newSeal]; ok && next.aura == old.aura {
		// The same seal again, another rank or a reseal, leaves no Echo.
		return
	}
	paladin.echoApply[old.aura] = old.apply
	old.aura.Activate(sim)
}
