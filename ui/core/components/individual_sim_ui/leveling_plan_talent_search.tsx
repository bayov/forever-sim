import tippy from 'tippy.js';
import { ref } from 'tsx-vanilla';

import { IndividualSimUI } from '../../individual_sim_ui';
import { countsOf, pickLevel, planLevels, sortedBuilds, talentName, talentPointsAt, talentsAt, talentsStringToCounts } from '../../leveling_plan';
import { Player } from '../../player';
import { Spec } from '../../proto/common';
import {
	LevelingPlanSearchProgress,
	LevelingPlanSearchRequest,
	LevelingPlanSearchResult,
	LevelingPlanSearchUpdate,
	LevelingTalentBuild,
	LevelingTalentPick,
	LevelingTalentRequirement,
} from '../../proto/ui';
import { TypedEvent } from '../../typed_event';
import { ContentBlock } from '../content_block';
import { keyTalents, LevelingTalentsSection, treeSplit } from './leveling_plan_talents';

// The Go search's local server (tools/levelingplan -serve).
const SEARCH_SERVER = 'http://localhost:3001/search';

// What a search found, as we keep it in the browser's storage.
interface SearchResult {
	time: number;
	fromLevel: number;
	toLevel: number;
	iterations: number;
	confirmIterations: number;
	sims: number;
	// The seconds the search took, in searches since it runs in Go.
	seconds?: number;
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

// The talent search of the Talents part (tools/levelingplan, docs/leveling_plan.md).
//
// It finds the order of talent points that levels us the fastest over the whole plan, and at
// each respec level (20, 30, 40 and 50 by default) the build we'd respec into there. It tells us
// how many hours each saves, and we decide whether to take the route or a respec. Nothing in the
// plan changes until we click to take one.
export class LevelingTalentSearch {
	private readonly player: Player<Spec>;
	private readonly statusElem: HTMLElement;
	private readonly requiredElem: HTMLElement;
	private readonly addRequiredSelect: HTMLSelectElement;
	private readonly resultElem: HTMLElement;
	private readonly button: HTMLButtonElement;
	private readonly respecInput: HTMLInputElement;
	private readonly sessionsInput: HTMLInputElement;
	private abort: AbortController | null = null;
	private result: SearchResult | null;

