package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Totems take a 1 sec global cooldown, not the 1.5 sec of other spells (wowhead's Forever
// spell pages, Searing and Flametongue Totem among them).
const totemGCD = time.Second

func (shaman *Shaman) newTotemSpellConfig(flatCost float64, spellID int32) core.SpellConfig {
	return core.SpellConfig{
		ActionID: core.ActionID{SpellID: spellID},
		Flags:    SpellFlagShaman | SpellFlagTotem | core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			FlatCost: flatCost,
			Factor:   shaman.totemicFocusFactor(),
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: totemGCD,
			},
		},
	}
}

// placeTotem puts the totem down in its element's slot.
//
// We have at most one totem of each element (earth, air, fire and water) down at a time,
// because the client summons each totem into its element's slot. A new totem takes the place
// of the one in its slot, so we end what that one was doing: its buff on us, its attacks, its
// heals or its mana ticks. Putting down the same totem again ends the old one too, and the new
// one starts over.
//
// end ends this totem's effects when the next totem in the slot comes down. It can be nil
// for a totem that does nothing in the sim, like Tremor Totem.
func (shaman *Shaman) placeTotem(sim *core.Simulation, slot int, spell *core.Spell, duration time.Duration, end func(sim *core.Simulation)) {
	if previous := shaman.totemEnds[slot]; previous != nil {
		shaman.totemEnds[slot] = nil
		previous(sim)
	}
	shaman.TotemExpirations[slot] = sim.CurrentTime + duration
	shaman.ActiveTotems[slot] = spell
	shaman.totemEnds[slot] = end
	shaman.totemPositions[slot] = sim.Encounter.LevelingPosition(sim)
}

// Fire totem ranges, from the client's SpellRange and SpellRadius (build 70338): Searing
// Totem's bolt reaches 20 yards (3606), Magma Totem's pulse 8 yards around it (8187), and
// Fire Nova goes off 10 yards around the totem (408423), which has to be within 30 yards of
// us to cast it (408341).
const (
	searingTotemRange = 20
	magmaTotemRadius  = 8
	fireNovaRadius    = 10
	fireNovaCastRange = 30
)

// nearFireTotem is whether the mob is within yards of our fire totem. Only a leveling
// session moves us and the mobs around (core/leveling.go); in any other fight every mob is.
func (shaman *Shaman) nearFireTotem(sim *core.Simulation, mob *core.Unit, yards float64) bool {
	return sim.Encounter.LevelingDistanceFrom(sim, shaman.totemPositions[FireTotem], mob) <= yards
}

// fireTotemMobs are the mobs fighting us within yards of our fire totem.
func (shaman *Shaman) fireTotemMobs(sim *core.Simulation, yards float64) []*core.Unit {
	if !sim.Encounter.IsLeveling() {
		return sim.Encounter.ActiveTargetUnits
	}
	return core.FilterSlice(sim.Encounter.ActiveTargetUnits, func(mob *core.Unit) bool {
		return shaman.nearFireTotem(sim, mob, yards)
	})
}

// searingTotemTarget is the mob Searing Totem shoots: our target, or in a leveling session
// the first mob fighting us in its range when our target isn't one. Nil when none is.
func (shaman *Shaman) searingTotemTarget(sim *core.Simulation) *core.Unit {
	if !sim.Encounter.IsLeveling() {
		return shaman.CurrentTarget
	}
	mobs := shaman.fireTotemMobs(sim, searingTotemRange)
	for _, mob := range mobs {
		if mob == shaman.CurrentTarget {
			return mob
		}
	}
	if len(mobs) > 0 {
		return mobs[0]
	}
	return nil
}

// endTotemBuff ends a totem whose effect is a buff aura on us.
//
// Another shaman's totem in the raid buffs shares the aura and makes it permanent. We leave
// that one up, because our totem never took it away from the other shaman.
func endTotemBuff(aura *core.Aura) func(sim *core.Simulation) {
	return func(sim *core.Simulation) {
		if aura.Duration != core.NeverExpires {
			aura.Deactivate(sim)
		}
	}
}

// ownTotemAura keeps the buff of a totem we put down ourselves off the stats panel.
//
// The panel turns on every aura with a build phase, so our own Strength of Earth and Grace
// of Air showed whenever the shaman knew the spell, even when the rotation never casts it.
// The Level 60 preset puts down Windfury Totem, yet the panel showed Grace of Air's 89
// Agility. The fight only gets the buff once we cast the totem.
//
// When another shaman's totem is picked in the raid buffs, the aura is already permanent
// by the time we get it, and we leave it on the panel.
func ownTotemAura(aura *core.Aura) *core.Aura {
	if aura.Duration != core.NeverExpires {
		aura.BuildPhase = core.CharacterBuildPhaseNone
	}
	return aura
}
