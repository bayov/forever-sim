package database

import (
	"slices"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
)

// The Forever client's ItemSparse class mask goes over Classic's, a row only the ItemSet
// table lists keeps Classic's, and a mask with every class limits nothing.
func TestForeverClientClassMask(t *testing.T) {
	wago := map[int32]WagoDbItem{
		1: {ClassMask: 16},
		2: {ClassMask: 4},
	}
	OverlayForeverClientItems(wago, "ID,Name,ItemLevel,Flags_1,RequiredPVPRank,ItemSet,DamageType,AllowableClass\n"+
		"289296,Benediction,40,0,0,0,0,16\n"+
		"1,Open,40,0,0,0,0,-1\n"+
		"2,,,,,0,,\n"+
		"3,Everyone,40,0,0,0,0,1535\n")

	if got := wago[289296].ClassAllowlist(); !slices.Equal(got, []proto.Class{proto.Class_ClassPriest}) {
		t.Errorf("Benediction's limit is %v, want priests", got)
	}
	if got := wago[1].ClassAllowlist(); got != nil {
		t.Errorf("an item the Forever client lets every class wear has limit %v", got)
	}
	if got := wago[2].ClassAllowlist(); !slices.Equal(got, []proto.Class{proto.Class_ClassHunter}) {
		t.Errorf("an ItemSet only row lost Classic's limit: %v", got)
	}
	if got := wago[3].ClassAllowlist(); got != nil {
		t.Errorf("a mask with every class has limit %v", got)
	}
}
