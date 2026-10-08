import tippy, { Instance as TippyInstance } from 'tippy.js';

import { getRacials, Racial, racialIconUrl } from '../../forever/racials';
import { Player } from '../../player';
import { Ruleset } from '../../proto/api';
import { Spec } from '../../proto/common';
import { Component } from '../component';
import { gameSpellTooltip } from '../game_spell_tooltip';

// The four racials of the selected race, on a piece of parchment like the spellbook in the
// game. The actives are in the left column with a square frame, and the passives are in the
// right one with a round frame. Each has the game's tooltip. Only Forever has these, so the
// block is empty under Classic.
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
		const parchment = (<div className="racials-picker-parchment"></div>) as HTMLElement;
		[...racials.filter(racial => !racial.passive), ...racials.filter(racial => racial.passive)].forEach(racial => {
			const entry = (
				<div className={`racials-picker-racial ${racial.passive ? 'passive' : 'active'}`}>
					<span className="racials-picker-frame">
						<span className="racials-picker-icon" style={{ backgroundImage: `url('${racialIconUrl(racial.icon)}')` }}></span>
					</span>
					<span className="racials-picker-name">{racial.name}</span>
				</div>
			) as HTMLElement;
			parchment.appendChild(entry);
			this.tooltips.push(tippy(entry, { theme: 'game-spell', content: spellTooltip(racial) }));
		});

		this.rootElem.appendChild(parchment);
	}
}

// The client data writes the cost line as one string, like "5 yd range; Instant; 2 min
// cooldown", so we split it. The range goes on its own line on the right, like in the game,
// then the cast time with the cooldown on its right (or Passive).
function spellTooltip(racial: Racial): HTMLElement {
	const parts = racial.cost?.split('; ') ?? [];
	const range = parts.find(part => part.endsWith('range'));
	const cooldown = parts.find(part => part.endsWith('cooldown'));
	const cast = parts.find(part => part !== range && part !== cooldown);

	const lines: Array<[string, string?]> = [];
	if (range) lines.push(['', range]);
	lines.push(racial.passive ? ['Passive'] : [cast ?? '', cooldown]);
	return gameSpellTooltip({ name: racial.name, lines, description: racial.description });
}
