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
	return newForeverRogueSimItems(t, talents, map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotMainHand: mainHandID,
		proto.ItemSlot_ItemSlotOffHand:  offHandID,
	}, numTargets)
}

func newForeverRogueSimItems(t *testing.T, talents string, itemIDs map[proto.ItemSlot]int32, numTargets int) (*core.Simulation, *DpsRogue) {
	return newForeverRogueSimRace(t, proto.Race_RaceHuman, talents, itemIDs, numTargets)
}

func newForeverRogueSimRace(t *testing.T, race proto.Race, talents string, itemIDs map[proto.ItemSlot]int32, numTargets int) (*core.Simulation, *DpsRogue) {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotOffHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{Id: itemIDs[proto.ItemSlot(i)]}
	}
	player := &proto.Player{
		Race:          race,
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

// The Subtlety preset's talents (tools/dps_report/req/rogue__subtlety.json), with Hemorrhage and
// Serrated Blades.
const subtletyPresetTalents = "0053031035140004--3021203310113011"

// Under Forever Rupture's ticks read attack power, the damage modifiers and Hemorrhage when they
// land, and keep only the combo points from the cast (wowsims/forever 36cfc58328, "Periodic damage
// does not snapshot in Forever").
func TestForeverRuptureTicksOnLiveStats(t *testing.T) {
	sim, rogue := newForeverRogueSim(t, subtletyPresetTalents, 18805, 18805, 1) // Core Hound Tooth
	if rogue.HemorrhageAuras == nil {
		t.Fatal("the talents give no Hemorrhage")
	}
	target := rogue.CurrentTarget
	rogue.Rupture.BonusCritRating -= 1000 * core.CritRatingPerCritChance
	rogue.AddComboPoints(sim, 5, target, rogue.Rupture.ComboPointMetrics())
	rogue.AddStatDynamic(sim, stats.MeleeHit, 20*core.MeleeHitRatingPerHitChance)
	for !rogue.Rupture.Dot(target).IsActive() {
		rogue.AddEnergy(sim, 100, rogue.EnergyRefundMetrics)
		rogue.SetGCDTimer(sim, sim.CurrentTime)
		rogue.Rupture.Cast(sim, target)
	}
	dot := rogue.Rupture.Dot(target)
	if dot.NumberOfTicks != 8 {
		t.Fatalf("a 5 point Rupture has %d ticks, want 8", dot.NumberOfTicks)
	}
	metrics := &rogue.Rupture.SpellMetrics[target.UnitIndex]
	tick := func() float64 {
		before := metrics.TotalDamage
		dot.TickOnce(sim)
		return metrics.TotalDamage - before
	}

	for i, extra := range []float64{0, 1000, 0} {
		rogue.AddStatDynamic(sim, stats.AttackPower, extra)
		rogue.PseudoStats.DamageDealtMultiplier *= 1.1
		hemorrhage := 1.0
		if i == 2 {
			rogue.HemorrhageAuras.Get(target).Activate(sim)
			hemorrhage = 1.15
		}
		ap := rogue.Rupture.MeleeAttackPower(target)
		multiplier := rogue.Rupture.AttackerDamageMultiplier(rogue.AttackTables[target.UnitIndex][proto.CastType_CastTypeMainHand], true)
		want := (469.0/8 + 0.03*ap) * hemorrhage * multiplier
		if got := tick(); math.Abs(got-want) >= 1 {
			t.Errorf("Rupture at %.0f attack power ticks for %.3f, want %.3f", ap, got, want)
		}
	}
}

// Under Forever Garrote's ticks read attack power and the damage modifiers when they land
// (wowsims/forever 36cfc58328).
func TestForeverGarroteTicksOnLiveStats(t *testing.T) {
	sim, rogue := newForeverRogueSim(t, "", 18805, 18805, 1)
	target := rogue.CurrentTarget
	dot := rogue.Garrote.Dot(target)
	dot.Apply(sim)
	metrics := &rogue.Garrote.SpellMetrics[target.UnitIndex]
	for _, extra := range []float64{0, 1000} {
		rogue.AddStatDynamic(sim, stats.AttackPower, extra)
		rogue.PseudoStats.DamageDealtMultiplier *= 1.1
		ap := rogue.Garrote.MeleeAttackPower(target)
		multiplier := rogue.Garrote.AttackerDamageMultiplier(rogue.AttackTables[target.UnitIndex][proto.CastType_CastTypeMainHand], true)
		want := (92 + 0.03*ap) * multiplier
		before := metrics.TotalDamage
		dot.TickOnce(sim)
		if got := metrics.TotalDamage - before; math.Abs(got-want) >= 1 {
			t.Errorf("Garrote at %.0f attack power ticks for %.3f, want %.3f", ap, got, want)
		}
	}
}

// Under Forever Ambush needs the rogue behind the target, from Stealth and through Cutthroat alike
// (wowsims/forever ambush.go).
func TestForeverAmbushFromBehind(t *testing.T) {
	sim, rogue := newForeverRogueSim(t, "--302120331011301105", 18805, 18805, 1) // 5 points of Cutthroat
	if rogue.CutthroatAura == nil {
		t.Fatal("the talents give no Cutthroat")
	}
	target := rogue.CurrentTarget
	rogue.CutthroatAura.Activate(sim)
	rogue.AddEnergy(sim, 100, rogue.EnergyRefundMetrics)
	for _, inFront := range []bool{false, true} {
		rogue.PseudoStats.InFrontOfTarget = inFront
		if got := rogue.Ambush.CanCast(sim, target); got == inFront {
			t.Errorf("Ambush through Cutthroat with the rogue in front %v: castable %v", inFront, got)
		}
	}
	rogue.CutthroatAura.Deactivate(sim)
	rogue.StealthAura.Activate(sim)
	if rogue.Ambush.CanCast(sim, target) {
		t.Error("Ambush from Stealth in front of the target is castable")
	}
	rogue.PseudoStats.InFrontOfTarget = false
	if !rogue.Ambush.CanCast(sim, target) {
		t.Error("Ambush from Stealth behind the target isn't castable")
	}
}

// A Night Elf's Shadowmeld (20580) counts as stealth, so Ambush (from behind) and Garrote can
// follow it in combat, and Ambush ends it.
func TestForeverShadowmeldIntoAmbush(t *testing.T) {
	sim, rogue := newForeverRogueSimRace(t, proto.Race_RaceNightElf, "", map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotMainHand: 18805, // Core Hound Tooth
	}, 1)
	target := rogue.CurrentTarget
	rogue.AddEnergy(sim, 100, rogue.EnergyRefundMetrics)
	rogue.SetGCDTimer(sim, sim.CurrentTime)
	if rogue.Ambush.CanCast(sim, target) || rogue.Garrote.CanCast(sim, target) {
		t.Fatal("Ambush or Garrote is castable out of stealth")
	}
	if !rogue.Shadowmeld.Cast(sim, target) {
		t.Fatal("Shadowmeld didn't cast")
	}
	if !rogue.IsStealthed() || !rogue.Ambush.CanCast(sim, target) || !rogue.Garrote.CanCast(sim, target) {
		t.Fatal("Ambush or Garrote isn't castable from Shadowmeld")
	}
	if !rogue.Ambush.Cast(sim, target) || rogue.IsStealthed() {
		t.Error("Ambush didn't cast or left Shadowmeld up")
	}
}

