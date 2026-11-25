package core

import (
	"time"

	"github.com/wowsims/classic/sim/core/proto"
)

// What a melee class needs of a leveling session (leveling.go): charging at a mob, stunning one,
// the area abilities that only reach the mobs near us, what a kill gives, and Cannibalize.

// Charge and Intercept run us to the target at this speed, 25 yards in 0.6 sec (cmangos
// SPEED_CHARGE). The client doesn't hold it.
const levelingChargeSpeed = 42

// LevelingInCombat is whether we're in combat. Outside a session we always are once the fight has
// started.
func (encounter *Encounter) LevelingInCombat(sim *Simulation) bool {
	if encounter.leveling == nil {
		return sim.CurrentTime > 0
	}
	return encounter.leveling.state == levelingCombat
}

// LevelingCanCharge is whether the target stands between minRange and maxRange yards from us: the
// mob we found, or one on its way to us. A Charge also asks for us to be out of combat, which it
// checks itself.
func (encounter *Encounter) LevelingCanCharge(sim *Simulation, target *Unit, minRange, maxRange float64) bool {
	s := encounter.leveling
	if s == nil || target == nil || target.Type != EnemyUnit || s.sitting() {
		return false
	}
	if state := s.mobs[target.Index].state; state != mobFound && state != mobRunning {
		return false
	}
	if s.mobs[target.Index].incoming {
		return false
	}
	distance := s.distance(sim, target)
	return distance >= minRange-levelingYardTolerance && distance <= maxRange+levelingYardTolerance
}

// LevelingCharge runs us up to the target in well under a second. The mob we found is pulled by
// it. stunned is how long the charge's stun holds the mob, from now.
func (encounter *Encounter) LevelingCharge(sim *Simulation, target *Unit, stunned time.Duration) {
	s := encounter.leveling
	if sim.Log != nil {
		s.player.Log(sim, "Leveling: charging %s", target.Label)
	}
	if s.mobs[target.Index].state == mobFound {
		s.aggro(sim, target)
	}
	if target != s.player.CurrentTarget {
		s.retarget(sim, target)
	}
	s.move(sim, goalNone, Vec2{}, 0, nil)

	// The mob runs at us all the while. Its stun keeps it from that in the game, which makes a
	// tenth of a second of difference.
	here, there := s.playerPosition(sim), s.mobPosition(sim, target)
	gap := max(there.Dist(here)-levelingMeleeRange, 0)
	stunEnd := sim.CurrentTime + stunned
	s.goal, s.goalPoint, s.goalWithin = goalPoint, there, levelingMeleeRange
	if gap > 0 {
		s.setVelocity(sim, there.Sub(here).unit().Scale(levelingChargeSpeed))
	}
	took := DurationFromSeconds(gap / (levelingChargeSpeed + mobRunSpeed(target)))
	s.after(sim, &s.goalReached, took+time.Millisecond, func(sim *Simulation) {
		s.move(sim, goalNone, Vec2{}, 0, nil)
		target.holdMeleeUntil(sim, stunEnd)
	})
	target.Stun(sim, stunned)
}

// Stun stuns a mob: it doesn't swing until the stun is over, and can't dodge, parry or block
// meanwhile. A boss is immune, so only a leveling session's mobs are stunned.
func (unit *Unit) Stun(sim *Simulation, duration time.Duration) {
	if unit.Type != EnemyUnit || unit.Env.Encounter.leveling == nil || unit.dead || unit.stunAura == nil {
		return
	}
	if unit.stunAura.IsActive() && unit.stunAura.RemainingDuration(sim) >= duration {
		return
	}
	unit.stunAura.Duration = duration
	unit.stunAura.Activate(sim)
	unit.holdMeleeUntil(sim, sim.CurrentTime+duration)
}

// RegisterStuns lets Stun stun the session's mobs. A class that stuns calls it as it is built.
func (env *Environment) RegisterStuns() {
	if env.Encounter.leveling == nil {
		return
	}
	for _, target := range env.Encounter.TargetUnits {
		if target.stunAura != nil {
			continue
		}
		target.stunAura = target.RegisterAura(Aura{
			Label:    "Stunned",
			Duration: time.Second,
			OnGain: func(aura *Aura, _ *Simulation) {
				aura.Unit.PseudoStats.Stunned = true
			},
			OnExpire: func(aura *Aura, _ *Simulation) {
				aura.Unit.PseudoStats.Stunned = false
			},
		})
	}
}

