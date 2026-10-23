package druid

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

type RipRankInfo struct {
	id              int32
	level           int32
	dmgTickBase     float64
	dmgTickPerCombo float64
}

// Rip under Forever (1.60.1.70338, wago.tools): rank 6 ticks for 15 plus 25.5 a combo
// point (1.12 has 17 and 28), 10% less. The attack power part is left to the server, the
// sim keeps 1.12's (see newRipSpellConfig).
var ripRanks = []RipRankInfo{
	{id: 1079, level: 20, dmgTickBase: 3, dmgTickPerCombo: 4.4},
	{id: 9492, level: 28, dmgTickBase: 4, dmgTickPerCombo: 7.2},
	{id: 9493, level: 36, dmgTickBase: 6, dmgTickPerCombo: 8.5},
	{id: 9752, level: 44, dmgTickBase: 8, dmgTickPerCombo: 12.7},
	{id: 9894, level: 52, dmgTickBase: 11, dmgTickPerCombo: 18.2},
	{id: 9896, level: 60, dmgTickBase: 15, dmgTickPerCombo: 25.5},
}

const RipTicks int32 = 6

func (druid *Druid) registerRipSpell() {
	// Add highest available Rip rank for level.
	for rank := len(ripRanks) - 1; rank >= 0; rank-- {
		if druid.Level >= ripRanks[rank].level {
			config := druid.newRipSpellConfig(ripRanks[rank])
			druid.Rip = druid.RegisterSpell(Cat, config)
			return
		}
	}
}

func (druid *Druid) newRipSpellConfig(ripRank RipRankInfo) core.SpellConfig {
	energyCost := 30.0
	// Every Forever Rip rank carries ATTR_EX_1_DISCOUNT_POWER_ON_MISS (0x8000000, client rows
	// 1079 to 9896), so a miss gives back 80% of the cost (wowsims/forever's rip.go reads it).
	refund := core.TernaryFloat64(druid.Env.IsForever(), 0.8, 0)

	// The combo points the Rip went up with, per target.
	comboPoints := map[int32]float64{}
	// AP scaling is 6% per combo point from 1 to 4, and 24% again for 5
	tickDamage := func(target *core.Unit, spell *core.Spell) float64 {
		cp := comboPoints[target.UnitIndex]
		cpScaling := core.TernaryFloat64(cp == 5, 4, cp)
		return ripRank.dmgTickBase + ripRank.dmgTickPerCombo*cp + 0.01*cpScaling*spell.MeleeAttackPower(target)
	}

	return core.SpellConfig{
		SpellCode:   SpellCode_DruidRip,
		ActionID:    core.ActionID{SpellID: ripRank.id},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       SpellFlagOmen | core.SpellFlagMeleeMetrics | core.SpellFlagAPL | core.SpellFlagPureDot,

		EnergyCost: core.EnergyCostOptions{
			Cost:   energyCost,
			Refund: refund,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return druid.ComboPoints() > 0
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Rip",
			},
			NumberOfTicks: RipTicks + druid.RipBonusTicks,
			TickLength:    time.Second * 2,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				comboPoints[target.UnitIndex] = float64(druid.ComboPoints())
				dot.Snapshot(target, tickDamage(target, dot.Spell), isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				// Under Forever each tick reads attack power and the damage modifiers when it
				// lands; only the combo points are kept from the cast (wowsims/forever's rip.go,
				// as for Rake, Rend, Rupture and Garrote: 36cfc58328, "Periodic damage does not
				// snapshot in Forever").
				if sim.IsForever() {
					damage := tickDamage(target, dot.Spell) * druid.RendAndTearMultiplier(target)
					dot.Spell.CalcAndDealPeriodicDamage(sim, target, damage, dot.OutcomeTick)
					return
				}
				rendAndTear := druid.RendAndTearMultiplier(target)
				dot.SnapshotAttackerMultiplier *= rendAndTear
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
				dot.SnapshotAttackerMultiplier /= rendAndTear
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHitNoHitCounter)
			if result.Landed() {
				dot := spell.Dot(target)
				dot.Apply(sim)
				druid.SpendComboPoints(sim, spell)
			} else {
				spell.IssueRefund(sim)
			}
			spell.DealOutcome(sim, result)
		},
	}
}

func (druid *Druid) CurrentRipCost() float64 {
	return druid.Rip.Cost.GetCurrentCost()
}
