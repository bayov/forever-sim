package main

import (
	"github.com/wowsims/classic/sim/core/proto"
)

// Shared pieces of the Arms and Fury templates (warrior_arms.go, warrior_fury.go). The spells are
// the level 60 ranks, the warrior templates only build level 60 rotations.
var (
	battleShout      = spellIDRank(11551, 6)
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
)

// The 30 sec Death Wish is the anchor of a warrior's burst window.
var deathWishDuration = seconds(30)

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
		{Name: "racialHoldForDw", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "potionHoldForDw", Default: 1, Min: 0, Max: 2, Step: 1},
		{Name: "potionAt", Default: 30, Min: 10, Max: 60, Step: 5},
		// The synthetic on-use trinket: 0 on cooldown (with the other cooldowns), otherwise its last
		// use waits for the last trinketAt seconds when the wait costs no use.
		{Name: "trinketAt", Default: 0, Min: 0, Max: 40, Step: 5},
		// Battle Shout when the raid's is not up. The shared raid already has one.
		{Name: "battleShout", Default: 0, Min: 0, Max: 1, Step: 1},
	}
}

// warriorCooldowns are the lines that go at the top of either rotation: the burst cooldowns, the
// racials and Rage on demand.
func warriorCooldowns(k Knobs) []*proto.APLListItem {
	var items []*proto.APLListItem
	if k["battleShout"] == 1 {
		items = append(items, cast(battleShout, not(auraIsActive(battleShout)), "Battle Shout whenever it is down."))
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
	hold := func(spell *proto.ActionID, on bool) value {
		if !on {
			return nil
		}
		return or(
			auraIsActive(deathWish),
			eq(spellUsesRemaining(deathWish, deathWishDuration), num(0)),
			gt(spellUsesLostByDelay(spell, spellTimeToReady(deathWish), buffDuration), num(0)),
		)
	}
	racial := k["racialHoldForDw"] == 1
	items = append(items,
		cast(bloodFury, hold(bloodFury, racial), "Orc. Blood Fury waits for Death Wish when the wait costs no use. Other races can ignore the warning on this line."),
		cast(berserking, hold(berserking, racial), "Troll. Berserking waits for Death Wish when the wait costs no use. Other races can ignore the warning on this line."),
		cast(elunesLight, hold(elunesLight, racial), "Night Elf. Elune's Light waits for Death Wish when the wait costs no use. Other races can ignore the warning on this line."),
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
	items = append(items, autocastOtherCooldowns(nil, "Anything not listed above (Sapper, trinkets)."))
	return items
}

// potionLine is the Mighty Rage Potion: on cooldown, held for Death Wish, or held for the last
// potionAt seconds, in each case only when the wait costs no use.
func potionLine(k Knobs, hold func(*proto.ActionID, bool) value) *proto.APLListItem {
	if k["potionHoldForDw"] == 2 {
		return cast(potion, or(
			le(remainingTime(), seconds(k["potionAt"])),
			gt(spellUsesLostByDelay(potion, sub(remainingTime(), seconds(k["potionAt"])), seconds(15)), num(0)),
		), "Mighty Rage Potion, the last one held for the end of the fight when the wait costs no use.")
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
