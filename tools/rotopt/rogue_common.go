package main

import (
	"fmt"

	"github.com/wowsims/classic/sim/core/proto"
)

// The cooldown rules every rogue rotation shares, with the knobs that tune them.
//
// Adrenaline Rush is the anchor of the big window. The timed buffs that multiply with
// the extra energy (Blood Fury, Elune's Light) wait for it whenever the wait costs them
// no use. Cold Blood and Eureka! are spent by the next hits, so they wait for a 5 point
// Eviscerate, which also lines them up with each other. Blade Flurry and Berserking are
// auto attack speed, which does not feed the energy window, so they go on cooldown
// unless the optimizer finds that holding them helps.

var (
	sinisterStrike = spellIDRank(11294, 8)
	sliceAndDice   = spellIDRank(6774, 2)
	eviscerate     = spellID(31016)
	adrenalineRush = spellID(13750)
	bladeFlurry    = spellID(13877)
	coldBlood      = spellID(14177)
	thistleTea     = itemID(7676)
	bloodFury      = spellID(20572)
	berserking     = spellIDTag(26297, 2)
	elunesLight    = spellID(460531)
	eureka         = spellID(460532)
	earthstrike    = itemID(21180)
)

// The 15 second buffs are only worth a cast when the fight lasts that long after it.
var buffDuration = seconds(15)

func rogueCooldownKnobs() []Knob {
	return []Knob{
		{Name: "sndRefresh", Default: 3, Min: 1, Max: 9, Step: 2},
		// A fresh Slice and Dice (none up) waits for this many combo points. At 60 a 1 point
		// one goes up at once. At level 20 a Backstab is 60 energy, so a point comes every
		// 6 sec and a 1 point Slice and Dice (9 sec) is gone before the next point: the
		// rogue would refresh it forever and never reach a Rupture. Waiting for 4 points
		// (18 sec) is worth 1.4 DPS there. The refresh before it drops stays at 5 points,
		// every lower value lost.
		{Name: "sndCp", Default: 1, Min: 1, Max: 5, Step: 1},
		{Name: "arEnergy", Default: 59, Min: 39, Max: 99, Step: 10},
		{Name: "arSnd", Default: 1, Min: 0, Max: 18, Step: 3},
		{Name: "bfSnd", Default: 1, Min: 0, Max: 24, Step: 3},
		{Name: "bfHoldForAr", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "cbSnd", Default: 6, Min: 0, Max: 24, Step: 3},
		// Build to 5 points for a Cold Blood Eviscerate while Cold Blood is ready, instead
		// of spending them on a Rupture. Only matters with a Rupture threshold under 5.
		{Name: "cbPool", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "eurekaSnd", Default: 8, Min: 0, Max: 12, Step: 2},
		// Eureka! right before a 5 point Eviscerate (5), or on cooldown (0). Below 60 the
		// finishers go out at 3 points and 5 never comes.
		{Name: "eurekaCp", Default: 5, Min: 0, Max: 5, Step: 5},
		{Name: "teaEnergy", Default: 10, Min: 0, Max: 50, Step: 10},
		{Name: "bloodFurySnd", Default: 0, Min: 0, Max: 24, Step: 4},
		{Name: "bloodFuryHoldForAr", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "trinketSnd", Default: 0, Min: 0, Max: 24, Step: 4},
		{Name: "trinketHoldForAr", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "eluneHoldForAr", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "berserkingHoldForBf", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "evisEnergy", Default: 79, Min: 59, Max: 99, Step: 10},
		{Name: "bfPrepull", Default: 0, Min: 0, Max: 1, Step: 1},
	}
}

// rogueCooldowns holds the shared rules built from one set of knobs. A template lays
// them out around its own builders and finishers.
type rogueCooldowns struct {
	// sliceAndDice is the Slice and Dice line, the first thing in every rotation.
	sliceAndDice *proto.APLListItem
	// cooldowns are the lines between Slice and Dice and the finishers, in order.
	cooldowns []*proto.APLListItem
	// fivePointEviscerate is true when the next action can be a 5 point Eviscerate.
	fivePointEviscerate value
	// notRefreshing is true when the next finisher is not going to be a Slice and Dice
	// refresh for a waiting cooldown. Nil when no cooldown waits on Slice and Dice.
	notRefreshing value
	// evisWhen is the condition on the plain 5 point Eviscerate.
	evisWhen value
	// ruptureHold is true while Rupture and the Slice and Dice refresh should leave the
	// points alone, nil when nothing is waiting on them.
	ruptureHold value
	// holdNotes is the sentence for the lines ruptureHold is on, empty when it is nil.
	holdNotes string
	// prepull holds the Blade Flurry prepull when the knob asks for one.
	prepull []*proto.APLPrepullAction
}

