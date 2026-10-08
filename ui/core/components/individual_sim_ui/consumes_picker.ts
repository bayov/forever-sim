import tippy from 'tippy.js';

import { IndividualSimUI } from '../../individual_sim_ui';
import { Player } from '../../player';
import { Spec, Stat } from '../../proto/common';
import { TypedEvent } from '../../typed_event';
import { Component } from '../component';
import { IconEnumPicker } from '../icon_enum_picker';
import { buildIconInput } from '../icon_inputs.js';
import { IconPicker } from '../icon_picker';
import * as ConsumablesInputs from '../inputs/consumables';
import { relevantStatOptions } from '../inputs/stat_options';
import { MultiIconPicker } from '../multi_icon_picker';

export class ConsumesPicker extends Component {
	protected simUI: IndividualSimUI<Spec>;

	constructor(parentElem: HTMLElement, simUI: IndividualSimUI<Spec>) {
		super(parentElem, 'consumes-picker-root');
		this.simUI = simUI;

		this.simUI.sim.waitForInit().then(() => {
			this.buildPotionsPicker();
			this.buildConjuredPicker();
			this.buildFlaskPicker();
			this.buildWeaponImbuePicker();
			this.buildFoodPicker();
			this.buildPhysicalBuffPickers();
			this.buildDefensiveBuffPickers();
			this.buildSpellPowerBuffPickers();
			this.buildMiscConsumesPickers();
			this.buildEngPickers();
			this.buildPetPicker();

			// The dropdowns open 6px below their slot instead of over its bottom edge.
			this.rootElem.querySelectorAll<HTMLElement>('[data-bs-toggle=dropdown]').forEach(toggle => (toggle.dataset.bsOffset = '0,6'));
		});
	}

	private buildPotionsPicker() {
		const [row, potionsElem] = this.buildGroup('Potion');

		const potionsOptions = ConsumablesInputs.makePotionsInput(relevantStatOptions(ConsumablesInputs.POTIONS_CONFIG, this.simUI));

		const pickers = [buildIconInput(potionsElem, this.simUI.player, potionsOptions)];

		TypedEvent.onAny([this.simUI.player.professionChangeEmitter]).on(() => this.updateRow(row, pickers));
		this.updateRow(row, pickers);
	}

	// Runes, Thistle Tea and the healthstones don't share the potions' cooldown, so they get their
	// own slot.
	private buildConjuredPicker() {
		const [row, conjuredElem] = this.buildGroup('Special', 'Runes, Thistle Tea and healthstones. They have their own cooldown, apart from potions.');

		const conjuredOptions = ConsumablesInputs.makeConjuredInput(relevantStatOptions(ConsumablesInputs.CONJURED_CONFIG, this.simUI));

		const pickers = [buildIconInput(conjuredElem, this.simUI.player, conjuredOptions)];

		TypedEvent.onAny([this.simUI.player.professionChangeEmitter]).on(() => this.updateRow(row, pickers));
		this.updateRow(row, pickers);
	}

	private buildFlaskPicker() {
		const [row, flasksElem] = this.buildGroup('Flask');

		const flasksOptions = ConsumablesInputs.makeFlasksInput(relevantStatOptions(ConsumablesInputs.FLASKS_CONFIG, this.simUI));

		const pickers = [buildIconInput(flasksElem, this.simUI.player, flasksOptions)];

		TypedEvent.onAny([this.simUI.player.professionChangeEmitter]).on(() => this.updateRow(row, pickers));
		this.updateRow(row, pickers);
	}

	private buildWeaponImbuePicker() {
		const [row, imbuesElem] = this.buildGroup('Weapon');

		const mhImbueOptions = ConsumablesInputs.makeMainHandImbuesInput(relevantStatOptions(ConsumablesInputs.WEAPON_IMBUES_MH_CONFIG, this.simUI));
		const ohImbueOptions = ConsumablesInputs.makeOffHandImbuesInput(relevantStatOptions(ConsumablesInputs.WEAPON_IMBUES_OH_CONFIG, this.simUI), 'Off-Hand');

		const pickers = [buildIconInput(imbuesElem, this.simUI.player, mhImbueOptions), buildIconInput(imbuesElem, this.simUI.player, ohImbueOptions)];

		TypedEvent.onAny([this.simUI.player.gearChangeEmitter, this.simUI.player.raceChangeEmitter]).on(() => this.updateRow(row, pickers));
		this.updateRow(row, pickers);
	}

