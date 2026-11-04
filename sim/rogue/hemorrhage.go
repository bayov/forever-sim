package rogue

import (
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Hemorrhage no longer weakens the target for the whole raid, it makes the rogue's own
// Rupture and Garrote hit harder. Garrote came with Blizzard's November 2 development notes;
// client 16511 in 1.60.1.70753 still masks Rupture only (1048576).
const HemorrhageRuptureMultiplier = 1.15

func (rogue *Rogue) registerHemorrhageSpell() {
	if !rogue.Talents.Hemorrhage {
		return
	}

	// The client's one Forever rank (16511).
	spellID := int32(16511)

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
		// Cold Blood only names Sinister Strike, Backstab, Ambush, Eviscerate and Mutilate.
		Flags: rogue.builderFlags() &^ SpellFlagColdBlooded,

		// Efficient Strikes covers Hemorrhage too (November 2 development notes).
		EnergyCost: core.EnergyCostOptions{
			Cost:   35 - rogue.hemorrhageEfficientStrikes(),
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
			// Normalized weapon damage, the client's Hemorrhage (16511) uses effect 121 like Sinister Strike.
			baseDamage := spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))

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

// The Rupture and Garrote damage multiplier from the rogue's Hemorrhage on the target.
func (rogue *Rogue) hemorrhageMultiplier(target *core.Unit) float64 {
	return core.TernaryFloat64(rogue.isHemorrhaging(target), HemorrhageRuptureMultiplier, 1)
}

// Garrote's share of Hemorrhage: Forever only, from the November 2 development notes (standards 37).
func (rogue *Rogue) garroteHemorrhageMultiplier(target *core.Unit) float64 {
	if !rogue.Env.IsForever() {
		return 1
	}
	return rogue.hemorrhageMultiplier(target)
}
