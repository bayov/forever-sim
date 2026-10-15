package core

import (
	"time"
)

// Out of combat between pulls in a leveling session (leveling.go): resting and bandaging.
//
// When the last mob dies we're ready: the rotation can sit down to eat, drink or just rest
// (Rest (Leveling)), bandage (Bandage (Leveling)), heal or buff. Once it isn't casting and the
// global cooldown is up, we go on to the next pull. It can also rest while we search, or with
// a mob found and not pulled yet. A rotation without the rest actions gets the fixed policy
// instead (eat_below, drink_below, no_bandages).
//
// Sitting adds a third to Spirit's health regen (sittingHealthRegenMultiplier). Food and water
// come on top of Spirit's regen, health and mana alike. While we sit or bandage the rotation
// does nothing: anything else would stand us up.

// levelingRest is what we sit down for: eating, drinking, or just sitting, until health and
// mana are at these shares of max.
type levelingRest struct {
	eat, drink             bool
	healthUntil, manaUntil float64
	// Another food or drink when one runs out short of the share (the rotation's rests). The
	// fixed policy stands up once both have run out.
	refill bool
}

// ready: out of combat, deciding. We search once the rotation has looked with its global
// cooldown up and found nothing to do (onRotationIdle), or after levelingReadyAtMost.
func (s *levelingSession) ready(sim *Simulation) {
	s.setState(sim, levelingReady)
	s.move(sim, goalNone, Vec2{}, 0, nil)
	s.after(sim, &s.next, levelingReadyAtMost, s.search)
}

// A rotation that keeps acting out of combat still searches after this.
const levelingReadyAtMost = 10 * time.Second

// onRotationIdle: the rotation looked and did nothing. Ready, we go on to the next pull.
func (s *levelingSession) onRotationIdle(sim *Simulation, unit *Unit) {
	if unit != s.player || s.state != levelingReady || s.player.IsCasting(sim) {
		return
	}
	s.after(sim, &s.next, 0, s.search)
}

// outOfCombatState is whether we're out of combat and free to rest or bandage.
func (s *levelingSession) outOfCombatState() bool {
	return s.state == levelingReady || s.state == levelingSearching || s.state == levelingPulling
}

// needsRest is whether a rest would do anything: health or mana short of its share.
func (s *levelingSession) needsRest(rest levelingRest) bool {
	healthShort := s.player.CurrentHealthPercent() < rest.healthUntil
	manaShort := s.player.HasManaBar() && s.manaPercent() < rest.manaUntil
	switch {
	case rest.eat && rest.drink:
		return healthShort || manaShort
	case rest.eat:
		return healthShort
	case rest.drink:
		return manaShort
	}
	return healthShort || manaShort
}

// startRest sits us down.
func (s *levelingSession) startRest(sim *Simulation, rest levelingRest) {
	s.setState(sim, levelingResting)
	s.move(sim, goalNone, Vec2{}, 0, nil)
	s.rest = rest
	s.foodLeft, s.drinkLeft = 0, 0
	s.refill(sim)
	if sim.Log != nil {
		s.player.Log(sim, "Leveling: sitting down (eat: %v, drink: %v) at %.0f%% health and %.0f%% mana", rest.eat, rest.drink,
			100*s.player.CurrentHealthPercent(), 100*s.manaPercent())
	}
	s.busy(sim, time.Second)
}

// refill starts the next food or drink when the last one has run out and we still want it.
func (s *levelingSession) refill(sim *Simulation) {
	r := &s.rest
	if r.eat && s.foodLeft <= 0 && s.player.CurrentHealthPercent() < r.healthUntil {
		s.foodLeft = s.food.duration
		s.foods++
	}
	if r.drink && s.drinkLeft <= 0 && s.player.HasManaBar() && s.manaPercent() < r.manaUntil {
		s.drinkLeft = s.drink.duration
		s.drinks++
	}
}

