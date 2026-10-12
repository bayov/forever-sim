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
		{Name: "strikerMana", Default: 0, Min: 0, Max: 40, Step: 10},
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
		cast(hunterMongooseBite, nil, "Mongoose Bite whenever Expose Prey or a dodge opens it."),
		cast(hunterStriderKick, floor("strikerMana"), "Strider Kick on cooldown, with the talent."),
		cast(hunterRaptorStrikeQ, nil, "Queue Raptor Strike on the next swing when it is ready."),
	}
	if k["lacerate"] == 1 {
		items = append(items, cast(hunterLacerate, and(not(dotIsActive(hunterLacerate)), floor("lacerateMana")), "Lacerate when it is down."))
	}
	if k["trap"] == 1 {
		items = append(items, cast(hunterImmolationTrap, floor("trapMana"), "Immolation Trap at the hunter's feet on cooldown."))
	}
	if k["wingClip"] == 1 {
		items = append(items, cast(hunterWingClip, floor("wingClipMana"), "Wing Clip as a filler with the mana to spare."))
	}

	rot := rotation(items...)
	rot.PrepullActions = append(rot.PrepullActions, prepull(hunterAspectOfTheBeast, 3))
	return rot
}
