package main

import (
	"github.com/wowsims/classic/sim/core/proto"
)

func init() {
	templates["hunter_ranged"] = hunterRanged{}
}

// The level 60 hunter spells the templates name.
var (
	hunterAimedShot        = spellID(20904)
	hunterArcaneShot       = spellID(14287)
	hunterMultiShot        = spellID(2643)
	hunterSerpentSting     = spellID(13555) // rank 8, rank 9 is an AQ book
	hunterSummonHawk       = spellID(1293527)
	hunterSniperShot       = spellID(1310786)
	hunterRapidFire        = spellID(3045)
	hunterIntimidation     = spellID(19577)
	hunterBestialWrath     = spellID(19574)
	hunterAspectOfTheHawk  = spellID(14321) // rank 5, see sim/hunter/aspects.go
	hunterAspectOfTheBeast = spellID(1299447)
	hunterRaptorStrikeQ    = spellIDTag(14266, 3)
	hunterRaptorStrike     = spellID(14266)
	hunterMongooseBite     = spellID(14271)
	hunterStriderKick      = spellID(1317257)
	hunterLacerate         = spellID(1299332)
	hunterWingClip         = spellID(14268)
	hunterImmolationTrap   = spellID(14305)
	nightElfElunesLight    = spellID(1259799)
	hunterHuntersMark      = spellID(14325) // rank 4, sim/hunter/hunters_mark.go
	hunterScorpidSting     = spellID(3043)  // Forever's one rank, sim/hunter/scorpid_sting.go
	// The ranged sets' synthetic on-use trinket (sim/hunter/forever_synthetic_trinkets.go).
	hunterRangedOnUse = itemID(993016)
)

// Beast Mastery and Marksmanship under Forever rules. Auto Shot keeps firing through casts
// (sim/hunter/hunter.go), so the shots only compete for the GCD and the mana. Aimed Shot and
// Multi-Shot share a 6 sec cooldown, Arcane Shot and Summon Hawk another, Sniper Shot has its
// own 15 sec one. The lines that need a talent the build lacks simply never fire.
type hunterRanged struct{}

