package main

import (
	"fmt"
	"strings"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/priest"
)

// Shadow under Forever rules. Shadowform goes up before the pull (40% of base mana). The
// damage over time spells come first, then the cooldown spells (Mind Blast, Shadow Word:
// Death, Starshards for a Night Elf, Chastise for a Dwarf), then a Troll's Shadowguard, and
// Mind Flay fills. Every line names the highest rank the player's level knows. Lines for
// spells the build or race cannot have only warn, and a knob at 0 drops its line.
//
// The real decisions are knobs: whether Devouring Plague and Shadow Word: Death are worth
// their mana, whether Mind Flay is clipped for a Mind Blast that comes off cooldown, where
// Starshards sits against Mind Blast, what Inner Focus is spent on, and mana floors for the
// expensive spells so a long fight does not run dry.
type priestShadow struct{}

func init() {
	templates["priest_shadow"] = priestShadow{}
}

var (
	innerFocus      = spellID(14751)
	shadowform      = spellID(15473)
	powerInfusion   = spellIDTag(10060, 0)
	vampiricEmbrace = spellID(15286)
	wandShoot       = spellID(5019)
)

// majorSpellblastingPotion (druid_balance.go) is the Major Spellblasting Potion's item id.

func (priestShadow) Knobs() []Knob {
	return []Knob{
		// Shadow Word: Pain when it is down, with at least this many seconds of fight left.
		{Name: "swpMinTime", Default: 9, Min: 0, Max: 18, Step: 3},
		// Devouring Plague when it is down and off its 24 sec cooldown, with at least this
		// many seconds of fight left and above a mana floor in percent.
		{Name: "dp", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "dpMinTime", Default: 12, Min: 0, Max: 24, Step: 6},
		{Name: "dpMana", Default: 0, Min: 0, Max: 60, Step: 10},
		// Shadow Word: Death on cooldown (18 sec), 2 for only in the execute phase.
		{Name: "swd", Default: 1, Min: 0, Max: 2, Step: 1},
		{Name: "swdMana", Default: 0, Min: 0, Max: 60, Step: 10},
		// Shadow Word: Death above Mind Blast in the priority. On since 70425 (18 sec cooldown,
		// 200% crits in Shadowform): +1.5 / +1.8 / +1.7 at 120 / 180 / 300 sec (2026-10-16).
		{Name: "swdFirst", Default: 1, Min: 0, Max: 1, Step: 1},
		// Mind Blast above a mana floor.
		{Name: "mbMana", Default: 0, Min: 0, Max: 40, Step: 10},
		// Inner Focus: 0 never, 1 on Mind Blast, 2 on cooldown with whatever comes next, 3 on
		// Devouring Plague (free through the fourth class mask word, client 14751).
		{Name: "innerFocus", Default: 1, Min: 0, Max: 3, Step: 1},
		// Clip Mind Flay for a ready Mind Blast once this many ticks are in (0 never). The
		// check runs after each tick and when the GCD comes back 1.5 sec in.
		{Name: "mfClip", Default: 0, Min: 0, Max: 2, Step: 1},
		// Clip Mind Flay once this many ticks are in when Shadow Word: Pain or Devouring Plague
		// fell off during it (0 never), so the damage over time spell goes back up sooner.
		{Name: "mfClipDots", Default: 2, Min: 0, Max: 2, Step: 1},
		// Night Elf: Starshards on cooldown, 1 below Mind Blast, 2 above it. Its ticks hit no
		// harder than Mind Flay's in Shadowform and its 6 sec channel holds up Mind Blast, so
		// it costs 30 DPS at any fight length (2026-10-12).
		{Name: "starshards", Default: 0, Min: 0, Max: 2, Step: 1},
		// Dwarf: Chastise on cooldown against a Humanoid. A GCD of Mind Flay is worth more,
		// 4.7 DPS at 180 sec.
		{Name: "chastise", Default: 0, Min: 0, Max: 1, Step: 1},
		// Undead: Dark Sacrifice under this much mana in percent (0 never). 10% since 70425:
		// Shadow + Holy now runs short (Devouring Plague every 24 sec) and casts it only when
		// it must, +2.9 / +5.8 / +2.1 at 120 / 180 / 300 sec against 40% (2026-10-16).
		{Name: "darkSacrifice", Default: 10, Min: 0, Max: 80, Step: 10},
		// Vampiric Embrace kept on the target. It heals the party for 20% of the priest's
		// Shadow damage and deals none, so each cast is a GCD of damage given up.
		{Name: "vampiricEmbrace", Default: 0, Min: 0, Max: 1, Step: 1},
		// Power Infusion on cooldown (needs the 31 point Discipline talent, which no level 60
		// build takes).
		{Name: "powerInfusion", Default: 0, Min: 0, Max: 1, Step: 1},
		// Troll: Shadowguard (3 charges, castable in Shadowform) before the pull, with Shadowform
		// a GCD earlier for it, and again before Mind Flay once it is down with at least
		// shadowguardRecast seconds left (0 never) and above shadowguardMana percent mana. The
		// raid's 2 hits a minute spend it. On the Level 60 build the pre-pull cast is +10.6 /
		// +7.9 / +4.9 at 120 / 180 / 300 sec and the recast +0.0 / +1.1 / +2.3 more. Shadow +
		// Holy runs short of mana at 300 sec, where the recast without the floor cost 3.6 (1.4
		// with it) for +1.0 at 180 sec (2026-10-23).
		{Name: "shadowguard", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "shadowguardRecast", Default: 60, Min: 0, Max: 100, Step: 20},
		{Name: "shadowguardMana", Default: 40, Min: 0, Max: 80, Step: 20},
		// Wand Shoot from wandLow percent mana up to wandHigh (0 never) with at least wandMinTime
		// seconds of fight left, to regenerate outside the five second rule, and wandFallback
		// when nothing else can be cast (out of mana). Shoot gives way to the next cast after
		// each shot (sim/core/wand.go). wowsims/forever's Smite line (Shoot at 10% or less) and
		// 20% to 40% lost to 10% to 20% (2026-10-23).
		{Name: "wandLow", Default: 10, Min: 0, Max: 30, Step: 5},
		{Name: "wandHigh", Default: 20, Min: 10, Max: 50, Step: 10},
		{Name: "wandMinTime", Default: 60, Min: 0, Max: 120, Step: 30},
		{Name: "wandFallback", Default: 1, Min: 0, Max: 1, Step: 1},
		// Before the pull: 0 Shadowform only (1.5 sec out), 1 Shadowform 3 sec out and a Mind
		// Blast 1.5 sec out that lands at 0, 2 the same with Inner Focus on that Mind Blast.
		{Name: "prepull", Default: 1, Min: 0, Max: 2, Step: 1},
		// The Level 60 on-use trinket (1) and also the Major Spellblasting Potion (2) at -0.1 sec,
		// so the pre-cast Mind Blast landing at 0 gets them (rotation auditor, 2026-10-13).
		{Name: "prepullCDs", Default: 2, Min: 0, Max: 2, Step: 1},
	}
}

