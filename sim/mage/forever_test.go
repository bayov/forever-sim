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

// Since 1.60.1.70425 each Arcane Blast stack is 15% of Intellect as spell damage and 175% more
// mana on Arcane Blast. The buff lasts 15 sec from the first stack, and other spells don't end it.
func TestMageArcaneBlastStacks(t *testing.T) {
	// Arcane to Arcane Wrath, with Missile Barrage 2.
	sim, mage := newMageSim(t, "050200500310001")
	arcaneBlast := topRank(t, "Arcane Blast", mage.ArcaneBlast)
	fireball := topRank(t, "Fireball", mage.Fireball)
	if got := arcaneBlast.CastTime(); got != 1500*time.Millisecond {
		t.Errorf("Arcane Blast casts in %s, want 1.5s", got)
	}
	baseCost := arcaneBlast.Cost.GetCurrentCost()
	if want := 0.15 * mage.BaseMana; math.Abs(baseCost-want) > 1e-6 {
		t.Fatalf("Arcane Blast costs %.1f, want %.1f (15%% of base mana)", baseCost, want)
	}
	baseSpellPower := mage.GetStat(stats.SpellPower)
	baseMultiplier := fireball.DamageMultiplierAdditive

	cast := func(spell *core.Spell) {
		mage.SetGCDTimer(sim, sim.CurrentTime)
		if !spell.Cast(sim, mage.CurrentTarget) {
			t.Fatalf("%s didn't cast", spell.ActionID)
		}
		for !sim.Step() && mage.Hardcast.Expires > sim.CurrentTime {
		}
	}
	for i := 0; i < 3; i++ {
		cast(arcaneBlast)
	}
	if got := mage.ArcaneBlastAura.GetStacks(); got != 3 {
		t.Fatalf("%d Arcane Blast stacks after 3 casts, want 3", got)
	}
	if got, want := arcaneBlast.Cost.GetCurrentCost(), baseCost*(1+3*1.75); math.Abs(got-want) > 1e-6 {
		t.Errorf("Arcane Blast costs %.1f at 3 stacks, want %.1f", got, want)
	}
	if got, want := mage.GetStat(stats.SpellPower)-baseSpellPower, 0.45*mage.GetStat(stats.Intellect); math.Abs(got-want) > 1e-6 {
		t.Errorf("3 stacks give %.2f spell damage, want %.2f (45%% of Intellect)", got, want)
	}
	if got := fireball.DamageMultiplierAdditive; math.Abs(got-baseMultiplier) > 1e-9 {
		t.Errorf("Fireball's multiplier is %.2f at 3 stacks, want %.2f", got, baseMultiplier)
	}
	// The first stack came 1.5 sec into the first cast, and later stacks don't refresh it.
	if got, want := mage.ArcaneBlastAura.RemainingDuration(sim), 15*time.Second-3*time.Second; got != want {
		t.Errorf("the stacks have %s left, want %s", got, want)
	}

	cast(fireball)
	if got := mage.ArcaneBlastAura.GetStacks(); got != 3 {
		t.Errorf("%d Arcane Blast stacks after a Fireball, want 3", got)
	}
	// 15 sec after the first stack, which came at 1.5 sec.
	sim.AddPendingAction(&core.PendingAction{NextActionAt: 16500 * time.Millisecond, OnAction: func(*core.Simulation) {}})
	for sim.CurrentTime < 16500*time.Millisecond && !sim.Step() {
	}
	if mage.ArcaneBlastAura.IsActive() {
		t.Errorf("the Arcane Blast stacks are still up %s in", sim.CurrentTime)
	}
	if got := mage.GetStat(stats.SpellPower); math.Abs(got-baseSpellPower) > 1e-6 {
		t.Errorf("spell damage is %.2f after the stacks ended, want %.2f", got, baseSpellPower)
	}
}

// Arcane Wrath casts in 3.5 sec, has a 6 sec cooldown and gets 5% of maximum mana as spell damage
// at a full coefficient.
func TestMageArcaneWrath(t *testing.T) {
	sim, mage := newMageSim(t, "050200500310001")
	arcaneWrath := topRank(t, "Arcane Wrath", mage.ArcaneWrath)
	if got := arcaneWrath.CastTime(); got != 3500*time.Millisecond {
		t.Errorf("Arcane Wrath casts in %s, want 3.5s", got)
	}
	if got := arcaneWrath.CD.Duration; got != 6*time.Second {
		t.Errorf("Arcane Wrath cooldown %s, want 6s", got)
	}
	if got := arcaneWrath.Cost.GetCurrentCost(); got != 395 {
		t.Errorf("Arcane Wrath costs %.0f, want 395", got)
	}
	mage.SetGCDTimer(sim, sim.CurrentTime)
	if !arcaneWrath.Cast(sim, mage.CurrentTarget) {
		t.Fatalf("Arcane Wrath didn't cast")
	}
	for !sim.Step() && sim.CurrentTime < 5*time.Second {
	}
	metrics := arcaneWrath.SpellMetrics[mage.CurrentTarget.UnitIndex]
	if metrics.Casts != 1 {
		t.Fatalf("%d Arcane Wrath casts, want 1", metrics.Casts)
	}
	// No gear: the spell damage is all from the mana, and there are no damage modifiers.
	low := 405.67 + 0.05*mage.MaxMana()
	high := 496.33 + 0.05*mage.MaxMana()
	damage := metrics.TotalDamage
	if metrics.Crits > 0 {
		damage /= 1.5 + arcaneWrath.CritDamageBonus
	}
	if metrics.Hits+metrics.Crits == 1 && (damage < low-1e-6 || damage > high+1e-6) {
		t.Errorf("Arcane Wrath hit for %.1f, want %.1f to %.1f", damage, low, high)
	}
}

