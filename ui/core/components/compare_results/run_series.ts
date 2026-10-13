import { IterationLog, RaidSimResult, ResourceType } from '../../proto/api';
import { SimLog } from '../../proto_utils/logs_parser';
import { SimResult, UnitMetrics } from '../../proto_utils/sim_result';

// The DPS, resources and threat of one iteration over the fight, like the Events tab's chart,
// for the Compare tab's Over Time view.
//
// We keep these instead of the iteration's logs. The logs of an iteration are about 250 KB, and
// these are a few KB: a value every half second. That's enough for a chart of the whole fight.
export interface IterationSeries {
	// In seconds.
	duration: number;
	dps: number;
	// The DPS of the last 15 seconds at each step, like the Events tab's DPS line.
	dpsOverTime: number[];
	threat: number[];
	resources: ResourceSeries[];
}

export interface ResourceSeries {
	type: ResourceType;
	// The full pool, for showing mana as a share of it.
	max: number;
	values: number[];
}

export type IterationKind = 'typical' | 'lucky' | 'unlucky';

export type RunSeries = Partial<Record<IterationKind, IterationSeries>>;

// The time between two values of a series.
export const SERIES_STEP = 0.5;

// The resources that go up and down over a fight. Health and combo points don't tell much here.
const RESOURCE_TYPES = [ResourceType.ResourceTypeMana, ResourceType.ResourceTypeEnergy, ResourceType.ResourceTypeRage];

// The series of the typical, lucky and unlucky iterations of a run. The run's own logs are
// the typical iteration's, and the lucky and unlucky ones come with their own logs, which we
// read here.
export async function buildRunSeries(result: SimResult): Promise<RunSeries> {
	const raw = result.result;
	const series: RunSeries = {};
	const typical = iterationSeries(result, raw.typicalIteration, raw.firstIterationDuration);
	if (typical) series.typical = typical;

	for (const [kind, iteration] of [
		['lucky', raw.luckyIteration],
		['unlucky', raw.unluckyIteration],
	] as Array<[IterationKind, IterationLog | undefined]>) {
		if (!iteration?.logs) continue;
		try {
			const iterationResult = await SimResult.makeNew(
				result.request,
				RaidSimResult.create({ ...raw, logs: iteration.logs, firstIterationDuration: iteration.duration }),
			);
			const built = iterationSeries(iterationResult, iteration, iteration.duration);
			if (built) series[kind] = built;
		} catch (e) {
			console.warn(`Failed to read the ${kind} iteration of a run: ` + e);
		}
	}
	return series;
}

function iterationSeries(result: SimResult, iteration: IterationLog | undefined, fallbackDuration: number): IterationSeries | null {
	const player = result.getFirstPlayer();
	if (!player || !player.dpsLogs.length) return null;
	const duration = iteration?.duration || fallbackDuration || result.encounterMetrics.durationSeconds;
	return {
		duration,
		dps: iteration?.dps ?? player.dps.avg,
		dpsOverTime: sample(player.dpsLogs, duration, log => log.dps, 0),
		threat: sample(player.threatLogs, duration, log => log.threatAfter, 0),
		resources: resourceSeries(player, duration),
	};
}

function resourceSeries(player: UnitMetrics, duration: number): ResourceSeries[] {
	return RESOURCE_TYPES.flatMap(type => {
		const logs = player.groupedResourceLogs[type] ?? [];
		if (!logs.length) return [];
		// The full pool is the most we ever had, prepull included, like the Events tab's mana axis.
		const max = Math.max(...logs.map(log => Math.max(log.valueBefore, log.valueAfter)));
		return [{ type, max, values: sample(logs, duration, log => log.valueAfter, logs[0].valueBefore) }];
	});
}

// The value at each step from 0 to the end of the fight: the value after the last change at or
// before it. Prepull changes count for the value at 0.
function sample<T extends SimLog>(logs: T[], duration: number, value: (log: T) => number, initial: number): number[] {
	const values: number[] = [];
	let current = initial;
	let i = 0;
	const steps = Math.floor(duration / SERIES_STEP + 1e-9);
	for (let step = 0; step <= steps; step++) {
		const time = step * SERIES_STEP;
		while (i < logs.length && logs[i].timestamp <= time) current = value(logs[i++]);
		values.push(Math.round(current * 10) / 10);
	}
	return values;
}
