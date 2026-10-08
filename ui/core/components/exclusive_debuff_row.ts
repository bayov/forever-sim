import tippy from 'tippy.js';

import { Player } from '../player';
import { Debuffs, TristateEffect } from '../proto/common';
import { TypedEvent } from '../typed_event';
import { Component } from './component';
import { IconPicker, IconPickerConfig } from './icon_picker';

// The debuffs we turn on with one icon, whether the field is a plain on or off or a tristate.
export type DebuffToggleField =
	| 'sunderArmor'
	| 'exposeArmor'
	| 'curseOfRecklessness'
	| 'faerieFire'
	| 'demoralizingShout'
	| 'demoralizingRoar'
	| 'thunderClap'
	| 'thunderfury'
	| 'judgementOfTheCrusader'
	| 'huntersMark';

// A tristate debuff counts as on when it's regular or improved.
export function isDebuffOn(debuffs: Debuffs, field: DebuffToggleField): boolean {
	const value = debuffs[field];
	return typeof value === 'number' ? value !== TristateEffect.TristateEffectMissing : value;
}

// Turns a debuff on or off. A tristate debuff turns on as the regular one.
export function setDebuffOn(debuffs: Debuffs, field: DebuffToggleField, on: boolean) {
	if (typeof debuffs[field] === 'number') {
		(debuffs[field] as TristateEffect) = on ? TristateEffect.TristateEffectRegular : TristateEffect.TristateEffectMissing;
	} else {
		(debuffs[field] as boolean) = on;
	}
}

export interface ExclusiveDebuffRowConfig {
	fields: Array<DebuffToggleField>;
	// One icon per field, in the same order. Turning one on turns the others off.
	options: Array<IconPickerConfig<Player<any>, boolean>>;
	// When the debuffs stack after all, like Faerie Fire and Curse of Recklessness in Classic.
	// Then the icons work on their own and the row has no empty slot.
	exclusiveWhen: (player: Player<any>) => boolean;
}

// A row of debuffs that don't stack, so we pick at most one of them, like Sunder Armor and
// Expose Armor.
//
// The row starts with an empty slot that's lit when none of them is on. Clicking it turns them
// all off.
export class ExclusiveDebuffRow extends Component {
	constructor(parent: HTMLElement, player: Player<any>, config: ExclusiveDebuffRowConfig) {
		super(parent, 'exclusive-debuff-row');
		const raid = player.getRaid()!;

		const none = document.createElement('div');
		none.classList.add('icon-picker', 'exclusive-debuff-none');
		const noneButton = document.createElement('a');
		noneButton.classList.add('icon-picker-button');
		none.appendChild(noneButton);
		this.rootElem.appendChild(none);
		tippy(noneButton, { content: 'None' });
		noneButton.addEventListener('click', event => {
			event.preventDefault();
			const debuffs = raid.getDebuffs();
			config.fields.forEach(field => setDebuffOn(debuffs, field, false));
			raid.setDebuffs(TypedEvent.nextEventID(), debuffs);
		});

		config.options.forEach(option => new IconPicker(this.rootElem, player, option));

		const update = () => {
			const debuffs = raid.getDebuffs();
			noneButton.classList.toggle('active', !config.fields.some(field => isDebuffOn(debuffs, field)));
			none.classList.toggle('hide', !config.exclusiveWhen(player));
		};
		update();
		TypedEvent.onAny([raid.debuffsChangeEmitter, player.sim.rulesetChangeEmitter]).on(update);
	}
}
