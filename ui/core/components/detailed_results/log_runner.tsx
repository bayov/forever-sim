import { delegate } from 'tippy.js';
import { ref } from 'tsx-vanilla';

import { ResourceType } from '../../proto/api.js';
import { SimLog } from '../../proto_utils/logs_parser';
import { SimResult } from '../../proto_utils/sim_result.js';
import { TypedEvent } from '../../typed_event.js';
import { formatToNumber } from '../../utils.js';
import { EventCategory, EventFilter } from './events_filter.js';
import { badgeOutcomeInLogLine } from './hit_outcome.js';
import {
	formatClock,
	formatSpan,
	getLevelingSession,
	LevelingSession,
	SectionKind,
	SESSION_STATES,
	SessionSection,
	SessionState,
	sectionTitle,
} from './leveling_session.js';
import { ResultComponent, ResultComponentConfig, SimResultData } from './result_component.js';

export interface LogRunnerConfig extends ResultComponentConfig {
	filter: EventFilter;
	// Called when we click the time of an event, to show that moment in the timeline.
	onTimeClick: (time: number) => void;
}

interface LogRow {
	log: SimLog;
	category: EventCategory;
	// The row and its lower case text, made the first time we show or search it. A leveling
	// session has some 30,000 lines, and making them all up front took seconds.
	elem: HTMLTableRowElement | null;
	text: string | null;
	// In a leveling session, the section the line is in, and whether it's one of the lines we
	// fold away in a rest, a search or a corpse run.
	section: SessionSection | null;
	folded: boolean;
}

// The lines of a rest, a search or a corpse run that we fold into one row. They are the health
// and mana we got back and the rotation waiting.
interface Fold {
	section: SessionSection;
	health: number;
	mana: number;
	elem: HTMLTableRowElement;
	label: HTMLElement;
}

// The leveling sections' colors are those of the state they are mostly in.
const SECTION_STATE: Record<SectionKind, SessionState> = {
	pull: 'fighting',
	rest: 'resting',
	search: 'searching',
	death: 'dead',
};

// The iteration's events one per line, oldest first.
//
// In a leveling session the log has a header row at the start of each pull, rest, search and
// death. Clicking one shows only that section. The regen and the rotation's waiting in a rest,
// a search or a corpse run fold into one row each, which opens on a click.
export class LogRunner extends ResultComponent {
	private readonly filter: EventFilter;
	private readonly onTimeClick: (time: number) => void;
	private readonly virtualScroll: CustomVirtualScroll;
	private readonly scrollContainer: HTMLDivElement;
	private readonly emptyElem: HTMLElement;

	private rows: Array<LogRow> = [];
	private rowsKey: SimResult | null = null;
	private session: LevelingSession | null = null;
	private headers = new Map<SessionSection, HTMLTableRowElement>();
	private folds = new Map<SessionSection, Fold>();
	// The sections whose folded lines we opened.
	private readonly expanded = new Set<SessionSection>();
	// A line between the prepull events and the ones after the pull. It's a row of the same height
	// as the others, so the virtual scroll can place it like any row.
	private readonly pullRow = (
		<tr className="log-pull-row">
			<td colSpan={2}>
				<span>Pull</span>
			</td>
		</tr>
	) as HTMLTableRowElement;

	// How many events the filter lets through, out of all of them.
	readonly countEmitter = new TypedEvent<void>('LogRunnerCount');
	shownCount = 0;
	totalCount = 0;

	constructor(config: LogRunnerConfig) {
		config.rootCssClass = 'log-runner-root';
		super(config);
		this.filter = config.filter;
		this.onTimeClick = config.onTimeClick;

		const scrollContainerRef = ref<HTMLDivElement>();
		const contentContainerRef = ref<HTMLTableSectionElement>();
		const emptyRef = ref<HTMLDivElement>();

		this.rootElem.appendChild(
			<>
				<div ref={scrollContainerRef} className="log-runner-scroll">
					<table className="log-runner-table">
						<thead>
							<tr>
								<th className="log-timestamp">Time</th>
								<th>Event</th>
							</tr>
						</thead>
						<tbody ref={contentContainerRef} className="log-runner-logs"></tbody>
					</table>
				</div>
				<div ref={emptyRef} className="events-empty hide">
					No events match the filter.
				</div>
			</>,
		);

		this.scrollContainer = scrollContainerRef.value!;
		this.emptyElem = emptyRef.value!;
		this.virtualScroll = new CustomVirtualScroll({
			scrollContainer: this.scrollContainer,
			contentContainer: contentContainerRef.value!,
			itemHeight: LOG_ROW_HEIGHT,
		});

		// A log can have thousands of rows, so the rows' tooltips are made when we hover them.
		delegate(this.scrollContainer, { target: '[data-tippy-content]' });

		this.filter.changeEmitter.on(() => this.applyFilter());
	}

