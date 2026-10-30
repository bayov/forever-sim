import { Chart, Plugin } from 'chart.js';
import tippy, { delegate } from 'tippy.js';
import { ref } from 'tsx-vanilla';

import { TOOLTIP_METRIC_LABELS } from '../../constants/tooltips';
import { DistributionMetrics, ResourceType } from '../../proto/api';
import { ActionId } from '../../proto_utils/action_id';
import { resourceNames } from '../../proto_utils/names';
import { ActionMetrics, AuraMetrics, ResourceMetrics, SimResult, UnitMetrics } from '../../proto_utils/sim_result';
import { orderedResourceTypes } from '../../proto_utils/utils';
import type { SimUI } from '../../sim_ui';
import { bucket, formatToNumber } from '../../utils';
import { Component } from '../component';
import { panel, syncSubTabWithUrl, trackStuckToolbar } from '../detailed_results';
import { getLevelingMetrics, mobsKilledError } from '../leveling_results';
import { changesList } from '../preset_tree';
import { RaidSimResultsManager, ResultMetric } from '../raid_sim_action';
import { runLabel, SimHistory, SimRun } from '../sim_history';
import { Better, Breakdown, BreakdownLine, changeElem, CompareColumn, CompareRow, CompareTable, formatFixed, formatPercent } from './compare_table';

// A side of the comparison is a run from the history, or the latest or the one before it. The
// last two follow new runs, so after each Simulate we see what the change did.
type RunChoice = 'latest' | 'previous' | string;

type Side = 'a' | 'b';

const SIDE_NAMES: Record<Side, string> = { a: 'Baseline', b: 'Compared' };

const CHANGED_ONLY_KEY = '__compare_changed_only__';

// The topline results where less is better.
const LOWER_IS_BETTER = new Set(['DTPS', 'TMI', 'COD', 'OOM', 'DUR', 'Deaths']);

// Bigger changes than this many standard errors are real, the same Z test at 95% as the sim
// history.
const Z_SIGNIFICANT = 1.96;

const COLOR_A = '#7fb3f0';
const COLOR_A_BAR = 'rgba(127, 179, 240, 0.5)';
const COLOR_B = '#c8a24a';
const COLOR_B_BAR = 'rgba(200, 162, 74, 0.5)';
const MUTED = '#8e948f';

// Each point of the distribution gets at least this many pixels.
// The same bins as the DPS Distribution of the Results tab, see DpsHistogram.
const MIN_BIN_PX = 6;
const MIN_ITERATIONS_PER_BIN = 10;
const NICE_BIN_WIDTHS = [1, 2, 3, 4, 5, 10, 15, 20, 25, 50, 100, 200, 250, 500, 1000];

// The rows in the list of biggest changes. The rest go in one Other row.
const TOP_CHANGES = 10;

type ChangesMode = 'dps' | 'casts' | 'uptime';
type DistributionMode = 'distribution' | 'cumulative';

interface Tab {
	id: string;
	label: string;
	icon: string;
	className?: string;
}

const tabs: Tab[] = [
	{ id: 'compareOverviewTab', label: 'Overview', icon: 'fa-scale-balanced' },
	{ id: 'compareDamageTab', label: 'Damage', icon: 'fa-bolt' },
	{ id: 'compareThreatTab', label: 'Threat', icon: 'fa-skull', className: 'cmp-threat-tab' },
	{ id: 'compareHealingTab', label: 'Healing', icon: 'fa-heart-pulse', className: 'cmp-healing-tab' },
	{ id: 'compareDamageTakenTab', label: 'Damage Taken', icon: 'fa-shield-halved' },
	{ id: 'compareAurasTab', label: 'Auras', icon: 'fa-wand-magic-sparkles' },
	{ id: 'compareCastsTab', label: 'Casts', icon: 'fa-hand-sparkles' },
	{ id: 'compareResourcesTab', label: 'Resources', icon: 'fa-droplet' },
];

// The two runs we compare, with their results.
interface Pair {
	a: SimRun;
	b: SimRun;
	resultA: SimResult;
	resultB: SimResult;
}

// The Compare tab puts two runs from the sim history side by side: their results, where the
// difference in DPS comes from, the settings that changed between them, and each spell, aura,
// cast and resource with its change.
//
// It compares the latest run with the one before it until we pick other runs. The history keeps
// the full results of each run in the browser, see SimHistory. Runs from before we kept them
// only have their DPS, so we can't pick them.
export class CompareResults extends Component {
	static readonly VIEW_PARAM = 'compare';

	private readonly simUI: SimUI;
	private readonly history: SimHistory;
	// What the empty state suggests changing between two runs, like "a talent or an item".
	private readonly changeExample: string;
	// Who a member is across runs, from its name. The raid page's names changed form, so its old and new
	// runs name one member two ways.
	private readonly memberKey: (name: string) => string;
	// The raid member we compare, by name, or '' for the whole raid. A spec page's runs have one player, so
	// there it's always ''.
	private member = '';
	private readonly memberElem: HTMLElement;
	private readonly memberSelect: HTMLSelectElement;
	private choices: Record<Side, RunChoice> = { a: 'previous', b: 'latest' };
	private readonly selects: Record<Side, HTMLSelectElement>;
	private readonly metaElems: Record<Side, HTMLElement>;
	private readonly contentElem: HTMLElement;
	private readonly emptyElem: HTMLElement;
	private readonly emptyTextElem: HTMLElement;
	// The runs we drew last, so a history change that keeps them draws nothing.
	private drawnKey = '';
	// We draw when the tab shows, so runs that come while we look at other tabs cost nothing.
	private pendingDraw = false;
	private drawToken = 0;
	private readonly resultCache = new Map<string, SimResult>();

	private changesMode: ChangesMode = 'dps';
	private distributionMode: DistributionMode = 'distribution';
	private pair: Pair | null = null;
	private chart: Chart | null = null;
	// The width we drew the distribution at, see the resize observer.
	private distributionWidth = 0;

	private readonly overviewElems: {
		verdict: HTMLElement;
		cards: HTMLElement;
		changes: HTMLElement;
		settings: HTMLElement;
		distribution: HTMLElement;
		changesButtons: HTMLButtonElement[];
		distributionButtons: HTMLButtonElement[];
	};
	private readonly resourceCardsElem: HTMLElement;
	private readonly tables: {
		damage: CompareTable<ActionMetrics>;
		threat: CompareTable<ActionMetrics>;
		healing: CompareTable<ActionMetrics>;
		damageTaken: CompareTable<ActionMetrics>;
		buffs: CompareTable<AuraMetrics>;
		debuffs: CompareTable<AuraMetrics>;
		casts: CompareTable<ActionMetrics>;
		resources: Map<ResourceType, { panel: HTMLElement; table: CompareTable<ResourceMetrics> }>;
	};

