// The comp editor (docs/raid_sim.md, The page, M2b): a raid of 10, 20 or 40 in parties of 5, each slot a
// DPS Report build or a tank or healer placeholder, with its assignments, the encounter, the faction and
// the camp buffs. Run builds a RaidComp, has the resolver turn it into the raid's request (resolveComp),
// runs that request in the sim's workers and shows the raid's results.
import { ChangeCategory } from '../core/components/dirty_settings';
import { EncounterPicker } from '../core/components/encounter_picker';
import { SimProgress } from '../core/components/sim_progress';
import {
	Assignment,
	Blessing,
	CompEncounter,
	CompMemberBuild,
	CompResult,
	CompScenarioResult,
	CompSearchRequest,
	CompSlot,
	CompValuesResult,
	Duty,
	PaladinJudgement,
	Player,
	ProgressMetrics,
	RaidAssumptions,
	RaidComp,
	RaidCompResolution,
	RaidPackage,
	RaidPreset,
	RaidRole,
	RaidSimRequest,
	RaidSimResult,
	RaidTotem,
	SimOptions,
} from '../core/proto/api';
import { Class, Encounter as EncounterProto, Faction, MobType, Stat } from '../core/proto/common';
import { SimResult } from '../core/proto_utils/sim_result';
import { classNames, cssClassForClass } from '../core/proto_utils/utils';
import { RequestTypes } from '../core/sim_signal_manager';
import { EventID, TypedEvent } from '../core/typed_event';
import { assignmentPicker, Choice, openChoiceMenu, segmented, toolbarItem } from './choice_menu';
import {
	assignedElement,
	BLESSING_CLASSES,
	blessingName,
	buildLabel,
	BuildVariant,
	CLASS_ICONS,
	dutyName,
	factionName,
	judgementName,
	loadRaidSimData,
	loadRaidSimPresets,
	packageName,
	presetInfo,
	presetsFor,
	RAID_SIZES,
	RaidSimData,
	ReportBuild,
	roleName,
	TOTEM_ELEMENTS,
	TotemElement,
	totemElement,
	totemName,
} from './data';
import {
	BLESSING_ICONS,
	buildIcon,
	DUTY_ICONS,
	effectsOf,
	INNERVATE_ICON,
	JUDGEMENT_ICONS,
	MANA_TIDE_ICON,
	memberIcon,
	PACKAGE_ICON,
	packageHasTide,
	POWER_INFUSION_ICON,
	RAID_EFFECTS,
	RaidEffect,
	TOTEM_ICONS,
} from './raid_effects';
import type { RaidSimUI } from './raid_sim_ui';
import { DETAILS_TAB } from './raid_sim_ui';
import { check } from './search_form';
import {
	badge,
	ci95Labelled,
	empty,
	errorStatus,
	fmt,
	fmtSigned,
	iconAlt,
	list,
	MEMBER_VALUE_COLUMNS,
	NA,
	partsToggle,
	plural,
	section,
	table,
} from './view_utils';

const STORAGE_KEY = '__forever_raid_sim__comp__';
const PARTY_SIZE = 5;

// One member of the raid. Its id is the CompSlot id, which assignments name, so it stays the same when the
// member moves.
interface Member {
	id: number;
	buildId?: string; // a base build of the report
	preset?: RaidPreset;
	presetInnervate?: boolean;
}

// The jobs set by hand. Whatever isn't here is automatic.
interface HandAssignments {
	blessings: Record<number, Partial<Record<Class, Blessing>>>;
	judgements: Record<number, PaladinJudgement>;
	totems: Record<number, Partial<Record<TotemElement, RaidTotem>>>;
	duties: Record<number, Duty>;
	powerInfusion: Record<number, number>; // holder: target
	innervate: Record<number, number>;
}

type Boss = 'Generic' | 'Onyxia' | 'Custom';

interface EditorState {
	size: number;
	cells: Array<Member | null>;
	nextId: number;
	faction: Faction;
	campBuffs: boolean;
	boss: Boss;
	encounter?: unknown; // protojson
	// The sidebar's iterations, which Run uses.
	iterations?: number;
	hand: HandAssignments;
}

const emptyHand = (): HandAssignments => ({ blessings: {}, judgements: {}, totems: {}, duties: {}, powerInfusion: {}, innervate: {} });

// The raid as the sim history, the presets and links keep it.
export type RaidState = EditorState;

// A built-in raid of the sidebar's presets.
export interface RaidPresetEntry {
	name: string;
	group: string;
	tooltip: string;
	state: RaidState;
}

// A palette entry or a slot being dragged.
type DragSource = { kind: 'cell'; index: number } | { kind: 'new'; member: Omit<Member, 'id'> };

interface RunOutput {
	comp: RaidComp;
	resolution: RaidCompResolution;
	request?: RaidSimRequest;
	result?: RaidSimResult;
	simResult?: SimResult;
}

export class CompEditor {
	// Fires when the raid changes (its members, assignments, faction, camps or fight), for the presets.
	readonly changeEmitter = new TypedEvent<void>('RaidCompChange');
	private readonly simUI: RaidSimUI;
	private data: RaidSimData | null = null;
	private state: EditorState;
	private readonly rootElem: HTMLElement;
	private readonly controlsHost = (<div className="rs-editor-controls" />) as HTMLElement;
	private readonly paletteHost = (<div className="rs-palette" />) as HTMLElement;
	private readonly raidHost = (<div className="rs-raid" />) as HTMLElement;
	private readonly assignmentsHost = (<div className="rs-assignments" />) as HTMLElement;
	private readonly warningsHost = (<ul className="rs-assignment-warnings" attributes={{ 'aria-live': 'polite' }} />) as HTMLElement;
	private readonly encounterHost = (<div className="rs-encounter" />) as HTMLElement;
	private readonly overviewHost = (<div className="rs-overview" />) as HTMLElement;
	private readonly bossHost = (<div className="rs-boss" />) as HTMLElement;
	private readonly runStatus = (<div className="rs-run-status" />) as HTMLElement;
	private readonly resultsHost = (<div className="rs-run-results" />) as HTMLElement;
	private readonly runButton: HTMLButtonElement;
	private readonly valuesButton: HTMLButtonElement;
	private readonly valuesHost = (<div className="rs-member-values" />) as HTMLElement;
	private drag: DragSource | null = null;
	// The bench's filter, and its list of members to add.
	private readonly benchFilter = (<input className="form-control rs-bench-filter" type="search" placeholder="Filter: fire, wolf, kings…" />) as HTMLInputElement;
	private readonly benchList = (<div className="rs-bench-list" />) as HTMLElement;
	// The slot a member was just added to, which flashes once it's drawn.
	private addedCell: number | null = null;
	private running = false;
	// The running raid sim was stopped from the page.
	private stopping = false;
	private lastRun: RunOutput | null = null;
	// The raid changed after the last run (or member values): its results are kept but marked stale, and
	// its resolved assignments no longer fill the Auto labels.
	private stale = false;
	private valuesStale = false;
	// Counts the edits, so a run or member values that an edit came during are stale when they end.
	private revision = 0;
	private valuesAbort: AbortController | null = null;
	// A slot picked up with the keyboard, moved by Enter or Space on another slot.
	private moving: number | null = null;
	// The slot to focus once the raid is drawn again (after a keyboard move).
	private focusCell: number | null = null;
	private readonly slotHelpId = 'rs-slot-help';
	// The event of the last fight a boss preset or a loaded comp set, which leaves the boss as it is. Any
	// other change to the fight is an edit by hand, which makes it a custom one.
	private presetEvent: EventID | null = null;
	// The raid as it was before the last New raid, Clear, smaller size or loaded comp, for one Undo. The next
	// edit forgets it.
	private undo: { state: EditorState; what: string } | null = null;
	private readonly undoButton: HTMLButtonElement;
	// Why Member values can't run here (a page served as static files has no native server), if it can't.
	private valuesUnavailable: string | null = null;
	private readonly valuesNote = (<span className="rs-static-note rs-static-off d-none" />) as HTMLElement;
	// The notes that mark the sidebar's raid DPS and Raid results as another raid's after an edit.
	private readonly sidebarStaleNote = (<p className="rs-stale-note rs-sidebar-stale d-none">The raid changed after this run.</p>) as HTMLElement;
	private readonly detailsStaleNote = (
		<p className="rs-stale-note d-none" attributes={{ role: 'status' }}>
			The raid changed after this run, so these results are of the raid as it was. Run it again to update them.
		</p>
	) as HTMLElement;

	constructor(parent: HTMLElement, simUI: RaidSimUI) {
		this.simUI = simUI;
		this.state = this.loadState();
		this.runButton = (
			<button className="btn btn-primary rs-run-button" type="button">
				<i className="fas fa-play me-1" />
				Run
			</button>
		) as HTMLButtonElement;
		this.runButton.addEventListener('click', () => this.run());
		this.valuesButton = (
			<button
				className="btn btn-outline-primary rs-values-button"
				type="button"
				title="Each member’s worth to this raid, from single-player sims on the native server">
				<i className="fas fa-calculator me-1" />
				Member values
			</button>
		) as HTMLButtonElement;
		this.valuesButton.addEventListener('click', () => this.memberValues());
		this.undoButton = (
			<button className="btn btn-outline-warning d-none" type="button">
				<i className="fas fa-rotate-left me-1" />
				Undo
			</button>
		) as HTMLButtonElement;
		this.undoButton.addEventListener('click', () => this.undoLast());

		this.rootElem = (
			<div className="rs-editor">
				<div className="rs-toolbar">
					{this.controlsHost}
					<div className="rs-toolbar-actions">
						{this.undoButton}
						{this.valuesButton}
						{this.runButton}
					</div>
				</div>
				{this.valuesNote}
				{this.runStatus}
				{this.overviewHost}
				<div className="rs-editor-layout">
					<div className="rs-editor-main">
						<p className="visually-hidden" id={this.slotHelpId}>
							Enter or Space picks this slot’s member up to move it to another slot. Escape lets go, Delete removes the member, and the arrow keys
							move between slots.
						</p>
						{this.raidHost}
						<p className="rs-hint">
							<i className="fas fa-arrows-up-down-left-right me-1" />
							Drag a member to move or swap it. Click one for its other builds.
						</p>
						{section('Assignments', this.assignmentsHost, { className: 'rs-assignments-section' })}
						{section('Encounter', this.encounterHost, { className: 'rs-encounter-section' })}
						{this.valuesHost}
						{this.resultsHost}
					</div>
					<aside className="rs-bench">{this.paletteHost}</aside>
				</div>
			</div>
		) as HTMLElement;
		this.benchFilter.setAttribute('aria-label', 'Filter the members to add');
		this.benchFilter.addEventListener('input', () => this.renderPalette());
		this.paletteHost.append(
			<div className="rs-bench-head">
				<span className="rs-bench-title">Add members</span>
				<span className="rs-muted">click or drag</span>
			</div>,
			this.benchFilter,
			this.benchList,
		);
		parent.appendChild(this.rootElem);
		simUI.resultsViewer.contentElem.after(this.sidebarStaleNote);
		simUI.tabElem(DETAILS_TAB).prepend(this.detailsStaleNote);

		this.renderControls();
		this.renderRaid();
		this.renderPalette();
		this.renderAssignments();
		this.initEncounter();

		loadRaidSimData()
			.then(data => {
				this.data = data;
				// A fresh page starts with the report's fight.
				if (!this.state.encounter) this.applyBoss(this.state.boss);
				this.renderAll();
			})
			.catch(e => this.setStatus(`The report’s builds didn’t load: ${e}`, true));

		// Member values come from the native server's single-player sims, as a search does.
		simUI.sim.isWasm().then(isWasm => {
			if (!isWasm) return;
			this.valuesUnavailable = 'Member values need the native server (the downloadable sim or make host); this page is served as static files.';
			this.valuesNote.textContent = this.valuesUnavailable;
			this.valuesNote.classList.remove('d-none');
			this.valuesButton.disabled = true;
		});
	}

	// ---- State ----

	private loadState(): EditorState {
		try {
			const saved = window.localStorage.getItem(STORAGE_KEY);
			if (saved) {
				const s = JSON.parse(saved) as EditorState;
				if (RAID_SIZES[s.size] && Array.isArray(s.cells)) return { ...this.defaultState(s.size), ...s, hand: { ...emptyHand(), ...s.hand } };
			}
		} catch (e) {
			console.warn('Failed to read the saved comp:', e);
		}
		return this.defaultState(20);
	}

	private saveState() {
		try {
			this.state.encounter = EncounterProto.toJson(this.simUI.sim.encounter.toProto());
			window.localStorage.setItem(STORAGE_KEY, JSON.stringify(this.state));
		} catch (e) {
			// A private window: the comp just isn't remembered.
		}
		this.changeEmitter.emit(TypedEvent.nextEventID());
	}

	// A new raid: the size's default tanks and healers, placed by the Groups rule (a package giver per
	// party first, the other placeholders in parties without one), and empty DPS slots.
	private defaultState(size: number): EditorState {
		const info = RAID_SIZES[size];
		const healerCycle = [
			RaidPreset.RaidPresetRestorationShaman,
			RaidPreset.RaidPresetHolyPaladin,
			RaidPreset.RaidPresetDisciplinePriest,
			RaidPreset.RaidPresetRestorationDruid,
		];
		const tankCycle = [RaidPreset.RaidPresetProtectionWarrior, RaidPreset.RaidPresetProtectionPaladin, RaidPreset.RaidPresetFeralTank];
		const placeholders: RaidPreset[] = [
			...[...Array(info.tanks).keys()].map(i => tankCycle[i % tankCycle.length]),
			...[...Array(info.healers).keys()].map(i => healerCycle[i % healerCycle.length]),
		];
		const state: EditorState = {
			size,
			cells: Array(size).fill(null),
			nextId: 1,
			faction: Faction.Horde,
			campBuffs: false,
			boss: 'Generic',
			hand: emptyHand(),
		};
		this.placeByGroupsRule(state, placeholders);
		return state;
	}

	private placeByGroupsRule(state: EditorState, presets: RaidPreset[]) {
		const parties = state.size / PARTY_SIZE;
		const freeIn = (p: number) => [...Array(PARTY_SIZE).keys()].map(i => p * PARTY_SIZE + i).find(i => !state.cells[i]);
		const givers = presets.filter(p => presetInfo(p)?.package);
		const others = presets.filter(p => !presetInfo(p)?.package);
		const packaged = new Set<number>();
		givers.forEach((p, i) => {
			const party = i % parties;
			const cell = freeIn(party);
			if (cell === undefined) return;
			state.cells[cell] = { id: state.nextId++, preset: p };
			packaged.add(party);
		});
		// The other placeholders go to the parties without a package while they have room, so the packaged
		// seats are left for members who gain from them (docs/raid_sim.md, Groups, 3). Among those, the party
		// with the fewest placeholders so far takes the next one, which spreads them.
		const placed = Array(parties).fill(0);
		others.forEach(p => {
			const party = [...Array(parties).keys()]
				.filter(q => freeIn(q) !== undefined)
				.sort((a, b) => +packaged.has(a) - +packaged.has(b) || placed[a] - placed[b] || a - b)[0];
			if (party === undefined) return;
			state.cells[freeIn(party)!] = { id: state.nextId++, preset: p };
			placed[party]++;
		});
	}