	onSimResult(resultData: SimResultData) {
		if (this.rowsKey !== resultData.result) {
			this.rowsKey = resultData.result;
			this.session = getLevelingSession(resultData.result);
			this.rootElem.classList.toggle('leveling', !!this.session);
			this.expanded.clear();
			// Cast completed lines repeat the cast began ones, so we leave them out.
			this.rows = resultData.result.logs
				.filter(log => log.raw.trim() != '')
				.filter(log => !log.isCastCompleted())
				.map(log => ({ log, category: EventFilter.categoryOf(log), elem: null, text: null, section: null, folded: false }));
			this.buildSections();
		}
		this.applyFilter();
	}

	// Puts each line in its section, and makes the sections' header and fold rows.
	private buildSections() {
		this.headers = new Map();
		this.folds = new Map();
		const session = this.session;
		if (!session) return;

		const sections = [...session.sections].sort((a, b) => a.start - b.start);
		sections.forEach(section => this.headers.set(section, this.makeHeader(section)));

		let i = 0;
		for (const row of this.rows) {
			const t = row.log.timestamp;
			while (i < sections.length && sections[i].end <= t) i++;
			const section = i < sections.length && sections[i].start <= t ? sections[i] : null;
			row.section = section;
			row.folded = !!section && section.kind != 'pull' && FOLDED_CATEGORIES.includes(row.category);
			if (!row.folded) continue;

			let fold = this.folds.get(section!);
			if (!fold) {
				fold = this.makeFold(section!);
				this.folds.set(section!, fold);
			}
			const log = row.log;
			if (log.isResourceChanged() && log.source?.isPlayer && !log.target) {
				const delta = log.valueAfter - log.valueBefore;
				if (log.resourceType == ResourceType.ResourceTypeHealth) fold.health += delta;
				else if (log.resourceType == ResourceType.ResourceTypeMana) fold.mana += delta;
			}
		}
	}

	private applyFilter() {
		const searching = this.filter.hasSearch();
		const shownRows = this.rows.filter(
			row => this.filter.matchesLogExceptText(row.log, row.category) && (!searching || this.filter.matchesText(this.textOf(row))),
		);
		this.shownCount = shownRows.length;
		this.totalCount = this.rows.length;
		this.emptyElem.classList.toggle('hide', shownRows.length > 0 || this.rows.length == 0);

		const shown = this.session ? this.levelingItems(shownRows) : this.fightItems(shownRows);
		this.virtualScroll.setItems(shown);
		this.countEmitter.emit(TypedEvent.nextEventID());
	}

	private fightItems(shownRows: Array<LogRow>): Array<RowItem> {
		const shown: Array<RowItem> = shownRows.map(row => () => this.elemOf(row));
		const pullIndex = shownRows.findIndex(row => row.log.timestamp >= 0);
		if (pullIndex > 0) shown.splice(pullIndex, 0, this.pullRow);
		return shown;
	}

	// The rows of a leveling session: each section's header before its first line, then the
	// lines, with the folded ones in one row where the first of them would be.
	private levelingItems(shownRows: Array<LogRow>): Array<RowItem> {
		// While we search, only the sections with a match get their header.
		const withRows = this.filter.hasSearch() ? new Set(shownRows.map(row => row.section)) : null;
		const sections = [...this.headers.keys()].filter(
			// A section that only touches the range, like the search that ends as the pull starts, isn't
			// in it.
			section =>
				(this.filter.to == null || section.start < this.filter.to) &&
				(this.filter.from == null || section.end > this.filter.from) &&
				(!withRows || withRows.has(section)),
		);
		const items: Array<RowItem> = [];
		const foldCounts = new Map<SessionSection, number>();
		let next = 0;
		const headersUntil = (t: number) => {
			while (next < sections.length && sections[next].start <= t) items.push(this.headers.get(sections[next++])!);
		};

		for (const row of shownRows) {
			headersUntil(row.log.timestamp);
			// The lines a search finds stay open, folded or not.
			if (row.folded && row.section && !withRows) {
				const count = foldCounts.get(row.section) ?? 0;
				if (count == 0) items.push(this.folds.get(row.section)!.elem);
				foldCounts.set(row.section, count + 1);
				if (!this.expanded.has(row.section)) continue;
			}
			items.push(() => this.elemOf(row));
		}
		headersUntil(Infinity);

		const range = this.filter.from != null && this.filter.to != null ? [this.filter.from, this.filter.to] : null;
		this.headers.forEach((header, section) => {
			const selected = !!range && Math.abs(section.start - range[0]) < 1e-6 && Math.abs(section.end - range[1]) < 1e-6;
			header.classList.toggle('selected', selected);
			const button = header.querySelector('.log-section-button');
			button?.setAttribute('data-tippy-content', selected ? 'Show the whole session' : `Show only ${sectionTitle(section)}`);
		});
		this.folds.forEach((fold, section) => this.updateFold(fold, foldCounts.get(section) ?? 0));
		return items;
	}

