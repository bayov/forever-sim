package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Forever's Five Thunders (set 1679) in a mix of its three variants: Strength wrist, waist,
// shoulders and legs, the caster head and the healer feet.
var fiveThundersMix = []struct {
	id   int32
	slot proto.ItemSlot
}{
	{227017, proto.ItemSlot_ItemSlotWrist},
	{227018, proto.ItemSlot_ItemSlotWaist},
	{227019, proto.ItemSlot_ItemSlotShoulder},
	{227020, proto.ItemSlot_ItemSlotLegs},
	{227002, proto.ItemSlot_ItemSlotHead},
	{227015, proto.ItemSlot_ItemSlotFeet},
}

func fiveThundersSim(t *testing.T, pieces int) *EnhancementShaman {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	for _, piece := range fiveThundersMix[:pieces] {
		if _, ok := core.ItemsByID[piece.id]; !ok {
			t.Fatalf("no item %d in the database", piece.id)
		}
		items[piece.slot] = &proto.ItemSpec{Id: piece.id}
	}
	player := newOrcShaman(60, restoRiptidePurif, &proto.EnhancementShaman_Options{})
	player.Equipment = &proto.EquipmentSpec{Items: items}
	_, sh := newShamanSim(player, &proto.Debuffs{}, 30)
	return sh
}

// TestForeverFiveThunders: pieces of the three variants count toward one set. 3 pieces give 8
// mana per 5 sec, 4 pieces The Furious Storm proc and 6 pieces 23 spell power.
func TestForeverFiveThunders(t *testing.T) {
	setStat := func(sh *EnhancementShaman, stat stats.Stat) float64 {
		return sh.GetStat(stat) - sh.EquipStats()[stat]
	}
	two, three := fiveThundersSim(t, 2), fiveThundersSim(t, 3)
	if got := setStat(three, stats.MP5) - setStat(two, stats.MP5); math.Abs(got-8) > 0.01 {
		t.Errorf("the 3 piece gives %.1f mp5, want 8", got)
	}
	furiousStorm := core.ActionID{SpellID: 27775}
	if three.GetAuraByID(furiousStorm) != nil {
		t.Errorf("3 pieces have The Furious Storm")
	}
	if fiveThundersSim(t, 4).GetAuraByID(furiousStorm) == nil {
		t.Errorf("4 pieces lack The Furious Storm")
	}
	five, six := fiveThundersSim(t, 5), fiveThundersSim(t, 6)
	if got := setStat(six, stats.SpellPower) - setStat(five, stats.SpellPower); math.Abs(got-23) > 0.01 {
		t.Errorf("the 6 piece gives %.1f spell power, want 23", got)
	}
}
