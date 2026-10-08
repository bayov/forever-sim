import tippy, { Instance } from 'tippy.js';

import { IndividualSimUI } from '../../individual_sim_ui';
import { Player } from '../../player';
import { Spec, Stat, TristateEffect } from '../../proto/common';
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

	private staminaSlot?: IconEnumPicker<Player<Spec>, number>;
	private intellectSlot?: IconEnumPicker<Player<Spec>, number>;
	private spiritSlot?: IconEnumPicker<Player<Spec>, number>;

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

		const pickers = [buildIconInput(foodsElem, this.simUI.player, foodOptions)];

		this.updateRow(row, pickers);
	}

	// The attributes' elixirs and scrolls in one category, with Stamina last. Each
	// attribute has one slot with its elixirs and its scroll, and its short name at the bottom. A
	// scroll doesn't stack with a raid buff of the same attribute either, see markUnstacked.
	private buildAttributePickers() {
		const player = this.simUI.player;
		const raidBuffShown = (config: unknown) => relevantStatOptions(BuffDebuffInputs.RAID_BUFFS_CONFIG, this.simUI).some(option => option.config == config);
		const [row, elem] = this.buildGroup('Attributes');
		const pickers: Array<IconPicker<Player<Spec>, any> | IconEnumPicker<Player<Spec>, any>> = [];
		const add = <T extends IconPicker<Player<Spec>, any> | IconEnumPicker<Player<Spec>, any>>(picker: T, tag: string): T => {
			pickers.push(this.tagSlot(picker, tag));
			return picker;
		};

		const strengthOptions = ConsumablesInputs.makeStrengthConsumeInput(relevantStatOptions(ConsumablesInputs.STRENGTH_CONSUMES_CONFIG, this.simUI));
		const agilityOptions = ConsumablesInputs.makeAgilityConsumeInput(relevantStatOptions(ConsumablesInputs.AGILITY_CONSUMES_CONFIG, this.simUI));
		const slot = (config: ReturnType<typeof ConsumablesInputs.makeScrollSlotInput>, tag: string) =>
			add(buildIconInput(elem, player, config) as IconEnumPicker<Player<Spec>, number>, tag);

		add(buildIconInput(elem, player, strengthOptions) as IconEnumPicker<Player<Spec>, number>, 'STR');
		add(buildIconInput(elem, player, agilityOptions) as IconEnumPicker<Player<Spec>, number>, 'AGI');
		this.intellectSlot = slot(
			ConsumablesInputs.makeScrollSlotInput({
				scrollId: 10308,
				scrollField: 'scrollOfIntellect',
				showScroll: raidBuffShown(BuffDebuffInputs.IntellectBuff),
				elixirs: relevantStatOptions(ConsumablesInputs.INTELLECT_CONFIG, this.simUI),
			}),
			'INT',
		);
		this.spiritSlot = slot(
			ConsumablesInputs.makeScrollSlotInput({ scrollId: 10306, scrollField: 'scrollOfSpirit', showScroll: raidBuffShown(BuffDebuffInputs.SpiritBuff) }),
			'SPI',
		);
		this.staminaSlot = slot(
			ConsumablesInputs.makeScrollSlotInput({ scrollId: 10307, scrollField: 'scrollOfStamina', showScroll: raidBuffShown(BuffDebuffInputs.StaminaBuff) }),
			'STA',
		);

		this.updateRow(row, pickers);
	}

	private buildPhysicalBuffPickers() {
		const includeAgi = this.simUI.individualConfig.epStats.includes(Stat.StatAgility);
		const includeStr = this.simUI.individualConfig.epStats.includes(Stat.StatStrength);
		const includeHit = this.simUI.individualConfig.epStats.includes(Stat.StatMeleeHit);

		if (!includeAgi && !includeStr && !includeHit) return;

		const [row, physicalConsumesElem] = this.buildGroup('Physical');

		const apBuffOptions = ConsumablesInputs.makeAttackPowerConsumeInput(relevantStatOptions(ConsumablesInputs.ATTACK_POWER_CONSUMES_CONFIG, this.simUI));
		const hitConsumableOptions = ConsumablesInputs.makeHitConsumableInput(relevantStatOptions(ConsumablesInputs.HIT_CONSUMABLE_CONFIG, this.simUI));

		const pickers = [
			this.tagSlot(buildIconInput(physicalConsumesElem, this.simUI.player, apBuffOptions), 'AP'),
			this.tagSlot(buildIconInput(physicalConsumesElem, this.simUI.player, hitConsumableOptions), 'HIT'),
		];

		this.updateRow(row, pickers);
	}

	private buildDefensiveBuffPickers() {
		const [row, defensiveConsumesElem] = this.buildGroup('Defensive');

		const healthBuffOptions = ConsumablesInputs.makeHealthConsumeInput(relevantStatOptions(ConsumablesInputs.HEALTH_CONSUMES_CONFIG, this.simUI));
		const armorBuffOptions = ConsumablesInputs.makeArmorConsumeInput(relevantStatOptions(ConsumablesInputs.ARMOR_CONSUMES_CONFIG, this.simUI));
		const trollsBloodOptions = ConsumablesInputs.makeTrollsBloodInput(relevantStatOptions(ConsumablesInputs.TROLLS_BLOOD_CONFIG, this.simUI));

		const pickers = [
			this.tagSlot(buildIconInput(defensiveConsumesElem, this.simUI.player, healthBuffOptions), 'HP'),
			this.tagSlot(buildIconInput(defensiveConsumesElem, this.simUI.player, armorBuffOptions), 'ARMOR'),
			this.tagSlot(buildIconInput(defensiveConsumesElem, this.simUI.player, trollsBloodOptions), 'HP5'),
		];
		this.markNotSimulated(pickers[2], "Not simulated: The sim has no health regen yet, so Troll's Blood doesn't affect results.");

		this.updateRow(row, pickers);
	}

	private buildSpellPowerBuffPickers() {
		const [row, spellsCnsumesElem] = this.buildGroup('Spell');

		const spBuffOptions = ConsumablesInputs.makeSpellPowerConsumeInput(relevantStatOptions(ConsumablesInputs.SPELL_POWER_CONFIG, this.simUI));
		const fireBuffOptions = ConsumablesInputs.makeFirePowerConsumeInput(relevantStatOptions(ConsumablesInputs.FIRE_POWER_CONFIG, this.simUI));
		const frostBuffOptions = ConsumablesInputs.makeFrostPowerConsumeInput(relevantStatOptions(ConsumablesInputs.FROST_POWER_CONFIG, this.simUI));
		const shadowBuffOptions = ConsumablesInputs.makeShadowPowerConsumeInput(relevantStatOptions(ConsumablesInputs.SHADOW_POWER_CONFIG, this.simUI));
		const mp5BuffOptions = ConsumablesInputs.makeMp5ConsumeInput(relevantStatOptions(ConsumablesInputs.MP5_CONFIG, this.simUI));

		const pickers = [
			this.tagSlot(buildIconInput(spellsCnsumesElem, this.simUI.player, spBuffOptions), 'SP'),
			this.tagSlot(buildIconInput(spellsCnsumesElem, this.simUI.player, fireBuffOptions), 'FIRE'),
			this.tagSlot(buildIconInput(spellsCnsumesElem, this.simUI.player, frostBuffOptions), 'FROST'),
			this.tagSlot(buildIconInput(spellsCnsumesElem, this.simUI.player, shadowBuffOptions), 'SHAD'),
			this.tagSlot(buildIconInput(spellsCnsumesElem, this.simUI.player, mp5BuffOptions), 'MP5'),
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
			buildIconInput(
				miscConsumesElem,
				this.simUI.player,
				ConsumablesInputs.makeAlcoholInput(relevantStatOptions(ConsumablesInputs.ALCOHOL_CONFIG, this.simUI), 'Alcohol'),
			),
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

	// A slot's short name along its bottom, like STR or AP, so we can tell the slots of a category
	// apart without hovering them.
	private tagSlot<T extends { rootElem: HTMLElement }>(picker: T, tag: string): T {
		const tagElem = document.createElement('span');
		tagElem.classList.add('consumes-slot-tag');
		tagElem.textContent = tag;
		picker.rootElem.appendChild(tagElem);
		return picker;
	}

	// A slot the sim doesn't read yet, like the talents we don't simulate. It gets a dashed edge
	// and a yellow "!" whose tooltip says so. The slot keeps the item's tooltip.
	private markNotSimulated<T extends { rootElem: HTMLElement }>(picker: T, note: string) {
		picker.rootElem.classList.add('consumes-not-simulated');
		const badge = document.createElement('span');
		badge.classList.add('consumes-slot-badge', 'consumes-not-simulated-badge');
		badge.textContent = '!';
		picker.rootElem.appendChild(badge);
		const tooltip = tippy(badge, { content: note, theme: 'consumes-not-simulated' });
		this.addOnDisposeCallback(() => tooltip.destroy());
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

	// We mark a scroll in red when the sim ignores it, because the raid buff of the same stat is on
	// and is bigger. The badge's tooltip names that buff. Keep the rules in sync with the raid
	// buffs in sim/core/buffs.go.
	private markUnstacked() {
		const marks = new Map<IconEnumPicker<Player<Spec>, number>, string>();
		const badges = new Map<HTMLElement, { elem: HTMLElement; tooltip: Instance }>();

		const update = () => {
			const player = this.simUI.player;
			const buffs = player.getRaid()!.getBuffs();
			marks.clear();

			const replaces = (bigger: string) => `Doesn't stack with ${bigger}, which is bigger, so this adds nothing.`;
			if (this.staminaSlot && buffs.scrollOfStamina && buffs.powerWordFortitude != TristateEffect.TristateEffectMissing) {
				marks.set(this.staminaSlot, replaces('Power Word: Fortitude'));
			}
			if (this.intellectSlot && buffs.scrollOfIntellect && buffs.arcaneBrilliance) {
				marks.set(this.intellectSlot, replaces('Arcane Intellect'));
			}
			if (this.spiritSlot && buffs.scrollOfSpirit && buffs.divineSpirit) {
				marks.set(this.spiritSlot, replaces('Divine Spirit'));
			}

			[this.staminaSlot, this.intellectSlot, this.spiritSlot].forEach(picker => {
				if (!picker) return;
				const note = marks.get(picker);
				picker.rootElem.classList.toggle('consumes-unstacked', !!note);
				// The badge has the note in its tooltip. The slot keeps the item's tooltip.
				let badge = badges.get(picker.rootElem);
				if (note && !badge) {
					const elem = document.createElement('span');
					elem.classList.add('consumes-slot-badge', 'consumes-unstacked-badge');
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
