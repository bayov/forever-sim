package enhancement

import (
	"fmt"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestOrcShamanElementalAlacrity checks how much Elemental Alacrity takes off our cast times
// (shaman_audit.md 8.8).
//
// The client takes 0.17, 0.33 and 0.5 sec off Lightning Bolt, Chain Lightning and Lava Burst
// for 1, 2 and 3 points, on every rank. The cut comes before haste, so with 3 points at 25%
// haste a top rank Lightning Bolt casts in (2.5 - 0.5) / 1.25 = 1.6 sec.
func TestOrcShamanElementalAlacrity(t *testing.T) {
	for points, cut := range []time.Duration{0, 170, 330, 500} {
		cut *= time.Millisecond
		// Elemental Alacrity points and Lava Burst.
		sim, enh := newShamanSim(newOrcShaman(60, fmt.Sprintf("0000000%d00000001", points), &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 10)
		if enh.Talents.ElementalAlacrity != int32(points) {
			t.Fatalf("got Elemental Alacrity %d, want %d", enh.Talents.ElementalAlacrity, points)
		}
		if enh.LightningBolt[1] == nil {
			t.Fatalf("no Lightning Bolt rank 1")
		}
		spells := []struct {
			name  string
			spell *core.Spell
			base  time.Duration
		}{
			{"Lightning Bolt", topRank(t, "Lightning Bolt", enh.LightningBolt), 2500 * time.Millisecond},
			{"Lightning Bolt rank 1", enh.LightningBolt[1], 1500 * time.Millisecond},
			{"Chain Lightning", topRank(t, "Chain Lightning", enh.ChainLightning), 2000 * time.Millisecond},
			{"Lava Burst", enh.LavaBurst, 2500 * time.Millisecond},
		}
		check := func(sim *core.Simulation, label string, haste float64) {
			for _, s := range spells {
				want := time.Duration(float64(s.base-cut) / haste)
				if got := s.spell.CastTime(); (got - want).Abs() > time.Microsecond {
					t.Errorf("Elemental Alacrity %d%s: %s casts in %s, want %s", points, label, s.name, got, want)
				}
			}
		}
		at(sim, 1, func(sim *core.Simulation) {
			check(sim, "", 1)
			enh.MultiplyCastSpeed(1.25)
			check(sim, " at 25% haste", 1.25)
		})
		runSim(sim)
	}
}
