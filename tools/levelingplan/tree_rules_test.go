package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

func readTreeFile(t *testing.T, path string) *proto.LevelingTalentShape {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	shape, err := treeFileShape(data)
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return shape
}

func talentIndex(t *testing.T, shape *proto.LevelingTalentShape, tree int, name string) int32 {
	t.Helper()
	for i, talent := range shape.Trees[tree].Talents {
		if talent.Name == name {
			return int32(i)
		}
	}
	t.Fatalf("no %s in %s", name, shape.Trees[tree].Name)
	return -1
}

// TestShippedRoutesKeepTheRules: every leveling plan a spec ships (ui/<spec>/apls/leveling_plan*.json)
// keeps its class's talent tree rules at every pick, with the trees as they are now. A tree
// change that moves a talent the route takes early fails here, not in game.
func TestShippedRoutesKeepTheRules(t *testing.T) {
	files, err := filepath.Glob("../../ui/*/apls/leveling_plan*.json")
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		var raw struct{ Route []json.RawMessage }
		if err := json.Unmarshal(data, &raw); err != nil {
			t.Fatalf("%s: %v", file, err)
		}
		if len(raw.Route) == 0 {
			continue
		}
		spec := filepath.Base(filepath.Dir(filepath.Dir(file)))
		class := spec[strings.LastIndex(spec, "_")+1:]
		shape := readTreeFile(t, "../../ui/core/talents/trees/"+class+".json")
		build := &proto.LevelingTalentBuild{Level: 10}
		for _, r := range raw.Route {
			pick := &proto.LevelingTalentPick{}
			if err := protojson.Unmarshal(r, pick); err != nil {
				t.Fatalf("%s: %v", file, err)
			}
			build.Picks = append(build.Picks, pick)
		}
		plan := &proto.LevelingPlanSettings{FromLevel: 10, ToLevel: 61, Builds: []*proto.LevelingTalentBuild{build}}
		if err := routeProblem(plan, shape); err != nil {
			t.Errorf("%s: %v", file, err)
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("found no leveling plan routes to check")
	}
}

// TestRouteProblem: the mistake of 2026-11-04, Shamanistic Focus (Enhancement row 3) with only
// 8 points in the tree, is caught with its level and why. Past the plan's end it doesn't count.
func TestRouteProblem(t *testing.T) {
	shape := readTreeFile(t, "../../ui/core/talents/trees/shaman.json")
	thundering := talentIndex(t, shape, 1, "Thundering Strikes")
	dexterity := talentIndex(t, shape, 1, "Mental Dexterity")
	focus := talentIndex(t, shape, 1, "Shamanistic Focus")
	var picks []*proto.LevelingTalentPick
	for i := 0; i < 5; i++ {
		picks = append(picks, &proto.LevelingTalentPick{Tree: 1, Talent: thundering})
	}
	for i := 0; i < 3; i++ {
		picks = append(picks, &proto.LevelingTalentPick{Tree: 1, Talent: dexterity})
	}
	picks = append(picks, &proto.LevelingTalentPick{Tree: 1, Talent: focus})
	plan := &proto.LevelingPlanSettings{FromLevel: 1, ToLevel: 60, Builds: []*proto.LevelingTalentBuild{{Level: 10, Picks: picks}}}

	err := routeProblem(plan, shape)
	if err == nil || !strings.Contains(err.Error(), "Shamanistic Focus at level 18") || !strings.Contains(err.Error(), "row isn't open") {
		t.Errorf("got %v, want Shamanistic Focus at level 18 with its row not open", err)
	}
	plan.ToLevel = 18
	if err := routeProblem(plan, shape); err != nil {
		t.Errorf("a plan that ends before the pick: got %v, want nil", err)
	}
}

// TestShapeDifference: the trees from before November 2 (Improved Ghost Wolf on row 2) aren't
// today's, and today's match themselves.
func TestShapeDifference(t *testing.T) {
	now := readTreeFile(t, "../../ui/core/talents/trees/shaman.json")
	old := readTreeFile(t, "testdata/shaman_tree_70707.json")
	if diff := shapeDifference(old, now); diff == "" {
		t.Error("the 70707 trees read the same as today's")
	}
	if diff := shapeDifference(now, now); diff != "" {
		t.Errorf("today's trees differ from themselves: %s", diff)
	}
}
