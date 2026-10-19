// Package aplkit has small builders for APL protos, so a rotation template reads like the
// condition it expresses rather than like nested struct literals. tools/rotopt and
// tools/levelingplan both use it, and so the leveling rotation templates live here too.
package aplkit

import (
	"strconv"

	"github.com/wowsims/classic/sim/core/proto"
)

type Value = *proto.APLValue

func Constant(val string) Value {
	return &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: val}}}
}

func Num(v float64) Value {
	return Constant(strconv.FormatFloat(v, 'f', -1, 64))
}

func Seconds(v float64) Value {
	return Constant(strconv.FormatFloat(v, 'f', -1, 64) + "s")
}

// And drops nil operands, so a template can pass nil for a rule that is switched off.
// With every rule off it is nil too, so the line has no condition at all.
func And(vals ...Value) Value {
	if len(NonNil(vals)) == 0 {
		return nil
	}
	return &proto.APLValue{Value: &proto.APLValue_And{And: &proto.APLValueAnd{Vals: NonNil(vals)}}}
}

func Or(vals ...Value) Value {
	return &proto.APLValue{Value: &proto.APLValue_Or{Or: &proto.APLValueOr{Vals: NonNil(vals)}}}
}

func NonNil(vals []Value) []Value {
	var out []Value
	for _, v := range vals {
		if v != nil {
			out = append(out, v)
		}
	}
	return out
}

func Not(val Value) Value {
	return &proto.APLValue{Value: &proto.APLValue_Not{Not: &proto.APLValueNot{Val: val}}}
}

func Cmp(lhs Value, op proto.APLValueCompare_ComparisonOperator, rhs Value) Value {
	return &proto.APLValue{Value: &proto.APLValue_Cmp{Cmp: &proto.APLValueCompare{Op: op, Lhs: lhs, Rhs: rhs}}}
}

func Ge(lhs, rhs Value) Value { return Cmp(lhs, proto.APLValueCompare_OpGe, rhs) }
func Gt(lhs, rhs Value) Value { return Cmp(lhs, proto.APLValueCompare_OpGt, rhs) }
func Le(lhs, rhs Value) Value { return Cmp(lhs, proto.APLValueCompare_OpLe, rhs) }
func Lt(lhs, rhs Value) Value { return Cmp(lhs, proto.APLValueCompare_OpLt, rhs) }
func Eq(lhs, rhs Value) Value { return Cmp(lhs, proto.APLValueCompare_OpEq, rhs) }

func Add(lhs, rhs Value) Value {
	return &proto.APLValue{Value: &proto.APLValue_Math{Math: &proto.APLValueMath{Op: proto.APLValueMath_OpAdd, Lhs: lhs, Rhs: rhs}}}
}

func Sub(lhs, rhs Value) Value {
	return &proto.APLValue{Value: &proto.APLValue_Math{Math: &proto.APLValueMath{Op: proto.APLValueMath_OpSub, Lhs: lhs, Rhs: rhs}}}
}

func MainHandTimeToNext() Value {
	return &proto.APLValue{Value: &proto.APLValue_AutoTimeToNext{AutoTimeToNext: &proto.APLValueAutoTimeToNext{AutoType: proto.APLValueAutoTimeToNext_MainHand}}}
}

func MainHandSwingTime() Value {
	return &proto.APLValue{Value: &proto.APLValue_AutoSwingTime{AutoSwingTime: &proto.APLValueAutoSwingTime{AutoType: proto.APLValueAutoSwingTime_MainHand}}}
}

func ComboPoints() Value {
	return &proto.APLValue{Value: &proto.APLValue_CurrentComboPoints{CurrentComboPoints: &proto.APLValueCurrentComboPoints{}}}
}

func Energy() Value {
	return &proto.APLValue{Value: &proto.APLValue_CurrentEnergy{CurrentEnergy: &proto.APLValueCurrentEnergy{}}}
}

func RemainingTime() Value {
	return &proto.APLValue{Value: &proto.APLValue_RemainingTime{RemainingTime: &proto.APLValueRemainingTime{}}}
}

func AuraIsActive(aura *proto.ActionID) Value {
	return &proto.APLValue{Value: &proto.APLValue_AuraIsActive{AuraIsActive: &proto.APLValueAuraIsActive{AuraId: aura}}}
}

func AuraRemainingTime(aura *proto.ActionID) Value {
	return &proto.APLValue{Value: &proto.APLValue_AuraRemainingTime{AuraRemainingTime: &proto.APLValueAuraRemainingTime{AuraId: aura}}}
}

func SpellIsReady(spell *proto.ActionID) Value {
	return &proto.APLValue{Value: &proto.APLValue_SpellIsReady{SpellIsReady: &proto.APLValueSpellIsReady{SpellId: spell}}}
}

