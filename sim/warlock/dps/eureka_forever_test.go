package dps

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
)

// Gnome Eureka! for a Warlock covers what client 1.60.1.70425's spell 1259821 names: the cost
// and direct damage of Shadow Bolt, Immolate, Searing Pain and the like, the ticks of Drain
// Life, Drain Soul and Wrack, and nothing on Corruption, the banes, Siphon Life or Incinerate
// (development notes 2026-10-01: "Eureka! no longer benefits periodic effects at all.
// Channeled spells do not count as periodics.").
func TestForeverEurekaWarlockScope(t *testing.T) {
	player := &proto.Player{
		Race:      proto.Race_RaceGnome,
		Class:     proto.Class_ClassWarlock,
		Level:     60,
		Equipment: &proto.EquipmentSpec{},
		Rotation:  &proto.APLRotation{},
		Spec:      &proto.Player_Warlock{Warlock: &proto.Warlock{Options: &proto.WarlockOptions{}}},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 120, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	lock := sim.Raid.Parties[0].Players[0].(*DpsWarlock)

	top := func(spells []*core.Spell) *core.Spell { return spells[len(spells)-1] }
	type snapshot struct{ mult, impact, cost float64 }
	snap := func(s *core.Spell) snapshot {
		cost := 0.0
		if s.Cost != nil {
			cost = float64(s.Cost.Multiplier) / 100 * s.Cost.Factor
		}
		return snapshot{s.DamageMultiplier, s.ImpactDamageMultiplierAdditive, cost}
	}
	cases := []struct {
		name               string
		spell              *core.Spell
		mult, impact, cost float64 // the change Eureka! makes; the cost is a factor of its own
	}{
		{"Shadow Bolt", top(lock.ShadowBolt), 1.1, 0, 0.9},
		{"Immolate", top(lock.Immolate), 1, 0.1, 0.9},
		{"Drain Life", top(lock.DrainLife), 1.1, 0, 0.9},
		{"Corruption", top(lock.Corruption), 1, 0, 1},
		{"Bane of Agony", top(lock.BaneOfAgony), 1, 0, 1},
	}
	before := make([]snapshot, len(cases))
	for i, c := range cases {
		before[i] = snap(c.spell)
	}

	eureka := lock.GetAura("Eureka!")
	if eureka == nil {
		t.Fatal("the Gnome has no Eureka!")
	}
	eureka.Activate(sim)
	for i, c := range cases {
		got := snap(c.spell)
		if math.Abs(got.mult/before[i].mult-c.mult) > 1e-9 || math.Abs(got.impact-before[i].impact-c.impact) > 1e-9 ||
			math.Abs(got.cost/before[i].cost-c.cost) > 1e-9 {
			t.Errorf("%s (%d) with Eureka!: damage x%.3f, hit +%.2f, cost x%.3f; want x%.3f, +%.2f, x%.3f", c.name, c.spell.SpellID,
				got.mult/before[i].mult, got.impact-before[i].impact, got.cost/before[i].cost, c.mult, c.impact, c.cost)
		}
	}

	eureka.Deactivate(sim)
	for i, c := range cases {
		if got := snap(c.spell); math.Abs(got.mult-before[i].mult) > 1e-9 || math.Abs(got.impact-before[i].impact) > 1e-9 || math.Abs(got.cost-before[i].cost) > 1e-9 {
			t.Errorf("%s not restored after Eureka!: %+v, want %+v", c.name, got, before[i])
		}
	}
}
