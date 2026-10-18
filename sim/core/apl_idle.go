package core

import (
	"cmp"
	"slices"
	"time"

	"github.com/wowsims/classic/sim/core/proto"
)

// A rotation that finds nothing to do waits 50ms and checks its whole priority list again (see
// DoNextAction). A rogue waiting for energy does that thousands of times a fight, and almost
// every check finds nothing again. This file lets those checks be skipped without changing what
// the rotation does.
//
// When a check finds nothing to do, we work out the earliest time any action in the priority
// list could be ready, as long as nothing happens in the sim before then. For a spell that's when
// its cooldown and the GCD are up and we have the energy for it (Spell.castableNotBefore). For a
// condition that is false, it's when the condition could turn true: a comparison of the time left
// in the fight against 90s turns at a time we can work out, and an aura we don't have can't come
// up without an event (aplSteadyUntil). Anything we can't tell counts as "any time now", and then
// we don't skip.
//
// The next 50ms wake-ups skip the check while the sim counted no event since (Simulation.events)
// and that earliest time hasn't come. They still wait 50ms more, the same as a check that found
// nothing, so the rotation acts at the same moments as before.

// The time we take off a time we worked out from floats (energy regen, or when a comparison
// turns), so that float rounding can't make us skip a check that would have found something.
const aplIdleMargin = time.Millisecond

// doNextActionOnWakeUp is DoNextAction for the rotation's own wake-up (the GCD action), which may
// skip the check, see canSkipCheck.
func (apl *APLRotation) doNextActionOnWakeUp(sim *Simulation) {
	apl.wakeUp = true
	apl.DoNextAction(sim)
	apl.wakeUp = false
}

// canSkipCheck is whether this wake-up would find nothing to do, like the check before it.
//
// That holds when nothing happened in the sim since that check except this wake-up, and the
// earliest time an action could be ready hasn't come yet. We only work that time out here, at
// the first wake-up after the check. Most checks that find nothing are followed by an event
// (a swing, a proc), and then we never need it.
func (apl *APLRotation) canSkipCheck(sim *Simulation) bool {
	if !apl.wakeUp || !apl.idle || sim.events != apl.idleEvents+1 {
		apl.idleHopeless = false
		return false
	}
	if !apl.idleBounded {
		// When the last time we worked it out we couldn't tell, and nothing but our wake-ups
		// happened since, we couldn't tell now either. So we don't try again until an event.
		if apl.idleHopeless {
			return false
		}
		apl.idleBounded = true
		apl.idleUntil = apl.readyNotBefore(sim)
		apl.idleHopeless = apl.idleUntil <= sim.CurrentTime
	}
	return sim.CurrentTime < apl.idleUntil
}

// noteIdle is for a check that found nothing to do, see canSkipCheck.
func (apl *APLRotation) noteIdle(sim *Simulation) {
	// A leveling session moves us and the mobs around in ways the bounds don't know about.
	if len(apl.controllingActions) != 0 || sim.Encounter.leveling != nil {
		return
	}
	apl.idle = true
	apl.idleEvents = sim.events
	apl.idleBounded = false
}

// readyNotBefore is the earliest time an action of the priority list could be ready when nothing
// happens in the sim before then, or the current time when we can't tell.
func (apl *APLRotation) readyNotBefore(sim *Simulation) time.Duration {
	// First we work out when each action itself could be ready, which is cheap. Then we look at
	// the conditions, starting with the action that could be ready soonest, and stop once no other
	// action could come before what we found. That way we look at as few conditions as we can.
	if len(apl.idleOrder) != len(apl.priorityList) {
		apl.idleOrder = make([]int, len(apl.priorityList))
		apl.idleBounds = make([]time.Duration, len(apl.priorityList))
	}
	for i, action := range apl.priorityList {
		if action.bound == nil {
			return sim.CurrentTime
		}
		t := action.bound.readyNotBefore(sim)
		if t <= sim.CurrentTime && action.condition == nil {
			return sim.CurrentTime
		}
		apl.idleBounds[i] = t
		apl.idleOrder[i] = i
	}
	slices.SortFunc(apl.idleOrder, func(a, b int) int { return cmp.Compare(apl.idleBounds[a], apl.idleBounds[b]) })

	until := NeverExpires
	for _, i := range apl.idleOrder {
		t := apl.idleBounds[i]
		if t >= until {
			break
		}
		if condition := apl.priorityList[i].condition; condition != nil && !condition.GetBool(sim) {
			t = max(t, aplSteadyUntil(sim, condition))
		}
		if until = min(until, t); until <= sim.CurrentTime {
			return sim.CurrentTime
		}
	}
	return until
}

