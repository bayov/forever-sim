package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/tools/rotopt/aplkit"
	"google.golang.org/protobuf/encoding/protojson"
	goproto "google.golang.org/protobuf/proto"
)

// A planConfig is a short JSON file -make turns into a LevelingPlanSearchRequest, so we can plan a
// class without the browser (docs/leveling_plan.md, A plan without the browser). Paths in it go
// from the file's own folder.
type planConfig struct {
	// Player fields as protojson: the race, class, spec with its options, and professions. The
	// level, talents, gear and rotation come from the plan.
	Player json.RawMessage `json:"player"`
	// LevelingOptions as protojson, for the mobs, searches and deaths. The sim's defaults without.
	Leveling json.RawMessage `json:"leveling"`

	FromLevel         int32 `json:"fromLevel"`
	ToLevel           int32 `json:"toLevel"`
	FirstTalentLevel  int32 `json:"firstTalentLevel"`
	BonusTalentPoints int32 `json:"bonusTalentPoints"`

	Styles []styleConfig `json:"styles"`
	// The talent builds. Each one after the first is a respec at its level.
	Talents   []talentConfig   `json:"talents"`
	Required  []requiredConfig `json:"requiredTalents"`
	Rotations []rotationConfig `json:"rotations"`
	// The template the rotation search tunes. Empty for the spec's default.
	RotationTemplate string `json:"rotationTemplate"`
	// What we start from, and the fields above replace: a plan as the UI exports it
	// (LevelingPlanSettings), a preset file (ui/enhancement_shaman/apls/leveling_plan_audited.json)
	// or a search's result, whose route and rotations we take.
	PlanFile string `json:"planFile"`
	// A LevelingTrainingSettings file (-training writes one).
	TrainingFile string `json:"trainingFile"`
	RouteFile    string `json:"routeFile"`
	// The stat weights files by style: "free" for every style without its own, "twoHand",
	// "dualWield" and "shield".
	Weights map[string]string `json:"weights"`

	GearSeed int32           `json:"gearSeed"`
	Pins     json.RawMessage `json:"pins"`
	Seed     int64           `json:"seed"`
	// The search's sessions and respec levels.
	Sessions        int32   `json:"sessions"`
	ConfirmSessions int32   `json:"confirmSessions"`
	RespecLevels    []int32 `json:"respecLevels"`
	RotationSearch  bool    `json:"rotationSearch"`
}

type styleConfig struct {
	Level int32  `json:"level"`
	Style string `json:"style"`
}

type talentConfig struct {
	Level int32 `json:"level"`
	// The talents in the order we put their points in, by name with how many points in a row
	// ("Cruelty:5", or "Cruelty" for one).
	Order []string `json:"order"`
}

type requiredConfig struct {
	Talent string `json:"talent"`
	Points int32  `json:"points"`
	Level  int32  `json:"level"`
}

// The rotation from a level until the next one's: a template with the knobs that aren't at their
// default, or a file with an APL rotation.
type rotationConfig struct {
	Level    int32        `json:"level"`
	Template string       `json:"template"`
	Knobs    aplkit.Knobs `json:"knobs"`
	File     string       `json:"file"`
	Imbue    string       `json:"imbue"`
}

var styleNames = map[string]proto.LevelingWeaponStyle{
	"free":      proto.LevelingWeaponStyle_LevelingWeaponStyleFree,
	"twoHand":   proto.LevelingWeaponStyle_LevelingWeaponStyleTwoHand,
	"dualWield": proto.LevelingWeaponStyle_LevelingWeaponStyleDualWield,
	"shield":    proto.LevelingWeaponStyle_LevelingWeaponStyleOneHandShield,
}

func styleNamed(name string) (proto.LevelingWeaponStyle, error) {
	style, ok := styleNames[name]
	if !ok {
		return 0, fmt.Errorf("no weapon style %q, want free, twoHand, dualWield or shield", name)
	}
	return style, nil
}

func styleName(style proto.LevelingWeaponStyle) string {
	for name, s := range styleNames {
		if s == style {
			return name
		}
	}
	return ""
}

// makeRequestFile writes the request of a config file.
func makeRequestFile(config, out string) error {
	req, err := requestFromConfig(config)
	if err != nil {
		return err
	}
	data, err := protojson.MarshalOptions{Multiline: true}.Marshal(req)
	if err != nil {
		return err
	}
	if out == "" {
		_, err = os.Stdout.Write(append(data, '\n'))
		return err
	}
	return os.WriteFile(out, data, 0o644)
}

