package mage

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core"
)

const PyroblastRanks = 8

var PyroblastSpellId = [PyroblastRanks + 1]int32{0, 11366, 12505, 12522, 12523, 12524, 12525, 12526, 18809}

// Damage at each rank's cap (or at 60) and its growth per level up to there, from the Forever
// client (1.60.1.70338). Rank 8 is 520 to 646 plus 212 over 12 sec (Classic 716 to 890 plus 268).
var PyroblastBaseDamage = [PyroblastRanks + 1][]float64{{0}, {99.73, 132.27}, {124.79, 163.61}, {178.41, 228.79}, {228.19, 290.61}, {288.87, 365.53}, {365.5, 458.5}, {445.69, 557.11}, {519.84, 646.16}}
var PyroblastScaling = [PyroblastRanks + 1]core.RankScaling{{}, {24, 1.5}, {30, 1.7}, {36, 2.1}, {42, 2.4}, {48, 2.7}, {54, 3}, {60, 3.4}, {}}
var PyroblastDotDamage = [PyroblastRanks + 1]float64{0, 44, 56, 76, 100, 124, 152, 184, 212}
var PyroblastManaCost = [PyroblastRanks + 1]float64{0, 125, 150, 195, 240, 285, 335, 385, 440}
var PyroblastLevel = [PyroblastRanks + 1]int{0, 20, 24, 30, 36, 42, 48, 54, 60}

func (mage *Mage) registerPyroblastSpell() {
	if !mage.Talents.Pyroblast {
		return
	}

	mage.Pyroblast = make([]*core.Spell, PyroblastRanks+1)

	for rank := 1; rank <= PyroblastRanks; rank++ {
		config := mage.newPyroblastSpellConfig(rank)

		if config.RequiredLevel <= int(mage.Level) {
			mage.Pyroblast[rank] = mage.GetOrRegisterSpell(config)
		}
	}
}

func (mage *Mage) newPyroblastSpellConfig(rank int) core.SpellConfig {

	numTicks := int32(4)
	tickLength := time.Second * 3

	spellId := PyroblastSpellId[rank]
	baseDamageLow := PyroblastScaling[rank].At(PyroblastBaseDamage[rank][0], mage.Level)
	baseDamageHigh := PyroblastScaling[rank].At(PyroblastBaseDamage[rank][1], mage.Level)
	baseDotDamage := PyroblastDotDamage[rank] / float64(numTicks)
	manaCost := PyroblastManaCost[rank]
	level := PyroblastLevel[rank]

	spellCoeff := 1.0
	// 15% a tick, the client's coefficient on the periodic effect.
	dotCoeff := .15
	castTime := time.Second * 6

	actionID := core.ActionID{SpellID: spellId}

	spellConfig := core.SpellConfig{
		ActionID:     actionID,
		SpellCode:    SpellCode_MagePyroblast,
		SpellSchool:  core.SpellSchoolFire,
		DefenseType:  core.DefenseTypeMagic,
		ProcMask:     core.ProcMaskSpellDamage,
		Flags:        SpellFlagMage | core.SpellFlagAPL,
		MissileSpeed: 24,

		RequiredLevel: level,
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: castTime,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: spellCoeff,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label:    fmt.Sprintf("Pyroblast (Rank %d)", rank),
				ActionID: actionID.WithTag(1),
			},
			NumberOfTicks:    numTicks,
			TickLength:       tickLength,
			BonusCoefficient: dotCoeff,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, baseDotDamage, isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

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

	return spellConfig
}
