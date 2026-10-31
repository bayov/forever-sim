import ApexCharts from 'apexcharts';
import tippy, { followCursor, Props } from 'tippy.js';
import { ref } from 'tsx-vanilla';

import { ResourceType, Ruleset } from '../../proto/api.js';
import { OtherAction } from '../../proto/common.js';
import { ActionId, resourceTypeToIcon } from '../../proto_utils/action_id.js';
import { AuraUptimeLog, CastLog, DamageDealtLog, DpsLog, ResourceChangedLogGroup, SimLog, ThreatLogGroup } from '../../proto_utils/logs_parser.js';
import { resourceNames } from '../../proto_utils/names.js';
import { UnitMetrics } from '../../proto_utils/sim_result.js';
import { classColors, orderedResourceTypes, specToClass } from '../../proto_utils/utils.js';
import { TypedEvent } from '../../typed_event.js';
import { bucket, distinct, formatToCompactNumber, formatToNumber, stringComparator, sum } from '../../utils.js';
import { actionColors } from './color_settings.js';
import { EventCategory, EventFilter } from './events_filter.js';
import { hitOutcome, HitOutcomeKind, outcomeBadge } from './hit_outcome.js';
import { iconColor } from './icon_color.js';
import { formatClock, getLevelingSession, LevelingSession, mobName, SESSION_STATES, valueAt } from './leveling_session.js';
import { ResultComponent, ResultComponentConfig, SimResultData } from './result_component.js';

type TooltipHandler = (dataPointIndex: number) => Element;

const dpsColor = '#ed5653';
const manaColor = '#2E93fA';
// The resource the chart draws beside the DPS, the first of these the unit has, in the game's
// colors.
const chartResources = [ResourceType.ResourceTypeMana, ResourceType.ResourceTypeRage, ResourceType.ResourceTypeEnergy];
const resourceChartColors: Partial<Record<ResourceType, string>> = {
	[ResourceType.ResourceTypeMana]: manaColor,
	[ResourceType.ResourceTypeRage]: '#ff0000',
	[ResourceType.ResourceTypeEnergy]: '#ffd700',
};
// Rage is red like the DPS line, so next to rage the DPS line takes the warrior's class color.
const rageDpsColor = '#c79c6e';
const threatColor = '#b56d07';
// The leveling chart's colors. Health and mana share the lower chart, and the pair passes the
// dataviz palette checks on the dark card. Mobs killed is the only solid line of the upper chart,
// and the average pace is a dashed gray line beside it.
const killsColor = '#c8a24a';
const paceColor = '#8e948f';
const levelingHealthColor = '#3fa85a';
const levelingManaColor = '#3d8ee0';

// How the timeline shows the iteration. The rotation has a row per ability, aura and resource.
// The chart draws DPS, the main resource and threat over time.
export type TimelineMode = 'rotation' | 'chart';

export interface TimelineConfig extends ResultComponentConfig {
	filter: EventFilter;
}

// One row of the rotation, with the label on the left and the events on the right. We keep them
// so the filter can show, hide and fade rows without drawing the rotation again.
interface RotationRow {
	label: HTMLElement;
	row: HTMLElement;
	name: string;
	actionId: ActionId | null;
	category: EventCategory;
	events: Array<RowEvent>;
	// An aura that's up from the pull to the end of the fight.
	alwaysUp?: boolean;
	// The icons in the row, at the time their cast lands or their aura starts. An aura's stacks
	// at that moment go in the icon's corner.
	icons?: Array<{ time: number; elem: HTMLElement }>;
}

// One thing drawn in a row, like a cast or an aura's uptime, with the text of the logs it came
// from. The search looks in that text.
interface RowEvent {
	elem: Element;
	text: string;
	// The ability of the event, in a row that mixes abilities like the GCD row. Hiding the
	// ability hides the event.
	actionId?: ActionId;
}

// The search looks at what happened, so we leave out the time and the source that start every
// log line, like "[12.34] [Player (#1)] ". Otherwise a search for "1" would match every event.
const logPrefixRegex = /^\[[^\]]*\]\s*(\[[^\]]*\]\s*)?/;
// Some lines name the ability only by its ID, like "from {SpellID: 10497}", so we add each
// log's ability name too.
export const searchText = (logs: Array<SimLog>): string =>
	logs
		.map(log => `${log.raw.replace(logPrefixRegex, '')} ${log.actionId?.name ?? ''}`)
		.join(' ')
		.toLowerCase();

// The sim deals a melee swing's damage 10 ms after the swing (SpellBatchWindow), and the logs
// parser reads that delay as travel time. A real projectile takes much longer, so we treat 10 ms
// or less as no travel at all.
const MAX_BATCH_DELAY = 0.0105;
const travelTimeOf = (castLog: CastLog): number => (castLog.travelTime > MAX_BATCH_DELAY ? castLog.travelTime : 0);

// A cast icon gets a small badge when a hit landed in a way worth seeing, like a crit or a parry.
// Plain hits and ticks get none, so the badges stand out. When the hits of one cast landed in
// different ways, we show the most notable one, in this order.
const notableOutcomes: Array<HitOutcomeKind> = ['crit', 'avoid', 'glance'];

// The badge is narrow, so the longer words are short.
const badgeLabels: Record<string, string> = {
	Crit: 'CRIT',
	Crush: 'CRSH',
	Glance: 'GLNC',
	Miss: 'MISS',
	Dodge: 'DODGE',
	Parry: 'PARRY',
	Block: 'BLOCK',
	'Blocked Crit': 'BLOCK',
	Resist: 'RES',
};

// The most notable outcome of a cast's hits, or null when every hit was plain. A partial resist
// on a plain hit counts as a resist.
const notableOutcome = (castLog: CastLog): { label: string; kind: HitOutcomeKind; mixed: boolean } | null => {
	const outcomes = castLog.damageDealtLogs.map(ddl => {
		const outcome = hitOutcome(ddl);
		const label = outcome.label.replace(/^(Heal|Shield) /, '');
		return outcome.resisted > 0 && outcome.kind == 'hit' ? { label: 'Resist', kind: 'avoid' as HitOutcomeKind } : { label, kind: outcome.kind };
	});
	for (const kind of notableOutcomes) {
		const found = outcomes.find(outcome => outcome.kind == kind);
		if (found) return { ...found, mixed: outcomes.some(outcome => outcome.label != found.label) };
	}
	return null;
};

// A bar can be much wider than the view, like an aura that lasts the whole fight. Its tooltip
// would sit at the middle or the start of the bar, maybe far off screen, so it follows the mouse
// along the timeline instead.
const followAlong: Partial<Props> = { followCursor: 'horizontal', plugins: [followCursor] };

// The parts of the hover tooltips. Each starts with a title (the ability's icon and name), then
// a muted line with the timing, then the hits or the changes, then key and value rows.
const formatSeconds = (time: number): string => `${time.toFixed(2)}s`;

const formatAmount = (value: number, digits = 0): string => formatToNumber(value, { minimumFractionDigits: digits, maximumFractionDigits: digits });

export const tooltipTitle = (actionId: ActionId, note?: string): JSX.Element => (
	<div className="timeline-tooltip-title">
		{actionId.iconUrl && <img className="timeline-tooltip-icon" src={actionId.iconUrl} />}
		<span className="timeline-tooltip-name">{actionId.name}</span>
		{note && <span className="timeline-tooltip-note">{note}</span>}
	</div>
);

export const tooltipRow = (key: string, value: string, valueClass = ''): JSX.Element => (
	<div className="timeline-tooltip-row">
		<span className="timeline-tooltip-key">{key}</span>
		<span className={`timeline-tooltip-value ${valueClass}`}>{value}</span>
	</div>
);

// One hit: its time, how it landed and the amount. The chart lists hits of several abilities,
// so there we name the ability too.
export const tooltipHit = (ddl: DamageDealtLog, withName = false): JSX.Element => {
	const outcome = hitOutcome(ddl);
	const landed = !ddl.miss && !ddl.dodge && !ddl.parry;
	return (
		<li>
			<span className="timeline-tooltip-time">{formatSeconds(ddl.timestamp)}</span>
			<span className="timeline-tooltip-outcome">
				{withName && ddl.actionId?.iconUrl && <img className="timeline-tooltip-icon small" src={ddl.actionId.iconUrl} />}
				{withName && ddl.actionId && <span>{ddl.actionId.name}</span>}
				{outcomeBadge(outcome.label, outcome.kind)}
				{outcome.resisted > 0 && <span className="timeline-tooltip-muted">{outcome.resisted}% resisted</span>}
			</span>
			<span className="timeline-tooltip-amount">
				{landed ? formatAmount(ddl.amount) : ''}
				{ddl.source?.isTarget && <span className="timeline-tooltip-muted threat-metrics"> {formatAmount(ddl.threat, 1)} threat</span>}
			</span>
		</li>
	);
};

// One change to a resource or to threat, with what caused it.
const tooltipChange = (log: SimLog, value: string, valueClass = ''): JSX.Element => (
	<li>
		{log.actionId?.iconUrl && <img className="timeline-tooltip-icon small" src={log.actionId.iconUrl} />}
		<span className="timeline-tooltip-outcome">{log.actionId?.name ?? ''}</span>
		<span className={`timeline-tooltip-amount ${valueClass}`}>{value}</span>
	</li>
);

// How many seconds of the rotation fit in the view, or 'fight' for the whole time range. A
// leveling session also has 'selection', which fits the From and To range the session strip
// picked, and 'session', the whole session.
export type ZoomSpan = number | 'fight' | 'selection' | 'session';

export const zoomPresets: Array<{ span: ZoomSpan; label: string }> = [
	{ span: 10, label: '10s' },
	{ span: 30, label: '30s' },
	{ span: 60, label: '1m' },
	{ span: 'fight', label: 'Fight' },
];

// A session is 20 min, so the zoom goes further out. Selection fits the pull, rest or death we
// picked on the session strip, or the whole session when we picked none.
export const levelingZoomPresets: Array<{ span: ZoomSpan; label: string }> = [
	{ span: 10, label: '10s' },
	{ span: 30, label: '30s' },
	{ span: 60, label: '1m' },
	{ span: 300, label: '5m' },
	{ span: 'selection', label: 'Selection' },
	{ span: 'session', label: 'Session' },
];

// The closest we zoom in, in pixels per second.
const maxPxPerSecond = 400;
// How fast the zoom eases to its target. After this many milliseconds it has gone about 63% of
// the way.
const zoomAnimationMs = 60;

// A zoom on its way to its target. We keep the time at anchorX pixels from the left of the view
// in place.
interface ZoomAnimation {
	anchorTime: number;
	anchorX: number;
	lastFrame: number;
}

export class Timeline extends ResultComponent {
	private readonly filter: EventFilter;

	private readonly dpsResourcesPlotElem: HTMLElement;
	private readonly chartLegend: HTMLElement;
	private dpsResourcesPlot: any;
	// In a leveling session the chart is two charts on one time axis: mobs killed on top, and
	// health and mana under it. They have different scales, so they don't share an axis.
	private readonly vitalsPlotElem: HTMLElement;
	private readonly vitalsLegend: HTMLElement;
	private vitalsPlot: any = null;
	private chartIsLeveling = false;

	private readonly rotationPlotElem: HTMLElement;
	private readonly rotationLabels: HTMLElement;
	private readonly rotationTimeline: HTMLElement;
	private readonly emptyElem: HTMLElement;

	private resultData: SimResultData | null;
	// The leveling session of the iteration, or null for a normal fight.
	private session: LevelingSession | null = null;
	private rendered: boolean;
	// Set when something changed while the timeline was hidden, so we draw it once it shows.
	private dirty = true;
	private mode: TimelineMode = 'rotation';

	private rows: Array<RotationRow | 'separator'> = [];
	private separators: Array<HTMLElement[]> = [];

	// The part of the iteration we draw, in seconds, from the filter's time range.
	private viewStart = 0;
	private viewEnd = 1;
	// The first second of the prepull, or 0 when there's no prepull.
	private prepullStart = 0;
	private zoomSpan: ZoomSpan = 10;
	// From the zoom and the width of the view. It changes on every frame while the zoom animates.
	private pxPerSecond = 100;
	// The pixels per second and the width of the view when we last drew the rows.
	private drawnPxPerSecond = 0;
	private drawnWidth = 0;
	private contentOrigin = 0;
	private zoomAnimation: ZoomAnimation | null = null;
	private settleTimer = 0;

	// Fires when the zoom changes, also from Alt + scroll.
	readonly zoomEmitter = new TypedEvent<void>('TimelineZoom');
	// Fires when the part of the iteration we can see changes, from a zoom or a scroll. The
	// session strip draws that window over the session.
	readonly viewEmitter = new TypedEvent<void>('TimelineView');
	private viewFrame = 0;
	// The dimmed parts outside the selection, in a leveling session. They follow the selection
	// without drawing the rows again.
	private selectionElem: HTMLElement | null = null;

	// The chart's lines we don't show, by name. Threat starts hidden, because its axis climbs all
	// fight and squeezes the other lines, and most of the time we look at DPS and mana.
	private readonly hiddenChartSeries = new Set<string>(['Threat']);

	private resetCallbacks: (() => void)[] = [];

