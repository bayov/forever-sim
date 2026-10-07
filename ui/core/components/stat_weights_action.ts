import tippy from 'tippy.js';

import { BooleanPicker } from '../components/boolean_picker.js';
import { NumberPicker } from '../components/number_picker.js';
import { IndividualSimUI } from '../individual_sim_ui.js';
import { Player } from '../player.js';
import { ProgressMetrics, StatWeightsResult, StatWeightValues } from '../proto/api.js';
import { PseudoStat, Stat, UnitStats } from '../proto/common.js';
import { getClassStatName } from '../proto_utils/names.js';
import { Stats, UnitStat } from '../proto_utils/stats.js';
import { RequestTypes } from '../sim_signal_manager';
import { EventID, TypedEvent } from '../typed_event.js';
import { stDevToConf90 } from '../utils.js';
import { Component } from './component.js';
import { ContentBlock } from './content_block.js';
import { dirtySettings, PresetSource } from './dirty_settings.js';
import { presetListTooltip, PresetTree } from './preset_tree.js';
import { ResultsViewer } from './results_viewer.js';

// Stat weights are a section of the Gear tab, under the gear, because the weights are for
// picking gear. The raid sim's player editor leaves it out, like the sidebar actions.
export function addStatWeightsSection(
	parent: HTMLElement,
	simUI: IndividualSimUI<any>,
	epStats: Array<Stat>,
	epPseudoStats: Array<PseudoStat> | undefined,
	epReferenceStat: Stat,
) {
	new EpWeightsMenu(parent, simUI, epStats, epPseudoStats || [], epReferenceStat);
}

// A stat's calculated EP (or weight) over every metric, each scaled by its EP ratio. For a DPS
// spec that's the DPS EP, because only the DPS ratio is set.
function scaledValue(stat: UnitStat, epRatios: number[], result: StatWeightsResult | null, kind: 'epValues' | 'weights'): number {
	if (!result) return 0;
	return [result.dps, result.hps, result.tps, result.dtps, result.tmi, result.pDeath].reduce((total, values, i) => {
		const protoValues = values?.[kind];
		return total + (protoValues ? epRatios[i] * stat.getProtoValue(protoValues) : 0);
	}, 0);
}

function scaledEpValue(stat: UnitStat, epRatios: number[], result: StatWeightsResult | null): number {
	return scaledValue(stat, epRatios, result, 'epValues');
}

// The values we copy into the Current EP keep two decimals, like the table shows them.
const roundEp = (value: number) => Math.round(value * 100) / 100;

const DEFAULT_ITERATIONS = 3000;

class EpWeightsMenu extends Component {
	private readonly simUI: IndividualSimUI<any>;
	private readonly body: HTMLElement;
	private readonly container: HTMLElement;
	private readonly table: HTMLElement;
	private readonly tableBody: HTMLElement;
	private readonly resultsViewer: ResultsViewer;

	private statsType: string;
	private epStats: Array<Stat>;
	private epPseudoStats: Array<PseudoStat>;
	private epReferenceStat: Stat;
	private showAllStats = false;
	private iterations: number;
	private readonly iterationsChangeEmitter = new TypedEvent<void>();
	private readonly applyAllButton: HTMLButtonElement;