// restTick: a second of food and water, and standing up once we're done.
func (s *levelingSession) restTick(sim *Simulation, dt time.Duration) {
	p := s.player
	seconds := dt.Seconds()
	if s.foodLeft > 0 {
		s.heal(sim, s.food.amount*seconds/s.food.duration.Seconds(), s.foodMetrics)
		s.foodLeft -= dt
	}
	if s.drinkLeft > 0 && p.HasManaBar() {
		p.AddMana(sim, min(s.drink.amount*seconds/s.drink.duration.Seconds(), p.MaxMana()-p.CurrentMana()), s.drinkMetrics)
		s.drinkLeft -= dt
	}
	r := &s.rest
	if r.refill {
		s.refill(sim)
	}
	healthDone := p.CurrentHealthPercent() >= r.healthUntil
	manaDone := !p.HasManaBar() || s.manaPercent() >= r.manaUntil
	if !r.refill {
		// The fixed policy: until full or until what we ate and drank has run out.
		healthDone = healthDone || s.foodLeft <= 0
		manaDone = manaDone || s.drinkLeft <= 0
	} else if r.eat && !r.drink {
		manaDone = true
	} else if r.drink && !r.eat {
		healthDone = true
	}
	if healthDone && manaDone {
		if sim.Log != nil {
			p.Log(sim, "Leveling: standing up at %.0f%% health and %.0f%% mana", 100*p.CurrentHealthPercent(), 100*s.manaPercent())
		}
		s.foodLeft, s.drinkLeft = 0, 0
		s.outOfCombatDone(sim)
	} else {
		s.busy(sim, dt)
	}
}

// startBandage: a bandage heals its amount over its time, and we're Recently Bandaged for a
// minute.
func (s *levelingSession) startBandage(sim *Simulation) {
	s.setState(sim, levelingBandaging)
	s.move(sim, goalNone, Vec2{}, 0, nil)
	s.bandageLeft = s.bandage.duration
	s.bandageReadyAt = sim.CurrentTime + recentlyBandagedDuration
	s.bandages++
	s.busy(sim, s.bandage.duration)
	if sim.Log != nil {
		s.player.Log(sim, "Leveling: bandaging")
	}
	s.after(sim, &s.next, s.bandage.duration, s.outOfCombatDone)
}

// outOfCombatDone: a rest or bandage is over, and the rotation or the fixed policy decides
// what's next.
func (s *levelingSession) outOfCombatDone(sim *Simulation) {
	if s.rotationRests {
		s.ready(sim)
	} else {
		s.outOfCombat(sim)
	}
}

// The rotation's side (apl_values_leveling.go).

// LevelingRotationRests has the session leave resting and bandaging to the rotation.
func (encounter *Encounter) LevelingRotationRests() {
	if encounter.leveling != nil {
		encounter.leveling.rotationRests = true
	}
}

// LevelingCanRest is whether we can sit down for this rest now: out of combat, not casting,
// and short of health or mana for it.
func (encounter *Encounter) LevelingCanRest(sim *Simulation, eat, drink bool, healthUntil, manaUntil float64) bool {
	s := encounter.leveling
	return s != nil && s.outOfCombatState() && !s.player.IsCasting(sim) &&
		s.needsRest(levelingRest{eat: eat, drink: drink, healthUntil: healthUntil, manaUntil: manaUntil})
}

// LevelingRest sits us down to eat, drink or just rest until health and mana are at these
// shares of max, with another food or drink whenever one runs out short of that.
func (encounter *Encounter) LevelingRest(sim *Simulation, eat, drink bool, healthUntil, manaUntil float64) {
	if encounter.LevelingCanRest(sim, eat, drink, healthUntil, manaUntil) {
		encounter.leveling.startRest(sim, levelingRest{eat: eat, drink: drink, healthUntil: healthUntil, manaUntil: manaUntil, refill: true})
	}
}

// LevelingBandageReady is whether we're out of combat and not Recently Bandaged.
func (encounter *Encounter) LevelingBandageReady(sim *Simulation) bool {
	s := encounter.leveling
	return s != nil && s.opts.bandages && s.outOfCombatState() && sim.CurrentTime >= s.bandageReadyAt
}

// LevelingCanBandage is whether a bandage would do anything now.
func (encounter *Encounter) LevelingCanBandage(sim *Simulation) bool {
	s := encounter.leveling
	return encounter.LevelingBandageReady(sim) && !s.player.IsCasting(sim) && s.player.CurrentHealth() < s.player.MaxHealth()
}

// LevelingBandage puts on a bandage.
func (encounter *Encounter) LevelingBandage(sim *Simulation) {
	if encounter.LevelingCanBandage(sim) {
		encounter.leveling.startBandage(sim)
	}
}
