package dpswarrior

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
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
