package main

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/tools/rotopt/aplkit"
)

// The Enhancement leveling template is in aplkit (aplkit/shaman_enh_leveling.go), because the
// leveling plan's rotation search (tools/levelingplan) uses it too. Here we build it for the
// -level we sim.
type shamanEnhLeveling struct{}

func (shamanEnhLeveling) Knobs() []Knob { return aplkit.ShamanEnhLeveling{}.Knobs() }

func (shamanEnhLeveling) Build(k Knobs) *proto.APLRotation {
	return aplkit.ShamanEnhLeveling{Level: playerLevel}.Build(k)
}
