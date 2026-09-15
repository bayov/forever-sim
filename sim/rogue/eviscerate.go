package rogue

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (rogue *Rogue) registerEviscerate() {
	flatDamage := rankAt(rogue.Level, map[int32]float64{
		1: 1, 8: 3, 16: 6, 24: 10, 32: 15, 40: 22, 48: 34, 56: 48,
		60: core.TernaryFloat64(core.IncludeAQ, 54, 48),
	})
	comboDamageBonus := rankAt(rogue.Level, map[int32]float64{
		1: 5, 8: 11, 16: 19, 24: 31, 32: 45, 40: 77, 48: 110, 56: 151,
		60: core.TernaryFloat64(core.IncludeAQ, 170, 151),
	})
	damageVariance := rankAt(rogue.Level, map[int32]float64{
		1: 4, 8: 8, 16: 14, 24: 20, 32: 30, 40: 44, 48: 68, 56: 96,
		60: core.TernaryFloat64(core.IncludeAQ, 108, 96),
	})
	spellID := rankSpellID(rogue.Level, map[int32]int32{
		1: 2098, 8: 6760, 16: 6761, 24: 6762, 32: 8623, 40: 8624, 48: 11299, 56: 11300,
		60: core.TernaryInt32(core.IncludeAQ, 31016, 11300),
	})
	core.RegisterSpellRanks(11300, 31016)

	rogue.Eviscerate = rogue.RegisterSpell(core.SpellConfig{
		SpellCode:    SpellCode_RogueEviscerate,
		ActionID:     core.ActionID{SpellID: spellID},
		SpellSchool:  core.SpellSchoolPhysical,
		DefenseType:  core.DefenseTypeMelee,
		ProcMask:     core.ProcMaskMeleeMHSpecial,
		Flags:        rogue.finisherFlags() | SpellFlagColdBlooded,
		MetricSplits: 6,

		EnergyCost: core.EnergyCostOptions{
			Cost:   35,
			Refund: 0,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
			ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
				spell.SetMetricsSplit(spell.Unit.ComboPoints())
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return rogue.ComboPoints() > 0
		},

		DamageMultiplier: 1 +
			[]float64{0, 0.07, 0.14, 0.21}[rogue.Talents.ImprovedEviscerate] +
			[]float64{0, 0.02, 0.04, 0.06}[rogue.Talents.Aggression],
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)

			comboPoints := rogue.ComboPoints()
			flatBaseDamage := flatDamage + comboDamageBonus*float64(comboPoints)

			baseDamage := sim.Roll(flatBaseDamage, flatBaseDamage+damageVariance) +
				0.03*float64(comboPoints)*spell.MeleeAttackPower(target)

			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

			if result.Landed() {
				rogue.SpendComboPoints(sim, spell)
			} else {
				spell.IssueRefund(sim)
			}

			spell.DealDamage(sim, result)
		},
	})
	rogue.Finishers = append(rogue.Finishers, rogue.Eviscerate)
}
