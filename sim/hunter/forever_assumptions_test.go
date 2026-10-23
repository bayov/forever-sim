package hunter

import (
	"math"
	"testing"

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
