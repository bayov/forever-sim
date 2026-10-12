package druid

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Prowl and Ravage, the cat's opener (1.60.1.70338, wago.tools). Ravage needs Prowl and the
// target's back: 350% weapon damage on the weapon damage plus the rank's flat bonus (98 at
// rank 4, level 58), 60 Energy, 1 combo point. Prowl is cast before the pull; the cat's
// first swing waits for the opener, and the first damage it deals breaks Prowl.
var RavageSpellId = []int32{0, 6785, 6787, 9866, 9867}
var RavageFlatDamage = []float64{0, 42, 62, 78, 98}
var RavageLevel = []int32{0, 32, 42, 50, 58}

const RavageWeaponMultiplier = 3.5

var ProwlSpellId = []int32{0, 5215, 6783, 9913}
var ProwlLevel = []int32{0, 10, 24, 40}

func (druid *Druid) registerProwlSpell() {
	rank := core.HighestRankAt(druid.Level, ProwlLevel)
	if rank == 0 {
		return
	}
	actionID := core.ActionID{SpellID: ProwlSpellId[rank]}

	druid.ProwlAura = druid.RegisterAura(core.Aura{
		Label:    "Prowl",
		ActionID: actionID,
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			if sim.CurrentTime < 0 {
				// The opener goes before the first swing.
				druid.AutoAttacks.DelayMeleeBy(sim, time.Millisecond)
			}
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			aura.Deactivate(sim)
		},
	})

	druid.Prowl = druid.RegisterSpell(Cat, core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagAPL,
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: time.Second * 10,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			// Only out of combat, so before the pull.
			return sim.CurrentTime < 0
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			druid.ProwlAura.Activate(sim)
		},
	})
}

func (druid *Druid) registerRavageSpell() {
	rank := core.HighestRankAt(druid.Level, RavageLevel)
	if rank == 0 || druid.ProwlAura == nil {
		return
	}
	flatDamageBonus := RavageFlatDamage[rank]

	druid.Ravage = druid.RegisterSpell(Cat, core.SpellConfig{
		SpellCode:   SpellCode_DruidRavage,
		ActionID:    core.ActionID{SpellID: RavageSpellId[rank]},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL | SpellFlagOmen | SpellFlagBuilder,

		Rank:          rank,
		RequiredLevel: int(RavageLevel[rank]),

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
			return druid.ProwlAura.IsActive() && !druid.PseudoStats.InFrontOfTarget
		},

		DamageMultiplier: RavageWeaponMultiplier,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := flatDamageBonus + spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target))

			rendAndTear := druid.RendAndTearMultiplier(target)
			spell.DamageMultiplier *= rendAndTear
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
			spell.DamageMultiplier /= rendAndTear

			if result.Landed() {
				druid.AddComboPoints(sim, 1, target, spell.ComboPointMetrics())
			} else {
				spell.IssueRefund(sim)
			}
		},
	})
}
