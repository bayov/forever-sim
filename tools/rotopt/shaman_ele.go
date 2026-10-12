package main

import (
	"math"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

// Elemental under Forever rules, a caster at range in a raid. Every line names the highest rank
// the player's level knows, except the Lightning Bolt that downranks when the mana runs low.
// Lines for spells the build cannot have (Lava Burst without the talent, Mana Tide Totem) only
// warn, and a knob at 0 drops its line entirely.
//
// The shaman stands 20 yards or more from the boss, so Searing Totem (20 yards) is the only fire
// totem that reaches it. Magma Totem and Fire Nova hit around the totem (8 and 10 yards), so the
// template leaves them out.
//
// Forever has no Elemental Mastery. The burst is the racials and the trinkets, which go out on
// cooldown with the other cooldowns (potions, Demonic Rune, Nature's Swiftness).
type shamanEle struct{}

func init() {
	templates["shaman_ele"] = shamanEle{}
}

func (shamanEle) Knobs() []Knob {
	return []Knob{
		// 0 none, 1 Mana Spring Totem.
		{Name: "waterTotem", Default: 1, Min: 0, Max: 1, Step: 1},
		// Mana Tide Totem (the Restoration talent) once the shaman is down to this much mana,
		// 0 never. It takes Mana Spring Totem's place for its 12 sec.
		{Name: "manaTideMana", Default: 0, Min: 0, Max: 80, Step: 20},
		// 0 none, 1 Searing Totem.
		{Name: "fireTotem", Default: 1, Min: 0, Max: 1, Step: 1},
		// No new fire totem with less than this long to go, its mana is better in a bolt.
		{Name: "fireTotemMinTime", Default: 20, Min: 10, Max: 30, Step: 10},
		// Flame Shock whenever its DoT is down, 0 never. Lava Burst does 20% more while it ticks.
		{Name: "flameShock", Default: 1, Min: 0, Max: 1, Step: 1},
		// 0 never, 1 Lava Burst on cooldown, 2 only while Flame Shock ticks.
		{Name: "lavaBurst", Default: 1, Min: 0, Max: 2, Step: 1},
		// 0 never, 1 Chain Lightning on cooldown (one target, so it is a 2 sec bolt on a 6 sec
		// cooldown).
		{Name: "chainLightning", Default: 1, Min: 0, Max: 1, Step: 1},
		// Chain Lightning ahead of Lava Burst when both are ready.
		{Name: "chainLightningFirst", Default: 0, Min: 0, Max: 1, Step: 1},
		// Chain Lightning only above this much mana, or while Clearcasting makes it free.
		{Name: "chainLightningMana", Default: 0, Min: 0, Max: 80, Step: 10},
		// Earth Shock on the shared shock cooldown when Flame Shock does not need it, 0 never.
		{Name: "earthShock", Default: 0, Min: 0, Max: 1, Step: 1},
		// Shocks only above this much mana.
		{Name: "shockMana", Default: 0, Min: 0, Max: 80, Step: 10},
		// Below lightningBoltDownrankMana, Lightning Bolt drops this many ranks to save mana, 0
		// never. Forever gave the low ranks close to full coefficients.
		{Name: "lightningBoltDownrank", Default: 0, Min: 0, Max: 4, Step: 1},
		{Name: "lightningBoltDownrankMana", Default: 20, Min: 10, Max: 60, Step: 10},
		// 1 puts the totems down in the prepull actions.
		{Name: "prepullTotems", Default: 1, Min: 0, Max: 1, Step: 1},
		// 1 starts a Lightning Bolt 2.9 sec before the pull, so it lands just after the fight
		// starts. The totems go down around it.
		{Name: "prepullBolt", Default: 0, Min: 0, Max: 1, Step: 1},
		// Lava Burst waits while Flame Shock has this many seconds or less left, since the 20%
		// counts when the cast ends. 0 never waits.
		{Name: "lavaBurstFlameShockLeft", Default: 0, Min: 0, Max: 3, Step: 1},
	}
}

func (shamanEle) Build(k Knobs) *proto.APLRotation {
	searingTotem := rankID(shaman.SearingTotemLevel[:], shaman.SearingTotemSpellId[:])
	manaSpringTotem := rankID(shaman.ManaSpringTotemLevel[:], shaman.ManaSpringTotemSpellId[:])
	manaTideTotem := rankID(shaman.ManaTideTotemLevel[:], shaman.ManaTideTotemSpellId[:])
	earthShock := rankID(shaman.EarthShockLevel[:], shaman.EarthShockSpellId[:])
	flameShock := rankID(shaman.FlameShockLevel[:], shaman.FlameShockSpellId[:])
	lightningBolt := rankID(shaman.LightningBoltLevel[:], shaman.LightningBoltSpellId[:])
	chainLightning := rankID(shaman.ChainLightningLevel[:], shaman.ChainLightningSpellId[:])
	lavaBurst := rankID(shaman.LavaBurstLevel, shaman.LavaBurstSpellId)

	manaFloor := func(pct float64) value {
		if pct == 0 {
			return nil
		}
		return ge(currentManaPercent(), num(pct/100))
	}

	var items []*proto.APLListItem
	var totems []*proto.ActionID
	items = append(items, autocastOtherCooldowns(nil, "Racials, trinkets, potions, Demonic Rune and Nature's Swiftness on cooldown."))
	if k["manaTideMana"] > 0 {
		items = append(items, cast(manaTideTotem, le(currentManaPercent(), num(k["manaTideMana"]/100)),
			"Mana Tide Totem once the mana runs low. Without the talent ignore the warning on this line."))
	}
	if k["waterTotem"] == 1 {
		totems = append(totems, manaSpringTotem)
		when := not(auraIsActive(manaSpringTotem))
		if k["manaTideMana"] > 0 {
			// Mana Spring would knock the tide down at once, so it waits for it to end.
			when = and(when, not(auraIsActive(manaTideTotem)))
		}
		items = append(items, cast(manaSpringTotem, when, "Mana Spring Totem whenever it is down."))
	}
	if k["fireTotem"] == 1 {
		totems = append(totems, searingTotem)
		items = append(items, cast(searingTotem, and(
			not(dotIsActive(searingTotem)),
			ge(remainingTime(), seconds(k["fireTotemMinTime"])),
		), "Searing Totem whenever it is down, unless the fight is about to end. It reaches 20 yards."))
	}
	if k["flameShock"] == 1 {
		items = append(items, cast(flameShock, and(not(dotIsActive(flameShock)), manaFloor(k["shockMana"])), "Flame Shock when its DoT is down."))
	}
	lavaBurstLine := func() {
		switch k["lavaBurst"] {
		case 1:
			var when value
			notes := "Lava Burst on cooldown. Without the talent ignore the warning on this line."
			if left := k["lavaBurstFlameShockLeft"]; left > 0 && k["flameShock"] == 1 {
				when = or(not(dotIsActive(flameShock)), gt(dotRemainingTime(flameShock), seconds(left)))
				notes = "Lava Burst on cooldown, unless Flame Shock runs out before the cast ends. Without the talent ignore the warning on this line."
			}
			items = append(items, cast(lavaBurst, when, notes))
		case 2:
			items = append(items, cast(lavaBurst, dotIsActive(flameShock),
				"Lava Burst while Flame Shock ticks (20% more). Without the talent ignore the warning on this line."))
		}
	}
	chainLightningLine := func() {
		if k["chainLightning"] == 1 {
			when := manaFloor(k["chainLightningMana"])
			if when != nil {
				// The sim drops a condition on an aura the build does not have, which leaves just
				// the mana floor without Elemental Focus.
				when = or(when, auraIsActive(clearcasting))
			}
			items = append(items, cast(chainLightning, when, "Chain Lightning on cooldown."))
		}
	}
	if k["chainLightningFirst"] == 1 {
		chainLightningLine()
		lavaBurstLine()
	} else {
		lavaBurstLine()
		chainLightningLine()
	}
	if k["earthShock"] == 1 {
		when := manaFloor(k["shockMana"])
		if k["flameShock"] == 1 {
			// Flame Shock shares the shock cooldown, so Earth Shock only goes when the DoT will
			// still be up once the cooldown is back.
			when = and(when, gt(dotRemainingTime(flameShock), seconds(6)))
		}
		items = append(items, cast(earthShock, when, "Earth Shock on the shock cooldown."))
	}
	if r := int(k["lightningBoltDownrank"]); r > 0 {
		top := len(shaman.LightningBoltLevel) - 1
		for top > 0 && shaman.LightningBoltLevel[top] > int(playerLevel) {
			top--
		}
		low := spellID(shaman.LightningBoltSpellId[max(1, top-r)])
		items = append(items, cast(low, lt(currentManaPercent(), num(k["lightningBoltDownrankMana"]/100)),
			"A lower rank of Lightning Bolt once the mana runs low."))
	}
	items = append(items, cast(lightningBolt, nil, "Lightning Bolt."))

	rot := rotation(items...)
	// The bolt starts 2.9 sec before the pull. Its 2 sec cast (Elemental Alacrity 3) ends 0.9 sec
	// before the pull, and the missile flies 1 sec from 20 yards, so it lands 0.1 sec into the fight.
	// The other totems go down one totem GCD (1 sec) apart before it. Searing Totem goes down right
	// after it, 0.9 sec before the pull: its first attack comes 2.5 sec later, after the pull, and
	// the standard only counts damage from the pull on.
	boltLead := 0.0
	if k["prepullBolt"] == 1 {
		boltLead = 2.9
	}
	if k["prepullTotems"] == 1 {
		var early []*proto.ActionID
		for _, id := range totems {
			if k["prepullBolt"] == 1 && id == searingTotem {
				continue
			}
			early = append(early, id)
		}
		for i, id := range early {
			rot.PrepullActions = append(rot.PrepullActions, prepull(id, boltLead+float64(len(early)-i)))
		}
	}
	if k["prepullBolt"] == 1 {
		rot.PrepullActions = append(rot.PrepullActions, prepull(lightningBolt, boltLead))
		if k["prepullTotems"] == 1 && k["fireTotem"] == 1 {
			rot.PrepullActions = append(rot.PrepullActions, prepull(searingTotem, math.Round((boltLead-2)*10)/10))
		}
	}
	return rot
}
