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
	// The synthetic level 60 on-use trinket (sim/mage/forever_synthetic_trinkets.go).
	mageSyntheticTalisman = itemID(996032)
	// Mana Agate, Jade, Citrine and Ruby.
	manaGems = []*proto.ActionID{itemID(5514), itemID(5513), itemID(8007), itemID(8008)}
)

type mageSpells struct {
	fireball, frostbolt, frostfireBolt, scorch, fireBlast, pyroblast *proto.ActionID
	arcaneMissiles, arcaneBlast, iceLance, arcaneWrath, coneOfCold   *proto.ActionID
}

// mageRanks leaves out the ranks only the Ahn'Qiraj books teach (Fireball XII, Frostbolt XI,
// Arcane Missiles VIII): the templates build Forever presets, and Forever has no source for the
// books (core's IncludeAQRanks).
func mageRanks() mageSpells {
	return mageSpells{
		fireball:       rankID(mage.FireballLevel[:mage.FireballRanks], mage.FireballSpellId[:]),
		frostbolt:      rankID(mage.FrostboltLevel[:mage.FrostboltRanks], mage.FrostboltSpellId[:]),
		frostfireBolt:  rankID(mage.FrostfireBoltLevel[:], mage.FrostfireBoltSpellId[:]),
		scorch:         rankID(mage.ScorchLevel[:], mage.ScorchSpellId[:]),
		fireBlast:      rankID(mage.FireBlastLevel[:], mage.FireBlastSpellId[:]),
		pyroblast:      rankID(mage.PyroblastLevel[:], mage.PyroblastSpellId[:]),
		arcaneMissiles: rankID(mage.ArcaneMissilesLevel[:mage.ArcaneMissilesRanks], mage.ArcaneMissilesSpellId[:]),
		arcaneBlast:    rankID(mage.ArcaneBlastLevel[:], mage.ArcaneBlastSpellId[:]),
		iceLance:       rankID(mage.IceLanceLevel[:], mage.IceLanceSpellId[:]),
		arcaneWrath:    rankID(mage.ArcaneWrathLevel[:], mage.ArcaneWrathSpellId[:]),
		coneOfCold:     rankID(mage.ConeOfColdLevel[:], mage.ConeOfColdSpellId[:]),
	}
}

// arcaneKnobs are the Arcane talents every mage template can use since the Forever client
// 1.60.1.70425: Arcane Blast's stacks are 15% of Intellect as spell damage each for 15 sec from
// the first (more stacks don't refresh it, other spells don't end it, each stack costs 175% more
// on the next Arcane Blast), and Arcane Wrath is a 3.5 sec cast on a 6 sec cooldown with 5% of
// maximum mana as extra spell damage.
func arcaneKnobs(stacks float64, wrath float64) []Knob {
	return []Knob{
		// Arcane Blast up to this many stacks, 0 never.
		{Name: "blastBuff", Default: stacks, Min: 0, Max: 4, Step: 1},
		// More stacks only while the buff has more than this many seconds left, so the
		// expensive ones aren't bought for a buff about to end. A new buff always starts. 0
		// stacks whenever.
		{Name: "blastRefresh", Default: 9, Min: 0, Max: 12, Step: 3},
		// Arcane Wrath on cooldown (1), never (0).
		{Name: "arcaneWrath", Default: wrath, Min: 0, Max: 1, Step: 1},
	}
}

// arcaneBlastBuffLines keep the Arcane Blast stacks up, for builds with the talent. extra is an
// added condition, nil for none.
func arcaneBlastBuffLines(k Knobs, s mageSpells, extra value) []*proto.APLListItem {
	if k["blastBuff"] == 0 {
		return nil
	}
	stacks := auraNumStacks(arcaneBlastAura)
	var window value
	if k["blastRefresh"] > 0 {
		window = or(not(auraIsActive(arcaneBlastAura)), gt(auraRemainingTime(arcaneBlastAura), seconds(k["blastRefresh"])))
	}
	cond := and(auraIsKnown(arcaneBlastAura), lt(stacks, num(k["blastBuff"])), window,
		// Not in the last seconds, when the spell damage has nothing left to go on.
		gt(remainingTime(), seconds(4)), extra)
	return []*proto.APLListItem{cast(s.arcaneBlast, cond, "Arcane Blast up to the stack target: each stack is 15% of Intellect as spell damage for 15 sec from the first.")}
}

