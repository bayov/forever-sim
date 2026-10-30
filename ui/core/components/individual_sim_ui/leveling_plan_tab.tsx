import tippy from 'tippy.js';
import { ref } from 'tsx-vanilla';

import { setItemQualityCssClass } from '../../css_utils';
import { IndividualSimUI } from '../../individual_sim_ui';
import { LEVELING_GEAR_SLOTS } from '../../leveling_gear';
import {
	BONUS_TALENT_POINTS_LEVEL,
	buildIndexAt,
	buildLevels,
	hoursAtLevel,
	LEVELING_PLAN_MAX_LEVEL,
	planLevels,
	rangeAt,
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
import { Consumes, Debuffs, EquipmentSpec, IndividualBuffs, PartyBuffs, RaidBuffs, Spec } from '../../proto/common';
import { LevelingGearClassRules, LevelingGearWeights, LevelingPlanSearchRequest, LevelingPlanSettings, LevelingTalentShape } from '../../proto/ui';
import { ActionId } from '../../proto_utils/action_id';
import { Database } from '../../proto_utils/database';
import { classToEligibleRangedWeaponTypes, classToEligibleWeaponTypes, dualWieldClasses, maxArmorTypeAt, raceToFaction } from '../../proto_utils/utils';
import { encounterMode } from '../../sim_mode';
import { TypedEvent } from '../../typed_event';
import { BooleanPicker } from '../boolean_picker';
import { ContentBlock } from '../content_block';
import { NumberPicker } from '../number_picker';
import { SimTab } from '../sim_tab';
import { LevelingPlanCharacter } from './leveling_plan_character';
import { LevelingGearSection } from './leveling_plan_gear';
import { LevelingRotationsSection } from './leveling_plan_rotations';
import { LevelingTalentSearch } from './leveling_plan_talent_search';
import { LevelingTalentsSection } from './leveling_plan_talents';

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
	private lastRun: PlanRun | null = null;
	private running: PlanRun | null = null;
	private progress: ProgressMetrics | null = null;
	private readonly sections = new Map<Section, HTMLElement>();
	private readonly navButtons = new Map<Section, HTMLButtonElement>();
	private readonly tableElem: HTMLElement;
	private readonly respecsElem: HTMLElement;
	private readonly tilesElem: HTMLElement;
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

		// The overview: the totals of the last run and the plan's levels on top, then the
		// results by level.
		const overview = this.sections.get('overview')!;
		const side = (<div className="leveling-plan-side" />) as HTMLElement;
		const main = (<div className="leveling-plan-main" />) as HTMLElement;
		this.tilesElem = (<div className="leveling-plan-tiles" />) as HTMLElement;
		overview.append(
			<div className="leveling-plan-overview-top">
				{this.tilesElem}
				{side}
			</div>,
			main,
		);
		const levelsBlock = new ContentBlock(side, 'leveling-plan-levels-block', {
			header: {
				title: 'Levels',
				tooltip:
					'We sim a grinding session at each level from the first level up to the last, and add up the hours each level takes. Quests, dungeons and travel are left out, so it is the time of a plan that only grinds.',
			},
		});
		const reset = ref<HTMLButtonElement>();
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
				getValue: player => get(player.readLevelingPlan()),
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
			getValue: player => player.readLevelingPlan().useSettingsBuffs,
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

		// The Gear and Rotations parts each show the plan's character at a level of their own.
		const makeRequest = (level: number) => this.makeRequest(level);
		const characterKey = (level: number) => this.characterKey(level);
		this.gear = new LevelingGearSection(
			this.sections.get('gear')!,
			simUI,
			new LevelingPlanCharacter(simUI.sim, this.player.spec, makeRequest),
			characterKey,
			level => this.grind(level),
		);
		this.talents = new LevelingTalentsSection(this.sections.get('talents')!, simUI);
		new LevelingTalentSearch(this.talents.rootElem, simUI, this.talents, () => this.searchRequest());
		this.rotations = new LevelingRotationsSection(
			this.sections.get('rotations')!,
			simUI,
			new LevelingPlanCharacter(simUI.sim, this.player.spec, makeRequest),
			characterKey,
		);
		this.gear.planChangeEmitter.on(() => this.rotations.refresh());
		this.showSection('overview');
		simUI.sim.encounter.changeEmitter.on(() => this.syncActiveSections());

		this.lastRun = this.loadRun();
		simUI.levelingPlanRun = onProgress => this.run(onProgress);
		simUI.levelingPlanSummary = () => this.sidebarSummary();

		simUI.sim.waitForInit().then(() => {
			// Settings saved before the plan had talents and rotations get the spec's plan.
			const plan = this.player.getLevelingPlan();
			const defaults = this.simUI.individualConfig.defaults.levelingPlan;
			if (defaults && !plan.builds.length && !plan.rotations.length) this.player.setLevelingPlan(TypedEvent.nextEventID(), defaults);
			this.render();
			this.moveRotationsToTopRanks();
			TypedEvent.onAny([this.player.levelingPlanChangeEmitter, this.gear.planChangeEmitter]).on(() => this.render());
		});
	}

	// Plans saved before 2026-10-17 name the ranks their rotations had at one level, so a
	// rotation cast those ranks over its whole level range. Now a rotation names an ability's
	// top rank to cast the highest rank we know (the rotation editor's "Highest rank"). We move
	// a saved plan over once. Each library rotation goes to the level of the first range that
	// plays it, and its spells at that level's highest rank name the top rank instead. A lower
	// rank stays, because it was a downrank on purpose.
	private async moveRotationsToTopRanks() {
		const plan = this.player.getLevelingPlan();
		if (plan.rotationsNameTopRanks || !plan.rotationLibrary.length) return;
		const result = LevelingPlanSettings.clone(plan);
		for (const preset of result.rotationLibrary) {
			if (!preset.rotation) continue;
			const levels = result.rotations.filter(range => range.name == preset.name).map(range => Math.max(range.level, result.fromLevel));
			const request = this.makeRequest(levels.length ? Math.min(...levels) : result.fromLevel);
			if (!request) continue;
			const metadata = (await this.simUI.sim.computeStatsOfRequest(request)).raidStats?.parties[0]?.players[0]?.metadata;
			if (!metadata) continue;
			const tops = new Map<number, number>();
			for (const known of [...metadata.spells, ...metadata.auras]) {
				if (known.id?.rawId.oneofKind == 'spellId' && known.highestRankId) tops.set(known.id.rawId.spellId, known.highestRankId);
			}
			const json = JSON.stringify(APLRotation.toJson(preset.rotation));
			preset.rotation = APLRotation.fromJson(JSON.parse(json, (key, value) => (key == 'spellId' && tops.has(value) ? tops.get(value) : value)));
		}
		result.rotationsNameTopRanks = true;
		// The plan may have changed while we asked the sim, and then we try again next time.
		if (this.player.levelingPlanEquals(plan)) this.player.setLevelingPlan(TypedEvent.nextEventID(), result);
	}

	protected buildTabContent() {
		// The constructor builds it, because the parts need the tab's fields.
	}

	private showSection(section: Section) {
		this.section = section;
		this.syncActiveSections();
		this.sections.forEach((elem, id) => elem.classList.toggle('hide', id != section));
		this.navButtons.forEach((button, id) => {
			button.classList.toggle('active', id == section);
			button.setAttribute('aria-pressed', String(id == section));
		});
	}

	// The request of one level: the plan's gear, talents and rotation there, in the grinding
	// session of the Encounter settings. With buildIndex we take that talent build instead of
	// the plan's, like the one we had before a respec, and with talents those talents. With
	// seed, the sessions are those of that seed, so a respec and the build we had play the same
	// mobs and the same rolls.
	private makeRequest(level: number, options: { seed?: bigint; buildIndex?: number } = {}): RaidSimRequest | null {
		const plan = this.player.getLevelingPlan();
		const gear = this.gear.gearAt(level);
		if (!gear) return null;
		const request = this.simUI.sim.makeRaidSimRequest(false);
		request.simOptions!.debugFirstIteration = false;
		if (options.seed !== undefined) request.simOptions!.randomSeed = options.seed;
		const player = request.raid!.parties[0].players[0];
		player.level = level;
		player.bonusTalentPoints = level >= BONUS_TALENT_POINTS_LEVEL ? plan.bonusTalentPoints : 0;
		player.talentsString = talentsAt(plan, level, this.talents.shape, options.buildIndex);
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
		if (rotation) player.rotation = rotation;
		const imbue = rangeAt(plan, level)?.imbue;
		if (imbue) {
			const spec = player.spec as any;
			const options = spec?.[spec.oneofKind]?.options;
			if (options && 'shamanImbue' in options) options.shamanImbue = imbue;
		}
		if (request.encounter?.leveling) request.encounter.leveling.plan = false;
		return request;
	}

	// What the Go talent search needs (tools/levelingplan): the settings' request, which it makes
	// each level's request from like makeRequest does, the plan and its gear plan, and our gear
	// at each level, which it checks its own gear plan against. Null before the database loads.
	private searchRequest(): LevelingPlanSearchRequest | null {
		const plan = this.player.getLevelingPlan();
		const levels = planLevels(plan);
		const gear = levels.map(level => this.gear.gearAt(level));
		if (gear.some(g => !g)) return null;
		const playerClass = this.player.getClass();
		const shape = this.talents.shape;
		return LevelingPlanSearchRequest.create({
			base: this.simUI.sim.makeRaidSimRequest(false),
			plan,
			gearPlan: this.player.getLevelingGearPlan(),
			route: this.gear.route(),
			gearWeights: this.gear.weightPresets().map(preset => LevelingGearWeights.create({ level: preset.level, weights: preset.weights.toProto() })),
			gearDefaultWeights: this.player.getEpWeights().toProto(),
			classRules: LevelingGearClassRules.create({
				maxArmorType: Array.from({ length: LEVELING_PLAN_MAX_LEVEL }, (_, i) => maxArmorTypeAt(playerClass, i + 1)),
				weaponTypes: classToEligibleWeaponTypes[playerClass].map(wt => wt.weaponType),
				twoHandWeaponTypes: classToEligibleWeaponTypes[playerClass].filter(wt => wt.canUseTwoHand).map(wt => wt.weaponType),
				rangedWeaponTypes: classToEligibleRangedWeaponTypes[playerClass],
				canDualWield: dualWieldClasses.includes(playerClass),
			}),
			faction: raceToFaction[this.player.getRace()],
			talentShape: LevelingTalentShape.create({
				pointsPerRow: shape.pointsPerRow,
				trees: shape.trees.map(tree => ({
					name: tree.name,
					talents: tree.talents.map(talent => ({
						name: talent.name,
						row: talent.row,
						col: talent.col,
						maxPoints: talent.maxPoints,
						prereq: talent.prereq ?? -1,
						notSimulated: !!talent.notSimulated,
					})),
				})),
			}),
			uiGear: gear.map(g => g?.asSpec() ?? EquipmentSpec.create()),
		});
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
				const request = this.makeRequest(job.level, {
					buildIndex: job.kept !== undefined ? job.kept - 1 : undefined,
					seed: seed + BigInt(job.level) * 1_000_003n,
				});
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

	// The Gear and Rotations parts load the plan's character only while we see them, so the
	// other modes compute no stats for it.
	private syncActiveSections() {
		const shown = encounterMode(this.simUI.sim.encounter) == 'levelingPlan';
		this.gear.setActive(shown && this.section == 'gear');
		this.rotations.setActive(shown && this.section == 'rotations');
	}

	// What the plan's character at a level is made of, besides its rotation. When it's the same
	// as when we loaded the character, the character is still right.
	private characterKey(level: number): string {
		const plan = this.player.readLevelingPlan();
		return JSON.stringify([
			level,
			talentsAt(plan, level, this.talents.shape),
			plan.bonusTalentPoints,
			plan.useSettingsBuffs,
			rangeAt(plan, level)?.imbue ?? 0,
			this.gear.gearAt(level)?.asSpec() ?? null,
		]);
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
		const imbue = rangeAt(plan, level)?.imbue;
		const eventID = TypedEvent.nextEventID();
		TypedEvent.freezeAllAndDo(() => {
			this.player.setLevel(eventID, level);
			this.player.setBonusTalentPoints(eventID, level >= BONUS_TALENT_POINTS_LEVEL ? plan.bonusTalentPoints : 0);
			this.player.setTalentsString(eventID, talentsAt(plan, level, this.talents.shape));
			if (gear) this.player.setGear(eventID, gear);
			if (rotation) this.player.setAplRotation(eventID, rotation);
			if (imbue) {
				const options = this.player.getSpecOptions() as any;
				if ('shamanImbue' in options) this.player.setSpecOptions(eventID, { ...options, shamanImbue: imbue });
			}
		});
	}

	private render() {
		this.renderTiles();
		this.renderRespecs();
		this.renderTable();
	}

	private totalHours(levels: LevelResult[]): number {
		return levels.reduce((sum, level) => sum + level.hours, 0);
	}

	// The totals of the run we show, a tile each.
	private renderTiles() {
		const plan = this.player.readLevelingPlan();
		const run = this.shownRun();
		const levels = run?.levels ?? [];
		const hours = this.totalHours(levels);
		const deaths = levels.reduce((sum, level) => sum + level.deathsPerHour * level.hours, 0);
		const respecs = sortedBuilds(plan).length - 1;
		const tile = (value: string, label: string, tooltip: string, className = '') => {
			const elem = (
				<div className={`leveling-plan-tile ${className}`}>
					<span className="leveling-plan-tile-value">{value}</span>
					<span className="leveling-plan-tile-label">{label}</span>
				</div>
			) as HTMLElement;
			tippy(elem, { content: tooltip });
			return elem;
		};
		const done = levels.length ? levels[levels.length - 1].level + 1 : plan.fromLevel;
		this.tilesElem.replaceChildren(
			tile(
				levels.length ? formatHours(hours) : '-',
				this.running ? 'So far' : 'Grinding time',
				levels.length
					? `The hours of grinding from level ${plan.fromLevel} to ${done}, by the ${this.running ? 'run going now' : 'last run'}.`
					: 'Simulate to see how long the plan takes.',
				'leveling-plan-tile-main',
			),
			tile(`${plan.fromLevel} to ${plan.toLevel}`, 'Levels', `${planLevels(plan).length} levels, a session sim each.`),
			tile(levels.length ? String(Math.round(deaths)) : '-', 'Deaths', 'Deaths over all the levels, from the deaths an hour at each one.'),
			tile(String(respecs), respecs == 1 ? 'Respec' : 'Respecs', 'Talent builds after the first one, see the Talents part.'),
			tile(run ? String(run.iterations) : '-', 'Sessions a level', 'The iterations of the run, one grinding session each.'),
		);
	}

	private renderRespecs() {
		const run = this.shownRun();
		const plan = this.player.readLevelingPlan();
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
		const plan = this.player.readLevelingPlan();
		const run = this.shownRun();
		const gearPlan = this.gear.getPlan();
		const builds = sortedBuilds(plan);
		const shape = this.talents.shape;
		const results = planLevels(plan).map(level => run?.levels.find(l => l.level == level));
		const maxXp = Math.max(1, ...results.map(result => result?.xpPerHour ?? 0));
		const maxHours = Math.max(0.01, ...results.map(result => (result && isFinite(result.hours) ? result.hours : 0)));
		const table = (<table className="leveling-plan-table table table-sm" />) as HTMLTableElement;
		table.appendChild(
			<thead>
				<tr>
					<th>Level</th>
					<th>Talents</th>
					<th>New gear</th>
					<th>Rotation</th>
					<th className="leveling-plan-bar-col">XP / h</th>
					<th className="text-end">Mobs / h</th>
					<th className="text-end">Deaths / h</th>
					<th className="leveling-plan-bar-col">Hours</th>
					<th className="text-end">Total</th>
					<th />
				</tr>
			</thead>,
		);
		const tbody = document.createElement('tbody');
		let total = 0;
		let lastRange = '';
		planLevels(plan).forEach((level, index) => {
			const buildIndex = buildIndexAt(plan, level);
			const respec = buildIndex > 0 && builds[buildIndex].level == level;
			const before = level > plan.fromLevel && !respec ? talentsAt(plan, level - 1, shape) : '';
			const gained = newTalents(before, talentsAt(plan, level, shape), shape);
			const range = rangeAt(plan, level);
			const rangeName = range?.name || "Rotation tab's";
			const rangeStart = rangeName != lastRange;
			lastRange = rangeName;
			const result = results[index];
			if (result) total += result.hours;

			const talentsCell = ref<HTMLTableCellElement>();
			const gearCell = ref<HTMLTableCellElement>();
			const rotationCell = ref<HTMLTableCellElement>();
			const hoursCell = ref<HTMLTableCellElement>();
			const grind = ref<HTMLButtonElement>();
			const row = (
				<tr className={respec ? 'leveling-plan-respec-row' : ''}>
					<td className="leveling-plan-level-cell">{String(level)}</td>
					<td className="leveling-plan-talents-cell" ref={talentsCell} />
					<td className="leveling-plan-gear-cell" ref={gearCell} />
					<td className="leveling-plan-rotation-cell" ref={rotationCell} />
					<td className="leveling-plan-bar-col">
						{result ? bar(result.xpPerHour / maxXp, Math.round(result.xpPerHour).toLocaleString(), 'xp') : null}
					</td>
					<td className="text-end">{result ? result.mobsPerHour.toFixed(1) : ''}</td>
					<td className={`text-end ${result && result.deathsPerHour >= 1 ? 'leveling-plan-deaths-high' : ''}`}>
						{result ? result.deathsPerHour.toFixed(2) : ''}
					</td>
					<td className="leveling-plan-bar-col" ref={hoursCell}>
						{result ? bar(isFinite(result.hours) ? result.hours / maxHours : 1, result.hours.toFixed(2), 'hours') : null}
					</td>
					<td className="text-end leveling-plan-total-cell">{result ? total.toFixed(1) : ''}</td>
					<td>
						<button className="btn btn-link leveling-plan-grind" ref={grind}>
							<i className="fas fa-person-running" />
						</button>
					</td>
				</tr>
			) as HTMLElement;

			if (respec) talentsCell.value!.appendChild(<span className="leveling-plan-badge">Respec</span>);
			gained.forEach(({ tree, talent, rank, maxPoints, name }) => {
				const icon = (<span className="leveling-plan-talent-icon small" />) as HTMLElement;
				this.talents.talentIcon(tree, talent, url => (icon.style.backgroundImage = `url('${url}')`));
				talentsCell.value!.appendChild(
					<span className="leveling-plan-talent-gain">
						{icon}
						{!respec || gained.length == 1 ? <span>{`${name} ${rank}/${maxPoints}`}</span> : null}
					</span>,
				);
				if (respec && gained.length > 1) tippy(icon, { content: `${name} ${rank}/${maxPoints}` });
			});

			const levelGear = gearPlan?.levels[level - 1];
			levelGear?.changed
				.filter(slot => LEVELING_GEAR_SLOTS.includes(slot))
				.forEach(slot => {
					const piece = levelGear.gear[slot];
					if (!piece) return;
					const icon = (<a className="leveling-plan-item-icon" href="javascript:void(0)" dataset={{ whtticon: 'false' }} />) as HTMLAnchorElement;
					setItemQualityCssClass(icon, piece.item.quality);
					ActionId.fromItemId(piece.item.id, undefined, piece.randomSuffix?.id).fillAndSet(icon, true, true);
					gearCell.value!.appendChild(icon);
				});

			const chip = (
				<span
					className={`leveling-plan-rotation-chip ${rangeStart ? '' : 'continued'}`}
					style={{ '--segment-color': this.rotations.colorOf(plan, range?.name ?? '') } as any}>
					{rangeStart ? <span>{rangeName}</span> : null}
				</span>
			) as HTMLElement;
			rotationCell.value!.appendChild(chip);

			if (result) {
				tippy(hoursCell.value!, {
					content: `${xpToNextLevel(level).toLocaleString()} XP to level ${level + 1}, ±${(result.hours * result.relativeError * 1.96).toFixed(
						2,
					)} hours`,
				});
			}
			tippy(grind.value!, { content: `Go to the Grinding mode as the plan's level ${level}` });
			grind.value!.addEventListener('click', () => this.grind(level));
			tbody.appendChild(row);
		});
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

// The talent ranks we got from one talents string to the next.
function newTalents(before: string, now: string, shape: TalentShape) {
	const a = talentsStringToCounts(before, shape);
	const b = talentsStringToCounts(now, shape);
	const gained: Array<{ tree: number; talent: number; rank: number; maxPoints: number; name: string }> = [];
	shape.trees.forEach((tree, t) =>
		tree.talents.forEach((talent, i) => {
			if (b[t][i] > a[t][i]) gained.push({ tree: t, talent: i, rank: b[t][i], maxPoints: talent.maxPoints, name: talent.name });
		}),
	);
	return gained;
}

// A bar of the share of the column's largest value, with the value on it.
function bar(share: number, text: string, kind: string): HTMLElement {
	return (
		<div className={`leveling-plan-bar leveling-plan-bar-${kind}`}>
			<span className="leveling-plan-bar-fill" style={{ width: `${Math.round(Math.min(Math.max(share, 0), 1) * 100)}%` }} />
			<span className="leveling-plan-bar-text">{text}</span>
		</div>
	) as HTMLElement;
}

function formatHours(hours: number): string {
	const h = Math.floor(hours);
	const m = Math.round((hours - h) * 60);
	return m == 60 ? `${h + 1}h 0m` : `${h}h ${m}m`;
}