	constructor(parent: HTMLElement, simUI: SimUI, history: SimHistory, opts: { changeExample?: string; memberKey?: (name: string) => string } = {}) {
		super(parent, 'cmp-root');
		this.simUI = simUI;
		this.changeExample = opts.changeExample ?? 'a talent or an item';
		this.memberKey = opts.memberKey ?? (name => name);
		this.history = history;
		this.memberSelect = (<select className="form-select cmp-member-select" />) as HTMLSelectElement;
		this.memberElem = (
			<div className="cmp-member hide">
				<label className="cmp-member-label">Member</label>
				{this.memberSelect}
			</div>
		) as HTMLElement;
		this.memberSelect.addEventListener('change', () => {
			this.member = this.memberSelect.value;
			if (!this.pair) return;
			this.drawOverview();
			this.drawTables();
			this.drawResourceCards();
		});

		const simulateRef = ref<HTMLButtonElement>();
		const swapRef = ref<HTMLButtonElement>();
		const emptyTextRef = ref<HTMLParagraphElement>();
		const changedOnlyRef = ref<HTMLButtonElement>();

		this.tables = {
			damage: new CompareTable<ActionMetrics>(damageColumns(), 0),
			threat: new CompareTable<ActionMetrics>(threatColumns(), 0),
			healing: new CompareTable<ActionMetrics>(healingColumns(), 0),
			damageTaken: new CompareTable<ActionMetrics>(damageTakenColumns(), 0),
			buffs: new CompareTable<AuraMetrics>(auraColumns(), 0),
			debuffs: new CompareTable<AuraMetrics>(auraColumns(), 0),
			casts: new CompareTable<ActionMetrics>(castColumns(), 0),
			resources: new Map(),
		};
		const resourcePanels = orderedResourceTypes.map(type => {
			const table = new CompareTable<ResourceMetrics>(resourceColumns(), 0);
			const resourcePanel = panel(resourceNames.get(type) ?? 'Resource', 'hide', table.rootElem) as HTMLElement;
			this.tables.resources.set(type, { panel: resourcePanel, table });
			return resourcePanel;
		});

		const segmented = <M extends string>(modes: Array<[M, string]>, onPick: (mode: M) => void, buttons: HTMLButtonElement[]) => (
			<div className="segmented" attributes={{ role: 'group' }}>
				{modes.map(([mode, label]) => {
					const button = (
						<button className="segmented-option" type="button" dataset={{ mode }}>
							{label}
						</button>
					) as HTMLButtonElement;
					button.addEventListener('click', () => onPick(mode));
					buttons.push(button);
					return button;
				})}
			</div>
		);
		const changesButtons: HTMLButtonElement[] = [];
		const distributionButtons: HTMLButtonElement[] = [];
		const changesElem = (<div className="cmp-changes" />) as HTMLElement;
		const distributionElem = (<div className="cmp-distribution" />) as HTMLElement;

		this.rootElem.appendChild(
			<>
				<div className="dr-toolbar cmp-toolbar">
					<span className="dr-toolbar-title">
						Compare
						<i className="fas fa-chevron-right" />
					</span>
					<ul className="nav dr-tabs" attributes={{ role: 'tablist' }}>
						{tabs.map(({ id, label, icon, className }, i) => (
							<li className={`nav-item ${className ?? ''}`} attributes={{ role: 'presentation' }}>
								<button
									className={`nav-link${i == 0 ? ' active' : ''}`}
									type="button"
									attributes={{
										role: 'tab',
										// @ts-expect-error
										'aria-controls': id,
										'aria-selected': i == 0,
									}}
									dataset={{ bsToggle: 'tab', bsTarget: `#${id}` }}>
									<i className={`fas ${icon}`} />
									{label}
								</button>
							</li>
						))}
					</ul>
					<div className="tabs-filler" />
					<button ref={changedOnlyRef} className="filter-chip cmp-changed-only" type="button">
						<i className="fas fa-filter" />
						Changed only
					</button>
				</div>
				<div className="cmp-runs">
					{this.runPicker('a')}
					<button ref={swapRef} className="cmp-swap" type="button" attributes={{ 'aria-label': 'Swap the runs' }}>
						<i className="fas fa-right-left" />
					</button>
					{this.runPicker('b')}
				</div>
				{this.memberElem}
				<div className="dr-empty cmp-empty hide">
					<i className="fas fa-code-compare dr-empty-icon" />
					<h5 className="dr-empty-title">Nothing to compare yet</h5>
					<p ref={emptyTextRef} className="dr-empty-text" />
					<button ref={simulateRef} className="btn btn-primary" type="button">
						Simulate
					</button>
				</div>
				<div className="tab-content cmp-content">
					<div id="compareOverviewTab" className="tab-pane dr-tab-content fade active show">
						<div className="cmp-verdict" />
						<div className="cmp-cards" />
						<div className="cmp-grid">
							{panel(
								'Biggest Changes',
								'cmp-changes-panel',
								changesElem,
								segmented<ChangesMode>(
									[
										['dps', 'DPS'],
										['casts', 'Casts'],
										['uptime', 'Uptime'],
									],
									mode => this.setChangesMode(mode),
									changesButtons,
								),
							)}
							{panel('Settings Changed', 'cmp-settings-panel', <div className="cmp-settings" />)}
						</div>
						{panel(
							'DPS Distribution',
							'cmp-distribution-panel',
							distributionElem,
							segmented<DistributionMode>(
								[
									['distribution', 'Distribution'],
									['cumulative', 'Cumulative'],
								],
								mode => this.setDistributionMode(mode),
								distributionButtons,
							),
						)}
					</div>
					<div id="compareDamageTab" className="tab-pane dr-tab-content fade">
						{panel('Abilities', '', this.tables.damage.rootElem)}
					</div>
					<div id="compareThreatTab" className="tab-pane dr-tab-content fade">
						{panel('Abilities', '', this.tables.threat.rootElem)}
					</div>
					<div id="compareHealingTab" className="tab-pane dr-tab-content fade">
						{panel('Abilities', '', this.tables.healing.rootElem)}
					</div>
					<div id="compareDamageTakenTab" className="tab-pane dr-tab-content fade">
						{panel('Abilities', '', this.tables.damageTaken.rootElem)}
					</div>
					<div id="compareAurasTab" className="tab-pane dr-tab-content fade">
						<div className="dr-auras-grid cmp-auras-grid">
							{panel('Buffs', '', this.tables.buffs.rootElem)}
							{panel('Debuffs', '', this.tables.debuffs.rootElem)}
						</div>
					</div>
					<div id="compareCastsTab" className="tab-pane dr-tab-content fade">
						{panel('Casts', '', this.tables.casts.rootElem)}
					</div>
					<div id="compareResourcesTab" className="tab-pane dr-tab-content fade">
						<div className="cmp-cards cmp-resource-cards" />
						{resourcePanels}
					</div>
				</div>
			</>,
		);

		this.selects = {
			a: this.rootElem.querySelector('.cmp-run-a select')!,
			b: this.rootElem.querySelector('.cmp-run-b select')!,
		};
		this.metaElems = {
			a: this.rootElem.querySelector('.cmp-run-a .cmp-run-meta')!,
			b: this.rootElem.querySelector('.cmp-run-b .cmp-run-meta')!,
		};
		this.contentElem = this.rootElem.querySelector('.cmp-content')!;
		this.emptyElem = this.rootElem.querySelector('.cmp-empty')!;
		this.emptyTextElem = emptyTextRef.value!;
		this.overviewElems = {
			verdict: this.rootElem.querySelector('.cmp-verdict')!,
			cards: this.rootElem.querySelector('.cmp-cards')!,
			changes: changesElem,
			settings: this.rootElem.querySelector('.cmp-settings')!,
			distribution: distributionElem,
			changesButtons,
			distributionButtons,
		};
		this.updateModeButtons();
		delegate(this.rootElem.querySelector('#compareOverviewTab')!, { target: '[data-tippy-content]' });
		delegate(this.rootElem.querySelector('.cmp-resource-cards')!, { target: '[data-tippy-content]' });
		this.resourceCardsElem = this.rootElem.querySelector('.cmp-resource-cards')!;

		(['a', 'b'] as Side[]).forEach(side =>
			this.selects[side].addEventListener('change', () => {
				this.choices[side] = this.selects[side].value;
				this.update();
			}),
		);
		swapRef.value!.addEventListener('click', () => {
			this.choices = { a: this.choices.b, b: this.choices.a };
			this.update();
		});
		tippy(swapRef.value!, { content: 'Swap the runs' });
		this.bindChangedOnly(changedOnlyRef.value!);
		simulateRef.value!.addEventListener('click', () => {
			this.simUI.simActionsContainer.querySelector<HTMLButtonElement>('.dps-action')?.click();
		});

		const toolbar = this.rootElem.querySelector<HTMLElement>('.dr-toolbar')!;
		trackStuckToolbar(toolbar, simUI);
		syncSubTabWithUrl(toolbar, CompareResults.VIEW_PARAM);

		const updateMetricTabs = () => {
			this.rootElem.classList.toggle('hide-threat-metrics', !simUI.sim.getShowThreatMetrics());
			this.rootElem.classList.toggle('hide-healing-metrics', !simUI.sim.getShowHealingMetrics());
		};
		simUI.sim.showThreatMetricsChangeEmitter.on(updateMetricTabs);
		simUI.sim.showHealingMetricsChangeEmitter.on(updateMetricTabs);
		updateMetricTabs();

		// A tab that opens with runs that came while it was hidden draws them now.
		document.addEventListener('shown.bs.tab', () => {
			if (this.pendingDraw && this.isVisible()) this.update();
		});
		history.changeEmitter.on(() => this.update());
		this.addOnDisposeCallback(() => this.chart?.destroy());

		// A wider or narrower chart can take a different bin width, like the Results tab's DPS
		// Distribution. We check once a frame at most, because a window resize sends many changes.
		let frame = 0;
		const resizeObserver = new ResizeObserver(() => {
			cancelAnimationFrame(frame);
			frame = requestAnimationFrame(() => {
				const width = this.overviewElems.distribution.clientWidth;
				if (this.pair && this.distributionMode === 'distribution' && width && width !== this.distributionWidth) this.drawDistribution();
			});
		});
		resizeObserver.observe(this.overviewElems.distribution);
		this.addOnDisposeCallback(() => resizeObserver.disconnect());
		this.update();
	}

