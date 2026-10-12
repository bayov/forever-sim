package main

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/paladin"
)

// Retribution under Forever rules. A Judgement no longer consumes the seal, so the
// paladin keeps one seal up (resealing when its 30 sec run out) and judges it on every
// cooldown. Judgement of the Crusader lasts 40 sec and every melee hit refreshes it, so
// with jotc=1 the fight opens with Seal of the Crusader and one Judgement, and the damage
// seal goes up after that. Holy Strike is the new baseline strike, Consecration is
// baseline too but costs a third of the level 20 mana pool, so it is a knob with a mana
// floor. Every line names the highest rank the player's level knows.
//
// With Twist of Light (the 31 point Retribution talent) replacing Seal of Command or Seal of
// Righteousness leaves an Echo of it, which the next melee swing applies on top of the new
// seal. Twisting puts the other damage seal up just before each swing, so every swing gets
// both seals' effects for one seal's mana. The twist knobs say how close to the swing and
// down to how much mana.
type paladinRet struct{}

var judgement = spellID(paladin.JudgementSpellId)
var judgementOfTheCrusader = spellID(paladin.JudgementOfTheCrusaderSpellId)
var echoOfCommand = spellID(1311703)
var echoOfRighteousness = spellID(1311704)

func (paladinRet) Knobs() []Knob {
	return []Knob{
		// 1 Seal of Righteousness, 2 Seal of Command (the 11 point Retribution talent).
		{Name: "seal", Default: 1, Min: 1, Max: 2, Step: 1},
		// Open with Seal of the Crusader and a Judgement so Judgement of the Crusader is on
		// the target for the fight.
		{Name: "jotc", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "holyStrike", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "consecration", Default: 0, Min: 0, Max: 1, Step: 1},
		// Consecration above the seal and judgement lines, for a pack of enemies where it
		// is most of the damage and a delayed cast is a lost tick on every one of them.
		{Name: "consecrationFirst", Default: 0, Min: 0, Max: 1, Step: 1},
		// Consecration only above this much mana, so the seals and judgements keep going.
		{Name: "consecrationMana", Default: 40, Min: 0, Max: 80, Step: 20},
		// Exorcism only lands on Undead and Demons, the line is harmless elsewhere.
		{Name: "exorcism", Default: 1, Min: 0, Max: 1, Step: 1},
		// Holy Shock on cooldown, a 20 point Holy talent from level 30 (10 sec under Forever).
		{Name: "holyShock", Default: 0, Min: 0, Max: 1, Step: 1},
		// Hammer of Wrath (from level 44) below 20% health, above this much mana. 100 never.
		{Name: "hammerOfWrathMana", Default: 100, Min: 0, Max: 100, Step: 20},
		// Twist the seals for Twist of Light's Echo: 0 never, 1 swap Seal of Command and Seal
		// of Righteousness within twistWindow tenths of a second of the next swing, above
		// twistMana percent of mana.
		{Name: "twist", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "twistWindow", Default: 10, Min: 5, Max: 30, Step: 5},
		{Name: "twistMana", Default: 40, Min: 0, Max: 80, Step: 20},
		// Holy Strike only above this much mana, 0 always.
		{Name: "holyStrikeMana", Default: 0, Min: 0, Max: 40, Step: 20},
	}
}

