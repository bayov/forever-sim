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
// 2026-10-07. The beta stops at level 30, so level 60 pins the 1.12 values until we can
// check them on release.
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
		check("melee crit", got[stats.MeleeCrit]/core.CritRatingPerCritChance, want.meleeCrit, 0.005)
		check("spell crit", got[stats.SpellCrit]/core.CritRatingPerCritChance, want.spellCrit, 0.005)
		check("dodge", got[stats.Dodge]/core.DodgeRatingPerDodgeChance, want.dodge, 0.005)
	}
}
