package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// historySettle is how long after a change BayovCore records it, at most. The addon waits for 2
// seconds of quiet (SETTLE in BayovCore's Modules/Export.lua), and we leave a few more for a slow frame.
const historySettle = 5 * time.Second

// historyRecord is one line of ~/forever-logs/history/<key>.jsonl, which
// tools/combat_log/history.py copies out of BayovCore's export module. The key is the
// character's GUID, or Name-Realm in the histories from before 2026-10-16 (characters.go says
// how we place those).
//
// A "login" or "mark" record holds the whole character, like an export. A "change" record
// holds only what changed: the gear slots (false for one that was emptied), the talents, the
// level, or the skills that went up. A "logout" record holds nothing. The addon writes one at
// logout and /reload, just before WoW saves the history, so its time says the history is
// complete up to then. We still rebuild an export for it, the same as the one before. The sheet ("stats") and the full talent list ("ref") are
// ids into <key>.blobs.json, because the addon stores each distinct one once.
type historyRecord struct {
	Kind    string           `json:"kind"`
	Date    string           `json:"date"`
	When    json.RawMessage  `json:"when"`
	Who     map[string]any   `json:"who"`
	Gear    historyGear      `json:"gear"`
	Skills  []map[string]any `json:"skills"`
	Imbues  json.RawMessage  `json:"imbues"`
	Talents map[string]any   `json:"talents"`
	Level   *int             `json:"level"`
	Stats   string           `json:"stats"`
	Buffs   json.RawMessage  `json:"buffs"`
}

// historyGear is a record's gear by slot number, as a string ("1" to "19").
//
// The addon saves gear as a Lua table, and its JSON encoder (isArray in BayovCore's
// Modules/Export.lua) writes a table as a list when its entry count equals #t. Most records
// come out as an object keyed by slot. But a table with holes can pass that test too: with the
// shoulders to the boots (slots 4 to 8) and the gloves, cloak and main hand (10, 15 and 16) on,
// there are 8 entries and #t can be 8. Then the list holds slots 1 to 8, with null for each
// empty one, and the slots past the hole are lost. Each null in the list stands for one slot
// past its end that the encoder dropped, so we keep that count in dropped.
type historyGear struct {
	slots map[string]json.RawMessage
	// listLen is the list's length when the record had a list, else 0.
	listLen int
	// dropped is how many filled slots past listLen the encoder left out.
	dropped int
}

func (g *historyGear) UnmarshalJSON(data []byte) error {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err == nil {
		g.slots = obj
		return nil
	}
	var list []json.RawMessage
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	g.slots = map[string]json.RawMessage{}
	g.listLen = len(list)
	for i, item := range list {
		if len(item) == 0 || string(item) == "null" {
			g.dropped++
			continue
		}
		g.slots[strconv.Itoa(i+1)] = item
	}
	return nil
}

// loadHistory rebuilds an export for each history record in dir. A character's records before
// its first login record can't be placed on a whole character, so we skip them.
func loadHistory(dir string) ([]*export, error) {
	paths, _ := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	var all []*export
	for _, path := range paths {
		exports, err := replayHistory(path)
		if err != nil {
			return nil, err
		}
		all = append(all, exports...)
	}
	return all, nil
}

func replayHistory(path string) ([]*export, error) {
	blobs := map[string]json.RawMessage{}
	if data, err := os.ReadFile(strings.TrimSuffix(path, ".jsonl") + ".blobs.json"); err == nil {
		if err := json.Unmarshal(data, &blobs); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// The character as of the last record, in the export's own shape.
	var state map[string]any
	gear := map[int]json.RawMessage{}
	var exports []*export
	logins := 0
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1<<20), 1<<24)
	for line := 1; scanner.Scan(); line++ {
		var rec historyRecord
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			return nil, fmt.Errorf("%s:%d: %w", path, line, err)
		}
		whole := rec.Kind == "login" || rec.Kind == "mark"
		if !whole && state == nil {
			continue
		}
		if rec.Kind == "login" {
			logins++
		}
		if whole {
			state = map[string]any{"when": rec.When, "who": rec.Who, "imbues": rec.Imbues, "skills": rec.Skills}
			before := gear
			gear = map[int]json.RawMessage{}
			// A login whose gear list lost slots past its end (see historyGear) still wears
			// what we last knew in those slots. On 2026-11-05 a login dropped the gloves, cloak
			// and staff, and the sim swung unarmed for half an hour.
			if rec.Gear.dropped > 0 {
				for slot, item := range before {
					if slot > rec.Gear.listLen {
						gear[slot] = item
					}
				}
			}
		}
		for slot, item := range rec.Gear.slots {
			n, err := strconv.Atoi(slot)
			if err != nil {
				return nil, fmt.Errorf("%s:%d: gear slot %q", path, line, slot)
			}
			if string(item) == "false" {
				delete(gear, n)
			} else {
				gear[n] = item
			}
		}
		if rec.Talents != nil {
			if ref, ok := rec.Talents["ref"].(string); ok {
				rec.Talents["talents"] = blobs[ref]
			}
			state["talents"] = rec.Talents
		}
		if rec.Level != nil {
			who := map[string]any{}
			for k, v := range state["who"].(map[string]any) {
				who[k] = v
			}
			who["level"] = *rec.Level
			state["who"] = who
		}
		if !whole && len(rec.Skills) > 0 {
			state["skills"] = mergeSkills(state["skills"].([]map[string]any), rec.Skills)
		}
		if rec.Stats != "" {
			state["stats"] = blobs[rec.Stats]
			// The buffs go with the sheet they were read with.
			state["buffs"] = rec.Buffs
		}

		// The export's time is its when.date, so each rebuilt export takes the record's.
		var when map[string]any
		if err := json.Unmarshal(state["when"].(json.RawMessage), &when); err != nil {
			return nil, fmt.Errorf("%s:%d: when: %w", path, line, err)
		}
		when["date"] = rec.Date
		// Slots 1 to 19 in order, with an empty entry for an empty slot, so two exports'
		// gear lines up slot by slot.
		list := make([]json.RawMessage, 19)
		for i := range list {
			list[i] = json.RawMessage("{}")
			if item, ok := gear[i+1]; ok {
				list[i] = item
			}
		}
		out := map[string]any{}
		for k, v := range state {
			out[k] = v
		}
		out["when"] = when
		out["gear"] = list
		data, err := json.Marshal(out)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %w", path, line, err)
		}
		ex, err := parseExport(data, fmt.Sprintf("%s:%d (history, %s)", filepath.Base(path), line, rec.Kind))
		if err != nil {
			return nil, err
		}
		ex.history = true
		ex.kind = rec.Kind
		ex.session = fmt.Sprintf("%s#%d", path, logins)
		exports = append(exports, ex)
	}
	return exports, scanner.Err()
}

// mergeSkills puts the skills that went up in place of their old lines, by skill ID or else
// by name, and adds a skill we didn't have.
func mergeSkills(skills, ups []map[string]any) []map[string]any {
	key := func(line map[string]any) string {
		if id, ok := line["skillID"]; ok {
			return fmt.Sprint(id)
		}
		return fmt.Sprint(line["name"])
	}
	merged := append([]map[string]any(nil), skills...)
	for _, up := range ups {
		found := false
		for i, line := range merged {
			if key(line) == key(up) {
				merged[i] = up
				found = true
			}
		}
		if !found {
			merged = append(merged, up)
		}
	}
	return merged
}
