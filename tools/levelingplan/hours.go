package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	goproto "google.golang.org/protobuf/proto"
)

// The hours of a plan, over several gear histories (-hours), to compare plans with.
//
// One gear seed is one random history of finds, and a lucky weapon moves a plan's hours more than
// the sessions' noise does. So we play the plan with several gear seeds and average. Every plan
// plays the same gear seeds and the same session seeds, so the differences between plans are
// paired: a change in the rotation or the talents doesn't also get other finds.

const bandLevels = 10

// hoursOptions are what every plan plays.
type hoursOptions struct {
	// Gear seeds firstGearSeed, firstGearSeed+1 and so on.
	firstGearSeed, gearSeeds int32
	sessions                 int32
	seed                     int64
}

// The mean over the gear seeds and its standard error.
type meanError struct {
	Mean  float64 `json:"mean"`
	Error float64 `json:"se"`
}

func meanErrorOf(values []float64) meanError {
	n := float64(len(values))
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	out := meanError{Mean: sum / n}
	if len(values) < 2 {
		return out
	}
	squares := 0.0
	for _, v := range values {
		squares += (v - out.Mean) * (v - out.Mean)
	}
	out.Error = math.Sqrt(squares/(n-1)) / math.Sqrt(n)
	return out
}

// hoursRow is a level or a band of levels.
type hoursRow struct {
	Level int32     `json:"level"`
	To    int32     `json:"to,omitempty"`
	Hours meanError `json:"hours"`
	// The level's part of the plan's hours.
	HoursShare float64 `json:"hoursShare"`
	XpPerHour  float64 `json:"xpPerHour"`
	// The deaths while we play the level, not a session's.
	Deaths      meanError `json:"deaths"`
	MobsPerPull float64   `json:"mobsPerPull"`
	MobsPerHour float64   `json:"mobsPerHour"`
	// Parts of the time, of 1. Searching counts walking up to a mob and waiting to leave combat.
	CombatShare float64 `json:"combatShare"`
	RestShare   float64 `json:"restShare"`
	SearchShare float64 `json:"searchShare"`
	DeadShare   float64 `json:"deadShare"`
	Food        float64 `json:"food"`
	Drinks      float64 `json:"drinks"`
	Bandages    float64 `json:"bandages"`
}

type planHours struct {
	Plan      string    `json:"plan"`
	GearSeeds []int32   `json:"gearSeeds"`
	Sessions  int32     `json:"sessions"`
	Seed      int64     `json:"seed"`
	Hours     meanError `json:"hours"`
	// The hours on each gear seed, in the order of gearSeeds.
	HoursBySeed []float64 `json:"hoursBySeed"`
	// The hours less the first plan's, on the same gear seeds and sessions.
	AgainstFirst *meanError `json:"againstFirst,omitempty"`
	Deaths       meanError  `json:"deaths"`
	Bands        []hoursRow `json:"bands"`
	Levels       []hoursRow `json:"levels"`
}

// One sim: a level with a gear seed, and what the level took.
type hoursSim struct {
	level   int32
	hours   float64
	metrics *proto.LevelingMetrics
	// The seconds of a session.
	session float64
}

// hoursPlan plays a plan with every gear seed. The result has the sims by gear seed and then by
// level.
type hoursPlan struct {
	name   string
	req    *proto.LevelingPlanSearchRequest
	levels []int32
	opts   hoursOptions
	sims   [][]hoursSim
}

// withGearSeed is the request with another gear history.
func withGearSeed(req *proto.LevelingPlanSearchRequest, seed int32) *proto.LevelingPlanSearchRequest {
	out := goproto.Clone(req).(*proto.LevelingPlanSearchRequest)
	if out.GearPlan == nil {
		out.GearPlan = &proto.LevelingGearPlanSettings{}
	}
	out.GearPlan.Seed = seed
	// The UI's gear is its own seed's.
	out.UiGear = nil
	return out
}

func (h *hoursPlan) seeds() []int32 {
	var out []int32
	for i := range h.opts.gearSeeds {
		out = append(out, h.opts.firstGearSeed+i)
	}
	return out
}

