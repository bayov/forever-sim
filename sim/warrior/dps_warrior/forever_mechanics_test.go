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
// the damage not yet ticked to the new one. Either hand's crit takes the main hand's average weapon
// damage, and the damage modifiers come once, at the tick (wowsims/forever 3281b13bbe, 36cfc58328).
func TestForeverWarriorDeepWoundsRolls(t *testing.T) {
	sim, war := newForeverWarriorArmed(t, "00305003", 991014, 991015)
	war.AutoAttacks.CancelAutoSwing(sim)
	target := war.CurrentTarget
	dot := war.DeepWounds.Dot(target)
	crit := func(auto *core.Spell) float64 {
		result := auto.NewResult(target)
		result.Outcome = core.OutcomeCrit
		result.Damage = 300
		auto.DealDamage(sim, result)
		return dot.SnapshotBaseDamage
	}

	want := war.AutoAttacks.MH().AverageDamage() * 0.6 / 4
	war.PseudoStats.DamageDealtMultiplier *= 1.2
	if got := crit(war.AutoAttacks.MHAuto()); math.Abs(got-want) > 1e-6 {
		t.Errorf("a main hand crit stored a %.3f tick, want the main hand's weapon damage only: %.3f", got, want)
	}
	if got := crit(war.AutoAttacks.OHAuto()); math.Abs(got-2*want) > 1e-6 {
		t.Errorf("an off hand crit before any tick gave a %.3f tick, want the first's 4 ticks and the main hand's share: %.3f", got, 2*want)
	}

	metrics := &war.DeepWounds.SpellMetrics[target.UnitIndex]
	before := metrics.TotalDamage
	dot.TickOnce(sim)
	multiplier := war.DeepWounds.AttackerDamageMultiplier(war.AttackTables[target.UnitIndex][proto.CastType_CastTypeMainHand], true)
	// Forever rounds damage to whole numbers.
	if got := metrics.TotalDamage - before; math.Abs(got-2*want*multiplier) > 0.5 {
		t.Errorf("a tick dealt %.3f, want the stored %.3f times the modifiers once (%.3f): %.3f", got, 2*want, multiplier, 2*want*multiplier)
	}

	dot.Cancel(sim)
	war.AddStatDynamic(sim, stats.AttackPower, 1000)
	if got := crit(war.AutoAttacks.MHAuto()); math.Abs(got-want) > 1e-6 {
		t.Errorf("with 1000 more attack power a crit gave a %.3f tick, want %.3f", got, want)
	}
}

// TestForeverWarriorDeepWoundsKeepsTimer checks that under Forever a crit while Deep Wounds is up
// keeps the first application's 3 sec tick timer and gives it 4 ticks from the next one
// (wowsims/forever c4a5402aa9, from a Forever combat log).
func TestForeverWarriorDeepWoundsKeepsTimer(t *testing.T) {
	sim, war := newForeverWarrior(t, "00305003")
	war.AutoAttacks.CancelAutoSwing(sim)
	war.AutoAttacks.MH().BaseDamageMin = 100
	war.AutoAttacks.MH().BaseDamageMax = 200
	target := war.CurrentTarget
	dot := war.DeepWounds.Dot(target)
	crit := func(sim *core.Simulation) {
		mh := war.AutoAttacks.MHAuto()
		result := mh.NewResult(target)
		result.Outcome = core.OutcomeCrit
		result.Damage = 300
		mh.DealDamage(sim, result)
	}

	start := sim.CurrentTime
	crit(sim)
	sim.AddPendingAction(&core.PendingAction{NextActionAt: start + 2*time.Second, OnAction: crit})
	for !sim.Step() && sim.CurrentTime < start+2*time.Second {
	}
	if got, want := dot.NextTickAt(), start+3*time.Second; got != want {
		t.Errorf("after the second crit the next tick is at %s, want the first's timer: %s", got, want)
	}
	if got, want := dot.ExpiresAt(), start+12*time.Second; got != want {
		t.Errorf("after the second crit the bleed ends at %s, want 4 ticks from the next one: %s", got, want)
	}

	for !sim.Step() && sim.CurrentTime <= start+15*time.Second {
	}
	if got := war.DeepWounds.SpellMetrics[target.UnitIndex].Ticks; got != 4 {
		t.Errorf("the bleed ticked %d times, want 4", got)
	}
}

