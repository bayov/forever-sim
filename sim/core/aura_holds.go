package core

import (
	"slices"
	"time"
)

// Buffs that raid members keep up with something they cast, which identical casts from
// several members share, like two shamans' Strength of Earth Totems on one receiver. The buff
// is one aura on the receiver, and each caster holds it until their own effect ends. It stays
// while one of them still holds it.

type auraHold struct {
	holder *Unit
	until  time.Duration
}

// AddHold records that holder keeps the aura up until the time, in place of its earlier hold,
// and returns the latest time any holder keeps it up.
func (aura *Aura) AddHold(sim *Simulation, holder *Unit, until time.Duration) time.Duration {
	aura.RemoveHold(sim, holder)
	aura.holds = append(aura.holds, auraHold{holder: holder, until: until})
	return aura.latestHold()
}

// RemoveHold ends holder's hold, and returns the latest time a remaining holder keeps the aura
// up, 0 when none does.
func (aura *Aura) RemoveHold(sim *Simulation, holder *Unit) time.Duration {
	aura.holds = slices.DeleteFunc(aura.holds, func(hold auraHold) bool {
		return hold.holder == holder || hold.until <= sim.CurrentTime
	})
	return aura.latestHold()
}

func (aura *Aura) latestHold() time.Duration {
	latest := time.Duration(0)
	for _, hold := range aura.holds {
		latest = max(latest, hold.until)
	}
	return latest
}

// HoldUntil puts the buff up for holder until the time, or keeps it up longer when another
// caster's hold ends later. A permanent one, from the settings, is only refreshed: the given
// buff and the cast one share the aura.
func (aura *Aura) HoldUntil(sim *Simulation, holder *Unit, until time.Duration) {
	aura.Activate(sim)
	if aura.IsPermanent() || !aura.IsActive() {
		return
	}
	if latest := aura.AddHold(sim, holder, until); latest != aura.expires {
		aura.UpdateExpires(sim, latest)
	}
}

// ReleaseHold ends holder's hold. The buff goes, unless another caster still holds it, and a
// permanent one stays.
func (aura *Aura) ReleaseHold(sim *Simulation, holder *Unit) {
	if aura.IsPermanent() {
		return
	}
	if latest := aura.RemoveHold(sim, holder); latest > sim.CurrentTime && aura.IsActive() {
		if latest != aura.expires {
			aura.UpdateExpires(sim, latest)
		}
		return
	}
	aura.Deactivate(sim)
}
