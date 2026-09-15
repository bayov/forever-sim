package main

import (
	"github.com/wowsims/classic/sim/core/proto"
)

// Combat Sinister Strike under Forever rules (Restless Blades and the new racials). The
// cooldown rules are the shared ones in rogue_common.go, this adds the builder.
type rogueSS struct{}

func (rogueSS) Knobs() []Knob {
	return append(rogueCooldownKnobs(),
		// Rupture at this many combo points when it is not ticking, 0 for never. At 60 the
		// points are worth more in Eviscerate, at low level the early Rupture ranks win.
		Knob{Name: "ruptureCp", Default: 0, Min: 0, Max: 5, Step: 1},
		// Rupture only with this much Slice and Dice left, so the points are not needed there.
		Knob{Name: "ruptureSnd", Default: 6, Min: 0, Max: 12, Step: 3},
		Knob{Name: "ssAfterSwing", Default: 0.5, Min: 0, Max: 1, Step: 0.25},
		Knob{Name: "ssEnergy", Default: 79, Min: 0, Max: 100, Step: 10},
	)
}

func (rogueSS) Build(k Knobs) *proto.APLRotation {
	c := buildRogueCooldowns(k)
	items := []*proto.APLListItem{c.sliceAndDice}
	items = append(items, c.cooldowns...)
	if k["ruptureCp"] > 0 {
		items = append(items, cast(rupture, and(
			ge(comboPoints(), num(k["ruptureCp"])),
			not(dotIsActive(rupture)),
			ge(auraRemainingTime(sliceAndDice), seconds(k["ruptureSnd"])),
			gt(remainingTime(), seconds(12)),
		), "Rupture when it is not ticking and Slice and Dice does not need the points."))
	}
	items = append(items,
		cast(eviscerate, c.evisWhen, ""),
		cast(sinisterStrike, builderWhen(k["ssAfterSwing"], k["ssEnergy"]),
			"Sinister Strike right after a main hand swing so an extra attack proc restarts a swing that just happened, or when energy is about to cap."),
	)
	rot := rotation(items...)
	rot.PrepullActions = c.prepull
	return rot
}
