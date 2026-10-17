import tippy from 'tippy.js';
import { ref } from 'tsx-vanilla';

import { IndividualSimUI } from '../../individual_sim_ui';
import { countsOf, hoursAtLevel, pickLevel, planLevels, sortedBuilds, talentName, talentPointsAt, talentsAt } from '../../leveling_plan';
import { LevelHours, RouteSearchProgress, TalentRouteSearch } from '../../leveling_talent_search';
import { Player } from '../../player';
import { ErrorOutcomeType, RaidSimRequest } from '../../proto/api';
import { Spec } from '../../proto/common';
import { LevelingTalentBuild, LevelingTalentPick } from '../../proto/ui';
import { RequestTypes } from '../../sim_signal_manager';
import { TypedEvent } from '../../typed_event';
import { ContentBlock } from '../content_block';
import { keyTalents, LevelingTalentsSection, treeSplit } from './leveling_plan_talents';

export interface LevelRequestOptions {
	talents?: string;
	seed?: bigint;
	iterations?: number;
}

type MakeLevelRequest = (level: number, options: LevelRequestOptions) => RaidSimRequest | null;

// What a search found, as we keep it in the browser's storage.
interface SearchResult {
	time: number;
	fromLevel: number;
	toLevel: number;
	iterations: number;
	confirmIterations: number;
	sims: number;
	stopped?: boolean;
	// The best route without a respec, and the plan's own builds, over all the levels on the
	// confirm sessions.
	route: unknown[];
	routeHours: number;
	planHours: number;
	// The 95% noise of the hours the route saves over the plan.
	noise: number;
	respecs: Array<{
		level: number;
		route: unknown[];
		// Both over the levels from the respec on, on the confirm sessions.
		respecHours: number;
		keptHours: number;
		noise: number;
	}>;
}

const SEARCH_RESULT_KEY = '__levelingTalentSearch__';
const SEARCH_OPTIONS_KEY = '__levelingTalentSearchOptions__';
const DEFAULT_RESPEC_LEVELS = [20, 30, 40, 50];
const DEFAULT_SEARCH_SESSIONS = 50;
// A respec that saves less than this share of its levels' hours isn't worth the gold, so we
// don't mark it.
const BIG_SAVING = 0.02;

// The talent search of the Talents part (leveling_talent_search.ts, docs/leveling_plan.md).
//
// It finds the order of talent points that levels us the fastest over the whole plan, and at
// each respec level (20, 30, 40 and 50 by default) the build we'd respec into there. It tells us
// how many hours each saves, and we decide whether to take the route or a respec. Nothing in the
// plan changes until we click to take one.
export class LevelingTalentSearch {
	private readonly player: Player<Spec>;
	private readonly statusElem: HTMLElement;
	private readonly resultElem: HTMLElement;
	private readonly button: HTMLButtonElement;
	private readonly respecInput: HTMLInputElement;
	private readonly sessionsInput: HTMLInputElement;
	private search: TalentRouteSearch | null = null;
	private stopRequested = false;
	private result: SearchResult | null;

