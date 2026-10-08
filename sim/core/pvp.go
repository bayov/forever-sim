package core

import (
	"time"
)

// Time out of melee range (Encounter.pvp_melee_downtime), in PvP or not.
//
// An enemy player does not stand still in melee range the way a boss does. They kite,
// they run behind a pillar, and they stun or root us. A boss can also make us move, like
// when we run out of a fire or the boss knocks us back. So we take every player out of
// melee range for random stretches of 1 to 10 sec. While out of range, a player has no
// white hits and cannot use melee abilities, but spells still go out (a shock reaches 20
// yards). Fire Nova goes off around the fire totem, which the enemy has left as well, so
// it is out too.
//
// The stretches in melee range are drawn so that the share of the fight spent out of
// range comes to Encounter.pvp_melee_downtime on average. A fight starts at a random
// point of that cycle, so a short fight is out of range for the same share as a long
// one. With no downtime we never move anyone, and that's the default.
const (
	pvpMinDowntime = time.Second
	pvpMaxDowntime = time.Second * 10
	// Out of melee range, but still in range of a shock.
	pvpOutOfRangeDistance = 20
)

// Whether the unit can reach its target in melee. Only the melee downtime moves a player
// out of range during the fight.
func (unit *Unit) IsInMeleeRange() bool {
	return unit.DistanceFromTarget <= MaxMeleeAttackDistance
}

// Whether the melee downtime keeps this spell from being cast right now: a melee ability,
// or a spell that goes off from a totem, while the player is out of melee range.
func (spell *Spell) outOfPvPRange() bool {
	if spell.Unit.Env.Encounter.PvPMeleeDowntime <= 0 || spell.Unit.IsInMeleeRange() {
		return false
	}
	return spell.ProcMask.Matches(ProcMaskMelee) || spell.Flags.Matches(SpellFlagCastFromTotem)
}

func (env *Environment) startPvPDowntime(sim *Simulation) {
	downtime := env.Encounter.PvPMeleeDowntime
	if downtime <= 0 {
		return
	}
	for _, unit := range env.Raid.AllPlayerUnits {
		startPvPDowntimeFor(sim, unit, downtime)
	}
}

func startPvPDowntimeFor(sim *Simulation, unit *Unit, downtime float64) {
	inRange := unit.StartDistanceFromTarget
	outOfRange := max(inRange, pvpOutOfRangeDistance)

	// The average stretch out of range is 5.5 sec, so the average stretch in range is
	// whatever makes the out of range share come to downtime.
	meanDown := (pvpMinDowntime + pvpMaxDowntime).Seconds() / 2
	meanUp := meanDown * (1 - downtime) / max(downtime, 0.01)

	setDistance := func(sim *Simulation, distance float64) {
		unit.AutoAttacks.CancelAutoSwing(sim)
		unit.DistanceFromTarget = distance
		unit.AutoAttacks.EnableAutoSwing(sim)
	}

	downStretch := func(sim *Simulation) time.Duration {
		return pvpMinDowntime + time.Duration(sim.RandomFloat("PvP Downtime")*float64(pvpMaxDowntime-pvpMinDowntime))
	}
	upStretch := func(sim *Simulation) time.Duration {
		return DurationFromSeconds(meanUp * (0.5 + sim.RandomFloat("PvP Uptime")))
	}

	var goOutOfRange, comeBack func(sim *Simulation)
	goOutOfRange = func(sim *Simulation) {
		setDistance(sim, outOfRange)
		StartDelayedAction(sim, DelayedActionOptions{DoAt: sim.CurrentTime + downStretch(sim), OnAction: comeBack})
	}
	comeBack = func(sim *Simulation) {
		setDistance(sim, inRange)
		StartDelayedAction(sim, DelayedActionOptions{DoAt: sim.CurrentTime + upStretch(sim), OnAction: goOutOfRange})
	}

	if downtime >= 1 {
		unit.DistanceFromTarget = outOfRange
		return
	}
	// The fight starts somewhere inside a stretch, out of range with the downtime's
	// chance.
	if sim.RandomFloat("PvP Start") < downtime {
		unit.DistanceFromTarget = outOfRange
		left := time.Duration(sim.RandomFloat("PvP Start") * float64(downStretch(sim)))
		StartDelayedAction(sim, DelayedActionOptions{DoAt: left, OnAction: comeBack})
	} else {
		left := time.Duration(sim.RandomFloat("PvP Start") * float64(upStretch(sim)))
		StartDelayedAction(sim, DelayedActionOptions{DoAt: left, OnAction: goOutOfRange})
	}
}
