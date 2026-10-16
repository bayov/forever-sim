package paladin

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// ConcentrationAuraSpellId is Concentration Aura (client 1.60.1.70425, spell 19746, level 22,
// the same as 70338 and Classic): the party in it has a 35% chance to ignore the pushback
// (aura 149, all schools) when hit while casting. As the paladin's own aura (the "Aura"
// option) it protects the paladin's own casts, Light's Vigil, Hammer of Wrath and Holy Wrath,
// from the raid hits. Voice of Truth (1310897, 100% for 6 sec on a 3 min cooldown) is not in
// the sim.
const ConcentrationAuraSpellId = 19746

func (paladin *Paladin) registerConcentrationAura() {
	if paladin.Options.Aura != proto.PaladinAura_ConcentrationAura || paladin.Level < 22 {
		return
	}

	paladin.OnSpellRegistered(func(spell *core.Spell) {
		spell.PushbackReduction = max(spell.PushbackReduction, 0.35)
	})

	core.MakePermanent(paladin.RegisterAura(core.Aura{
		Label:    "Concentration Aura",
		ActionID: core.ActionID{SpellID: ConcentrationAuraSpellId},
	}))
}