	constructor(config: TimelineConfig) {
		config.rootCssClass = 'timeline-root';
		super(config);
		this.filter = config.filter;
		this.resultData = null;
		this.rendered = false;

		this.rootElem.appendChild(
			<div className="timeline-plots-container">
				<div className="timeline-plot dps-resources-plot hide">
					<div className="chart-legend" attributes={{ role: 'group', 'aria-label': 'Lines' }}></div>
					<div className="dps-resources-chart"></div>
					<div className="leveling-vitals hide">
						<div className="chart-legend" attributes={{ role: 'group', 'aria-label': 'Lines' }}></div>
						<div className="leveling-vitals-chart"></div>
					</div>
				</div>
				<div className="timeline-plot rotation-plot">
					<div className="rotation-container">
						<div className="rotation-labels"></div>
						<div className="rotation-timeline" draggable={true}></div>
					</div>
				</div>
				<div className="events-empty hide">No rows match the filter.</div>
			</div>,
		);

		this.dpsResourcesPlotElem = this.rootElem.querySelector('.dps-resources-plot')!;
		this.chartLegend = this.dpsResourcesPlotElem.querySelector('.chart-legend')!;
		this.vitalsPlotElem = this.dpsResourcesPlotElem.querySelector('.leveling-vitals')!;
		this.vitalsLegend = this.vitalsPlotElem.querySelector('.chart-legend')!;
		this.dpsResourcesPlot = this.makeDpsResourcesPlot();
		this.addOnDisposeCallback(() => {
			this.dpsResourcesPlot.destroy();
			this.vitalsPlot?.destroy();
		});

		this.rotationPlotElem = this.rootElem.querySelector('.rotation-plot')!;
		this.rotationLabels = this.rootElem.querySelector('.rotation-labels')!;
		this.rotationTimeline = this.rootElem.querySelector('.rotation-timeline')!;
		this.emptyElem = this.rootElem.querySelector('.events-empty')!;

		let isMouseDown = false;
		let startX = 0;
		let scrollLeft = 0;
		this.rotationTimeline.addEventListener('dragstart', event => {
			event.preventDefault();
		});
		this.rotationTimeline.addEventListener('mousedown', event => {
			isMouseDown = true;
			startX = event.pageX - this.rotationTimeline.offsetLeft;
			scrollLeft = this.rotationTimeline.scrollLeft;
			this.rotationTimeline.classList.add('dragging');
		});
		const stopDrag = () => {
			isMouseDown = false;
			this.rotationTimeline.classList.remove('dragging');
		};
		this.rotationTimeline.addEventListener('mouseleave', stopDrag);
		this.rotationTimeline.addEventListener('mouseup', stopDrag);
		this.rotationTimeline.addEventListener('mousemove', event => {
			if (!isMouseDown) return;
			event.preventDefault();
			const x = event.pageX - this.rotationTimeline.offsetLeft;
			const walk = x - startX;
			this.rotationTimeline.scrollLeft = scrollLeft - walk;
		});
		this.rotationTimeline.addEventListener('scroll', () => this.emitView());

		this.setUpWheelZoom();

		// The zoom is in seconds, so a wider or narrower view needs new pixels per second. The view
		// also gets its width back when the tab opens again.
		const resizeObserver = new ResizeObserver(() => this.refitIfResized());
		resizeObserver.observe(this.rotationTimeline);
		this.addOnDisposeCallback(() => resizeObserver.disconnect());

		let lastRange = [this.filter.from, this.filter.to];
		this.filter.changeEmitter.on(() => {
			// A new time range moves everything, so we draw the timeline again. The other filters
			// only show and hide rows.
			const rangeChanged = lastRange[0] != this.filter.from || lastRange[1] != this.filter.to;
			lastRange = [this.filter.from, this.filter.to];
			// A session's rotation always has the whole session drawn. A new range only zooms and
			// scrolls to it, which is much faster than drawing 20 min of rows again.
			if (rangeChanged && this.session && this.mode == 'rotation') {
				if (this.hasSelection()) this.zoomSpan = 'selection';
				// While another view shows, we zoom to the selection when the timeline draws next.
				if (!this.dirty && this.isVisible()) {
					this.applyFilter();
					this.showSelection();
				} else {
					this.invalidate();
				}
			} else if (rangeChanged) {
				this.invalidate();
			} else {
				this.applyFilter();
			}
		});
	}

	// The chart's base options. A session's chart takes other options, so going back to a normal
	// fight starts over from these.
	private makeDpsResourcesPlot(): any {
		return new ApexCharts(this.dpsResourcesPlotElem.querySelector('.dps-resources-chart')!, {
			chart: {
				animations: {
					enabled: false,
				},
				background: 'transparent',
				foreColor: '#8e948f',
				fontFamily: 'inherit',
				// A fixed height, legend included. With '100%' ApexCharts measured the wrong parent and
				// drew past the bottom of the card, which cut the legend off.
				height: 512,
				id: 'dpsResources',
				type: 'line',
				zoom: {
					enabled: true,
					allowMouseWheelZoom: false,
				},
			},
			grid: {
				borderColor: 'rgba(255, 255, 255, 0.06)',
			},
			series: [], // Set dynamically
			xaxis: {
				title: {
					text: 'Time (s)',
				},
			},
			noData: {
				text: 'Waiting for data...',
			},
			stroke: {
				width: 2,
				curve: 'straight',
			},
		});
	}

	onSimResult(resultData: SimResultData) {
		this.resultData = resultData;
		const session = getLevelingSession(resultData.result);
		// A session opens on its selection, or a minute from its start. A normal fight after one
		// gets the fight's zoom back.
		if (session && !this.session) {
			this.zoomSpan = this.hasSelection() ? 'selection' : 60;
		} else if (!session && (this.zoomSpan == 'selection' || this.zoomSpan == 'session')) {
			this.zoomSpan = 'fight';
		}
		this.session = session;
		this.invalidate();
	}

	// The part of the iteration we can see, in seconds. In the chart it's the time range of the
	// axis.
	getViewRange(): { start: number; end: number } {
		if (this.mode != 'rotation' || !this.drawnWidth) return { start: this.viewStart, end: this.viewEnd };
		const start = this.leftToTime(this.rotationTimeline.scrollLeft);
		return { start, end: Math.min(this.viewEnd, start + this.rotationTimeline.clientWidth / this.pxPerSecond) };
	}

	// Lets the strip know the view moved, once a frame at most.
	private emitView() {
		if (this.viewFrame) return;
		this.viewFrame = requestAnimationFrame(() => {
			this.viewFrame = 0;
			this.viewEmitter.emit(TypedEvent.nextEventID());
		});
	}

	private hasSelection(): boolean {
		return this.filter.from != null || this.filter.to != null;
	}

	// The selection of a session, clipped to it. The whole session when there is none.
	private selectionRange(): { start: number; end: number } {
		const start = Math.max(this.viewStart, Math.min(this.filter.from ?? this.viewStart, this.viewEnd));
		const end = Math.min(this.viewEnd, this.filter.to ?? this.viewEnd);
		return end > start ? { start, end } : { start, end: Math.min(this.viewEnd, start + 1) };
	}

	// Zooms the session's rotation to fit the selection and dims the rest. With no selection we
	// keep the zoom and only take the dimming away.
	private showSelection() {
		this.drawSelection();
		if (!this.hasSelection()) return;
		this.zoomSpan = 'selection';
		this.zoomEmitter.emit(TypedEvent.nextEventID());
		this.zoomAnimation = null;
		window.clearTimeout(this.settleTimer);
		this.setPxPerSecond(this.fitPxPerSecond(this.drawnWidth));
		this.rotationTimeline.scrollLeft = this.timeToLeftValue(this.selectionRange().start);
		// The rows were drawn for another zoom. The resource charts stretch, but the labels on
		// them should be as many as fit, so we draw again once the zoom settles.
		this.settleTimer = window.setTimeout(() => {
			if (Math.abs(this.pxPerSecond - this.drawnPxPerSecond) > 1e-6) this.redrawKeepingScroll();
		}, 150);
		this.emitView();
	}

	// The parts of the session outside the selection get a dark veil, and the selection's edges a
	// gold line, so we can scroll past the selection and still see where it was.
	private drawSelection() {
		this.selectionElem?.remove();
		this.selectionElem = null;
		if (!this.session || !this.hasSelection()) return;
		const { start, end } = this.selectionRange();
		this.selectionElem = (
			<div className="rotation-selection">
				<div className="rotation-selection-veil" style={{ left: '0px', width: this.timeToPx(start - this.viewStart) }} />
				<div className="rotation-selection-edge" style={{ left: this.timeToLeft(start) }} />
				<div className="rotation-selection-edge" style={{ left: this.timeToLeft(end) }} />
				<div className="rotation-selection-veil" style={{ left: this.timeToLeft(end), width: this.timeToPx(this.viewEnd - end) }} />
			</div>
		) as HTMLElement;
		this.rotationTimeline.appendChild(this.selectionElem);
	}

	setMode(mode: TimelineMode) {
		if (this.mode == mode) return;
		this.mode = mode;
		this.dpsResourcesPlotElem.classList.toggle('hide', mode != 'chart');
		this.rotationPlotElem.classList.toggle('hide', mode != 'rotation');
		this.invalidate();
	}

	getZoom(): ZoomSpan {
		return this.zoomSpan;
	}

	// Shows this many seconds of the rotation, keeping the time at the left of the view where it
	// is.
	// Selection zooms to the selection again even when it's the zoom already, because we may have
	// scrolled away from it.
	setZoom(span: ZoomSpan) {
		if (span == this.zoomSpan && span != 'selection') return;
		// Selection and Session start at their own start, the others keep the time at the left.
		const anchorTime = span == 'selection' ? this.selectionRange().start : span == 'session' ? this.viewStart : undefined;
		this.zoomAround(span, 0, anchorTime);
	}

	// Zooms so the time at this many pixels from the left of the view stays there, or puts
	// anchorTime there when we give one. While the timeline is hidden, we only remember the zoom
	// for when it draws.
	private zoomAround(span: ZoomSpan, anchorX: number, anchorTimeOverride?: number) {
		this.zoomSpan = span;
		this.zoomEmitter.emit(TypedEvent.nextEventID());
		if (this.dirty || !this.isVisible()) {
			this.invalidate();
			return;
		}
		const anchorTime = anchorTimeOverride ?? this.leftToTime(this.rotationTimeline.scrollLeft + anchorX);
		if (!this.zoomAnimation) requestAnimationFrame(now => this.stepZoomAnimation(now));
		this.zoomAnimation = { anchorTime, anchorX, lastFrame: this.zoomAnimation?.lastFrame ?? performance.now() };
	}

	// Each frame moves the zoom part of the way to its target, so it eases in and out. Every
	// position on the timeline is a time times the --px-per-second variable, so a frame only sets
	// that variable and draws the ruler again. Once the zoom stops, we draw the rows again for the
	// new zoom. That brings back the right amount of resource labels and the detail of the
	// resource charts.
	private stepZoomAnimation(now: number) {
		const animation = this.zoomAnimation;
		if (!animation) return;
		const target = this.fitPxPerSecond(this.drawnWidth);
		// The first frame can come with a timestamp from before the wheel event, so we count every
		// frame as at least a few milliseconds.
		const progress = 1 - Math.exp(-Math.max(4, now - animation.lastFrame) / zoomAnimationMs);
		animation.lastFrame = now;
		// We step in log space, so zooming in and zooming out by the same factor take the same time.
		const remaining = Math.log(target / this.pxPerSecond);
		const done = Math.abs(remaining) < 0.002;
		this.setPxPerSecond(done ? target : this.pxPerSecond * Math.exp(remaining * progress));
		this.rotationTimeline.scrollLeft = this.timeToLeftValue(animation.anchorTime) - animation.anchorX;
		this.emitView();

		window.clearTimeout(this.settleTimer);
		if (done) {
			this.zoomAnimation = null;
			// The wheel sends a notch at a time, so we wait a little for the next one before we draw
			// the rows.
			this.settleTimer = window.setTimeout(() => {
				if (!this.zoomAnimation && Math.abs(this.pxPerSecond - this.drawnPxPerSecond) > 1e-6) this.redrawKeepingScroll();
			}, 150);
		} else {
			requestAnimationFrame(next => this.stepZoomAnimation(next));
		}
	}

	private setPxPerSecond(pxPerSecond: number) {
		this.pxPerSecond = pxPerSecond;
		this.rotationTimeline.style.setProperty('--px-per-second', `${pxPerSecond}px`);
		const header = this.rotationTimeline.querySelector<HTMLElement>('.rotation-timeline-header');
		if (header) this.drawRotationTimeRuler(header);
	}

	// Alt + scroll zooms around the mouse. Each notch of the wheel shows about 15% more or less
	// of the rotation.
	private setUpWheelZoom() {
		this.rotationTimeline.addEventListener(
			'wheel',
			event => {
				if (!event.altKey || this.dirty) return;
				event.preventDefault();
				const delta = (event.deltaY || event.deltaX) * (event.deltaMode == WheelEvent.DOM_DELTA_LINE ? 33 : 1);
				if (!delta) return;
				const anchorX = event.clientX - this.rotationTimeline.getBoundingClientRect().left;
				// A notch during the animation goes on from where the last one was going.
				const seconds = this.zoomSeconds(this.zoomSpan) * Math.exp(delta * 0.0015);
				// Once the whole time range fits, we stop there and call it the whole fight, or the whole
				// session.
				const whole = this.session ? 'session' : 'fight';
				this.zoomAround(seconds >= this.viewEnd - this.viewStart ? whole : this.zoomSeconds(seconds), anchorX);
			},
			{ passive: false },
		);
	}

	// How many seconds this zoom shows in the view. It shows no more than the time range, and no
	// less than we can see at the closest zoom.
	private zoomSeconds(span: ZoomSpan): number {
		const rangeSeconds = this.viewEnd - this.viewStart;
		if (span == 'fight' || span == 'session') return rangeSeconds;
		if (span == 'selection') {
			const { start, end } = this.selectionRange();
			span = end - start;
		}
		return Math.min(rangeSeconds, Math.max(this.drawnWidth / maxPxPerSecond, span));
	}

	// The pixels per second that show the zoom's seconds in a view this wide.
	private fitPxPerSecond(width: number): number {
		return width / this.zoomSeconds(this.zoomSpan);
	}

	// Draws the rotation again when its view got wider or narrower since we drew it.
	private refitIfResized() {
		const width = this.rotationTimeline.clientWidth;
		if (this.mode != 'rotation' || this.dirty || !width || Math.abs(width - this.drawnWidth) < 1) return;
		this.redrawKeepingScroll();
	}

	// Draws the rotation again, and keeps the time at the left of the view where it is.
	private redrawKeepingScroll() {
		const leftTime = this.leftToTime(this.rotationTimeline.scrollLeft);
		this.dirty = true;
		this.renderIfDirty();
		this.rotationTimeline.scrollLeft = this.timeToLeftValue(leftTime);
	}

	// Scrolls the rotation so this time shows near its left side.
	scrollToTime(time: number) {
		this.renderIfDirty();
		this.rotationTimeline.scrollLeft = Math.max(0, this.timeToLeftValue(time) - this.rotationTimeline.clientWidth / 5);
		this.rotationTimeline.querySelector('.rotation-timeline-marker')?.remove();
		const marker = <div className="rotation-timeline-marker" style={{ left: this.timeToLeft(time) }} />;
		this.rotationTimeline.appendChild(marker);
	}

	private invalidate() {
		this.dirty = true;
		if (this.isVisible()) this.renderIfDirty();
	}

	private isVisible(): boolean {
		return this.rootElem.offsetParent != null;
	}

	// The timeline needs its size to draw, so it only draws while we can see it.
	renderIfDirty() {
		if (!this.dirty || !this.isVisible()) return;
		this.dirty = false;
		this.reset();
		if (!this.rendered) {
			this.dpsResourcesPlot.render();
			this.rendered = true;
		}
		this.updatePlot();
		// The labels on the left get their width from the rows we just drew. When that changed
		// the width of the view, we draw once more to fit it.
		this.refitIfResized();
	}