	// The chip that leaves out the rows of the tables where nothing changed. We remember it in
	// the browser, like the DPS Distribution's mode.
	private bindChangedOnly(button: HTMLButtonElement) {
		let changedOnly = false;
		try {
			changedOnly = window.localStorage.getItem(CHANGED_ONLY_KEY) === 'true';
		} catch {
			// We start with every row.
		}
		const apply = () => {
			button.classList.toggle('active', changedOnly);
			this.allTables().forEach(table => table.setChangedOnly(changedOnly));
		};
		button.addEventListener('click', () => {
			changedOnly = !changedOnly;
			try {
				window.localStorage.setItem(CHANGED_ONLY_KEY, String(changedOnly));
			} catch {
				// We only lose the remembered choice.
			}
			apply();
		});
		tippy(button, { content: 'Only show the rows of the tables where something changed between the runs' });
		apply();
	}

	private allTables(): CompareTable<any>[] {
		const { resources, ...tables } = this.tables;
		return [...Object.values(tables), ...Array.from(resources.values(), ({ table }) => table)];
	}

	private runPicker(side: Side): Element {
		return (
			<div className={`cmp-run cmp-run-${side}`}>
				<span className="cmp-run-badge">{side.toUpperCase()}</span>
				<div className="cmp-run-body">
					<span className="cmp-run-label">{SIDE_NAMES[side]}</span>
					<select className="form-select cmp-run-select" />
					<span className="cmp-run-meta" />
				</div>
			</div>
		);
	}

	private isVisible(): boolean {
		return this.rootElem.offsetParent !== null;
	}

	// The run a choice stands for, or null when there's no such run.
	private resolve(choice: RunChoice): SimRun | null {
		const { pinned, recent } = this.history.getRuns();
		if (choice === 'latest') return recent[0] ?? null;
		if (choice === 'previous') return recent[1] ?? null;
		return recent.find(run => run.id === choice) ?? pinned.find(run => run.id === choice) ?? null;
	}

	private update() {
		this.fillSelects();
		const a = this.resolve(this.choices.a);
		const b = this.resolve(this.choices.b);
		(['a', 'b'] as Side[]).forEach(side => this.fillMeta(side, side === 'a' ? a : b));

		const usable = this.history.getRuns().recent.filter(run => this.history.hasResults(run)).length;
		const problem = !a || !b ? this.missingText(usable) : !this.history.hasResults(a) || !this.history.hasResults(b) ? this.noResultsText() : null;
		this.emptyElem.classList.toggle('hide', !problem);
		this.contentElem.classList.toggle('hide', !!problem);
		this.rootElem.querySelector('.dr-toolbar .dr-tabs')!.classList.toggle('invisible', !!problem);
		if (problem) {
			this.emptyTextElem.textContent = problem;
			this.drawnKey = '';
			this.pair = null;
			return;
		}

		const key = `${a!.id}|${b!.id}`;
		if (key === this.drawnKey) return;
		if (!this.isVisible()) {
			this.pendingDraw = true;
			return;
		}
		this.pendingDraw = false;
		this.drawnKey = key;
		this.draw(a!, b!);
	}

	private missingText(usable: number): string {
		if (usable == 0)
			return 'Each run of Simulate goes in the sim history. Run the sim, change something, and run it again to see here what the change did.';
		return `We need one more run. Change something, like ${this.changeExample}, and run the sim again to see here what the change did.`;
	}

	private noResultsText(): string {
		return 'One of the picked runs ran before we kept the full results of each run, so we only know its DPS. Pick another run, or run the sim again.';
	}

	private fillSelects() {
		const { pinned, recent } = this.history.getRuns();
		const option = (value: string, text: string, disabled = false) => {
			const elem = document.createElement('option');
			elem.value = value;
			elem.textContent = text;
			elem.disabled = disabled;
			return elem;
		};
		const runOption = (run: SimRun) => {
			const score = run.leveling ? `${run.leveling.mobsKilled.toFixed(1)} mobs` : `${run.dps.toFixed(1)} DPS`;
			return option(run.id, `${runLabel(run)} · ${score}${this.history.hasResults(run) ? '' : ' (DPS only)'}`, !this.history.hasResults(run));
		};
		const group = (label: string, options: HTMLOptionElement[]) => {
			const elem = document.createElement('optgroup');
			elem.label = label;
			elem.append(...options);
			return elem;
		};

		(['a', 'b'] as Side[]).forEach(side => {
			const select = this.selects[side];
			const groups = [
				group('Follows new runs', [option('latest', 'Latest run'), option('previous', 'Previous run')]),
				...(pinned.length ? [group('Pinned', pinned.map(runOption))] : []),
				...(recent.length ? [group('Recent', recent.map(runOption))] : []),
			];
			select.replaceChildren(...groups);
			// A picked run that's gone from the history goes back to following new runs.
			if (!Array.from(select.options).some(o => o.value === this.choices[side])) this.choices[side] = side === 'a' ? 'previous' : 'latest';
			select.value = this.choices[side];
		});
	}

	private fillMeta(side: Side, run: SimRun | null) {
		const elem = this.metaElems[side];
		if (!run) {
			elem.textContent = 'No run yet';
			return;
		}
		const label = this.choices[side] === 'latest' || this.choices[side] === 'previous' ? `${runLabel(run)} · ` : '';
		// A leveling run goes by its mobs killed a session, with the standard error.
		const score = run.leveling
			? `${run.leveling.mobsKilled.toFixed(2)} mobs ±${mobsKilledError(run.leveling, run.iterations).toFixed(2)}`
			: `${run.dps.toFixed(1)} DPS ±${run.stdev.toFixed(0)}`;
		elem.textContent = `${label}${score} · ${formatToNumber(run.iterations)} iterations`;
	}

	private async loadResult(run: SimRun): Promise<SimResult | null> {
		const cached = this.resultCache.get(run.id);
		if (cached) return cached;
		const result = await this.history.loadResults(run);
		if (result) {
			// The two shown and a few before them are enough. Each holds every metric of a run.
			if (this.resultCache.size >= 6) this.resultCache.delete(this.resultCache.keys().next().value!);
			this.resultCache.set(run.id, result);
		}
		return result;
	}

	private async draw(a: SimRun, b: SimRun) {
		const token = ++this.drawToken;
		this.contentElem.classList.add('loading');
		const [resultA, resultB] = await Promise.all([this.loadResult(a), this.loadResult(b)]);
		if (token !== this.drawToken) return;
		this.contentElem.classList.remove('loading');
		if (!resultA || !resultB) {
			this.emptyTextElem.textContent = this.noResultsText();
			this.emptyElem.classList.remove('hide');
			this.contentElem.classList.add('hide');
			this.drawnKey = '';
			return;
		}

		this.pair = { a, b, resultA, resultB };
		this.fillMembers();
		this.drawOverview();
		this.drawTables();
		this.drawResourceCards();
	}

