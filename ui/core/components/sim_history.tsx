import tippy, { Instance as TippyInstance } from 'tippy.js';

import type { IndividualSimUI } from '../individual_sim_ui';
import { IndividualSimSettings } from '../proto/ui';
import { SimResult } from '../proto_utils/sim_result';
import { TypedEvent } from '../typed_event';
import { BaseModal } from './base_modal';
import { ChangeCategory, dirtySettings } from './dirty_settings';
import { changesList } from './preset_tree';
import { keyClickHint } from './sticky_tooltips';
import { RaidSimResultsManager, ResultMetric } from './raid_sim_action';

// A sim run as we keep it: the results the sidebar shows, and the settings it ran with so we
// can list what changed since.
interface SimRun {
	id: string;
	// When the run finished, in ms since the epoch.
	time: number;
	// The name we gave it. Without one, the run goes by its date and time.
	name?: string;
	dps: number;
	stdev: number;
	iterations: number;
	metrics: ResultMetric[];
	// IndividualSimSettings as JSON.
	settings: unknown;
}

// How a run's DPS compares with the latest run. The difference only counts when it's bigger
// than the noise of the two runs (a Z test at 95%), otherwise it's 'same'.
type Comparison = 'higher' | 'lower' | 'same';

// The runs we keep. A run in the settings is about 3 KB, so 100 runs fit the browser's
// storage easily.
const MAX_HISTORY = 100;
// The recent runs in the sidebar. The rest are in the history dialog.
const SIDEBAR_RUNS = 5;

const HISTORY_KEY = '__simHistory__';
const PINNED_KEY = '__pinnedSims__';

// Where a run's row shows. Each place offers different actions on it.
type RowPlace = 'pinned' | 'recent' | 'dialog';

// The results in the sidebar of an individual sim, with the runs before them.
//
// Every run we start with Simulate goes in the history, which keeps the last 100. The sidebar
// shows the latest run's results, and the last few runs under them. The history dialog shows
// all of them. Each run shows how the latest run compares with it: green when the latest did
// better, red when it did worse, and white when the difference is within the noise.
//
// We pin a run to keep it as a reference. Pinned runs stay until we unpin them, even after
// they drop out of the history, and they show at the top. We can name them and drag them
// into any order.
//
// Hovering a run lists the settings that changed since it, like the tooltip of a modified
// preset. Ctrl+click on a run loads the settings it ran with. The history stays in the
// browser's storage, so it's still there when we come back.
export class SimHistory {
	private readonly simUI: IndividualSimUI<any>;
	// Newest first.
	private history: SimRun[];
	// In the order we put them.
	private pinned: SimRun[];
	// The settings when we pressed Simulate. We keep them for the run, because the settings
	// may change while it runs. Runs without them (the one-iteration runs for the timeline)
	// don't go in the history.
	private runSettings: unknown = null;
	// While a run goes, the sidebar shows its progress in place of the latest run's results.
	private running = false;
	private draggedId: string | null = null;
	// The pinned and recent runs. They stay in view while a run goes, under its progress.
	private readonly listElem: HTMLElement;
	private tooltips: TippyInstance[] = [];
	private latestTooltip: TippyInstance | null = null;
	private dialog: SimHistoryModal | null = null;

	constructor(simUI: IndividualSimUI<any>, resultsManager: RaidSimResultsManager) {
		this.simUI = simUI;
		resultsManager.simHistory = this;
		this.history = this.load(HISTORY_KEY);
		this.pinned = this.load(PINNED_KEY);
		this.listElem = (<div className="sim-history" />) as HTMLElement;
		// The warnings go under the runs.
		simUI.resultsViewer.rootElem.insertBefore(this.listElem, simUI.resultsViewer.warningElem);
		this.render();
	}

	startRun() {
		this.runSettings = IndividualSimSettings.toJson(this.simUI.toProto());
		this.running = true;
	}

	// A run that we stopped, or that failed, leaves its progress in place of the results, so
	// we show the latest run's results again.
	finishRun() {
		this.runSettings = null;
		this.running = false;
		this.render();
	}