func (hunterRanged) Knobs() []Knob {
	return []Knob{
		// 1 Aimed Shot on the shared cooldown, 0 Multi-Shot.
		{Name: "aimed", Default: 1, Min: 0, Max: 1, Step: 1},
		// What goes on the Arcane Shot cooldown: 0 Arcane Shot, 1 Summon Hawk, 2 Summon Hawk while
		// fewer than 2 hawks are out and Arcane Shot otherwise.
		{Name: "hawk", Default: 2, Min: 0, Max: 2, Step: 1},
		{Name: "serpent", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "sniper", Default: 1, Min: 0, Max: 1, Step: 1},
		// Sniper Shot above Aimed Shot when both are ready.
		{Name: "sniperFirst", Default: 0, Min: 0, Max: 1, Step: 1},
		// The Arcane Shot cooldown above Aimed Shot when both are ready.
		{Name: "arcaneFirst", Default: 0, Min: 0, Max: 1, Step: 1},
		// Sniper Shot only while the mana covers this much a second of the fight left (0 off). It is
		// the least mana efficient shot, so in a long fight it goes first when mana runs short.
		{Name: "sniperManaPerSec", Default: 0, Min: 0, Max: 40, Step: 5},
		// Rapid Fire 0.1 sec before the pull, so the second Auto Shot is already hasted.
		{Name: "rapidPrepull", Default: 0, Min: 0, Max: 1, Step: 1},
		// Serpent Sting below Aimed Shot, so a refresh never delays Aimed Shot's cooldown, which starts
		// when the cast ends (1), or last, below the Arcane Shot cooldown too (2).
		{Name: "serpentLate", Default: 0, Min: 0, Max: 2, Step: 1},
		{Name: "intimidation", Default: 0, Min: 0, Max: 1, Step: 1},
		// Mana floors, in percent: Serpent Sting and the Arcane Shot cooldown only above them.
		{Name: "serpentMana", Default: 30, Min: 0, Max: 60, Step: 10},
		{Name: "arcaneMana", Default: 20, Min: 0, Max: 60, Step: 10},
		// Aimed Shot only above this much mana (Multi-Shot is a third of the cost).
		{Name: "aimedMana", Default: 0, Min: 0, Max: 40, Step: 10},
		// Hold Rapid Fire, and the Night Elf's Elune's Light, for the on-use trinket when the wait
		// costs no use. Only a fight long enough for a third trinket use sees a difference.
		{Name: "holdRapid", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "holdElune", Default: 0, Min: 0, Max: 1, Step: 1},
		// Hunter's Mark duty (standards 23): the hunter keeps the mark up itself, with the raid's
		// Hunter's Mark debuff off in the request. The base presets leave it to the raid.
		{Name: "mark", Default: 0, Min: 0, Max: 1, Step: 1},
		// Seconds before the pull the duty casts the mark, 0 not before the pull. Hunter's Mark doesn't
		// pull, so it may go up before it (standards 23, user ruling 2026-10-15).
		{Name: "markPrepull", Default: 0, Min: 0, Max: 3, Step: 0.5},
		// Scorpid Sting duty (standards 23): Scorpid Sting kept up in the sting slot in place of
		// Serpent Sting, with the raid's Scorpid Sting off.
		{Name: "scorpid", Default: 0, Min: 0, Max: 1, Step: 1},
		// Seconds left on the sting when the duty refreshes it.
		{Name: "scorpidLead", Default: 2, Min: 1, Max: 6, Step: 1},
		// Melee weaving (hunterWeave): between Auto Shots the hunter walks into melee for a main hand
		// swing, with Raptor Strike queued on it, and back out past the deadzone. 0 stands at range.
		{Name: "weave", Default: 0, Min: 0, Max: 1, Step: 1},
		// The most a weave may hold the next Auto Shot back, in seconds.
		{Name: "weaveDelay", Default: 1.2, Min: 0, Max: 2.5, Step: 0.1},
		// The reaction the weave allows for, in seconds: after an Auto Shot before walking in, and
		// after a walk out before Auto Shot starts again (the sim's reaction time, see
		// core.AutoAttacks.resumeRangedAfterMove).
		{Name: "weaveReact", Default: 0.2, Min: 0, Max: 0.4, Step: 0.05},
		// Seconds in melee after the swing before walking out.
		{Name: "weaveMeleeWait", Default: 0.2, Min: 0, Max: 0.4, Step: 0.05},
		// Aimed Shot and Sniper Shot, which hold the hunter still, only when the cast ends before the
		// walk in (1), or whenever ready (0).
		{Name: "weaveCasts", Default: 1, Min: 0, Max: 1, Step: 1},
		// Walk in only when Arcane Shot and the Aimed Shot / Multi-Shot cooldown stay down until the
		// hunter is back at range, less this many seconds (off at 3).
		{Name: "weaveShotSlack", Default: 3, Min: 0, Max: 3, Step: 0.25},
		// Walk in only for a swing Raptor Strike is ready for (1), or for white swings too (0).
		{Name: "weaveRaptor", Default: 0, Min: 0, Max: 1, Step: 1},
		// Walk in only while the mana covers this much a second of the fight left (0 off). Raptor
		// Strike costs mana and a weave can push a shot back, so a long fight stops weaving when the
		// mana runs short.
		{Name: "weaveManaPerSec", Default: 0, Min: 0, Max: 60, Step: 5},
	}
}

