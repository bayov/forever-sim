package core

import (
	"fmt"
	"time"
)

// A cast corresponds to any action which causes the in-game castbar to be
// shown, and activates the GCD. Note that a cast can also be instant, i.e.
// the effects are applied immediately even though the GCD is still activated.

// Callback for when a cast is finished, i.e. when the in-game castbar reaches full.
type OnCastComplete func(aura *Aura, sim *Simulation, spell *Spell)

type Hardcast struct {
	Expires    time.Duration
	ActionID   ActionID
	OnComplete func(*Simulation, *Unit)
	Target     *Unit
	Pushback   float64
	// The cast's full time. Pushback can't take the cast bar back past its start.
	CastTime time.Duration
	// When the cast started.
	Start time.Duration
	// Under Forever the melee swing waits out the cast, then starts a full swing when it
	// completes.
	ResetsSwing bool
}

// Input for constructing the CastSpell function for a spell.
type CastConfig struct {
	// Default cast values with all static effects applied.
	DefaultCast Cast

	// Dynamic modifications for each cast.
	ModifyCast func(*Simulation, *Spell, *Cast)

	// Ignores haste when calculating the GCD and cast time for this cast.
	// Automatically set if GCD and cast times are all 0, e.g. for empty casts.
	IgnoreHaste bool

	CD       Cooldown
	SharedCD Cooldown

	CastTime func(spell *Spell) time.Duration
}

type Cast struct {
	// Amount of resource that will be consumed by this cast.
	Cost float64

	// The length of time the GCD will be on CD as a result of this cast.
	GCD time.Duration

	// The amount of time between the call to spell.Cast() and when the spell
	// effects are invoked.
	CastTime time.Duration
}

func (cast *Cast) EffectiveTime() time.Duration {
	gcd := cast.GCD
	if cast.GCD != 0 {
		// TODO: isn't this wrong for spells like shadowfury, that have a reduced GCD?
		gcd = max(GCDMin, gcd)
	}
	return max(gcd, cast.CastTime)
}

type CastFunc func(*Simulation, *Unit)
type CastSuccessFunc func(*Simulation, *Unit) bool

func (spell *Spell) castFailureHelper(sim *Simulation, message string, vals ...any) bool {
	if sim.CurrentTime < 0 && spell.Unit.Rotation != nil {
		spell.Unit.Rotation.ValidationWarning(fmt.Sprintf(spell.ActionID.String()+" failed to cast: "+message, vals...))
	} else {
		if sim.Log != nil && !spell.Flags.Matches(SpellFlagNoLogs) {
			spell.Unit.Log(sim, fmt.Sprintf(spell.ActionID.String()+" failed to cast: "+message, vals...))
		}
	}
	return false
}

// Under Forever every cast with a cast time holds the melee swing and starts a full swing when
// it completes, at the attack speed it completes with (wowsims/forever 4fb2849945). A cast that
// fails or is interrupted lets the held swings land at once (wowsims/forever 7146db7c30). Spells
// the client marks Do Not Reset Combat Timers (Attributes[2] 0x20000) carry
// SpellFlagKeepsSwingTimer: Aimed Shot, Sniper Shot, Slam with Improved Slam and Flamestrike.
// Channels and Auto Shot are left out, and instants only reset it by SpellFlagResetAttackSwing.
func (spell *Spell) resetsSwingOnCast(sim *Simulation) bool {
	return sim.IsForever() &&
		!spell.Flags.Matches(SpellFlagChanneled|SpellFlagKeepsSwingTimer) &&
		!spell.ProcMask.Matches(ProcMaskRangedAuto)
}

