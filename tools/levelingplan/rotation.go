package main

import (
	"context"
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/tools/rotopt/aplkit"
	goproto "google.golang.org/protobuf/proto"
)

// The rotation search (docs/leveling_plan.md, Rotation search). It splits the plan's levels
// into ranges at each level that adds an ability the rotation template can use (a spell, a
// talent like Stormstrike, an imbue) or a new rank of a spell it casts, and tunes the
// template's knobs for each range, starting from the range before. A range only tunes the
// knobs that what's new there can change:
//   - At the plan's first level, and at a new ability anchorEvery levels after we last did,
//     every knob.
//   - At a new ability, the rest and mana thresholds, the imbue and the knobs of that ability.
//   - At a new rank only, the thresholds, since the rank costs more mana.
//
// Each change is scored by the hours of up to bandSampleLevels levels of the range (its first
// level for a new rank). The candidates of a round all start at the first halvingSessions,
// and after each step the best quarter goes on to the next, so only the close ones get the
// sessions. A threshold only tries the values up to thresholdReach steps from where it is.
const (
	bandSampleLevels = 3
	thresholdReach   = 2
	// A change must save this many times its noise at the last step. One of 70 changes that
	// only beats its noise once is mostly luck.
	winnerNoise = 1
	anchorEvery = 15
	// The most rounds of changes we make to one range's knobs.
	fullRounds    = 3
	touchUpRounds = 1
	// The most times we search the rotations and then the route, in turns.
	maxRotationPasses = 3
)

var halvingSessions = []int32{50, 200, 800}

// A rotation template the plan search can tune (tools/rotopt/aplkit).
type levelTemplate interface {
	Knobs() []aplkit.Knob
	Build(level int32, k aplkit.Knobs) *proto.APLRotation
	Abilities(level int32, hasTalent func(name string) bool) []string
	NewRank(level int32) bool
}

type shamanEnhTemplate struct{}

func (shamanEnhTemplate) Knobs() []aplkit.Knob { return aplkit.ShamanEnhLeveling{}.Knobs() }

func (shamanEnhTemplate) Build(level int32, k aplkit.Knobs) *proto.APLRotation {
	return aplkit.ShamanEnhLeveling{Level: level}.Build(k)
}

func (shamanEnhTemplate) Abilities(level int32, hasTalent func(string) bool) []string {
	return aplkit.ShamanEnhLeveling{Level: level}.Abilities(hasTalent)
}

func (shamanEnhTemplate) NewRank(level int32) bool {
	return aplkit.ShamanEnhLeveling{Level: level}.NewRank()
}

func templateFor(player *proto.Player) (levelTemplate, error) {
	if _, ok := player.GetSpec().(*proto.Player_EnhancementShaman); ok {
		return shamanEnhTemplate{}, nil
	}
	return nil, fmt.Errorf("The rotation search has a leveling rotation template only for Enhancement shamans.")
}

// The imbues an ability name stands for, in the order we list them.
var imbueAbilities = []struct {
	ability string
	imbue   proto.WeaponImbue
}{
	{"rockbiterWeapon", proto.WeaponImbue_RockbiterWeapon},
	{"flametongueWeapon", proto.WeaponImbue_FlametongueWeapon},
	{"windfuryWeapon", proto.WeaponImbue_WindfuryWeapon},
}

// One level range of the rotation search.
type rotationBand struct {
	level  int32
	levels []int32
	// The abilities new at its level, none for a range where only a rank is new.
	added []string
	// Whether it tunes every knob: the plan's first range, and the anchors.
	every     bool
	abilities []string
	// The talents of its levels, to tell whether a later pass changed them.
	talents string
	knobs   aplkit.Knobs
	imbue   proto.WeaponImbue
	sims    int
}

func (b *rotationBand) kind() string {
	switch {
	case b.every:
		return "every knob"
	case len(b.added) > 0:
		return "new " + strings.Join(b.added, ", ")
	}
	return "new ranks"
}

type rotationSearch struct {
	p        *planner
	ctx      context.Context
	template levelTemplate
	shape    *proto.LevelingTalentShape
	levels   []int32
	progress func(string)
	// Each elimination step plays new sessions, so a winner isn't the luck of the step
	// before.
	step int64
}

// A candidate: knob values and an imbue.
type rotationCandidate struct {
	knobs aplkit.Knobs
	imbue proto.WeaponImbue
	// The knob it changes from the current, "imbue", or "imbue+" and a knob, for combining
	// winners.
	changed string
}

