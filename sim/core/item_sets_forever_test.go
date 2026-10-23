package core

import (
	"slices"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// A Forever piece carries only the client's set ID (its set name comes from Wowhead's Classic
// tooltips), and a reworked set grants its Forever bonuses, at their own piece counts, only under the
// Forever ruleset.
func TestForeverSetByIDAndRuleset(t *testing.T) {
	set := &ItemSet{ID: 99999, Name: "Test Set", Bonuses: RulesetSetBonuses(
		map[int32]ApplyEffect{2: SetBonusStats(stats.Stats{stats.Armor: 200})},
		map[int32]ApplyEffect{2: SetBonusStats(stats.Stats{stats.ArcaneResistance: 8}), 3: SetBonusStats(stats.Stats{stats.AttackPower: 30})},
	)}
	sets = append(sets, set)
	defer func() { sets = sets[:len(sets)-1] }()

	for _, tc := range []struct {
		ruleset proto.Ruleset
		want    stats.Stats
	}{
		{proto.Ruleset_RulesetForever, stats.Stats{stats.ArcaneResistance: 8, stats.AttackPower: 30}},
		{proto.Ruleset_RulesetClassic, stats.Stats{stats.Armor: 200}},
	} {
		agent := &FakeAgent{Character: Character{Unit: Unit{Env: &Environment{Ruleset: tc.ruleset}}}}
		character := agent.GetCharacter()
		for i, slot := range []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotLegs} {
			character.Equipment[slot] = Item{ID: int32(990000 + i), SetID: set.ID}
		}

		if names := character.GetActiveSetBonusNames(); !slices.Equal(names, []string{"Test Set (2pc)", "Test Set (3pc)"}) {
			t.Fatalf("%v: active set bonuses %v", tc.ruleset, names)
		}
		character.applyItemSetBonusEffects(agent)
		for _, stat := range []stats.Stat{stats.Armor, stats.ArcaneResistance, stats.AttackPower} {
			if got := character.GetStat(stat); got != tc.want[stat] {
				t.Errorf("%v: %s is %.0f, want %.0f", tc.ruleset, stat.StatName(), got, tc.want[stat])
			}
		}
	}
}