	private members(): Array<{ member: Member; cell: number }> {
		return this.state.cells.map((member, cell) => ({ member, cell })).filter((m): m is { member: Member; cell: number } => !!m.member);
	}

	private memberById(id: number) {
		return this.members().find(m => m.member.id == id);
	}

	private buildOf(member: Member): ReportBuild | undefined {
		return member.buildId ? this.data?.byId.get(member.buildId)?.build : undefined;
	}

	private classOf(member: Member): Class {
		if (member.preset) return presetInfo(member.preset)?.class ?? Class.ClassUnknown;
		return this.buildOf(member)?.class ?? Class.ClassUnknown;
	}

	private memberName(member: Member): string {
		if (member.preset) return presetInfo(member.preset)?.name ?? 'Placeholder';
		const build = this.buildOf(member);
		return build ? buildLabel(build) : member.buildId ?? '';
	}

	private memberShortName(member: Member, cell: number): string {
		return `${this.memberName(member)} (${Math.floor(cell / PARTY_SIZE) + 1}.${(cell % PARTY_SIZE) + 1})`;
	}

	private changed() {
		this.saveState();
		this.markStale();
		this.renderRaid();
		this.renderAssignments();
	}

	// A change to the raid, its assignments, faction, camps or encounter: the last run and the member values
	// are of another raid now.
	private markStale() {
		this.revision++;
		this.setUndo(null);
		const hadRun = !!this.lastRun && !this.stale;
		const hadValues = this.valuesHost.childElementCount > 0 && !this.valuesStale;
		if (this.lastRun) this.stale = true;
		if (this.valuesHost.childElementCount) this.valuesStale = true;
		if (hadRun) {
			this.renderResults();
			// The Auto labels and slot DPS were the run's.
			this.renderRaid();
			this.renderAssignments();
		}
		if (hadValues) this.renderValuesStale();
	}

	// Forget the last run and the member values (New raid, Clear, a comp loaded from a search or a preset).
	// The sidebar keeps the sim history's latest run and Raid results keep its results, both marked as
	// another raid's.
	private forgetResults() {
		this.lastRun = null;
		this.stale = false;
		this.valuesStale = false;
		this.valuesAbort?.abort();
		this.valuesHost.replaceChildren();
		this.renderResults();
	}

	// The sidebar's latest run is of another raid (RaidSimUI.updateStale).
	setSidebarStale(stale: boolean) {
		this.sidebarStaleNote.classList.toggle('d-none', !stale);
		this.simUI.resultsViewer.contentElem.classList.toggle('rs-stale', stale);
	}

	// Raid results are of another raid (RaidSimUI.updateStale).
	setDetailsStale(stale: boolean) {
		this.detailsStaleNote.classList.toggle('d-none', !stale);
		this.simUI.tabElem(DETAILS_TAB).classList.toggle('rs-details-stale', stale);
	}

	// ---- Undo ----

	// The raid as it stands, its fight included, for an Undo.
	private snapshot(): EditorState {
		this.saveState();
		return JSON.parse(JSON.stringify(this.state));
	}

	// Offers one Undo of a change that replaced the raid (or forgets it, with null).
	private setUndo(state: EditorState | null, what = '') {
		this.undo = state ? { state, what } : null;
		this.undoButton.classList.toggle('d-none', !state);
		this.undoButton.title = state ? `Put the raid back as it was before ${what}` : '';
		this.undoButton.replaceChildren(<i className="fas fa-rotate-left me-1" />, state ? `Undo ${what}` : 'Undo');
	}

	private undoLast() {
		const undo = this.undo;
		if (!undo) return;
		this.state = undo.state;
		this.moving = null;
		if (this.state.encounter) {
			try {
				this.setPresetEncounter(EncounterProto.fromJson(this.state.encounter as any, { ignoreUnknownFields: true }));
			} catch (e) {
				console.warn('Failed to read the undone encounter:', e);
			}
		}
		this.initBossSelect();
		this.renderControls();
		this.changed();
		this.setStatus(`Undid ${undo.what}: the raid is back as it was.`);
	}

	// The last run, while it's of the raid as it stands.
	private get freshRun(): RunOutput | null {
		return this.stale ? null : this.lastRun;
	}

	// A hand-set assignment changed: the raid is another one now.
	private assignmentChanged() {
		this.saveState();
		this.markStale();
		// The cards show the new pick, and the slots their icons.
		this.renderRaid();
		this.renderAssignments();
	}

	private renderAll() {
		this.renderControls();
		this.renderRaid();
		this.renderPalette();
		this.renderAssignments();
	}

	// ---- Controls: size, faction, camps ----

	private renderControls() {
		const s = this.state;
		const newRaid = (
			<button className="btn btn-link rs-toolbar-link" type="button" title="Start over with this size’s usual tanks and healers">
				<i className="fas fa-file me-1" />
				New
			</button>
		) as HTMLButtonElement;
		newRaid.addEventListener('click', () => {
			const before = this.snapshot();
			const next = this.defaultState(this.state.size);
			this.state = { ...next, faction: s.faction, campBuffs: s.campBuffs, boss: s.boss, encounter: s.encounter, iterations: s.iterations };
			this.forgetResults();
			this.changed();
			this.setStatus('');
			this.setUndo(before, 'New raid');
		});
		const clear = (
			<button className="btn btn-link rs-toolbar-link" type="button" title="Empty every slot">
				<i className="fas fa-eraser me-1" />
				Clear
			</button>
		) as HTMLButtonElement;
		clear.addEventListener('click', () => {
			const before = this.snapshot();
			this.state.cells = Array(this.state.size).fill(null);
			this.state.hand = emptyHand();
			this.forgetResults();
			this.changed();
			this.setStatus('');
			this.setUndo(before, 'Clear');
		});

		this.controlsHost.replaceChildren(
			toolbarItem(
				'Size',
				segmented(
					[10, 20, 40].map(n => ({ value: `${n}`, label: `${n}`, title: `${n} players` })),
					`${s.size}`,
					v => this.resize(parseInt(v)),
					'Raid size',
				),
			),
			toolbarItem(
				'Faction',
				segmented(
					[Faction.Horde, Faction.Alliance].map(f => ({ value: `${f}`, label: factionName(f), title: 'Each build takes its best race of this faction' })),
					`${s.faction}`,
					v => {
						this.state.faction = parseInt(v);
						this.assignmentChanged();
						this.renderControls();
					},
					'Faction',
				),
			),
			toolbarItem(
				'Camp buffs',
				segmented(
					[
						{ value: 'off', label: 'Off' },
						{ value: 'on', label: 'On', title: 'The raid relies on camp buffs' },
					],
					s.campBuffs ? 'on' : 'off',
					v => {
						this.state.campBuffs = v == 'on';
						this.assignmentChanged();
						this.renderControls();
					},
					'Camp buffs',
				),
			),
			<div className="rs-toolbar-edit">
				{newRaid}
				{clear}
			</div>,
		);
	}

	private resize(size: number) {
		const before = this.snapshot();
		const cells: Array<Member | null> = Array(size).fill(null);
		// Members keep their place; the ones past the new size are dropped, with their assignments, and the
		// status line names them.
		const dropped: string[] = [];
		this.state.cells.forEach((m, i) => {
			if (!m) return;
			if (i < size) {
				cells[i] = m;
				return;
			}
			dropped.push(this.memberShortName(m, i));
			this.dropAssignmentsOf(m.id);
		});
		if (this.moving !== null && this.moving >= size) this.moving = null;
		this.state.size = size;
		this.state.cells = cells;
		this.changed();
		this.renderControls();
		if (dropped.length) {
			const who = dropped.length == 1 ? 'this member' : `these ${dropped.length} members`;
			this.setStatus(`Down to ${size} players: ${who} past group ${size / PARTY_SIZE} left the raid: ${dropped.join(', ')}. Undo brings them back.`);
			this.setUndo(before, 'the resize');
		}
	}

	// ---- The raid: parties of 5 with drag and drop ----

	// The last run's DPS of the member `id`, while the run is of the raid as it stands.
	private dpsOf(id: number): number | undefined {
		const run = this.freshRun;
		if (!run?.simResult || !run.request) return undefined;
		const k = run.resolution.slotOfPlayer.indexOf(id);
		if (k < 0) return undefined;
		const raidIndex = playersInOrder(run.request)[k];
		return raidIndex !== undefined ? run.simResult.getPlayerWithRaidIndex(raidIndex)?.dps.avg : undefined;
	}

	private renderRaid() {
		const parties = this.state.size / PARTY_SIZE;
		// The slot with the keyboard's focus keeps it when the raid is drawn again.
		const active = document.activeElement;
		const activeCell = active instanceof HTMLElement && this.raidHost.contains(active) && active.dataset.cell ? parseInt(active.dataset.cell) : null;
		const keep = this.focusCell ?? activeCell;
		this.raidHost.replaceChildren(<div className={`rs-parties rs-parties-${parties}`}>{[...Array(parties).keys()].map(p => this.partyElem(p))}</div>);
		if (keep !== null) this.raidHost.querySelector<HTMLElement>(`[data-cell="${keep}"]`)?.focus();
		this.focusCell = null;
		if (this.addedCell !== null) this.raidHost.querySelector(`[data-cell="${this.addedCell}"]`)?.classList.add('rs-slot-added');
		this.addedCell = null;
		this.renderOverview();
	}

	// A party: its name, its own buffs and its DPS in the last run, then its five slots.
	private partyElem(p: number) {
		const cells = [...Array(PARTY_SIZE).keys()].map(i => p * PARTY_SIZE + i);
		const partyDps = cells.reduce((sum, c) => {
			const m = this.state.cells[c];
			return sum + (m ? this.dpsOf(m.id) ?? 0 : 0);
		}, 0);
		return (
			<div className="rs-party">
				<div className="rs-party-header">
					<span className="rs-party-name">Group {p + 1}</span>
					<span className="rs-party-buffs">{this.partyBuffs(cells)}</span>
					{partyDps ? (
						<span className="rs-party-dps" title="The group’s DPS members in the last run">
							{fmt(partyDps)} <span className="rs-unit">DPS</span>
						</span>
					) : null}
				</div>
				<div className="rs-party-slots">{cells.map(c => this.slotElem(c))}</div>
			</div>
		);
	}

	private packageOf(member: Member): RaidPackage {
		return (member.preset ? presetInfo(member.preset)?.package : this.buildOf(member)?.package) ?? RaidPackage.RaidPackageNone;
	}

	// The buff that reaches only this party: a shaman's Mana Spring, with Mana Tide from some. A party with
	// mana users and no shaman shows the empty place where it would be.
	private partyBuffs(cells: number[]): Node[] {
		const members = cells.map(c => this.state.cells[c]).filter((m): m is Member => !!m);
		const giver = members.find(m => this.packageOf(m) != RaidPackage.RaidPackageNone);
		if (giver) {
			const pkg = this.packageOf(giver);
			return [
				<img className="rs-icon rs-party-buff" src={PACKAGE_ICON} alt="Mana Spring" title={`${packageName(pkg)} for this group, from ${this.memberName(giver)}`} />,
				...(packageHasTide(pkg) ? [<img className="rs-icon rs-party-buff" src={MANA_TIDE_ICON} alt="Mana Tide" title={`Mana Tide Totem, from ${this.memberName(giver)}`} />] : []),
			];
		}
		const manaUsers = members.filter(m => !m.preset && this.usesMana(m));
		if (!manaUsers.length) return [];
		return [
			<span className="rs-icon rs-party-gap" title={`No Mana Spring: ${plural(manaUsers.length, 'mana user')} here and no shaman`}>
				<img src={PACKAGE_ICON} alt="No Mana Spring" />
			</span>,
		];
	}

	// What a slot's keys do: Enter or Space picks this slot's member up and puts it down on another slot (a
	// move, or a swap); Escape lets go; Delete removes; the arrow keys go up and down a group and across the
	// groups.
	private slotKey(e: KeyboardEvent, cell: number) {
		const member = this.state.cells[cell];
		const parties = this.state.size / PARTY_SIZE;
		const go = (to: number) => {
			e.preventDefault();
			this.raidHost.querySelector<HTMLElement>(`[data-cell="${to}"]`)?.focus();
		};
		switch (e.key) {
			case 'Enter':
			case ' ':
				e.preventDefault();
				this.activateSlot(cell);
				return;
			case 'Escape':
				if (this.moving === null) return;
				e.preventDefault();
				this.moving = null;
				this.focusCell = cell;
				this.setStatus('');
				this.renderRaid();
				return;
			case 'Delete':
			case 'Backspace':
				if (!member) return;
				e.preventDefault();
				this.removeMember(cell);
				this.focusCell = cell;
				this.renderRaid();
				return;
			case 'ArrowDown':
				return go(cell % PARTY_SIZE < PARTY_SIZE - 1 ? cell + 1 : cell);
			case 'ArrowUp':
				return go(cell % PARTY_SIZE > 0 ? cell - 1 : cell);
			case 'ArrowRight':
				return go(Math.floor(cell / PARTY_SIZE) < parties - 1 ? cell + PARTY_SIZE : cell);
			case 'ArrowLeft':
				return go(cell >= PARTY_SIZE ? cell - PARTY_SIZE : cell);
		}
	}

	// Enter or Space on a slot: pick its member up, or put the one picked up here.
	private activateSlot(cell: number) {
		if (this.moving !== null) {
			const from = this.moving;
			this.moving = null;
			this.focusCell = cell;
			if (from == cell) {
				this.setStatus('');
				this.renderRaid();
			} else {
				this.dropOn(cell, { kind: 'cell', index: from });
				this.setStatus('');
			}
			return;
		}
		const member = this.state.cells[cell];
		if (!member) return;
		this.moving = cell;
		this.focusCell = cell;
		this.setStatus(`Moving ${this.memberShortName(member, cell)}: Enter on another slot puts it there (a swap if it’s taken). Escape lets go.`);
		this.renderRaid();
	}

	private removeMember(cell: number) {
		const member = this.state.cells[cell];
		if (!member) return;
		this.state.cells[cell] = null;
		this.dropAssignmentsOf(member.id);
		if (this.moving == cell) this.moving = null;
		this.changed();
	}

	// A member's name in a slot. The icon and color say the class, so a build goes by its own name.
	private slotName(member: Member) {
		return member.preset ? this.memberName(member) : this.memberName(member).replace(/^[^:]*: /, '');
	}

