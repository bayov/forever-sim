package main

import (
	"math"
	"testing"

	"github.com/wowsims/classic/tools/healscore"
)

// The linear estimate is k + (1 - f)/s inside the trust band, and refused outside it.
func TestHealEstimate(t *testing.T) {
	for _, c := range []struct {
		k, s, f, want float64
		ok            bool
	}{
		{1.2, 5, 1, 1.2, true},
		{1.2, 5, 0.5, 1.3, true},
		{1.2, 5, 2, 1.0, true},
		{1.2, 5, 0, 0, false},
		{1.2, 5, 0.2, 0, false},
		{1.2, 5, 3.5, 0, false},
		{1.2, 0, 1, 0, false},
		{1.2, math.Inf(1), 1, 0, false},
	} {
		got, ok := estimate(c.k, c.s, c.f)
		if ok != c.ok || (ok && math.Abs(got-c.want) > 1e-12) {
			t.Errorf("estimate(%g, %g, %g) = %g, %t, want %g, %t", c.k, c.s, c.f, got, ok, c.want, c.ok)
		}
	}
}

// The fallback finds the line from either side of it on a steep, curved failure, where the
// linear estimate would be far off, and agrees with the full carry's search.
func TestHealBracket(t *testing.T) {
	line := 1.37
	failure := func(x float64) float64 { return math.Pow(x/line, 8) }
	sim := func(x float64) (float64, int, error) { return failure(x), int(1000 * 0.02 * failure(x)), nil }
	full, _ := healscore.CarrySearch(func(x float64) (float64, error) { return failure(x), nil })
	for _, k := range []float64{0.8, 1.0, 1.6, 2.5} {
		c, _, err := healBracket(k, failure(k), 0, sim)
		if err != nil {
			t.Fatal(err)
		}
		if c.Lo > line || c.Hi < line {
			t.Errorf("from %g: bracket [%g, %g] misses the line at %g", k, c.Lo, c.Hi, line)
		}
		if math.Abs(c.K-line) > 0.01 {
			t.Errorf("from %g: carry %.4f, the line is at %g (full search %.4f)", k, c.K, line, full.K)
		}
	}
	if math.Abs(full.K-line) > 0.01 {
		t.Errorf("full search: carry %.4f, the line is at %g", full.K, line)
	}
}

// Two heal results pair fight by fight: the same fights give no noise, and results that don't
// pair (a fallback, or no parts on one side) still give an error and don't crash.
func TestHealGainNoise(t *testing.T) {
	slot := &healSlot{role: "tank", slope: 4}
	fails := make([]bool, 1000)
	for i := 0; i < 20; i++ {
		fails[i*50] = true
	}
	a := result{dps: 120, stderr: 3, healParts: []healPart{{slot: slot, failed: fails, stderr: 3}}}
	if n := gainNoise(a, a); n != 0 {
		t.Errorf("the same fights: noise %g, want 0", n)
	}
	moved := append([]bool(nil), fails...)
	moved[1], moved[2] = true, true
	b := result{dps: 118, stderr: 3, healParts: []healPart{{slot: slot, failed: moved, stderr: 3}}}
	if n := gainNoise(a, b); !(n > 0 && n < 3) {
		t.Errorf("two more failed fights: noise %g, want a paired error under the unpaired 3", n)
	}
	fb := result{dps: 118, stderr: 4, healParts: []healPart{{slot: slot, stderr: 4}}}
	if n := gainNoise(a, fb); math.Abs(n-5) > 1e-9 {
		t.Errorf("against a fallback: noise %g, want 5", n)
	}
	if n := gainNoise(a, result{dps: 100, stderr: 4}); math.Abs(n-5) > 1e-9 {
		t.Errorf("against a result without parts: noise %g, want 5", n)
	}
}
