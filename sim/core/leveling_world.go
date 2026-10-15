package core

import (
	"math"
	"time"
)

// Where things are in a leveling session (leveling.go): a line we move along. We have a
// position on it, and so does every mob out of the pool and every totem (LevelingPosition).
//
//   - We run forward while we search, and walk to a mob or back to a point (a totem) when the
//     session or the rotation says so. While we move no hard cast goes out (Moving), but we
//     keep swinging at a mob on us.
//   - The mob we find stands ahead of us, 10 to 40 yards away (found_min / found_max).
//   - A pulled mob runs at us at its level's speed, following us if we move, and reaches us
//     5 yards away (melee range). It then stays on us wherever we go.
//   - A mob we haven't pulled notices us once we're within its aggro radius, and pulls itself:
//     20 yards for a mob of our level, a yard more for each level it has over us and a yard
//     less for each level under, never under 5 (cmangos Creature::GetAttackDistance). One
//     we find that close notices us at once.

const (
	// Where a mob that reached us stands.
	levelingMeleeRange = 5
	// How long the rotation has to pull a mob before we walk up to it, counted from when we
	// found it or last stopped moving.
	levelingFacePullAfter = time.Second
	// How often the session checks whether to walk up to the mob we found.
	levelingPullPoll = 250 * time.Millisecond
)

type levelingMobState int

const (
	mobInPool levelingMobState = iota
	mobFound
	mobRunning
	mobArrived
)

type levelingMob struct {
	state levelingMobState
	// Position at xAt: where a found mob stands, or a running one was, running in dir.
	x, dir  float64
	xAt     time.Duration
	arrival *PendingAction
	// A spell that attacks it is on its way (a bolt in flight): it's pulled, and no longer
	// the rotation's Pull Target.
	incoming bool
	// Found by pack_size's search rather than the rotation's, so we walk up to it if the
	// rotation doesn't pull it.
	autoPull bool
}

// Why we're moving.
type levelingGoal int

const (
	goalNone levelingGoal = iota
	// Running around looking for a mob, until the search ends.
	goalSearch
	// Walking up to the mob we found, until we're within goalWithin yards of it or it
	// notices us.
	goalMob
	// Walking to goalPoint, until we're within goalWithin yards of it.
	goalPoint
)

func (s *levelingSession) playerPosition(sim *Simulation) float64 {
	return s.pos + s.vel*(sim.CurrentTime-s.posAt).Seconds()
}

func mobRunSpeed(mob *Unit) float64 {
	return NormalMobStatsAt(mob.Level).RunSpeed
}

// mobPosition is where the mob is; one on us is where we are.
func (s *levelingSession) mobPosition(sim *Simulation, mob *Unit) float64 {
	m := &s.mobs[mob.Index]
	switch m.state {
	case mobFound:
		return m.x
	case mobRunning:
		return m.x + m.dir*mobRunSpeed(mob)*(sim.CurrentTime-m.xAt).Seconds()
	}
	return s.playerPosition(sim)
}

// distance is how far the mob is from us: melee range once it has reached us, and never
// closer while it runs in.
func (s *levelingSession) distance(sim *Simulation, mob *Unit) float64 {
	switch s.mobs[mob.Index].state {
	case mobFound:
		return math.Abs(s.mobPosition(sim, mob) - s.playerPosition(sim))
	case mobRunning:
		return max(levelingMeleeRange, math.Abs(s.mobPosition(sim, mob)-s.playerPosition(sim)))
	case mobArrived:
		return levelingMeleeRange
	}
	return math.Inf(1)
}

// aggroRadius is how close we can get to a mob before it notices us.
func (s *levelingSession) aggroRadius(mob *Unit) float64 {
	levelDiff := max(s.level-mob.Level, -25)
	return max(20-float64(levelDiff), levelingMeleeRange)
}

