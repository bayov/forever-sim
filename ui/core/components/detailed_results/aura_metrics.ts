import { OtherAction } from '../../proto/common';
import { AuraMetrics } from '../../proto_utils/sim_result';
import { TypedEvent } from '../../typed_event';
import { ColumnSortType, MetricsTable } from './metrics_table/metrics_table';
import { ResultComponentConfig, SimResultData } from './result_component';

// An aura that never came up in the whole run, like the lower ranks of a spell we never cast.
const isUnused = (metric: AuraMetrics) => metric.averageProcs == 0 && metric.uptimePercent == 0;

export class AuraMetricsTable extends MetricsTable<AuraMetrics> {
	private readonly useDebuffs: boolean;

	showUnused = false;
	// How many rows we leave out because their auras never came up.
	unusedCount = 0;

	constructor(config: ResultComponentConfig, useDebuffs: boolean) {
		if (useDebuffs) {
			config.rootCssClass = 'debuff-metrics-root';
		} else {
			config.rootCssClass = 'buff-metrics-root';
		}
		super(config, [
			MetricsTable.nameCellConfig((metric: AuraMetrics) => {
				return {
					name: metric.name,
					actionId: metric.actionId,
					metricType: metric?.constructor?.name,
				};
			}),
			{
				name: 'Procs',
				getValue: (metric: AuraMetrics) => metric.averageProcs,
				getDisplayString: (metric: AuraMetrics) => metric.averageProcs.toFixed(2),
			},
			{
				name: 'PPM',
				getValue: (metric: AuraMetrics) => metric.ppm,
				getDisplayString: (metric: AuraMetrics) => metric.ppm.toFixed(2),
			},
			{
				name: 'Uptime',
				sort: ColumnSortType.Descending,
				columnClass: 'uptime-cell',
				getValue: (metric: AuraMetrics) => metric.uptimePercent,
				fillCell: (metric: AuraMetrics, cellElem: HTMLElement) => {
					const bar = document.createElement('div');
					bar.classList.add('uptime-bar');
					bar.style.setProperty('--uptime', `${Math.min(100, metric.uptimePercent)}%`);
					const value = document.createElement('span');
					value.textContent = metric.uptimePercent.toFixed(2) + '%';
					cellElem.append(bar, value);
				},
			},
		]);
		this.useDebuffs = useDebuffs;
	}

	setShowUnused(showUnused: boolean) {
		this.showUnused = showUnused;
		// The table skips events it has seen, so the redraw needs an event of its own.
		if (this.hasLastSimResult()) this.onSimResult({ ...this.getLastSimResult(), eventID: TypedEvent.nextEventID() });
	}

	getGroupedMetrics(resultData: SimResultData): Array<Array<AuraMetrics>> {
		const groups = this.getAllGroupedMetrics(resultData);
		const usedGroups = groups.filter(group => !group.every(isUnused));
		this.unusedCount = groups.length - usedGroups.length;
		return this.showUnused ? groups : usedGroups;
	}

	private getAllGroupedMetrics(resultData: SimResultData): Array<Array<AuraMetrics>> {
		if (this.useDebuffs) {
			return AuraMetrics.groupById(resultData.result.getDebuffMetrics(resultData.filter));
		} else {
			const players = resultData.result.getRaidIndexedPlayers(resultData.filter);
			if (players.length != 1) {
				return [];
			}
			const player = players[0];

			const auras = this.filterMetrics(player.auras);
			const actionGroups = AuraMetrics.groupById(auras);
			const petGroups = player.pets.map(pet => this.filterMetrics(pet.auras));
			return actionGroups.concat(petGroups);
		}
	}

	mergeMetrics(metrics: Array<AuraMetrics>): AuraMetrics {
		return AuraMetrics.merge(metrics, {
			removeTag: true,
			actionIdOverride: metrics[0].unit?.petActionId || undefined,
		});
	}

	shouldCollapse(metric: AuraMetrics): boolean {
		return !metric.unit?.isPet;
	}

	filterMetrics(metrics: Array<AuraMetrics>): Array<AuraMetrics> {
		return metrics.filter(aura => !(aura.unit?.isPet && aura.actionId.otherId === OtherAction.OtherActionMove));
	}
}
