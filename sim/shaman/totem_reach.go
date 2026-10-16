package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Totems in a leveling session (core/leveling_world.go), where we move around and a totem
// stays where we dropped it.
//
// A buff totem's buff reaches 30 yards around it (Strength of Earth, Stoneskin, Grace of
// Air, Windfury, Flametongue, Mana Spring, Healing Stream and Mana Tide Totem, the client's
// SpellRadius, build 70338). Once we're further away we lose the buff, and get it back when we
// come back, as long as the totem stands. The game checks area auras about once a second, and
// so do we. Mana Spring's buff comes and goes the same way, and Healing Stream and Mana Tide
// give nothing on a pulse while we're out of reach. Outside a session we never move, so every
// totem reaches us.
//
// Our totems vanish when we die. Totemic Recall takes them all back for a quarter of their
// mana, and Totemic Projection moves them all to where we stand.

const (
	totemBuffRadius = 30
	// How often an area aura checks who's in reach.
	totemReachCheck = time.Second
)

// totemReach turns a totem's buff on us on and off as we come and go.
type totemReach struct {
	on, off func(sim *core.Simulation)
	inReach bool
}

// inTotemReach is whether we're within the slot's totem's buff radius.
func (shaman *Shaman) inTotemReach(sim *core.Simulation, slot int) bool {
	return sim.Encounter.LevelingPosition(sim).Dist(shaman.totemPositions[slot]) <= totemBuffRadius
}

// setTotemReach says how the totem just put down in the slot turns its buff on and off.
func (shaman *Shaman) setTotemReach(slot int, on, off func(sim *core.Simulation)) {
	shaman.totemReaches[slot] = &totemReach{on: on, off: off, inReach: true}
}

// setTotemAuraReach is setTotemReach for a totem whose buff is an aura on us that lasts as
// long as the totem.
func (shaman *Shaman) setTotemAuraReach(slot int, aura *core.Aura) {
	if aura == nil || aura.Duration == core.NeverExpires {
		return
	}
	shaman.setTotemReach(slot,
		func(sim *core.Simulation) {
			aura.Activate(sim)
			aura.UpdateExpires(sim, shaman.TotemExpirations[slot])
		},
		func(sim *core.Simulation) {
			aura.Deactivate(sim)
		})
}

// checkTotemReach turns each standing totem's buff on or off by where we are.
func (shaman *Shaman) checkTotemReach(sim *core.Simulation) {
	for slot, reach := range shaman.totemReaches {
		if reach == nil || !shaman.totemIsDown(sim, slot) {
			continue
		}
		if inReach := shaman.inTotemReach(sim, slot); inReach != reach.inReach {
			reach.inReach = inReach
			if inReach {
				reach.on(sim)
			} else {
				reach.off(sim)
			}
		}
	}
}

// destroyTotems ends every totem we have down, as when we die or recall them.
func (shaman *Shaman) destroyTotems(sim *core.Simulation) {
	for slot := range shaman.ActiveTotems {
		if end := shaman.totemEnds[slot]; end != nil {
			shaman.totemEnds[slot] = nil
			end(sim)
		}
		shaman.ActiveTotems[slot] = nil
		shaman.TotemExpirations[slot] = 0
		shaman.totemReaches[slot] = nil
		shaman.totemCosts[slot] = 0
	}
}

func (shaman *Shaman) anyTotemDown(sim *core.Simulation) bool {
	for slot := range shaman.ActiveTotems {
		if shaman.totemIsDown(sim, slot) {
			return true
		}
	}
	return false
}

// registerTotemReach starts the once a second reach check in a leveling session, and has
// our totems vanish when we die there.
func (shaman *Shaman) registerTotemReach() {
	if !shaman.Env.Encounter.IsLeveling() {
		return
	}
	shaman.RegisterResetEffect(func(sim *core.Simulation) {
		core.StartPeriodicAction(sim, core.PeriodicActionOptions{
			Period:   totemReachCheck,
			OnAction: shaman.checkTotemReach,
		})
	})
	shaman.RegisterLevelingDeathEffect(shaman.destroyTotems)
}

// Totemic Recall (36936, level 20) takes back every totem we have down, and gives us a
// quarter of the mana each cost (effect 110, 25; Forever's 2026-10-01 development notes: it
// now restores mana). No cost, a 1.5 sec global cooldown.
func (shaman *Shaman) registerTotemicRecallSpell() {
	if shaman.Level < 20 {
		return
	}
	actionID := core.ActionID{SpellID: 36936}
	manaMetrics := shaman.NewManaMetrics(actionID)
	shaman.TotemicRecall = shaman.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    SpellFlagShaman | core.SpellFlagAPL,

		RequiredLevel: 20,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, _ *core.Unit) bool {
			return shaman.anyTotemDown(sim)
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			refund := 0.0
			for slot := range shaman.ActiveTotems {
				if shaman.totemIsDown(sim, slot) {
					refund += 0.25 * shaman.totemCosts[slot]
				}
			}
			shaman.destroyTotems(sim)
			if refund > 0 {
				shaman.AddMana(sim, refund, manaMetrics)
			}
		},
	})
}

// Totemic Projection (437009, level 22) moves every totem we have down to where we stand, each
// as far from us and in the same direction as when we dropped it.
// 25% of base mana, 1 min cooldown, 1.5 sec global cooldown. Totemic Focus discounts it (its
// mask covers it, shaman_audit.md 8.36). The totems keep their time left.
func (shaman *Shaman) registerTotemicProjectionSpell() {
	if shaman.Level < 22 {
		return
	}
	shaman.TotemicProjection = shaman.RegisterSpell(core.SpellConfig{
		ActionID: core.ActionID{SpellID: 437009},
		Flags:    SpellFlagShaman | core.SpellFlagAPL,

		RequiredLevel: 22,

		ManaCost: core.ManaCostOptions{
			BaseCost: 0.25,
			Factor:   shaman.totemicFocusFactor(),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    shaman.NewTimer(),
				Duration: time.Minute,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, _ *core.Unit) bool {
			return shaman.anyTotemDown(sim)
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			// The totems keep their places around us.
			here := sim.Encounter.LevelingPosition(sim)
			for slot := range shaman.ActiveTotems {
				if shaman.totemIsDown(sim, slot) {
					shaman.setTotemPosition(sim, slot, shaman.ActiveTotems[slot], here.Add(shaman.totemOffsets[slot]))
				}
			}
			shaman.checkTotemReach(sim)
		},
	})
}
