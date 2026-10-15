package core

import (
	"math"
	"time"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
	googleProto "google.golang.org/protobuf/proto"
)

// A grinding session (Encounter.leveling, docs/leveling_sim.md): instead of one fight, we
// find a mob, pull it, maybe chain pull more, kill them, recover, and go again until the
// session ends. The score is how many mobs we kill.
//
// The mobs come from a pool built with the sim: max pack size normal mobs (the mob table,
// mob_stats.go) at each level of the spread, all of one creature type. A mob's level and type
// are fixed when the sim is built, because its attack tables and the talents and items that
// read its type are. A dead mob goes back to the pool and comes again in a later pull, with
// full health and its permanent auras back.
//
// Where things are: each mob has a distance from us, and we have a position along the way we
// run, so a totem stays where we dropped it while we move on (LevelingPosition).
//
// One loop runs the session:
//   - Searching: we run around for a random 5 to 11 sec (search_min / search_max), so we
//     can't hard cast, and find a mob standing 10 to 40 yards away (found_min / found_max).
//   - Pulling: the mob waits there, as our target and the rotation's Pull Target. A harmful
//     spell that lands on it pulls it, and spells only reach it within their range
//     (Spell.MaxRange). When the rotation hasn't pulled it a second after we found it and
//     isn't casting, we walk up to it, and it notices us in melee range (a face pull). The
//     rotation can still pull it on the way, with an instant once it's in range.
//   - Combat: a pulled mob runs to us at its level's speed. We can cast at it on the way,
//     but it can't hit us and we can't reach it in melee until it's here. Its first swing
//     comes at a random point of its swing timer. While the pull has taken fewer than
//     pack_size mobs, a search finds the next one (2 to 4 sec, running with the mobs on us following
//     three times out of four, when only instants and melee go out). It's pulled the same
//     way: the rotation's spell at the Pull Target, or we walk up to it with the mobs on us
//     in tow.
//   - When the last one dies, we're out of combat: a bandage when our health is low and
//     we're not Recently Bandaged, then we sit to eat and drink when health or mana is low,
//     until both are full or both buffs run out. Then the mob we found while fighting if
//     there is one, or a new search.
//   - Death: the mobs on us leave, and after the corpse run we come back with part of our
//     health and mana and search again.
//
// Out of combat we regen health from Spirit (outOfCombatHealthRegen), and mana goes on as in
// combat (the five-second rule). While no mob is on us the rotation can cast what helps us
// (buffs, shields, heals), but nothing harmful except at the mob it pulls.
type levelingSession struct {
	env  *Environment
	opts levelingOptions

	character *Character
	player    *Unit
	level     int32

	// The pool by level, and each target's state and position by Target.Index.
	pool   map[int32][]*Unit
	levels []int32
	mobs   []levelingMob

	// The mob we found and haven't pulled, out of combat or in.
	found   *Unit
	foundAt time.Duration

	// Our position along the way we run, and since when we've been moving. walkTo is the
	// found mob we're walking up to, if we are.
	position    float64
	moving      bool
	movingSince time.Duration
	walkTo      *Unit

	state      levelingState
	stateSince time.Duration
	seconds    [levelingStateCount]time.Duration

	// How many mobs the current pull has taken, and the session's pulls and mobs pulled.
	pullMobs   int
	pulls      int
	mobsPulled int
	// Food, water and bandage left on the current rest or bandage.
	foodLeft, drinkLeft, bandageLeft time.Duration
	food, drink, bandage             restorer
	bandageReadyAt                   time.Duration
	// A pending search, chain search or revive, cancelled when the state moves on.
	next      *PendingAction
	chainNext *PendingAction

	kills, xp float64
	deaths    int

	healthMetrics, foodMetrics, bandageMetrics, reviveHealthMetrics *ResourceMetrics
	drinkMetrics, reviveManaMetrics                                 *ResourceMetrics

	// Sums over iterations, for LevelingMetrics.
	agg struct {
		n                                int
		kills, kills2, xp, deaths        float64
		pulls, pulled                    float64
		combat, resting, searching, dead float64
	}
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
	// Distance from us at since: where a found mob stands (since we last walked up to it),
	// or where a running one started.
	distance float64
	since    time.Duration
	arrival  *PendingAction
	// A harmful spell at a found mob is on its way (a bolt in flight), so the mob is pulled
	// and no longer the rotation's Pull Target.
	incoming bool
}

