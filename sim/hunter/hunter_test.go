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
		hunterSuite(BeastMasteryTalents, "level60_phase1_synthetic_ranged", "bm", RangedConsumes, PlayerOptionsCat),
	}))
}

// Beast Mastery with Summon Hawk for a point of Spirit Bond, which no Level 60 preset takes: the
// rotation sends a hawk on the cooldown it shares with Arcane Shot while fewer than 2 are out. It
// runs the turret rotation, which stands at range, so the goldens keep one that doesn't weave.
func TestBeastMasteryHawk(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		hunterSuite(BeastMasteryHawkTalents, "level60_phase1_synthetic_ranged", "bm_turret", RangedConsumes, PlayerOptionsCat),
	}))
}

func TestMarksmanship(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		hunterSuite(MarksmanshipTalents, "level60_phase1_synthetic_ranged", "mm", RangedConsumes, PlayerOptionsCat),
	}))
}

// Lone Wolf is Marksmanship without a pet.
func TestLoneWolf(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		hunterSuite(LoneWolfTalents, "level60_phase1_synthetic_ranged", "lw", RangedConsumes, PlayerOptionsNoPet),
	}))
}

// The test suite puts every player 5 yards from the target. The ranged rotations walk to 12 yards
// before the pull and melee weave from there (Marksmanship and Lone Wolf while the mana lasts; the
// turret rotation steps back to 25 yards at the pull), Survival melees.
func TestSurvival(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		hunterSuite(SurvivalTalents, "level60_phase1_synthetic_survival", "sv", SurvivalConsumes, PlayerOptionsCat),
	}))
}

func hunterSuite(talents string, gear string, apl string, consumes core.ConsumesCombo, options *proto.Player_Hunter) core.CharacterSuiteConfig {
	return core.CharacterSuiteConfig{
		Class:      proto.Class_ClassHunter,
		Phase:      1,
		Race:       proto.Race_RaceOrc,
		OtherRaces: []proto.Race{proto.Race_RaceNightElf},

		Talents:     talents,
		GearSet:     core.GetGearSet("../../ui/hunter/gear_sets", gear),
		Rotation:    core.GetAplRotation("../../ui/hunter/apls", apl),
		Buffs:       core.ForeverBuffs,
		Consumes:    consumes,
		SpecOptions: core.SpecOptionsCombo{Label: "Options", SpecOptions: options},

		ItemFilter:      ItemFilters,
		EPReferenceStat: proto.Stat_StatRangedAttackPower,
		StatsToWeigh:    Stats,

		Ruleset: proto.Ruleset_RulesetForever,
	}
}

// The level 60 builds in ui/hunter/presets.ts.
var BeastMasteryTalents = "5020121505021241-00520503005"
var BeastMasteryHawkTalents = "5020121505111241-00520503005"
var MarksmanshipTalents = "502-0053551501503051-5"
var LoneWolfTalents = "5-0053552511503051-5"
var SurvivalTalents = "502-005005-500230230250122151"

// The level 60 consumes from ui/hunter/presets.ts.
var RangedConsumes = core.ConsumesCombo{
	Label: "Ranged",
	Consumes: &proto.Consumes{
		AgilityElixir:   proto.AgilityElixir_ElixirOfTheMongoose,
		AttackPowerBuff: proto.AttackPowerBuff_JujuMight,
		DefaultConjured: proto.Conjured_ConjuredDemonicRune,
		DefaultPotion:   proto.Potions_MajorManaPotion,
		FillerExplosive: proto.Explosive_ExplosiveGoblinLandMine,
		Flask:           proto.Flask_FlaskOfDistilledWisdom,
		Food:            proto.Food_FoodFlankAuPoivre,
		IntellectElixir: proto.IntellectElixir_ElixirOfGreaterIntellect,
		ManaRegenElixir: proto.ManaRegenElixir_GreaterMagebloodElixir,
		SapperExplosive: proto.SapperExplosive_SapperGoblinSapper,
		StrengthBuff:    proto.StrengthBuff_ElixirOfFerocity,
		ZanzaBuff:       proto.ZanzaBuff_GroundScorpokAssay,
	},
}

var SurvivalConsumes = core.ConsumesCombo{
	Label: "Survival",
	Consumes: &proto.Consumes{
		AgilityElixir:     proto.AgilityElixir_ElixirOfTheMongoose,
		AttackPowerBuff:   proto.AttackPowerBuff_JujuMight,
		DragonBreathChili: true,
		MainHandImbue:     proto.WeaponImbue_DenseSharpeningStone,
		OffHandImbue:      proto.WeaponImbue_DenseSharpeningStone,
		DefaultPotion:     proto.Potions_MajorFrenzyPotion,
		FillerExplosive:   proto.Explosive_ExplosiveGoblinLandMine,
		Food:              proto.Food_FoodFlankAuPoivre,
		IntellectElixir:   proto.IntellectElixir_ElixirOfGreaterIntellect,
		SapperExplosive:   proto.SapperExplosive_SapperGoblinSapper,
		StrengthBuff:      proto.StrengthBuff_ElixirOfFerocity,
		ZanzaBuff:         proto.ZanzaBuff_GroundScorpokAssay,
	},
}

var PlayerOptionsCat = &proto.Player_Hunter{
	Hunter: &proto.Hunter{
		Options: &proto.Hunter_Options{
			Ammo:           proto.Hunter_Options_SwiftfeatherArrow,
			QuiverBonus:    proto.Hunter_Options_Speed15,
			PetType:        proto.Hunter_Options_Cat,
			PetUptime:      1,
			PetAttackSpeed: proto.Hunter_Options_One,
			CobraReflexes:  true,
		},
	},
}

var PlayerOptionsNoPet = &proto.Player_Hunter{
	Hunter: &proto.Hunter{
		Options: &proto.Hunter_Options{
			Ammo:        proto.Hunter_Options_SwiftfeatherArrow,
			QuiverBonus: proto.Hunter_Options_Speed15,
			PetType:     proto.Hunter_Options_PetNone,
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
