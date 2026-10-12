package dps

import (
	"testing"

	_ "github.com/wowsims/classic/sim/common"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

func init() {
	RegisterDpsWarlock()
}

// The five Level 60 builds of ui/warlock/presets.ts on the synthetic Phase 1 sets.
func warlockSuite(t *testing.T, talents string, gear string, apl string, label string, options *proto.Player_Warlock) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class: proto.Class_ClassWarlock,
			Phase: 1,
			Race:  proto.Race_RaceHuman,

			Talents:     talents,
			GearSet:     core.GetGearSet("../../../ui/warlock/gear_sets", gear),
			Rotation:    core.GetAplRotation("../../../ui/warlock/apls/", apl),
			Buffs:       core.FullBuffs,
			Consumes:    Consumes,
			SpecOptions: core.SpecOptionsCombo{Label: label, SpecOptions: options},

			ItemFilter:      ItemFilters,
			EPReferenceStat: proto.Stat_StatSpellPower,
			StatsToWeigh:    Stats,
		},
	}))
}

func TestWarlockDemonology(t *testing.T) {
	warlockSuite(t, TalentsDemonology, "level60_phase1_synthetic_demonology", "level60_demonology", "Demonology Warlock", DemonologyWarlock)
}

func TestWarlockDestruction(t *testing.T) {
	warlockSuite(t, TalentsDestruction, "level60_phase1_synthetic", "level60_destruction", "Destruction Warlock", SuccubusWarlock)
}

func TestWarlockAffliction(t *testing.T) {
	warlockSuite(t, TalentsAffliction, "level60_phase1_synthetic", "level60_affliction", "Affliction Warlock", SuccubusWarlock)
}

func TestWarlockDestructionFire(t *testing.T) {
	warlockSuite(t, TalentsDestructionFire, "level60_phase1_synthetic", "level60_destruction_fire", "Destruction Fire Warlock", FireWarlock)
}

func TestWarlockAfflictionDoT(t *testing.T) {
	warlockSuite(t, TalentsAfflictionDoT, "level60_phase1_synthetic", "level60_affliction_dot", "Affliction DoT Warlock", SuccubusWarlock)
}

// Forever tree (tools/forever_talents/client/warlock.json).
var TalentsDemonology = "011-0305003131221001351-0550005003"
var TalentsDestruction = "25510000035--055030500310035"
var TalentsAffliction = "2553000003520005--0550305003"
var TalentsDestructionFire = "252-0050003201-0050345003101351"
var TalentsAfflictionDoT = "25530000035001351--0550005003"

// Sacrifices the Succubus before the pull for 15% Fire damage and keeps no demon out.
var FireWarlock = &proto.Player_Warlock{
	Warlock: &proto.Warlock{
		Options: &proto.WarlockOptions{
			Armor:         proto.WarlockOptions_DemonArmor,
			SacrificedPet: proto.WarlockOptions_Succubus,
			WeaponImbue:   proto.WarlockOptions_Firestone,
		},
	},
}

// Sacrifices the Imp before the pull and keeps the Succubus out, which Demonic Pact allows.
var DemonologyWarlock = &proto.Player_Warlock{
	Warlock: &proto.Warlock{
		Options: &proto.WarlockOptions{
			Armor:         proto.WarlockOptions_DemonArmor,
			Summon:        proto.WarlockOptions_Succubus,
			SacrificedPet: proto.WarlockOptions_Imp,
			WeaponImbue:   proto.WarlockOptions_Spellstone,
		},
	},
}

var SuccubusWarlock = &proto.Player_Warlock{
	Warlock: &proto.Warlock{
		Options: &proto.WarlockOptions{
			Armor:       proto.WarlockOptions_DemonArmor,
			Summon:      proto.WarlockOptions_Succubus,
			WeaponImbue: proto.WarlockOptions_Spellstone,
		},
	},
}

var Consumes = core.ConsumesCombo{
	Label: "Consumes",
	Consumes: &proto.Consumes{
		DefaultPotion:   proto.Potions_MajorManaPotion,
		DefaultConjured: proto.Conjured_ConjuredDemonicRune,
		Flask:           proto.Flask_FlaskOfSupremePower,
		FirePowerBuff:   proto.FirePowerBuff_ElixirOfFirepower,
		ShadowPowerBuff: proto.ShadowPowerBuff_ElixirOfShadowPower,
		Food:            proto.Food_FoodLobsterStew,
		IntellectElixir: proto.IntellectElixir_ElixirOfGreaterIntellect,
		MainHandImbue:   proto.WeaponImbue_WizardOil,
		ManaRegenElixir: proto.ManaRegenElixir_GreaterMagebloodElixir,
		SpellPowerBuff:  proto.SpellPowerBuff_GreaterArcaneElixir,
		SpiritElixir:    proto.SpiritElixir_ElixirOfSages,
		ZanzaBuff:       proto.ZanzaBuff_CerebralCortexCompound,
	},
}

var ItemFilters = core.ItemFilter{
	WeaponTypes: []proto.WeaponType{
		proto.WeaponType_WeaponTypeSword,
		proto.WeaponType_WeaponTypeDagger,
	},
	HandTypes: []proto.HandType{
		proto.HandType_HandTypeOffHand,
	},
	ArmorType: proto.ArmorType_ArmorTypeCloth,
	RangedWeaponTypes: []proto.RangedWeaponType{
		proto.RangedWeaponType_RangedWeaponTypeWand,
	},
}

var Stats = []proto.Stat{
	proto.Stat_StatIntellect,
	proto.Stat_StatSpellPower,
	proto.Stat_StatSpellHit,
	proto.Stat_StatSpellCrit,
}
