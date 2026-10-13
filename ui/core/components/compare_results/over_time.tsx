import { Chart } from 'chart.js';

import { ResourceType } from '../../proto/api';
import { resourceNames } from '../../proto_utils/names';
import { formatToNumber } from '../../utils';
import { formatFixed } from './compare_table';
import { IterationKind, IterationSeries, RunSeries, SERIES_STEP } from './run_series';

const COLOR_A = '#7fb3f0';
const COLOR_B = '#c8a24a';
const MUTED = '#8e948f';

const ITERATIONS: Array<{ kind: IterationKind; label: string; tooltip: string }> = [
	{ kind: 'typical', label: 'Typical', tooltip: 'The iteration of each run with the DPS closest to its average.' },
	{ kind: 'lucky', label: 'Lucky', tooltip: 'The iteration of each run with the highest DPS.' },
	{ kind: 'unlucky', label: 'Unlucky', tooltip: 'The iteration of each run with the lowest DPS.' },
];

// The DPS, resources and threat of both runs over the fight, like the Events tab's chart.
//
// One iteration of a run is chance more than anything, so we pick the same kind of iteration of
// each run: the typical ones, the lucky ones or the unlucky ones. The DPS is over the last 15
// seconds, like in the Events tab. Mana shows as a share of the pool, so runs with different
// gear compare.
export class OverTimeView {
	readonly rootElem: HTMLElement;
	readonly controlsElem: HTMLElement;
	private readonly bodyElem: HTMLElement;
	private readonly buttons: HTMLButtonElement[] = [];
	private iteration: IterationKind = 'typical';
	private seriesA: RunSeries | null = null;
	private seriesB: RunSeries | null = null;
	private showThreat = false;
	private charts: Chart[] = [];

	constructor() {
		this.controlsElem = (
			<div className="segmented" attributes={{ role: 'group', 'aria-label': 'Iteration' }}>
				{ITERATIONS.map(({ kind, label, tooltip }) => {
					const button = (
						<button className="segmented-option" type="button" dataset={{ mode: kind, tippyContent: tooltip }}>
							{label}
						</button>
					) as HTMLButtonElement;
					button.addEventListener('click', () => {
						this.iteration = kind;
						this.render();
					});
					this.buttons.push(button);
					return button;
				})}
			</div>
		) as HTMLElement;
		this.bodyElem = (<div className="cmp-over-time" />) as HTMLElement;
		this.rootElem = this.bodyElem;
	}

	setSeries(seriesA: RunSeries | null, seriesB: RunSeries | null, showThreat: boolean) {
		this.seriesA = seriesA;
		this.seriesB = seriesB;
		this.showThreat = showThreat;
		this.render();
	}

	// Whether both runs have their series. They come a moment after a run, see SimHistory.
	get isComplete(): boolean {
		return !!this.seriesA?.typical && !!this.seriesB?.typical;
	}

	private render() {
		this.charts.forEach(chart => chart.destroy());
		this.charts = [];
		this.buttons.forEach(button => {
			const kind = button.dataset.mode as IterationKind;
			button.classList.toggle('active', kind === this.iteration);
			button.disabled = !this.seriesA?.[kind] || !this.seriesB?.[kind];
		});

		const a = this.seriesA?.[this.iteration];
		const b = this.seriesB?.[this.iteration];
		if (!a || !b) {
			const text =
				!this.seriesA || !this.seriesB
					? 'One of the runs ran before we kept its DPS and resources over the fight. Run the sim again to see them here.'
					: 'One of the runs has no such iteration, like a run of a single iteration. Pick another one.';
			this.bodyElem.replaceChildren(<p className="cmp-note">{text}</p>);
			return;
		}

		const sections: Array<{ title: string; canvas: HTMLCanvasElement; draw: (canvas: HTMLCanvasElement) => Chart; tall?: boolean }> = [];
		sections.push({ title: 'DPS', canvas: document.createElement('canvas'), draw: canvas => dpsChart(canvas, a, b), tall: true });
		const types = Array.from(new Set([...a.resources, ...b.resources].map(resource => resource.type)));
		types.forEach(type => {
			const name = resourceNames.get(type) ?? 'Resource';
			sections.push({
				title: type === ResourceType.ResourceTypeMana ? `${name} (% of the pool)` : name,
				canvas: document.createElement('canvas'),
				draw: canvas => resourceChart(canvas, type, a, b),
			});
		});
		if (this.showThreat) sections.push({ title: 'Threat', canvas: document.createElement('canvas'), draw: canvas => threatChart(canvas, a, b) });

		this.bodyElem.replaceChildren(
			<div className="cmp-legend">
				<span className="cmp-legend-item cmp-legend-a">
					<span className="cmp-legend-swatch" />
					{`Baseline · ${formatFixed(1)(a.dps)} DPS over ${formatFixed(1)(a.duration)} s`}
				</span>
				<span className="cmp-legend-item cmp-legend-b">
					<span className="cmp-legend-swatch" />
					{`Compared · ${formatFixed(1)(b.dps)} DPS over ${formatFixed(1)(b.duration)} s`}
				</span>
			</div>,
			...sections.map(section => (
				<div className="cmp-over-time-section">
					<span className="cmp-over-time-title">{section.title}</span>
					<div className={`cmp-over-time-chart ${section.tall ? 'tall' : ''}`}>{section.canvas}</div>
				</div>
			)),
		);
		this.charts = sections.map(section => section.draw(section.canvas));
	}
}

