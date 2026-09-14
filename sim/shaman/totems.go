package shaman

import (
	"github.com/wowsims/classic/sim/core"
)

// Enhancing Totems is gone from the Forever tree, so Strength of Earth and Grace of Air always land
// at their improved value.
// TODO: Assumed baseline rather than deleted, beta will confirm it.
const enhancingTotemsMultiplier = 1.15

func (shaman *Shaman) newTotemSpellConfig(flatCost float64, spellID int32) core.SpellConfig {
	return core.SpellConfig{
		ActionID: core.ActionID{SpellID: spellID},
		Flags:    SpellFlagShaman | SpellFlagTotem | core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			FlatCost:   flatCost,
			Multiplier: shaman.totemManaMultiplier(),
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},
	}
}
