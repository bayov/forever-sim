package main

import (
	"strconv"

	"github.com/wowsims/classic/sim/core/proto"
)

// Small builders for APL protos, so a rotation template reads like the condition it
// expresses rather than like nested struct literals.

type value = *proto.APLValue

func constant(val string) value {
	return &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: val}}}
}

func num(v float64) value {
	return constant(strconv.FormatFloat(v, 'f', -1, 64))
}

func seconds(v float64) value {
	return constant(strconv.FormatFloat(v, 'f', -1, 64) + "s")
}

// and drops nil operands, so a template can pass nil for a rule that is switched off.
// With every rule off it is nil too, so the line has no condition at all.
func and(vals ...value) value {
	if len(nonNil(vals)) == 0 {
		return nil
	}
	return &proto.APLValue{Value: &proto.APLValue_And{And: &proto.APLValueAnd{Vals: nonNil(vals)}}}
}

func or(vals ...value) value {
	return &proto.APLValue{Value: &proto.APLValue_Or{Or: &proto.APLValueOr{Vals: nonNil(vals)}}}
}

func nonNil(vals []value) []value {
	var out []value
	for _, v := range vals {
		if v != nil {
			out = append(out, v)
		}
	}
	return out
}

func not(val value) value {
	return &proto.APLValue{Value: &proto.APLValue_Not{Not: &proto.APLValueNot{Val: val}}}
}

func cmp(lhs value, op proto.APLValueCompare_ComparisonOperator, rhs value) value {
	return &proto.APLValue{Value: &proto.APLValue_Cmp{Cmp: &proto.APLValueCompare{Op: op, Lhs: lhs, Rhs: rhs}}}
}

func ge(lhs, rhs value) value { return cmp(lhs, proto.APLValueCompare_OpGe, rhs) }
func gt(lhs, rhs value) value { return cmp(lhs, proto.APLValueCompare_OpGt, rhs) }
func le(lhs, rhs value) value { return cmp(lhs, proto.APLValueCompare_OpLe, rhs) }
func lt(lhs, rhs value) value { return cmp(lhs, proto.APLValueCompare_OpLt, rhs) }
func eq(lhs, rhs value) value { return cmp(lhs, proto.APLValueCompare_OpEq, rhs) }

func add(lhs, rhs value) value {
	return &proto.APLValue{Value: &proto.APLValue_Math{Math: &proto.APLValueMath{Op: proto.APLValueMath_OpAdd, Lhs: lhs, Rhs: rhs}}}
}

func sub(lhs, rhs value) value {
	return &proto.APLValue{Value: &proto.APLValue_Math{Math: &proto.APLValueMath{Op: proto.APLValueMath_OpSub, Lhs: lhs, Rhs: rhs}}}
}

func mainHandTimeToNext() value {
	return &proto.APLValue{Value: &proto.APLValue_AutoTimeToNext{AutoTimeToNext: &proto.APLValueAutoTimeToNext{AutoType: proto.APLValueAutoTimeToNext_MainHand}}}
}

func mainHandSwingTime() value {
	return &proto.APLValue{Value: &proto.APLValue_AutoSwingTime{AutoSwingTime: &proto.APLValueAutoSwingTime{AutoType: proto.APLValueAutoSwingTime_MainHand}}}
}

func comboPoints() value {
	return &proto.APLValue{Value: &proto.APLValue_CurrentComboPoints{CurrentComboPoints: &proto.APLValueCurrentComboPoints{}}}
}

func energy() value {
	return &proto.APLValue{Value: &proto.APLValue_CurrentEnergy{CurrentEnergy: &proto.APLValueCurrentEnergy{}}}
}

func remainingTime() value {
	return &proto.APLValue{Value: &proto.APLValue_RemainingTime{RemainingTime: &proto.APLValueRemainingTime{}}}
}

func auraIsActive(aura *proto.ActionID) value {
	return &proto.APLValue{Value: &proto.APLValue_AuraIsActive{AuraIsActive: &proto.APLValueAuraIsActive{AuraId: aura}}}
}

func auraRemainingTime(aura *proto.ActionID) value {
	return &proto.APLValue{Value: &proto.APLValue_AuraRemainingTime{AuraRemainingTime: &proto.APLValueAuraRemainingTime{AuraId: aura}}}
}

func spellIsReady(spell *proto.ActionID) value {
	return &proto.APLValue{Value: &proto.APLValue_SpellIsReady{SpellIsReady: &proto.APLValueSpellIsReady{SpellId: spell}}}
}

