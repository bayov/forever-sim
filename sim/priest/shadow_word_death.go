package priest

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

const ShadowWordDeathRanks = 4

// Shadow Word: Death is new in Forever, trained from level 32 (foreverchanges.pro
// spellbook). Damage at each rank's cap, growth per level and coefficient from the
// client (1.60.1.70338, wago.tools). Instant, 15 sec cooldown. When the target lives,
// the priest takes 10% of their maximum health, which the sim leaves out.
var ShadowWordDeathSpellId = [ShadowWordDeathRanks + 1]int32{0, 1309595, 1309633, 1309635, 1309636}
var ShadowWordDeathBaseDamage = [ShadowWordDeathRanks + 1][]float64{{0}, {293.35, 311.65}, {368.02, 390.98}, {401.48, 426.52}, {444.15, 471.85}}
var ShadowWordDeathScaling = [ShadowWordDeathRanks + 1]core.RankScaling{{}, {37, 1.5}, {45, 1.9}, {53, 2.2}, {60, 2.5}}
var ShadowWordDeathManaCost = [ShadowWordDeathRanks + 1]float64{0, 175, 205, 250, 340}
var ShadowWordDeathLevel = [ShadowWordDeathRanks + 1]int{0, 32, 40, 48, 56}

func (priest *Priest) registerShadowWordDeathSpell() {
	priest.ShadowWordDeath = make([]*core.Spell, ShadowWordDeathRanks+1)
	cdTimer := priest.NewTimer()

	for rank := 1; rank <= ShadowWordDeathRanks; rank++ {
		if ShadowWordDeathLevel[rank] <= int(priest.Level) {
			priest.ShadowWordDeath[rank] = priest.GetOrRegisterSpell(priest.getShadowWordDeathConfig(rank, cdTimer))
		}
	}
}

func (priest *Priest) getShadowWordDeathConfig(rank int, cdTimer *core.Timer) core.SpellConfig {
	baseDamageLow := ShadowWordDeathScaling[rank].At(ShadowWordDeathBaseDamage[rank][0], priest.Level)
	baseDamageHigh := ShadowWordDeathScaling[rank].At(ShadowWordDeathBaseDamage[rank][1], priest.Level)

	// Early Demise: 15% crit a rank on targets at or below 20% health.
	earlyDemise := 15 * float64(priest.Talents.EarlyDemise) * core.SpellCritRatingPerCritChance

	return core.SpellConfig{
		SpellCode:   SpellCode_PriestShadowWordDeath,
		ActionID:    core.ActionID{SpellID: ShadowWordDeathSpellId[rank]},
		SpellSchool: core.SpellSchoolShadow,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagPriest | core.SpellFlagAPL,

		RequiredLevel: ShadowWordDeathLevel[rank],
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: ShadowWordDeathManaCost[rank],
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    cdTimer,
				Duration: time.Second * 15,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 0.429,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			bonusCrit := 0.0
			if earlyDemise > 0 && sim.IsExecutePhase20() {
				bonusCrit = earlyDemise
			}
			spell.BonusCritRating += bonusCrit
			result := spell.CalcDamage(sim, target, sim.Roll(baseDamageLow, baseDamageHigh), spell.OutcomeMagicHitAndCrit)
			spell.BonusCritRating -= bonusCrit

			if result.Landed() {
				priest.AddShadowWeavingStack(sim, target)
			}
			spell.DealDamage(sim, result)
		},
	}
}
