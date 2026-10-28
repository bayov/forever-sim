package main

import (
	"github.com/wowsims/classic/sim/core/proto"
)

// The warlock under Forever rules, one template for the three specs. A bane and a curse are
// two slots now, so the bane line (Agony or Doom) and the curse line (the warlock's own
// Curse of the Elements or Recklessness, for when the raid has none) are separate knobs. The
// damage over time spells go up when they are down and there is time left for most of their
// ticks, then the procs and cooldowns, then the filler. Lines for spells the build cannot
// have (Conflagrate, Shadowburn, Incinerate and Wrack are talents) only warn, and a knob at 0
// drops its line. The demon, Demonic Sacrifice and the weapon imbue are WarlockOptions, not
// knobs.
//
// Every line names the top rank at level 60 that Forever teaches (no Ahn'Qiraj book ranks:
// Shadow Bolt IX, Immolate VII, Corruption VI).
type warlockRotation struct{}

func init() {
	templates["warlock"] = warlockRotation{}
}

var (
	lockShadowBolt   = spellID(11661)
	lockIncinerate   = spellID(1293813)
	lockSearingPain  = spellID(17923)
	lockImmolate     = spellID(11668)
	lockCorruption   = spellID(11672)
	lockBaneOfAgony  = spellID(11713)
	lockBaneOfDoom   = spellID(603)
	lockSiphonLife   = spellID(18881)
	lockConflagrate  = spellID(18932)
	lockShadowburn   = spellID(18871)
	lockSoulFire     = spellID(17924)
	lockDeathCoil    = spellID(17926)
	lockWrack        = spellID(1316697)
	lockDrainLife    = spellID(11700)
	lockDrainSoul    = spellID(11675)
	lockLifeTap      = spellID(11689)
	lockAmplifyCurse = spellID(18288)
	lockSoulLink     = spellID(19028)
	// The synthetic on-use trinket, tools/database/forever_synthetic_items_warlock.go.
	lockSyntheticTalisman = itemID(997022)
	lockCurseOfElement    = spellID(1311680)
	lockCurseOfReckl      = spellID(11717)
	lockDecimation        = spellID(440873)
	lockShadowTrance      = spellID(17941)
	// Gnome Eureka! (sim/core/racials_forever.go): 10% less cost and 10% more damage for 8 sec,
	// 2 min. The warlock's own mask covers Shadow Bolt, Searing Pain, Soul Fire, Conflagrate and
	// Immolate's hit, and not Incinerate or the damage over time spells.
	lockEureka = spellID(1259812)
)

// gcdIsReady is true when the global cooldown is free.
func gcdIsReady() value {
	return &proto.APLValue{Value: &proto.APLValue_GcdIsReady{GcdIsReady: &proto.APLValueGCDIsReady{}}}
}

