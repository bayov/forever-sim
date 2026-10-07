package main

import (
	"fmt"
	"math"
	"sort"

	"github.com/wowsims/classic/sim/core/proto"
	goproto "google.golang.org/protobuf/proto"
)

// Slot trade-offs, for a PvP set. For each slot we swap in every item the slot can take,
// and every enchant the current item can take, and keep the swaps that no other swap
// beats on both damage and survival. Those are the slot's frontier: going up it, each
// step buys more health and armor for some damage. We print what each step costs in DPS
// for each 100 health, so we can decide slot by slot where survival is worth it.
//
// Survival is health plus armor at the ratio of -armor-weight to -health-weight (a
// quarter by default), all without buffs, the same numbers the weighted gear search uses.
//
// Every swap is against the rest of the set as it is. Rings and trinkets don't offer the
// item the other slot holds. A main hand only offers two handers, because a one hander
// needs a second swap for the off hand, and the off hand is skipped next to a two hander.

type tradeoff struct {
	name          string
	spec          *proto.ItemSpec
	dps, stderr   float64
	health, armor float64
	survival      float64
	current       bool
}

func (s *searcher) survivalOf(health, armor float64) float64 {
	ratio := 0.25
	if s.healthWeight != 0 {
		ratio = s.armorWeight / s.healthWeight
	}
	return health + ratio*armor
}

// The swaps no other swap beats on both damage and survival. A swap within margin DPS
// of a better one still counts as beaten only when it has no more survival.
func frontier(options []tradeoff, margin func(a, b tradeoff) float64) []tradeoff {
	var out []tradeoff
	for i, c := range options {
		beaten := false
		for j, d := range options {
			if i == j {
				continue
			}
			if d.survival >= c.survival && d.dps >= c.dps+margin(c, d) && (d.survival > c.survival || d.dps > c.dps) {
				beaten = true
				break
			}
		}
		if !beaten {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].survival < out[j].survival })
	return out
}

