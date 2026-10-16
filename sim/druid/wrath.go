package druid

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

const WrathRanks = 8

var WrathSpellId = [WrathRanks + 1]int32{0, 5176, 5177, 5178, 5179, 5180, 6780, 8905, 9912}

// Damage at each rank's cap (or 60), its growth per level up to there, and the spell power
// coefficient, from the Forever client (1.60.1.70338, wago.tools SpellEffect and SpellLevels).
// Forever cut Wrath to about 40% of its 1.12 damage (rank 8 is 92 to 102 at 60, 1.12 has
// 248 to 277) and made it cheaper. The low ranks keep close to full coefficients.
var WrathBaseDamage = [WrathRanks + 1][]float64{{0}, {14.58, 17.02}, {21.11, 24.49}, {30.25, 35.75}, {37.61, 43.59}, {46.01, 52.39}, {55.79, 63.01}, {69.34, 77.46}, {91.57, 102.43}}
var WrathScaling = [WrathRanks + 1]core.RankScaling{{}, {5, 0.2}, {12, 0.3}, {20, 0.5}, {28, 0.6}, {36, 0.7}, {44, 0.9}, {52, 0.9}, {60, 1}}
var WrathSpellCoeff = [WrathRanks + 1]float64{0, 0.429, 0.486, 0.571, 0.571, 0.571, 0.571, 0.571, 0.571}
var WrathManaCost = [WrathRanks + 1]float64{0, 10, 20, 40, 50, 70, 80, 100, 120}
var WrathCastTime = [WrathRanks + 1]int{0, 1500, 1700, 2000, 2000, 2000, 2000, 2000, 2000}
var WrathLevel = [WrathRanks + 1]int{0, 1, 6, 14, 22, 30, 38, 46, 54}

// Report assumption druid_wrath8_tuning, user ruling 2026-10-15: rank 8's base damage x1.085,
// half the raise that would bring it to the level 30 rank's 0.42 of Era (beta build 70291
// retuned ranks 1 to 5 only, and the higher ranks keep 0.34 to 0.36).
const WrathRank8DamageScale = 1.085

func (druid *Druid) registerWrathSpell() {
	druid.Wrath = make([]*DruidSpell, WrathRanks+1)

	for rank := 1; rank <= WrathRanks; rank++ {
		config := druid.newWrathSpellConfig(rank)

		if config.RequiredLevel <= int(druid.Level) {
			druid.Wrath[rank] = druid.RegisterSpell(Humanoid|Moonkin, config)
		}
	}
}

// WrathBaseDamageAt is a rank's damage range at a level, before spell power.
func WrathBaseDamageAt(rank int, level int32) (float64, float64) {
	low := WrathScaling[rank].At(WrathBaseDamage[rank][0], level)
	high := WrathScaling[rank].At(WrathBaseDamage[rank][1], level)
	if rank == 8 {
		low *= WrathRank8DamageScale
		high *= WrathRank8DamageScale
	}
	return low, high
}

func (druid *Druid) newWrathSpellConfig(rank int) core.SpellConfig {
	spellId := WrathSpellId[rank]
	baseDamageLow, baseDamageHigh := WrathBaseDamageAt(rank, druid.Level)
	spellCoeff := WrathSpellCoeff[rank]
	manaCost := WrathManaCost[rank]
	castTime := WrathCastTime[rank]
	level := WrathLevel[rank]

	return core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellId},
		SpellCode:   SpellCode_DruidWrath,
		SpellSchool: core.SpellSchoolNature,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagOmen | core.SpellFlagAPL | core.SpellFlagResetAttackSwing,

		PushbackReduction: druid.naturesFocusChance(),

		RequiredLevel: level,
		Rank:          rank,
		MissileSpeed:  20,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond * time.Duration(castTime),
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: spellCoeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)

			// Nature's Grace procs when the cast finishes, not when the bolt lands.
			if result.DidCrit() && druid.NaturesGraceProcAura != nil {
				druid.NaturesGraceProcAura.Activate(sim)
			}

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	}
}
