package druid

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

type FerociousBiteRankInfo struct {
	id           int32
	level        int32
	dmgBase      float64
	dmgRange     float64
	dmgPerCombo  float64
	dmgPerEnergy float64
	// foreverOnly is a rank only Forever has.
	foreverOnly bool
}

var ferociousBiteRanks = []FerociousBiteRankInfo{
	// Forever's new rank 1 at 24 (1328618), from the launch realms' 2026-10-30 hotfix (wago.tools
	// hotfixes, region 91, build 70707): 14 base points and 24 a combo point (33 to 43 damage with 1
	// combo point, up to 129 to 139 with 5, its tooltip), and 0.6 damage a point of extra Energy
	// (effect 1, 60 / 100). The old ranks became ranks 2 to 6, their rows unchanged.
	{
		id:           1328618,
		level:        24,
		dmgBase:      9.0,
		dmgRange:     10.0,
		dmgPerCombo:  24.0,
		dmgPerEnergy: 0.6,
		foreverOnly:  true,
	},
	{
		id:           22568,
		level:        32,
		dmgBase:      14.0,
		dmgRange:     16.0,
		dmgPerCombo:  36.0,
		dmgPerEnergy: 1.0,
	},
	{
		id:           22827,
		level:        40,
		dmgBase:      20.0,
		dmgRange:     24.0,
		dmgPerCombo:  59.0,
		dmgPerEnergy: 1.5,
	},
	{
		id:           22828,
		level:        48,
		dmgBase:      30.0,
		dmgRange:     40.0,
		dmgPerCombo:  92.0,
		dmgPerEnergy: 2.0,
	},
	{
		id:           22829,
		level:        56,
		dmgBase:      45.0,
		dmgRange:     50.0,
		dmgPerCombo:  128.0,
		dmgPerEnergy: 2.5,
	},
	{
		id:           31018,
		level:        60,
		dmgBase:      52.0,
		dmgRange:     60.0,
		dmgPerCombo:  147.0,
		dmgPerEnergy: 2.7,
	},
}

func (druid *Druid) registerFerociousBiteSpell() {
	// Forever's ranks from level 32 match 1.12's (1.60.1.70338, wago.tools). The last comes from an
	// Ahn'Qiraj book, which Forever leaves out (Environment.IncludeAQRanks).
	maxRank := len(ferociousBiteRanks) - core.TernaryInt(druid.Env.IncludeAQRanks(), 1, 2)
	rank := -1
	for i := 0; i <= maxRank; i++ {
		if ferociousBiteRanks[i].foreverOnly && !druid.Env.IsForever() {
			continue
		}
		if druid.Level >= ferociousBiteRanks[i].level {
			rank = i
		}
	}
	if rank < 0 {
		return
	}
	config := druid.newFerociousBiteSpellConfig(ferociousBiteRanks[rank])
	druid.FerociousBite = druid.RegisterSpell(Cat, config)
}

func (druid *Druid) newFerociousBiteSpellConfig(rank FerociousBiteRankInfo) core.SpellConfig {
	// A Bite that misses, is dodged or parried spends its 35 Energy and keeps the extra Energy and the
	// combo points, as wowsims/forever has it since c91498c01 ("in game, a missed... Ferocious Bite costs
	// its base energy and keeps its extra energy"). That reverses their earlier c742250b22, on a beta
	// report in their PR #149 that "a failed Bite leaves zero energy", which the report no longer assumes.
	// The rows carry ATTR_EX_1_DISCOUNT_POWER_ON_MISS (0x8000000, 22568 to 31018), as Rip's do, but the
	// game gives no refund of the 35 for it on Bite.
	//
	// Predatory Discipline (1328613, a Feral row 4 talent since the launch realms' 2026-10-30
	// hotfix): "Your Ferocious Bite ability no longer converts extra Energy into damage, but still
	// deals extra damage based on your excess Energy." Its one effect is a dummy aura with base
	// points 20 that the server reads, and the tooltip names no number. The sim takes the most
	// literal reading, as wowsims/forever does: the Bite still deals its rank's damage a point of
	// extra Energy, and spends only its 35. The dummy's 20 is not used (the report's
	// druid_predatory_discipline assumption).
	keepsExtraEnergy := druid.Talents.PredatoryDiscipline

	// The November 2 development notes raise the base points and the attack power part by about
	// 25% (tuning_70753.go); the damage a point of extra Energy stays.
	scale := core.TernaryFloat64(druid.Env.IsForever(), nov2BleedScale, 1)

	return core.SpellConfig{
		SpellCode:   SpellCode_DruidFerociousBite,
		ActionID:    core.ActionID{SpellID: rank.id},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       SpellFlagOmen | core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		EnergyCost: core.EnergyCostOptions{
			Cost: 35,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return druid.ComboPoints() > 0
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			comboPoints := float64(druid.ComboPoints())
			attackPower := spell.MeleeAttackPower(target)
			excessEnergy := druid.CurrentEnergy()

			baseDamage := scale*(rank.dmgBase+rank.dmgRange*sim.RandomFloat("Ferocious Bite")+
				rank.dmgPerCombo*comboPoints+
				attackPower*0.03*comboPoints) +
				rank.dmgPerEnergy*excessEnergy

			rendAndTear := druid.RendAndTearMultiplier(target)
			spell.DamageMultiplier *= rendAndTear
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
			spell.DamageMultiplier /= rendAndTear

			if result.Landed() && !keepsExtraEnergy {
				druid.SpendEnergy(sim, excessEnergy, spell.EnergyMetrics())
			}
			if result.Landed() {
				druid.SpendComboPoints(sim, spell)
			}
		},
	}
}

func (druid *Druid) CurrentFerociousBiteCost() float64 {
	return druid.FerociousBite.Cost.GetCurrentCost()
}
