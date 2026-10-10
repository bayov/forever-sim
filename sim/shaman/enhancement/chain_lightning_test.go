package enhancement

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
	googleProto "google.golang.org/protobuf/proto"
)

// TestOrcShamanChainLightningJumps checks whom Chain Lightning hits, for how much, and when
// (shaman_audit.md 7.13).
//
// It hits our target and jumps to the next 2, each jump dealing 70% of the one before. So with
// 4 targets the first three take 100%, 70% and 49% of a roll and the fourth nothing. Each hit
// rolls its own damage and hit. A Stormstrike mark on the second target goes to the jump that
// hits it, which then deals 84%. Rank 4 deals 119.19 to 133.21 at level 60 with no spell power,
// so the four ranges don't overlap. The targets are at our level, and we take all our spell
// crit away.
//
// The 2 sec cast lands all three hits at its end, and the 6 sec cooldown starts there.
func TestOrcShamanChainLightningJumps(t *testing.T) {
	target := googleProto.Clone(core.NewDefaultTarget()).(*proto.Target)
	target.Level = 60
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(newOrcShaman(60, stormstrike, &proto.EnhancementShaman_Options{}), &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 20, Targets: []*proto.Target{target, target, target, target}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	enh := sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
	cl := topRank(t, "Chain Lightning", enh.ChainLightning)
	if cl.Rank != 4 {
		t.Fatalf("level 60 casts Chain Lightning rank %d, want 4", cl.Rank)
	}
	targets := sim.Encounter.TargetUnits
	mark := targets[1].GetAuraByID(enh.Stormstrike.ActionID)

	// hits returns the damage each target took and whether it was hit or missed since the last
	// call.
	type hit struct {
		damage float64
		rolled bool
	}
	var last [4]core.SpellMetrics
	hits := func() (got [4]hit) {
		for i, target := range targets {
			m := cl.SpellMetrics[target.UnitIndex]
			got[i] = hit{m.TotalDamage - last[i].TotalDamage, m.Hits+m.Crits+m.Misses > last[i].Hits+last[i].Crits+last[i].Misses}
			if m.Crits > last[i].Crits {
				t.Fatalf("Chain Lightning crit with no spell crit")
			}
			last[i] = m
		}
		return got
	}

	misses := 0
	at(sim, 1, func(sim *core.Simulation) {
		enh.AddStatDynamic(sim, stats.SpellCrit, -100*core.SpellCritRatingPerCritChance)
		for i := range 200 {
			marked := i%2 == 0
			if marked {
				mark.Activate(sim)
			}
			cl.ApplyEffects(sim, targets[0], cl)
			if mark.IsActive() {
				t.Fatalf("the jump to the marked target left the mark up")
			}
			jump := 0.7
			if marked {
				jump = 0.84
			}
			for j, h := range hits() {
				share := []float64{1, jump, 0.49, 0}[j]
				switch {
				case j == 3:
					if h.rolled {
						t.Fatalf("Chain Lightning hit the fourth target")
					}
				case !h.rolled:
					t.Fatalf("Chain Lightning didn't roll on target %d", j+1)
				case h.damage == 0:
					misses++
				case h.damage < 119.19*share-1e-6 || h.damage > 133.21*share+1e-6:
					t.Fatalf("target %d took %.2f, want %.0f%% of 119.19 to 133.21", j+1, h.damage, 100*share)
				}
			}
		}
		if misses == 0 {
			t.Errorf("none of 600 hits missed")
		}

		enh.AddMana(sim, enh.MaxMana()-enh.CurrentMana(), enh.GetManaNotCastingMetrics())
		castNow(t, sim, enh, cl)
		if readyAt, want := cl.CD.ReadyAt(), sim.CurrentTime+8*time.Second; readyAt != want {
			t.Errorf("Chain Lightning cast at %s is ready at %s, want %s", sim.CurrentTime, readyAt, want)
		}
	})
	at(sim, 2.99, func(sim *core.Simulation) {
		for j, h := range hits() {
			if h.rolled {
				t.Errorf("target %d was hit at 2.99 sec, before the cast ended", j+1)
			}
		}
	})
	at(sim, 3.01, func(sim *core.Simulation) {
		for j, h := range hits() {
			if h.rolled != (j < 3) {
				t.Errorf("at the end of the cast, target %d hit: %v, want %v", j+1, h.rolled, j < 3)
			}
		}
	})
	runSim(sim)
}
