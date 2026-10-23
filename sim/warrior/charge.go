package warrior

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Charge only matters here for its Rage, so it is an opener: it can't be used in combat (client
// row 11578, Attributes_0 0x10000000), which the sim reads as before the pull, time 0 included.
// It deals no damage, the 1 sec stun does nothing to a boss, and the sim leaves the run in out.
// Rank 3 gives 15 Rage (row 11578, energize 150 tenths) and Improved Charge 3 more a point (row
// 12285). Vanguard makes it usable in Defensive Stance (row 1310317 swaps in 1240289, stance mask
// Battle and Defensive). wowsims/forever's Arms and Fury APLs open with it (charge.go).
func (warrior *Warrior) registerChargeSpell() {
	if warrior.Level < 4 {
		return
	}

	rank := struct {
		spellID int32
		rage    float64
	}{100, 9}
	switch {
	case warrior.Level >= 46:
		rank.spellID, rank.rage = 11578, 15
	case warrior.Level >= 26:
		rank.spellID, rank.rage = 6178, 12
	}

	actionID := core.ActionID{SpellID: rank.spellID}
	rageMetrics := warrior.NewRageMetrics(actionID)
	rage := rank.rage + 3*float64(warrior.Talents.ImprovedCharge)

	stances := BattleStance
	if warrior.Talents.Vanguard {
		stances |= DefensiveStance
	}

	warrior.Charge = warrior.RegisterSpell(stances, core.SpellConfig{
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolPhysical,
		Flags:       core.SpellFlagAPL,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    warrior.NewTimer(),
				Duration: time.Second * 15,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, _ *core.Unit) bool {
			return sim.CurrentTime <= 0
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			warrior.AddRage(sim, rage, rageMetrics)
		},
	})
}