// arcaneWrathLines cast Arcane Wrath on cooldown, for builds with the talent.
func arcaneWrathLines(k Knobs, s mageSpells) []*proto.APLListItem {
	if k["arcaneWrath"] != 1 {
		return nil
	}
	return []*proto.APLListItem{cast(s.arcaneWrath, gt(remainingTime(), add(spellCastTime(s.arcaneWrath), seconds(0.5))), "Arcane Wrath on cooldown. Without the talent ignore the warning on this line.")}
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
		// Shoot the wand when nothing else can be cast (out of mana), 0 never.
		{Name: "wand", Default: 1, Min: 0, Max: 1, Step: 1},
		readLeylineKnob(),
	}
}

// wandLines shoot the wand as the last line, when the mana runs out (Shoot, sim/core/wand.go).
func wandLines(k Knobs) []*proto.APLListItem {
	if k["wand"] != 1 {
		return nil
	}
	return []*proto.APLListItem{cast(spellID(5019), nil, "Shoot the wand when nothing else can be cast (out of mana).")}
}

// scorchFirstKnob holds Combustion until Scorch has 5 stacks of Fire Vulnerability up, so its
// crits go on the spells that take all 15%.
func scorchFirstKnob() Knob {
	return Knob{Name: "cooldownsAfterScorch", Default: 0, Min: 0, Max: 1, Step: 1}
}

// trinketAfterScorchKnob holds the synthetic on-use trinket (+140 spell power) the same way, for
// the Fire spells that take Fire Vulnerability's 15%. Without Improved Scorch it goes on cooldown.
func trinketAfterScorchKnob() Knob {
	return Knob{Name: "trinketAfterScorch", Default: 0, Min: 0, Max: 1, Step: 1}
}

