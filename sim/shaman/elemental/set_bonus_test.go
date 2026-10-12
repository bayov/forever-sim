package elemental

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

// newSyntheticSetSim equips the given synthetic Phase 1 elemental pieces on a level 60 Orc
// with Lava Burst, and starts a Forever sim.
func newSyntheticSetSim(items []*proto.ItemSpec) *ElementalShaman {
	player := &proto.Player{
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassShaman,
		Level:         60,
		Equipment:     &proto.EquipmentSpec{Items: items},
		TalentsString: "5505301300103051",
		Rotation:      &proto.APLRotation{},
		Spec: &proto.Player_ElementalShaman{ElementalShaman: &proto.ElementalShaman{
			Options: &proto.ElementalShaman_Options{},
		}},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim.Raid.Parties[0].Players[0].(*ElementalShaman)
}

// TestSpiritcallersStormBonuses checks The Spiritcaller's Storm bonuses on the synthetic
// Phase 1 elemental set: 1% hit at 2 pieces and Lava Burst's cooldown 1 sec shorter at 5.
func TestSpiritcallersStormBonuses(t *testing.T) {
	// Head, shoulders, chest and feet carry no hit of their own. Hands and legs have 10 hit
	// rating each.
	head, shoulder, chest, hands, legs, feet := int32(990101), int32(990103), int32(990105), int32(990107), int32(990109), int32(990110)
	equip := func(ids ...int32) []*proto.ItemSpec {
		items := make([]*proto.ItemSpec, 17)
		slots := map[int32]int{head: 0, shoulder: 2, chest: 4, hands: 6, legs: 8, feet: 9}
		for i := range items {
			items[i] = &proto.ItemSpec{}
		}
		for _, id := range ids {
			items[slots[id]] = &proto.ItemSpec{Id: id}
		}
		return items
	}

	for _, tc := range []struct {
		name      string
		items     []*proto.ItemSpec
		spellHit  float64
		meleeHit  float64
		lavaBurst time.Duration
	}{
		{"1 piece", equip(head), 0, 0, 10 * time.Second},
		{"2 pieces", equip(head, shoulder), 1, 1, 10 * time.Second},
		{"4 pieces", equip(head, shoulder, chest, feet), 1, 1, 10 * time.Second},
		{"5 pieces", equip(head, shoulder, chest, feet, hands), 2, 2, 9 * time.Second},
		{"6 pieces", equip(head, shoulder, chest, feet, hands, legs), 3, 3, 9 * time.Second},
	} {
		ele := newSyntheticSetSim(tc.items)
		if got := ele.GetStat(stats.SpellHit) / core.SpellHitRatingPerHitChance; got != tc.spellHit {
			t.Errorf("%s: spell hit %.2f%%, want %.2f%%", tc.name, got, tc.spellHit)
		}
		if got := ele.GetStat(stats.MeleeHit) / core.MeleeHitRatingPerHitChance; got != tc.meleeHit {
			t.Errorf("%s: melee hit %.2f%%, want %.2f%%", tc.name, got, tc.meleeHit)
		}
		if got := ele.LavaBurst.CD.Duration; got != tc.lavaBurst {
			t.Errorf("%s: Lava Burst cooldown %v, want %v", tc.name, got, tc.lavaBurst)
		}
	}
}