// planBands splits the levels into ranges, with the talents talentsAt has at each level.
func (rs *rotationSearch) planBands(talentsAt func(level int32) string) []*rotationBand {
	var bands []*rotationBand
	var prev []string
	lastEvery := int32(math.MinInt32 / 2)
	for i, level := range rs.levels {
		talents := talentsAt(level)
		abilities := rs.template.Abilities(level, rs.hasTalent(talents))
		var added []string
		for _, a := range abilities {
			if !slices.Contains(prev, a) {
				added = append(added, a)
			}
		}
		prev = abilities
		var band *rotationBand
		switch {
		case i == 0:
			band = &rotationBand{every: true}
			lastEvery = level
		case len(added) > 0:
			band = &rotationBand{added: added, every: level-lastEvery >= anchorEvery}
			if band.every {
				lastEvery = level
			}
		case rs.template.NewRank(level):
			band = &rotationBand{}
		}
		if band != nil {
			band.level = level
			band.abilities = abilities
			bands = append(bands, band)
		}
		last := bands[len(bands)-1]
		last.levels = append(last.levels, level)
		last.talents += talents + ";"
	}
	return bands
}

// Whether the talents have a point in the talent, by the template's name for it: the trees'
// names without spaces and in any case, like maelstromWeapon for Maelstrom Weapon.
func (rs *rotationSearch) hasTalent(talents string) func(string) bool {
	counts := parseTalents(talents, rs.shape)
	return func(name string) bool {
		for t, tree := range rs.shape.Trees {
			for i, talent := range tree.Talents {
				if strings.EqualFold(strings.ReplaceAll(talent.Name, " ", ""), name) && counts[t][i] > 0 {
					return true
				}
			}
		}
		return false
	}
}

// The knobs a range tunes, and whether it tunes the imbue.
func (rs *rotationSearch) knobsToTune(b *rotationBand) ([]aplkit.Knob, bool) {
	var out []aplkit.Knob
	for _, knob := range rs.template.Knobs() {
		if len(knob.Needs) > 0 && !slices.ContainsFunc(knob.Needs, func(a string) bool { return slices.Contains(b.abilities, a) }) {
			continue
		}
		newNeed := slices.ContainsFunc(knob.Needs, func(a string) bool { return slices.Contains(b.added, a) })
		switch {
		case b.every, knob.Kind == aplkit.KnobThreshold, len(b.added) > 0 && newNeed:
			out = append(out, knob)
		}
	}
	// Any new ability can change which imbue is best, not only a new imbue. Windfury Totem at
	// 32 makes Rockbiter beat Windfury Weapon again, say.
	return out, b.every || len(b.added) > 0
}

// The knobs to try together with each imbue: the ones the range has the abilities for.
func (rs *rotationSearch) imbueKnobs(b *rotationBand) []aplkit.Knob {
	var out []aplkit.Knob
	for _, knob := range rs.template.Knobs() {
		if knob.WithImbue && slices.ContainsFunc(knob.Needs, func(a string) bool { return slices.Contains(b.abilities, a) }) {
			out = append(out, knob)
		}
	}
	return out
}

func (rs *rotationSearch) imbues(b *rotationBand) []proto.WeaponImbue {
	var out []proto.WeaponImbue
	for _, x := range imbueAbilities {
		if slices.Contains(b.abilities, x.ability) {
			out = append(out, x.imbue)
		}
	}
	return out
}

// tune tunes each range that needs it, in order, and says how many it tuned. A range of a
// pass before with the same levels and talents keeps its knobs. Any other starts from that
// pass's knobs at its level, or the range before it.
func (rs *rotationSearch) tune(bands, before []*rotationBand, defaults aplkit.Knobs, defaultImbue func(level int32) proto.WeaponImbue, talentsAt func(int32) string) (int, error) {
	tuned := 0
	for i, b := range bands {
		if old := sameBand(before, b); old != nil {
			b.knobs, b.imbue, b.sims = maps(old.knobs), old.imbue, old.sims
			continue
		}
		switch old := bandAt(before, b.level); {
		case old != nil:
			b.knobs, b.imbue = maps(old.knobs), old.imbue
		case i > 0:
			b.knobs, b.imbue = maps(bands[i-1].knobs), bands[i-1].imbue
		default:
			b.knobs, b.imbue = maps(defaults), defaultImbue(b.level)
		}
		if !slices.Contains(rs.imbues(b), b.imbue) {
			if imbues := rs.imbues(b); len(imbues) > 0 {
				b.imbue = imbues[0]
			}
		}
		if err := rs.tuneBand(b, talentsAt); err != nil {
			return tuned, err
		}
		tuned++
	}
	return tuned, nil
}

// The range of the pass before with the same levels, talents and reason to start, or nil.
func sameBand(before []*rotationBand, b *rotationBand) *rotationBand {
	for _, old := range before {
		if old.level == b.level && slices.Equal(old.levels, b.levels) && old.talents == b.talents && old.every == b.every && slices.Equal(old.added, b.added) {
			return old
		}
	}
	return nil
}

