package warrior

import (
	"github.com/wowsims/classic/sim/core"
)

func (warrior *Warrior) registerExecuteSpell() {

	flatDamage := 600.0
	convertedRageDamage := 15.0
	spellID := int32(20662)

	forever := warrior.Env.IsForever()

	// Under Forever the cast strikes on its own before the damage, and that strike rolls Windfury
	// Totem too, so a landed Execute rolls it twice. wowsims/forever's EV-0155 (their 56473177e7)
	// counts Windfury Totem hits off Execute against the same warriors' other abilities, where two
	// rolls read about 1.81: after hotfix push 114313 (build 1.60.1.71037, 2026-11-20 23:47 UTC to
	// 11-21 08:05 UTC) 1.68 [1.38, 1.98] on 39 warriors with Execute ranks 1-4 and 2.09 [1.89, 2.31]
	// on 50 with rank 5; EV-0002 (2026-11-14) read 31.6% of 846 landed Executes, where one roll
	// gives 20%. It flips with Blizzard's pushes: their reports of 2026-11-20 before the push read
	// one roll (1.01 and 1.05 on 112 warriors), after Blizzard's Nov 16 post had called the second
	// roll a bug ("Mutilate and Execute would have an extra attempt to activate Windfury Totem",
	// fixed). Push 114313 sets Suppress Weapon Procs on the cast rows, so the strike rolls the
	// totem and no weapon proc; client 71037 also makes those rows Always Hit, so the strike rolls
	// it whatever the damage does (their avoided Executes drew a Windfury hit 28 of 171 times after
	// the push, against none of 188 before it). DPS Report assumption warr_exec_wf2 (0: one roll).
	castRollsTotem := forever && core.Assume("warr_exec_wf2", 1) != 0 // DPS Report assumption knob

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
			if castRollsTotem {
				warrior.RollWindfuryTotemOnCast(sim, spell)
			}

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