	constructor(
		parentElem: HTMLElement,
		private readonly simUI: IndividualSimUI<Spec>,
		private readonly talents: LevelingTalentsSection,
		private readonly makeRequest: MakeLevelRequest,
	) {
		this.player = simUI.player;
		const block = new ContentBlock(parentElem, 'leveling-plan-search-block', {
			header: {
				title: 'Talent search',
				tooltip:
					'Finds the order of talent points that levels us the fastest over all the levels of the plan, with its gear and rotations. It scores each order by the hours of all the levels together, so a point that does the most at level 10 but slows us later loses. At each respec level it also finds the best build to respec into there, and how many hours that saves over keeping the route. Nothing changes in the plan until we take a route or a respec.',
			},
		});
		// Under the builds, the part of the plan the search changes.
		parentElem.querySelector('.leveling-plan-builds-block')?.after(block.rootElem);
		const button = ref<HTMLButtonElement>();
		const respecs = ref<HTMLInputElement>();
		const sessions = ref<HTMLInputElement>();
		const status = ref<HTMLParagraphElement>();
		const result = ref<HTMLDivElement>();
		const options = this.loadOptions();
		block.bodyElement.appendChild(
			<>
				<div className="leveling-plan-search-controls">
					<label className="form-label">
						Respec levels
						<input className="form-control" type="text" value={options.respecLevels.join(', ')} ref={respecs} />
					</label>
					<label className="form-label">
						Sessions a sim
						<input className="form-control" type="number" min="10" value={String(options.sessions)} ref={sessions} />
					</label>
					<button className="btn btn-primary" ref={button} />
				</div>
				<p className="leveling-plan-note" ref={status} />
				<div className="leveling-plan-search-result" ref={result} />
			</>,
		);
		this.button = button.value!;
		this.respecInput = respecs.value!;
		this.sessionsInput = sessions.value!;
		this.statusElem = status.value!;
		this.resultElem = result.value!;
		tippy(this.respecInput.closest('label')!, {
			content: 'The levels we look for a respec at. Only the ones inside the plan count.',
		});
		tippy(this.sessionsInput.closest('label')!, {
			content:
				"The search sims each level with these many sessions. The route it finds and the plan are simmed again with the run's iterations, on other sessions, for the numbers we show.",
		});
		const saveOptions = () => this.storeOptions();
		this.respecInput.addEventListener('change', saveOptions);
		this.sessionsInput.addEventListener('change', saveOptions);
		this.button.addEventListener('click', () => (this.search ? this.stop() : this.run()));

		this.result = this.loadResult();
		TypedEvent.onAny([this.player.levelingPlanChangeEmitter]).on(() => this.render());
		this.render();
	}

	private respecLevels(): number[] {
		const levels = this.respecInput.value
			.split(/[^0-9]+/)
			.map(Number)
			.filter(level => level > 0);
		return [...new Set(levels)].sort((a, b) => a - b);
	}

	private sessions(): number {
		return Math.max(10, Math.round(Number(this.sessionsInput.value) || DEFAULT_SEARCH_SESSIONS));
	}

	private stop() {
		this.stopRequested = true;
		this.search?.stop();
		this.simUI.sim.signalManager.abortType(RequestTypes.RaidSim);
	}

	// Sims each level with its talents on its own sessions. A level's seed comes from its number,
	// so two talents strings at a level play the same mobs and the same rolls.
	private async simLevels(jobs: Array<{ level: number; talents: string }>, iterations: number, seed: bigint) {
		const results: LevelHours[] = new Array(jobs.length);
		let next = 0;
		const simOne = async () => {
			while (next < jobs.length && !this.stopRequested) {
				const i = next++;
				const { level, talents } = jobs[i];
				const request = this.makeRequest(level, { talents, iterations, seed: seed + BigInt(level) * 1_000_003n });
				if (!request) continue;
				const result = await this.simUI.sim.runRaidSimRequest(request, () => {}, false);
				if (result.error?.type == ErrorOutcomeType.ErrorOutcomeAborted) {
					this.stop();
					return;
				}
				const metrics = result.encounterMetrics?.leveling;
				if (!metrics) continue;
				results[i] = {
					hours: hoursAtLevel(level, metrics.xpPerHour),
					relativeError: metrics.mobsKilled > 0 ? metrics.mobsKilledStdev / Math.sqrt(iterations) / metrics.mobsKilled : 0,
				};
			}
		};
		// More sims in flight than workers, so a worker never waits for the next one.
		const inFlight = 2 * Math.max(2, this.simUI.sim.getWasmConcurrency());
		await Promise.all(Array.from({ length: inFlight }, simOne));
		return results;
	}

