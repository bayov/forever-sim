import { Chart, ChartDataset } from 'chart.js';

import { AggroMetrics } from '../../proto/api.js';
import { UnitMetrics } from '../../proto_utils/sim_result.js';
import { classColors, specToClass } from '../../proto_utils/utils.js';
import { formatToPercent } from '../../utils.js';
import { ResultComponent, ResultComponentConfig, SimResultData } from './result_component.js';

const MUTED = '#8e948f';
// The members most likely to pass get a line, so a raid of hunters who each pass now and then stays readable.
const MAX_LINES = 8;

// The second of the fight by which half the passes have come.
export const medianFirstPass = (aggro: AggroMetrics): number => {
	const passes = aggro.firstPass.reduce((a, b) => a + b, 0);
	let seen = 0;
	for (let second = 0; second < aggro.firstPass.length; second++) {
		seen += aggro.firstPass[second];
		if (seen * 2 >= passes) return second;
	}
	return 0;
};

// What the Aggro column's tooltip says about a player or pet, like "Passes the tank in 12% of fights,
// half of them by 3s in".
export const aggroLine = (name: string, aggro: AggroMetrics): string =>
	aggro.chance > 0
		? `${name}: passes in ${percent(aggro.chance)} of fights, half by ${medianFirstPass(aggro)}s in`
		: `${name}: never passes, at most ${percent(aggro.peak?.max ?? 0)} of the pull`;

export const percent = (share: number) => formatToPercent(share * 100, { maximumFractionDigits: share < 0.1 ? 1 : 0 });

// A player's own aggro and their pets', for the players table.
export const unitAggros = (player: UnitMetrics): Array<{ name: string; aggro: AggroMetrics }> =>
	[player, ...player.pets].filter(unit => unit.aggro).map(unit => ({ name: unit.isPet ? unit.name : 'Self', aggro: unit.aggro! }));

// The share of fights in which each player or pet has passed the tank by each second of the fight
// (docs/raid_sim.md, Threat).
//
// The tank starts with little threat, so most passes come in the first seconds. One line per member who
// ever passes, so a raid where nobody does shows only a line of text.
export class AggroChart extends ResultComponent {
	private chart: Chart | null = null;
	private readonly chartElem: HTMLElement;
	private readonly noteElem: HTMLElement;

	constructor(config: ResultComponentConfig) {
		config.rootCssClass = 'aggro-chart-root';
		super(config);
		this.noteElem = (<p className="aggro-chart-note" />) as HTMLElement;
		this.chartElem = (<div className="aggro-chart-canvas" />) as HTMLElement;
		this.rootElem.replaceChildren(this.noteElem, this.chartElem);
	}

	onSimResult(resultData: SimResultData) {
		const result = resultData.result;
		const players = result.getPlayers(resultData.filter);
		const row = this.rootElem.closest<HTMLElement>('.dr-row');
		const checked = players.some(player => player.aggro);
		if (row) row.hidden = !checked;
		this.chart?.destroy();
		this.chart = null;
		this.chartElem.replaceChildren();
		if (!checked) return;

		const iterations = result.iterations;
		const seconds = Math.ceil(result.duration);
		const passing = players.flatMap(player => [player, ...player.pets]).filter(unit => unit.aggro && unit.aggro.chance > 0);
		const units = passing.sort((a, b) => b.aggro!.chance - a.aggro!.chance).slice(0, MAX_LINES);
		const tps = (players[0]?.aggro && result.request.raid?.threatCheck?.tankTps) || 0;
		const headStart = result.request.raid?.threatCheck?.headStart ?? 0;
		const tank = `tank ${tps} TPS, ${headStart}s head start`;
		this.noteElem.textContent = units.length
			? [`Share of fights passed by then`, tank, ...(passing.length > units.length ? [`top ${units.length} of ${passing.length}`] : [])].join(' · ')
			: `Nobody passes the tank · ${tank}`;
		if (!units.length) return;

		const datasets: Array<ChartDataset<'line'>> = units.map(unit => {
			const owner = unit.isPet ? result.getPlayerWithRaidIndex(unit.index) : unit;
			const color = owner ? classColors[specToClass[owner.spec]] : MUTED;
			let passed = 0;
			const data = [...Array(seconds + 1).keys()].map(second => {
				passed += unit.aggro!.firstPass[second] ?? 0;
				return { x: second, y: (passed / iterations) * 100 };
			});
			return {
				label: unit.isPet ? `${owner?.label ?? ''} - ${unit.name}` : unit.label,
				data,
				borderColor: color,
				backgroundColor: color,
				borderDash: unit.isPet ? [4, 3] : [],
				borderWidth: 2,
				pointRadius: 0,
				pointHoverRadius: 3,
				stepped: 'after',
			};
		});

		const canvas = document.createElement('canvas');
		this.chartElem.appendChild(canvas);
		this.chart = new Chart(canvas, {
			type: 'line',
			data: { datasets },
			options: {
				maintainAspectRatio: false,
				animation: false,
				interaction: { mode: 'nearest', axis: 'x', intersect: false },
				plugins: {
					legend: { display: units.length > 1, labels: { color: MUTED, boxWidth: 12, boxHeight: 2 } },
					tooltip: {
						callbacks: {
							title: items => `${items[0].parsed.x}s`,
							label: item => `${item.dataset.label}: ${formatToPercent(item.parsed.y ?? 0, { maximumFractionDigits: 1 })}`,
						},
					},
				},
				scales: {
					x: {
						type: 'linear',
						min: 0,
						max: seconds,
						grid: { display: false },
						ticks: { color: MUTED, maxRotation: 0, autoSkipPadding: 16, callback: value => `${value}s` },
					},
					y: {
						min: 0,
						suggestedMax: 10,
						grid: { color: 'rgba(255, 255, 255, 0.05)' },
						border: { display: false },
						ticks: { color: MUTED, callback: value => `${value}%` },
					},
				},
			},
		});
	}
}
