package core

import (
	"slices"

	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// RotationReferencesSpell is whether the rotation names any of the spell IDs anywhere: as a spell
// it casts, an aura it checks or any other action ID.
//
// A class registers an ability only a few rotations use with it, so the auras it would add
// (which the results list even when they never go up) don't change the result of a rotation
// that doesn't use it.
func RotationReferencesSpell(rotation *proto.APLRotation, spellIDs ...int32) bool {
	if rotation == nil {
		return false
	}
	actionIDName := (&proto.ActionID{}).ProtoReflect().Descriptor().FullName()
	found := false
	var walk func(m protoreflect.Message)
	walk = func(m protoreflect.Message) {
		if found {
			return
		}
		if m.Descriptor().FullName() == actionIDName {
			id := m.Interface().(*proto.ActionID)
			if slices.Contains(spellIDs, id.GetSpellId()) {
				found = true
			}
			return
		}
		m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
			switch {
			case fd.IsList() && fd.Message() != nil:
				list := v.List()
				for i := 0; i < list.Len() && !found; i++ {
					walk(list.Get(i).Message())
				}
			case fd.IsMap():
				// The APL protos have no message maps.
			case fd.Message() != nil:
				walk(v.Message())
			}
			return !found
		})
	}
	walk(rotation.ProtoReflect())
	return found
}
