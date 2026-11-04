package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
	goproto "google.golang.org/protobuf/proto"
)

// Hit-trade gear steps (-hit-steps, standards 4).
//
// Every Level 60 set carries gear hit up to the spec's cap. A hit talent or racial on top of
// that does nothing unless the gear gives the same hit back as other stats, so each spec with a
// hit talent has gear sets one step apart: <set>_hit-1, _hit-2, ... carry 10 hit rating (1%) less
// per step, spent on the spec's best other stat at the budget price, and <set>_hit+1, ... carry
// more, for a spec that sits below its cap by design. The talent search wears, for each talent
// build, the step that brings its total hit (talents, racial, buffs) back to the cap, so a hit
// talent is scored as "talent + gear with that much less hit". With -hit-span, it also tries the
// steps around that one and keeps whichever sims best, for specs where hit past the yellow cap
// still pays (a dual wielder's white swings, a hybrid's spells).
type hitSteps struct {
	base   string                       // the gear set path without .gear.json
	gear   map[int]*proto.EquipmentSpec // by step: the change in gear hit in percent, 0 the set itself
	lo, hi int

	kind proto.Stat // MeleeHit (melee and ranged) or SpellHit
	cap  float64
	// School-only hit talents the stat sheet doesn't show (Shadow Focus), per rank in percent.
	school map[string]float64
	trees  []talentTree

	span int

	auto map[string]int // the cap step per talent string
	best map[string]int // the step a span search picked per talent string
}

// loadHitSteps reads <base>.gear.json and every <base>_hit±N.gear.json next to it. capSpec is
// "melee:8", "ranged:8" or "spell:16"; school is "fieldName:pct,..." for school-only hit talents.
func loadHitSteps(base, capSpec, school string, span int, class string) (*hitSteps, error) {
	h := &hitSteps{base: strings.TrimSuffix(base, ".gear.json"), gear: map[int]*proto.EquipmentSpec{},
		span: span, school: map[string]float64{}, auto: map[string]int{}, best: map[string]int{}}
	read := func(step int, path string) (bool, error) {
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			return false, nil
		} else if err != nil {
			return false, err
		}
		eq := &proto.EquipmentSpec{}
		if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(data, eq); err != nil {
			return false, fmt.Errorf("%s: %v", path, err)
		}
		h.gear[step] = eq
		return true, nil
	}
	if ok, err := read(0, h.base+".gear.json"); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("no gear set %s.gear.json", h.base)
	}
	for _, dir := range []int{-1, 1} {
		for n := 1; ; n++ {
			sign := "+"
			if dir < 0 {
				sign = "-"
			}
			ok, err := read(dir*n, fmt.Sprintf("%s_hit%s%d.gear.json", h.base, sign, n))
			if err != nil {
				return nil, err
			}
			if !ok {
				break
			}
			if dir < 0 {
				h.lo = -n
			} else {
				h.hi = n
			}
		}
	}

	kind, val, ok := strings.Cut(capSpec, ":")
	cap, err := strconv.ParseFloat(val, 64)
	if !ok || err != nil {
		return nil, fmt.Errorf("-hit-cap %q: want melee:8, ranged:8 or spell:16", capSpec)
	}
	switch kind {
	case "melee", "ranged":
		h.kind = proto.Stat_StatMeleeHit
	case "spell":
		h.kind = proto.Stat_StatSpellHit
	default:
		return nil, fmt.Errorf("-hit-cap %q: want melee, ranged or spell", capSpec)
	}
	h.cap = cap

	if school != "" {
		if h.trees, err = loadTalentTrees(class); err != nil {
			return nil, err
		}
		for _, kv := range strings.Split(school, ",") {
			name, pct, _ := strings.Cut(kv, ":")
			v, err := strconv.ParseFloat(pct, 64)
			if err != nil {
				return nil, fmt.Errorf("-hit-school %q: want fieldName:percent per rank", kv)
			}
			h.school[name] = v
		}
	}
	return h, nil
}

// The steps a search with -hit-span tries for this talent string, the cap step first: span
// steps either side of the cap step, and for a set worn past the cap (a dual wielder's, whose
// white swings still gain), up to span steps past the set itself too, so the search can keep or
// raise the set's own hit (phase 5: Fury on the set's 10% has cap step -2, and +2 won).
func (h *hitSteps) candidates(s *setup, talents string) []int {
	auto := h.capStep(s, talents)
	steps := []int{auto}
	for st := auto - h.span; st <= max(auto, 0)+h.span; st++ {
		if st != auto && st >= h.lo && st <= h.hi {
			steps = append(steps, st)
		}
	}
	return steps
}

