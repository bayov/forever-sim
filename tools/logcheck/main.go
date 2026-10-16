// logcheck compares a WoW Forever combat log with the sim.
//
// It takes a log window and the /bayov export of the character that played it, builds that
// character in the sim, and replays our casts at the times the log has them, against a mob of
// the same level, many times over. Then it compares the two side by side for each ability:
// how often it hit, crit, glanced, missed, was dodged and so on, its raw damage (before the
// mob's armor and resists), its partial resists, and how often each proc came per landed
// white swing. A row the sim can't explain gets a flag, so each new log rechecks everything
// we've settled so far.
//
//	go run --tags=with_db ./tools/logcheck -from 20:25:00 -to 21:10:00
//
// With no -log it archives the game's logs first (tools/combat_log/archive.sh) and reads the
// newest one in ~/forever-logs. With no -export it takes the newest export taken before the
// window ends, from ~/forever-logs/exports and from BayovCore's history in
// ~/forever-logs/history (one export for each record). Save each pasted /bayov export in
// exports/ as export-<date>_<time>.json, with the date and time from its when.date.
//
// It has to run from the repo root, because it reads the talent trees and the item database
// from there.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/core/proto"
)

var repoRoot = "."

func main() {
	home, _ := os.UserHomeDir()
	archive := flag.String("archive", filepath.Join(home, "forever-logs"), "the log archive: logs under <flavor>/, exports under exports/, client tables under wago/")
	flavor := flag.String("flavor", "_classic_beta_", "the game folder whose newest log we read when -log is empty")
	logPath := flag.String("log", "", "the combat log to read (default: the newest one in the archive)")
	exportPath := flag.String("export", "", "the /bayov export JSON of the character (default: the newest one before the window ends)")
	fromFlag := flag.String("from", "", "start of the window, HH:MM:SS on the log's first day (a time before the log starts is taken as the next day), or YYYY-MM-DD HH:MM:SS")
	toFlag := flag.String("to", "", "end of the window, like -from")
	iterations := flag.Int("iters", 200, "replays of the window per mob level")
	minAttempts := flag.Int("min", 5, "leave out abilities with fewer attempts than this in the log")
	rounding := flag.String("rounding", "floor", "how the server turns damage into whole points, floor or round")
	fullSkill := flag.Bool("full-skill", false, "sim full weapon skill instead of the export's")
	behind := flag.Bool("behind", false, "sim us behind the mob, where it can't parry or block (by default we're in front, as when questing)")
	charFlag := flag.String("char", "", "the character who played the window, as Name-Realm (default: the one of ours with lines in the window)")
	flagsOut := flag.String("flags-out", "", "also write the flagged list to this file as JSON")
	audit := flag.Bool("audit", false, "check every stretch of the archived logs we haven't checked yet (audit.go), instead of one window")
	redo := flag.Bool("redo", false, "with -audit, check again the windows the ledger already has, with the sim as it is now")
	mobs := flag.Bool("mobs", false, "measure the mobs that swung at our characters, in every archived log or -log, against the leveling sim's normal mob table (mobs.go), instead of one window")
	pacingFlag := flag.Bool("pacing", false, "measure how we paced our grinding (time from a kill to the next pull, pulls, run-ins, pack sizes) in every archived log or -log, within -from and -to, for the leveling sim's session defaults (pacing.go), instead of one window")
	playerOut := flag.String("player-out", "", "write the sim player built from -char's newest export (or -export) to this file as JSON and exit, without a log (player.go)")
	flag.Parse()
	sim.RegisterAll()

	if *rounding != "floor" && *rounding != "round" {
		log.Fatal("-rounding is floor or round")
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "ui/core/talents/trees")); err != nil {
		log.Fatal("run logcheck from the repo root")
	}

	if *playerOut != "" {
		if err := writePlayer(*archive, *exportPath, *charFlag, *fullSkill, *playerOut); err != nil {
			log.Fatal(err)
		}
		return
	}

	if *mobs {
		if err := runMobs(*archive, *flavor, *logPath); err != nil {
			log.Fatal(err)
		}
		return
	}

	if *pacingFlag {
		if err := runPacing(*archive, *flavor, *logPath, *fromFlag, *toFlag); err != nil {
			log.Fatal(err)
		}
		return
	}

	if *audit {
		if err := runAudit(*archive, *flavor, *logPath, *iterations, *redo); err != nil {
			log.Fatal(err)
		}
		return
	}

	if *logPath == "" {
		if err := exec.Command(filepath.Join(repoRoot, "tools/combat_log/archive.sh")).Run(); err != nil {
			fmt.Fprintln(os.Stderr, "archive.sh:", err)
		}
		*logPath = newestLog(filepath.Join(*archive, *flavor))
		if *logPath == "" {
			log.Fatalf("no combat logs in %s", filepath.Join(*archive, *flavor))
		}
	}
	all, err := loadLog(*logPath)
	if err != nil {
		log.Fatal(err)
	}
	if len(all) == 0 {
		log.Fatalf("%s has no lines", *logPath)
	}
	from, err := parseWindow(*fromFlag, all[0].t, all[0].t)
	if err != nil {
		log.Fatal(err)
	}
	to, err := parseWindow(*toFlag, all[0].t, all[len(all)-1].t)
	if err != nil {
		log.Fatal(err)
	}

	exports, err := loadExports(*archive)
	if err != nil {
		log.Fatal(err)
	}
	var ex *export
	switch {
	case *exportPath != "":
		ex, err = loadExport(*exportPath)
	case *charFlag != "":
		ex, err = pickExport(exports, to, map[string]bool{*charFlag: true})
	default:
		// The log has every player near us, and we may have played more than one of our
		// characters in it, so we take an export of one that played in the window.
		chars := logChars(all, exports, from, to)
		if len(chars) == 0 {
			log.Fatal("none of the characters we have exports of played in the window")
		}
		if len(chars) > 1 {
			log.Fatalf("more than one of our characters played in the window (%s), pick one with -char",
				strings.Join(slices.Sorted(maps.Keys(chars)), ", "))
		}
		ex, err = pickExport(exports, to, chars)
	}
	if err != nil {
		log.Fatal(err)
	}

	cl, err := newCombatLog(all, charKey(ex), from, to)
	if err != nil {
		log.Fatal(err)
	}
	build := fmt.Sprintf("%s.%d", ex.When.Version, ex.When.Build)
	tables, err := loadClientTables(filepath.Join(*archive, "wago"), build)
	if err != nil {
		log.Fatal(err)
	}
	char, err := buildPlayer(ex, tables)
	if err != nil {
		log.Fatal(err)
	}
	char.player.InFrontOfTarget = !*behind
	if !*fullSkill {
		setSkillGaps(char.player, ex)
	}
	setImbue(char, all, cl, ex, tables)

	r := &report{cl: cl, ex: ex, char: char, tables: tables, logPath: *logPath, minAttempts: *minAttempts, missing: map[string]bool{}, aliased: map[int32]bool{}, pooled: map[abilityKey]*pooledHits{}, rounding: *rounding}
	r.header()
	r.sheet()
	r.weaponSkills(*fullSkill)
	r.exportChanges(exportsAround(exports, ex, from, to))
	r.logNotes()

	logTallies := tallyLog(cl)
	for _, p := range plans(cl) {
		simTallies, stats := replay(char.player, p, *iterations)
		r.level(p, logTallies, simTallies, stats)
	}
	r.amountAllLevels()
	fmt.Println()
	r.flags()
	if *flagsOut != "" {
		data, err := json.Marshal(runResult{Export: ex.path, Findings: r.flagged})
		if err == nil {
			err = os.WriteFile(*flagsOut, data, 0o644)
		}
		if err != nil {
			log.Fatal(err)
		}
	}
}

