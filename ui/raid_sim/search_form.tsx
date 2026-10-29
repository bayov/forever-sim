// The comp search's form (docs/raid_sim.md, Comp search). It posts a CompSearchRequest to the native
// server's /compSearchAsync and shows the phase, the sims done and the best comp so far while it runs. A
// page served as static files has no server to search with, so its button is off.
import { CompEncounter, CompSearchRequest, CompSearchResult, ProgressMetrics, RaidPreset, RaidRole } from '../core/proto/api';
import { Faction, MobType } from '../core/proto/common';
import { RequestTypes } from '../core/sim_signal_manager';
import { buildLabel, loadRaidSimData, presetName, presetsFor, RAID_SIZES, RaidSimData } from './data';
import type { RaidSimUI } from './raid_sim_ui';
import { fmt, fmtInt, memberLabel } from './view_utils';

const STORAGE_KEY = '__forever_raid_sim__searchForm__';

type Boss = 'Generic' | 'Onyxia';
type Camps = 'off' | 'on' | 'both';

interface FormState {
	size: number;
	tanks: number;
	healers: number;
	fixedTanks: RaidPreset[]; // per tank slot, Unknown for any
	fixedHealers: RaidPreset[];
	fixedDps: string[];
	boss: Boss;
	useHealth: boolean;
	duration: number;
	variation: number;
	bossHealth: number;
	horde: boolean;
	alliance: boolean;
	camps: Camps;
	iterations: number;
	contenderIterations: number;
	skipSensitivity: boolean;
	maxCombos: number;
}

const defaultState = (): FormState => ({
	size: 20,
	tanks: RAID_SIZES[20].tanks,
	healers: RAID_SIZES[20].healers,
	fixedTanks: [],
	fixedHealers: [],
	fixedDps: [],
	boss: 'Generic',
	useHealth: false,
	duration: 180,
	variation: 20,
	bossHealth: 2_000_000,
	horde: true,
	alliance: true,
	camps: 'both',
	iterations: 1000,
	contenderIterations: 10000,
	skipSensitivity: false,
	maxCombos: 0,
});

// The request a form makes, with the encounter set the way tools/compsearch sets it: Generic is the
// report's boss in the natural flask zone, Onyxia a Dragonkin outside it.
export const searchRequest = (s: FormState): CompSearchRequest => {
	const encounter = CompEncounter.create({
		name: s.boss,
		duration: s.duration,
		durationVariation: s.variation,
		useHealth: s.useHealth,
		bossHealth: s.useHealth ? s.bossHealth : 0,
	});
	if (s.boss == 'Onyxia') encounter.mobType = MobType.MobTypeDragonkin;
	else encounter.naturalFlaskZone = true;
	const factions: Faction[] = [];
	if (s.horde) factions.push(Faction.Horde);
	if (s.alliance) factions.push(Faction.Alliance);
	return CompSearchRequest.create({
		size: s.size,
		tanks: s.tanks,
		healers: s.healers,
		// Fixed presets go to the first slots, so the ones set come first.
		fixedTanks: s.fixedTanks.slice(0, s.tanks).filter(p => p != RaidPreset.RaidPresetUnknown),
		fixedHealers: s.fixedHealers.slice(0, s.healers).filter(p => p != RaidPreset.RaidPresetUnknown),
		fixedDps: s.fixedDps.slice(0, s.size - s.tanks - s.healers),
		encounter,
		factions,
		campBuffs: s.camps == 'both' ? [false, true] : [s.camps == 'on'],
		iterations: s.iterations,
		contenderIterations: s.contenderIterations,
		skipSensitivity: s.skipSensitivity,
		maxCombos: s.maxCombos,
	});
};

export class SearchForm {
	private readonly simUI: RaidSimUI;
	private readonly onResult: (result: CompSearchResult) => void;
	private state: FormState;
	private data: RaidSimData | null = null;
	private readonly formHost: HTMLElement;
	private readonly searchButton: HTMLButtonElement;
	private readonly abortButton: HTMLButtonElement;
	private readonly staticNote: HTMLElement;
	private readonly progressElem: HTMLElement;
	private running = false;
	private native = false;

