package core

import (
	"math"
	"time"

	"github.com/wowsims/classic/sim/core/stats"
)

// Which mobs are behind us in a leveling session (docs/leveling_sim.md, "Big pulls"), from 38 of
// our own combat logs, which hold where both of us stood and looked at every swing (2026-11-25).
//
// The positions in leveling_world.go are too coarse for it, since every mob that reached us is
// where we are. So each swing rolls for it:
//
//   - While we run looking ahead, the mobs on us are behind: 95% of the swings of a lone mob, and 70
//     to 82% with two mobs or more. We don't swing or use melee abilities at them then (0.03 swings
//     a second in the logs).
//   - But half of the swings at someone running came from in front, because people strafe with the
//     mobs at their side and keep swinging (0.2 a second, 0.3 when they stand). So half of our runs
//     go like that (run_forward), as if we stood.
//   - While we stand, up to three mobs stay in front: 5% came from behind, the same with one mob
//     as with three, as we turned or stepped. Past that it grows with each mob, to a quarter.
//
// A swing from behind can't be dodged, parried or blocked (1,334 such swings in the logs had 0.4%
// to 0.5% of each, where 9.0%, 4.2% and 1.0% came from in front). One that hurts can daze us.
const (
	behindRunningOne  = 0.95
	behindRunningMore = 0.75
	// The share from behind that each mob past the third adds while we stand, and the most.
	behindPerMob  = 0.04
	behindMost    = 0.25
	mobsInFront   = 3
	dazeDuration  = 4 * time.Second
	dazeMoveSpeed = 0.5
	// Our distance to our target while we run with it on us: out of reach of a swing.
	levelingRunningDistance = levelingMeleeRange + 1
)

// dazeChance is the chance that a mob's melee hit from behind dazes us: 20% at our own level, and
// 0.65% a level under 30, then 0.16% more for each point of the mob's weapon skill over our
// Defense, at most 40% (cmangos Unit::CalculateEffectiveDazeChance). The logs have 1% at level 5,
// 5% at 10, 16% at 18 and 13 to 21% from 30 on.
func dazeChance(level, mobLevel int32, bonusDefense float64) float64 {
	chance := 20.0
	if level < 30 {
		chance = 0.65*float64(level) + 0.5
	}
	chance += (float64(5*mobLevel) - float64(5*level) - bonusDefense) * 0.16
	return min(max(chance, 0), 40) / 100
}

// behindChance is the chance that a swing of a mob on us comes from behind.
func (s *levelingSession) behindChance() float64 {
	if s.opts.ignoreFacing {
		return 0
	}
	running := s.runningAhead()
	// Most swings are in a fight with few mobs on us, so we needn't count them.
	if !running && (len(s.env.Encounter.ActiveTargetUnits) <= mobsInFront || s.opts.surround == 0) {
		return 0
	}
	onUs := s.mobsOnUs()
	if running {
		if onUs <= 1 {
			return behindRunningOne
		}
		return behindRunningMore
	}
	return min(behindPerMob*float64(max(onUs-mobsInFront, 0)), behindMost) * s.opts.surround
}

// runningAhead is whether we run with our back to the mobs on us.
func (s *levelingSession) runningAhead() bool {
	return s.runForward && !s.vel.IsZero()
}

// startRun rolls which way we look on a run that starts now.
func (s *levelingSession) startRun(sim *Simulation) {
	s.runForward = !s.opts.ignoreFacing && sim.RandomFloat("Leveling Run Forward") < s.opts.runForward
}

func (s *levelingSession) mobsOnUs() int {
	onUs := 0
	for _, mob := range s.env.Encounter.ActiveTargetUnits {
		if s.mobs[mob.Index].state == mobArrived {
			onUs++
		}
	}
	return onUs
}

// swingFromBehind rolls whether a mob's swing at us comes from behind.
func (s *levelingSession) swingFromBehind(sim *Simulation) bool {
	chance := s.behindChance()
	s.hitBehind = chance > 0 && sim.RandomFloat("Leveling Swing From Behind") < chance
	return s.hitBehind
}

// onHitFromBehind: a mob's swing from behind hit us, and may daze us.
func (s *levelingSession) onHitFromBehind(sim *Simulation, mob *Unit, damage float64) {
	s.hitBehind = false
	if damage <= 0 || s.player.CurrentHealth() <= 0 {
		return
	}
	if sim.RandomFloat("Leveling Daze") < dazeChance(s.level, mob.Level, s.player.stats[stats.Defense]) {
		s.dazed.Activate(sim)
	}
}

func (s *levelingSession) registerDaze() {
	actionID := ActionID{SpellID: 1604}
	s.dazed = s.character.RegisterAura(Aura{
		Label:    "Dazed",
		ActionID: actionID,
		Duration: dazeDuration,
		OnGain: func(_ *Aura, sim *Simulation) {
			s.player.AddMoveSpeedModifier(&s.dazed.ActionID, dazeMoveSpeed)
			s.speedChanged(sim)
		},
		OnExpire: func(_ *Aura, sim *Simulation) {
			s.player.RemoveMoveSpeedModifier(&s.dazed.ActionID)
			s.speedChanged(sim)
		},
	})
}

// speedChanged has a run we're on go on at our new speed. A Charge has a speed of its own, so it
// goes on as it was.
func (s *levelingSession) speedChanged(sim *Simulation) {
	if s.vel.IsZero() || math.Abs(s.vel.Len()-s.runSpeed) > 1e-6 {
		return
	}
	switch s.goal {
	case goalSearch:
		s.runSpeed = s.player.MovementHandler.MoveSpeed
		s.setVelocity(sim, s.vel.unit().Scale(s.runSpeed))
	case goalMob, goalPoint:
		s.move(sim, s.goal, s.goalPoint, s.goalWithin, s.goalMob)
	}
}

// targetDistance is how far our target is for our swings and melee abilities: a mob on us is out
// of their reach while we run from it looking ahead.
func (s *levelingSession) targetDistance(sim *Simulation, mob *Unit) float64 {
	if s.runningAhead() && s.mobs[mob.Index].state == mobArrived {
		return levelingRunningDistance
	}
	return s.distance(sim, mob)
}
