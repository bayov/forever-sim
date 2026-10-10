package enhancement

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/shaman"
)

// TestOrcShamanFlameShockTicks checks when Flame Shock's DoT ticks, what a recast does to it,
// and that its ticks crit under Forever.
//
// It ticks every 3 sec from when it goes up, 4 times. A recast drops the old DoT, so the time
// since the last tick is lost and the next tick comes 3 sec after the recast. The beta log
// showed both (shaman_audit.md 5.7). Rank 3 ticks for 14 with no spell power, and a crit tick
// deals 1.5 times that. The roll uses our spell crit at the time of the tick, so we give 100%
// crit for the first DoT and take it all away before the recast.
func TestOrcShamanFlameShockTicks(t *testing.T) {
	sim, enh := newTargetLevelSim(30, 30, whirlwind, 0, "")
	sim.PrePull()
	flameShock := topRank(t, "Flame Shock", enh.FlameShock)
	if flameShock.Rank != 3 {
		t.Fatalf("level 30 casts Flame Shock rank %d, want 3", flameShock.Rank)
	}

	type tick struct {
		at     time.Duration
		damage float64
		crit   bool
	}
	var ticks []tick
	metrics := &flameShock.SpellMetrics[enh.CurrentTarget.UnitIndex]
	dot := flameShock.Dot(enh.CurrentTarget)
	onTick := dot.OnTick
	dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		damage, critDamage := metrics.TotalTickDamage, metrics.TotalCritTickDamage
		onTick(sim, target, dot)
		ticks = append(ticks, tick{sim.CurrentTime, metrics.TotalTickDamage - damage, metrics.TotalCritTickDamage > critDamage})
	}

	allCrit := 100.0 * core.SpellCritRatingPerCritChance
	at(sim, 1, func(sim *core.Simulation) {
		enh.AddStatDynamic(sim, stats.SpellCrit, allCrit)
		dot.Apply(sim)
	})
	at(sim, 8, func(sim *core.Simulation) {
		enh.AddStatDynamic(sim, stats.SpellCrit, -2*allCrit)
	})
	at(sim, 8.5, func(sim *core.Simulation) {
		dot.Apply(sim)
	})
	runSim(sim)

	// The first DoT's ticks at 10 and 13 sec never come.
	want := []tick{
		{4 * time.Second, 21, true},
		{7 * time.Second, 21, true},
		{11500 * time.Millisecond, 14, false},
		{14500 * time.Millisecond, 14, false},
		{17500 * time.Millisecond, 14, false},
		{20500 * time.Millisecond, 14, false},
	}
	if len(ticks) != len(want) {
		t.Fatalf("got %d ticks %v, want %d", len(ticks), ticks, len(want))
	}
	for i, w := range want {
		if ticks[i].at != w.at || math.Abs(ticks[i].damage-w.damage) > 1e-9 || ticks[i].crit != w.crit {
			t.Errorf("tick %d: got %+v, want %+v", i+1, ticks[i], w)
		}
	}
}

// TestOrcShamanHealingStreamCrits checks that Healing Stream's heal crits 5% of the time under
// Forever, whatever our spell crit, and that its crits leave Water Shield's globes alone.
//
// On the beta our totems' heals crit on 27 of 549 with 12.3% spell crit on the sheet
// (shaman_audit.md 5.7). We heal 1000 times with 100% spell crit and 1000 times with none, and
// expect about 50 crits each time. Rank 2 heals for 6 with no healing power, and 9 on a crit.
// Water Shield's proc flags in the client leave periodic heals out, and Forever's Healing
// Stream is a periodic heal.
func TestOrcShamanHealingStreamCrits(t *testing.T) {
	// Water Shield.
	sim, enh := newShamanSim(newOrcShaman(30, "--000000001", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 10)
	heal := enh.GetSpell(core.ActionID{SpellID: shaman.HealingStreamTotemHealId[2]})
	if heal == nil || enh.WaterShield == nil {
		t.Fatalf("no Healing Stream Totem rank 2 or no Water Shield at level 30")
	}
	metrics := &heal.SpellMetrics[enh.UnitIndex]

	heals := func(sim *core.Simulation) int32 {
		crits := metrics.Crits
		for range 1000 {
			heal.Cast(sim, &enh.Unit)
		}
		return metrics.Crits - crits
	}
	allCrit := 100.0 * core.SpellCritRatingPerCritChance
	var withCrit, withoutCrit int32
	at(sim, 1, func(sim *core.Simulation) {
		castNow(t, sim, enh, enh.WaterShield)
		enh.AddStatDynamic(sim, stats.SpellCrit, allCrit)
		withCrit = heals(sim)
		enh.AddStatDynamic(sim, stats.SpellCrit, -2*allCrit)
		withoutCrit = heals(sim)
	})
	runSim(sim)

	// 50 crits give or take 7.
	if withCrit < 25 || withCrit > 75 || withoutCrit < 25 || withoutCrit > 75 {
		t.Errorf("got %d crits in 1000 heals with 100%% spell crit and %d with none, want about 50 each", withCrit, withoutCrit)
	}
	crits, hits := float64(metrics.Crits), float64(metrics.Hits)
	if crits+hits != 2000 || math.Abs(metrics.TotalCritHealing-9*crits) > 1e-6 || math.Abs(metrics.TotalHealing-9*crits-6*hits) > 1e-6 {
		t.Errorf("got %.0f crits and %.0f normal heals for %.1f (%.1f from crits), want 9 a crit and 6 a normal heal",
			crits, hits, metrics.TotalHealing, metrics.TotalCritHealing)
	}
	if stacks := enh.WaterShieldAura.GetStacks(); stacks != 3 {
		t.Errorf("Water Shield has %d globes, want 3", stacks)
	}
}

// TestOrcShamanHealingStreamTalents checks the talents that raise Healing Stream's heal.
//
// Restorative Totems adds 10% a point and Purification nothing, so with 5 points in each rank
// 2 heals for 6 * 1.5 = 9. Purification only lists Healing Wave, Lesser Healing Wave and Chain
// Heal in the client.
func TestOrcShamanHealingStreamTalents(t *testing.T) {
	// Restorative Totems 5 and Purification 5.
	sim, enh := newShamanSim(newOrcShaman(30, "--000000000050005", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 6)
	if enh.Talents.RestorativeTotems != 5 || enh.Talents.Purification != 5 {
		t.Fatalf("got Restorative Totems %d and Purification %d, want 5 and 5", enh.Talents.RestorativeTotems, enh.Talents.Purification)
	}
	stream := topRank(t, "Healing Stream Totem", enh.HealingStreamTotem)
	heal := enh.GetSpell(core.ActionID{SpellID: shaman.HealingStreamTotemHealId[2]})

	at(sim, 1, func(sim *core.Simulation) {
		castNow(t, sim, enh, stream)
	})
	runSim(sim)

	// A crit heals for 13.5.
	metrics := heal.SpellMetrics[enh.UnitIndex]
	if want := 9*float64(metrics.Hits) + 13.5*float64(metrics.Crits); metrics.Hits+metrics.Crits != 2 || math.Abs(metrics.TotalHealing-want) > 1e-9 {
		t.Errorf("got %d normal heals and %d crits for %.2f, want 2 heals for 9 each (13.5 on a crit)", metrics.Hits, metrics.Crits, metrics.TotalHealing)
	}
}
