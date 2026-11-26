package main

import (
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
)

// The game gives the first talent point at 10. With the Legacy perk Talented 5/5 it comes at 5,
// and the 51 points are full at 55.
func TestTalentPointsByFirstLevel(t *testing.T) {
	for _, c := range []struct {
		first, bonus, level int32
		points              int
		playerBonus         int32
	}{
		{0, 0, 9, 0, 0},
		{0, 0, 10, 1, 0},
		{10, 0, 59, 50, 0},
		{0, 0, 60, 51, 0},
		{0, 5, 24, 15, 0},
		{0, 5, 25, 21, 5},
		{0, 5, 60, 51, 5},
		{5, 0, 4, 0, 0},
		{5, 0, 5, 1, 1},
		{5, 0, 9, 5, 5},
		{5, 0, 10, 6, 5},
		{5, 0, 55, 51, 5},
		{5, 0, 60, 51, 0},
		{9, 0, 9, 1, 1},
	} {
		points := talentPointsOf(&proto.LevelingPlanSettings{FirstTalentLevel: c.first, BonusTalentPoints: c.bonus})
		if got := points.at(c.level); got != c.points {
			t.Errorf("first %d, bonus %d, level %d: %d points, want %d", c.first, c.bonus, c.level, got, c.points)
		}
		if got := points.bonusAt(c.level); got != c.playerBonus {
			t.Errorf("first %d, bonus %d, level %d: %d on top of the player's own, want %d", c.first, c.bonus, c.level, got, c.playerBonus)
		}
	}
}

func TestPickLevelByFirstLevel(t *testing.T) {
	for _, c := range []struct {
		first, start int32
		index        int
		want         int32
	}{
		{0, 1, 0, 10},
		{0, 1, 5, 15},
		{0, 30, 5, 30},
		{5, 1, 0, 5},
		{5, 1, 5, 10},
		{5, 1, 50, 55},
		{5, 1, 51, 61},
	} {
		points := talentPoints{first: c.first}
		if got := points.pickLevel(c.index, c.start); got != c.want {
			t.Errorf("first %d, point %d of a build from %d: level %d, want %d", c.first, c.index, c.start, got, c.want)
		}
	}
}
