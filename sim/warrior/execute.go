package warrior

import (
	"github.com/wowsims/classic/sim/core"
)

func (warrior *Warrior) registerExecuteSpell() {

	flatDamage := 600.0
	convertedRageDamage := 15.0
	spellID := int32(20662)

	forever := warrior.Env.IsForever()
	var rageMetrics *core.ResourceMetrics
	warrior.Execute = warrior.RegisterSpell(BattleStance|BerserkerStance, core.SpellConfig{
		SpellCode:   SpellCode_WarriorExecute,
		ActionID:    core.ActionID{SpellID: spellID},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL | SpellFlagOffensive,

		RageCost: core.RageCostOptions{
			Cost:   15 - []float64{0, 3, 5}[warrior.Talents.ImprovedExecute], // client 70338: 3 and 5
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return sim.IsExecutePhase20()
		},

		CritDamageBonus: warrior.impale(),

		DamageMultiplier: 1,
		ThreatMultiplier: 1.25,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			extraRage := spell.Unit.CurrentRage()
			baseDamage := flatDamage + convertedRageDamage*(extraRage)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

			// Under Forever the extra rage goes only when the Execute lands: a miss, dodge or parry
			// keeps it and refunds none of the base cost (36 avoided Executes in 1.60.1.70862 logs,
			// wowsims/forever 7025f28b98: 46.9 Rage to 37.9 after a parry and one Bloodrage tick, at
			// a base cost of 10).
			if result.Landed() || !forever {
				warrior.SpendRage(sim, extraRage, rageMetrics)
				// We must count this rage event if the spell itself cost 0,
				// otherwise we could end up with 0 events even though rage was spent.
				if spell.Cost.GetCurrentCost() > 0 {
					rageMetrics.Events--
				}
			}
			if !result.Landed() && !forever {
				spell.IssueRefund(sim)
			}
			spell.DealDamage(sim, result)
		},
	})
	rageMetrics = warrior.Execute.Cost.SpellCostFunctions.(*core.RageCost).ResourceMetrics
}
