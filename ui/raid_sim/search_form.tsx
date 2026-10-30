// The comp search's form (docs/raid_sim.md, Comp search). It posts a CompSearchRequest to the native
// server's /compSearchAsync and shows the phase, the sims done and the best comp so far while it runs. A
// page served as static files has no server to search with, so its button is off.
import { CompEncounter, CompSearchRequest, CompSearchResult, ProgressMetrics, RaidPreset, RaidRole } from '../core/proto/api';
import { Faction, MobType } from '../core/proto/common';
import { RequestTypes } from '../core/sim_signal_manager';
import { buildLabel, loadRaidSimData, presetName, presetsFor, RAID_SIZES, RaidSimData } from './data';
import type { RaidSimUI } from './raid_sim_ui';
import { errorStatus, fmt, fmtInt, memberLabel, plural } from './view_utils';

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
	confirm: boolean;
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
	confirm: false,
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
		// The confirmation's iterations are the search's defaults (5,000, and 20,000 for the top 2).
		confirm: s.confirm,
	});
};

// Why a form can't search as it stands, if it can't. The search (sim/compsearch/search.go, Defaults)
// reads 0 tanks and 0 healers as the raid size's defaults, so the form doesn't send that: it would run a
// different raid from the one set here.
export const formProblems = (s: FormState): string[] => {
	const problems: string[] = [];
	const def = RAID_SIZES[s.size];
	if (s.tanks == 0 && s.healers == 0) {
		problems.push(
			`A raid with no tank and no healer slot can’t be searched: the search reads 0 and 0 as the size’s defaults (${def?.tanks ?? '?'} and ${
				def?.healers ?? '?'
			}). Set at least one tank or healer slot.`,
		);
	}
	if (s.size - s.tanks - s.healers < 1)
		problems.push(`${plural(s.tanks, 'tank')} and ${plural(s.healers, 'healer')} leave no DPS slot in a raid of ${s.size}.`);
	if (!s.horde && !s.alliance) problems.push('Pick a faction.');
	return problems;
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
	private readonly problemsElem: HTMLElement;
	private readonly progressElem: HTMLElement;
	// The number fields whose text isn't a valid value, with why. The state keeps their last valid value,
	// so the search waits until they're fixed.
	private readonly invalid = new Map<string, string>();
	private running = false;
	// The running search was aborted from the page.
	private aborted = false;
	private native = false;
	private readonly baseTitle = document.title;

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
		this.problemsElem = (<ul className="rs-form-problems text-danger" attributes={{ 'aria-live': 'polite' }} />) as HTMLElement;
		this.progressElem = (<div className="rs-progress d-none" attributes={{ 'aria-live': 'polite' }} />) as HTMLElement;

		parent.appendChild(
			<div className="rs-search">
				<p className="rs-intro">
					Finds the DPS builds, and the tank and healer presets, that give a raid the most DPS, for each faction and camp setting. It runs on the
					native server, on every core, and caches each single-player sim, so a search it has run before takes seconds.
				</p>
				{this.formHost}
				{this.problemsElem}
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
			this.aborted = true;
			this.showProgress(<p className="rs-muted">Aborting the search…</p>);
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
			this.staticNote.textContent = isWasm
				? 'Searching needs the native server (the downloadable sim or make host); this page is served as static files.'
				: 'Runs on the native server serving this page.';
			this.staticNote.classList.toggle('rs-static-off', isWasm);
			this.updateButton();
		});
	}

	private loadState(): FormState {
		try {
			const saved = window.localStorage.getItem(STORAGE_KEY);
			if (saved) {
				const s: FormState = { ...defaultState(), ...JSON.parse(saved) };
				// An older page took any number: the counts are whole numbers.
				s.tanks = Math.max(0, Math.round(s.tanks));
				s.healers = Math.max(0, Math.round(s.healers));
				return s;
			}
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
		else this.updateButton();
	}

	// A number field of the form. Counts and iterations are whole numbers.
	private num<K extends keyof FormState>(key: K, min: number, max: number, opts: { integer?: boolean; rerender?: boolean } = {}) {
		const integer = opts.integer ?? true;
		return numberInput(this.state[key] as number, min, max, v => this.set(key, v as FormState[K], opts.rerender), {
			integer,
			onInvalid: reason => {
				if (reason) this.invalid.set(key, reason);
				else this.invalid.delete(key);
				this.updateButton();
			},
		});
	}

	private render() {
		const s = this.state;
		const sizeInfo = RAID_SIZES[s.size];
		const dpsSlots = s.size - s.tanks - s.healers;
		this.invalid.clear();

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
		const tanksInput = this.num('tanks', 0, s.size - 1, { rerender: true });
		const healersInput = this.num('healers', 0, s.size - 1, { rerender: true });

		const slotPickers = (role: RaidRole, count: number, fixed: RaidPreset[], key: 'fixedTanks' | 'fixedHealers') =>
			[...Array(count).keys()].map(i => {
				const elem = select(
					[{ value: '0', label: 'any (searched)' }, ...presetsFor(role).map(p => ({ value: `${p.id}`, label: p.name }))],
					`${fixed[i] ?? 0}`,
					v => {
						const next = [...this.state[key]];
						while (next.length <= i) next.push(RaidPreset.RaidPresetUnknown);
						next[i] = parseInt(v);
						this.set(key, next);
					},
				);
				elem.setAttribute('aria-label', `${role == RaidRole.RaidRoleTank ? 'Tank' : 'Healer'} slot ${i + 1}`);
				return elem;
			});

		// The fixed DPS members: a guild's roster, the rest filled by the search.
		const fixedDps = (
			<div className="rs-fixed-dps">
				{s.fixedDps.map((id, i) => {
					const remove = (
						<button className="btn btn-link link-danger rs-chip-remove" type="button" title="Remove" attributes={{ 'aria-label': `Remove ${id}` }}>
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
					(() => {
						const add = select(
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
						);
						add.setAttribute('aria-label', 'Add a fixed DPS member');
						return add;
					})()
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
				{field(s.useHealth ? 'First guess (s)' : 'Duration (s)', this.num('duration', 10, 1200, { integer: false }))}
				{s.useHealth ? field('Boss health', this.num('bossHealth', 1, 1e9)) : field('Variation (s)', this.num('variation', 0, 600, { integer: false }))}
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
						{field('Iterations', this.num('iterations', 10, 100000), 'the screening sims')}
						{field('Contender iterations', this.num('contenderIterations', 10, 1000000), 'the contenders’ re-sims')}
						{field('Max combos', this.num('maxCombos', 0, 100000), '0 ranks every tank and healer combo')}
						{field(
							'Sensitivity',
							check('Skip the sensitivity scenarios', s.skipSensitivity, v => this.set('skipSensitivity', v)),
						)}
						{field(
							'Confirm',
							check('Raid-sim the top comps', s.confirm, v => this.set('confirm', v)),
							'whole-raid sims of the best buff sets, which then rank them; slower',
						)}
					</div>
				</fieldset>
			</div>,
		);
		this.updateButton();
	}

	// Why the form can't search now: its own problems and the fields that hold no valid value.
	private problems(): string[] {
		return [...formProblems(this.state), ...[...this.invalid.entries()].map(([key, reason]) => `${fieldNames[key] ?? key}: ${reason}.`)];
	}

	private updateButton() {
		const problems = this.problems();
		this.problemsElem.replaceChildren(...problems.map(p => <li>{p}</li>));
		this.searchButton.disabled = !this.native || this.running || problems.length > 0;
	}

	private async search() {
		if (this.running) return;
		if (this.problems().length) {
			this.updateButton();
			return;
		}
		const request = searchRequest(this.state);
		this.running = true;
		this.aborted = false;
		this.searchButton.disabled = true;
		this.abortButton.disabled = false;
		this.abortButton.classList.remove('d-none');
		this.showProgress(<p>Starting the search…</p>);
		this.setTitle('Searching');
		const started = Date.now();
		const seconds = () => fmt((Date.now() - started) / 1000, 0);
		try {
			const result = await this.simUI.sim.runCompSearch(request, (progress: ProgressMetrics) => this.onProgress(progress, started));
			if (result.error && !result.scenarios.length && this.aborted) {
				// Nothing of an aborted search is kept, so its last progress (the best so far) goes too.
				this.showProgress(<p className="rs-muted">Search aborted after {seconds()} s. Nothing was kept.</p>);
				this.finishTitle('Search aborted');
			} else if (result.error && !result.scenarios.length) {
				this.showProgress(<div className="alert alert-danger">{errorStatus('The search stopped', result.error.message)}</div>);
				this.finishTitle('Search failed');
			} else {
				this.showProgress(<p>Done in {seconds()} s. The result is in the Search result tab.</p>);
				this.finishTitle('Search done');
				this.onResult(result);
			}
		} catch (e) {
			this.showProgress(<div className="alert alert-danger">{errorStatus('The search failed', `${e instanceof Error ? e.message : e}`)}</div>);
			this.finishTitle('Search failed');
		} finally {
			this.running = false;
			this.abortButton.classList.add('d-none');
			this.render();
		}
	}

	private onProgress(progress: ProgressMetrics, started: number) {
		if (progress.finalCompSearchResult || this.aborted) return;
		const best = progress.compSearchBest;
		this.setTitle(`Searching: ${progress.compSearchPhase || 'starting'}`);
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

	// The page's title says how a search goes, for a page in a background tab. Its outcome stays there until
	// the page is looked at again.
	private setTitle(status: string | null) {
		document.title = status ? `${status} · ${this.baseTitle}` : this.baseTitle;
	}

	private finishTitle(status: string) {
		if (!document.hidden) {
			this.setTitle(null);
			return;
		}
		this.setTitle(status);
		const back = () => {
			if (document.hidden) return;
			document.removeEventListener('visibilitychange', back);
			if (!this.running) this.setTitle(null);
		};
		document.addEventListener('visibilitychange', back);
	}

	private showProgress(content: Element) {
		this.progressElem.classList.remove('d-none');
		this.progressElem.replaceChildren(content);
	}
}

const fieldNames: Record<string, string> = {
	tanks: 'Tank slots',
	healers: 'Healer slots',
	duration: 'Duration',
	variation: 'Variation',
	bossHealth: 'Boss health',
	iterations: 'Iterations',
	contenderIterations: 'Contender iterations',
	maxCombos: 'Max combos',
};

const rangeText = (r?: [number, number]) => (!r ? '' : r[0] == r[1] ? `${r[0]}` : `${r[0]} to ${r[1]}`);

// Plain form controls, styled like the sim's pickers.

let controlSeq = 0;
const nextId = (prefix: string) => `${prefix}-${++controlSeq}`;

const isFormControl = (n: Node): n is HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement =>
	n instanceof HTMLInputElement || n instanceof HTMLSelectElement || n instanceof HTMLTextAreaElement;

// A labelled field. A single form control gets a <label for> of its own. A field of several controls, or
// of a checkbox with its own label, is a group the label names, so a click on that label toggles nothing.
export const field = (label: string, control: Node, hint?: string): HTMLElement => {
	const hintElem = hint ? ((<span className="rs-field-hint">{hint}</span>) as HTMLElement) : null;
	if (hintElem) hintElem.id = nextId('rs-hint');
	if (isFormControl(control) && control.type != 'checkbox') {
		if (!control.id) control.id = nextId('rs-control');
		if (hintElem) control.setAttribute('aria-describedby', hintElem.id);
		return (
			<div className="rs-field">
				<label className="form-label rs-field-label" htmlFor={control.id}>
					{label}
				</label>
				{control}
				{hintElem}
			</div>
		) as HTMLElement;
	}
	const labelElem = (<span className="form-label rs-field-label">{label}</span>) as HTMLElement;
	labelElem.id = nextId('rs-label');
	const group = (
		<div className="rs-field" attributes={{ role: 'group' }}>
			{labelElem}
			{control}
			{hintElem}
		</div>
	) as HTMLElement;
	group.setAttribute('aria-labelledby', labelElem.id);
	if (hintElem) group.setAttribute('aria-describedby', hintElem.id);
	return group;
};

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

// A number input. A value out of range, or not a whole number where one is needed, is marked invalid and
// not taken: onInvalid hears why, and onChange gets only valid values.
export const numberInput = (
	value: number,
	min: number,
	max: number,
	onChange: (value: number) => void,
	opts: { integer?: boolean; onInvalid?: (reason: string | null) => void } = {},
) => {
	const elem = (
		<input className="form-control rs-number" type="number" value={`${value}`} min={`${min}`} max={`${max}`} step={opts.integer ? '1' : 'any'} />
	) as HTMLInputElement;
	elem.inputMode = opts.integer ? 'numeric' : 'decimal';
	elem.addEventListener('change', () => {
		const text = elem.value.trim();
		const v = text === '' ? NaN : Number(text);
		let reason: string | null = null;
		if (isNaN(v)) reason = 'it needs a number';
		else if (opts.integer && !Number.isInteger(v)) reason = 'it needs a whole number';
		else if (v < min || v > max) reason = `it needs ${min} to ${max}`;
		elem.classList.toggle('is-invalid', !!reason);
		elem.setAttribute('aria-invalid', reason ? 'true' : 'false');
		elem.title = reason ? `Not taken: ${reason}` : '';
		opts.onInvalid?.(reason);
		if (!reason) onChange(v);
	});
	return elem;
};

// A checkbox with a label of its own.
export const check = (label: string, value: boolean, onChange: (value: boolean) => void) => {
	const input = (<input className="form-check-input" type="checkbox" checked={value} />) as HTMLInputElement;
	input.id = nextId('rs-check');
	input.addEventListener('change', () => onChange(input.checked));
	return (
		<span className="form-check rs-check">
			{input}
			<label className="form-check-label" htmlFor={input.id}>
				{label}
			</label>
		</span>
	);
};
