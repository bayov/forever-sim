package main

import (
	"github.com/wowsims/classic/sim/core/proto"
)

func init() {
	templates["hunter_sv"] = hunterSV{}
}

// Survival melees under Forever: Mongoose Bite whenever Expose Prey (or a dodge) opens it, Strider
// Kick and Raptor Strike on cooldown, Lacerate kept up, Immolation Trap at the hunter's feet, Wing
// Clip as a filler. Aspect of the Beast for its melee attack power.
type hunterSV struct{}

func (hunterSV) Knobs() []Knob {
	return []Knob{
		{Name: "lacerate", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "trap", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "wingClip", Default: 0, Min: 0, Max: 1, Step: 1},
		// Mana floors, in percent.
		{Name: "lacerateMana", Default: 20, Min: 0, Max: 60, Step: 10},
		{Name: "trapMana", Default: 30, Min: 0, Max: 60, Step: 10},
		{Name: "wingClipMana", Default: 50, Min: 0, Max: 80, Step: 10},
		// Hold Wing Clip when Mongoose Bite (opened) or Strider Kick is ready within 1 sec (1), or
		// for Mongoose Bite only (2).
		{Name: "wingClipHold", Default: 0, Min: 0, Max: 2, Step: 1},
		// Immolation Trap right after Strider Kick, ahead of Raptor Strike and Lacerate (1).
		{Name: "trapFirst", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "strikerMana", Default: 0, Min: 0, Max: 40, Step: 10},
		// The hunter's own Hunter's Mark before the pull (standards 23), as in hunter_ranged.
		{Name: "markPrepull", Default: 1.5, Min: 0, Max: 3, Step: 0.5},
		// The High Order's Read Leyline below this mana percent (0 never), as in hunter_ranged.
		{Name: "leylineMana", Default: 0, Min: 0, Max: 100, Step: 10},
		// Windfury Totem timing (standards 17: every landed main hand hit rolls it, with a 100 ms
		// cooldown after a proc, and a proc from a special attack leaves a second charge of its attack
		// power for a white swing within 1 sec). swingGate picks what the gates hold: 0 nothing, 1 Wing
		// Clip, 2 every instant main hand attack (Mongoose Bite, Strider Kick, Lacerate, Wing Clip;
		// Raptor Strike is the swing itself). swingGap holds them until this long after a main hand
		// swing, preSwing while the next swing is this close, and swingAlign until the next swing is
		// this close (0 off). The rotation waits out the hold when nothing else goes.
		{Name: "swingGate", Default: 0, Min: 0, Max: 2, Step: 1},
		{Name: "swingGap", Default: 0, Min: 0, Max: 0.3, Step: 0.1},
		{Name: "preSwing", Default: 0, Min: 0, Max: 0.3, Step: 0.1},
		{Name: "swingAlign", Default: 0, Min: 0, Max: 1.5, Step: 0.5},
	}
}