	constructor(parent: HTMLElement, simUI: IndividualSimUI<any>, epStats: Array<Stat>, epPseudoStats: Array<PseudoStat>, epReferenceStat: Stat) {
		super(parent, 'ep-weights-menu');
		this.rootElem.classList.add('within-raid-sim-hide');
		// The changes on a modified preset are listed under this name, like under a tab.
		this.rootElem.dataset.presetCategory = 'Stat Weights';
		this.simUI = simUI;
		this.statsType = 'ep';
		this.epStats = epStats;
		this.epPseudoStats = epPseudoStats;
		this.epReferenceStat = epReferenceStat;

		const contentBlock = new ContentBlock(this.rootElem, 'ep-weights-block', {
			header: { title: 'Stat Weights', tooltip: 'How much DPS each stat is worth, from sims that add a little of each stat.' },
		});
		this.body = contentBlock.bodyElement;
		this.body.innerHTML = `
			<div class="ep-weights-layout">
				<div class="ep-weights-settings">
					<div class="ep-weights-presets">
						<span>Presets:</span>
						<div class="saved-data-presets"></div>
					</div>
					<div class="ep-reference-options experimental">
						<div class="damage-metrics">
							<span>DPS/TPS reference:</span>
							<select class="ref-stat-select form-select damage-metrics"></select>
						</div>
						<div class="healing-metrics">
							<span>Healing reference:</span>
							<select class="ref-stat-select form-select healing-metrics"></select>
						</div>
						<div class="threat-metrics">
							<span>Mitigation reference:</span>
							<select class="ref-stat-select form-select threat-metrics"></select>
						</div>
						<p>The above stat selectors control which reference stat is used for EP normalisation for the different EP columns.</p>
					</div>
					<div class="ep-weights-actions">
						<div class="ep-iterations-container"></div>
						<button class="btn btn-primary calc-weights">
							<i class="fas fa-calculator"></i>
							Calculate
						</button>
						<button class="btn btn-outline-primary apply-all-ep">
							<i class="fas fa-arrow-left"></i>
							Apply All
						</button>
					</div>
				</div>
				<div class="ep-weights-results">
					<div class="ep-weights-options">
						<div class="ep-type-container">
							<label class="form-label">Show</label>
							<select class="ep-type-select form-select">
								<option value="ep">EP</option>
								<option value="weight">Weights</option>
							</select>
						</div>
						<div class="show-all-stats-container"></div>
					</div>
					<div class="results-ep-table-container">
						<div class="results-pending-overlay"></div>
						<table class="results-ep-table">
							<thead>
								<tr>
									<th>Stat</th>
									<th class="current-ep-header"><span>Current EP</span></th>
									<th class="damage-metrics type-weight metric-header"><span>Calculated Weight</span></th>
									<th class="damage-metrics type-ep metric-header"><span>Calculated EP</span></th>
									<th class="healing-metrics type-weight metric-header"><span>HPS Weight</span></th>
									<th class="healing-metrics type-ep metric-header"><span>HPS EP</span></th>
									<th class="threat-metrics type-weight metric-header"><span>TPS Weight</span></th>
									<th class="threat-metrics type-ep metric-header"><span>TPS EP</span></th>
									<th class="threat-metrics type-weight metric-header"><span>DTPS Weight</span></th>
									<th class="threat-metrics type-ep metric-header"><span>DTPS EP</span></th>
									<th class="threat-metrics experimental type-weight metric-header"><span>TMI Weight</span></th>
									<th class="threat-metrics experimental type-ep metric-header"><span>TMI EP</span></th>
									<th class="threat-metrics experimental type-weight metric-header"><span>Death Weight</span></th>
									<th class="threat-metrics experimental type-ep metric-header"><span>Death EP</span></th>
								</tr>
								<tr class="ep-ratios">
									<td>EP Ratio</td>
									<td></td>
									<td class="damage-metrics type-ratio type-weight"></td>
									<td class="damage-metrics type-ratio type-ep"></td>
									<td class="healing-metrics type-ratio type-weight"></td>
									<td class="healing-metrics type-ratio type-ep"></td>
									<td class="threat-metrics type-ratio type-weight"></td>
									<td class="threat-metrics type-ratio type-ep"></td>
									<td class="threat-metrics type-ratio type-weight"></td>
									<td class="threat-metrics type-ratio type-ep"></td>
									<td class="threat-metrics experimental type-ratio type-weight"></td>
									<td class="threat-metrics experimental type-ratio type-ep"></td>
									<td class="threat-metrics experimental type-ratio type-weight"></td>
									<td class="threat-metrics experimental type-ratio type-ep"></td>
								</tr>
							</thead>
							<tbody></tbody>
						</table>
					</div>
				</div>
			</div>
		`;

		this.container = this.rootElem.querySelector('.results-ep-table-container') as HTMLElement;
		this.table = this.rootElem.querySelector('.results-ep-table') as HTMLElement;
		this.tableBody = this.rootElem.querySelector('.results-ep-table tbody') as HTMLElement;
		this.applyAllButton = this.rootElem.querySelector('.apply-all-ep') as HTMLButtonElement;
		this.iterations = this.loadIterations();
		if (!this.simUI.prevEpSimResult) this.loadResult();

		const resultsElem = this.rootElem.querySelector('.results-pending-overlay') as HTMLElement;
		this.resultsViewer = new ResultsViewer(resultsElem);

		const updateType = () => {
			if (this.statsType == 'ep') {
				this.table.classList.remove('stats-type-weight');
				this.table.classList.add('stats-type-ep');
			} else {
				this.table.classList.add('stats-type-weight');
				this.table.classList.remove('stats-type-ep');
			}
		};

		const selectElem = this.rootElem.getElementsByClassName('ep-type-select')[0] as HTMLSelectElement;
		selectElem.addEventListener('input', _event => {
			this.statsType = selectElem.value;
			updateType();
		});
		selectElem.value = this.statsType;
		updateType();

		const getNameFromStat = (stat: Stat | undefined) => {
			return stat !== undefined ? getClassStatName(stat, this.simUI.player.getClass()) : '??';
		};

		const getStatFromName = (value: string) => {
			for (const stat of this.epStats) {
				if (getNameFromStat(stat) == value) {
					return stat;
				}
			}

			return undefined;
		};

		const updateEpRefStat = () => {
			this.simUI.player.epRefStatChangeEmitter.emit(TypedEvent.nextEventID());
			this.simUI.prevEpSimResult = this.calculateEp(this.getPrevSimResult());
			this.saveResult();
			this.updateTable();
		};

		const epRefSelects = this.rootElem.querySelectorAll('.ref-stat-select') as NodeListOf<HTMLSelectElement>;
		epRefSelects.forEach((epSelect: HTMLSelectElement, _idx: number) => {
			this.epStats.forEach(stat => {
				epSelect.options[epSelect.options.length] = new Option(getNameFromStat(stat));
			});
			if (epSelect.classList.contains('damage-metrics')) {
				epSelect.addEventListener('input', _event => {
					this.simUI.dpsRefStat = getStatFromName(epSelect.value);
					updateEpRefStat();
				});
				epSelect.value = getNameFromStat(this.getDpsEpRefStat());
			} else if (epSelect.classList.contains('healing-metrics')) {
				epSelect.addEventListener('input', _event => {
					this.simUI.healRefStat = getStatFromName(epSelect.value);
					updateEpRefStat();
				});
				epSelect.value = getNameFromStat(this.getHealEpRefStat());
			} else if (epSelect.classList.contains('threat-metrics')) {
				epSelect.addEventListener('input', _event => {
					this.simUI.tankRefStat = getStatFromName(epSelect.value);
					updateEpRefStat();
				});
				epSelect.value = getNameFromStat(this.getTankEpRefStat());
			}
		});

		const calcButton = this.rootElem.getElementsByClassName('calc-weights')[0] as HTMLButtonElement;
		let isRunning = false;
		calcButton.addEventListener('click', async _event => {
			if (isRunning) return;
			isRunning = true;

			try {
				await this.simUI.sim.signalManager.abortType(RequestTypes.All);
			} catch (error) {
				console.error(error);
				return;
			}

			calcButton.disabled = true;

			const previousContents = calcButton.innerHTML;
			calcButton.style.width = `${calcButton.getBoundingClientRect().width.toFixed(3)}px`;
			calcButton.innerHTML = `<i class="fa fa-spinner fa-spin"></i>&nbsp;Running`;
			this.container.scrollTo({ top: 0 });
			this.container.classList.add('pending');
			this.resultsViewer.setPending();
			const iterations = this.iterations;

			let waitAbort = false;
			this.resultsViewer.addAbortButton(async () => {
				if (waitAbort) return;
				try {
					waitAbort = true;
					await simUI.sim.signalManager.abortType(RequestTypes.StatWeights);
				} catch (error) {
					console.error('Error on stat weight abort!');
					console.error(error);
				} finally {
					waitAbort = false;
					if (!isRunning) {
						calcButton.disabled = false;
						calcButton.innerHTML = previousContents;
					}
				}
			});

			const result = await this.simUI.player.computeStatWeights(
				TypedEvent.nextEventID(),
				this.epStats,
				this.epPseudoStats,
				this.epReferenceStat,
				(progress: ProgressMetrics) => {
					this.setSimProgress(progress);
				},
				iterations,
			);
			this.container.classList.remove('pending');
			this.resultsViewer.hideAll();

			isRunning = false;
			if (!waitAbort) {
				calcButton.disabled = false;
				calcButton.innerHTML = previousContents;
			}
			if (!result) return;

			this.simUI.prevEpIterations = iterations;
			this.simUI.prevEpSimResult = this.calculateEp(result);
			this.saveResult();
			this.updateTable();
		});

		this.addHeaderTooltips();
		const showAllStatsContainer = this.rootElem.getElementsByClassName('show-all-stats-container')[0] as HTMLElement;
		new BooleanPicker(showAllStatsContainer, this, {
			id: 'ep-show-all-stats',
			label: 'Show All Stats',
			inline: true,
			changedEvent: () => new TypedEvent(),
			getValue: () => this.showAllStats,
			setValue: (eventID: EventID, menu: EpWeightsMenu, newValue: boolean) => {
				this.showAllStats = newValue;
				this.updateTable();
			},
		});

		new NumberPicker(this.rootElem.querySelector('.ep-iterations-container') as HTMLElement, this, {
			id: 'ep-iterations',
			label: 'Iterations',
			labelTooltip: "How many iterations each stat weights sim runs. The sidebar's Iterations is for the main sim only.",
			positive: true,
			changedEvent: () => this.iterationsChangeEmitter,
			getValue: () => this.iterations,
			setValue: (eventID: EventID, _menu: EpWeightsMenu, newValue: number) => {
				this.iterations = newValue > 0 ? newValue : DEFAULT_ITERATIONS;
				this.saveIterations();
				this.iterationsChangeEmitter.emit(eventID);
			},
		});

		tippy(calcButton, {
			content: `
				<p class="mb-1">We sim the character once as it is, and then once with each stat raised by 1 and once with it lowered by 1. All of them use the same random rolls, so the difference is the stat alone.</p>
				<p class="mb-1">A stat's weight is the DPS it adds per point. Its EP is that weight divided by the weight of ${getNameFromStat(
					this.getDpsEpRefStat(),
				)}, so ${getNameFromStat(this.getDpsEpRefStat())} is always 1.</p>
				<p class="mb-0">More iterations give steadier numbers, but take longer.</p>
			`,
			allowHTML: true,
		});

		tippy(this.applyAllButton, { content: 'Copy every calculated EP into the Current EP.' });
		this.applyAllButton.addEventListener('click', () => this.applyAll());

		this.buildPresets();
		this.updateTable();

		const makeEpRatioCell = (cell: HTMLElement, idx: number) => {
			new NumberPicker(cell, this.simUI.player, {
				id: `ep-ratio-${idx}`,
				float: true,
				changedEvent: (player: Player<any>) => player.epRatiosChangeEmitter,
				getValue: (_player: Player<any>) => this.simUI.player.getEpRatios()[idx],
				setValue: (eventID: EventID, player: Player<any>, newValue: number) => {
					const epRatios = player.getEpRatios();
					epRatios[idx] = newValue;
					player.setEpRatios(eventID, epRatios);
				},
			});
		};
		const epRatioCells = this.body.querySelectorAll('.type-ratio.type-ep') as NodeListOf<HTMLElement>;
		epRatioCells.forEach(makeEpRatioCell);
		this.simUI.player.epRatiosChangeEmitter.on(_eventID => this.updateTable());

		const weightRatioCells = this.body.querySelectorAll('.type-ratio.type-weight') as NodeListOf<HTMLElement>;
		weightRatioCells.forEach(makeEpRatioCell);
	}

