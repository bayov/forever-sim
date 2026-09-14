package rogue

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Improved Expose Armor no longer scales the armor reduction, it discounts the finisher
// and hands a combo point back on a full spend.
// TODO: assumed baseline, beta will confirm. The raid reads this debuff, so the Classic
// 2/2 armor value is treated as baseline rather than deleted.
const exposeArmorBaselineRank = 2

func (rogue *Rogue) registerExposeArmorSpell() {
	rogue.ExposeArmorAuras = rogue.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.ExposeArmorAura(target, exposeArmorBaselineRank)
	})

	spellID := map[int32]int32{
		25: 8647,
		40: 8650,
		50: 11197,
		60: 11198,
	}[rogue.Level]

	arpenPerCombo := map[int32]float64{
		25: 80,
		40: 210,
		50: 275,
		60: 340,
	}[rogue.Level]

	arpenPerCombo *= []float64{1, 1.25, 1.5}[exposeArmorBaselineRank]

	// TODO: Only rank 1 was seen, the Energy discount is assumed to scale linearly while the
	// refund and the 5 combo point trigger stay put. Beta will confirm.
	energyCost := 25.0 - 5*float64(rogue.Talents.ImprovedExposeArmor)
	cpMetrics := rogue.NewComboPointMetrics(core.ActionID{SpellID: 14169})

	// share ExtraCastCondition() state with ApplyEffects()
	var arpen float64
	var eaAura *core.Aura

	rogue.ExposeArmor = rogue.RegisterSpell(core.SpellConfig{
		SpellCode:    SpellCode_RogueExposeArmor,
		ActionID:     core.ActionID{SpellID: spellID},
		SpellSchool:  core.SpellSchoolPhysical,
		DefenseType:  core.DefenseTypeMelee,
		ProcMask:     core.ProcMaskMeleeMHSpecial,
		Flags:        rogue.finisherFlags(),
		MetricSplits: 6,

		EnergyCost: core.EnergyCostOptions{
			Cost:   energyCost,
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
			if rogue.ComboPoints() == 0 {
				return false
			}

			eaAura = rogue.ExposeArmorAuras.Get(target)
			arpen = float64(rogue.ComboPoints()) * arpenPerCombo

			if curActive := eaAura.ExclusiveEffects[0].Category.GetActiveEffect(); curActive != nil {
				return arpen >= curActive.Priority
			}
			return true
		},

		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)

			comboPoints := rogue.ComboPoints()

			result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHit)
			if result.Landed() {
				eaAura.ExclusiveEffects[0].Priority = arpen
				eaAura.Activate(sim)
				rogue.SpendComboPoints(sim, spell)
				if rogue.Talents.ImprovedExposeArmor > 0 && comboPoints == 5 {
					rogue.AddComboPoints(sim, 1, target, cpMetrics)
				}
			} else {
				spell.IssueRefund(sim)
			}
			spell.DealOutcome(sim, result)
		},

		RelatedAuras: []core.AuraArray{rogue.ExposeArmorAuras},
	})
	rogue.Finishers = append(rogue.Finishers, rogue.ExposeArmor)
}