// aplActionReadyBound is for actions that can tell, without side effects, the earliest time they
// could be ready when nothing happens in the sim before then.
type aplActionReadyBound interface {
	// readyNotBefore returns sim.CurrentTime when the action may be ready now.
	readyNotBefore(sim *Simulation) time.Duration
	// mayBeReadyNow is false when readyNotBefore would return a later time, but it's cheaper.
	mayBeReadyNow(sim *Simulation) bool
}

// readyNotBefore is the earliest time the action could be ready when nothing happens in the sim
// before then, or the current time when we can't tell.
func (action *APLAction) readyNotBefore(sim *Simulation) time.Duration {
	if action.bound == nil {
		return sim.CurrentTime
	}
	t := action.bound.readyNotBefore(sim)
	if action.condition == nil || action.condition.GetBool(sim) {
		return t
	}
	return max(t, aplSteadyUntil(sim, action.condition))
}

// aplHasSideEffects is whether evaluating the value can change the sim, so we must not skip it.
//
// Those are the values that ask whether a spell can be cast. Asking that about a spell we lack
// the mana for starts the time we count as out of mana.
func aplHasSideEffects(value APLValue) bool {
	if value == nil {
		return false
	}
	switch value.(type) {
	case *APLValueSpellCanCast, *APLValueSequenceIsReady, *APLValueSequenceTimeToReady:
		return true
	}
	for _, inner := range value.GetInnerValues() {
		if aplHasSideEffects(inner) {
			return true
		}
	}
	return false
}

func (action *APLActionCastSpell) readyNotBefore(sim *Simulation) time.Duration {
	if action.target.IsDynamic() && action.target.Get() == nil {
		return sim.CurrentTime
	}
	return action.spell.castableNotBefore(sim)
}

func (action *APLActionCastSpell) mayBeReadyNow(sim *Simulation) bool {
	return action.spell.castableNow(sim)
}

func (action *APLActionChannelSpell) readyNotBefore(sim *Simulation) time.Duration {
	return action.spell.castableNotBefore(sim)
}
func (action *APLActionChannelSpell) mayBeReadyNow(sim *Simulation) bool {
	return action.spell.castableNow(sim)
}

// The first major cooldown that is ready and should be used, once the GCD is up.
//
// A cooldown that is ready gets asked whether it can be cast (getFirstReadyMCD), which for one
// we lack the mana for starts the time we count as out of mana. So we only say it can wait for
// an event when it costs no mana.
func (action *APLActionAutocastOtherCooldowns) readyNotBefore(sim *Simulation) time.Duration {
	character := action.character
	first := NeverExpires
	for _, mcd := range character.majorCooldowns {
		t := max(mcd.ReadyAt(), mcd.Spell.castableNotBefore(sim))
		if t <= sim.CurrentTime && mcd.ShouldActivateNeedsEvent && mcd.numUsages >= len(mcd.timings) &&
			(mcd.Spell.Cost == nil || mcd.Spell.Cost.CostType() != CostTypeMana) && !mcd.ShouldActivate(sim, character) {
			t = NeverExpires
		}
		first = min(first, t)
	}
	return max(sim.CurrentTime, character.GCD.ReadyAt(), character.minReady, first)
}

// IsReady asks every ready cooldown whether it can be cast before it looks at the GCD, so we
// can only skip it when no cooldown is ready, like getFirstReadyMCD does.
func (action *APLActionAutocastOtherCooldowns) mayBeReadyNow(sim *Simulation) bool {
	return sim.CurrentTime >= action.character.minReady
}

func (action *APLActionSequence) mayBeReadyNow(sim *Simulation) bool {
	return action.curIdx < len(action.subactions)
}
func (action *APLActionSequence) readyNotBefore(sim *Simulation) time.Duration {
	if action.curIdx >= len(action.subactions) {
		// Only another action of the rotation can reset the sequence.
		return NeverExpires
	}
	return action.subactions[action.curIdx].readyNotBefore(sim)
}