func (hunterRanged) Build(k Knobs) *proto.APLRotation {
	floor := func(name string) value {
		if k[name] <= 0 {
			return nil
		}
		return ge(currentManaPercent(), num(k[name]/100))
	}

	weave := hunterWeave{k}
	items := []*proto.APLListItem{
		{
			Notes: "Step back to 25 yards when the fight starts in melee range, inside the 8 yard minimum.",
			Action: &proto.APLAction{
				Condition: inMeleeRange(),
				Action:    &proto.APLAction_Move{Move: &proto.APLActionMove{RangeFromTarget: num(25)}},
			},
		},
		cast(hunterAspectOfTheHawk, not(auraIsActive(hunterAspectOfTheHawk)), "Aspect of the Hawk if it ever drops."),
	}
	if weave.on() {
		items = append(weave.melee(), items[1:]...)
	}
	if k["holdRapid"] == 1 {
		items = append(items, cast(hunterRapidFire, alignedWith(hunterRapidFire, hunterRangedOnUse), "Rapid Fire with the on-use trinket, unless waiting would cost a use."))
	}
	if k["holdElune"] == 1 {
		items = append(items, cast(nightElfElunesLight, alignedWith(nightElfElunesLight, hunterRangedOnUse), "Elune's Light with the on-use trinket, unless waiting would cost a use."))
	}
	items = append(items, autocastOtherCooldowns(nil, "Rapid Fire, Bestial Wrath, racials, trinkets and potions on cooldown."))
	if k["mark"] == 1 {
		items = append(items, hunterMarkDuty())
	}
	// One sting a hunter: on Scorpid Sting duty it takes Serpent Sting's place.
	serpentOn := k["serpent"] == 1 && k["scorpid"] == 0
	if k["scorpid"] == 1 {
		items = append(items, hunterScorpidDuty(k))
	}
	if k["intimidation"] == 1 {
		items = append(items, cast(hunterIntimidation, nil, "Intimidation on cooldown for the pet's crit, with the talent."))
	}
	serpent := cast(hunterSerpentSting, and(not(dotIsActive(hunterSerpentSting)), floor("serpentMana")), "Serpent Sting when it is down.")
	if serpentOn && k["serpentLate"] == 0 {
		items = append(items, serpent)
	}

	arcaneFloor := floor("arcaneMana")
	var arcane []*proto.APLListItem
	switch k["hawk"] {
	case 1:
		arcane = append(arcane, cast(hunterSummonHawk, arcaneFloor, "Summon Hawk on the cooldown it shares with Arcane Shot, with the talent."))
	case 2:
		hawksOut := and(dotIsActive(spellIDTag(1293527, 1)), dotIsActive(spellIDTag(1293527, 2)))
		arcane = append(arcane, cast(hunterSummonHawk, and(not(hawksOut), arcaneFloor), "Summon Hawk while fewer than 2 hawks are out, with the talent."))
	}
	arcane = append(arcane, cast(hunterArcaneShot, arcaneFloor, "Arcane Shot on its cooldown."))

	if k["arcaneFirst"] == 1 {
		items = append(items, arcane...)
	}
	var sniperGate value
	if k["sniperManaPerSec"] > 0 {
		sniperGate = ge(hunterCurrentMana(), hunterMul(remainingTime(), num(k["sniperManaPerSec"])))
	}
	sniper := cast(hunterSniperShot, weave.castFits(hunterSniperShot, sniperGate), "Sniper Shot on cooldown, with the talent, while the mana lasts the fight.")
	if k["sniper"] == 1 && k["sniperFirst"] == 1 {
		items = append(items, sniper)
	}
	if k["aimed"] == 1 {
		items = append(items, cast(hunterAimedShot, weave.castFits(hunterAimedShot, floor("aimedMana")), "Aimed Shot on the cooldown it shares with Multi-Shot."))
	}
	items = append(items, cast(hunterMultiShot, nil, "Multi-Shot when Aimed Shot is not taken."))
	if serpentOn && k["serpentLate"] == 1 {
		items = append(items, serpent)
	}
	if k["sniper"] == 1 && k["sniperFirst"] == 0 {
		items = append(items, sniper)
	}
	if k["arcaneFirst"] == 0 {
		items = append(items, arcane...)
	}
	if serpentOn && k["serpentLate"] == 2 {
		items = append(items, serpent)
	}
	if weave.on() {
		items = append(items, weave.walkIn()...)
	}

	rot := rotation(items...)
	if weave.on() {
		// Into position before the pull, from wherever the fight starts.
		rot.PrepullActions = append(rot.PrepullActions, &proto.APLPrepullAction{
			Action:    &proto.APLAction{Action: &proto.APLAction_Move{Move: &proto.APLActionMove{RangeFromTarget: num(hunterWeaveShootSpot)}}},
			DoAtValue: seconds(-6),
		})
	}
	rot.PrepullActions = append(rot.PrepullActions, prepull(hunterAspectOfTheHawk, 3))
	if k["mark"] == 1 && k["markPrepull"] > 0 {
		rot.PrepullActions = append(rot.PrepullActions, prepull(hunterHuntersMark, k["markPrepull"]))
	}
	if k["rapidPrepull"] == 1 {
		rot.PrepullActions = append(rot.PrepullActions, prepull(hunterRapidFire, 0.1))
	}
	return rot
}

func hunterCurrentMana() value {
	return &proto.APLValue{Value: &proto.APLValue_CurrentMana{CurrentMana: &proto.APLValueCurrentMana{}}}
}

