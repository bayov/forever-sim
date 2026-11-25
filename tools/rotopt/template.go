package main

import (
	"os"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/tools/rotopt/aplkit"
	"google.golang.org/protobuf/encoding/protojson"
)

// The knobs live in aplkit, with the leveling templates.
type Knob = aplkit.Knob
type Knobs = aplkit.Knobs

// A Template turns knob values into a full rotation. The rules are fixed by the
// template, the optimizer only picks the numbers.
type Template interface {
	Knobs() []Knob
	Build(knobs Knobs) *proto.APLRotation
}

var templates = map[string]Template{
	"rogue_ss":     rogueSS{},
	"rogue_bs":     rogueBS{},
	"shaman_enh":   shamanEnh{},
	"paladin_ret":  paladinRet{},
	"paladin_tank": paladinTank{},
}

// The leveling templates, for leveling sessions (settings with leveling options). They're in
// aplkit because the leveling plan's rotation search (tools/levelingplan) uses them too.
func init() {
	for _, name := range aplkit.LevelingTemplateNames() {
		templates[name] = levelingTemplate{aplkit.LevelingTemplateNamed(name)}
	}
}

// A leveling template, built for the -level we sim.
type levelingTemplate struct {
	template aplkit.LevelingTemplate
}

func (t levelingTemplate) Knobs() []Knob { return t.template.Knobs() }

func (t levelingTemplate) Build(k Knobs) *proto.APLRotation {
	return t.template.Build(playerLevel, k)
}

func defaultKnobs(t Template) Knobs {
	knobs := Knobs{}
	for _, k := range t.Knobs() {
		knobs[k.Name] = k.Default
	}
	return knobs
}

// fixedRotation is a hand written rotation file simmed as is, so a preset can be
// compared against the templates with the same settings and seed.
type fixedRotation struct {
	rot *proto.APLRotation
}

func (fixedRotation) Knobs() []Knob                    { return nil }
func (f fixedRotation) Build(Knobs) *proto.APLRotation { return f.rot }

func loadRotation(path string) (*proto.APLRotation, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	rot := &proto.APLRotation{}
	if err := protojson.Unmarshal(data, rot); err != nil {
		return nil, err
	}
	return rot, nil
}
