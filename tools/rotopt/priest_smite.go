package main

import (
	"fmt"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/priest"
)

// A Holy damage priest under Forever rules: Smite as the filler, Holy Fire kept on the
// target for Power in Light, Penance on cooldown, and the Shadow spells a priest without
// Shadowform still has (Shadow Word: Pain, Devouring Plague, Mind Blast, Shadow Word: Death)
// as knobs, since without Shadowform they cost full mana. Power Infusion and Inner Focus go
// with the 31 point Discipline build. Every line names the highest rank the player's level
// knows. Lines for spells the build or race cannot have only warn, and a knob at 0 drops it.
type priestSmite struct{}

func init() {
	templates["priest_smite"] = priestSmite{}
}

func (priestSmite) Knobs() []Knob {
	return []Knob{
		// Holy Fire whenever its damage over time is down (Power in Light needs it), above a
		// mana floor in percent. 2 casts it on cooldown like a nuke.
		{Name: "holyFire", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "holyFireMana", Default: 0, Min: 0, Max: 60, Step: 10},
		{Name: "penance", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "swp", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "swpMana", Default: 0, Min: 0, Max: 60, Step: 10},
		{Name: "dp", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "dpMana", Default: 20, Min: 0, Max: 80, Step: 10},
		{Name: "mb", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "mbMana", Default: 20, Min: 0, Max: 60, Step: 10},
		{Name: "swd", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "swdMana", Default: 20, Min: 0, Max: 60, Step: 10},
		// Inner Focus: 0 never, 1 on Holy Fire, 2 on Smite, 3 on Mind Blast.
		{Name: "innerFocus", Default: 2, Min: 0, Max: 3, Step: 1},
		{Name: "powerInfusion", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "chastise", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "darkSacrifice", Default: 40, Min: 0, Max: 80, Step: 10},
		// Smite of this rank (0 never) once the mana would no longer last the fight at
		// smiteManaRate mana a second. Ranks 3 and up keep the full coefficient under Forever,
		// so a lower rank keeps most of the damage for far less mana. Below 20 mana a second,
		// rank 3 is best at every fight length (2026-10-12), rank 4 only 0.9 behind at 300 sec.
		// Rank 3 is learned at level 14, and Classic cuts spells learned below 20, so the
		// default is rank 4, which no such rule reaches.
		{Name: "smiteDownrank", Default: 4, Min: 0, Max: 7, Step: 1},
		{Name: "smiteManaRate", Default: 20, Min: 15, Max: 45, Step: 5},
	}
}

func (priestSmite) Build(k Knobs) *proto.APLRotation {
	smite := rankID(priest.SmiteLevel[:], priest.SmiteSpellId[:])
	holyFire := rankID(priest.HolyFireLevel[:], priest.HolyFireSpellId[:])
	penance := rankID(priest.PenanceLevel[:], priest.PenanceSpellId[:])
	shadowWordPain := rankID(priest.ShadowWordPainLevel[:], priest.ShadowWordPainSpellId[:])
	devouringPlague := rankID(priest.DevouringPlagueLevel[:], priest.DevouringPlagueSpellId[:])
	mindBlast := rankID(priest.MindBlastLevel[:], priest.MindBlastSpellId[:])
	shadowWordDeath := rankID(priest.ShadowWordDeathLevel[:], priest.ShadowWordDeathSpellId[:])
	chastise := rankID(priest.ChastiseLevel[:], priest.ChastiseSpellId[:])
	darkSacrifice := rankID(priest.DarkSacrificeLevel[:], priest.DarkSacrificeSpellId[:])

	var items []*proto.APLListItem
	if k["darkSacrifice"] > 0 {
		items = append(items, cast(darkSacrifice, lt(currentManaPercent(), num(k["darkSacrifice"]/100)), "Undead. Dark Sacrifice when mana runs low. Other races can ignore the warning on this line."))
	}
	if k["powerInfusion"] == 1 {
		items = append(items, cast(powerInfusion, nil, "Power Infusion on cooldown. Without the talent ignore the warning on this line."))
	}
	items = append(items, autocastOtherCooldowns(nil, "Racials and trinkets on cooldown."))

	if k["holyFire"] == 1 {
		if k["innerFocus"] == 1 {
			items = append(items, cast(innerFocus, and(not(dotIsActive(holyFire)), manaAbove(k["holyFireMana"])), "Inner Focus right before Holy Fire. Without the talent ignore the warning on this line."))
		}
		items = append(items, cast(holyFire, and(not(dotIsActive(holyFire)), manaAbove(k["holyFireMana"])), "Holy Fire whenever its damage over time is down, for Power in Light."))
	}
	if k["penance"] == 1 {
		items = append(items, channel(penance, nil, nil, "Penance on cooldown. Without the talent ignore the warning on this line."))
	}
	if k["swp"] == 1 {
		items = append(items, cast(shadowWordPain, and(not(dotIsActive(shadowWordPain)), ge(remainingTime(), seconds(9)), manaAbove(k["swpMana"])), "Shadow Word: Pain whenever it is down."))
	}
	if k["dp"] == 1 {
		items = append(items, cast(devouringPlague, and(not(dotIsActive(devouringPlague)), ge(remainingTime(), seconds(12)), manaAbove(k["dpMana"])), "Devouring Plague whenever it is down and off cooldown."))
	}
	if k["chastise"] == 1 {
		items = append(items, cast(chastise, nil, "Dwarf. Chastise on cooldown against a Humanoid. Other races can ignore the warning on this line."))
	}
	if k["mb"] == 1 {
		if k["innerFocus"] == 3 {
			items = append(items, cast(innerFocus, and(spellIsReady(mindBlast), manaAbove(k["mbMana"])), "Inner Focus right before Mind Blast. Without the talent ignore the warning on this line."))
		}
		items = append(items, cast(mindBlast, manaAbove(k["mbMana"]), "Mind Blast on cooldown."))
	}
	if k["swd"] == 1 {
		items = append(items, cast(shadowWordDeath, manaAbove(k["swdMana"]), "Shadow Word: Death on cooldown."))
	}
	if k["innerFocus"] == 2 {
		items = append(items, cast(innerFocus, nil, "Inner Focus on cooldown, for the next Smite. Without the talent ignore the warning on this line."))
	}
	if rank := int(k["smiteDownrank"]); rank >= 3 && rank < priest.SmiteRanks {
		lasts := ge(&proto.APLValue{Value: &proto.APLValue_Math{Math: &proto.APLValueMath{
			Op:  proto.APLValueMath_OpDiv,
			Lhs: &proto.APLValue{Value: &proto.APLValue_CurrentMana{CurrentMana: &proto.APLValueCurrentMana{}}},
			Rhs: num(k["smiteManaRate"]),
		}}}, remainingTime())
		items = append(items, cast(smite, lasts, fmt.Sprintf("Smite while the mana lasts the fight at %.0f mana a second.", k["smiteManaRate"])))
		items = append(items, cast(spellID(priest.SmiteSpellId[rank]), nil, fmt.Sprintf("Rank %d Smite once it would not: most of the damage for far less mana.", rank)))
	} else {
		items = append(items, cast(smite, nil, "Smite fills."))
	}
	return rotation(items...)
}
