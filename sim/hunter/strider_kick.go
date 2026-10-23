package hunter

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Strider Kick, the 21 point Survival talent (beta client 1317257, build 1.60.1.70338): 100%
// normalized melee weapon damage, an 8 sec cooldown, 5.81% of base mana, and 30% run speed for 3 sec
// (its row's third effect; wowsims/forever b8c06a0e07), which only a weave walks with. Like every run
// speed bonus only the highest counts (MoveHeap). Savage Strikes and Predator's Edge cover it, and since client 1.60.1.70425
// Resourcefulness does too (its mask gained Strider Kick's bit); Efficiency doesn't.
func (hunter *Hunter) registerStriderKickSpell() {
	if !hunter.Talents.StriderKick {
		return
	}

	speedAura := hunter.RegisterAura(core.Aura{
		Label:    "Strider Kick",
		ActionID: core.ActionID{SpellID: 1317257},
		Duration: time.Second * 3,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.AddMoveSpeedModifier(&aura.ActionID, 1.30)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.RemoveMoveSpeedModifier(&aura.ActionID)
		},
	})

	hunter.StriderKick = hunter.RegisterSpell(core.SpellConfig{
		SpellCode:     SpellCode_HunterStriderKick,
		ActionID:      core.ActionID{SpellID: 1317257},
		SpellSchool:   core.SpellSchoolPhysical,
		DefenseType:   core.DefenseTypeMelee,
		ProcMask:      core.ProcMaskMeleeMHSpecial,
		Flags:         core.SpellFlagMeleeMetrics | core.SpellFlagAPL | SpellFlagStrike,
		RequiredLevel: 30,

		ManaCost: core.ManaCostOptions{
			BaseCost: 0.0581,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    hunter.NewTimer(),
				Duration: time.Second * 8,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return hunter.inMeleeRange()
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			damage := hunter.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
			spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
			speedAura.Activate(sim)
		},
	})
}