	private updatePlot() {
		// We keep the width even when we draw no rotation, like before the first result or for more
		// than one player. Otherwise refitIfResized would draw again and again for a width we never
		// take.
		this.drawnWidth = this.rotationTimeline.clientWidth;
		if (this.resultData == null) {
			return;
		}

		const duration = this.resultData!.result.result.firstIterationDuration || 1;
		if (this.session && this.mode == 'rotation') {
			// A session has no prepull, and its rotation always has the whole session, so we can
			// scroll from one pull to the next. The From and To range zooms to its part instead,
			// see showSelection.
			this.prepullStart = 0;
			this.viewStart = 0;
			this.viewEnd = duration;
		} else {
			// The prepull actions, like a totem 3 seconds before the pull, happen at negative times.
			// Without a From time, the timeline starts at the first of them, on a whole second.
			const firstEvent = this.session ? 0 : this.resultData!.result.logs.reduce((first, log) => Math.min(first, log.timestamp), 0);
			this.prepullStart = Math.floor(firstEvent);
			this.viewStart = Math.max(this.prepullStart, Math.min(this.filter.from ?? this.prepullStart, duration));
			this.viewEnd = Math.min(duration, this.filter.to ?? duration);
			if (this.viewEnd <= this.viewStart) {
				this.viewEnd = Math.min(duration, this.viewStart + 1);
			}
		}

		const players = this.resultData!.result.getRaidIndexedPlayers(this.resultData!.filter);

		// The rotation follows one player. With more than one we can only show the chart.
		if (this.mode == 'rotation' && players.length == 1) {
			this.zoomAnimation = null;
			window.clearTimeout(this.settleTimer);
			this.drawnPxPerSecond = this.fitPxPerSecond(this.drawnWidth);
			this.setPxPerSecond(this.drawnPxPerSecond);
			try {
				this.updateRotationChart(players[0], duration);
			} catch (e) {
				console.log('Failed to update rotation chart: ', e);
			}
			if (this.session) {
				this.drawSelection();
				if (this.zoomSpan == 'selection') this.rotationTimeline.scrollLeft = this.timeToLeftValue(this.selectionRange().start);
			}
			this.emitView();
			return;
		}
		this.clearRotationChart();
		this.dpsResourcesPlotElem.classList.remove('hide');
		this.rotationPlotElem.classList.add('hide');

		// The chart keeps options from the last time we drew it, so a switch between a session
		// and a normal fight makes it again from the base options.
		const levelingChart = !!this.session && players.length == 1;
		if (levelingChart != this.chartIsLeveling) {
			this.chartIsLeveling = levelingChart;
			this.dpsResourcesPlot.destroy();
			this.dpsResourcesPlot = this.makeDpsResourcesPlot();
			this.dpsResourcesPlot.render();
		}
		this.vitalsPlotElem.classList.toggle('hide', !levelingChart);
		if (levelingChart) {
			this.updateLevelingChart(this.session!, duration);
			return;
		}

		const options: any = {
			theme: {
				mode: 'dark',
			},
			series: [],
			colors: [],
			xaxis: {
				type: 'numeric',
				min: this.viewStart,
				max: this.viewEnd,
				tickAmount: 10,
				decimalsInFloat: 1,
				labels: {
					show: true,
				},
				title: {
					text: 'Time (s)',
				},
			},
			yaxis: [],
			chart: {
				events: {
					beforeResetZoom: () => {
						return {
							xaxis: {
								min: this.viewStart,
								max: this.viewEnd,
							},
						};
					},
				},
				toolbar: {
					show: false,
				},
			},
		};

		let tooltipHandlers: Array<TooltipHandler | null> = [];
		options.tooltip = {
			enabled: true,
			custom: (data: { series: any; seriesIndex: number; dataPointIndex: number; w: any }) => {
				if (tooltipHandlers[data.seriesIndex]) {
					return tooltipHandlers[data.seriesIndex]!(data.dataPointIndex);
				} else {
					throw new Error('No tooltip handler for series ' + data.seriesIndex);
				}
			},
		};

		if (players.length == 1) {
			const player = players[0];
			const resourceType = chartResources.find(type => player.groupedResourceLogs[type]?.some(log => log.timestamp >= 0));
			const color = resourceType == ResourceType.ResourceTypeRage ? rageDpsColor : dpsColor;
			const dpsData = this.addDpsSeries(player, options, '', duration, color);
			this.addDpsYAxis(dpsData.maxDps, options, color);
			tooltipHandlers.push(dpsData.tooltipHandler);
			const continuousRegen = this.resultData!.result.request.simOptions?.ruleset != Ruleset.RulesetClassic;
			if (resourceType !== undefined) tooltipHandlers.push(this.addResourceSeries(player, resourceType, options, continuousRegen));
			tooltipHandlers.push(this.addThreatSeries(player, options, ''));
			tooltipHandlers = tooltipHandlers.filter(handler => handler != null);

			this.addMajorCooldownAnnotations(player, options);
		} else {
			// Each player's line is named for them, so the legend tells them apart and the DPS axis
			// can name all of them.
			let maxDps = 0;
			players.forEach(player => {
				const dpsData = this.addDpsSeries(player, options, classColors[specToClass[player.spec]], duration, '', player.label);
				maxDps = Math.max(maxDps, dpsData.maxDps);
				tooltipHandlers.push(dpsData.tooltipHandler);
			});
			this.addDpsYAxis(maxDps, options, dpsColor, players.map(player => player.label));
		}

		options.series.forEach((series: { name: string; hidden?: boolean }) => (series.hidden = this.hiddenChartSeries.has(series.name)));
		options.legend = { show: false };
		this.dpsResourcesPlot.updateOptions(options);
		this.drawChartLegend(
			options.series.map((series: { name: string }, i: number) => ({ name: series.name, color: options.colors[i] })),
			this.dpsResourcesPlot,
			this.chartLegend,
		);
	}

	// A session's chart, as two charts on one time axis. The upper one has the mobs killed so far
	// against the average session's pace, so we can see where this session fell behind or got
	// ahead. The lower one has health and mana in percent. Both get the session's states as faint
	// bands and a red line at each death. The DPS over a session's downtime says little, so it
	// isn't here.
	//
	// The axis covers the selection, or the whole session without one. A drag on either chart
	// zooms both.
	private updateLevelingChart(session: LevelingSession, duration: number) {
		const xMin = this.hasSelection() ? this.selectionRange().start : 0;
		const xMax = this.hasSelection() ? this.selectionRange().end : duration;
		const averageKills = this.resultData!.result.result.encounterMetrics?.leveling?.mobsKilled ?? 0;

		const bands = session.spans
			.filter(span => span.state != 'fighting' && span.end > span.start)
			.map(span => ({ x: span.start, x2: span.end, fillColor: SESSION_STATES[span.state].color, opacity: 0.1, borderColor: 'transparent' }));
		const deathLines = session.deaths.map(death => ({
			x: death.t,
			borderColor: SESSION_STATES.dead.color,
			strokeDashArray: 0,
			label: {
				text: 'Died',
				borderColor: 'transparent',
				orientation: 'horizontal',
				style: { background: 'transparent', color: SESSION_STATES.dead.color, fontSize: '11px' },
			},
		}));
		const annotations = { position: 'back', xaxis: [...bands, ...deathLines] };

		const xaxis = {
			type: 'numeric',
			min: xMin,
			max: xMax,
			// A tick a minute over the whole session, and ApexCharts' own steps over a short selection.
			tickAmount: xMax - xMin >= 300 ? Math.min(20, Math.round((xMax - xMin) / 60)) : 12,
			labels: { formatter: (value: string) => formatClock(parseFloat(value)) },
			title: { text: undefined },
			tooltip: { enabled: false },
		};
		const chart = (id: string, height: number) => ({
			id,
			group: 'leveling-session',
			height,
			type: 'line',
			animations: { enabled: false },
			background: 'transparent',
			foreColor: '#8e948f',
			fontFamily: 'inherit',
			toolbar: { show: false },
			zoom: { enabled: true, allowMouseWheelZoom: false },
			events: { beforeResetZoom: () => ({ xaxis: { min: xMin, max: xMax } }) },
		});

		// Mobs killed is a step up at each kill. The average pace is a straight line from none to the
		// average session's kills at the end.
		const killPoints: Array<{ x: number; y: number }> = [{ x: 0, y: 0 }];
		session.kills.forEach(({ t }, i) => killPoints.push({ x: t, y: i }, { x: t, y: i + 1 }));
		killPoints.push({ x: duration, y: session.kills.length });
		const killsAt = (t: number) => session.kills.filter(kill => kill.t <= t).length;
		const paceAt = (t: number) => (averageKills * t) / duration;
		const stateLine = (t: number) => {
			const state = session.stateAt(t);
			const pull = session.sectionAt(t, 'pull')?.pull;
			return `${SESSION_STATES[state].label}${state == 'fighting' && pull ? `, pull ${pull.number}` : ''}`;
		};
		const tooltipAt = (t: number, rows: Array<JSX.Element>) => (
			<div className="timeline-tooltip">
				<div className="timeline-tooltip-title">
					<span className="timeline-tooltip-name">{formatClock(t, true)}</span>
					<span className="timeline-tooltip-note">{stateLine(t)}</span>
				</div>
				<div className="timeline-tooltip-rows">{rows}</div>
			</div>
		);
		const xOf = (w: any, seriesIndex: number, dataPointIndex: number): number => w.globals.seriesX[seriesIndex]?.[dataPointIndex] ?? 0;

		const yMax = Math.max(1, Math.ceil(Math.max(session.kills.length, averageKills)));
		const killsSeries = [
			{ name: 'Mobs killed', type: 'line', data: killPoints },
			{
				name: 'Average pace',
				type: 'line',
				data: [
					{ x: 0, y: 0 },
					{ x: duration, y: averageKills },
				],
			},
		];
		const killsOptions: any = {
			theme: { mode: 'dark' },
			chart: chart('dpsResources', 300),
			colors: [killsColor, paceColor],
			stroke: { width: [2, 2], curve: 'straight', dashArray: [0, 6] },
			series: killsSeries,
			xaxis: { ...xaxis, labels: { show: false } },
			yaxis: [
				{
					min: 0,
					max: yMax,
					tickAmount: Math.min(10, yMax),
					decimalsInFloat: 0,
					title: { text: 'Mobs killed' },
					labels: { minWidth: 40 },
				},
			],
			grid: { borderColor: 'rgba(255, 255, 255, 0.06)' },
			annotations,
			legend: { show: false },
			tooltip: {
				enabled: true,
				custom: ({ seriesIndex, dataPointIndex, w }: any) => {
					const t = xOf(w, seriesIndex, dataPointIndex);
					return tooltipAt(t, [
						tooltipRow('Mobs killed', String(killsAt(t))),
						tooltipRow('Average pace', formatAmount(paceAt(t), 1)),
					]);
				},
			},
		};
		killsOptions.series.forEach((series: { name: string; hidden?: boolean }) => (series.hidden = this.hiddenChartSeries.has(series.name)));
		// The two charts are in one group so they zoom and point at the same time together. An
		// update would also go to the other chart of the group by default, and give it this chart's
		// lines and dashes, so we keep each update to its own chart.
		this.dpsResourcesPlot.updateOptions(killsOptions, false, false, false);
		this.drawChartLegend(
			killsSeries.map((series, i) => ({ name: series.name, color: killsOptions.colors[i] })),
			this.dpsResourcesPlot,
			this.chartLegend,
		);

		// Health and mana change many times a second, so we keep one point per small step of the
		// session. Health keeps the lowest value of each step too, so a near death still shows.
		const step = duration / 2400;
		const thin = (samples: Array<{ t: number; value: number }>, keepLowest: boolean) => {
			const points: Array<{ x: number; y: number }> = [];
			let bucketEnd = -Infinity;
			let lowest: { t: number; value: number } | null = null;
			let last: { t: number; value: number } | null = null;
			const flush = () => {
				if (keepLowest && lowest && last && lowest !== last) points.push({ x: lowest.t, y: lowest.value * 100 });
				if (last) points.push({ x: last.t, y: last.value * 100 });
			};
			samples.forEach(sample => {
				// After a long gap with no change, like the corpse run, the old value held until now. A
				// straight line to the new one would draw health coming back slowly.
				if (last && sample.t - last.t > 2 * step) {
					flush();
					points.push({ x: sample.t, y: last.value * 100 });
					lowest = null;
					last = null;
					bucketEnd = sample.t + step;
				}
				if (sample.t >= bucketEnd) {
					flush();
					bucketEnd = sample.t + step;
					lowest = null;
				}
				if (!lowest || sample.value < lowest.value) lowest = sample;
				last = sample;
			});
			flush();
			if (points.length) points.push({ x: duration, y: points[points.length - 1].y });
			return points;
		};
		const healthPoints = [{ x: 0, y: 100 }, ...thin(session.health, true)];
		const manaPoints = session.mana.length ? [{ x: 0, y: 100 }, ...thin(session.mana, false)] : [];
		const vitalsSeries = [{ name: 'Health', type: 'line', data: healthPoints }];
		if (manaPoints.length) vitalsSeries.push({ name: 'Mana', type: 'line', data: manaPoints });
		const colors = [levelingHealthColor, levelingManaColor].slice(0, vitalsSeries.length);
		const vitalsOptions: any = {
			theme: { mode: 'dark' },
			chart: chart('levelingVitals', 240),
			colors,
			stroke: { width: 2, curve: 'straight', dashArray: 0 },
			series: vitalsSeries,
			noData: { text: '' },
			xaxis,
			yaxis: [
				{
					min: 0,
					max: 100,
					tickAmount: 4,
					title: { text: 'Health and mana' },
					labels: { minWidth: 40, formatter: (value: number) => `${value.toFixed(0)}%` },
				},
			],
			grid: { borderColor: 'rgba(255, 255, 255, 0.06)' },
			annotations,
			legend: { show: false },
			tooltip: {
				enabled: true,
				custom: ({ seriesIndex, dataPointIndex, w }: any) => {
					const t = xOf(w, seriesIndex, dataPointIndex);
					const rows = [tooltipRow('Health', `${(valueAt(session.health, t, 1) * 100).toFixed(0)}%`)];
					if (session.mana.length) rows.push(tooltipRow('Mana', `${(valueAt(session.mana, t, 1) * 100).toFixed(0)}%`));
					return tooltipAt(t, rows);
				},
			},
		};
		vitalsOptions.series.forEach((series: { name: string; hidden?: boolean }) => (series.hidden = this.hiddenChartSeries.has(series.name)));
		if (!this.vitalsPlot) {
			this.vitalsPlot = new ApexCharts(this.vitalsPlotElem.querySelector('.leveling-vitals-chart')!, vitalsOptions);
			this.vitalsPlot.render();
		} else {
			this.vitalsPlot.updateOptions(vitalsOptions, false, false, false);
		}
		this.drawChartLegend(
			vitalsSeries.map((series, i) => ({ name: series.name, color: colors[i] })),
			this.vitalsPlot,
			this.vitalsLegend,
		);
	}

