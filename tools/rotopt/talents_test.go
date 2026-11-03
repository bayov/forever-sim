package main

import (
	"os"
	"testing"
)

// shamanTrees loads the shaman trees from the repo root, where rotopt runs.
func shamanTrees(t *testing.T) []talentTree {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir("../.."); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(wd) })
	trees, err := loadTalentTrees("shaman")
	if err != nil {
		t.Fatal(err)
	}
	return trees
}

// TestTalentPoints checks how many points the search may spend (shaman_audit.md 8.51).
//
// A character has one point a level from level 10, and the bonus points come on top. The
// beta's level 30 exports all have 26 points, which is 21 plus 5. The bonus points never
// take a character past 51.
func TestTalentPoints(t *testing.T) {
	for _, c := range []struct {
		level, bonus int32
		want         int
	}{{30, 0, 21}, {30, 5, 26}, {25, 5, 21}, {9, 0, 0}, {60, 0, 51}, {60, 5, 51}, {0, 5, 51}} {
		if got := talentPoints(c.level, c.bonus); got != c.want {
			t.Errorf("level %d with %d bonus points: %d points, want %d", c.level, c.bonus, got, c.want)
		}
	}
}

// TestTalentTreeRules checks the tree rules the search keeps to (shaman_audit.md 8.51).
//
// A row opens at 5 points per row above it in the same tree, a talent with a prerequisite
// needs that one maxed, and a required talent must be maxed. Each broken build is a legal 26
// point build (the Level 30 + 5 preset until the November 2 notes moved Improved Ghost Wolf to
// row 4, which left it 13 points above row 4) with one change, so it breaks only the rule it names.
func TestTalentTreeRules(t *testing.T) {
	trees := shamanTrees(t)
	defer func(n int) { maxTalentPoints = n }(maxTalentPoints)
	maxTalentPoints = 26

	const preset = "-0550301100510122"
	check := func(name, s string, want bool) {
		t.Helper()
		if got := parseBuild(trees, s).valid(trees); got != want {
			t.Errorf("%s (%s): valid %v, want %v", name, s, got, want)
		}
	}
	check("the 26 point build", preset, true)
	check("a 27th point in Toughness", "-0550301101510122", false)
	check("row 4 with 14 points above it", "-0550301000510122", false)
	check("Improved Stormstrike without Stormstrike", "-0550301100500122", false)
	check("Flurry with Mental Dexterity at 2/3", "-0551201100510122", false)
	check("no Spirit Weapons", "-0550301100510022", true)

	maxTalentPoints = 27
	check("a 27th point in Toughness with 27 points", "-0550301101510122", true)

	if err := requireTalents(trees, "spiritWeapons"); err != nil {
		t.Fatal(err)
	}
	check("no Spirit Weapons when it's required", "-0550301100510022", false)
	check("the preset when Spirit Weapons is required", preset, true)
}
