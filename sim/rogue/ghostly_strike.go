package rogue

import (
	"time"

	"github.com/wowsims/classic/sim/core/stats"

	"github.com/wowsims/classic/sim/core"
)

func (rogue *Rogue) registerGhostlyStrikeSpell() {
	if !rogue.Talents.GhostlyStrike {
		return
	}

	ghostlyStrikeAura := rogue.RegisterAura(core.Aura{
		Label:    "Ghostly Strike Buff",
		ActionID: core.ActionID{SpellID: 14278},
		Duration: time.Second * 7,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			rogue.AddStatDynamic(sim, stats.Dodge, 15*core.DodgeRatingPerDodgeChance)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			rogue.AddStatDynamic(sim, stats.Dodge, -15*core.DodgeRatingPerDodgeChance)
		},
	})

	// Blizzard's November 2 development notes: Ghostly Strike "now also causes your next Ambush or
	// Garrote cast within 7 seconds to not require Stealth". No client spell for it yet (1.60.1.70753),
	// so the aura borrows Ghostly Strike's ID with a tag, and is taken to start on the cast like the
	// dodge buff.
	rogue.GhostlyStrikeOpenerAura = rogue.RegisterAura(core.Aura{
		Label:    "Ghostly Strike Opener",
		ActionID: core.ActionID{SpellID: 14278, Tag: 1},
		Duration: time.Second * 7,
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell == rogue.Ambush || spell == rogue.Garrote {
				aura.Deactivate(sim)
			}
		},
	})

	rogue.GhostlyStrike = rogue.RegisterSpell(core.SpellConfig{
		SpellCode:   SpellCode_RogueGhostlyStrike,
		ActionID:    ghostlyStrikeAura.ActionID,
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		// Cold Blood only names Sinister Strike, Backstab, Ambush, Eviscerate and Mutilate.
		Flags: rogue.builderFlags() &^ SpellFlagColdBlooded,
		EnergyCost: core.EnergyCostOptions{
			Cost:   40.0,
			Refund: 0.8,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			CD: core.Cooldown{
				Timer:    rogue.NewTimer(),
				Duration: time.Second * 20,
			},
			IgnoreHaste: true,
		},

		CritDamageBonus: rogue.lethality(),

		DamageMultiplier: core.TernaryFloat64(rogue.HasDagger(core.MainHand), 1.8, 1.25),
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)
			baseDamage := spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target)) * rogue.quietusMultiplier(sim)

			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

			ghostlyStrikeAura.Activate(sim)
			rogue.GhostlyStrikeOpenerAura.Activate(sim)

			if result.Landed() {
				rogue.AddComboPoints(sim, 1, target, spell.ComboPointMetrics())
			} else {
				spell.IssueRefund(sim)
			}
		},
	})
}
