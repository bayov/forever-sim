package warlock

import (
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (warlock *Warlock) getImmolateConfig(rank int) core.SpellConfig {
	data := immolateRanks[rank]
	baseDamage := data.damageAt(warlock.Level)
	dotData := immolateDotRanks[rank]
	dotDamage := dotData.damageAt(warlock.Level)

	return core.SpellConfig{
		SpellCode:   SpellCode_WarlockImmolate,
		ActionID:    core.ActionID{SpellID: data.SpellID},
		SpellSchool: core.SpellSchoolFire,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       core.SpellFlagAPL | core.SpellFlagResetAttackSwing | core.SpellFlagBinary | WarlockFlagDestruction,

		Rank:          rank,
		RequiredLevel: data.Level,

		ManaCost: core.ManaCostOptions{
			FlatCost: data.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: data.castTime(),
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: data.Coef,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Immolate-" + warlock.Label + strconv.Itoa(rank),
			},

			NumberOfTicks:    5,
			TickLength:       time.Second * 3,
			BonusCoefficient: dotData.Coef,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, dotDamage, isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)

			if result.Landed() {
				dot := spell.Dot(target)
				dot.Apply(sim)
			}

			spell.DealDamage(sim, result)
		},
		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
			if useSnapshot {
				dot := spell.Dot(target)
				return dot.CalcSnapshotDamage(sim, target, dot.Spell.OutcomeExpectedMagicAlwaysHit)
			} else {
				return spell.CalcPeriodicDamage(sim, target, dotDamage, spell.OutcomeExpectedMagicAlwaysHit)
			}
		},
	}
}

func (warlock *Warlock) getActiveImmolateSpell(target *core.Unit) *core.Spell {
	for _, immolateSpell := range warlock.Immolate {
		if immolateSpell.Dot(target).IsActive() {
			return immolateSpell
		}
	}
	return nil
}

func (warlock *Warlock) registerImmolateSpell() {
	warlock.Immolate = make([]*core.Spell, 0)

	maxRank := core.TernaryInt(core.IncludeAQ, len(immolateRanks)-1, len(immolateRanks)-2)
	for rank := 1; rank <= maxRank; rank++ {
		config := warlock.getImmolateConfig(rank)

		if config.RequiredLevel <= int(warlock.Level) {
			warlock.Immolate = append(warlock.Immolate, warlock.GetOrRegisterSpell(config))
		}
	}
}
