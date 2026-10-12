package retribution

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestHolyPowerOnItsSpellsOnly checks that Holy Power (client 1.60.1.70338, spell 5923) is a
// crit modifier on the paladin's own spells and not a spell crit stat, which would also reach
// Dragonbreath Chili, Shadow Oil and the other consumables. At 5 points Holy Shock gets 15%
// and Judgement of Righteousness 5%.
func TestHolyPowerOnItsSpellsOnly(t *testing.T) {
	const withHolyPower = "55003003201121251--05225230001"
	const withoutHolyPower = "55003003201121201--05225230001"

	_, with := retSim(t, withHolyPower, arcaniteReaper)
	_, without := retSim(t, withoutHolyPower, arcaniteReaper)

	if got, want := with.GetStat(stats.SpellCrit), without.GetStat(stats.SpellCrit); got != want {
		t.Errorf("spell crit stat with Holy Power = %.2f, want %.2f (no Holy Power)", got, want)
	}

	checks := []struct {
		name    string
		spellID int32
		want    float64
	}{
		{"Holy Shock", 20930, 15},
		{"Judgement of Righteousness", 20286, 5},
	}
	for _, c := range checks {
		spell := with.GetSpell(core.ActionID{SpellID: c.spellID})
		if spell == nil {
			t.Fatalf("%s (%d) is not registered", c.name, c.spellID)
		}
		if spell.BonusCritRating != c.want {
			t.Errorf("%s bonus crit = %.1f, want %.1f", c.name, spell.BonusCritRating, c.want)
		}
	}
}

// TestConsecratedGroundOnConsecration checks that Consecrated Ground (client 1310905, 5% a
// point to Holy spells against targets in the Consecration) reaches Consecration's own ticks:
// at 2 points the ticks snapshot 10% more than without the talent.
func TestConsecratedGroundOnConsecration(t *testing.T) {
	snapshot := func(talents string) float64 {
		sim, ret := retSim(t, talents, arcaniteReaper)
		castByID(sim, ret, 20924)
		return ret.GetSpell(core.ActionID{SpellID: 20924}).AOEDot().SnapshotAttackerMultiplier
	}
	with := snapshot("55003003201121251--05225230001")
	without := snapshot("55003003201121051--05225230001")
	if ratio := with / without; ratio < 1.0999 || ratio > 1.1001 {
		t.Errorf("Consecration tick multiplier with Consecrated Ground 2 / without = %.4f, want 1.1000", ratio)
	}
}
