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

// The build the checks below use, on the trees of the launch realms' 2026-10-30 hotfix: Ferocity
// 5, Furor 5, Shredding Attacks 3, Savage Fury 2, Sharpened Claws 2, Predatory Strikes 3, Predatory
// Discipline, Blood Frenzy 2, Leader of the Pack, Natural Instinct 2, Rend and Tear 5 and Shifting
// Power, and Heart of the Wild 5 and Natural Shapeshifter 3 in Restoration.
const catTalents = "-55000032023012125001-05003"

// TestCatCosts checks the Energy costs with the talents (Shred 60 less 6 a point of Shredding
// Attacks, Claw and Rake 1 a point of Ferocity off), Shifting Power's mana (55% of base mana,
// 30% off with Natural Shapeshifter 3) and its 8 sec cooldown (the hotfix's; it took Improved
// Shifting Power out of the tree), and that the Forever druid has no Tiger's Fury.
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

// TestCatPredatoryDiscipline checks Predatory Discipline (1328613) on the literal reading of its
// tooltip (the report's druid_predatory_discipline assumption): a Bite at 100 Energy deals the same
// damage with the talent as without it, extra Energy included, and keeps the 65 extra Energy, where
// without it the Bite spends all 100. The two builds differ only in that point (Lacerate, a bear
// ability the cat never uses, takes it in the other), so the fights roll the same.
func TestCatPredatoryDiscipline(t *testing.T) {
	const withoutPD = "-55000032000100000000" // Ferocity 5, Furor 5, Shredding Attacks 3, Savage Fury 2, Lacerate
	const withPD = "-55000032000010000000"    // the same with Predatory Discipline for Lacerate
	bite := func(talents string) (damage, energyLeft float64) {
		sim, cat := newCatSim(t, talents)
		target := noCritsNoArmor(sim, cat)
		cat.AddStatDynamic(sim, stats.MeleeHit, 100*core.MeleeHitRatingPerHitChance)
		cat.AttackTables[target.UnitIndex][proto.CastType_CastTypeMainHand].BaseDodgeChance = 0
		cat.AttackTables[target.UnitIndex][proto.CastType_CastTypeMainHand].BaseParryChance = 0
		cat.AttackTables[target.UnitIndex][proto.CastType_CastTypeMainHand].BaseGlanceChance = 0
		spell := cat.FerociousBite
		cat.AddComboPoints(sim, 5, target, spell.ComboPointMetrics())
		cat.AddEnergy(sim, cat.MaxEnergy(), spell.EnergyMetrics())
		cat.SetGCDTimer(sim, sim.CurrentTime)
		if !spell.Cast(sim, target) {
			t.Fatalf("%q: Bite didn't cast", talents)
		}
		return spell.SpellMetrics[target.UnitIndex].TotalDamage, cat.CurrentEnergy()
	}
	plainDamage, plainEnergy := bite(withoutPD)
	pdDamage, pdEnergy := bite(withPD)
	if plainDamage <= 0 || math.Abs(pdDamage-plainDamage) > 1e-6 {
		t.Errorf("Bite at 100 Energy deals %.2f with Predatory Discipline, %.2f without, want the same", pdDamage, plainDamage)
	}
	if math.Abs(plainEnergy) > 1e-6 {
		t.Errorf("Bite at 100 Energy without Predatory Discipline leaves %.1f Energy, want 0", plainEnergy)
	}
	if math.Abs(pdEnergy-65) > 1e-6 {
		t.Errorf("Bite at 100 Energy with Predatory Discipline leaves %.1f Energy, want 65", pdEnergy)
	}
}

