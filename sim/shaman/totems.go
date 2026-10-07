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

// ownTotemAura keeps the buff of a totem we put down ourselves off the stats panel.
//
// The panel turns on every aura with a build phase, so our own Strength of Earth and Grace
// of Air showed whenever the shaman knew the spell, even when the rotation never casts it.
// The Level 60 preset puts down Windfury Totem, yet the panel showed Grace of Air's 89
// Agility. The fight only gets the buff once we cast the totem.
//
// When another shaman's totem is picked in the raid buffs, the aura is already permanent
// by the time we get it, and we leave it on the panel.
func ownTotemAura(aura *core.Aura) *core.Aura {
	if aura.Duration != core.NeverExpires {
		aura.BuildPhase = core.CharacterBuildPhaseNone
	}
	return aura
}