func (h *hoursPlan) run(ctx context.Context, db *proto.UIDatabase) error {
	if err := routeProblem(h.req.Plan, h.req.TalentShape); err != nil {
		return err
	}
	// Every planner first, since making one adds its plain items to the item database, which the
	// sims read without a lock.
	var planners []*planner
	for _, seed := range h.seeds() {
		p, err := newPlanner(withGearSeed(h.req, seed), db)
		if err != nil {
			return err
		}
		planners = append(planners, p)
	}
	h.sims = make([][]hoursSim, len(planners))
	for i := range h.sims {
		h.sims[i] = make([]hoursSim, len(h.levels))
	}
	errs := make([]error, len(planners)*len(h.levels))
	parallel(ctx, len(errs), func(i int) {
		g, l := i/len(h.levels), i%len(h.levels)
		level := h.levels[l]
		// Each gear seed plays its own sessions too, so their noise averages out as well.
		seed := levelSeed(h.opts.seed+int64(g)*104_729, level)
		h.sims[g][l], errs[i] = planners[g].simLevel(level, h.opts.sessions, seed)
		if errs[i] != nil {
			errs[i] = fmt.Errorf("level %d with gear seed %d: %w", level, h.seeds()[g], errs[i])
		}
	})
	if ctx.Err() != nil {
		return errStopped
	}
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

// simLevel plays the plan's level.
func (p *planner) simLevel(level, sessions int32, seed int64) (hoursSim, error) {
	request := p.request(levelJob{level: level, talents: planTalentsAt(p.plan, level, p.shape)}, sessions, seed)
	res := core.RunRaidSim(request)
	if res.Error != nil {
		return hoursSim{}, fmt.Errorf("%s", firstLine(res.Error.Message))
	}
	m := res.EncounterMetrics.GetLeveling()
	if m == nil {
		return hoursSim{}, fmt.Errorf("the request has no leveling session")
	}
	out := hoursSim{level: level, hours: hoursAtLevel(level, m.XpPerHour), metrics: m, session: m.Xp / m.XpPerHour * 3600}
	if math.IsInf(out.hours, 0) {
		return out, fmt.Errorf("we killed nothing in %d sessions", sessions)
	}
	p.sims.Add(1)
	return out, nil
}

func firstLine(text string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(text), "\n")
	return line
}

// row sums the levels, one by one or as a band.
func (h *hoursPlan) row(levels []int32, planTotal float64) hoursRow {
	out := hoursRow{Level: levels[0]}
	if len(levels) > 1 {
		out.To = levels[len(levels)-1]
	}
	var hours, deaths []float64
	// The rest goes by the hours we spend, so a band's rest share is the share of its time.
	var total, xp, mobs, pulls float64
	for g := range h.sims {
		seedHours, seedDeaths := 0.0, 0.0
		for _, s := range h.sims[g] {
			if s.level < levels[0] || s.level > levels[len(levels)-1] {
				continue
			}
			m := s.metrics
			sessions := s.hours * 3600 / s.session
			seedHours += s.hours
			seedDeaths += m.Deaths * sessions
			total += s.hours
			xp += m.Xp * sessions
			mobs += m.MobsKilled * sessions
			pulls += m.MobsPerPull * s.hours
			out.CombatShare += m.SecondsInCombat / s.session * s.hours
			out.RestShare += m.SecondsResting / s.session * s.hours
			out.SearchShare += m.SecondsSearching / s.session * s.hours
			out.DeadShare += m.SecondsDead / s.session * s.hours
			out.Food += m.Food * sessions / float64(len(h.sims))
			out.Drinks += m.Drinks * sessions / float64(len(h.sims))
			out.Bandages += m.Bandages * sessions / float64(len(h.sims))
		}
		hours, deaths = append(hours, seedHours), append(deaths, seedDeaths)
	}
	out.Hours, out.Deaths = meanErrorOf(hours), meanErrorOf(deaths)
	out.HoursShare = out.Hours.Mean / planTotal
	out.XpPerHour, out.MobsPerHour, out.MobsPerPull = xp/total, mobs/total, pulls/total
	for _, share := range []*float64{&out.CombatShare, &out.RestShare, &out.SearchShare, &out.DeadShare} {
		*share /= total
	}
	return out
}

