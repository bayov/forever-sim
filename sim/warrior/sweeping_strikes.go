package warrior

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (warrior *Warrior) registerSweepingStrikesCD() {
	if !warrior.Talents.SweepingStrikes {
		return
	}

	// Procs from auto attacks and most abilities https://www.wowhead.com/classic/spell=12723/sweeping-strikes
	//
	// Under Forever the copy is the first hit's damage, its crit included, with the first target's
	// armor taken out and the second target's put in, and the attacker's damage bonuses apply to it
	// a second time, except one that needs a weapon: Two-Handed Weapon Specialization counts once
	// (wowsims/forever EV-0182, 20,650 copies from 148 warriors: Enrage's +10% moves the copy's
	// ratio to its hit by +9.85%, and a two-hander's copy equals its hit on equal armor). The
	// target's damage taken modifiers went into the hit and are not taken again. From 2026-10-23 to
	// 2026-11-25 the copy skipped the attacker's bonuses too, on their earlier reading of the row
	// (replaced by the user's ruling of 2026-11-25). Classic runs the copy through every damage
	// multiplier a second time.
	var curDmg float64
	copyFlags := core.SpellFlagMeleeMetrics | core.SpellFlagNoOnCastComplete
	if warrior.Env.IsForever() {
		copyFlags |= core.SpellFlagIgnoreTargetModifiers
	}
	// 12723 states no DefenseType, so no chance-on-hit weapon or combat enchant hears it
	// (EV-0007: 0 procs on 567 copies), and it keeps no proc mask.
	hitSchoolDamagWithValue := warrior.RegisterSpell(AnyStance, core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 12723},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskEmpty, // No proc mask, so it won't proc itself.
		Flags:       copyFlags,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			damage := curDmg
			if sim.IsForever() && warrior.twoHandSpecPhysical > 0 {
				damage /= warrior.twoHandSpecPhysical
			}
			spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeAlwaysHit)
		},
	})

	// Procs from WW, also Execute? https://www.wowhead.com/classic/spell=26654/sweeping-strikes
	//
	// Under Forever the normalized hit has the melee DefenseType and is a main-hand special, so the
	// main hand's chance-on-hit weapons roll on it (R-0042, EV-0007: 0.83 of the white hit rate).
	// Classic gives it no proc mask.
	hitSpecialNormalized := warrior.RegisterSpell(AnyStance, core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 26654},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.Ternary(warrior.Env.IsForever(), core.ProcMaskMeleeMHSpecial, core.ProcMaskEmpty),
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagNoOnCastComplete,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			damage := spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
			spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMeleeSpecialCritOnly)
		},
	})

	actionID := core.ActionID{SpellID: 12292}

	// Forever's Sweeping Strikes lasts 20 sec (client row 12292, DurationMs 20000; Wowhead's Forever
	// buff tooltip reads "20 seconds remaining").
	ssAura := warrior.RegisterAura(core.Aura{
		Label:     "Sweeping Strikes",
		ActionID:  actionID,
		Duration:  core.TernaryDuration(warrior.Env.IsForever(), time.Second*20, time.Second*10),
		MaxStacks: 5,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.SetStacks(sim, 5)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if aura.GetStacks() == 0 || result.Damage <= 0 || !spell.ProcMask.Matches(core.ProcMaskMelee) || spell == hitSpecialNormalized.Spell {
				return
			}

			var spellToUse *WarriorSpell
			if spell.SpellCode == SpellCode_WarriorWhirlwind {
				spellToUse = hitSpecialNormalized
			} else {
				curDmg = result.Damage
				curDmg /= result.ResistanceMultiplier // Undo armor reduction to get the raw damage value.
				spellToUse = hitSchoolDamagWithValue
			}

			if sim.Encounter.AoETargetCount(0) > 1 {
				target := warrior.Env.NextTargetUnit(result.Target)
				spellToUse.Cast(sim, target)
				spellToUse.SpellMetrics()[target.UnitIndex].Casts--
			} else if warrior.Env.IsForever() {
				// "Sweeping Strikes no longer loses charges if no second target is hit"
				// (Blizzard hotfixes of 2026-11-10).
				return
			}

			aura.RemoveStack(sim)
		},
	})

	SweepingStrikes := warrior.RegisterSpell(BattleStance, core.SpellConfig{
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolPhysical,
		Flags:       core.SpellFlagHelpful,

		RageCost: core.RageCostOptions{
			Cost: 30,
		},
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    warrior.NewTimer(),
				Duration: time.Second * 30,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			ssAura.Activate(sim)
		},
	})

	warrior.AddMajorCooldown(core.MajorCooldown{
		Spell: SweepingStrikes.Spell,
		Type:  core.CooldownTypeDPS,
		ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
			return sim.Encounter.AoETargetCount(0) >= 2
		},
	})
}