	// The tooltips on the column titles. They say what each column holds, which stat the EP is
	// normalized by, and how the calculated values get into the Current EP.
	private addHeaderTooltips() {
		const getName = (stat: Stat) => getClassStatName(stat, this.simUI.player.getClass());
		const copyNote =
			'Green is above the Current EP and red is below it. The ± is the range we are 90% sure the value is in. The arrow on a row, or Apply All, copies it into the Current EP.';
		const metrics: Array<[string, (() => Stat) | undefined]> = [
			['How much DPS (damage per second) one point of each stat adds, from the last Calculate.', undefined],
			['The EP (equivalency points) of each stat for DPS, from the last Calculate.', () => this.getDpsEpRefStat()],
			['How much HPS (healing per second) one point of each stat adds.', undefined],
			['The EP of each stat for HPS (healing per second).', () => this.getHealEpRefStat()],
			['How much TPS (threat per second) one point of each stat adds.', undefined],
			['The EP of each stat for TPS (threat per second).', () => this.getDpsEpRefStat()],
			['How much DTPS (damage taken per second) one point of each stat takes away.', undefined],
			['The EP of each stat for DTPS (damage taken per second).', () => this.getTankEpRefStat()],
			['How much TMI (Theck-Meloree Index) one point of each stat takes away.', undefined],
			['The EP of each stat for TMI (Theck-Meloree Index).', () => this.getTankEpRefStat()],
			['How much p(death) one point of each stat takes away.', undefined],
			['The EP of each stat for p(death).', () => this.getTankEpRefStat()],
		];
		this.rootElem.querySelectorAll<HTMLElement>('.metric-header > span').forEach((label, i) => {
			const [text, refStat] = metrics[i];
			tippy(label, { content: () => [text, refStat ? `It's normalized by ${getName(refStat())}.` : '', copyNote].filter(Boolean).join(' ') });
		});
		tippy(this.rootElem.querySelector('.current-ep-header > span') as HTMLElement, {
			content:
				'The EP the item lists use to sort and compare gear. Type a value to change it. Ctrl+click the selected stat weights preset to go back to its values.',
		});
	}