	// A raid's runs compare the whole raid or one member, picked by name. The members are those of either run,
	// so a member the compared raid dropped (or added) still shows, with nothing on the other side.
	private fillMembers() {
		const { resultA, resultB } = this.pair!;
		const names = [...new Set([...resultA.getRaidIndexedPlayers(), ...resultB.getRaidIndexedPlayers()].map(player => this.memberKey(player.name)))];
		const isRaid = resultA.getRaidIndexedPlayers().length > 1 || resultB.getRaidIndexedPlayers().length > 1;
		this.memberElem.classList.toggle('hide', !isRaid);
		if (!isRaid || !names.includes(this.member)) this.member = '';
		const option = (value: string, text: string) => {
			const elem = document.createElement('option');
			elem.value = value;
			elem.textContent = text;
			return elem;
		};
		this.memberSelect.replaceChildren(option('', 'Whole raid'), ...names.map(name => option(name, name)));
		this.memberSelect.value = this.member;
	}

	// The players we compare in a run: the picked member, or every player.
	private players(result: SimResult): UnitMetrics[] {
		const players = result.getRaidIndexedPlayers();
		return this.member ? players.filter(player => this.memberKey(player.name) === this.member) : players;
	}

	// The DPS we compare in a run: the picked member's, or the raid's.
	private dpsOf(result: SimResult): DistributionMetrics {
		if (!this.member) return result.raidMetrics.dps;
		return this.players(result)[0]?.dps ?? DistributionMetrics.create();
	}

	private drawOverview() {
		const { a, b, resultA, resultB } = this.pair!;
		const { verdict, cards, settings } = this.overviewElems;

		// A picked member's own results, by its raid index in each run.
		const filter = (result: SimResult) => {
			const player = this.member ? this.players(result)[0] : undefined;
			return this.member ? { player: player?.index ?? -1 } : undefined;
		};
		const columnsA = RaidSimResultsManager.toplineColumns(resultA, filter(resultA), { showOutOfMana: true });
		const columnsB = RaidSimResultsManager.toplineColumns(resultB, filter(resultB), { showOutOfMana: true });
		const metrics = columnsB
			.map(metricB => ({ b: metricB, a: columnsA.find(metricA => metricA.name === metricB.name) }))
			.filter((pair): pair is { a: ResultMetric; b: ResultMetric } => !!pair.a && (pair.a.average !== 0 || pair.b.average !== 0));

		verdict.replaceChildren(this.verdict(a, b, resultA, resultB));
		cards.replaceChildren(...metrics.map(({ a: metricA, b: metricB }) => this.card(metricA, metricB, resultA.iterations, resultB.iterations)));

		if (a.id === b.id) {
			settings.replaceChildren(<p className="cmp-note">Both sides are the same run.</p>);
		} else {
			const categories = this.history.changesBetween(a, b);
			settings.replaceChildren(
				categories === null ? (
					<p className="cmp-note">We couldn't read the settings of one of the runs.</p>
				) : categories.length ? (
					<div className="cmp-settings-list">{changesList(categories)}</div>
				) : (
					<p className="cmp-note">Both runs used the same settings. The difference is chance, from the random rolls of each iteration.</p>
				),
			);
		}

		this.drawChanges();
		this.drawDistribution();
	}

