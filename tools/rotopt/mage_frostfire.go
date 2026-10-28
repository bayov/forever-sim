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
	return append(append(append(mageCommonKnobs(), arcaneKnobs(0, 0)...), finisherKnobs(1.5)...),
		// Scorch up to 5 stacks of Fire Vulnerability and again with less than this many seconds
		// left, 0 never.
		Knob{Name: "scorchRefresh", Default: 5, Min: 0, Max: 9, Step: 3},
		// Pyroblast at this many Heating Up stacks, 0 never.
		Knob{Name: "pyroStacks", Default: 3, Min: 0, Max: 3, Step: 1},
		// Ice Lance while Fingers of Frost is up: 0 never, 1 every charge, 2 the last one, 3 right
		// after a Frostfire Bolt completes (both read the target as Frozen), 4 the same on the
		// last charge only.
		Knob{Name: "iceLance", Default: 1, Min: 0, Max: 4, Step: 1},
		// The Fingers of Frost Ice Lance ahead of Scorch and Pyroblast.
		Knob{Name: "iceLanceFirst", Default: 0, Min: 0, Max: 1, Step: 1},
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
		trinketAfterScorchKnob(),
		racialAfterScorchKnob(),
		// Rank 2 Frostfire Bolt (285 mana instead of 370, about 9% less damage) once Evocation is
		// spent or no longer worth it (evocationMinTime sec or less left, the cutoff the
		// Evocation line uses), and the mana is under this many per second of the fight left.
		// 0 never.
		Knob{Name: "lowManaRank", Default: 0, Min: 0, Max: 90, Step: 30},
		// Fire Blast with the mana the fight can't use: above this many per second of the fight
		// left. 0 never.
		Knob{Name: "fireBlastSurplus", Default: 0, Min: 0, Max: 200, Step: 50},
		cooldownWaitKnob(),
		// Hold a cast shorter than the 1.5 sec GCD while Fingers of Frost is up: 0 never, 1
		// Pyroblast, 2 Pyroblast and the Scorch refresh (while Fire Vulnerability has its 5 stacks
		// and more than 4 sec left, so the hold never lets it drop). Unhasted, Pyroblast at 3
		// Heating Up stacks and Scorch are 1.5 sec and end with the GCD, so the Ice Lance on the
		// last charge goes out in the same spell batch, before the cast spends the charge (a
		// batch after it completes), and the charge buys both Frozen. With haste (Skyborne Wind
		// Blessed, Berserking) they end before the GCD and spend the last charge before Ice Lance
		// can go: the Ice Lance is lost. Held, a Frostfire Bolt (longer than the GCD) takes the
		// charge and the Ice Lance its batch. 3 also keeps the Ice Lance on the last charge
		// (iceLance 2) for a Frostfire Bolt's batch while hasted, where it would otherwise go
		// alone a moment after a Scorch or Pyroblast. Unhasted the lines never hold.
		Knob{Name: "lanceHold", Default: 0, Min: 0, Max: 3, Step: 1},
	)
}

// lanceHoldCond lets spell go unless it would end before the GCD with Fingers of Frost up and
// also (nil for always) holds (the lanceHold knob), nil (no condition) when the knob doesn't
// cover it.
func lanceHoldCond(k Knobs, spell *proto.ActionID, level float64, also value) value {
	if k["lanceHold"] < level {
		return nil
	}
	return not(known(fingersOfFrost, and(auraIsActive(fingersOfFrost), lt(spellCastTime(spell), seconds(1.49)), also)))
}

func (mageFrostfire) Build(k Knobs) *proto.APLRotation {
	s := mageRanks()
	items := mageCooldowns(k, s)
	iceLance := func() {
		switch k["iceLance"] {
		case 1:
			items = append(items, cast(s.iceLance, known(fingersOfFrost, auraIsActive(fingersOfFrost)), "Ice Lance while Fingers of Frost is up."))
		case 2:
			last := eq(auraNumStacks(fingersOfFrost), num(1))
			if k["lanceHold"] >= 3 {
				// Hasted (Scorch under the GCD), only in a Frostfire Bolt's batch, either rank.
				rank2 := spellID(mage.FrostfireBoltSpellId[2])
				last = and(last, or(not(lt(spellCastTime(s.scorch), seconds(1.49))), spellInFlight(s.frostfireBolt), spellInFlight(rank2)))
			}
			items = append(items, cast(s.iceLance, known(fingersOfFrost, last), "Ice Lance on the last Fingers of Frost charge."))
		default:
			items = append(items, iceLanceAfterBolt(k, s.frostfireBolt, s.iceLance)...)
		}
	}
	if k["iceLanceFirst"] == 1 {
		iceLance()
	}
	if k["scorchRefresh"] > 0 {
		items = append(items, cast(s.scorch,
			and(targetAuraIsKnown(fireVulnerability), or(lt(targetAuraNumStacks(fireVulnerability), num(5)), lt(targetAuraRemainingTime(fireVulnerability), seconds(k["scorchRefresh"]))),
				lanceHoldCond(k, s.scorch, 2, and(ge(targetAuraNumStacks(fireVulnerability), num(5)), gt(targetAuraRemainingTime(fireVulnerability), seconds(4))))),
			"Scorch up to 5 stacks of Fire Vulnerability, and before it runs out."))
	}
	if p := k["pyroStacks"]; p > 0 {
		items = append(items, cast(s.pyroblast, and(known(heatingUp, ge(auraNumStacks(heatingUp), num(p))), manaAbove(k["pyroMana"]), lanceHoldCond(k, s.pyroblast, 1, nil)), "Pyroblast once Heating Up has cut its cast enough."))
	}
	if k["iceLanceFirst"] != 1 {
		iceLance()
	}
	items = append(items, arcaneBlastBuffLines(k, s, nil)...)
	items = append(items, arcaneWrathLines(k, s)...)
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
		// Once Evocation won't refill the mana: it is on cooldown, or the fight is past
		// evocationMinTime, the Evocation line's own cutoff, so the two lines make one decision
		// and there is no stretch where neither fires (full-rank bolts drain the mana and the
		// wand finishes).
		rank2 := spellID(mage.FrostfireBoltSpellId[2])
		items = append(items, cast(rank2, and(
			or(not(spellIsReady(evocation)), le(remainingTime(), seconds(k["evocationMinTime"]))),
			lt(currentMana(), mul(remainingTime(), num(k["lowManaRank"]))),
			gt(remainingTime(), seconds(10)),
		), "Rank 2 Frostfire Bolt when the mana will not last the fight."))
	}
	items = append(items, finisherLines(k, s.frostfireBolt, s.iceLance)...)
	var blastConds []value
	if k["fireBlastMana"] > 0 {
		blastConds = append(blastConds, ge(currentManaPercent(), num(k["fireBlastMana"]/100)))
	}
	if k["fireBlastSurplus"] > 0 {
		blastConds = append(blastConds, gt(currentMana(), mul(remainingTime(), num(k["fireBlastSurplus"]))))
	}
	if len(blastConds) > 0 {
		items = append(items, cooldownWaitLines(k, s.fireBlast, or(blastConds...))...)
	}
	items = append(items, cast(s.frostfireBolt, nil, "Frostfire Bolt."))
	items = append(items, wandLines(k)...)
	rot := rotation(items...)
	if k["prepull"] == 1 {
		rot.PrepullActions = append(rot.PrepullActions, prepull(s.frostfireBolt, 3))
	}
	return rot
}
