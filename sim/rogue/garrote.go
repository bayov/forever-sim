package rogue

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (rogue *Rogue) registerGarrote() {
	if rogue.Level < 14 {
		return
	}
	// Damage per tick, the tooltip total over 6 ticks.
	baseDamage := rankAt(rogue.Level, map[int32]float64{
		14: 24, 22: 34, 30: 47, 38: 59, 46: 74, 54: 92,
	})
	spellID := rankSpellID(rogue.Level, map[int32]int32{
		14: 703, 22: 8631, 30: 8632, 38: 8633, 46: 11289, 54: 11290,
	})

	rogue.Garrote = rogue.GetOrRegisterSpell(core.SpellConfig{
		SpellCode:   SpellCode_RogueGarrote,
		ActionID:    core.ActionID{SpellID: spellID},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       SpellFlagBuilder | core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		EnergyCost: core.EnergyCostOptions{
			Cost:   50.0 - 10*float64(rogue.Talents.DirtyDeeds),
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			if !rogue.IsStealthed() {
				return false
			}
			// Dirty Deeds drops the positional requirement.
			return rogue.Talents.DirtyDeeds > 0 || !rogue.PseudoStats.InFrontOfTarget
		},

		DamageMultiplier: 1 +
			0.05*float64(rogue.Talents.Opportunity),
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Garrote",
			},
			NumberOfTicks: 6,
			TickLength:    time.Second * 3,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				damage := baseDamage + dot.Spell.MeleeAttackPower(target)*0.03
				dot.Snapshot(target, damage, isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)
			result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialNoBlockDodgeParryNoCritNoHitCounter)
			if result.Landed() {
				rogue.AddComboPoints(sim, 1, target, spell.ComboPointMetrics())
				spell.Dot(target).Apply(sim)
			} else {
				spell.IssueRefund(sim)
			}
			spell.DealOutcome(sim, result)
		},
	})
}