func (h *hoursPlan) report() *planHours {
	out := &planHours{Plan: h.name, GearSeeds: h.seeds(), Sessions: h.opts.sessions, Seed: h.opts.seed}
	var deaths []float64
	for _, sims := range h.sims {
		total, dead := 0.0, 0.0
		for _, s := range sims {
			total += s.hours
			dead += s.metrics.Deaths * s.hours * 3600 / s.session
		}
		out.HoursBySeed = append(out.HoursBySeed, total)
		deaths = append(deaths, dead)
	}
	out.Hours, out.Deaths = meanErrorOf(out.HoursBySeed), meanErrorOf(deaths)
	var band []int32
	for i, level := range h.levels {
		out.Levels = append(out.Levels, h.row([]int32{level}, out.Hours.Mean))
		band = append(band, level)
		if i+1 == len(h.levels) || h.levels[i+1]/bandLevels != level/bandLevels {
			out.Bands = append(out.Bands, h.row(band, out.Hours.Mean))
			band = nil
		}
	}
	return out
}

// against is this plan's hours less the first plan's, gear seed by gear seed.
func (r *planHours) against(first *planHours) *meanError {
	var diffs []float64
	for i := range r.HoursBySeed {
		diffs = append(diffs, r.HoursBySeed[i]-first.HoursBySeed[i])
	}
	m := meanErrorOf(diffs)
	return &m
}

// hoursFiles plays each plan file and writes the results as a JSON list.
func hoursFiles(db *proto.UIDatabase, files []string, out string, opts hoursOptions, levels []int32) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	var results []*planHours
	for _, file := range files {
		req, err := readPlanRequest(file)
		if err != nil {
			return err
		}
		h := &hoursPlan{name: strings.TrimSuffix(filepath.Base(file), ".json"), req: req, opts: opts}
		for _, level := range planLevels(req.Plan) {
			if len(levels) == 0 || level >= levels[0] && level <= levels[1] {
				h.levels = append(h.levels, level)
			}
		}
		if len(h.levels) == 0 {
			return fmt.Errorf("%s: the plan has none of the levels", file)
		}
		if err := h.run(ctx, db); err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
		r := h.report()
		if len(results) > 0 && results[0].Levels[0].Level == r.Levels[0].Level && len(results[0].Levels) == len(r.Levels) {
			r.AgainstFirst = r.against(results[0])
		}
		results = append(results, r)
		fmt.Fprintln(os.Stderr, r.summary())
	}
	data, err := json.MarshalIndent(results, "", "\t")
	if err != nil {
		return err
	}
	data = roundNumbers(data)
	if out == "" {
		_, err = os.Stdout.Write(append(data, '\n'))
		return err
	}
	return os.WriteFile(out, append(data, '\n'), 0o644)
}

var longNumber = regexp.MustCompile(`-?\d+\.\d{5,}(e-?\d+)?`)

// roundNumbers writes the JSON's long numbers with 4 digits after the point.
func roundNumbers(data []byte) []byte {
	return longNumber.ReplaceAllFunc(data, func(number []byte) []byte {
		x, err := strconv.ParseFloat(string(number), 64)
		if err != nil {
			return number
		}
		return strconv.AppendFloat(nil, round(x, 4), 'f', -1, 64)
	})
}

func (r *planHours) summary() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s: %.2fh ± %.2f over %d gear seeds, %.1f deaths", r.Plan, r.Hours.Mean, r.Hours.Error, len(r.GearSeeds), r.Deaths.Mean)
	if r.AgainstFirst != nil {
		fmt.Fprintf(&b, ", %+.2fh ± %.2f against the first plan", r.AgainstFirst.Mean, r.AgainstFirst.Error)
	}
	for _, band := range r.Bands {
		to := max(band.To, band.Level)
		fmt.Fprintf(&b, "\n  %2d to %2d: %6.2fh ± %.2f, %.0f XP/h, %.2f mobs a pull, %.0f%% resting, %.1f deaths",
			band.Level, to, band.Hours.Mean, band.Hours.Error, band.XpPerHour, band.MobsPerPull, band.RestShare*100, band.Deaths.Mean)
	}
	return b.String()
}

// readPlanRequest reads a request, or makes the one of a config (config.go).
func readPlanRequest(path string) (*proto.LevelingPlanSearchRequest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if _, isConfig := fields["player"]; isConfig {
		return requestFromConfig(path)
	}
	return readRequest(data)
}

// parallel calls run with 0 to n-1 on all the cores, until ctx ends.
func parallel(ctx context.Context, n int, run func(i int)) {
	var next atomic.Int64
	var wg sync.WaitGroup
	for range min(runtime.NumCPU(), n) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ctx.Err() == nil {
				i := int(next.Add(1) - 1)
				if i >= n {
					return
				}
				run(i)
			}
		}()
	}
	wg.Wait()
}
