package core

import (
	"math"
	"sort"

	"github.com/wowsims/classic/sim/core/proto"
)

// The tank score (Player.tank_scoring, docs/tank_sim.md, The optimizer's objective): what the tank
// optimizer weighs, threat against survival.
//
//   - Survival: each iteration's least stand-in HPS the tank lives through. Without
//     Encounter.track_health the damage the tank takes and the stand-in healer's timing don't depend
//     on the healer's HPS, only the heals' sizes do (each is the HPS times the gap since the last).
//     So the iteration's damage and heals are recorded and replayed at any HPS after it ends. The
//     share of iterations whose least HPS is over h is the chance of death at h, and healing needed
//     is a percentile of them (TankScoreSummary). It holds only while nothing the tank does reads
//     its health: a health potion, a defensive at a health share, an APL condition on health.
//   - Threat: SafeDPS, the steady DPS a reference DPS member could do without passing the tank on
//     the boss at any moment. The member starts dps_start seconds after the pull and makes
//     threat_per_damage threat a point of damage, so at time t it has threat_per_damage × DPS ×
//     (t − dps_start), and it passes at pull_ratio times the tank's threat. Threat only grows when
//     it's booked, so the tightest moment is just before a booking or at the end.
type tankScore struct {
	dpsStart        float64
	threatPerDamage float64
	pullRatio       float64

	// The replay, when the tank has the stand-in healer and no healing fight.
	replay        bool
	hps           float64
	streamMetrics *ResourceMetrics
	startHealth   float64
	events        []tankReplayEvent

	// This iteration's threat on the boss since the pull, and the lowest SafeDPS so far.
	threat  float64
	safeMin float64

	// Over the iterations.
	survive  []float64
	safe     DistributionMetrics
	safeEach []float64
}

// tankReplayEvent is damage taken, a heal of a fixed size, or a stand-in heal of units times the
// HPS, each capped at the max health then. Damage carries the max health too: when max health drops
// (Last Stand ending), the live run cuts current health to it (unit.go, processDynamicBonus), so the
// replay does the same before taking the damage.
type tankReplayEvent struct {
	damage    float64
	heal      float64
	units     float64
	maxHealth float64
}

// setupTankScore gives a tank the boss attacks its score, when the player asks for one. It comes
// after trackChanceOfDeath, which sets up the stand-in healer.
func (character *Character) setupTankScore(scoring *proto.TankScoring, healingModel *proto.HealingModel) {
	if scoring == nil || !character.Metrics.isTanking {
		return
	}
	ts := &tankScore{
		dpsStart:        scoring.DpsStart,
		threatPerDamage: scoring.ThreatPerDamage,
		pullRatio:       scoring.PullRatio,
		safe:            NewDistributionMetrics(),
	}
	if ts.dpsStart == 0 {
		ts.dpsStart = 2
	}
	if ts.threatPerDamage == 0 {
		// 1 threat a point of damage: Forever removed Blessing of Salvation (the October 30 hotfix,
		// and the November 2 notes), which was 0.7.
		ts.threatPerDamage = 1
	}
	if ts.pullRatio == 0 {
		ts.pullRatio = MeleePullRatio
	}
	if healingModel != nil && healingModel.Hps > 0 && character.healingModelMetrics != nil &&
		!character.Env.Encounter.TrackHealth && character.Env.Encounter.leveling == nil {
		ts.replay = true
		ts.hps = healingModel.Hps
		ts.streamMetrics = character.healingModelMetrics
	}
	character.Unit.tankScore = ts
}

// readsNoHealth is whether the unit is scored by a tank score's replay, where nothing may read its
// health: the score replays the iteration at other HPS, so a choice made on this run's health
// wouldn't follow (docs/tank_sim.md).
func (unit *Unit) readsNoHealth() bool {
	return unit.tankScore != nil && unit.tankScore.replay
}

func (ts *tankScore) reset() {
	ts.events = ts.events[:0]
	ts.threat = 0
	ts.safeMin = math.Inf(1)
}

// startPull is the health the iteration starts the fight with, after the fight's permanent auras.
func (ts *tankScore) startPull(unit *Unit) {
	ts.startHealth = unit.CurrentHealth()
}