// racialAfterScorchKnob holds the Troll's Berserking and the Orc's Blood Fury the same way (1),
// where the autocast line (0, the default, which adds no line) uses them at the pull, while
// Scorch's 1.5 sec casts take nothing from haste (the GCD holds them).
func racialAfterScorchKnob() Knob {
	return Knob{Name: "racialAfterScorch", Default: 0, Min: 0, Max: 1, Step: 1}
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
	if k["powerWithTrinket"] == 1 {
		// With the trinket up, or when it won't be back before the last 15 sec.
		withTrinket := or(auraIsActive(mageSyntheticTalisman), gt(spellTimeToReady(mageSyntheticTalisman), sub(remainingTime(), seconds(15))))
		items = append(items,
			cast(arcanePower, withTrinket, "Arcane Power with the on-use trinket, unless the trinket won't be back in time. Without the talent ignore the warning on this line."),
			cast(presenceOfMind, withTrinket, "Presence of Mind with Arcane Power and the trinket. Without the talent ignore the warning on this line."),
		)
	}
	if k["trinketAfterScorch"] == 1 {
		fullStacks := or(and(targetAuraIsKnown(fireVulnerability), ge(targetAuraNumStacks(fireVulnerability), num(5))), not(targetAuraIsKnown(fireVulnerability)))
		items = append(items,
			cast(mageSyntheticTalisman, fullStacks, "The on-use trinket once Fire Vulnerability has 5 stacks (or on cooldown without Improved Scorch). Without the trinket ignore the warning on this line."),
		)
	}
	if k["racialAfterScorch"] == 1 {
		fullStacks := or(and(targetAuraIsKnown(fireVulnerability), ge(targetAuraNumStacks(fireVulnerability), num(5))), not(targetAuraIsKnown(fireVulnerability)))
		items = append(items,
			cast(berserking, fullStacks, "Troll. Berserking once Fire Vulnerability has 5 stacks. Other races ignore the warning on this line."),
			cast(bloodFury, fullStacks, "Orc. Blood Fury once Fire Vulnerability has 5 stacks. Other races ignore the warning on this line."),
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
	items = append(items, readLeylineLines(k, nil)...)
	if k["gems"] == 0 {
		for _, gem := range manaGems {
			items = append(items, cast(gem, gt(num(0), num(1)), "Never: the mana gems share the Demonic Rune's cooldown, and the rune gives more."))
		}
	}
	return items
}

// cooldownWaitKnob waits for a cooldown spell (Fire Blast) that is this close to ready instead of
// starting a filler, 0 never. Without haste a Fireball (3 sec), a Fire Blast and a 1.5 sec
// Pyroblast or Scorch make 6 sec, Fire Blast's cooldown, so Fire Blast is ready as each Fireball
// ends; 2% haste (Skyborne Wind Blessed) cuts the Fireball to 2.94 sec, Fire Blast comes up 60 ms
// after it and waits a whole Fireball. The default (0) adds no line.
func cooldownWaitKnob() Knob {
	return Knob{Name: "fireBlastWait", Default: 0, Min: 0, Max: 0.3, Step: 0.1}
}

// cooldownWaitLines wait for spell when it is ready within the fireBlastWait knob and cond (its
// own line's condition, nil for none) holds, ahead of the filler.
func cooldownWaitLines(k Knobs, spell *proto.ActionID, cond value) []*proto.APLListItem {
	if k["fireBlastWait"] <= 0 {
		return nil
	}
	soon := and(gt(spellTimeToReady(spell), seconds(0)), le(spellTimeToReady(spell), seconds(k["fireBlastWait"])), cond)
	return []*proto.APLListItem{action(soon, "Wait for a cooldown that is about to be ready instead of starting the filler.",
		&proto.APLAction{Action: &proto.APLAction_Wait{Wait: &proto.APLActionWait{Duration: spellTimeToReady(spell)}}})}
}

// finisherKnobs: in the last moments of the fight, when the filler's cast would not land before
// the end, cast instants instead (they land, the filler would be lost). finishPad is the margin
// added to the filler's cast time.
func finisherKnobs(pad float64) []Knob {
	return []Knob{
		{Name: "finisher", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "finishPad", Default: pad, Min: 0, Max: 2, Step: 0.25},
	}
}

// finisherLines are the instants to cast when the filler would not land before the end.
func finisherLines(k Knobs, filler *proto.ActionID, instants ...*proto.ActionID) []*proto.APLListItem {
	if k["finisher"] != 1 {
		return nil
	}
	tooLate := lt(remainingTime(), add(spellCastTime(filler), seconds(k["finishPad"])))
	var items []*proto.APLListItem
	for _, instant := range instants {
		items = append(items, cast(instant, tooLate, "An instant when the filler would not land before the end of the fight."))
	}
	return items
}

// iceLanceAfterBolt is the iceLance knob's 3 and 4: Ice Lance in the same spell batch as the
// bolt that just completed. Fingers of Frost spends a charge a batch (10 ms) after each cast
// (sim/mage/talents.go), and the bolt rolls its damage when its cast completes, so both read the
// target as Frozen: on the last charge one charge buys two Frozen spells. The bolt still in
// flight tells the two apart from an Ice Lance a GCD later. At 3 every charge goes that way, at 4
// only the last (with two charges the first goes on the bolt alone).
func iceLanceAfterBolt(k Knobs, bolt, iceLance *proto.ActionID) []*proto.APLListItem {
	var charges value
	switch k["iceLance"] {
	case 3:
		charges = auraIsActive(fingersOfFrost)
	case 4:
		charges = eq(auraNumStacks(fingersOfFrost), num(1))
	default:
		return nil
	}
	return []*proto.APLListItem{cast(iceLance, known(fingersOfFrost, and(charges, spellInFlight(bolt))), "Ice Lance as the bolt completes: the Fingers of Frost charge is spent a spell batch late, so both read the target as Frozen.")}
}