func (hunterSV) Build(k Knobs) *proto.APLRotation {
	floor := func(name string) value {
		if k[name] <= 0 {
			return nil
		}
		return ge(currentManaPercent(), num(k[name]/100))
	}

	sinceSwing := sub(mainHandSwingTime(), mainHandTimeToNext())
	var open value
	var waits []*proto.APLListItem
	if k["swingGate"] > 0 {
		gcdReady := le(hunterGCDTimeToReady(), seconds(0))
		var opens []value
		if gap := k["swingGap"]; gap > 0 {
			opens = append(opens, ge(sinceSwing, seconds(gap)))
			waits = append(waits, hunterWait(add(sub(seconds(gap), sinceSwing), seconds(0.001)), and(gcdReady, lt(sinceSwing, seconds(gap))),
				"Wait out the hold after the main hand swing."))
		}
		if pre := k["preSwing"]; pre > 0 {
			opens = append(opens, ge(mainHandTimeToNext(), seconds(pre)))
			waits = append(waits, hunterWait(add(mainHandTimeToNext(), seconds(k["swingGap"]+0.001)), and(gcdReady, lt(mainHandTimeToNext(), seconds(pre))),
				"Wait out the hold around the coming main hand swing."))
		}
		if align := k["swingAlign"]; align > 0 {
			opens = append(opens, le(mainHandTimeToNext(), seconds(align)))
			waits = append(waits, hunterWait(add(sub(mainHandTimeToNext(), seconds(align)), seconds(0.001)), and(gcdReady, gt(mainHandTimeToNext(), seconds(align))),
				"Wait until the next main hand swing is close."))
		}
		if len(opens) > 0 {
			open = and(opens...)
		}
	}
	gated := func(c value, all bool) value {
		if open == nil || (!all && k["swingGate"] < 2) {
			return c
		}
		if c == nil {
			return open
		}
		return and(c, open)
	}

	items := []*proto.APLListItem{
		{
			Notes: "Walk into melee range when the fight starts at range.",
			Action: &proto.APLAction{
				Condition: not(inMeleeRange()),
				Action:    &proto.APLAction_Move{Move: &proto.APLActionMove{RangeFromTarget: num(0)}},
			},
		},
		cast(hunterAspectOfTheBeast, not(auraIsActive(hunterAspectOfTheBeast)), "Aspect of the Beast if it ever drops."),
		autocastOtherCooldowns(nil, "Rapid Fire, racials, trinkets and potions on cooldown."),
	}
	if leyline := hunterLeyline(k, nil); leyline != nil {
		items = append(items, leyline)
	}
	items = append(items, hunterMark())
	trap := cast(hunterImmolationTrap, floor("trapMana"), "Immolation Trap at the hunter's feet on cooldown.")
	items = append(items,
		cast(hunterMongooseBite, gated(nil, false), "Mongoose Bite whenever Expose Prey or a dodge opens it."),
		cast(hunterStriderKick, gated(floor("strikerMana"), false), "Strider Kick on cooldown, with the talent."),
	)
	if k["trap"] == 1 && k["trapFirst"] == 1 {
		items = append(items, trap)
	}
	items = append(items, cast(hunterRaptorStrikeQ, nil, "Queue Raptor Strike on the next swing when it is ready."))
	if k["lacerate"] == 1 {
		items = append(items, cast(hunterLacerate, gated(and(not(dotIsActive(hunterLacerate)), floor("lacerateMana")), false), "Lacerate when it is down."))
	}
	if k["trap"] == 1 && k["trapFirst"] == 0 {
		items = append(items, trap)
	}
	if k["wingClip"] == 1 {
		var hold value
		mongooseHold := not(and(auraIsActive(spellID(1310726)), lt(spellTimeToReady(hunterMongooseBite), seconds(1))))
		note := "Wing Clip as a filler with the mana to spare."
		switch k["wingClipHold"] {
		case 1:
			hold = and(mongooseHold, ge(spellTimeToReady(hunterStriderKick), seconds(1)))
			note = "Wing Clip as a filler with the mana to spare, held when Mongoose Bite or Strider Kick is about ready."
		case 2:
			hold = and(mongooseHold)
			note = "Wing Clip as a filler with the mana to spare, held when an opened Mongoose Bite is about ready."
		}
		items = append(items, cast(hunterWingClip, gated(and(floor("wingClipMana"), hold), true), note))
	}

	if open != nil {
		items = append(items, waits...)
	}
	rot := rotation(items...)
	// Walk in before the pull: the UI starts every hunter at 25 yards.
	rot.PrepullActions = append(rot.PrepullActions, &proto.APLPrepullAction{
		Action:    &proto.APLAction{Action: &proto.APLAction_Move{Move: &proto.APLActionMove{RangeFromTarget: num(0)}}},
		DoAtValue: seconds(-6),
	}, prepull(hunterAspectOfTheBeast, 3))
	if k["markPrepull"] > 0 {
		// In melee by then: the walk ends before it.
		rot.PrepullActions = append(rot.PrepullActions, prepull(hunterHuntersMark, k["markPrepull"]))
	}
	return rot
}
