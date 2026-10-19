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

// confirmNoise is the most noise we want on each number the confirm sims show, in hours.
//
// The confirm sims only measure the route, the respecs and the rotations the search already
// picked, so it doesn't change what we find. ±0.1h is about 6 minutes on a 10 to 60 plan of
// about 97 hours, and it takes about 530 sessions a level there.
//
// We start with confirmFirstSessions a level, and sim more sessions only while a number's noise
// is above confirmNoise, up to the request's confirm_sessions.
const confirmNoise = 0.1
const confirmFirstSessions = 200

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
//     sessions, for the numbers we show. We sim more sessions until their noise is at most
//     confirmNoise, up to confirm_sessions.
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
			progress:   func(inner string, best float64) { report(stage+": "+strings.ToLower(inner), best) },
			bigChanges: true,
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
	jobKey := func(job levelJob) string {
		if job.plan != nil {
			return key("own", job.level, job.talents)
		}
		return key("found", job.level, job.talents)
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
	worstNoise := func() float64 {
		_, _, worst := compare(levels, "found", routeAt, ownTag, planAt)
		if req.RotationSearch {
			_, _, noise := compare(levels, "found", planAt, ownTag, planAt)
			worst = max(worst, noise)
		}
		for _, r := range respecs {
			respecAt := func(level int32) string { return search.talentsAt(r.route, level) }
			_, _, noise := compare(levelsFrom(levels, r.level), "found", respecAt, "found", routeAt)
			worst = max(worst, noise)
		}
		return worst
	}
	sessions := int32(0)
	for round := int64(1); sessions < result.ConfirmSessions; round++ {
		want := min(result.ConfirmSessions, confirmFirstSessions)
		if sessions > 0 {
			noise := worstNoise()
			if noise <= confirmNoise {
				break
			}
			// The noise goes down with the square root of the sessions. We aim a bit lower, so
			// a second round is almost always the last.
			want = min(result.ConfirmSessions, int32(math.Ceil(float64(sessions)*1.1*(noise/confirmNoise)*(noise/confirmNoise))))
		}
		report(fmt.Sprintf("Simming the route and the respecs again, %d sessions a level", want), inf)
		// Each round plays new sessions, which both sides of a comparison share again.
		results, done := p.simLevels(ctx, jobs, want-sessions, p.seed+round*confirmSeedOffset)
		for i, job := range jobs {
			if !done[i] {
				return fail(errStopped)
			}
			k := jobKey(job)
			if sessions == 0 {
				confirm[k] = results[i]
			} else {
				confirm[k] = confirm[k].merged(sessions, results[i], want-sessions)
			}
		}
		sessions = want
	}
	result.ConfirmSessions = sessions
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