// capStep is the step that brings total hit on the set (step 0) back to the cap: the hit
// past the cap, rounded to whole percent, taken off the gear, or the hit short of it added.
func (h *hitSteps) capStep(s *setup, talents string) int {
	if st, ok := h.auto[talents]; ok {
		return st
	}
	excess := h.totalHit(s, talents) - h.cap
	st := -int(math.Floor(excess + 0.5 + 1e-9))
	st = max(h.lo, min(h.hi, st))
	h.auto[talents] = st
	return st
}

// totalHit is the player's final hit of the cap's kind on the set itself (step 0), with the
// school-only hit talents added.
func (h *hitSteps) totalHit(s *setup, talents string) float64 {
	request := goproto.Clone(s.request).(*proto.RaidSimRequest)
	player := request.Raid.Parties[0].Players[0]
	player.TalentsString = talents
	player.Equipment = goproto.Clone(h.gear[0]).(*proto.EquipmentSpec)
	stats := core.ComputeStats(&proto.ComputeStatsRequest{Raid: request.Raid, Encounter: request.Encounter, Ruleset: request.SimOptions.Ruleset})
	hit := stats.RaidStats.Parties[0].Players[0].FinalStats.Stats[h.kind]
	if len(h.school) > 0 {
		b := parseBuild(h.trees, talents)
		for ti, tree := range h.trees {
			for i, t := range tree.talents {
				if pct, ok := h.school[t.name]; ok {
					hit += pct * float64(b[ti][i])
				}
			}
		}
	}
	return hit
}

// step is the step this talent string wears: a span search's pick when there is one, else
// the cap step.
func (h *hitSteps) step(s *setup, talents string) int {
	if s.hitStep != nil {
		return *s.hitStep
	}
	if st, ok := h.best[talents]; ok {
		return st
	}
	return h.capStep(s, talents)
}

func (h *hitSteps) name(step int) string {
	if step == 0 {
		return h.base
	}
	return fmt.Sprintf("%s_hit%+d", h.base, step)
}

// effectiveTalents is the talent string a run uses.
func (s *setup) effectiveTalents() string {
	if s.talents != "" {
		return s.talents
	}
	return s.player().TalentsString
}

// wearHitStep puts the talent string's step on the request's player.
func (s *setup) wearHitStep(request *proto.RaidSimRequest) {
	if s.hit == nil {
		return
	}
	st := s.hit.step(s, s.effectiveTalents())
	request.Raid.Parties[0].Players[0].Equipment = goproto.Clone(s.hit.gear[st]).(*proto.EquipmentSpec)
}

// The cache key part for the step, empty without -hit-steps.
func (s *setup) hitKey() string {
	if s.hit == nil {
		return ""
	}
	return fmt.Sprintf("|hit%+d", s.hit.step(s, s.effectiveTalents()))
}

// prepareSteps is prepare for every step the talent string tries: one without -hit-span.
func (s *searcher) prepareSteps(knobs Knobs) ([]*pendingScore, []int) {
	h := s.setup.hit
	if h == nil || h.span == 0 || s.setup.hitStep != nil {
		return []*pendingScore{s.prepare(knobs)}, nil
	}
	steps := h.candidates(s.setup, s.setup.effectiveTalents())
	var pending []*pendingScore
	for i := range steps {
		s.setup.hitStep = &steps[i]
		pending = append(pending, s.prepare(knobs))
	}
	s.setup.hitStep = nil
	return pending, steps
}

// pickStep keeps the best of a talent string's step scores and remembers its step.
func (s *searcher) pickStep(talents string, results []result, steps []int) result {
	if steps == nil {
		return results[0]
	}
	bi := 0
	for i := range results {
		if results[i].dps > results[bi].dps {
			bi = i
		}
	}
	s.setup.hit.best[talents] = steps[bi]
	return results[bi]
}

// hitStepNote says which step the talents wear, for the search's output.
func (s *setup) hitStepNote() string {
	if s.hit == nil {
		return ""
	}
	t := s.effectiveTalents()
	st := s.hit.step(s, t)
	return fmt.Sprintf("hit step %+d (%s.gear.json; cap step %+d, total hit on the set %.2f%% against %g%%)",
		st, s.hit.name(st), s.hit.capStep(s, t), s.hit.totalHit(s, t), s.hit.cap)
}
