package main

import (
	"github.com/wowsims/classic/sim/core/proto"
)

// Subtlety under Forever rules. It is the Backstab template with the Subtlety lines on:
// Ghostly Strike on cooldown, Hemorrhage before a Rupture (Forever's Hemorrhage makes
// the rogue's own Rupture hit 15% harder), Premeditation out of Stealth and Rupture at 5
// points for Serrated Blades and Thousand Cuts. The talent and knob searches decide
// which of them stay.
type rogueSub struct{}

func init() {
	templates["rogue_sub"] = rogueSub{}
}

func (rogueSub) Knobs() []Knob {
	return withDefaults(rogueBS{}.Knobs(), map[string]float64{
		"ghostly":        1,
		"hemoForRupture": 1,
		"ruptureCp":      5,
		"premedCp":       3,
		// No Adrenaline Rush or Blade Flurry in the build, see rogue_mut.go.
		"arSnd": 0,
		"bfSnd": 0,
	})
}

func (rogueSub) Build(k Knobs) *proto.APLRotation {
	return rogueBS{}.Build(k)
}
