//go:build with_db

package common_test

import (
	"testing"

	"github.com/wowsims/classic/sim/common"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

// A level 60 character on ruleset wielding weapon in the main hand, against one target with
// debuffs. It returns the sim and the target.
func newWielderVs(t *testing.T, ruleset proto.Ruleset, weapon int32, debuffs *proto.Debuffs) (*core.Simulation, *core.Unit) {
	t.Helper()
	specs := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range specs {
		specs[i] = &proto.ItemSpec{}
	}
	specs[proto.ItemSlot_ItemSlotMainHand].Id = weapon
	player := &proto.Player{
		Race:      proto.Race_RaceGnome,
		Class:     proto.Class_ClassMage,
		Level:     60,
		Equipment: &proto.EquipmentSpec{Items: specs},
		Spec:      &proto.Player_Mage{Mage: &proto.Mage{Options: &proto.Mage_Options{}}},
		Rotation:  &proto.APLRotation{},
	}
	s := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, debuffs),
		Encounter:  &proto.Encounter{Duration: 600, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: ruleset, RandomSeed: 1},
	}, simsignals.CreateSignals())
	s.Reset()
	s.PrePull()
	return s, &s.Encounter.Targets[0].Unit
}

// The armor a full stack of the weapon's debuff (label) takes off the target, on top of what
// the given debuffs already took.
func fullStackArmor(t *testing.T, ruleset proto.Ruleset, weapon int32, label string, debuffs *proto.Debuffs) float64 {
	t.Helper()
	s, target := newWielderVs(t, ruleset, weapon, debuffs)
	aura := target.GetAura(label)
	if aura == nil {
		t.Fatalf("no %s on the target", label)
	}
	before := target.GetStat(stats.Armor)
	aura.Activate(s)
	// A stronger debuff in the slot keeps it from landing at all.
	if aura.IsActive() {
		aura.SetStacks(s, aura.MaxStacks)
	}
	return before - target.GetStat(stats.Armor)
}

// Forever's Armor Shatter and Puncture Armor sit in the minor armor slot and lose it to Faerie
// Fire and Curse of Recklessness, which take more (EV-0147: neither proc lands under either, and
// either landing removes them). Beside Sunder Armor they count in full.
func TestForeverWeaponArmorProcsShareMinorSlot(t *testing.T) {
	forever := proto.Ruleset_RulesetForever
	for _, c := range []struct {
		weapon int32
		label  string
		full   float64
	}{
		{common.Annihlator, "Armor Shatter", 3 * 165},
		{common.Rivenspike, "Puncture Armor", 3 * 100},
	} {
		if got := fullStackArmor(t, forever, c.weapon, c.label, &proto.Debuffs{}); got != c.full {
			t.Errorf("%s alone takes %v armor, want %v", c.label, got, c.full)
		}
		if got := fullStackArmor(t, forever, c.weapon, c.label, &proto.Debuffs{SunderArmor: true}); got != c.full {
			t.Errorf("%s beside Sunder Armor takes %v armor, want %v", c.label, got, c.full)
		}
		if got := fullStackArmor(t, forever, c.weapon, c.label, &proto.Debuffs{FaerieFire: true}); got != 0 {
			t.Errorf("%s under Faerie Fire takes %v more armor, want 0", c.label, got)
		}
		if got := fullStackArmor(t, forever, c.weapon, c.label, &proto.Debuffs{CurseOfRecklessness: true}); got != 0 {
			t.Errorf("%s under Curse of Recklessness takes %v more armor, want 0", c.label, got)
		}
	}
	// Classic keeps its 200 a stack.
	if got := fullStackArmor(t, proto.Ruleset_RulesetClassic, common.Rivenspike, "Puncture Armor", &proto.Debuffs{}); got != 600 {
		t.Errorf("Classic Puncture Armor takes %v armor, want 600", got)
	}
}
