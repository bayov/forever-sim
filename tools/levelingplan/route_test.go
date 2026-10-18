package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
)

// TestRouteSearchMatchesTS runs the route search on a made up problem with the shaman's
// trees: hours from a worth for each talent and a little jitter, Improved Ghost Wolf required
// by 16 and not simmed, levels 10 to 39, then respecs at 20 and 30 and the route again, like
// runSearch. The UI's TS search (leveling_talent_search.ts, before the port) found these
// routes with these sims on the same problem, so a change to the search that moves them is
// a change from the search we had.
func TestRouteSearchMatchesTS(t *testing.T) {
	shape := shamanShape(t)
	worth := func(tree, i int) float64 {
		w := float64((tree*31+i*17)%23) / 1000
		if tree == 1 && i == 12 {
			w += 0.25
		}
		if tree == 1 && i == 11 {
			w += 0.03
		}
		return w
	}
	hours := func(level int32, talents string) float64 {
		counts := parseTalents(talents, shape)
		v := 0.0
		sums := make([]int, len(counts))
		for tree, c := range counts {
			for i, n := range c {
				v += float64(n) * worth(tree, i)
				sums[tree] += n
			}
		}
		v -= 0.002 * float64(min(sums[0], sums[1]))
		jitter := 0.004 * (float64(djb2(fmt.Sprintf("%d|%s", level, talents)))/4294967296 - 0.5)
		return (1 + float64(level)/10) / (1 + v) * (1 + jitter)
	}
	var levels []int32
	for l := int32(10); l < 40; l++ {
		levels = append(levels, l)
	}
	s := &routeSearch{
		shape:    shape,
		levels:   levels,
		required: []*proto.LevelingTalentRequirement{{Tree: 1, Talent: 5, Points: 2, Level: 16}},
		ctx:      context.Background(),
		cache:    map[string]levelHours{},
		progress: func(string, float64) {},
		sim: func(jobs []levelJob) ([]levelHours, []bool) {
			out := make([]levelHours, len(jobs))
			done := make([]bool, len(jobs))
			for i, j := range jobs {
				out[i], done[i] = levelHours{hours: hours(j.level, j.talents), relativeError: 0.003}, true
			}
			return out, done
		},
	}
	total := func(route []pick, levels []int32) float64 {
		sum := 0.0
		for _, l := range levels {
			sum += s.hoursAt(l, s.talentsAt(route, l))
		}
		return sum
	}
	from := func(level int32) []int32 {
		var out []int32
		for _, l := range levels {
			if l >= level {
				out = append(out, l)
			}
		}
		return out
	}
	check := func(stage string, route []pick, levels []int32, wantHours string, wantSims int, wantRoute string) {
		t.Helper()
		got := fmt.Sprintf("%.6f %d %s", total(route, levels), s.sims, routeString(route))
		want := fmt.Sprintf("%s %d %s", wantHours, wantSims, wantRoute)
		if got != want {
			t.Errorf("%s:\n got %s\nwant %s", stage, got, want)
		}
	}

	route, err := s.search([][]pick{nil}, levels)
	if err != nil {
		t.Fatal(err)
	}
	check("route", route, levels, "74.566120", 2257, "1,2 1,2 1,2 1,2 1,2 1,5 1,5 1,6 1,3 1,3 1,7 1,0 1,0 1,4 1,4 1,12 1,4 1,11 1,11 1,11 1,11 1,11 1,13 1,6 1,6 1,7 1,10 1,10 1,14 1,14")
	respec20, _ := s.search([][]pick{route, nil}, from(20))
	check("respec 20", respec20, from(20), "51.202579", 4090, "1,2 1,2 1,2 1,2 1,2 1,6 1,3 1,3 1,6 1,5 1,5 1,6 1,4 1,4 1,4 1,12 1,11 1,11 1,11 1,11 1,11 1,13 2,0 1,14 1,10 1,10 2,0 2,0 2,0 2,0")
	respec30, _ := s.search([][]pick{route, nil}, from(30))
	check("respec 30", respec30, from(30), "25.941936", 5243, "1,2 1,2 1,2 1,2 1,2 1,6 1,3 1,3 1,0 1,4 1,4 1,4 1,6 1,6 1,7 1,11 1,11 1,5 1,5 1,12 1,11 1,11 1,11 1,13 1,10 0,1 2,0 1,7 2,0 1,14")
	again, _ := s.search([][]pick{route, respec20, respec30}, levels)
	check("again", again, levels, "73.635098", 5550, "1,2 1,2 1,2 1,2 1,2 1,5 1,5 1,6 1,6 1,6 1,3 1,3 1,4 1,4 1,4 1,12 1,11 1,11 1,11 1,11 1,11 1,10 1,13 1,10 0,1 2,0 1,14 2,0 1,10 1,10")
}

func routeString(route []pick) string {
	parts := make([]string, len(route))
	for i, p := range route {
		parts[i] = fmt.Sprintf("%d,%d", p.tree, p.talent)
	}
	return strings.Join(parts, " ")
}

func djb2(s string) uint32 {
	h := uint32(5381)
	for _, c := range []byte(s) {
		h = h*33 + uint32(c)
	}
	return h
}

// The shaman's trees from the UI's talent files, with Improved Ghost Wolf not simmed.
func shamanShape(t *testing.T) *proto.LevelingTalentShape {
	data, err := os.ReadFile("../../ui/core/talents/trees/shaman.json")
	if err != nil {
		t.Fatal(err)
	}
	type location struct{ RowIdx, ColIdx int32 }
	var trees []struct {
		Name    string
		Talents []struct {
			FieldName      string
			MaxPoints      int32
			Location       location
			PrereqLocation *location
		}
	}
	if err := json.Unmarshal(data, &trees); err != nil {
		t.Fatal(err)
	}
	shape := &proto.LevelingTalentShape{PointsPerRow: 5}
	for ti, tree := range trees {
		out := &proto.LevelingTalentTree{Name: tree.Name}
		for i, talent := range tree.Talents {
			prereq := int32(-1)
			if talent.PrereqLocation != nil {
				for j, other := range tree.Talents {
					if other.Location == *talent.PrereqLocation {
						prereq = int32(j)
					}
				}
			}
			out.Talents = append(out.Talents, &proto.LevelingTalentConfig{
				Name: talent.FieldName, Row: talent.Location.RowIdx, Col: talent.Location.ColIdx,
				MaxPoints: talent.MaxPoints, Prereq: prereq, NotSimulated: ti == 1 && i == 5,
			})
		}
		shape.Trees = append(shape.Trees, out)
	}
	return shape
}