type levelingState int

const (
	levelingSearching levelingState = iota
	levelingPulling
	levelingCombat
	levelingResting
	levelingBandaging
	levelingDead
	levelingStateCount
)

// levelingOptions are the session's options with their defaults filled in.
type levelingOptions struct {
	sessionLength                  time.Duration
	mobLevelOffset, mobSpread      int32
	mobType                        proto.MobType
	armorMultiplier                float64
	packSize, maxPack              int
	searchMin, searchMax           time.Duration
	chainSearchMin, chainSearchMax time.Duration
	chainSearchMoving              float64
	foundMin, foundMax             float64
	eatBelow, drinkBelow           float64
	bandages                       bool
	deathPenalty                   time.Duration
	reviveFraction                 float64
}

const (
	// Where a mob that reached us stands.
	levelingMeleeRange = 5
	// How long the rotation has to pull a mob we found before we walk up to it.
	levelingFacePullAfter = time.Second
	// How often the session looks at the pull: whether to walk up, how close we are.
	levelingPullPoll = 250 * time.Millisecond
)

func resolveLevelingOptions(o *proto.LevelingOptions) levelingOptions {
	or := func(v, def float64) float64 {
		if v > 0 {
			return v
		}
		return def
	}
	seconds := func(v, def float64) time.Duration { return DurationFromSeconds(or(v, def)) }
	opts := levelingOptions{
		sessionLength:     seconds(o.SessionLength, 1200),
		mobLevelOffset:    o.MobLevelOffset,
		mobSpread:         1,
		mobType:           o.MobType,
		armorMultiplier:   or(o.ArmorMultiplier, 1),
		packSize:          int(max(o.PackSize, 1)),
		maxPack:           int(max(o.MaxPackSize, 0)),
		searchMin:         seconds(o.SearchMin, 5),
		searchMax:         seconds(o.SearchMax, 11),
		chainSearchMin:    seconds(o.ChainSearchMin, 2),
		chainSearchMax:    seconds(o.ChainSearchMax, 4),
		chainSearchMoving: 0.75,
		foundMin:          math.Round(or(o.FoundMin, 10)),
		foundMax:          math.Round(or(o.FoundMax, 40)),
		eatBelow:          or(o.EatBelow, 0.5),
		drinkBelow:        or(o.DrinkBelow, 0.4),
		bandages:          !o.NoBandages,
		deathPenalty:      seconds(o.DeathPenalty, 180),
		reviveFraction:    or(o.ReviveFraction, 0.5),
	}
	if o.MobLevelSpread != nil {
		opts.mobSpread = max(*o.MobLevelSpread, 0)
	}
	if o.ChainSearchMoving != nil {
		opts.chainSearchMoving = min(max(*o.ChainSearchMoving, 0), 1)
	}
	if opts.mobType == proto.MobType_MobTypeUnknown {
		opts.mobType = proto.MobType_MobTypeBeast
	}
	if opts.maxPack == 0 {
		opts.maxPack = 5
	}
	opts.packSize = min(opts.packSize, opts.maxPack)
	opts.foundMax = max(opts.foundMax, opts.foundMin)
	return opts
}

// levelingPlayerLevel is the level of the raid's first player, the one who grinds.
func levelingPlayerLevel(raidProto *proto.Raid) int32 {
	for _, party := range raidProto.Parties {
		for _, player := range party.Players {
			if player != nil && player.Class != proto.Class_ClassUnknown {
				if player.Level == 0 {
					return CharacterMaxLevel
				}
				return player.Level
			}
		}
	}
	return CharacterMaxLevel
}

// prepareLevelingEncounter turns a leveling encounter into one the sim can build: a timed
// fight of the session's length with the mob pool as its targets. Any other encounter comes
// back as it was.
func prepareLevelingEncounter(raidProto *proto.Raid, encounterProto *proto.Encounter) *proto.Encounter {
	if encounterProto == nil || encounterProto.Leveling == nil {
		return encounterProto
	}
	opts := resolveLevelingOptions(encounterProto.Leveling)
	level := levelingPlayerLevel(raidProto)

	prepared := googleProto.Clone(encounterProto).(*proto.Encounter)
	prepared.Duration = opts.sessionLength.Seconds()
	prepared.DurationVariation = 0
	prepared.UseHealth = false
	prepared.Targets = nil
	for _, mobLevel := range levelingMobLevels(level, opts) {
		for range opts.maxPack {
			mob := NormalMobTarget(mobLevel, opts.mobType, opts.armorMultiplier)
			mob.TankIndex = -1 // We tank them all, which newLevelingSession sets.
			prepared.Targets = append(prepared.Targets, mob)
		}
	}
	return prepared
}

