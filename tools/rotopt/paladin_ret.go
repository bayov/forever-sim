package main

import (
	"fmt"

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
// With Unsealed Wrath (the 31 point Retribution talent since the launch realms' October 30
// hotfix, sim/paladin/unsealed_wrath.go) every landed white swing stores a copy of the active
// seal, up to 3 for 15 sec, and the first landed swing under the other damage seal spends them
// all, each copy getting the stored seal's own chance. Twisting keeps one seal up while its
// copies build to twistStacks, puts the other one up for the swing that spends them (which also
// stores a copy of the other), and goes back once the other has twistBack copies. The knobs say
// how many copies, how close to the swing and down to how much mana.
type paladinRet struct{}

var judgement = spellID(paladin.JudgementSpellId)
var judgementOfTheCrusader = spellID(paladin.JudgementOfTheCrusaderSpellId)
var judgementOfWisdom = spellID(20355) // the rank 3 debuff (sim/paladin/sow.go)
var echoOfCommand = spellID(paladin.EchoOfCommandSpellId)
var echoOfRighteousness = spellID(paladin.EchoOfRighteousnessSpellId)
var swiftJudgement = spellID(paladin.SwiftJudgementSpellId)

const syntheticSpellPowerOnUseTrinket = paladin.SyntheticSpellPowerOnUseTrinket
const goblinSapperCharge = 10646
const majorManaPotion = 13444
const demonicRune = 12662

func (paladinRet) Knobs() []Knob {
	return []Knob{
		// 1 Seal of Righteousness, 2 Seal of Command (the 11 point Retribution talent).
		{Name: "seal", Default: 1, Min: 1, Max: 2, Step: 1},
		// The raid duty (standards 23): 1 open with Seal of the Crusader and a Judgement so
		// Judgement of the Crusader is on the target for the fight, 2 the same with Seal of Wisdom
		// for Judgement of Wisdom. 0 none, the raid's paladins keep them up.
		{Name: "jotc", Default: 1, Min: 0, Max: 2, Step: 1},
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
		// Hammer of Wrath (from level 44) below 20% health, above this much mana. 100 never.
		{Name: "hammerOfWrathMana", Default: 100, Min: 0, Max: 100, Step: 20},
		// Where Hammer of Wrath goes: 0 after Holy Strike, 1 just above Consecration, 2 at the
		// top of the list, 3 just above Holy Strike. In the execute phase it beats the
		// Consecration tick or the Holy Strike it delays.
		{Name: "hammerOfWrathFirst", Default: 0, Min: 0, Max: 3, Step: 1},
		// Twist the seals for Unsealed Wrath: 0 never, 1 put the other damage seal up once the
		// seal's Echo holds twistStacks copies, and the seal back once the other's Echo holds
		// twistBack, within twistWindow tenths of a second of the next swing (0: as soon as the
		// copies are there), above twistMana percent of mana.
		{Name: "twist", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "twistStacks", Default: 3, Min: 1, Max: 3, Step: 1},
		{Name: "twistBack", Default: 1, Min: 1, Max: 3, Step: 1},
		{Name: "twistWindow", Default: 0, Min: 0, Max: 30, Step: 5},
		{Name: "twistMana", Default: 0, Min: 0, Max: 80, Step: 20},
		// Each swap costs a seal's mana, so when the mana won't last the fight the swap away waits
		// for all 3 copies instead of twistStacks: fewer swaps, more copies spent a swap. The mana
		// lasts while the bar covers twistThriftRate mana a second of the fight left. 0 off.
		{Name: "twistThriftRate", Default: 0, Min: 0, Max: 30, Step: 5},
		// Holy Strike only above this much mana, 0 always.
		{Name: "holyStrikeMana", Default: 0, Min: 0, Max: 40, Step: 20},
		// Holy Strike ahead of the seal and twist lines instead of after them.
		{Name: "holyStrikeFirst", Default: 0, Min: 0, Max: 1, Step: 1},
		// With twisting, judge Seal of Righteousness only when the next swing is more than this
		// many tenths of a second away, so most Judgements land on Seal of Command (its
		// judgement rolls on the melee table and hits harder). 0 judges whichever seal is up.
		{Name: "judgeRighteousnessWindow", Default: 0, Min: 0, Max: 30, Step: 5},
		// Swift Judgement (the 15 point Protection talent, a free Judgement of its own on a 1 min
		// cooldown since the hotfix) on cooldown, on the same seals as Judgement. Off the global
		// cooldown, so it costs no other cast.
		{Name: "swiftJudgement", Default: 1, Min: 0, Max: 1, Step: 1},
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
		// Without twisting: put the seal back up ahead of Consecration and Holy Strike, so a
		// seal that drops (after the Judgement of the Crusader opener, or after 30 sec) is
		// recast first.
		{Name: "sealFirst", Default: 0, Min: 0, Max: 1, Step: 1},
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
	holyWrath := rankID(paladin.HolyWrathLevel[:], paladin.HolyWrathSpellId[:])

	seal, otherSeal := sealOfRighteousness, sealOfCommand
	sealNotes := "Seal of Righteousness whenever it is down. Judgement does not consume it under Forever."
	if k["seal"] == 2 {
		seal, otherSeal = sealOfCommand, sealOfRighteousness
		sealNotes = "Seal of Command whenever it is down. Judgement does not consume it under Forever. Without the talent ignore the warning on this line."
	}

	var hammerOfWrathLine *proto.APLListItem
	if k["hammerOfWrathMana"] < 100 {
		var floor value
		if k["hammerOfWrathMana"] > 0 {
			floor = ge(currentManaPercent(), num(k["hammerOfWrathMana"]/100))
		}
		hammerOfWrathLine = cast(hammerOfWrath, floor, "Hammer of Wrath below 20% health.")
	}

	var items []*proto.APLListItem
	if hammerOfWrathLine != nil && k["hammerOfWrathFirst"] == 2 {
		items = append(items, hammerOfWrathLine)
	}
	if k["cooldownsFirst"] == 1 {
		items = append(items,
			trinketLine(k["trinketExecute"] == 1),
			cast(itemID(goblinSapperCharge), nil, "Goblin Sapper Charge on cooldown."),
		)
	}
	var jotcUp value
	dutySeal, dutyName, dutyJudgement := sealOfTheCrusader, "Seal of the Crusader", judgementOfTheCrusader
	if k["jotc"] == 2 {
		dutySeal, dutyName, dutyJudgement = rankID(paladin.SealOfWisdomLevel[:], paladin.SealOfWisdomSpellId[:]), "Seal of Wisdom", judgementOfWisdom
	}
	if k["jotc"] > 0 {
		jotcUp = targetAuraIsActive(dutyJudgement)
		items = append(items,
			cast(judgement, auraIsActive(dutySeal), "Judge "+dutyName+" first, the judgement lasts 40 sec and every melee hit refreshes it."),
			cast(dutySeal, and(not(jotcUp), not(auraIsActive(dutySeal))), dutyName+" when its judgement is not on the target."),
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
	twist := k["twist"] == 1
	sealLine := cast(seal, and(not(auraIsActive(seal)), jotcUp), sealNotes)
	if !twist && k["sealFirst"] == 1 {
		items = append(items, sealLine)
	}
	if hammerOfWrathLine != nil && k["hammerOfWrathFirst"] == 1 && consecrationLine != nil && k["consecrationFirst"] == 1 {
		items = append(items, hammerOfWrathLine)
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
	swiftLine := func(when value, notes string) {
		if k["swiftJudgement"] == 1 {
			items = append(items, cast(swiftJudgement, when, notes+" Without the talent ignore the warning on this line."))
		}
	}
	if !twist {
		swiftLine(and(auraIsActive(seal), jotcUp), "Swift Judgement on cooldown, off the global cooldown.")
		items = append(items, cast(judgement, and(auraIsActive(seal), jotcUp), "Judgement on cooldown."))
		if k["sealFirst"] == 0 {
			items = append(items, sealLine)
		}
	} else {
		sealUp := or(auraIsActive(seal), auraIsActive(otherSeal))
		echoOf := func(s *proto.ActionID) *proto.ActionID {
			if s == sealOfCommand {
				return echoOfCommand
			}
			return echoOfRighteousness
		}
		// The swap away from the seal waits for its copies, the swap back for the other's.
		ready := func(s *proto.ActionID, n float64) value {
			stacked := ge(auraNumStacks(echoOf(s)), num(n))
			if k["twistWindow"] > 0 {
				return and(stacked, lt(mainHandTimeToNext(), seconds(k["twistWindow"]/10)))
			}
			return stacked
		}
		var floor value
		if k["twistMana"] > 0 {
			floor = ge(currentManaPercent(), num(k["twistMana"]/100))
		}
		judgeWhen, judgeNotes := sealUp, "on whichever damage seal is up."
		if k["judgeRighteousnessWindow"] > 0 {
			judgeWhen = or(auraIsActive(sealOfCommand), and(auraIsActive(sealOfRighteousness), gt(mainHandTimeToNext(), seconds(k["judgeRighteousnessWindow"]/10))))
			judgeNotes = "on Seal of Command, and on Seal of Righteousness only when the next swing is a while off (Judgement of Command hits harder)."
		}
		swiftLine(and(judgeWhen, jotcUp), "Swift Judgement on cooldown, off the global cooldown, "+judgeNotes)
		stacks, back := k["twistStacks"], k["twistBack"]
		away := ready(seal, stacks)
		awayNotes := fmt.Sprintf("Unsealed Wrath: once the seal has stored %s, the other damage seal, so the next swing spends them all.", copies(stacks))
		if rate := k["twistThriftRate"]; rate > 0 && stacks < unsealedWrathCopies {
			lasts := ge(currentMana(), mul(remainingTime(), num(rate)))
			away = or(and(lasts, away), ready(seal, unsealedWrathCopies))
			awayNotes = fmt.Sprintf("Unsealed Wrath: once the seal has stored %s while the mana lasts the fight (%d mana a second left), and all 3 when it won't, the other damage seal, so the next swing spends them all.", copies(stacks), int(rate))
		}
		items = append(items,
			cast(judgement, and(judgeWhen, jotcUp), "Judgement on cooldown, "+judgeNotes),
			cast(seal, and(not(sealUp), jotcUp), sealNotes),
			cast(otherSeal, and(auraIsActive(seal), away, floor), awayNotes),
			cast(seal, and(auraIsActive(otherSeal), ready(otherSeal, back), floor),
				fmt.Sprintf("And back once the other seal has stored %s, which the next swing spends.", copies(back))),
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
	if hammerOfWrathLine != nil && k["hammerOfWrathFirst"] == 3 {
		items = append(items, hammerOfWrathLine)
	}
	if k["holyStrikeFirst"] == 0 {
		holyStrikeLine()
	}
	if hammerOfWrathLine != nil && (k["hammerOfWrathFirst"] == 0 || k["hammerOfWrathFirst"] == 1 && (consecrationLine == nil || k["consecrationFirst"] == 0)) {
		items = append(items, hammerOfWrathLine)
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
	if k["jotc"] > 0 {
		opener = dutySeal
	}
	rot.PrepullActions = append(rot.PrepullActions, prepull(opener, 1.5))
	return rot
}

// unsealedWrathCopies is the most copies Unsealed Wrath stores (sim/paladin/unsealed_wrath.go).
const unsealedWrathCopies = 3

// copies reads a copy count out for the notes.
func copies(n float64) string {
	if n == 1 {
		return "a copy"
	}
	return fmt.Sprintf("%d copies", int(n))
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