	private slotElem(cell: number) {
		const member = this.state.cells[cell];
		let elem: HTMLElement;
		const where = `Group ${Math.floor(cell / PARTY_SIZE) + 1}, slot ${(cell % PARTY_SIZE) + 1}`;
		if (!member) {
			elem = (
				<div className="rs-slot rs-slot-empty" title="Drag a member here, or click one under Add members">
					<i className="fas fa-plus" />
				</div>
			) as HTMLElement;
			elem.setAttribute('aria-label', `${where}: empty`);
		} else {
			const cls = this.classOf(member);
			const info = member.preset ? presetInfo(member.preset) : undefined;
			const build = this.buildOf(member);
			const role = info?.role ?? RaidRole.RaidRoleDps;
			const dps = this.dpsOf(member.id);
			// A build this page's data doesn't have (a result from a newer report): Run refuses it.
			const unknownBuild = !member.preset && !!this.data && !build;
			// The slot itself takes Delete, so its remove button stays out of the tab order.
			const remove = (
				<button type="button" className="rs-slot-remove" title="Remove" attributes={{ 'aria-hidden': 'true' }}>
					<i className="fas fa-xmark" />
				</button>
			) as HTMLButtonElement;
			remove.tabIndex = -1;
			remove.addEventListener('click', e => {
				e.stopPropagation();
				this.removeMember(cell);
			});
			const sub = unknownBuild ? 'not one of this page’s builds' : role == RaidRole.RaidRoleDps ? classNames[cls] : `${roleName(role)} · ${classNames[cls]}`;
			elem = (
				<div className={`rs-slot rs-slot-filled rs-class-${cssClassForClass(cls)} ${unknownBuild ? 'rs-slot-unknown' : ''}`} draggable={true}>
					<span className="rs-slot-icon">
						<img src={memberIcon(info, build, cls)} draggable={false} alt={iconAlt(cls)} />
						{role != RaidRole.RaidRoleDps ? (
							<span className={`rs-role rs-role-${role == RaidRole.RaidRoleTank ? 'tank' : 'healer'}`}>
								<i className={`fas ${role == RaidRole.RaidRoleTank ? 'fa-shield-alt' : 'fa-plus'}`} />
							</span>
						) : null}
					</span>
					<span className="rs-slot-text">
						<span className={`rs-slot-name text-${cssClassForClass(cls)}`}>{this.slotName(member)}</span>
						<span className="rs-slot-sub">{sub}</span>
					</span>
					<span className="rs-slot-jobs">{this.jobIcons(member)}</span>
					{dps !== undefined ? <span className="rs-slot-dps">{fmt(dps)}</span> : null}
					{remove}
				</div>
			) as HTMLElement;
			elem.title = `${this.memberName(member)}${info ? ` · ${info.gives}` : build ? ` · gives ${build.gives}` : ''}`;
			elem.setAttribute(
				'aria-label',
				`${where}: ${this.memberName(member)}, ${roleName(role)}${dps !== undefined ? `, ${fmt(dps)} DPS` : ''}${
					this.moving == cell ? ', picked up to move' : ''
				}`,
			);
			if (this.moving == cell) elem.classList.add('rs-slot-moving');
			elem.addEventListener('dragstart', e => {
				this.drag = { kind: 'cell', index: cell };
				e.dataTransfer!.effectAllowed = 'move';
				e.dataTransfer!.setData('text/plain', `rs-cell:${cell}`);
				elem.classList.add('dragfrom');
				this.raidHost.classList.add('rs-dragging');
			});
			elem.addEventListener('dragend', () => {
				elem.classList.remove('dragfrom');
				this.raidHost.classList.remove('rs-dragging');
				this.drag = null;
			});
			// A click offers the member's other builds, or the other placeholders of its role.
			elem.addEventListener('click', () => {
				if (this.moving !== null) {
					this.activateSlot(cell);
					return;
				}
				openChoiceMenu(elem, {
					title: `${where}: swap for`,
					choices: [...this.swapChoices(member), { value: 'remove', label: 'Remove', faIcon: 'fa-xmark', danger: true }],
					value: chipKey(member),
					onPick: v => (v == 'remove' ? this.removeMember(cell) : this.swapMember(cell, v)),
				});
			});
		}
		let depth = 0;
		elem.addEventListener('dragenter', e => {
			if (!this.drag) return;
			e.preventDefault();
			depth++;
			elem.classList.add('dragto');
		});
		elem.addEventListener('dragleave', () => {
			if (!this.drag) return;
			depth = Math.max(0, depth - 1);
			if (depth == 0) elem.classList.remove('dragto');
		});
		elem.addEventListener('dragover', e => {
			if (this.drag) e.preventDefault();
		});
		elem.addEventListener('drop', e => {
			if (!this.drag) return;
			e.preventDefault();
			e.stopPropagation();
			depth = 0;
			elem.classList.remove('dragto');
			this.dropOn(cell, this.drag);
			this.drag = null;
		});
		if (!member) elem.addEventListener('click', () => this.moving !== null && this.activateSlot(cell));
		elem.tabIndex = 0;
		elem.dataset.cell = `${cell}`;
		elem.setAttribute('role', 'button');
		elem.setAttribute('aria-describedby', this.slotHelpId);
		elem.addEventListener('keydown', e => this.slotKey(e, cell));
		if (this.moving !== null && this.moving != cell) elem.classList.add('rs-slot-target');
		return elem;
	}

	// The builds a member can turn into: its class's other builds, or the other tanks or healers.
	private swapChoices(member: Member): Choice[] {
		if (member.preset) {
			const role = presetInfo(member.preset)?.role ?? RaidRole.RaidRoleHealer;
			return presetsFor(role).map(p => ({ value: chipKey({ preset: p.id }), label: p.name, icon: p.icon, labelClass: `text-${cssClassForClass(p.class)}` }));
		}
		const cls = this.classOf(member);
		return (this.data?.builds ?? [])
			.filter(b => b.class == cls)
			.map(b => ({
				value: chipKey({ buildId: b.id }),
				label: buildLabel(b).replace(/^[^:]*: /, ''),
				icon: buildIcon(b),
				hint: `${fmt(b.published, 0)} DPS`,
				labelClass: `text-${cssClassForClass(b.class)}`,
			}));
	}

	// Turns the member at `cell` into another build or placeholder (a chip key). It keeps its place and its
	// assignments, but those the new one can't hold go.
	private swapMember(cell: number, key: string) {
		const member = this.state.cells[cell];
		if (!member || key == chipKey(member)) return;
		const next: Member = key.startsWith('preset-') ? { id: member.id, preset: parseInt(key.slice(7)) } : { id: member.id, buildId: key.slice(6) };
		this.state.cells[cell] = next;
		const h = this.state.hand;
		const id = member.id;
		const info = next.preset ? presetInfo(next.preset) : undefined;
		if (this.classOf(next) != Class.ClassPaladin) delete h.blessings[id];
		if (!info?.paladin) delete h.judgements[id];
		if (!info?.shaman) delete h.totems[id];
		if (!info?.powerInfusion) delete h.powerInfusion[id];
		if (!info?.innervate && !info?.innervateOptional) delete h.innervate[id];
		const duty = h.duties[id];
		if (duty !== undefined && (next.preset || (duty != Duty.DutyNone && !this.buildOf(next)?.duties.some(d => d.duty == duty)))) delete h.duties[id];
		this.addedCell = cell;
		this.changed();
	}

	private dropOn(cell: number, source: DragSource) {
		const cells = this.state.cells;
		if (source.kind == 'cell') {
			if (source.index == cell) return;
			// A move to an empty slot, or a swap.
			[cells[cell], cells[source.index]] = [cells[source.index], cells[cell]];
		} else {
			const old = cells[cell];
			if (old) this.dropAssignmentsOf(old.id);
			cells[cell] = { ...source.member, id: this.state.nextId++ };
			this.addedCell = cell;
		}
		this.changed();
	}

	// A member from the bench, in the first empty slot.
	private addMember(member: Omit<Member, 'id'>) {
		const cell = this.state.cells.findIndex(c => !c);
		if (cell < 0) {
			this.setStatus('The raid is full. Drag a member onto a slot to replace who is there.', true);
			return;
		}
		this.setStatus('');
		this.dropOn(cell, { kind: 'new', member });
	}

	private dropAssignmentsOf(id: number) {
		const h = this.state.hand;
		delete h.blessings[id];
		delete h.judgements[id];
		delete h.totems[id];
		delete h.duties[id];
		delete h.powerInfusion[id];
		delete h.innervate[id];
		for (const map of [h.powerInfusion, h.innervate]) {
			for (const [holder, target] of Object.entries(map)) if (target == id) delete map[parseInt(holder)];
		}
	}

	// ---- What each member does, as icons on its slot ----

	// The member's assignments: what the last run's resolver gave it while the run is of this raid, else
	// what's set by hand. A click on one goes to its card under Assignments.
	private jobIcons(member: Member): Node[] {
		const jobs: Array<{ icon: string; title: string; hand: boolean }> = [];
		const id = member.id;
		const run = this.freshRun;
		const targetName = (target: number) => {
			const t = this.memberById(target);
			return t ? this.memberShortName(t.member, t.cell) : 'nobody';
		};
		if (run) {
			const blessings = new Map<Blessing, { classes: string[]; hand: boolean }>();
			run.resolution.assignments
				.filter(a => a.holder == id)
				.forEach(a => {
					const j = a.job;
					if (j.oneofKind == 'blessing') {
						if (j.blessing.blessing == Blessing.BlessingNone) return;
						const b = blessings.get(j.blessing.blessing) ?? { classes: [], hand: false };
						b.classes.push(`${classNames[j.blessing.class]}s`);
						b.hand ||= a.setByHand;
						blessings.set(j.blessing.blessing, b);
					} else if (j.oneofKind == 'judgement' && j.judgement != PaladinJudgement.PaladinJudgementNone) {
						jobs.push({ icon: JUDGEMENT_ICONS[j.judgement]!, title: `Judgement of ${judgementName(j.judgement)}`, hand: a.setByHand });
					} else if (j.oneofKind == 'totem' && j.totem.totem != RaidTotem.RaidTotemNone) {
						jobs.push({ icon: TOTEM_ICONS[j.totem.totem]!, title: `${totemName(j.totem.totem)} Totem`, hand: a.setByHand });
					} else if (j.oneofKind == 'duty' && j.duty != Duty.DutyNone) {
						jobs.push({ icon: DUTY_ICONS[j.duty]!, title: dutyName(j.duty), hand: a.setByHand });
					} else if (j.oneofKind == 'cooldown') {
						const pi = j.cooldown.powerInfusion;
						jobs.push({ icon: pi ? POWER_INFUSION_ICON : INNERVATE_ICON, title: `${pi ? 'Power Infusion' : 'Innervate'} on ${targetName(j.cooldown.target)}`, hand: a.setByHand });
					}
				});
			[...blessings.entries()].forEach(([b, { classes, hand }]) =>
				jobs.unshift({ icon: BLESSING_ICONS[b]!, title: `Greater Blessing of ${blessingName(b)}: ${classes.join(', ')}`, hand }),
			);
		} else {
			const h = this.state.hand;
			const byBlessing = new Map<Blessing, string[]>();
			Object.entries(h.blessings[id] ?? {}).forEach(([c, b]) => {
				if (b == Blessing.BlessingNone || b === undefined) return;
				byBlessing.set(b, [...(byBlessing.get(b) ?? []), `${classNames[Number(c) as Class]}s`]);
			});
			byBlessing.forEach((classes, b) => jobs.push({ icon: BLESSING_ICONS[b]!, title: `Greater Blessing of ${blessingName(b)}: ${classes.join(', ')}`, hand: true }));
			const j = h.judgements[id];
			if (j !== undefined && j != PaladinJudgement.PaladinJudgementNone) jobs.push({ icon: JUDGEMENT_ICONS[j]!, title: `Judgement of ${judgementName(j)}`, hand: true });
			Object.values(h.totems[id] ?? {}).forEach(t => {
				if (t !== undefined && t != RaidTotem.RaidTotemNone) jobs.push({ icon: TOTEM_ICONS[t]!, title: `${totemName(t)} Totem`, hand: true });
			});
			const d = h.duties[id];
			if (d !== undefined && d != Duty.DutyNone) jobs.push({ icon: DUTY_ICONS[d]!, title: dutyName(d), hand: true });
			if (h.powerInfusion[id] !== undefined) jobs.push({ icon: POWER_INFUSION_ICON, title: `Power Infusion on ${targetName(h.powerInfusion[id])}`, hand: true });
			if (h.innervate[id] !== undefined) jobs.push({ icon: INNERVATE_ICON, title: `Innervate on ${targetName(h.innervate[id])}`, hand: true });
		}
		// Four fit beside a name. Past that, three and how many more.
		const shown = jobs.length > 4 ? jobs.slice(0, 3) : jobs;
		const icons: Node[] = shown.map(job => {
			const icon = (
				<button type="button" className={`rs-job ${job.hand ? 'rs-job-hand' : ''}`} title={`${job.title}${job.hand ? ' (set by hand)' : ' (auto)'}`} tabIndex={-1}>
					<img className="rs-icon" src={job.icon} alt={job.title} draggable={false} />
				</button>
			) as HTMLButtonElement;
			icon.addEventListener('click', e => {
				e.stopPropagation();
				this.showAssignmentsOf(id);
			});
			return icon;
		});
		if (jobs.length > 4) icons.push(<span className="rs-job-more" title={jobs.slice(3).map(j => j.title).join('\n')}>+{jobs.length - 3}</span>);
		return icons;
	}

	// Scrolls to a member's assignments and flashes them.
	private showAssignmentsOf(id: number) {
		const rows = [...this.assignmentsHost.querySelectorAll<HTMLElement>(`[data-holder="${id}"]`)];
		if (!rows.length) return;
		rows[0].scrollIntoView({ behavior: 'smooth', block: 'center' });
		rows.forEach(r => {
			r.classList.remove('rs-flash');
			void r.offsetWidth;
			r.classList.add('rs-flash');
		});
	}

	// ---- Over the raid: its roles, and the buffs and debuffs it has ----