// levelingMobLevels are the mob levels of the session: the spread around our level plus the
// offset, kept within the mob table.
func levelingMobLevels(level int32, opts levelingOptions) []int32 {
	center := level + opts.mobLevelOffset
	var levels []int32
	for l := center - opts.mobSpread; l <= center+opts.mobSpread; l++ {
		if l >= 1 && l <= int32(len(normalMobStatsByLevel)) {
			levels = append(levels, l)
		}
	}
	if len(levels) == 0 {
		levels = append(levels, min(max(center, 1), int32(len(normalMobStatsByLevel))))
	}
	return levels
}

// newLevelingSession is built with the environment, once the raid is there.
func newLevelingSession(env *Environment, encounterProto *proto.Encounter) *levelingSession {
	session := &levelingSession{
		env:  env,
		opts: resolveLevelingOptions(encounterProto.Leveling),
		pool: map[int32][]*Unit{},
	}
	for _, unit := range env.Raid.AllPlayerUnits {
		if agent := env.Raid.GetPlayerFromUnit(unit); agent != nil {
			session.character = agent.GetCharacter()
			session.player = unit
			break
		}
	}
	session.level = session.player.Level
	session.levels = levelingMobLevels(session.level, session.opts)
	session.mobs = make([]levelingMob, len(env.Encounter.Targets))
	for _, target := range env.Encounter.TargetUnits {
		session.pool[target.Level] = append(session.pool[target.Level], target)
		target.CurrentTarget = session.player
	}

	session.food = bestRestorer(levelingFood, session.level)
	session.drink = bestRestorer(levelingWater, session.level)
	session.bandage = bestRestorer(levelingBandages, session.level)

	c := session.character
	session.healthMetrics = c.NewHealthMetrics(ActionID{SpellID: 20577, Tag: 1}) // Health regen out of combat.
	session.foodMetrics = c.NewHealthMetrics(ActionID{ItemID: session.food.itemID})
	session.bandageMetrics = c.NewHealthMetrics(ActionID{ItemID: session.bandage.itemID})
	session.reviveHealthMetrics = c.NewHealthMetrics(ActionID{SpellID: 8326}) // Ghost
	if c.HasManaBar() {
		session.drinkMetrics = c.NewManaMetrics(ActionID{ItemID: session.drink.itemID})
		session.reviveManaMetrics = c.NewManaMetrics(ActionID{SpellID: 8326})
	}
	return session
}

// reset is the start of an iteration, after every unit's own reset: every mob back in the
// pool and none near us. Before the sim's pull at 0, so no mob starts swinging.
func (s *levelingSession) reset(sim *Simulation) {
	enc := &s.env.Encounter
	for _, target := range enc.TargetUnits {
		target.dead = true
		target.enabled = false
		s.mobs[target.Index] = levelingMob{}
	}
	enc.ActiveTargetUnits = nil
	s.found, s.walkTo = nil, nil
	s.position, s.moving = 0, false
	s.player.MovementHandler.Moving = false
	s.player.DistanceFromTarget = s.opts.foundMax

	s.state, s.stateSince = levelingSearching, 0
	s.seconds = [levelingStateCount]time.Duration{}
	s.pullMobs, s.pulls, s.mobsPulled = 0, 0, 0
	s.kills, s.xp, s.deaths = 0, 0, 0
	s.bandageReadyAt = 0
	s.next, s.chainNext = nil, nil

	sim.AddPendingAction(&PendingAction{
		NextActionAt: 0,
		Priority:     ActionPriorityPrePull - 1,
		OnAction:     s.search,
	})
	s.startTicks(sim)
}

func (s *levelingSession) setState(sim *Simulation, state levelingState) {
	s.seconds[s.state] += sim.CurrentTime - s.stateSince
	s.state, s.stateSince = state, sim.CurrentTime
	cancel(sim, &s.next)
}

func cancel(sim *Simulation, pa **PendingAction) {
	if *pa != nil {
		(*pa).Cancel(sim)
		*pa = nil
	}
}

