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
	sim, mage := newMageSim(t, "-355000012301")
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

// Since 1.60.1.70425 every Fire Blast that deals damage adds a Heating Up stack, and a crit adds
// a second. A Pyroblast made instant by Presence of Mind leaves the stacks alone.
func TestMageHeatingUpFireBlastAndPresenceOfMind(t *testing.T) {
	// Arcane to Presence of Mind, Fire to Heating Up.
	sim, mage := newMageSim(t, "000225203100001-355000012301")
	fireBlast := topRank(t, "Fire Blast", mage.FireBlast)
	for i := 0; i < 8; i++ {
		mage.HeatingUpAura.Deactivate(sim)
		fireBlast.CD.Reset()
		mage.SetGCDTimer(sim, sim.CurrentTime)
		crits := fireBlast.SpellMetrics[mage.CurrentTarget.UnitIndex].Crits
		hits := fireBlast.SpellMetrics[mage.CurrentTarget.UnitIndex].Hits
		if !fireBlast.Cast(sim, mage.CurrentTarget) {
			t.Fatalf("Fire Blast %d didn't cast (mana %.0f)", i, mage.CurrentMana())
		}
		crit := fireBlast.SpellMetrics[mage.CurrentTarget.UnitIndex].Crits > crits
		hit := fireBlast.SpellMetrics[mage.CurrentTarget.UnitIndex].Hits > hits
		want := int32(0)
		if crit {
			want = 2
		} else if hit {
			want = 1
		}
		if got := mage.HeatingUpAura.GetStacks(); got != want {
			t.Errorf("Fire Blast (crit %v, hit %v) left %d Heating Up stacks, want %d", crit, hit, got, want)
		}
	}

	pyroblast := topRank(t, "Pyroblast", mage.Pyroblast)
	mage.HeatingUpAura.Activate(sim)
	mage.HeatingUpAura.SetStacks(sim, 2)
	mage.PresenceOfMindAura.Activate(sim)
	mage.SetGCDTimer(sim, sim.CurrentTime)
	if !pyroblast.Cast(sim, mage.CurrentTarget) {
		t.Fatalf("Pyroblast didn't cast")
	}
	if mage.PresenceOfMindAura.IsActive() {
		t.Errorf("Presence of Mind is still up after the Pyroblast")
	}
	if got := mage.HeatingUpAura.GetStacks(); got != 2 {
		t.Errorf("an instant Pyroblast with Presence of Mind left %d Heating Up stacks, want 2", got)
	}
}