// applySpellPushback: a hit that lands and does damage pushes back the cast or channel we're
// in, as on the cmangos Classic server (Unit::InterruptOrDelaySpell).
//
// Only direct hits do it: melee swings, ranged attacks and spell hits. Ticks of a DoT don't,
// and neither does a hit that does no damage (a miss, a dodge or a full absorb).
//
// The first hit takes 1 sec, then 0.8, 0.6, 0.4 and 0.2 sec for every hit after that
// (Spell::GetNextDelayAtDamageMsTime). Each cast or channel starts over at 1 sec. A cast's bar
// goes back, so it lands later (Spell::Delayed). A channel ends sooner, and its ticks past the
// new end are lost (Spell::DelayedChannel). Talents like Eye of the Storm, Healing Focus and Fel
// Concentration give the spell we're casting a chance to ignore each hit.
func (unit *Unit) applySpellPushback() {
	// The channel we've pushed back, by its dot and start, how much its next pushback takes,
	// and the action that ends it early.
	var channel *Dot
	var channelStart time.Duration
	var channelPushback float64
	var channelEnd *PendingAction

	unit.RegisterAura(Aura{
		Label:    "Spell Pushback",
		Duration: NeverExpires,
		OnReset: func(aura *Aura, sim *Simulation) {
			aura.Activate(sim)
			channel = nil
			channelEnd = nil
		},
		OnSpellHitTaken: func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			//No pushback for bosses/NPCs
			if unit.Type == EnemyUnit {
				return
			}

			if !result.Landed() {
				return
			}

			if result.Damage <= 0 {
				return
			}

			if !spell.ProcMask.Matches(ProcMaskDirect) {
				return
			}

			if aura.Unit.IsCasting(sim) {
				// We take the chance to ignore pushback from the spell we're casting.
				//
				// Talents like Eye of the Storm let our Lightning Bolt shrug off a hit. The spell
				// that hits us has nothing to do with it.
				hc := aura.Unit.Hardcast
				hcSpell := aura.Unit.GetSpell(hc.ActionID)
				if sim.Roll(0, 1.0) <= hcSpell.PushbackReduction {
					return
				}
				pushback := DurationFromSeconds(max(0.2, hc.Pushback))
				aura.Unit.Hardcast.Pushback -= 0.2

				// The cast bar goes back by the pushback, but not past its start, as on the
				// cmangos Classic server (Spell::Delayed). A hit half a second into a 3 sec
				// cast takes it back half a second, not the full 1 sec.
				if hc.CastTime > 0 {
					pushback = min(pushback, max(0, sim.CurrentTime+hc.CastTime-hc.Expires))
				}
				if sim.Log != nil {
					aura.Unit.Log(sim, "Unit Hardcast extended by %s due to spell hit taken, will now occur at %s", pushback, hc.Expires+pushback)
				}

				aura.Unit.Hardcast.Expires += pushback
				if metricsTarget := aura.Unit.CurrentTarget; metricsTarget != nil {
					hcSpell.SpellMetrics[metricsTarget.UnitIndex].TotalCastTime += pushback
				}

				// Update GCDTimer
				aura.Unit.SetGCDTimer(sim, aura.Unit.Hardcast.Expires)

				// Update Swing timer. Under Forever a cast that resets the swing holds it to its
				// new end (WeaponAttack.swing), and one that keeps it leaves it alone.
				if !sim.IsForever() {
					aura.Unit.AutoAttacks.StopMeleeUntil(sim, aura.Unit.Hardcast.Expires, false)
				}
				return
			}

			// Our own damage doesn't shorten our channel (Hellfire burns us while we channel it),
			// and wand Shoot has no channel time to lose.
			dot := aura.Unit.ChanneledDot
			if dot == nil || !dot.IsActive() || spell.Unit == aura.Unit || (aura.Unit.wand != nil && dot.Spell == aura.Unit.wand.Shoot) {
				return
			}
			if sim.Roll(0, 1.0) <= dot.Spell.PushbackReduction {
				return
			}
			if dot != channel || dot.StartedAt() != channelStart {
				channel, channelStart, channelPushback = dot, dot.StartedAt(), 1.0
			}
			pushback := DurationFromSeconds(max(0.2, channelPushback))
			channelPushback -= 0.2

			newExpires := max(sim.CurrentTime, dot.ExpiresAt()-pushback)
			if sim.Log != nil {
				aura.Unit.Log(sim, "Unit channel shortened by %s due to spell hit taken, will now end at %s", dot.ExpiresAt()-newExpires, newExpires)
				aura.Unit.Log(sim, "Channel %s pushed back to end at %0.3f", dot.Spell.ActionID, newExpires.Seconds())
			}
			dot.UpdateExpires(sim, newExpires)

			// The channel ends at its new end, with the tick due then if one is, and the
			// rotation goes on from there, as it does after a channel's last tick (Dot.TickOnce).
			if channelEnd != nil {
				channelEnd.Cancel(sim)
			}
			channelEnd = &PendingAction{
				NextActionAt: newExpires,
				OnAction: func(sim *Simulation) {
					channelEnd = nil
					// Nothing to do when the rotation stopped the channel before this, or
					// started it again.
					if dot.StartedAt() != channelStart {
						return
					}
					if dot.IsActive() {
						dot.Cancel(sim)
					} else if dot.fadeTime != sim.CurrentTime {
						return
					}
					if !aura.Unit.IsCasting(sim) && aura.Unit.GCD.IsReady(sim) {
						aura.Unit.WaitUntil(sim, sim.CurrentTime+aura.Unit.ChannelClipDelay)
					}
				},
			}
			sim.AddPendingAction(channelEnd)
		},
	})
}

