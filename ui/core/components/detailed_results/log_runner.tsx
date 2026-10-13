import { delegate } from 'tippy.js';
import { ref } from 'tsx-vanilla';

import { SimLog } from '../../proto_utils/logs_parser';
import { SimResult } from '../../proto_utils/sim_result.js';
import { TypedEvent } from '../../typed_event.js';
import { EventFilter } from './events_filter.js';
import { badgeOutcomeInLogLine } from './hit_outcome.js';
import { ResultComponent, ResultComponentConfig, SimResultData } from './result_component.js';

export interface LogRunnerConfig extends ResultComponentConfig {
	filter: EventFilter;
	// Called when we click the time of an event, to show that moment in the timeline.
	onTimeClick: (time: number) => void;
}

interface LogRow {
	log: SimLog;
	elem: HTMLTableRowElement;
	text: string;
}

// The iteration's events one per line, oldest first.
export class LogRunner extends ResultComponent {
	private readonly filter: EventFilter;
	private readonly onTimeClick: (time: number) => void;
	private readonly virtualScroll: CustomVirtualScroll;
	private readonly scrollContainer: HTMLDivElement;
	private readonly emptyElem: HTMLElement;

	private rows: Array<LogRow> = [];
	private rowsKey: SimResult | null = null;
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
			// Cast completed lines repeat the cast began ones, so we leave them out.
			this.rows = resultData.result.logs
				.filter(log => log.raw.trim() != '')
				.filter(log => !log.isCastCompleted())
				.map(log => this.makeRow(log));
		}
		this.applyFilter();
	}

	private applyFilter() {
		const shownRows = this.rows.filter(row => this.filter.matchesLog(row.log, row.text));
		this.shownCount = shownRows.length;
		this.totalCount = this.rows.length;
		this.emptyElem.classList.toggle('hide', shownRows.length > 0 || this.rows.length == 0);

		const shown: Element[] = shownRows.map(row => row.elem);
		const pullIndex = shownRows.findIndex(row => row.log.timestamp >= 0);
		if (pullIndex > 0) shown.splice(pullIndex, 0, this.pullRow);
		this.virtualScroll.setItems(shown);
		this.countEmitter.emit(TypedEvent.nextEventID());
	}

	private makeRow(log: SimLog): LogRow {
		const timeRef = ref<HTMLButtonElement>();
		const eventRef = ref<HTMLTableCellElement>();
		const elem = (
			<tr>
				<td className="log-timestamp">
					<button ref={timeRef} className="log-time-link" type="button" dataset={{ tippyContent: 'Show in timeline', tippyPlacement: 'right' }}>
						{formatEventTime(log.timestamp)}
					</button>
				</td>
				<td ref={eventRef} className="log-event">
					<span className="log-event-text">{log.toHTML(false)}</span>
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

		return { log, elem, text: text.toLowerCase() };
	}
}

const LOG_ROW_HEIGHT = 30;

// Times read like a clock, minutes then seconds with milliseconds: 1:05.250. Events before the
// pull get a minus sign.
export const formatEventTime = (time: number): string => {
	const positive = Math.abs(time);
	const minutes = Math.floor(positive / 60);
	const seconds = positive - minutes * 60;
	return `${time < 0 ? '-' : ''}${minutes}:${seconds.toFixed(3).padStart(6, '0')}`;
};

class CustomVirtualScroll {
	private scrollContainer: HTMLElement;
	private contentContainer: HTMLElement;
	private items: Element[];
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

	setItems(newItems: CustomVirtualScroll['items']): void {
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
		const visibleItems = this.items.slice(this.startIndex, endIndex);
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
