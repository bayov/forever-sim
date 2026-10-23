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
	flatDamage := rankAt(rogue.Level, map[int32]float64{1: 17.25, 40: 24.75, 50: 36, 60: 50.25})

	rogue.mutilateOH = rogue.RegisterSpell(core.SpellConfig{
		ActionID:    MutilateActionID.WithTag(2),
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeOHSpecial,
		Flags:       SpellFlagBuilder | SpellFlagColdBlooded | core.SpellFlagMeleeMetrics | core.SpellFlagNoOnCastComplete,

		// Puncturing Wounds raises the crit chance of the whole Mutilate, both halves.
		BonusCritRating: 5 * core.CritRatingPerCritChance * float64(rogue.Talents.PuncturingWounds),

		CritDamageBonus: rogue.lethality(),

		DamageMultiplier: rogue.AutoAttacks.OHConfig().DamageMultiplier * []float64{1, 1.05, 1.1}[rogue.Talents.Opportunity],
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := rogue.mutilateDamage(target, flatDamage, rogue.OHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target)))
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

		// 60 Energy in the client (1310707).
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
			// Both hands need a dagger: the client (SpellEquippedItems, 1.60.1.70338) asks for one
			// on the cast and on both halves, every rank (1310707, 1310706, 1310705, 399961, 1241588,
			// 1241590), and the class deep dive says "both must be daggers".
			return rogue.HasDagger(core.MainHand) && rogue.HasDagger(core.OffHand)
		},

		BonusCritRating: 5 * core.CritRatingPerCritChance * float64(rogue.Talents.PuncturingWounds),

		CritDamageBonus: rogue.lethality(),

		DamageMultiplier: []float64{1, 1.05, 1.1}[rogue.Talents.Opportunity],
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)

			// Cold Blood covers both halves and is spent by the off hand half (talents.go).
			baseDamage := rogue.mutilateDamage(target, flatDamage, rogue.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target)))
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
// the rogue's lingering poisons is on the target. The flat bonus is the Forever rank the
// level knows. The client adds 23, 33, 48 and 67 to the weapon damage before the 75%
// (1310706, 399960, 1241585, 1241586 in 1.60.1.70338), so 17.25 from the talent, then
// 24.75, 36 and 50.25 from the trainer at 40, 50 and 60.
func (rogue *Rogue) mutilateDamage(target *core.Unit, flatDamage float64, weaponDamage float64) float64 {
	baseDamage := flatDamage + 0.75*weaponDamage
	if rogue.isPoisoned(target) {
		baseDamage *= 1.2
	}
	return baseDamage
}

// Instant Poison leaves nothing behind, so only the two lingering poisons count.
func (rogue *Rogue) isPoisoned(target *core.Unit) bool {
	return rogue.deadlyPoisonTick.Dot(target).IsActive() || rogue.woundPoisonDebuffAuras.Get(target).IsActive()
}
