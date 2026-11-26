package core

import (
	"math"
	"time"
)

// The mobs around us in a leveling session when each has its own place (camp_mobs,
// docs/leveling_sim.md, "Big pulls"). Without a camp there is always another mob 10 to 40 yards
// away, which is fine for pulls of one to three. A pull of ten takes every mob in sight, so the
// walk to each one is longer than the last, and the spot is empty until they respawn.
//
//   - Around the spot we start at, camp_mobs mobs are strewn at random over a circle of 100
//     yards, and camp_outside_mobs over each circle of that size outside it. 1.12's spawns
//     (tools/leveling_density.py) have the same walks from mob to mob.
//   - The mob we find is the nearest one. Out of combat we loot what we killed, look around and
//     walk up until it's found_max yards away. In combat we look around, and the rotation walks
//     to it (Walk To Pull Target) with the mobs on us following, or leaves it.
//   - Any mob notices us once we're within its aggro radius, also the ones we only pass and
//     the ones that respawn next to us, and brings the mobs within 10 yards of it along.
//   - A dead mob is back at its place after respawn seconds.
//
// We know no walls, roads or other players here, and the mobs don't wander.
const (
	campRadius = 100
	// How far out the mobs are strewn, further than a session gets.
	campOutsideRadius = 600
	// A mob that's pulled brings the mobs this close to it along (cmangos' call for help).
	campLinkRadius = 10
	// How far from our corpse we come back to life.
	campReviveReach = 35
	// How far we see a mob to look for another one in combat. The rotation decides whether it's
	// worth the walk (the Pull Target's Leveling Distance).
	campSight = 100
)

// campMob is a mob's place, and the mob of the pool that stands for it while it's found or
// pulled.
type campMob struct {
	x     Vec2
	level int32
	// When it stands there again.
	upAt time.Duration
	unit *Unit
}

func (c *campMob) up(sim *Simulation) bool {
	return c.unit == nil && c.upAt <= sim.CurrentTime
}

// A usualCamp is how many mobs there are at a good spot from a level on: 1 in 4 mobs of our level
// stands at a denser one (tools/leveling_density.py, docs/leveling_sim.md, Big pulls).
type usualCamp struct {
	level   int32
	mobs    int
	outside float64
	respawn time.Duration
}

var usualCamps = []usualCamp{
	{1, 20, 10, 250 * time.Second},
	{8, 15, 10, 275 * time.Second},
	{16, 12, 7, 300 * time.Second},
	{36, 11, 7, 300 * time.Second},
	{53, 12, 7, 333 * time.Second},
}

func usualCampAt(level int32) usualCamp {
	camp := usualCamps[0]
	for _, c := range usualCamps {
		if c.level <= level {
			camp = c
		}
	}
	return camp
}

// resetCamp strews the session's mobs anew.
func (s *levelingSession) resetCamp(sim *Simulation) {
	s.notice, s.linking, s.toLoot = nil, false, 0
	s.camp = s.camp[:0]
	if s.opts.campMobs == 0 {
		s.camp = nil
		return
	}
	sim.reseedWithin("Leveling Camp", 0)
	add := func(inner, outer float64) {
		r := math.Sqrt(inner*inner + sim.RandomFloat("Leveling Camp Distance")*(outer*outer-inner*inner))
		x := vec2FromAngle(2 * math.Pi * sim.RandomFloat("Leveling Camp Direction")).Scale(r)
		level := s.levels[int(sim.RandomFloat("Leveling Camp Level")*float64(len(s.levels)))%len(s.levels)]
		s.camp = append(s.camp, campMob{x: x, level: level})
	}
	for range s.opts.campMobs {
		add(0, campRadius)
	}
	outside := s.opts.campOutside * (campOutsideRadius*campOutsideRadius/(campRadius*campRadius) - 1)
	for range int(math.Round(outside)) {
		add(campRadius, campOutsideRadius)
	}
}

// nearestUp is the place of the nearest mob that stands within yards of us, or -1.
func (s *levelingSession) nearestUp(sim *Simulation, yards float64) (int, float64) {
	here := s.playerPosition(sim)
	nearest := -1
	for i := range s.camp {
		if c := &s.camp[i]; c.up(sim) {
			if d := c.x.Dist(here); d <= yards {
				nearest, yards = i, d
			}
		}
	}
	return nearest, yards
}

// campSearch: out of combat, we loot what we killed, look around and walk to the next mob.
func (s *levelingSession) campSearch(sim *Simulation) {
	s.move(sim, goalNone, Vec2{}, 0, nil)
	s.scheduleNotice(sim)
	delay := time.Duration(s.toLoot)*s.opts.lootTime + randomDuration(sim, "Leveling Search", s.opts.searchMin, s.opts.searchMax)
	s.toLoot = 0
	s.after(sim, &s.next, delay, s.campApproach)
}

