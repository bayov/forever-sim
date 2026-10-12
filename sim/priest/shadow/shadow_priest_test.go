package shadow

import (
	"testing"

	_ "github.com/wowsims/classic/sim/common" // imported to get caster sets included.
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

func init() {
	RegisterShadowPriest()
}

func TestP1Shadow(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassPriest,
			Phase:      1,
			Race:       proto.Race_RaceUndead,
			OtherRaces: []proto.Race{proto.Race_RaceHuman, proto.Race_RaceNightElf, proto.Race_RaceDwarf},

			Talents:     P1Talents,
			GearSet:     core.GetGearSet("../../../ui/shadow_priest/gear_sets", "level60_phase1_synthetic"),
			Rotation:    core.GetAplRotation("../../../ui/shadow_priest/apls", "level60"),
			Buffs:       core.ForeverBuffs,
			Consumes:    P1Consumes,
			SpecOptions: core.SpecOptionsCombo{Label: "Basic", SpecOptions: PlayerOptionsBasic},

			ItemFilter:      ItemFilters,
			EPReferenceStat: proto.Stat_StatSpellPower,
			StatsToWeigh:    Stats,
		},
	}))
}

var P1Talents = "005030211305--500321501201302051"

var P1Consumes = core.ConsumesCombo{
	Label: "Level60-Consumes",
	Consumes: &proto.Consumes{
		DefaultPotion:   proto.Potions_MajorManaPotion,
		DefaultConjured: proto.Conjured_ConjuredDemonicRune,
		Flask:           proto.Flask_FlaskOfSupremePower,
		Food:            proto.Food_FoodNightfinSoup,
		IntellectElixir: proto.IntellectElixir_ElixirOfWisdom,
		MainHandImbue:   proto.WeaponImbue_WizardOil,
		SapperExplosive: proto.SapperExplosive_SapperGoblinSapper,
		SpellPowerBuff:  proto.SpellPowerBuff_GreaterArcaneElixir,
		ShadowPowerBuff: proto.ShadowPowerBuff_ElixirOfShadowPower,
		ZanzaBuff:       proto.ZanzaBuff_CerebralCortexCompound,
	},
}

var PlayerOptionsBasic = &proto.Player_ShadowPriest{
	ShadowPriest: &proto.ShadowPriest{
		Options: &proto.ShadowPriest_Options{
			Armor: proto.ShadowPriest_Options_InnerFire,
		},
	},
}

var ItemFilters = core.ItemFilter{
	WeaponTypes: []proto.WeaponType{
		proto.WeaponType_WeaponTypeDagger,
		proto.WeaponType_WeaponTypeMace,
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
	proto.Stat_StatShadowPower,
	proto.Stat_StatSpellHit,
	proto.Stat_StatSpellCrit,
}