// Fire Power: 1% more Fire damage per point for each of the mage's own Fire effects on the
// target, at most 4. Frostfire Bolt's DoT is two schools and doesn't count.
func TestMageFirePower(t *testing.T) {
	// Fire to Fire Power 3, with Improved Scorch 3.
	sim, mage := newMageSim(t, "-2355000013010003")
	target := mage.CurrentTarget
	at := mage.AttackTables[target.UnitIndex][proto.CastType_CastTypeMainHand]
	fireball := topRank(t, "Fireball", mage.Fireball)
	frostbolt := topRank(t, "Frostbolt", mage.Frostbolt)
	pyroblast := topRank(t, "Pyroblast", mage.Pyroblast)
	frostfireBolt := topRank(t, "Frostfire Bolt", mage.FrostfireBolt)

	check := func(what string, spell *core.Spell, want float64) {
		t.Helper()
		if got := spell.TargetDamageMultiplier(at, false); math.Abs(got-want) > 1e-9 {
			t.Errorf("%s: %s takes %.4f, want %.4f", what, spell.ActionID, got, want)
		}
	}
	check("no effects", fireball, 1)
	frostfireBolt.Dot(target).Activate(sim)
	check("Frostfire Bolt DoT", fireball, 1)
	mage.Ignite.Dot(target).Activate(sim)
	check("Ignite", fireball, 1.03)
	fireball.Dot(target).Activate(sim)
	check("Ignite and Fireball", fireball, 1.06)
	check("Ignite and Fireball", frostbolt, 1)
	check("Ignite and Fireball", mage.Ignite, 1)
	aura := mage.ImprovedScorchAuras.Get(target)
	aura.Activate(sim)
	aura.SetStacks(sim, 5)
	check("three effects", fireball, 1.15*1.09)
	pyroblast.Dot(target).Activate(sim)
	topRank(t, "Flamestrike", mage.Flamestrike).AOEDot().Activate(sim)
	check("five effects", fireball, 1.15*1.12)
	check("five effects", frostfireBolt, 1.15*1.12)
	check("five effects", mage.Ignite, 1)
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

// Shatter's crit comes with Fingers of Frost, on every spell. Winter's Chill gives no crit since
// 1.60.1.70425.
func TestMageShatter(t *testing.T) {
	// Frost to Winter's Chill 2, with Shatter 3 and Fingers of Frost 2.
	sim, mage := newMageSim(t, "--055500033100130122")
	target := mage.CurrentTarget
	frostbolt := topRank(t, "Frostbolt", mage.Frostbolt)
	iceLance := topRank(t, "Ice Lance", mage.IceLance)
	blizzard := topRank(t, "Blizzard", mage.Blizzard)
	rating := float64(core.SpellCritRatingPerCritChance)
	if got := mage.personalCritRating(frostbolt, target); got != 0 {
		t.Errorf("Frostbolt gets %.1f crit rating without Fingers of Frost, want 0", got)
	}

	mage.FingersOfFrostAura.Activate(sim)
	mage.FingersOfFrostAura.SetStacks(sim, 2)
	if got := mage.personalCritRating(iceLance, target); got != 50*rating {
		t.Errorf("Ice Lance gets %.1f crit rating with Fingers of Frost, want %.1f", got, 50*rating)
	}
	if got := mage.personalCritRating(blizzard, target); got != 50*rating {
		t.Errorf("Blizzard gets %.1f crit rating with Fingers of Frost, want %.1f", got, 50*rating)
	}
}

// Blizzard's and Flamestrike's ticks come from the area trigger's own spells, which carry the
// client's coefficients (4.2% and 3.2% a tick), not the 3% dummy on Blizzard itself. Combustion
// stacks to 10 (28682).
func TestMageAreaCoefficientsAndCombustion(t *testing.T) {
	_, mage := newMageSim(t, "-235500001301")
	if got := topRank(t, "Blizzard", mage.Blizzard).AOEDot().BonusCoefficient; math.Abs(got-0.042) > 1e-9 {
		t.Errorf("Blizzard tick coefficient %v, want 0.042", got)
	}
	for rank, spell := range mage.Flamestrike {
		if spell == nil {
			continue
		}
		if got := spell.AOEDot().BonusCoefficient; math.Abs(got-0.032) > 1e-9 {
			t.Errorf("Flamestrike rank %d DoT coefficient %v, want 0.032", rank, got)
		}
	}
	if mage.CombustionAura != nil && mage.CombustionAura.MaxStacks != 10 {
		t.Errorf("Combustion stacks to %d, want 10", mage.CombustionAura.MaxStacks)
	}
}

// The report assumptions the user ruled on 2026-10-15: Ice Lance takes SoD's 42.9% of spell power
// on every rank, and Frostbolt rank 10 deals 4% more than the client (rank 9 is unchanged).
func TestMageRankTuning(t *testing.T) {
	_, mage := newMageSim(t, "--055500033100130122")
	for rank, spell := range mage.IceLance {
		if spell == nil {
			continue
		}
		if got := spell.BonusCoefficient; math.Abs(got-0.429) > 1e-9 {
			t.Errorf("Ice Lance rank %d coefficient %v, want 0.429", rank, got)
		}
	}
	for _, c := range []struct {
		rank      int
		low, high float64
	}{
		{10, 382.44 * 1.04, 412.76 * 1.04},
		{9, 305.42, 331.38},
	} {
		low, high := FrostboltBaseDamageAt(c.rank, 60)
		if math.Abs(low-c.low) > 1e-9 || math.Abs(high-c.high) > 1e-9 {
			t.Errorf("Frostbolt rank %d at 60: %.2f to %.2f, want %.2f to %.2f", c.rank, low, high, c.low, c.high)
		}
	}
}

// The staff imbue scrolls add their stat only on a staff.
func TestMageStaffImbue(t *testing.T) {
	for _, tc := range []struct {
		imbue proto.Mage_Options_StaffImbue
		stat  stats.Stat
		want  float64
	}{
		{proto.Mage_Options_ImbueGreaterFlame, stats.FirePower, 20},
		{proto.Mage_Options_ImbueGreaterFrost, stats.FrostPower, 20},
		{proto.Mage_Options_ImbuePrecision, stats.SpellCrit, 1 * core.SpellCritRatingPerCritChance},
	} {
		for _, staff := range []bool{true, false} {
			equipment := &proto.EquipmentSpec{}
			if staff {
				equipment.Items = make([]*proto.ItemSpec, 15)
				for i := range equipment.Items {
					equipment.Items[i] = &proto.ItemSpec{}
				}
				// Staff of Dominance.
				equipment.Items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 18842}
			}
			player := &proto.Player{
				Race:      proto.Race_RaceGnome,
				Class:     proto.Class_ClassMage,
				Level:     60,
				Equipment: equipment,
				Spec:      &proto.Player_Mage{Mage: &proto.Mage{Options: &proto.Mage_Options{StaffImbue: tc.imbue}}},
				Rotation:  &proto.APLRotation{},
			}
			base := *player
			base.Spec = &proto.Player_Mage{Mage: &proto.Mage{Options: &proto.Mage_Options{}}}
			got := mageStat(player, tc.stat) - mageStat(&base, tc.stat)
			want := tc.want
			if !staff {
				want = 0
			}
			if math.Abs(got-want) > 1e-9 {
				t.Errorf("%v, staff %v: %v gives %v, want %v", tc.imbue, staff, tc.stat, got, want)
			}
		}
	}
}

