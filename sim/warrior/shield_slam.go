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
	// Forever's rank 4 deals 655 ± 4.57% (625 to 685) plus twice the block value, with no attack
	// power (client 70425, spell 23925). EV-0004: Rank 1 at block value 28 reads 481.1 against 481.1
	// for 2 x and 453.4 for 1 x, and no hit in 205 tops the 2 x ceiling at block values 40 to 48.
	// EV-0050: its threat is the damage plus the rank's own flat, added once on a crit (Ranks 1 to 3
	// 178, 203, 229), so Rank 4 takes 255 on that line until a level 60 log measures it.
	//
	// It is struck with the shield, so it is an off-hand special: no Windfury Totem or main hand
	// weapon procs (wowsims/forever b6ce30c972). With no off-hand weapon behind it, it takes none
	// of the off-hand weapon modifiers (isShieldStrike), and its attack table stays the main hand's.
	if warrior.Env.IsForever() {
		damageLow, damageHigh = 625, 685
		apCoef = 0
		threat = 255
		procMask = core.ProcMaskMeleeOHSpecial
	}

	threatBonus := core.TernaryFloat64(warrior.Env.IsForever(), threat, threat*2)

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
		FlatThreatBonus:  threatBonus,
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
