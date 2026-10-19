package core

import (
	"fmt"

	"github.com/wowsims/classic/sim/core/proto"
)

// Talent Points is a number that doesn't change in a sim, so it's a constant to the rest of
// the rotation (and to the idle check skipping, apl_idle.go).
func (rot *APLRotation) newValueTalentPoints(config *proto.APLValueTalentPoints) APLValue {
	value := &APLValueConst{valType: proto.APLValueType_ValueTypeInt, stringVal: fmt.Sprintf("Talent Points(%s)", config.Talent)}
	character := rot.unit.Env.Raid.GetPlayerFromUnit(rot.unit)
	if character == nil {
		return value
	}
	points, ok := character.GetCharacter().TalentPoints(config.Talent)
	if !ok {
		// Read as 0 rather than dropped, so the line it gates doesn't run without the talent.
		rot.ValidationWarning("%s has no talent %q", rot.unit.Label, config.Talent)
	}
	value.intVal = points
	value.floatVal = float64(points)
	return value
}
