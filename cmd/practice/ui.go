//go:build practice

package main

import (
	"bufio"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// The user's in-game UI: the EllesmereUI profile (where the frames and bars sit, their sizes and
// colors), its font, and the class keybind notes. The room draws its HUD from them so it looks
// like the game.

// ellesmereUI gives the parts of the account's active EllesmereUI profile the room draws: action
// bars, unit frames, resource bars, the damage meter, nameplates and the tooltip spot. It is nil
// when the account has no EllesmereUI.
func (c *wowClient) ellesmereUI(account string) map[string]any {
	if account == "" {
		return nil
	}
	data, err := os.ReadFile(filepath.Join(c.forever, "WTF", "Account", account, "SavedVariables", "EllesmereUI.lua"))
	if err != nil {
		return nil
	}
	vars, err := parseSavedVariables(string(data))
	if err != nil {
		return nil
	}
	db := luaTable(vars["EllesmereUIDB"])
	name := luaString(db, "activeProfile")
	if name == "" {
		name = "Default"
	}
	profile := luaTable(db, "profiles", name)
	if profile == nil {
		return nil
	}
	addons := luaTable(profile, "addons")
	out := map[string]any{
		"profile":         name,
		"uiScale":         db["ppUIScale"],
		"actionBars":      addons["EllesmereUIActionBars"],
		"unitFrames":      addons["EllesmereUIUnitFrames"],
		"resourceBars":    addons["EllesmereUIResourceBars"],
		"nameplates":      addons["EllesmereUINameplates"],
		"tooltipFixedPos": profile["tooltipFixedPos"],
		"customColors":    profile["customColors"],
		"fonts":           profile["fonts"],
	}
	if dm := luaTable(addons, "EllesmereUIDamageMeters", "dm"); dm != nil {
		out["damageMeter"] = map[string]any{
			"window":       luaTable(dm, "windows", "1"),
			"barHeight":    dm["barHeight"],
			"spellHistory": dm["spellHistory"],
		}
	}
	return out
}

// handleFont serves the HUD's font from the client: Avant Garde from SharedMedia_MyMedia when any
// of the install's game folders has it, else EllesmereUI's Avant Garde Naowh from the Forever
// folder.
func (s *server) handleFont(w http.ResponseWriter, r *http.Request) {
	if s.client == nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "max-age=86400")
	http.ServeFile(w, r, s.client.fontPath())
}

// fontPath is the font handleFont serves. We look in the Forever folder first, then in the other
// game folders, since SharedMedia_MyMedia may only be installed for Classic Era or Anniversary.
func (c *wowClient) fontPath() string {
	shared := filepath.Join("Interface", "AddOns", "SharedMedia_MyMedia", "font", "Avant Garde.ttf")
	dirs := []string{c.forever}
	if entries, err := os.ReadDir(c.dataRoot); err == nil {
		for _, e := range entries {
			if p := filepath.Join(c.dataRoot, e.Name()); e.IsDir() && strings.HasPrefix(e.Name(), "_") && p != c.forever {
				dirs = append(dirs, p)
			}
		}
	}
	for _, d := range dirs {
		if _, err := os.Stat(filepath.Join(d, shared)); err == nil {
			return filepath.Join(d, shared)
		}
	}
	return filepath.Join(c.forever, "Interface", "AddOns", "EllesmereUI", "media", "fonts", "Avant Garde Naowh.ttf")
}

// noteBinds reads the "## Keybinds" block of notes/<class>-keybinds-forever.md in the WoW folder:
// the keys the user plans for each spell, item and macro. The room puts a toolkit spell the
// character's bars don't have on its planned key's slot.
//
// The block is pairs of lines: keys, then what each one does, in the same columns.
//
//	Q                       SQ                       CQ
//	Earth Shock             Stormstrike              Earth Shock (r1)
func (c *wowClient) noteBinds(class string) map[string]string {
	if class == "" {
		return nil
	}
	data, err := os.ReadFile(filepath.Join(c.dataRoot, "notes", strings.ToLower(class)+"-keybinds-forever.md"))
	if err != nil {
		return nil
	}
	return parseNoteBinds(string(data))
}

func parseNoteBinds(text string) map[string]string {
	var lines []string
	inSection, inBlock := false, false
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "## "):
			inSection = strings.TrimSpace(line) == "## Keybinds"
		case inSection && strings.HasPrefix(line, "```"):
			inBlock = !inBlock
		case inSection && inBlock:
			lines = append(lines, line)
		}
	}
	out := map[string]string{}
	for i := 0; i+1 < len(lines); i++ {
		starts, keys := keyColumns(lines[i])
		if keys == nil {
			continue
		}
		labels := lines[i+1]
		for k, key := range keys {
			start := starts[k]
			if start >= len(labels) {
				continue
			}
			end := len(labels)
			if k+1 < len(starts) && starts[k+1] < end {
				end = starts[k+1]
			}
			if label := strings.TrimSpace(labels[start:end]); label != "" {
				out[key] = label
			}
		}
		i++
	}
	return out
}

var tokenPattern = regexp.MustCompile(`\S+`)

// keyColumns reads a line of key names like "Q   SQ   CQ" into WoW key names and the column
// each starts at. It returns nil when any word isn't a key, so label lines are skipped.
func keyColumns(line string) ([]int, []string) {
	var starts []int
	var keys []string
	for _, loc := range tokenPattern.FindAllStringIndex(line, -1) {
		key, ok := noteKey(line[loc[0]:loc[1]])
		if !ok {
			return nil, nil
		}
		starts = append(starts, loc[0])
		keys = append(keys, key)
	}
	return starts, keys
}

// noteKey turns the notes' short key names into WoW's: SQ is SHIFT-Q, CF1 is CTRL-F1, M3 is the
// middle mouse button and N7 is numpad 7.
func noteKey(token string) (string, bool) {
	if key, ok := noteBaseKey(token); ok {
		return key, true
	}
	if len(token) > 1 {
		mods := map[byte]string{'S': "SHIFT-", 'C': "CTRL-", 'A': "ALT-"}
		if mod, ok := mods[token[0]]; ok {
			if key, ok := noteBaseKey(token[1:]); ok {
				return mod + key, true
			}
		}
	}
	return "", false
}

var (
	fKeyPattern   = regexp.MustCompile(`^F([1-9]|1[0-2])$`)
	mouseKey      = regexp.MustCompile(`^M([3-5])$`)
	numpadNames   = map[string]string{"-": "NUMPADMINUS", "+": "NUMPADPLUS", ".": "NUMPADDECIMAL", "/": "NUMPADDIVIDE", "*": "NUMPADMULTIPLY"}
	plainKeyRunes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789`-=[];',./\\"
)

func noteBaseKey(token string) (string, bool) {
	switch {
	case len(token) == 1 && strings.Contains(plainKeyRunes, token):
		return token, true
	case fKeyPattern.MatchString(token):
		return token, true
	case mouseKey.MatchString(token):
		return "BUTTON" + token[1:], true
	case len(token) == 2 && token[0] == 'N':
		if token[1] >= '0' && token[1] <= '9' {
			return "NUMPAD" + token[1:], true
		}
		if name, ok := numpadNames[token[1:]]; ok {
			return name, true
		}
	}
	return "", false
}
