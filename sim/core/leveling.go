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
// Where things are: we, the mobs and our totems have positions on a 2D plane
// (leveling_world.go).
//
// One loop runs the session:
//   - Searching: we run around for a random 2 to 7 sec (search_min / search_max), so we
//     can't hard cast, and find a mob standing 10 to 40 yards away (found_min / found_max).
//     One that close to its aggro radius notices us at once.
//   - Pulling: the mob waits there, as our target and the rotation's Pull Target. A spell
//     that attacks it pulls it when it lands, and spells only reach it within their range
//     (Spell.MaxRange). The rotation can walk toward it (Walk To Pull Target). When it hasn't
//     pulled it, isn't casting and hasn't moved for a second, we walk up to it until it
//     notices us (a face pull). The rotation can still pull it on the way with an instant.
//   - Combat: a pulled mob runs at us at its level's speed, following us if we move. We can
//     cast at it on the way, but it can't hit us and we can't reach it in melee until it's
//     here. Its first swing comes at a random point of its swing timer. The rotation decides
//     whether to look for another mob (Look For Mob: 2 to 4 sec, running with the mobs on us
//     following three times out of four, when only instants and melee go out), and pulls it
//     the same ways, or leaves it for our next pull.
//   - When the last one dies, we're out of combat and ready (leveling_rest.go): the rotation
//     can bandage, sit to eat and drink, heal or buff. Once it has nothing to do, the mob we
//     found while fighting if there is one, or a new search.
//   - Death: the mobs on us leave, and after the corpse run we come back with part of our
//     health and mana and search again.
//
// Out of combat we regen health from Spirit (outOfCombatHealthRegen), a third more sitting,
// and mana goes on as in combat (the five-second rule). While no mob is on us the rotation can cast what helps us
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

	// Where we are (leveling_world.go): our position at posAt, moving at vel yards a sec,
	// toward a goal, and when we last stopped. goalReached ends the walk, proximity is the
	// mob we found noticing us, and chaseTurn turns the mobs running at us while we move.
	pos, vel, goalPoint               Vec2
	goalWithin                        float64
	posAt, stoppedAt                  time.Duration
	goal                              levelingGoal
	goalMob                           *Unit
	goalReached, proximity, chaseTurn *PendingAction

	state      levelingState
	stateSince time.Duration
	seconds    [levelingStateCount]time.Duration

	// How many mobs the current pull has taken, and the session's pulls and mobs pulled.
	pullMobs   int
	pulls      int
	mobsPulled int
	// The current rest (leveling_rest.go), with the food, water and bandage left on it.
	rest                             levelingRest
	foodLeft, drinkLeft, bandageLeft time.Duration
	food, drink, bandage             restorer
	bandageReadyAt                   time.Duration
	// Food, water and bandages used this session.
	foods, drinks, bandages int
	// A pending search, chain search or revive, cancelled when the state moves on.
	next      *PendingAction
	chainNext *PendingAction

	kills, xp float64
	deaths    int
	// The last session's score, mobs on us at the end counted by the share of their health
	// we took, for the iteration picks (lastIteration).
	lastKills float64

	healthMetrics, foodMetrics, bandageMetrics, reviveHealthMetrics *ResourceMetrics
	drinkMetrics, reviveManaMetrics                                 *ResourceMetrics

	// Sums over iterations, for LevelingMetrics.
	agg struct {
		n                                int
		kills, kills2, xp, deaths        float64
		pulls, pulled                    float64
		combat, resting, searching, dead float64
		foods, drinks, bandages          float64
	}
}

type levelingState int

const (
	levelingSearching levelingState = iota
	levelingPulling
	levelingCombat
	levelingResting
	levelingBandaging
	levelingDead
	// Out of combat and deciding: the rotation can rest, bandage or buff before we search.
	levelingReady
	levelingStateCount
)

// levelingOptions are the session's options with their defaults filled in.
type levelingOptions struct {
	sessionLength                  time.Duration
	mobLevelOffset, mobSpread      int32
	mobType                        proto.MobType
	armorMultiplier                float64
	maxPack                        int
	searchMin, searchMax           time.Duration
	chainSearchMin, chainSearchMax time.Duration
	chainSearchMoving              float64
	foundMin, foundMax             float64
	deathPenalty                   time.Duration
	reviveFraction                 float64
}

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
		maxPack:           int(max(o.MaxPackSize, 0)),
		searchMin:         seconds(o.SearchMin, 2),
		searchMax:         seconds(o.SearchMax, 7),
		chainSearchMin:    seconds(o.ChainSearchMin, 2),
		chainSearchMax:    seconds(o.ChainSearchMax, 4),
		chainSearchMoving: 0.75,
		foundMin:          math.Round(or(o.FoundMin, 10)),
		foundMax:          math.Round(or(o.FoundMax, 40)),
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
	// Execute phases go by time over the whole encounter, which says nothing about each mob's
	// health in a session, so they never start; and a session isn't a PvP fight.
	prepared.ExecuteProportion_20 = 0
	prepared.ExecuteProportion_25 = 0
	prepared.ExecuteProportion_35 = 0
	prepared.Pvp = false
	prepared.PvpMeleeDowntime = 0
	// The mobs we fight hit us themselves; there is no raid.
	prepared.RaidDamageHitsPerMinute = 0
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
	s.found = nil
	s.pos, s.vel, s.posAt, s.stoppedAt = Vec2{}, Vec2{}, 0, 0
	s.goal, s.goalMob, s.goalReached, s.proximity, s.chaseTurn = goalNone, nil, nil, nil, nil
	s.player.MovementHandler.Moving = false
	s.player.DistanceFromTarget = s.opts.foundMax

	s.state, s.stateSince = levelingSearching, 0
	s.seconds = [levelingStateCount]time.Duration{}
	s.pullMobs, s.pulls, s.mobsPulled = 0, 0, 0
	s.kills, s.xp, s.deaths = 0, 0, 0
	s.bandageReadyAt = 0
	s.rest = levelingRest{}
	s.foods, s.drinks, s.bandages = 0, 0, 0
	s.next, s.chainNext = nil, nil

	sim.AddPendingAction(&PendingAction{
		NextActionAt: 0,
		Priority:     ActionPriorityPrePull - 1,
		OnAction:     s.search,
	})
	s.startTicks(sim)
}

