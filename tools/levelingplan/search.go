package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/tools/rotopt/aplkit"
)

// The confirm sims play other sessions than the search's, so the numbers we show can't come
// from the luck of the sessions the search picked the route on.
const confirmSeedOffset = 7_777_777

var errStopped = errors.New("Stopped.")

// runSearch finds the best talent route over the plan's levels and the best respec at each
// respec level:
//  1. The route from the plan's first build. With rotation_search, first the rotations with
//     the plan's talents (rotationSearch), then the route with them, then the rotations of
//     the ranges whose talents the route changed, and so on until the rotations or the route
//     stay the same, or maxRotationPasses.
//  2. A respec search at each respec level inside the plan, over the levels from it on.
//  3. The route again from the respec builds. A respec's build is also a route from the first
//     level, and it can be the better route, like when it only moves a point the route could
//     have put there too. Then a respec only shows what a respec saves.
//  4. The route, the respecs and the plan's own talents and rotations again on other
//     sessions, at confirm_sessions, for the numbers we show.
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
	fail := func(err error) *proto.LevelingPlanSearchResult {
		result.Error = err.Error()
		result.Sims = p.sims.Load()
		return result
	}

	stage := ""
	report := func(text string, best float64) {
		update := &proto.LevelingPlanSearchProgress{Stage: text, Sims: p.sims.Load()}
		if !math.IsInf(best, 0) {
			update.BestHours = best
		}
		progress(update)
	}
	newSearch := func() *routeSearch {
		return &routeSearch{
			shape:    p.shape,
			levels:   levels,
			bonus:    plan.BonusTalentPoints,
			required: plan.RequiredTalents,
			ctx:      ctx,
			cache:    map[string]levelHours{},
			sim: func(jobs []levelJob) ([]levelHours, []bool) {
				return p.simLevels(ctx, jobs, result.Sessions, p.seed)
			},
			progress: func(inner string, best float64) { report(stage+": "+strings.ToLower(inner), best) },
		}
	}

	var start []pick
	if builds := sortedBuilds(plan); len(builds) > 0 {
		start = picksFromProto(builds[0].Picks)
	}
	// The plan as the UI has it, for the numbers we compare against. p.plan becomes the plan
	// with the rotations we found.
	ownPlan := plan
	search := newSearch()
	var route []pick
	var bands []*rotationBand
	if !req.RotationSearch {
		stage = "Route"
		found, err := search.search([][]pick{start}, levels)
		if err != nil {
			return fail(err)
		}
		route = found
	} else {
		template, err := templateFor(req.Base.GetRaid().GetParties()[0].GetPlayers()[0])
		if err != nil {
			return fail(err)
		}
		rs := &rotationSearch{p: p, ctx: ctx, template: template, shape: p.shape, levels: levels}
		defaults := aplkit.Knobs{}
		for _, knob := range template.Knobs() {
			defaults[knob.Name] = knob.Default
		}
		defaultImbue := func(level int32) proto.WeaponImbue {
			if r := rangeAt(plan, level); r != nil {
				return r.Imbue
			}
			return proto.WeaponImbue_WeaponImbueUnknown
		}
		talentsAt := func(level int32) string { return planTalentsAt(plan, level, p.shape) }
		starts := [][]pick{start}
		var before []*rotationBand
		for pass := int32(1); pass <= maxRotationPasses; pass++ {
			stage = fmt.Sprintf("Rotations, pass %d", pass)
			rs.progress = func(text string) { report(stage+": "+text, inf) }
			passBands := rs.planBands(talentsAt)
			tuned, err := rs.tune(passBands, before, defaults, defaultImbue, talentsAt)
			if err != nil {
				return fail(err)
			}
			// The same rotations find the same route again.
			if pass > 1 && tuned == 0 {
				break
			}
			bands, before = passBands, passBands
			p.plan = rs.planWith(ownPlan, bands)
			result.Passes = pass
			stage = fmt.Sprintf("Route, pass %d", pass)
			search = newSearch()
			found, err := search.search(starts, levels)
			if err != nil {
				return fail(err)
			}
			// The same route keeps the rotations' talents, so they'd stay the same.
			same := route != nil && routeKey(found) == routeKey(route)
			route = found
			starts = [][]pick{route}
			talentsAt = func(level int32) string { return search.talentsAt(found, level) }
			if same {
				break
			}
		}
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
		found, err := search.search([][]pick{route, start}, levelsFrom(levels, level))
		if err != nil {
			return fail(err)
		}
		respecs = append(respecs, respec{level, found})
	}
	if len(respecs) > 0 && !search.stopped() {
		stage = "Route again"
		starts := [][]pick{route}
		for _, r := range respecs {
			starts = append(starts, r.route)
		}
		found, err := search.search(starts, levels)
		if err != nil {
			return fail(err)
		}
		route = found
	}
	result.Sims = p.sims.Load()
	if search.stopped() {
		return fail(errStopped)
	}

	report(fmt.Sprintf("Simming the route and the respecs again with %d sessions a level", result.ConfirmSessions), inf)
	// "own" is the plan's own rotations and "found" the ones we found. They're the same
	// without rotation_search.
	ownTag := "own"
	if ownPlan == p.plan {
		ownTag = "found"
	}
	confirm := map[string]levelHours{}
	var jobs []levelJob
	key := func(tag string, level int32, talents string) string { return tag + "|" + cacheKey(level, talents) }
	add := func(tag string, level int32, talents string) {
		k := key(tag, level, talents)
		if _, ok := confirm[k]; ok {
			return
		}
		confirm[k] = levelHours{}
		job := levelJob{level: level, talents: talents}
		if tag == "own" {
			job.plan = ownPlan
		}
		jobs = append(jobs, job)
	}
	planAt := func(level int32) string { return planTalentsAt(plan, level, p.shape) }
	routeAt := func(level int32) string { return search.talentsAt(route, level) }
	for _, level := range levels {
		add(ownTag, level, planAt(level))
		add("found", level, routeAt(level))
		if req.RotationSearch {
			add("found", level, planAt(level))
		}
	}
	for _, r := range respecs {
		for _, level := range levelsFrom(levels, r.level) {
			add("found", level, search.talentsAt(r.route, level))
		}
	}
	results, done := p.simLevels(ctx, jobs, result.ConfirmSessions, p.seed+confirmSeedOffset)
	for i, job := range jobs {
		if !done[i] {
			return fail(errStopped)
		}
		tag := "found"
		if job.plan != nil {
			tag = "own"
		}
		confirm[key(tag, job.level, job.talents)] = results[i]
	}
	// Both play the same sessions, so their noise mostly cancels, and adding it up is the most
	// it can be.
	compare := func(levels []int32, tagA string, a func(level int32) string, tagB string, b func(level int32) string) (hoursA, hoursB, noise float64) {
		variance := 0.0
		for _, level := range levels {
			x, y := confirm[key(tagA, level, a(level))], confirm[key(tagB, level, b(level))]
			hoursA += x.hours
			hoursB += y.hours
			variance += (x.hours*x.relativeError)*(x.hours*x.relativeError) + (y.hours*y.relativeError)*(y.hours*y.relativeError)
		}
		return hoursA, hoursB, 1.96 * math.Sqrt(variance)
	}
	result.Route = picksToProto(route)
	result.RouteHours, result.PlanHours, result.Noise = compare(levels, "found", routeAt, ownTag, planAt)
	if req.RotationSearch {
		result.RotationLibrary = slices.DeleteFunc(slices.Clone(p.plan.RotationLibrary), func(r *proto.LevelingRotationPreset) bool { return !strings.HasPrefix(r.Name, searchedPrefix) })
		result.Rotations = p.plan.Rotations
		result.Bands = bandsToProto(bands)
		result.RotationOnlyHours, _, result.RotationOnlyNoise = compare(levels, "found", planAt, ownTag, planAt)
	}
	for _, r := range respecs {
		respecAt := func(level int32) string { return search.talentsAt(r.route, level) }
		respecHours, keptHours, noise := compare(levelsFrom(levels, r.level), "found", respecAt, "found", routeAt)
		result.Respecs = append(result.Respecs, &proto.LevelingRespecResult{
			Level: r.level, Route: picksToProto(r.route), RespecHours: respecHours, KeptHours: keptHours, Noise: noise,
		})
	}
	result.Seconds = time.Since(started).Seconds()
	return result
}

// The levels from this one on.
func levelsFrom(levels []int32, level int32) []int32 {
	var out []int32
	for _, l := range levels {
		if l >= level {
			out = append(out, l)
		}
	}
	return out
}
