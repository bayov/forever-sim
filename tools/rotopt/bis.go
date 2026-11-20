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

	"github.com/wowsims/classic/sim/core/proto"
)

// BiS lists (docs/bis_lists.md). Once the gear search has settled on a set, we rank what
// else each slot could hold. Every item the pool offers for the slot goes into the set in
// place of the pick, with the rest of the set as it is, and we write down its DPS and what
// it loses against the set. A player who can't get the pick for a slot reads down the
// list to the next item they can get.
//
// The swaps follow the gear search's own rules. A swap carries the slot's enchant over,
// the two rings and the two trinkets never hold the same item, a two hander empties the
// off hand, and an item a character can only hold one of never goes in both hands.
//
// Rings and trinkets are one list each. The two picks come first, and every other item is
// scored in both slots and listed with the better of the two, since a player would put it
// in place of the pick it loses less to.
//
// A one hander in place of a two hander needs an off hand, or it loses the off hand's
// stats and looks far worse than it is. So when the pick is a two hander, we first find the
// best one hand setup: the best one hander with nothing in the off hand, then the best off
// hand next to it, and then we score every one hander next to that off hand. Each one hand
// entry names the off hand it was scored with. The off hand list is then what we'd hold if
// we went one hand, scored next to that one hander, and none of its entries is a pick.
//
// With -gear-rank-two-hand-only we skip the one hand setup. When the pick is a two hander,
// the main hand list has only two handers and the off hand list is empty, since the off hand
// stays empty while we wield a two hander. The Phase 0 list uses it.

var (
	gearRankOut  = flag.String("gear-rank-out", "", "gear search: once the search settles, rank every slot's options against the set and write the BiS list (docs/bis_lists.md) to this file")
	gearRankKeep = flag.Int("gear-rank-keep", 10, "gear search: how many items each slot of the -gear-rank-out list keeps, the BiS picks included")
	gearRankName = flag.String("gear-rank-name", "", "the -gear-rank-out list's label, e.g. \"Level 30\"")
	gearRankSpec = flag.String("gear-rank-spec", "", "the -gear-rank-out list's spec, by its mksettings name (enh, ele, ...)")
	gearRankSim  = flag.String("gear-rank-sim", "", "the forever-sim commit the -gear-rank-out list was simmed on, for the list's sim field")

	gearRankConfirmIters = flag.Int("gear-rank-confirm-iters", 0, "gear search: the iterations that confirm an item the -gear-rank-out ranking puts ahead of the set's pick before it takes the pick's place (0 is 10 times -iters)")

	gearRankOnly = flag.Bool("gear-rank-only", false, "gear search: skip the search and rank the starting gear as the set for -gear-rank-out, to redo a list's ranking from its picks (bis_phase0.sh RANK_ONLY=1)")

	gearRankSlots = flag.String("gear-rank-slots", "", "with -gear-rank-only and -gear-rank-from: rank only these lists (head, ..., finger, trinket, mainhand, offhand, ranged, comma separated) and copy every other list and enchant list from the -gear-rank-from list")
	gearRankFrom  = flag.String("gear-rank-from", "", "the list -gear-rank-slots copies the other slots from, usually the one whose picks are the starting gear")

	gearRankTwoHandOnly = flag.Bool("gear-rank-two-hand-only", false, "gear search: when the -gear-rank-out main hand pick is a two hander, list only two handers in the main hand and leave the off hand list empty")
)

// One line of a slot's list, as the list file has it.
type bisEntry struct {
	ID        int32   `json:"id"`
	Suffix    int32   `json:"suffix,omitempty"`
	Name      string  `json:"name"`
	DPS       float64 `json:"dps"`
	Loss      float64 `json:"loss"`
	Unsourced bool    `json:"unsourced,omitempty"`
	// With is the item in the other hand: the off hand a one hand main hand was scored
	// with, or the main hand an off hand was scored with.
	With int32 `json:"with,omitempty"`
	BiS  bool  `json:"bis,omitempty"`
	// Enchant is a pick's enchant, by its effect ID in the sim (the enchants list's id).
	Enchant int32 `json:"enchant,omitempty"`
	// Conditional is where the item would go on this list with a condition met that makes its
	// effect bigger, one for each condition (-gear-rank-conditions, bis_conditions.go).
	Conditional []bisConditional `json:"conditional,omitempty"`
	// Cooldown is where an item with a use would go on this list with its use never pressed,
	// as while it is on cooldown (-gear-rank-cooldowns, bis_cooldowns.go).
	Cooldown *bisCooldown `json:"cooldown,omitempty"`

	// use is the item's use value in the set it was scored in, with -use-dur (use_values.go).
	use float64
}

type bisSlot struct {
	name    string
	entries []bisEntry
}

// An option and its score in the best place we found for it, and the gear set of that place.
type rankedGear struct {
	option gearOption
	result result
	gear   []*proto.ItemSpec
}

