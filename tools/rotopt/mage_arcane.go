package main

import "github.com/wowsims/classic/sim/core/proto"

// Arcane under Forever rules. Arcane Blast stacks to 4 for 10% more damage on the mage's other
// spells per stack, but each stack makes the next Arcane Blast cost 175% more. Arcane Missiles
// ends the stacks and takes their bonus, and Missile Barrage (40% per Arcane Blast) makes it
// free and twice as fast. So the rotation builds stacks, then spends them on Arcane Missiles,
// and the knobs are how many stacks, whether the missiles wait for Missile Barrage, and when
// the mana says to stop stacking.
type mageArcane struct{}

func init() {
	templates["mage_arcane"] = mageArcane{}
}

func (mageArcane) Knobs() []Knob {
	return append(mageCommonKnobs(),
		// Arcane Blast up to this many stacks before Arcane Missiles.
		Knob{Name: "blastStacks", Default: 3, Min: 1, Max: 4, Step: 1},
		// Arcane Missiles only with Missile Barrage up (1), or at the stack count whatever
		// (0). With 1 and no proc, Arcane Blast goes on to 4 stacks and then waits on the filler.
		Knob{Name: "missilesNeedBarrage", Default: 0, Min: 0, Max: 1, Step: 1},
		// Below this much mana Arcane Blast only goes to one stack, 0 never.
		Knob{Name: "conserveMana", Default: 30, Min: 0, Max: 60, Step: 10},
		// The filler once Arcane Blast is done: 1 Arcane Missiles, 2 Frostbolt, 3 Fireball.
		Knob{Name: "filler", Default: 1, Min: 1, Max: 3, Step: 1},
		// Missile Barrage procs are used right away, before more Arcane Blasts.
		Knob{Name: "barrageFirst", Default: 1, Min: 0, Max: 1, Step: 1},
		// With barrageFirst, the Missile Barrage missiles wait for this many stacks, 0 the stack
		// target.
		Knob{Name: "barrageStacks", Default: 0, Min: 0, Max: 4, Step: 1},
		// Precast an Arcane Blast (2.5 sec cast) so it lands at the pull.
		Knob{Name: "prepull", Default: 1, Min: 0, Max: 1, Step: 1},
	)
}

func (mageArcane) Build(k Knobs) *proto.APLRotation {
	s := mageRanks()
	items := mageCooldowns(k, s)
	stacks := auraNumStacks(arcaneBlastAura)
	abKnown := auraIsKnown(arcaneBlastAura)
	maxStacks := num(k["blastStacks"])
	if k["conserveMana"] > 0 {
		// Below the floor the stack target is 1.
		lowMana := lt(currentManaPercent(), num(k["conserveMana"]/100))
		items = append(items,
			cast(s.arcaneMissiles, and(abKnown, lowMana, ge(stacks, num(1))), "Low on mana: Arcane Missiles after a single Arcane Blast."),
			cast(s.arcaneBlast, and(abKnown, lowMana, lt(stacks, num(1))), "Low on mana: one Arcane Blast."),
		)
	}
	if k["barrageFirst"] == 1 {
		barrageStacks := maxStacks
		if k["barrageStacks"] > 0 {
			barrageStacks = num(k["barrageStacks"])
		}
		items = append(items, cast(s.arcaneMissiles, and(abKnown, auraIsKnown(missileBarrage), auraIsActive(missileBarrage), ge(stacks, barrageStacks)), "Arcane Missiles on Missile Barrage once the stacks are up."))
	}
	if k["missilesNeedBarrage"] == 0 {
		items = append(items, cast(s.arcaneMissiles, and(abKnown, ge(stacks, maxStacks)), "Arcane Missiles at the stack target, it ends the stacks and takes their bonus."))
	}
	stackCap := maxStacks
	if k["missilesNeedBarrage"] == 1 {
		stackCap = num(4)
	}
	items = append(items, cast(s.arcaneBlast, and(abKnown, lt(stacks, stackCap)), "Arcane Blast up to the stack target."))
	filler := s.arcaneMissiles
	switch k["filler"] {
	case 2:
		filler = s.frostbolt
	case 3:
		filler = s.fireball
	}
	items = append(items, cast(filler, nil, "The filler."))
	rot := rotation(items...)
	if k["prepull"] == 1 {
		rot.PrepullActions = append(rot.PrepullActions, prepull(s.arcaneBlast, 2.5))
	}
	return rot
}
