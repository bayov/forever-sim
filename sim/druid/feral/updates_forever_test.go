package feral

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/druid"
)

// Changes from the Forever beta development notes that the client data confirms.

// "Faerie Fire no longer resets your swing timer" (2026-10-01, build 1.60.1.70170): the client
// dropped InterruptFlags 8, the auto attack reset, from the caster ranks, and Faerie Fire
// (Feral) has none. Classic keeps the reset.
func TestForeverFaerieFireKeepsSwing(t *testing.T) {
	_, cat := newFormSim(t, proto.Ruleset_RulesetForever, &proto.Consumes{})
	if cat.FaerieFire.Flags.Matches(core.SpellFlagResetAttackSwing) {
		t.Error("Forever Faerie Fire (Feral) resets the swing timer")
	}
	_, era := newFormSim(t, proto.Ruleset_RulesetClassic, &proto.Consumes{})
	if !era.FaerieFire.Flags.Matches(core.SpellFlagResetAttackSwing) {
		t.Error("Classic Faerie Fire (Feral) no longer resets the swing timer")
	}
}

// "Shifting Power can no longer be used while already at Full Energy" (2026-10-08, build
// 1.60.1.70291).
func TestForeverShiftingPowerNotAtFullEnergy(t *testing.T) {
	sim, cat := newCatSim(t, catTalents)
	cat.AddEnergy(sim, cat.MaxEnergy(), cat.Shred.EnergyMetrics())
	cat.SetGCDTimer(sim, sim.CurrentTime)
	if cat.ShiftingPower.Cast(sim, cat.CurrentTarget) {
		t.Error("Shifting Power cast at full Energy")
	}
	cat.SpendEnergy(sim, 1, cat.Shred.EnergyMetrics())
	if !cat.ShiftingPower.Cast(sim, cat.CurrentTarget) {
		t.Error("Shifting Power didn't cast at 99 Energy")
	}
}

// Wolfshead Helm (17768): "an additional 5 Energy from activating Shifting Power" (spell
// 1310990; 20 in build 1.60.1.70170, 5 since 70291, development notes 2026-10-08), and nothing
// on the shift into Cat Form.
func TestForeverWolfsheadHelm(t *testing.T) {
	player := &proto.Player{
		Race:          proto.Race_RaceTauren,
		Class:         proto.Class_ClassDruid,
		Level:         60,
		Equipment:     &proto.EquipmentSpec{Items: []*proto.ItemSpec{{Id: druid.WolfsheadHelm}}},
		TalentsString: catTalents,
		Rotation:      &proto.APLRotation{},
		Spec: &proto.Player_FeralDruid{FeralDruid: &proto.FeralDruid{Options: &proto.FeralDruid_Options{
			InnervateTarget: &proto.UnitReference{},
		}}},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	cat := sim.Raid.Parties[0].Players[0].(*FeralDruid)

	cat.SpendEnergy(sim, cat.CurrentEnergy(), cat.Shred.EnergyMetrics())
	cat.SetGCDTimer(sim, sim.CurrentTime)
	if !cat.ShiftingPower.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Shifting Power didn't cast")
	}
	if got := cat.CurrentEnergy(); math.Abs(got-45) > 1e-6 {
		t.Errorf("Energy after Shifting Power with Wolfshead Helm %.2f, want 45", got)
	}
}