// campApproach walks up to the nearest mob until it's near enough to be the one we found.
func (s *levelingSession) campApproach(sim *Simulation) {
	if s.state != levelingSearching {
		return
	}
	i, yards := s.nearestUp(sim, math.Inf(1))
	again := time.Second
	if i >= 0 && yards <= s.opts.foundMax+levelingYardTolerance {
		s.move(sim, goalNone, Vec2{}, 0, nil)
		sim.reseedWithin("Leveling Search Find", s.searches)
		if s.campFind(sim) != nil {
			s.startPull(sim)
			return
		}
	} else if i >= 0 {
		x := s.camp[i].x
		if s.goal != goalPoint || s.goalPoint != x {
			s.move(sim, goalPoint, x, s.opts.foundMax, nil)
		}
		again = min(again, DurationFromSeconds((yards-s.opts.foundMax)/s.runSpeed)+time.Millisecond)
	}
	s.after(sim, &s.next, again, s.campApproach)
}

// campFind makes the nearest mob the one we found, in combat only when it's in sight.
func (s *levelingSession) campFind(sim *Simulation) *Unit {
	sight := math.Inf(1)
	if s.state == levelingCombat {
		sight = campSight
	}
	i, _ := s.nearestUp(sim, sight)
	if i < 0 {
		return nil
	}
	mob := s.campMob(sim, i)
	if mob != nil {
		s.foundMob(sim, mob)
	}
	return mob
}

// campMob takes a mob of the pool for a place, nil when every one of the pool is out.
func (s *levelingSession) campMob(sim *Simulation, i int) *Unit {
	c := &s.camp[i]
	mob := s.idleMob(c.level)
	if mob == nil {
		return nil
	}
	c.unit = mob
	s.place(sim, mob, c.x)
	s.mobs[mob.Index].camp = i + 1
	return mob
}

// campKilled: the mob died, and stands at its place again later.
func (s *levelingSession) campKilled(sim *Simulation, mob *Unit) {
	if i := s.mobs[mob.Index].camp; i > 0 {
		s.camp[i-1].upAt = sim.CurrentTime + s.opts.respawnTime
		s.toLoot++
	}
}

// campReturned: the mob left us, dead or not. One that's alive stands at its place at once.
func (s *levelingSession) campReturned(mob *Unit) {
	if i := s.mobs[mob.Index].camp; i > 0 {
		s.camp[i-1].unit = nil
	}
}

// scheduleNotice times when the next mob we haven't found notices us.
func (s *levelingSession) scheduleNotice(sim *Simulation) {
	cancel(sim, &s.notice)
	if s.camp == nil || s.state == levelingDead {
		return
	}
	here := s.playerPosition(sim)
	first, at := -1, math.Inf(1)
	for i := range s.camp {
		c := &s.camp[i]
		if c.unit != nil {
			continue
		}
		wait := max(c.upAt-sim.CurrentTime, 0).Seconds()
		if wait >= at {
			continue
		}
		t, ok := whenWithin(c.x.Sub(here.Add(s.vel.Scale(wait))), s.vel.Scale(-1), s.aggroRadiusAt(c.level))
		if ok && wait+t < at {
			first, at = i, wait+t
		}
	}
	if first < 0 || at > s.opts.sessionLength.Seconds() {
		return
	}
	s.after(sim, &s.notice, DurationFromSeconds(at), func(sim *Simulation) {
		if !s.camp[first].up(sim) {
			s.scheduleNotice(sim)
			return
		}
		mob := s.campMob(sim, first)
		if mob == nil {
			// Every mob of the pool is out. It gets its turn when one dies.
			s.after(sim, &s.notice, time.Second, s.scheduleNotice)
			return
		}
		if sim.Log != nil {
			s.player.Log(sim, "Leveling: %s noticed us", mob.Label)
		}
		s.aggro(sim, mob)
	})
}

// campPulled: a mob was pulled at x, and the ones next to it come along.
func (s *levelingSession) campPulled(sim *Simulation, x Vec2) {
	if s.camp == nil || s.linking {
		return
	}
	s.linking = true
	if found := s.found; found != nil && s.mobs[found.Index].x.Dist(x) <= campLinkRadius {
		s.aggro(sim, found)
	}
	for i := range s.camp {
		if c := &s.camp[i]; c.up(sim) && c.x.Dist(x) <= campLinkRadius {
			if mob := s.campMob(sim, i); mob != nil {
				s.aggro(sim, mob)
			}
		}
	}
	s.linking = false
	s.scheduleNotice(sim)
}

// campRevive has us come back to life near our corpse where the fewest mobs notice us, as the
// game lets us pick a spot that far from it.
func (s *levelingSession) campRevive(sim *Simulation) {
	if s.camp == nil {
		return
	}
	corpse := s.playerPosition(sim)
	// How far the nearest mob is from noticing us there.
	room := func(spot Vec2) float64 {
		least := math.Inf(1)
		for i := range s.camp {
			if c := &s.camp[i]; c.up(sim) {
				least = min(least, c.x.Dist(spot)-s.aggroRadiusAt(c.level))
			}
		}
		return least
	}
	best, most := corpse, room(corpse)
	for step := range 16 {
		for _, yards := range []float64{campReviveReach / 2, campReviveReach} {
			spot := corpse.Add(vec2FromAngle(2 * math.Pi * float64(step) / 16).Scale(yards))
			if r := room(spot); r > most {
				best, most = spot, r
			}
		}
	}
	s.pos, s.posAt = best, sim.CurrentTime
	s.toLoot = 0
	s.logPosition(sim)
}
