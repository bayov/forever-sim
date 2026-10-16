import tippy from 'tippy.js';
import { ref } from 'tsx-vanilla';

import { IndividualSimUI } from '../../individual_sim_ui';
import {
	buildLevels,
	countsOf,
	countsToTalentsString,
	emptyCounts,
	pickLevel,
	pickProblems,
	picksAfterEdit,
	picksToward,
	sortedBuilds,
	talentPointsAt,
	TalentShape,
	talentShape,
	talentsStringToCounts,
} from '../../leveling_plan';
import { Player } from '../../player';
import { Spec } from '../../proto/common';
import { LevelingPlanSettings, LevelingTalentBuild, LevelingTalentPick } from '../../proto/ui';
import { ActionId } from '../../proto_utils/action_id';
import { getSpecIcon } from '../../proto_utils/utils';
import { classTalentsConfig } from '../../talents/factory';
import { talentIconUrl, TalentsConfig, TalentsPicker } from '../../talents/talents_picker';
import { EventID, TypedEvent } from '../../typed_event';
import { ContentBlock } from '../content_block';

// The colors of a class's three trees, in their order, for the order list and the builds.
const TREE_COLORS = ['#5b8ff0', '#e8892b', '#45b86b'];

// A run of picks in a row of a build's order that go into the same talent.
interface PickRun {
	tree: number;
	talent: number;
	// The first pick of the run in the build's order.
	start: number;
	// The level and rank of each pick of the run.
	points: Array<{ level: number; rank: number; problem: string | null }>;
}

// The Talents part of the Leveling Plan tab: the builds we play on the way up, and the order we
// put each one's points in.
//
// The builds are on top, with the levels each one plays. The trees show the selected build with
// all its points. A click on a talent puts the next point in it, and a right click takes its
// last point out, so the order we click in is the order we level in. Each talent shows the level
// its first point comes at. The order on the right has a row for each talent we put points in
// one after the other, with the level of each point, and the arrows move the row earlier or
// later. A build after the first is a respec at its level.
export class LevelingTalentsSection {
	readonly rootElem: HTMLElement;
	private readonly player: Player<Spec>;
	private readonly trees: TalentsConfig<any>;
	readonly shape: TalentShape;
	private selected = 0;
	private readonly selectedEmitter = new TypedEvent<void>('LevelingTalentsSelected');
	private readonly picker: TalentsPicker<any>;

	private readonly timelineElem: HTMLElement;
	private readonly buildsElem: HTMLElement;
	private readonly orderElem: HTMLElement;
	private readonly orderTitleElem: HTMLElement;
	private readonly orderFootElem: HTMLElement;
	// Each talent's icon, by tree and talent, once we've looked them up.
	private readonly icons = new Map<string, string>();