	private async run() {
		const plan = this.player.getLevelingPlan();
		const levels = planLevels(plan);
		const shape = this.talents.shape;
		if (!levels.length) return;
		const iterations = this.sessions();
		const seed = this.simUI.sim.makeRaidSimRequest(false).simOptions!.randomSeed;
		const pointsAt = (level: number) => talentPointsAt(level, plan.bonusTalentPoints);
		let stage = '';
		const search = new TalentRouteSearch(
			{ shape, levels, pointsAt, allowed: (tree, talent) => !shape.trees[tree].talents[talent].notSimulated },
			jobs => this.simLevels(jobs, iterations, seed),
			progress => this.showProgress(stage, progress),
		);
		this.search = search;
		this.stopRequested = false;
		this.render();
		try {
			const start = sortedBuilds(plan)[0]?.picks ?? [];
			stage = 'Route';
			let route = await search.search([start]);
			const respecs: Array<{ level: number; picks: LevelingTalentPick[] }> = [];
			for (const level of this.respecLevels().filter(level => level > plan.fromLevel && level < plan.toLevel)) {
				if (search.isStopped) break;
				stage = `Respec at ${level}`;
				const found = await search.search(
					[route.picks, start],
					levels.filter(l => l >= level),
				);
				respecs.push({ level, picks: found.picks });
			}
			// A respec's build is also a route from the first level, and it can be the better
			// route, like when it only moves a point the route could have put there too. So we
			// search the route again from them, and a respec then only saves what a respec does.
			if (respecs.length && !search.isStopped) {
				stage = 'Route again';
				route = await search.search([route.picks, ...respecs.map(respec => respec.picks)]);
			}

			// The numbers we show come from other sessions, at the run's iterations, so the search
			// can't have picked the route by the luck of its own sessions.
			const confirmIterations = this.simUI.sim.getIterations();
			const confirmSeed = seed + 7_777_777n;
			const jobs = new Map<string, { level: number; talents: string }>();
			const add = (level: number, talents: string) => jobs.set(`${level}|${talents}`, { level, talents });
			levels.forEach(level => {
				add(level, talentsAt(plan, level, shape));
				add(level, search.talentsAt(route.picks, level));
			});
			respecs.forEach(respec => levels.filter(l => l >= respec.level).forEach(level => add(level, search.talentsAt(respec.picks, level))));
			if (search.isStopped) {
				this.statusElem.textContent = 'Stopped.';
				return;
			}
			this.statusElem.textContent = `Simming the route${respecs.length ? ' and the respecs' : ''} again with ${confirmIterations} sessions a level...`;
			const list = [...jobs.values()];
			const results = await this.simLevels(list, confirmIterations, confirmSeed);
			if (this.stopRequested || results.some(r => !r)) {
				this.statusElem.textContent = 'Stopped.';
				return;
			}
			const byKey = new Map(list.map((job, i) => [`${job.level}|${job.talents}`, results[i]]));
			const at = (level: number, talents: string) => byKey.get(`${level}|${talents}`)!;
			// Both play the same sessions, so their noise mostly cancels, and adding it up is the
			// most it can be.
			const compare = (forLevels: number[], a: (level: number) => string, b: (level: number) => string) => {
				const hoursA = forLevels.reduce((sum, level) => sum + at(level, a(level)).hours, 0);
				const hoursB = forLevels.reduce((sum, level) => sum + at(level, b(level)).hours, 0);
				const variance = forLevels.reduce((sum, level) => {
					const x = at(level, a(level));
					const y = at(level, b(level));
					return sum + (x.hours * x.relativeError) ** 2 + (y.hours * y.relativeError) ** 2;
				}, 0);
				return { hoursA, hoursB, noise: 1.96 * Math.sqrt(variance) };
			};
			const main = compare(
				levels,
				level => search.talentsAt(route.picks, level),
				level => talentsAt(plan, level, shape),
			);
			this.result = {
				time: Date.now(),
				fromLevel: plan.fromLevel,
				toLevel: plan.toLevel,
				iterations,
				confirmIterations,
				sims: search.sims,
				route: route.picks.map(pick => LevelingTalentPick.toJson(pick)),
				routeHours: main.hoursA,
				planHours: main.hoursB,
				noise: main.noise,
				respecs: respecs.map(respec => {
					const c = compare(
						levels.filter(l => l >= respec.level),
						level => search.talentsAt(respec.picks, level),
						level => search.talentsAt(route.picks, level),
					);
					return {
						level: respec.level,
						route: respec.picks.map(pick => LevelingTalentPick.toJson(pick)),
						respecHours: c.hoursA,
						keptHours: c.hoursB,
						noise: c.noise,
					};
				}),
			};
			this.storeResult(this.result);
			this.statusElem.textContent = '';
		} catch (e) {
			this.simUI.handleCrash(e);
		} finally {
			this.search = null;
			this.render();
		}
	}

	private showProgress(stage: string, progress: RouteSearchProgress) {
		const best = isFinite(progress.bestHours) ? `, best so far ${formatHours(progress.bestHours)}` : '';
		this.statusElem.textContent = `${stage}: ${progress.stage.toLowerCase()}, ${progress.sims} sims of a level${best}.`;
	}

