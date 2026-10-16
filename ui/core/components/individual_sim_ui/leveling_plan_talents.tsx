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
	talentName,
	talentPointsAt,
	TalentShape,
	talentShape,
	talentsStringToCounts,
} from '../../leveling_plan';
import { Player } from '../../player';
import { Spec } from '../../proto/common';
import { LevelingPlanSettings, LevelingTalentBuild } from '../../proto/ui';
import { classTalentsConfig } from '../../talents/factory';
import { TalentsPicker } from '../../talents/talents_picker';
import { EventID, TypedEvent } from '../../typed_event';
import { ContentBlock } from '../content_block';

// The Talents part of the Leveling Plan tab: the builds we play on the way up, and the order we
// put each one's points in.
//
// The trees show the selected build with all its points. A click on a talent puts the next
// point in it, and a right click takes its last point out, so the order we click in is the order
// we level in. The list next to the trees has each point with the level we put it in, and the
// arrows move a point earlier or later. A build after the first is a respec at its level.
export class LevelingTalentsSection {
	readonly rootElem: HTMLElement;
	private readonly simUI: IndividualSimUI<Spec>;
	private readonly player: Player<Spec>;
	readonly shape: TalentShape;
	private selected = 0;
	private readonly selectedEmitter = new TypedEvent<void>('LevelingTalentsSelected');

	private readonly buildsElem: HTMLElement;
	private readonly orderElem: HTMLElement;
	private readonly orderTitleElem: HTMLElement;