// castableNow is a quick check, without side effects, of what castableNotBefore looks at: when
// it returns false, CanCast returns false too.
func (spell *Spell) castableNow(sim *Simulation) bool {
	if spell == nil {
		return true
	}
	if spell.DefaultCast.GCD > 0 && !spell.Unit.GCD.IsReady(sim) {
		return false
	}
	if !BothTimersReady(spell.CD.Timer, spell.SharedCD.Timer, sim) {
		return false
	}
	if spell.Cost == nil {
		return true
	}
	switch spell.Cost.CostType() {
	case CostTypeEnergy:
		return spell.Unit.CurrentEnergy() >= spell.Cost.GetCurrentCost()
	case CostTypeRage:
		return spell.Unit.CurrentRage() >= spell.Cost.GetCurrentCost()
	case CostTypeFocus:
		return spell.Unit.CurrentFocus() >= spell.Cost.GetCurrentCost()
	}
	return true
}

// castableNotBefore is the earliest time CanCast could return true when nothing happens in the
// sim before then. It looks at the GCD, the cooldowns, and energy, rage or focus, without side
// effects. It returns the current time when the spell may be castable now.
//
// Rage and focus only come with an event (a hit, a focus tick), so when we lack them, nothing
// changes until one. Under Forever energy comes at a steady rate, so we know when we will have
// enough. We leave mana to CanCast, because its check also keeps track of the time we spend out
// of mana.
func (spell *Spell) castableNotBefore(sim *Simulation) time.Duration {
	t := sim.CurrentTime
	if spell == nil {
		return t
	}
	if spell.DefaultCast.GCD > 0 {
		t = max(t, spell.Unit.GCD.ReadyAt())
	}
	t = max(t, BothTimersReadyAt(spell.CD.Timer, spell.SharedCD.Timer))
	if spell.Cost == nil {
		return t
	}
	switch spell.Cost.CostType() {
	case CostTypeEnergy:
		if need := spell.Cost.GetCurrentCost() - spell.Unit.CurrentEnergy(); need > 0 {
			t = max(t, spell.Unit.energyBar.gainedNotBefore(sim, need))
		}
	case CostTypeRage:
		if spell.Unit.CurrentRage() < spell.Cost.GetCurrentCost() {
			return NeverExpires
		}
	case CostTypeFocus:
		if spell.Unit.CurrentFocus() < spell.Cost.GetCurrentCost() {
			return NeverExpires
		}
	}
	return t
}

// gainedNotBefore is the earliest time we have gained the given energy when nothing happens in
// the sim before then. Under Classic energy comes in ticks, which are events.
func (eb *energyBar) gainedNotBefore(sim *Simulation, energy float64) time.Duration {
	rate := EnergyPerSecond * eb.EnergyTickMultiplier
	if !eb.continuous || rate <= 0 {
		return NeverExpires
	}
	return sim.CurrentTime + DurationFromSeconds(energy/rate) - aplIdleMargin
}

// energyPerSecond is how fast energy goes up when nothing happens in the sim, and until when.
func (eb *energyBar) energyPerSecond(sim *Simulation) (float64, time.Duration) {
	rate := EnergyPerSecond * eb.EnergyTickMultiplier
	if !eb.continuous || rate <= 0 || eb.currentEnergy >= eb.maxEnergy {
		return 0, NeverExpires
	}
	return rate, sim.CurrentTime + DurationFromSeconds((eb.maxEnergy-eb.currentEnergy)/rate) - aplIdleMargin
}

// manaPerSecond is how fast mana goes up when nothing happens in the sim, and until when. See
// manaBar.settle: under Forever mana comes at the casting rate until the 5 second rule runs out,
// and at the full rate after. Under Classic it comes in ticks, which are events.
func (mb *manaBar) manaPerSecond(sim *Simulation) (float64, time.Duration) {
	unit := mb.unit
	if !mb.continuous || !unit.IsEnabled() || unit.currentMana >= unit.MaxMana() {
		return 0, NeverExpires
	}
	rate, until := max(0, mb.regenPerSecondNotCasting), NeverExpires
	if fsr := unit.PseudoStats.FiveSecondRuleRefreshTime; sim.CurrentTime < fsr {
		rate, until = max(0, mb.regenPerSecondCasting), fsr
	}
	if rate > 0 {
		until = min(until, sim.CurrentTime+DurationFromSeconds((unit.MaxMana()-unit.currentMana)/rate)-aplIdleMargin)
	}
	return rate, until
}

