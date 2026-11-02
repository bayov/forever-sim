//go:build practice

package main

import (
	"bufio"
	_ "embed"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Reading the WoW Forever client's files: BayovCore records, key bindings and CVars. We only ever
// read the WoW folder.

//go:embed defaults/bindings.txt
var defaultBindings string

type wowClient struct {
	// dataRoot is the WoW install (with .build.info and Data), forever the Forever folder in it.
	dataRoot string
	forever  string
}

// findClient finds the Forever folder in the data root by its .flavor.info, which names wowf.
func findClient(dataRoot string) (*wowClient, error) {
	entries, err := os.ReadDir(dataRoot)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), "_") {
			continue
		}
		flavor, err := os.ReadFile(filepath.Join(dataRoot, e.Name(), ".flavor.info"))
		if err == nil && strings.Contains(string(flavor), "wowf") {
			return &wowClient{dataRoot: dataRoot, forever: filepath.Join(dataRoot, e.Name())}, nil
		}
	}
	if _, err := os.Stat(filepath.Join(dataRoot, "_forever_")); err == nil {
		return &wowClient{dataRoot: dataRoot, forever: filepath.Join(dataRoot, "_forever_")}, nil
	}
	return nil, errors.New("no Forever folder (with a wowf .flavor.info) in " + dataRoot)
}

func (c *wowClient) accounts() []string {
	entries, _ := os.ReadDir(filepath.Join(c.forever, "WTF", "Account"))
	var out []string
	for _, e := range entries {
		if e.IsDir() && e.Name() != "SavedVariables" {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out
}

// record is one character's BayovCore CharSetup record.
type record struct {
	GUID    string                    `json:"guid"`
	Name    string                    `json:"name"`
	Realm   string                    `json:"realm"`
	Class   string                    `json:"class"`
	Race    string                    `json:"race"`
	Level   int                       `json:"level"`
	Seen    float64                   `json:"seen"`
	Account string                    `json:"account"`
	Slots   map[string]map[string]any `json:"slots,omitempty"`
	Macros  []map[string]any          `json:"macros,omitempty"`
	// AccountMacros are the account's own macros, the fallback for a macro slot.
	AccountMacros []map[string]any `json:"accountMacros,omitempty"`
}

// records reads every account's BayovCore file and keeps the newest copy of each GUID.
func (c *wowClient) records() ([]*record, error) {
	byGUID := map[string]*record{}
	for _, account := range c.accounts() {
		path := filepath.Join(c.forever, "WTF", "Account", account, "SavedVariables", "BayovCore.lua")
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		vars, err := parseSavedVariables(string(data))
		if err != nil {
			return nil, err
		}
		setup := luaTable(vars["BayovCoreDB"], "modules", "charsetup")
		accountMacros := luaList(luaTable(setup, "account")["macros"])
		for guid, raw := range luaTable(setup, "chars") {
			m, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			rec := &record{
				GUID:          guid,
				Name:          luaString(m, "name"),
				Realm:         luaString(m, "realm"),
				Class:         luaString(m, "class"),
				Race:          luaString(m, "race"),
				Level:         int(luaNumber(m, "level")),
				Seen:          luaNumber(m, "seen"),
				Account:       account,
				Slots:         map[string]map[string]any{},
				Macros:        luaList(m["macros"]),
				AccountMacros: accountMacros,
			}
			for slot, v := range luaTable(m, "slots") {
				if s, ok := v.(map[string]any); ok {
					rec.Slots[slot] = s
				}
			}
			if old := byGUID[guid]; old == nil || rec.Seen > old.Seen {
				byGUID[guid] = rec
			}
		}
	}
	var out []*record
	for _, rec := range byGUID {
		out = append(out, rec)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Seen > out[j].Seen })
	return out, nil
}

func luaList(v any) []map[string]any {
	m, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	var out []map[string]any
	for i := 1; ; i++ {
		item, ok := m[strconv.Itoa(i)].(map[string]any)
		if !ok {
			return out
		}
		out = append(out, item)
	}
}

// parseBindings reads `bind KEY COMMAND` lines into binds, in order. NONE unbinds the key.
func parseBindings(text string, binds map[string]string) {
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "bind ") {
			continue
		}
		rest := strings.TrimPrefix(line, "bind ")
		key, command, ok := strings.Cut(rest, " ")
		if !ok {
			continue
		}
		command = strings.TrimSpace(command)
		if command == "NONE" || command == "" {
			delete(binds, key)
			continue
		}
		binds[key] = command
	}
}

// parseCVars reads `SET name "value"` lines into cvars.
func parseCVars(text string, cvars map[string]string) {
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "SET ") {
			continue
		}
		name, value, ok := strings.Cut(strings.TrimPrefix(line, "SET "), " ")
		if !ok {
			continue
		}
		cvars[name] = strings.Trim(value, `"`)
	}
}

// characterDir is the character's own WTF folder, joined by name, empty when there is none.
func (c *wowClient) characterDir(account, name string) string {
	if name == "" {
		return ""
	}
	matches, _ := filepath.Glob(filepath.Join(c.forever, "WTF", "Account", account, "*", name))
	if len(matches) > 0 {
		return matches[0]
	}
	return ""
}

// bindings are Blizzard's defaults, then the account's file, then the character's own.
func (c *wowClient) bindings(account, name string) map[string]string {
	binds := map[string]string{}
	parseBindings(defaultBindings, binds)
	if account == "" {
		return binds
	}
	if data, err := os.ReadFile(filepath.Join(c.forever, "WTF", "Account", account, "bindings-cache.wtf")); err == nil {
		parseBindings(string(data), binds)
	}
	if dir := c.characterDir(account, name); dir != "" {
		if data, err := os.ReadFile(filepath.Join(dir, "bindings-cache.wtf")); err == nil {
			parseBindings(string(data), binds)
		}
	}
	return binds
}

// cvars are Config.wtf, then the account's, then the character's. A missing CVar keeps its
// default, which the page knows.
func (c *wowClient) cvars(account, name string) map[string]string {
	cvars := map[string]string{}
	if data, err := os.ReadFile(filepath.Join(c.forever, "WTF", "Config.wtf")); err == nil {
		parseCVars(string(data), cvars)
	}
	if account == "" {
		return cvars
	}
	if data, err := os.ReadFile(filepath.Join(c.forever, "WTF", "Account", account, "config-cache.wtf")); err == nil {
		parseCVars(string(data), cvars)
	}
	if dir := c.characterDir(account, name); dir != "" {
		if data, err := os.ReadFile(filepath.Join(dir, "config-cache.wtf")); err == nil {
			parseCVars(string(data), cvars)
		}
	}
	return cvars
}