// rankGear scores each option and returns the best of them, best first. It keeps keep
// items by name, the way trimBis counts them: the rolls of one item and suffix name count
// once.
//
// placements gives the gear sets an option can be scored in, and the option counts with
// the best of them. A ring is placed in either ring slot, say. An option with no
// placement is left out.
//
// empty gives the gear sets with the slot left empty, or nil when we don't compare with an
// empty slot. An option that scores no better than the best of them does nothing in the
// sim (a Hook of the Master Angler in a trinket slot), so it isn't worth listing.
//
// A slot can have hundreds of options, so we don't score them all at full iterations. We
// race them instead (rankStages): everything at screenIterations first, then what is left
// at more iterations, and the last few at full iterations. Between two steps we drop an
// option only when it is clearly out: behind the keep-th best, or behind the empty slot,
// by more than confidence standard errors of the gap (raceCut). The runs share their seed,
// and we keep each iteration's DPS, so that error is the paired one (gainNoise).
//
// We used to keep a plain top 50 of the 100 iteration screen. It lost real items, because a
// screen that short is several DPS off, and the trinkets that matter are within a few DPS
// of each other and of an empty slot. That cut lost Blackhand's Breadth from the Phase 0
// list of 2026-11-13, which is 7.7 DPS better than an empty slot at 2000 iterations.
func (s *searcher) rankGear(label string, options []gearOption, placements func(gearOption) [][]*proto.ItemSpec, empty [][]*proto.ItemSpec, keep int, knobs Knobs) []rankedGear {
	if keep <= 0 {
		return nil
	}
	counts := []string{fmt.Sprint(len(options))}
	options = bestRolls(options)
	counts = append(counts, fmt.Sprintf("%d best rolls", len(options)))
	stages := rankStages(s.screenIterations, s.iterations)
	if empty == nil && len(options) <= keep {
		stages = stages[len(stages)-1:]
	}
	var ranked []rankedGear
	for i, iterations := range stages {
		var floor *result
		ranked, floor = s.scoreOptions(options, placements, empty, iterations, knobs)
		if i == len(stages)-1 {
			if floor != nil {
				ranked = slices.DeleteFunc(ranked, func(r rankedGear) bool { return r.result.dps <= floor.dps })
			}
			ranked = topByName(ranked, keep)
			break
		}
		options = raceCut(ranked, floor, keep, s.confidence)
		counts = append(counts, fmt.Sprintf("%d after %d iterations", len(options), iterations))
	}
	if label != "" {
		fmt.Printf("  %s: %s, %d listed\n", label, strings.Join(counts, ", "), len(ranked))
	}
	return ranked
}

// rankStages is the iterations of each step of rankGear's race: screen, about the geometric
// mean of screen and full, and full. 100 and 2000 give 100, 447 and 2000.
func rankStages(screen, full int32) []int32 {
	if screen <= 0 || screen >= full {
		return []int32{full}
	}
	mid := int32(math.Sqrt(float64(screen) * float64(full)))
	if mid < 2*screen || 2*mid > full {
		return []int32{screen, full}
	}
	return []int32{screen, mid, full}
}

// scoreOptions scores every option in the best of its placements at the iterations, best
// first, and the empty slot (nil without empty sets). The sims run side by side (scoreAll).
func (s *searcher) scoreOptions(options []gearOption, placements func(gearOption) [][]*proto.ItemSpec, empty [][]*proto.ItemSpec,
	iterations int32, knobs Knobs) ([]rankedGear, *result) {
	full := s.iterations
	s.iterations, s.saveValues = iterations, true
	defer func() { s.iterations, s.saveValues = full, false }()
	var pending []*pendingScore
	// owner[i] is the option pending[i] belongs to, and -1 for an empty slot. gears[i] is the
	// gear set it was scored in.
	var owner []int
	var gears [][]*proto.ItemSpec
	for i, o := range options {
		for _, gear := range placements(o) {
			if gear != nil {
				s.setGear(gear)
				pending = append(pending, s.prepare(knobs))
				owner = append(owner, i)
				gears = append(gears, gear)
			}
		}
	}
	for _, gear := range empty {
		s.setGear(gear)
		pending = append(pending, s.prepare(knobs))
		owner = append(owner, -1)
		gears = append(gears, gear)
	}
	best := make([]*result, len(options))
	bestGear := make([][]*proto.ItemSpec, len(options))
	var floor *result
	for i, r := range s.scoreAll(pending) {
		at := &floor
		if owner[i] >= 0 {
			at = &best[owner[i]]
		}
		if *at == nil || r.dps > (*at).dps {
			*at = &r
			if owner[i] >= 0 {
				bestGear[owner[i]] = gears[i]
			}
		}
	}
	var ranked []rankedGear
	for i, r := range best {
		if r != nil {
			ranked = append(ranked, rankedGear{options[i], *r, bestGear[i]})
		}
	}
	sort.SliceStable(ranked, func(i, j int) bool { return ranked[i].result.dps > ranked[j].result.dps })
	return ranked, floor
}

