package druid

import (
	"github.com/wowsims/classic/sim/core"
)

// Poised, a druid passive Forever adds in client 1.60.1.70753 (1328829, level 32): "While at
// maximum Energy, and out of combat, reduce the Energy cost of your next ability by 10." The buff
// (1328827) is aura 107 misc 14, a flat power cost modifier of -10, with one proc charge and no
// end (DurationIndex 21). Its class mask (8622080 / 536870912) names Claw, Shred and Rake
// (0x1000), Rip and Ferocious Bite (0x800000), Ravage (0x10000), Pounce (0x20000) and Cower; the
// sim casts the first six. Not in Blizzard's notes; foreverchanges.pro lists it as new in 70753
// (finding client_update-20261103-1).
//
// The sim gives it at the pull, as the rogue's: one charge if the druid comes into the fight at
// full Energy, used up by the next of those abilities, whether or not it lands.
func (druid *Druid) registerPoised() {
	if !druid.Env.IsForever() || druid.Level < 32 {
		return
	}

	var affected []*core.Spell
	druid.PoisedAura = druid.RegisterAura(core.Aura{
		Label:    "Poised",
		ActionID: core.ActionID{SpellID: 1328827},
		Duration: core.NeverExpires,
		OnInit: func(_ *core.Aura, _ *core.Simulation) {
			for _, spell := range []*DruidSpell{druid.Claw, druid.Shred, druid.Rake, druid.Rip, druid.Ravage, druid.FerociousBite} {
				if spell != nil {
					affected = append(affected, spell.Spell)
				}
			}
		},
		OnGain: func(_ *core.Aura, _ *core.Simulation) {
			for _, spell := range affected {
				spell.Cost.FlatModifier -= 10
			}
		},
		OnExpire: func(_ *core.Aura, _ *core.Simulation) {
			for _, spell := range affected {
				spell.Cost.FlatModifier += 10
			}
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			for _, s := range affected {
				if s == spell {
					aura.Deactivate(sim)
					return
				}
			}
		},
	})

	// At time 0, before the pull starts the rotation.
	druid.RegisterPrepullAction(0, func(sim *core.Simulation) {
		if druid.CurrentEnergy() >= druid.MaxEnergy() {
			druid.PoisedAura.Activate(sim)
		}
	})
}