	constructor(parentElem: HTMLElement, simUI: IndividualSimUI<Spec>) {
		this.player = simUI.player;
		this.trees = classTalentsConfig[this.player.getClass()];
		this.shape = talentShape(this.trees);
		this.rootElem = (<div className="leveling-plan-talents" />) as HTMLElement;
		parentElem.appendChild(this.rootElem);

		const buildsBlock = new ContentBlock(this.rootElem, 'leveling-plan-builds-block', {
			header: {
				title: 'Builds',
				tooltip:
					'The talent builds we level with. The first one starts at the first level of the plan, and each one after it is a respec at its level, where we put all our points back by its own order.',
			},
		});
		const timeline = ref<HTMLDivElement>();
		const buildsList = ref<HTMLDivElement>();
		const addRespec = ref<HTMLButtonElement>();
		const fromTalents = ref<HTMLButtonElement>();
		buildsBlock.bodyElement.appendChild(
			<>
				<div className="leveling-plan-timeline" ref={timeline} />
				<div className="leveling-plan-builds">
					<div className="leveling-plan-build-cards" ref={buildsList} />
					<div className="leveling-plan-build-actions">
						<button className="btn btn-outline-primary" ref={addRespec}>
							<i className="fas fa-plus me-1" />
							Add respec
						</button>
						<button className="btn btn-outline-primary" ref={fromTalents}>
							<i className="fas fa-file-import me-1" />
							From Talents tab
						</button>
					</div>
				</div>
			</>,
		);
		this.timelineElem = timeline.value!;
		this.buildsElem = buildsList.value!;
		tippy(addRespec.value!, { content: 'Respec 5 levels after the last build, into a copy of it that we can change.' });
		tippy(fromTalents.value!, {
			content: "Make the selected build the Talents tab's build, with its points in the order of the rows, lowest row first.",
		});
		addRespec.value!.addEventListener('click', () => this.addRespec());
		fromTalents.value!.addEventListener('click', () => this.loadFromTalentsTab());

		const main = (<div className="leveling-plan-talents-main" />) as HTMLElement;
		this.rootElem.appendChild(main);
		const treesCol = (<div className="leveling-plan-talents-trees" />) as HTMLElement;
		const orderCol = (<div className="leveling-plan-talents-order-col" />) as HTMLElement;
		main.append(treesCol, orderCol);

		treesCol.appendChild(
			<p className="leveling-plan-note leveling-plan-trees-note">
				A click on a talent takes its next point at the end of the order, and a right click takes its last point out. The badge on a talent is the level
				of its first point.
			</p>,
		);
		this.picker = new TalentsPicker(treesCol, this.player, {
			klass: this.player.getClass(),
			trees: this.trees,
			pointsPerRow: this.shape.pointsPerRow,
			tracksPresets: false,
			getMaxPoints: () => {
				const plan = this.plan();
				return Math.max(talentPointsAt(plan.toLevel - 1, plan.bonusTalentPoints), this.selectedBuild()?.picks.length ?? 0);
			},
			changedEvent: () => TypedEvent.onAny([this.player.levelingPlanChangeEmitter, this.selectedEmitter]),
			getValue: () => countsToTalentsString(countsOf(this.selectedBuild()?.picks ?? [], this.shape)),
			setValue: (eventID: EventID, _: Player<Spec>, newValue: string) => {
				this.updateSelected(eventID, build => (build.picks = picksAfterEdit(build.picks, newValue, this.shape)));
			},
		});

		const orderBlock = new ContentBlock(orderCol, 'leveling-plan-order-block', {
			header: {
				title: 'Order',
				tooltip:
					'The talents of the selected build in the order we put points in them, with the level of each point. The arrows move a talent earlier or later. A point in red breaks the tree, like one before its row is open.',
			},
		});
		const orderTitle = ref<HTMLParagraphElement>();
		const order = ref<HTMLDivElement>();
		const orderFoot = ref<HTMLDivElement>();
		orderBlock.bodyElement.appendChild(
			<>
				<p className="leveling-plan-note" ref={orderTitle} />
				<div className="leveling-plan-order" ref={order} />
				<div className="leveling-plan-order-foot" ref={orderFoot} />
			</>,
		);
		this.orderTitleElem = orderTitle.value!;
		this.orderElem = order.value!;
		this.orderFootElem = orderFoot.value!;

		TypedEvent.onAny([this.player.levelingPlanChangeEmitter, this.selectedEmitter]).on(() => this.render());
		this.render();
	}

	private plan(): LevelingPlanSettings {
		return this.player.getLevelingPlan();
	}

	private selectedBuild(): LevelingTalentBuild | undefined {
		return sortedBuilds(this.plan())[this.selected];
	}

	// Changes the selected build. The builds are kept by level, so the change finds it there.
	private updateSelected(eventID: EventID, change: (build: LevelingTalentBuild) => void) {
		const plan = this.plan();
		plan.builds = sortedBuilds(plan);
		const build = plan.builds[this.selected];
		if (!build) return;
		change(build);
		this.player.setLevelingPlan(eventID, plan);
	}

	private select(index: number) {
		this.selected = index;
		this.selectedEmitter.emit(TypedEvent.nextEventID());
	}

