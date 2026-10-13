import { Chart } from 'chart.js';

import { ResultComponent, ResultComponentConfig, SimResultData } from './result_component.js';

export class DpsHistogram extends ResultComponent {
	constructor(config: ResultComponentConfig) {
		config.rootCssClass = 'dps-histogram-root';
		super(config);
	}

	onSimResult(resultData: SimResultData) {
		const chartBounds = this.rootElem.getBoundingClientRect();

		this.rootElem.textContent = '';
		const chartCanvas = document.createElement('canvas');
		chartCanvas.height = chartBounds.height;
		chartCanvas.width = chartBounds.width;

		const damageMetrics = resultData.result.getDamageMetrics(resultData.filter);

		// Bars within one standard deviation of the average are solid, the tails are faint.
		const min = damageMetrics.avg - damageMetrics.stdev;
		const max = damageMetrics.avg + damageMetrics.stdev;
		const vals: Array<number> = [];
		const colors: Array<string> = [];

		const labels = Object.keys(damageMetrics.hist);
		labels.forEach(k => {
			vals.push(damageMetrics.hist[Number(k)]);
			const val = parseInt(k);
			colors.push(val > min && val < max ? 'rgba(127, 179, 240, 0.85)' : 'rgba(127, 179, 240, 0.3)');
		});

		const ctx = chartCanvas.getContext('2d')!;
		this.rootElem.appendChild(chartCanvas);

		const mutedColor = '#8e948f';
		const chart = new Chart(ctx, {
			type: 'bar',
			data: {
				labels: labels,
				datasets: [
					{
						data: vals,
						backgroundColor: colors,
						hoverBackgroundColor: '#c8a24a',
						borderRadius: 2,
						categoryPercentage: 0.9,
						barPercentage: 1,
					},
				],
			},
			options: {
				maintainAspectRatio: false,
				plugins: {
					title: {
						display: false,
					},
					legend: {
						display: false,
					},
					tooltip: {
						displayColors: false,
						callbacks: {
							title: items => `${items[0].label} DPS`,
							label: item => `${item.formattedValue} iterations`,
						},
					},
				},
				scales: {
					x: {
						grid: {
							display: false,
						},
						ticks: {
							color: mutedColor,
							maxRotation: 0,
							autoSkipPadding: 16,
						},
					},
					y: {
						beginAtZero: true,
						grid: {
							color: 'rgba(255, 255, 255, 0.05)',
						},
						border: {
							display: false,
						},
						ticks: {
							display: false,
						},
					},
				},
			},
		});

		this.addOnDisposeCallback(() => chart.destroy());
	}
}
