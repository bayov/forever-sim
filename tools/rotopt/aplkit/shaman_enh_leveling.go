package aplkit

import (
	"github.com/wowsims/classic/sim/core"
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
//     Mana Spring Totem paid in audit round 2, and so does the cheapest Searing Totem with
//     two mobs pulled. Since pass 7 (Forever's mob health, the user's search times) Mana
//     Spring is off, and a heal below 40% between pulls does part of the resting.
//   - The fight: Stormstrike, the shocks above a mana floor, Fire Nova on packs, Lightning
//     Shield, a heal below a threshold, and the cooldowns.
type ShamanEnhLeveling struct {
	// The level we build the rotation for. It has the abilities we know there, by their top
	// rank, which the sim casts as the highest rank we know (core.TopRankID). So it also
	// plays at the levels after, as the Leveling Plan's level ranges do.
	Level int32
}

func (ShamanEnhLeveling) Knobs() []Knob {
	return []Knob{
		// Out of combat.
		// Totemic Recall's global cooldown before each search costs more than a quarter of
		// the totems' mana gives back (audit round 2).
		{Name: "recall", Default: 0, Min: 0, Max: 1, Step: 1, Needs: []string{"totemicRecall"}, Kind: KnobSteady},
		{Name: "oocHealBelow", Default: 40, Min: 0, Max: 60, Step: 20, Needs: []string{"healingWave", "lesserHealingWave"}, Kind: KnobThreshold},
		{Name: "bandageBelow", Default: 0, Min: 0, Max: 80, Step: 20, Kind: KnobThreshold},
		{Name: "restHealth", Default: 90, Min: 20, Max: 90, Step: 10, Kind: KnobThreshold},
		{Name: "restMana", Default: 50, Min: 10, Max: 70, Step: 10, Kind: KnobThreshold},
		{Name: "restUntil", Default: 90, Min: 60, Max: 100, Step: 10, Kind: KnobThreshold},
		{Name: "restUntilMana", Default: 80, Min: 60, Max: 100, Step: 10, Kind: KnobThreshold},
		// Pulling: 0 none (walk up), 1 Earth Shock rank 1, 2 Earth Shock, 3 Flame Shock.
		{Name: "pullSpell", Default: 3, Min: 0, Max: 3, Step: 1, Needs: []string{"earthShock", "flameShock"}},
		{Name: "pullBolt", Default: 0, Min: 0, Max: 1, Step: 1, Kind: KnobSteady},
		// Chain pulls: 0 off, else look while fewer than chainMax mobs are pulled, and pull
		// what we find above chainHealth and chainMana.
		{Name: "chainMax", Default: 2, Min: 0, Max: 4, Step: 1},
		{Name: "chainHealth", Default: 70, Min: 30, Max: 90, Step: 10, Kind: KnobThreshold},
		{Name: "chainMana", Default: 50, Min: 0, Max: 80, Step: 10, Kind: KnobThreshold},
		// Look for the next mob above this health even when we won't pull it yet; 0 looks
		// only when we'd pull it (chainHealth and chainMana).
		{Name: "lookAheadHealth", Default: 40, Min: 0, Max: 70, Step: 10, Kind: KnobThreshold},
		// With one mob pulled and our target below this share of its health, pull the next
		// one as it dies, above chainDyingHealth and chainDyingMana; 0 off.
		{Name: "chainDyingBelow", Default: 20, Min: 0, Max: 40, Step: 10},
		{Name: "chainDyingHealth", Default: 50, Min: 50, Max: 90, Step: 10, Kind: KnobThreshold},
		{Name: "chainDyingMana", Default: 40, Min: 0, Max: 80, Step: 20, Kind: KnobThreshold},
		// How we look for the next mob while fighting: 0 run around for it, 1 stand and keep
		// fighting (a longer search), 2 stand while a totem of ours is down and in reach, so
		// we keep its buff, and run otherwise.
		{Name: "chainStand", Default: 1, Min: 0, Max: 2, Step: 1, Kind: KnobSteady},
		// In combat, walk up to the mob we found only when it's within this many yards, and
		// otherwise leave it for after the fight; 0 always walks up. Out of combat we always do.
		{Name: "chainWalkMax", Default: 0, Min: 0, Max: 40, Step: 10, Kind: KnobSteady},
		// Totems, once a mob is pulled. Fire: 0 none, 1 Flametongue, 2 Searing, 3 Magma
		// with 2 mobs on us.
		{Name: "strengthOfEarth", Default: 1, Min: 0, Max: 1, Step: 1, Needs: []string{"strengthOfEarth"}, Kind: KnobSteady},
		{Name: "manaSpring", Default: 1, Min: 0, Max: 1, Step: 1, Needs: []string{"manaSpring"}},
		{Name: "fireTotem", Default: 2, Min: 0, Max: 3, Step: 1, Needs: []string{"searingTotem", "magmaTotem", "flametongueTotem"}},
		// Searing Totem: rank 1 (25 mana) or the top rank, and from how many mobs pulled.
		{Name: "searingRank1", Default: 0, Min: 0, Max: 1, Step: 1, Needs: []string{"searingTotem"}, Kind: KnobSteady},
		{Name: "searingMinPulled", Default: 2, Min: 1, Max: 2, Step: 1, Needs: []string{"searingTotem"}},
		// When we've moved out of Mana Spring Totem's reach in a fight and it's still up, walk
		// back to it (1) instead of dropping a new one (0).
		{Name: "totemWalkBack", Default: 0, Min: 0, Max: 1, Step: 1, Needs: []string{"manaSpring"}, Kind: KnobSteady},
		// The fight.
		{Name: "shockMana", Default: 60, Min: 0, Max: 60, Step: 10, Kind: KnobThreshold},
		{Name: "flameShock", Default: 0, Min: 0, Max: 1, Step: 1, Needs: []string{"flameShock"}},
		{Name: "fireNova", Default: 1, Min: 0, Max: 1, Step: 1, Needs: []string{"fireNova"}},
		{Name: "lightningShield", Default: 1, Min: 0, Max: 1, Step: 1, Needs: []string{"lightningShield"}, Kind: KnobSteady},
		{Name: "combatHealBelow", Default: 30, Min: 0, Max: 50, Step: 10, Needs: []string{"healingWave", "lesserHealingWave"}, Kind: KnobThreshold},
		// Past level 30, off by default so the Level 30 rotations stay as they were.
		// Lightning Bolt at this many Maelstrom Weapon stacks (5 is instant), 0 never.
		{Name: "maelstromStacks", Default: 0, Min: 0, Max: 5, Step: 1, Needs: []string{"maelstromWeapon"}},
		// Chain Lightning with at least this many mobs on us, above shockMana; 0 never.
		{Name: "chainLightning", Default: 0, Min: 0, Max: 3, Step: 1, Needs: []string{"chainLightning"}},
		// Air totem once a mob is pulled: 0 none, 1 Grace of Air, 2 Windfury Totem. Windfury
		// Weapon turns Windfury Totem's buff off for us.
		{Name: "airTotem", Default: 0, Min: 0, Max: 2, Step: 1, Needs: []string{"windfuryTotem", "graceOfAir"}},
	}
}

func (t ShamanEnhLeveling) Build(k Knobs) *proto.APLRotation {
	strengthOfEarth := topRankID(t.Level, shaman.StrengthOfEarthTotemLevel[:], shaman.StrengthOfEarthTotemSpellId[:])
	searingTotem := topRankID(t.Level, shaman.SearingTotemLevel[:], shaman.SearingTotemSpellId[:])
	magmaTotem := topRankID(t.Level, shaman.MagmaTotemLevel[:], shaman.MagmaTotemSpellId[:])
	flametongueTotem := topRankID(t.Level, shaman.FlametongueTotemLevel[:], shaman.FlametongueTotemSpellId[:])
	manaSpringTotem := topRankID(t.Level, shaman.ManaSpringTotemLevel[:], shaman.ManaSpringTotemSpellId[:])
	fireNova := topRankID(t.Level, shaman.FireNovaLevel[:], shaman.FireNovaSpellId[:])
	lightningShield := topRankID(t.Level, shaman.LightningShieldLevel[:], shaman.LightningShieldSpellId[:])
	earthShock := topRankID(t.Level, shaman.EarthShockLevel[:], shaman.EarthShockSpellId[:])
	earthShock1 := SpellID(shaman.EarthShockSpellId[1])
	flameShock := topRankID(t.Level, shaman.FlameShockLevel[:], shaman.FlameShockSpellId[:])
	lightningBolt1 := SpellID(shaman.LightningBoltSpellId[1])
	lesserHealingWave := SpellID(shaman.TopHealRankID(shaman.LesserHealingWaveRanks, t.Level))
	healingWave := SpellID(shaman.TopHealRankID(shaman.HealingWaveRanks, t.Level))
	heal := lesserHealingWave
	if lesserHealingWave.GetSpellId() == 0 {
		heal = healingWave
	}
	totemRecall := SpellID(36936)
	lightningBolt := topRankID(t.Level, shaman.LightningBoltLevel[:], shaman.LightningBoltSpellId[:])
	chainLightning := topRankID(t.Level, shaman.ChainLightningLevel[:], shaman.ChainLightningSpellId[:])
	windfuryTotem := topRankID(t.Level, shaman.WindfuryTotemLevel[:], shaman.WindfuryTotemSpellId[:])
	graceOfAir := topRankID(t.Level, shaman.GraceOfAirTotemLevel[:], shaman.GraceOfAirTotemSpellId[:])

	noMobs := Eq(MobsPulled(), Num(0))
	inCombat := Ge(MobsPulled(), Num(1))
	var items []*proto.APLListItem

	// Between pulls.
	if k["recall"] == 1 && t.Level >= 20 {
		items = append(items, Cast(totemRecall, And(noMobs, Eq(PullTargetDistance(), Num(0))),
			"Totemic Recall once the fight is over: a quarter of the totems' mana back, and they'd be left behind anyway."))
	}
	if k["oocHealBelow"] > 0 {
		items = append(items, Cast(heal, And(noMobs, Lt(CurrentHealthPercent(), Pct(k["oocHealBelow"]))),
			"A heal between pulls instead of eating."))
	}
	if k["bandageBelow"] > 0 {
		items = append(items, Bandage(Lt(CurrentHealthPercent(), Pct(k["bandageBelow"])), "A bandage between pulls."))
	}
	needRest := Or(Lt(CurrentHealthPercent(), Pct(k["restHealth"])), Lt(CurrentManaPercent(), Pct(k["restMana"])))
	items = append(items, Rest(true, true, k["restUntil"], k["restUntilMana"], needRest, "Sit down to eat and drink when health or mana is low."))
	// We pull the mob we found once we're fit for it: in combat above the chain thresholds,
	// out of combat once we don't need a rest. Until then it waits.
	chainFit := And(Gt(CurrentHealthPercent(), Pct(k["chainHealth"])), Ge(CurrentManaPercent(), Pct(k["chainMana"])))
	pullFit := Or(And(inCombat, chainFit), And(noMobs, Not(needRest)))
	if k["chainDyingBelow"] > 0 {
		// The next mob as our target dies, while we're fit enough for one more.
		pullFit = Or(pullFit, And(Eq(MobsPulled(), Num(1)),
			Lt(TargetHealthPercent(), Pct(k["chainDyingBelow"])),
			Gt(CurrentHealthPercent(), Pct(k["chainDyingHealth"])),
			Ge(CurrentManaPercent(), Pct(k["chainDyingMana"]))))
	}

	// The fight's heal comes before pulling more.
	if k["combatHealBelow"] > 0 {
		items = append(items, Cast(heal, And(inCombat, Lt(CurrentHealthPercent(), Pct(k["combatHealBelow"]))),
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
		items = append(items, CastAt(pullShock, PullTarget, pullFit, "Pull the mob we found with a shock once it's in range."))
	}
	if k["pullBolt"] == 1 {
		items = append(items,
			CastAt(lightningBolt1, PullTarget, And(pullFit, Eq(MobsOnUs(), Num(0))), "Pull with rank 1 Lightning Bolt from beyond shock range while nothing is on us."),
			WalkToPullTarget(30, And(pullFit, Eq(MobsOnUs(), Num(0))), "Walk into bolt range."))
	}
	shockRange := 0.0
	if pullShock != nil {
		shockRange = 20
	}
	walkFit := pullFit
	if k["chainWalkMax"] > 0 {
		walkFit = And(pullFit, Or(noMobs, Le(PullTargetDistance(), Num(k["chainWalkMax"]))))
	}
	items = append(items, WalkToPullTarget(shockRange, walkFit, "Walk up to the mob we found, into shock range or until it notices us."))

	// Chain pulls.
	if k["chainMax"] > 0 {
		look := chainFit
		if k["lookAheadHealth"] > 0 {
			look = Gt(CurrentHealthPercent(), Pct(k["lookAheadHealth"]))
		}
		look = And(Lt(MobsPulled(), Num(k["chainMax"])), look)
		notes := "Look for another mob while fighting: the next pull, now or after a rest."
		switch k["chainStand"] {
		case 0:
			items = append(items, LookForMob(false, look, notes+" We run around for it."))
		case 1:
			items = append(items, LookForMob(true, look, notes+" We stand and keep fighting while we look."))
		case 2:
			// The totems we'd run out of reach of.
			var up []Value
			if k["manaSpring"] == 1 && manaSpringTotem.GetSpellId() != 0 {
				up = append(up, And(Gt(TotemRemainingTime(proto.ShamanTotems_Water), Num(0)), Le(TotemDistance(proto.ShamanTotems_Water), Num(30))))
			}
			if k["fireTotem"] > 0 {
				up = append(up, Gt(TotemRemainingTime(proto.ShamanTotems_Fire), Num(0)))
			}
			if len(up) > 0 {
				items = append(items, LookForMob(true, And(look, Or(up...)), notes+" We stand while a totem of ours is down, to keep its buff."))
			}
			items = append(items, LookForMob(false, look, notes+" We run around for it otherwise."))
		}
	}

	// Totems once a mob is pulled, again when we've moved out of their reach.
	dropTotem := func(totem *proto.ActionID, slot proto.ShamanTotems_TotemType, reach float64, extra Value, notes string) {
		items = append(items, Cast(totem, And(inCombat, extra, Or(Eq(TotemRemainingTime(slot), Num(0)), Gt(TotemDistance(slot), Num(reach)))), notes))
	}
	if k["strengthOfEarth"] == 1 {
		dropTotem(strengthOfEarth, proto.ShamanTotems_Earth, 30, nil, "Strength of Earth Totem in combat.")
	}
	if k["manaSpring"] == 1 && manaSpringTotem.GetSpellId() != 0 {
		if k["totemWalkBack"] == 1 {
			water := proto.ShamanTotems_Water
			items = append(items, WalkToTotem(water, 20, And(inCombat, Gt(TotemRemainingTime(water), Num(0)), Gt(TotemDistance(water), Num(30))),
				"Walk back into Mana Spring Totem's reach, with the mobs following, instead of dropping a new one."))
		}
		dropTotem(manaSpringTotem, proto.ShamanTotems_Water, 30, nil, "Mana Spring Totem in combat.")
	}
	switch k["fireTotem"] {
	case 1:
		dropTotem(flametongueTotem, proto.ShamanTotems_Fire, 30, nil, "Flametongue Totem in combat.")
	case 2:
		searing := searingTotem
		if k["searingRank1"] == 1 {
			searing = SpellID(shaman.SearingTotemSpellId[1])
		}
		dropTotem(searing, proto.ShamanTotems_Fire, 15, Ge(MobsPulled(), Num(k["searingMinPulled"])), "Searing Totem with enough mobs pulled.")
	case 3:
		dropTotem(magmaTotem, proto.ShamanTotems_Fire, 4, Ge(MobsOnUs(), Num(2)), "Magma Totem with two mobs on us.")
	}

	switch k["airTotem"] {
	case 1:
		if graceOfAir.GetSpellId() != 0 {
			dropTotem(graceOfAir, proto.ShamanTotems_Air, 30, nil, "Grace of Air Totem in combat.")
		}
	case 2:
		if windfuryTotem.GetSpellId() != 0 {
			dropTotem(windfuryTotem, proto.ShamanTotems_Air, 30, nil, "Windfury Totem in combat. Windfury Weapon turns its buff off for us.")
		}
	}

	// The fight.
	shockFloor := Ge(CurrentManaPercent(), Pct(k["shockMana"]))
	items = append(items, Cast(SpellID(17364), nil, "Stormstrike on cooldown. Without the talent ignore the warning on this line."))
	if k["maelstromStacks"] > 0 {
		items = append(items, Cast(lightningBolt, And(inCombat, Ge(AuraNumStacks(SpellID(408505)), Num(k["maelstromStacks"]))),
			"Lightning Bolt at enough Maelstrom Weapon stacks, 5 is instant. Without the talent ignore the warning on this line."))
	}
	if k["chainLightning"] > 0 && chainLightning.GetSpellId() != 0 {
		items = append(items, Cast(chainLightning, And(Ge(MobsOnUs(), Num(k["chainLightning"])), shockFloor), "Chain Lightning with enough mobs on us."))
	}
	if k["flameShock"] == 1 {
		items = append(items, Cast(flameShock, And(inCombat, Not(DotIsActive(flameShock)), shockFloor), "Flame Shock when its DoT is down."))
	}
	items = append(items, Cast(earthShock, And(inCombat, shockFloor), "Earth Shock otherwise."))
	if k["fireNova"] == 1 {
		items = append(items, Cast(fireNova, And(Ge(MobsOnUs(), Num(2)), shockFloor), "Fire Nova with two mobs on us, around the fire totem."))
	}
	if k["lightningShield"] == 1 {
		items = append(items, Cast(lightningShield, Not(AuraIsActive(lightningShield)), "Lightning Shield whenever it is down."))
	}
	items = append(items, AutocastOtherCooldowns(inCombat, "Racials and trinkets in combat."))
	return Rotation(items...)
}

// topRankID names the ability's top rank once the level knows any rank of it, which the sim
// casts as the highest rank the character knows (core.TopRankID). It's spell ID 0 before the
// first rank.
func topRankID[L int | int32](level int32, levels []L, ids []int32) *proto.ActionID {
	if core.HighestRankAt(level, levels) == 0 {
		return SpellID(0)
	}
	return SpellID(ids[len(ids)-1])
}

// Abilities are what the rotation can use at the level, with the talents hasTalent says we
// have (by their field names in the talent trees). The leveling plan's rotation search tunes
// a rotation again from each level that adds one. The weapon imbues are here too, since the
// search picks the imbue of each level range.
func (t ShamanEnhLeveling) Abilities(hasTalent func(name string) bool) []string {
	var out []string
	learned := func(name string, levels []int32) {
		if core.HighestRankAt(t.Level, levels) > 0 {
			out = append(out, name)
		}
	}
	learned("healingWave", shaman.HealRankLevels(shaman.HealingWaveRanks))
	learned("lesserHealingWave", shaman.HealRankLevels(shaman.LesserHealingWaveRanks))
	learned("earthShock", toInt32(shaman.EarthShockLevel[:]))
	learned("flameShock", toInt32(shaman.FlameShockLevel[:]))
	learned("lightningShield", toInt32(shaman.LightningShieldLevel[:]))
	learned("strengthOfEarth", toInt32(shaman.StrengthOfEarthTotemLevel[:]))
	learned("searingTotem", toInt32(shaman.SearingTotemLevel[:]))
	learned("magmaTotem", toInt32(shaman.MagmaTotemLevel[:]))
	learned("flametongueTotem", toInt32(shaman.FlametongueTotemLevel[:]))
	learned("manaSpring", toInt32(shaman.ManaSpringTotemLevel[:]))
	learned("fireNova", toInt32(shaman.FireNovaLevel[:]))
	learned("chainLightning", toInt32(shaman.ChainLightningLevel[:]))
	learned("windfuryTotem", toInt32(shaman.WindfuryTotemLevel[:]))
	learned("graceOfAir", toInt32(shaman.GraceOfAirTotemLevel[:]))
	learned("rockbiterWeapon", shaman.RockbiterWeaponLevel[:])
	learned("flametongueWeapon", shaman.FlametongueWeaponLevel[:])
	learned("windfuryWeapon", shaman.WindfuryWeaponLevel[:])
	if t.Level >= 20 {
		out = append(out, "totemicRecall")
	}
	for _, talent := range []string{"stormstrike", "maelstromWeapon"} {
		if hasTalent(talent) {
			out = append(out, talent)
		}
	}
	return out
}

// NewRank is whether a spell the rotation casts gets a new rank at the level, past its first.
// A new rank costs more mana, so the search tunes the rest and mana thresholds again there.
func (t ShamanEnhLeveling) NewRank() bool {
	tables := [][]int32{
		shaman.HealRankLevels(shaman.HealingWaveRanks), shaman.HealRankLevels(shaman.LesserHealingWaveRanks),
		toInt32(shaman.EarthShockLevel[:]), toInt32(shaman.FlameShockLevel[:]), toInt32(shaman.LightningShieldLevel[:]),
		toInt32(shaman.LightningBoltLevel[:]), toInt32(shaman.ChainLightningLevel[:]), toInt32(shaman.FireNovaLevel[:]),
		toInt32(shaman.StrengthOfEarthTotemLevel[:]), toInt32(shaman.SearingTotemLevel[:]), toInt32(shaman.MagmaTotemLevel[:]),
		toInt32(shaman.FlametongueTotemLevel[:]), toInt32(shaman.ManaSpringTotemLevel[:]), toInt32(shaman.WindfuryTotemLevel[:]),
		toInt32(shaman.GraceOfAirTotemLevel[:]),
	}
	for _, levels := range tables {
		for rank := 2; rank < len(levels); rank++ {
			if levels[rank] == t.Level {
				return true
			}
		}
	}
	return false
}

func toInt32(levels []int) []int32 {
	out := make([]int32, len(levels))
	for i, l := range levels {
		out[i] = int32(l)
	}
	return out
}
