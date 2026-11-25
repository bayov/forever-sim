package core

import (
	"time"

	"github.com/wowsims/classic/sim/core/proto"
)

type APLValueAutoTimeToNext struct {
	DefaultAPLValueImpl
	unit     *Unit
	autoType proto.APLValueAutoTimeToNext_AttackType
}

func (rot *APLRotation) newValueAutoTimeToNext(config *proto.APLValueAutoTimeToNext) APLValue {
	return &APLValueAutoTimeToNext{
		unit:     rot.unit,
		autoType: config.AutoType,
	}
}
func (value *APLValueAutoTimeToNext) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueAutoTimeToNext) GetDuration(sim *Simulation) time.Duration {
	switch value.autoType {
	case proto.APLValueAutoTimeToNext_Melee:
		return max(0, min(value.unit.AutoAttacks.MainhandScheduledSwingAt(), value.unit.AutoAttacks.OffhandSwingAt())-sim.CurrentTime)
	case proto.APLValueAutoTimeToNext_MainHand:
		return max(0, value.unit.AutoAttacks.MainhandScheduledSwingAt()-sim.CurrentTime)
	case proto.APLValueAutoTimeToNext_OffHand:
		return max(0, value.unit.AutoAttacks.OffhandSwingAt()-sim.CurrentTime)
	case proto.APLValueAutoTimeToNext_Ranged:
		return max(0, value.unit.AutoAttacks.NextRangedAttackAt()-sim.CurrentTime)
	}
	// defaults to Any
	aa := value.unit.AutoAttacks
	return max(0, min(aa.MainhandScheduledSwingAt(), aa.OffhandSwingAt(), aa.NextRangedAttackAt())-sim.CurrentTime)
}
func (value *APLValueAutoTimeToNext) String() string {
	return "Auto Time To Next"
}

type APLValueAutoSwingTime struct {
	DefaultAPLValueImpl
	unit     *Unit
	autoType proto.APLValueAutoSwingTime_SwingType
}

func (rot *APLRotation) newValueAutoSwingTime(config *proto.APLValueAutoSwingTime) APLValue {
	return &APLValueAutoSwingTime{
		unit:     rot.unit,
		autoType: config.AutoType,
	}
}
func (value *APLValueAutoSwingTime) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueAutoSwingTime) GetDuration(sim *Simulation) time.Duration {
	switch value.autoType {
	case proto.APLValueAutoSwingTime_MainHand:
		return max(0, value.unit.AutoAttacks.MainhandSwingSpeed())
	case proto.APLValueAutoSwingTime_OffHand:
		return max(0, value.unit.AutoAttacks.OffhandSwingSpeed())
	case proto.APLValueAutoSwingTime_Ranged:
		return max(0, value.unit.AutoAttacks.RangedSwingSpeed())
	}
	// defaults to 0
	return 0
}
func (value *APLValueAutoSwingTime) String() string {
	return "Auto Swing Time"
}
