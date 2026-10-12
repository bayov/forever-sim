package druid

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Claw under Forever (1.60.1.70338, wago.tools): 110% weapon damage on the weapon damage
// plus the rank's flat bonus (1.12 has 100%), so rank 5's 115 comes to 126.5.
var ClawSpellId = []int32{0, 1082, 3029, 5201, 9849, 9850}
var ClawFlatDamage = []float64{0, 27, 39, 57, 88, 115}
var ClawLevel = []int32{0, 20, 28, 38, 48, 58}

const ClawWeaponMultiplier = 1.1

func (druid *Druid) registerClawSpell() {
	rank := core.HighestRankAt(druid.Level, ClawLevel)
	if rank == 0 {
		return
	}
	flatDamageBonus := ClawFlatDamage[rank]

	druid.Claw = druid.RegisterSpell(Cat, core.SpellConfig{
		SpellCode:   SpellCode_DruidClaw,
		ActionID:    core.ActionID{SpellID: ClawSpellId[rank]},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL | SpellFlagOmen | SpellFlagBuilder,

		Rank:          rank,
		RequiredLevel: int(ClawLevel[rank]),

		EnergyCost: core.EnergyCostOptions{
			// Ferocity takes 1 a point off.
			Cost:   45 - float64(druid.Talents.Ferocity),
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
		},

		DamageMultiplier: ClawWeaponMultiplier,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := flatDamageBonus + spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target))

			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

			if result.Landed() {
				druid.AddComboPoints(sim, 1, target, spell.ComboPointMetrics())
			} else {
				spell.IssueRefund(sim)
			}
		},
	})
}
