package druid

import (
	"math"
	"testing"
	"time"
)

// Every rank of the Restoration druid's heals against the Forever client (1.60.1.70755: SpellEffect,
// SpellLevels, SpellPower, SpellMisc), as a level 60 character gets it: the direct heal's range, the
// HoT's tick (grown to the rank's max level), the mana cost and the cast time. A rank's growth stops at
// its max level (healer_mechanics-20261104-9).
func TestDruidHealRanksMatchClient(t *testing.T) {
	type direct struct {
		spellID    int32
		cast, mana float64
		low, high  float64
	}
	healingTouch := []direct{
		{5185, 1500, 25, 39.7, 54.7},
		{5186, 2000, 55, 93.7, 119.3},
		{5187, 2500, 110, 187.4, 233.6},
		{5188, 3000, 190, 337.4, 413.6},
		{5189, 3500, 280, 509.2, 617.8},
		{6778, 3500, 350, 659.5, 794.5},
		{8903, 3500, 425, 838.1, 1002.9},
		{9758, 3500, 520, 1094.9, 1303.1},
		{9888, 3500, 630, 1414.8, 1676.2},
		{9889, 3500, 755, 1785.8, 2107.0},
		{25297, 3500, 840, 2138.6, 2525.4},
	}
	for i, c := range healingTouch {
		r := HealingTouchRanks[i+1]
		low, high := r.At(60)
		if r.spellID != c.spellID || r.manaCost != c.mana || r.castTime != time.Duration(c.cast)*time.Millisecond ||
			math.Abs(low-c.low) > 0.06 || math.Abs(high-c.high) > 0.06 {
			t.Errorf("Healing Touch rank %d: %d, %.0f mana, %v, %.1f-%.1f; client %+v", i+1, r.spellID, r.manaCost, r.castTime, low, high, c)
		}
	}

	type hot struct {
		spellID   int32
		mana      float64
		low, high float64 // Regrowth's direct heal
		tick      float64
	}
	for name, c := range map[string]struct {
		ranks []healRank
		want  []hot
	}{
		"Regrowth": {RegrowthRanks, []hot{
			{8936, 70, 88.6, 103.4, 13}, {8938, 125, 164.5, 188.5, 22}, {8939, 170, 233.9, 267.1, 32},
			{8940, 210, 308.6, 349.4, 42}, {8941, 250, 389.5, 439.5, 52}, {9750, 305, 496.4, 558.6, 68},
			{9856, 370, 633.3, 709.7, 88}, {9857, 445, 799.6, 894.4, 113}, {9858, 525, 965.2, 1076.8, 142},
		}},
		"Rejuvenation": {RejuvenationRanks, []hot{
			{774, 25, 0, 0, 8}, {1058, 40, 0, 0, 12}, {1430, 75, 0, 0, 23}, {2090, 105, 0, 0, 32},
			{2091, 135, 0, 0, 42}, {3627, 160, 0, 0, 51}, {8910, 195, 0, 0, 71}, {9839, 235, 0, 0, 94},
			{9840, 280, 0, 0, 124}, {9841, 335, 0, 0, 161}, {25299, 360, 0, 0, 194},
		}},
		"Tranquility": {TranquilityRanks, []hot{
			{740, 375, 0, 0, 90.6}, {8918, 505, 0, 0, 133.2}, {9862, 695, 0, 0, 201.4}, {9863, 925, 0, 0, 285},
		}},
		"Wild Growth": {WildGrowthRanks, []hot{
			{408120, 550, 0, 0, 48}, {1238214, 755, 0, 0, 70.4}, {1238215, 1050, 0, 0, 97},
		}},
	} {
		if len(c.ranks) != len(c.want)+1 {
			t.Errorf("%s: %d ranks, client %d", name, len(c.ranks)-1, len(c.want))
		}
		for i, w := range c.want {
			r := c.ranks[i+1]
			low, high := r.At(60)
			if r.base == 0 {
				low, high = 0, 0
			}
			if r.spellID != w.spellID || r.manaCost != w.mana || math.Abs(r.TickAt(60)-w.tick) > 0.06 ||
				math.Abs(low-w.low) > 0.06 || math.Abs(high-w.high) > 0.06 {
				t.Errorf("%s rank %d: %d, %.0f mana, %.1f-%.1f, tick %.1f; client %+v", name, i+1, r.spellID, r.manaCost, low, high, r.TickAt(60), w)
			}
		}
	}

	// The client's cast times, coefficients and HoT lengths, the same for every rank.
	if RegrowthCastTime != 2*time.Second || WildGrowthCastTime != 3*time.Second || RegrowthCoefficient != 0.286 ||
		RegrowthTickCoefficient != 0.071 || RejuvenationTickCoefficient != 0.2 || TranquilityTickCoefficient != 0.067 ||
		WildGrowthTickCoefficient != 0.033 || RegrowthTicks != 7 || RejuvenationTicks != 4 || TranquilityTicks != 5 || WildGrowthTicks != 7 {
		t.Error("a cast time, coefficient or tick count differs from the client's")
	}
	for i, want := range []float64{0.429, 0.571, 0.714, 0.857, 1, 1, 1, 1, 1, 1, 1} {
		if HealingTouchRanks[i+1].coefficient != want {
			t.Errorf("Healing Touch rank %d coefficient %.3f, client %.3f", i+1, HealingTouchRanks[i+1].coefficient, want)
		}
	}
}
