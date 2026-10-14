package main

// The audit checks every stretch of our archived logs that we haven't checked yet.
//
//	go run --tags=with_db ./tools/logcheck -audit
//
// It splits each log into windows where the sim's character holds still: one of our
// characters, with the same gear, talents and level all through. It cuts at each change that
// BayovCore's history records, at each change of our logged item level (a gear swap that has no
// history), at each logout or /reload, and at breaks of more than 30 minutes. Then it runs
// logcheck on each window it hasn't run before, and saves the whole report in
// ~/forever-logs/audit/reports/. Each window goes in ~/forever-logs/audit/ledger.jsonl with what
// it flagged, so the next audit skips it and can tell which flags are new.
//
// A window waits until the history covers it. WoW saves the history only at logout and
// /reload, so while we play, a gear swap from 20 minutes ago may not be on disk yet. The addon
// writes a logout record just before WoW saves, and every record says the history is complete
// up to its time. So a window is ready once the character has a record from after it ends. A
// character with no history at all waits until the log hasn't changed for 30 minutes.
//
// With -redo it runs every window again, with the sim as it is now, and the newest run of a
// window is the one that counts.

import (
	"bufio"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

const (
	// auditBreak is the longest gap between two fights that stays in one window.
	auditBreak = 30 * time.Minute
	// auditIdle is how long a log has to stay unchanged before we check a character that has
	// no history.
	auditIdle = 30 * time.Minute
	// auditMinAttacks is the fewest attack lines of ours on mobs that make a window worth
	// running. A shorter one is too small for any row to flag, so we only note it.
	auditMinAttacks = 60
	// auditWorkers is how many windows run at once.
	auditWorkers = 4
)

// auditWindow is one stretch of one log where the sim's character holds still.
type auditWindow struct {
	Log  string `json:"log"`
	Char string `json:"char"`
	From string `json:"from"`
	To   string `json:"to"`
	// Ends says why the window ends: what changed, a break, a logout, or the end of the log.
	Ends    string `json:"ends"`
	Attacks int    `json:"attacks"`
}

func (w auditWindow) key() string { return w.Log + " " + w.Char + " " + w.From }

// ledgerEntry is one line of the ledger: a window and what checking it found.
type ledgerEntry struct {
	auditWindow
	Ran    string `json:"ran"`
	Report string `json:"report,omitempty"`
	Export string `json:"export,omitempty"`
	// Skipped says why we didn't run logcheck on the window.
	Skipped  string    `json:"skipped,omitempty"`
	Error    string    `json:"error,omitempty"`
	Findings []finding `json:"findings,omitempty"`
}

// cut is a stretch between two windows: what happened in it belongs to neither.
type cut struct {
	from, to time.Time
	why      string
}

func runAudit(archive, flavor, onlyLog string, iterations int, redo bool) error {
	if err := exec.Command(filepath.Join(repoRoot, "tools/combat_log/archive.sh")).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "archive.sh:", err)
	}
	dir := filepath.Join(archive, "audit")
	ledgerPath := filepath.Join(dir, "ledger.jsonl")
	if err := os.MkdirAll(filepath.Join(dir, "reports"), 0o755); err != nil {
		return err
	}
	ledger, err := readLedger(ledgerPath)
	if err != nil {
		return err
	}
	done := map[string]bool{}
	for _, entry := range ledger {
		done[entry.key()] = true
	}

	exports, err := loadExports(archive)
	if err != nil {
		return err
	}
	logs := []string{onlyLog}
	if onlyLog == "" {
		logs, _ = filepath.Glob(filepath.Join(archive, flavor, "WoWCombatLog-*.txt"))
		slices.Sort(logs)
	}

	var todo []auditWindow
	var waiting []string
	paths := map[string]string{}
	for _, path := range logs {
		paths[filepath.Base(path)] = path
		windows, wait, err := planLog(path, exports)
		if err != nil {
			return err
		}
		waiting = append(waiting, wait...)
		for _, w := range windows {
			if redo || !done[w.key()] {
				todo = append(todo, w)
			}
		}
	}

	self, err := os.Executable()
	if err != nil {
		return err
	}
	results := make([]ledgerEntry, len(todo))
	var wg sync.WaitGroup
	next := make(chan int)
	for range auditWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range next {
				results[i] = runWindow(self, archive, paths[todo[i].Log], todo[i], iterations)
			}
		}()
	}
	for i := range todo {
		next <- i
	}
	close(next)
	wg.Wait()

	f, err := os.OpenFile(ledgerPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	for _, entry := range results {
		data, _ := json.Marshal(entry)
		if _, err := f.Write(append(data, '\n')); err != nil {
			f.Close()
			return err
		}
	}
	if err := f.Close(); err != nil {
		return err
	}
	return printDigest(ledger, results, waiting)
}

