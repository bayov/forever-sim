package main

import (
	goproto "google.golang.org/protobuf/proto"

	"github.com/wowsims/classic/sim/core/proto"
)

// Shared pieces of the Arms and Fury templates (warrior_arms.go, warrior_fury.go). The spells are
// the level 60 ranks, the warrior templates only build level 60 rotations.
var (
	battleShout      = spellIDRank(11551, 6)
	sunderArmor      = spellID(11597)
	charge           = spellIDRank(11578, 3)
	battleStance     = spellID(2457)
	berserkerStance  = spellID(2458)
	bloodrage        = spellID(2687)
	berserkerRage    = spellID(18499)
	deathWish        = spellID(12328)
	recklessness     = spellID(1719)
	bloodthirst      = spellIDRank(23894, 4)
	mortalStrike     = spellIDRank(21553, 4)
	whirlwind        = spellID(1680)
	slam             = spellIDRank(11605, 5)
	execute          = spellIDRank(20662, 5)
	overpower        = spellIDRank(11585, 4)
	rend             = spellIDRank(11574, 7)
	hamstring        = spellIDRank(7373, 3)
	spearingStrike   = spellID(1310222)
	heroicStrikeNext = &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: 11567}, Tag: 1, Rank: 8}
	potion           = &proto.ActionID{RawId: &proto.ActionID_OtherId{OtherId: proto.OtherAction_OtherActionPotion}}
	// The synthetic on-use trinket (sim/warrior/forever_synthetic_trinkets.go): 20 sec, 2 min.
	battleTrinket = itemID(991018)
	// The Goblin Sapper Charge and the Goblin Land Mine share a 1 min explosives cooldown.
	warriorSapper   = itemID(10646)
	warriorLandMine = itemID(4395)
)

// The 30 sec Death Wish is the anchor of a warrior's burst window.
var deathWishDuration = seconds(30)

// warriorMax is the largest of the values.
func warriorMax(vals ...value) value {
	return &proto.APLValue{Value: &proto.APLValue_Max{Max: &proto.APLValueMax{Vals: vals}}}
}

func warriorCurrentTime() value {
	return &proto.APLValue{Value: &proto.APLValue_CurrentTime{CurrentTime: &proto.APLValueCurrentTime{}}}
}

func currentRage() value {
	return &proto.APLValue{Value: &proto.APLValue_CurrentRage{CurrentRage: &proto.APLValueCurrentRage{}}}
}

func executePhase() value {
	return &proto.APLValue{Value: &proto.APLValue_IsExecutePhase{IsExecutePhase: &proto.APLValueIsExecutePhase{
		Threshold: proto.APLValueIsExecutePhase_E20,
	}}}
}