	private addRespec() {
		const plan = this.plan();
		const builds = sortedBuilds(plan);
		const last = builds[builds.length - 1];
		const level = Math.min(Math.max(builds.length > 1 ? last.level + 5 : plan.fromLevel + 5, plan.fromLevel + 1), plan.toLevel - 1);
		plan.builds = [...builds, LevelingTalentBuild.create({ level, picks: last ? [...last.picks] : [] })];
		this.selected = plan.builds.length - 1;
		this.player.setLevelingPlan(TypedEvent.nextEventID(), plan);
	}

	private loadFromTalentsTab() {
		const target = talentsStringToCounts(this.player.getTalentsString(), this.shape);
		const eventID = TypedEvent.nextEventID();
		if (!this.plan().builds.length) {
			const plan = this.plan();
			plan.builds = [LevelingTalentBuild.create({ level: plan.fromLevel })];
			this.selected = 0;
			this.player.setLevelingPlan(eventID, plan);
		}
		this.updateSelected(eventID, build => (build.picks = picksToward(emptyCounts(this.shape), target, this.shape)));
	}

	private startLevel(plan: LevelingPlanSettings, index: number, build: LevelingTalentBuild): number {
		return index == 0 ? plan.fromLevel : build.level;
	}

	private render() {
		const plan = this.plan();
		const builds = sortedBuilds(plan);
		if (this.selected >= builds.length) this.selected = Math.max(builds.length - 1, 0);
		this.renderTimeline(plan, builds);
		this.renderBuilds(plan, builds);
		this.renderOrder(plan, builds);
		this.renderTreeBadges(plan, builds);
	}

	// A talent's icon, like the trees show it.
	talentIcon(tree: number, talent: number, onIcon: (url: string) => void) {
		const key = `${tree}-${talent}`;
		const known = this.icons.get(key);
		if (known) return onIcon(known);
		const config = this.trees[tree]?.talents[talent];
		if (!config) return;
		if (config.icon || !config.spellIds.some(id => id)) {
			const url = talentIconUrl(config.icon ?? 'inv_misc_questionmark');
			this.icons.set(key, url);
			return onIcon(url);
		}
		ActionId.fromSpellId(config.spellIds.find(id => id)!)
			.fill()
			.then(id => {
				this.icons.set(key, id.iconUrl);
				onIcon(id.iconUrl);
			});
	}

	private renderTimeline(plan: LevelingPlanSettings, builds: LevelingTalentBuild[]) {
		const segments = builds.flatMap((build, i) => {
			const levels = buildLevels(plan, i);
			if (!levels.length) return [];
			const counts = countsOf(build.picks.slice(0, talentPointsAt(levels[levels.length - 1], plan.bonusTalentPoints)), this.shape);
			const segment = (
				<button
					className={`leveling-plan-segment ${i == this.selected ? 'active' : ''}`}
					type="button"
					style={{ flexGrow: String(levels.length), '--segment-color': TREE_COLORS[mainTree(counts)] } as any}>
					<span className="leveling-plan-segment-levels">{levels.length == 1 ? `${levels[0]}` : `${levels[0]}-${levels[levels.length - 1]}`}</span>
					<span className="leveling-plan-segment-name">{i == 0 ? 'Starting build' : `Respec at ${build.level}`}</span>
					<span className="leveling-plan-segment-imbue">{treeSplit(counts)}</span>
				</button>
			) as HTMLButtonElement;
			segment.addEventListener('click', () => this.select(i));
			return [segment];
		});
		// One build plays every level, so a timeline would tell us nothing more than its card.
		this.timelineElem.replaceChildren(...(segments.length > 1 ? segments : []));
		this.timelineElem.classList.toggle('hide', segments.length < 2);
	}

