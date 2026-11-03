package main

import (
	"fmt"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/druid"
)

// Feral cat under Forever rules. Forever's Furor gives back only part of the Energy a cat
// left its form with, so powershifting is gone. Shifting Power (the Feral row 7 talent since the
// launch realms' 2026-10-30 hotfix) takes its place: 40 Energy for 55% of base mana on an 8 sec
// cooldown, cast from Cat Form. Predatory Discipline, which it needs, makes Ferocious Bite keep its
// extra Energy and still deal the damage for it (sim/druid/ferocious_bite.go), so a Bite at high
// Energy costs only its 35. The same hotfix took Berserk out of the tree. Tiger's Fury is gone from
// Forever (sim/druid/tigers_fury.go); its lines are for the Classic ruleset.
type druidFeral struct{}

func init() {
	templates["druid_feral"] = druidFeral{}
}

var (
	druidRakeLevel   = []int32{0, 24, 34, 44, 54}
	druidRakeSpellId = []int32{0, 1822, 1823, 1824, 9904}
	druidRipLevel    = []int32{0, 20, 28, 36, 44, 52, 60}
	druidRipSpellId  = []int32{0, 1079, 9492, 9493, 9752, 9894, 9896}
	// Ferocious Bite: Forever's new rank 1 at 24 (1328618, the 2026-10-30 hotfix), then 1.12's. Its
	// rank at 60 (31018) is an Ahn'Qiraj book rank, which Forever leaves out.
	druidBiteLevel     = []int32{0, 24, 32, 40, 48, 56}
	druidBiteSpellId   = []int32{0, 1328618, 22568, 22827, 22828, 22829}
	druidTigersFury    = spellID(5217)
	druidShiftingPower = spellID(1322605)
	druidShadowmeld    = spellID(20580)
	druidRavage        = spellID(9867)
	// The filler explosive (the dps report's engineering bomb slot).
	druidBomb = itemID(260817)
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
		// Bite only with this much Energy, its extra Energy turns into damage (and with
		// Predatory Discipline stays after the Bite).
		{Name: "biteEnergy", Default: 35, Min: 35, Max: 95, Step: 10},
		// With 1, Ferocious Bite at 5 combo points anyway once the Energy reaches
		// biteCapEnergy, so a cat waiting on Rip does not sit at full Energy. 0 never.
		{Name: "biteAtCap", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "biteCapEnergy", Default: 90, Min: 60, Max: 100, Step: 10},
		// Keep Rake up, 0/1, but not with rakeMinTime seconds of the fight or less left.
		{Name: "rake", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "rakeMinTime", Default: 0, Min: 0, Max: 9, Step: 1},
		// The Night Elf's Shadowmeld into Ravage ahead of the builder, from behind, 0/1: on
		// Clearcasting, or in the last shadowmeldEnd seconds with the Energy for it. It does nothing
		// for other races.
		{Name: "shadowmeld", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "shadowmeldEnd", Default: 10, Min: 0, Max: 20, Step: 5},
		// With 1 the Shadowmeld line goes ahead of Rake, with 0 between Rake and the builder.
		{Name: "meldBeforeRake", Default: 1, Min: 0, Max: 1, Step: 1},
		// The builder: 0 Shred (from behind), 1 Claw.
		{Name: "builder", Default: 0, Min: 0, Max: 1, Step: 1},
		// Shifting Power at or below this much Energy, 0 never. Above 60 its 40 Energy can go past
		// the cap.
		{Name: "shiftEnergy", Default: 50, Min: 0, Max: 80, Step: 10},
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
		// Cooldown stacking (rotation auditor), 0/1: the Night Elf's Elune's Light (10% crit) held
		// for the synthetic Strength trinket when holding loses no use. Until the 2026-10-30 hotfix
		// took Berserk out of the tree, it held Berserk for the trinket and Elune's Light out of
		// Berserk.
		{Name: "cdStack", Default: 0, Min: 0, Max: 1, Step: 1},
		// Elune's Light (the Night Elf's racial) in the cdStack line, 0/1. Other races leave it out;
		// a Night Elf without the line still uses it through the autocast line.
		{Name: "elune", Default: 1, Min: 0, Max: 1, Step: 1},
		// Tiger's Fury before the pull, so it is ready again sooner, 0/1.
		{Name: "prepullTigersFury", Default: 0, Min: 0, Max: 1, Step: 1},
		// Faerie Fire duty (standards 23), 0/1: keep Faerie Fire on the target, cast again
		// once it has less than ffRefresh seconds left. ffPos 0 puts it ahead of everything but the
		// cooldowns, 1 after Rip.
		{Name: "faerieFire", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "ffRefresh", Default: 1, Min: 0, Max: 8, Step: 1},
		{Name: "ffPos", Default: 0, Min: 0, Max: 1, Step: 1},
		// The filler explosive, 0/1: with 0 the autocast line throws it on cooldown; with 1 its own
		// line throws it only at or below bombEnergy Energy (not at the cap, where the 1 sec cast
		// wastes Energy) and right after a main hand swing (the cast restarts the swing timer).
		{Name: "bomb", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "bombEnergy", Default: 100, Min: 20, Max: 100, Step: 20},
		// Rake also at 5 combo points, 0/1: the bleed is worth more than the wasted point.
		{Name: "rakeAt5", Default: 0, Min: 0, Max: 1, Step: 1},
		// The Shadowmeld Ravage line at 5 combo points too on Clearcasting, 0/1: a free Ravage is
		// worth more than the wasted point.
		{Name: "meldAt5", Default: 0, Min: 0, Max: 1, Step: 1},
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
	if k["bomb"] == 1 {
		items = append(items, cast(druidBomb, and(le(energy(), num(k["bombEnergy"])), ge(mainHandTimeToNext(), sub(mainHandSwingTime(), seconds(0.4)))),
			"The filler explosive right after a main hand swing (its 1 sec cast restarts the swing timer), and not near the Energy cap, where the cast would waste Energy."))
	}
	items = append(items, autocastOtherCooldowns(nil, "Racials, trinkets, potions and runes."))
	switch k["tigersFury"] {
	case 1:
		items = append(items, cast(druidTigersFury, nil, "Tiger's Fury on cooldown, it is free and off the global cooldown."))
	case 2:
		items = append(items, cast(druidTigersFury, le(energy(), num(k["tigersFuryEnergy"])), "Tiger's Fury once the Energy is low."))
	}
	if k["cdStack"] == 1 && k["elune"] == 1 {
		items = append(items, cast(elunesLight, holdFor(elunesLight, itemID(druid.FeralSyntheticStrengthTrinket)),
			"Night Elf: Elune's Light with the Strength trinket, unless waiting for it would lose a use. Other races can ignore the warning on this line."))
	}
	// Forever's Faerie Fire and Curse of Recklessness share the minor armor slot (standards 20), and
	// the sim won't put ours up while the raid's Faerie Fire or a Curse of Recklessness holds it, so
	// the line waits for the slot to be free instead of recasting on every cooldown.
	// Forever has no Faerie Fire (Feral): the cat casts the caster Faerie Fire (druid_balance.go's
	// druidFaerieFire), free in Cat Form, on a 1.5 sec global cooldown and no cooldown. It is the same
	// debuff as the raid's, so the line only refreshes it.
	ffItem := cast(druidFaerieFire, and(not(targetAuraIsActive(druidCoRAura)),
		lt(targetAuraRemainingTime(druidFaerieFire), seconds(k["ffRefresh"]))),
		"Faerie Fire on the target, free in Cat Form, refreshed before it drops, unless Curse of Recklessness holds the slot (Faerie Fire duty).")
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
		note := "Ferocious Bite at 5 combo points while Rip has time left" + biteEnergyNote(k) + "."
		if k["biteAtCap"] == 1 {
			note = "Ferocious Bite at 5 combo points while Rip has time left" + biteEnergyNote(k) + ", or at full Energy."
		}
		items = append(items, cast(bite, cond, note))
	case 2:
		items = append(items, cast(bite, and(eq(comboPoints(), num(5)), ge(energy(), num(k["biteEnergy"]))),
			"Ferocious Bite at 5 combo points"+biteEnergyNote(k)+"."))
	}
	var meld *proto.APLListItem
	if k["shadowmeld"] == 1 {
		trigger := auraIsActive(druidClearcasting)
		if k["shadowmeldEnd"] > 0 {
			trigger = or(trigger, and(ge(energy(), spellCurrentCost(druidRavage)), lt(remainingTime(), seconds(k["shadowmeldEnd"]))))
		}
		meldCP := lt(comboPoints(), num(5))
		if k["meldAt5"] == 1 {
			meldCP = or(meldCP, auraIsActive(druidClearcasting))
		}
		meld = &proto.APLListItem{
			Notes: "Night Elf: Shadowmeld into Ravage on Clearcasting, or in the last seconds with the Energy for it. Other races can ignore the warning on this line.",
			Action: &proto.APLAction{
				Condition: and(spellIsReady(druidShadowmeld), not(&proto.APLValue{Value: &proto.APLValue_FrontOfTarget{FrontOfTarget: &proto.APLValueFrontOfTarget{}}}),
					meldCP, trigger),
				Action: &proto.APLAction_StrictSequence{StrictSequence: &proto.APLActionStrictSequence{Actions: []*proto.APLAction{
					{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: druidShadowmeld}}},
					{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: druidRavage}}},
				}}},
			},
		}
	}
	if meld != nil && k["meldBeforeRake"] == 1 {
		items = append(items, meld)
	}
	if k["rake"] == 1 {
		cond := not(dotIsActive(rake))
		if k["rakeAt5"] == 0 {
			cond = and(cond, lt(comboPoints(), num(5)))
		}
		note := "Rake whenever it is down."
		if k["rakeMinTime"] > 0 {
			cond = and(cond, gt(remainingTime(), seconds(k["rakeMinTime"])))
			note = "Rake whenever it is down, but not at the very end of the fight."
		}
		items = append(items, cast(rake, cond, note))
	}
	if meld != nil && k["meldBeforeRake"] == 0 {
		items = append(items, meld)
	}
	items = append(items, cast(builder, lt(comboPoints(), num(5)), "The builder."))

	rot := rotation(items...)
	if k["prepullTigersFury"] == 1 {
		rot.PrepullActions = append(rot.PrepullActions, prepull(druidTigersFury, 1))
	}
	return rot
}

// biteEnergyNote says why the Bite waits for Energy: the extra Energy adds damage, and with
// Predatory Discipline the Bite keeps it for the next builders.
func biteEnergyNote(k Knobs) string {
	if k["biteEnergy"] <= 35 {
		return ""
	}
	return fmt.Sprintf(", once the Energy reaches %.0f (the extra Energy adds damage, and Predatory Discipline keeps it)", k["biteEnergy"])
}
