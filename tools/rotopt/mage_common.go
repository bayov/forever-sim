package main

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/mage"
)

// What the three mage templates share: the spells, the mana rules and the cooldown line.

var (
	evocation         = spellID(12051)
	presenceOfMind    = spellID(12043)
	arcanePower       = spellID(12042)
	combustion        = spellID(11129)
	arcaneBlastAura   = spellID(400573)
	missileBarrage    = spellID(400589)
	heatingUp         = spellID(400625)
	fingersOfFrost    = spellID(400670)
	fireVulnerability = spellID(22959)
	wintersChill      = spellID(12579)
	mageClearcasting  = spellID(12536)
	// Mana Agate, Jade, Citrine and Ruby.
	manaGems = []*proto.ActionID{itemID(5514), itemID(5513), itemID(8007), itemID(8008)}
)

type mageSpells struct {
	fireball, frostbolt, frostfireBolt, scorch, fireBlast, pyroblast *proto.ActionID
	arcaneMissiles, arcaneBlast, iceLance                            *proto.ActionID
}

func mageRanks() mageSpells {
	return mageSpells{
		fireball:       rankID(mage.FireballLevel[:], mage.FireballSpellId[:]),
		frostbolt:      rankID(mage.FrostboltLevel[:], mage.FrostboltSpellId[:]),
		frostfireBolt:  rankID(mage.FrostfireBoltLevel[:], mage.FrostfireBoltSpellId[:]),
		scorch:         rankID(mage.ScorchLevel[:], mage.ScorchSpellId[:]),
		fireBlast:      rankID(mage.FireBlastLevel[:], mage.FireBlastSpellId[:]),
		pyroblast:      rankID(mage.PyroblastLevel[:], mage.PyroblastSpellId[:]),
		arcaneMissiles: rankID(mage.ArcaneMissilesLevel[:], mage.ArcaneMissilesSpellId[:]),
		arcaneBlast:    rankID(mage.ArcaneBlastLevel[:], mage.ArcaneBlastSpellId[:]),
		iceLance:       rankID(mage.IceLanceLevel[:], mage.IceLanceSpellId[:]),
	}
}

func targetAuraNumStacks(aura *proto.ActionID) value {
	return &proto.APLValue{Value: &proto.APLValue_AuraNumStacks{AuraNumStacks: &proto.APLValueAuraNumStacks{
		SourceUnit: &proto.UnitReference{Type: proto.UnitReference_CurrentTarget},
		AuraId:     aura,
	}}}
}

func targetAuraRemainingTime(aura *proto.ActionID) value {
	return &proto.APLValue{Value: &proto.APLValue_AuraRemainingTime{AuraRemainingTime: &proto.APLValueAuraRemainingTime{
		SourceUnit: &proto.UnitReference{Type: proto.UnitReference_CurrentTarget},
		AuraId:     aura,
	}}}
}

func auraIsKnown(aura *proto.ActionID) value {
	return &proto.APLValue{Value: &proto.APLValue_AuraIsKnown{AuraIsKnown: &proto.APLValueAuraIsKnown{AuraId: aura}}}
}

func targetAuraIsKnown(aura *proto.ActionID) value {
	return &proto.APLValue{Value: &proto.APLValue_AuraIsKnown{AuraIsKnown: &proto.APLValueAuraIsKnown{
		SourceUnit: &proto.UnitReference{Type: proto.UnitReference_CurrentTarget},
		AuraId:     aura,
	}}}
}

// known guards a condition on a talent's aura. Without the talent the sim drops a condition on
// an aura it does not know, and the line would fire every time, so the guard turns it off instead.
func known(aura *proto.ActionID, v value) value {
	return and(auraIsKnown(aura), v)
}

func currentMana() value {
	return &proto.APLValue{Value: &proto.APLValue_CurrentMana{CurrentMana: &proto.APLValueCurrentMana{}}}
}

