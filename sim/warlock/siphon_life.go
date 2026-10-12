package warlock

import (
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core"
)

// The talent teaches rank 1 and the trainer the other three (client: learned at 30, 38, 48
// and 58).
func (warlock *Warlock) getSiphonLifeBaseConfig(rank int) core.SpellConfig {
	data := siphonLifeRanks[rank]
	baseDamage := data.damageAt(warlock.Level)
	spellId := data.SpellID
	manaCost := data.Cost
	level := data.Level

	spellCoeff := data.Coef
	actionID := core.ActionID{SpellID: spellId}
	healthMetrics := warlock.NewHealthMetrics(actionID)

	return core.SpellConfig{
		ActionID:      actionID,
		SpellCode:     SpellCode_WarlockSiphonLife,
		SpellSchool:   core.SpellSchoolShadow,
		DefenseType:   core.DefenseTypeMagic,
		ProcMask:      core.ProcMaskSpellDamage,
		Flags:         core.SpellFlagAPL | core.SpellFlagResetAttackSwing | core.SpellFlagBinary | WarlockFlagAffliction,
		RequiredLevel: level,
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		DamageMultiplierAdditive: 1,
		DamageMultiplier:         1,
		ThreatMultiplier:         1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "SiphonLife-" + warlock.Label + strconv.Itoa(rank),
			},
			NumberOfTicks:       10,
			TickLength:          3 * time.Second,
			AffectedByCastSpeed: false,
			BonusCoefficient:    spellCoeff,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, baseDamage, isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				// The target modifiers count per tick, like the warlock's other damage over time
				// spells, so the warlock's own Improved Shadow Bolt debuff only counts while it is
				// up (1.12's Siphon Life took them when it was cast).
				result := dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)

				health := result.Damage
				warlock.GainHealth(sim, health, healthMetrics)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
			if result.Landed() {
				dot := spell.Dot(target)
				dot.Apply(sim)
			}
		},
		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
			if useSnapshot {
				dot := spell.Dot(target)
				return dot.CalcSnapshotDamage(sim, target, spell.OutcomeExpectedMagicAlwaysHit)
			} else {
				return spell.CalcPeriodicDamage(sim, target, baseDamage, spell.OutcomeExpectedMagicAlwaysHit)
			}
		},
	}
}

func (warlock *Warlock) registerSiphonLifeSpell() {
	if !warlock.Talents.SiphonLife {
		return
	}

	warlock.SiphonLife = make([]*core.Spell, 0)
	for rank := 1; rank < len(siphonLifeRanks); rank++ {
		config := warlock.getSiphonLifeBaseConfig(rank)

		if config.RequiredLevel <= int(warlock.Level) {
			warlock.SiphonLife = append(warlock.SiphonLife, warlock.GetOrRegisterSpell(config))
		}
	}
}
