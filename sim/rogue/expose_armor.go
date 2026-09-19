package rogue

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Improved Expose Armor no longer scales the armor reduction, it discounts the finisher
// and hands a combo point back on a full spend. The Forever ranks carry a bigger armor
// value instead (450 per point at rank 5, Classic 340 or 425 with the old 2/2 talent).
// The raid debuff option reads the same value through this rank.
const exposeArmorBaselineRank = 2

func (rogue *Rogue) registerExposeArmorSpell() {
	rogue.ExposeArmorAuras = rogue.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.ExposeArmorAura(target, exposeArmorBaselineRank)
	})

	if rogue.Level < 14 {
		return
	}
	spellID := rankSpellID(rogue.Level, map[int32]int32{
		14: 8647, 26: 8649, 36: 8650, 46: 11197, 56: 11198,
	})
	arpenPerCombo := rankAt(rogue.Level, map[int32]float64{
		14: 90, 26: 180, 36: 270, 46: 360, 56: 450,
	})

	// Improved Expose Armor takes 5 Energy off per rank and refunds a combo point per rank
	// when cast at 5 points.
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
					rogue.AddComboPoints(sim, rogue.Talents.ImprovedExposeArmor, target, cpMetrics)
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
