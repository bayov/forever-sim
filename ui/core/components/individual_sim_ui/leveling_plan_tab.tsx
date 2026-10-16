import tippy from 'tippy.js';
import { ref } from 'tsx-vanilla';

import { IndividualSimUI } from '../../individual_sim_ui';
import { LEVELING_GEAR_SLOTS } from '../../leveling_gear';
import {
	BONUS_TALENT_POINTS_LEVEL,
	buildIndexAt,
	buildLevels,
	hoursAtLevel,
	LEVELING_PLAN_MAX_LEVEL,
	planLevels,
	rotationAt,
	sortedBuilds,
	talentsAt,
	TalentShape,
	talentsStringToCounts,
	xpToNextLevel,
} from '../../leveling_plan';
import { Player } from '../../player';
import { ErrorOutcomeType, ProgressMetrics, RaidSimRequest } from '../../proto/api';
import { APLRotation } from '../../proto/apl';
import { Consumes, Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs, Spec } from '../../proto/common';
import { LevelingPlanSettings } from '../../proto/ui';
import { slotNames } from '../../proto_utils/names';
import { Database } from '../../proto_utils/database';
import { Gear } from '../../proto_utils/gear';
import { encounterMode } from '../../sim_mode';
import { TypedEvent } from '../../typed_event';
import { ContentBlock } from '../content_block';
import { BooleanPicker } from '../boolean_picker';
import { NumberPicker } from '../number_picker';
import { SimTab } from '../sim_tab';
import { LevelingGearSection } from './leveling_plan_gear';
import { LevelingRotationsSection } from './leveling_plan_rotations';
import { LevelingTalentsSection } from './leveling_plan_talents';

interface SavedCharacter {
	level: number;
	bonusTalentPoints: number;
	talents: string;
	gear: Gear;
	rotation: APLRotation;
	specOptions: ReturnType<Player<Spec>['getSpecOptions']>;
}

// What one level of a plan run came to.
interface LevelResult {
	level: number;
	xpPerHour: number;
	mobsPerHour: number;
	deathsPerHour: number;
	// The share of the hours that is noise, from the mobs killed a session.
	relativeError: number;
	hours: number;
}

// A run of the plan: each level as the plan plays it, and the levels of each respec as we'd
// play them if we kept the build we had.
interface PlanRun {
	time: number;
	iterations: number;
	fromLevel: number;
	toLevel: number;
	levels: LevelResult[];
	// By the respec's build number in the plan (1 for the first respec).
	kept: Record<number, LevelResult[]>;
	respecs: Array<{ build: number; level: number }>;
	stopped?: boolean;
}

const PLAN_RUN_KEY = '__levelingPlanRun__';

type Section = 'overview' | 'gear' | 'talents' | 'rotations';
const SECTIONS: Array<{ id: Section; name: string }> = [
	{ id: 'overview', name: 'Overview' },
	{ id: 'gear', name: 'Gear' },
	{ id: 'talents', name: 'Talents' },
	{ id: 'rotations', name: 'Rotations' },
];

// The Leveling Plan tab, the page of the Leveling mode (sim_mode.ts, leveling_plan.ts).
//
// The plan says what we wear, which talents we have and which rotation we play at each level.
// The Overview has the levels we sim and the results of the last run, a row per level. Gear,
// Talents and Rotations each edit their part of the plan. In the Leveling mode this tab takes
// the place of the Gear, Talents and Rotation tabs, which edit one character at one level.
export class LevelingPlanTab extends SimTab {
	protected simUI: IndividualSimUI<Spec>;
	private readonly player: Player<Spec>;
	private readonly gear: LevelingGearSection;
	private readonly talents: LevelingTalentsSection;
	private readonly rotations: LevelingRotationsSection;
	private section: Section = 'overview';
	// The character as it was before the Rotations part made it the plan's at a level, so we
	// can put it back when we leave that part.
	private savedCharacter: SavedCharacter | null = null;
	private lastRun: PlanRun | null = null;
	private running: PlanRun | null = null;
	private progress: ProgressMetrics | null = null;
	private readonly sections = new Map<Section, HTMLElement>();
	private readonly navButtons = new Map<Section, HTMLButtonElement>();
	private readonly tableElem: HTMLElement;
	private readonly respecsElem: HTMLElement;
	private readonly summaryElem: HTMLElement;
	// Fires when the run's results change, so the sidebar shows them.
	readonly runChangeEmitter = new TypedEvent<void>('LevelingPlanRun');

