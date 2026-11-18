package core

import (
	"math/rand"
	"testing"
	"time"
	"unsafe"
)

// healthAtLinear is healthAt as a walk from the newest change back, which the binary search must
// match.
func (hb *healthBar) healthAtLinear(t time.Duration) float64 {
	health := hb.currentHealth
	for k := 1; k <= healthHistorySize; k++ {
		c := hb.history[(hb.next-k+healthHistorySize)%healthHistorySize]
		if !c.set || c.at <= t {
			break
		}
		health = c.before
	}
	return health
}

func TestHealthAtMatchesLinearWalk(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for trial := 0; trial < 2000; trial++ {
		hb := healthBar{currentHealth: 5000}
		at := time.Duration(rng.Intn(1000)) * time.Millisecond
		changes := rng.Intn(3 * healthHistorySize)
		for i := 0; i < changes; i++ {
			// Some changes share a time, as several hits in one instant do.
			if rng.Intn(3) > 0 {
				at += time.Duration(rng.Intn(400)) * time.Millisecond
			}
			hb.history[hb.next] = healthChange{at: at, before: hb.currentHealth, set: true}
			hb.next = (hb.next + 1) % healthHistorySize
			hb.currentHealth = float64(rng.Intn(5000))
		}
		for q := 0; q < 50; q++ {
			ts := time.Duration(rng.Intn(int(at/time.Millisecond)+500)-200) * time.Millisecond
			if got, want := hb.healthAt(ts), hb.healthAtLinear(ts); got != want {
				t.Fatalf("trial %d, %d changes: healthAt(%v) = %v, want %v", trial, changes, ts, got, want)
			}
		}
	}
}

func TestSpellMetricsSizeIsWords(t *testing.T) {
	if unsafe.Sizeof(SpellMetrics{})%8 != 0 {
		t.Fatalf("SpellMetrics is %d bytes, not a whole number of words (allZeroBytes)", unsafe.Sizeof(SpellMetrics{}))
	}
}
