package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestOrcShamanBaseStats pins a naked Orc shaman with no talents to the Forever beta.
//
// The level 1 and 30 numbers come from the beta's Lua API (UnitStat, UnitHealthMax,
// UnitPowerMax, UnitAttackPower, UnitArmor, GetSpellCritChance, GetDodgeChance) on
// 2026-10-07. The beta stops at level 30, so level 60 pins the 1.12 values with Forever's
// Orc shift (1 less Stamina, 2 more Intellect, 3 less Spirit, seen on a level 1 Orc shaman
// and warlock on 2026-10-10) until we can check them on release.
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
		{1, 24, 17, 22, 20, 22, 67, 75, 30, 34, 4.504, 4.875, 4.504},
		{30, 51, 30, 51, 46, 53, 665, 1075, 142, 60, 4.304, 3.933, 4.304},
		{60, 88, 52, 96, 89, 100, 2060, 2575, 276, 104, 4.342, 3.804, 4.342},
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
		check("melee crit", got[stats.MeleeCrit]/core.CritRatingPerCritChance, want.meleeCrit, 0.005)
		check("spell crit", got[stats.SpellCrit]/core.CritRatingPerCritChance, want.spellCrit, 0.005)
		check("dodge", got[stats.Dodge]/core.DodgeRatingPerDodgeChance, want.dodge, 0.005)
	}
}

// TestOrcShamanAttributeRates checks what each attribute gives well above the naked values,
// against a geared level 30 Orc shaman on the Forever beta (2026-10-08). We give a naked
// Orc the same attribute bonuses as that gear (+55 Strength, +13 Agility, +55 Stamina, +91
// Intellect, +30 Spirit) and nothing else.
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
	}
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
	check("melee crit", got[stats.MeleeCrit]/core.CritRatingPerCritChance, 5.43, 0.005)
	check("dodge", got[stats.Dodge]/core.DodgeRatingPerDodgeChance, 5.43, 0.005)
	check("spell crit", got[stats.SpellCrit]/core.CritRatingPerCritChance, 7.16, 0.01)
	// 83 per 5 sec on the beta's tooltip, 12.5 + 33 / 8 = 16.625 a second.
	check("mana regen", character.ManaRegenPerSecondWhileNotCasting(), 16.625, 1e-9)
}
