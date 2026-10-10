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

// TestOrcShamanHealingStreamCrits checks that Healing Stream's heal crits for 1.5 times under
// Forever, at our spell crit at the time of the heal, and that its crits leave Water Shield's
// globes alone.
//
// Rank 2 heals for 6 every 2 sec with no healing power. We give 100% crit for the heals at 3
// and 5 sec and take it all away before the ones at 7 and 9. Water Shield's proc flags in
// the client leave periodic heals out, and Forever's Healing Stream is a periodic heal.
func TestOrcShamanHealingStreamCrits(t *testing.T) {
	// Water Shield.
	sim, enh := newShamanSim(newOrcShaman(30, "--000000001", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 10)
	stream := topRank(t, "Healing Stream Totem", enh.HealingStreamTotem)
	heal := enh.GetSpell(core.ActionID{SpellID: shaman.HealingStreamTotemHealId[2]})
	if stream.Rank != 2 || heal == nil || enh.WaterShield == nil {
		t.Fatalf("no Healing Stream Totem rank 2 or no Water Shield at level 30")
	}

	allCrit := 100.0 * core.SpellCritRatingPerCritChance
	at(sim, 1, func(sim *core.Simulation) {
		enh.AddStatDynamic(sim, stats.SpellCrit, allCrit)
		castNow(t, sim, enh, enh.WaterShield)
		castNow(t, sim, enh, stream)
	})
	at(sim, 6, func(sim *core.Simulation) {
		enh.AddStatDynamic(sim, stats.SpellCrit, -2*allCrit)
	})
	runSim(sim)

	metrics := heal.SpellMetrics[enh.UnitIndex]
	if metrics.Crits != 2 || metrics.Hits != 2 || metrics.TotalCritHealing != 18 || metrics.TotalHealing != 30 {
		t.Errorf("got %d crits and %d normal heals for %.1f (%.1f from crits), want 2 and 2 for 30 (18)",
			metrics.Crits, metrics.Hits, metrics.TotalHealing, metrics.TotalCritHealing)
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
		// No crits.
		enh.AddStatDynamic(sim, stats.SpellCrit, -100*core.SpellCritRatingPerCritChance)
		castNow(t, sim, enh, stream)
	})
	runSim(sim)

	if metrics := heal.SpellMetrics[enh.UnitIndex]; metrics.Hits != 2 || math.Abs(metrics.TotalHealing-18) > 1e-9 {
		t.Errorf("got %d heals for %.2f, want 2 for 18", metrics.Hits, metrics.TotalHealing)
	}
}
