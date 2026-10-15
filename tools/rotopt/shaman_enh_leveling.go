package main

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

// Enhancement while leveling solo (a leveling session, sim/core/leveling.go): the score is
// mobs killed in a 20 min session. The rules come from docs/leveling_sim.md and the leveling
// auditor's findings (orchestration/auditor_leveling.md); the knobs are the thresholds and
// the yes/no choices that depend on the level, gear and mobs:
//
//   - Out of combat: Totemic Recall, a heal, a bandage, or sitting down to eat and drink.
//   - Pulling the mob we found: a shock within 20 yards, a rank 1 Lightning Bolt from
//     further while nothing is on us, or walking up into range. A rotation that does none
//     of these walks up until the mob notices us.
//   - Chain pulls: look for another mob while health and mana are above thresholds and
//     fewer than a number of mobs are pulled.
//   - Totems only once a mob is pulled, dropped again when out of reach.
//   - The fight: Stormstrike, the shocks above a mana floor, Fire Nova on packs, Lightning
//     Shield, a heal below a threshold, and the cooldowns.
type shamanEnhLeveling struct{}

func (shamanEnhLeveling) Knobs() []Knob {
	return []Knob{
		// Out of combat.
		{Name: "recall", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "oocHealBelow", Default: 0, Min: 0, Max: 60, Step: 20},
		{Name: "bandageBelow", Default: 0, Min: 0, Max: 80, Step: 20},
		{Name: "restHealth", Default: 50, Min: 20, Max: 80, Step: 10},
		{Name: "restMana", Default: 40, Min: 10, Max: 70, Step: 10},
		{Name: "restUntil", Default: 100, Min: 60, Max: 100, Step: 10},
		// Pulling: 0 none (walk up), 1 Earth Shock rank 1, 2 Earth Shock, 3 Flame Shock.
		{Name: "pullSpell", Default: 2, Min: 0, Max: 3, Step: 1},
		{Name: "pullBolt", Default: 0, Min: 0, Max: 1, Step: 1},
		// Chain pulls: 0 off, else look while fewer than chainMax mobs are pulled.
		{Name: "chainMax", Default: 2, Min: 0, Max: 4, Step: 1},
		{Name: "chainHealth", Default: 60, Min: 30, Max: 90, Step: 10},
		{Name: "chainMana", Default: 40, Min: 0, Max: 80, Step: 20},
		// Totems, once a mob is pulled. Fire: 0 none, 1 Flametongue, 2 Searing, 3 Magma
		// with 2 mobs on us.
		{Name: "strengthOfEarth", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "manaSpring", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "fireTotem", Default: 0, Min: 0, Max: 3, Step: 1},
		// The fight.
		{Name: "shockMana", Default: 10, Min: 0, Max: 60, Step: 10},
		{Name: "flameShock", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "fireNova", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "lightningShield", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "combatHealBelow", Default: 0, Min: 0, Max: 50, Step: 10},
	}
}

