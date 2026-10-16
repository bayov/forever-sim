import tippy from 'tippy.js';

import { formatToNumber } from '../../utils.js';
import { EventFilter } from './events_filter.js';
import { formatClock, formatSpan, getLevelingSession, LevelingSession, MobLife, mobName, Pull } from './leveling_session.js';
import { ResultComponent, ResultComponentConfig, SimResultData } from './result_component.js';

export interface PullsTableConfig extends ResultComponentConfig {
	filter: EventFilter;
	// Called when we click a pull, to show only it in the other views.
	onPick: (start: number, end: number) => void;
}

interface Column {
	label: string;
	tooltip: string;
	// What the column sorts by. Columns without one don't sort.
	sortValue?: (pull: Pull) => number;
	cell: (pull: Pull) => Element | string;
	footer: (pulls: Array<Pull>, session: LevelingSession) => Element | string;
}

// A leveling session's pulls, one row each: what we pulled and how, how long we fought, what
// it cost us and the rest after. Clicking a row shows only that pull in the other views, and
// the row of the pull we look at stands out.
export class PullsTable extends ResultComponent {
	private readonly filter: EventFilter;
	private readonly onPick: (start: number, end: number) => void;
	private readonly headRow: HTMLTableRowElement;
	private readonly body: HTMLTableSectionElement;
	private readonly footRow: HTMLTableRowElement;
	private readonly emptyElem: HTMLElement;

	private session: LevelingSession | null = null;
	private rows = new Map<Pull, HTMLTableRowElement>();
	private sortColumn = 0;
	private sortDescending = false;

	constructor(config: PullsTableConfig) {
		config.rootCssClass = 'pulls-table-root';
		super(config);
		this.filter = config.filter;
		this.onPick = config.onPick;

		this.headRow = (<tr className="metrics-table-header-row" />) as HTMLTableRowElement;
		this.body = (<tbody className="metrics-table-body" />) as HTMLTableSectionElement;
		this.footRow = (<tr className="metrics-table-footer-row" />) as HTMLTableRowElement;
		this.emptyElem = (<div className="events-empty hide">This run has no pulls. The Pulls view is for the Leveling mode.</div>) as HTMLElement;
		this.rootElem.appendChild(
			<>
				<div className="pulls-table-scroll">
					<table className="metrics-table pulls-table">
						<thead className="metrics-table-header">{this.headRow}</thead>
						{this.body}
						<tfoot className="metrics-table-footer">{this.footRow}</tfoot>
					</table>
				</div>
				{this.emptyElem}
			</>,
		);

		COLUMNS.forEach((column, i) => {
			const cell = (
				<th className="metrics-table-header-cell">
					<span className="tablesorter-header-inner">{column.label}</span>
				</th>
			) as HTMLTableCellElement;
			tippy(cell, { content: column.tooltip });
			if (column.sortValue) {
				cell.addEventListener('click', () => {
					if (this.sortColumn == i) this.sortDescending = !this.sortDescending;
					else {
						this.sortColumn = i;
						// The first column reads in session order. The others start with the most.
						this.sortDescending = i != 0;
					}
					this.render();
				});
			} else {
				cell.classList.add('unsortable');
			}
			this.headRow.appendChild(cell);
		});

		this.filter.changeEmitter.on(() => this.updateSelected());
	}

	onSimResult(resultData: SimResultData) {
		const session = getLevelingSession(resultData.result);
		if (session === this.session) return;
		this.session = session;
		this.rows = new Map();
		(session?.pulls ?? []).forEach(pull => this.rows.set(pull, this.makeRow(pull)));
		this.render();
	}

	private render() {
		const session = this.session;
		const pulls = session?.pulls ?? [];
		this.emptyElem.classList.toggle('hide', pulls.length > 0);
		this.rootElem.querySelector('.pulls-table-scroll')!.classList.toggle('hide', pulls.length == 0);

		const column = COLUMNS[this.sortColumn];
		const sorted = [...pulls];
		if (column.sortValue) {
			const value = column.sortValue;
			sorted.sort((a, b) => (value(a) - value(b)) * (this.sortDescending ? -1 : 1) || a.number - b.number);
		}
		this.body.replaceChildren(...sorted.map(pull => this.rows.get(pull)!));

		this.headRow.querySelectorAll('th').forEach((cell, i) => {
			if (i == this.sortColumn) cell.setAttribute('aria-sort', this.sortDescending ? 'descending' : 'ascending');
			else cell.removeAttribute('aria-sort');
		});
		this.footRow.replaceChildren(...(session ? COLUMNS.map(column => <td>{column.footer(pulls, session)}</td>) : []));
		this.updateSelected();
	}

