package priest

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

const ChastiseRanks = 5

// Chastise is a new Dwarf priest spell in Forever (SkillLineAbility race mask 4): an
// instant Holy hit with a 2 min cooldown that only works on Humanoids. Damage and the
// 14.3% coefficient are from the client (1.60.1.70338, wago.tools). The 2 sec root does
// nothing in the sim.
var ChastiseSpellId = [ChastiseRanks + 1]int32{0, 1277331, 1277332, 1277333, 1277334, 1277335}
var ChastiseBaseDamage = [ChastiseRanks + 1][]float64{{0}, {43.24, 48.76}, {87.42, 98.58}, {131.6, 148.4}, {201.16, 226.84}, {271.66, 306.34}}
var ChastiseManaCost = [ChastiseRanks + 1]float64{0, 50, 100, 135, 180, 225}
var ChastiseLevel = [ChastiseRanks + 1]int{0, 20, 30, 40, 50, 60}

func (priest *Priest) registerChastiseSpell() {
	if priest.Race != proto.Race_RaceDwarf {
		return
	}

	priest.Chastise = make([]*core.Spell, ChastiseRanks+1)
	cdTimer := priest.NewTimer()

	for rank := 1; rank <= ChastiseRanks; rank++ {
		if ChastiseLevel[rank] > int(priest.Level) {
			continue
		}
		baseDamageLow := ChastiseBaseDamage[rank][0]
		baseDamageHigh := ChastiseBaseDamage[rank][1]

		priest.Chastise[rank] = priest.RegisterSpell(core.SpellConfig{
			SpellCode:   SpellCode_PriestChastise,
			ActionID:    core.ActionID{SpellID: ChastiseSpellId[rank]},
			SpellSchool: core.SpellSchoolHoly,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskSpellDamage,
			Flags:       SpellFlagPriest | core.SpellFlagAPL,

			RequiredLevel: ChastiseLevel[rank],
			Rank:          rank,

			ManaCost: core.ManaCostOptions{
				FlatCost: ChastiseManaCost[rank],
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
				CD: core.Cooldown{
					Timer:    cdTimer,
					Duration: time.Minute * 2,
				},
			},
			ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
				return target.MobType == proto.MobType_MobTypeHumanoid
			},

			DamageMultiplier: 1,
			// "Very low threat", not given a number.
			ThreatMultiplier: 1,
			BonusCoefficient: 0.143,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealDamage(sim, target, sim.Roll(baseDamageLow, baseDamageHigh), spell.OutcomeMagicHitAndCrit)
			},
		})
	}
}
