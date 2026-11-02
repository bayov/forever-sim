//go:build practice

package practice

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/shaman"
)

// Casting your spells.
//
// Each cast input names a spell id, which we look up in the session's own toolkit table, never
// with Unit.GetSpell: the Flametongue and Frostbrand proc spells share the imbue's spell id. Then
// we run the game's checks in its order and answer with its error text when one fails. Only then
// do we call Spell.Cast, because a failed Cast can use up the cooldown.

func (r *run) pressCast(sim *core.Simulation, in Input) {
	spell := r.toolkit[in.SpellID]
	if spell == nil {
		r.s.answer(in.ID, false, errUnknownSpell)
		return
	}
	r.press(sim, in, spell)
}

func (r *run) pressItem(sim *core.Simulation, in Input) {
	spell := r.items[in.ItemID]
	if spell == nil {
		r.s.answer(in.ID, false, "You don't have that item")
		return
	}
	r.press(sim, in, spell)
}

// press handles a press of a spell or item: it casts now, queues it for the GCD's or the cast's
// end when that's within the spell queue window, or fails.
func (r *run) press(sim *core.Simulation, in Input, spell *core.Spell) {
	readyAt := r.readyAt(sim, spell)
	if readyAt > sim.CurrentTime {
		window := fromMs(r.s.settings.SpellQueueWindowMs)
		if readyAt-sim.CurrentTime > window {
			r.s.answer(in.ID, false, r.notReadyError(spell))
			return
		}
		// The last press in the window wins.
		r.cancelQueued()
		var queued *action
		queued = r.enqueueWorld(readyAt, func(sim *core.Simulation) {
			if r.queued == queued {
				r.queued = nil
			}
			r.tryCast(sim, in, spell)
		})
		r.queued = queued
		r.queuedInput = in.ID
		r.queuedSpell = spell
		return
	}
	r.tryCast(sim, in, spell)
}

// readyAt is when the spell's GCD, cooldown and the cast in progress all allow it.
func (r *run) readyAt(sim *core.Simulation, spell *core.Spell) time.Duration {
	at := sim.CurrentTime
	if spell.DefaultCast.GCD > 0 {
		at = max(at, r.char.GCD.ReadyAt())
	}
	if r.char.IsCasting(sim) && !spell.Flags.Matches(core.SpellFlagCastWhileCasting) {
		at = max(at, r.char.Hardcast.Expires)
	}
	// ReadyAt is 0 for a spell with no cooldown, which before the pull would look busy.
	if !spell.IsReady(sim) {
		at = max(at, spell.ReadyAt())
	}
	return at
}

func (r *run) notReadyError(spell *core.Spell) string {
	if spell.ActionID.ItemID != 0 {
		return errItemNotReady
	}
	if spell.ProcMask.Matches(core.ProcMaskMelee) {
		return errAbilityReady
	}
	return errNotReady
}

// harmful is whether the spell attacks its target, as the core decides it (Spell.attacksTarget).
func harmful(spell *core.Spell) bool {
	return !spell.Flags.Matches(core.SpellFlagHelpful) && spell.ProcMask != core.ProcMaskEmpty &&
		spell.ProcMask != core.ProcMaskUnknown && !spell.Flags.Matches(SpellFlagTotemSpell)
}

// SpellFlagTotemSpell is the shaman totem flag, for spells that hit from where the totem stands.
const SpellFlagTotemSpell = shaman.SpellFlagTotem

// tryCast runs the checks and casts.
func (r *run) tryCast(sim *core.Simulation, in Input, spell *core.Spell) {
	if errText := r.checkCast(sim, spell); errText != "" {
		r.s.answer(in.ID, false, errText)
		return
	}
	target := &r.char.Unit
	if harmful(spell) {
		target = r.dummy
	}
	if !spell.Cast(sim, target) {
		r.s.answer(in.ID, false, errCantDoThat)
		return
	}
	r.afterCast(sim, spell)
	r.s.answer(in.ID, true, "")
}

// checkCast runs the game's checks in its order: moving, range and facing, timers and cost,
// then the core's own CanCast.
func (r *run) checkCast(sim *core.Simulation, spell *core.Spell) string {
	now := sim.CurrentTime
	isHarmful := harmful(spell)
	if spell.DefaultCast.CastTime > 0 && (r.me.moving() || r.me.airborne(now)) {
		return errMoving
	}
	if isHarmful {
		if !r.hasTarget {
			return errNoTarget
		}
		if errText := r.rangeError(sim, spell); errText != "" {
			return errText
		}
	}
	if r.shaman != nil && r.isFireNova(spell) {
		if !r.totemPosSet[shaman.FireTotem] || r.shaman.ActiveTotems[shaman.FireTotem] == nil {
			return errCantDoThat
		}
		if r.totemPos[shaman.FireTotem].dist(r.me.posAt(now)) > 30 {
			return errOutOfRange
		}
	}
	if spell.DefaultCast.GCD > 0 && !r.char.GCD.IsReady(sim) {
		return r.notReadyError(spell)
	}
	if !spell.IsReady(sim) {
		return r.notReadyError(spell)
	}
	if spell.Cost != nil && !spell.Cost.MeetsRequirement(sim, spell) {
		return costError(spell.Cost.CostFailureReason(sim, spell))
	}
	if r.char.IsCasting(sim) && !spell.Flags.Matches(core.SpellFlagCastWhileCasting) {
		return errNotReady
	}
	if !spell.CanCast(sim, &r.char.Unit) {
		return errCantDoThat
	}
	return ""
}

