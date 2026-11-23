package dpswarrior

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/warrior"
)

// TestForeverWindfuryTotemFurySpecials checks Blizzard's hotfixes of November 13, 2026
// (t/2403040) on a dual wielding Fury warrior: "Windfury Totem is now correctly activated by all
// melee abilities ... Strikes from abilities that involve your offhand weapon like Whirlwind or
// Mutilate do not have a chance to activate Windfury. The main-hand strike from those abilities
// still does." Bloodthirst and Whirlwind's main hand strike roll it at 20%, Whirlwind's off-hand
// strike and the off-hand white swing never do (standards 17).
func TestForeverWindfuryTotemFurySpecials(t *testing.T) {
	const n = 400
	items := make([]*proto.ItemSpec, 17)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 991014}
	items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: 991015}
	player := &proto.Player{
		Race:          proto.Race_RaceTauren,
		Class:         proto.Class_ClassWarrior,
		Level:         60,
		Equipment:     &proto.EquipmentSpec{Items: items},
		TalentsString: P1Talents, // Bloodthirst
		Rotation:      &proto.APLRotation{},
		Spec:          &proto.Player_Warrior{Warrior: &proto.Warrior{Options: &proto.Warrior_Options{}}},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(player, &proto.PartyBuffs{},
			&proto.RaidBuffs{TotemWeaponBuff: proto.TotemWeaponBuff_TotemWeaponBuffWindfury}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 600, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	war := sim.Raid.Parties[0].Players[0].(*DpsWarrior).GetWarrior()
	totem, buff := war.GetAura("Windfury Totem Raid"), war.GetAura("Windfury Attack")
	if totem == nil || buff == nil || war.Bloodthirst == nil || war.Whirlwind == nil {
		t.Fatal("no Windfury Totem, Bloodthirst or Whirlwind on the warrior")
	}

	for _, c := range []struct {
		name  string
		spell *core.Spell
		rolls bool
	}{
		{"Bloodthirst", war.Bloodthirst.Spell, true},
		{"Whirlwind's main hand strike", war.Whirlwind.Spell, true},
		{"Whirlwind's off-hand strike", war.GetSpell(core.ActionID{SpellID: 1680}.WithTag(2)), false},
		{"off-hand white swing", war.AutoAttacks.OHAuto(), false},
	} {
		if c.spell == nil {
			t.Fatalf("%s isn't registered", c.name)
		}
		procs := 0
		for range n {
			sim.CurrentTime += 2 * time.Second
			buff.Deactivate(sim)
			result := c.spell.NewResult(war.CurrentTarget)
			result.Outcome = core.OutcomeHit
			totem.OnSpellHitDealt(totem, sim, c.spell, result)
			if buff.IsActive() {
				procs++
			}
		}
		if c.rolls && (procs < n/10 || procs > n/3) {
			t.Errorf("%s procced Windfury Totem %d times in %d landed hits, want about a fifth", c.name, procs, n)
		}
		if !c.rolls && procs != 0 {
			t.Errorf("%s procced Windfury Totem %d times in %d landed hits, want none", c.name, procs, n)
		}
	}
}

// TestForeverBloodthrillNotFromWindfuryTotemSwing: under Forever Windfury Totem's white extra
// swing doesn't proc Bloodthrill (wowsims/forever EV-0056 and EV-0136: 10 of 1,633 totem hits, no
// more than the ability before them gives alone), while another proc's extra swing (Hand of
// Justice) still does, as a white swing.
func TestForeverBloodthrillNotFromWindfuryTotemSwing(t *testing.T) {
	sim, war := newForeverWarriorArmed(t, armsPresetTalents, 991014, 991015)
	war.AutoAttacks.CancelAutoSwing(sim)
	target := war.CurrentTarget
	war.Rend.Dot(target).Apply(sim)
	procs := func(source int32) int {
		restore := war.AutoAttacks.SetExtraAttackSourceForTest(core.ActionID{SpellID: source})
		defer restore()
		n := 0
		for range 200 {
			war.BloodthrillAura.Deactivate(sim)
			result := war.AutoAttacks.MHAuto().NewResult(target)
			result.Outcome = core.OutcomeHit
			result.Damage = 100
			war.AutoAttacks.MHAuto().DealDamage(sim, result)
			if war.BloodthrillAura.IsActive() {
				n++
			}
		}
		return n
	}
	if n := procs(core.WindfuryBuffSpellId[3]); n != 0 {
		t.Errorf("Windfury Totem's extra swing procced Bloodthrill %d times in 200 hits, want none", n)
	}
	if procs(15601) == 0 { // Hand of Justice
		t.Error("Hand of Justice's extra swing never procced Bloodthrill in 200 hits")
	}
}

// TestForeverFlurryNotSpentByWindfuryTotemSwing: under Forever Windfury Totem's white extra swing
// spends no Flurry charge, and its crit still sets Flurry to 3 (wowsims/forever EV-0138 and
// EV-0140: 68 of 74 totem hits with Flurry up removed none, 22 of 22 totem crits applied or
// refreshed it). Another proc's extra swing still spends one.
func TestForeverFlurryNotSpentByWindfuryTotemSwing(t *testing.T) {
	sim, war := newForeverWarriorArmed(t, P1Talents, 991014, 991015)
	war.AutoAttacks.CancelAutoSwing(sim)
	flurry := war.GetAura("Flurry Proc (12972)")
	if flurry == nil {
		t.Fatal("the talents should give 3 points of Flurry")
	}
	swing := func(source int32, outcome core.HitOutcome) {
		restore := war.AutoAttacks.SetExtraAttackSourceForTest(core.ActionID{SpellID: source})
		defer restore()
		auto := war.AutoAttacks.MHAuto()
		result := auto.NewResult(war.CurrentTarget)
		result.Outcome = outcome
		result.Damage = 500
		result.PreOutcomeDamage = 500
		auto.DealDamage(sim, result)
	}
	totem := core.WindfuryBuffSpellId[3]

	flurry.Activate(sim)
	flurry.SetStacks(sim, 3)
	swing(totem, core.OutcomeHit)
	if got := flurry.GetStacks(); got != 3 {
		t.Errorf("a Windfury Totem hit left Flurry at %d charges, want 3", got)
	}
	flurry.SetStacks(sim, 1)
	swing(totem, core.OutcomeCrit)
	if got := flurry.GetStacks(); !flurry.IsActive() || got != 3 {
		t.Errorf("a Windfury Totem crit left Flurry at %d charges, want 3", got)
	}
	flurry.Deactivate(sim)
	swing(totem, core.OutcomeCrit)
	if got := flurry.GetStacks(); !flurry.IsActive() || got != 3 {
		t.Errorf("a Windfury Totem crit with Flurry down left it at %d charges, want 3", got)
	}
	swing(15601, core.OutcomeHit) // Hand of Justice
	if got := flurry.GetStacks(); got != 2 {
		t.Errorf("Hand of Justice's extra swing left Flurry at %d charges, want 2", got)
	}
}

// TestForeverExecuteRollsWindfuryTotemTwice: under Forever an Execute's cast rolls Windfury Totem
// on its own, whatever its damage does, and the landed damage rolls it again (wowsims/forever
// EV-0155, after hotfix push 114313: 1.68 and 2.09 times one roll's procs, where two rolls give
// about 1.81). Within the 100 ms cooldown that is 1 - 0.8 x (1 - 0.2 x landed), about a third of
// casts. DPS Report assumption warr_exec_wf2=0 rolls once, on the landed damage.
func TestForeverExecuteRollsWindfuryTotemTwice(t *testing.T) {
	const n = 4000
	rate := func(env string) float64 {
		core.AssumeForTest(env)
		t.Cleanup(func() { core.AssumeForTest("") })
		sim, war := newForeverWarriorBuffed(t, &proto.RaidBuffs{TotemWeaponBuff: proto.TotemWeaponBuff_TotemWeaponBuffWindfury})
		buff := war.GetAura("Windfury Attack")
		if buff == nil || war.Execute == nil {
			t.Fatal("no Windfury Totem or Execute on the warrior")
		}
		procs := 0
		for range n {
			sim.CurrentTime += 2 * time.Second
			buff.Deactivate(sim)
			war.Execute.ApplyEffects(sim, war.CurrentTarget, war.Execute.Spell)
			if buff.IsActive() {
				procs++
			}
		}
		return float64(procs) / n
	}
	twice, once := rate(""), rate("warr_exec_wf2=0")
	t.Logf("Windfury Totem procs a cast: %.3f with two rolls, %.3f with one", twice, once)
	if twice < 0.30 || twice > 0.38 {
		t.Errorf("Execute procced Windfury Totem on %.3f of casts, want about a third (two rolls)", twice)
	}
	if once < 0.13 || once > 0.21 {
		t.Errorf("with warr_exec_wf2=0 Execute procced Windfury Totem on %.3f of casts, want just under 0.2 (the casts that land)", once)
	}
}

// newForeverWarriorBuffed is a level 60 Tauren Fury warrior under Forever, dual wielding, with the
// raid buffs given.
func newForeverWarriorBuffed(t *testing.T, raidBuffs *proto.RaidBuffs) (*core.Simulation, *warrior.Warrior) {
	t.Helper()
	items := make([]*proto.ItemSpec, 17)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 991014}
	items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: 991015}
	player := &proto.Player{
		Race:          proto.Race_RaceTauren,
		Class:         proto.Class_ClassWarrior,
		Level:         60,
		Equipment:     &proto.EquipmentSpec{Items: items},
		TalentsString: P1Talents,
		Rotation:      &proto.APLRotation{},
		Spec:          &proto.Player_Warrior{Warrior: &proto.Warrior{Options: &proto.Warrior_Options{}}},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, raidBuffs, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 600, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*DpsWarrior).GetWarrior()
}