	// One sentence on the DPS: how much more or less the compared run does, and whether that's
	// more than chance. Two leveling runs compare their mobs killed a session instead.
	private verdict(a: SimRun, b: SimRun, resultA: SimResult, resultB: SimResult): Element {
		const levelingA = getLevelingMetrics(resultA);
		const levelingB = getLevelingMetrics(resultB);
		const leveling = !!levelingA && !!levelingB;
		const unit = leveling ? 'mobs' : 'DPS';
		const dpsA = leveling ? { avg: levelingA.mobsKilled, stdev: levelingA.mobsKilledStdev } : this.dpsOf(resultA);
		const dpsB = leveling ? { avg: levelingB.mobsKilled, stdev: levelingB.mobsKilledStdev } : this.dpsOf(resultB);
		const digits = leveling ? 2 : 1;
		const diff = dpsB.avg - dpsA.avg;
		const percent = (diff / (dpsA.avg || 1)) * 100;
		const z = zScore(dpsA.avg, dpsA.stdev, resultA.iterations, dpsB.avg, dpsB.stdev, resultB.iterations);
		const significant = z > Z_SIGNIFICANT;

		const tone = a.id === b.id || !diff || !significant ? 'flat' : diff > 0 ? 'better' : 'worse';
		const noise = Math.sqrt(dpsA.stdev ** 2 / Math.max(resultA.iterations, 1) + dpsB.stdev ** 2 / Math.max(resultB.iterations, 1));
		// The runs take the colors of their pickers, and the amount the color of the change. A
		// change within the noise gets no green or red, because it may well be chance.
		const compared = <span className="cmp-word-b">compared run</span>;
		const baseline = <span className="cmp-word-a">baseline</span>;
		const amount = (
			<span className={`cmp-verdict-amount cmp-amount-${tone}`}>
				{`${formatFixed(digits)(Math.abs(diff))} ${unit} (${formatFixed(2)(Math.abs(percent))}%) ${diff > 0 ? 'more' : 'less'}`}
			</span>
		);

		let title: Element;
		if (a.id === b.id) title = <>Both sides are the same run. Pick another run to compare with.</>;
		else if (!diff) title = <>{leveling ? 'Both runs killed the same mobs a session.' : 'Both runs did the same DPS.'}</>;
		else if (leveling)
			title = (
				<>
					The {compared} kills {amount} a session than the {baseline}.
				</>
			);
		else
			title = (
				<>
					The {compared} does {amount} than the {baseline}.
				</>
			);

		return (
			<div className={`cmp-verdict-inner cmp-verdict-${tone}`}>
				<i className={`fas ${tone === 'better' ? 'fa-arrow-trend-up' : tone === 'worse' ? 'fa-arrow-trend-down' : 'fa-equals'} cmp-verdict-icon`} />
				<div className="cmp-verdict-text">
					<span className="cmp-verdict-title">{title}</span>
					{a.id !== b.id && diff ? (
						<span className="cmp-verdict-detail">
							{significant ? (
								<>
									That's more than chance: the difference is <span className="cmp-verdict-number">{`${z.toFixed(1)} times`}</span> its
									standard error (Z test at 95%).
								</>
							) : (
								<>
									That's within the noise of the two runs (
									<span className="cmp-verdict-number">{`±${formatFixed(digits)(noise * Z_SIGNIFICANT)} ${unit}`}</span> at 95%). Run more
									iterations to tell them apart.
								</>
							)}
						</span>
					) : null}
				</div>
				{a.id !== b.id && diff ? (
					<span className={`cmp-verdict-chip ${significant ? 'significant' : ''}`}>{significant ? 'Significant' : 'Within noise'}</span>
				) : null}
			</div>
		);
	}

	private card(metricA: ResultMetric, metricB: ResultMetric, iterationsA: number, iterationsB: number): Element {
		const better: Better = LOWER_IS_BETTER.has(metricB.name) ? 'lower' : 'higher';
		const format = (value: number) => RaidSimResultsManager.formatResultValue(value, metricB.unit, metricB.digits);
		const diff = metricB.average - metricA.average;
		const percent = metricA.average ? (diff / Math.abs(metricA.average)) * 100 : 0;
		// Only results with a spread can be tested, like DPS. TMI and the others without one show
		// the change alone.
		const z = metricA.stdev && metricB.stdev ? zScore(metricA.average, metricA.stdev, iterationsA, metricB.average, metricB.stdev, iterationsB) : null;
		const label = <span className="results-card-label">{metricB.name}</span>;
		if (TOOLTIP_METRIC_LABELS[metricB.name as keyof typeof TOOLTIP_METRIC_LABELS]) {
			tippy(label, { content: TOOLTIP_METRIC_LABELS[metricB.name as keyof typeof TOOLTIP_METRIC_LABELS], ignoreAttributes: true });
		}
		return (
			<div
				className={`cmp-card ${metricB.name === (getLevelingMetrics(this.pair!.resultB) ? 'Mobs Killed' : 'DPS') ? 'cmp-card-lead' : ''} ${
					z !== null && z <= Z_SIGNIFICANT ? 'cmp-card-noise' : ''
				}`}>
				{label}
				<div className="cmp-card-values">
					<span className="cmp-card-b">{format(metricB.average)}</span>
					<span className="cmp-card-a">{`from ${format(metricA.average)}`}</span>
				</div>
				<div className="cmp-card-change">
					{changeElem(metricA.average, metricB.average, format, better)}
					{metricB.unit !== 'percentage' && metricA.average ? (
						<span className="cmp-card-percent">{`${percent >= 0 ? '+' : '-'}${formatFixed(2)(Math.abs(percent))}%`}</span>
					) : null}
					{z !== null && diff ? (
						<span
							className={`cmp-card-sig ${z > Z_SIGNIFICANT ? 'significant' : ''}`}
							dataset={{ tippyContent: `The difference is ${z.toFixed(2)} times its standard error. Above 1.96 it's more than chance.` }}>
							{z > Z_SIGNIFICANT ? 'significant' : 'noise'}
						</span>
					) : null}
				</div>
			</div>
		);
	}

	private setChangesMode(mode: ChangesMode) {
		this.changesMode = mode;
		this.updateModeButtons();
		if (this.pair) this.drawChanges();
	}

	private setDistributionMode(mode: DistributionMode) {
		this.distributionMode = mode;
		this.updateModeButtons();
		if (this.pair) this.drawDistribution();
	}

	private updateModeButtons() {
		this.overviewElems.changesButtons.forEach(button => button.classList.toggle('active', button.dataset.mode === this.changesMode));
		this.overviewElems.distributionButtons.forEach(button => button.classList.toggle('active', button.dataset.mode === this.distributionMode));
	}

	// The spells, casts or auras that changed the most, as bars to the right for more and to
	// the left for less.
	private drawChanges() {
		const { resultA, resultB } = this.pair!;
		let items: ChangeItem[];
		let format: (value: number) => string;
		let better: Better;
		let unit: string;
		if (this.changesMode === 'dps') {
			items = changeItems(pairRows(damageRows(resultA, this.players(resultA)), damageRows(resultB, this.players(resultB))), m => m.dps);
			format = formatFixed(1);
			better = 'higher';
			unit = 'DPS';
		} else if (this.changesMode === 'casts') {
			items = changeItems(pairRows(castRows(resultA, this.players(resultA)), castRows(resultB, this.players(resultB))), m => m.casts);
			format = formatFixed(1);
			better = 'none';
			unit = 'casts';
		} else {
			items = changeItems(
				[...pairRows(buffRows(resultA, this.players(resultA)), buffRows(resultB, this.players(resultB)), true), ...pairRows(debuffRows(resultA, this.players(resultA)), debuffRows(resultB, this.players(resultB)), true)],
				m => m.uptimePercent,
			);
			format = formatPercent(1);
			better = 'none';
			unit = 'uptime';
		}
		this.overviewElems.changes.replaceChildren(changesChart(items, format, better, unit));
	}

	private drawDistribution() {
		const { resultA, resultB } = this.pair!;
		this.chart?.destroy();
		this.chart = null;
		const elem = this.overviewElems.distribution;
		this.distributionWidth = elem.clientWidth;
		const countsA = dpsCounts(this.dpsOf(resultA));
		const countsB = dpsCounts(this.dpsOf(resultB));
		if (!countsA.length || !countsB.length) {
			elem.replaceChildren(<p className="cmp-note">No DPS distribution in one of the runs.</p>);
			return;
		}
		const medianA = percentile(countsA, 0.5);
		const medianB = percentile(countsB, 0.5);
		const canvas = document.createElement('canvas');
		elem.replaceChildren(
			<div className="cmp-legend">
				<span className="cmp-legend-item cmp-legend-a">
					<span className="cmp-legend-swatch" />
					{`Baseline · median ${formatFixed(1)(medianA)}`}
				</span>
				<span className="cmp-legend-item cmp-legend-b">
					<span className="cmp-legend-swatch" />
					{`Compared · median ${formatFixed(1)(medianB)}`}
				</span>
			</div>,
			<div className="cmp-distribution-chart">{canvas}</div>,
		);
		const markers = [
			{ value: medianA, color: COLOR_A },
			{ value: medianB, color: COLOR_B },
		];
		this.chart =
			this.distributionMode === 'distribution'
				? drawDistribution(canvas, countsA, countsB, markers, elem.clientWidth)
				: drawCumulative(canvas, countsA, countsB, markers);
	}

	// How much of each resource the runs gained, spent and wasted, from all their iterations. The
	// rates are per second, so runs of different lengths compare.
	private drawResourceCards() {
		const { resultA, resultB } = this.pair!;
		const cards: Element[] = [];
		RESOURCE_SUMMARY_TYPES.forEach(type => {
			const rowsA = resourceRows(this.players(resultA), type).map(row => row.metric);
			const rowsB = resourceRows(this.players(resultB), type).map(row => row.metric);
			if (!rowsA.length && !rowsB.length) return;
			const name = resourceNames.get(type) ?? 'Resource';
			const sum = (rows: ResourceMetrics[], value: (metric: ResourceMetrics) => number) => rows.reduce((total, metric) => total + value(metric), 0);
			const gained = (rows: ResourceMetrics[]) => sum(rows, m => Math.max(0, m.gainPerSecond));
			const spent = (rows: ResourceMetrics[]) => sum(rows, m => Math.max(0, -m.gainPerSecond));
			const wasted = (rows: ResourceMetrics[]) => sum(rows, m => m.wastedGain);
			cards.push(
				compareCard(`${name} gained / s`, gained(rowsA), gained(rowsB), formatFixed(2), 'higher'),
				compareCard(`${name} spent / s`, spent(rowsA), spent(rowsB), formatFixed(2), 'none'),
				compareCard(
					`${name} wasted`,
					wasted(rowsA),
					wasted(rowsB),
					formatFixed(1),
					'lower',
					'Gains that came while the pool was full, in an iteration.',
				),
			);
			if (type === ResourceType.ResourceTypeMana) {
				// A raid's average over the players we compare.
				const oom = (result: SimResult) => {
					const players = this.players(result);
					return players.reduce((sum, player) => sum + player.secondsOomAvg, 0) / Math.max(players.length, 1);
				};
				cards.push(
					compareCard(
						'Time out of mana',
						oom(resultA),
						oom(resultB),
						value => `${formatFixed(1)(value)}s`,
						'lower',
						'How long an iteration spends out of mana, on average.',
					),
				);
			}
		});
		this.resourceCardsElem.replaceChildren(...cards);
		this.resourceCardsElem.classList.toggle('hide', !cards.length);
	}

	private drawTables() {
		const { resultA, resultB } = this.pair!;
		this.tables.damage.setRows(pairRows(damageRows(resultA, this.players(resultA)), damageRows(resultB, this.players(resultB))));
		this.tables.threat.setRows(pairRows(threatRows(resultA, this.players(resultA)), threatRows(resultB, this.players(resultB))));
		this.tables.healing.setRows(pairRows(healingRows(resultA, this.players(resultA)), healingRows(resultB, this.players(resultB))));
		this.tables.damageTaken.setRows(pairRows(damageTakenRows(resultA, this.players(resultA)), damageTakenRows(resultB, this.players(resultB))));
		this.tables.buffs.setRows(pairRows(buffRows(resultA, this.players(resultA)), buffRows(resultB, this.players(resultB)), true));
		this.tables.debuffs.setRows(pairRows(debuffRows(resultA, this.players(resultA)), debuffRows(resultB, this.players(resultB)), true));
		this.tables.casts.setRows(pairRows(castRows(resultA, this.players(resultA)), castRows(resultB, this.players(resultB))));
		this.tables.resources.forEach(({ panel, table }, type) => {
			table.setRows(pairRows(resourceRows(this.players(resultA), type), resourceRows(this.players(resultB), type)));
			panel.classList.toggle('hide', table.isEmpty);
		});
	}
}

// The resources with a summary on the Resources tab.
const RESOURCE_SUMMARY_TYPES = [ResourceType.ResourceTypeMana, ResourceType.ResourceTypeEnergy, ResourceType.ResourceTypeRage];