func mul(lhs, rhs value) value {
	return &proto.APLValue{Value: &proto.APLValue_Math{Math: &proto.APLValueMath{Op: proto.APLValueMath_OpMul, Lhs: lhs, Rhs: rhs}}}
}

// manaAbove is a mana floor in percent, nil (no condition) at 0.
func manaAbove(percent float64) value {
	if percent <= 0 {
		return nil
	}
	return ge(currentManaPercent(), num(percent/100))
}

// mageCommonKnobs are the mana and cooldown knobs of every mage template.
func mageCommonKnobs() []Knob {
	return []Knob{
		// Evocation below this much mana, 0 never.
		{Name: "evocationMana", Default: 20, Min: 0, Max: 40, Step: 10},
		// Evocation only with at least this many seconds of the fight left, so its 8 sec channel
		// buys mana the fight still needs.
		{Name: "evocationMinTime", Default: 30, Min: 0, Max: 90, Step: 30},
		// Arcane Power and Presence of Mind wait for the other cooldowns: 0 on cooldown, 1
		// Arcane Power waits for the trinket and racial window (they go with it).
		{Name: "cooldownsTogether", Default: 0, Min: 0, Max: 1, Step: 1},
		// Presence of Mind on Pyroblast (1) or on the next spell whatever it is (0).
		{Name: "pomPyroblast", Default: 0, Min: 0, Max: 1, Step: 1},
		// The mana gems (1), or none so the Demonic Rune gets the cooldown they share (0). The
		// client puts gems and runes in one category with a 2 min cooldown, and the sim's
		// autocast takes the smallest gem first.
		{Name: "gems", Default: 1, Min: 0, Max: 1, Step: 1},
	}
}

// scorchFirstKnob holds Combustion until Scorch has 5 stacks of Fire Vulnerability up, so its
// crits go on the spells that take all 15%.
func scorchFirstKnob() Knob {
	return Knob{Name: "cooldownsAfterScorch", Default: 0, Min: 0, Max: 1, Step: 1}
}

// mageCooldowns is the cooldown line, ahead of the spells.
func mageCooldowns(k Knobs, s mageSpells) []*proto.APLListItem {
	var items []*proto.APLListItem
	if k["pomPyroblast"] == 1 {
		items = append(items,
			cast(s.pyroblast, known(presenceOfMind, auraIsActive(presenceOfMind)), "Pyroblast while Presence of Mind is up. Without the talent ignore the warning on this line."),
			cast(presenceOfMind, nil, "Presence of Mind for an instant Pyroblast. Without the talent ignore the warning on this line."),
		)
	}
	if k["cooldownsTogether"] == 1 {
		items = append(items, cast(arcanePower, nil, "Arcane Power on cooldown, ahead of the trinkets and racials. Without the talent ignore the warning on this line."))
	}
	if k["cooldownsAfterScorch"] == 1 {
		fullStacks := and(targetAuraIsKnown(fireVulnerability), ge(targetAuraNumStacks(fireVulnerability), num(5)))
		items = append(items,
			cast(combustion, fullStacks, "Combustion once Fire Vulnerability has 5 stacks. Without the talent ignore the warning on this line."),
		)
	}
	items = append(items, autocastOtherCooldowns(nil, "Arcane Power, Presence of Mind, Combustion, trinkets, racials and the mana gems on a free GCD."))
	if k["evocationMana"] > 0 {
		var timeLeft value
		if k["evocationMinTime"] > 0 {
			timeLeft = gt(remainingTime(), seconds(k["evocationMinTime"]))
		}
		items = append(items, cast(evocation, and(lt(currentManaPercent(), num(k["evocationMana"]/100)), timeLeft), "Evocation once the mana runs low, with enough of the fight left to spend it."))
	}
	if k["gems"] == 0 {
		for _, gem := range manaGems {
			items = append(items, cast(gem, gt(num(0), num(1)), "Never: the mana gems share the Demonic Rune's cooldown, and the rune gives more."))
		}
	}
	return items
}