const points = (values: number[], scale = 1) => values.map((value, i) => ({ x: i * SERIES_STEP, y: value * scale }));

// The charts share one time axis from 0 to the longer of the two fights, so they line up.
function lineChart(
	canvas: HTMLCanvasElement,
	a: IterationSeries,
	b: IterationSeries,
	dataA: Array<{ x: number; y: number }>,
	dataB: Array<{ x: number; y: number }>,
	options: { stepped?: boolean; yMax?: number; format: (value: number) => string },
): Chart {
	const dataset = (label: string, data: Array<{ x: number; y: number }>, color: string) => ({
		label,
		data,
		borderColor: color,
		backgroundColor: color,
		borderWidth: 1.75,
		pointRadius: 0,
		pointHoverRadius: 3,
		stepped: options.stepped ? ('before' as const) : false,
		tension: 0,
	});
	return new Chart(canvas, {
		type: 'line',
		data: { datasets: [dataset('Baseline', dataA, COLOR_A), dataset('Compared', dataB, COLOR_B)] },
		options: {
			maintainAspectRatio: false,
			animation: false,
			interaction: { mode: 'index', intersect: false },
			plugins: {
				legend: { display: false },
				tooltip: {
					callbacks: {
						title: items => `${formatToNumber(items[0].parsed.x, { maximumFractionDigits: 1 })} s`,
						label: item => `${item.dataset.label}: ${options.format(item.parsed.y ?? 0)}`,
					},
				},
			},
			scales: {
				x: {
					type: 'linear',
					min: 0,
					max: Math.ceil(Math.max(a.duration, b.duration)),
					grid: { color: 'rgba(255, 255, 255, 0.04)' },
					ticks: { color: MUTED, maxRotation: 0, autoSkipPadding: 16, callback: value => (Number.isInteger(Number(value)) ? `${value}s` : '') },
				},
				y: {
					min: 0,
					max: options.yMax,
					grid: { color: 'rgba(255, 255, 255, 0.05)' },
					border: { display: false },
					ticks: { color: MUTED, maxTicksLimit: 6, callback: value => options.format(Number(value)) },
				},
			},
		},
	});
}

function dpsChart(canvas: HTMLCanvasElement, a: IterationSeries, b: IterationSeries): Chart {
	return lineChart(canvas, a, b, points(a.dpsOverTime), points(b.dpsOverTime), { format: value => formatToNumber(value, { maximumFractionDigits: 0 }) });
}

function threatChart(canvas: HTMLCanvasElement, a: IterationSeries, b: IterationSeries): Chart {
	return lineChart(canvas, a, b, points(a.threat), points(b.threat), { format: value => formatToNumber(value, { notation: 'compact' }) });
}

function resourceChart(canvas: HTMLCanvasElement, type: ResourceType, a: IterationSeries, b: IterationSeries): Chart {
	const resourceA = a.resources.find(resource => resource.type === type);
	const resourceB = b.resources.find(resource => resource.type === type);
	const isMana = type === ResourceType.ResourceTypeMana;
	const data = (resource: typeof resourceA) => (resource ? points(resource.values, isMana && resource.max ? 100 / resource.max : 1) : []);
	return lineChart(canvas, a, b, data(resourceA), data(resourceB), {
		stepped: true,
		yMax: isMana ? 100 : undefined,
		format: value => (isMana ? `${formatToNumber(value, { maximumFractionDigits: 0 })}%` : formatToNumber(value, { maximumFractionDigits: 0 })),
	});
}
