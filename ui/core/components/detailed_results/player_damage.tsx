import tippy from 'tippy.js';

import { ActionMetrics, UnitMetrics } from '../../proto_utils/sim_result.js';
import { maxIndex, sum } from '../../utils.js';
import { ColumnSortType, MetricsTable } from './metrics_table/metrics_table.js';
import { MetricsTotalBar } from './metrics_table/metrics_total_bar';
import { ResultComponentConfig, SimResultData } from './result_component.js';
import { ResultsFilter } from './results_filter.js';
import { SourceChart } from './source_chart.js';

// Each player's damage, or with metric 'threat' each player's threat, as a share of the raid's.
export class PlayerDamageMetricsTable extends MetricsTable<UnitMetrics> {
	private readonly resultsFilter: ResultsFilter;
	private readonly metric: 'damage' | 'threat';

	// Cached values from most recent result.
	private raidDps: number;
	private maxDps: number;

	constructor(config: ResultComponentConfig, resultsFilter: ResultsFilter, metric: 'damage' | 'threat' = 'damage') {
		config.rootCssClass = metric == 'threat' ? 'player-threat-metrics-root' : 'player-damage-metrics-root';
		const perSecond = metric == 'threat' ? 'TPS' : 'DPS';
		super(config, [
			MetricsTable.playerNameCellConfig(),
			{
				name: 'Amount',
				tooltip: metric == 'threat' ? 'Player Threat / Raid Threat' : 'Player Damage / Raid Damage',
				headerCellClass: 'amount-header-cell text-center',
				fillCell: (player: UnitMetrics, cellElem: HTMLElement, rowElem: HTMLElement) => {
					cellElem.classList.add('amount-cell');

					let chart: HTMLElement | null = null;
					const makeChart = () => {
						const chartContainer = document.createElement('div');
						rowElem.appendChild(chartContainer);
						const playerActions = player
							.getPlayerAndPetActions()
							.map(action => action.forTarget(this.resultsFilter.getFilter()))
							.flat();
						new SourceChart(chartContainer, playerActions, this.metric);
						return chartContainer;
					};

					tippy(rowElem, {
						content: 'Loading...',
						placement: 'bottom',
						ignoreAttributes: true,
						onShow(instance: any) {
							if (!chart) {
								chart = makeChart();
								instance.setContent(chart);
							}
						},
					});

					const playerDps = this.getPlayerDps(player);

					cellElem.appendChild(
						<MetricsTotalBar
							classColor={player.classColor}
							max={this.maxDps}
							percentage={(playerDps / this.raidDps) * 100}
							value={playerDps}
							total={playerDps * this.getLastSimResult().result.duration}
						/>,
					);
				},
			},
			{
				name: perSecond,
				tooltip: metric == 'threat' ? 'Threat / Encounter Duration' : 'Damage / Encounter Duration',
				columnClass: 'dps-cell',
				sort: ColumnSortType.Descending,
				getValue: (player: UnitMetrics) => this.getPlayerDps(player),
				getDisplayString: (player: UnitMetrics) => this.getPlayerDps(player).toFixed(1),
			},
		]);
		this.resultsFilter = resultsFilter;
		this.metric = metric;
		this.raidDps = 0;
		this.maxDps = 0;
	}

	private getPlayerDps(player: UnitMetrics): number {
		const playerActions = player
			.getPlayerAndPetActions()
			.map(action => action.forTarget(this.resultsFilter.getFilter()))
			.flat();
		return sum(playerActions.map(action => this.perSecond(action)));
	}

	// Named for damage, the amounts here are threat with metric 'threat'.
	private perSecond(action: ActionMetrics): number {
		return this.metric == 'threat' ? action.tps : action.dps;
	}

	customizeRowElem(player: UnitMetrics, rowElem: HTMLElement) {
		rowElem.classList.add('player-damage-row');
		rowElem.addEventListener('click', event => {
			this.resultsFilter.setPlayer(this.getLastSimResult().eventID, player.index);
		});
	}

	getGroupedMetrics(resultData: SimResultData): Array<Array<UnitMetrics>> {
		const players = resultData.result.getPlayers(resultData.filter);

		const targetActions = players.map(player => player.getPlayerAndPetActions().map(action => action.forTarget(resultData.filter))).flat();

		this.raidDps = sum(targetActions.map(action => this.perSecond(action)));
		const maxDpsIndex = maxIndex(
			players.map(player => {
				const targetActions = player
					.getPlayerAndPetActions()
					.map(action => action.forTarget(resultData.filter))
					.flat();
				return sum(targetActions.map(action => this.perSecond(action)));
			}),
		)!;

		const maxDpsTargetActions = players[maxDpsIndex]
			.getPlayerAndPetActions()
			.map(action => action.forTarget(resultData.filter))
			.flat();
		this.maxDps = sum(maxDpsTargetActions.map(action => this.perSecond(action)));

		return players.map(player => [player]);
	}
}
