package main

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/mage"
)

// Arcane under Forever rules (client 1.60.1.70425). Arcane Blast's stacks are spell damage now:
// 15% of Intellect each, up to 4, for 15 sec from the first stack, and each stack makes the next
// Arcane Blast cost 175% more. Nothing ends them early. So the rotation buys a few stacks at the
// start of each 15 sec, then casts Arcane Wrath on cooldown (3.5 sec, 6 sec cooldown, 5% of
// maximum mana as extra spell damage), Arcane Missiles on Missile Barrage, and the filler. The
// knobs are how many stacks, the filler and when the mana says to stop stacking.
type mageArcane struct{}

func init() {
	templates["mage_arcane"] = mageArcane{}
}

func (mageArcane) Knobs() []Knob {
	return append(append(append(mageCommonKnobs(), finisherKnobs(0)...), arcaneKnobs(2, 1)...),
		// Below this much mana Arcane Blast only goes to one stack, 0 never.
		Knob{Name: "conserveMana", Default: 30, Min: 0, Max: 60, Step: 10},
		// The filler: 1 Arcane Missiles, 2 Frostbolt, 3 Fireball.
		Knob{Name: "filler", Default: 1, Min: 1, Max: 3, Step: 1},
		// Arcane Missiles on a Missile Barrage proc.
		Knob{Name: "missileBarrage", Default: 1, Min: 0, Max: 1, Step: 1},
		// Missile Barrage missiles ahead of Arcane Wrath.
		Knob{Name: "barrageFirst", Default: 0, Min: 0, Max: 1, Step: 1},
		// Missile Barrage missiles ahead of the Arcane Blasts, so a proc during the stacks is used
		// before the next Arcane Blast can roll another over it.
		Knob{Name: "barrageBeforeBlast", Default: 0, Min: 0, Max: 1, Step: 1},
		// Precast an Arcane Blast (1.5 sec cast) so it lands at the pull.
		Knob{Name: "prepull", Default: 1, Min: 0, Max: 1, Step: 1},
		// How many seconds before the pull that Arcane Blast starts. It hits as its cast ends, and
		// damage before the pull doesn't count, so with haste it must start later: 1.47 for the
		// High Order's 2% (a 1.4706 sec cast), where 1.5 lands it 30 ms before the pull.
		Knob{Name: "prepullAt", Default: 1.5, Min: 1.47, Max: 1.5, Step: 0.03},
		// Hold Arcane Power and Presence of Mind for the synthetic on-use trinket.
		Knob{Name: "powerWithTrinket", Default: 0, Min: 0, Max: 1, Step: 1},
		// Presence of Mind on Arcane Wrath, the longest cast.
		Knob{Name: "pomWrath", Default: 0, Min: 0, Max: 1, Step: 1},
	)
}

func (mageArcane) Build(k Knobs) *proto.APLRotation {
	s := mageRanks()
	var items []*proto.APLListItem
	if k["pomWrath"] == 1 {
		items = append(items,
			cast(s.arcaneWrath, known(presenceOfMind, auraIsActive(presenceOfMind)), "Arcane Wrath while Presence of Mind is up. Without the talents ignore the warning on this line."),
			cast(presenceOfMind, spellIsReady(s.arcaneWrath), "Presence of Mind for an instant Arcane Wrath. Without the talents ignore the warning on this line."),
		)
	}
	items = append(items, mageCooldowns(k, s)...)
	filler := s.arcaneMissiles
	switch k["filler"] {
	case 2:
		filler = s.frostbolt
	case 3:
		filler = s.fireball
	}
	arcaneExplosion := rankID(mage.ArcaneExplosionLevel[:], mage.ArcaneExplosionSpellId[:])
	items = append(items, finisherLines(k, filler, s.fireBlast, arcaneExplosion)...)
	barrage := []*proto.APLListItem{}
	if k["missileBarrage"] == 1 {
		barrage = append(barrage, cast(s.arcaneMissiles, known(missileBarrage, auraIsActive(missileBarrage)), "Arcane Missiles on a Missile Barrage proc, free and twice as fast."))
	}
	if k["barrageBeforeBlast"] == 1 {
		items = append(items, barrage...)
		barrage = nil
	}
	var enoughMana value
	if k["conserveMana"] > 0 && k["blastBuff"] > 1 {
		// Below the floor the stack target is 1.
		lowMana := lt(currentManaPercent(), num(k["conserveMana"]/100))
		items = append(items, cast(s.arcaneBlast, and(auraIsKnown(arcaneBlastAura), lowMana, not(auraIsActive(arcaneBlastAura)), gt(remainingTime(), seconds(4))), "Low on mana: one Arcane Blast for the spell damage."))
		enoughMana = not(lowMana)
	}
	items = append(items, arcaneBlastBuffLines(k, s, enoughMana)...)
	if k["barrageFirst"] == 1 {
		items = append(items, barrage...)
		items = append(items, arcaneWrathLines(k, s)...)
	} else {
		items = append(items, arcaneWrathLines(k, s)...)
		items = append(items, barrage...)
	}
	items = append(items, cast(filler, nil, "The filler."))
	items = append(items, wandLines(k)...)
	rot := rotation(items...)
	if k["prepull"] == 1 {
		rot.PrepullActions = append(rot.PrepullActions, prepull(s.arcaneBlast, k["prepullAt"]))
	}
	return rot
}
