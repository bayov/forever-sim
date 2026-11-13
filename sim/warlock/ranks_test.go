package warlock

import (
	"math"
	"testing"
)

// TestWarlockRanksAtSixty checks the top ranks at level 60 against the Forever client's
// tooltips (ForeverChanges, build 1.60.1.70338): Shadow Bolt 253 to 283, Searing Pain 107 to
// 126, Soul Fire 389 to 488, Shadowburn 258 to 288, Conflagrate 251 to 313, Incinerate 201 to
// 233, Immolate 158 and 5 ticks of 55, Corruption 6 ticks of 73, Bane of Agony 552 over 12
// ticks, Siphon Life 41 a tick and Life Tap 424 before Spirit.
func TestWarlockRanksAtSixty(t *testing.T) {
	ranges := []struct {
		name      string
		rank      spellRank
		low, high float64
	}{
		{"Shadow Bolt", shadowBoltRanks[10], 253.29, 282.71},
		{"Searing Pain", searingPainRanks[6], 106.96, 125.84},
		{"Soul Fire", soulFireRanks[2], 389.32, 487.88},
		{"Shadowburn", shadowburnRanks[6], 258.28, 288.12},
		{"Conflagrate", conflagrateRanks[6], 251.10, 312.90},
		{"Incinerate", incinerateRanks[3], 200.72, 233.28},
		{"Immolate", immolateRanks[8], 158, 158},
		{"Immolate tick", immolateDotRanks[8], 55, 55},
		{"Corruption tick", corruptionRanks[7], 73, 73},
		{"Bane of Agony tick", baneOfAgonyRanks[6], 46, 46},
		{"Siphon Life tick", siphonLifeRanks[4], 41, 41},
		{"Life Tap", lifeTapRanks[6], 424, 424},
		{"Imp Firebolt", impFireboltRanks[7], 42.63, 47.77},
		{"Lash of Pain", lashOfPainRanks[6], 50, 50},
	}
	for _, c := range ranges {
		low, high := c.rank.rangeAt(60)
		if math.Abs(low-c.low) > 0.01 || math.Abs(high-c.high) > 0.01 {
			t.Errorf("%s: got %.2f to %.2f at 60, want %.2f to %.2f", c.name, low, high, c.low, c.high)
		}
	}
}

// TestImpFireboltRankOne checks Imp Firebolt rank 1 after hotfix 113766 (SpellEffect 681913: base
// 6, variance 0.333, 0.3 a level): 5 to 7 at level 1 and 6 to 8.4 at 5, its top level.
func TestImpFireboltRankOne(t *testing.T) {
	for _, c := range []struct {
		level     int32
		low, high float64
	}{{1, 5.00, 7.00}, {5, 6.00, 8.40}, {60, 6.00, 8.40}} {
		low, high := impFireboltRanks[1].rangeAt(c.level)
		if math.Abs(low-c.low) > 0.01 || math.Abs(high-c.high) > 0.01 {
			t.Errorf("Imp Firebolt rank 1 at %d: got %.2f to %.2f, want %.2f to %.2f", c.level, low, high, c.low, c.high)
		}
	}
}

// TestImpFireboltRanksTwoToFour checks Imp Firebolt ranks 2 to 4 against client 1.60.1.70938
// (SpellEffect #0: base 11 / 15 / 18, 0.4 a level, variance 0.1818 / 0.2353 / 0.2222) at their
// learn and top levels.
func TestImpFireboltRanksTwoToFour(t *testing.T) {
	for _, c := range []struct {
		rank      int
		level     int32
		low, high float64
	}{
		{2, 8, 10.00, 12.00}, {2, 13, 11.82, 14.18},
		{3, 18, 13.24, 16.76}, {3, 23, 15.00, 19.00},
		{4, 28, 16.00, 20.00}, {4, 33, 17.78, 22.22},
	} {
		low, high := impFireboltRanks[c.rank].rangeAt(c.level)
		if math.Abs(low-c.low) > 0.01 || math.Abs(high-c.high) > 0.01 {
			t.Errorf("Imp Firebolt rank %d at %d: got %.2f to %.2f, want %.2f to %.2f", c.rank, c.level, low, high, c.low, c.high)
		}
	}
}
