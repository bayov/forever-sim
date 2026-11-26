package aplkit

import (
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

func shamanPlayer() *proto.Player {
	return &proto.Player{Class: proto.Class_ClassShaman, Spec: &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{}}}
}

func warriorPlayer() *proto.Player {
	return &proto.Player{Class: proto.Class_ClassWarrior, Spec: &proto.Player_TankWarrior{TankWarrior: &proto.TankWarrior{}}}
}

// TestLevelingTemplateFor: a spec gets its default template, or the named one when it plays the
// spec, and only the shaman's has imbues.
func TestLevelingTemplateFor(t *testing.T) {
	for _, c := range []struct {
		player *proto.Player
		name   string
		// The template we want, or a part of the error.
		want, err string
	}{
		{shamanPlayer(), "", "shaman_enh_leveling", ""},
		{shamanPlayer(), "shaman_enh_leveling", "shaman_enh_leveling", ""},
		{warriorPlayer(), "", "warrior_leveling", ""},
		{warriorPlayer(), "shaman_enh_leveling", "", "want one of warrior_leveling"},
		{shamanPlayer(), "no_such", "", `No leveling rotation template "no_such"`},
		{&proto.Player{Class: proto.Class_ClassMage}, "", "", "no leveling rotation template for this Mage"},
	} {
		template, err := LevelingTemplateFor(c.player, c.name)
		if c.err != "" {
			if err == nil || !strings.Contains(err.Error(), c.err) {
				t.Errorf("%s %q: %v, want an error with %q", c.player.Class, c.name, err, c.err)
			}
			continue
		}
		if err != nil || template != LevelingTemplateNamed(c.want) {
			t.Errorf("%s %q: %T (%v), want %s", c.player.Class, c.name, template, err, c.want)
			continue
		}
		_, imbues := template.(ImbueTemplate)
		if want := c.want == "shaman_enh_leveling"; imbues != want {
			t.Errorf("%s: imbues %v, want %v", c.want, imbues, want)
		}
	}
	if names := LevelingTemplateNames(); !slices.Contains(names, "warrior_leveling") || !slices.Contains(names, "shaman_enh_leveling") {
		t.Errorf("templates %v", names)
	}
}

// TestWarriorLevelingAbilities: a warrior has Charge at 4, and its second rank at 26.
func TestWarriorLevelingAbilities(t *testing.T) {
	none := func(string) bool { return false }
	if got := (WarriorLeveling{Level: 3}).Abilities(none); slices.Contains(got, "charge") {
		t.Errorf("level 3 has %v", got)
	}
	if got := (WarriorLeveling{Level: 4}).Abilities(none); !slices.Contains(got, "charge") {
		t.Errorf("level 4 has %v", got)
	}
	if (WarriorLeveling{Level: 25}).NewRank() || !(WarriorLeveling{Level: 26}).NewRank() {
		t.Error("the second rank of Charge isn't at 26")
	}
}

// TestWarriorLevelingGoreDrinker: the lines that feed Gore Drinker (Reckless Swings, Bloodrage and
// Berserker Rage with its stance) cast only with the talent, and switched off there are none.
func TestWarriorLevelingGoreDrinker(t *testing.T) {
	knobs := Knobs{}
	for _, k := range (WarriorLeveling{}).Knobs() {
		knobs[k.Name] = k.Default
	}
	knobs["goreFeedBelow"], knobs["goreDance"] = 80, 1
	goreLines := func(k Knobs) []*proto.APLListItem {
		var lines []*proto.APLListItem
		for _, item := range (WarriorLeveling{Level: 50}).Build(k).PriorityList {
			if strings.Contains(item.Notes, "Gore Drinker") || strings.Contains(item.Notes, "for Berserker Rage") {
				lines = append(lines, item)
			}
		}
		return lines
	}
	lines := goreLines(knobs)
	if len(lines) != 4 {
		t.Fatalf("%d Gore Drinker lines at 50 with goreDance, want Reckless Swings, Bloodrage, Berserker Stance and Berserker Rage", len(lines))
	}
	for _, item := range lines {
		condition, err := protojson.Marshal(item.Action.Condition)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(condition), `"talent":"goreDrinker"`) {
			t.Errorf("%q casts without the talent: %s", item.Notes, condition)
		}
	}
	knobs["goreFeedBelow"] = 0
	if lines := goreLines(knobs); len(lines) != 0 {
		t.Errorf("%d Gore Drinker lines with goreFeedBelow 0, want none", len(lines))
	}
}
