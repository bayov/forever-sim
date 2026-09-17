import tippy, { Instance as TippyInstance } from 'tippy.js';

import { getRacials, Racial, racialIconUrl } from '../../forever/racials';
import { Player } from '../../player';
import { Ruleset } from '../../proto/api';
import { Spec } from '../../proto/common';
import { Component } from '../component';

// The four racials of the selected race, under the race picker: icon, name and a tooltip
// with the client's text. Only Forever has these, so the block is empty under Classic.
export class RacialsPicker extends Component {
	private readonly player: Player<Spec>;
	private tooltips: Array<TippyInstance> = [];

	constructor(parent: HTMLElement, player: Player<Spec>) {
		super(parent, 'racials-picker-root');
		this.player = player;

		this.render();
		const handler = () => this.render();
		player.raceChangeEmitter.on(handler);
		player.sim.rulesetChangeEmitter.on(handler);
		this.addOnDisposeCallback(() => {
			player.raceChangeEmitter.off(handler);
			player.sim.rulesetChangeEmitter.off(handler);
		});
	}

	private render() {
		this.tooltips.forEach(tooltip => tooltip.destroy());
		this.tooltips = [];
		this.rootElem.replaceChildren();

		if (this.player.sim.getRuleset() !== Ruleset.RulesetForever) {
			return;
		}

		const racials = getRacials(this.player.getRace(), this.player.getClass());
		const actives = racials.filter(racial => !racial.passive);
		const passives = racials.filter(racial => racial.passive);

		this.rootElem.appendChild(
			<>
				<label className="form-label">Racials</label>
				<div className="racials-picker-groups">
					{this.renderGroup('Active', actives)}
					{this.renderGroup('Passive', passives)}
				</div>
			</>,
		);
	}

	private renderGroup(title: string, racials: Array<Racial>): HTMLElement {
		const group = (
			<div className="racials-picker-group">
				<span className="racials-picker-group-title">{title}</span>
			</div>
		) as HTMLElement;

		racials.forEach(racial => {
			const entry = (
				<div className="racials-picker-racial">
					<span className="racials-picker-icon" style={{ backgroundImage: `url('${racialIconUrl(racial.icon)}')` }}></span>
					<span className="racials-picker-name">{racial.name}</span>
				</div>
			) as HTMLElement;
			group.appendChild(entry);

			this.tooltips.push(
				tippy(entry, {
					content: (
						<div className="racials-picker-tooltip">
							<span className="racials-picker-tooltip-name">{racial.name}</span>
							{racial.cost && <span className="racials-picker-tooltip-cost">{racial.cost}</span>}
							{racial.passive && <span className="racials-picker-tooltip-cost">Passive</span>}
							<p className="racials-picker-tooltip-description">{racial.description}</p>
						</div>
					),
				}),
			);
		});

		return group;
	}
}
