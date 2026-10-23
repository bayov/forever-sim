package core

import (
	"time"

	"github.com/wowsims/classic/sim/core/proto"
)

// Forever's wand Shoot (5019) for mages, priests and warlocks, as wowsims/forever has it
// (7ddd02e9e2, dcc2aaf371, 4fd24939d8, 924140e5b2). Shoot is an autorepeat channel: it fires
// one shot per wand speed, hasted by cast speed when it starts, until the rotation casts
// something else, which it looks for after each shot and every 50 ms in between. The first
// shot comes 0.5 sec after Shoot starts, and no sooner than one wand speed after the wand's
// timer last restarted, which a completed cast, the end of another channel and every shot
// do. Moving ends it, and it can't start on the move (the row's interrupt flags hold
// movement and casting).
//
// A shot rolls the wand's damage in the wand's school, with no attack power and no spell
// power, travels at speed 20, misses on the ranged weapon table at wand skill (physical hit
// lowers it, spell hit doesn't), crits on spell crit for 1.5 times, and takes no partial
// resist (OutcomeRangedHitAndMagicCrit).
const WandShootSpellID = 5019

const wandWindup = 500 * time.Millisecond

type Wand struct {
	unit   *Unit
	Shoot  *Spell
	weapon Weapon

	target   *Unit
	interval time.Duration // The shot interval, fixed when Shoot starts.
	slot     *PendingAction

	timerSet bool
	timerAt  time.Duration // When the wand's timer last restarted.
}

// RegisterWandShoot registers Shoot for a character holding a wand under Forever, and
// returns it; nil otherwise.
func (character *Character) RegisterWandShoot() *Spell {
	if character.Env == nil || !character.Env.IsForever() {
		return nil
	}
	item := character.GetRangedWeapon()
	if item == nil || item.RangedWeaponType != proto.RangedWeaponType_RangedWeaponTypeWand || item.SwingSpeed <= 0 {
		return nil
	}

	wand := &Wand{unit: &character.Unit, weapon: character.WeaponFromRanged()}
	character.wand = wand

	wand.Shoot = character.RegisterSpell(SpellConfig{
		ActionID:    ActionID{SpellID: WandShootSpellID},
		SpellSchool: wand.weapon.GetSpellSchool(),
		DefenseType: DefenseTypeMagic,
		ProcMask:    ProcMaskRangedAuto,
		CastType:    proto.CastType_CastTypeRanged,
		Flags:       SpellFlagChanneled | SpellFlagAPL | SpellFlagNoPartialResist | SpellFlagNoOnCastComplete,

		MissileSpeed: 20,
		MaxRange:     30,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ExtraCastCondition: func(_ *Simulation, _ *Unit) bool {
			return !character.IsMoving()
		},

		// The channel carries the channeling state; the shots come from the wand's own slot,
		// so the dot never ticks.
		Dot: DotConfig{
			Aura: Aura{
				Label: "Shoot",
				OnExpire: func(_ *Aura, sim *Simulation) {
					if wand.slot != nil {
						wand.slot.Cancel(sim)
						wand.slot = nil
					}
				},
			},
			NumberOfTicks: 1,
			TickLength:    wandWindup,
			OnTick:        func(_ *Simulation, _ *Unit, _ *Dot) {},
		},

		ApplyEffects: func(sim *Simulation, target *Unit, _ *Spell) {
			wand.start(sim, target)
		},
	})

	character.RegisterResetEffect(func(_ *Simulation) {
		wand.timerSet = false
		wand.slot = nil
	})

	// Every completed cast of a spell the rotation or a cooldown uses restarts the timer.
	// Procs and auto attacks leave it alone.
	MakePermanent(character.RegisterAura(Aura{
		Label: "Wand Timer",
		OnCastComplete: func(_ *Aura, sim *Simulation, spell *Spell) {
			if spell != wand.Shoot && spell.Flags.Matches(SpellFlagAPL|SpellFlagMCD) {
				wand.restartTimer(sim)
			}
		},
	}))

	return wand.Shoot
}

func (wand *Wand) restartTimer(sim *Simulation) {
	wand.timerSet = true
	wand.timerAt = sim.CurrentTime
}

// Another channel's end restarts the timer (dot.go).
func (wand *Wand) channelEnded(sim *Simulation, dot *Dot) {
	if wand != nil && dot.Spell != wand.Shoot {
		wand.restartTimer(sim)
	}
}