// setVelocity starts, stops or turns our movement. Starting to move stops a hard cast. The
// mobs running at us change course, and the mob we found may notice us sooner or later.
func (s *levelingSession) setVelocity(sim *Simulation, vel float64) {
	if vel == s.vel {
		return
	}
	for _, mob := range s.env.Encounter.ActiveTargetUnits {
		if m := &s.mobs[mob.Index]; m.state == mobRunning {
			m.x, m.xAt = s.mobPosition(sim, mob), sim.CurrentTime
		}
	}
	s.pos, s.posAt = s.playerPosition(sim), sim.CurrentTime
	if s.vel == 0 && s.player.Hardcast.Expires > sim.CurrentTime {
		s.player.interruptHardcast(sim)
	}
	s.vel = vel
	s.player.MovementHandler.Moving = vel != 0
	for _, mob := range s.env.Encounter.ActiveTargetUnits {
		if s.mobs[mob.Index].state == mobRunning {
			s.scheduleArrival(sim, mob)
		}
	}
	if s.found != nil && s.found == s.player.CurrentTarget {
		s.player.DistanceFromTarget = s.distance(sim, s.found)
	}
	s.scheduleProximity(sim)
}

// move sets off toward a goal, or stops with goalNone.
func (s *levelingSession) move(sim *Simulation, goal levelingGoal, toward float64, within float64, mob *Unit) {
	s.goal, s.goalPoint, s.goalWithin, s.goalMob = goal, toward, within, mob
	cancel(sim, &s.goalReached)
	if goal == goalNone {
		s.setVelocity(sim, 0)
		s.stoppedAt = sim.CurrentTime
		return
	}
	speed := s.player.MovementHandler.MoveSpeed
	if goal == goalSearch {
		s.setVelocity(sim, speed)
		return
	}
	gap := toward - s.playerPosition(sim)
	if math.Abs(gap) <= within {
		s.move(sim, goalNone, 0, 0, nil)
		return
	}
	s.setVelocity(sim, math.Copysign(speed, gap))
	s.after(sim, &s.goalReached, DurationFromSeconds((math.Abs(gap)-within)/speed), func(sim *Simulation) {
		s.move(sim, goalNone, 0, 0, nil)
		// Exactly there, not a rounding error short of a spell's range.
		s.pos = toward - math.Copysign(within, gap)
		s.scheduleProximity(sim)
	})
}

// scheduleArrival times a running mob's arrival from where it and we are now.
func (s *levelingSession) scheduleArrival(sim *Simulation, mob *Unit) {
	m := &s.mobs[mob.Index]
	cancel(sim, &m.arrival)
	p := s.playerPosition(sim)
	gap := math.Abs(p - m.x)
	if gap <= levelingMeleeRange {
		s.after(sim, &m.arrival, 0, func(sim *Simulation) { s.arrive(sim, mob) })
		return
	}
	m.dir = math.Copysign(1, p-m.x)
	closing := mobRunSpeed(mob) - m.dir*s.vel
	if closing <= 0 {
		return // We outrun it, for now.
	}
	s.after(sim, &m.arrival, DurationFromSeconds((gap-levelingMeleeRange)/closing), func(sim *Simulation) { s.arrive(sim, mob) })
}

// scheduleProximity times when the mob we found notices us, if we're heading its way.
func (s *levelingSession) scheduleProximity(sim *Simulation) {
	cancel(sim, &s.proximity)
	mob := s.found
	if mob == nil || s.mobs[mob.Index].incoming {
		return
	}
	gap := s.distance(sim, mob) - s.aggroRadius(mob)
	switch {
	case gap <= 0:
	case s.vel != 0 && math.Signbit(s.vel) == math.Signbit(s.mobs[mob.Index].x-s.playerPosition(sim)):
		gap /= math.Abs(s.vel)
	default:
		return
	}
	s.after(sim, &s.proximity, DurationFromSeconds(max(gap, 0)), func(sim *Simulation) {
		if s.found == mob {
			if sim.Log != nil {
				s.player.Log(sim, "Leveling: %s noticed us", mob.Label)
			}
			s.aggro(sim, mob)
		}
	})
}

// search: out of combat and ready, we run around looking for the next mob.
func (s *levelingSession) search(sim *Simulation) {
	if s.found != nil {
		// We saw it while fighting, and it's still there.
		s.startPull(sim)
		return
	}
	s.setState(sim, levelingSearching)
	s.move(sim, goalSearch, 0, 0, nil)
	if sim.Log != nil {
		s.player.Log(sim, "Leveling: searching for a mob")
	}
	s.after(sim, &s.next, randomDuration(sim, "Leveling Search", s.opts.searchMin, s.opts.searchMax), func(sim *Simulation) {
		s.move(sim, goalNone, 0, 0, nil)
		if s.find(sim, false) != nil {
			s.startPull(sim)
		} else {
			s.search(sim)
		}
	})
}

