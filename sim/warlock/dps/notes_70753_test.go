package dps

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The Warlock lines of Blizzard's development notes of 2026-11-02 (client 1.60.1.70753).

// TestWarlockWrack70753 checks Wrack at 60: 86 base damage a second (36 at 40, 2.5 more a level) and
// 0.286 of spell damage a tick, twice the client's 0.143. Spell damage is added after the first cast
// to read both off the snapshot.
func TestWarlockWrack70753(t *testing.T) {
	sim, lock := newWarlockSim(t, "00000000000000001", &proto.WarlockOptions{})
	target := lock.CurrentTarget
	snapshot := func() float64 {
		dot := lock.Wrack.Dot(target)
		if !dot.IsActive() {
			t.Fatalf("at %s: Wrack not up", sim.CurrentTime)
		}
		return dot.SnapshotBaseDamage - 0.286*lock.Wrack.GetBonusDamage(target)
	}

	at(sim, 0.5, func(sim *core.Simulation) {
		lock.AddStatDynamic(sim, stats.SpellHit, 20*core.SpellHitRatingPerHitChance)
		castNow(t, sim, lock, lock.Wrack)
	})
	at(sim, 1, func(sim *core.Simulation) {
		if got := snapshot(); math.Abs(got-86) > 1e-6 {
			t.Errorf("Wrack's base tick at 60: %.3f, want 86", got)
		}
	})
	at(sim, 8, func(sim *core.Simulation) {
		lock.AddStatDynamic(sim, stats.ShadowPower, 300)
		castNow(t, sim, lock, lock.Wrack)
	})
	at(sim, 8.5, func(sim *core.Simulation) {
		if got := snapshot(); math.Abs(got-86) > 1e-6 {
			t.Errorf("Wrack's tick at +300 Shadow damage less 0.286 of it: %.3f, want 86", got)
		}
	})
	runSim(sim)
}

// TestWarlockShadowAndFlame70753 checks Shadow and Flame 5/5: a Conflagrate that lands gives 20%
// Shadow damage (4% a point, was 2%).
func TestWarlockShadowAndFlame70753(t *testing.T) {
	sim, lock := newWarlockSim(t, "--000000000010005", &proto.WarlockOptions{})
	shadow := func() float64 { return lock.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexShadow] }
	var before float64

	at(sim, 0.5, func(sim *core.Simulation) {
		lock.AddStatDynamic(sim, stats.SpellHit, 20*core.SpellHitRatingPerHitChance)
		castNow(t, sim, lock, last(lock.Immolate))
	})
	at(sim, 3, func(sim *core.Simulation) {
		before = shadow()
		lock.SetGCDTimer(sim, sim.CurrentTime)
		if !last(lock.Conflagrate).Cast(sim, lock.CurrentTarget) {
			t.Fatalf("Conflagrate didn't cast")
		}
	})
	at(sim, 3.5, func(sim *core.Simulation) {
		if got := shadow() / before; math.Abs(got-1.20) > 1e-9 {
			t.Errorf("Shadow damage after Conflagrate with Shadow and Flame 5/5: x%.4f, want x1.20", got)
		}
	})
	runSim(sim)
}

// TestWarlockBaneOfAgonyLive checks that Bane of Agony's ticks read spell damage when they land
// ("now correctly dynamically updates its damage"): Shadow damage gained after the cast raises
// the next tick by the rank's coefficient.
func TestWarlockBaneOfAgonyLive(t *testing.T) {
	sim, lock := newWarlockSim(t, "", &proto.WarlockOptions{})
	target := lock.CurrentTarget
	agony := last(lock.BaneOfAgony)
	tick := func() float64 {
		return agony.Dot(target).CalcSnapshotDamage(sim, target, agony.OutcomeExpectedMagicAlwaysHit).Damage
	}
	var before float64

	at(sim, 0.5, func(sim *core.Simulation) {
		lock.AddStatDynamic(sim, stats.SpellHit, 20*core.SpellHitRatingPerHitChance)
		castNow(t, sim, lock, agony)
	})
	at(sim, 1, func(sim *core.Simulation) {
		before = tick()
		lock.AddStatDynamic(sim, stats.ShadowPower, 300)
	})
	at(sim, 1.5, func(sim *core.Simulation) {
		after := tick()
		if after <= before {
			t.Errorf("Bane of Agony's tick %.1f before +300 Shadow damage and %.1f after, want it to grow", before, after)
		}
		if got, want := after-before, 300*agony.Dot(target).BonusCoefficient*before/agony.Dot(target).SnapshotBaseDamage; math.Abs(got-want) > 1e-6 {
			t.Errorf("Bane of Agony's tick grew %.3f with +300 Shadow damage, want %.3f", got, want)
		}
	})
	runSim(sim)
}

// TestWarlockAmplifyCurseAgony checks that Amplify Curse raises Bane of Agony's base damage by 50%
// ("now correctly affects Curse of Exhaustion and Bane of Agony", the 2026-11-02 notes), not its
// spell damage part, and is spent by it.
func TestWarlockAmplifyCurseAgony(t *testing.T) {
	bases := map[bool]float64{}
	for _, amplify := range []bool{false, true} {
		sim, lock := newWarlockSim(t, "000000001", &proto.WarlockOptions{})
		target := lock.CurrentTarget
		agony := last(lock.BaneOfAgony)

		at(sim, 0.5, func(sim *core.Simulation) {
			lock.AddStatDynamic(sim, stats.SpellHit, 20*core.SpellHitRatingPerHitChance)
			if amplify {
				lock.AmplifyCurse.Cast(sim, target)
			}
			castNow(t, sim, lock, agony)
		})
		at(sim, 1, func(sim *core.Simulation) {
			dot := agony.Dot(target)
			bases[amplify] = dot.SnapshotBaseDamage - dot.BonusCoefficient*agony.GetBonusDamage(target)
			if lock.AmplifyCurseAura.IsActive() {
				t.Errorf("Amplify Curse still up after Bane of Agony")
			}
		})
		runSim(sim)
	}
	if got := bases[true] / bases[false]; math.Abs(got-1.5) > 1e-9 {
		t.Errorf("Amplify Curse makes Bane of Agony's base x%.4f, want x1.5", got)
	}
}
