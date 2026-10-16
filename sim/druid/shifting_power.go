package druid

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Shifting Power, the Feral talent that takes powershifting's place under Forever
// (1.60.1.70338, wago.tools 1322605): turns 55% of base mana into 40 Energy, Cat Form
// only, on a 16 sec cooldown and a 1 sec global cooldown. Natural Shapeshifter takes 10% a
// point off its cost like Cat Form's, and Improved Shifting Power 4 sec a point off the
// cooldown (1322670).
func (druid *Druid) registerShiftingPowerSpell() {
	if !druid.Talents.ShiftingPower {
		return
	}

	actionID := core.ActionID{SpellID: 1322605}
	energyMetrics := druid.NewEnergyMetrics(actionID)

	// Wolfshead Helm (17768) adds 5 Energy to Shifting Power, through spell 1310990 (20 in
	// 1.60.1.70170, 5 since 70291, development notes 2026-10-08).
	energy := 40.0
	if druid.hasWolfsheadHelm() {
		energy += 5
	}

	druid.ShiftingPower = druid.RegisterSpell(Cat, core.SpellConfig{
		SpellCode: SpellCode_DruidShiftingPower,
		ActionID:  actionID,
		Flags:     core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			BaseCost: 0.55,
			Factor:   druid.naturalShapeshifterFactor(),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: time.Second*16 - time.Second*4*time.Duration(druid.Talents.ImprovedShiftingPower),
			},
		},
		// "Shifting Power can no longer be used while already at Full Energy" (development
		// notes, 2026-10-08, build 1.60.1.70291).
		ExtraCastCondition: func(_ *core.Simulation, _ *core.Unit) bool {
			return druid.CurrentEnergy() < druid.MaxEnergy()
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			druid.AddEnergy(sim, energy, energyMetrics)
		},
	})
}

func (druid *Druid) hasWolfsheadHelm() bool {
	head := druid.Equipment.Head()
	return head != nil && head.ID == WolfsheadHelm
}

// Natural Shapeshifter takes 10% a point off the cost of shapeshifting and Shifting Power.
func (druid *Druid) naturalShapeshifterFactor() float64 {
	return 1 - 0.1*float64(druid.Talents.NaturalShapeshifter)
}
