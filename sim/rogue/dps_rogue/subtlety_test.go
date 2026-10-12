package dpsrogue

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
)

// newSubtletyRogueSim starts a Forever sim of a level 60 rogue with 21 Subtlety points
// (Premeditation and Preparation) and an empty rotation.
func newSubtletyRogueSim() (*core.Simulation, *DpsRogue) {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotOffHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 18832} // Brutality Blade
	player := &proto.Player{
		Race:          proto.Race_RaceHuman,
		Class:         proto.Class_ClassRogue,
		Level:         60,
		TalentsString: "--502020331001301",
		Equipment:     &proto.EquipmentSpec{Items: items},
		Rotation:      &proto.APLRotation{},
		Spec:          DefaultRogue,
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{{Level: 63, MobType: proto.MobType_MobTypeHumanoid}}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*DpsRogue)
}

// TestPremeditationOutOfStealth checks that Forever's Premeditation needs no Stealth, and
// that Preparation brings it back.
func TestPremeditationOutOfStealth(t *testing.T) {
	sim, rogue := newSubtletyRogueSim()
	target := rogue.CurrentTarget

	if rogue.IsStealthed() {
		t.Fatalf("the rogue starts in Stealth")
	}
	if !rogue.Premeditation.Cast(sim, target) {
		t.Fatalf("Premeditation did not go out of Stealth")
	}
	if got := rogue.ComboPoints(); got != 2 {
		t.Errorf("Premeditation gave %d combo points, want 2", got)
	}
	if rogue.Premeditation.CD.IsReady(sim) {
		t.Fatalf("Premeditation has no cooldown")
	}
	if !rogue.Preparation.Cast(sim, target) {
		t.Fatalf("Preparation did not go out")
	}
	if !rogue.Premeditation.CD.IsReady(sim) {
		t.Errorf("Preparation did not reset Premeditation")
	}
}
