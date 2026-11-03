package main

import (
	"fmt"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/paladin"
)

// The Protection paladin tank under Forever (docs/tank_paladin.md). Righteous Fury is an option,
// on for the whole fight. The seal stays up (a Judgement no longer consumes it) and is judged on
// cooldown: Seal of Fury, the tanking seal (Holy damage on every swing, an absorb of half of it with
// a shield, and a Judgement that taunts), or Seal of Righteousness. Holy Shield is kept up (30%
// block, 4 charges, its block damage at 1.2 times Holy threat), Holy Strike goes on cooldown (Iron
// Creed's threat and damage cut, Shieldsworn's Holy Shield hit), Consecration with the mana to
// spare. Every line names the rank the player's level knows. Nothing reads the tank's health, so the
// tank score's replay holds (docs/tank_sim.md): Templar's Bulwark, when the template casts it, goes
// out on a timer.
//
// The paladin keeps none of the raid's debuffs itself (tankscore.OwnDebuffs): Judgement of Fury
// leaves no debuff, and the raid's Judgements of Wisdom, Light and the Crusader are other paladins'.
type paladinTank struct{}

func (paladinTank) Knobs() []Knob {
	return []Knob{
		// 1 Seal of Fury, 2 Seal of Righteousness.
		{Name: "seal", Default: 1, Min: 1, Max: 2, Step: 1},
		// Put the seal back up when it has less than this many tenths of a second left (0: once it
		// is down).
		{Name: "resealWindow", Default: 0, Min: 0, Max: 30, Step: 10},
		// Holy Shield on cooldown above this much mana (percent), and ahead of Holy Strike.
		{Name: "holyShieldMana", Default: 0, Min: 0, Max: 60, Step: 20},
		{Name: "holyShieldFirst", Default: 1, Min: 0, Max: 1, Step: 1},
		// Holy Shield cast before the pull, so it is up for the first swings.
		{Name: "prepullHolyShield", Default: 1, Min: 0, Max: 1, Step: 1},
		// Holy Strike on cooldown above this much mana.
		{Name: "holyStrikeMana", Default: 0, Min: 0, Max: 40, Step: 20},
		// Judgement on cooldown (off the global cooldown) above this much mana.
		{Name: "judgementMana", Default: 0, Min: 0, Max: 60, Step: 20},
		// Swift Judgement (the 15 point Protection talent) on cooldown.
		{Name: "swiftJudgement", Default: 1, Min: 0, Max: 1, Step: 1},
		// Consecration on cooldown above consecrationMana percent of mana (100 never), after Holy
		// Strike, or ahead of it with consecrationFirst.
		{Name: "consecrationMana", Default: 60, Min: 0, Max: 100, Step: 20},
		{Name: "consecrationFirst", Default: 0, Min: 0, Max: 1, Step: 1},
		// Templar's Bulwark on cooldown from bulwarkAt seconds into the fight (a timer, no health
		// read): -1 never.
		{Name: "bulwarkAt", Default: -1, Min: -1, Max: 59, Step: 15},
	}
}

func currentSealRemainingTime() value {
	return &proto.APLValue{Value: &proto.APLValue_CurrentSealRemainingTime{CurrentSealRemainingTime: &proto.APLValueCurrentSealRemainingTime{}}}
}

func currentTime() value {
	return &proto.APLValue{Value: &proto.APLValue_CurrentTime{CurrentTime: &proto.APLValueCurrentTime{}}}
}

func manaFloor(pct float64) value {
	if pct <= 0 {
		return nil
	}
	return ge(currentManaPercent(), num(pct/100))
}

func (paladinTank) Build(k Knobs) *proto.APLRotation {
	seal := rankID(paladin.SealOfFuryLevel[:], paladin.SealOfFuryRankSpellId[:])
	sealName := "Seal of Fury"
	if k["seal"] == 2 {
		seal = rankID(paladin.SealOfRighteousnessLevel[:], paladin.SealOfRighteousnessSpellId[:])
		sealName = "Seal of Righteousness"
	}
	holyShield := rankID(paladin.HolyShieldLevel[:], paladin.HolyShieldSpellId[:])
	holyStrike := rankID(paladin.HolyStrikeLevel[:], paladin.HolyStrikeSpellId[:])
	consecration := rankID(paladin.ConsecrationLevel[:], paladin.ConsecrationSpellId[:])
	bulwark := spellID(paladin.TemplarsBulwarkSpellId)

	var items []*proto.APLListItem
	if k["bulwarkAt"] >= 0 {
		items = append(items, cast(bulwark, ge(currentTime(), seconds(k["bulwarkAt"])),
			fmt.Sprintf("Templar's Bulwark on cooldown from %d sec in, on a timer (a scoring run reads no health). Off the global cooldown. Without the talent ignore the warning on this line.", int(k["bulwarkAt"]))))
	}
	sealDown := not(auraIsActive(seal))
	sealNotes := sealName + " whenever it is down. Judgement does not consume it under Forever."
	if w := k["resealWindow"]; w > 0 {
		sealDown = lt(currentSealRemainingTime(), seconds(w/10))
		sealNotes = fmt.Sprintf("%s again when it has less than %.1f sec left.", sealName, w/10)
	}
	items = append(items, cast(seal, sealDown, sealNotes))
	if k["swiftJudgement"] == 1 {
		items = append(items, cast(spellID(paladin.SwiftJudgementSpellId), auraIsActive(seal),
			"Swift Judgement on cooldown, off the global cooldown. Without the talent ignore the warning on this line."))
	}
	items = append(items, cast(judgement, and(nonNil([]value{auraIsActive(seal), manaFloor(k["judgementMana"])})...),
		"Judgement on cooldown, off the global cooldown."))

	holyShieldLine := cast(holyShield, manaFloor(k["holyShieldMana"]), "Holy Shield on cooldown: 30% block for 10 sec or 4 blocks, its damage at 1.2 times the threat.")
	holyStrikeLine := cast(holyStrike, manaFloor(k["holyStrikeMana"]), "Holy Strike on cooldown: Iron Creed's threat and damage cut, and with Shieldsworn Holy Shield's damage.")
	var consecrationLine *proto.APLListItem
	if k["consecrationMana"] < 100 {
		consecrationLine = cast(consecration, manaFloor(k["consecrationMana"]), "Consecration with the mana to spare.")
	}
	if k["holyShieldFirst"] == 1 {
		items = append(items, holyShieldLine)
	}
	if consecrationLine != nil && k["consecrationFirst"] == 1 {
		items = append(items, consecrationLine)
	}
	items = append(items, holyStrikeLine)
	if k["holyShieldFirst"] == 0 {
		items = append(items, holyShieldLine)
	}
	if consecrationLine != nil && k["consecrationFirst"] == 0 {
		items = append(items, consecrationLine)
	}
	items = append(items, autocastOtherCooldowns(nil, "Racials and trinkets on a free GCD."))

	rot := rotation(items...)
	rot.PrepullActions = append(rot.PrepullActions, prepull(seal, 3))
	if k["prepullHolyShield"] == 1 {
		rot.PrepullActions = append(rot.PrepullActions, prepull(holyShield, 1.5))
	}
	return rot
}
