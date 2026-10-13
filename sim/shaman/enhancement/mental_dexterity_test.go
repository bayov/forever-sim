package enhancement

import (
	"fmt"
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestOrcShamanMentalDexterity checks the attack power Mental Dexterity gives (shaman_audit.md
// 8.21).
//
// 1, 2 and 3 points give attack power equal to 33, 67 and 100% of our Intellect, as the client
// says (415140). It follows Intellect we gain mid-fight. On the beta 3 more Intellect gave 3
// more attack power at 3/3.
func TestOrcShamanMentalDexterity(t *testing.T) {
	ap := func(points int) (*core.Simulation, *EnhancementShaman) {
		// Mental Dexterity is the fifth Enhancement talent.
		return newShamanSim(newOrcShaman(30, fmt.Sprintf("-0000%d", points), &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 2)
	}
	_, none := ap(0)
	for points, share := range []float64{0, .33, .67, 1} {
		sim, enh := ap(points)
		if enh.Talents.MentalDexterity != int32(points) {
			t.Fatalf("got Mental Dexterity %d, want %d", enh.Talents.MentalDexterity, points)
		}
		intellect := enh.GetStat(stats.Intellect)
		if got, want := enh.GetStat(stats.AttackPower)-none.GetStat(stats.AttackPower), intellect*share; math.Abs(got-want) > 1e-9 {
			t.Errorf("Mental Dexterity %d: %.2f more attack power from %.0f Intellect, want %.2f", points, got, intellect, want)
		}
		if points != 3 {
			continue
		}
		at(sim, 1, func(sim *core.Simulation) {
			before := enh.GetStat(stats.AttackPower)
			enh.AddStatDynamic(sim, stats.Intellect, 10)
			if got := enh.GetStat(stats.AttackPower) - before; math.Abs(got-10) > 1e-9 {
				t.Errorf("10 Intellect mid-fight gave %.2f attack power at 3/3, want 10", got)
			}
		})
		runSim(sim)
	}
}
