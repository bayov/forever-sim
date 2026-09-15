package main

import (
	"github.com/wowsims/classic/sim/core/proto"
)

// Combat Sinister Strike under Forever rules (Restless Blades and the new racials).
//
// Adrenaline Rush is the anchor of the big window. The timed buffs that multiply with
// the extra energy (Blood Fury, Elune's Light) wait for it whenever the wait costs them
// no use. Cold Blood and Eureka! are spent by the next hits, so they wait for a 5 point
// Eviscerate, which also lines them up with each other. Blade Flurry and Berserking are
// auto attack speed, which does not feed the energy window, so they go on cooldown
// unless the optimizer finds that holding them helps.
type rogueSS struct{}

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

func (rogueSS) Knobs() []Knob {
	return []Knob{
		{Name: "sndRefresh", Default: 3, Min: 1, Max: 9, Step: 2},
		{Name: "arEnergy", Default: 59, Min: 39, Max: 99, Step: 10},
		{Name: "arSnd", Default: 1, Min: 0, Max: 18, Step: 3},
		{Name: "bfSnd", Default: 1, Min: 0, Max: 24, Step: 3},
		{Name: "bfHoldForAr", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "cbSnd", Default: 6, Min: 0, Max: 24, Step: 3},
		{Name: "eurekaSnd", Default: 8, Min: 0, Max: 12, Step: 2},
		{Name: "teaEnergy", Default: 10, Min: 0, Max: 50, Step: 10},
		{Name: "bloodFurySnd", Default: 0, Min: 0, Max: 24, Step: 4},
		{Name: "bloodFuryHoldForAr", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "trinketSnd", Default: 0, Min: 0, Max: 24, Step: 4},
		{Name: "trinketHoldForAr", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "eluneHoldForAr", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "berserkingHoldForBf", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "evisEnergy", Default: 79, Min: 59, Max: 99, Step: 10},
		{Name: "ssAfterSwing", Default: 0.5, Min: 0, Max: 1, Step: 0.25},
		{Name: "ssEnergy", Default: 79, Min: 0, Max: 100, Step: 10},
		{Name: "bfPrepull", Default: 1, Min: 0, Max: 1, Step: 1},
	}
}