func (r *run) isFireNova(spell *core.Spell) bool {
	for _, nova := range r.shaman.FireNova {
		if nova == spell {
			return true
		}
	}
	return false
}

// rangeError checks range and facing for a harmful spell.
func (r *run) rangeError(sim *core.Simulation, spell *core.Spell) string {
	now := sim.CurrentTime
	me, target := r.me.posAt(now), r.targetPos(sim)
	dist := me.dist(target)
	if spell.ProcMask.Matches(core.ProcMaskMelee) {
		if dist > meleeRange(&r.me, &r.target) {
			return errOutOfRange
		}
	} else if spell.MaxRange > 0 && dist > spellRange(spell.MaxRange, &r.me, &r.target) {
		return errOutOfRange
	}
	if !inArc(me, r.me.facingAt(now), target, spellArc) {
		return errFacing
	}
	return ""
}

// afterCast keeps what the HUD and the summary need, turns auto attack on for a melee ability,
// and checks range and facing at a cast's end.
func (r *run) afterCast(sim *core.Simulation, spell *core.Spell) {
	now := sim.CurrentTime
	if spell.DefaultCast.GCD > 0 {
		r.gcdStart = now
	}
	// Idle GCD time counts from the pull: the time nothing kept the GCD busy.
	if now >= 0 && r.s.pulled {
		if now > r.busyUntil {
			r.idleGCD += now - max(r.busyUntil, 0)
		}
	}
	busy := max(r.char.GCD.ReadyAt(), now)
	if r.char.IsCasting(sim) {
		busy = max(busy, r.char.Hardcast.Expires)
	}
	r.busyUntil = max(r.busyUntil, busy)

	key := hitKey{spell.ActionID.SpellID, spell.ActionID.ItemID, int32(spell.ActionID.OtherID), spell.ActionID.Tag}
	if now >= 0 {
		r.casts[key]++
	}
	if !r.quiet {
		r.s.batch.Casts = append(r.s.batch.Casts, CastEvent{At: toMs(now), SpellID: spell.ActionID.SpellID, ItemID: spell.ActionID.ItemID})
	}

	if spell.Flags.Matches(shaman.SpellFlagTotem) && r.shaman != nil {
		r.placeTotem(sim, spell)
	}

	if spell.ProcMask.Matches(core.ProcMaskMelee) && !r.attackOn {
		r.attackOn = true
		r.updateWorld(sim)
	}

	if r.char.IsCasting(sim) {
		r.scheduleCastEndCheck(sim)
	}
}

// placeTotem remembers where a totem stands: about 2 yards from you, at a fixed angle for each
// element.
func (r *run) placeTotem(sim *core.Simulation, spell *core.Spell) {
	for slot, active := range r.shaman.ActiveTotems {
		if active != spell {
			continue
		}
		angles := [4]float64{shaman.AirTotem: 0.6, shaman.EarthTotem: 1.4, shaman.FireTotem: -0.6, shaman.WaterTotem: -1.4}
		facing := r.me.facingAt(sim.CurrentTime)
		r.totemPos[slot] = r.me.posAt(sim.CurrentTime).add(fromAngle(facing + angles[slot]).scale(2))
		r.totemPosSet[slot] = true
	}
}

// scheduleCastEndCheck checks range and facing again at the cast's end, before the cast lands,
// and interrupts the cast when it fails.
func (r *run) scheduleCastEndCheck(sim *core.Simulation) {
	if r.castCheck != nil {
		r.castCheck.Cancel(sim)
	}
	hc := r.char.Hardcast
	spell := r.char.GetSpell(hc.ActionID)
	if spell == nil || !harmful(spell) {
		return
	}
	start := hc.Start
	var pa *core.PendingAction
	// It runs 1ns before the end, because at the end the core already counts the cast as done
	// and an interrupt does nothing.
	pa = &core.PendingAction{
		NextActionAt: hc.Expires - 1,
		Priority:     castEndCheckPriority,
		OnAction: func(sim *core.Simulation) {
			if r.castCheck == pa {
				r.castCheck = nil
			}
			now := r.char.Hardcast

			if now.Start != start || now.ActionID != hc.ActionID {
				return
			}
			if now.Expires > sim.CurrentTime+1 {
				// Pushed back: check at the new end.
				r.scheduleCastEndCheck(sim)
				return
			}
			r.settleWorld(sim)
			if errText := r.castEndError(sim, spell); errText != "" {
				r.char.InterruptHardcast(sim)
				if !r.quiet {
					r.s.batch.Errors = append(r.s.batch.Errors, errText)
				}
			}
		},
	}
	r.castCheck = pa
	sim.AddPendingAction(pa)
}

// castEndError is the range and facing check at a cast's end, with about 10% of the max range
// (at most 3 yards) of extra room.
func (r *run) castEndError(sim *core.Simulation, spell *core.Spell) string {
	if !r.hasTarget {
		return errNoTarget
	}
	now := sim.CurrentTime
	me, target := r.me.posAt(now), r.targetPos(sim)
	if spell.MaxRange > 0 {
		extra := min(spell.MaxRange*0.1, 3)
		if me.dist(target) > spellRange(spell.MaxRange+extra, &r.me, &r.target) {
			return "Out of range"
		}
	}
	if !inArc(me, r.me.facingAt(now), target, spellArc) {
		return errFacing
	}
	return ""
}
