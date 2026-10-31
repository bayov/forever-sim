import { APLValue, APLValueCompare_ComparisonOperator as ComparisonOperator, APLValueMath_MathOperator as MathOperator } from '../../proto/apl';
import { ActionID as ActionIdProto, UnitReference, UnitReference_Type } from '../../proto/common';
import { ActionId } from '../../proto_utils/action_id';
import { aplValueKindLabel } from './apl_values';

// A condition as one line of text, like "Mobs On You + Mobs Running In ≥ 1 and Health (%) < 0.3".
//
// A collapsed row of the rotation editor shows it in place of the condition's dropdowns. We name a
// value by its kind's label in the editor, with the spell, item or aura it points at and the unit
// it reads from when that isn't us ("Aura Active · Flurry", "Health (%) · Current Target"). Other
// fields, like a totem type, stay out, so the line stays short. The editor shows them when we
// expand the row.
export async function aplValueSummary(value: APLValue | undefined): Promise<string> {
	if (!value?.value.oneofKind) return '';
	const v = value.value;
	switch (v.oneofKind) {
		case 'const':
			return v.const.val;
		case 'and':
			return (await Promise.all(v.and.vals.map(val => grouped(val, 'or')))).filter(s => s).join(' and ');
		case 'or':
			return (await Promise.all(v.or.vals.map(val => grouped(val, 'and')))).filter(s => s).join(' or ');
		case 'not':
			return `not ${await grouped(v.not.val, 'and', 'or', 'cmp', 'math')}`;
		case 'cmp': {
			// The editor's percentages are fractions ("Health (%) < 0.4"), and we read them as 40%.
			const lhs = await aplValueSummary(v.cmp.lhs);
			let rhs = await aplValueSummary(v.cmp.rhs);
			if (lhs.includes('(%)') && v.cmp.rhs?.value.oneofKind == 'const' && /^-?\d*\.?\d+$/.test(rhs)) {
				rhs = `${Math.round(parseFloat(rhs) * 1000) / 10}%`;
			}
			return `${lhs} ${COMPARISON[v.cmp.op] ?? '?'} ${rhs}`;
		}
		case 'math':
			return `${await grouped(v.math.lhs, 'math')} ${MATH[v.math.op] ?? '?'} ${await grouped(v.math.rhs, 'math')}`;
		case 'max':
		case 'min': {
			const vals = v.oneofKind == 'max' ? v.max.vals : v.min.vals;
			return `${aplValueKindLabel(v.oneofKind)}(${(await Promise.all(vals.map(aplValueSummary))).join(', ')})`;
		}
	}
	const fields = Object.values((v as Record<string, unknown>)[v.oneofKind] ?? {});
	const parts = (await Promise.all(fields.map(fieldSummary))).filter(s => s);
	return [aplValueKindLabel(v.oneofKind), ...parts].join(' · ');
}

// A value inside another one, in parentheses when it is one of the given kinds, so "a or b" inside
// an "All of" keeps its meaning.
async function grouped(value: APLValue | undefined, ...kinds: Array<APLValue['value']['oneofKind']>): Promise<string> {
	const text = await aplValueSummary(value);
	return value && kinds.includes(value.value.oneofKind) ? `(${text})` : text;
}

async function fieldSummary(field: unknown): Promise<string> {
	if (!field || typeof field != 'object') return '';
	if ('rawId' in field) {
		const id = ActionId.fromProto(field as ActionIdProto);
		return id.anyId() ? (await id.fill()).name : '';
	}
	if ('value' in field && 'oneofKind' in ((field as APLValue).value ?? {})) {
		return aplValueSummary(field as APLValue);
	}
	if ('type' in field && 'index' in field) {
		return UNITS[(field as UnitReference).type] ?? '';
	}
	return '';
}

const COMPARISON: Partial<Record<ComparisonOperator, string>> = {
	[ComparisonOperator.OpEq]: '=',
	[ComparisonOperator.OpNe]: '≠',
	[ComparisonOperator.OpGe]: '≥',
	[ComparisonOperator.OpGt]: '>',
	[ComparisonOperator.OpLe]: '≤',
	[ComparisonOperator.OpLt]: '<',
};

const MATH: Partial<Record<MathOperator, string>> = {
	[MathOperator.OpAdd]: '+',
	[MathOperator.OpSub]: '-',
	[MathOperator.OpMul]: '×',
	[MathOperator.OpDiv]: '/',
};

// The units a value can read from, when it isn't us.
const UNITS: Partial<Record<UnitReference_Type, string>> = {
	[UnitReference_Type.Target]: 'Target',
	[UnitReference_Type.Pet]: 'Pet',
	[UnitReference_Type.CurrentTarget]: 'Current Target',
	[UnitReference_Type.AllPlayers]: 'All Players',
	[UnitReference_Type.AllTargets]: 'All Targets',
	[UnitReference_Type.PullTarget]: 'Pull Target',
};