// holdMeleeUntil keeps the unit's swings from landing before t. The swing timer goes on while it's
// held, so a swing that came due lands right at t.
func (unit *Unit) holdMeleeUntil(sim *Simulation, t time.Duration) {
	aa := &unit.AutoAttacks
	if !aa.AutoSwingMelee {
		return
	}
	aa.mh.swingAt = max(aa.mh.swingAt, t)
	if aa.IsDualWielding {
		aa.oh.swingAt = max(aa.oh.swingAt, t)
	}
	if aa.enabled {
		sim.rescheduleWeaponAttack(aa.mh.swingAt)
	}
}

// AreaTargets are the enemies an area ability with this radius hits, at most limit of them, the
// target first. In a leveling session those are the mobs we pulled that are near enough: a mob
// still running at us, or one we haven't pulled, isn't hit. Anywhere else every living target is
// in reach.
func (encounter *Encounter) AreaTargets(sim *Simulation, target *Unit, radius float64, limit int, buf []*Unit) []*Unit {
	buf = buf[:0]
	s := encounter.leveling
	if s == nil {
		for range encounter.AoETargetCount(int32(limit)) {
			buf = append(buf, target)
			target = sim.Environment.NextTargetUnit(target)
		}
		return buf
	}
	near := func(mob *Unit) bool {
		return s.pulled(mob) && s.distance(sim, mob) <= max(radius, levelingMeleeRange)+levelingYardTolerance
	}
	if target != nil && !target.dead && near(target) {
		buf = append(buf, target)
	}
	for _, mob := range encounter.ActiveTargetUnits {
		if len(buf) >= limit {
			break
		}
		if mob != target && near(mob) {
			buf = append(buf, mob)
		}
	}
	return buf
}

// InExecuteRange is whether an ability that asks for a target at or below this share of its health
// can go at it. A leveling session's mobs each have their own health, and elsewhere the fight's
// execute phase says it.
func (sim *Simulation) InExecuteRange(target *Unit, percent int32) bool {
	if sim.Encounter.leveling != nil {
		return target != nil && target.Type == EnemyUnit && target.HasHealthBar() &&
			target.CurrentHealthPercent() <= float64(percent)/100
	}
	return sim.executePhase <= percent
}

// RegisterKillEffect runs effect when a mob we pulled dies in a leveling session, giving whether
// it was one that gives experience (a non-trivial enemy).
func (character *Character) RegisterKillEffect(effect func(sim *Simulation, target *Unit, gaveXP bool)) {
	character.levelingKillEffects = append(character.levelingKillEffects, effect)
}

// onKill: a mob we pulled died. Runs before it goes back to the pool, so it still has its place.
func (s *levelingSession) onKill(sim *Simulation, target *Unit) {
	gaveXP := mobKillXP(s.level, target.Level) > 0
	for _, effect := range s.character.levelingKillEffects {
		effect(sim, target, gaveXP)
	}
}

// Cannibalize: an Undead eats a Humanoid or Undead corpse within 5 yards for 7% of max health (and
// mana) every 2 sec for 10 sec. Moving, acting or taking damage ends it (client 71091, spells
// 20577 and 20578). We eat out of combat only, standing over what we killed last: in combat a
// hit would end it.
const (
	cannibalizeRange    = 5
	cannibalizeShare    = 0.07
	cannibalizeTicks    = 5
	cannibalizeTick     = 2 * time.Second
	cannibalizeCooldown = 2 * time.Minute
)

