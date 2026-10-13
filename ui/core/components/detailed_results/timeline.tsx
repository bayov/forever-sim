import ApexCharts from 'apexcharts';
import tippy, { followCursor, Props } from 'tippy.js';
import { ref } from 'tsx-vanilla';

import { ResourceType } from '../../proto/api.js';
import { OtherAction } from '../../proto/common.js';
import { ActionId, resourceTypeToIcon } from '../../proto_utils/action_id.js';
import { AuraUptimeLog, CastLog, DamageDealtLog, DpsLog, ResourceChangedLogGroup, SimLog, ThreatLogGroup } from '../../proto_utils/logs_parser.js';
import { resourceNames } from '../../proto_utils/names.js';
import { UnitMetrics } from '../../proto_utils/sim_result.js';
import { orderedResourceTypes } from '../../proto_utils/utils.js';
import { bucket, distinct, formatToCompactNumber, formatToNumber, maxIndex, stringComparator } from '../../utils.js';
import { actionColors } from './color_settings.js';
import { EventCategory, EventFilter } from './events_filter.js';
import { hitOutcome, HitOutcomeKind, outcomeBadge } from './hit_outcome.js';
import { iconColor } from './icon_color.js';
import { ResultComponent, ResultComponentConfig, SimResultData } from './result_component.js';

type TooltipHandler = (dataPointIndex: number) => Element;

const dpsColor = '#ed5653';
const manaColor = '#2E93fA';
const threatColor = '#b56d07';

// How the timeline shows the iteration. The rotation has a row per ability, aura and resource.
// The chart draws DPS, mana and threat over time.
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
const searchText = (logs: Array<SimLog>): string =>
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

const tooltipTitle = (actionId: ActionId, note?: string): JSX.Element => (
	<div className="timeline-tooltip-title">
		{actionId.iconUrl && <img className="timeline-tooltip-icon" src={actionId.iconUrl} />}
		<span className="timeline-tooltip-name">{actionId.name}</span>
		{note && <span className="timeline-tooltip-note">{note}</span>}
	</div>
);

const tooltipRow = (key: string, value: string, valueClass = ''): JSX.Element => (
	<div className="timeline-tooltip-row">
		<span className="timeline-tooltip-key">{key}</span>
		<span className={`timeline-tooltip-value ${valueClass}`}>{value}</span>
	</div>
);

