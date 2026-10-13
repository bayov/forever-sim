package dpswarrior

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/warrior"
)

// newForeverWarrior builds a naked level 60 Tauren warrior with no rotation (no racial procs), so
// only what a test does happens.
func newForeverWarrior(t *testing.T, talents string) (*core.Simulation, *warrior.Warrior) {
	player := &proto.Player{
		Race:          proto.Race_RaceTauren,
		Class:         proto.Class_ClassWarrior,
		Level:         60,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: talents,
		Rotation:      &proto.APLRotation{},
		Spec:          &proto.Player_Warrior{Warrior: &proto.Warrior{Options: &proto.Warrior_Options{}}},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*DpsWarrior).GetWarrior()
}

// TestForeverWarriorCritRage checks that a critical white hit gives twice the Rage of a normal hit
// with the same damage (the 1 October development notes: 100% more Rage from critical strikes), on
// top of what the damage gives.
func TestForeverWarriorCritRage(t *testing.T) {
	sim, war := newForeverWarrior(t, "")
	mh := war.AutoAttacks.MHAuto()

	gained := func(outcome core.HitOutcome) float64 {
		war.SpendRage(sim, war.CurrentRage(), war.RageRefundMetrics)
		result := mh.NewResult(war.CurrentTarget)
		result.Outcome = outcome
		result.Damage = 300
		mh.DealDamage(sim, result)
		return war.CurrentRage()
	}
	hit := gained(core.OutcomeHit)
	crit := gained(core.OutcomeCrit)
	if want := 300 * 7.5 / core.GetRageConversion(60); math.Abs(hit-want) > 1e-6 {
		t.Errorf("a 300 damage hit gave %.3f Rage, want %.3f", hit, want)
	}
	if math.Abs(crit-2*hit) > 1e-6 {
		t.Errorf("a 300 damage crit gave %.3f Rage, want twice the hit's %.3f", crit, 2*hit)
	}
}

// TestForeverWarriorSpellValues checks the Forever client values (1.60.1.70338) that are easy to get
// wrong: Slam's cooldown and the talent Rage discounts.
func TestForeverWarriorSpellValues(t *testing.T) {
	// Improved Slam 2/2 and Improved Execute 2/2, Booming Voice 5/5 and Raging Blows.
	_, war := newForeverWarrior(t, "00000000000000200-50000000102")
	if got := war.Slam.CD.Duration; got != 15*time.Second {
		t.Errorf("Slam with Improved Slam has a %s cooldown, want 15s", got)
	}
	costs := []struct {
		name  string
		spell *warrior.WarriorSpell
		want  float64
	}{
		{"Execute", war.Execute, 10},
		{"Whirlwind", war.Whirlwind, 22},
		{"Cleave", war.Cleave, 17},
		{"Battle Shout", war.BattleShout, 7.5},
	}
	for _, c := range costs {
		if got := c.spell.DefaultCast.Cost; got != c.want {
			t.Errorf("%s costs %g Rage, want %g", c.name, got, c.want)
		}
	}

	_, war = newForeverWarrior(t, "")
	if got := war.Slam.CD.Duration; got != 18*time.Second {
		t.Errorf("Slam has a %s cooldown, want 18s", got)
	}
}

// TestForeverWarriorDeepWoundsAndFlurry checks two client 70338 values: Deep Wounds' damage spell
// (412613) cannot crit, so its ticks never do, and the Flurry buff (12966) lasts 15 sec.
func TestForeverWarriorDeepWoundsAndFlurry(t *testing.T) {
	_, war := newForeverWarrior(t, "30305013002-25053005112110501-2")
	if war.DeepWounds == nil {
		t.Fatal("the talents should give Deep Wounds")
	}
	if !war.DeepWounds.Flags.Matches(core.SpellFlagNoPeriodicCrit) {
		t.Error("Deep Wounds ticks should not be able to crit under Forever")
	}
	flurry := war.GetAura("Flurry Proc (12974)")
	if flurry == nil {
		t.Fatal("the talents should give 5 points of Flurry")
	}
	if flurry.Duration != 15*time.Second {
		t.Errorf("Flurry lasts %v, want 15s", flurry.Duration)
	}
}

// TestForeverWarriorDeepWoundsRolls checks the 8 October development notes: under Forever, Deep
// Wounds doesn't scale with attack power, and it is a rolling DoT, so a crit while it is up adds
// the damage not yet ticked to the new one.
func TestForeverWarriorDeepWoundsRolls(t *testing.T) {
	sim, war := newForeverWarrior(t, "00305003")
	war.AutoAttacks.MH().BaseDamageMin = 100
	war.AutoAttacks.MH().BaseDamageMax = 200
	mh := war.AutoAttacks.MHAuto()
	dot := war.DeepWounds.Dot(war.CurrentTarget)
	crit := func() float64 {
		result := mh.NewResult(war.CurrentTarget)
		result.Outcome = core.OutcomeCrit
		result.Damage = 300
		mh.DealDamage(sim, result)
		return dot.SnapshotBaseDamage
	}

	first := crit()
	if first <= 0 {
		t.Fatalf("a crit should apply Deep Wounds, its tick is %.3f", first)
	}
	if got := crit(); math.Abs(got-2*first) > 1e-6 {
		t.Errorf("a second crit before any tick gave a %.3f tick, want the first's 4 ticks rolled in: %.3f", got, 2*first)
	}

	dot.Cancel(sim)
	war.AddStatDynamic(sim, stats.AttackPower, 1000)
	if got := crit(); math.Abs(got-first) > 1e-6 {
		t.Errorf("with 1000 more attack power a crit gave a %.3f tick, want %.3f", got, first)
	}
}
