package warlock

import (
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Rain of Fire is an 8 sec channel whose periodic dummy fires a wave every 2 sec at every enemy
// in the area, each wave a hit that rolls hit and crit on each target (wowsims/forever
// f1533984eb, rain_of_fire.go). The wave rows carry Rain of Fire's class flags, so Ruin,
// Cataclysm, Agonizing Flames and Intensity reach it.
func (warlock *Warlock) getRainOfFireBaseConfig(rank int) core.SpellConfig {
	data := rainOfFireRanks[rank]
	waveDamage := data.damageAt(warlock.Level)

	flags := core.SpellFlagAPL | core.SpellFlagResetAttackSwing | WarlockFlagDestruction | core.SpellFlagChanneled

	return core.SpellConfig{
		SpellCode:     SpellCode_WarlockRainOfFire,
		ActionID:      core.ActionID{SpellID: data.SpellID},
		SpellSchool:   core.SpellSchoolFire,
		DefenseType:   core.DefenseTypeMagic,
		ProcMask:      core.ProcMaskSpellDamage,
		Flags:         flags,
		RequiredLevel: data.Level,
		Rank:          rank,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: data.Coef,

		ManaCost: core.ManaCostOptions{
			FlatCost: data.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},
		Dot: core.DotConfig{
			IsAOE: true,
			Aura: core.Aura{
				Label: "RainOfFire-" + warlock.Label + strconv.Itoa(rank),
			},
			NumberOfTicks: 4,
			TickLength:    time.Second * 2,

			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
					dot.Spell.CalcAndDealDamage(sim, aoeTarget, waveDamage, dot.Spell.OutcomeMagicHitAndCrit)
				}
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.AOEDot().Apply(sim)
		},
	}
}

func (warlock *Warlock) registerRainOfFireSpell() {
	warlock.RainOfFire = make([]*core.Spell, 0)
	for rank := 1; rank < len(rainOfFireRanks); rank++ {
		config := warlock.getRainOfFireBaseConfig(rank)

		if config.RequiredLevel <= int(warlock.Level) {
			warlock.RainOfFire = append(warlock.RainOfFire, warlock.GetOrRegisterSpell(config))
		}
	}
}
