package core

import (
	"math"
	"time"
)

// Where things are in a leveling session (leveling.go): a flat 2D plane, in yards. We have a
// position on it, and so does every mob out of the pool and every totem (LevelingPosition).
//
//   - Each search we run off in a random direction, and we walk to a mob or to a point (a
//     totem) when the session or the rotation says so. While we move no hard cast goes out
//     (Moving), but we keep swinging at a mob on us.
//   - The mob we find stands in a random direction from us, 10 to 40 yards away after a
//     search we run for (found_min / found_max) and 20 to 60 after one we stand for
//     (found_standing_min / found_standing_max).
//   - A pulled mob runs at us at its level's speed and reaches us 5 yards away (melee range).
//     While we move it keeps turning toward where we are, every 0.1 sec. Once it has reached
//     us it stays on us wherever we go.
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
	// How often a mob running at us turns toward where we are while we move.
	levelingChaseTurn = 100 * time.Millisecond
	// Distances closer than this to a range are in it: a walk that stops at 30 yards ends a
	// rounding error away from it.
	levelingYardTolerance = 1e-6
)

// Vec2 is a point or a movement on the leveling session's plane, in yards.
type Vec2 struct {
	X, Y float64
}

func (v Vec2) Add(o Vec2) Vec2         { return Vec2{v.X + o.X, v.Y + o.Y} }
func (v Vec2) Sub(o Vec2) Vec2         { return Vec2{v.X - o.X, v.Y - o.Y} }
func (v Vec2) Scale(k float64) Vec2    { return Vec2{v.X * k, v.Y * k} }
func (v Vec2) Dot(o Vec2) float64      { return v.X*o.X + v.Y*o.Y }
func (v Vec2) Len() float64            { return math.Hypot(v.X, v.Y) }
func (v Vec2) Dist(o Vec2) float64     { return v.Sub(o).Len() }
func (v Vec2) IsZero() bool            { return v.X == 0 && v.Y == 0 }
func vec2FromAngle(angle float64) Vec2 { return Vec2{math.Cos(angle), math.Sin(angle)} }
func (v Vec2) unit() Vec2              { return v.Scale(1 / v.Len()) }
func (v Vec2) at(d time.Duration) Vec2 { return v.Scale(d.Seconds()) }

// toward is k yards from v toward o.
func (v Vec2) toward(o Vec2, k float64) Vec2 {
	if v == o {
		return v
	}
	return v.Add(o.Sub(v).unit().Scale(k))
}

// whenWithin is the first time from now, in seconds, that a point at rel (from us) moving
// at relVel comes within r of us, and false if it never does.
func whenWithin(rel, relVel Vec2, r float64) (float64, bool) {
	c := rel.Dot(rel) - r*r
	if c <= 0 {
		return 0, true
	}
	a := relVel.Dot(relVel)
	b := rel.Dot(relVel)
	if a == 0 || b >= 0 {
		return 0, false
	}
	disc := b*b - a*c
	if disc < 0 {
		return 0, false
	}
	return (-b - math.Sqrt(disc)) / a, true
}

type levelingMobState int

const (
	mobInPool levelingMobState = iota
	mobFound
	mobRunning
	mobArrived
)

