package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Talent search. Every move takes one point (or a whole talent's worth of points) from a
// talent that has some and puts it somewhere else the tree rules allow. The best move of a
// round is kept when it beats the current build by more than the noise, and that repeats
// until no move helps. The rotation is held fixed at its starting knobs, so this runs
// first and the knob search after it on the build it finds.
//
// The tree layout (rows, ranks, prerequisites, which talents the sim ignores) comes from
// the UI's tree data under ui/core/talents/trees, the same file the talent picker draws.

type talent struct {
	tree, index  int
	name         string
	maxPoints    int
	row          int
	prereq       int // index in the same tree, or -1
	notSimulated bool
	// The search keeps this talent at max rank. For talents the sim gives nothing for but
	// a player takes anyway, like the threat cut of the shaman's Spirit Weapons.
	required bool
}

type talentTree struct {
	name    string
	talents []talent
}

// Points the build may spend, one per level from 10 plus the player's bonus points. Set
// from the player in main.
var maxTalentPoints = 51

func loadTalentTrees(class string) ([]talentTree, error) {
	path := filepath.Join("ui", "core", "talents", "trees", class+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw []struct {
		Name    string `json:"name"`
		Talents []struct {
			FieldName    string `json:"fieldName"`
			MaxPoints    int    `json:"maxPoints"`
			NotSimulated bool   `json:"notSimulated"`
			Location     struct {
				Row int `json:"rowIdx"`
				Col int `json:"colIdx"`
			} `json:"location"`
			Prereq *struct {
				Row int `json:"rowIdx"`
				Col int `json:"colIdx"`
			} `json:"prereqLocation"`
		} `json:"talents"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	var trees []talentTree
	for ti, t := range raw {
		tree := talentTree{name: t.Name}
		for i, x := range t.Talents {
			tree.talents = append(tree.talents, talent{
				tree: ti, index: i, name: x.FieldName, maxPoints: x.MaxPoints,
				row: x.Location.Row, prereq: -1, notSimulated: x.NotSimulated,
			})
		}
		for i, x := range t.Talents {
			if x.Prereq == nil {
				continue
			}
			for j, y := range t.Talents {
				if y.Location.Row == x.Prereq.Row && y.Location.Col == x.Prereq.Col {
					tree.talents[i].prereq = j
				}
			}
		}
		trees = append(trees, tree)
	}
	return trees, nil
}

// A build is the points in every talent, laid out like the trees.
type build [][]int

func parseBuild(trees []talentTree, s string) build {
	b := make(build, len(trees))
	for ti, tree := range trees {
		b[ti] = make([]int, len(tree.talents))
	}
	for ti, treeStr := range strings.Split(s, "-") {
		for i, ch := range treeStr {
			if ti < len(b) && i < len(b[ti]) {
				b[ti][i] = int(ch - '0')
			}
		}
	}
	return b
}

// The string form drops trailing zeros in each tree and trailing empty trees, the way
// the UI writes it.
func (b build) String() string {
	var parts []string
	for _, tree := range b {
		s := ""
		for _, p := range tree {
			s += fmt.Sprint(p)
		}
		parts = append(parts, strings.TrimRight(s, "0"))
	}
	for len(parts) > 1 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return strings.Join(parts, "-")
}

func (b build) clone() build {
	c := make(build, len(b))
	for i, tree := range b {
		c[i] = append([]int(nil), tree...)
	}
	return c
}

func (b build) points() int {
	n := 0
	for _, tree := range b {
		for _, p := range tree {
			n += p
		}
	}
	return n
}

// requireTalents marks the named talents (field names, comma separated) as required and
// says which names matched nothing.
func requireTalents(trees []talentTree, names string) error {
	for _, name := range strings.Split(names, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		found := false
		for ti := range trees {
			for i := range trees[ti].talents {
				if strings.EqualFold(trees[ti].talents[i].name, name) {
					trees[ti].talents[i].required = true
					found = true
				}
			}
		}
		if !found {
			return fmt.Errorf("no talent named %q", name)
		}
	}
	return nil
}

// valid checks the tree rules: a row opens at 5 points per row above it in the same
// tree, a talent with a prerequisite needs that maxed, no talent over its rank count,
// and every required talent is maxed.
func (b build) valid(trees []talentTree) bool {
	if b.points() > maxTalentPoints {
		return false
	}
	for ti, tree := range trees {
		var rowPoints [10]int
		for i, t := range tree.talents {
			p := b[ti][i]
			if p < 0 || p > t.maxPoints {
				return false
			}
			if t.required && p < t.maxPoints {
				return false
			}
			if p > 0 && t.prereq >= 0 && b[ti][t.prereq] < tree.talents[t.prereq].maxPoints {
				return false
			}
			rowPoints[t.row] += p
		}
		above := 0
		for row, p := range rowPoints {
			if p > 0 && above < 5*row {
				return false
			}
			above += p
		}
	}
	return true
}

// moves lists every build one move away: a point or a whole talent moved from a talent
// that has points to another one, plus a point added when the build has room.
func (b build) moves(trees []talentTree) []build {
	var out []build
	seen := map[string]bool{b.String(): true}
	add := func(c build) {
		if c.valid(trees) && !seen[c.String()] {
			seen[c.String()] = true
			out = append(out, c)
		}
	}
	type slot struct{ tree, index int }
	var from, to []slot
	for ti, tree := range trees {
		for i, t := range tree.talents {
			if b[ti][i] > 0 {
				from = append(from, slot{ti, i})
			}
			if b[ti][i] < t.maxPoints {
				to = append(to, slot{ti, i})
			}
		}
	}
	if b.points() < maxTalentPoints {
		for _, d := range to {
			c := b.clone()
			c[d.tree][d.index]++
			add(c)
		}
	}
	for _, s := range from {
		for _, d := range to {
			if s == d {
				continue
			}
			c := b.clone()
			c[s.tree][s.index]--
			c[d.tree][d.index]++
			add(c)

			// The whole talent at once, for the cases where every single step loses
			// (Lethality into Seal Fate needs five points to move before it pays off).
			c = b.clone()
			n := min(c[s.tree][s.index], trees[d.tree].talents[d.index].maxPoints-c[d.tree][d.index])
			if n > 1 {
				c[s.tree][s.index] -= n
				c[d.tree][d.index] += n
				add(c)
			}
		}
	}
	return out
}

// Which talents differ between two builds, as "name -1 +2" pieces.
func (b build) diff(trees []talentTree, other build) string {
	var parts []string
	for ti, tree := range trees {
		for i, t := range tree.talents {
			if d := other[ti][i] - b[ti][i]; d != 0 {
				parts = append(parts, fmt.Sprintf("%s%+d", t.name, d))
			}
		}
	}
	return strings.Join(parts, " ")
}

func (s *searcher) searchTalents(trees []talentTree, start string, knobs Knobs) (string, result) {
	best := parseBuild(trees, start)
	if !best.valid(trees) {
		fmt.Fprintf(os.Stderr, "starting talents %s break the tree rules\n", start)
		os.Exit(2)
	}
	s.setup.talents = best.String()
	bestScore := s.score(knobs)
	fmt.Printf("start %s: %s\n", best, bestScore)

	for round := 1; ; round++ {
		type candidate struct {
			b     build
			score result
			gain  float64
			noise float64
		}
		var cands []candidate
		for _, c := range best.moves(trees) {
			s.setup.talents = c.String()
			score := s.score(knobs)
			gain := score.dps - bestScore.dps
			noise := math.Sqrt(score.stderr*score.stderr + bestScore.stderr*bestScore.stderr)
			cands = append(cands, candidate{c, score, gain, noise})
		}
		sort.Slice(cands, func(i, j int) bool { return cands[i].gain > cands[j].gain })
		for _, c := range cands[:min(8, len(cands))] {
			fmt.Printf("  %s: %s (%+.1f, noise %.1f)\n", best.diff(trees, c.b), c.score, c.gain, c.noise)
		}
		if len(cands) == 0 || cands[0].gain <= s.confidence*cands[0].noise {
			s.setup.talents = best.String()
			return best.String(), bestScore
		}
		fmt.Printf("round %d: %s, now %s %s\n", round, best.diff(trees, cands[0].b), cands[0].b, cands[0].score)
		best, bestScore = cands[0].b, cands[0].score
	}
}