	// Our own legend above the chart, with a chip for each line like the filters. A chip shows or
	// hides its line, and we remember that for the next iteration we draw.
	private drawChartLegend(lines: Array<{ name: string; color: string }>, plot: any, legend: HTMLElement) {
		legend.replaceChildren(
			...lines.map(({ name, color }) => {
				const chip = (
					<button className="filter-chip chart-legend-chip" type="button">
						<span className="chart-legend-swatch" />
						{name}
					</button>
				) as HTMLButtonElement;
				chip.style.setProperty('--swatch', color);
				const update = () => {
					const shown = !this.hiddenChartSeries.has(name);
					chip.classList.toggle('active', shown);
					chip.setAttribute('aria-pressed', String(shown));
				};
				chip.addEventListener('click', () => {
					if (this.hiddenChartSeries.delete(name)) {
						plot.showSeries(name);
					} else {
						this.hiddenChartSeries.add(name);
						plot.hideSeries(name);
					}
					update();
				});
				update();
				return chip;
			}),
		);
	}

	private addDpsYAxis(maxDps: number, options: any, color = dpsColor, seriesName: string | Array<string> = 'DPS') {
		const dpsAxisMax = Math.ceil(maxDps / 100) * 100;
		options.yaxis.push({
			color,
			seriesName,
			min: 0,
			max: dpsAxisMax,
			tickAmount: 10,
			decimalsInFloat: 0,
			title: {
				text: 'DPS',
				style: {
					color,
				},
			},
			axisBorder: {
				show: true,
				color,
			},
			axisTicks: {
				color,
			},
			labels: {
				minWidth: 30,
				style: {
					colors: [color],
				},
			},
		});
	}

	// Returns a function for drawing the tooltip, or null if no series was added.
	//
	// The DPS is the damage of the last 15 sec, so it changes at two kinds of moments: when we
	// deal damage, and when an old hit leaves the window 15 sec later. We draw a point at both.
	// When nothing hits for a while, like while we wait on mana, the line then falls at the
	// moments it really falls. Each change is a jump, so we draw the value just before it and
	// just after it at the same time.
	//
	// Between 5 and 15 sec into the fight we divide by the time so far (see DpsLog.windowAt), so
	// the value also falls slowly between changes. We add a point every half second there.
	//
	// lineColor replaces the usual red, like next to rage (see rageDpsColor).
	private addDpsSeries(
		unit: UnitMetrics,
		options: any,
		colorOverride: string,
		duration: number,
		lineColor = '',
		name = 'DPS',
	): { maxDps: number; tooltipHandler: TooltipHandler } {
		const dpsLogs = unit.dpsLogs;
		const changes = new Map<number, number>();
		const addChange = (timestamp: number, amount: number) => changes.set(timestamp, (changes.get(timestamp) ?? 0) + amount);
		const logsAt = new Map<number, DpsLog>();
		dpsLogs.forEach(log => {
			const amount = sum(log.damageLogs.map(damageLog => damageLog.amount));
			addChange(log.timestamp, amount);
			addChange(log.timestamp + DpsLog.DPS_WINDOW, -amount);
			logsAt.set(log.timestamp, log);
		});
		for (let t = DpsLog.DPS_MIN_WINDOW; t < DpsLog.DPS_WINDOW; t += 0.5) addChange(t, 0);
		addChange(0, 0);
		addChange(duration, 0);

		// Each point keeps the hits at its time (none at a window exit) and the last hits at or
		// before it, for the auras in the tooltip.
		const points: Array<{ timestamp: number; dps: number; log: DpsLog | null; aurasLog: DpsLog | null }> = [];
		let damage = 0;
		let lastLog: DpsLog | null = null;
		[...changes.keys()]
			.sort((a, b) => a - b)
			.forEach(timestamp => {
				const window = DpsLog.windowAt(timestamp);
				const before = damage / window;
				damage += changes.get(timestamp)!;
				// Floating point sums can leave a tiny leftover when the window empties.
				if (Math.abs(damage) < 1e-6) damage = 0;
				const log = logsAt.get(timestamp) ?? null;
				if (log) lastLog = log;
				// The hits before the pull count towards the window, but we don't draw them.
				if (timestamp < 0 || timestamp > duration) return;
				const after = damage / window;
				if (Math.abs(after - before) > 1e-9) points.push({ timestamp, dps: before, log: null, aurasLog: lastLog });
				points.push({ timestamp, dps: after, log, aurasLog: lastLog });
			});

		options.colors.push(lineColor || colorOverride || dpsColor);
		options.series.push({
			name,
			type: 'line',
			data: points.map(point => ({ x: point.timestamp, y: point.dps })),
		});

		return {
			maxDps: Math.max(1, ...points.map(point => point.dps)),
			tooltipHandler: (dataPointIndex: number) => {
				const point = points[dataPointIndex];
				return this.dpsTooltip(point.timestamp, point.dps, point.log, point.aurasLog, unit, colorOverride);
			},
		};
	}

	// Returns a function for drawing the tooltip. The resource is the unit's main one: mana, or
	// rage or energy for a class without mana.
	//
	// A spend is a jump, so we draw a point after each change at a moment, like the regen up to
	// it and then the cost of a cast. Under Forever, mana and energy regenerate all the time and
	// the sim adds the regen since the last event at every event, so a straight line between
	// moments is right. Under Classic, mana only changes at the 2 sec ticks and the casts, and
	// rage only changes when we hit, get hit or spend it. Then we also draw the value before the
	// moment, and the line stays flat until it.
	private addResourceSeries(unit: UnitMetrics, resourceType: ResourceType, options: any, continuousRegen: boolean): TooltipHandler {
		const allLogs = unit.groupedResourceLogs[resourceType];
		const logs = allLogs.filter(log => log.timestamp >= 0);
		const name = resourceNames.get(resourceType)!;
		const color = resourceChartColors[resourceType]!;
		const isMana = resourceType == ResourceType.ResourceTypeMana;
		const steps = !continuousRegen || resourceType == ResourceType.ResourceTypeRage;
		// The full mana pool is the mana before the first change, prepull included. The first
		// change after the pull can come after prepull totems already spent some, and then regen
		// would draw the line over the top of the axis. Rage and energy top out at 100.
		const highest = Math.max(...allLogs.map(log => Math.max(log.valueBefore, log.valueAfter)));
		const maxValue = isMana ? highest : Math.max(100, highest);

		const points: Array<{ value: number; log: ResourceChangedLogGroup }> = [];
		logs.forEach(log => {
			if (steps) points.push({ value: log.valueBefore, log });
			log.logs.forEach(changedLog => points.push({ value: changedLog.valueAfter, log }));
		});

		options.colors.push(color);
		options.series.push({
			name,
			type: 'line',
			data: points.map(point => ({ x: point.log.timestamp, y: point.value })),
		});
		options.yaxis.push({
			seriesName: name,
			opposite: true, // Appear on right side
			min: 0,
			max: maxValue,
			tickAmount: isMana ? 10 : 5,
			title: {
				text: name,
				style: {
					color,
				},
			},
			axisBorder: {
				show: true,
				color,
			},
			axisTicks: {
				color,
			},
			labels: {
				minWidth: 30,
				style: {
					colors: [color],
				},
				formatter: (val: string) => {
					const v = parseFloat(val);
					return isMana ? `${v.toFixed(0)} (${((v / maxValue) * 100).toFixed(0)}%)` : v.toFixed(0);
				},
			},
		} as any);

		return (dataPointIndex: number) => this.resourceTooltip(points[dataPointIndex].log, maxValue, true);
	}

	// Returns a function for drawing the tooltip, or null if no series was added. Threat adds up
	// over the fight to far more than the DPS, so it gets its own axis on the right. It only
	// changes when we act, so we draw the value before and after each change at the same time.
	private addThreatSeries(unit: UnitMetrics, options: any, colorOverride: string): TooltipHandler | null {
		const threatLogs = unit.threatLogs.filter(log => log.timestamp >= 0);
		if (threatLogs.length == 0) {
			return null;
		}
		const color = colorOverride || threatColor;
		const maxThreat = Math.max(...threatLogs.map(log => log.threatAfter));

		options.colors.push(color);
		options.series.push({
			name: 'Threat',
			type: 'line',
			data: threatLogs.flatMap(log => [
				{ x: log.timestamp, y: log.threatBefore },
				{ x: log.timestamp, y: log.threatAfter },
			]),
		});
		options.yaxis.push({
			seriesName: 'Threat',
			opposite: true,
			min: 0,
			max: maxThreat > 0 ? maxThreat : 1,
			tickAmount: 10,
			title: {
				text: 'Threat',
				style: {
					color,
				},
			},
			axisBorder: {
				show: true,
				color,
			},
			axisTicks: {
				color,
			},
			labels: {
				minWidth: 30,
				style: {
					colors: [color],
				},
				formatter: (val: string) => formatToCompactNumber(parseFloat(val)),
			},
		} as any);

		return (dataPointIndex: number) => {
			const log = threatLogs[Math.floor(dataPointIndex / 2)];
			return this.threatTooltip(log, true, unit, colorOverride);
		};
	}

	private addMajorCooldownAnnotations(unit: UnitMetrics, options: any) {
		const mcdLogs = unit.majorCooldownLogs;
		const mcdAuraLogs = unit.majorCooldownAuraUptimeLogs;

		// Figure out how much to vertically offset cooldown icons, for cooldowns
		// used very close to each other. This is so the icons don't overlap.
		const MAX_ALLOWED_DIST = 10;
		const cooldownIconOffsets = mcdLogs.map(
			(mcdLog, mcdIdx) => mcdLogs.filter((cdLog, cdIdx) => cdIdx < mcdIdx && cdLog.timestamp > mcdLog.timestamp - MAX_ALLOWED_DIST).length,
		);

		const distinctMcdAuras = distinct(mcdAuraLogs, (a, b) => a.actionId!.equalsIgnoringTag(b.actionId!));
		// Sort by name so auras keep their same colors even if timings change.
		distinctMcdAuras.sort((a, b) => stringComparator(a.actionId!.name, b.actionId!.name));
		const mcdAuraColors = mcdAuraLogs.map(
			mcdAuraLog => actionColors[distinctMcdAuras.findIndex(dAura => dAura.actionId!.equalsIgnoringTag(mcdAuraLog.actionId!))],
		);

		options.annotations = {
			position: 'back',
			xaxis: mcdAuraLogs.map((log, i) => {
				return {
					x: log.gainedAt,
					x2: log.fadedAt,
					fillColor: mcdAuraColors[i],
				};
			}),
			points: mcdLogs.map((log, i) => {
				return {
					x: log.timestamp,
					y: 0,
					image: {
						path: log.actionId!.iconUrl,
						width: 20,
						height: 20,
						offsetY: cooldownIconOffsets[i] * -25,
					},
				};
			}),
		};
	}

	private clearRotationChart() {
		this.rotationLabels.innerText = '';
		this.rotationLabels.appendChild(<div className="rotation-label-header"></div>);

		this.rotationTimeline.innerText = '';
		this.rotationTimeline.appendChild(<div className="rotation-timeline-header"></div>);
		this.rows = [];
		this.separators = [];
	}

	private updateRotationChart(player: UnitMetrics, duration: number) {
		// A session fights a pool of mobs, Target 1 to N, and any of them can be the one we're on.
		// So we take them all, and their casts and debuffs share a row by spell: one Flame Shock row
		// whichever mob it's on.
		const targets = this.session ? this.resultData!.result.getTargets() : this.resultData!.result.getTargets(this.resultData!.filter);
		if (targets.length == 0) {
			return;
		}
		const target = targets[0];
		const targetAuraLogs = this.session ? targets.flatMap(t => t.auraUptimeLogs).sort((a, b) => a.gainedAt - b.gainedAt) : target.auraUptimeLogs;

		this.clearRotationChart();
		this.dpsResourcesPlotElem.classList.add('hide');
		this.rotationPlotElem.classList.remove('hide');

		if (this.session) this.addSessionLane();
		orderedResourceTypes.forEach(resourceType => this.addResourceRow(resourceType, player.groupedResourceLogs[resourceType], duration));
		this.addGcdRow(player, duration);

		const buffsById = Object.values(bucket(player.auraUptimeLogs, log => log.actionId!.toString()));
		buffsById.sort((a, b) => stringComparator(a[0].actionId!.name, b[0].actionId!.name));
		const debuffsById = Object.values(bucket(targetAuraLogs, log => log.actionId!.toString()));
		debuffsById.sort((a, b) => stringComparator(a[0].actionId!.name, b[0].actionId!.name));
		const buffsAndDebuffsById = buffsById.concat(debuffsById);

		const playerCastsByAbility = this.getSortedCastsByAbility(player);
		playerCastsByAbility.forEach(castLogs => this.addCastRow(castLogs, buffsAndDebuffsById));

		if (player.pets.length > 0) {
			const playerPets = new Map<string, UnitMetrics>();
			player.pets.forEach(petsLog => {
				const petCastsByAbility = this.getSortedCastsByAbility(petsLog);
				if (petCastsByAbility.length > 0) {
					// Because mulle pets can have the same name and we parse cast logs
					// by pet name each individual pet ends up with all the casts of pets
					// with the same name. Because of this we can just grab the first pet
					// of each name and visualize only that.
					if (!playerPets.has(petsLog.name)) {
						playerPets.set(petsLog.name, petsLog);
					}
				}
			});

			playerPets.forEach(pet => {
				this.addSeparatorRow();
				this.addPetRow(pet.name);
				orderedResourceTypes.forEach(resourceType => this.addResourceRow(resourceType, pet.groupedResourceLogs[resourceType], duration));
				const petCastsByAbility = this.getSortedCastsByAbility(pet);
				petCastsByAbility.forEach(castLogs => this.addCastRow(castLogs, buffsAndDebuffsById));
			});
		}

		// Don't add a row for buffs that were already visualized in a cast row.
		const buffsToShow = buffsById.filter(auraUptimeLogs =>
			playerCastsByAbility.findIndex(casts => casts[0].actionId!.equalsIgnoringTag(auraUptimeLogs[0].actionId!)),
		);
		if (buffsToShow.length > 0) {
			this.addSeparatorRow();
			buffsToShow.forEach(auraUptimeLogs => this.addAuraRow(auraUptimeLogs, duration));
		}

		const targetCastsByAbility = this.session ? this.getSortedCastsByAbility(target, targets.flatMap(t => t.castLogs).sort((a, b) => a.timestamp - b.timestamp)) : this.getSortedCastsByAbility(target);
		if (targetCastsByAbility.length > 0) {
			this.addSeparatorRow();
			targetCastsByAbility.forEach(castLogs => this.addCastRow(castLogs, buffsAndDebuffsById));
		}

		// Add a row for all debuffs, even those which have already been visualized in a cast row.
		const debuffsToShow = debuffsById;
		if (debuffsToShow.length > 0) {
			this.addSeparatorRow();
			debuffsToShow.forEach(auraUptimeLogs => this.addAuraRow(auraUptimeLogs, duration));
		}

		this.applyFilter();

		// The ruler and the grid go last. The rows give the labels on the left their width, and that
		// tells us where the view starts on the screen. See crispLeft.
		const header = this.rotationTimeline.querySelector<HTMLElement>('.rotation-timeline-header')!;
		this.drawRotationTimeRuler(header);
		this.drawPull();
		this.drawGrid(header);
		if (this.session) this.drawSession(header, this.session);
	}

