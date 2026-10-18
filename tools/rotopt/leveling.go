package main

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// Leveling settings (an encounter with leveling options, sim/core/leveling.go): instead of
// DPS, a candidate scores the mobs it kills in a session, or with -objective xp its XP an
// hour, for comparing mob levels. The searches, -eval and -apl work the same.

// What a leveling run scores: "mobs" or "xp".
var levelingObjective = "mobs"

// The random seed of every scoring run. It's fixed so two candidates see the same sessions
// (or fights), and -seed changes it to check a gain holds on others.
var simSeed int64 = 1

func levelingResult(m *proto.LevelingMetrics, iterations int32) result {
	n := math.Sqrt(float64(iterations))
	r := result{leveling: m, dps: m.MobsKilled, stderr: m.MobsKilledStdev / n}
	if levelingObjective == "xp" && m.MobsKilled > 0 {
		// The spread of XP follows the spread of kills, at the session's XP a kill.
		r.dps = m.XpPerHour
		r.stderr = m.XpPerHour / m.MobsKilled * m.MobsKilledStdev / n
	}
	return r
}

func levelingString(r result) string {
	m := r.leveling
	out := fmt.Sprintf("%.2f(%.2f)", r.dps, r.stderr)
	if levelingObjective == "xp" {
		out = fmt.Sprintf("%.0f(%.0f) XP/h mobs %.1f", r.dps, r.stderr, m.MobsKilled)
	}
	return out + fmt.Sprintf(" deaths %.2f mobs/pull %.2f combat %.0fs rest %.0fs search %.0fs dead %.0fs food %.1f drinks %.1f bandages %.1f",
		m.Deaths, m.MobsPerPull, m.SecondsInCombat, m.SecondsResting, m.SecondsSearching, m.SecondsDead, m.Food, m.Drinks, m.Bandages)
}

// session runs one session with the combat log and boils it down to what an auditor needs
// to judge the rotation's choices: each search, pull and rest, with health and mana before
// and after, how each mob was pulled, the casts in each fight and between fights, and the
// time each took.
func (s *setup) session(rot *proto.APLRotation, seed int64) string {
	request := s.clone()
	request.Raid.Parties[0].Players[0].Rotation = rot
	request.SimOptions.Iterations = 1
	request.SimOptions.RandomSeed = seed
	request.SimOptions.Debug = true

	var b strings.Builder
	stats := core.ComputeStats(&proto.ComputeStatsRequest{Raid: request.Raid, Encounter: request.Encounter, Ruleset: request.SimOptions.Ruleset})
	if stats.RaidStats == nil {
		return "computing stats failed"
	}
	player := stats.RaidStats.Parties[0].Players[0]
	for i, item := range player.RotationStats.GetPriorityList() {
		for _, w := range item.Warnings {
			fmt.Fprintf(&b, "rotation line %d: %s\n", i+1, w)
		}
	}
	maxHealth := player.FinalStats.Stats[proto.Stat_StatHealth]
	maxMana := player.FinalStats.Stats[proto.Stat_StatMana]

	res := core.RunRaidSim(request)
	if res.Error != nil {
		return "sim failed: " + res.Error.Message
	}
	m := res.EncounterMetrics.GetLeveling()
	fmt.Fprintf(&b, "session (seed %d): %.1f mobs, %.0f XP, %.0f deaths, %.0fs combat, %.0fs resting, %.0fs searching and pulling, %.0fs dead, %.0f food, %.0f drinks, %.0f bandages\n",
		seed, m.MobsKilled, m.Xp, m.Deaths, m.SecondsInCombat, m.SecondsResting, m.SecondsSearching, m.SecondsDead, m.Food, m.Drinks, m.Bandages)
	fmt.Fprintf(&b, "max health %.0f, max mana %.0f; health and mana below are %% of max\n\n", maxHealth, maxMana)
	length := request.Encounter.GetLeveling().GetSessionLength()
	if length <= 0 {
		length = 1200
	}
	b.WriteString(summarizeSession(res.Logs, maxHealth, maxMana, length))
	return b.String()
}

// sessionLog is one session's whole log, the way the UI gets it (debug lines too).
func (s *setup) sessionLog(rot *proto.APLRotation, seed int64) string {
	request := s.clone()
	request.Raid.Parties[0].Players[0].Rotation = rot
	request.SimOptions.Iterations = 1
	request.SimOptions.RandomSeed = seed
	request.SimOptions.Debug = true
	res := core.RunRaidSim(request)
	if res.Error != nil {
		return "sim failed: " + res.Error.Message
	}
	return res.Logs
}

var (
	sessionTime     = regexp.MustCompile(`^\s*\[(-?[0-9.]+)\] `)
	sessionPlayer   = regexp.MustCompile(`^\s*\[[0-9.]+\] \[[^\]]*\(#1\)\] `)
	sessionHealth   = regexp.MustCompile(`\] \[[^\]]*\(#1\)\] (?:Spent|Gained) [0-9.]+ health from .*\(([0-9.]+) --> ([0-9.]+)\)`)
	sessionMana     = regexp.MustCompile(`\] \[[^\]]*\(#1\)\] (?:Spent|Gained) [0-9.]+ mana from .*\(([0-9.]+) --> ([0-9.]+)\)`)
	sessionDamage   = regexp.MustCompile(`\] \[[^\]]*\(#1\)\] Spent ([0-9.]+) health from \{OtherID: 11\}`)
	sessionCast     = regexp.MustCompile(`\] \[[^\]]*\(#1\)\] Casting \{SpellID: (\d+)\}`)
	sessionLeveling = regexp.MustCompile(`Leveling: (.*)$`)
	sessionMobDied  = regexp.MustCompile(`\] \[(Target \d+)\] Died`)
)

