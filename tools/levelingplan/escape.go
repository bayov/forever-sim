package main

import (
	"fmt"
	"math"
	"slices"
	"sort"
)

// Bigger changes to a route that no small change improves (docs/leveling_plan.md, Talent
// search).
//
// The small changes move one pick at most swapWindow places. They can't take several points far
// at once, like Tidal Focus 4 and 5 from 56 and 59 to 38 with Mindfulness after them, because
// each step on the way is worse than the route we have. So once the best start stops improving,
// we price bigger changes with a worth table, sim only the few that look best, and search on
// from one that saves hours.
//
// The worth table is the hours at a few screening levels with one point more in each talent we
// could take there, and one point less in each we have, the tree rules ignored. A change's worth
// is about the sum, over the levels, of the points it adds and takes away there, each at its
// worth at the nearest screening level. Only the table costs sims, and a change gets simmed at
// every level only when the table says it saves hours.
//
// The changes:
//   - All of a talent's points moved together to any place in the route.
//   - From any level, going straight for a talent: the points its row needs in its tree (the
//     route's own picks there first), the talent it needs, and its ranks, then the rest of the
//     route.

const (
	// The levels we sim the worth table at.
	escapeScreenLevels = 6
	// How many of the changes that look best by the table we sim at every level.
	escapeKept = 4
	// How many times we take a bigger change and search on from it.
	escapeRounds = 3
)

// Takes the bigger change that saves the most hours, searches on from it with the small
// changes, and does that again, until no bigger change saves any.
func (s *routeSearch) escape(route []pick, hours float64, levels []int32) ([]pick, float64) {
	for round := 1; round <= escapeRounds && !s.stopped(); round++ {
		s.progress(fmt.Sprintf("Bigger changes, round %d", round), hours)
		moves := s.escapeMoves(route, levels)
		if len(moves) == 0 || s.stopped() {
			break
		}
		scores := s.score(moves, levels)
		best := -1
		for i, score := range scores {
			bar := math.Min(minGainNoise*s.noiseOfChange(route, moves[i], levels), minGain*hours)
			if score < hours-bar-1e-9 && (best < 0 || score < scores[best]) {
				best = i
			}
		}
		if best < 0 || s.stopped() {
			break
		}
		route, hours = s.improve(moves[best], scores[best], levels, fmt.Sprintf("After bigger change %d", round), hours)
	}
	return route, hours
}

