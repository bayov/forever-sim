package core

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core/proto"
)

type APLValueSpellIsKnown struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueSpellIsKnown(config *proto.APLValueSpellIsKnown) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	return &APLValueSpellIsKnown{
		spell: spell,
	}
}
func (value *APLValueSpellIsKnown) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueSpellIsKnown) GetBool(sim *Simulation) bool {
	return value.spell != nil
}
func (value *APLValueSpellIsKnown) String() string {
	return fmt.Sprintf("Is Known(%s)", value.spell.ActionID)
}

type APLValueSpellCanCast struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueSpellCanCast(config *proto.APLValueSpellCanCast) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellCanCast{
		spell: spell,
	}
}
func (value *APLValueSpellCanCast) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueSpellCanCast) GetBool(sim *Simulation) bool {
	return value.spell.CanCast(sim, value.spell.Unit.CurrentTarget)
}
func (value *APLValueSpellCanCast) String() string {
	return fmt.Sprintf("Can Cast(%s)", value.spell.ActionID)
}

type APLValueSpellIsReady struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueSpellIsReady(config *proto.APLValueSpellIsReady) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellIsReady{
		spell: spell,
	}
}
func (value *APLValueSpellIsReady) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueSpellIsReady) GetBool(sim *Simulation) bool {
	return value.spell.IsReady(sim)
}
func (value *APLValueSpellIsReady) String() string {
	return fmt.Sprintf("Is Ready(%s)", value.spell.ActionID)
}

type APLValueSpellTimeToReady struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueSpellTimeToReady(config *proto.APLValueSpellTimeToReady) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellTimeToReady{
		spell: spell,
	}
}
func (value *APLValueSpellTimeToReady) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueSpellTimeToReady) GetDuration(sim *Simulation) time.Duration {
	return value.spell.TimeToReady(sim)
}
func (value *APLValueSpellTimeToReady) String() string {
	return fmt.Sprintf("Time To Ready(%s)", value.spell.ActionID)
}

type APLValueSpellCastTime struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueSpellCastTime(config *proto.APLValueSpellCastTime) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellCastTime{
		spell: spell,
	}
}
func (value *APLValueSpellCastTime) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueSpellCastTime) GetDuration(_ *Simulation) time.Duration {
	return value.spell.CastTime()
}
func (value *APLValueSpellCastTime) String() string {
	return fmt.Sprintf("Cast Time(%s)", value.spell.ActionID)
}

type APLValueSpellTravelTime struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueSpellTravelTime(config *proto.APLValueSpellTravelTime) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellTravelTime{
		spell: spell,
	}
}
func (value *APLValueSpellTravelTime) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueSpellTravelTime) GetDuration(_ *Simulation) time.Duration {
	return value.spell.TravelTime()
}
func (value *APLValueSpellTravelTime) String() string {
	return fmt.Sprintf("Travel Time(%s)", value.spell.ActionID)
}

type APLValueSpellInFlight struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueSpellInFlight(config *proto.APLValueSpellInFlight) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellInFlight{
		spell: spell,
	}
}
func (value *APLValueSpellInFlight) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueSpellInFlight) GetBool(sim *Simulation) bool {
	return value.spell.MissileSpeed > 0 && value.spell.casts > 0 && (sim.CurrentTime <= value.spell.LastCastAt+value.spell.TravelTime())
}
func (value *APLValueSpellInFlight) String() string {
	return fmt.Sprintf("In Flight(%s)", value.spell.ActionID)
}

type APLValueSpellCPM struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueSpellCPM(config *proto.APLValueSpellCPM) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellCPM{
		spell: spell,
	}
}
func (value *APLValueSpellCPM) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueSpellCPM) GetFloat(sim *Simulation) float64 {
	return value.spell.CurCPM(sim)
}
func (value *APLValueSpellCPM) String() string {
	return fmt.Sprintf("CPM(%s)", value.spell.ActionID)
}

type APLValueSpellIsChanneling struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueSpellIsChanneling(config *proto.APLValueSpellIsChanneling) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellIsChanneling{
		spell: spell,
	}
}
func (value *APLValueSpellIsChanneling) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueSpellIsChanneling) GetBool(_ *Simulation) bool {
	return value.spell.Unit.ChanneledDot != nil && value.spell.Unit.ChanneledDot.Spell == value.spell
}
func (value *APLValueSpellIsChanneling) String() string {
	return fmt.Sprintf("IsChanneling(%s)", value.spell.ActionID)
}

type APLValueSpellChanneledTicks struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueSpellChanneledTicks(config *proto.APLValueSpellChanneledTicks) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellChanneledTicks{
		spell: spell,
	}
}
func (value *APLValueSpellChanneledTicks) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeInt
}
func (value *APLValueSpellChanneledTicks) GetInt(_ *Simulation) int32 {
	channeledDot := value.spell.Unit.ChanneledDot
	if channeledDot == nil {
		return 0
	} else {
		return channeledDot.TickCount
	}
}
func (value *APLValueSpellChanneledTicks) String() string {
	return fmt.Sprintf("ChanneledTicks(%s)", value.spell.ActionID)
}