// warriorKnobs are the knobs both specs share.
func warriorKnobs() []Knob {
	return []Knob{
		// Death Wish: 0 never, 1 on cooldown, 2 on cooldown but held for the last dwEnd seconds (the
		// execute phase) when the wait costs no use.
		{Name: "deathWish", Default: 2, Min: 0, Max: 2, Step: 1},
		{Name: "dwEnd", Default: 36, Min: 15, Max: 60, Step: 3},
		// Recklessness (30 min cooldown, so once a fight): 0 no line of its own (the other cooldowns
		// line casts it at the pull), 1 at the pull, 2 when this many seconds are left (recklessAt),
		// 3 while the fight's last Death Wish is up (the end-of-fight dump: with two Death Wishes
		// the first goes at the pull without it), just ahead of the other cooldowns line, so that
		// line's trinket waits for the GCD after it instead of going out in the same pass (the
		// rotation auditor's pass 23).
		{Name: "recklessness", Default: 2, Min: 0, Max: 3, Step: 1},
		{Name: "recklessAt", Default: 15, Min: 10, Max: 30, Step: 5},
		// Bloodrage below this much Rage.
		{Name: "bloodrageRage", Default: 70, Min: 0, Max: 100, Step: 10},
		// Berserker Rage below this much Rage (it only gives Rage with the talent), 0 never.
		{Name: "berserkerRageRage", Default: 70, Min: 0, Max: 100, Step: 10},
		// Heroic Strike on the next swing above this much Rage, and 0/1 to keep queueing it in the
		// execute phase.
		{Name: "hsRage", Default: 50, Min: 20, Max: 90, Step: 10},
		{Name: "hsExecute", Default: 0, Min: 0, Max: 1, Step: 1},
		// Execute ahead of the main strike in the execute phase (1) or after it (0).
		{Name: "executeFirst", Default: 1, Min: 0, Max: 1, Step: 1},
		// Blood Fury, Berserking and the Mighty Rage Potion wait for Death Wish when the wait costs
		// no use. (Elune's Light has its own rule, elunesLightWhen: never with Recklessness.) The potion can instead (2) wait for the last potionAt seconds when
		// the wait costs no use, for Arms' Recklessness or a Death Wish that is held to the end.
		{Name: "racialHoldForDw", Default: 1, Min: 0, Max: 2, Step: 1},
		{Name: "potionHoldForDw", Default: 1, Min: 0, Max: 2, Step: 1},
		{Name: "potionAt", Default: 30, Min: 10, Max: 60, Step: 5},
		// With potionHoldForDw 2: 1 also drinks at the pull (the first 3 sec), which matters in
		// fights too short for the hold to lose a use.
		{Name: "potionPull", Default: 0, Min: 0, Max: 1, Step: 1},
		// racialHoldForDw 2: Blood Fury and Berserking wait for the last racialAt seconds when the
		// wait costs no use, instead of waiting for Death Wish.
		{Name: "racialAt", Default: 25, Min: 15, Max: 40, Step: 5},
		// The synthetic on-use trinket: 0 on cooldown (with the other cooldowns), otherwise its last
		// use waits for the last trinketAt seconds when the wait costs no use.
		{Name: "trinketAt", Default: 0, Min: 0, Max: 40, Step: 5},
		// 1: the Goblin Sapper Charge on cooldown, and the Goblin Land Mine right after a main hand
		// swing, since its 1 sec cast restarts both swing timers. 0: both with the other cooldowns.
		{Name: "mineSync", Default: 1, Min: 0, Max: 1, Step: 1},
		// The Battle Shout duty (standards 23): 1 shouts before the pull, with a Bloodrage for the
		// Rage, and again before it drops when the fight outlasts it. The base presets leave it to
		// the raid (0), and the duty presets switch the raid's off.
		{Name: "battleShout", Default: 0, Min: 0, Max: 1, Step: 1},
		// Refresh Battle Shout when this many seconds of it are left. 35 puts the refresh (only in
		// fights longer than about 177 sec) at about 142 sec, ahead of the execute phase, rather than
		// at about 174 sec inside it.
		{Name: "shoutRefresh", Default: 35, Min: 0, Max: 60, Step: 5},
	}
}

// warriorCooldowns are the lines that go at the top of either rotation: the burst cooldowns, the
// racials and Rage on demand.
func warriorCooldowns(k Knobs) []*proto.APLListItem {
	var items []*proto.APLListItem
	if k["battleShout"] == 1 {
		items = append(items, cast(battleShout, and(
			lt(auraRemainingTime(battleShout), seconds(k["shoutRefresh"])),
			gt(remainingTime(), auraRemainingTime(battleShout)),
		), "Battle Shout before it drops, unless it lasts the fight."))
	}
	switch k["deathWish"] {
	case 1:
		items = append(items, cast(deathWish, nil, "Without the talent ignore the warning on this line."))
	case 2:
		items = append(items, cast(deathWish, or(
			le(remainingTime(), seconds(k["dwEnd"])),
			gt(spellUsesLostByDelay(deathWish, sub(remainingTime(), seconds(k["dwEnd"])), deathWishDuration), num(0)),
		), "Without the talent ignore the warning on this line."))
	}
	switch k["recklessness"] {
	case 1:
		items = append(items, cast(recklessness, nil, "Recklessness at the pull."))
	case 2:
		items = append(items, cast(recklessness, le(remainingTime(), seconds(k["recklessAt"])), ""))
	}
	// The wait for Death Wish: until it is ready, and when it is held for the last dwEnd seconds,
	// until then too.
	dwWait := spellTimeToReady(deathWish)
	if k["deathWish"] == 2 {
		dwWait = warriorMax(dwWait, sub(remainingTime(), seconds(k["dwEnd"])))
	}
	hold := func(spell *proto.ActionID, on bool) value {
		if !on {
			return nil
		}
		return or(
			auraIsActive(deathWish),
			eq(spellUsesRemaining(deathWish, deathWishDuration), num(0)),
			gt(spellUsesLostByDelay(spell, dwWait, buffDuration), num(0)),
		)
	}
	racial := k["racialHoldForDw"] == 1
	racialHold := func(spell *proto.ActionID) value {
		if k["racialHoldForDw"] == 2 {
			return or(
				le(remainingTime(), seconds(k["racialAt"])),
				gt(spellUsesLostByDelay(spell, sub(remainingTime(), seconds(k["racialAt"])), seconds(15)), num(0)),
			)
		}
		return hold(spell, racial)
	}
	items = append(items,
		cast(bloodFury, racialHold(bloodFury), "Other races can ignore the warning on this line."),
		cast(berserking, racialHold(berserking), "Other races can ignore the warning on this line."),
		cast(elunesLight, elunesLightWhen(k), "Elune's Light with the cooldowns, never during Recklessness, whose 100% crit makes its 10% worthless. Other races can ignore the warning on this line."),
		potionLine(k, hold),
	)
	if k["trinketAt"] > 0 {
		items = append(items, cast(battleTrinket, or(
			le(remainingTime(), seconds(k["trinketAt"])),
			gt(spellUsesLostByDelay(battleTrinket, sub(remainingTime(), seconds(k["trinketAt"])), seconds(15)), num(0)),
		), ""))
	}
	items = append(items,
		cast(bloodrage, lt(currentRage(), num(k["bloodrageRage"])), ""),
	)
	if k["berserkerRageRage"] > 0 {
		items = append(items, cast(berserkerRage, lt(currentRage(), num(k["berserkerRageRage"])),
			"Berserker Rage for its Rage (Improved Berserker Rage). It needs Berserker Stance."))
	}
	if k["mineSync"] == 1 {
		items = append(items,
			cast(warriorSapper, nil, ""),
			cast(warriorLandMine, ge(mainHandTimeToNext(), sub(mainHandSwingTime(), seconds(0.4))),
				"Its 1s cast then loses no swing time."),
		)
	}
	if k["recklessness"] == 3 {
		// Recklessness is the end-of-fight dump with the fight's last Death Wish: in a fight where
		// two Death Wishes fit, the first goes at the pull and Recklessness waits for the second.
		items = append(items, cast(recklessness, and(auraIsActive(deathWish), lastDeathWish()),
			"Recklessness with the fight's last Death Wish."))
	}
	items = append(items, autocastOtherCooldowns(nil, "Sapper and trinkets."))
	return items
}