// Since 1.60.1.70425 each Arcane Blast stack is 15% of Intellect as spell damage and 175% more
// mana on Arcane Blast. The buff lasts 15 sec from the first stack, and other spells don't end it.
func TestMageArcaneBlastStacks(t *testing.T) {
	// Arcane to Arcane Wrath, with Missile Barrage 2.
	sim, mage := newMageSim(t, "00000520310001")
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
	sim, mage := newMageSim(t, "00000520310001")
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
	sim, mage := newMageSim(t, "-3550000123010003")
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
	sim, mage := newMageSim(t, "-355000012301")
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
	sim, mage := newMageSim(t, "--05350003010130122")
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
	_, mage := newMageSim(t, "-355000012301")
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
	_, mage := newMageSim(t, "--05350003010130122")
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
	_, mage := newMageSim(t, "05000520310001-000003")
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

// Since the 2026-10-20 hotfix Ice Barrier raises Frost damage by 3% while the shield holds, and
// every cast gives Frigid Power, 3% more for a minute, each its own factor. The absorb is the rank's base and 10% of
// Frost spell power, and Frost Warding raises it 10% a point. Precision is 2% hit with all spells a
// point.
func TestMageIceBarrier(t *testing.T) {
	// Frost Warding 2, Precision 3, to Ice Barrier.
	sim, mage := newMageSim(t, "--253500133101301201")
	frostbolt := topRank(t, "Frostbolt", mage.Frostbolt)
	fireBlast := topRank(t, "Fire Blast", mage.FireBlast)
	if got, want := mage.GetStat(stats.SpellHit), 6*float64(core.SpellHitRatingPerHitChance); math.Abs(got-want) > 1e-9 {
		t.Errorf("Precision 3 gives %.1f spell hit, want %.1f", got, want)
	}
	frostBase, fireBase := frostbolt.DamageMultiplier, fireBlast.DamageMultiplier

	mage.SetGCDTimer(sim, sim.CurrentTime)
	if !topRank(t, "Ice Barrier", mage.IceBarrier).Cast(sim, &mage.Unit) {
		t.Fatalf("Ice Barrier didn't cast")
	}
	if got := frostbolt.DamageMultiplier / frostBase; math.Abs(got-1.03*1.03) > 1e-9 {
		t.Errorf("Frostbolt does %.4f times the damage with the shield and Frigid Power up, want 1.0609", got)
	}
	if got := fireBlast.DamageMultiplier; got != fireBase {
		t.Errorf("Fire Blast's damage multiplier moved to %.3f, want %.3f", got, fireBase)
	}

	// Rank 4 at 60 is 811 + 2 levels of 4, times 1.2 for Frost Warding 2.
	if got, want := IceBarrierAbsorbAt(4, 60, 0, 2), 819*1.2; math.Abs(got-want) > 1e-9 {
		t.Errorf("absorb %.1f with no spell power, want %.1f", got, want)
	}
	absorb := IceBarrierAbsorbAt(4, 60, mage.frostSpellPower(), 2)
	hit := func(damage float64) float64 {
		result := &core.SpellResult{Damage: damage}
		for _, modifier := range mage.DynamicDamageTakenModifiers {
			modifier(sim, nil, result)
		}
		return result.Damage
	}
	if got := hit(500); got != 0 {
		t.Errorf("a 500 hit does %.1f through the shield, want 0", got)
	}
	if got, want := hit(1000), 1500-absorb; math.Abs(got-want) > 1e-9 {
		t.Errorf("a 1000 hit does %.1f through what is left of the shield, want %.1f", got, want)
	}
	if mage.IceBarrierAuras[4].IsActive() {
		t.Errorf("the shield is still up after it absorbed %.1f", absorb)
	}
	if !mage.FrigidPowerAura.IsActive() {
		t.Errorf("Frigid Power went with the shield")
	}
	if got := frostbolt.DamageMultiplier / frostBase; math.Abs(got-1.03) > 1e-9 {
		t.Errorf("Frostbolt does %.4f times the damage with Frigid Power alone, want 1.03", got)
	}
	if got := hit(1000); got != 1000 {
		t.Errorf("a 1000 hit does %.1f with the shield gone, want 1000", got)
	}
}

// Under Forever each percent damage and cost row is its own factor (wowsims/forever fef2bcaebd and
// 810101380a). Arcane Power with Arcane Instability 3 is 1.3 x 1.03, four Arcane Blast stacks
// under Arcane Power cost 8 x 1.3 times the base, a Missile Barrage Arcane Missiles is free under
// Arcane Power, and the Gnome's Eureka! multiplies Arcane Power's cost too.
func TestMageModifiersMultiply(t *testing.T) {
	// The Arcane preset's tree: Missile Barrage 2, Arcane Blast, Arcane Instability 3, Arcane Power.
	sim, mage := newMageSim(t, "050205203100301531")
	fireball := topRank(t, "Fireball", mage.Fireball)
	arcaneBlast := topRank(t, "Arcane Blast", mage.ArcaneBlast)
	arcaneMissiles := topRank(t, "Arcane Missiles", mage.ArcaneMissiles)
	if got := fireball.DamageMultiplier; math.Abs(got-1.03) > 1e-9 {
		t.Errorf("Fireball does %.4f times the damage with Arcane Instability 3, want 1.03", got)
	}
	fireballCost := fireball.Cost.GetCurrentCost()
	blastCost := arcaneBlast.Cost.GetCurrentCost()
	missilesCost := arcaneMissiles.Cost.GetCurrentCost()

	mage.ArcanePowerAura.Activate(sim)
	if got := fireball.DamageMultiplier; math.Abs(got-1.339) > 1e-9 {
		t.Errorf("Fireball does %.4f times the damage with Arcane Power and Arcane Instability 3, want 1.339", got)
	}
	mage.ArcaneBlastAura.Activate(sim)
	mage.ArcaneBlastAura.SetStacks(sim, 4)
	if got, want := arcaneBlast.Cost.GetCurrentCost(), blastCost*8*1.3; math.Abs(got-want) > 1e-6 {
		t.Errorf("Arcane Blast costs %.1f at 4 stacks under Arcane Power, want %.1f (10.4 times the base)", got, want)
	}
	mage.MissileBarrageAura.Activate(sim)
	if got := arcaneMissiles.Cost.GetCurrentCost(); got != 0 {
		t.Errorf("Arcane Missiles costs %.1f with Missile Barrage under Arcane Power, want 0", got)
	}
	eureka := mage.GetAura("Eureka!")
	eureka.Activate(sim)
	if got, want := fireball.Cost.GetCurrentCost(), fireballCost*1.3*0.9; math.Abs(got-want) > 1e-6 {
		t.Errorf("Fireball costs %.1f with Arcane Power and Eureka!, want %.1f", got, want)
	}

	eureka.Deactivate(sim)
	mage.MissileBarrageAura.Deactivate(sim)
	mage.ArcaneBlastAura.Deactivate(sim)
	mage.ArcanePowerAura.Deactivate(sim)
	for _, c := range []struct {
		name      string
		spell     *core.Spell
		got, want float64
	}{
		{"Fireball", fireball, fireball.Cost.GetCurrentCost(), fireballCost},
		{"Arcane Blast", arcaneBlast, arcaneBlast.Cost.GetCurrentCost(), blastCost},
		{"Arcane Missiles", arcaneMissiles, arcaneMissiles.Cost.GetCurrentCost(), missilesCost},
	} {
		if math.Abs(c.got-c.want) > 1e-6 {
			t.Errorf("%s costs %.3f once the buffs are gone, want %.3f", c.name, c.got, c.want)
		}
	}
	if got := fireball.DamageMultiplier; math.Abs(got-1.03) > 1e-9 {
		t.Errorf("Fireball does %.4f times the damage once Arcane Power is gone, want 1.03", got)
	}
}

// Frost Channeling's cost row (11160 effect 0) takes Frostfire Bolt by word 1 bit 12, the bit Ice
// Shards and Critical Mass reach it by, and Ice Barrier too.
func TestMageFrostChanneling(t *testing.T) {
	// Frost to Ice Barrier, with Frost Channeling 3.
	_, mage := newMageSim(t, "--053500030101301221")
	for _, c := range []struct {
		name  string
		spell *core.Spell
	}{
		{"Frostbolt", topRank(t, "Frostbolt", mage.Frostbolt)},
		{"Frostfire Bolt", topRank(t, "Frostfire Bolt", mage.FrostfireBolt)},
		{"Ice Barrier", topRank(t, "Ice Barrier", mage.IceBarrier)},
	} {
		if got, want := c.spell.Cost.GetCurrentCost(), c.spell.Cost.BaseCost*0.85; math.Abs(got-want) > 1e-6 {
			t.Errorf("%s costs %.1f with Frost Channeling 3, want %.1f", c.name, got, want)
		}
	}
	if got := topRank(t, "Fireball", mage.Fireball); got.Cost.GetCurrentCost() != got.Cost.BaseCost {
		t.Errorf("Fireball costs %.1f with Frost Channeling 3, want %.1f", got.Cost.GetCurrentCost(), got.Cost.BaseCost)
	}
}

// Fingers of Frost, as wowsims/forever has it from their logs (1373a0b96f): Shatter's crit reaches
// every mage damage spell, a cast spends its charge a spell batch after it finishes, a channel
// spends none at its cast, and each Arcane Missiles missile that lands spends one.
func TestMageFingersOfFrostCharges(t *testing.T) {
	// Frost to Winter's Chill 2, with Shatter 3 and Fingers of Frost 2.
	sim, mage := newMageSim(t, "--05350003010130122")
	target := mage.CurrentTarget
	fireBlast := topRank(t, "Fire Blast", mage.FireBlast)
	arcaneMissiles := topRank(t, "Arcane Missiles", mage.ArcaneMissiles)
	fof := mage.FingersOfFrostAura
	advance := func(by time.Duration) {
		until := sim.CurrentTime + by
		sim.AddPendingAction(&core.PendingAction{NextActionAt: until, OnAction: func(*core.Simulation) {}})
		for sim.CurrentTime < until && !sim.Step() {
		}
	}
	// The charges come a little before the cast: a cast at the moment they came granted them.
	grant := func() {
		fof.Activate(sim)
		fof.SetStacks(sim, 2)
		advance(100 * time.Millisecond)
	}
	// Free, so the mana lasts.
	fireBlast.Cost.Multiplier = 0
	arcaneMissiles.Cost.Multiplier = 0

	// 51% crit without Shatter and every spell hits, so only Shatter's 50% makes every cast crit.
	fireBlast.BonusHitRating += 100 * core.SpellHitRatingPerHitChance
	fireBlast.BonusCritRating += 51*core.SpellCritRatingPerCritChance - 100*fireBlast.SpellCritChance(target)*core.SpellCritRatingPerCritChance
	const casts = 12
	for i := 0; i < casts; i++ {
		grant()
		fireBlast.CD.Reset()
		mage.SetGCDTimer(sim, sim.CurrentTime)
		if !fireBlast.Cast(sim, target) {
			t.Fatalf("Fire Blast didn't cast")
		}
		if got := fof.GetStacks(); got != 2 {
			t.Fatalf("%d Fingers of Frost charges right after a Fire Blast, want 2 until the batch ends", got)
		}
		advance(core.SpellBatchWindow)
		if got := fof.GetStacks(); got != 1 {
			t.Fatalf("%d Fingers of Frost charges a batch after a Fire Blast, want 1", got)
		}
	}
	metrics := fireBlast.SpellMetrics[target.UnitIndex]
	if got := metrics.Crits + metrics.ResistedCrits; got != casts {
		t.Errorf("%d of %d Fire Blasts crit at 51%% crit with Fingers of Frost, want all (Shatter's 50%%)", got, casts)
	}

	// A cast spends the charge a batch late, so a spell in the same batch still reads Frozen.
	grant()
	fof.SetStacks(sim, 1)
	fireBlast.CD.Reset()
	mage.SetGCDTimer(sim, sim.CurrentTime)
	if !fireBlast.Cast(sim, target) {
		t.Fatalf("Fire Blast didn't cast")
	}
	if !mage.isTargetFrozen(target) {
		t.Errorf("the target stopped counting as Frozen at the cast that spends the last charge")
	}
	advance(core.SpellBatchWindow)
	if fof.IsActive() {
		t.Errorf("Fingers of Frost is still up a batch after its last charge was spent")
	}

	// The channel spends nothing; each missile that lands spends a charge.
	grant()
	for _, tick := range mage.ArcaneMissilesTickSpell {
		if tick != nil {
			tick.BonusHitRating += 100 * core.SpellHitRatingPerHitChance
		}
	}
	mage.SetGCDTimer(sim, sim.CurrentTime)
	if !arcaneMissiles.Cast(sim, target) {
		t.Fatalf("Arcane Missiles didn't cast")
	}
	advance(core.SpellBatchWindow)
	if got := fof.GetStacks(); got != 2 {
		t.Errorf("%d Fingers of Frost charges after Arcane Missiles started, want 2", got)
	}
	advance(time.Second)
	if got := fof.GetStacks(); got != 1 {
		t.Errorf("%d Fingers of Frost charges after the first missile, want 1", got)
	}
	advance(time.Second)
	if fof.IsActive() {
		t.Errorf("Fingers of Frost is still up after two missiles")
	}
}

// Master of Elements gives back its share of the base cost on any Fire or Frost crit, a
// Clearcasting cast's too (wowsims/forever c999a66b3f).
func TestMageMasterOfElements(t *testing.T) {
	// The Fire preset's tree: Arcane Concentration 5, Master of Elements 3.
	sim, mage := newMageSim(t, "0002052-35500301230133031")
	target := mage.CurrentTarget
	fireBlast := topRank(t, "Fire Blast", mage.FireBlast)
	fireBlast.BonusHitRating += 100 * core.SpellHitRatingPerHitChance
	fireBlast.BonusCritRating += 100 * core.SpellCritRatingPerCritChance

	mage.SpendMana(sim, 1000, mage.NewManaMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionNone}))
	mage.ClearcastingAura.Activate(sim)
	if got := fireBlast.Cost.GetCurrentCost(); got != 0 {
		t.Fatalf("Fire Blast costs %.1f with Clearcasting, want 0", got)
	}
	before := mage.CurrentMana()
	mage.SetGCDTimer(sim, sim.CurrentTime)
	if !fireBlast.Cast(sim, target) {
		t.Fatalf("Fire Blast didn't cast")
	}
	if got, want := mage.CurrentMana()-before, 0.3*fireBlast.Cost.BaseCost; math.Abs(got-want) > 1e-6 {
		t.Errorf("a free Fire Blast crit gave back %.1f mana, want %.1f (30%% of its base cost)", got, want)
	}
}

