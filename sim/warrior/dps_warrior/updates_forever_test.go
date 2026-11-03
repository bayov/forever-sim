package dpswarrior

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/warrior"
)

// The Arms preset's talents (tools/dps_report/req/warrior__arms.json), with Bloodthrill and
// Improved Slam.
const armsPresetTalents = "20305213032515201-55050000001"

// "Slam is no longer subject to spell pushback" (development notes, 2026-10-15, build
// 1.60.1.70425). Classic keeps the pushback.
func TestForeverWarriorSlamNoPushback(t *testing.T) {
	_, war := newForeverWarrior(t, armsPresetTalents)
	if got := war.Slam.PushbackReduction; got != 1 {
		t.Errorf("Forever Slam ignores pushback %.0f%% of the time, want 100%%", 100*got)
	}

	player := &proto.Player{
		Race:          proto.Race_RaceTauren,
		Class:         proto.Class_ClassWarrior,
		Level:         60,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: armsPresetTalents,
		Rotation:      &proto.APLRotation{},
		Spec:          &proto.Player_Warrior{Warrior: &proto.Warrior{Options: &proto.Warrior_Options{}}},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetClassic, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	era := sim.Raid.Parties[0].Players[0].(*DpsWarrior).GetWarrior()
	if got := era.Slam.PushbackReduction; got != 0 {
		t.Errorf("Classic Slam ignores pushback %.0f%% of the time, want 0", 100*got)
	}
}

// Bloodthrill's Overpower window is spell 1282733, 5 sec, since client 70009.
func TestForeverWarriorBloodthrillWindow(t *testing.T) {
	_, war := newForeverWarrior(t, armsPresetTalents)
	if war.BloodthrillAura == nil {
		t.Fatal("no Bloodthrill with the Arms preset's talents")
	}
	if got := war.BloodthrillAura.Duration; got != 5*time.Second {
		t.Errorf("Bloodthrill's Overpower window %s, want 5s", got)
	}
}

// Improved Tactical Mastery keeps 5 Rage a point over 3 points, on top of Tactical Mastery's 10
// (development notes post 9, 2026-11-02: was 3 a point over 5 points). An old 5 point string
// keeps the 3 point 15.
func TestForeverWarriorImprovedTacticalMastery(t *testing.T) {
	for _, c := range []struct {
		talents string
		want    float64
	}{
		{"", 10},
		{"00002", 20},
		{"00003", 25},
		{"00005", 25},
	} {
		sim, war := newForeverWarrior(t, c.talents)
		war.Stance = warrior.BattleStance
		war.AddRage(sim, 100, war.RageRefundMetrics)
		if !war.BerserkerStance.Cast(sim, war.CurrentTarget) {
			t.Fatalf("%q: Berserker Stance didn't cast", c.talents)
		}
		if got := war.CurrentRage(); got != c.want {
			t.Errorf("%q keeps %g Rage through a stance change, want %g", c.talents, got, c.want)
		}
	}
}

// Bloodthrill: 10/20% a main hand hit on a Rend target (development notes post 9, 2026-11-02:
// was 4% a point over 5 points). The Arms preset's old 5 point string gets 20%.
func TestForeverWarriorBloodthrillChance(t *testing.T) {
	for _, c := range []struct {
		talents string
		want    float64
	}{
		{"00000000000100000", 0.1},
		{"00000000000200000", 0.2},
		{armsPresetTalents, 0.2},
	} {
		sim, war := newForeverWarriorArmed(t, c.talents, 991016, 0)
		war.AutoAttacks.CancelAutoSwing(sim)
		target := war.CurrentTarget
		war.Rend.Dot(target).Apply(sim)
		const n = 4000
		procs := 0
		for i := 0; i < n; i++ {
			war.BloodthrillAura.Deactivate(sim)
			result := war.AutoAttacks.MHAuto().NewResult(target)
			result.Outcome = core.OutcomeHit
			result.Damage = 100
			war.AutoAttacks.MHAuto().DealDamage(sim, result)
			if war.BloodthrillAura.IsActive() {
				procs++
			}
		}
		if got := float64(procs) / n; got < c.want-0.025 || got > c.want+0.025 {
			t.Errorf("%q: Bloodthrill procced on %.3f of hits, want %.2f", c.talents, got, c.want)
		}
	}
}

// Flurry: 10/20/30% attack speed over 3 points (development notes post 9, 2026-11-02: was 5 points
// of 5%).
func TestForeverWarriorFlurrySpeed(t *testing.T) {
	for _, c := range []struct {
		points int
		want   float64
	}{{1, 1.1}, {2, 1.2}, {3, 1.3}, {5, 1.3}} {
		talents := "-" + strings.Repeat("0", 14) + string(rune('0'+c.points))
		sim, war := newForeverWarrior(t, talents)
		var flurry *core.Aura
		for _, label := range []string{"Flurry Proc (12319)", "Flurry Proc (12971)", "Flurry Proc (12972)"} {
			if a := war.GetAura(label); a != nil {
				flurry = a
			}
		}
		if flurry == nil {
			t.Fatalf("%d points: no Flurry aura", c.points)
		}
		before := war.PseudoStats.MeleeSpeedMultiplier
		flurry.Activate(sim)
		if got := war.PseudoStats.MeleeSpeedMultiplier / before; math.Abs(got-c.want) > 1e-9 {
			t.Errorf("%d points of Flurry speed attacks up by %.3f, want %.2f", c.points, got, c.want)
		}
	}
}