func (wand *Wand) start(sim *Simulation, target *Unit) {
	// A channel action that started Shoot sets its interruptIf once Shoot has started. Any
	// other start drops what an earlier channel action left, like a Mind Flay that missed and
	// so never channeled (wowsims/forever 4fd24939d8).
	if rot := wand.unit.Rotation; rot != nil {
		rot.interruptChannelIf = nil
		rot.allowChannelRecastOnInterrupt = false
	}
	wand.target = target
	wand.interval = wand.unit.ApplyCastSpeedForSpell(DurationFromSeconds(wand.weapon.SwingSpeed), wand.Shoot).Round(time.Millisecond)

	firstShotAt := sim.CurrentTime + wandWindup
	if wand.timerSet {
		firstShotAt = max(firstShotAt, wand.timerAt+wand.interval)
	}

	dot := wand.Shoot.Dot(target)
	dot.TickLength = max(sim.GetRemainingDuration(), firstShotAt-sim.CurrentTime) + wand.interval
	dot.Apply(sim)
	if tickAction := dot.tickAction; tickAction != nil {
		dot.tickAction = nil
		tickAction.Cancel(sim)
	}

	wand.schedule(sim, firstShotAt)
}

func (wand *Wand) schedule(sim *Simulation, at time.Duration) {
	wand.slot = &PendingAction{NextActionAt: at, OnAction: wand.fire}
	sim.AddPendingAction(wand.slot)
}

func (wand *Wand) wanding() bool {
	return wand.unit.ChanneledDot != nil && wand.unit.ChanneledDot.Spell == wand.Shoot
}

// Fires the shot due now. Then Shoot gives way to the rotation if it would do something
// else, and otherwise shoots again one interval later.
func (wand *Wand) fire(sim *Simulation) {
	wand.slot = nil
	if !wand.wanding() {
		return
	}
	dot := wand.unit.ChanneledDot

	spell := wand.Shoot
	result := spell.CalcDamage(sim, wand.target, wand.weapon.BaseDamage(sim), spell.OutcomeRangedHitAndMagicCrit)
	spell.WaitTravelTime(sim, func(sim *Simulation) {
		spell.DealDamage(sim, result)
	})
	wand.restartTimer(sim)

	if !wand.wanding() {
		return
	}
	if wand.givesWay(sim) {
		dot.Cancel(sim)
		// The rotation acts at once, with no channel clip delay, or when the GCD ends.
		wand.unit.WaitUntil(sim, max(sim.CurrentTime, wand.unit.GCD.ReadyAt()))
		return
	}

	next := sim.CurrentTime + wand.interval
	// A fight that runs past its expected length keeps the channel going.
	if dot.ExpiresAt() <= next {
		dot.UpdateExpires(sim, next+wand.interval)
	}
	wand.schedule(sim, next)
}

// Whether the rotation, with Shoot set aside, would now do something other than Shoot. With
// nothing else ready Shoot keeps shooting. A channel action's interruptIf has to hold first,
// so Shoot started that way runs until it does. The rotation asks this whenever it looks
// (APLRotation.shouldInterruptChannel), and the wand after each shot.
func (wand *Wand) givesWay(sim *Simulation) bool {
	rot := wand.unit.Rotation
	if rot == nil {
		return false
	}
	if interruptIf := rot.interruptChannelIf; interruptIf != nil && !interruptIf.GetBool(sim) {
		return false
	}
	channeledDot := wand.unit.ChanneledDot
	wand.unit.ChanneledDot = nil
	next := rot.getNextAction(sim)
	wand.unit.ChanneledDot = channeledDot
	return next != nil && !wand.isShootAction(next)
}

func (wand *Wand) isShootAction(action *APLAction) bool {
	if wand == nil {
		return false
	}
	switch impl := action.impl.(type) {
	case *APLActionCastSpell:
		return impl.spell == wand.Shoot
	case *APLActionChannelSpell:
		return impl.spell == wand.Shoot
	}
	return false
}

// Wand Specialization (mage 6057, priest 14524 under Forever): +bonus damage with wands.
// The rows' school mask (126) is every school but physical, so a physical wand gets nothing
// (wowsims/forever 96c0c178ed).
func (character *Character) ApplyWandSpecialization(bonus float64) {
	if bonus == 0 {
		return
	}
	character.OnSpellRegistered(func(spell *Spell) {
		if spell.SpellID == WandShootSpellID && !spell.SpellSchool.Matches(SpellSchoolPhysical) {
			spell.DamageMultiplier *= 1 + bonus
		}
	})
}
