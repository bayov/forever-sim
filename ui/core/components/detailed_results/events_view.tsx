// @ts-expect-error
import debounce from 'lodash/debounce';
import tippy from 'tippy.js';
import { ref } from 'tsx-vanilla';

import { IterationLog, RaidSimResult } from '../../proto/api.js';
import { SimResult } from '../../proto_utils/sim_result.js';
import { SimUI } from '../../sim_ui.js';
import { TypedEvent } from '../../typed_event.js';
import { formatToNumber } from '../../utils.js';
import { eventCategories, EventCategory, EventFilter, levelingOnlyCategories } from './events_filter.js';
import { formatClock, getLevelingSession, isLevelingResult } from './leveling_session.js';
import { LogRunner } from './log_runner.js';
import { PullsTable } from './pulls_table.js';
import { Replay } from './replay.js';
import { ResultComponent, ResultComponentConfig, SimResultData } from './result_component.js';
import { SessionStrip } from './session_strip.js';
import { levelingZoomPresets, Timeline, zoomPresets, ZoomSpan } from './timeline.js';

// Pulls is a leveling session's table of pulls, and only shows in a leveling run.
type EventsFormat = 'timeline' | 'chart' | 'log' | 'pulls' | 'replay';
// 'new' is the iteration we ran with Sim 1 iteration, after the run. 'died' is a leveling
// run's session with the most deaths.
export type IterationKind = 'typical' | 'lucky' | 'unlucky' | 'died' | 'new';

const formats: Array<{ format: EventsFormat; label: string; icon: string }> = [
	{ format: 'timeline', label: 'Timeline', icon: 'fa-bars-staggered' },
	{ format: 'chart', label: 'Chart', icon: 'fa-chart-line' },
	{ format: 'log', label: 'Log', icon: 'fa-list' },
	{ format: 'pulls', label: 'Pulls', icon: 'fa-table-list' },
	{ format: 'replay', label: 'Replay', icon: 'fa-circle-play' },
];

const iterationKinds: Array<{ kind: IterationKind; label: string; tooltip: string }> = [
	{ kind: 'typical', label: 'Typical', tooltip: 'The iteration with the DPS closest to the average.' },
	{ kind: 'lucky', label: 'Lucky', tooltip: 'The iteration with the highest DPS.' },
	{ kind: 'unlucky', label: 'Unlucky', tooltip: 'The iteration with the lowest DPS.' },
	{ kind: 'died', label: 'Died', tooltip: 'The session with the most deaths.' },
	{ kind: 'new', label: 'New', tooltip: 'Sim one more iteration with a fresh seed and the current settings.' },
];

// The filters that mean something in each view. The others are greyed out there.
const appliesTo: Record<EventsFormat, { search: boolean; categories: Array<EventCategory>; hidden: boolean; alwaysUp: boolean }> = {
	timeline: { search: true, categories: ['casts', 'auras', 'resources'], hidden: true, alwaysUp: true },
	chart: { search: false, categories: [], hidden: false, alwaysUp: false },
	log: { search: true, categories: ['casts', 'damage', 'auras', 'resources', 'session', 'other', 'debug'], hidden: true, alwaysUp: false },
	pulls: { search: false, categories: [], hidden: false, alwaysUp: false },
	replay: { search: false, categories: [], hidden: false, alwaysUp: false },
};

// In a leveling run the iterations are sessions, and they go by mobs killed.
const levelingTooltips: Partial<Record<IterationKind, string>> = {
	typical: 'The session with the mobs killed closest to the average.',
	lucky: 'The session with the most mobs killed.',
	unlucky: 'The session with the fewest mobs killed.',
	died: 'The session with the most deaths, and of those the one with the mobs killed closest to the average.',
};

const FORMAT_STORAGE_KEY = '__detailed_results_events_format__';

export interface EventsViewConfig extends ResultComponentConfig {
	simUI: SimUI | null;
	// What the iteration buttons say on hover, when the results picked their iterations some
	// other way. The DPS Report shows the fights at the 10th, 50th and 90th percentile.
	iterationTooltips?: Partial<Record<IterationKind, string>>;
	// Called when we pick another iteration, like the DPS Report's aura bars that follow it.
	onIterationChange?: (kind: IterationKind) => void;
}