// A card with the compared run's value, the baseline's under it, and the change.
function compareCard(name: string, a: number, b: number, format: (value: number) => string, better: Better, tooltip?: string): Element {
	const percent = a ? ((b - a) / Math.abs(a)) * 100 : 0;
	return (
		<div className="cmp-card" dataset={tooltip ? { tippyContent: tooltip } : {}}>
			<span className="results-card-label">{name}</span>
			<div className="cmp-card-values">
				<span className="cmp-card-b">{format(b)}</span>
				<span className="cmp-card-a">{`from ${format(a)}`}</span>
			</div>
			<div className="cmp-card-change">
				{changeElem(a, b, format, better)}
				{a && format(Math.abs(b - a)) !== format(0) ? (
					<span className="cmp-card-percent">{`${percent >= 0 ? '+' : '-'}${formatFixed(2)(Math.abs(percent))}%`}</span>
				) : null}
			</div>
		</div>
	);
}

// A spell, aura or pet of one run, by the key we match it with in the other run.
interface Keyed<T> {
	key: string;
	actionId: ActionId;
	name: string;
	metric: T;
}

// Matches the rows of the two runs by their keys. The order doesn't matter, the tables sort.
function pairRows<T>(rowsA: Keyed<T>[], rowsB: Keyed<T>[], isAura = false): CompareRow<T>[] {
	const rows = new Map<string, CompareRow<T>>();
	rowsA.forEach(row => rows.set(row.key, { key: row.key, actionId: row.actionId, name: row.name, isAura, a: row.metric }));
	rowsB.forEach(row => {
		const existing = rows.get(row.key);
		if (existing) existing.b = row.metric;
		else rows.set(row.key, { key: row.key, actionId: row.actionId, name: row.name, isAura, b: row.metric });
	});
	return Array.from(rows.values());
}

// The players' spells, with the ranks and hands of one spell in one row, like the Results
// tab's rows before we open them. Each pet is one row. Several players (a whole raid) add up
// their casts of a spell in its row.
function actionRows(players: UnitMetrics[], pick: (unit: UnitMetrics) => ActionMetrics[]): Keyed<ActionMetrics>[] {
	if (!players.length) return [];
	const own = ActionMetrics.groupById(players.flatMap(pick)).map(group => {
		const metric = ActionMetrics.merge(group, { removeTag: true });
		return { key: metric.actionId.toString(), actionId: metric.actionId, name: metric.name, metric };
	});
	const pets = Object.values(bucket(players.flatMap(player => player.pets), pet => pet.name)).flatMap(pets => {
		const actions = pets.flatMap(pick);
		if (!actions.length) return [];
		const metric = ActionMetrics.merge(actions, { removeTag: true, actionIdOverride: pets[0].petActionId || undefined });
		return [{ key: `pet-${pets[0].name}`, actionId: metric.actionId, name: pets[0].name, metric }];
	});
	return own.concat(pets);
}

const damageRows = (_result: SimResult, players: UnitMetrics[]) => actionRows(players, unit => unit.getDamageActions());
const threatRows = (_result: SimResult, players: UnitMetrics[]) => actionRows(players, unit => unit.getThreatActions());
const healingRows = (_result: SimResult, players: UnitMetrics[]) =>
	actionRows(players, unit => unit.getHealingActions().filter(action => action.hps !== 0));
const castRows = (_result: SimResult, players: UnitMetrics[]) => actionRows(players, unit => unit.actions.filter(action => action.casts !== 0));

// The enemy's spells and melee against the players.
function damageTakenRows(result: SimResult, players: UnitMetrics[]): Keyed<ActionMetrics>[] {
	const actions = result.getTargets().flatMap(target =>
		target.getDamageActions().flatMap(action =>
			players.flatMap(player => {
				try {
					return [action.forTarget({ player: player.unitIndex })];
				} catch {
					return [];
				}
			}),
		),
	);
	return ActionMetrics.groupById(actions).map(group => {
		const metric = ActionMetrics.merge(group, { removeTag: true });
		return { key: metric.actionId.toString(), actionId: metric.actionId, name: metric.name, metric };
	});
}

const isUsedAura = (aura: AuraMetrics) => aura.uptimePercent > 0 || aura.averageProcs > 0;

function auraRows(auras: AuraMetrics[]): Keyed<AuraMetrics>[] {
	return AuraMetrics.groupById(auras)
		.map(group => AuraMetrics.merge(group, { removeTag: true }))
		.filter(isUsedAura)
		.map(metric => ({ key: metric.actionId.toString(), actionId: metric.actionId, name: metric.name, metric }));
}

// A whole raid's buff shows its best uptime on any player, see AuraMetrics.merge.
function buffRows(_result: SimResult, players: UnitMetrics[]): Keyed<AuraMetrics>[] {
	return auraRows(players.flatMap(player => player.auras));
}

function debuffRows(result: SimResult, _players: UnitMetrics[]): Keyed<AuraMetrics>[] {
	return auraRows(result.getDebuffMetrics());
}

function resourceRows(players: UnitMetrics[], type: ResourceType): Keyed<ResourceMetrics>[] {
	return ResourceMetrics.groupById(players.flatMap(player => player.getResourceMetrics(type))).map(group => {
		const metric = ResourceMetrics.merge(group, { removeTag: true });
		return { key: metric.actionId.toString(), actionId: metric.actionId, name: metric.name, metric };
	});
}

// The hit types of a spell's damage, in the order of the Results tab's hover tables. The good
// ones are better when there's more of them, from the side of whoever deals the damage.
const HIT_TYPES: Array<{ name: string; key: keyof ActionMetrics['damageDone']; better: Better }> = [
	{ name: 'Hit', key: 'hit', better: 'none' },
	{ name: 'Resisted Hit', key: 'resistedHit', better: 'lower' },
	{ name: 'Critical Hit', key: 'critHit', better: 'higher' },
	{ name: 'Resisted Critical Hit', key: 'resistedCritHit', better: 'lower' },
	{ name: 'Tick', key: 'tick', better: 'none' },
	{ name: 'Resisted Tick', key: 'resistedTick', better: 'lower' },
	{ name: 'Critical Tick', key: 'critTick', better: 'higher' },
	{ name: 'Resisted Critical Tick', key: 'resistedCritTick', better: 'lower' },
	{ name: 'Glancing Blow', key: 'glance', better: 'lower' },
	{ name: 'Blocked Hit', key: 'block', better: 'lower' },
	{ name: 'Blocked Critical Hit', key: 'blockedCrit', better: 'lower' },
];

// For damage taken the good side flips: the player wants fewer crits and more resists.
const flip = (better: Better): Better => (better === 'higher' ? 'lower' : better === 'lower' ? 'higher' : 'none');
const forTaker = (breakdown: Breakdown<ActionMetrics>): Breakdown<ActionMetrics> => ({
	...breakdown,
	lines: metric => breakdown.lines(metric).map(line => ({ ...line, better: flip(line.better ?? 'none') })),
});

// The DPS of each hit type. The Results tab shows the damage of an iteration, and we turn it into
// DPS to match the column.
const dpsByHitType: Breakdown<ActionMetrics> = {
	format: formatFixed(2),
	lines: m => {
		const perSecond = m.avgDamage ? m.dps / m.avgDamage : 0;
		return HIT_TYPES.map(({ name, key, better }) => ({ name, value: m.damageDone[key].value * perSecond, better }));
	},
};

const averageByHitType: Breakdown<ActionMetrics> = {
	format: formatFixed(1),
	lines: m => HIT_TYPES.map(({ name, key }) => ({ name, value: m.damageDone[key].average, better: 'higher' as Better })),
};

