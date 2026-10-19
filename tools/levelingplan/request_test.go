package main

import (
	"math"
	"testing"
)

func TestMergedHours(t *testing.T) {
	near := func(a, b float64) bool { return math.Abs(a-b) < 1e-9 }
	// 10 XP an hour over 100 sessions and 20 over 300 sessions average to 17.5, so a level
	// of 35 XP takes 2 hours.
	got := levelHours{hours: 3.5, relativeError: 0.02}.merged(100, levelHours{hours: 1.75, relativeError: 0.01}, 300)
	if !near(got.hours, 2) {
		t.Errorf("hours %v, want 2", got.hours)
	}
	if want := math.Sqrt(100*100*0.02*0.02+300*300*0.01*0.01) / 400; !near(got.relativeError, want) {
		t.Errorf("relative error %v, want %v", got.relativeError, want)
	}
	// Twice the sessions of the same sim have the noise over the square root of 2.
	same := levelHours{hours: 4, relativeError: 0.03}.merged(200, levelHours{hours: 4, relativeError: 0.03}, 200)
	if !near(same.hours, 4) || !near(same.relativeError, 0.03/math.Sqrt2) {
		t.Errorf("same sims merged to %+v", same)
	}
}
