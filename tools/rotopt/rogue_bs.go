package main

import (
	"github.com/wowsims/classic/sim/core/proto"
)

// Backstab under Forever rules. The cooldown rules are the shared ones in rogue_common.go.
// This template carries a line for every dagger talent a build might take (Ghostly
// Strike, Hemorrhage, Rupture with Serrated Blades and Thousand Cuts, Cutthroat Ambush,
// Vanish into Premeditation and Ambush), so one template serves the talent search as well
// as the rotation search. A line for a talent the build does not have is dropped by the
// sim with a warning and costs nothing.
type rogueBS struct{}

var (
	backstab      = spellID(25300)
	rupture       = spellID(11275)
	ghostlyStrike = spellID(14278)
	hemorrhage    = spellID(17348)
	ambush        = spellID(11269)
	vanish        = spellID(1856)
	premeditation = spellID(14183)
	preparation   = spellID(14185)
	stealth       = spellID(1787)
	cutthroat     = spellID(460534)
	thousandCuts  = spellID(460535)
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
		// Ghostly Strike on cooldown as a builder.
		Knob{Name: "ghostly", Default: 1, Min: 0, Max: 1, Step: 1},
		// Vanish for an Ambush (and Premeditation) at this many combo points or fewer, 9 for never.
		Knob{Name: "vanishCp", Default: 9, Min: 0, Max: 9, Step: 3},
		// Backstab only at this many Thousand Cuts stacks, or when energy is pooled. 0 ignores stacks.
		Knob{Name: "bsStacks", Default: 0, Min: 0, Max: 5, Step: 1},
		Knob{Name: "bsAfterSwing", Default: 0, Min: 0, Max: 1, Step: 0.25},
		Knob{Name: "bsEnergy", Default: 0, Min: 0, Max: 100, Step: 10},
	)
}

func (rogueBS) Build(k Knobs) *proto.APLRotation {
	c := buildRogueCooldowns(k)
	items := []*proto.APLListItem{c.sliceAndDice}
	items = append(items, c.cooldowns...)

	// Rupture is the other finisher. It snapshots Hemorrhage at cast, so the debuff only
	// has to be up then. The finisher goes out when Slice and Dice will not need the
	// points and the fight lasts long enough for the ticks.
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
			), "Hemorrhage right before a Rupture when the debuff is missing, Rupture snapshots it."+c.holdNotes))
		}
		items = append(items, cast(rupture, ruptureWhen, "Rupture when it is not ticking and Slice and Dice does not need the points."+c.holdNotes))
	}

	items = append(items, cast(eviscerate, c.evisWhen, ""))

	// Vanish is only an Ambush (and Premeditation, which needs stealth) at a low combo
	// count, so the points are not wasted over 5. Premeditation has no GCD, Ambush breaks
	// the stealth right after. Preparation gives a second Vanish once the first is spent,
	// which with Restless Blades pulling Vanish in anyway is the one use it gets. Cutthroat
	// Ambushes ride on the same line.
	if k["vanishCp"] < 9 {
		items = append(items,
			cast(vanish, and(
				le(comboPoints(), num(k["vanishCp"])),
				ge(energy(), spellCurrentCost(ambush)),
				auraIsActive(sliceAndDice),
				gt(remainingTime(), seconds(10)),
			), "Vanish for an Ambush at low combo points. Premeditation goes out in the stealth first."),
			cast(premeditation, auraIsActive(stealth), "Premeditation needs stealth, so right after Vanish."),
			cast(ambush, or(auraIsActive(stealth), auraIsActive(cutthroat)),
				"Ambush from stealth after Vanish, or when Cutthroat lets it go out after a Backstab."),
			cast(preparation, and(
				not(auraIsActive(stealth)),
				le(comboPoints(), num(k["vanishCp"])),
				gt(spellTimeToReady(vanish), seconds(60)),
				gt(remainingTime(), seconds(10)),
			), "Preparation resets Vanish (and Cold Blood) for a second Ambush once the first is spent."),
		)
	}

	if k["ghostly"] == 1 {
		items = append(items, cast(ghostlyStrike, lt(comboPoints(), num(5)),
			"Ghostly Strike is 180% weapon damage for 40 energy with a dagger, more per energy than Backstab, so on cooldown."))
	}

	// The builder. Backstab waits for Thousand Cuts stacks when asked, with the energy
	// pool as the fallback so the wait never caps energy.
	when := builderWhen(k["bsAfterSwing"], k["bsEnergy"])
	if k["hemoFiller"] == 1 {
		items = append(items, cast(hemorrhage, when, "Hemorrhage as the builder."))
	}
	if k["bsStacks"] > 0 {
		pooled := ge(energy(), num(k["bsEnergy"]))
		if k["bsEnergy"] == 0 {
			pooled = ge(energy(), num(90))
		}
		when = and(when, or(ge(auraNumStacks(thousandCuts), num(k["bsStacks"])), pooled, lt(remainingTime(), seconds(6))))
	}
	items = append(items, cast(backstab, when, "Backstab."))

	rot := rotation(items...)
	rot.PrepullActions = c.prepull
	return rot
}
