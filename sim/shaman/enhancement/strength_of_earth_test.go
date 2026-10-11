package enhancement

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestOrcShamanStrengthOfEarthRanks checks the Strength our own Strength of Earth Totem gives
// by rank (shaman_audit.md 7.21).
//
// Ranks 1 to 5 come at levels 10, 24, 38, 52 and 60 and give 7, 14, 25, 42 and 53 Strength, as
// the Forever client's party auras do. At level 30 rank 2's 14 Strength is 28 attack power. On
// the beta our attack power went from 869 to 897 when we put it down. The Strength goes when a
// Tremor Totem takes the earth slot.
func TestOrcShamanStrengthOfEarthRanks(t *testing.T) {
	for _, c := range []struct {
		level    int32
		strength float64
	}{{10, 7}, {24, 14}, {30, 14}, {38, 25}, {52, 42}, {60, 53}} {
		sim, enh := newShamanSim(newOrcShaman(c.level, "", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 5)
		at(sim, 1, func(sim *core.Simulation) {
			strength, ap := enh.GetStat(stats.Strength), enh.GetStat(stats.AttackPower)
			castNow(t, sim, enh, topRank(t, "Strength of Earth Totem", enh.StrengthOfEarthTotem))
			if got := enh.GetStat(stats.Strength) - strength; got != c.strength {
				t.Errorf("level %d: Strength of Earth Totem gave %.0f Strength, want %.0f", c.level, got, c.strength)
			}
			if got := enh.GetStat(stats.AttackPower) - ap; c.level == 30 && got != 28 {
				t.Errorf("level 30: Strength of Earth Totem gave %.0f attack power, want 28 as on the beta", got)
			}
			if c.level >= 18 {
				castNow(t, sim, enh, enh.TremorTotem)
				if got := enh.GetStat(stats.Strength) - strength; got != 0 {
					t.Errorf("level %d: %.0f Strength left after Tremor Totem took the earth slot", c.level, got)
				}
			}
		})
		runSim(sim)
	}
}