func (s *levelingSession) after(sim *Simulation, slot **PendingAction, delay time.Duration, fn func(*Simulation)) {
	cancel(sim, slot)
	// Before the rotation at the same moment, so it acts on what the session just did.
	pa := &PendingAction{NextActionAt: sim.CurrentTime + delay, Priority: ActionPriorityGCD + 1}
	pa.OnAction = func(sim *Simulation) {
		if *slot == pa {
			*slot = nil
		}
		fn(sim)
	}
	*slot = pa
	sim.AddPendingAction(pa)
}

func randomDuration(sim *Simulation, label string, lo, hi time.Duration) time.Duration {
	return lo + time.Duration(sim.RandomFloat(label)*float64(hi-lo))
}

// Moving: we run (searching, or walking up to a mob), so no hard casts. Mobs on us follow
// and keep swinging, and so do we.
func (s *levelingSession) setMoving(sim *Simulation, moving bool) {
	if s.moving == moving {
		return
	}
	if s.moving {
		if s.walkTo != nil {
			m := &s.mobs[s.walkTo.Index]
			m.distance, m.since = s.distance(sim, s.walkTo), sim.CurrentTime
			s.walkTo = nil
		}
		s.position += s.player.MovementHandler.MoveSpeed * (sim.CurrentTime - s.movingSince).Seconds()
	} else {
		s.movingSince = sim.CurrentTime
		if s.player.Hardcast.Expires > sim.CurrentTime {
			s.player.interruptHardcast(sim)
		}
	}
	s.moving = moving
	s.player.MovementHandler.Moving = moving
}

// playerPosition is where we are along the way we run.
func (s *levelingSession) playerPosition(sim *Simulation) float64 {
	if s.moving {
		return s.position + s.player.MovementHandler.MoveSpeed*(sim.CurrentTime-s.movingSince).Seconds()
	}
	return s.position
}

// distance is how far the mob is from us: where a found mob stands (closer as we walk up to
// it), how far a running one has still to come, and melee range for one on us.
func (s *levelingSession) distance(sim *Simulation, mob *Unit) float64 {
	m := &s.mobs[mob.Index]
	switch m.state {
	case mobFound:
		if mob == s.walkTo {
			return max(levelingMeleeRange, m.distance-s.player.MovementHandler.MoveSpeed*(sim.CurrentTime-m.since).Seconds())
		}
		return m.distance
	case mobRunning:
		speed := NormalMobStatsAt(mob.Level).RunSpeed
		return max(levelingMeleeRange, m.distance-speed*(sim.CurrentTime-m.since).Seconds())
	case mobArrived:
		return levelingMeleeRange
	}
	return math.Inf(1)
}

// search: out of combat and ready, we run around looking for the next mob.
func (s *levelingSession) search(sim *Simulation) {
	if s.found != nil {
		// We saw it while fighting, and it's still there.
		s.startPull(sim)
		return
	}
	s.setState(sim, levelingSearching)
	s.setMoving(sim, true)
	if sim.Log != nil {
		s.player.Log(sim, "Leveling: searching for a mob")
	}
	s.after(sim, &s.next, randomDuration(sim, "Leveling Search", s.opts.searchMin, s.opts.searchMax), func(sim *Simulation) {
		s.setMoving(sim, false)
		if s.find(sim) != nil {
			s.startPull(sim)
		} else {
			s.search(sim)
		}
	})
}

// find picks a mob from the pool to stand somewhere around us, waiting to be pulled.
func (s *levelingSession) find(sim *Simulation) *Unit {
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
	s.mobs[mob.Index] = levelingMob{state: mobFound, distance: distance, since: sim.CurrentTime}
	s.found, s.foundAt = mob, sim.CurrentTime
	if sim.Log != nil {
		s.player.Log(sim, "Leveling: found %s (level %d) %.0f yards away", mob.Label, mob.Level, distance)
	}
	return mob
}

// startPull: out of combat with a mob found. It's our target, and the rotation pulls it.
func (s *levelingSession) startPull(sim *Simulation) {
	s.setState(sim, levelingPulling)
	s.setMoving(sim, false)
	s.foundAt = sim.CurrentTime
	s.retarget(sim, s.found)
	s.pollPull(sim)
}