type levelingMob struct {
	state levelingMobState
	// Position at xAt: where a found mob stands, or a running one was, running along dir.
	x, dir  Vec2
	xAt     time.Duration
	arrival *PendingAction
	// A spell that attacks it is on its way (a bolt in flight): it's pulled, and no longer
	// the rotation's Pull Target.
	incoming bool
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

func (s *levelingSession) playerPosition(sim *Simulation) Vec2 {
	return s.pos.Add(s.vel.at(sim.CurrentTime - s.posAt))
}

func mobRunSpeed(mob *Unit) float64 {
	return NormalMobStatsAt(mob.Level).RunSpeed
}

// mobPosition is where the mob is; one on us is where we are.
func (s *levelingSession) mobPosition(sim *Simulation, mob *Unit) Vec2 {
	m := &s.mobs[mob.Index]
	switch m.state {
	case mobFound:
		return m.x
	case mobRunning:
		return m.x.Add(m.dir.Scale(mobRunSpeed(mob) * (sim.CurrentTime - m.xAt).Seconds()))
	}
	return s.playerPosition(sim)
}

// distance is how far the mob is from us: melee range once it has reached us, and never
// closer while it runs in.
func (s *levelingSession) distance(sim *Simulation, mob *Unit) float64 {
	switch s.mobs[mob.Index].state {
	case mobFound:
		return s.mobPosition(sim, mob).Dist(s.playerPosition(sim))
	case mobRunning:
		return max(levelingMeleeRange, s.mobPosition(sim, mob).Dist(s.playerPosition(sim)))
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
// mobs running at us turn toward us, and the mob we found may notice us sooner or later.
func (s *levelingSession) setVelocity(sim *Simulation, vel Vec2) {
	if vel == s.vel {
		return
	}
	s.pos, s.posAt = s.playerPosition(sim), sim.CurrentTime
	if s.vel.IsZero() && s.player.Hardcast.Expires > sim.CurrentTime {
		s.player.interruptHardcast(sim)
	}
	s.vel = vel
	s.logPosition(sim)
	s.player.MovementHandler.Moving = !vel.IsZero()
	s.chase(sim)
	if s.found != nil && s.found == s.player.CurrentTarget {
		s.player.DistanceFromTarget = s.distance(sim, s.found)
	}
	s.scheduleProximity(sim)
}

// move sets off toward a goal, or stops with goalNone.
func (s *levelingSession) move(sim *Simulation, goal levelingGoal, toward Vec2, within float64, mob *Unit) {
	s.goal, s.goalPoint, s.goalWithin, s.goalMob = goal, toward, within, mob
	cancel(sim, &s.goalReached)
	if goal == goalNone {
		s.setVelocity(sim, Vec2{})
		s.stoppedAt = sim.CurrentTime
		return
	}
	speed := s.player.MovementHandler.MoveSpeed
	if goal == goalSearch {
		heading := vec2FromAngle(2 * math.Pi * sim.RandomFloat("Leveling Search Heading"))
		s.setVelocity(sim, heading.Scale(speed))
		return
	}
	here := s.playerPosition(sim)
	gap := toward.Dist(here)
	if gap <= within {
		s.move(sim, goalNone, Vec2{}, 0, nil)
		return
	}
	s.setVelocity(sim, toward.Sub(here).unit().Scale(speed))
	s.after(sim, &s.goalReached, DurationFromSeconds((gap-within)/speed), func(sim *Simulation) {
		from := s.playerPosition(sim)
		s.move(sim, goalNone, Vec2{}, 0, nil)
		// Exactly there, not a rounding error short of a spell's range.
		s.pos = toward.toward(from, within)
		s.logPosition(sim)
		s.scheduleProximity(sim)
	})
}

// chase brings the mobs running at us up to now, turns them toward us, and times their
// arrival. While we move they turn again every levelingChaseTurn.
func (s *levelingSession) chase(sim *Simulation) {
	cancel(sim, &s.chaseTurn)
	p := s.playerPosition(sim)
	running := false
	for _, mob := range s.env.Encounter.ActiveTargetUnits {
		m := &s.mobs[mob.Index]
		if m.state != mobRunning {
			continue
		}
		cancel(sim, &m.arrival)
		m.x, m.xAt = s.mobPosition(sim, mob), sim.CurrentTime
		rel := m.x.Sub(p)
		if rel.Len() <= levelingMeleeRange {
			s.after(sim, &m.arrival, 0, func(sim *Simulation) { s.arrive(sim, mob) })
			continue
		}
		m.dir = rel.Scale(-1 / rel.Len())
		s.logMobPosition(sim, mob)
		t, ok := whenWithin(rel, m.dir.Scale(mobRunSpeed(mob)).Sub(s.vel), levelingMeleeRange)
		if ok && (s.vel.IsZero() || t <= levelingChaseTurn.Seconds()) {
			s.after(sim, &m.arrival, DurationFromSeconds(t), func(sim *Simulation) { s.arrive(sim, mob) })
		} else {
			running = true
		}
	}
	if running && !s.vel.IsZero() {
		s.after(sim, &s.chaseTurn, levelingChaseTurn, s.chase)
	}
}

// scheduleProximity times when the mob we found notices us, if our path takes us within its
// aggro radius.
func (s *levelingSession) scheduleProximity(sim *Simulation) {
	cancel(sim, &s.proximity)
	mob := s.found
	if mob == nil || s.mobs[mob.Index].incoming {
		return
	}
	t, ok := whenWithin(s.mobs[mob.Index].x.Sub(s.playerPosition(sim)), s.vel.Scale(-1), s.aggroRadius(mob))
	if !ok {
		return
	}
	s.after(sim, &s.proximity, DurationFromSeconds(t), func(sim *Simulation) {
		if s.found == mob {
			if sim.Log != nil {
				s.player.Log(sim, "Leveling: %s noticed us", mob.Label)
			}
			s.aggro(sim, mob)
		}
	})
}

// logPosition logs where we are and how we move, for the replay's map. We move in straight
// lines between these.
func (s *levelingSession) logPosition(sim *Simulation) {
	if sim.Log != nil {
		s.player.Log(sim, "Leveling: we're at (%.2f, %.2f), moving (%.2f, %.2f)", s.pos.X, s.pos.Y, s.vel.X, s.vel.Y)
	}
}

// logMobPosition logs where a mob we found or pulled is and how it moves, for the replay's
// map. A running mob turns toward us now and then, and each turn logs a new line.
func (s *levelingSession) logMobPosition(sim *Simulation, mob *Unit) {
	if sim.Log == nil {
		return
	}
	m := &s.mobs[mob.Index]
	vel := Vec2{}
	if m.state == mobRunning {
		vel = m.dir.Scale(mobRunSpeed(mob))
	}
	s.player.Log(sim, "Leveling: %s at (%.2f, %.2f), moving (%.2f, %.2f)", mob.Label, m.x.X, m.x.Y, vel.X, vel.Y)
}

// search: out of combat and ready, we run around looking for the next mob.
func (s *levelingSession) search(sim *Simulation) {
	if s.found != nil {
		// We saw it while fighting, and it's still there.
		s.startPull(sim)
		return
	}
	s.setState(sim, levelingSearching)
	s.move(sim, goalSearch, Vec2{}, 0, nil)
	if sim.Log != nil {
		s.player.Log(sim, "Leveling: searching for a mob")
	}
	s.after(sim, &s.next, randomDuration(sim, "Leveling Search", s.opts.searchMin, s.opts.searchMax), func(sim *Simulation) {
		s.move(sim, goalNone, Vec2{}, 0, nil)
		if s.find(sim, false) != nil {
			s.startPull(sim)
		} else {
			s.search(sim)
		}
	})
}

// find picks a mob from the pool to stand in a random direction from us, waiting to be
// pulled. A mob we find while standing is further away.
func (s *levelingSession) find(sim *Simulation, stood bool) *Unit {
	level := s.levels[int(sim.RandomFloat("Leveling Mob Level")*float64(len(s.levels)))%len(s.levels)]
	mob := s.idleMob(level)
	if mob == nil {
		return nil
	}
	lo, hi := s.opts.foundMin, s.opts.foundMax
	if stood {
		lo, hi = s.opts.standingFoundMin, s.opts.standingFoundMax
	}
	distance := math.Round(lo + sim.RandomFloat("Leveling Mob Distance")*(hi-lo))
	direction := vec2FromAngle(2 * math.Pi * sim.RandomFloat("Leveling Mob Direction"))
	mob.dead = false
	mob.enabled = true
	mob.healthBar.currentHealth = mob.MaxHealth()
	mob.auraTracker.expireAll(sim) // Anything cast at it while it lay in the pool.
	mob.auraTracker.respawn(sim)
	s.mobs[mob.Index] = levelingMob{state: mobFound, x: s.playerPosition(sim).Add(direction.Scale(distance)), xAt: sim.CurrentTime}
	s.found, s.foundAt = mob, sim.CurrentTime
	s.logMobPosition(sim, mob)
	if sim.Log != nil {
		s.player.Log(sim, "Leveling: found %s (level %d) %.0f yards away, bearing %.0f°", mob.Label, mob.Level, distance, math.Mod(math.Atan2(direction.Y, direction.X)*180/math.Pi+360, 360))
	}
	s.scheduleProximity(sim)
	return mob
}

// startPull: out of combat with a mob found. It's our target, and the rotation pulls it.
func (s *levelingSession) startPull(sim *Simulation) {
	s.setState(sim, levelingPulling)
	s.move(sim, goalNone, Vec2{}, 0, nil)
	s.foundAt = sim.CurrentTime
	s.retarget(sim, s.found)
	s.pollPull(sim)
}

// pollPull watches the mob we found out of combat: when the rotation hasn't pulled it, isn't
// casting and hasn't moved for a second, we walk up to it until it notices us. That keeps a
// rotation without pull lines grinding; one with them pulls or walks up within the second. In
// combat the rotation looked for the mob, and it decides what to do with it.
func (s *levelingSession) pollPull(sim *Simulation) {
	mob := s.found
	if mob == nil || s.state != levelingPulling || s.mobs[mob.Index].incoming {
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
		s.move(sim, goalNone, Vec2{}, 0, nil)
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
		s.move(sim, goalNone, Vec2{}, 0, nil)
	}
	if s.state != levelingCombat {
		s.setState(sim, levelingCombat)
		if s.goal == goalSearch {
			s.move(sim, goalNone, Vec2{}, 0, nil)
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
	s.chase(sim)
	if sim.Log != nil {
		if arrival := s.mobs[mob.Index].arrival; arrival != nil {
			s.player.Log(sim, "Leveling: pulled %s (level %d) from %.0f yards, here in %s", mob.Label, mob.Level, distance, arrival.NextActionAt-sim.CurrentTime)
		} else {
			s.player.Log(sim, "Leveling: pulled %s (level %d) from %.0f yards, chasing us", mob.Label, mob.Level, distance)
		}
	}
}

// chainSearch: while fighting, the rotation looks for another mob (LevelingLookForMob). It
// runs around for it with the mobs on us following, or stands and keeps fighting while it
// looks, which takes longer.
func (s *levelingSession) chainSearch(sim *Simulation, stand bool) {
	lo, hi := s.opts.chainSearchMin, s.opts.chainSearchMax
	if stand {
		lo, hi = s.opts.standingSearchMin, s.opts.standingSearchMax
	} else {
		s.move(sim, goalSearch, Vec2{}, 0, nil)
	}
	if sim.Log != nil {
		s.player.Log(sim, "Leveling: looking for another mob, running: %v", !stand)
	}
	s.after(sim, &s.chainNext, randomDuration(sim, "Leveling Chain Search", lo, hi), func(sim *Simulation) {
		if s.goal == goalSearch {
			s.move(sim, goalNone, Vec2{}, 0, nil)
		}
		if s.state == levelingCombat {
			s.find(sim, stand)
		}
	})
}

// arrive: the mob reached us and starts swinging, its first swing anywhere in its timer.
func (s *levelingSession) arrive(sim *Simulation, mob *Unit) {
	m := &s.mobs[mob.Index]
	if m.state != mobRunning {
		return
	}
	m.state = mobArrived
	if sim.Log != nil {
		s.player.Log(sim, "Leveling: %s reached us", mob.Label)
	}
	aa := &mob.AutoAttacks
	swing := DurationFromSeconds(aa.MH().SwingSpeed / mob.SwingSpeed())
	aa.mh.swingAt = sim.CurrentTime + time.Duration(sim.RandomFloat("Leveling First Swing")*float64(swing))
	aa.EnableAutoSwing(sim)
	if s.player.CurrentTarget == mob {
		s.setPlayerDistance(sim, levelingMeleeRange)
	}
}

// The leveling session's view of where things are, for spells, totems and the rotation.
// Outside a session everything is in reach: every position is the origin.

// IsLeveling is whether this is a leveling session.
func (encounter *Encounter) IsLeveling() bool {
	return encounter.leveling != nil
}

// LevelingPosition is where we are on the plane, the origin outside a session. A totem
// keeps the position it was dropped at.
func (encounter *Encounter) LevelingPosition(sim *Simulation) Vec2 {
	if encounter.leveling == nil {
		return Vec2{}
	}
	return encounter.leveling.playerPosition(sim)
}

// A totem doesn't drop on our feet but somewhere around us, this many yards away.
const (
	levelingTotemDropMin = 1.5
	levelingTotemDropMax = 3.5
)

// LevelingTotemOffset is where a totem we drop now lands, from where we stand: a few yards
// away in a random direction. Outside a session totems stand on us, so it's zero.
func (encounter *Encounter) LevelingTotemOffset(sim *Simulation) Vec2 {
	if encounter.leveling == nil {
		return Vec2{}
	}
	angle := 2 * math.Pi * sim.RandomFloat("Leveling Totem Drop Angle")
	yards := levelingTotemDropMin + (levelingTotemDropMax-levelingTotemDropMin)*sim.RandomFloat("Leveling Totem Drop Distance")
	return Vec2{yards * math.Cos(angle), yards * math.Sin(angle)}
}

// LevelingDistanceFrom is how far the mob is from a point (a totem), 0 outside a session.
// A mob on us is where we are.
func (encounter *Encounter) LevelingDistanceFrom(sim *Simulation, point Vec2, mob *Unit) float64 {
	s := encounter.leveling
	if s == nil {
		return 0
	}
	return s.mobPosition(sim, mob).Dist(point)
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

// LevelingMobsThisPull is how many mobs the current pull has taken, the dead ones too: 0 out of
// combat.
func (encounter *Encounter) LevelingMobsThisPull() int {
	s := encounter.leveling
	if s == nil || s.state != levelingCombat {
		return 0
	}
	return s.pullMobs
}

// LevelingCanLookForMob is whether the rotation can look for another mob now: fighting,
// not already looking or walking, no mob found and not pulled yet, and room in the pack.
func (encounter *Encounter) LevelingCanLookForMob() bool {
	s := encounter.leveling
	return s != nil && s.state == levelingCombat && s.found == nil && s.chainNext == nil && s.goal == goalNone &&
		len(encounter.ActiveTargetUnits) > 0 && len(encounter.ActiveTargetUnits) < s.opts.maxPack
}

// LevelingLookForMob starts the rotation's search for another mob while fighting, running
// around for it, or standing and fighting while we look.
func (encounter *Encounter) LevelingLookForMob(sim *Simulation, stand bool) {
	if encounter.LevelingCanLookForMob() {
		encounter.leveling.chainSearch(sim, stand)
	}
}

// LevelingCanWalkToPullTarget is whether we can walk up to the Pull Target to within yards
// of it: there is one, we're not moving or casting, and it's further than that.
func (encounter *Encounter) LevelingCanWalkToPullTarget(sim *Simulation, within float64) bool {
	s := encounter.leveling
	mob := encounter.LevelingPullTarget()
	return mob != nil && s.goal == goalNone && !s.sitting() && !s.player.IsCasting(sim) &&
		s.distance(sim, mob) > max(within, levelingMeleeRange)+levelingYardTolerance
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
func (encounter *Encounter) LevelingCanWalkTo(sim *Simulation, point Vec2, within float64) bool {
	s := encounter.leveling
	return s != nil && s.goal == goalNone && !s.player.IsCasting(sim) &&
		(s.state == levelingPulling || s.state == levelingCombat) &&
		point.Dist(s.playerPosition(sim)) > within+levelingYardTolerance
}

// LevelingWalkTo walks to within yards of a point. The mobs on us follow, and the mob we
// found may notice us on the way.
func (encounter *Encounter) LevelingWalkTo(sim *Simulation, point Vec2, within float64) {
	if encounter.LevelingCanWalkTo(sim, point, within) {
		s := encounter.leveling
		if sim.Log != nil {
			s.player.Log(sim, "Leveling: walking %.0f yards toward a point", point.Dist(s.playerPosition(sim))-within)
		}
		s.move(sim, goalPoint, point, within, nil)
	}
}
