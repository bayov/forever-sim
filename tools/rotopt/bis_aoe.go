package main

import (
	"fmt"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// Uses against several mobs (docs/bis_lists.md). A dungeon pull is often several mobs, and some
// uses hit them all: Smokey's Lighter breathes fire in a cone, and Chained Essence of Eranikus
// leaves a poison cloud. The list is ranked on one boss, where they hit only the boss.
//
// So with -use-aoe-targets 4 an item with a use gets one more conditional rank, "vs 4 mobs". Only
// its use counts against the 4 mobs: the use runs at -use-dur fight 4 copies of the boss, all
// alive for the whole fight, and the rest of the score stays on the one boss at -durs. That is
// what the use adds on a pull, and we don't switch the rotation or the gear to an AoE build for
// it. The rotation still hits the 4 mobs where it can (Magma Totem, Fire Nova), but a use value
// is the run with the use less the run without it, both against the 4 mobs, so that part
// cancels out. What stays is what the use adds to the AoE damage. A haste use like Manual Crowd
// Pummeler's gains a few DPS that way, since more swings give more Dragonbreath Chili procs, and
// the chili's fire hits all 4 mobs.
//
// Most uses gain little or nothing on 4 mobs. A line for them would only repeat the normal rank,
// so an item gets the line only when its use is worth more than aoeMinGain more against the 4
// mobs than against the boss.

// aoeMinGain is how much more a use must be worth against the mobs than against the boss before
// its item gets the "vs N mobs" line. Use values are about 0.1 to 0.3 DPS apart from one run to
// the next at -use-iters (2026-11-20), so a gain over 1 DPS is real.
const aoeMinGain = 1.0

// bisCondition is a condition the BiS list ranks an item under: one of the sim's item conditions
// (vs Undead, in snow), or its use against several mobs.
type bisCondition struct {
	item core.ItemCondition
	// useTargets is how many mobs the use runs fight, for the "vs N mobs" condition. 0 for an
	// item condition.
	useTargets int
}

func (c bisCondition) label() string {
	if c.useTargets > 0 {
		return fmt.Sprintf("vs %d mobs", c.useTargets)
	}
	return c.item.Label
}

// aoeCondition is the "vs N mobs" condition of -use-aoe-targets, and whether it is on.
func aoeCondition() (bisCondition, bool) {
	return bisCondition{useTargets: *useAoeTargets}, *useAoeTargets > 0
}

// itemBisConditions is the conditions the item is ranked under, but not the ones the fight
// already meets: its item conditions, and the "vs N mobs" one when it has a use. The item must
// have been in a sim run before, so core.ItemHasUse knows it.
func itemBisConditions(id int32, encounter *proto.Encounter) []bisCondition {
	var out []bisCondition
	for _, c := range core.ItemConditions(id) {
		if !c.MetBy(encounter) {
			out = append(out, bisCondition{item: c})
		}
	}
	if aoe, ok := aoeCondition(); ok && core.ItemHasUse(id) {
		out = append(out, aoe)
	}
	return out
}

// meet sets up the runs we prepare next with the condition met, on the fight's own encounter.
// The zero condition sets the fight back as it was.
func (s *searcher) meet(c bisCondition, encounter *proto.Encounter) {
	s.setup.request.Encounter = encounter
	if c.item != (core.ItemCondition{}) {
		s.setup.request.Encounter = c.item.Apply(encounter)
	}
	s.useTargets = c.useTargets
}

// aoeGains says whether a use worth single against the boss gains enough against the mobs,
// where it is worth aoe, to get its "vs N mobs" line.
func aoeGains(single, aoe float64) bool {
	return aoe-single > aoeMinGain
}
