package rogue

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (rogue *Rogue) registerAmbushSpell() {
	if rogue.Level < 18 {
		return
	}
	// The tooltip's flat bonus divided by the 250% weapon damage multiplier.
	flatDamageBonus := rankAt(rogue.Level, map[int32]float64{
		18: 28, 26: 40, 34: 50, 42: 74, 50: 92, 58: 116,
	})
	spellID := rankSpellID(rogue.Level, map[int32]int32{
		18: 8676, 26: 8724, 34: 8725, 42: 11267, 50: 11268, 58: 11269,
	})

	damageMultiplier := 2.5 * rogue.opportunityMultiplier()

	rogue.Ambush = rogue.RegisterSpell(core.SpellConfig{
		SpellCode:   SpellCode_RogueAmbush,
		ActionID:    core.ActionID{SpellID: spellID},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       rogue.builderFlags(),

		EnergyCost: core.EnergyCostOptions{
			Cost:   60,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			if !rogue.HasDagger(core.MainHand) {
				return false
			}
			// Under Forever Ambush needs the rogue behind the target (wowsims/forever ambush.go).
			if rogue.Env.IsForever() && rogue.PseudoStats.InFrontOfTarget {
				return false
			}
			// Ghostly Strike lets the Stealth requirement slide for 7 sec.
			return rogue.IsStealthed() || rogue.canOpenWithoutStealth()
		},

		BonusCritRating:  15 * core.CritRatingPerCritChance * float64(rogue.Talents.ImprovedAmbush),
		DamageMultiplier: damageMultiplier,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)
			baseDamage := (flatDamageBonus + spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target)))

			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialNoBlockDodgeParry)

			if result.Landed() {
				rogue.AddComboPoints(sim, 1, target, spell.ComboPointMetrics())
			} else {
				spell.IssueRefund(sim)
			}
		},
	})
}
