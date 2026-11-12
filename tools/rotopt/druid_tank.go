package main

import (
	"fmt"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/druid"
)

// The Feral tank (sim/druid/bear.go, bear_attacks.go; docs/tank_druid.md), searched with -tank on
// the scoring encounters. What a bear decides: what its Rage goes to (Primal Bite, Lacerate, Maul,
// Swipe, Demoralizing Roar), when Maul is queued and how much Rage it leaves for Primal Bite, whether
// Lacerate is kept at 5 stacks, Enrage before the pull or in the fight (its armor cut), a shift
// before the pull for Furor's Rage, and Barkskin and Frenzied Regeneration on a timer (a scoring run
// reads no health, docs/tank_sim.md). The bear keeps its own Demoralizing Roar up
// (tankscore.OwnDebuffs drops the raid's Roar and Shout).
type druidTank struct{}

func init() {
	templates["druid_tank"] = druidTank{}
}

var (
	bearForm          = spellID(9634)
	bearDemoRoar      = spellID(9898)
	bearPrimalBite    = spellID(1238070)
	bearLacerate      = spellID(1235827)
	bearMaul          = spellID(9881)
	bearSwipe         = spellID(9908)
	bearEnrage        = spellID(5229)
	bearBarkskin      = spellID(22812)
	bearFrenziedRegen = spellID(22842)
)

func (druidTank) Knobs() []Knob {
	return []Knob{
		// Demoralizing Roar again once it has less than roarRefresh seconds left on the boss.
		{Name: "roarRefresh", Default: 3, Min: 0, Max: 9, Step: 1},
		// Demoralizing Roar's line just below Lacerate's keep-up line, 0/1 (0: ahead of the attacks,
		// where it takes the pull's first GCD and Rage; tank_rotations-20261104-1).
		{Name: "roarAfterLacerate", Default: 1, Min: 0, Max: 1, Step: 1},
		// Primal Bite on cooldown, 0/1.
		{Name: "primalBite", Default: 1, Min: 0, Max: 1, Step: 1},
		// Lacerate: 0 never, 1 build it to 5 stacks and keep them up (again under lacerateRefresh
		// seconds left), 2 also as the filler whenever the GCD is free.
		{Name: "lacerate", Default: 2, Min: 0, Max: 2, Step: 1},
		{Name: "lacerateRefresh", Default: 4, Min: 1, Max: 9, Step: 1},
		// Lacerate's keep-up line ahead of Primal Bite, 0/1.
		{Name: "lacerateFirst", Default: 0, Min: 0, Max: 1, Step: 1},
		// Swipe as the filler (single target) at swipeRage Rage or more, 0 never. It goes after
		// Lacerate's lines.
		{Name: "swipeRage", Default: 0, Min: 0, Max: 80, Step: 10},
		// Maul queued at maulRage Rage or more.
		{Name: "maulRage", Default: 30, Min: 10, Max: 80, Step: 5},
		// While Primal Bite is ready within a GCD, Maul only with its cost on top of maulRage, 0/1.
		{Name: "maulReserve", Default: 0, Min: 0, Max: 1, Step: 1},
		// Enrage: 0 never, 1 before the pull (enragePrepull seconds), 2 in the fight under enrageRage
		// Rage, 3 both.
		{Name: "enrage", Default: 0, Min: 0, Max: 3, Step: 1},
		{Name: "enragePrepull", Default: 10, Min: 1, Max: 10, Step: 1},
		{Name: "enrageRage", Default: 20, Min: 0, Max: 100, Step: 10},
		// A shift into Dire Bear Form before the pull, for Furor's 10 Rage, 0/1 (with Enrage before the
		// pull it comes just before it, else 1.5 sec before the pull).
		{Name: "prepullShift", Default: 0, Min: 0, Max: 1, Step: 1},
		// Barkskin on cooldown, 20% less Physical damage for 15 sec of each minute, for a GCD: 0 never,
		// 1 ahead of the attacks, 2 in the filler's place (after Primal Bite and Lacerate's keep-up), from
		// barkskinAfter seconds into the fight.
		{Name: "barkskin", Default: 0, Min: 0, Max: 2, Step: 1},
		{Name: "barkskinAfter", Default: 0, Min: 0, Max: 30, Step: 5},
		// Frenzied Regeneration on cooldown once the Rage is at frRage or more, 0 never.
		// Step 10: 70 beat 60, 80 and 100 at Max survival (tank_rotations-20261112-1).
		{Name: "frRage", Default: 0, Min: 0, Max: 100, Step: 10},
		// The synthetic armor trinket (a defensive on-use, standards 36) on cooldown, 0/1: the autocast
		// line leaves defensive cooldowns to a health threshold, which a scoring run can't read.
		{Name: "armorTrinket", Default: 0, Min: 0, Max: 1, Step: 1},
		// 1 drinks the potion slot's potion (the Greater Stoneshield) on cooldown from its own line, above the
		// autocast one: autocast never drinks a defensive potion in a tank score (tank_fairness-20261104-10).
		{Name: "potion", Default: 0, Min: 0, Max: 1, Step: 1},
	}
}

