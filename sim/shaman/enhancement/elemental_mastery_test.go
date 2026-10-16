package enhancement

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestOrcShamanElementalMastery checks Elemental Mastery, the Elemental tree's 30-point talent
// since client 1.60.1.70425 (16166).
//
// For 15 sec it takes 0.5 sec off Lightning Bolt's and Lava Burst's casts and takes Chain
// Lightning's cooldown away. It is off the global cooldown and comes back after 3 min. A Chain
// Lightning cast under it starts no cooldown, and once it ends the casts and the cooldown are
// back to normal.
func TestOrcShamanElementalMastery(t *testing.T) {
	// Lava Burst and Elemental Mastery.
	sim, enh := newShamanSim(newOrcShaman(60, "0000000000001001", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 30)
	if !enh.Talents.ElementalMastery {
		t.Fatalf("no Elemental Mastery talent")
	}
	mastery := enh.GetSpell(core.ActionID{SpellID: 16166})
	aura := enh.GetAura("Elemental Mastery")
	if mastery == nil || aura == nil {
		t.Fatalf("no Elemental Mastery spell or aura")
	}
	bolt := topRank(t, "Lightning Bolt", enh.LightningBolt)
	chain := topRank(t, "Chain Lightning", enh.ChainLightning)
	lava := enh.LavaBurst
	chainCooldown := chain.CD.Duration
	if chainCooldown <= 0 {
		t.Fatalf("Chain Lightning has no cooldown to remove")
	}

	checkCasts := func(sim *core.Simulation, label string, bolt0, lava0 time.Duration) {
		if got := bolt.CastTime(); got != bolt0 {
			t.Errorf("%s: Lightning Bolt casts in %s, want %s", label, got, bolt0)
		}
		if got := lava.CastTime(); got != lava0 {
			t.Errorf("%s: Lava Burst casts in %s, want %s", label, got, lava0)
		}
	}

	at(sim, 1, func(sim *core.Simulation) {
		checkCasts(sim, "before", 2500*time.Millisecond, 2500*time.Millisecond)
		enh.SetGCDTimer(sim, sim.CurrentTime)
		if !mastery.Cast(sim, enh.CurrentTarget) {
			t.Fatalf("Elemental Mastery didn't cast")
		}
		if !enh.GCD.IsReady(sim) {
			t.Errorf("Elemental Mastery started the global cooldown")
		}
		if readyAt, want := mastery.ReadyAt(), sim.CurrentTime+3*time.Minute; readyAt != want {
			t.Errorf("Elemental Mastery is ready again at %s, want %s", readyAt, want)
		}
		if !aura.IsActive() || aura.ExpiresAt() != sim.CurrentTime+15*time.Second {
			t.Errorf("Elemental Mastery's buff is up %v until %s, want 15 sec", aura.IsActive(), aura.ExpiresAt())
		}
		checkCasts(sim, "under Elemental Mastery", 2000*time.Millisecond, 2000*time.Millisecond)
		castNow(t, sim, enh, chain)
	})
	// The Chain Lightning above ends at 3 sec, so the next one goes right away.
	at(sim, 3.5, func(sim *core.Simulation) {
		if !chain.CD.IsReady(sim) {
			t.Errorf("Chain Lightning is on cooldown under Elemental Mastery, ready at %s", chain.CD.ReadyAt())
		}
		castNow(t, sim, enh, chain)
	})
	at(sim, 17, func(sim *core.Simulation) {
		if aura.IsActive() {
			t.Fatalf("Elemental Mastery's buff is still up after 16 sec")
		}
		checkCasts(sim, "after Elemental Mastery", 2500*time.Millisecond, 2500*time.Millisecond)
		castNow(t, sim, enh, chain)
	})
	// That cast ends at 19 sec and starts the cooldown again.
	at(sim, 19.5, func(sim *core.Simulation) {
		if chain.CD.IsReady(sim) {
			t.Errorf("Chain Lightning has no cooldown after Elemental Mastery, want its %s back", chainCooldown)
		}
	})
	runSim(sim)
}
