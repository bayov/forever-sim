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
// A round where no move stands out of the noise is scored again at refineIterations
// before the search stops, on a short list of the best moves by DPS and by end mana.
//
// When no move gains damage, mana is the tie-breaker. A move is kept when its damage is
// within the noise of the best build seen so far and it ends the fight with more mana by
// more than the noise. That is how a spare point lands in Elemental Focus or Convection
// instead of a talent that does nothing. Measuring against the best build seen so far
// (and not the current one) stops a run of tie-breaks from giving up damage bit by bit.
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

type talentCandidate struct {
	b     build
	score result
	gain  float64
	noise float64
}

// scoreBuilds scores each build against the current one, best DPS gain first.
func (s *searcher) scoreBuilds(builds []build, current result, knobs Knobs) []talentCandidate {
	var cands []talentCandidate
	for _, c := range builds {
		s.setup.talents = c.String()
		score := s.score(knobs)
		gain := score.dps - current.dps
		noise := math.Hypot(score.stderr, current.stderr)
		cands = append(cands, talentCandidate{c, score, gain, noise})
	}
	sort.Slice(cands, func(i, j int) bool { return cands[i].gain > cands[j].gain })
	return cands
}

// pickTalentMove is the candidate to move to, or nil. A DPS gain beyond the noise wins
// first. Otherwise, when tieBreak is set, the most end mana among the candidates whose
// DPS is within the noise of peak and whose mana beats the current build's by more than
// the noise. A build the search has been at before is never picked again.
func (s *searcher) pickTalentMove(cands []talentCandidate, current, peak result, tieBreak bool, visited map[string]bool) (*talentCandidate, bool) {
	var fresh []talentCandidate
	for _, c := range cands {
		if !visited[c.b.String()] {
			fresh = append(fresh, c)
		}
	}
	cands = fresh
	if len(cands) > 0 && cands[0].gain > s.confidence*cands[0].noise {
		return &cands[0], false
	}
	if !tieBreak {
		return nil, false
	}
	var tie *talentCandidate
	for i, c := range cands {
		dpsNoise := math.Hypot(c.score.stderr, peak.stderr)
		manaGain := c.score.mana - current.mana
		manaNoise := math.Hypot(c.score.manaStderr, current.manaStderr)
		if c.score.dps < peak.dps-s.confidence*dpsNoise || manaGain <= s.confidence*manaNoise {
			continue
		}
		if tie == nil || c.score.mana > tie.score.mana {
			tie = &cands[i]
		}
	}
	return tie, tie != nil
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
	peakBuild, peak := best, bestScore
	visited := map[string]bool{best.String(): true}

	// The mana tie-break only runs on scores at refineIterations. At search iterations
	// the DPS noise is wide enough that a tie-break would undo a move the refined
	// scores just made (Ancestral Knowledge 4 into Elemental Focus and back).
	canRefine := s.refineIterations > s.iterations

	for round := 1; ; round++ {
		cands := s.scoreBuilds(best.moves(trees), bestScore, knobs)
		for _, c := range cands[:min(8, len(cands))] {
			fmt.Printf("  %s: %s (%+.1f, noise %.1f)\n", best.diff(trees, c.b), c.score, c.gain, c.noise)
		}
		next, tieBreak := s.pickTalentMove(cands, bestScore, peak, !canRefine, visited)

		// Nothing stands out of the noise at search iterations. Before we stop, the
		// current build, the peak and the best few moves by DPS and by mana are scored
		// again at refineIterations, where a gain of 1 DPS is no longer lost in the noise.
		if next == nil && canRefine {
			short := map[string]build{}
			for _, c := range cands[:min(8, len(cands))] {
				short[c.b.String()] = c.b
			}
			byMana := append([]talentCandidate(nil), cands...)
			sort.Slice(byMana, func(i, j int) bool { return byMana[i].score.mana > byMana[j].score.mana })
			for _, c := range byMana[:min(8, len(byMana))] {
				short[c.b.String()] = c.b
			}
			var builds []build
			for _, b := range short {
				builds = append(builds, b)
			}

			searchIterations := s.iterations
			s.iterations = s.refineIterations
			s.setup.talents = best.String()
			bestScore = s.score(knobs)
			s.setup.talents = peakBuild.String()
			peak = s.score(knobs)
			cands = s.scoreBuilds(builds, bestScore, knobs)
			s.iterations = searchIterations

			fmt.Printf("  refined at %d iterations, current %s\n", s.refineIterations, bestScore)
			for _, c := range cands {
				fmt.Printf("  %s: %s (%+.1f, noise %.1f)\n", best.diff(trees, c.b), c.score, c.gain, c.noise)
			}
			next, tieBreak = s.pickTalentMove(cands, bestScore, peak, true, visited)
		}

		if next == nil {
			s.setup.talents = best.String()
			return best.String(), bestScore
		}
		how := ""
		if tieBreak {
			how = " (mana tie-break)"
		}
		fmt.Printf("round %d%s: %s, now %s %s\n", round, how, best.diff(trees, next.b), next.b, next.score)
		best, bestScore = next.b, next.score
		visited[best.String()] = true
		if bestScore.dps > peak.dps {
			peakBuild, peak = best, bestScore
		}
	}
}