	// Whether the last Calculate weighed the stat. The others have an empty Calculated EP and
	// keep their Current EP.
	private isCalculated(stat: UnitStat): boolean {
		return stat.isStat() ? this.epStats.includes(stat.getStat()) : this.epPseudoStats.includes(stat.getPseudoStat());
	}

	// The stat's value from the last Calculate, as the table shows it (EP or weight), or
	// undefined when we have none.
	private calculatedValue(stat: UnitStat): number | undefined {
		const result = this.simUI.prevEpSimResult;
		if (!result || !this.isCalculated(stat)) return undefined;
		return scaledValue(stat, this.simUI.player.getEpRatios(), result, this.statsType == 'ep' ? 'epValues' : 'weights');
	}

	private differsFromCurrent(stat: UnitStat): boolean {
		const calculated = this.calculatedValue(stat);
		return calculated !== undefined && calculated.toFixed(2) !== this.simUI.player.getEpWeights().getUnitStat(stat).toFixed(2);
	}

	// Copies every calculated value into the Current EP. Stats we didn't calculate keep theirs,
	// like the Stamina of the PvP presets.
	private applyAll() {
		let weights = this.simUI.player.getEpWeights();
		EpWeightsMenu.epUnitStats.forEach(stat => {
			const calculated = this.calculatedValue(stat);
			if (calculated !== undefined) weights = weights.withUnitStat(stat, roundEp(calculated));
		});
		this.simUI.player.setEpWeights(TypedEvent.nextEventID(), weights);
	}

