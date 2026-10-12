package core

import (
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

func foreverTestCharacter() *Character {
	return &Character{Unit: Unit{Env: &Environment{Ruleset: proto.Ruleset_RulesetForever}, Level: 60}}
}

func expectStats(t *testing.T, name string, got stats.Stats, want stats.Stats) {
	t.Helper()
	for s := range got {
		if got[s] != want[s] {
			t.Errorf("%s: %s is %.2f, want %.2f", name, stats.Stat(s).StatName(), got[s], want[s])
		}
	}
}

// The values from client 1.60.1.70338 (SpellEffect of each item's spell).
func TestForeverConsumeStats(t *testing.T) {
	crit := stats.Stats{stats.MeleeCrit: 2 * CritRatingPerCritChance, stats.SpellCrit: 2 * SpellCritRatingPerCritChance}
	cases := []struct {
		name     string
		consumes *proto.Consumes
		want     stats.Stats
	}{
		{"Elixir of the Owl", &proto.Consumes{IntellectElixir: proto.IntellectElixir_ElixirOfTheOwl}, crit.Add(stats.Stats{stats.Intellect: 25})},
		{"Elixir of Greater Intellect", &proto.Consumes{IntellectElixir: proto.IntellectElixir_ElixirOfGreaterIntellect}, stats.Stats{stats.Intellect: 25}},
		{"Elixir of Intellect", &proto.Consumes{IntellectElixir: proto.IntellectElixir_ElixirOfIntellect}, stats.Stats{stats.Intellect: 10}},
		{"Elixir of Sages", &proto.Consumes{SpiritElixir: proto.SpiritElixir_ElixirOfSages}, crit.Add(stats.Stats{stats.Spirit: 25})},
		{"Owl and Sages", &proto.Consumes{IntellectElixir: proto.IntellectElixir_ElixirOfTheOwl, SpiritElixir: proto.SpiritElixir_ElixirOfSages}, crit.Add(crit).Add(stats.Stats{stats.Intellect: 25, stats.Spirit: 25})},
		{"Elixir of the Grizzly", &proto.Consumes{StrengthBuff: proto.StrengthBuff_ElixirOfTheGrizzly}, crit.Add(stats.Stats{stats.Strength: 25})},
		{"Elixir of Ferocity", &proto.Consumes{StrengthBuff: proto.StrengthBuff_ElixirOfFerocity}, stats.Stats{stats.Strength: 18, stats.Agility: 18}},
		{"Elixir of Cunning", &proto.Consumes{AgilityElixir: proto.AgilityElixir_ElixirOfCunning}, stats.Stats{stats.Agility: 25, stats.Intellect: 25}},
		{"Elixir of Nature Power", &proto.Consumes{NaturePowerBuff: proto.NaturePowerBuff_ElixirOfNaturePower}, stats.Stats{stats.NaturePower: 40}},
		{"Greater Mageblood Elixir", &proto.Consumes{ManaRegenElixir: proto.ManaRegenElixir_GreaterMagebloodElixir}, stats.Stats{stats.MP5: 20}},
		{"Greater Cleric's Elixir", &proto.Consumes{HealingPowerBuff: proto.HealingPowerBuff_GreaterClericsElixir}, stats.Stats{stats.HealingPower: 40}},
		{"Lobster Stew", &proto.Consumes{Food: proto.Food_FoodLobsterStew}, stats.Stats{stats.SpellDamage: 28}},
		{"Mightfish Steak", &proto.Consumes{Food: proto.Food_FoodMightfishSteak}, stats.Stats{stats.AttackPower: 40, stats.RangedAttackPower: 40}},
		{"Flank au Poivre", &proto.Consumes{Food: proto.Food_FoodFlankAuPoivre}, stats.Stats{stats.Agility: 20}},
		{"Savory Stag Sliders", &proto.Consumes{Food: proto.Food_FoodSavoryStagSliders}, stats.Stats{stats.Strength: 15}},
		{"Prehistoric Pulled Raptor", &proto.Consumes{Food: proto.Food_FoodPrehistoricPulledRaptor}, stats.Stats{stats.Intellect: 20}},
		{"Wicked Smoothie", &proto.Consumes{Food: proto.Food_FoodWickedSmoothie}, stats.Stats{stats.Spirit: 20}},
	}
	for _, c := range cases {
		character := foreverTestCharacter()
		applyFoodConsumes(character, c.consumes)
		applyPhysicalBuffConsumes(character, c.consumes)
		applySpellBuffConsumes(character, c.consumes, &proto.RaidBuffs{})
		expectStats(t, c.name, character.GetStats(), c.want)
	}
}

// A stat elixir adds only what it gives above the Scroll of the same stat, but its crit counts in full.
func TestForeverElixirOverScroll(t *testing.T) {
	character := foreverTestCharacter()
	consumes := &proto.Consumes{IntellectElixir: proto.IntellectElixir_ElixirOfTheOwl, SpiritElixir: proto.SpiritElixir_ElixirOfSages}
	applySpellBuffConsumes(character, consumes, &proto.RaidBuffs{ScrollOfIntellect: true, ScrollOfSpirit: true})
	want := stats.Stats{
		stats.Intellect: 25 - 16,
		stats.Spirit:    25 - 15,
		stats.MeleeCrit: 4 * CritRatingPerCritChance,
		stats.SpellCrit: 4 * SpellCritRatingPerCritChance,
	}
	expectStats(t, "Owl and Sages over scrolls", character.GetStats(), want)
}

func TestForeverStatPotions(t *testing.T) {
	cases := []struct {
		potion proto.Potions
		itemID int32
		stat   stats.Stat
		amount float64
	}{
		{proto.Potions_MajorSpellblastingPotion, 250937, stats.SpellDamage, 47},
		{proto.Potions_MajorFrenzyPotion, 250943, stats.AttackPower, 80},
		{proto.Potions_MajorFrenzyPotion, 250943, stats.RangedAttackPower, 80},
	}
	for _, c := range cases {
		sim := NewSim(&proto.RaidSimRequest{
			SimOptions: &proto.SimOptions{RandomSeed: 100, Ruleset: proto.Ruleset_RulesetForever},
			Raid: &proto.Raid{Parties: []*proto.Party{{
				Players: []*proto.Player{{
					Name:      "Caster",
					Class:     proto.Class_ClassShaman,
					Level:     60,
					Consumes:  &proto.Consumes{DefaultPotion: c.potion},
					Buffs:     &proto.IndividualBuffs{},
					Spec:      &proto.Player_ElementalShaman{},
					Equipment: &proto.EquipmentSpec{},
				}},
				Buffs: &proto.PartyBuffs{},
			}}},
			Encounter: &proto.Encounter{
				Targets:  []*proto.Target{{Name: "target", Level: 63}},
				Duration: 180,
			},
		}, simsignals.CreateSignals())
		sim.Reset()

		character := sim.Raid.Parties[0].Players[0].GetCharacter()
		mcd := character.GetMajorCooldown(ActionID{ItemID: c.itemID})
		if mcd == nil {
			t.Fatalf("%s: no potion cooldown", c.potion)
		}
		before := character.GetStat(c.stat)
		mcd.Spell.Cast(sim, nil)
		if got := character.GetStat(c.stat) - before; got != c.amount {
			t.Errorf("%s: %s went up %.2f, want %.2f", c.potion, c.stat.StatName(), got, c.amount)
		}
	}
}