// One iteration of the run, event by event. We can look at it as a timeline, a chart or a log,
// or play it back like a video of the fight, and pick the typical, the lucky or the unlucky
// iteration. The filter above is shared by the timeline, the chart and the log.
export class EventsView extends ResultComponent {
	private readonly filter = new EventFilter();
	private readonly iterationEmitter = new TypedEvent<SimResultData | null>('EventsIteration');
	private readonly timeline: Timeline;
	private readonly logRunner: LogRunner;
	private readonly replay: Replay;
	private readonly pullsTable: PullsTable;
	private readonly strip: SessionStrip;
	private readonly iterationTooltips?: Partial<Record<IterationKind, string>>;
	// Whether the latest run was in the leveling mode.
	private leveling = false;

	private format: EventsFormat = 'timeline';
	private iteration: IterationKind = 'typical';
	// The SimResult of each iteration of the latest run. The typical one is the run's own result.
	private iterationResults = new Map<IterationKind, Promise<SimResult>>();
	private latestData: SimResultData | null = null;
	// The iteration from Sim 1 iteration, until the next run.
	private newIteration: SimResult | null = null;
	private selectToken = 0;
	private simulatingNew = false;
	private readonly simUI: SimUI | null;
	private readonly onIterationChange?: (kind: IterationKind) => void;

	private readonly ui: {
		formatButtons: Map<EventsFormat, HTMLButtonElement>;
		iterationGroup: HTMLElement;
		iterationButtons: Map<IterationKind, HTMLButtonElement>;
		rerollButton: HTMLButtonElement;
		zoomGroup: HTMLElement;
		// The fight's presets and a leveling session's. We show the ones that fit the run.
		zoomButtons: Array<{ span: ZoomSpan; leveling: boolean; button: HTMLButtonElement }>;
		search: HTMLInputElement;
		searchGroup: HTMLElement;
		from: HTMLInputElement;
		to: HTMLInputElement;
		categoryButtons: Map<EventCategory, HTMLButtonElement>;
		alwaysUpButton: HTMLButtonElement;
		count: HTMLElement;
		reset: HTMLButtonElement;
		hiddenRow: HTMLElement;
		hiddenList: HTMLElement;
		timelineHost: HTMLElement;
		logHost: HTMLElement;
		pullsHost: HTMLElement;
		replayHost: HTMLElement;
		filterRow: HTMLElement;
		loading: HTMLElement;
		stripHost: HTMLElement;
	};

