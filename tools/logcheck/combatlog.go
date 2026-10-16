package main

import (
	"bufio"
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

// event is one line of a WoW combat log.
//
// The fields keep the client's order, so f[0] is the event name and f[1] the source GUID.
// Every line we read carries the advanced unit fields, even with ADVANCED_LOG_ENABLED,0 in the
// header (tools/combat_log/README.md has the details).
type event struct {
	t       time.Time
	kind    string
	srcGUID string
	src     string
	dstGUID string
	dst     string
	spellID int32
	spell   string
	f       []string
}

// Where the advanced unit fields sit. A swing has no spell prefix, so its fields start 3
// earlier than a spell's. The unit's level is the last advanced field: a mob's level, or a
// player's item level.
const (
	swingInfo = 9
	spellInfo = 12
	infoLevel = 18
)

// loadLog reads a combat log and returns its lines in time order.
//
// The file isn't always in time order around a log restart, so we sort it. Lines with the
// same time keep their file order, because a proc logged before a swing in the same ms didn't
// come from that swing.
//
// A log the audit compressed (compressLog) ends in .gz, and we read it the same way.
func loadLog(path string) ([]*event, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var in io.Reader = file
	if strings.HasSuffix(path, ".gz") {
		gz, err := gzip.NewReader(file)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		defer gz.Close()
		in = gz
	}
	return readEvents(in)
}

// readEvents reads combat log lines and returns them in time order, as loadLog says.
func readEvents(in io.Reader) ([]*event, error) {
	var events []*event
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		line := scanner.Text()
		stamp, rest, ok := strings.Cut(line, "  ")
		if !ok {
			continue
		}
		t, err := time.Parse("1/2/2006 15:04:05.0000", stamp)
		if err != nil {
			continue
		}
		reader := csv.NewReader(strings.NewReader(rest))
		reader.LazyQuotes = true
		f, err := reader.Read()
		if err != nil || len(f) == 0 {
			continue
		}
		e := &event{t: t, kind: f[0], f: f}
		if len(f) > 6 {
			e.srcGUID, e.src, e.dstGUID, e.dst = f[1], f[2], f[5], f[6]
		}
		if (strings.HasPrefix(e.kind, "SPELL_") || strings.HasPrefix(e.kind, "RANGE_")) && len(f) > 10 {
			id, _ := strconv.Atoi(f[9])
			e.spellID, e.spell = int32(id), f[10]
		}
		events = append(events, e)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	slices.SortStableFunc(events, func(a, b *event) int { return a.t.Compare(b.t) })
	return events, nil
}

// infoStart is where the event's advanced unit fields start, or -1 when it has none.
func (e *event) infoStart() int {
	start := spellInfo
	if strings.HasPrefix(e.kind, "SWING_") {
		start = swingInfo
	}
	if len(e.f) <= start+infoLevel {
		return -1
	}
	return start
}

// infoGUID is the unit the advanced fields describe. On a swing that's the source, and on
// spell damage it's the target.
func (e *event) infoGUID() string {
	if start := e.infoStart(); start >= 0 {
		return e.f[start]
	}
	return ""
}

func (e *event) infoInt(offset int) int {
	start := e.infoStart()
	if start < 0 {
		return 0
	}
	n, _ := strconv.ParseFloat(e.f[start+offset], 64)
	return int(n)
}

// damage is the suffix of a *_DAMAGE line.
//
// raw is the hit before the target's armor, the crit, any partial resist and any damage taken
// modifier on the target, so it compares straight to the sim's damage before the target's
// modifiers.
type damage struct {
	amount, raw, overkill, school, resisted, blocked, absorbed int
	crit, glance, crush                                        bool
}

func (e *event) damage() (damage, bool) {
	start := e.infoStart()
	if start < 0 || !strings.HasSuffix(e.kind, "_DAMAGE") {
		return damage{}, false
	}
	f := e.f[start+infoLevel+1:]
	if len(f) < 10 {
		return damage{}, false
	}
	n := func(s string) int {
		v, _ := strconv.ParseFloat(s, 64)
		return int(v)
	}
	return damage{
		amount: n(f[0]), raw: n(f[1]), overkill: n(f[2]), school: n(f[3]),
		resisted: n(f[4]), blocked: n(f[5]), absorbed: n(f[6]),
		crit: f[7] == "1", glance: f[8] == "1", crush: f[9] == "1",
	}, true
}

// missType is MISS, DODGE, PARRY, BLOCK, RESIST, IMMUNE, ABSORB, EVADE and so on.
func (e *event) missType() string {
	i := spellInfo
	if e.kind == "SWING_MISSED" {
		i = swingInfo
	}
	if len(e.f) > i {
		return e.f[i]
	}
	return ""
}

func isCreature(guid string) bool {
	return strings.HasPrefix(guid, "Creature-") || strings.HasPrefix(guid, "Vehicle-")
}

// combatLog is the part of a log we look at, seen from one player.
type combatLog struct {
	events []*event
	me     string          // our GUID
	ours   map[string]bool // our GUID and the GUIDs of our totems
	levels map[string]int  // mob GUID to level
	from   time.Time
	to     time.Time
}

// newCombatLog takes the lines from `from` to `to`, and finds us by the character's
// Name-Realm.
func newCombatLog(all []*event, char string, from, to time.Time) (*combatLog, error) {
	cl := &combatLog{ours: map[string]bool{}, levels: map[string]int{}, from: from, to: to}
	for _, e := range all {
		if cl.me == "" && isChar(e.src, char) && strings.HasPrefix(e.srcGUID, "Player-") {
			cl.me = e.srcGUID
		}
		// Mob levels come from every line where the mob is the info unit, even outside the
		// window, because a mob we only hit with white swings never is.
		if guid := e.infoGUID(); isCreature(guid) {
			if level := e.infoInt(infoLevel); level > 0 {
				cl.levels[guid] = level
			}
		}
	}
	if cl.me == "" {
		return nil, fmt.Errorf("no lines from %s in the log", char)
	}
	cl.ours[cl.me] = true
	for _, e := range all {
		if e.t.Before(from) || e.t.After(to) {
			continue
		}
		cl.events = append(cl.events, e)
		if e.kind == "SPELL_SUMMON" && e.srcGUID == cl.me {
			cl.ours[e.dstGUID] = true
		}
	}
	if len(cl.events) == 0 {
		return nil, fmt.Errorf("no lines between %s and %s", from.Format(stampFormat), to.Format(stampFormat))
	}
	return cl, nil
}

// fight is the time we spent on one mob.
type fight struct {
	guid       string
	name       string
	level      int
	start, end time.Time
}

// fights splits our time into fights, one per mob, from the first line between us and the mob
// to the last. A mob we leave alone for more than 15 sec and meet again is a new fight.
func (cl *combatLog) fights() []*fight {
	const gap = 15 * time.Second
	open := map[string]*fight{}
	var fights []*fight
	for _, e := range cl.events {
		var mob, name string
		switch {
		case cl.ours[e.srcGUID] && isCreature(e.dstGUID) && isAttack(e.kind):
			mob, name = e.dstGUID, e.dst
		case e.dstGUID == cl.me && isCreature(e.srcGUID) && isAttack(e.kind):
			mob, name = e.srcGUID, e.src
		default:
			continue
		}
		f := open[mob]
		if f == nil || e.t.Sub(f.end) > gap {
			f = &fight{guid: mob, name: name, level: cl.levels[mob], start: e.t, end: e.t}
			open[mob] = f
			fights = append(fights, f)
		}
		f.end = e.t
	}
	return fights
}

// logName is a log's file name without the .gz of a compressed one, so the ledger knows a log
// by one name before and after we compress it.
func logName(path string) string {
	return strings.TrimSuffix(filepath.Base(path), ".gz")
}

// archivedLogs are the logs in dir, plain and compressed.
func archivedLogs(dir string) []string {
	plain, _ := filepath.Glob(filepath.Join(dir, "WoWCombatLog-*.txt"))
	compressed, _ := filepath.Glob(filepath.Join(dir, "WoWCombatLog-*.txt.gz"))
	logs := append(plain, compressed...)
	slices.SortFunc(logs, func(a, b string) int { return strings.Compare(logName(a), logName(b)) })
	return logs
}

// isChar says whether a name in the log is the character Name-Realm. The log adds a dash
// and a region after the realm, which is empty on the beta ("Bayov-ClassicBetaPvP-").
//
// Since the 2026-10-16 logs, the beta can also put an extra part between the name and the
// realm ("Bayov-Oom-ClassicBetaPvE2-"), and a log can have both forms. So we match the name
// and the realm, and skip what's between them.
func isChar(name, char string) bool {
	if name == char || strings.HasPrefix(name, char+"-") {
		return true
	}
	charName, realm, ok := strings.Cut(char, "-")
	if !ok {
		return false
	}
	parts := strings.Split(name, "-")
	return len(parts) > 2 && parts[0] == charName && slices.Contains(parts[2:], realm)
}

// logChars is which of the characters we have exports of played between from and to.
func logChars(all []*event, exports []*export, from, to time.Time) map[string]bool {
	known := map[string]bool{}
	for _, ex := range exports {
		known[charKey(ex)] = true
	}
	names := map[string]bool{}
	for _, e := range all {
		if !e.t.Before(from) && !e.t.After(to) && strings.HasPrefix(e.srcGUID, "Player-") {
			names[e.src] = true
		}
	}
	found := map[string]bool{}
	for char := range known {
		for name := range names {
			if isChar(name, char) {
				found[char] = true
			}
		}
	}
	return found
}

func isAttack(kind string) bool {
	return strings.HasSuffix(kind, "_DAMAGE") || strings.HasSuffix(kind, "_MISSED")
}

const stampFormat = "2006-01-02 15:04:05"
