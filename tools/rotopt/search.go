package main

import (
	"fmt"
	"math"
	"os"
	"sync"

	"github.com/wowsims/classic/sim/core/proto"
)

// Coordinate descent over the knobs. Every knob in turn tries each of its values with the
// others held fixed, and the best value is kept when it beats the current one by more than
// the noise. That repeats until a full round changes nothing. It is not global, but the
// knobs here are thresholds with one clear optimum each, and the sim is too slow for
// anything that scales worse.
type searcher struct {
	setup      *setup
	template   Template
	durations  []float64
	iterations int32

	// Standard errors of improvement a change has to show before it is kept.
	confidence float64

	// The gear search first scores every item of a slot at screenIterations and only
	// runs the best screenKeep of them again at full iterations. 0 scores all of them
	// at full iterations.
	screenIterations int32
	screenKeep       int

	// The gear search also searches enchants, from the level 20 list in enchants.go,
	// leaving out enchantExclude.
	enchantSearch  bool
	enchantExclude []int32
	// The character level, for the kits' level requirement.
	level int32

	// The talent search scores a round again at refineIterations when nothing in it
	// stands out at iterations. 0 turns that off.
	refineIterations int32

	// The gear search can ask for a minimum of unbuffed health and armor, for a PvP set.
	// A set below the minimum loses survivalPenalty DPS for each point of health it is
	// short, and a fifth of that for each point of armor. So the search first climbs to
	// the minimum, and then looks for the most damage above it. survival also reports
	// the health and armor without asking for a minimum.
	minHealth, minArmor float64
	survivalPenalty     float64
	survival            bool

	// The gear search can also put a price on unbuffed health and armor, in DPS for each
	// 100 points. A PvP set then trades damage for health and armor wherever the trade
	// is cheaper than the price, the same way a stat weight picks gear. This takes one
	// search per set, where a minimum needs a search for every floor we try.
	healthWeight, armorWeight float64

	cache map[string]result
	evals int
}

// Mean DPS and end mana over the fight lengths, with the standard errors of those means.
func (s *searcher) score(knobs Knobs) result {
	return s.scoreAll([]*pendingScore{s.prepare(knobs)})[0]
}

// pendingScore is a score still to work out: a sim request for each fight length, made from
// the setup as it was when we asked. So the gear and talent searches can change the setup for
// the next candidate before the runs happen, see scoreAll.
type pendingScore struct {
	knobs Knobs
	// The cache it was made against. The gear search starts a new one for each gear set, so
	// the results must go into this one.
	cache    map[string]result
	keys     []string
	requests []*proto.RaidSimRequest // nil where the cache had the result
	results  []result

	health, armor float64
}

func (s *searcher) prepare(knobs Knobs) *pendingScore {
	rot := s.template.Build(knobs)
	p := &pendingScore{knobs: knobs, cache: s.cache}
	for _, d := range s.durations {
		key := fmt.Sprintf("%s|%g|%s|%d", s.setup.talents, d, knobs, s.iterations)
		r, ok := s.cache[key]
		var request *proto.RaidSimRequest
		if !ok {
			request = s.setup.runRequest(rot, d, s.iterations)
		}
		p.keys = append(p.keys, key)
		p.requests = append(p.requests, request)
		p.results = append(p.results, r)
	}
	if s.survival || s.minHealth > 0 || s.minArmor > 0 || s.healthWeight != 0 || s.armorWeight != 0 {
		p.health, p.armor = s.setup.unbuffedStats()
	}
	return p
}

// How many sim runs scoreAll runs side by side. Each run already splits its iterations over
// all the CPUs, but at the iterations a search uses, the splits are short, and the CPUs wait
// for the slowest one before the next run starts. A few runs side by side keep them busy.
const scoreParallel = 4