	constructor(config: EventsViewConfig) {
		config.rootCssClass = 'events-view-root';
		super(config);

		try {
			const stored = window.localStorage.getItem(FORMAT_STORAGE_KEY) as EventsFormat | null;
			if (stored && formats.some(f => f.format == stored)) this.format = stored;
		} catch {
			// The format is a convenience, the view works without it.
		}

		const refs = {
			iterationGroup: ref<HTMLDivElement>(),
			reroll: ref<HTMLButtonElement>(),
			zoomGroup: ref<HTMLDivElement>(),
			search: ref<HTMLInputElement>(),
			searchGroup: ref<HTMLDivElement>(),
			from: ref<HTMLInputElement>(),
			to: ref<HTMLInputElement>(),
			count: ref<HTMLSpanElement>(),
			reset: ref<HTMLButtonElement>(),
			alwaysUp: ref<HTMLButtonElement>(),
			hiddenRow: ref<HTMLDivElement>(),
			hiddenList: ref<HTMLDivElement>(),
			timelineHost: ref<HTMLDivElement>(),
			logHost: ref<HTMLDivElement>(),
			pullsHost: ref<HTMLDivElement>(),
			replayHost: ref<HTMLDivElement>(),
			filterRow: ref<HTMLDivElement>(),
			loading: ref<HTMLDivElement>(),
			stripHost: ref<HTMLDivElement>(),
		};
		const formatButtons = new Map<EventsFormat, HTMLButtonElement>();
		const iterationButtons = new Map<IterationKind, HTMLButtonElement>();
		const zoomButtons: Array<{ span: ZoomSpan; leveling: boolean; button: HTMLButtonElement }> = [];
		const categoryButtons = new Map<EventCategory, HTMLButtonElement>();

		this.rootElem.appendChild(
			<>
				<div className="content-block events-controls">
					<div className="events-controls-row">
						<div className="segmented events-format" attributes={{ role: 'group', 'aria-label': 'View' }}>
							{formats.map(({ format, label, icon }) => {
								const button = (
									<button className="segmented-option" type="button" dataset={{ format }}>
										<i className={`fas ${icon}`} />
										{label}
									</button>
								) as HTMLButtonElement;
								button.addEventListener('click', () => this.setFormat(format));
								formatButtons.set(format, button);
								return button;
							})}
						</div>
						<div ref={refs.iterationGroup} className="events-iteration-group">
							<span className="events-control-label">Iteration</span>
							<div className="segmented events-iteration" attributes={{ role: 'group', 'aria-label': 'Iteration' }}>
								{iterationKinds.map(({ kind, label, tooltip: defaultTooltip }) => {
									const tooltip = config.iterationTooltips?.[kind] ?? defaultTooltip;
									const button = (
										<button className={`segmented-option iteration-${kind}`} type="button">
											{kind == 'new' && <i className="fas fa-dice iteration-new-icon" />}
											<span className="iteration-label">{label}</span>
											{kind == 'died' && <i className="fas fa-skull iteration-died-icon" />}
											<span className="iteration-dps"></span>
										</button>
									) as HTMLButtonElement;
									button.addEventListener('click', () => (kind == 'new' && !this.newIteration ? this.simNewIteration() : this.setIteration(kind)));
									tippy(button, { content: tooltip });
									iterationButtons.set(kind, button);
									return button;
								})}
								<button
									ref={refs.reroll}
									className="segmented-option iteration-reroll hide"
									type="button"
									attributes={{ 'aria-label': 'Sim another new iteration' }}>
									<i className="fas fa-rotate-right" />
								</button>
							</div>
						</div>
						<div className="events-controls-end">
							<div ref={refs.zoomGroup} className="events-zoom">
								<span className="events-hint">Drag or Shift + scroll to move, Alt + scroll to zoom</span>
								<div className="segmented" attributes={{ role: 'group', 'aria-label': 'Zoom' }}>
									{[...zoomPresets.map(p => ({ ...p, leveling: false })), ...levelingZoomPresets.map(p => ({ ...p, leveling: true }))].map(({ span, label, leveling }) => {
										const button = (
											<button className="segmented-option" type="button">
												{label}
											</button>
										) as HTMLButtonElement;
										button.addEventListener('click', () => this.timeline.setZoom(span));
										zoomButtons.push({ span, leveling, button });
										return button;
									})}
								</div>
							</div>
						</div>
					</div>
					<div ref={refs.stripHost} className="events-strip-row"></div>
					<div ref={refs.filterRow} className="events-filter-row">
						<div ref={refs.searchGroup} className="events-search">
							<i className="fas fa-magnifying-glass" />
							<input ref={refs.search} type="search" className="form-control" placeholder="Search abilities and events" />
						</div>
						<div className="events-range">
							<span className="events-control-label">From</span>
							<input ref={refs.from} type="text" className="form-control" attributes={{ inputmode: 'decimal' }} placeholder="start" />
							<span className="events-control-label">to</span>
							<input ref={refs.to} type="text" className="form-control" attributes={{ inputmode: 'decimal' }} placeholder="end" />
						</div>
						<div className="events-categories">
							{eventCategories.map(({ category, label }) => {
								const button = (
									<button className="filter-chip" type="button">
										{label}
									</button>
								) as HTMLButtonElement;
								button.addEventListener('click', () => this.filter.toggleCategory(category));
								categoryButtons.set(category, button);
								return button;
							})}
							<button ref={refs.alwaysUp} className="filter-chip events-always-up" type="button">
								Hide 100% auras
							</button>
						</div>
						<span ref={refs.count} className="events-count"></span>
						<button ref={refs.reset} className="btn btn-link events-reset" type="button">
							<i className="fas fa-rotate-left me-1" />
							Reset filters
						</button>
					</div>
					<div ref={refs.hiddenRow} className="events-hidden-row hide">
						<span className="events-control-label">Hidden</span>
						<div ref={refs.hiddenList} className="events-hidden-list"></div>
					</div>
				</div>
				<div className="content-block events-body">
					<div ref={refs.loading} className="events-loading hide">
						<div className="spinner-border spinner-border-sm" />
						Loading the iteration
					</div>
					<div ref={refs.timelineHost} className="events-timeline"></div>
					<div ref={refs.logHost} className="events-log"></div>
					<div ref={refs.pullsHost} className="events-pulls"></div>
					<div ref={refs.replayHost} className="events-replay"></div>
				</div>
			</>,
		);

		this.ui = {
			formatButtons,
			iterationGroup: refs.iterationGroup.value!,
			iterationButtons,
			rerollButton: refs.reroll.value!,
			zoomGroup: refs.zoomGroup.value!,
			zoomButtons,
			search: refs.search.value!,
			searchGroup: refs.searchGroup.value!,
			from: refs.from.value!,
			to: refs.to.value!,
			categoryButtons,
			alwaysUpButton: refs.alwaysUp.value!,
			count: refs.count.value!,
			reset: refs.reset.value!,
			hiddenRow: refs.hiddenRow.value!,
			hiddenList: refs.hiddenList.value!,
			timelineHost: refs.timelineHost.value!,
			logHost: refs.logHost.value!,
			pullsHost: refs.pullsHost.value!,
			replayHost: refs.replayHost.value!,
			filterRow: refs.filterRow.value!,
			loading: refs.loading.value!,
			stripHost: refs.stripHost.value!,
		};
		this.iterationTooltips = config.iterationTooltips;

		this.strip = new SessionStrip({ parent: this.ui.stripHost, filter: this.filter });

		this.timeline = new Timeline({
			parent: this.ui.timelineHost,
			resultsEmitter: this.iterationEmitter,
			filter: this.filter,
		});
		this.logRunner = new LogRunner({
			parent: this.ui.logHost,
			resultsEmitter: this.iterationEmitter,
			filter: this.filter,
			onTimeClick: time => this.showInTimeline(time),
		});
		this.pullsTable = new PullsTable({
			parent: this.ui.pullsHost,
			resultsEmitter: this.iterationEmitter,
			filter: this.filter,
			onPick: (start, end) => this.filter.setRange(start, end),
		});
		this.replay = new Replay({
			parent: this.ui.replayHost,
			resultsEmitter: this.iterationEmitter,
			filter: this.filter,
		});

		this.ui.search.addEventListener(
			'input',
			debounce(() => this.filter.setSearch(this.ui.search.value), 150),
		);
		const onRangeChange = () => {
			const from = parseSeconds(this.ui.from.value);
			const to = parseSeconds(this.ui.to.value);
			this.filter.setRange(from, to != null && from != null && to <= from ? null : to);
		};
		this.ui.from.addEventListener('change', onRangeChange);
		this.ui.to.addEventListener('change', onRangeChange);
		[this.ui.from, this.ui.to].forEach(input => input.addEventListener('keydown', event => event.key == 'Enter' && onRangeChange()));
		tippy(this.ui.from, { content: 'Seconds, or minutes and seconds like 4:05.' });
		tippy(this.ui.to, { content: 'Seconds, or minutes and seconds like 4:05.' });
		this.ui.reset.addEventListener('click', () => this.filter.reset());
		this.ui.alwaysUpButton.addEventListener('click', () => this.filter.toggleHideAlwaysUp());
		tippy(this.ui.alwaysUpButton, { content: 'Hide the buffs and debuffs that are up for the whole fight, like Blessing of Kings.' });
		this.timeline.zoomEmitter.on(() => this.updateZoomButtons());
		// The strip shows where the replay is, and which part of the session the timeline or the
		// chart has in view.
		this.replay.timeEmitter.on((_, time) => this.strip.setCursor(this.shownFormat() == 'replay' ? time : null));
		this.timeline.viewEmitter.on(() => this.updateStripView());

		this.simUI = config.simUI;
		this.onIterationChange = config.onIterationChange;
		this.ui.rerollButton.addEventListener('click', () => this.simNewIteration());
		tippy(this.ui.rerollButton, { content: 'Sim another new iteration, with a fresh seed and the current settings.' });

		this.filter.changeEmitter.on(() => this.updateFilterControls());
		this.logRunner.countEmitter.on(() => this.updateCount());

		// The timeline and the replay draw only while we can see them, so we let them know when a
		// tab opens.
		const onTabShown = () => {
			this.timeline.renderIfDirty();
			this.replay.setShown(this.shownFormat() == 'replay');
		};
		document.addEventListener('shown.bs.tab', onTabShown);
		this.addOnDisposeCallback(() => document.removeEventListener('shown.bs.tab', onTabShown));

		this.applyFormat();
		this.updateFilterControls();
	}