	// The last calculated weights are kept in local storage with their iterations, so they are
	// still in the table after a reload.
	private loadResult() {
		try {
			const stored = window.localStorage.getItem(this.simUI.getStorageKey('__epResult__'));
			if (!stored) return;
			const { iterations, result } = JSON.parse(stored);
			this.simUI.prevEpSimResult = StatWeightsResult.fromJson(result);
			this.simUI.prevEpIterations = iterations ?? 0;
		} catch {
			// A result we can't read is one we calculate again.
		}
	}

	private saveResult() {
		if (!this.simUI.prevEpSimResult) return;
		try {
			window.localStorage.setItem(
				this.simUI.getStorageKey('__epResult__'),
				JSON.stringify({ iterations: this.simUI.prevEpIterations, result: StatWeightsResult.toJson(this.simUI.prevEpSimResult) }),
			);
		} catch {
			// Without storage the weights last until the page closes.
		}
	}

	private loadIterations(): number {
		try {
			const stored = parseInt(window.localStorage.getItem(this.simUI.getStorageKey('__epIterations__')) ?? '');
			return stored > 0 ? stored : DEFAULT_ITERATIONS;
		} catch {
			return DEFAULT_ITERATIONS;
		}
	}

	private saveIterations() {
		try {
			window.localStorage.setItem(this.simUI.getStorageKey('__epIterations__'), String(this.iterations));
		} catch {
			// Without storage we start from the default on the next visit.
		}
	}

