package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Totems take a 1 sec global cooldown, not the 1.5 sec of other spells (wowhead's Forever
// spell pages, Searing and Flametongue Totem among them).
const totemGCD = time.Second

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
				GCD: totemGCD,
			},
		},
	}
}
