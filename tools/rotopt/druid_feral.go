package main

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/druid"
)

// Feral cat under Forever rules. Forever's Furor gives back only part of the Energy a cat
// left its form with, so powershifting is gone. Shifting Power (the Feral talent) takes its
// place: 40 Energy for 55% of base mana on a 16 sec cooldown (8 with Improved Shifting
// Power), cast from Cat Form. Tiger's Fury is free, off the global cooldown and 15% more
// Physical damage for 6 sec on a 30 sec cooldown, and Berserk (the 31 point talent) makes
// every builder crit for 15 sec, which Blood Frenzy turns into extra combo points.
type druidFeral struct{}

func init() {
	templates["druid_feral"] = druidFeral{}
}

var (
	druidRakeLevel   = []int32{0, 24, 34, 44, 54}
	druidRakeSpellId = []int32{0, 1822, 1823, 1824, 9904}
	druidRipLevel    = []int32{0, 20, 28, 36, 44, 52, 60}
	druidRipSpellId  = []int32{0, 1079, 9492, 9493, 9752, 9894, 9896}
	// Ferocious Bite rank 5 (60, 31018) is an Ahn'Qiraj book rank, which Forever leaves out.
	druidBiteLevel     = []int32{0, 32, 40, 48, 56}
	druidBiteSpellId   = []int32{0, 22568, 22827, 22828, 22829}
	druidTigersFury    = spellID(5217)
	druidShiftingPower = spellID(1322605)
	druidBerserk       = spellID(417141)
	// Faerie Fire (Feral): a trainer spell at 25 under Forever, cast from Cat Form, free, 1 sec
	// global cooldown, 6 sec cooldown.
	druidFaerieFireFeral = spellID(17392)
)

func (druidFeral) Knobs() []Knob {
	return []Knob{
		// Rip at this many combo points while it is down, 0 never.
		{Name: "ripCP", Default: 4, Min: 0, Max: 5, Step: 1},
		// No new Rip with less than this many seconds of the fight left.
		{Name: "ripMinTime", Default: 10, Min: 4, Max: 12, Step: 2},
		// Ferocious Bite at 5 combo points: 0 never, 1 only while Rip has at least
		// biteRipTime seconds left (or with no Rip in the rotation), 2 always.
		{Name: "bite", Default: 1, Min: 0, Max: 2, Step: 1},
		{Name: "biteRipTime", Default: 4, Min: 0, Max: 10, Step: 2},
		// Bite only with this much Energy, its extra Energy turns into damage.
		{Name: "biteEnergy", Default: 35, Min: 35, Max: 75, Step: 10},
		// With 1, Ferocious Bite at 5 combo points anyway once the Energy reaches
		// biteCapEnergy, so a cat waiting on Rip does not sit at full Energy. 0 never.
		{Name: "biteAtCap", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "biteCapEnergy", Default: 90, Min: 60, Max: 100, Step: 10},
		// Keep Rake up, 0/1.
		{Name: "rake", Default: 0, Min: 0, Max: 1, Step: 1},
		// The builder: 0 Shred (from behind), 1 Claw.
		{Name: "builder", Default: 0, Min: 0, Max: 1, Step: 1},
		// Shifting Power at or below this much Energy, 0 never.
		{Name: "shiftEnergy", Default: 50, Min: 0, Max: 60, Step: 10},
		// Shifting Power only above this much mana, so the pool lasts the fight.
		{Name: "shiftMana", Default: 0, Min: 0, Max: 40, Step: 10},
		// Hold Shifting Power while a Ferocious Bite is due (5 combo points with Rip up for
		// biteRipTime), so the Energy it gives goes into Shreds and not into the Bite, 0/1.
		{Name: "shiftHoldForBite", Default: 1, Min: 0, Max: 1, Step: 1},
		// Ferocious Bite at any combo points in the last endBite seconds of the fight, so they
		// are not wasted, 0 never.
		{Name: "endBite", Default: 0, Min: 0, Max: 4, Step: 1},
		// Tiger's Fury: 1 on cooldown, 2 only at or below tigersFuryEnergy too, 0 never.
		{Name: "tigersFury", Default: 1, Min: 0, Max: 2, Step: 1},
		{Name: "tigersFuryEnergy", Default: 40, Min: 20, Max: 80, Step: 20},
		// Berserk: 1 on cooldown, 2 with Tiger's Fury up, 0 never.
		{Name: "berserk", Default: 1, Min: 0, Max: 2, Step: 1},
		// Cooldown stacking (rotation auditor), 0/1: Berserk's later uses held for the synthetic
		// Strength trinket when holding loses no use, and the Night Elf's Elune's Light (10% crit)
		// kept out of Berserk, where every builder crits anyway.
		{Name: "cdStack", Default: 0, Min: 0, Max: 1, Step: 1},
		// Elune's Light (the Night Elf's racial) in the cdStack line, 0/1. Other races leave it out;
		// a Night Elf without the line still uses it through the autocast line.
		{Name: "elune", Default: 1, Min: 0, Max: 1, Step: 1},
		// Tiger's Fury before the pull, so it is ready again sooner, 0/1.
		{Name: "prepullTigersFury", Default: 0, Min: 0, Max: 1, Step: 1},
		// Faerie Fire duty (standards 23), 0/1: keep Faerie Fire (Feral) on the target, cast again
		// once it has less than ffRefresh seconds left. ffPos 0 puts it ahead of everything but the
		// cooldowns, 1 after Rip.
		{Name: "faerieFire", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "ffRefresh", Default: 1, Min: 0, Max: 8, Step: 1},
		{Name: "ffPos", Default: 0, Min: 0, Max: 1, Step: 1},
	}
}

