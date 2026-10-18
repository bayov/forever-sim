package aplkit

import (
	"fmt"
	"sort"
)

// A Knob is one number in a rotation the optimizer is free to move: an energy threshold,
// a minimum Slice and Dice duration, or a 0/1 switch for a whole rule.
type Knob struct {
	Name    string
	Default float64
	Min     float64
	Max     float64
	Step    float64
	// For the leveling plan's rotation search (tools/levelingplan), which tunes only some
	// knobs at each level. Needs are the abilities (a template's Abilities) the knob does
	// something with, and the search tunes it again at the level we get one. Kind says how
	// often else the search tunes it.
	Needs []string
	Kind  KnobKind
	// The knob's worth depends on the weapon imbue, so the search tries its values together
	// with each imbue too. Windfury Totem does nothing next to Windfury Weapon, so from
	// Windfury Weapon neither the totem alone nor Rockbiter alone would win.
	WithImbue bool
}

// How often the leveling plan's rotation search tunes a knob.
type KnobKind int

const (
	// A choice of what the rotation does, tuned where the rotation gains an ability and every
	// so many levels.
	KnobChoice KnobKind = iota
	// A health or mana threshold, tuned also at each level a spell we cast gets a new rank,
	// since the new rank costs more mana.
	KnobThreshold
	// A choice the level hardly changes, tuned at the plan's first level and every so many
	// levels.
	KnobSteady
)

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
