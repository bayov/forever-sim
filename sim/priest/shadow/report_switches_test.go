package shadow

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestDashRacialsOffPriestSpells: DASH_RACIALS_OFF leaves the priest racial spells out, and
// DASH_RACIALS_OFF_CLASS keeps it to the class it names. Unset, every race has its spell.
func TestDashRacialsOffPriestSpells(t *testing.T) {
	const talents = "00503523103--50052101231302051"
	learned := func(race proto.Race) bool {
		_, sp := newShadowPriestSimRace(t, talents, race)
		var spells []*core.Spell
		switch race {
		case proto.Race_RaceDwarf:
			spells = sp.Chastise
		case proto.Race_RaceUndead:
			spells = []*core.Spell{sp.DarkSacrifice}
		case proto.Race_RaceNightElf:
			spells = sp.Starshards
		case proto.Race_RaceTroll:
			spells = sp.Shadowguard
		}
		for _, s := range spells {
			if s != nil {
				return true
			}
		}
		return false
	}
	races := map[proto.Race]string{
		proto.Race_RaceDwarf:    "chastise",
		proto.Race_RaceUndead:   "darksacrifice",
		proto.Race_RaceNightElf: "starshards",
		proto.Race_RaceTroll:    "shadowguard",
	}
	for _, c := range []struct {
		off, class string
		want       func(part string) bool // whether the race's spell stays
	}{
		{"", "", func(string) bool { return true }},
		{"all", "", func(string) bool { return false }},
		{"starshards,chastise", "", func(p string) bool { return p != "starshards" && p != "chastise" }},
		{"all", "Priest", func(string) bool { return false }},
		{"all", "Warrior", func(string) bool { return true }},
	} {
		t.Run(c.off+"/"+c.class, func(t *testing.T) {
			t.Cleanup(core.ReloadReportSwitches)
			t.Setenv("DASH_RACIALS_OFF", c.off)
			t.Setenv("DASH_RACIALS_OFF_CLASS", c.class)
			core.ReloadReportSwitches()
			for race, part := range races {
				if got := learned(race); got != c.want(part) {
					t.Errorf("%s priest: %s learned %v, want %v", race, part, got, c.want(part))
				}
			}
		})
	}
}
