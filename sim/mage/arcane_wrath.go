package mage

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

const ArcaneWrathRanks = 4

var ArcaneWrathSpellId = [ArcaneWrathRanks + 1]int32{0, 1326079, 1326080, 1326081, 1326082}

// Arcane Wrath is a talent from the Forever client 1.60.1.70425 (Arcane, row 4, behind Arcane
// Blast), with trainer ranks behind it. Damage at each rank's cap (or at 60) and its growth per
// level up to there are the client's: base points plus 20.1% variance, so rank 4 is 406 to 496 at
// 60, at a full coefficient. 5% of the mage's maximum mana counts as extra spell damage (the
// client's second effect), so it goes through the same coefficient. 3.5 sec cast, 6 sec cooldown.
var ArcaneWrathBaseDamage = [ArcaneWrathRanks + 1][]float64{{0}, {120.53, 147.47}, {190.96, 233.64}, {293.6, 359.2}, {405.67, 496.33}}
var ArcaneWrathScaling = [ArcaneWrathRanks + 1]core.RankScaling{{}, {35, 1.4}, {47, 2.3}, {59, 2.4}, {}}
var ArcaneWrathManaCost = [ArcaneWrathRanks + 1]float64{0, 140, 220, 305, 395}
var ArcaneWrathLevel = [ArcaneWrathRanks + 1]int{0, 25, 36, 48, 60}

const ArcaneWrathSpellCoeff = 1.0
const ArcaneWrathMaxManaSpellDamage = 0.05

func (mage *Mage) registerArcaneWrathSpell() {
	if !mage.Talents.ArcaneWrath {
		return
	}

	// The ranks share the cooldown.
	cdTimer := mage.NewTimer()

	mage.ArcaneWrath = make([]*core.Spell, ArcaneWrathRanks+1)
	for rank := 1; rank <= ArcaneWrathRanks; rank++ {
		if ArcaneWrathLevel[rank] <= int(mage.Level) {
			mage.ArcaneWrath[rank] = mage.RegisterSpell(mage.newArcaneWrathSpellConfig(rank, cdTimer))
		}
	}
}

func (mage *Mage) newArcaneWrathSpellConfig(rank int, cdTimer *core.Timer) core.SpellConfig {
	baseDamageLow := ArcaneWrathScaling[rank].At(ArcaneWrathBaseDamage[rank][0], mage.Level)
	baseDamageHigh := ArcaneWrathScaling[rank].At(ArcaneWrathBaseDamage[rank][1], mage.Level)

	return core.SpellConfig{
		ActionID:     core.ActionID{SpellID: ArcaneWrathSpellId[rank]},
		SpellCode:    SpellCode_MageArcaneWrath,
		SpellSchool:  core.SpellSchoolArcane,
		DefenseType:  core.DefenseTypeMagic,
		ProcMask:     core.ProcMaskSpellDamage,
		Flags:        SpellFlagMage | core.SpellFlagAPL,
		MissileSpeed: 24,

		RequiredLevel: ArcaneWrathLevel[rank],
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: ArcaneWrathManaCost[rank],
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond * 3500,
			},
			CD: core.Cooldown{
				Timer:    cdTimer,
				Duration: time.Second * 6,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: ArcaneWrathSpellCoeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(baseDamageLow, baseDamageHigh) + ArcaneWrathSpellCoeff*ArcaneWrathMaxManaSpellDamage*mage.MaxMana()
			result := mage.calcMageDamage(sim, target, baseDamage, spell)
			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	}
}
