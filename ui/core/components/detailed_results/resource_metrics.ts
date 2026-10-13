import { ResourceType } from '../../proto/api';
import { resourceNames } from '../../proto_utils/names';
import { ResourceMetrics } from '../../proto_utils/sim_result';
import { orderedResourceTypes } from '../../proto_utils/utils';
import { ColumnSortType, MetricsTable } from './metrics_table/metrics_table';
import { ResultComponent, ResultComponentConfig, SimResultData } from './result_component';

export class ResourceMetricsTable extends ResultComponent {
	constructor(config: ResultComponentConfig) {
		config.rootCssClass = 'resource-metrics-root';
		super(config);

		orderedResourceTypes.forEach(resourceType => {
			// Each resource gets its own card, titled like the other panels of the results.
			const panelElem = document.createElement('div');
			panelElem.classList.add('content-block', 'dr-panel', 'resource-metrics-table-container', 'hide');
			panelElem.innerHTML = `<div class="content-block-header"><h6 class="content-block-title">${resourceNames.get(resourceType)}</h6></div>`;
			const containerElem = document.createElement('div');
			containerElem.classList.add('content-block-body');
			panelElem.appendChild(containerElem);
			this.rootElem.appendChild(panelElem);

			const childConfig = config;
			childConfig.parent = containerElem;
			const table = new TypedResourceMetricsTable(childConfig, resourceType);
			table.onUpdate.on(() => {
				panelElem.classList.toggle('hide', table.rootElem.classList.contains('hide'));
			});
		});
	}

	// eslint-disable-next-line @typescript-eslint/no-empty-function
	onSimResult() {}
}

export class TypedResourceMetricsTable extends MetricsTable<ResourceMetrics> {
	readonly resourceType: ResourceType;

	constructor(config: ResultComponentConfig, resourceType: ResourceType) {
		config.rootCssClass = 'resource-metrics-table-root';
		super(config, [
			MetricsTable.nameCellConfig((metric: ResourceMetrics) => {
				return {
					name: metric.name,
					actionId: metric.actionId,
					metricType: metric.constructor?.name,
				};
			}),
			{
				name: 'Casts',
				getValue: (metric: ResourceMetrics) => metric.events,
				getDisplayString: (metric: ResourceMetrics) => metric.events.toFixed(1),
			},
			{
				name: 'Gain',
				sort: ColumnSortType.Descending,
				getValue: (metric: ResourceMetrics) => metric.gain,
				getDisplayString: (metric: ResourceMetrics) => metric.gain.toFixed(1),
			},
			{
				name: 'Gain / s',
				getValue: (metric: ResourceMetrics) => metric.gainPerSecond,
				getDisplayString: (metric: ResourceMetrics) => metric.gainPerSecond.toFixed(1),
			},
			{
				name: 'Avg Gain',
				getValue: (metric: ResourceMetrics) => metric.avgGain,
				getDisplayString: (metric: ResourceMetrics) => metric.avgGain.toFixed(1),
			},
			{
				name: 'Wasted Gain',
				getValue: (metric: ResourceMetrics) => metric.wastedGain,
				getDisplayString: (metric: ResourceMetrics) => metric.wastedGain.toFixed(1),
			},
		]);
		this.resourceType = resourceType;
	}

	getGroupedMetrics(resultData: SimResultData): Array<Array<ResourceMetrics>> {
		const players = resultData.result.getRaidIndexedPlayers(resultData.filter);
		if (players.length != 1) {
			return [];
		}
		const player = players[0];

		const resources = player.getResourceMetrics(this.resourceType);
		const resourceGroups = ResourceMetrics.groupById(resources);
		return resourceGroups;
	}

	mergeMetrics(metrics: Array<ResourceMetrics>): ResourceMetrics {
		return ResourceMetrics.merge(metrics, {
			removeTag: true,
			actionIdOverride: metrics[0].unit?.petActionId || undefined,
		});
	}
}
