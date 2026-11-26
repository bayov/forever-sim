package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/classic/assets/database"
	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

func gearRequest(t *testing.T, class proto.Class, weights map[proto.LevelingWeaponStyle]string, routeFile string, seed int32) *proto.LevelingPlanSearchRequest {
	t.Helper()
	rules, err := classRules(class)
	if err != nil {
		t.Fatal(err)
	}
	req := &proto.LevelingPlanSearchRequest{
		Base: &proto.RaidSimRequest{Raid: &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{{
			Class: class,
		}}}}}},
		Plan:               &proto.LevelingPlanSettings{},
		GearPlan:           &proto.LevelingGearPlanSettings{Seed: seed},
		GearDefaultWeights: &proto.UnitStats{},
		Faction:            proto.Faction_Horde,
		ClassRules:         rules,
	}
	for _, style := range []proto.LevelingWeaponStyle{free, twoHand, dualWield, oneHandShield} {
		path, ok := weights[style]
		if !ok {
			continue
		}
		file, err := readWeightsFile(path)
		if err != nil {
			t.Fatal(err)
		}
		presets, err := file.presets(style)
		if err != nil {
			t.Fatal(err)
		}
		req.GearWeights = append(req.GearWeights, presets...)
	}
	if routeFile != "" {
		data, err := os.ReadFile(routeFile)
		if err != nil {
			t.Fatal(err)
		}
		req.Route = &proto.LevelingRoute{}
		if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(data, req.Route); err != nil {
			t.Fatal(err)
		}
	}
	return req
}

const (
	free          = proto.LevelingWeaponStyle_LevelingWeaponStyleFree
	twoHand       = proto.LevelingWeaponStyle_LevelingWeaponStyleTwoHand
	dualWield     = proto.LevelingWeaponStyle_LevelingWeaponStyleDualWield
	oneHandShield = proto.LevelingWeaponStyle_LevelingWeaponStyleOneHandShield

	shamanWeights       = "../../ui/enhancement_shaman/leveling_weights.json"
	shamanShieldWeights = "../../ui/enhancement_shaman/leveling_weights_shield.json"
	shamanRoute         = "../../ui/enhancement_shaman/leveling_route.json"
	warriorWeights      = "testdata/warrior_weights.json"
)

// The Orc shaman's gear plan request, the way the UI makes it for its Leveling 1-60 and Leveling
// 1-60 Shield presets.
func shamanGearRequest(t *testing.T, shield, withRoute bool, seed int32) *proto.LevelingPlanSearchRequest {
	weights, route := shamanWeights, ""
	if shield {
		weights = shamanShieldWeights
	}
	if withRoute {
		route = shamanRoute
	}
	req := gearRequest(t, proto.Class_ClassShaman, map[proto.LevelingWeaponStyle]string{free: weights}, route, seed)
	req.Plan.OneHandAndShield = shield
	return req
}

// A warrior's gear plan request with the weapon styles, from style and level pairs.
func warriorGearRequest(t *testing.T, seed int32, styles ...*proto.LevelingWeaponStyleRange) *proto.LevelingPlanSearchRequest {
	req := gearRequest(t, proto.Class_ClassWarrior, map[proto.LevelingWeaponStyle]string{free: warriorWeights, oneHandShield: shamanShieldWeights}, "", seed)
	req.Plan.WeaponStyles = styles
	return req
}

func from(level int32, style proto.LevelingWeaponStyle) *proto.LevelingWeaponStyleRange {
	return &proto.LevelingWeaponStyleRange{Level: level, Style: style}
}

// gearLines is a gear plan as text, a line for each level: the slots with their items and suffixes.
func gearLines(plan *gearPlan) []string {
	var lines []string
	for _, l := range plan.levels {
		var b strings.Builder
		fmt.Fprintf(&b, "%d", l.level)
		for _, slot := range gearSlots {
			if p, ok := l.gear[slot]; ok {
				fmt.Fprintf(&b, " %d:%d/%d", slot, p.item.Id, p.suffix.GetId())
			}
		}
		lines = append(lines, b.String())
	}
	return lines
}

// TestGearPlanMatchesTS runs the UI's gear plan (leveling_gear.ts) on the same requests as ours and
// compares them level by level. It needs node, node_modules and the generated TS protos, and skips
// without them.
func TestGearPlanMatchesTS(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("no node")
	}
	for _, path := range []string{"../../node_modules/esbuild", "../../ui/core/proto/ui.ts"} {
		if _, err := os.Stat(path); err != nil {
			t.Skip(path + " is missing")
		}
	}
	requests := map[string]*proto.LevelingPlanSearchRequest{}
	for seed := int32(1); seed <= 5; seed++ {
		for _, shield := range []bool{false, true} {
			for _, route := range []bool{false, true} {
				requests[fmt.Sprintf("shaman shield %v route %v seed %d", shield, route, seed)] = shamanGearRequest(t, shield, route, seed)
			}
		}
		for name, styles := range map[string][]*proto.LevelingWeaponStyleRange{
			"none":      nil,
			"free":      {from(1, free)},
			"two-hand":  {from(1, twoHand)},
			"dual":      {from(1, dualWield)},
			"shield":    {from(1, oneHandShield)},
			"switching": {from(40, dualWield), from(1, twoHand), from(25, oneHandShield)},
		} {
			requests[fmt.Sprintf("warrior %s seed %d", name, seed)] = warriorGearRequest(t, seed, styles...)
		}
	}
	var names []string
	var list []json.RawMessage
	for name := range requests {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		data, err := protojson.Marshal(requests[name])
		if err != nil {
			t.Fatal(err)
		}
		list = append(list, data)
	}
	data, err := json.Marshal(list)
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "requests.json")
	if err := os.WriteFile(file, data, 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(node, "gear_twin.mjs", "../../assets/database/db.bin", file)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	var plans [][]string
	if err := json.Unmarshal(out, &plans); err != nil {
		t.Fatal(err)
	}
	db := database.Load()
	for i, name := range names {
		ours := gearLines(planLevelingGear(gearInputs(requests[name], db)))
		for l, line := range ours {
			if line != plans[i][l] {
				t.Errorf("%s:\n  Go %s\n  TS %s", name, line, plans[i][l])
				break
			}
		}
	}
}

