package enhancement

import (
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

// TestOrcShamanLavaBurstRanks checks that a rotation naming Lava Burst rank 3 casts the rank
// the level knows.
//
// The talent teaches rank 1 at 40, and the trainer has rank 2 at 50 and rank 3 at 60. Each
// rank has its own Forever spell ID.
func TestOrcShamanLavaBurstRanks(t *testing.T) {
	cases := []struct {
		level int32
		rank  int
	}{{45, 1}, {55, 2}, {60, 3}}
	for _, c := range cases {
		// Lava Burst.
		player := newOrcShaman(c.level, "0000000000000001", &proto.EnhancementShaman_Options{})
		player.Rotation = &proto.APLRotation{PriorityList: []*proto.APLListItem{{Action: &proto.APLAction{
			Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{
				SpellId: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: shaman.LavaBurstSpellId[3]}},
			}},
		}}}}
		sim, enh := newShamanSim(player, &proto.Debuffs{}, 15)
		runSim(sim)

		lavaBurst := enh.LavaBurst
		if got, want := lavaBurst.ActionID.SpellID, shaman.LavaBurstSpellId[c.rank]; got != want || lavaBurst.Rank != c.rank {
			t.Errorf("level %d: got Lava Burst %d rank %d, want %d rank %d", c.level, got, lavaBurst.Rank, want, c.rank)
		}
		if casts := lavaBurst.SpellMetrics[enh.CurrentTarget.UnitIndex].Casts; casts < 2 {
			t.Errorf("level %d: the rotation cast Lava Burst %d times in 15 sec", c.level, casts)
		}
	}
}
