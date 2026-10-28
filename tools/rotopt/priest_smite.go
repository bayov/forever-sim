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
		// Holy Fire whenever its damage over time is down, above a mana floor in percent. 2
		// casts it on cooldown like a nuke. Off since Power in Light also counts Shadow Word:
		// Pain (70425): its mana goes to Smite, 744.0 / 724.6 / 639.7 at 120 / 180 / 300 sec
		// against 729.3 / 716.8 / 646.6 with it (2026-10-16).
		{Name: "holyFire", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "holyFireMana", Default: 0, Min: 0, Max: 60, Step: 10},
		{Name: "penance", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "swp", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "swpMana", Default: 0, Min: 0, Max: 60, Step: 10},
		{Name: "dp", Default: 1, Min: 0, Max: 1, Step: 1},
		// Devouring Plague above 80% mana. Its 24 sec cooldown (client 70425, was 1 min) would
		// spend Smite's mana: above 40% it is 733.2 / 709.5 / 617.9 at 120 / 180 / 300 sec,
		// above 60% with Inner Focus on cooldown 727.8 / 714.6 / 642.2 (2026-10-16). Without
		// Holy Fire (below) 80% is best at 180 sec.
		{Name: "dpMana", Default: 80, Min: 0, Max: 80, Step: 10},
		{Name: "mb", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "mbMana", Default: 20, Min: 0, Max: 60, Step: 10},
		{Name: "swd", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "swdMana", Default: 20, Min: 0, Max: 60, Step: 10},
		// Shadow Word: Death once a Smite would no longer land before the fight ends. With swd
		// off it is always ready then: +5.8 / +3.0 / +1.2 at 120 / 180 / 300 sec (rotation
		// audit 19b). Mind Blast there adds nothing on top.
		{Name: "swdEnd", Default: 1, Min: 0, Max: 1, Step: 1},
		// Inner Focus: 0 never, 1 on Holy Fire, 2 on Smite, 3 on Mind Blast, 4 on Devouring
		// Plague (free through the fourth class mask word, client 14751).
		{Name: "innerFocus", Default: 2, Min: 0, Max: 4, Step: 1},
		{Name: "powerInfusion", Default: 1, Min: 0, Max: 1, Step: 1},
		// Dwarf: Chastise on cooldown against a Humanoid (it hits nothing else). Its GCD is worth
		// more as Smite: a Dwarf on a Humanoid boss is 737.4 without it against 728.9 with it
		// at 180 sec (2026-10-28).
		{Name: "chastise", Default: 0, Min: 0, Max: 1, Step: 1},
		// Night Elf: Starshards on cooldown, 1 below Devouring Plague, 2 above Holy Fire. Its
		// 6 sec channel deals about 480 DPS against Smite's 700, and the mana it saves doesn't
		// make that up: the Night Elf rotation is 741.1 without it, 723.7 / 718.4 with it, and
		// a knob search started from 1 drops it (2026-10-28).
		{Name: "starshards", Default: 0, Min: 0, Max: 2, Step: 1},
		{Name: "darkSacrifice", Default: 40, Min: 0, Max: 80, Step: 10},
		// Smite of this rank (0 never) once the mana would no longer last the fight at
		// smiteManaRate mana a second. Forever's downranking penalty (Blizzard's 2026-10-01
		// development notes, foreverchanges.pro/downranking: 5% of the coefficient lost per
		// level past 17 below the caster, sim/priest/priest.go) leaves rank 7 (level 46) its
		// full coefficient at 60 and ranks 4 and lower none. Rank 7 below 35 mana a second is
		// 706.5 / 643.8 at 180 / 300 sec against 701.4 / 545.0 for the earlier rank 4 below 20.
		{Name: "smiteDownrank", Default: 7, Min: 0, Max: 7, Step: 1},
		{Name: "smiteManaRate", Default: 35, Min: 15, Max: 45, Step: 5},
		// Wand Shoot from wandLow percent mana up to wandHigh (0 never) with at least wandMinTime
		// seconds of fight left, to regenerate outside the five second rule, and wandFallback
		// when nothing else can be cast (out of mana). Shoot gives way to the next cast after
		// each shot (sim/core/wand.go). wowsims/forever's Smite line (Shoot at 10% or less) and
		// 20% to 40% lost to 10% to 20% (2026-10-23).
		{Name: "wandLow", Default: 10, Min: 0, Max: 30, Step: 5},
		{Name: "wandHigh", Default: 20, Min: 10, Max: 50, Step: 10},
		{Name: "wandMinTime", Default: 60, Min: 0, Max: 120, Step: 30},
		{Name: "wandFallback", Default: 1, Min: 0, Max: 1, Step: 1},
		// A Smite 2 sec before the pull (Divine Fury 5/5), landing at 0.
		{Name: "prepullSmite", Default: 1, Min: 0, Max: 1, Step: 1},
		// The Level 60 on-use trinket at -0.1 sec, so the pre-cast Smite landing at 0 gets it.
		{Name: "prepullTrinket", Default: 1, Min: 0, Max: 1, Step: 1},
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
	starshards := rankID(priest.StarshardsLevel[:], priest.StarshardsSpellId[:])
	darkSacrifice := rankID(priest.DarkSacrificeLevel[:], priest.DarkSacrificeSpellId[:])

	var items []*proto.APLListItem
	if k["darkSacrifice"] > 0 {
		items = append(items, cast(darkSacrifice, lt(currentManaPercent(), num(k["darkSacrifice"]/100)), "Undead. Dark Sacrifice when mana runs low. Other races can ignore the warning on this line."))
	}
	if k["powerInfusion"] == 1 {
		items = append(items, cast(powerInfusion, nil, "Power Infusion on cooldown. Without the talent ignore the warning on this line."))
	}
	items = append(items, autocastOtherCooldowns(nil, "Racials and trinkets on cooldown."))
	items = append(items, priestWandRegen(k)...)

	starshardsLine := channel(starshards, nil, nil, "Night Elf. Starshards on cooldown. Other races can ignore the warning on this line.")
	if k["starshards"] == 2 {
		items = append(items, starshardsLine)
	}
	if k["holyFire"] == 1 {
		hfCond := and(not(dotIsActive(holyFire)), priestManaAbove(k["holyFireMana"]))
		if k["innerFocus"] == 1 {
			items = append(items, cast(innerFocus, hfCond, "Inner Focus right before Holy Fire. Without the talent ignore the warning on this line."))
		}
		items = append(items, cast(holyFire, hfCond, "Holy Fire whenever its damage over time is down, for Power in Light."))
	}
	if k["penance"] == 1 {
		items = append(items, channel(penance, nil, nil, "Penance on cooldown. Without the talent ignore the warning on this line."))
	}
	if k["swp"] == 1 {
		items = append(items, cast(shadowWordPain, and(not(dotIsActive(shadowWordPain)), ge(remainingTime(), seconds(9)), priestManaAbove(k["swpMana"])), "Shadow Word: Pain whenever it is down."))
	}
	if k["dp"] == 1 {
		dpCond := and(not(dotIsActive(devouringPlague)), ge(remainingTime(), seconds(12)), priestManaAbove(k["dpMana"]))
		if k["innerFocus"] == 4 {
			// The lines above would take the free cast, so only while their spells are up.
			var above []value
			if k["holyFire"] == 1 {
				above = append(above, dotIsActive(holyFire))
			}
			if k["swp"] == 1 {
				above = append(above, dotIsActive(shadowWordPain))
			}
			items = append(items, cast(innerFocus, and(append([]value{dpCond, spellIsReady(devouringPlague)}, above...)...), "Inner Focus right before Devouring Plague. Without the talent ignore the warning on this line."))
		}
		items = append(items, cast(devouringPlague, dpCond, "Devouring Plague whenever it is down and off cooldown."))
	}
	if k["starshards"] == 1 {
		items = append(items, starshardsLine)
	}
	if k["chastise"] == 1 {
		items = append(items, cast(chastise, nil, "Dwarf. Chastise on cooldown against a Humanoid. Other races can ignore the warning on this line."))
	}
	if k["mb"] == 1 {
		if k["innerFocus"] == 3 {
			items = append(items, cast(innerFocus, and(spellIsReady(mindBlast), priestManaAbove(k["mbMana"])), "Inner Focus right before Mind Blast. Without the talent ignore the warning on this line."))
		}
		items = append(items, cast(mindBlast, priestManaAbove(k["mbMana"]), "Mind Blast on cooldown."))
	}
	if k["swd"] == 1 {
		items = append(items, cast(shadowWordDeath, priestManaAbove(k["swdMana"]), "Shadow Word: Death on cooldown."))
	}
	if k["swdEnd"] == 1 {
		smiteCastTime := &proto.APLValue{Value: &proto.APLValue_SpellCastTime{SpellCastTime: &proto.APLValueSpellCastTime{SpellId: smite}}}
		items = append(items, cast(shadowWordDeath, lt(remainingTime(), smiteCastTime), "Shadow Word: Death when Smite would not land before the fight ends."))
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
	items = append(items, priestWandFallback(k)...)
	rot := rotation(items...)
	if k["prepullSmite"] == 1 {
		rot.PrepullActions = append(rot.PrepullActions, prepull(smite, 2))
	}
	if k["prepullTrinket"] == 1 {
		rot.PrepullActions = append(rot.PrepullActions, prepull(itemID(priest.PriestSyntheticOnUseTrinket), 0.1))
	}
	return rot
}