// The share of each hit type in the landed hits, and of each tick type in the landed ticks.
const hitShares: Breakdown<ActionMetrics> = {
	format: formatPercent(2),
	lines: m => {
		const ofHits = (count: number) => (m.landedHits ? (count / m.landedHits) * 100 : 0);
		const ofTicks = (count: number) => (m.landedTicks ? (count / m.landedTicks) * 100 : 0);
		const lines: BreakdownLine[] = [
			{ name: 'Hit', value: ofHits(m.hits - m.resistedHits) },
			{ name: 'Resisted Hit', value: ofHits(m.resistedHits), better: 'lower' },
			{ name: 'Critical Hit', value: ofHits(m.crits - m.resistedCrits), better: 'higher' },
			{ name: 'Blocked Critical Hit', value: ofHits(m.blockedCrits), better: 'lower' },
			{ name: 'Resisted Critical Hit', value: ofHits(m.resistedCrits), better: 'lower' },
			{ name: 'Glancing Blow', value: ofHits(m.glances), better: 'lower' },
			{ name: 'Blocked Hit', value: ofHits(m.blocks), better: 'lower' },
			{ name: 'Tick', value: ofTicks(m.ticks - m.resistedTicks) },
			{ name: 'Resisted Tick', value: ofTicks(m.resistedTicks), better: 'lower' },
			{ name: 'Critical Tick', value: ofTicks(m.critTicks - m.resistedCritTicks), better: 'higher' },
			{ name: 'Resisted Critical Tick', value: ofTicks(m.resistedCritTicks), better: 'lower' },
		];
		return lines;
	},
};

const castOutcomes: Breakdown<ActionMetrics> = {
	format: formatFixed(1),
	lines: m => [
		{ name: 'Hits', value: m.landedHits || m.casts - m.totalMisses },
		{ name: 'Miss', value: m.misses, better: 'lower' },
		{ name: 'Parry', value: m.parries, better: 'lower' },
		{ name: 'Dodge', value: m.dodges, better: 'lower' },
	],
};

const missTypes: Breakdown<ActionMetrics> = {
	format: formatPercent(2),
	lines: m => [
		{ name: 'Miss', value: m.missPercent, better: 'lower' },
		{ name: 'Parry', value: m.parryPercent, better: 'lower' },
		{ name: 'Dodge', value: m.dodgePercent, better: 'lower' },
	],
};

const damageColumns = (): CompareColumn<ActionMetrics>[] => [
	{ name: 'DPS', value: m => m.dps, format: formatFixed(2), better: 'higher', bars: true, breakdown: dpsByHitType },
	{ name: 'Casts', value: m => m.casts, format: formatFixed(1), breakdown: castOutcomes },
	{ name: 'Hits', value: m => m.landedHits || m.landedTicks, format: formatFixed(1), breakdown: hitShares },
	{ name: 'Avg Hit', value: m => m.avgHit || m.avgTick, format: formatFixed(1), better: 'higher', breakdown: averageByHitType },
	{ name: 'Crit %', value: m => m.critPercent + m.blockedCritPercent || m.critTickPercent, format: formatPercent(2), better: 'higher' },
	{ name: 'Miss %', value: m => m.totalMissesPercent, format: formatPercent(2), better: 'lower', breakdown: missTypes },
];

const threatColumns = (): CompareColumn<ActionMetrics>[] => [
	{ name: 'TPS', value: m => m.tps, format: formatFixed(2), better: 'higher', bars: true },
	{ name: 'Casts', value: m => m.casts, format: formatFixed(1) },
	{ name: 'Threat / Cast', value: m => m.avgCastThreat, format: formatFixed(1), better: 'higher' },
];

const healingColumns = (): CompareColumn<ActionMetrics>[] => [
	{ name: 'HPS', value: m => m.hps, format: formatFixed(2), better: 'higher', bars: true },
	{ name: 'Casts', value: m => m.casts, format: formatFixed(1) },
	{ name: 'Avg Cast', value: m => m.avgCastHealing, format: formatFixed(1), better: 'higher' },
	{ name: 'Crit %', value: m => m.healingCritPercent, format: formatPercent(2), better: 'higher' },
];

const damageTakenColumns = (): CompareColumn<ActionMetrics>[] => [
	{ name: 'DTPS', value: m => m.dps, format: formatFixed(2), better: 'lower', bars: true, breakdown: forTaker(dpsByHitType) },
	{ name: 'Hits', value: m => m.landedHits || m.landedTicks, format: formatFixed(1), better: 'lower', breakdown: forTaker(hitShares) },
	{ name: 'Avg Hit', value: m => m.avgHit || m.avgTick, format: formatFixed(1), better: 'lower', breakdown: forTaker(averageByHitType) },
	{ name: 'Crit %', value: m => m.critPercent, format: formatPercent(2), better: 'lower' },
	{
		name: 'Avoided %',
		tooltip: 'Misses, dodges and parries.',
		value: m => m.totalMissesPercent,
		format: formatPercent(2),
		better: 'higher',
		breakdown: forTaker(missTypes),
	},
];

const auraColumns = (): CompareColumn<AuraMetrics>[] => [
	{ name: 'Uptime', value: m => m.uptimePercent, format: formatPercent(2), better: 'higher', bars: true },
	{ name: 'Procs', value: m => m.averageProcs, format: formatFixed(2) },
	{ name: 'PPM', value: m => m.ppm, format: formatFixed(2) },
];

const castColumns = (): CompareColumn<ActionMetrics>[] => [
	{ name: 'Casts', value: m => m.casts, format: formatFixed(1), bars: true },
	{ name: 'CPM', value: m => m.castsPerMinute, format: formatFixed(2) },
];

const resourceColumns = (): CompareColumn<ResourceMetrics>[] => [
	{ name: 'Gain', value: m => m.gain, format: formatFixed(1), better: 'higher', bars: true, total: true },
	{ name: 'Gain / s', value: m => m.gainPerSecond, format: formatFixed(2), better: 'higher' },
	{ name: 'Wasted', value: m => m.wastedGain, format: formatFixed(1), better: 'lower', total: true },
	{ name: 'Casts', value: m => m.events, format: formatFixed(1) },
];

// How far apart two averages are, in standard errors of their difference.
function zScore(avgA: number, stdevA: number, iterationsA: number, avgB: number, stdevB: number, iterationsB: number): number {
	const error = Math.sqrt(stdevA ** 2 / Math.max(iterationsA, 1) + stdevB ** 2 / Math.max(iterationsB, 1));
	if (!error) return avgA === avgB ? 0 : Infinity;
	return Math.abs(avgB - avgA) / error;
}

interface ChangeItem {
	name: string;
	actionId?: ActionId;
	isAura?: boolean;
	a: number;
	b: number;
	diff: number;
}

// The rows with the biggest changes first. The rest add up in one Other row, so the bars still
// add up to the whole change.
function changeItems<T>(rows: CompareRow<T>[], value: (metric: T) => number): ChangeItem[] {
	const items = rows
		.map(row => {
			const a = row.a ? value(row.a) : 0;
			const b = row.b ? value(row.b) : 0;
			return { name: row.name, actionId: row.actionId, isAura: row.isAura, a, b, diff: b - a };
		})
		.filter(item => Math.abs(item.diff) > 1e-6)
		.sort((x, y) => Math.abs(y.diff) - Math.abs(x.diff));
	if (items.length <= TOP_CHANGES + 1) return items;
	const rest = items.slice(TOP_CHANGES);
	return [
		...items.slice(0, TOP_CHANGES),
		{
			name: `${rest.length} others`,
			a: rest.reduce((sum, item) => sum + item.a, 0),
			b: rest.reduce((sum, item) => sum + item.b, 0),
			diff: rest.reduce((sum, item) => sum + item.diff, 0),
		},
	];
}

// A bar for each change, from a line in the middle: to the right for more, to the left for less.
function changesChart(items: ChangeItem[], format: (value: number) => string, better: Better, unit: string): Element {
	if (!items.length) return <p className="cmp-note">{`No ${unit} changed between the runs.`}</p>;
	const max = Math.max(...items.map(item => Math.abs(item.diff)));
	return (
		<div className="cmp-diverging">
			{items.map(item => {
				const icon = (<span className="metrics-action-icon cmp-diverging-icon" />) as HTMLElement;
				if (item.actionId) {
					item.actionId.setBackgroundAndHref(icon as HTMLAnchorElement);
				} else {
					icon.classList.add('cmp-diverging-icon-other');
				}
				const width = `${(Math.abs(item.diff) / max) * 50}%`;
				const tone = better === 'none' ? (item.diff > 0 ? 'more' : 'less') : item.diff > 0 === (better === 'higher') ? 'better' : 'worse';
				return (
					<div className={`cmp-diverging-row cmp-${tone}`} dataset={{ tippyContent: `${item.name}: ${format(item.a)} → ${format(item.b)} ${unit}` }}>
						{icon}
						<span className="cmp-diverging-name text-truncate">{item.name}</span>
						<span className="cmp-diverging-track">
							<span className="cmp-diverging-bar" style={item.diff > 0 ? { left: '50%', width } : { right: '50%', width }} />
						</span>
						<span className="cmp-diverging-value">
							{format(Math.abs(item.diff)) === format(0) ? '±0' : `${item.diff > 0 ? '+' : '-'}${format(Math.abs(item.diff))}`}
						</span>
					</div>
				);
			})}
		</div>
	);
}

