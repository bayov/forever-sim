package enhancement

import (
	"fmt"
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestOrcShamanMentalQuickness checks the spell power Mental Quickness gives (shaman_audit.md
// 8.31).
//
// 1 and 2 points give spell power equal to 15 and 30% of our Intellect, as the client says
// (30812). It follows Intellect we gain mid-fight. On the beta dropping 2/2 took 42 spell
// power off the sheet, which is 30% of our 141 Intellect.
func TestOrcShamanMentalQuickness(t *testing.T) {
	sp := func(points int) (*core.Simulation, *EnhancementShaman) {
		// Mental Quickness is the fifteenth Enhancement talent.
		return newShamanSim(newOrcShaman(30, fmt.Sprintf("-00000000000000%d", points), &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 2)
	}
	_, none := sp(0)
	for points, share := range []float64{0, .15, .30} {
		sim, enh := sp(points)
		if enh.Talents.MentalQuickness != int32(points) {
			t.Fatalf("got Mental Quickness %d, want %d", enh.Talents.MentalQuickness, points)
		}
		intellect := enh.GetStat(stats.Intellect)
		if got, want := enh.GetStat(stats.SpellPower)-none.GetStat(stats.SpellPower), intellect*share; math.Abs(got-want) > 1e-9 {
			t.Errorf("Mental Quickness %d: %.2f more spell power from %.0f Intellect, want %.2f", points, got, intellect, want)
		}
		if points != 2 {
			continue
		}
		at(sim, 1, func(sim *core.Simulation) {
			before := enh.GetStat(stats.SpellPower)
			enh.AddStatDynamic(sim, stats.Intellect, 10)
			if got := enh.GetStat(stats.SpellPower) - before; math.Abs(got-3) > 1e-9 {
				t.Errorf("10 Intellect mid-fight gave %.2f spell power at 2/2, want 3", got)
			}
		})
		runSim(sim)
	}
}
