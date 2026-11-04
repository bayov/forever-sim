package core

import "strconv"

// Absorbs: a shield on a unit that takes damage before its health does, as Power Word: Shield
// and Divine Aegis do (docs/healing_priest.md, The absorb pipeline).
//
// Damage an opponent deals goes through the target's absorbs, oldest first, as the last step of
// the damage calculation (ApplyPostOutcomeDamageModifiers, after armor, resists, damage taken
// modifiers, the outcome and the other absorbs like Ice Barrier). So everything after it sees the
// damage the shield let through: the result's Damage, OnSpellHitTaken and OnPeriodicDamageTaken,
// RemoveHealth and with it the tank score's damageTaken. An absorb never touches GainHealth or
// healTaken, and a hit the shields take whole takes no health.
//
// The healing it counts (the open default, docs/healing_sim.md, Open decisions): what a shield
// absorbs is its caster's healing on the target as it absorbs, all of it effective, and what it
// never absorbs (it ran out of time, was replaced, or the target died) is booked then as healing
// and overhealing. A shield still up when the fight ends books nothing for its rest. Shields aren't
// booked as Shielding, so HPS counts them once and EHPS counts what they absorbed.
type Absorb struct {
	// The spell (rank) that put the shield up, which books its healing.
	Spell *Spell
	// The shield's aura on the target.
	Aura *Aura

	remaining float64
}

type AbsorbArray []*Absorb

func (absorbs AbsorbArray) Get(target *Unit) *Absorb {
	return absorbs[target.UnitIndex]
}

// NewAbsorbArray registers a shield of the caster's on every unit that isn't its opponent, each
// with an aura made from config. Its ranks share it: a new one replaces the old.
func (caster *Unit) NewAbsorbArray(config Aura) AbsorbArray {
	absorbs := make(AbsorbArray, len(caster.Env.AllUnits))
	for _, target := range caster.Env.AllUnits {
		if caster.IsOpponent(target) {
			continue
		}
		absorb := &Absorb{}
		auraConfig := config
		auraConfig.Label = config.Label + "-" + strconv.Itoa(int(caster.UnitIndex))
		onExpire := config.OnExpire
		auraConfig.OnExpire = func(aura *Aura, sim *Simulation) {
			absorb.end(sim)
			if onExpire != nil {
				onExpire(aura, sim)
			}
		}
		absorb.Aura = target.GetOrRegisterAura(auraConfig)
		absorbs[target.UnitIndex] = absorb
	}
	return absorbs
}

// Remaining is what the shield can still absorb, 0 when it's down.
func (absorb *Absorb) Remaining() float64 {
	if !absorb.Aura.IsActive() {
		return 0
	}
	return absorb.remaining
}

// Apply puts the shield up for amount, for its full duration. An old one still up is replaced, and
// what it had left is booked as overhealing.
func (absorb *Absorb) Apply(sim *Simulation, spell *Spell, amount float64) {
	absorb.Aura.Deactivate(sim)
	absorb.Spell = spell
	absorb.remaining = amount
	absorb.Aura.Activate(sim)
	unit := absorb.Aura.Unit
	unit.absorbs = append(unit.absorbs, absorb)
	if sim.Log != nil {
		spell.Unit.Log(sim, "%s %s shields for %0.3f.", unit.LogLabel(), spell.ActionID, amount)
	}
}

// Add tops the shield up by amount, to at most limit, and refreshes it. One that's down goes up
// at amount (to the limit).
func (absorb *Absorb) Add(sim *Simulation, spell *Spell, amount, limit float64) {
	if !absorb.Aura.IsActive() {
		absorb.Apply(sim, spell, min(amount, limit))
		return
	}
	added := min(absorb.remaining+amount, limit) - absorb.remaining
	if added < 0 {
		added = 0
	}
	// The part booked so far belongs to the spell that put it there; the new part is this spell's.
	absorb.Spell = spell
	absorb.remaining += added
	absorb.Aura.Refresh(sim)
	if sim.Log != nil {
		spell.Unit.Log(sim, "%s %s adds %0.3f to its shield (%0.3f).", absorb.Aura.Unit.LogLabel(), spell.ActionID, added, absorb.remaining)
	}
}

// end takes the shield off its unit's list as its aura goes, booking what it never absorbed.
func (absorb *Absorb) end(sim *Simulation) {
	unit := absorb.Aura.Unit
	for i, a := range unit.absorbs {
		if a == absorb {
			unit.absorbs = append(unit.absorbs[:i], unit.absorbs[i+1:]...)
			break
		}
	}
	// The fight is over once Cleanup has set the time to the fight's length.
	if absorb.remaining > 0 && absorb.Spell != nil && sim.CurrentTime < sim.Duration {
		metrics := &absorb.Spell.SpellMetrics[unit.UnitIndex]
		metrics.TotalHealing += absorb.remaining
		metrics.TotalOverhealing += absorb.remaining
	}
	absorb.remaining = 0
}

// absorbDamage runs a hit an opponent dealt through the unit's shields, oldest first.
func (unit *Unit) absorbDamage(sim *Simulation, spell *Spell, result *SpellResult) {
	if len(unit.absorbs) == 0 || result.Damage <= 0 || !spell.Unit.IsOpponent(unit) {
		return
	}
	for len(unit.absorbs) > 0 && result.Damage > 0 {
		absorb := unit.absorbs[0]
		taken := min(absorb.remaining, result.Damage)
		absorb.remaining -= taken
		result.Damage -= taken
		absorb.Spell.SpellMetrics[unit.UnitIndex].TotalHealing += taken
		if sim.Log != nil {
			absorb.Spell.Unit.Log(sim, "%s %s absorbs %0.3f of %s, %0.3f left.", unit.LogLabel(), absorb.Spell.ActionID, taken, spell.ActionID, absorb.remaining)
		}
		if absorb.remaining <= 0 {
			// Its aura's expiry takes it off the list.
			absorb.Aura.Deactivate(sim)
		}
	}
}
