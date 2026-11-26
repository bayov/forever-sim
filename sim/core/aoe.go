package core

import "math"

// An ability that hits more than one target is one of three kinds (docs/leveling_sim.md, and
// the Forever mechanics auditor's AoE scaling check, orchestration/aoe_scaling.md):
//
//   - Hard capped: it hits at most N targets (Cleave, Chain Lightning's jumps). Its
//     ApplyEffects takes AoETargetCount(N) targets.
//   - Soft capped: it hits every target, but each one takes less past N targets. Its damage
//     takes AoESoftCapMultiplier.
//   - Unlimited: every living target (ActiveTargetUnits) at full damage.
//
// All three count only the living targets, so a pack hurts less as its mobs die. A leveling
// session's hard capped ones also only reach the mobs near the one they go at (inArea).

// AoETargetCount is how many targets an ability that hits at most maxTargets of them reaches
// now: the living ones, at most maxTargets (0 for no limit).
func (encounter *Encounter) AoETargetCount(maxTargets int32) int32 {
	n := int32(len(encounter.ActiveTargetUnits))
	if encounter.leveling != nil {
		n = encounter.leveling.areaTargetCount()
	}
	if maxTargets > 0 {
		n = min(n, maxTargets)
	}
	return n
}

// TargetsNear are the enemies within yards of center, for an ability without a cap that blasts an
// area. Outside a leveling session that's every living target. In one it's the mobs we pulled, and
// center itself.
func (encounter *Encounter) TargetsNear(sim *Simulation, center *Unit, yards float64) []*Unit {
	s := encounter.leveling
	if s == nil {
		return encounter.ActiveTargetUnits
	}
	here := s.playerPosition(sim)
	if center != nil && center.Type == EnemyUnit {
		here = s.mobPosition(sim, center)
	}
	near := FilterSlice(encounter.ActiveTargetUnits, func(mob *Unit) bool {
		return s.mobPosition(sim, mob).Dist(here) <= yards+levelingYardTolerance
	})
	if center != nil && center.Type == EnemyUnit && s.mobs[center.Index].state == mobFound {
		near = append(near, center)
	}
	return near
}

// AoESoftCapRule is how a soft capped ability's damage drops past its cap.
type AoESoftCapRule int

const (
	// The damage of cap targets split over all of them: cap / targets each. Classic's
	// 10 target cap (Encounter.AOECapMultiplier).
	AoESoftCapSplit AoESoftCapRule = iota
	// sqrt(cap / targets) each, like modern WoW's AoE abilities "reduced beyond N targets".
	AoESoftCapSqrt
)

// AoESoftCapMultiplier is each target's damage multiplier when a soft capped ability hits
// targets of them: 1 up to the cap, less past it.
func AoESoftCapMultiplier(targets, cap int32, rule AoESoftCapRule) float64 {
	if cap <= 0 || targets <= cap {
		return 1
	}
	share := float64(cap) / float64(targets)
	if rule == AoESoftCapSqrt {
		return math.Sqrt(share)
	}
	return share
}
