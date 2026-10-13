import ApexCharts from 'apexcharts';
import tippy from 'tippy.js';
import { ref } from 'tsx-vanilla';

import { ResourceType } from '../../proto/api.js';
import { OtherAction } from '../../proto/common.js';
import { ActionId, resourceTypeToIcon } from '../../proto_utils/action_id.js';
import { AuraUptimeLog, CastLog, DamageDealtLog, DpsLog, ResourceChangedLogGroup, SimLog, ThreatLogGroup } from '../../proto_utils/logs_parser.js';
import { resourceNames } from '../../proto_utils/names.js';
import { UnitMetrics } from '../../proto_utils/sim_result.js';
import { orderedResourceTypes } from '../../proto_utils/utils.js';
import { bucket, distinct, maxIndex, stringComparator } from '../../utils.js';
import { actionColors } from './color_settings.js';
import { EventCategory, EventFilter } from './events_filter.js';
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
// so the filter can show and hide rows without drawing the rotation again.
interface RotationRow {
	label: HTMLElement;
	row: HTMLElement;
	name: string;
	actionId: ActionId | null;
	category: EventCategory;
}

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
		this.viewStart = Math.max(0, Math.min(this.filter.from ?? 0, duration));
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
		const manaLogs = unit.groupedResourceLogs[ResourceType.ResourceTypeMana].filter(log => log.timestamp >= 0);
		if (manaLogs.length == 0) {
			return null;
		}
		const maxMana = manaLogs[0].valueBefore;

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

	// Returns a function for drawing the tooltip, or null if no series was added.
	private addThreatSeries(unit: UnitMetrics, options: any, colorOverride: string): TooltipHandler | null {
		options.colors.push(colorOverride || threatColor);
		options.series.push({
			name: 'Threat',
			type: 'line',
			data: unit.threatLogs
				.filter(log => log.timestamp >= 0)
				.map(log => {
					return {
						x: log.timestamp,
						y: log.threatAfter,
					};
				}),
		});

		return (dataPointIndex: number) => {
			const log = unit.threatLogs[dataPointIndex];
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

		orderedResourceTypes.forEach(resourceType => this.addResourceRow(resourceType, player.groupedResourceLogs[resourceType], duration));

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
	// a row.
	private applyFilter() {
		let anyShown = false;
		this.rows.forEach(entry => {
			if (entry == 'separator') return;
			const shown = this.isRowShown(entry);
			entry.label.classList.toggle('hide', !shown);
			entry.row.classList.toggle('hide', !shown);
			anyShown ||= shown;
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
		if (row.actionId && this.filter.isHidden(row.actionId)) return false;
		return this.filter.matchesText(row.name.toLowerCase());
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
		this.addRow({ label: labelElem, row: this.makeRowElem(), name: petName, actionId: null, category: 'casts' });

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

	private addResourceRow(resourceType: ResourceType, resourceLogs: Array<ResourceChangedLogGroup>, duration: number) {
		if (resourceLogs.length == 0) {
			return;
		}
		const startValue = resourceLogs[0].valueBefore;

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

		resourceLogs.forEach((resourceLogGroup, i) => {
			const cNames = resourceNames.get(resourceType)!.toLowerCase().replaceAll(' ', '-');
			const resourceElem = (
				<div
					className={`rotation-timeline-resource series-color ${cNames}`}
					style={{
						left: this.timeToLeft(resourceLogGroup.timestamp),
						width: this.timeToPx((resourceLogs[i + 1]?.timestamp || duration) - resourceLogGroup.timestamp),
					}}></div>
			);

			if (percentageResources.includes(resourceType)) {
				resourceElem.textContent = ((resourceLogGroup.valueAfter / startValue) * 100).toFixed(0) + '%';
			} else {
				if (resourceType == ResourceType.ResourceTypeEnergy) {
					const bgElem = document.createElement('div');
					bgElem.classList.add('rotation-timeline-resource-fill');
					bgElem.classList.add(cNames);
					bgElem.style.height = ((resourceLogGroup.valueAfter / startValue) * 100).toFixed(0) + '%';
					resourceElem.appendChild(bgElem);
				} else {
					resourceElem.textContent = Math.floor(resourceLogGroup.valueAfter).toFixed(0);
				}
			}
			rowElem.appendChild(resourceElem);

			const tooltip = tippy(resourceElem, {
				placement: 'bottom',
				content: this.resourceTooltipElem(resourceLogGroup, startValue, false),
			});
			this.addOnResetCallback(() => tooltip.destroy());
		});
		this.addRow({ label: labelElem, row: rowElem, name: resourceNames.get(resourceType)!, actionId: null, category: 'resources' });
	}

	private addCastRow(castLogs: Array<CastLog>, aurasById: Array<Array<AuraUptimeLog>>) {
		const actionId = castLogs[0].actionId!;
		const rowElem = this.makeRowElem();
		this.addRow({ label: this.makeLabelElem(actionId), row: rowElem, name: actionId.name, actionId, category: 'casts' });

		let stackedIconCount = 1;
		let stackedDamageCount = 1;

		castLogs.forEach((castLog, index) => {
			const castElem = (
				<div
					className="rotation-timeline-cast"
					style={{
						left: this.timeToLeft(castLog.timestamp),
						minWidth: this.timeToPx(castLog.castTime + castLog.travelTime),
					}}></div>
			);

			rowElem.appendChild(castElem);

			if (castLog.travelTime != 0) {
				const travelTimeElem = (
					<div
						className="rotation-timeline-travel-time"
						style={{
							left: this.timeToPx(castLog.castTime),
							minWidth: this.timeToPx(castLog.travelTime),
						}}
					/>
				);
				castElem.appendChild(travelTimeElem);
			}

			if (castLog.damageDealtLogs.length > 0) {
				const ddl = castLog.damageDealtLogs[0];
				if (ddl.miss || ddl.dodge || ddl.parry) {
					castElem.classList.add('outcome-miss');
				} else if (ddl.glance || ddl.block || ddl.partialResist1_4 || ddl.partialResist2_4 || ddl.partialResist3_4) {
					castElem.classList.add('outcome-partial');
				} else if (ddl.crit) {
					castElem.classList.add('outcome-crit');
				} else {
					castElem.classList.add('outcome-hit');
				}
			}

			const iconElem = (<a className="rotation-timeline-cast-icon" />) as HTMLAnchorElement;
			actionId.setBackground(iconElem);
			castElem.appendChild(iconElem);

			const travelTimeStr = castLog.travelTime == 0 ? '' : ` + ${castLog.travelTime.toFixed(2)}s travel time`;
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

			const tt = (
				<div className="timeline-tooltip">
					<span>
						{castLog.actionId!.name} from {castLog.timestamp.toFixed(2)}s to {(castLog.timestamp + castLog.castTime).toFixed(2)}s (
						{castLog.castTime > 0 && `${castLog.castTime.toFixed(2)}s, `}
						{castLog.effectiveTime > 0 && `${castLog.effectiveTime.toFixed(2)}s GCD Time`}){travelTimeStr.length > 0 && travelTimeStr}
					</span>
					{aggregatedData.damageDealtLogs.length > 0 && (
						<ul className="rotation-timeline-cast-damage-list">
							{aggregatedData.damageDealtLogs.map((ddl, ddlIndex) => (
								<li>
									<span>
										{ddl.timestamp.toFixed(2)}s - {ddl.result()}
									</span>
									{ddl.source?.isTarget && <span className="threat-metrics"> ({ddl.threat.toFixed(1)} Threat)</span>}
								</li>
							))}
						</ul>
					)}
					{totalDamage > 0 && (
						<span>
							Total: {aggregatedData.totalDamage.toFixed(2)} ({(aggregatedData.totalDamage / (castLog.effectiveTime || 1)).toFixed(2)} DPET)
						</span>
					)}
				</div>
			);

			tippy(castElem, {
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

					const tt = (
						<div className="timeline-tooltip">
							<span>
								{ddl.timestamp.toFixed(2)}s - {ddl.actionId!.name} {ddl.result()}
							</span>
							{ddl.source?.isTarget && <span className="threat-metrics"> ({ddl.threat.toFixed(1)} Threat)</span>}
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
			.forEach(auraUptimeLogs => this.applyAuraUptimeLogsToRow(auraUptimeLogs, rowElem));
	}

	private addAuraRow(auraUptimeLogs: Array<AuraUptimeLog>) {
		const actionId = auraUptimeLogs[0].actionId!;

		const rowElem = this.makeRowElem();
		this.addRow({ label: this.makeLabelElem(actionId, true), row: rowElem, name: actionId.name, actionId, category: 'auras' });

		this.applyAuraUptimeLogsToRow(auraUptimeLogs, rowElem);
	}

	private applyAuraUptimeLogsToRow(auraUptimeLogs: Array<AuraUptimeLog>, rowElem: JSX.Element) {
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
			rowElem.appendChild(auraElem);

			const tt = (
				<div className="timeline-tooltip">
					<span>
						{aul.actionId!.name}: {aul.gainedAt.toFixed(2)}s - {aul.fadedAt.toFixed(2)}s
					</span>
				</div>
			);

			const tooltip = tippy(auraElem, {
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
					<div className="rotation-ruler-label" style={{ left: this.timeToLeft(time) }}>
						{formatRulerTime(time)}
					</div>,
				);
			}
		}
		header.replaceChildren(...ticks);
	}

	private dpsTooltip(log: DpsLog, _includeAuras: boolean, player: UnitMetrics, colorOverride: string) {
		const showPlayerLabel = colorOverride != '';
		return (
			<div className="timeline-tooltip dps">
				<div className="timeline-tooltip-header">
					{showPlayerLabel ? (
						<>
							<img className="timeline-tooltip-icon" src="${player.iconUrl}" />
							<span className="" style="color: ${colorOverride}">
								{player.label}
							</span>
							<span> - </span>
						</>
					) : null}
					<span className="bold">{log.timestamp.toFixed(2)}s</span>
				</div>
				<div className="timeline-tooltip-body">
					<ul className="timeline-dps-events">{log.damageLogs.map(damageLog => this.tooltipLogItem(damageLog, damageLog.result()))}</ul>
					<div className="timeline-tooltip-body-row">
						<span className="series-color">DPS: {log.dps.toFixed(2)}</span>
					</div>
				</div>
				{this.tooltipAurasSection(log)}
			</div>
		);
	}

	private threatTooltip(log: ThreatLogGroup, includeAuras: boolean, player: UnitMetrics, colorOverride: string) {
		const showPlayerLabel = colorOverride != '';
		return (
			<div className="timeline-tooltip threat">
				<div className="timeline-tooltip-header">
					{showPlayerLabel ? (
						<>
							<img className="timeline-tooltip-icon" src={player.iconUrl} />
							<span className="" style={{ color: colorOverride }}>
								{player.label}
							</span>
							<span> - </span>
						</>
					) : null}
					<span className="bold">{log.timestamp.toFixed(2)}s</span>
				</div>
				<div className="timeline-tooltip-body">
					<div className="timeline-tooltip-body-row">
						<span className="series-color">Before: {log.threatBefore.toFixed(1)}</span>
					</div>
					<ul className="timeline-threat-events">{log.logs.map(log => this.tooltipLogItem(log, <>{log.threat.toFixed(1)} Threat</>))}</ul>
					<div className="timeline-tooltip-body-row">
						<span className="series-color">After: {log.threatAfter.toFixed(1)}</span>
					</div>
				</div>
				{includeAuras ? this.tooltipAurasSection(log) : null}
			</div>
		);
	}

	private resourceTooltipElem(log: ResourceChangedLogGroup, maxValue: number, includeAuras: boolean) {
		const valToDisplayString = percentageResources.includes(log.resourceType)
			? (val: number) => `${val.toFixed(1)} (${((val / maxValue) * 100).toFixed(0)}%)`
			: (val: number) => `${val.toFixed(1)}`;

		return (
			<div className={`timeline-tooltip ${resourceNames.get(log.resourceType)!.toLowerCase().replaceAll(' ', '-')}`}>
				<div className="timeline-tooltip-header">
					<span className="bold">{log.timestamp.toFixed(2)}s</span>
				</div>
				<div className="timeline-tooltip-body">
					<div className="timeline-tooltip-body-row">
						<span className="series-color">Before: {valToDisplayString(log.valueBefore)}</span>
					</div>
					<ul className="timeline-mana-events">
						{log.logs.map(manaChangedLog => this.tooltipLogItemElem(manaChangedLog, <>{manaChangedLog.resultString()}</>))}
					</ul>
					<div className="timeline-tooltip-body-row">
						<span className="series-color">After: {valToDisplayString(log.valueAfter)}</span>
					</div>
				</div>
				{includeAuras && this.tooltipAurasSectionElem(log)}
			</div>
		);
	}

	private resourceTooltip(log: ResourceChangedLogGroup, maxValue: number, includeAuras: boolean) {
		return this.resourceTooltipElem(log, maxValue, includeAuras);
	}

	private tooltipLogItem(log: SimLog, value: Element) {
		return this.tooltipLogItemElem(log, value);
	}

	private tooltipLogItemElem(log: SimLog, value: Element): JSX.Element {
		return (
			<li>
				{log.actionId && log.actionId.iconUrl && <img className="timeline-tooltip-icon" src={log.actionId.iconUrl}></img>}
				{log.actionId && <span>{log.actionId.name}</span>}
				<span className="series-color">{value}</span>
			</li>
		);
	}

	private tooltipAurasSection(log: SimLog) {
		if (log.activeAuras.length == 0) {
			return '';
		}
		return this.tooltipAurasSectionElem(log);
	}

	private tooltipAurasSectionElem(log: SimLog): JSX.Element {
		if (log.activeAuras.length == 0) {
			return <></>;
		}

		return (
			<div className="timeline-tooltip-auras">
				<div className="timeline-tooltip-body-row">
					<span className="bold">Active Auras</span>
				</div>
				<ul className="timeline-active-auras">
					{log.activeAuras.map(auraLog => (
						<li>
							{auraLog.actionId!.iconUrl && <img className="timeline-tooltip-icon" src={auraLog.actionId!.iconUrl}></img>}
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
	if (time < 60) return `${Number(time.toFixed(1))}s`;
	const minutes = Math.floor(time / 60);
	const seconds = Math.round(time - minutes * 60);
	return `${minutes}:${String(seconds).padStart(2, '0')}`;
};
