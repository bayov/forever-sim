//go:build with_db

package item_sets

import (
	"slices"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// Forever added the Imperial Plate Gauntlets (250589) to Imperial Plate. The database gives it only
// the client's set ID, and it still counts toward the set's Forever 6-piece bonus.
func TestForeverImperialPlateGauntlets(t *testing.T) {
	character := &core.Character{}
	for slot, id := range map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotHead:  12427,
		proto.ItemSlot_ItemSlotChest: 12422,
		proto.ItemSlot_ItemSlotWrist: 12425,
		proto.ItemSlot_ItemSlotWaist: 12424,
		proto.ItemSlot_ItemSlotLegs:  12429,
		proto.ItemSlot_ItemSlotHands: 250589,
	} {
		item, ok := core.ItemsByID[id]
		if !ok {
			t.Fatalf("no item %d in the database", id)
		}
		character.Equipment[slot] = item
	}
	names := character.GetActiveSetBonusNames()
	for _, want := range []string{"Imperial Plate (5pc)", "Imperial Plate (6pc)"} {
		if !slices.Contains(names, want) {
			t.Errorf("active set bonuses %v lack %s", names, want)
		}
	}
}
