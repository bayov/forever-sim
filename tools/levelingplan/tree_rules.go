package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wowsims/classic/sim/core/proto"
)

// routeProblem says which pick of the plan's builds breaks the talent tree's rules first, or
// nil when none does.
//
// The sim takes a talents string at each level and doesn't check it, so an illegal route sims
// fine and its hours look real. That happened on 2026-11-04: the November 2 notes moved
// Improved Ghost Wolf from row 2 to row 4, its two points moved from 15-16 to 25-26 by hand,
// and that left Shamanistic Focus at 18 and Stormstrike at 23 before their rows were open.
// We check only the picks the plan plays (before its last level), so a route that runs past a
// short plan's end doesn't stop it.
func routeProblem(plan *proto.LevelingPlanSettings, shape *proto.LevelingTalentShape) error {
	for b, build := range sortedBuilds(plan) {
		start := build.Level
		if b == 0 {
			start = plan.FromLevel
		}
		for i, reason := range pickReasons(picksFromProto(build.Picks), shape) {
			level := pickLevel(i, start, plan.BonusTalentPoints)
			if level >= plan.ToLevel {
				break
			}
			if reason == "" {
				continue
			}
			p := build.Picks[i]
			name := fmt.Sprintf("talent %d of tree %d", p.Talent, p.Tree)
			if int(p.Tree) < len(shape.Trees) && int(p.Talent) < len(shape.Trees[p.Tree].Talents) {
				name = shape.Trees[p.Tree].Talents[p.Talent].Name
			}
			respec := ""
			if b > 0 {
				respec = fmt.Sprintf(" (the respec at %d)", build.Level)
			}
			return fmt.Errorf("the talent route breaks the tree's rules: %s at level %d%s, %s", name, level, respec, reason)
		}
	}
	return nil
}

// treeFileShape reads a class's talent trees from the UI's talent file
// (ui/core/talents/trees/<class>.json) in the shape the requests carry, without the
// NotSimulated marks, which the UI adds.
func treeFileShape(data []byte) (*proto.LevelingTalentShape, error) {
	type location struct{ RowIdx, ColIdx int32 }
	var trees []struct {
		Name    string
		Talents []struct {
			Name           string
			MaxPoints      int32
			Location       location
			PrereqLocation *location
		}
	}
	if err := json.Unmarshal(data, &trees); err != nil {
		return nil, err
	}
	shape := &proto.LevelingTalentShape{PointsPerRow: 5}
	for _, tree := range trees {
		out := &proto.LevelingTalentTree{Name: tree.Name}
		for _, talent := range tree.Talents {
			prereq := int32(-1)
			if talent.PrereqLocation != nil {
				for j, other := range tree.Talents {
					if other.Location == *talent.PrereqLocation {
						prereq = int32(j)
					}
				}
			}
			out.Talents = append(out.Talents, &proto.LevelingTalentConfig{
				Name: talent.Name, Row: talent.Location.RowIdx, Col: talent.Location.ColIdx, MaxPoints: talent.MaxPoints, Prereq: prereq,
			})
		}
		shape.Trees = append(shape.Trees, out)
	}
	return shape, nil
}

// shapeDifference says how a request's talent shape differs from the trees, or "" when it
// has the same talents in the same places.
func shapeDifference(got, want *proto.LevelingTalentShape) string {
	if got.PointsPerRow != want.PointsPerRow {
		return fmt.Sprintf("%d points a row, the trees have %d", got.PointsPerRow, want.PointsPerRow)
	}
	if len(got.Trees) != len(want.Trees) {
		return fmt.Sprintf("%d trees, the class has %d", len(got.Trees), len(want.Trees))
	}
	for t, tree := range want.Trees {
		if len(got.Trees[t].Talents) != len(tree.Talents) {
			return fmt.Sprintf("%d talents in %s, the tree has %d", len(got.Trees[t].Talents), tree.Name, len(tree.Talents))
		}
		for i, w := range tree.Talents {
			g := got.Trees[t].Talents[i]
			if g.Row != w.Row || g.Col != w.Col || g.MaxPoints != w.MaxPoints || g.Prereq != w.Prereq {
				return fmt.Sprintf("%s talent %d as %s at row %d, column %d with %d ranks, the tree has %s at row %d, column %d with %d ranks",
					tree.Name, i+1, g.Name, g.Row, g.Col, g.MaxPoints, w.Name, w.Row, w.Col, w.MaxPoints)
			}
		}
	}
	return ""
}

// checkShape refuses a request whose talent shape isn't the class's trees in this repo.
//
// A request keeps the shape of the trees from when the UI made it. One saved before a tree
// change (Improved Ghost Wolf moving to row 4 on November 2) still sims, with the old rows,
// and gives wrong hours (88.0h against 85.9h for the shield plan). When we can't find the
// trees (the tool runs outside the repo), we don't check.
func checkShape(req *proto.LevelingPlanSearchRequest) error {
	players := req.GetBase().GetRaid().GetParties()
	if len(players) == 0 || len(players[0].Players) == 0 {
		return nil
	}
	class := strings.ToLower(strings.TrimPrefix(players[0].Players[0].Class.String(), "Class"))
	path := findTreeFile(class)
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	want, err := treeFileShape(data)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}
	if diff := shapeDifference(req.TalentShape, want); diff != "" {
		return fmt.Errorf("the request's talent trees aren't %s's: it has %s. Make the request again from the UI", path, diff)
	}
	return nil
}

// findTreeFile looks for ui/core/talents/trees/<class>.json in the working directory and the
// ones above it.
func findTreeFile(class string) string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		path := filepath.Join(dir, "ui", "core", "talents", "trees", class+".json")
		if _, err := os.Stat(path); err == nil {
			return path
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
