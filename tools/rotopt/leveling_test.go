package main

import (
	"strings"
	"testing"

	"github.com/wowsims/classic/sim"
)

// TestLevelingSession: the Level 30 solo settings run with the leveling template, score
// mobs killed, and the session summary has searches, pulls, fights and rests in order.
func TestLevelingSession(t *testing.T) {
	sim.RegisterAll()
	setup, err := loadSetup("leveling/level30_enh_solo.json")
	if err != nil {
		t.Fatal(err)
	}
	playerLevel = setup.player().Level
	template := shamanEnhLeveling{}
	rot := template.Build(defaultKnobs(template))

	r, err := setup.run(rot, 0, 50)
	if err != nil {
		t.Fatal(err)
	}
	if r.leveling == nil || r.dps < 40 || r.dps != r.leveling.MobsKilled {
		t.Fatalf("scored %s, want the mobs killed in a session", r)
	}

	summary := setup.session(rot, 1)
	for _, kind := range []string{" search ", " pull ", " fight ", " ready ", " rest "} {
		if !strings.Contains(summary, kind) {
			t.Errorf("the session summary has no%sline:\n%s", kind, summary[:min(len(summary), 2000)])
		}
	}
	if strings.Contains(summary, "rotation line") {
		t.Errorf("the template's rotation has warnings:\n%s", summary[:min(len(summary), 2000)])
	}
}
