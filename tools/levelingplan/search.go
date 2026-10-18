package main

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/wowsims/classic/sim/core/proto"
)

// The confirm sims play other sessions than the search's, so the numbers we show can't come
// from the luck of the sessions the search picked the route on.
const confirmSeedOffset = 7_777_777

// runSearch finds the best talent route over the plan's levels and the best respec at each
// respec level, the way the Talents part's search did in the browser:
//  1. The route from the plan's first build.
//  2. A respec search at each respec level inside the plan, over the levels from it on.
//  3. The route again from the respec builds. A respec's build is also a route from the first
//     level, and it can be the better route, like when it only moves a point the route could
//     have put there too. Then a respec only shows what a respec saves.
//  4. The route, the respecs and the plan's own talents again on other sessions, at
//     confirm_sessions, for the numbers we show.
func runSearch(ctx context.Context, req *proto.LevelingPlanSearchRequest, p *planner, progress func(*proto.LevelingPlanSearchProgress)) *proto.LevelingPlanSearchResult {
	started := time.Now()
	plan := req.Plan
	levels := planLevels(plan)
	result := &proto.LevelingPlanSearchResult{
		FromLevel: plan.FromLevel, ToLevel: plan.ToLevel,
		Sessions: max(req.Sessions, 10), ConfirmSessions: max(req.ConfirmSessions, 1),
	}
	if len(levels) == 0 {
		result.Error = "The plan has no levels."
		return result
	}

	stage := ""
	search := &routeSearch{
		shape:    p.shape,
		levels:   levels,
		bonus:    plan.BonusTalentPoints,
		required: plan.RequiredTalents,
		ctx:      ctx,
		cache:    map[string]levelHours{},
		sim: func(jobs []levelJob) ([]levelHours, []bool) {
			return p.simLevels(ctx, jobs, result.Sessions, p.seed)
		},
	}
	search.progress = func(inner string, best float64) {
		update := &proto.LevelingPlanSearchProgress{Stage: stage + ": " + strings.ToLower(inner), Sims: int32(search.sims)}
		if !math.IsInf(best, 0) {
			update.BestHours = best
		}
		progress(update)
	}

	var start []pick
	if builds := sortedBuilds(plan); len(builds) > 0 {
		start = picksFromProto(builds[0].Picks)
	}
	stage = "Route"
	route, err := search.search([][]pick{start}, levels)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	type respec struct {
		level int32
		route []pick
	}
	var respecs []respec
	for _, level := range req.RespecLevels {
		if level <= plan.FromLevel || level >= plan.ToLevel || search.stopped() {
			continue
		}
		stage = fmt.Sprintf("Respec at %d", level)
		var from []int32
		for _, l := range levels {
			if l >= level {
				from = append(from, l)
			}
		}
		found, err := search.search([][]pick{route, start}, from)
		if err != nil {
			result.Error = err.Error()
			return result
		}
		respecs = append(respecs, respec{level, found})
	}
	if len(respecs) > 0 && !search.stopped() {
		stage = "Route again"
		starts := [][]pick{route}
		for _, r := range respecs {
			starts = append(starts, r.route)
		}
		if route, err = search.search(starts, levels); err != nil {
			result.Error = err.Error()
			return result
		}
	}
	result.Sims = int32(search.sims)
	if search.stopped() {
		result.Error = "Stopped."
		return result
	}

	progress(&proto.LevelingPlanSearchProgress{Stage: fmt.Sprintf("Simming the route and the respecs again with %d sessions a level", result.ConfirmSessions), Sims: int32(search.sims)})
	confirm := map[string]levelHours{}
	var jobs []levelJob
	add := func(level int32, talents string) {
		key := cacheKey(level, talents)
		if _, ok := confirm[key]; !ok {
			confirm[key] = levelHours{}
			jobs = append(jobs, levelJob{level, talents})
		}
	}
	for _, level := range levels {
		add(level, planTalentsAt(plan, level, p.shape))
		add(level, search.talentsAt(route, level))
	}
	for _, r := range respecs {
		for _, level := range levels {
			if level >= r.level {
				add(level, search.talentsAt(r.route, level))
			}
		}
	}
	results, done := p.simLevels(ctx, jobs, result.ConfirmSessions, p.seed+confirmSeedOffset)
	for i, job := range jobs {
		if !done[i] {
			result.Error = "Stopped."
			return result
		}
		confirm[cacheKey(job.level, job.talents)] = results[i]
	}
	// Both play the same sessions, so their noise mostly cancels, and adding it up is the most
	// it can be.
	compare := func(levels []int32, a, b func(level int32) string) (hoursA, hoursB, noise float64) {
		variance := 0.0
		for _, level := range levels {
			x, y := confirm[cacheKey(level, a(level))], confirm[cacheKey(level, b(level))]
			hoursA += x.hours
			hoursB += y.hours
			variance += (x.hours*x.relativeError)*(x.hours*x.relativeError) + (y.hours*y.relativeError)*(y.hours*y.relativeError)
		}
		return hoursA, hoursB, 1.96 * math.Sqrt(variance)
	}
	routeAt := func(level int32) string { return search.talentsAt(route, level) }
	result.Route = picksToProto(route)
	result.RouteHours, result.PlanHours, result.Noise = compare(levels, routeAt, func(level int32) string { return planTalentsAt(plan, level, p.shape) })
	for _, r := range respecs {
		var from []int32
		for _, l := range levels {
			if l >= r.level {
				from = append(from, l)
			}
		}
		respecHours, keptHours, noise := compare(from, func(level int32) string { return search.talentsAt(r.route, level) }, routeAt)
		result.Respecs = append(result.Respecs, &proto.LevelingRespecResult{
			Level: r.level, Route: picksToProto(r.route), RespecHours: respecHours, KeptHours: keptHours, Noise: noise,
		})
	}
	result.Seconds = time.Since(started).Seconds()
	return result
}