// pollPull watches the mob we found, out of combat or in: once we've walked into melee range
// of it, it's pulled, and when the rotation hasn't pulled it a second after we found it and
// isn't casting, we walk up to it.
func (s *levelingSession) pollPull(sim *Simulation) {
	mob := s.found
	if mob == nil || (s.state != levelingPulling && s.state != levelingCombat) {
		return
	}
	distance := s.distance(sim, mob)
	if s.mobs[mob.Index].incoming {
		return
	}
	if distance <= levelingMeleeRange {
		s.aggro(sim, mob)
		return
	}
	if s.walkTo == nil && !s.moving && sim.CurrentTime-s.foundAt >= levelingFacePullAfter && !s.player.IsCasting(sim) {
		if sim.Log != nil {
			s.player.Log(sim, "Leveling: walking up to %s", mob.Label)
		}
		s.setMoving(sim, true)
		s.walkTo = mob
		s.mobs[mob.Index].since = sim.CurrentTime
	}
	if mob == s.player.CurrentTarget {
		// Still out of melee range, so no swings to move.
		s.player.DistanceFromTarget = distance
	}
	poll := levelingPullPoll
	if s.walkTo == mob {
		poll = min(poll, DurationFromSeconds((distance-levelingMeleeRange)/s.player.MovementHandler.MoveSpeed))
	}
	s.after(sim, &s.next, max(poll, time.Millisecond), s.pollPull)
}

