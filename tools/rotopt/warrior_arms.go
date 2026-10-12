package main

import "github.com/wowsims/classic/sim/core/proto"

// Arms under Forever rules with a two-hander. Mortal Strike is the main strike. In Battle Stance
// (stance 1) Overpower, Rend (for Bloodthrill) and Spearing Strike are open and Whirlwind is not, in
// Berserker Stance (stance 2) it is the other way round with 3% more crit. Slam has an 18 sec cooldown
// under Forever (15 with Improved Slam, which also stops it delaying the swing). Execute takes over in
// the last 20% of the boss's health, and Heroic Strike spends the Rage above a threshold. The burst
// cooldowns are in warrior_common.go.
type warriorArms struct{}

func (warriorArms) Knobs() []Knob {
	return append(warriorKnobs(),
		// 1 Battle Stance, 2 Berserker Stance.
		Knob{Name: "stance", Default: 1, Min: 1, Max: 2, Step: 1},
		// Slam on cooldown above this much Rage, 0 never.
		Knob{Name: "slamRage", Default: 30, Min: 0, Max: 60, Step: 15},
		// Slam only right after a white swing, so a Slam without Improved Slam loses little of
		// the swing: 0 any time, 1 when the next swing is at least this far.
		Knob{Name: "slamAfterSwing", Default: 0, Min: 0, Max: 1, Step: 1},
		// Battle Stance only: Overpower above this much Rage, 0 never.
		Knob{Name: "overpowerRage", Default: 5, Min: 0, Max: 45, Step: 10},
		// Battle Stance only: keep Rend up (it lets Bloodthrill proc Overpower).
		Knob{Name: "rend", Default: 1, Min: 0, Max: 1, Step: 1},
		// Battle Stance only: Spearing Strike on cooldown above this much Rage, 0 never.
		Knob{Name: "spearingRage", Default: 0, Min: 0, Max: 60, Step: 15},
		// Berserker Stance only: Whirlwind on cooldown above this much Rage, 0 never.
		Knob{Name: "wwRage", Default: 25, Min: 0, Max: 55, Step: 10},
		// Mortal Strike ahead of Whirlwind and Slam (1), or Slam first (0).
		Knob{Name: "msFirst", Default: 1, Min: 0, Max: 1, Step: 1},
		// Hamstring as a filler above this much Rage outside the execute phase, 0 never. Its hits
		// proc Windfury Totem, the sword's extra attacks and Deep Wounds.
		Knob{Name: "hamstringRage", Default: 0, Min: 0, Max: 60, Step: 10},
		// Battle Stance only: switch to Berserker Stance for Recklessness (it needs Berserker Stance)
		// when recklessAt seconds are left, Recklessness is ready and the Rage fits in what Tactical
		// Mastery keeps, and stay there for the rest of the fight.
		Knob{Name: "reckSwap", Default: 0, Min: 0, Max: 1, Step: 1},
	)
}

func (warriorArms) Build(k Knobs) *proto.APLRotation {
	battle := k["stance"] == 1
	reckSwap := battle && k["reckSwap"] == 1 && k["recklessness"] == 2
	var items []*proto.APLListItem
	if reckSwap {
		items = append(items, cast(berserkerStance, and(
			le(remainingTime(), seconds(k["recklessAt"])),
			not(auraIsActive(berserkerStance)),
			spellIsReady(recklessness),
			le(currentRage(), num(25)),
		), "Berserker Stance for Recklessness at the end of the fight, with no more Rage than Tactical Mastery keeps."))
	}
	items = append(items, warriorCooldowns(k)...)

	if k["executeFirst"] == 1 {
		items = append(items, cast(execute, nil, "Execute in the execute phase."))
	}
	ms := cast(mortalStrike, nil, "Mortal Strike on cooldown.")
	var slamLine *proto.APLListItem
	if k["slamRage"] > 0 {
		when := and(ge(currentRage(), num(k["slamRage"])), not(executePhase()))
		if k["slamAfterSwing"] == 1 {
			when = and(when, ge(mainHandTimeToNext(), seconds(1.5)))
		}
		slamLine = cast(slam, when, "Slam on cooldown with Rage to spare.")
	}
	if k["msFirst"] == 1 || slamLine == nil {
		items = append(items, ms)
		if slamLine != nil {
			items = append(items, slamLine)
		}
	} else {
		items = append(items, slamLine, ms)
	}
	if battle {
		if k["overpowerRage"] > 0 {
			items = append(items, cast(overpower, ge(currentRage(), num(k["overpowerRage"])), "Overpower after a dodge or a Bloodthrill proc."))
		}
		if k["rend"] == 1 {
			items = append(items, cast(rend, and(not(dotIsActive(rend)), not(executePhase())), "Rend whenever it is down, for Bloodthrill."))
		}
		if k["spearingRage"] > 0 {
			items = append(items, cast(spearingStrike, ge(currentRage(), num(k["spearingRage"])), "Spearing Strike on cooldown. Without the talent ignore the warning on this line."))
		}
	} else if k["wwRage"] > 0 {
		items = append(items, cast(whirlwind, ge(currentRage(), num(k["wwRage"])), "Whirlwind on cooldown with the Rage for it."))
	}
	if k["executeFirst"] == 0 {
		items = append(items, cast(execute, nil, "Execute in the execute phase."))
	}
	items = append(items, heroicStrikeLine(k))
	if k["hamstringRage"] > 0 {
		items = append(items, cast(hamstring, and(ge(currentRage(), num(k["hamstringRage"])), not(executePhase())), "Hamstring with Rage to spare, for the procs."))
	}
	stance := battleStance
	if !battle {
		stance = berserkerStance
	}
	var back value = not(auraIsActive(stance))
	if reckSwap {
		back = and(back, gt(remainingTime(), seconds(k["recklessAt"])))
	}
	items = append(items, cast(stance, back, "Back to the rotation's stance."))
	rot := rotation(items...)
	rot.PrepullActions = append(rot.PrepullActions, prepull(stance, 1))
	return rot
}
