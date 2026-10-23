package core

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core/proto"
)

type APLValueChannelClipDelay struct {
	DefaultAPLValueImpl
	unit *Unit
}

func (rot *APLRotation) newValueChannelClipDelay(config *proto.APLValueChannelClipDelay) APLValue {
	return &APLValueChannelClipDelay{
		unit: rot.unit,
	}
}
func (value *APLValueChannelClipDelay) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueChannelClipDelay) GetDuration(sim *Simulation) time.Duration {
	return value.unit.ChannelClipDelay
}
func (value *APLValueChannelClipDelay) String() string {
	return "Channel Clip Delay()"
}

type APLValueFrontOfTarget struct {
	DefaultAPLValueImpl
	unit *Unit
}

func (rot *APLRotation) newValueFrontOfTarget(config *proto.APLValueFrontOfTarget) APLValue {
	return &APLValueFrontOfTarget{
		unit: rot.unit,
	}
}
func (value *APLValueFrontOfTarget) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueFrontOfTarget) GetBool(sim *Simulation) bool {
	return value.unit.PseudoStats.InFrontOfTarget
}
func (value *APLValueFrontOfTarget) String() string {
	return "Front of Target()"
}

type APLValueInMeleeRange struct {
	DefaultAPLValueImpl
	unit *Unit
}

func (rot *APLRotation) newValueInMeleeRange(config *proto.APLValueInMeleeRange) APLValue {
	return &APLValueInMeleeRange{
		unit: rot.unit,
	}
}
func (value *APLValueInMeleeRange) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueInMeleeRange) GetBool(sim *Simulation) bool {
	return value.unit.IsInMeleeRange()
}
func (value *APLValueInMeleeRange) String() string {
	return "In Melee Range()"
}

type APLValueMoveTime struct {
	DefaultAPLValueImpl
	unit      *Unit
	moveRange APLValue
}

func (rot *APLRotation) newValueMoveTime(config *proto.APLValueMoveTime) APLValue {
	moveRange := rot.coerceTo(rot.newAPLValue(config.RangeFromTarget), proto.APLValueType_ValueTypeFloat)
	if moveRange == nil {
		return nil
	}
	return &APLValueMoveTime{
		unit:      rot.unit,
		moveRange: moveRange,
	}
}
func (value *APLValueMoveTime) GetInnerValues() []APLValue {
	return []APLValue{value.moveRange}
}
func (value *APLValueMoveTime) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueMoveTime) GetDuration(sim *Simulation) time.Duration {
	return value.unit.TimeToMoveTo(value.moveRange.GetFloat(sim))
}
func (value *APLValueMoveTime) String() string {
	return fmt.Sprintf("Move Time(%s)", value.moveRange)
}
