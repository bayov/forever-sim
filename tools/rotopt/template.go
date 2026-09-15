package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

// A Knob is one number in a rotation the optimizer is free to move: an energy threshold,
// a minimum Slice and Dice duration, or a 0/1 switch for a whole rule.
type Knob struct {
	Name    string
	Default float64
	Min     float64
	Max     float64
	Step    float64
}

// Every candidate value of the knob, from Min to Max.
func (k Knob) Values() []float64 {
	var vals []float64
	for v := k.Min; v <= k.Max+1e-9; v += k.Step {
		vals = append(vals, v)
	}
	return vals
}

type Knobs map[string]float64

func (k Knobs) String() string {
	names := make([]string, 0, len(k))
	for name := range k {
		names = append(names, name)
	}
	sort.Strings(names)
	s := ""
	for i, name := range names {
		if i > 0 {
			s += " "
		}
		s += fmt.Sprintf("%s=%g", name, k[name])
	}
	return s
}

// A Template turns knob values into a full rotation. The rules are fixed by the
// template, the optimizer only picks the numbers.
type Template interface {
	Knobs() []Knob
	Build(knobs Knobs) *proto.APLRotation
}

var templates = map[string]Template{
	"rogue_ss": rogueSS{},
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
