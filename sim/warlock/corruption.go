package warlock

import (
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Ranks 1 and 2 last 12 and 15 sec, the rest 18 (client SpellDuration), all with 3 sec ticks.
var corruptionTicks = []int32{0, 4, 5, 6, 6, 6, 6, 6}

func (warlock *Warlock) getCorruptionConfig(rank int) core.SpellConfig {
	data := corruptionRanks[rank]
	ticks := corruptionTicks[rank]
	baseDamage := data.damageAt(warlock.Level)

	return core.SpellConfig{
		ActionID:      core.ActionID{SpellID: data.SpellID},
		SpellSchool:   core.SpellSchoolShadow,
		SpellCode:     SpellCode_WarlockCorruption,
		ProcMask:      core.ProcMaskSpellDamage,
		DefenseType:   core.DefenseTypeMagic,
		Flags:         core.SpellFlagAPL | core.SpellFlagResetAttackSwing | core.SpellFlagPureDot | WarlockFlagAffliction,
		Rank:          rank,
		RequiredLevel: data.Level,

		ManaCost: core.ManaCostOptions{
			FlatCost: data.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				CastTime: data.castTime() - 400*time.Millisecond*time.Duration(warlock.Talents.ImprovedCorruption),
				GCD:      core.GCDDefault,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Corruption-" + warlock.Label + strconv.Itoa(rank),
			},

			NumberOfTicks:    ticks,
			TickLength:       time.Second * 3,
			BonusCoefficient: data.Coef,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, baseDamage, isRollover)
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
				return dot.CalcSnapshotDamage(sim, target, dot.Spell.OutcomeExpectedMagicAlwaysHit)
			} else {
				return spell.CalcPeriodicDamage(sim, target, baseDamage, spell.OutcomeExpectedMagicAlwaysHit)
			}
		},
	}
}

func (warlock *Warlock) registerCorruptionSpell() {
	warlock.Corruption = make([]*core.Spell, 0)

	maxRank := core.TernaryInt(core.IncludeAQ, len(corruptionRanks)-1, len(corruptionRanks)-2)
	for rank := 1; rank <= maxRank; rank++ {
		config := warlock.getCorruptionConfig(rank)

		if config.RequiredLevel <= int(warlock.Level) {
			warlock.Corruption = append(warlock.Corruption, warlock.GetOrRegisterSpell(config))
		}
	}
}
