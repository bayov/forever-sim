package mage

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

// Mechanics of the Forever talents that a DPS number alone would not catch.

func newMageSim(t *testing.T, talents string) (*core.Simulation, *Mage) {
	player := &proto.Player{
		Race:          proto.Race_RaceGnome,
		Class:         proto.Class_ClassMage,
		Level:         60,
		TalentsString: talents,
		Equipment:     &proto.EquipmentSpec{},
		Spec:          &proto.Player_Mage{Mage: &proto.Mage{Options: &proto.Mage_Options{}}},
		Rotation:      &proto.APLRotation{},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*Mage)
}

func topRank(t *testing.T, name string, ranks []*core.Spell) *core.Spell {
	for i := len(ranks) - 1; i >= 0; i-- {
		if ranks[i] != nil {
			return ranks[i]
		}
	}
	t.Fatalf("no %s rank", name)
	return nil
}

// Heating Up takes 25% off Pyroblast's 6 sec cast per stack, to 1.5 sec at 3 stacks, and the
// next Pyroblast uses up every stack.
func TestMageHeatingUp(t *testing.T) {
	// Fire to Heating Up.
	sim, mage := newMageSim(t, "-235500001301")
	pyroblast := topRank(t, "Pyroblast", mage.Pyroblast)
	for stacks, want := range []time.Duration{6000, 4500, 3000, 1500} {
		if stacks > 0 {
			mage.HeatingUpAura.Activate(sim)
			mage.HeatingUpAura.SetStacks(sim, int32(stacks))
		}
		if got := pyroblast.CastTime(); got != want*time.Millisecond {
			t.Errorf("%d stacks: Pyroblast casts in %s, want %s", stacks, got, want*time.Millisecond)
		}
	}
	mage.SetGCDTimer(sim, sim.CurrentTime)
	if !pyroblast.Cast(sim, mage.CurrentTarget) {
		t.Fatalf("Pyroblast didn't cast")
	}
	for !sim.Step() && mage.Hardcast.Expires > sim.CurrentTime {
	}
	if mage.HeatingUpAura.IsActive() {
		t.Errorf("Heating Up is still up after the Pyroblast")
	}
	if got := pyroblast.CastTime(); got != 6*time.Second {
		t.Errorf("Pyroblast casts in %s after it used the stacks, want 6s", got)
	}
}

// Each Arcane Blast stack is 10% more damage on the other spells and 175% more mana on Arcane
// Blast. Any other damage spell ends them.
func TestMageArcaneBlastStacks(t *testing.T) {
	// Arcane to Arcane Blast and Missile Barrage.
	sim, mage := newMageSim(t, "05022500110001")
	arcaneBlast := topRank(t, "Arcane Blast", mage.ArcaneBlast)
	fireball := topRank(t, "Fireball", mage.Fireball)
	baseCost := arcaneBlast.Cost.GetCurrentCost()
	if want := 0.15 * mage.BaseMana; math.Abs(baseCost-want) > 1e-6 {
		t.Fatalf("Arcane Blast costs %.1f, want %.1f (15%% of base mana)", baseCost, want)
	}
	baseMultiplier := fireball.DamageMultiplierAdditive
	mage.ArcaneBlastAura.Activate(sim)
	mage.ArcaneBlastAura.SetStacks(sim, 3)
	if got, want := arcaneBlast.Cost.GetCurrentCost(), baseCost*(1+3*1.75); math.Abs(got-want) > 1e-6 {
		t.Errorf("Arcane Blast costs %.1f at 3 stacks, want %.1f", got, want)
	}
	if got := fireball.DamageMultiplierAdditive - baseMultiplier; math.Abs(got-0.3) > 1e-9 {
		t.Errorf("Fireball gets %.2f more damage at 3 stacks, want 0.30", got)
	}
	if got := arcaneBlast.DamageMultiplierAdditive; got != 1 {
		t.Errorf("Arcane Blast's own multiplier is %.2f, want 1", got)
	}

	mage.SetGCDTimer(sim, sim.CurrentTime)
	if !fireball.Cast(sim, mage.CurrentTarget) {
		t.Fatalf("Fireball didn't cast")
	}
	for !sim.Step() && mage.Hardcast.Expires > sim.CurrentTime {
	}
	if mage.ArcaneBlastAura.IsActive() {
		t.Errorf("the Arcane Blast stacks are still up after a Fireball")
	}
	if got := fireball.DamageMultiplierAdditive; math.Abs(got-baseMultiplier) > 1e-9 {
		t.Errorf("Fireball's multiplier is %.2f after the stacks ended, want %.2f", got, baseMultiplier)
	}
}

// Fire Vulnerability counts only for the mage's own Fire damage, 3% a stack.
func TestMageFireVulnerability(t *testing.T) {
	sim, mage := newMageSim(t, "-235500001301")
	target := mage.CurrentTarget
	aura := mage.ImprovedScorchAuras.Get(target)
	aura.Activate(sim)
	aura.SetStacks(sim, 5)

	at := mage.AttackTables[target.UnitIndex][proto.CastType_CastTypeMainHand]
	fireball := topRank(t, "Fireball", mage.Fireball)
	frostbolt := topRank(t, "Frostbolt", mage.Frostbolt)
	if got := fireball.TargetDamageMultiplier(at, false); math.Abs(got-1.15) > 1e-9 {
		t.Errorf("Fireball takes %.3f at 5 stacks, want 1.15", got)
	}
	if got := frostbolt.TargetDamageMultiplier(at, false); got != 1 {
		t.Errorf("Frostbolt takes %.3f, want 1", got)
	}
	if got := target.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexFire]; got != 1 {
		t.Errorf("the target takes %.3f more Fire damage from everyone, want 1", got)
	}
}

// Winter's Chill stacks up to its rank, and gives only Frostbolt and Ice Lance its crit.
// Shatter's crit comes with Fingers of Frost.
func TestMageWintersChillAndShatter(t *testing.T) {
	// Frost to Winter's Chill 2, with Shatter 3 and Fingers of Frost 2.
	sim, mage := newMageSim(t, "--055500033100130122")
	target := mage.CurrentTarget
	aura := mage.WintersChillAuras.Get(target)
	if aura.MaxStacks != 2 {
		t.Fatalf("Winter's Chill stacks to %d at rank 2, want 2", aura.MaxStacks)
	}
	aura.Activate(sim)
	aura.SetStacks(sim, 2)

	frostbolt := topRank(t, "Frostbolt", mage.Frostbolt)
	iceLance := topRank(t, "Ice Lance", mage.IceLance)
	blizzard := topRank(t, "Blizzard", mage.Blizzard)
	rating := float64(core.SpellCritRatingPerCritChance)
	if got := mage.personalCritRating(frostbolt, target); got != 4*rating {
		t.Errorf("Frostbolt gets %.1f crit rating from Winter's Chill, want %.1f", got, 4*rating)
	}
	if got := mage.personalCritRating(blizzard, target); got != 0 {
		t.Errorf("Blizzard gets %.1f crit rating from Winter's Chill, want 0", got)
	}

	mage.FingersOfFrostAura.Activate(sim)
	mage.FingersOfFrostAura.SetStacks(sim, 2)
	if got := mage.personalCritRating(iceLance, target); got != 54*rating {
		t.Errorf("Ice Lance gets %.1f crit rating with Fingers of Frost, want %.1f", got, 54*rating)
	}
	if got := mage.personalCritRating(blizzard, target); got != 50*rating {
		t.Errorf("Blizzard gets %.1f crit rating with Fingers of Frost, want %.1f", got, 50*rating)
	}
}