func (spell *Spell) makeCastFunc(config CastConfig) CastSuccessFunc {
	return func(sim *Simulation, target *Unit) bool {
		spell.Unit.completeDueHardcast(sim)

		// A Heroic Strike that a swing through a cast or channel lands (WeaponAttack.swing).
		throughCast := spell.Unit.swingSpellThroughCast == spell
		spell.Unit.swingSpellThroughCast = nil

		spell.CurCast = spell.DefaultCast

		if config.ModifyCast != nil {
			config.ModifyCast(sim, spell, &spell.CurCast)
			if spell.CurCast.Cost != spell.DefaultCast.Cost {
				// Costs need to be modified using the unit and spell multipliers, so that
				// their affects are also visible in the spell.CanCast() function, which
				// does not invoke ModifyCast.
				panic("May not modify cost in ModifyCast!")
			}
		}

		if spell.ExtraCastCondition != nil {
			if !spell.ExtraCastCondition(sim, target) {
				return spell.castFailureHelper(sim, "extra spell condition")
			}
		}

		if spell.Cost != nil {
			if !spell.Cost.MeetsRequirement(sim, spell) {
				return spell.castFailureHelper(sim, spell.Cost.CostFailureReason(sim, spell))
			}
		}

		if !config.IgnoreHaste {
			// Vanilla has no natural GCD reduction besides abilities with 1s GCDs
			// spell.CurCast.GCD = spell.Unit.ApplyFlatCastSpeed(spell.CurCast.GCD)
			spell.CurCast.CastTime = config.CastTime(spell)
		}

		if config.CD.Timer != nil {
			// By panicking if spell is on CD, we force each sim to properly check for their own CDs.
			if !spell.CD.IsReady(sim) {
				return spell.castFailureHelper(sim, "still on cooldown for %s, curTime = %s", spell.CD.TimeToReady(sim), sim.CurrentTime)
			}

			spell.CD.Set(sim.CurrentTime + spell.CurCast.CastTime + spell.CD.Duration)
			spell.logCooldown(sim, spell.CD.Timer)
		}

		if config.SharedCD.Timer != nil {
			// By panicking if spell is on CD, we force each sim to properly check for their own CDs.
			if !spell.SharedCD.IsReady(sim) {
				return spell.castFailureHelper(sim, "still on shared cooldown for %s, curTime = %s", spell.SharedCD.TimeToReady(sim), sim.CurrentTime)
			}

			spell.SharedCD.Set(sim.CurrentTime + spell.CurCast.CastTime + spell.SharedCD.Duration)
			spell.logCooldown(sim, spell.SharedCD.Timer)
		}

		// By panicking if spell is on CD, we force each sim to properly check for their own CDs.
		if spell.CurCast.GCD != 0 && !spell.Unit.GCD.IsReady(sim) {
			return spell.castFailureHelper(sim, "GCD on cooldown for %s, curTime = %s", spell.Unit.GCD.TimeToReady(sim), sim.CurrentTime)
		}

		if hc := spell.Unit.Hardcast; spell.Unit.IsCasting(sim) && !throughCast {
			return spell.castFailureHelper(sim, "casting/channeling %v for %s, curTime = %s", hc.ActionID, hc.Expires-sim.CurrentTime, sim.CurrentTime)
		}

		if dot := spell.Unit.ChanneledDot; spell.Unit.IsChanneling(sim) && !throughCast {
			return spell.castFailureHelper(sim, "channeling %v for %s, curTime = %s", dot.ActionID, dot.expires-sim.CurrentTime, sim.CurrentTime)
		}

		if effectiveTime := spell.CurCast.EffectiveTime(); effectiveTime != 0 {
			if spell.Flags.Matches(SpellFlagCastTimeNoGCD) {
				effectiveTime = max(effectiveTime, spell.Unit.GCD.TimeToReady(sim))
			}
			// do not add channeled time here as they have variable cast length
			// cast time for channels is handled in dot.OnExpire
			if !spell.Flags.Matches(SpellFlagChanneled) {
				spell.SpellMetrics[target.UnitIndex].TotalCastTime += effectiveTime
			}
			spell.Unit.SetGCDTimer(sim, sim.CurrentTime+effectiveTime)
		}

		if (spell.CurCast.CastTime > 0) && spell.Unit.IsMoving() {
			return spell.castFailureHelper(sim, "casting/channeling while moving not allowed!")
		}

		// Non melee casts. Under Forever a cast with a cast time resets the swing when it
		// completes instead (Hardcast.ResetsSwing); instants and channels keep this.
		if spell.Flags.Matches(SpellFlagResetAttackSwing) && spell.Unit.AutoAttacks.enabled && !(sim.IsForever() && spell.CurCast.CastTime > 0) {
			restartMeleeAt := sim.CurrentTime + spell.CurCast.CastTime
			spell.Unit.AutoAttacks.StopMeleeUntil(sim, restartMeleeAt, false)
		}

		// Hardcasts
		if spell.CurCast.CastTime > 0 {
			resetsSwing := spell.resetsSwingOnCast(sim)

			if sim.Log != nil && !spell.Flags.Matches(SpellFlagNoLogs) {
				spell.Unit.Log(sim, "Casting %s (Cost = %0.03f, Cast Time = %s, Effective Time = %s)",
					spell.ActionID, max(0, spell.CurCast.Cost), spell.CurCast.CastTime, spell.CurCast.EffectiveTime())
			}

			spell.Unit.Hardcast = Hardcast{
				Expires:  sim.CurrentTime + spell.CurCast.CastTime,
				ActionID: spell.ActionID,
				Pushback: 1.0,
				CastTime: spell.CurCast.CastTime,
				Start:    sim.CurrentTime,
				OnComplete: func(sim *Simulation, target *Unit) {
					spell.LastCastAt = sim.CurrentTime

					if sim.Log != nil && !spell.Flags.Matches(SpellFlagNoLogs) {
						spell.Unit.Log(sim, "Completed cast %s", spell.ActionID)
					}

					if spell.Cost != nil {
						if !spell.Cost.MeetsRequirement(sim, spell) {
							spell.castFailureHelper(sim, spell.Cost.CostFailureReason(sim, spell))
							spell.Unit.AutoAttacks.releaseMeleeHeldByCast(sim)
							return
						}
					}

					if resetsSwing {
						spell.Unit.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime, false)
					}

					if spell.Cost != nil {
						spell.Cost.SpendCost(sim, spell)
					}

					spell.applyEffects(sim, target)

					if !spell.Flags.Matches(SpellFlagNoOnCastComplete) {
						spell.Unit.OnCastComplete(sim, spell)
					}

					if !sim.Options.Interactive {
						spell.Unit.Rotation.DoNextAction(sim)
					}
				},
				Target:      target,
				ResetsSwing: resetsSwing,
			}

			if spell.Unit.Hardcast.Expires != spell.Unit.NextGCDAt() {
				spell.Unit.newHardcastAction(sim)
			}

			return true
		}

		spell.LastCastAt = sim.CurrentTime

		if sim.Log != nil && !spell.Flags.Matches(SpellFlagNoLogs) {
			spell.Unit.Log(sim, "Casting %s (Cost = %0.03f, Cast Time = %s, Effective Time = %s)",
				spell.ActionID, max(0, spell.CurCast.Cost), spell.CurCast.CastTime, spell.CurCast.EffectiveTime())
			spell.Unit.Log(sim, "Completed cast %s", spell.ActionID)
		}

		if spell.Cost != nil {
			spell.Cost.SpendCost(sim, spell)
		}

		spell.applyEffects(sim, target)

		if !spell.Flags.Matches(SpellFlagNoOnCastComplete) {
			spell.Unit.OnCastComplete(sim, spell)
		}

		return true
	}
}

