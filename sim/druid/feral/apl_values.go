package feral

import (
	"cmp"
	"slices"
	"time"

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

// The Energy the cat can spend now without delaying a coming Rip or Rake refresh, as
// wowsims/forever has it (excess_energy.go, d9e8603294).
//
// A refresh counts when its bleed is up and the fight lasts long enough after it runs out for a
// new one to be worth casting: Rip at 5 combo points with 10 sec of fight left after it, Rake
// with a full Rake left. Taking the refreshes in time order, the Energy regenerated before each
// one pays for it, and the first one after Shifting Power is ready costs Shifting Power's Energy
// less; what is left has to be kept from now.
func (value *APLValueCatExcessEnergy) GetFloat(sim *core.Simulation) float64 {
	return value.cat.ExcessEnergy(sim)
}

type pendingRefresh struct {
	at   time.Duration
	cost float64
}

func (cat *FeralDruid) ExcessEnergy(sim *core.Simulation) float64 {
	target := cat.CurrentTarget
	remaining := sim.GetRemainingDuration()
	refreshes := make([]pendingRefresh, 0, 2)
	if cat.Rip != nil && cat.ComboPoints() >= 5 {
		if dot := cat.Rip.Dot(target); dot.IsActive() && dot.RemainingDuration(sim) <= remaining-10*time.Second {
			refreshes = append(refreshes, pendingRefresh{dot.ExpiresAt(), cat.Rip.DefaultCast.Cost})
		}
	}
	if cat.Rake != nil {
		if dot := cat.Rake.Dot(target); dot.IsActive() && dot.RemainingDuration(sim) <= remaining-dot.Duration {
			refreshes = append(refreshes, pendingRefresh{dot.ExpiresAt(), cat.Rake.DefaultCast.Cost})
		}
	}
	slices.SortStableFunc(refreshes, func(a, b pendingRefresh) int { return cmp.Compare(a.at, b.at) })

	regen := core.EnergyPerSecond * cat.EnergyTickMultiplier
	if !sim.IsForever() {
		regen = core.EnergyPerTick * cat.EnergyTickMultiplier / core.EnergyTickDuration.Seconds()
	}
	// When Shifting Power gives more than the refresh costs, the cost goes below zero and the
	// Energy left over goes to the next refresh.
	shiftEnergy := 0.0
	shiftReadyAt := core.NeverExpires
	if cat.ShiftingPower != nil {
		shiftEnergy = cat.ShiftingPowerEnergy
		shiftReadyAt = cat.ShiftingPower.ReadyAt()
	}
	keep := 0.0
	// The time up to which the regeneration has already paid for earlier refreshes.
	regenUsedUntil := sim.CurrentTime
	for _, refresh := range refreshes {
		if shiftEnergy > 0 && shiftReadyAt <= refresh.at {
			refresh.cost -= shiftEnergy
			shiftEnergy = 0
		}
		if regen <= 0 {
			keep += max(0, refresh.cost)
			continue
		}
		if gain := (refresh.at - regenUsedUntil).Seconds() * regen; gain < refresh.cost {
			// Regeneration falls short: keep the difference, and all of it up to this refresh is used.
			keep += refresh.cost - gain
			regenUsedUntil = refresh.at
		} else {
			// Regeneration covers it: only the time it takes to pay is used. A cost below zero
			// gives time back, so the next refresh gets that much more regeneration.
			regenUsedUntil += core.DurationFromSeconds(refresh.cost / regen)
		}
	}
	return cat.CurrentEnergy() - keep
}

func (value *APLValueCatExcessEnergy) String() string {
	return "Cat Excess Energy()"
}

// The 1.12 "optimal" cat rotation was built on powershifting, which Forever's Furor and
// Shifting Power do away with, so it is gone. The rotations are plain APLs now.
func (cat *FeralDruid) NewAPLAction(rot *core.APLRotation, config *proto.APLAction) core.APLActionImpl {
	return nil
}
