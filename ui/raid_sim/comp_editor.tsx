// The comp editor (docs/raid_sim.md, The page, M2b): a raid of 10, 20 or 40 in parties of 5, each slot a
// DPS Report build or a tank or healer placeholder, with its assignments, the encounter, the faction and
// the camp buffs. Run builds a RaidComp, has the resolver turn it into the raid's request (resolveComp),
// runs that request in the sim's workers and shows the raid's results.
import { EncounterPicker } from '../core/components/encounter_picker';
import {
	Assignment,
	Blessing,
	CompMemberBuild,
	CompResult,
	CompScenarioResult,
	CompSearchRequest,
	CompSlot,
	Duty,
	PaladinJudgement,
	Player,
	ProgressMetrics,
	RaidAssumptions,
	RaidComp,
	RaidCompResolution,
	RaidPreset,
	RaidRole,
	RaidSimRequest,
	RaidSimResult,
	RaidTotem,
	SimOptions,
} from '../core/proto/api';
import { Class, Encounter as EncounterProto, Faction, MobType } from '../core/proto/common';
import { SimResult } from '../core/proto_utils/sim_result';
import { classNames, cssClassForClass } from '../core/proto_utils/utils';
import { RequestTypes } from '../core/sim_signal_manager';
import { TypedEvent } from '../core/typed_event';
import {
	BLESSING_CLASSES,
	blessingName,
	buildLabel,
	BuildVariant,
	CLASS_ICONS,
	dutyName,
	factionName,
	judgementName,
	loadRaidSimData,
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
import { badge, empty, fmt, fmtSigned, list, section, table } from './view_utils';

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
	private readonly encounterHost = (<div className="rs-encounter" />) as HTMLElement;
	private readonly runStatus = (<div className="rs-run-status" />) as HTMLElement;
	private readonly resultsHost = (<div className="rs-run-results" />) as HTMLElement;
	private readonly runButton: HTMLButtonElement;
	private drag: DragSource | null = null;
	// The palette entry picked with a click, placed by clicking a slot (touch screens have no drag and drop).
	private picked: Omit<Member, 'id'> | null = null;
	private running = false;
	private lastRun: RunOutput | null = null;

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

		this.rootElem = (
			<div className="rs-editor">
				<p className="rs-intro">
					Build a raid from the DPS Report’s builds and the tank and healer placeholders, drag members between groups, set assignments by hand or
					leave them automatic, and run it. Each member’s gear, talents, rotation and consumes come whole from its report build.
				</p>
				{this.controlsHost}
				<div className="rs-editor-layout">
					<div className="rs-editor-main">
						{section('The raid', this.raidHost, {
							note: 'Drag a member onto another slot to move or swap it, or pick one below and click a slot.',
						})}
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
					{this.runStatus}
				</div>
				{this.resultsHost}
			</div>
		) as HTMLElement;
		parent.appendChild(this.rootElem);

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
		const order = [...Array(parties).keys()].sort((a, b) => +packaged.has(a) - +packaged.has(b));
		others.forEach(p => {
			const party = order.find(q => freeIn(q) !== undefined);
			if (party === undefined) return;
			state.cells[freeIn(party)!] = { id: state.nextId++, preset: p };
			// Spread them: the next one goes to the next party.
			order.push(order.splice(order.indexOf(party), 1)[0]);
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
			<button className="btn btn-outline-secondary" type="button" title="Start over with this size’s default tanks and healers">
				New raid
			</button>
		) as HTMLButtonElement;
		newRaid.addEventListener('click', () => {
			const next = this.defaultState(this.state.size);
			this.state = { ...next, faction: s.faction, campBuffs: s.campBuffs, boss: s.boss, encounter: s.encounter };
			this.lastRun = null;
			this.changed();
			this.renderResults();
		});
		const clear = (
			<button className="btn btn-outline-secondary" type="button">
				Clear
			</button>
		) as HTMLButtonElement;
		clear.addEventListener('click', () => {
			this.state.cells = Array(this.state.size).fill(null);
			this.state.hand = emptyHand();
			this.changed();
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
						this.saveState();
					},
				),
				'each build takes its best race in it',
			),
			field(
				'Camp buffs',
				check('The raid relies on camp buffs', s.campBuffs, v => {
					this.state.campBuffs = v;
					this.saveState();
				}),
			),
			<div className="rs-editor-buttons">
				{newRaid}
				{clear}
			</div>,
		);
	}

	private resize(size: number) {
		const cells: Array<Member | null> = Array(size).fill(null);
		// Members keep their place; the ones past the new size are dropped.
		this.state.cells.forEach((m, i) => {
			if (m && i < size) cells[i] = m;
		});
		this.state.size = size;
		this.state.cells = cells;
		this.changed();
	}

	// ---- The raid: parties of 5 with drag and drop ----

	private renderRaid() {
		const parties = this.state.size / PARTY_SIZE;
		const run = this.lastRun;
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
	}

	private slotElem(cell: number, dpsOf: (id: number) => number | undefined) {
		const member = this.state.cells[cell];
		let elem: HTMLElement;
		if (!member) {
			elem = (
				<div className="player player-picker-root rs-slot rs-slot-empty" title="An empty slot: drop a member here, or pick one below and click">
					<span className="rs-slot-empty-text">empty</span>
				</div>
			) as HTMLElement;
		} else {
			const cls = this.classOf(member);
			const info = member.preset ? presetInfo(member.preset) : undefined;
			const icon = info?.icon ?? CLASS_ICONS[cls] ?? '';
			const dps = dpsOf(member.id);
			const remove = (
				<a href="javascript:void(0)" className="player-delete link-danger" attributes={{ role: 'button' }} title="Remove">
					<i className="fa fa-times fa-lg" />
				</a>
			) as HTMLElement;
			remove.addEventListener('click', e => {
				e.stopPropagation();
				this.state.cells[cell] = null;
				this.dropAssignmentsOf(member.id);
				this.changed();
			});
			elem = (
				<div className={`player player-picker-root rs-slot bg-${cssClassForClass(cls)}-dampened`} draggable={true}>
					<div className="player-label">
						<img className="player-icon" src={icon} draggable={false} />
						<div className="player-details">
							<span className={`rs-slot-name text-${cssClassForClass(cls)}`}>{this.memberName(member)}</span>
							<span className="rs-slot-sub">
								{info ? (info.role == RaidRole.RaidRoleTank ? 'Tank' : 'Healer') : 'DPS'}
								{dps !== undefined ? ` · ${fmt(dps)} DPS` : ''}
							</span>
						</div>
					</div>
					<div className="player-options">{remove}</div>
				</div>
			) as HTMLElement;
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
			if (!this.picked) return;
			this.dropOn(cell, { kind: 'new', member: this.picked });
		});
		if (this.picked) elem.classList.add('rs-slot-target');
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
					{icon ? <img className="rs-palette-icon" src={icon} draggable={false} /> : null}
					<span>{label}</span>
				</button>
			) as HTMLButtonElement;
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
	}

	private isPicked(member: Omit<Member, 'id'>) {
		return !!this.picked && this.picked.buildId == member.buildId && this.picked.preset == member.preset;
	}

	// ---- Assignments ----

	private resolvedText(job: (a: Assignment) => boolean): string | null {
		const found = this.lastRun?.resolution.assignments.find(job);
		if (!found) return null;
		switch (found.job.oneofKind) {
			case 'blessing':
				return blessingName(found.job.blessing.blessing);
			case 'judgement':
				return judgementName(found.job.judgement);
			case 'totem':
				return totemName(found.job.totem.totem);
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
								this.saveState();
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
								this.saveState();
							},
						),
					),
				);
			});
			rows.push(this.assignmentRow(member, cell, 'Greater Blessings for each class', controls));
		});

		// Each placeholder shaman: a totem for each element. A DPS shaman's rotation drops fixed totems.
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
									label: this.autoLabel(
										a => a.holder == member.id && a.job.oneofKind == 'totem' && totemElement(a.job.totem.totem) == element,
									),
								},
								...totems.map(t => ({ value: `${t}`, label: totemName(t) })),
							],
							h.totems[member.id]?.[element] !== undefined ? `${h.totems[member.id]![element]}` : '',
							v => {
								const mine = (h.totems[member.id] ??= {});
								if (v === '') delete mine[element];
								else mine[element] = parseInt(v);
								this.saveState();
							},
						),
					),
				);
				rows.push(this.assignmentRow(member, cell, 'Totems', controls));
			});

		// Duties: a DPS member with duty variants, and a druid placeholder's Faerie Fire.
		members.forEach(({ member, cell }) => {
			const duties = member.preset
				? presetInfo(member.preset)?.faerieFire
					? [Duty.DutyFaerieFire]
					: []
				: (this.buildOf(member)?.duties ?? []).map(d => d.duty!).filter(d => d !== undefined);
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
								{ value: `${Duty.DutyNone}`, label: 'none' },
							],
							current !== undefined ? `${current}` : '',
							v => {
								if (v === '') delete h.duties[member.id];
								else h.duties[member.id] = parseInt(v);
								this.saveState();
							},
						),
					),
				]),
			);
		});

		// Cooldowns: Power Infusion from a Discipline priest, Innervate from a Restoration druid (or a Feral
		// tank with it switched on). Only who gets them is chosen.
		const dpsMembers = members.filter(m => !m.member.preset);
		const manaUsers = dpsMembers.filter(m => ![Class.ClassWarrior, Class.ClassRogue].includes(this.classOf(m.member)));
		const cooldown = (holder: { member: Member; cell: number }, pi: boolean) => {
			const map = pi ? h.powerInfusion : h.innervate;
			const targets = pi ? dpsMembers : manaUsers;
			return field(
				pi ? 'Power Infusion' : 'Innervate',
				select(
					[
						{
							value: '',
							label: this.autoLabel(a => a.holder == holder.member.id && a.job.oneofKind == 'cooldown' && a.job.cooldown.powerInfusion == pi),
						},
						...targets.map(t => ({ value: `${t.member.id}`, label: this.memberShortName(t.member, t.cell) })),
					],
					map[holder.member.id] !== undefined ? `${map[holder.member.id]}` : '',
					v => {
						if (v === '') delete map[holder.member.id];
						else map[holder.member.id] = parseInt(v);
						this.saveState();
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
								this.changed();
							}),
						),
					);
				}
				if (info.innervate || m.member.presetInnervate) controls.push(cooldown(m, false));
				rows.push(this.assignmentRow(m.member, m.cell, 'Cooldown', controls));
			});

		const dpsShamans = dpsMembers.filter(m => this.classOf(m.member) == Class.ClassShaman);
		this.assignmentsHost.replaceChildren(
			...(rows.length ? rows : [empty('Nobody in this raid holds an assignment yet. Paladins, shamans, priests, druids and the duty builds do.')]),
			...(dpsShamans.length
				? [<p className="rs-muted">A DPS shaman’s rotation drops its own fixed totems, so it has no totem choice yet (docs/raid_sim.md, M1b).</p>]
				: []),
		);
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
				this.saveState();
			},
		);
		this.encounterHost.replaceChildren(field('Boss', bossSelect));
		const pickerHost = (<div className="rs-encounter-picker" />) as HTMLElement;
		this.encounterHost.appendChild(pickerHost);

		sim.waitForInit().then(() => {
			if (this.state.encounter) {
				try {
					sim.encounter.fromProto(TypedEvent.nextEventID(), EncounterProto.fromJson(this.state.encounter as any, { ignoreUnknownFields: true }));
				} catch (e) {
					console.warn('Failed to read the saved encounter:', e);
				}
			}
			new EncounterPicker(pickerHost, sim.encounter, { showExecuteProportion: true }, this.simUI);
			sim.encounter.changeEmitter.on(() => this.saveState());
		});
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
		this.simUI.sim.encounter.fromProto(TypedEvent.nextEventID(), encounter);
	}

	// ---- Load a comp from a search result ----

	loadFromResult(comp: CompResult, scenario: CompScenarioResult, request?: CompSearchRequest) {
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
					const totems: Partial<Record<TotemElement, RaidTotem>> = {};
					m.totems.forEach(t => {
						const e = totemElement(t);
						if (e) totems[e] = t;
					});
					state.hand.totems[member.id] = totems;
				}
			} else {
				// A duty variant holds its duty; the editor keeps the base build and the duty by hand.
				const found = this.data?.byId.get(m.buildId);
				member.buildId = found?.build.id ?? m.buildId;
				if (m.duties.length) state.hand.duties[member.id] = m.duties[0];
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
		if (e) {
			this.state.boss = e.name == 'Onyxia' ? 'Onyxia' : 'Generic';
			this.applyBoss(this.state.boss);
			const enc = this.simUI.sim.encounter;
			const eventID = TypedEvent.nextEventID();
			TypedEvent.freezeAllAndDo(() => {
				enc.setDuration(eventID, e.duration);
				enc.setDurationVariation(eventID, e.durationVariation);
				enc.setUseHealth(eventID, e.useHealth);
			});
		}
		this.lastRun = null;
		this.saveState();
		this.renderAll();
		this.renderResults();
		this.setStatus(
			`Loaded comp from the search (${factionName(scenario.faction)}, camps ${scenario.campBuffs ? 'on' : 'off'}): ${fmt(comp.dps)} raid DPS there.`,
		);
	}

	// ---- Run ----

	// The comp as the resolver takes it.
	async buildComp(): Promise<RaidComp> {
		const sim = this.simUI.sim;
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
			if (!build) return;
			// Against a creature type the build has a variant for, it runs that variant, as the search does.
			const variant = build.encounters.find(v => v.mobType == mobType) ?? build;
			const base = playerOf(variant, name);
			if (!base) return;
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
		Object.entries(h.totems).forEach(([holder, byElement]) =>
			Object.values(byElement).forEach(totem => hand(parseInt(holder), { oneofKind: 'totem', totem: { totem: totem as RaidTotem } })),
		);
		Object.entries(h.duties).forEach(([holder, duty]) => hand(parseInt(holder), { oneofKind: 'duty', duty }));
		Object.entries(h.powerInfusion).forEach(
			([holder, target]) => has(target) && hand(parseInt(holder), { oneofKind: 'cooldown', cooldown: { powerInfusion: true, target } }),
		);
		Object.entries(h.innervate).forEach(
			([holder, target]) => has(target) && hand(parseInt(holder), { oneofKind: 'cooldown', cooldown: { powerInfusion: false, target } }),
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
			await sim.signalManager.abortType(RequestTypes.RaidSim);
			return;
		}
		if (!this.data) {
			this.setStatus('The report’s builds haven’t loaded yet.', true);
			return;
		}
		if (!this.members().some(m => !m.member.preset)) {
			this.setStatus('The raid has no DPS member to sim. Add some from the report’s builds.', true);
			return;
		}
		this.running = true;
		this.setRunning(true);
		try {
			this.setStatus('Resolving the comp…');
			const comp = await this.buildComp();
			const resolution = await sim.resolveComp(comp);
			const out: RunOutput = { comp, resolution };
			this.lastRun = out;
			if (resolution.error || !resolution.request) {
				this.setStatus(`The resolver couldn’t resolve this comp: ${resolution.error?.message || 'it returned no request.'}`, true);
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
			this.setStatus(`Simming ${request.raid?.parties.flatMap(p => p.players).length ?? 0} players…`);
			const result = await sim.runRaidSimRequest(request, (progress: ProgressMetrics) => {
				this.simUI.resultsManager.setSimProgress(progress);
				if (progress.totalIterations)
					this.setStatus(`Simming: ${progress.completedIterations} of ${progress.totalIterations} iterations, ${fmt(progress.dps)} DPS`);
			});
			out.result = result;
			if (result.error) {
				this.setStatus(`The raid sim stopped: ${result.error.message || 'aborted'}`, true);
			} else {
				out.simResult = await SimResult.makeNew(request, result);
				this.simUI.resultsManager.setSimResult(TypedEvent.nextEventID(), out.simResult);
				this.setStatus(`Done: ${fmt(result.raidMetrics?.dps?.avg)} raid DPS.`);
			}
			this.renderRaid();
			this.renderAssignments();
			this.renderResults();
		} catch (e) {
			console.error(e);
			this.setStatus(`The run failed: ${e instanceof Error ? e.message : e}`, true);
		} finally {
			this.running = false;
			this.setRunning(false);
		}
	}

	private setRunning(running: boolean) {
		this.runButton.replaceChildren(<i className={`fas ${running ? 'fa-stop' : 'fa-play'} me-1`} />, running ? 'Stop' : 'Run');
		const sidebar = this.simUI.simActionsContainer.querySelector<HTMLButtonElement>('.dps-action');
		if (sidebar) sidebar.textContent = running ? 'Stop' : 'Run comp';
	}

	private setStatus(text: string, error = false) {
		this.runStatus.textContent = text;
		this.runStatus.classList.toggle('text-danger', error);
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
					<span className="rs-ci">±{fmt((1.96 * raidDps.stdev) / Math.sqrt(iterations))} (95%)</span>
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
						published ? fmtSigned(dps - published) : '',
						duties.join(', '),
						this.getsText(slot, resolution),
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
						['Group', 'Member', { label: 'DPS', numeric: true }, { label: 'vs the DPS Report', numeric: true }, 'Duties', 'Gets'],
						rows.map(r => r.row),
						'rs-table-run-members',
					),
					{ note: 'Each member’s DPS here, and its difference from its report build’s published DPS.' },
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
								simResult ? uptimeText(simResult, b.effect, b.scope) : '',
							]),
					  )
					: empty('The resolver recorded no buffs.'),
				{ note: 'From the resolver’s record. Uptimes come from the raid sim’s aura metrics, where the sim tracks an aura of that name.' },
			),
			section('What the raid goes without', resolution.misses.length ? list(resolution.misses, 'rs-misses') : empty('Nothing.')),
			section(
				'Assignments as resolved',
				resolution.assignments.length ? list(resolution.assignments.map(a => this.assignmentText(a, slotById))) : empty('None.'),
			),
		);
		this.resultsHost.replaceChildren(<h3 className="rs-results-title">The run</h3>, ...parts);
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
				job = `${blessingName(a.job.blessing.blessing)} to ${classNames[a.job.blessing.class]}s`;
				break;
			case 'judgement':
				job = `judges ${judgementName(a.job.judgement)}`;
				break;
			case 'totem':
				job = `drops ${totemName(a.job.totem.totem)}`;
				break;
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

const normalizeName = (s: string) =>
	s
		.toLowerCase()
		.replace(/\btotem\b|\bgreater\b|\bblessing of\b|\bprayer of\b|\bjudgement of( the)?\b/g, '')
		.replace(/[^a-z]/g, '');

// The uptime of an effect in a raid sim: the average over the units that carry an aura of its name, the
// boss for a debuff and the players for a buff.
const uptimeText = (simResult: SimResult, effect: string, scope: string): string => {
	const want = normalizeName(effect.split(/[(×]/)[0]);
	if (!want) return '';
	const units = /boss|target/i.test(scope) ? simResult.getTargets() : simResult.getPlayers();
	const uptimes: number[] = [];
	units.forEach(u => {
		const aura = u.auras.find(a => {
			const n = normalizeName(a.name);
			return n && (n == want || n.includes(want) || want.includes(n));
		});
		if (aura) uptimes.push(aura.uptimePercent);
	});
	if (!uptimes.length) return '—';
	return `${fmt(uptimes.reduce((a, b) => a + b, 0) / uptimes.length, 1)}%`;
};