// The range of the pass before that has the level, or nil.
func bandAt(before []*rotationBand, level int32) *rotationBand {
	var found *rotationBand
	for _, old := range before {
		if old.level <= level {
			found = old
		}
	}
	return found
}

func maps(k aplkit.Knobs) aplkit.Knobs {
	out := aplkit.Knobs{}
	for name, v := range k {
		out[name] = v
	}
	return out
}

// tuneBand makes rounds of changes to the range's knobs: in each, every one-knob change goes
// through the elimination, and we take the winners. More than one winner, of different knobs,
// we try together too, and take that when it saves more than the best one alone.
func (rs *rotationSearch) tuneBand(b *rotationBand, talentsAt func(int32) string) error {
	knobs, tuneImbue := rs.knobsToTune(b)
	rounds := fullRounds
	if !b.every && len(b.added) == 0 {
		rounds = touchUpRounds
	}
	current := rotationCandidate{knobs: b.knobs, imbue: b.imbue}
	for round := 1; round <= rounds; round++ {
		var candidates []rotationCandidate
		for _, knob := range knobs {
			for _, v := range knob.Values() {
				near := knob.Kind != aplkit.KnobThreshold || math.Abs(v-current.knobs[knob.Name]) <= thresholdReach*knob.Step+1e-9
				if v != current.knobs[knob.Name] && near {
					next := maps(current.knobs)
					next[knob.Name] = v
					candidates = append(candidates, rotationCandidate{knobs: next, imbue: current.imbue, changed: knob.Name})
				}
			}
		}
		if tuneImbue {
			for _, imbue := range rs.imbues(b) {
				if imbue == current.imbue {
					continue
				}
				candidates = append(candidates, rotationCandidate{knobs: current.knobs, imbue: imbue, changed: "imbue"})
				for _, knob := range rs.imbueKnobs(b) {
					for _, v := range knob.Values() {
						if v != current.knobs[knob.Name] {
							next := maps(current.knobs)
							next[knob.Name] = v
							candidates = append(candidates, rotationCandidate{knobs: next, imbue: imbue, changed: "imbue+" + knob.Name})
						}
					}
				}
			}
		}
		rs.progress(fmt.Sprintf("level %d (%s), round %d, %d changes", b.level, b.kind(), round, len(candidates)))
		winners, err := rs.eliminate(b, current, candidates, talentsAt)
		if err != nil {
			return err
		}
		if len(winners) == 0 {
			break
		}
		best := winners[0]
		// The best winner of each other knob, on top of the best one.
		together := rotationCandidate{knobs: maps(best.c.knobs), imbue: best.c.imbue, changed: "together"}
		seen := map[string]bool{}
		for _, name := range strings.Split(best.c.changed, "+") {
			seen[name] = true
		}
		combined := false
		for _, w := range winners[1:] {
			names := strings.Split(w.c.changed, "+")
			if slices.ContainsFunc(names, func(name string) bool { return seen[name] }) {
				continue
			}
			for _, name := range names {
				seen[name] = true
				if name == "imbue" {
					together.imbue = w.c.imbue
				} else {
					together.knobs[name] = w.c.knobs[name]
				}
			}
			combined = true
		}
		next := best.c
		if combined {
			scores, err := rs.score(b, []rotationCandidate{current, best.c, together}, halvingSessions[len(halvingSessions)-1], talentsAt)
			if err != nil {
				return err
			}
			if scores[2].hours < scores[1].hours && scores[0].hours-scores[2].hours > winnerNoise*scores[2].noiseWith(scores[0]) {
				next = together
			}
		}
		next.changed = ""
		current = next
	}
	b.knobs, b.imbue = current.knobs, current.imbue
	return nil
}

// What a candidate's sims came to over a range: hours and their variance.
type bandScore struct {
	hours, variance float64
}

func (s bandScore) noiseWith(other bandScore) float64 { return math.Sqrt(s.variance + other.variance) }

type winner struct {
	c     rotationCandidate
	saved float64
}

// eliminate sims the current knobs and every candidate at the first sessions, and the best
// quarter again at the next, until a few are left or the last sessions. It gives the candidates
// that save more than their noise at the last step, the most first.
func (rs *rotationSearch) eliminate(b *rotationBand, current rotationCandidate, candidates []rotationCandidate, talentsAt func(int32) string) ([]winner, error) {
	alive := candidates
	for step, sessions := range halvingSessions {
		last := step == len(halvingSessions)-1
		if !last && len(alive) <= 3 {
			continue
		}
		scores, err := rs.score(b, append([]rotationCandidate{current}, alive...), sessions, talentsAt)
		if err != nil {
			return nil, err
		}
		var ws []winner
		for i, c := range alive {
			ws = append(ws, winner{c, scores[0].hours - scores[i+1].hours})
		}
		sort.SliceStable(ws, func(i, j int) bool { return ws[i].saved > ws[j].saved })
		if last {
			var out []winner
			for i, w := range ws {
				idx := slices.IndexFunc(alive, func(c rotationCandidate) bool { return c.changed == w.c.changed && sameCandidate(c, w.c) })
				if w.saved > winnerNoise*scores[idx+1].noiseWith(scores[0]) {
					out = append(out, ws[i])
				}
			}
			return out, nil
		}
		alive = alive[:0:0]
		for _, w := range ws[:(len(ws)+3)/4] {
			alive = append(alive, w.c)
		}
	}
	return nil, nil
}