// raceCut is the options of ranked (best first) worth scoring at more iterations. It drops
// an option that is behind the keep-th best by name, or behind the empty slot floor, by
// more than confidence standard errors of the gap. It also drops an option that scores
// exactly what the empty slot scores. The runs share their random seed, so that item
// changes nothing the sim sees (Stamina and Spirit trinkets, say).
func raceCut(ranked []rankedGear, floor *result, keep int, confidence float64) []gearOption {
	var kth *result
	if i := keepthByName(ranked, keep); i >= 0 {
		kth = &ranked[i].result
	}
	behind := func(r, ahead result) bool { return r.dps+confidence*gainNoise(r, ahead) < ahead.dps }
	var out []gearOption
	for _, r := range ranked {
		if floor != nil && (r.result.dps == floor.dps || behind(r.result, *floor)) {
			continue
		}
		if kth != nil && behind(r.result, *kth) {
			continue
		}
		out = append(out, r.option)
	}
	return out
}

// topByName is the first keep options of ranked (best first) by name: a weaker roll of an
// item and suffix name already listed comes along but doesn't count, since trimBis drops it.
func topByName(ranked []rankedGear, keep int) []rankedGear {
	if i := keepthByName(ranked, keep+1); i >= 0 {
		return ranked[:i]
	}
	return ranked
}

// keepthByName is the index in ranked (best first) of the keep-th option by name, or -1
// when ranked has fewer names.
func keepthByName(ranked []rankedGear, keep int) int {
	type named struct {
		id   int32
		name string
	}
	seen := map[named]bool{}
	for i, r := range ranked {
		key := named{r.option.spec.Id, r.option.name}
		if !seen[key] {
			seen[key] = true
			if len(seen) == keep {
				return i
			}
		}
	}
	return -1
}

// bestRolls drops the weaker rolls of each random suffix. Two rolls of an item with the same
// suffix name ("of the Tiger") only differ in how much of each stat they give, so a roll with
// no more of any stat than another one can't do better, and we don't sim it. A slot of
// greens has hundreds of rolls, and most of them go this way.
func bestRolls(options []gearOption) []gearOption {
	// dominates is whether a gives at least as much of every stat as b. Of two equal rolls
	// the first one stays.
	dominates := func(a, b gearOption) bool {
		for i := range max(len(a.suffixStats), len(b.suffixStats)) {
			var x, y float64
			if i < len(a.suffixStats) {
				x = a.suffixStats[i]
			}
			if i < len(b.suffixStats) {
				y = b.suffixStats[i]
			}
			if x < y {
				return false
			}
		}
		return true
	}
	var out []gearOption
	for i, o := range options {
		weaker := o.suffixStats != nil && slices.ContainsFunc(options, func(other gearOption) bool {
			if other.suffixStats == nil || other.spec.Id != o.spec.Id || other.name != o.name || other.spec.RandomSuffix == o.spec.RandomSuffix {
				return false
			}
			if !dominates(other, o) {
				return false
			}
			// Equal rolls: the first one stays.
			return !dominates(o, other) || slices.IndexFunc(options, func(x gearOption) bool { return sameGear(x.spec, other) }) < i
		})
		if !weaker {
			out = append(out, o)
		}
	}
	return out
}

// Whether the option is the item and suffix of spec.
func sameGear(spec *proto.ItemSpec, o gearOption) bool {
	return spec != nil && spec.Id == o.spec.Id && spec.RandomSuffix == o.spec.RandomSuffix
}

