package rogue

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

var MutilateActionID = core.ActionID{SpellID: 1329}

func (rogue *Rogue) registerMutilateSpell() {
	if !rogue.Talents.Mutilate {
		return
	}

	rogue.mutilateOH = rogue.RegisterSpell(core.SpellConfig{
		ActionID:    MutilateActionID.WithTag(2),
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeOHSpecial,
		Flags:       SpellFlagBuilder | core.SpellFlagMeleeMetrics | core.SpellFlagNoOnCastComplete,

		CritDamageBonus: rogue.lethality(),

		DamageMultiplier: rogue.AutoAttacks.OHConfig().DamageMultiplier * []float64{1, 1.05, 1.1}[rogue.Talents.Opportunity],
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := rogue.mutilateDamage(target, rogue.OHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target)))
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		},
	})

	rogue.Mutilate = rogue.RegisterSpell(core.SpellConfig{
		SpellCode:   SpellCode_RogueMutilate,
		ActionID:    MutilateActionID,
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       rogue.builderFlags(),

		// TODO: The tooltip showed no Energy cost, the 60 is taken from the Classic Mutilate.
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
			// TODO: The tooltip didn't repeat the Classic dagger requirement, it's assumed to still apply.
			return rogue.HasDagger(core.MainHand) && rogue.HasDagger(core.OffHand)
		},

		BonusCritRating: 5 * core.CritRatingPerCritChance * float64(rogue.Talents.PuncturingWounds),

		CritDamageBonus: rogue.lethality(),

		DamageMultiplier: []float64{1, 1.05, 1.1}[rogue.Talents.Opportunity],
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)

			// Cold Blood is spent on the main hand half, which is the larger of the two.
			baseDamage := rogue.mutilateDamage(target, rogue.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target)))
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
			rogue.mutilateOH.Cast(sim, target)

			if result.Landed() {
				rogue.AddComboPoints(sim, 2, target, spell.ComboPointMetrics())
			} else {
				spell.IssueRefund(sim)
			}
		},
	})
}

// Each half strikes for 75% weapon damage plus a flat bonus, and hits harder while one of
// the rogue's lingering poisons is on the target.
func (rogue *Rogue) mutilateDamage(target *core.Unit, weaponDamage float64) float64 {
	baseDamage := 13 + 0.75*weaponDamage
	if rogue.isPoisoned(target) {
		baseDamage *= 1.2
	}
	return baseDamage
}

// Instant Poison leaves nothing behind, so only the two lingering poisons count.
func (rogue *Rogue) isPoisoned(target *core.Unit) bool {
	return rogue.deadlyPoisonTick.Dot(target).IsActive() || rogue.woundPoisonDebuffAuras.Get(target).IsActive()
}
