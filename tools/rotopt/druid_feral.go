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
	druidRakeLevel     = []int32{0, 24, 34, 44, 54}
	druidRakeSpellId   = []int32{0, 1822, 1823, 1824, 9904}
	druidRipLevel      = []int32{0, 20, 28, 36, 44, 52, 60}
	druidRipSpellId    = []int32{0, 1079, 9492, 9493, 9752, 9894, 9896}
	druidBiteLevel     = []int32{0, 32, 40, 48, 56, 60}
	druidBiteSpellId   = []int32{0, 22568, 22827, 22828, 22829, 31018}
	druidTigersFury    = spellID(5217)
	druidShiftingPower = spellID(1322605)
	druidBerserk       = spellID(417141)
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
		// Tiger's Fury: 1 on cooldown, 2 only at or below tigersFuryEnergy too, 0 never.
		{Name: "tigersFury", Default: 1, Min: 0, Max: 2, Step: 1},
		{Name: "tigersFuryEnergy", Default: 40, Min: 20, Max: 80, Step: 20},
		// Berserk: 1 on cooldown, 2 with Tiger's Fury up, 0 never.
		{Name: "berserk", Default: 1, Min: 0, Max: 2, Step: 1},
		// Tiger's Fury before the pull, so it is ready again sooner, 0/1.
		{Name: "prepullTigersFury", Default: 0, Min: 0, Max: 1, Step: 1},
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
	switch k["berserk"] {
	case 1:
		items = append(items, cast(druidBerserk, nil, "Berserk on cooldown. Without the talent ignore the warning on this line."))
	case 2:
		items = append(items, cast(druidBerserk, auraIsActive(druidTigersFury), "Berserk with Tiger's Fury."))
	}
	if k["shiftEnergy"] > 0 {
		cond := le(energy(), num(k["shiftEnergy"]))
		if k["shiftMana"] > 0 {
			cond = and(cond, ge(currentManaPercent(), num(k["shiftMana"]/100)))
		}
		items = append(items, cast(druidShiftingPower, cond, "Shifting Power once the Energy is low. Without the talent ignore the warning on this line."))
	}
	if k["ripCP"] > 0 {
		items = append(items, cast(rip, and(ge(comboPoints(), num(k["ripCP"])), not(dotIsActive(rip)), gt(remainingTime(), seconds(k["ripMinTime"]))),
			"Rip whenever it is down."))
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
		items = append(items, cast(bite, cond, "Ferocious Bite at 5 combo points while Rip has time left, or at full Energy."))
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
