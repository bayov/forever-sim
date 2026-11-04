package main

import (
	"slices"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

// shamanTankLine is the priority list index of the first cast of any of spellIDs, or -1.
func shamanTankLine(rot *proto.APLRotation, spellIDs ...int32) int {
	for i, item := range rot.PriorityList {
		if id := item.GetAction().GetCastSpell().GetSpellId(); id != nil && slices.Contains(spellIDs, id.GetSpellId()) {
			return i
		}
	}
	return -1
}

// TestShamanTankShockAboveStormstrike checks the Warden template's line order
// (tank_rotations-20261104-4): by default the shock goes above Stormstrike, since Earth Shock makes
// more threat a GCD, and shockFirst=0 puts it back below.
func TestShamanTankShockAboveStormstrike(t *testing.T) {
	const stormstrikeID = 17364
	earthShockIDs := shaman.EarthShockSpellId[1:] // the rank the template picks depends on the level it's built for
	for _, c := range []struct {
		shockFirst float64
		above      bool
	}{{1, true}, {0, false}} {
		k := defaultKnobs(shamanTank{})
		if k["shockFirst"] != 1 {
			t.Fatalf("shockFirst defaults to %v, want 1", k["shockFirst"])
		}
		k["shockFirst"] = c.shockFirst
		rot := shamanTank{}.Build(k)
		ss, es := shamanTankLine(rot, stormstrikeID), shamanTankLine(rot, earthShockIDs...)
		if ss < 0 || es < 0 {
			t.Fatalf("shockFirst=%v: Stormstrike at line %d, Earth Shock at %d, want both", c.shockFirst, ss, es)
		}
		if got := es < ss; got != c.above {
			t.Errorf("shockFirst=%v: Earth Shock (line %d) above Stormstrike (line %d) is %v, want %v", c.shockFirst, es, ss, got, c.above)
		}
	}
}