// aplValueSteady is for values that can tell how long they keep their value when nothing happens
// in the sim.
type aplValueSteady interface {
	// steadyUntil returns sim.CurrentTime when the value could change any moment.
	steadyUntil(sim *Simulation) time.Duration
}

// aplValueTrend is for numbers that can tell how they change with time when nothing happens in
// the sim.
type aplValueTrend interface {
	// trend returns how much the value changes each second (a duration in seconds), and until
	// when it keeps changing that way. ok is false when we can't tell.
	trend(sim *Simulation) (perSecond float64, until time.Duration, ok bool)
}

// aplSteadyUntil is how long the value stays the same when nothing happens in the sim, or the
// current time when we can't tell.
func aplSteadyUntil(sim *Simulation, value APLValue) time.Duration {
	if steady, ok := value.(aplValueSteady); ok {
		return max(sim.CurrentTime, steady.steadyUntil(sim))
	}
	if perSecond, until, ok := aplTrend(sim, value); ok && perSecond == 0 {
		return max(sim.CurrentTime, until)
	}
	return sim.CurrentTime
}

func aplTrend(sim *Simulation, value APLValue) (float64, time.Duration, bool) {
	if trend, ok := value.(aplValueTrend); ok {
		return trend.trend(sim)
	}
	return 0, sim.CurrentTime, false
}

// aplNumber is a number value as a float, with durations in seconds.
func aplNumber(sim *Simulation, value APLValue) float64 {
	switch value.Type() {
	case proto.APLValueType_ValueTypeInt:
		return float64(value.GetInt(sim))
	case proto.APLValueType_ValueTypeFloat:
		return value.GetFloat(sim)
	case proto.APLValueType_ValueTypeDuration:
		return value.GetDuration(sim).Seconds()
	}
	return 0
}

func aplIsNumber(value APLValue) bool {
	switch value.Type() {
	case proto.APLValueType_ValueTypeInt, proto.APLValueType_ValueTypeFloat, proto.APLValueType_ValueTypeDuration:
		return true
	}
	return false
}

// aplSteadyTrend is the trend of a number that only changes with an event.
func aplSteadyTrend() (float64, time.Duration, bool) {
	return 0, NeverExpires, true
}

// aplCountdownTrend is the trend of a time left until the given moment, which stops at 0.
func aplCountdownTrend(sim *Simulation, at time.Duration) (float64, time.Duration, bool) {
	if at <= sim.CurrentTime {
		return aplSteadyTrend()
	}
	return -1, at, true
}

func (value *APLValueConst) steadyUntil(_ *Simulation) time.Duration {
	return NeverExpires
}
func (value *APLValueConst) trend(_ *Simulation) (float64, time.Duration, bool) {
	return aplSteadyTrend()
}

func (value *APLValueCoerced) steadyUntil(sim *Simulation) time.Duration {
	if value.inner.Type() == proto.APLValueType_ValueTypeBool {
		return aplSteadyUntil(sim, value.inner)
	}
	if perSecond, until, ok := aplTrend(sim, value.inner); ok && perSecond == 0 {
		return until
	}
	return sim.CurrentTime
}
func (value *APLValueCoerced) trend(sim *Simulation) (float64, time.Duration, bool) {
	if value.inner.Type() == proto.APLValueType_ValueTypeBool {
		return 0, aplSteadyUntil(sim, value.inner), true
	}
	perSecond, until, ok := aplTrend(sim, value.inner)
	// A float or a duration turned into an int is cut to a whole number, so it only keeps the
	// trend while it doesn't change.
	if value.valueType == proto.APLValueType_ValueTypeInt && value.inner.Type() != proto.APLValueType_ValueTypeInt && perSecond != 0 {
		return 0, sim.CurrentTime, false
	}
	return perSecond, until, ok
}

