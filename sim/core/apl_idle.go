package core

import (
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

// Working out the earliest time an action could be ready costs about as much as two or three
// checks of the priority list. For most rotations it's worth it, because we skip many checks each
// time: a rogue waiting for energy skips about six. But a hunter's auto shots and pet attacks
// come so often that it skips only about two, and then working it out costs more than it saves.
//
// So each time a rotation has tried it aplIdleTries times, we look at how many checks it skipped
// each time. When it's fewer than aplIdleMinSkips, we don't try for the next aplIdlePause
// wake-ups, and then try again. A leveling rotation goes through stretches where it pays and
// stretches where it doesn't, so we don't stop for good. That only changes how fast we run, not
// what the rotation does.
const (
	aplIdleTries    = 200
	aplIdleMinSkips = 3
	aplIdlePause    = 2000
)

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
		if apl.idlePause > 0 {
			apl.idlePause--
			return false
		}
		apl.idleBounded = true
		apl.idleUntil = apl.readyNotBefore(sim)
		apl.idleHopeless = apl.idleUntil <= sim.CurrentTime
		if apl.idleTries++; apl.idleTries == aplIdleTries {
			if apl.idleSkips < aplIdleMinSkips*aplIdleTries {
				apl.idlePause = aplIdlePause
			}
			apl.idleTries, apl.idleSkips = 0, 0
		}
	}
	if sim.CurrentTime < apl.idleUntil {
		apl.idleSkips++
		return true
	}
	return false
}

