package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"sort"

	"github.com/wowsims/classic/sim/core/proto"
)

// The talent route search (docs/leveling_plan.md, Talent search), ported from the UI's
// leveling_talent_search.ts, which it replaces.
//
// A route is the order we put our talent points in from the plan's first level, one pick a
// point. We score a route by the hours of all its levels together, never one level at a
// time. Taking the talent that's best at level 10 can leave us slower over the whole way, like
// a point that does the most at 10 but keeps us out of a row we want at 20. So a change to the
// route is kept only when the plan's total hours go down.
//
// The search starts from many routes: the plan's own build, one that fills each tree from its
// top row down, and one for each deeper talent that goes straight for it (like Stormstrike at
// 25), with the fewest points before it that its row needs. A change at a time can't get from
// one of those to another, because the points in between are worse, so we score them all and
// search from the best few. From each start we keep making the change to the route that saves
// the most hours over all the levels, until no change saves any. The changes are:
//   - Swapping two picks up to swapWindow points apart, or moving a pick up to that far with
//     the picks in between shifting over, so a talent comes earlier or later.
//   - Taking out the last point of a talent and putting a point in another talent instead, in
//     its place or at the end of the route.
//
// Each level is simmed on its own fixed sessions, so two routes with the same talents at a
// level get the same hours there, and we sim each level and talents only once.
//
// Required talents (like Improved Ghost Wolf for walking) are rules, not something we sim for.
// Every route we look at has them by their level.

const (
	// How far apart two picks can be for a swap.
	swapWindow = 3
	// How many of the point moves that look best at the screening levels we sim over all levels.
	setMovesKept = 6
	// The least a change must save to be kept, so we don't follow the luck of the sessions: the
	// noise of the levels it changes, or 0.05% of the hours when that's less. The noise adds up
	// the noise of each level's sessions, but two routes play the same sessions and most of it
	// cancels. So for a change of many levels it's far more than the real noise, and 0.05% is
	// the bar.
	minGainNoise = 1
	minGain      = 0.0005
	// How many starts we search from, the best ones by their own hours.
	startsSearched = 3
)

// errNoRoute is a search that can't run, like one whose required talents no route can have.
var errNoRoute = errors.New("No talent route has the required talents by their levels. A talent needs its row open by then.")

type routeSearch struct {
	shape    *proto.LevelingTalentShape
	levels   []int32
	bonus    int32
	required []*proto.LevelingTalentRequirement
	// Sims the jobs and gives back each one's hours, or false for one it didn't get to.
	sim      func(jobs []levelJob) ([]levelHours, []bool)
	ctx      context.Context
	progress func(stage string, best float64)
	cache    map[string]levelHours
	sims     int
}

func (s *routeSearch) pointsAt(level int32) int { return talentPointsAt(level, s.bonus) }

func (s *routeSearch) allowed(tree, talent int) bool {
	return !s.shape.Trees[tree].Talents[talent].NotSimulated
}

func (s *routeSearch) stopped() bool { return s.ctx.Err() != nil }

// The number of points a route has, the points of the last level.
func (s *routeSearch) routeLength() int { return s.pointsAt(s.levels[len(s.levels)-1]) }

func (s *routeSearch) talentsAt(route []pick, level int32) string {
	return countsToTalentsString(countsOf(route[:min(s.pointsAt(level), len(route))], s.shape))
}

func cacheKey(level int32, talents string) string { return fmt.Sprintf("%d|%s", level, talents) }

// The hours of each route over the levels. We sim the levels and talents we haven't simmed
// yet, all in one batch.
func (s *routeSearch) score(routes [][]pick, levels []int32) []float64 {
	var jobs []levelJob
	for _, route := range routes {
		for _, level := range levels {
			jobs = append(jobs, levelJob{level: level, talents: s.talentsAt(route, level)})
		}
	}
	s.simMissing(jobs)
	scores := make([]float64, len(routes))
	for i, route := range routes {
		for _, level := range levels {
			scores[i] += s.hoursAt(level, s.talentsAt(route, level))
		}
	}
	return scores
}

