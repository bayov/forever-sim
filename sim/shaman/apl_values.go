package shaman

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

func (shaman *Shaman) NewAPLValue(rot *core.APLRotation, config *proto.APLValue) core.APLValue {
	switch config.Value.(type) {
	case *proto.APLValue_TotemRemainingTime:
		return shaman.newValueTotemRemainingTime(rot, config.GetTotemRemainingTime())
	case *proto.APLValue_TotemDistance:
		return shaman.newValueTotemDistance(rot, config.GetTotemDistance())
	default:
		return nil
	}
}

func (shaman *Shaman) NewAPLAction(rot *core.APLRotation, config *proto.APLAction) core.APLActionImpl {
	switch config.Action.(type) {
	case *proto.APLAction_ShamanWalkToTotem:
		return shaman.newActionWalkToTotem(rot, config.GetShamanWalkToTotem())
	case *proto.APLAction_CastAssignedTotem:
		return shaman.newActionCastAssignedTotem(rot, config.GetCastAssignedTotem())
	default:
		return nil
	}
}

// newActionCastAssignedTotem casts our assigned totem of the element (Player.assigned_totems)
// at the highest rank we know, as a plain Cast Spell. So two shamans with one rotation drop
// the totems the raid gave each of them. Without one assigned there, the action is left out.
func (shaman *Shaman) newActionCastAssignedTotem(rot *core.APLRotation, config *proto.APLActionCastAssignedTotem) core.APLActionImpl {
	if totemSlot(config.TotemType) < 0 {
		rot.ValidationWarning("Totem Type required.")
		return nil
	}
	spell := shaman.assignedTotem(config.TotemType)
	if spell == nil {
		rot.ValidationWarning("No %s totem assigned, or none known at this level.", config.TotemType)
		return nil
	}
	return rot.NewCastSpellAction(spell)
}

// assignedTotem is the highest rank we know of our assigned totem of the element, nil without
// one.
func (shaman *Shaman) assignedTotem(totemType proto.ShamanTotems_TotemType) *core.Spell {
	assigned := shaman.AssignedTotems
	if assigned == nil {
		return nil
	}
	var ranks []*core.Spell
	switch totemType {
	case proto.ShamanTotems_Earth:
		switch assigned.Earth {
		case proto.EarthTotem_StrengthOfEarthTotem:
			ranks = shaman.StrengthOfEarthTotem
		case proto.EarthTotem_StoneskinTotem:
			ranks = shaman.StoneskinTotem
		case proto.EarthTotem_TremorTotem:
			ranks = []*core.Spell{shaman.TremorTotem}
		}
	case proto.ShamanTotems_Air:
		switch assigned.Air {
		case proto.AirTotem_WindfuryTotem:
			ranks = shaman.WindfuryTotem
		case proto.AirTotem_GraceOfAirTotem:
			ranks = shaman.GraceOfAirTotem
		}
	case proto.ShamanTotems_Fire:
		switch assigned.Fire {
		case proto.FireTotem_SearingTotem:
			ranks = shaman.SearingTotem
		case proto.FireTotem_MagmaTotem:
			ranks = shaman.MagmaTotem
		case proto.FireTotem_FireNovaTotem:
			ranks = shaman.FireNova
		case proto.FireTotem_FlametongueTotem:
			ranks = shaman.FlametongueTotem
		}
	case proto.ShamanTotems_Water:
		switch assigned.Water {
		case proto.WaterTotem_ManaSpringTotem:
			ranks = shaman.ManaSpringTotem
		case proto.WaterTotem_HealingStreamTotem:
			ranks = shaman.HealingStreamTotem
		}
	}
	for i := len(ranks) - 1; i >= 0; i-- {
		if ranks[i] != nil {
			return ranks[i]
		}
	}
	return nil
}

// totemSlot is the totem slot of an APL totem type, -1 for none.
func totemSlot(totemType proto.ShamanTotems_TotemType) int {
	switch totemType {
	case proto.ShamanTotems_Earth:
		return EarthTotem
	case proto.ShamanTotems_Air:
		return AirTotem
	case proto.ShamanTotems_Fire:
		return FireTotem
	case proto.ShamanTotems_Water:
		return WaterTotem
	}
	return -1
}

// totemIsDown is whether the slot has a totem standing.
func (shaman *Shaman) totemIsDown(sim *core.Simulation, slot int) bool {
	return shaman.ActiveTotems[slot] != nil && shaman.TotemExpirations[slot] > sim.CurrentTime
}

// totemDistance is how far we are from the slot's totem in a leveling session, 0 outside one
// or without that totem.
func (shaman *Shaman) totemDistance(sim *core.Simulation, slot int) float64 {
	if !shaman.totemIsDown(sim, slot) {
		return 0
	}
	return sim.Encounter.LevelingPosition(sim).Dist(shaman.totemPositions[slot])
}

type APLValueTotemDistance struct {
	core.DefaultAPLValueImpl
	shaman    *Shaman
	totemType proto.ShamanTotems_TotemType
}

