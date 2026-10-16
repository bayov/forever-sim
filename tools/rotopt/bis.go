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

var (
	gearRankOut  = flag.String("gear-rank-out", "", "gear search: once the search settles, rank every slot's options against the set and write the BiS list (docs/bis_lists.md) to this file")
	gearRankKeep = flag.Int("gear-rank-keep", 10, "gear search: how many items each slot of the -gear-rank-out list keeps, the BiS picks included")
	gearRankName = flag.String("gear-rank-name", "", "the -gear-rank-out list's label, e.g. \"Level 30\"")
	gearRankSpec = flag.String("gear-rank-spec", "", "the -gear-rank-out list's spec, by its mksettings name (enh, ele, ...)")
	gearRankSim  = flag.String("gear-rank-sim", "", "the forever-sim commit the -gear-rank-out list was simmed on, for the list's sim field")
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

// rankGear scores each option and returns the best keep of them, best first.
//
// placements gives the gear sets an option can be scored in, and the option counts with
// the best of them. A ring is placed in either ring slot, say. An option with no
// placement is left out. We screen the options the same way screenGear does: everything
// at screenIterations first when the slot has more options than we keep, and only the
// best of those at full iterations.
func (s *searcher) rankGear(options []gearOption, placements func(gearOption) [][]*proto.ItemSpec, keep int, knobs Knobs) []rankedGear {
	if keep <= 0 {
		return nil
	}
	scoreAt := func(o gearOption, iterations int32) (result, bool) {
		full := s.iterations
		s.iterations = iterations
		defer func() { s.iterations = full }()
		var best result
		found := false
		for _, gear := range placements(o) {
			if gear == nil {
				continue
			}
			s.setGear(gear)
			if r := s.score(knobs); !found || r.dps > best.dps {
				best, found = r, true
			}
		}
		return best, found
	}
	screenKeep := max(s.screenKeep, keep)
	if s.screenIterations > 0 && len(options) > screenKeep {
		var screened []rankedGear
		for _, o := range options {
			if r, ok := scoreAt(o, s.screenIterations); ok {
				screened = append(screened, rankedGear{o, r})
			}
		}
		sort.SliceStable(screened, func(i, j int) bool { return screened[i].result.dps > screened[j].result.dps })
		options = nil
		for i := 0; i < len(screened) && i < screenKeep; i++ {
			options = append(options, screened[i].option)
		}
	}
	var ranked []rankedGear
	for _, o := range options {
		if r, ok := scoreAt(o, s.iterations); ok {
			ranked = append(ranked, rankedGear{o, r})
		}
	}
	sort.SliceStable(ranked, func(i, j int) bool { return ranked[i].result.dps > ranked[j].result.dps })
	if len(ranked) > keep {
		ranked = ranked[:keep]
	}
	return ranked
}

// Whether the option is the item and suffix of spec.
func sameGear(spec *proto.ItemSpec, o gearOption) bool {
	return spec != nil && spec.Id == o.spec.Id && spec.RandomSuffix == o.spec.RandomSuffix
}

// rankBis ranks every slot's options against the set best and writes the list to path.
// names and unsourced are the set's own, as the search printed them.
func (s *searcher) rankBis(path string, pool [][]gearOption, best []*proto.ItemSpec, names []string, unsourced []bool, knobs Knobs) error {
	keep := *gearRankKeep
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
		e := bisEntry{ID: item.Id, Suffix: item.RandomSuffix, Name: name, DPS: round2(base.dps), Unsourced: unsourced[slot], BiS: true}
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

	// We score more than we keep, because trimBis drops the weaker rolls of an item.
	wide := max(s.screenKeep, 2*keep)
	// An item that scores no better than leaving the slot empty does nothing in the sim
	// (a Hook of the Master Angler in a trinket slot), so it isn't worth listing. For the
	// rings and trinkets we empty whichever slot loses less.
	emptied := func(slots ...int) float64 {
		nothing := math.Inf(-1)
		for _, slot := range slots {
			candidate := append([]*proto.ItemSpec(nil), best...)
			candidate[slot] = nil
			s.setGear(candidate)
			nothing = max(nothing, s.score(knobs).dps)
		}
		return nothing
	}
	useful := func(ranked []rankedGear, nothing float64) []rankedGear {
		return slices.DeleteFunc(ranked, func(r rankedGear) bool { return r.result.dps <= nothing })
	}
	var slots []bisSlot
	single := func(slot int) {
		picks := pickOf(slot, nil)
		ranked := s.rankGear(others(slot, best[slot]), func(o gearOption) [][]*proto.ItemSpec {
			return [][]*proto.ItemSpec{swapped(slot, o.spec)}
		}, wide, knobs)
		ranked = useful(ranked, emptied(slot))
		slots = append(slots, bisSlot{gearSlotNames[slot], trimBis(append(picks, entriesOf(ranked, noOtherHand)...), keep)})
	}
	pair := func(name string, a, b int) {
		picks := append(pickOf(a, nil), pickOf(b, nil)...)
		// Each slot takes any item but the one the other slot holds.
		ranked := s.rankGear(others(a, best[a], best[b]), func(o gearOption) [][]*proto.ItemSpec {
			var out [][]*proto.ItemSpec
			if best[b] == nil || best[b].Id != o.spec.Id {
				out = append(out, swapped(a, o.spec))
			}
			if best[a] == nil || best[a].Id != o.spec.Id {
				out = append(out, swapped(b, o.spec))
			}
			return out
		}, wide, knobs)
		ranked = useful(ranked, emptied(a, b))
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
	return s.writeBis(path, base, slots)
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
		ranked := s.rankGear(append(oneHanders, twoHanders...), func(o gearOption) [][]*proto.ItemSpec {
			return [][]*proto.ItemSpec{withMainHand(o, best[oh])}
		}, keep-len(picks), knobs)
		mainList := append(picks, byDPS(entriesOf(ranked, mainHandWith(best[oh])))...)
		offPicks := pickOf(oh, best[mh])
		offOptions := slices.DeleteFunc(slices.Clone(pool[oh]), func(o gearOption) bool { return sameGear(best[oh], o) })
		offRanked := s.rankGear(offOptions, func(o gearOption) [][]*proto.ItemSpec {
			return [][]*proto.ItemSpec{withOffHand(o, best[mh])}
		}, keep-len(offPicks), knobs)
		return mainList, append(offPicks, entriesOf(offRanked, func(rankedGear) *proto.ItemSpec { return best[mh] })...)
	}

	// A two hand pick. The best one hander with nothing in the off hand, then the best off
	// hand next to it, then every one hander next to that off hand.
	picks := pickOf(mh, nil)
	twoRanked := s.rankGear(twoHanders, func(o gearOption) [][]*proto.ItemSpec {
		return [][]*proto.ItemSpec{withMainHand(o, nil)}
	}, keep-len(picks), knobs)
	mainList := append(picks, entriesOf(twoRanked, noOtherHand)...)
	bare := s.rankGear(oneHanders, func(o gearOption) [][]*proto.ItemSpec {
		return [][]*proto.ItemSpec{withMainHand(o, nil)}
	}, 1, knobs)
	if len(bare) == 0 {
		return mainList, nil
	}
	oneHand := s.withEnchantOf(mh, bare[0].option.spec, best[mh])
	offRanked := s.rankGear(pool[oh], func(o gearOption) [][]*proto.ItemSpec {
		return [][]*proto.ItemSpec{withOffHand(o, oneHand)}
	}, keep, knobs)
	offList := entriesOf(offRanked, func(rankedGear) *proto.ItemSpec { return oneHand })
	if len(offRanked) == 0 {
		return mainList, offList
	}
	offHand := s.withEnchantOf(oh, offRanked[0].option.spec, best[oh])
	oneRanked := s.rankGear(oneHanders, func(o gearOption) [][]*proto.ItemSpec {
		return [][]*proto.ItemSpec{withMainHand(o, offHand)}
	}, keep-len(picks), knobs)
	if len(oneRanked) > 0 {
		r := oneRanked[0].result
		if gain := r.dps - base.dps; gain > s.confidence*math.Hypot(r.stderr, base.stderr) {
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
func (s *searcher) writeBis(path string, base result, slots []bisSlot) error {
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
	b.WriteString("  }\n}\n")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return err
	}
	fmt.Printf("  wrote %s\n", path)
	return nil
}