	private elemOf(row: LogRow): HTMLTableRowElement {
		if (!row.elem) this.makeRow(row);
		return row.elem!;
	}

	private textOf(row: LogRow): string {
		if (row.text == null) this.makeRow(row);
		return row.text!;
	}

	private makeRow(row: LogRow) {
		const log = row.log;
		const timeRef = ref<HTMLButtonElement>();
		const eventRef = ref<HTMLTableCellElement>();
		const leveling = levelingLine(log);
		const elem = (
			<tr className={leveling ? 'log-session-line' : ''}>
				<td className="log-timestamp">
					<button ref={timeRef} className="log-time-link" type="button" dataset={{ tippyContent: 'Show in timeline', tippyPlacement: 'right' }}>
						{formatEventTime(log.timestamp, this.session ? 2 : 3)}
					</button>
				</td>
				<td ref={eventRef} className="log-event">
					<span className="log-event-text">
						{leveling != null ? (
							<>
								<i className="fas fa-person-hiking log-session-icon" />
								{leveling}
							</>
						) : (
							log.toHTML(false)
						)}
					</span>
				</td>
			</tr>
		) as HTMLTableRowElement;

		timeRef.value!.addEventListener('click', () => this.onTimeClick(log.timestamp));
		if (log.isDamageDealt()) badgeOutcomeInLogLine(eventRef.value!);

		const text = eventRef.value!.textContent?.trim() ?? '';
		eventRef.value!.title = text;

		if (log.actionId) {
			const actionId = log.actionId;
			const hideButton = (
				<button className="log-hide-action" type="button" dataset={{ tippyContent: `Hide ${actionId.name}` }}>
					<i className="fas fa-eye-slash" />
				</button>
			) as HTMLButtonElement;
			hideButton.addEventListener('click', () => this.filter.hide(actionId));
			eventRef.value!.appendChild(hideButton);
		}

		row.elem = elem;
		row.text = text.toLowerCase();
	}

	// A section's header, like "Pull 7  2:05  3 mobs (30, 31, 31), 42 s, died". It's a row of
	// the same height as the others. Clicking it shows only that section, and again the whole
	// session.
	private makeHeader(section: SessionSection): HTMLTableRowElement {
		const buttonRef = ref<HTMLButtonElement>();
		const elem = (
			<tr className={`log-section-row log-section-${section.kind}`}>
				<td colSpan={2}>
					<button ref={buttonRef} className="log-section-button" type="button">
						<span className="log-section-title">{sectionTitle(section)}</span>
						<span className="log-section-time">{formatClock(section.start)}</span>
						<span className="log-section-summary">{sectionSummary(section)}</span>
					</button>
				</td>
			</tr>
		) as HTMLTableRowElement;
		elem.style.setProperty('--section-color', SESSION_STATES[SECTION_STATE[section.kind]].color);
		buttonRef.value!.addEventListener('click', () => {
			const selected = this.filter.from == section.start && this.filter.to == section.end;
			if (selected) this.filter.setRange(null, null);
			else this.filter.setRange(section.start, section.end);
		});
		return elem;
	}

	private makeFold(section: SessionSection): Fold {
		const labelRef = ref<HTMLSpanElement>();
		const elem = (
			<tr className="log-fold-row">
				<td className="log-timestamp"></td>
				<td className="log-event">
					<button className="log-fold-button" type="button">
						<i className="fas fa-caret-right log-fold-caret" />
						<span ref={labelRef}></span>
					</button>
				</td>
			</tr>
		) as HTMLTableRowElement;
		elem.querySelector('button')!.addEventListener('click', () => {
			if (this.expanded.has(section)) this.expanded.delete(section);
			else this.expanded.add(section);
			const scrollTop = this.scrollContainer.scrollTop;
			this.applyFilter();
			this.scrollContainer.scrollTop = scrollTop;
		});
		return { section, health: 0, mana: 0, elem, label: labelRef.value! };
	}