// requestFromConfig is the request the UI would make for the config's plan.
func requestFromConfig(path string) (*proto.LevelingPlanSearchRequest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config := &planConfig{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	// A field we don't know is most likely a typo, and the plan would silently differ.
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(config); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	req, err := config.request(filepath.Dir(path))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return req, nil
}

func (c *planConfig) request(dir string) (*proto.LevelingPlanSearchRequest, error) {
	resolve := func(path string) string {
		if filepath.IsAbs(path) {
			return path
		}
		return filepath.Join(dir, path)
	}

	player := &proto.Player{
		ReactionTimeMs: 200, ChannelClipDelayMs: 50, DistanceFromTarget: 5, InFrontOfTarget: true,
		Consumes: &proto.Consumes{}, Buffs: &proto.IndividualBuffs{}, Equipment: &proto.EquipmentSpec{},
		Rotation: &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
	}
	if err := mergeJSON(c.Player, player, "player"); err != nil {
		return nil, err
	}
	if player.Class == proto.Class_ClassUnknown || player.Race == proto.Race_RaceUnknown || player.Spec == nil {
		return nil, fmt.Errorf(`player needs a race, a class and a spec, like "tankWarrior": {"options": {}}`)
	}
	leveling := &proto.LevelingOptions{}
	if err := mergeJSON(c.Leveling, leveling, "leveling"); err != nil {
		return nil, err
	}
	raid := core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
	raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player}}

	shape, err := classTreeShape(player.Class)
	if err != nil {
		return nil, err
	}
	rules, err := classRules(player.Class)
	if err != nil {
		return nil, err
	}
	req := &proto.LevelingPlanSearchRequest{
		Base: &proto.RaidSimRequest{
			Raid:       raid,
			Encounter:  &proto.Encounter{Leveling: leveling},
			SimOptions: &proto.SimOptions{Iterations: 1000, RandomSeed: c.Seed},
		},
		GearPlan:           &proto.LevelingGearPlanSettings{Seed: c.GearSeed},
		GearDefaultWeights: &proto.UnitStats{},
		ClassRules:         rules,
		Faction:            factionOf(player.Race),
		TalentShape:        shape,
		RespecLevels:       c.RespecLevels,
		Sessions:           c.Sessions,
		ConfirmSessions:    c.ConfirmSessions,
		RotationSearch:     c.RotationSearch,
		RotationTemplate:   c.RotationTemplate,
	}
	if req.Base.SimOptions.RandomSeed == 0 {
		req.Base.SimOptions.RandomSeed = 1
	}
	if req.GearPlan.Seed == 0 {
		req.GearPlan.Seed = 1
	}
	if req.Sessions == 0 {
		req.Sessions = 100
	}
	if req.ConfirmSessions == 0 {
		req.ConfirmSessions = 2000
	}
	if len(c.Pins) > 0 {
		pins := &proto.LevelingGearPlanSettings{}
		if err := mergeJSON(json.RawMessage(`{"pins":`+string(c.Pins)+`}`), pins, "pins"); err != nil {
			return nil, err
		}
		req.GearPlan.Pins = pins.Pins
	}
	if c.RouteFile != "" {
		req.Route = &proto.LevelingRoute{}
		if err := readProto(resolve(c.RouteFile), req.Route); err != nil {
			return nil, err
		}
	}
	if _, err := aplkit.LevelingTemplateFor(player, c.RotationTemplate); err != nil && c.RotationTemplate != "" {
		return nil, err
	}

	if req.Plan, err = c.plan(player, shape, resolve); err != nil {
		return nil, err
	}
	if err := routeProblem(req.Plan, shape); err != nil {
		return nil, err
	}
	// The presets go by style and then by level. Within a style the order is the file's.
	names := []string{"free", "twoHand", "dualWield", "shield"}
	for name := range c.Weights {
		if !slices.Contains(names, name) {
			_, err := styleNamed(name)
			return nil, fmt.Errorf("weights: %w", err)
		}
	}
	for _, name := range names {
		path, ok := c.Weights[name]
		if !ok {
			continue
		}
		file, err := readWeightsFile(resolve(path))
		if err != nil {
			return nil, err
		}
		presets, err := file.presets(styleNames[name])
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		req.GearWeights = append(req.GearWeights, presets...)
	}
	return req, nil
}

