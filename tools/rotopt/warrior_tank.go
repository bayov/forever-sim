package main

import "github.com/wowsims/classic/sim/core/proto"

// The Protection warrior tank under Forever rules, sword and board in Defensive Stance, scored with
// -tank (docs/tank_sim.md, docs/tank_warrior.md). It decides what a tank decides: Shield Block's
// upkeep, which of its own debuffs to keep up (Demoralizing Shout and Thunder Clap: tankscore.OwnDebuffs
// takes the raid's off, and its Sunder Armor ramp), the order of Shield Slam and Revenge, what Rage
// goes to Sunder Armor and Heroic Strike, Bloodrage and the opener. Nothing in it reads the tank's
// health (the replay that scores survival can't follow it): Shield Wall goes on a timer or not at all,
// and Last Stand stays out.
type warriorTank struct{}

var (
	defensiveStance   = spellID(71)
	shieldBlock       = spellID(2565)
	shieldSlam        = spellIDRank(23925, 4)
	revenge           = spellIDRank(11601, 5)
	thunderClap       = spellIDRank(11581, 6)
	demoralizingShout = spellIDRank(11556, 5)
	shieldWall        = spellID(871)
)

func (warriorTank) Knobs() []Knob {
	return []Knob{
		// Shield Block: 1 whenever its buff is down, 2 on cooldown (a recast refreshes both charges),
		// 0 never.
		{Name: "shieldBlock", Default: 1, Min: 0, Max: 2, Step: 1},
		// Shield Block above this much Rage only (0: whenever it's ready).
		{Name: "sbRage", Default: 0, Min: 0, Max: 40, Step: 10},
		// Demoralizing Shout kept up, refreshed under demoRefresh seconds left (0: never).
		{Name: "demo", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "demoRefresh", Default: 3, Min: 0, Max: 9, Step: 3},
		// Thunder Clap kept up, refreshed under tcRefresh seconds left (0: never). 2: on cooldown, for
		// its damage and threat too.
		{Name: "tc", Default: 1, Min: 0, Max: 2, Step: 1},
		{Name: "tcRefresh", Default: 2, Min: 0, Max: 8, Step: 2},
		// The debuffs ahead of Shield Slam and Revenge (1) or after them (0).
		{Name: "debuffsFirst", Default: 1, Min: 0, Max: 1, Step: 1},
		// Revenge ahead of Shield Slam (1) or after it (0).
		{Name: "revengeFirst", Default: 0, Min: 0, Max: 1, Step: 1},
		// Sunder Armor up to 5 stacks, then refreshed in its last 3 sec (1), or only as the filler (0).
		{Name: "sunderUp", Default: 1, Min: 0, Max: 1, Step: 1},
		// Sunder Armor as the filler above this much Rage (0: never).
		{Name: "sunderRage", Default: 40, Min: 0, Max: 60, Step: 10},
		// Heroic Strike on the next swing above this much Rage (100: never).
		{Name: "hsRage", Default: 50, Min: 20, Max: 100, Step: 10},
		// Bloodrage below this much Rage (0: never). It costs 20% of base health.
		{Name: "bloodrageRage", Default: 30, Min: 0, Max: 70, Step: 10},
		// The opener: 0 none, 1 Bloodrage before the pull, 2 Charge (from Defensive Stance with
		// Vanguard, or from Battle Stance and back: chargeFrom), 3 both.
		{Name: "opener", Default: 0, Min: 0, Max: 3, Step: 1},
		// Where the opener's Charge goes from: 0 Defensive Stance (needs Vanguard), 1 Battle Stance.
		{Name: "chargeFrom", Default: 0, Min: 0, Max: 1, Step: 1},
		// The Mighty Rage Potion: 0 never, 1 at the pull, 2 below potionRage Rage.
		{Name: "potion", Default: 0, Min: 0, Max: 2, Step: 1},
		{Name: "potionRage", Default: 20, Min: 10, Max: 40, Step: 10},
		// Shield Wall on a timer: 0 never, 1 at the pull, 2 in the last wallAt seconds.
		{Name: "shieldWall", Default: 0, Min: 0, Max: 2, Step: 1},
		{Name: "wallAt", Default: 12, Min: 12, Max: 60, Step: 12},
	}
}