// Cone of Cold under Forever: rank 5 is 328 to 358 at 60 with a 12.9% coefficient for 555 mana,
// on a 10 sec cooldown that Cold Snap finishes, castable only within its 10 yards. Improved Cone
// of Cold is 12 / 23 / 35%. Winter's Chill stacks on the target from landed Frostbolts and from
// Ice Lances cast at a Frozen target, and each is 5% more damage and 7% of spell power on the
// mage's next Cone of Cold, which spends them all (the 2026-10-20 hotfix, push 112694).
func TestMageConeOfColdAndWintersChill(t *testing.T) {
	// wowsims/forever's Frost build (41214cc8d0): Frost Channeling 3, Improved Cone of Cold 3,
	// Cold Snap, Fingers of Frost 2 and Winter's Chill 5.
	sim, mage := newMageSim(t, "0002052011003--053500030100331251")
	target := mage.CurrentTarget
	cone := topRank(t, "Cone of Cold", mage.ConeOfCold)
	frostbolt := topRank(t, "Frostbolt", mage.Frostbolt)
	iceLance := topRank(t, "Ice Lance", mage.IceLance)
	if cone.SpellID != 10161 {
		t.Fatalf("top Cone of Cold rank is %d, want 10161", cone.SpellID)
	}
	if got, want := cone.Cost.GetCurrentCost(), 555*0.85; math.Abs(got-want) > 1e-6 {
		t.Errorf("Cone of Cold costs %.1f with Frost Channeling 3, want %.1f", got, want)
	}
	if !cone.Flags.Matches(core.SpellFlagBinary) || iceLance.Flags.Matches(core.SpellFlagBinary) {
		t.Errorf("Cone of Cold binary %v, Ice Lance binary %v, want true and false",
			cone.Flags.Matches(core.SpellFlagBinary), iceLance.Flags.Matches(core.SpellFlagBinary))
	}
	if low, high := ConeOfColdScaling[5].At(ConeOfColdBaseDamage[5][0], 60), ConeOfColdScaling[5].At(ConeOfColdBaseDamage[5][1], 60); low != 328 || high != 358 {
		t.Errorf("rank 5 is %.0f to %.0f at 60, want 328 to 358", low, high)
	}

	for _, spell := range []*core.Spell{cone, frostbolt, iceLance} {
		spell.BonusHitRating += 100 * core.SpellHitRatingPerHitChance
		spell.Cost.Multiplier = 0
	}
	advance := func(by time.Duration) {
		until := sim.CurrentTime + by
		sim.AddPendingAction(&core.PendingAction{NextActionAt: until, OnAction: func(*core.Simulation) {}})
		for sim.CurrentTime < until && !sim.Step() {
		}
	}
	castNow := func(spell *core.Spell) {
		t.Helper()
		mage.SetGCDTimer(sim, sim.CurrentTime)
		if !spell.Cast(sim, target) {
			t.Fatalf("%s didn't cast", spell.ActionID)
		}
	}

	// Out of its radius it can't be cast.
	mage.DistanceFromTarget = 25
	if cone.CanCast(sim, target) {
		t.Errorf("Cone of Cold can be cast at 25 yards")
	}
	mage.DistanceFromTarget = 10

	// Winter's Chill 5 is a sure stack on every landed Frostbolt.
	for i := 0; i < 2; i++ {
		castNow(frostbolt)
		advance(4 * time.Second)
	}
	if got := mage.wintersChillStacks(target); got != 2 {
		t.Fatalf("%d Winter's Chill stacks after two Frostbolts, want 2", got)
	}

	// An Ice Lance adds one only when the target counts as Frozen. Every rank shares its 8 sec
	// cooldown.
	castNow(iceLance)
	if mage.IceLance[1].CD.IsReady(sim) {
		t.Errorf("rank 1 Ice Lance is ready right after rank 6, want both on an 8 sec cooldown")
	}
	advance(IceLanceCooldown)
	if got := mage.wintersChillStacks(target); got != 2 {
		t.Errorf("%d Winter's Chill stacks after an Ice Lance at a target not Frozen, want 2", got)
	}
	mage.FingersOfFrostAura.Activate(sim)
	mage.FingersOfFrostAura.SetStacks(sim, 2)
	castNow(iceLance)
	advance(2 * time.Second)
	if got := mage.wintersChillStacks(target); got != 3 {
		t.Errorf("%d Winter's Chill stacks after an Ice Lance at a Frozen target, want 3", got)
	}
	mage.FingersOfFrostAura.Deactivate(sim)

	// The cone reads the stacks as it hits, and spends them.
	baseMultiplier, baseCoefficient := cone.DamageMultiplier, cone.BonusCoefficient
	if math.Abs(baseMultiplier-1.35) > 1e-9 || math.Abs(baseCoefficient-ConeOfColdSpellCoeff) > 1e-9 {
		t.Errorf("Cone of Cold has %.4f damage and %.3f coefficient, want 1.35 (Improved Cone of Cold 3) and 0.129", baseMultiplier, baseCoefficient)
	}
	if multiplier, coefficient := mage.wintersChillBonus(target); math.Abs(multiplier-1.15) > 1e-9 || math.Abs(coefficient-0.21) > 1e-9 {
		t.Errorf("3 Winter's Chill stacks give Cone of Cold %.4f times the damage and %.3f more coefficient, want 1.15 and 0.21", multiplier, coefficient)
	}
	mage.fingersOfFrostChance = 0
	castNow(cone)
	if got := mage.wintersChillStacks(target); got != 0 {
		t.Errorf("%d Winter's Chill stacks after Cone of Cold, want 0", got)
	}
	if cone.DamageMultiplier != baseMultiplier || cone.BonusCoefficient != baseCoefficient {
		t.Errorf("Cone of Cold kept the stacks' bonus after the hit")
	}

	// Cold Snap finishes the cooldown.
	if cone.CD.IsReady(sim) {
		t.Fatalf("Cone of Cold is ready right after its cast")
	}
	coldSnap := mage.GetSpell(core.ActionID{SpellID: 12472})
	coldSnap.Cast(sim, &mage.Unit)
	if !cone.CD.IsReady(sim) {
		t.Errorf("Cone of Cold still on cooldown after Cold Snap")
	}

	// A Cone of Cold that grants Fingers of Frost spends none of the charges it granted, and one
	// cast with charges up spends one.
	mage.fingersOfFrostChance = 1
	advance(2 * time.Second)
	castNow(cone)
	advance(core.SpellBatchWindow)
	if got := mage.FingersOfFrostAura.GetStacks(); got != 2 {
		t.Errorf("%d Fingers of Frost charges after the Cone of Cold that granted them, want 2", got)
	}
	mage.fingersOfFrostChance = 0
	cone.CD.Reset()
	advance(2 * time.Second)
	castNow(cone)
	advance(core.SpellBatchWindow)
	if got := mage.FingersOfFrostAura.GetStacks(); got != 1 {
		t.Errorf("%d Fingers of Frost charges after a Cone of Cold cast with 2 up, want 1", got)
	}
}

