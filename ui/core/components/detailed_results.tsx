import { Tab as BootstrapTab } from 'bootstrap';
import { ref } from 'tsx-vanilla';

import { DetailedResultsUpdate, SimRun, SimRunData } from '../proto/ui';
import { SimResult } from '../proto_utils/sim_result';
import { SimUI } from '../sim_ui';
import { TypedEvent } from '../typed_event';
import { Component } from './component';
import { AuraMetricsTable } from './detailed_results/aura_metrics';
import { CastMetricsTable } from './detailed_results/cast_metrics';
import { DamageMetricsTable } from './detailed_results/damage_metrics';
import { DpsHistogram } from './detailed_results/dps_histogram';
import { DtpsMetricsTable } from './detailed_results/dtps_metrics';
import { EventsView } from './detailed_results/events_view';
import { HealingMetricsTable } from './detailed_results/healing_metrics';
import { PlayerChooser } from './detailed_results/player_chooser';
import { PlayerDamageMetricsTable } from './detailed_results/player_damage';
import { PlayerDamageTakenMetricsTable } from './detailed_results/player_damage_taken';
import { ResourceMetricsTable } from './detailed_results/resource_metrics';
import { SimResultData } from './detailed_results/result_component';
import { ResultsFilter } from './detailed_results/results_filter';
import { ThreatMetricsTable } from './detailed_results/threat_metrics';
import { ToplineResults } from './detailed_results/topline_results';
import { RaidSimResultsManager } from './raid_sim_action';

type Tab = {
	isActive?: boolean;
	targetId: string;
	label: string;
	icon: string;
	classes?: string[];
};

const tabs: Tab[] = [
	{
		isActive: true,
		targetId: 'damageTab',
		label: 'Damage',
		icon: 'fa-bolt',
		classes: ['damage-metrics-tab'],
	},
	{
		// Every spec makes threat, so this tab is always there. The tank results (Damage Taken, DTPS,
		// TMI, chance of death) stay behind the Threat/Tank setting.
		targetId: 'threatTab',
		label: 'Threat',
		icon: 'fa-skull',
	},
	{
		targetId: 'healingTab',
		label: 'Healing',
		icon: 'fa-heart-pulse',
		classes: ['healing-metrics-tab'],
	},
	{
		targetId: 'damageTakenTab',
		label: 'Damage Taken',
		icon: 'fa-shield-halved',
		classes: ['threat-metrics-tab'],
	},
	{
		targetId: 'aurasTab',
		label: 'Auras',
		icon: 'fa-wand-magic-sparkles',
	},
	{
		targetId: 'castsTab',
		label: 'Casts',
		icon: 'fa-hand-sparkles',
	},
	{
		targetId: 'resourcesTab',
		label: 'Resources',
		icon: 'fa-droplet',
	},
	{
		targetId: 'eventsTab',
		label: 'Events',
		icon: 'fa-bars-staggered',
	},
];

// A titled card around one part of the results, like the Settings tab's sections.
export const panel = (title: string, className: string, body: Element, headerExtra?: Element) => (
	<div className={`content-block dr-panel ${className}`}>
		<div className="content-block-header">
			<h6 className="content-block-title">{title}</h6>
			{headerExtra ?? null}
		</div>
		<div className="content-block-body">{body}</div>
	</div>
);

export abstract class DetailedResults extends Component {
	static readonly VIEW_PARAM = 'view';

	protected readonly simUI: SimUI | null;
	protected latestRun: SimRunData | null = null;

	private currentSimResult: SimResult | null = null;
	private resultsEmitter: TypedEvent<SimResultData | null> = new TypedEvent<SimResultData | null>();
	private resultsFilter: ResultsFilter;
	private rootDiv: HTMLElement;

