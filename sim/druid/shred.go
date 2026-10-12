package druid

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Shred under Forever (1.60.1.70338, wago.tools): 155% weapon damage (1.12 has 225%), on
// the weapon damage plus the rank's flat bonus, which the percent scales too. The tooltip
// still says "plus 180" at rank 5, 1.12's 80 times 2.25, but the client's effects are 80
// and 155%, so the bonus comes to 124.
var ShredSpellId = []int32{0, 5221, 6800, 8992, 9829, 9830}
var ShredFlatDamage = []float64{0, 24, 32, 44, 64, 80}
var ShredLevel = []int32{0, 22, 30, 38, 46, 54}

const ShredWeaponMultiplier = 1.55

func (druid *Druid) registerShredSpell() {
	rank := core.HighestRankAt(druid.Level, ShredLevel)
	if rank == 0 {
		return
	}
	flatDamageBonus := ShredFlatDamage[rank]

	druid.Shred = druid.RegisterSpell(Cat, core.SpellConfig{
		SpellCode:   SpellCode_DruidShred,
		ActionID:    core.ActionID{SpellID: ShredSpellId[rank]},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL | SpellFlagOmen | SpellFlagBuilder,

		Rank:          rank,
		RequiredLevel: int(ShredLevel[rank]),

		EnergyCost: core.EnergyCostOptions{
			// Shredding Attacks takes 6 a point off.
			Cost:   60 - 6*float64(druid.Talents.ShreddingAttacks),
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return !druid.PseudoStats.InFrontOfTarget
		},

		DamageMultiplier: ShredWeaponMultiplier,
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
		ExpectedInitialDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
			baseDamage := flatDamageBonus + spell.Unit.AutoAttacks.MH().CalculateAverageWeaponDamage(spell.MeleeAttackPower(target))

			baseres := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeExpectedMagicAlwaysHit)

			attackTable := spell.Unit.AttackTables[target.UnitIndex][spell.CastType]
			critChance := spell.PhysicalCritChance(attackTable)
			critMod := critChance * (spell.CritMultiplier(attackTable) - 1)

			baseres.Damage *= 1 + critMod

			return baseres
		},
	})
}
