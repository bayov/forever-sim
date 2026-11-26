package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
)

// TestOrcShamanOverloadRollsElementalFocus checks that a Lightning Overload's launch rolls
// Elemental Focus once, as a cast does (EV-0137: 31 of 248 overloaded bolts gained Clearcasting at
// the impact, 1 of 565 without an overload).
//
// Elemental Focus and Lightning Overload 3/3. Each bolt is cast with Clearcasting down. A bolt
// without an overload gains it 10% of the time, from its own cast; one with an overload 19%
// (1 - 0.9 x 0.9). An overload launched with Clearcasting up keeps the charge.
func TestOrcShamanOverloadRollsElementalFocus(t *testing.T) {
	const casts = 4000
	sim, enh := newShamanSim(newOrcShaman(60, "00000010000003", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 4*casts+10)
	target := enh.CurrentTarget
	lb := topRank(t, "Lightning Bolt", enh.LightningBolt)
	overload := enh.LightningBoltOverload[lb.Rank]
	cc := enh.ClearcastingAura
	if overload == nil || cc == nil {
		t.Fatalf("no overload or Clearcasting")
	}

	at(sim, 0.5, func(sim *core.Simulation) {
		cc.Activate(sim)
		cc.SetStacks(sim, 1)
		overload.Cast(sim, target)
		if !cc.IsActive() {
			t.Errorf("an overload launch spent Clearcasting")
		}
		cc.Deactivate(sim)
	})

	var gains, launches [2]int // indexed by whether the bolt overloaded
	for i := range casts {
		var before int32
		at(sim, float64(1+4*i), func(sim *core.Simulation) {
			enh.AddMana(sim, enh.MaxMana(), enh.GetManaNotCastingMetrics())
			cc.Deactivate(sim)
			before = overload.SpellMetrics()[target.UnitIndex].Casts
			castNow(t, sim, enh, lb)
		})
		at(sim, float64(1+4*i)+3.5, func(sim *core.Simulation) {
			overloaded := 0
			if overload.SpellMetrics()[target.UnitIndex].Casts > before {
				overloaded = 1
			}
			launches[overloaded]++
			if cc.IsActive() {
				gains[overloaded]++
			}
		})
	}
	runSim(sim)

	for overloaded, want := range []float64{.10, .19} {
		rate := float64(gains[overloaded]) / float64(launches[overloaded])
		if math.Abs(rate-want) > .035 {
			t.Errorf("overloaded %t: %d of %d bolts gained Clearcasting (%.1f%%), want %.0f%%",
				overloaded == 1, gains[overloaded], launches[overloaded], rate*100, want*100)
		}
	}
}

// TestOrcShamanChainLightningOverloadPerTarget checks that each target Chain Lightning lands on
// rolls its own overload at a third of Lightning Bolt's chance, 10% at 3/3, and the overload
// starts from that target (EV-0137: 9.3% [7.6, 11.2] on every landed target over two logs).
func TestOrcShamanChainLightningOverloadPerTarget(t *testing.T) {
	const casts = 3000
	player := newOrcShaman(60, "00000000000003", &proto.EnhancementShaman_Options{})
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter: &proto.Encounter{Duration: 3*casts + 10, Targets: []*proto.Target{
			core.NewDefaultTarget(), core.NewDefaultTarget(), core.NewDefaultTarget(),
		}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1, ExactDamage: true},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	enh := sim.Raid.Parties[0].Players[0].(*EnhancementShaman)

	cl := topRank(t, "Chain Lightning", enh.ChainLightning)
	overload := enh.ChainLightningOverload[cl.Rank]
	if overload == nil {
		t.Fatalf("no overload for Chain Lightning rank %d", cl.Rank)
	}
	for i := range casts {
		at(sim, float64(1+3*i), func(sim *core.Simulation) {
			enh.AddMana(sim, enh.MaxMana(), enh.GetManaNotCastingMetrics())
			cl.CD.Reset()
			castNow(t, sim, enh, cl)
		})
	}
	runSim(sim)

	var landed, overloads int32
	for _, target := range sim.Encounter.TargetUnits {
		m := cl.SpellMetrics()[target.UnitIndex]
		hits := m.Hits + m.Crits
		started := overload.SpellMetrics()[target.UnitIndex].Casts
		landed += hits
		overloads += started
		if rate := float64(started) / float64(hits); math.Abs(rate-.10) > .025 {
			t.Errorf("target %d: %d overloads on %d landed hits (%.1f%%), want 10%%", target.UnitIndex, started, hits, rate*100)
		}
	}
	if rate := float64(overloads) / float64(landed); math.Abs(rate-.10) > .012 {
		t.Errorf("%d overloads on %d landed hits (%.1f%%), want 10%%", overloads, landed, rate*100)
	}
}

// TestOrcShamanClearcastingKeptByMaelstrom checks EV-0180: a Lightning Bolt that consumes five
// Maelstrom Weapon stacks keeps Clearcasting. A Chain Lightning cast while Maelstrom Ready! is up
// costs nothing and takes off Maelstrom Ready! with the stacks left, and since Blizzard's Nov 25
// hotfixes it spends the charge like any other Chain Lightning (Elemental Focus gives it back 10%
// of the time). Maelstrom Ready! goes up at the fifth stack and comes off with the stacks or at a
// Healing Wave.
func TestOrcShamanClearcastingKeptByMaelstrom(t *testing.T) {
	const trials = 40
	sim, enh := weaponSim{level: 60, talents: talentProcTalents, weapon: &proto.ItemSpec{Id: darkEdge}, imbue: proto.WeaponImbue_WindfuryWeapon, seconds: 10*trials + 20}.start()
	sim.PrePull()
	at(sim, 0.5, enh.AutoAttacks.CancelAutoSwing)
	mw, ready, cc := enh.MaelstromWeaponAura, enh.MaelstromReadyAura, enh.ClearcastingAura
	if ready == nil {
		t.Fatalf("no Maelstrom Ready!")
	}
	cl := topRank(t, "Chain Lightning", enh.ChainLightning)
	lb := topRank(t, "Lightning Bolt", enh.LightningBolt)
	hw := topRank(t, "Healing Wave", enh.HealingWave)

	charge := func(sim *core.Simulation) {
		cc.Activate(sim)
		cc.SetStacks(sim, 1)
	}
	var clReadySpent, lbKept, clSpent int
	for i := range trials {
		start := float64(1 + 10*i)
		at(sim, start, func(sim *core.Simulation) {
			enh.AddMana(sim, enh.MaxMana(), enh.GetManaNotCastingMetrics())
			mw.Activate(sim)
			mw.SetStacks(sim, 5)
			if !ready.IsActive() {
				t.Fatalf("five Maelstrom Weapon stacks left Maelstrom Ready! down")
			}

			charge(sim)
			cl.CD.Reset()
			castNow(t, sim, enh, cl)
		})
		at(sim, start+3, func(sim *core.Simulation) {
			if !cc.IsActive() {
				clReadySpent++
			}
			if ready.IsActive() || mw.GetStacks() != 5 {
				t.Errorf("Chain Lightning left Maelstrom Ready! up: %t, %d stacks, want down and 5", ready.IsActive(), mw.GetStacks())
			}

			// Five stacks without Maelstrom Ready!: the bolt is instant and still keeps the charge.
			charge(sim)
			castNow(t, sim, enh, lb)
			if cc.IsActive() {
				lbKept++
			}
			if mw.IsActive() {
				t.Errorf("Lightning Bolt left %d Maelstrom Weapon stacks", mw.GetStacks())
			}

			charge(sim)
			cl.CD.Reset()
			castNow(t, sim, enh, cl)
		})
		at(sim, start+6, func(sim *core.Simulation) {
			if !cc.IsActive() {
				clSpent++
			}
			mw.Activate(sim)
			mw.SetStacks(sim, 5)
			enh.SetGCDTimer(sim, sim.CurrentTime)
			if !hw.Cast(sim, &enh.Unit) {
				t.Fatalf("Healing Wave didn't cast")
			}
		})
		at(sim, start+9.5, func(sim *core.Simulation) {
			if ready.IsActive() || mw.GetStacks() != 5 {
				t.Errorf("Healing Wave left Maelstrom Ready! up: %t, %d stacks, want down and 5", ready.IsActive(), mw.GetStacks())
			}
			mw.SetStacks(sim, 4)
			mw.SetStacks(sim, 5)
			if !ready.IsActive() {
				t.Errorf("the fifth stack left Maelstrom Ready! down")
			}
			mw.Deactivate(sim)
			if ready.IsActive() {
				t.Errorf("Maelstrom Ready! outlasted the stacks")
			}
		})
	}
	runSim(sim)

	if lbKept != trials {
		t.Errorf("Clearcasting kept by %d of %d five stack bolts, want all", lbKept, trials)
	}
	if clReadySpent < trials*7/10 {
		t.Errorf("Chain Lightning under Maelstrom Ready! spent Clearcasting %d of %d times, want about 90%%", clReadySpent, trials)
	}
	if clSpent < trials*7/10 {
		t.Errorf("Chain Lightning without Maelstrom Ready! spent Clearcasting %d of %d times, want about 90%%", clSpent, trials)
	}
}
