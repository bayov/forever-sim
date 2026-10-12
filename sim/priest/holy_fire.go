package priest

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core"
)

const HolyFireRanks = 8

var HolyFireSpellId = [HolyFireRanks + 1]int32{0, 14914, 15262, 15263, 15264, 15265, 15266, 15267, 15261}

// Damage at each rank's cap, growth per level and the tick damage from the Forever
// client (1.60.1.70338, wago.tools). Forever halved it (rank 8: 184 to 232 plus 75 over
// 10 sec, Classic 355 to 449 plus 145) and gave every rank 75% direct and 5% a tick.
var HolyFireBaseDamage = [HolyFireRanks + 1][]float64{{0}, {56.02, 70.38}, {63.47, 79.33}, {78.85, 99.15}, {89.12, 112.88}, {101.56, 128.44}, {131.59, 166.81}, {162.49, 206.31}, {183.68, 232.32}}
var HolyFireScaling = [HolyFireRanks + 1]core.RankScaling{{}, {24, 1.3}, {30, 1.4}, {36, 1.5}, {42, 1.5}, {48, 1.5}, {54, 1.7}, {60, 1.9}, {60, 2}}
var HolyFireTickDamage = [HolyFireRanks + 1]float64{0, 4, 5, 6, 7, 13, 10, 13, 15}
var HolyFireManaCost = [HolyFireRanks + 1]float64{0, 85, 95, 125, 145, 170, 200, 230, 255}
var HolyFireLevel = [HolyFireRanks + 1]int{0, 20, 24, 30, 36, 42, 48, 54, 60}

func (priest *Priest) registerHolyFire() {
	priest.HolyFire = make([]*core.Spell, HolyFireRanks+1)

	for rank := 1; rank <= HolyFireRanks; rank++ {
		config := priest.getHolyFireConfig(rank)

		if config.RequiredLevel <= int(priest.Level) {
			priest.HolyFire[rank] = priest.GetOrRegisterSpell(config)
		}
	}
}

func (priest *Priest) getHolyFireConfig(rank int) core.SpellConfig {
	ticks := int32(5)

	spellId := HolyFireSpellId[rank]
	baseDamageLow := HolyFireScaling[rank].At(HolyFireBaseDamage[rank][0], priest.Level)
	baseDamageHigh := HolyFireScaling[rank].At(HolyFireBaseDamage[rank][1], priest.Level)
	dotDamage := HolyFireTickDamage[rank]
	manaCost := HolyFireManaCost[rank]
	level := HolyFireLevel[rank]

	directCoeff := 0.75
	dotCoeff := 0.05
	castTime := time.Millisecond * 3500

	return core.SpellConfig{
		SpellCode:   SpellCode_PriestHolyFire,
		ActionID:    core.ActionID{SpellID: spellId},
		SpellSchool: core.SpellSchoolHoly,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagPriest | core.SpellFlagAPL,

		RequiredLevel: level,
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: castTime - time.Millisecond*100*time.Duration(priest.Talents.DivineFury),
			},
		},

		BonusCoefficient: directCoeff,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: fmt.Sprintf("Holy Fire (Rank %d)", rank),
			},

			NumberOfTicks:    ticks,
			TickLength:       time.Second * 2,
			BonusCoefficient: dotCoeff,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, dotDamage, isRollover)
			},

			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				// A waiting Inner Focus is only for the direct hit of the next spell.
				innerFocus := priest.InnerFocusAura != nil && priest.InnerFocusAura.IsActive()
				if innerFocus {
					dot.Spell.BonusCritRating -= 25 * core.SpellCritRatingPerCritChance
				}
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
				if innerFocus {
					dot.Spell.BonusCritRating += 25 * core.SpellCritRatingPerCritChance
				}
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			if result.Landed() {
				spell.Dot(target).Apply(sim)
			}
			spell.DealDamage(sim, result)
		},
	}
}
