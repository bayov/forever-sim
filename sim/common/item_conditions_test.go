package common_test

import (
	"slices"
	"testing"

	"github.com/wowsims/classic/sim/common"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// The items whose effect grows under a condition declare it, so the BiS lists can rank them with
// the condition met. A new conditional effect that forgets to declare its condition never gets
// that second ranking, so we check the known ones here.
func TestItemConditions(t *testing.T) {
	cases := []struct {
		id    int32
		label string
	}{
		{common.BlackhandsBreadth, "vs Orcs and Ogres"},
		{common.BoltsChargedEyeball, "vs Dinosaurs"},
		{common.ChainedEssenceOfEranikus, "vs Trolls"},
		{common.GoldenBanana, "in forest or grassland"},
		{common.FrozenHeartOfTheMountain, "in snow"},
		{common.ArgentAvenger, "vs Undead"},
		{common.TreantsBane, "vs Plants"},
		{common.SmokeysLighter, "vs Plants"},
		{common.DraconicInfusedEmblem, "vs Dragonkin"},
		{common.RuneOfTheGuardCaptain, "in forest or grassland"},
		{common.JewelOfTheGuardCaptain, "in forest or grassland"},
		{common.BriarwoodReed, "in marsh or swamp"},
		{common.SecondWind, "in mountains"},
		{common.HandOfJustice, "vs Dwarves"},
		// The attack power and spell damage helpers declare their creature types.
		{common.SealOfTheDawn, "vs Undead"},
		{common.TigulesHarpoon, "vs Beasts"},
		{common.GrileksCarver, "vs Dragonkin"},
	}
	for _, c := range cases {
		labels := []string{}
		for _, condition := range core.ItemConditions(c.id) {
			labels = append(labels, condition.Label)
		}
		if !slices.Contains(labels, c.label) {
			t.Errorf("item %d has conditions %q, want %q", c.id, labels, c.label)
		}
	}
}

// A condition goes on every target of the encounter, and a met condition reads as met.
func TestItemConditionApply(t *testing.T) {
	encounter := &proto.Encounter{Targets: []*proto.Target{{}, {}}}
	for _, c := range []core.ItemCondition{core.ConditionVsOrcsAndOgres, core.ConditionInSnow, core.MobTypeCondition(proto.MobType_MobTypeUndead)} {
		if c.MetBy(encounter) {
			t.Errorf("%s: met before Apply", c.Label)
		}
		met := c.Apply(encounter)
		if !c.MetBy(met) {
			t.Errorf("%s: not met after Apply: %v", c.Label, met)
		}
	}
	if encounter.Targets[0].MobFamily != proto.MobFamily_MobFamilyUnknown {
		t.Errorf("Apply changed the encounter it was given")
	}
}
