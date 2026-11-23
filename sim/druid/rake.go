package druid

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

type RakeRankInfo struct {
	id            int32
	level         int32
	initialDamage float64
	dotTickDamage float64
}

// Rake under Forever (1.60.1.70338, wago.tools): a little more than 1.12 on every rank
// (rank 4 is 61 and 34 a tick, 1.12 has 58 and 32).
var rakeSpells = []RakeRankInfo{
	{id: 1822, level: 24, initialDamage: 23, dotTickDamage: 16},
	{id: 1823, level: 34, initialDamage: 31, dotTickDamage: 21},
	{id: 1824, level: 44, initialDamage: 45, dotTickDamage: 26},
	{id: 9904, level: 54, initialDamage: 61, dotTickDamage: 34},
}

// rakeTickAPCoefficient is the attack power share of a Forever Rake tick, every rank, outside the
// November 2 scale. The client leaves it to the server. wowsims/forever's logs (builds 70862 and
// 70888, after the November 2 notes) read the hit's share on every rank: EV-0031 rank 1
// (2026-11-07, 0.0652, 55 of 55 ticks at 221 attack power), EV-0033 rank 2 (2026-11-08, 1,157 of
// 1,232 ticks), EV-0034 rank 3 (2026-11-10, 683 of 686 ticks) and EV-0036 rank 4 (2026-11-12, 237
// of 238 ticks), all at 0.065 (user ruling 2026-11-22: a share measured from a Forever log wins).
// The sim had testers' 0.055 (PR #149, 2026-10-01) × 1.25 = 0.06875. The tick reads the attack
// power when it lands (EV-0032: 334 of 337 ticks fit the live attack power).
const rakeTickAPCoefficient = 0.065

// rakeHitAPCoefficient is the attack power share of a Forever Rake's direct hit, every rank, read
// at the hit and outside the November 2 scale (the user's ruling, 2026-11-12). wowsims/forever
// 0384d8f98c fits 0.0648 to a 1.60.1.70862 combat log of a Tauren druid at levels 22 to 25: rank
// 1's hit read 45 on 26 of 26 hits at 221 attack power with Savage Fury 2/2, a weapon damage buff
// didn't move it. Their later logs read 0.065 on ranks 2 to 4 too (EV-0033, EV-0034 and EV-0036,
// 2026-11-08 to 12: 514 of 530, 347 of 349 and 121 of 121 hits). Their older PR #149 tester had
// the hit take none.
const rakeHitAPCoefficient = 0.065

func (druid *Druid) registerRakeSpell() {
	// Add highest available rake rank for level.
	for rank := len(rakeSpells) - 1; rank >= 0; rank-- {
		if druid.Level >= rakeSpells[rank].level {
			config := druid.newRakeSpellConfig(rakeSpells[rank])
			druid.Rake = druid.RegisterSpell(Cat, config)
			return
		}
	}
}

func (druid *Druid) newRakeSpellConfig(rakeRank RakeRankInfo) core.SpellConfig {
	// The November 2 development notes raise the base points and the attack power part by about
	// 25% (tuning_70753.go).
	scale := core.TernaryFloat64(druid.Env.IsForever(), nov2BleedScale, 1)
	baseDamageInitial := scale * rakeRank.initialDamage
	baseDamageTick := scale * rakeRank.dotTickDamage
	energyCost := 40 - float64(druid.Talents.Ferocity)
	tickAPCoefficient := core.TernaryFloat64(druid.Env.IsForever(), rakeTickAPCoefficient, 0)
	hitAPCoefficient := core.TernaryFloat64(druid.Env.IsForever(), rakeHitAPCoefficient, 0)
	hitBase := func(target *core.Unit, spell *core.Spell) float64 {
		return baseDamageInitial + hitAPCoefficient*spell.MeleeAttackPower(target)
	}
	tickBase := func(target *core.Unit, spell *core.Spell) float64 {
		return (baseDamageTick + tickAPCoefficient*spell.MeleeAttackPower(target)) * druid.RendAndTearMultiplier(target)
	}

	return core.SpellConfig{
		SpellCode:   SpellCode_DruidRake,
		ActionID:    core.ActionID{SpellID: rakeRank.id},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagIgnoreResists | core.SpellFlagBinary | core.SpellFlagAPL | SpellFlagOmen | SpellFlagBuilder,

		EnergyCost: core.EnergyCostOptions{
			Cost:   energyCost,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Rake",
			},
			NumberOfTicks: 3,
			TickLength:    time.Second * 3,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				damage := baseDamageTick
				dot.Snapshot(target, damage, isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.SnapshotBaseDamage = tickBase(target, dot.Spell) * dot.Spell.BaseDamageMultiplierAdditive
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := hitBase(target, spell)
			rendAndTear := druid.RendAndTearMultiplier(target)
			spell.DamageMultiplier *= rendAndTear
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
			spell.DamageMultiplier /= rendAndTear

			if result.Landed() {
				druid.AddComboPoints(sim, 1, target, spell.ComboPointMetrics())
				spell.Dot(target).Apply(sim)
			} else {
				spell.IssueRefund(sim)
			}
		},

		ExpectedInitialDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
			baseDamage := hitBase(target, spell)
			initial := spell.CalcPeriodicDamage(sim, target, baseDamage, spell.OutcomeExpectedMagicAlwaysHit)

			attackTable := spell.Unit.AttackTables[target.UnitIndex][spell.CastType]
			critChance := spell.PhysicalCritChance(attackTable)
			critMod := critChance * (spell.CritMultiplier(attackTable) - 1)
			initial.Damage *= 1 + critMod
			return initial
		},
		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
			ticks := spell.CalcPeriodicDamage(sim, target, tickBase(target, spell), spell.OutcomeExpectedMagicAlwaysHit)
			return ticks
		},
	}
}

func (druid *Druid) CurrentRakeCost() float64 {
	return druid.Rake.Cost.GetCurrentCost()
}