// TestForeverWarriorRendAttackPower checks report assumption warr_rend_ap (user ruling
// 2026-10-15): under Forever each Rend tick adds 2% of attack power to rank 7's 21, read when the
// tick lands with the damage modifiers of that moment (wowsims/forever 1172f1c745, 36cfc58328).
// Its ticks crit, with Impale (872e6def91).
func TestForeverWarriorRendAttackPower(t *testing.T) {
	sim, war := newForeverWarrior(t, "00305003032")
	target := war.CurrentTarget
	dot := war.Rend.Dot(target)
	dot.Apply(sim)
	metrics := &war.Rend.SpellMetrics[target.UnitIndex]
	tick := func() float64 {
		before := metrics.TotalDamage
		dot.TickOnce(sim)
		return metrics.TotalDamage - before
	}

	war.Rend.BonusCritRating -= 1000 * core.CritRatingPerCritChance
	for _, extra := range []float64{0, 1000} {
		war.AddStatDynamic(sim, stats.AttackPower, extra)
		war.PseudoStats.DamageDealtMultiplier *= 1.1
		ap := war.Rend.MeleeAttackPower(target)
		multiplier := war.Rend.AttackerDamageMultiplier(war.AttackTables[target.UnitIndex][proto.CastType_CastTypeMainHand], true)
		want := (21 + 0.02*ap) * multiplier
		if got := tick(); math.Abs(got-want) > 0.5 {
			t.Errorf("Rend at %.0f attack power ticks for %.3f, want %.3f", ap, got, want)
		}
	}

	if got := war.Rend.CritDamageBonus; got != 1.2 {
		t.Errorf("Rend's crit damage bonus is %.2f, want 1.2 with 2 points of Impale", got)
	}
	war.Rend.BonusCritRating += 2000 * core.CritRatingPerCritChance
	before := metrics.CritTicks
	tick()
	if metrics.CritTicks != before+1 {
		t.Error("a Rend tick at 100% crit chance didn't crit")
	}
}

// TestForeverWarriorBloodthrillSources checks that Bloodthrill procs only from main hand white
// swings, Heroic Strike and Cleave (wowsims/forever 788f62ce0f: Hamstring 0/99, Overpower 0/9,
// Execute 0/7, off hand 0/200 in their logs).
func TestForeverWarriorBloodthrillSources(t *testing.T) {
	sim, war := newForeverWarriorArmed(t, armsPresetTalents, 991014, 991015)
	war.AutoAttacks.CancelAutoSwing(sim)
	target := war.CurrentTarget
	war.Rend.Dot(target).Apply(sim)
	procs := func(spell *core.Spell) int {
		n := 0
		for i := 0; i < 200; i++ {
			war.BloodthrillAura.Deactivate(sim)
			result := spell.NewResult(target)
			result.Outcome = core.OutcomeHit
			result.Damage = 100
			spell.DealDamage(sim, result)
			if war.BloodthrillAura.IsActive() {
				n++
			}
		}
		return n
	}
	for _, spell := range []*core.Spell{war.AutoAttacks.MHAuto(), war.HeroicStrike.Spell, war.Cleave.Spell} {
		if procs(spell) == 0 {
			t.Errorf("%s never procced Bloodthrill in 200 hits", spell.ActionID)
		}
	}
	for _, spell := range []*core.Spell{war.AutoAttacks.OHAuto(), war.Hamstring.Spell, war.Overpower.Spell, war.MortalStrike.Spell} {
		if n := procs(spell); n != 0 {
			t.Errorf("%s procced Bloodthrill %d times in 200 hits, want none", spell.ActionID, n)
		}
	}
}

// TestForeverWarriorSweepingStrikesAndCharge checks two client rows: Sweeping Strikes lasts 20 sec
// (12292), and Charge gives 15 Rage (11578), 3 more a point of Improved Charge (12285), before the
// pull only.
func TestForeverWarriorSweepingStrikesAndCharge(t *testing.T) {
	sim, war := newForeverWarrior(t, "00020000000010001")
	if ss := war.GetAura("Sweeping Strikes"); ss == nil || ss.Duration != 20*time.Second {
		t.Errorf("Sweeping Strikes should last 20s, got %v", ss)
	}
	if war.Charge == nil {
		t.Fatal("no Charge at 60")
	}
	war.Stance = warrior.BattleStance
	war.SpendRage(sim, war.CurrentRage(), war.RageRefundMetrics)
	if !war.Charge.Cast(sim, war.CurrentTarget) {
		t.Fatal("Charge should cast in Battle Stance before the pull")
	}
	if got := war.CurrentRage(); got != 21 {
		t.Errorf("Charge with 2 points of Improved Charge gave %.0f Rage, want 21", got)
	}
	sim.AddPendingAction(&core.PendingAction{NextActionAt: 20 * time.Second, OnAction: func(*core.Simulation) {}})
	for !sim.Step() && sim.CurrentTime < 20*time.Second {
	}
	if war.Charge.CanCast(sim, war.CurrentTarget) {
		t.Error("Charge should not be usable in combat")
	}
}

