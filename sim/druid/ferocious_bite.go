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
}

var ferociousBiteRanks = []FerociousBiteRankInfo{
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
	// Forever's ranks match 1.12's (1.60.1.70338, wago.tools). Rank 5 comes from an Ahn'Qiraj
	// book, which Forever leaves out (Environment.IncludeAQRanks).
	maxRank := core.TernaryInt(druid.Env.IncludeAQRanks(), 4, 3)
	rank := -1
	for i := 0; i <= maxRank; i++ {
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
	// Under Forever a Bite that misses spends its 35 Energy and all the extra Energy and gives
	// nothing back, and the combo points stay (wowsims/forever c742250b22, on a beta report in
	// their PR #149 that "a failed Bite leaves zero energy"). The rows carry
	// ATTR_EX_1_DISCOUNT_POWER_ON_MISS (0x8000000, 22568 to 31018), as Rip's do, but the game
	// gives no refund for it on Bite.
	spendExtraOnMiss := druid.Env.IsForever()

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

			baseDamage := rank.dmgBase + rank.dmgRange*sim.RandomFloat("Ferocious Bite") +
				rank.dmgPerCombo*comboPoints +
				attackPower*0.03*comboPoints +
				rank.dmgPerEnergy*excessEnergy

			rendAndTear := druid.RendAndTearMultiplier(target)
			spell.DamageMultiplier *= rendAndTear
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
			spell.DamageMultiplier /= rendAndTear

			if result.Landed() || spendExtraOnMiss {
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
