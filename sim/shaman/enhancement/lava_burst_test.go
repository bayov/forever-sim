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

// TestOrcShamanLavaBurstRanks checks that a rotation naming Lava Burst rank 3 casts the rank
// the level knows.
//
// The talent teaches rank 1 at 40, and the trainer has rank 2 at 50 and rank 3 at 60. Each
// rank has its own Forever spell ID.
func TestOrcShamanLavaBurstRanks(t *testing.T) {
	cases := []struct {
		level int32
		rank  int
	}{{45, 1}, {55, 2}, {60, 3}}
	for _, c := range cases {
		// Lava Burst.
		player := newOrcShaman(c.level, "0000000000000001", &proto.EnhancementShaman_Options{})
		player.Rotation = &proto.APLRotation{PriorityList: []*proto.APLListItem{{Action: &proto.APLAction{
			Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{
				SpellId: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: shaman.LavaBurstSpellId[3]}},
			}},
		}}}}
		sim, enh := newShamanSim(player, &proto.Debuffs{}, 15)
		runSim(sim)

		lavaBurst := enh.LavaBurst
		if got, want := lavaBurst.ActionID.SpellID, shaman.LavaBurstSpellId[c.rank]; got != want || lavaBurst.Rank != c.rank {
			t.Errorf("level %d: got Lava Burst %d rank %d, want %d rank %d", c.level, got, lavaBurst.Rank, want, c.rank)
		}
		if casts := lavaBurst.SpellMetrics[enh.CurrentTarget.UnitIndex].Casts; casts < 2 {
			t.Errorf("level %d: the rotation cast Lava Burst %d times in 15 sec", c.level, casts)
		}
	}
}

// TestOrcShamanLavaBurstDamage checks Lava Burst's damage, its bonus with Flame Shock up, its
// cast, travel time and cooldown (shaman_audit.md 7.14).
//
// Rank 3 deals 192.14 to 247.86 at level 60 with no spell power, and 20% more with our Flame
// Shock on the target. We take all our spell crit away, and the target is at our level. Half
// of 1000 hits have Flame Shock up, so the two averages should come out near 220 and 264.
//
// The cast takes 2.5 sec and the lava flies at 20 yards a second, so from 20 yards it lands 1
// sec after the cast ends. The 10 sec cooldown counts from the end of the cast.
func TestOrcShamanLavaBurstDamage(t *testing.T) {
	// Lava Burst.
	sim, enh := newTargetLevelSim(60, 60, 0, 0, "0000000000000001")
	sim.PrePull()
	lavaBurst := enh.LavaBurst
	if lavaBurst == nil || lavaBurst.Rank != 3 {
		t.Fatalf("level 60 has no Lava Burst rank 3")
	}
	target := enh.CurrentTarget
	flameShock := topRank(t, "Flame Shock", enh.FlameShock).Dot(target)
	metrics := &lavaBurst.SpellMetrics[target.UnitIndex]

	const low, high = 192.14, 247.86
	multipliers := []float64{1, 1.2}
	var sums [2]float64
	var hits [2]int
	// The lava lands in an action of its own even at 0 yards, so we read each hit in the next
	// cast's action.
	last, damage, misses := -1, 0.0, int32(0)
	read := func() {
		if last >= 0 && metrics.Misses == misses {
			got, m := metrics.TotalDamage-damage, multipliers[last]
			if got < low*m-1e-6 || got > high*m+1e-6 {
				t.Fatalf("Lava Burst hit for %.2f, want %.2f to %.2f", got, low*m, high*m)
			}
			sums[last] += got
			hits[last]++
		}
		damage, misses = metrics.TotalDamage, metrics.Misses
	}
	at(sim, 0.5, func(sim *core.Simulation) {
		enh.AddStatDynamic(sim, stats.SpellCrit, -100*core.SpellCritRatingPerCritChance)
	})
	for i := range 1000 {
		at(sim, 1+0.001*float64(i), func(sim *core.Simulation) {
			read()
			last = i % 2
			if last == 1 {
				flameShock.Apply(sim)
			} else {
				flameShock.Deactivate(sim)
			}
			lavaBurst.ApplyEffects(sim, target, lavaBurst)
		})
	}
	at(sim, 2, func(sim *core.Simulation) {
		read()
		flameShock.Deactivate(sim)
		for i, m := range multipliers {
			// 4 standard errors of the mean of a flat roll.
			mean, want := sums[i]/float64(hits[i]), (low+high)/2*m
			if tolerance := 4 * (high - low) * m / math.Sqrt(12*float64(hits[i])); math.Abs(mean-want) > tolerance {
				t.Errorf("Lava Burst hit for %.2f on average with Flame Shock bonus %.1f, want %.2f", mean, m, want)
			}
		}
		if metrics.Crits > 0 {
			t.Errorf("Lava Burst crit %d times with no spell crit", metrics.Crits)
		}
		if metrics.Misses == 0 {
			t.Errorf("none of 1000 Lava Bursts missed")
		}
	})
	at(sim, 2.5, func(sim *core.Simulation) {
		enh.DistanceFromTarget = 20
		enh.AddMana(sim, enh.MaxMana()-enh.CurrentMana(), enh.GetManaNotCastingMetrics())
		castNow(t, sim, enh, lavaBurst)
		if readyAt, want := lavaBurst.CD.ReadyAt(), sim.CurrentTime+12500*time.Millisecond; readyAt != want {
			t.Errorf("Lava Burst cast at %s is ready at %s, want %s", sim.CurrentTime, readyAt, want)
		}
	})
	// The cast ends at 5 sec, and the hit roll comes then.
	var hitsBefore, missesBefore int32
	var damageBefore float64
	at(sim, 4.99, func(sim *core.Simulation) {
		hitsBefore, missesBefore, damageBefore = metrics.Hits, metrics.Misses, metrics.TotalDamage
	})
	at(sim, 5.01, func(sim *core.Simulation) {
		if metrics.Misses > missesBefore {
			t.Fatalf("the cast missed, so we can't see when it lands")
		}
		if metrics.Hits != hitsBefore+1 {
			t.Errorf("Lava Burst didn't roll at the end of its 2.5 sec cast")
		}
		if metrics.TotalDamage != damageBefore {
			t.Errorf("Lava Burst landed when the cast ended, 1 sec before the lava gets there")
		}
	})
	at(sim, 6.01, func(sim *core.Simulation) {
		if metrics.TotalDamage == damageBefore {
			t.Errorf("Lava Burst hadn't landed 1 sec after the cast ended")
		}
	})
	runSim(sim)
}