// summarizeSession reads a one session debug log (see session).
func summarizeSession(logs string, maxHealth, maxMana, length float64) string {
	var b strings.Builder
	health, mana := maxHealth, maxMana
	pct := func(v, max float64) string {
		if max <= 0 {
			return "-"
		}
		return fmt.Sprintf("%.0f%%", 100*v/max)
	}
	state := fmt.Sprintf // shorthand

	// What we're doing now, since when, and what happened in it.
	type phase struct {
		kind         string // search, pull, fight, rest, bandage, ready, dead
		start        float64
		health, mana float64
		casts        map[int32]int
		damage       float64
		events       []string
		mobs, kills  int
	}
	var cur *phase
	begin := func(kind string, at float64) {
		cur = &phase{kind: kind, start: at, health: health, mana: mana, casts: map[int32]int{}}
	}
	castList := func(casts map[int32]int) string {
		type kv struct {
			id int32
			n  int
		}
		var list []kv
		for id, n := range casts {
			list = append(list, kv{id, n})
		}
		sort.Slice(list, func(i, j int) bool { return list[i].n > list[j].n || list[i].n == list[j].n && list[i].id < list[j].id })
		var parts []string
		for _, c := range list {
			name, ok := spellNames[c.id]
			if !ok {
				name = strconv.Itoa(int(c.id))
			}
			parts = append(parts, fmt.Sprintf("%s %d", name, c.n))
		}
		return strings.Join(parts, ", ")
	}
	end := func(at float64) {
		if cur == nil {
			return
		}
		line := state("%7.1f  %-7s %5.1fs  health %s→%s  mana %s→%s", cur.start, cur.kind, at-cur.start,
			pct(cur.health, maxHealth), pct(health, maxHealth), pct(cur.mana, maxMana), pct(mana, maxMana))
		if cur.kind == "fight" {
			line += state("  %d mobs, %d killed, %.0f damage taken", cur.mobs, cur.kills, cur.damage)
		}
		if len(cur.casts) > 0 {
			line += "  casts: " + castList(cur.casts)
		}
		b.WriteString(line + "\n")
		for _, e := range cur.events {
			b.WriteString("           " + e + "\n")
		}
		cur = nil
	}

	for _, line := range strings.Split(logs, "\n") {
		tm := sessionTime.FindStringSubmatch(line)
		if tm == nil {
			continue
		}
		at, _ := strconv.ParseFloat(tm[1], 64)
		if h := sessionHealth.FindStringSubmatch(line); h != nil {
			health, _ = strconv.ParseFloat(h[2], 64)
		}
		if mm := sessionMana.FindStringSubmatch(line); mm != nil {
			mana, _ = strconv.ParseFloat(mm[2], 64)
		}
		if cur != nil {
			if d := sessionDamage.FindStringSubmatch(line); d != nil {
				v, _ := strconv.ParseFloat(d[1], 64)
				cur.damage += v
			}
			if c := sessionCast.FindStringSubmatch(line); c != nil {
				id, _ := strconv.Atoi(c[1])
				cur.casts[int32(id)]++
			}
			if sessionMobDied.MatchString(line) && cur.kind == "fight" {
				cur.kills++
			}
		}
		if !sessionPlayer.MatchString(line) {
			continue
		}
		lv := sessionLeveling.FindStringSubmatch(line)
		if lv == nil {
			continue
		}
		msg := lv[1]
		switch {
		case strings.HasPrefix(msg, "searching"):
			end(at)
			begin("search", at)
		case strings.HasPrefix(msg, "found"):
			if cur != nil && cur.kind == "fight" {
				cur.events = append(cur.events, state("%.1f %s", at, msg))
				continue
			}
			end(at)
			begin("pull", at)
			cur.events = append(cur.events, msg)
		case strings.HasPrefix(msg, "pulled"):
			if cur == nil || cur.kind != "fight" {
				if cur != nil && cur.kind == "pull" {
					cur.events = append(cur.events, state("%.1f %s", at, msg))
				}
				end(at)
				begin("fight", at)
			} else {
				cur.events = append(cur.events, state("%.1f %s", at, msg))
			}
			cur.mobs++
		case strings.HasPrefix(msg, "out of combat"):
			end(at)
			begin("ready", at)
		case strings.HasPrefix(msg, "sitting down"):
			end(at)
			begin("rest", at)
			cur.events = append(cur.events, msg)
		case strings.HasPrefix(msg, "bandaging"):
			end(at)
			begin("bandage", at)
		case strings.HasPrefix(msg, "died"):
			end(at)
			begin("dead", at)
		default:
			if cur != nil {
				cur.events = append(cur.events, state("%.1f %s", at, msg))
			}
		}
	}
	end(length)
	return b.String()
}
