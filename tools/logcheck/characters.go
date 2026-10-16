package main

// We know each character by its GUID (Player-4620-01872808), never by its name.
//
// The name in the log is the game's to change, and it has changed. The beta wrote
// "Bayov-ClassicBetaPvE2-" on 2026-10-15 and "Bayov-Oom-ClassicBetaPvE2-" for the same
// character on 2026-10-16. Two characters can also share a name: we had two Bayovs on
// ClassicBetaPvE2 the same morning, and BayovCore wrote both into one history file.
//
// Each log tells us its own player. The client flags every unit it writes with how it relates
// to us, and only our own character has a player's flags with the "mine" affiliation (0x511).
// Party members are 0x512 and everyone else 0x518. So a log's lines say which GUID was us at
// each moment, with no name involved.
//
// Exports and history records made since 2026-10-16 carry the GUID (who.guid). The older ones
// don't, so we give each of them the GUID that the logs had as ours at its time.

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

// The parts of a unit's flags that say it's a player character of ours.
const (
	unitAffiliationMask = 0xF
	unitAffiliationMine = 0x1
	unitTypeMask        = 0xFC00
	unitTypePlayer      = 0x400
)

// isMine says whether a unit's flags (like "0x511") are our own character's.
func isMine(flags string) bool {
	n, err := strconv.ParseUint(strings.TrimPrefix(flags, "0x"), 16, 64)
	return err == nil && n&unitAffiliationMask == unitAffiliationMine && n&unitTypeMask == unitTypePlayer
}

// meGUID is our character's GUID on the line, or "" when we aren't on it.
func (e *event) meGUID() string {
	if len(e.f) > 7 {
		if strings.HasPrefix(e.srcGUID, "Player-") && isMine(e.f[3]) {
			return e.srcGUID
		}
		if strings.HasPrefix(e.dstGUID, "Player-") && isMine(e.f[7]) {
			return e.dstGUID
		}
	}
	return ""
}

