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
type paladinRet struct{}

var judgement = spellID(paladin.JudgementSpellId)
var judgementOfTheCrusader = spellID(paladin.JudgementOfTheCrusaderSpellId)

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
	}
}

func (paladinRet) Build(k Knobs) *proto.APLRotation {
	sealOfRighteousness := rankID(paladin.SealOfRighteousnessLevel[:], paladin.SealOfRighteousnessSpellId[:])
	sealOfCommand := rankID(paladin.SealOfCommandLevel[:], paladin.SealOfCommandSpellId[:])
	sealOfTheCrusader := rankID(paladin.SealOfTheCrusaderLevel[:], paladin.SealOfTheCrusaderSpellId[:])
	holyStrike := rankID(paladin.HolyStrikeLevel[:], paladin.HolyStrikeSpellId[:])
	consecration := rankID(paladin.ConsecrationLevel[:], paladin.ConsecrationSpellId[:])
	exorcism := rankID(paladin.ExorcismLevel[:], paladin.ExorcismSpellId[:])

	seal := sealOfRighteousness
	sealNotes := "Seal of Righteousness whenever it is down. Judgement does not consume it under Forever."
	if k["seal"] == 2 {
		seal = sealOfCommand
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
		if k["consecrationMana"] > 0 {
			floor = ge(currentManaPercent(), num(k["consecrationMana"]/100))
		}
		consecrationLine = cast(consecration, floor, "Consecration with the mana to spare.")
	}
	if consecrationLine != nil && k["consecrationFirst"] == 1 {
		items = append(items, consecrationLine)
	}
	items = append(items,
		cast(judgement, and(auraIsActive(seal), jotcUp), "Judgement on cooldown."),
		cast(seal, and(not(auraIsActive(seal)), jotcUp), sealNotes),
	)
	if k["holyStrike"] == 1 {
		items = append(items, cast(holyStrike, nil, "Holy Strike on cooldown."))
	}
	if k["exorcism"] == 1 {
		items = append(items, cast(exorcism, nil, "Exorcism on cooldown, only against Undead and Demons."))
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