	private renderBuilds(plan: LevelingPlanSettings, builds: LevelingTalentBuild[]) {
		const cards = builds.map((build, i) => {
			const levels = buildLevels(plan, i);
			const pointsAtEnd = levels.length ? talentPointsAt(levels[levels.length - 1], plan.bonusTalentPoints) : 0;
			const counts = countsOf(build.picks.slice(0, pointsAtEnd), this.shape);
			const level = ref<HTMLInputElement>();
			const remove = ref<HTMLButtonElement>();
			const trees = ref<HTMLDivElement>();
			const keys = ref<HTMLDivElement>();
			const card = (
				<div
					className={`leveling-plan-build ${i == this.selected ? 'active' : ''}`}
					style={{ '--segment-color': TREE_COLORS[mainTree(counts)] } as any}>
					<div className="leveling-plan-build-head">
						<span className="leveling-plan-build-name">
							{i == 0 ? (
								'Starting build'
							) : (
								<>
									Respec at <input className="form-control leveling-plan-level-input" type="number" value={String(build.level)} ref={level} />
								</>
							)}
						</span>
						{i == 0 ? null : (
							<button className="btn btn-link link-danger leveling-plan-remove" ref={remove}>
								<i className="fas fa-times" />
							</button>
						)}
					</div>
					<span className="leveling-plan-build-where">
						{levels.length ? `Levels ${levels[0]} to ${levels[levels.length - 1]}, ${pointsAtEnd} points` : 'Plays no levels'}
					</span>
					<div className="leveling-plan-build-trees" ref={trees} />
					<div className="leveling-plan-build-keys" ref={keys} />
				</div>
			) as HTMLElement;
			this.shape.trees.forEach((tree, t) => {
				const points = counts[t].reduce((sum, n) => sum + n, 0);
				const elem = (
					<span className={`leveling-plan-build-tree ${points ? '' : 'empty'}`}>
						<img src={getSpecIcon(this.player.getClass(), t)} />
						<span>{String(points)}</span>
					</span>
				) as HTMLElement;
				tippy(elem, { content: `${points} in ${tree.name}` });
				trees.value!.appendChild(elem);
			});
			// The talents of the two deepest rows we have, which say the most about a build.
			keyTalents(counts, this.shape).forEach(({ tree, talent }) => {
				const icon = (<span className="leveling-plan-talent-icon" />) as HTMLElement;
				this.talentIcon(tree, talent, url => (icon.style.backgroundImage = `url('${url}')`));
				const config = this.shape.trees[tree].talents[talent];
				tippy(icon, { content: `${config.name} ${counts[tree][talent]}/${config.maxPoints}` });
				keys.value!.appendChild(icon);
			});
			card.addEventListener('click', event => {
				if ((event.target as HTMLElement).closest('input, button')) return;
				this.select(i);
			});
			level.value?.addEventListener('change', () => {
				const newLevel = Math.round(Number(level.value!.value));
				if (!(newLevel > plan.fromLevel && newLevel < plan.toLevel)) {
					level.value!.value = String(build.level);
					return;
				}
				const newPlan = this.plan();
				newPlan.builds = sortedBuilds(newPlan);
				newPlan.builds[i].level = newLevel;
				this.player.setLevelingPlan(TypedEvent.nextEventID(), newPlan);
			});
			if (remove.value) {
				tippy(remove.value, { content: 'Drop this respec' });
				remove.value.addEventListener('click', () => {
					const newPlan = this.plan();
					newPlan.builds = sortedBuilds(newPlan).filter((_, j) => j != i);
					if (this.selected >= i) this.selected = Math.max(this.selected - 1, 0);
					this.player.setLevelingPlan(TypedEvent.nextEventID(), newPlan);
				});
			}
			return card;
		});
		this.buildsElem.replaceChildren(...(cards.length ? cards : [<p className="leveling-plan-note">No builds yet.</p>]));
	}

	// The picks of a build in runs of the same talent, with each pick's level and rank.
	private runsOf(plan: LevelingPlanSettings, build: LevelingTalentBuild, start: number): PickRun[] {
		const problems = pickProblems(build.picks, this.shape);
		const ranks = emptyCounts(this.shape);
		const runs: PickRun[] = [];
		build.picks.forEach((pick, i) => {
			ranks[pick.tree][pick.talent]++;
			const point = { level: pickLevel(i, start, plan.bonusTalentPoints), rank: ranks[pick.tree][pick.talent], problem: problems[i] };
			const last = runs[runs.length - 1];
			if (last && last.tree == pick.tree && last.talent == pick.talent) last.points.push(point);
			else runs.push({ tree: pick.tree, talent: pick.talent, start: i, points: [point] });
		});
		return runs;
	}