	onSimResult(resultData: SimResultData) {
		// A new run starts over from its typical iteration. The player and target filter of the
		// results gives us new data too, and then we keep the iteration we look at.
		const isNewRun = this.latestData?.result !== resultData.result;
		this.latestData = resultData;
		if (isNewRun) {
			this.iterationResults = new Map([['typical', Promise.resolve(resultData.result)]]);
			this.iteration = 'typical';
			this.newIteration = null;
			// A new session has other pulls, so the part we picked in the last one means nothing.
			const leveling = isLevelingResult(resultData.result);
			if (leveling || this.leveling) this.filter.setRange(null, null);
			this.leveling = leveling;
			this.rootElem.classList.toggle('leveling', leveling);
			this.applyFormat();
		}
		this.updateIterationButtons();
		this.selectIteration();
	}

	private setFormat(format: EventsFormat) {
		this.format = format;
		try {
			window.localStorage.setItem(FORMAT_STORAGE_KEY, format);
		} catch {
			// See the constructor.
		}
		this.applyFormat();
	}

	// The format we show: the one we picked, but the timeline when we picked the pulls and the
	// run isn't a leveling one.
	private shownFormat(): EventsFormat {
		return this.format == 'pulls' && !this.leveling ? 'timeline' : this.format;
	}

