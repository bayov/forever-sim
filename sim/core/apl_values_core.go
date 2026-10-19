package core

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core/proto"
)

// A DoT value looks up the DoT on its target each time, because the current target changes
// (a leveling session's next mob, or a rotation that swaps targets).
type aplDotRef struct {
	spell  *Spell
	target UnitReference
}

func (rot *APLRotation) newAPLDotRef(targetRef *proto.UnitReference, spellId *proto.ActionID) *aplDotRef {
	target := rot.GetTargetUnit(targetRef)
	if rot.GetAPLDot(target, spellId) == nil {
		return nil
	}
	return &aplDotRef{spell: rot.GetAPLSpell(spellId), target: target}
}

func (ref *aplDotRef) get() *Dot {
	if dot := ref.spell.AOEDot(); dot != nil {
		return dot
	}
	if target := ref.target.Get(); target != nil {
		return ref.spell.Dot(target)
	}
	return ref.spell.CurDot()
}

type APLValueDotIsActive struct {
	DefaultAPLValueImpl
	dot *aplDotRef
}

func (rot *APLRotation) newValueDotIsActive(config *proto.APLValueDotIsActive) APLValue {
	dot := rot.newAPLDotRef(config.TargetUnit, config.SpellId)
	if dot == nil {
		return nil
	}
	return &APLValueDotIsActive{
		dot: dot,
	}
}
func (value *APLValueDotIsActive) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueDotIsActive) GetBool(sim *Simulation) bool {
	dot := value.dot.get()
	return dot != nil && dot.IsActive()
}
func (value *APLValueDotIsActive) String() string {
	return fmt.Sprintf("Dot Is Active(%s)", value.dot.spell.ActionID)
}

type APLValueDotRemainingTime struct {
	DefaultAPLValueImpl
	dot *aplDotRef
}

func (rot *APLRotation) newValueDotRemainingTime(config *proto.APLValueDotRemainingTime) APLValue {
	dot := rot.newAPLDotRef(config.TargetUnit, config.SpellId)
	if dot == nil {
		return nil
	}
	return &APLValueDotRemainingTime{
		dot: dot,
	}
}
func (value *APLValueDotRemainingTime) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueDotRemainingTime) GetDuration(sim *Simulation) time.Duration {
	dot := value.dot.get()
	if dot == nil {
		return 0
	}
	return dot.RemainingDuration(sim)
}
func (value *APLValueDotRemainingTime) String() string {
	return fmt.Sprintf("Dot Remaining Time(%s)", value.dot.spell.ActionID)
}