func (spell *Spell) makeCastFuncSimple() CastSuccessFunc {
	return func(sim *Simulation, target *Unit) bool {
		if spell.ExtraCastCondition != nil {
			if !spell.ExtraCastCondition(sim, target) {
				return spell.castFailureHelper(sim, "extra spell condition")
			}
		}

		if spell.CD.Timer != nil {
			// By panicking if spell is on CD, we force each sim to properly check for their own CDs.
			if !spell.CD.IsReady(sim) {
				return spell.castFailureHelper(sim, "still on cooldown for %s, curTime = %s", spell.CD.TimeToReady(sim), sim.CurrentTime)
			}

			spell.CD.Set(sim.CurrentTime + spell.CD.Duration)
			spell.logCooldown(sim, spell.CD.Timer)
		}

		if spell.SharedCD.Timer != nil {
			// By panicking if spell is on CD, we force each sim to properly check for their own CDs.
			if !spell.SharedCD.IsReady(sim) {
				return spell.castFailureHelper(sim, "still on shared cooldown for %s, curTime = %s", spell.SharedCD.TimeToReady(sim), sim.CurrentTime)
			}

			spell.SharedCD.Set(sim.CurrentTime + spell.SharedCD.Duration)
			spell.logCooldown(sim, spell.SharedCD.Timer)
		}

		if sim.Log != nil && !spell.Flags.Matches(SpellFlagNoLogs) {
			spell.Unit.Log(sim, "Casting %s (Cost = %0.03f, Cast Time = %s, Effective Time = %s)",
				spell.ActionID, 0.0, "0s", "0s")
			spell.Unit.Log(sim, "Completed cast %s", spell.ActionID)
		}

		spell.applyEffects(sim, target)

		if !spell.Flags.Matches(SpellFlagNoOnCastComplete) {
			spell.Unit.OnCastComplete(sim, spell)
		}

		return true
	}
}

