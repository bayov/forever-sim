package main

import (
	"fmt"
	"math"
	"strings"
)

// The amount check.
//
// The log's raw field is truncated, so a hit of 8.8 always logs raw 8, and the raw row can't
// tell 8.8 from 8.1. The amount the target took is rounded up with a chance equal to the
// fraction (9 on 80% of 8.8 hits), so how often it rounds up pins the fraction. A hit of v
// shows amount a with a chance of max(0, 1 - |v - a|).
//
// It only works for an ability whose every sim hit has the same value, like Flametongue's
// attack, Flame Shock's ticks and Lightning Shield's orbs. A spell with a damage range lands
// anywhere between two whole points, so its round ups say nothing about the formula.

// steadyHit is an ability's one hit value in the sim: raw is the hit before crit and damage
// taken modifiers, and hit and crit are the amounts. hasCrit is false when the sim never crit.
type steadyHit struct {
	raw, hit, crit float64
	hasCrit        bool
}

// steadyHitOf returns the sim's one hit value, or false when its hits vary.
func steadyHitOf(hits []landedHit) (steadyHit, bool) {
	var raws, normal, crits []float64
	for _, h := range hits {
		raws = append(raws, h.raw)
		if h.crit {
			crits = append(crits, h.amount)
		} else {
			normal = append(normal, h.amount)
		}
	}
	steady := func(values []float64) bool {
		s := spreadOf(values)
		return s.max-s.min < 0.005
	}
	if len(normal) == 0 || !steady(raws) || !steady(normal) || !steady(crits) {
		return steadyHit{}, false
	}
	s := steadyHit{raw: spreadOf(raws).mean(), hit: spreadOf(normal).mean()}
	if len(crits) > 0 {
		s.crit, s.hasCrit = spreadOf(crits).mean(), true
	}
	return s, s.raw > 0
}

// amountHit is a log hit the check uses: the amount it showed, and the multiplier from the
// raw hit to the amount (the sim's damage taken modifiers, times the crit for a crit).
type amountHit struct {
	amount, multiplier float64
}

// roundUpChance is how likely a hit of v shows amount a.
func roundUpChance(v, a float64) float64 {
	return math.Max(0, 1-math.Abs(v-a))
}

// amount prints the amount row for one ability, on one mob level or on all of them together
// (where says which, for the flagged list).
func (r *report) amount(where, name string, lt, st *tally) {
	sim, ok := steadyHitOf(st.hits)
	if !ok {
		return
	}
	base := math.Floor(sim.raw + 1e-6)
	var hits []amountHit
	otherRaw, modified := 0, 0
	for _, h := range lt.hits {
		multiplier := sim.hit / sim.raw
		if h.crit {
			if !sim.hasCrit {
				continue
			}
			multiplier = sim.crit / sim.raw
		}
		// A hit with another raw had other stats (a buff, a gear swap), and the raw row
		// already shows it.
		if h.raw != base {
			otherRaw++
			continue
		}
		// An amount no hit from base to base + 1 can give comes from a modifier the sim
		// doesn't have on the mob, like the Stormstrike mark.
		if h.amount <= multiplier*base-1 || h.amount >= multiplier*(base+1)+1 {
			modified++
			continue
		}
		hits = append(hits, amountHit{amount: h.amount, multiplier: multiplier})
	}
	if len(hits) == 0 {
		if otherRaw+modified > 0 {
			fmt.Printf("  amount   no hit to check: the sim's raw is %.3f, and the log has %d hits with another raw and %d with a modifier the sim doesn't have\n", sim.raw, otherRaw, modified)
		}
		return
	}

	// Under the sim's hit, each log hit either rounded up or down, with its own chance. The
	// count of round ups then has a Poisson binomial distribution.
	up, impossible := 0, 0
	chances := make([]float64, len(hits))
	expected := 0.0
	for i, h := range hits {
		v := h.multiplier * sim.raw
		chances[i] = v - math.Floor(v)
		expected += chances[i]
		switch h.amount {
		case math.Floor(v) + 1:
			up++
		case math.Floor(v):
		default:
			impossible++
		}
	}
	pValue := 0.0
	if impossible == 0 {
		pValue = poissonBinomialP(up, chances)
	}
	lo, hi, fits := hitInterval(hits, base)

	fit := "no single hit value fits these amounts"
	if fits {
		fit = fmt.Sprintf("the log puts the hit at %.2f-%.2f", lo, hi)
	}
	var left []string
	if otherRaw > 0 {
		left = append(left, fmt.Sprintf("%d with another raw", otherRaw))
	}
	if modified > 0 {
		left = append(left, fmt.Sprintf("%d with a modifier the sim doesn't have", modified))
	}
	leftOut := ""
	if len(left) > 0 {
		leftOut = fmt.Sprintf(" (left out %s)", strings.Join(left, ", "))
	}
	mark := pMark(pValue)
	fmt.Printf("  amount   rounded up on %d of %d, the sim's %.3f expects %.1f   p %s%s   %s%s\n",
		up, len(hits), sim.raw, expected, pString(pValue), mark, fit, leftOut)
	if mark == " !!" {
		r.flagged = append(r.flagged, fmt.Sprintf("%s, %s: amount rounded up on %d of %d, the sim's hit of %.3f expects %.1f (%s)", where, name, up, len(hits), sim.raw, expected, fit))
	}
}