	// The session's states as bands behind the rows: hatched while we search, rest, wait or lie
	// dead, a faint blue while a found mob waits for its pull, and nothing while we fight. Each
	// kill gets a tick under the ruler and each death a red line through the rows, with a
	// tooltip. They're all in --px-per-second, so they follow the zoom without drawing again.
	private drawSession(header: HTMLElement, session: LevelingSession) {
		const top = `${header.offsetTop + header.offsetHeight}px`;
		const bands = session.spans
			.filter(span => span.state != 'fighting' && span.end > span.start)
			.map(span => {
				const band = (
					<div className={`rotation-session-band state-${span.state}`} style={{ left: this.timeToLeft(span.start), width: this.timeToPx(span.end - span.start), top }}>
						{/* A name that doesn't fit would be cut off, so a narrow band goes without. */}
						{(span.end - span.start) * this.pxPerSecond >= 72 && <span className="rotation-session-band-label">{SESSION_STATES[span.state].label}</span>}
					</div>
				) as HTMLElement;
				band.style.setProperty('--band-color', SESSION_STATES[span.state].color);
				return band;
			});
		this.rotationTimeline.prepend(<div className="rotation-session-bands">{bands}</div>);

		const markers = (<div className="rotation-session-markers" style={{ top: `${header.offsetTop}px`, height: `${header.offsetHeight}px` }} />) as HTMLElement;
		session.kills.forEach(({ t, mob }) => {
			const tick = (<div className="rotation-kill-tick" style={{ left: this.timeToLeft(t) }} />) as HTMLElement;
			markers.appendChild(tick);
			this.lazyTooltip(tick, () => (
				<div className="timeline-tooltip">
					<div className="timeline-tooltip-title">
						<span className="timeline-tooltip-name">{`${mobName(mob)} died`}</span>
						<span className="timeline-tooltip-note">{formatClock(t, true)}</span>
					</div>
					<div className="timeline-tooltip-meta">{`Level ${mob.level}${mob.pull ? `, pull ${mob.pull.number}` : ''}`}</div>
				</div>
			));
		});
		session.deaths.forEach(death => {
			const line = (<div className="rotation-death-line" style={{ left: this.timeToLeft(death.t), top }} />) as HTMLElement;
			this.rotationTimeline.appendChild(line);
			const skull = (
				<div className="rotation-death-marker" style={{ left: this.timeToLeft(death.t) }}>
					<i className="fas fa-skull" />
				</div>
			) as HTMLElement;
			markers.appendChild(skull);
			this.lazyTooltip(skull, () => (
				<div className="timeline-tooltip">
					<div className="timeline-tooltip-title">
						<span className="timeline-tooltip-name">We died</span>
						<span className="timeline-tooltip-note">{formatClock(death.t, true)}</span>
					</div>
					<div className="timeline-tooltip-meta">{`${death.pull ? `In pull ${death.pull.number}. ` : ''}Back at ${formatClock(death.backAt)} after the corpse run.`}</div>
				</div>
			));
		});
		this.rotationTimeline.appendChild(markers);
	}

	// A tooltip that we only make the first time the mouse comes over its element. A session has
	// thousands of casts, and a tippy for each one up front made drawing the rotation slow.
	private lazyTooltip(elem: HTMLElement, content: () => Element, props: Partial<Props> = {}) {
		const onEnter = () => {
			elem.removeEventListener('mouseenter', onEnter);
			const tooltip = tippy(elem, { placement: 'bottom', ...props, content: content() });
			tooltip.show();
			this.addOnResetCallback(() => tooltip.destroy());
		};
		elem.addEventListener('mouseenter', onEnter);
	}

	// Shows the rows the filter lets through, and the separators between groups that still have
	// a row. The categories and the hidden abilities hide rows. The search only fades them, see
	// applySearch, so the rows keep their places.
	private applyFilter() {
		const searching = this.filter.search.trim().length > 0;
		let anyShown = false;
		this.rows.forEach(entry => {
			if (entry == 'separator') return;
			const shown = this.isRowShown(entry);
			entry.label.classList.toggle('hide', !shown);
			entry.row.classList.toggle('hide', !shown);
			entry.events.forEach(event => {
				if (event.actionId) event.elem.classList.toggle('hide', this.filter.isHidden(event.actionId));
			});
			anyShown ||= shown;
			this.applySearch(entry, shown && searching);
		});

		// A separator shows only when there are rows shown both above and below it.
		let shownAbove = false;
		let pending: HTMLElement[] | null = null;
		let separatorIndex = 0;
		this.rows.forEach(entry => {
			if (entry == 'separator') {
				const separator = this.separators[separatorIndex++];
				separator.forEach(e => e.classList.add('hide'));
				if (shownAbove) pending = separator;
				return;
			}
			if (!entry.label.classList.contains('hide')) {
				pending?.forEach(e => e.classList.remove('hide'));
				pending = null;
				shownAbove = true;
			}
		});

		this.emptyElem.classList.toggle('hide', anyShown || this.rows.length == 0 || this.mode != 'rotation');
	}

	private isRowShown(row: RotationRow): boolean {
		if (!this.filter.categories.has(row.category)) return false;
		if (row.alwaysUp && this.filter.hideAlwaysUp) return false;
		return !(row.actionId && this.filter.isHidden(row.actionId));
	}

	// A row whose name matches the search stays as it is. When only some of its events match,
	// like the crits of Stormstrike for "crit", we outline those events and fade the others. A
	// row with no match at all fades.
	private applySearch(row: RotationRow, searching: boolean) {
		const nameMatches = !searching || this.filter.matchesText(row.name.toLowerCase());
		let anyEventMatches = false;
		row.events.forEach(event => {
			const hit = !nameMatches && this.filter.matchesText(event.text);
			event.elem.classList.toggle('search-hit', hit);
			anyEventMatches ||= hit;
		});
		row.row.classList.toggle('search-partial', anyEventMatches);
		const dim = !nameMatches && !anyEventMatches;
		row.label.classList.toggle('search-dim', dim);
		row.row.classList.toggle('search-dim', dim);
	}

	private addRow(row: RotationRow) {
		this.rows.push(row);
		this.rotationLabels.appendChild(row.label);
		this.rotationTimeline.appendChild(row.row);
	}

	// The casts of the unit by ability. castLogs takes the place of the unit's own, like the casts of
	// every mob of a session.
	private getSortedCastsByAbility(player: UnitMetrics, castLogs: Array<CastLog> = player.castLogs): Array<Array<CastLog>> {
		const meleeActionIds = player.getMeleeActions().map(action => action.actionId);
		const spellActionIds = player.getSpellActions().map(action => action.actionId);
		const getActionCategory = (actionId: ActionId): number => {
			const fixedCategory = idToCategoryMap[actionId.anyId()];
			if (fixedCategory != null) {
				return fixedCategory;
			} else if (meleeActionIds.find(meleeActionId => meleeActionId.equals(actionId))) {
				return MELEE_ACTION_CATEGORY;
			} else if (spellActionIds.find(spellActionId => spellActionId.equals(actionId))) {
				return SPELL_ACTION_CATEGORY;
			} else {
				return DEFAULT_ACTION_CATEGORY;
			}
		};

		const castsByAbility = Object.values(
			bucket(castLogs, log => {
				if (idsToGroupForRotation.includes(log.actionId!.spellId)) {
					return log.actionId!.toStringIgnoringTag();
				} else {
					return log.actionId!.toString();
				}
			}),
		);

		castsByAbility.sort((a, b) => {
			const categoryA = getActionCategory(a[0].actionId!);
			const categoryB = getActionCategory(b[0].actionId!);
			if (categoryA != categoryB) {
				return categoryA - categoryB;
			} else if (a[0].actionId!.anyId() == b[0].actionId!.anyId()) {
				return a[0].actionId!.tag - b[0].actionId!.tag;
			} else {
				return stringComparator(a[0].actionId!.name, b[0].actionId!.name);
			}
		});

		return castsByAbility;
	}

	private makeLabelElem(actionId: ActionId, isAura?: boolean): HTMLElement {
		const labelText = idsToGroupForRotation.includes(actionId.spellId) ? actionId.baseName : actionId.name;
		const labelIcon = ref<HTMLAnchorElement>();
		const hideElem = ref<HTMLButtonElement>();
		const labelElem = (
			<div className="rotation-label rotation-row">
				<a ref={labelIcon} className="rotation-label-icon"></a>
				<span className="rotation-label-text">{labelText}</span>
				<button ref={hideElem} className="rotation-label-hide" type="button">
					<i className="fas fa-eye-slash" />
				</button>
			</div>
		) as HTMLElement;
		const onClickHandler = () => this.filter.hide(actionId);
		hideElem.value!.addEventListener('click', onClickHandler);
		const tooltip = tippy(hideElem.value!, {
			placement: 'bottom',
			content: 'Hide row',
		});

		actionId.setBackgroundAndHref(labelIcon.value!);
		actionId.setWowheadDataset(labelIcon.value!, { useBuffAura: isAura });

		this.addOnResetCallback(() => {
			hideElem.value?.removeEventListener('click', onClickHandler);
			tooltip.destroy();
		});

		return labelElem;
	}

	private makeRowElem(): HTMLElement {
		return (
			<div
				className="rotation-timeline-row rotation-row"
				style={{
					width: this.viewWidthPx(),
				}}></div>
		) as HTMLElement;
	}

	private addPetRow(petName: string) {
		const actionId = ActionId.fromPetName(petName);
		const labelElem = (<div className="rotation-label rotation-row rotation-pet-label"></div>) as HTMLElement;
		this.addRow({ label: labelElem, row: this.makeRowElem(), name: petName, actionId: null, category: 'casts', events: [] });

		actionId.fill().then(filledActionId => {
			const labelText = idsToGroupForRotation.includes(filledActionId.spellId) ? filledActionId.baseName : filledActionId.name;
			const labelIcon = ref<HTMLAnchorElement>();
			labelElem.appendChild(
				<>
					<a ref={labelIcon} className="rotation-label-icon"></a>
					<span className="rotation-label-text">{labelText}</span>
				</>,
			);
			filledActionId.setBackgroundAndHref(labelIcon.value!);
		});
	}

	// An empty lane at the top of a session, where the state bands put their names (Searching,
	// Resting). Without it the names sit on the health line.
	private addSessionLane() {
		this.rotationLabels.appendChild(<div className="rotation-session-lane" />);
		this.rotationTimeline.appendChild(<div className="rotation-session-lane" style={{ width: this.viewWidthPx() }} />);
	}

	private addSeparatorRow() {
		const labelSeparator = (<div className="rotation-timeline-separator"></div>) as HTMLElement;
		const rowSeparator = (<div className="rotation-timeline-separator" style={{ width: this.viewWidthPx() }}></div>) as HTMLElement;
		this.rotationLabels.appendChild(labelSeparator);
		this.rotationTimeline.appendChild(rowSeparator);
		this.rows.push('separator');
		this.separators.push([labelSeparator, rowSeparator]);
	}

	// A resource like mana, drawn as a step chart of its value across the row. Mana changes many
	// times a second with Forever's continuous regen, so a label on every change would pile up.
	// We label a change only when there's room since the last label, and one tooltip for the row
	// shows the change under the mouse.
	private addResourceRow(resourceType: ResourceType, resourceLogs: Array<ResourceChangedLogGroup>, duration: number) {
		if (resourceLogs.length == 0) {
			return;
		}
		const startValue = resourceLogs[0].valueBefore;
		const isPercentage = percentageResources.includes(resourceType);
		// The top of the chart is the resource's cap: the full mana pool (the highest mana we saw),
		// 100 rage or energy, and 5 combo points. Others use the highest value we saw.
		const caps: Partial<Record<ResourceType, number>> = {
			[ResourceType.ResourceTypeRage]: 100,
			[ResourceType.ResourceTypeEnergy]: 100,
			[ResourceType.ResourceTypeComboPoints]: 5,
		};
		const highest = Math.max(startValue, ...resourceLogs.map(log => log.valueAfter));
		const maxValue = isPercentage ? highest : Math.max(caps[resourceType] ?? 0, highest);
		const labelText = (log: ResourceChangedLogGroup) =>
			isPercentage ? ((log.valueAfter / maxValue) * 100).toFixed(0) + '%' : Math.floor(log.valueAfter).toFixed(0);
		const cNames = resourceNames.get(resourceType)!.toLowerCase().replaceAll(' ', '-');

		const labelElem = (
			<div className="rotation-label rotation-row">
				<a
					className="rotation-label-icon"
					style={{
						backgroundImage: `url('${resourceTypeToIcon[resourceType]}')`,
					}}></a>
				<span className="rotation-label-text">{resourceNames.get(resourceType)}</span>
			</div>
		) as HTMLElement;

		const rowElem = this.makeRowElem();
		const row: RotationRow = { label: labelElem, row: rowElem, name: resourceNames.get(resourceType)!, actionId: null, category: 'resources', events: [] };

		// The chart. Changes within the same pixel only keep the last value, so a long fight at a
		// low zoom stays a short path.
		// The inside of the lane, without its line at the bottom.
		const height = 30;
		const pad = 3;
		const yOf = (value: number) => (height - pad - (Math.max(0, Math.min(1, value / (maxValue || 1))) * (height - 2 * pad))).toFixed(1);
		const steps: Array<{ x: number; y: string }> = [];
		resourceLogs.forEach(log => {
			const x = Math.round(this.timeToLeftValue(log.timestamp));
			const y = yOf(log.valueAfter);
			if (steps.length && steps[steps.length - 1].x == x) steps[steps.length - 1].y = y;
			else steps.push({ x, y });
		});
		const endX = Math.round(this.timeToLeftValue(duration));
		// Each value holds until the next change, so the path goes across, then up or down.
		let line = `M ${steps[0].x} ${steps[0].y}`;
		steps.forEach((step, i) => {
			if (i > 0) line += ` H ${step.x} V ${step.y}`;
		});
		line += ` H ${endX}`;
		const area = `${line} V ${height} H ${steps[0].x} Z`;

		const svgNs = 'http://www.w3.org/2000/svg';
		const svg = document.createElementNS(svgNs, 'svg');
		svg.setAttribute('class', `rotation-resource-chart series-color ${cNames}`);
		// The path is in pixels at the zoom we drew it. While the zoom animates, we stretch it.
		svg.setAttribute('viewBox', `0 0 ${Math.max(endX, 1)} ${height}`);
		svg.setAttribute('preserveAspectRatio', 'none');
		svg.style.width = this.timeToPx(duration - this.viewStart);
		svg.setAttribute('height', String(height));
		const areaPath = document.createElementNS(svgNs, 'path');
		areaPath.setAttribute('class', 'rotation-resource-area');
		areaPath.setAttribute('d', area);
		const linePath = document.createElementNS(svgNs, 'path');
		linePath.setAttribute('class', 'rotation-resource-line');
		linePath.setAttribute('d', line);
		linePath.setAttribute('vector-effect', 'non-scaling-stroke');
		svg.append(areaPath, linePath);
		rowElem.appendChild(svg);

		// The labels, about one per 48px. The search only looks at the lane's name, like "mana".
		// Its changes name the abilities that spent or gave it, like "from Earth Shock", and the
		// lane would light up for every ability search.
		let lastLabelX = -Infinity;
		resourceLogs.forEach(resourceLogGroup => {
			const x = this.timeToLeftValue(resourceLogGroup.timestamp);
			// In a session, a death and the revive after it always get their label (0%, then the
			// health we came back with), so the empty lane through the corpse run reads as 0%.
			const deathOrRevive =
				!!this.session && resourceType == ResourceType.ResourceTypeHealth && (resourceLogGroup.valueAfter <= 0 || resourceLogGroup.valueBefore <= 0);
			if (x - lastLabelX >= 48 || deathOrRevive) {
				lastLabelX = x;
				rowElem.appendChild(
					<span className={`rotation-resource-label series-color ${cNames}`} style={{ left: this.timeToLeft(resourceLogGroup.timestamp) }}>
						{labelText(resourceLogGroup)}
					</span>,
				);
			}
		});

		// One tooltip for the row, with the change at the mouse. It finds the last change at or
		// before the mouse's time.
		let shownIndex = -1;
		const tooltip = tippy(rowElem, {
			...followAlong,
			placement: 'bottom',
			content: '',
			onShow: () => (shownIndex >= 0 ? undefined : false),
		});
		const onMove = (event: MouseEvent) => {
			const time = this.leftToTime(event.clientX - rowElem.getBoundingClientRect().left);
			let lo = 0;
			let hi = resourceLogs.length - 1;
			while (lo < hi) {
				const mid = Math.ceil((lo + hi) / 2);
				if (resourceLogs[mid].timestamp <= time) lo = mid;
				else hi = mid - 1;
			}
			if (lo == shownIndex) return;
			shownIndex = lo;
			tooltip.setContent(this.resourceTooltipElem(resourceLogs[lo], maxValue, false));
			if (!tooltip.state.isVisible) tooltip.show();
		};
		rowElem.addEventListener('mousemove', onMove);
		this.addOnResetCallback(() => {
			rowElem.removeEventListener('mousemove', onMove);
			tooltip.destroy();
		});
		this.addRow(row);
	}