// rankBis ranks every slot's options against the set best and writes the list to path.
// names and unsourced are the set's own, as the search printed them.
//
// The list's picks follow its own ranking. When a slot's ranking shows an item ahead of the
// pick and a confirm at more iterations holds it up (followRanking), we put it in the set and
// rank everything again on the new set.
// The gear search can keep an item the ranking then puts behind, because the ranking runs on
// the set's best enchants and races every item of the pool, where the search screened them
// at a few hundred iterations. On 2026-11-14 the Phase 0 list kept Wildheart Bracers while
// its own wrist ranking had Windtalker's Wristguards 2.5 DPS ahead.
func (s *searcher) rankBis(path string, pool [][]gearOption, best []*proto.ItemSpec, names []string, unsourced []bool, knobs Knobs) error {
	from, err := loadRankFrom()
	if err != nil {
		return err
	}
	if *gearRankCooldowns && !s.useScores {
		return fmt.Errorf("-gear-rank-cooldowns needs -use-dur")
	}
	for pass := 1; ; pass++ {
		enchants, base, slots, sources, challengers, err := s.rankBisPass(pool, best, names, unsourced, knobs)
		if err != nil {
			return err
		}
		if pass < maxRankPasses && s.followRanking(pool, best, names, unsourced, base, challengers, knobs) {
			fmt.Println("ranking again on the new set:")
			continue
		}
		if pass == maxRankPasses && len(challengers) > 0 {
			fmt.Printf("  stopped after %d rankings, the picks may still be behind:\n", pass)
			for _, c := range challengers {
				fmt.Printf("    %s: %s %s\n", c.slot, c.ranked.option.name, c.ranked.result)
			}
		}
		if *gearRankConditions {
			slots = s.rankConditions(slots, sources, best, base, knobs)
		}
		if *gearRankCooldowns {
			slots = rankCooldowns(slots)
		}
		s.setGear(best)
		for _, slot := range slots {
			fmt.Printf("  %s:\n", slot.name)
			for _, e := range slot.entries {
				fmt.Printf("    %6.1f %6.1f  %s", e.DPS, 0-e.Loss, e.Name)
				if e.With != 0 {
					if it := itemsByID[e.With]; it != nil {
						fmt.Printf(" (with %s)", it.Name)
					}
				}
				if e.Unsourced {
					fmt.Print("  no source")
				}
				for _, c := range e.Conditional {
					fmt.Printf("  %s: rank %d (%+.1f)", c.Label, c.Rank, 0-c.Loss)
				}
				if c := e.Cooldown; c != nil {
					fmt.Printf("  on cooldown: rank %d (%+.1f)", c.Rank, 0-c.Loss)
				}
				fmt.Println()
			}
		}
		return s.writeBis(path, base, slots, enchants, from)
	}
}

// How many times rankBis ranks the slots at most: the first ranking and two more after picks
// changed. The confirmed gains are real, so a third change of picks is rare, and each
// ranking takes several minutes.
const maxRankPasses = 3

// An item a slot's ranking puts ahead of the set's pick, in the gear set it was scored in.
type rankChallenger struct {
	slot   string
	ranked rankedGear
}

