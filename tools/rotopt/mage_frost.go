package main

import "github.com/wowsims/classic/sim/core/proto"

// Frost under Forever rules. Frostbolt is the filler. Fingers of Frost (15% on a chill) makes
// the next spells treat the boss as Frozen, which turns on Shatter and Ice Lance's 4 times damage
// (the launch realms' 2026-10-30 hotfix), so the main decision is what to spend the charges on.
// Winter's Chill is a Cone of Cold debuff since the client 1.60.1.70425. Ice Barrier, the 31 point
// talent, raises Frost damage by 5% while the shield holds and gives Frigid Power, 15% more for a
// minute (the 2026-10-30 hotfix; 3% and 3% since the 2026-10-20 one).
type mageFrost struct{}

var (
	iceBarrier  = spellID(13033)
	frigidPower = spellID(1326891)
	coldSnap    = spellID(12472)
)

func init() {
	templates["mage_frost"] = mageFrost{}
}

func (mageFrost) Knobs() []Knob {
	return append(append(append(append(append(mageCommonKnobs(), arcaneKnobs(0, 0)...), finisherKnobs(1)...), iceBarrierKnobs(2)...), blastSpendKnobs()...),
		// Ice Lance while Fingers of Frost is up: 0 never, 1 every charge, 2 only the last
		// charge (the first goes on a Frostbolt), 3 right after a Frostbolt completes (a charge
		// is spent a spell batch after the cast, so the Frostbolt and the Ice Lance both read
		// the target as Frozen), 4 the same on the last charge only.
		Knob{Name: "iceLance", Default: 1, Min: 0, Max: 4, Step: 1},
		// Arcane Missiles on a Missile Barrage proc (Frostbolt has 20%), for builds with the
		// Arcane talent.
		Knob{Name: "missileBarrage", Default: 1, Min: 0, Max: 1, Step: 1},
		// Fire Blast on cooldown above this much mana, 0 never.
		Knob{Name: "fireBlastMana", Default: 0, Min: 0, Max: 80, Step: 20},
		// Precast Frostbolt so the first one lands at the pull.
		Knob{Name: "prepull", Default: 1, Min: 0, Max: 1, Step: 1},
		// Cone of Cold on cooldown (1), never (0). It needs the boss within its 10 yards, so
		// only a mage standing that close (the report's melee-range assumption) can cast it.
		Knob{Name: "coneOfCold", Default: 0, Min: 0, Max: 1, Step: 1},
		// Cold Snap once Cone of Cold has at least this many seconds of cooldown left, 0 never.
		Knob{Name: "coldSnapCone", Default: 0, Min: 0, Max: 9, Step: 1},
	)
}

func (mageFrost) Build(k Knobs) *proto.APLRotation {
	s := mageRanks()
	items := mageCooldowns(k, s)
	switch k["iceLance"] {
	case 1:
		items = append(items, cast(s.iceLance, known(fingersOfFrost, auraIsActive(fingersOfFrost)), "Ice Lance while Fingers of Frost makes the boss count as Frozen (4 times the damage, Shatter's crit)."))
	case 2:
		items = append(items, cast(s.iceLance, known(fingersOfFrost, eq(auraNumStacks(fingersOfFrost), num(1))), "Ice Lance on the last Fingers of Frost charge, the first goes on a Frostbolt."))
	default:
		items = append(items, iceLanceAfterBolt(k, s.frostbolt, s.iceLance)...)
	}
	if k["coneOfCold"] == 1 {
		if k["coldSnapCone"] > 0 {
			items = append(items, cast(coldSnap, ge(spellTimeToReady(s.coneOfCold), seconds(k["coldSnapCone"])), "Cold Snap to finish Cone of Cold's cooldown."))
		}
		items = append(items, cast(s.coneOfCold, nil, "Cone of Cold on cooldown, within its 10 yards: Winter's Chill's stacks are 20% more damage and 7% of spell power each on it."))
	}
	items = append(items, iceBarrierLines(k)...)
	items = append(items, arcaneBlastBuffLines(k, s, nil)...)
	items = append(items, blastClearcastLines(k, s)...)
	items = append(items, arcaneWrathLines(k, s)...)
	if k["missileBarrage"] == 1 {
		items = append(items, cast(s.arcaneMissiles, known(missileBarrage, auraIsActive(missileBarrage)), "Arcane Missiles on a Missile Barrage proc, free and twice as fast."))
	}
	if k["fireBlastMana"] > 0 {
		items = append(items, cast(s.fireBlast, ge(currentManaPercent(), num(k["fireBlastMana"]/100)), "Fire Blast on cooldown with the mana to spare."))
	}
	items = append(items, finisherLines(k, s.frostbolt, s.iceLance)...)
	items = append(items, blastFillerLines(k, s)...)
	items = append(items, cast(s.frostbolt, nil, "Frostbolt."))
	items = append(items, wandLines(k)...)
	rot := rotation(items...)
	iceBarrierPrepull(k, rot)
	if k["prepull"] == 1 {
		rot.PrepullActions = append(rot.PrepullActions, prepull(s.frostbolt, 2.5))
	}
	return rot
}

// iceBarrierKnobs are the Ice Barrier knobs, for the builds with the 31 point Frost talent.
func iceBarrierKnobs(mode float64) []Knob {
	return []Knob{
		// Ice Barrier: 0 never, 1 whenever the shield is down (a raid hit breaks it), 2 only
		// when Frigid Power has under iceBarrierRefresh seconds left, about once a minute.
		{Name: "iceBarrier", Default: mode, Min: 0, Max: 2, Step: 1},
		{Name: "iceBarrierRefresh", Default: 2, Min: 0, Max: 20, Step: 2},
		// Cast Ice Barrier before the pull, ahead of the filler's precast.
		{Name: "iceBarrierPrepull", Default: 1, Min: 0, Max: 1, Step: 1},
	}
}

// iceBarrierLines keep Ice Barrier and Frigid Power up. Not in the last seconds, when the 15% has
// too little left to work on to pay for the GCD.
func iceBarrierLines(k Knobs) []*proto.APLListItem {
	switch k["iceBarrier"] {
	case 1:
		return []*proto.APLListItem{cast(iceBarrier, and(not(auraIsActive(iceBarrier)), gt(remainingTime(), seconds(10))), "Ice Barrier whenever the shield is down: 5% Frost damage while it holds, and Frigid Power's 15% for a minute. Without the talent ignore the warning on this line.")}
	case 2:
		return []*proto.APLListItem{cast(iceBarrier, and(lt(auraRemainingTime(frigidPower), seconds(k["iceBarrierRefresh"])), gt(remainingTime(), seconds(10))), "Ice Barrier when Frigid Power is about to run out: its 15% for another minute, and 5% more while the shield holds. Without the talent ignore the warning on this line.")}
	}
	return nil
}

// iceBarrierPrepull puts Ice Barrier up before the pull, ahead of the filler's precast.
func iceBarrierPrepull(k Knobs, rot *proto.APLRotation) {
	if k["iceBarrier"] > 0 && k["iceBarrierPrepull"] == 1 {
		rot.PrepullActions = append(rot.PrepullActions, prepull(iceBarrier, 4))
	}
}
