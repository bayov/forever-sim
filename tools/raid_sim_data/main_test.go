package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// The page's data file is what the search's pool gives now (races, players, variants), so the page sims
// each member as the search does. When it fails, run: go run -tags=with_db ./tools/raid_sim_data
func TestDataFileIsFresh(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	want, _, pool := generate(root)
	got, err := os.ReadFile(filepath.Join(root, "ui", "raid_sim", "data", "builds.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Error("ui/raid_sim/data/builds.json is stale: run go run -tags=with_db ./tools/raid_sim_data")
	}
	// The page's table is the search's (sim/compsearch/table.go).
	got, err = os.ReadFile(filepath.Join(root, tableFile))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, table(pool)) {
		t.Error(tableFile + " is stale: run go run -tags=with_db ./tools/raid_sim_data")
	}
}

// The built-in raids are the committed write-ups' best comps. When it fails, run the same command.
func TestPresetsFileIsFresh(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	want, _ := presets(root)
	got, err := os.ReadFile(filepath.Join(root, presetsFile))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Error(presetsFile + " is stale: run go run -tags=with_db ./tools/raid_sim_data")
	}
}
