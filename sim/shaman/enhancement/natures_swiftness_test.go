package enhancement

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestOrcShamanNaturesSwiftness checks that Nature's Swiftness makes one Lightning Bolt instant
// and that its cooldown starts when the buff is used up (shaman_audit.md 8.48).
//
// The next Nature spell with a cast time is instant and uses up the buff. The client (70338)
// flags the 3 min cooldown to start when the buff ends. We press it at 1 sec on a level 60
// shaman, cast Lightning Bolts at 3 and 7 sec and want it ready at 183 sec, not 181.
func TestOrcShamanNaturesSwiftness(t *testing.T) {
	// Nature's Swiftness is the fourteenth Restoration talent.
	sim, enh := newShamanSim(newOrcShaman(60, "--00000000000001", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 200)
	ns := enh.GetSpell(core.ActionID{SpellID: 16188})
	if ns == nil {
		t.Fatal("no Nature's Swiftness")
	}
	buff := enh.GetAura("Natures Swiftness")
	lb := topRank(t, "Lightning Bolt", enh.LightningBolt)
	full := lb.DefaultCast.CastTime

	at(sim, 1, func(sim *core.Simulation) {
		if !ns.Cast(sim, enh.CurrentTarget) {
			t.Fatal("Nature's Swiftness didn't cast")
		}
	})
	at(sim, 3, func(sim *core.Simulation) {
		if got := lb.CastTime(); got != 0 {
			t.Errorf("Lightning Bolt casts in %s under Nature's Swiftness, want instant", got)
		}
		castNow(t, sim, enh, lb)
		if buff.IsActive() {
			t.Error("the instant Lightning Bolt didn't use up Nature's Swiftness")
		}
	})
	at(sim, 7, func(sim *core.Simulation) {
		if got := lb.CastTime(); got != full {
			t.Errorf("the next Lightning Bolt casts in %s, want %s", got, full)
		}
		castNow(t, sim, enh, lb)
	})
	at(sim, 2, func(sim *core.Simulation) {
		if ns.IsReady(sim) {
			t.Error("Nature's Swiftness is ready again while its buff is up")
		}
	})
	at(sim, 182.9, func(sim *core.Simulation) {
		if ns.IsReady(sim) {
			t.Error("Nature's Swiftness is ready before 3 min after the buff was used")
		}
	})
	at(sim, 183.1, func(sim *core.Simulation) {
		if !ns.IsReady(sim) {
			t.Errorf("Nature's Swiftness isn't ready 3 min after the buff was used (ready in %s)", ns.CD.TimeToReady(sim))
		}
	})
	runSim(sim)
}
