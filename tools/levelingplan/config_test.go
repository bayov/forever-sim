package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

var exampleConfigs = []string{"undead_warrior_arms_two_hand", "undead_warrior_fury_dual_wield", "undead_warrior_prot_shield"}

func exampleRequest(t *testing.T, name string) *proto.LevelingPlanSearchRequest {
	t.Helper()
	// The route comes from tools/leveling_route. Without it we can't make the examples' requests.
	if _, err := os.Stat("../../ui/warrior/leveling_route.json"); err != nil {
		t.Skip(err)
	}
	req, err := requestFromConfig("configs/" + name + ".json")
	if err != nil {
		t.Fatal(err)
	}
	return req
}

func configFile(t *testing.T, text string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "plan.json")
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestExampleConfigs: the example warriors have 51 points from 5 to 55 and legal trees, and the
// rotation gets a range where we learn Charge.
func TestExampleConfigs(t *testing.T) {
	for _, name := range exampleConfigs {
		req := exampleRequest(t, name)
		if n := len(req.Plan.Builds[0].Picks); n != maxTalentPoints {
			t.Errorf("%s: %d talent points, want %d", name, n, maxTalentPoints)
		}
		if got := planTalentsAt(req.Plan, 4, req.TalentShape); strings.Trim(got, "-0") != "" {
			t.Errorf("%s: talents %q at 4, before the first point", name, got)
		}
		if got := planTalentsAt(req.Plan, 5, req.TalentShape); strings.Trim(got, "-0") != "1" {
			t.Errorf("%s: talents %q at 5, want one point", name, got)
		}
		if len(req.Plan.Rotations) < 2 || req.Plan.Rotations[0].Level != 1 || req.Plan.Rotations[1].Level != 4 {
			t.Errorf("%s: rotations %v, want one at 1 and one at 4", name, req.Plan.Rotations)
		}
		if len(req.Route.GetQuests()) < 100 {
			t.Errorf("%s: %d quests on the route", name, len(req.Route.GetQuests()))
		}
		if req.Faction != proto.Faction_Horde || req.Base.Encounter.Leveling == nil {
			t.Errorf("%s: faction %s, leveling %v", name, req.Faction, req.Base.Encounter.Leveling)
		}
		// The search reads the request back from its file.
		data, err := protojson.Marshal(req)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := readRequest(data); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

// TestWarriorLevelingConfigs: our warrior plan and the builds we compared it with (the
// "Leveling: ..." presets of ui/tank_warrior) make whole requests from the repo's files. Their
// weapon styles are the build's, and only Fury Prot judges shields by its own weights.
func TestWarriorLevelingConfigs(t *testing.T) {
	shieldWeights := func(req *proto.LevelingPlanSearchRequest) []float64 {
		for _, w := range req.GearWeights {
			if w.Style == proto.LevelingWeaponStyle_LevelingWeaponStyleOneHandShield {
				return w.Weights.Stats
			}
		}
		return nil
	}
	prot := shieldWeights(exampleRequest(t, "undead_warrior_leveling_prot"))
	for name, styles := range map[string]int{"1_60": 2, "prot": 2, "fury_prot": 1, "arms": 1, "dual_wield": 1, "arms_prot_dw": 3} {
		req := exampleRequest(t, "undead_warrior_leveling_"+name)
		if n := len(req.Plan.WeaponStyles); n != styles {
			t.Errorf("%s: %d weapon styles, want %d", name, n, styles)
		}
		if req.Plan.FirstTalentLevel != 5 || len(req.Plan.Rotations) < 10 || req.Base.Encounter.Leveling.RecoveryFromLevel != 20 {
			t.Errorf("%s: first talent at %d, %d rotations, leveling %v", name, req.Plan.FirstTalentLevel, len(req.Plan.Rotations), req.Base.Encounter.Leveling)
		}
		if same := slices.Equal(shieldWeights(req), prot); same == (name == "fury_prot") {
			t.Errorf("%s: shield weights the same as Prot's: %v", name, same)
		}
	}
}

// TestShamanConfig: the shaman's plan preset as a plan file gives its talents, rotations and
// imbues, and the leveling options go through.
func TestShamanConfig(t *testing.T) {
	dir, err := filepath.Abs("../../ui/enhancement_shaman")
	if err != nil {
		t.Fatal(err)
	}
	req, err := requestFromConfig(configFile(t, `{
		"player": {"race": "RaceOrc", "class": "ClassShaman", "enhancementShaman": {"options": {"shamanImbue": "RockbiterWeapon"}}},
		"leveling": {"mobLevelOffset": -1},
		"planFile": "`+dir+`/apls/leveling_plan_audited.json",
		"routeFile": "`+dir+`/leveling_route.json",
		"requiredTalents": [{"talent": "Improved Ghost Wolf", "points": 2, "level": 27}],
		"weights": {"free": "`+dir+`/leveling_weights.json"},
		"gearSeed": 7
	}`))
	if err != nil {
		t.Fatal(err)
	}
	plan := req.Plan
	if len(plan.Builds) != 1 || len(plan.Builds[0].Picks) < 50 || len(plan.Rotations) < 5 || len(plan.RotationLibrary) < 5 {
		t.Errorf("%d builds, %d rotations, %d in the library", len(plan.Builds), len(plan.Rotations), len(plan.RotationLibrary))
	}
	if plan.Rotations[0].Imbue != proto.WeaponImbue_RockbiterWeapon {
		t.Errorf("imbue %s", plan.Rotations[0].Imbue)
	}
	if r := plan.RequiredTalents[0]; r.Tree != 1 || r.Talent != 12 {
		t.Errorf("Improved Ghost Wolf is %v, want tree 1 talent 12", r)
	}
	if req.Base.Encounter.Leveling.MobLevelOffset != -1 || req.GearPlan.Seed != 7 || len(req.Route.Quests) < 100 || len(req.GearWeights) < 10 {
		t.Errorf("leveling %v, gear seed %d, %d weights", req.Base.Encounter.Leveling, req.GearPlan.Seed, len(req.GearWeights))
	}
}

// TestConfigMistakes: a typo stops with what's wrong.
func TestConfigMistakes(t *testing.T) {
	const player = `"player": {"race": "RaceUndead", "class": "ClassWarrior", "tankWarrior": {}}`
	for _, c := range []struct{ config, want string }{
		{`{` + player + `, "style": []}`, `unknown field "style"`},
		{`{"player": {"race": "RaceUndead", "class": "ClassWarrior"}}`, "needs a race, a class and a spec"},
		{`{` + player + `, "styles": [{"level": 1, "style": "shields"}]}`, `no weapon style "shields"`},
		{`{` + player + `, "talents": [{"level": 5, "order": ["Crulty:5"]}]}`, `no talent "Crulty"`},
		{`{` + player + `, "talents": [{"level": 5, "order": ["Mortal Strike"]}]}`, "Mortal Strike at level 10"},
		{`{` + player + `, "rotations": [{"level": 1, "template": "shaman_enh_leveling"}]}`, "want one of warrior_leveling"},
		{`{` + player + `, "rotations": [{"level": 1, "knobs": {"restHelth": 1}}]}`, `no knob "restHelth"`},
		{`{` + player + `, "weights": {"shields": "x"}}`, `no weapon style "shields"`},
		{`{` + player + `, "leveling": {"mobs": 1}}`, "leveling:"},
	} {
		_, err := requestFromConfig(configFile(t, c.config))
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: %v, want an error with %q", c.config, err, c.want)
		}
	}
}