	private renderOverview() {
		const members = this.members();
		const size = this.state.size;
		const info = RAID_SIZES[size];
		const roleOf = (m: Member) => (m.preset ? presetInfo(m.preset)?.role ?? RaidRole.RaidRoleHealer : RaidRole.RaidRoleDps);
		const count = (role: RaidRole) => members.filter(m => roleOf(m.member) == role).length;
		const roleChip = (icon: string, n: number, word: string, range?: [number, number]) => {
			const off = !!range && (n < range[0] || n > range[1]);
			const usual = range ? (range[0] == range[1] ? `${range[0]}` : `${range[0]} to ${range[1]}`) : '';
			return (
				<span className={`rs-role-count ${off ? 'rs-off' : ''}`} title={range ? `A ${size}-player raid usually has ${usual} ${word}` : ''}>
					<i className={`fas ${icon}`} />
					<b>{n}</b> {word}
				</span>
			);
		};
		const open = size - members.length;

		// Who could give each effect, and after a run of this raid, what was up.
		const givers = new Map<string, string[]>();
		members.forEach(({ member, cell }) =>
			effectsOf(member.preset, this.buildOf(member)).forEach(e => givers.set(e, [...(givers.get(e) ?? []), this.memberShortName(member, cell)])),
		);
		const run = this.freshRun?.request ? this.freshRun : null;
		const records = new Map((run?.resolution.buffs ?? []).map(b => [b.effect, b]));
		const blessed = !!run?.resolution.assignments.some(a => a.job.oneofKind == 'blessing' && a.job.blessing.blessing != Blessing.BlessingNone);
		const tile = (e: RaidEffect) => {
			const who = givers.get(e.name) ?? [];
			const record = records.get(e.name);
			const up = run ? !!record || (e.name == 'Blessings' && blessed) : false;
			const state = run ? (up ? 'up' : who.length ? 'off' : 'none') : who.length ? 'can' : 'none';
			const name = e.label ?? e.name;
			const why = {
				up: `Up, from ${record?.provider || who.join(', ')}`,
				can: `From ${who.join(', ')}`,
				off: `Not up in the last run. ${who.join(', ')} could give it.`,
				none: 'Nobody here gives it',
			}[state];
			return (
				<span className={`rs-effect rs-effect-${state}`} title={`${name}\n${why}`}>
					<img className="rs-icon" src={e.icon} alt={name} />
				</span>
			);
		};
		const last = this.lastRun?.simResult && this.lastRun.request ? this.lastRun : null;
		let dps: Node | null = null;
		if (last) {
			const iterations = last.request!.simOptions?.iterations || 1;
			const raidDps = last.simResult!.raidMetrics.dps;
			dps = (
				<div className={`rs-overview-dps ${this.stale ? 'rs-stale' : ''}`} title={this.stale ? 'The raid changed after this run' : ''}>
					<span className="rs-overview-dps-value">{fmt(raidDps.avg)}</span>
					<span className="rs-overview-dps-label">
						raid DPS
						<span className="rs-ci" title="The average’s 95% interval">
							±{fmt((1.96 * raidDps.stdev) / Math.sqrt(iterations))}
						</span>
					</span>
					<button className="btn btn-link rs-details-link" type="button" onclick={() => this.simUI.activate(DETAILS_TAB)}>
						Details <i className="fas fa-arrow-right" />
					</button>
				</div>
			);
		}
		this.overviewHost.replaceChildren(
			<div className="rs-overview-roles">
				{roleChip('fa-shield-alt', count(RaidRole.RaidRoleTank), 'tanks', info.tankRange)}
				{roleChip('fa-plus', count(RaidRole.RaidRoleHealer), 'healers', info.healerRange)}
				{roleChip('fa-khanda', count(RaidRole.RaidRoleDps), 'DPS')}
				{open ? <span className="rs-role-count rs-open">{open} open</span> : null}
			</div>,
			<div className="rs-strip" title="">
				<div className="rs-strip-group">
					<span className="rs-strip-label">Buffs</span>
					{RAID_EFFECTS.filter(e => !e.debuff).map(tile)}
				</div>
				<div className="rs-strip-group">
					<span className="rs-strip-label">Debuffs</span>
					{RAID_EFFECTS.filter(e => e.debuff).map(tile)}
				</div>
			</div>,
			...(dps ? [dps] : []),
		);
	}

	// ---- The bench: presets and the report's builds ----

	private renderPalette() {
		const filter = this.benchFilter.value.trim().toLowerCase();
		const matches = (...words: string[]) => !filter || words.some(w => w.toLowerCase().includes(filter));
		const entry = (member: Omit<Member, 'id'>, label: string, icon: string, cls: Class, title: string, hint = '') => {
			const elem = (
				<button className="rs-bench-item" type="button" draggable={true} title={title}>
					<img className="rs-icon" src={icon} draggable={false} alt={iconAlt(cls)} />
					<span className={`rs-bench-name text-${cssClassForClass(cls)}`}>{label}</span>
					{hint ? <span className="rs-bench-hint">{hint}</span> : null}
				</button>
			) as HTMLButtonElement;
			elem.dataset.key = chipKey(member);
			elem.addEventListener('dragstart', e => {
				this.drag = { kind: 'new', member };
				e.dataTransfer!.effectAllowed = 'copy';
				e.dataTransfer!.setData('text/plain', 'rs-new');
				this.raidHost.classList.add('rs-dragging');
			});
			elem.addEventListener('dragend', () => {
				this.drag = null;
				this.raidHost.classList.remove('rs-dragging');
			});
			elem.addEventListener('click', () => this.addMember(member));
			return elem;
		};
		const group = (title: string, icon: Node, items: HTMLElement[]) =>
			items.length ? (
				<div className="rs-bench-group">
					<div className="rs-bench-heading">
						{icon}
						{title}
					</div>
					<div className="rs-bench-items">{items}</div>
				</div>
			) : null;
		const presetGroup = (title: string, role: RaidRole, faIcon: string) =>
			group(
				title,
				<i className={`fas ${faIcon} rs-bench-role`} />,
				presetsFor(role)
					.filter(p => matches(p.name, title))
					.map(p => entry({ preset: p.id }, p.name, p.icon, p.class, `${p.name}: ${p.gives}`)),
			);

		// The entry with the keyboard's focus keeps it when the bench is drawn again.
		const active = document.activeElement;
		const activeKey = active instanceof HTMLElement && this.benchList.contains(active) ? active.dataset.key : undefined;
		const byClass = new Map<string, ReportBuild[]>();
		this.data?.builds.forEach(b => {
			if (!byClass.has(b.className)) byClass.set(b.className, []);
			byClass.get(b.className)!.push(b);
		});
		const groups = [
			presetGroup('Tanks', RaidRole.RaidRoleTank, 'fa-shield-alt'),
			presetGroup('Healers', RaidRole.RaidRoleHealer, 'fa-plus'),
			...(this.data
				? [...byClass.entries()].map(([className, builds]) =>
						group(
							className,
							<img className="rs-icon" src={CLASS_ICONS[builds[0].class] ?? ''} alt="" />,
							builds
								.filter(b => matches(className, b.spec, buildLabel(b), b.gives))
								.map(b =>
									entry(
										{ buildId: b.id },
										buildLabel(b).replace(/^[^:]*: /, ''),
										buildIcon(b),
										b.class,
										`${buildLabel(b)} · ${fmt(b.published)} DPS in the report · gives ${b.gives}${b.summary ? `\n${b.summary}` : ''}`,
										fmt(b.published, 0),
									),
								),
						),
				  )
				: [<p className="rs-muted">Loading the report’s builds…</p>]),
		].filter((g): g is HTMLElement => !!g);
		this.benchList.replaceChildren(...(groups.length ? groups : [<p className="rs-muted">No build matches “{this.benchFilter.value}”.</p>]));
		if (activeKey) this.benchList.querySelector<HTMLElement>(`[data-key="${activeKey}"]`)?.focus();
	}

	// ---- Assignments ----

	// The job the last run's resolver gave, while the run is of this raid.
	private resolved(job: (a: Assignment) => boolean): Assignment | undefined {
		return this.freshRun?.resolution.assignments.find(job);
	}

	// A member as a row's heading on an assignment card: its icon, its name and where it stands.
	private who(member: Member, cell: number) {
		const cls = this.classOf(member);
		const info = member.preset ? presetInfo(member.preset) : undefined;
		return (
			<span className="rs-who" title={this.memberName(member)}>
				<img className="rs-icon" src={memberIcon(info, this.buildOf(member), cls)} alt={iconAlt(cls)} />
				<span className={`rs-who-name text-${cssClassForClass(cls)}`}>{this.slotName(member)}</span>
				<span className="rs-who-seat">
					{Math.floor(cell / PARTY_SIZE) + 1}.{(cell % PARTY_SIZE) + 1}
				</span>
			</span>
		);
	}

	// A member as a choice, for a cooldown's target.
	private memberChoice(member: Member, cell: number, hint = ''): Choice {
		const cls = this.classOf(member);
		const info = member.preset ? presetInfo(member.preset) : undefined;
		return {
			value: `${member.id}`,
			label: `${this.slotName(member)} (${Math.floor(cell / PARTY_SIZE) + 1}.${(cell % PARTY_SIZE) + 1})`,
			icon: memberIcon(info, this.buildOf(member), cls),
			labelClass: `text-${cssClassForClass(cls)}`,
			hint,
		};
	}

