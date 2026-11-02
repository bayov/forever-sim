//go:build practice

package practice

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// How the world talks to the sim. Every write happens inside a session action, because the
// core's auto attack and cast calls use the current sim time.

func (r *run) targetPos(sim *core.Simulation) vec2 {
	return r.target.posAt(sim.CurrentTime)
}

// move takes a move input: where you are, how you move from now on, and how fast you turn.
func (r *run) move(sim *core.Simulation, in Input) {
	now := sim.CurrentTime
	r.settleWorld(sim)
	wasMoving := r.me.moving()
	r.me.pos = vec2(in.Pos)
	r.me.vel = vec2(in.Velocity)
	r.me.facing = normalizeAngle(in.Facing)
	r.me.turnRate = in.TurnRate
	r.me.at = now
	// Moving stops a cast. Turning in place doesn't.
	if r.me.moving() && !wasMoving {
		r.interruptCast(sim)
	}
	r.updateWorld(sim)
}

// settleWorld counts the time out of range up to now.
func (r *run) settleWorld(sim *core.Simulation) {
	now := sim.CurrentTime
	if now > r.lastWorldAt {
		if r.s.pulled && !r.inMelee {
			from := max(r.lastWorldAt, 0)
			if now > from {
				r.outOfRange += now - from
			}
		}
		r.lastWorldAt = now
	}
}

// updateWorld works out range and facing now and tells the sim: your distance, whether you're
// in front of the dummy, and whether your auto attack swings.
func (r *run) updateWorld(sim *core.Simulation) {
	r.settleWorld(sim)
	now := sim.CurrentTime
	me, target := r.me.posAt(now), r.targetPos(sim)
	dist := me.dist(target)
	melee := meleeRange(&r.me, &r.target)

	r.inMelee = r.hasTarget && dist <= melee
	// The core knows melee range as a flat 5 yards. We give it the distance that makes its rule
	// agree with ours.
	r.char.DistanceFromTarget = max(0, dist-(melee-nominalMeleeRange))
	if !r.hasTarget {
		r.char.DistanceFromTarget = 100
	}
	// Parry and block happen only from the front of the dummy.
	r.char.PseudoStats.InFrontOfTarget = inArc(target, r.target.facingAt(now), me, spellArc)

	swing := r.attackOn && r.inMelee && inArc(me, r.me.facingAt(now), target, swingArc)
	if swing != r.swinging {
		r.swinging = swing
		if swing {
			r.char.AutoAttacks.EnableAutoSwing(sim)
			// Before the pull, auto attack on in range is the pull: the first swing comes at once.
			if now < 0 && r.pullAt == nil {
				at := now
				r.pullAt = &at
			}
		} else {
			r.char.AutoAttacks.CancelAutoSwing(sim)
		}
	}

	r.keepTicking(sim)
}

// keepTicking runs a world tick every 50ms while you move or turn.
//
// Before the pull with auto attack off, a tick can't change anything the sim sees: range and
// facing only matter to swings and hits. So we skip them then, which keeps the pull's swap
// quick after a long prepull of walking around (spike 4).
func (r *run) keepTicking(sim *core.Simulation) {
	active := r.me.moving() || r.me.turnRate != 0
	needed := r.attackOn || r.s.pulled || sim.CurrentTime >= 0
	if !active || !needed || r.ticking {
		return
	}
	r.ticking = true
	var tick func(sim *core.Simulation)
	tick = func(sim *core.Simulation) {
		r.ticking = false
		r.updateWorld(sim)
	}
	r.enqueueWorld(sim.CurrentTime+worldTick, tick)
}

func (r *run) unitStates() []UnitState {
	now := r.sim.CurrentTime
	me := r.me.posAt(now)
	target := r.target.posAt(now)
	states := []UnitState{
		{Ref: UnitRef{Kind: "player"}, Pos: me, Facing: normalizeAngle(r.me.facingAt(now)), Reach: r.me.reach, Height: r.me.height},
		{Ref: UnitRef{Kind: "enemy"}, Pos: target, Facing: r.target.facingAt(now), Reach: r.target.reach, Height: r.target.height},
	}
	if r.me.airborne(now) {
		states[0].AirborneUntil = toMs(r.me.airborneUntil)
	}
	return states
}