	showRun(simResult: SimResult) {
		if (this.runSettings !== null) {
			const dps = simResult.raidMetrics.dps;
			this.history.unshift({
				id: `${Date.now()}-${Math.random().toString(36).slice(2)}`,
				time: Date.now(),
				dps: dps.avg,
				stdev: dps.stdev,
				iterations: simResult.iterations,
				metrics: RaidSimResultsManager.toplineColumns(simResult),
				settings: this.runSettings,
			});
			this.history.splice(MAX_HISTORY);
			this.runSettings = null;
			this.store();
		}
		this.running = false;
		this.render();
	}

	private get latest(): SimRun | null {
		return this.history[0] ?? null;
	}

	// Draws the sidebar, and the history dialog when it's open.
	private render() {
		this.renderLatest();
		this.tooltips.forEach(tooltip => tooltip.destroy());
		this.tooltips = [];
		this.dialog?.renderList();
		this.listElem.replaceChildren(
			...(this.pinned.length ? [this.section('Pinned', this.pinned, 'pinned')] : []),
			...(this.history.length ? [this.section('Recent', this.history.slice(0, SIDEBAR_RUNS), 'recent', this.historyButton())] : []),
		);
	}

	// The latest run's results, with the same tooltips as after a run.
	private renderLatest() {
		const latest = this.latest;
		if (this.running || !latest) return;
		this.latestTooltip?.destroy();
		this.latestTooltip = null;
		const content = (<div className="results-sim">{RaidSimResultsManager.buildResultsList(latest.metrics)}</div>) as HTMLElement;
		this.simUI.resultsViewer.setContent(content);

		RaidSimResultsManager.addMetricTooltips(content, ['dps']);
		const dpsElem = content.querySelector<HTMLElement>(`.${RaidSimResultsManager.resultMetricClasses['dps']}`);
		if (dpsElem) this.latestTooltip = this.addTooltip(dpsElem, () => this.runTooltip(latest), []);
	}

	private section(title: string, runs: SimRun[], place: RowPlace, action?: Element): Element {
		return (
			<div className={`sim-history-section sim-history-${place}`}>
				<div className="sim-history-section-header">
					<span className="sim-history-section-title">{title}</span>
					{action}
				</div>
				<div className="sim-history-list">{runs.map(run => this.buildRow(run, place))}</div>
			</div>
		) as Element;
	}

	private historyButton(): Element {
		const button = (
			<button className="sim-history-open btn btn-link">
				{`History (${this.history.length})`}
				<i className="fas fa-up-right-from-square ms-1" />
			</button>
		) as HTMLButtonElement;
		button.addEventListener('click', () => this.openDialog());
		return button;
	}

	private openDialog() {
		const dialog = new SimHistoryModal(this.simUI.rootElem, this);
		dialog.addOnHideCallback(() => {
			if (this.dialog === dialog) this.dialog = null;
		});
		this.dialog = dialog;
		dialog.renderList();
		dialog.open();
	}

	// Rows for the history dialog, newest first, with their tooltips. The dialog destroys the
	// tooltips when it draws the list again.
	dialogRows(tooltips: TippyInstance[]): HTMLElement[] {
		return this.history.map(run => this.buildRow(run, 'dialog', tooltips));
	}

