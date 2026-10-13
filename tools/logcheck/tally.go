package main

import (
	"math"
	"strings"
)

// abilityKey is one row of the report. A white swing has spell 0. A DoT's ticks get their own
// row, apart from the spell's direct hit.
type abilityKey struct {
	spell    int32
	periodic bool
}

// tally counts how one ability's attempts on one mob level came out.
type tally struct {
	attempts int
	outcomes map[string]int
	// raw is each landed hit's damage before the target's armor, resists, crit and damage
	// taken modifiers. That's the log's raw field and the sim's RawDamage.
	raw []float64
	// spellLanded counts the landed hits of a school other than Physical, which a resist can
	// cut, and partial counts them by how much was resisted, in steps of 10%.
	spellLanded int
	partial     map[int]int
}

func newTally() *tally {
	return &tally{outcomes: map[string]int{}, partial: map[int]int{}}
}

// tallies holds every ability by mob level, for the log or for all sim iterations together.
type tallies struct {
	byLevel map[int]map[abilityKey]*tally
	// whiteLanded counts the landed white swings by mob level, which proc rates are per.
	whiteLanded map[int]int
	iterations  int
}

func newTallies() *tallies {
	return &tallies{byLevel: map[int]map[abilityKey]*tally{}, whiteLanded: map[int]int{}}
}

func (ts *tallies) get(level int, key abilityKey) *tally {
	byKey := ts.byLevel[level]
	if byKey == nil {
		byKey = map[abilityKey]*tally{}
		ts.byLevel[level] = byKey
	}
	t := byKey[key]
	if t == nil {
		t = newTally()
		byKey[key] = t
	}
	return t
}

// add records one attempt. outcome is one of the names below, and the rest only matter when
// it landed.
func (ts *tallies) add(level int, key abilityKey, outcome string, raw float64, physical bool, resistedShare float64) {
	t := ts.get(level, key)
	t.attempts++
	t.outcomes[outcome]++
	if !landed(outcome) {
		return
	}
	if key.spell == 0 && !key.periodic {
		ts.whiteLanded[level]++
	}
	t.raw = append(t.raw, raw)
	if !physical {
		t.spellLanded++
		if resistedShare > 0.001 {
			t.partial[int(math.Round(resistedShare*10))*10]++
		}
	}
}

func (ts *tallies) merge(other *tallies) {
	for level, byKey := range other.byLevel {
		for key, o := range byKey {
			t := ts.get(level, key)
			t.attempts += o.attempts
			for name, n := range o.outcomes {
				t.outcomes[name] += n
			}
			t.raw = append(t.raw, o.raw...)
			t.spellLanded += o.spellLanded
			for share, n := range o.partial {
				t.partial[share] += n
			}
		}
	}
	for level, n := range other.whiteLanded {
		ts.whiteLanded[level] += n
	}
	ts.iterations += other.iterations
}

// The outcomes, in the order the report lists them. A crit, glance, crush or block landed, and
// a hit is a landed attempt that was none of those. A full resist counts as a miss, because the
// client logs a spell's hit roll miss as RESIST and the sim calls it a miss.
var outcomeOrder = []string{"hit", "crit", "glance", "crush", "block", "miss", "dodge", "parry", "immune", "absorb", "evade", "other"}

func landed(outcome string) bool {
	switch outcome {
	case "hit", "crit", "glance", "crush", "block":
		return true
	}
	return false
}

// logOutcome names a log line's outcome.
func logOutcome(e *event, d damage, isDamage bool) string {
	if isDamage {
		switch {
		case d.crit:
			return "crit"
		case d.glance:
			return "glance"
		case d.crush:
			return "crush"
		case d.blocked > 0:
			return "block"
		}
		return "hit"
	}
	switch miss := strings.ToLower(e.missType()); miss {
	case "miss", "resist":
		return "miss"
	case "dodge", "parry", "block", "immune", "absorb", "evade":
		return miss
	}
	return "other"
}

// wilson is the 95% confidence interval of k successes in n tries.
func wilson(k, n int) (float64, float64) {
	if n == 0 {
		return 0, 1
	}
	const z = 1.96
	p := float64(k) / float64(n)
	nf := float64(n)
	center := (p + z*z/(2*nf)) / (1 + z*z/nf)
	half := z * math.Sqrt(p*(1-p)/nf+z*z/(4*nf*nf)) / (1 + z*z/nf)
	return math.Max(0, center-half), math.Min(1, center+half)
}

// binomialP is the two-sided chance of a count at least as far from n * p as k, when each of
// n tries succeeds with chance p. We use the exact sum rather than a z-score, because most
// rows have few tries and the normal approximation overstates them.
func binomialP(k, n int, p float64) float64 {
	if p <= 0 || p >= 1 {
		if (p <= 0 && k == 0) || (p >= 1 && k == n) {
			return 1
		}
		return 0
	}
	logPMF := func(i int) float64 {
		a, _ := math.Lgamma(float64(n + 1))
		b, _ := math.Lgamma(float64(i + 1))
		c, _ := math.Lgamma(float64(n - i + 1))
		return a - b - c + float64(i)*math.Log(p) + float64(n-i)*math.Log(1-p)
	}
	low, high := 0.0, 0.0
	for i := 0; i <= n; i++ {
		pmf := math.Exp(logPMF(i))
		if i <= k {
			low += pmf
		}
		if i >= k {
			high += pmf
		}
	}
	return math.Min(1, 2*math.Min(low, high))
}

// poissonP is the two-sided chance of a count at least as far from the mean as k, for counts
// that can pass one per try, like procs per landed swing.
func poissonP(k int, mean float64) float64 {
	if mean <= 0 {
		if k == 0 {
			return 1
		}
		return 0
	}
	low, high, pmf := 0.0, 0.0, math.Exp(-mean)
	limit := max(k, int(mean*3)+20)
	for i := 0; i <= limit; i++ {
		if i > 0 {
			pmf *= mean / float64(i)
		}
		if i <= k {
			low += pmf
		}
		if i >= k {
			high += pmf
		}
	}
	if k > limit {
		high = 0
	}
	return math.Min(1, 2*math.Min(low, high))
}

type spread struct {
	n                    int
	min, max, sum, sumSq float64
}

func spreadOf(values []float64) spread {
	s := spread{n: len(values), min: math.Inf(1), max: math.Inf(-1)}
	for _, v := range values {
		s.min = math.Min(s.min, v)
		s.max = math.Max(s.max, v)
		s.sum += v
		s.sumSq += v * v
	}
	return s
}

func (s spread) mean() float64 {
	if s.n == 0 {
		return 0
	}
	return s.sum / float64(s.n)
}

func (s spread) variance() float64 {
	if s.n < 2 {
		return 0
	}
	mean := s.mean()
	return math.Max(0, s.sumSq/float64(s.n)-mean*mean)
}

func normalCDF(z float64) float64 {
	return 0.5 * math.Erfc(-z/math.Sqrt2)
}