// runResult is what -flags-out writes, for the audit.
type runResult struct {
	Export   string    `json:"export"`
	Findings []finding `json:"findings"`
}

func newestLog(dir string) string {
	paths := archivedLogs(dir)
	newest := ""
	var newestTime time.Time
	for _, path := range paths {
		if info, err := os.Stat(path); err == nil && info.ModTime().After(newestTime) {
			newest, newestTime = path, info.ModTime()
		}
	}
	return newest
}

// parseWindow reads a -from or -to value. An empty one gives def.
func parseWindow(value string, logStart, def time.Time) (time.Time, error) {
	if value == "" {
		return def, nil
	}
	if t, err := time.Parse(stampFormat, value); err == nil {
		return t, nil
	}
	clock, err := time.Parse("15:04:05", value)
	if err != nil {
		return time.Time{}, fmt.Errorf("window time %q: want HH:MM:SS or YYYY-MM-DD HH:MM:SS", value)
	}
	t := time.Date(logStart.Year(), logStart.Month(), logStart.Day(), clock.Hour(), clock.Minute(), clock.Second(), 0, time.UTC)
	if t.Before(logStart.Truncate(time.Second)) {
		t = t.AddDate(0, 0, 1)
	}
	return t, nil
}

// setImbue gives the sim our weapon imbue.
//
// Under Forever a shaman's imbue is a buff on the character, and the export doesn't see it, so
// we take the last imbue we cast before the window ends. An oil or a stone does show in the
// export as the weapon's temporary enchant.
func setImbue(char *character, all []*event, cl *combatLog, ex *export, tables *clientTables) {
	var imbue proto.WeaponImbue
	name := ""
	var during []string
	for _, e := range all {
		if e.kind != "SPELL_CAST_SUCCESS" || e.srcGUID != cl.me || e.t.After(cl.to) {
			continue
		}
		if value, ok := imbueFromName(e.spell); ok {
			imbue, name = value, e.spell
			if !e.t.Before(cl.from) && !slices.Contains(during, e.spell) {
				during = append(during, e.spell)
			}
		}
	}
	// The imbue lasts 30 minutes and survives a relog, so its cast can be in an earlier log
	// file. Then its procs in the window tell which one it is.
	if imbue == 0 {
		for _, e := range cl.events {
			if e.srcGUID != cl.me || !isAttack(e.kind) {
				continue
			}
			if value, ok := imbueProcs[e.spellID]; ok {
				imbue, name = value, e.spell+" procs in the window"
				break
			}
		}
	}
	if enh := char.player.GetEnhancementShaman(); enh != nil && imbue != 0 {
		enh.Options.ShamanImbue = imbue
		char.notes = append(char.notes, "shaman imbue from the log: "+name)
	}
	if len(during) > 1 {
		char.notes = append(char.notes, "imbues changed in the window ("+strings.Join(during, ", ")+"), the sim keeps the last one for all of it")
	}

	for hand, temp := range ex.Imbues {
		for _, spell := range tables.enchantSpells[temp.Enchant] {
			value, ok := imbueFromName(tables.spellNames[spell])
			if !ok {
				continue
			}
			if hand == "off" {
				char.player.Consumes.OffHandImbue = value
			} else {
				char.player.Consumes.MainHandImbue = value
			}
			break
		}
	}
}

