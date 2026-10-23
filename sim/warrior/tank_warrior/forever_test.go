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

// newForeverTank builds a level 60 Orc warrior in Defensive Stance with a sword and a shield and
// no rotation.
func newForeverTank(t *testing.T, talents string) (*core.Simulation, *warrior.Warrior) {
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
		TalentsString: talents,
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
	return sim, war
}

// TestForeverShieldWallAndBlock checks the Forever rows (client 70425): Shield Wall takes 60% less
// damage for 12 sec on a 15 min cooldown and the global cooldown (871), less 5.5 min a point of
// Improved Shield Wall (12312), and Shield Block blocks 2 attacks over 7 sec (2565).
func TestForeverShieldWallAndBlock(t *testing.T) {
	sim, war := newForeverTank(t, P1Talents) // Improved Shield Wall 2/2

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

// TestForeverShieldBlockRecastRefills checks that casting Shield Block while its buff is still up
// (7 sec on a 5 sec cooldown) refreshes it with both charges, as wowsims/forever's shield_block.go
// does.
func TestForeverShieldBlockRecastRefills(t *testing.T) {
	sim, war := newForeverTank(t, P1Talents)
	aura := war.ShieldBlockAura
	cast := func() {
		t.Helper()
		war.AddRage(sim, 100, war.RageRefundMetrics)
		if !war.ShieldBlock.Cast(sim, war.CurrentTarget) {
			t.Fatalf("Shield Block didn't cast at %s", sim.CurrentTime)
		}
	}

	cast()
	aura.RemoveStack(sim) // a block spends a charge
	if got := aura.GetStacks(); got != 1 {
		t.Fatalf("Shield Block has %d charges after a block, want 1", got)
	}
	sim.AddPendingAction(&core.PendingAction{NextActionAt: 5 * time.Second, OnAction: func(*core.Simulation) {}})
	for !sim.Step() && sim.CurrentTime < 5*time.Second {
	}
	if !aura.IsActive() {
		t.Fatal("Shield Block ended before its cooldown came back")
	}
	cast()
	if got := aura.GetStacks(); got != 2 {
		t.Errorf("Shield Block recast while up has %d charges, want 2", got)
	}
	if got, want := aura.ExpiresAt(), sim.CurrentTime+7*time.Second; got != want {
		t.Errorf("Shield Block recast at %s ends at %s, want %s", sim.CurrentTime, got, want)
	}
}

// TestForeverShieldSlamIsOffHand checks that under Forever Shield Slam is an off-hand special,
// struck with the shield (wowsims/forever b6ce30c972): main hand procs such as Windfury Totem and
// weapon enchants never see it, Dual Wield Specialization and Furious Precision leave it alone,
// and a procs-per-minute rate on it goes by the main hand's speed.
func TestForeverShieldSlamIsOffHand(t *testing.T) {
	_, bare := newForeverTank(t, "--251533120331221351")
	_, specced := newForeverTank(t, "-00003005-251533120331221351") // Furious Precision 3/3, Dual Wield Specialization 5/5
	slam := specced.ShieldSlam.Spell

	if slam.ProcMask != core.ProcMaskMeleeOHSpecial {
		t.Errorf("Shield Slam's proc mask is %v, want the off-hand special mask", slam.ProcMask)
	}
	if slam.CastType != proto.CastType_CastTypeMainHand {
		t.Errorf("Shield Slam uses the %v attack table, want the main hand's", slam.CastType)
	}
	if bareSlam := bare.ShieldSlam.Spell; slam.DamageMultiplier != bareSlam.DamageMultiplier || slam.BonusHitRating != bareSlam.BonusHitRating {
		t.Errorf("Shield Slam with the off-hand talents: damage %v, hit %v; want the untalented %v and %v",
			slam.DamageMultiplier, slam.BonusHitRating, bareSlam.DamageMultiplier, bareSlam.BonusHitRating)
	}

	mainHandOnly := specced.AutoAttacks.NewPPMManager(6, core.ProcMaskMeleeMH)
	if got := mainHandOnly.Chance(slam.ProcMask); got != 0 {
		t.Errorf("a main hand weapon proc rolls on Shield Slam at %v, want never", got)
	}
	anyMelee := specced.AutoAttacks.NewPPMManager(6, core.ProcMaskMelee)
	if got, want := anyMelee.Chance(slam.ProcMask), specced.AutoAttacks.MH().SwingSpeed*6/60; math.Abs(got-want) > 1e-9 {
		t.Errorf("a 6 PPM melee proc rolls on Shield Slam at %v, want the main hand's %v", got, want)
	}
}