	constructor(parentElem: HTMLElement, simUI: IndividualSimUI<Spec>) {
		super(parentElem, simUI, { identifier: 'leveling-plan-tab', title: 'Leveling Plan' });
		this.simUI = simUI;
		this.player = simUI.player;

		const nav = (<div className="leveling-plan-nav segmented" attributes={{ role: 'group', 'aria-label': 'Part of the plan' }} />) as HTMLElement;
		const body = (<div className="leveling-plan-body" />) as HTMLElement;
		this.contentContainer.appendChild(
			<div className="leveling-plan">
				{nav}
				{body}
			</div>,
		);
		for (const { id, name } of SECTIONS) {
			const button = (
				<button className="segmented-option" type="button" dataset={{ section: id }}>
					{name}
				</button>
			) as HTMLButtonElement;
			button.addEventListener('click', () => this.showSection(id));
			nav.appendChild(button);
			this.navButtons.set(id, button);
			const section = (<div className={`leveling-plan-section leveling-plan-section-${id}`} />) as HTMLElement;
			body.appendChild(section);
			this.sections.set(id, section);
		}

		// The overview: the plan's levels on the left, the results by level on the right.
		const overview = this.sections.get('overview')!;
		const side = (<div className="leveling-plan-side" />) as HTMLElement;
		const main = (<div className="leveling-plan-main" />) as HTMLElement;
		overview.append(side, main);
		const levelsBlock = new ContentBlock(side, 'leveling-plan-levels-block', {
			header: {
				title: 'Levels',
				tooltip:
					'We sim a grinding session at each level from the first level up to the last, and add up the hours each level takes. Quests, dungeons and travel are left out, so it is the time of a plan that only grinds.',
			},
		});
		const reset = ref<HTMLButtonElement>();
		const summary = ref<HTMLParagraphElement>();
		levelsBlock.bodyElement.appendChild(
			<>
				<p className="leveling-plan-summary" ref={summary} />
			</>,
		);
		this.summaryElem = summary.value!;
		const levelInput = (
			id: string,
			label: string,
			tooltip: string,
			get: (plan: LevelingPlanSettings) => number,
			set: (plan: LevelingPlanSettings, value: number) => void,
		) =>
			new NumberPicker<Player<Spec>>(levelsBlock.bodyElement, this.player, {
				id,
				label,
				labelTooltip: tooltip,
				positive: true,
				changedEvent: player => player.levelingPlanChangeEmitter,
				getValue: player => get(player.getLevelingPlan()),
				setValue: (eventID, player, newValue) => {
					const plan = player.getLevelingPlan();
					set(plan, newValue);
					player.setLevelingPlan(eventID, plan);
				},
			});
		levelInput(
			'leveling-plan-from',
			'From level',
			'The level we start the plan at.',
			plan => plan.fromLevel,
			(plan, value) => (plan.fromLevel = Math.min(Math.max(Math.round(value), 1), plan.toLevel - 1)),
		);
		levelInput(
			'leveling-plan-to',
			'To level',
			'The level the plan ends at. We sim each level before it.',
			plan => plan.toLevel,
			(plan, value) => (plan.toLevel = Math.min(Math.max(Math.round(value), plan.fromLevel + 1), LEVELING_PLAN_MAX_LEVEL)),
		);
		levelInput(
			'leveling-plan-bonus',
			`Extra talent points from ${BONUS_TALENT_POINTS_LEVEL}`,
			`The talent points Forever gives on top of one a level, from level ${BONUS_TALENT_POINTS_LEVEL} on.`,
			plan => plan.bonusTalentPoints,
			(plan, value) => (plan.bonusTalentPoints = Math.max(Math.round(value), 0)),
		);
		new BooleanPicker<Player<Spec>>(levelsBlock.bodyElement, this.player, {
			id: 'leveling-plan-settings-buffs',
			label: "Settings tab's buffs",
			labelTooltip:
				"Level with the Settings tab's buffs, debuffs and consumables. Without them we level solo, because the Settings tab usually has a raid's.",
			inline: true,
			changedEvent: player => player.levelingPlanChangeEmitter,
			getValue: player => player.getLevelingPlan().useSettingsBuffs,
			setValue: (eventID, player, newValue) => {
				const plan = player.getLevelingPlan();
				plan.useSettingsBuffs = newValue;
				player.setLevelingPlan(eventID, plan);
			},
		});
		levelsBlock.bodyElement.appendChild(
			<div className="leveling-plan-buttons">
				<button className="btn btn-outline-primary" ref={reset}>
					<i className="fas fa-rotate-left me-1" />
					Spec's plan
				</button>
			</div>,
		);
		tippy(reset.value!, { content: "Put back the spec's own talent builds, rotations and levels. The gear plan stays as it is." });
		reset.value!.addEventListener('click', () => {
			const plan = this.simUI.individualConfig.defaults.levelingPlan;
			if (plan) this.player.setLevelingPlan(TypedEvent.nextEventID(), plan);
		});

		const resultsBlock = new ContentBlock(main, 'leveling-plan-results-block', {
			header: {
				title: 'By Level',
				tooltip:
					'A row per level: the talent points and the gear we get there, the rotation we play, and what the last run measured. Hours is the XP to the next level over the XP an hour. Grind takes the level to the Grinding mode.',
			},
		});
		const table = ref<HTMLDivElement>();
		const respecs = ref<HTMLDivElement>();
		resultsBlock.bodyElement.appendChild(
			<>
				<div className="leveling-plan-respecs" ref={respecs} />
				<div className="leveling-plan-table-wrap" ref={table} />
			</>,
		);
		this.tableElem = table.value!;
		this.respecsElem = respecs.value!;

		this.gear = new LevelingGearSection(this.sections.get('gear')!, simUI, level => this.grind(level));
		this.talents = new LevelingTalentsSection(this.sections.get('talents')!, simUI);
		this.rotations = new LevelingRotationsSection(
			this.sections.get('rotations')!,
			simUI,
			level => this.applyLevel(level),
			() => this.savedCharacter?.rotation ?? this.player.aplRotation,
		);
		this.showSection('overview');
		simUI.sim.encounter.changeEmitter.on(() => this.syncRotationsEditing());

		this.lastRun = this.loadRun();
		simUI.levelingPlanRun = onProgress => this.run(onProgress);
		simUI.levelingPlanSummary = () => this.sidebarSummary();

		simUI.sim.waitForInit().then(() => {
			// Settings saved before the plan had talents and rotations get the spec's plan.
			const plan = this.player.getLevelingPlan();
			const defaults = this.simUI.individualConfig.defaults.levelingPlan;
			if (defaults && !plan.builds.length && !plan.rotations.length) this.player.setLevelingPlan(TypedEvent.nextEventID(), defaults);
			this.render();
			TypedEvent.onAny([this.player.levelingPlanChangeEmitter, this.gear.planChangeEmitter]).on(() => this.render());
		});
	}

