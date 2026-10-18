package aplkit

import "github.com/wowsims/classic/sim/core/proto"

// APL helpers for leveling sessions (sim/core/leveling.go).

var PullTarget = &proto.UnitReference{Type: proto.UnitReference_PullTarget}

func CurrentHealthPercent() Value {
	return &proto.APLValue{Value: &proto.APLValue_CurrentHealthPercent{CurrentHealthPercent: &proto.APLValueCurrentHealthPercent{}}}
}

// TargetHealthPercent is our current target's health, as a share of its max.
func TargetHealthPercent() Value {
	return &proto.APLValue{Value: &proto.APLValue_CurrentHealthPercent{CurrentHealthPercent: &proto.APLValueCurrentHealthPercent{
		SourceUnit: &proto.UnitReference{Type: proto.UnitReference_CurrentTarget}}}}
}

func Pct(v float64) Value {
	return Num(v / 100)
}

// PullTargetDistance is 0 when there is no Pull Target.
func PullTargetDistance() Value {
	return &proto.APLValue{Value: &proto.APLValue_LevelingDistance{LevelingDistance: &proto.APLValueLevelingDistance{TargetUnit: PullTarget}}}
}

func MobsOnUs() Value {
	return &proto.APLValue{Value: &proto.APLValue_LevelingMobsOnUs{LevelingMobsOnUs: &proto.APLValueLevelingMobsOnUs{}}}
}

func MobsRunningIn() Value {
	return &proto.APLValue{Value: &proto.APLValue_LevelingMobsRunningIn{LevelingMobsRunningIn: &proto.APLValueLevelingMobsRunningIn{}}}
}

// MobsPulled are the mobs on us and on their way.
func MobsPulled() Value {
	return Add(MobsOnUs(), MobsRunningIn())
}

func IsMoving() Value {
	return &proto.APLValue{Value: &proto.APLValue_IsMoving{IsMoving: &proto.APLValueIsMoving{}}}
}

func BandageReady() Value {
	return &proto.APLValue{Value: &proto.APLValue_LevelingBandageReady{LevelingBandageReady: &proto.APLValueLevelingBandageReady{}}}
}

func TotemDistance(totem proto.ShamanTotems_TotemType) Value {
	return &proto.APLValue{Value: &proto.APLValue_TotemDistance{TotemDistance: &proto.APLValueTotemDistance{TotemType: totem}}}
}

func TotemRemainingTime(totem proto.ShamanTotems_TotemType) Value {
	return &proto.APLValue{Value: &proto.APLValue_TotemRemainingTime{TotemRemainingTime: &proto.APLValueTotemRemainingTime{TotemType: totem}}}
}

// CastAt is a cast at a given unit, like the Pull Target.
func CastAt(spell *proto.ActionID, target *proto.UnitReference, condition Value, notes string) *proto.APLListItem {
	item := Cast(spell, condition, notes)
	item.Action.GetCastSpell().Target = target
	return item
}

func Action(condition Value, notes string, a *proto.APLAction) *proto.APLListItem {
	a.Condition = condition
	return &proto.APLListItem{Notes: notes, Action: a}
}

// LookForMob looks for another mob while fighting, running around for it, or standing and
// fighting while we look.
func LookForMob(stand bool, condition Value, notes string) *proto.APLListItem {
	return Action(condition, notes, &proto.APLAction{Action: &proto.APLAction_LevelingLookForMob{LevelingLookForMob: &proto.APLActionLevelingLookForMob{Stand: stand}}})
}

func WalkToPullTarget(within float64, condition Value, notes string) *proto.APLListItem {
	return Action(condition, notes, &proto.APLAction{Action: &proto.APLAction_LevelingWalkToPullTarget{
		LevelingWalkToPullTarget: &proto.APLActionLevelingWalkToPullTarget{Within: Num(within)}}})
}

func WalkToTotem(totem proto.ShamanTotems_TotemType, within float64, condition Value, notes string) *proto.APLListItem {
	return Action(condition, notes, &proto.APLAction{Action: &proto.APLAction_ShamanWalkToTotem{
		ShamanWalkToTotem: &proto.APLActionShamanWalkToTotem{TotemType: totem, Within: Num(within)}}})
}

// Rest sits down to eat and drink (or just sit) until health and mana are at these shares.
func Rest(eat, drink bool, healthUntilPct, manaUntilPct float64, condition Value, notes string) *proto.APLListItem {
	return Action(condition, notes, &proto.APLAction{Action: &proto.APLAction_LevelingRest{LevelingRest: &proto.APLActionLevelingRest{
		Eat: eat, Drink: drink, HealthUntil: Pct(healthUntilPct), ManaUntil: Pct(manaUntilPct)}}})
}

func Bandage(condition Value, notes string) *proto.APLListItem {
	return Action(condition, notes, &proto.APLAction{Action: &proto.APLAction_LevelingBandage{LevelingBandage: &proto.APLActionLevelingBandage{}}})
}
