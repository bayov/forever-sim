//go:build with_db

package hunter

import (
	"testing"

	"github.com/wowsims/classic/sim/common"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

// Marksman Bands, Stoneshatter and Sash of the Grand Hunt each give 1% hit with ranged weapons
// in Forever. It goes on the hunter's ranged shots only: not on melee, not on the hunter's hit
// stat, and not on the pet.
func TestForeverRangedHitItems(t *testing.T) {
	newSim := func(equip bool) *Hunter {
		items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
		for i := range items {
			items[i] = &proto.ItemSpec{}
		}
		if equip {
			items[proto.ItemSlot_ItemSlotWrist].Id = common.MarksmanBands
			items[proto.ItemSlot_ItemSlotWaist].Id = common.SashOfTheGrandHunt
			items[proto.ItemSlot_ItemSlotRanged].Id = common.Stoneshatter
		}
		player := &proto.Player{
			Class:     proto.Class_ClassHunter,
			Race:      proto.Race_RaceHuman,
			Level:     60,
			Equipment: &proto.EquipmentSpec{Items: items},
			Spec: &proto.Player_Hunter{Hunter: &proto.Hunter{Options: &proto.Hunter_Options{
				PetType: proto.Hunter_Options_Cat, PetUptime: 1, PetAttackSpeed: proto.Hunter_Options_Two,
			}}},
		}
		sim := core.NewSim(&proto.RaidSimRequest{
			Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
			SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: 1},
		}, simsignals.CreateSignals())
		sim.Reset()
		return sim.Raid.Parties[0].Players[0].(HunterAgent).GetHunter()
	}
	without, with := newSim(false), newSim(true)

	ranged := with.AutoAttacks.RangedAuto().BonusHitRating - without.AutoAttacks.RangedAuto().BonusHitRating
	if want := 3.0 * core.MeleeHitRatingPerHitChance; ranged != want {
		t.Errorf("the three items add %v hit to Auto Shot, want %v", ranged, want)
	}
	if got := with.AutoAttacks.MHAuto().BonusHitRating - without.AutoAttacks.MHAuto().BonusHitRating; got != 0 {
		t.Errorf("the three items add %v hit to melee swings, want 0", got)
	}
	for _, stat := range []stats.Stat{stats.MeleeHit, stats.SpellHit} {
		if got := with.GetStat(stat) - without.GetStat(stat); got != 0 {
			t.Errorf("the three items add %v to the hunter's %v, want 0", got, stat)
		}
		if got := with.pet.GetStat(stat) - without.pet.GetStat(stat); got != 0 {
			t.Errorf("the three items add %v to the pet's %v, want 0", got, stat)
		}
	}
}
