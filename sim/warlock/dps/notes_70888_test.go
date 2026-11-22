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
// while Wrack is channelled Corruption ticks the half points of its own grid too, and still ends 18 sec
// after it was cast, on a full tick (wowsims/forever's EV-0085). Corruption from 1 s ticks every 3 s
// from 4; Wrack from 2 to 8 adds its half points 2.5 and 5.5, so 2.5, 4, 5.5 and 7 in the channel,
// then 10, 13, 16 and 19: 8 ticks by the time it fades at 19, not 6.
func TestWarlockWrackSpeedsDots(t *testing.T) {
	sim, lock := newWarlockSim(t, wrackTalents, &proto.WarlockOptions{})
	target := lock.CurrentTarget
	corruption := last(lock.Corruption).Dot(target)

	at(sim, 1, func(sim *core.Simulation) { corruption.Apply(sim) })
	at(sim, 2, func(sim *core.Simulation) { lock.Wrack.Dot(target).Apply(sim) })
	tickCheck(t, sim, corruption, 2.4, 0)
	tickCheck(t, sim, corruption, 2.6, 1)
	tickCheck(t, sim, corruption, 7.1, 4)
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

// TestWarlockWrackStoppedEarly checks a channel cut short, and that a dot Wrack sped up still ticks
// in full at its expiry (EV-0085: all 4 logged effects that ran out after a channel ended on a full
// tick). Wrack from 2 to 3.5 s: Corruption from 1 s ticks its half point at 2.5, then its normal grid
// 4, 7, ..., 19, the last as it fades: 7 ticks, one more than without Wrack.
func TestWarlockWrackStoppedEarly(t *testing.T) {
	sim, lock := newWarlockSim(t, wrackTalents, &proto.WarlockOptions{})
	target := lock.CurrentTarget
	corruption := last(lock.Corruption).Dot(target)

	at(sim, 1, func(sim *core.Simulation) { corruption.Apply(sim) })
	at(sim, 2, func(sim *core.Simulation) { lock.Wrack.Dot(target).Apply(sim) })
	at(sim, 3.5, func(sim *core.Simulation) { lock.Wrack.Dot(target).Cancel(sim) })
	tickCheck(t, sim, corruption, 2.6, 1)
	tickCheck(t, sim, corruption, 3.9, 1)
	tickCheck(t, sim, corruption, 4.1, 2)
	tickCheck(t, sim, corruption, 18.9, 6)
	at(sim, 19.1, func(sim *core.Simulation) {
		if corruption.IsActive() || corruption.TickCount != 7 {
			t.Errorf("Corruption at 19.1 s: up %v, %d ticks, want gone after 7", corruption.IsActive(), corruption.TickCount)
		}
	})
	runSim(sim)
}

// TestWarlockWrackHalfPoints checks a channel that lands on a half point and one that ends on one.
// Corruption from 1 s has its half points at 2.5, 5.5, 8.5 and so on. Wrack cast at 2.5 ticks that
// half point at once (EV-0085: Wrack one global cooldown after Siphon Life ticked its 1.5 s half point
// in 13 of 16 openers). The channel ends at 8.5, on a half point, which ticks too: not measured (the
// logs' 0.05 s steps cannot split it), kept from our earlier model. So 2.5, 4, 5.5, 7, 8.5, then 10,
// 13, 16 and 19: 9 ticks.
func TestWarlockWrackHalfPoints(t *testing.T) {
	sim, lock := newWarlockSim(t, wrackTalents, &proto.WarlockOptions{})
	target := lock.CurrentTarget
	corruption := last(lock.Corruption).Dot(target)

	at(sim, 1, func(sim *core.Simulation) { corruption.Apply(sim) })
	at(sim, 2.5, func(sim *core.Simulation) { lock.Wrack.Dot(target).Apply(sim) })
	tickCheck(t, sim, corruption, 2.51, 1)
	tickCheck(t, sim, corruption, 8.4, 4)
	tickCheck(t, sim, corruption, 8.6, 5)
	tickCheck(t, sim, corruption, 9.9, 5)
	tickCheck(t, sim, corruption, 10.1, 6)
	at(sim, 19.1, func(sim *core.Simulation) {
		if corruption.IsActive() || corruption.TickCount != 9 {
			t.Errorf("Corruption at 19.1 s: up %v, %d ticks, want gone after 9", corruption.IsActive(), corruption.TickCount)
		}
	})
	runSim(sim)
}

// TestWarlockWrackAgonyMidChannel checks a dot applied while Wrack is channelled, and Bane of
// Agony's steps past its 12th tick. Wrack from 1 to 7 s; Bane of Agony at 2.25 ticks the half points
// of its 2 s grid from its application (not measured, EV-0085 leaves it open): 3.25 to 6.25, then
// 8.25 and every 2 s to 26.25, the last as it fades: 14 ticks. Ticks 13 and 14 stay at the last
// step, 150%.
func TestWarlockWrackAgonyMidChannel(t *testing.T) {
	sim, lock := newWarlockSim(t, wrackTalents, &proto.WarlockOptions{})
	target := lock.CurrentTarget
	agony := last(lock.BaneOfAgony).Dot(target)
	var afterTwelve float64

	at(sim, 1, func(sim *core.Simulation) { lock.Wrack.Dot(target).Apply(sim) })
	at(sim, 2.25, func(sim *core.Simulation) { agony.Apply(sim) })
	tickCheck(t, sim, agony, 3.3, 1)
	tickCheck(t, sim, agony, 6.3, 4)
	tickCheck(t, sim, agony, 8.2, 4)
	tickCheck(t, sim, agony, 8.3, 5)
	at(sim, 22.3, func(sim *core.Simulation) {
		if agony.TickCount != 12 {
			t.Errorf("Bane of Agony at 22.3 s: %d ticks, want 12", agony.TickCount)
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