func (rogueSS) Build(k Knobs) *proto.APLRotation {
	// Slice and Dice left, or nil when the knob is 0 so the rule drops out entirely.
	sndAtLeast := func(secs float64) value {
		if secs == 0 {
			return nil
		}
		return ge(auraRemainingTime(sliceAndDice), seconds(secs))
	}
	// When a cooldown wants a long Slice and Dice, a 5 point refresh goes out first so
	// the cooldown is not waiting on a 1 point one to run out. The Slice and Dice check
	// goes through the spell's time to ready (zero once it is ready) so that a cooldown the
	// character does not have (Blood Fury on a Gnome, no Earthstrike) takes its whole
	// clause out instead of leaving a bare Slice and Dice check behind.
	var waiting []value
	for _, w := range []struct {
		spell *proto.ActionID
		snd   float64
	}{
		{adrenalineRush, k["arSnd"]},
		{bladeFlurry, k["bfSnd"]},
		{bloodFury, k["bloodFurySnd"]},
		{earthstrike, k["trinketSnd"]},
	} {
		if w.snd > 0 {
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

	// Enough for a full Eviscerate. Cold Blood and Eureka! wait for this.
	fivePointEviscerate := and(
		ge(comboPoints(), num(5)),
		ge(energy(), spellCurrentCost(eviscerate)),
	)
	// Not when the next finisher is going to be that Slice and Dice refresh instead.
	var notRefreshing value
	if sndForWindow != nil {
		notRefreshing = not(sndForWindow)
	}
	holdFor := func(spell *proto.ActionID, anchor *proto.ActionID, hold bool) value {
		if !hold {
			return nil
		}
		return alignedWith(spell, anchor)
	}

	// Sinister Strike right after a main hand swing, or once energy is pooled. Extra
	// attacks (Hack and Slash, Hand of Justice) move the main hand swing to now and
	// restart the timer, so a proc off a Sinister Strike cast late in the swing wins
	// almost nothing, the swing was about to land anyway. Right after a swing the same
	// proc is a whole free swing. The energy fallback keeps the wait from capping energy.
	// Both knobs at 0 means on every GCD it can afford.
	var ssWhen value
	if k["ssAfterSwing"] > 0 || k["ssEnergy"] > 0 {
		var afterSwing, pooled value
		if k["ssAfterSwing"] > 0 {
			afterSwing = gt(mainHandTimeToNext(), sub(mainHandSwingTime(), seconds(k["ssAfterSwing"])))
		}
		if k["ssEnergy"] > 0 {
			pooled = ge(energy(), num(k["ssEnergy"]))
		}
		ssWhen = or(afterSwing, pooled, lt(remainingTime(), seconds(6)))
	}

	rot := rotation(
		cast(sliceAndDice, or(
			and(ge(comboPoints(), num(1)), not(auraIsActive(sliceAndDice)), ge(remainingTime(), seconds(6))),
			and(ge(comboPoints(), num(5)), lt(auraRemainingTime(sliceAndDice), seconds(k["sndRefresh"])), gt(remainingTime(), seconds(9))),
			sndForWindow,
		), "Slice and Dice up at any combo points, refresh at 5 when it is about to drop or when Adrenaline Rush is waiting on a longer one."),

		cast(adrenalineRush, and(
			lt(energy(), num(k["arEnergy"])),
			lt(timeToEnergyTick(), seconds(1)),
			sndAtLeast(k["arSnd"]),
		), "Adrenaline Rush anchors the big window. Cast it low on energy so none of the extra regen is wasted."),

		cast(bladeFlurry, and(
			sndAtLeast(k["bfSnd"]),
			holdFor(bladeFlurry, adrenalineRush, k["bfHoldForAr"] == 1),
		), "Blade Flurry is auto attack speed and does not feed the energy window, so it goes on cooldown."),

		cast(coldBlood, and(fivePointEviscerate, sndAtLeast(k["cbSnd"]), notRefreshing),
			"Cold Blood is spent by the next hit, so only ever right before a 5 point Eviscerate."),

		cast(eureka, and(fivePointEviscerate, sndAtLeast(k["eurekaSnd"]), notRefreshing),
			"Gnome. Eureka! buffs the next 3 abilities, so right before Eviscerate and the two builders after it. Other races can ignore the warning on this line."),

		cast(bloodFury, and(sndAtLeast(k["bloodFurySnd"]), holdFor(bloodFury, adrenalineRush, k["bloodFuryHoldForAr"] == 1)),
			"Orc. Blood Fury waits for Adrenaline Rush when the wait costs no use. Other races can ignore the warning on this line."),

		cast(elunesLight, holdFor(elunesLight, adrenalineRush, k["eluneHoldForAr"] == 1),
			"Night Elf. Elune's Light waits for Adrenaline Rush when the wait costs no use. Other races can ignore the warning on this line."),

		cast(berserking, holdFor(berserking, bladeFlurry, k["berserkingHoldForBf"] == 1),
			"Troll. Berserking is auto attack speed like Blade Flurry, so it goes on cooldown. Other races can ignore the warning on this line."),

		cast(earthstrike, and(sndAtLeast(k["trinketSnd"]), holdFor(earthstrike, adrenalineRush, k["trinketHoldForAr"] == 1)),
			"Earthstrike waits for Adrenaline Rush or a long Slice and Dice. Without the trinket ignore the warning on this line."),

		cast(thistleTea, and(le(energy(), num(k["teaEnergy"])), gt(timeToEnergyTick(), seconds(1))),
			"Thistle Tea when there is room for the 100 energy. Right after a tick, since the next one would go over the cap before the next ability spends any."),

		autocastOtherCooldowns(auraIsActive(sliceAndDice), "Anything not listed above (Sapper, trinkets)."),

		cast(eviscerate, and(
			ge(comboPoints(), num(5)),
			or(auraIsActive(sliceAndDice), ge(energy(), num(k["evisEnergy"])), lt(remainingTime(), seconds(6))),
		), ""),

		cast(sinisterStrike, ssWhen, "Sinister Strike right after a main hand swing so an extra attack proc restarts a swing that just happened, or when energy is about to cap."),
	)
	if k["bfPrepull"] == 1 {
		rot.PrepullActions = []*proto.APLPrepullAction{prepull(bladeFlurry, 1)}
	}
	return rot
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
