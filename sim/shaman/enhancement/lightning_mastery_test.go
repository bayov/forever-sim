package enhancement

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestOrcShamanLightningMastery checks Lightning Mastery (16166), the Elemental tree's 30-point
// talent since client 1.60.1.70425, Lightning Mastery until the November 2 development notes
// renamed it and removed Lava Burst.
//
// For 15 sec it takes 0.5 sec off Lightning Bolt's cast and takes Chain
// Lightning's cooldown away. It is off the global cooldown and comes back after 3 min. A Chain
// Lightning cast under it starts no cooldown, and once it ends the casts and the cooldown are
// back to normal.
func TestOrcShamanLightningMastery(t *testing.T) {
	// Call of Thunder and Lightning Mastery.
	sim, enh := newShamanSim(newOrcShaman(60, "0000000000001001", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 30)
	if !enh.Talents.LightningMastery {
		t.Fatalf("no Lightning Mastery talent")
	}
	mastery := enh.GetSpell(core.ActionID{SpellID: 16166})
	aura := enh.GetAura("Lightning Mastery")
	if mastery == nil || aura == nil {
		t.Fatalf("no Lightning Mastery spell or aura")
	}
	bolt := topRank(t, "Lightning Bolt", enh.LightningBolt)
	chain := topRank(t, "Chain Lightning", enh.ChainLightning)
	chainCooldown := chain.CD.Duration
	if chainCooldown <= 0 {
		t.Fatalf("Chain Lightning has no cooldown to remove")
	}

	checkCasts := func(sim *core.Simulation, label string, bolt0 time.Duration) {
		if got := bolt.CastTime(); got != bolt0 {
			t.Errorf("%s: Lightning Bolt casts in %s, want %s", label, got, bolt0)
		}
	}

	at(sim, 1, func(sim *core.Simulation) {
		checkCasts(sim, "before", 2500*time.Millisecond)
		enh.SetGCDTimer(sim, sim.CurrentTime)
		if !mastery.Cast(sim, enh.CurrentTarget) {
			t.Fatalf("Lightning Mastery didn't cast")
		}
		if !enh.GCD.IsReady(sim) {
			t.Errorf("Lightning Mastery started the global cooldown")
		}
		if readyAt, want := mastery.ReadyAt(), sim.CurrentTime+3*time.Minute; readyAt != want {
			t.Errorf("Lightning Mastery is ready again at %s, want %s", readyAt, want)
		}
		if !aura.IsActive() || aura.ExpiresAt() != sim.CurrentTime+15*time.Second {
			t.Errorf("Lightning Mastery's buff is up %v until %s, want 15 sec", aura.IsActive(), aura.ExpiresAt())
		}
		checkCasts(sim, "under Lightning Mastery", 2000*time.Millisecond)
		castNow(t, sim, enh, chain)
	})
	// The Chain Lightning above ends at 3.5 sec (a 2.5 sec cast since the November 2 notes), so
	// the next one goes right away.
	at(sim, 4, func(sim *core.Simulation) {
		if !chain.CD.IsReady(sim) {
			t.Errorf("Chain Lightning is on cooldown under Lightning Mastery, ready at %s", chain.CD.ReadyAt())
		}
		castNow(t, sim, enh, chain)
	})
	at(sim, 17, func(sim *core.Simulation) {
		if aura.IsActive() {
			t.Fatalf("Lightning Mastery's buff is still up after 16 sec")
		}
		checkCasts(sim, "after Lightning Mastery", 2500*time.Millisecond)
		castNow(t, sim, enh, chain)
	})
	// That cast ends at 19.5 sec and starts the cooldown again.
	at(sim, 20, func(sim *core.Simulation) {
		if chain.CD.IsReady(sim) {
			t.Errorf("Chain Lightning has no cooldown after Lightning Mastery, want its %s back", chainCooldown)
		}
	})
	runSim(sim)
}

// TestOrcShamanCallOfThunder checks Call of Thunder (16120), which takes Lava Burst's place under
// the November 2 development notes: 5% more crit for Lightning Bolt and Chain Lightning, nothing
// for the shocks.
func TestOrcShamanCallOfThunder(t *testing.T) {
	_, none := newShamanSim(newOrcShaman(60, "", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 2)
	_, with := newShamanSim(newOrcShaman(60, "00000000000013", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 2)
	if !with.Talents.CallOfThunder || none.Talents.CallOfThunder {
		t.Fatalf("Call of Thunder: %v with, %v without, want true and false", with.Talents.CallOfThunder, none.Talents.CallOfThunder)
	}
	for _, c := range []struct {
		name       string
		with, none *core.Spell
		want       float64
	}{
		{"Lightning Bolt", topRank(t, "Lightning Bolt", with.LightningBolt), topRank(t, "Lightning Bolt", none.LightningBolt), 5},
		{"Lightning Bolt overload", topRank(t, "Lightning Bolt overload", with.LightningBoltOverload), topRank(t, "Lightning Bolt", none.LightningBolt), 5},
		{"Chain Lightning", topRank(t, "Chain Lightning", with.ChainLightning), topRank(t, "Chain Lightning", none.ChainLightning), 5},
		{"Earth Shock", topRank(t, "Earth Shock", with.EarthShock), topRank(t, "Earth Shock", none.EarthShock), 0},
	} {
		if got := (c.with.BonusCritRating - c.none.BonusCritRating) / core.SpellCritRatingPerCritChance; got != c.want {
			t.Errorf("%s: Call of Thunder gives %.1f%% crit, want %.0f%%", c.name, got, c.want)
		}
	}
}
