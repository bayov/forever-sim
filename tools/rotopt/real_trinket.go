package main

import (
	"fmt"
	"os"
	"slices"
	"sync"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	goproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// The DPS templates write their on-use trinket as a synthetic item (warrior_common.go's
// battleTrinket, mage_common.go's mageSyntheticTalisman, ...), from the days the presets wore
// one. The Phase 0 sets wear real trinkets, so every run and every rotation we write casts the
// gear's own on-use trinkets in their place. A rotation's synthetic trinkets, in the order they
// first appear, take the worn trinkets with a use in slot order (the first trinket slot's, then
// the second's), so a template with two (warrior_tank.go's bulwark and battle trinkets) casts both
// of a set's on-use trinkets. One with more synthetic trinkets than the gear has uses puts the
// rest on the last of them. With no such trinket the lines that cast a synthetic one go. Gear
// that holds a synthetic trinket (the tank and healer sets) keeps the rotation as built.

var (
	trinketMu     sync.Mutex
	trinketProbed = map[int32]bool{}
)

// useTrinkets is the request's trinkets with a use, in slot order (none if none), and whether the
// gear holds a synthetic trinket. The sim only knows an item has a use (core.ItemHasUse) after it
// built a character wearing it, so a trinket we haven't seen costs a one iteration, one second run.
func useTrinkets(request *proto.RaidSimRequest) ([]int32, bool) {
	slots := equipmentSlots(request.Raid.Parties[0].Players[0].Equipment)
	var ids []int32
	for _, slot := range []proto.ItemSlot{proto.ItemSlot_ItemSlotTrinket1, proto.ItemSlot_ItemSlotTrinket2} {
		if int(slot) < len(slots) && slots[slot] != nil {
			if isSyntheticItem(slots[slot].Id) {
				return nil, true
			}
			ids = append(ids, slots[slot].Id)
		}
	}
	trinketMu.Lock()
	defer trinketMu.Unlock()
	if slices.ContainsFunc(ids, func(id int32) bool { return !trinketProbed[id] }) {
		probe := goproto.Clone(request).(*proto.RaidSimRequest)
		pl := probe.Raid.Parties[0].Players[0]
		pl.ItemUsesOff, pl.ItemEffectsOff = nil, nil
		probe.Encounter.Duration = 1
		probe.Encounter.DurationVariation = 0
		probe.SimOptions.Iterations = 1
		if res := core.RunRaidSim(probe); res.Error != nil {
			fmt.Fprintf(os.Stderr, "probing the trinkets' uses: %s\n", res.Error.Message)
		}
		for _, id := range ids {
			trinketProbed[id] = true
		}
	}
	return slices.DeleteFunc(ids, func(id int32) bool { return !core.ItemHasUse(id) }), false
}

// setRotation puts rot on the request's player, with the gear's real on-use trinkets in place of
// the synthetic ones (see above). rot itself is left as is.
func setRotation(request *proto.RaidSimRequest, rot *proto.APLRotation) {
	request.Raid.Parties[0].Players[0].Rotation = withRealTrinket(rot, request)
}

func withRealTrinket(rot *proto.APLRotation, request *proto.RaidSimRequest) *proto.APLRotation {
	if rot == nil || !hasSyntheticItem(rot.ProtoReflect()) {
		return rot
	}
	real, keep := useTrinkets(request)
	if keep {
		return rot
	}
	rot = goproto.Clone(rot).(*proto.APLRotation)
	if len(real) > 0 {
		replaceSyntheticItems(rot.ProtoReflect(), real)
		return rot
	}
	castsSynthetic := func(a *proto.APLAction) bool {
		return isSyntheticItem(a.GetCastSpell().GetSpellId().GetItemId())
	}
	rot.PrepullActions = slices.DeleteFunc(rot.PrepullActions, func(p *proto.APLPrepullAction) bool { return castsSynthetic(p.Action) })
	rot.PriorityList = slices.DeleteFunc(rot.PriorityList, func(l *proto.APLListItem) bool { return castsSynthetic(l.Action) })
	return rot
}

// actionIDs calls f on every ActionID in m.
func actionIDs(m protoreflect.Message, f func(*proto.ActionID)) {
	if id, ok := m.Interface().(*proto.ActionID); ok {
		f(id)
		return
	}
	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if fd.Kind() != protoreflect.MessageKind {
			return true
		}
		switch {
		case fd.IsList():
			for i, l := 0, v.List(); i < l.Len(); i++ {
				actionIDs(l.Get(i).Message(), f)
			}
		case fd.IsMap():
		default:
			actionIDs(v.Message(), f)
		}
		return true
	})
}

func hasSyntheticItem(m protoreflect.Message) bool {
	found := false
	actionIDs(m, func(id *proto.ActionID) { found = found || isSyntheticItem(id.GetItemId()) })
	return found
}

// replaceSyntheticItems puts real's trinkets in place of m's synthetic items: the first synthetic
// id m names (in actionIDs' order) becomes real[0], the second real[1], and any past the end of
// real the last of it.
func replaceSyntheticItems(m protoreflect.Message, real []int32) {
	var order []int32
	actionIDs(m, func(id *proto.ActionID) {
		item := id.GetItemId()
		if !isSyntheticItem(item) {
			return
		}
		i := slices.Index(order, item)
		if i < 0 {
			i = len(order)
			order = append(order, item)
		}
		id.RawId = &proto.ActionID_ItemId{ItemId: real[min(i, len(real)-1)]}
	})
}
