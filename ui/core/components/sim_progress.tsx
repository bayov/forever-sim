import { ProgressMetrics } from '../proto/api';

// A progress bar for a sim that runs, with how far along it is and the time it has left.
//
// The bar fills as the iterations complete, and shows the percent, the iterations (and the sims,
// when there's more than one, like for stat weights), and an estimate of the time left. We
// estimate the time from how fast the iterations went so far. The presims at the start don't
// count, so while they run, the bar moves without filling and says so.
//
// The sidebar's bar is too narrow for all of that on one line. So there the caller passes a
// countParent (the Iterations label), and the iteration count goes there instead of on the bar.
//
// Progress comes in every 100ms or so, so a short sim can jump from 40% to done, and the caller
// would remove the bar before we ever see it full. So when the sim is done, the bar fills the rest
// of the way and flashes, and the caller awaits finish() before it removes the bar.
export class SimProgress {
	readonly rootElem: HTMLElement;
	private readonly fillElem: HTMLElement;
	private readonly doneElem: HTMLElement;
	private readonly etaElem: HTMLElement;
	private readonly countElem: HTMLElement | null;
	// When the first iterations completed, and how many had, so the presims don't slow down the
	// estimate.
	private start: { time: number; iterations: number } | null = null;
	// The last counts we showed, so the final progress (which may not have counts, like for stat
	// weights) can show them full.
	private last: { total: number; totalSims: number } = { total: 0, totalSims: 0 };
	// When the sim finished well, or null while it runs or when it was aborted or failed.
	private completedAt: number | null = null;

	constructor(className = '', countParent?: HTMLElement) {
		this.countElem = countParent ? ((<span className="sim-progress-count" />) as HTMLElement) : null;
		if (this.countElem) countParent!.appendChild(this.countElem);
		this.fillElem = (<div className="sim-progress-fill" />) as HTMLElement;
		this.doneElem = (<span className="sim-progress-done">Starting</span>) as HTMLElement;
		this.etaElem = (<span className="sim-progress-eta" />) as HTMLElement;
		this.rootElem = (
			<div className={`sim-progress sim-progress-indeterminate ${className}`} attributes={{ role: 'progressbar' }}>
				{this.fillElem}
				<div className="sim-progress-label">
					{this.doneElem}
					{this.etaElem}
				</div>
			</div>
		) as HTMLElement;
	}

	update(progress: ProgressMetrics) {
		const final = progress.finalRaidResult ?? progress.finalWeightResult ?? progress.finalBulkResult;
		if (final) {
			if (!final.error) this.complete();
			return;
		}

		const total = progress.totalIterations;
		const completed = progress.completedIterations;
		const running = !progress.presimRunning && total > 0 && completed > 0;
		this.rootElem.classList.toggle('sim-progress-indeterminate', !running);
		if (!running) {
			this.doneElem.textContent = progress.presimRunning ? 'Presimulating' : 'Starting';
			this.etaElem.textContent = '';
			if (this.countElem) this.countElem.textContent = '';
			return;
		}

		this.last = { total, totalSims: progress.totalSims };
		// The leveling plan sends no final result, so we also count a run whose iterations all
		// completed as done.
		if (completed >= total) {
			this.complete();
			return;
		}

		const fraction = completed / total;
		const percent = Math.floor(fraction * 100);
		this.fillElem.style.width = `${fraction * 100}%`;
		this.setCounts(percent, completed, total, progress.completedSims, progress.totalSims);

		const now = performance.now();
		if (!this.start) this.start = { time: now, iterations: completed };
		const rate = (completed - this.start.iterations) / (now - this.start.time);
		// The first few updates go too fast or too slow to tell, so we wait for a second of them.
		this.etaElem.textContent = now - this.start.time > 1000 && rate > 0 ? `${timeLeft((total - completed) / rate)} left` : '';
	}

	// Waits until the finish animation has played, when the sim finished well. When it was
	// aborted or failed, the bar didn't fill, and we return right away.
	async finish() {
		if (this.completedAt === null) return;
		const wait = this.completedAt + FINISH_MS - performance.now();
		if (wait > 0) await new Promise(resolve => setTimeout(resolve, wait));
	}

	private complete() {
		if (this.completedAt !== null) return;
		this.completedAt = performance.now();
		this.rootElem.classList.remove('sim-progress-indeterminate');
		this.rootElem.classList.add('sim-progress-complete');
		this.fillElem.style.width = '100%';
		const { total, totalSims } = this.last;
		this.setCounts(100, total, total, totalSims, totalSims);
		this.etaElem.replaceChildren(<i className="fas fa-check me-1" />, 'Done');
	}

	private setCounts(percent: number, completed: number, total: number, completedSims: number, totalSims: number) {
		this.rootElem.setAttribute('aria-valuenow', String(percent));
		const sims = totalSims > 1 ? ` · ${completedSims} / ${totalSims} sims` : '';
		const iterations = total > 0 ? `${count(completed, total)} / ${count(total, total)}${sims}` : '';
		if (this.countElem) {
			this.doneElem.textContent = `${percent}%`;
			this.countElem.textContent = iterations;
		} else {
			this.doneElem.textContent = iterations ? `${percent}% · ${iterations}` : `${percent}%`;
		}
	}

	remove() {
		this.rootElem.remove();
		this.countElem?.remove();
	}
}

// How long the bar takes to fill the rest of the way and flash when the sim is done. It matches
// .sim-progress-complete in _raid_sim_action.scss.
const FINISH_MS = 600;

// An iteration count, like '2,500', or '110K' when the total is big, so the text fits the
// sidebar's bar.
function count(value: number, total: number): string {
	return total >= 10000 ? new Intl.NumberFormat(undefined, { notation: 'compact', maximumFractionDigits: 1 }).format(value) : value.toLocaleString();
}

// Like '<1s', '42s' or '3m 05s'.
function timeLeft(ms: number): string {
	const seconds = Math.round(ms / 1000);
	if (seconds < 1) return '<1s';
	if (seconds < 60) return `${seconds}s`;
	return `${Math.floor(seconds / 60)}m ${String(seconds % 60).padStart(2, '0')}s`;
}
