package dpswarrior

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
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
