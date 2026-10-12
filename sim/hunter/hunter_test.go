package hunter

import (
	"testing"

	_ "github.com/wowsims/classic/sim/common" // imported to get item effects included.
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

func init() {
	RegisterHunter()
}

func TestBeastMastery(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		hunterSuite(BeastMasteryTalents, "level60_phase1_synthetic_ranged", "bm"),
	}))
}

func TestMarksmanship(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		hunterSuite(MarksmanshipTalents, "level60_phase1_synthetic_ranged", "mm"),
	}))
}

// Survival melees under Forever. The test suite puts every player 30 yards out, so the rotation
// walks in first.
func TestSurvival(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		hunterSuite(SurvivalTalents, "level60_phase1_synthetic_survival", "sv"),
	}))
}

func hunterSuite(talents string, gear string, apl string) core.CharacterSuiteConfig {
	return core.CharacterSuiteConfig{
		Class:      proto.Class_ClassHunter,
		Phase:      1,
		Race:       proto.Race_RaceOrc,
		OtherRaces: []proto.Race{proto.Race_RaceNightElf},

		Talents:     talents,
		GearSet:     core.GetGearSet("../../ui/hunter/gear_sets", gear),
		Rotation:    core.GetAplRotation("../../ui/hunter/apls", apl),
		Buffs:       core.ForeverBuffs,
		Consumes:    Phase1Consumes,
		SpecOptions: core.SpecOptionsCombo{Label: "Cat", SpecOptions: PlayerOptionsCat},

		ItemFilter:      ItemFilters,
		EPReferenceStat: proto.Stat_StatRangedAttackPower,
		StatsToWeigh:    Stats,

		Ruleset: proto.Ruleset_RulesetForever,
	}
}

var BeastMasteryTalents = "5020021505121251-0050551"
var MarksmanshipTalents = "502002-0053552501523051"
var SurvivalTalents = "2-0050552-500230031050222151"

var Phase1Consumes = core.ConsumesCombo{
	Label: "P1-Consumes",
	Consumes: &proto.Consumes{
		AgilityElixir:   proto.AgilityElixir_ElixirOfTheMongoose,
		AttackPowerBuff: proto.AttackPowerBuff_JujuMight,
		DefaultPotion:   proto.Potions_MajorManaPotion,
		Food:            proto.Food_FoodGrilledSquid,
		ManaRegenElixir: proto.ManaRegenElixir_MagebloodPotion,
		ZanzaBuff:       proto.ZanzaBuff_GroundScorpokAssay,
	},
}

var PlayerOptionsCat = &proto.Player_Hunter{
	Hunter: &proto.Hunter{
		Options: &proto.Hunter_Options{
			Ammo:           proto.Hunter_Options_ThoriumHeadedArrow,
			QuiverBonus:    proto.Hunter_Options_Speed15,
			PetType:        proto.Hunter_Options_Cat,
			PetUptime:      1,
			PetAttackSpeed: proto.Hunter_Options_One,
		},
	},
}

var ItemFilters = core.ItemFilter{
	ArmorType: proto.ArmorType_ArmorTypeMail,
	WeaponTypes: []proto.WeaponType{
		proto.WeaponType_WeaponTypeAxe,
		proto.WeaponType_WeaponTypeDagger,
		proto.WeaponType_WeaponTypeFist,
		proto.WeaponType_WeaponTypeMace,
		proto.WeaponType_WeaponTypeOffHand,
		proto.WeaponType_WeaponTypePolearm,
		proto.WeaponType_WeaponTypeStaff,
		proto.WeaponType_WeaponTypeSword,
	},
	RangedWeaponTypes: []proto.RangedWeaponType{
		proto.RangedWeaponType_RangedWeaponTypeBow,
		proto.RangedWeaponType_RangedWeaponTypeCrossbow,
		proto.RangedWeaponType_RangedWeaponTypeGun,
	},
}

var Stats = []proto.Stat{
	proto.Stat_StatAgility,
	proto.Stat_StatAttackPower,
	proto.Stat_StatRangedAttackPower,
	proto.Stat_StatMeleeCrit,
	proto.Stat_StatMeleeHit,
}
