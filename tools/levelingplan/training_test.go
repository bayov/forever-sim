package main

import (
	"slices"
	"testing"
)

func TestCheapestCover(t *testing.T) {
	a := &spellRank{id: 1}
	b := &spellRank{id: 2}
	c := &spellRank{id: 3}
	items := []knapsackItem{
		{rank: a, weight: 5000, cost: 3},
		{rank: b, weight: 3000, cost: 1},
		{rank: c, weight: 3000, cost: 1},
	}
	// 50s short: b and c together (60s for 2) beat a alone (50s for 3).
	var got []int32
	for _, item := range cheapestCover(items, 5000) {
		got = append(got, item.rank.id)
	}
	slices.Sort(got)
	if !slices.Equal(got, []int32{2, 3}) {
		t.Errorf("50s short picked %v, want [2 3]", got)
	}
	// 20s short: one of the cheap ones is enough.
	if got := cheapestCover(items, 2000); len(got) != 1 || got[0].cost != 1 {
		t.Errorf("20s short picked %v, want one rank of cost 1", got)
	}
	// More than all of them weigh: all of them.
	if got := cheapestCover(items, 20000); len(got) != 3 {
		t.Errorf("2g short picked %d ranks, want all 3", len(got))
	}
}

func TestSkipTails(t *testing.T) {
	ranks := []*spellRank{
		{id: 1, spell: "Shock", level: 10},
		{id: 2, spell: "Shock", level: 20},
		{id: 3, spell: "Shock", level: 30},
		{id: 4, spell: "Shock", level: 40},
		{id: 5, spell: "Totem", level: 20},
		{id: 6, spell: "Totem", level: 30},
	}
	// Shock 2 costs a little but Shock 3 saves more, so we keep both. Shock 4 saves nothing.
	// Totem 2 costs time.
	minutes := map[int32]float64{1: 5, 2: -1, 3: 4, 4: 0, 5: 3, 6: -2}
	got := skipTails(ranks, minutes)
	want := schedule{4: neverLevel, 6: neverLevel}
	if !sameSchedule(got, want) {
		t.Errorf("skipped %v, want %v", got, want)
	}
}