	// One row for each of the spec's stat weight presets. A click makes it the Current EP,
	// and the row stays lit while the Current EP matches it.
	private buildPresets() {
		const container = this.body.querySelector('.ep-weights-presets') as HTMLElement;
		const player = this.simUI.player;
		const presets = (this.simUI.individualConfig.presets.epWeights ?? []).filter(preset => !preset.enableWhen || preset.enableWhen(player));
		if (!presets.length) {
			container.remove();
			return;
		}

		const tree = new PresetTree(this.simUI.getStorageKey('__epWeightsOpenFolders__'), this.simUI.getStorageKey('__selectedEpWeights__'));
		container.querySelector('.saved-data-presets')!.appendChild(tree.rootElem);
		tippy(container.querySelector('span')!, {
			content: presetListTooltip('Loading stat weights changes:', ['The weight of every stat, which the EP in the gear picker uses']),
		});

		const presetByItem = new Map<HTMLElement, (typeof presets)[number]>();
		const source: PresetSource = {
			selected: () => {
				const preset = presetByItem.get(tree.selected()?.item as HTMLElement);
				return preset && { apply: () => player.setEpWeights(TypedEvent.nextEventID(), preset.epWeights) };
			},
		};
		presets.forEach(preset => {
			const item = PresetTree.makeItem(preset.name);
			presetByItem.set(item, preset);
			tree.onClick(item, () => {
				player.setEpWeights(TypedEvent.nextEventID(), preset.epWeights);
				preset.onLoad?.(player);
				// The preset may weigh stats the table hides, like Stamina for PvP.
				this.updateTable();
			});
			tree.attachTooltip(item, preset.tooltip, 'right', () => dirtySettings.changesFor(source));
			tree.track(item, { name: preset.name, matches: () => player.getEpWeights().equals(preset.epWeights) });
			tree.add(item, preset.group);
		});

		// The Current EP column follows the weights we load anywhere, like with a preset
		// configuration.
		const listener = player.epWeightsChangeEmitter.on(() => {
			tree.refresh();
			this.updateTable();
		});
		const disposeSource = dirtySettings.addSource(source);
		this.addOnDisposeCallback(() => {
			listener.dispose();
			disposeSource();
		});
	}

	private setSimProgress(progress: ProgressMetrics) {
		this.resultsViewer.setContent(`
			<div class="results-sim">
				<div class=""> ${progress.completedSims} / ${progress.totalSims}<br>simulations complete</div>
				<div class="">
					${progress.completedIterations} / ${progress.totalIterations}<br>iterations complete
				</div>
			</div>
		`);
	}

	private updateTable() {
		// The table is built again when the Current EP changes, like after we type one in and
		// tab to the next row. We put the focus back on the same row's field.
		const focusedRow =
			document.activeElement instanceof HTMLInputElement
				? Array.from(this.tableBody.children).findIndex(row => row.contains(document.activeElement))
				: -1;
		this.tableBody.innerHTML = ``;

		EpWeightsMenu.epUnitStats.forEach(stat => {
			// Don't show extra stats when 'Show all stats' is not selected
			// A stat the spec doesn't weigh still shows when the Current EP prices it, as the
			// PvP presets do with Stamina and Armor.
			const hasCurrentEp = this.simUI.player.getEpWeights().getUnitStat(stat) != 0;
			const isExtra = !hasCurrentEp && stat.isStat() && !this.epStats.includes(stat.getStat());
			if ((!this.showAllStats && isExtra) || (stat.isPseudoStat() && !this.epPseudoStats.includes(stat.getPseudoStat()))) {
				return;
			}
			const row = this.makeTableRow(stat);
			// The stats only Show All Stats lists have their name in gray.
			row.classList.toggle('ep-extra-stat', isExtra);
			this.tableBody.appendChild(row);
		});

		if (focusedRow >= 0) this.tableBody.children[focusedRow]?.querySelector('input')?.focus();
		this.applyAllButton.disabled = !EpWeightsMenu.epUnitStats.some(stat => this.differsFromCurrent(stat));
	}

