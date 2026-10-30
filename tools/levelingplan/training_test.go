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

func TestDeferRuns(t *testing.T) {
	ranks := []*spellRank{
		{id: 1, spell: "Bolt", level: 10},
		{id: 2, spell: "Bolt", level: 20},
		{id: 3, spell: "Bolt", level: 30},
		{id: 4, spell: "Bolt", level: 40},
		{id: 5, spell: "Shock", level: 10},
		{id: 6, spell: "Shock", level: 12},
		{id: 7, spell: "Shock", level: 14},
		{id: 8, spell: "Shock", level: 16},
	}
	// Bolt 1 to 3 aren't cast while they're our highest, so they come with Bolt 4. Shock 2 costs
	// time at 12 and saves less at 13, so it waits until 13. Shock 3 costs time at both its
	// levels and waits for Shock 4. Shock 1 saves time from its first level.
	loss := map[int32]map[int32]float64{
		5: {10: 0.1, 11: 0.1},
		6: {12: -0.2, 13: 0.1},
		7: {14: -0.1, 15: -0.1},
		8: {16: 0.2},
	}
	cast := map[int32]bool{4: true, 5: true, 6: true, 7: true, 8: true}
	got := deferRuns(ranks, loss, cast, schedule{})
	want := schedule{1: 40, 2: 40, 3: 40, 6: 13, 7: 16}
	if !sameSchedule(got, want) {
		t.Errorf("deferred %v, want %v", got, want)
	}
}