// find picks a mob from the pool to stand ahead of us, waiting to be pulled.
func (s *levelingSession) find(sim *Simulation, autoPull bool) *Unit {
	level := s.levels[int(sim.RandomFloat("Leveling Mob Level")*float64(len(s.levels)))%len(s.levels)]
	mob := s.idleMob(level)
	if mob == nil {
		return nil
	}
	distance := math.Round(s.opts.foundMin + sim.RandomFloat("Leveling Mob Distance")*(s.opts.foundMax-s.opts.foundMin))
	mob.dead = false
	mob.enabled = true
	mob.healthBar.currentHealth = mob.MaxHealth()
	mob.auraTracker.expireAll(sim) // Anything cast at it while it lay in the pool.
	mob.auraTracker.respawn(sim)
	s.mobs[mob.Index] = levelingMob{state: mobFound, x: s.playerPosition(sim) + distance, xAt: sim.CurrentTime, autoPull: autoPull}
	s.found, s.foundAt = mob, sim.CurrentTime
	if sim.Log != nil {
		s.player.Log(sim, "Leveling: found %s (level %d) %.0f yards away", mob.Label, mob.Level, distance)
	}
	s.scheduleProximity(sim)
	return mob
}

// startPull: out of combat with a mob found. It's our target, and the rotation pulls it.
func (s *levelingSession) startPull(sim *Simulation) {
	s.setState(sim, levelingPulling)
	s.move(sim, goalNone, 0, 0, nil)
	s.foundAt = sim.CurrentTime
	s.retarget(sim, s.found)
	s.pollPull(sim)
}

// pollPull watches the mob we found: when the rotation hasn't pulled it, isn't casting and
// hasn't moved for a second, we walk up to it until it notices us. Out of combat that's
// always so; in combat only for a mob pack_size's search found, since the rotation that
// looked for one decides what to do with it.
func (s *levelingSession) pollPull(sim *Simulation) {
	mob := s.found
	if mob == nil || (s.state != levelingPulling && s.state != levelingCombat) {
		return
	}
	m := &s.mobs[mob.Index]
	if s.state == levelingCombat && !m.autoPull {
		return
	}
	if m.incoming {
		return
	}
	if s.goal == goalNone && sim.CurrentTime-max(s.foundAt, s.stoppedAt) >= levelingFacePullAfter && !s.player.IsCasting(sim) {
		s.walkToFound(sim, 0)
	}
	s.after(sim, &s.next, levelingPullPoll, s.pollPull)
}

// walkToFound walks up to the mob we found until we're within yards of it, or it notices us.
func (s *levelingSession) walkToFound(sim *Simulation, within float64) {
	mob := s.found
	if sim.Log != nil {
		s.player.Log(sim, "Leveling: walking up to %s", mob.Label)
	}
	s.move(sim, goalMob, s.mobs[mob.Index].x, max(within, levelingMeleeRange), mob)
}

// onCast: a spell of ours went off at a mob. One that attacks the mob we found pulls it when
// it lands (onHit), and until then the rotation shouldn't cast another. A spell that never
// lands (no damage, no miss) pulls it once its travel time is over.
func (s *levelingSession) onCast(sim *Simulation, mob *Unit, spell *Spell) {
	m := &s.mobs[mob.Index]
	if m.state != mobFound || m.incoming || !spell.attacksTarget() {
		return
	}
	m.incoming = true
	cancel(sim, &s.proximity)
	if s.goal == goalMob && s.goalMob == mob {
		s.move(sim, goalNone, 0, 0, nil)
	}
	StartDelayedAction(sim, DelayedActionOptions{
		DoAt: sim.CurrentTime + spell.TravelTime() + 100*time.Millisecond,
		OnAction: func(sim *Simulation) {
			if s.mobs[mob.Index].state == mobFound && s.mobs[mob.Index].incoming {
				s.aggro(sim, mob)
			}
		},
	})
}

