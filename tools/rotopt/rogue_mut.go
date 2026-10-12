package main

import (
	"github.com/wowsims/classic/sim/core/proto"
)

// Assassination Mutilate under Forever rules. It is the Backstab template with Mutilate
// as the builder and Venom as a finisher, so the cooldown, Rupture and Eviscerate rules
// stay the same as the dagger builds'. Backstab stays in as the fallback builder for
// when the main hand is not a dagger, it never fires with one.
type rogueMut struct{}

func init() {
	templates["rogue_mut"] = rogueMut{}
}

func (rogueMut) Knobs() []Knob {
	return withDefaults(rogueBS{}.Knobs(), map[string]float64{
		"mutilate": 1,
		"ghostly":  0,
		"venomCp":  0,
		// No Adrenaline Rush or Blade Flurry in the build, so nothing waits for them. A
		// Slice and Dice rule that waits for a spell the build lacks comes out as a bare
		// "5 combo points" and takes every finisher.
		"arSnd": 0,
		"bfSnd": 0,
	})
}

func (rogueMut) Build(k Knobs) *proto.APLRotation {
	return rogueBS{}.Build(k)
}

// withDefaults returns the knobs with some defaults changed, for a template that is
// another one with different starting values.
func withDefaults(knobs []Knob, defaults map[string]float64) []Knob {
	out := make([]Knob, len(knobs))
	for i, k := range knobs {
		if v, ok := defaults[k.Name]; ok {
			k.Default = v
		}
		out[i] = k
	}
	return out
}