func (druidTank) Build(k Knobs) *proto.APLRotation {
	var items []*proto.APLListItem
	if k["potion"] == 1 {
		items = append(items, cast(potion, nil, "Greater Stoneshield Potion on cooldown."))
	}
	items = append(items, autocastOtherCooldowns(nil, "Racials, trinkets and potions."))
	if k["armorTrinket"] == 1 {
		items = append(items, cast(itemID(druid.FeralTankSyntheticArmorTrinket), nil,
			"The armor trinket on cooldown (on a timer: a scoring run reads no health). Without it ignore the warning on this line."))
	}
	var barkskin *proto.APLListItem
	if k["barkskin"] > 0 {
		var cond value
		note := "Barkskin on cooldown: 20% less Physical damage for 15 sec of each minute (on a timer: a scoring run reads no health)."
		if k["barkskinAfter"] > 0 {
			cond = gt(warriorCurrentTime(), seconds(k["barkskinAfter"]))
			note = fmt.Sprintf("Barkskin on cooldown from %.0f sec in: 20%% less Physical damage for 15 sec of each minute (on a timer: a scoring run reads no health).", k["barkskinAfter"])
		}
		barkskin = cast(bearBarkskin, cond, note)
	}
	if k["barkskin"] == 1 {
		items = append(items, barkskin)
	}
	if k["frRage"] > 0 {
		note := fmt.Sprintf("Frenzied Regeneration on cooldown at %.0f Rage or more (on a timer: a scoring run reads no health).", k["frRage"])
		if k["frRage"] >= 100 {
			note = "Frenzied Regeneration at 100 Rage: it fires at the pull's Rage peak, about once a fight (a scoring run reads no health)."
		}
		items = append(items, cast(bearFrenziedRegen, ge(currentRage(), num(k["frRage"])), note))
	}
	if k["enrage"] >= 2 {
		items = append(items, cast(bearEnrage, lt(currentRage(), num(k["enrageRage"])),
			fmt.Sprintf("Enrage in the fight under %.0f Rage (off the GCD; 16%% less armor from items for 10 sec).", k["enrageRage"])))
	}
	roar := cast(bearDemoRoar, lt(targetAuraRemainingTime(bearDemoRoar), seconds(k["roarRefresh"])),
		"Demoralizing Roar kept up on the boss: the tank applies its own (the bench drops the raid's Roar and Shout).")
	if k["roarAfterLacerate"] == 0 {
		items = append(items, roar)
	}

	keepLacerate := cast(bearLacerate, or(lt(targetAuraNumStacks(bearLacerate), num(5)), lt(dotRemainingTime(bearLacerate), seconds(k["lacerateRefresh"]))),
		fmt.Sprintf("Lacerate to 5 stacks, and again before they drop (under %.0f sec left).", k["lacerateRefresh"]))
	if k["lacerate"] >= 1 && k["lacerateFirst"] == 1 {
		items = append(items, keepLacerate)
	}
	if k["primalBite"] == 1 {
		items = append(items, cast(bearPrimalBite, nil, "Primal Bite on cooldown."))
	}
	if k["lacerate"] >= 1 && k["lacerateFirst"] == 0 {
		items = append(items, keepLacerate)
	}
	if k["roarAfterLacerate"] == 1 {
		items = append(items, roar)
	}
	if k["barkskin"] == 2 {
		items = append(items, barkskin)
	}
	if k["lacerate"] == 2 {
		items = append(items, cast(bearLacerate, nil, "Lacerate as the filler: it refreshes the bleed and hits for 50% of weapon damage at 5 stacks."))
	}
	if k["swipeRage"] > 0 {
		items = append(items, cast(bearSwipe, ge(currentRage(), num(k["swipeRage"])), fmt.Sprintf("Swipe as the filler at %.0f Rage or more.", k["swipeRage"])))
	}
	maulCond := ge(currentRage(), num(k["maulRage"]))
	maulNote := fmt.Sprintf("Maul queued at %.0f Rage or more.", k["maulRage"])
	if k["maulReserve"] == 1 && k["primalBite"] == 1 {
		maulCond = and(maulCond, or(gt(spellTimeToReady(bearPrimalBite), seconds(1.5)),
			ge(currentRage(), num(k["maulRage"]+15))))
		maulNote = fmt.Sprintf("Maul queued at %.0f Rage or more, with Primal Bite's Rage on top while it's due.", k["maulRage"])
	}
	items = append(items, cast(bearMaul, maulCond, maulNote))

	rot := rotation(items...)
	enrageBefore := k["enrage"] == 1 || k["enrage"] == 3
	if k["prepullShift"] == 1 {
		at := 1.5
		if enrageBefore {
			at = k["enragePrepull"] + 1.5
		}
		rot.PrepullActions = append(rot.PrepullActions, prepull(bearForm, at))
	}
	if enrageBefore {
		rot.PrepullActions = append(rot.PrepullActions, prepull(bearEnrage, k["enragePrepull"]))
	}
	return rot
}