type APLValueSpellCurrentCost struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueSpellCurrentCost(config *proto.APLValueSpellCurrentCost) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellCurrentCost{
		spell: spell,
	}
}
func (value *APLValueSpellCurrentCost) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueSpellCurrentCost) GetFloat(_ *Simulation) float64 {
	spell := value.spell
	if spell.Cost == nil {
		return 0
	}
	return spell.Cost.GetCurrentCost()
}
func (value *APLValueSpellCurrentCost) String() string {
	return fmt.Sprintf("CurrentCost(%s)", value.spell.ActionID)
}

// Casts of a cooldown that still fit in the fight when the next one waits at least delay
// and every one after it goes out the moment it is ready. A cast only counts when it
// lands with minTimeLeft of fight still to go, so a 15s buff can ask for 15s and not
// count a cast that would run past the end.
//
// The live cooldown timer is the estimate for the next cast. It only knows about
// reductions that already happened, so with Restless Blades the count is pessimistic
// early in the fight and settles as the cooldown gets closer.
func spellUsesRemaining(sim *Simulation, spell *Spell, delay time.Duration, minTimeLeft time.Duration) int32 {
	cooldown := spell.CdSpell.CD.Duration
	if cooldown == 0 {
		cooldown = spell.CdSpell.SharedCD.Duration
	}
	firstCast := max(spell.TimeToReady(sim), delay)
	window := sim.GetRemainingDuration() - minTimeLeft - firstCast
	if window < 0 {
		return 0
	}
	return 1 + int32(window/cooldown)
}

func (rot *APLRotation) getCooldownSpell(spellId *proto.ActionID) *Spell {
	spell := rot.GetAPLSpell(spellId)
	if spell == nil {
		return nil
	}
	if spell.CdSpell.CD.Duration == 0 && spell.CdSpell.SharedCD.Duration == 0 {
		rot.ValidationWarning("%s has no cooldown, so it has no number of uses", spell.ActionID)
		return nil
	}
	return spell
}

func (rot *APLRotation) durationValueOrZero(config *proto.APLValue) APLValue {
	value := rot.coerceTo(rot.newAPLValue(config), proto.APLValueType_ValueTypeDuration)
	if value == nil {
		value = rot.newValueConst(&proto.APLValueConst{Val: "0s"})
	}
	return value
}

type APLValueSpellUsesRemaining struct {
	DefaultAPLValueImpl
	spell       *Spell
	minTimeLeft APLValue
}

func (rot *APLRotation) newValueSpellUsesRemaining(config *proto.APLValueSpellUsesRemaining) APLValue {
	spell := rot.getCooldownSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellUsesRemaining{
		spell:       spell,
		minTimeLeft: rot.durationValueOrZero(config.MinTimeLeft),
	}
}
func (value *APLValueSpellUsesRemaining) GetInnerValues() []APLValue {
	return []APLValue{value.minTimeLeft}
}
func (value *APLValueSpellUsesRemaining) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeInt
}
func (value *APLValueSpellUsesRemaining) GetInt(sim *Simulation) int32 {
	return spellUsesRemaining(sim, value.spell, 0, value.minTimeLeft.GetDuration(sim))
}
func (value *APLValueSpellUsesRemaining) String() string {
	return fmt.Sprintf("Uses Remaining(%s)", value.spell.ActionID)
}

// How many casts of a cooldown are lost by holding it for delay. Zero means the hold is
// free, which is the test for lining a short cooldown up with a longer one.
type APLValueSpellUsesLostByDelay struct {
	DefaultAPLValueImpl
	spell       *Spell
	delay       APLValue
	minTimeLeft APLValue
}

func (rot *APLRotation) newValueSpellUsesLostByDelay(config *proto.APLValueSpellUsesLostByDelay) APLValue {
	spell := rot.getCooldownSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	// A delay that is given but does not resolve (the time to ready of a cooldown the
	// character does not have, like Adrenaline Rush below level 40) takes the whole value
	// out, the same as a missing spell does. A delay of zero instead would make "uses
	// lost > 0" false forever, and a Blood Fury waiting on it would never be cast.
	var delay APLValue
	if config.Delay != nil {
		delay = rot.coerceTo(rot.newAPLValue(config.Delay), proto.APLValueType_ValueTypeDuration)
		if delay == nil {
			return nil
		}
	} else {
		delay = rot.durationValueOrZero(nil)
	}
	return &APLValueSpellUsesLostByDelay{
		spell:       spell,
		delay:       delay,
		minTimeLeft: rot.durationValueOrZero(config.MinTimeLeft),
	}
}
func (value *APLValueSpellUsesLostByDelay) GetInnerValues() []APLValue {
	return []APLValue{value.delay, value.minTimeLeft}
}
func (value *APLValueSpellUsesLostByDelay) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeInt
}
func (value *APLValueSpellUsesLostByDelay) GetInt(sim *Simulation) int32 {
	minTimeLeft := value.minTimeLeft.GetDuration(sim)
	return spellUsesRemaining(sim, value.spell, 0, minTimeLeft) - spellUsesRemaining(sim, value.spell, value.delay.GetDuration(sim), minTimeLeft)
}
func (value *APLValueSpellUsesLostByDelay) String() string {
	return fmt.Sprintf("Uses Lost By Delay(%s, %s)", value.spell.ActionID, value.delay)
}
