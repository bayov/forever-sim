package feral

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/druid"
)

func newCatSim(t *testing.T, talents string) (*core.Simulation, *FeralDruid) {
	player := &proto.Player{
		Race:          proto.Race_RaceTauren,
		Class:         proto.Class_ClassDruid,
		Level:         60,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: talents,
		Rotation:      &proto.APLRotation{},
		Spec: &proto.Player_FeralDruid{FeralDruid: &proto.FeralDruid{Options: &proto.FeralDruid_Options{
			InnervateTarget: &proto.UnitReference{},
		}}},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1, ExactDamage: true},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*FeralDruid)
}

// The build the checks below use: Ferocity 5, Shredding Attacks 3, Savage Fury 2, Sharpened
// Claws 2, Shifting Power, Predatory Strikes 3, Blood Frenzy 2, Improved Shifting Power 2,
// Leader of the Pack, Natural Instinct 2, Rend and Tear 5, Berserk, and Furor 5 and Natural
// Shapeshifter 3 in Restoration.
const catTalents = "-50000032021032212051-05003"

// TestCatCosts checks the Energy costs with the talents (Shred 60 less 6 a point of Shredding
// Attacks, Claw and Rake 1 a point of Ferocity off), Shifting Power's mana (55% of base mana,
// 30% off with Natural Shapeshifter 3) and its cooldown (16 sec, 8 off with Improved Shifting
// Power 2), and Tiger's Fury, free with a 30 sec cooldown.
func TestCatCosts(t *testing.T) {
	_, cat := newCatSim(t, catTalents)
	costs := []struct {
		name  string
		spell *druid.DruidSpell
		want  float64
	}{
		{"Shred", cat.Shred, 42},
		{"Claw", cat.Claw, 40},
		{"Rake", cat.Rake, 35},
		{"Rip", cat.Rip, 30},
		{"Ferocious Bite", cat.FerociousBite, 35},
		{"Shifting Power", cat.ShiftingPower, 0.55 * cat.BaseMana * 0.7},
	}
	for _, c := range costs {
		if got := c.spell.DefaultCast.Cost; math.Abs(got-c.want) > 1e-6 {
			t.Errorf("%s costs %.2f, want %.2f", c.name, got, c.want)
		}
	}
	if got := cat.ShiftingPower.CD.Duration; got != 8*time.Second {
		t.Errorf("Shifting Power cooldown %s, want 8s", got)
	}
	// Forever removed Tiger's Fury (sim/druid/tigers_fury.go).
	if cat.TigersFury != nil {
		t.Error("the Forever druid has Tiger's Fury")
	}
}

// TestCatShiftingPower checks that Shifting Power gives 40 Energy.
func TestCatShiftingPower(t *testing.T) {
	sim, cat := newCatSim(t, catTalents)
	cat.SpendEnergy(sim, cat.CurrentEnergy(), cat.Shred.EnergyMetrics())
	cat.SetGCDTimer(sim, sim.CurrentTime)
	if !cat.ShiftingPower.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Shifting Power didn't cast")
	}
	if got := cat.CurrentEnergy(); math.Abs(got-40) > 1e-6 {
		t.Errorf("Energy after Shifting Power %.2f, want 40", got)
	}
}

// TestCatBuffs checks Berserk (100% crit on the builders, not on Rip or Ferocious Bite).
func TestCatBuffs(t *testing.T) {
	sim, cat := newCatSim(t, catTalents)

	shredCrit, biteCrit := cat.Shred.BonusCritRating, cat.FerociousBite.BonusCritRating
	cat.BerserkAura.Activate(sim)
	if got := cat.Shred.BonusCritRating - shredCrit; got != 100 {
		t.Errorf("Berserk gives Shred %.0f crit, want 100", got)
	}
	if got := cat.FerociousBite.BonusCritRating - biteCrit; got != 0 {
		t.Errorf("Berserk gives Ferocious Bite %.0f crit, want 0", got)
	}
}

// TestCatShredDamage checks Shred against the client's effects: (weapon damage + 80) * 155%,
// with no talents, armor or crits in the way.
func TestCatShredDamage(t *testing.T) {
	sim, cat := newCatSim(t, "")
	target := cat.CurrentTarget
	target.AddStatDynamic(sim, stats.Armor, -target.GetStat(stats.Armor))
	ap := cat.Shred.MeleeAttackPower(target)
	weapon := cat.AutoAttacks.MH()
	want := (80 + (weapon.BaseDamageMin+weapon.BaseDamageMax)/2 + ap/14) * 1.55
	result := cat.Shred.CalcDamage(sim, target, 80+weapon.CalculateAverageWeaponDamage(ap), cat.Shred.OutcomeAlwaysHit)
	if math.Abs(result.Damage-want) > 1e-6 {
		t.Errorf("Shred hits for %.2f, want %.2f", result.Damage, want)
	}
}

// TestCatRendAndTear checks Rend and Tear: 2% a point more Shred damage while our Rake is on
// the target, and nothing without a bleed.
func TestCatRendAndTear(t *testing.T) {
	sim, cat := newCatSim(t, catTalents)
	target := cat.CurrentTarget
	if got := cat.RendAndTearMultiplier(target); got != 1 {
		t.Errorf("Rend and Tear with no bleed %.2f, want 1", got)
	}
	cat.Rake.Dot(target).Apply(sim)
	if got := cat.RendAndTearMultiplier(target); math.Abs(got-1.1) > 1e-9 {
		t.Errorf("Rend and Tear 5/5 on a bleeding target %.2f, want 1.1", got)
	}
}