// lastDeathWish is true when no Death Wish fits after the one that is up or about to go: the
// count is of the casts after the current cooldown that still have its 30 sec to run.
func lastDeathWish() value {
	return eq(spellUsesRemaining(deathWish, deathWishDuration), num(0))
}

// reckDue is true when Recklessness goes out within the next `within`, read from the same knobs
// as its own line (warriorCooldowns). Its 30 min cooldown makes it once a fight, and the caller
// checks whether it is already spent.
//   - recklessness 0 or 1: at the pull, so it is due until it is out.
//   - 2: in the last recklessAt seconds.
//   - 3: with the fight's last Death Wish, so when that Death Wish is up, or when the next Death
//     Wish is the last one and goes within `within`. That next Death Wish goes once it is ready,
//     and with deathWish=2 not before the last dwEnd seconds either: holding the last one costs no
//     use. (That wait is only "when Death Wish goes" for the last one: an earlier one goes as soon
//     as holding it would cost a use, so it is no estimate of when the next Death Wish goes.)
//
// nil means Recklessness never goes (recklessness=3 without Death Wish).
func reckDue(k Knobs, within float64) value {
	switch k["recklessness"] {
	case 2:
		return le(remainingTime(), seconds(k["recklessAt"]+within))
	case 3:
		if k["deathWish"] == 0 {
			return nil
		}
		next := spellTimeToReady(deathWish)
		if k["deathWish"] == 2 {
			next = warriorMax(next, sub(remainingTime(), seconds(k["dwEnd"])))
		}
		return or(
			and(auraIsActive(deathWish), lastDeathWish()),
			and(le(spellUsesRemaining(deathWish, deathWishDuration), num(1)), le(next, seconds(within))),
		)
	}
	return constant("true")
}

// elunesLightWhen is Night Elf Elune's Light (15 sec of 10% crit, 3 min) for warriors: with the
// cooldowns, never during Recklessness and not when Recklessness is due within its 15 sec, since
// Recklessness's 100% crit makes its 10% worthless. Otherwise it goes on cooldown, which puts the
// first at the pull with the potion, the trinket and, in fights where two fit, the first Death
// Wish. (Held for Death Wish instead, it went out with Death Wish and Recklessness at the end.)
func elunesLightWhen(k Knobs) value {
	due := reckDue(k, 15)
	if due == nil {
		return nil
	}
	return and(
		not(auraIsActive(recklessness)),
		or(eq(spellUsesRemaining(recklessness, seconds(0)), num(0)), not(due)),
	)
}