	constructor(parent: HTMLElement, simUI: RaidSimUI, onResult: (result: CompSearchResult) => void) {
		this.simUI = simUI;
		this.onResult = onResult;
		this.state = this.loadState();

		this.formHost = (<div className="rs-form" />) as HTMLElement;
		this.searchButton = (
			<button className="btn btn-primary rs-search-button" type="button" disabled>
				<i className="fas fa-magnifying-glass me-1" />
				Search
			</button>
		) as HTMLButtonElement;
		this.abortButton = (
			<button className="btn btn-outline-danger rs-abort-button d-none" type="button">
				<i className="fas fa-stop me-1" />
				Abort
			</button>
		) as HTMLButtonElement;
		this.staticNote = (<p className="rs-static-note">Checking for the native server…</p>) as HTMLElement;
		this.progressElem = (<div className="rs-progress d-none" />) as HTMLElement;

		parent.appendChild(
			<div className="rs-search">
				<p className="rs-intro">
					Finds the DPS builds, and the tank and healer presets, that give a raid the most DPS, for each faction and camp setting. It runs on the
					native server, on every core, and caches each single-player sim, so a search it has run before takes seconds.
				</p>
				{this.formHost}
				<div className="rs-search-actions">
					{this.searchButton}
					{this.abortButton}
					{this.staticNote}
				</div>
				{this.progressElem}
			</div>,
		);

		this.searchButton.addEventListener('click', () => this.search());
		this.abortButton.addEventListener('click', async () => {
			this.abortButton.disabled = true;
			await this.simUI.sim.signalManager.abortType(RequestTypes.CompSearch);
		});

		this.render();
		loadRaidSimData()
			.then(data => {
				this.data = data;
				this.render();
			})
			// Without the data file the fixed DPS members are just ids.
			.catch(() => this.render());

		// The pages send their sims to the native server when it serves them: its workers are net workers,
		// not WebAssembly. Served as static files, there's no server to search with.
		this.simUI.sim.isWasm().then(isWasm => {
			this.native = !isWasm;
			this.searchButton.disabled = isWasm;
			this.staticNote.textContent = isWasm
				? 'Searching needs the native server (the downloadable sim or make host); this page is served as static files.'
				: 'Runs on the native server serving this page.';
			this.staticNote.classList.toggle('rs-static-off', isWasm);
		});
	}

	private loadState(): FormState {
		try {
			const saved = window.localStorage.getItem(STORAGE_KEY);
			if (saved) return { ...defaultState(), ...JSON.parse(saved) };
		} catch (e) {
			console.warn('Failed to read the saved search form:', e);
		}
		return defaultState();
	}

	private saveState() {
		try {
			window.localStorage.setItem(STORAGE_KEY, JSON.stringify(this.state));
		} catch (e) {
			// A private window: the form just isn't remembered.
		}
	}

	private set<K extends keyof FormState>(key: K, value: FormState[K], rerender = false) {
		this.state[key] = value;
		this.saveState();
		if (rerender) this.render();
	}

