package warlock

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Soul Fire costs a Soul Shard, which the sim does not track (a raiding warlock brings
// enough). Decimation takes up to 90% off its 1 min cooldown and, after a Shadow Bolt or
// Searing Pain on a target under 35%, up to 40% off its cast.
func (warlock *Warlock) getSoulFireBaseConfig(rank int) core.SpellConfig {
	data := soulFireRanks[rank]
	damageLow, damageHigh := data.rangeAt(warlock.Level)

	return core.SpellConfig{
		SpellCode:     SpellCode_WarlockSoulFire,
		ActionID:      core.ActionID{SpellID: data.SpellID},
		SpellSchool:   core.SpellSchoolFire,
		DefenseType:   core.DefenseTypeMagic,
		ProcMask:      core.ProcMaskSpellDamage,
		Flags:         core.SpellFlagAPL | core.SpellFlagResetAttackSwing | WarlockFlagDestruction,
		RequiredLevel: data.Level,
		Rank:          rank,
		MissileSpeed:  24,

		ManaCost: core.ManaCostOptions{
			FlatCost: data.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: data.castTime(),
			},
			CD: core.Cooldown{
				Timer:    warlock.soulFireTimer(),
				Duration: time.Duration(float64(time.Minute) * (1 - 0.45*float64(warlock.Talents.Decimation))),
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: data.Coef,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcDamage(sim, target, sim.Roll(damageLow, damageHigh), spell.OutcomeMagicHitAndCrit)
			spell.WaitTravelTime(sim, func(s *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	}
}

// The ranks share one cooldown (category 631).
func (warlock *Warlock) soulFireTimer() *core.Timer {
	return warlock.GetOrInitTimer(&warlock.soulFireCD)
}

func (warlock *Warlock) registerSoulFireSpell() {
	warlock.SoulFire = make([]*core.Spell, 0)
	for rank := 1; rank < len(soulFireRanks); rank++ {
		config := warlock.getSoulFireBaseConfig(rank)

		if config.RequiredLevel <= int(warlock.Level) {
			warlock.SoulFire = append(warlock.SoulFire, warlock.GetOrRegisterSpell(config))
		}
	}
}
