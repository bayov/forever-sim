package database

import (
	"encoding/json"
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

// "+1 Weapon Damage" and "+20 to All Resistances" aren't in the file's stats, so the entries
// take them from foreverEquipStats, and an entry that gives its own wins. Wowhead's jsonequip
// lacks them too, so ToProto takes them the same way.
func TestForeverEquipStats(t *testing.T) {
	db := NewWowDatabase()
	db.Items[14551] = &proto.UIItem{Id: 14551, Name: "Edgemaster's Handguards", Type: proto.ItemType_ItemTypeHands}
	db.Items[12609] = &proto.UIItem{Id: 12609, Name: "Polychromatic Visionwrap", Type: proto.ItemType_ItemTypeChest}
	db.Items[274614] = &proto.UIItem{Id: 274614, Name: "Knucklebone Ring", Type: proto.ItemType_ItemTypeFinger}

	path := filepath.Join(t.TempDir(), "seen.json")
	entry := `{"items": {
		"14551": {"name": "Edgemaster's Handguards", "stats": {"Expertise": 1, "Armor": 201}},
		"12609": {"name": "Polychromatic Visionwrap", "stats": {"Stamina": 14, "Armor": 87}},
		"274614": {"name": "Knucklebone Ring", "stats": {"Stamina": 7}, "bonusPhysicalDamage": 3}}}`
	if err := os.WriteFile(path, []byte(entry), 0o644); err != nil {
		t.Fatal(err)
	}
	db.MergeItems(ReadForeverSeenInGame(path, db.Items))

	if got := db.Items[14551].BonusPhysicalDamage; got != 1 {
		t.Errorf("Edgemaster's Handguards have %v weapon damage, want 1", got)
	}
	if got := db.Items[274614].BonusPhysicalDamage; got != 3 {
		t.Errorf("Knucklebone Ring has %v weapon damage, want the entry's 3", got)
	}
	var want stats.Stats
	want[stats.Stamina], want[stats.Armor] = 14, 87
	want[stats.ArcaneResistance], want[stats.FireResistance], want[stats.FrostResistance] = 20, 20, 20
	want[stats.NatureResistance], want[stats.ShadowResistance] = 20, 20
	if got := stats.FromFloatArray(db.Items[12609].Stats); got != want {
		t.Errorf("Polychromatic Visionwrap has %s, want %s", got.FlatString(), want.FlatString())
	}

	shield, _ := ForeverWowheadItem{ID: 22198, Name: "Jagged Obsidian Shield", Slot: 14, Class: 4, Subclass: 6,
		Eq: map[string]json.RawMessage{"armor": json.RawMessage("2645")}}.ToProto()
	if got := shield.Stats[proto.Stat_StatShadowResistance]; got != 5 {
		t.Errorf("Jagged Obsidian Shield from Wowhead has %v shadow resistance, want 5", got)
	}
	talon, _ := ForeverWowheadItem{ID: 280012, Name: "Talon of Tor'Rel", Slot: 13, Class: 2, Subclass: 7}.ToProto()
	if got := talon.BonusPhysicalDamage; got != 2 {
		t.Errorf("Talon of Tor'Rel from Wowhead has %v weapon damage, want 2", got)
	}
}
