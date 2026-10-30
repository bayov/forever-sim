import { ref } from 'tsx-vanilla';

import { IndividualSimUI } from '../../individual_sim_ui';
import { Player } from '../../player';
import { Spec } from '../../proto/common';
import { LevelingRankWorth } from '../../proto/ui';
import { ActionId } from '../../proto_utils/action_id';
import { ContentBlock } from '../content_block';

// A buy level past the plan's last level is a rank we never train (tools/levelingplan/training.go).
const NEVER_LEVEL = 61;

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
		this.statusElem.textContent = parts.join(' · ');

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

// A price as the game says it, like 1.08g, 18s or 5c.
function formatCopper(copper: number): string {
	if (copper >= 10000) return `${(copper / 10000).toFixed(2)}g`;
	if (copper >= 100) return `${Math.round(copper / 100)}s`;
	return `${copper}c`;
}
