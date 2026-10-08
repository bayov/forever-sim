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
// They stand at the pull without costing mana, and each loses the time since we put it down.
// Strength of Earth went down 280 sec before the pull, so it has 20 sec left. Searing Totem
// (55 sec) went down 43 sec before, so it has 12 sec left and attacks every 2.5 sec from the
// pull, 4 times. Mana Spring went down at the pull. Grace of Air went down 6 min before, so
// it's gone.
func TestOrcShamanStartingTotems(t *testing.T) {
	player := newOrcShaman(60, "", &proto.EnhancementShaman_Options{StartingTotems: &proto.StartingTotems{
		Earth:                  proto.EarthTotem_StrengthOfEarthTotem,
		EarthSecondsBeforePull: 280,
		Air:                    proto.AirTotem_GraceOfAirTotem,
		AirSecondsBeforePull:   360,
		Fire:                   proto.FireTotem_SearingTotem,
		FireSecondsBeforePull:  43,
		Water:                  proto.WaterTotem_ManaSpringTotem,
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
			// The fire totem gets 1 ns on top, so its last attack lands before it goes.
			{"Searing Totem", shaman.FireTotem, 12*time.Second + 1},
			{"Grace of Air Totem", shaman.AirTotem, 0},
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
		if goa := enh.GetAura("Grace of Air Totem"); goa == nil || goa.IsActive() {
			t.Errorf("the Grace of Air buff is up, but the totem ran out before the pull")
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
