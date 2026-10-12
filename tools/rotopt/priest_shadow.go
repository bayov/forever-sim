package main

import (
	"strings"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/priest"
)

// Shadow under Forever rules. Shadowform goes up before the pull (40% of base mana). The
// damage over time spells come first, then the cooldown spells (Mind Blast, Shadow Word:
// Death, Starshards for a Night Elf, Chastise for a Dwarf), and Mind Flay fills. Every line
// names the highest rank the player's level knows. Lines for spells the build or race
// cannot have only warn, and a knob at 0 drops its line.
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
)

func (priestShadow) Knobs() []Knob {
	return []Knob{
		// Shadow Word: Pain when it is down, with at least this many seconds of fight left.
		{Name: "swpMinTime", Default: 9, Min: 0, Max: 18, Step: 3},
		// Devouring Plague when it is down and off its 1 min cooldown, with at least this
		// many seconds of fight left and above a mana floor in percent.
		{Name: "dp", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "dpMinTime", Default: 12, Min: 0, Max: 24, Step: 6},
		{Name: "dpMana", Default: 0, Min: 0, Max: 60, Step: 10},
		// Shadow Word: Death on cooldown (15 sec), 2 for only in the execute phase.
		{Name: "swd", Default: 1, Min: 0, Max: 2, Step: 1},
		{Name: "swdMana", Default: 0, Min: 0, Max: 60, Step: 10},
		// Shadow Word: Death above Mind Blast in the priority.
		{Name: "swdFirst", Default: 0, Min: 0, Max: 1, Step: 1},
		// Mind Blast above a mana floor.
		{Name: "mbMana", Default: 0, Min: 0, Max: 40, Step: 10},
		// Inner Focus: 0 never, 1 on Mind Blast, 2 on cooldown with whatever comes next.
		{Name: "innerFocus", Default: 1, Min: 0, Max: 2, Step: 1},
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
		// Undead: Dark Sacrifice under this much mana in percent (0 never).
		{Name: "darkSacrifice", Default: 40, Min: 0, Max: 80, Step: 10},
		// Vampiric Embrace kept on the target. It heals the party for 20% of the priest's
		// Shadow damage and deals none, so each cast is a GCD of damage given up.
		{Name: "vampiricEmbrace", Default: 0, Min: 0, Max: 1, Step: 1},
		// Power Infusion on cooldown (needs the 31 point Discipline talent, which no level 60
		// build takes).
		{Name: "powerInfusion", Default: 0, Min: 0, Max: 1, Step: 1},
		// Before the pull: 0 Shadowform only (1.5 sec out), 1 Shadowform 3 sec out and a Mind
		// Blast 1.5 sec out that lands at 0, 2 the same with Inner Focus on that Mind Blast.
		{Name: "prepull", Default: 1, Min: 0, Max: 2, Step: 1},
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

func manaAbove(percent float64) value {
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

	var items []*proto.APLListItem
	if k["darkSacrifice"] > 0 {
		items = append(items, cast(darkSacrifice, lt(currentManaPercent(), num(k["darkSacrifice"]/100)), "Undead. Dark Sacrifice when mana runs low. Other races can ignore the warning on this line."))
	}
	if k["powerInfusion"] == 1 {
		items = append(items, cast(powerInfusion, nil, "Power Infusion on cooldown. Without the talent ignore the warning on this line."))
	}
	items = append(items, autocastOtherCooldowns(nil, "Racials and trinkets on cooldown."))

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
		items = append(items, cast(devouringPlague, and(not(dotIsActive(devouringPlague)), dpTime, manaAbove(k["dpMana"])), "Devouring Plague whenever it is down and off cooldown."))
	}

	var swdLine *proto.APLListItem
	if k["swd"] > 0 {
		var cond value = manaAbove(k["swdMana"])
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
	items = append(items, cast(mindBlast, manaAbove(k["mbMana"]), "Mind Blast on cooldown."))
	if swdLine != nil && k["swdFirst"] == 0 {
		items = append(items, swdLine)
	}
	if starshardsLine != nil && k["starshards"] == 1 {
		items = append(items, starshardsLine)
	}
	if k["chastise"] == 1 {
		items = append(items, cast(chastise, nil, "Dwarf. Chastise on cooldown against a Humanoid. Other races can ignore the warning on this line."))
	}

	var clip value
	mfNotes := "Mind Flay fills."
	if k["mfClip"] > 0 {
		clip = and(spellIsReady(mindBlast), manaAbove(k["mbMana"]), ge(channeledTicks(mindFlay), num(k["mfClip"])))
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

	rot := rotation(items...)
	if k["prepull"] == 0 {
		rot.PrepullActions = append(rot.PrepullActions, prepull(shadowform, 1.5))
	} else {
		rot.PrepullActions = append(rot.PrepullActions, prepull(shadowform, 3))
		if k["prepull"] == 2 {
			rot.PrepullActions = append(rot.PrepullActions, prepull(innerFocus, 1.5))
		}
		rot.PrepullActions = append(rot.PrepullActions, prepull(mindBlast, 1.5))
	}
	return rot
}

func channeledTicks(spell *proto.ActionID) value {
	return &proto.APLValue{Value: &proto.APLValue_SpellChanneledTicks{SpellChanneledTicks: &proto.APLValueSpellChanneledTicks{SpellId: spell}}}
}

func isExecutePhase20() value {
	return &proto.APLValue{Value: &proto.APLValue_IsExecutePhase{IsExecutePhase: &proto.APLValueIsExecutePhase{
		Threshold: proto.APLValueIsExecutePhase_E20,
	}}}
}
