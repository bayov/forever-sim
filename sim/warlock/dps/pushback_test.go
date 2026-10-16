package dps

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestWarlockChannelPushback hits us while we channel Drain Life (5 ticks, one a second). Each
// hit ends the channel sooner: 1 sec for the first hit, then 0.8, 0.6 and so on, as on the
// cmangos Classic server (Spell::DelayedChannel). The ticks past the new end are lost, and a
// tick due right at the new end still comes.
func TestWarlockChannelPushback(t *testing.T) {
	sim, lock := newTankedWarlockSim(t)
	swing := sim.Encounter.TargetUnits[0].AutoAttacks.MHAuto()
	drainLife := lock.DrainLife[len(lock.DrainLife)-1]
	dot := drainLife.Dot(lock.CurrentTarget)

	hit := func(sim *core.Simulation) {
		swing.CalcAndDealDamage(sim, &lock.Unit, 10, swing.OutcomeAlwaysHit)
	}
	check := func(seconds float64, wantEnd float64, wantTicks int32) {
		at(sim, seconds+0.001, func(sim *core.Simulation) {
			if wantEnd == 0 {
				if lock.IsChanneling(sim) || dot.IsActive() {
					t.Errorf("at %.1f sec Drain Life is still channeling until %s, want it over", seconds, dot.ExpiresAt())
				}
			} else if got := dot.ExpiresAt(); (got - core.DurationFromSeconds(wantEnd)).Abs() > time.Microsecond {
				t.Errorf("at %.1f sec Drain Life ends at %s, want %.1f sec", seconds, got, wantEnd)
			}
			if dot.TickCount != wantTicks {
				t.Errorf("at %.1f sec Drain Life ticked %d times, want %d", seconds, dot.TickCount, wantTicks)
			}
		})
	}

	at(sim, 0.1, func(sim *core.Simulation) {
		lock.AddStatDynamic(sim, stats.SpellHit, 20*core.SpellHitRatingPerHitChance)
		castNow(t, sim, lock, drainLife)
		if !dot.IsActive() || (dot.ExpiresAt()-core.DurationFromSeconds(5.1)).Abs() > time.Microsecond {
			t.Fatalf("Drain Life isn't channeling until 5.1 sec")
		}
	})
	// 1 sec off: it ends at 4.1, with the tick due then.
	at(sim, 1.6, hit)
	check(1.6, 4.1, 1)
	check(4.1, 0, 4)
	runSim(sim)

	sim, lock = newTankedWarlockSim(t)
	swing = sim.Encounter.TargetUnits[0].AutoAttacks.MHAuto()
	drainLife = lock.DrainLife[len(lock.DrainLife)-1]
	dot = drainLife.Dot(lock.CurrentTarget)
	at(sim, 0.1, func(sim *core.Simulation) {
		lock.AddStatDynamic(sim, stats.SpellHit, 20*core.SpellHitRatingPerHitChance)
		castNow(t, sim, lock, drainLife)
	})
	at(sim, 1.6, hit)
	// 0.8 off, at 3.3: the tick at 4.1 is lost.
	at(sim, 2.6, hit)
	check(2.6, 3.3, 2)
	// 0.6 would take it past now, so it ends now, after 3 ticks.
	at(sim, 3.2, hit)
	check(3.2, 0, 3)
	runSim(sim)
}

// newTankedWarlockSim is a level 60 warlock that a level 60 mob attacks. The mob swings once
// at the pull and not again in the fight, so only the test's own hits land.
func newTankedWarlockSim(t *testing.T) (*core.Simulation, *DpsWarlock) {
	player := &proto.Player{
		Race:      proto.Race_RaceOrc,
		Class:     proto.Class_ClassWarlock,
		Level:     60,
		Equipment: &proto.EquipmentSpec{},
		Rotation:  &proto.APLRotation{},
		Spec:      &proto.Player_Warlock{Warlock: &proto.Warlock{Options: &proto.WarlockOptions{}}},
	}
	raid := core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
	raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
	mob := core.NormalMobTarget(60, proto.MobType_MobTypeBeast, 1)
	mob.SwingSpeed = 1000
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  &proto.Encounter{Duration: 120, Targets: []*proto.Target{mob}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*DpsWarlock)
}

// TestWarlockChannelPushbackRotation: a mob swings at us while we channel Drain Life over and
// over. Each time a hit ends the channel sooner, the rotation picks its next action right then:
// it channels again, or it finds nothing to do once we're out of mana.
func TestWarlockChannelPushbackRotation(t *testing.T) {
	player := &proto.Player{
		Race:      proto.Race_RaceOrc,
		Class:     proto.Class_ClassWarlock,
		Level:     30,
		Equipment: &proto.EquipmentSpec{},
		Rotation:  core.APLRotationFromJsonString(`{"type":"TypeAPL","priorityList":[{"action":{"channelSpell":{"spellId":{"spellId":709}}}}]}`),
		Spec:      &proto.Player_Warlock{Warlock: &proto.Warlock{Options: &proto.WarlockOptions{}}},
	}
	raid := core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
	raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
	mob := core.NormalMobTarget(30, proto.MobType_MobTypeBeast, 1)
	result := core.RunRaidSim(&proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  &proto.Encounter{Duration: 50, Targets: []*proto.Target{mob}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1, Iterations: 1, Debug: true},
	})
	if result.Error != nil {
		t.Fatal(result.Error.Message)
	}

	logTime := regexp.MustCompile(`^\[(-?[0-9.]+)\] `)
	shortened := regexp.MustCompile(`channel shortened by .* will now end at ([0-9.]+)(ms|s)$`)
	var ends []float64
	// When the rotation picked its next action.
	var casts []float64
	for _, line := range strings.Split(result.Logs, "\n") {
		m := logTime.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		now, _ := strconv.ParseFloat(m[1], 64)
		if s := shortened.FindStringSubmatch(line); s != nil {
			end, _ := strconv.ParseFloat(s[1], 64)
			if s[2] == "ms" {
				end /= 1000
			}
			if len(ends) > 0 && ends[len(ends)-1] > now {
				ends[len(ends)-1] = end
			} else {
				ends = append(ends, end)
			}
		}
		if strings.HasPrefix(line[len(m[0]):], "[ (#1)] Casting {SpellID: 709}") || strings.HasPrefix(line[len(m[0]):], "[ (#1)] No available actions!") {
			casts = append(casts, now)
		}
	}
	if len(ends) < 5 {
		t.Fatalf("%d channels shortened in a minute of swings, want more", len(ends))
	}
	for _, end := range ends {
		if end > 49 {
			continue
		}
		recast := false
		for _, c := range casts {
			if c >= end-0.005 && c <= end+0.015 {
				recast = true
			}
		}
		if !recast {
			t.Errorf("the rotation didn't act when the channel shortened to end at %.2f did", end)
		}
	}
}
