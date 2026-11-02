package core

import "time"

// What a healer sees of the raid in a healing fight (docs/healing_sim.md), as a Classic healer did
// with HealComm, the addon that shows the heals other healers are casting:
//
//   - Health a reaction time late. A healer with a reaction time (Player.reaction_time_ms) reads
//     other players' health as it was that long ago, so two healers don't jump on the same hit in
//     the same instant, and nobody heals a hit the moment it lands.
//   - The heals in flight: the heals with a cast time other healers are casting at a player, at
//     their average size (SpellConfig.ExpectedHealing). Most Hurt, Lowest Health and Health Missing
//     take them off what's missing, so a second healer doesn't start the same heal on the same tank.
//   - Cancelling (Player.stop_cast_overheal): a reaction time before its heal lands, a healer whose
//     heal would mostly overheal stops it, unpaid, and can cast again at once.
//
// Outside a healing fight none of this applies: health reads as it is.

// seenHealth is a player's health as viewer sees it now.
func (viewer *Unit) seenHealth(target *Unit) float64 {
	sim := viewer.Env.sim
	if sim == nil || !sim.Encounter.TrackHealth || viewer == target || viewer.ReactionTime <= 0 || target.dead {
		return target.CurrentHealth()
	}
	return target.healthAt(sim.CurrentTime - viewer.ReactionTime)
}

// IncomingHealing is the healing in flight at target from everyone but except: the heals being cast
// at it, at their average size.
func (target *Unit) IncomingHealing(except *Unit) float64 {
	sim := target.Env.sim
	if sim == nil || !sim.Encounter.TrackHealth {
		return 0
	}
	total := 0.0
	for _, unit := range target.Env.Raid.AllPlayerUnits {
		hc := &unit.Hardcast
		if unit == except || hc.Target != target || hc.spell == nil || hc.spell.expectedHealing == nil || !unit.IsCasting(sim) {
			continue
		}
		total += hc.spell.expectedHealing(target)
	}
	return total
}

// seenMissing is the health target is missing as viewer sees it: late by viewer's reaction time,
// less the heals in flight from the other healers. Never below 0.
func (viewer *Unit) seenMissing(target *Unit) float64 {
	return max(0, target.MaxHealth()-viewer.seenHealth(target)-target.IncomingHealing(viewer))
}

// onHealCastStart sets up cancelling for a heal with a cast time, when the healer cancels heals.
func (unit *Unit) onHealCastStart(sim *Simulation, spell *Spell, target *Unit) {
	if unit.stopCastOverheal <= 0 || !sim.Encounter.TrackHealth || target == nil || target.Type == EnemyUnit {
		return
	}
	start := unit.Hardcast.Start
	check := &PendingAction{}
	check.OnAction = func(sim *Simulation) {
		hc := &unit.Hardcast
		if hc.Start != start || hc.spell != spell || !unit.IsCasting(sim) {
			return
		}
		// Pushback moved the cast's end: look again a reaction time before the new one.
		if at := hc.Expires - unit.ReactionTime; at > sim.CurrentTime+time.Millisecond {
			check.NextActionAt = at
			sim.AddPendingAction(check)
			return
		}
		heal := spell.expectedHealing(target)
		if heal > 0 && unit.seenMissing(target) < (1-unit.stopCastOverheal)*heal {
			if sim.Log != nil {
				unit.Log(sim, "Stops casting %s: %s is nearly full", spell.ActionID, target.Label)
			}
			unit.interruptHardcast(sim)
		}
	}
	check.NextActionAt = max(sim.CurrentTime, unit.Hardcast.Expires-unit.ReactionTime)
	sim.AddPendingAction(check)
}

// ExpectedHealing is a heal's average before crits, from its base heal's average: its healing
// power bonus, our healing multipliers and the target's healing taken modifiers, as CalcHealing
// works it out.
func (spell *Spell) ExpectedHealing(target *Unit, averageBase float64) float64 {
	heal := averageBase
	if spell.BonusCoefficient > 0 {
		heal += spell.BonusCoefficient * spell.HealingPower(target)
	}
	heal *= spell.CasterHealingMultiplier()
	return spell.applyTargetHealingModifiers(heal, spell.Unit.AttackTables[target.UnitIndex][spell.CastType])
}
