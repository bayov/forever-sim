package core

import (
	"math"
	"testing"
)

func TestAoESoftCapMultiplier(t *testing.T) {
	for _, c := range []struct {
		targets, cap int32
		rule         AoESoftCapRule
		want         float64
	}{
		{1, 5, AoESoftCapSqrt, 1},
		{5, 5, AoESoftCapSqrt, 1},
		{20, 5, AoESoftCapSqrt, 0.5},
		{10, 10, AoESoftCapSplit, 1},
		{20, 10, AoESoftCapSplit, 0.5},
		{3, 0, AoESoftCapSplit, 1},
	} {
		if got := AoESoftCapMultiplier(c.targets, c.cap, c.rule); math.Abs(got-c.want) > 1e-12 {
			t.Errorf("%d targets, cap %d, rule %d: %v, want %v", c.targets, c.cap, c.rule, got, c.want)
		}
	}
}
