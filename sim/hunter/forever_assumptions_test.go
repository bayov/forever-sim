package hunter

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

// The report assumptions the user ruled on 2026-10-15 (hunter_shot_ap_shares, hunter_raptor_31plus):
// Arcane Shot takes 10% of ranged attack power and Serpent Sting 3% a tick (the forever-hunter
// wiki), and Raptor Strike rank 8 adds 96 (the client's 70 raised halfway to 123).
func TestHunterShotSharesAndRaptorStrike(t *testing.T) {
	if arcaneShotRAPCoefficient != 0.10 {
		t.Errorf("Arcane Shot takes %v of ranged attack power, want 0.10", arcaneShotRAPCoefficient)
	}
	if serpentStingRAPPerTick != 0.03 {
		t.Errorf("Serpent Sting takes %v of ranged attack power a tick, want 0.03", serpentStingRAPPerTick)
	}
	for rank, want := range map[int]float64{7: 55, 8: 96} {
		if got := RaptorStrikeBaseDamage[rank]; got != want {
			t.Errorf("Raptor Strike rank %d adds %v, want %v", rank, got, want)
		}
	}
}

// newHunterPetSim builds a level 60 hunter with a cat at the given attack speed, under the ruleset.
func newHunterPetSim(t *testing.T, speed proto.Hunter_Options_PetAttackSpeed, ruleset proto.Ruleset) (*core.Simulation, *Hunter) {
	player := &proto.Player{
		Class:     proto.Class_ClassHunter,
		Race:      proto.Race_RaceHuman,
		Level:     60,
		Equipment: &proto.EquipmentSpec{},
		Spec: &proto.Player_Hunter{Hunter: &proto.Hunter{Options: &proto.Hunter_Options{
			PetType:        proto.Hunter_Options_Cat,
			PetUptime:      1,
			PetAttackSpeed: speed,
		}}},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: ruleset, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(HunterAgent).GetHunter()
}

// Report assumption hunter_pet_speed (user ruling 2026-10-15): under Forever a pet's white hit
// scales with its attack speed as in Classic, so a 1.0 sec pet hits half as hard as a 2.0 sec pet.
// A 2.0 sec pet's hit is the creature dps for level 60 over 2 sec, 20% either side (wowsims/forever
// 643d636345). Classic keeps its 18.17 to 27.66 a second.
func TestForeverPetHitScalesWithSpeed(t *testing.T) {
	for _, c := range []struct {
		speed   proto.Hunter_Options_PetAttackSpeed
		seconds float64
	}{
		{proto.Hunter_Options_One, 1.0},
		{proto.Hunter_Options_OneFive, 1.5},
		{proto.Hunter_Options_Two, 2.0},
	} {
		for _, r := range []struct {
			ruleset  proto.Ruleset
			min, max float64
		}{
			{proto.Ruleset_RulesetForever, 60.4529 * 0.8, 60.4529 * 1.2},
			{proto.Ruleset_RulesetClassic, 18.17, 27.66},
		} {
			_, hunter := newHunterPetSim(t, c.speed, r.ruleset)
			mh := hunter.pet.AutoAttacks.MH()
			if math.Abs(mh.BaseDamageMin-r.min*c.seconds) > 1e-9 || math.Abs(mh.BaseDamageMax-r.max*c.seconds) > 1e-9 {
				t.Errorf("ruleset %v: a %.1f sec pet hits for %.2f to %.2f, want %.2f to %.2f", r.ruleset, c.seconds, mh.BaseDamageMin, mh.BaseDamageMax, r.min*c.seconds, r.max*c.seconds)
			}
			if mh.SwingSpeed != c.seconds {
				t.Errorf("ruleset %v: the pet swings every %.1f sec, want %.1f", r.ruleset, mh.SwingSpeed, c.seconds)
			}
		}
	}
}

// Under Forever a pet's attack power is its own 30% of the creature dps as attack power (254 at 60)
// and 10% of the hunter's ranged attack power, with none from Strength (wowsims/forever 643d636345).
// Its crit is the hunter's with none from Agility (e56aaeb1a9, 684209eb0a), and it takes the
// hunter's hit as melee hit and 17/8 of it as spell hit (9ce48b6aba). Classic's pet keeps 2 attack
// power a point of Strength and its Agility crit.
func TestForeverPetStats(t *testing.T) {
	sim, hunter := newHunterPetSim(t, proto.Hunter_Options_Two, proto.Ruleset_RulesetForever)
	pet := hunter.pet
	if got, want := pet.GetStat(stats.AttackPower), 0.3*14*60.4529+0.1*hunter.GetStat(stats.RangedAttackPower); math.Abs(got-want) > 1e-6 {
		t.Errorf("pet attack power %.2f, want %.2f", got, want)
	}
	if got, want := pet.GetStat(stats.MeleeCrit), hunter.GetStat(stats.MeleeCrit); math.Abs(got-want) > 1e-9 {
		t.Errorf("pet crit %.4f%%, want the hunter's %.4f%%", got, want)
	}

	ap, crit := pet.GetStat(stats.AttackPower), pet.GetStat(stats.MeleeCrit)
	pet.AddStatDynamic(sim, stats.Strength, 100)
	pet.AddStatDynamic(sim, stats.Agility, 100)
	if got := pet.GetStat(stats.AttackPower); got != ap {
		t.Errorf("100 pet Strength moves its attack power from %.2f to %.2f, want no change", ap, got)
	}
	if got := pet.GetStat(stats.MeleeCrit); got != crit {
		t.Errorf("100 pet Agility moves its crit from %.4f to %.4f, want no change", crit, got)
	}

	melee, spell := pet.GetStat(stats.MeleeHit), pet.GetStat(stats.SpellHit)
	hunter.AddStatDynamic(sim, stats.MeleeHit, 4*core.MeleeHitRatingPerHitChance)
	if got := pet.GetStat(stats.MeleeHit) - melee; math.Abs(got-4*core.MeleeHitRatingPerHitChance) > 1e-9 {
		t.Errorf("4%% hunter hit gives the pet %.4f melee hit, want 4%%", got)
	}
	if got := pet.GetStat(stats.SpellHit) - spell; math.Abs(got-8.5*core.SpellHitRatingPerHitChance) > 1e-9 {
		t.Errorf("4%% hunter hit gives the pet %.4f spell hit, want 8.5%%", got)
	}

	sim, hunter = newHunterPetSim(t, proto.Hunter_Options_Two, proto.Ruleset_RulesetClassic)
	ap = hunter.pet.GetStat(stats.AttackPower)
	hunter.pet.AddStatDynamic(sim, stats.Strength, 100)
	if got := hunter.pet.GetStat(stats.AttackPower) - ap; math.Abs(got-200) > 1e-9 {
		t.Errorf("Classic: 100 pet Strength gives %.2f attack power, want 200", got)
	}
}

// Lone Wolf with a pet out gives the hunter nothing and divides the pet's damage by 1.2
// (wowsims/forever 56e9bc08e0).
func TestLoneWolfWithAPet(t *testing.T) {
	_, hunter := newHunterSim(t, LoneWolfTalents, &proto.Debuffs{})
	if !hunter.Talents.LoneWolf {
		t.Fatal("LoneWolfTalents has no Lone Wolf")
	}
	if got := hunter.PseudoStats.DamageDealtMultiplier; got != 1 {
		t.Errorf("hunter damage multiplier %.4f with a pet out, want 1", got)
	}
	if got, want := hunter.pet.PseudoStats.DamageDealtMultiplier, 1.25/1.2; math.Abs(got-want) > 1e-9 {
		t.Errorf("pet damage multiplier %.4f, want happiness's 1.25 / 1.2 = %.4f", got, want)
	}
}

// Report assumption hunter_crit_per_agi (user ruling 2026-10-15): under Forever the hunter gets 22%
// more crit per Agility than Classic, and Classic's table is unchanged.
func TestForeverHunterCritPerAgility(t *testing.T) {
	for _, c := range []struct {
		ruleset proto.Ruleset
		scale   float64
	}{
		{proto.Ruleset_RulesetForever, 1.22},
		{proto.Ruleset_RulesetClassic, 1},
	} {
		sim, hunter := newHunterPetSim(t, proto.Hunter_Options_Two, c.ruleset)
		before := hunter.GetStat(stats.MeleeCrit)
		hunter.AddStatDynamic(sim, stats.Agility, 100)
		got := (hunter.GetStat(stats.MeleeCrit) - before) / core.CritRatingPerCritChance
		want := 100 * core.CritPerAgi(proto.Class_ClassHunter, 60) * c.scale
		if math.Abs(got-want) > 1e-9 {
			t.Errorf("ruleset %v: 100 Agility gives %.4f%% crit, want %.4f%%", c.ruleset, got, want)
		}
	}
}

// Summon Hawk's hawk, wowsims/forever's log fit (81651da664, b89878e6f9, 0fc13fcb30): a 2.5 sec swing
// of 0.0815 of the creature dps over the swing before Unleashed Fury, a third either side, on 80% of
// the hunter's ranged attack power at 112 × 1.15 a point of dps; half the hunter's crit plus
// Ferocity whole, its hit; Unleashed Fury and the hunter's damage done at the summon. It swings on
// arrival and follows the hunter's ranged speed but not its attack speed, and a third hawk replaces
// the one with the least time left.
func TestSummonHawkHawks(t *testing.T) {
	sim, hunter := newHunterSim(t, BeastMasteryHawkTalents, &proto.Debuffs{})
	if !hunter.Talents.SummonHawk || len(hunter.hawks) != 2 {
		t.Fatalf("Summon Hawk %v with %d hawks, want 2", hunter.Talents.SummonHawk, len(hunter.hawks))
	}
	target := hunter.CurrentTarget
	summon := func() { hunter.SummonHawk.ApplyEffects(sim, target, hunter.SummonHawk) }
	stepTo := func(at time.Duration) {
		for sim.CurrentTime < at && !sim.Step() {
		}
	}
	lifetime := func(i int) *core.Dot {
		return hunter.GetSpell(core.ActionID{SpellID: 1293527, Tag: int32(i + 1)}).AOEDot()
	}
	hawk1, hawk2 := hunter.hawks[0], hunter.hawks[1]

	mh := hawk1.AutoAttacks.MH()
	if avg := 0.0815 * 60.4529 * 2.5 / 1.15; math.Abs(mh.BaseDamageMin-avg*2/3) > 1e-9 || math.Abs(mh.BaseDamageMax-avg*4/3) > 1e-9 {
		t.Errorf("hawk swing %.2f to %.2f, want %.2f to %.2f", mh.BaseDamageMin, mh.BaseDamageMax, avg*2/3, avg*4/3)
	}
	if mh.SwingSpeed != 2.5 || math.Abs(mh.AttackPowerPerDPS-112*1.15) > 1e-9 {
		t.Errorf("hawk swings every %v sec at %v attack power a point of dps, want 2.5 and 128.8", mh.SwingSpeed, mh.AttackPowerPerDPS)
	}

	summon()
	if !hawk1.IsEnabled() || hawk2.IsEnabled() {
		t.Fatalf("after one Summon Hawk the hawks are out %v %v, want the first only", hawk1.IsEnabled(), hawk2.IsEnabled())
	}
	if got, want := hawk1.GetStat(stats.AttackPower), 0.8*hunter.GetStat(stats.RangedAttackPower); math.Abs(got-want) > 1e-6 {
		t.Errorf("hawk attack power %.2f, want %.2f", got, want)
	}
	if got, want := hawk1.GetStat(stats.MeleeCrit), 0.5*hunter.GetStat(stats.MeleeCrit)+10*core.CritRatingPerCritChance; math.Abs(got-want) > 1e-6 {
		t.Errorf("hawk crit %.4f, want half the hunter's plus Ferocity's 10%%, %.4f", got, want)
	}
	if got, want := hawk1.GetStat(stats.MeleeHit), hunter.GetStat(stats.MeleeHit); got != want {
		t.Errorf("hawk hit %.4f, want the hunter's %.4f", got, want)
	}
	if got, want := hawk1.PseudoStats.DamageDealtMultiplier, 1.15*hunter.PseudoStats.DamageDealtMultiplier; math.Abs(got-want) > 1e-9 {
		t.Errorf("hawk damage multiplier %.4f, want Unleashed Fury times the hunter's, %.4f", got, want)
	}
	if hawk1.AutoAttacks.MainhandSwingAt() > sim.CurrentTime {
		t.Errorf("the hawk first swings at %v, want on arrival at %v", hawk1.AutoAttacks.MainhandSwingAt(), sim.CurrentTime)
	}
	rap := hunter.GetStat(stats.RangedAttackPower)
	hunter.AddStatDynamic(sim, stats.RangedAttackPower, 100)
	if got := hawk1.GetStat(stats.AttackPower) - 0.8*rap; math.Abs(got-80) > 1e-6 {
		t.Errorf("100 hunter ranged attack power gives the hawk %.2f attack power, want 80", got)
	}
	hunter.AddStatDynamic(sim, stats.RangedAttackPower, -100)

	quiver := hunter.PseudoStats.RangedSpeedMultiplier
	if quiver < 1.15 {
		t.Fatalf("hunter ranged speed %v, want the 15%% quiver at least", quiver)
	}
	if got := hawk1.SwingSpeed(); math.Abs(got-quiver) > 1e-9 {
		t.Errorf("hawk speed %.4f, want the hunter's ranged %.4f", got, quiver)
	}
	hunter.MultiplyRangedSpeed(sim, 1.3)
	hunter.MultiplyAttackSpeed(sim, 1.1)
	if got := hawk1.SwingSpeed(); math.Abs(got-quiver*1.3) > 1e-9 {
		t.Errorf("hawk speed %.4f with Quick Shots and Berserking, want the ranged %.4f", got, quiver*1.3)
	}

	stepTo(time.Second)
	summon()
	if got := hawk2.SwingSpeed(); !hawk2.IsEnabled() || math.Abs(got-quiver*1.3) > 1e-9 {
		t.Errorf("second hawk out %v at speed %.4f, want out at %.4f", hawk2.IsEnabled(), got, quiver*1.3)
	}
	hunter.MultiplyAttackSpeed(sim, 1/1.1)
	hunter.MultiplyRangedSpeed(sim, 1/1.3)
	for i, hawk := range hunter.hawks {
		if got := hawk.SwingSpeed(); math.Abs(got-quiver) > 1e-9 {
			t.Errorf("hawk %d speed %.4f after the hastes fade, want %.4f", i+1, got, quiver)
		}
	}

	stepTo(2 * time.Second)
	left2 := lifetime(1).RemainingDuration(sim)
	summon()
	if got := lifetime(0).RemainingDuration(sim); got != hawkDuration || lifetime(1).RemainingDuration(sim) != left2 {
		t.Errorf("a third hawk left %v and %v, want the first replaced (18s) and the second at %v", got, lifetime(1).RemainingDuration(sim), left2)
	}
	if !hawk1.IsEnabled() || hawk1.AutoAttacks.MainhandSwingAt() > sim.CurrentTime {
		t.Errorf("the replaced hawk is out %v and swings at %v, want out and swinging at %v", hawk1.IsEnabled(), hawk1.AutoAttacks.MainhandSwingAt(), sim.CurrentTime)
	}

	stepTo(25 * time.Second)
	for i, hawk := range hunter.hawks {
		if hawk.IsEnabled() {
			t.Errorf("hawk %d still out after 18 sec", i+1)
		}
		if got := hawk.SwingSpeed(); math.Abs(got-1) > 1e-9 {
			t.Errorf("hawk %d keeps speed %.4f once gone, want 1", i+1, got)
		}
		if got := hawk.PseudoStats.DamageDealtMultiplier; math.Abs(got-1.15) > 1e-9 {
			t.Errorf("hawk %d keeps damage multiplier %.4f once gone, want Unleashed Fury's 1.15", i+1, got)
		}
	}
}