	protected buildTabContent() {}

	private showSection(section: Section) {
		this.section = section;
		this.syncRotationsEditing();
		this.sections.forEach((elem, id) => elem.classList.toggle('hide', id != section));
		this.navButtons.forEach((button, id) => {
			button.classList.toggle('active', id == section);
			button.setAttribute('aria-pressed', String(id == section));
		});
	}

	// The request of one level: the plan's gear, talents and rotation there, in the grinding
	// session of the Encounter settings. With buildIndex we take that talent build instead of
	// the plan's, like the one we had before a respec. With seed, the sessions are those of
	// that seed, so a respec and the build we had play the same mobs and the same rolls.
	private makeRequest(level: number, buildIndex?: number, seed?: bigint): RaidSimRequest | null {
		const plan = this.player.getLevelingPlan();
		const gear = this.gear.gearAt(level);
		if (!gear) return null;
		const request = this.simUI.sim.makeRaidSimRequest(false);
		request.simOptions!.debugFirstIteration = false;
		if (seed !== undefined) request.simOptions!.randomSeed = seed;
		const player = request.raid!.parties[0].players[0];
		player.level = level;
		player.bonusTalentPoints = level >= BONUS_TALENT_POINTS_LEVEL ? plan.bonusTalentPoints : 0;
		player.talentsString = talentsAt(plan, level, this.talents.shape, buildIndex);
		player.equipment = gear.asSpec();
		player.database = Database.mergeSimDatabases(gear.toDatabase(), this.player.getItemSwapGear().toDatabase());
		if (!plan.useSettingsBuffs) {
			player.consumes = Consumes.create();
			player.buffs = IndividualBuffs.create();
			request.raid!.buffs = RaidBuffs.create();
			request.raid!.debuffs = Debuffs.create();
			request.raid!.parties.forEach(party => (party.buffs = PartyBuffs.create()));
		}
		const rotation = rotationAt(plan, level);
		if (rotation?.rotation) player.rotation = rotation.rotation;
		if (rotation?.imbue) {
			const spec = player.spec as any;
			const options = spec?.[spec.oneofKind]?.options;
			if (options && 'shamanImbue' in options) options.shamanImbue = rotation.imbue;
		}
		if (request.encounter?.leveling) request.encounter.leveling.plan = false;
		return request;
	}

