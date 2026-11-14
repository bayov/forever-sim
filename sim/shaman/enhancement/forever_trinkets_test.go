package enhancement

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/common"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	googleProto "google.golang.org/protobuf/proto"
)

// newTrinketSim starts a Forever sim of a level 60 Orc shaman with the given trinkets and no
// weapon, against the given number of level 63 targets with no armor.
func newTrinketSim(t *testing.T, targets int, seconds float64, trinkets ...int32) (*core.Simulation, *EnhancementShaman) {
	player := newOrcShaman(60, stormstrike, &proto.EnhancementShaman_Options{})
	for range proto.ItemSlot_ItemSlotRanged + 1 {
		player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{})
	}
	for i, id := range trinkets {
		player.Equipment.Items[proto.ItemSlot_ItemSlotTrinket1+proto.ItemSlot(i)].Id = id
	}
	encounterTargets := make([]*proto.Target, targets)
	for i := range encounterTargets {
		target := googleProto.Clone(core.NewDefaultTarget()).(*proto.Target)
		target.Stats[proto.Stat_StatArmor] = 0
		encounterTargets[i] = target
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: seconds, Targets: encounterTargets},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1, ExactDamage: true},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
}

// landHit deals a landed hit of spell on the target, or a landed tick when periodic is set.
func landHit(sim *core.Simulation, spell *core.Spell, target *core.Unit, periodic bool) {
	result := spell.NewResult(target)
	result.Outcome = core.OutcomeHit
	result.Damage = 100
	if periodic {
		spell.DealPeriodicDamage(sim, result)
	} else {
		spell.DealDamage(sim, result)
	}
}

// Bolt's Charged Eyeball procs on 15% of our damaging hits and ticks, at most once every 3 sec,
// and half of the procs shock the target for 60 Nature (the other half hit us).
//
// We land hits 3 sec apart, so the cooldown is always over and each hit procs on its own roll.
// Then 7.5% of them shock the target. With 4000 hits we expect 300 shocks, and the bounds are
// about 3 standard deviations wide.
func TestForeverBoltsChargedEyeball(t *testing.T) {
	sim, enh := newTrinketSim(t, 1, 100000, common.BoltsChargedEyeball)
	shock := enh.GetSpell(core.ActionID{SpellID: 1315996})
	if shock == nil {
		t.Fatal("no Grounded Electroshock with Bolt's Charged Eyeball")
	}
	target := enh.CurrentTarget
	casts := func() int32 { return shock.SpellMetrics[target.UnitIndex].Casts }
	flameShock := topRank(t, "Flame Shock", enh.FlameShock)

	for _, tc := range []struct {
		name     string
		spell    *core.Spell
		periodic bool
	}{
		{"Stormstrike", enh.Stormstrike, false},
		{"Flame Shock tick", flameShock, true},
	} {
		before := casts()
		for range 4000 {
			sim.CurrentTime += 3 * time.Second
			landHit(sim, tc.spell, target, tc.periodic)
		}
		if got := casts() - before; got < 250 || got > 350 {
			t.Errorf("%d shocks on the target from 4000 %s hits 3 sec apart, want about 300", got, tc.name)
		}
	}

	// Hits 0.5 sec apart proc at most once every 3 sec.
	before := casts()
	for range 6000 {
		sim.CurrentTime += 500 * time.Millisecond
		landHit(sim, enh.Stormstrike, target, false)
	}
	if got := casts() - before; got > 1000 {
		t.Errorf("%d shocks in 3000 sec, want at most 1000", got)
	}

	// Each shock that lands deals 60 Nature, or 90 on a crit, before the target's resistance.
	metrics := shock.SpellMetrics[target.UnitIndex]
	landed := metrics.Hits + metrics.ResistedHits + metrics.Crits + metrics.ResistedCrits
	if landed == 0 || metrics.TotalDamage <= 0 || metrics.TotalDamage > 90*float64(landed) {
		t.Errorf("%d landed shocks dealt %.1f, want more than 0 and up to 90 each", landed, metrics.TotalDamage)
	}

	// The shock itself and other spells that don't proc equip effects never shock.
	before = casts()
	for range 1000 {
		sim.CurrentTime += 3 * time.Second
		landHit(sim, shock, target, false)
	}
	if got := casts() - before; got != 0 {
		t.Errorf("%d shocks from 1000 shocks, want none", got)
	}

}

