package main

import "github.com/wowsims/classic/sim/core/proto"

// Fury under Forever rules, dual wield or two-hand in Berserker Stance. Bloodthirst and Whirlwind
// are the two strikes (Whirlwind hits with both weapons now and Raging Blows takes 3 Rage off it),
// Execute takes over in the last 20% of the boss's health, and Heroic Strike spends the Rage above a
// threshold. Slam has an 18 sec cooldown under Forever and stops the swing for its cast without
// Improved Slam, so it is a knob. The burst cooldowns are in warrior_common.go.
type warriorFury struct{}

func (warriorFury) Knobs() []Knob {
	return append(warriorKnobs(),
		// Whirlwind ahead of Bloodthirst (1) or after it (0).
		Knob{Name: "wwFirst", Default: 0, Min: 0, Max: 1, Step: 1},
		// Whirlwind only above this much Rage, so Bloodthirst is never short.
		Knob{Name: "wwRage", Default: 25, Min: 25, Max: 55, Step: 10},
		// Whirlwind only when Bloodthirst is this many seconds away or more, 0 any time.
		Knob{Name: "wwBtGap", Default: 0, Min: 0, Max: 2, Step: 1},
		// Slam on cooldown above this much Rage, 0 never.
		Knob{Name: "slamRage", Default: 0, Min: 0, Max: 60, Step: 15},
		// Hamstring as a filler above this much Rage, 0 never. It procs Windfury Totem and the sword
		// extra attacks.
		Knob{Name: "hamstringRage", Default: 0, Min: 0, Max: 90, Step: 10},
	)
}

func (warriorFury) Build(k Knobs) *proto.APLRotation {
	items := warriorCooldowns(k)

	executeLine := cast(execute, nil, "Execute in the execute phase.")
	if k["executeFirst"] == 1 {
		items = append(items, executeLine)
	}
	wwWhen := ge(currentRage(), num(k["wwRage"]))
	if k["wwBtGap"] > 0 {
		wwWhen = and(wwWhen, ge(spellTimeToReady(bloodthirst), seconds(k["wwBtGap"])))
	}
	bt := cast(bloodthirst, nil, "Bloodthirst on cooldown.")
	ww := cast(whirlwind, wwWhen, "Whirlwind on cooldown with the Rage for it.")
	if k["wwFirst"] == 1 {
		items = append(items, ww, bt)
	} else {
		items = append(items, bt, ww)
	}
	if k["executeFirst"] == 0 {
		items = append(items, executeLine)
	}
	if k["slamRage"] > 0 {
		items = append(items, cast(slam, and(ge(currentRage(), num(k["slamRage"])), not(executePhase())), "Slam on cooldown with Rage to spare."))
	}
	items = append(items, heroicStrikeLine(k))
	if k["hamstringRage"] > 0 {
		items = append(items, cast(hamstring, and(ge(currentRage(), num(k["hamstringRage"])), not(executePhase())), "Hamstring with Rage to spare, for the procs."))
	}
	items = append(items, cast(berserkerStance, not(auraIsActive(berserkerStance)), "Back to Berserker Stance."))
	return rotation(items...)
}
