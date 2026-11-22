//go:build with_db

package common_test

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/common"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/mage"
)

// A level 60 Forever mage wearing item in slot and nothing else.
func newForeverMageWearing(t *testing.T, slot proto.ItemSlot, item int32) (*core.Simulation, *mage.Mage) {
	t.Helper()
	specs := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range specs {
		specs[i] = &proto.ItemSpec{}
	}
	specs[slot].Id = item
	player := &proto.Player{
		Race:      proto.Race_RaceGnome,
		Class:     proto.Class_ClassMage,
		Level:     60,
		Equipment: &proto.EquipmentSpec{Items: specs},
		Spec:      &proto.Player_Mage{Mage: &proto.Mage{Options: &proto.Mage_Options{}}},
		Rotation:  &proto.APLRotation{},
	}
	s := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 100000, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	s.Reset()
	s.PrePull()
	return s, s.Raid.Parties[0].Players[0].(*mage.Mage)
}

// Helm of Fire: Fireball with 33 every 2 sec for 8 sec, on a 5 min cooldown and 10 sec of the
// shared burst item cooldown under Forever.
func TestForeverHelmOfFire(t *testing.T) {
	s, m := newForeverMageWearing(t, proto.ItemSlot_ItemSlotHead, common.HelmOfFire)
	spell := m.GetSpell(core.ActionID{ItemID: common.HelmOfFire})
	if spell == nil {
		t.Fatal("no Helm of Fire use")
	}
	if got := spell.CD.Duration; got != 5*time.Minute {
		t.Errorf("Helm of Fire cooldown %v, want 5m", got)
	}
	if got := spell.SharedCD.Duration; got != 10*time.Second {
		t.Errorf("Helm of Fire shared cooldown %v, want 10s", got)
	}
	dot := spell.Dot(m.CurrentTarget)
	if dot.NumberOfTicks != 4 || dot.TickLength != 2*time.Second {
		t.Errorf("Helm of Fire burns %d ticks of %v, want 4 of 2s", dot.NumberOfTicks, dot.TickLength)
	}
	for !spell.Cast(s, m.CurrentTarget) || !dot.IsActive() {
		// A resisted Fireball puts no burn on. The cooldown lets us try again later.
		s.CurrentTime += 5 * time.Minute
	}
	if dot.SnapshotBaseDamage != 33 {
		t.Errorf("Helm of Fire burns for %.0f a tick, want 33", dot.SnapshotBaseDamage)
	}
	if spell.Cast(s, m.CurrentTarget) {
		t.Error("Helm of Fire cast again while on cooldown")
	}
}

// Circle of Flame: a 10 sec channel giving 79 mana a second with a variance of 0.5, 59 to 99.
func TestForeverCircleOfFlame(t *testing.T) {
	s, m := newForeverMageWearing(t, proto.ItemSlot_ItemSlotHead, common.CircleOfFlame)
	spell := m.GetSpell(core.ActionID{SpellID: 17447})
	if spell == nil {
		t.Fatal("no Circle of Flame use")
	}
	if !spell.Flags.Matches(core.SpellFlagChanneled) {
		t.Error("Circle of Flame isn't a channel")
	}
	if got := spell.CD.Duration; got != 5*time.Minute {
		t.Errorf("Circle of Flame cooldown %v, want 5m", got)
	}
	hot := spell.SelfHot()
	if hot.NumberOfTicks != 10 || hot.TickLength != time.Second {
		t.Errorf("Circle of Flame has %d ticks of %v, want 10 of 1s", hot.NumberOfTicks, hot.TickLength)
	}
	lo, hi, sum := 1e9, 0.0, 0.0
	const n = 2000
	for range n {
		m.SpendMana(s, m.CurrentMana(), m.NewManaMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionNone}))
		hot.OnTick(s, &m.Unit, hot)
		got := m.CurrentMana()
		lo, hi, sum = min(lo, got), max(hi, got), sum+got
	}
	// Mana comes in whole points, so 59.25 to 98.75 rounds to 59 to 99.
	if lo < 59 || hi > 99 {
		t.Errorf("Circle of Flame gives %.1f to %.1f a tick, want within 59 to 99", lo, hi)
	}
	if avg := sum / n; avg < 77 || avg > 81 {
		t.Errorf("Circle of Flame gives %.1f a tick on average, want about 79", avg)
	}
}

// Robe of the Archmage: 375 to 625 mana, off the GCD, on a 5 min cooldown of its own.
func TestForeverRobeOfTheArchmage(t *testing.T) {
	s, m := newForeverMageWearing(t, proto.ItemSlot_ItemSlotChest, common.RobeOfTheArchmage)
	spell := m.GetSpell(core.ActionID{SpellID: 18385})
	if spell == nil {
		t.Fatal("no Replenish Mana use")
	}
	if spell.DefaultCast.GCD != 0 {
		t.Error("Replenish Mana is on the GCD")
	}
	drain := m.NewManaMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionNone})
	lo, hi, sum := 1e9, 0.0, 0.0
	const n = 2000
	for range n {
		s.CurrentTime += 5 * time.Minute
		m.SpendMana(s, m.CurrentMana(), drain)
		if !spell.Cast(s, &m.Unit) {
			t.Fatal("Replenish Mana didn't cast off cooldown")
		}
		got := m.CurrentMana()
		lo, hi, sum = min(lo, got), max(hi, got), sum+got
	}
	if lo < 375 || hi > 625 {
		t.Errorf("Replenish Mana gives %.1f to %.1f, want within 375 to 625", lo, hi)
	}
	if avg := sum / n; avg < 490 || avg > 510 {
		t.Errorf("Replenish Mana gives %.1f on average, want about 500", avg)
	}
	if spell.Cast(s, &m.Unit) {
		t.Error("Replenish Mana cast again while on cooldown")
	}
}