// The hours of a level we simmed, or Inf when we stopped before it was.
func (s *routeSearch) hoursAt(level int32, talents string) float64 {
	if r, ok := s.cache[cacheKey(level, talents)]; ok {
		return r.hours
	}
	return inf
}

func (s *routeSearch) simMissing(jobs []levelJob) {
	seen := map[string]bool{}
	var missing []levelJob
	for _, job := range jobs {
		key := cacheKey(job.level, job.talents)
		if _, ok := s.cache[key]; ok || seen[key] {
			continue
		}
		seen[key] = true
		missing = append(missing, job)
	}
	if len(missing) == 0 || s.stopped() {
		return
	}
	results, done := s.sim(missing)
	for i, job := range missing {
		if done[i] {
			s.cache[cacheKey(job.level, job.talents)] = results[i]
		}
	}
	s.sims += len(missing)
}

// The level we check a requirement at: its own level, or the first level we score when that
// comes later (a respec puts all its points in there), or the last when the requirement's
// level is past the plan.
func (s *routeSearch) requirementLevel(r *proto.LevelingTalentRequirement, levels []int32) int32 {
	return min(max(r.Level, levels[0]), levels[len(levels)-1])
}

func countOf(route []pick, tree, talent int) int {
	n := 0
	for _, p := range route {
		if p.tree == tree && p.talent == talent {
			n++
		}
	}
	return n
}

// Whether the route keeps the tree's rules and has every required talent by its level.
func (s *routeSearch) valid(route []pick, levels []int32) bool {
	if !treeValid(route, s.shape) {
		return false
	}
	for _, r := range s.required {
		points := min(s.pointsAt(s.requirementLevel(r, levels)), len(route))
		if countOf(route[:points], int(r.Tree), int(r.Talent)) < int(r.Points) {
			return false
		}
	}
	return true
}

// The route with the required talents in it by their levels. Each missing point goes in as
// late as it can and still make its level, and the route loses its last points to make room.
// Nil when a requirement can't be met, like a talent whose row isn't open by its level.
func (s *routeSearch) withRequired(route []pick, levels []int32) []pick {
	result := slices.Clone(route)
	required := slices.Clone(s.required)
	sort.SliceStable(required, func(a, b int) bool { return required[a].Level < required[b].Level })
	for _, r := range required {
		deadline := s.pointsAt(s.requirementLevel(r, levels))
		target := pick{int(r.Tree), int(r.Talent)}
		// Its points after the deadline come before it instead.
		var kept []pick
		for i, p := range result {
			if i < deadline || p != target {
				kept = append(kept, p)
			}
		}
		result = kept
		have := countOf(result[:min(deadline, len(result))], target.tree, target.talent)
		for have < int(r.Points) {
			placed := false
			for at := min(deadline-1, len(result)); at >= 0 && !placed; at-- {
				next := slices.Concat(result[:at], []pick{target}, result[at:])
				if pickProblems(next[:at+1], s.shape)[at] {
					continue
				}
				// The point we put in pushes the last point before the deadline past it. When
				// that's one of this talent's, another point after ours goes past instead.
				if len(next) > deadline && next[deadline] == target {
					var swapped []pick
					for k := 0; k < deadline-at-1; k++ {
						j := deadline - 1 - k
						if next[j] == target {
							continue
						}
						t := slices.Clone(next)
						t[j], t[deadline] = t[deadline], t[j]
						if treeValid(t[:deadline+1], s.shape) {
							swapped = t
							break
						}
					}
					if swapped == nil {
						continue
					}
					next = swapped
				}
				result = next
				placed = true
			}
			if !placed {
				return nil
			}
			have++
		}
	}
	result = s.extend(result[:min(len(result), s.routeLength())], -1)
	if !s.valid(result, levels) {
		return nil
	}
	return result
}

