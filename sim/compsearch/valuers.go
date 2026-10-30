package compsearch

import (
	"fmt"
	"math"
)

// modelValuer values a seat with the model, without sims.
type modelValuer struct{ m *Model }

func (v modelValuer) V(b *Build, s Seat) (float64, float64, string) { return v.m.V(b, s), 0, "" }

// exactValuer values a seat with a single-player sim. One it doesn't have yet is recorded and valued by
// the model, so a fill can be run, its sims run in one batch, and the fill run again (EvalExact).
type exactValuer struct {
	sc      *Scenario
	iters   int32
	model   *Model
	vals    map[string]*Job
	missing map[string]*Job
}

func newExactValuer(sc *Scenario, iters int32, model *Model) *exactValuer {
	return &exactValuer{sc: sc, iters: iters, model: model, vals: map[string]*Job{}, missing: map[string]*Job{}}
}

func askKey(b *Build, s Seat) string { return b.ID + "#" + s.key() }

func (v *exactValuer) V(b *Build, s Seat) (float64, float64, string) {
	if v.sc.HealthLength > 0 {
		return v.lerp(b, s)
	}
	return v.one(b, s, v.sc.Fight)
}

func (v *exactValuer) one(b *Build, s Seat, fight Fight) (float64, float64, string) {
	k := askKey(b, s) + fmt.Sprint(fight.Duration)
	if j, ok := v.vals[k]; ok {
		return j.Value.DPS, j.Value.SE, j.key
	}
	if _, ok := v.missing[k]; !ok {
		j := v.sc.job(b, s, v.iters)
		j.Fight = fight
		v.missing[k] = j
	}
	// Until it's simmed: the model's guess, or nothing without a model (a caller that only reads values
	// after a flush).
	if v.model == nil {
		return 0, 0, ""
	}
	return v.model.V(b, s), 0, ""
}

// lerp is a health fight's value (docs/raid_sim.md, Fight length): the single-player sims run on a
// 5-second grid of fixed lengths, so the cache hits across sets, and a build's DPS is interpolated
// linearly between the two grid points around the fight's length.
func (v *exactValuer) lerp(b *Build, s Seat) (float64, float64, string) {
	L := v.sc.HealthLength
	lo := math.Floor(L/lengthGrid) * lengthGrid
	w := (L - lo) / lengthGrid
	f1, f2 := v.sc.Fight, v.sc.Fight
	f1.Duration, f2.Duration = lo, lo+lengthGrid
	f1.DurationVariation, f2.DurationVariation = 0, 0
	d1, se1, k1 := v.one(b, s, f1)
	if w == 0 {
		return d1, se1, k1
	}
	d2, se2, k2 := v.one(b, s, f2)
	key := ""
	if k1 != "" && k2 != "" {
		key = k1 + k2
	}
	return (1-w)*d1 + w*d2, (1-w)*se1 + w*se2, key
}

const lengthGrid = 5.0

// flush sims everything recorded as missing.
func (v *exactValuer) flush() (int, error) {
	if len(v.missing) == 0 {
		return 0, nil
	}
	jobs := make([]*Job, 0, len(v.missing))
	for _, j := range v.missing {
		jobs = append(jobs, j)
	}
	if err := v.sc.sim(jobs); err != nil {
		return 0, err
	}
	n := len(v.missing)
	for k, j := range v.missing {
		v.vals[k] = j
	}
	v.missing = map[string]*Job{}
	return n, nil
}

// EvalExact fills a plan with exact values: fill, sim what it asked for, and fill again until it asks for
// nothing new.
func (sc *Scenario) EvalExact(r *Raid, p Plan, f *Fillers, v *exactValuer) (*Comp, error) {
	for round := 0; ; round++ {
		comp := sc.Fill(r, p, f, v)
		n, err := v.flush()
		if err != nil {
			return nil, err
		}
		if n == 0 || round >= 8 {
			if comp != nil && n > 0 {
				comp = sc.Fill(r, p, f, v)
			}
			return comp, nil
		}
	}
}
