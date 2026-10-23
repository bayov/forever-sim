package warrior

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (warrior *Warrior) registerShieldSlamSpell() {
	if !warrior.Talents.ShieldSlam {
		return
	}

	spellID := int32(23925)
	damageLow := 421.0
	damageHigh := 439.0
	threat := 254.0

	apCoef := 0.15
	blockValueCoef := 2.0
	procMask := core.ProcMaskMeleeMHSpecial
	// Forever's rank 4 deals 655 ± 4.57% (625 to 685) plus the block value once, with no attack
	// power (client 70425, spell 23925; wowsims/forever's Shield Slam).
	//
	// It is struck with the shield, so it is an off-hand special: no Windfury Totem or main hand
	// weapon procs (wowsims/forever b6ce30c972). With no off-hand weapon behind it, it takes none
	// of the off-hand weapon modifiers (isShieldStrike), and its attack table stays the main hand's.
	if warrior.Env.IsForever() {
		damageLow, damageHigh = 625, 685
		apCoef = 0
		blockValueCoef = 1
		procMask = core.ProcMaskMeleeOHSpecial
	}

	warrior.ShieldSlam = warrior.RegisterSpell(AnyStance, core.SpellConfig{
		SpellCode:   SpellCode_WarriorShieldSlam,
		ActionID:    core.ActionID{SpellID: spellID},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    procMask,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL | SpellFlagOffensive,

		RageCost: core.RageCostOptions{
			Cost:   20,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    warrior.NewTimer(),
				Duration: time.Second * 6,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return warrior.PseudoStats.CanBlock
		},

		CritDamageBonus: warrior.impale(),

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		FlatThreatBonus:  threat * 2,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			damage := sim.Roll(damageLow, damageHigh) + warrior.BlockValue()*blockValueCoef + apCoef*spell.MeleeAttackPower(target)
			result := spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMeleeSpecialHitAndCrit)

			if !result.Landed() {
				spell.IssueRefund(sim)
			}
		},
	})
}