	private renderAssignments() {
		const members = this.members();
		const h = this.state.hand;
		const cards: Node[] = [];
		const card = (title: string, icon: string, body: Node[], note = '') => (
			<div className="rs-card">
				<div className="rs-card-head">
					<img className="rs-icon" src={icon} alt="" />
					<span className="rs-card-title">{title}</span>
				</div>
				{note ? <p className="rs-card-note">{note}</p> : null}
				{body}
			</div>
		);
		const AUTO: Choice = { value: '', label: 'Auto', faIcon: 'fa-wand-magic-sparkles', hint: 'the resolver picks' };
		const set = <T,>(record: Record<number, T>, id: number, v: string) => {
			if (v === '') delete record[id];
			else record[id] = parseInt(v) as T;
			this.assignmentChanged();
		};

		// Blessings: a row per paladin, a column per class in the raid, and a placeholder's judgement.
		const presentClasses = BLESSING_CLASSES.filter(c => members.some(m => !m.member.preset && this.classOf(m.member) == c));
		const paladins = members.filter(m => this.classOf(m.member) == Class.ClassPaladin);
		if (paladins.length) {
			const blessingChoices: Choice[] = [
				AUTO,
				...[Blessing.BlessingKings, Blessing.BlessingMight, Blessing.BlessingWisdom].map(b => ({
					value: `${b}`,
					label: `Greater Blessing of ${blessingName(b)}`,
					icon: BLESSING_ICONS[b],
				})),
				{ value: `${Blessing.BlessingNone}`, label: 'No blessing', faIcon: 'fa-ban' },
			];
			const judgementChoices: Choice[] = [
				AUTO,
				...[PaladinJudgement.PaladinJudgementCrusader, PaladinJudgement.PaladinJudgementWisdom].map(j => ({
					value: `${j}`,
					label: `Judgement of ${judgementName(j)}`,
					icon: JUDGEMENT_ICONS[j],
				})),
				{ value: `${PaladinJudgement.PaladinJudgementNone}`, label: 'No judgement', faIcon: 'fa-ban' },
			];
			const withJudgement = paladins.some(p => p.member.preset);
			const blessingOf = (a: Assignment | undefined) =>
				a?.job.oneofKind == 'blessing' ? blessingChoices.find(c => c.value == `${(a.job as any).blessing.blessing}`) : undefined;
			const judgementOf = (a: Assignment | undefined) =>
				a?.job.oneofKind == 'judgement' ? judgementChoices.find(c => c.value == `${(a.job as any).judgement}`) : undefined;
			cards.push(
				card(
					'Blessings',
					BLESSING_ICONS[Blessing.BlessingKings]!,
					[
						<div className="rs-table-wrap">
							<table className="rs-matrix">
								<thead>
									<tr>
										<th />
										{presentClasses.map(c => (
											<th title={`${classNames[c]}s`}>
												<img className="rs-icon" src={CLASS_ICONS[c]} alt={classNames[c]} />
											</th>
										))}
										{withJudgement ? <th className="rs-matrix-gap">Judge</th> : null}
									</tr>
								</thead>
								<tbody>
									{paladins.map(({ member, cell }) => (
										<tr dataset={{ holder: `${member.id}` }}>
											<th>{this.who(member, cell)}</th>
											{presentClasses.map(c => (
												<td>
													{assignmentPicker({
														label: `${this.slotName(member)}: blessing for ${classNames[c]}s`,
														choices: blessingChoices,
														value: h.blessings[member.id]?.[c] !== undefined ? `${h.blessings[member.id]![c]}` : '',
														resolved: blessingOf(
															this.resolved(a => a.holder == member.id && a.job.oneofKind == 'blessing' && a.job.blessing.class == c),
														),
														onChange: v => {
															const mine = (h.blessings[member.id] ??= {});
															if (v === '') delete mine[c];
															else mine[c] = parseInt(v);
															this.assignmentChanged();
														},
													})}
												</td>
											))}
											{withJudgement ? (
												<td className="rs-matrix-gap">
													{member.preset ? (
														assignmentPicker({
															label: `${this.slotName(member)}: judgement`,
															choices: judgementChoices,
															value: h.judgements[member.id] !== undefined ? `${h.judgements[member.id]}` : '',
															resolved: judgementOf(this.resolved(a => a.holder == member.id && a.job.oneofKind == 'judgement')),
															onChange: v => set(h.judgements, member.id, v),
														})
													) : (
														<span className="rs-muted" title="A DPS paladin judges by its duty, under Duties">
															duty
														</span>
													)}
												</td>
											) : null}
										</tr>
									))}
								</tbody>
							</table>
						</div>,
					],
					presentClasses.length ? '' : 'Add DPS members to give them blessings.',
				),
			);
		}

		// Totems: a row per shaman placeholder, a column per element. A DPS shaman's rotation drops its own.
		const shamans = members.filter(m => m.member.preset && presetInfo(m.member.preset)?.shaman);
		const dpsShamans = members.filter(m => !m.member.preset && this.classOf(m.member) == Class.ClassShaman);
		if (shamans.length || dpsShamans.length) {
			const totemChoices = (totems: RaidTotem[]): Choice[] => [
				AUTO,
				...totems.map(t => ({ value: `${t}`, label: `${totemName(t)} Totem`, icon: TOTEM_ICONS[t] })),
				{ value: `${RaidTotem.RaidTotemNone}`, label: 'No totem', faIcon: 'fa-ban' },
			];
			cards.push(
				card(
					'Totems',
					TOTEM_ICONS[RaidTotem.RaidTotemWindfury]!,
					[
						shamans.length ? (
							<div className="rs-table-wrap">
								<table className="rs-matrix">
									<thead>
										<tr>
											<th />
											{TOTEM_ELEMENTS.map(e => (
												<th>{e.label}</th>
											))}
										</tr>
									</thead>
									<tbody>
										{shamans.map(({ member, cell }) => (
											<tr dataset={{ holder: `${member.id}` }}>
												<th>{this.who(member, cell)}</th>
												{TOTEM_ELEMENTS.map(({ element: el, label, totems }) => {
													const choices = totemChoices(totems);
													const a = this.resolved(a => a.holder == member.id && a.job.oneofKind == 'totem' && assignedElement(a.job.totem) == el);
													return (
														<td>
															{assignmentPicker({
																label: `${this.slotName(member)}: ${label.toLowerCase()} totem`,
																choices,
																value: h.totems[member.id]?.[el] !== undefined ? `${h.totems[member.id]![el]}` : '',
																resolved: a?.job.oneofKind == 'totem' ? choices.find(c => c.value == `${(a.job as any).totem.totem}`) : undefined,
																onChange: v => {
																	const mine = (h.totems[member.id] ??= {});
																	if (v === '') delete mine[el];
																	else mine[el] = parseInt(v);
																	this.assignmentChanged();
																},
															})}
														</td>
													);
												})}
											</tr>
										))}
									</tbody>
								</table>
							</div>
						) : null,
						...dpsShamans.map(({ member, cell }) => (
							<div className="rs-job-row rs-job-row-fixed" dataset={{ holder: `${member.id}` }}>
								{this.who(member, cell)}
								<span className="rs-muted">drops its own totems</span>
							</div>
						)),
					].filter((n): n is HTMLElement => !!n),
				),
			);
		}

		// Duties: a DPS member with duty variants holds one of them, or none. A placeholder gives its own
		// (a druid's Faerie Fire, the warrior tank's Sunder Armor) without one.
		const dutyRows: Node[] = [];
		members.forEach(({ member, cell }) => {
			if (member.preset) return;
			const duties = (this.buildOf(member)?.duties ?? []).map(d => d.duty!).filter(d => d !== undefined);
			if (!duties.length) return;
			const choices: Choice[] = [
				AUTO,
				...duties.map(d => ({ value: `${d}`, label: dutyName(d), icon: DUTY_ICONS[d] })),
				{ value: `${Duty.DutyNone}`, label: 'No duty', faIcon: 'fa-ban', hint: 'the most DPS' },
			];
			const a = this.resolved(a => a.holder == member.id && a.job.oneofKind == 'duty');
			dutyRows.push(
				<div className="rs-job-row" dataset={{ holder: `${member.id}` }}>
					{this.who(member, cell)}
					{assignmentPicker({
						label: `${this.slotName(member)}: duty`,
						choices,
						value: h.duties[member.id] !== undefined ? `${h.duties[member.id]}` : '',
						resolved: a?.job.oneofKind == 'duty' ? choices.find(c => c.value == `${(a.job as any).duty}`) : undefined,
						onChange: v => set(h.duties, member.id, v),
					})}
				</div>,
			);
		});
		if (dutyRows.length) cards.push(card('Duties', DUTY_ICONS[Duty.DutySunderUpkeep]!, [<div className="rs-job-rows">{dutyRows}</div>], 'A duty costs its holder some DPS.'));

		// Cooldowns: Power Infusion from a Discipline priest, Innervate from a Restoration druid (or a Feral
		// tank with it switched on). Only who gets them is chosen.
		// Power Infusion goes to the casters its 20% magic damage is for, Innervate to the mana users, as the
		// resolver gives them (sim/compsearch, isCaster and usesManaPlayer).
		const dpsMembers = members.filter(m => !m.member.preset);
		const casters = dpsMembers.filter(m => this.isCaster(m.member));
		const manaUsers = dpsMembers.filter(m => this.usesMana(m.member));
		// The resolver's automatic Power Infusions and Innervates name their target, so the n-th holder's
		// is the n-th one.
		const piHolders = members.filter(m => m.member.preset && presetInfo(m.member.preset)?.powerInfusion).map(m => m.member.id);
		const innervateHolders = members
			.filter(m => m.member.preset && (presetInfo(m.member.preset)?.innervate || m.member.presetInnervate))
			.map(m => m.member.id);
		const cooldown = (holder: { member: Member; cell: number }, pi: boolean) => {
			const map = pi ? h.powerInfusion : h.innervate;
			const targets = [...(pi ? casters : manaUsers)];
			// A target set by hand that the list doesn't offer (a saved raid from before it was narrowed) stays
			// in it, marked, so the picker shows what Run sends.
			const current = dpsMembers.find(t => t.member.id == map[holder.member.id]);
			if (current && !targets.includes(current)) targets.push(current);
			const nth = (pi ? piHolders : innervateHolders).indexOf(holder.member.id);
			const resolved = this.freshRun?.resolution.assignments.filter(a => a.job.oneofKind == 'cooldown' && a.job.cooldown.powerInfusion == pi)[nth];
			const choices: Choice[] = [
				AUTO,
				...targets.map(t =>
					this.memberChoice(t.member, t.cell, t != current || (pi ? casters : manaUsers).includes(t) ? '' : pi ? 'no magic damage to boost' : 'no mana'),
				),
			];
			const target = resolved?.job.oneofKind == 'cooldown' ? resolved.job.cooldown.target : undefined;
			return (
				<span className="rs-cooldown">
					<img className="rs-icon" src={pi ? POWER_INFUSION_ICON : INNERVATE_ICON} alt="" title={pi ? 'Power Infusion' : 'Innervate'} />
					<i className="fas fa-arrow-right rs-muted" />
					{assignmentPicker({
						label: `${this.slotName(holder.member)}: ${pi ? 'Power Infusion' : 'Innervate'} on`,
						choices,
						value: map[holder.member.id] !== undefined ? `${map[holder.member.id]}` : '',
						resolved: target !== undefined ? choices.find(c => c.value == `${target}`) : undefined,
						onChange: v => set(map, holder.member.id, v),
					})}
				</span>
			);
		};
		const cooldownRows: Node[] = [];
		members
			.filter(m => m.member.preset && presetInfo(m.member.preset)?.powerInfusion)
			.forEach(m =>
				cooldownRows.push(
					<div className="rs-job-row" dataset={{ holder: `${m.member.id}` }}>
						{this.who(m.member, m.cell)}
						{cooldown(m, true)}
					</div>,
				),
			);
		members
			.filter(m => m.member.preset && (presetInfo(m.member.preset)?.innervate || presetInfo(m.member.preset)?.innervateOptional))
			.forEach(m => {
				const info = presetInfo(m.member.preset!)!;
				cooldownRows.push(
					<div className="rs-job-row" dataset={{ holder: `${m.member.id}` }}>
						{this.who(m.member, m.cell)}
						{info.innervate || m.member.presetInnervate ? cooldown(m, false) : null}
						{info.innervateOptional
							? check('Innervate', !!m.member.presetInnervate, v => {
									m.member.presetInnervate = v;
									// Without its Innervate, a target set by hand goes too: the resolver refuses a holder
									// with nothing to give.
									if (!v) delete h.innervate[m.member.id];
									this.changed();
							  })
							: null}
					</div>,
				);
			});
		if (cooldownRows.length) cards.push(card('Cooldowns', POWER_INFUSION_ICON, [<div className="rs-job-rows">{cooldownRows}</div>]));

		this.renderAssignmentWarnings();
		this.assignmentsHost.replaceChildren(
			<p className="rs-legend">
				<span className="rs-legend-item">
					<span className="rs-pick rs-pick-hand rs-legend-swatch" /> set by hand
				</span>
				<span className="rs-legend-item">
					<span className="rs-pick rs-pick-auto rs-legend-swatch" /> Auto: the resolver picks, and a run shows its pick
				</span>
			</p>,
			this.warningsHost,
			cards.length ? (
				<div className="rs-cards">{cards}</div>
			) : (
				empty('Nobody here holds an assignment yet. Paladins, shamans, priests, druids, warriors and warlocks do.')
			),
		);
	}

	// What the jobs set by hand waste: two holders of the one minor armor slot, or a Windfury Totem by hand
	// next to another shaman's (the raid gets one, so the second adds nothing). Drawn again on each change.
	private renderAssignmentWarnings() {
		const h = this.state.hand;
		const members = this.members();
		const warnings: string[] = [];
		const minorArmor = members.filter(m => h.duties[m.member.id] == Duty.DutyFaerieFire || h.duties[m.member.id] == Duty.DutyCurseOfRecklessness);
		if (minorArmor.length > 1) {
			warnings.push(
				`${minorArmor.map(m => `${this.memberShortName(m.member, m.cell)} holds ${dutyName(h.duties[m.member.id])}`).join(' and ')} by hand, but ` +
					'Faerie Fire and Curse of Recklessness fill one minor armor slot: only one of them counts, and the rest pay their duty cost for nothing.',
			);
		}
		const handWindfury = members.filter(m => h.totems[m.member.id]?.air == RaidTotem.RaidTotemWindfury);
		const otherWindfury = (id: number) =>
			members.filter(
				m =>
					m.member.id != id &&
					(h.totems[m.member.id]?.air == RaidTotem.RaidTotemWindfury || this.buildOf(m.member)?.gives.includes('Windfury Totem')),
			);
		handWindfury.forEach(m => {
			const others = otherWindfury(m.member.id);
			// Two placeholders by hand make one warning, from the first.
			if (!others.length || others.some(o => handWindfury.includes(o) && o.cell < m.cell)) return;
			warnings.push(
				`${this.memberShortName(m.member, m.cell)}’s Windfury Totem by hand duplicates ${others
					.map(o => this.memberShortName(o.member, o.cell))
					.join(' and ')}’s: the raid gets one Windfury Totem, so Grace of Air would add more.`,
			);
		});
		this.warningsHost.replaceChildren(...warnings.map(w => <li>{w}</li>));
	}

	// A member Power Infusion's 20% magic damage is for (sim/compsearch, isCaster).
	private isCaster(member: Member): boolean {
		const build = this.buildOf(member);
		if (!build) return false;
		const spec = Object.values(build.players)[0]?.spec.oneofKind;
		switch (build.class) {
			case Class.ClassMage:
			case Class.ClassWarlock:
			case Class.ClassPriest:
				return true;
			case Class.ClassDruid:
				return spec == 'balanceDruid';
			case Class.ClassShaman:
				return spec == 'elementalShaman';
		}
		return false;
	}

	// A member Innervate is for (sim/compsearch, usesManaPlayer): not a warrior, a rogue or a cat.
	private usesMana(member: Member): boolean {
		const build = this.buildOf(member);
		if (!build) return false;
		if (build.class == Class.ClassWarrior || build.class == Class.ClassRogue) return false;
		return Object.values(build.players)[0]?.spec.oneofKind != 'feralDruid';
	}

	private assignmentRow(member: Member, cell: number, title: string, controls: Node[]) {
		const cls = this.classOf(member);
		return (
			<div className="rs-assignment">
				<div className="rs-assignment-header">
					<span className={`text-${cssClassForClass(cls)}`}>{this.memberShortName(member, cell)}</span>
					<span className="rs-muted">{title}</span>
				</div>
				<div className="rs-assignment-controls">{controls}</div>
			</div>
		);
	}

	// ---- The encounter ----

	private initEncounter() {
		const sim = this.simUI.sim;
		const pickerHost = (<div className="rs-encounter-picker" />) as HTMLElement;
		this.encounterHost.replaceChildren(this.bossHost, pickerHost);
		this.initBossSelect();

		sim.waitForInit().then(() => {
			if (this.state.encounter) {
				try {
					sim.encounter.fromProto(TypedEvent.nextEventID(), EncounterProto.fromJson(this.state.encounter as any, { ignoreUnknownFields: true }));
				} catch (e) {
					console.warn('Failed to read the saved encounter:', e);
				}
			}
			new EncounterPicker(pickerHost, sim.encounter, { showExecuteProportion: true }, this.simUI);
			sim.encounter.changeEmitter.on(eventID => {
				if (eventID != this.presetEvent && this.state.boss != 'Custom') {
					this.state.boss = 'Custom';
					this.initBossSelect();
				}
				this.saveState();
				this.markStale();
			});
			if (this.state.iterations) sim.setIterations(TypedEvent.nextEventID(), this.state.iterations);
			sim.iterationsChangeEmitter.on(() => {
				this.state.iterations = sim.getIterations();
				this.saveState();
			});
		});
	}

	// The boss preset's select, drawn again when a loaded comp sets it.
	private initBossSelect() {
		const bosses: Array<{ value: Boss; label: string; title: string }> = [
			{ value: 'Generic', label: 'Generic', title: 'The DPS Report’s level 63 boss, in the natural flask zone' },
			{ value: 'Onyxia', label: 'Onyxia', title: 'A Dragonkin, outside the flask zone' },
			{ value: 'Custom', label: 'Custom', title: 'As set below. Any change below makes the fight custom.' },
		];
		const bossControl = segmented(
			bosses,
			this.state.boss,
			v => {
				this.state.boss = v as Boss;
				this.applyBoss(this.state.boss);
				this.assignmentChanged();
				this.initBossSelect();
			},
			'Boss',
		);
		this.bossHost.replaceChildren(
			toolbarItem('Boss', bossControl),
			<span className="rs-muted rs-boss-note">{bosses.find(b => b.value == this.state.boss)?.title}</span>,
		);
	}

	private applyBoss(boss: Boss) {
		if (boss == 'Custom' || !this.data) return;
		const encounter = EncounterProto.clone(this.data.encounter);
		if (boss == 'Onyxia') {
			encounter.naturalFlaskZone = false;
			encounter.targets.forEach(t => (t.mobType = MobType.MobTypeDragonkin));
		} else {
			encounter.naturalFlaskZone = true;
		}
		this.setPresetEncounter(encounter);
	}

	private setPresetEncounter(encounter: EncounterProto) {
		this.presetEvent = TypedEvent.nextEventID();
		this.simUI.sim.encounter.fromProto(this.presetEvent, encounter);
	}

	// ---- The raid as data: the sim history, the presets and links ----

	// The raid as it stands, its fight included, as a copy.
	getState(): RaidState {
		const state = { ...this.state, encounter: EncounterProto.toJson(this.simUI.sim.encounter.toProto()) };
		return JSON.parse(JSON.stringify(state));
	}

	// Reads a raid kept as JSON (a preset, a run in the sim history or an import), filling in what a raid
	// from before a field was added lacks. It throws when it isn't a raid of this page.
	static readState(raw: unknown): RaidState {
		const s = raw as Partial<EditorState> | null;
		if (!s || typeof s != 'object' || !s.size || !RAID_SIZES[s.size] || !Array.isArray(s.cells) || s.cells.length != s.size)
			throw new Error('It isn’t a raid of the raid sim page.');
		const cells = s.cells.map(c => (c && typeof c == 'object' && typeof c.id == 'number' ? c : null));
		const nextId = Math.max(s.nextId ?? 1, ...cells.map(c => (c?.id ?? 0) + 1));
		return {
			...s,
			size: s.size,
			cells,
			nextId,
			faction: s.faction ?? Faction.Horde,
			campBuffs: !!s.campBuffs,
			boss: s.boss ?? 'Generic',
			hand: { ...emptyHand(), ...s.hand },
		};
	}

