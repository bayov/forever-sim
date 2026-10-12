package mage

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core"
)

const FlamestrikeRanks = 6

var FlamestrikeSpellId = [FlamestrikeRanks + 1]int32{0, 2120, 2121, 8422, 8423, 10215, 10216}

// Damage at each rank's cap and its growth per level up to there, from the Forever client
// (1.60.1.70338), and the DoT from its tooltips. Rank 6 is 381 to 466 plus 332 over 8 sec
// (Classic 340), and rank 1 has the full 15.7%.
var FlamestrikeBaseDamage = [FlamestrikeRanks + 1][]float64{{0}, {54.6, 71.4}, {99.52, 126.48}, {158.45, 197.55}, {225.81, 279.19}, {297.72, 367.28}, {381.12, 466.48}}
var FlamestrikeScaling = [FlamestrikeRanks + 1]core.RankScaling{{}, {21, 0.6}, {29, 0.8}, {37, 1}, {45, 1.3}, {53, 1.5}, {60, 1.7}}
var FlamestrikeDotDamage = [FlamestrikeRanks + 1]float64{0, 44, 84, 132, 188, 256, 332}
var FlamestrikeSpellCoeff = [FlamestrikeRanks + 1]float64{0, .157, .157, .157, .157, .157, .157}
var FlamestrikeDotCoeff = [FlamestrikeRanks + 1]float64{0, .017, .02, .02, .02, .02, .02}
var FlamestrikeManaCost = [FlamestrikeRanks + 1]float64{0, 195, 330, 490, 650, 815, 990}
var FlamestrikeLevel = [FlamestrikeRanks + 1]int{0, 16, 24, 32, 40, 48, 56}

func (mage *Mage) registerFlamestrikeSpell() {
	mage.Flamestrike = make([]*core.Spell, FlamestrikeRanks+1)

	for rank := 1; rank <= FlamestrikeRanks; rank++ {
		config := mage.newFlamestrikeSpellConfig(rank)

		if config.RequiredLevel <= int(mage.Level) {
			mage.Flamestrike[rank] = mage.GetOrRegisterSpell(config)
		}
	}
}

func (mage *Mage) newFlamestrikeSpellConfig(rank int) core.SpellConfig {
	numTicks := int32(4)
	tickLength := time.Second * 2

	spellId := FlamestrikeSpellId[rank]
	baseDamageLow := FlamestrikeScaling[rank].At(FlamestrikeBaseDamage[rank][0], mage.Level)
	baseDamageHigh := FlamestrikeScaling[rank].At(FlamestrikeBaseDamage[rank][1], mage.Level)
	baseDotDamage := FlamestrikeDotDamage[rank] / float64(numTicks)
	spellCoeff := FlamestrikeSpellCoeff[rank]
	dotCoeff := FlamestrikeDotCoeff[rank]
	manaCost := FlamestrikeManaCost[rank]
	level := FlamestrikeLevel[rank]

	castTime := time.Second * 3

	return core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellId},
		SpellSchool: core.SpellSchoolFire,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagMage | core.SpellFlagAPL,
		SpellCode:   SpellCode_MageFlamestrike,

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

		BonusCritRating: float64(5 * mage.Talents.ImprovedFlamestrike * core.CritRatingPerCritChance),

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: spellCoeff,

		Dot: core.DotConfig{
			IsAOE: true,
			Aura: core.Aura{
				Label: fmt.Sprintf("Flamestrike (Rank %d)", rank),
			},
			NumberOfTicks:    numTicks,
			TickLength:       tickLength,
			BonusCoefficient: dotCoeff,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, baseDotDamage, isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				for _, aoeTarget := range sim.Encounter.TargetUnits {
					dot.CalcAndDealPeriodicSnapshotDamage(sim, aoeTarget, dot.OutcomeTick)
				}
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			for _, aoeTarget := range sim.Encounter.TargetUnits {
				baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
				spell.CalcAndDealDamage(sim, aoeTarget, baseDamage, spell.OutcomeMagicCrit)
			}
			spell.AOEDot().Apply(sim)
		},
	}
}