// plan is the config's plan, on top of its plan file.
func (c *planConfig) plan(player *proto.Player, shape *proto.LevelingTalentShape, resolve func(string) string) (*proto.LevelingPlanSettings, error) {
	plan := &proto.LevelingPlanSettings{}
	if c.PlanFile != "" {
		base, err := readPlanFile(resolve(c.PlanFile))
		if err != nil {
			return nil, err
		}
		plan = base
	}
	if c.FromLevel > 0 {
		plan.FromLevel = c.FromLevel
	}
	if c.ToLevel > 0 {
		plan.ToLevel = c.ToLevel
	}
	if plan.FromLevel == 0 {
		plan.FromLevel = 1
	}
	if plan.ToLevel == 0 {
		plan.ToLevel = maxPlanLevel
	}
	if c.FirstTalentLevel > 0 {
		plan.FirstTalentLevel = c.FirstTalentLevel
	}
	if c.BonusTalentPoints > 0 {
		plan.BonusTalentPoints = c.BonusTalentPoints
	}
	plan.RotationsNameTopRanks = true

	if len(c.Styles) > 0 {
		plan.WeaponStyles = nil
		plan.OneHandAndShield = false
	}
	for _, s := range c.Styles {
		style, err := styleNamed(s.Style)
		if err != nil {
			return nil, err
		}
		plan.WeaponStyles = append(plan.WeaponStyles, &proto.LevelingWeaponStyleRange{Level: s.Level, Style: style})
	}
	if len(c.Talents) > 0 {
		plan.Builds = nil
	}
	for _, t := range c.Talents {
		picks, err := picksFromNames(t.Order, shape)
		if err != nil {
			return nil, err
		}
		plan.Builds = append(plan.Builds, &proto.LevelingTalentBuild{Level: max(t.Level, plan.FromLevel), Picks: picks})
	}
	if len(c.Required) > 0 {
		plan.RequiredTalents = nil
	}
	for _, r := range c.Required {
		at, err := talentNamed(r.Talent, shape)
		if err != nil {
			return nil, err
		}
		plan.RequiredTalents = append(plan.RequiredTalents, &proto.LevelingTalentRequirement{
			Tree: int32(at.tree), Talent: int32(at.talent), Points: r.Points, Level: r.Level,
		})
	}
	if c.TrainingFile != "" {
		plan.Training = &proto.LevelingTrainingSettings{}
		if err := readProto(resolve(c.TrainingFile), plan.Training); err != nil {
			return nil, err
		}
	}
	if len(c.Rotations) > 0 {
		plan.Rotations, plan.RotationLibrary = nil, nil
	}
	rotations := slices.Clone(c.Rotations)
	slices.SortStableFunc(rotations, func(a, b rotationConfig) int { return int(a.Level - b.Level) })
	for i, r := range rotations {
		from, to := r.Level, plan.ToLevel
		if i == 0 {
			from = plan.FromLevel
		}
		if i+1 < len(rotations) {
			to = rotations[i+1].Level
		}
		if err := addRotation(plan, r, player, from, to, resolve); err != nil {
			return nil, err
		}
	}
	return plan, nil
}

// addRotation adds a rotation's level ranges to the plan.
//
// A template's rotation has only the abilities we know at the level it's built for, so a
// template gets a new range at each level its rotation changes, like where we learn a new one.
func addRotation(plan *proto.LevelingPlanSettings, r rotationConfig, player *proto.Player, from, to int32, resolve func(string) string) error {
	imbue := proto.WeaponImbue_WeaponImbueUnknown
	if r.Imbue != "" {
		value, ok := proto.WeaponImbue_value[r.Imbue]
		if !ok {
			return fmt.Errorf("no weapon imbue %q", r.Imbue)
		}
		imbue = proto.WeaponImbue(value)
	}
	add := func(level int32, rotation *proto.APLRotation) {
		name := fmt.Sprintf("From %d", level)
		plan.RotationLibrary = append(plan.RotationLibrary, &proto.LevelingRotationPreset{Name: name, Rotation: rotation})
		plan.Rotations = append(plan.Rotations, &proto.LevelingRotation{Level: level, Name: name, Imbue: imbue})
	}
	if r.File != "" {
		rotation := &proto.APLRotation{}
		if err := readProto(resolve(r.File), rotation); err != nil {
			return err
		}
		add(from, rotation)
		return nil
	}
	template, err := aplkit.LevelingTemplateFor(player, r.Template)
	if err != nil {
		return err
	}
	knobs := aplkit.Knobs{}
	for _, knob := range template.Knobs() {
		knobs[knob.Name] = knob.Default
	}
	for name, value := range r.Knobs {
		if _, ok := knobs[name]; !ok {
			return fmt.Errorf("the rotation template has no knob %q", name)
		}
		knobs[name] = value
	}
	var last *proto.APLRotation
	for level := from; level < max(to, from+1); level++ {
		rotation := template.Build(level, knobs)
		if last == nil || !goproto.Equal(last, rotation) {
			add(level, rotation)
		}
		last = rotation
	}
	return nil
}

