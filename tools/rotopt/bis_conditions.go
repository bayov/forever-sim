package main

import (
	"flag"
	"fmt"
	"math"
	"slices"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// Conditional ranks (docs/bis_lists.md). Some items do more in some fights: Blackhand's Breadth's
// mark doubles on Orcs and Ogres, Seal of the Dawn gives attack power only against Undead, and
// Golden Banana doubles in forest or grassland. The normal ranking is on a plain boss, where they
// get none of that, so a player farming Scholomance never learns from the list that Seal of the
// Dawn is their best trinket there.
//
// So with -gear-rank-conditions we rank each such item a second time with its condition met, and
// write where it would go on the same list. The item effects declare their conditions in the sim
// (core.RegisterItemConditions), so we never screen the whole pool under every condition. With
// -use-aoe-targets, every item with a use is ranked "vs 4 mobs" too, with only its use valued
// against the mobs (bis_aoe.go).
//
// For each condition we score the set with the condition met, and then each of its items in the
// set the way the list places it, with the condition met too. The item's conditional loss is what
// it loses against that set. A ring or trinket goes in place of either pick, and the better of
// the two counts, as on the normal list. These runs are at -gear-rank-confirm-iters and share
// their seed, since the gaps that matter are a few DPS.
//
// Its conditional rank is where that loss puts it among the list's normal entries (bisRank). The
// normal ranks never change.
//
// An item can be worth listing only for its condition: Seal of the Dawn is far down a trinket
// list against a plain boss. So we look at every conditional item of the slot's pool, not only
// the listed ones, and when its conditional rank is within -gear-rank-keep we add it to the list
// with its normal DPS, at its normal place, even past the list's length.

var gearRankConditions = flag.Bool("gear-rank-conditions", false, "gear search: with -gear-rank-out, also rank each item whose effect grows under a condition (vs Undead, in snow, ...) with the condition met, and add the ones that make the list that way (docs/bis_lists.md)")

// Where an item would go on its list with a condition met, as the list file has it.
type bisConditional struct {
	Label string  `json:"label"`
	Rank  int     `json:"rank"`
	Loss  float64 `json:"loss"`
}

// The options a list ranked, how it placed each one in the set, and the item it names in the
// other hand. A list can have several (the main hand's two handers and one handers).
type rankSource struct {
	options    []gearOption
	placements func(gearOption) [][]*proto.ItemSpec
	with       func(rankedGear) *proto.ItemSpec
}

// bisRank is the conditional rank of an item that loses loss to the set with its condition met.
// entries is its list, and self is the item's own entry there (-1 when it isn't listed), which
// doesn't count.
//
// A negative loss beats a pick, so the item is rank 1: BiS under the condition. Otherwise it comes
// after every other entry that loses as little or less. The picks lose 0, so that rank is past
// the picks, and a rank no higher than the slot's number of picks always means BiS.
func bisRank(entries []bisEntry, self int, loss float64) int {
	if loss < 0 {
		return 1
	}
	rank := 1
	for i, e := range entries {
		if i != self && e.Loss <= loss {
			rank++
		}
	}
	return rank
}

// rankConditions adds the conditional ranks to the lists, and the conditional items that make a
// list only with their condition met. sources are the lists' options (rankBisPass), and base is
// the set best's normal score.
func (s *searcher) rankConditions(slots []bisSlot, sources map[string][]rankSource, best []*proto.ItemSpec, base result, knobs Knobs) []bisSlot {
	keep := *gearRankKeep
	encounter := s.setup.request.Encounter
	defer s.meet(bisCondition{}, encounter)
	// The conditions the fight doesn't already meet. Against a plain boss that is all of them,
	// and "vs 4 mobs" for an item with a use (bis_aoe.go).
	conditionsOf := func(id int32) []bisCondition { return itemBisConditions(id, encounter) }

	// A conditional item of a list, and its scores.
	type candidate struct {
		list   int
		source rankSource
		option gearOption
		// conditions are the item's, and conditional[i] is its best score with conditions[i]
		// met.
		conditions  []bisCondition
		conditional []*result
		// normal is its best score against a plain boss and gear the set it was scored in, for
		// an item the list doesn't have yet.
		normal *result
		gear   []*proto.ItemSpec
		listed bool
	}
	type named struct {
		id   int32
		name string
	}
	var candidates []*candidate
	for i, slot := range slots {
		listed, seen := map[named]bool{}, map[named]bool{}
		for _, e := range slot.entries {
			listed[named{e.ID, e.Name}] = true
		}
		for _, src := range sources[slot.name] {
			for _, o := range bestRolls(src.options) {
				key := named{o.spec.Id, o.name}
				conditions := conditionsOf(o.spec.Id)
				if len(conditions) == 0 || seen[key] {
					continue
				}
				seen[key] = true
				candidates = append(candidates, &candidate{list: i, source: src, option: o, conditions: conditions,
					conditional: make([]*result, len(conditions)), listed: listed[key]})
			}
		}
	}

	// We prepare every run first and score them side by side. A run takes the encounter and the
	// gear as they are when it is prepared (prepare).
	iterations := int32(*gearRankConfirmIters)
	if iterations <= 0 {
		iterations = 10 * s.iterations
	}
	full := s.iterations
	s.saveValues = true
	defer func() { s.iterations, s.saveValues = full, false }()
	var pending []*pendingScore
	var done []func(result)
	run := func(gear []*proto.ItemSpec, then func(result)) {
		s.setGear(gear)
		pending = append(pending, s.prepare(knobs))
		done = append(done, then)
	}
	keepBest := func(at **result) func(result) {
		return func(r result) {
			if *at == nil || r.dps > (*at).dps {
				*at = &r
			}
		}
	}
	// The set with each condition met, measured once for all its items. A pick with a use needs
	// it for "vs 4 mobs" too, to tell whether its use gains on the mobs.
	var conditions []bisCondition
	for _, c := range candidates {
		for _, condition := range c.conditions {
			if !slices.Contains(conditions, condition) {
				conditions = append(conditions, condition)
			}
		}
	}
	if aoe, ok := aoeCondition(); ok && !slices.Contains(conditions, aoe) &&
		slices.ContainsFunc(slots, func(slot bisSlot) bool {
			return slices.ContainsFunc(slot.entries, func(e bisEntry) bool { return e.BiS && core.ItemHasUse(e.ID) })
		}) {
		conditions = append(conditions, aoe)
	}
	bases := make([]result, len(conditions))
	s.iterations = iterations
	for i, condition := range conditions {
		s.meet(condition, encounter)
		run(best, func(r result) { bases[i] = r })
		for _, c := range candidates {
			if j := slices.Index(c.conditions, condition); j >= 0 {
				for _, gear := range c.source.placements(c.option) {
					if gear != nil {
						run(gear, keepBest(&c.conditional[j]))
					}
				}
			}
		}
	}
	// An item the list doesn't have needs its normal score too, at the list's iterations, next
	// to the base the list was ranked on.
	s.meet(bisCondition{}, encounter)
	s.iterations = full
	for _, c := range candidates {
		if !c.listed {
			for _, gear := range c.source.placements(c.option) {
				if gear != nil {
					run(gear, func(r result) {
						if c.normal == nil || r.dps > c.normal.dps {
							c.normal, c.gear = &r, gear
						}
					})
				}
			}
		}
	}
	if len(pending) == 0 {
		return slots
	}
	fmt.Printf("ranking %d conditional items under %d conditions at %d iterations:\n", len(candidates), len(conditions), iterations)
	for i, r := range s.scoreAll(pending) {
		done[i](r)
	}

	// The conditional entries of an item, ranked on the list as it is. single is its use value
	// against the boss. A "vs 4 mobs" line is left out when the use gains too little on the mobs
	// (aoeGains).
	conditionalOf := func(c *candidate, entries []bisEntry, self int, single float64) []bisConditional {
		var out []bisConditional
		id := c.option.spec.Id
		for j, condition := range c.conditions {
			r := c.conditional[j]
			if r == nil {
				continue
			}
			if condition.useTargets > 0 {
				if !aoeGains(single, r.uses[id]) {
					fmt.Printf("  %s %s: its use is worth %+.1f, %+.1f against the boss, no line\n", c.option.name, condition.label(), r.uses[id], single)
					continue
				}
				fmt.Printf("  %s %s: its use is worth %+.1f, %+.1f against the boss\n", c.option.name, condition.label(), r.uses[id], single)
			}
			loss := round2(bases[slices.Index(conditions, condition)].dps - r.dps)
			out = append(out, bisConditional{Label: condition.label(), Rank: bisRank(entries, self, loss), Loss: loss})
		}
		return out
	}
	// First the items the list doesn't have, ranked on the list without them. One that makes the
	// list under any of its conditions goes in at the place its normal DPS gives it.
	for _, c := range candidates {
		if c.listed || c.normal == nil {
			continue
		}
		slot := &slots[c.list]
		conditional := conditionalOf(c, slot.entries, -1, c.normal.uses[c.option.spec.Id])
		for _, e := range conditional {
			fmt.Printf("  %s: %s %s: rank %d (%+.1f), %s normally\n", slot.name, c.option.name, e.Label, e.Rank, 0-e.Loss, *c.normal)
		}
		if !slices.ContainsFunc(conditional, func(e bisConditional) bool { return e.Rank <= keep }) {
			continue
		}
		entry := bisEntry{ID: c.option.spec.Id, Suffix: c.option.spec.RandomSuffix, Name: c.option.name,
			DPS: round2(c.normal.dps), Loss: round2(base.dps - c.normal.dps), Unsourced: c.option.unsourced, use: c.normal.uses[c.option.spec.Id]}
		if with := c.source.with(rankedGear{c.option, *c.normal, c.gear}); with != nil {
			entry.With = with.Id
		}
		at := len(slot.entries)
		for i, e := range slot.entries {
			if !e.BiS && e.DPS < entry.DPS {
				at = i
				break
			}
		}
		slot.entries = slices.Insert(slot.entries, at, entry)
		c.listed = true
	}
	// Then every listed one, on the list as it now is.
	for _, c := range candidates {
		if !c.listed {
			continue
		}
		slot := &slots[c.list]
		self := slices.IndexFunc(slot.entries, func(e bisEntry) bool { return e.ID == c.option.spec.Id && e.Name == c.option.name })
		slot.entries[self].Conditional = conditionalOf(c, slot.entries, self, slot.entries[self].use)
	}
	// A pick with a condition stays where it is, and loses nothing, since the set already holds
	// it. It gets a "vs 4 mobs" line only when its use gains on the mobs, as the others do.
	for i := range slots {
		for j, e := range slots[i].entries {
			if !e.BiS {
				continue
			}
			for _, condition := range conditionsOf(e.ID) {
				if condition.useTargets > 0 {
					aoe := bases[slices.Index(conditions, condition)].uses[e.ID]
					gains := aoeGains(e.use, aoe)
					fmt.Printf("  %s %s: its use is worth %+.1f, %+.1f against the boss, line %t\n", e.Name, condition.label(), aoe, e.use, gains)
					if !gains {
						continue
					}
				}
				slots[i].entries[j].Conditional = append(slots[i].entries[j].Conditional, bisConditional{Label: condition.label(), Rank: j + 1})
			}
		}
	}
	return slots
}

// round2 rounds to the list's two decimals, without a negative zero.
func round2(x float64) float64 {
	if r := math.Round(x*100) / 100; r != 0 {
		return r
	}
	return 0
}
