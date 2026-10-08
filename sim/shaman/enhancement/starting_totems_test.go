package enhancement

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

// TestOrcShamanStartingTotems checks the totems we start the fight with at level 60.
//
// They stand at the pull without costing mana, and each goes away when its time left runs
// out. Strength of Earth has 20 sec left, Searing Totem 12 sec, and Mana Spring was just put
// down. Searing Totem attacks every 2.5 sec from the pull, so 4 times before it goes.
func TestOrcShamanStartingTotems(t *testing.T) {
	player := newOrcShaman(60, "", &proto.EnhancementShaman_Options{StartingTotems: &proto.StartingTotems{
		Earth:            proto.EarthTotem_StrengthOfEarthTotem,
		EarthSecondsLeft: 20,
		Fire:             proto.FireTotem_SearingTotem,
		FireSecondsLeft:  12,
		Water:            proto.WaterTotem_ManaSpringTotem,
	}})
	sim, enh := newShamanSim(player, &proto.Debuffs{}, 30)
	soe := enh.GetAuraByID(core.ActionID{SpellID: shaman.StrengthOfEarthTotemSpellId[5]})
	searing := enh.GetSpell(core.ActionID{SpellID: shaman.SearingTotemAttackSpellId[6]})
	if soe == nil || searing == nil {
		t.Fatalf("missing Strength of Earth rank 5 or Searing Totem rank 6")
	}
	casts := func(spell *core.Spell) int32 { return spell.SpellMetrics[enh.CurrentTarget.UnitIndex].Casts }

	at(sim, 0.001, func(sim *core.Simulation) {
		if got, want := enh.CurrentMana(), enh.MaxMana(); got != want {
			t.Errorf("at the pull: %.2f mana, want the full %.2f", got, want)
		}
		expirations := []struct {
			name string
			slot int
			want time.Duration
		}{
			{"Strength of Earth Totem", shaman.EarthTotem, 20 * time.Second},
			{"Searing Totem", shaman.FireTotem, 12 * time.Second},
			{"Mana Spring Totem", shaman.WaterTotem, 5 * time.Minute},
		}
		for _, e := range expirations {
			if got := enh.TotemExpirations[e.slot]; got != e.want {
				t.Errorf("%s runs out at %s, want %s", e.name, got, e.want)
			}
		}
		if !soe.IsActive() || soe.ExpiresAt() != 20*time.Second {
			t.Errorf("the Strength of Earth buff isn't up until 20 sec")
		}
	})

	at(sim, 13, func(sim *core.Simulation) {
		if got := casts(searing); got != 4 {
			t.Errorf("Searing Totem attacked %d times by 13 sec, want 4", got)
		}
	})
	at(sim, 20.001, func(sim *core.Simulation) {
		if soe.IsActive() {
			t.Errorf("the Strength of Earth buff is still up after 20 sec")
		}
		if got := casts(searing); got != 4 {
			t.Errorf("Searing Totem attacked %d times by 20 sec, want 4", got)
		}
	})
	runSim(sim)
}
