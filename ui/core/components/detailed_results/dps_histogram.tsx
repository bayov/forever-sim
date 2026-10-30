import { Chart, ChartDataset, Plugin } from 'chart.js';
import { ref } from 'tsx-vanilla';

import { formatToNumber } from '../../utils.js';
import { ResultComponent, ResultComponentConfig, SimResultData } from './result_component.js';

type HistogramMode = 'distribution' | 'cumulative';

const MODE_STORAGE_KEY = '__dps_histogram_mode__';

// How many bins we draw depends on the chart's width. Each bar gets at least this many pixels,
// so it stays wide enough to read and to hover.
const MIN_BIN_PX = 6;
// Each bin holds at least this many iterations on average. With fewer, the shape gets spiky
// from chance alone, like a short run of 100 iterations spread over 50 bins.
const MIN_ITERATIONS_PER_BIN = 10;
// Before the chart has a width, like while the Results tab is hidden, we draw this many.
const DEFAULT_MAX_BINS = 60;
const NICE_BIN_WIDTHS = [1, 2, 3, 4, 5, 10, 15, 20, 25, 50, 100, 200, 250, 500, 1000];

const ACCENT = 'rgba(127, 179, 240, 0.85)';
const ACCENT_FAINT = 'rgba(127, 179, 240, 0.3)';
const GOLD = '#c8a24a';
const IVORY = '#e8e2d0';
const MUTED = '#8e948f';

interface Marker {
	label: string;
	value: number;
	color: string;
}

interface Percentiles {
	p10: number;
	median: number;
	p90: number;
}

// The DPS of each iteration, as a histogram or as the share of iterations at or below each DPS.
// With metric 'tps' it's the threat per second instead, of the one player the results show.
//
// The sim counts the iterations at each whole DPS. We group those counts into bins with a
// round width. A wider chart gets more bins, so on a wide screen a run with a spread of 500 DPS
// can get bins of 5 DPS, and on a narrow one bins of 10 or 20. We mark the 10th percentile, the median and the 90th percentile:
// one in ten iterations does worse than P10, and one in ten does better than P90.
export class DpsHistogram extends ResultComponent {
	private mode: HistogramMode = loadMode();
	private chart: Chart | null = null;
	private lastResult: SimResultData | null = null;
	private readonly statsElem: HTMLElement;
	private readonly chartElem: HTMLElement;
	private readonly modeButtons: HTMLButtonElement[] = [];
	// The bin width we drew with, so a resize only draws again when it changes the bins.
	private drawnBinWidth = 0;
	private readonly metric: 'dps' | 'tps';
	// The unit after each value, like "180 DPS".
	private readonly unit: string;

	constructor(config: ResultComponentConfig, metric: 'dps' | 'tps' = 'dps') {
		config.rootCssClass = 'dps-histogram-root';
		super(config);
		this.metric = metric;
		this.unit = metric == 'tps' ? 'TPS' : 'DPS';

		const statsRef = ref<HTMLDivElement>();
		const chartRef = ref<HTMLDivElement>();
		this.rootElem.replaceChildren(
			<div className="dps-histogram-head">
				<div className="dps-histogram-stats" ref={statsRef} />
				<div className="segmented dps-histogram-mode" attributes={{ role: 'group', 'aria-label': 'Chart' }}>
					{(['distribution', 'cumulative'] as HistogramMode[]).map(mode => {
						const button = (
							<button className="segmented-option" type="button" dataset={{ mode }}>
								{mode == 'distribution' ? 'Distribution' : 'Cumulative'}
							</button>
						) as HTMLButtonElement;
						button.addEventListener('click', () => this.setMode(mode));
						this.modeButtons.push(button);
						return button;
					})}
				</div>
			</div>,
			<div className="dps-histogram-chart" ref={chartRef} />,
		);
		this.statsElem = statsRef.value!;
		this.chartElem = chartRef.value!;
		this.updateModeButtons();

		this.addOnDisposeCallback(() => this.chart?.destroy());

		// A wider or narrower chart can take a different number of bins. We check once a frame at
		// most, because a window resize sends many changes.
		let frame = 0;
		const resizeObserver = new ResizeObserver(() => {
			cancelAnimationFrame(frame);
			frame = requestAnimationFrame(() => {
				if (this.mode == 'distribution' && this.lastResult && this.binWidthFor(this.lastCounts()) != this.drawnBinWidth) this.render();
			});
		});
		resizeObserver.observe(this.chartElem);
		this.addOnDisposeCallback(() => resizeObserver.disconnect());
	}

	onSimResult(resultData: SimResultData) {
		this.lastResult = resultData;
		this.render();
	}

