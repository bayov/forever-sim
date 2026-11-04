package main

import (
	"github.com/wowsims/classic/sim/core/proto"
)

// Backstab under Forever rules. The cooldown rules are the shared ones in rogue_common.go.
// This template carries a line for every dagger talent a build might take (Ghostly
// Strike and the Ambush or Garrote it opens, Hemorrhage, Rupture with Serrated Blades,
// Death Mark, Vanish into Premeditation and Ambush), so one template serves the talent search as well
// as the rotation search. A line for a talent the build does not have is dropped by the
// sim with a warning and costs nothing.
type rogueBS struct{}

var (
	backstab      = spellID(25300)
	rupture       = spellID(11275)
	ghostlyStrike = spellID(14278)
	hemorrhage    = spellID(16511)
	ambush        = spellID(11269)
	vanish        = spellID(1856)
	premeditation = spellID(14183)
	preparation   = spellID(14185)
	stealth       = spellID(1787)
	garrote       = spellIDRank(11290, 6)
	deathMark     = spellID(1325758)
	mutilate      = spellID(1241584)
	ghostlyOpener = spellIDTag(14278, 1)
	shadowmeld    = spellID(20580)
)

func (rogueBS) Knobs() []Knob {
	return append(rogueCooldownKnobs(),
		// Rupture at this many combo points when it is not ticking, 0 for never.
		Knob{Name: "ruptureCp", Default: 0, Min: 0, Max: 5, Step: 1},
		// Rupture only with this much Slice and Dice left, so the points are not needed there.
		Knob{Name: "ruptureSnd", Default: 6, Min: 0, Max: 12, Step: 3},
		// Hemorrhage before a Rupture when the debuff is missing.
		Knob{Name: "hemoForRupture", Default: 1, Min: 0, Max: 1, Step: 1},
		// Hemorrhage as the builder instead of Backstab.
		Knob{Name: "hemoFiller", Default: 0, Min: 0, Max: 1, Step: 1},
		// Mutilate as the builder instead of Backstab (a 20 point Assassination talent under
		// Forever, daggers in both hands).
		Knob{Name: "mutilate", Default: 0, Min: 0, Max: 1, Step: 1},
		// Ambush in the 7 sec Ghostly Strike leaves for one without Stealth (November 2 development
		// notes), or a Garrote there when it is not ticking (ghostlyGarrote).
		Knob{Name: "ghostlyAmbush", Default: 1, Min: 0, Max: 1, Step: 1},
		Knob{Name: "ghostlyGarrote", Default: 0, Min: 0, Max: 1, Step: 1},
		// Death Mark (the 31 point Subtlety talent) at this many combo points or more, 0 for never.
		Knob{Name: "deathMarkCp", Default: 5, Min: 0, Max: 5, Step: 1},
		// Premeditation (no longer needs Stealth under Forever) at this many combo points or
		// fewer, so the 2 points it adds are not lost over 5. 9 leaves it to the cooldown rule
		// in the sim, which uses it at 3 or fewer.
		Knob{Name: "premedCp", Default: 9, Min: 0, Max: 9, Step: 1},
		// Ghostly Strike on cooldown as a builder.
		Knob{Name: "ghostly", Default: 1, Min: 0, Max: 1, Step: 1},
		// Vanish for an Ambush (and Premeditation) at this many combo points or fewer, 9 for never.
		Knob{Name: "vanishCp", Default: 9, Min: 0, Max: 9, Step: 3},
		// Pool energy for the Ambush while Vanish is ready, so the builders do not keep it
		// below 60 (an energy starved build otherwise leaves Vanish idle).
		Knob{Name: "vanishPool", Default: 0, Min: 0, Max: 1, Step: 1},
		// Night Elf: Shadowmeld into Ambush at this many combo points or fewer, 9 for never.
		// The line does nothing for the other races.
		Knob{Name: "meldCp", Default: 3, Min: 0, Max: 9, Step: 3},
		Knob{Name: "bsAfterSwing", Default: 0, Min: 0, Max: 1, Step: 0.25},
		Knob{Name: "bsEnergy", Default: 0, Min: 0, Max: 100, Step: 10},
	)
}

