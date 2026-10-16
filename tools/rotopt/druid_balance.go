package main

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/druid"
)

// Balance under Forever rules. The moonkin keeps its dots up and fills with Starfire or
// Wrath. Eclipse (the Balance talent) gives each Wrath that lands 2 charges, at most 4, and
// each charge takes up to 0.5 sec off a Starfire, so the filler can weave Wrath to keep
// Starfire short. Clearcasting is not used up by Wrath, so a free Starfire can wait for it.
// Every line names the highest rank the player's level knows.
type druidBalance struct{}

func init() {
	templates["druid_balance"] = druidBalance{}
}

var (
	druidClearcasting = spellID(16870)
	druidEclipse      = spellID(408255)
	druidInnervate    = spellID(29166)
	druidFaerieFire   = spellID(9907)
	// The raid's Curse of Recklessness on the target (sim/core/debuffs.go).
	druidCoRAura = spellID(11717)
)

func (druidBalance) Knobs() []Knob {
	return []Knob{
		// Keep Moonfire and Insect Swarm up (Insect Swarm needs its talent), 0/1 each.
		{Name: "moonfire", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "insectSwarm", Default: 1, Min: 0, Max: 1, Step: 1},
		// No new dot with less than this many seconds of the fight left.
		{Name: "dotMinTime", Default: 8, Min: 0, Max: 12, Step: 2},
		// The filler: 0 Starfire, 1 Wrath, 2 Wrath until Eclipse has eclipseStacks charges
		// and then Starfire, 3 Wrath only once Eclipse is down to eclipseStacks charges or
		// fewer with no Wrath on its way, and Starfire otherwise. A Wrath flies as long as it
		// casts (1.5 sec), so with 2 a second Wrath goes out before the first one's charges
		// land, and a third one's charges hit the cap of 4.
		{Name: "filler", Default: 3, Min: 0, Max: 3, Step: 1},
		{Name: "eclipseStacks", Default: 2, Min: 0, Max: 4, Step: 1},
		// Count a Wrath still flying to the target as the charges it will give, so the filler
		// does not cast a second Wrath while the first is in the air, 0/1.
		{Name: "wrathInFlight", Default: 0, Min: 0, Max: 1, Step: 1},
		// Starfire while Clearcasting is up, whatever the filler says, 0/1.
		{Name: "clearcastStarfire", Default: 1, Min: 0, Max: 1, Step: 1},
		// Below this much mana the dots stop and the filler is Wrath, which costs a third of
		// Starfire for each second of casting. 0 never.
		{Name: "wrathBelowMana", Default: 0, Min: 0, Max: 40, Step: 10},
		// Innervate on ourself below this much mana, 0 never.
		{Name: "innervateMana", Default: 40, Min: 0, Max: 80, Step: 10},
		// A pre-pull cast timed to land at 0: 0 none, 1 Starfire, 2 Wrath, cast prepullTime
		// seconds before the pull (3 for Starfire with Improved Starfire, 1.5 for Wrath with
		// Improved Wrath, so the spell lands at 0 plus its travel time).
		{Name: "prepull", Default: 1, Min: 0, Max: 2, Step: 1},
		{Name: "prepullTime", Default: 3, Min: 1.5, Max: 3.5, Step: 0.5},
		// Use the synthetic spell power trinket 0.1 sec before the pull, so it is up when the
		// pre-pull cast lands, 0/1 (only with the trinket worn).
		{Name: "prepullTrinket", Default: 0, Min: 0, Max: 1, Step: 1},
		// Cooldown stacking (rotation auditor), 0/1: Major Spellblasting Potion and the Night Elf's
		// Elune's Light 0.1 sec before the pull as well, so the pre-pull cast lands inside them, and
		// Elune's later uses held for the synthetic trinket when holding loses no use.
		{Name: "cdStack", Default: 0, Min: 0, Max: 1, Step: 1},
		// Elune's Light (the Night Elf's racial) in the cdStack lines, 0/1. Other races leave it
		// out; a Night Elf without the lines still uses it through the autocast line.
		{Name: "elune", Default: 1, Min: 0, Max: 1, Step: 1},
		// Moonfire at the end of the fight once the filler would not land before it ends (rotation
		// auditor): 0 never, 1 when a Starfire would not, 2 when a Wrath would not. endMoonfireSlack
		// adds seconds to that cast time.
		{Name: "endMoonfire", Default: 0, Min: 0, Max: 2, Step: 1},
		{Name: "endMoonfireSlack", Default: 0, Min: 0, Max: 3, Step: 0.5},
		// Faerie Fire duty (standards 23), 0/1: keep Faerie Fire on the target, cast again once it
		// has less than ffRefresh seconds left. ffPos 0 puts it ahead of the dots, 1 after them.
		{Name: "faerieFire", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "ffRefresh", Default: 1.5, Min: 0, Max: 4, Step: 0.5},
		{Name: "ffPos", Default: 0, Min: 0, Max: 1, Step: 1},
	}
}

