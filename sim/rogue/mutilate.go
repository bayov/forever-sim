package rogue

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Mutilate's casts: 399956 at 40, 1241582 at 50, 1241584 at 60 (server hotfix push 113483 deleted
// the old level 30 rank, 1310707 and its halves 1310705/6); the off hand half is tag 2.
var MutilateActionID = core.ActionID{SpellID: 1241584}

func (rogue *Rogue) registerMutilateSpell() {
	if !rogue.Talents.Mutilate {
		return
	}
	flatDamage := rankAt(rogue.Level, map[int32]float64{1: 44, 50: 63, 60: 88})
	actionID := core.ActionID{SpellID: rankSpellID(rogue.Level, map[int32]int32{1: 399956, 50: 1241582, 60: 1241584})}

	rogue.mutilateOH = rogue.RegisterSpell(core.SpellConfig{
		ActionID:    actionID.WithTag(2),
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeOHSpecial,
		Flags:       SpellFlagBuilder | SpellFlagColdBlooded | core.SpellFlagMeleeMetrics | core.SpellFlagNoOnCastComplete,

		// Puncturing Wounds raises the crit chance of the whole Mutilate, both halves.
		BonusCritRating: 5 * core.CritRatingPerCritChance * float64(rogue.Talents.PuncturingWounds),

		CritDamageBonus: rogue.lethality(),

		DamageMultiplier: rogue.AutoAttacks.OHConfig().DamageMultiplier * rogue.opportunityMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := rogue.mutilateDamage(target, flatDamage, rogue.OHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target)))
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		},
	})

	rogue.Mutilate = rogue.RegisterSpell(core.SpellConfig{
		SpellCode:   SpellCode_RogueMutilate,
		ActionID:    actionID,
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

		DamageMultiplier: rogue.opportunityMultiplier(),
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

// Each half strikes for 90% weapon damage plus a flat bonus, and hits harder while one of the
// rogue's lingering poisons is on the target. Server hotfix push 113483 (launch realms, build 70707,
// 2026-11-03; Blizzard's November 2 notes announced it) sets the halves (399960/1, 1241585/8,
// 1241586/90) to 90% weapon damage (was 75) and a flat 44, 63 and 88 at 40, 50 and 60 (was 33, 48,
// 67), added before the percent: 88 x 0.9 = 79.2 a half at 60. It also deletes the level 30 rank
// (flat 23), so the talent teaches the level 40 rank. Client 1.60.1.70753 still has 75, which is kept
// outside Forever (Classic has no Mutilate).
func (rogue *Rogue) mutilateDamage(target *core.Unit, flatDamage float64, weaponDamage float64) float64 {
	baseDamage := core.TernaryFloat64(rogue.Env.IsForever(), 0.9, 0.75) * (flatDamage + weaponDamage)
	if rogue.isPoisoned(target) {
		baseDamage *= 1.2
	}
	return baseDamage
}

// Instant Poison leaves nothing behind, so only the two lingering poisons count.
func (rogue *Rogue) isPoisoned(target *core.Unit) bool {
	return rogue.deadlyPoisonTick.Dot(target).IsActive() || rogue.woundPoisonDebuffAuras.Get(target).IsActive()
}
