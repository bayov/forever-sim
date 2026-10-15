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

const syntheticSpellPowerOnUseTrinket = paladin.SyntheticSpellPowerOnUseTrinket
const goblinSapperCharge = 10646

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
		// Holy Shock on cooldown, a 20 point Holy talent from level 30 (10 sec under Forever):
		// 0 never, 1 after Holy Strike, 2 ahead of it.
		{Name: "holyShock", Default: 0, Min: 0, Max: 2, Step: 1},
		// Holy Shock only above this much mana, 0 always.
		{Name: "holyShockMana", Default: 0, Min: 0, Max: 60, Step: 20},
		// Light's Vigil (the 31 point Holy talent) when Holy Shock is ready within its 1.5 sec
		// cast, above this much mana, so the Holy Shock right after it detonates it and starts
		// no cooldown. 100 never.
		{Name: "lightsVigilMana", Default: 100, Min: 0, Max: 100, Step: 10},
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
		// Holy Strike ahead of the seal and twist lines instead of after them.
		{Name: "holyStrikeFirst", Default: 0, Min: 0, Max: 1, Step: 1},
		// With twisting, judge Seal of Righteousness only when the next swing is more than this
		// many tenths of a second away, so most Judgements land on Seal of Command (its
		// judgement rolls on the melee table and hits harder). 0 judges whichever seal is up.
		{Name: "judgeRighteousnessWindow", Default: 0, Min: 0, Max: 30, Step: 5},
		// Put the other seal up before the primary one in the prepull, so the first swing
		// gets its Echo (twisting), or Light's Vigil on the target (Holy Shock builds).
		{Name: "prepullTwo", Default: 0, Min: 0, Max: 1, Step: 1},
		// Light's Vigil whenever it is off cooldown and not on the target, above
		// lightsVigilMana, with Holy Shock waiting for the Vigil (or the mana floor) instead of
		// going out on cooldown. 0 ties the Vigil to Holy Shock's cooldown.
		{Name: "vigilChain", Default: 0, Min: 0, Max: 1, Step: 1},
		// The synthetic spell power on-use trinket and the Goblin Sapper Charge on cooldown at
		// the top of the list. On the autocast line at the bottom they wait for a free GCD,
		// which a GCD line above it nearly always takes, so they go out seconds late.
		{Name: "cooldownsFirst", Default: 0, Min: 0, Max: 1, Step: 1},
		// Against Undead and Demons: Exorcism ahead of Holy Strike instead of last, and only
		// above exorcismMana percent of mana.
		{Name: "exorcismFirst", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "exorcismMana", Default: 0, Min: 0, Max: 60, Step: 20},
		// Holy Wrath (from level 50, 2 sec cast, 1 min cooldown, only hits Undead and Demons)
		// after Exorcism above holyWrathMana percent of mana. 100 never.
		{Name: "holyWrathMana", Default: 100, Min: 0, Max: 100, Step: 20},
		// With cooldownsFirst, the on-use trinket at the pull and then held for the 20% execute
		// phase whenever holding it there loses no use of it in the fight.
		{Name: "trinketExecute", Default: 0, Min: 0, Max: 1, Step: 1},
		// In fights longer than vigilLongFight seconds, Light's Vigil (and Holy Shock's low mana
		// branch) keeps a higher floor of vigilLongMana percent, so the mana lasts; shorter
		// fights use lightsVigilMana. 0 off.
		{Name: "vigilLongFight", Default: 0, Min: 0, Max: 300, Step: 30},
		{Name: "vigilLongMana", Default: 60, Min: 30, Max: 80, Step: 10},
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
	lightsVigil := rankID(paladin.LightsVigilLevel[:], paladin.LightsVigilSpellId[:])
	lightsVigilMark := rankID(paladin.LightsVigilLevel[:], paladin.LightsVigilDebuffId[:])
	holyWrath := rankID(paladin.HolyWrathLevel[:], paladin.HolyWrathSpellId[:])

	seal, otherSeal := sealOfRighteousness, sealOfCommand
	sealNotes := "Seal of Righteousness whenever it is down. Judgement does not consume it under Forever."
	if k["seal"] == 2 {
		seal, otherSeal = sealOfCommand, sealOfRighteousness
		sealNotes = "Seal of Command whenever it is down. Judgement does not consume it under Forever. Without the talent ignore the warning on this line."
	}

	var items []*proto.APLListItem
	if k["cooldownsFirst"] == 1 {
		items = append(items,
			trinketLine(k["trinketExecute"] == 1),
			cast(itemID(goblinSapperCharge), nil, "Goblin Sapper Charge on cooldown."),
		)
	}
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
	holyStrikeLine := func() {
		if k["holyStrike"] != 1 {
			return
		}
		var floor value
		notes := "Holy Strike on cooldown."
		if k["holyStrikeMana"] > 0 {
			floor = ge(currentManaPercent(), num(k["holyStrikeMana"]/100))
			notes = "Holy Strike on cooldown with the mana to spare."
		}
		items = append(items, cast(holyStrike, floor, notes))
	}
	if k["holyStrikeFirst"] == 1 {
		holyStrikeLine()
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
		judgeWhen, judgeNotes := sealUp, "Judgement on cooldown, on whichever damage seal is up."
		if k["judgeRighteousnessWindow"] > 0 {
			judgeWhen = or(auraIsActive(sealOfCommand), and(auraIsActive(sealOfRighteousness), gt(mainHandTimeToNext(), seconds(k["judgeRighteousnessWindow"]/10))))
			judgeNotes = "Judgement on cooldown on Seal of Command, and on Seal of Righteousness only when the next swing is a while off (Judgement of Command hits harder)."
		}
		items = append(items,
			cast(judgement, and(judgeWhen, jotcUp), judgeNotes),
			cast(seal, and(not(sealUp), jotcUp), sealNotes),
			cast(otherSeal, and(auraIsActive(seal), swingSoon, floor),
				"Twist of Light: swap to the other damage seal just before the swing, which then gets the Echo of the old one too."),
			cast(seal, and(auraIsActive(otherSeal), swingSoon, floor), "And back again before the next swing."),
		)
	}
	holyShockLine := func() {
		if k["holyShock"] == 0 {
			return
		}
		var floor value
		notes := "Holy Shock on cooldown. Without the talent ignore the warning on this line."
		if k["holyShockMana"] > 0 {
			floor = ge(currentManaPercent(), num(k["holyShockMana"]/100))
			notes = "Holy Shock on cooldown with the mana to spare. Without the talent ignore the warning on this line."
		}
		if k["lightsVigilMana"] < 100 {
			var vigilFloor value
			if k["lightsVigilMana"] > 0 {
				vigilFloor = ge(currentManaPercent(), num(k["lightsVigilMana"]/100))
			}
			if k["vigilLongFight"] > 0 {
				// The fight is short when the time left is under vigilLongFight minus the time spent.
				now := &proto.APLValue{Value: &proto.APLValue_CurrentTime{CurrentTime: &proto.APLValueCurrentTime{}}}
				shortFight := lt(remainingTime(), sub(seconds(k["vigilLongFight"]), now))
				vigilFloor = and(vigilFloor, or(shortFight, ge(currentManaPercent(), num(k["vigilLongMana"]/100))))
			}
			if k["vigilChain"] == 1 {
				items = append(items, cast(lightsVigil, and(vigilFloor, not(targetAuraIsActive(lightsVigilMark))),
					"Light's Vigil whenever it is not on the target, with the mana to spare. Without the talent ignore the warning on this line."))
				lowMana := constant("false")
				if k["vigilLongFight"] > 0 {
					lowMana = not(vigilFloor)
				} else if k["lightsVigilMana"] > 0 {
					lowMana = lt(currentManaPercent(), num(k["lightsVigilMana"]/100))
				}
				items = append(items, cast(holyShock, and(or(targetAuraIsActive(lightsVigilMark), lowMana), floor),
					"Holy Shock on the Vigil, so it detonates it and starts no cooldown, or on cooldown once the mana is too low for the Vigil."))
				return
			}
			items = append(items, cast(lightsVigil, and(le(spellTimeToReady(holyShock), seconds(1.5)), vigilFloor),
				"Light's Vigil just before Holy Shock is ready, so the Holy Shock detonates it and starts no cooldown. Without the talent ignore the warning on this line."))
		}
		items = append(items, cast(holyShock, floor, notes))
	}
	if k["holyShock"] == 2 {
		holyShockLine()
	}
	exorcismLine := func() {
		if k["exorcism"] != 1 {
			return
		}
		var floor value
		notes := "Exorcism on cooldown, only against Undead and Demons."
		if k["exorcismMana"] > 0 {
			floor = ge(currentManaPercent(), num(k["exorcismMana"]/100))
			notes = "Exorcism on cooldown with the mana to spare, only against Undead and Demons."
		}
		items = append(items, cast(exorcism, floor, notes))
	}
	if k["exorcismFirst"] == 1 {
		exorcismLine()
	}
	if k["holyStrikeFirst"] == 0 {
		holyStrikeLine()
	}
	if k["hammerOfWrathMana"] < 100 {
		var floor value
		if k["hammerOfWrathMana"] > 0 {
			floor = ge(currentManaPercent(), num(k["hammerOfWrathMana"]/100))
		}
		items = append(items, cast(hammerOfWrath, floor, "Hammer of Wrath below 20% health."))
	}
	if k["exorcismFirst"] == 0 {
		exorcismLine()
	}
	if k["holyWrathMana"] < 100 {
		var floor value
		if k["holyWrathMana"] > 0 {
			floor = ge(currentManaPercent(), num(k["holyWrathMana"]/100))
		}
		items = append(items, cast(holyWrath, floor, "Holy Wrath on cooldown, only against Undead and Demons."))
	}
	if k["holyShock"] == 1 {
		holyShockLine()
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
	if k["prepullTwo"] == 1 {
		first, second := otherSeal, seal
		if k["holyShock"] > 0 && k["lightsVigilMana"] < 100 {
			first, second = seal, lightsVigil
		}
		if k["jotc"] == 1 {
			// The raid duty: the same two pre-pull casts 1.5 sec earlier, then Seal of the
			// Crusader last, judged at the pull. Replacing a damage seal with it still leaves
			// Twist of Light's Echo for the first swing.
			rot.PrepullActions = append(rot.PrepullActions, prepull(first, 4.5), prepull(second, 3))
		} else {
			rot.PrepullActions = append(rot.PrepullActions, prepull(first, 3))
			opener = second
		}
	}
	rot.PrepullActions = append(rot.PrepullActions, prepull(opener, 1.5))
	return rot
}

