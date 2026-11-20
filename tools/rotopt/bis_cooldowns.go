package main

import (
	"flag"
	"fmt"

	"github.com/wowsims/classic/sim/core"
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

var gearRankCooldowns = flag.Bool("gear-rank-cooldowns", false, "gear search: with -gear-rank-out and -use-dur, also rank each listed item that has a use without its use value, as it is while its use is on cooldown (docs/bis_lists.md)")

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

// rankCooldowns adds the cooldown ranks to the lists' items that have a use. Every entry was
// scored in this run, so core.ItemHasUse knows it.
func rankCooldowns(slots []bisSlot) []bisSlot {
	for i := range slots {
		slot := &slots[i]
		for j := range slot.entries {
			e := &slot.entries[j]
			if !core.ItemHasUse(e.ID) {
				continue
			}
			loss := round2(e.Loss + e.use)
			e.Cooldown = &bisCooldown{Rank: bisRank(slot.entries, j, loss), Loss: loss}
			fmt.Printf("  %s: %s on cooldown: rank %d (%+.1f, its use is worth %+.1f)\n", slot.name, e.Name, e.Cooldown.Rank, 0-loss, e.use)
		}
	}
	return slots
}