// pooledHits holds an ability's hits from every mob level, for the amount check on all of
// them together. The hit's value doesn't depend on the mob's level (resisted hits are left
// out), and one level alone often has only a handful of hits.
type pooledHits struct {
	log, sim *tally
	levels   int
}

func (r *report) pool(key abilityKey, lt, st *tally) {
	if len(lt.hits) == 0 || len(st.hits) == 0 {
		return
	}
	p := r.pooled[key]
	if p == nil {
		p = &pooledHits{log: newTally(), sim: newTally()}
		r.pooled[key] = p
		r.pooledOrder = append(r.pooledOrder, key)
	}
	p.log.hits = append(p.log.hits, lt.hits...)
	p.sim.hits = append(p.sim.hits, st.hits...)
	p.levels++
}

// amountAllLevels prints the amount row on all mob levels together, for the abilities we hit
// more than one level with.
func (r *report) amountAllLevels() {
	printed := false
	for _, key := range r.pooledOrder {
		p := r.pooled[key]
		if p.levels < 2 {
			continue
		}
		if _, ok := steadyHitOf(p.sim.hits); !ok {
			continue
		}
		if !printed {
			fmt.Println("\n=== Amounts on all mob levels together")
			printed = true
		}
		name := r.name(key)
		fmt.Printf("\n%s: %d levels\n", name, p.levels)
		r.amount("all levels", name, p.log, p.sim)
	}
}

// hitInterval is the range of raw hits from base to base + 1 that fit the log's amounts, as
// a 95% likelihood interval. fits is false when no hit gives every amount.
func hitInterval(hits []amountHit, base float64) (lo, hi float64, fits bool) {
	const steps = 1000
	logLikelihood := make([]float64, steps)
	best := math.Inf(-1)
	for i := range steps {
		x := base + (float64(i)+0.5)/steps
		sum := 0.0
		for _, h := range hits {
			sum += math.Log(roundUpChance(h.multiplier*x, h.amount))
		}
		logLikelihood[i] = sum
		best = math.Max(best, sum)
	}
	if math.IsInf(best, -1) {
		return 0, 0, false
	}
	lo, hi = math.Inf(1), math.Inf(-1)
	for i, ll := range logLikelihood {
		// 1.92 is half of 3.84, the 95% point of a chi-squared with one degree of freedom.
		if ll >= best-1.92 {
			x := base + (float64(i)+0.5)/steps
			lo, hi = math.Min(lo, x), math.Max(hi, x)
		}
	}
	return lo, hi, true
}

// poissonBinomialP is the two-sided chance of a count as unlikely as k, when each of the
// trials succeeds with its own chance.
func poissonBinomialP(k int, chances []float64) float64 {
	dist := []float64{1}
	for _, p := range chances {
		next := make([]float64, len(dist)+1)
		for j, q := range dist {
			next[j] += q * (1 - p)
			next[j+1] += q * p
		}
		dist = next
	}
	observed := dist[k]
	pValue := 0.0
	for _, q := range dist {
		if q <= observed*(1+1e-7) {
			pValue += q
		}
	}
	return math.Min(1, pValue)
}