	// Swaps a run with the one before it, so its talent comes first.
	private moveRunUp(runs: PickRun[], index: number) {
		if (index <= 0) return;
		const a = runs[index - 1];
		const b = runs[index];
		this.updateSelected(TypedEvent.nextEventID(), build => {
			const picks = [...build.picks];
			const first = picks.slice(a.start, a.start + a.points.length);
			const second = picks.slice(b.start, b.start + b.points.length);
			picks.splice(a.start, first.length + second.length, ...second, ...first);
			build.picks = picks;
		});
	}

	private renderOrder(plan: LevelingPlanSettings, builds: LevelingTalentBuild[]) {
		const build = builds[this.selected];
		if (!build) {
			this.orderTitleElem.textContent = '';
			this.orderElem.replaceChildren();
			this.orderFootElem.replaceChildren();
			return;
		}
		const start = this.startLevel(plan, this.selected, build);
		const lastLevel = plan.toLevel - 1;
		const pointsAtEnd = talentPointsAt(lastLevel, plan.bonusTalentPoints);
		const placed = Math.min(build.picks.length, pointsAtEnd);
		this.orderTitleElem.textContent =
			placed < pointsAtEnd
				? `${placed} of the ${pointsAtEnd} points of level ${lastLevel} are placed. A click on a talent places the next one.`
				: `All ${pointsAtEnd} points of level ${lastLevel} are placed.`;

		const runs = this.runsOf(plan, build, start);
		const rows = runs.map((run, r) => {
			const config = this.shape.trees[run.tree]?.talents[run.talent];
			const icon = ref<HTMLSpanElement>();
			const pips = ref<HTMLSpanElement>();
			const up = ref<HTMLButtonElement>();
			const down = ref<HTMLButtonElement>();
			const inPlan = run.points.filter(point => point.level <= lastLevel);
			// The levels of the points we sim, or of all of them when the run comes after the plan.
			const shown = inPlan.length ? inPlan : run.points;
			const from = shown[0].level;
			const to = shown[shown.length - 1].level;
			const problem = run.points.find(point => point.problem)?.problem;
			const row = (
				<div
					className={`leveling-plan-run ${problem ? 'leveling-plan-run-problem' : ''} ${inPlan.length ? '' : 'leveling-plan-run-past'}`}
					style={{ '--tree-color': TREE_COLORS[run.tree] } as any}>
					<span className="leveling-plan-run-levels">{from > 60 ? '-' : from == to ? `${from}` : `${from}-${Math.min(to, 60)}`}</span>
					<span className="leveling-plan-talent-icon" ref={icon} />
					<span className="leveling-plan-run-labels">
						<span className="leveling-plan-run-name">{config?.name ?? 'Unknown talent'}</span>
						<span className="leveling-plan-pips" ref={pips} />
					</span>
					<span className="leveling-plan-run-moves">
						<button className="btn btn-link leveling-plan-move" ref={up}>
							<i className="fas fa-chevron-up" />
						</button>
						<button className="btn btn-link leveling-plan-move" ref={down}>
							<i className="fas fa-chevron-down" />
						</button>
					</span>
				</div>
			) as HTMLElement;
			this.talentIcon(run.tree, run.talent, url => (icon.value!.style.backgroundImage = `url('${url}')`));
			// A pip per rank of the talent. The ranks this run puts in are bright, and the ones it
			// already had or doesn't reach are dim.
			const firstRank = run.points[0].rank;
			for (let rank = 1; rank <= (config?.maxPoints ?? run.points.length); rank++) {
				const point = run.points.find(p => p.rank == rank);
				const pip = (<span className={`leveling-plan-pip ${point ? 'on' : rank < firstRank ? 'had' : ''}`} />) as HTMLElement;
				if (point) {
					pip.classList.toggle('past', point.level > lastLevel);
					tippy(pip, { content: point.level > 60 ? `Rank ${rank}, after level 60` : `Rank ${rank} at level ${point.level}` });
				}
				pips.value!.appendChild(pip);
			}
			if (problem) tippy(row, { content: problem });
			up.value!.disabled = r == 0;
			down.value!.disabled = r == runs.length - 1;
			tippy(up.value!, { content: 'Earlier' });
			tippy(down.value!, { content: 'Later' });
			up.value!.addEventListener('click', () => this.moveRunUp(runs, r));
			down.value!.addEventListener('click', () => this.moveRunUp(runs, r + 1));
			return row;
		});
		// A line where the plan ends, when the build has points after it.
		const past = runs.findIndex(run => run.points[0].level > lastLevel);
		if (past > 0) rows.splice(past, 0, (<div className="leveling-plan-order-end">{`After level ${lastLevel}`}</div>) as HTMLElement);
		this.orderElem.replaceChildren(...(rows.length ? rows : [<p className="leveling-plan-note">Click the talents in the trees to put points in them.</p>]));

		const extra = build.picks.length - pointsAtEnd;
		if (extra > 0) {
			const drop = ref<HTMLButtonElement>();
			this.orderFootElem.replaceChildren(
				<div className="leveling-plan-order-extra">
					<span>{`${extra} point${extra == 1 ? ' comes' : 's come'} after level ${lastLevel}, so we don't sim ${extra == 1 ? 'it' : 'them'}.`}</span>
					<button className="btn btn-outline-primary btn-sm" ref={drop}>
						Drop {extra == 1 ? 'it' : 'them'}
					</button>
				</div>,
			);
			drop.value!.addEventListener('click', () => this.updateSelected(TypedEvent.nextEventID(), b => (b.picks = b.picks.slice(0, pointsAtEnd))));
		} else {
			this.orderFootElem.replaceChildren();
		}
	}

