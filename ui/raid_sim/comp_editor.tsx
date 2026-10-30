// The comp editor (docs/raid_sim.md, The page, M2b): a raid of 10, 20 or 40 in parties of 5, each slot a
// DPS Report build or a tank or healer placeholder, with its assignments, the encounter, the faction and
// the camp buffs. Run builds a RaidComp, has the resolver turn it into the raid's request (resolveComp),
// runs that request in the sim's workers and shows the raid's results.
import { EncounterPicker } from '../core/components/encounter_picker';
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
	packageName,
	presetInfo,
	presetsFor,
	RAID_SIZES,
	RaidSimData,
	ReportBuild,
	TOTEM_ELEMENTS,
	TotemElement,
	totemElement,
	totemName,
} from './data';
import type { RaidSimUI } from './raid_sim_ui';
import { DETAILS_TAB } from './raid_sim_ui';
import { check, field, select } from './search_form';
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
	private readonly bossHost = (<div className="rs-boss" />) as HTMLElement;
	private readonly runStatus = (<div className="rs-run-status" />) as HTMLElement;
	private readonly resultsHost = (<div className="rs-run-results" />) as HTMLElement;
	private readonly runButton: HTMLButtonElement;
	private readonly valuesButton: HTMLButtonElement;
	private readonly valuesHost = (<div className="rs-member-values" />) as HTMLElement;
	private drag: DragSource | null = null;
	// The palette entry picked with a click, placed by clicking a slot (touch screens have no drag and drop).
	private picked: Omit<Member, 'id'> | null = null;
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
				title="Each member’s contribution, replacement margin and duty cost, from the native server’s single-player sims">
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
				<p className="rs-intro">
					Build a raid from the DPS Report’s builds and the tank and healer placeholders, drag members between groups, set assignments by hand or
					leave them automatic, and run it. Each member’s gear, talents, rotation and consumes come whole from its report build.
				</p>
				{this.controlsHost}
				<div className="rs-editor-layout">
					<div className="rs-editor-main">
						{section(
							'The raid',
							[
								<p className="visually-hidden" id={this.slotHelpId}>
									Enter or Space places the member picked under Add members, or picks this slot’s member up to move it to another slot. Escape
									lets go, Delete removes the member, and the arrow keys move between slots.
								</p>,
								this.raidHost,
							],
							{
								note: 'Drag a member onto another slot to move or swap it, or pick one below and click a slot. With the keyboard, Enter on a member picks it up and Enter on another slot puts it there.',
							},
						)}
						{section('Add members', this.paletteHost)}
					</div>
					<div className="rs-editor-side">
						{section('Assignments', this.assignmentsHost, {
							note: 'Auto leaves a job to the resolver: biggest gain first, to a placeholder before a DPS member.',
						})}
						{section('The encounter', this.encounterHost)}
					</div>
				</div>
				<div className="rs-run-bar">
					{this.runButton}
					{this.valuesButton}
					{this.valuesNote}
					{this.runStatus}
				</div>
				{this.valuesHost}
				{this.resultsHost}
			</div>
		) as HTMLElement;
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
			if (this.lastRun?.simResult) this.setRunStale(true);
			this.renderResults();
			// The Auto labels and slot DPS were the run's.
			this.renderRaid();
			this.renderAssignments();
		}
		if (hadValues) this.renderValuesStale();
	}

	// Forget the last run and the member values (New raid, Clear, a comp loaded from a search).
	private forgetResults() {
		this.lastRun = null;
		this.stale = false;
		this.valuesStale = false;
		this.valuesAbort?.abort();
		this.valuesHost.replaceChildren();
		this.renderResults();
		this.clearSidebar();
	}

	// The sidebar's raid DPS (or a run's progress) and Raid results, which were the last run's. A run's
	// progress isn't run data, so it goes whether there is any or not.
	private clearSidebar() {
		if (this.simUI.resultsManager.getRunData()) this.simUI.resultsManager.clearSimResult(TypedEvent.nextEventID());
		this.simUI.resultsViewer.contentElem.replaceChildren();
		this.simUI.resultsViewer.hideAll();
		this.setRunStale(false);
	}

	// The sidebar's raid DPS and Raid results are of the raid as it was.
	private setRunStale(stale: boolean) {
		this.sidebarStaleNote.classList.toggle('d-none', !stale);
		this.detailsStaleNote.classList.toggle('d-none', !stale);
		this.simUI.resultsViewer.contentElem.classList.toggle('rs-stale', stale);
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
		this.renderAssignmentWarnings();
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
			<button className="btn btn-outline-secondary" type="button" title="Start over with this size’s default tanks and healers">
				New raid
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
			<button className="btn btn-outline-secondary" type="button">
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
			field(
				'Raid size',
				select(
					[10, 20, 40].map(n => ({ value: `${n}`, label: `${n}` })),
					`${s.size}`,
					v => this.resize(parseInt(v)),
				),
			),
			field(
				'Faction',
				select(
					[Faction.Horde, Faction.Alliance].map(f => ({ value: `${f}`, label: factionName(f) })),
					`${s.faction}`,
					v => {
						this.state.faction = parseInt(v);
						this.assignmentChanged();
					},
				),
				'each build takes its best race in it',
			),
			field(
				'Camp buffs',
				check('The raid relies on camp buffs', s.campBuffs, v => {
					this.state.campBuffs = v;
					this.assignmentChanged();
				}),
			),
			<div className="rs-editor-buttons">
				{newRaid}
				{clear}
				{this.undoButton}
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
		if (dropped.length) {
			const who = dropped.length == 1 ? 'this member' : `these ${dropped.length} members`;
			this.setStatus(`Down to ${size} players: ${who} past group ${size / PARTY_SIZE} left the raid: ${dropped.join(', ')}. Undo brings them back.`);
			this.setUndo(before, 'the resize');
		}
	}

	// ---- The raid: parties of 5 with drag and drop ----

	private renderRaid() {
		const parties = this.state.size / PARTY_SIZE;
		const run = this.freshRun;
		// The slot with the keyboard's focus keeps it when the raid is drawn again.
		const active = document.activeElement;
		const activeCell = active instanceof HTMLElement && this.raidHost.contains(active) && active.dataset.cell ? parseInt(active.dataset.cell) : null;
		const keep = this.focusCell ?? activeCell;
		const dpsOf = (id: number): number | undefined => {
			if (!run?.simResult || !run.request) return undefined;
			const k = run.resolution.slotOfPlayer.indexOf(id);
			if (k < 0) return undefined;
			return playersInOrder(run.request)[k] !== undefined ? run.simResult.getPlayerWithRaidIndex(playersInOrder(run.request)[k])?.dps.avg : undefined;
		};
		this.raidHost.replaceChildren(
			<div className={`rs-parties rs-parties-${parties}`}>
				{[...Array(parties).keys()].map(p => {
					const cells = [...Array(PARTY_SIZE).keys()].map(i => p * PARTY_SIZE + i);
					const partyDps = cells.reduce((sum, c) => {
						const m = this.state.cells[c];
						return sum + (m ? dpsOf(m.id) ?? 0 : 0);
					}, 0);
					return (
						<div className="rs-party party-picker-root active">
							<div className="rs-party-header">
								<span className="rs-party-name">Group {p + 1}</span>
								{partyDps ? <span className="rs-party-dps">{fmt(partyDps)} DPS</span> : null}
							</div>
							<div className="players-container">{cells.map(c => this.slotElem(c, dpsOf))}</div>
						</div>
					);
				})}
			</div>,
		);
		if (keep !== null) this.raidHost.querySelector<HTMLElement>(`[data-cell="${keep}"]`)?.focus();
		this.focusCell = null;
	}

	// What a slot's keys do: Enter or Space places the member picked below, or picks this slot's member up
	// and puts it down on another slot (a move, or a swap); Escape lets go; Delete removes; the arrow keys
	// go up and down a group and across the groups.
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
				if (this.moving === null && !this.picked) return;
				e.preventDefault();
				this.moving = null;
				this.picked = null;
				this.focusCell = cell;
				this.setStatus('');
				this.renderPalette();
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

	// A click, or Enter or Space, on a slot.
	private activateSlot(cell: number) {
		if (this.picked) {
			this.focusCell = cell;
			this.dropOn(cell, { kind: 'new', member: this.picked });
			return;
		}
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
		this.setStatus(`Moving ${this.memberShortName(member, cell)}: Enter or a click on another slot puts it there (a swap if it’s taken); Escape lets go.`);
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

	private slotElem(cell: number, dpsOf: (id: number) => number | undefined) {
		const member = this.state.cells[cell];
		let elem: HTMLElement;
		const where = `Group ${Math.floor(cell / PARTY_SIZE) + 1}, slot ${(cell % PARTY_SIZE) + 1}`;
		if (!member) {
			elem = (
				<div className="player player-picker-root rs-slot rs-slot-empty" title="An empty slot: drop a member here, or pick one below and click">
					<span className="rs-slot-empty-text">empty</span>
				</div>
			) as HTMLElement;
			elem.setAttribute('aria-label', `${where}: empty`);
		} else {
			const cls = this.classOf(member);
			const info = member.preset ? presetInfo(member.preset) : undefined;
			const icon = info?.icon ?? CLASS_ICONS[cls] ?? '';
			const dps = dpsOf(member.id);
			// A build this page's data doesn't have (a result from a newer report): Run refuses it.
			const unknownBuild = !member.preset && !!this.data && !this.buildOf(member);
			// The slot itself takes Delete, so its remove button stays out of the tab order.
			const remove = (
				<a href="javascript:void(0)" className="player-delete link-danger" attributes={{ role: 'button', 'aria-hidden': 'true' }} title="Remove">
					<i className="fa fa-times fa-lg" />
				</a>
			) as HTMLElement;
			remove.tabIndex = -1;
			remove.addEventListener('click', e => {
				e.stopPropagation();
				this.removeMember(cell);
			});
			elem = (
				<div className={`player player-picker-root rs-slot bg-${cssClassForClass(cls)}-dampened`} draggable={true}>
					<div className="player-label">
						<img className="player-icon" src={icon} draggable={false} alt={iconAlt(cls)} />
						<div className="player-details">
							{/* The icon and color say the class, so a slot names only the build, as the palette does. */}
							<span className={`rs-slot-name text-${cssClassForClass(cls)}`} title={this.memberName(member)}>
								{member.preset ? this.memberName(member) : this.memberName(member).replace(/^[^:]*: /, '')}
							</span>
							<span className="rs-slot-sub">
								{unknownBuild ? 'not one of this page’s builds' : info ? (info.role == RaidRole.RaidRoleTank ? 'Tank' : 'Healer') : 'DPS'}
								{dps !== undefined ? ` · ${fmt(dps)} DPS` : ''}
							</span>
						</div>
					</div>
					<div className="player-options">{remove}</div>
				</div>
			) as HTMLElement;
			elem.setAttribute(
				'aria-label',
				`${where}: ${this.memberName(member)}, ${info ? (info.role == RaidRole.RaidRoleTank ? 'tank' : 'healer') : 'DPS'}${
					dps !== undefined ? `, ${fmt(dps)} DPS` : ''
				}${this.moving == cell ? ', picked up to move' : ''}`,
			);
			if (this.moving == cell) elem.classList.add('rs-slot-moving');
			if (unknownBuild) elem.classList.add('rs-slot-unknown');
			elem.addEventListener('dragstart', e => {
				this.drag = { kind: 'cell', index: cell };
				e.dataTransfer!.effectAllowed = 'move';
				e.dataTransfer!.setData('text/plain', `rs-cell:${cell}`);
				elem.classList.add('dragfrom');
			});
			elem.addEventListener('dragend', () => {
				elem.classList.remove('dragfrom');
				this.drag = null;
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
		elem.addEventListener('click', () => {
			if (!this.picked && this.moving === null) return;
			this.activateSlot(cell);
		});
		elem.tabIndex = 0;
		elem.dataset.cell = `${cell}`;
		elem.setAttribute('role', 'button');
		elem.setAttribute('aria-describedby', this.slotHelpId);
		elem.addEventListener('keydown', e => this.slotKey(e, cell));
		if (this.picked || (this.moving !== null && this.moving != cell)) elem.classList.add('rs-slot-target');
		return elem;
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
		}
		this.changed();
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

	// ---- The palette: presets and the report's builds ----

	private renderPalette() {
		const chip = (member: Omit<Member, 'id'>, label: string, icon: string, cls: Class, title: string) => {
			const elem = (
				<button
					className={`rs-palette-chip ${this.isPicked(member) ? 'active' : ''} text-${cssClassForClass(cls)}`}
					type="button"
					draggable={true}
					title={title}>
					{icon ? <img className="rs-palette-icon" src={icon} draggable={false} alt={iconAlt(cls)} /> : null}
					<span>{label}</span>
				</button>
			) as HTMLButtonElement;
			elem.dataset.key = chipKey(member);
			elem.setAttribute('aria-pressed', this.isPicked(member) ? 'true' : 'false');
			elem.addEventListener('dragstart', e => {
				this.drag = { kind: 'new', member };
				e.dataTransfer!.effectAllowed = 'copy';
				e.dataTransfer!.setData('text/plain', 'rs-new');
			});
			elem.addEventListener('dragend', () => (this.drag = null));
			elem.addEventListener('click', () => {
				this.picked = this.isPicked(member) ? null : member;
				this.renderPalette();
				this.renderRaid();
			});
			return elem;
		};
		const presetChips = (role: RaidRole) =>
			presetsFor(role).map(p => chip({ preset: p.id }, p.name, p.icon, p.class, `${p.name} (${p.talents}): ${p.gives}`));

		// The chip with the keyboard's focus keeps it when the palette is drawn again.
		const active = document.activeElement;
		const activeKey = active instanceof HTMLElement && this.paletteHost.contains(active) ? active.dataset.key : undefined;
		const byClass = new Map<string, ReportBuild[]>();
		this.data?.builds.forEach(b => {
			if (!byClass.has(b.className)) byClass.set(b.className, []);
			byClass.get(b.className)!.push(b);
		});
		this.paletteHost.replaceChildren(
			<div className="rs-palette-group">
				<span className="rs-palette-label">Tanks</span>
				<div className="rs-palette-chips">{presetChips(RaidRole.RaidRoleTank)}</div>
			</div>,
			<div className="rs-palette-group">
				<span className="rs-palette-label">Healers</span>
				<div className="rs-palette-chips">{presetChips(RaidRole.RaidRoleHealer)}</div>
			</div>,
			...(this.data
				? [...byClass.entries()].map(([className, builds]) => (
						<div className="rs-palette-group">
							<span className="rs-palette-label">{className}</span>
							<div className="rs-palette-chips">
								{builds.map(b =>
									chip(
										{ buildId: b.id },
										buildLabel(b).replace(/^[^:]*: /, ''),
										'',
										b.class,
										`${buildLabel(b)}, ${fmt(b.published)} DPS in the report. Gives: ${b.gives}.${b.summary ? ' ' + b.summary : ''}`,
									),
								)}
							</div>
						</div>
				  ))
				: [<p className="rs-muted">Loading the report’s builds…</p>]),
		);
		if (activeKey) this.paletteHost.querySelector<HTMLElement>(`[data-key="${activeKey}"]`)?.focus();
	}

	private isPicked(member: Omit<Member, 'id'>) {
		return !!this.picked && this.picked.buildId == member.buildId && this.picked.preset == member.preset;
	}

	// ---- Assignments ----

	private resolvedText(job: (a: Assignment) => boolean): string | null {
		const found = this.freshRun?.resolution.assignments.find(job);
		if (!found) return null;
		switch (found.job.oneofKind) {
			case 'blessing':
				return blessingName(found.job.blessing.blessing);
			case 'judgement':
				return judgementName(found.job.judgement);
			case 'totem':
				return found.job.totem.totem == RaidTotem.RaidTotemNone ? 'none' : totemName(found.job.totem.totem);
			case 'duty':
				return dutyName(found.job.duty);
			case 'cooldown': {
				const target = this.memberById(found.job.cooldown.target);
				return target ? this.memberShortName(target.member, target.cell) : `slot ${found.job.cooldown.target}`;
			}
		}
		return null;
	}

	private autoLabel(job: (a: Assignment) => boolean) {
		const resolved = this.resolvedText(job);
		return resolved ? `Auto (${resolved})` : 'Auto';
	}

	private renderAssignments() {
		const members = this.members();
		const h = this.state.hand;
		const rows: Node[] = [];
		const presentClasses = BLESSING_CLASSES.filter(c => members.some(m => !m.member.preset && this.classOf(m.member) == c));

		// Each paladin: a Greater Blessing for each class, and a placeholder's judgement.
		const paladins = members.filter(m => this.classOf(m.member) == Class.ClassPaladin);
		paladins.forEach(({ member, cell }) => {
			const controls: Node[] = [];
			if (member.preset) {
				controls.push(
					field(
						'Judgement',
						select(
							[
								{ value: '', label: this.autoLabel(a => a.holder == member.id && a.job.oneofKind == 'judgement') },
								...[
									PaladinJudgement.PaladinJudgementCrusader,
									PaladinJudgement.PaladinJudgementWisdom,
									PaladinJudgement.PaladinJudgementNone,
								].map(j => ({
									value: `${j}`,
									label: judgementName(j),
								})),
							],
							h.judgements[member.id] !== undefined ? `${h.judgements[member.id]}` : '',
							v => {
								if (v === '') delete h.judgements[member.id];
								else h.judgements[member.id] = parseInt(v);
								this.assignmentChanged();
							},
						),
					),
				);
			}
			presentClasses.forEach(c => {
				const current = h.blessings[member.id]?.[c];
				controls.push(
					field(
						classNames[c],
						select(
							[
								{
									value: '',
									label: this.autoLabel(a => a.holder == member.id && a.job.oneofKind == 'blessing' && a.job.blessing.class == c),
								},
								...[Blessing.BlessingKings, Blessing.BlessingMight, Blessing.BlessingWisdom, Blessing.BlessingNone].map(b => ({
									value: `${b}`,
									label: blessingName(b),
								})),
							],
							current !== undefined ? `${current}` : '',
							v => {
								const mine = (h.blessings[member.id] ??= {});
								if (v === '') delete mine[c];
								else mine[c] = parseInt(v);
								this.assignmentChanged();
							},
						),
					),
				);
			});
			rows.push(this.assignmentRow(member, cell, 'Greater Blessings for each class', controls));
		});

		// Each placeholder shaman: a totem for each element. A DPS shaman's rotation drops fixed totems. Each
		// element set by hand is sent with its element (none drops nothing there); the elements left on Auto
		// stay the resolver's.
		members
			.filter(m => m.member.preset && presetInfo(m.member.preset)?.shaman)
			.forEach(({ member, cell }) => {
				const controls = TOTEM_ELEMENTS.map(({ element, label, totems }) =>
					field(
						label,
						select(
							[
								{
									value: '',
									label: this.autoLabel(a => a.holder == member.id && a.job.oneofKind == 'totem' && assignedElement(a.job.totem) == element),
								},
								...totems.map(t => ({ value: `${t}`, label: totemName(t) })),
								{ value: `${RaidTotem.RaidTotemNone}`, label: 'none' },
							],
							h.totems[member.id]?.[element] !== undefined ? `${h.totems[member.id]![element]}` : '',
							v => {
								const mine = (h.totems[member.id] ??= {});
								if (v === '') delete mine[element];
								else mine[element] = parseInt(v);
								this.assignmentChanged();
							},
						),
					),
				);
				rows.push(this.assignmentRow(member, cell, 'Totems', controls));
			});

		// Duties: a DPS member with duty variants. A placeholder gives its own (a druid's Faerie Fire, the
		// warrior tank's Sunder Armor) without one.
		members.forEach(({ member, cell }) => {
			if (member.preset) return;
			const duties = (this.buildOf(member)?.duties ?? []).map(d => d.duty!).filter(d => d !== undefined);
			if (!duties.length) return;
			const current = h.duties[member.id];
			rows.push(
				this.assignmentRow(member, cell, 'Duty', [
					field(
						'Holds',
						select(
							[
								{ value: '', label: this.autoLabel(a => a.holder == member.id && a.job.oneofKind == 'duty') },
								...duties.map(d => ({ value: `${d}`, label: dutyName(d) })),
								{ value: `${Duty.DutyNone}`, label: 'None (holds no duty)' },
							],
							current !== undefined ? `${current}` : '',
							v => {
								if (v === '') delete h.duties[member.id];
								else h.duties[member.id] = parseInt(v);
								this.assignmentChanged();
							},
						),
					),
				]),
			);
		});

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
			// in it, marked, so the select shows what Run sends.
			const current = dpsMembers.find(t => t.member.id == map[holder.member.id]);
			if (current && !targets.includes(current)) targets.push(current);
			const nth = (pi ? piHolders : innervateHolders).indexOf(holder.member.id);
			const resolved = this.freshRun?.resolution.assignments.filter(a => a.job.oneofKind == 'cooldown' && a.job.cooldown.powerInfusion == pi)[nth];
			return field(
				pi ? 'Power Infusion' : 'Innervate',
				select(
					[
						{
							value: '',
							label: this.autoLabel(a => a === resolved),
						},
						...targets.map(t => ({
							value: `${t.member.id}`,
							label: `${this.memberShortName(t.member, t.cell)}${
								t != current || (pi ? casters : manaUsers).includes(t) ? '' : pi ? ' (no magic damage to boost)' : ' (no mana)'
							}`,
						})),
					],
					map[holder.member.id] !== undefined ? `${map[holder.member.id]}` : '',
					v => {
						if (v === '') delete map[holder.member.id];
						else map[holder.member.id] = parseInt(v);
						this.assignmentChanged();
					},
				),
			);
		};
		members
			.filter(m => m.member.preset && presetInfo(m.member.preset)?.powerInfusion)
			.forEach(m => rows.push(this.assignmentRow(m.member, m.cell, 'Cooldown', [cooldown(m, true)])));
		members
			.filter(m => m.member.preset && (presetInfo(m.member.preset)?.innervate || presetInfo(m.member.preset)?.innervateOptional))
			.forEach(m => {
				const info = presetInfo(m.member.preset!)!;
				const controls: Node[] = [];
				if (info.innervateOptional) {
					controls.push(
						field(
							'Innervate',
							check('Casts Innervate (leaves Bear Form)', !!m.member.presetInnervate, v => {
								m.member.presetInnervate = v;
								// Without its Innervate, a target set by hand goes too: the resolver refuses a holder
								// with nothing to give.
								if (!v) delete h.innervate[m.member.id];
								this.changed();
							}),
						),
					);
				}
				if (info.innervate || m.member.presetInnervate) controls.push(cooldown(m, false));
				rows.push(this.assignmentRow(m.member, m.cell, 'Cooldown', controls));
			});

		const dpsShamans = dpsMembers.filter(m => this.classOf(m.member) == Class.ClassShaman);
		this.renderAssignmentWarnings();
		this.assignmentsHost.replaceChildren(
			this.warningsHost,
			...(rows.length ? rows : [empty('Nobody in this raid holds an assignment yet. Paladins, shamans, priests, druids and the duty builds do.')]),
			...(dpsShamans.length
				? [<p className="rs-muted">A DPS shaman’s rotation drops its own fixed totems, so it has no totem choice yet (docs/raid_sim.md, M1b).</p>]
				: []),
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
		const bossSelect = select(
			[
				{ value: 'Generic', label: 'Generic: the report’s level 63 boss, in the natural flask zone' },
				{ value: 'Onyxia', label: 'Onyxia: a Dragonkin, outside the flask zone' },
				{ value: 'Custom', label: 'Custom: as set below' },
			],
			this.state.boss,
			v => {
				this.state.boss = v as Boss;
				this.applyBoss(this.state.boss);
				this.assignmentChanged();
			},
		);
		this.bossHost.replaceChildren(field('Boss', bossSelect));
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

		this.state = state;
		const e = request?.encounter;
		if (e) this.applySearchEncounter(e);
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

	// A search's fight, as tools/compsearch runs it (sim/compsearch, compOf): the report's boss with the
	// search's length, its health for a health fight, its creature type and its natural flask zone.
	private applySearchEncounter(e: CompEncounter) {
		if (!this.data) return;
		const encounter = EncounterProto.clone(this.data.encounter);
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
		this.state.boss = (e.name || 'Generic') == 'Generic' && generic ? 'Generic' : e.name == 'Onyxia' && onyxia ? 'Onyxia' : 'Custom';
		this.setPresetEncounter(encounter);
		this.initBossSelect();
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
				this.simUI.resultsManager.setSimProgress(progress);
				if (progress.totalIterations)
					this.setStatus(`Simming: ${progress.completedIterations} of ${progress.totalIterations} iterations, ${fmt(progress.dps)} DPS`);
			});
			live = false;
			out.result = result;
			// The raid changed while it ran.
			if (this.revision != revision) this.stale = true;
			if (result.error && this.stopping) {
				// A stopped run keeps nothing: neither its progress in the sidebar nor the resolution.
				this.lastRun = null;
				this.stale = false;
				this.clearSidebar();
				this.setStatus('Run stopped. Nothing was kept.');
			} else if (result.error) {
				this.clearSidebar();
				this.setError('The raid sim stopped', result.error.message || 'aborted');
			} else {
				out.simResult = await SimResult.makeNew(request, result);
				this.simUI.resultsManager.setSimResult(TypedEvent.nextEventID(), out.simResult);
				this.simUI.resultsViewer.contentElem
					.querySelector('.results-sim')
					?.append(
						(
							<p className="rs-sidebar-note">
								± here is one standard deviation of the iterations; the Comp tab’s ±(95%) is the average’s 95% interval.
							</p>
						) as HTMLElement,
					);
				this.setRunStale(this.stale);
				this.setStatus(`Done: ${fmt(result.raidMetrics?.dps?.avg)} raid DPS.`);
			}
			this.renderRaid();
			this.renderAssignments();
			this.renderResults();
		} catch (e) {
			console.error(e);
			this.clearSidebar();
			this.setError('The run failed', e instanceof Error ? e.message : `${e}`);
		} finally {
			live = false;
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
		if (sidebar) sidebar.textContent = running ? 'Stop' : 'Run comp';
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
			const raidDps = simResult.raidMetrics.dps;
			parts.push(
				<div className="rs-run-topline">
					<span className="rs-run-dps">{fmt(raidDps.avg)}</span>
					<span className="rs-run-dps-label">raid DPS</span>
					<span className="rs-ci" title="The 95% interval of the average raid DPS (the sidebar’s ± is one standard deviation of the iterations)">
						±{fmt((1.96 * raidDps.stdev) / Math.sqrt(iterations))} (95%)
					</span>
					<span className="rs-muted">
						{iterations} iterations, {fmt(simResult.result.avgIterationDuration, 0)} s average fight
					</span>
					<button className="btn btn-link rs-details-link" type="button" onclick={() => this.simUI.activate(DETAILS_TAB)}>
						Every metric in Raid results <i className="fas fa-arrow-right" />
					</button>
				</div>,
			);

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
			const parties = simResult.raidMetrics.parties
				.map((p, i) => ({ i, dps: p.dps.avg, n: order.filter(r => Math.floor(r / PARTY_SIZE) == i).length }))
				.filter(p => p.n > 0);
			parts.push(
				section(
					'Groups',
					table(
						['Group', { label: 'Simmed members', numeric: true }, { label: 'DPS', numeric: true }],
						parties.map(p => [`${p.i + 1}`, `${p.n}`, fmt(p.dps)]),
					),
					{ note: 'Placeholders aren’t simmed, so a group’s DPS is its DPS members’.' },
				),
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
			section(
				'Assignments as resolved',
				resolution.assignments.length ? list(resolution.assignments.map(a => this.assignmentText(a, slotById))) : empty('None.'),
			),
		);
		if (this.stale) {
			parts.unshift(
				<p className="rs-stale-note" attributes={{ role: 'status' }}>
					The raid changed after this run, so these results are of the raid as it was. Run it again to update them.
				</p>,
			);
		}
		this.resultsHost.replaceChildren(<h3 className="rs-results-title">The run</h3>, ...parts);
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
