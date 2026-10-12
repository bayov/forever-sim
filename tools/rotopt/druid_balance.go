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
)

func (druidBalance) Knobs() []Knob {
	return []Knob{
		// Keep Moonfire and Insect Swarm up (Insect Swarm needs its talent), 0/1 each.
		{Name: "moonfire", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "insectSwarm", Default: 1, Min: 0, Max: 1, Step: 1},
		// No new dot with less than this many seconds of the fight left.
		{Name: "dotMinTime", Default: 8, Min: 0, Max: 12, Step: 2},
		// The filler: 0 Starfire, 1 Wrath, 2 Wrath until Eclipse has eclipseStacks charges
		// and then Starfire.
		{Name: "filler", Default: 2, Min: 0, Max: 2, Step: 1},
		{Name: "eclipseStacks", Default: 1, Min: 1, Max: 4, Step: 1},
		// Starfire while Clearcasting is up, whatever the filler says, 0/1.
		{Name: "clearcastStarfire", Default: 1, Min: 0, Max: 1, Step: 1},
		// Below this much mana the dots stop and the filler is Wrath, which costs a third of
		// Starfire for each second of casting. 0 never.
		{Name: "wrathBelowMana", Default: 0, Min: 0, Max: 40, Step: 10},
		// Innervate on ourself below this much mana, 0 never.
		{Name: "innervateMana", Default: 40, Min: 0, Max: 80, Step: 10},
	}
}

func (druidBalance) Build(k Knobs) *proto.APLRotation {
	wrath := rankID(druid.WrathLevel[:], druid.WrathSpellId[:])
	starfire := rankID(druid.StarfireLevel[:], druid.StarfireSpellId[:])
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
	if k["insectSwarm"] == 1 {
		items = append(items, cast(insectSwarm, and(not(dotIsActive(insectSwarm)), gt(remainingTime(), seconds(k["dotMinTime"])), manaOK),
			"Insect Swarm whenever it is down. Without the talent ignore the warning on this line."))
	}
	if k["moonfire"] == 1 {
		items = append(items, cast(moonfire, and(not(dotIsActive(moonfire)), gt(remainingTime(), seconds(k["dotMinTime"])), manaOK),
			"Moonfire whenever its dot is down."))
	}
	if k["clearcastStarfire"] == 1 {
		items = append(items, cast(starfire, auraIsActive(druidClearcasting),
			"A free Starfire on Clearcasting, which Wrath does not use up."))
	}
	if k["wrathBelowMana"] > 0 {
		items = append(items, cast(wrath, lt(currentManaPercent(), num(k["wrathBelowMana"]/100)),
			"Wrath only once the mana runs low."))
	}
	switch k["filler"] {
	case 0:
		items = append(items, cast(starfire, nil, "Starfire."))
	case 1:
		items = append(items, cast(wrath, nil, "Wrath."))
	case 2:
		items = append(items,
			cast(starfire, ge(auraNumStacks(druidEclipse), num(k["eclipseStacks"])), "Starfire once Eclipse has enough charges."),
			cast(wrath, nil, "Wrath for Eclipse charges."))
	}
	return rotation(items...)
}