func newMageSimWithItems(t *testing.T, items map[proto.ItemSlot]int32) (*core.Simulation, *Mage) {
	specs := make([]*proto.ItemSpec, 17)
	for i := range specs {
		specs[i] = &proto.ItemSpec{}
	}
	for slot, id := range items {
		specs[slot] = &proto.ItemSpec{Id: id}
	}
	player := &proto.Player{
		Race:      proto.Race_RaceGnome,
		Class:     proto.Class_ClassMage,
		Level:     60,
		Equipment: &proto.EquipmentSpec{Items: specs},
		Spec:      &proto.Player_Mage{Mage: &proto.Mage{Options: &proto.Mage_Options{}}},
		Rotation:  &proto.APLRotation{},
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

// Lord Sakrasis' Scepter smites the target with Holy Smite (1312330) on each harmful spell cast
// under Forever.
func TestForeverLordSakrasisScepter(t *testing.T) {
	sim, mage := newMageSimWithItems(t, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotOffHand: 5028})
	smite := mage.GetSpell(core.ActionID{SpellID: 1312330})
	if smite == nil {
		t.Fatal("no Holy Smite with Lord Sakrasis' Scepter")
	}
	topRank(t, "Fire Blast", mage.FireBlast).Cast(sim, mage.CurrentTarget)
	if got := smite.SpellMetrics[mage.CurrentTarget.UnitIndex].Casts; got != 1 {
		t.Errorf("Fire Blast cast %d Holy Smites, want 1", got)
	}
}

// Everlook Delivery Bot marks the target, and the 6th to 8th landed hit on it brings 467 Fire.
func TestForeverEverlookDeliveryBot(t *testing.T) {
	sim, mage := newMageSimWithItems(t, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotTrinket1: 274760})
	use := mage.GetSpell(core.ActionID{ItemID: 274760})
	delivery := mage.GetSpell(core.ActionID{SpellID: 1295633})
	if use == nil || delivery == nil {
		t.Fatal("no Everlook Delivery Bot use or delivery")
	}
	use.Cast(sim, mage.CurrentTarget)
	fireBlast := topRank(t, "Fire Blast", mage.FireBlast)
	hits := 0
	for delivery.SpellMetrics[mage.CurrentTarget.UnitIndex].Casts == 0 && hits < 20 {
		result := fireBlast.NewResult(mage.CurrentTarget)
		result.Outcome = core.OutcomeHit
		result.Damage = 100
		fireBlast.DealDamage(sim, result)
		hits++
	}
	if hits < 6 || hits > 8 {
		t.Errorf("the delivery came after %d hits, want 6 to 8", hits)
	}
}

