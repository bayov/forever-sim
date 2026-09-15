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

	damageMultiplier := 2.5 * []float64{1, 1.05, 1.1}[rogue.Talents.Opportunity]

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
			// Cutthroat lets the Stealth requirement slide for a short while after a Backstab.
			return rogue.IsStealthed() || (rogue.CutthroatAura != nil && rogue.CutthroatAura.IsActive())
		},

		BonusCritRating:  15 * core.CritRatingPerCritChance * float64(rogue.Talents.ImprovedAmbush),
		DamageMultiplier: damageMultiplier,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)
			if rogue.CutthroatAura != nil {
				rogue.CutthroatAura.Deactivate(sim)
			}
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