	// One row for the global cooldown. Each cast that starts the GCD draws a bar for as long as
	// the GCD lasts, in the color of its icon, so the row reads as the rotation. A cast that takes
	// longer than the GCD, like Lightning Bolt, fills the whole bar. Otherwise the cast time is
	// the solid part and the rest of the GCD the lighter part. The time when the GCD was ready
	// and we pressed nothing shows as a striped gap. The label says how much of the fight was
	// spent under the GCD.
	private addGcdRow(player: UnitMetrics, duration: number) {
		const casts = player.castLogs.filter(castLog => castLog.effectiveTime > 0).sort((a, b) => a.timestamp - b.timestamp);
		if (casts.length == 0) return;

		const rowElem = this.makeRowElem();
		const row: RotationRow = { label: document.createElement('div'), row: rowElem, name: 'GCD', actionId: null, category: 'casts', events: [] };

		// We count the idle time and the share under the GCD from the pull on, so the prepull
		// doesn't count. In a leveling session we only count the time we fight. Searching,
		// resting and the corpse run are no time to press anything.
		const fighting = this.session
			? this.session.spans.filter(span => span.state == 'fighting').map(span => [span.start, span.end] as [number, number])
			: [[0, duration] as [number, number]];
		const fightingTime = sum(fighting.map(([from, to]) => to - from));
		// How much of from to to is fighting time.
		const fightingOverlap = (from: number, to: number) => sum(fighting.map(([a, b]) => Math.max(0, Math.min(to, b) - Math.max(from, a))));
		let busyUntil = 0;
		let busyTime = 0;
		const addIdle = (from: number, to: number) => {
			if (this.session) {
				// Only the parts that fall in a fight are idle.
				fighting.forEach(([a, b]) => {
					if (Math.min(to, b) > Math.max(from, a)) drawIdle(Math.max(from, a), Math.min(to, b));
				});
			} else {
				drawIdle(from, to);
			}
		};
		const drawIdle = (from: number, to: number) => {
			// Gaps shorter than this are the sim's reaction time, not a missed GCD.
			if (to - from < 0.05) return;
			const idleElem = (<div className="rotation-gcd-idle" style={{ left: this.timeToLeft(from), width: this.timeToPx(to - from) }} />) as HTMLElement;
			rowElem.appendChild(idleElem);
			this.lazyTooltip(
				idleElem,
				() => (
					<div className="timeline-tooltip">
						<div className="timeline-tooltip-title">
							<span className="timeline-tooltip-name">Idle</span>
						</div>
						<div className="timeline-tooltip-meta">
							{`${formatSeconds(from)} to ${formatSeconds(to)} · ${formatSeconds(to - from)} with the GCD ready and nothing pressed`}
						</div>
					</div>
				),
				followAlong,
			);
		};

		casts.forEach(castLog => {
			const start = castLog.timestamp;
			const end = start + castLog.effectiveTime;
			if (start > busyUntil) addIdle(Math.max(0, busyUntil), start);
			busyTime += fightingOverlap(Math.max(start, busyUntil, 0), Math.min(end, duration));
			busyUntil = Math.max(busyUntil, end);

			const actionId = castLog.actionId!;
			const castPart = Math.min(castLog.castTime, castLog.effectiveTime);
			const gcdElem = (
				<div className="rotation-gcd" style={{ left: this.timeToLeft(start), width: this.timeToPx(castLog.effectiveTime) }}>
					{castPart > 0 && <div className="rotation-gcd-cast" style={{ width: this.timeToPx(castPart) }} />}
				</div>
			) as HTMLElement;
			const iconElem = (<span className="rotation-gcd-icon" />) as HTMLElement;
			actionId.setBackground(iconElem);
			gcdElem.appendChild(iconElem);
			this.tint(gcdElem, actionId.iconUrl);
			rowElem.appendChild(gcdElem);
			row.events.push({ elem: gcdElem, text: searchText([castLog]), actionId });

			const timing = [`${formatSeconds(start)} to ${formatSeconds(end)}`];
			if (castLog.castTime > 0) timing.push(`${formatSeconds(castLog.castTime)} cast`);
			timing.push(`${formatSeconds(castLog.effectiveTime)} GCD`);
			this.lazyTooltip(
				gcdElem,
				() => (
					<div className="timeline-tooltip">
						{tooltipTitle(actionId)}
						<div className="timeline-tooltip-meta">{timing.join(' · ')}</div>
					</div>
				),
				followAlong,
			);
		});
		if (busyUntil < duration) addIdle(Math.max(0, busyUntil), duration);

		const usage = fightingTime > 0 ? (busyTime / fightingTime) * 100 : 0;
		const labelElem = (
			<div className="rotation-label rotation-row rotation-gcd-label">
				<span className="rotation-label-icon rotation-gcd-label-icon">
					<i className="fas fa-hourglass-half" />
				</span>
				<span className="rotation-label-text">GCD</span>
				<span className="rotation-gcd-usage">{`${formatAmount(usage, 1)}% used`}</span>
			</div>
		) as HTMLElement;
		const labelTooltip = tippy(labelElem, {
			placement: 'right',
			content: this.session
				? 'The share of the time fighting that we spent under the global cooldown or casting. Searching, resting and the corpse run don\'t count. Striped gaps are time in a fight with the GCD ready and nothing pressed.'
				: 'The share of the fight, from the pull on, that we spent under the global cooldown or casting. Striped gaps are time with the GCD ready and nothing pressed.',
		});
		this.addOnResetCallback(() => labelTooltip.destroy());
		row.label = labelElem;
		this.addRow(row);
	}

	private addCastRow(castLogs: Array<CastLog>, aurasById: Array<Array<AuraUptimeLog>>) {
		const actionId = castLogs[0].actionId!;
		const rowElem = this.makeRowElem();
		const row: RotationRow = { label: this.makeLabelElem(actionId), row: rowElem, name: actionId.name, actionId, category: 'casts', events: [] };
		this.addRow(row);

		let stackedIconCount = 1;
		let stackedDamageCount = 1;

		castLogs.forEach((castLog, index) => {
			const castElem = (
				<div
					className="rotation-timeline-cast"
					style={{
						left: this.timeToLeft(castLog.timestamp),
						minWidth: this.timeToPx(castLog.castTime + travelTimeOf(castLog)),
					}}></div>
			);

			rowElem.appendChild(castElem);
			row.events.push({ elem: castElem, text: searchText([castLog, ...castLog.damageDealtLogs]) });

			const travelTime = travelTimeOf(castLog);
			if (travelTime != 0) {
				const travelTimeElem = (
					<div
						className="rotation-timeline-travel-time"
						style={{
							left: this.timeToPx(castLog.castTime),
							minWidth: this.timeToPx(travelTime),
						}}
					/>
				);
				castElem.appendChild(travelTimeElem);
			}

			this.tint(castElem as HTMLElement, actionId.iconUrl);

			const iconElem = (<a className="rotation-timeline-cast-icon" />) as HTMLAnchorElement;
			actionId.setBackground(iconElem);
			castElem.appendChild(iconElem);
			(row.icons ??= []).push({ time: castLog.timestamp + castLog.castTime + travelTime, elem: iconElem });

			const notable = notableOutcome(castLog);
			if (notable) {
				castElem.appendChild(<span className={`rotation-timeline-cast-badge outcome-${notable.kind}`}>{badgeLabels[notable.label] ?? notable.label}</span>);
			}

			const totalDamage = castLog.totalDamage();

			if (index > 0) {
				let timeDelta = null;

				for (let i = index - 1; i >= 0; i--) {
					if (castLog.timestamp != castLogs[i].timestamp) {
						timeDelta = castLog.timestamp - castLogs[i].timestamp;
						break;
					}
				}
				if (timeDelta != null && timeDelta < 0.21) {
					stackedDamageCount = stackedDamageCount + castLog.damageDealtLogs.length;
					stackedIconCount = stackedIconCount + 1;
				} else {
					stackedDamageCount = castLog.damageDealtLogs.length;
					stackedIconCount = 1;
				}
			}

			const startIndex = Math.max(0, 1 + index - stackedIconCount);
			const relevantCastLogs = castLogs.slice(startIndex, index + 1);

			// Reset and initialize aggregatedData for the current iteration
			const aggregatedData = relevantCastLogs.reduce<{
				damageDealtLogs: DamageDealtLog[];
				totalDamage: number;
			}>(
				(acc, log) => {
					// Correct type
					acc.damageDealtLogs.push(...log.damageDealtLogs);
					acc.totalDamage += log.totalDamage();
					return acc;
				},
				{
					damageDealtLogs: [], // Correctly typed as DamageDealtLog[]
					totalDamage: 0,
				},
			);

			const timing = [formatSeconds(castLog.timestamp)];
			if (castLog.castTime > 0) {
				timing[0] += ` to ${formatSeconds(castLog.timestamp + castLog.castTime)}`;
				timing.push(`${formatSeconds(castLog.castTime)} cast`);
			}
			if (castLog.effectiveTime > 0) timing.push(`${formatSeconds(castLog.effectiveTime)} GCD`);
			if (travelTime > 0) timing.push(`${formatSeconds(travelTime)} travel`);

			const tt = () => (
				<div className="timeline-tooltip">
					{tooltipTitle(castLog.actionId!)}
					<div className="timeline-tooltip-meta">{timing.join(' · ')}</div>
					{aggregatedData.damageDealtLogs.length > 0 && <ul className="timeline-tooltip-hits">{aggregatedData.damageDealtLogs.map(ddl => tooltipHit(ddl))}</ul>}
					{notable?.mixed && (
						<div className="timeline-tooltip-meta">The icon's badge shows the most notable result, {notable.label.toLowerCase()}.</div>
					)}
					{totalDamage > 0 && (aggregatedData.damageDealtLogs.length > 1 || castLog.effectiveTime > 0) && (
						<div className="timeline-tooltip-rows">
							{/* A single hit already shows its amount, so the total only shows for two or more. */}
							{aggregatedData.damageDealtLogs.length > 1 && tooltipRow('Total', formatAmount(aggregatedData.totalDamage))}
							{castLog.effectiveTime > 0 && tooltipRow('DPET', formatAmount(aggregatedData.totalDamage / castLog.effectiveTime, 1))}
						</div>
					)}
				</div>
			);

			this.lazyTooltip(castElem as HTMLElement, tt, followAlong);

			if (stackedIconCount > 1) {
				castElem.appendChild(<div className="stacked-icon-count">{String(stackedDamageCount)}</div>);
			}

			castLog.damageDealtLogs
				.filter(ddl => ddl.tick)
				.forEach(ddl => {
					const tickElem = (
						<div
							className="rotation-timeline-tick"
							style={{
								left: this.timeToLeft(ddl.timestamp),
							}}
						/>
					);
					rowElem.appendChild(tickElem);
					row.events.push({ elem: tickElem, text: searchText([ddl]) });

					this.lazyTooltip(tickElem as HTMLElement, () => (
						<div className="timeline-tooltip">
							{tooltipTitle(ddl.actionId!, 'Tick')}
							<ul className="timeline-tooltip-hits">{tooltipHit(ddl)}</ul>
						</div>
					));
				});
		});

		// If there are any auras that correspond to this cast, visualize them in the same row.
		aurasById
			.filter(auraUptimeLogs => auraUptimeLogs[0].actionId!.equalsIgnoringTag(actionId))
			.forEach(auraUptimeLogs => this.applyAuraUptimeLogsToRow(auraUptimeLogs, row));
	}

	private addAuraRow(auraUptimeLogs: Array<AuraUptimeLog>, duration: number) {
		const actionId = auraUptimeLogs[0].actionId!;

		const row: RotationRow = {
			label: this.makeLabelElem(actionId, true),
			row: this.makeRowElem(),
			name: actionId.name,
			actionId,
			category: 'auras',
			events: [],
			alwaysUp: isAlwaysUp(auraUptimeLogs, duration),
		};
		this.addRow(row);

		this.applyAuraUptimeLogsToRow(auraUptimeLogs, row, true);
	}

