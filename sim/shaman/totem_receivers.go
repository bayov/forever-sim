package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Our totems' buffs reach the raid's other players, or only our party's (docs/raid_sim.md, the
// provider table). Client 70685 makes Strength of Earth, Grace of Air and Windfury Totem raid
// area auras (effect 65) and Mana Spring a party one (effect 35). Flametongue Totem goes to the
// raid like Windfury. Mana Tide and Healing Stream already pay our party (water_totems.go).
// Pets get none of them until the launch checks say so, and the raid sim takes every player to
// be in reach.
//
// A receiver's buff is one aura that every shaman's identical totem shares, so two don't stack.
// Each totem holds it until the totem ends (core.Aura.HoldUntil), so the buff stays while a
// second shaman's totem still stands. We build the receivers' auras once every player has
// registered their own, so a shaman receiver's own totem buff is the one ours shares.
//
// With nobody else in the raid there's nothing to build, and a single-player sim runs as it
// did.

// A totem buff on one of the raid's other players.
type totemBuffReceiver struct {
	char *core.Character
	aura *core.Aura
}

type totemBuffReceivers []totemBuffReceiver

// newTotemBuffReceivers gives each of the raid's other players, or only our party's, the
// aura makeAura makes for them. makeAura may return nil for a player the buff does nothing for.
func (shaman *Shaman) newTotemBuffReceivers(partyOnly bool, makeAura func(char *core.Character) *core.Aura) *totemBuffReceivers {
	receivers := &totemBuffReceivers{}
	shaman.Env.RegisterPreFinalizeEffect(func() {
		parties := shaman.Env.Raid.Parties
		if partyOnly {
			parties = []*core.Party{shaman.Party}
		}
		for _, party := range parties {
			for _, agent := range party.Players {
				if _, ok := agent.(*core.TargetDummy); ok {
					continue
				}
				char := agent.GetCharacter()
				if char == &shaman.Character {
					continue
				}
				if aura := makeAura(char); aura != nil {
					*receivers = append(*receivers, totemBuffReceiver{char: char, aura: aura})
				}
			}
		}
	})
	return receivers
}

// hold puts each receiver's buff up until our totem ends.
func (receivers *totemBuffReceivers) hold(sim *core.Simulation, shaman *Shaman, until time.Duration) {
	for _, receiver := range *receivers {
		receiver.aura.HoldUntil(sim, &shaman.Unit, until)
	}
}

// release takes our totem's buff off each receiver, unless another shaman's totem holds it.
func (receivers *totemBuffReceivers) release(sim *core.Simulation, shaman *Shaman) {
	for _, receiver := range *receivers {
		receiver.aura.ReleaseHold(sim, &shaman.Unit)
	}
}

// holdWeapon gives each receiver our weapon totem's buff, unless their imbue blocks it
// (core.Character.HoldTotemWeaponBuff picks which weapon totem a receiver takes).
func (receivers *totemBuffReceivers) holdWeapon(sim *core.Simulation, shaman *Shaman, windfury bool) {
	for _, receiver := range *receivers {
		if !receiver.char.TotemWeaponBuffBlocked(windfury) {
			receiver.char.HoldTotemWeaponBuff(sim, receiver.aura, &shaman.Unit, windfury)
		}
	}
}

func (receivers *totemBuffReceivers) releaseWeapon(sim *core.Simulation, shaman *Shaman) {
	for _, receiver := range *receivers {
		receiver.char.ReleaseTotemWeaponBuff(sim, receiver.aura, &shaman.Unit)
	}
}

// holdTotemBuff puts our totem's buff up on us and the receivers until the totem in the slot
// ends.
func (shaman *Shaman) holdTotemBuff(sim *core.Simulation, slot int, aura *core.Aura, receivers *totemBuffReceivers) {
	aura.HoldUntil(sim, &shaman.Unit, shaman.TotemExpirations[slot])
	receivers.hold(sim, shaman, shaman.TotemExpirations[slot])
}

// endHeldTotemBuff ends a totem whose buff holdTotemBuff put up. The buff stays where another
// shaman's identical totem still holds it, and a permanent one from the raid buffs stays too
// (see endTotemBuff).
func (shaman *Shaman) endHeldTotemBuff(aura *core.Aura, receivers *totemBuffReceivers) func(sim *core.Simulation) {
	return func(sim *core.Simulation) {
		aura.ReleaseHold(sim, &shaman.Unit)
		receivers.release(sim, shaman)
	}
}
