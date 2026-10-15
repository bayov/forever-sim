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
func TestForeverPetHitScalesWithSpeed(t *testing.T) {
	for _, c := range []struct {
		speed   proto.Hunter_Options_PetAttackSpeed
		seconds float64
	}{
		{proto.Hunter_Options_One, 1.0},
		{proto.Hunter_Options_OneFive, 1.5},
		{proto.Hunter_Options_Two, 2.0},
	} {
		_, hunter := newHunterPetSim(t, c.speed, proto.Ruleset_RulesetForever)
		mh := hunter.pet.AutoAttacks.MH()
		if math.Abs(mh.BaseDamageMin-18.17*c.seconds) > 1e-9 || math.Abs(mh.BaseDamageMax-27.66*c.seconds) > 1e-9 {
			t.Errorf("a %.1f sec pet hits for %.2f to %.2f, want %.2f to %.2f", c.seconds, mh.BaseDamageMin, mh.BaseDamageMax, 18.17*c.seconds, 27.66*c.seconds)
		}
		if mh.SwingSpeed != c.seconds {
			t.Errorf("the pet swings every %.1f sec, want %.1f", mh.SwingSpeed, c.seconds)
		}
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