// A route of routeLength points that starts with the picks. We fill what's missing from the
// top row down, in the tree we name (or the one with the most points, for -1), and then the
// others.
func (s *routeSearch) extend(picks []pick, tree int) []pick {
	route := slices.Clone(picks[:min(len(picks), s.routeLength())])
	first := tree
	if first < 0 {
		counts := countsOf(route, s.shape)
		best := -1
		for t, c := range counts {
			n := 0
			for _, v := range c {
				n += v
			}
			// The TS sorts by points, most first, and the sort keeps the tree order on a tie.
			if best < 0 || n > best {
				first, best = t, n
			}
		}
	}
	trees := []int{first}
	for t := range s.shape.Trees {
		if t != first {
			trees = append(trees, t)
		}
	}
	for len(route) < s.routeLength() {
		found := false
		for _, t := range trees {
			order := make([]int, 0, len(s.shape.Trees[t].Talents))
			for i := range s.shape.Trees[t].Talents {
				if s.allowed(t, i) {
					order = append(order, i)
				}
			}
			talents := s.shape.Trees[t].Talents
			sort.SliceStable(order, func(a, b int) bool {
				if talents[order[a]].Row != talents[order[b]].Row {
					return talents[order[a]].Row < talents[order[b]].Row
				}
				return order[a] < order[b]
			})
			for _, i := range order {
				p := pick{t, i}
				if treeValid(append(slices.Clone(route), p), s.shape) {
					route = append(route, p)
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			break
		}
	}
	return route
}

// A route that goes straight for a deeper talent: the points its row needs in its tree (the
// talents the start has there first), then the talent it needs, then all its ranks, and the
// rest like extend. Nil when the route has too few points to get there.
func (s *routeSearch) rush(tree, talent int, start []pick) []pick {
	config := s.shape.Trees[tree].Talents[talent]
	var route []pick
	add := func(p pick) bool {
		if len(route) >= s.routeLength() || !treeValid(append(slices.Clone(route), p), s.shape) {
			return false
		}
		route = append(route, p)
		return true
	}
	var above []pick
	for _, p := range start {
		if p.tree == tree {
			above = append(above, p)
		}
	}
	talents := s.shape.Trees[tree].Talents
	order := make([]int, len(talents))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool {
		if talents[order[a]].Row != talents[order[b]].Row {
			return talents[order[a]].Row < talents[order[b]].Row
		}
		return order[a] < order[b]
	})
	for _, i := range order {
		for range talents[i].MaxPoints {
			above = append(above, pick{tree, i})
		}
	}
	treePoints := func() int { return countOfTree(route, tree) }
	for _, p := range above {
		c := talents[p.talent]
		if c.Row >= config.Row || int32(p.talent) == config.Prereq || !s.allowed(tree, p.talent) {
			continue
		}
		if treePoints() >= int(config.Row*s.shape.PointsPerRow) {
			break
		}
		add(p)
	}
	if config.Prereq >= 0 {
		for range talents[config.Prereq].MaxPoints {
			add(pick{tree, int(config.Prereq)})
		}
	}
	if !add(pick{tree, talent}) {
		return nil
	}
	for n := int32(1); n < config.MaxPoints; n++ {
		add(pick{tree, talent})
	}
	return s.extend(route, tree)
}

func countOfTree(route []pick, tree int) int {
	n := 0
	for _, p := range route {
		if p.tree == tree {
			n++
		}
	}
	return n
}

// The best route we find over the levels from the starts. Each start is cut or filled to the
// points of the last level.
func (s *routeSearch) search(starts [][]pick, levels []int32) ([]pick, error) {
	var candidates [][]pick
	for _, start := range starts {
		candidates = append(candidates, s.extend(start, -1))
	}
	for t := range s.shape.Trees {
		candidates = append(candidates, s.extend(nil, t))
	}
	var first []pick
	if len(starts) > 0 {
		first = starts[0]
	}
	for t, tree := range s.shape.Trees {
		for i, config := range tree.Talents {
			if config.Row > 0 && s.allowed(t, i) {
				candidates = append(candidates, s.rush(t, i, first))
			}
		}
	}
	seen := map[string]bool{}
	var unique [][]pick
	for _, route := range candidates {
		if route == nil || len(route) != s.routeLength() || !treeValid(route, s.shape) {
			continue
		}
		meeting := s.withRequired(route, levels)
		if meeting == nil || seen[routeKey(meeting)] {
			continue
		}
		seen[routeKey(meeting)] = true
		unique = append(unique, meeting)
	}
	if len(unique) == 0 {
		return nil, errNoRoute
	}

	// We score the starts at every other level, which is enough to tell the good ones.
	s.progress(fmt.Sprintf("Scoring %d starts", len(unique)), inf)
	var startLevels []int32
	for i, level := range levels {
		if i%2 == 0 || i == len(levels)-1 {
			startLevels = append(startLevels, level)
		}
	}
	scores := s.score(unique, startLevels)
	order := make([]int, len(unique))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool { return scores[order[a]] < scores[order[b]] })
	searched := order[:min(startsSearched, len(order))]
	var routes [][]pick
	for _, i := range searched {
		routes = append(routes, unique[i])
	}
	full := s.score(routes, levels)
	best, bestHours := routes[0], full[0]
	for n, route := range routes {
		found, hours := s.improve(route, full[n], levels, fmt.Sprintf("Start %d of %d", n+1, len(routes)), bestHours)
		if hours < bestHours {
			best, bestHours = found, hours
		}
		if s.stopped() {
			break
		}
	}
	return best, nil
}