	// Draws a bar for each time the aura is up. In the aura's own lane, the bar starts with the
	// aura's icon, like a cast. A cast lane shows its ability's auras without one, because the
	// cast that put the aura up already has it.
	private applyAuraUptimeLogsToRow(auraUptimeLogs: Array<AuraUptimeLog>, row: RotationRow, withIcon = false) {
		auraUptimeLogs.forEach(aul => {
			const auraElem = (
				<div
					className="rotation-timeline-aura"
					style={{
						left: this.timeToLeft(aul.gainedAt),
						minWidth: this.timeToPx(aul.fadedAt === aul.gainedAt ? 0.001 : aul.fadedAt - aul.gainedAt),
					}}
				/>
			);
			row.row.appendChild(auraElem);
			this.tint(auraElem as HTMLElement, aul.actionId?.iconUrl);
			if (withIcon && aul.actionId) {
				const iconElem = (<a className="rotation-timeline-cast-icon rotation-timeline-aura-icon" />) as HTMLAnchorElement;
				aul.actionId.setBackground(iconElem);
				auraElem.appendChild(iconElem);
				(row.icons ??= []).push({ time: aul.gainedAt, elem: iconElem });
			}
			row.events.push({ elem: auraElem, text: searchText([aul, ...aul.stacksChange]) });

			this.lazyTooltip(
				auraElem as HTMLElement,
				() => (
					<div className="timeline-tooltip">
						{tooltipTitle(aul.actionId!)}
						<div className="timeline-tooltip-meta">
							{`${formatSeconds(aul.gainedAt)} to ${formatSeconds(aul.fadedAt)} · ${formatSeconds(aul.fadedAt - aul.gainedAt)}`}
						</div>
					</div>
				),
				followAlong,
			);

			aul.stacksChange.forEach((scl, i) => {
				if (scl.timestamp == aul.fadedAt) {
					return;
				}

				// When an icon in the row is at the same moment, like the Sunder Armor cast that added
				// the stack, the count goes in the icon's bottom right corner, like on a buff in the
				// game. Otherwise, like when a Flurry charge gets used up, it goes in the bar.
				const icon = row.icons?.find(icon => Math.abs(icon.time - scl.timestamp) < 0.05);
				if (icon) {
					const countElem = icon.elem.querySelector('.rotation-icon-count') ?? icon.elem.appendChild(<span className="rotation-icon-count" />);
					countElem.textContent = String(scl.newStacks);
				}
				const stacksChangeElem = (
					<div
						className="rotation-timeline-stacks-change"
						style={{
							left: this.timeToPx(scl.timestamp - aul.timestamp),
							width: this.timeToPx(aul.stacksChange[i + 1] ? aul.stacksChange[i + 1].timestamp - scl.timestamp : aul.fadedAt - scl.timestamp),
						}}>
						{icon ? '' : String(scl.newStacks)}
					</div>
				);
				auraElem.appendChild(stacksChangeElem);
			});
		});
	}

	// Gives a bar the main color of its ability's icon, see icon_color.ts. The icon loads once,
	// and until then the bar has the default color.
	private tint(elem: HTMLElement, iconUrl: string | undefined) {
		if (!iconUrl) return;
		iconColor(iconUrl).then(rgb => {
			if (rgb) elem.style.setProperty('--bar-rgb', rgb);
		});
	}

	// The CSS for a length, like how long a cast or an aura lasts. It follows --px-per-second, so
	// it changes with the zoom without drawing the row again.
	private timeToPx(time: number): string {
		return `calc(var(--px-per-second) * ${time})`;
	}

	// Where the left edge of the scrolled content is on the screen, in CSS pixels.
	private measureContentOrigin() {
		const rect = this.rotationTimeline.getBoundingClientRect();
		this.contentOrigin = rect.left + this.rotationTimeline.clientLeft - this.rotationTimeline.scrollLeft;
		// A line one screen pixel wide. With Windows display scaling at 125%, a CSS pixel is 1.25
		// screen pixels, so a 1px line would always blur over two of them.
		const dpr = window.devicePixelRatio || 1;
		this.rotationTimeline.style.setProperty('--hairline', `${Math.max(1, Math.round(dpr)) / dpr}px`);
	}

	// Like timeToLeft, for the thin lines of the ruler and the grid. Most seconds land between two
	// screen pixels, like 3 sec at 88.5 px per second, at x = 265.5. The browser then blends a
	// 1px line over both, and it looks wider and brighter than one that lands on a pixel. So we
	// move each line to the nearest screen pixel.
	private crispLeft(time: number): string {
		const dpr = window.devicePixelRatio || 1;
		const exact = this.contentOrigin + this.timeToLeftValue(time);
		const shift = Math.round(exact * dpr) / dpr - exact;
		return `calc(${this.timeToLeft(time)} + ${shift.toFixed(3)}px)`;
	}

	// For where something starts, measured from the start of the time range we draw.
	private timeToLeftValue(time: number): number {
		return (time - this.viewStart) * this.pxPerSecond;
	}
	private timeToLeft(time: number): string {
		return this.timeToPx(time - this.viewStart);
	}
	private leftToTime(left: number): number {
		return this.viewStart + left / this.pxPerSecond;
	}

	private viewWidthPx(): string {
		return this.timeToPx(this.viewEnd - this.viewStart);
	}

	// The time ruler above the rotation. We pick the spacing of the labels from the zoom, so they
	// stay about 60px or more apart.
	// With a prepull, the time before the pull gets a hatched background labeled Prepull, and a
	// gold line marks the pull through every row. The background goes first, so the rows draw over
	// it, and the line goes last, over them.
	private drawPull() {
		if (this.prepullStart >= 0 || this.viewStart >= 0) return;
		const pullLeft = this.timeToLeft(Math.min(0, this.viewEnd));
		this.rotationTimeline.prepend(
			<div className="rotation-prepull" style={{ width: pullLeft }}>
				<span className="rotation-prepull-label">Prepull</span>
			</div>,
		);
		if (this.viewEnd > 0) this.rotationTimeline.appendChild(<div className="rotation-pull-line" style={{ left: pullLeft }} />);
	}

	// A faint line down the rows at every second, so we can follow a time from the ruler. Each line
	// sits where the ruler's tick for that second is, and starts under the ruler. It goes first, so
	// the rows draw over it.
	private drawGrid(header: HTMLElement) {
		const lines: Array<Element> = [];
		for (let time = Math.ceil(this.viewStart); time <= this.viewEnd; time++) {
			lines.push(<div className="rotation-grid-line" style={{ left: this.crispLeft(time) }} />);
		}
		this.rotationTimeline.prepend(
			<div className="rotation-grid" style={{ top: `${header.offsetTop + header.offsetHeight}px`, width: this.viewWidthPx() }}>
				{lines}
			</div>,
		);
	}

	private drawRotationTimeRuler(header: HTMLElement) {
		const [labelStep, tickStep] = rulerSteps.find(([label]) => label * this.pxPerSecond >= 60) ?? rulerSteps[rulerSteps.length - 1];
		header.style.width = this.viewWidthPx();
		this.measureContentOrigin();

		const ticks: Array<Element> = [];
		const firstTick = Math.ceil(this.viewStart / tickStep - 1e-6);
		const lastTick = Math.floor(this.viewEnd / tickStep + 1e-6);
		for (let i = firstTick; i <= lastTick; i++) {
			const time = i * tickStep;
			const isLabel = Math.abs(time / labelStep - Math.round(time / labelStep)) < 1e-6;
			ticks.push(<div className={`rotation-ruler-tick${isLabel ? ' major' : ''}`} style={{ left: this.crispLeft(time) }} />);
			if (isLabel) {
				ticks.push(
					<div className={`rotation-ruler-label${time == 0 && this.prepullStart < 0 ? ' pull' : ''}`} style={{ left: this.timeToLeft(time) }}>
						{time == 0 && this.prepullStart < 0 ? 'Pull' : formatRulerTime(time)}
					</div>,
				);
			}
		}
		header.replaceChildren(...ticks);
	}

	// The chart's tooltips. The time is the title, then the events at that moment, then the
	// value of the line.
	private chartTooltipTitle(timestamp: number, player: UnitMetrics, colorOverride: string) {
		return (
			<div className="timeline-tooltip-title">
				{colorOverride != '' && (
					<>
						<img className="timeline-tooltip-icon" src={player.iconUrl} />
						<span className="timeline-tooltip-name" style={{ color: colorOverride }}>
							{player.label}
						</span>
					</>
				)}
				<span className="timeline-tooltip-name">{formatSeconds(timestamp)}</span>
			</div>
		);
	}

	// The hits are those at this time, none at the moment an old hit leaves the window. The
	// auras are those at the last hits at or before it.
	private dpsTooltip(timestamp: number, dps: number, log: DpsLog | null, aurasLog: DpsLog | null, player: UnitMetrics, colorOverride: string) {
		return (
			<div className="timeline-tooltip dps">
				{this.chartTooltipTitle(timestamp, player, colorOverride)}
				{log && log.damageLogs.length > 0 && <ul className="timeline-tooltip-hits">{log.damageLogs.map(damageLog => tooltipHit(damageLog, true))}</ul>}
				<div className="timeline-tooltip-rows">{tooltipRow('DPS', formatAmount(dps, 1), 'series-color')}</div>
				{aurasLog && this.tooltipAurasSection(aurasLog)}
			</div>
		);
	}

	private threatTooltip(log: ThreatLogGroup, includeAuras: boolean, player: UnitMetrics, colorOverride: string) {
		return (
			<div className="timeline-tooltip threat">
				{this.chartTooltipTitle(log.timestamp, player, colorOverride)}
				<div className="timeline-tooltip-rows">{tooltipRow('Before', formatAmount(log.threatBefore, 1))}</div>
				<ul className="timeline-tooltip-changes">
					{log.logs.map(threatLog => tooltipChange(threatLog, `${formatAmount(threatLog.threat, 1)} threat`))}
				</ul>
				<div className="timeline-tooltip-rows">{tooltipRow('After', formatAmount(log.threatAfter, 1), 'series-color')}</div>
				{includeAuras ? this.tooltipAurasSection(log) : null}
			</div>
		);
	}

	private resourceTooltipElem(log: ResourceChangedLogGroup, maxValue: number, includeAuras: boolean) {
		const valToDisplayString = percentageResources.includes(log.resourceType)
			? (val: number) => `${formatAmount(val)} (${((val / maxValue) * 100).toFixed(0)}%)`
			: (val: number) => formatAmount(val, 1);
		const name = resourceNames.get(log.resourceType)!;

		return (
			<div className={`timeline-tooltip ${name.toLowerCase().replaceAll(' ', '-')}`}>
				<div className="timeline-tooltip-title">
					<img className="timeline-tooltip-icon" src={resourceTypeToIcon[log.resourceType]} />
					<span className="timeline-tooltip-name">{name}</span>
					<span className="timeline-tooltip-note">{formatSeconds(log.timestamp)}</span>
				</div>
				<div className="timeline-tooltip-rows">{tooltipRow('Before', valToDisplayString(log.valueBefore))}</div>
				<ul className="timeline-tooltip-changes">
					{log.logs.map(changedLog => {
						const delta = changedLog.valueAfter - changedLog.valueBefore;
						return tooltipChange(changedLog, `${delta < 0 ? '' : '+'}${formatAmount(delta, 1)}`, delta < 0 ? 'negative' : 'positive');
					})}
				</ul>
				<div className="timeline-tooltip-rows">{tooltipRow('After', valToDisplayString(log.valueAfter), 'series-color')}</div>
				{includeAuras && this.tooltipAurasSection(log)}
			</div>
		);
	}

	private resourceTooltip(log: ResourceChangedLogGroup, maxValue: number, includeAuras: boolean) {
		return this.resourceTooltipElem(log, maxValue, includeAuras);
	}

	// The auras up at the moment, as a row of small icons with their names.
	private tooltipAurasSection(log: SimLog): JSX.Element | null {
		if (log.activeAuras.length == 0) return null;
		return (
			<div className="timeline-tooltip-auras">
				<div className="timeline-tooltip-section">Active auras</div>
				<ul className="timeline-tooltip-aura-list">
					{log.activeAuras.map(auraLog => (
						<li>
							{auraLog.actionId!.iconUrl && <img className="timeline-tooltip-icon small" src={auraLog.actionId!.iconUrl} />}
							<span>{auraLog.actionId!.name}</span>
						</li>
					))}
				</ul>
			</div>
		);
	}

	addOnResetCallback(callback: () => void) {
		this.resetCallbacks.push(callback);
	}

	reset() {
		this.resetCallbacks.forEach(callback => callback());
		this.resetCallbacks = [];
	}
}

// Whether the aura is up from the pull to the end of the fight. We add up its uptime between
// those two, so an aura we put up in the prepull counts too. When we refresh an aura, it fades and
// comes back at the same moment, so that leaves no gap. We allow 10 ms for rounding.
const isAlwaysUp = (auraUptimeLogs: Array<AuraUptimeLog>, duration: number): boolean => {
	const spans = auraUptimeLogs
		.map(log => [Math.max(0, log.gainedAt), Math.min(duration, log.fadedAt)])
		.filter(([from, to]) => to > from)
		.sort((a, b) => a[0] - b[0]);
	let covered = 0;
	let end = 0;
	spans.forEach(([from, to]) => {
		if (to <= end) return;
		covered += to - Math.max(from, end);
		end = to;
	});
	return covered >= duration - 0.01;
};

const MELEE_ACTION_CATEGORY = 1;
const SPELL_ACTION_CATEGORY = 2;
const DEFAULT_ACTION_CATEGORY = 3;