	constructor(parent: HTMLElement, simUI: SimUI | null) {
		super(parent, 'detailed-results-manager-root');

		const simulateRef = ref<HTMLButtonElement>();
		const buffsUnusedRef = ref<HTMLButtonElement>();
		const debuffsUnusedRef = ref<HTMLButtonElement>();

		this.rootElem.appendChild(
			<div className="dr-root dr-no-results">
				<div className="dr-toolbar">
					<span className="dr-toolbar-title">
						Results
						<i className="fas fa-chevron-right" />
					</span>
					<ul className="nav dr-tabs" attributes={{ role: 'tablist' }}>
						{tabs.map(({ label, icon, targetId, isActive, classes }) => (
							<li className={`nav-item dr-tab-tab ${classes?.join(' ') || ''}`} attributes={{ role: 'presentation' }}>
								<button
									className={`nav-link${isActive ? ' active' : ''}`}
									type="button"
									attributes={{
										role: 'tab',
										// @ts-expect-error
										'aria-controls': targetId,
										'aria-selected': !!isActive,
									}}
									dataset={{
										bsToggle: 'tab',
										bsTarget: `#${targetId}`,
									}}>
									<i className={`fas ${icon}`} />
									{label}
								</button>
							</li>
						))}
					</ul>
					<div className="tabs-filler"></div>
					<div className="results-filter"></div>
				</div>
				<div className="tab-content">
					<div id="noResultsTab" className="tab-pane dr-tab-content fade active show">
						<div className="dr-empty">
							<i className="fas fa-chart-column dr-empty-icon" />
							<h5 className="dr-empty-title">No results yet</h5>
							<p className="dr-empty-text">
								Run a simulation to see where the damage comes from, how auras and casts add up, and the typical, lucky and unlucky iterations
								event by event.
							</p>
							{simUI ? (
								<button ref={simulateRef} className="btn btn-primary" type="button">
									Simulate
								</button>
							) : null}
						</div>
					</div>
					<div id="damageTab" className="tab-pane dr-tab-content damage-content fade active show">
						<div className="dr-row topline-results" />
						<div className="dr-row all-players-only">{panel('Players', 'damage-metrics', <div className="player-damage-metrics" />)}</div>
						<div className="dr-row single-player-only">{panel('Abilities', 'damage-metrics', <div className="damage-metrics-host" />)}</div>
						<div className="dr-row">{panel('DPS Distribution', 'damage-metrics', <div className="dps-histogram" />)}</div>
					</div>
					<div id="threatTab" className="tab-pane dr-tab-content threat-content fade">
						<div className="dr-row topline-results" />
						<div className="dr-row all-players-only">{panel('Players', '', <div className="player-threat-metrics" />)}</div>
						<div className="dr-row single-player-only">{panel('Abilities', '', <div className="threat-metrics-host" />)}</div>
						<div className="dr-row single-player-only">{panel('TPS Distribution', '', <div className="tps-histogram" />)}</div>
					</div>
					<div id="healingTab" className="tab-pane dr-tab-content healing-content fade">
						<div className="dr-row topline-results" />
						<div className="dr-row single-player-only">{panel('Abilities', 'healing-metrics', <div className="healing-spell-metrics" />)}</div>
					</div>
					<div id="damageTakenTab" className="tab-pane dr-tab-content damage-taken-content fade">
						<div className="dr-row topline-results" />
						<div className="dr-row all-players-only">{panel('Players', '', <div className="player-damage-taken-metrics" />)}</div>
						<div className="dr-row single-player-only">{panel('Abilities', '', <div className="dtps-metrics" />)}</div>
					</div>
					<div id="aurasTab" className="tab-pane dr-tab-content auras-content fade">
						<div className="dr-row dr-auras-grid">
							{panel(
								'Buffs',
								'',
								<div>
									<div className="all-players-only player-chooser-host" />
									<div className="buff-aura-metrics" />
								</div>,
								<button ref={buffsUnusedRef} className="btn btn-link dr-panel-action" type="button" />,
							)}
							{panel(
								'Debuffs',
								'',
								<div className="debuff-aura-metrics" />,
								<button ref={debuffsUnusedRef} className="btn btn-link dr-panel-action" type="button" />,
							)}
						</div>
					</div>
					<div id="castsTab" className="tab-pane dr-tab-content casts-content fade">
						<div className="dr-row">
							{panel(
								'Casts',
								'',
								<div>
									<div className="all-players-only player-chooser-host" />
									<div className="cast-metrics" />
								</div>,
							)}
						</div>
					</div>
					<div id="resourcesTab" className="tab-pane dr-tab-content resources-content fade">
						<div className="dr-row all-players-only">{panel('Resources', '', <div className="player-chooser-host" />)}</div>
						<div className="dr-row">
							<div className="resource-metrics" />
						</div>
					</div>
					<div id="eventsTab" className="tab-pane dr-tab-content events-content fade">
						<div className="dr-row">
							<div className="events-view" />
						</div>
					</div>
				</div>
			</div>,
		);
		this.rootDiv = this.rootElem.querySelector('.dr-root')!;
		this.simUI = simUI;

		this.simUI?.sim.settingsChangeEmitter.on(async () => await this.updateSettings());

		simulateRef.value?.addEventListener('click', () => {
			this.simUI?.simActionsContainer.querySelector<HTMLButtonElement>('.dps-action')?.click();
		});

		const toolbar = this.rootElem.querySelector('.dr-toolbar') as HTMLElement;
		trackStuckToolbar(toolbar, this.simUI);
		syncSubTabWithUrl(toolbar, DetailedResults.VIEW_PARAM);

		this.resultsFilter = new ResultsFilter({
			parent: this.rootElem.querySelector('.results-filter')!,
			resultsEmitter: this.resultsEmitter,
		});

		[...this.rootElem.querySelectorAll<HTMLElement>('.topline-results')]?.forEach(toplineResultsDiv => {
			new ToplineResults({ parent: toplineResultsDiv, resultsEmitter: this.resultsEmitter });
		});

		new CastMetricsTable({
			parent: this.rootElem.querySelector('.cast-metrics')!,
			resultsEmitter: this.resultsEmitter,
		});

		new ThreatMetricsTable({
			parent: this.rootElem.querySelector('.threat-metrics-host')!,
			resultsEmitter: this.resultsEmitter,
		});

		new DamageMetricsTable({
			parent: this.rootElem.querySelector('.damage-metrics-host')!,
			resultsEmitter: this.resultsEmitter,
		});

		new HealingMetricsTable({
			parent: this.rootElem.querySelector('.healing-spell-metrics')!,
			resultsEmitter: this.resultsEmitter,
		});
		new ResourceMetricsTable({
			parent: this.rootElem.querySelector('.resource-metrics')!,
			resultsEmitter: this.resultsEmitter,
		});
		new PlayerDamageMetricsTable(
			{ parent: this.rootElem.querySelector('.player-damage-metrics')!, resultsEmitter: this.resultsEmitter },
			this.resultsFilter,
		);
		new PlayerDamageMetricsTable(
			{ parent: this.rootElem.querySelector('.player-threat-metrics')!, resultsEmitter: this.resultsEmitter },
			this.resultsFilter,
			'threat',
		);
		new PlayerDamageTakenMetricsTable(
			{ parent: this.rootElem.querySelector('.player-damage-taken-metrics')!, resultsEmitter: this.resultsEmitter },
			this.resultsFilter,
		);
		const buffsTable = new AuraMetricsTable(
			{
				parent: this.rootElem.querySelector('.buff-aura-metrics')!,
				resultsEmitter: this.resultsEmitter,
			},
			false,
		);
		const debuffsTable = new AuraMetricsTable(
			{
				parent: this.rootElem.querySelector('.debuff-aura-metrics')!,
				resultsEmitter: this.resultsEmitter,
			},
			true,
		);
		bindUnusedToggle(buffsTable, buffsUnusedRef.value!);
		bindUnusedToggle(debuffsTable, debuffsUnusedRef.value!);

		new DpsHistogram({
			parent: this.rootElem.querySelector('.dps-histogram')!,
			resultsEmitter: this.resultsEmitter,
		});

		new DpsHistogram(
			{
				parent: this.rootElem.querySelector('.tps-histogram')!,
				resultsEmitter: this.resultsEmitter,
			},
			'tps',
		);

		new DtpsMetricsTable({
			parent: this.rootElem.querySelector('.dtps-metrics')!,
			resultsEmitter: this.resultsEmitter,
		});

		this.rootElem.querySelectorAll<HTMLElement>('.player-chooser-host').forEach(parent => {
			new PlayerChooser({ parent, resultsEmitter: this.resultsEmitter }, this.resultsFilter);
		});

		new EventsView({
			parent: this.rootElem.querySelector('.events-view')!,
			resultsEmitter: this.resultsEmitter,
			simUI: this.simUI,
			resultsFilter: this.resultsFilter,
		});

		this.rootElem.classList.add('hide-threat-metrics');

		this.resultsFilter.changeEmitter.on(() => this.updateResults());

		this.resultsEmitter.on((_, resultData) => {
			if (resultData?.filter.player || resultData?.filter.player === 0) {
				this.rootDiv.classList.remove('all-players');
				this.rootDiv.classList.add('single-player');
			} else {
				this.rootDiv.classList.add('all-players');
				this.rootDiv.classList.remove('single-player');
			}
		});
	}

