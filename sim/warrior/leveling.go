package warrior

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// What only matters to a warrior alone in the open world: abilities that need a kill, several mobs or
// a mob that can be stunned, and Rage between pulls (docs/leveling_warrior.md).

const (
	victoryRushID       = 402927
	retaliationStrikeID = 20240
)

// enableLeveling wires up what a leveling session adds: mobs we can stun, and Rage that decays out of
// combat. Anger Management has it decay 30% slower (12296), and Lingering Rage has it start 2 sec
// later a point (1323964, client 71091).
func (warrior *Warrior) enableLeveling() {
	if !warrior.Env.Encounter.IsLeveling() {
		return
	}
	warrior.Env.RegisterStuns()
	warrior.EnableLevelingRageDecay(core.RageDecayOptions{
		Multiplier: core.TernaryFloat64(warrior.Talents.AngerManagement, 0.7, 1),
		Delay:      2 * time.Second * time.Duration(warrior.Talents.LingeringRage),
		Metrics:    warrior.NewRageMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionRageGain, Tag: 2}),
	})
}

// Victory Rush deals 1 plus 15% of attack power, costs no Rage and heals 10% of max health. It can be
// used for 20 sec after we kill a mob that gives experience (client 71091, 402927 and 402975), so
// only a leveling session has it. The heal is an effect of its own on us, so we take it as healing
// on a miss too.
func (warrior *Warrior) registerVictoryRushSpell() {
	if !warrior.Env.Encounter.IsLeveling() || !warrior.KnowsRank(victoryRushID, 20) {
		return
	}
	actionID := core.ActionID{SpellID: victoryRushID}
	healthMetrics := warrior.NewHealthMetrics(actionID)

	victorious := warrior.RegisterAura(core.Aura{
		Label:    "Victorious",
		ActionID: core.ActionID{SpellID: 402975},
		Duration: time.Second * 20,
	})
	warrior.RegisterKillEffect(func(sim *core.Simulation, _ *core.Unit, gaveXP bool) {
		if gaveXP {
			victorious.Activate(sim)
		}
	})

	warrior.VictoryRush = warrior.RegisterSpell(AnyStance, core.SpellConfig{
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    warrior.NewTimer(),
				Duration: time.Second * 30,
			},
		},
		ExtraCastCondition: func(_ *core.Simulation, _ *core.Unit) bool {
			return victorious.IsActive()
		},

		CritDamageBonus: warrior.impale(),

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			victorious.Deactivate(sim)
			damage := 1 + 0.15*spell.MeleeAttackPower(target)
			spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMeleeSpecialHitAndCrit)
			warrior.GainHealth(sim, 0.1*warrior.MaxHealth()*warrior.PseudoStats.HealingTakenMultiplier, healthMetrics)
		},
	})
}

// Retaliation counterattacks every mob that hits us in melee for 15 sec, at most 30 times, with a
// plain weapon hit (client 71091, 20230 and 20240). It's on a 15 min cooldown, in Battle Stance. We
// take it to answer every melee hit that lands, blocked ones too.
func (warrior *Warrior) registerRetaliationSpell() {
	if !warrior.KnowsRank(20230, 20) {
		return
	}

	strike := warrior.RegisterSpell(AnyStance, core.SpellConfig{
		ActionID:    core.ActionID{SpellID: retaliationStrikeID},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

		CritDamageBonus: warrior.impale(),

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			damage := spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target))
			spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		},
	})

	aura := warrior.RegisterAura(core.Aura{
		Label:     "Retaliation",
		ActionID:  core.ActionID{SpellID: 20230},
		Duration:  time.Second * 15,
		MaxStacks: 30,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.SetStacks(sim, 30)
		},
		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.ProcMask.Matches(core.ProcMaskMelee) && result.Landed() && spell.Unit.Type == core.EnemyUnit {
				aura.RemoveStack(sim)
				strike.Cast(sim, spell.Unit)
			}
		},
	})

	warrior.Retaliation = warrior.RegisterSpell(BattleStance, core.SpellConfig{
		ActionID: core.ActionID{SpellID: 20230},
		Flags:    core.SpellFlagAPL | core.SpellFlagHelpful,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    warrior.NewTimer(),
				Duration: time.Minute * 15,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			aura.Activate(sim)
		},
	})
}

// Concussion Blow stuns a mob for 5 sec for 10 Rage, off the global cooldown, and deals no damage
// (client 71091, 12809). A boss is immune, so only a leveling session has it.
func (warrior *Warrior) registerConcussionBlowSpell() {
	if !warrior.Env.Encounter.IsLeveling() || !warrior.Talents.ConcussionBlow {
		return
	}

	warrior.ConcussionBlow = warrior.RegisterSpell(AnyStance, core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 12809},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL | SpellFlagOffensive,

		RageCost: core.RageCostOptions{
			Cost:   10,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    warrior.NewTimer(),
				Duration: time.Second * 45,
			},
		},

		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMeleeSpecialHit)
			if result.Landed() {
				target.Stun(sim, time.Second*5)
			} else {
				spell.IssueRefund(sim)
			}
		},
	})
}

// Gore Drinker: after Enrage, Berserker Rage, Bloodrage, Death Wish, Reckless Swings or Bloodthirst,
// our next 3 melee attacks that land within 8 sec each heal 0.5 or 1% of our max health (client
// 71091: 1323967, whose buff 1323969 procs on melee swings and melee abilities, and curve 125264).
// No tank can reach it and the sim heals no DPS in a raid, so only a leveling session has it.
func (warrior *Warrior) applyGoreDrinker() {
	points := warrior.Talents.GoreDrinker
	if points == 0 || !warrior.Env.Encounter.IsLeveling() {
		return
	}
	healthMetrics := warrior.NewHealthMetrics(core.ActionID{SpellID: 1323968})
	share := 0.005 * float64(points)

	warrior.GoreDrinkerAura = warrior.RegisterAura(core.Aura{
		Label:     "Gore Drinker",
		ActionID:  core.ActionID{SpellID: 1323969},
		Duration:  time.Second * 8,
		MaxStacks: 3,
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !spell.ProcMask.Matches(core.ProcMaskMelee) || !result.Landed() || result.Target.Type != core.EnemyUnit {
				return
			}
			heal := share * warrior.MaxHealth() * warrior.PseudoStats.HealingTakenMultiplier
			warrior.GainHealth(sim, heal, healthMetrics)
			aura.RemoveStack(sim)
		},
	})

	core.MakePermanent(warrior.RegisterAura(core.Aura{
		Label: "Gore Drinker Trigger",
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			for _, trigger := range []*WarriorSpell{warrior.BerserkerRage, warrior.Bloodrage, warrior.DeathWish,
				warrior.RecklessSwings, warrior.Bloodthirst} {
				if trigger.IsEqual(spell) {
					warrior.goreDrinker(sim)
					return
				}
			}
		},
	}))
}

// goreDrinker starts Gore Drinker, or refills it.
func (warrior *Warrior) goreDrinker(sim *core.Simulation) {
	if warrior.GoreDrinkerAura != nil {
		warrior.GoreDrinkerAura.Activate(sim)
		warrior.GoreDrinkerAura.SetStacks(sim, 3)
	}
}