// meSpan is a stretch of a log that one character of ours wrote.
type meSpan struct {
	GUID string    `json:"guid"`
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// meSpans splits the log by which character of ours is on its lines. A log has more than one
// when we log out and back in on another character without restarting the game.
func meSpans(all []*event) []meSpan {
	var spans []meSpan
	for _, e := range all {
		guid := e.meGUID()
		if guid == "" {
			continue
		}
		if n := len(spans); n > 0 && spans[n-1].GUID == guid {
			spans[n-1].To = e.t
			continue
		}
		spans = append(spans, meSpan{GUID: guid, From: e.t, To: e.t})
	}
	return spans
}

// meGUIDs are the characters of ours with lines between from and to.
func meGUIDs(all []*event, from, to time.Time) map[string]bool {
	guids := map[string]bool{}
	for _, e := range all {
		if !e.t.Before(from) && !e.t.After(to) {
			if guid := e.meGUID(); guid != "" {
				guids[guid] = true
			}
		}
	}
	return guids
}

// flavorDirs are the game folders the archive has logs of, like _classic_beta_. The release
// client will have a folder of its own (maybe _forever_), and we take every one we find.
func flavorDirs(archive string) []string {
	dirs, _ := filepath.Glob(filepath.Join(archive, "_*_"))
	return dirs
}

// allArchivedLogs are the logs of every game folder in the archive.
func allArchivedLogs(archive string) []string {
	var logs []string
	for _, dir := range flavorDirs(archive) {
		logs = append(logs, archivedLogs(dir)...)
	}
	return logs
}

// meCache is <archive>/audit/me.json: the meSpans of each log, so we read each log once
// rather than every time we place an export. A log that changed since (it grew, or the audit
// compressed it) is read again.
type meCache map[string]struct {
	Size  int64     `json:"size"`
	Mod   time.Time `json:"mod"`
	Spans []meSpan  `json:"spans"`
}

// archivedMeSpans is the meSpans of every log in the archive.
func archivedMeSpans(archive string) ([]meSpan, error) {
	path := filepath.Join(archive, "audit", "me.json")
	cache := meCache{}
	if data, err := os.ReadFile(path); err == nil {
		json.Unmarshal(data, &cache)
	}
	changed := false
	var spans []meSpan
	for _, log := range allArchivedLogs(archive) {
		info, err := os.Stat(log)
		if err != nil {
			return nil, err
		}
		key := filepath.Base(log)
		entry, ok := cache[key]
		if !ok || entry.Size != info.Size() || !entry.Mod.Equal(info.ModTime()) {
			all, err := loadLog(log)
			if err != nil {
				return nil, err
			}
			entry.Size, entry.Mod, entry.Spans = info.Size(), info.ModTime(), meSpans(all)
			cache[key] = entry
			changed = true
		}
		spans = append(spans, entry.Spans...)
	}
	if changed {
		if data, err := json.Marshal(cache); err == nil && os.MkdirAll(filepath.Dir(path), 0o755) == nil {
			tmp := path + ".tmp"
			if os.WriteFile(tmp, data, 0o644) == nil {
				os.Rename(tmp, path)
			}
		}
	}
	return spans, nil
}

// placeNear is how far from an export's time we still look for a log of ours to give it the
// GUID of. We take an export or a history record while we play, so the log has us within
// minutes of it.
const placeNear = 30 * time.Minute

// placeExports gives each export that doesn't carry a GUID the one the logs had as ours at its
// time, or the nearest one within placeNear. An export we can't place keeps an empty GUID, and
// no window uses it.
//
// The history records of one session (from a login to the next) are all one character, so they
// all take the GUID most of them were placed on. That also places a record from a quiet moment
// with no log lines near it.
func placeExports(exports []*export, spans []meSpan) {
	nearest := func(t time.Time) string {
		best, bestGap := "", placeNear+1
		for _, s := range spans {
			gap := time.Duration(0)
			if t.Before(s.From) {
				gap = s.From.Sub(t)
			} else if t.After(s.To) {
				gap = t.Sub(s.To)
			}
			if gap < bestGap {
				best, bestGap = s.GUID, gap
			}
		}
		if bestGap > placeNear {
			return ""
		}
		return best
	}

	sessions := map[string][]*export{}
	for _, ex := range exports {
		if ex.Who.GUID != "" {
			ex.guid = ex.Who.GUID
			continue
		}
		ex.guid = nearest(ex.time)
		if ex.history {
			sessions[ex.session] = append(sessions[ex.session], ex)
		}
	}
	for _, list := range sessions {
		votes := map[string]int{}
		for _, ex := range list {
			if ex.guid != "" {
				votes[ex.guid]++
			}
		}
		best := ""
		for _, guid := range slices.Sorted(maps.Keys(votes)) {
			if best == "" || votes[guid] > votes[best] {
				best = guid
			}
		}
		for _, ex := range list {
			ex.guid = best
		}
	}
}

// charName is how a report names the character: Name-Realm and the GUID.
func charName(ex *export) string {
	return ex.Who.Name + "-" + ex.Who.Realm + " (" + ex.guid + ")"
}

// ourChars are the characters of ours with lines between from and to that we have an export
// or a history of.
func ourChars(all []*event, exports []*export, from, to time.Time) map[string]bool {
	known := map[string]bool{}
	for _, ex := range exports {
		known[ex.guid] = true
	}
	chars := map[string]bool{}
	for guid := range meGUIDs(all, from, to) {
		if known[guid] {
			chars[guid] = true
		}
	}
	return chars
}

// findChar is the GUID that -char names. It takes a GUID, or a name as Name or Name-Realm for
// when we type it by hand. A name that more than one character had is the one with the newest
// export.
func findChar(exports []*export, char string) (string, error) {
	if strings.HasPrefix(char, "Player-") {
		return char, nil
	}
	var newest *export
	guids := map[string]bool{}
	for _, ex := range exports {
		if ex.guid == "" || !(strings.EqualFold(ex.Who.Name, char) || strings.EqualFold(ex.Who.Name+"-"+ex.Who.Realm, char)) {
			continue
		}
		guids[ex.guid] = true
		if newest == nil || ex.time.After(newest.time) {
			newest = ex
		}
	}
	if newest == nil {
		return "", fmt.Errorf("no exports and no history of a character named %s", char)
	}
	if len(guids) > 1 {
		fmt.Fprintf(os.Stderr, "%d characters were named %s (%s), we take %s, the one with the newest export\n",
			len(guids), char, strings.Join(slices.Sorted(maps.Keys(guids)), ", "), newest.guid)
	}
	return newest.guid, nil
}

// flavorLogs are the archived logs of one game folder, or of every one when flavor is "".
func flavorLogs(archive, flavor string) []string {
	if flavor == "" {
		return allArchivedLogs(archive)
	}
	return archivedLogs(filepath.Join(archive, flavor))
}
