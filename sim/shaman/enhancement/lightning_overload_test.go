package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestOrcShamanLightningOverload checks the copies Lightning Overload casts (shaman_audit.md
// 8.13).
//
// With 3/3 each Lightning Bolt has a 10% chance to cast a second one of the same rank on the
// same target. The copy is free and deals half damage with no threat, and it never makes a
// copy of its own. We cast 2000 rank 10 bolts at level 60 and count the copies, then cast the
// copy itself 2000 times and want no more casts than that.
func TestOrcShamanLightningOverload(t *testing.T) {
	const casts = 2000
	// Lightning Overload 3.
	sim, enh := newShamanSim(newOrcShaman(60, "0000000000003", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 2*casts*3+10)
	target := enh.CurrentTarget
	lb := topRank(t, "Lightning Bolt", enh.LightningBolt)
	overload := enh.LightningBoltOverload[lb.Rank]
	if overload == nil {
		t.Fatalf("no overload for Lightning Bolt rank %d", lb.Rank)
	}

	at(sim, 0.5, func(sim *core.Simulation) {
		if got := overload.ThreatMultiplier; got != 0 {
			t.Errorf("the overload's threat multiplier is %.2f, want 0", got)
		}
		bolt := lb.CalcDamage(sim, target, 100, forcedOutcome(core.OutcomeHit)).Damage
		overloadHit := overload.CalcDamage(sim, target, 100, forcedOutcome(core.OutcomeHit)).Damage
		if math.Abs(overloadHit-bolt/2) > 1e-9 {
			t.Errorf("a 100 damage overload hit deals %.4f, want half of the bolt's %.4f", overloadHit, bolt)
		}
		mana := enh.CurrentMana()
		overload.Cast(sim, target)
		if got := enh.CurrentMana(); got != mana {
			t.Errorf("the overload cost %.2f mana, want 0", mana-got)
		}
	})

	for i := range casts {
		at(sim, float64(1+3*i), func(sim *core.Simulation) {
			enh.AddMana(sim, enh.MaxMana(), enh.GetManaNotCastingMetrics())
			castNow(t, sim, enh, lb)
		})
	}
	var copies int32
	at(sim, float64(1+3*casts), func(sim *core.Simulation) {
		copies = overload.SpellMetrics[target.UnitIndex].Casts - 1
		if rate := float64(copies) / casts; math.Abs(rate-.10) > .02 {
			t.Errorf("%d of %d bolts overloaded (%.1f%%), want 10%%", copies, casts, rate*100)
		}
	})

	// The copy cast on its own makes no copy.
	for i := range casts {
		at(sim, float64(2+3*(casts+i)), func(sim *core.Simulation) {
			overload.Cast(sim, target)
		})
	}
	runSim(sim)
	if got := overload.SpellMetrics[target.UnitIndex].Casts; got != 1+copies+casts {
		t.Errorf("casting the overload %d times made %d more", casts, got-1-copies-casts)
	}
}
