package rogue

import (
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Hemorrhage no longer weakens the target for the whole raid, it makes the rogue's own
// Rupture hit harder.
const HemorrhageRuptureMultiplier = 1.15

func (rogue *Rogue) registerHemorrhageSpell() {
	if !rogue.Talents.Hemorrhage {
		return
	}

	spellID := int32(17348)

	actionID := core.ActionID{SpellID: spellID}

	rogue.HemorrhageAuras = rogue.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.GetOrRegisterAura(core.Aura{
			Label:    "Hemorrhage-" + strconv.Itoa(int(rogue.Index)),
			ActionID: actionID,
			Duration: time.Second * 15,
		})
	})

	rogue.Hemorrhage = rogue.RegisterSpell(core.SpellConfig{
		SpellCode:   SpellCode_RogueHemorrhage,
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       rogue.builderFlags(),

		EnergyCost: core.EnergyCostOptions{
			Cost:   35.0,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
		},

		CritDamageBonus: rogue.lethality(),

		DamageMultiplier: core.TernaryFloat64(rogue.HasDagger(core.MainHand), 1.45, 1),
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)
			baseDamage := spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target)) * rogue.quietusMultiplier(sim)

			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

			if result.Landed() {
				rogue.AddComboPoints(sim, 1, target, spell.ComboPointMetrics())
				rogue.HemorrhageAuras.Get(target).Activate(sim)
			} else {
				spell.IssueRefund(sim)
			}
		},
	})
}

// Hemorrhage is optional, so the debuff has to be looked up defensively.
func (rogue *Rogue) isHemorrhaging(target *core.Unit) bool {
	return rogue.HemorrhageAuras != nil && rogue.HemorrhageAuras.Get(target).IsActive()
}
