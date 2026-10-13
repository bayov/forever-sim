package warlock

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Forever's Life Tap converts its points plus one health per point of Spirit, times Improved
// Life Tap (client 11689: ($m1 + $SPI) * (1 + $18182m1 / 100), 420 + 1 a level to 66), and no
// longer gains from spell power (1.12 gave it 80% of it).
func (warlock *Warlock) getLifeTapBaseConfig(rank int) core.SpellConfig {
	data := lifeTapRanks[rank]
	baseAmount := data.damageAt(warlock.Level)

	actionID := core.ActionID{SpellID: data.SpellID}

	manaMetrics := warlock.NewManaMetrics(actionID)
	for _, pet := range warlock.BasePets {
		pet.LifeTapManaMetrics = pet.NewManaMetrics(actionID)
	}

	impLifeTap := 1 + 0.1*float64(warlock.Talents.ImprovedLifeTap)
	// Demonic Energies hands the demon 50% or 100% of the mana.
	petShare := []float64{0, 0.5, 1}[warlock.Talents.DemonicEnergies]

	return core.SpellConfig{
		ActionID:      actionID,
		SpellSchool:   core.SpellSchoolShadow,
		SpellCode:     SpellCode_WarlockLifeTap,
		ProcMask:      core.ProcMaskEmpty,
		Flags:         core.SpellFlagAPL | core.SpellFlagResetAttackSwing | WarlockFlagAffliction,
		RequiredLevel: data.Level,
		Rank:          rank,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			health := (baseAmount + warlock.GetStat(stats.Spirit)) * impLifeTap
			restore := health * (1 + warlock.lifeTapBonus)

			if warlock.IsTanking() {
				warlock.RemoveHealth(sim, health)
			}
			warlock.AddMana(sim, restore, manaMetrics)

			if pet := warlock.ActivePet; petShare > 0 && pet != nil && pet.IsEnabled() {
				pet.AddMana(sim, restore*petShare, pet.LifeTapManaMetrics)
			}
		},
	}
}

func (warlock *Warlock) registerLifeTapSpell() {
	warlock.LifeTap = make([]*core.Spell, 0)
	for i := 1; i < len(lifeTapRanks); i++ {
		config := warlock.getLifeTapBaseConfig(i)

		if config.RequiredLevel <= int(warlock.Level) {
			warlock.LifeTap = append(warlock.LifeTap, warlock.GetOrRegisterSpell(config))
		}
	}
}