	// Puts a raid in place: a preset, a run's raid from the sim history, or an import. The iterations stay as
	// they are. Undo puts back the raid it replaced. It throws when it can't read the raid.
	applyState(raw: unknown, what: string) {
		const next = CompEditor.readState(raw);
		const before = this.snapshot();
		next.iterations = this.state.iterations;
		this.state = next;
		this.moving = null;
		this.setStateEncounter();
		this.initBossSelect();
		this.forgetResults();
		this.renderAll();
		this.changed();
		this.setUndo(before, what);
		this.setStatus('');
	}

	// The fight the raid keeps, or its boss's when it keeps none.
	private setStateEncounter() {
		if (this.state.encounter) {
			try {
				this.setPresetEncounter(EncounterProto.fromJson(this.state.encounter as any, { ignoreUnknownFields: true }));
				return;
			} catch (e) {
				console.warn('Failed to read the raid’s encounter:', e);
			}
		}
		this.applyBoss(this.state.boss);
	}

	// Whether two raids are the same: the same members in the same slots, with the same assignments, faction,
	// camps and fight. The member ids and the iterations don't count.
	static sameRaid(a: RaidState, b: RaidState): boolean {
		return stableJson(canonicalRaid(a)) == stableJson(canonicalRaid(b));
	}

	// What changed from one raid to the other, in lines like 'Group 2, slot 3: Fire → Beast Mastery', under
	// Raid, Members, Assignments and Encounter.
	describeChanges(from: RaidState, to: RaidState): ChangeCategory[] {
		const categories: ChangeCategory[] = [];
		const add = (name: string, lines: string[]) => lines.length && categories.push({ name, lines });
		const line = (name: string, a: string, b: string) => (a == b ? [] : [`${name}: ${a} → ${b}`]);
		const onOff = (on: boolean) => (on ? 'on' : 'off');

		add('Raid', [
			...line('Size', `${from.size}`, `${to.size}`),
			...line('Faction', factionName(from.faction), factionName(to.faction)),
			...line('Camp buffs', onOff(from.campBuffs), onOff(to.campBuffs)),
			...line('Boss', from.boss, to.boss),
		]);

		const slotName = (cell: number) => `Group ${Math.floor(cell / PARTY_SIZE) + 1}, slot ${(cell % PARTY_SIZE) + 1}`;
		const memberAt = (state: RaidState, cell: number) => {
			const m = state.cells[cell];
			return m ? this.memberName(m) : 'empty';
		};
		const cells = Math.max(from.size, to.size);
		add(
			'Members',
			[...Array(cells).keys()].flatMap(cell => line(slotName(cell), memberAt(from, cell), memberAt(to, cell))),
		);

		const fromJobs = this.jobsOf(from);
		const toJobs = this.jobsOf(to);
		add(
			'Assignments',
			[...new Set([...fromJobs.keys(), ...toJobs.keys()])].flatMap(job => line(job, fromJobs.get(job) ?? 'Auto', toJobs.get(job) ?? 'Auto')),
		);

		const fromFight = flattenJson(from.encounter ?? {});
		const toFight = flattenJson(to.encounter ?? {});
		add(
			'Encounter',
			[...new Set([...fromFight.keys(), ...toFight.keys()])].flatMap(key => line(humanize(key), fromFight.get(key) ?? '-', toFight.get(key) ?? '-')),
		);
		return categories;
	}

	// The raid's assignments set by hand, by who holds them, like 'Holy paladin (4.3): Hunter blessing' →
	// 'Kings'.
	private jobsOf(state: RaidState): Map<string, string> {
		const jobs = new Map<string, string>();
		const cellOf = new Map<number, number>();
		state.cells.forEach((m, cell) => m && cellOf.set(m.id, cell));
		const holder = (id: number) => {
			const cell = cellOf.get(id);
			return cell === undefined ? undefined : this.memberShortName(state.cells[cell]!, cell);
		};
		const each = <T,>(record: Record<number, T>, put: (who: string, value: T) => void) =>
			Object.entries(record).forEach(([id, value]) => {
				const who = holder(Number(id));
				if (who) put(who, value);
			});
		const h = state.hand;
		each(h.blessings, (who, byClass) =>
			Object.entries(byClass).forEach(([c, b]) => jobs.set(`${who}: ${classNames[Number(c) as Class]} blessing`, blessingName(b as Blessing))),
		);
		each(h.judgements, (who, j) => jobs.set(`${who}: Judgement`, judgementName(j)));
		each(h.totems, (who, totems) =>
			TOTEM_ELEMENTS.forEach(({ element, label }) => {
				const t = totems[element];
				if (t !== undefined) jobs.set(`${who}: ${label} totem`, t == RaidTotem.RaidTotemNone ? 'none' : totemName(t));
			}),
		);
		each(h.duties, (who, d) => jobs.set(`${who}: Duty`, dutyName(d)));
		each(h.powerInfusion, (who, target) => jobs.set(`${who}: Power Infusion`, holder(target) ?? 'nobody'));
		each(h.innervate, (who, target) => jobs.set(`${who}: Innervate`, holder(target) ?? 'nobody'));
		return jobs;
	}

	// The built-in raids: each size's new raid, and the best comp of each scenario of the committed searches.
	async builtInPresets(): Promise<RaidPresetEntry[]> {
		this.data ??= await loadRaidSimData();
		const generic = (state: EditorState): EditorState => ({ ...state, boss: 'Generic', encounter: EncounterProto.toJson(this.genericEncounter()) });
		const presets: RaidPresetEntry[] = [10, 20, 40].map(size => ({
			name: `${size} players`,
			group: 'New raids',
			tooltip: `${size} players: the usual tanks and healers, no DPS yet.`,
			state: generic(this.defaultState(size)),
		}));
		for (const result of await loadRaidSimPresets()) {
			for (const scenario of result.scenarios) {
				const comp = scenario.comps[0];
				if (!comp) continue;
				const { state } = this.stateFromResult(comp, scenario, result.request);
				const size = result.request?.size ?? state.size;
				const faction = factionName(scenario.faction);
				const dps = comp.confirmed ? `${fmt(comp.raidSimDps)} raid DPS` : `${fmt(comp.dps)} score`;
				presets.push({
					name: `${faction} ${size}${scenario.campBuffs ? ', camps' : ''}`,
					group: 'Best comps',
					tooltip: `The comp search’s best ${size}-player ${faction} raid, camp buffs ${scenario.campBuffs ? 'on' : 'off'} · ${dps}`,
					state: { ...state, encounter: state.encounter ?? EncounterProto.toJson(this.genericEncounter()) },
				});
			}
		}
		return presets;
	}

	// The report's fight, the Generic boss.
	private genericEncounter(): EncounterProto {
		const encounter = EncounterProto.clone(this.data!.encounter);
		encounter.naturalFlaskZone = true;
		return encounter;
	}

	// ---- Load a comp from a search result ----

	async loadFromResult(comp: CompResult, scenario: CompScenarioResult, request?: CompSearchRequest) {
		// The members are the data file's builds, so it has to be here first.
		if (!this.data) {
			try {
				this.data = await loadRaidSimData();
			} catch (e) {
				this.setError('Can’t load this comp: the report’s builds didn’t load', `${e}`);
				return;
			}
		}
		const before = this.snapshot();
		const { state, unknown } = this.stateFromResult(comp, scenario, request);
		this.state = state;
		this.setStateEncounter();
		this.initBossSelect();
		this.forgetResults();
		this.saveState();
		this.renderAll();
		this.setUndo(before, 'the loaded comp');
		// A confirmed comp has its raid sim DPS; otherwise the search only scored it.
		const there = comp.confirmed
			? `${fmt(comp.raidSimDps)} raid sim DPS there`
			: `a score of ${fmt(comp.dps)} there (its members’ single-player sims added up; Run sims it as a raid)`;
		const loaded = `Loaded comp from the search (${factionName(scenario.faction)}, camps ${
			scenario.campBuffs ? 'on' : 'off'
		}): ${there}. Undo puts back the raid it replaced.`;
		if (unknown.length) {
			this.setStatus(
				`${loaded} This page’s builds don’t have ${[...new Set(unknown)].join(
					', ',
				)}, so Run and Member values refuse this raid until they’re replaced.`,
				true,
			);
		} else {
			this.setStatus(loaded);
		}
	}

	// A search's comp as the editor's raid, with its fight when the search's request has one, and the builds
	// this page's data doesn't have. It needs the data file.
	private stateFromResult(comp: CompResult, scenario: CompScenarioResult, request?: CompSearchRequest): { state: EditorState; unknown: string[] } {
		const unknown: string[] = [];
		const size = request?.size && RAID_SIZES[request.size] ? request.size : comp.members.length <= 10 ? 10 : comp.members.length <= 20 ? 20 : 40;
		const state: EditorState = {
			...this.state,
			size,
			cells: Array(size).fill(null),
			nextId: 1,
			faction: scenario.faction,
			campBuffs: scenario.campBuffs,
			hand: emptyHand(),
		};
		const placed: Array<{ member: Member; result: (typeof comp.members)[number] }> = [];
		const nextInParty = new Map<number, number>();
		comp.members.forEach(m => {
			const party = Math.max(0, m.party - 1);
			const index = nextInParty.get(party) ?? 0;
			nextInParty.set(party, index + 1);
			const cell = party * PARTY_SIZE + index;
			if (cell >= size || index >= PARTY_SIZE) return;
			const member: Member = { id: state.nextId++ };
			if (m.role != RaidRole.RaidRoleDps) {
				member.preset = m.preset;
				if (presetInfo(m.preset)?.paladin && m.judgement != PaladinJudgement.PaladinJudgementNone) state.hand.judgements[member.id] = m.judgement;
				if (m.totems.length) {
					const totems: Partial<Record<TotemElement, RaidTotem>> = {
						earth: RaidTotem.RaidTotemNone,
						air: RaidTotem.RaidTotemNone,
						water: RaidTotem.RaidTotemNone,
						fire: RaidTotem.RaidTotemNone,
					};
					m.totems.forEach(t => {
						const e = totemElement(t);
						if (e) totems[e] = t;
					});
					state.hand.totems[member.id] = totems;
				}
			} else {
				// A duty variant holds its duty; the editor keeps the base build and the duty by hand. A base
				// build that could hold a duty holds none by hand, as in the search, so the resolver doesn't
				// give it one.
				const found = this.data?.byId.get(m.buildId);
				member.buildId = found?.build.id ?? m.buildId;
				if (!found) unknown.push(m.buildId);
				const duty = m.duties.find(d => d != Duty.DutyNone);
				if (duty !== undefined) state.hand.duties[member.id] = duty;
				else if (found?.build.duties.length) state.hand.duties[member.id] = Duty.DutyNone;
			}
			state.cells[cell] = member;
			placed.push({ member, result: m });
		});

		// The blessings each member gets, as the search filled them, back into each paladin's blessing per
		// class: the k-th paladin gives each class its k-th blessing.
		const paladins = placed.filter(p => (p.member.preset ? presetInfo(p.member.preset)?.class : this.buildOf(p.member)?.class) == Class.ClassPaladin);
		BLESSING_CLASSES.forEach(c => {
			const got = placed.find(p => !p.member.preset && this.buildOf(p.member)?.class == c)?.result.blessings ?? [];
			if (!placed.some(p => !p.member.preset && this.buildOf(p.member)?.class == c)) return;
			paladins.forEach((p, k) => {
				(state.hand.blessings[p.member.id] ??= {})[c] = got[k] ?? Blessing.BlessingNone;
			});
		});

		// Power Infusion and Innervate: whoever gets them, from the first placeholder that has them.
		const piHolder = placed.find(p => p.member.preset && presetInfo(p.member.preset)?.powerInfusion);
		const piTarget = placed.find(p => p.result.powerInfusion);
		if (piHolder && piTarget) state.hand.powerInfusion[piHolder.member.id] = piTarget.member.id;
		const innervateTarget = placed.find(p => p.result.innervate);
		if (innervateTarget) {
			const holder =
				placed.find(p => p.member.preset && presetInfo(p.member.preset)?.innervate) ??
				placed.find(p => p.member.preset && presetInfo(p.member.preset)?.innervateOptional);
			if (holder) {
				if (presetInfo(holder.member.preset!)?.innervateOptional) holder.member.presetInnervate = true;
				state.hand.innervate[holder.member.id] = innervateTarget.member.id;
			}
		}

		const e = request?.encounter;
		if (e) {
			const { encounter, boss } = this.searchEncounter(e);
			state.encounter = EncounterProto.toJson(encounter);
			state.boss = boss;
		}
		return { state, unknown };
	}

	// A search's fight, as tools/compsearch runs it (sim/compsearch, compOf): the report's boss with the
	// search's length, its health for a health fight, its creature type and its natural flask zone.
	private searchEncounter(e: CompEncounter): { encounter: EncounterProto; boss: Boss } {
		const encounter = EncounterProto.clone(this.data!.encounter);
		encounter.duration = e.duration || 180;
		encounter.durationVariation = e.durationVariation;
		encounter.useHealth = e.useHealth;
		encounter.naturalFlaskZone = e.naturalFlaskZone;
		encounter.targets.forEach(t => {
			if (e.mobType != MobType.MobTypeUnknown) t.mobType = e.mobType;
			if (e.useHealth && e.bossHealth > 0) {
				while (t.stats.length <= Stat.StatHealth) t.stats.push(0);
				t.stats[Stat.StatHealth] = e.bossHealth;
			}
		});
		const generic = e.mobType == MobType.MobTypeUnknown && e.naturalFlaskZone;
		const onyxia = e.mobType == MobType.MobTypeDragonkin && !e.naturalFlaskZone;
		const boss = (e.name || 'Generic') == 'Generic' && generic ? 'Generic' : e.name == 'Onyxia' && onyxia ? 'Onyxia' : 'Custom';
		return { encounter, boss };
	}

	// ---- Run ----