// rankBisPass ranks every slot's options against the set best, with the best enchants first
// when -gear-rank-enchants asks for them. challengers are the slots' best items that beat
// the set at the ranking's iterations, by any margin.
//
// sources are the options each list ranked and how it placed them, for rankConditions.
func (s *searcher) rankBisPass(pool [][]gearOption, best []*proto.ItemSpec, names []string, unsourced []bool, knobs Knobs) (
	enchants []bisEnchantSlot, base result, slots []bisSlot, sources map[string][]rankSource, challengers []rankChallenger, err error) {
	keep := *gearRankKeep
	sources = map[string][]rankSource{}
	source := func(list string, options []gearOption, placements func(gearOption) [][]*proto.ItemSpec, with func(rankedGear) *proto.ItemSpec) {
		sources[list] = append(sources[list], rankSource{options, placements, with})
	}
	// With -gear-rank-enchants the set gets its best enchants first, so the items are
	// ranked on them (bis_enchants.go).
	if *gearRankEnchants != "" {
		fmt.Println("ranking every enchant of the set:")
		if enchants, err = s.rankEnchants(best, knobs); err != nil {
			return
		}
	}
	// The base keeps each iteration's DPS, so a challenger's gain has its paired error. The
	// last step of rankGear's race runs at the same iterations.
	//
	// With -use-dur the use values are measured again on this base, so the base's own uses are
	// valued on the set itself.
	s.resetUseValues()
	s.setGear(best)
	s.saveValues = true
	base = s.score(knobs)
	s.saveValues = false
	fmt.Printf("ranking every slot against the set: %s\n", base)
	// ahead notes a slot's best item when it beats the set at all. ranked is best first.
	//
	// We don't ask for a clear margin here, because the confirm step does that at many more
	// iterations. On 2026-11-14 a margin of two standard errors at the ranking's iterations
	// let Don Julio's Band stay 2.3 DPS ahead of the Phase 0 ring picks, and it was 2.4 ahead
	// at 20000 iterations.
	ahead := func(slot string, ranked []rankedGear) {
		if len(ranked) == 0 || ranked[0].gear == nil {
			return
		}
		if ranked[0].result.dps > base.dps {
			challengers = append(challengers, rankChallenger{slot, ranked[0]})
		}
	}

	entryOf := func(o gearOption, r result, with *proto.ItemSpec) bisEntry {
		e := bisEntry{ID: o.spec.Id, Suffix: o.spec.RandomSuffix, Name: o.name, DPS: round2(r.dps), Loss: round2(base.dps - r.dps), Unsourced: o.unsourced, use: r.uses[o.spec.Id]}
		if with != nil {
			e.With = with.Id
		}
		return e
	}
	// The set's own item in a slot, as the first entries of its list.
	pickOf := func(slot int, with *proto.ItemSpec) []bisEntry {
		item := best[slot]
		if item == nil {
			return nil
		}
		name := names[slot]
		if name == "" {
			if it := itemsByID[item.Id]; it != nil {
				name = it.Name
			}
		}
		e := bisEntry{ID: item.Id, Suffix: item.RandomSuffix, Name: name, DPS: round2(base.dps), Unsourced: unsourced[slot], BiS: true, Enchant: item.Enchant, use: base.uses[item.Id]}
		if with != nil {
			e.With = with.Id
		}
		return []bisEntry{e}
	}
	entriesOf := func(ranked []rankedGear, with func(rankedGear) *proto.ItemSpec) []bisEntry {
		var out []bisEntry
		for _, r := range ranked {
			out = append(out, entryOf(r.option, r.result, with(r)))
		}
		return out
	}
	// The set with spec in the slot, carrying the slot's enchant over.
	swapped := func(slot int, spec *proto.ItemSpec) []*proto.ItemSpec {
		candidate := append([]*proto.ItemSpec(nil), best...)
		candidate[slot] = s.withEnchantOf(slot, spec, best[slot])
		return candidate
	}
	others := func(slot int, picks ...*proto.ItemSpec) []gearOption {
		return slices.DeleteFunc(slices.Clone(pool[slot]), func(o gearOption) bool {
			return slices.ContainsFunc(picks, func(p *proto.ItemSpec) bool { return sameGear(p, o) })
		})
	}

	// We rank more names than we keep, so an item near the end of the list is never the one
	// rankGear's race cuts at.
	wide := keep + keep/2
	// The set with the slots emptied one at a time, for rankGear's empty slot. For the rings
	// and trinkets we empty whichever slot loses less.
	emptied := func(slots ...int) [][]*proto.ItemSpec {
		var out [][]*proto.ItemSpec
		for _, slot := range slots {
			candidate := append([]*proto.ItemSpec(nil), best...)
			candidate[slot] = nil
			out = append(out, candidate)
		}
		return out
	}
	single := func(slot int) {
		if !rankListSelected(gearSlotNames[slot]) {
			return
		}
		picks := pickOf(slot, nil)
		options := others(slot, best[slot])
		placements := func(o gearOption) [][]*proto.ItemSpec {
			return [][]*proto.ItemSpec{swapped(slot, o.spec)}
		}
		source(gearSlotNames[slot], options, placements, noOtherHand)
		ranked := s.rankGear(gearSlotNames[slot], options, placements, emptied(slot), wide, knobs)
		ahead(gearSlotNames[slot], ranked)
		slots = append(slots, bisSlot{gearSlotNames[slot], trimBis(append(picks, entriesOf(ranked, noOtherHand)...), keep)})
	}
	pair := func(name string, a, b int) {
		if !rankListSelected(name) {
			return
		}
		picks := append(pickOf(a, nil), pickOf(b, nil)...)
		// Each slot takes any item but the one the other slot holds.
		options := others(a, best[a], best[b])
		placements := func(o gearOption) [][]*proto.ItemSpec {
			var out [][]*proto.ItemSpec
			if best[b] == nil || best[b].Id != o.spec.Id {
				out = append(out, swapped(a, o.spec))
			}
			if best[a] == nil || best[a].Id != o.spec.Id {
				out = append(out, swapped(b, o.spec))
			}
			return out
		}
		source(name, options, placements, noOtherHand)
		ranked := s.rankGear(name, options, placements, emptied(a, b), wide, knobs)
		ahead(name, ranked)
		slots = append(slots, bisSlot{name, trimBis(append(picks, entriesOf(ranked, noOtherHand)...), keep)})
	}

	for slot := range 10 {
		single(slot)
	}
	pair("finger", int(proto.ItemSlot_ItemSlotFinger1), int(proto.ItemSlot_ItemSlotFinger2))
	pair("trinket", int(proto.ItemSlot_ItemSlotTrinket1), int(proto.ItemSlot_ItemSlotTrinket2))
	// The two hands are ranked together, since each list is scored next to the other hand.
	if rankListSelected("mainhand") || rankListSelected("offhand") {
		main, off := s.rankWeapons(pool, best, wide, knobs, pickOf, entriesOf, base, ahead, source)
		for _, list := range []bisSlot{{"mainhand", trimBis(main, keep)}, {"offhand", trimBis(off, keep)}} {
			if rankListSelected(list.name) {
				slots = append(slots, list)
			}
		}
	}
	single(int(proto.ItemSlot_ItemSlotRanged))
	return
}

