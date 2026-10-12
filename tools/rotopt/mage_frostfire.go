package main

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/mage"
)

// Frostfire under Forever rules: Frostfire Bolt as the filler, a Fire and Frost spell at once, so
// it takes the talents of both trees. Its crits build Heating Up for Pyroblast, its chill rolls
// Fingers of Frost for Ice Lance, and it rolls Missile Barrage like Fireball. The Manaflare set's
// 5 piece bonus adds 10% to the last two and 10% crit during Combustion.
type mageFrostfire struct{}

func init() {
	templates["mage_frostfire"] = mageFrostfire{}
}

func (mageFrostfire) Knobs() []Knob {
	return append(mageCommonKnobs(),
		// Scorch up to 5 stacks of Fire Vulnerability and again with less than this many seconds
		// left, 0 never.
		Knob{Name: "scorchRefresh", Default: 5, Min: 0, Max: 9, Step: 3},
		// Pyroblast at this many Heating Up stacks, 0 never.
		Knob{Name: "pyroStacks", Default: 3, Min: 0, Max: 3, Step: 1},
		// Ice Lance while Fingers of Frost is up: 0 never, 1 every charge, 2 the last one.
		Knob{Name: "iceLance", Default: 1, Min: 0, Max: 2, Step: 1},
		// Pyroblast only above this much mana, 0 any time. A Pyroblast costs more than the filler,
		// so a long fight saves it for the Heating Up stacks the mana allows.
		Knob{Name: "pyroMana", Default: 0, Min: 0, Max: 60, Step: 10},
		// Arcane Missiles on a Missile Barrage proc.
		Knob{Name: "missileBarrage", Default: 1, Min: 0, Max: 1, Step: 1},
		// Fire Blast on cooldown above this much mana, 0 never.
		Knob{Name: "fireBlastMana", Default: 20, Min: 0, Max: 80, Step: 20},
		// Precast the filler so the first one lands at the pull.
		Knob{Name: "prepull", Default: 1, Min: 0, Max: 1, Step: 1},
		scorchFirstKnob(),
		// Rank 2 Frostfire Bolt (285 mana instead of 370, about 9% less damage) once Evocation is
		// spent or less than 60 sec are left, and the mana is under this many per second of the
		// fight left. 0 never.
		Knob{Name: "lowManaRank", Default: 0, Min: 0, Max: 90, Step: 30},
		// Fire Blast with the mana the fight can't use: above this many per second of the fight
		// left. 0 never.
		Knob{Name: "fireBlastSurplus", Default: 0, Min: 0, Max: 200, Step: 50},
	)
}

func (mageFrostfire) Build(k Knobs) *proto.APLRotation {
	s := mageRanks()
	items := mageCooldowns(k, s)
	if k["scorchRefresh"] > 0 {
		items = append(items, cast(s.scorch,
			and(targetAuraIsKnown(fireVulnerability), or(lt(targetAuraNumStacks(fireVulnerability), num(5)), lt(targetAuraRemainingTime(fireVulnerability), seconds(k["scorchRefresh"])))),
			"Scorch up to 5 stacks of Fire Vulnerability, and before it runs out."))
	}
	if p := k["pyroStacks"]; p > 0 {
		items = append(items, cast(s.pyroblast, and(known(heatingUp, ge(auraNumStacks(heatingUp), num(p))), manaAbove(k["pyroMana"])), "Pyroblast once Heating Up has cut its cast enough."))
	}
	switch k["iceLance"] {
	case 1:
		items = append(items, cast(s.iceLance, known(fingersOfFrost, auraIsActive(fingersOfFrost)), "Ice Lance while Fingers of Frost is up."))
	case 2:
		items = append(items, cast(s.iceLance, known(fingersOfFrost, eq(auraNumStacks(fingersOfFrost), num(1))), "Ice Lance on the last Fingers of Frost charge."))
	}
	if k["missileBarrage"] == 1 {
		items = append(items, cast(s.arcaneMissiles, known(missileBarrage, auraIsActive(missileBarrage)), "Arcane Missiles on a Missile Barrage proc, free and twice as fast."))
	}
	if k["fireBlastMana"] > 0 {
		items = append(items, cast(s.fireBlast, ge(currentManaPercent(), num(k["fireBlastMana"]/100)), "Fire Blast on cooldown with the mana to spare."))
	}
	if k["fireBlastSurplus"] > 0 {
		items = append(items, cast(s.fireBlast, and(
			gt(currentMana(), mul(remainingTime(), num(k["fireBlastSurplus"]))),
			gt(remainingTime(), seconds(1)),
		), "Fire Blast with the mana the fight can't use."))
	}
	if k["lowManaRank"] > 0 {
		rank2 := spellID(mage.FrostfireBoltSpellId[2])
		items = append(items, cast(rank2, and(
			or(not(spellIsReady(evocation)), lt(remainingTime(), seconds(60))),
			lt(currentMana(), mul(remainingTime(), num(k["lowManaRank"]))),
			gt(remainingTime(), seconds(10)),
		), "Rank 2 Frostfire Bolt when the mana will not last the fight."))
	}
	items = append(items, cast(s.frostfireBolt, nil, "Frostfire Bolt."))
	rot := rotation(items...)
	if k["prepull"] == 1 {
		rot.PrepullActions = append(rot.PrepullActions, prepull(s.frostfireBolt, 3))
	}
	return rot
}
