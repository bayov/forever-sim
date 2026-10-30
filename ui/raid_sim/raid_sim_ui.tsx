// The raid sim page (docs/raid_sim.md, The page): a results viewer for the comp search's JSON, the search
// through the native server, and a comp editor that runs one raid in the sim's workers. It is a page of its
// own, not a spec's page or the DPS Report.
import { CompareResults } from '../core/components/compare_results/compare_results';
import { EmbeddedDetailedResults } from '../core/components/detailed_results';
import { RaidSimResultsManager } from '../core/components/raid_sim_action';
import { SimHeader } from '../core/components/sim_header';
import { SimHistory } from '../core/components/sim_history';
import { raidSimStatus } from '../core/launched_sims';
import { CompResult, CompScenarioResult, CompSearchResult } from '../core/proto/api';
import { Sim } from '../core/sim';
import { SimUI } from '../core/sim_ui';
import { EventID, TypedEvent } from '../core/typed_event';
import { CompEditor, RaidState } from './comp_editor';
import { loadRaidSimData, RaidSimData } from './data';
import { addRaidPresets, loadRaidFromLink, raidHistoryHost,RaidJsonExporter, RaidJsonImporter, RaidLinkExporter } from './raid_settings';
import { RESULT_PARAM, ResultLoader } from './result_loader';
import { ResultView } from './result_view';
import { SearchForm } from './search_form';

export const RESULT_TAB = 'raid-sim-result';
export const SEARCH_TAB = 'raid-sim-search';
export const EDITOR_TAB = 'raid-sim-editor';
export const DETAILS_TAB = 'raid-sim-details';
export const COMPARE_TAB = 'raid-sim-compare';

// The last search run on this page, so a reload shows it again. Save JSON on the result keeps it for good.
const LAST_SEARCH_KEY = '__forever_raid_sim__lastSearch__';

export class RaidSimUI extends SimUI {
	readonly resultsManager: RaidSimResultsManager;
	readonly dataPromise: Promise<RaidSimData | null>;
	private readonly resultHost: HTMLElement;
	readonly editor: CompEditor;
	readonly history: SimHistory;
	// The raid whose results Raid results shows, or null when it shows none.
	private shownRaid: RaidState | null = null;

	constructor(parentElem: HTMLElement) {
		super(parentElem, new Sim(), {
			cssClass: 'raid-sim-page',
			cssScheme: 'raid',
			spec: null,
			simStatus: raidSimStatus,
			knownIssues: [
				'Tanks and healers are placeholders: they give what their preset gives, and deal, take and heal nothing.',
				'The comp search runs only through the native server (the downloadable sim, or make host). A page served as static files shows results and runs single raids.',
			],
		});
		this.rootElem.classList.add('raid-sim-ui');

		// The data file is only for names and the editor's members, so a page without it still shows a result.
		this.dataPromise = loadRaidSimData().catch(e => {
			console.error('The raid sim data file failed to load:', e);
			return null;
		});

		this.resultsManager = new RaidSimResultsManager(this);

		// Building a raid and running it comes first. The comp search, which finds the best raids, comes after.
		this.addTab('Comp', EDITOR_TAB, '');
		this.addTab('Raid results', DETAILS_TAB, '<div class="detailed-results"></div>');
		this.addTab('Compare', COMPARE_TAB, '<div class="compare-results"></div>');
		this.addTab('Search', SEARCH_TAB, '');
		this.addTab('Search result', RESULT_TAB, '');

		const resultTab = this.tabElem(RESULT_TAB);
		// A result linked with the page's address opens its tab unless the address names another one
		// (?result=…&tab=comp stays on the Comp tab).
		const loader = new ResultLoader(resultTab, (result, source, fromUrl) =>
			this.showResult(result, source, !fromUrl || !new URLSearchParams(window.location.search).has(SimHeader.TAB_PARAM)),
		);
		this.resultHost = (<div className="rs-result-host" />) as HTMLElement;
		resultTab.appendChild(this.resultHost);
		this.resultHost.appendChild(loader.emptyState());

		new SearchForm(this.tabElem(SEARCH_TAB), this, (result: CompSearchResult) => {
			this.showResult(result, 'the search run on this page');
			this.keepLastSearch(result);
		});

		this.editor = new CompEditor(this.tabElem(EDITOR_TAB), this);
		this.addAction('Run comp', 'dps-action', () => this.editor.run());

		new EmbeddedDetailedResults(this.tabElem(DETAILS_TAB).querySelector('.detailed-results') as HTMLElement, this, this.resultsManager);

		// What a spec page has for its character, for the raid (raid_settings.tsx): preset raids in the sidebar,
		// the sim history under the raid DPS with the Compare tab, and the raid's JSON and link.
		addRaidPresets(this.rootElem.querySelector('.sim-sidebar-presets') as HTMLElement, this, this.editor);
		this.history = new SimHistory(raidHistoryHost(this, this.editor), this.resultsManager);
		// A run's results the history shows (Ctrl+click on a run) are of the raid it just put in place.
		this.sim.simResultEmitter.on((eventID, result) => {
			this.resultsManager.setSimResult(eventID, result);
			this.setShownRaid(this.editor.getState());
		});
		this.history.changeEmitter.on(() => {
			this.addSidebarNote();
			this.updateStale();
		});
		this.editor.changeEmitter.on(() => this.updateStale());
		this.addSidebarNote();
		new CompareResults(this.tabElem(COMPARE_TAB).querySelector('.compare-results') as HTMLElement, this, this.history);
		this.simHeader.addImportLink('JSON', new RaidJsonImporter(this.rootElem, this, this.editor));
		this.simHeader.addExportLink('Link', new RaidLinkExporter(this.rootElem, this, this.editor));
		this.simHeader.addExportLink('JSON', new RaidJsonExporter(this.rootElem, this, this.editor));
		loadRaidFromLink(this.editor);
		// A link pasted into this page's own tab changes only the hash, which doesn't reload the page.
		window.addEventListener('hashchange', () => loadRaidFromLink(this.editor));
		this.showLatestRun();

		this.simHeader.syncTabWithUrl();
		if (new URLSearchParams(window.location.search).has(RESULT_PARAM)) loader.loadFromUrlParameter();
		else this.showLastSearch();
	}

