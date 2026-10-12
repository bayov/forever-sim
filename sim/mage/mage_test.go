package mage

import (
	"testing"

	_ "github.com/wowsims/classic/sim/common"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

func init() {
	RegisterMage()
}

func TestP1Mage(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class: proto.Class_ClassMage,
			Phase: 1,
			Race:  proto.Race_RaceTroll,
			// Not converted to Forever yet, so it keeps running under Classic rules.
			Ruleset:    proto.Ruleset_RulesetClassic,
			OtherRaces: []proto.Race{proto.Race_RaceUndead},

			Talents:     P1Talents,
			GearSet:     core.GetGearSet("../../ui/mage/gear_sets", "level60_phase1_synthetic"),
			Rotation:    core.GetAplRotation("../../ui/mage/apls", "level60_frostfire"),
			Buffs:       core.FullBuffs,
			Consumes:    P1Consumes,
			SpecOptions: core.SpecOptionsCombo{Label: "DPS", SpecOptions: PlayerOptions},

			ItemFilter:      ItemFilters,
			EPReferenceStat: proto.Stat_StatSpellPower,
			StatsToWeigh:    Stats,
		},
	}))
}

// The Level 60 Frostfire build (ui/mage/presets.ts).
var P1Talents = "-1355000013013304-00550003310003002"

var PlayerOptions = &proto.Player_Mage{
	Mage: &proto.Mage{
		Options: &proto.Mage_Options{
			Armor: proto.Mage_Options_MageArmor,
		},
	},
}

var P1Consumes = core.ConsumesCombo{
	Label: "P1-Consumes",
	Consumes: &proto.Consumes{
		DefaultPotion:   proto.Potions_MajorManaPotion,
		FirePowerBuff:   proto.FirePowerBuff_ElixirOfFirepower,
		Flask:           proto.Flask_FlaskOfSupremePower,
		Food:            proto.Food_FoodNightfinSoup,
		FrostPowerBuff:  proto.FrostPowerBuff_ElixirOfFrostPower,
		IntellectElixir: proto.IntellectElixir_ElixirOfWisdom,
		MainHandImbue:   proto.WeaponImbue_BrilliantWizardOil,
		ManaRegenElixir: proto.ManaRegenElixir_MagebloodPotion,
		SpellPowerBuff:  proto.SpellPowerBuff_GreaterArcaneElixir,
	},
}

var ItemFilters = core.ItemFilter{
	WeaponTypes: []proto.WeaponType{
		proto.WeaponType_WeaponTypeDagger,
		proto.WeaponType_WeaponTypeSword,
		proto.WeaponType_WeaponTypeOffHand,
		proto.WeaponType_WeaponTypeStaff,
	},
	ArmorType: proto.ArmorType_ArmorTypeCloth,
	RangedWeaponTypes: []proto.RangedWeaponType{
		proto.RangedWeaponType_RangedWeaponTypeWand,
	},
}

var Stats = []proto.Stat{
	proto.Stat_StatIntellect,
	proto.Stat_StatSpellPower,
	proto.Stat_StatArcanePower,
	proto.Stat_StatFirePower,
	proto.Stat_StatFrostPower,
	proto.Stat_StatSpellHit,
	proto.Stat_StatSpellCrit,
}