// trinketLine is the spell power on-use trinket on cooldown, or with holdForExecute at the pull
// and then held for the 20% execute phase when that loses no use. The time until execute is
// (4 x remaining - now) / 5: APL math truncates a duration times a fractional constant to 0, so
// it multiplies by 4 and divides by 5 instead of taking 0.8.
func trinketLine(holdForExecute bool) *proto.APLListItem {
	trinket := itemID(syntheticSpellPowerOnUseTrinket)
	if !holdForExecute {
		return cast(trinket, nil, "The spell power on-use trinket on cooldown.")
	}
	now := &proto.APLValue{Value: &proto.APLValue_CurrentTime{CurrentTime: &proto.APLValueCurrentTime{}}}
	math := func(op proto.APLValueMath_MathOperator, lhs, rhs value) value {
		return &proto.APLValue{Value: &proto.APLValue_Math{Math: &proto.APLValueMath{Op: op, Lhs: lhs, Rhs: rhs}}}
	}
	untilExecute := math(proto.APLValueMath_OpDiv, sub(math(proto.APLValueMath_OpMul, remainingTime(), num(4)), now), num(5))
	execute := &proto.APLValue{Value: &proto.APLValue_IsExecutePhase{IsExecutePhase: &proto.APLValueIsExecutePhase{Threshold: proto.APLValueIsExecutePhase_E20}}}
	return cast(trinket, or(lt(now, seconds(1)), execute, gt(spellUsesLostByDelay(trinket, untilExecute, seconds(20)), num(0))),
		"The spell power on-use trinket at the pull, then held for the execute phase when that loses no use.")
}
