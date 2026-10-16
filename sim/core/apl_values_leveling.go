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

type APLActionLevelingLookForMob struct {
	defaultAPLActionImpl
	unit  *Unit
	stand bool
}

func (rot *APLRotation) newActionLevelingLookForMob(config *proto.APLActionLevelingLookForMob) APLActionImpl {
	return &APLActionLevelingLookForMob{unit: rot.unit, stand: config.Stand}
}
func (action *APLActionLevelingLookForMob) IsReady(sim *Simulation) bool {
	return sim.Encounter.LevelingCanLookForMob()
}
func (action *APLActionLevelingLookForMob) Execute(sim *Simulation) {
	sim.Encounter.LevelingLookForMob(sim, action.stand)
}
func (action *APLActionLevelingLookForMob) String() string {
	if action.stand {
		return "Look For Mob (Leveling, standing)"
	}
	return "Look For Mob (Leveling)"
}

type APLActionLevelingWalkToPullTarget struct {
	defaultAPLActionImpl
	within APLValue
}

func (rot *APLRotation) newActionLevelingWalkToPullTarget(config *proto.APLActionLevelingWalkToPullTarget) APLActionImpl {
	return &APLActionLevelingWalkToPullTarget{within: rot.coerceTo(rot.newAPLValue(config.Within), proto.APLValueType_ValueTypeFloat)}
}
func (action *APLActionLevelingWalkToPullTarget) withinYards(sim *Simulation) float64 {
	if action.within == nil {
		return 0
	}
	return action.within.GetFloat(sim)
}
func (action *APLActionLevelingWalkToPullTarget) IsReady(sim *Simulation) bool {
	return sim.Encounter.LevelingCanWalkToPullTarget(sim, action.withinYards(sim))
}
func (action *APLActionLevelingWalkToPullTarget) Execute(sim *Simulation) {
	sim.Encounter.LevelingWalkToPullTarget(sim, action.withinYards(sim))
}
func (action *APLActionLevelingWalkToPullTarget) String() string {
	return fmt.Sprintf("Walk To Pull Target (Leveling)(%v)", action.within)
}

// NewAPLFloatValue builds a value from its config as a number, for class actions.
func (rot *APLRotation) NewAPLFloatValue(config *proto.APLValue) APLValue {
	return rot.coerceTo(rot.newAPLValue(config), proto.APLValueType_ValueTypeFloat)
}

// shareValue reads an optional share of max (a "90%" const), 100% when there's none.
func shareValue(sim *Simulation, value APLValue) float64 {
	if value == nil {
		return 1
	}
	return value.GetFloat(sim)
}

type APLActionLevelingRest struct {
	defaultAPLActionImpl
	eat, drink             bool
	healthUntil, manaUntil APLValue
}

func (rot *APLRotation) newActionLevelingRest(config *proto.APLActionLevelingRest) APLActionImpl {
	action := &APLActionLevelingRest{eat: config.Eat, drink: config.Drink}
	if config.HealthUntil != nil {
		action.healthUntil = rot.NewAPLFloatValue(config.HealthUntil)
	}
	if config.ManaUntil != nil {
		action.manaUntil = rot.NewAPLFloatValue(config.ManaUntil)
	}
	return action
}
func (action *APLActionLevelingRest) IsReady(sim *Simulation) bool {
	return sim.Encounter.LevelingCanRest(sim, action.eat, action.drink, shareValue(sim, action.healthUntil), shareValue(sim, action.manaUntil))
}
func (action *APLActionLevelingRest) Execute(sim *Simulation) {
	sim.Encounter.LevelingRest(sim, action.eat, action.drink, shareValue(sim, action.healthUntil), shareValue(sim, action.manaUntil))
}
func (action *APLActionLevelingRest) String() string {
	return fmt.Sprintf("Rest (Leveling)(eat: %v, drink: %v)", action.eat, action.drink)
}

type APLActionLevelingBandage struct {
	defaultAPLActionImpl
}

func (rot *APLRotation) newActionLevelingBandage(_ *proto.APLActionLevelingBandage) APLActionImpl {
	return &APLActionLevelingBandage{}
}
func (action *APLActionLevelingBandage) IsReady(sim *Simulation) bool {
	return sim.Encounter.LevelingCanBandage(sim)
}
func (action *APLActionLevelingBandage) Execute(sim *Simulation) {
	sim.Encounter.LevelingBandage(sim)
}
func (action *APLActionLevelingBandage) String() string {
	return "Bandage (Leveling)"
}

type APLValueLevelingBandageReady struct {
	DefaultAPLValueImpl
}

func (rot *APLRotation) newValueLevelingBandageReady(_ *proto.APLValueLevelingBandageReady) APLValue {
	return &APLValueLevelingBandageReady{}
}
func (value *APLValueLevelingBandageReady) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueLevelingBandageReady) GetBool(sim *Simulation) bool {
	return sim.Encounter.LevelingBandageReady(sim)
}
func (value *APLValueLevelingBandageReady) String() string {
	return "Bandage Ready (Leveling)"
}

type APLValueLevelingMobsThisPull struct {
	DefaultAPLValueImpl
}

func (rot *APLRotation) newValueLevelingMobsThisPull(_ *proto.APLValueLevelingMobsThisPull) APLValue {
	return &APLValueLevelingMobsThisPull{}
}
func (value *APLValueLevelingMobsThisPull) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeInt
}
func (value *APLValueLevelingMobsThisPull) GetInt(sim *Simulation) int32 {
	return int32(sim.Encounter.LevelingMobsThisPull())
}
func (value *APLValueLevelingMobsThisPull) String() string {
	return "Leveling Mobs This Pull()"
}
