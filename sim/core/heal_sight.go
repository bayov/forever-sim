package core

import (
	"time"

	"github.com/wowsims/classic/sim/core/proto"
)

// healSight is what a healer sees of the raid within one rotation check (Simulation.aplCheck),
// worked out once: the other healers' heals in flight, and its Most Hurt and Lowest Health. A
// healer's rotation reads them many times in a check, in each line's target and in values like
// Health Missing and Expected Healing, at the same instant.
//
// Nothing in the sim changes during a check (aplMemo), and the same terms give the same sums,
// so these are the values seenMissing and pickHealTarget give. To be safe the sight is also dropped
// whenever something it reads could have changed: a player's health, death or stats, a cast
// starting, moving or ending, an aura (Simulation.sightEpoch).
type healSight struct {
	check uint64
	at    time.Duration
	epoch uint64
	// healCasters(viewer)
	casters []*Unit
	// Most Hurt and Lowest Health, once picked in this sight.
	picks [2]healSightPick
}

type healSightPick struct {
	picked  bool
	best    *Unit
	missing float64 // seenMissing(best)
}

// sight is the healer's sight now, nil outside a rotation check or a healing fight.
func (viewer *Unit) sight(sim *Simulation) *healSight {
	if sim == nil || sim.aplCheck == 0 || !sim.Encounter.TrackHealth {
		return nil
	}
	s := &viewer.healSight
	if s.check != sim.aplCheck || s.at != sim.CurrentTime || s.epoch != sim.sightEpoch {
		s.check, s.at, s.epoch = sim.aplCheck, sim.CurrentTime, sim.sightEpoch
		s.casters = sim.healCasters(viewer, s.casters[:0])
		s.picks = [2]healSightPick{}
	}
	return s
}

// current is whether nothing the sight read changed since it was made.
func (s *healSight) current(sim *Simulation) bool {
	return s.check == sim.aplCheck && s.at == sim.CurrentTime && s.epoch == sim.sightEpoch
}

func healSightIndex(pick proto.UnitReference_Type) int {
	if pick == proto.UnitReference_LowestHealth {
		return 1
	}
	return 0
}