func mageStat(player *proto.Player, stat stats.Stat) float64 {
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim.Raid.Parties[0].Players[0].GetCharacter().GetStat(stat)
}

// Pushback protection from the client masks: Burning Soul on Fireball, Scorch, Pyroblast,
// Flamestrike and Frostfire Bolt, Improved Channeling on Arcane Missiles, Arcane Blast and
// Arcane Wrath.
func TestMagePushbackProtection(t *testing.T) {
	// Improved Channeling 5 with Arcane Wrath, and Burning Soul 3.
	_, mage := newMageSim(t, "055200500310001-00000030")
	for _, c := range []struct {
		name  string
		spell *core.Spell
		want  float64
	}{
		{"Arcane Missiles", topRank(t, "Arcane Missiles", mage.ArcaneMissiles), 1.0},
		{"Arcane Blast", topRank(t, "Arcane Blast", mage.ArcaneBlast), 0.70},
		{"Arcane Wrath", topRank(t, "Arcane Wrath", mage.ArcaneWrath), 0.70},
		{"Fireball", topRank(t, "Fireball", mage.Fireball), 0.70},
		{"Frostfire Bolt", topRank(t, "Frostfire Bolt", mage.FrostfireBolt), 0.70},
		{"Frostbolt", topRank(t, "Frostbolt", mage.Frostbolt), 0},
	} {
		if math.Abs(c.spell.PushbackReduction-c.want) > 1e-9 {
			t.Errorf("%s ignores pushback %.2f of the time, want %.2f", c.name, c.spell.PushbackReduction, c.want)
		}
	}
}