	abstract postMessage(update: DetailedResultsUpdate): Promise<void>;

	protected async setSimRunData(simRunData: SimRunData) {
		this.latestRun = simRunData;
		await this.postMessage(
			DetailedResultsUpdate.create({
				data: {
					oneofKind: 'runData',
					runData: simRunData,
				},
			}),
		);
	}

	protected async updateSettings() {
		if (!this.simUI) return;
		await this.postMessage(
			DetailedResultsUpdate.create({
				data: {
					oneofKind: 'settings',
					settings: this.simUI.sim.toProto(),
				},
			}),
		);
	}

	private updateResults() {
		const eventID = TypedEvent.nextEventID();
		if (this.currentSimResult == null) {
			this.rootDiv.classList.add('dr-no-results');
			this.resultsEmitter.emit(eventID, null);
		} else {
			this.rootDiv.classList.remove('dr-no-results');
			this.resultsEmitter.emit(eventID, {
				eventID: eventID,
				result: this.currentSimResult,
				filter: this.resultsFilter.getFilter(),
			});
		}
	}

	protected async handleMessage(data: DetailedResultsUpdate) {
		switch (data.data.oneofKind) {
			case 'runData':
				const runData = data.data.runData;
				this.currentSimResult = runData.run ? await SimResult.fromProto(runData.run) : null;
				this.updateResults();
				break;
			case 'settings':
				const settings = data.data.settings;
				if (settings.showDamageMetrics) {
					this.rootElem.classList.remove('hide-damage-metrics');
				} else {
					this.rootElem.classList.add('hide-damage-metrics');
					const damageTabEl = document.getElementById('damageTab')!;
					const healingTabEl = document.getElementById('healingTab')!;
					if (damageTabEl.classList.contains('active')) {
						damageTabEl.classList.remove('active', 'show');
						healingTabEl.classList.add('active', 'show');

						const toolbar = this.rootElem.querySelector('.dr-toolbar') as HTMLElement;
						toolbar.querySelector('[data-bs-target="#damageTab"]')?.classList.remove('active');
						toolbar.querySelector('[data-bs-target="#healingTab"]')?.classList.add('active');
					}
				}
				this.rootElem.classList[settings.showThreatMetrics ? 'remove' : 'add']('hide-threat-metrics');
				this.rootElem.classList[settings.showHealingMetrics ? 'remove' : 'add']('hide-healing-metrics');
				this.rootElem.classList[settings.showExperimental ? 'remove' : 'add']('hide-experimental');
				break;
		}
	}
}

