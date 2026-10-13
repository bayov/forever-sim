package dpswarrior

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/warrior"
)

// newForeverWarrior builds a naked level 60 Tauren warrior with no rotation (no racial procs), so
// only what a test does happens.
func newForeverWarrior(t *testing.T, talents string) (*core.Simulation, *warrior.Warrior) {
	player := &proto.Player{
		Race:          proto.Race_RaceTauren,
		Class:         proto.Class_ClassWarrior,
		Level:         60,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: talents,
		Rotation:      &proto.APLRotation{},
		Spec:          &proto.Player_Warrior{Warrior: &proto.Warrior{Options: &proto.Warrior_Options{}}},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*DpsWarrior).GetWarrior()
}

// newForeverWarriorArmed is newForeverWarrior with weapons: the synthetic Fury swords (991014,
// 2.6 sec, and 991015, 2.5 sec) or the synthetic two-hander (991016, 3.4 sec).
func newForeverWarriorArmed(t *testing.T, talents string, mainHand, offHand int32) (*core.Simulation, *warrior.Warrior) {
	// A boss that swings, for Rage from damage taken.
	target := core.NewDefaultTarget()
	target.MinBaseDamage = 3000
	target.SwingSpeed = 2
	items := make([]*proto.ItemSpec, 17)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: mainHand}
	items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: offHand}
	player := &proto.Player{
		Race:          proto.Race_RaceTauren,
		Class:         proto.Class_ClassWarrior,
		Level:         60,
		Equipment:     &proto.EquipmentSpec{Items: items},
		TalentsString: talents,
		Rotation:      &proto.APLRotation{},
		Spec:          &proto.Player_Warrior{Warrior: &proto.Warrior{Options: &proto.Warrior_Options{}}},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{target}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*DpsWarrior).GetWarrior()
}

// whiteRage is the Rage one white hit with this outcome and damage gives, from an empty bar.
func whiteRage(sim *core.Simulation, war *warrior.Warrior, auto *core.Spell, outcome core.HitOutcome, damage float64) float64 {
	war.SpendRage(sim, war.CurrentRage(), war.RageRefundMetrics)
	result := auto.NewResult(war.CurrentTarget)
	result.Outcome = outcome
	result.Damage = damage
	result.PreOutcomeDamage = damage
	auto.DealDamage(sim, result)
	return war.CurrentRage()
}