func (druidFeral) Build(k Knobs) *proto.APLRotation {
	shred := rankID(druid.ShredLevel, druid.ShredSpellId)
	claw := rankID(druid.ClawLevel, druid.ClawSpellId)
	rake := rankID(druidRakeLevel, druidRakeSpellId)
	rip := rankID(druidRipLevel, druidRipSpellId)
	bite := rankID(druidBiteLevel, druidBiteSpellId)

	builder := shred
	if k["builder"] == 1 {
		builder = claw
	}

	var items []*proto.APLListItem
	items = append(items, autocastOtherCooldowns(nil, "Racials, trinkets, potions and runes."))
	switch k["tigersFury"] {
	case 1:
		items = append(items, cast(druidTigersFury, nil, "Tiger's Fury on cooldown, it is free and off the global cooldown."))
	case 2:
		items = append(items, cast(druidTigersFury, le(energy(), num(k["tigersFuryEnergy"])), "Tiger's Fury once the Energy is low."))
	}
	switch {
	case k["berserk"] == 1 && k["cdStack"] == 1:
		items = append(items, cast(druidBerserk, holdFor(druidBerserk, itemID(druid.FeralSyntheticStrengthTrinket)),
			"Berserk with the trinket, unless waiting for it would lose a use. Without the talent ignore the warning on this line."))
	case k["berserk"] == 1:
		items = append(items, cast(druidBerserk, nil, "Berserk on cooldown. Without the talent ignore the warning on this line."))
	case k["berserk"] == 2:
		items = append(items, cast(druidBerserk, auraIsActive(druidTigersFury), "Berserk with Tiger's Fury."))
	}
	if k["cdStack"] == 1 && k["elune"] == 1 {
		items = append(items, cast(elunesLight, or(
			and(not(auraIsActive(druidBerserk)), not(spellIsReady(druidBerserk))),
			gt(spellUsesLostByDelay(elunesLight, auraRemainingTime(druidBerserk), seconds(0)), num(0))),
			"Elune's Light outside Berserk, unless waiting would lose a use."))
	}
	// Forever's Faerie Fire and Curse of Recklessness share the minor armor slot (standards 20), and
	// the sim won't put ours up while the raid's Faerie Fire or a Curse of Recklessness holds it, so
	// the line waits for the slot to be free instead of recasting on every cooldown.
	ffItem := cast(druidFaerieFireFeral, and(not(targetAuraIsActive(druidFaerieFire)), not(targetAuraIsActive(druidCoRAura)),
		lt(targetAuraRemainingTime(druidFaerieFireFeral), seconds(k["ffRefresh"]))),
		"Faerie Fire (Feral) on the target, refreshed before it drops, unless another minor armor debuff holds the slot (Faerie Fire duty).")
	if k["faerieFire"] == 1 && k["ffPos"] == 0 {
		items = append(items, ffItem)
	}
	if k["shiftEnergy"] > 0 {
		cond := le(energy(), num(k["shiftEnergy"]))
		if k["shiftMana"] > 0 {
			cond = and(cond, ge(currentManaPercent(), num(k["shiftMana"]/100)))
		}
		if k["shiftHoldForBite"] == 1 {
			cond = and(cond, not(and(eq(comboPoints(), num(5)), ge(dotRemainingTime(rip), seconds(k["biteRipTime"])))))
		}
		note := "Shifting Power once the Energy is low. Without the talent ignore the warning on this line."
		if k["shiftHoldForBite"] == 1 {
			note = "Shifting Power once the Energy is low, but not while a Ferocious Bite is due (its Energy goes into Shreds). Without the talent ignore the warning on this line."
		}
		items = append(items, cast(druidShiftingPower, cond, note))
	}
	if k["endBite"] > 0 {
		items = append(items, cast(bite, and(ge(comboPoints(), num(1)), lt(remainingTime(), seconds(k["endBite"]))),
			"Ferocious Bite at any combo points in the last seconds of the fight."))
	}
	if k["ripCP"] > 0 {
		items = append(items, cast(rip, and(ge(comboPoints(), num(k["ripCP"])), not(dotIsActive(rip)), gt(remainingTime(), seconds(k["ripMinTime"]))),
			"Rip whenever it is down."))
	}
	if k["faerieFire"] == 1 && k["ffPos"] == 1 {
		items = append(items, ffItem)
	}
	switch k["bite"] {
	case 1:
		cond := and(eq(comboPoints(), num(5)), ge(energy(), num(k["biteEnergy"])))
		if k["ripCP"] > 0 {
			cond = and(cond, or(ge(dotRemainingTime(rip), seconds(k["biteRipTime"])), le(remainingTime(), seconds(k["ripMinTime"]))))
		}
		if k["biteAtCap"] == 1 {
			cond = or(cond, and(eq(comboPoints(), num(5)), ge(energy(), num(k["biteCapEnergy"]))))
		}
		note := "Ferocious Bite at 5 combo points while Rip has time left."
		if k["biteAtCap"] == 1 {
			note = "Ferocious Bite at 5 combo points while Rip has time left, or at full Energy."
		}
		items = append(items, cast(bite, cond, note))
	case 2:
		items = append(items, cast(bite, and(eq(comboPoints(), num(5)), ge(energy(), num(k["biteEnergy"]))), "Ferocious Bite at 5 combo points."))
	}
	if k["rake"] == 1 {
		items = append(items, cast(rake, and(not(dotIsActive(rake)), lt(comboPoints(), num(5))), "Rake whenever it is down."))
	}
	items = append(items, cast(builder, lt(comboPoints(), num(5)), "The builder."))

	rot := rotation(items...)
	if k["prepullTigersFury"] == 1 {
		rot.PrepullActions = append(rot.PrepullActions, prepull(druidTigersFury, 1))
	}
	return rot
}
