package retribution

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestUndeadOnlySpells checks the paladin's creature-type spells against the two bosses the
// presets use (standards 19): on the generic boss (type Unknown) Exorcism cannot be cast and
// Holy Wrath hits nothing, and against an Undead boss both land.
func TestUndeadOnlySpells(t *testing.T) {
	for _, c := range []struct {
		mobType proto.MobType
		lands   bool
	}{
		{proto.MobType_MobTypeUnknown, false},
		{proto.MobType_MobTypeUndead, true},
	} {
		sim, ret := retSim(t, "55003003201121251--05225230001", arcaniteReaper)
		target := ret.CurrentTarget
		target.MobType = c.mobType

		exorcism := ret.GetSpell(core.ActionID{SpellID: 10314})
		holyWrath := ret.GetSpell(core.ActionID{SpellID: 10318})
		if exorcism == nil || holyWrath == nil {
			t.Fatalf("Exorcism (10314) or Holy Wrath (10318) is not registered")
		}
		if got := exorcism.CanCast(sim, target); got != c.lands {
			t.Errorf("%v: Exorcism castable = %v, want %v", c.mobType, got, c.lands)
		}
		metrics := &holyWrath.SpellMetrics[target.UnitIndex]
		for range 10 {
			holyWrath.ApplyEffects(sim, target, holyWrath)
		}
		if got := metrics.Hits+metrics.Crits+metrics.Misses > 0; got != c.lands {
			t.Errorf("%v: Holy Wrath landed = %v, want %v", c.mobType, got, c.lands)
		}
	}
}