// A comparison of two numbers only turns when their difference crosses 0. We know when that
// happens while both change at a steady rate.
func (value *APLValueCompare) steadyUntil(sim *Simulation) time.Duration {
	if !aplIsNumber(value.lhs) {
		if value.lhs.Type() == proto.APLValueType_ValueTypeBool {
			return min(aplSteadyUntil(sim, value.lhs), aplSteadyUntil(sim, value.rhs))
		}
		return sim.CurrentTime
	}
	lhsPerSecond, lhsUntil, lhsOk := aplTrend(sim, value.lhs)
	rhsPerSecond, rhsUntil, rhsOk := aplTrend(sim, value.rhs)
	if !lhsOk || !rhsOk {
		return sim.CurrentTime
	}
	until := min(lhsUntil, rhsUntil)
	perSecond := lhsPerSecond - rhsPerSecond
	if perSecond == 0 {
		return until
	}
	crossesIn := -(aplNumber(sim, value.lhs) - aplNumber(sim, value.rhs)) / perSecond
	if crossesIn < 0 {
		// They move apart.
		return until
	}
	return max(sim.CurrentTime, min(until, sim.CurrentTime+DurationFromSeconds(crossesIn)-aplIdleMargin))
}

func (value *APLValueMath) trend(sim *Simulation) (float64, time.Duration, bool) {
	lhsPerSecond, lhsUntil, lhsOk := aplTrend(sim, value.lhs)
	rhsPerSecond, rhsUntil, rhsOk := aplTrend(sim, value.rhs)
	if !lhsOk || !rhsOk {
		return 0, sim.CurrentTime, false
	}
	until := min(lhsUntil, rhsUntil)
	switch value.op {
	case proto.APLValueMath_OpAdd:
		return lhsPerSecond + rhsPerSecond, until, true
	case proto.APLValueMath_OpSub:
		return lhsPerSecond - rhsPerSecond, until, true
	}
	// A product or a quotient of numbers that don't change doesn't change either.
	if lhsPerSecond == 0 && rhsPerSecond == 0 {
		return 0, until, true
	}
	return 0, sim.CurrentTime, false
}

func aplSteadyValuesTrend(sim *Simulation, values []APLValue) (float64, time.Duration, bool) {
	until := NeverExpires
	for _, value := range values {
		perSecond, valueUntil, ok := aplTrend(sim, value)
		if !ok || perSecond != 0 {
			return 0, sim.CurrentTime, false
		}
		until = min(until, valueUntil)
	}
	return 0, until, true
}
func (value *APLValueMax) trend(sim *Simulation) (float64, time.Duration, bool) {
	return aplSteadyValuesTrend(sim, value.vals)
}
func (value *APLValueMin) trend(sim *Simulation) (float64, time.Duration, bool) {
	return aplSteadyValuesTrend(sim, value.vals)
}

// An AND that is false stays false while any of its false values does. One that is true stays
// true until any of its values turns.
func (value *APLValueAnd) steadyUntil(sim *Simulation) time.Duration {
	return aplLogicSteadyUntil(sim, value.vals, false)
}

// An OR that is true stays true while any of its true values does. One that is false stays
// false until any of its values turns.
func (value *APLValueOr) steadyUntil(sim *Simulation) time.Duration {
	return aplLogicSteadyUntil(sim, value.vals, true)
}

// aplLogicSteadyUntil is for an AND (deciding is false) or an OR (deciding is true).
func aplLogicSteadyUntil(sim *Simulation, vals []APLValue, deciding bool) time.Duration {
	// The values of the first few, so we don't evaluate them twice.
	var results [16]bool
	decided := false
	for i, val := range vals {
		result := val.GetBool(sim)
		if i < len(results) {
			results[i] = result
		}
		decided = decided || result == deciding
	}
	if decided {
		at := sim.CurrentTime
		for i, val := range vals {
			result := false
			if i < len(results) {
				result = results[i]
			} else {
				result = val.GetBool(sim)
			}
			if result == deciding {
				if at = max(at, aplSteadyUntil(sim, val)); at == NeverExpires {
					break
				}
			}
		}
		return at
	}
	until := NeverExpires
	for _, val := range vals {
		if until = min(until, aplSteadyUntil(sim, val)); until <= sim.CurrentTime {
			break
		}
	}
	return until
}

