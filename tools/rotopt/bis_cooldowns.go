package main

import (
	"flag"
	"fmt"
	"slices"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// On cooldown ranks (docs/bis_lists.md). The user's addon swaps trinkets between fights, so it
// wants to know what a trinket is worth while its use is on cooldown: Chained Essence of Eranikus
// may be the best trinket for the pull, and only filler for the next fight.
//
// With -use-dur a set's score is its DPS with every use off, plus each use's value on a shorter
// fight (use_values.go). An item's score with its use on cooldown is the same without its own
// use value, while the other items' uses still count. So its cooldown loss is its loss plus its
// use value, and that needs no runs of its own. We write where that puts it on the same list in
// its cooldown field, with the rank rule of the conditional ranks (bisRank). The picks get it too,
// since the addon takes a pick off as well when its use is on cooldown.
//
// A trinket can add nothing at all while its use is on cooldown (Barov Peasant Caller), or on the
// plain boss the list is ranked on (Rune of the Dawn, until the enemy is undead). A rank there only
// says where an empty slot would go, so we don't write one (user, 2026-11-24), see zeroTrinkets.

var gearRankCooldowns = flag.Bool("gear-rank-cooldowns", false, "gear search: with -gear-rank-out and -use-dur, also rank each listed item that has a use without its use value, as it is while its use is on cooldown (docs/bis_lists.md)")

// Where an item would go on its list with its use on cooldown, as the list file has it. A trinket
// that adds nothing then has no rank.
type bisCooldown struct {
	Rank int     `json:"rank,omitempty"`
	Loss float64 `json:"loss"`
	// BiS says the item would still be the pick with its use on cooldown (bisWins).
	BiS bool `json:"bis,omitempty"`
}

// setItemSwitches sets the item switches of the runs we prepare next, see Player.item_uses_off
// and Player.item_effects_off. nil clears them.
func (s *searcher) setItemSwitches(usesOff, effectsOff []int32) {
	player := s.setup.player()
	player.ItemUsesOff, player.ItemEffectsOff = usesOff, effectsOff
}

// rankCooldowns adds the cooldown ranks to the lists' items that have a use. Every entry was
// scored in this run, so core.ItemHasUse knows it. zero is zeroTrinkets' answer.
func rankCooldowns(slots []bisSlot, zero map[int32]bool) []bisSlot {
	for i := range slots {
		slot := &slots[i]
		for j := range slot.entries {
			e := &slot.entries[j]
			if !core.ItemHasUse(e.ID) {
				continue
			}
			loss := round2(e.Loss + e.use)
			if zero[e.ID] {
				e.Cooldown = &bisCooldown{Loss: loss}
				fmt.Printf("  %s: %s on cooldown: adds nothing (%+.1f, its use is worth %+.1f)\n", slot.name, e.Name, 0-loss, e.use)
				continue
			}
			e.Cooldown = &bisCooldown{Rank: bisRank(slot.entries, j, loss), Loss: loss, BiS: bisWins(*e, loss)}
			fmt.Printf("  %s: %s on cooldown: rank %d (%+.1f, its use is worth %+.1f)\n", slot.name, e.Name, e.Cooldown.Rank, 0-loss, e.use)
		}
	}
	return slots
}

// zeroTrinkets says which of the listed trinkets add nothing with their use off: the set scores
// the same with the trinket as with its slot empty, on the list's own fight.
//
// So it counts the trinkets without stats (Barov Peasant Caller) and the ones whose stats do
// nothing for the spec: Flamebreak Talisman's resistance, or Blackhand's Breadth's melee crit on
// a mage. The runs share their seed and give each kind of roll its own random source, so a
// trinket that does nothing scores the same as the empty slot. It's the same 0.1 DPS rounding as
// the effects file's cooldown value (bis_effects.go).
func (s *searcher) zeroTrinkets(slots []bisSlot, best []*proto.ItemSpec, knobs Knobs) map[int32]bool {
	at := slices.IndexFunc(slots, func(slot bisSlot) bool { return slot.name == "trinket" })
	if at < 0 {
		return nil
	}
	full := s.iterations
	s.iterations, s.labeledRands = int32(*gearRankConfirmIters), true
	if s.iterations <= 0 {
		s.iterations = 10 * full
	}
	defer func() {
		s.iterations, s.labeledRands = full, false
		s.setGear(best)
	}()
	var pending []*pendingScore
	prepare := func(slot proto.ItemSlot, p *proto.ItemSpec) int {
		gear := slices.Clone(best)
		gear[slot] = p
		s.setGear(gear)
		pending = append(pending, s.prepare(knobs))
		return len(pending) - 1
	}
	// Each trinket's run and its slot's empty one. A pick stays in its own slot.
	runs := map[int32][2]int{}
	empty := map[proto.ItemSlot]int{}
	for _, e := range slots[at].entries {
		slot := proto.ItemSlot_ItemSlotTrinket2
		if p := best[proto.ItemSlot_ItemSlotTrinket1]; p != nil && p.Id == e.ID {
			slot = proto.ItemSlot_ItemSlotTrinket1
		}
		if _, ok := empty[slot]; !ok {
			empty[slot] = prepare(slot, nil)
		}
		runs[e.ID] = [2]int{prepare(slot, &proto.ItemSpec{Id: e.ID, RandomSuffix: e.Suffix}), empty[slot]}
	}
	scores := s.scoreAll(pending)
	zero := map[int32]bool{}
	for id, r := range runs {
		with, without := scores[r[0]], scores[r[1]]
		zero[id] = round1((with.dps-with.useSum)-(without.dps-without.useSum)) == 0
	}
	return zero
}

// markZero marks the listed trinkets that add nothing on the list's own fight, and takes out the
// ones that add nothing under any condition either. A pick stays, whatever it adds.
func markZero(slots []bisSlot, zero map[int32]bool) []bisSlot {
	for i := range slots {
		if slots[i].name != "trinket" {
			continue
		}
		slots[i].entries = slices.DeleteFunc(slots[i].entries, func(e bisEntry) bool {
			return zero[e.ID] && e.use == 0 && !e.BiS && len(e.Conditional) == 0
		})
		for j := range slots[i].entries {
			e := &slots[i].entries[j]
			e.Zero = zero[e.ID] && e.use == 0 && !e.BiS
		}
	}
	return slots
}
