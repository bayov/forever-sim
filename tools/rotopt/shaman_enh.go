package main

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

// Enhancement under Forever rules. Every line names the highest rank the player's level
// knows, so the rotation file reads right in the UI at that level. Lines for spells the
// build cannot have (Stormstrike under 16 talent points, Rage of the Farseer without the
// talent) only warn, and a knob at 0 drops its line entirely.
//
// Rage of the Farseer is the anchor of the burst window (30% attack and cast speed for
// 25 sec). Blood Fury and Berserking multiply with it, so they can wait for it the way
// the rogue cooldowns wait for Adrenaline Rush, when the wait costs them no use. The
// totems, the shield and the shock are 0/1 knobs, since what each is worth depends on
// whether a raid is already providing the buff and whether the boss is hitting the shaman.
type shamanEnh struct{}

// playerLevel is the level the templates build for, set once the settings are loaded.
var playerLevel int32 = core.CharacterMaxLevel

// The spell ID of the highest rank known at playerLevel, from a rank table.
func rankID[L int | int32](levels []L, ids []int32) *proto.ActionID {
	return spellID(ids[core.HighestRankAt(playerLevel, levels)])
}

var (
	stormstrike      = spellID(17364)
	maelstromWeapon  = spellID(51530)
	rageOfTheFarseer = spellID(2825)
	naturesSwiftness = spellID(16188)
)

func (shamanEnh) Knobs() []Knob {
	return []Knob{
		{Name: "strengthOfEarth", Default: 1, Min: 0, Max: 1, Step: 1},
		// 0 none, 1 Grace of Air, 2 Windfury Totem. The sim gives the shaman nothing for
		// its own Windfury Totem (it does not stack with Windfury Weapon), so 2 is for the
		// melee group's sake, and the raid buffs already give Grace of Air when a second
		// shaman is set.
		{Name: "airTotem", Default: 1, Min: 0, Max: 2, Step: 1},
		// 0 none, 1 Searing Totem, 2 Magma Totem.
		{Name: "fireTotem", Default: 1, Min: 0, Max: 2, Step: 1},
		// No new fire totem with less than this long to go, its mana is better in a shock.
		{Name: "fireTotemMinTime", Default: 20, Min: 10, Max: 30, Step: 10},
		{Name: "shield", Default: 0, Min: 0, Max: 1, Step: 1},
		// 0 none, 1 Earth Shock, 2 Flame Shock, 3 Frost Shock, 4 Flame Shock when its DoT
		// is down and Earth Shock otherwise, 5 Earth Shock while the Stormstrike debuff (20%
		// Nature damage) is on the target and Frost Shock otherwise, 6 Flame Shock when its
		// DoT is down and Frost Shock otherwise.
		{Name: "shock", Default: 1, Min: 0, Max: 6, Step: 1},
		// Shocks only above this much mana, so the pool is not gone by the middle of the fight.
		{Name: "shockMana", Default: 0, Min: 0, Max: 80, Step: 20},
		// Lightning Bolt at this many Maelstrom Weapon stacks (5 is instant), 0 never.
		{Name: "maelstromStacks", Default: 5, Min: 0, Max: 5, Step: 1},
		// Hard cast Lightning Bolt with the mana to spare, on top of the Maelstrom ones.
		{Name: "lightningBolt", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "lightningBoltMana", Default: 60, Min: 0, Max: 80, Step: 20},
		{Name: "bloodFuryHoldForRage", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "berserkingHoldForRage", Default: 1, Min: 0, Max: 1, Step: 1},
		// Rage of the Farseer itself waits for Stormstrike to be ready, so the window opens
		// with it.
		{Name: "rageWithStormstrike", Default: 0, Min: 0, Max: 1, Step: 1},
	}
}