func (warlockRotation) Knobs() []Knob {
	return []Knob{
		// 0 no bane, 1 Bane of Agony, 2 Bane of Doom with Bane of Agony while Doom is on
		// cooldown or the fight ends before Doom would land.
		{Name: "bane", Default: 2, Min: 0, Max: 2, Step: 1},
		// Amplify Curse ahead of each Bane of Agony (the Affliction talent, 3 min cooldown).
		{Name: "amplifyCurse", Default: 0, Min: 0, Max: 1, Step: 1},
		// The curse slot (separate from the bane under Forever): 0 none, the raid has its curses
		// from another warlock; 1 Curse of the Elements and 2 Curse of Recklessness when this
		// warlock is on duty for it (standards 23, the duty presets).
		{Name: "curse", Default: 0, Min: 0, Max: 2, Step: 1},
		{Name: "corruption", Default: 1, Min: 0, Max: 1, Step: 1},
		// A new Corruption only with this many seconds left in the fight (its 6 ticks take 18).
		{Name: "corruptionMinTime", Default: 9, Min: 3, Max: 18, Step: 3},
		{Name: "immolate", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "immolateMinTime", Default: 6, Min: 3, Max: 15, Step: 3},
		// Recast Corruption and Immolate this many seconds before they run out, so the new one
		// lands as the old one ends instead of after the filler that was being cast (0 waits for
		// them to fall off). About their cast time: 1.6 sec for Corruption with one point of
		// Improved Corruption and 1.5 for Immolate with Bane 5/5.
		{Name: "dotLead", Default: 0, Min: 0, Max: 2.5, Step: 0.5},
		// 1 recasts Immolate once it has less left than its own cast time, which clips no tick
		// (an early Corruption, an instant, clips its last one). 0 leaves Immolate on dotLead.
		{Name: "immolateLead", Default: 0, Min: 0, Max: 1, Step: 1},
		// Siphon Life (the Affliction talent) when its 30 sec are down.
		{Name: "siphonLife", Default: 0, Min: 0, Max: 1, Step: 1},
		// Conflagrate on cooldown while Immolate is up.
		{Name: "conflagrate", Default: 0, Min: 0, Max: 1, Step: 1},
		// Shadowburn: 0 never, 1 on cooldown, 2 only in the last 20% of the boss's health.
		// Each costs a Soul Shard, which the sim does not count.
		{Name: "shadowburn", Default: 0, Min: 0, Max: 2, Step: 1},
		// Soul Fire: 0 never, 1 on cooldown, 2 only while Decimation (the Demonology talent)
		// cuts its cast.
		{Name: "soulFire", Default: 0, Min: 0, Max: 2, Step: 1},
		// Death Coil on cooldown.
		{Name: "deathCoil", Default: 0, Min: 0, Max: 1, Step: 1},
		// At the very end, when the filler would not land before the fight is over: Death Coil, then
		// Searing Pain if it still lands.
		{Name: "endInstants", Default: 0, Min: 0, Max: 1, Step: 1},
		// Wrack (the Affliction capstone): 0 never, 1 while Corruption and the bane are both up, 2
		// only while Corruption and Bane of Agony are up (its 10% counts on those two).
		{Name: "wrack", Default: 0, Min: 0, Max: 2, Step: 1},
		// The filler: 1 Shadow Bolt, 2 Incinerate while Immolate is up and Shadow Bolt
		// otherwise, 3 Searing Pain, 4 Drain Life, 5 Drain Soul.
		{Name: "filler", Default: 1, Min: 1, Max: 5, Step: 1},
		// Drain Soul in place of the filler below this much of the boss's health: 0 never, 1
		// below 25%, 2 below 20%. Soul Siphon and Improved Drains make it the Affliction execute.
		{Name: "drainSoulExecute", Default: 0, Min: 0, Max: 2, Step: 1},
		// Life Tap below this much mana.
		{Name: "lifeTapMana", Default: 20, Min: 5, Max: 50, Step: 5},
		// Life Tap anyway in the last seconds when the mana is no use any more: 0 never.
		{Name: "lifeTapEnd", Default: 0, Min: 0, Max: 1, Step: 1},
		// Soul Link before the pull (the Demonology talent).
		{Name: "soulLink", Default: 0, Min: 0, Max: 1, Step: 1},
		// Life Tap before the pull, to pay back Soul Link's mana.
		{Name: "prepullLifeTap", Default: 0, Min: 0, Max: 1, Step: 1},
		// How many seconds before the pull the opening cast starts. Damage that lands before the
		// pull does not count, so it should land just after: 2.5 for a 2.45 sec Shadow Bolt and
		// its travel, 1.5 for Immolate.
		{Name: "prepullAt", Default: 2.5, Min: 1, Max: 4, Step: 0.5},
		// Gnome Eureka!: 0 with the other cooldowns (at the pull and on cooldown, where its 8 sec
		// go on the bane, Corruption and Life Tap), 1 just before a filler with the GCD free, so
		// its 8 sec are Shadow Bolts, 2 just before an Immolate that is due with Conflagrate back
		// within 5 sec, so both hit inside it (the Incinerate build, whose filler it doesn't
		// cover). Other races ignore the warning on the line.
		{Name: "eureka", Default: 0, Min: 0, Max: 2, Step: 1},
	}
}