func sameCandidate(a, b rotationCandidate) bool {
	if a.imbue != b.imbue || len(a.knobs) != len(b.knobs) {
		return false
	}
	for name, v := range a.knobs {
		if b.knobs[name] != v {
			return false
		}
	}
	return true
}

// score sims each candidate on the range's sampled levels, all on the same new sessions, and
// gives the range's hours: each sampled level counts for its share of the range's levels.
func (rs *rotationSearch) score(b *rotationBand, candidates []rotationCandidate, sessions int32, talentsAt func(int32) string) ([]bandScore, error) {
	sample := sampleLevels(b.levels)
	if !b.every && len(b.added) == 0 {
		sample = b.levels[:1]
	}
	weight := float64(len(b.levels)) / float64(len(sample))
	var jobs []levelJob
	for _, c := range candidates {
		rotation := rs.template.Build(b.level, c.knobs)
		for _, level := range sample {
			jobs = append(jobs, levelJob{level: level, talents: talentsAt(level), rotation: rotation, imbue: c.imbue})
		}
	}
	rs.step++
	results, done := rs.p.simLevels(rs.ctx, jobs, sessions, rs.p.seed+rs.step*7_919)
	b.sims += len(jobs)
	out := make([]bandScore, len(candidates))
	for i := range jobs {
		if !done[i] {
			return nil, errStopped
		}
		r := results[i]
		s := &out[i/len(sample)]
		s.hours += r.hours * weight
		s.variance += (r.hours * r.relativeError * weight) * (r.hours * r.relativeError * weight)
	}
	return out, nil
}

// The first, middle and last of the levels.
func sampleLevels(levels []int32) []int32 {
	if len(levels) <= bandSampleLevels {
		return levels
	}
	return []int32{levels[0], levels[len(levels)/2], levels[len(levels)-1]}
}

// planWith is the plan with the ranges' rotations in its library ("Searched 32") and its level
// ranges, and the plan's own rotations kept in the library.
func (rs *rotationSearch) planWith(plan *proto.LevelingPlanSettings, bands []*rotationBand) *proto.LevelingPlanSettings {
	out := goproto.Clone(plan).(*proto.LevelingPlanSettings)
	out.RotationLibrary = slices.DeleteFunc(out.RotationLibrary, func(r *proto.LevelingRotationPreset) bool { return strings.HasPrefix(r.Name, searchedPrefix) })
	out.Rotations = nil
	for _, b := range bands {
		name := fmt.Sprintf("%s%d", searchedPrefix, b.level)
		out.RotationLibrary = append(out.RotationLibrary, &proto.LevelingRotationPreset{Name: name, Rotation: rs.template.Build(b.level, b.knobs)})
		out.Rotations = append(out.Rotations, &proto.LevelingRotation{Level: b.level, Name: name, Imbue: b.imbue})
	}
	out.RotationsNameTopRanks = true
	return out
}

const searchedPrefix = "Searched "

// The ranges as the result has them, each with the knobs it changed from the one before.
func bandsToProto(bands []*rotationBand) []*proto.LevelingRotationBand {
	var out []*proto.LevelingRotationBand
	var prev aplkit.Knobs
	prevImbue := proto.WeaponImbue_WeaponImbueUnknown
	for _, b := range bands {
		var changes []string
		for _, name := range slices.Sorted(keys(b.knobs)) {
			if prev == nil || prev[name] != b.knobs[name] {
				changes = append(changes, fmt.Sprintf("%s=%g", name, b.knobs[name]))
			}
		}
		if prev == nil {
			changes = []string{b.knobs.String()}
		}
		if b.imbue != prevImbue && prev != nil {
			changes = append(changes, "imbue="+b.imbue.String())
		}
		out = append(out, &proto.LevelingRotationBand{
			Level: b.level, NewAbilities: b.added, Changes: strings.Join(changes, " "), Imbue: b.imbue, Sims: int32(b.sims),
		})
		prev, prevImbue = b.knobs, b.imbue
	}
	return out
}

func keys(k aplkit.Knobs) func(func(string) bool) {
	return func(yield func(string) bool) {
		for name := range k {
			if !yield(name) {
				return
			}
		}
	}
}
