package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestOrcShamanBaseStats pins a naked Orc shaman with no talents.
//
// The level 1 and 30 numbers came from the beta's Lua API (UnitStat, UnitHealthMax,
// UnitPowerMax, UnitAttackPower, UnitArmor, GetSpellCritChance, GetDodgeChance) on
// 2026-10-07, when the Forever Orc had 1 less Stamina, 2 more Intellect and 3 less Spirit than
// Classic's (betaOrcShift). Client 1.60.1.70753 gave every race Classic's attributes
// (development notes 2026-11-02), so the rows below are the sheets with that shift taken off:
// 10 more health, 30 less mana and the spell crit of 2 less Intellect. Level 60 is the 1.12
// row until we can check it on release.
//
// The crit readings come from client 70338, when the Orc's crit racial needed an axe. Client
// 70425's Savage gives 1% more melee and spell crit with any weapon, so the sim's crit is 1%
// above them (savageCrit).
//
// The beta's level 1 dodge reads 4.3% because the new character has 1 of 5 defense, and
// each missing point takes 0.04% off. The sim doesn't model defense skill, so we pin the
// 4.5% a trained character has.
func TestOrcShamanBaseStats(t *testing.T) {
	type expected struct {
		level                       int32
		str, agi, sta, int, spi     float64
		health, mana, ap, armor     float64
		meleeCrit, spellCrit, dodge float64
	}
	cases := []expected{
		{1, 24, 17, 23, 18, 25, 77, 45, 30, 34, 4.504, 4.617, 4.504},
		{30, 51, 30, 52, 44, 56, 675, 1045, 142, 60, 4.304, 3.864, 4.304},
		{60, 88, 52, 97, 87, 103, 2070, 2545, 276, 104, 4.342, 3.770, 4.342},
	}

	for _, want := range cases {
		player := &proto.Player{
			Race:      proto.Race_RaceOrc,
			Class:     proto.Class_ClassShaman,
			Level:     want.level,
			Equipment: &proto.EquipmentSpec{},
			Spec: &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{
				Options: &proto.EnhancementShaman_Options{},
			}},
		}
		result := core.ComputeStats(&proto.ComputeStatsRequest{
			Raid:    core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Ruleset: proto.Ruleset_RulesetForever,
		})
		got := stats.FromFloatArray(result.RaidStats.Parties[0].Players[0].BaseStats.Stats)

		check := func(name string, have, want, tolerance float64) {
			if math.Abs(have-want) > tolerance {
				t.Errorf("level %d %s: got %.3f, want %.3f", player.Level, name, have, want)
			}
		}
		check("Strength", got[stats.Strength], want.str, 0)
		check("Agility", got[stats.Agility], want.agi, 0)
		check("Stamina", got[stats.Stamina], want.sta, 0)
		check("Intellect", got[stats.Intellect], want.int, 0)
		check("Spirit", got[stats.Spirit], want.spi, 0)
		check("health", got[stats.Health], want.health, 0)
		check("mana", got[stats.Mana], want.mana, 0)
		check("attack power", got[stats.AttackPower], want.ap, 0)
		check("armor", got[stats.Armor], want.armor, 0)
		check("melee crit", got[stats.MeleeCrit]/core.CritRatingPerCritChance, want.meleeCrit+savageCrit, 0.005)
		check("spell crit", got[stats.SpellCrit]/core.CritRatingPerCritChance, want.spellCrit+savageCrit, 0.005)
		check("dodge", got[stats.Dodge]/core.DodgeRatingPerDodgeChance, want.dodge, 0.005)
	}
}

// savageCrit is the Orc's crit racial on top of the beta readings below, taken on client 70338
// when it needed an axe and our characters had none. Client 70425's Savage gave 1% with any
// weapon; the development notes of 2026-10-26 remove Savage, so it is 0 again.
const savageCrit = 0.0

// betaOrcShift is the beta's Orc of 2026-10-07 and 2026-10-08 against Classic's: 1 less Stamina,
// 2 more Intellect and 3 less Spirit (our level 1 Orc shaman and warlock sheets of 2026-10-10).
// Client 1.60.1.70753 put the races on Classic's attributes, so a test that checks a mechanic
// against a beta sheet adds it to the character's bonus stats to stand on the sheet's numbers.
var betaOrcShift = stats.Stats{stats.Stamina: -1, stats.Intellect: 2, stats.Spirit: -3}

