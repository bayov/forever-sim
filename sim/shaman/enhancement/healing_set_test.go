package enhancement

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
	"github.com/wowsims/classic/sim/shaman"
)

// The synthetic Phase 1 Restoration head and chest carry the Tier 1 Restoration set, the shoulder,
// hands and legs don't (standards 6).
const (
	restoSyntheticHead     = 990201
	restoSyntheticShoulder = 990203
	restoSyntheticChest    = 990205
	restoSyntheticHands    = 990207
	restoSyntheticLegs     = 990209
)

func restoSetSim(t *testing.T, ids ...int32) (*core.Simulation, *EnhancementShaman) {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotLegs+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	slots := map[int32]proto.ItemSlot{
		restoSyntheticHead: proto.ItemSlot_ItemSlotHead, restoSyntheticShoulder: proto.ItemSlot_ItemSlotShoulder,
		restoSyntheticChest: proto.ItemSlot_ItemSlotChest, restoSyntheticHands: proto.ItemSlot_ItemSlotHands,
		restoSyntheticLegs: proto.ItemSlot_ItemSlotLegs,
	}
	for _, id := range ids {
		if _, ok := core.ItemsByID[id]; !ok {
			t.Fatalf("no item %d in the database", id)
		}
		items[slots[id]] = &proto.ItemSpec{Id: id}
	}
	player := newOrcShaman(60, restoRiptidePurif, &proto.EnhancementShaman_Options{})
	player.Equipment = &proto.EquipmentSpec{Items: items}
	return newShamanSim(player, &proto.Debuffs{}, 30)
}

// TestRestorationSyntheticSet (healer_mechanics-20261104-6): the preset's head and chest give the
// Tier 1 2 piece, +10 Spirit (1300960), and nothing else; worn as 4 and 5 pieces the set gives 26
// healing (9 spell power and 17 healing, 1301093) and 1 sec off Riptide's cooldown (1301713).
func TestRestorationSyntheticSet(t *testing.T) {
	for _, id := range []int32{restoSyntheticHead, restoSyntheticChest} {
		if got := core.ItemsByID[id].SetName; got != shaman.ItemSetRestorationSyntheticPhase1.Name {
			t.Errorf("item %d is in set %q, want %q", id, got, shaman.ItemSetRestorationSyntheticPhase1.Name)
		}
	}
	for _, id := range []int32{restoSyntheticShoulder, restoSyntheticHands, restoSyntheticLegs} {
		if got := core.ItemsByID[id].SetName; got != "" {
			t.Errorf("item %d is in set %q, want none (only the 2 piece is live)", id, got)
		}
	}

	_, head := restoSetSim(t, restoSyntheticHead)
	_, two := restoSetSim(t, restoSyntheticHead, restoSyntheticChest)
	_, chest := restoSetSim(t, restoSyntheticChest)
	headSpirit := head.GetStat(stats.Spirit) - head.EquipStats()[stats.Spirit]
	twoSpirit := two.GetStat(stats.Spirit) - two.EquipStats()[stats.Spirit]
	if got := twoSpirit - headSpirit; math.Abs(got-10) > 0.6 {
		t.Errorf("the 2 piece gives %.1f Spirit, want 10 (one piece %.1f)", got, chest.GetStat(stats.Spirit)-chest.EquipStats()[stats.Spirit]-headSpirit)
	}
	if got := two.Riptide.CD.Duration; got != shaman.RiptideNotesCooldown {
		t.Errorf("Riptide's cooldown is %v with 2 pieces, want %v", got, shaman.RiptideNotesCooldown)
	}

	// Three more pieces in the set, for this test only.
	for _, id := range []int32{restoSyntheticShoulder, restoSyntheticHands, restoSyntheticLegs} {
		item := core.ItemsByID[id]
		saved := item
		item.SetName = shaman.ItemSetRestorationSyntheticPhase1.Name
		core.ItemsByID[id] = item
		defer func() { core.ItemsByID[id] = saved }()
	}
	_, four := restoSetSim(t, restoSyntheticHead, restoSyntheticChest, restoSyntheticShoulder, restoSyntheticHands)
	_, fourOff := restoSetSim(t, restoSyntheticHead, restoSyntheticChest, restoSyntheticShoulder)
	if got := four.GetStat(stats.HealingPower) - four.EquipStats()[stats.HealingPower] -
		(fourOff.GetStat(stats.HealingPower) - fourOff.EquipStats()[stats.HealingPower]); math.Abs(got-17) > 0.01 {
		t.Errorf("the 4 piece gives %.1f healing, want 17 on top of 9 spell power", got)
	}
	if got := four.GetStat(stats.SpellPower) - four.EquipStats()[stats.SpellPower] -
		(fourOff.GetStat(stats.SpellPower) - fourOff.EquipStats()[stats.SpellPower]); math.Abs(got-9) > 0.01 {
		t.Errorf("the 4 piece gives %.1f spell power, want 9", got)
	}
	if got := four.Riptide.CD.Duration; got != shaman.RiptideNotesCooldown {
		t.Errorf("Riptide's cooldown is %v with 4 pieces, want %v", got, shaman.RiptideNotesCooldown)
	}
	_, five := restoSetSim(t, restoSyntheticHead, restoSyntheticChest, restoSyntheticShoulder, restoSyntheticHands, restoSyntheticLegs)
	if got, want := five.Riptide.CD.Duration, shaman.RiptideNotesCooldown-time.Second; got != want {
		t.Errorf("Riptide's cooldown is %v with 5 pieces, want %v", got, want)
	}
}
