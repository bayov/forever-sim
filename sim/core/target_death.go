package core

// killTarget is a target dying in a health fight: its health ran out (spell_result.go). It
// stops attacking and casting, its auras and our debuffs and dots on it end, and it leaves
// ActiveTargetUnits, so it takes no more damage and no AoE picks it. A cast at it stops
// unfinished and unpaid, and whoever had it as their target moves on to the next living one.
// When the last one dies, the fight ends.
//
// Only a health fight kills targets, and only those with health (environment.go), so a timed
// fight never gets here.
func (encounter *Encounter) killTarget(sim *Simulation, target *Unit) {
	if target.dead {
		return
	}
	target.dead = true
	encounter.removeActiveTarget(target)

	if sim.Log != nil {
		target.Log(sim, "Died")
	}

	if target.gcdAction != nil {
		target.CancelGCDTimer(sim)
	}
	target.AutoAttacks.CancelAutoSwing(sim)
	target.Hardcast = Hardcast{}
	target.auraTracker.expireAll(sim)

	for _, unit := range target.Env.Raid.AllUnits {
		// A heal or buff "at" the mob (a rotation's default target) lands on us, so only a
		// harmful cast stops.
		if unit.Hardcast.Target == target && unit.IsCasting(sim) {
			if spell := unit.GetSpell(unit.Hardcast.ActionID); spell == nil || !spell.Flags.Matches(SpellFlagHelpful) {
				unit.interruptHardcast(sim)
			}
		}
		if unit.CurrentTarget == target {
			unit.CurrentTarget = target.Env.NextTargetUnit(target)
		}
	}

	if encounter.leveling != nil {
		encounter.leveling.onMobDeath(sim, target)
	} else if len(encounter.ActiveTargetUnits) == 0 && encounter.EndFightAtHealth > 0 {
		// Step ends the fight once the damage taken passes this.
		sim.endOfCombatDamage = -1
	}
}

// IsDead says whether this target died in a health fight (killTarget).
func (unit *Unit) IsDead() bool {
	return unit.dead
}

// removeActiveTarget takes a target out of ActiveTargetUnits. The list is copied, never
// changed in place, because an AoE that kills a target may still be looping over it: that
// loop keeps the old list, and the dead target in it takes no damage (spell_result.go).
func (encounter *Encounter) removeActiveTarget(target *Unit) {
	active := make([]*Unit, 0, len(encounter.ActiveTargetUnits))
	for _, u := range encounter.ActiveTargetUnits {
		if u != target {
			active = append(active, u)
		}
	}
	encounter.ActiveTargetUnits = active
}

// addActiveTarget puts a target in ActiveTargetUnits, on a copy for the same reason.
func (encounter *Encounter) addActiveTarget(target *Unit) {
	active := make([]*Unit, 0, len(encounter.ActiveTargetUnits)+1)
	encounter.ActiveTargetUnits = append(append(active, encounter.ActiveTargetUnits...), target)
}