// TestOrcShamanAttributeRates checks what each attribute gives well above the naked values,
// against a geared level 30 Orc shaman on the Forever beta (2026-10-08). We give a naked
// Orc the same attribute bonuses as that gear (+55 Strength, +13 Agility, +55 Stamina, +91
// Intellect, +30 Spirit) and the beta's Orc attributes (betaOrcShift), and nothing else.
//
// The beta showed 212 AP from Strength, 86 armor from Agility, 880 health from Stamina, 1775
// mana from Intellect, 5.43% dodge, 5.39% melee crit with a 149 of 150 weapon skill (each
// missing point takes 0.04% off, so 5.43% at full skill), 7.16% spell crit and 83 mana per 5
// sec from Spirit. Health and mana add the base 335 and 665.
//
// The beta's spell crit is exactly 0.0355% per Intellect at level 30 (GetSpellCritChance on
// the naked character). The sim scales the level 60 rate along the client's curve and gets
// 0.03554%, which puts it 0.006% above the beta at 137 Intellect, so spell crit has a looser
// tolerance than the rest.
func TestOrcShamanAttributeRates(t *testing.T) {
	bonus := stats.Stats{
		stats.Strength:  55,
		stats.Agility:   13,
		stats.Stamina:   55,
		stats.Intellect: 91,
		stats.Spirit:    30,
	}.Add(betaOrcShift)
	player := &proto.Player{
		Race:       proto.Race_RaceOrc,
		Class:      proto.Class_ClassShaman,
		Level:      30,
		Equipment:  &proto.EquipmentSpec{},
		BonusStats: &proto.UnitStats{Stats: bonus.ToFloatArray()},
		Spec: &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{
			Options: &proto.EnhancementShaman_Options{},
		}},
	}
	raid := core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
	env, _, _ := core.NewEnvironment(raid, &proto.Encounter{}, proto.Ruleset_RulesetForever, false)
	character := env.Raid.Parties[0].Players[0].GetCharacter()
	got := character.GetStats()

	check := func(name string, have, want, tolerance float64) {
		if math.Abs(have-want) > tolerance {
			t.Errorf("%s: got %.3f, want %.3f", name, have, want)
		}
	}
	check("Strength", got[stats.Strength], 106, 0)
	check("Agility", got[stats.Agility], 43, 0)
	check("Stamina", got[stats.Stamina], 106, 0)
	check("Intellect", got[stats.Intellect], 137, 0)
	check("Spirit", got[stats.Spirit], 83, 0)
	check("attack power", got[stats.AttackPower], 40+212, 0)
	check("armor", got[stats.Armor], 86, 0)
	check("health", got[stats.Health], 335+880, 0)
	check("mana", got[stats.Mana], 665+1775, 0)
	check("melee crit", got[stats.MeleeCrit]/core.CritRatingPerCritChance, 5.43+savageCrit, 0.005)
	check("dodge", got[stats.Dodge]/core.DodgeRatingPerDodgeChance, 5.43, 0.005)
	check("spell crit", got[stats.SpellCrit]/core.CritRatingPerCritChance, 7.16+savageCrit, 0.01)
	// 83 per 5 sec on the beta's tooltip, 12.5 + 33 / 8 = 16.625 a second.
	check("mana regen", character.ManaRegenPerSecondWhileNotCasting(), 16.625, 1e-9)
}

// TestOrcShamanHealthTalents checks Toughness and Improved Reincarnation on a naked level 30
// Orc shaman.
//
// Toughness gives 2% more Stamina a point and drops the fraction, so 5/5 turns 52 Stamina
// into 57. Improved Reincarnation gives 2% more max health a point, Stamina's part included.
// Both are the Forever client's values (shaman_audit.md 6.6). Base health is 335, the first
// 20 Stamina give 1 health each and the rest 10.
func TestOrcShamanHealthTalents(t *testing.T) {
	cases := []struct {
		name    string
		talents string
		stamina float64
		health  float64
	}{
		{"no talents", "", 52, 675},
		{"Toughness 5/5", "-00000000005", 57, 725},
		{"Improved Reincarnation 2/2", "--000002", 52, 675 * 1.04},
		{"both", "-00000000005-000002", 57, 725 * 1.04},
	}
	for _, c := range cases {
		_, enh := newShamanSim(newOrcShaman(30, c.talents, &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 10)
		if got := enh.GetStat(stats.Stamina); got != c.stamina {
			t.Errorf("%s: %.2f Stamina, want %.0f", c.name, got, c.stamina)
		}
		if got := enh.MaxHealth(); math.Abs(got-c.health) > 1e-6 {
			t.Errorf("%s: %.2f health, want %.2f", c.name, got, c.health)
		}
	}
}
