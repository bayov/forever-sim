import { Chart } from 'chart.js';

import { Component } from '../../components/component.js';
import { ActionMetrics } from '../../proto_utils/sim_result.js';
import { sum } from '../../utils.js';
import { actionColors } from './color_settings.js';

// Each action's share of a player's damage, or of their threat, as a pie.
export class SourceChart extends Component {
	constructor(parentElem: HTMLElement, allActionMetrics: Array<ActionMetrics>, metric: 'damage' | 'threat' = 'damage') {
		const chartCanvas = document.createElement('canvas');
		super(parentElem, 'source-chart-root', chartCanvas);

		chartCanvas.style.height = '400px';
		chartCanvas.style.width = '600px';
		chartCanvas.height = 400;
		chartCanvas.width = 600;

		const amount = (actionMetric: ActionMetrics) => (metric == 'threat' ? actionMetric.threat : actionMetric.damage);
		const actionMetrics = allActionMetrics.filter(actionMetric => amount(actionMetric) > 0).sort((a, b) => amount(b) - amount(a));
		const names = actionMetrics.map(am => am.name);
		const total = sum(actionMetrics.map(amount));
		const vals = actionMetrics.map(actionMetric => amount(actionMetric) / total);
		const bgColors = actionColors.slice(0, actionMetrics.length);

		const ctx = chartCanvas.getContext('2d')!;
		const chart = new Chart(ctx, {
			type: 'pie',
			data: {
				labels: names,
				datasets: [
					{
						data: vals,
						backgroundColor: bgColors,
					},
				],
			},
			options: {
				plugins: {
					legend: {
						display: true,
						position: 'right',
					},
				},
			},
		});

		this.addOnDisposeCallback(() => chart.destroy());
	}
}
