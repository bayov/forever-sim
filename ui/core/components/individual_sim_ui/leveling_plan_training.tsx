import { ref } from 'tsx-vanilla';

import { IndividualSimUI } from '../../individual_sim_ui';
import { Player } from '../../player';
import { Spec } from '../../proto/common';
import { LevelingRankWorth, LevelingTrainingSettings } from '../../proto/ui';
import { TypedEvent } from '../../typed_event';
import { ActionId } from '../../proto_utils/action_id';
import { ContentBlock } from '../content_block';

// A buy level past the plan's last level is a rank we never train (tools/levelingplan/training.go).
const NEVER_LEVEL = 61;
// The budget as training.go has it: the gold earned by a level grows with the level to this power,
// and we keep Training's flight money (BayovCore's RESERVE) after training.
const GOLD_GROWTH = 3.5;
const FLIGHT_RESERVE: Array<[number, number]> = [
	[50, 20000],
	[40, 10000],
	[30, 5000],
	[20, 2500],
	[11, 1000],
	[1, 100],
];

// The Training part of the Leveling Plan tab: when we train each spell rank the plan casts, what
// it costs and the minutes it saves.
//
// The training search (tools/levelingplan -training) makes it. It measures each rank against
// the rank below it at the levels it's our highest, and it skips the last ranks of a spell when
// they save no time in all. Where the gold budget runs short, it trains the ranks that save the
// least for their gold later. The sims of the plan take the schedule, so a rank we train later
// isn't cast before then.
export class LevelingTrainingSection {
	readonly rootElem: HTMLElement;
	private readonly player: Player<Spec>;
	private readonly statusElem: HTMLElement;
	private readonly bodyElem: HTMLElement;
	private readonly othersElem: HTMLElement;

	constructor(parentElem: HTMLElement, simUI: IndividualSimUI<Spec>) {
		this.player = simUI.player;
		this.rootElem = (<div className="leveling-plan-training" />) as HTMLElement;
		parentElem.appendChild(this.rootElem);

		const block = new ContentBlock(this.rootElem, 'leveling-plan-training-block', {
			header: {
				title: 'Training',
				tooltip: 'When we train each rank the plan casts. Saves is the time it saves over the levels it is our best rank.',
			},
		});
		const status = ref<HTMLDivElement>();
		const body = ref<HTMLTableSectionElement>();
		block.bodyElement.appendChild(
			<>
				<div className="leveling-plan-training-status" ref={status} />
				<div className="leveling-plan-table-wrap">
					<table className="leveling-plan-table leveling-plan-training-table table">
						<thead>
							<tr>
								<th>Level</th>
								<th>Rank</th>
								<th>Price</th>
								<th>Saves</th>
								<th>Train</th>
							</tr>
						</thead>
						<tbody ref={body} />
					</table>
				</div>
			</>,
		);
		this.statusElem = status.value!;
		this.bodyElem = body.value!;

		const othersBlock = new ContentBlock(this.rootElem, 'leveling-plan-training-others-block', {
			header: {
				title: 'Other spells',
				tooltip: "Spells the plan never casts. A ticked one is trained at each rank's level, and its gold counts in the budget.",
			},
		});
		const others = ref<HTMLTableSectionElement>();
		othersBlock.bodyElement.appendChild(
			<div className="leveling-plan-table-wrap">
				<table className="leveling-plan-table leveling-plan-training-table table">
					<thead>
						<tr>
							<th>Always</th>
							<th>Spell</th>
							<th>Levels</th>
							<th>By 40</th>
							<th>To 60</th>
						</tr>
					</thead>
					<tbody ref={others} />
				</table>
			</div>,
		);
		this.othersElem = others.value!;
		this.player.levelingPlanChangeEmitter.on(() => this.render());
		this.render();
	}

	private render() {
		const training = this.player.readLevelingPlan().training;
		const worth = [...(training?.worth ?? [])].sort((a, b) => a.level - b.level || a.spellId - b.spellId);
		const buys = new Map((training?.buys ?? []).map(buy => [buy.spellId, buy.level]));
		this.bodyElem.replaceChildren();
		if (!training || !worth.length) {
			this.statusElem.textContent = 'Every rank at its level';
			return;
		}

		const never = [...buys.values()].filter(level => level >= NEVER_LEVEL).length;
		const later = buys.size - never;
		const parts = [`${training.goldAt40}g by 40`];
		if (training.ridingLevel) parts.push(`riding ${training.ridingGold}g at ${training.ridingLevel}`);
		parts.push(`${never} skipped`, `${later} later`);
		const short = shortest(training);
		parts.push(short ? `${formatCopper(short.copper)} short at ${short.level}` : 'gold enough');
		this.statusElem.textContent = parts.join(' · ');
		this.statusElem.classList.toggle('leveling-plan-training-short', !!short);
		this.renderOthers(training);

		worth.forEach(w => {
			const row = this.row(w, buys.get(w.spellId));
			this.bodyElem.appendChild(row.elem);
			ActionId.fromSpellId(w.spellId)
				.fill()
				.then(id => {
					row.icon.style.backgroundImage = `url('${id.iconUrl}')`;
					row.name.textContent = `${id.name.replace(/ \(Rank \d+\)$/, '')} ${w.rank}`;
				});
		});
	}