	private applyFormat() {
		// A fight has no pulls. We keep the choice, so the next leveling run opens on them again.
		const format = this.shownFormat();
		this.ui.formatButtons.forEach((button, f) => {
			button.classList.toggle('active', f == format);
			button.setAttribute('aria-pressed', String(f == format));
			button.classList.toggle('hide', f == 'pulls' && !this.leveling);
		});
		this.ui.timelineHost.classList.toggle('hide', format != 'timeline' && format != 'chart');
		this.ui.logHost.classList.toggle('hide', format != 'log');
		this.ui.pullsHost.classList.toggle('hide', format != 'pulls');
		this.ui.replayHost.classList.toggle('hide', this.shownFormat() != 'replay');
		// The filters don't change what the replay shows, so we hide them there.
		this.ui.filterRow.classList.toggle('hide', this.shownFormat() == 'replay');
		this.replay.setShown(this.shownFormat() == 'replay');
		this.ui.zoomGroup.classList.toggle('hide', this.shownFormat() != 'timeline');
		this.updateStripView();
		if (this.shownFormat() == 'timeline' || this.shownFormat() == 'chart') {
			this.timeline.setMode(this.shownFormat() == 'timeline' ? 'rotation' : 'chart');
			this.timeline.renderIfDirty();
		}
		this.updateFilterControls();
	}

	// Sims one more iteration and shows it as New. A second one replaces the first.
	private async simNewIteration() {
		if (!this.simUI || this.simulatingNew) return;
		this.simulatingNew = true;
		this.updateIterationButtons();
		this.ui.loading.classList.remove('hide');
		const result = await this.simUI.runSimOnce();
		this.simulatingNew = false;
		this.ui.loading.classList.add('hide');
		if (result) {
			this.showNewIteration(result);
		} else {
			this.updateIterationButtons();
		}
	}

	private showNewIteration(result: SimResult) {
		this.newIteration = result;
		this.iterationResults.set('new', Promise.resolve(result));
		this.iteration = 'new';
		this.updateIterationButtons();
		this.selectIteration();
	}

	setIteration(kind: IterationKind) {
		if (this.iteration == kind) return;
		this.iteration = kind;
		this.updateIterationButtons();
		this.selectIteration();
		this.onIterationChange?.(kind);
	}