interface DpsCount {
	dps: number;
	count: number;
}

function dpsCounts(dps: DistributionMetrics): DpsCount[] {
	return Object.entries(dps.hist)
		.map(([dps, count]) => ({ dps: Number(dps), count }))
		.filter(({ count }) => count > 0)
		.sort((x, y) => x.dps - y.dps);
}

// The DPS at a share of the iterations, spreading each whole DPS step over the half DPS on each
// side, like the DPS Distribution of the Results tab.
function percentile(counts: DpsCount[], share: number): number {
	const total = counts.reduce((sum, { count }) => sum + count, 0);
	const target = share * total;
	let below = 0;
	for (const { dps, count } of counts) {
		if (below + count >= target) return dps - 0.5 + (target - below) / count;
		below += count;
	}
	return counts[counts.length - 1].dps + 0.5;
}

interface Marker {
	value: number;
	color: string;
}

// Each run as the share of its iterations in each bin, so runs with different iteration counts
// compare. Both use the same bins, wide enough for the run with fewer iterations.
function drawDistribution(canvas: HTMLCanvasElement, countsA: DpsCount[], countsB: DpsCount[], markers: Marker[], chartWidth: number): Chart {
	const all = [...countsA, ...countsB];
	const minDps = Math.min(...all.map(c => c.dps));
	const maxDps = Math.max(...all.map(c => c.dps));
	const totalA = countsA.reduce((sum, c) => sum + c.count, 0);
	const totalB = countsB.reduce((sum, c) => sum + c.count, 0);
	const maxBins = Math.max(1, Math.min(chartWidth ? Math.floor(chartWidth / MIN_BIN_PX) : 60, Math.floor(Math.min(totalA, totalB) / MIN_ITERATIONS_PER_BIN)));
	const span = maxDps - minDps + 1;
	const width = NICE_BIN_WIDTHS.find(w => span / w <= maxBins) ?? NICE_BIN_WIDTHS[NICE_BIN_WIDTHS.length - 1];
	const binOf = (dps: number) => Math.floor((dps + 0.5) / width);
	const firstBin = binOf(minDps);
	const binCount = binOf(maxDps) - firstBin + 1;
	const shares = (counts: DpsCount[], total: number) => {
		const bins = new Array(binCount).fill(0);
		counts.forEach(({ dps, count }) => (bins[binOf(dps) - firstBin] += count));
		return bins.map((count, i) => ({ x: (firstBin + i + 0.5) * width - 0.5, y: (count / total) * 100 }));
	};
	// The two bars of a bin take its full width, one over the other. They're see-through, so
	// we see both where they overlap.
	const dataset = (label: string, data: { x: number; y: number }[], color: string) => ({
		label,
		data,
		backgroundColor: color,
		hoverBackgroundColor: color,
		borderRadius: 2,
		categoryPercentage: 0.9,
		barPercentage: 1,
		grouped: false,
	});

	return new Chart(canvas, {
		type: 'bar',
		data: {
			datasets: [dataset('Baseline', shares(countsA, totalA), COLOR_A_BAR), dataset('Compared', shares(countsB, totalB), COLOR_B_BAR)],
		},
		plugins: [markerPlugin(markers)],
		options: {
			maintainAspectRatio: false,
			animation: false,
			interaction: { mode: 'index', intersect: false },
			plugins: {
				legend: { display: false },
				tooltip: {
					callbacks: {
						title: items => {
							const center = items[0].parsed.x + 0.5;
							return width == 1 ? `${center} DPS` : `${center - width / 2} to ${center + width / 2 - 1} DPS`;
						},
						label: item => `${item.dataset.label}: ${formatToNumber(item.parsed.y ?? 0, { maximumFractionDigits: 1 })}% of iterations`,
					},
				},
			},
			scales: {
				x: {
					type: 'linear',
					min: firstBin * width - 0.5,
					max: (firstBin + binCount) * width - 0.5,
					offset: false,
					grid: { display: false },
					ticks: { color: MUTED, maxRotation: 0, autoSkipPadding: 16, callback: wholeDpsTick },
				},
				y: {
					beginAtZero: true,
					grid: { color: 'rgba(255, 255, 255, 0.05)' },
					border: { display: false },
					ticks: { color: MUTED, callback: value => `${value}%` },
				},
			},
		},
	});
}

// The share of each run's iterations at or below each DPS. The curve further right is the
// better run, and where the curves cross the runs trade places.
function drawCumulative(canvas: HTMLCanvasElement, countsA: DpsCount[], countsB: DpsCount[], markers: Marker[]): Chart {
	const points = (counts: DpsCount[]) => {
		const total = counts.reduce((sum, c) => sum + c.count, 0);
		let below = 0;
		const result = [{ x: counts[0].dps - 0.5, y: 0 }];
		counts.forEach(({ dps, count }) => {
			below += count;
			result.push({ x: dps + 0.5, y: (below / total) * 100 });
		});
		return result;
	};
	const pointsA = points(countsA);
	const pointsB = points(countsB);
	const dataset = (label: string, data: { x: number; y: number }[], color: string) => ({
		label,
		data,
		borderColor: color,
		backgroundColor: color,
		borderWidth: 2,
		pointRadius: 0,
		pointHoverRadius: 4,
		tension: 0,
	});
	return new Chart(canvas, {
		type: 'line',
		data: { datasets: [dataset('Baseline', pointsA, COLOR_A), dataset('Compared', pointsB, COLOR_B)] },
		plugins: [markerPlugin(markers)],
		options: {
			maintainAspectRatio: false,
			animation: false,
			interaction: { mode: 'nearest', axis: 'x', intersect: false },
			plugins: {
				legend: { display: false },
				tooltip: {
					callbacks: {
						title: items => `${formatToNumber(items[0].parsed.x, { maximumFractionDigits: 1 })} DPS`,
						label: item => `${item.dataset.label}: ${formatToNumber(item.parsed.y ?? 0, { maximumFractionDigits: 1 })}% at or below`,
					},
				},
			},
			scales: {
				x: {
					type: 'linear',
					min: Math.min(pointsA[0].x, pointsB[0].x),
					max: Math.max(pointsA[pointsA.length - 1].x, pointsB[pointsB.length - 1].x),
					grid: { display: false },
					ticks: { color: MUTED, maxRotation: 0, autoSkipPadding: 16, callback: wholeDpsTick },
				},
				y: {
					min: 0,
					max: 100,
					grid: { color: 'rgba(255, 255, 255, 0.05)' },
					border: { display: false },
					ticks: { color: MUTED, stepSize: 25, callback: value => `${value}%` },
				},
			},
		},
	});
}

function wholeDpsTick(value: string | number): string {
	return Number.isInteger(Number(value)) ? formatToNumber(Number(value)) : '';
}

// A dashed line at each run's median.
function markerPlugin(markers: Marker[]): Plugin {
	return {
		id: 'compareMarkers',
		afterDatasetsDraw: chart => {
			const { ctx, chartArea, scales } = chart;
			ctx.save();
			markers.forEach(marker => {
				const x = scales.x.getPixelForValue(marker.value);
				if (x < chartArea.left || x > chartArea.right) return;
				ctx.strokeStyle = marker.color;
				ctx.lineWidth = 1;
				ctx.setLineDash([4, 3]);
				ctx.beginPath();
				ctx.moveTo(x, chartArea.top);
				ctx.lineTo(x, chartArea.bottom);
				ctx.stroke();
			});
			ctx.restore();
		},
	};
}