// TestCatBiteRanks checks the hotfix's new Ferocious Bite rank 1 (1328618) at 24, that the old
// rank 1 (22568) still comes at 32, and that level 60 casts 22829 (31018 is an Ahn'Qiraj book
// rank). Classic has no rank at 24.
func TestCatBiteRanks(t *testing.T) {
	for _, tc := range []struct {
		ruleset proto.Ruleset
		level   int32
		id      int32
	}{
		{proto.Ruleset_RulesetForever, 23, 0},
		{proto.Ruleset_RulesetForever, 24, 1328618},
		{proto.Ruleset_RulesetForever, 31, 1328618},
		{proto.Ruleset_RulesetForever, 32, 22568},
		{proto.Ruleset_RulesetForever, 60, 22829},
		{proto.Ruleset_RulesetClassic, 24, 0},
		{proto.Ruleset_RulesetClassic, 32, 22568},
	} {
		player := &proto.Player{
			Race: proto.Race_RaceTauren, Class: proto.Class_ClassDruid, Level: tc.level,
			Equipment: &proto.EquipmentSpec{}, Rotation: &proto.APLRotation{},
			Spec: &proto.Player_FeralDruid{FeralDruid: &proto.FeralDruid{Options: &proto.FeralDruid_Options{
				InnervateTarget: &proto.UnitReference{},
			}}},
		}
		sim := core.NewSim(&proto.RaidSimRequest{
			Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
			SimOptions: &proto.SimOptions{Ruleset: tc.ruleset, RandomSeed: 1},
		}, simsignals.CreateSignals())
		cat := sim.Raid.Parties[0].Players[0].(*FeralDruid)
		got := int32(0)
		if cat.FerociousBite != nil {
			got = cat.FerociousBite.ActionID.SpellID
		}
		if got != tc.id {
			t.Errorf("%s level %d: Ferocious Bite is %d, want %d", tc.ruleset, tc.level, got, tc.id)
		}
	}
}

// TestCatRavageDamage checks Ravage against the client's effects (9867) at the November 2
// development notes' 435%: (weapon damage + 98) * 435%, with no talents, armor or crits in the way.
func TestCatRavageDamage(t *testing.T) {
	sim, cat := newCatSim(t, "")
	target := cat.CurrentTarget
	target.AddStatDynamic(sim, stats.Armor, -target.GetStat(stats.Armor))
	ap := cat.Ravage.MeleeAttackPower(target)
	weapon := cat.AutoAttacks.MH()
	want := (98 + (weapon.BaseDamageMin+weapon.BaseDamageMax)/2 + ap/14) * 4.35
	result := cat.Ravage.CalcDamage(sim, target, 98+weapon.CalculateAverageWeaponDamage(ap), cat.Ravage.OutcomeAlwaysHit)
	if math.Abs(result.Damage-want) > 1e-6 {
		t.Errorf("Ravage hits for %.2f, want %.2f", result.Damage, want)
	}
}

// TestCatShredDamage checks Shred against the client's effects at the November 2 development
// notes' 200%: (weapon damage + 80) * 200%,
// with no talents, armor or crits in the way.
func TestCatShredDamage(t *testing.T) {
	sim, cat := newCatSim(t, "")
	target := cat.CurrentTarget
	target.AddStatDynamic(sim, stats.Armor, -target.GetStat(stats.Armor))
	ap := cat.Shred.MeleeAttackPower(target)
	weapon := cat.AutoAttacks.MH()
	want := (80 + (weapon.BaseDamageMin+weapon.BaseDamageMax)/2 + ap/14) * 2.0
	result := cat.Shred.CalcDamage(sim, target, 80+weapon.CalculateAverageWeaponDamage(ap), cat.Shred.OutcomeAlwaysHit)
	if math.Abs(result.Damage-want) > 1e-6 {
		t.Errorf("Shred hits for %.2f, want %.2f", result.Damage, want)
	}
}