func (druidBalance) Build(k Knobs) *proto.APLRotation {
	wrath := rankID(druid.WrathLevel[:], druid.WrathSpellId[:])
	// Starfire rank 7 is an Ahn'Qiraj book rank, which Forever leaves out.
	starfire := rankID(druid.StarfireLevel[:druid.StarfireRanks], druid.StarfireSpellId[:druid.StarfireRanks])
	moonfire := rankID(druid.MoonfireLevel[:], druid.MoonfireSpellId[:])
	insectSwarm := rankID(druid.InsectSwarmLevel[:], druid.InsectSwarmSpellId[:])

	manaOK := value(nil)
	if k["wrathBelowMana"] > 0 {
		manaOK = ge(currentManaPercent(), num(k["wrathBelowMana"]/100))
	}

	var items []*proto.APLListItem
	if k["innervateMana"] > 0 {
		items = append(items, cast(druidInnervate, le(currentManaPercent(), num(k["innervateMana"]/100)),
			"Innervate on ourself once the mana runs low."))
	}
	items = append(items, autocastOtherCooldowns(nil, "Racials, trinkets, potions and runes."))
	if k["cdStack"] == 1 && k["elune"] == 1 {
		items = append(items, cast(elunesLight, holdFor(elunesLight, itemID(druid.BalanceSyntheticSpellPowerTrinket)),
			"Elune's Light with the trinket, unless waiting for it would lose a use."))
	}
	// Forever's Faerie Fire and Curse of Recklessness share the minor armor slot (standards 20), and
	// the sim won't put ours up while a Curse of Recklessness holds it.
	ffItem := cast(druidFaerieFire, and(not(targetAuraIsActive(druidCoRAura)),
		lt(targetAuraRemainingTime(druidFaerieFire), seconds(k["ffRefresh"]))),
		"Faerie Fire on the target, refreshed before it drops, unless a Curse of Recklessness holds the slot (Faerie Fire duty).")
	if k["faerieFire"] == 1 && k["ffPos"] == 0 {
		items = append(items, ffItem)
	}
	if k["insectSwarm"] == 1 {
		items = append(items, cast(insectSwarm, and(not(dotIsActive(insectSwarm)), gt(remainingTime(), seconds(k["dotMinTime"])), manaOK),
			"Insect Swarm whenever it is down. Without the talent ignore the warning on this line."))
	}
	if k["moonfire"] == 1 {
		items = append(items, cast(moonfire, and(not(dotIsActive(moonfire)), gt(remainingTime(), seconds(k["dotMinTime"])), manaOK),
			"Moonfire whenever its dot is down."))
	}
	if k["faerieFire"] == 1 && k["ffPos"] == 1 {
		items = append(items, ffItem)
	}
	if k["clearcastStarfire"] == 1 {
		items = append(items, cast(starfire, auraIsActive(druidClearcasting),
			"A free Starfire on Clearcasting, which Wrath does not use up."))
	}
	if k["wrathBelowMana"] > 0 {
		items = append(items, cast(wrath, lt(currentManaPercent(), num(k["wrathBelowMana"]/100)),
			"Wrath only once the mana runs low."))
	}
	if k["endMoonfire"] > 0 {
		filler, name := starfire, "Starfire"
		if k["endMoonfire"] == 2 {
			filler, name = wrath, "Wrath"
		}
		castTime := spellCastTime(filler)
		if k["endMoonfireSlack"] > 0 {
			castTime = add(castTime, seconds(k["endMoonfireSlack"]))
		}
		items = append(items, cast(moonfire, lt(remainingTime(), castTime),
			"Moonfire when a "+name+" would not land before the end."))
	}
	switch k["filler"] {
	case 0:
		items = append(items, cast(starfire, nil, "Starfire."))
	case 1:
		items = append(items, cast(wrath, nil, "Wrath."))
	case 3:
		items = append(items,
			cast(wrath, and(le(auraNumStacks(druidEclipse), num(k["eclipseStacks"])), not(spellInFlight(wrath))),
				"Wrath once Eclipse is down to its last charge and no Wrath is on its way."),
			cast(starfire, nil, "Starfire."))
	case 2:
		ready := ge(auraNumStacks(druidEclipse), num(k["eclipseStacks"]))
		if k["wrathInFlight"] == 1 {
			ready = or(ready, spellInFlight(wrath))
		}
		items = append(items,
			cast(starfire, ready, "Starfire once Eclipse has enough charges, or a Wrath that will give them is on its way."),
			cast(wrath, nil, "Wrath for Eclipse charges."))
	}
	rot := rotation(items...)
	switch k["prepull"] {
	case 1:
		rot.PrepullActions = append(rot.PrepullActions, prepull(starfire, k["prepullTime"]))
	case 2:
		rot.PrepullActions = append(rot.PrepullActions, prepull(wrath, k["prepullTime"]))
	}
	if k["prepullTrinket"] == 1 {
		rot.PrepullActions = append(rot.PrepullActions, prepull(itemID(druid.BalanceSyntheticSpellPowerTrinket), 0.1))
	}
	if k["cdStack"] == 1 {
		rot.PrepullActions = append(rot.PrepullActions, prepull(itemID(majorSpellblastingPotion), 0.1))
		if k["elune"] == 1 {
			rot.PrepullActions = append(rot.PrepullActions, prepull(elunesLight, 0.1))
		}
	}
	return rot
}

func spellInFlight(spell *proto.ActionID) value {
	return &proto.APLValue{Value: &proto.APLValue_SpellInFlight{SpellInFlight: &proto.APLValueSpellInFlight{SpellId: spell}}}
}

// elunesLight (rogue_common.go) is the Night Elf's Forever racial (sim/core/racials_forever.go):
// 10% crit for a while on a long cooldown. Other races do not have it; their rotation skips the line.
const majorSpellblastingPotion = 250937

// holdFor casts spell while trigger's aura is up, or when trigger will not be ready again in
// the fight (with 10 sec to spare), or when waiting for it would cost spell a use.
func holdFor(spell, trigger *proto.ActionID) value {
	return or(
		auraIsActive(trigger),
		gt(spellTimeToReady(trigger), sub(remainingTime(), seconds(10))),
		gt(spellUsesLostByDelay(spell, spellTimeToReady(trigger), seconds(15)), num(0)))
}
