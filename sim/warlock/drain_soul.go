package warlock

import (
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Forever's Drain Soul hits harder than 1.12's (85 over 15 sec at rank 1, 1.12 55) and stops
// when another spell is cast (the 2026-10-08 build), which a channel does in the sim anyway.
func (warlock *Warlock) getDrainSoulBaseConfig(rank int) core.SpellConfig {
	numTicks := int32(5)
	tickLength := time.Second * 3

	data := drainSoulRanks[rank]
	spellId := data.SpellID
	spellCoeff := data.Coef
	baseDamage := data.damageAt(warlock.Level)
	manaCost := data.Cost
	level := data.Level

	return core.SpellConfig{
		SpellCode:   SpellCode_WarlockDrainSoul,
		ActionID:    core.ActionID{SpellID: spellId},
		SpellSchool: core.SpellSchoolShadow,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       core.SpellFlagAPL | core.SpellFlagChanneled | core.SpellFlagResetAttackSwing | WarlockFlagAffliction,

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

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "DrainSoul-" + warlock.Label + strconv.Itoa(rank),
			},
			NumberOfTicks:    numTicks,
			TickLength:       tickLength,
			BonusCoefficient: spellCoeff,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, baseDamage, isRollover)
				dot.SnapshotAttackerMultiplier *= warlock.soulSiphonMultiplier(target)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
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
}

func (warlock *Warlock) registerDrainSoulSpell() {
	warlock.DrainSoul = make([]*core.Spell, 0)
	for rank := 1; rank < len(drainSoulRanks); rank++ {
		config := warlock.getDrainSoulBaseConfig(rank)

		if config.RequiredLevel <= int(warlock.Level) {
			warlock.DrainSoul = append(warlock.DrainSoul, warlock.GetOrRegisterSpell(config))
		}
	}
}