// scoreAll works out the scores, running the sims side by side. The results are the same as
// scoring them one by one, because each run has its own fixed random seed.
func (s *searcher) scoreAll(pending []*pendingScore) []result {
	type run struct {
		p   *pendingScore
		i   int
		err error
	}
	var runs []*run
	for _, p := range pending {
		for i, request := range p.requests {
			if request != nil {
				runs = append(runs, &run{p: p, i: i})
			}
		}
	}
	var wg sync.WaitGroup
	slots := make(chan struct{}, scoreParallel)
	for _, r := range runs {
		wg.Add(1)
		slots <- struct{}{}
		go func() {
			defer wg.Done()
			r.p.results[r.i], r.err = simulate(r.p.requests[r.i])
			<-slots
		}()
	}
	wg.Wait()
	for _, r := range runs {
		if r.err != nil {
			fmt.Fprintf(os.Stderr, "sim failed for %s at %gs: %v\n", r.p.knobs, s.durations[r.i], r.err)
			os.Exit(1)
		}
		r.p.cache[r.p.keys[r.i]] = r.p.results[r.i]
		s.evals++
	}

	results := make([]result, len(pending))
	for i, p := range pending {
		results[i] = s.combine(p)
	}
	return results
}

// combine is the mean over the fight lengths of a score whose runs are done.
func (s *searcher) combine(p *pendingScore) result {
	var sum, variance, mana, manaVariance, tto float64
	for _, r := range p.results {
		sum += r.dps
		variance += r.stderr * r.stderr
		mana += r.mana
		manaVariance += r.manaStderr * r.manaStderr
		tto += r.tto
	}
	n := float64(len(s.durations))
	r := result{
		dps: sum / n, stderr: math.Sqrt(variance) / n,
		mana: mana / n, manaStderr: math.Sqrt(manaVariance) / n,
		tto: tto / n,
		health: p.health, armor: p.armor,
	}
	if len(s.durations) == 1 {
		r.leveling = p.results[0].leveling
	}
	return r
}

// What the gear search maximizes: DPS and the price of the set's health and armor, less
// the penalty for a set below the minimum health and armor.
func (s *searcher) objective(r result) float64 {
	return r.dps + (s.healthWeight*r.health+s.armorWeight*r.armor)/100 -
		s.survivalPenalty*(max(s.minHealth-r.health, 0)+0.2*max(s.minArmor-r.armor, 0))
}

// frozenKnobs are the knobs the search leaves at their starting values (-freeze).
var frozenKnobs = map[string]bool{}

func (s *searcher) search(start Knobs) (Knobs, result) {
	best := Knobs{}
	for name, v := range start {
		best[name] = v
	}
	bestScore := s.score(best)
	fmt.Printf("start %s: %s\n", best, bestScore)

	for round := 1; ; round++ {
		improved := false
		for _, knob := range s.template.Knobs() {
			if frozenKnobs[knob.Name] {
				continue
			}
			current := best[knob.Name]
			var values []float64
			var pending []*pendingScore
			for _, v := range knob.Values() {
				if v == current {
					continue
				}
				candidate := Knobs{}
				for name, val := range best {
					candidate[name] = val
				}
				candidate[knob.Name] = v
				values = append(values, v)
				pending = append(pending, s.prepare(candidate))
			}
			var bestValue float64
			bestGain := 0.0
			for i, score := range s.scoreAll(pending) {
				v := values[i]
				gain := score.dps - bestScore.dps
				noise := math.Sqrt(score.stderr*score.stderr + bestScore.stderr*bestScore.stderr)
				fmt.Printf("  %s=%g: %s (%+.1f, noise %.1f)\n", knob.Name, v, score, gain, noise)
				if gain > s.confidence*noise && gain > bestGain {
					bestGain, bestValue = gain, v
				}
			}
			if bestGain > 0 {
				best[knob.Name] = bestValue
				bestScore = s.score(best)
				improved = true
				fmt.Printf("round %d: %s=%g, now %s\n", round, knob.Name, bestValue, bestScore)
			}
		}
		if !improved {
			return best, bestScore
		}
	}
}