	// One run in a row: its name, its DPS, and how the latest run compares with it.
	//
	// The rows of the sidebar's sections share the columns of one grid, so the numbers line up
	// from section to section, see .sim-history.
	private buildRow(run: SimRun, place: RowPlace, tooltips = this.tooltips): HTMLElement {
		const latest = this.latest;
		const isLatest = latest?.id === run.id;
		const isPinned = this.isPinned(run);
		const comparison = this.compare(run);
		const diff = latest && !isLatest ? latest.dps - run.dps : null;

		const pinButton = (
			<button className={`sim-run-pin ${isPinned ? 'active' : ''}`} attributes={{ 'aria-label': isPinned ? 'Unpin' : 'Pin' }}>
				<i
					className={`fas ${
						!isPinned ? 'fa-circle sim-run-pin-placeholder' : place === 'pinned' ? 'fa-grip-vertical' : 'fa-thumbtack'
					} sim-run-pin-idle`}
				/>
				<i className="fas fa-thumbtack sim-run-pin-hover" />
			</button>
		) as HTMLButtonElement;
		const renameButton = (
			<button className="sim-run-action sim-run-rename" attributes={{ 'aria-label': 'Rename' }}>
				<i className="fas fa-pen" />
			</button>
		) as HTMLButtonElement;

		const row = (
			<div className={`sim-run sim-run-${comparison} ${isLatest ? 'sim-run-latest' : ''}`}>
				{pinButton}
				<span className="sim-run-title">
					<span className={`sim-run-name ${run.name ? '' : 'sim-run-unnamed'}`}>{runLabel(run)}</span>
					{renameButton}
				</span>
				{place === 'dialog' ? <span className="sim-run-iterations">{`${run.iterations} it.`}</span> : undefined}
				<span className="sim-run-dps">{run.dps.toFixed(1)}</span>
				{isLatest ? <span className="sim-run-latest-tag">latest</span> : undefined}
				{diff !== null ? <span className="sim-run-diff">{signed(diff, 1)}</span> : undefined}
				{diff !== null ? <span className="sim-run-percent">{`(${signed((diff / (run.dps || 1)) * 100, 1)}%)`}</span> : undefined}
			</div>
		) as HTMLElement;
		row.dataset.id = run.id;

		pinButton.addEventListener('click', () => this.togglePin(run));
		renameButton.addEventListener('click', () => this.rename(row, run));
		row.querySelector('.sim-run-name')!.addEventListener('dblclick', () => this.rename(row, run));
		row.addEventListener('click', event => {
			if ((event.ctrlKey || event.metaKey) && !(event.target as Element).closest('.sim-run-action, .sim-run-pin, input')) this.applySettings(run);
		});
		if (place === 'pinned') this.addDragAndDrop(row, run);
		const rowTooltip = this.addTooltip(row, () => this.runTooltip(run), tooltips);
		this.addButtonTooltip(pinButton, isPinned ? 'Unpin' : 'Pin', rowTooltip, tooltips);
		this.addButtonTooltip(renameButton, 'Rename', rowTooltip, tooltips);
		return row;
	}

	private isPinned(run: SimRun): boolean {
		return this.pinned.some(other => other.id === run.id);
	}

	private togglePin(run: SimRun) {
		if (this.isPinned(run)) this.pinned = this.pinned.filter(other => other.id !== run.id);
		else this.pinned.push({ ...run });
		this.store();
		this.render();
	}

	// Loads the settings the run ran with, all of them: gear, talents, rotation, buffs and the
	// encounter.
	private applySettings(run: SimRun) {
		try {
			const settings = IndividualSimSettings.fromJson(run.settings as any);
			this.tooltips.forEach(tooltip => tooltip.hide());
			this.simUI.fromProto(TypedEvent.nextEventID(), settings);
		} catch (e) {
			console.warn('Failed to load the settings of a saved run: ' + e);
		}
	}

	// Puts a text field in place of the name. Enter or a click elsewhere keeps the new name,
	// and Escape keeps the old one. An empty name goes back to the run's time.
	private rename(row: HTMLElement, run: SimRun) {
		const nameElem = row.querySelector('.sim-run-name');
		if (!nameElem) return;
		const input = (<input className="sim-run-name-input" type="text" value={run.name ?? ''} />) as HTMLInputElement;
		input.placeholder = defaultLabel(run);
		row.draggable = false;
		row.classList.add('renaming');
		this.tooltips.forEach(tooltip => tooltip.hide());
		nameElem.replaceWith(input);
		input.focus();
		input.select();

		let done = false;
		const finish = (keep: boolean) => {
			if (done) return;
			done = true;
			if (keep) {
				const name = input.value.trim() || undefined;
				// A pinned run is a copy of the one in the history, so we rename both.
				[...this.history, ...this.pinned].filter(other => other.id === run.id).forEach(other => (other.name = name));
				this.store();
			}
			this.render();
		};
		input.addEventListener('keydown', event => {
			if (event.key === 'Enter') finish(true);
			else if (event.key === 'Escape') finish(false);
		});
		input.addEventListener('blur', () => finish(true));
	}

