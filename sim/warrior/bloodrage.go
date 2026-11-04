package warrior

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

func (warrior *Warrior) registerBloodrageCD() {
	actionID := core.ActionID{SpellID: 2687}
	rageMetrics := warrior.NewRageMetrics(actionID)

	// Improved Bloodrage scales all of the Rage the ability makes now, not just the instant hit.
	rageMultiplier := 1 + 0.25*float64(warrior.Talents.ImprovedBloodrage)
	instantRage := 10.0 * rageMultiplier
	ragePerSec := 1.0 * rageMultiplier

	// Under Forever it costs 20% of base health (client 70753, spell 2687: power type health,
	// PowerCostPct 20; wowsims/forever's bloodrage.go takes it off the base stats' health too). A
	// tank's survival score counts it as damage taken (healing needed), but not TMI or DTPS
	// (PayHealthCost).
	// To verify: base health here is our class base health at the level (emulator convention:
	// 1689 for an Orc warrior at 60). wowsims/forever 841a787806 takes it from a base that counts
	// base Stamina (2629), with no evidence cited; a level 60 Bloodrage in a combat log settles it.
	healthCost := 0.0
	if warrior.Env.IsForever() {
		healthCost = 0.2 * warrior.GetBaseStats()[stats.Health]
	}

	warrior.BloodrageAura = warrior.RegisterAura(core.Aura{
		Label:    "Bloodrage",
		ActionID: actionID,
		Duration: time.Second * 10,
	})

	warrior.Bloodrage = warrior.RegisterSpell(AnyStance, core.SpellConfig{
		ActionID: actionID,
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    warrior.NewTimer(),
				Duration: time.Minute,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			warrior.BloodrageAura.Activate(sim)
			warrior.AddRage(sim, instantRage, rageMetrics)
			// Only a tank's: the sim heals no DPS, so the cost would only run a DPS warrior's health
			// down to 0, where its survival cooldowns (Berserker Rage) fire.
			if healthCost > 0 && warrior.IsTanking() {
				warrior.PayHealthCost(sim, healthCost)
			}

			core.StartPeriodicAction(sim, core.PeriodicActionOptions{
				NumTicks: 10,
				Period:   time.Second * 1,
				OnAction: func(sim *core.Simulation) {
					warrior.AddRage(sim, ragePerSec, rageMetrics)
				},
			})
		},
	})

	warrior.AddMajorCooldown(core.MajorCooldown{
		Spell: warrior.Bloodrage.Spell,
		Type:  core.CooldownTypeDPS,
	})
}
