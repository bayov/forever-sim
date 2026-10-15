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

// hunterStingRange is the sim's minimum ranged distance (core.MinRangedAttackDistance).
const hunterStingRange = 12

func (hunterSV) Knobs() []Knob {
	return []Knob{
		{Name: "lacerate", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "trap", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "wingClip", Default: 0, Min: 0, Max: 1, Step: 1},
		// Mana floors, in percent.
		{Name: "lacerateMana", Default: 20, Min: 0, Max: 60, Step: 10},
		{Name: "trapMana", Default: 30, Min: 0, Max: 60, Step: 10},
		{Name: "wingClipMana", Default: 50, Min: 0, Max: 80, Step: 10},
		// Hold Wing Clip when Mongoose Bite (opened) or Strider Kick is ready within 1 sec.
		{Name: "wingClipHold", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "strikerMana", Default: 0, Min: 0, Max: 40, Step: 10},
		// Hunter's Mark duty (standards 23), as in hunter_ranged.
		{Name: "mark", Default: 0, Min: 0, Max: 1, Step: 1},
		// Scorpid Sting duty (standards 23): Survival steps out to the sting's minimum range to
		// keep it up, with the raid's Scorpid Sting off.
		{Name: "scorpid", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "scorpidLead", Default: 2, Min: 1, Max: 6, Step: 1},
	}
}

func (hunterSV) Build(k Knobs) *proto.APLRotation {
	floor := func(name string) value {
		if k[name] <= 0 {
			return nil
		}
		return ge(currentManaPercent(), num(k[name]/100))
	}

	var items []*proto.APLListItem
	walkIn := not(inMeleeRange())
	if k["scorpid"] == 1 {
		due := and(hunterScorpidDue(k), not(spellInFlight(hunterScorpidSting)))
		// Back into melee only once the sting is on its way.
		walkIn = and(walkIn, not(due))
		items = append(items,
			hunterScorpidDuty(k),
			&proto.APLListItem{
				Notes: "Scorpid Sting duty: step back to the sting's minimum range when it is due.",
				Action: &proto.APLAction{
					Condition: due,
					Action:    &proto.APLAction_Move{Move: &proto.APLActionMove{RangeFromTarget: num(hunterStingRange)}},
				},
			},
		)
	}
	items = append(items, []*proto.APLListItem{
		{
			Notes: "Walk into melee range when the fight starts at range.",
			Action: &proto.APLAction{
				Condition: walkIn,
				Action:    &proto.APLAction_Move{Move: &proto.APLActionMove{RangeFromTarget: num(0)}},
			},
		},
		cast(hunterAspectOfTheBeast, not(auraIsActive(hunterAspectOfTheBeast)), "Aspect of the Beast if it ever drops."),
		autocastOtherCooldowns(nil, "Rapid Fire, racials, trinkets and potions on cooldown."),
	}...)
	if k["mark"] == 1 {
		items = append(items, hunterMarkDuty())
	}
	items = append(items,
		cast(hunterMongooseBite, nil, "Mongoose Bite whenever Expose Prey or a dodge opens it."),
		cast(hunterStriderKick, floor("strikerMana"), "Strider Kick on cooldown, with the talent."),
		cast(hunterRaptorStrikeQ, nil, "Queue Raptor Strike on the next swing when it is ready."),
	)
	if k["lacerate"] == 1 {
		items = append(items, cast(hunterLacerate, and(not(dotIsActive(hunterLacerate)), floor("lacerateMana")), "Lacerate when it is down."))
	}
	if k["trap"] == 1 {
		items = append(items, cast(hunterImmolationTrap, floor("trapMana"), "Immolation Trap at the hunter's feet on cooldown."))
	}
	if k["wingClip"] == 1 {
		var hold value
		if k["wingClipHold"] == 1 {
			hold = and(
				not(and(auraIsActive(spellID(1310726)), lt(spellTimeToReady(hunterMongooseBite), seconds(1)))),
				ge(spellTimeToReady(hunterStriderKick), seconds(1)),
			)
		}
		items = append(items, cast(hunterWingClip, and(floor("wingClipMana"), hold), "Wing Clip as a filler with the mana to spare, held when Mongoose Bite or Strider Kick is about ready."))
	}

	rot := rotation(items...)
	// Walk in before the pull: the UI starts every hunter at 25 yards. On Scorpid Sting duty too: the
	// sim keeps Auto Shot firing in melee when the walk ends at range before the pull.
	rot.PrepullActions = append(rot.PrepullActions, &proto.APLPrepullAction{
		Action:    &proto.APLAction{Action: &proto.APLAction_Move{Move: &proto.APLActionMove{RangeFromTarget: num(0)}}},
		DoAtValue: seconds(-6),
	}, prepull(hunterAspectOfTheBeast, 3))
	return rot
}
