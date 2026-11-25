package common

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Enchants that matter to a tank or a character alone: what heals us, and what hurts whoever hits us.

// RecoveryEffectID is Forever's Enchant Weapon - Recovery, by the game's own enchant ID.
const RecoveryEffectID = 8721

// recoveryFromOurAttacks is which dodges and parries give Recovery: those of our own attacks, or with
// false those of the attacks at us.
//
// The client's formula (item 249480) says "when your attacks are dodged or parried", and the buff's
// proc mask is the attacks we deal. Our logs agree: 377 of 379 heals came with a dodge or parry of the
// wearer's own swing or ability, and 371 of 372 of those healed when the cooldown was up. Of some 840
// attacks the wearers dodged or parried, 13 had a heal at the same moment.
const recoveryFromOurAttacks = true

func init() {
	core.AddEffectsToTest = false
	defer func() { core.AddEffectsToTest = true }()

	// Weapon - Recovery: heals 5% of max health, at most once in 10 sec (client 71091: 1248761 and its
	// heal 1248759). Any of our melee attacks, with either hand. It can crit, 5 of 379 times in our logs.
	core.NewEnchantEffect(RecoveryEffectID, func(agent core.Agent) {
		character := agent.GetCharacter()
		const label = "Recovery"
		if character.HasAura(label) {
			return // On the other weapon too. They share the buff.
		}

		heal := character.RegisterSpell(core.SpellConfig{
			ActionID:         core.ActionID{SpellID: 1248759},
			SpellSchool:      core.SpellSchoolPhysical,
			DefenseType:      core.DefenseTypeMagic,
			ProcMask:         core.ProcMaskEmpty,
			Flags:            core.SpellFlagHelpful | core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,
			DamageMultiplier: 1,
			ThreatMultiplier: 1,
			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealHealing(sim, target, 0.05*character.MaxHealth(), spell.OutcomeHealingCrit)
			},
		})
		icd := core.Cooldown{Timer: character.NewTimer(), Duration: time.Second * 10}
		onAttack := func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.Outcome.Matches(core.OutcomeDodge|core.OutcomeParry) && spell.ProcMask.Matches(core.ProcMaskMelee) &&
				character.HasHealthBar() && icd.IsReady(sim) &&
				character.GetProcMaskForEnchant(RecoveryEffectID) != core.ProcMaskUnknown {
				icd.Use(sim)
				heal.Cast(sim, &character.Unit)
			}
		}
		config := core.Aura{
			Label:    label,
			ActionID: core.ActionID{SpellID: 1248761},
			Duration: core.NeverExpires,
			OnReset: func(aura *core.Aura, sim *core.Simulation) {
				aura.Activate(sim)
			},
		}
		if recoveryFromOurAttacks {
			config.OnSpellHitDealt = onAttack
		} else {
			config.OnSpellHitTaken = onAttack
		}
		aura := character.RegisterAura(config)
		character.ItemSwap.RegisterOnSwapItemForEffect(RecoveryEffectID, aura)
	})

	// Shield spikes: each melee attack we block, in full or in part, hurts the attacker (client 71091,
	// aura 43). In our logs (1,079 hits) every block does, however soon after the last: the 5 sec
	// cooldown of the buffs' client rows isn't in the game. It ignores armor and never misses or crits.
	registerShieldSpike(43, 9784, "Iron Shield Spike", 8, 12)
	registerShieldSpike(463, 9782, "Mithril Shield Spike", 16, 20)
	registerShieldSpike(1704, 16624, "Thorium Shield Spike", 20, 30)
}

func registerShieldSpike(effectID, spellID int32, label string, low, high float64) {
	core.NewEnchantEffect(effectID, func(agent core.Agent) {
		character := agent.GetCharacter()
		if character.HasAura(label) {
			return
		}

		spike := character.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: spellID},
			SpellSchool: core.SpellSchoolPhysical,
			DefenseType: core.DefenseTypeMelee,
			ProcMask:    core.ProcMaskEmpty,
			Flags: core.SpellFlagIgnoreResists | core.SpellFlagMeleeMetrics | core.SpellFlagNoOnCastComplete |
				core.SpellFlagPassiveSpell,
			DamageMultiplier: 1,
			ThreatMultiplier: 1,
			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealDamage(sim, target, sim.Roll(low, high), spell.OutcomeAlwaysHit)
			},
		})
		aura := character.RegisterAura(core.Aura{
			Label:    label,
			ActionID: core.ActionID{SpellID: spellID},
			Duration: core.NeverExpires,
			OnReset: func(aura *core.Aura, sim *core.Simulation) {
				aura.Activate(sim)
			},
			OnSpellHitTaken: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if result.DidBlock() && spell.ProcMask.Matches(core.ProcMaskMelee) && spell.Unit.Type == core.EnemyUnit &&
					character.GetProcMaskForEnchant(effectID) != core.ProcMaskUnknown {
					spike.Cast(sim, spell.Unit)
				}
			},
		})
		character.ItemSwap.RegisterOnSwapItemForEffect(effectID, aura)
	})
}