func (spell *Spell) makeCastFuncAutosOrProcs() CastSuccessFunc {
	return func(sim *Simulation, target *Unit) bool {
		if sim.Log != nil && !spell.Flags.Matches(SpellFlagNoLogs) {
			spell.Unit.Log(sim, "Casting %s (Cost = %0.03f, Cast Time = %s, Effective Time = %s)",
				spell.ActionID, 0.0, "0s", "0s")
			spell.Unit.Log(sim, "Completed cast %s", spell.ActionID)
		}

		spell.applyEffects(sim, target)

		if !spell.Flags.Matches(SpellFlagNoOnCastComplete) {
			spell.Unit.OnCastComplete(sim, spell)
		}

		return true
	}
}

// logCooldown writes when the spell is off cooldown again, like
// `{SpellID: 17364} on cooldown until 12.500 (timer 4)`, for the Replay view of the Events tab.
//
// The timer number is the timer's place among the unit's timers. Spells that share a timer
// (like Earth Shock and Flame Shock) log the same number, so the replay puts all of them on
// cooldown when we cast one.
func (spell *Spell) logCooldown(sim *Simulation, timer *Timer) {
	if sim.Log == nil || spell.Flags.Matches(SpellFlagNoLogs) || timer.ReadyAt() <= sim.CurrentTime {
		return
	}
	timerIndex := -1
	for i, unitTimer := range spell.Unit.cdTimers {
		if unitTimer == timer {
			timerIndex = i
			break
		}
	}
	spell.Unit.Log(sim, "%s on cooldown until %0.3f (timer %d)", spell.ActionID, timer.ReadyAt().Seconds(), timerIndex)
}