func (value *APLValueNot) steadyUntil(sim *Simulation) time.Duration {
	return aplSteadyUntil(sim, value.val)
}

func (value *APLValueRemainingTime) trend(sim *Simulation) (float64, time.Duration, bool) {
	// See Simulation.GetRemainingDuration. When the fight ends at a health, after 5s the time
	// left is an estimate from the damage done, which changes with each hit.
	if sim.Encounter.EndFightAtHealth > 0 && sim.Encounter.DurationIsEstimate {
		if sim.CurrentTime >= time.Second*5 {
			return 0, sim.CurrentTime, false
		}
		return -1, time.Second * 5, true
	}
	return -1, NeverExpires, true
}

func (value *APLValueCurrentTime) trend(_ *Simulation) (float64, time.Duration, bool) {
	return 1, NeverExpires, true
}

func (value *APLValueCurrentEnergy) trend(sim *Simulation) (float64, time.Duration, bool) {
	perSecond, until := value.unit.energyBar.energyPerSecond(sim)
	return perSecond, until, true
}

func (value *APLValueCurrentMana) trend(sim *Simulation) (float64, time.Duration, bool) {
	perSecond, until := value.unit.Get().manaBar.manaPerSecond(sim)
	return perSecond, until, true
}

func (value *APLValueCurrentManaPercent) trend(sim *Simulation) (float64, time.Duration, bool) {
	unit := value.unit.Get()
	perSecond, until := unit.manaBar.manaPerSecond(sim)
	return perSecond / unit.MaxMana(), until, true
}

func (value *APLValueCurrentRage) trend(_ *Simulation) (float64, time.Duration, bool) {
	return aplSteadyTrend()
}

func (value *APLValueCurrentComboPoints) trend(_ *Simulation) (float64, time.Duration, bool) {
	return aplSteadyTrend()
}

func (value *APLValueIsExecutePhase) steadyUntil(_ *Simulation) time.Duration {
	return NeverExpires
}

func (value *APLValueAuraIsKnown) steadyUntil(_ *Simulation) time.Duration {
	return NeverExpires
}

// seenChangesAt is when a rotation with the given reaction time next sees the aura or its stacks
// change when nothing happens in the sim: when it sees a proc that came less than the reaction
// time ago.
func (aura *Aura) seenChangesAt(sim *Simulation, reactionTime time.Duration) time.Duration {
	if !aura.IsActive() {
		return NeverExpires
	}
	at := NeverExpires
	if aura.procGained && sim.CurrentTime-aura.procGainedAt < reactionTime {
		at = aura.procGainedAt + reactionTime
	}
	for _, gain := range aura.procStackGains {
		if sim.CurrentTime-gain.at < reactionTime {
			at = min(at, gain.at+reactionTime)
		}
	}
	return at
}

func (value *APLValueAuraIsActive) steadyUntil(sim *Simulation) time.Duration {
	return value.aura.Get().seenChangesAt(sim, value.reactionTime)
}

func (value *APLValueAuraIsActiveWithReactionTime) steadyUntil(sim *Simulation) time.Duration {
	aura := value.aura.Get()
	if aura.IsActive() && aura.TimeActive(sim) < value.reactionTime {
		return aura.startTime + value.reactionTime
	}
	return NeverExpires
}

func (value *APLValueAuraNumStacks) trend(sim *Simulation) (float64, time.Duration, bool) {
	return 0, value.aura.Get().seenChangesAt(sim, value.reactionTime), true
}

// See GetDuration. An aura that ends is an event, but one that isn't up keeps counting down from
// when it last ran out.
func (value *APLValueAuraRemainingTime) trend(sim *Simulation) (float64, time.Duration, bool) {
	aura := value.aura.Get()
	if aura.IsActive() && !aura.SeenActive(sim, value.reactionTime) {
		return 0, aura.seenChangesAt(sim, value.reactionTime), true
	}
	if aura.expires == NeverExpires {
		return aplSteadyTrend()
	}
	return -1, NeverExpires, true
}

func (value *APLValueDotIsActive) steadyUntil(_ *Simulation) time.Duration {
	return NeverExpires
}