// TestForeverWarriorNormalizedRage checks standards 21: under Forever a landed white hit gives
// k × the weapon's swing speed in Rage, whatever its damage (Blizzard's Priest & Warrior deep dive:
// "Rage gained is now based on the speed of the weapon that dealt the damage"), with k = 3.46 for
// a one-hander and 4.5 for a two-hander, the off hand at the same rate, 100% more on a crit (the
// 2 October post, https://us.forums.blizzard.com/en/wow/t/2369360) and dodges and parries giving
// the same as a hit. Abilities and Deep Wounds give no Rage for their damage.
func TestForeverWarriorNormalizedRage(t *testing.T) {
	near := func(name string, got, want float64) {
		t.Helper()
		if math.Abs(got-want) > 1e-6 {
			t.Errorf("%s gave %.3f Rage, want %.3f", name, got, want)
		}
	}

	// At level 60 with the default s = 0.5, k is 1.5 times its level 30 value.
	const l60 = 1.5
	sim, war := newForeverWarriorArmed(t, "", 991014, 991015)
	mh, oh := war.AutoAttacks.MHAuto(), war.AutoAttacks.OHAuto()
	near("a 300 damage main hand hit", whiteRage(sim, war, mh, core.OutcomeHit, 300), l60*3.46*2.6)
	near("a 900 damage main hand hit", whiteRage(sim, war, mh, core.OutcomeHit, 900), l60*3.46*2.6)
	near("a main hand crit", whiteRage(sim, war, mh, core.OutcomeCrit, 600), 2*l60*3.46*2.6)
	near("a dodged main hand swing", whiteRage(sim, war, mh, core.OutcomeDodge, 0), l60*3.46*2.6)
	near("a parried main hand swing", whiteRage(sim, war, mh, core.OutcomeParry, 0), l60*3.46*2.6)
	near("a glancing main hand swing", whiteRage(sim, war, mh, core.OutcomeGlance, 200), l60*3.46*2.6)
	near("a missed main hand swing", whiteRage(sim, war, mh, core.OutcomeMiss, 0), 0)
	near("an off hand hit", whiteRage(sim, war, oh, core.OutcomeHit, 150), l60*3.46*2.5)

	// Dual Wield Specialization's off hand Rage (10% a point) stays on top.
	sim, war = newForeverWarriorArmed(t, "-00000005", 991014, 991015)
	near("an off hand hit with Dual Wield Specialization 5/5", whiteRage(sim, war, war.AutoAttacks.OHAuto(), core.OutcomeHit, 150), 1.5*l60*3.46*2.5)

	sim, war = newForeverWarriorArmed(t, "", 991016, 0)
	near("a two-hander hit", whiteRage(sim, war, war.AutoAttacks.MHAuto(), core.OutcomeHit, 500), l60*4.5*3.4)
	near("a two-hander crit", whiteRage(sim, war, war.AutoAttacks.MHAuto(), core.OutcomeCrit, 1000), 2*l60*4.5*3.4)

	// Only white hits make Rage: an ability's damage and Deep Wounds' ticks make none.
	sim, war = newForeverWarriorArmed(t, "00305003", 991014, 991015)
	near("a Hamstring hit", whiteRage(sim, war, war.Hamstring.Spell, core.OutcomeHit, 500), 0)
	near("a Deep Wounds hit", whiteRage(sim, war, war.DeepWounds.Spell, core.OutcomeHit, 500), 0)
}

// TestForeverWarriorNormalizedRageLevels checks standards 21's level scaling: k(L) = k30 × (1 + s ×
// (L − 30) / 30) above level 30, with the default s = 0.5: ×1 at 30 and below, ×1.25 at 45 and ×1.5
// at 60.
func TestForeverWarriorNormalizedRageLevels(t *testing.T) {
	if core.ForeverNormalizedRageScaling != 0.5 {
		t.Fatalf("the default s is %g, want 0.5", core.ForeverNormalizedRageScaling)
	}
	sword := &core.Weapon{SwingSpeed: 2.6, NormalizedSwingSpeed: 2.4}
	twoHander := &core.Weapon{SwingSpeed: 3.4, NormalizedSwingSpeed: 3.3}
	for _, c := range []struct {
		level  int32
		factor float64
	}{{20, 1}, {30, 1}, {45, 1.25}, {60, 1.5}} {
		if got, want := core.ForeverNormalizedRage(sword, c.level), c.factor*3.46*2.6; math.Abs(got-want) > 1e-9 {
			t.Errorf("a 2.6 sec one-hander at level %d gives %.4f Rage, want %.4f", c.level, got, want)
		}
		if got, want := core.ForeverNormalizedRage(twoHander, c.level), c.factor*4.5*3.4; math.Abs(got-want) > 1e-9 {
			t.Errorf("a 3.4 sec two-hander at level %d gives %.4f Rage, want %.4f", c.level, got, want)
		}
	}
}

// TestForeverWarriorDamageTakenRage checks that Rage from taking damage stays Classic's (standards
// 21 only changes Rage from dealing damage): damage × 2.5 / the conversion at the attacker's
// level. The damage comes from a level 60 special attack, which gives no Rage for dealing it.
func TestForeverWarriorDamageTakenRage(t *testing.T) {
	sim, war := newForeverWarriorArmed(t, "", 991014, 991015)
	war.SpendRage(sim, war.CurrentRage(), war.RageRefundMetrics)
	attack := war.Hamstring.Spell
	result := attack.NewResult(&war.Unit)
	result.Outcome = core.OutcomeHit
	result.Damage = 1000
	attack.DealDamage(sim, result)
	if got, want := war.CurrentRage(), 1000*2.5/core.GetRageConversion(60); math.Abs(got-want) > 1e-6 {
		t.Errorf("taking 1000 damage gave %.3f Rage, want %.3f", got, want)
	}
}