	// Sims each level of the plan, then each respec's levels with the build we had before it.
	async run(onProgress: (progress: ProgressMetrics) => void) {
		const plan = this.player.getLevelingPlan();
		const builds = sortedBuilds(plan);
		const jobs: Array<{ level: number; kept?: number }> = planLevels(plan).map(level => ({ level }));
		const respecs = builds.slice(1).map((build, i) => ({ build: i + 1, level: build.level }));
		respecs.forEach(respec => buildLevels(plan, respec.build).forEach(level => jobs.push({ level, kept: respec.build })));

		const iterations = this.simUI.sim.getIterations();
		const run: PlanRun = { time: Date.now(), iterations, fromLevel: plan.fromLevel, toLevel: plan.toLevel, levels: [], kept: {}, respecs };
		this.running = run;
		this.progress = null;
		this.changed();
		const total = jobs.length * iterations;
		const seed = this.simUI.sim.makeRaidSimRequest(false).simOptions!.randomSeed;
		try {
			for (let i = 0; i < jobs.length; i++) {
				const job = jobs[i];
				const request = this.makeRequest(job.level, job.kept !== undefined ? job.kept - 1 : undefined, seed + BigInt(job.level) * 1_000_003n);
				if (!request) break;
				const result = await this.simUI.sim.runRaidSimRequest(request, (progress: ProgressMetrics) => {
					this.progress = ProgressMetrics.create({
						completedIterations: i * iterations + (progress.completedIterations ?? 0),
						totalIterations: total,
					});
					onProgress(this.progress);
				});
				if (result.error?.type == ErrorOutcomeType.ErrorOutcomeAborted) {
					run.stopped = true;
					break;
				}
				const metrics = result.encounterMetrics?.leveling;
				if (!metrics) break;
				const sessionHours = (request.encounter?.leveling?.sessionLength || 1200) / 3600;
				const levelResult: LevelResult = {
					level: job.level,
					xpPerHour: metrics.xpPerHour,
					mobsPerHour: metrics.mobsKilled / sessionHours,
					deathsPerHour: metrics.deaths / sessionHours,
					relativeError: metrics.mobsKilled > 0 ? metrics.mobsKilledStdev / Math.sqrt(Math.max(iterations, 1)) / metrics.mobsKilled : 0,
					hours: hoursAtLevel(job.level, metrics.xpPerHour),
				};
				if (job.kept === undefined) run.levels.push(levelResult);
				else (run.kept[job.kept] ??= []).push(levelResult);
				this.changed();
			}
			if (run.levels.length) {
				this.lastRun = run;
				this.storeRun(run);
			}
		} catch (e) {
			this.simUI.handleCrash(e);
		} finally {
			this.running = null;
			this.progress = null;
			this.changed();
		}
	}