func (value *APLValueDotRemainingTime) trend(sim *Simulation) (float64, time.Duration, bool) {
	if value.dot.expires == NeverExpires {
		return aplSteadyTrend()
	}
	return -1, NeverExpires, true
}

func (value *APLValueSequenceIsComplete) steadyUntil(_ *Simulation) time.Duration {
	return NeverExpires
}

func (value *APLValueSpellIsReady) steadyUntil(sim *Simulation) time.Duration {
	if value.spell.IsReady(sim) {
		return NeverExpires
	}
	return value.spell.ReadyAt()
}

func (value *APLValueSpellTimeToReady) trend(sim *Simulation) (float64, time.Duration, bool) {
	return aplCountdownTrend(sim, value.spell.ReadyAt())
}

func (value *APLValueAutoTimeToNext) trend(sim *Simulation) (float64, time.Duration, bool) {
	attacks := value.unit.AutoAttacks
	switch value.autoType {
	case proto.APLValueAutoTimeToNext_Melee:
		return aplCountdownTrend(sim, attacks.NextAttackAt())
	case proto.APLValueAutoTimeToNext_MainHand:
		return aplCountdownTrend(sim, attacks.MainhandSwingAt())
	case proto.APLValueAutoTimeToNext_OffHand:
		return aplCountdownTrend(sim, attacks.OffhandSwingAt())
	case proto.APLValueAutoTimeToNext_Ranged:
		return aplCountdownTrend(sim, attacks.NextRangedAttackAt())
	}
	return aplCountdownTrend(sim, attacks.NextAnyAttackAt())
}

// The swing speed, cast time and cost only change with an aura, which is an event.
func (value *APLValueAutoSwingTime) trend(_ *Simulation) (float64, time.Duration, bool) {
	return aplSteadyTrend()
}
func (value *APLValueSpellCastTime) trend(_ *Simulation) (float64, time.Duration, bool) {
	return aplSteadyTrend()
}
func (value *APLValueSpellCurrentCost) trend(_ *Simulation) (float64, time.Duration, bool) {
	return aplSteadyTrend()
}

func (value *APLValueSpellUsesRemaining) trend(sim *Simulation) (float64, time.Duration, bool) {
	return 0, usesRemainingSteadyUntil(sim, value.spell, value.minTimeLeft, nil), true
}

func (value *APLValueSpellUsesLostByDelay) trend(sim *Simulation) (float64, time.Duration, bool) {
	return 0, min(usesRemainingSteadyUntil(sim, value.spell, value.minTimeLeft, nil), usesRemainingSteadyUntil(sim, value.spell, value.minTimeLeft, value.delay)), true
}

// usesRemainingSteadyUntil is how long spellUsesRemaining keeps its value when nothing happens in
// the sim, or the current time when we can't tell.
//
// The uses are 1 + window / cooldown, where the window is the time left in the fight less the
// time to the first cast. While the spell is on cooldown, both count down together and the
// window stays the same. Otherwise the window shrinks, and the uses drop when it goes below a
// whole number of cooldowns.
func usesRemainingSteadyUntil(sim *Simulation, spell *Spell, minTimeLeftValue APLValue, delayValue APLValue) time.Duration {
	remainingPerSecond, until, ok := (&APLValueRemainingTime{}).trend(sim)
	if !ok || remainingPerSecond != -1 {
		return sim.CurrentTime
	}
	minTimeLeft := minTimeLeftValue.GetDuration(sim)
	until = min(until, aplSteadyUntil(sim, minTimeLeftValue))
	delay := time.Duration(0)
	if delayValue != nil {
		delay = delayValue.GetDuration(sim)
		until = min(until, aplSteadyUntil(sim, delayValue))
	}
	cooldown := spell.CdSpell.CD.Duration
	if cooldown == 0 {
		cooldown = spell.CdSpell.SharedCD.Duration
	}
	timeToReady := spell.TimeToReady(sim)
	if timeToReady > delay {
		return min(until, sim.CurrentTime+timeToReady-delay)
	}
	window := sim.GetRemainingDuration() - minTimeLeft - delay
	if window < 0 {
		return until
	}
	return min(until, sim.CurrentTime+window-time.Duration(window/cooldown)*cooldown)
}