	// We drop a pinned run above or below another one, by which half of the row we're over.
	private addDragAndDrop(row: HTMLElement, run: SimRun) {
		row.draggable = true;
		const clearMarks = () => row.classList.remove('drop-before', 'drop-after');
		row.addEventListener('dragstart', event => {
			this.draggedId = run.id;
			row.classList.add('dragging');
			event.dataTransfer!.effectAllowed = 'move';
			event.dataTransfer!.setData('text/plain', runLabel(run));
			this.tooltips.forEach(tooltip => {
				tooltip.hide();
				tooltip.disable();
			});
		});
		row.addEventListener('dragend', () => {
			this.draggedId = null;
			this.render();
		});
		row.addEventListener('dragover', event => {
			if (!this.draggedId || this.draggedId === run.id) return;
			event.preventDefault();
			event.dataTransfer!.dropEffect = 'move';
			const after = isLowerHalf(row, event);
			row.classList.toggle('drop-before', !after);
			row.classList.toggle('drop-after', after);
		});
		row.addEventListener('dragleave', clearMarks);
		row.addEventListener('drop', event => {
			event.preventDefault();
			clearMarks();
			const dragged = this.pinned.find(other => other.id === this.draggedId);
			if (!dragged || dragged.id === run.id) return;
			const after = isLowerHalf(row, event);
			this.pinned = this.pinned.filter(other => other !== dragged);
			const target = this.pinned.findIndex(other => other.id === run.id);
			this.pinned.splice(target + (after ? 1 : 0), 0, dragged);
			this.store();
		});
	}

	private addTooltip(elem: HTMLElement, content: () => Element, tooltips = this.tooltips): TippyInstance {
		const tooltip = tippy(elem, {
			placement: 'right',
			// The changed items and enchants are links, and we move the mouse onto them to
			// see their tooltips. The tooltip goes in body, because the sidebar would cut it
			// off.
			interactive: true,
			appendTo: () => document.body,
			onShow: instance => {
				if (elem.classList.contains('renaming')) return false;
				if (elem.querySelector('.sim-run-pin:hover, .sim-run-action:hover')) return false;
				instance.setContent(content());
				return undefined;
			},
		});
		tooltips.push(tooltip);
		return tooltip;
	}

	// A small button in the row says what it does in its own tooltip. The run's tooltip steps
	// aside while we hover the button, and comes back when we move on to the rest of the row.
	private addButtonTooltip(button: HTMLElement, label: string, rowTooltip: TippyInstance, tooltips: TippyInstance[]) {
		tooltips.push(tippy(button, { content: label }));
		button.addEventListener('mouseenter', () => rowTooltip.hide());
		button.addEventListener('mouseleave', () => {
			if (rowTooltip.reference.matches(':hover')) rowTooltip.show();
		});
	}

	// The run's results, how the latest run compares with it, and the settings that changed
	// since it.
	private runTooltip(run: SimRun): Element {
		const latest = this.latest;
		const isLatest = latest?.id === run.id;
		let verdict = '';
		if (latest && !isLatest) {
			const diff = latest.dps - run.dps;
			const percent = (Math.abs(diff) / (run.dps || 1)) * 100;
			const amount = `${Math.abs(diff).toFixed(2)} DPS (${percent.toFixed(2)}%)`;
			const comparison = this.compare(run);
			if (comparison === 'higher') verdict = `The latest run did ${amount} better.`;
			else if (comparison === 'lower') verdict = `The latest run did ${amount} worse.`;
			else verdict = `The latest run is within the noise of this one (${signed(diff, 2)} DPS).`;
		}
		// When it ran with the current settings, there's nothing to load.
		const categories = this.changesSince(run);
		return (
			<>
				{run.name ? <p className="mb-0 fw-bold">{run.name}</p> : undefined}
				<p className="mb-0 sim-run-time">{defaultLabel(run)}</p>
				<p className="mb-0">{`${run.dps.toFixed(2)} DPS (±${run.stdev.toFixed(0)}) over ${run.iterations} iterations`}</p>
				{verdict ? <p className={`mb-0 sim-run-verdict sim-run-${this.compare(run)}`}>{verdict}</p> : undefined}
				{this.changesNote(categories)}
				{categories?.length === 0 ? undefined : keyClickHint('ctrl', 'load its settings')}
			</>
		) as Element;
	}

