package main

import "github.com/wowsims/classic/sim/core/proto"

// Fire under Forever rules. Improved Scorch is the mage's own under Forever (3% more Fire damage
// per stack, from that mage only), so the fire mage keeps 5 stacks up with its own Scorches.
// Heating Up takes 25% off the next Pyroblast's cast per Fireball, Fire Blast or Scorch crit,
// to 75% at 3 stacks, so the question is how many stacks a Pyroblast waits for. Fire Blast is
// instant, and with Wake of Fire it is up every 6 sec.
type mageFire struct{}

func init() {
	templates["mage_fire"] = mageFire{}
}

func (mageFire) Knobs() []Knob {
	return append(mageCommonKnobs(),
		// Scorch until Fire Vulnerability has 5 stacks, and again once it has less than this
		// many seconds left. 0 never Scorches.
		Knob{Name: "scorchRefresh", Default: 5, Min: 0, Max: 9, Step: 3},
		// Pyroblast at this many Heating Up stacks (3 is a 1.5 sec cast), 0 never, 4 hard cast
		// on cooldown of nothing (whenever, as a filler).
		Knob{Name: "pyroStacks", Default: 3, Min: 0, Max: 4, Step: 1},
		// Fire Blast on cooldown above this much mana, 0 never.
		Knob{Name: "fireBlastMana", Default: 20, Min: 0, Max: 80, Step: 20},
		// Fire Blast only while Heating Up is below 3 stacks, so its crits are not wasted.
		Knob{Name: "fireBlastHeating", Default: 0, Min: 0, Max: 1, Step: 1},
		// Pyroblast only above this much mana, 0 any time. A Pyroblast costs more than the filler,
		// so a long fight saves it for the Heating Up stacks the mana allows.
		Knob{Name: "pyroMana", Default: 0, Min: 0, Max: 60, Step: 10},
		// Arcane Missiles on a Missile Barrage proc (Fireball has 20%), for builds with the
		// Arcane talent.
		Knob{Name: "missileBarrage", Default: 1, Min: 0, Max: 1, Step: 1},
		// The filler: 1 Fireball, 2 Frostfire Bolt, 3 Scorch.
		Knob{Name: "filler", Default: 1, Min: 1, Max: 3, Step: 1},
		// Precast the filler so the first one lands at the pull.
		Knob{Name: "prepull", Default: 1, Min: 0, Max: 1, Step: 1},
		scorchFirstKnob(),
	)
}

func (mageFire) Build(k Knobs) *proto.APLRotation {
	s := mageRanks()
	items := mageCooldowns(k, s)
	if k["scorchRefresh"] > 0 {
		items = append(items, cast(s.scorch,
			and(targetAuraIsKnown(fireVulnerability), or(lt(targetAuraNumStacks(fireVulnerability), num(5)), lt(targetAuraRemainingTime(fireVulnerability), seconds(k["scorchRefresh"])))),
			"Scorch up to 5 stacks of Fire Vulnerability, and before it runs out. Under Forever it only counts for the mage who put it up."))
	}
	switch p := k["pyroStacks"]; {
	case p >= 1 && p <= 3:
		items = append(items, cast(s.pyroblast, and(known(heatingUp, ge(auraNumStacks(heatingUp), num(p))), manaAbove(k["pyroMana"])), "Pyroblast once Heating Up has cut its cast enough."))
	}
	if k["missileBarrage"] == 1 {
		items = append(items, cast(s.arcaneMissiles, known(missileBarrage, auraIsActive(missileBarrage)), "Arcane Missiles on a Missile Barrage proc, free and twice as fast."))
	}
	if k["fireBlastMana"] > 0 {
		cond := ge(currentManaPercent(), num(k["fireBlastMana"]/100))
		if k["fireBlastHeating"] == 1 {
			cond = and(cond, or(not(auraIsKnown(heatingUp)), lt(auraNumStacks(heatingUp), num(3))))
		}
		items = append(items, cast(s.fireBlast, cond, "Fire Blast on cooldown with the mana to spare."))
	}
	if k["pyroStacks"] == 4 {
		items = append(items, cast(s.pyroblast, nil, "Pyroblast as the filler."))
	}
	filler := s.fireball
	switch k["filler"] {
	case 2:
		filler = s.frostfireBolt
	case 3:
		filler = s.scorch
	}
	items = append(items, cast(filler, nil, "The filler."))
	rot := rotation(items...)
	if k["prepull"] == 1 {
		rot.PrepullActions = append(rot.PrepullActions, prepull(filler, 3))
	}
	return rot
}
