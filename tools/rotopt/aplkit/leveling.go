package aplkit

import (
	"fmt"
	"slices"
	"strings"

	"github.com/wowsims/classic/sim/core/proto"
)

// A LevelingTemplate is a leveling rotation with knobs, for the leveling plan's rotation search
// (tools/levelingplan) and rotopt. It builds the rotation of a level, and says which abilities we
// know there, since the search tunes the rotation again from each level that adds one.
type LevelingTemplate interface {
	Knobs() []Knob
	Build(level int32, k Knobs) *proto.APLRotation
	// What the rotation can use at the level, with the talents hasTalent says we have. It takes
	// a talent's name without spaces, in any case.
	Abilities(level int32, hasTalent func(name string) bool) []string
	// Whether a spell the rotation casts gets a new rank at the level, past its first.
	NewRank(level int32) bool
}

// An ImbueTemplate also has weapon imbues to pick from, which are options of the spec and not
// something the rotation casts. Only the shaman has them.
type ImbueTemplate interface {
	Imbues() []Imbue
}

// An Imbue goes with the ability of that name in Abilities.
type Imbue struct {
	Ability string
	Imbue   proto.WeaponImbue
}

type levelingEntry struct {
	name     string
	fits     func(player *proto.Player) bool
	template LevelingTemplate
}

// Every leveling template, in the order we look for one that fits a player.
var levelingTemplates []levelingEntry

// RegisterLevelingTemplate adds a template by name, from an init in the template's file. fits says
// which specs it plays, and the first one registered that fits is a spec's default.
func RegisterLevelingTemplate(name string, template LevelingTemplate, fits func(player *proto.Player) bool) {
	if slices.ContainsFunc(levelingTemplates, func(e levelingEntry) bool { return e.name == name }) {
		panic("two leveling templates named " + name)
	}
	levelingTemplates = append(levelingTemplates, levelingEntry{name, fits, template})
}

// LevelingTemplateNamed is the template of a name, nil for none.
func LevelingTemplateNamed(name string) LevelingTemplate {
	for _, e := range levelingTemplates {
		if e.name == name {
			return e.template
		}
	}
	return nil
}

// LevelingTemplateFor is the template of a name if it plays the player's spec, or the spec's
// default one without a name.
func LevelingTemplateFor(player *proto.Player, name string) (LevelingTemplate, error) {
	var fitting []string
	for _, e := range levelingTemplates {
		if !e.fits(player) {
			continue
		}
		if name == "" || e.name == name {
			return e.template, nil
		}
		fitting = append(fitting, e.name)
	}
	class := strings.TrimPrefix(player.GetClass().String(), "Class")
	switch {
	case len(fitting) > 0:
		return nil, fmt.Errorf("No leveling rotation template %q for this %s, want one of %s.", name, class, strings.Join(fitting, ", "))
	case name != "":
		return nil, fmt.Errorf("The leveling rotation template %q isn't for this %s.", name, class)
	}
	return nil, fmt.Errorf("The rotation search has no leveling rotation template for this %s.", class)
}

// LevelingTemplateNames are the names of every template.
func LevelingTemplateNames() []string {
	var out []string
	for _, e := range levelingTemplates {
		out = append(out, e.name)
	}
	return out
}