func hunterMul(lhs, rhs value) value {
	return &proto.APLValue{Value: &proto.APLValue_Math{Math: &proto.APLValueMath{Op: proto.APLValueMath_OpMul, Lhs: lhs, Rhs: rhs}}}
}

// hunterMarkDuty keeps the hunter's own Hunter's Mark on the target: the first global cooldown
// after the pull when it didn't go up before it (markPrepull), then again in the last 3 sec of its
// 2 min.
func hunterMarkDuty() *proto.APLListItem {
	return cast(hunterHuntersMark, lt(targetAuraRemainingTime(hunterHuntersMark), seconds(3)),
		"Hunter's Mark duty: the mark at the pull and again before it drops (the raid's mark is off).")
}

// hunterScorpidDuty keeps the hunter's own Scorpid Sting on the target: the first global cooldown
// after the pull, then again in its last scorpidLead sec, unless one is already on its way.
func hunterScorpidDuty(k Knobs) *proto.APLListItem {
	return cast(hunterScorpidSting, and(hunterScorpidDue(k), not(spellInFlight(hunterScorpidSting))),
		"Scorpid Sting duty: the sting at the pull and again before it drops, in place of Serpent Sting (the raid's Scorpid Sting is off).")
}

func hunterScorpidDue(k Knobs) value {
	return lt(targetAuraRemainingTime(hunterScorpidSting), seconds(k["scorpidLead"]))
}

func hunterMax(vals ...value) value {
	return &proto.APLValue{Value: &proto.APLValue_Max{Max: &proto.APLValueMax{Vals: vals}}}
}

func hunterMin(vals ...value) value {
	return &proto.APLValue{Value: &proto.APLValue_Min{Min: &proto.APLValueMin{Vals: vals}}}
}

func hunterRangedTimeToNext() value {
	return &proto.APLValue{Value: &proto.APLValue_AutoTimeToNext{AutoTimeToNext: &proto.APLValueAutoTimeToNext{AutoType: proto.APLValueAutoTimeToNext_Ranged}}}
}

func hunterRangedSwingTime() value {
	return &proto.APLValue{Value: &proto.APLValue_AutoSwingTime{AutoSwingTime: &proto.APLValueAutoSwingTime{AutoType: proto.APLValueAutoSwingTime_Ranged}}}
}

func hunterGCDTimeToReady() value {
	return &proto.APLValue{Value: &proto.APLValue_GcdTimeToReady{GcdTimeToReady: &proto.APLValueGCDTimeToReady{}}}
}

func hunterMoveTime(yards float64) value {
	return &proto.APLValue{Value: &proto.APLValue_MoveTime{MoveTime: &proto.APLValueMoveTime{RangeFromTarget: num(yards)}}}
}

func hunterMove(yards float64, condition value, notes string) *proto.APLListItem {
	return action(condition, notes, &proto.APLAction{Action: &proto.APLAction_Move{Move: &proto.APLActionMove{RangeFromTarget: num(yards)}}})
}

func hunterWait(duration value, condition value, notes string) *proto.APLListItem {
	return action(condition, notes, &proto.APLAction{Action: &proto.APLAction_Wait{Wait: &proto.APLActionWait{Duration: duration}}})
}

// Where a weave shoots from and walks in to. Forever's deadzone ends at 11.67 yd (8 yd past both
// combat reaches, core.ForeverMinRangedAttackDistance); the weave shoots from 12, the third of a yard
// past it a player overshoots by (wowsims/forever 4e02d3ce95), and walks in to just inside melee range.
const (
	hunterWeaveShootSpot = 12
	hunterWeaveMeleeSpot = 4.99
)

// hunterWeave is the weave of the hunter_ranged template: the knobs weave, weaveDelay, weaveReact,
// weaveMeleeWait, weaveCasts, weaveShotSlack, weaveRaptor and weaveManaPerSec. A weave walks in when
// the main hand swing is ready by the time the hunter gets there, and only when it holds the next
// Auto Shot back by weaveDelay or less: Auto Shot
// can't go off on the move, and after the walk out it winds up for 0.5 sec a reaction time after the
// hunter stops (core.AutoAttacks.resumeRangedAfterMove). The walks take the hunter's run speed
// (Move Time), so Strider Kick and the run speed buffs shorten them.
type hunterWeave struct{ k Knobs }