export class EmbeddedDetailedResults extends DetailedResults {
	constructor(parent: HTMLElement, simUI: SimUI, simResultsManager: RaidSimResultsManager) {
		super(parent, simUI);

		if (simUI.isIndividualSim()) {
			this.rootElem.classList.add('individual-sim');
		}

		simResultsManager.currentChangeEmitter.on(async () => {
			// Without a run, the tab shows no results.
			const runData = simResultsManager.getRunData() ?? SimRunData.create();
			await this.updateSettings();
			await this.setSimRunData(runData);
		});
	}

	async postMessage(update: DetailedResultsUpdate) {
		await this.handleMessage(update);
	}
}

// The toolbar gets the header's frosted background once it sticks under the header, so the two
// read as one bar. Before that it has no background, so it doesn't look like a second header.
// The page scrolls in .sim-ui, and its scroll events don't bubble, so we listen in the capture
// phase.
export const trackStuckToolbar = (toolbar: HTMLElement, simUI: SimUI | null) => {
	const updateStuck = () => {
		if (!toolbar.offsetParent) {
			toolbar.classList.remove('stuck');
			return;
		}
		const header = simUI?.simHeader.rootElem;
		const headerBottom = header ? header.getBoundingClientRect().bottom : 0;
		toolbar.classList.toggle('stuck', toolbar.getBoundingClientRect().top <= headerBottom + 1);
	};
	document.addEventListener('scroll', updateStuck, { capture: true, passive: true });
	window.addEventListener('resize', updateStuck, { passive: true });
	document.addEventListener('shown.bs.tab', updateStuck);
};

// The URL says which of the toolbar's tabs is open, like ?tab=results&view=events, so a refresh
// or a shared link opens the same one. The main tab is in the URL too, see SimHeader.
export const syncSubTabWithUrl = (toolbar: HTMLElement, param: string) => {
	const tabName = (link: Element) => link.textContent!.trim().toLowerCase().replace(/\s+/g, '-');
	const links = Array.from(toolbar.querySelectorAll<HTMLElement>('.dr-tabs [data-bs-toggle="tab"]'));

	const wanted = new URLSearchParams(window.location.search).get(param);
	const link = links.find(link => tabName(link) === wanted);
	if (link) BootstrapTab.getOrCreateInstance(link).show();

	toolbar.addEventListener('shown.bs.tab', event => {
		const url = new URL(window.location.href);
		url.searchParams.set(param, tabName(event.target as Element));
		window.history.replaceState(window.history.state, '', url);
	});
};

// The Show unused button of a buffs or debuffs panel. Auras that never came up (no procs and no
// uptime) stay out of the table until we ask for them, like the ranks of a spell we never cast.
const bindUnusedToggle = (table: AuraMetricsTable, button: HTMLButtonElement) => {
	const update = () => {
		button.classList.toggle('hide', table.unusedCount == 0);
		button.textContent = table.showUnused ? 'Hide unused' : `Show unused (${table.unusedCount})`;
	};
	button.addEventListener('click', () => table.setShowUnused(!table.showUnused));
	table.onUpdate.on(update);
	update();
};
