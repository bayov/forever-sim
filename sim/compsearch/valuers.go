package compsearch

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
	k := askKey(b, s)
	if j, ok := v.vals[k]; ok {
		return j.Value.DPS, j.Value.SE, j.key
	}
	if _, ok := v.missing[k]; !ok {
		v.missing[k] = v.sc.job(b, s, v.iters)
	}
	return v.model.V(b, s), 0, ""
}

// flush sims everything recorded as missing.
func (v *exactValuer) flush() (int, error) {
	if len(v.missing) == 0 {
		return 0, nil
	}
	jobs := make([]*Job, 0, len(v.missing))
	for _, j := range v.missing {
		jobs = append(jobs, j)
	}
	if err := v.sc.Sim.Sim(jobs); err != nil {
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