// imbueProcs are the spells our shaman imbues hit with, by the log's spell ID. Flametongue
// Totem's attack has the same name as the imbue's (16368 against 10444), so we go by ID.
var imbueProcs = map[int32]proto.WeaponImbue{
	439440: proto.WeaponImbue_WindfuryWeapon,
	10444:  proto.WeaponImbue_FlametongueWeapon,
}

// tallyLog counts our attacks on mobs in the window, by mob level.
func tallyLog(cl *combatLog) *tallies {
	ts := newTallies()
	ts.iterations = 1
	for _, e := range cl.events {
		if !cl.ours[e.srcGUID] || !isCreature(e.dstGUID) {
			continue
		}
		level := cl.levels[e.dstGUID]
		if level == 0 {
			continue
		}
		key := abilityKey{spell: e.spellID}
		switch e.kind {
		case "SWING_DAMAGE", "SWING_MISSED":
			key.spell = 0
		case "SPELL_DAMAGE", "SPELL_MISSED", "RANGE_DAMAGE", "RANGE_MISSED":
		case "SPELL_PERIODIC_DAMAGE", "SPELL_PERIODIC_MISSED":
			key.periodic = true
		default:
			continue
		}
		d, isDamage := e.damage()
		outcome := logOutcome(e, d, isDamage)
		physical := key.spell == 0 || (isDamage && d.school == 1) || (!isDamage && len(e.f) > 11 && e.f[11] == "0x1")
		share := 0.0
		if isDamage && d.resisted > 0 {
			share = float64(d.resisted) / float64(d.amount+d.resisted+d.absorbed)
		}
		ts.add(level, key, outcome, float64(d.raw), physical, share)
		if isDamage && (outcome == "hit" || outcome == "crit") && d.resisted == 0 && d.blocked == 0 && d.absorbed == 0 {
			t := ts.get(level, key)
			t.hits = append(t.hits, landedHit{raw: float64(d.raw), amount: float64(d.amount), crit: d.crit})
		}
	}
	return ts
}