func (paladinRet) Build(k Knobs) *proto.APLRotation {
	sealOfRighteousness := rankID(paladin.SealOfRighteousnessLevel[:], paladin.SealOfRighteousnessSpellId[:])
	sealOfCommand := rankID(paladin.SealOfCommandLevel[:], paladin.SealOfCommandSpellId[:])
	sealOfTheCrusader := rankID(paladin.SealOfTheCrusaderLevel[:], paladin.SealOfTheCrusaderSpellId[:])
	holyStrike := rankID(paladin.HolyStrikeLevel[:], paladin.HolyStrikeSpellId[:])
	consecration := rankID(paladin.ConsecrationLevel[:], paladin.ConsecrationSpellId[:])
	exorcism := rankID(paladin.ExorcismLevel[:], paladin.ExorcismSpellId[:])
	holyShock := rankID(paladin.HolyShockLevel[:], paladin.HolyShockSpellId[:])
	hammerOfWrath := rankID(paladin.HammerOfWrathLevel[:], paladin.HammerOfWrathSpellId[:])

	seal, otherSeal := sealOfRighteousness, sealOfCommand
	sealNotes := "Seal of Righteousness whenever it is down. Judgement does not consume it under Forever."
	if k["seal"] == 2 {
		seal, otherSeal = sealOfCommand, sealOfRighteousness
		sealNotes = "Seal of Command whenever it is down. Judgement does not consume it under Forever. Without the talent ignore the warning on this line."
	}

	var items []*proto.APLListItem
	var jotcUp value
	if k["jotc"] == 1 {
		jotcUp = targetAuraIsActive(judgementOfTheCrusader)
		items = append(items,
			cast(judgement, auraIsActive(sealOfTheCrusader), "Judge Seal of the Crusader first, the judgement lasts 40 sec and every melee hit refreshes it."),
			cast(sealOfTheCrusader, and(not(jotcUp), not(auraIsActive(sealOfTheCrusader))), "Seal of the Crusader when its judgement is not on the target."),
		)
	}
	var consecrationLine *proto.APLListItem
	if k["consecration"] == 1 {
		var floor value
		notes := "Consecration on cooldown, down to the last point of mana."
		if k["consecrationMana"] > 0 {
			floor = ge(currentManaPercent(), num(k["consecrationMana"]/100))
			notes = "Consecration with the mana to spare."
		}
		consecrationLine = cast(consecration, floor, notes)
	}
	if consecrationLine != nil && k["consecrationFirst"] == 1 {
		items = append(items, consecrationLine)
	}
	twist := k["twist"] == 1
	if !twist {
		items = append(items,
			cast(judgement, and(auraIsActive(seal), jotcUp), "Judgement on cooldown."),
			cast(seal, and(not(auraIsActive(seal)), jotcUp), sealNotes),
		)
	} else {
		sealUp := or(auraIsActive(seal), auraIsActive(otherSeal))
		// One swap per swing: none while the last swap's Echo waits for its swing. Two swaps
		// before a swing would leave both Echoes up, and we do not know the server allows that.
		swingSoon := and(lt(mainHandTimeToNext(), seconds(k["twistWindow"]/10)),
			not(auraIsActive(echoOfCommand)), not(auraIsActive(echoOfRighteousness)))
		var floor value
		if k["twistMana"] > 0 {
			floor = ge(currentManaPercent(), num(k["twistMana"]/100))
		}
		items = append(items,
			cast(judgement, and(sealUp, jotcUp), "Judgement on cooldown, on whichever damage seal is up."),
			cast(seal, and(not(sealUp), jotcUp), sealNotes),
			cast(otherSeal, and(auraIsActive(seal), swingSoon, floor),
				"Twist of Light: swap to the other damage seal just before the swing, which then gets the Echo of the old one too."),
			cast(seal, and(auraIsActive(otherSeal), swingSoon, floor), "And back again before the next swing."),
		)
	}
	if k["holyStrike"] == 1 {
		var floor value
		notes := "Holy Strike on cooldown."
		if k["holyStrikeMana"] > 0 {
			floor = ge(currentManaPercent(), num(k["holyStrikeMana"]/100))
			notes = "Holy Strike on cooldown with the mana to spare."
		}
		items = append(items, cast(holyStrike, floor, notes))
	}
	if k["hammerOfWrathMana"] < 100 {
		var floor value
		if k["hammerOfWrathMana"] > 0 {
			floor = ge(currentManaPercent(), num(k["hammerOfWrathMana"]/100))
		}
		items = append(items, cast(hammerOfWrath, floor, "Hammer of Wrath below 20% health."))
	}
	if k["exorcism"] == 1 {
		items = append(items, cast(exorcism, nil, "Exorcism on cooldown, only against Undead and Demons."))
	}
	if k["holyShock"] == 1 {
		items = append(items, cast(holyShock, nil, "Holy Shock on cooldown. Without the talent ignore the warning on this line."))
	}
	if consecrationLine != nil && k["consecrationFirst"] == 0 {
		items = append(items, consecrationLine)
	}
	items = append(items, autocastOtherCooldowns(nil, "Racials and trinkets on a free GCD."))

	rot := rotation(items...)
	opener := seal
	if k["jotc"] == 1 {
		opener = sealOfTheCrusader
	}
	rot.PrepullActions = append(rot.PrepullActions, prepull(opener, 1.5))
	return rot
}