func (shaman *Shaman) newValueTotemDistance(rot *core.APLRotation, config *proto.APLValueTotemDistance) core.APLValue {
	if totemSlot(config.TotemType) < 0 {
		rot.ValidationWarning("Totem Type required.")
		return nil
	}
	return &APLValueTotemDistance{shaman: shaman, totemType: config.TotemType}
}
func (value *APLValueTotemDistance) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueTotemDistance) GetFloat(sim *core.Simulation) float64 {
	return value.shaman.totemDistance(sim, totemSlot(value.totemType))
}
func (value *APLValueTotemDistance) String() string {
	return fmt.Sprintf("Totem Distance(%s)", value.totemType.String())
}

type APLActionWalkToTotem struct {
	shaman *Shaman
	slot   int
	within core.APLValue
}

func (shaman *Shaman) newActionWalkToTotem(rot *core.APLRotation, config *proto.APLActionShamanWalkToTotem) core.APLActionImpl {
	slot := totemSlot(config.TotemType)
	if slot < 0 {
		rot.ValidationWarning("Totem Type required.")
		return nil
	}
	action := &APLActionWalkToTotem{shaman: shaman, slot: slot}
	if config.Within != nil {
		action.within = rot.NewAPLFloatValue(config.Within)
	}
	return action
}
func (action *APLActionWalkToTotem) GetInnerActions() []*core.APLAction             { return nil }
func (action *APLActionWalkToTotem) GetAPLValues() []core.APLValue                  { return nil }
func (action *APLActionWalkToTotem) Finalize(*core.APLRotation)                     {}
func (action *APLActionWalkToTotem) Reset(*core.Simulation)                         {}
func (action *APLActionWalkToTotem) GetNextAction(*core.Simulation) *core.APLAction { return nil }
func (action *APLActionWalkToTotem) withinYards(sim *core.Simulation) float64 {
	if action.within == nil {
		return 0
	}
	return action.within.GetFloat(sim)
}
func (action *APLActionWalkToTotem) IsReady(sim *core.Simulation) bool {
	return action.shaman.totemIsDown(sim, action.slot) &&
		sim.Encounter.LevelingCanWalkTo(sim, action.shaman.totemPositions[action.slot], action.withinYards(sim))
}
func (action *APLActionWalkToTotem) Execute(sim *core.Simulation) {
	sim.Encounter.LevelingWalkTo(sim, action.shaman.totemPositions[action.slot], action.withinYards(sim))
}
func (action *APLActionWalkToTotem) String() string {
	return fmt.Sprintf("Walk To Totem(%d)", action.slot)
}

type APLValueTotemRemainingTime struct {
	core.DefaultAPLValueImpl
	shaman    *Shaman
	totemType proto.ShamanTotems_TotemType
}

func (shaman *Shaman) newValueTotemRemainingTime(rot *core.APLRotation, config *proto.APLValueTotemRemainingTime) core.APLValue {
	if config.TotemType == proto.ShamanTotems_TypeUnknown {
		rot.ValidationWarning("Totem Type required.")
		return nil
	}
	return &APLValueTotemRemainingTime{
		shaman:    shaman,
		totemType: config.TotemType,
	}
}
func (value *APLValueTotemRemainingTime) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueTotemRemainingTime) GetDuration(sim *core.Simulation) time.Duration {
	if value.totemType == proto.ShamanTotems_Earth {
		return max(0, value.shaman.TotemExpirations[EarthTotem]-sim.CurrentTime)
	} else if value.totemType == proto.ShamanTotems_Air {
		return max(0, value.shaman.TotemExpirations[AirTotem]-sim.CurrentTime)
	} else if value.totemType == proto.ShamanTotems_Fire {
		return max(0, value.shaman.TotemExpirations[FireTotem]-sim.CurrentTime)
	} else if value.totemType == proto.ShamanTotems_Water {
		return max(0, value.shaman.TotemExpirations[WaterTotem]-sim.CurrentTime)
	} else {
		return 0
	}
}
func (value *APLValueTotemRemainingTime) String() string {
	return fmt.Sprintf("Totem Remaining Time(%s)", value.totemType.String())
}

// The totem's time left counts down until it expires, see core.APLValueTrend.
func (value *APLValueTotemRemainingTime) Trend(sim *core.Simulation) (float64, time.Duration, bool) {
	slot := totemSlot(value.totemType)
	if slot < 0 {
		return core.APLSteadyTrend()
	}
	return core.APLCountdownTrend(sim, value.shaman.TotemExpirations[slot])
}

// A totem stays where we dropped it, so the distance to it only changes while we move. Once the
// totem expires the distance is 0, which happens as time passes, without an event.
func (value *APLValueTotemDistance) Trend(sim *core.Simulation) (float64, time.Duration, bool) {
	if value.shaman.IsMoving() {
		return 0, sim.CurrentTime, false
	}
	slot := totemSlot(value.totemType)
	if !value.shaman.totemIsDown(sim, slot) {
		return core.APLSteadyTrend()
	}
	return 0, value.shaman.TotemExpirations[slot], true
}

// We only walk to a totem that stands, which only drops with an event. We only walk when we
// stand, and then the distance to it stays the same. While we move we only stop with an event.
func (action *APLActionWalkToTotem) ReadyNotBefore(sim *core.Simulation) time.Duration {
	if action.IsReady(sim) {
		return sim.CurrentTime
	}
	if action.within == nil {
		return core.NeverExpires
	}
	return core.APLSteadyUntil(sim, action.within)
}
func (action *APLActionWalkToTotem) MayBeReadyNow(sim *core.Simulation) bool {
	return action.shaman.totemIsDown(sim, action.slot)
}