func handsAt(plan *gearPlan, level int32) (mainHand, offHand *proto.UIItem) {
	gear := plan.levels[level-1].gear
	if p, ok := gear[proto.ItemSlot_ItemSlotMainHand]; ok {
		mainHand = p.item
	}
	if p, ok := gear[proto.ItemSlot_ItemSlotOffHand]; ok {
		offHand = p.item
	}
	return mainHand, offHand
}

func isShield(item *proto.UIItem) bool {
	return item != nil && item.WeaponType == proto.WeaponType_WeaponTypeShield
}

// Each style wears its own set from the level it can, whatever the weights say.
func TestWeaponStyles(t *testing.T) {
	db := database.Load()
	for seed := int32(1); seed <= 10; seed++ {
		plan := func(styles ...*proto.LevelingWeaponStyleRange) *gearPlan {
			return planLevelingGear(gearInputs(warriorGearRequest(t, seed, styles...), db))
		}
		twoHanded, dual, shield := plan(from(1, twoHand)), plan(from(1, dualWield)), plan(from(1, oneHandShield))
		for level := int32(1); level <= 60; level++ {
			if mh, oh := handsAt(shield, level); mh == nil || mh.HandType == proto.HandType_HandTypeTwoHand || !isShield(oh) {
				t.Fatalf("seed %d, shield at %d: %v and %v", seed, level, mh.GetName(), oh.GetName())
			}
			if level >= 5 {
				if mh, oh := handsAt(twoHanded, level); mh.GetHandType() != proto.HandType_HandTypeTwoHand || oh != nil {
					t.Fatalf("seed %d, two-hander at %d: %v and %v", seed, level, mh.GetName(), oh.GetName())
				}
			}
			if level >= 13 {
				mh, oh := handsAt(dual, level)
				if mh == nil || oh == nil || oh.WeaponSpeed == 0 || mh.HandType == proto.HandType_HandTypeTwoHand {
					t.Fatalf("seed %d, dual wield at %d: %v and %v", seed, level, mh.GetName(), oh.GetName())
				}
			}
			if _, oh := handsAt(dual, level); level < 13 && oh != nil && oh.WeaponSpeed > 0 {
				t.Fatalf("seed %d: a weapon in the off hand at %d, before Dual Wield", seed, level)
			}
		}
	}
}

// A plan that switches styles has the weapons it came across before the switch: at each level it
// holds exactly what a plan with that style all the way holds, and nothing else of it changes.
func TestWeaponStyleSwitchKeepsTheBag(t *testing.T) {
	db := database.Load()
	for seed := int32(1); seed <= 10; seed++ {
		plan := func(styles ...*proto.LevelingWeaponStyleRange) *gearPlan {
			req := warriorGearRequest(t, seed, styles...)
			// With the same weights for every style, the other slots are the same in every plan.
			req.GearWeights = slices.DeleteFunc(req.GearWeights, func(w *proto.LevelingGearWeights) bool { return w.Style != free })
			return planLevelingGear(gearInputs(req, db))
		}
		twoHanded, shield, both := plan(from(1, twoHand)), plan(from(1, oneHandShield)), plan(from(1, twoHand), from(25, oneHandShield))
		for level := int32(1); level <= 60; level++ {
			want := twoHanded
			if level >= 25 {
				want = shield
			}
			mh, oh := handsAt(both, level)
			wantMh, wantOh := handsAt(want, level)
			if level >= 5 {
				if mh.GetId() == 0 || (level >= 25) != isShield(oh) {
					t.Fatalf("seed %d, level %d: %v and %v", seed, level, mh.GetName(), oh.GetName())
				}
			}
			if level < 25 && (mh.GetId() != wantMh.GetId() || oh.GetId() != wantOh.GetId()) {
				t.Fatalf("seed %d, level %d: %v and %v, want %v and %v", seed, level, mh.GetName(), oh.GetName(), wantMh.GetName(), wantOh.GetName())
			}
		}
		// At the switch, the shield is the best of those the bag has, not the white one we started with.
		if _, oh := handsAt(both, 25); oh.Ilvl < 5 {
			t.Errorf("seed %d: still in the starting shield at 25", seed)
		}
	}
}

// Without weapon styles a warrior's plan is the old plan: what the weights like, with Dual Wield
// from level 1.
func TestNoWeaponStylesIsTheOldPlan(t *testing.T) {
	db := database.Load()
	req := warriorGearRequest(t, 1)
	plan := planLevelingGear(gearInputs(req, db))
	if mh, oh := handsAt(plan, 1); mh.GetHandType() != proto.HandType_HandTypeMainHand || oh != nil {
		t.Fatalf("level 1: %v and %v", mh, oh)
	}
}
