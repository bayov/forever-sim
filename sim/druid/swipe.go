package druid

import (
	"github.com/wowsims/classic/sim/core"
)

const SwipeRanks = 5

var SwipeSpellId = [SwipeRanks + 1]int32{0, 779, 780, 769, 9754, 9908}
var SwipeBaseDamage = [SwipeRanks + 1]float64{0, 18, 25, 36, 60, 83}
var SwipeLevel = [SwipeRanks + 1]int32{0, 16, 24, 34, 44, 54}

// swipeAPCoefficient is the 3% of attack power wowsims/forever gives Swipe on top of its flat
// damage (an override of theirs: the client row states none).
const swipeAPCoefficient = 0.03

// registerSwipeBearSpell is the bear's Swipe under Forever (rank 5, 9908: 83 to up to 3 enemies,
// 20 Rage, Ferocity 1 a point less, Idol of Brutality 2 less; Clearcasting makes it free). Feral Instinct is now 10% more Swipe damage a point (the
// client tree; it was threat). Savage Fury's 5% a point is in talents.go. No threat on top of the
// form's (bear_attacks.go): SoD's +101% Swipe threat (its tuning passive 436895) isn't Forever's.
func (druid *Druid) registerSwipeBearSpell() {
	rank := 0
	for r := 1; r <= SwipeRanks; r++ {
		if druid.Level >= SwipeLevel[r] {
			rank = r
		}
	}
	if rank == 0 {
		return
	}
	baseDamage := SwipeBaseDamage[rank]
	numHits := min(3, druid.Env.GetNumTargets())
	results := make([]*core.SpellResult, numHits)

	druid.SwipeBear = druid.RegisterSpell(Bear, core.SpellConfig{
		ActionID:    core.ActionID{SpellID: SwipeSpellId[rank]},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL | SpellFlagOmen,

		Rank:          rank,
		RequiredLevel: int(SwipeLevel[rank]),

		RageCost: core.RageCostOptions{
			Cost: 20 - float64(druid.Talents.Ferocity) - druid.brutalityRage(),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{GCD: core.GCDDefault},
			IgnoreHaste: true,
		},

		DamageMultiplier: 1 + 0.1*float64(druid.Talents.FeralInstinct),
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// Only living targets, each one once.
			hits := results[:sim.Encounter.AoETargetCount(int32(len(results)))]
			for idx := range hits {
				damage := (baseDamage + swipeAPCoefficient*spell.MeleeAttackPower(target)) * druid.RendAndTearMultiplier(target)
				results[idx] = spell.CalcDamage(sim, target, damage, spell.OutcomeMeleeSpecialHitAndCrit)
				target = sim.Environment.NextTargetUnit(target)
			}
			for _, result := range hits {
				spell.DealDamage(sim, result)
			}
		},
	})
}