func spellTimeToReady(spell *proto.ActionID) value {
	return &proto.APLValue{Value: &proto.APLValue_SpellTimeToReady{SpellTimeToReady: &proto.APLValueSpellTimeToReady{SpellId: spell}}}
}

func spellUsesRemaining(spell *proto.ActionID, minTimeLeft value) value {
	return &proto.APLValue{Value: &proto.APLValue_SpellUsesRemaining{SpellUsesRemaining: &proto.APLValueSpellUsesRemaining{SpellId: spell, MinTimeLeft: minTimeLeft}}}
}

func spellUsesLostByDelay(spell *proto.ActionID, delay value, minTimeLeft value) value {
	return &proto.APLValue{Value: &proto.APLValue_SpellUsesLostByDelay{SpellUsesLostByDelay: &proto.APLValueSpellUsesLostByDelay{SpellId: spell, Delay: delay, MinTimeLeft: minTimeLeft}}}
}

func spellID(id int32) *proto.ActionID {
	return &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}}
}

func spellIDRank(id int32, rank int32) *proto.ActionID {
	return &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}, Rank: rank}
}

func spellIDTag(id int32, tag int32) *proto.ActionID {
	return &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}, Tag: tag}
}

func itemID(id int32) *proto.ActionID {
	return &proto.ActionID{RawId: &proto.ActionID_ItemId{ItemId: id}}
}

// cast is one priority list line. A nil condition means always.
func cast(spell *proto.ActionID, condition value, notes string) *proto.APLListItem {
	return &proto.APLListItem{
		Notes: notes,
		Action: &proto.APLAction{
			Condition: condition,
			Action:    &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: spell}},
		},
	}
}

func autocastOtherCooldowns(condition value, notes string) *proto.APLListItem {
	return &proto.APLListItem{
		Notes: notes,
		Action: &proto.APLAction{
			Condition: condition,
			Action:    &proto.APLAction_AutocastOtherCooldowns{AutocastOtherCooldowns: &proto.APLActionAutocastOtherCooldowns{}},
		},
	}
}

func rotation(items ...*proto.APLListItem) *proto.APLRotation {
	return &proto.APLRotation{Type: proto.APLRotation_TypeAPL, PriorityList: items}
}

// prepull casts a spell that many seconds before the pull.
func prepull(spell *proto.ActionID, secondsBefore float64) *proto.APLPrepullAction {
	return &proto.APLPrepullAction{
		Action:    &proto.APLAction{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: spell}}},
		DoAtValue: seconds(-secondsBefore),
	}
}

func spellCurrentCost(spell *proto.ActionID) value {
	return &proto.APLValue{Value: &proto.APLValue_SpellCurrentCost{SpellCurrentCost: &proto.APLValueSpellCurrentCost{SpellId: spell}}}
}

// targetAuraIsActive checks a debuff the player put on the current target.
func targetAuraIsActive(aura *proto.ActionID) value {
	return &proto.APLValue{Value: &proto.APLValue_AuraIsActive{AuraIsActive: &proto.APLValueAuraIsActive{
		SourceUnit: &proto.UnitReference{Type: proto.UnitReference_CurrentTarget},
		AuraId:     aura,
	}}}
}

func dotIsActive(spell *proto.ActionID) value {
	return &proto.APLValue{Value: &proto.APLValue_DotIsActive{DotIsActive: &proto.APLValueDotIsActive{SpellId: spell}}}
}

func dotRemainingTime(spell *proto.ActionID) value {
	return &proto.APLValue{Value: &proto.APLValue_DotRemainingTime{DotRemainingTime: &proto.APLValueDotRemainingTime{SpellId: spell}}}
}

func auraNumStacks(aura *proto.ActionID) value {
	return &proto.APLValue{Value: &proto.APLValue_AuraNumStacks{AuraNumStacks: &proto.APLValueAuraNumStacks{AuraId: aura}}}
}

// inMeleeRange is false only in PvP mode, while the player is out of melee range.
func inMeleeRange() value {
	return &proto.APLValue{Value: &proto.APLValue_InMeleeRange{InMeleeRange: &proto.APLValueInMeleeRange{}}}
}

func currentManaPercent() value {
	return &proto.APLValue{Value: &proto.APLValue_CurrentManaPercent{CurrentManaPercent: &proto.APLValueCurrentManaPercent{}}}
}