// The High Order's Read Leyline is a 2 sec cast on a 2 min cooldown, and its Energized raises
// the Spirit part of the mana regen by 50% for 15 sec (development notes 2026-10-26).
func TestForeverReadLeyline(t *testing.T) {
	sim, mage := newMageWithRace(proto.Race_RaceSkyborneHighOrder)

	readLeyline := mage.GetSpell(core.ActionID{SpellID: 1259705})
	if readLeyline == nil {
		t.Fatal("a High Order mage has no Read Leyline")
	}
	if readLeyline.DefaultCast.CastTime != 2*time.Second || readLeyline.CD.Duration != 2*time.Minute {
		t.Errorf("Read Leyline casts in %v on a %v cooldown, want 2s and 2m", readLeyline.DefaultCast.CastTime, readLeyline.CD.Duration)
	}
	energized := mage.GetAura("Energized")
	if energized.Duration != 15*time.Second {
		t.Errorf("Energized lasts %v, want 15s", energized.Duration)
	}
	regen := mage.ManaRegenPerSecondWhileNotCasting()
	spirit := mage.SpiritManaRegenPerSecondDefault()
	energized.Activate(sim)
	if got := mage.ManaRegenPerSecondWhileNotCasting() - regen; math.Abs(got-spirit/2) > 1e-9 {
		t.Errorf("Energized added %.3f mana a second, want half the Spirit regen %.3f", got, spirit/2)
	}

	if _, gnome := newMageWithRace(proto.Race_RaceGnome); gnome.GetSpell(core.ActionID{SpellID: 1259705}) != nil {
		t.Error("a Gnome mage has Read Leyline")
	}
}