// followRanking puts the challengers that hold up in the set, and says whether the set
// changed. Each challenger beat the set at the ranking's iterations, so we score it again at
// -gear-rank-confirm-iters next to the set, and it counts only when it still beats the set by
// more than confidence paired standard errors there. Then the confirmed ones go in, the
// biggest gain first, each on slots no other one changed. When two changes together score
// worse than the biggest one alone (two rings that each needed the hit the other one gives),
// only the biggest goes in, and the next ranking finds the rest again.
func (s *searcher) followRanking(pool [][]gearOption, best []*proto.ItemSpec, names []string, unsourced []bool,
	base result, challengers []rankChallenger, knobs Knobs) bool {
	if len(challengers) == 0 {
		return false
	}
	iterations := int32(*gearRankConfirmIters)
	if iterations <= 0 {
		iterations = 10 * s.iterations
	}
	full := s.iterations
	s.iterations, s.saveValues = iterations, true
	defer func() { s.iterations, s.saveValues = full, false }()
	scoreSets := func(sets ...[]*proto.ItemSpec) []result {
		var pending []*pendingScore
		for _, gear := range sets {
			s.setGear(gear)
			pending = append(pending, s.prepare(knobs))
		}
		return s.scoreAll(pending)
	}
	sets := [][]*proto.ItemSpec{best}
	for _, c := range challengers {
		sets = append(sets, c.ranked.gear)
	}
	results := scoreSets(sets...)
	type confirmed struct {
		rankChallenger
		gain float64
	}
	var good []confirmed
	for i, c := range challengers {
		r := results[i+1]
		gain, noise := r.dps-results[0].dps, gainNoise(results[0], r)
		verdict := "not confirmed"
		if gain > s.confidence*noise {
			verdict = "confirmed"
			good = append(good, confirmed{c, gain})
		}
		fmt.Printf("  %s: %s is %+.1f (%.1f) over the pick at %d iterations, %s\n", c.slot, c.ranked.option.name, gain, noise, iterations, verdict)
	}
	if len(good) == 0 {
		return false
	}
	sort.SliceStable(good, func(i, j int) bool { return good[i].gain > good[j].gain })
	changes := func(gear []*proto.ItemSpec) []int { return gearChanges(best, gear) }
	apply := func(gear []*proto.ItemSpec, picks []confirmed) []*proto.ItemSpec {
		out := slices.Clone(gear)
		for _, c := range picks {
			for _, slot := range changes(c.ranked.gear) {
				out[slot] = c.ranked.gear[slot]
			}
		}
		return out
	}
	var gears [][]*proto.ItemSpec
	for _, c := range good {
		gears = append(gears, c.ranked.gear)
	}
	var picks []confirmed
	for _, i := range disjointChanges(best, gears) {
		picks = append(picks, good[i])
	}
	if len(picks) > 1 {
		both := scoreSets(apply(best, picks), apply(best, picks[:1]))
		if both[0].dps <= both[1].dps {
			fmt.Printf("  the changes together score %.1f, below the biggest alone at %.1f, so only that one goes in\n", both[0].dps, both[1].dps)
			picks = picks[:1]
		}
	}
	// The names of the new items: a challenger's own item, or the item its gear set holds in
	// another slot (the off hand of a one hand setup).
	nameOf := func(slot int, spec *proto.ItemSpec) (string, bool) {
		for _, o := range pool[slot] {
			if sameGear(spec, o) {
				return o.name, o.unsourced
			}
		}
		if it := itemsByID[spec.Id]; it != nil {
			return it.Name, false
		}
		return "", false
	}
	next := apply(best, picks)
	for _, c := range picks {
		for _, slot := range changes(c.ranked.gear) {
			if next[slot] == nil {
				names[slot], unsourced[slot] = "", false
			} else {
				names[slot], unsourced[slot] = nameOf(slot, next[slot])
			}
			old := "nothing"
			if best[slot] != nil {
				old, _ = nameOf(slot, best[slot])
			}
			fmt.Printf("  %s: %s takes the place of %s (%+.1f)\n", gearSlotNames[slot], names[slot], old, c.gain)
		}
	}
	copy(best, next)
	return true
}

// gearChanges is the slots where gear differs from best: another item, roll or enchant.
func gearChanges(best, gear []*proto.ItemSpec) []int {
	var out []int
	for slot := range best {
		a, b := best[slot], gear[slot]
		if (a == nil) != (b == nil) || a != nil && (a.Id != b.Id || a.RandomSuffix != b.RandomSuffix || a.Enchant != b.Enchant) {
			out = append(out, slot)
		}
	}
	return out
}

// disjointChanges is the indexes of the gear sets (best first) whose changes from best can go
// in together: each one in turn, unless it changes a slot an earlier one already changed.
func disjointChanges(best []*proto.ItemSpec, gears [][]*proto.ItemSpec) []int {
	taken := map[int]bool{}
	var out []int
	for i, gear := range gears {
		slots := gearChanges(best, gear)
		if slices.ContainsFunc(slots, func(slot int) bool { return taken[slot] }) {
			continue
		}
		for _, slot := range slots {
			taken[slot] = true
		}
		out = append(out, i)
	}
	return out
}

