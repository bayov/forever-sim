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
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
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
	if cat.TigersFury.DefaultCast.Cost != 0 || cat.TigersFury.CD.Duration != 30*time.Second {
		t.Errorf("Tiger's Fury costs %.0f with a %s cooldown, want free and 30s", cat.TigersFury.DefaultCast.Cost, cat.TigersFury.CD.Duration)
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

// TestCatBuffs checks Tiger's Fury (15% Physical damage) and Berserk (100% crit on the
// builders, not on Rip or Ferocious Bite).
func TestCatBuffs(t *testing.T) {
	sim, cat := newCatSim(t, catTalents)
	physical := cat.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical]
	cat.TigersFuryAura.Activate(sim)
	if got := cat.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical]; math.Abs(got-physical*1.15) > 1e-9 {
		t.Errorf("Physical damage with Tiger's Fury %.4f, want %.4f", got, physical*1.15)
	}
	cat.TigersFuryAura.Deactivate(sim)

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
