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
		// Hunter's Mark duty (standards 23), as in hunter_ranged.
		{Name: "mark", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "markPrepull", Default: 0, Min: 0, Max: 3, Step: 0.5},
	}
}

func (hunterSV) Build(k Knobs) *proto.APLRotation {
	floor := func(name string) value {
		if k[name] <= 0 {
			return nil
		}
		return ge(currentManaPercent(), num(k[name]/100))
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
	if k["mark"] == 1 {
		items = append(items, hunterMarkDuty())
	}
	trap := cast(hunterImmolationTrap, floor("trapMana"), "Immolation Trap at the hunter's feet on cooldown.")
	items = append(items,
		cast(hunterMongooseBite, nil, "Mongoose Bite whenever Expose Prey or a dodge opens it."),
		cast(hunterStriderKick, floor("strikerMana"), "Strider Kick on cooldown, with the talent."),
	)
	if k["trap"] == 1 && k["trapFirst"] == 1 {
		items = append(items, trap)
	}
	items = append(items, cast(hunterRaptorStrikeQ, nil, "Queue Raptor Strike on the next swing when it is ready."))
	if k["lacerate"] == 1 {
		items = append(items, cast(hunterLacerate, and(not(dotIsActive(hunterLacerate)), floor("lacerateMana")), "Lacerate when it is down."))
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
		items = append(items, cast(hunterWingClip, and(floor("wingClipMana"), hold), note))
	}

	rot := rotation(items...)
	// Walk in before the pull: the UI starts every hunter at 25 yards.
	rot.PrepullActions = append(rot.PrepullActions, &proto.APLPrepullAction{
		Action:    &proto.APLAction{Action: &proto.APLAction_Move{Move: &proto.APLActionMove{RangeFromTarget: num(0)}}},
		DoAtValue: seconds(-6),
	}, prepull(hunterAspectOfTheBeast, 3))
	if k["mark"] == 1 && k["markPrepull"] > 0 {
		// In melee by then: the walk ends before it.
		rot.PrepullActions = append(rot.PrepullActions, prepull(hunterHuntersMark, k["markPrepull"]))
	}
	return rot
}
