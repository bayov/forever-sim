package priest

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core"
)

const ShadowWordPainRanks = 8

// Damage per tick and the coefficient per tick from the Forever client (1.60.1.70338,
// wago.tools SpellEffect). Forever cut the damage (rank 8: 762 over 18 sec, Classic 852)
// and gave every rank 20% a tick, 120% over the 6 ticks (Classic 16.7% a tick at the top
// ranks).
var ShadowWordPainSpellId = [ShadowWordPainRanks + 1]int32{0, 589, 594, 970, 992, 2767, 10892, 10893, 10894}
var ShadowWordPainTickDamage = [ShadowWordPainRanks + 1]float64{0, 5, 10, 18, 30, 48, 71, 97, 127}
var ShadowWordPainManaCost = [ShadowWordPainRanks + 1]float64{0, 25, 50, 95, 155, 230, 305, 385, 470}
var ShadowWordPainLevel = [ShadowWordPainRanks + 1]int{0, 4, 10, 18, 26, 34, 42, 50, 58}

const ShadowWordPainTickCoef = 0.2

func (priest *Priest) registerShadowWordPainSpell() {
	priest.ShadowWordPain = make([]*core.Spell, ShadowWordPainRanks+1)

	for rank := 1; rank <= ShadowWordPainRanks; rank++ {
		config := priest.getShadowWordPainConfig(rank)

		if config.RequiredLevel <= int(priest.Level) {
			priest.ShadowWordPain[rank] = priest.GetOrRegisterSpell(config)
		}
	}
}

func (priest *Priest) getShadowWordPainConfig(rank int) core.SpellConfig {
	ticks := int32(6)

	spellId := ShadowWordPainSpellId[rank]
	baseDotDamage := ShadowWordPainTickDamage[rank]
	manaCost := ShadowWordPainManaCost[rank]
	level := ShadowWordPainLevel[rank]

	return core.SpellConfig{
		SpellCode:   SpellCode_PriestShadowWordPain,
		ActionID:    core.ActionID{SpellID: spellId},
		SpellSchool: core.SpellSchoolShadow,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagPriest | core.SpellFlagAPL | core.SpellFlagPureDot,

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
				Label: fmt.Sprintf("Shadow Word: Pain (Rank %d)", rank),
			},

			// Improved Shadow Word: Pain adds 3 sec (one tick) a rank.
			NumberOfTicks:    ticks + priest.Talents.ImprovedShadowWordPain,
			TickLength:       time.Second * 3,
			BonusCoefficient: ShadowWordPainTickCoef,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, baseDotDamage, isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)

			if result.Landed() {
				priest.AddShadowWeavingStack(sim, result.Target)
				spell.Dot(result.Target).Apply(sim)
			}
			spell.DealOutcome(sim, result)
		},

		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
			if useSnapshot {
				return spell.Dot(target).CalcSnapshotDamage(sim, target, spell.OutcomeExpectedMagicAlwaysHit)
			}
			return spell.CalcPeriodicDamage(sim, target, baseDotDamage, spell.OutcomeExpectedMagicAlwaysHit)
		},
	}
}
