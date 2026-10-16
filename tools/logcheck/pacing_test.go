package main

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

const (
	pacingMe  = "Player-1-0001"
	pacingMob = "Creature-0-1-2-3-100-000"
)

// pacingLine is a combat log line at sec seconds. info is the GUID of the unit its advanced
// fields describe, standing at (x, y), or "" for a line without them.
func pacingLine(sec float64, kind, src, dst, spell, info string, x, y float64) string {
	name := func(guid string) string {
		if guid == pacingMe {
			return `"Bayov-Beta-"`
		}
		return `"Mob"`
	}
	units := []string{kind, src, name(src), "0x511", "0x0", dst, name(dst), "0xa48", "0x0"}
	if spell != "" {
		units = append(units, "8050", `"`+spell+`"`, "0x4")
	}
	if info != "" {
		level := "30"
		fields := []string{info, "0000000000000000", "900", "1366", "0", "0", "1200"}
		for len(fields) < infoX {
			fields = append(fields, "0")
		}
		fields = append(fields, fmt.Sprint(x), fmt.Sprint(y), "1441", "0.5")
		for len(fields) < infoLevel {
			fields = append(fields, "0")
		}
		units = append(units, append(fields, level)...)
		if strings.HasSuffix(kind, "_DAMAGE") {
			units = append(units, "100", "120", "-1", "4", "0", "0", "0", "nil", "nil", "nil")
		}
	}
	stamp := time.Date(2026, 10, 10, 18, 0, 0, 0, time.UTC).Add(time.Duration(sec * float64(time.Second)))
	return stamp.Format("1/2/2006 15:04:05.0000") + "  " + strings.Join(units, ",")
}

// TestPacing: two kills of level 30 mobs at our level 30. We pull the first with Flame Shock
// from 20 yd, and it swings at us 1.5 sec later. It dies at 10 sec, we eat from 11 to 14 sec,
// and walk up to the second and swing at 18 sec, so the gap less the rest is 5 sec.
func TestPacing(t *testing.T) {
	mob1, mob2 := pacingMob+"1", pacingMob+"2"
	lines := []string{
		pacingLine(0, "SPELL_CAST_SUCCESS", pacingMe, mob1, "Flame Shock", pacingMe, 0, 0),
		pacingLine(0, "SPELL_DAMAGE", pacingMe, mob1, "Flame Shock", mob1, 20, 0),
		pacingLine(1.5, "SWING_DAMAGE", mob1, pacingMe, "", mob1, 1, 0),
		pacingLine(5, "SWING_DAMAGE", pacingMe, mob1, "", pacingMe, 0, 0),
		pacingLine(10, "UNIT_DIED", "0000000000000000", mob1, "", "", 0, 0),
		pacingLine(11, "SPELL_AURA_APPLIED", pacingMe, pacingMe, "Food", "", 0, 0),
		pacingLine(14, "SPELL_AURA_REMOVED", pacingMe, pacingMe, "Food", "", 0, 0),
		pacingLine(18, "SWING_DAMAGE", pacingMe, mob2, "", pacingMe, 30, 0),
		pacingLine(19, "SWING_DAMAGE", mob2, pacingMe, "", mob2, 31, 0),
		pacingLine(25, "UNIT_DIED", "0000000000000000", mob2, "", "", 0, 0),
	}
	events, err := parseLines(strings.Join(lines, "\n"))
	if err != nil || len(events) != len(lines) {
		t.Fatalf("parsed %d of %d lines: %v", len(events), len(lines), err)
	}
	ex := &export{time: time.Date(2026, 10, 10, 17, 0, 0, 0, time.UTC)}
	ex.Who.Name, ex.Who.Realm, ex.Who.Level = "Bayov", "Beta", 30
	p := &pacing{hours: map[string]*pacingHour{}}
	p.read(events, map[string][]*export{pacingMe: {ex}}, events[0].t, events[len(events)-1].t)

	if p.fights != 2 || p.pulls["Flame Shock"] != 1 || p.pulls["our white swing"] != 1 {
		t.Errorf("%d fights, pulls %v, want 2 fights, one Flame Shock pull and one white swing", p.fights, p.pulls)
	}
	if len(p.gaps) != 1 || p.gaps[0] != 5 || p.rawGaps[0] != 8 {
		t.Errorf("gaps %v (raw %v), want 5 sec less the 3 sec of eating", p.gaps, p.rawGaps)
	}
	if d := p.pullDistance["Flame Shock"]; len(d) != 1 || d[0] != 20 {
		t.Errorf("Flame Shock pulled from %v yd, want 20", d)
	}
	if r := p.runIn["Flame Shock"]; len(r) != 1 || r[0] != 1.5 {
		t.Errorf("run-in %v sec, want 1.5", r)
	}
	if len(p.travel) != 1 || p.travel[0] != 30 {
		t.Errorf("travel %v yd, want 30", p.travel)
	}
	if p.mobsPerFight[1] != 2 {
		t.Errorf("mobs a fight %v, want two fights of one", p.mobsPerFight)
	}
}
