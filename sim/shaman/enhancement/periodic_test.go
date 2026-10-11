package enhancement

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/shaman"
	googleProto "google.golang.org/protobuf/proto"
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

// TestOrcShamanFlameShockMiss checks that a missed Flame Shock puts no DoT up and leaves a
// running one alone (shaman_audit.md 7.9).
//
// On the beta none of 6 misses on a mob without the DoT put one up. One miss on a Thundering
// Boulderkin that had the DoT left it running, and its last tick came on time. We cast every
// 0.25 sec at a level 35 target, which our spells miss 39% of the time, and take the DoT down
// before every other cast.
func TestOrcShamanFlameShockMiss(t *testing.T) {
	sim, enh := newTargetLevelSim(30, 35, whirlwind, 0, "")
	sim.PrePull()
	flameShock := topRank(t, "Flame Shock", enh.FlameShock)
	dot := flameShock.Dot(enh.CurrentTarget)
	metrics := &flameShock.SpellMetrics[enh.CurrentTarget.UnitIndex]

	missesWithoutDot, missesWithDot := 0, 0
	for i := range 100 {
		at(sim, 1+0.25*float64(i), func(sim *core.Simulation) {
			if i%2 == 0 {
				dot.Deactivate(sim)
			}
			up, expiresAt, tickCount, misses := dot.IsActive(), dot.ExpiresAt(), dot.TickCount, metrics.Misses
			flameShock.ApplyEffects(sim, enh.CurrentTarget, flameShock)
			switch {
			case metrics.Misses == misses:
				if !dot.IsActive() || dot.ExpiresAt() != sim.CurrentTime+12*time.Second {
					t.Errorf("at %s: a landed Flame Shock didn't put the DoT up for 12 sec", sim.CurrentTime)
				}
			case up:
				missesWithDot++
				if !dot.IsActive() || dot.ExpiresAt() != expiresAt || dot.TickCount != tickCount {
					t.Errorf("at %s: a missed Flame Shock changed the running DoT", sim.CurrentTime)
				}
			default:
				missesWithoutDot++
				if dot.IsActive() {
					t.Errorf("at %s: a missed Flame Shock put the DoT up", sim.CurrentTime)
				}
			}
		})
	}
	runSim(sim)

	if missesWithoutDot == 0 || missesWithDot == 0 {
		t.Errorf("%d misses without the DoT and %d with it, want some of each", missesWithoutDot, missesWithDot)
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

// TestOrcShamanSearingTotemBolts checks how many bolts a Searing Totem fires and when
// (shaman_audit.md 7.22).
//
// Rank 3 lasts 40 sec and shoots every 2.5 sec from 2.5 sec after the drop, so it fires 16
// bolts and the last one as it runs out. On the beta 72 gaps averaged 2.53 sec, and the
// longest run of a rank 3 totem was 16 bolts.
func TestOrcShamanSearingTotemBolts(t *testing.T) {
	sim, enh := newShamanSim(newOrcShaman(30, "", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 50)
	searing := topRank(t, "Searing Totem", enh.SearingTotem)
	if searing.Rank != 3 {
		t.Fatalf("level 30 casts Searing Totem rank %d, want 3", searing.Rank)
	}
	bolt := enh.GetSpell(core.ActionID{SpellID: shaman.SearingTotemAttackSpellId[3]})
	var bolts []time.Duration
	dot := searing.Dot(enh.CurrentTarget)
	onTick := dot.OnTick
	dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		casts := bolt.SpellMetrics[target.UnitIndex].Casts
		onTick(sim, target, dot)
		if bolt.SpellMetrics[target.UnitIndex].Casts == casts+1 {
			bolts = append(bolts, sim.CurrentTime)
		}
	}
	at(sim, 1, func(sim *core.Simulation) { castNow(t, sim, enh, searing) })
	runSim(sim)

	if len(bolts) != 16 {
		t.Fatalf("Searing Totem fired %d bolts at %v, want 16", len(bolts), bolts)
	}
	for i, at := range bolts {
		if want := time.Second + time.Duration(i+1)*2500*time.Millisecond; at != want {
			t.Errorf("bolt %d at %s, want %s", i+1, at, want)
		}
	}
}

// TestOrcShamanMagmaTotemPulses checks when Magma Totem pulses, whom it hits and for how much
// (shaman_audit.md 7.23).
//
// It pulses every 2 sec from 2 sec after the drop, 10 times in its 20 sec, as the client's
// passive triggers its pulse every 2000 ms. Each pulse rolls on every target in the fight.
// Rank 4 deals 73 with no spell power. We take all our spell crit away, and the 3 targets are
// at our level, so a pulse deals 73 to each or misses.
func TestOrcShamanMagmaTotemPulses(t *testing.T) {
	target := googleProto.Clone(core.NewDefaultTarget()).(*proto.Target)
	target.Level = 60
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(newOrcShaman(60, "", &proto.EnhancementShaman_Options{}), &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 30, Targets: []*proto.Target{target, target, target}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	enh := sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
	magma := topRank(t, "Magma Totem", enh.MagmaTotem)
	if magma.Rank != 4 {
		t.Fatalf("level 60 casts Magma Totem rank %d, want 4", magma.Rank)
	}
	pulse := enh.GetSpell(core.ActionID{SpellID: shaman.MagmaTotemAoeSpellId[4]})
	targets := sim.Encounter.TargetUnits

	var pulses []time.Duration
	dot := magma.Dot(enh.CurrentTarget)
	onTick := dot.OnTick
	dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		var before []core.SpellMetrics
		for _, u := range targets {
			before = append(before, pulse.SpellMetrics[u.UnitIndex])
		}
		onTick(sim, target, dot)
		pulses = append(pulses, sim.CurrentTime)
		for i, u := range targets {
			m, b := pulse.SpellMetrics[u.UnitIndex], before[i]
			hits, misses, damage := m.Hits-b.Hits, m.Misses-b.Misses, m.TotalDamage-b.TotalDamage
			if hits+misses != 1 || m.Crits != b.Crits || math.Abs(damage-73*float64(hits)) > 1e-9 {
				t.Errorf("at %s, target %d: %d hits and %d misses for %.2f, want one hit for 73 or one miss",
					sim.CurrentTime, i+1, hits, misses, damage)
			}
		}
	}
	at(sim, 1, func(sim *core.Simulation) {
		enh.AddStatDynamic(sim, stats.SpellCrit, -100*core.SpellCritRatingPerCritChance)
		castNow(t, sim, enh, magma)
	})
	runSim(sim)

	if len(pulses) != 10 {
		t.Fatalf("Magma Totem pulsed %d times at %v, want 10", len(pulses), pulses)
	}
	for i, at := range pulses {
		if want := time.Second + time.Duration(i+1)*2*time.Second; at != want {
			t.Errorf("pulse %d at %s, want %s", i+1, at, want)
		}
	}
}
