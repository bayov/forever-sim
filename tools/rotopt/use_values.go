package main

import (
	"flag"
	"fmt"
	"slices"
	"strings"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	goproto "google.golang.org/protobuf/proto"
)

// Use values (docs/bis_lists.md, "How uses count"). The user's fights at level 60 are short and
// an addon swaps trinkets between them, but there's little time to regen mana between fights.
// So a set is scored on the full fight with every use off, and each use counts for what it adds
// on a shorter fight (user, 2026-11-20):
//
//	score(S) = DPS at -durs of S with every use off + the sum of useValue(i, S) over S's uses
//	useValue(i, S) = DPS at -use-dur of S - the same with only i's use off
//
// The short fight is 90s. Swapping a usable item in also puts it on a 30s equip cooldown, so
// sometimes its use never gets pressed in the fight, and 90s takes that into account roughly.
//
// With -use-dur every gear search score works this way (prepare): the search, the BiS list's
// ranks and picks, its losses, the conditional ranks and the enchants.
//
// The use values of a set depend mostly on which uses it holds, so we measure them once for
// each such group of uses (useSet) and reuse them for every set that holds the same uses. An
// option without a use then costs only its run at -durs, and one with a use adds the runs at
// -use-dur for its group. The BiS ranking starts over for each base set, so the base's own
// values are measured on the base itself.

var (
	useDur   = flag.Float64("use-dur", 0, "gear search: score each item's use on a fight this many seconds long, and everything else at -durs with every use off (docs/bis_lists.md, How uses count). 0 scores the uses at -durs like the rest")
	useIters = flag.Int("use-iters", 20000, "the iterations of each -use-dur run")
	// -use-aoe-targets is the "vs 4 mobs" conditional of bis_aoe.go.
	useAoeTargets = flag.Int("use-aoe-targets", 0, "gear search: with -gear-rank-conditions or -gear-effects-out, also value each item's use against this many mobs, and list the items whose use gains on them (docs/bis_lists.md, Uses against several mobs). 0 leaves it out")
)

// A sim run at -use-dur, and its result once it ran.
type useRun struct {
	request *proto.RaidSimRequest
	result  result
}

// useSet is the use values of one group of uses on one fight: a run with every use pressed, and
// one with each use off.
type useSet struct {
	on   *useRun
	offs map[int32]*useRun
}

// value is the use value of item, and its paired standard error.
func (u *useSet) value(item int32) (float64, float64) {
	off := u.offs[item]
	return u.on.result.dps - off.result.dps, gainNoise(off.result, u.on.result)
}

// resetUseValues forgets the use values, so the next set we score measures them again. The BiS
// ranking calls it for each new base set.
func (s *searcher) resetUseValues() {
	s.useSets = map[string]*useSet{}
}

// gearUses is the items of the setup's gear that have a use. The sim only tells us after it built
// a character wearing the item (core.ItemHasUse), so we build one for the items we haven't seen
// yet: one iteration of a one second fight.
func (s *searcher) gearUses() []int32 {
	player := s.setup.player()
	if s.probed == nil {
		s.probed = map[int32]bool{}
	}
	var ids []int32
	probe := false
	for _, item := range player.Equipment.GetItems() {
		if item.GetId() != 0 {
			ids = append(ids, item.Id)
			probe = probe || !s.probed[item.Id]
		}
	}
	if probe {
		request := goproto.Clone(s.setup.request).(*proto.RaidSimRequest)
		request.Raid.Parties[0].Players[0].ItemUsesOff = nil
		request.Raid.Parties[0].Players[0].ItemEffectsOff = nil
		request.Encounter.Duration = 1
		request.Encounter.DurationVariation = 0
		request.SimOptions.Iterations = 1
		if res := core.RunRaidSim(request); res.Error != nil {
			fmt.Printf("probing the gear's uses: %s\n", res.Error.Message)
		}
		for _, id := range ids {
			s.probed[id] = true
		}
	}
	return slices.DeleteFunc(ids, func(id int32) bool { return !core.ItemHasUse(id) })
}

// prepareUses makes the score p a use score: its run at -durs goes with every use of the gear off,
// and it adds the use values of the uses the item switches leave on (setItemSwitches). The use
// values come from the group's useSet, which we prepare runs for when it's new.
func (s *searcher) prepareUses(p *pendingScore, rot *proto.APLRotation, knobs Knobs) {
	uses := s.gearUses()
	player := s.setup.player()
	for _, request := range p.requests {
		if request != nil {
			off := &request.Raid.Parties[0].Players[0].ItemUsesOff
			for _, id := range uses {
				if !slices.Contains(*off, id) {
					*off = append(*off, id)
				}
			}
		}
	}
	for _, id := range uses {
		if !slices.Contains(player.ItemUsesOff, id) && !slices.Contains(player.ItemEffectsOff, id) {
			p.useItems = append(p.useItems, id)
		}
	}
	if len(uses) == 0 {
		return
	}
	sorted := slices.Clone(uses)
	slices.Sort(sorted)
	enc := s.setup.request.Encounter
	var targets []string
	for _, t := range enc.GetTargets() {
		targets = append(targets, fmt.Sprintf("%d/%d", t.MobType, t.MobFamily))
	}
	key := fmt.Sprintf("%s%s|%s|%s|%d|%d|%v", s.setup.talents, s.setup.hitKey(), knobs, strings.Join(targets, ","), enc.GetOutdoorTerrain(), s.useTargets, sorted)
	set := s.useSets[key]
	if set == nil {
		set = &useSet{offs: map[int32]*useRun{}}
		s.useSets[key] = set
		run := func(off int32) *useRun {
			request := s.setup.runRequest(rot, *useDur, int32(*useIters))
			if s.useTargets > 0 {
				setTargetCount(request.Encounter, s.useTargets)
			}
			pl := request.Raid.Parties[0].Players[0]
			pl.ItemUsesOff, pl.ItemEffectsOff = nil, nil
			if off != 0 {
				pl.ItemUsesOff = []int32{off}
			}
			request.SimOptions.SaveAllValues = true
			request.SimOptions.UseLabeledRands = true
			r := &useRun{request: request}
			s.useRuns = append(s.useRuns, r)
			return r
		}
		set.on = run(0)
		for _, id := range sorted {
			set.offs[id] = run(id)
		}
	}
	p.uses = set
}

// takeUseRuns hands scoreAll the use runs prepared since it last ran, to run side by side with
// its other runs.
func (s *searcher) takeUseRuns() []*useRun {
	runs := s.useRuns
	s.useRuns = nil
	return runs
}

// addUses adds the use values to a score's result at -durs.
func addUses(r *result, p *pendingScore) {
	if p.uses == nil {
		return
	}
	r.uses, r.useErrs = map[int32]float64{}, map[int32]float64{}
	for _, id := range p.useItems {
		v, se := p.uses.value(id)
		r.uses[id], r.useErrs[id] = v, se
		r.useSum += v
	}
	r.dps += r.useSum
}