func readLedger(path string) ([]ledgerEntry, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var ledger []ledgerEntry
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1<<20), 1<<24)
	for line := 1; scanner.Scan(); line++ {
		var entry ledgerEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			return nil, fmt.Errorf("%s:%d: %w", path, line, err)
		}
		ledger = append(ledger, entry)
	}
	return ledger, scanner.Err()
}

// planLog splits one log into windows, for each of our characters that played in it. It
// returns the windows that are ready, and a line for each one that waits for history.
func planLog(path string, exports []*export) ([]auditWindow, []string, error) {
	all, err := loadLog(path)
	if err != nil || len(all) == 0 {
		return nil, nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, nil, err
	}
	idle := time.Since(info.ModTime()) > auditIdle
	logStart, logEnd := all[0].t, all[len(all)-1].t

	var windows []auditWindow
	var waiting []string
	chars := logChars(all, exports, logStart, logEnd)
	for _, char := range slices.Sorted(maps.Keys(chars)) {
		cl, err := newCombatLog(all, char, logStart, logEnd)
		if err != nil {
			return nil, nil, err
		}
		var history []*export
		for _, ex := range exports {
			if ex.history && charKey(ex) == char {
				history = append(history, ex)
			}
		}
		slices.SortStableFunc(history, func(a, b *export) int { return a.time.Compare(b.time) })
		cuts := append(historyCuts(history), itemLevelCuts(cl, history)...)
		fights := cl.fights()
		for i := 1; i < len(fights); i++ {
			if gap := fights[i].start.Sub(fights[i-1].end); gap > auditBreak {
				cuts = append(cuts, cut{fights[i-1].end, fights[i].start, fmt.Sprintf("a break of %s", gap.Round(time.Minute))})
			}
		}
		slices.SortFunc(cuts, func(a, b cut) int { return a.from.Compare(b.from) })

		// The windows run from cut to cut, trimmed to the fights in them.
		start := logStart
		for i := 0; i <= len(cuts); i++ {
			end, ends := logEnd, "the end of the log"
			if i < len(cuts) {
				end, ends = cuts[i].from, cuts[i].why
			}
			from, to, attacks := fightSpan(cl, fights, start, end)
			if i < len(cuts) {
				start = cuts[i].to
			}
			if attacks == 0 {
				continue
			}
			// The window's times are whole seconds, so the end goes up to the next one.
			w := auditWindow{Log: filepath.Base(path), Char: char, From: from.Format(stampFormat), To: to.Add(time.Second - 1).Format(stampFormat), Ends: ends, Attacks: attacks}
			switch {
			case len(history) > 0 && history[len(history)-1].time.Before(to):
				waiting = append(waiting, fmt.Sprintf("%s %s to %s: its history ends at %s, /reload in game to save it",
					char, w.From, w.To, history[len(history)-1].time.Format(stampFormat)))
				continue
			case len(history) == 0 && !idle:
				waiting = append(waiting, fmt.Sprintf("%s %s to %s: the log is still growing", char, w.From, w.To))
				continue
			}
			windows = append(windows, w)
		}
	}
	return windows, waiting, nil
}

// historyCuts are the changes the character's history recorded that move the sim: gear,
// talents and level (skill ups get a warning in the report instead, because low skill goes up
// every few swings). And each logout, because the next login is a new session.
func historyCuts(history []*export) []cut {
	var cuts []cut
	for i, ex := range history {
		if ex.kind == "logout" {
			cuts = append(cuts, cut{ex.time, ex.time, "a logout or /reload"})
			continue
		}
		if i == 0 {
			continue
		}
		changes, _ := exportChanges(history[i-1], ex)
		if len(changes) > 0 {
			// The addon records a change about 2 sec after it, so we leave out the few
			// seconds before the record.
			cuts = append(cuts, cut{ex.time.Add(-historySettle), ex.time, strings.Join(changes, ", ")})
		}
	}
	return cuts
}

