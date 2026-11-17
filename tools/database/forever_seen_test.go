package database

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// An entry in forever_seen_in_game.json lists all of an item's stats, so a stat the entry leaves
// out, or gives as 0, is cleared on the item, and so are hit and crit it doesn't give. Bad Hat
// (260212) lost its +23 Shadow damage that way in the 2026-11-13 hotfixes (wago push 114063).
func TestForeverSeenInGameClearsStats(t *testing.T) {
	var old stats.Stats
	old[stats.Intellect], old[stats.Spirit], old[stats.ShadowPower], old[stats.MP5] = 17, 17, 23, 4
	db := NewWowDatabase()
	db.Items[260212] = &proto.UIItem{Id: 260212, Name: "Bad Hat", Type: proto.ItemType_ItemTypeHead,
		Stats: old.ToFloatArray(), HitRating: 10, CritRating: 10, Ilvl: 56}
	db.Items[1] = &proto.UIItem{Id: 1, Name: "Untouched", Stats: old.ToFloatArray(), HitRating: 10}

	path := filepath.Join(t.TempDir(), "seen.json")
	entry := `{"items": {"260212": {"name": "Bad Hat", "ilvl": 56, "quality": 3,
		"stats": {"Intellect": 17, "Spirit": 17, "MP5": 0}}}}`
	if err := os.WriteFile(path, []byte(entry), 0o644); err != nil {
		t.Fatal(err)
	}
	db.MergeItems(ReadForeverSeenInGame(path, db.Items))

	hat := db.Items[260212]
	var want stats.Stats
	want[stats.Intellect], want[stats.Spirit] = 17, 17
	if got := stats.FromFloatArray(hat.Stats); got != want {
		t.Errorf("Bad Hat has %s, want %s", got.FlatString(), want.FlatString())
	}
	if hat.HitRating != 0 || hat.CritRating != 0 {
		t.Errorf("Bad Hat has hit %v and crit %v, want none", hat.HitRating, hat.CritRating)
	}
	if hat.Type != proto.ItemType_ItemTypeHead {
		t.Errorf("Bad Hat is type %v, want it kept as head", hat.Type)
	}
	if other := db.Items[1]; stats.FromFloatArray(other.Stats) != old || other.HitRating != 10 {
		t.Errorf("an item with no entry changed: %s, hit %v", stats.FromFloatArray(other.Stats).FlatString(), other.HitRating)
	}
}
