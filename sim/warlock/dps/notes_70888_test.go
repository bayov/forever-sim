package dps

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The Warlock lines of server hotfix 113766 (build 1.60.1.70888) and the 2026-11-10 notes.

// Siphon Life, Soul Siphon 3/3 and Wrack.
const wrackTalents = "00000000000001301"

// TestWarlockWrack70888 checks Wrack at 60 after hotfix 113766 (SpellEffect 1349556): 60.2 base
// damage a second (25.2 at 40, 1.75 more a level) and 0.2002 of spell damage a tick. Spell damage
// is added after the first cast to read both off the snapshot.
func TestWarlockWrack70888(t *testing.T) {
	sim, lock := newWarlockSim(t, "00000000000000001", &proto.WarlockOptions{})
	target := lock.CurrentTarget
	snapshot := func() float64 {
		dot := lock.Wrack.Dot(target)
		if !dot.IsActive() {
			t.Fatalf("at %s: Wrack not up", sim.CurrentTime)
		}
		return dot.SnapshotBaseDamage - 0.2002*lock.Wrack.GetBonusDamage(target)
	}

	at(sim, 0.5, func(sim *core.Simulation) {
		lock.AddStatDynamic(sim, stats.SpellHit, 20*core.SpellHitRatingPerHitChance)
		castNow(t, sim, lock, lock.Wrack)
	})
	at(sim, 1, func(sim *core.Simulation) {
		if got := snapshot(); math.Abs(got-60.2) > 1e-6 {
			t.Errorf("Wrack's base tick at 60: %.3f, want 60.2", got)
		}
	})
	at(sim, 8, func(sim *core.Simulation) {
		lock.AddStatDynamic(sim, stats.ShadowPower, 300)
		castNow(t, sim, lock, lock.Wrack)
	})
	at(sim, 8.5, func(sim *core.Simulation) {
		if got := snapshot(); math.Abs(got-60.2) > 1e-6 {
			t.Errorf("Wrack's tick at +300 Shadow damage less 0.2002 of it: %.3f, want 60.2", got)
		}
	})
	runSim(sim)
}

// tickCheck expects the dot to have ticked want times by the given second.
func tickCheck(t *testing.T, sim *core.Simulation, dot *core.Dot, seconds float64, want int32) {
	at(sim, seconds, func(sim *core.Simulation) {
		if dot.TickCount != want {
			t.Errorf("at %.2fs: %s ticked %d times, want %d", seconds, dot.Spell.ActionID, dot.TickCount, want)
		}
	})
}

// TestWarlockWrackSpeedsDots checks Wrack's second effect (SpellEffect 1349557, aura 4 value 100):
// Corruption ticks twice as fast while Wrack is channelled and still ends 18 sec after it was cast,
// and the time skipped is kept after the channel ends. Corruption from 1 s ticks at 3 (2 s left at
// Wrack's start halves to 1), 4.5, 6 and 7.5; at 8 the channel ends 1 s (fast) from the next tick,
// which is 2 s at the normal speed, so 10, 13, 16 and 19: 8 ticks by the time it fades at 19, not 6.
func TestWarlockWrackSpeedsDots(t *testing.T) {
	sim, lock := newWarlockSim(t, wrackTalents, &proto.WarlockOptions{})
	target := lock.CurrentTarget
	corruption := last(lock.Corruption).Dot(target)

	at(sim, 1, func(sim *core.Simulation) { corruption.Apply(sim) })
	at(sim, 2, func(sim *core.Simulation) { lock.Wrack.Dot(target).Apply(sim) })
	tickCheck(t, sim, corruption, 2.9, 0)
	tickCheck(t, sim, corruption, 3.1, 1)
	tickCheck(t, sim, corruption, 7.6, 4)
	tickCheck(t, sim, corruption, 9.9, 4)
	tickCheck(t, sim, corruption, 10.1, 5)
	tickCheck(t, sim, corruption, 18.9, 7)
	at(sim, 19.1, func(sim *core.Simulation) {
		if corruption.IsActive() {
			t.Errorf("Corruption still up at 19.1 s, want it to end 18 s after it was applied")
		}
		if corruption.TickCount != 8 {
			t.Errorf("Corruption ticked %d times, want 8", corruption.TickCount)
		}
	})
	runSim(sim)
}