// itemLevelCuts are the gear swaps the history doesn't have. Each of our casts carries our
// item level, so a swap happened between the last cast with the old one and the first with the
// new one. A swap the history recorded in that stretch is already a cut.
func itemLevelCuts(cl *combatLog, history []*export) []cut {
	var cuts []cut
	var lastTime time.Time
	lastLevel := ""
	for _, e := range cl.events {
		if e.kind != "SPELL_CAST_SUCCESS" || e.srcGUID != cl.me || len(e.f) <= spellInfo+infoLevel {
			continue
		}
		level := e.f[spellInfo+infoLevel]
		if lastLevel != "" && level != lastLevel {
			recorded := slices.ContainsFunc(history, func(ex *export) bool {
				return ex.kind == "change" && !ex.time.Before(lastTime) && !ex.time.After(e.t.Add(historySettle))
			})
			if !recorded {
				cuts = append(cuts, cut{lastTime, e.t, fmt.Sprintf("item level %s to %s", lastLevel, level)})
			}
		}
		lastTime, lastLevel = e.t, level
	}
	return cuts
}

// fightSpan is the part of start to end that has fights in it, and how many attack lines of
// ours on mobs it has.
func fightSpan(cl *combatLog, fights []*fight, start, end time.Time) (from, to time.Time, attacks int) {
	for _, f := range fights {
		if f.end.Before(start) || f.start.After(end) {
			continue
		}
		s, e := f.start, f.end
		if s.Before(start) {
			s = start
		}
		if e.After(end) {
			e = end
		}
		if from.IsZero() || s.Before(from) {
			from = s
		}
		if e.After(to) {
			to = e
		}
	}
	if from.IsZero() {
		return from, to, 0
	}
	for _, e := range cl.events {
		if !e.t.Before(from) && !e.t.After(to) && cl.ours[e.srcGUID] && isCreature(e.dstGUID) && isAttack(e.kind) {
			attacks++
		}
	}
	return from, to, attacks
}

// runWindow runs logcheck on one window, as its own process, and saves its report.
func runWindow(self, archive, logPath string, w auditWindow, iterations int) ledgerEntry {
	entry := ledgerEntry{auditWindow: w, Ran: time.Now().Format(stampFormat)}
	if w.Attacks < auditMinAttacks {
		entry.Skipped = fmt.Sprintf("only %d attacks", w.Attacks)
		return entry
	}
	name := fmt.Sprintf("%s_%s_%s", strings.TrimSuffix(w.Log, ".txt"), w.Char, strings.NewReplacer("-", "", ":", "", " ", "_").Replace(w.From))
	entry.Report = filepath.Join(archive, "audit", "reports", name+".txt")
	flagsPath := filepath.Join(archive, "audit", "reports", name+".json")
	cmd := exec.Command(self, "-archive", archive, "-log", logPath, "-from", w.From, "-to", w.To,
		"-char", w.Char, "-iters", fmt.Sprint(iterations), "-flags-out", flagsPath)
	out, err := cmd.CombinedOutput()
	if writeErr := os.WriteFile(entry.Report, out, 0o644); writeErr != nil && err == nil {
		err = writeErr
	}
	if err != nil {
		entry.Error = fmt.Sprintf("%v: %s", err, lastLine(out))
		return entry
	}
	var result runResult
	data, err := os.ReadFile(flagsPath)
	if err == nil {
		err = json.Unmarshal(data, &result)
	}
	if err != nil {
		entry.Error = err.Error()
		return entry
	}
	os.Remove(flagsPath)
	entry.Export, entry.Findings = result.Export, result.Findings
	return entry
}

func lastLine(out []byte) string {
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	return lines[len(lines)-1]
}

// knownFlag is a flag we've looked into, from tools/logcheck/known.json. The digest marks it,
// so a new run shows what still needs a look.
type knownFlag struct {
	Kind    string `json:"kind"`
	Ability string `json:"ability,omitempty"`
	// Levels limits it to these mob levels. Empty is every level.
	Levels []int `json:"levels,omitempty"`
	// Why says what we found, and where it's tracked (a Need to Verify item in notes.md, a
	// sim fix, or a known limit of logcheck).
	Why string `json:"why"`
}