// TestCatRendAndTear checks Rend and Tear: 3% a point (the November 2 development notes) more Shred damage while our Rake is on
// the target, and nothing without a bleed.
func TestCatRendAndTear(t *testing.T) {
	sim, cat := newCatSim(t, catTalents)
	target := cat.CurrentTarget
	if got := cat.RendAndTearMultiplier(target); got != 1 {
		t.Errorf("Rend and Tear with no bleed %.2f, want 1", got)
	}
	cat.Rake.Dot(target).Apply(sim)
	if got := cat.RendAndTearMultiplier(target); math.Abs(got-1.15) > 1e-9 {
		t.Errorf("Rend and Tear 5/5 on a bleeding target %.2f, want 1.15", got)
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
// attack power read when the tick lands (wowsims/forever's fit to PR #149's tests), both 25% up
// with the November 2 development notes.
func TestCatRakeTickAttackPower(t *testing.T) {
	sim, cat := newCatSim(t, "")
	target := noCritsNoArmor(sim, cat)
	dot := cat.Rake.Dot(target)
	dot.Apply(sim)
	ap := cat.Rake.MeleeAttackPower(target)
	if got, want := tickDamage(sim, cat.Rake, target), 1.25*(34+0.055*ap)*dot.SnapshotAttackerMultiplier; !nearDamage(got, want) {
		t.Errorf("Rake ticks for %.3f at %.0f attack power, want %.3f", got, ap, want)
	}
	cat.AddStatDynamic(sim, stats.AttackPower, 1000)
	if got, want := tickDamage(sim, cat.Rake, target), 1.25*(34+0.055*(ap+1000))*dot.SnapshotAttackerMultiplier; !nearDamage(got, want) {
		t.Errorf("Rake ticks for %.3f after 1000 attack power more, want %.3f", got, want)
	}
}

// TestCatRakeHitAttackPower checks Rake's direct hit under Forever: 61 at the November 2 scale
// plus 0.065 attack power read at the hit, which the scale leaves alone (the user's ruling,
// 2026-11-12, on wowsims/forever 0384d8f98c's combat log fit).
func TestCatRakeHitAttackPower(t *testing.T) {
	sim, cat := newCatSim(t, "")
	target := noCritsNoArmor(sim, cat)
	cat.AddStatDynamic(sim, stats.MeleeHit, 100*core.MeleeHitRatingPerHitChance)
	target.PseudoStats.DodgeReduction += 1
	// The multiplier the hit carries, on a base of 1.
	multiplier := cat.Rake.CalcDamage(sim, target, 1, cat.Rake.OutcomeAlwaysHit).Damage
	hit := func() float64 {
		metrics := &cat.Rake.SpellMetrics[target.UnitIndex]
		before, hits := metrics.TotalDamage, metrics.Hits
		cat.Rake.ApplyEffects(sim, target, cat.Rake.Spell)
		if metrics.Hits != hits+1 {
			t.Fatal("Rake's hit didn't land")
		}
		return metrics.TotalDamage - before
	}
	ap := cat.Rake.MeleeAttackPower(target)
	low := hit()
	if want := (1.25*61 + 0.065*ap) * multiplier; !nearDamage(low, want) {
		t.Errorf("Rake hits for %.3f at %.0f attack power, want %.3f", low, ap, want)
	}
	cat.AddStatDynamic(sim, stats.AttackPower, 1000)
	if got, want := hit()-low, 0.065*(cat.Rake.MeleeAttackPower(target)-ap)*multiplier; !nearDamage(got, want) {
		t.Errorf("1000 attack power more raises Rake's hit by %.3f, want %.3f", got, want)
	}
}

// TestCatRendAndTearBleeds checks that Rend and Tear 5/5 raises the Rake and Rip ticks by 15%
// (a tester's table on wowsims/forever PR #149, at the November 2 development notes' 3% a point).
func TestCatRendAndTearBleeds(t *testing.T) {
	for _, talents := range []string{"", catTalents} {
		sim, cat := newCatSim(t, talents)
		target := noCritsNoArmor(sim, cat)
		rendAndTear := core.TernaryFloat64(talents == "", 1, 1.15)

		rake := cat.Rake.Dot(target)
		rake.Apply(sim)
		ap := cat.Rake.MeleeAttackPower(target)
		want := 1.25 * (34 + 0.055*ap) * cat.Rake.BaseDamageMultiplierAdditive * rake.SnapshotAttackerMultiplier * rendAndTear
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

// TestCatFinisherMissRefunds checks that a dodged Rip gives back 80% of its cost
// (ATTR_EX_1_DISCOUNT_POWER_ON_MISS), that a dodged Ferocious Bite spends its cost and keeps the
// extra Energy (wowsims/forever c91498c01), and that both keep the combo points.
func TestCatFinisherMissRefunds(t *testing.T) {
	for _, tc := range []struct {
		spell      func(*FeralDruid) *druid.DruidSpell
		energyLeft func(cost float64) float64
	}{
		{func(cat *FeralDruid) *druid.DruidSpell { return cat.Rip }, func(cost float64) float64 { return 100 - 0.2*cost }},
		{func(cat *FeralDruid) *druid.DruidSpell { return cat.FerociousBite }, func(cost float64) float64 { return 100 - cost }},
	} {
		sim, cat := newCatSim(t, "")
		target := cat.CurrentTarget
		cat.AttackTables[target.UnitIndex][proto.CastType_CastTypeMainHand].BaseDodgeChance = 1
		cat.AddStatDynamic(sim, stats.MeleeHit, 100*core.MeleeHitRatingPerHitChance)
		finisher := tc.spell(cat)
		cat.AddComboPoints(sim, 5, target, finisher.ComboPointMetrics())
		cat.AddEnergy(sim, cat.MaxEnergy(), finisher.EnergyMetrics())
		cat.SetGCDTimer(sim, sim.CurrentTime)
		if !finisher.Cast(sim, target) {
			t.Fatalf("%s didn't cast", finisher.ActionID)
		}
		if got, want := cat.CurrentEnergy(), tc.energyLeft(finisher.DefaultCast.Cost); math.Abs(got-want) > 1e-6 {
			t.Errorf("%s dodged at 100 Energy leaves %.1f Energy, want %.1f", finisher.ActionID, got, want)
		}
		if cat.ComboPoints() != 5 {
			t.Errorf("%s dodged leaves %d combo points, want 5", finisher.ActionID, cat.ComboPoints())
		}
	}
}

// newCatSimWith builds the cat with one item on (by slot index in the equipment list) and the
// given options.
func newCatSimWith(t *testing.T, talents string, slot int, item int32, options *proto.FeralDruid_Options) (*core.Simulation, *FeralDruid) {
	items := make([]*proto.ItemSpec, 17)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[slot].Id = item
	options.InnervateTarget = &proto.UnitReference{}
	player := &proto.Player{
		Race:          proto.Race_RaceTauren,
		Class:         proto.Class_ClassDruid,
		Level:         60,
		Equipment:     &proto.EquipmentSpec{Items: items},
		TalentsString: talents,
		Rotation:      &proto.APLRotation{},
		Spec:          &proto.Player_FeralDruid{FeralDruid: &proto.FeralDruid{Options: options}},
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

// TestCatNaturalInstinctRip checks that Natural Instinct's crit damage bonus reaches Rip,
// whose bit its class mask shares with Ferocious Bite (1223242, word 0 0x839800).
func TestCatNaturalInstinctRip(t *testing.T) {
	_, cat := newCatSim(t, catTalents)
	for _, spell := range []*druid.DruidSpell{cat.Claw, cat.Rake, cat.Shred, cat.Ravage, cat.FerociousBite, cat.Rip} {
		if !near(spell.CritDamageBonus, 1.2) {
			t.Errorf("%s has %.2f crit damage bonus with Natural Instinct 2, want 1.20", spell.ActionID, spell.CritDamageBonus)
		}
	}
}

// TestCatRipTicksLive checks that a Forever Rip tick reads attack power and the damage
// multipliers when it lands, with the combo points kept from the cast: 15 + 25.5 a point and
// 1% of attack power a point, the fifth counted as the fourth, all 25% up with the November 2
// development notes.
func TestCatRipTicksLive(t *testing.T) {
	sim, cat := newCatSim(t, "")
	target := noCritsNoArmor(sim, cat)
	cat.AddComboPoints(sim, 5, target, cat.Rip.ComboPointMetrics())
	dot := cat.Rip.Dot(target)
	dot.Apply(sim)
	cat.SpendComboPoints(sim, cat.Rip.Spell)
	ap := cat.Rip.MeleeAttackPower(target)
	multiplier := dot.SnapshotAttackerMultiplier
	if got, want := tickDamage(sim, cat.Rip, target), 1.25*(15+25.5*5+0.04*ap)*multiplier; !nearDamage(got, want) {
		t.Errorf("Rip ticks for %.3f at %.0f attack power, want %.3f", got, ap, want)
	}
	cat.AddStatDynamic(sim, stats.AttackPower, 1000)
	cat.PseudoStats.DamageDealtMultiplier *= 1.1
	if got, want := tickDamage(sim, cat.Rip, target), 1.25*(15+25.5*5+0.04*(ap+1000))*multiplier*1.1; !nearDamage(got, want) {
		t.Errorf("Rip ticks for %.3f after 1000 attack power and 10%% damage more, want %.3f", got, want)
	}
}

// TestCatIdolOfTheDream checks that Idol of the Dream (220606, spell 446212: Rip +2000 ms)
// gives Rip a seventh tick of the same damage.
func TestCatIdolOfTheDream(t *testing.T) {
	for _, c := range []struct {
		item  int32
		ticks int32
	}{{0, 6}, {druid.IdolOfTheDream, 7}} {
		sim, cat := newCatSimWith(t, "", 16, c.item, &proto.FeralDruid_Options{})
		cat.AddComboPoints(sim, 5, cat.CurrentTarget, cat.Rip.ComboPointMetrics())
		dot := cat.Rip.Dot(cat.CurrentTarget)
		dot.Apply(sim)
		if dot.NumberOfTicks != c.ticks || dot.Duration != time.Duration(c.ticks)*2*time.Second {
			t.Errorf("item %d: Rip has %d ticks over %v, want %d", c.item, dot.NumberOfTicks, dot.Duration, c.ticks)
		}
	}
}

// TestCatPrepullClearcasting checks that the prepull_clearcasting option opens the fight with
// Clearcasting, which went up a global cooldown before the pull, and Omen of Clarity's 10 sec
// cooldown running from then.
func TestCatPrepullClearcasting(t *testing.T) {
	for _, on := range []bool{false, true} {
		sim, cat := newCatSimWith(t, "", 0, 0, &proto.FeralDruid_Options{PrepullClearcasting: on})
		for sim.CurrentTime < 0 && !sim.Step() {
		}
		if active := cat.ClearcastingAura.IsActive(); active != on {
			t.Errorf("option %v: Clearcasting up at the pull is %v", on, active)
		}
		if on && cat.ClearcastingAura.ExpiresAt() != 15*time.Second-core.GCDDefault {
			t.Errorf("Clearcasting expires at %v, want %v", cat.ClearcastingAura.ExpiresAt(), 15*time.Second-core.GCDDefault)
		}
	}
}

// TestCatExcessEnergy checks the Excess Energy value: with 50 Energy and a Rip at 5 combo
// points running out in 2 sec, the 20 Energy regenerated by then pays for 20 of the new Rip's
// 30, so 10 is kept.
func TestCatExcessEnergy(t *testing.T) {
	sim, cat := newCatSim(t, "")
	target := cat.CurrentTarget
	cat.AddComboPoints(sim, 5, target, cat.Rip.ComboPointMetrics())
	if got := cat.ExcessEnergy(sim); !near(got, cat.CurrentEnergy()) {
		t.Errorf("excess Energy is %.2f with no bleed up, want all %.2f", got, cat.CurrentEnergy())
	}
	dot := cat.Rip.Dot(target)
	dot.Apply(sim)
	dot.UpdateExpires(sim, sim.CurrentTime+2*time.Second)
	cat.SpendEnergy(sim, cat.CurrentEnergy()-50, cat.Rip.EnergyMetrics())
	if got := cat.ExcessEnergy(sim); !near(got, 40) {
		t.Errorf("excess Energy is %.2f, want 40", got)
	}
}

// TestCatFerociousBiteNov2 checks Ferocious Bite rank 5 (22829) with the November 2 development
// notes' 25% more base points and attack power part: 1.25 * (45 to 95 + 128 a combo point + 3% of
// attack power a combo point), with no extra Energy, crits or armor in the way.
func TestCatFerociousBiteNov2(t *testing.T) {
	for i := 0; i < 20; i++ {
		sim, cat := newCatSim(t, "")
		target := noCritsNoArmor(sim, cat)
		cat.AddComboPoints(sim, 5, target, cat.FerociousBite.ComboPointMetrics())
		cat.SpendEnergy(sim, cat.CurrentEnergy()-35, cat.FerociousBite.EnergyMetrics())
		ap := cat.FerociousBite.MeleeAttackPower(target)
		multiplier := cat.FerociousBite.CalcDamage(sim, target, 1000, cat.FerociousBite.OutcomeAlwaysHit).Damage / 1000
		metrics := &cat.FerociousBite.SpellMetrics[target.UnitIndex]
		cat.SetGCDTimer(sim, sim.CurrentTime)
		if !cat.FerociousBite.Cast(sim, target) {
			t.Fatal("Ferocious Bite not cast")
		}
		if metrics.Hits == 0 {
			continue
		}
		base := metrics.TotalDamage / multiplier
		low, high := 1.25*(45+128*5+0.15*ap), 1.25*(95+128*5+0.15*ap)
		if base < low-1e-6 || base > high+1e-6 {
			t.Errorf("Ferocious Bite's base damage is %.2f, want %.2f to %.2f", base, low, high)
		}
		return
	}
	t.Fatal("no Ferocious Bite landed")
}

// TestCatClawDamage checks Claw (9850) at the November 2 development notes' 140%: (weapon damage +
// 115) * 140%, with no talents, armor or crits in the way.
func TestCatClawDamage(t *testing.T) {
	sim, cat := newCatSim(t, "")
	target := cat.CurrentTarget
	target.AddStatDynamic(sim, stats.Armor, -target.GetStat(stats.Armor))
	ap := cat.Claw.MeleeAttackPower(target)
	weapon := cat.AutoAttacks.MH()
	want := (115 + (weapon.BaseDamageMin+weapon.BaseDamageMax)/2 + ap/14) * 1.4
	result := cat.Claw.CalcDamage(sim, target, 115+weapon.CalculateAverageWeaponDamage(ap), cat.Claw.OutcomeAlwaysHit)
	if math.Abs(result.Damage-want) > 1e-6 {
		t.Errorf("Claw hits for %.2f, want %.2f", result.Damage, want)
	}
}
