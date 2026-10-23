package tankwarrior

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/warrior"
)

// TestForeverShieldWallAndBlock checks the Forever rows (client 70425): Shield Wall takes 60% less
// damage for 12 sec on a 15 min cooldown and the global cooldown (871), less 5.5 min a point of
// Improved Shield Wall (12312), and Shield Block blocks 2 attacks over 7 sec (2565).
func TestForeverShieldWallAndBlock(t *testing.T) {
	items := make([]*proto.ItemSpec, 17)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 991014}
	items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: 19349} // Elementium Reinforced Bulwark
	player := &proto.Player{
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassWarrior,
		Level:         60,
		Equipment:     &proto.EquipmentSpec{Items: items},
		TalentsString: P1Talents, // Improved Shield Wall 2/2
		Rotation:      &proto.APLRotation{},
		Spec:          PlayerOptionsBasic,
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	war := sim.Raid.Parties[0].Players[0].(*TankWarrior).GetWarrior()
	war.Stance = warrior.DefensiveStance

	wall := war.GetSpell(core.ActionID{SpellID: 871})
	if got, want := wall.CD.Duration, 4*time.Minute; got != want {
		t.Errorf("Shield Wall with Improved Shield Wall 2/2 has a %s cooldown, want %s", got, want)
	}
	if got := wall.DefaultCast.GCD; got != core.GCDDefault {
		t.Errorf("Shield Wall has a %s global cooldown, want %s", got, core.GCDDefault)
	}
	taken := war.PseudoStats.DamageTakenMultiplier
	if !wall.Cast(sim, war.CurrentTarget) {
		t.Fatal("Shield Wall didn't cast in Defensive Stance with a shield")
	}
	aura := war.GetAura("Shield Wall")
	if aura.Duration != 12*time.Second {
		t.Errorf("Shield Wall lasts %s, want 12s", aura.Duration)
	}
	if got := war.PseudoStats.DamageTakenMultiplier / taken; math.Abs(got-0.4) > 1e-9 {
		t.Errorf("Shield Wall leaves %.2f of the damage taken, want 0.4", got)
	}

	if got := war.ShieldBlockAura.Duration; got != 7*time.Second {
		t.Errorf("Shield Block lasts %s, want 7s", got)
	}
	if got := war.ShieldBlockAura.MaxStacks; got != 2 {
		t.Errorf("Shield Block blocks %d attacks, want 2", got)
	}
	if got := war.ShieldBlock.CD.Duration; got != 5*time.Second {
		t.Errorf("Shield Block has a %s cooldown, want 5s", got)
	}
}
