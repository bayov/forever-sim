package core

import (
	"container/heap"
	"math"
	"time"

	"github.com/wowsims/classic/sim/core/proto"
)

type MoveModifier struct {
	ActionId *ActionID
	Modifier float64
}

type MoveHeap []MoveModifier

func (h MoveHeap) Len() int           { return len(h) }
func (h MoveHeap) Less(i, j int) bool { return h[i].Modifier > h[j].Modifier }
func (h MoveHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *MoveHeap) Push(x any) {
	*h = append(*h, x.(MoveModifier))
}

func (h *MoveHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h *MoveHeap) activeModifier() float64 {
	heap := *h
	if len(heap) < 1 {
		panic("Movement Heaps should never be missing their base case!!!")
	}
	n := heap[0]
	return n.Modifier
}

func (h *MoveHeap) Find(actionId *ActionID) int {
	for i, mod := range *h {
		if mod.ActionId == actionId {
			return i
		}
	}
	return -1
}

func (move *MovementHandler) addMoveSpeedModifier(moveHeap *MoveHeap, moveMod MoveModifier) {
	heap.Push(moveHeap, moveMod)
	move.updateMoveSpeed()
}

func (move *MovementHandler) removeMoveSpeedModifier(moveHeap *MoveHeap, actionID *ActionID) {
	index := moveHeap.Find(actionID)
	if index == -1 {
		return
	}
	heap.Remove(moveHeap, index)
	move.updateMoveSpeed()
}

func (move *MovementHandler) updateMoveSpeed() {
	move.MoveSpeed = move.baseSpeed * move.getActveModifier(move.moveSpeedBonuses) * (1 - move.getActveModifier(move.moveSpeedPenalties))
}

func (move *MovementHandler) getActveModifier(moveHeap *MoveHeap) float64 {
	return moveHeap.activeModifier()
}

type MovementHandler struct {
	Moving    bool
	MoveSpeed float64

	baseSpeed          float64
	moveAura           *Aura
	moveSpell          *Spell
	moveSpeedBonuses   *MoveHeap
	moveSpeedPenalties *MoveHeap
}

func (unit *Unit) initMovement() {
	unit.MovementHandler = &MovementHandler{
		moveSpeedBonuses: &MoveHeap{
			MoveModifier{
				Modifier: 1,
			},
		},
		moveSpeedPenalties: &MoveHeap{
			MoveModifier{
				Modifier: 0,
			},
		},
		baseSpeed: 7.0,
	}
	unit.MovementHandler.updateMoveSpeed()

	unit.MovementHandler.moveAura = unit.GetOrRegisterAura(Aura{
		Label:     "Movement",
		ActionID:  ActionID{OtherID: proto.OtherAction_OtherActionMove},
		Duration:  NeverExpires,
		MaxStacks: 30,

		OnGain: func(aura *Aura, sim *Simulation) {
			if unit.IsChanneling(sim) {
				unit.ChanneledDot.Cancel(sim)
			}
			unit.AutoAttacks.CancelAutoSwing(sim)
			unit.MovementHandler.Moving = true
		},
		OnExpire: func(aura *Aura, sim *Simulation) {
			unit.MovementHandler.Moving = false
			unit.AutoAttacks.EnableAutoSwing(sim)

			// Simulate the delay from starting attack. Under Forever only a walk into melee range
			// starts it: one that ends out of range has no attack to start.
			if !sim.IsForever() || unit.DistanceFromTarget <= MaxMeleeAttackDistance {
				unit.AutoAttacks.DelayMeleeBy(sim, time.Millisecond*50)
			}
			unit.AutoAttacks.resumeRangedAfterMove(sim)
		},
	})

	unit.MovementHandler.moveSpell = unit.GetOrRegisterSpell(SpellConfig{
		ActionID: ActionID{OtherID: proto.OtherAction_OtherActionMove},
		Flags:    SpellFlagMeleeMetrics,

		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			unit.MovementHandler.moveAura.Activate(sim)
			unit.MovementHandler.moveAura.SetStacks(sim, int32(unit.DistanceFromTarget))
		},
	})
}