func (warlockRotation) Build(k Knobs) *proto.APLRotation {
	var items []*proto.APLListItem
	items = append(items, autocastOtherCooldowns(nil, "Racials and trinkets on cooldown."))

	// The filler can cost more than the Life Tap threshold, so tap also when it cannot be cast
	// (otherwise the warlock waits for regen).
	filler := lockShadowBolt
	switch k["filler"] {
	case 2:
		filler = lockIncinerate
	case 3:
		filler = lockSearingPain
	case 4:
		filler = lockDrainLife
	case 5:
		filler = lockDrainSoul
	}
	items = append(items, cast(lockLifeTap, or(lt(currentManaPercent(), num(k["lifeTapMana"]/100)), not(spellCanCast(filler))),
		"Life Tap when the mana runs low or the filler costs more than is left."))

	// The raid's curse when this warlock is the one on duty for it (standards 23): put up right
	// after the opener and recast before it drops, ahead of the bane.
	switch k["curse"] {
	case 1:
		items = append(items, cast(lockCurseOfElement, and(lt(targetAuraRemainingTime(lockCurseOfElement), seconds(1.5)), gt(remainingTime(), seconds(2))),
			"Curse of the Elements on duty for the raid, recast before it drops."))
	case 2:
		items = append(items, cast(lockCurseOfReckl, and(lt(targetAuraRemainingTime(lockCurseOfReckl), seconds(1.5)), gt(remainingTime(), seconds(2))),
			"Curse of Recklessness on duty for the raid, recast before it drops."))
	}
	switch k["bane"] {
	case 1:
		if k["amplifyCurse"] == 1 {
			items = append(items, cast(lockAmplifyCurse, not(dotIsActive(lockBaneOfAgony)), "Amplify Curse ahead of Bane of Agony."))
		}
		items = append(items, cast(lockBaneOfAgony, and(not(dotIsActive(lockBaneOfAgony)), gt(remainingTime(), seconds(12))), "Bane of Agony when it is down."))
	case 2:
		items = append(items,
			cast(lockBaneOfDoom, and(not(dotIsActive(lockBaneOfAgony)), gt(remainingTime(), seconds(60))), "Bane of Doom when it lands before the fight ends."),
			cast(lockBaneOfAgony, and(not(dotIsActive(lockBaneOfDoom)), not(dotIsActive(lockBaneOfAgony)), gt(remainingTime(), seconds(12)),
				or(gt(spellTimeToReady(lockBaneOfDoom), seconds(12)), le(remainingTime(), seconds(60)))),
				"Bane of Agony while Bane of Doom is on cooldown or would not land."),
		)
	}
	if k["corruption"] == 1 {
		items = append(items, cast(lockCorruption, and(dotDown(lockCorruption, k["dotLead"]), gt(remainingTime(), seconds(k["corruptionMinTime"]))), "Corruption when it is down."))
	}
	if k["siphonLife"] == 1 {
		items = append(items, cast(lockSiphonLife, and(not(dotIsActive(lockSiphonLife)), gt(remainingTime(), seconds(9))), "Siphon Life when it is down."))
	}
	if k["immolate"] == 1 {
		immolateDown := dotDown(lockImmolate, k["dotLead"])
		if k["immolateLead"] == 1 {
			immolateDown = or(not(dotIsActive(lockImmolate)), lt(dotRemainingTime(lockImmolate), spellCastTime(lockImmolate)))
		}
		immolateWhen := and(immolateDown, gt(remainingTime(), seconds(k["immolateMinTime"])))
		if k["eureka"] == 2 {
			items = append(items, cast(lockEureka, and(gcdIsReady(), immolateWhen, lt(spellTimeToReady(lockConflagrate), seconds(5))),
				"Gnome. Eureka! for an Immolate with Conflagrate after it. Other races ignore the warning on this line."))
		}
		items = append(items, cast(lockImmolate, immolateWhen, "Immolate when it is down."))
	}
	if k["conflagrate"] == 1 {
		items = append(items, cast(lockConflagrate, dotIsActive(lockImmolate), "Conflagrate on cooldown."))
	}
	switch k["shadowburn"] {
	case 1:
		items = append(items, cast(lockShadowburn, nil, "Shadowburn on cooldown."))
	case 2:
		items = append(items, cast(lockShadowburn, isExecutePhase(proto.APLValueIsExecutePhase_E20), "Shadowburn in the last 20% of the boss."))
	}
	if k["deathCoil"] == 1 {
		items = append(items, cast(lockDeathCoil, nil, "Death Coil on cooldown."))
	}
	switch k["soulFire"] {
	case 1:
		items = append(items, cast(lockSoulFire, nil, "Soul Fire on cooldown."))
	case 2:
		items = append(items, cast(lockSoulFire, auraIsActive(lockDecimation), "Soul Fire while Decimation cuts its cast."))
	}
	switch k["wrack"] {
	case 1:
		items = append(items, cast(lockWrack, and(dotIsActive(lockCorruption), or(dotIsActive(lockBaneOfAgony), dotIsActive(lockBaneOfDoom))), "Wrack while Corruption and the bane are up."))
	case 2:
		items = append(items, cast(lockWrack, and(dotIsActive(lockCorruption), dotIsActive(lockBaneOfAgony)), "Wrack while Corruption and Bane of Agony are up."))
	}
	if k["lifeTapEnd"] == 1 {
		items = append(items, cast(lockLifeTap, and(lt(remainingTime(), seconds(3)), lt(currentManaPercent(), num(0.5))), "Life Tap at the very end is free."))
	}
	switch k["drainSoulExecute"] {
	case 1:
		items = append(items, cast(lockDrainSoul, isExecutePhase(proto.APLValueIsExecutePhase_E25), "Drain Soul in the last 25% of the boss."))
	case 2:
		items = append(items, cast(lockDrainSoul, isExecutePhase(proto.APLValueIsExecutePhase_E20), "Drain Soul in the last 20% of the boss."))
	}
	if k["endInstants"] == 1 {
		fillerCast := lockShadowBolt
		if k["filler"] == 2 {
			fillerCast = lockIncinerate
		}
		tooLate := lt(remainingTime(), add(spellCastTime(fillerCast), seconds(0.25)))
		items = append(items,
			cast(lockDeathCoil, tooLate, "Death Coil when the filler would not land before the end."),
			cast(lockSearingPain, and(tooLate, gt(remainingTime(), spellCastTime(lockSearingPain))),
				"Searing Pain when the filler would not land before the end."))
	}
	if k["eureka"] == 1 {
		items = append(items, cast(lockEureka, gcdIsReady(),
			"Gnome. Eureka! just before a filler, so its 8 sec are fillers. Other races ignore the warning on this line."))
	}
	switch k["filler"] {
	case 4:
		items = append(items, cast(lockDrainLife, nil, "Drain Life as the filler."))
	case 5:
		items = append(items, cast(lockDrainSoul, nil, "Drain Soul as the filler."))
	case 2:
		items = append(items,
			cast(lockIncinerate, dotIsActive(lockImmolate), "Incinerate while Immolate is up."),
			cast(lockShadowBolt, nil, "Shadow Bolt otherwise."))
	case 3:
		items = append(items, cast(lockSearingPain, nil, "Searing Pain as the filler."))
	default:
		items = append(items, cast(lockShadowBolt, nil, "Shadow Bolt as the filler."))
	}

	rot := rotation(items...)
	if k["soulLink"] == 1 {
		at := 4.5
		if k["prepullLifeTap"] == 1 {
			at = 6
		}
		rot.PrepullActions = append(rot.PrepullActions, prepull(lockSoulLink, at))
	}
	if k["prepullLifeTap"] == 1 {
		rot.PrepullActions = append(rot.PrepullActions, prepull(lockLifeTap, 4.5))
	}
	switch k["filler"] {
	case 2:
		// Incinerate wants Immolate on the target first.
		rot.PrepullActions = append(rot.PrepullActions, prepull(lockImmolate, k["prepullAt"]))
	case 3:
		rot.PrepullActions = append(rot.PrepullActions, prepull(lockSearingPain, k["prepullAt"]))
	default:
		rot.PrepullActions = append(rot.PrepullActions, prepull(lockShadowBolt, k["prepullAt"]))
	}
	// The synthetic on-use trinket of the Level 60 sets (140 spell power, 20 sec, 2 min), used just
	// before the opener lands so it counts too. Without the trinket the action is dropped.
	rot.PrepullActions = append(rot.PrepullActions, prepull(lockSyntheticTalisman, 0.1))
	return rot
}

// dotDown is true when the damage over time spell is off the target, or has less than lead
// seconds left.
func dotDown(spell *proto.ActionID, lead float64) value {
	if lead == 0 {
		return not(dotIsActive(spell))
	}
	return or(not(dotIsActive(spell)), lt(dotRemainingTime(spell), seconds(lead)))
}

func isExecutePhase(threshold proto.APLValueIsExecutePhase_ExecutePhaseThreshold) value {
	return &proto.APLValue{Value: &proto.APLValue_IsExecutePhase{IsExecutePhase: &proto.APLValueIsExecutePhase{Threshold: threshold}}}
}

func spellCanCast(spell *proto.ActionID) value {
	return &proto.APLValue{Value: &proto.APLValue_SpellCanCast{SpellCanCast: &proto.APLValueSpellCanCast{SpellId: spell}}}
}

func spellCastTime(spell *proto.ActionID) value {
	return &proto.APLValue{Value: &proto.APLValue_SpellCastTime{SpellCastTime: &proto.APLValueSpellCastTime{SpellId: spell}}}
}
