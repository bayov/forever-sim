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
	rot := s.template.Build(knobs)
	var sum, variance, mana, manaVariance, tto float64
	for _, d := range s.durations {
		key := fmt.Sprintf("%s|%g|%s|%d", s.setup.talents, d, knobs, s.iterations)
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
		mana += r.mana
		manaVariance += r.manaStderr * r.manaStderr
		tto += r.tto
	}
	n := float64(len(s.durations))
	r := result{
		dps: sum / n, stderr: math.Sqrt(variance) / n,
		mana: mana / n, manaStderr: math.Sqrt(manaVariance) / n,
		tto: tto / n,
	}
	if s.survival || s.minHealth > 0 || s.minArmor > 0 || s.healthWeight != 0 || s.armorWeight != 0 {
		r.health, r.armor = s.setup.unbuffedStats()
	}
	return r
}

// What the gear search maximizes: DPS and the price of the set's health and armor, less
// the penalty for a set below the minimum health and armor.
func (s *searcher) objective(r result) float64 {
	return r.dps + (s.healthWeight*r.health+s.armorWeight*r.armor)/100 -
		s.survivalPenalty*(max(s.minHealth-r.health, 0)+0.2*max(s.minArmor-r.armor, 0))
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
