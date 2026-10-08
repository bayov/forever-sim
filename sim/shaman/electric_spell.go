package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Totem Item IDs
const (
	StormfuryTotem           = 31031
	TotemOfAncestralGuidance = 32330
	TotemOfStorms            = 23199
	TotemOfTheVoid           = 28248
	TotemOfHex               = 40267
	VentureCoLightningRod    = 38361
	ThunderfallTotem         = 45255
)

// Shared precomputation logic for LB and CL.
// stopMeleeForCast holds our swings until a spell's cast ends, and the swing timer starts
// over from there.
//
// An instant cast leaves the swing timer alone, like a shock (the user, 2026-10-08). That's
// Lightning Bolt with 5 Maelstrom Weapon stacks or Nature's Swiftness.
func (shaman *Shaman) stopMeleeForCast(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
	if castTime := shaman.ApplyCastSpeedForSpell(cast.CastTime, spell); castTime > 0 {
		shaman.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime+castTime, false)
	}
}

func (shaman *Shaman) newElectricSpellConfig(actionID core.ActionID, baseCost float64, baseCastTime time.Duration) core.SpellConfig {
	spell := core.SpellConfig{
		ActionID:     actionID,
		SpellSchool:  core.SpellSchoolNature,
		DefenseType:  core.DefenseTypeMagic,
		ProcMask:     core.ProcMaskSpellDamage,
		Flags:        SpellFlagShaman | SpellFlagLightning | core.SpellFlagAPL,
		MetricSplits: 6,

		ManaCost: core.ManaCostOptions{
			FlatCost:   baseCost,
			Multiplier: 100 - 2*shaman.Talents.Convection,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				CastTime: baseCastTime - shaman.elementalAlacrityReduction(),
				GCD:      core.GCDDefault,
			},
			ModifyCast: shaman.stopMeleeForCast,
		},

		BonusCritRating: core.TernaryFloat64(shaman.Talents.CallOfThunder, 3, 0) * core.SpellCritRatingPerCritChance,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
	}

	return spell
}