// plans builds one replay per mob level we fought.
func plans(cl *combatLog) []*plan {
	t0 := cl.from
	byLevel := map[int]*plan{}
	var levels []int
	for _, f := range cl.fights() {
		if f.level == 0 {
			continue
		}
		p := byLevel[f.level]
		if p == nil {
			p = &plan{level: f.level}
			byLevel[f.level] = p
			levels = append(levels, f.level)
		}
		p.fights = append(p.fights, f)
		p.windows = append(p.windows, [2]time.Duration{f.start.Sub(t0), f.end.Sub(t0)})
		p.duration = max(p.duration, f.end.Sub(t0)+2*time.Second)
	}
	slices.Sort(levels)

	// The spells that dealt damage themselves. Our other casts are buffs and totems. We go by
	// name, because a cast can hit under another ID (Fire Nova casts 408342 and hits with
	// 408424).
	damaging := map[string]bool{}
	for _, e := range cl.events {
		if e.srcGUID == cl.me && e.spellID != 0 && isAttack(e.kind) {
			damaging[e.spell] = true
		}
	}
	// A cast with a cast time goes from its SPELL_CAST_START.
	starts := map[int32]time.Time{}
	var result []*plan
	for _, level := range levels {
		p := byLevel[level]
		inFight := func(t time.Duration) bool {
			for _, w := range p.windows {
				if t >= w[0]-time.Second && t <= w[1]+time.Second {
					return true
				}
			}
			return false
		}
		clear(starts)
		for _, e := range cl.events {
			if e.srcGUID != cl.me {
				continue
			}
			if e.kind == "SPELL_CAST_START" {
				starts[e.spellID] = e.t
				continue
			}
			if e.kind != "SPELL_CAST_SUCCESS" {
				continue
			}
			if _, ok := imbueFromName(e.spell); ok {
				continue
			}
			start := e.t
			if s, ok := starts[e.spellID]; ok && e.t.Sub(s) < 5*time.Second {
				start = s
				delete(starts, e.spellID)
			}
			at := start.Sub(t0)
			if at > p.duration {
				continue
			}
			if damaging[e.spell] {
				if isCreature(e.dstGUID) {
					if cl.levels[e.dstGUID] != level {
						continue
					}
				} else if !inFight(at) {
					continue
				}
			}
			p.casts = append(p.casts, replayCast{at: at, spellID: e.spellID, name: e.spell})
		}

		hits := 0
		for _, e := range cl.events {
			if e.dstGUID == cl.me && isCreature(e.srcGUID) && strings.HasSuffix(e.kind, "_DAMAGE") && cl.levels[e.srcGUID] == level {
				hits++
			}
		}
		p.hitsPerMinute = float64(hits) / p.duration.Minutes()
		result = append(result, p)
	}
	return result
}
