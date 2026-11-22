package warrior

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

func (warrior *Warrior) applyDeepWounds() {
	if warrior.Talents.DeepWounds == 0 {
		return
	}

	spellID := map[int32]int32{
		1: 12834,
		2: 12849,
		3: 12867,
	}[warrior.Talents.DeepWounds]

	flags := core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell | core.SpellFlagNoPeriodicCrit
	if warrior.Env.IsForever() {
		// The warrior's damage bonuses, Death Wish and Enrage among them, don't scale the ticks
		// (wowsims/forever 3af55b0ab8, EV-0139: a level 60 Forever log ticks 17 and 15 with and
		// without Death Wish).
		flags |= core.SpellFlagIgnoreAttackerModifiers
	}

	warrior.DeepWounds = warrior.RegisterSpell(AnyStance, core.SpellConfig{
		SpellCode:   SpellCode_WarriorDeepWounds,
		ActionID:    core.ActionID{SpellID: spellID},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskEmpty,
		// Forever's Deep Wounds (412609) deals its damage through 412613, which cannot crit
		// (client 70338: SPELL_ATTR2_CANT_CRIT, and no Attributes_8 0x200 like the DoTs that crit).
		Flags: flags,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Deep Wounds",
			},
			NumberOfTicks: 4,
			TickLength:    time.Second * 3,

			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				// Under Forever the stored damage is weapon damage only, and the tick takes no
				// damage bonus of the warrior's (the flag above).
				if sim.IsForever() {
					dot.Spell.CalcAndDealPeriodicDamage(sim, target, dot.SnapshotBaseDamage, dot.OutcomeTick)
					return
				}
				attackTable := warrior.AttackTables[target.UnitIndex][proto.CastType_CastTypeMainHand]
				dot.SnapshotAttackerMultiplier = dot.Spell.AttackerDamageMultiplier(attackTable, true) // Double dips on attackers mods
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			dot := spell.Dot(target)
			if sim.IsForever() && dot.IsActive() {
				// Under Forever a crit while the bleed is up keeps the first application's 3 sec
				// tick timer and gives it a full 4 ticks from the next one (wowsims/forever
				// c4a5402aa9, from a Forever combat log).
				nextTick := dot.NextTickAt()
				dot.ApplyOrRefresh(sim)
				dot.UpdateExpires(sim, nextTick+dot.TickPeriod()*time.Duration(dot.NumberOfTicks-1))
			} else {
				dot.Apply(sim) //Resets the tick counter with Apply vs ApplyorRefresh
			}
			spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHitNoHitCounter)
		},
	})

	core.MakePermanent(warrior.RegisterAura(core.Aura{
		Label: "Deep Wounds Talent",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.ProcMask.Matches(core.ProcMaskEmpty) || !spell.SpellSchool.Matches(core.SpellSchoolPhysical) {
				return
			}

			// Ravager doesn't proc Deep Wounds
			if spell.ActionID.SpellID == 9633 {
				return
			}

			// Under Forever a crit on the magic hit table, Thunder Clap's, starts no bleed
			// (wowsims/forever EV-0139: 0 of 79 Thunder Clap crits in Warcraft Logs reports).
			if warrior.Env.IsForever() && spell.DefenseType == core.DefenseTypeMagic {
				return
			}

			if result.Outcome.Matches(core.OutcomeCrit) {
				warrior.procDeepWounds(sim, result.Target, spell.IsOH())
			}
		},
	}))
}

func (warrior *Warrior) procDeepWounds(sim *core.Simulation, target *core.Unit, isOh bool) {
	dot := warrior.DeepWounds.Dot(target)

	if warrior.Env.IsForever() {
		// Under Forever, Deep Wounds doesn't scale with attack power and is a rolling DoT (the
		// 8 October development notes): a new crit adds the damage not yet ticked to the new DoT.
		// The critting hand's average weapon damage sizes it, an off-hand crit's without the
		// off-hand penalty (wowsims/forever EV-0139: level 60 Forever log, off-hand crits tick 15
		// against the main hand's 17, the hands' averages 104.5 and 114.5).
		weapon := core.Ternary(isOh, warrior.AutoAttacks.OH(), warrior.AutoAttacks.MH())
		newDamage := weapon.AverageDamage() * 0.2 * float64(warrior.Talents.DeepWounds)
		if dot.IsActive() {
			newDamage += dot.SnapshotBaseDamage * float64(dot.MaxTicksRemaining())
		}
		dot.SnapshotBaseDamage = newDamage / float64(dot.NumberOfTicks)
		dot.SnapshotAttackerMultiplier = 1

		warrior.DeepWounds.Cast(sim, target)
		return
	}

	ap := dot.Spell.MeleeAttackPower(target)

	var awd float64
	if isOh {
		attackTableOh := warrior.AttackTables[target.UnitIndex][proto.CastType_CastTypeOffHand]
		adm := warrior.AutoAttacks.OHAuto().AttackerDamageMultiplier(attackTableOh, true)
		awd = warrior.AutoAttacks.OH().CalculateAverageWeaponDamage(ap) * 0.5 * adm
	} else { // MH
		attackTableMh := warrior.AttackTables[target.UnitIndex][proto.CastType_CastTypeMainHand]
		adm := warrior.AutoAttacks.MHAuto().AttackerDamageMultiplier(attackTableMh, true)
		awd = warrior.AutoAttacks.MH().CalculateAverageWeaponDamage(ap) * adm
	}

	newDamage := awd * 0.2 * float64(warrior.Talents.DeepWounds) // 60% of average attackers damage

	dot.SnapshotBaseDamage = newDamage / 4.0 // spread over 4 ticks of the dot
	dot.SnapshotAttackerMultiplier = 1

	warrior.DeepWounds.Cast(sim, target)
}
