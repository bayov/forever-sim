import { delegate } from 'tippy.js';

import { ActionId } from '../../proto_utils/action_id';
import { formatToNumber } from '../../utils';

// Which way of a change is good. DPS going up is good, damage taken going up is bad, and casts
// going up is neither.
export type Better = 'higher' | 'lower' | 'none';

export interface CompareColumn<T> {
	name: string;
	tooltip?: string;
	value: (metric: T) => number;
	format: (value: number) => string;
	better?: Better;
	// The main column gets a bar for each run under its numbers, so we see their sizes at a
	// glance, like the Damage Done bars of the Results tab.
	bars?: boolean;
	// The footer adds the column up, like the DPS of every ability.
	total?: boolean;
}

// One row: a spell, an aura or a pet, with its metrics in each run. A row that's only in one
// run is new or gone.
export interface CompareRow<T> {
	key: string;
	actionId: ActionId;
	name: string;
	// Auras link to their buff's tooltip.
	isAura?: boolean;
	a?: T;
	b?: T;
}

// How a column sorts: by the compared run's value, or by the size of the change.
type SortMode = 'value' | 'change';

// A change smaller than this share of the value shows muted, so the rows that really moved
// stand out. The sim has no spread per ability to test the change against.
const SMALL_CHANGE = 0.005;

// A table of spells or auras with the compared run's values and the change from the baseline.
//
// Each cell shows the compared run's value, and under it the change from the baseline, green
// when it's better and red when it's worse. Hovering a cell shows both values. Clicking a
// header sorts by its values, and clicking it again sorts by the biggest changes.
export class CompareTable<T> {
	readonly rootElem: HTMLElement;
	private readonly columns: CompareColumn<T>[];
	private readonly headElem: HTMLElement;
	private readonly bodyElem: HTMLElement;
	private readonly footElem: HTMLElement;
	private rows: CompareRow<T>[] = [];
	private sortColumn: number;
	private sortMode: SortMode = 'value';
	private changedOnly = false;

	constructor(columns: CompareColumn<T>[], sortColumn = 0) {
		this.columns = columns;
		this.sortColumn = sortColumn;
		this.rootElem = (
			<div className="cmp-table-wrap">
				<table className="metrics-table cmp-table">
					<thead>
						<tr className="metrics-table-header-row" />
					</thead>
					<tbody className="metrics-table-body" />
					<tfoot className="metrics-table-footer" />
				</table>
			</div>
		) as HTMLElement;
		this.headElem = this.rootElem.querySelector('thead tr')!;
		this.bodyElem = this.rootElem.querySelector('tbody')!;
		this.footElem = this.rootElem.querySelector('tfoot')!;
		delegate(this.rootElem, { target: '[data-tippy-content]', allowHTML: false });
		this.renderHeader();
	}

	setRows(rows: CompareRow<T>[]) {
		this.rows = rows;
		this.renderBody();
	}

	get isEmpty(): boolean {
		return !this.rows.length;
	}

	// Leaves out the rows where nothing we show changed, like the buffs that are up the whole
	// fight in both runs.
	setChangedOnly(changedOnly: boolean) {
		this.changedOnly = changedOnly;
		this.renderBody();
	}

	private hasChange(row: CompareRow<T>): boolean {
		if (!row.a || !row.b) return true;
		return this.columns.some(column => column.format(column.value(row.a!)) !== column.format(column.value(row.b!)));
	}

	private renderHeader() {
		this.headElem.replaceChildren(
			<th className="metrics-table-header-cell">Name</th>,
			...this.columns.map((column, i) => {
				const cell = (
					<th className="metrics-table-header-cell cmp-sortable">
						<span>{column.name}</span>
					</th>
				) as HTMLElement;
				if (i === this.sortColumn) cell.dataset.sort = this.sortMode;
				const hint = this.sortMode === 'change' && i === this.sortColumn ? 'Sorted by the biggest change' : 'Click to sort, again to sort by change';
				cell.dataset.tippyContent = column.tooltip ? `${column.tooltip} ${hint}.` : `${hint}.`;
				cell.addEventListener('click', () => {
					if (this.sortColumn === i) this.sortMode = this.sortMode === 'value' ? 'change' : 'value';
					else {
						this.sortColumn = i;
						this.sortMode = 'value';
					}
					this.renderHeader();
					this.renderBody();
				});
				return cell;
			}),
		);
	}

