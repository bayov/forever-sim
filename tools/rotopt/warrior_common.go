package main

import (
	"github.com/wowsims/classic/sim/core/proto"
)

// Shared pieces of the Arms and Fury templates (warrior_arms.go, warrior_fury.go). The spells are
// the level 60 ranks, the warrior templates only build level 60 rotations.
var (
	battleShout      = spellIDRank(25289, 7)
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
	heroicStrikeNext = &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: 25286}, Tag: 1, Rank: 9}
	potion           = &proto.ActionID{RawId: &proto.ActionID_OtherId{OtherId: proto.OtherAction_OtherActionPotion}}
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
		// the wait costs no use.
		{Name: "racialHoldForDw", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "potionHoldForDw", Default: 1, Min: 0, Max: 1, Step: 1},
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
		cast(potion, hold(potion, k["potionHoldForDw"] == 1), "Mighty Rage Potion with Death Wish."),
		cast(bloodrage, lt(currentRage(), num(k["bloodrageRage"])), "Bloodrage when there is room for the Rage."),
	)
	if k["berserkerRageRage"] > 0 {
		items = append(items, cast(berserkerRage, lt(currentRage(), num(k["berserkerRageRage"])),
			"Berserker Rage for its Rage (Improved Berserker Rage). It needs Berserker Stance."))
	}
	items = append(items, autocastOtherCooldowns(nil, "Anything not listed above (Sapper, trinkets)."))
	return items
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
