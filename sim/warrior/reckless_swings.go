package warrior

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Reckless Swings, a new Fury talent: "15 second cooldown. Costs 25% of base health. Your next 4
// melee attacks generate 5 additional Rage. Lasts 8 seconds." (development notes post 9,
// 2026-11-02; client 70753 unchanged, it has no such spell yet, so the action id is a placeholder.)
// Assumptions, until a client row or a log shows them:
//   - The health cost, 25% of base health as Bloodrage takes its 20% (bloodrage.go), is taken from
//     the warrior's health: nothing for a DPS, damage taken for a tank's survival score.
//   - It is off the global cooldown, like Bloodrage, the other warrior ability paid for in health.
//   - The 4 attacks are landed melee swings, Heroic Strike and Cleave included (the melee auto
//     attack proc type, as Unbridled Wrath's), each giving a flat 5 Rage that a crit doesn't double.
func (warrior *Warrior) registerRecklessSwingsCD() {
	if !warrior.Talents.RecklessSwings {
		return
	}

	actionID := core.ActionID{SpellID: 1399001}
	rageMetrics := warrior.NewRageMetrics(actionID)
	healthCost := 0.25 * warrior.GetBaseStats()[stats.Health]

	warrior.RecklessSwingsAura = warrior.RegisterAura(core.Aura{
		Label:     "Reckless Swings",
		ActionID:  actionID,
		Duration:  time.Second * 8,
		MaxStacks: 4,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.SetStacks(sim, 4)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || !spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) {
				return
			}
			warrior.AddRage(sim, 5, rageMetrics)
			aura.RemoveStack(sim)
		},
	})

	warrior.RecklessSwings = warrior.RegisterSpell(AnyStance, core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagAPL,
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    warrior.NewTimer(),
				Duration: time.Second * 15,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			warrior.RecklessSwingsAura.Activate(sim)
			warrior.RemoveHealth(sim, healthCost)
		},
	})
}