	// The sidebar's ± is a standard deviation of the iterations, and the Comp tab's the average's 95% interval.
	private addSidebarNote() {
		const results = this.resultsViewer.contentElem.querySelector('.results-sim');
		if (!results || results.querySelector('.rs-sidebar-note')) return;
		results.append(
			(
				<p className="rs-sidebar-note">± here is one standard deviation of the iterations; the Comp tab’s ±(95%) is the average’s 95% interval.</p>
			) as HTMLElement,
		);
	}

	// After a reload, Raid results show the latest run's results again, as a spec page's Results do. They're
	// marked as another raid's when the raid changed since.
	private async showLatestRun() {
		await this.history.ready;
		const latest = this.history.getRuns().recent[0];
		if (!latest || this.resultsManager.getRunData() || this.history.isRunning()) return;
		const result = await this.history.loadResults(latest, true);
		if (!result || this.resultsManager.getRunData() || this.history.isRunning()) return;
		this.resultsManager.setSimResult(TypedEvent.nextEventID(), result);
		try {
			this.setShownRaid(CompEditor.readState(latest.settings));
		} catch (e) {
			console.warn('Failed to read the latest run’s raid:', e);
		}
	}

	setShownRaid(raid: RaidState) {
		this.shownRaid = raid;
		this.updateStale();
	}

	// The sidebar's latest run and Raid results each say so when they're of another raid than the one on the
	// Comp tab.
	private updateStale() {
		const current = this.editor.getState();
		const differs = (raid: unknown) => {
			try {
				return !CompEditor.sameRaid(CompEditor.readState(raid), current);
			} catch (e) {
				return true;
			}
		};
		const latest = this.history.getRuns().recent[0];
		this.editor.setSidebarStale(!!latest && !this.history.isRunning() && differs(latest.settings));
		this.editor.setDetailsStale(!!this.resultsManager.getRunData() && (!this.shownRaid || differs(this.shownRaid)));
	}

	tabElem(cssClass: string): HTMLElement {
		return this.simTabContentsContainer.querySelector(`#${cssClass}-tab`) as HTMLElement;
	}

	activate(cssClass: string) {
		this.simHeader.activateTab(`${cssClass}-tab`);
	}

	private keepLastSearch(result: CompSearchResult) {
		try {
			window.localStorage.setItem(LAST_SEARCH_KEY, JSON.stringify({ at: Date.now(), result: CompSearchResult.toJson(result) }));
		} catch (e) {
			// A private window, or a result past the storage's quota: Save JSON still keeps it.
			console.warn('Failed to keep the last search:', e);
		}
	}

	// The last search run on this page, in its tab, without opening that tab.
	private showLastSearch() {
		try {
			const saved = window.localStorage.getItem(LAST_SEARCH_KEY);
			if (!saved) return;
			const { at, result } = JSON.parse(saved);
			const when = new Date(at).toLocaleString();
			this.showResult(CompSearchResult.fromJson(result, { ignoreUnknownFields: true }), `the last search run on this page, ${when}`, false);
		} catch (e) {
			console.warn('Failed to read the last search:', e);
		}
	}

	async showResult(result: CompSearchResult, source: string, activate = true) {
		const data = await this.dataPromise;
		const view = new ResultView(result, data, {
			source,
			onOpenComp: (comp: CompResult, scenario: CompScenarioResult) => {
				this.editor.loadFromResult(comp, scenario, result.request);
				this.activate(EDITOR_TAB);
			},
		});
		this.resultHost.replaceChildren(view.rootElem);
		if (activate) this.activate(RESULT_TAB);
	}

	getStorageKey(postfix: string): string {
		return '__forever_raid_sim__' + postfix;
	}

	// The page keeps its own settings (the comp editor and the search form), so there are none to default.
	applyDefaults(_eventID: EventID) {
		return;
	}

	toLink(): string {
		return window.location.href;
	}
}