// One hit: its time, how it landed and the amount. The chart lists hits of several abilities,
// so there we name the ability too.
const tooltipHit = (ddl: DamageDealtLog, withName = false): JSX.Element => {
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

// Pixels per second of the rotation at each zoom step.
const zoomLevels = [10, 20, 40, 70, 100, 150, 250, 400];
const defaultZoomIndex = zoomLevels.indexOf(100);

export class Timeline extends ResultComponent {
	private readonly filter: EventFilter;

	private readonly dpsResourcesPlotElem: HTMLElement;
	private dpsResourcesPlot: any;

	private readonly rotationPlotElem: HTMLElement;
	private readonly rotationLabels: HTMLElement;
	private readonly rotationTimeline: HTMLElement;
	private readonly emptyElem: HTMLElement;

	private resultData: SimResultData | null;
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
	private zoomIndex = defaultZoomIndex;

	private resetCallbacks: (() => void)[] = [];

	constructor(config: TimelineConfig) {
		config.rootCssClass = 'timeline-root';
		super(config);
		this.filter = config.filter;
		this.resultData = null;
		this.rendered = false;

		this.rootElem.appendChild(
			<div className="timeline-plots-container">
				<div className="timeline-plot dps-resources-plot hide"></div>
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
		this.dpsResourcesPlot = new ApexCharts(this.dpsResourcesPlotElem, {
			chart: {
				animations: {
					enabled: false,
				},
				background: 'transparent',
				foreColor: '#8e948f',
				fontFamily: 'inherit',
				height: '100%',
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
		this.addOnDisposeCallback(() => this.dpsResourcesPlot.destroy());

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

		let lastRange = [this.filter.from, this.filter.to];
		this.filter.changeEmitter.on(() => {
			// A new time range moves everything, so we draw the timeline again. The other filters
			// only show and hide rows.
			const rangeChanged = lastRange[0] != this.filter.from || lastRange[1] != this.filter.to;
			lastRange = [this.filter.from, this.filter.to];
			if (rangeChanged) {
				this.invalidate();
			} else {
				this.applyFilter();
			}
		});
	}

	onSimResult(resultData: SimResultData) {
		this.resultData = resultData;
		this.invalidate();
	}

	setMode(mode: TimelineMode) {
		if (this.mode == mode) return;
		this.mode = mode;
		this.dpsResourcesPlotElem.classList.toggle('hide', mode != 'chart');
		this.rotationPlotElem.classList.toggle('hide', mode != 'rotation');
		this.invalidate();
	}

	canZoomIn(): boolean {
		return this.zoomIndex < zoomLevels.length - 1;
	}

	canZoomOut(): boolean {
		return this.zoomIndex > 0;
	}

	// Zooms the rotation in or out a step, keeping the time in the middle of the view where it is.
	zoom(step: 1 | -1) {
		const newIndex = Math.min(zoomLevels.length - 1, Math.max(0, this.zoomIndex + step));
		if (newIndex == this.zoomIndex) return;
		const middle = this.leftToTime(this.rotationTimeline.scrollLeft + this.rotationTimeline.clientWidth / 2);
		this.zoomIndex = newIndex;
		this.invalidate();
		this.rotationTimeline.scrollLeft = this.timeToLeftValue(middle) - this.rotationTimeline.clientWidth / 2;
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
	}

	private updatePlot() {
		if (this.resultData == null) {
			return;
		}

		const duration = this.resultData!.result.result.firstIterationDuration || 1;
		// The prepull actions, like a totem 3 seconds before the pull, happen at negative times.
		// Without a From time, the timeline starts at the first of them, on a whole second.
		const firstEvent = this.resultData!.result.logs.reduce((first, log) => Math.min(first, log.timestamp), 0);
		this.prepullStart = Math.floor(firstEvent);
		this.viewStart = Math.max(this.prepullStart, Math.min(this.filter.from ?? this.prepullStart, duration));
		this.viewEnd = Math.min(duration, this.filter.to ?? duration);
		if (this.viewEnd <= this.viewStart) {
			this.viewEnd = Math.min(duration, this.viewStart + 1);
		}

		const players = this.resultData!.result.getRaidIndexedPlayers(this.resultData!.filter);

		// The rotation follows one player. With more than one we can only show the chart.
		if (this.mode == 'rotation' && players.length == 1) {
			try {
				this.updateRotationChart(players[0], duration);
			} catch (e) {
				console.log('Failed to update rotation chart: ', e);
			}
			return;
		}
		this.clearRotationChart();
		this.dpsResourcesPlotElem.classList.remove('hide');
		this.rotationPlotElem.classList.add('hide');

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
			const dpsData = this.addDpsSeries(player, options, '');
			this.addDpsYAxis(dpsData.maxDps, options);
			tooltipHandlers.push(dpsData.tooltipHandler);
			tooltipHandlers.push(this.addManaSeries(player, options));
			tooltipHandlers.push(this.addThreatSeries(player, options, ''));
			tooltipHandlers = tooltipHandlers.filter(handler => handler != null);

			this.addMajorCooldownAnnotations(player, options);
		} else {
			let maxDps = 0;
			players.forEach(player => {
				const dpsData = this.addDpsSeries(player, options, `var(--bs-${player.classColor}`);
				maxDps = Math.max(maxDps, dpsData.maxDps);
				tooltipHandlers.push(dpsData.tooltipHandler);
			});
			this.addDpsYAxis(maxDps, options);
		}

		this.dpsResourcesPlot.updateOptions(options);
	}

	private addDpsYAxis(maxDps: number, options: any) {
		const dpsAxisMax = Math.ceil(maxDps / 100) * 100;
		options.yaxis.push({
			color: dpsColor,
			seriesName: 'DPS',
			min: 0,
			max: dpsAxisMax,
			tickAmount: 10,
			decimalsInFloat: 0,
			title: {
				text: 'DPS',
				style: {
					color: dpsColor,
				},
			},
			axisBorder: {
				show: true,
				color: dpsColor,
			},
			axisTicks: {
				color: dpsColor,
			},
			labels: {
				minWidth: 30,
				style: {
					colors: [dpsColor],
				},
			},
		});
	}

	// Returns a function for drawing the tooltip, or null if no series was added.
	private addDpsSeries(unit: UnitMetrics, options: any, colorOverride: string): { maxDps: number; tooltipHandler: TooltipHandler } {
		const dpsLogs = unit.dpsLogs.filter(log => log.timestamp >= 0);

		options.colors.push(colorOverride || dpsColor);
		options.series.push({
			name: 'DPS',
			type: 'line',
			data: dpsLogs.map(log => {
				return {
					x: log.timestamp,
					y: log.dps,
				};
			}),
		});

		return {
			maxDps: dpsLogs[maxIndex(dpsLogs.map(l => l.dps))!].dps,
			tooltipHandler: (dataPointIndex: number) => {
				const log = dpsLogs[dataPointIndex];
				return this.dpsTooltip(log, true, unit, colorOverride);
			},
		};
	}

	// Returns a function for drawing the tooltip, or null if no series was added.
	private addManaSeries(unit: UnitMetrics, options: any): TooltipHandler | null {
		const allManaLogs = unit.groupedResourceLogs[ResourceType.ResourceTypeMana];
		const manaLogs = allManaLogs.filter(log => log.timestamp >= 0);
		if (manaLogs.length == 0) {
			return null;
		}
		// The full pool is the mana before the first change, prepull included. The first change
		// after the pull can come after prepull totems already spent some, and then regen would
		// draw the line over the top of the axis.
		const maxMana = Math.max(...allManaLogs.map(log => Math.max(log.valueBefore, log.valueAfter)));

		options.colors.push(manaColor);
		options.series.push({
			name: 'Mana',
			type: 'line',
			data: manaLogs.map(log => {
				return {
					x: log.timestamp,
					y: log.valueAfter,
				};
			}),
		});
		options.yaxis.push({
			seriesName: 'Mana',
			opposite: true, // Appear on right side
			min: 0,
			max: maxMana,
			tickAmount: 10,
			title: {
				text: 'Mana',
				style: {
					color: manaColor,
				},
			},
			axisBorder: {
				show: true,
				color: manaColor,
			},
			axisTicks: {
				color: manaColor,
			},
			labels: {
				minWidth: 30,
				style: {
					colors: [manaColor],
				},
				formatter: (val: string) => {
					const v = parseFloat(val);
					return `${v.toFixed(0)} (${((v / maxMana) * 100).toFixed(0)}%)`;
				},
			},
		} as any);

		return (dataPointIndex: number) => {
			const log = manaLogs[dataPointIndex];
			return this.resourceTooltip(log, maxMana, true);
		};
	}

	// Returns a function for drawing the tooltip, or null if no series was added. Threat adds up
	// over the fight to far more than the DPS, so it gets its own axis on the right.
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
			data: threatLogs.map(log => {
				return {
					x: log.timestamp,
					y: log.threatAfter,
				};
			}),
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
			const log = threatLogs[dataPointIndex];
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
		const targets = this.resultData!.result.getTargets(this.resultData!.filter);
		if (targets.length == 0) {
			return;
		}
		const target = targets[0];

		this.clearRotationChart();
		this.dpsResourcesPlotElem.classList.add('hide');
		this.rotationPlotElem.classList.remove('hide');

		this.drawRotationTimeRuler(this.rotationTimeline.querySelector('.rotation-timeline-header')!);
		this.drawPull();

		orderedResourceTypes.forEach(resourceType => this.addResourceRow(resourceType, player.groupedResourceLogs[resourceType], duration));
		this.addGcdRow(player, duration);

		const buffsById = Object.values(bucket(player.auraUptimeLogs, log => log.actionId!.toString()));
		buffsById.sort((a, b) => stringComparator(a[0].actionId!.name, b[0].actionId!.name));
		const debuffsById = Object.values(bucket(target.auraUptimeLogs, log => log.actionId!.toString()));
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
			buffsToShow.forEach(auraUptimeLogs => this.addAuraRow(auraUptimeLogs));
		}

		const targetCastsByAbility = this.getSortedCastsByAbility(target);
		if (targetCastsByAbility.length > 0) {
			this.addSeparatorRow();
			targetCastsByAbility.forEach(castLogs => this.addCastRow(castLogs, buffsAndDebuffsById));
		}

		// Add a row for all debuffs, even those which have already been visualized in a cast row.
		const debuffsToShow = debuffsById;
		if (debuffsToShow.length > 0) {
			this.addSeparatorRow();
			debuffsToShow.forEach(auraUptimeLogs => this.addAuraRow(auraUptimeLogs));
		}

		this.applyFilter();
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

	private getSortedCastsByAbility(player: UnitMetrics): Array<Array<CastLog>> {
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
			bucket(player.castLogs, log => {
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
		const height = 28;
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
		svg.setAttribute('width', String(Math.max(endX, 1)));
		svg.setAttribute('height', String(height));
		const areaPath = document.createElementNS(svgNs, 'path');
		areaPath.setAttribute('class', 'rotation-resource-area');
		areaPath.setAttribute('d', area);
		const linePath = document.createElementNS(svgNs, 'path');
		linePath.setAttribute('class', 'rotation-resource-line');
		linePath.setAttribute('d', line);
		svg.append(areaPath, linePath);
		rowElem.appendChild(svg);

		// The labels, about one per 48px. The search only looks at the lane's name, like "mana".
		// Its changes name the abilities that spent or gave it, like "from Earth Shock", and the
		// lane would light up for every ability search.
		let lastLabelX = -Infinity;
		resourceLogs.forEach(resourceLogGroup => {
			const x = this.timeToLeftValue(resourceLogGroup.timestamp);
			if (x - lastLabelX >= 48) {
				lastLabelX = x;
				rowElem.appendChild(
					<span className={`rotation-resource-label series-color ${cNames}`} style={{ left: `${x}px` }}>
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
		// doesn't count.
		let busyUntil = 0;
		let busyTime = 0;
		const addIdle = (from: number, to: number) => {
			// Gaps shorter than this are the sim's reaction time, not a missed GCD.
			if (to - from < 0.05) return;
			const idleElem = (<div className="rotation-gcd-idle" style={{ left: this.timeToLeft(from), width: this.timeToPx(to - from) }} />) as HTMLElement;
			rowElem.appendChild(idleElem);
			const tooltip = tippy(idleElem, {
				...followAlong,
				placement: 'bottom',
				content: (
					<div className="timeline-tooltip">
						<div className="timeline-tooltip-title">
							<span className="timeline-tooltip-name">Idle</span>
						</div>
						<div className="timeline-tooltip-meta">
							{`${formatSeconds(from)} to ${formatSeconds(to)} · ${formatSeconds(to - from)} with the GCD ready and nothing pressed`}
						</div>
					</div>
				),
			});
			this.addOnResetCallback(() => tooltip.destroy());
		};

		casts.forEach(castLog => {
			const start = castLog.timestamp;
			const end = start + castLog.effectiveTime;
			if (start > busyUntil) addIdle(Math.max(0, busyUntil), start);
			busyTime += Math.max(0, Math.min(end, duration) - Math.max(start, busyUntil, 0));
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
			const tooltip = tippy(gcdElem, {
				...followAlong,
				placement: 'bottom',
				content: (
					<div className="timeline-tooltip">
						{tooltipTitle(actionId)}
						<div className="timeline-tooltip-meta">{timing.join(' · ')}</div>
					</div>
				),
			});
			this.addOnResetCallback(() => tooltip.destroy());
		});
		if (busyUntil < duration) addIdle(Math.max(0, busyUntil), duration);

		const usage = duration > 0 ? (busyTime / duration) * 100 : 0;
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
			content: 'The share of the fight, from the pull on, that we spent under the global cooldown or casting. Striped gaps are time with the GCD ready and nothing pressed.',
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

			const tt = (
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

			tippy(castElem, {
				...followAlong,
				placement: 'bottom',
				content: tt,
			});

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

					const tt = (
						<div className="timeline-tooltip">
							{tooltipTitle(ddl.actionId!, 'Tick')}
							<ul className="timeline-tooltip-hits">{tooltipHit(ddl)}</ul>
						</div>
					);

					tippy(tickElem, {
						placement: 'bottom',
						content: tt,
					});
				});
		});

		// If there are any auras that correspond to this cast, visualize them in the same row.
		aurasById
			.filter(auraUptimeLogs => auraUptimeLogs[0].actionId!.equalsIgnoringTag(actionId))
			.forEach(auraUptimeLogs => this.applyAuraUptimeLogsToRow(auraUptimeLogs, row));
	}

	private addAuraRow(auraUptimeLogs: Array<AuraUptimeLog>) {
		const actionId = auraUptimeLogs[0].actionId!;

		const row: RotationRow = { label: this.makeLabelElem(actionId, true), row: this.makeRowElem(), name: actionId.name, actionId, category: 'auras', events: [] };
		this.addRow(row);

		this.applyAuraUptimeLogsToRow(auraUptimeLogs, row);
	}

	private applyAuraUptimeLogsToRow(auraUptimeLogs: Array<AuraUptimeLog>, row: RotationRow) {
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
			row.events.push({ elem: auraElem, text: searchText([aul, ...aul.stacksChange]) });

			const tt = (
				<div className="timeline-tooltip">
					{tooltipTitle(aul.actionId!)}
					<div className="timeline-tooltip-meta">
						{`${formatSeconds(aul.gainedAt)} to ${formatSeconds(aul.fadedAt)} · ${formatSeconds(aul.fadedAt - aul.gainedAt)}`}
					</div>
				</div>
			);

			const tooltip = tippy(auraElem, {
				...followAlong,
				placement: 'bottom',
				content: tt,
			});
			this.addOnResetCallback(() => tooltip.destroy());

			aul.stacksChange.forEach((scl, i) => {
				if (scl.timestamp == aul.fadedAt) {
					return;
				}

				const stacksChangeElem = (
					<div
						className="rotation-timeline-stacks-change"
						style={{
							left: this.timeToPx(scl.timestamp - aul.timestamp),
							width: this.timeToPx(aul.stacksChange[i + 1] ? aul.stacksChange[i + 1].timestamp - scl.timestamp : aul.fadedAt - scl.timestamp),
						}}>
						{String(scl.newStacks)}
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

	private get pxPerSecond(): number {
		return zoomLevels[this.zoomIndex];
	}

	// For lengths, like how long a cast or an aura lasts.
	private timeToPxValue(time: number): number {
		return time * this.pxPerSecond;
	}
	private timeToPx(time: number): string {
		return this.timeToPxValue(time) + 'px';
	}

	// For where something starts, measured from the start of the time range we draw.
	private timeToLeftValue(time: number): number {
		return (time - this.viewStart) * this.pxPerSecond;
	}
	private timeToLeft(time: number): string {
		return this.timeToLeftValue(time) + 'px';
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
		const pullLeft = this.timeToLeftValue(Math.min(0, this.viewEnd));
		this.rotationTimeline.prepend(
			<div className="rotation-prepull" style={{ width: `${pullLeft}px` }}>
				<span className="rotation-prepull-label">Prepull</span>
			</div>,
		);
		if (this.viewEnd > 0) this.rotationTimeline.appendChild(<div className="rotation-pull-line" style={{ left: `${pullLeft}px` }} />);
	}

	private drawRotationTimeRuler(header: HTMLElement) {
		const [labelStep, tickStep] = rulerSteps.find(([label]) => label * this.pxPerSecond >= 60) ?? rulerSteps[rulerSteps.length - 1];
		header.style.width = this.viewWidthPx();

		// Faint lines down the rows at each label, so we can follow a time from the ruler.
		const gridStep = labelStep * this.pxPerSecond;
		this.rotationTimeline.style.setProperty('--grid-step', `${gridStep}px`);
		this.rotationTimeline.style.setProperty('--grid-offset', `${-((this.viewStart % labelStep) * this.pxPerSecond)}px`);

		const ticks: Array<Element> = [];
		const firstTick = Math.ceil(this.viewStart / tickStep - 1e-6);
		const lastTick = Math.floor(this.viewEnd / tickStep + 1e-6);
		for (let i = firstTick; i <= lastTick; i++) {
			const time = i * tickStep;
			const isLabel = Math.abs(time / labelStep - Math.round(time / labelStep)) < 1e-6;
			ticks.push(<div className={`rotation-ruler-tick${isLabel ? ' major' : ''}`} style={{ left: this.timeToLeft(time) }} />);
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
	private chartTooltipTitle(log: SimLog, player: UnitMetrics, colorOverride: string) {
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
				<span className="timeline-tooltip-name">{formatSeconds(log.timestamp)}</span>
			</div>
		);
	}

	private dpsTooltip(log: DpsLog, _includeAuras: boolean, player: UnitMetrics, colorOverride: string) {
		return (
			<div className="timeline-tooltip dps">
				{this.chartTooltipTitle(log, player, colorOverride)}
				{log.damageLogs.length > 0 && <ul className="timeline-tooltip-hits">{log.damageLogs.map(damageLog => tooltipHit(damageLog, true))}</ul>}
				<div className="timeline-tooltip-rows">{tooltipRow('DPS', formatAmount(log.dps, 1), 'series-color')}</div>
				{this.tooltipAurasSection(log)}
			</div>
		);
	}

	private threatTooltip(log: ThreatLogGroup, includeAuras: boolean, player: UnitMetrics, colorOverride: string) {
		return (
			<div className="timeline-tooltip threat">
				{this.chartTooltipTitle(log, player, colorOverride)}
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
];

const formatRulerTime = (time: number): string => {
	if (time < 0) return `-${formatRulerTime(-time)}`;
	if (time < 60) return `${Number(time.toFixed(1))}s`;
	const minutes = Math.floor(time / 60);
	const seconds = Math.round(time - minutes * 60);
	return `${minutes}:${String(seconds).padStart(2, '0')}`;
};