func (unit *Unit) IsMoving() bool {
	return unit.MovementHandler.Moving
}

func (unit *Unit) MoveTo(moveRange float64, sim *Simulation) {
	if moveRange == unit.DistanceFromTarget {
		return
	}

	moveDistance := moveRange - unit.DistanceFromTarget
	moveTicks, partial := unit.MovementHandler.moveSteps(math.Abs(moveDistance))
	moveInterval := math.Copysign(1, moveDistance)

	unit.MovementHandler.moveSpell.Cast(sim, unit.CurrentTarget)

	// A walk ends where it was going, a fraction of a yard past its last whole yard included.
	arrive := func(sim *Simulation) {
		unit.DistanceFromTarget = moveRange
		unit.MovementHandler.moveAura.SetStacks(sim, int32(unit.DistanceFromTarget))
		unit.MovementHandler.moveAura.Deactivate(sim)
	}
	arriveAfter := func(sim *Simulation, delay time.Duration) {
		if delay <= 0 {
			arrive(sim)
			return
		}
		sim.AddPendingAction(&PendingAction{NextActionAt: sim.CurrentTime + delay, OnAction: arrive})
	}
	if moveTicks == 0 {
		arriveAfter(sim, partial)
		return
	}

	ticks := 0
	sim.AddPendingAction(NewPeriodicAction(sim, PeriodicActionOptions{
		Period:          unit.MovementHandler.yardTime(),
		NumTicks:        moveTicks,
		TickImmediately: false,

		OnAction: func(sim *Simulation) {
			unit.DistanceFromTarget += moveInterval
			unit.MovementHandler.moveAura.SetStacks(sim, int32(unit.DistanceFromTarget))

			if ticks++; ticks == moveTicks {
				arriveAfter(sim, partial)
			}
		},
	}))
}

// How long one yard takes at the current speed, in whole milliseconds.
func (move *MovementHandler) yardTime() time.Duration {
	return time.Millisecond * time.Duration(1000/move.MoveSpeed)
}

// A walk of the given yards takes whole yards at yardTime each, then the fraction left over at
// the same pace.
func (move *MovementHandler) moveSteps(yards float64) (int, time.Duration) {
	whole := math.Floor(yards)
	if fraction := yards - whole; fraction > 1-1e-9 {
		whole++
	}
	fraction := max(0, yards-whole)
	if fraction < 1e-9 {
		fraction = 0
	}
	return int(whole), time.Duration(fraction * float64(move.yardTime()))
}

// How long MoveTo takes from where the unit stands to the given range, at the current speed.
func (unit *Unit) TimeToMoveTo(moveRange float64) time.Duration {
	ticks, partial := unit.MovementHandler.moveSteps(math.Abs(moveRange - unit.DistanceFromTarget))
	return time.Duration(ticks)*unit.MovementHandler.yardTime() + partial
}

// A move speed increase of 30% should be represented as 1.30 and a move speed slow of 70% should be respresented as 0.70
func (unit *Unit) AddMoveSpeedModifier(actionId *ActionID, modifier float64) {
	moveSpeedMod := MoveModifier{
		ActionId: actionId,
		Modifier: modifier,
	}
	if moveSpeedMod.Modifier < 1 {
		unit.MovementHandler.addMoveSpeedModifier(unit.MovementHandler.moveSpeedPenalties, moveSpeedMod)
	} else {
		unit.MovementHandler.addMoveSpeedModifier(unit.MovementHandler.moveSpeedBonuses, moveSpeedMod)
	}

}

func (unit *Unit) RemoveMoveSpeedModifier(actionID *ActionID) {
	unit.MovementHandler.removeMoveSpeedModifier(unit.MovementHandler.moveSpeedPenalties, actionID)
	unit.MovementHandler.removeMoveSpeedModifier(unit.MovementHandler.moveSpeedBonuses, actionID)
}