	private renderBody() {
		const column = this.columns[this.sortColumn];
		const valueOf = (metric: T | undefined) => (metric ? column.value(metric) : 0);
		const sortKey = (row: CompareRow<T>) =>
			this.sortMode === 'value' ? (row.b ? valueOf(row.b) : valueOf(row.a)) : Math.abs(valueOf(row.b) - valueOf(row.a));
		const shown = this.changedOnly ? this.rows.filter(row => this.hasChange(row)) : this.rows;
		const rows = [...shown].sort((x, y) => sortKey(y) - sortKey(x));

		// The bars of a column share one scale, the biggest value in either run.
		const barMax = this.columns.map(column =>
			column.bars ? Math.max(0, ...this.rows.flatMap(row => [row.a, row.b].filter(m => m !== undefined).map(m => column.value(m!)))) : 0,
		);

		if (!rows.length) {
			const text = this.rows.length ? 'Nothing here changed between the runs.' : 'Nothing here in either run.';
			this.bodyElem.replaceChildren(
				<tr className="cmp-table-empty">
					<td colSpan={this.columns.length + 1}>{text}</td>
				</tr>,
			);
			this.footElem.replaceChildren();
			return;
		}

		this.bodyElem.replaceChildren(
			...rows.map(row => (
				<tr className={row.a && row.b ? '' : row.b ? 'cmp-row-new' : 'cmp-row-gone'}>
					<td>{this.nameCell(row)}</td>
					{this.columns.map((column, i) =>
						this.valueCell(column, row.a ? column.value(row.a) : undefined, row.b ? column.value(row.b) : undefined, barMax[i]),
					)}
				</tr>
			)),
		);

		const totals = this.columns.map(column => {
			if (!column.total) return null;
			const sum = (side: 'a' | 'b') => this.rows.reduce((total, row) => total + (row[side] ? column.value(row[side]!) : 0), 0);
			return { a: sum('a'), b: sum('b') };
		});
		this.footElem.replaceChildren(
			...(totals.some(total => total)
				? [
						<tr className="metrics-table-footer-row">
							<td>Total</td>
							{this.columns.map((column, i) => (totals[i] ? this.valueCell(column, totals[i]!.a, totals[i]!.b, 0) : <td />))}
						</tr>,
				  ]
				: []),
		);
	}

	private nameCell(row: CompareRow<T>): Element {
		const icon = (<a className="metrics-action-icon" />) as HTMLAnchorElement;
		row.actionId.setBackgroundAndHref(icon);
		row.actionId.setWowheadDataset(icon, { useBuffAura: !!row.isAura });
		const tag = !row.a ? 'new' : !row.b ? 'gone' : null;
		return (
			<div className="metrics-action">
				{icon}
				<span className="metrics-action-name text-truncate">{row.name}</span>
				{tag ? (
					<span className={`cmp-tag cmp-tag-${tag}`} dataset={{ tippyContent: tag === 'new' ? 'Only in the compared run' : 'Only in the baseline' }}>
						{tag}
					</span>
				) : null}
			</div>
		);
	}

	private valueCell(column: CompareColumn<T>, a: number | undefined, b: number | undefined, barMax: number): Element {
		const numbers = (
			<span className="cmp-numbers">
				<span className="cmp-value">{b === undefined ? '-' : column.format(b)}</span>
				{changeElem(a ?? 0, b ?? 0, column.format, column.better ?? 'none', true)}
			</span>
		);
		// The bars go in front of the numbers, the baseline's over the compared run's.
		const cell = (
			<td className={column.bars ? 'cmp-cell cmp-cell-bars' : 'cmp-cell'}>
				{column.bars ? (
					<span className="cmp-cell-inner">
						<span className="cmp-bars">
							<span className="cmp-bar cmp-bar-a" style={{ width: `${barMax > 0 ? (Math.max(0, a ?? 0) / barMax) * 100 : 0}%` }} />
							<span className="cmp-bar cmp-bar-b" style={{ width: `${barMax > 0 ? (Math.max(0, b ?? 0) / barMax) * 100 : 0}%` }} />
						</span>
						{numbers}
					</span>
				) : (
					numbers
				)}
			</td>
		) as HTMLElement;
		cell.dataset.tippyContent = `Baseline ${a === undefined ? '-' : column.format(a)} → Compared ${b === undefined ? '-' : column.format(b)}`;
		return cell;
	}
}

// The change from a to b, like '+12.3', colored by whether it's better. A change that's tiny
// next to the values shows muted. No change (at the digits we show) is '±0', or nothing in a
// table, where a column of '±0' would only be noise.
export function changeElem(a: number, b: number, format: (value: number) => string, better: Better, blankWhenSame = false): Element {
	const className = 'cmp-change';
	const diff = b - a;
	const scale = Math.max(Math.abs(a), Math.abs(b));
	if (!diff || format(Math.abs(diff)) === format(0)) return <span className={`${className} cmp-flat`}>{blankWhenSame ? '' : '±0'}</span>;
	const small = scale > 0 && Math.abs(diff) / scale < SMALL_CHANGE;
	const tone = small || better === 'none' ? 'cmp-flat' : diff > 0 === (better === 'higher') ? 'cmp-better' : 'cmp-worse';
	return <span className={`${className} ${tone}`}>{`${diff > 0 ? '+' : '-'}${format(Math.abs(diff))}`}</span>;
}

export const formatFixed =
	(digits: number) =>
	(value: number): string =>
		formatToNumber(value, { minimumFractionDigits: digits, maximumFractionDigits: digits });

export const formatPercent =
	(digits: number) =>
	(value: number): string =>
		`${formatToNumber(value, { minimumFractionDigits: digits, maximumFractionDigits: digits })}%`;