	private setMode(mode: HistogramMode) {
		this.mode = mode;
		try {
			window.localStorage.setItem(MODE_STORAGE_KEY, mode);
		} catch {
			// We only lose the remembered choice.
		}
		this.updateModeButtons();
		this.render();
	}

	private updateModeButtons() {
		this.modeButtons.forEach(button => button.classList.toggle('active', button.dataset.mode == this.mode));
	}

	private render() {
		this.chart?.destroy();
		this.chart = null;
		this.chartElem.replaceChildren();
		this.statsElem.replaceChildren();
		if (!this.lastResult) return;

		const counts = this.lastCounts();
		const total = counts.reduce((sum, { count }) => sum + count, 0);
		if (!total) return;

		const percentiles = computePercentiles(counts, total);
		const markers: Marker[] = [
			{ label: 'P10', value: percentiles.p10, color: MUTED },
			{ label: 'Median', value: percentiles.median, color: GOLD },
			{ label: 'P90', value: percentiles.p90, color: MUTED },
		];
		this.statsElem.replaceChildren(
			...markers.map(marker => (
				<span className={`dps-histogram-stat ${marker.label == 'Median' ? 'median' : ''}`}>
					<span className="dps-histogram-stat-label">{marker.label}</span>
					<span className="dps-histogram-stat-value">{formatToNumber(marker.value, { maximumFractionDigits: 1 })}</span>
				</span>
			)),
		);

		const canvas = document.createElement('canvas');
		this.chartElem.appendChild(canvas);
		this.chart =
			this.mode == 'distribution'
				? this.drawDistribution(canvas, counts, total, percentiles, markers)
				: this.drawCumulative(canvas, counts, total, markers);
	}

	private lastCounts(): DpsCount[] {
		if (!this.lastResult) return [];
		const { result, filter } = this.lastResult;
		let distribution;
		if (this.metric == 'tps') {
			// The sim keeps a threat distribution for each player, not for the raid.
			const players = result.getRaidIndexedPlayers(filter);
			if (players.length != 1) return [];
			distribution = players[0].tps;
		} else {
			distribution = result.getDamageMetrics(filter);
		}
		return Object.entries(distribution.hist)
			.map(([dps, count]) => ({ dps: Number(dps), count }))
			.filter(({ count }) => count > 0)
			.sort((a, b) => a.dps - b.dps);
	}

	// The smallest round bin width that keeps the bars wide enough for the chart and each bin
	// full enough. Bins of 1 DPS are the finest, because the sim counts whole DPS.
	private binWidthFor(counts: DpsCount[]): number {
		if (!counts.length) return 0;
		const total = counts.reduce((sum, { count }) => sum + count, 0);
		const chartWidth = this.chartElem.clientWidth;
		const maxBins = Math.max(1, Math.min(chartWidth ? Math.floor(chartWidth / MIN_BIN_PX) : DEFAULT_MAX_BINS, Math.floor(total / MIN_ITERATIONS_PER_BIN)));
		const span = counts[counts.length - 1].dps - counts[0].dps + 1;
		return NICE_BIN_WIDTHS.find(w => span / w <= maxBins) ?? NICE_BIN_WIDTHS[NICE_BIN_WIDTHS.length - 1];
	}

	private drawDistribution(canvas: HTMLCanvasElement, counts: DpsCount[], total: number, percentiles: Percentiles, markers: Marker[]): Chart {
		const minDps = counts[0].dps;
		const maxDps = counts[counts.length - 1].dps;
		// Each whole DPS step covers half a DPS on each side, so a bin of width 1 at 180 holds
		// every iteration from 179.5 to 180.5.
		const width = this.binWidthFor(counts);
		this.drawnBinWidth = width;
		const binOf = (dps: number) => Math.floor((dps + 0.5) / width);
		const firstBin = binOf(minDps);
		const bins = Array.from({ length: binOf(maxDps) - firstBin + 1 }, (_, i) => ({
			start: (firstBin + i) * width - 0.5,
			end: (firstBin + i + 1) * width - 0.5,
			count: 0,
		}));
		counts.forEach(({ dps, count }) => (bins[binOf(dps) - firstBin].count += count));

		// The bars between P10 and P90 are solid, the tails are faint.
		const dataset: ChartDataset<'bar', { x: number; y: number }[]> = {
			data: bins.map(bin => ({ x: (bin.start + bin.end) / 2, y: bin.count })),
			backgroundColor: bins.map(bin => {
				const center = (bin.start + bin.end) / 2;
				return center >= percentiles.p10 && center <= percentiles.p90 ? ACCENT : ACCENT_FAINT;
			}),
			hoverBackgroundColor: GOLD,
			borderRadius: 2,
			categoryPercentage: 0.9,
			barPercentage: 1,
		};

		return new Chart(canvas, {
			type: 'bar',
			data: { datasets: [dataset] },
			plugins: [markerPlugin(markers)],
			options: {
				maintainAspectRatio: false,
				animation: false,
				layout: { padding: { top: 18 } },
				plugins: {
					legend: { display: false },
					tooltip: {
						displayColors: false,
						callbacks: {
							title: items => {
								const bin = bins[items[0].dataIndex];
								return width == 1 ? `${bin.start + 0.5} ${this.unit}` : `${bin.start + 0.5} to ${bin.end - 0.5} ${this.unit}`;
							},
							label: item => {
								const count = bins[item.dataIndex].count;
								return `${formatToNumber(count)} iterations (${formatToNumber((count / total) * 100, { maximumFractionDigits: 1 })}%)`;
							},
						},
					},
				},
				scales: {
					x: {
						type: 'linear',
						min: bins[0].start,
						max: bins[bins.length - 1].end,
						offset: false,
						grid: { display: false },
						ticks: { color: MUTED, maxRotation: 0, autoSkipPadding: 16, callback: wholeDpsTick },
					},
					y: {
						beginAtZero: true,
						grid: { color: 'rgba(255, 255, 255, 0.05)' },
						border: { display: false },
						ticks: { display: false },
					},
				},
			},
		});
	}