// readPlanFile reads a plan in any of the shapes we have one: LevelingPlanSettings, the plan preset
// files (with a route in place of the builds), or a LevelingPlanSearchResult.
func readPlanFile(path string) (*proto.LevelingPlanSettings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	plan := &proto.LevelingPlanSettings{}
	lenient := protojson.UnmarshalOptions{DiscardUnknown: true}
	if err := lenient.Unmarshal(data, plan); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	result := &proto.LevelingPlanSearchResult{}
	if err := lenient.Unmarshal(data, result); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if len(plan.Builds) == 0 && len(result.Route) > 0 {
		plan.Builds = []*proto.LevelingTalentBuild{{Level: plan.FromLevel, Picks: result.Route}}
	}
	if len(plan.Builds) == 0 && len(plan.Rotations) == 0 {
		return nil, fmt.Errorf("%s: no talents and no rotations in it", path)
	}
	return plan, nil
}

// picksFromNames are the points of names like "Cruelty:5", in order.
func picksFromNames(order []string, shape *proto.LevelingTalentShape) ([]*proto.LevelingTalentPick, error) {
	var picks []*proto.LevelingTalentPick
	for _, entry := range order {
		name, count, hasCount := strings.Cut(entry, ":")
		points := 1
		if hasCount {
			n, err := strconv.Atoi(strings.TrimSpace(count))
			if err != nil || n < 1 {
				return nil, fmt.Errorf("talent %q: want a name with a number of points, like \"Cruelty:5\"", entry)
			}
			points = n
		}
		at, err := talentNamed(name, shape)
		if err != nil {
			return nil, err
		}
		for range points {
			picks = append(picks, &proto.LevelingTalentPick{Tree: int32(at.tree), Talent: int32(at.talent)})
		}
	}
	return picks, nil
}

// talentNamed finds a talent by its name in any case. Where two trees have one, the name goes with
// its tree, like "Fury/Cruelty".
func talentNamed(name string, shape *proto.LevelingTalentShape) (pick, error) {
	name = strings.TrimSpace(name)
	var found []pick
	for t, tree := range shape.Trees {
		for i, talent := range tree.Talents {
			if strings.EqualFold(talent.Name, name) || strings.EqualFold(tree.Name+"/"+talent.Name, name) {
				found = append(found, pick{t, i})
			}
		}
	}
	switch len(found) {
	case 1:
		return found[0], nil
	case 0:
		return pick{}, fmt.Errorf("no talent %q in the trees", name)
	}
	return pick{}, fmt.Errorf("%d trees have a talent %q, write it with its tree, like %q", len(found), name, shape.Trees[found[0].tree].Name+"/"+name)
}

// classTreeShape is the class's talent trees, from the UI's file.
func classTreeShape(class proto.Class) (*proto.LevelingTalentShape, error) {
	name := strings.ToLower(strings.TrimPrefix(class.String(), "Class"))
	path := findTreeFile(name)
	if path == "" {
		return nil, fmt.Errorf("no talent trees for %s, run it inside the repo", name)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return treeFileShape(data)
}

// factionOf is a race's side (raceToFaction in proto_utils/utils.ts).
func factionOf(race proto.Race) proto.Faction {
	switch race {
	case proto.Race_RaceOrc, proto.Race_RaceTauren, proto.Race_RaceTroll, proto.Race_RaceUndead, proto.Race_RaceSkyborneWindshaper:
		return proto.Faction_Horde
	case proto.Race_RaceUnknown:
		return proto.Faction_Unknown
	}
	return proto.Faction_Alliance
}

func mergeJSON(data json.RawMessage, into goproto.Message, field string) error {
	if len(data) == 0 {
		return nil
	}
	fresh := into.ProtoReflect().New().Interface()
	if err := protojson.Unmarshal(data, fresh); err != nil {
		return fmt.Errorf("%s: %w", field, err)
	}
	goproto.Merge(into, fresh)
	return nil
}

// readProto reads a file's fields. It skips the ones the message doesn't have, since our data
// files have notes.
func readProto(path string, into goproto.Message) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(data, into); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}