// Everlook Pathcarver's use deals 607 Physical damage split between up to 4 targets. It rolls
// spell hit and crit (150%), has a 5 min cooldown and puts the other burst trinkets on a 15 sec
// cooldown.
func TestForeverEverlookPathcarver(t *testing.T) {
	for _, tc := range []struct {
		targets int
		each    float64
		hit     int
	}{
		{1, 607, 1},
		{3, 607.0 / 3, 3},
		{6, 607.0 / 4, 4},
	} {
		sim, enh := newTrinketSim(t, tc.targets, 1000000, common.EverlookPathcarver, 274760)
		use := enh.GetSpell(core.ActionID{ItemID: common.EverlookPathcarver})
		if use == nil {
			t.Fatal("no Everlook Pathcarver use")
		}
		if enh.GetMajorCooldown(use.ActionID) == nil {
			t.Fatal("Everlook Pathcarver is not a major cooldown, so the rotation won't use it")
		}
		deliveryBot := enh.GetSpell(core.ActionID{ItemID: 274760})

		if !use.Cast(sim, enh.CurrentTarget) {
			t.Fatal("Everlook Pathcarver didn't cast")
		}
		if got := use.CD.TimeToReady(sim); got != 5*time.Minute {
			t.Errorf("cooldown %s after the use, want 5m", got)
		}
		if deliveryBot.CanCast(sim, enh.CurrentTarget) {
			t.Error("Everlook Delivery Bot is ready right after Everlook Pathcarver")
		}
		sim.CurrentTime += 15 * time.Second
		if !deliveryBot.CanCast(sim, enh.CurrentTarget) {
			t.Error("Everlook Delivery Bot still not ready 15 sec after Everlook Pathcarver")
		}

		for range 999 {
			sim.CurrentTime += 5 * time.Minute
			use.Cast(sim, enh.CurrentTarget)
		}

		hitTargets := 0
		for _, target := range sim.Encounter.TargetUnits {
			m := use.SpellMetrics[target.UnitIndex]
			if m.Casts+m.Hits+m.Crits+m.Misses == 0 {
				continue
			}
			hitTargets++
			if m.Hits+m.Crits+m.Misses != 1000 {
				t.Errorf("%d targets: %d outcomes on a target from 1000 uses", tc.targets, m.Hits+m.Crits+m.Misses)
			}
			if m.Misses == 0 || m.Crits == 0 {
				t.Errorf("%d targets: %d misses and %d crits in 1000 uses, want some of each", tc.targets, m.Misses, m.Crits)
			}
			wantHit := tc.each * float64(m.Hits)
			wantCrit := tc.each * 1.5 * float64(m.Crits)
			if math.Abs(m.TotalCritDamage-wantCrit) > 1e-6 || math.Abs(m.TotalDamage-wantHit-wantCrit) > 1e-6 {
				t.Errorf("%d targets: %d hits and %d crits dealt %.2f (%.2f on crits), want %.2f a hit",
					tc.targets, m.Hits, m.Crits, m.TotalDamage, m.TotalCritDamage, tc.each)
			}
		}
		if hitTargets != tc.hit {
			t.Errorf("%d targets: %d took damage, want %d", tc.targets, hitTargets, tc.hit)
		}
	}
}

// Chained Essence of Eranikus' cloud ticks 140 Nature on every enemy every 2 sec, 9 times
// before it ends at 20 sec. The ticks can crit (210). It has a 10 min cooldown and puts the
// other burst trinkets on a 20 sec cooldown.
func TestForeverChainedEssenceOfEranikus(t *testing.T) {
	sim, enh := newTrinketSim(t, 3, 30, common.ChainedEssenceOfEranikus, 274760)
	use := enh.GetSpell(core.ActionID{ItemID: common.ChainedEssenceOfEranikus})
	cloud := enh.GetSpell(core.ActionID{SpellID: 1300240})
	if use == nil || cloud == nil {
		t.Fatal("no Chained Essence of Eranikus use or cloud")
	}
	if enh.GetMajorCooldown(use.ActionID) == nil {
		t.Fatal("Chained Essence of Eranikus is not a major cooldown, so the rotation won't use it")
	}
	deliveryBot := enh.GetSpell(core.ActionID{ItemID: 274760})

	at(sim, 1, func(sim *core.Simulation) {
		if !use.Cast(sim, enh.CurrentTarget) {
			t.Fatal("Chained Essence of Eranikus didn't cast")
		}
		if got := use.CD.TimeToReady(sim); got != 10*time.Minute {
			t.Errorf("cooldown %s after the use, want 10m", got)
		}
	})
	at(sim, 20.9, func(sim *core.Simulation) {
		if deliveryBot.CanCast(sim, enh.CurrentTarget) {
			t.Error("Everlook Delivery Bot is ready 19.9 sec after Chained Essence of Eranikus")
		}
	})
	at(sim, 21, func(sim *core.Simulation) {
		if !deliveryBot.CanCast(sim, enh.CurrentTarget) {
			t.Error("Everlook Delivery Bot still not ready 20 sec after Chained Essence of Eranikus")
		}
	})
	runSim(sim)

	clouded := 0
	for _, target := range sim.Encounter.TargetUnits {
		m := cloud.SpellMetrics[target.UnitIndex]
		if m.Misses > 0 {
			continue
		}
		clouded++
		if ticks := m.Ticks + m.ResistedTicks + m.CritTicks + m.ResistedCritTicks; ticks != 9 {
			t.Errorf("%s took %d ticks, want 9", target.Label, ticks)
		}
		if m.TotalTickDamage <= 0 || m.TotalTickDamage > 9*210 {
			t.Errorf("%s took %.1f from the cloud, want more than 0 and up to 9 * 210", target.Label, m.TotalTickDamage)
		}
	}
	if clouded == 0 {
		t.Error("the cloud missed every target")
	}
}
