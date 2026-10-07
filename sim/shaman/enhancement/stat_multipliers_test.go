package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestOrcShamanStatMultipliers checks Ancestral Knowledge and Blessing of Kings against a
// geared level 30 Orc shaman on the Forever beta (2026-10-08).
//
// The gear brought Intellect to 137. With Ancestral Knowledge 5/5 the beta showed 150
// Intellect and 2635 max mana, and with Blessing of Kings on top 165 and 2860. So the two
// multiply (137 × 1.10 × 1.10 = 165.77, adding them would give 164), and the server drops
// the fraction before it works out mana (665 + 20 + 145 × 15 = 2860).
//
// We also give the character 7 Intellect mid-fight under both. 144 × 1.21 = 174.24, so it
// should gain 9 Intellect and not the 8.47 the change makes on its own.
func TestOrcShamanStatMultipliers(t *testing.T) {
	bonus := stats.Stats{
		stats.Strength:  55,
		stats.Agility:   13,
		stats.Stamina:   55,
		stats.Intellect: 91,
		stats.Spirit:    30,
	}
	newPlayer := func(kings bool) *proto.Player {
		return &proto.Player{
			Race:          proto.Race_RaceOrc,
			Class:         proto.Class_ClassShaman,
			Level:         30,
			Equipment:     &proto.EquipmentSpec{},
			BonusStats:    &proto.UnitStats{Stats: bonus.ToFloatArray()},
			TalentsString: "-005", // Ancestral Knowledge 5/5
			Buffs:         &proto.IndividualBuffs{BlessingOfKings: kings},
			Spec: &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{
				Options: &proto.EnhancementShaman_Options{},
			}},
		}
	}
	newSim := func(kings bool) (*core.Simulation, *core.Character) {
		sim := core.NewSim(&proto.RaidSimRequest{
			Raid:       core.SinglePlayerRaidProto(newPlayer(kings), &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter:  &proto.Encounter{},
			SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
		}, simsignals.CreateSignals())
		sim.Reset()
		return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
	}
	check := func(name string, have, want float64) {
		if have != want {
			t.Errorf("%s: got %.3f, want %.3f", name, have, want)
		}
	}

	_, character := newSim(false)
	check("Intellect with Ancestral Knowledge", character.GetStat(stats.Intellect), 150)
	check("mana with Ancestral Knowledge", character.MaxMana(), 2635)

	// The other four attributes with Kings on the beta: 116 Strength (106 × 1.10 = 116.6),
	// 47 Agility (47.3), 116 Stamina (116.6) and 91 Spirit (91.3). What they convert to
	// counts the whole numbers too: 272 AP (40 + 2 × 116), 1315 health (665 + 65 × 10),
	// melee crit 1.7% + 47 × 0.0868% (the beta's 5.7396% is 0.04% lower for a 149 of 150
	// weapon skill) and 17.625 mana a second from Spirit (12.5 + 41 × 0.125).
	sim, character := newSim(true)
	check("Intellect with Kings", character.GetStat(stats.Intellect), 165)
	check("mana with Kings", character.MaxMana(), 2860)
	check("Strength with Kings", character.GetStat(stats.Strength), 116)
	check("Agility with Kings", character.GetStat(stats.Agility), 47)
	check("Stamina with Kings", character.GetStat(stats.Stamina), 116)
	check("Spirit with Kings", character.GetStat(stats.Spirit), 91)
	check("attack power with Kings", character.GetStat(stats.AttackPower), 272)
	check("health with Kings", character.MaxHealth(), 1315)
	check("mana regen with Kings", character.ManaRegenPerSecondWhileNotCasting(), 17.625)
	if crit := character.GetStat(stats.MeleeCrit) / core.CritRatingPerCritChance; math.Abs(crit-(1.7+47*0.0868)) > 0.005 {
		t.Errorf("melee crit with Kings: got %.4f, want %.4f", crit, 1.7+47*0.0868)
	}

	character.AddStatDynamic(sim, stats.Intellect, 7)
	check("Intellect after +7", character.GetStat(stats.Intellect), 174)
	check("mana after +7", character.MaxMana(), 665+20+154*15)

	character.AddStatDynamic(sim, stats.Intellect, -7)
	check("Intellect after -7", character.GetStat(stats.Intellect), 165)
	check("mana after -7", character.MaxMana(), 2860)
}