	// Sends the chosen iteration to the views. We parse the lucky and unlucky logs the first time
	// we pick them, because most of the time we only look at the typical one.
	private async selectIteration() {
		const data = this.latestData;
		if (!data) return;
		const kind = this.iteration;
		const token = ++this.selectToken;

		let resultPromise = this.iterationResults.get(kind);
		if (!resultPromise) {
			const iterationLog = getIterationLog(data.result.result, kind);
			if (!iterationLog?.logs) return;
			resultPromise = SimResult.makeNew(
				data.result.request,
				RaidSimResult.create({
					...data.result.result,
					logs: iterationLog.logs,
					firstIterationDuration: iterationLog.duration,
				}),
			);
			this.iterationResults.set(kind, resultPromise);
		}

		this.ui.loading.classList.remove('hide');
		const result = await resultPromise;
		if (token != this.selectToken) return;
		this.ui.loading.classList.add('hide');
		this.strip.setSession(getLevelingSession(result), data.result.result.encounterMetrics?.leveling?.mobsKilled);
		this.iterationEmitter.emit(TypedEvent.nextEventID(), { eventID: data.eventID, result, filter: data.filter });
	}

	private updateIterationButtons() {
		const result = this.latestData?.result.result;
		// A run of one iteration has no lucky or unlucky one, and older results don't have the
		// iterations at all. Then we only offer the run's own iteration and the new one.
		const hasPicked = !!result?.typicalIteration && (result?.iterationsDone ?? 0) > 1;
		const leveling = this.leveling;
		this.ui.iterationGroup.classList.toggle('hide', !hasPicked && !this.simUI);
		if (!result) return;

		// Before we sim a new iteration, New is the button that sims it. After, it picks it, and
		// the small button next to it sims another.
		this.ui.iterationButtons.forEach((button, kind) => {
			const iterationLog = getIterationLog(result, kind);
			const shown = kind == 'typical' || (kind == 'new' ? !!this.simUI : kind == 'died' ? hasPicked && !!result.diedIteration : hasPicked);
			button.classList.toggle('hide', !shown);
			button.disabled = kind == 'new' ? this.simulatingNew : kind != 'typical' && !iterationLog?.logs;
			if (kind == 'new') {
				button.classList.toggle('not-simmed', !this.newIteration);
				button.querySelector('.iteration-new-icon')!.classList.toggle('fa-spin', this.simulatingNew);
				(button as any)._tippy?.setContent(
					this.newIteration
						? 'The iteration we simmed with a fresh seed and the current settings.'
						: 'Sim one more iteration with a fresh seed and the current settings.',
				);
			}
			button.classList.toggle('active', kind == this.iteration);
			button.setAttribute('aria-pressed', String(kind == this.iteration));
			if (kind != 'new') {
				(button as any)._tippy?.setContent((leveling && levelingTooltips[kind]) || this.iterationTooltips?.[kind] || iterationKinds.find(k => k.kind == kind)!.tooltip);
			}
			// A leveling session shows its mobs killed, and the died one its deaths too.
			let value: string;
			if (leveling) {
				const kills = kind == 'new' ? this.newIteration?.result.encounterMetrics?.leveling?.mobsKilled : iterationLog?.kills;
				value = kills !== undefined ? `${formatToNumber(kills, { maximumFractionDigits: 1 })} mobs` : '';
				if (kind == 'died' && iterationLog) value += ` · ${iterationLog.deaths}`;
			} else {
				const dps = kind == 'new' ? this.newIteration?.raidMetrics.dps.avg : iterationLog?.dps;
				value = dps !== undefined ? formatToNumber(dps, { maximumFractionDigits: 1 }) : '';
			}
			button.querySelector('.iteration-dps')!.textContent = value;
		});
		this.ui.rerollButton.classList.toggle('hide', !this.newIteration);
		this.ui.rerollButton.disabled = this.simulatingNew;
		this.ui.rerollButton.querySelector('i')!.classList.toggle('fa-spin', this.simulatingNew);
	}

	private showInTimeline(time: number) {
		if (!this.filter.inRange(time)) {
			this.filter.setRange(null, null);
		}
		this.setFormat('timeline');
		this.timeline.scrollToTime(time);
	}

	private updateStripView() {
		if (this.shownFormat() != 'replay') this.strip.setCursor(null);
		this.strip.setView(this.shownFormat() == 'timeline' || this.shownFormat() == 'chart' ? this.timeline.getViewRange() : null);
	}

