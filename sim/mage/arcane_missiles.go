package mage

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core"
)

const ArcaneMissilesRanks = 8

var ArcaneMissilesSpellId = [ArcaneMissilesRanks + 1]int32{0, 5143, 5144, 5145, 8416, 8417, 10211, 10212, 25345}

// Damage per missile at each rank's cap (or at 60) and its growth per level up to there, from the
// missile spells in the Forever client (1.60.1.70338, 7268 to 25346). Every rank has 28.6% per
// missile, 143% for a full channel (Classic 24% per missile, less on ranks 1 and 2). Rank 8 is
// 209 a missile (Classic 230). Rank 8's missile spell is learned at 60, so a level 56 to 59 mage
// has rank 7.
var ArcaneMissilesBaseTickDamage = [ArcaneMissilesRanks + 1]float64{0, 24.8, 35.2, 48.6, 68.4, 97.8, 133.2, 174.6, 209}
var ArcaneMissilesScaling = [ArcaneMissilesRanks + 1]core.RankScaling{{}, {12, 0.2}, {20, 0.3}, {28, 0.4}, {36, 0.6}, {44, 0.7}, {52, 0.8}, {60, 0.9}, {}}
var ArcaneMissilesSpellCoeff = [ArcaneMissilesRanks + 1]float64{0, .286, .286, .286, .286, .286, .286, .286, .286}
var ArcaneMissilesCastTime = [ArcaneMissilesRanks + 1]int32{0, 3, 4, 5, 5, 5, 5, 5, 5}
var ArcaneMissilesManaCost = [ArcaneMissilesRanks + 1]float64{0, 85, 140, 235, 320, 410, 500, 595, 655}
var ArcaneMissilesLevel = [ArcaneMissilesRanks + 1]int{0, 8, 16, 24, 32, 40, 48, 56, 60}

func (mage *Mage) registerArcaneMissilesSpell() {
	mage.ArcaneMissiles = make([]*core.Spell, ArcaneMissilesRanks+1)
	mage.ArcaneMissilesTickSpell = make([]*core.Spell, ArcaneMissilesRanks+1)

	maxRank := core.TernaryInt(core.IncludeAQ, ArcaneMissilesRanks, ArcaneMissilesRanks-1)
	for rank := 1; rank <= maxRank; rank++ {
		config := mage.getArcaneMissilesSpellConfig(rank)

		if config.RequiredLevel <= int(mage.Level) {
			mage.ArcaneMissiles[rank] = mage.GetOrRegisterSpell(config)
		}
	}
}

func (mage *Mage) getArcaneMissilesSpellConfig(rank int) core.SpellConfig {
	spellId := ArcaneMissilesSpellId[rank]
	baseTickDamage := ArcaneMissilesScaling[rank].At(ArcaneMissilesBaseTickDamage[rank], mage.Level)
	castTime := ArcaneMissilesCastTime[rank]
	manaCost := ArcaneMissilesManaCost[rank]
	level := ArcaneMissilesLevel[rank]

	numTicks := castTime
	tickLength := time.Second

	tickSpell := mage.getArcaneMissilesTickSpell(rank)
	mage.ArcaneMissilesTickSpell[rank] = tickSpell

	return core.SpellConfig{
		SpellCode:   SpellCode_MageArcaneMissiles,
		ActionID:    core.ActionID{SpellID: spellId},
		SpellSchool: core.SpellSchoolArcane,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagMage | core.SpellFlagAPL | core.SpellFlagChanneled | core.SpellFlagNoMetrics,

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

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: fmt.Sprintf("ArcaneMissiles-%d-%d", +rank, numTicks),
				OnExpire: func(aura *core.Aura, sim *core.Simulation) {
					// TODO: This check is necessary to ensure the final tick occurs before
					// Arcane Blast stacks are dropped. To fix this, ticks need to reliably
					// occur before aura expirations.

					//TODO: Test interaction in classic code without aura
					dot := mage.ArcaneMissiles[rank].Dot(aura.Unit)
					if dot.TickCount < dot.NumberOfTicks {
						dot.TickCount++
						dot.TickOnce(sim)
					}
				},
			},
			NumberOfTicks: numTicks,
			TickLength:    tickLength,
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				tickSpell.Cast(sim, target)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			dot := spell.Dot(target)
			// Missile Barrage halves the channel: the same missiles every 0.5 sec.
			dot.TickLength = tickLength
			if mage.MissileBarrageAura != nil && mage.MissileBarrageAura.IsActive() {
				dot.TickLength = tickLength / 2
				mage.MissileBarrageAura.Deactivate(sim)
			}
			// The missiles keep the Arcane Blast bonus the channel started with. Arcane
			// Missiles is a damage spell other than Arcane Blast, so it ends the stacks.
			mage.arcaneMissilesBlastBonus = mage.arcaneBlastBonus()
			if mage.ArcaneBlastAura != nil {
				mage.ArcaneBlastAura.Deactivate(sim)
			}
			dot.Apply(sim)
		},
		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
			return tickSpell.CalcDamage(sim, target, baseTickDamage, spell.OutcomeExpectedMagicHitAndCrit)
		},
	}
}

func (mage *Mage) getArcaneMissilesTickSpell(rank int) *core.Spell {
	spellId := ArcaneMissilesSpellId[rank]
	baseTickDamage := ArcaneMissilesScaling[rank].At(ArcaneMissilesBaseTickDamage[rank], mage.Level)
	spellCoeff := ArcaneMissilesSpellCoeff[rank]

	return mage.RegisterSpell(core.SpellConfig{
		SpellCode:    SpellCode_MageArcaneMissilesTick,
		ActionID:     core.ActionID{SpellID: spellId}.WithTag(1),
		SpellSchool:  core.SpellSchoolArcane,
		DefenseType:  core.DefenseTypeMagic,
		ProcMask:     core.ProcMaskSpellDamage,
		Flags:        SpellFlagMage,
		MissileSpeed: 20,

		Rank: 1,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: spellCoeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			bonus := mage.arcaneMissilesBlastBonus
			spell.DamageMultiplierAdditive += bonus
			result := mage.calcMageDamage(sim, target, baseTickDamage, spell)
			spell.DamageMultiplierAdditive -= bonus

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	})
}
