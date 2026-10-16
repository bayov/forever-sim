package main

import "github.com/wowsims/classic/sim/core/proto"

// Frost under Forever rules. Frostbolt is the filler. Fingers of Frost (15% on a chill) makes
// the next spells treat the boss as Frozen, which turns on Shatter and Ice Lance's 3 times damage,
// so the main decision is what to spend the charges on. Winter's Chill is a Cone of Cold debuff
// since the client 1.60.1.70425.
type mageFrost struct{}

func init() {
	templates["mage_frost"] = mageFrost{}
}

func (mageFrost) Knobs() []Knob {
	return append(append(append(mageCommonKnobs(), arcaneKnobs(0, 0)...), finisherKnobs(1)...),
		// Ice Lance while Fingers of Frost is up: 0 never, 1 every charge, 2 only the last
		// charge (the first goes on a Frostbolt).
		Knob{Name: "iceLance", Default: 1, Min: 0, Max: 2, Step: 1},
		// Arcane Missiles on a Missile Barrage proc (Frostbolt has 20%), for builds with the
		// Arcane talent.
		Knob{Name: "missileBarrage", Default: 1, Min: 0, Max: 1, Step: 1},
		// Fire Blast on cooldown above this much mana, 0 never.
		Knob{Name: "fireBlastMana", Default: 0, Min: 0, Max: 80, Step: 20},
		// Precast Frostbolt so the first one lands at the pull.
		Knob{Name: "prepull", Default: 1, Min: 0, Max: 1, Step: 1},
	)
}

func (mageFrost) Build(k Knobs) *proto.APLRotation {
	s := mageRanks()
	items := mageCooldowns(k, s)
	switch k["iceLance"] {
	case 1:
		items = append(items, cast(s.iceLance, known(fingersOfFrost, auraIsActive(fingersOfFrost)), "Ice Lance while Fingers of Frost makes the boss count as Frozen (3 times the damage, Shatter's crit)."))
	case 2:
		items = append(items, cast(s.iceLance, known(fingersOfFrost, eq(auraNumStacks(fingersOfFrost), num(1))), "Ice Lance on the last Fingers of Frost charge, the first goes on a Frostbolt."))
	}
	items = append(items, arcaneBlastBuffLines(k, s, nil)...)
	items = append(items, arcaneWrathLines(k, s)...)
	if k["missileBarrage"] == 1 {
		items = append(items, cast(s.arcaneMissiles, known(missileBarrage, auraIsActive(missileBarrage)), "Arcane Missiles on a Missile Barrage proc, free and twice as fast."))
	}
	if k["fireBlastMana"] > 0 {
		items = append(items, cast(s.fireBlast, ge(currentManaPercent(), num(k["fireBlastMana"]/100)), "Fire Blast on cooldown with the mana to spare."))
	}
	items = append(items, finisherLines(k, s.frostbolt, s.iceLance)...)
	items = append(items, cast(s.frostbolt, nil, "Frostbolt."))
	rot := rotation(items...)
	if k["prepull"] == 1 {
		rot.PrepullActions = append(rot.PrepullActions, prepull(s.frostbolt, 2.5))
	}
	return rot
}
