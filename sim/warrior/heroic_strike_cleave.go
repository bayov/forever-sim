package warrior

import (
	"github.com/wowsims/classic/sim/core"
)

func (warrior *Warrior) registerHeroicStrikeSpell(realismICD *core.Cooldown) {
	levels := []int32{0, 1, 8, 16, 24, 32, 40, 48, 56}
	spellIDs := []int32{0, 78, 284, 285, 1608, 11564, 11565, 11566, 11567}
	// Ranks from the Ahn'Qiraj books are out under Forever (no source in client 70338).
	aq := warrior.Env.IncludeAQRanks()
	if aq {
		levels, spellIDs = append(levels, 60), append(spellIDs, 25286)
	}
	rank := warrior.knownRank(levels, spellIDs)
	if rank == 0 {
		return
	}
	flatDamageBonus := []float64{0, 11, 21, 32, 44, 58, 80, 111, 138, 157}[rank]
	spellID := spellIDs[rank]
	// No known equation under Classic.
	threat := core.TernaryFloat64(aq, 173, 145)
	// Under Forever its flat threat is 2.5 a spell level, 140 for rank 8 (level 56), times the
	// stance: EV-0004 (Rank 4 +60.08 over 104 clean windows), EV-0048 (Rank 5 +80.5) and EV-0174
	// (Rank 6 +100.06, the user's own log). The client states none.
	if warrior.Env.IsForever() {
		threat = 2.5 * float64(levels[rank])
	}

	warrior.HeroicStrike = warrior.RegisterSpell(AnyStance, core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellID},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeMHAuto,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagNoOnCastComplete | SpellFlagOffensive,

		RageCost: core.RageCostOptions{
			Cost:   15 - float64(warrior.Talents.ImprovedHeroicStrike),
			Refund: 0.8,
		},

		CritDamageBonus: warrior.impale(),

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		FlatThreatBonus:  threat,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := flatDamageBonus + spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target))
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

			if !result.Landed() {
				spell.IssueRefund(sim)
			}

			spell.DealDamage(sim, result)
			if warrior.curQueueAura != nil {
				warrior.curQueueAura.Deactivate(sim)
			}
		},
	})
	warrior.HeroicStrikeQueue = warrior.makeQueueSpellsAndAura(warrior.HeroicStrike, realismICD)
}

func (warrior *Warrior) registerCleaveSpell(realismICD *core.Cooldown) {
	levels := []int32{0, 20, 30, 40, 50, 60}
	spellIDs := []int32{0, 845, 7369, 11608, 11609, 20569}
	rank := warrior.knownRank(levels, spellIDs)
	if rank == 0 {
		return
	}
	flatDamageBonus := []float64{0, 5, 10, 18, 32, 50}[rank]
	spellID := spellIDs[rank]
	threat := 100.0
	// Under Forever 96 for rank 5 (level 60), split over the targets it lands on: EV-0050 measured
	// Rank 3 at 62 and Rank 4 at 80 (1.6 a spell level), and Rank 5 carries Rank 4's 1.6 until a level
	// 60 log measures it.
	if warrior.Env.IsForever() {
		threat = 1.6 * float64(levels[rank])
	}

	// Improved Cleave is gone from the Forever tree. Raging Blows, which took 3 Rage off Cleave and
	// Whirlwind, is gone too: hotfix push 113539 (2026-11-03) turns its spell 1310315 into Improved
	// Whirlwind and drops the cost effect.
	rageCost := 20.0

	results := make([]*core.SpellResult, min(int32(2), warrior.Env.GetNumTargets()))

	warrior.Cleave = warrior.RegisterSpell(AnyStance, core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellID},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeMHAuto,
		Flags:       core.SpellFlagMeleeMetrics | SpellFlagOffensive,

		RageCost: core.RageCostOptions{
			Cost: rageCost,
		},

		CritDamageBonus: warrior.impale(),

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		FlatThreatBonus:  threat,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// Only living targets, each one once.
			targets := warrior.areaTargets(sim, target, meleeReach, len(results))
			hits := results[:len(targets)]
			for idx, target := range targets {
				baseDamage := flatDamageBonus + spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target))
				results[idx] = spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
			}

			// Under Forever Cleave's flat threat is its total, split between the targets it lands
			// on (wowsims/forever 913a5bcf69, the owner's in-game threat measurements). Classic
			// gives each target all of it.
			if sim.IsForever() {
				landed := 0
				for _, result := range hits {
					if result.Landed() {
						landed++
					}
				}
				if landed > 1 {
					flatThreat := spell.FlatThreatBonus
					spell.FlatThreatBonus /= float64(landed)
					for _, result := range hits {
						result.Threat = spell.ThreatFromDamage(result.Outcome, result.Damage)
					}
					spell.FlatThreatBonus = flatThreat
				}
			}

			for _, result := range hits {
				spell.DealDamage(sim, result)
			}

			if warrior.curQueueAura != nil {
				warrior.curQueueAura.Deactivate(sim)
			}
		},
	})
	warrior.CleaveQueue = warrior.makeQueueSpellsAndAura(warrior.Cleave, realismICD)
}