// onHit: something of ours landed on a mob. Anything harmful pulls the mob we found.
func (s *levelingSession) onHit(sim *Simulation, mob *Unit, spell *Spell) {
	if s.mobs[mob.Index].state == mobFound && !spell.Flags.Matches(SpellFlagHelpful) {
		s.aggro(sim, mob)
	}
}

// aggro: the mob is pulled and runs at us from where it stands.
func (s *levelingSession) aggro(sim *Simulation, mob *Unit) {
	distance := s.distance(sim, mob)
	x := s.mobPosition(sim, mob)
	if mob == s.found {
		s.found = nil
		cancel(sim, &s.proximity)
	}
	if s.goal == goalMob && s.goalMob == mob {
		s.move(sim, goalNone, 0, 0, nil)
	}
	if s.state != levelingCombat {
		s.setState(sim, levelingCombat)
		if s.goal == goalSearch {
			s.move(sim, goalNone, 0, 0, nil)
		}
		s.pulls++
		s.pullMobs = 0
	}
	s.pullMobs++
	s.mobsPulled++
	s.mobs[mob.Index] = levelingMob{state: mobRunning, x: x, xAt: sim.CurrentTime}
	s.env.Encounter.addActiveTarget(mob)
	if mob == s.player.CurrentTarget {
		s.setPlayerDistance(sim, max(distance, levelingMeleeRange))
	}
	s.scheduleArrival(sim, mob)
	if sim.Log != nil {
		if arrival := s.mobs[mob.Index].arrival; arrival != nil {
			s.player.Log(sim, "Leveling: pulled %s (level %d) from %.0f yards, here in %s", mob.Label, mob.Level, distance, arrival.NextActionAt-sim.CurrentTime)
		} else {
			s.player.Log(sim, "Leveling: pulled %s (level %d) from %.0f yards, not catching up with us", mob.Label, mob.Level, distance)
		}
	}
	s.chainSearch(sim, true)
}

// chainSearch: while fighting, we look for another mob, running around with the mobs on us
// following most of the time. pack_size's search (auto) goes until a pull has taken that
// many mobs; the rotation's (LevelingLookForMob) whenever it asks.
func (s *levelingSession) chainSearch(sim *Simulation, auto bool) bool {
	if s.state != levelingCombat || s.found != nil || s.chainNext != nil || s.goal != goalNone ||
		len(s.env.Encounter.ActiveTargetUnits) >= s.opts.maxPack || (auto && s.pullMobs >= s.opts.packSize) {
		return false
	}
	moving := sim.RandomFloat("Leveling Chain Search Moving") < s.opts.chainSearchMoving
	if moving {
		s.move(sim, goalSearch, 0, 0, nil)
	}
	if sim.Log != nil {
		s.player.Log(sim, "Leveling: looking for another mob, running: %v", moving)
	}
	s.after(sim, &s.chainNext, randomDuration(sim, "Leveling Chain Search", s.opts.chainSearchMin, s.opts.chainSearchMax), func(sim *Simulation) {
		if s.goal == goalSearch {
			s.move(sim, goalNone, 0, 0, nil)
		}
		if s.state != levelingCombat || s.find(sim, auto) == nil {
			return
		}
		s.pollPull(sim)
	})
	return true
}

// arrive: the mob reached us and starts swinging, its first swing anywhere in its timer.
func (s *levelingSession) arrive(sim *Simulation, mob *Unit) {
	m := &s.mobs[mob.Index]
	if m.state != mobRunning {
		return
	}
	m.state = mobArrived
	aa := &mob.AutoAttacks
	swing := DurationFromSeconds(aa.MH().SwingSpeed / mob.SwingSpeed())
	aa.mh.swingAt = sim.CurrentTime + time.Duration(sim.RandomFloat("Leveling First Swing")*float64(swing))
	aa.EnableAutoSwing(sim)
	if s.player.CurrentTarget == mob {
		s.setPlayerDistance(sim, levelingMeleeRange)
	}
}

// The leveling session's view of where things are, for spells, totems and the rotation.
// Outside a session everything is in reach: distances are 0.

// IsLeveling is whether this is a leveling session.
func (encounter *Encounter) IsLeveling() bool {
	return encounter.leveling != nil
}

// LevelingPosition is where we are on the line we move along, 0 outside a session. A totem
// keeps the position it was dropped at.
func (encounter *Encounter) LevelingPosition(sim *Simulation) float64 {
	if encounter.leveling == nil {
		return 0
	}
	return encounter.leveling.playerPosition(sim)
}

