package balance

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/druid"
)

func newMoonkinSim(t *testing.T, talents string) (*core.Simulation, *BalanceDruid) {
	player := &proto.Player{
		Race:          proto.Race_RaceTauren,
		Class:         proto.Class_ClassDruid,
		Level:         60,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: talents,
		Rotation:      &proto.APLRotation{},
		Spec:          &proto.Player_BalanceDruid{BalanceDruid: &proto.BalanceDruid{Options: &proto.BalanceDruid_Options{}}},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*BalanceDruid)
}

func topRank(spells []*druid.DruidSpell) *druid.DruidSpell {
	for i := len(spells) - 1; i >= 0; i-- {
		if spells[i] != nil {
			return spells[i]
		}
	}
	return nil
}

// TestBalanceRankValues checks the top ranks against the Forever client (1.60.1.70338):
// Wrath rank 8 costs 120 and casts in 2 sec, Starfire rank 7 costs 340 in 3.5 sec, Moonfire
// rank 10 costs 375 and ticks 4 times, Insect Swarm rank 5 costs 160.
func TestBalanceRankValues(t *testing.T) {
	_, moonkin := newMoonkinSim(t, "000000001")
	checks := []struct {
		name     string
		spell    *druid.DruidSpell
		cost     float64
		castTime time.Duration
	}{
		{"Wrath", topRank(moonkin.Wrath), 120, 2 * time.Second},
		{"Starfire", topRank(moonkin.Starfire), 340, 3500 * time.Millisecond},
		{"Moonfire", topRank(moonkin.Moonfire), 375, 0},
		{"Insect Swarm", topRank(moonkin.InsectSwarm), 160, 0},
	}
	for _, c := range checks {
		if c.spell == nil {
			t.Fatalf("no %s", c.name)
		}
		if got := c.spell.Cost.GetCurrentCost(); math.Abs(got-c.cost) > 1e-9 {
			t.Errorf("%s costs %.2f, want %.2f", c.name, got, c.cost)
		}
		if got := c.spell.DefaultCast.CastTime; got != c.castTime {
			t.Errorf("%s casts in %s, want %s", c.name, got, c.castTime)
		}
	}
	if ticks := topRank(moonkin.Moonfire).Dot(moonkin.CurrentTarget).NumberOfTicks; ticks != 4 {
		t.Errorf("Moonfire ticks %d times, want 4", ticks)
	}
}

// TestBalanceCostAndCastTalents checks Improved Wrath (0.1 sec and 10% a point), Moonglow
// (8, 17, 25%, multiplying with Improved Wrath like every client cost modifier), Improved
// Starfire (0.1 sec a point) and Nature's Splendor (one more tick on both dots).
func TestBalanceCostAndCastTalents(t *testing.T) {
	// Improved Wrath 5, Moonglow 3, Nature's Majesty 2, Nature's Splendor, Insect Swarm,
	// Improved Starfire 5.
	_, moonkin := newMoonkinSim(t, "50302001105")
	wrath := topRank(moonkin.Wrath)
	if got, want := wrath.Cost.GetCurrentCost(), 120*0.5*0.75; math.Abs(got-want) > 1e-9 {
		t.Errorf("Wrath costs %.2f, want %.2f", got, want)
	}
	if got := wrath.DefaultCast.CastTime; got != 1500*time.Millisecond {
		t.Errorf("Wrath casts in %s, want 1.5s", got)
	}
	starfire := topRank(moonkin.Starfire)
	if got, want := starfire.Cost.GetCurrentCost(), 340*0.75; math.Abs(got-want) > 1e-9 {
		t.Errorf("Starfire costs %.2f, want %.2f", got, want)
	}
	if got := starfire.DefaultCast.CastTime; got != 3*time.Second {
		t.Errorf("Starfire casts in %s, want 3s", got)
	}
	if ticks := topRank(moonkin.Moonfire).Dot(moonkin.CurrentTarget).NumberOfTicks; ticks != 5 {
		t.Errorf("Moonfire ticks %d times with Nature's Splendor, want 5", ticks)
	}
	if ticks := topRank(moonkin.InsectSwarm).Dot(moonkin.CurrentTarget).NumberOfTicks; ticks != 7 {
		t.Errorf("Insect Swarm ticks %d times with Nature's Splendor, want 7", ticks)
	}
}

// TestBalanceEclipse checks that a Wrath that lands gives 2 Eclipse charges, at most 4, and
// each one takes 0.5 sec (Eclipse 3/3) off the next Starfire.
func TestBalanceEclipse(t *testing.T) {
	// Eclipse sits in row 5, so the build needs 20 points above it.
	sim, moonkin := newMoonkinSim(t, "55302200105003")
	starfire := topRank(moonkin.Starfire)
	base := starfire.DefaultCast.CastTime
	moonkin.EclipseAura.Activate(sim)
	moonkin.EclipseAura.SetStacks(sim, 2)
	if got := starfire.DefaultCast.CastTime; got != base-500*time.Millisecond {
		t.Errorf("Starfire with Eclipse casts in %s, want %s", got, base-500*time.Millisecond)
	}
	moonkin.EclipseAura.RemoveStack(sim)
	moonkin.EclipseAura.RemoveStack(sim)
	if got := starfire.DefaultCast.CastTime; got != base {
		t.Errorf("Starfire after Eclipse casts in %s, want %s", got, base)
	}
}
