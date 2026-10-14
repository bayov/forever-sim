package enhancement

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

// TestOrcShamanMaelstromWeaponRanks checks what each Maelstrom Weapon stack does at every rank,
// and how long the stacks last (shaman_audit.md 8.33).
//
// Each stack takes 4% a point off the next Lightning Bolt's cast time and cost, as the client
// says (408498), so 20% a stack at 5/5. The stacks last 30 sec (408505). We set 1 to 5 stacks at
// level 60 for each rank, then stop our swings, give one stack at 1 sec and want it gone between 30.9 and 31.1.
func TestOrcShamanMaelstromWeaponRanks(t *testing.T) {
	for points := 1; points <= 5; points++ {
		// Maelstrom Weapon is the seventeenth Enhancement talent.
		sim, enh := newShamanSim(newOrcShaman(60, fmt.Sprintf("-0000000000000000%d", points), &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 40)
		if enh.Talents.MaelstromWeapon != int32(points) {
			t.Fatalf("got Maelstrom Weapon %d, want %d", enh.Talents.MaelstromWeapon, points)
		}
		lb := topRank(t, "Lightning Bolt", enh.LightningBolt)
		cost := shaman.LightningBoltManaCost[lb.Rank]
		mw := enh.MaelstromWeaponAura

		at(sim, 0.5, func(sim *core.Simulation) {
			for stacks := 1; stacks <= 5; stacks++ {
				mw.Activate(sim)
				mw.SetStacks(sim, int32(stacks))
				share := max(0, 1-.04*float64(points*stacks))
				if got := lb.Cost.GetCurrentCost(); math.Abs(got-cost*share) > 1e-9 {
					t.Errorf("Maelstrom Weapon %d with %d stacks: Lightning Bolt costs %.2f, want %.2f", points, stacks, got, cost*share)
				}
				want := time.Duration(float64(lb.DefaultCast.CastTime) * share)
				if got := lb.CastTime(); (got - want).Abs() > time.Microsecond {
					t.Errorf("Maelstrom Weapon %d with %d stacks: Lightning Bolt casts in %s, want %s", points, stacks, got, want)
				}
			}
			mw.Deactivate(sim)
		})
		// We stop our swings so they don't give more stacks.
		at(sim, 1, func(sim *core.Simulation) {
			enh.AutoAttacks.CancelAutoSwing(sim)
			mw.Activate(sim)
			mw.AddStack(sim)
		})
		at(sim, 30.9, func(sim *core.Simulation) {
			if !mw.IsActive() {
				t.Errorf("Maelstrom Weapon %d: the stack ran out before 30 sec", points)
			}
		})
		at(sim, 31.1, func(sim *core.Simulation) {
			if mw.IsActive() {
				t.Errorf("Maelstrom Weapon %d: the stack lasted past 30 sec", points)
			}
		})
		runSim(sim)
	}
}