// onCast: a spell of ours went off at a mob. A harmful one at the mob we found pulls it
// when it lands (onHit), and until then the rotation shouldn't cast another. A spell that
// never lands (no damage, no miss) pulls it once its travel time is over.
func (s *levelingSession) onCast(sim *Simulation, mob *Unit, spell *Spell) {
	m := &s.mobs[mob.Index]
	if m.state != mobFound || m.incoming || !spell.attacksTarget() {
		return
	}
	m.incoming = true
	if s.walkTo == mob {
		s.setMoving(sim, false)
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

// onHit: something of ours landed on a mob. A harmful hit pulls the mob we found.
func (s *levelingSession) onHit(sim *Simulation, mob *Unit, spell *Spell) {
	if s.mobs[mob.Index].state == mobFound && !spell.Flags.Matches(SpellFlagHelpful) {
		s.aggro(sim, mob)
	}
}

// aggro: the mob is pulled and runs to us from where it stands.
func (s *levelingSession) aggro(sim *Simulation, mob *Unit) {
	distance := s.distance(sim, mob)
	if mob == s.found {
		s.found = nil
	}
	if s.walkTo == mob {
		s.setMoving(sim, false)
	}
	if s.state != levelingCombat {
		s.setState(sim, levelingCombat)
		s.setMoving(sim, false)
		s.pulls++
		s.pullMobs = 0
	}
	s.pullMobs++
	s.mobsPulled++
	s.mobs[mob.Index] = levelingMob{state: mobRunning, distance: distance, since: sim.CurrentTime}
	s.env.Encounter.addActiveTarget(mob)

	runIn := DurationFromSeconds(max(distance-levelingMeleeRange, 0) / NormalMobStatsAt(mob.Level).RunSpeed)
	if sim.Log != nil {
		s.player.Log(sim, "Leveling: pulled %s (level %d) from %.0f yards, here in %s", mob.Label, mob.Level, distance, runIn)
	}
	if mob == s.player.CurrentTarget {
		s.setPlayerDistance(sim, max(distance, levelingMeleeRange))
	}
	m := &s.mobs[mob.Index]
	if runIn <= 0 {
		s.arrive(sim, mob)
	} else {
		m.arrival = &PendingAction{NextActionAt: sim.CurrentTime + runIn, OnAction: func(sim *Simulation) { s.arrive(sim, mob) }}
		sim.AddPendingAction(m.arrival)
	}
	s.chainSearch(sim)
}

// chainSearch: while the pull has taken fewer than pack_size mobs, we look for the next one
// as we fight, running around with the mobs on us following most of the time.
func (s *levelingSession) chainSearch(sim *Simulation) {
	if s.state != levelingCombat || s.found != nil || s.chainNext != nil || s.pullMobs >= s.opts.packSize {
		return
	}
	moving := sim.RandomFloat("Leveling Chain Search Moving") < s.opts.chainSearchMoving
	if moving {
		s.setMoving(sim, true)
	}
	if sim.Log != nil {
		s.player.Log(sim, "Leveling: looking for another mob, running: %v", moving)
	}
	s.after(sim, &s.chainNext, randomDuration(sim, "Leveling Chain Search", s.opts.chainSearchMin, s.opts.chainSearchMax), func(sim *Simulation) {
		s.setMoving(sim, false)
		if s.state != levelingCombat || s.find(sim) == nil {
			return
		}
		s.pollPull(sim)
	})
}

func (s *levelingSession) idleMob(level int32) *Unit {
	for _, mob := range s.pool[level] {
		if s.mobs[mob.Index].state == mobInPool {
			return mob
		}
	}
	for _, l := range s.levels {
		for _, mob := range s.pool[l] {
			if s.mobs[mob.Index].state == mobInPool {
				return mob
			}
		}
	}
	return nil
}

// arrive: the mob reached us and starts swinging, its first swing anywhere in its timer.
func (s *levelingSession) arrive(sim *Simulation, mob *Unit) {
	m := &s.mobs[mob.Index]
	if m.state != mobRunning {
		return
	}
	m.state, m.arrival = mobArrived, nil
	aa := &mob.AutoAttacks
	swing := DurationFromSeconds(aa.MH().SwingSpeed / mob.SwingSpeed())
	aa.mh.swingAt = sim.CurrentTime + time.Duration(sim.RandomFloat("Leveling First Swing")*float64(swing))
	aa.EnableAutoSwing(sim)
	if s.player.CurrentTarget == mob {
		s.setPlayerDistance(sim, levelingMeleeRange)
	}
}

// retarget makes the mob our target, in melee range once it has reached us.
func (s *levelingSession) retarget(sim *Simulation, mob *Unit) {
	s.player.CurrentTarget = mob
	for _, pet := range s.character.Pets {
		pet.CurrentTarget = mob
	}
	s.setPlayerDistance(sim, s.distance(sim, mob))
}

func (s *levelingSession) setPlayerDistance(sim *Simulation, distance float64) {
	if math.IsInf(distance, 1) {
		distance = s.opts.foundMax
	}
	if s.player.DistanceFromTarget == distance {
		return
	}
	s.player.AutoAttacks.CancelAutoSwing(sim)
	s.player.DistanceFromTarget = distance
	s.player.AutoAttacks.EnableAutoSwing(sim)
}

// onMobDeath: a mob we pulled died (killTarget). It counts, our target has moved on to the
// next living one, and when it was the last one we're out of combat.
func (s *levelingSession) onMobDeath(sim *Simulation, mob *Unit) {
	s.kills++
	s.xp += mobKillXP(s.level, mob.Level)
	s.toPool(sim, mob)
	active := s.env.Encounter.ActiveTargetUnits
	if len(active) == 0 {
		cancel(sim, &s.chainNext)
		s.setMoving(sim, false)
		s.outOfCombat(sim)
		return
	}
	// killTarget moved our target to the next living mob, which may be one we haven't
	// pulled: we stay on the ones we have.
	if target := s.player.CurrentTarget; target == nil || !s.pulled(target) {
		s.retarget(sim, active[0])
	} else {
		s.retarget(sim, target)
	}
	s.chainSearch(sim)
}

// pulled is whether the mob is one we pulled, on its way or on us.
func (s *levelingSession) pulled(mob *Unit) bool {
	state := s.mobs[mob.Index].state
	return mob.Type == EnemyUnit && (state == mobRunning || state == mobArrived)
}

// toPool sends a mob back to the pool.
func (s *levelingSession) toPool(sim *Simulation, mob *Unit) {
	m := &s.mobs[mob.Index]
	if m.arrival != nil {
		m.arrival.Cancel(sim)
	}
	if s.walkTo == mob {
		s.setMoving(sim, false)
	}
	*m = levelingMob{}
	mob.enabled = false
	if s.found == mob {
		s.found = nil
	}
}

// outOfCombat: the pull is over. A bandage when health is low, then food and water when
// health or mana is, then the next pull.
func (s *levelingSession) outOfCombat(sim *Simulation) {
	health := s.player.CurrentHealthPercent()
	if s.opts.bandages && health < s.opts.eatBelow && sim.CurrentTime >= s.bandageReadyAt {
		s.setState(sim, levelingBandaging)
		s.bandageLeft = s.bandage.duration
		s.bandageReadyAt = sim.CurrentTime + recentlyBandagedDuration
		s.busy(sim, s.bandage.duration)
		if sim.Log != nil {
			s.player.Log(sim, "Leveling: bandaging")
		}
		s.after(sim, &s.next, s.bandage.duration, s.outOfCombat)
		return
	}

	mana := 1.0
	if s.player.HasManaBar() {
		mana = s.player.CurrentManaPercent()
	}
	if health < s.opts.eatBelow || mana < s.opts.drinkBelow {
		s.setState(sim, levelingResting)
		s.foodLeft, s.drinkLeft = 0, 0
		if health < 1 {
			s.foodLeft = s.food.duration
		}
		if s.player.HasManaBar() && mana < 1 {
			s.drinkLeft = s.drink.duration
		}
		if sim.Log != nil {
			s.player.Log(sim, "Leveling: eating and drinking at %.0f%% health and %.0f%% mana", 100*health, 100*mana)
		}
		s.busy(sim, time.Second)
		return
	}

	s.search(sim)
}

// busy keeps the rotation from acting while we sit, bandage or run back as a ghost.
func (s *levelingSession) busy(sim *Simulation, d time.Duration) {
	if s.player.Hardcast.Expires > sim.CurrentTime {
		s.player.interruptHardcast(sim)
	}
	s.player.WaitUntil(sim, sim.CurrentTime+d)
}

// startTicks starts the session's 1 sec tick: health regen out of combat, food, water and
// bandage, and standing up when we're done resting.
func (s *levelingSession) startTicks(sim *Simulation) {
	StartPeriodicAction(sim, PeriodicActionOptions{
		Period: time.Second,
		OnAction: func(sim *Simulation) {
			s.tick(sim, time.Second)
		},
	})
}

func (s *levelingSession) tick(sim *Simulation, dt time.Duration) {
	p := s.player
	seconds := dt.Seconds()
	switch s.state {
	case levelingSearching, levelingPulling, levelingResting, levelingBandaging:
		regen := outOfCombatHealthRegen(s.character.Class, p.GetStat(stats.Spirit)) * seconds
		if s.state == levelingResting {
			regen *= 1.5
		}
		s.heal(sim, regen, s.healthMetrics)
	}

	switch s.state {
	case levelingBandaging:
		if s.bandageLeft > 0 {
			s.heal(sim, s.bandage.amount*seconds/s.bandage.duration.Seconds(), s.bandageMetrics)
			s.bandageLeft -= dt
		}
	case levelingResting:
		if s.foodLeft > 0 {
			s.heal(sim, s.food.amount*seconds/s.food.duration.Seconds(), s.foodMetrics)
			s.foodLeft -= dt
		}
		if s.drinkLeft > 0 && p.HasManaBar() {
			p.AddMana(sim, min(s.drink.amount*seconds/s.drink.duration.Seconds(), p.MaxMana()-p.CurrentMana()), s.drinkMetrics)
			s.drinkLeft -= dt
		}
		healthDone := s.foodLeft <= 0 || p.CurrentHealth() >= p.MaxHealth()
		manaDone := s.drinkLeft <= 0 || !p.HasManaBar() || p.CurrentMana() >= p.MaxMana()
		if healthDone && manaDone {
			s.search(sim)
		} else {
			s.busy(sim, dt)
		}
	}
}

func (s *levelingSession) heal(sim *Simulation, amount float64, metrics *ResourceMetrics) {
	amount = min(amount, s.player.MaxHealth()-s.player.CurrentHealth())
	if amount > 0 {
		s.player.GainHealth(sim, amount, metrics)
	}
}

// onPlayerDeath: the mobs on us leave, and we run back as a ghost.
func (s *levelingSession) onPlayerDeath(sim *Simulation) {
	if s.state == levelingDead {
		return
	}
	s.deaths++
	s.setState(sim, levelingDead)
	cancel(sim, &s.chainNext)
	s.setMoving(sim, false)
	if sim.Log != nil {
		s.player.Log(sim, "Leveling: died, back in %s", s.opts.deathPenalty)
	}
	enc := &s.env.Encounter
	for len(enc.ActiveTargetUnits) > 0 {
		s.despawn(sim, enc.ActiveTargetUnits[0])
	}
	if s.found != nil {
		s.despawn(sim, s.found)
	}
	s.setPlayerDistance(sim, s.opts.foundMax)
	s.busy(sim, s.opts.deathPenalty)
	s.after(sim, &s.next, s.opts.deathPenalty, s.revive)
}

// despawn sends a mob back to the pool without a kill.
func (s *levelingSession) despawn(sim *Simulation, mob *Unit) {
	mob.dead = true
	mob.AutoAttacks.CancelAutoSwing(sim)
	mob.auraTracker.expireAll(sim)
	s.env.Encounter.removeActiveTarget(mob)
	s.toPool(sim, mob)
}

func (s *levelingSession) revive(sim *Simulation) {
	p := s.player
	target := s.opts.reviveFraction * p.MaxHealth()
	if p.CurrentHealth() < target {
		p.GainHealth(sim, target-p.CurrentHealth(), s.reviveHealthMetrics)
	}
	if p.HasManaBar() && p.CurrentMana() < s.opts.reviveFraction*p.MaxMana() {
		p.AddMana(sim, s.opts.reviveFraction*p.MaxMana()-p.CurrentMana(), s.reviveManaMetrics)
	}
	p.Metrics.Died = true
	s.search(sim)
}

// doneIteration adds the session to the totals: the mobs on us at the end count by the share
// of their health we took.
func (s *levelingSession) doneIteration(sim *Simulation) {
	s.seconds[s.state] += sim.CurrentTime - s.stateSince
	s.stateSince = sim.CurrentTime
	s.setMoving(sim, false)
	kills, xp := s.kills, s.xp
	for _, mob := range s.env.Encounter.ActiveTargetUnits {
		share := 1 - mob.CurrentHealth()/mob.MaxHealth()
		kills += share
		xp += share * mobKillXP(s.level, mob.Level)
	}

	a := &s.agg
	a.n++
	a.kills += kills
	a.kills2 += kills * kills
	a.xp += xp
	a.deaths += float64(s.deaths)
	a.pulls += float64(s.pulls)
	a.pulled += float64(s.mobsPulled)
	a.combat += s.seconds[levelingCombat].Seconds()
	a.resting += (s.seconds[levelingResting] + s.seconds[levelingBandaging]).Seconds()
	a.searching += (s.seconds[levelingSearching] + s.seconds[levelingPulling]).Seconds()
	a.dead += s.seconds[levelingDead].Seconds()
}

func (s *levelingSession) metrics() *proto.LevelingMetrics {
	a := &s.agg
	if a.n == 0 {
		return &proto.LevelingMetrics{}
	}
	n := float64(a.n)
	mean := a.kills / n
	m := &proto.LevelingMetrics{
		MobsKilled:       mean,
		MobsKilledStdev:  math.Sqrt(max(a.kills2/n-mean*mean, 0)),
		Xp:               a.xp / n,
		XpPerHour:        a.xp / n * 3600 / s.opts.sessionLength.Seconds(),
		Deaths:           a.deaths / n,
		SecondsInCombat:  a.combat / n,
		SecondsResting:   a.resting / n,
		SecondsSearching: a.searching / n,
		SecondsDead:      a.dead / n,
	}
	if a.pulls > 0 {
		m.MobsPerPull = a.pulled / a.pulls
	}
	return m
}

// The leveling session's view of where things are, for spells and totems (0 and in reach of
// everything outside a session).

// IsLeveling is whether this is a leveling session.
func (encounter *Encounter) IsLeveling() bool {
	return encounter.leveling != nil
}

// LevelingPosition is where we are along the way we run in a leveling session, 0 outside
// one. A totem keeps the position it was dropped at.
func (encounter *Encounter) LevelingPosition(sim *Simulation) float64 {
	if encounter.leveling == nil {
		return 0
	}
	return encounter.leveling.playerPosition(sim)
}

// LevelingDistanceFrom is how far the mob is from a point we stood at (a totem), 0 outside a
// leveling session. A mob on us is where we are, and one still on its way or waiting to be
// pulled is that much further, taken on the far side.
func (encounter *Encounter) LevelingDistanceFrom(sim *Simulation, point float64, mob *Unit) float64 {
	s := encounter.leveling
	if s == nil {
		return 0
	}
	return math.Abs(s.playerPosition(sim)-point) + max(0, s.distance(sim, mob)-levelingMeleeRange)
}

// LevelingDistance is how far the target is from us in a leveling session, 0 outside one.
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