	constructor(
		parentElem: HTMLElement,
		private readonly simUI: IndividualSimUI<Spec>,
		private readonly talents: LevelingTalentsSection,
		private readonly makeSearchRequest: () => LevelingPlanSearchRequest | null,
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
		const required = ref<HTMLDivElement>();
		const addRequired = ref<HTMLSelectElement>();
		const options = this.loadOptions();
		block.bodyElement.appendChild(
			<>
				<div className="leveling-plan-required-block">
					<span className="form-label leveling-plan-required-title">Required talents</span>
					<div className="leveling-plan-required" ref={required} />
					<select className="form-select leveling-plan-required-add" ref={addRequired} />
				</div>
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
		this.requiredElem = required.value!;
		this.addRequiredSelect = addRequired.value!;
		tippy(this.requiredElem.previousElementSibling!, {
			content:
				"Talents we take whatever the sim says, like Improved Ghost Wolf for walking to quests and grind spots, which a grinding session doesn't sim. Every route the search finds has their points by their level, and each respec has them from its level on.",
		});
		this.fillAddRequired();
		this.addRequiredSelect.addEventListener('change', () => {
			const [tree, talent] = this.addRequiredSelect.value.split(',').map(Number);
			this.addRequiredSelect.value = '';
			if (!isNaN(tree) && !isNaN(talent)) this.addRequirement(tree, talent);
		});
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
		this.button.addEventListener('click', () => (this.abort ? this.stop() : this.run()));

		this.result = this.loadResult();
		TypedEvent.onAny([this.player.levelingPlanChangeEmitter]).on(() => this.render());
		this.render();
	}

	// The talents we can add as required, by tree.
	private fillAddRequired() {
		const shape = this.talents.shape;
		this.addRequiredSelect.replaceChildren(
			<option value="">Add a required talent...</option>,
			...shape.trees.map((tree, t) => (
				<optgroup label={tree.name}>
					{tree.talents.map((talent, i) => (
						<option value={`${t},${i}`}>{talent.name}</option>
					))}
				</optgroup>
			)),
		);
	}

	// The earliest level a talent can have points by: its row must be open and the talent it
	// needs full.
	private earliestLevel(tree: number, talent: number, points: number): number {
		const plan = this.player.getLevelingPlan();
		const shape = this.talents.shape;
		const config = shape.trees[tree].talents[talent];
		const before = Math.max(config.row * shape.pointsPerRow, config.prereq !== undefined ? shape.trees[tree].talents[config.prereq].maxPoints : 0);
		return pickLevel(before + points - 1, plan.fromLevel, plan.bonusTalentPoints);
	}

	private addRequirement(tree: number, talent: number) {
		const plan = this.player.getLevelingPlan();
		const points = this.talents.shape.trees[tree].talents[talent].maxPoints;
		plan.requiredTalents = [
			...plan.requiredTalents.filter(r => r.tree != tree || r.talent != talent),
			LevelingTalentRequirement.create({ tree, talent, points, level: this.earliestLevel(tree, talent, points) }),
		];
		this.player.setLevelingPlan(TypedEvent.nextEventID(), plan);
	}

	private changeRequirement(index: number, change: (requirement: LevelingTalentRequirement) => void) {
		const plan = this.player.getLevelingPlan();
		const requirement = plan.requiredTalents[index];
		if (!requirement) return;
		change(requirement);
		this.player.setLevelingPlan(TypedEvent.nextEventID(), plan);
	}

	private renderRequired() {
		const plan = this.player.getLevelingPlan();
		const shape = this.talents.shape;
		const rows = plan.requiredTalents.map((requirement, index) => {
			const config = shape.trees[requirement.tree]?.talents[requirement.talent];
			if (!config) return null;
			const icon = (<span className="leveling-plan-talent-icon small" />) as HTMLElement;
			this.talents.talentIcon(requirement.tree, requirement.talent, url => (icon.style.backgroundImage = `url('${url}')`));
			const points = ref<HTMLSelectElement>();
			const level = ref<HTMLInputElement>();
			const remove = ref<HTMLButtonElement>();
			// The plan's own builds may not have it, like ones from before we required it.
			const missing = planLevels(plan).find(
				l =>
					l >= requirement.level &&
					(talentsStringToCounts(talentsAt(plan, l, shape), shape)[requirement.tree]?.[requirement.talent] ?? 0) < requirement.points,
			);
			const earliest = this.earliestLevel(requirement.tree, requirement.talent, requirement.points);
			const row = (
				<div className="leveling-plan-required-row">
					{icon}
					<span>{config.name}</span>
					<select className="form-select" ref={points}>
						{Array.from({ length: config.maxPoints }, (_, n) => (
							<option value={String(n + 1)} selected={n + 1 == requirement.points}>{`${n + 1}/${config.maxPoints}`}</option>
						))}
					</select>
					<span>by level</span>
					<input className="form-control leveling-plan-level-input" type="number" value={String(requirement.level)} ref={level} />
					<button className="btn btn-link link-danger leveling-plan-remove" ref={remove}>
						<i className="fas fa-times" />
					</button>
					{requirement.level < earliest ? (
						<span className="leveling-plan-search-warning">{`Its row opens at ${earliest} at the earliest.`}</span>
					) : missing !== undefined ? (
						<span className="leveling-plan-search-warning">{`The plan's talents don't have it at ${missing}.`}</span>
					) : null}
				</div>
			) as HTMLElement;
			points.value!.addEventListener('change', () => this.changeRequirement(index, r => (r.points = Number(points.value!.value))));
			level.value!.addEventListener('change', () => {
				const value = Math.round(Number(level.value!.value));
				if (value > 0) this.changeRequirement(index, r => (r.level = value));
				else level.value!.value = String(requirement.level);
			});
			tippy(remove.value!, { content: 'No longer require it' });
			remove.value!.addEventListener('click', () => {
				const newPlan = this.player.getLevelingPlan();
				newPlan.requiredTalents = newPlan.requiredTalents.filter((_, i) => i != index);
				this.player.setLevelingPlan(TypedEvent.nextEventID(), newPlan);
			});
			return row;
		});
		this.requiredElem.replaceChildren(...rows.filter((row): row is HTMLElement => !!row));
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
		this.abort?.abort();
	}

	// Sends the plan to the Go search (tools/levelingplan) and shows its progress until it
	// answers. The search runs natively on all the cores, which the browser's wasm sims were too
	// slow for.
	private async run() {
		const request = this.makeSearchRequest();
		if (!request) {
			this.statusElem.textContent = "The gear plan hasn't loaded yet.";
			return;
		}
		request.respecLevels = this.respecLevels();
		request.sessions = this.sessions();
		request.confirmSessions = this.simUI.sim.getIterations();
		const abort = new AbortController();
		this.abort = abort;
		this.render();
		this.statusElem.textContent = 'Starting the search...';
		try {
			const response = await fetch(SEARCH_SERVER, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(LevelingPlanSearchRequest.toJson(request)),
				signal: abort.signal,
			});
			if (!response.ok || !response.body) {
				this.statusElem.textContent = (await response.text()).trim() || `The search server answered ${response.status}.`;
				return;
			}
			const reader = response.body.getReader();
			const decoder = new TextDecoder();
			let buffer = '';
			for (;;) {
				const { done, value } = await reader.read();
				if (done) break;
				buffer += decoder.decode(value, { stream: true });
				const lines = buffer.split('\n');
				buffer = lines.pop() ?? '';
				for (const line of lines) {
					if (!line.trim()) continue;
					const update = LevelingPlanSearchUpdate.fromJsonString(line, { ignoreUnknownFields: true }).update;
					if (update.oneofKind == 'progress') this.showProgress(update.progress);
					else if (update.oneofKind == 'result') this.takeResult(update.result);
				}
			}
		} catch (e) {
			this.statusElem.textContent = abort.signal.aborted
				? 'Stopped.'
				: `The search runs in Go, and nothing answers at ${SEARCH_SERVER}. Start it with: go run --tags=with_db ./tools/levelingplan -serve localhost:3001`;
		} finally {
			this.abort = null;
			this.render();
		}
	}