// registerCannibalize gives an Undead the racial in a leveling session against Humanoids or
// Undead. The rotation casts it, and waits while we eat.
func (character *Character) registerCannibalize() {
	s := character.Env.Encounter.leveling
	if s == nil || s.character != character {
		return
	}
	if mobType := s.opts.mobType; mobType != proto.MobType_MobTypeHumanoid && mobType != proto.MobType_MobTypeUndead {
		return
	}
	actionID := ActionID{SpellID: 20577}
	metrics := character.NewHealthMetrics(actionID)
	var manaMetrics *ResourceMetrics
	if character.HasManaBar() {
		manaMetrics = character.NewManaMetrics(actionID)
	}

	// Where the last mob died.
	var corpse Vec2
	hasCorpse := false
	character.RegisterResetEffect(func(*Simulation) {
		hasCorpse = false
	})
	character.RegisterKillEffect(func(sim *Simulation, target *Unit, _ bool) {
		corpse, hasCorpse = s.mobPosition(sim, target), true
	})

	character.RegisterSpell(SpellConfig{
		ActionID: actionID,
		ProcMask: ProcMaskEmpty,
		Flags:    SpellFlagAPL | SpellFlagHelpful | SpellFlagNoOnCastComplete,
		Cast: CastConfig{
			CD: Cooldown{Timer: character.NewTimer(), Duration: cannibalizeCooldown},
		},
		ExtraCastCondition: func(sim *Simulation, _ *Unit) bool {
			return s.outOfCombatState() && hasCorpse && !character.IsCasting(sim) && character.GCD.IsReady(sim) &&
				corpse.Dist(s.playerPosition(sim)) <= cannibalizeRange+levelingYardTolerance &&
				character.CurrentHealth() < character.MaxHealth()
		},
		ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
			hasCorpse = false
			s.startCannibalize(sim, metrics, manaMetrics)
		},
	})
}

// startCannibalize keeps us at the corpse like a bandage does, with the rotation waiting.
func (s *levelingSession) startCannibalize(sim *Simulation, metrics, manaMetrics *ResourceMetrics) {
	s.setState(sim, levelingBandaging)
	s.move(sim, goalNone, Vec2{}, 0, nil)
	s.bandageLeft = 0
	if sim.Log != nil {
		s.player.Log(sim, "Leveling: cannibalizing")
	}
	total := cannibalizeTicks * cannibalizeTick
	s.busy(sim, total)
	started := sim.CurrentTime
	StartPeriodicAction(sim, PeriodicActionOptions{
		Period:   cannibalizeTick,
		NumTicks: cannibalizeTicks,
		// Before the session moves on at the last tick's moment.
		Priority: ActionPriorityDOT,
		OnAction: func(sim *Simulation) {
			if s.state != levelingBandaging || s.stateSince != started {
				return
			}
			p := s.player
			s.heal(sim, cannibalizeShare*p.MaxHealth(), metrics)
			if p.HasManaBar() {
				p.AddMana(sim, min(cannibalizeShare*p.MaxMana(), p.MaxMana()-p.CurrentMana()), manaMetrics)
			}
		},
	})
	s.after(sim, &s.next, total, s.outOfCombatDone)
}

// Out of combat a rage bar empties by 1.25 rage a second, from the moment we leave combat (client
// 71091, PowerType row RAGE: RegenPeace -12.5 tenths, RegenInterruptTimeMS 0).
const rageDecayPerSecond = 1.25

// RageDecayOptions are what talents change about it.
type RageDecayOptions struct {
	Multiplier float64       // Anger Management: 0.7
	Delay      time.Duration // Lingering Rage: how long after leaving combat it starts
	Metrics    *ResourceMetrics
}

// EnableLevelingRageDecay has the rage bar empty out of combat in a leveling session, and at once
// when we die. Elsewhere there is no out of combat.
func (character *Character) EnableLevelingRageDecay(opts RageDecayOptions) {
	const period = time.Second
	character.RegisterResetEffect(func(sim *Simulation) {
		s := sim.Encounter.leveling
		if s == nil || s.character != character {
			return
		}
		var leftCombatAt time.Duration
		inCombat := false
		StartPeriodicAction(sim, PeriodicActionOptions{
			Period: period,
			OnAction: func(sim *Simulation) {
				wasInCombat := inCombat
				inCombat = s.state == levelingCombat
				if !inCombat && wasInCombat {
					// Some time since the last tick.
					leftCombatAt = max(sim.CurrentTime-period, s.stateSince)
				}
				rage := character.CurrentRage()
				if inCombat || rage <= 0 || sim.CurrentTime < leftCombatAt+opts.Delay {
					sim.NothingHappened()
					return
				}
				decay := rageDecayPerSecond * opts.Multiplier * period.Seconds()
				character.SpendRage(sim, min(decay, rage), opts.Metrics)
			},
		})
	})
	character.RegisterLevelingDeathEffect(func(sim *Simulation) {
		if rage := character.CurrentRage(); rage > 0 {
			character.SpendRage(sim, rage, opts.Metrics)
		}
	})
}