// rankWeapons is the main hand and off hand lists. See the top of the file for how a one
// hander is scored next to a two hander pick.
func (s *searcher) rankWeapons(pool [][]gearOption, best []*proto.ItemSpec, keep int, knobs Knobs,
	pickOf func(int, *proto.ItemSpec) []bisEntry, entriesOf func([]rankedGear, func(rankedGear) *proto.ItemSpec) []bisEntry, base result,
	ahead func(string, []rankedGear), source func(string, []gearOption, func(gearOption) [][]*proto.ItemSpec, func(rankedGear) *proto.ItemSpec)) ([]bisEntry, []bisEntry) {
	mh, oh := int(proto.ItemSlot_ItemSlotMainHand), int(proto.ItemSlot_ItemSlotOffHand)
	// The set with o in the main hand and offHand in the off hand. A two hander leaves the
	// off hand empty.
	withMainHand := func(o gearOption, offHand *proto.ItemSpec) []*proto.ItemSpec {
		if o.twoHand {
			offHand = nil
		}
		if o.once && offHand != nil && offHand.Id == o.spec.Id {
			return nil
		}
		candidate := append([]*proto.ItemSpec(nil), best...)
		candidate[mh] = s.withEnchantOf(mh, o.spec, best[mh])
		candidate[oh] = offHand
		return candidate
	}
	// The set with mainHand in the main hand and o in the off hand.
	withOffHand := func(o gearOption, mainHand *proto.ItemSpec) []*proto.ItemSpec {
		if o.once && mainHand != nil && mainHand.Id == o.spec.Id {
			return nil
		}
		candidate := append([]*proto.ItemSpec(nil), best...)
		candidate[mh] = mainHand
		candidate[oh] = s.withEnchantOf(oh, o.spec, best[oh])
		return candidate
	}
	var oneHanders, twoHanders []gearOption
	for _, o := range pool[mh] {
		if sameGear(best[mh], o) {
			continue
		}
		if o.twoHand {
			twoHanders = append(twoHanders, o)
		} else {
			oneHanders = append(oneHanders, o)
		}
	}
	byDPS := func(entries []bisEntry) []bisEntry {
		sort.SliceStable(entries, func(i, j int) bool { return entries[i].DPS > entries[j].DPS })
		return entries
	}
	mainHandWith := func(offHand *proto.ItemSpec) func(rankedGear) *proto.ItemSpec {
		return func(r rankedGear) *proto.ItemSpec {
			if r.option.twoHand {
				return nil
			}
			return offHand
		}
	}

	if !isTwoHand(pool, best[mh]) {
		// A one hand pick: every other main hand next to the off hand the set holds, and
		// every other off hand next to the set's main hand.
		picks := pickOf(mh, best[oh])
		mainOptions := append(oneHanders, twoHanders...)
		mainPlacements := func(o gearOption) [][]*proto.ItemSpec {
			return [][]*proto.ItemSpec{withMainHand(o, best[oh])}
		}
		source("mainhand", mainOptions, mainPlacements, mainHandWith(best[oh]))
		ranked := s.rankGear("mainhand", mainOptions, mainPlacements, nil, keep-len(picks), knobs)
		ahead("mainhand", ranked)
		mainList := append(picks, byDPS(entriesOf(ranked, mainHandWith(best[oh])))...)
		offPicks := pickOf(oh, best[mh])
		offOptions := slices.DeleteFunc(slices.Clone(pool[oh]), func(o gearOption) bool { return sameGear(best[oh], o) })
		offPlacements := func(o gearOption) [][]*proto.ItemSpec {
			return [][]*proto.ItemSpec{withOffHand(o, best[mh])}
		}
		source("offhand", offOptions, offPlacements, func(rankedGear) *proto.ItemSpec { return best[mh] })
		offRanked := s.rankGear("offhand", offOptions, offPlacements, nil, keep-len(offPicks), knobs)
		ahead("offhand", offRanked)
		return mainList, append(offPicks, entriesOf(offRanked, func(rankedGear) *proto.ItemSpec { return best[mh] })...)
	}

	// A two hand pick. The best one hander with nothing in the off hand, then the best off
	// hand next to it, then every one hander next to that off hand.
	picks := pickOf(mh, nil)
	twoPlacements := func(o gearOption) [][]*proto.ItemSpec {
		return [][]*proto.ItemSpec{withMainHand(o, nil)}
	}
	source("mainhand", twoHanders, twoPlacements, noOtherHand)
	twoRanked := s.rankGear("mainhand two handers", twoHanders, twoPlacements, nil, keep-len(picks), knobs)
	ahead("mainhand", twoRanked)
	mainList := append(picks, entriesOf(twoRanked, noOtherHand)...)
	if *gearRankTwoHandOnly {
		return mainList, nil
	}
	bare := s.rankGear("", oneHanders, func(o gearOption) [][]*proto.ItemSpec {
		return [][]*proto.ItemSpec{withMainHand(o, nil)}
	}, nil, 1, knobs)
	if len(bare) == 0 {
		return mainList, nil
	}
	oneHand := s.withEnchantOf(mh, bare[0].option.spec, best[mh])
	offPlacements := func(o gearOption) [][]*proto.ItemSpec {
		return [][]*proto.ItemSpec{withOffHand(o, oneHand)}
	}
	source("offhand", pool[oh], offPlacements, func(rankedGear) *proto.ItemSpec { return oneHand })
	offRanked := s.rankGear("offhand", pool[oh], offPlacements, nil, keep, knobs)
	offList := entriesOf(offRanked, func(rankedGear) *proto.ItemSpec { return oneHand })
	if len(offRanked) == 0 {
		return mainList, offList
	}
	offHand := s.withEnchantOf(oh, offRanked[0].option.spec, best[oh])
	onePlacements := func(o gearOption) [][]*proto.ItemSpec {
		return [][]*proto.ItemSpec{withMainHand(o, offHand)}
	}
	source("mainhand", oneHanders, onePlacements, mainHandWith(offHand))
	oneRanked := s.rankGear("mainhand one handers", oneHanders, onePlacements, nil, keep-len(picks), knobs)
	// A one hand setup that beats the two hander pick goes in the set like any other item
	// that beats its pick. The search tries one handers next to the off hand it holds, and a
	// two hander leaves none, so it may never have tried this setup.
	ahead("mainhand", oneRanked)
	others := byDPS(append(mainList[len(picks):], entriesOf(oneRanked, mainHandWith(offHand))...))
	mainList = append(picks, others...)
	if len(mainList) > keep {
		mainList = mainList[:keep]
	}
	return mainList, offList
}

