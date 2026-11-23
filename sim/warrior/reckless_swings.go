package warrior

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
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
//   - 25% of base health (SpellPower 315180), taken from the warrior's health: nothing for a DPS,
//     damage taken for a tank's survival score (healing needed), not TMI or DTPS (PayHealthCost).
//     Base health is Bloodrage's, to verify the same way (see bloodrage.go).
//
// Server hotfix push 114097 (2026-11-17) sets the buff's ProcTypeMask from 4 to 20, melee auto
// attacks and melee abilities (SpellAuraOptions 249144). Of the abilities only Heroic Strike and Cleave,
// which take the swing's place, use a charge, once a cast, landed or not, however many targets a
// Cleave hits, and each charge gives its 5 Rage. wowsims/forever's EV-0138 (2026-11-16, build 70938
// with hotfixes to push 114115, a level 60 Fury warrior): Cleave 19 casts and Heroic Strike 3 at 1
// a cast, on 1 or 2 targets; Whirlwind (12 casts), Bloodthirst (11) and Hamstring (6) none. A white
// swing uses one only when it lands (report assumption warr_rs_landed, measured in the same log:
// 106 of 106 landed swings, 0 of 22 missed), Windfury Totem's swing among them (14 charges).
func (warrior *Warrior) registerRecklessSwingsCD() {
	if !warrior.Talents.RecklessSwings {
		return
	}

	actionID := core.ActionID{SpellID: 1329555}
	rageMetrics := warrior.NewRageMetrics(core.ActionID{SpellID: 1329554})
	healthCost := 0.25 * warrior.GetBaseStats()[stats.Health]

	queuedSpentAt := time.Duration(-1)
	warrior.RecklessSwingsAura = warrior.RegisterAura(core.Aura{
		Label:     "Reckless Swings",
		ActionID:  actionID,
		Duration:  time.Second * 8,
		MaxStacks: 4,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.SetStacks(sim, 4)
			queuedSpentAt = -1
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) {
				return
			}
			if warrior.isQueuedSwing(spell) {
				// A Cleave's hits all land at its cast.
				if queuedSpentAt == sim.CurrentTime {
					return
				}
				queuedSpentAt = sim.CurrentTime
			} else if !result.Landed() {
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
			if warrior.IsTanking() { // as Bloodrage's: the sim heals no DPS
				warrior.PayHealthCost(sim, healthCost)
			}
		},
	})
}
