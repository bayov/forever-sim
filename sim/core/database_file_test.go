//go:build with_db

package core

import (
	"slices"
	"testing"

	"github.com/wowsims/classic/assets/database"
	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// The sim decodes db.bin as SimDatabaseFile, which must give each field its namesake's number
// and type in UIDatabase and UIItem, and so the same values.
func TestSimDatabaseFileMatchesUI(t *testing.T) {
	sameFields := func(slim, full protoreflect.MessageDescriptor) {
		fields := slim.Fields()
		for i := 0; i < fields.Len(); i++ {
			f := fields.Get(i)
			u := full.Fields().ByName(f.Name())
			switch {
			case u == nil:
				t.Errorf("%s.%s: %s has no such field", slim.Name(), f.Name(), full.Name())
			case f.Number() != u.Number() || f.Kind() != u.Kind() || f.Cardinality() != u.Cardinality() || f.IsPacked() != u.IsPacked():
				t.Errorf("%s.%s is %d %v %v, %s's is %d %v %v", slim.Name(), f.Name(), f.Number(), f.Cardinality(), f.Kind(), full.Name(), u.Number(), u.Cardinality(), u.Kind())
			case f.Enum() != nil && f.Enum().FullName() != u.Enum().FullName():
				t.Errorf("%s.%s is a %s, %s's a %s", slim.Name(), f.Name(), f.Enum().FullName(), full.Name(), u.Enum().FullName())
			case f.Message() != nil && f.Message().FullName() != u.Message().FullName() && f.Message().FullName() != "proto.SimDatabaseFileItem":
				t.Errorf("%s.%s is a %s, %s's a %s", slim.Name(), f.Name(), f.Message().FullName(), full.Name(), u.Message().FullName())
			}
		}
	}
	sameFields((&proto.SimDatabaseFile{}).ProtoReflect().Descriptor(), (&proto.UIDatabase{}).ProtoReflect().Descriptor())
	sameFields((&proto.SimDatabaseFileItem{}).ProtoReflect().Descriptor(), (&proto.UIItem{}).ProtoReflect().Descriptor())

	full, slim := database.Load(), database.LoadSim()
	if len(full.Items) != len(slim.Items) || len(full.Enchants) != len(slim.Enchants) || len(full.RandomSuffixes) != len(slim.RandomSuffixes) {
		t.Fatalf("SimDatabaseFile has %d items, %d enchants and %d suffixes, UIDatabase %d, %d and %d",
			len(slim.Items), len(slim.Enchants), len(slim.RandomSuffixes), len(full.Items), len(full.Enchants), len(full.RandomSuffixes))
	}
	for i, item := range slim.Items {
		s, u := item.ProtoReflect(), full.Items[i].ProtoReflect()
		fields := s.Descriptor().Fields()
		for j := 0; j < fields.Len(); j++ {
			f := fields.Get(j)
			if !s.Get(f).Equal(u.Get(u.Descriptor().Fields().ByName(f.Name()))) {
				t.Fatalf("item %d: %s differs", item.Id, f.Name())
			}
		}
	}
	for i := range slim.Enchants {
		if slim.Enchants[i].EffectId != full.Enchants[i].EffectId || !slices.Equal(slim.Enchants[i].Stats, full.Enchants[i].Stats) {
			t.Fatalf("enchant %d differs", full.Enchants[i].EffectId)
		}
	}
	for i := range slim.RandomSuffixes {
		if slim.RandomSuffixes[i].Id != full.RandomSuffixes[i].Id || slim.RandomSuffixes[i].Name != full.RandomSuffixes[i].Name || !slices.Equal(slim.RandomSuffixes[i].Stats, full.RandomSuffixes[i].Stats) {
			t.Fatalf("random suffix %d differs", full.RandomSuffixes[i].Id)
		}
	}
}
