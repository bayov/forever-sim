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
