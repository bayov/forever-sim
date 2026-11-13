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
}

type bisSlot struct {
	name    string
	entries []bisEntry
}

// An option and its score in the best place we found for it.
type rankedGear struct {
	option gearOption
	result result
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
	// owner[i] is the option pending[i] belongs to, and -1 for an empty slot.
	var owner []int
	for i, o := range options {
		for _, gear := range placements(o) {
			if gear != nil {
				s.setGear(gear)
				pending = append(pending, s.prepare(knobs))
				owner = append(owner, i)
			}
		}
	}
	for _, gear := range empty {
		s.setGear(gear)
		pending = append(pending, s.prepare(knobs))
		owner = append(owner, -1)
	}
	best := make([]*result, len(options))
	var floor *result
	for i, r := range s.scoreAll(pending) {
		at := &floor
		if owner[i] >= 0 {
			at = &best[owner[i]]
		}
		if *at == nil || r.dps > (*at).dps {
			*at = &r
		}
	}
	var ranked []rankedGear
	for i, r := range best {
		if r != nil {
			ranked = append(ranked, rankedGear{options[i], *r})
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
func (s *searcher) rankBis(path string, pool [][]gearOption, best []*proto.ItemSpec, names []string, unsourced []bool, knobs Knobs) error {
	keep := *gearRankKeep
	// With -gear-rank-enchants the set gets its best enchants first, so the items are
	// ranked on them (bis_enchants.go).
	var enchants []bisEnchantSlot
	if *gearRankEnchants != "" {
		fmt.Println("ranking every enchant of the set:")
		var err error
		if enchants, err = s.rankEnchants(best, knobs); err != nil {
			return err
		}
	}
	s.setGear(best)
	base := s.score(knobs)
	fmt.Printf("ranking every slot against the set: %s\n", base)

	round2 := func(x float64) float64 {
		if r := math.Round(x*100) / 100; r != 0 {
			return r
		}
		return 0
	}
	entryOf := func(o gearOption, r result, with *proto.ItemSpec) bisEntry {
		e := bisEntry{ID: o.spec.Id, Suffix: o.spec.RandomSuffix, Name: o.name, DPS: round2(r.dps), Loss: round2(base.dps - r.dps), Unsourced: o.unsourced}
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
		e := bisEntry{ID: item.Id, Suffix: item.RandomSuffix, Name: name, DPS: round2(base.dps), Unsourced: unsourced[slot], BiS: true, Enchant: item.Enchant}
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
	var slots []bisSlot
	single := func(slot int) {
		picks := pickOf(slot, nil)
		ranked := s.rankGear(gearSlotNames[slot], others(slot, best[slot]), func(o gearOption) [][]*proto.ItemSpec {
			return [][]*proto.ItemSpec{swapped(slot, o.spec)}
		}, emptied(slot), wide, knobs)
		slots = append(slots, bisSlot{gearSlotNames[slot], trimBis(append(picks, entriesOf(ranked, noOtherHand)...), keep)})
	}
	pair := func(name string, a, b int) {
		picks := append(pickOf(a, nil), pickOf(b, nil)...)
		// Each slot takes any item but the one the other slot holds.
		ranked := s.rankGear(name, others(a, best[a], best[b]), func(o gearOption) [][]*proto.ItemSpec {
			var out [][]*proto.ItemSpec
			if best[b] == nil || best[b].Id != o.spec.Id {
				out = append(out, swapped(a, o.spec))
			}
			if best[a] == nil || best[a].Id != o.spec.Id {
				out = append(out, swapped(b, o.spec))
			}
			return out
		}, emptied(a, b), wide, knobs)
		slots = append(slots, bisSlot{name, trimBis(append(picks, entriesOf(ranked, noOtherHand)...), keep)})
	}

	for slot := range 10 {
		single(slot)
	}
	pair("finger", int(proto.ItemSlot_ItemSlotFinger1), int(proto.ItemSlot_ItemSlotFinger2))
	pair("trinket", int(proto.ItemSlot_ItemSlotTrinket1), int(proto.ItemSlot_ItemSlotTrinket2))
	main, off := s.rankWeapons(pool, best, wide, knobs, pickOf, entriesOf, base)
	slots = append(slots, bisSlot{"mainhand", trimBis(main, keep)}, bisSlot{"offhand", trimBis(off, keep)})
	single(int(proto.ItemSlot_ItemSlotRanged))

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
			fmt.Println()
		}
	}
	return s.writeBis(path, base, slots, enchants)
}

// rankWeapons is the main hand and off hand lists. See the top of the file for how a one
// hander is scored next to a two hander pick.
func (s *searcher) rankWeapons(pool [][]gearOption, best []*proto.ItemSpec, keep int, knobs Knobs,
	pickOf func(int, *proto.ItemSpec) []bisEntry, entriesOf func([]rankedGear, func(rankedGear) *proto.ItemSpec) []bisEntry, base result) ([]bisEntry, []bisEntry) {
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
		ranked := s.rankGear("mainhand", append(oneHanders, twoHanders...), func(o gearOption) [][]*proto.ItemSpec {
			return [][]*proto.ItemSpec{withMainHand(o, best[oh])}
		}, nil, keep-len(picks), knobs)
		mainList := append(picks, byDPS(entriesOf(ranked, mainHandWith(best[oh])))...)
		offPicks := pickOf(oh, best[mh])
		offOptions := slices.DeleteFunc(slices.Clone(pool[oh]), func(o gearOption) bool { return sameGear(best[oh], o) })
		offRanked := s.rankGear("offhand", offOptions, func(o gearOption) [][]*proto.ItemSpec {
			return [][]*proto.ItemSpec{withOffHand(o, best[mh])}
		}, nil, keep-len(offPicks), knobs)
		return mainList, append(offPicks, entriesOf(offRanked, func(rankedGear) *proto.ItemSpec { return best[mh] })...)
	}

	// A two hand pick. The best one hander with nothing in the off hand, then the best off
	// hand next to it, then every one hander next to that off hand.
	picks := pickOf(mh, nil)
	twoRanked := s.rankGear("mainhand two handers", twoHanders, func(o gearOption) [][]*proto.ItemSpec {
		return [][]*proto.ItemSpec{withMainHand(o, nil)}
	}, nil, keep-len(picks), knobs)
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
	offRanked := s.rankGear("offhand", pool[oh], func(o gearOption) [][]*proto.ItemSpec {
		return [][]*proto.ItemSpec{withOffHand(o, oneHand)}
	}, nil, keep, knobs)
	offList := entriesOf(offRanked, func(rankedGear) *proto.ItemSpec { return oneHand })
	if len(offRanked) == 0 {
		return mainList, offList
	}
	offHand := s.withEnchantOf(oh, offRanked[0].option.spec, best[oh])
	oneRanked := s.rankGear("mainhand one handers", oneHanders, func(o gearOption) [][]*proto.ItemSpec {
		return [][]*proto.ItemSpec{withMainHand(o, offHand)}
	}, nil, keep-len(picks), knobs)
	if len(oneRanked) > 0 {
		r := oneRanked[0].result
		if gain := r.dps - base.dps; gain > s.confidence*gainNoise(base, r) {
			fmt.Printf("  %s with %s beats the two hander pick by %.1f, the search never tried a one hand setup from here\n",
				oneRanked[0].option.name, offRanked[0].option.name, gain)
		}
	}
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

// writeBis writes the list file. The format is the contract BayovCore's bis.py reads, see
// docs/bis_lists.md. Each entry goes on a line of its own, so a diff of two runs reads
// item by item.
func (s *searcher) writeBis(path string, base result, slots []bisSlot, enchants []bisEnchantSlot) error {
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
		{"dps", math.Round(base.dps*100) / 100},
		{"made", time.Now().Format("2006-01-02")},
		{"sim", *gearRankSim},
	}
	var b strings.Builder
	b.WriteString("{\n")
	for _, h := range header {
		data, err := json.Marshal(h.value)
		if err != nil {
			return err
		}
		fmt.Fprintf(&b, "  %q: %s,\n", h.key, data)
	}
	b.WriteString("  \"slots\": {\n")
	for i, slot := range slots {
		fmt.Fprintf(&b, "    %q: [", slot.name)
		for j, e := range slot.entries {
			data, err := json.Marshal(e)
			if err != nil {
				return err
			}
			if j > 0 {
				b.WriteString(",")
			}
			fmt.Fprintf(&b, "\n      %s", data)
		}
		if len(slot.entries) > 0 {
			b.WriteString("\n    ")
		}
		b.WriteString("]")
		if i < len(slots)-1 {
			b.WriteString(",")
		}
		b.WriteString("\n")
	}
	b.WriteString("  }")
	if len(enchants) > 0 {
		b.WriteString(",\n  \"enchants\": {\n")
		for i, slot := range enchants {
			fmt.Fprintf(&b, "    %q: [", slot.name)
			for j, e := range slot.entries {
				data, err := json.Marshal(e)
				if err != nil {
					return err
				}
				if j > 0 {
					b.WriteString(",")
				}
				fmt.Fprintf(&b, "\n      %s", data)
			}
			if len(slot.entries) > 0 {
				b.WriteString("\n    ")
			}
			b.WriteString("]")
			if i < len(enchants)-1 {
				b.WriteString(",")
			}
			b.WriteString("\n")
		}
		b.WriteString("  }")
	}
	b.WriteString("\n}\n")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return err
	}
	fmt.Printf("  wrote %s\n", path)
	return nil
}
