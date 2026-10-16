package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

const LavaBurstFlameShockBonus = .2

// Lava Burst ranks from the Forever client (1.60.1.70425, wago.tools): damage at each rank's
// cap, its growth per level up to there, and mana. The talent (at the 20-point gate since
// 70425) teaches rank 1 at 30 and the trainer the others at 40, 50 and 60. 70425 moved the
// three old ranks ten levels down (rank 1 now 123 + 0.8 a level, rank 2 164 + 0.9, rank 3
// 196 + 1.1, each cheaper) and added rank 4 (1326047) with what was rank 3: 220 + 1.3 a level,
// 265 mana. At 60 the shaman casts rank 4, the same spell as before under a new ID.
var LavaBurstBaseDamage = [][]float64{{0}, {113.01, 145.79}, {149.52, 192.88}, {178.86, 230.74}, {192.14, 247.86}}
var LavaBurstScaling = []core.RankScaling{{}, {38, 0.8}, {48, 0.9}, {58, 1.1}, {60, 1.3}}

// The spell IDs are Forever's, and the talent teaches rank 1.
var LavaBurstSpellId = []int32{0, 408490, 1238299, 1238300, 1326047}
var LavaBurstManaCost = []float64{0, 135, 165, 230, 265}
var LavaBurstLevel = []int32{0, 30, 40, 50, 60}

func (shaman *Shaman) registerLavaBurstSpell() {
	if !shaman.Talents.LavaBurst {
		return
	}

	rank := max(core.HighestRankAt(shaman.Level, LavaBurstLevel), 1)
	baseDamageLow := LavaBurstScaling[rank].At(LavaBurstBaseDamage[rank][0], shaman.Level)
	baseDamageHigh := LavaBurstScaling[rank].At(LavaBurstBaseDamage[rank][1], shaman.Level)
	spellCoeff := 0.714
	castTime := time.Millisecond * 2500

	shaman.LavaBurst = shaman.RegisterSpell(core.SpellConfig{
		SpellCode:   SpellCode_ShamanLavaBurst,
		ActionID:    core.ActionID{SpellID: LavaBurstSpellId[rank]},
		SpellSchool: core.SpellSchoolFire,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagShaman | core.SpellFlagAPL,
		Rank:        rank,

		MissileSpeed: 20,
		MaxRange:     30,

		ManaCost: core.ManaCostOptions{
			FlatCost: LavaBurstManaCost[rank],
			Factor:   shaman.convectionFactor(),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				CastTime: castTime - shaman.elementalAlacrityReduction(),
				GCD:      core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    shaman.NewTimer(),
				Duration: time.Second * 10,
			},
			ModifyCast: shaman.stopMeleeForCast,
		},

		DamageMultiplier:  1,
		ThreatMultiplier:  1,
		BonusCoefficient:  spellCoeff,
		PushbackReduction: shaman.eyeOfTheStormChance(),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			if shaman.hasActiveFlameShock(target) {
				spell.DamageMultiplier *= 1 + LavaBurstFlameShockBonus
				defer func() { spell.DamageMultiplier /= 1 + LavaBurstFlameShockBonus }()
			}

			baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	})
}

func (shaman *Shaman) hasActiveFlameShock(target *core.Unit) bool {
	for _, spell := range shaman.FlameShock {
		if spell != nil && spell.Dot(target).IsActive() {
			return true
		}
	}
	return false
}