// trimBis keeps the first keep entries of a slot's list, best first.
//
// A random suffix item comes in several rolls of the same name ("of the Tiger" with more or
// less of each stat), and they'd crowd the list with one item. So only the best roll of
// each name stays, and a weaker roll of a pick's name goes too.
func trimBis(entries []bisEntry, keep int) []bisEntry {
	type named struct {
		id   int32
		name string
	}
	seen := map[named]bool{}
	var out []bisEntry
	for _, e := range entries {
		if len(out) == keep {
			break
		}
		if seen[named{e.ID, e.Name}] {
			continue
		}
		seen[named{e.ID, e.Name}] = true
		out = append(out, e)
	}
	return out
}

// For the entries that name no item in the other hand.
func noOtherHand(rankedGear) *proto.ItemSpec { return nil }

// useDurOf is the list's use_dur, the fight length its uses count on (-use-dur), or nil when
// they count on the list's own fight.
func useDurOf(s *searcher) any {
	if !s.useScores {
		return nil
	}
	return *useDur
}

// writeBis writes the list file. The format is the contract BayovCore's bis.py reads, see
// docs/bis_lists.md. Each entry goes on a line of its own, so a diff of two runs reads
// item by item.
//
// from, when set (-gear-rank-slots), is the list whose other slots we copy, with its top-level
// dps.
func (s *searcher) writeBis(path string, base result, slots []bisSlot, enchants []bisEnchantSlot, from *rankFrom) error {
	player := s.setup.player()
	talents := player.TalentsString
	if s.setup.talents != "" {
		talents = s.setup.talents
	}
	faction := "Alliance"
	if contains([]proto.Race{proto.Race_RaceOrc, proto.Race_RaceTroll, proto.Race_RaceTauren, proto.Race_RaceUndead}, player.Race) {
		faction = "Horde"
	}
	header := []struct {
		key   string
		value any
	}{
		{"name", *gearRankName},
		{"spec", *gearRankSpec},
		{"class", strings.ToUpper(strings.TrimPrefix(player.Class.String(), "Class"))},
		{"race", strings.TrimPrefix(player.Race.String(), "Race")},
		{"faction", faction},
		{"level", player.Level},
		{"talents", talents},
		{"durations", s.durations},
		{"use_dur", useDurOf(s)},
		{"dps", listDPS(base, from)},
		{"made", time.Now().Format("2006-01-02")},
		{"sim", *gearRankSim},
	}
	var b strings.Builder
	b.WriteString("{\n")
	for _, h := range header {
		if h.value == nil {
			continue
		}
		data, err := json.Marshal(h.value)
		if err != nil {
			return err
		}
		fmt.Fprintf(&b, "  %q: %s,\n", h.key, data)
	}
	var slotLists, enchantLists []rawList
	for _, slot := range slots {
		list, err := rawListOf(slot.name, slot.entries)
		if err != nil {
			return err
		}
		slotLists = append(slotLists, list)
	}
	for _, slot := range enchants {
		list, err := rawListOf(slot.name, slot.entries)
		if err != nil {
			return err
		}
		enchantLists = append(enchantLists, list)
	}
	if from != nil {
		slotLists = mergeLists(bisListNames, slotLists, from.Slots)
		enchantLists = mergeLists(gearSlotNames, enchantLists, from.Enchants)
	}
	b.WriteString("  \"slots\": {\n")
	writeLists(&b, slotLists)
	b.WriteString("  }")
	if len(enchantLists) > 0 {
		b.WriteString(",\n  \"enchants\": {\n")
		writeLists(&b, enchantLists)
		b.WriteString("  }")
	}
	b.WriteString("\n}\n")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return err
	}
	fmt.Printf("  wrote %s\n", path)
	return nil
}