	// "Resting 0:42 to 1:05: +480 health, +600 mana, 230 lines".
	private updateFold(fold: Fold, count: number) {
		const { section } = fold;
		const what = { pull: 'Fighting', rest: 'Resting', search: 'Searching', death: 'Corpse run' }[section.kind];
		const parts = [
			fold.health != 0 ? `${signed(fold.health)} health` : '',
			fold.mana != 0 ? `${signed(fold.mana)} mana` : '',
			`${formatToNumber(count)} ${count == 1 ? 'line' : 'lines'}`,
		].filter(Boolean);
		const open = this.expanded.has(section);
		fold.label.textContent = `${what} ${formatClock(section.start)} to ${formatClock(section.end)}: ${parts.join(', ')}${open ? '' : ' folded'}`;
		fold.elem.classList.toggle('open', open);
	}
}

// The categories we fold in a rest, a search or a corpse run: the regen, the rotation's
// waiting and the debug lines. Casts, auras, damage and the session's own lines stay.
const FOLDED_CATEGORIES: Array<EventCategory> = ['resources', 'other', 'debug'];

const signed = (amount: number): string => `${amount > 0 ? '+' : amount < 0 ? '-' : ''}${formatToNumber(Math.abs(amount), { maximumFractionDigits: 0 })}`;

// What a section was, after its title and time.
const sectionSummary = (section: SessionSection): string => {
	const length = formatSpan(section.end - section.start);
	switch (section.kind) {
		case 'pull': {
			const pull = section.pull;
			if (!pull) return length;
			const levels = pull.mobs.map(mob => mob.level).join(', ');
			const mobs = `${pull.mobs.length} ${pull.mobs.length == 1 ? 'mob' : 'mobs'}${levels ? ` (${levels})` : ''}`;
			return [mobs, length, pull.died ? 'died' : ''].filter(Boolean).join(', ');
		}
		case 'rest': {
			const did = [section.ate && section.drank ? 'ate and drank' : section.ate ? 'ate' : section.drank ? 'drank' : '', section.bandaged ? 'bandaged' : '']
				.filter(Boolean)
				.join(', ');
			return [length, did].filter(Boolean).join(', ');
		}
		case 'search':
			return length;
		case 'death':
			return `corpse run ${length}`;
	}
};

// A leveling session's line in plain words, like "Found Mob 11 (level 31), 23 yards away", or
// null when the line isn't one.
export const levelingLine = (log: SimLog): string | null => {
	const at = log.raw.indexOf('] Leveling: ');
	if (at < 0) return null;
	return levelingText(log.raw.substring(at + '] Leveling: '.length).trim());
};