// The bigger changes that look best by the worth table, at most escapeKept of them.
func (s *routeSearch) escapeMoves(route []pick, levels []int32) [][]pick {
	var screen []int32
	for k := range escapeScreenLevels {
		level := levels[k*(len(levels)-1)/max(1, escapeScreenLevels-1)]
		if !slices.Contains(screen, level) {
			screen = append(screen, level)
		}
	}

	// The table: the share of the hours one point more (gain) or one point less (loss) saves or
	// costs at each screening level. NaN where the talent has no rank left, or none to take out.
	changed := func(level int32, tree, talent, change int) string {
		counts := countsOf(route[:min(s.pointsAt(level), len(route))], s.shape)
		counts[tree][talent] += change
		return countsToTalentsString(counts)
	}
	var jobs []levelJob
	for _, level := range screen {
		jobs = append(jobs, levelJob{level: level, talents: s.talentsAt(route, level)})
		counts := countsOf(route[:min(s.pointsAt(level), len(route))], s.shape)
		for t, tree := range s.shape.Trees {
			for i, config := range tree.Talents {
				if !s.allowed(t, i) {
					continue
				}
				if counts[t][i] < int(config.MaxPoints) {
					jobs = append(jobs, levelJob{level: level, talents: changed(level, t, i, 1)})
				}
				if counts[t][i] > 0 {
					jobs = append(jobs, levelJob{level: level, talents: changed(level, t, i, -1)})
				}
			}
		}
	}
	s.simMissing(jobs)
	if s.stopped() {
		return nil
	}
	gain := make([][][]float64, len(screen))
	loss := make([][][]float64, len(screen))
	for k, level := range screen {
		base := s.hoursAt(level, s.talentsAt(route, level))
		counts := countsOf(route[:min(s.pointsAt(level), len(route))], s.shape)
		gain[k], loss[k] = make([][]float64, len(s.shape.Trees)), make([][]float64, len(s.shape.Trees))
		for t, tree := range s.shape.Trees {
			gain[k][t], loss[k][t] = make([]float64, len(tree.Talents)), make([]float64, len(tree.Talents))
			for i, config := range tree.Talents {
				gain[k][t][i], loss[k][t][i] = math.NaN(), math.NaN()
				if !s.allowed(t, i) {
					gain[k][t][i], loss[k][t][i] = 0, 0
					continue
				}
				if counts[t][i] < int(config.MaxPoints) {
					gain[k][t][i] = (base - s.hoursAt(level, changed(level, t, i, 1))) / base
				}
				if counts[t][i] > 0 {
					loss[k][t][i] = (s.hoursAt(level, changed(level, t, i, -1)) - base) / base
				}
			}
		}
	}
	// A talent's worth at a level, from the nearest screening level that has it.
	nearest := func(table [][][]float64, level int32, t, i int) float64 {
		order := make([]int, len(screen))
		for k := range order {
			order[k] = k
		}
		sort.SliceStable(order, func(a, b int) bool {
			return abs(int(screen[order[a]]-level)) < abs(int(screen[order[b]]-level))
		})
		for _, k := range order {
			if v := table[k][t][i]; !math.IsNaN(v) {
				return v
			}
		}
		return 0
	}
	type worth struct{ gain, loss [][]float64 }
	byLevel := make([]worth, len(levels))
	baseHours := make([]float64, len(levels))
	routeCounts := make([][][]int, len(levels))
	for li, level := range levels {
		w := worth{make([][]float64, len(s.shape.Trees)), make([][]float64, len(s.shape.Trees))}
		for t, tree := range s.shape.Trees {
			w.gain[t], w.loss[t] = make([]float64, len(tree.Talents)), make([]float64, len(tree.Talents))
			for i := range tree.Talents {
				w.gain[t][i], w.loss[t][i] = nearest(gain, level, t, i), nearest(loss, level, t, i)
			}
		}
		byLevel[li] = w
		baseHours[li] = s.hoursAt(level, s.talentsAt(route, level))
		routeCounts[li] = countsOf(route[:min(s.pointsAt(level), len(route))], s.shape)
	}
	// The hours a change saves (less than 0) or costs, by the table.
	estimate := func(move []pick) float64 {
		total := 0.0
		for li, level := range levels {
			counts := countsOf(move[:min(s.pointsAt(level), len(move))], s.shape)
			share := 0.0
			for t := range counts {
				for i, n := range counts[t] {
					switch d := n - routeCounts[li][t][i]; {
					case d > 0:
						share -= float64(d) * byLevel[li].gain[t][i]
					case d < 0:
						share -= float64(d) * byLevel[li].loss[t][i]
					}
				}
			}
			total += share * baseHours[li]
		}
		return total
	}

	original := routeKey(route)
	seen := map[string]bool{}
	type candidate struct {
		route    []pick
		estimate float64
	}
	var candidates []candidate
	add := func(move []pick) {
		if move == nil || len(move) != s.routeLength() {
			return
		}
		key := routeKey(move)
		if key == original || seen[key] || !s.valid(move, levels) {
			return
		}
		seen[key] = true
		if e := estimate(move); e < 0 {
			candidates = append(candidates, candidate{move, e})
		}
	}
	// All of a talent's points moved together.
	var taken []pick
	for _, p := range route {
		if !slices.Contains(taken, p) {
			taken = append(taken, p)
		}
	}
	for _, p := range taken {
		n := countOf(route, p.tree, p.talent)
		var rest []pick
		for _, q := range route {
			if q != p {
				rest = append(rest, q)
			}
		}
		for j := 0; j <= len(rest); j++ {
			add(slices.Concat(rest[:j], slices.Repeat([]pick{p}, n), rest[j:]))
		}
	}
	// Going straight for a talent from a place in the route.
	for at := range route {
		for t, tree := range s.shape.Trees {
			for i := range tree.Talents {
				if s.allowed(t, i) {
					add(s.rushFrom(route, at, t, i))
				}
			}
		}
	}

	sort.SliceStable(candidates, func(a, b int) bool { return candidates[a].estimate < candidates[b].estimate })
	var out [][]pick
	for _, c := range candidates[:min(escapeKept, len(candidates))] {
		out = append(out, c.route)
	}
	return out
}

// The route up to at, then straight for the talent: the points its row needs in its tree (the
// route's own later picks there first, then from the top row down), the talent it needs, and
// all its ranks, then the rest of the route's picks in their order. Nil when that adds no
// point of the talent before where the route had it.
func (s *routeSearch) rushFrom(route []pick, at, tree, talent int) []pick {
	talents := s.shape.Trees[tree].Talents
	config := talents[talent]
	current := slices.Clone(route[:at])
	rest := slices.Clone(route[at:])
	take := func(p pick) bool {
		if !treeValid(append(slices.Clone(current), p), s.shape) {
			return false
		}
		current = append(current, p)
		if i := slices.Index(rest, p); i >= 0 {
			rest = slices.Delete(rest, i, i+1)
		}
		return true
	}
	order := make([]int, len(talents))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool { return talents[order[a]].Row < talents[order[b]].Row })
	for countOfTree(current, tree) < int(config.Row*s.shape.PointsPerRow) {
		found := false
		for _, p := range slices.Clone(rest) {
			if p.tree == tree && p.talent != talent && talents[p.talent].Row < config.Row && take(p) {
				found = true
				break
			}
		}
		for _, i := range order {
			if found {
				break
			}
			if i != talent && talents[i].Row < config.Row && s.allowed(tree, i) && take(pick{tree, i}) {
				found = true
			}
		}
		if !found {
			return nil
		}
	}
	if config.Prereq >= 0 {
		for countOf(current, tree, int(config.Prereq)) < int(talents[config.Prereq].MaxPoints) {
			if !take(pick{tree, int(config.Prereq)}) {
				return nil
			}
		}
	}
	before := countOf(route[:min(len(current), len(route))], tree, talent)
	for countOf(current, tree, talent) < int(config.MaxPoints) && take(pick{tree, talent}) {
	}
	if countOf(current, tree, talent) <= before {
		return nil
	}
	move := append(current, rest...)
	return s.extend(move[:min(len(move), s.routeLength())], -1)
}