// TestForeverWarriorSpellValues checks the Forever client values (1.60.1.70338) that are easy to get
// wrong: Slam's cooldown and the talent Rage discounts.
func TestForeverWarriorSpellValues(t *testing.T) {
	// Improved Slam 2/2 and Improved Execute 2/2, Booming Voice 5/5 and Raging Blows.
	_, war := newForeverWarrior(t, "00000000000000200-50000000102")
	if got := war.Slam.CD.Duration; got != 15*time.Second {
		t.Errorf("Slam with Improved Slam has a %s cooldown, want 15s", got)
	}
	costs := []struct {
		name  string
		spell *warrior.WarriorSpell
		want  float64
	}{
		{"Execute", war.Execute, 10},
		{"Whirlwind", war.Whirlwind, 22},
		{"Cleave", war.Cleave, 17},
		{"Battle Shout", war.BattleShout, 7.5},
	}
	for _, c := range costs {
		if got := c.spell.DefaultCast.Cost; got != c.want {
			t.Errorf("%s costs %g Rage, want %g", c.name, got, c.want)
		}
	}

	_, war = newForeverWarrior(t, "")
	if got := war.Slam.CD.Duration; got != 18*time.Second {
		t.Errorf("Slam has a %s cooldown, want 18s", got)
	}
}

// TestForeverWarriorDeepWoundsAndFlurry checks two client 70338 values: Deep Wounds' damage spell
// (412613) cannot crit, so its ticks never do, and the Flurry buff (12966) lasts 15 sec.
func TestForeverWarriorDeepWoundsAndFlurry(t *testing.T) {
	_, war := newForeverWarrior(t, "30305013002-25053005112110501-2")
	if war.DeepWounds == nil {
		t.Fatal("the talents should give Deep Wounds")
	}
	if !war.DeepWounds.Flags.Matches(core.SpellFlagNoPeriodicCrit) {
		t.Error("Deep Wounds ticks should not be able to crit under Forever")
	}
	flurry := war.GetAura("Flurry Proc (12974)")
	if flurry == nil {
		t.Fatal("the talents should give 5 points of Flurry")
	}
	if flurry.Duration != 15*time.Second {
		t.Errorf("Flurry lasts %v, want 15s", flurry.Duration)
	}
}

// TestForeverWarriorDeepWoundsRolls checks the 8 October development notes: under Forever, Deep
// Wounds doesn't scale with attack power, and it is a rolling DoT, so a crit while it is up adds
// the damage not yet ticked to the new one.
func TestForeverWarriorDeepWoundsRolls(t *testing.T) {
	sim, war := newForeverWarrior(t, "00305003")
	war.AutoAttacks.MH().BaseDamageMin = 100
	war.AutoAttacks.MH().BaseDamageMax = 200
	mh := war.AutoAttacks.MHAuto()
	dot := war.DeepWounds.Dot(war.CurrentTarget)
	crit := func() float64 {
		result := mh.NewResult(war.CurrentTarget)
		result.Outcome = core.OutcomeCrit
		result.Damage = 300
		mh.DealDamage(sim, result)
		return dot.SnapshotBaseDamage
	}

	first := crit()
	if first <= 0 {
		t.Fatalf("a crit should apply Deep Wounds, its tick is %.3f", first)
	}
	if got := crit(); math.Abs(got-2*first) > 1e-6 {
		t.Errorf("a second crit before any tick gave a %.3f tick, want the first's 4 ticks rolled in: %.3f", got, 2*first)
	}

	dot.Cancel(sim)
	war.AddStatDynamic(sim, stats.AttackPower, 1000)
	if got := crit(); math.Abs(got-first) > 1e-6 {
		t.Errorf("with 1000 more attack power a crit gave a %.3f tick, want %.3f", got, first)
	}
}
