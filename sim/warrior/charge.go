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
//
// In a leveling session it opens every pull it's ready for. It takes a mob 8 to 25 yards away while
// we're out of combat, pulls it, puts us on it and stuns it for 1 sec (the stun is spell 7922).
func (warrior *Warrior) registerChargeSpell() {
	spellIDs := []int32{0, 100, 6178, 11578}
	rank := warrior.knownRank([]int32{0, 4, 26, 46}, spellIDs)
	if rank == 0 {
		return
	}

	actionID := core.ActionID{SpellID: spellIDs[rank]}
	rageMetrics := warrior.NewRageMetrics(actionID)
	// Under Forever Charge's Rage makes no threat: a threat capture's Charge window shows the melee
	// hit's threat alone, and the client flags every rank and Vanguard's 1240289 with
	// ATTR_EX_1_NO_THREAT (wowsims/forever 6067c8469d).
	rageMetrics.NoThreat = warrior.Env.IsForever()
	rage := []float64{0, 9, 12, 15}[rank] + 3*float64(warrior.Talents.ImprovedCharge)

	stances := BattleStance
	if warrior.Talents.Vanguard {
		stances |= DefensiveStance
	}

	leveling := warrior.Env.Encounter.IsLeveling()
	warrior.Charge = warrior.RegisterSpell(stances, core.SpellConfig{
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolPhysical,
		Flags:       core.SpellFlagAPL,
		MaxRange:    core.TernaryFloat64(leveling, chargeMaxRange, 0),

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    warrior.NewTimer(),
				Duration: time.Second * 15,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			if leveling {
				return !sim.Encounter.LevelingInCombat(sim) &&
					sim.Encounter.LevelingCanCharge(sim, target, chargeMinRange, chargeMaxRange)
			}
			return sim.CurrentTime <= 0
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
			warrior.AddRage(sim, rage, rageMetrics)
			if leveling {
				sim.Encounter.LevelingCharge(sim, target, time.Second)
			}
		},
	})
}

// Intercept is a Charge that works in combat, in Berserker Stance. It costs 10 Rage, deals 25, 45 or
// 65 damage by rank and stuns for 3 sec (client 71091, 20252 and its 20253). A boss fight never needs
// it, so we only have it in a leveling session, for a mob still on its way to us or one we found.
func (warrior *Warrior) registerInterceptSpell() {
	if !warrior.Env.Encounter.IsLeveling() {
		return
	}
	spellIDs := []int32{0, 20252, 20616, 20617}
	rank := warrior.knownRank([]int32{0, 30, 42, 52}, spellIDs)
	if rank == 0 {
		return
	}
	damage := []float64{0, 25, 45, 65}[rank]

	warrior.Intercept = warrior.RegisterSpell(BerserkerStance, core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellIDs[rank]},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
		MaxRange:    chargeMaxRange,

		RageCost: core.RageCostOptions{
			Cost: 10,
		},
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    warrior.NewTimer(),
				Duration: time.Second * time.Duration(30-5*warrior.Talents.ImprovedIntercept),
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return sim.Encounter.LevelingCanCharge(sim, target, chargeMinRange, chargeMaxRange)
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			sim.Encounter.LevelingCharge(sim, target, time.Second*3)
			spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMeleeSpecialHitAndCrit)
		},
	})
}
