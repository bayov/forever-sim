package core

import (
	"time"
)

// Note that this is only used when the hardcast and GCD actions happen at different times.
func (unit *Unit) newHardcastAction(sim *Simulation) {
	if unit.hardcastAction != nil && !unit.hardcastAction.consumed {
		unit.hardcastAction.Cancel(sim)
		unit.hardcastAction = nil
	}

	if unit.hardcastAction == nil {
		pa := &PendingAction{
			NextActionAt: unit.Hardcast.Expires,
			OnAction: func(sim *Simulation) {
				if hc := &unit.Hardcast; hc.Expires != startingCDTime && !unit.IsCasting(sim) {
					hc.Expires = startingCDTime
					if hc.OnComplete != nil {
						hc.OnComplete(sim, hc.Target)
					}
				}
			},
		}
		unit.hardcastAction = pa
	} else {
		unit.hardcastAction.cancelled = false
		unit.hardcastAction.NextActionAt = unit.Hardcast.Expires

		if sim.Log != nil {
			unit.Log(sim, "Unit Hardcast extended to %s", unit.Hardcast.Expires)
		}
	}

	sim.AddPendingAction(unit.hardcastAction)
}

func (unit *Unit) NextGCDAt() time.Duration {
	return unit.gcdAction.NextActionAt
}

func (unit *Unit) SetGCDTimer(sim *Simulation, gcdReadyAt time.Duration) {
	if unit.gcdAction == nil {
		return
	}

	unit.GCD.Set(gcdReadyAt)

	if unit.gcdAction.consumed {
		unit.gcdAction.cancelled = false
		unit.gcdAction.NextActionAt = gcdReadyAt
	} else {
		unit.gcdAction.Cancel(sim)
		oldAction := unit.gcdAction.OnAction
		unit.gcdAction = &PendingAction{
			NextActionAt: gcdReadyAt,
			Priority:     ActionPriorityGCD,
			OnAction:     oldAction,
		}
	}
	sim.AddPendingAction(unit.gcdAction)
}

// Call this to stop the GCD loop for a unit.
// This is mostly used for pets that get summoned / expire.
func (unit *Unit) CancelGCDTimer(sim *Simulation) {
	unit.gcdAction.Cancel(sim)
}

func (unit *Unit) WaitUntil(sim *Simulation, readyTime time.Duration) {
	if readyTime < sim.CurrentTime {
		panic(unit.Label + ": cannot wait negative time")
	}

	unit.SetGCDTimer(sim, readyTime)
	if sim.Log != nil && readyTime > sim.CurrentTime {
		unit.Log(sim, "Pausing GCD for %s due to rotation / CDs.", readyTime-sim.CurrentTime)
	}
}

// interruptHardcast stops the cast in progress without finishing it: it costs nothing (a
// cast pays when it completes), does nothing, and gives back the cooldown it started (the
// spell was ready when the cast began). The unit can act again once the GCD the cast started
// has run out. A target dying mid-cast does this (target_death.go), as the game does when
// the target of a cast dies.
func (unit *Unit) interruptHardcast(sim *Simulation) {
	castStart := unit.Hardcast.Start
	if sim.Log != nil {
		unit.Log(sim, "Cast of %s interrupted", unit.Hardcast.ActionID)
	}
	if spell := unit.GetSpell(unit.Hardcast.ActionID); spell != nil {
		if spell.CD.Timer != nil {
			spell.CD.Reset()
		}
		if spell.SharedCD.Timer != nil {
			spell.SharedCD.Reset()
		}
	}
	unit.Hardcast = Hardcast{Expires: startingCDTime}
	unit.ChanneledDot = nil
	if unit.hardcastAction != nil && !unit.hardcastAction.consumed {
		unit.hardcastAction.Cancel(sim)
		unit.hardcastAction = nil
	}
	if unit.gcdAction != nil {
		unit.SetGCDTimer(sim, max(sim.CurrentTime, castStart+GCDDefault))
	}
}
