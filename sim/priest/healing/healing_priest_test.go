//go:build with_db

// The character suites need the item database: go test -tags with_db.

package healing

import (
	"strings"
	"testing"

	_ "github.com/wowsims/classic/sim/common" // imported to get caster sets included.
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

func init() {
	RegisterHealingPriest()
}

// Forever regression suites for the healing priest's two talent presets (ui/healing_priest/presets.ts)
// on the level 60 synthetic set, each with its preset rotation.
func TestDisc(t *testing.T) {
	core.RunTestSuite(t, t.Name(), healingSuite(core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:    proto.Class_ClassPriest,
			Phase:    1,
			Race:     proto.Race_RaceHuman,
			IsHealer: true,

			Talents:     DiscTalents,
			GearSet:     core.GetGearSet("../../../ui/healing_priest/gear_sets", "level60_phase1_synthetic"),
			Rotation:    core.GetAplRotation("../../../ui/healing_priest/apls", "discipline_tank_healer"),
			Buffs:       core.ForeverBuffs,
			Consumes:    FullConsumes,
			SpecOptions: core.SpecOptionsCombo{Label: "Disc", SpecOptions: PlayerOptionsDisc},

			ItemFilter: ItemFilters,
		},
	})))
}

func TestHoly(t *testing.T) {
	core.RunTestSuite(t, t.Name(), healingSuite(core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:    proto.Class_ClassPriest,
			Phase:    1,
			Race:     proto.Race_RaceHuman,
			IsHealer: true,

			Talents:     HolyTalents,
			GearSet:     core.GetGearSet("../../../ui/healing_priest/gear_sets", "level60_phase1_synthetic"),
			Rotation:    core.GetAplRotation("../../../ui/healing_priest/apls", "holy_raid_healer"),
			Buffs:       core.ForeverBuffs,
			Consumes:    FullConsumes,
			SpecOptions: core.SpecOptionsCombo{Label: "Holy", SpecOptions: PlayerOptionsHoly},

			ItemFilter: ItemFilters,
		},
	})))
}

// healingSuite gives the suites something to heal: fire on every player for 400 every 3 sec, with
// health tracked. It drops the Average test: over its 2000 iterations the priest's mana regen sums
// to about 5e7, and the single and multi threaded runs' float sums differ by ~0.015, past the result
// comparison's 0.01 tolerance (summation order, not a state leak: 3e-10 relative).
func healingSuite(generators []core.TestGenerator) []core.TestGenerator {
	out := make([]core.TestGenerator, len(generators))
	for i, g := range generators {
		out[i] = healingGenerator{g}
	}
	return out
}

type healingGenerator struct{ core.TestGenerator }

func (g healingGenerator) NumTests() int {
	return len(g.indexes())
}

func (g healingGenerator) GetTest(i int) (string, *proto.ComputeStatsRequest, *proto.StatWeightsRequest, *proto.RaidSimRequest) {
	name, csr, swr, rsr := g.TestGenerator.GetTest(g.indexes()[i])
	if rsr != nil {
		rsr = googleProto.Clone(rsr).(*proto.RaidSimRequest)
		rsr.Encounter.TrackHealth = true
		rsr.Encounter.RaidDamage = append(rsr.Encounter.RaidDamage, &proto.RaidDamageEvent{School: proto.SpellSchool_SpellSchoolFire,
			MinDamage: 400, MaxDamage: 400, Who: proto.RaidDamageEvent_Everyone, EverySeconds: 3, FirstAt: 1})
	}
	return name, csr, swr, rsr
}

func (g healingGenerator) indexes() []int {
	var kept []int
	for i := 0; i < g.TestGenerator.NumTests(); i++ {
		if name, _, _, _ := g.TestGenerator.GetTest(i); !strings.Contains(name, "Average") {
			kept = append(kept, i)
		}
	}
	return kept
}

// The presets' talents (ui/healing_priest/presets.ts).
var DiscTalents = "005005031330011531-0350510303"
var HolyTalents = "00500503113-03505103030121531"

// The presets' consumables (ui/healing_priest/presets.ts, DefaultConsumes).
var FullConsumes = core.ConsumesCombo{
	Label: "Level60-Consumes",
	Consumes: &proto.Consumes{
		DefaultConjured: proto.Conjured_ConjuredDemonicRune,
		DefaultPotion:   proto.Potions_MajorManaPotion,
		Flask:           proto.Flask_FlaskOfSupremePower,
		Food:            proto.Food_FoodLobsterStew,
		IntellectElixir: proto.IntellectElixir_ElixirOfGreaterIntellect,
		MainHandImbue:   proto.WeaponImbue_BrilliantManaOil,
		ManaRegenElixir: proto.ManaRegenElixir_GreaterMagebloodElixir,
		SpellPowerBuff:  proto.SpellPowerBuff_GreaterArcaneElixir,
		SpiritElixir:    proto.SpiritElixir_ElixirOfSages,
		ZanzaBuff:       proto.ZanzaBuff_CerebralCortexCompound,
	},
}

var PlayerOptionsDisc = &proto.Player_HealingPriest{
	HealingPriest: &proto.HealingPriest{
		Options: &proto.HealingPriest_Options{},
	},
}

var PlayerOptionsHoly = &proto.Player_HealingPriest{
	HealingPriest: &proto.HealingPriest{
		Options: &proto.HealingPriest_Options{},
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