// noCritsNoArmor takes the cat's crits and the target's armor away, for exact damage checks.
func noCritsNoArmor(sim *core.Simulation, cat *FeralDruid) *core.Unit {
	target := cat.CurrentTarget
	target.AddStatDynamic(sim, stats.Armor, -target.GetStat(stats.Armor))
	cat.AddStatDynamic(sim, stats.MeleeCrit, -1000*core.CritRatingPerCritChance)
	return target
}

func nearDamage(got, want float64) bool { return math.Abs(got-want) < 1e-6 }

// tickDamage is what one tick of the dot deals.
func tickDamage(sim *core.Simulation, spell *druid.DruidSpell, target *core.Unit) float64 {
	metrics := &spell.SpellMetrics[target.UnitIndex]
	before := metrics.TotalDamage
	spell.Dot(target).TickOnce(sim)
	return metrics.TotalDamage - before
}

// TestCatRakeTickAttackPower checks Rake's tick under Forever: 34 + 0.055 attack power, the
// attack power read when the tick lands (wowsims/forever's fit to PR #149's tests).
func TestCatRakeTickAttackPower(t *testing.T) {
	sim, cat := newCatSim(t, "")
	target := noCritsNoArmor(sim, cat)
	dot := cat.Rake.Dot(target)
	dot.Apply(sim)
	ap := cat.Rake.MeleeAttackPower(target)
	if got, want := tickDamage(sim, cat.Rake, target), (34+0.055*ap)*dot.SnapshotAttackerMultiplier; !nearDamage(got, want) {
		t.Errorf("Rake ticks for %.3f at %.0f attack power, want %.3f", got, ap, want)
	}
	cat.AddStatDynamic(sim, stats.AttackPower, 1000)
	if got, want := tickDamage(sim, cat.Rake, target), (34+0.055*(ap+1000))*dot.SnapshotAttackerMultiplier; !nearDamage(got, want) {
		t.Errorf("Rake ticks for %.3f after 1000 attack power more, want %.3f", got, want)
	}
}

// TestCatRendAndTearBleeds checks that Rend and Tear 5/5 raises the Rake and Rip ticks by 10%
// (a tester's table on wowsims/forever PR #149).
func TestCatRendAndTearBleeds(t *testing.T) {
	for _, talents := range []string{"", catTalents} {
		sim, cat := newCatSim(t, talents)
		target := noCritsNoArmor(sim, cat)
		rendAndTear := core.TernaryFloat64(talents == "", 1, 1.1)

		rake := cat.Rake.Dot(target)
		rake.Apply(sim)
		ap := cat.Rake.MeleeAttackPower(target)
		want := (34 + 0.055*ap) * cat.Rake.BaseDamageMultiplierAdditive * rake.SnapshotAttackerMultiplier * rendAndTear
		if got := tickDamage(sim, cat.Rake, target); !nearDamage(got, want) {
			t.Errorf("talents %q: Rake ticks for %.3f, want %.3f", talents, got, want)
		}

		cat.AddComboPoints(sim, 5, target, cat.Rip.ComboPointMetrics())
		rip := cat.Rip.Dot(target)
		rip.Apply(sim)
		want = rip.SnapshotBaseDamage * rip.SnapshotAttackerMultiplier * rendAndTear
		if got := tickDamage(sim, cat.Rip, target); !nearDamage(got, want) {
			t.Errorf("talents %q: Rip ticks for %.3f, want %.3f", talents, got, want)
		}
	}
}

// TestCatFinisherMissRefunds checks that a dodged Rip or Ferocious Bite gives back 80% of its
// cost (ATTR_EX_1_DISCOUNT_POWER_ON_MISS on every rank), keeps the combo points, and that Bite
// keeps its extra Energy.
func TestCatFinisherMissRefunds(t *testing.T) {
	for _, spell := range []func(*FeralDruid) *druid.DruidSpell{
		func(cat *FeralDruid) *druid.DruidSpell { return cat.Rip },
		func(cat *FeralDruid) *druid.DruidSpell { return cat.FerociousBite },
	} {
		sim, cat := newCatSim(t, "")
		target := cat.CurrentTarget
		cat.AttackTables[target.UnitIndex][proto.CastType_CastTypeMainHand].BaseDodgeChance = 1
		cat.AddStatDynamic(sim, stats.MeleeHit, 100*core.MeleeHitRatingPerHitChance)
		finisher := spell(cat)
		cat.AddComboPoints(sim, 5, target, finisher.ComboPointMetrics())
		cat.AddEnergy(sim, cat.MaxEnergy(), finisher.EnergyMetrics())
		cat.SetGCDTimer(sim, sim.CurrentTime)
		if !finisher.Cast(sim, target) {
			t.Fatalf("%s didn't cast", finisher.ActionID)
		}
		if got, want := cat.CurrentEnergy(), 100-0.2*finisher.DefaultCast.Cost; math.Abs(got-want) > 1e-6 {
			t.Errorf("%s dodged at 100 Energy leaves %.1f Energy, want %.1f", finisher.ActionID, got, want)
		}
		if cat.ComboPoints() != 5 {
			t.Errorf("%s dodged leaves %d combo points, want 5", finisher.ActionID, cat.ComboPoints())
		}
	}
}