// The first GCD after the pull is one Sunder Armor for every warrior (standards 23), so every
// spell on the GCD waits for the sunderOpener sequence.
const sunderOpener = "sunder"

func sequenceIsComplete(name string) value {
	return &proto.APLValue{Value: &proto.APLValue_SequenceIsComplete{SequenceIsComplete: &proto.APLValueSequenceIsComplete{SequenceName: name}}}
}

// warriorGCDSpells are the spells the templates cast that take the GCD (or, the Land Mine, a cast).
var warriorGCDSpells = []*proto.ActionID{battleShout, deathWish, recklessness, berserkerRage, bloodthirst,
	mortalStrike, whirlwind, slam, execute, overpower, rend, hamstring, spearingStrike, warriorLandMine}

// warriorOpener puts the Sunder Armor sequence ahead of the first spell on the GCD and makes
// every such spell wait for it. Prepull actions don't count: Battle Shout goes up before the pull.
func warriorOpener(rot *proto.APLRotation) *proto.APLRotation {
	gcd := func(id *proto.ActionID) bool {
		for _, s := range warriorGCDSpells {
			if goproto.Equal(s, id) {
				return true
			}
		}
		return false
	}
	done := sequenceIsComplete(sunderOpener)
	first := -1
	for i, item := range rot.PriorityList {
		cs := item.Action.GetCastSpell()
		if cs == nil || !gcd(cs.SpellId) {
			continue
		}
		if first < 0 {
			first = i
		}
		item.Action.Condition = and(item.Action.Condition, done)
	}
	if first < 0 {
		first = len(rot.PriorityList)
	}
	sunder := &proto.APLListItem{
		Action: &proto.APLAction{Action: &proto.APLAction_Sequence{Sequence: &proto.APLActionSequence{
			Name:    sunderOpener,
			Actions: []*proto.APLAction{{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: sunderArmor}}}},
		}}},
		Notes: "Sunder Armor once at the pull. The tanks keep it up after.",
	}
	rot.PriorityList = append(rot.PriorityList[:first], append([]*proto.APLListItem{sunder}, rot.PriorityList[first:]...)...)
	return rot
}

// warriorShoutPrepull is the Battle Shout duty's pull: Bloodrage for the Rage, then the shout, early
// enough that its GCD is over before Arms' stance and the pull.
func warriorShoutPrepull(k Knobs) []*proto.APLPrepullAction {
	if k["battleShout"] != 1 {
		return nil
	}
	return []*proto.APLPrepullAction{prepull(bloodrage, 3), prepull(battleShout, 3)}
}

// warriorChargePrepull opens with Charge from Battle Stance (15 Rage, out of combat only, so before
// the pull; standards 13). Fury goes back to Berserker Stance after it, and Improved Tactical Mastery
// keeps the Rage.
func warriorChargePrepull(stance *proto.ActionID) []*proto.APLPrepullAction {
	if stance == battleStance {
		return []*proto.APLPrepullAction{prepull(battleStance, 1), prepull(charge, 0.5)}
	}
	return []*proto.APLPrepullAction{prepull(battleStance, 2.5), prepull(charge, 1.5), prepull(stance, 0.5)}
}

// potionLine is the Mighty Rage Potion: on cooldown, held for Death Wish, or held for the last
// potionAt seconds, in each case only when the wait costs no use.
func potionLine(k Knobs, hold func(*proto.ActionID, bool) value) *proto.APLListItem {
	if k["potionHoldForDw"] == 2 {
		when := or(
			le(remainingTime(), seconds(k["potionAt"])),
			gt(spellUsesLostByDelay(potion, sub(remainingTime(), seconds(k["potionAt"])), seconds(15)), num(0)),
		)
		if k["potionPull"] == 1 {
			when = or(le(warriorCurrentTime(), seconds(3)), when)
		}
		return cast(potion, when, "")
	}
	return cast(potion, hold(potion, k["potionHoldForDw"] == 1), "Mighty Rage Potion with Death Wish.")
}

// heroicStrikeLine queues Heroic Strike with Rage to spare.
func heroicStrikeLine(k Knobs) *proto.APLListItem {
	when := ge(currentRage(), num(k["hsRage"]))
	if k["hsExecute"] == 0 {
		when = and(when, not(executePhase()))
	}
	return cast(heroicStrikeNext, when, "")
}

func init() {
	templates["warrior_fury"] = warriorFury{}
	templates["warrior_arms"] = warriorArms{}
}
