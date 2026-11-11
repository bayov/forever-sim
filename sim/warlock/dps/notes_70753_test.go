package dps

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

// The Warlock lines of Blizzard's development notes of 2026-11-02 (client 1.60.1.70753).

// Wrack's damage is checked at hotfix 113766's numbers in notes_70888_test.go.

// TestWarlockShadowAndFlame70753 checks Shadow and Flame 5/5: a Conflagrate that lands gives 20%
// Shadow damage (4% a point, was 2%).
func TestWarlockShadowAndFlame70753(t *testing.T) {
	sim, lock := newWarlockSim(t, "--000000000010005", &proto.WarlockOptions{})
	shadow := func() float64 { return lock.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexShadow] }
	var before float64

	at(sim, 0.5, func(sim *core.Simulation) {
		lock.AddStatDynamic(sim, stats.SpellHit, 20*core.SpellHitRatingPerHitChance)
		castNow(t, sim, lock, last(lock.Immolate))
	})
	at(sim, 3, func(sim *core.Simulation) {
		before = shadow()
		lock.SetGCDTimer(sim, sim.CurrentTime)
		if !last(lock.Conflagrate).Cast(sim, lock.CurrentTarget) {
			t.Fatalf("Conflagrate didn't cast")
		}
	})
	at(sim, 3.5, func(sim *core.Simulation) {
		if got := shadow() / before; math.Abs(got-1.20) > 1e-9 {
			t.Errorf("Shadow damage after Conflagrate with Shadow and Flame 5/5: x%.4f, want x1.20", got)
		}
	})
	runSim(sim)
}

// TestWarlockBaneOfAgonyLive checks that Bane of Agony's ticks read spell damage when they land
// ("now correctly dynamically updates its damage"): Shadow damage gained after the cast raises
// the next tick by the rank's coefficient.
func TestWarlockBaneOfAgonyLive(t *testing.T) {
	sim, lock := newWarlockSim(t, "", &proto.WarlockOptions{})
	target := lock.CurrentTarget
	agony := last(lock.BaneOfAgony)
	tick := func() float64 {
		return agony.Dot(target).CalcSnapshotDamage(sim, target, agony.OutcomeExpectedMagicAlwaysHit).Damage
	}
	var before float64

	at(sim, 0.5, func(sim *core.Simulation) {
		lock.AddStatDynamic(sim, stats.SpellHit, 20*core.SpellHitRatingPerHitChance)
		castNow(t, sim, lock, agony)
	})
	at(sim, 1, func(sim *core.Simulation) {
		before = tick()
		lock.AddStatDynamic(sim, stats.ShadowPower, 300)
	})
	at(sim, 1.5, func(sim *core.Simulation) {
		after := tick()
		if after <= before {
			t.Errorf("Bane of Agony's tick %.1f before +300 Shadow damage and %.1f after, want it to grow", before, after)
		}
		if got, want := after-before, 300*agony.Dot(target).BonusCoefficient*before/agony.Dot(target).SnapshotBaseDamage; math.Abs(got-want) > 1e-6 {
			t.Errorf("Bane of Agony's tick grew %.3f with +300 Shadow damage, want %.3f", got, want)
		}
	})
	runSim(sim)
}

// TestWarlockAmplifyCurseAgony checks that Amplify Curse raises Bane of Agony's base damage by 50%
// ("now correctly affects Curse of Exhaustion and Bane of Agony", the 2026-11-02 notes), not its
// spell damage part, and is spent by it.
func TestWarlockAmplifyCurseAgony(t *testing.T) {
	bases := map[bool]float64{}
	for _, amplify := range []bool{false, true} {
		sim, lock := newWarlockSim(t, "000000001", &proto.WarlockOptions{})
		target := lock.CurrentTarget
		agony := last(lock.BaneOfAgony)

		at(sim, 0.5, func(sim *core.Simulation) {
			lock.AddStatDynamic(sim, stats.SpellHit, 20*core.SpellHitRatingPerHitChance)
			if amplify {
				lock.AmplifyCurse.Cast(sim, target)
			}
			castNow(t, sim, lock, agony)
		})
		at(sim, 1, func(sim *core.Simulation) {
			dot := agony.Dot(target)
			bases[amplify] = dot.SnapshotBaseDamage - dot.BonusCoefficient*agony.GetBonusDamage(target)
			if lock.AmplifyCurseAura.IsActive() {
				t.Errorf("Amplify Curse still up after Bane of Agony")
			}
		})
		runSim(sim)
	}
	if got := bases[true] / bases[false]; math.Abs(got-1.5) > 1e-9 {
		t.Errorf("Amplify Curse makes Bane of Agony's base x%.4f, want x1.5", got)
	}
}

// TestWarlockOrcCommand checks the Orc's Command (the 2026-11-02 notes: "re-added, increasing
// Warlock and Hunter pet damage by 5%") on the warlock's demons.
func TestWarlockOrcCommand(t *testing.T) {
	multipliers := map[proto.Race]float64{}
	for _, race := range []proto.Race{proto.Race_RaceOrc, proto.Race_RaceHuman} {
		player := &proto.Player{
			Race:      race,
			Class:     proto.Class_ClassWarlock,
			Level:     60,
			Equipment: &proto.EquipmentSpec{},
			Rotation:  &proto.APLRotation{},
			Spec:      &proto.Player_Warlock{Warlock: &proto.Warlock{Options: &proto.WarlockOptions{Summon: proto.WarlockOptions_Succubus}}},
		}
		sim := core.NewSim(&proto.RaidSimRequest{
			Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
			SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
		}, simsignals.CreateSignals())
		sim.Reset()
		lock := sim.Raid.Parties[0].Players[0].(*DpsWarlock)
		multipliers[race] = lock.Succubus.PseudoStats.DamageDealtMultiplier
	}
	if got := multipliers[proto.Race_RaceOrc] / multipliers[proto.Race_RaceHuman]; math.Abs(got-1.05) > 1e-9 {
		t.Errorf("Orc Succubus damage x%.4f of the Human's, want x1.05", got)
	}
}
