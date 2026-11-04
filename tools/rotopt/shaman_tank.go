package main

import (
	"fmt"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

// The Warden shaman, Forever's shaman tank (docs/tank_shaman.md): Rockbiter Weapon on a one-hander
// and a shield, Spirit Weapons for the parry and 1.302 times the threat. It's searched with -tank on
// the tank encounters (docs/tank_sim.md, The optimizer's objective).
//
// Its threat is the white swings with Rockbiter's 72 threat a second of weapon speed, Stormstrike
// (reset by our dodges and parries with Improved Stormstrike), Earth Shock (twice its damage in
// threat) and Lightning Shield's orbs. What a tank decides: which totems (Stoneskin Totem's 30 before
// armor for us, or Strength of Earth's block value when the raid has none), which shield (Lightning
// Shield's threat or Water Shield's mana), which shock and how much mana to keep for it, and whether
// to put a GCD into Lesser Healing Wave. While the shaman casts it can't dodge, parry or block, so a
// heal costs avoidance as well as mana.
//
// Nothing here reads the tank's health: a scoring run's replay can't follow it (docs/tank_sim.md).
// The emergency heal (Nature's Swiftness and Healing Wave at low health) is in the UI's rotation only.
type shamanTank struct{}

func init() {
	templates["shaman_tank"] = shamanTank{}
}

func (shamanTank) Knobs() []Knob {
	return []Knob{
		// 0 none, 1 Stoneskin Totem, 2 Strength of Earth Totem. The tank bench's raid has another
		// shaman's Strength of Earth, Grace of Air and Windfury Totem (raid-wide since 70425), so
		// ours would only stand in for theirs.
		{Name: "earthTotem", Default: 1, Min: 0, Max: 2, Step: 1},
		// 0 none, 1 Grace of Air Totem, 2 Windfury Totem.
		{Name: "airTotem", Default: 0, Min: 0, Max: 2, Step: 1},
		// 0 none, 1 Searing Totem, 2 Magma Totem, 3 Flametongue Totem. A totem's own damage makes the
		// totem threat, not us, so only Fire Nova (below) and Flametongue Totem's buff feed our threat.
		{Name: "fireTotem", Default: 0, Min: 0, Max: 3, Step: 1},
		// 0 none, 1 Mana Spring Totem, 2 Healing Stream Totem (its heals count as our own, fixed in
		// size, in the replay).
		{Name: "waterTotem", Default: 0, Min: 0, Max: 2, Step: 1},
		// 0 none, 1 Lightning Shield, 2 Water Shield (the Restoration talent).
		{Name: "shield", Default: 1, Min: 0, Max: 2, Step: 1},
		// The shield goes up again only below this much mana (Water Shield), 0 whenever it's down.
		{Name: "shieldMana", Default: 0, Min: 0, Max: 80, Step: 20},
		// 0 none, 1 Earth Shock, 2 Flame Shock, 3 Frost Shock, 4 Flame Shock while its DoT is down and
		// Earth Shock otherwise.
		{Name: "shock", Default: 1, Min: 0, Max: 4, Step: 1},
		// Shocks only above this much mana.
		{Name: "shockMana", Default: 0, Min: 0, Max: 60, Step: 10},
		// 1 holds Earth Shock while Stormstrike's +20% is coming within 2 sec, so the shock takes it.
		{Name: "earthShockHold", Default: 0, Min: 0, Max: 1, Step: 1},
		// 1: the shock line above Stormstrike (it makes more threat a GCD); 0: below it, as before 2026-11-04.
		{Name: "shockFirst", Default: 1, Min: 0, Max: 1, Step: 1},
		// Fire Nova (a shaman spell, our threat) around the fire totem: 0 never, 1 after the shock.
		{Name: "fireNova", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "fireNovaMana", Default: 40, Min: 0, Max: 80, Step: 20},
		// Lightning Bolt at this many Maelstrom Weapon stacks (5 is instant), 0 never.
		{Name: "maelstromStacks", Default: 0, Min: 0, Max: 5, Step: 1},
		// Lesser Healing Wave on ourselves whenever the mana is over this, 0 never. It reads no health.
		{Name: "selfHealMana", Default: 0, Min: 0, Max: 90, Step: 10},
		// 1 uses the synthetic defensive on-use trinket (1,665 armor for 20 sec every 2 min) on cooldown. A
		// defensive cooldown isn't autocast (that reads health), so it has its own line.
		{Name: "armorTrinket", Default: 1, Min: 0, Max: 1, Step: 1},
		// With armorTrinket 1: 0 uses it on cooldown, else holds it for the last trinketAt seconds when the
		// wait costs no use (tank_rotations-20261104-9, as the warrior's trinketAt).
		{Name: "trinketAt", Default: 0, Min: 0, Max: 60, Step: 5},
		// 1 puts the totems and the shield down before the pull, 0 after it.
		{Name: "prepull", Default: 1, Min: 0, Max: 1, Step: 1},
		// No new fire or water totem with less than this long to go.
		{Name: "totemMinTime", Default: 10, Min: 0, Max: 30, Step: 10},
	}
}

func (shamanTank) Build(k Knobs) *proto.APLRotation {
	soeLevels, soeIDs := shaman.StrengthOfEarthTotemLevel[:], shaman.StrengthOfEarthTotemSpellId[:]
	goaLevels, goaIDs := shaman.GraceOfAirTotemLevel[:], shaman.GraceOfAirTotemSpellId[:]
	strengthOfEarth := rankID(soeLevels[:len(soeLevels)-1], soeIDs[:len(soeIDs)-1])
	graceOfAir := rankID(goaLevels[:len(goaLevels)-1], goaIDs[:len(goaIDs)-1])
	stoneskin := rankID(shaman.StoneskinTotemLevel[:], shaman.StoneskinTotemSpellId[:])
	windfuryTotem := rankID(shaman.WindfuryTotemLevel[:], shaman.WindfuryTotemSpellId[:])
	searingTotem := rankID(shaman.SearingTotemLevel[:], shaman.SearingTotemSpellId[:])
	magmaTotem := rankID(shaman.MagmaTotemLevel[:], shaman.MagmaTotemSpellId[:])
	flametongueTotem := rankID(shaman.FlametongueTotemLevel[:], shaman.FlametongueTotemSpellId[:])
	fireNova := rankID(shaman.FireNovaLevel[:], shaman.FireNovaSpellId[:])
	manaSpringTotem := rankID(shaman.ManaSpringTotemLevel[:], shaman.ManaSpringTotemSpellId[:])
	healingStreamTotem := rankID(shaman.HealingStreamTotemLevel[:], shaman.HealingStreamTotemSpellId[:])
	lightningShield := rankID(shaman.LightningShieldLevel[:], shaman.LightningShieldSpellId[:])
	earthShock := rankID(shaman.EarthShockLevel[:], shaman.EarthShockSpellId[:])
	flameShock := rankID(shaman.FlameShockLevel[:], shaman.FlameShockSpellId[:])
	frostShock := rankID(shaman.FrostShockLevel[:], shaman.FrostShockSpellId[:])
	lightningBolt := rankID(shaman.LightningBoltLevel[:], shaman.LightningBoltSpellId[:])
	lesserHealingWave := spellID(shaman.TopHealRankID(shaman.LesserHealingWaveRanks, playerLevel))

	manaFloor := func(pct float64) value {
		if pct == 0 {
			return nil
		}
		return ge(currentManaPercent(), num(pct/100))
	}
	totemDown := func(slot proto.ShamanTotems_TotemType, minTime bool) value {
		when := le(totemRemainingTime(slot), seconds(0))
		if minTime && k["totemMinTime"] > 0 {
			when = and(when, ge(remainingTime(), seconds(k["totemMinTime"])))
		}
		return when
	}

	var items []*proto.APLListItem
	var pre []*proto.ActionID // the prepull casts, in order

	// The shield first: Lightning Shield's orbs are threat, and it's gone after 3 of the boss's hits.
	var shieldID *proto.ActionID
	switch k["shield"] {
	case 1:
		shieldID = lightningShield
	case 2:
		shieldID = waterShield
	}
	if shieldID != nil {
		pre = append(pre, shieldID)
		note := "Lightning Shield whenever its orbs are gone: their threat beats the 370 mana."
		if k["shield"] == 2 {
			note = "Water Shield whenever its globes are gone (it needs the Restoration talent)."
		}
		items = append(items, cast(shieldID, and(not(auraIsActive(shieldID)), lowManaOnly(k["shieldMana"])), note))
	}

	var earthID *proto.ActionID
	switch k["earthTotem"] {
	case 1:
		earthID = stoneskin
	case 2:
		earthID = strengthOfEarth
	}
	if earthID != nil {
		pre = append(pre, earthID)
		items = append(items, cast(earthID, totemDown(proto.ShamanTotems_Earth, false), "The earth totem whenever it's down."))
	}
	var airID *proto.ActionID
	switch k["airTotem"] {
	case 1:
		airID = graceOfAir
	case 2:
		airID = windfuryTotem
	}
	if airID != nil {
		pre = append(pre, airID)
		items = append(items, cast(airID, totemDown(proto.ShamanTotems_Air, false), "The air totem whenever it's down."))
	}

	if k["armorTrinket"] == 1 {
		trinket := itemID(shaman.WardenSyntheticOnUseTrinket)
		var when value
		note := "The defensive on-use trinket on cooldown, on a timer. Without it ignore the warning on this line."
		if k["trinketAt"] > 0 {
			note = fmt.Sprintf("The defensive on-use trinket, held for the last %d sec unless the wait costs a use. Without it ignore the warning on this line.", int(k["trinketAt"]))
			when = heldForLast(trinket, k["trinketAt"], 20)
		}
		items = append(items, cast(trinket, when, note))
	}
	items = append(items,
		autocastOtherCooldowns(nil, "Rage of the Farseer, racials, trinkets on cooldown."),
	)
	// Stormstrike goes below the shock by default: Earth Shock makes about 2,225 threat a GCD against
	// Stormstrike's 1,500, so when both are ready the shock goes first (tank_rotations-20261104-4).
	ssItem := cast(stormstrike, nil, "Stormstrike on cooldown: our dodges and parries reset it with Improved Stormstrike.")
	if k["shockFirst"] == 0 {
		items = append(items, ssItem)
	}

	when := manaFloor(k["shockMana"])
	esWhen := when
	if k["earthShockHold"] == 1 {
		esWhen = and(esWhen, or(targetAuraIsActive(stormstrike), gt(spellTimeToReady(stormstrike), seconds(2))))
	}
	switch k["shock"] {
	case 1:
		items = append(items, cast(earthShock, esWhen, "Earth Shock: twice its damage in threat."))
	case 2:
		items = append(items, cast(flameShock, when, "Flame Shock."))
	case 3:
		items = append(items, cast(frostShock, when, "Frost Shock."))
	case 4:
		items = append(items,
			cast(flameShock, and(not(dotIsActive(flameShock)), when), "Flame Shock while its DoT is down."),
			cast(earthShock, esWhen, "Earth Shock otherwise."))
	}
	if k["shockFirst"] == 1 {
		items = append(items, ssItem)
	}

	var fireID *proto.ActionID
	switch k["fireTotem"] {
	case 1:
		fireID = searingTotem
	case 2:
		fireID = magmaTotem
	case 3:
		fireID = flametongueTotem
	}
	if fireID != nil {
		pre = append(pre, fireID)
		items = append(items, cast(fireID, totemDown(proto.ShamanTotems_Fire, true), "The fire totem whenever it's down."))
		if k["fireNova"] == 1 {
			items = append(items, cast(fireNova, manaFloor(k["fireNovaMana"]), "Fire Nova around the fire totem: our own threat."))
		}
	}
	var waterID *proto.ActionID
	switch k["waterTotem"] {
	case 1:
		waterID = manaSpringTotem
	case 2:
		waterID = healingStreamTotem
	}
	if waterID != nil {
		pre = append(pre, waterID)
		items = append(items, cast(waterID, totemDown(proto.ShamanTotems_Water, true), "The water totem whenever it's down."))
	}

	if k["maelstromStacks"] > 0 {
		items = append(items, cast(lightningBolt, ge(auraNumStacks(maelstromWeapon), num(k["maelstromStacks"])),
			"Lightning Bolt at enough Maelstrom Weapon stacks. Without the talent ignore the warning on this line."))
	}
	if k["selfHealMana"] > 0 {
		items = append(items, cast(lesserHealingWave, manaFloor(k["selfHealMana"]),
			"Lesser Healing Wave on ourselves with the mana to spare (it reads no health; we can't dodge, parry or block while casting)."))
	}

	rot := rotation(items...)
	if k["prepull"] == 1 {
		// One GCD apart (totems 1 sec, the shield 1.5), the last one at the pull.
		at := 0.0
		for i := len(pre) - 1; i >= 0; i-- {
			at += 1.5
			rot.PrepullActions = append([]*proto.APLPrepullAction{prepull(pre[i], at)}, rot.PrepullActions...)
		}
	}
	return rot
}

// lowManaOnly is a condition that the mana is under pct, nil (always) at 0.
func lowManaOnly(pct float64) value {
	if pct == 0 {
		return nil
	}
	return lt(currentManaPercent(), num(pct/100))
}