func channel(spell *proto.ActionID, condition value, interruptIf value, notes string) *proto.APLListItem {
	return &proto.APLListItem{
		Notes: notes,
		Action: &proto.APLAction{
			Condition: condition,
			Action: &proto.APLAction_ChannelSpell{ChannelSpell: &proto.APLActionChannelSpell{
				SpellId:     spell,
				InterruptIf: interruptIf,
			}},
		},
	}
}

func priestManaAbove(percent float64) value {
	if percent <= 0 {
		return nil
	}
	return ge(currentManaPercent(), num(percent/100))
}

func (priestShadow) Build(k Knobs) *proto.APLRotation {
	shadowWordPain := rankID(priest.ShadowWordPainLevel[:], priest.ShadowWordPainSpellId[:])
	devouringPlague := rankID(priest.DevouringPlagueLevel[:], priest.DevouringPlagueSpellId[:])
	mindBlast := rankID(priest.MindBlastLevel[:], priest.MindBlastSpellId[:])
	mindFlay := rankID(priest.MindFlayLevel[:], priest.MindFlaySpellId[:])
	shadowWordDeath := rankID(priest.ShadowWordDeathLevel[:], priest.ShadowWordDeathSpellId[:])
	starshards := rankID(priest.StarshardsLevel[:], priest.StarshardsSpellId[:])
	chastise := rankID(priest.ChastiseLevel[:], priest.ChastiseSpellId[:])
	darkSacrifice := rankID(priest.DarkSacrificeLevel[:], priest.DarkSacrificeSpellId[:])
	shadowguard := rankID(priest.ShadowguardLevel[:], priest.ShadowguardSpellId[:])

	var items []*proto.APLListItem
	if k["darkSacrifice"] > 0 {
		items = append(items, cast(darkSacrifice, lt(currentManaPercent(), num(k["darkSacrifice"]/100)), "Undead. Dark Sacrifice when mana runs low. Other races can ignore the warning on this line."))
	}
	if k["powerInfusion"] == 1 {
		items = append(items, cast(powerInfusion, nil, "Power Infusion on cooldown. Without the talent ignore the warning on this line."))
	}
	items = append(items, autocastOtherCooldowns(nil, "Racials and trinkets on cooldown."))
	items = append(items, priestWandRegen(k)...)

	if k["vampiricEmbrace"] == 1 {
		items = append(items, cast(vampiricEmbrace, not(targetAuraIsActive(vampiricEmbrace)), "Vampiric Embrace whenever it is down, for the party's healing. Without the talent ignore the warning on this line."))
	}

	var swpTime value
	if k["swpMinTime"] > 0 {
		swpTime = ge(remainingTime(), seconds(k["swpMinTime"]))
	}
	items = append(items, cast(shadowWordPain, and(not(dotIsActive(shadowWordPain)), swpTime), "Shadow Word: Pain whenever it is down."))
	if k["dp"] == 1 {
		// Raiments of Conviction (5) take its whole cooldown, so it goes up whenever it is down.
		var dpTime value
		if k["dpMinTime"] > 0 {
			dpTime = ge(remainingTime(), seconds(k["dpMinTime"]))
		}
		dpCond := and(not(dotIsActive(devouringPlague)), dpTime, priestManaAbove(k["dpMana"]))
		if k["innerFocus"] == 3 {
			// Shadow Word: Pain above would take the free cast, so only while it is up.
			items = append(items, cast(innerFocus, and(dpCond, spellIsReady(devouringPlague), dotIsActive(shadowWordPain)), "Inner Focus right before Devouring Plague. Without the talent ignore the warning on this line."))
		}
		items = append(items, cast(devouringPlague, dpCond, "Devouring Plague whenever it is down and off cooldown."))
	}

	var swdLine *proto.APLListItem
	if k["swd"] > 0 {
		var cond value = priestManaAbove(k["swdMana"])
		notes := "Shadow Word: Death on cooldown."
		if k["swd"] == 2 {
			cond = and(cond, isExecutePhase20())
			notes = "Shadow Word: Death in the execute phase."
		}
		swdLine = cast(shadowWordDeath, cond, notes)
	}
	var starshardsLine *proto.APLListItem
	if k["starshards"] > 0 {
		starshardsLine = channel(starshards, nil, nil, "Night Elf. Starshards on cooldown. Other races can ignore the warning on this line.")
	}

	if starshardsLine != nil && k["starshards"] == 2 {
		items = append(items, starshardsLine)
	}
	if swdLine != nil && k["swdFirst"] == 1 {
		items = append(items, swdLine)
	}
	if k["innerFocus"] == 1 {
		// Only once the lines above Mind Blast have nothing to cast, or one of them takes
		// the free spell.
		var dpReady value
		if k["dp"] == 1 {
			dpReady = or(dotIsActive(devouringPlague), not(spellIsReady(devouringPlague)))
		}
		items = append(items, cast(innerFocus, and(spellIsReady(mindBlast), dotIsActive(shadowWordPain), dpReady), "Inner Focus right before Mind Blast. Without the talent ignore the warning on this line."))
	} else if k["innerFocus"] == 2 {
		items = append(items, cast(innerFocus, nil, "Inner Focus on cooldown. Without the talent ignore the warning on this line."))
	}
	items = append(items, cast(mindBlast, priestManaAbove(k["mbMana"]), "Mind Blast on cooldown."))
	if swdLine != nil && k["swdFirst"] == 0 {
		items = append(items, swdLine)
	}
	if starshardsLine != nil && k["starshards"] == 1 {
		items = append(items, starshardsLine)
	}
	if k["chastise"] == 1 {
		items = append(items, cast(chastise, nil, "Dwarf. Chastise on cooldown against a Humanoid. Other races can ignore the warning on this line."))
	}

	if k["shadowguard"] == 1 && k["shadowguardRecast"] > 0 {
		items = append(items, cast(shadowguard, and(not(auraIsActive(shadowguard)), ge(remainingTime(), seconds(k["shadowguardRecast"])), priestManaAbove(k["shadowguardMana"])),
			"Troll. Shadowguard again once its charges are spent. Other races can ignore the warning on this line."))
	}

	var clip value
	mfNotes := "Mind Flay fills."
	if k["mfClip"] > 0 {
		clip = and(spellIsReady(mindBlast), priestManaAbove(k["mbMana"]), ge(channeledTicks(mindFlay), num(k["mfClip"])))
		mfNotes = "Mind Flay fills, clipped for a ready Mind Blast."
	}
	if k["mfClipDots"] > 0 {
		dotDown := not(dotIsActive(shadowWordPain))
		if k["dp"] == 1 {
			dotDown = or(dotDown, and(not(dotIsActive(devouringPlague)), spellIsReady(devouringPlague)))
		}
		clip = or(clip, and(dotDown, ge(channeledTicks(mindFlay), num(k["mfClipDots"]))))
		mfNotes = strings.TrimSuffix(mfNotes, ".") + ", and when a damage over time spell fell off."
	}
	items = append(items, channel(mindFlay, nil, clip, mfNotes))
	items = append(items, priestWandFallback(k)...)

	rot := rotation(items...)
	// Shadowguard needs Shadowform, so Shadowform goes a GCD earlier for it.
	var guard float64
	if k["shadowguard"] == 1 {
		guard = 1.5
	}
	if k["prepull"] == 0 {
		rot.PrepullActions = append(rot.PrepullActions, prepull(shadowform, 1.5+guard))
		if guard > 0 {
			rot.PrepullActions = append(rot.PrepullActions, prepull(shadowguard, 1.5))
		}
	} else {
		rot.PrepullActions = append(rot.PrepullActions, prepull(shadowform, 3+guard))
		if guard > 0 {
			rot.PrepullActions = append(rot.PrepullActions, prepull(shadowguard, 3))
		}
		if k["prepull"] == 2 {
			rot.PrepullActions = append(rot.PrepullActions, prepull(innerFocus, 1.5))
		}
		rot.PrepullActions = append(rot.PrepullActions, prepull(mindBlast, 1.5))
	}
	if k["prepullCDs"] >= 1 {
		rot.PrepullActions = append(rot.PrepullActions, prepull(itemID(priest.PriestSyntheticOnUseTrinket), 0.1))
	}
	if k["prepullCDs"] >= 2 {
		rot.PrepullActions = append(rot.PrepullActions, prepull(itemID(majorSpellblastingPotion), 0.1))
	}
	return rot
}