	private buildFoodPicker() {
		const [row, foodsElem] = this.buildGroup('Food');

		const foodOptions = ConsumablesInputs.makeFoodInput(relevantStatOptions(ConsumablesInputs.FOOD_CONFIG, this.simUI));
		const alcoholOptions = ConsumablesInputs.makeAlcoholInput(relevantStatOptions(ConsumablesInputs.ALCOHOL_CONFIG, this.simUI));

		const pickers = [
			buildIconInput(foodsElem, this.simUI.player, foodOptions),
			buildIconInput(foodsElem, this.simUI.player, ConsumablesInputs.DragonBreathChili),
			buildIconInput(foodsElem, this.simUI.player, alcoholOptions),
		];

		this.updateRow(row, pickers);
	}

	private buildPhysicalBuffPickers() {
		const includeAgi = this.simUI.individualConfig.epStats.includes(Stat.StatAgility);
		const includeStr = this.simUI.individualConfig.epStats.includes(Stat.StatStrength);
		const includeHit = this.simUI.individualConfig.epStats.includes(Stat.StatMeleeHit);

		if (!includeAgi && !includeStr && !includeHit) return;

		const [row, physicalConsumesElem] = this.buildGroup('Physical');

		const apBuffOptions = ConsumablesInputs.makeAttackPowerConsumeInput(
			relevantStatOptions(ConsumablesInputs.ATTACK_POWER_CONSUMES_CONFIG, this.simUI),
			'Attack Power',
		);
		const agiBuffOptions = ConsumablesInputs.makeAgilityConsumeInput(relevantStatOptions(ConsumablesInputs.AGILITY_CONSUMES_CONFIG, this.simUI), 'Agility');
		const strBuffOptions = ConsumablesInputs.makeStrengthConsumeInput(
			relevantStatOptions(ConsumablesInputs.STRENGTH_CONSUMES_CONFIG, this.simUI),
			'Strength',
		);
		const hitConsumableOptions = ConsumablesInputs.makeHitConsumableInput(relevantStatOptions(ConsumablesInputs.HIT_CONSUMABLE_CONFIG, this.simUI), 'Hit');

		const pickers = [
			buildIconInput(physicalConsumesElem, this.simUI.player, apBuffOptions),
			buildIconInput(physicalConsumesElem, this.simUI.player, agiBuffOptions),
			buildIconInput(physicalConsumesElem, this.simUI.player, strBuffOptions),
			buildIconInput(physicalConsumesElem, this.simUI.player, hitConsumableOptions),
		];

		this.updateRow(row, pickers);
	}

	private buildDefensiveBuffPickers() {
		const [row, defensiveConsumesElem] = this.buildGroup('Defensive');

		const healthBuffOptions = ConsumablesInputs.makeHealthConsumeInput(relevantStatOptions(ConsumablesInputs.HEALTH_CONSUMES_CONFIG, this.simUI), 'Health');

		const armorBuffOptions = ConsumablesInputs.makeArmorConsumeInput(relevantStatOptions(ConsumablesInputs.ARMOR_CONSUMES_CONFIG, this.simUI), 'Armor');

		const pickers = [
			buildIconInput(defensiveConsumesElem, this.simUI.player, healthBuffOptions),
			buildIconInput(defensiveConsumesElem, this.simUI.player, armorBuffOptions),
		];

		this.updateRow(row, pickers);
	}

	private buildSpellPowerBuffPickers() {
		const [row, spellsCnsumesElem] = this.buildGroup('Spell');

		const spBuffOptions = ConsumablesInputs.makeSpellPowerConsumeInput(
			relevantStatOptions(ConsumablesInputs.SPELL_POWER_CONFIG, this.simUI),
			'Spell Damage',
		);
		const fireBuffOptions = ConsumablesInputs.makeFirePowerConsumeInput(
			relevantStatOptions(ConsumablesInputs.FIRE_POWER_CONFIG, this.simUI),
			'Fire Damage',
		);
		const frostBuffOptions = ConsumablesInputs.makeFrostPowerConsumeInput(
			relevantStatOptions(ConsumablesInputs.FROST_POWER_CONFIG, this.simUI),
			'Frost Damage',
		);
		const shadowBuffOptions = ConsumablesInputs.makeShadowPowerConsumeInput(
			relevantStatOptions(ConsumablesInputs.SHADOW_POWER_CONFIG, this.simUI),
			'Shadow Damage',
		);
		const mp5BuffOptions = ConsumablesInputs.makeMp5ConsumeInput(relevantStatOptions(ConsumablesInputs.MP5_CONFIG, this.simUI), 'Mana Regen');
		const intBuffOptions = ConsumablesInputs.makeIntellectConsumeInput(relevantStatOptions(ConsumablesInputs.INTELLECT_CONFIG, this.simUI), 'Intellect');

		const pickers = [
			buildIconInput(spellsCnsumesElem, this.simUI.player, spBuffOptions),
			buildIconInput(spellsCnsumesElem, this.simUI.player, fireBuffOptions),
			buildIconInput(spellsCnsumesElem, this.simUI.player, frostBuffOptions),
			buildIconInput(spellsCnsumesElem, this.simUI.player, shadowBuffOptions),
			buildIconInput(spellsCnsumesElem, this.simUI.player, mp5BuffOptions),
			buildIconInput(spellsCnsumesElem, this.simUI.player, intBuffOptions),
		];

		this.updateRow(row, pickers);
	}

