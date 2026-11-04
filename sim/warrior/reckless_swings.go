package warrior

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Reckless Swings, a new Fury talent: "15 second cooldown. Costs 25% of base health. Your next 4
// melee attacks generate 5 additional Rage. Lasts 8 seconds." (development notes post 9,
// 2026-11-02). Hotfix push 113539 (2026-11-03) brings the spells: 1329555 is the cast and the buff
// (talent definition 145966), 1329554 the Rage. From its rows:
//   - 15 sec cooldown and a 1 sec global cooldown (SpellCooldowns 102442: StartRecoveryTime 1000,
//     category 133). The sim had it off the GCD before the hotfix.
//   - 8 sec, 4 charges, proc type 4 (the melee auto attack, as Flurry's and Unbridled Wrath's), so
//     Heroic Strike and Cleave use a charge as they do for those (SpellAuraOptions 249144).
//   - Each charge is a flat 5 Rage (1329554 effect 0, energize 50 tenths), which a crit doesn't double.
//   - 25% of base health (SpellPower 315180). Nothing for a DPS, so it isn't modelled.
//
// Still an assumption: only landed swings use a charge (the hit mask is server side, not in the data).
func (warrior *Warrior) registerRecklessSwingsCD() {
	if !warrior.Talents.RecklessSwings {
		return
	}

	actionID := core.ActionID{SpellID: 1329555}
	rageMetrics := warrior.NewRageMetrics(core.ActionID{SpellID: 1329554})

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
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			CD: core.Cooldown{
				Timer:    warrior.NewTimer(),
				Duration: time.Second * 15,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			warrior.RecklessSwingsAura.Activate(sim)
		},
	})
}
