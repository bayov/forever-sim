package main

import "github.com/wowsims/classic/sim/core/proto"

// APL helpers for leveling sessions (sim/core/leveling.go).

var pullTarget = &proto.UnitReference{Type: proto.UnitReference_PullTarget}

func currentHealthPercent() value {
	return &proto.APLValue{Value: &proto.APLValue_CurrentHealthPercent{CurrentHealthPercent: &proto.APLValueCurrentHealthPercent{}}}
}

// targetHealthPercent is our current target's health, as a share of its max.
func targetHealthPercent() value {
	return &proto.APLValue{Value: &proto.APLValue_CurrentHealthPercent{CurrentHealthPercent: &proto.APLValueCurrentHealthPercent{
		SourceUnit: &proto.UnitReference{Type: proto.UnitReference_CurrentTarget}}}}
}

func pct(v float64) value {
	return num(v / 100)
}

// pullTargetDistance is 0 when there is no Pull Target.
func pullTargetDistance() value {
	return &proto.APLValue{Value: &proto.APLValue_LevelingDistance{LevelingDistance: &proto.APLValueLevelingDistance{TargetUnit: pullTarget}}}
}

func mobsOnUs() value {
	return &proto.APLValue{Value: &proto.APLValue_LevelingMobsOnUs{LevelingMobsOnUs: &proto.APLValueLevelingMobsOnUs{}}}
}

func mobsRunningIn() value {
	return &proto.APLValue{Value: &proto.APLValue_LevelingMobsRunningIn{LevelingMobsRunningIn: &proto.APLValueLevelingMobsRunningIn{}}}
}

// mobsPulled are the mobs on us and on their way.
func mobsPulled() value {
	return add(mobsOnUs(), mobsRunningIn())
}

func isMoving() value {
	return &proto.APLValue{Value: &proto.APLValue_IsMoving{IsMoving: &proto.APLValueIsMoving{}}}
}

func bandageReady() value {
	return &proto.APLValue{Value: &proto.APLValue_LevelingBandageReady{LevelingBandageReady: &proto.APLValueLevelingBandageReady{}}}
}

func totemDistance(totem proto.ShamanTotems_TotemType) value {
	return &proto.APLValue{Value: &proto.APLValue_TotemDistance{TotemDistance: &proto.APLValueTotemDistance{TotemType: totem}}}
}

func totemRemainingTime(totem proto.ShamanTotems_TotemType) value {
	return &proto.APLValue{Value: &proto.APLValue_TotemRemainingTime{TotemRemainingTime: &proto.APLValueTotemRemainingTime{TotemType: totem}}}
}

// castAt is a cast at a given unit, like the Pull Target.
func castAt(spell *proto.ActionID, target *proto.UnitReference, condition value, notes string) *proto.APLListItem {
	item := cast(spell, condition, notes)
	item.Action.GetCastSpell().Target = target
	return item
}

func action(condition value, notes string, a *proto.APLAction) *proto.APLListItem {
	a.Condition = condition
	return &proto.APLListItem{Notes: notes, Action: a}
}

func lookForMob(condition value, notes string) *proto.APLListItem {
	return action(condition, notes, &proto.APLAction{Action: &proto.APLAction_LevelingLookForMob{LevelingLookForMob: &proto.APLActionLevelingLookForMob{}}})
}

func walkToPullTarget(within float64, condition value, notes string) *proto.APLListItem {
	return action(condition, notes, &proto.APLAction{Action: &proto.APLAction_LevelingWalkToPullTarget{
		LevelingWalkToPullTarget: &proto.APLActionLevelingWalkToPullTarget{Within: num(within)}}})
}

func walkToTotem(totem proto.ShamanTotems_TotemType, within float64, condition value, notes string) *proto.APLListItem {
	return action(condition, notes, &proto.APLAction{Action: &proto.APLAction_ShamanWalkToTotem{
		ShamanWalkToTotem: &proto.APLActionShamanWalkToTotem{TotemType: totem, Within: num(within)}}})
}

// rest sits down to eat and drink (or just sit) until health and mana are at these shares.
func rest(eat, drink bool, healthUntilPct, manaUntilPct float64, condition value, notes string) *proto.APLListItem {
	return action(condition, notes, &proto.APLAction{Action: &proto.APLAction_LevelingRest{LevelingRest: &proto.APLActionLevelingRest{
		Eat: eat, Drink: drink, HealthUntil: pct(healthUntilPct), ManaUntil: pct(manaUntilPct)}}})
}

func bandage(condition value, notes string) *proto.APLListItem {
	return action(condition, notes, &proto.APLAction{Action: &proto.APLAction_LevelingBandage{LevelingBandage: &proto.APLActionLevelingBandage{}}})
}
