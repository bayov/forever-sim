package core

import (
	"time"

	"github.com/wowsims/classic/sim/core/proto"
)

// Out of combat between pulls in a leveling session (leveling.go): resting and bandaging.
//
// When the last mob dies we're ready: the rotation can sit down to eat, drink or just rest
// (Rest (Leveling)), bandage (Bandage (Leveling)), heal or buff. Once it isn't casting and the
// global cooldown is up, we go on to the next pull. It can also rest while we search, or with
// a mob found and not pulled yet. Only the rotation rests: a rotation with no Rest or Bandage
// line gets default ones in front of its own (levelingDefaultRests).
//
// Sitting adds a third to Spirit's health regen (sittingHealthRegenMultiplier). Food and water
// come on top of Spirit's regen, health and mana alike. While we sit or bandage the rotation
// does nothing: anything else would stand us up.
//
// The food, water and bandage are spells of ours, like in the game's combat log. Each food or
// drink we start and each bandage is a cast, with an aura while it lasts (Food, Drink and First
// Aid, then Recently Bandaged), and the food and bandage heal in ticks. So they show in the
// log, the healing metrics and the replay's action history like any other spell.

// levelingRest is what we sit down for: eating, drinking, or just sitting, until health and
// mana are at these shares of max, with another food or drink when one runs out short.
type levelingRest struct {
	eat, drink             bool
	healthUntil, manaUntil float64
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

// sitting is whether we're sitting, bandaging or dead: nothing else goes until that's over,
// since anything would stand us up.
func (s *levelingSession) sitting() bool {
	return s.state == levelingResting || s.state == levelingBandaging || s.state == levelingDead
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
		s.foodSpell.Cast(sim, s.player)
	}
	if r.drink && s.drinkLeft <= 0 && s.player.HasManaBar() && s.manaPercent() < r.manaUntil {
		s.drinkLeft = s.drink.duration
		s.drinks++
		s.drinkSpell.Cast(sim, s.player)
	}
}

// restTick: a second of food and water, and standing up once we're done.
func (s *levelingSession) restTick(sim *Simulation, dt time.Duration) {
	p := s.player
	seconds := dt.Seconds()
	if s.foodLeft > 0 {
		s.healTick(sim, s.foodSpell, s.food.amount*seconds/s.food.duration.Seconds())
		s.foodLeft -= dt
	}
	if s.drinkLeft > 0 && p.HasManaBar() {
		p.AddMana(sim, min(s.drink.amount*seconds/s.drink.duration.Seconds(), p.MaxMana()-p.CurrentMana()), s.drinkMetrics)
		s.drinkLeft -= dt
	}
	r := &s.rest
	s.refill(sim)
	healthDone := p.CurrentHealthPercent() >= r.healthUntil
	manaDone := !p.HasManaBar() || s.manaPercent() >= r.manaUntil
	if r.eat && !r.drink {
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
	s.bandageSpell.Cast(sim, s.player)
	s.recentlyBandagedAura.Activate(sim)
	s.busy(sim, s.bandage.duration)
	if sim.Log != nil {
		s.player.Log(sim, "Leveling: bandaging")
	}
	s.after(sim, &s.next, s.bandage.duration, s.outOfCombatDone)
}

// registerRestSpells makes the food, water and bandage spells of ours, each with its aura.
// They go off the rotation's global cooldown and don't proc anything: we're sitting.
func (s *levelingSession) registerRestSpells() {
	c := s.character
	restSpell := func(r restorer, label string) (*Spell, *Aura) {
		actionID := ActionID{ItemID: r.itemID}
		aura := c.RegisterAura(Aura{Label: label, ActionID: actionID, Duration: r.duration})
		spell := c.RegisterSpell(SpellConfig{
			ActionID:    actionID,
			SpellSchool: SpellSchoolPhysical,
			ProcMask:    ProcMaskEmpty,
			// Food and bandages heal what they say, whatever our healing modifiers.
			Flags:            SpellFlagHelpful | SpellFlagIgnoreModifiers | SpellFlagNoOnCastComplete,
			DamageMultiplier: 1,
			ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
				aura.Activate(sim)
			},
		})
		return spell, aura
	}
	s.foodSpell, s.foodAura = restSpell(s.food, "Food")
	s.drinkSpell, s.drinkAura = restSpell(s.drink, "Drink")
	s.bandageSpell, s.bandageAura = restSpell(s.bandage, "First Aid")
	s.recentlyBandagedAura = c.RegisterAura(Aura{
		Label:    "Recently Bandaged",
		ActionID: ActionID{SpellID: 11196},
		Duration: recentlyBandagedDuration,
	})
}

