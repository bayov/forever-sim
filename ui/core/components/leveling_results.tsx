import tippy from 'tippy.js';

import { LevelingMetrics } from '../proto/api';
import { SimResult } from '../proto_utils/sim_result';
import { formatToNumber } from '../utils';
import type { ResultMetric } from './raid_sim_action';

// The results of a leveling run (docs/leveling_sim.md), where each iteration is a grinding
// session. Its score is the mobs killed a session, in place of the DPS. Every result without
// leveling metrics (DPS and PvP runs) shows as before.

// The leveling metrics of a run, or undefined when it wasn't a leveling run.
export function getLevelingMetrics(simResult: SimResult): LevelingMetrics | undefined {
	return simResult.result.encounterMetrics?.leveling;
}

// The standard error of the mobs killed a session, the ± of the score.
export function mobsKilledError(leveling: LevelingMetrics, iterations: number): number {
	return leveling.mobsKilledStdev / Math.sqrt(Math.max(iterations, 1));
}

// The parts of a session, in the order they're drawn.
const SESSION_PARTS: Array<{ key: keyof LevelingMetrics; label: string; className: string; tooltip: string }> = [
	{ key: 'secondsInCombat', label: 'Combat', className: 'combat', tooltip: 'Fighting' },
	{ key: 'secondsResting', label: 'Resting', className: 'resting', tooltip: 'Eating, drinking and bandaging' },
	{ key: 'secondsSearching', label: 'Searching', className: 'searching', tooltip: 'Finding the next mob and pulling it' },
	{ key: 'secondsDead', label: 'Dead', className: 'dead', tooltip: 'Running back to the corpse' },
];

// Like 10:13.
function minutesAndSeconds(seconds: number): string {
	const rounded = Math.round(seconds);
	return `${Math.floor(rounded / 60)}:${String(rounded % 60).padStart(2, '0')}`;
}

// The headline results of a leveling run, for the cards of the Results tab and the compare
// tab. stdev stays the spread of a session, like the DPS's, so the Z tests work the same; the
// cards show the standard error of the mobs killed instead.
export function levelingColumns(leveling: LevelingMetrics, iterations: number): ResultMetric[] {
	return [
		{
			name: 'Mobs Killed',
			average: leveling.mobsKilled,
			stdev: leveling.mobsKilledStdev,
			error: mobsKilledError(leveling, iterations),
			digits: 2,
			note: 'a session',
			classes: 'results-sim-mobs',
		},
		{ name: 'XP/Hour', average: leveling.xpPerHour, digits: 0, classes: 'results-sim-xp' },
		{ name: 'Deaths', average: leveling.deaths, digits: 2, note: 'a session', classes: 'results-sim-deaths' },
		{ name: 'Mobs/Pull', average: leveling.mobsPerPull, digits: 2, classes: 'results-sim-mobs-per-pull' },
	];
}

// The results of a leveling run in the sidebar: the mobs killed a session leads, with its
// standard error, then XP an hour, deaths and mobs a pull, how the session's time split, and
// what we ate, drank and bandaged. The DPS (over the whole session, downtime included) comes
// last and small.
export function buildLevelingResults(leveling: LevelingMetrics, iterations: number, dps?: number): HTMLElement {
	const total = SESSION_PARTS.reduce((sum, part) => sum + (leveling[part.key] as number), 0) || 1;
	const parts = SESSION_PARTS.map(part => ({ ...part, seconds: leveling[part.key] as number })).filter(part => part.seconds > 0);
	const stat = (value: string, label: string, tooltip: string, className = '') => (
		<span className={`leveling-stat ${className}`} dataset={{ tippyContent: tooltip }}>
			<span className="leveling-stat-value">{value}</span> {label}
		</span>
	);

	const elem = (
		<div className="leveling-results">
			<div className="results-metric results-sim-mobs">
				<span className="topline-result-avg">{leveling.mobsKilled.toFixed(2)}</span>
				<span className="topline-result-stdev">
					(<i className="fas fa-plus-minus fa-xs"></i>
					{mobsKilledError(leveling, iterations).toFixed(2)})
				</span>
				<div className="results-reference hide">
					<span className="results-reference-diff"></span> vs ref
				</div>
			</div>
			<div className="leveling-stats">
				{stat(formatToNumber(leveling.xpPerHour, { maximumFractionDigits: 0 }), 'XP/h', 'XP an hour of grinding')}
				{stat(leveling.deaths.toFixed(2), 'deaths', 'Deaths a session')}
				{stat(leveling.mobsPerPull.toFixed(2), 'mobs/pull', 'Mobs a pull came to, chain pulls included')}
			</div>
			<div className="leveling-session">
				<div className="leveling-session-bar">
					{parts.map(part => (
						<span
							className={`leveling-session-part leveling-session-${part.className}`}
							style={{ width: `${(part.seconds / total) * 100}%` }}
							dataset={{
								tippyContent: `${part.tooltip}: ${minutesAndSeconds(part.seconds)} (${formatToNumber((part.seconds / total) * 100, {
									maximumFractionDigits: 0,
								})}%)`,
							}}
						/>
					))}
				</div>
				<div className="leveling-session-legend">
					{parts.map(part => (
						<span className="leveling-session-key" dataset={{ tippyContent: part.tooltip }}>
							<span className={`leveling-session-swatch leveling-session-${part.className}`} />
							{`${part.label} ${minutesAndSeconds(part.seconds)}`}
						</span>
					))}
				</div>
			</div>
			<div className="leveling-stats leveling-consumables">
				{stat(formatToNumber(leveling.food, { maximumFractionDigits: 1 }), 'food', 'Food eaten a session')}
				{stat(formatToNumber(leveling.drinks, { maximumFractionDigits: 1 }), 'drinks', 'Drinks a session')}
				{stat(formatToNumber(leveling.bandages, { maximumFractionDigits: 1 }), 'bandages', 'Bandages a session')}
				{dps !== undefined ? stat(dps.toFixed(1), 'DPS', 'Damage a second over the whole session, downtime included', 'leveling-dps') : undefined}
			</div>
		</div>
	) as HTMLElement;
	tippy(elem.querySelectorAll<HTMLElement>('[data-tippy-content]'));
	return elem;
}
