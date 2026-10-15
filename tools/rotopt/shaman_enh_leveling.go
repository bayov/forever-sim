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
//   - Chain pulls: look for another mob while fewer than a number of mobs are pulled, and
//     pull it while health and mana are above thresholds. With look-ahead we look at lower
//     health too, and a mob we don't pull yet waits while we rest: the next pull without a
//     search (the leveling auditor's first finding, orchestration/leveling_audit_log.md).
//     Nothing pulls the mob we found while we need a rest.
//   - Totems only once a mob is pulled, dropped again when out of reach. Rests wait on mana
//     more than health at level 30 (eating heals 4% a second, drinking 1.8%), so a cheap
//     Mana Spring Totem pays, and so does the cheapest Searing Totem with two mobs pulled
//     (audit round 2).
//   - The fight: Stormstrike, the shocks above a mana floor, Fire Nova on packs, Lightning
//     Shield, a heal below a threshold, and the cooldowns.
type shamanEnhLeveling struct{}

func (shamanEnhLeveling) Knobs() []Knob {
	return []Knob{
		// Out of combat.
		// Totemic Recall's global cooldown before each search costs more than a quarter of
		// the totems' mana gives back (audit round 2).
		{Name: "recall", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "oocHealBelow", Default: 0, Min: 0, Max: 60, Step: 20},
		{Name: "bandageBelow", Default: 0, Min: 0, Max: 80, Step: 20},
		{Name: "restHealth", Default: 90, Min: 20, Max: 90, Step: 10},
		{Name: "restMana", Default: 50, Min: 10, Max: 70, Step: 10},
		{Name: "restUntil", Default: 90, Min: 60, Max: 100, Step: 10},
		{Name: "restUntilMana", Default: 80, Min: 60, Max: 100, Step: 10},
		// Pulling: 0 none (walk up), 1 Earth Shock rank 1, 2 Earth Shock, 3 Flame Shock.
		{Name: "pullSpell", Default: 0, Min: 0, Max: 3, Step: 1},
		{Name: "pullBolt", Default: 0, Min: 0, Max: 1, Step: 1},
		// Chain pulls: 0 off, else look while fewer than chainMax mobs are pulled, and pull
		// what we find above chainHealth and chainMana.
		{Name: "chainMax", Default: 2, Min: 0, Max: 4, Step: 1},
		{Name: "chainHealth", Default: 90, Min: 30, Max: 90, Step: 10},
		{Name: "chainMana", Default: 50, Min: 0, Max: 80, Step: 10},
		// Look for the next mob above this health even when we won't pull it yet; 0 looks
		// only when we'd pull it (chainHealth and chainMana).
		{Name: "lookAheadHealth", Default: 40, Min: 0, Max: 70, Step: 10},
		// With one mob pulled and our target below this share of its health, pull the next
		// one as it dies, above chainDyingHealth and chainDyingMana; 0 off.
		{Name: "chainDyingBelow", Default: 20, Min: 0, Max: 40, Step: 10},
		{Name: "chainDyingHealth", Default: 70, Min: 50, Max: 90, Step: 10},
		{Name: "chainDyingMana", Default: 40, Min: 0, Max: 80, Step: 20},
		// Totems, once a mob is pulled. Fire: 0 none, 1 Flametongue, 2 Searing, 3 Magma
		// with 2 mobs on us.
		{Name: "strengthOfEarth", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "manaSpring", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "fireTotem", Default: 2, Min: 0, Max: 3, Step: 1},
		// Searing Totem: rank 1 (25 mana) or the top rank, and from how many mobs pulled.
		{Name: "searingRank1", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "searingMinPulled", Default: 2, Min: 1, Max: 2, Step: 1},
		// The fight.
		{Name: "shockMana", Default: 10, Min: 0, Max: 60, Step: 10},
		{Name: "flameShock", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "fireNova", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "lightningShield", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "combatHealBelow", Default: 30, Min: 0, Max: 50, Step: 10},
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
	needRest := or(lt(currentHealthPercent(), pct(k["restHealth"])), lt(currentManaPercent(), pct(k["restMana"])))
	items = append(items, rest(true, true, k["restUntil"], k["restUntilMana"], needRest, "Sit down to eat and drink when health or mana is low."))
	// We pull the mob we found once we're fit for it: in combat above the chain thresholds,
	// out of combat once we don't need a rest. Until then it waits.
	chainFit := and(gt(currentHealthPercent(), pct(k["chainHealth"])), ge(currentManaPercent(), pct(k["chainMana"])))
	pullFit := or(and(inCombat, chainFit), and(noMobs, not(needRest)))
	if k["chainDyingBelow"] > 0 {
		// The next mob as our target dies, while we're fit enough for one more.
		pullFit = or(pullFit, and(eq(mobsPulled(), num(1)),
			lt(targetHealthPercent(), pct(k["chainDyingBelow"])),
			gt(currentHealthPercent(), pct(k["chainDyingHealth"])),
			ge(currentManaPercent(), pct(k["chainDyingMana"]))))
	}

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
		items = append(items, castAt(pullShock, pullTarget, pullFit, "Pull the mob we found with a shock once it's in range."))
	}
	if k["pullBolt"] == 1 {
		items = append(items,
			castAt(lightningBolt1, pullTarget, and(pullFit, eq(mobsOnUs(), num(0))), "Pull with rank 1 Lightning Bolt from beyond shock range while nothing is on us."),
			walkToPullTarget(30, and(pullFit, eq(mobsOnUs(), num(0))), "Walk into bolt range."))
	}
	shockRange := 0.0
	if pullShock != nil {
		shockRange = 20
	}
	items = append(items, walkToPullTarget(shockRange, pullFit, "Walk up to the mob we found, into shock range or until it notices us."))

	// Chain pulls.
	if k["chainMax"] > 0 {
		look := chainFit
		if k["lookAheadHealth"] > 0 {
			look = gt(currentHealthPercent(), pct(k["lookAheadHealth"]))
		}
		items = append(items, lookForMob(and(lt(mobsPulled(), num(k["chainMax"])), look),
			"Look for another mob while fighting: the next pull, now or after a rest."))
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
		searing := searingTotem
		if k["searingRank1"] == 1 {
			searing = spellID(shaman.SearingTotemSpellId[1])
		}
		dropTotem(searing, proto.ShamanTotems_Fire, 15, ge(mobsPulled(), num(k["searingMinPulled"])), "Searing Totem with enough mobs pulled.")
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
