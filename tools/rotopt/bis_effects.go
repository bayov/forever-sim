package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// Effect values (docs/bis_lists.md). BayovCore scores an item by stat weights, so it can't
// score an item's effects (a proc, a use, Blackhand's Breadth's mark) and says so in the
// tooltip. With -gear-effects-out we measure, for every item with an effect the sim models,
// the DPS its effects add on the BiS set, and BayovCore adds that to the item's stat score.
//
// An item's dps is the set with the item in its slot, less the same set with all of the
// item's effects left out (Player.item_effects_off), so only its stats count. A ring or
// trinket goes in place of the weaker pick, the one the set loses less without. An item with a
// use also gets its cooldown value, what its effects add while the use is on cooldown: the set
// with only the use never pressed (Player.item_uses_off), less the set with all its effects
// off. So BayovCore can swap the item's dps for it then. An item with a condition
// (core.RegisterItemConditions) gets its dps with the condition met too.
//
// The items are every one of the gear pool (the search's own filters) that has an item effect
// in the sim, and the set's picks. The list's other items come from the same pool.
//
// Every pair of runs shares its seed, so the noise is the paired one, printed for each item. The
// runs also give each kind of roll its own random source (SimOptions.use_labeled_rands). Without
// it, one Hand of Justice proc shifts every roll after it in the iteration, and the two runs pair
// about as badly as two unrelated ones. On 2026-11-20 at 45s and 2000 iterations that took the
// median noise from 2.5 to 0.6 DPS.

var (
	gearEffectsOut   = flag.String("gear-effects-out", "", "gear search: write the DPS each item's effects add on the set to this file (docs/bis_lists.md, Effect values). With -gear-rank-only and no -gear-rank-out it is all we do")
	gearEffectsIters = flag.Int("gear-effects-iters", 0, "the iterations of each -gear-effects-out run (0 is -gear-rank-confirm-iters)")
)

// One item of the effects file.
type effectsItem struct {
	Name string  `json:"name"`
	DPS  float64 `json:"dps"`
	// Cooldown is what the effects add while the use is on cooldown, for an item with a use.
	Cooldown    *float64             `json:"cooldown,omitempty"`
	Conditional []effectsConditional `json:"conditional,omitempty"`
}

type effectsConditional struct {
	Label string  `json:"label"`
	DPS   float64 `json:"dps"`
}

