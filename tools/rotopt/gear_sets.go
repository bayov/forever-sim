package main

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// The gear search swaps one slot at a time, so a set piece only gets in when it beats the item in
// its slot on its own. Its set bonus counts only once enough pieces are worn, and the first
// pieces never get there. Forever's Five Thunders is like that: its 4 piece proc and 6 piece
// spell power are much of what the set is worth, but each piece alone is a middling item.
//
// So after each round we also try whole sets. For every set with bonuses that the pool has pieces
// of, we first score each slot's pieces alone (at the screen's iterations) and take the best piece
// of each slot. Then for each bonus's piece count we put on the pieces that cost the least,
// keeping the pieces already worn, and score that at full iterations. The best of those goes in
// when it beats the current gear the same way a single swap must. The next round then tries every
// slot again, so a set piece that isn't worth its share of the bonus can still come off.
//
// We leave out the weapon slots and the paired ring and trinket slots, because no set worth
// trying there would fit the swap rules of those slots in a simple way.
func (s *searcher) searchSets(pool [][]gearOption, best []*proto.ItemSpec, names []string, unsourced []bool, bestScore *result, knobs Knobs, round int) bool {
	type piece struct {
		slot   int
		option gearOption
	}
	bySet := map[*core.ItemSet][]piece{}
	var order []*core.ItemSet
	for slot, options := range pool {
		if weaponPair(slot) >= 0 || pairedSlot(slot) >= 0 || proto.ItemSlot(slot) == proto.ItemSlot_ItemSlotRanged {
			continue
		}
		for _, o := range options {
			set := core.ItemSetOf(core.ItemsByID[o.spec.Id])
			if set == nil || len(set.Bonuses) == 0 {
				continue
			}
			if _, ok := bySet[set]; !ok {
				order = append(order, set)
			}
			bySet[set] = append(bySet[set], piece{slot, o})
		}
	}

	improved := false
	for _, set := range order {
		pieces := bySet[set]
		inSet := func(spec *proto.ItemSpec) bool {
			return spec != nil && core.ItemSetOf(core.ItemsByID[spec.Id]) == set
		}
		worn := 0
		for slot, spec := range best {
			if inSet(spec) && weaponPair(slot) < 0 && pairedSlot(slot) < 0 {
				worn++
			}
		}

		// Each open slot's best piece, scored alone at the screen's iterations.
		full := s.iterations
		if s.screenIterations > 0 {
			s.iterations = s.screenIterations
		}
		var tried []piece
		var pending []*pendingScore
		for _, p := range pieces {
			if inSet(best[p.slot]) {
				continue
			}
			candidate := append([]*proto.ItemSpec(nil), best...)
			candidate[p.slot] = s.withEnchantOf(p.slot, p.option.spec, best[p.slot])
			s.setGear(candidate)
			tried = append(tried, p)
			pending = append(pending, s.prepare(knobs))
		}
		bestOfSlot := map[int]piece{}
		slotDPS := map[int]float64{}
		for i, score := range s.scoreAll(pending) {
			p := tried[i]
			if dps, ok := slotDPS[p.slot]; !ok || s.objective(score) > dps {
				bestOfSlot[p.slot], slotDPS[p.slot] = p, s.objective(score)
			}
		}
		s.iterations = full
		var open []int
		for slot := range bestOfSlot {
			open = append(open, slot)
		}
		sort.Slice(open, func(i, j int) bool { return slotDPS[open[i]] > slotDPS[open[j]] })

		// Every bonus's piece count the open slots can reach.
		var counts []int
		for n := range set.Bonuses {
			if int(n) > worn && int(n) <= worn+len(open) {
				counts = append(counts, int(n))
			}
		}
		slices.Sort(counts)
		if len(counts) == 0 {
			continue
		}
		var candidates [][]*proto.ItemSpec
		pending = nil
		for _, n := range counts {
			candidate := append([]*proto.ItemSpec(nil), best...)
			for _, slot := range open[:n-worn] {
				candidate[slot] = s.withEnchantOf(slot, bestOfSlot[slot].option.spec, best[slot])
			}
			s.setGear(candidate)
			candidates = append(candidates, candidate)
			pending = append(pending, s.prepare(knobs))
		}
		bestIndex, bestGain := -1, 0.0
		for i, score := range s.scoreAll(pending) {
			gain := s.objective(score) - s.objective(*bestScore)
			if gain > s.confidence*gainNoise(*bestScore, score) && gain > bestGain {
				bestIndex, bestGain = i, gain
			}
		}
		if bestIndex < 0 {
			s.setGear(best)
			continue
		}
		n := counts[bestIndex]
		var put []string
		for _, slot := range open[:n-worn] {
			best[slot] = candidates[bestIndex][slot]
			names[slot], unsourced[slot] = bestOfSlot[slot].option.name, bestOfSlot[slot].option.unsourced
			put = append(put, fmt.Sprintf("%s = %s", gearSlotNames[slot], names[slot]))
		}
		s.setGear(best)
		*bestScore = s.score(knobs)
		improved = true
		fmt.Printf("round %d: %s %d pieces: %s (%+.1f), now %s\n", round, set.Name, n, strings.Join(put, ", "), bestGain, *bestScore)
	}
	return improved
}