// The wand line that regenerates mana outside the five second rule, from the wandLow,
// wandHigh and wandMinTime knobs, for priestShadow and priestSmite.
func priestWandRegen(k Knobs) []*proto.APLListItem {
	if k["wandLow"] <= 0 {
		return nil
	}
	var minTime value
	if k["wandMinTime"] > 0 {
		minTime = ge(remainingTime(), seconds(k["wandMinTime"]))
	}
	return []*proto.APLListItem{channel(wandShoot, and(le(currentManaPercent(), num(k["wandLow"]/100)), minTime), ge(currentManaPercent(), num(k["wandHigh"]/100)),
		fmt.Sprintf("Shoot the wand from %.0f%% mana to %.0f%% with %.0f sec or more to go, to regenerate outside the five second rule.", k["wandLow"], k["wandHigh"], k["wandMinTime"]))}
}

// The wand when nothing else can be cast (out of mana), from the wandFallback knob.
func priestWandFallback(k Knobs) []*proto.APLListItem {
	if k["wandFallback"] != 1 {
		return nil
	}
	return []*proto.APLListItem{cast(wandShoot, nil, "Shoot the wand when nothing else can be cast (out of mana).")}
}

func channeledTicks(spell *proto.ActionID) value {
	return &proto.APLValue{Value: &proto.APLValue_SpellChanneledTicks{SpellChanneledTicks: &proto.APLValueSpellChanneledTicks{SpellId: spell}}}
}

func isExecutePhase20() value {
	return &proto.APLValue{Value: &proto.APLValue_IsExecutePhase{IsExecutePhase: &proto.APLValueIsExecutePhase{
		Threshold: proto.APLValueIsExecutePhase_E20,
	}}}
}