func (rogueBS) Build(k Knobs) *proto.APLRotation {
	c := buildRogueCooldowns(k)
	var items []*proto.APLListItem
	// Out of a Vanish the Ambush goes first, before anything breaks the Stealth.
	if k["vanishCp"] < 9 {
		items = append(items, cast(ambush, auraIsActive(stealth), "Ambush out of a Vanish before anything else."))
	}
	items = append(items, c.sliceAndDice)
	items = append(items, c.cooldowns...)

	// Rupture is the other finisher. Under Forever its ticks read Hemorrhage as they land
	// (since 2026-10-23, sim/rogue/rupture.go), so the debuff counts only while it is up, not
	// at the cast. The finisher goes out when Slice and Dice will not need the points and the
	// fight lasts long enough for the ticks.
	var ruptureWhen value
	if k["ruptureCp"] > 0 {
		ruptureWhen = and(
			ge(comboPoints(), num(k["ruptureCp"])),
			not(dotIsActive(rupture)),
			ge(auraRemainingTime(sliceAndDice), seconds(k["ruptureSnd"])),
			gt(remainingTime(), seconds(12)),
			c.ruptureHold,
		)
		if k["hemoForRupture"] == 1 {
			items = append(items, cast(hemorrhage, and(
				ge(comboPoints(), num(k["ruptureCp"]-1)),
				lt(comboPoints(), num(5)),
				not(dotIsActive(rupture)),
				not(targetAuraIsActive(hemorrhage)),
				ge(auraRemainingTime(sliceAndDice), seconds(k["ruptureSnd"])),
				gt(remainingTime(), seconds(12)),
				c.ruptureHold,
			), "Hemorrhage right before a Rupture when the debuff is missing, so the first ticks land with it (each tick reads it as it lands)."+c.holdNotes))
		}
		items = append(items, cast(rupture, ruptureWhen, "Rupture when it is not ticking and Slice and Dice does not need the points."+c.holdNotes))
	}

	if k["premedCp"] < 9 {
		items = append(items, cast(premeditation, le(comboPoints(), num(k["premedCp"])),
			"Premeditation needs no Stealth under Forever, so it goes out whenever the 2 points fit."))
	}

	if k["deathMarkCp"] > 0 {
		items = append(items, cast(deathMark, and(ge(comboPoints(), num(k["deathMarkCp"])), gt(remainingTime(), seconds(6))),
			"Death Mark on cooldown at enough combo points while its 6 sec store can still go off in the fight. Without the talent ignore the warning on this line."))
	}
	items = append(items, cast(eviscerate, c.evisWhen, ""))

	// Vanish is only an Ambush at a low combo count, so the points are not wasted over 5.
	// Ambush breaks the stealth right after. Preparation gives a second Vanish once the
	// first is spent. Cutthroat Ambushes ride on the same line.
	// Shadowmeld is a Night Elf's second Vanish (standards 13: only in the fight), off the
	// GCD, and needs the target's back. The Ambush goes out in the same sequence.
	meld := &proto.APLListItem{
		Notes: "Night Elf: Shadowmeld into Ambush, as the Vanish line. Other races can ignore the warning on this line.",
		Action: &proto.APLAction{
			Condition: and(
				spellIsReady(shadowmeld),
				not(&proto.APLValue{Value: &proto.APLValue_FrontOfTarget{FrontOfTarget: &proto.APLValueFrontOfTarget{}}}),
				not(auraIsActive(stealth)),
				le(comboPoints(), num(k["meldCp"])),
				ge(energy(), spellCurrentCost(ambush)),
				auraIsActive(sliceAndDice),
			),
			Action: &proto.APLAction_StrictSequence{StrictSequence: &proto.APLActionStrictSequence{Actions: []*proto.APLAction{
				{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: shadowmeld}}},
				{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: ambush}}},
			}}},
		},
	}
	if k["vanishCp"] < 9 {
		items = append(items, cast(vanish, and(
			not(auraIsActive(stealth)),
			le(comboPoints(), num(k["vanishCp"])),
			ge(energy(), spellCurrentCost(ambush)),
			auraIsActive(sliceAndDice),
			gt(remainingTime(), seconds(10)),
		), "Vanish for an Ambush at low combo points."))
	}
	if k["meldCp"] < 9 {
		items = append(items, meld)
	}
	if k["vanishCp"] < 9 {
		items = append(items,
			cast(ambush, auraIsActive(stealth), "Ambush from stealth after Vanish."),
			cast(preparation, and(
				not(auraIsActive(stealth)),
				le(comboPoints(), num(k["vanishCp"])),
				gt(spellTimeToReady(vanish), seconds(60)),
				gt(remainingTime(), seconds(10)),
			), "Preparation resets Vanish (and Cold Blood) for a second Ambush once the first is spent."),
		)
	}

	// While Vanish is ready for an Ambush, the builders leave its 60 energy alone.
	var pool value
	if k["vanishPool"] == 1 && k["vanishCp"] < 9 {
		pool = or(not(spellIsReady(vanish)), le(remainingTime(), seconds(10)), ge(comboPoints(), num(4)), not(auraIsActive(sliceAndDice)))
	}

	if k["ghostly"] == 1 {
		items = append(items, cast(ghostlyStrike, and(lt(comboPoints(), num(5)), pool),
			"Ghostly Strike on cooldown: a builder, and it opens a 7 sec window for an Ambush without Stealth."))
	}
	if k["ghostlyGarrote"] == 1 {
		items = append(items, cast(garrote, and(auraIsActive(ghostlyOpener), not(dotIsActive(garrote)), gt(remainingTime(), seconds(9))),
			"Garrote in the window Ghostly Strike opens, when it is not ticking and the fight lasts for its ticks. Without Ghostly Strike ignore the warning on this line."))
	}
	if k["ghostlyAmbush"] == 1 {
		items = append(items, cast(ambush, and(auraIsActive(ghostlyOpener), lt(comboPoints(), num(5))),
			"Ambush in the window Ghostly Strike opens, the hardest hitting builder. Without Ghostly Strike ignore the warning on this line."))
	}
	if s := c.shank; s != nil {
		items = append(items, s)
	}

	// The builder. Backstab waits for Thousand Cuts stacks when asked, with the energy
	// pool as the fallback so the wait never caps energy.
	when := and(builderWhen(k["bsAfterSwing"], k["bsEnergy"]), pool)
	poolNotes := ""
	if pool != nil {
		poolNotes = " While Vanish is ready for an Ambush it leaves the Ambush's energy alone."
	}
	if k["hemoFiller"] == 1 {
		items = append(items, cast(hemorrhage, when, builderNotes("Hemorrhage as the builder,", k["bsAfterSwing"], k["bsEnergy"])+poolNotes))
	}
	if k["mutilate"] == 1 {
		items = append(items, cast(mutilate, when,
			builderNotes("Mutilate as the builder (2 combo points for the energy of a Backstab),", k["bsAfterSwing"], k["bsEnergy"])+poolNotes+" Backstab takes over when the off hand is not a dagger."))
	}
	items = append(items, cast(backstab, when, builderNotes("Backstab", k["bsAfterSwing"], k["bsEnergy"])+poolNotes))

	rot := rotation(items...)
	rot.PrepullActions = c.prepull
	return rot
}
