package core

import (
	"fmt"

	"github.com/wowsims/classic/sim/core/proto"
)

// APL values for a leveling session (leveling.go), for the rotation to decide how to pull.

type APLValueIsMoving struct {
	DefaultAPLValueImpl
	unit *Unit
}

func (rot *APLRotation) newValueIsMoving(_ *proto.APLValueIsMoving) APLValue {
	return &APLValueIsMoving{unit: rot.unit}
}
func (value *APLValueIsMoving) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueIsMoving) GetBool(_ *Simulation) bool {
	return value.unit.IsMoving()
}
func (value *APLValueIsMoving) String() string {
	return "Is Moving()"
}

type APLValueLevelingDistance struct {
	DefaultAPLValueImpl
	target UnitReference
}

func (rot *APLRotation) newValueLevelingDistance(config *proto.APLValueLevelingDistance) APLValue {
	return &APLValueLevelingDistance{target: rot.GetTargetUnit(config.TargetUnit)}
}
func (value *APLValueLevelingDistance) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueLevelingDistance) GetFloat(sim *Simulation) float64 {
	return sim.Encounter.LevelingDistance(sim, value.target.Get())
}
func (value *APLValueLevelingDistance) String() string {
	return fmt.Sprintf("Leveling Distance(%s)", value.target.String())
}

type APLValueLevelingMobsRunningIn struct {
	DefaultAPLValueImpl
}

func (rot *APLRotation) newValueLevelingMobsRunningIn(_ *proto.APLValueLevelingMobsRunningIn) APLValue {
	return &APLValueLevelingMobsRunningIn{}
}
func (value *APLValueLevelingMobsRunningIn) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeInt
}
func (value *APLValueLevelingMobsRunningIn) GetInt(sim *Simulation) int32 {
	running, _ := sim.Encounter.LevelingMobCounts()
	return int32(running)
}
func (value *APLValueLevelingMobsRunningIn) String() string {
	return "Leveling Mobs Running In()"
}

type APLValueLevelingMobsOnUs struct {
	DefaultAPLValueImpl
}

func (rot *APLRotation) newValueLevelingMobsOnUs(_ *proto.APLValueLevelingMobsOnUs) APLValue {
	return &APLValueLevelingMobsOnUs{}
}
func (value *APLValueLevelingMobsOnUs) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeInt
}
func (value *APLValueLevelingMobsOnUs) GetInt(sim *Simulation) int32 {
	_, onUs := sim.Encounter.LevelingMobCounts()
	return int32(onUs)
}
func (value *APLValueLevelingMobsOnUs) String() string {
	return "Leveling Mobs On Us()"
}