const levelingText = (message: string): string => {
	let m: RegExpMatchArray | null;
	if ((m = message.match(/^found Target (\d+) \(level (\d+)\) (\d+) yards away/))) {
		return `Found Mob ${m[1]} (level ${m[2]}), ${m[3]} yards away`;
	}
	if ((m = message.match(/^pulled Target (\d+) \(level (\d+)\) from (\d+) yards, (?:here in (\S+)|(chasing us))/))) {
		const how = m[5] ? 'chasing us' : `here in ${goDuration(m[4]).toFixed(1)} s`;
		return `Pulled Mob ${m[1]} (level ${m[2]}) from ${m[3]} yards, ${how}`;
	}
	if ((m = message.match(/^looking for another mob, running: (\w+)/))) {
		return `Looking for another mob, ${m[1] == 'true' ? 'running with the mobs on us' : 'standing still'}`;
	}
	if ((m = message.match(/^sitting down \(eat: (\w+), drink: (\w+)\) (.*)/))) {
		const eat = m[1] == 'true';
		const drink = m[2] == 'true';
		return `Sitting down to ${eat && drink ? 'eat and drink' : eat ? 'eat' : drink ? 'drink' : 'rest'} ${m[3]}`;
	}
	if ((m = message.match(/^died, back in (\S+)/))) {
		return `Died, back in ${formatClock(goDuration(m[1]))}`;
	}
	if ((m = message.match(/^we're at \((.*?)\), moving \((-?[\d.]+), (-?[\d.]+)\)/))) {
		return `We're at (${m[1]}), ${movement(+m[2], +m[3])}`;
	}
	if ((m = message.match(/^Target (\d+) at \((.*?)\), moving \((-?[\d.]+), (-?[\d.]+)\)/))) {
		return `Mob ${m[1]} at (${m[2]}), ${movement(+m[3], +m[4])}`;
	}
	const plain = message.replace(/Target (\d+)/g, 'Mob $1');
	return plain.charAt(0).toUpperCase() + plain.substring(1);
};

const movement = (vx: number, vy: number): string => {
	const speed = Math.hypot(vx, vy);
	if (speed == 0) return 'standing';
	const heading = (Math.atan2(vy, vx) * 180) / Math.PI;
	return `moving ${speed.toFixed(1)} yards a sec, bearing ${((heading + 360) % 360).toFixed(0)}°`;
};

// Seconds in a Go duration, like 1.97s, 737.1ms or 3m0s.
const goDuration = (text: string): number => {
	let seconds = 0;
	for (const [, value, unit] of text.matchAll(/(-?[\d.]+)(h|ms|m|s|µs|us|ns)/g)) {
		const scale = { h: 3600, m: 60, s: 1, ms: 1e-3, µs: 1e-6, us: 1e-6, ns: 1e-9 }[unit as 'h' | 'm' | 's' | 'ms' | 'µs' | 'us' | 'ns'];
		seconds += parseFloat(value) * scale;
	}
	return seconds;
};

const LOG_ROW_HEIGHT = 30;

// Times read like a clock, minutes then seconds with milliseconds: 1:05.250. Events before the
// pull get a minus sign. A leveling session's log gives hundredths, 1:05.25, because its times
// only have those.
export const formatEventTime = (time: number, digits = 3): string => {
	const positive = Math.abs(time);
	const minutes = Math.floor(positive / 60);
	const seconds = positive - minutes * 60;
	return `${time < 0 ? '-' : ''}${minutes}:${seconds.toFixed(digits).padStart(digits + 3, '0')}`;
};

// A row of the virtual scroll: an element, or a function that makes it the first time it shows.
type RowItem = Element | (() => Element);

class CustomVirtualScroll {
	private scrollContainer: HTMLElement;
	private contentContainer: HTMLElement;
	private items: Array<RowItem>;
	private itemHeight: number;
	private visibleItemsCount: number;
	private startIndex: number;
	private placeholderTop: HTMLElement;
	private placeholderBottom: HTMLElement;

	constructor({ scrollContainer, contentContainer, itemHeight }: { scrollContainer: HTMLElement; contentContainer: HTMLElement; itemHeight: number }) {
		this.scrollContainer = scrollContainer;
		this.contentContainer = contentContainer;
		this.items = [];
		this.itemHeight = itemHeight;
		this.visibleItemsCount = 50; // +1 for buffer
		this.startIndex = 0;

		this.placeholderTop = document.createElement('tr');
		this.placeholderBottom = document.createElement('tr');
		contentContainer.prepend(this.placeholderTop);
		contentContainer.append(this.placeholderBottom);

		this.attachScrollListener();
	}

	scrollToTop(): void {
		this.scrollContainer.scrollTop = 0;
		this.startIndex = 0; // Reset startIndex to ensure items are updated correctly
		this.updateVisibleItems(); // Update the visible items after scrolling to top
	}

	setItems(newItems: Array<RowItem>): void {
		this.items = newItems;
		this.scrollToTop();
	}

	private attachScrollListener(): void {
		this.scrollContainer.addEventListener('scroll', () => {
			const newIndex = Math.floor(this.scrollContainer.scrollTop / this.itemHeight);
			if (newIndex !== this.startIndex) {
				this.startIndex = newIndex;
				this.updateVisibleItems();
			}
		});
	}

	private updateVisibleItems(): void {
		const endIndex = this.startIndex + this.visibleItemsCount;
		const visibleItems = this.items.slice(this.startIndex, endIndex).map(item => (typeof item == 'function' ? item() : item));
		const remainingItems = Math.max(0, this.items.length - endIndex);

		// Update the height of the placeholders before it's placed in the dom to prevent rerender
		this.placeholderTop.style.height = `${this.startIndex * this.itemHeight}px`;
		this.placeholderBottom.style.height = `${remainingItems * this.itemHeight}px`;
		this.contentContainer.replaceChildren(
			<>
				{this.placeholderTop}
				{visibleItems.map(item => item)}
				{this.placeholderBottom}
			</>,
		);
	}
}