	// A row per spell the plan never casts, with its ranks' levels and prices.
	private renderOthers(training: LevelingTrainingSettings) {
		this.othersElem.replaceChildren();
		const bySpell = new Map<string, LevelingRankWorth[]>();
		for (const w of training.others) bySpell.set(w.spell, [...(bySpell.get(w.spell) ?? []), w]);
		const spells = [...bySpell.entries()].map(([spell, ranks]) => ({ spell, ranks: ranks.sort((a, b) => a.level - b.level) }));
		spells.sort((a, b) => a.ranks[0].level - b.ranks[0].level || a.spell.localeCompare(b.spell));
		const always = new Set(training.alwaysSpells);
		for (const { spell, ranks } of spells) {
			const icon = ref<HTMLSpanElement>();
			const box = ref<HTMLInputElement>();
			const sum = (to: number) => ranks.filter(r => r.level <= to).reduce((total, r) => total + r.price, 0);
			this.othersElem.appendChild(
				<tr className={always.has(spell) ? '' : 'leveling-plan-training-never'}>
					<td>
						<input type="checkbox" className="form-check-input" ref={box} attributes={{ 'aria-label': spell }} />
					</td>
					<td>
						<span className="leveling-plan-spell" ref={icon} />
						{spell}
					</td>
					<td>{ranks.map(r => r.level).join(', ')}</td>
					<td>{formatCopper(sum(40))}</td>
					<td>{formatCopper(sum(60))}</td>
				</tr>,
			);
			// A checked attribute of false would still tick it, so we set the property.
			box.value!.checked = always.has(spell);
			box.value!.addEventListener('change', () => this.setAlways(spell, box.value!.checked));
			ActionId.fromSpellId(ranks[0].spellId)
				.fill()
				.then(id => (icon.value!.style.backgroundImage = `url('${id.iconUrl}')`));
		}
	}

	private setAlways(spell: string, on: boolean) {
		const plan = this.player.getLevelingPlan();
		const training = plan.training!;
		training.alwaysSpells = training.alwaysSpells.filter(s => s != spell);
		if (on) training.alwaysSpells.push(spell);
		this.player.setLevelingPlan(TypedEvent.nextEventID(), plan);
	}

	private row(w: LevelingRankWorth, buy: number | undefined) {
		const icon = ref<HTMLSpanElement>();
		const name = ref<HTMLSpanElement>();
		let train = `${w.level}`;
		let trainClass = '';
		if (buy !== undefined && buy >= NEVER_LEVEL) {
			train = 'Never';
			trainClass = 'leveling-plan-training-never';
		} else if (buy !== undefined) {
			train = `${buy}`;
			trainClass = 'leveling-plan-training-later';
		}
		const elem = (
			<tr className={trainClass}>
				<td>{`${w.level}`}</td>
				<td>
					<span className="leveling-plan-spell" ref={icon} />
					<span ref={name}>{`${w.spellId}`}</span>
				</td>
				<td>{formatCopper(w.price)}</td>
				<td className={w.minutes < 0 ? 'leveling-plan-training-costs' : ''}>{`${w.minutes > 0 ? '+' : ''}${w.minutes.toFixed(1)} min`}</td>
				<td>{train}</td>
			</tr>
		) as HTMLElement;
		return { elem, icon: icon.value!, name: name.value! };
	}
}

// The first level where the gold runs short under the schedule, and by how much in copper, or
// null when it never does. Like training.go's short: the ranks bought by a level, the spells we
// always train, riding and the flight money, against the gold earned.
function shortest(training: LevelingTrainingSettings): { level: number; copper: number } | null {
	const buys = new Map(training.buys.map(buy => [buy.spellId, buy.level]));
	const always = new Set(training.alwaysSpells);
	for (let level = 10; level <= 60; level++) {
		let spent = FLIGHT_RESERVE.find(([from]) => level >= from)?.[1] ?? 0;
		for (const w of training.worth) if ((buys.get(w.spellId) ?? w.level) <= level) spent += w.price;
		for (const w of training.others) if (always.has(w.spell) && w.level <= level) spent += w.price;
		if (training.ridingLevel && level >= training.ridingLevel) spent += training.ridingGold * 10000;
		const short = spent - training.goldAt40 * 10000 * Math.pow(level / 40, GOLD_GROWTH);
		if (short > 0) return { level, copper: Math.round(short) };
	}
	return null;
}

// A price as the game says it, like 1.08g, 18s or 5c.
function formatCopper(copper: number): string {
	if (copper >= 10000) return `${(copper / 10000).toFixed(2)}g`;
	if (copper >= 100) return `${Math.round(copper / 100)}s`;
	return `${copper}c`;
}