// The client's Grimstitch Armor pieces carry set ID 2099 and no set name, so they count toward
// the set registered with that ID: two pieces give 1% hit.
func TestForeverGrimstitchArmorBySetID(t *testing.T) {
	_, rogue := newForeverRogueSimItems(t, "", map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotHead:     280877, // Grimstitch Mask
		proto.ItemSlot_ItemSlotShoulder: 280876, // Grimstitch Spaulders
		proto.ItemSlot_ItemSlotChest:    280879, // Grimstitch Chestpiece
		proto.ItemSlot_ItemSlotHands:    280875, // Grimstitch Gloves
		proto.ItemSlot_ItemSlotLegs:     280872, // Grimstitch Pants
		proto.ItemSlot_ItemSlotFeet:     280878, // Grimstitch Boots
	}, 1)
	var pieces []int32
	for _, bonus := range rogue.GetActiveSetBonuses() {
		if bonus.Name == "Grimstitch Armor" {
			pieces = append(pieces, bonus.NumPieces)
		}
	}
	if len(pieces) != 4 {
		t.Errorf("six Grimstitch pieces give the bonuses at %v pieces, want 2, 3, 4 and 5", pieces)
	}
	if got := rogue.GetStat(stats.MeleeHit); got != float64(core.MeleeHitRatingPerHitChance) {
		t.Errorf("six Grimstitch pieces give %.2f melee hit rating, want %.2f", got, float64(core.MeleeHitRatingPerHitChance))
	}
}
