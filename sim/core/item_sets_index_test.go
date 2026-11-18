//go:build with_db

package core

import (
	"fmt"
	"testing"
)

// NewItemSet looks sets up in an index of the items' set IDs and names, built once. It must
// find what a scan of every item finds, and panic the same way for a set no item is in.
func TestNewItemSetIndexMatchesScan(t *testing.T) {
	setIndexOnce.Do(indexItemSets)
	check := func(set ItemSet) {
		foundID := set.ID == 0 || itemSetIDs[set.ID]
		foundName := set.Name != "" && itemSetNames[set.Name]
		foundAlternativeName := set.AlternativeName == "" || itemSetNames[set.AlternativeName]
		scanID, scanName, scanAlternativeName := findSetItems(&set)
		if foundID != scanID || foundName != scanName || foundAlternativeName != scanAlternativeName {
			t.Errorf("set %+v: index found %v %v %v, scan %v %v %v", set, foundID, foundName, foundAlternativeName, scanID, scanName, scanAlternativeName)
		}
	}
	type key struct {
		id   int32
		name string
	}
	seen := map[key]bool{}
	n := 0
	for _, item := range ItemsByID {
		k := key{item.SetID, item.SetName}
		if k == (key{}) || seen[k] {
			continue
		}
		seen[k] = true
		n++
		check(ItemSet{ID: item.SetID, Name: item.SetName})
		check(ItemSet{Name: item.SetName, AlternativeName: item.SetName})
		check(ItemSet{ID: item.SetID + 1_000_000, Name: item.SetName + "?", AlternativeName: item.SetName})
	}
	if n == 0 {
		t.Fatal("no set items in the database")
	}
}

func TestNewItemSetPanics(t *testing.T) {
	var name string
	var id int32
	for _, item := range ItemsByID {
		if item.SetName != "" && item.SetID > 0 {
			name, id = item.SetName, item.SetID
			break
		}
	}
	if name == "" {
		t.Fatal("no item with both a set name and a set ID")
	}

	for _, tc := range []struct {
		set  ItemSet
		want string
	}{
		{ItemSet{ID: 987654, Name: name}, "No items found for set id 987654"},
		{ItemSet{Name: "No Such Set"}, "No items found for set No Such Set"},
		{ItemSet{Name: "No Such Set", AlternativeName: name}, "No items found for set No Such Set"},
		{ItemSet{Name: name, AlternativeName: "No Such Alternative"}, "No items found for set alternative No Such Alternative"},
		{ItemSet{ID: id, Name: "No Such Set", AlternativeName: "No Such Alternative"}, "No items found for set alternative No Such Alternative"},
	} {
		n := len(sets)
		got := func() (msg string) {
			defer func() { msg = fmt.Sprint(recover()) }()
			NewItemSet(tc.set)
			return "no panic"
		}()
		if got != tc.want {
			t.Errorf("NewItemSet(%+v) panicked with %q, want %q", tc.set, got, tc.want)
		}
		if len(sets) != n {
			t.Errorf("NewItemSet(%+v) registered the set although it panicked", tc.set)
			sets = sets[:n]
		}
	}
}

// A set whose items were added after the index was built is still found, by a scan.
func TestNewItemSetItemsAddedLater(t *testing.T) {
	setIndexOnce.Do(indexItemSets)
	const id = 2_000_001
	ItemsByID[id] = Item{ID: id, SetName: "Test Set Added Later", SetID: 987655}
	defer delete(ItemsByID, id)
	n := len(sets)
	defer func() { sets = sets[:n] }()

	NewItemSet(ItemSet{Name: "Test Set Added Later"})
	NewItemSet(ItemSet{ID: 987655, Name: "Test Set Added Later Too"}) // a set Forever added, found by ID
	NewItemSet(ItemSet{Name: "Test Set Added Later", AlternativeName: "Test Set Added Later"})
	if len(sets) != n+3 {
		t.Fatalf("registered %d sets, want 3", len(sets)-n)
	}
}
