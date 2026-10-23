package druid

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Ravage under Forever (client rows 6785, 6787, 9866, 9867, as wowsims/forever's ravage.go reads
// them): 350% weapon damage on the weapon damage plus the rank's flat bonus, which the percent
// scales too, so rank 4's 98 comes to 343. 60 Energy, 80% back on a miss (Discount Power On Miss)
// and 1 combo point. It needs stealth (Attributes[0] 0x20000), Prowl or a Night Elf's Shadowmeld,
// and the target's back, and has no active defense (Attributes[0] 0x200000): it misses or lands,
// never dodged, parried or blocked. Prowl is never up in a fight (prowl.go), so Shadowmeld in
// combat is the only way in.
//
// The client's class masks give it Clearcasting, Berserk and Natural Instinct, not Savage Fury or
// Shredding Attacks. Rend and Tear is a dummy on the melee abilities, so it gets that too, and
// Blood Frenzy, as a combo point builder.
var RavageSpellId = []int32{0, 6785, 6787, 9866, 9867}
var RavageFlatDamage = []float64{0, 42, 62, 78, 98}
var RavageLevel = []int32{0, 32, 42, 50, 58}

const RavageWeaponMultiplier = 3.5

func (druid *Druid) registerRavageSpell() {
	rank := core.HighestRankAt(druid.Level, RavageLevel)
	if rank == 0 {
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
			return druid.Stealthed() && !druid.PseudoStats.InFrontOfTarget
		},

		DamageMultiplier: RavageWeaponMultiplier,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			druid.BreakStealth(sim)
			baseDamage := flatDamageBonus + spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target))

			rendAndTear := druid.RendAndTearMultiplier(target)
			spell.DamageMultiplier *= rendAndTear
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialNoBlockDodgeParry)
			spell.DamageMultiplier /= rendAndTear

			if result.Landed() {
				druid.AddComboPoints(sim, 1, target, spell.ComboPointMetrics())
			} else {
				spell.IssueRefund(sim)
			}
		},
	})
}