	constructor(parentElem: HTMLElement, simUI: IndividualSimUI<Spec>) {
		this.simUI = simUI;
		this.player = simUI.player;
		const trees = classTalentsConfig[this.player.getClass()];
		this.shape = talentShape(trees);
		this.rootElem = (<div className="leveling-plan-talents" />) as HTMLElement;
		parentElem.appendChild(this.rootElem);

		const side = (<div className="leveling-plan-talents-side" />) as HTMLElement;
		const treesCol = (<div className="leveling-plan-talents-trees" />) as HTMLElement;
		const orderCol = (<div className="leveling-plan-talents-order-col" />) as HTMLElement;
		this.rootElem.append(side, treesCol, orderCol);

		const buildsBlock = new ContentBlock(side, 'leveling-plan-builds-block', {
			header: {
				title: 'Builds',
				tooltip:
					'The talent builds we level with. The first one starts at the first level of the plan, and each one after it is a respec at its level, where we put all our points back by its own order.',
			},
		});
		const buildsList = ref<HTMLDivElement>();
		const addRespec = ref<HTMLButtonElement>();
		const fromTalents = ref<HTMLButtonElement>();
		buildsBlock.bodyElement.appendChild(
			<>
				<div className="leveling-plan-builds" ref={buildsList} />
				<div className="leveling-plan-buttons">
					<button className="btn btn-outline-primary" ref={addRespec}>
						<i className="fas fa-plus me-1" />
						Add respec
					</button>
					<button className="btn btn-outline-primary" ref={fromTalents}>
						<i className="fas fa-file-import me-1" />
						From Talents tab
					</button>
				</div>
			</>,
		);
		this.buildsElem = buildsList.value!;
		tippy(addRespec.value!, { content: 'Respec 5 levels after the last build, into a copy of it that we can change.' });
		tippy(fromTalents.value!, {
			content: "Make the selected build the Talents tab's build, with its points in the order of the rows, lowest row first.",
		});
		addRespec.value!.addEventListener('click', () => this.addRespec());
		fromTalents.value!.addEventListener('click', () => this.loadFromTalentsTab());

		new TalentsPicker(treesCol, this.player, {
			klass: this.player.getClass(),
			trees,
			pointsPerRow: this.shape.pointsPerRow,
			tracksPresets: false,
			getMaxPoints: () => {
				const plan = this.plan();
				return talentPointsAt(plan.toLevel, plan.bonusTalentPoints);
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
					'Each point of the selected build, in the order we put them in, with the level we put it in. The arrows move a point earlier or later. A point in red breaks the tree, like one before its row is open.',
			},
		});
		const orderTitle = ref<HTMLParagraphElement>();
		const order = ref<HTMLDivElement>();
		orderBlock.bodyElement.appendChild(
			<>
				<p className="leveling-plan-note" ref={orderTitle} />
				<div className="leveling-plan-order" ref={order} />
			</>,
		);
		this.orderTitleElem = orderTitle.value!;
		this.orderElem = order.value!;

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

	private render() {
		const plan = this.plan();
		const builds = sortedBuilds(plan);
		if (this.selected >= builds.length) this.selected = Math.max(builds.length - 1, 0);
		this.renderBuilds(plan, builds);
		this.renderOrder(plan, builds);
	}

	private renderBuilds(plan: LevelingPlanSettings, builds: LevelingTalentBuild[]) {
		const rows = builds.map((build, i) => {
			const levels = buildLevels(plan, i);
			const talents = countsToTalentsString(countsOf(build.picks, this.shape));
			const level = ref<HTMLInputElement>();
			const remove = ref<HTMLButtonElement>();
			const row = (
				<div className={`leveling-plan-build ${i == this.selected ? 'active' : ''}`}>
					<div className="leveling-plan-build-labels">
						<span className="leveling-plan-build-name">
							{i == 0 ? `From level ${plan.fromLevel}` : 'Respec at level '}
							{i == 0 ? '' : <input className="form-control leveling-plan-level-input" type="number" value={String(build.level)} ref={level} />}
						</span>
						<span className="leveling-plan-build-where">
							{levels.length ? `Levels ${levels[0]} to ${levels[levels.length - 1]}, ` : 'No levels, '}
							{String(build.picks.length)} points, {talents || 'none'}
						</span>
					</div>
					{i == 0 ? null : (
						<button className="btn btn-link link-danger leveling-plan-remove" ref={remove}>
							<i className="fas fa-times" />
						</button>
					)}
				</div>
			) as HTMLElement;
			row.addEventListener('click', event => {
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
			return row;
		});
		this.buildsElem.replaceChildren(...(rows.length ? rows : [<p className="leveling-plan-note">No builds yet.</p>]));
	}

	private renderOrder(plan: LevelingPlanSettings, builds: LevelingTalentBuild[]) {
		const build = builds[this.selected];
		if (!build) {
			this.orderTitleElem.textContent = '';
			this.orderElem.replaceChildren();
			return;
		}
		const start = this.selected == 0 ? plan.fromLevel : build.level;
		const pointsAtEnd = talentPointsAt(plan.toLevel, plan.bonusTalentPoints);
		this.orderTitleElem.textContent =
			build.picks.length < pointsAtEnd
				? `${pointsAtEnd - build.picks.length} of the ${pointsAtEnd} points we have at level ${plan.toLevel} are still to place.`
				: `All ${pointsAtEnd} points we have at level ${plan.toLevel} are placed.`;
		const problems = pickProblems(build.picks, this.shape);
		const ranks = emptyCounts(this.shape);
		const rows = build.picks.map((pick, i) => {
			ranks[pick.tree][pick.talent]++;
			const config = this.shape.trees[pick.tree]?.talents[pick.talent];
			const level = pickLevel(i, start, plan.bonusTalentPoints);
			const up = ref<HTMLButtonElement>();
			const down = ref<HTMLButtonElement>();
			const row = (
				<div className={`leveling-plan-pick ${problems[i] ? 'leveling-plan-pick-problem' : ''}`}>
					<span className="leveling-plan-pick-level">{level > 60 ? '-' : String(level)}</span>
					<span className="leveling-plan-pick-name">
						{talentName(pick, this.shape)} {`${ranks[pick.tree][pick.talent]}/${config?.maxPoints ?? '?'}`}
					</span>
					<button className="btn btn-link leveling-plan-move" ref={up}>
						<i className="fas fa-arrow-up" />
					</button>
					<button className="btn btn-link leveling-plan-move" ref={down}>
						<i className="fas fa-arrow-down" />
					</button>
				</div>
			) as HTMLElement;
			up.value!.disabled = i == 0;
			down.value!.disabled = i == build.picks.length - 1;
			if (problems[i]) tippy(row, { content: problems[i]! });
			const move = (to: number) =>
				this.updateSelected(TypedEvent.nextEventID(), b => {
					const picks = [...b.picks];
					[picks[i], picks[to]] = [picks[to], picks[i]];
					b.picks = picks;
				});
			up.value!.addEventListener('click', () => move(i - 1));
			down.value!.addEventListener('click', () => move(i + 1));
			return row;
		});
		this.orderElem.replaceChildren(...(rows.length ? rows : [<p className="leveling-plan-note">Click the talents to put points in them.</p>]));
	}
}