	// Puts the level of each talent's first point of the selected build on its icon in the trees.
	private renderTreeBadges(plan: LevelingPlanSettings, builds: LevelingTalentBuild[]) {
		const build = builds[this.selected];
		const first = new Map<string, number>();
		if (build) {
			const start = this.startLevel(plan, this.selected, build);
			build.picks.forEach((pick, i) => {
				const key = `${pick.tree}-${pick.talent}`;
				if (!first.has(key)) first.set(key, pickLevel(i, start, plan.bonusTalentPoints));
			});
		}
		this.picker.trees.forEach((tree, t) =>
			tree.talents.forEach((talent, i) => {
				let badge = talent.rootElem.querySelector<HTMLElement>('.leveling-plan-talent-level');
				const level = first.get(`${t}-${i}`);
				if (level === undefined) {
					badge?.remove();
					return;
				}
				if (!badge) {
					badge = (<span className="leveling-plan-talent-level" />) as HTMLElement;
					talent.rootElem.appendChild(badge);
				}
				badge.textContent = level > 60 ? '-' : String(level);
				badge.classList.toggle('past', level > plan.toLevel - 1);
			}),
		);
	}
}

// The tree with the most points, the first one when none has any.
function mainTree(counts: number[][]): number {
	let best = 0;
	let bestPoints = -1;
	counts.forEach((tree, t) => {
		const points = tree.reduce((sum, n) => sum + n, 0);
		if (points > bestPoints) {
			best = t;
			bestPoints = points;
		}
	});
	return best;
}

// The points in each tree, like "0/21/0".
function treeSplit(counts: number[][]): string {
	return counts.map(tree => tree.reduce((sum, n) => sum + n, 0)).join('/');
}

// The talents with points in the two deepest rows of a build, deepest first.
function keyTalents(counts: number[][], shape: TalentShape): LevelingTalentPick[] {
	const talents: Array<{ tree: number; talent: number; row: number }> = [];
	counts.forEach((tree, t) =>
		tree.forEach((points, i) => {
			if (points > 0) talents.push({ tree: t, talent: i, row: shape.trees[t].talents[i].row });
		}),
	);
	const rows = [...new Set(talents.map(talent => talent.row))].sort((a, b) => b - a).slice(0, 2);
	return talents
		.filter(talent => rows.includes(talent.row))
		.sort((a, b) => b.row - a.row)
		.map(({ tree, talent }) => LevelingTalentPick.create({ tree, talent }));
}