// TestForeverWindfuryTotemSwingAsWhite checks wowsims/forever's EV-0138 log (2026-11-16, build
// 70938 with hotfixes to push 114115): Windfury Totem's swing gives the Rage of a main-hand white
// hit (crit doubled), rolls Unbridled Wrath at the white swing's rate (61 of 103, 59%) and uses a
// Reckless Swings charge when it lands (14 charges), as the sim's white extra swing does.
func TestForeverWindfuryTotemSwingAsWhite(t *testing.T) {
	totem := core.ActionID{SpellID: core.WindfuryBuffSpellId[3]}
	sim, war := newForeverWarriorArmed(t, "", 991014, 991015)
	war.AutoAttacks.CancelAutoSwing(sim)
	auto := war.AutoAttacks.MHAuto()
	for _, outcome := range []core.HitOutcome{core.OutcomeHit, core.OutcomeCrit} {
		white := whiteRage(sim, war, auto, outcome, 500)
		restore := war.AutoAttacks.SetExtraAttackSourceForTest(totem)
		got := whiteRage(sim, war, auto, outcome, 500)
		restore()
		if white <= 0 || math.Abs(got-white) > 1e-9 {
			t.Errorf("a Windfury Totem swing (outcome %v) gives %.3f Rage, want a white hit's %.3f", outcome, got, white)
		}
	}

	sim, war = newForeverWarriorArmed(t, "30313013-250530055121100031", 991014, 991015) // P1Talents with Reckless Swings
	war.AutoAttacks.CancelAutoSwing(sim)
	auto = war.AutoAttacks.MHAuto()
	restore := war.AutoAttacks.SetExtraAttackSourceForTest(totem)
	defer restore()
	war.RecklessSwingsAura.Activate(sim)
	whiteRage(sim, war, auto, core.OutcomeHit, 500)
	if got := war.RecklessSwingsAura.GetStacks(); got != 3 {
		t.Errorf("a landed Windfury Totem swing left Reckless Swings at %d charges, want 3", got)
	}
	war.RecklessSwingsAura.Deactivate(sim)
	// Unbridled Wrath's share: the swings that gave more than the least any gave.
	const n = 4000
	rages := make([]float64, n)
	base := math.Inf(1)
	for i := range rages {
		rages[i] = whiteRage(sim, war, auto, core.OutcomeHit, 500)
		base = math.Min(base, rages[i])
	}
	procs := 0
	for _, rage := range rages {
		if rage > base+0.5 {
			procs++
		}
	}
	chance := 0.12 * float64(war.Talents.UnbridledWrath)
	if got := float64(procs) / n; math.Abs(got-chance) > 0.03 {
		t.Errorf("Unbridled Wrath came on %.3f of Windfury Totem swings, want %.2f", got, chance)
	}
}
