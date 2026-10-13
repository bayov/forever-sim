package feral

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

func (cat *FeralDruid) NewAPLValue(rot *core.APLRotation, config *proto.APLValue) core.APLValue {
	switch config.Value.(type) {
	case *proto.APLValue_CatExcessEnergy:
		return cat.newValueCatExcessEnergy(rot, config.GetCatExcessEnergy())
	default:
		return nil
	}
}

type APLValueCatExcessEnergy struct {
	core.DefaultAPLValueImpl
	cat *FeralDruid
}

func (cat *FeralDruid) newValueCatExcessEnergy(rot *core.APLRotation, config *proto.APLValueCatExcessEnergy) core.APLValue {
	return &APLValueCatExcessEnergy{
		cat: cat,
	}
}
func (value *APLValueCatExcessEnergy) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}

// The Energy not set aside for a bleed or finisher refresh. Nothing is set aside yet, so
// it is the current Energy.
func (value *APLValueCatExcessEnergy) GetFloat(sim *core.Simulation) float64 {
	return value.cat.CurrentEnergy()
}
func (value *APLValueCatExcessEnergy) String() string {
	return "Cat Excess Energy()"
}

// The 1.12 "optimal" cat rotation was built on powershifting, which Forever's Furor and
// Shifting Power do away with, so it is gone. The rotations are plain APLs now.
func (cat *FeralDruid) NewAPLAction(rot *core.APLRotation, config *proto.APLAction) core.APLActionImpl {
	return nil
}
