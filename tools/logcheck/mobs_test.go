package main

import (
	"strings"
	"testing"
	"time"
)

// mobLine is a combat log line from a level 30 mob with 1000 health and 1200 armor, swinging
// at our character. tail is what comes after the mob's advanced fields (the damage), or the
// miss type right after the units on a miss.
func mobLine(t *testing.T, at float64, kind, tail string) *event {
	units := "Creature-0-1-2-3-100-0001,\"Mob\",0xa48,0x0,Player-1-0001,\"Bayov-Beta-\",0x511,0x0"
	line := kind + "," + units + ","
	if kind == "SWING_DAMAGE" {
		// guid, owner, health, max health, AP, SP, armor, then the rest up to the level.
		info := []string{"Creature-0-1-2-3-100-0001", "0000000000000000", "900", "1000", "0", "0", "1200"}
		for len(info) < infoLevel {
			info = append(info, "0")
		}
		info = append(info, "30")
		line += strings.Join(info, ",") + ","
	}
	line += tail
	stamp := time.Date(2026, 10, 10, 15, 0, 0, 0, time.UTC).Add(time.Duration(at * float64(time.Second)))
	events, err := parseLines(stamp.Format("1/2/2006 15:04:05.0000") + "  " + line)
	if err != nil || len(events) != 1 {
		t.Fatalf("parsing %q: %v", line, err)
	}
	return events[0]
}

func parseLines(text string) ([]*event, error) {
	return readEvents(strings.NewReader(text))
}

func TestMobTally(t *testing.T) {
	ex := &export{time: time.Date(2026, 10, 10, 14, 0, 0, 0, time.UTC)}
	ex.Who.Name, ex.Who.Realm, ex.Who.Level = "Bayov", "Beta", 29
	byChar := map[string][]*export{"Bayov-Beta": {ex}}

	// amount, raw, overkill, school, resisted, blocked, absorbed, crit, glance, crush
	events := []*event{
		mobLine(t, 0, "SWING_DAMAGE", "40,50,0,1,0,0,0,nil,nil,nil"),
		mobLine(t, 2, "SWING_MISSED", "DODGE,nil"),
		mobLine(t, 4, "SWING_DAMAGE", "88,55,0,1,0,0,0,1,nil,nil"),
		mobLine(t, 6, "SWING_MISSED", "MISS,nil"),
	}
	m := &mobTally{levels: map[int]*mobLevel{}, diffs: map[int]*swingOutcomes{}, seen: map[string]bool{}}
	m.read(events, byChar)

	ml := m.levels[30]
	if ml == nil {
		t.Fatal("no level 30 mobs")
	}
	if len(ml.health) != 1 || ml.health[0] != 1000 || ml.armor[0] != 1200 {
		t.Errorf("health %v, armor %v, want one mob with 1000 and 1200", ml.health, ml.armor)
	}
	if median(ml.gaps) != 2 {
		t.Errorf("swing gaps %v, want 2 sec each", ml.gaps)
	}
	if len(ml.raw) != 2 || mean(ml.raw) != 52.5 {
		t.Errorf("raw damage %v, want 50 and 55", ml.raw)
	}
	d := m.diffs[1]
	if d == nil || d.swings != 4 || d.miss != 1 || d.dodge != 1 || d.crit != 1 {
		t.Errorf("at +1: %+v, want 4 swings with a miss, a dodge and a crit", d)
	}
}