func (shamanEnhLeveling) Build(k Knobs) *proto.APLRotation {
	soeLevels, soeIDs := shaman.StrengthOfEarthTotemLevel[:], shaman.StrengthOfEarthTotemSpellId[:]
	strengthOfEarth := rankID(soeLevels[:len(soeLevels)-1], soeIDs[:len(soeIDs)-1])
	searingTotem := rankID(shaman.SearingTotemLevel[:], shaman.SearingTotemSpellId[:])
	magmaTotem := rankID(shaman.MagmaTotemLevel[:], shaman.MagmaTotemSpellId[:])
	flametongueTotem := rankID(shaman.FlametongueTotemLevel[:], shaman.FlametongueTotemSpellId[:])
	manaSpringTotem := rankID(shaman.ManaSpringTotemLevel[:], shaman.ManaSpringTotemSpellId[:])
	fireNova := rankID(shaman.FireNovaLevel[:], shaman.FireNovaSpellId[:])
	lightningShield := rankID(shaman.LightningShieldLevel[:], shaman.LightningShieldSpellId[:])
	earthShock := rankID(shaman.EarthShockLevel[:], shaman.EarthShockSpellId[:])
	earthShock1 := spellID(shaman.EarthShockSpellId[1])
	flameShock := rankID(shaman.FlameShockLevel[:], shaman.FlameShockSpellId[:])
	lightningBolt1 := spellID(shaman.LightningBoltSpellId[1])
	lesserHealingWave := spellID(shaman.HighestHealRankID(shaman.LesserHealingWaveRanks, playerLevel))
	healingWave := spellID(shaman.HighestHealRankID(shaman.HealingWaveRanks, playerLevel))
	heal := lesserHealingWave
	if lesserHealingWave.GetSpellId() == 0 {
		heal = healingWave
	}
	totemRecall := spellID(36936)

	noMobs := eq(mobsPulled(), num(0))
	inCombat := ge(mobsPulled(), num(1))
	var items []*proto.APLListItem

	// Between pulls.
	if k["recall"] == 1 && playerLevel >= 20 {
		items = append(items, cast(totemRecall, and(noMobs, eq(pullTargetDistance(), num(0))),
			"Totemic Recall once the fight is over: a quarter of the totems' mana back, and they'd be left behind anyway."))
	}
	if k["oocHealBelow"] > 0 {
		items = append(items, cast(heal, and(noMobs, lt(currentHealthPercent(), pct(k["oocHealBelow"]))),
			"A heal between pulls instead of eating."))
	}
	if k["bandageBelow"] > 0 {
		items = append(items, bandage(lt(currentHealthPercent(), pct(k["bandageBelow"])), "A bandage between pulls."))
	}
	items = append(items, rest(true, true, k["restUntil"],
		or(lt(currentHealthPercent(), pct(k["restHealth"])), lt(currentManaPercent(), pct(k["restMana"]))),
		"Sit down to eat and drink when health or mana is low."))

	// The fight's heal comes before pulling more.
	if k["combatHealBelow"] > 0 {
		items = append(items, cast(heal, and(inCombat, lt(currentHealthPercent(), pct(k["combatHealBelow"]))),
			"A heal in combat when health runs low."))
	}

	// Pulling the mob we found.
	var pullShock *proto.ActionID
	switch k["pullSpell"] {
	case 1:
		pullShock = earthShock1
	case 2:
		pullShock = earthShock
	case 3:
		pullShock = flameShock
	}
	if pullShock != nil {
		items = append(items, castAt(pullShock, pullTarget, nil, "Pull the mob we found with a shock once it's in range."))
	}
	if k["pullBolt"] == 1 {
		items = append(items,
			castAt(lightningBolt1, pullTarget, eq(mobsOnUs(), num(0)), "Pull with rank 1 Lightning Bolt from beyond shock range while nothing is on us."),
			walkToPullTarget(30, eq(mobsOnUs(), num(0)), "Walk into bolt range."))
	}
	shockRange := 0.0
	if pullShock != nil {
		shockRange = 20
	}
	items = append(items, walkToPullTarget(shockRange, nil, "Walk up to the mob we found, into shock range or until it notices us."))

	// Chain pulls.
	if k["chainMax"] > 0 {
		items = append(items, lookForMob(and(
			lt(mobsPulled(), num(k["chainMax"])),
			gt(currentHealthPercent(), pct(k["chainHealth"])),
			ge(currentManaPercent(), pct(k["chainMana"])),
		), "Look for another mob while health and mana allow."))
	}

	// Totems once a mob is pulled, again when we've moved out of their reach.
	dropTotem := func(totem *proto.ActionID, slot proto.ShamanTotems_TotemType, reach float64, extra value, notes string) {
		items = append(items, cast(totem, and(inCombat, extra, or(eq(totemRemainingTime(slot), num(0)), gt(totemDistance(slot), num(reach)))), notes))
	}
	if k["strengthOfEarth"] == 1 {
		dropTotem(strengthOfEarth, proto.ShamanTotems_Earth, 30, nil, "Strength of Earth Totem in combat.")
	}
	if k["manaSpring"] == 1 && manaSpringTotem.GetSpellId() != 0 {
		dropTotem(manaSpringTotem, proto.ShamanTotems_Water, 30, nil, "Mana Spring Totem in combat.")
	}
	switch k["fireTotem"] {
	case 1:
		dropTotem(flametongueTotem, proto.ShamanTotems_Fire, 30, nil, "Flametongue Totem in combat.")
	case 2:
		dropTotem(searingTotem, proto.ShamanTotems_Fire, 15, ge(mobsOnUs(), num(1)), "Searing Totem with a mob on us.")
	case 3:
		dropTotem(magmaTotem, proto.ShamanTotems_Fire, 4, ge(mobsOnUs(), num(2)), "Magma Totem with two mobs on us.")
	}

	// The fight.
	shockFloor := ge(currentManaPercent(), pct(k["shockMana"]))
	items = append(items, cast(stormstrike, nil, "Stormstrike on cooldown. Without the talent ignore the warning on this line."))
	if k["flameShock"] == 1 {
		items = append(items, cast(flameShock, and(inCombat, not(dotIsActive(flameShock)), shockFloor), "Flame Shock when its DoT is down."))
	}
	items = append(items, cast(earthShock, and(inCombat, shockFloor), "Earth Shock otherwise."))
	if k["fireNova"] == 1 {
		items = append(items, cast(fireNova, and(ge(mobsOnUs(), num(2)), shockFloor), "Fire Nova with two mobs on us, around the fire totem."))
	}
	if k["lightningShield"] == 1 {
		items = append(items, cast(lightningShield, not(auraIsActive(lightningShield)), "Lightning Shield whenever it is down."))
	}
	items = append(items, autocastOtherCooldowns(inCombat, "Racials and trinkets in combat."))
	return rotation(items...)
}