	private takeResult(result: LevelingPlanSearchResult) {
		if (result.error) {
			this.statusElem.textContent = result.error;
			return;
		}
		this.result = {
			time: Date.now(),
			fromLevel: result.fromLevel,
			toLevel: result.toLevel,
			iterations: result.sessions,
			confirmIterations: result.confirmSessions,
			sims: result.sims,
			seconds: result.seconds,
			route: result.route.map(pick => LevelingTalentPick.toJson(pick)),
			routeHours: result.routeHours,
			planHours: result.planHours,
			noise: result.noise,
			respecs: result.respecs.map(respec => ({
				level: respec.level,
				route: respec.route.map(pick => LevelingTalentPick.toJson(pick)),
				respecHours: respec.respecHours,
				keptHours: respec.keptHours,
				noise: respec.noise,
			})),
		};
		this.storeResult(this.result);
		this.statusElem.textContent = '';
	}

	private showProgress(progress: LevelingPlanSearchProgress) {
		const best = progress.bestHours > 0 ? `, best so far ${formatHours(progress.bestHours)}` : '';
		this.statusElem.textContent = `${progress.stage}, ${progress.sims} sims of a level${best}.`;
	}

	private render() {
		this.button.replaceChildren(
			this.abort ? <i className="fas fa-stop me-1" /> : <i className="fas fa-magnifying-glass me-1" />,
			this.abort ? 'Stop' : 'Search',
		);
		this.button.classList.toggle('btn-danger', !!this.abort);
		this.button.classList.toggle('btn-primary', !this.abort);
		this.renderRequired();
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
				{`Searched ${new Date(result.time).toLocaleString()}, ${result.sims} sims of a level with ${result.iterations} sessions each${
					result.seconds ? ` in ${formatSeconds(result.seconds)}` : ''
				}, then ${result.confirmIterations} sessions a level for the numbers below.`}
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

function formatSeconds(seconds: number): string {
	return seconds < 90 ? `${Math.round(seconds)}s` : `${Math.round(seconds / 60)} min`;
}

export function formatHours(hours: number): string {
	if (!isFinite(hours)) return '-';
	const h = Math.floor(hours);
	const m = Math.round((hours - h) * 60);
	return m == 60 ? `${h + 1}h 0m` : `${h}h ${m}m`;
}
