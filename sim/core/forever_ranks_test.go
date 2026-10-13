package core

import "testing"

// TestForeverBestLearnedRank checks that Forever's Trueshot Aura and Hunter's Mark use the
// strongest rank learned by a level, not the highest. Client 70338 gives Trueshot Aura rank 5
// (level 60) 50 against rank 4's 75, and Hunter's Mark rank 4 (level 58) 71 against rank 3's 98
// (standards 8a: players cast any rank they know).
func TestForeverBestLearnedRank(t *testing.T) {
	for _, c := range []struct {
		name   string
		levels []int
		values []float64
		level  int32
		want   float64
	}{
		{"Trueshot Aura", foreverTrueshotAuraLevel[:], foreverTrueshotAuraRAP[:], 24, 0},
		{"Trueshot Aura", foreverTrueshotAuraLevel[:], foreverTrueshotAuraRAP[:], 45, 50},
		{"Trueshot Aura", foreverTrueshotAuraLevel[:], foreverTrueshotAuraRAP[:], 50, 75},
		{"Trueshot Aura", foreverTrueshotAuraLevel[:], foreverTrueshotAuraRAP[:], 60, 75},
		{"Hunter's Mark", foreverHuntersMarkLevel[:], foreverHuntersMark[:], 30, 59},
		{"Hunter's Mark", foreverHuntersMarkLevel[:], foreverHuntersMark[:], 40, 98},
		{"Hunter's Mark", foreverHuntersMarkLevel[:], foreverHuntersMark[:], 60, 98},
	} {
		if got := bestLearnedRank(c.level, c.levels, c.values); got != c.want {
			t.Errorf("%s at level %d: %.0f, want %.0f", c.name, c.level, got, c.want)
		}
	}
}
