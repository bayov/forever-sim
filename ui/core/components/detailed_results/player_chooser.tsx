import { UnitMetrics } from '../../proto_utils/sim_result.js';
import { ResultComponent, ResultComponentConfig, SimResultData } from './result_component.js';
import { ResultsFilter } from './results_filter.js';

// The raid's players as buttons, for a tab that shows one player at a time.
//
// With All Players picked, the Buffs, Casts, Resources and Events views have nothing to show. We show this
// in their place, so picking a player is one click away instead of a trip to the filter in the toolbar.
export class PlayerChooser extends ResultComponent {
	private readonly resultsFilter: ResultsFilter;

	constructor(config: ResultComponentConfig, resultsFilter: ResultsFilter) {
		config.rootCssClass = 'player-chooser-root';
		super(config);
		this.resultsFilter = resultsFilter;
	}

	onSimResult(resultData: SimResultData) {
		const players = resultData.result.getPlayers();
		this.rootElem.replaceChildren(
			<div className="player-chooser-title">Pick a player</div>,
			<div className="player-chooser-list">{players.map(player => this.button(player, resultData))}</div>,
		);
	}

	private button(player: UnitMetrics, resultData: SimResultData) {
		const button = (
			<button className="player-chooser-item" type="button">
				<img className="metrics-action-icon" src={player.iconUrl} />
				<span className={`text-${player.classColor} text-truncate`}>{player.label}</span>
			</button>
		);
		button.addEventListener('click', () => this.resultsFilter.setPlayer(resultData.eventID, player.index));
		return button;
	}
}