func buildRogueCooldowns(k Knobs) rogueCooldowns {
	// Slice and Dice left, or nil when the knob is 0 so the rule drops out entirely.
	sndAtLeast := func(secs float64) value {
		if secs == 0 {
			return nil
		}
		return ge(auraRemainingTime(sliceAndDice), seconds(secs))
	}
	// Adrenaline Rush is 31 points into Combat and Blade Flurry 21, so their lines only
	// exist once the build can have them. A line for a spell the character lacks would
	// only warn, but a condition that mentions one is quietly dropped by the sim, and
	// the Slice and Dice window clause below would come out as a bare "5 combo points"
	// once every cooldown in it is missing: at level 20 that refreshed Slice and Dice on
	// every 5 points and never let Cold Blood fire.
	hasAdrenalineRush := maxTalentPoints >= 31
	hasBladeFlurry := maxTalentPoints >= 21
	// When a cooldown wants a long Slice and Dice, a 5 point refresh goes out first so
	// the cooldown is not waiting on a 1 point one to run out. The Slice and Dice check
	// goes through the spell's time to ready (zero once it is ready) so that a cooldown the
	// character does not have (Blood Fury on a Gnome, no Earthstrike) takes its whole
	// clause out instead of leaving a bare Slice and Dice check behind.
	var waiting []value
	for _, w := range []struct {
		spell *proto.ActionID
		snd   float64
		has   bool
	}{
		{adrenalineRush, k["arSnd"], hasAdrenalineRush},
		{bladeFlurry, k["bfSnd"], hasBladeFlurry},
		{bloodFury, k["bloodFurySnd"], true},
		{earthstrike, k["trinketSnd"], true},
	} {
		if w.snd > 0 && w.has {
			waiting = append(waiting, and(
				spellIsReady(w.spell),
				gt(spellUsesRemaining(w.spell, buffDuration), num(0)),
				lt(add(auraRemainingTime(sliceAndDice), spellTimeToReady(w.spell)), seconds(w.snd)),
			))
		}
	}
	var sndForWindow value
	if len(waiting) > 0 {
		sndForWindow = and(ge(comboPoints(), num(5)), or(waiting...))
	}

	c := rogueCooldowns{}
	// While Cold Blood is ready the 5 points are for its Eviscerate: neither Rupture nor
	// the Slice and Dice refresh gets them. Slice and Dice goes back up on the point
	// Ruthlessness leaves after the Eviscerate.
	if k["cbPool"] == 1 {
		c.ruptureHold = not(spellIsReady(coldBlood))
	}
	// Enough for a full Eviscerate. Cold Blood and Eureka! wait for this.
	c.fivePointEviscerate = and(
		ge(comboPoints(), num(5)),
		ge(energy(), spellCurrentCost(eviscerate)),
	)
	// Not when the next finisher is going to be that Slice and Dice refresh instead.
	if sndForWindow != nil {
		c.notRefreshing = not(sndForWindow)
	}
	holdFor := func(spell *proto.ActionID, anchor *proto.ActionID, hold bool) value {
		if !hold || (anchor == adrenalineRush && !hasAdrenalineRush) || (anchor == bladeFlurry && !hasBladeFlurry) {
			return nil
		}
		return alignedWith(spell, anchor)
	}

	// The notes say what the knobs made of the line, so the file reads right in the UI.
	sndNotes := "Slice and Dice up at any combo points, refresh at 5 when it is about to drop"
	if k["sndCp"] > 1 {
		sndNotes = fmt.Sprintf("Slice and Dice at %g combo points once it is down (a short one at fewer points would eat every point before a Rupture), refresh at 5 when it is about to drop", k["sndCp"])
	}
	if sndForWindow != nil {
		sndNotes += " or when Adrenaline Rush is waiting on a longer one"
	}
	sndNotes += "."
	if c.ruptureHold != nil {
		c.holdNotes = " While Cold Blood is ready the points are kept for its Eviscerate. Without the talent ignore the warning on this line."
		sndNotes += c.holdNotes
	}
	c.sliceAndDice = cast(sliceAndDice, or(
		and(ge(comboPoints(), num(k["sndCp"])), not(auraIsActive(sliceAndDice)), ge(remainingTime(), seconds(6))),
		and(ge(comboPoints(), num(5)), lt(auraRemainingTime(sliceAndDice), seconds(k["sndRefresh"])), gt(remainingTime(), seconds(9)), c.ruptureHold),
		sndForWindow,
	), sndNotes)

	// Forever regenerates energy continuously, so there is no tick to line Adrenaline
	// Rush up with: a tick's worth of energy no longer arrives in one lump.
	if hasAdrenalineRush {
		c.cooldowns = append(c.cooldowns, cast(adrenalineRush, and(
			lt(energy(), num(k["arEnergy"])),
			sndAtLeast(k["arSnd"]),
		), "Adrenaline Rush anchors the big window. Cast it low on energy so none of the extra regen is wasted."))
	}
	if hasBladeFlurry {
		c.cooldowns = append(c.cooldowns, cast(bladeFlurry, and(
			sndAtLeast(k["bfSnd"]),
			holdFor(bladeFlurry, adrenalineRush, k["bfHoldForAr"] == 1),
		), "Blade Flurry is auto attack speed and does not feed the energy window, so it goes on cooldown."))
	}
	eurekaLine := cast(eureka, and(c.fivePointEviscerate, sndAtLeast(k["eurekaSnd"]), c.notRefreshing),
		"Gnome. Eureka! buffs the next 3 abilities, so right before Eviscerate and the two builders after it. Other races can ignore the warning on this line.")
	if k["eurekaCp"] == 0 {
		eurekaLine = cast(eureka, sndAtLeast(k["eurekaSnd"]),
			"Gnome. Eureka! on cooldown, the next 3 abilities cost half. Other races can ignore the warning on this line.")
	}
	berserkingNotes := "Berserking is auto attack speed, so it goes on cooldown."
	if k["berserkingHoldForBf"] == 1 && hasBladeFlurry {
		berserkingNotes = "Berserking waits for Blade Flurry when the wait costs no use."
	}
	c.cooldowns = append(c.cooldowns,
		cast(coldBlood, and(c.fivePointEviscerate, sndAtLeast(k["cbSnd"]), c.notRefreshing),
			"Cold Blood is spent by the next hit, so only ever right before a 5 point Eviscerate."),

		eurekaLine,

		cast(bloodFury, and(sndAtLeast(k["bloodFurySnd"]), holdFor(bloodFury, adrenalineRush, k["bloodFuryHoldForAr"] == 1)),
			"Orc. "+timedBuffNotes("Blood Fury", k["bloodFuryHoldForAr"] == 1 && hasAdrenalineRush, k["bloodFurySnd"])+" Other races can ignore the warning on this line."),

		cast(elunesLight, holdFor(elunesLight, adrenalineRush, k["eluneHoldForAr"] == 1),
			"Night Elf. "+timedBuffNotes("Elune's Light", k["eluneHoldForAr"] == 1 && hasAdrenalineRush, 0)+" Other races can ignore the warning on this line."),

		cast(berserking, holdFor(berserking, bladeFlurry, k["berserkingHoldForBf"] == 1),
			"Troll. "+berserkingNotes+" Other races can ignore the warning on this line."),

		cast(earthstrike, and(sndAtLeast(k["trinketSnd"]), holdFor(earthstrike, adrenalineRush, k["trinketHoldForAr"] == 1)),
			timedBuffNotes("Earthstrike", k["trinketHoldForAr"] == 1 && hasAdrenalineRush, k["trinketSnd"])+" Without the trinket ignore the warning on this line."),

		cast(thistleTea, le(energy(), num(k["teaEnergy"])),
			"Thistle Tea when there is room for the 100 energy."),

		autocastOtherCooldowns(auraIsActive(sliceAndDice), "Anything not listed above (Sapper, trinkets)."),
	)

	c.evisWhen = and(
		ge(comboPoints(), num(5)),
		or(auraIsActive(sliceAndDice), ge(energy(), num(k["evisEnergy"])), lt(remainingTime(), seconds(6))),
	)

	if k["bfPrepull"] == 1 {
		c.prepull = []*proto.APLPrepullAction{prepull(bladeFlurry, 1)}
	}
	return c
}