	private makeTableRow(stat: UnitStat): HTMLElement {
		const row = document.createElement('tr');
		// Stats the calculation doesn't weigh (like Stamina, which the PvP presets price by
		// hand) leave the Calculated EP empty, not a 0 that looks like a result.
		const result = this.isCalculated(stat) ? this.simUI.prevEpSimResult : null;
		const epRatios = this.simUI.player.getEpRatios();
		const rowTotalEp = scaledEpValue(stat, epRatios, result);
		row.innerHTML = `
			<td>${stat.getName(this.simUI.player.getClass())}</td>
			<td class="current-ep"></td>
			${this.makeTableRowCells(stat, result?.dps, 'damage-metrics', rowTotalEp, epRatios[0])}
			${this.makeTableRowCells(stat, result?.hps, 'healing-metrics', rowTotalEp, epRatios[1])}
			${this.makeTableRowCells(stat, result?.tps, 'threat-metrics', rowTotalEp, epRatios[2])}
			${this.makeTableRowCells(stat, result?.dtps, 'threat-metrics', rowTotalEp, epRatios[3])}
			${this.makeTableRowCells(stat, result?.tmi, 'threat-metrics experimental', rowTotalEp, epRatios[4])}
			${this.makeTableRowCells(stat, result?.pDeath, 'threat-metrics experimental', rowTotalEp, epRatios[5])}
		`;

		const currentEpCell = row.querySelector('.current-ep') as HTMLElement;
		const currentEpPicker = new NumberPicker(currentEpCell, this.simUI.player, {
			id: `ep-weight-stat-${stat}`,
			float: true,
			changedEvent: (player: Player<any>) => player.epWeightsChangeEmitter,
			getValue: (_player: Player<any>) => this.simUI.player.getEpWeights().getUnitStat(stat),
			setValue: (eventID: EventID, player: Player<any>, newValue: number) => {
				const epWeights = player.getEpWeights().withUnitStat(stat, newValue);
				player.setEpWeights(eventID, epWeights);
			},
		});

		// The arrow inside the field copies the calculated value into the Current EP. It only
		// shows when they differ.
		const calculated = this.calculatedValue(stat);
		if (calculated !== undefined && this.differsFromCurrent(stat)) {
			const applyButton = document.createElement('button');
			applyButton.className = 'btn btn-link ep-apply-row';
			applyButton.innerHTML = '<i class="fas fa-arrow-left"></i>';
			tippy(applyButton, { content: `Use the calculated ${calculated.toFixed(2)} as the Current EP` });
			applyButton.addEventListener('click', () => {
				const player = this.simUI.player;
				player.setEpWeights(TypedEvent.nextEventID(), player.getEpWeights().withUnitStat(stat, roundEp(calculated)));
			});
			currentEpPicker.rootElem.appendChild(applyButton);
		}

		return row;
	}

	private makeTableRowCells(stat: UnitStat, statWeights: StatWeightValues | undefined, className: string, epTotal: number, epRatio: number): string {
		let weightCell, epCell;
		if (statWeights) {
			const weightAvg = stat.getProtoValue(statWeights.weights!);
			const weightStdev = stat.getProtoValue(statWeights.weightsStdev!);
			weightCell = this.makeTableCellContents(weightAvg, weightStdev);

			const epAvg = stat.getProtoValue(statWeights.epValues!);
			const epStdev = stat.getProtoValue(statWeights.epValuesStdev!);
			epCell = this.makeTableCellContents(epAvg, epStdev);
		} else {
			weightCell = `<span class="results-avg notapplicable"></span>`;
			epCell = weightCell;
		}

		const template = document.createElement('template');
		template.innerHTML = `
			<td class="stdev-cell ${className} type-weight">
				${weightCell}
			</td>
			<td class="stdev-cell ${className} type-ep">
				${epCell}
			</td>
		`;

		if (!statWeights) return template.innerHTML;

		if (epRatio == 0) {
			const cells = template.content.querySelectorAll('.stdev-cell');
			cells.forEach(cell => cell.classList.add('unused-ep'));
			return template.innerHTML;
		}

		const epCurrent = this.simUI.player.getEpWeights().getUnitStat(stat);
		const epDelta = epTotal - epCurrent;

		const epAvgElem = template.content.querySelector('.type-ep .results-avg') as HTMLElement;
		if (epDelta.toFixed(2) == '0.00') epAvgElem; // no-op
		else if (epDelta > 0) epAvgElem.classList.add('positive');
		else if (epDelta < 0) epAvgElem.classList.add('negative');

		return template.innerHTML;
	}

