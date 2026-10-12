package hunter

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Forever's Mongoose Bite (beta client, build 1.60.1.70338) is normalized melee weapon damage plus
// 15/22/37/57 by rank, where Classic dealt 25/45/75/115 alone. A dodge opens it for 5 sec as in
// Classic, and Expose Prey opens it for 10 (talents.go).
func (hunter *Hunter) getMongooseBiteConfig(rank int) core.SpellConfig {
	spellId := [5]int32{0, 1495, 14269, 14270, 14271}[rank]
	baseDamage := [5]float64{0, 15, 22, 37, 57}[rank]
	manaCost := [5]float64{0, 30, 40, 50, 65}[rank]
	level := [5]int{0, 16, 30, 44, 58}[rank]

	spellConfig := core.SpellConfig{
		SpellCode:     SpellCode_HunterMongooseBite,
		ActionID:      core.ActionID{SpellID: spellId},
		SpellSchool:   core.SpellSchoolPhysical,
		DefenseType:   core.DefenseTypeMelee,
		ProcMask:      core.ProcMaskMeleeMHSpecial,
		Flags:         core.SpellFlagMeleeMetrics | core.SpellFlagAPL | SpellFlagStrike,
		Rank:          rank,
		RequiredLevel: level,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    hunter.NewTimer(),
				Duration: time.Second * 5,
			},
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return hunter.inMeleeRange() && hunter.DefensiveState.IsActive()
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			hunter.DefensiveState.Deactivate(sim)
			damage := baseDamage + hunter.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
			result := spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

			if hunter.LaceratingStrikes != nil && result.Landed() {
				hunter.procLaceratingStrikes(sim, result)
			}
		},
	}

	return spellConfig
}

func (hunter *Hunter) registerMongooseBiteSpell() {
	// Only a pre-requisite for Mongoose Bite: a dodge opens it, and so does Expose Prey.
	hunter.DefensiveState = hunter.RegisterAura(core.Aura{
		Label:    "Defensive State",
		ActionID: core.ActionID{SpellID: 1310726},
		Duration: time.Second * 5,
	})
	core.MakePermanent(hunter.RegisterAura(core.Aura{
		Label: "Defensive State Trigger",
		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.DidDodge() {
				hunter.DefensiveState.Duration = time.Second * 5
				hunter.DefensiveState.Activate(sim)
			}
		},
	}))

	for rank := 4; rank >= 1; rank-- {
		config := hunter.getMongooseBiteConfig(rank)
		if config.RequiredLevel <= int(hunter.Level) {
			hunter.MongooseBite = hunter.GetOrRegisterSpell(config)
			break
		}
	}

	hunter.registerLaceratingStrikes()
}

// Lacerating Strikes, the 31 point Survival talent: a landed Mongoose Bite also bleeds the target for
// 40% of its damage over 21 sec, a tick every 3 sec (client 1310536). The ticks can crit since the
// 2026-10-01 build. A new bite replaces the bleed. The 40% is of the bite as it landed, so the ticks
// skip the damage modifiers it already took.
func (hunter *Hunter) registerLaceratingStrikes() {
	if !hunter.Talents.LaceratingStrikes || hunter.MongooseBite == nil {
		return
	}

	hunter.LaceratingStrikes = hunter.RegisterSpell(core.SpellConfig{
		SpellCode:   SpellCode_HunterLaceratingStrikes,
		ActionID:    core.ActionID{SpellID: 1310536},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagNoOnCastComplete | core.SpellFlagIgnoreModifiers,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Lacerating Strikes" + hunter.Label,
			},
			NumberOfTicks: 7,
			TickLength:    time.Second * 3,

			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.Dot(target).Apply(sim)
		},
	})
}

func (hunter *Hunter) procLaceratingStrikes(sim *core.Simulation, result *core.SpellResult) {
	dot := hunter.LaceratingStrikes.Dot(result.Target)
	hunter.LaceratingStrikes.Cast(sim, result.Target)
	dot.SnapshotBaseDamage = result.Damage * 0.4 / float64(dot.NumberOfTicks)
	dot.SnapshotAttackerMultiplier = 1
}
