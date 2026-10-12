package mage

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core"
)

const FireballRanks = 12

var FireballSpellId = [FireballRanks + 1]int32{0, 133, 143, 145, 3140, 8400, 8401, 8402, 10148, 10149, 10150, 10151, 25306}

// Damage at each rank's cap (or at 60), its growth per level up to there, and the spell power
// coefficient, from the Forever client (1.60.1.70338, wago.tools SpellEffect and SpellLevels).
// Forever cut the top ranks hard: rank 12 is 425 to 541 (Classic 596 to 760) with a 60 damage
// DoT (Classic 76), at the same full coefficient. The low ranks keep close to full coefficients
// instead of the 1.12 penalty for spells learned below level 20.
var FireballBaseDamage = [FireballRanks + 1][]float64{{0}, {14.62, 22.98}, {27.08, 39.32}, {44.92, 61.88}, {67.7, 93.5}, {107.11, 144.09}, {137.41, 182.99}, {169.6, 222.8}, {211.83, 275.77}, {270.3, 348.9}, {337.23, 432.37}, {396.57, 505.43}, {424.58, 541.42}}
var FireballScaling = [FireballRanks + 1]core.RankScaling{{}, {5, 0.2}, {10, 0.3}, {16, 0.6}, {22, 0.9}, {28, 1.4}, {34, 1.8}, {40, 2.3}, {46, 2.2}, {52, 2.4}, {58, 2.7}, {}, {}}
var FireballDotDamage = [FireballRanks + 1]float64{0, 2, 3, 6, 12, 16, 24, 24, 32, 40, 48, 56, 60}
var FireballSpellCoeff = [FireballRanks + 1]float64{0, .429, .571, .714, .857, 1, 1, 1, 1, 1, 1, 1, 1}
var FireballCastTime = [FireballRanks + 1]int32{0, 1500, 2000, 2500, 3000, 3500, 3500, 3500, 3500, 3500, 3500, 3500, 3500}
var FireballManaCost = [FireballRanks + 1]float64{0, 30, 45, 65, 95, 140, 185, 220, 260, 305, 350, 395, 410}
var FireballLevel = [FireballRanks + 1]int{0, 1, 6, 12, 18, 24, 30, 36, 42, 48, 54, 60, 60}

// The DoT ticks every 2 sec for the rank's duration (4 sec for rank 1, 6 sec for ranks 2 and 3, 8 sec after).
var FireballDotTicks = [FireballRanks + 1]int32{0, 2, 3, 3, 4, 4, 4, 4, 4, 4, 4, 4, 4}

func (mage *Mage) registerFireballSpell() {
	mage.Fireball = make([]*core.Spell, FireballRanks+1)

	maxRank := core.TernaryInt(core.IncludeAQ, FireballRanks, FireballRanks-1)
	for rank := 1; rank <= maxRank; rank++ {
		config := mage.newFireballSpellConfig(rank)

		if config.RequiredLevel <= int(mage.Level) {
			mage.Fireball[rank] = mage.GetOrRegisterSpell(config)
		}
	}
}

func (mage *Mage) newFireballSpellConfig(rank int) core.SpellConfig {
	numTicks := FireballDotTicks[rank]
	tickLength := time.Second * 2

	spellId := FireballSpellId[rank]
	baseDamageLow := FireballScaling[rank].At(FireballBaseDamage[rank][0], mage.Level)
	baseDamageHigh := FireballScaling[rank].At(FireballBaseDamage[rank][1], mage.Level)
	baseDotDamage := FireballDotDamage[rank] / float64(numTicks)
	spellCoeff := FireballSpellCoeff[rank]
	castTime := FireballCastTime[rank]
	manaCost := FireballManaCost[rank]
	level := FireballLevel[rank]

	actionID := core.ActionID{SpellID: spellId}

	return core.SpellConfig{
		ActionID:     actionID,
		SpellCode:    SpellCode_MageFireball,
		SpellSchool:  core.SpellSchoolFire,
		DefenseType:  core.DefenseTypeMagic,
		ProcMask:     core.ProcMaskSpellDamage,
		Flags:        core.SpellFlagAPL | SpellFlagMage,
		MissileSpeed: 24,

		RequiredLevel: level,
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond*time.Duration(castTime) - time.Millisecond*100*time.Duration(mage.Talents.ImprovedFireball),
			},
		},

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label:    fmt.Sprintf("Fireball (Rank %d)", rank),
				ActionID: actionID.WithTag(1),
			},
			NumberOfTicks: numTicks,
			TickLength:    tickLength,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, baseDotDamage, isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: spellCoeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
			result := mage.calcMageDamage(sim, target, baseDamage, spell)
			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)

				if result.Landed() {
					spell.Dot(target).Apply(sim)
				}
			})
		},
	}
}
