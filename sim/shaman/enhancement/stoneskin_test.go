package enhancement

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestOrcShamanStoneskinTotem checks that our own Stoneskin Totem starts when we cast it and
// that its melee damage reduction goes away with it.
//
// The buff used to be built on the first cast, which crashed the sim. It also added its -30
// a second time when it ended, where it should take it off.
func TestOrcShamanStoneskinTotem(t *testing.T) {
	sim, enh := newShamanSim(newOrcShaman(60, "", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 10)
	aura := enh.GetAura("Stoneskin")
	reduction := func() float64 { return enh.PseudoStats.BonusDamageTakenAfterModifiers[core.DefenseTypeMelee] }
	at(sim, 1, func(sim *core.Simulation) {
		if aura == nil || aura.IsActive() || reduction() != 0 {
			t.Fatalf("Stoneskin is up before we cast the totem")
		}
		castNow(t, sim, enh, topRank(t, "Stoneskin Totem", enh.StoneskinTotem))
		if !aura.IsActive() || reduction() != -30 {
			t.Errorf("after the cast: got %.0f melee damage reduction, want -30", reduction())
		}
		aura.Deactivate(sim)
		if reduction() != 0 {
			t.Errorf("after the buff ends: got %.0f melee damage reduction, want 0", reduction())
		}
	})
	runSim(sim)
}