func (ts *tankScore) damageTaken(unit *Unit, amount float64) {
	if ts.replay {
		ts.events = append(ts.events, tankReplayEvent{damage: amount, maxHealth: unit.MaxHealth()})
	}
}

func (ts *tankScore) healTaken(unit *Unit, amount float64, metrics *ResourceMetrics) {
	if !ts.replay {
		return
	}
	if metrics == ts.streamMetrics {
		ts.events = append(ts.events, tankReplayEvent{units: amount / ts.hps, maxHealth: unit.MaxHealth()})
	} else {
		ts.events = append(ts.events, tankReplayEvent{heal: amount, maxHealth: unit.MaxHealth()})
	}
}

// checkSafe takes the SafeDPS of this moment, with the tank's threat so far.
func (ts *tankScore) checkSafe(t float64) {
	if t > ts.dpsStart {
		ts.safeMin = min(ts.safeMin, ts.pullRatio*ts.threat/(ts.threatPerDamage*(t-ts.dpsStart)))
	}
}

func (ts *tankScore) threatBooked(sim *Simulation, threat float64) {
	ts.checkSafe(sim.CurrentTime.Seconds())
	ts.threat += threat
}

func (ts *tankScore) doneIteration(sim *Simulation) {
	ts.checkSafe(sim.Duration.Seconds())
	if math.IsInf(ts.safeMin, 1) {
		ts.safeMin = 0
	}
	// DistributionMetrics divides by the fight length.
	ts.safe.Total = ts.safeMin * sim.Duration.Seconds()
	ts.safe.doneIteration(sim)
	ts.safeEach = append(ts.safeEach, ts.safeMin)
	if ts.replay {
		ts.survive = append(ts.survive, ts.leastHps())
	}
}

// lives replays the iteration at the HPS: whether the tank's health stays over 0.
func (ts *tankScore) lives(hps float64) bool {
	health := ts.startHealth
	for _, e := range ts.events {
		switch {
		case e.damage > 0:
			health = min(e.maxHealth, health) - e.damage
			if health <= 0 {
				return false
			}
		case e.units > 0:
			health = min(e.maxHealth, health+hps*e.units)
		case e.heal > 0:
			health = min(e.maxHealth, health+e.heal)
		}
	}
	return true
}

// leastHps is the least stand-in HPS the iteration lives through, to a millionth, +Inf when every
// heal topping the tank off still isn't enough.
func (ts *tankScore) leastHps() float64 {
	if ts.lives(0) {
		return 0
	}
	if !ts.lives(math.Inf(1)) {
		return math.Inf(1)
	}
	lo, hi := 0.0, max(ts.hps, 1)
	for !ts.lives(hi) {
		lo, hi = hi, 2*hi
	}
	for hi-lo > 1e-6*hi {
		mid := (lo + hi) / 2
		if ts.lives(mid) {
			hi = mid
		} else {
			lo = mid
		}
	}
	return hi
}

func (ts *tankScore) toProto() *proto.TankScore {
	return &proto.TankScore{
		SurviveHps:  append([]float64(nil), ts.survive...),
		SafeDps:     ts.safe.ToProto(),
		SafeDpsEach: append([]float64(nil), ts.safeEach...),
	}
}

// TankScoreSummary is healing needed, the stand-in HPS at which the tank dies in at most death of
// the iterations, with its standard error from the order statistics, and the death floor, the
// share no healing saves. Healing needed is -1 when the floor is over death.
func TankScoreSummary(score *proto.TankScore, death float64) (needed, stdErr, floor float64) {
	values := append([]float64(nil), score.GetSurviveHps()...)
	n := len(values)
	if n == 0 {
		return 0, 0, 0
	}
	sort.Float64s(values)
	for _, v := range values {
		if math.IsInf(v, 1) {
			floor++
		}
	}
	floor /= float64(n)
	if floor > death {
		return -1, 0, floor
	}
	// The (1 − death) quantile, and the order statistics 1.96 binomial standard errors either side
	// for its 95% interval.
	at := func(k float64) float64 {
		i := int(math.Ceil(k)) - 1
		return values[max(0, min(n-1, i))]
	}
	q := (1 - death) * float64(n)
	needed = at(q)
	spread := 1.96 * math.Sqrt(float64(n)*death*(1-death))
	stdErr = (at(q+spread) - at(q-spread)) / (2 * 1.96)
	return needed, stdErr, floor
}
