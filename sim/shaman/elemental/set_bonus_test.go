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

// TestSpiritcallersStormBonuses checks The Spiritcaller's Storm on the synthetic Phase 1
// elemental set: the hands and legs are its only pieces, so it gives the 2 piece bonus (1% melee
// and spell hit) and never reaches the 5 piece one (Lava Burst 1 sec sooner).
func TestSpiritcallersStormBonuses(t *testing.T) {
	// The hands and legs have 10 hit rating each, the back none and no set.
	back, hands, legs := int32(990104), int32(990107), int32(990109)
	equip := func(ids ...int32) []*proto.ItemSpec {
		items := make([]*proto.ItemSpec, 17)
		slots := map[int32]int{back: 3, hands: 6, legs: 8}
		for i := range items {
			items[i] = &proto.ItemSpec{}
		}
		for _, id := range ids {
			items[slots[id]] = &proto.ItemSpec{Id: id}
		}
		return items
	}

	for _, tc := range []struct {
		name     string
		items    []*proto.ItemSpec
		hit      float64
		cooldown time.Duration
	}{
		{"back", equip(back), 0, 10 * time.Second},
		{"hands", equip(hands), 1, 10 * time.Second},
		{"hands and back", equip(hands, back), 1, 10 * time.Second},
		{"hands and legs", equip(hands, legs), 3, 10 * time.Second},
	} {
		ele := newSyntheticSetSim(tc.items)
		if got := ele.GetStat(stats.SpellHit) / core.SpellHitRatingPerHitChance; got != tc.hit {
			t.Errorf("%s: spell hit %.2f%%, want %.2f%%", tc.name, got, tc.hit)
		}
		if got := ele.GetStat(stats.MeleeHit) / core.MeleeHitRatingPerHitChance; got != tc.hit {
			t.Errorf("%s: melee hit %.2f%%, want %.2f%%", tc.name, got, tc.hit)
		}
		if got := ele.LavaBurst.CD.Duration; got != tc.cooldown {
			t.Errorf("%s: Lava Burst cooldown %v, want %v", tc.name, got, tc.cooldown)
		}
	}
}
