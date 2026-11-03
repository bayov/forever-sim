package druid

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// The bear's attacks under Forever (bear.go has the sources). None of them states a threat effect
// in the client, and their tooltips' "high amount of threat" (Lacerate, Primal Bite) isn't a number
// anywhere, so each threatens for its damage times the form's 1.5, as wowsims/forever has it.
// Classic's Maul and Swipe had 1.75 on top; that's the first thing a log of a bear tank settles.

// Maul ranks: level and the flat damage on top of the weapon's (client 9881 rank 7 at 58: 128).
var maulRanks = []struct {
	level   int32
	spellID int32
	damage  float64
}{
	{10, 6807, 18}, {18, 6808, 27}, {26, 6809, 37}, {34, 8972, 49}, {42, 9745, 71}, {50, 9880, 101}, {58, 9881, 128},
}

// registerMaulSpell is Maul, the bear's next swing: the weapon's damage plus the rank's, 15 Rage
// (Ferocity 1 a point less), queued like a warrior's Heroic Strike and struck in the swing's place.
func (druid *Druid) registerMaulSpell() {
	rank := maulRanks[0]
	for _, r := range maulRanks {
		if druid.Level >= r.level {
			rank = r
		}
	}
	actionID := core.ActionID{SpellID: rank.spellID}

	maul := druid.RegisterSpell(Bear, core.SpellConfig{
		ActionID:    actionID.WithTag(1),
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagNoOnCastComplete,

		RageCost: core.RageCostOptions{
			Cost:   15 - float64(druid.Talents.Ferocity),
			Refund: 0.8,
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			if druid.MaulQueueAura != nil {
				druid.MaulQueueAura.Deactivate(sim)
			}
			damage := (rank.damage + spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target))) *
				druid.RendAndTearMultiplier(target)
			result := spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
			if !result.Landed() {
				spell.IssueRefund(sim)
			}
		},
	})
	druid.Maul = maul

	realismICD := &core.Cooldown{Timer: druid.NewTimer(), Duration: time.Millisecond * 50}
	queued := false
	druid.MaulQueueAura = druid.RegisterAura(core.Aura{
		Label:    "Maul Queue",
		ActionID: actionID,
		Duration: core.NeverExpires,
		OnReset: func(_ *core.Aura, _ *core.Simulation) {
			queued = false
		},
	})
	druid.MaulQueueSpell = druid.RegisterSpell(Bear, core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagMeleeMetrics | core.SpellFlagAPL | core.SpellFlagCastTimeNoGCD,
		ExtraCastCondition: func(sim *core.Simulation, _ *core.Unit) bool {
			return !druid.MaulQueueAura.IsActive() && !queued &&
				druid.CurrentRage() >= maul.DefaultCast.Cost && realismICD.IsReady(sim)
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			queued = true
			realismICD.Use(sim)
			sim.AddPendingAction(&core.PendingAction{
				NextActionAt: sim.CurrentTime + realismICD.Duration,
				OnAction: func(sim *core.Simulation) {
					queued = false
					if druid.InForm(Bear) {
						druid.MaulQueueAura.Activate(sim)
					}
				},
			})
		},
	})
}

// TryMaul strikes the queued Maul in the swing's place when it can, else the swing.
func (druid *Druid) TryMaul(sim *core.Simulation, mhSwingSpell *core.Spell) *core.Spell {
	if druid.MaulQueueAura == nil || !druid.MaulQueueAura.IsActive() {
		return mhSwingSpell
	}
	if !druid.Maul.CanCast(sim, druid.CurrentTarget) {
		druid.MaulQueueAura.Deactivate(sim)
		return mhSwingSpell
	}
	return druid.Maul.Spell
}

// Lacerate (Feral row 4 talent; rank 5, 1235827 at 58): 15 Rage (Shredding Attacks 1 a point less),
// a bleed of 15 a stack every 3 sec for 15 sec, up to 5 stacks, and a hit of 10% of weapon damage
// for each stack already on the target (the tree's "plus 10% weapon damage per existing
// application", so the first one hits for nothing).
func (druid *Druid) registerLacerateSpell() {
	if !druid.Talents.Lacerate {
		return
	}
	type lacerateRank struct {
		level   int32
		spellID int32
		tick    float64
	}
	ranks := []lacerateRank{{42, 414644, 10}, {50, 1235826, 12}, {58, 1235827, 15}}
	if druid.Level < ranks[0].level {
		return
	}
	rank := ranks[0]
	for _, r := range ranks {
		if druid.Level >= r.level {
			rank = r
		}
	}

	druid.Lacerate = druid.RegisterSpell(Bear, core.SpellConfig{
		ActionID:    core.ActionID{SpellID: rank.spellID},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		RageCost: core.RageCostOptions{
			Cost:   15 - float64(druid.Talents.ShreddingAttacks),
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{GCD: core.GCDDefault},
			IgnoreHaste: true,
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label:     "Lacerate",
				MaxStacks: 5,
			},
			NumberOfTicks: 5,
			TickLength:    time.Second * 3,
			OnSnapshot: func(_ *core.Simulation, target *core.Unit, dot *core.Dot, _ bool) {
				attackTable := dot.Spell.Unit.AttackTables[target.UnitIndex][dot.Spell.CastType]
				dot.SnapshotAttackerMultiplier = dot.Spell.AttackerDamageMultiplier(attackTable, true)
				dot.SnapshotBaseDamage = rank.tick * float64(dot.GetStacks())
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			dot := spell.Dot(target)
			existing := int32(0)
			if dot.IsActive() {
				existing = dot.GetStacks()
			}
			weapon := spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target))
			damage := 0.1 * float64(existing) * weapon * druid.RendAndTearMultiplier(target)
			result := spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
			if !result.Landed() {
				spell.IssueRefund(sim)
				return
			}
			dot.ApplyOrRefresh(sim)
			if dot.GetStacks() < dot.MaxStacks {
				dot.AddStack(sim)
			}
			dot.TakeSnapshot(sim, true)
		},
	})
}

// Primal Bite (Feral row 7 talent; rank 4, 1238073 at 60, bear only): 100% weapon damage plus 77,
// 20 Rage (Ferocity 1 a point less), 6 sec cooldown.
func (druid *Druid) registerPrimalBiteSpell() {
	if !druid.Talents.PrimalBite {
		return
	}
	ranks := []struct {
		level   int32
		spellID int32
		damage  float64
	}{{25, 407995, 26}, {36, 1238069, 38}, {48, 1238070, 59}, {60, 1238073, 77}}
	if druid.Level < ranks[0].level {
		return
	}
	rank := ranks[0]
	for _, r := range ranks {
		if druid.Level >= r.level {
			rank = r
		}
	}

	druid.PrimalBite = druid.RegisterSpell(Bear, core.SpellConfig{
		ActionID:    core.ActionID{SpellID: rank.spellID},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		RageCost: core.RageCostOptions{
			Cost:   20 - float64(druid.Talents.Ferocity),
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{GCD: core.GCDDefault},
			IgnoreHaste: true,
			CD:          core.Cooldown{Timer: druid.NewTimer(), Duration: time.Second * 6},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			damage := (rank.damage + spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target))) *
				druid.RendAndTearMultiplier(target)
			result := spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
			if !result.Landed() {
				spell.IssueRefund(sim)
			}
		},
	})
}
