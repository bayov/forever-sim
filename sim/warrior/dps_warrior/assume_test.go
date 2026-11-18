package dpswarrior

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

// assumeWarrior is a level 60 Orc Fury warrior under Forever, built after DASH_ASSUME is set to
// env (the DPS Report's assumption knobs, sim/core/report_switches.go).
func assumeWarrior(t *testing.T, env string) (*core.Simulation, *DpsWarrior) {
	t.Helper()
	core.AssumeForTest(env)
	t.Cleanup(func() { core.AssumeForTest("") })
	items := make([]*proto.ItemSpec, 17)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 991014}
	items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: 991015}
	player := &proto.Player{
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassWarrior,
		Level:         60,
		Equipment:     &proto.EquipmentSpec{Items: items},
		TalentsString: P1Talents,
		Rotation:      &proto.APLRotation{},
		Spec:          &proto.Player_Warrior{Warrior: &proto.Warrior{Options: &proto.Warrior_Options{}}},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 600, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*DpsWarrior)
}

// TestAssumeBloodFuryAPOnly: with orc_bf_weapon=0 Blood Fury gives its 10% attack power and no
// weapon damage. By default it adds 10% of each weapon's average damage too.
func TestAssumeBloodFuryAPOnly(t *testing.T) {
	for _, c := range []struct {
		env    string
		weapon bool
	}{{"", true}, {"orc_bf_weapon=0", false}} {
		sim, dw := assumeWarrior(t, c.env)
		war := dw.GetWarrior()
		bf := war.GetAura("Blood Fury")
		if bf == nil {
			t.Fatal("no Blood Fury on the Orc")
		}
		mh, oh := war.AutoAttacks.MH(), war.AutoAttacks.OH()
		ap, mhMin, ohMax, mhAvg := war.GetStat(stats.AttackPower), mh.BaseDamageMin, oh.BaseDamageMax, mh.AverageDamage()
		bf.Activate(sim)
		if got := war.GetStat(stats.AttackPower); math.Abs(got-1.1*ap) > 0.01 {
			t.Errorf("%q: attack power %.2f under Blood Fury, want %.2f", c.env, got, 1.1*ap)
		}
		wantMin := mhMin
		if c.weapon {
			wantMin += 0.1 * mhAvg
		}
		if math.Abs(mh.BaseDamageMin-wantMin) > 1e-9 {
			t.Errorf("%q: main hand minimum %.2f under Blood Fury, want %.2f", c.env, mh.BaseDamageMin, wantMin)
		}
		if !c.weapon && oh.BaseDamageMax != ohMax {
			t.Errorf("%q: off-hand maximum %.2f under Blood Fury, want %.2f", c.env, oh.BaseDamageMax, ohMax)
		}
		bf.Deactivate(sim)
		if mh.BaseDamageMin != mhMin || war.GetStat(stats.AttackPower) != ap {
			t.Errorf("%q: Blood Fury's expiry left the main hand at %.2f and attack power at %.2f, want %.2f and %.2f",
				c.env, mh.BaseDamageMin, war.GetStat(stats.AttackPower), mhMin, ap)
		}
	}
}
