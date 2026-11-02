//go:build practice

package main

import (
	"encoding/json"
	"flag"
	"os"
	"testing"
)

var update = flag.Bool("update", false, "rewrite the golden files")

// golden compares v as JSON with testdata/<name>.golden.json.
func golden(t *testing.T, name string, v any) {
	t.Helper()
	got, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	path := "testdata/" + name + ".golden.json"
	if *update {
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (run with -update)", err)
	}
	if string(got) != string(want) {
		t.Errorf("%s differs from its golden file:\n%s", name, got)
	}
}

func testClient(t *testing.T) *wowClient {
	c, err := findClient("testdata/wow")
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestLuaParser(t *testing.T) {
	vars, err := parseSavedVariables(`
-- a comment
A = { 1, "two", [5] = true, ["k"] = { x = -1.5, y = "q\"uote\n" }, nested = { {}, }, }
B = 0x10
C = nil
D = 'single\\'
`)
	if err != nil {
		t.Fatal(err)
	}
	golden(t, "lua", vars)
}

// Each GUID's newest copy wins across accounts, and its bindings and CVars come from that
// copy's account.
func TestRecordsAndBindings(t *testing.T) {
	c := testClient(t)
	records, err := c.records()
	if err != nil {
		t.Fatal(err)
	}
	golden(t, "records", records)
	if len(records) != 1 || records[0].Name != "Bayov Test" || records[0].Account != "1#2" {
		t.Fatalf("records: %+v", records)
	}
	golden(t, "bindings_1_1", c.bindings("1#1", ""))
	golden(t, "bindings_1_2", c.bindings("1#2", "Bayov Test"))
	golden(t, "cvars_1_1", c.cvars("1#1", ""))
}

// The real BayovCore files parse, when this PC has them.
func TestRealClientFiles(t *testing.T) {
	c, err := findClient("/mnt/d/Gaming/World of Warcraft")
	if err != nil {
		t.Skip("no WoW client here")
	}
	records, err := c.records()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) == 0 {
		t.Skip("no BayovCore records")
	}
	for _, rec := range records {
		t.Logf("%s %s (%s) level %d, %d slots, account %s", rec.GUID, rec.Name, rec.Class, rec.Level, len(rec.Slots), rec.Account)
		if len(c.bindings(rec.Account, rec.Name)) < 30 {
			t.Errorf("%s: few bindings", rec.Name)
		}
	}
}

// The class keybind notes give each key's planned spell, by WoW's key names.
func TestNoteBinds(t *testing.T) {
	golden(t, "note_binds", testClient(t).noteBinds("SHAMAN"))
}

// The real notes and EllesmereUI profile read, when this PC has them.
func TestRealUIFiles(t *testing.T) {
	c, err := findClient("/mnt/d/Gaming/World of Warcraft")
	if err != nil {
		t.Skip("no WoW client here")
	}
	binds := c.noteBinds("SHAMAN")
	t.Logf("%d shaman note binds, SHIFT-Q is %q", len(binds), binds["SHIFT-Q"])
	for _, account := range c.accounts() {
		if ui := c.ellesmereUI(account); ui != nil {
			if ui["actionBars"] == nil || ui["unitFrames"] == nil {
				t.Errorf("%s: EllesmereUI profile %v has no bars or unit frames", account, ui["profile"])
			}
		}
	}
}