// Makes the change that saves the most hours, again and again, until none saves any.
func (s *routeSearch) improve(route []pick, hours float64, levels []int32, stage string, bestSoFar float64) ([]pick, float64) {
	current, currentHours := route, hours
	for round := 1; !s.stopped(); round++ {
		s.progress(fmt.Sprintf("%s, change %d", stage, round), math.Min(bestSoFar, currentHours))
		moves := append(s.swaps(current, levels), s.pointMoves(current, levels)...)
		if len(moves) == 0 || s.stopped() {
			break
		}
		scores := s.score(moves, levels)
		bestIndex := -1
		for i, score := range scores {
			bar := math.Min(minGainNoise*s.noiseOfChange(current, moves[i], levels), minGain*currentHours)
			if score < currentHours-bar-1e-9 && (bestIndex < 0 || score < scores[bestIndex]) {
				bestIndex = i
			}
		}
		if bestIndex < 0 || s.stopped() {
			break
		}
		current, currentHours = moves[bestIndex], scores[bestIndex]
	}
	return current, currentHours
}

// The noise of the hours a change saves: that of each level whose talents it changes, added
// up for both routes.
func (s *routeSearch) noiseOfChange(from, to []pick, levels []int32) float64 {
	variance := 0.0
	for _, level := range levels {
		a, b := s.talentsAt(from, level), s.talentsAt(to, level)
		if a == b {
			continue
		}
		for _, talents := range []string{a, b} {
			if r, ok := s.cache[cacheKey(level, talents)]; ok {
				variance += (r.hours * r.relativeError) * (r.hours * r.relativeError)
			}
		}
	}
	return math.Sqrt(variance)
}

// Whether some level we score has the pick at index i but not the one at j (i < j), so a
// swap of the two changes its talents.
func (s *routeSearch) swapMatters(i, j int, levels []int32) bool {
	for _, level := range levels {
		points := s.pointsAt(level)
		if i < points && points <= j {
			return true
		}
	}
	return false
}

