import { ArcElement, Chart, DoughnutController } from 'chart.js';
import tippy from 'tippy.js';

import { Component } from '../../components/component.js';
import { ActionMetrics } from '../../proto_utils/sim_result.js';
import { formatToCompactNumber, formatToPercent, sum } from '../../utils.js';

Chart.register(DoughnutController, ArcElement);

// The slices' colors, picked to read on the dark tooltip. The first is the page's gold, so the
// biggest source matches the accent of the rest of the results.
const SLICE_COLORS = ['#dd9933', '#4fb3bf', '#6c8ef0', '#b07cd8', '#e06c75', '#8cc265', '#d4c06a', '#5fa8d3', '#c98bb0'];
const OTHER_COLOR = '#5c6370';

// Each action's share of a player's damage, or of their threat, as a ring with a legend.
//
// The legend lists the biggest sources with their icon, share and amount, and folds the rest into
// one "Other" slice, so a hunter with a dozen small procs still fits in one tooltip.
export class SourceChart extends Component {
	constructor(parentElem: HTMLElement, allActionMetrics: Array<ActionMetrics>, metric: 'damage' | 'threat' = 'damage') {
		const canvas = document.createElement('canvas');
		canvas.width = canvas.height = 132;
		super(parentElem, 'source-chart-root');

		const amount = (action: ActionMetrics) => (metric == 'threat' ? action.avgThreat : action.avgDamage);
		const actions = allActionMetrics.filter(action => amount(action) > 0).sort((a, b) => amount(b) - amount(a));
		const total = sum(actions.map(amount));
		if (!total) {
			this.rootElem.appendChild(<div className="source-chart-empty">No {metric}</div>);
			return;
		}

		const shown = actions.length > SLICE_COLORS.length ? actions.slice(0, SLICE_COLORS.length - 1) : actions;
		const rest = actions.slice(shown.length);
		const slices = [
			...shown.map((action, i) => ({ action: action as ActionMetrics | null, value: amount(action), color: SLICE_COLORS[i] })),
			...(rest.length ? [{ action: null, value: sum(rest.map(amount)), color: OTHER_COLOR }] : []),
		];

		const rows = slices.map(({ action, value, color }) => {
			const icon = (<span className="metrics-action-icon" />) as HTMLElement;
			if (action) action.actionId.setBackground(icon);
			return (
				<div className="source-chart-row">
					<span className="source-chart-swatch" style={{ backgroundColor: color }} />
					{action ? icon : <span className="metrics-action-icon source-chart-other-icon" />}
					<span className="source-chart-name">
						{action ? action.name : `Other (${rest.length})`}
						{action?.unit?.isPet ? <span className="source-chart-pet">{action.unit.name}</span> : null}
					</span>
					<span className="source-chart-percent">{formatToPercent((value / total) * 100, { maximumFractionDigits: 1 })}</span>
					<span className="source-chart-amount">{formatToCompactNumber(value)}</span>
				</div>
			);
		});

		this.rootElem.append(<div className="source-chart-ring">{canvas}</div>, <div className="source-chart-legend">{rows}</div>);

		const chart = new Chart(canvas, {
			type: 'doughnut',
			data: {
				datasets: [
					{
						data: slices.map(slice => slice.value),
						backgroundColor: slices.map(slice => slice.color),
						borderWidth: 0,
						spacing: 4,
					},
				],
			},
			options: {
				responsive: false,
				animation: false,
				cutout: '62%',
				layout: { padding: 4 },
				plugins: { legend: { display: false }, tooltip: { enabled: false } },
				events: [],
			},
		});

		this.addOnDisposeCallback(() => chart.destroy());
	}
}

// Shows a player's source chart when we point at their row. We make the chart on the first show,
// and not at all while the row is open, since its abilities table is right under it.
export const sourceChartTooltip = (rowElem: HTMLElement, makeChart: () => HTMLElement) => {
	let chart: HTMLElement | null = null;
	tippy(rowElem, {
		content: '',
		placement: 'bottom',
		theme: 'source-chart',
		maxWidth: 'none',
		delay: [250, 0],
		ignoreAttributes: true,
		onShow(instance) {
			if (rowElem.classList.contains('expanded')) return false;
			if (!chart) {
				chart = makeChart();
				instance.setContent(chart);
			}
		},
	});
};