	private makeRow(pull: Pull): HTMLTableRowElement {
		const row = (
			<tr className={`pulls-table-row ${pull.died ? 'died' : ''}`}>
				{COLUMNS.map(column => (
					<td>{column.cell(pull)}</td>
				))}
			</tr>
		) as HTMLTableRowElement;
		row.addEventListener('click', () => this.onPick(pull.start, pull.end));
		row.querySelectorAll('[data-tip]').forEach(elem => tippy(elem, { content: (elem as HTMLElement).dataset.tip!, allowHTML: true }));
		return row;
	}

	// The row of the pull the filter's range is on.
	private updateSelected() {
		const { from, to } = this.filter;
		this.rows.forEach((row, pull) => {
			const selected = from != null && to != null && Math.abs(pull.start - from) < 1e-6 && Math.abs(pull.end - to) < 1e-6;
			row.classList.toggle('selected', selected);
		});
	}
}

const percent = (share: number | null): string => (share == null ? '' : `${Math.round(share * 100)}%`);
const amount = (value: number): string => formatToNumber(value, { maximumFractionDigits: 0 });
const average = (pulls: Array<Pull>, value: (pull: Pull) => number): number =>
	pulls.length ? pulls.reduce((sum, pull) => sum + value(pull), 0) / pulls.length : 0;
const total = (pulls: Array<Pull>, value: (pull: Pull) => number): number => pulls.reduce((sum, pull) => sum + value(pull), 0);
const combatSeconds = (pull: Pull): number => pull.end - pull.start;
const muted = (text: string) => <span className="pulls-table-muted">{text}</span>;

// A chain pull is a mob we pulled while we already fought. A mob we found while fighting the
// last pull and pulled after it is the first of its own pull.
const chainPulled = (pull: Pull, mob: MobLife): boolean => mob.pulledAt != null && mob.pulledAt > pull.start + 1e-6;

// The mobs of a pull: how many, their levels, a link for the chain pulls, and an icon for how
// the first one was pulled.
const mobsCell = (pull: Pull): Element => {
	const first = pull.mobs[0];
	const chained = pull.mobs.filter(mob => chainPulled(pull, mob)).length;
	const names = pull.mobs
		.map(mob => `${mobName(mob)} (level ${mob.level})${chainPulled(pull, mob) ? ', chain pulled' : ''}${mob.diedAt == null ? ', not killed' : ''}`)
		.join('<br>');
	return (
		<span className="pulls-table-mobs">
			<span className="pulls-table-mob-count" dataset={{ tip: names || 'No mobs' }}>
				{pull.mobs.length}
			</span>
			<span className="pulls-table-levels">{pull.mobs.map(mob => mob.level).join(' ')}</span>
			{chained > 0 && (
				<span className="pulls-table-chain" dataset={{ tip: `${chained} of them found and pulled while we fought the others` }}>
					<i className="fas fa-link" />
					{chained}
				</span>
			)}
			{first?.pulledBy && (
				<i
					className={`fas ${first.pulledBy == 'spell' ? 'fa-wand-sparkles' : 'fa-person-walking'} pulls-table-pulled-by`}
					dataset={{
						tip:
							first.pulledBy == 'spell'
								? `Pulled with a spell from ${first.pulledFrom} yards`
								: `Face pulled: it noticed us ${first.pulledFrom} yards away`,
					}}
				/>
			)}
		</span>
	) as Element;
};

const fromTo = (start: number | null, end: number | null, lowest?: number): Element =>
	(
		<span className="pulls-table-from-to">
			{percent(start)}
			<i className="fas fa-arrow-right" />
			{percent(end)}
			{lowest != null && lowest < Math.min(start ?? 1, end ?? 1) - 0.005 && muted(` low ${percent(lowest)}`)}
		</span>
	) as Element;