func (s *routeSearch) swaps(route []pick, levels []int32) [][]pick {
	var moves [][]pick
	seen := map[string]bool{}
	for i := 0; i < len(route); i++ {
		for j := i + 1; j < min(len(route), i+1+swapWindow); j++ {
			if route[i] == route[j] || !s.swapMatters(i, j, levels) {
				continue
			}
			move := slices.Clone(route)
			move[i], move[j] = move[j], move[i]
			if s.valid(move, levels) {
				moves = append(moves, move)
				seen[routeKey(move)] = true
			}
		}
	}
	// A pick moved to another place, with the picks in between shifting over one. A swap can't
	// take a talent past one that needs it, like Shamanistic Focus past the points that open
	// its row.
	original := routeKey(route)
	for i := 0; i < len(route); i++ {
		for j := max(0, i-swapWindow); j < min(len(route), i+1+swapWindow); j++ {
			if abs(i-j) < 2 || !s.swapMatters(min(i, j), max(i, j), levels) {
				continue
			}
			move := slices.Clone(route)
			p := move[i]
			move = slices.Delete(move, i, i+1)
			move = slices.Insert(move, j, p)
			key := routeKey(move)
			if seen[key] || key == original || !s.valid(move, levels) {
				continue
			}
			seen[key] = true
			moves = append(moves, move)
		}
	}
	return moves
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// The routes that take the last point of a talent out and put a point in another talent
// instead. There are too many to sim each at every level, so we first sim the full build at
// two levels, the middle and the last, without the talent we take from and with the one we
// add. The moves that look best by those (taken as if each point counted on its own) get
// simmed at every level.
func (s *routeSearch) pointMoves(route []pick, levels []int32) [][]pick {
	counts := countsOf(route, s.shape)
	screenLevels := []int32{levels[(len(levels)-1)/2]}
	if last := levels[len(levels)-1]; last != screenLevels[0] {
		screenLevels = append(screenLevels, last)
	}

	type move struct {
		route    []pick
		from, to pick
	}
	var moves []move
	var taken []pick
	for _, p := range route {
		if !slices.Contains(taken, p) {
			taken = append(taken, p)
		}
	}
	for _, from := range taken {
		index := -1
		for i, p := range route {
			if p == from {
				index = i
			}
		}
		for t, tree := range s.shape.Trees {
			for i, config := range tree.Talents {
				to := pick{t, i}
				if !s.allowed(t, i) || to == from || counts[t][i] >= int(config.MaxPoints) {
					continue
				}
				inPlace := slices.Clone(route)
				inPlace[index] = to
				atEnd := append(slices.Concat(route[:index], route[index+1:]), to)
				switch {
				case s.valid(inPlace, levels):
					moves = append(moves, move{inPlace, from, to})
				case s.valid(atEnd, levels):
					moves = append(moves, move{atEnd, from, to})
				}
			}
		}
	}
	if len(moves) == 0 {
		return nil
	}

	// A talent's worth at a screening level: the hours there without its last point, or with
	// one more point.
	withChange := func(level int32, p pick, change int) string {
		changed := countsOf(route[:min(s.pointsAt(level), len(route))], s.shape)
		changed[p.tree][p.talent] = max(0, changed[p.tree][p.talent]+change)
		return countsToTalentsString(changed)
	}
	var jobs []levelJob
	for _, level := range screenLevels {
		jobs = append(jobs, levelJob{level: level, talents: s.talentsAt(route, level)})
		for _, p := range taken {
			jobs = append(jobs, levelJob{level: level, talents: withChange(level, p, -1)})
		}
		for _, m := range moves {
			jobs = append(jobs, levelJob{level: level, talents: withChange(level, m.to, 1)})
		}
	}
	s.simMissing(jobs)
	if s.stopped() {
		return nil
	}
	changes := make([]float64, len(moves))
	for k, m := range moves {
		for _, level := range screenLevels {
			now := s.hoursAt(level, s.talentsAt(route, level))
			changes[k] += s.hoursAt(level, withChange(level, m.from, -1)) - now + (s.hoursAt(level, withChange(level, m.to, 1)) - now)
		}
	}
	order := make([]int, len(moves))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool { return changes[order[a]] < changes[order[b]] })
	var out [][]pick
	for _, k := range order[:min(setMovesKept, len(order))] {
		out = append(out, moves[k].route)
	}
	return out
}