	private changed() {
		this.render();
		this.runChangeEmitter.emit(TypedEvent.nextEventID());
		if (encounterMode(this.simUI.sim.encounter) == 'levelingPlan') this.simUI.resultsViewer.setContent(this.sidebarSummary());
	}

	private shownRun(): PlanRun | null {
		return this.running ?? this.lastRun;
	}

	// The Rotations part edits the character's own rotation, so while it's open in the Leveling
	// mode the character is the plan's at a level. Before that we keep the character as it was,
	// and we put it back when we leave the part or the mode.
	private syncRotationsEditing() {
		const editing = this.section == 'rotations' && encounterMode(this.simUI.sim.encounter) == 'levelingPlan';
		if (editing && !this.savedCharacter) {
			this.savedCharacter = {
				level: this.player.getLevel(),
				bonusTalentPoints: this.player.getBonusTalentPoints(),
				talents: this.player.getTalentsString(),
				gear: this.player.getGear(),
				rotation: APLRotation.clone(this.player.aplRotation),
				specOptions: this.player.getSpecOptions(),
			};
			this.rotations.setActive(true);
		} else if (!editing && this.savedCharacter) {
			this.rotations.setActive(false);
			const saved = this.savedCharacter;
			this.savedCharacter = null;
			const eventID = TypedEvent.nextEventID();
			TypedEvent.freezeAllAndDo(() => {
				this.player.setLevel(eventID, saved.level);
				this.player.setBonusTalentPoints(eventID, saved.bonusTalentPoints);
				this.player.setTalentsString(eventID, saved.talents);
				this.player.setGear(eventID, saved.gear);
				this.player.setAplRotation(eventID, saved.rotation);
				this.player.setSpecOptions(eventID, saved.specOptions);
			});
		}
	}

	// Takes a level of the plan to the Grinding mode: its level, gear, talents and rotation.
	grind(level: number) {
		this.applyLevel(level);
		this.simUI.rootElem.querySelector<HTMLButtonElement>('.sim-mode-picker [data-mode="leveling"]')?.click();
		this.simUI.simHeader.activateTab('gear-tab');
	}

	// Makes the character the plan's at a level: its level, gear, talents, rotation and imbue.
	private applyLevel(level: number) {
		const plan = this.player.getLevelingPlan();
		const gear = this.gear.gearAt(level);
		const rotation = rotationAt(plan, level);
		const eventID = TypedEvent.nextEventID();
		TypedEvent.freezeAllAndDo(() => {
			this.player.setLevel(eventID, level);
			this.player.setBonusTalentPoints(eventID, level >= BONUS_TALENT_POINTS_LEVEL ? plan.bonusTalentPoints : 0);
			this.player.setTalentsString(eventID, talentsAt(plan, level, this.talents.shape));
			if (gear) this.player.setGear(eventID, gear);
			if (rotation?.rotation) this.player.setAplRotation(eventID, rotation.rotation);
			if (rotation?.imbue) {
				const options = this.player.getSpecOptions() as any;
				if ('shamanImbue' in options) this.player.setSpecOptions(eventID, { ...options, shamanImbue: rotation.imbue });
			}
		});
	}

	private render() {
		this.renderSummary();
		this.renderRespecs();
		this.renderTable();
	}

	private totalHours(levels: LevelResult[]): number {
		return levels.reduce((sum, level) => sum + level.hours, 0);
	}

