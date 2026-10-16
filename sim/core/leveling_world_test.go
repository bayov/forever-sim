package core

import (
	"math"
	"testing"
)

// TestWhenWithin checks the time a moving point first comes within a radius of us.
func TestWhenWithin(t *testing.T) {
	for _, c := range []struct {
		name        string
		rel, relVel Vec2
		r, want     float64
		ok          bool
	}{
		{"already within", Vec2{3, 4}, Vec2{1, 0}, 5, 0, true},
		{"head on", Vec2{30, 0}, Vec2{-5, 0}, 5, 5, true},
		{"moving away", Vec2{30, 0}, Vec2{5, 0}, 5, 0, false},
		{"standing still", Vec2{30, 0}, Vec2{}, 5, 0, false},
		// Passing 3 yards to the side comes within 5 when 4 yards short: 26 yards at 2 a sec.
		{"passing by", Vec2{30, 3}, Vec2{-2, 0}, 5, 13, true},
		{"passing too wide", Vec2{30, 6}, Vec2{-2, 0}, 5, 0, false},
	} {
		got, ok := whenWithin(c.rel, c.relVel, c.r)
		if ok != c.ok || math.Abs(got-c.want) > 1e-9 {
			t.Errorf("%s: got %.4f, %v, want %.4f, %v", c.name, got, ok, c.want, c.ok)
		}
	}
}
