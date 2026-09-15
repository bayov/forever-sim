package main

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

// Enhancement under Forever rules. Every line names the highest rank the player's level
// knows, so the rotation file reads right in the UI at that level. The 60 presets are
// hand written (the air totem twisting is not a knob), this template is for lower levels
// and for checking how much each spell is worth: a 0/1 knob per spell, and mana floors
// for the ones that burn the pool.
type shamanEnh struct{}

// playerLevel is the level the templates build for, set once the settings are loaded.
var playerLevel int32 = core.CharacterMaxLevel

// The spell ID of the highest rank known at playerLevel, from a rank table.
func rankID[L int | int32](levels []L, ids []int32) *proto.ActionID {
	return spellID(ids[core.HighestRankAt(playerLevel, levels)])
}

func (shamanEnh) Knobs() []Knob {
	return []Knob{
		{Name: "strengthOfEarth", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "searing", Default: 1, Min: 0, Max: 1, Step: 1},
		// No new Searing Totem with less than this long to go, its mana is better in a shock.
		{Name: "searingMinTime", Default: 20, Min: 10, Max: 30, Step: 10},
		{Name: "shield", Default: 0, Min: 0, Max: 1, Step: 1},
		// 0 none, 1 Earth Shock, 2 Flame Shock, 3 Frost Shock, 4 Flame Shock when its DoT
		// is down and Earth Shock otherwise.
		{Name: "shock", Default: 1, Min: 0, Max: 4, Step: 1},
		// Shocks only above this much mana, so the pool is not gone by the middle of the fight.
		{Name: "shockMana", Default: 0, Min: 0, Max: 80, Step: 20},
		{Name: "lightningBolt", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "lightningBoltMana", Default: 60, Min: 0, Max: 80, Step: 20},
	}
}

func (shamanEnh) Build(k Knobs) *proto.APLRotation {
	strengthOfEarth := rankID(shaman.StrengthOfEarthTotemLevel[:], shaman.StrengthOfEarthTotemSpellId[:])
	searingTotem := rankID(shaman.SearingTotemLevel[:], shaman.SearingTotemSpellId[:])
	lightningShield := rankID(shaman.LightningShieldLevel[:], shaman.LightningShieldSpellId[:])
	earthShock := rankID(shaman.EarthShockLevel[:], shaman.EarthShockSpellId[:])
	flameShock := rankID(shaman.FlameShockLevel[:], shaman.FlameShockSpellId[:])
	frostShock := rankID(shaman.FrostShockLevel[:], shaman.FrostShockSpellId[:])
	lightningBolt := rankID(shaman.LightningBoltLevel[:], shaman.LightningBoltSpellId[:])
	stormstrike := spellID(17364)

	var items []*proto.APLListItem
	if k["strengthOfEarth"] == 1 {
		items = append(items, cast(strengthOfEarth, not(auraIsActive(strengthOfEarth)), "Strength of Earth Totem whenever it is down."))
	}
	if k["searing"] == 1 {
		items = append(items, cast(searingTotem, and(
			not(dotIsActive(searingTotem)),
			ge(remainingTime(), seconds(k["searingMinTime"])),
		), "Searing Totem whenever it is down, unless the fight is about to end."))
	}
	items = append(items, autocastOtherCooldowns(nil, "Racials and trinkets on cooldown."))
	items = append(items, cast(stormstrike, nil, "Stormstrike on cooldown. Below 40 there is none, ignore the warning on this line."))
	if k["shield"] == 1 {
		items = append(items, cast(lightningShield, not(auraIsActive(lightningShield)), "Lightning Shield whenever it is down."))
	}
	manaFloor := func(pct float64) value {
		if pct == 0 {
			return nil
		}
		return ge(currentManaPercent(), num(pct/100))
	}
	switch k["shock"] {
	case 1:
		items = append(items, cast(earthShock, manaFloor(k["shockMana"]), "Earth Shock on cooldown."))
	case 2:
		items = append(items, cast(flameShock, and(not(dotIsActive(flameShock)), manaFloor(k["shockMana"])), "Flame Shock when its DoT is down."))
	case 3:
		items = append(items, cast(frostShock, manaFloor(k["shockMana"]), "Frost Shock on cooldown."))
	case 4:
		items = append(items,
			cast(flameShock, and(not(dotIsActive(flameShock)), manaFloor(k["shockMana"])), "Flame Shock when its DoT is down."),
			cast(earthShock, manaFloor(k["shockMana"]), "Earth Shock otherwise."),
		)
	}
	if k["lightningBolt"] == 1 {
		items = append(items, cast(lightningBolt, manaFloor(k["lightningBoltMana"]), "Lightning Bolt with the mana to spare. It stops the swing timer for the cast."))
	}
	return rotation(items...)
}