	private changesNote(categories: ChangeCategory[] | null): Element {
		if (categories === null) return (<p className="mt-2 mb-0">We couldn't compare its settings with the current ones.</p>) as Element;
		if (!categories.length) return (<p className="mt-2 mb-0">It ran with the current settings.</p>) as Element;
		return (
			<div className="mt-2">
				<p className="mb-1 preset-tree-changed-title">Changed since this run:</p>
				{changesList(categories)}
			</div>
		) as Element;
	}

	// The settings that changed since the run, in lines like 'Duration: 120 → 90'. The stat
	// weights don't change the DPS, so we leave them out.
	private changesSince(run: SimRun): ChangeCategory[] | null {
		try {
			const settings = IndividualSimSettings.fromJson(run.settings as any);
			return dirtySettings
				.changesFrom(() => this.simUI.fromProto(TypedEvent.nextEventID(), IndividualSimSettings.clone(settings)))
				.filter(category => category.name !== 'Stat Weights');
		} catch (e) {
			console.warn('Failed to compare a saved run with the current settings: ' + e);
			return null;
		}
	}

	private compare(run: SimRun): Comparison {
		const latest = this.latest;
		if (!latest || latest.id === run.id) return 'same';
		const error = (run: SimRun) => run.stdev / Math.sqrt(Math.max(run.iterations, 1));
		const diff = latest.dps - run.dps;
		const noise = Math.sqrt(error(latest) ** 2 + error(run) ** 2);
		if (diff === 0 || (noise > 0 && Math.abs(diff) / noise <= 1.96)) return 'same';
		return diff > 0 ? 'higher' : 'lower';
	}

	private load(key: string): SimRun[] {
		try {
			const json = window.localStorage.getItem(this.simUI.getStorageKey(key));
			const runs = json ? JSON.parse(json) : [];
			return Array.isArray(runs) ? runs : [];
		} catch (e) {
			console.warn('Failed to load the sim history: ' + e);
			return [];
		}
	}

	private store() {
		try {
			window.localStorage.setItem(this.simUI.getStorageKey(HISTORY_KEY), JSON.stringify(this.history));
			window.localStorage.setItem(this.simUI.getStorageKey(PINNED_KEY), JSON.stringify(this.pinned));
		} catch (e) {
			console.warn('Failed to save the sim history: ' + e);
		}
	}
}

// Every run we kept, newest first, with the same actions as in the sidebar.
class SimHistoryModal extends BaseModal {
	private readonly history: SimHistory;
	private tooltips: TippyInstance[] = [];

	constructor(parent: HTMLElement, history: SimHistory) {
		super(parent, 'sim-history-modal', { title: 'Sim History', size: 'lg', scrollContents: true, disposeOnClose: true });
		this.history = history;
		this.addOnHideCallback(() => this.tooltips.forEach(tooltip => tooltip.destroy()));
	}

	renderList() {
		this.tooltips.forEach(tooltip => tooltip.destroy());
		this.tooltips = [];
		const rows = this.history.dialogRows(this.tooltips);
		this.body.replaceChildren(
			<p className="sim-history-modal-intro">
				{`The last ${MAX_HISTORY} runs. Pin a run to keep it as a reference. Hover a run to see what changed since, and Ctrl+click it to load its settings.`}
			</p>,
			rows.length ? <div className="sim-history-list">{rows}</div> : <p className="mb-0">No runs yet.</p>,
		);
	}
}

function signed(value: number, digits: number): string {
	return `${value >= 0 ? '+' : '-'}${Math.abs(value).toFixed(digits)}`;
}

// A run's name, or its date and time when it has none.
function runLabel(run: SimRun): string {
	return run.name ?? defaultLabel(run);
}

// The date and time of the run, like '2026-05-06 22:33'.
function defaultLabel(run: SimRun): string {
	const date = new Date(run.time);
	const pad = (value: number) => String(value).padStart(2, '0');
	return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

function isLowerHalf(row: HTMLElement, event: DragEvent): boolean {
	const rect = row.getBoundingClientRect();
	return event.clientY > rect.top + rect.height / 2;
}
