package main

import (
	"github.com/wowsims/classic/sim/core/proto"
)

// Combat Sinister Strike under Forever rules (Restless Blades and the new racials). The
// cooldown rules are the shared ones in rogue_common.go, this adds the builder.
type rogueSS struct{}

func (rogueSS) Knobs() []Knob {
	return append(rogueCooldownKnobs(),
		Knob{Name: "ssAfterSwing", Default: 0.5, Min: 0, Max: 1, Step: 0.25},
		Knob{Name: "ssEnergy", Default: 79, Min: 0, Max: 100, Step: 10},
	)
}

func (rogueSS) Build(k Knobs) *proto.APLRotation {
	c := buildRogueCooldowns(k)
	items := []*proto.APLListItem{c.sliceAndDice}
	items = append(items, c.cooldowns...)
	items = append(items,
		cast(eviscerate, c.evisWhen, ""),
		cast(sinisterStrike, builderWhen(k["ssAfterSwing"], k["ssEnergy"]),
			"Sinister Strike right after a main hand swing so an extra attack proc restarts a swing that just happened, or when energy is about to cap."),
	)
	rot := rotation(items...)
	rot.PrepullActions = c.prepull
	return rot
}