	private render() {
		this.button.replaceChildren(
			this.search ? <i className="fas fa-stop me-1" /> : <i className="fas fa-magnifying-glass me-1" />,
			this.search ? 'Stop' : 'Search',
		);
		this.button.classList.toggle('btn-danger', !!this.search);
		this.button.classList.toggle('btn-primary', !this.search);
		this.renderResult();
	}

	private renderResult() {
		const result = this.result;
		if (!result) {
			this.resultElem.replaceChildren();
			return;
		}
		const plan = this.player.getLevelingPlan();
		const shape = this.talents.shape;
		const route = result.route.map(pick => LevelingTalentPick.fromJson(pick as any));
		const samePlan = result.fromLevel == plan.fromLevel && result.toLevel == plan.toLevel;
		const saved = result.planHours - result.routeHours;
		const routeLine = (
			<p className={`leveling-plan-respec ${saved > result.noise ? 'leveling-plan-better' : ''}`}>
				{`Best route: ${formatHours(result.routeHours)} from ${result.fromLevel} to ${result.toLevel}, ${
					saved >= 0 ? `${formatHours(saved)} faster` : `${formatHours(-saved)} slower`
				} than the plan's talents (±${formatHours(result.noise)}).`}
			</p>
		);
		const useRoute = ref<HTMLButtonElement>();
		const elems: Element[] = [
			<p className="leveling-plan-note">
				{`Searched ${new Date(result.time).toLocaleString()}, ${result.sims} sims of a level with ${result.iterations} sessions each, then ${
					result.confirmIterations
				} sessions a level for the numbers below.`}
				{samePlan ? '' : ' The plan has other levels now, so search again.'}
			</p>,
			routeLine,
			this.routeElem(route, result.fromLevel, plan.bonusTalentPoints),
			<div className="leveling-plan-buttons">
				<button className="btn btn-outline-primary" ref={useRoute}>
					<i className="fas fa-check me-1" />
					Use this route
				</button>
			</div>,
		];
		tippy(useRoute.value!, { content: "Make this route the plan's starting build. The respecs stay as they are." });
		useRoute.value!.addEventListener('click', () => {
			const newPlan = this.player.getLevelingPlan();
			newPlan.builds = sortedBuilds(newPlan);
			if (newPlan.builds.length) newPlan.builds[0].picks = route;
			else newPlan.builds = [LevelingTalentBuild.create({ level: newPlan.fromLevel, picks: route })];
			this.player.setLevelingPlan(TypedEvent.nextEventID(), newPlan);
		});

		for (const respec of result.respecs) {
			const picks = respec.route.map(pick => LevelingTalentPick.fromJson(pick as any));
			const respecSaved = respec.keptHours - respec.respecHours;
			const big = respecSaved > respec.noise && respecSaved > BIG_SAVING * respec.keptHours;
			const points = talentPointsAt(respec.level, plan.bonusTalentPoints);
			const counts = countsOf(picks.slice(0, points), shape);
			const kept = countsOf(route.slice(0, points), shape);
			const add = ref<HTMLButtonElement>();
			elems.push(
				<div className={`leveling-plan-respec ${big ? 'leveling-plan-better' : ''}`}>
					<div>
						{`Respec at ${respec.level} into ${treeSplit(counts)}: ${respecSaved >= 0 ? 'saves' : 'costs'} ${formatHours(
							Math.abs(respecSaved),
						)} (±${formatHours(respec.noise)}) over levels ${respec.level} to ${result.toLevel - 1}, against keeping the route.`}
					</div>
					{this.changeElem(kept, counts)}
					{respecSaved > respec.noise ? (
						<button className="btn btn-link leveling-plan-add-respec" ref={add}>
							Add this respec to the plan
						</button>
					) : null}
				</div>,
			);
			add.value?.addEventListener('click', () => {
				const newPlan = this.player.getLevelingPlan();
				const builds = sortedBuilds(newPlan);
				newPlan.builds = [
					...builds.filter((build, i) => i == 0 || build.level != respec.level),
					LevelingTalentBuild.create({ level: respec.level, picks }),
				];
				this.player.setLevelingPlan(TypedEvent.nextEventID(), newPlan);
			});
		}
		if (!result.respecs.length) elems.push(<p className="leveling-plan-note">No respec level inside the plan's levels.</p>);
		this.resultElem.replaceChildren(...elems);
	}

	// The route's talents in the order we take them, a chip for each run of points in one talent.
	private routeElem(route: LevelingTalentPick[], fromLevel: number, bonusPoints: number): Element {
		const shape = this.talents.shape;
		const elem = (<div className="leveling-plan-search-route" />) as HTMLElement;
		let i = 0;
		while (i < route.length) {
			let j = i;
			while (j + 1 < route.length && route[j + 1].tree == route[i].tree && route[j + 1].talent == route[i].talent) j++;
			const rank = route.slice(0, j + 1).filter(pick => pick.tree == route[i].tree && pick.talent == route[i].talent).length;
			const first = pickLevel(i, fromLevel, bonusPoints);
			const last = pickLevel(j, fromLevel, bonusPoints);
			const icon = (<span className="leveling-plan-talent-icon small" />) as HTMLElement;
			this.talents.talentIcon(route[i].tree, route[i].talent, url => (icon.style.backgroundImage = `url('${url}')`));
			elem.appendChild(
				<span className="leveling-plan-talent-gain">
					{icon}
					<span>{`${talentName(route[i], shape)} ${rank}/${shape.trees[route[i].tree].talents[route[i].talent].maxPoints}`}</span>
					<span className="leveling-plan-search-levels">{first == last ? `${first}` : `${first}-${last}`}</span>
				</span>,
			);
			i = j + 1;
		}
		return elem;
	}

	// What a respec changes at its level: the talents it takes points out of and puts them in.
	private changeElem(before: number[][], after: number[][]): Element {
		const shape = this.talents.shape;
		const elem = (<div className="leveling-plan-search-route" />) as HTMLElement;
		const key = new Set(keyTalents(after, shape).map(pick => `${pick.tree},${pick.talent}`));
		shape.trees.forEach((tree, t) =>
			tree.talents.forEach((config, i) => {
				const change = after[t][i] - before[t][i];
				if (!change) return;
				const icon = (<span className="leveling-plan-talent-icon small" />) as HTMLElement;
				this.talents.talentIcon(t, i, url => (icon.style.backgroundImage = `url('${url}')`));
				elem.appendChild(
					<span className={`leveling-plan-talent-gain ${change < 0 ? 'leveling-plan-search-out' : ''} ${key.has(`${t},${i}`) ? 'key' : ''}`}>
						{icon}
						<span>{`${change > 0 ? '+' : ''}${change} ${config.name}`}</span>
					</span>,
				);
			}),
		);
		return elem;
	}

	private loadOptions(): { respecLevels: number[]; sessions: number } {
		try {
			const json = window.localStorage.getItem(this.simUI.getStorageKey(SEARCH_OPTIONS_KEY));
			const options = json ? JSON.parse(json) : null;
			if (options && Array.isArray(options.respecLevels))
				return { respecLevels: options.respecLevels, sessions: options.sessions || DEFAULT_SEARCH_SESSIONS };
		} catch (e) {
			console.warn('Failed to load the talent search options: ' + e);
		}
		return { respecLevels: DEFAULT_RESPEC_LEVELS, sessions: DEFAULT_SEARCH_SESSIONS };
	}

	private storeOptions() {
		try {
			window.localStorage.setItem(
				this.simUI.getStorageKey(SEARCH_OPTIONS_KEY),
				JSON.stringify({ respecLevels: this.respecLevels(), sessions: this.sessions() }),
			);
		} catch (e) {
			console.warn('Failed to save the talent search options: ' + e);
		}
	}

	private loadResult(): SearchResult | null {
		try {
			const json = window.localStorage.getItem(this.simUI.getStorageKey(SEARCH_RESULT_KEY));
			const result = json ? JSON.parse(json) : null;
			return result && Array.isArray(result.route) ? result : null;
		} catch (e) {
			console.warn('Failed to load the last talent search: ' + e);
			return null;
		}
	}

	private storeResult(result: SearchResult) {
		try {
			window.localStorage.setItem(this.simUI.getStorageKey(SEARCH_RESULT_KEY), JSON.stringify(result));
		} catch (e) {
			console.warn('Failed to save the talent search: ' + e);
		}
	}
}

export function formatHours(hours: number): string {
	if (!isFinite(hours)) return '-';
	const h = Math.floor(hours);
	const m = Math.round((hours - h) * 60);
	return m == 60 ? `${h + 1}h 0m` : `${h}h ${m}m`;
}
