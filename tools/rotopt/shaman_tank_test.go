package main

import (
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
)

// shamanTankLine is the priority list index of the first cast of spellID, or -1.
func shamanTankLine(rot *proto.APLRotation, spellID int32) int {
	for i, item := range rot.PriorityList {
		if id := item.GetAction().GetCastSpell().GetSpellId(); id != nil && id.GetSpellId() == spellID {
			return i
		}
	}
	return -1
}

// TestShamanTankShockAboveStormstrike checks the Warden template's line order
// (tank_rotations-20261104-4): by default the shock goes above Stormstrike, since Earth Shock makes
// more threat a GCD, and shockFirst=0 puts it back below.
func TestShamanTankShockAboveStormstrike(t *testing.T) {
	const stormstrikeID, earthShockID = 17364, 10414
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
		ss, es := shamanTankLine(rot, stormstrikeID), shamanTankLine(rot, earthShockID)
		if ss < 0 || es < 0 {
			t.Fatalf("shockFirst=%v: Stormstrike at line %d, Earth Shock at %d, want both", c.shockFirst, ss, es)
		}
		if got := es < ss; got != c.above {
			t.Errorf("shockFirst=%v: Earth Shock (line %d) above Stormstrike (line %d) is %v, want %v", c.shockFirst, es, ss, got, c.above)
		}
	}
}