	private renderSummary() {
		const plan = this.player.getLevelingPlan();
		const run = this.shownRun();
		const levels = planLevels(plan).length;
		let text = `${levels} levels, ${plan.fromLevel} to ${plan.toLevel}.`;
		if (run?.levels.length) {
			text += ` The ${this.running ? 'run going now' : 'last run'} has ${run.levels.length} of them at ${run.iterations} sessions each, ${formatHours(
				this.totalHours(run.levels),
			)} of grinding.`;
		}
		this.summaryElem.textContent = text;
	}

	private renderRespecs() {
		const run = this.shownRun();
		const plan = this.player.getLevelingPlan();
		const rows = (run?.respecs ?? []).map(respec => {
			const kept = run!.kept[respec.build] ?? [];
			const played = run!.levels.filter(level => kept.some(k => k.level == level.level));
			if (!kept.length || !played.length) return <p className="leveling-plan-note">{`Respec at ${respec.level}: not simmed yet.`}</p>;
			const saved = this.totalHours(kept.slice(0, played.length)) - this.totalHours(played);
			// Both play the same sessions, so their noise mostly cancels, and adding it up is
			// the most it can be.
			const noise =
				1.96 * Math.sqrt([...kept.slice(0, played.length), ...played].reduce((sum, level) => sum + (level.hours * level.relativeError) ** 2, 0));
			const from = played[0].level;
			const to = played[played.length - 1].level;
			return (
				<p className={`leveling-plan-respec ${saved > noise ? 'leveling-plan-better' : saved < -noise ? 'leveling-plan-worse' : ''}`}>
					{`Respec at ${respec.level}: ${saved >= 0 ? 'saves' : 'costs'} ${formatHours(Math.abs(saved))} (±${formatHours(
						noise,
					)}) over levels ${from} to ${to}, against keeping the build we had.`}
				</p>
			);
		});
		if (!run && plan.builds.length > 1) rows.push(<p className="leveling-plan-note">Simulate to see what each respec saves.</p>);
		this.respecsElem.replaceChildren(...rows);
	}

	private renderTable() {
		const plan = this.player.getLevelingPlan();
		const run = this.shownRun();
		const gearPlan = this.gear.getPlan();
		const builds = sortedBuilds(plan);
		const shape = this.talents.shape;
		const table = (<table className="leveling-plan-table table table-sm" />) as HTMLTableElement;
		table.appendChild(
			<thead>
				<tr>
					<th>Level</th>
					<th>Talents</th>
					<th>Gear</th>
					<th>Rotation</th>
					<th className="text-end">XP / h</th>
					<th className="text-end">Mobs / h</th>
					<th className="text-end">Deaths / h</th>
					<th className="text-end">Hours</th>
					<th className="text-end">Total</th>
					<th />
				</tr>
			</thead>,
		);
		const tbody = document.createElement('tbody');
		let total = 0;
		let lastRotation = '';
		for (const level of planLevels(plan)) {
			const buildIndex = buildIndexAt(plan, level);
			const respec = buildIndex > 0 && builds[buildIndex].level == level;
			const before = level > plan.fromLevel && !respec ? talentsAt(plan, level - 1, shape) : '';
			const now = talentsAt(plan, level, shape);
			const talentsText = respec ? `Respec: ${now || 'none'}` : newTalents(before, now, shape);
			const levelGear = gearPlan?.levels[level - 1];
			const gearText =
				levelGear?.changed
					.filter(slot => LEVELING_GEAR_SLOTS.includes(slot))
					.map(slot => levelGear.gear[slot]?.item.name ?? slotNames.get(slot) ?? '')
					.join(', ') ?? '';
			const rotation = rotationAt(plan, level);
			const rotationName = rotation ? rotation.name || 'Rotation tab copy' : 'Rotation tab';
			const rotationText = rotationName != lastRotation ? rotationName : '';
			lastRotation = rotationName;
			const result = run?.levels.find(l => l.level == level);
			if (result) total += result.hours;
			const grind = ref<HTMLButtonElement>();
			const row = (
				<tr className={respec ? 'leveling-plan-respec-row' : ''}>
					<td>{String(level)}</td>
					<td className="leveling-plan-cell-text">{talentsText}</td>
					<td className="leveling-plan-cell-text">{gearText}</td>
					<td className="leveling-plan-cell-text">{rotationText}</td>
					<td className="text-end">{result ? Math.round(result.xpPerHour).toLocaleString() : ''}</td>
					<td className="text-end">{result ? result.mobsPerHour.toFixed(1) : ''}</td>
					<td className="text-end">{result ? result.deathsPerHour.toFixed(2) : ''}</td>
					<td className="text-end">{result ? result.hours.toFixed(2) : ''}</td>
					<td className="text-end">{result ? total.toFixed(1) : ''}</td>
					<td>
						<button className="btn btn-link leveling-plan-grind" ref={grind}>
							<i className="fas fa-person-running" />
						</button>
					</td>
				</tr>
			) as HTMLElement;
			if (result) {
				tippy(row.children[7] as HTMLElement, {
					content: `${xpToNextLevel(level).toLocaleString()} XP to level ${level + 1}, ±${(result.hours * result.relativeError * 1.96).toFixed(
						2,
					)} hours`,
				});
			}
			tippy(grind.value!, { content: `Go to the Grinding mode as the plan's level ${level}` });
			grind.value!.addEventListener('click', () => this.grind(level));
			tbody.appendChild(row);
		}
		table.appendChild(tbody);
		this.tableElem.replaceChildren(table);
	}

