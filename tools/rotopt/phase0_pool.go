package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// The gear search's dry run (-gear-pool-report, bis_phase0.sh DRY_RUN=1): what a spec's pool
// holds and which of its items the sim may get wrong, without simming anything. It is the
// input of the hand check that comes before a spec's first Phase 0 search.
//
// An item's effect is unmodelled when its Classic tooltip (wowhead_item_tooltips.csv) has a
// "Use:" or "Chance on hit:" line and the sim registers no item or weapon effect for it
// (core.HasItemEffect, core.HasWeaponEffect). Forever's own items have no tooltip there, so
// they are counted apart. An item's stats are a guess when Forever still hides them
// (UIItem.HiddenStats): Classic's stats, or only part of Forever's.

var (
	gearPoolReport = flag.String("gear-pool-report", "", "gear search dry run: write the pool's size, the items with unmodelled use or proc effects, the items with hidden stats and the items -gear-content doesn't list to this JSON file, then stop")
	gearPoolSpec   = flag.String("gear-pool-spec", "", "the spec -gear-pool-report names, the key of tools/rotopt/phase0_specs.json")
	gearContent    = flag.String("gear-content", "", "the hand check of what is open (tools/rotopt/phase0_content.json): its faction tags for the gear search's faction check, and -gear-pool-report's unlisted items")
	itemTooltips   = flag.String("item-tooltips", "assets/db_inputs/wowhead_item_tooltips.csv", "wowhead's Classic item tooltips, id and JSON per row, for -gear-pool-report")
)

type poolItem struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	Slot string `json:"slot"`
	// What the tooltip says the sim doesn't model ("Use", "Chance on hit"), or the kind of
	// hidden stats ("classic", "partial").
	Kind string `json:"kind,omitempty"`
}

type poolReport struct {
	Spec    string         `json:"spec"`
	Options map[string]int `json:"options"`
	// Items counts the distinct items over every slot, a random suffix item once.
	Items       int        `json:"items"`
	Unmodelled  []poolItem `json:"unmodelled"`
	HiddenStats []poolItem `json:"hiddenStats"`
	NoTooltip   int        `json:"foreverNoTooltip"`
	// Unlisted counts the items -gear-content doesn't list. Most of a pool is, so the run's
	// check of the picks and top 5s (bis_phase0.sh) is what names them.
	Unlisted int `json:"unlisted"`
	// Faction is the build's, and FactionDropped what the faction check left out of the pool:
	// the items only the other faction can get, with what says so (faction.go).
	Faction        string     `json:"faction"`
	FactionDropped []poolItem `json:"factionDropped"`
}

// readTooltips maps an item ID to its tooltip HTML. Each line is the ID, a comma and the
// JSON wowhead served, like orchestration/tools/bis_data.py reads it.
func readTooltips(path string) (map[int32]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[int32]string{}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1<<20), 1<<24)
	for scanner.Scan() {
		key, rest, ok := strings.Cut(scanner.Text(), ",")
		if !ok {
			continue
		}
		id, err := strconv.Atoi(key)
		if err != nil {
			continue
		}
		var entry struct {
			Tooltip string `json:"tooltip"`
		}
		if json.Unmarshal([]byte(rest), &entry) == nil {
			out[int32(id)] = entry.Tooltip
		}
	}
	return out, scanner.Err()
}

// unmodelledEffect is what an item's tooltip has that the sim doesn't model, or "".
func unmodelledEffect(id int32, tooltip string) string {
	if core.HasItemEffect(id) || core.HasWeaponEffect(id) {
		return ""
	}
	var kinds []string
	if strings.Contains(tooltip, ">Use: ") || strings.Contains(tooltip, "\nUse: ") {
		kinds = append(kinds, "Use")
	}
	if strings.Contains(tooltip, "Chance on hit:") {
		kinds = append(kinds, "Chance on hit")
	}
	return strings.Join(kinds, ", ")
}

// writePoolReport writes the dry run's report of the pool, for the spec named spec.
func writePoolReport(path, spec string, pool [][]gearOption, factions *factionCheck) error {
	tooltips, err := readTooltips(*itemTooltips)
	if err != nil {
		return err
	}
	content := map[string]json.RawMessage{}
	if *gearContent != "" {
		data, err := os.ReadFile(*gearContent)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, &content); err != nil {
			return err
		}
	}
	report := poolReport{Spec: spec, Options: map[string]int{}}
	seen := map[int32]bool{}
	for slot, options := range pool {
		report.Options[gearSlotNames[slot]] = len(options)
		for _, o := range options {
			id := o.spec.Id
			if seen[id] {
				continue
			}
			seen[id] = true
			item := itemsByID[id]
			name := item.GetName()
			entry := poolItem{ID: id, Name: name, Slot: gearSlotNames[slot]}
			tooltip, ok := tooltips[id]
			if !ok && id >= foreverFirstItemID {
				report.NoTooltip++
			}
			if kind := unmodelledEffect(id, tooltip); kind != "" {
				e := entry
				e.Kind = kind
				report.Unmodelled = append(report.Unmodelled, e)
			}
			switch item.GetHiddenStats() {
			case proto.UIItem_HIDDEN_STATS_CLASSIC, proto.UIItem_HIDDEN_STATS_PARTIAL:
				e := entry
				e.Kind = strings.ToLower(strings.TrimPrefix(item.GetHiddenStats().String(), "HIDDEN_STATS_"))
				report.HiddenStats = append(report.HiddenStats, e)
			}
			if _, ok := content[strconv.Itoa(int(id))]; *gearContent != "" && !ok {
				report.Unlisted++
			}
		}
	}
	report.Items = len(seen)
	report.FactionDropped = []poolItem{}
	if factions != nil {
		report.Faction = factions.faction.String()
		for _, e := range factions.dropped {
			report.FactionDropped = append(report.FactionDropped, e)
		}
	}
	for _, list := range [][]poolItem{report.Unmodelled, report.HiddenStats, report.FactionDropped} {
		sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