	private render() {
		const s = this.state;
		const sizeInfo = RAID_SIZES[s.size];
		const dpsSlots = s.size - s.tanks - s.healers;

		const sizeSelect = select(
			[10, 20, 40].map(n => ({ value: `${n}`, label: `${n}` })),
			`${s.size}`,
			v => {
				const size = parseInt(v);
				this.state.size = size;
				this.state.tanks = RAID_SIZES[size].tanks;
				this.state.healers = RAID_SIZES[size].healers;
				this.saveState();
				this.render();
			},
		);
		const tanksInput = numberInput(s.tanks, 0, s.size - 1, v => this.set('tanks', v, true));
		const healersInput = numberInput(s.healers, 0, s.size - 1, v => this.set('healers', v, true));

		const slotPickers = (role: RaidRole, count: number, fixed: RaidPreset[], key: 'fixedTanks' | 'fixedHealers') =>
			[...Array(count).keys()].map(i =>
				select(
					[{ value: '0', label: 'any (searched)' }, ...presetsFor(role).map(p => ({ value: `${p.id}`, label: p.name }))],
					`${fixed[i] ?? 0}`,
					v => {
						const next = [...this.state[key]];
						while (next.length <= i) next.push(RaidPreset.RaidPresetUnknown);
						next[i] = parseInt(v);
						this.set(key, next);
					},
				),
			);

		// The fixed DPS members: a guild's roster, the rest filled by the search.
		const fixedDps = (
			<div className="rs-fixed-dps">
				{s.fixedDps.map((id, i) => {
					const remove = (
						<button className="btn btn-link link-danger rs-chip-remove" type="button" title="Remove">
							<i className="fas fa-times" />
						</button>
					) as HTMLButtonElement;
					remove.addEventListener('click', () => {
						const next = [...this.state.fixedDps];
						next.splice(i, 1);
						this.set('fixedDps', next, true);
					});
					return (
						<span className={`rs-chip ${i >= dpsSlots ? 'rs-chip-over' : ''}`}>
							{memberLabel(this.data, id)}
							{remove}
						</span>
					);
				})}
				{this.data ? (
					select(
						[
							{ value: '', label: s.fixedDps.length >= dpsSlots ? 'Every DPS slot is fixed' : 'Add a member…' },
							...this.data.builds.map(b => ({ value: b.id, label: buildLabel(b) })),
						],
						'',
						v => {
							if (!v) return;
							this.set('fixedDps', [...this.state.fixedDps, v], true);
						},
						s.fixedDps.length >= dpsSlots,
					)
				) : (
					<span className="rs-muted">Loading the report’s builds…</span>
				)}
			</div>
		);

		const encounter = (
			<div className="rs-field-group">
				{field(
					'Boss',
					select(
						[
							{ value: 'Generic', label: 'Generic (level 63, natural flask zone)' },
							{ value: 'Onyxia', label: 'Onyxia (Dragonkin, outside the flask zone)' },
						],
						s.boss,
						v => this.set('boss', v as Boss),
					),
				)}
				{field(
					'Length',
					select(
						[
							{ value: 'duration', label: 'A fixed duration' },
							{ value: 'health', label: 'The boss’s health' },
						],
						s.useHealth ? 'health' : 'duration',
						v => this.set('useHealth', v == 'health', true),
					),
				)}
				{field(
					s.useHealth ? 'First guess (s)' : 'Duration (s)',
					numberInput(s.duration, 10, 1200, v => this.set('duration', v)),
				)}
				{s.useHealth
					? field(
							'Boss health',
							numberInput(s.bossHealth, 1, 1e9, v => this.set('bossHealth', v)),
					  )
					: field(
							'Variation (s)',
							numberInput(s.variation, 0, 600, v => this.set('variation', v)),
					  )}
			</div>
		);

		this.formHost.replaceChildren(
			<div className="rs-form-grid">
				<fieldset className="rs-fieldset">
					<legend>The raid</legend>
					<div className="rs-field-group">
						{field('Raid size', sizeSelect)}
						{field('Tank slots', tanksInput, `usually ${rangeText(sizeInfo?.tankRange)}`)}
						{field('Healer slots', healersInput, `usually ${rangeText(sizeInfo?.healerRange)}`)}
						{field('DPS slots', <span className={`rs-value ${dpsSlots < 1 ? 'text-danger' : ''}`}>{dpsSlots}</span>)}
					</div>
					{s.tanks > 0
						? field(
								'Tank presets',
								<div className="rs-slot-pickers">{slotPickers(RaidRole.RaidRoleTank, s.tanks, s.fixedTanks, 'fixedTanks')}</div>,
						  )
						: null}
					{s.healers > 0
						? field(
								'Healer presets',
								<div className="rs-slot-pickers">{slotPickers(RaidRole.RaidRoleHealer, s.healers, s.fixedHealers, 'fixedHealers')}</div>,
						  )
						: null}
					{field('Fixed DPS members', fixedDps, 'report builds that must be in the comp')}
				</fieldset>
				<fieldset className="rs-fieldset">
					<legend>The fight</legend>
					{encounter}
					<div className="rs-field-group">
						{field(
							'Factions',
							<div className="rs-checks">
								{check('Horde', s.horde, v => this.set('horde', v))}
								{check('Alliance', s.alliance, v => this.set('alliance', v))}
							</div>,
						)}
						{field(
							'Camp buffs',
							select(
								[
									{ value: 'off', label: 'off' },
									{ value: 'on', label: 'on' },
									{ value: 'both', label: 'both' },
								],
								s.camps,
								v => this.set('camps', v as Camps),
							),
						)}
					</div>
				</fieldset>
				<fieldset className="rs-fieldset">
					<legend>The search</legend>
					<div className="rs-field-group">
						{field(
							'Iterations',
							numberInput(s.iterations, 10, 100000, v => this.set('iterations', v)),
							'the screening sims',
						)}
						{field(
							'Contender iterations',
							numberInput(s.contenderIterations, 10, 1000000, v => this.set('contenderIterations', v)),
							'the contenders’ re-sims',
						)}
						{field(
							'Max combos',
							numberInput(s.maxCombos, 0, 100000, v => this.set('maxCombos', v)),
							'0 ranks every tank and healer combo',
						)}
						{field(
							'Sensitivity',
							check('Skip the sensitivity scenarios', s.skipSensitivity, v => this.set('skipSensitivity', v)),
						)}
					</div>
				</fieldset>
			</div>,
		);
		this.searchButton.disabled = !this.native || this.running || dpsSlots < 1;
	}

