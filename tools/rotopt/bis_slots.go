package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
)

// Ranking only some slots (-gear-rank-slots). After a change that touches one slot's items (a
// trinket's effect fixed in the sim, say), we don't need the whole ranking again. With
// -gear-rank-only the list's picks are the set, we rank only the named lists and their enchants,
// and we copy every other list and enchant list from the -gear-rank-from list as it is.
//
// The copied lists keep their numbers, so the list's top-level dps stays the old one too. The
// ranked lists' losses are against the set's DPS as we measure it now, which can differ from the
// old dps by the noise or by the change itself. We print both.

// The lists of a BiS list file, in its order.
var bisListNames = []string{"head", "neck", "shoulder", "back", "chest", "wrist", "hands", "waist", "legs", "feet", "finger", "trinket", "mainhand", "offhand", "ranged"}

// The -gear-rank-from list, as much of it as we copy.
type rankFrom struct {
	DPS      float64                      `json:"dps"`
	Slots    map[string][]json.RawMessage `json:"slots"`
	Enchants map[string][]json.RawMessage `json:"enchants"`
}

// loadRankFrom checks the -gear-rank-slots flags and reads the -gear-rank-from list, or returns
// nil when we rank every slot.
func loadRankFrom() (*rankFrom, error) {
	if *gearRankSlots == "" {
		return nil, nil
	}
	if !*gearRankOnly || *gearRankFrom == "" {
		return nil, fmt.Errorf("-gear-rank-slots needs -gear-rank-only and -gear-rank-from, so the other slots' lists fit the set")
	}
	for _, name := range strings.Split(*gearRankSlots, ",") {
		if !slices.Contains(bisListNames, name) {
			return nil, fmt.Errorf("-gear-rank-slots: no list %q, the lists are %s", name, strings.Join(bisListNames, ", "))
		}
	}
	data, err := os.ReadFile(*gearRankFrom)
	if err != nil {
		return nil, err
	}
	from := &rankFrom{}
	if err := json.Unmarshal(data, from); err != nil {
		return nil, fmt.Errorf("%s: %w", *gearRankFrom, err)
	}
	return from, nil
}

// rankListSelected says whether we rank the list (a slot list, or an enchant list by its gear
// slot name, like finger1). Without -gear-rank-slots we rank them all.
func rankListSelected(name string) bool {
	if *gearRankSlots == "" {
		return true
	}
	return slices.Contains(strings.Split(*gearRankSlots, ","), listOfName(name))
}

// listOfName is the slot list a gear slot belongs to: finger1 and finger2 share the finger list.
func listOfName(name string) string {
	switch name {
	case "finger1", "finger2":
		return "finger"
	case "trinket1", "trinket2":
		return "trinket"
	}
	return name
}

// listDPS is the list's top-level dps: the set's as we measured it, or the -gear-rank-from list's
// when we copy most of it.
func listDPS(base result, from *rankFrom) float64 {
	if from == nil {
		return round2(base.dps)
	}
	fmt.Printf("  the set scores %.2f now, the list keeps its dps of %.2f (%+.2f)\n", base.dps, from.DPS, base.dps-from.DPS)
	return from.DPS
}

// A list as the file has it: its name and one JSON line for each entry.
type rawList struct {
	name    string
	entries []json.RawMessage
}

func rawListOf[T any](name string, entries []T) (rawList, error) {
	list := rawList{name: name}
	for _, e := range entries {
		data, err := json.Marshal(e)
		if err != nil {
			return list, err
		}
		list.entries = append(list.entries, data)
	}
	return list, nil
}

// mergeLists is the lists in order: a list we ranked from fresh, and any other from old. A list
// we ranked that came out empty stays out, even when old has it.
func mergeLists(order []string, fresh []rawList, old map[string][]json.RawMessage) []rawList {
	var out []rawList
	for _, name := range order {
		if rankListSelected(name) {
			if i := slices.IndexFunc(fresh, func(l rawList) bool { return l.name == name }); i >= 0 {
				out = append(out, fresh[i])
			}
			continue
		}
		if entries, ok := old[name]; ok {
			list := rawList{name: name}
			for _, e := range entries {
				var compact bytes.Buffer
				if json.Compact(&compact, e) == nil {
					e = compact.Bytes()
				}
				list.entries = append(list.entries, e)
			}
			out = append(out, list)
		}
	}
	return out
}

// writeLists writes the lists of a "slots" or "enchants" object, each entry on a line of its own.
func writeLists(b *strings.Builder, lists []rawList) {
	for i, list := range lists {
		fmt.Fprintf(b, "    %q: [", list.name)
		for j, e := range list.entries {
			if j > 0 {
				b.WriteString(",")
			}
			fmt.Fprintf(b, "\n      %s", e)
		}
		if len(list.entries) > 0 {
			b.WriteString("\n    ")
		}
		b.WriteString("]")
		if i < len(lists)-1 {
			b.WriteString(",")
		}
		b.WriteString("\n")
	}
}
