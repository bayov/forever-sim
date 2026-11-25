package hunter

import (
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// The Forever trap cooldown is 30 sec (Classic 15). The damage is Classic's (client 14315, build
// 1.60.1.70338: 208 to 264 and 330 over 20 sec at rank 3), and under Forever the hit adds 10% of
// ranged attack power, read when the trap goes off; the base rolls and the share does not (EV-0134:
// 2,369 level 60 first hits). The ticks take no attack power (EV-0068: two level 34 and 36 combat
// logs, none of 1,451 ticks missed and none moved with ranged attack power). Before, the sim gave the
// ticks 1% a tick, wowsims/forever's earlier reading of the 15 October dev notes' "Fire traps now
// correctly scale with Attack Power" (3993d79540).
const ExplosiveTrapRAPOnHit = 0.10

func (hunter *Hunter) getExplosiveTrapConfig(rank int, timer *core.Timer) core.SpellConfig {
	spellId := [4]int32{0, 13813, 14316, 14317}[rank]
	dotDamage := [4]float64{0, 15, 24, 33}[rank]
	minDamage := [4]float64{0, 104, 145, 208}[rank]
	maxDamage := [4]float64{0, 135, 193, 264}[rank]
	manaCost := [4]float64{0, 275, 395, 520}[rank]
	level := [4]int{0, 34, 44, 54}[rank]

	return core.SpellConfig{
		SpellCode:     SpellCode_HunterExplosiveTrap,
		ActionID:      core.ActionID{SpellID: spellId},
		SpellSchool:   core.SpellSchoolFire,
		DefenseType:   core.DefenseTypeMagic,
		ProcMask:      core.ProcMaskSpellDamage,
		Flags:         core.SpellFlagAPL | SpellFlagTrap,
		Rank:          rank,
		RequiredLevel: level,
		MissileSpeed:  24,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    timer,
				Duration: time.Second * 30,
			},
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true, // Hunter GCD is locked at 1.5s
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			IsAOE: true,
			Aura: core.Aura{
				Label: "ExplosiveTrap" + hunter.Label + strconv.Itoa(rank),
				Tag:   "ExplosiveTrap",
			},
			NumberOfTicks: 10,
			TickLength:    time.Second * 2,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, dotDamage, isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
					// Explosive Trap DoT only does damage if the target does not have an immolation trap ticking on them
					if !aoeTarget.HasActiveAuraWithTag("ImmolationTrap") {
						dot.CalcAndDealPeriodicSnapshotDamage(sim, aoeTarget, dot.OutcomeTick)
					}
				}
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			if hunter.DistanceFromTarget > 5 {
				return
			}

			spell.WaitTravelTime(sim, func(s *core.Simulation) {
				// Classic's traps gain no benefit from hit bonuses except for the Trap Mastery talent,
				// since this is a unique interaction this is my workaround. Forever's roll with the
				// hunter's spell hit, Survival Tactics' trap hit on top (wowsims/forever's traps.go).
				spellHit := 0.0
				if !hunter.Env.IsForever() {
					spellHit = spell.Unit.GetStat(stats.SpellHit) + target.PseudoStats.BonusSpellHitRatingTaken
				}
				if spellHit != 0 {
					spell.Unit.AddStatDynamic(sim, stats.SpellHit, -spellHit)
				}
				for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
					// EV-0134: the base rolls, the attack power share is added after the roll.
					baseDamage := sim.Roll(minDamage, maxDamage)
					if sim.IsForever() {
						baseDamage += ExplosiveTrapRAPOnHit * spell.RangedAttackPower(aoeTarget, false)
					}
					baseDamage *= sim.Encounter.AOECapMultiplier()
					spell.CalcAndDealDamage(sim, aoeTarget, baseDamage, spell.OutcomeMagicHitAndCrit)
				}
				if spellHit != 0 {
					spell.Unit.AddStatDynamic(sim, stats.SpellHit, spellHit)
				}
				spell.AOEDot().ApplyOrReset(sim)
			})
		},
	}
}

func (hunter *Hunter) registerExplosiveTrapSpell(timer *core.Timer) {
	maxRank := 3
	for i := 1; i <= maxRank; i++ {
		config := hunter.getExplosiveTrapConfig(i, timer)

		if config.RequiredLevel <= int(hunter.Level) {
			hunter.ExplosiveTrap = hunter.GetOrRegisterSpell(config)
		}
	}
}