// TestWarlockWrackStoppedEarly checks a channel cut short: Wrack from 2 to 3.5 s. Corruption from
// 1 s ticks at 3 and has run 1 s of its next 3 s tick by 3.5, so ticks at 5.5, 8.5, 11.5, 14.5 and
// 17.5. The progress past 17.5 is lost when it fades at 19: 6 ticks, as without Wrack.
func TestWarlockWrackStoppedEarly(t *testing.T) {
	sim, lock := newWarlockSim(t, wrackTalents, &proto.WarlockOptions{})
	target := lock.CurrentTarget
	corruption := last(lock.Corruption).Dot(target)

	at(sim, 1, func(sim *core.Simulation) { corruption.Apply(sim) })
	at(sim, 2, func(sim *core.Simulation) { lock.Wrack.Dot(target).Apply(sim) })
	at(sim, 3.5, func(sim *core.Simulation) { lock.Wrack.Dot(target).Cancel(sim) })
	tickCheck(t, sim, corruption, 3.1, 1)
	tickCheck(t, sim, corruption, 5.4, 1)
	tickCheck(t, sim, corruption, 5.6, 2)
	tickCheck(t, sim, corruption, 17.6, 6)
	at(sim, 19.1, func(sim *core.Simulation) {
		if corruption.IsActive() || corruption.TickCount != 6 {
			t.Errorf("Corruption at 19.1 s: up %v, %d ticks, want gone after 6", corruption.IsActive(), corruption.TickCount)
		}
	})
	runSim(sim)
}

// TestWarlockWrackAgonyMidChannel checks a dot applied while Wrack is channelled, and Bane of
// Agony's steps past its 12th tick. Wrack from 1 to 7 s; Bane of Agony at 2.25 ticks every second
// (3.25 to 6.25), is a quarter of the way at 7, so ticks at 7.5 and every 2 s to 25.5 before it
// fades at 26.25: 14 ticks. Ticks 13 and 14 stay at the last step, 150%.
func TestWarlockWrackAgonyMidChannel(t *testing.T) {
	sim, lock := newWarlockSim(t, wrackTalents, &proto.WarlockOptions{})
	target := lock.CurrentTarget
	agony := last(lock.BaneOfAgony).Dot(target)
	var afterTwelve float64

	at(sim, 1, func(sim *core.Simulation) { lock.Wrack.Dot(target).Apply(sim) })
	at(sim, 2.25, func(sim *core.Simulation) { agony.Apply(sim) })
	tickCheck(t, sim, agony, 3.3, 1)
	tickCheck(t, sim, agony, 6.3, 4)
	tickCheck(t, sim, agony, 7.4, 4)
	tickCheck(t, sim, agony, 7.6, 5)
	at(sim, 21.6, func(sim *core.Simulation) {
		if agony.TickCount != 12 {
			t.Errorf("Bane of Agony at 21.6 s: %d ticks, want 12", agony.TickCount)
		}
		afterTwelve = agony.SnapshotBaseDamage
	})
	at(sim, 26.4, func(sim *core.Simulation) {
		if agony.IsActive() || agony.TickCount != 14 {
			t.Errorf("Bane of Agony at 26.4 s: up %v, %d ticks, want gone after 14", agony.IsActive(), agony.TickCount)
		}
		if agony.SnapshotBaseDamage != afterTwelve {
			t.Errorf("Bane of Agony's tick past the 12th: %.2f, want %.2f (the last step)", agony.SnapshotBaseDamage, afterTwelve)
		}
	})
	runSim(sim)
}

// TestWarlockSoulSiphon70888 checks Soul Siphon 3/3 on Wrack: 12% for each of Corruption, Siphon
// Life and Bane of Agony on the target, 36% with all three, and nothing for a curse.
func TestWarlockSoulSiphon70888(t *testing.T) {
	multiplier := func(withDots bool) float64 {
		sim, lock := newWarlockSim(t, wrackTalents, &proto.WarlockOptions{})
		target := lock.CurrentTarget
		var got float64
		at(sim, 1, func(sim *core.Simulation) {
			if withDots {
				last(lock.Corruption).Dot(target).Apply(sim)
				last(lock.SiphonLife).Dot(target).Apply(sim)
				agony := last(lock.BaneOfAgony).Dot(target)
				agony.Apply(sim)
			}
			lock.Wrack.Dot(target).Apply(sim)
			got = lock.Wrack.Dot(target).SnapshotAttackerMultiplier
		})
		runSim(sim)
		return got
	}
	if got := multiplier(true) / multiplier(false); math.Abs(got-1.36) > 1e-9 {
		t.Errorf("Wrack with Corruption, Siphon Life and Bane of Agony up and Soul Siphon 3/3: x%.4f, want x1.36", got)
	}
}
