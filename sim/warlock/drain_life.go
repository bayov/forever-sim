package warlock

import (
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (warlock *Warlock) getDrainLifeBaseConfig(rank int) core.SpellConfig {
	numTicks := int32(5)

	data := drainLifeRanks[rank]
	spellId := data.SpellID
	spellCoeff := data.Coef
	baseDamage := data.damageAt(warlock.Level)
	manaCost := data.Cost
	level := data.Level

	actionID := core.ActionID{SpellID: spellId}

	healingSpell := warlock.GetOrRegisterSpell(core.SpellConfig{
		ActionID:    actionID.WithTag(1),
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskSpellHealing,
		Flags:       core.SpellFlagPassiveSpell | core.SpellFlagHelpful,

		DamageMultiplier: 1,
		ThreatMultiplier: 0,
	})

	spellConfig := core.SpellConfig{
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolShadow,
		SpellCode:   SpellCode_WarlockDrainLife,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       core.SpellFlagAPL | core.SpellFlagResetAttackSwing | WarlockFlagAffliction | core.SpellFlagChanneled,

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
				Label: "DrainLife-" + warlock.Label + strconv.Itoa(rank),
			},
			NumberOfTicks:    numTicks,
			TickLength:       1 * time.Second,
			BonusCoefficient: spellCoeff,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, baseDamage, isRollover)
				dot.SnapshotAttackerMultiplier *= warlock.soulSiphonMultiplier(target)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				result := dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
				health := result.Damage
				healingSpell.CalcAndDealHealing(sim, healingSpell.Unit, health, healingSpell.OutcomeHealing)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
			if result.Landed() {

				dot := spell.Dot(target)
				dot.Apply(sim)
			}
			spell.DealOutcome(sim, result)
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

	return spellConfig
}

func (warlock *Warlock) registerDrainLifeSpell() {
	warlock.DrainLife = make([]*core.Spell, 0)
	for rank := 1; rank < len(drainLifeRanks); rank++ {
		config := warlock.getDrainLifeBaseConfig(rank)

		if config.RequiredLevel <= int(warlock.Level) {
			warlock.DrainLife = append(warlock.DrainLife, warlock.GetOrRegisterSpell(config))
		}
	}
}