func newMageWithRace(race proto.Race) (*core.Simulation, *Mage) {
	player := &proto.Player{
		Race: race, Class: proto.Class_ClassMage, Level: 60, Equipment: &proto.EquipmentSpec{},
		Spec:     &proto.Player_Mage{Mage: &proto.Mage{Options: &proto.Mage_Options{}}},
		Rotation: &proto.APLRotation{},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*Mage)
}

// Ignite takes no percent damage modifier of its own under Forever: the crit it comes from
// already has them (development notes 2026-09-24, "no longer double dips on % damage increase
// modifiers"). Arcane Instability and Arcane Power raise Fireball but not the Ignite ticks.
func TestMageIgniteTakesNoDamagePercent(t *testing.T) {
	// The Arcane preset's tree (Arcane Instability 3, Arcane Power) and Ignite 5.
	sim, mage := newMageSim(t, "050205203100301531-005")
	fireball := topRank(t, "Fireball", mage.Fireball)
	at := mage.AttackTables[mage.CurrentTarget.UnitIndex][mage.Ignite.CastType]
	mage.ArcanePowerAura.Activate(sim)
	if got := fireball.DamageMultiplier; math.Abs(got-1.339) > 1e-9 {
		t.Errorf("Fireball does %.4f times the damage with Arcane Power and Arcane Instability 3, want 1.339", got)
	}
	if got := mage.Ignite.AttackerDamageMultiplier(at, true); got != 1 {
		t.Errorf("Ignite ticks for %.4f times their share with Arcane Power and Arcane Instability 3, want 1", got)
	}
}