// landedRange casts the spell's effects n times and returns the lowest and highest damage of its
// plain hits (no crit, block or avoid).
func landedRange(sim *core.Simulation, spell *core.Spell, target *core.Unit, n int) (float64, float64) {
	low, high := math.Inf(1), math.Inf(-1)
	metrics := &spell.SpellMetrics[target.UnitIndex]
	for i := 0; i < n; i++ {
		before := *metrics
		spell.ApplyEffects(sim, target, spell)
		if metrics.Hits == before.Hits || metrics.Crits != before.Crits || metrics.Blocks != before.Blocks {
			continue
		}
		damage := metrics.TotalDamage - before.TotalDamage
		low, high = min(low, damage), max(high, damage)
	}
	return low, high
}

// TestForeverWarriorProtectionValues checks the Protection rows under Forever (client 70425):
// Revenge rank 5 rolls 97 to 145 plus 28% of attack power (11601; wowsims/forever 3a6c7df46e),
// Shield Slam 625 to 685 plus the block value once and no attack power (23925), Thunder Clap 103
// plus 7% of attack power (11581; 1172f1c745), Demoralizing Shout lasts 45 sec (11556), and
// Bastion raises Physical damage only (16538).
func TestForeverWarriorProtectionValues(t *testing.T) {
	sim, war := newForeverWarriorArmed(t, "--251533120331221351", 991014, 19349)
	war.AutoAttacks.CancelAutoSwing(sim)
	target := war.CurrentTarget
	target.AddStatDynamic(sim, stats.Armor, -target.GetStat(stats.Armor))
	war.AddStatDynamic(sim, stats.AttackPower, 1000)
	at := war.AttackTables[target.UnitIndex][proto.CastType_CastTypeMainHand]

	if got := war.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical]; math.Abs(got-1.1) > 1e-9 {
		t.Errorf("Bastion 5/5 with a shield raises Physical damage by %.3f, want 1.1", got)
	}
	if got := war.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexFire]; got != 1 {
		t.Errorf("Bastion raises Fire damage by %.3f, want nothing", got)
	}

	check := func(name string, spell *core.Spell, low, high, apCoef float64) {
		t.Helper()
		spell.BonusCritRating -= 200 * core.CritRatingPerCritChance
		ap := spell.MeleeAttackPower(target)
		m := spell.AttackerDamageMultiplier(at, false)
		wantLow, wantHigh := (low+apCoef*ap)*m, (high+apCoef*ap)*m
		gotLow, gotHigh := landedRange(sim, spell, target, 600)
		// Forever rounds damage to whole numbers (before the last multipliers, so within 1), and 600
		// rolls come within 1% of the ends.
		slack := 1 + 0.01*(wantHigh-wantLow)
		if gotLow < wantLow-1 || gotLow > wantLow+slack || gotHigh > wantHigh+1 || gotHigh < wantHigh-slack {
			t.Errorf("%s at %.0f attack power hits for %.1f to %.1f, want %.1f to %.1f", name, ap, gotLow, gotHigh, wantLow, wantHigh)
		}
	}
	check("Revenge", war.Revenge.Spell, 97, 145, 0.28)
	check("Shield Slam", war.ShieldSlam.Spell, 625+war.BlockValue(), 685+war.BlockValue(), 0)
	check("Thunder Clap", war.ThunderClap.Spell, 103, 103, 0.07)

	if got := war.Revenge.FlatThreatBonus; got != 0 {
		t.Errorf("Revenge has %.0f flat threat, want none on top of 2.25 times its damage", got)
	}
	if got := war.DemoralizingShoutAuras.Get(target).Duration; got != 45*time.Second {
		t.Errorf("Demoralizing Shout lasts %s, want 45s", got)
	}
}

// TestForeverWarriorTwoHandedSpecialization checks that under Forever Two-Handed Weapon
// Specialization raises all Physical damage with a two-hander, Rend included, and nothing else
// (client 70425, spell 12163).
func TestForeverWarriorTwoHandedSpecialization(t *testing.T) {
	_, war := newForeverWarriorArmed(t, "0000000003", 991016, 0)
	if got := war.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical]; math.Abs(got-1.03) > 1e-9 {
		t.Errorf("3 points with a two-hander raise Physical damage by %.3f, want 1.03", got)
	}
	if got := war.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexFire]; got != 1 {
		t.Errorf("3 points raise Fire damage by %.3f, want nothing", got)
	}
	if got := war.Rend.DamageMultiplier; got != 1 {
		t.Errorf("Rend's own multiplier is %.3f; the talent belongs on the school", got)
	}
}