func (s *searcher) slotTradeoffs(pool [][]gearOption, knobs Knobs) {
	gear := equipmentSlots(s.setup.player().Equipment)
	s.setGear(gear)
	base := s.score(knobs)
	fmt.Printf("set: %s\n", base)
	baseSurvival := s.survivalOf(base.health, base.armor)

	names := make([]string, len(gear))
	for slot, item := range gear {
		if item == nil {
			continue
		}
		names[slot] = fmt.Sprint(item.Id)
		if item := itemsByID[item.Id]; item != nil {
			names[slot] = item.Name
		}
		for _, o := range pool[slot] {
			if o.spec.Id == item.Id && o.spec.RandomSuffix == item.RandomSuffix {
				names[slot] = o.name
			}
		}
	}

	for slot := range gear {
		current := gear[slot]
		if proto.ItemSlot(slot) == proto.ItemSlot_ItemSlotOffHand && isTwoHand(pool, gear[proto.ItemSlot_ItemSlotMainHand]) {
			continue
		}
		var specs []*proto.ItemSpec
		var labels []string
		add := func(spec *proto.ItemSpec, label string) {
			specs = append(specs, spec)
			labels = append(labels, label)
		}
		if current != nil {
			add(current, names[slot]+enchantSuffix(current.Enchant)+"  (now)")
			for _, e := range append([]int32{0}, enchantOptions(slot, current.Id, s.enchantExclude, s.level)...) {
				if e == current.Enchant {
					continue
				}
				spec := goproto.Clone(current).(*proto.ItemSpec)
				spec.Enchant = e
				add(spec, names[slot]+enchantSuffix(e))
			}
		}
		for _, o := range pool[slot] {
			if current != nil && o.spec.Id == current.Id && o.spec.RandomSuffix == current.RandomSuffix {
				continue
			}
			if p := pairedSlot(slot); p >= 0 && gear[p] != nil && gear[p].Id == o.spec.Id {
				continue
			}
			if proto.ItemSlot(slot) == proto.ItemSlot_ItemSlotMainHand && !o.twoHand {
				continue
			}
			spec := s.withEnchantOf(slot, o.spec, current)
			add(spec, o.name+enchantSuffix(spec.Enchant))
		}
		if len(specs) <= 1 {
			continue
		}

		scoreAt := func(spec *proto.ItemSpec, iterations int32) tradeoff {
			candidate := append([]*proto.ItemSpec(nil), gear...)
			candidate[slot] = spec
			s.setGear(candidate)
			full := s.iterations
			s.iterations = iterations
			r := s.score(knobs)
			s.iterations = full
			return tradeoff{spec: spec, dps: r.dps, stderr: r.stderr, health: r.health, armor: r.armor, survival: s.survivalOf(r.health, r.armor)}
		}

		// A quick pass over everything, then the near frontier again at full iterations.
		var screened []tradeoff
		for i, spec := range specs {
			t := scoreAt(spec, s.screenIterations)
			t.name, t.current = labels[i], i == 0 && current != nil
			screened = append(screened, t)
		}
		near := frontier(screened, func(a, b tradeoff) float64 {
			return 1.5 * math.Sqrt(a.stderr*a.stderr+b.stderr*b.stderr)
		})
		var scored []tradeoff
		for _, t := range near {
			full := scoreAt(t.spec, s.iterations)
			full.name, full.current = t.name, t.current
			scored = append(scored, full)
		}
		// The current item stays in the table even when something beats it.
		hasCurrent := false
		for _, t := range scored {
			hasCurrent = hasCurrent || t.current
		}
		if !hasCurrent && current != nil {
			t := scoreAt(current, s.iterations)
			t.name, t.current = labels[0], true
			scored = append(scored, t)
		}
		// An empty slot is the set as it is.
		if current == nil {
			scored = append(scored, tradeoff{name: "(empty)  (now)", dps: base.dps, stderr: base.stderr, health: base.health,
				armor: base.armor, survival: baseSurvival, current: true})
		}
		front := frontier(scored, func(a, b tradeoff) float64 { return 0 })
		inFront := map[string]bool{}
		for _, t := range front {
			inFront[t.name] = true
		}
		for _, t := range scored {
			if t.current && !inFront[t.name] {
				front = append(front, t)
			}
		}
		sort.Slice(front, func(i, j int) bool { return front[i].survival < front[j].survival })

		// The pick at the -health-weight and -armor-weight prices.
		bestObjective, chosen := math.Inf(-1), -1
		for i, t := range front {
			if v := s.objective(result{dps: t.dps, health: t.health, armor: t.armor}); v > bestObjective {
				bestObjective, chosen = v, i
			}
		}

		var now tradeoff
		for _, t := range front {
			if t.current {
				now = t
			}
		}
		fmt.Printf("\n%s (%d swaps, %d near the frontier)\n", gearSlotNames[slot], len(specs)-1, len(near))
		fmt.Printf("  %-58s %7s %7s %6s %7s %10s\n", "", "DPS", "±", "HP", "armor", "DPS/100HP")
		for i, t := range front {
			// The step from the option below: DPS given up for each 100 survival bought.
			step := ""
			if i > 0 {
				dv := t.survival - front[i-1].survival
				if dv > 0 {
					step = fmt.Sprintf("%10.1f", -(t.dps-front[i-1].dps)/dv*100)
				}
			}
			mark := " "
			if i == chosen && (s.healthWeight != 0 || s.armorWeight != 0) {
				mark = "*"
			}
			if !inFront[t.name] {
				mark = "x"
			}
			fmt.Printf("%s %-58.58s %+7.1f %7.1f %+6.0f %+7.0f %s\n", mark, t.name, t.dps-now.dps, t.stderr, t.health-now.health, t.armor-now.armor, step)
		}
	}
	fmt.Printf("\nset survival %.0f (health %.0f, armor %.0f)\n", baseSurvival, base.health, base.armor)
	s.setGear(gear)
}

func enchantSuffix(e int32) string {
	if e == 0 {
		return ""
	}
	return ", " + enchantName(e)
}