// healTick heals us by a tick of food or a bandage, up to our max health.
func (s *levelingSession) healTick(sim *Simulation, spell *Spell, amount float64) {
	amount = min(amount, s.player.MaxHealth()-s.player.CurrentHealth())
	if amount <= 0 {
		return
	}
	spell.DealPeriodicHealing(sim, spell.CalcHealing(sim, s.player, amount, spell.OutcomeHealing))
}

// outOfCombatDone: a rest or bandage is over, and the rotation decides what's next.
func (s *levelingSession) outOfCombatDone(sim *Simulation) {
	s.ready(sim)
}

// The rotation's side (apl_values_leveling.go).

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
		encounter.leveling.startRest(sim, levelingRest{eat: eat, drink: drink, healthUntil: healthUntil, manaUntil: manaUntil})
	}
}

// LevelingBandageReady is whether we're out of combat and not Recently Bandaged.
func (encounter *Encounter) LevelingBandageReady(sim *Simulation) bool {
	s := encounter.leveling
	return s != nil && s.outOfCombatState() && sim.CurrentTime >= s.bandageReadyAt
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

// levelingDefaultRests are the rest lines a leveling rotation gets in front of its own when it
// has no Rest or Bandage line, so that one written for a single fight still recovers between
// pulls: a bandage below 50% health, then eating and drinking to full below 50% health or 40%
// mana, never while we run (searching). They are the thresholds the eat_below and drink_below
// options had. Outside a session, for a pet, or for a rotation that rests itself, none. A
// disabled Rest or Bandage line counts too, so a rotation that never rests just has one
// switched off.
func (rot *APLRotation) levelingDefaultRests(config *proto.APLRotation) []*APLAction {
	s := rot.unit.Env.Encounter.leveling
	if s == nil || rot.unit != s.player {
		return nil
	}
	for _, item := range config.PriorityList {
		if aplActionRests(item.Action) {
			return nil
		}
	}
	below := func(value *proto.APLValue, share string) *proto.APLValue {
		return &proto.APLValue{Value: &proto.APLValue_Cmp{Cmp: &proto.APLValueCompare{
			Op:  proto.APLValueCompare_OpLt,
			Lhs: value,
			Rhs: &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: share}}},
		}}}
	}
	and := func(vals ...*proto.APLValue) *proto.APLValue {
		return &proto.APLValue{Value: &proto.APLValue_And{And: &proto.APLValueAnd{Vals: vals}}}
	}
	standing := &proto.APLValue{Value: &proto.APLValue_Not{Not: &proto.APLValueNot{
		Val: &proto.APLValue{Value: &proto.APLValue_IsMoving{IsMoving: &proto.APLValueIsMoving{}}}}}}
	lowHealth := below(&proto.APLValue{Value: &proto.APLValue_CurrentHealthPercent{CurrentHealthPercent: &proto.APLValueCurrentHealthPercent{}}}, "50%")
	needRest := lowHealth
	if rot.unit.HasManaBar() {
		lowMana := below(&proto.APLValue{Value: &proto.APLValue_CurrentManaPercent{CurrentManaPercent: &proto.APLValueCurrentManaPercent{}}}, "40%")
		needRest = &proto.APLValue{Value: &proto.APLValue_Or{Or: &proto.APLValueOr{Vals: []*proto.APLValue{lowHealth, lowMana}}}}
	}
	configs := []*proto.APLAction{
		{Condition: and(standing, lowHealth), Action: &proto.APLAction_LevelingBandage{LevelingBandage: &proto.APLActionLevelingBandage{}}},
		{Condition: and(standing, needRest), Action: &proto.APLAction_LevelingRest{LevelingRest: &proto.APLActionLevelingRest{Eat: true, Drink: rot.unit.HasManaBar()}}},
	}
	var actions []*APLAction
	for _, config := range configs {
		rot.doAndRecordWarnings(nil, false, func() {
			if action := rot.newAPLAction(config); action != nil {
				action.Finalize(rot)
				actions = append(actions, action)
			}
		})
	}
	return actions
}

// aplActionRests is whether the action is a Rest or Bandage, or a sequence with one in it.
func aplActionRests(action *proto.APLAction) bool {
	switch {
	case action == nil:
		return false
	case action.GetLevelingRest() != nil, action.GetLevelingBandage() != nil:
		return true
	}
	var inner []*proto.APLAction
	if seq := action.GetSequence(); seq != nil {
		inner = seq.Actions
	} else if seq := action.GetStrictSequence(); seq != nil {
		inner = seq.Actions
	}
	for _, a := range inner {
		if aplActionRests(a) {
			return true
		}
	}
	return false
}