func (warrior *Warrior) makeQueueSpellsAndAura(srcSpell *WarriorSpell, realismICD *core.Cooldown) *WarriorSpell {
	isQueueQueued := false

	queueAura := warrior.RegisterAura(core.Aura{
		Label:    "HS/Cleave Queue Aura-" + srcSpell.ActionID.String(),
		ActionID: srcSpell.ActionID.WithTag(1),
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			isQueueQueued = false
		},
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			if warrior.curQueueAura != nil {
				warrior.curQueueAura.Deactivate(sim)
			}
			// Under Forever a queued Heroic Strike or Cleave no longer raises the off-hand's hit
			// chance (the 1 October development notes).
			warrior.PseudoStats.DisableDWMissPenalty = !warrior.Env.IsForever()
			warrior.curQueueAura = aura
			warrior.curQueuedAutoSpell = srcSpell
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			warrior.PseudoStats.DisableDWMissPenalty = false
			warrior.curQueueAura = nil
			warrior.curQueuedAutoSpell = nil
		},
	})

	queueSpell := warrior.RegisterSpell(AnyStance, core.SpellConfig{
		ActionID: srcSpell.ActionID.WithTag(1),
		Flags:    core.SpellFlagMeleeMetrics | core.SpellFlagAPL | core.SpellFlagCastTimeNoGCD,

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return warrior.curQueueAura == nil &&
				!isQueueQueued &&
				warrior.CurrentRage() >= srcSpell.DefaultCast.Cost &&
				!warrior.IsCasting(sim) &&
				realismICD.IsReady(sim)
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			if realismICD.IsReady(sim) {
				isQueueQueued = true
				realismICD.Use(sim)
				sim.AddPendingAction(&core.PendingAction{
					NextActionAt: sim.CurrentTime + realismICD.Duration,
					OnAction: func(sim *core.Simulation) {
						queueAura.Activate(sim)
						isQueueQueued = false
					},
				})
			}
		},
	})

	return queueSpell
}

func (warrior *Warrior) TryHSOrCleave(sim *core.Simulation, mhSwingSpell *core.Spell) *core.Spell {
	if !warrior.curQueueAura.IsActive() {
		return mhSwingSpell
	}

	if !warrior.curQueuedAutoSpell.CanCast(sim, warrior.CurrentTarget) {
		warrior.curQueueAura.Deactivate(sim)
		return mhSwingSpell
	}

	return warrior.curQueuedAutoSpell.Spell
}

// isQueuedSwing says whether spell is Heroic Strike or Cleave, the strikes that take the main
// hand swing's place.
func (warrior *Warrior) isQueuedSwing(spell *core.Spell) bool {
	return (warrior.HeroicStrike != nil && spell == warrior.HeroicStrike.Spell) ||
		(warrior.Cleave != nil && spell == warrior.Cleave.Spell)
}