	private makeTableCellContents(value: number, stdev: number): string {
		const iterations = this.simUI.prevEpIterations || 1;
		return `
			<span class="results-avg">${value.toFixed(2)}</span>
			<span class="results-stdev">
				<i class="fas fa-plus-minus fa-xs"></i>${stDevToConf90(stdev, iterations).toFixed(2)}
			</span>
		`;
	}

	private calculateEp(weights: StatWeightsResult) {
		const result = StatWeightsResult.clone(weights);
		const normaliseValue = (refStat: Stat, values: StatWeightValues) => {
			const refUnitStat = UnitStat.fromStat(refStat);
			const refWeight = refUnitStat.getProtoValue(values.weights!);
			const refStdev = refUnitStat.getProtoValue(values.weightsStdev!);
			EpWeightsMenu.epUnitStats.forEach(stat => {
				const value = stat.getProtoValue(values.weights!);
				stat.setProtoValue(values.epValues!, refWeight == 0 ? 0 : value / refWeight);

				const valueStdev = stat.getProtoValue(values.weightsStdev!);
				stat.setProtoValue(values.epValuesStdev!, refStdev == 0 ? 0 : valueStdev / refStdev);
			});
		};

		if (this.simUI.dpsRefStat !== undefined) {
			normaliseValue(this.simUI.dpsRefStat, result.dps!);
			normaliseValue(this.simUI.dpsRefStat, result.tps!);
		}
		if (this.simUI.healRefStat !== undefined) normaliseValue(this.simUI.healRefStat, result.hps!);
		if (this.simUI.tankRefStat !== undefined) {
			normaliseValue(this.simUI.tankRefStat, result.dtps!);
			normaliseValue(this.simUI.tankRefStat, result.tmi!);
			normaliseValue(this.simUI.tankRefStat, result.pDeath!);
		}
		return result;
	}

	private getDpsEpRefStat(): Stat {
		return this.simUI.dpsRefStat !== undefined ? this.simUI.dpsRefStat : this.epReferenceStat;
	}

	private getHealEpRefStat(): Stat {
		return this.simUI.healRefStat !== undefined ? this.simUI.healRefStat : this.epReferenceStat;
	}

	private getTankEpRefStat(): Stat {
		return this.simUI.tankRefStat !== undefined ? this.simUI.tankRefStat : Stat.StatArmor;
	}

	private getPrevSimResult(): StatWeightsResult {
		return (
			this.simUI.prevEpSimResult ||
			StatWeightsResult.create({
				dps: {
					weights: new Stats().toProto(),
					weightsStdev: new Stats().toProto(),
					epValues: new Stats().toProto(),
					epValuesStdev: new Stats().toProto(),
				},
				hps: {
					weights: new Stats().toProto(),
					weightsStdev: new Stats().toProto(),
					epValues: new Stats().toProto(),
					epValuesStdev: new Stats().toProto(),
				},
				tps: {
					weights: new Stats().toProto(),
					weightsStdev: new Stats().toProto(),
					epValues: new Stats().toProto(),
					epValuesStdev: new Stats().toProto(),
				},
				dtps: {
					weights: new Stats().toProto(),
					weightsStdev: new Stats().toProto(),
					epValues: new Stats().toProto(),
					epValuesStdev: new Stats().toProto(),
				},
				tmi: {
					weights: new Stats().toProto(),
					weightsStdev: new Stats().toProto(),
					epValues: new Stats().toProto(),
					epValuesStdev: new Stats().toProto(),
				},
				pDeath: {
					weights: new Stats().toProto(),
					weightsStdev: new Stats().toProto(),
					epValues: new Stats().toProto(),
					epValuesStdev: new Stats().toProto(),
				},
			})
		);
	}

	private static epUnitStats: Array<UnitStat> = UnitStat.getAll().filter(stat => {
		if (stat.isStat()) {
			return true;
		} else {
			return [
				PseudoStat.PseudoStatMainHandDps,
				PseudoStat.PseudoStatOffHandDps,
				PseudoStat.PseudoStatRangedDps,
				PseudoStat.BonusPhysicalDamage,
				PseudoStat.PseudoStatMeleeSpeedMultiplier,
				PseudoStat.PseudoStatRangedSpeedMultiplier,
				PseudoStat.PseudoStatCastSpeedMultiplier,
			].includes(stat.getPseudoStat());
		}
	});
}