	private buildMiscConsumesPickers() {
		const [row, miscConsumesElem] = this.buildGroup('Misc');

		const zanzaBuffOptions = ConsumablesInputs.makeZanzaBuffConsumesInput(
			relevantStatOptions(ConsumablesInputs.ZANZA_BUFF_CONSUMES_CONFIG, this.simUI),
			'Zanza Buffs',
		);
		const miscOffensiveConsumesOptions = relevantStatOptions(ConsumablesInputs.MISC_OFFENSIVE_CONSUMES_CONFIG, this.simUI);
		const miscDefensiveConsumesOptions = relevantStatOptions(ConsumablesInputs.MISC_DEFENSIVE_CONSUMES_CONFIG, this.simUI);

		const pickers = [
			buildIconInput(miscConsumesElem, this.simUI.player, zanzaBuffOptions),
			ConsumablesInputs.makeMiscOffensiveConsumesInput(miscConsumesElem, this.simUI.player, this.simUI, miscOffensiveConsumesOptions),
			ConsumablesInputs.makeMiscDefensiveConsumesInput(miscConsumesElem, this.simUI.player, this.simUI, miscDefensiveConsumesOptions),
		];

		this.updateRow(row, pickers);
	}

	private buildEngPickers() {
		const [row, engiConsumesElem] = this.buildGroup('Engineering');

		const explosiveOptions = ConsumablesInputs.makeExplosivesInput(relevantStatOptions(ConsumablesInputs.EXPLOSIVES_CONFIG, this.simUI), 'Explosives');
		const sapperOptions = ConsumablesInputs.makeSappersInput(relevantStatOptions(ConsumablesInputs.SAPPER_CONFIG, this.simUI), 'Sappers');

		const pickers = [
			buildIconInput(engiConsumesElem, this.simUI.player, sapperOptions),
			buildIconInput(engiConsumesElem, this.simUI.player, explosiveOptions),
		];

		TypedEvent.onAny([this.simUI.player.professionChangeEmitter]).on(() => this.updateRow(row, pickers));
		this.updateRow(row, pickers);
	}

	private buildPetPicker() {
		if (!this.simUI.individualConfig.petConsumeInputs?.length) return;

		const [row, petConsumesElem] = this.buildGroup('Pet');

		// const miscPetConsumesOptions = relevantStatOptions(ConsumablesInputs.MISC_PET_CONSUMES, this.simUI);

		const pickers = [
			...this.simUI.individualConfig.petConsumeInputs.map(iconInput => buildIconInput(petConsumesElem, this.simUI.player, iconInput)),
			// ConsumablesInputs.makeMiscPetConsumesInput(petConsumesElem, this.simUI.player, this.simUI, miscPetConsumesOptions),
		];

		this.updateRow(row, pickers);
	}

	// A category with its name above its slots. The categories sit side by side and wrap, so
	// several of them share a line.
	//
	// A category's tooltip shows on both its name and its slots. We give one only when the name
	// alone doesn't say what goes in it, like the Special slot.
	private buildGroup(label: string, tooltip?: string): [HTMLElement, HTMLElement] {
		const group = document.createElement('div');
		group.classList.add('consumes-group');

		const labelElem = document.createElement('label');
		labelElem.classList.add('consumes-group-label');
		labelElem.textContent = label;

		const slots = document.createElement('div');
		slots.classList.add('picker-group', 'icon-group', 'consumes-group-slots');

		group.append(labelElem, slots);
		this.rootElem.appendChild(group);
		if (tooltip) {
			const instance = tippy(group, { content: tooltip });
			this.addOnDisposeCallback(() => instance.destroy());
		}
		return [group, slots];
	}

	private updateRow(rowElem: HTMLElement, pickers: (IconPicker<Player<Spec>, any> | IconEnumPicker<Player<Spec>, any> | MultiIconPicker<Player<Spec>>)[]) {
		if (!!pickers.find(p => p?.showWhen())) {
			rowElem.classList.remove('hide');
		} else {
			rowElem.classList.add('hide');
		}
	}
}