// writeEffects measures the effect values of the pool's items on the set best and writes them
// to path.
func (s *searcher) writeEffects(path string, pool [][]gearOption, best []*proto.ItemSpec, names []string, knobs Knobs) error {
	if len(s.durations) != 1 {
		return fmt.Errorf("-gear-effects-out needs one fight length, not %v", s.durations)
	}
	started := time.Now()
	iterations := int32(*gearEffectsIters)
	if iterations <= 0 {
		iterations = int32(*gearRankConfirmIters)
	}
	if iterations <= 0 {
		iterations = 10 * s.iterations
	}
	full := s.iterations
	encounter := s.setup.request.Encounter
	s.iterations, s.saveValues, s.labeledRands = iterations, true, true
	defer func() {
		s.iterations, s.saveValues, s.labeledRands = full, false, false
		s.setup.request.Encounter = encounter
		s.setItemSwitches(nil, nil)
		s.setGear(best)
	}()

	var pending []*pendingScore
	var done []func(result)
	run := func(gear []*proto.ItemSpec, usesOff, effectsOff []int32, enc *proto.Encounter, then func(result)) {
		s.setItemSwitches(usesOff, effectsOff)
		s.setup.request.Encounter = enc
		s.setGear(gear)
		pending = append(pending, s.prepare(knobs))
		done = append(done, then)
	}
	score := func() {
		for i, r := range s.scoreAll(pending) {
			done[i](r)
		}
		pending, done = nil, nil
	}

	// The items, one roll of each, in the first slot of the pool that has them. A pick is the
	// set's own roll.
	type candidate struct {
		id         int32
		name       string
		gear       []*proto.ItemSpec
		conditions []core.ItemCondition
		normal     result
		off        result
		useOff     *result
		cond       []result // normal and all off for each condition, in pairs
		pick       bool
	}
	isPick := func(id int32) int {
		return slices.IndexFunc(best, func(p *proto.ItemSpec) bool { return p != nil && p.Id == id })
	}
	var candidates []*candidate
	seen := map[int32]bool{}
	var skipped []string
	// A ring or trinket goes in place of the weaker pick, the one the set loses less without.
	weaker := map[int]int{}
	for _, pair := range [][2]int{{int(proto.ItemSlot_ItemSlotFinger1), int(proto.ItemSlot_ItemSlotFinger2)}, {int(proto.ItemSlot_ItemSlotTrinket1), int(proto.ItemSlot_ItemSlotTrinket2)}} {
		a, b := pair[0], pair[1]
		switch {
		case best[a] == nil:
			weaker[a], weaker[b] = a, a
		case best[b] == nil:
			weaker[a], weaker[b] = b, b
		default:
			var without [2]result
			for i, slot := range pair {
				gear := slices.Clone(best)
				gear[slot] = nil
				run(gear, nil, nil, encounter, func(r result) { without[i] = r })
			}
			score()
			w := a
			if without[1].dps > without[0].dps {
				w = b
			}
			weaker[a], weaker[b] = w, w
			fmt.Printf("effects: rings and trinkets go in place of the weaker pick, %s for %s\n", names[w], listOfName(gearSlotNames[w]))
		}
	}
	twoHandMain := best[proto.ItemSlot_ItemSlotMainHand] != nil && isTwoHand(pool, best[proto.ItemSlot_ItemSlotMainHand])
	for slot, options := range pool {
		for _, o := range options {
			id := o.spec.Id
			if seen[id] || !core.HasItemEffect(id) {
				continue
			}
			seen[id] = true
			c := &candidate{id: id, name: o.name}
			if p := isPick(id); p >= 0 {
				c.gear, c.pick = slices.Clone(best), true
				if names[p] != "" {
					c.name = names[p]
				}
			} else {
				at := slot
				if w, ok := weaker[slot]; ok {
					at = w
				}
				if slot == int(proto.ItemSlot_ItemSlotOffHand) && twoHandMain {
					// An off hand only item can't go next to the two hander, and we have no one
					// hand setup to put it in (-gear-rank-two-hand-only).
					skipped = append(skipped, o.name)
					continue
				}
				c.gear = slices.Clone(best)
				c.gear[at] = s.withEnchantOf(at, o.spec, best[at])
				if slot == int(proto.ItemSlot_ItemSlotMainHand) && o.twoHand {
					c.gear[proto.ItemSlot_ItemSlotOffHand] = nil
				}
			}
			if o.suffixStats != nil {
				// The effect doesn't depend on the roll, so the name drops the suffix.
				if it := itemsByID[id]; it != nil {
					c.name = it.Name
				}
			}
			for _, condition := range core.ItemConditions(id) {
				if !condition.MetBy(encounter) {
					c.conditions = append(c.conditions, condition)
				}
			}
			candidates = append(candidates, c)
		}
	}
	for slot, p := range best {
		if p != nil && core.HasItemEffect(p.Id) && !seen[p.Id] {
			fmt.Printf("effects: the %s pick %s isn't in the pool, left out\n", gearSlotNames[slot], names[slot])
		}
	}

	// The set itself, for the file's dps, then every item with its effects on and off.
	var base result
	run(best, nil, nil, encounter, func(r result) { base = r })
	conditionRuns := 0
	for _, c := range candidates {
		run(c.gear, nil, nil, encounter, func(r result) { c.normal = r })
		run(c.gear, nil, []int32{c.id}, encounter, func(r result) { c.off = r })
		c.cond = make([]result, 2*len(c.conditions))
		for i, condition := range c.conditions {
			met := condition.Apply(encounter)
			run(c.gear, nil, nil, met, func(r result) { c.cond[2*i] = r })
			run(c.gear, nil, []int32{c.id}, met, func(r result) { c.cond[2*i+1] = r })
			conditionRuns += 2
		}
	}
	fmt.Printf("effects: %d items with an effect at %gs, %d runs at %d iterations\n", len(candidates), s.durations[0], len(pending), iterations)
	score()
	// The items with a use, which we know now that each was in a sim.
	uses := 0
	for _, c := range candidates {
		if core.ItemHasUse(c.id) {
			run(c.gear, []int32{c.id}, nil, encounter, func(r result) { c.useOff = &r })
			uses++
		}
	}
	fmt.Printf("effects: %d items with a use, %d runs\n", uses, len(pending))
	score()

	items := map[int32]effectsItem{}
	var noise []float64
	for _, c := range candidates {
		item := effectsItem{Name: c.name, DPS: round1(c.normal.dps - c.off.dps)}
		n := gainNoise(c.off, c.normal)
		noise = append(noise, n)
		line := fmt.Sprintf("  %-40s %+6.1f ± %.1f", c.name, c.normal.dps-c.off.dps, n)
		if c.useOff != nil {
			v := round1(c.useOff.dps - c.off.dps)
			item.Cooldown = &v
			line += fmt.Sprintf("  on cooldown %+.1f ± %.1f", c.useOff.dps-c.off.dps, gainNoise(c.off, *c.useOff))
		}
		for i, condition := range c.conditions {
			on, off := c.cond[2*i], c.cond[2*i+1]
			item.Conditional = append(item.Conditional, effectsConditional{Label: condition.Label, DPS: round1(on.dps - off.dps)})
			line += fmt.Sprintf("  %s %+.1f ± %.1f", condition.Label, on.dps-off.dps, gainNoise(off, on))
		}
		fmt.Println(line)
		items[c.id] = item
	}
	for _, name := range skipped {
		fmt.Printf("effects: %s is an off hand only item, left out next to the two hander\n", name)
	}
	slices.Sort(noise)
	if len(noise) > 0 {
		fmt.Printf("effects: paired noise (1 standard error) median %.2f, max %.2f DPS\n", noise[len(noise)/2], noise[len(noise)-1])
	}
	if err := writeEffectsFile(path, s.durations[0], base.dps, iterations, items); err != nil {
		return err
	}
	fmt.Printf("effects: wrote %s, %d items in %s\n", path, len(items), time.Since(started).Round(time.Second))
	return nil
}

// writeEffectsFile writes the effects file, one item a line in ID order, so a diff of two
// runs reads item by item.
func writeEffectsFile(path string, dur, dps float64, iterations int32, items map[int32]effectsItem) error {
	var b strings.Builder
	fmt.Fprintf(&b, "{\n  \"dur\": %g,\n  \"dps\": %g,\n  \"iters\": %d,\n  \"made\": %q,\n  \"sim\": %q,\n  \"items\": {",
		dur, round2(dps), iterations, time.Now().Format("2006-01-02"), *gearRankSim)
	ids := make([]int32, 0, len(items))
	for id := range items {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for i, id := range ids {
		data, err := json.Marshal(items[id])
		if err != nil {
			return err
		}
		if i > 0 {
			b.WriteString(",")
		}
		fmt.Fprintf(&b, "\n    \"%d\": %s", id, data)
	}
	b.WriteString("\n  }\n}\n")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// round1 rounds to one decimal, without a negative zero. The effect values are a few tenths
// apart at best.
func round1(x float64) float64 {
	if r := math.Round(x*10) / 10; r != 0 {
		return r
	}
	return 0
}