	// The comp as the resolver takes it.
	// It refuses a raid with a member it can't build (a build this page's data doesn't have, or one with no
	// player), naming them, rather than run the raid without them.
	async buildComp(): Promise<RaidComp> {
		const sim = this.simUI.sim;
		const unbuildable: string[] = [];
		await sim.waitForInit();
		const db = sim.db;
		const faction = this.state.faction;
		const encounter = sim.encounter.toProto();
		const mobType = encounter.targets[0]?.mobType ?? MobType.MobTypeUnknown;

		const playerOf = (variant: BuildVariant, name: string): Player | undefined => {
			const p = variant.players[faction] ?? Object.values(variant.players)[0];
			if (!p) return undefined;
			const player = Player.clone(p);
			player.name = name;
			// The WebAssembly lib has no item database of its own: each player brings its items.
			if (player.equipment) player.database = db.lookupEquipmentSpec(player.equipment).toDatabase();
			return player;
		};

		const slots: CompSlot[] = [];
		this.members().forEach(({ member, cell }) => {
			const party = Math.floor(cell / PARTY_SIZE);
			const index = cell % PARTY_SIZE;
			const name = this.memberShortName(member, cell);
			if (member.preset) {
				const info = presetInfo(member.preset);
				slots.push(
					CompSlot.create({
						id: member.id,
						party,
						index,
						role: info?.role ?? RaidRole.RaidRoleHealer,
						member: { oneofKind: 'placeholder', placeholder: member.preset },
						name,
						presetInnervate: !!member.presetInnervate,
					}),
				);
				return;
			}
			const build = this.buildOf(member);
			if (!build) {
				unbuildable.push(`${name} isn’t one of this page’s builds`);
				return;
			}
			// Against a creature type the build has a variant for, it runs that variant, as the search does.
			const variant = build.encounters.find(v => v.mobType == mobType) ?? build;
			const base = playerOf(variant, name);
			if (!base) {
				unbuildable.push(`${name}: its build has no player`);
				return;
			}
			slots.push(
				CompSlot.create({
					id: member.id,
					party,
					index,
					role: RaidRole.RaidRoleDps,
					member: {
						oneofKind: 'build',
						build: CompMemberBuild.create({
							base,
							duties: build.duties.map(d => ({ duty: d.duty!, player: playerOf(d, name)! })).filter(d => d.player && d.duty !== undefined),
						}),
					},
					reportId: variant.id,
					name,
				}),
			);
		});

		if (unbuildable.length) {
			throw new Error(
				`${unbuildable.length == 1 ? 'A member' : `${unbuildable.length} members`} can’t be built: ${unbuildable.join('; ')}. Replace or remove ${
					unbuildable.length == 1 ? 'it' : 'them'
				}.`,
			);
		}

		const h = this.state.hand;
		const assignments: Assignment[] = [];
		const has = (id: number) => slots.some(s => s.id == id);
		const hand = (holder: number, job: Assignment['job']) => has(holder) && assignments.push(Assignment.create({ holder, setByHand: true, job }));
		Object.entries(h.blessings).forEach(([holder, byClass]) =>
			Object.entries(byClass).forEach(([c, blessing]) =>
				hand(parseInt(holder), { oneofKind: 'blessing', blessing: { class: parseInt(c), blessing: blessing as Blessing } }),
			),
		);
		Object.entries(h.judgements).forEach(([holder, judgement]) => hand(parseInt(holder), { oneofKind: 'judgement', judgement }));
		// A shaman's totems: only the elements set by hand, each with its element, so none drops nothing
		// there and the elements left on Auto stay the resolver's.
		Object.entries(h.totems).forEach(([holder, byElement]) =>
			TOTEM_ELEMENTS.forEach(({ element, proto }) => {
				const totem = byElement[element];
				if (totem !== undefined) hand(parseInt(holder), { oneofKind: 'totem', totem: { totem, element: proto } });
			}),
		);
		Object.entries(h.duties).forEach(([holder, duty]) => hand(parseInt(holder), { oneofKind: 'duty', duty }));
		// A cooldown goes only from a holder that has it to give: a saved raid can name one that no longer does
		// (a Feral tank with its Innervate switched off), which the resolver would refuse.
		const gives = (holder: number, pi: boolean) => {
			const m = this.memberById(holder)?.member;
			const info = m?.preset ? presetInfo(m.preset) : undefined;
			return !!info && (pi ? !!info.powerInfusion : !!info.innervate || (!!info.innervateOptional && !!m!.presetInnervate));
		};
		Object.entries(h.powerInfusion).forEach(
			([holder, target]) =>
				has(target) && gives(parseInt(holder), true) && hand(parseInt(holder), { oneofKind: 'cooldown', cooldown: { powerInfusion: true, target } }),
		);
		Object.entries(h.innervate).forEach(
			([holder, target]) =>
				has(target) && gives(parseInt(holder), false) && hand(parseInt(holder), { oneofKind: 'cooldown', cooldown: { powerInfusion: false, target } }),
		);

		return RaidComp.create({
			size: this.state.size,
			slots,
			encounter,
			assignments,
			faction,
			campBuffs: this.state.campBuffs,
			assumptions: RaidAssumptions.create(),
			simOptions: SimOptions.create({
				iterations: sim.getIterations(),
				randomSeed: BigInt(sim.getFixedRngSeed() || Math.floor(Math.random() * (2 ** 32 - 1))),
				debugFirstIteration: true,
				ruleset: sim.getRuleset(),
			}),
		});
	}

	async run() {
		const sim = this.simUI.sim;
		if (this.running) {
			this.stopping = true;
			await sim.signalManager.abortType(RequestTypes.RaidSim);
			return;
		}
		if (!this.data) {
			this.setStatus('The report’s builds haven’t loaded yet.', true);
			return;
		}
		if (this.valuesAbort) {
			this.setStatus('Member values are running: abort them or wait for them before a run.', true);
			return;
		}
		if (!this.members().some(m => !m.member.preset)) {
			this.setStatus('The raid has no DPS member to sim. Add some from the report’s builds.', true);
			return;
		}
		this.running = true;
		this.stopping = false;
		this.setRunning(true);
		// The sidebar's progress bar, as on a spec page. It slides without filling while the resolver runs.
		const progressBar = new SimProgress('', this.simUI.iterationsPicker.querySelector<HTMLElement>('.form-label') ?? undefined);
		this.simUI.iterationsPicker.appendChild(progressBar.rootElem);
		this.simUI.simActionsContainer.classList.add('running');
		// The sim history keeps the raid as it was when Run was pressed, with the run's results.
		const history = this.simUI.resultsManager.simHistory;
		const raid = this.getState();
		history?.startRun();
		const revision = this.revision;
		// A worker's last progress can come after the run ends or is stopped: the sidebar keeps none of it then.
		let live = true;
		try {
			this.setStatus('Resolving the comp…');
			const comp = await this.buildComp();
			const resolution = await sim.resolveComp(comp);
			const out: RunOutput = { comp, resolution };
			this.lastRun = out;
			this.stale = false;
			if (resolution.error || !resolution.request) {
				this.setError('The resolver couldn’t resolve this comp', resolution.error?.message || 'it returned no request.');
				this.renderAssignments();
				this.renderResults();
				return;
			}
			const request = RaidSimRequest.clone(resolution.request);
			request.simOptions = SimOptions.create({ ...comp.simOptions, ...(request.simOptions ?? {}), debugFirstIteration: true });
			if (!request.simOptions.iterations) request.simOptions.iterations = comp.simOptions!.iterations;
			// The resolver keeps each player's items, but a player it made itself brings none.
			request.raid?.parties.forEach(party =>
				party.players.forEach(player => {
					if (player.equipment && !player.database) player.database = sim.db.lookupEquipmentSpec(player.equipment).toDatabase();
				}),
			);
			out.request = request;
			this.setStatus(`Simming ${plural(request.raid?.parties.flatMap(p => p.players).length ?? 0, 'player')}…`);
			const result = await sim.runRaidSimRequest(request, (progress: ProgressMetrics) => {
				if (!live || this.stopping) return;
				progressBar.update(progress);
				this.simUI.resultsManager.setSimProgress(progress);
				if (progress.totalIterations)
					this.setStatus(`Simming: ${progress.completedIterations} of ${progress.totalIterations} iterations, ${fmt(progress.dps)} DPS`);
			});
			live = false;
			out.result = result;
			// The raid changed while it ran.
			if (this.revision != revision) this.stale = true;
			if (result.error && this.stopping) {
				// A stopped run keeps nothing: neither the resolution nor a place in the sim history. The sidebar
				// goes back to the latest run (history.finishRun, below).
				this.lastRun = null;
				this.stale = false;
				this.setStatus('Run stopped. Nothing was kept.');
			} else if (result.error) {
				this.setError('The raid sim stopped', result.error.message || 'aborted');
			} else {
				out.simResult = await SimResult.makeNew(request, result);
				// With the sim history, this adds the run to it and shows it in the sidebar.
				this.simUI.resultsManager.setSimResult(TypedEvent.nextEventID(), out.simResult);
				this.simUI.setShownRaid(raid);
				// The raid DPS shows over the raid, so the status line has nothing more to say.
					this.setStatus('');
			}
			this.renderRaid();
			this.renderAssignments();
			this.renderResults();
		} catch (e) {
			console.error(e);
			this.setError('The run failed', e instanceof Error ? e.message : `${e}`);
		} finally {
			live = false;
			// A run that failed or was stopped isn't in the history, and the sidebar shows the latest run again.
			if (history?.isRunning()) history.finishRun();
			await progressBar.finish();
			progressBar.remove();
			this.simUI.simActionsContainer.classList.remove('running');
			this.running = false;
			this.setRunning(false);
		}
	}

	// A hand-built raid's member values (docs/raid_sim.md, Member value): the native server runs the
	// single-player sims its cache lacks. A second click aborts: the page stops waiting, but the server has
	// stops starting new sims when the call goes, and finishes the ones running (into its cache, so a rerun
	// is quicker).
	async memberValues() {
		if (this.valuesAbort) {
			this.valuesAbort.abort();
			return;
		}
		if (this.valuesUnavailable) {
			this.setStatus(this.valuesUnavailable, true);
			return;
		}
		if (!this.data) {
			this.setStatus('The report’s builds haven’t loaded yet.', true);
			return;
		}
		const abort = new AbortController();
		this.valuesAbort = abort;
		this.setValuesRunning(true);
		const started = Date.now();
		const revision = this.revision;
		const tick = () =>
			this.setStatus(
				`Simming the member values: ${fmt(
					(Date.now() - started) / 1000,
					0,
				)} s. The native server sims what its cache lacks, which can take minutes for a new raid.`,
			);
		tick();
		const timer = window.setInterval(tick, 1000);
		try {
			const comp = await this.buildComp();
			const out = await this.simUI.sim.compValues(comp, abort.signal);
			if (abort.signal.aborted) {
				this.setStatus(
					'Member values aborted. The native server starts no new sims and finishes the ones running; they go into its cache, so asking again is quicker.',
				);
				return;
			}
			if (out.error) {
				this.setError('No member values', out.error.message);
				this.valuesHost.replaceChildren();
				return;
			}
			this.valuesStale = false;
			this.valuesHost.replaceChildren(this.valuesTable(comp, out));
			if (this.revision != revision) {
				this.valuesStale = true;
				this.renderValuesStale();
			}
			this.setStatus(`Member values done in ${fmt((Date.now() - started) / 1000, 0)} s.`);
		} catch (e) {
			console.error(e);
			this.setError('Member values failed', e instanceof Error ? e.message : `${e}`);
		} finally {
			window.clearInterval(timer);
			this.valuesAbort = null;
			this.setValuesRunning(false);
		}
	}

	// The member values of a hand-built raid, one row per distinct member, named by the members it stands for.
	private valuesTable(comp: RaidComp, out: CompValuesResult): HTMLElement {
		const slotById = new Map(comp.slots.map(sl => [sl.id, sl]));
		const rows = out.memberValues.map(mv => {
			const found = this.data?.byId.get(mv.buildId);
			const cls = found?.build.class ?? Class.ClassUnknown;
			const label = found ? buildLabel(found.build, found.variant) : mv.buildId;
			const seats = mv.slots.map(id => slotById.get(id)).filter((sl): sl is CompSlot => !!sl);
			const seatText = seats.length ? seats.map(sl => `${sl.party + 1}.${sl.index + 1}`).join(', ') : '';
			const pkg = seats[0] ? this.packageOfParty(seats[0].party) : RaidPackage.RaidPackageNone;
			const duty = found?.variant.duty;
			const replacement = mv.bestReplacement ? this.data?.byId.get(mv.bestReplacement) : undefined;
			return [
				<span className={`rs-member-label ${cls == Class.ClassUnknown ? '' : `text-${cssClassForClass(cls)}`}`} title={mv.buildId}>
					{CLASS_ICONS[cls] ? <img className="rs-member-icon" src={CLASS_ICONS[cls]} alt={iconAlt(cls)} /> : null}
					{label}
					{seats.length > 1 ? <span className="rs-member-count">×{seats.length}</span> : null}
					<span className="rs-member-seat">
						{seats.length == 1 ? 'slot' : 'slots'} {seatText}
						{pkg != RaidPackage.RaidPackageNone ? ` · ${packageName(pkg)}` : ''}
					</span>
				</span>,
				fmt(mv.ownDps),
				fmtSigned(mv.contribution - mv.ownDps),
				fmt(mv.contribution),
				mv.bestReplacement ? fmt(mv.replacementMargin) : NA,
				mv.bestReplacement ? (replacement ? buildLabel(replacement.build, replacement.variant) : mv.bestReplacement) : NA,
				duty !== undefined && duty != Duty.DutyNone ? dutyName(duty) : NA,
				duty !== undefined && duty != Duty.DutyNone ? fmt(mv.dutyCost) : NA,
				mv.reportDifference || mv.reportDifferenceParts.length ? fmtSigned(mv.reportDifference) : NA,
				partsToggle(mv.reportDifferenceParts),
			];
		});
		return section('Member values', table(MEMBER_VALUE_COLUMNS, rows, 'rs-table-values'), {
			note: `One row per distinct member, with the slots (group.place) it stands for. The members’ single-player sum under this comp’s set: ${fmt(
				out.score,
			)} ${ci95Labelled(out.se)}.`,
		});
	}

	// The package a party's shaman gives, by the resolver's record when there is one, else by its preset.
	private packageOfParty(party: number): RaidPackage {
		for (let i = 0; i < PARTY_SIZE; i++) {
			const m = this.state.cells[party * PARTY_SIZE + i];
			if (!m) continue;
			const pkg = m.preset ? presetInfo(m.preset)?.package : this.buildOf(m)?.package;
			if (pkg) return pkg;
		}
		return RaidPackage.RaidPackageNone;
	}

	// The member values, kept but marked as another raid's after an edit.
	private renderValuesStale() {
		const values = this.valuesHost.querySelector('.rs-section');
		if (!values || values.querySelector('.rs-stale-note')) return;
		values.classList.add('rs-stale');
		values
			.querySelector('.rs-section-title')
			?.after((<p className="rs-stale-note">The raid changed after these values: ask for member values again to update them.</p>) as HTMLElement);
	}

	private setRunning(running: boolean) {
		this.runButton.replaceChildren(<i className={`fas ${running ? 'fa-stop' : 'fa-play'} me-1`} />, running ? 'Stop' : 'Run');
		const sidebar = this.simUI.simActionsContainer.querySelector<HTMLButtonElement>('.dps-action');
		if (sidebar) {
			// Stop is shorter than Run comp, and the button keeps its width so the bar beside it doesn't grow.
			sidebar.style.minWidth = running ? `${sidebar.getBoundingClientRect().width}px` : '';
			if (running) sidebar.replaceChildren(<i className="fas fa-stop me-1" />, 'Stop');
			else sidebar.textContent = 'Run comp';
		}
		this.valuesButton.disabled = running || !!this.valuesUnavailable;
	}

