package rogue

import (
	"github.com/wowsims/classic/sim/core"
)

// Poised (client 1.60.1.70753, passive 1328825 at level 18): "While at maximum Energy, and out of
// combat, reduce the Energy cost of your next ability by 10." The buff 1328826 is a flat cost
// modifier (aura 107 misc 14, bp -10) with one charge that lasts until used (DurationIndex 21).
// Its class mask (255727518 / 2099456 / 67108864) covers every Energy ability the sim has except
// Death Mark (mask_2 4); Shank isn't in the client yet and is taken to be covered.
//
// The sim's fight is the combat: the rogue starts the prepull at full Energy with Poised, and
// gets it back at the pull if a prepull ability spent it and the Energy has filled up again
// (finding client_update-20261103-1, written the same way for the druid).
func (rogue *Rogue) registerPoised() {
	if !rogue.Env.IsForever() || rogue.Level < 18 {
		return
	}

	var covered []*core.Spell
	reduction := rogue.poisedReduction()

	rogue.PoisedAura = rogue.RegisterAura(core.Aura{
		Label:    "Poised",
		ActionID: core.ActionID{SpellID: 1328826},
		Duration: core.NeverExpires,
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range []*core.Spell{
				rogue.SinisterStrike, rogue.Backstab, rogue.Garrote, rogue.Ambush, rogue.Eviscerate,
				rogue.SliceAndDice, rogue.ExposeArmor, rogue.Rupture, rogue.Hemorrhage, rogue.GhostlyStrike,
				rogue.Feint, rogue.Riposte, rogue.BladeFlurry, rogue.Mutilate, rogue.Shank,
			} {
				if spell != nil && spell.Cost != nil {
					covered = append(covered, spell)
				}
			}
		},
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
			core.StartDelayedAction(sim, core.DelayedActionOptions{
				DoAt:     0,
				Priority: core.ActionPriorityPrePull + 1,
				OnAction: func(sim *core.Simulation) {
					if !aura.IsActive() && rogue.CurrentEnergy() >= rogue.MaxEnergy() {
						aura.Activate(sim)
					}
				},
			})
		},
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range covered {
				spell.Cost.FlatModifier -= reduction
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range covered {
				spell.Cost.FlatModifier += reduction
			}
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			for _, s := range covered {
				if s == spell {
					aura.Deactivate(sim)
					return
				}
			}
		},
	})
}

// Poised's 10, plus Composure's 5 a rank once the tree has it.
func (rogue *Rogue) poisedReduction() int32 {
	return 10
}