// timedBuffNotes says what a timed buff's line waits for, from the same knobs as its
// condition.
func timedBuffNotes(name string, holdsForAr bool, snd float64) string {
	switch {
	case holdsForAr && snd > 0:
		return fmt.Sprintf("%s waits for Adrenaline Rush when the wait costs no use, and for a Slice and Dice with %g sec left.", name, snd)
	case holdsForAr:
		return name + " waits for Adrenaline Rush when the wait costs no use."
	case snd > 0:
		return fmt.Sprintf("%s waits for a Slice and Dice with %g sec left.", name, snd)
	}
	return name + " on cooldown."
}

// builderWhen is the condition for the filler builder: right after a main hand swing,
// or once energy is pooled. Extra attacks (Hack and Slash, Hand of Justice) move the main
// hand swing to now and restart the timer, so a proc off a builder cast late in the swing
// wins almost nothing, the swing was about to land anyway. Right after a swing the same
// proc is a whole free swing. The energy fallback keeps the wait from capping energy.
// Both knobs at 0 means on every GCD it can afford.
func builderWhen(afterSwing, pooledEnergy float64) value {
	if afterSwing == 0 && pooledEnergy == 0 {
		return nil
	}
	var swing, pooled value
	if afterSwing > 0 {
		swing = gt(mainHandTimeToNext(), sub(mainHandSwingTime(), seconds(afterSwing)))
	}
	if pooledEnergy > 0 {
		pooled = ge(energy(), num(pooledEnergy))
	}
	return or(swing, pooled, lt(remainingTime(), seconds(6)))
}

// alignedWith is true when spell can go out now without giving up an overlap with the
// anchor: the anchor is up, it is never coming back, or waiting for it would cost the
// spell a use. Spells this gates wait for the anchor otherwise.
func alignedWith(spell *proto.ActionID, anchor *proto.ActionID) value {
	return or(
		auraIsActive(anchor),
		eq(spellUsesRemaining(anchor, buffDuration), num(0)),
		gt(spellUsesLostByDelay(spell, spellTimeToReady(anchor), buffDuration), num(0)),
	)
}
