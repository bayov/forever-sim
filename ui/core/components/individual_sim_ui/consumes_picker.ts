import tippy, { Instance } from 'tippy.js';

import { IndividualSimUI } from '../../individual_sim_ui';
import { Player } from '../../player';
import { IntellectElixir, Spec, Stat, TristateEffect } from '../../proto/common';
import { TypedEvent } from '../../typed_event';
import { Component } from '../component';
import { IconEnumPicker } from '../icon_enum_picker';
import { buildIconInput } from '../icon_inputs.js';
import { IconPicker } from '../icon_picker';
import * as BuffDebuffInputs from '../inputs/buffs_debuffs';
import * as ConsumablesInputs from '../inputs/consumables';
import { relevantStatOptions } from '../inputs/stat_options';
import { MultiIconPicker } from '../multi_icon_picker';

export class ConsumesPicker extends Component {
	protected simUI: IndividualSimUI<Spec>;

	private intellectElixir?: IconEnumPicker<Player<Spec>, number>;
	private scrollOfStamina?: IconPicker<Player<Spec>, boolean>;
	private scrollOfIntellect?: IconPicker<Player<Spec>, boolean>;
	private scrollOfSpirit?: IconPicker<Player<Spec>, boolean>;

	constructor(parentElem: HTMLElement, simUI: IndividualSimUI<Spec>) {
		super(parentElem, 'consumes-picker-root');
		this.simUI = simUI;

		this.simUI.sim.waitForInit().then(() => {
			this.buildPotionsPicker();
			this.buildConjuredPicker();
			this.buildFlaskPicker();
			this.buildWeaponImbuePicker();
			this.buildFoodPicker();
			this.buildAttributePickers();
			this.buildPhysicalBuffPickers();
			this.buildSpellPowerBuffPickers();
			this.buildDefensiveBuffPickers();
			this.buildMiscConsumesPickers();
			this.buildEngPickers();
			this.buildPetPicker();
			this.markUnstacked();

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

		const pickers = [buildIconInput(foodsElem, this.simUI.player, foodOptions), buildIconInput(foodsElem, this.simUI.player, alcoholOptions)];

		this.updateRow(row, pickers);
	}

	// A category for each attribute, with its elixirs and its scroll. A scroll doesn't stack with
	// an elixir or a raid buff of the same attribute, see markUnstacked.
	private buildAttributePickers() {
		const player = this.simUI.player;
		const raidBuffShown = (config: unknown) => relevantStatOptions(BuffDebuffInputs.RAID_BUFFS_CONFIG, this.simUI).some(option => option.config == config);

		const [strengthRow, strengthElem] = this.buildGroup('Strength');
		const strengthOptions = ConsumablesInputs.makeStrengthConsumeInput(relevantStatOptions(ConsumablesInputs.STRENGTH_CONSUMES_CONFIG, this.simUI));
		this.updateRow(strengthRow, [buildIconInput(strengthElem, player, strengthOptions)]);

		const [agilityRow, agilityElem] = this.buildGroup('Agility');
		const agilityOptions = ConsumablesInputs.makeAgilityConsumeInput(relevantStatOptions(ConsumablesInputs.AGILITY_CONSUMES_CONFIG, this.simUI));
		this.updateRow(agilityRow, [buildIconInput(agilityElem, player, agilityOptions)]);

		const [staminaRow, staminaElem] = this.buildGroup('Stamina');
		if (raidBuffShown(BuffDebuffInputs.StaminaBuff)) {
			this.scrollOfStamina = buildIconInput(staminaElem, player, ConsumablesInputs.ScrollOfStamina) as IconPicker<Player<Spec>, boolean>;
		}
		this.updateRow(staminaRow, this.scrollOfStamina ? [this.scrollOfStamina] : []);

		const [intellectRow, intellectElem] = this.buildGroup('Intellect');
		const intellectOptions = ConsumablesInputs.makeIntellectConsumeInput(relevantStatOptions(ConsumablesInputs.INTELLECT_CONFIG, this.simUI));
		this.intellectElixir = buildIconInput(intellectElem, player, intellectOptions) as IconEnumPicker<Player<Spec>, number>;
		if (raidBuffShown(BuffDebuffInputs.IntellectBuff)) {
			this.scrollOfIntellect = buildIconInput(intellectElem, player, ConsumablesInputs.ScrollOfIntellect) as IconPicker<Player<Spec>, boolean>;
		}
		this.updateRow(intellectRow, this.scrollOfIntellect ? [this.intellectElixir, this.scrollOfIntellect] : [this.intellectElixir]);

		const [spiritRow, spiritElem] = this.buildGroup('Spirit');
		if (raidBuffShown(BuffDebuffInputs.SpiritBuff)) {
			this.scrollOfSpirit = buildIconInput(spiritElem, player, ConsumablesInputs.ScrollOfSpirit) as IconPicker<Player<Spec>, boolean>;
		}
		this.updateRow(spiritRow, this.scrollOfSpirit ? [this.scrollOfSpirit] : []);

		const [armorRow, armorElem] = this.buildGroup('Armor');
		const armorOptions = ConsumablesInputs.makeArmorConsumeInput(relevantStatOptions(ConsumablesInputs.ARMOR_CONSUMES_CONFIG, this.simUI));
		this.updateRow(armorRow, [buildIconInput(armorElem, player, armorOptions)]);
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
		const hitConsumableOptions = ConsumablesInputs.makeHitConsumableInput(relevantStatOptions(ConsumablesInputs.HIT_CONSUMABLE_CONFIG, this.simUI), 'Hit');

		const pickers = [
			buildIconInput(physicalConsumesElem, this.simUI.player, apBuffOptions),
			buildIconInput(physicalConsumesElem, this.simUI.player, hitConsumableOptions),
		];

		this.updateRow(row, pickers);
	}

	private buildDefensiveBuffPickers() {
		const [row, defensiveConsumesElem] = this.buildGroup('Defensive');

		const healthBuffOptions = ConsumablesInputs.makeHealthConsumeInput(relevantStatOptions(ConsumablesInputs.HEALTH_CONSUMES_CONFIG, this.simUI));

		const pickers = [buildIconInput(defensiveConsumesElem, this.simUI.player, healthBuffOptions)];

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

		const pickers = [
			buildIconInput(spellsCnsumesElem, this.simUI.player, spBuffOptions),
			buildIconInput(spellsCnsumesElem, this.simUI.player, fireBuffOptions),
			buildIconInput(spellsCnsumesElem, this.simUI.player, frostBuffOptions),
			buildIconInput(spellsCnsumesElem, this.simUI.player, shadowBuffOptions),
			buildIconInput(spellsCnsumesElem, this.simUI.player, mp5BuffOptions),
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
			buildIconInput(miscConsumesElem, this.simUI.player, ConsumablesInputs.DragonBreathChili),
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

	// We mark a pick in red when the sim ignores it, because a bigger buff of the same stat is on
	// and they don't stack. Its tooltip names that buff. Keep the rules in sync with the raid
	// buffs in sim/core/buffs.go and applySpellBuffConsumes in sim/core/consumes.go.
	private markUnstacked() {
		const marks = new Map<IconPicker<Player<Spec>, any> | IconEnumPicker<Player<Spec>, any>, string>();
		const badges = new Map<HTMLElement, { elem: HTMLElement; tooltip: Instance }>();

		const update = () => {
			const player = this.simUI.player;
			const buffs = player.getRaid()!.getBuffs();
			const consumes = player.getConsumes();
			marks.clear();

			const replaces = (bigger: string) => `Doesn't stack with ${bigger}, which is bigger, so this adds nothing.`;
			if (this.scrollOfStamina && buffs.scrollOfStamina && buffs.powerWordFortitude != TristateEffect.TristateEffectMissing) {
				marks.set(this.scrollOfStamina, replaces('Power Word: Fortitude'));
			}
			if (this.scrollOfSpirit && buffs.scrollOfSpirit && buffs.divineSpirit) {
				marks.set(this.scrollOfSpirit, replaces('Divine Spirit'));
			}
			if (this.scrollOfIntellect && buffs.scrollOfIntellect) {
				if (buffs.arcaneBrilliance) {
					marks.set(this.scrollOfIntellect, replaces('Arcane Intellect'));
				} else if (this.intellectElixir && consumes.intellectElixir != IntellectElixir.IntellectElixirUnknown) {
					// The sim keeps the bigger of the two. A Scroll of Intellect I gives 4, under
					// the elixir's 6, and the later ranks give more.
					const scroll = ConsumablesInputs.scrollOfIntellectValue(player.getEffectiveLevel());
					const elixir = ConsumablesInputs.INTELLECT_ELIXIR_VALUE;
					if (scroll >= elixir) {
						marks.set(this.intellectElixir, replaces(`the Scroll of Intellect (${scroll} Intellect)`));
					} else {
						marks.set(this.scrollOfIntellect, replaces(`the Intellect elixir (${elixir} Intellect)`));
					}
				}
			}

			[this.scrollOfStamina, this.scrollOfIntellect, this.scrollOfSpirit, this.intellectElixir].forEach(picker => {
				if (!picker) return;
				const note = marks.get(picker);
				picker.rootElem.classList.toggle('consumes-unstacked', !!note);
				// The badge has the note in its tooltip. The slot keeps the item's tooltip.
				let badge = badges.get(picker.rootElem);
				if (note && !badge) {
					const elem = document.createElement('span');
					elem.classList.add('consumes-unstacked-badge');
					elem.textContent = '!';
					picker.rootElem.appendChild(elem);
					const tooltip = tippy(elem, { theme: 'consumes-unstacked' });
					this.addOnDisposeCallback(() => tooltip.destroy());
					badge = { elem, tooltip };
					badges.set(picker.rootElem, badge);
				}
				if (badge) {
					badge.elem.classList.toggle('hide', !note);
					badge.tooltip.setContent(note ?? '');
				}
			});
		};

		update();
		this.simUI.changeEmitter.on(update);
	}

	private updateRow(rowElem: HTMLElement, pickers: (IconPicker<Player<Spec>, any> | IconEnumPicker<Player<Spec>, any> | MultiIconPicker<Player<Spec>>)[]) {
		if (!!pickers.find(p => p?.showWhen())) {
			rowElem.classList.remove('hide');
		} else {
			rowElem.classList.add('hide');
		}
	}
}
