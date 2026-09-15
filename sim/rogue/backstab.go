package rogue

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (rogue *Rogue) registerBackstabSpell() {
	if rogue.Level < 4 {
		return
	}
	// The tooltip's flat bonus divided by the 150% weapon damage multiplier.
	flatDamageBonus := rankAt(rogue.Level, map[int32]float64{
		4: 10, 12: 20, 20: 32, 28: 46, 36: 60, 44: 90, 52: 110,
		60: core.TernaryFloat64(core.IncludeAQ, 150, 140),
	})
	spellID := rankSpellID(rogue.Level, map[int32]int32{
		4: 53, 12: 2589, 20: 2590, 28: 2591, 36: 8721, 44: 11279, 52: 11280,
		60: core.TernaryInt32(core.IncludeAQ, 25300, 11281),
	})
	core.RegisterSpellRanks(11281, 25300)

	damageMultiplier := 1.5 *
		[]float64{1, 1.05, 1.1}[rogue.Talents.Opportunity] *
		[]float64{1, 1.02, 1.04, 1.06}[rogue.Talents.Aggression]

	// TODO: Only rank 1 of Puncturing Wounds was seen, the extra combo point chance is
	// assumed to scale linearly. Beta will confirm.
	extraComboPointChance := 0.15 * float64(rogue.Talents.PuncturingWounds)
	cpMetrics := rogue.NewComboPointMetrics(core.ActionID{SpellID: 13866})

	rogue.Backstab = rogue.RegisterSpell(core.SpellConfig{
		SpellCode:   SpellCode_RogueBackstab,
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
			return !rogue.PseudoStats.InFrontOfTarget
		},

		BonusCritRating: 10 * core.CritRatingPerCritChance * float64(rogue.Talents.PuncturingWounds),

		CritDamageBonus: rogue.lethality(),

		DamageMultiplier: damageMultiplier,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)
			baseDamage := (flatDamageBonus + spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target)))
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

			if result.Landed() {
				rogue.AddComboPoints(sim, 1, target, spell.ComboPointMetrics())
				if sim.Proc(extraComboPointChance, "Puncturing Wounds") {
					rogue.AddComboPoints(sim, 1, target, cpMetrics)
				}
			} else {
				spell.IssueRefund(sim)
			}
		},
	})
}
