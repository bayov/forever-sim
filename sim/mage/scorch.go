package mage

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

const ScorchRanks = 7

var ScorchSpellId = [ScorchRanks + 1]int32{0, 2948, 8444, 8445, 8446, 10205, 10206, 10207}

// Damage at each rank's cap (or at 60) and its growth per level up to there, from the Forever
// client (1.60.1.70338). Rank 7 is 166 to 196 at 60 (Classic 237 to 280).
var ScorchBaseDamage = [ScorchRanks + 1][]float64{{0}, {37.91, 46.49}, {53.45, 64.55}, {66.36, 79.64}, {89.09, 106.51}, {111.29, 131.91}, {142.11, 169.89}, {166.4, 196.4}}
var ScorchScaling = [ScorchRanks + 1]core.RankScaling{{}, {26, 0.8}, {32, 1}, {38, 1}, {44, 1.2}, {50, 1.4}, {56, 1.5}, {60, 1.7}}
var ScorchManaCost = [ScorchRanks + 1]float64{0, 50, 65, 80, 100, 115, 135, 150}
var ScorchLevel = [ScorchRanks + 1]int{0, 22, 28, 34, 40, 46, 52, 58}

func (mage *Mage) registerScorchSpell() {
	mage.Scorch = make([]*core.Spell, ScorchRanks+1)

	for rank := 1; rank <= ScorchRanks; rank++ {
		config := mage.getScorchConfig(rank)

		if config.RequiredLevel <= int(mage.Level) {
			mage.Scorch[rank] = mage.GetOrRegisterSpell(config)
		}
	}
}

func (mage *Mage) getScorchConfig(rank int) core.SpellConfig {
	spellId := ScorchSpellId[rank]
	baseDamageLow := ScorchScaling[rank].At(ScorchBaseDamage[rank][0], mage.Level)
	baseDamageHigh := ScorchScaling[rank].At(ScorchBaseDamage[rank][1], mage.Level)
	manaCost := ScorchManaCost[rank]
	level := ScorchLevel[rank]

	spellCoeff := .429
	debuffProcChance := []float64{0, .33, .67, 1}[mage.Talents.ImprovedScorch]

	return core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellId},
		SpellCode:   SpellCode_MageScorch,
		SpellSchool: core.SpellSchoolFire,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       core.SpellFlagAPL | SpellFlagMage,

		RequiredLevel: level,
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond * 1500,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: spellCoeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)

			// Since the 2026-09-24 build the vulnerability no longer rolls to be resisted on its
			// own, so it lands with every Scorch that hits.
			if result.Landed() && debuffProcChance > 0 && sim.RandomFloat("Improved Scorch") < debuffProcChance {
				aura := mage.ImprovedScorchAuras.Get(target)
				aura.Activate(sim)
				aura.AddStack(sim)
			}
		},

		RelatedAuras: []core.AuraArray{mage.ImprovedScorchAuras},
	}
}
