import tippy from 'tippy.js';

import { ActionMetrics, UnitMetrics } from '../../proto_utils/sim_result.js';
import { maxIndex, sum } from '../../utils.js';
import { TypedEvent } from '../../typed_event.js';
import { DamageMetricsTable } from './damage_metrics.js';
import { ColumnSortType, MetricsTable } from './metrics_table/metrics_table.js';
import { MetricsTotalBar } from './metrics_table/metrics_total_bar';
import { ResultComponentConfig, SimResultData } from './result_component.js';
import { ResultsFilter } from './results_filter.js';
import { ThreatMetricsTable } from './threat_metrics.js';

// Each player's damage, or with metric 'threat' each player's threat, as a share of the raid's.
export class PlayerDamageMetricsTable extends MetricsTable<UnitMetrics> {
	private readonly resultsFilter: ResultsFilter;
	private readonly metric: 'damage' | 'threat';
	// The players whose breakdown is open, by raid index, so a new result or filter keeps them open.
	private readonly expanded = new Set<number>();
	private readonly rows = new Map<number, HTMLElement>();

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
				fillCell: (player: UnitMetrics, cellElem: HTMLElement) => {
					cellElem.classList.add('amount-cell');

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
				fillCell: (player: UnitMetrics, cellElem: HTMLElement) => {
					const filterButton = (
						<button className="player-filter-button" type="button">
							<i className="fas fa-filter" />
						</button>
					) as HTMLButtonElement;
					tippy(filterButton, { content: 'Show only this player' });
					filterButton.addEventListener('click', event => {
						event.stopPropagation();
						this.resultsFilter.setPlayer(this.getLastSimResult().eventID, player.index);
					});
					cellElem.append(<span>{this.getPlayerDps(player).toFixed(1)}</span>, filterButton);
				},
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
		rowElem.firstElementChild?.prepend(<i className="fas fa-chevron-right player-expand-caret" />);
		this.rows.set(player.index, rowElem);
		rowElem.addEventListener('click', () => {
			if (this.expanded.has(player.index)) {
				this.expanded.delete(player.index);
				this.closeBreakdown(rowElem);
			} else {
				this.expanded.add(player.index);
				this.openBreakdown(player, rowElem);
			}
			this.resort();
		});
	}

	onSimResult(resultData: SimResultData) {
		super.onSimResult(resultData);
		this.expanded.forEach(index => {
			const rowElem = this.rows.get(index);
			if (rowElem) this.openBreakdown(this.getLastSimResult().result.getUnitWithIndex(index)!, rowElem);
		});
		if (this.expanded.size) this.resort();
	}

	// The player's own abilities table goes in a row under theirs, as the Abilities table it shows with
	// only that player picked.
	private openBreakdown(player: UnitMetrics, rowElem: HTMLElement) {
		const host = <div className="player-breakdown" />;
		const breakdownRow = (
			<tr className="child-metric player-breakdown-row">
				<td colSpan={rowElem.children.length}>{host}</td>
			</tr>
		);
		rowElem.after(breakdownRow);
		rowElem.classList.add('expanded');

		const emitter = new TypedEvent<SimResultData | null>();
		const config = { parent: host as HTMLElement, resultsEmitter: emitter };
		if (this.metric == 'threat') new ThreatMetricsTable(config);
		else new DamageMetricsTable(config);
		const last = this.getLastSimResult();
		emitter.emit(last.eventID, { ...last, filter: { ...last.filter, player: player.index } });
	}

	private closeBreakdown(rowElem: HTMLElement) {
		if (rowElem.nextElementSibling?.classList.contains('player-breakdown-row')) rowElem.nextElementSibling.remove();
		rowElem.classList.remove('expanded');
	}

	getGroupedMetrics(resultData: SimResultData): Array<Array<UnitMetrics>> {
		this.rows.clear();
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
