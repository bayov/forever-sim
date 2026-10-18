package main

import (
	"slices"
	"strconv"
	"strings"

	"github.com/wowsims/classic/sim/core/proto"
)

// The leveling plan, as ui/core/leveling_plan.ts has it: which talents we have and which
// rotation we play at each level. The gear is in gear.go.

// XP from each level to the next, in Classic 1.12. Forever may have changed it (Need to Verify
// in notes.md). We have no XP numbers from the beta: the combat log has the kills but not
// their XP.
var xpToNextLevel = map[int32]float64{
	1: 400, 2: 900, 3: 1400, 4: 2100, 5: 2800, 6: 3600, 7: 4500, 8: 5400, 9: 6500, 10: 7600,
	11: 8800, 12: 10100, 13: 11400, 14: 12900, 15: 14400, 16: 16000, 17: 17700, 18: 19400, 19: 21300, 20: 23200,
	21: 25200, 22: 27300, 23: 29400, 24: 31700, 25: 34000, 26: 36400, 27: 38900, 28: 41400, 29: 44300, 30: 47400,
	31: 50800, 32: 54500, 33: 58600, 34: 62800, 35: 67100, 36: 71600, 37: 76100, 38: 80800, 39: 85700, 40: 90700,
	41: 95800, 42: 101000, 43: 106300, 44: 111800, 45: 117500, 46: 123200, 47: 129100, 48: 135100, 49: 141200, 50: 147500,
	51: 153900, 52: 160400, 53: 167100, 54: 173900, 55: 180800, 56: 187900, 57: 195000, 58: 202300, 59: 209800,
}

const (
	maxPlanLevel = 60
	// Forever gives its extra talent points from level 25 on.
	bonusTalentPointsLevel = 25
	maxTalentPoints        = 51
)

// The hours a level takes us, when we get xpPerHour.
func hoursAtLevel(level int32, xpPerHour float64) float64 {
	if xpPerHour <= 0 {
		return inf
	}
	return xpToNextLevel[level] / xpPerHour
}

func talentPointsAt(level, bonusPoints int32) int {
	bonus := int32(0)
	if level >= bonusTalentPointsLevel {
		bonus = bonusPoints
	}
	return int(min(maxTalentPoints, max(0, level-9)+bonus))
}

// The level we put in a build's point number index (0 for the first point), when the build
// starts at startLevel.
func pickLevel(index int, startLevel, bonusPoints int32) int32 {
	for level := max(startLevel, 10); level <= maxPlanLevel; level++ {
		if talentPointsAt(level, bonusPoints) > index {
			return level
		}
	}
	return maxPlanLevel + 1
}

// The levels we sim, from the plan's first level up to, but not including, its last.
func planLevels(plan *proto.LevelingPlanSettings) []int32 {
	var levels []int32
	for level := plan.FromLevel; level < plan.ToLevel; level++ {
		levels = append(levels, level)
	}
	return levels
}

// The builds by level, lowest first. The first one starts at the plan's first level whatever
// its own level says.
func sortedBuilds(plan *proto.LevelingPlanSettings) []*proto.LevelingTalentBuild {
	builds := slices.Clone(plan.Builds)
	slices.SortStableFunc(builds, func(a, b *proto.LevelingTalentBuild) int { return int(a.Level - b.Level) })
	return builds
}

func buildIndexAt(plan *proto.LevelingPlanSettings, level int32) int {
	index := 0
	for i, build := range sortedBuilds(plan) {
		if i > 0 && build.Level <= level {
			index = i
		}
	}
	return index
}

// The plan's talents string at a level, with the build it plays there.
func planTalentsAt(plan *proto.LevelingPlanSettings, level int32, shape *proto.LevelingTalentShape) string {
	builds := sortedBuilds(plan)
	var picks []*proto.LevelingTalentPick
	if len(builds) > 0 {
		picks = builds[buildIndexAt(plan, level)].Picks
	}
	n := min(talentPointsAt(level, plan.BonusTalentPoints), len(picks))
	return countsToTalentsString(countsOf(picks[:n], shape))
}

// The level range we play at a level.
func rangeAt(plan *proto.LevelingPlanSettings, level int32) *proto.LevelingRotation {
	ranges := slices.Clone(plan.Rotations)
	slices.SortStableFunc(ranges, func(a, b *proto.LevelingRotation) int { return int(a.Level - b.Level) })
	if len(ranges) == 0 {
		return nil
	}
	found := ranges[0]
	for _, r := range ranges {
		if r.Level <= level {
			found = r
		}
	}
	return found
}

