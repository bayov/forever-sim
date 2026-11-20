package main

import (
	"flag"
	"fmt"
	"slices"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// On cooldown ranks (docs/bis_lists.md). The list is ranked with every use pressed at the pull.
// The user's addon swaps trinkets between fights, so it wants to know what a trinket is worth
// while its use is on cooldown: Adaptive Combat Assistant may be the best trinket for the pull,
// and only filler for the next fight two minutes later.
//
// So with -gear-rank-cooldowns we score each listed item that has a use a second time with the
// use never pressed (Player.item_uses_off), so only its stats, equip effects and procs count.
// We write where that puts it on the same list in its cooldown field, with the rank rule of the
// conditional ranks (bisRank). The picks get it too, since the addon takes a pick off as well
// when its use is on cooldown.
//
// The runs are at -gear-rank-confirm-iters and share their seed, as for the conditional ranks.
// They also give each kind of roll its own random source (SimOptions.use_labeled_rands), which
// pairs them about four times better (see writeEffects). The base is the set itself (its uses
// pressed), scored once the same way. An item that isn't a pick goes in
// the places its list scored it in (a ring or trinket in place of either pick), and the best of
// them counts, which is in place of the weaker pick.

var gearRankCooldowns = flag.Bool("gear-rank-cooldowns", false, "gear search: with -gear-rank-out, also rank each listed item that has a use with the use never pressed, as it is while its use is on cooldown (docs/bis_lists.md)")

// Where an item would go on its list with its use on cooldown, as the list file has it.
type bisCooldown struct {
	Rank int     `json:"rank"`
	Loss float64 `json:"loss"`
}

// setItemSwitches sets the item switches of the runs we prepare next, see Player.item_uses_off
// and Player.item_effects_off. nil clears them.
func (s *searcher) setItemSwitches(usesOff, effectsOff []int32) {
	player := s.setup.player()
	player.ItemUsesOff, player.ItemEffectsOff = usesOff, effectsOff
}

// rankCooldowns adds the cooldown ranks to the lists' items that have a use. sources are the
// lists' options (rankBisPass). Every entry was scored in this run, so core.ItemHasUse knows
// it.
func (s *searcher) rankCooldowns(slots []bisSlot, sources map[string][]rankSource, best []*proto.ItemSpec, knobs Knobs) []bisSlot {
	iterations := int32(*gearRankConfirmIters)
	if iterations <= 0 {
		iterations = 10 * s.iterations
	}
	full := s.iterations
	s.iterations, s.saveValues, s.labeledRands = iterations, true, true
	defer func() {
		s.iterations, s.saveValues, s.labeledRands = full, false, false
		s.setItemSwitches(nil, nil)
	}()

	var pending []*pendingScore
	var done []func(result)
	run := func(gear []*proto.ItemSpec, usesOff []int32, then func(result)) {
		s.setItemSwitches(usesOff, nil)
		s.setGear(gear)
		pending = append(pending, s.prepare(knobs))
		done = append(done, then)
	}

	// An entry with a use, and its best score with the use off.
	type candidate struct {
		list, entry int
		off         *result
	}
	var base result
	run(best, nil, func(r result) { base = r })
	var candidates []*candidate
	for i, slot := range slots {
		for j, e := range slot.entries {
			if !core.ItemHasUse(e.ID) {
				continue
			}
			c := &candidate{list: i, entry: j}
			candidates = append(candidates, c)
			keepBest := func(r result) {
				if c.off == nil || r.dps > c.off.dps {
					c.off = &r
				}
			}
			usesOff := []int32{e.ID}
			if e.BiS {
				run(best, usesOff, keepBest)
				continue
			}
			for _, src := range sources[slot.name] {
				i := slices.IndexFunc(src.options, func(o gearOption) bool {
					return o.spec.Id == e.ID && o.spec.RandomSuffix == e.Suffix && o.name == e.Name
				})
				if i < 0 {
					continue
				}
				for _, gear := range src.placements(src.options[i]) {
					if gear != nil {
						run(gear, usesOff, keepBest)
					}
				}
				break
			}
		}
	}
	if len(candidates) == 0 {
		return slots
	}
	fmt.Printf("ranking %d items with a use, with the use on cooldown, at %d iterations:\n", len(candidates), iterations)
	for i, r := range s.scoreAll(pending) {
		done[i](r)
	}
	for _, c := range candidates {
		slot := &slots[c.list]
		e := &slot.entries[c.entry]
		if c.off == nil {
			fmt.Printf("  %s: %s: no place to score it in\n", slot.name, e.Name)
			continue
		}
		loss := round2(base.dps - c.off.dps)
		e.Cooldown = &bisCooldown{Rank: bisRank(slot.entries, c.entry, loss), Loss: loss}
		fmt.Printf("  %s: %s on cooldown: rank %d (%+.1f ± %.1f)\n", slot.name, e.Name, e.Cooldown.Rank, 0-loss, gainNoise(base, *c.off))
	}
	return slots
}
