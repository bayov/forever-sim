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

// seenHealthChangesAt is when viewer next sees a heal candidate's health change while nothing
// happens in the sim: a reaction time after the first hit or heal on one that it doesn't see yet.
// NeverExpires when it sees them all.
func (viewer *Unit) seenHealthChangesAt(sim *Simulation) time.Duration {
	at := NeverExpires
	if !sim.Encounter.TrackHealth || viewer.ReactionTime <= 0 {
		return at
	}
	seen := sim.CurrentTime - viewer.ReactionTime
	for _, player := range viewer.HealCandidates() {
		if player == viewer || player.dead || !player.HasHealthBar() {
			continue
		}
		// From the newest change back to the first one viewer sees (healthAt).
		hb := &player.healthBar
		for k := 1; k <= healthHistorySize; k++ {
			c := hb.history[(hb.next-k+healthHistorySize)%healthHistorySize]
			if !c.set || c.at <= seen {
				break
			}
			at = min(at, c.at+viewer.ReactionTime)
		}
	}
	return at
}

// IncomingHealing is the healing in flight at target from everyone but except: the heals being cast
// at it, at their average size.
func (target *Unit) IncomingHealing(except *Unit) float64 {
	sim := target.Env.sim
	if sim == nil || !sim.Encounter.TrackHealth {
		return 0
	}
	var buf [16]*Unit
	return target.incomingFrom(sim.healCasters(except, buf[:0]))
}

// healCasters appends to casters the players other than except casting a heal with an average size
// now, in raid order: the heals in flight, for IncomingHealing on any target.
func (sim *Simulation) healCasters(except *Unit, casters []*Unit) []*Unit {
	for _, unit := range sim.Raid.AllPlayerUnits {
		hc := &unit.Hardcast
		if unit == except || hc.spell == nil || hc.spell.expectedHealing == nil || !unit.IsCasting(sim) {
			continue
		}
		casters = append(casters, unit)
	}
	return casters
}

// incomingFrom is the healing in flight at target from casters (healCasters): the same terms in the
// same order as a walk over the raid, so the sum is the same to the bit.
func (target *Unit) incomingFrom(casters []*Unit) float64 {
	total := 0.0
	for _, unit := range casters {
		if hc := &unit.Hardcast; hc.Target == target {
			total += hc.spell.expectedHealing(target)
		}
	}
	return total
}

// seenMissing is the health target is missing as viewer sees it: late by viewer's reaction time,
// less the heals in flight from the other healers. Never below 0.
func (viewer *Unit) seenMissing(target *Unit) float64 {
	// Within a rotation check, from the sight (heal_sight.go): the same terms.
	if s := viewer.sight(target.Env.sim); s != nil {
		for i := range s.picks {
			if p := &s.picks[i]; p.picked && p.best == target {
				return p.missing
			}
		}
		return viewer.seenMissingWith(target, s.casters)
	}
	return max(0, target.MaxHealth()-viewer.seenHealth(target)-target.IncomingHealing(viewer))
}

// seenMissingWith is seenMissing with the heals in flight from casters, healCasters(viewer), worked
// out once for a pick over many targets.
func (viewer *Unit) seenMissingWith(target *Unit, casters []*Unit) float64 {
	return max(0, target.MaxHealth()-viewer.seenHealth(target)-target.incomingFrom(casters))
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
