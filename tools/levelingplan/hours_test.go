package main

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wowsims/classic/assets/database"
	"github.com/wowsims/classic/sim"
	"github.com/wowsims/classic/sim/core/proto"
)

// TestHours: two copies of a plan play the same gear and sessions, so they differ by nothing, and
// the levels add up to the bands and the plan.
func TestHours(t *testing.T) {
	exampleRequest(t, "undead_warrior_arms_two_hand")
	sim.RegisterAll()
	out := filepath.Join(t.TempDir(), "hours.json")
	files := []string{"configs/undead_warrior_arms_two_hand.json", "configs/undead_warrior_arms_two_hand.json", "configs/undead_warrior_prot_shield.json"}
	opts := hoursOptions{firstGearSeed: 1, gearSeeds: 2, sessions: 5, seed: 1}
	if err := hoursFiles(database.Load(), files, out, opts, []int32{8, 11}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var results []planHours
	if err := json.Unmarshal(data, &results); err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Fatalf("%d results", len(results))
	}
	if d := results[1].AgainstFirst; d == nil || d.Mean != 0 || d.Error != 0 {
		t.Errorf("the same plan is %v against itself", d)
	}
	if d := results[2].AgainstFirst; d == nil || d.Mean == 0 {
		t.Errorf("another plan is %v against the first", d)
	}
	r := results[0]
	if len(r.Levels) != 4 || len(r.Bands) != 2 || r.Bands[0].Level != 8 || r.Bands[0].To != 9 || r.Bands[1].To != 11 {
		t.Fatalf("%d levels and bands %+v", len(r.Levels), r.Bands)
	}
	levels, bands := 0.0, 0.0
	for _, row := range r.Levels {
		levels += row.Hours.Mean
		if shares := row.CombatShare + row.RestShare + row.SearchShare + row.DeadShare; shares < 0.99 || shares > 1.01 {
			t.Errorf("level %d: the shares add up to %.3f", row.Level, shares)
		}
	}
	for _, row := range r.Bands {
		bands += row.Hours.Mean
	}
	if math.Abs(levels-r.Hours.Mean) > 1e-3 || math.Abs(bands-r.Hours.Mean) > 1e-3 {
		t.Errorf("the levels take %.4fh and the bands %.4fh, the plan %.4fh", levels, bands, r.Hours.Mean)
	}
}

// TestMeasureWeights: attack power is 1, Strength about 2 and weapon DPS about 14 attack power
// for a warrior, and a style's weights have that style's stats.
func TestMeasureWeights(t *testing.T) {
	sim.RegisterAll()
	req := exampleRequest(t, "undead_warrior_prot_shield")
	req.Plan.FromLevel, req.Plan.ToLevel = 14, 16
	style := proto.LevelingWeaponStyle_LevelingWeaponStyleOneHandShield
	opts := weightsOptions{hoursOptions: hoursOptions{firstGearSeed: 1, gearSeeds: 1, sessions: 150, seed: 1}, rounds: 2, style: &style}
	rounds := 0
	file, err := measureWeights(context.Background(), database.Load(), req, opts, func(int, *weightsFile) { rounds++ })
	if err != nil {
		t.Fatal(err)
	}
	if rounds != 2 || len(file.Levels) != 1 || file.Levels[0].Level != 15 {
		t.Fatalf("%d rounds, levels %+v", rounds, file.Levels)
	}
	w := file.Levels[0].Weights
	if w["AttackPower"] != 1 || w["Strength"] < 1.5 || w["Strength"] > 2.6 || w["MainHandDps"] < 11 || w["MainHandDps"] > 17 {
		t.Errorf("weights %v", w)
	}
	if _, ok := w["BlockValue"]; !ok {
		t.Errorf("no block value with a shield: %v", w)
	}
	if _, ok := w["OffHandDps"]; ok {
		t.Errorf("off hand DPS with a shield: %v", w)
	}
	// What the config's weights and the UI's presets read.
	if _, err := file.presets(style); err != nil {
		t.Error(err)
	}
}

// TestWeightsFill: hit is the larger of 1% more and 1% less, and a weight below 0 is 0.
func TestWeightsFill(t *testing.T) {
	row := weightsRow{Level: 30}
	gains := map[string]float64{"AttackPower": 0.1, "MeleeHit": 0.2, "MeleeHit-": 1.5, "Spirit": -0.01}
	if err := row.fill(gains, map[string]float64{"AttackPower": 0.01, "MeleeHit-": 0.3}); err != nil {
		t.Fatal(err)
	}
	if row.Weights["MeleeHit"] != 15 || row.Errors["MeleeHit"] != 3 || row.Weights["Spirit"] != 0 {
		t.Errorf("weights %v, errors %v", row.Weights, row.Errors)
	}
	if _, ok := row.Weights["MeleeHit-"]; ok {
		t.Error("the 1% less probe has a weight of its own")
	}
	row = weightsRow{Level: 30}
	if err := row.fill(map[string]float64{"AttackPower": 0.001}, map[string]float64{"AttackPower": 0.01}); err == nil || !strings.Contains(err.Error(), "too close to nothing") {
		t.Errorf("attack power within its noise: %v", err)
	}
}