func (warriorTank) Build(k Knobs) *proto.APLRotation {
	var items []*proto.APLListItem
	if k["bloodrageRage"] > 0 {
		items = append(items, cast(bloodrage, lt(currentRage(), num(k["bloodrageRage"])), "Bloodrage for Rage (20% of base health)."))
	}
	switch k["potion"] {
	case 1:
		items = append(items, cast(potion, nil, "Mighty Rage Potion at the pull."))
	case 2:
		items = append(items, cast(potion, lt(currentRage(), num(k["potionRage"])), "Mighty Rage Potion when short of Rage."))
	}
	switch k["shieldWall"] {
	case 1:
		items = append(items, cast(shieldWall, nil, "Shield Wall at the pull (on a timer: the score can't follow health)."))
	case 2:
		items = append(items, cast(shieldWall, le(remainingTime(), seconds(k["wallAt"])), "Shield Wall for the end of the fight."))
	}
	if k["shieldBlock"] > 0 {
		var when value
		if k["shieldBlock"] == 1 {
			when = not(auraIsActive(shieldBlock))
		}
		if k["sbRage"] > 0 {
			when = and(when, ge(currentRage(), num(k["sbRage"])))
		}
		items = append(items, cast(shieldBlock, when, "Shield Block: 2 blocks over 7 sec on a 5 sec cooldown."))
	}

	var debuffs []*proto.APLListItem
	if k["demo"] == 1 {
		debuffs = append(debuffs, cast(demoralizingShout, and(
			lt(targetAuraRemainingTime(demoralizingShout), seconds(k["demoRefresh"])),
			gt(remainingTime(), targetAuraRemainingTime(demoralizingShout)),
		), "Demoralizing Shout kept up: the tank applies its own."))
	}
	switch k["tc"] {
	case 1:
		debuffs = append(debuffs, cast(thunderClap, and(
			lt(targetAuraRemainingTime(thunderClap), seconds(k["tcRefresh"])),
			gt(remainingTime(), targetAuraRemainingTime(thunderClap)),
		), "Thunder Clap kept up: the tank applies its own slow."))
	case 2:
		debuffs = append(debuffs, cast(thunderClap, nil, "Thunder Clap on cooldown."))
	}
	strikes := []*proto.APLListItem{
		cast(shieldSlam, nil, "Shield Slam on cooldown."),
		cast(revenge, nil, "Revenge after a block, dodge or parry."),
	}
	if k["revengeFirst"] == 1 {
		strikes[0], strikes[1] = strikes[1], strikes[0]
	}
	if k["debuffsFirst"] == 1 {
		items = append(items, debuffs...)
		items = append(items, strikes...)
	} else {
		items = append(items, strikes...)
		items = append(items, debuffs...)
	}
	if k["sunderUp"] == 1 {
		items = append(items, cast(sunderArmor, and(
			or(lt(targetAuraNumStacks(sunderArmor), num(5)), lt(targetAuraRemainingTime(sunderArmor), seconds(3))),
			gt(remainingTime(), targetAuraRemainingTime(sunderArmor)),
		), "Sunder Armor to 5 stacks, refreshed in its last 3 sec."))
	}
	if k["hsRage"] < 100 {
		items = append(items, cast(heroicStrikeNext, ge(currentRage(), num(k["hsRage"])), "Heroic Strike with Rage to spare."))
	}
	if k["sunderRage"] > 0 {
		items = append(items, cast(sunderArmor, ge(currentRage(), num(k["sunderRage"])), "Sunder Armor as the filler, for its threat."))
	}
	items = append(items, cast(defensiveStance, not(auraIsActive(defensiveStance)), ""))

	rot := rotation(items...)
	opener := int(k["opener"])
	if opener&1 != 0 {
		rot.PrepullActions = append(rot.PrepullActions, prepull(bloodrage, 3))
	}
	switch {
	case opener&2 != 0 && k["chargeFrom"] == 0:
		// Vanguard: Charge from Defensive Stance, keeping all its Rage.
		rot.PrepullActions = append(rot.PrepullActions, prepull(defensiveStance, 1), prepull(charge, 0.5))
	case opener&2 != 0:
		// Without Vanguard: from Battle Stance and back, Tactical Mastery keeping 10 Rage (and 3 a
		// point of Improved Tactical Mastery).
		rot.PrepullActions = append(rot.PrepullActions,
			prepull(battleStance, 2.5), prepull(charge, 1.5), prepull(defensiveStance, 0.5))
	default:
		rot.PrepullActions = append(rot.PrepullActions, prepull(defensiveStance, 1))
	}
	return rot
}

func init() {
	templates["warrior_tank"] = warriorTank{}
}