	private setValuesRunning(running: boolean) {
		this.valuesButton.replaceChildren(
			<i className={`fas ${running ? 'fa-stop' : 'fa-calculator'} me-1`} />,
			running ? 'Abort member values' : 'Member values',
		);
		this.valuesButton.classList.toggle('btn-outline-danger', running);
		this.valuesButton.classList.toggle('btn-outline-primary', !running);
		this.runButton.disabled = running;
	}

	private setStatus(text: string, error = false) {
		this.runStatus.textContent = text;
		this.runStatus.classList.toggle('text-danger', error);
	}

	// An error in the status line: its first line, with the rest (a Go stack trace) in a collapsible.
	private setError(prefix: string, message: string) {
		this.runStatus.replaceChildren(errorStatus(prefix, message));
		this.runStatus.classList.add('text-danger');
	}

	// ---- The run's results ----

	private renderResults() {
		const run = this.lastRun;
		if (!run) {
			this.resultsHost.replaceChildren();
			return;
		}
		const { resolution, request, result, simResult } = run;
		const parts: Node[] = [];
		const slotById = new Map(run.comp.slots.map(s => [s.id, s]));

		if (simResult && result && request) {
			const iterations = request.simOptions?.iterations || 1;

			// The members, in the request's order, back to their slots.
			const order = playersInOrder(request);
			const rows = order.map((raidIndex, k) => {
				const metrics = simResult.getPlayerWithRaidIndex(raidIndex);
				const slot = slotById.get(resolution.slotOfPlayer[k]);
				const reportId = slot?.reportId;
				const published = reportId ? this.data?.byId.get(reportId)?.variant.published : undefined;
				const duties = resolution.assignments
					.filter(a => a.holder == slot?.id && a.job.oneofKind == 'duty' && a.job.duty != Duty.DutyNone)
					.map(a => (a.job.oneofKind == 'duty' ? dutyName(a.job.duty) : ''));
				const dps = metrics?.dps.avg ?? 0;
				return {
					party: Math.floor(raidIndex / PARTY_SIZE) + 1,
					row: [
						`${Math.floor(raidIndex / PARTY_SIZE) + 1}`,
						<span className={slot ? `text-${cssClassForClass(this.slotClass(slot))}` : ''}>
							{slot?.name ?? metrics?.name ?? `player ${k + 1}`}
						</span>,
						<span>
							{fmt(dps)} <span className="rs-ci">{metrics ? `±${fmt((1.96 * metrics.dps.stdev) / Math.sqrt(iterations))}` : ''}</span>
						</span>,
						metrics ? fmt(metrics.tps.avg) : NA,
						published ? fmtSigned(dps - published) : NA,
						duties.length ? duties.join(', ') : NA,
						this.getsText(slot, resolution) || NA,
					],
				};
			});
			parts.push(
				section(
					'Members',
					table(
						[
							'Group',
							'Member',
							{ label: 'DPS ±95%', numeric: true, tooltip: 'Its DPS in the raid sim, with its 95% interval' },
							{ label: 'TPS', numeric: true, tooltip: 'Threat per second' },
							{ label: 'vs the DPS Report', numeric: true },
							'Duties',
							'Gets',
						],
						rows.map(r => r.row),
						'rs-table-run-members',
					),
					{
						note: 'Each member’s DPS and threat per second here, and its difference from its report build’s published DPS. Each duty’s cost is in Member values.',
					},
				),
			);
		}

		parts.push(
			section(
				'Buffs and debuffs',
				resolution.buffs.length
					? table(
							['Effect', 'Scope', 'How', { label: 'Receivers', numeric: true }, 'Provider', { label: 'Uptime', numeric: true }],
							resolution.buffs.map(b => [
								b.effect,
								b.scope,
								b.given ? 'given' : 'cast',
								`${b.receivers}`,
								b.provider,
								b.given ? (
									<span title="A given effect is in each receiver’s own request from the pull, so it’s up all fight">up all fight</span>
								) : simResult ? (
									uptimeText(simResult, b.effect, b.scope)
								) : (
									''
								),
							]),
					  )
					: empty('The resolver recorded no buffs.'),
				{
					note: 'From the resolver’s record. A given effect is up all fight; a cast one’s uptime is its aura’s in the raid sim (the boss’s for a debuff; for a buff, the average over the members it was up on, leaving out a member whose aura never came up, like an Elemental shaman’s Windfury Totem), and shows — where the sim has no aura for it.',
				},
			),
			section('What the raid goes without', resolution.misses.length ? list(resolution.misses, 'rs-misses') : empty('Nothing.')),
			<details className="rs-resolved">
				<summary>Every assignment as resolved</summary>
				{resolution.assignments.length ? list(resolution.assignments.map(a => this.assignmentText(a, slotById))) : empty('None.')}
			</details>,
		);
		if (this.stale) {
			parts.unshift(
				<p className="rs-stale-note" attributes={{ role: 'status' }}>
					The raid changed after this run, so these results are of the raid as it was. Run it again to update them.
				</p>,
			);
		}
		this.resultsHost.replaceChildren(<h3 className="rs-results-title">Last run</h3>, ...parts);
		this.resultsHost.classList.toggle('rs-stale', this.stale);
	}

	private slotClass(slot: CompSlot): Class {
		if (slot.member.oneofKind == 'placeholder') return presetInfo(slot.member.placeholder)?.class ?? Class.ClassUnknown;
		if (slot.member.oneofKind == 'build') return slot.member.build.base?.class ?? Class.ClassUnknown;
		return Class.ClassUnknown;
	}

	// What a member gets from the assignments: its class's blessings, Power Infusion and Innervate.
	private getsText(slot: CompSlot | undefined, resolution: RaidCompResolution): string {
		if (!slot) return '';
		const cls = this.slotClass(slot);
		const gets: string[] = [];
		resolution.assignments.forEach(a => {
			if (a.job.oneofKind == 'blessing' && a.job.blessing.class == cls && a.job.blessing.blessing != Blessing.BlessingNone)
				gets.push(blessingName(a.job.blessing.blessing));
			if (a.job.oneofKind == 'cooldown' && a.job.cooldown.target == slot.id) gets.push(a.job.cooldown.powerInfusion ? 'Power Infusion' : 'Innervate');
		});
		return gets.join(', ');
	}

	private assignmentText(a: Assignment, slotById: Map<number, CompSlot>): Node {
		const holder = slotById.get(a.holder)?.name ?? `slot ${a.holder}`;
		let job = '';
		switch (a.job.oneofKind) {
			case 'blessing':
				job = `${a.job.blessing.blessing == Blessing.BlessingNone ? 'no blessing' : blessingName(a.job.blessing.blessing)} to ${
					classNames[a.job.blessing.class]
				}s`;
				break;
			case 'judgement':
				job = a.job.judgement == PaladinJudgement.PaladinJudgementNone ? 'judges nothing' : `judges ${judgementName(a.job.judgement)}`;
				break;
			case 'totem': {
				const element = assignedElement(a.job.totem);
				job =
					a.job.totem.totem != RaidTotem.RaidTotemNone
						? `drops ${totemName(a.job.totem.totem)}`
						: element
						? `drops no ${element} totem`
						: 'drops no totem';
				break;
			}
			case 'duty':
				job = `holds ${dutyName(a.job.duty)}`;
				break;
			case 'cooldown':
				job = `${a.job.cooldown.powerInfusion ? 'Power Infusion' : 'Innervate'} on ${
					slotById.get(a.job.cooldown.target)?.name ?? `slot ${a.job.cooldown.target}`
				}`;
				break;
		}
		return (
			<span>
				{holder}: {job} {a.setByHand ? badge('by hand', 'secondary') : badge('auto', 'info')}
			</span>
		);
	}
}

// A palette entry's key, which keeps its focus across a redraw.
const chipKey = (member: Omit<Member, 'id'>) => (member.preset ? `preset-${member.preset}` : `build-${member.buildId}`);

// The raid index of each player of a request, in order (parties and players in order), the order the
// resolution's slot_of_player follows.
const playersInOrder = (request: RaidSimRequest): number[] => {
	const out: number[] = [];
	request.raid?.parties.forEach((party, p) =>
		party.players.forEach((player, i) => {
			if (player.class != Class.ClassUnknown) out.push(p * PARTY_SIZE + i);
		}),
	);
	return out;
};

// The sim's aura names for each effect the resolver records (sim/compsearch's giftNames), without their
// rank ("Sunder Armor (Rank 5)" is Sunder Armor), so an uptime is that effect's aura and not some other
// aura with a like name. The boss's for a debuff, the members' for a buff.
const EFFECT_AURAS: Record<string, string[]> = {
	'Battle Shout': ['Battle Shout'],
	'Sunder upkeep': ['Sunder Armor'],
	'Minor armor slot': ['Faerie Fire', 'Faerie Fire (Feral)', 'Curse of Recklessness', 'Dust Cloud'],
	'Curse of the Elements': ['Curse of Elements', 'Curse of the Elements'],
	'Judgement of the Crusader': ['Judgement of the Crusader'],
	'Judgement of Wisdom': ['Judgement of Wisdom'],
	'Strength of Earth': ['Strength of Earth Totem', 'Strength of Earth'],
	'Grace of Air': ['Grace of Air Totem', 'Grace of Air'],
	// The shaman's totem aura; the receivers' buff has no metrics of its own.
	'Windfury Totem': ['Windfury Totem Passive', 'Windfury Totem'],
	'Flametongue Totem': ['Flametongue Totem'],
	'Moonkin Aura': ['Moonkin Aura'],
	'Leader of the Pack': ['Leader of the Pack'],
	'Trueshot Aura': ['Trueshot Aura'],
	'Arcane Brilliance': ['Arcane Brilliance', 'Arcane Intellect'],
	'Gift of the Wild': ['Gift of the Wild', 'Mark of the Wild'],
	'Prayer of Spirit': ['Prayer of Spirit', 'Divine Spirit'],
	Fortitude: ['Prayer of Fortitude', 'Power Word: Fortitude'],
	'Furious Howl': ['Furious Howl'],
	'Mana Spring': ['Mana Spring Totem', 'Mana Spring'],
	'Mana Spring and Mana Tide': ['Mana Spring Totem', 'Mana Spring'],
	'Mana Spring ×1.25 and Mana Tide': ['Mana Spring Totem', 'Mana Spring'],
};

// The uptime of a cast effect in a raid sim: the average over the units it was up on, the boss for a
// debuff and the players for a buff. A unit that has the aura but never had it up isn't a giver or a
// receiver of it (an Elemental shaman registers Windfury Totem and never drops it, next to an
// Enhancement shaman who does), so it's left out, and the text says how many were. An effect the page
// knows no aura for, or whose aura no unit carries, shows none.
// The Windfury Totem's 1 sec attack power buff on each proc (sim/core's WindfuryBuffSpellId): its spell is
// named Windfury Totem too, but its uptime is the procs', not the totem's.
const WINDFURY_ATTACK_SPELLS = [8516, 10608, 10610];

const auraBaseName = (name: string) => name.replace(/\s*\(Rank \d+\)$/, '');

const uptimeText = (simResult: SimResult, effect: string, scope: string): Node | string => {
	const names = EFFECT_AURAS[effect];
	if (!names) return NA;
	const units = /boss|target/i.test(scope) ? simResult.getTargets() : simResult.getPlayers();
	const uptimes: number[] = [];
	units.forEach(u => {
		const auras = u.auras.filter(a => names.includes(auraBaseName(a.name)) && !WINDFURY_ATTACK_SPELLS.includes(a.actionId.spellId));
		if (auras.length) uptimes.push(Math.max(...auras.map(a => a.uptimePercent)));
	});
	if (!uptimes.length) return NA;
	const up = uptimes.filter(u => u > 0);
	if (!up.length) return '0%';
	const average = `${fmt(up.reduce((a, b) => a + b, 0) / up.length, 1)}%`;
	if (up.length == uptimes.length) return average;
	return (
		<span title={`The average over the ${plural(up.length, 'unit')} it was up on; ${uptimes.length - up.length} more have its aura and never had it up`}>
			{average}{' '}
			<span className="rs-muted">
				on {up.length} of {uptimes.length}
			</span>
		</span>
	);
};

// A raid as sameRaid compares it: the members by slot, and the assignments by their holder's slot (and a
// Power Infusion's or Innervate's target by its slot), so the member ids don't count.
function canonicalRaid(state: RaidState) {
	const cellOf = new Map<number, number>();
	state.cells.forEach((m, cell) => m && cellOf.set(m.id, cell));
	const byCell = <T,>(record: Record<number, T>, value: (v: T) => unknown = v => v) =>
		Object.fromEntries(
			Object.entries(record)
				.filter(([id]) => cellOf.has(Number(id)))
				.map(([id, v]) => [cellOf.get(Number(id))!, value(v)]),
		);
	const target = (id: number) => cellOf.get(id) ?? -1;
	return {
		size: state.size,
		faction: state.faction,
		campBuffs: state.campBuffs,
		boss: state.boss,
		encounter: state.encounter ?? null,
		cells: state.cells.map(m => (m ? { buildId: m.buildId ?? null, preset: m.preset ?? null, presetInnervate: !!m.presetInnervate } : null)),
		hand: {
			blessings: byCell(state.hand.blessings),
			judgements: byCell(state.hand.judgements),
			totems: byCell(state.hand.totems),
			duties: byCell(state.hand.duties),
			powerInfusion: byCell(state.hand.powerInfusion, target),
			innervate: byCell(state.hand.innervate, target),
		},
	};
}

// JSON with every object's keys sorted, so two equal values write the same text.
function stableJson(value: unknown): string {
	if (Array.isArray(value)) return `[${value.map(stableJson).join(',')}]`;
	if (value && typeof value == 'object')
		return `{${Object.keys(value)
			.sort()
			.map(key => `${JSON.stringify(key)}:${stableJson((value as Record<string, unknown>)[key])}`)
			.join(',')}}`;
	return JSON.stringify(value) ?? 'null';
}

// A JSON value's leaves by path, like 'targets.0.stats.12' → '6000000'.
function flattenJson(value: unknown, path = '', out = new Map<string, string>()): Map<string, string> {
	if (Array.isArray(value)) value.forEach((v, i) => flattenJson(v, path ? `${path}.${i}` : `${i}`, out));
	else if (value && typeof value == 'object') Object.entries(value).forEach(([k, v]) => flattenJson(v, path ? `${path}.${k}` : k, out));
	else out.set(path, `${value}`);
	return out;
}

// A path of flattenJson in words, like 'targets.0.mobType' → 'Target 1 mob type'.
function humanize(path: string): string {
	const words = path
		.split('.')
		.map(part => (/^\d+$/.test(part) ? `${Number(part) + 1}` : part.replace(/([a-z])([A-Z0-9])/g, '$1 $2').toLowerCase()))
		.join(' ')
		.replace(/^targets /, 'target ');
	return words.charAt(0).toUpperCase() + words.slice(1);
}
