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
const armsPresetTalents = "20303213032215201-550500000001"

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
// (development notes post 9, 2026-11-02: was 3 a point over 5 points).
func TestForeverWarriorImprovedTacticalMastery(t *testing.T) {
	for _, c := range []struct {
		talents string
		want    float64
	}{
		{"", 10},
		{"00002", 20},
		{"00003", 25},
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
// was 4% a point over 5 points).
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
	}{{1, 1.1}, {2, 1.2}, {3, 1.3}} {
		talents := "-" + strings.Repeat("0", 16) + string(rune('0'+c.points))
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

// Reckless Swings (development notes post 9, 2026-11-02): 15 sec cooldown, off the GCD here, and the
// next 4 landed melee swings in 8 sec give 5 more Rage each.
func TestForeverWarriorRecklessSwings(t *testing.T) {
	sim, war := newForeverWarriorArmed(t, "-"+strings.Repeat("0", 10)+"1", 991014, 991015)
	war.AutoAttacks.CancelAutoSwing(sim)
	if war.RecklessSwings == nil {
		t.Fatal("no Reckless Swings with the talent")
	}
	swing := func(spell *core.Spell, outcome core.HitOutcome) float64 {
		war.SpendRage(sim, war.CurrentRage(), war.RageRefundMetrics)
		result := spell.NewResult(war.CurrentTarget)
		result.Outcome = outcome
		result.Damage = 100
		spell.DealDamage(sim, result)
		return war.CurrentRage()
	}
	mh, oh := war.AutoAttacks.MHAuto(), war.AutoAttacks.OHAuto()
	baseMH, baseOH := swing(mh, core.OutcomeHit), swing(oh, core.OutcomeHit)

	gcdBefore := war.GCD.TimeToReady(sim)
	if !war.RecklessSwings.Cast(sim, war.CurrentTarget) {
		t.Fatal("Reckless Swings didn't cast")
	}
	if got := war.RecklessSwings.CD.Duration; got != 15*time.Second {
		t.Errorf("Reckless Swings has a %s cooldown, want 15s", got)
	}
	if got := war.GCD.TimeToReady(sim); got != gcdBefore {
		t.Errorf("Reckless Swings took the GCD (%s, was %s)", got, gcdBefore)
	}
	if got := swing(mh, core.OutcomeMiss); got != 0 {
		t.Errorf("a miss gave %g Rage, want 0", got)
	}
	for i, c := range []struct {
		spell *core.Spell
		base  float64
	}{{mh, baseMH}, {oh, baseOH}, {mh, baseMH}, {oh, baseOH}} {
		if got := swing(c.spell, core.OutcomeHit) - c.base; math.Abs(got-5) > 1e-9 {
			t.Errorf("swing %d gave %g more Rage, want 5", i+1, got)
		}
	}
	if war.RecklessSwingsAura.IsActive() {
		t.Error("Reckless Swings still up after 4 swings")
	}
	if got := swing(mh, core.OutcomeHit) - baseMH; math.Abs(got) > 1e-9 {
		t.Errorf("a fifth swing gave %g more Rage, want 0", got)
	}
}

// Improved Whirlwind (development notes post 9, 2026-11-02): 1 sec a point off Whirlwind's
// cooldown, and 15% a point more damage to the closest enemy only.
func TestForeverWarriorImprovedWhirlwind(t *testing.T) {
	damage := func(talents string) (time.Duration, float64, float64) {
		sim, war := newForeverWarriorTargets(t, talents, 991014, 991015, 2)
		war.AutoAttacks.CancelAutoSwing(sim)
		war.Stance = warrior.BerserkerStance
		war.AddRage(sim, 100, war.RageRefundMetrics)
		first, second := war.CurrentTarget, sim.Environment.NextTargetUnit(war.CurrentTarget)
		var d1, d2 float64
		const n = 2000
		for i := 0; i < n; i++ {
			war.Whirlwind.CD.Reset()
			war.GCD.Reset()
			war.AddRage(sim, 100, war.RageRefundMetrics)
			war.Whirlwind.Cast(sim, first)
		}
		d1 = war.Whirlwind.SpellMetrics[first.UnitIndex].TotalDamage
		d2 = war.Whirlwind.SpellMetrics[second.UnitIndex].TotalDamage
		return war.Whirlwind.CD.Duration, d1, d2
	}
	cd0, a0, b0 := damage("")
	cd2, a2, b2 := damage("-" + strings.Repeat("0", 17) + "2")
	if cd0 != 10*time.Second || cd2 != 8*time.Second {
		t.Errorf("Whirlwind's cooldown %s and %s with 2 points, want 10s and 8s", cd0, cd2)
	}
	// Same seed, same rolls: the first target's hits are 30% bigger, the second's the same.
	if r := a2 / a0; math.Abs(r-1.3) > 1e-6 {
		t.Errorf("2 points raise the closest enemy's Whirlwind damage by %.4f, want 1.3", r)
	}
	if r := b2 / b0; math.Abs(r-1) > 1e-6 {
		t.Errorf("2 points change the second enemy's Whirlwind damage by %.4f, want 1", r)
	}
}