// levelingStateNames are the states as the "Leveling: now ..." log lines say them, for the
// results' session strip.
var levelingStateNames = [levelingStateCount]string{
	levelingSearching: "searching",
	levelingPulling:   "pulling",
	levelingCombat:    "fighting",
	levelingResting:   "resting",
	levelingBandaging: "bandaging",
	levelingDead:      "dead",
	levelingReady:     "ready",
}

func (s *levelingSession) setState(sim *Simulation, state levelingState) {
	s.seconds[s.state] += sim.CurrentTime - s.stateSince
	if sim.Log != nil && (state != s.state || sim.CurrentTime == 0) {
		s.player.Log(sim, "Leveling: now %s", levelingStateNames[state])
	}
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
		s.move(sim, goalNone, Vec2{}, 0, nil)
		if sim.Log != nil {
			s.player.Log(sim, "Leveling: out of combat")
		}
		// killTarget moved our target to the next living mob: the one we found, if there
		// is one, still where it stands, out of reach of a swing.
		if s.found != nil {
			s.retarget(sim, s.found)
		} else {
			s.setPlayerDistance(sim, s.opts.foundMax)
		}
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
	if s.goal == goalMob && s.goalMob == mob {
		s.move(sim, goalNone, Vec2{}, 0, nil)
	}
	*m = levelingMob{}
	mob.enabled = false
	if s.found == mob {
		s.found = nil
	}
}

// outOfCombat: the pull is over, and the rotation decides what's next (ready).
func (s *levelingSession) outOfCombat(sim *Simulation) {
	s.ready(sim)
}

func (s *levelingSession) manaPercent() float64 {
	if !s.player.HasManaBar() {
		return 1
	}
	return s.player.CurrentManaPercent()
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
	regen := outOfCombatHealthRegen(s.character.Class, s.level, p.GetStat(stats.Spirit)) * seconds
	if s.character.Race == proto.Race_RaceTroll {
		regen *= trollHealthRegenMultiplier
	}
	switch s.state {
	case levelingSearching, levelingPulling, levelingResting, levelingBandaging:
		if s.state == levelingResting {
			regen *= sittingHealthRegenMultiplier
		}
		s.heal(sim, regen, s.healthMetrics)
	case levelingCombat:
		if s.character.Race == proto.Race_RaceTroll {
			s.heal(sim, regen*trollCombatHealthRegen, s.healthMetrics)
		}
	}

	switch s.state {
	case levelingBandaging:
		if s.bandageLeft > 0 {
			s.heal(sim, s.bandage.amount*seconds/s.bandage.duration.Seconds(), s.bandageMetrics)
			s.bandageLeft -= dt
		}
	case levelingResting:
		s.restTick(sim, dt)
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
	s.move(sim, goalNone, Vec2{}, 0, nil)
	if sim.Log != nil {
		s.player.Log(sim, "Leveling: died, back in %s", s.opts.deathPenalty)
	}
	for _, effect := range s.character.levelingDeathEffects {
		effect(sim)
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
	if sim.Log != nil {
		s.player.Log(sim, "Leveling: %s left", mob.Label)
	}
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
	s.ready(sim)
}

// doneIteration adds the session to the totals: the mobs on us at the end count by the share
// of their health we took.
func (s *levelingSession) doneIteration(sim *Simulation) {
	s.seconds[s.state] += sim.CurrentTime - s.stateSince
	s.stateSince = sim.CurrentTime
	kills, xp := s.kills, s.xp
	for _, mob := range s.env.Encounter.ActiveTargetUnits {
		share := 1 - mob.CurrentHealth()/mob.MaxHealth()
		kills += share
		xp += share * mobKillXP(s.level, mob.Level)
	}
	s.lastKills = kills

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
	a.searching += (s.seconds[levelingSearching] + s.seconds[levelingPulling] + s.seconds[levelingReady]).Seconds()
	a.foods += float64(s.foods)
	a.drinks += float64(s.drinks)
	a.bandages += float64(s.bandages)
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
		Food:             a.foods / n,
		Drinks:           a.drinks / n,
		Bandages:         a.bandages / n,
	}
	if a.pulls > 0 {
		m.MobsPerPull = a.pulled / a.pulls
	}
	return m
}
