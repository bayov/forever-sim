package main

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"
)

func TestIsBehind(t *testing.T) {
	// The mob stands at (10, 10). Facing 0 is along +x and pi/2 along +y.
	cases := []struct {
		name   string
		facing float64
		x, y   float64
		behind bool
	}{
		{"straight in front", 0, 13, 10, false},
		{"straight behind", 0, 7, 10, true},
		{"front side, 60 degrees off", 0, 11, 10 + math.Sqrt(3), false},
		{"back side, 120 degrees off", 0, 9, 10 + math.Sqrt(3), true},
		{"behind a mob facing +y", math.Pi / 2, 10, 8, true},
		{"in front of a mob facing +y", math.Pi / 2, 10, 12, false},
		// A facing past 2 pi, or one just under it, must wrap around.
		{"facing just under 2 pi, we're in front", 2*math.Pi - 0.1, 13, 10, false},
		{"facing just under 2 pi, we're behind", 2*math.Pi - 0.1, 7, 10, true},
		{"facing 4.8950, from the Nov 22 log", 4.8950, 79.98 - 78.66 + 10, -248.40 + 247.08 + 10, false},
	}
	for _, c := range cases {
		if got := isBehind(10, 10, c.facing, c.x, c.y); got != c.behind {
			t.Errorf("%s: isBehind = %v, want %v", c.name, got, c.behind)
		}
	}
}

// unitInfo is a unit's advanced fields in a combat log line, as the client writes them.
func unitInfo(guid string, x, y, facing float64, level int) string {
	return fmt.Sprintf("%s,0000000000000000,1000,1000,100,0,1000,0,0,0,1,0,0,0,%.2f,%.2f,0,%.4f,%d", guid, x, y, facing, level)
}

func TestMeleeSides(t *testing.T) {
	const (
		me   = "Player-6940-0038ED1A"
		tank = "Player-6856-005CB67B"
		mob  = "Creature-0-6866-229-118070-9097-001182B6CA"
		mage = "Creature-0-6866-229-118070-9240-000082B6CA"
	)
	units := func(src, dst string) string {
		return fmt.Sprintf("%s,\"A\",0x511,0x80000000,%s,\"B\",0x10a48,0x80000000", src, dst)
	}
	damage := ",300,300,-1,1,0,0,0,nil,nil,nil"
	at := func(sec float64) string {
		stamp := time.Date(2026, 11, 22, 11, 43, 0, 0, time.UTC).Add(time.Duration(sec * float64(time.Second)))
		return stamp.Format("1/2/2006 15:04:05.0000") + "  "
	}
	lines := []string{
		// The mob at (10, 10) swings at the tank at (12, 10). Its logged facing (pi) points the
		// other way, as a lagging facing can, but the swing says it faces the tank.
		at(0) + "SWING_DAMAGE," + units(mob, tank) + "," + unitInfo(mob, 10, 10, math.Pi, 60) + damage,
		at(0) + "SWING_DAMAGE_LANDED," + units(mob, tank) + "," + unitInfo(tank, 12, 10, math.Pi, 60) + damage,
		// We swing from (8, 10), behind it.
		at(0.5) + "SWING_DAMAGE," + units(me, mob) + "," + unitInfo(me, 8, 10, 0, 60) + damage,
		// Stormstrike gets parried, with no advanced fields, so our position comes from the
		// swing above. Still behind.
		at(0.7) + "SPELL_MISSED," + units(me, mob) + ",17364,\"Stormstrike\",0x1,PARRY,nil,ST",
		// A Fire Nova hit isn't a melee attempt.
		at(0.8) + "SPELL_DAMAGE," + units(me, mob) + ",408428,\"Fire Nova\",0x4," + unitInfo(mob, 10, 10, math.Pi, 60) + damage + ",ST",
		// We step to (11, 12), on the mob's front side, and swing again.
		at(1.0) + "SWING_DAMAGE," + units(me, mob) + "," + unitInfo(me, 11, 12, 0, 60) + damage,
		// The caster at (20, 20) never swings, so its logged facing (+x) decides. We're at
		// (18, 20), behind it.
		at(1.2) + "SPELL_DAMAGE," + units(me, mage) + ",17364,\"Stormstrike\",0x1," + unitInfo(mage, 20, 20, 0, 59) + damage + ",ST",
		at(1.2) + "SPELL_ENERGIZE," + units(me, me) + ",408511,\"Water Shield\",0x8," + unitInfo(me, 18, 20, 0, 60) + ",93.0000,0.0000,0,4645",
		// 5 seconds on, the mob's swing at the tank is too old to go by, and we have no
		// position of ours within a second, so this one's side is unknown.
		at(5.5) + "SWING_MISSED," + units(me, mob) + ",DODGE,nil",
	}
	events, err := readEvents(strings.NewReader(strings.Join(lines, "\n")))
	if err != nil {
		t.Fatal(err)
	}
	cl := &combatLog{events: events, me: me, ours: map[string]bool{me: true}, levels: map[string]int{mob: 60, mage: 59}}
	sides := cl.meleeSides()
	if got, want := *sides[60], (meleeSide{attempts: 4, known: 3, behind: 2}); got != want {
		t.Errorf("level 60: %+v, want %+v", got, want)
	}
	if got, want := *sides[59], (meleeSide{attempts: 1, known: 1, behind: 1}); got != want {
		t.Errorf("level 59: %+v, want %+v", got, want)
	}
	if got := sides[60].share(); math.Abs(got-2.0/3) > 1e-9 {
		t.Errorf("level 60 share %.3f, want 0.667", got)
	}
	if got := (meleeSide{attempts: 5}).share(); got != 0 {
		t.Errorf("share with no positions %.3f, want 0 (in front)", got)
	}
}