const restCell = (pull: Pull): Element =>
	(
		<span className="pulls-table-rest">
			{pull.restSeconds > 0 ? formatSpan(pull.restSeconds) : muted('none')}
			{pull.ate && <i className="fas fa-drumstick-bite" dataset={{ tip: 'Ate' }} />}
			{pull.drank && <i className="fas fa-mug-hot" dataset={{ tip: 'Drank' }} />}
			{pull.bandaged && <i className="fas fa-bandage" dataset={{ tip: 'Bandaged' }} />}
		</span>
	) as Element;

const COLUMNS: Array<Column> = [
	{
		label: 'Pull',
		tooltip: 'The pulls in session order.',
		sortValue: pull => pull.number,
		cell: pull => <span className="pulls-table-number">{pull.number}</span>,
		footer: pulls => `${pulls.length} pulls`,
	},
	{
		label: 'Start',
		tooltip: 'When the first mob was pulled.',
		sortValue: pull => pull.start,
		cell: pull => formatClock(pull.start),
		footer: () => '',
	},
	{
		label: 'Mobs',
		tooltip:
			'The mobs we fought and their levels. The link counts the chain pulls, the mobs we found and pulled while we fought. The last icon is how we pulled the first mob: with a spell, or by walking up to it.',
		sortValue: pull => pull.mobs.length,
		cell: mobsCell,
		footer: pulls =>
			`${formatToNumber(
				average(pulls, pull => pull.mobs.length),
				{ maximumFractionDigits: 2 },
			)} a pull`,
	},
	{
		label: 'Killed',
		tooltip: "The mobs that died. One we left when we died or the session ended doesn't count.",
		sortValue: pull => pull.kills,
		cell: pull => String(pull.kills),
		footer: pulls => `${total(pulls, pull => pull.kills)} total`,
	},
	{
		label: 'Combat',
		tooltip: 'How long we fought, from the first mob pulled until the last died or we did.',
		sortValue: combatSeconds,
		cell: pull => formatSpan(combatSeconds(pull)),
		footer: pulls => `${formatSpan(average(pulls, combatSeconds))} avg`,
	},
	{
		label: 'Taken',
		tooltip: 'Damage the mobs did to us.',
		sortValue: pull => pull.damageTaken,
		cell: pull => amount(pull.damageTaken),
		footer: pulls => `${amount(average(pulls, pull => pull.damageTaken))} avg`,
	},
	{
		label: 'Done',
		tooltip: 'Damage we did to the mobs, our pets and totems too.',
		sortValue: pull => pull.damageDone,
		cell: pull => amount(pull.damageDone),
		footer: pulls => `${amount(average(pulls, pull => pull.damageDone))} avg`,
	},
	{
		label: 'Health',
		tooltip: 'Our health when the pull started and when it ended, and the lowest it went when that was lower.',
		sortValue: pull => pull.lowestHealth,
		cell: pull => fromTo(pull.healthStart, pull.healthEnd, pull.lowestHealth),
		footer: pulls => `low ${percent(average(pulls, pull => pull.lowestHealth))} avg`,
	},
	{
		label: 'Mana',
		tooltip: 'Our mana when the pull started and when it ended.',
		sortValue: pull => pull.manaEnd ?? 0,
		cell: pull => (pull.manaStart == null ? muted('') : fromTo(pull.manaStart, pull.manaEnd)),
		footer: pulls => (pulls[0]?.manaEnd == null ? '' : `end ${percent(average(pulls, pull => pull.manaEnd ?? 0))} avg`),
	},
	{
		label: 'Rest after',
		tooltip: 'How long we sat or bandaged before the next pull, and whether we ate, drank or bandaged.',
		sortValue: pull => pull.restSeconds,
		cell: restCell,
		footer: (pulls, session) => `${formatSpan(total(pulls, pull => pull.restSeconds))} total of ${formatClock(session.length)}`,
	},
	{
		label: 'Died',
		tooltip: 'Whether we died in the pull.',
		sortValue: pull => (pull.died ? 1 : 0),
		cell: pull => (pull.died ? <i className="fas fa-skull pulls-table-skull" /> : ''),
		footer: (_, session) => `${session.deaths.length} ${session.deaths.length == 1 ? 'death' : 'deaths'}`,
	},
];