	// Alt + scroll zooms between the presets, and then none of them is active.
	private updateZoomButtons() {
		const zoom = this.timeline.getZoom();
		this.ui.zoomButtons.forEach(({ span, leveling, button }) => {
			button.classList.toggle('hide', leveling != this.leveling);
			button.classList.toggle('active', span == zoom);
			button.setAttribute('aria-pressed', String(span == zoom));
		});
	}

	// Brings the controls in line with the filter, and greys out what the view doesn't use.
	private updateFilterControls() {
		const applies = appliesTo[this.shownFormat()];

		if (this.ui.search.value != this.filter.search && document.activeElement != this.ui.search) {
			this.ui.search.value = this.filter.search;
		}
		this.ui.search.disabled = !applies.search;
		this.ui.searchGroup.classList.toggle('disabled', !applies.search);
		const formatTime = (t: number | null) => (t == null ? '' : this.leveling ? formatClock(t, true) : String(t));
		if (document.activeElement != this.ui.from) this.ui.from.value = formatTime(this.filter.from);
		if (document.activeElement != this.ui.to) this.ui.to.value = formatTime(this.filter.to);

		this.ui.categoryButtons.forEach((button, category) => {
			const active = this.filter.categories.has(category);
			button.classList.toggle('active', active);
			button.setAttribute('aria-pressed', String(active));
			button.disabled = !applies.categories.includes(category);
			button.classList.toggle('hide', levelingOnlyCategories.includes(category) && !this.leveling);
		});
		this.ui.alwaysUpButton.classList.toggle('active', this.filter.hideAlwaysUp);
		this.ui.alwaysUpButton.setAttribute('aria-pressed', String(this.filter.hideAlwaysUp));
		this.ui.alwaysUpButton.disabled = !applies.alwaysUp || !this.filter.categories.has('auras');

		this.ui.reset.classList.toggle('invisible', !this.filter.isActive());

		this.ui.hiddenRow.classList.toggle('hide', this.filter.hidden.length == 0 || this.shownFormat() == 'replay');
		this.ui.hiddenRow.classList.toggle('disabled', !applies.hidden);
		this.ui.hiddenList.replaceChildren(
			...this.filter.hidden.map(actionId => {
				const iconRef = ref<HTMLSpanElement>();
				const chip = (
					<button className="filter-chip hidden-action" type="button">
						<span ref={iconRef} className="hidden-action-icon" />
						{actionId.name}
						<i className="fas fa-xmark" />
					</button>
				) as HTMLButtonElement;
				actionId.setBackground(iconRef.value! as HTMLAnchorElement);
				chip.addEventListener('click', () => this.filter.unhide(actionId));
				return chip;
			}),
			...(this.filter.hidden.length > 1
				? [
						(() => {
							const showAll = (
								<button className="btn btn-link events-show-all" type="button">
									Show all
								</button>
							) as HTMLButtonElement;
							showAll.addEventListener('click', () => this.filter.unhideAll());
							return showAll;
						})(),
				  ]
				: []),
		);

		this.updateCount();
		this.updateZoomButtons();
	}

	private updateCount() {
		const { shownCount, totalCount } = this.logRunner;
		const show = this.shownFormat() == 'log' && totalCount > 0;
		this.ui.count.classList.toggle('hide', !show);
		this.ui.count.textContent =
			shownCount == totalCount ? `${formatToNumber(totalCount)} events` : `${formatToNumber(shownCount)} of ${formatToNumber(totalCount)} events`;
	}
}

const getIterationLog = (result: RaidSimResult, kind: IterationKind): IterationLog | undefined => {
	switch (kind) {
		case 'typical':
			return result.typicalIteration;
		case 'lucky':
			return result.luckyIteration;
		case 'unlucky':
			return result.unluckyIteration;
		case 'died':
			return result.diedIteration;
		case 'new':
			return undefined;
	}
};

// Times before the pull are negative, like -3 for a prepull action 3 seconds before it. A time
// can also be minutes and seconds, like 4:05 or 4:05.5, for a 20 min leveling session.
const parseSeconds = (value: string): number | null => {
	const text = value.trim();
	if (text == '') return null;
	const clock = text.match(/^(-?)(\d+):(\d+(?:\.\d*)?)$/);
	if (clock) return (clock[1] ? -1 : 1) * (parseInt(clock[2]) * 60 + parseFloat(clock[3]));
	const seconds = parseFloat(text);
	return isNaN(seconds) ? null : seconds;
};
