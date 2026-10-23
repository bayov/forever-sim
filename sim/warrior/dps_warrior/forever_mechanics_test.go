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

// newForeverWarriorArmed is newForeverWarrior with weapons: the synthetic Fury main hand and off
// hand (991014, a 2.8 sec sword, and 991015, a 1.7 sec axe) or the synthetic Arms two-hander
// (991016, a 3.7 sec axe).
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
// a one-hander and 4.5 for a two-hander, the off hand at half the rate (measured, see
// core.ForeverNormalizedRageOffHand) and 100% more on a crit (the 2 October post,
// https://us.forums.blizzard.com/en/wow/t/2369360). A two-hander's crit adds 100% of the swing's
// one-hand-rate Rage, so (4.5 + 3.46) × speed (wowsims/forever 5ac4ffbcb6, Rule of Rage 1322574).
// Misses, dodges and parries give nothing (wowsims/forever f9f9f21883, level 8 and 9 logs).
// Abilities and Deep Wounds give no Rage for their damage.
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
	near("a 300 damage main hand hit", whiteRage(sim, war, mh, core.OutcomeHit, 300), l60*3.46*2.8)
	near("a 900 damage main hand hit", whiteRage(sim, war, mh, core.OutcomeHit, 900), l60*3.46*2.8)
	near("a main hand crit", whiteRage(sim, war, mh, core.OutcomeCrit, 600), 2*l60*3.46*2.8)
	near("a dodged main hand swing", whiteRage(sim, war, mh, core.OutcomeDodge, 0), 0)
	near("a parried main hand swing", whiteRage(sim, war, mh, core.OutcomeParry, 0), 0)
	near("a glancing main hand swing", whiteRage(sim, war, mh, core.OutcomeGlance, 200), l60*3.46*2.8)
	near("a missed main hand swing", whiteRage(sim, war, mh, core.OutcomeMiss, 0), 0)
	near("an off hand hit", whiteRage(sim, war, oh, core.OutcomeHit, 150), 0.5*l60*3.46*2.9)
	near("an off hand crit", whiteRage(sim, war, oh, core.OutcomeCrit, 300), 2*0.5*l60*3.46*2.9)

	// Dual Wield Specialization's off hand Rage (10% a point) stays on top.
	sim, war = newForeverWarriorArmed(t, "-00000005", 991014, 991015)
	near("an off hand hit with Dual Wield Specialization 5/5", whiteRage(sim, war, war.AutoAttacks.OHAuto(), core.OutcomeHit, 150), 1.5*0.5*l60*3.46*2.9)

	sim, war = newForeverWarriorArmed(t, "", 991016, 0)
	near("a two-hander hit", whiteRage(sim, war, war.AutoAttacks.MHAuto(), core.OutcomeHit, 500), l60*4.5*3.7)
	near("a two-hander crit", whiteRage(sim, war, war.AutoAttacks.MHAuto(), core.OutcomeCrit, 1000), l60*(4.5+3.46)*3.7)

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

// TestForeverWarriorDamageTakenRage checks Forever's Rage from taking damage (standards 21):
// 2.5 × (1 − A) × damage / the Rage conversion at the attacker's level, with A the armor assumed
// for a level 60 attacker (39.5%, between 38% at 50 and 40% at 63), so 6.6 Rage for 1000 damage,
// where Classic's damage × 2.5 / the same conversion gives 10.8 (wowsims/forever ee2d69b16a,
// 73b317c7cf, de550b7df9). The damage is the attacker's, before anything of ours: our damage
// taken modifiers (Defensive Stance, Shield Wall), a block and an absorb take no Rage away
// (report assumption warr_dt_rage_mods, None). A crushing blow counts at 1.5, a periodic tick
// like a hit, and Berserker Rage doubles it. The damage comes from a level 60 special attack,
// which gives no Rage for dealing it.
func TestForeverWarriorDamageTakenRage(t *testing.T) {
	perPoint := 2.5 * (1 - (0.38 + 0.02*10/13)) / core.GetRageConversion(60)
	cases := []struct {
		name       string
		outcome    core.HitOutcome
		taken      float64 // the damage after all of our mitigation
		takenMulti float64 // our damage taken multiplier while the hit lands
		periodic   bool
		berserker  bool    // Berserker Rage up
		want       float64 // Rage, for 1000 attacker damage
	}{
		{"a hit", core.OutcomeHit, 1000, 1, false, false, 1000 * perPoint},
		{"a hit in Defensive Stance", core.OutcomeHit, 900, 0.9, false, false, 1000 * perPoint},
		{"a hit under Shield Wall", core.OutcomeHit, 250, 0.25, false, false, 1000 * perPoint},
		{"a hit fully absorbed", core.OutcomeHit, 0, 1, false, false, 1000 * perPoint},
		{"a blocked hit", core.OutcomeBlock, 700, 1, false, false, 1000 * perPoint},
		{"a crushing blow", core.OutcomeCrush, 1500, 1, false, false, 1500 * perPoint},
		{"a dodge", core.OutcomeDodge, 0, 1, false, false, 0},
		{"a periodic tick", core.OutcomeHit, 1000, 1, true, false, 1000 * perPoint},
		{"a hit under Berserker Rage", core.OutcomeHit, 1000, 1, false, true, 2 * 1000 * perPoint},
	}
	for _, c := range cases {
		sim, war := newForeverWarriorArmed(t, "", 991014, 991015)
		if c.berserker {
			war.BerserkerRageAura.Activate(sim)
		}
		war.SpendRage(sim, war.CurrentRage(), war.RageRefundMetrics)
		war.PseudoStats.DamageTakenMultiplier *= c.takenMulti
		attack := war.Hamstring.Spell
		result := attack.NewResult(&war.Unit)
		result.Outcome = c.outcome
		result.AttackerDamage = 1000
		result.Damage = c.taken
		if c.periodic {
			attack.DealPeriodicDamage(sim, result)
		} else {
			attack.DealDamage(sim, result)
		}
		if got := war.CurrentRage(); math.Abs(got-c.want) > 1e-6 {
			t.Errorf("%s gave %.3f Rage, want %.3f", c.name, got, c.want)
		}
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

// TestForeverWarriorRendAttackPower checks report assumption warr_rend_ap (user ruling
// 2026-10-15): under Forever each Rend tick adds 2% of attack power to rank 4's 21.
func TestForeverWarriorRendAttackPower(t *testing.T) {
	sim, war := newForeverWarrior(t, "")
	target := war.CurrentTarget
	tick := func() float64 {
		dot := war.Rend.Dot(target)
		dot.Apply(sim)
		got := dot.SnapshotBaseDamage
		dot.Cancel(sim)
		return got
	}
	for _, extra := range []float64{0, 1000} {
		war.AddStatDynamic(sim, stats.AttackPower, extra)
		ap := war.Rend.MeleeAttackPower(target)
		want := (21 + 0.02*ap) * war.Rend.BaseDamageMultiplierAdditive
		if got := tick(); math.Abs(got-want) > 1e-6 {
			t.Errorf("Rend at %.0f attack power ticks for %.3f, want %.3f", ap, got, want)
		}
	}
}