func SpellTimeToReady(spell *proto.ActionID) Value {
	return &proto.APLValue{Value: &proto.APLValue_SpellTimeToReady{SpellTimeToReady: &proto.APLValueSpellTimeToReady{SpellId: spell}}}
}

func SpellUsesRemaining(spell *proto.ActionID, minTimeLeft Value) Value {
	return &proto.APLValue{Value: &proto.APLValue_SpellUsesRemaining{SpellUsesRemaining: &proto.APLValueSpellUsesRemaining{SpellId: spell, MinTimeLeft: minTimeLeft}}}
}

func SpellUsesLostByDelay(spell *proto.ActionID, delay Value, minTimeLeft Value) Value {
	return &proto.APLValue{Value: &proto.APLValue_SpellUsesLostByDelay{SpellUsesLostByDelay: &proto.APLValueSpellUsesLostByDelay{SpellId: spell, Delay: delay, MinTimeLeft: minTimeLeft}}}
}

func SpellID(id int32) *proto.ActionID {
	return &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}}
}

func SpellIDRank(id int32, rank int32) *proto.ActionID {
	return &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}, Rank: rank}
}

func SpellIDTag(id int32, tag int32) *proto.ActionID {
	return &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}, Tag: tag}
}

func ItemID(id int32) *proto.ActionID {
	return &proto.ActionID{RawId: &proto.ActionID_ItemId{ItemId: id}}
}

// Cast is one priority list line. A nil condition means always.
func Cast(spell *proto.ActionID, condition Value, notes string) *proto.APLListItem {
	return &proto.APLListItem{
		Notes: notes,
		Action: &proto.APLAction{
			Condition: condition,
			Action:    &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: spell}},
		},
	}
}

func AutocastOtherCooldowns(condition Value, notes string) *proto.APLListItem {
	return &proto.APLListItem{
		Notes: notes,
		Action: &proto.APLAction{
			Condition: condition,
			Action:    &proto.APLAction_AutocastOtherCooldowns{AutocastOtherCooldowns: &proto.APLActionAutocastOtherCooldowns{}},
		},
	}
}

func Rotation(items ...*proto.APLListItem) *proto.APLRotation {
	return &proto.APLRotation{Type: proto.APLRotation_TypeAPL, PriorityList: items}
}

// Prepull casts a spell that many seconds before the pull.
func Prepull(spell *proto.ActionID, secondsBefore float64) *proto.APLPrepullAction {
	return &proto.APLPrepullAction{
		Action:    &proto.APLAction{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: spell}}},
		DoAtValue: Seconds(-secondsBefore),
	}
}

func SpellCurrentCost(spell *proto.ActionID) Value {
	return &proto.APLValue{Value: &proto.APLValue_SpellCurrentCost{SpellCurrentCost: &proto.APLValueSpellCurrentCost{SpellId: spell}}}
}

// TargetAuraIsActive checks a debuff the player put on the current target.
func TargetAuraIsActive(aura *proto.ActionID) Value {
	return &proto.APLValue{Value: &proto.APLValue_AuraIsActive{AuraIsActive: &proto.APLValueAuraIsActive{
		SourceUnit: &proto.UnitReference{Type: proto.UnitReference_CurrentTarget},
		AuraId:     aura,
	}}}
}

func DotIsActive(spell *proto.ActionID) Value {
	return &proto.APLValue{Value: &proto.APLValue_DotIsActive{DotIsActive: &proto.APLValueDotIsActive{SpellId: spell}}}
}

func DotRemainingTime(spell *proto.ActionID) Value {
	return &proto.APLValue{Value: &proto.APLValue_DotRemainingTime{DotRemainingTime: &proto.APLValueDotRemainingTime{SpellId: spell}}}
}

func AuraNumStacks(aura *proto.ActionID) Value {
	return &proto.APLValue{Value: &proto.APLValue_AuraNumStacks{AuraNumStacks: &proto.APLValueAuraNumStacks{AuraId: aura}}}
}

// TalentPoints is our points in a talent, by its field name in the class's talents.
func TalentPoints(talent string) Value {
	return &proto.APLValue{Value: &proto.APLValue_TalentPoints{TalentPoints: &proto.APLValueTalentPoints{Talent: talent}}}
}

// InMeleeRange is false only in PvP mode, while the player is out of melee range.
func InMeleeRange() Value {
	return &proto.APLValue{Value: &proto.APLValue_InMeleeRange{InMeleeRange: &proto.APLValueInMeleeRange{}}}
}

func CurrentManaPercent() Value {
	return &proto.APLValue{Value: &proto.APLValue_CurrentManaPercent{CurrentManaPercent: &proto.APLValueCurrentManaPercent{}}}
}
