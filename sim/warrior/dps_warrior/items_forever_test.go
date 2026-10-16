package dpswarrior

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
)

func newWarriorWithMainHand(t *testing.T, ruleset proto.Ruleset, mainHand int32) (*core.Simulation, *DpsWarrior) {
	items := make([]*proto.ItemSpec, 17)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: mainHand}
	player := &proto.Player{
		Race:      proto.Race_RaceTauren,
		Class:     proto.Class_ClassWarrior,
		Level:     60,
		Equipment: &proto.EquipmentSpec{Items: items},
		Rotation:  &proto.APLRotation{},
		Spec:      &proto.Player_Warrior{Warrior: &proto.Warrior{Options: &proto.Warrior_Options{}}},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 600, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: ruleset, RandomSeed: 1, ExactDamage: true},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*DpsWarrior)
}

// Alcor's Sunrazor's Firebolt (18833) is 105 with variance 0.2, 94.5 to 115.5, since Forever
// client 1.60.1.70425 (75 to 105 before, and in Classic).
func TestForeverAlcorsSunrazorFirebolt(t *testing.T) {
	for _, c := range []struct {
		ruleset  proto.Ruleset
		min, max float64
	}{
		{proto.Ruleset_RulesetForever, 94.5, 115.5},
		{proto.Ruleset_RulesetClassic, 75, 105},
	} {
		sim, war := newWarriorWithMainHand(t, c.ruleset, 14555)
		firebolt := war.GetSpell(core.ActionID{SpellID: 18833})
		if firebolt == nil {
			t.Fatalf("%s: no Firebolt with Alcor's Sunrazor", c.ruleset)
		}
		// Read each landed, uncritted cast's damage off the metrics, with resists off.
		firebolt.Flags |= core.SpellFlagIgnoreResists
		metrics := &firebolt.SpellMetrics[war.CurrentTarget.UnitIndex]
		lo, hi := 1e9, 0.0
		for i := 0; i < 2000; i++ {
			damage, hits := metrics.TotalDamage, metrics.Hits
			firebolt.Cast(sim, war.CurrentTarget)
			if metrics.Hits == hits+1 {
				d := metrics.TotalDamage - damage
				lo, hi = min(lo, d), max(hi, d)
			}
		}
		// 2000 rolls come within 1% of each end.
		if lo < c.min-1e-6 || hi > c.max+1e-6 || lo > c.min+0.3 || hi < c.max-0.3 {
			t.Errorf("%s: Firebolt %.1f to %.1f, want %.1f to %.1f", c.ruleset, lo, hi, c.min, c.max)
		}
	}
}

// Shadowstrike's Drain Life (21170) gets no spell power under Forever since client 70009.
func TestForeverShadowstrikeNoSpellPower(t *testing.T) {
	_, war := newWarriorWithMainHand(t, proto.Ruleset_RulesetForever, 17074)
	if got := war.GetSpell(core.ActionID{SpellID: 461683}).BonusCoefficient; got != 0 {
		t.Errorf("Forever Shadowstrike proc coefficient %.2f, want 0", got)
	}
	_, era := newWarriorWithMainHand(t, proto.Ruleset_RulesetClassic, 17074)
	if got := era.GetSpell(core.ActionID{SpellID: 461683}).BonusCoefficient; got != 1 {
		t.Errorf("Classic Shadowstrike proc coefficient %.2f, want 1", got)
	}
}