	// The share of iterations at or below each DPS. Reading it the other way, 100% minus the
	// value is the chance to do better than that DPS.
	private drawCumulative(canvas: HTMLCanvasElement, counts: DpsCount[], total: number, markers: Marker[]): Chart {
		let below = 0;
		const points = [{ x: counts[0].dps - 0.5, y: 0 }];
		counts.forEach(({ dps, count }) => {
			below += count;
			points.push({ x: dps + 0.5, y: (below / total) * 100 });
		});

		return new Chart(canvas, {
			type: 'line',
			data: {
				datasets: [
					{
						data: points,
						borderColor: ACCENT,
						backgroundColor: 'rgba(127, 179, 240, 0.12)',
						fill: 'origin',
						borderWidth: 2,
						pointRadius: 0,
						pointHoverRadius: 4,
						pointHoverBackgroundColor: GOLD,
						tension: 0,
					},
				],
			},
			plugins: [markerPlugin(markers)],
			options: {
				maintainAspectRatio: false,
				animation: false,
				layout: { padding: { top: 18 } },
				interaction: { mode: 'nearest', axis: 'x', intersect: false },
				plugins: {
					legend: { display: false },
					tooltip: {
						displayColors: false,
						callbacks: {
							title: items => `${formatToNumber(items[0].parsed.x, { maximumFractionDigits: 1 })} ${this.unit}`,
							label: item => {
								const share = item.parsed.y ?? 0;
								return [
									`${formatToNumber(share, { maximumFractionDigits: 1 })}% of iterations at or below`,
									`${formatToNumber(100 - share, { maximumFractionDigits: 1 })}% above`,
								];
							},
						},
					},
				},
				scales: {
					x: {
						type: 'linear',
						min: points[0].x,
						max: points[points.length - 1].x,
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
}

interface DpsCount {
	dps: number;
	count: number;
}

// The axis ends sit half a DPS off a whole number, see the bins. We only label whole DPS.
function wholeDpsTick(value: string | number): string {
	return Number.isInteger(Number(value)) ? formatToNumber(Number(value)) : '';
}

function loadMode(): HistogramMode {
	try {
		return window.localStorage.getItem(MODE_STORAGE_KEY) == 'cumulative' ? 'cumulative' : 'distribution';
	} catch {
		return 'distribution';
	}
}

// Each whole DPS step holds the iterations from half a DPS below it to half a DPS above it.
// We spread them evenly over that range, so a percentile can land between two steps.
function computePercentiles(counts: DpsCount[], total: number): Percentiles {
	const at = (share: number) => {
		const target = share * total;
		let below = 0;
		for (const { dps, count } of counts) {
			if (below + count >= target) return dps - 0.5 + (target - below) / count;
			below += count;
		}
		return counts[counts.length - 1].dps + 0.5;
	};
	return { p10: at(0.1), median: at(0.5), p90: at(0.9) };
}

// Draws a dashed line with a label at each marker, over the bars or the line.
function markerPlugin(markers: Marker[]): Plugin {
	return {
		id: 'dpsMarkers',
		afterDatasetsDraw: chart => {
			const { ctx, chartArea, scales } = chart;
			ctx.save();
			ctx.font = '600 11px sans-serif';
			ctx.textAlign = 'center';
			ctx.textBaseline = 'bottom';
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
				ctx.fillStyle = marker.color == GOLD ? GOLD : IVORY;
				ctx.fillText(marker.label, x, chartArea.top - 3);
			});
			ctx.restore();
		},
	};
}