func (shamanEnh) Build(k Knobs) *proto.APLRotation {
	strengthOfEarth := rankID(shaman.StrengthOfEarthTotemLevel[:], shaman.StrengthOfEarthTotemSpellId[:])
	graceOfAir := rankID(shaman.GraceOfAirTotemLevel[:], shaman.GraceOfAirTotemSpellId[:])
	windfuryTotem := rankID(shaman.WindfuryTotemLevel[:], shaman.WindfuryTotemSpellId[:])
	// The totem's buff is its own aura, refreshed every 5 sec while the totem stands.
	windfuryTotemAura := rankID(shaman.WindfuryTotemLevel[:], shaman.WindfuryBuffAuraId[:])
	searingTotem := rankID(shaman.SearingTotemLevel[:], shaman.SearingTotemSpellId[:])
	magmaTotem := rankID(shaman.MagmaTotemLevel[:], shaman.MagmaTotemSpellId[:])
	lightningShield := rankID(shaman.LightningShieldLevel[:], shaman.LightningShieldSpellId[:])
	earthShock := rankID(shaman.EarthShockLevel[:], shaman.EarthShockSpellId[:])
	flameShock := rankID(shaman.FlameShockLevel[:], shaman.FlameShockSpellId[:])
	frostShock := rankID(shaman.FrostShockLevel[:], shaman.FrostShockSpellId[:])
	lightningBolt := rankID(shaman.LightningBoltLevel[:], shaman.LightningBoltSpellId[:])

	var items []*proto.APLListItem
	// The totems go down before the pull too, one GCD apart, fire totem last.
	var totems []*proto.ActionID
	totem := func(id *proto.ActionID, aura *proto.ActionID, notes string) {
		totems = append(totems, id)
		items = append(items, cast(id, not(auraIsActive(aura)), notes))
	}
	if k["strengthOfEarth"] == 1 {
		totem(strengthOfEarth, strengthOfEarth, "Strength of Earth Totem whenever it is down.")
	}
	switch k["airTotem"] {
	case 1:
		totem(graceOfAir, graceOfAir, "Grace of Air Totem whenever it is down.")
	case 2:
		totem(windfuryTotem, windfuryTotemAura, "Windfury Totem whenever it is down, for the melee group. The shaman's own Windfury Weapon does not stack with it.")
	}
	fireTotem := func(id *proto.ActionID, notes string) {
		totems = append(totems, id)
		items = append(items, cast(id, and(
			not(dotIsActive(id)),
			ge(remainingTime(), seconds(k["fireTotemMinTime"])),
		), notes))
	}
	switch k["fireTotem"] {
	case 1:
		fireTotem(searingTotem, "Searing Totem whenever it is down, unless the fight is about to end.")
	case 2:
		fireTotem(magmaTotem, "Magma Totem whenever it is down, unless the fight is about to end.")
	}

	holdFor := func(spell *proto.ActionID, anchor *proto.ActionID, hold bool) value {
		if !hold {
			return nil
		}
		return alignedWith(spell, anchor)
	}
	var rageWhen value
	if k["rageWithStormstrike"] == 1 {
		rageWhen = spellIsReady(stormstrike)
	}
	// Rage of the Farseer takes 31 talent points, so the burst window only exists from
	// level 40 (or fewer levels with bonus points). Below that the racials go last, on a GCD nothing else wants: at 20 Blood
	// Fury's 10% of 400 attack power is worth about what the GCD it takes from the shocks
	// costs. Stormstrike sits behind 15 talent points under Forever and is the 16th point
	// itself, so its line exists once the build has 16 points (level 25, or the Forever
	// beta's level 20 with 5 extra points).
	if maxTalentPoints >= 31 {
		items = append(items,
			cast(rageOfTheFarseer, rageWhen, "Rage of the Farseer anchors the burst window. Without the talent ignore the warning on this line."),
			cast(bloodFury, holdFor(bloodFury, rageOfTheFarseer, k["bloodFuryHoldForRage"] == 1),
				"Orc. Blood Fury waits for Rage of the Farseer when the wait costs no use. Other races can ignore the warning on this line."),
			cast(berserking, holdFor(berserking, rageOfTheFarseer, k["berserkingHoldForRage"] == 1),
				"Troll. Berserking waits for Rage of the Farseer when the wait costs no use. Other races can ignore the warning on this line."),
			autocastOtherCooldowns(nil, "Anything not listed above (Sapper, trinkets, potions)."),
		)
	}
	if maxTalentPoints >= 16 {
		items = append(items, cast(stormstrike, nil, "Stormstrike on cooldown. Without the talent ignore the warning on this line."))
	}
	if k["shield"] == 1 {
		items = append(items, cast(lightningShield, not(auraIsActive(lightningShield)), "Lightning Shield whenever it is down."))
	}
	if k["maelstromStacks"] > 0 {
		items = append(items, cast(lightningBolt, ge(auraNumStacks(maelstromWeapon), num(k["maelstromStacks"])),
			"Lightning Bolt once Maelstrom Weapon has stacked enough to make it quick. Without the talent ignore the warning on this line."))
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
	case 5:
		items = append(items,
			cast(earthShock, and(targetAuraIsActive(stormstrike), manaFloor(k["shockMana"])), "Earth Shock while the Stormstrike debuff is up (20% Nature damage)."),
			cast(frostShock, manaFloor(k["shockMana"]), "Frost Shock otherwise."),
		)
	case 6:
		items = append(items,
			cast(flameShock, and(not(dotIsActive(flameShock)), manaFloor(k["shockMana"])), "Flame Shock when its DoT is down."),
			cast(frostShock, manaFloor(k["shockMana"]), "Frost Shock otherwise."),
		)
	}
	if k["lightningBolt"] == 1 {
		items = append(items, cast(lightningBolt, manaFloor(k["lightningBoltMana"]), "Lightning Bolt with the mana to spare. It stops the swing timer for the cast."))
	}
	if maxTalentPoints < 31 {
		items = append(items, autocastOtherCooldowns(nil, "Racials and trinkets on a free GCD."))
	}
	rot := rotation(items...)
	for i, id := range totems {
		rot.PrepullActions = append(rot.PrepullActions, prepull(id, 1.5*float64(len(totems)-i)))
	}
	return rot
}
