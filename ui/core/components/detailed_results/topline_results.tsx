import { Spec } from '../../proto/common.js';
import { formatToNumber } from '../../utils.js';
import { RaidSimResultsManager } from '../raid_sim_action';
import { ResultComponent, ResultComponentConfig, SimResultData } from './result_component';

export class ToplineResults extends ResultComponent {
	constructor(config: ResultComponentConfig) {
		config.rootCssClass = 'topline-results-root';
		super(config);

		this.rootElem.classList.add('results-sim');
	}

	onSimResult(resultData: SimResultData) {
		const noManaSpecs = [Spec.SpecFeralTankDruid, Spec.SpecRogue, Spec.SpecWarrior, Spec.SpecTankWarrior];
		const players = resultData.result.getRaidIndexedPlayers(resultData.filter);

		const content = RaidSimResultsManager.makeToplineResultsContent(resultData.result, resultData.filter, {
			showOutOfMana: players.length === 1 && !noManaSpecs.includes(players[0].spec),
			asCards: true,
		});

		// The last card says how big the run was, so we know how much to trust the numbers.
		const result = resultData.result.result;
		content.appendChild(
			<div className="results-card results-card-run">
				<div className="results-card-label">Iterations</div>
				<div className="topline-result-avg">{formatToNumber(result.iterationsDone || resultData.result.iterations)}</div>
				<div className="results-card-details">
					<span>{formatToNumber(resultData.result.duration, { maximumFractionDigits: 1 })} s average fight</span>
				</div>
			</div>,
		);

		this.rootElem.replaceChildren(content);
	}
}
