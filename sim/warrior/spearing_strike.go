package warrior

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// Spearing Strike (spell 1310222): 15 Rage, 20 sec cooldown, 75% of normalized weapon damage, and
// 75% more on top against Giants and Dragonkin. Hotfix push 113539 (2026-11-03): SpellEffect 1339573
// (effect 1, weapon %) 40 -> 75 and 1339768 (effect 2, the creature-type bonus's multiplier) 2 -> 1,
// so the bonus goes from 80% to 75% (client 70755 still 40 / 80). ForeverChanges says it needs
// Battle Stance since the 1 October build.
func (warrior *Warrior) registerSpearingStrikeSpell() {
	if !warrior.Talents.SpearingStrike {
		return
	}

	warrior.SpearingStrike = warrior.RegisterSpell(BattleStance, core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 1310222},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL | SpellFlagOffensive,

		RageCost: core.RageCostOptions{
			Cost:   15,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    warrior.NewTimer(),
				Duration: time.Second * 20,
			},
		},

		CritDamageBonus: warrior.impale(),

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			weaponPercent := 0.75
			if target.MobType == proto.MobType_MobTypeGiant || target.MobType == proto.MobType_MobTypeDragonkin {
				weaponPercent += 0.75
			}
			baseDamage := weaponPercent * spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
			if !result.Landed() {
				spell.IssueRefund(sim)
			}
		},
	})
}
