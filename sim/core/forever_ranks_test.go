package core

import "testing"

// TestForeverBestLearnedRank checks that Forever's Trueshot Aura and Hunter's Mark use the
// strongest rank learned by a level, not the highest (standards 8a: players cast any rank they
// know). The launch realms' 2026-10-28 hotfix puts Trueshot Aura on a rising curve (15 / 22 / 30 /
// 40, rank 5 the client's 50), and the 2026-10-21 server hotfix gives Hunter's Mark rank 4 143
// (the client's is 71 against rank 3's 98), so the top ranks are cast.
func TestForeverBestLearnedRank(t *testing.T) {
	for _, c := range []struct {
		name   string
		levels []int
		values []float64
		level  int32
		want   float64
	}{
		{"Trueshot Aura", foreverTrueshotAuraLevel[:], foreverTrueshotAuraRAP[:], 24, 0},
		{"Trueshot Aura", foreverTrueshotAuraLevel[:], foreverTrueshotAuraRAP[:], 45, 30},
		{"Trueshot Aura", foreverTrueshotAuraLevel[:], foreverTrueshotAuraRAP[:], 50, 40},
		{"Trueshot Aura", foreverTrueshotAuraLevel[:], foreverTrueshotAuraRAP[:], 59, 40},
		{"Trueshot Aura", foreverTrueshotAuraLevel[:], foreverTrueshotAuraRAP[:], 60, 50},
		{"Hunter's Mark", foreverHuntersMarkLevel[:], foreverHuntersMark[:], 30, 59},
		{"Hunter's Mark", foreverHuntersMarkLevel[:], foreverHuntersMark[:], 40, 98},
		{"Hunter's Mark", foreverHuntersMarkLevel[:], foreverHuntersMark[:], 57, 98},
		{"Hunter's Mark", foreverHuntersMarkLevel[:], foreverHuntersMark[:], 60, 143},
	} {
		if got := bestLearnedRank(c.level, c.levels, c.values); got != c.want {
			t.Errorf("%s at level %d: %.0f, want %.0f", c.name, c.level, got, c.want)
		}
	}
}
