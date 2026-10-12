package warlock

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Shadowburn is a talent with trainer ranks behind it under Forever (client: learned at 20,
// 24, 32, 40, 48 and 56). It costs a Soul Shard, which the sim does not track.
func (warlock *Warlock) registerShadowBurnBaseConfig(rank int) core.SpellConfig {
	data := shadowburnRanks[rank]
	damageLow, damageHigh := data.rangeAt(warlock.Level)

	return core.SpellConfig{
		ActionID:      core.ActionID{SpellID: data.SpellID},
		SpellCode:     SpellCode_WarlockShadowburn,
		SpellSchool:   core.SpellSchoolShadow,
		DefenseType:   core.DefenseTypeMagic,
		ProcMask:      core.ProcMaskSpellDamage,
		Flags:         core.SpellFlagAPL | core.SpellFlagResetAttackSwing | core.SpellFlagBinary | WarlockFlagDestruction,
		RequiredLevel: data.Level,
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: data.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    warlock.GetOrInitTimer(&warlock.shadowburnCD),
				Duration: time.Second * 15,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: data.Coef,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, sim.Roll(damageLow, damageHigh), spell.OutcomeMagicHitAndCrit)
		},
	}
}

func (warlock *Warlock) registerShadowBurnSpell() {
	if !warlock.Talents.Shadowburn {
		return
	}

	warlock.Shadowburn = make([]*core.Spell, 0)
	for rank := 1; rank < len(shadowburnRanks); rank++ {
		config := warlock.registerShadowBurnBaseConfig(rank)

		if config.RequiredLevel <= int(warlock.Level) {
			warlock.Shadowburn = append(warlock.Shadowburn, warlock.GetOrRegisterSpell(config))
		}
	}
}
