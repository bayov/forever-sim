package sim

// The Restoration shaman's healing talents on a healing fight (docs/healing_shaman.md). Needs the item
// database: go test -tags with_db.

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

// The bench's talents with Healing Focus's 3 points in Ancestral Healing.
const restoAncestralHealingTalents = "-035-5530503015513151"

// Ancestral Healing 3: a crit from the shaman's heal on the tank gives the tank 25% more of all its
// armor (Ancestral Fortitude, 16177, aura 101 as Inspiration), its Agility's and buffs' too, not
// only its armor from items (healer_mechanics-20261104-3), and it goes again 15 sec later.
func TestHealingFightShamanAncestralFortitude(t *testing.T) {
	rsr := healingRaid(t, 1, 1, 0)
	rsr.Raid.Parties[0].Players[1].TalentsString = restoAncestralHealingTalents
	rsr.Encounter.Duration = 120
	sim := core.NewSim(rsr, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	tank := sim.Raid.Parties[0].Players[0].GetCharacter()
	buff := tank.GetAura("Ancestral Fortitude")
	if buff == nil {
		t.Fatal("no Ancestral Fortitude on the tank")
	}
	armor, itemArmor := tank.Armor(), tank.EquipStats()[stats.Armor]
	if itemArmor <= 0 || itemArmor >= armor {
		t.Fatalf("the tank has %.0f armor, %.0f from items", armor, itemArmor)
	}
	gained := false
	for !sim.Step() {
		if buff.IsActive() && !gained {
			gained = true
			// Its own buffs may change its armor by now (Shield Block doesn't), so read the
			// buff's part: all armor, 25% more.
			if got, want := tank.Armor()-armor, 0.25*armor; math.Abs(got-want) > 1 {
				t.Errorf("Ancestral Fortitude added %.0f armor, want %.0f (25%% of all %.0f; from items only would be %.0f)",
					got, want, armor, 0.25*itemArmor)
			}
		}
		if gained && !buff.IsActive() {
			if got := tank.Armor(); math.Abs(got-armor) > 1 {
				t.Errorf("armor %.0f after the buff, want %.0f", got, armor)
			}
			return
		}
	}
	t.Fatal("no heal crit on the tank gave Ancestral Fortitude, or it never ended")
}