// The rotation we play at a level, or nil to keep the settings' one.
func rotationAt(plan *proto.LevelingPlanSettings, level int32) *proto.APLRotation {
	r := rangeAt(plan, level)
	if r == nil {
		return nil
	}
	for _, preset := range plan.RotationLibrary {
		if preset.Name == r.Name {
			return preset.Rotation
		}
	}
	return nil
}

// A pick is one talent point: its tree and the talent's place in the tree.
type pick struct{ tree, talent int }

func picksFromProto(picks []*proto.LevelingTalentPick) []pick {
	out := make([]pick, len(picks))
	for i, p := range picks {
		out[i] = pick{int(p.Tree), int(p.Talent)}
	}
	return out
}

func picksToProto(picks []pick) []*proto.LevelingTalentPick {
	out := make([]*proto.LevelingTalentPick, len(picks))
	for i, p := range picks {
		out[i] = &proto.LevelingTalentPick{Tree: int32(p.tree), Talent: int32(p.talent)}
	}
	return out
}

// A route's key, for the sets and maps that need one.
func routeKey(route []pick) string {
	var b strings.Builder
	for _, p := range route {
		b.WriteString(strconv.Itoa(p.tree))
		b.WriteByte(',')
		b.WriteString(strconv.Itoa(p.talent))
		b.WriteByte(';')
	}
	return b.String()
}

func emptyCounts(shape *proto.LevelingTalentShape) [][]int {
	counts := make([][]int, len(shape.Trees))
	for t, tree := range shape.Trees {
		counts[t] = make([]int, len(tree.Talents))
	}
	return counts
}

// countsOf takes either proto picks or our own.
func countsOf[P *proto.LevelingTalentPick | pick](picks []P, shape *proto.LevelingTalentShape) [][]int {
	counts := emptyCounts(shape)
	for _, p := range picks {
		var tree, talent int
		switch v := any(p).(type) {
		case *proto.LevelingTalentPick:
			tree, talent = int(v.Tree), int(v.Talent)
		case pick:
			tree, talent = v.tree, v.talent
		}
		if tree < len(counts) && talent < len(counts[tree]) {
			counts[tree][talent]++
		}
	}
	return counts
}

// The talents string of some counts, like "-0540320010051122".
func countsToTalentsString(counts [][]int) string {
	parts := make([]string, len(counts))
	for t, tree := range counts {
		var b strings.Builder
		for _, n := range tree {
			b.WriteString(strconv.Itoa(n))
		}
		parts[t] = strings.TrimRight(b.String(), "0")
	}
	return strings.TrimRight(strings.Join(parts, "-"), "-")
}

// Whether each pick of a route keeps the tree's rules when we come to it: it has a rank left,
// its row is open and the talent it needs is full.
func pickProblems(route []pick, shape *proto.LevelingTalentShape) []bool {
	counts := emptyCounts(shape)
	treePoints := make([]int, len(shape.Trees))
	problems := make([]bool, len(route))
	for i, p := range route {
		if p.tree >= len(shape.Trees) || p.talent >= len(shape.Trees[p.tree].Talents) {
			problems[i] = true
			continue
		}
		config := shape.Trees[p.tree].Talents[p.talent]
		switch {
		case counts[p.tree][p.talent] >= int(config.MaxPoints):
			problems[i] = true
		case treePoints[p.tree] < int(config.Row*shape.PointsPerRow):
			problems[i] = true
		case config.Prereq >= 0 && counts[p.tree][config.Prereq] < int(shape.Trees[p.tree].Talents[config.Prereq].MaxPoints):
			problems[i] = true
		}
		counts[p.tree][p.talent]++
		treePoints[p.tree]++
	}
	return problems
}

func treeValid(route []pick, shape *proto.LevelingTalentShape) bool {
	return !slices.Contains(pickProblems(route, shape), true)
}

// parseTalents reads a talents string, like 05-0532-1, into each talent's points.
func parseTalents(talents string, shape *proto.LevelingTalentShape) [][]int {
	counts := emptyCounts(shape)
	for t, part := range strings.Split(talents, "-") {
		for i, c := range part {
			if t < len(counts) && i < len(counts[t]) {
				counts[t][i] = int(c - '0')
			}
		}
	}
	return counts
}