	// What the sidebar shows in the Leveling mode: the hours of the last run, or the progress.
	private sidebarSummary(): HTMLElement {
		const run = this.shownRun();
		if (!run?.levels.length) {
			return (<p className="sim-history-empty">{this.running ? 'Simming the first level...' : 'No Leveling runs yet.'}</p>) as HTMLElement;
		}
		const hours = this.totalHours(run.levels);
		const last = run.levels[run.levels.length - 1].level;
		const deaths = run.levels.reduce((sum, level) => sum + level.deathsPerHour * level.hours, 0);
		return (
			<div className="results-sim">
				<div className="leveling-results">
					<div className="results-metric results-sim-hours">
						<span className="topline-result-avg">{formatHours(hours)}</span>
					</div>
					<div className="leveling-stats">
						<span>{`Levels ${run.fromLevel} to ${last + 1}${this.running ? ', so far' : run.stopped ? ', stopped' : ''}`}</span>
						<span>{`${Math.round(deaths)} deaths`}</span>
						<span>{`${run.iterations} sessions a level`}</span>
					</div>
				</div>
			</div>
		) as HTMLElement;
	}

	private loadRun(): PlanRun | null {
		try {
			const json = window.localStorage.getItem(this.simUI.getStorageKey(PLAN_RUN_KEY));
			const run = json ? JSON.parse(json) : null;
			return run && Array.isArray(run.levels) ? run : null;
		} catch (e) {
			console.warn('Failed to load the last leveling plan run: ' + e);
			return null;
		}
	}

	private storeRun(run: PlanRun) {
		try {
			window.localStorage.setItem(this.simUI.getStorageKey(PLAN_RUN_KEY), JSON.stringify(run));
		} catch (e) {
			console.warn('Failed to save the leveling plan run: ' + e);
		}
	}
}

// The talents we got from one talents string to the next, like "Flurry 2/5".
function newTalents(before: string, now: string, shape: TalentShape): string {
	const a = talentsStringToCounts(before, shape);
	const b = talentsStringToCounts(now, shape);
	const names: string[] = [];
	shape.trees.forEach((tree, t) =>
		tree.talents.forEach((talent, i) => {
			if (b[t][i] > a[t][i]) names.push(`${talent.name} ${b[t][i]}/${talent.maxPoints}`);
		}),
	);
	return names.join(', ');
}

function formatHours(hours: number): string {
	const h = Math.floor(hours);
	const m = Math.round((hours - h) * 60);
	return m == 60 ? `${h + 1}h 0m` : `${h}h ${m}m`;
}
