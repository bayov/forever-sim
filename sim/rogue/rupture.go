package rogue

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (rogue *Rogue) registerRupture() {
	if rogue.Level < 20 {
		return
	}
	spellID := rankSpellID(rogue.Level, map[int32]int32{
		20: 1943, 28: 8639, 36: 8640, 44: 11273, 52: 11274, 60: 11275,
	})

	rogue.Rupture = rogue.RegisterSpell(core.SpellConfig{
		SpellCode:    SpellCode_RogueRupture,
		ActionID:     core.ActionID{SpellID: spellID},
		SpellSchool:  core.SpellSchoolPhysical,
		DefenseType:  core.DefenseTypeMelee,
		ProcMask:     core.ProcMaskMeleeMHSpecial,
		Flags:        rogue.finisherFlags(),
		MetricSplits: 6,

		EnergyCost: core.EnergyCostOptions{
			Cost:   25,
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

		DamageMultiplier: []float64{1, 1.1, 1.2, 1.3}[rogue.Talents.SerratedBlades],
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Rupture",
			},
			NumberOfTicks: 0, // Set dynamically
			TickLength:    time.Second * 2,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				damage := rogue.RuptureDamage(target, rogue.ComboPoints())
				if rogue.isHemorrhaging(target) {
					damage *= HemorrhageRuptureMultiplier
				}
				dot.Snapshot(target, damage, isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)
			result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHitNoHitCounter)
			if result.Landed() {
				dot := spell.Dot(target)
				dot.Spell = spell
				dot.NumberOfTicks = rogue.RuptureTicks(rogue.ComboPoints())
				dot.Apply(sim)
				rogue.SpendComboPoints(sim, spell)
			} else {
				spell.IssueRefund(sim)
			}
			spell.DealOutcome(sim, result)
		},
	})
	rogue.Finishers = append(rogue.Finishers, rogue.Rupture)
}

// The Forever tooltip totals per combo point, indexed by points. Forever cut Rupture to
// about 60% of Classic (469 at 5 points for rank 6, Classic 800). The totals no longer
// split into a clean base plus per point tick, so we keep the totals and divide by the
// tick count. The attack power part is the Classic one, the tooltip does not show it.
var ruptureTotalDamage = map[int32][6]float64{
	20: {0, 25, 37, 51, 68, 87},
	28: {0, 35, 53, 74, 99, 127},
	36: {0, 53, 79, 109, 143, 183},
	44: {0, 76, 110, 149, 195, 246},
	52: {0, 105, 151, 207, 270, 342},
	60: {0, 159, 222, 295, 377, 469},
}

func (rogue *Rogue) RuptureDamage(target *core.Unit, comboPoints int32) float64 {
	baseTickDamage := rankAt(rogue.Level, ruptureTotalDamage)[comboPoints] / float64(rogue.RuptureTicks(comboPoints))

	return baseTickDamage +
		[]float64{0, 0.04 / 4, 0.10 / 5, 0.18 / 6, 0.21 / 7, 0.24 / 8}[comboPoints]*rogue.Rupture.MeleeAttackPower(target)
}

func (rogue *Rogue) RuptureTicks(comboPoints int32) int32 {
	return 3 + comboPoints
}

func (rogue *Rogue) RuptureDuration(comboPoints int32) time.Duration {
	return time.Duration(rogue.RuptureTicks(comboPoints)) * time.Second * 2
}