	private async search() {
		if (this.running) return;
		const request = searchRequest(this.state);
		if (!request.factions.length) {
			this.showProgress(<p className="text-danger">Pick a faction.</p>);
			return;
		}
		this.running = true;
		this.searchButton.disabled = true;
		this.abortButton.disabled = false;
		this.abortButton.classList.remove('d-none');
		this.showProgress(<p>Starting the search…</p>);
		const started = Date.now();
		try {
			const result = await this.simUI.sim.runCompSearch(request, (progress: ProgressMetrics) => this.onProgress(progress, started));
			if (result.error && !result.scenarios.length) {
				this.showProgress(<div className="alert alert-danger">The search stopped: {result.error.message}</div>);
			} else {
				this.showProgress(<p>Done in {fmt((Date.now() - started) / 1000, 0)} s.</p>);
				this.onResult(result);
			}
		} catch (e) {
			this.showProgress(<div className="alert alert-danger">The search failed: {`${e}`}</div>);
		} finally {
			this.running = false;
			this.abortButton.classList.add('d-none');
			this.render();
		}
	}

	private onProgress(progress: ProgressMetrics, started: number) {
		if (progress.finalCompSearchResult) return;
		const best = progress.compSearchBest;
		this.showProgress(
			<div className="rs-progress-body">
				<div className="rs-progress-line">
					<span className="rs-progress-phase">
						<i className="fas fa-spinner fa-spin me-2" />
						{progress.compSearchPhase || 'searching'}
					</span>
					<span>{fmtInt(progress.completedSims)} sims</span>
					<span>{fmt((Date.now() - started) / 1000, 0)} s</span>
				</div>
				{best ? (
					<div className="rs-progress-best">
						Best so far: <strong>{fmt(best.dps)} raid DPS</strong>
						{best.combo.length ? `, ${best.combo.map(presetName).join(', ')}` : ''}
						{best.buffSet ? <div className="rs-muted">{best.buffSet}</div> : null}
					</div>
				) : null}
			</div>,
		);
	}

	private showProgress(content: Element) {
		this.progressElem.classList.remove('d-none');
		this.progressElem.replaceChildren(content);
	}
}

const rangeText = (r?: [number, number]) => (!r ? '' : r[0] == r[1] ? `${r[0]}` : `${r[0]} to ${r[1]}`);

// Plain form controls, styled like the sim's pickers.
export const field = (label: string, control: Node, hint?: string) => (
	<label className="rs-field">
		<span className="form-label rs-field-label">{label}</span>
		{control}
		{hint ? <span className="rs-field-hint">{hint}</span> : null}
	</label>
);

export const select = (options: { value: string; label: string }[], value: string, onChange: (value: string) => void, disabled = false) => {
	const elem = (
		<select className="form-select rs-select" disabled={disabled}>
			{options.map(o => (
				<option value={o.value} selected={o.value == value}>
					{o.label}
				</option>
			))}
		</select>
	) as HTMLSelectElement;
	elem.addEventListener('change', () => onChange(elem.value));
	return elem;
};

export const numberInput = (value: number, min: number, max: number, onChange: (value: number) => void) => {
	const elem = (<input className="form-control rs-number" type="number" value={`${value}`} min={`${min}`} max={`${max}`} />) as HTMLInputElement;
	elem.addEventListener('change', () => {
		const v = parseFloat(elem.value);
		if (!isNaN(v)) onChange(Math.min(max, Math.max(min, v)));
	});
	return elem;
};

export const check = (label: string, value: boolean, onChange: (value: boolean) => void) => {
	const input = (<input className="form-check-input" type="checkbox" checked={value} />) as HTMLInputElement;
	input.addEventListener('change', () => onChange(input.checked));
	return (
		<span className="form-check rs-check">
			{input}
			<span className="form-check-label">{label}</span>
		</span>
	);
};
