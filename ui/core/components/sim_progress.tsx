import { ProgressMetrics } from '../proto/api';

// A progress bar for a sim that runs, with how far along it is and the time it has left.
//
// The bar fills as the iterations complete, and shows the percent, the iterations (and the sims,
// when there's more than one, like for stat weights), and an estimate of the time left. We
// estimate the time from how fast the iterations went so far. The presims at the start don't
// count, so while they run, the bar moves without filling and says so.
export class SimProgress {
	readonly rootElem: HTMLElement;
	private readonly fillElem: HTMLElement;
	private readonly doneElem: HTMLElement;
	private readonly etaElem: HTMLElement;
	// When the first iterations completed, and how many had, so the presims don't slow down the
	// estimate.
	private start: { time: number; iterations: number } | null = null;

	constructor(className = '') {
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
		const total = progress.totalIterations;
		const completed = progress.completedIterations;
		const running = !progress.presimRunning && total > 0 && completed > 0;
		this.rootElem.classList.toggle('sim-progress-indeterminate', !running);
		if (!running) {
			this.doneElem.textContent = progress.presimRunning ? 'Presimulating' : 'Starting';
			this.etaElem.textContent = '';
			return;
		}

		const fraction = Math.min(completed / total, 1);
		const percent = Math.floor(fraction * 100);
		this.fillElem.style.width = `${fraction * 100}%`;
		this.rootElem.setAttribute('aria-valuenow', String(percent));
		const sims = progress.totalSims > 1 ? ` · ${progress.completedSims} / ${progress.totalSims} sims` : '';
		this.doneElem.textContent = `${percent}% · ${count(completed, total)} / ${count(total, total)}${sims}`;

		const now = performance.now();
		if (!this.start) this.start = { time: now, iterations: completed };
		const rate = (completed - this.start.iterations) / (now - this.start.time);
		// The first few updates go too fast or too slow to tell, so we wait for a second of them.
		this.etaElem.textContent = now - this.start.time > 1000 && rate > 0 ? `${timeLeft((total - completed) / rate)} left` : '';
	}
}

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