func readKnown() ([]knownFlag, error) {
	data, err := os.ReadFile(filepath.Join(repoRoot, "tools/logcheck/known.json"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var known []knownFlag
	if err := json.Unmarshal(data, &known); err != nil {
		return nil, fmt.Errorf("tools/logcheck/known.json: %w", err)
	}
	return known, nil
}

func knownWhy(known []knownFlag, f finding) (string, bool) {
	for _, k := range known {
		if k.Kind == f.Kind && k.Ability == f.Ability && (len(k.Levels) == 0 || slices.Contains(k.Levels, f.Level)) {
			return k.Why, true
		}
	}
	return "", false
}

// tainted says whether the character changed in the window in a way the sim can't follow. Its
// other flags may only show that, so the counts across the ledger leave them out.
func tainted(entry ledgerEntry) bool {
	return slices.ContainsFunc(entry.Findings, func(f finding) bool { return f.Kind == "change" })
}

// printDigest prints what this audit ran and found, and how often each flag came back across
// the whole ledger.
func printDigest(earlier, now []ledgerEntry, waiting []string) error {
	known, err := readKnown()
	if err != nil {
		return err
	}
	// The newest run of each window counts, so a -redo replaces the older ones.
	latest := map[string]ledgerEntry{}
	var order []string
	for _, entry := range append(slices.Clone(earlier), now...) {
		if _, ok := latest[entry.key()]; !ok {
			order = append(order, entry.key())
		}
		latest[entry.key()] = entry
	}
	type seen struct {
		windows int
		last    string
	}
	before := map[string]bool{}
	for _, entry := range earlier {
		for _, f := range entry.Findings {
			before[f.Kind+" "+f.Ability] = true
		}
	}
	counts := map[string]*seen{}
	var countOrder []string
	ran, left := 0, 0
	for _, key := range order {
		entry := latest[key]
		if entry.Report == "" || entry.Error != "" {
			continue
		}
		if tainted(entry) {
			left++
			continue
		}
		ran++
		inWindow := map[string]bool{}
		for _, f := range entry.Findings {
			k := f.Kind + " " + f.Ability
			if inWindow[k] {
				continue
			}
			inWindow[k] = true
			if counts[k] == nil {
				counts[k] = &seen{}
				countOrder = append(countOrder, k)
			}
			counts[k].windows++
			counts[k].last = entry.Log + " " + entry.From
		}
	}

	if len(now) == 0 {
		fmt.Println("No new windows to check.")
	}
	for _, entry := range now {
		fmt.Printf("\n%s  %s  %s to %s  (%d attacks, ends at %s)\n", entry.Log, entry.Char, entry.From, entry.To[11:], entry.Attacks, entry.Ends)
		switch {
		case entry.Skipped != "":
			fmt.Printf("  skipped: %s\n", entry.Skipped)
			continue
		case entry.Error != "":
			fmt.Printf("  ERROR: %s\n  report: %s\n", entry.Error, entry.Report)
			continue
		}
		fmt.Printf("  export: %s\n  report: %s\n", entry.Export, entry.Report)
		if len(entry.Findings) == 0 {
			fmt.Println("  nothing flagged")
		}
		if tainted(entry) {
			fmt.Println("  the character changed in the window, so the other flags may only show that")
		}
		for _, f := range entry.Findings {
			mark := "      "
			if _, ok := knownWhy(known, f); ok {
				mark = "known "
			} else if !before[f.Kind+" "+f.Ability] {
				mark = "NEW   "
			}
			fmt.Printf("  %s%s\n", mark, f.String())
		}
	}
	if len(waiting) > 0 {
		fmt.Println("\nWaiting:")
		for _, w := range waiting {
			fmt.Println("  " + w)
		}
	}
	if len(countOrder) > 0 {
		slices.SortStableFunc(countOrder, func(a, b string) int { return counts[b].windows - counts[a].windows })
		fmt.Printf("\nFlags across the ledger, by how many of its %d windows had them", ran)
		if left > 0 {
			fmt.Printf(" (%d more windows left out, where the character changed)", left)
		}
		fmt.Println(":")
		for _, k := range countOrder {
			kind, ability, _ := strings.Cut(k, " ")
			what := kind
			if ability != "" {
				what = ability + ": " + kind
			}
			// A known flag can be limited to some levels, so the table only says when it's
			// known on every level.
			note := ""
			if why, ok := knownWhy(known, finding{Kind: kind, Ability: ability}); ok {
				note = "  known: " + why
			}
			fmt.Printf("  %3d  %s (last %s)%s\n", counts[k].windows, what, counts[k].last, note)
		}
	}
	return nil
}
