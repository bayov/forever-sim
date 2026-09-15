package main

import (
	"fmt"
	"math"
	"os"
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

	cache map[string]result
	evals int
}

// Mean DPS over the fight lengths, with the standard error of that mean.
func (s *searcher) score(knobs Knobs) result {
	rot := s.template.Build(knobs)
	var sum, variance float64
	for _, d := range s.durations {
		key := fmt.Sprintf("%g|%s", d, knobs)
		r, ok := s.cache[key]
		if !ok {
			var err error
			r, err = s.setup.run(rot, d, s.iterations)
			if err != nil {
				fmt.Fprintf(os.Stderr, "sim failed for %s at %gs: %v\n", knobs, d, err)
				os.Exit(1)
			}
			s.cache[key] = r
			s.evals++
		}
		sum += r.dps
		variance += r.stderr * r.stderr
	}
	n := float64(len(s.durations))
	return result{dps: sum / n, stderr: math.Sqrt(variance) / n}
}

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
			current := best[knob.Name]
			var bestValue float64
			bestGain := 0.0
			for _, v := range knob.Values() {
				if v == current {
					continue
				}
				candidate := Knobs{}
				for name, val := range best {
					candidate[name] = val
				}
				candidate[knob.Name] = v
				score := s.score(candidate)
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
