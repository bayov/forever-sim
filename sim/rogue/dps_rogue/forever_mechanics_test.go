package dpsrogue

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

func newForeverRogueSim(t *testing.T, talents string, mainHandID, offHandID int32, numTargets int) (*core.Simulation, *DpsRogue) {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotOffHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: mainHandID}
	items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: offHandID}
	player := &proto.Player{
		Race:          proto.Race_RaceHuman,
		Class:         proto.Class_ClassRogue,
		Level:         60,
		TalentsString: talents,
		Equipment:     &proto.EquipmentSpec{Items: items},
		Rotation:      &proto.APLRotation{},
		Spec:          DefaultRogue,
	}
	targets := make([]*proto.Target, numTargets)
	for i := range targets {
		targets[i] = &proto.Target{Level: 63, MobType: proto.MobType_MobTypeHumanoid}
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 600, Targets: targets},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*DpsRogue)
}

// Under Forever Blade Flurry copies the first hit's final damage to the second target, so the
// damage multipliers that went into the hit don't apply a second time (wowsims/forever's
// talents_combat.go copies result.Damage with IgnoreResists and IgnoreModifiers).
func TestForeverBladeFlurryCopiesFinalDamage(t *testing.T) {
	sim, rogue := newForeverRogueSim(t, CombatSwordsTalents, 18832, 18805, 2) // Brutality Blade, Core Hound Tooth
	if rogue.BladeFlurryAura == nil {
		t.Fatal("the talents give no Blade Flurry")
	}
	first, second := sim.Encounter.TargetUnits[0], sim.Encounter.TargetUnits[1]
	rogue.PseudoStats.DamageDealtMultiplier *= 1.5
	rogue.BladeFlurryAura.Activate(sim)
	for i := 0; i < 50; i++ {
		rogue.AddEnergy(sim, 100, rogue.EnergyRefundMetrics)
		rogue.SetGCDTimer(sim, sim.CurrentTime)
		rogue.SinisterStrike.Cast(sim, first)
	}

	var meleeDamage float64
	for _, spell := range rogue.Spellbook {
		if spell.ProcMask.Matches(core.ProcMaskMelee) {
			meleeDamage += spell.SpellMetrics[first.UnitIndex].TotalDamage
		}
	}
	copied := rogue.GetSpell(core.ActionID{SpellID: 22482}).SpellMetrics[second.UnitIndex].TotalDamage
	if meleeDamage == 0 || math.Abs(copied-meleeDamage) > 1e-6*meleeDamage {
		t.Errorf("Blade Flurry copied %.1f damage of %.1f", copied, meleeDamage)
	}
}

// Every Forever finisher carries ATTR_EX_1_DISCOUNT_POWER_ON_MISS, so a missed Eviscerate gives
// back 80% of its Energy and keeps its combo points. No talents, so nothing else gives Energy or
// combo points back.
func TestForeverFinisherMissRefund(t *testing.T) {
	sim, rogue := newForeverRogueSim(t, "", 18832, 18805, 1)
	target := rogue.CurrentTarget
	misses, hits := 0, 0
	for i := 0; i < 400 && misses < 5; i++ {
		rogue.SpendComboPoints(sim, rogue.Eviscerate)
		rogue.AddComboPoints(sim, 3, target, rogue.Eviscerate.ComboPointMetrics())
		rogue.AddEnergy(sim, 100, rogue.EnergyRefundMetrics)
		rogue.SetGCDTimer(sim, sim.CurrentTime)
		energy := rogue.CurrentEnergy()
		if !rogue.Eviscerate.Cast(sim, target) {
			t.Fatalf("Eviscerate didn't cast at %s", sim.CurrentTime)
		}
		cost, spent := rogue.Eviscerate.CurCast.Cost, energy-rogue.CurrentEnergy()
		if rogue.ComboPoints() == 0 {
			hits++
			if math.Abs(spent-cost) > 1e-6 {
				t.Errorf("a landed Eviscerate cost %.2f Energy, want %.0f", spent, cost)
			}
			continue
		}
		misses++
		if rogue.ComboPoints() != 3 {
			t.Errorf("a missed Eviscerate left %d combo points, want 3", rogue.ComboPoints())
		}
		if math.Abs(spent-0.2*cost) > 1e-6 {
			t.Errorf("a missed Eviscerate cost %.2f Energy, want %.2f (20%% of %.0f)", spent, 0.2*cost, cost)
		}
	}
	if misses == 0 || hits == 0 {
		t.Fatalf("%d Eviscerates landed and %d missed", hits, misses)
	}
}

// Forever's Instant Poison adds 0.5% of attack power a hit (wowsims/forever, "Confirmed by Zirene").
// The same seed gives the same rolls, so 2000 more attack power adds 10 to every hit before the
// multipliers, and half as much again on a crit.
func TestForeverInstantPoisonAttackPower(t *testing.T) {
	const casts = 500
	damage := func(bonusAP float64) (float64, *core.Spell) {
		sim, rogue := newForeverRogueSim(t, "", 18832, 18805, 1)
		rogue.AddStatDynamic(sim, stats.AttackPower, bonusAP)
		for i := 0; i < casts; i++ {
			rogue.InstantPoison.Cast(sim, rogue.CurrentTarget)
		}
		return rogue.InstantPoison.SpellMetrics[rogue.CurrentTarget.UnitIndex].TotalDamage, rogue.InstantPoison
	}
	base, _ := damage(0)
	more, poison := damage(2000)
	metrics := poison.SpellMetrics[poison.Unit.CurrentTarget.UnitIndex]
	want := 0.005 * 2000 * poison.DamageMultiplier * (float64(metrics.Hits) + 1.5*float64(metrics.Crits))
	if got := more - base; math.Abs(got-want) > 0.01*want {
		t.Errorf("2000 attack power added %.1f Instant Poison damage over %d casts, want %.1f", got, casts, want)
	}
}