// Hard-coded spell categories for controlling rotation ordering.
const idToCategoryMap: Record<number, number> = {
	[OtherAction.OtherActionShoot]: 0,
	[OtherAction.OtherActionMove]: 0.33,
	[OtherAction.OtherActionAttack]: 0.66,

	// Druid
	[48480]: 0.1, // Maul
	[48564]: MELEE_ACTION_CATEGORY + 0.1, // Mangle (Bear)
	[48568]: MELEE_ACTION_CATEGORY + 0.2, // Lacerate
	[48562]: MELEE_ACTION_CATEGORY + 0.3, // Swipe (Bear)

	[48566]: MELEE_ACTION_CATEGORY + 0.1, // Mangle (Cat)
	[48572]: MELEE_ACTION_CATEGORY + 0.2, // Shred
	[49800]: MELEE_ACTION_CATEGORY + 0.51, // Rip
	[52610]: MELEE_ACTION_CATEGORY + 0.52, // Savage Roar
	[48577]: MELEE_ACTION_CATEGORY + 0.53, // Ferocious Bite

	[48465]: SPELL_ACTION_CATEGORY + 0.1, // Starfire
	[48461]: SPELL_ACTION_CATEGORY + 0.2, // Wrath
	[439748]: SPELL_ACTION_CATEGORY + 0.3, // Starfall
	[48468]: SPELL_ACTION_CATEGORY + 0.4, // Insect Swarm
	[48463]: SPELL_ACTION_CATEGORY + 0.5, // Moonfire

	// Hunter
	[48996]: 0.1, // Raptor Strike
	[53217]: 0.6, // Wild Quiver
	[53209]: MELEE_ACTION_CATEGORY + 0.1, // Chimera Shot
	[53353]: MELEE_ACTION_CATEGORY + 0.11, // Chimera Shot Serpent
	[60053]: MELEE_ACTION_CATEGORY + 0.1, // Explosive Shot
	[49050]: MELEE_ACTION_CATEGORY + 0.2, // Aimed Shot
	[49048]: MELEE_ACTION_CATEGORY + 0.21, // Multi Shot
	[49045]: MELEE_ACTION_CATEGORY + 0.22, // Arcane Shot
	[49052]: MELEE_ACTION_CATEGORY + 0.27, // Steady Shot
	[61006]: MELEE_ACTION_CATEGORY + 0.28, // Kill Shot
	[34490]: MELEE_ACTION_CATEGORY + 0.29, // Silencing Shot
	[49001]: MELEE_ACTION_CATEGORY + 0.3, // Serpent Sting
	[3043]: MELEE_ACTION_CATEGORY + 0.3, // Scorpid Sting
	[53238]: MELEE_ACTION_CATEGORY + 0.31, // Piercing Shots
	[63672]: MELEE_ACTION_CATEGORY + 0.32, // Black Arrow
	[49067]: MELEE_ACTION_CATEGORY + 0.33, // Explosive Trap
	[58434]: MELEE_ACTION_CATEGORY + 0.34, // Volley

	// Paladin
	[35395]: MELEE_ACTION_CATEGORY + 0.1, // Crusader Strike
	[53385]: MELEE_ACTION_CATEGORY + 0.2, // Divine Storm
	[42463]: MELEE_ACTION_CATEGORY + 0.3, // Seal of Vengeance
	[61840]: MELEE_ACTION_CATEGORY + 0.4, // Righteous Vengeance
	[61411]: MELEE_ACTION_CATEGORY + 0.5, // Shield of Righteousness
	[53595]: MELEE_ACTION_CATEGORY + 0.51, // Hammer of Righteousness
	[20182]: MELEE_ACTION_CATEGORY + 0.6, // Reckoning
	[48952]: SPELL_ACTION_CATEGORY + 0.1, // Holy Shield
	[31803]: SPELL_ACTION_CATEGORY + 0.2, // Holy Vengeance
	[48801]: SPELL_ACTION_CATEGORY + 0.3, // Exorcism
	[48819]: SPELL_ACTION_CATEGORY + 0.4, // Consecration
	[53408]: SPELL_ACTION_CATEGORY + 0.51, // Judgement of Wisdom
	[20271]: SPELL_ACTION_CATEGORY + 0.52, // Judgement of Light
	[31804]: SPELL_ACTION_CATEGORY + 0.53, // Judgement of Vengeance
	[20467]: SPELL_ACTION_CATEGORY + 0.54, // Judgement of Command
	[20187]: SPELL_ACTION_CATEGORY + 0.55, // Judgement of Righteousness
	[31878]: SPELL_ACTION_CATEGORY + 0.56, // Judgements of the Wise
	[48817]: SPELL_ACTION_CATEGORY + 0.5, // Holy Wrath
	[48806]: SPELL_ACTION_CATEGORY + 0.6, // Hammer of Wrath
	[54428]: SPELL_ACTION_CATEGORY + 0.7, // Divine Plea
	[66233]: SPELL_ACTION_CATEGORY + 0.8, // Ardent Defender

	// Priest
	[48300]: SPELL_ACTION_CATEGORY + 0.11, // Devouring Plague
	[48125]: SPELL_ACTION_CATEGORY + 0.12, // Shadow Word: Pain
	[48160]: SPELL_ACTION_CATEGORY + 0.13, // Vampiric Touch
	[48135]: SPELL_ACTION_CATEGORY + 0.14, // Holy Fire
	[48123]: SPELL_ACTION_CATEGORY + 0.19, // Smite
	[48127]: SPELL_ACTION_CATEGORY + 0.2, // Mind Blast
	[48158]: SPELL_ACTION_CATEGORY + 0.3, // Shadow Word: Death
	[48156]: SPELL_ACTION_CATEGORY + 0.4, // Mind Flay

	// Rogue
	[6774]: MELEE_ACTION_CATEGORY + 0.1, // Slice and Dice
	[8647]: MELEE_ACTION_CATEGORY + 0.2, // Expose Armor
	[48672]: MELEE_ACTION_CATEGORY + 0.3, // Rupture
	[57993]: MELEE_ACTION_CATEGORY + 0.3, // Envenom
	[48668]: MELEE_ACTION_CATEGORY + 0.4, // Eviscerate
	[48666]: MELEE_ACTION_CATEGORY + 0.5, // Mutilate
	[48665]: MELEE_ACTION_CATEGORY + 0.6, // Mutilate (MH)
	[48664]: MELEE_ACTION_CATEGORY + 0.7, // Mutilate (OH)
	[48638]: MELEE_ACTION_CATEGORY + 0.5, // Sinister Strike
	[51723]: MELEE_ACTION_CATEGORY + 0.8, // Fan of Knives
	[57973]: SPELL_ACTION_CATEGORY + 0.1, // Deadly Poison
	[57968]: SPELL_ACTION_CATEGORY + 0.2, // Instant Poison

	// Shaman
	[58804]: 0.11, // Windfury Weapon
	[58790]: 0.12, // Flametongue Weapon
	[58796]: 0.12, // Frostbrand Weapon
	[17364]: MELEE_ACTION_CATEGORY + 0.1, // Stormstrike
	[60103]: MELEE_ACTION_CATEGORY + 0.2, // Lava Lash
	[49233]: SPELL_ACTION_CATEGORY + 0.21, // Flame Shock
	[49231]: SPELL_ACTION_CATEGORY + 0.22, // Earth Shock
	[49236]: SPELL_ACTION_CATEGORY + 0.23, // Frost Shock
	[60043]: SPELL_ACTION_CATEGORY + 0.31, // Lava Burst
	[49238]: SPELL_ACTION_CATEGORY + 0.32, // Lightning Bolt
	[49271]: SPELL_ACTION_CATEGORY + 0.33, // Chain Lightning
	[61657]: SPELL_ACTION_CATEGORY + 0.41, // Fire Nova
	[58734]: SPELL_ACTION_CATEGORY + 0.42, // Magma Totem
	[58704]: SPELL_ACTION_CATEGORY + 0.43, // Searing Totem
	[49281]: SPELL_ACTION_CATEGORY + 0.51, // Lightning Shield
	[49279]: SPELL_ACTION_CATEGORY + 0.52, // Lightning Shield (Proc)
	[2825]: DEFAULT_ACTION_CATEGORY + 0.1, // Bloodlust

	// Warlock
	[47867]: SPELL_ACTION_CATEGORY + 0.01, // Curse of Doom
	[47864]: SPELL_ACTION_CATEGORY + 0.02, // Curse of Agony
	[47813]: SPELL_ACTION_CATEGORY + 0.1, // Corruption
	[59164]: SPELL_ACTION_CATEGORY + 0.2, // Haunt
	[47843]: SPELL_ACTION_CATEGORY + 0.3, // Unstable Affliction
	[47811]: SPELL_ACTION_CATEGORY + 0.31, // Immolate
	[17962]: SPELL_ACTION_CATEGORY + 0.32, // Conflagrate
	[59172]: SPELL_ACTION_CATEGORY + 0.49, // Chaos Bolt
	[47809]: SPELL_ACTION_CATEGORY + 0.5, // Shadow Bolt
	[47838]: SPELL_ACTION_CATEGORY + 0.51, // Incinerate
	[47825]: SPELL_ACTION_CATEGORY + 0.52, // Soul Fire
	[47855]: SPELL_ACTION_CATEGORY + 0.6, // Drain Soul
	[57946]: SPELL_ACTION_CATEGORY + 0.7, // Life Tap
	[47241]: SPELL_ACTION_CATEGORY + 0.8, // Metamorphosis
	[50589]: SPELL_ACTION_CATEGORY + 0.81, // Immolation Aura
	[47193]: SPELL_ACTION_CATEGORY + 0.82, // Demonic Empowerment

	// Mage
	[42842]: SPELL_ACTION_CATEGORY + 0.01, // Frostbolt
	[47610]: SPELL_ACTION_CATEGORY + 0.02, // Frostfire Bolt
	[42897]: SPELL_ACTION_CATEGORY + 0.02, // Arcane Blast
	[42833]: SPELL_ACTION_CATEGORY + 0.02, // Fireball
	[42859]: SPELL_ACTION_CATEGORY + 0.03, // Scorch
	[42891]: SPELL_ACTION_CATEGORY + 0.1, // Pyroblast
	[10212]: SPELL_ACTION_CATEGORY + 0.1, // Arcane Missiles
	[44572]: SPELL_ACTION_CATEGORY + 0.1, // Deep Freeze
	[44781]: SPELL_ACTION_CATEGORY + 0.2, // Arcane Barrage
	[42914]: SPELL_ACTION_CATEGORY + 0.2, // Ice Lance
	[55360]: SPELL_ACTION_CATEGORY + 0.2, // Living Bomb
	[55362]: SPELL_ACTION_CATEGORY + 0.21, // Living Bomb (Explosion)
	[12654]: SPELL_ACTION_CATEGORY + 0.3, // Ignite
	[12472]: SPELL_ACTION_CATEGORY + 0.4, // Icy Veins
	[11129]: SPELL_ACTION_CATEGORY + 0.4, // Combustion
	[12042]: SPELL_ACTION_CATEGORY + 0.4, // Arcane Power
	[11958]: SPELL_ACTION_CATEGORY + 0.41, // Cold Snap
	[12043]: SPELL_ACTION_CATEGORY + 0.41, // Presence of Mind
	[31687]: SPELL_ACTION_CATEGORY + 0.41, // Water Elemental
	[55342]: SPELL_ACTION_CATEGORY + 0.5, // Mirror Image
	[33312]: SPELL_ACTION_CATEGORY + 0.51, // Mana Gems
	[12051]: SPELL_ACTION_CATEGORY + 0.52, // Evocate
	[44401]: SPELL_ACTION_CATEGORY + 0.6, // Missile Barrage
	[44448]: SPELL_ACTION_CATEGORY + 0.6, // Hot Streak
	[44545]: SPELL_ACTION_CATEGORY + 0.6, // Fingers of Frost
	[44549]: SPELL_ACTION_CATEGORY + 0.61, // Brain Freeze
	[12536]: SPELL_ACTION_CATEGORY + 0.61, // Clearcasting

	// Warrior
	[845]: 0.1, // Cleave
	[11608]: 0.1, // Cleave
	[11609]: 0.1, // Cleave
	[20569]: 0.1, // Cleave
	[1608]: 0.1, // Heroic Strike
	[11565]: 0.1, // Heroic Strike
	[11566]: 0.1, // Heroic Strike
	[11567]: 0.1, // Heroic Strike
	[25286]: 0.1, // Heroic Strike
	[8820]: MELEE_ACTION_CATEGORY + 0.05, // Slam
	[11604]: MELEE_ACTION_CATEGORY + 0.05, // Slam
	[11605]: MELEE_ACTION_CATEGORY + 0.05, // Slam
	[23881]: MELEE_ACTION_CATEGORY + 0.1, // Bloodthirst
	[47486]: MELEE_ACTION_CATEGORY + 0.1, // Mortal Strike
	[30356]: MELEE_ACTION_CATEGORY + 0.1, // Shield Slam
	[47498]: MELEE_ACTION_CATEGORY + 0.21, // Devastate
	[47467]: MELEE_ACTION_CATEGORY + 0.22, // Sunder Armor
	[57823]: MELEE_ACTION_CATEGORY + 0.23, // Revenge
	[1680]: MELEE_ACTION_CATEGORY + 0.24, // Whirlwind
	[7384]: MELEE_ACTION_CATEGORY + 0.25, // Overpower
	[47471]: MELEE_ACTION_CATEGORY + 0.42, // Execute
	[12867]: SPELL_ACTION_CATEGORY + 0.51, // Deep Wounds
	[58874]: SPELL_ACTION_CATEGORY + 0.52, // Damage Shield
	[46924]: SPELL_ACTION_CATEGORY + 0.61, // Bladestorm
	[2565]: SPELL_ACTION_CATEGORY + 0.62, // Shield Block
	[64382]: SPELL_ACTION_CATEGORY + 0.65, // Shattering Throw
	[71]: DEFAULT_ACTION_CATEGORY + 0.1, // Defensive Stance
	[2457]: DEFAULT_ACTION_CATEGORY + 0.1, // Battle Stance
	[2458]: DEFAULT_ACTION_CATEGORY + 0.1, // Berserker Stance

	// Generic
	[53307]: SPELL_ACTION_CATEGORY + 0.931, // Thorns
	[54043]: SPELL_ACTION_CATEGORY + 0.932, // Retribution Aura
	[54758]: SPELL_ACTION_CATEGORY + 0.933, // Hyperspeed Acceleration
	[42641]: SPELL_ACTION_CATEGORY + 0.941, // Sapper
	[40536]: SPELL_ACTION_CATEGORY + 0.942, // Explosive Decoy
	[41119]: SPELL_ACTION_CATEGORY + 0.943, // Saronite Bomb
	[40771]: SPELL_ACTION_CATEGORY + 0.944, // Cobalt Frag Bomb
};

const idsToGroupForRotation: Array<number> = [
	6774, // Slice and Dice
	8647, // Expose Armor
	48668, // Eviscerate
	48672, // Rupture
	51690, // Killing Spree
	57993, // Envenom
];

const percentageResources: Array<ResourceType> = [ResourceType.ResourceTypeHealth, ResourceType.ResourceTypeMana];

// Pairs of seconds between ruler labels and seconds between its small ticks, from fine to coarse.
const rulerSteps: Array<[number, number]> = [
	[0.5, 0.1],
	[1, 0.2],
	[2, 0.5],
	[5, 1],
	[10, 2],
	[15, 5],
	[30, 5],
	[60, 10],
	[120, 30],
];

const formatRulerTime = (time: number): string => {
	if (time < 0) return `-${formatRulerTime(-time)}`;
	if (time < 60) return `${Number(time.toFixed(1))}s`;
	const minutes = Math.floor(time / 60);
	const seconds = Math.round(time - minutes * 60);
	return `${minutes}:${String(seconds).padStart(2, '0')}`;
};