// LevelingDistanceFrom is how far the mob is from a point (a totem), 0 outside a session.
// A mob on us is where we are.
func (encounter *Encounter) LevelingDistanceFrom(sim *Simulation, point float64, mob *Unit) float64 {
	s := encounter.leveling
	if s == nil {
		return 0
	}
	return math.Abs(s.mobPosition(sim, mob) - point)
}

// LevelingDistance is how far the target is from us in a session, 0 outside one.
func (encounter *Encounter) LevelingDistance(sim *Simulation, target *Unit) float64 {
	if encounter.leveling == nil || target == nil || target.Type != EnemyUnit {
		return 0
	}
	return encounter.leveling.distance(sim, target)
}

// LevelingPullTarget is the mob we found and haven't pulled, nil when there isn't one or
// a spell on its way to it already pulls it.
func (encounter *Encounter) LevelingPullTarget() *Unit {
	s := encounter.leveling
	if s == nil || s.found == nil || s.mobs[s.found.Index].incoming {
		return nil
	}
	return s.found
}

// LevelingMobCounts are the mobs we pulled that are still coming and those on us.
func (encounter *Encounter) LevelingMobCounts() (running, onUs int) {
	if encounter.leveling == nil {
		return 0, len(encounter.ActiveTargetUnits)
	}
	for _, mob := range encounter.ActiveTargetUnits {
		switch encounter.leveling.mobs[mob.Index].state {
		case mobRunning:
			running++
		case mobArrived:
			onUs++
		}
	}
	return
}

// LevelingCanLookForMob is whether the rotation can look for another mob now: fighting,
// not already looking or walking, no mob found and not pulled yet, and room in the pack.
func (encounter *Encounter) LevelingCanLookForMob() bool {
	s := encounter.leveling
	return s != nil && s.state == levelingCombat && s.found == nil && s.chainNext == nil && s.goal == goalNone &&
		len(encounter.ActiveTargetUnits) < s.opts.maxPack
}

// LevelingLookForMob starts the rotation's search for another mob while fighting.
func (encounter *Encounter) LevelingLookForMob(sim *Simulation) {
	if encounter.LevelingCanLookForMob() {
		encounter.leveling.chainSearch(sim, false)
	}
}

// LevelingCanWalkToPullTarget is whether we can walk up to the Pull Target to within yards
// of it: there is one, we're not moving or casting, and it's further than that.
func (encounter *Encounter) LevelingCanWalkToPullTarget(sim *Simulation, within float64) bool {
	s := encounter.leveling
	mob := encounter.LevelingPullTarget()
	return mob != nil && s.goal == goalNone && !s.player.IsCasting(sim) &&
		s.distance(sim, mob) > max(within, levelingMeleeRange)
}

// LevelingWalkToPullTarget walks up to the Pull Target until we're within yards of it (0
// for melee range), or it notices us. The mobs on us follow.
func (encounter *Encounter) LevelingWalkToPullTarget(sim *Simulation, within float64) {
	if encounter.LevelingCanWalkToPullTarget(sim, within) {
		encounter.leveling.walkToFound(sim, within)
	}
}

// LevelingCanWalkTo is whether we can walk to within yards of a point (a totem): we're in a
// session, not moving or casting, and further away than that.
func (encounter *Encounter) LevelingCanWalkTo(sim *Simulation, point, within float64) bool {
	s := encounter.leveling
	return s != nil && s.goal == goalNone && !s.player.IsCasting(sim) &&
		(s.state == levelingPulling || s.state == levelingCombat) &&
		math.Abs(point-s.playerPosition(sim)) > within
}

// LevelingWalkTo walks to within yards of a point. The mobs on us follow, and the mob we
// found may notice us on the way.
func (encounter *Encounter) LevelingWalkTo(sim *Simulation, point, within float64) {
	if encounter.LevelingCanWalkTo(sim, point, within) {
		s := encounter.leveling
		if sim.Log != nil {
			s.player.Log(sim, "Leveling: walking %.0f yards back", math.Abs(point-s.playerPosition(sim))-within)
		}
		s.move(sim, goalPoint, point, within, nil)
	}
}