func (r *run) playerState() *PlayerState {
	sim := r.sim
	now := sim.CurrentTime
	c := r.char
	p := &PlayerState{
		GCDStart:       toMs(r.gcdStart),
		GCDEnd:         toMs(c.GCD.ReadyAt()),
		AttackOn:       r.attackOn,
		Swinging:       r.swinging,
		InMeleeRange:   r.inMelee,
		Moving:         r.me.moving(),
		HasTarget:      r.hasTarget,
		DamageDone:     r.damage,
		OutOfRangeTime: toMs(r.outOfRange),
	}
	if c.HasManaBar() {
		p.Mana = c.CurrentMana()
		p.MaxMana = c.MaxMana()
	}
	p.MaxHealth = c.MaxHealth()
	p.Health = p.MaxHealth
	if c.HasHealthBar() {
		p.Health = c.CurrentHealth()
	}
	if mh := c.AutoAttacks.MH(); mh != nil {
		p.SwingMs = toMs(c.AutoAttacks.MainhandSwingSpeed())
	}
	if r.swinging {
		p.NextSwingAt = toMs(c.AutoAttacks.MainhandSwingAt())
	}
	if c.IsCasting(sim) {
		hc := c.Hardcast
		p.Cast = &Cast{SpellID: hc.ActionID.SpellID, Start: toMs(hc.Start), End: toMs(hc.Expires)}
	}
	if r.queued != nil && !r.queued.done && r.queuedSpell != nil {
		p.QueuedSpellID = r.queuedSpell.ActionID.SpellID
	}
	for _, aura := range c.GetAuras() {
		if state, ok := auraState(aura, now); ok {
			p.Auras = append(p.Auras, state)
		}
	}
	for id, spell := range r.toolkit {
		if ready := spell.ReadyAt(); !spell.IsReady(sim) {
			p.Cooldowns = append(p.Cooldowns, Cooldown{SpellID: id, ReadyAt: toMs(ready), Length: toMs(spell.CD.Duration)})
		}
	}
	for id, spell := range r.items {
		if ready := spell.ReadyAt(); !spell.IsReady(sim) {
			p.Cooldowns = append(p.Cooldowns, Cooldown{ItemID: id, ReadyAt: toMs(ready), Length: toMs(spell.CD.Duration)})
		}
	}
	if r.shaman != nil {
		for slot, spell := range r.shaman.ActiveTotems {
			if spell == nil || r.shaman.TotemExpirations[slot] <= now {
				continue
			}
			p.Totems = append(p.Totems, TotemState{
				Element:   []string{"air", "earth", "fire", "water"}[slot],
				SpellID:   spell.ActionID.SpellID,
				ExpiresAt: toMs(r.shaman.TotemExpirations[slot]),
				Pos:       r.totemPos[slot],
			})
		}
		p.MainHandImbue = r.shaman.ActiveImbue().String()
	}
	return p
}

// auraState is an aura the HUD shows: active, not a hidden helper, not expired.
func auraState(aura *core.Aura, now time.Duration) (AuraState, bool) {
	if !aura.IsActive() || aura.ActionID.IsEmptyAction() {
		return AuraState{}, false
	}
	if aura.ActionID.SpellID == 0 && aura.ActionID.ItemID == 0 {
		return AuraState{}, false
	}
	if aura.Label == "Practice Watcher" {
		return AuraState{}, false
	}
	expires := aura.ExpiresAt()
	if expires != core.NeverExpires && expires <= now {
		return AuraState{}, false
	}
	state := AuraState{
		SpellID:   aura.ActionID.SpellID,
		ItemID:    aura.ActionID.ItemID,
		Label:     aura.Label,
		Stacks:    aura.GetStacks(),
		StartedAt: toMs(aura.StartedAt()),
	}
	if expires != core.NeverExpires {
		state.ExpiresAt = toMs(expires)
		state.Duration = toMs(aura.Duration)
	}
	return state, true
}

func (r *run) targetState() *TargetState {
	now := r.sim.CurrentTime
	t := &TargetState{Name: r.dummy.Label, Level: r.dummy.Level, DamageTaken: r.damage}
	for _, aura := range r.dummy.GetAuras() {
		if aura.IsPermanent() {
			continue
		}
		if state, ok := auraState(aura, now); ok {
			t.Debuffs = append(t.Debuffs, state)
		}
	}
	return t
}
