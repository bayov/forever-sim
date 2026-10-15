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
		// Recklessness (30 min cooldown, so once a fight): 0 never, 1 at the pull, 2 when this many
		// seconds are left (recklessAt).
		{Name: "recklessness", Default: 2, Min: 0, Max: 2, Step: 1},
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
		// Blood Fury, Berserking, Elune's Light and the Mighty Rage Potion wait for Death Wish when
		// the wait costs no use. The potion can instead (2) wait for the last potionAt seconds when
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
		items = append(items, cast(deathWish, nil, "Death Wish on cooldown. Without the talent ignore the warning on this line."))
	case 2:
		items = append(items, cast(deathWish, or(
			le(remainingTime(), seconds(k["dwEnd"])),
			gt(spellUsesLostByDelay(deathWish, sub(remainingTime(), seconds(k["dwEnd"])), deathWishDuration), num(0)),
		), "Death Wish on cooldown, but held for the end of the fight when the wait costs no use. Without the talent ignore the warning on this line."))
	}
	switch k["recklessness"] {
	case 1:
		items = append(items, cast(recklessness, nil, "Recklessness at the pull."))
	case 2:
		items = append(items, cast(recklessness, le(remainingTime(), seconds(k["recklessAt"])), "Recklessness for the last seconds of the fight, in the execute phase."))
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
		cast(bloodFury, racialHold(bloodFury), "Orc. Blood Fury waits for Death Wish (or the end of the fight) when the wait costs no use. Other races can ignore the warning on this line."),
		cast(berserking, racialHold(berserking), "Troll. Berserking waits for Death Wish (or the end of the fight) when the wait costs no use. Other races can ignore the warning on this line."),
		// Elune's Light keeps waiting for Death Wish: held for the end it lost 7 at 180 sec on
		// two-hand Fury, where Blood Fury and Berserking gained.
		cast(elunesLight, hold(elunesLight, k["racialHoldForDw"] >= 1), "Night Elf. Elune's Light waits for Death Wish when the wait costs no use. Other races can ignore the warning on this line."),
		potionLine(k, hold),
	)
	if k["trinketAt"] > 0 {
		items = append(items, cast(battleTrinket, or(
			le(remainingTime(), seconds(k["trinketAt"])),
			gt(spellUsesLostByDelay(battleTrinket, sub(remainingTime(), seconds(k["trinketAt"])), seconds(15)), num(0)),
		), "The on-use trinket, the last use held for the end of the fight when the wait costs no use."))
	}
	items = append(items,
		cast(bloodrage, lt(currentRage(), num(k["bloodrageRage"])), "Bloodrage when there is room for the Rage."),
	)
	if k["berserkerRageRage"] > 0 {
		items = append(items, cast(berserkerRage, lt(currentRage(), num(k["berserkerRageRage"])),
			"Berserker Rage for its Rage (Improved Berserker Rage). It needs Berserker Stance."))
	}
	if k["mineSync"] == 1 {
		items = append(items,
			cast(warriorSapper, nil, "Goblin Sapper Charge on cooldown."),
			cast(warriorLandMine, ge(mainHandTimeToNext(), sub(mainHandSwingTime(), seconds(0.4))),
				"Goblin Land Mine right after a main hand swing, so its 1 sec cast loses no swing time."),
		)
	}
	items = append(items, autocastOtherCooldowns(nil, "Anything not listed above (Sapper, trinkets)."))
	return items
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
		Notes: "Sunder Armor as the first GCD after the pull, as every warrior does; the raid's tanks keep it up after.",
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
		notes := "Mighty Rage Potion, the last one held for the end of the fight when the wait costs no use."
		if k["potionPull"] == 1 {
			notes = "Mighty Rage Potion at the pull, the last one held for the end of the fight when the wait costs no use."
		}
		return cast(potion, when, notes)
	}
	return cast(potion, hold(potion, k["potionHoldForDw"] == 1), "Mighty Rage Potion with Death Wish.")
}

// heroicStrikeLine queues Heroic Strike with Rage to spare.
func heroicStrikeLine(k Knobs) *proto.APLListItem {
	when := ge(currentRage(), num(k["hsRage"]))
	if k["hsExecute"] == 0 {
		when = and(when, not(executePhase()))
	}
	return cast(heroicStrikeNext, when, "Heroic Strike on the next swing with Rage to spare.")
}

func init() {
	templates["warrior_fury"] = warriorFury{}
	templates["warrior_arms"] = warriorArms{}
}