// noteIdle is for a check that found nothing to do, see canSkipCheck.
func (apl *APLRotation) noteIdle(sim *Simulation) {
	if len(apl.controllingActions) != 0 {
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
	if len(apl.idleBounds) != len(apl.priorityList) {
		apl.idleBounds = make([]time.Duration, len(apl.priorityList))
	}
	bounds := apl.idleBounds
	for i, action := range apl.priorityList {
		if action.bound == nil {
			return sim.CurrentTime
		}
		t := action.bound.readyNotBefore(sim)
		if t <= sim.CurrentTime && action.condition == nil {
			return sim.CurrentTime
		}
		bounds[i] = t
	}

	// We pick the soonest action each time with a scan instead of sorting them all, because we
	// usually stop after a few. We mark an action we looked at with NeverExpires.
	until := NeverExpires
	for {
		next := -1
		for i, t := range bounds {
			if t < until && (next < 0 || t < bounds[next]) {
				next = i
			}
		}
		if next < 0 {
			return until
		}
		t := bounds[next]
		bounds[next] = NeverExpires
		if condition := apl.priorityList[next].condition; condition != nil {
			t = max(t, aplFalseUntil(sim, condition))
		}
		if until = min(until, t); until <= sim.CurrentTime {
			return sim.CurrentTime
		}
	}
}

// aplActionReadyBound is for actions that can tell, without side effects, the earliest time they
// could be ready when nothing happens in the sim before then.
type aplActionReadyBound interface {
	// readyNotBefore returns sim.CurrentTime when the action may be ready now.
	readyNotBefore(sim *Simulation) time.Duration
	// mayBeReadyNow is false when readyNotBefore would return a later time, but it's cheaper.
	mayBeReadyNow(sim *Simulation) bool
}

// APLActionReadyBound, APLValueSteady and APLValueTrend are aplActionReadyBound, aplValueSteady
// and aplValueTrend for the actions and values of class packages, which can't have their
// unexported methods.
type APLActionReadyBound interface {
	ReadyNotBefore(sim *Simulation) time.Duration
	MayBeReadyNow(sim *Simulation) bool
}
type APLValueSteady interface {
	SteadyUntil(sim *Simulation) time.Duration
}
type APLValueTrend interface {
	Trend(sim *Simulation) (perSecond float64, until time.Duration, ok bool)
}

type exportedReadyBound struct {
	APLActionReadyBound
}

func (b exportedReadyBound) readyNotBefore(sim *Simulation) time.Duration {
	return b.ReadyNotBefore(sim)
}
func (b exportedReadyBound) mayBeReadyNow(sim *Simulation) bool {
	return b.MayBeReadyNow(sim)
}

// readyNotBefore is the earliest time the action could be ready when nothing happens in the sim
// before then, or the current time when we can't tell.
func (action *APLAction) readyNotBefore(sim *Simulation) time.Duration {
	if action.bound == nil {
		return sim.CurrentTime
	}
	t := action.bound.readyNotBefore(sim)
	if action.condition == nil {
		return t
	}
	return max(t, aplFalseUntil(sim, action.condition))
}

// aplFalseUntil is how long a condition that is false stays false when nothing happens in the
// sim. It returns the current time when the condition is true, or when we can't tell.
//
// It's the same as checking the condition and then asking aplSteadyUntil, but it looks at each
// value once, and it stops as soon as it finds a false part of an AND that stays false until an
// event, like no mob on us. Working out the bounds costs a lot less that way.
func aplFalseUntil(sim *Simulation, value APLValue) time.Duration {
	return aplHoldsUntil(sim, value, false)
}

// aplHoldsUntil is how long the condition keeps the given result, or the current time when it
// doesn't have it now or we can't tell.
func aplHoldsUntil(sim *Simulation, value APLValue, result bool) time.Duration {
	switch v := value.(type) {
	case *APLValueAnd:
		// A false AND stays false while any of its false values does, and a true one stays true
		// while all of its values do.
		return aplLogicHoldsUntil(sim, v.vals, result, !result)
	case *APLValueOr:
		return aplLogicHoldsUntil(sim, v.vals, result, result)
	case *APLValueNot:
		return aplHoldsUntil(sim, v.val, !result)
	}
	if value.GetBool(sim) != result {
		return sim.CurrentTime
	}
	return aplSteadyUntil(sim, value)
}

// aplLogicHoldsUntil is aplHoldsUntil for the values of an AND or an OR. When any is enough, the
// result holds while any of the values keeps it (an AND that is false, an OR that is true).
// Otherwise it holds while all of them keep it.
func aplLogicHoldsUntil(sim *Simulation, vals []APLValue, result bool, any bool) time.Duration {
	if any {
		until := sim.CurrentTime
		for _, val := range vals {
			if until = max(until, aplHoldsUntil(sim, val, result)); until == NeverExpires {
				break
			}
		}
		return until
	}
	until := NeverExpires
	for _, val := range vals {
		if until = min(until, aplHoldsUntil(sim, val, result)); until <= sim.CurrentTime {
			return sim.CurrentTime
		}
	}
	return until
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

// A cast at the leveling session's Pull Target waits for one, which only comes with an event.
func (action *APLActionCastSpell) readyNotBefore(sim *Simulation) time.Duration {
	if action.target.IsDynamic() && action.target.Get() == nil {
		return NeverExpires
	}
	return action.spell.castableNotBefore(sim, action.target.Get())
}

func (action *APLActionCastSpell) mayBeReadyNow(sim *Simulation) bool {
	return action.spell.castableNow(sim, &action.target)
}

func (action *APLActionChannelSpell) readyNotBefore(sim *Simulation) time.Duration {
	return action.spell.castableNotBefore(sim, action.target.Get())
}
func (action *APLActionChannelSpell) mayBeReadyNow(sim *Simulation) bool {
	return action.spell.castableNow(sim, &action.target)
}

// Moving waits for us to stop moving and to finish a cast, and it only goes somewhere else than
// where we are. We only stop moving with an event, and we only get somewhere with one, like the
// PvP downtime that takes us out of range and back.
func (action *APLActionMove) readyNotBefore(sim *Simulation) time.Duration {
	if sim.CurrentTime < 0 || aplSteadyUntil(sim, action.moveRange) <= sim.CurrentTime {
		return sim.CurrentTime
	}
	if action.unit.IsMoving() || action.moveRange.GetFloat(sim) == action.unit.DistanceFromTarget {
		return NeverExpires
	}
	return max(sim.CurrentTime, action.unit.Hardcast.Expires)
}
func (action *APLActionMove) mayBeReadyNow(sim *Simulation) bool {
	return !action.unit.IsMoving() && !action.unit.IsCasting(sim)
}

// The first major cooldown that is ready and should be used, once the GCD is up.
func (action *APLActionAutocastOtherCooldowns) readyNotBefore(sim *Simulation) time.Duration {
	character := action.character
	// No cooldown goes before this, so once one could go by then, we needn't look at the others.
	floor := max(sim.CurrentTime, character.GCD.ReadyAt(), character.minReady)
	first := NeverExpires
	for _, mcd := range character.majorCooldowns {
		if mcd.ReadyAt() >= first {
			continue
		}
		t := max(mcd.ReadyAt(), mcd.Spell.castableNotBefore(sim, character.CurrentTarget))
		if t <= sim.CurrentTime && mcd.waitsForEvent(sim, character) {
			t = NeverExpires
		}
		if first = min(first, t); first <= floor {
			break
		}
	}
	return max(floor, first)
}

// waitsForEvent is whether a cooldown that is ready won't be used before something happens in
// the sim, and asking about it now has no side effects (see shouldActivateHelper).
//
// A cooldown that is ready gets asked whether it can be cast. For one we lack the mana for, that
// starts the time we count as out of mana, and for one we have the mana for, it ends that time.
// So we only say a cooldown with a mana cost waits for an event when we have the mana for it and
// aren't out of mana. Mana only goes up as time passes, so that stays true until an event.
func (mcd *MajorCooldown) waitsForEvent(sim *Simulation, character *Character) bool {
	if mcd.numUsages < len(mcd.timings) {
		return false
	}
	if cost := mcd.Spell.Cost; cost != nil && cost.CostType() == CostTypeMana &&
		(character.IsOOM() || character.CurrentMana() < cost.GetCurrentCost()) {
		return false
	}
	// We only use survival cooldowns below some health, and health only changes with a hit or a
	// heal.
	if mcd.Type.Matches(CooldownTypeSurvival) && character.CurrentHealthPercent() > character.cooldownConfigs.HpPercentForDefensives {
		return true
	}
	return mcd.ShouldActivateNeedsEvent && !mcd.ShouldActivate(sim, character)
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
//
// It runs for every action of every check, so we only look up the target when the spell's own
// condition needs it.
func (spell *Spell) castableNow(sim *Simulation, target *UnitReference) bool {
	if spell == nil {
		return true
	}
	if spell.DefaultCast.GCD > 0 && !spell.Unit.GCD.IsReady(sim) {
		return false
	}
	if !BothTimersReady(spell.CD.Timer, spell.SharedCD.Timer, sim) {
		return false
	}
	if spell.Cost != nil {
		switch spell.Cost.CostType() {
		case CostTypeEnergy:
			if spell.Unit.CurrentEnergy() < spell.Cost.GetCurrentCost() {
				return false
			}
		case CostTypeRage:
			if spell.Unit.CurrentRage() < spell.Cost.GetCurrentCost() {
				return false
			}
		case CostTypeFocus:
			if spell.Unit.CurrentFocus() < spell.Cost.GetCurrentCost() {
				return false
			}
		case CostTypeMana:
			if spell.Unit.IsOOM() && spell.Unit.CurrentMana() < spell.Cost.GetCurrentCost() {
				return false
			}
		}
	}
	return !spell.extraCastConditionNeedsEvent || spell.ExtraCastCondition(sim, target.Get())
}

// castableNotBefore is the earliest time CanCast could return true when nothing happens in the
// sim before then. It looks at the GCD, the cooldowns, and energy, rage or focus, without side
// effects. It returns the current time when the spell may be castable now.
//
// Rage and focus only come with an event (a hit, a focus tick), so when we lack them, nothing
// changes until one. Under Forever energy comes at a steady rate, so we know when we will have
// enough. Mana also comes at a steady rate, but its check keeps track of the time we spend out of
// mana, see manaBar.castableNotBefore. The spell's own condition we only look at when it says it
// changes only with an event, like Hammer of Wrath's execute phase.
func (spell *Spell) castableNotBefore(sim *Simulation, target *Unit) time.Duration {
	t := sim.CurrentTime
	if spell == nil {
		return t
	}
	if spell.extraCastConditionNeedsEvent && !spell.ExtraCastCondition(sim, target) {
		return NeverExpires
	}
	// We only get in or out of melee range with an event, see APLActionMove.
	if spell.outOfPvPRange() {
		return NeverExpires
	}
	if l := spell.Unit.Env.Encounter.leveling; l != nil {
		if spell.Unit == l.player && l.sitting() {
			return NeverExpires
		}
		if spell.MaxRange > 0 && target != nil && target.Type == EnemyUnit {
			t = max(t, l.inRangeNotBefore(sim, target, spell.MaxRange))
		}
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
	case CostTypeMana:
		t = max(t, spell.Unit.manaBar.castableNotBefore(sim, spell))
	}
	return t
}

// castableNotBefore is the earliest time a spell that costs mana could pass its mana check when
// nothing happens in the sim before then, or the current time when we can't tell.
//
// The check has side effects (ManaCost.MeetsRequirement). When we lack the mana, it starts the
// time we count as out of mana, and when we have it, it ends that time. So we can only tell
// while we're already out of mana and still lack it, because then the check changes nothing,
// and it stays that way until regen brings us the mana.
func (mb *manaBar) castableNotBefore(sim *Simulation, spell *Spell) time.Duration {
	if !mb.IsOOM() {
		return sim.CurrentTime
	}
	cost := spell.Cost.GetCurrentCost()
	if cost <= 0 || mb.unit.CurrentMana() >= cost {
		return sim.CurrentTime
	}
	rate, until := mb.manaPerSecond(sim)
	if rate <= 0 {
		return until
	}
	return min(until, sim.CurrentTime+DurationFromSeconds((cost-mb.unit.CurrentMana())/rate)-aplIdleMargin)
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
	if steady, ok := value.(APLValueSteady); ok {
		return max(sim.CurrentTime, steady.SteadyUntil(sim))
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
	if trend, ok := value.(APLValueTrend); ok {
		return trend.Trend(sim)
	}
	return 0, sim.CurrentTime, false
}

// APLSteadyUntil, APLSteadyTrend and APLCountdownTrend are aplSteadyUntil, aplSteadyTrend and
// aplCountdownTrend for class packages.
func APLSteadyUntil(sim *Simulation, value APLValue) time.Duration {
	return aplSteadyUntil(sim, value)
}
func APLSteadyTrend() (float64, time.Duration, bool) {
	return aplSteadyTrend()
}
func APLCountdownTrend(sim *Simulation, at time.Duration) (float64, time.Duration, bool) {
	return aplCountdownTrend(sim, at)
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

// Health only changes with a hit or a heal.
func (value *APLValueCurrentHealth) trend(_ *Simulation) (float64, time.Duration, bool) {
	return aplSteadyTrend()
}
func (value *APLValueCurrentHealthPercent) trend(_ *Simulation) (float64, time.Duration, bool) {
	return aplSteadyTrend()
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

// We only get in or out of range with an event, see APLActionMove.
func (value *APLValueInMeleeRange) steadyUntil(_ *Simulation) time.Duration {
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
	if dot := value.dot.get(); dot == nil || dot.expires == NeverExpires {
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

// In a leveling session (leveling.go) what the session is doing, the mob we found and the mobs we
// pulled only change with an event. Distances change as we and the mobs move, but a spell out of
// range only makes an action less ready, so the bounds still hold. Only a value that says how far
// something is must look at whether anything moves.

func (value *APLValueIsMoving) steadyUntil(_ *Simulation) time.Duration {
	return NeverExpires
}

func (value *APLValueLevelingDistance) trend(sim *Simulation) (float64, time.Duration, bool) {
	if l := sim.Encounter.leveling; l != nil && !l.still() {
		return 0, sim.CurrentTime, false
	}
	return aplSteadyTrend()
}

func (value *APLValueLevelingMobsRunningIn) trend(_ *Simulation) (float64, time.Duration, bool) {
	return aplSteadyTrend()
}

func (value *APLValueLevelingMobsOnUs) trend(_ *Simulation) (float64, time.Duration, bool) {
	return aplSteadyTrend()
}

// aplOptionalSteadyUntil is aplSteadyUntil for a value an action may leave out.
func aplOptionalSteadyUntil(sim *Simulation, value APLValue) time.Duration {
	if value == nil {
		return NeverExpires
	}
	return aplSteadyUntil(sim, value)
}

func (action *APLActionLevelingLookForMob) readyNotBefore(sim *Simulation) time.Duration {
	if action.IsReady(sim) {
		return sim.CurrentTime
	}
	return NeverExpires
}
func (action *APLActionLevelingLookForMob) mayBeReadyNow(sim *Simulation) bool {
	return action.IsReady(sim)
}

// We only walk up to the Pull Target when we stand, and then the distance to it stays the same.
// While we move we only stop with an event.
func (action *APLActionLevelingWalkToPullTarget) readyNotBefore(sim *Simulation) time.Duration {
	if action.IsReady(sim) {
		return sim.CurrentTime
	}
	return aplOptionalSteadyUntil(sim, action.within)
}
func (action *APLActionLevelingWalkToPullTarget) mayBeReadyNow(sim *Simulation) bool {
	return sim.Encounter.LevelingPullTarget() != nil
}

// A rest we don't need now we don't need until an event either: health only changes with a hit
// or a heal, and mana only goes up as time passes.
func (action *APLActionLevelingRest) readyNotBefore(sim *Simulation) time.Duration {
	if action.IsReady(sim) {
		return sim.CurrentTime
	}
	return min(aplOptionalSteadyUntil(sim, action.healthUntil), aplOptionalSteadyUntil(sim, action.manaUntil))
}
func (action *APLActionLevelingRest) mayBeReadyNow(sim *Simulation) bool {
	s := sim.Encounter.leveling
	return s != nil && s.outOfCombatState() && !s.player.IsCasting(sim)
}

// A bandage waits for Recently Bandaged to run out.
func (action *APLActionLevelingBandage) readyNotBefore(sim *Simulation) time.Duration {
	s := sim.Encounter.leveling
	if s == nil || !s.outOfCombatState() || s.player.IsCasting(sim) || s.player.CurrentHealth() >= s.player.MaxHealth() {
		return NeverExpires
	}
	return max(sim.CurrentTime, s.bandageReadyAt)
}
func (action *APLActionLevelingBandage) mayBeReadyNow(sim *Simulation) bool {
	return sim.Encounter.LevelingBandageReady(sim)
}
