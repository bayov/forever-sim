//go:build practice

package main

import (
	"strconv"

	"github.com/wowsims/classic/sim/core/proto"
)

// lookAlikes give the made up Synthetic items (tools/database/forever_synthetic_items.go) a real
// item to look like, since the client has no models for them. They're the Era Phase 1 pieces an
// enhancement shaman would wear in their place, and Spinal Reaper for the two-hander.
var lookAlikes = map[int32]int32{
	990001: 18817, // Crown of Destruction
	990003: 12927, // Truestrike Shoulders
	990004: 13340, // Cape of the Black Baron
	990005: 11726, // Savage Gladiator Chain
	990006: 18812, // Wristguards of True Flight
	990007: 15063, // Devilsaur Gauntlets
	990008: 11686, // Girdle of Beastial Fury
	990009: 15062, // Devilsaur Leggings
	990010: 13210, // Pads of the Dread Wolf
	990012: 17104, // Spinal Reaper
}

// look is what the character looks like, for its model: the client's race and class ids and the
// gear by inventory slot (1 head, 3 shoulders, 16 main hand, ...).
func (s *server) look(player *proto.Player) map[string]any {
	races := map[proto.Race]int{
		proto.Race_RaceHuman: 1, proto.Race_RaceOrc: 2, proto.Race_RaceDwarf: 3, proto.Race_RaceNightElf: 4,
		proto.Race_RaceUndead: 5, proto.Race_RaceTauren: 6, proto.Race_RaceGnome: 7, proto.Race_RaceTroll: 8,
	}
	classes := map[proto.Class]int{
		proto.Class_ClassWarrior: 1, proto.Class_ClassPaladin: 2, proto.Class_ClassHunter: 3, proto.Class_ClassRogue: 4,
		proto.Class_ClassPriest: 5, proto.Class_ClassShaman: 7, proto.Class_ClassMage: 8, proto.Class_ClassWarlock: 9,
		proto.Class_ClassDruid: 11,
	}
	slots := map[proto.ItemSlot]int{
		proto.ItemSlot_ItemSlotHead: 1, proto.ItemSlot_ItemSlotShoulder: 3, proto.ItemSlot_ItemSlotBack: 15,
		proto.ItemSlot_ItemSlotChest: 5, proto.ItemSlot_ItemSlotWrist: 9, proto.ItemSlot_ItemSlotHands: 10,
		proto.ItemSlot_ItemSlotWaist: 6, proto.ItemSlot_ItemSlotLegs: 7, proto.ItemSlot_ItemSlotFeet: 8,
		proto.ItemSlot_ItemSlotMainHand: 16, proto.ItemSlot_ItemSlotOffHand: 17, proto.ItemSlot_ItemSlotRanged: 18,
	}
	items := map[string]int32{}
	for i, it := range player.GetEquipment().GetItems() {
		inv, ok := slots[proto.ItemSlot(i)]
		if !ok || it == nil || it.Id <= 0 {
			continue
		}
		id := it.Id
		if alike, ok := lookAlikes[id]; ok {
			id = alike
		} else if id >= 990000 && id < 1000000 {
			// A made up item with no look-alike yet: the slot shows bare.
			continue
		}
		items[strconv.Itoa(inv)] = id
	}
	out := map[string]any{"race": races[player.GetRace()], "class": classes[player.GetClass()], "items": items}
	// The main hand's weapon subclass picks the swing and impact sounds (1 two-hand axe, 7 one-hand
	// sword, ...). It's -1 for bare hands.
	out["weaponSubclass"] = -1
	if mh, ok := items["16"]; ok && s.db2 != nil {
		if row := s.db2.MustTable("Item").Get(int(mh)); row.Ok() {
			out["weaponSubclass"] = row.Int("SubclassID")
		}
	}
	return out
}