func (w hunterWeave) on() bool { return w.k["weave"] == 1 }

// The walk either way, as seen from the shoot spot.
func (w hunterWeave) walk() value { return hunterMoveTime(hunterWeaveMeleeSpot) }

// When the hunter is back at the shoot spot after a weave that starts now or once the swing allows.
func (w hunterWeave) back() value {
	return add(add(hunterMax(mainHandTimeToNext(), w.walk()), seconds(w.k["weaveMeleeWait"])), w.walk())
}

// The weave fits when the next Auto Shot goes off no more than weaveDelay late: the hunter is back
// out, with the reaction and the windup, by then. With weaveShotSlack below 3 the shots on the
// global cooldown must also stay on cooldown until about then.
func (w hunterWeave) fits() value {
	fits := ge(add(hunterRangedTimeToNext(), seconds(w.k["weaveDelay"])), add(w.back(), seconds(w.k["weaveReact"]+0.5)))
	if w.k["weaveShotSlack"] >= 3 {
		return fits
	}
	shotsDown := sub(w.back(), seconds(w.k["weaveShotSlack"]))
	return and(fits, ge(spellTimeToReady(hunterArcaneShot), shotsDown), ge(spellTimeToReady(hunterMultiShot), shotsDown))
}

// The condition of a cast that holds the hunter still: with weaveCasts it goes out only when it ends
// before the walk in, or when no weave fits before the next Auto Shot.
func (w hunterWeave) castFits(spell *proto.ActionID, condition value) value {
	if !w.on() || w.k["weaveCasts"] == 0 {
		return condition
	}
	return and(condition, or(not(w.fits()), le(add(spellCastTime(spell), w.walk()), mainHandTimeToNext())))
}

// In melee: Raptor Strike queued, the swing, a moment, and back out to the shoot spot. Nothing
// else goes out in melee.
func (w hunterWeave) melee() []*proto.APLListItem {
	sinceSwing := sub(mainHandSwingTime(), mainHandTimeToNext())
	return []*proto.APLListItem{
		cast(hunterRaptorStrikeQ, nil, "Weave: Raptor Strike queued on the swing."),
		cast(hunterMongooseBite, nil, "Weave: Mongoose Bite when a dodge opens it."),
		cast(hunterStriderKick, nil, "Weave: Strider Kick, with the talent, for its run speed out."),
		hunterWait(add(mainHandTimeToNext(), seconds(0.001)), and(inMeleeRange(), gt(sinceSwing, seconds(0.5))),
			"Weave: in melee, wait for the swing."),
		hunterWait(sub(seconds(w.k["weaveMeleeWait"]), sinceSwing), and(inMeleeRange(), lt(sinceSwing, seconds(w.k["weaveMeleeWait"]))),
			"Weave: a moment after the swing."),
		hunterMove(hunterWeaveShootSpot, inMeleeRange(), "Weave: back out past the deadzone."),
	}
}

// At range, below every shot: walk in when the swing is ready on arrival and the weave fits, a
// reaction after the Auto Shot that just went off, or wait for that moment.
func (w hunterWeave) walkIn() []*proto.APLListItem {
	sinceShot := sub(hunterRangedSwingTime(), hunterRangedTimeToNext())
	react := seconds(w.k["weaveReact"])
	ready := le(mainHandTimeToNext(), w.walk())
	if w.k["weaveRaptor"] == 1 {
		ready = and(ready, le(spellTimeToReady(hunterRaptorStrike), w.walk()))
	}
	if w.k["weaveManaPerSec"] > 0 {
		ready = and(ready, ge(hunterCurrentMana(), hunterMul(remainingTime(), num(w.k["weaveManaPerSec"]))))
	}
	return []*proto.APLListItem{
		hunterWait(sub(react, sinceShot), and(ready, w.fits(), lt(sinceShot, react)),
			"Weave: a reaction after the Auto Shot."),
		hunterMove(hunterWeaveMeleeSpot, and(ready, w.fits()), "Weave: walk in for the swing."),
		hunterWait(hunterMin(sub(mainHandTimeToNext(), w.walk()), hunterMax(hunterGCDTimeToReady(), seconds(0.05))),
			and(not(ready), w.fits()), "Weave: wait for the walk in."),
	}
}
