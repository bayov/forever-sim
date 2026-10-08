import tippy from 'tippy.js';

import { MAX_CHARACTER_LEVEL } from '../../constants/mechanics';
import { Player } from '../../player';
import { Profession, Race, Spec } from '../../proto/common';
import { professionNames, raceNames } from '../../proto_utils/names';
import { specToEligibleRaces } from '../../proto_utils/utils';
import { TypedEvent } from '../../typed_event';
import { EnumPickerConfig, EnumValueConfig } from '../enum_picker';
import { Input, InputConfig } from '../input';

const iconUrl = (icon: string) => `https://wow.zamimg.com/images/wow/icons/large/${icon}.jpg`;

export interface LevelSliderConfig<ModObject> extends InputConfig<ModObject, number> {
	id: string;
	// The highest level, like 60 for a player or 63 for a raid boss.
	max: number;
}

// A level as a slider from 1 to the highest level, with a button on each side to step it by one
// and a field to type it in. The player's level and the targets' levels use it.
//
// A level of 0 means the highest level. The slider moves the field while we drag it, but we only
// set the level when we let go, because each change of level runs a new sim (and for the player,
// reloads the gear picker).
export class LevelSlider<ModObject> extends Input<ModObject, number> {
	private readonly max: number;
	private readonly slider: HTMLInputElement;
	private readonly field: HTMLInputElement;
	private readonly down: HTMLButtonElement;
	private readonly up: HTMLButtonElement;

	constructor(parent: HTMLElement, modObject: ModObject, config: LevelSliderConfig<ModObject>) {
		super(parent, 'level-picker-root', modObject, config);
		this.max = config.max;

		this.down = (
			<button type="button" className="level-picker-step" aria-label="One level down">
				<i className="fas fa-minus"></i>
			</button>
		) as HTMLButtonElement;
		this.slider = (<input type="range" className="level-picker-slider" min="1" max={String(this.max)} step="1" />) as HTMLInputElement;
		this.up = (
			<button type="button" className="level-picker-step" aria-label="One level up">
				<i className="fas fa-plus"></i>
			</button>
		) as HTMLButtonElement;
		this.field = (<input type="text" id={config.id} inputMode="numeric" className="form-control level-picker-field" />) as HTMLInputElement;

		this.rootElem.appendChild(
			<div className="level-picker-controls">
				{this.down}
				{this.slider}
				{this.up}
				{this.field}
			</div>,
		);

		this.slider.addEventListener('input', () => this.show(this.slider.valueAsNumber), { signal: this.signal });
		this.slider.addEventListener('change', () => this.commit(this.slider.valueAsNumber), { signal: this.signal });
		this.field.addEventListener('change', () => this.commit(parseInt(this.field.value)), { signal: this.signal });
		this.down.addEventListener('click', () => this.commit(this.getInputValue() - 1), { signal: this.signal });
		this.up.addEventListener('click', () => this.commit(this.getInputValue() + 1), { signal: this.signal });

		this.init();
	}

	getInputElem(): HTMLElement {
		return this.slider;
	}

	getInputValue(): number {
		return this.slider.valueAsNumber;
	}

	setInputValue(newValue: number) {
		this.show(newValue > 0 && newValue < this.max ? newValue : this.max);
	}

	// A typed level outside 1 to the highest level goes to the nearest end, and anything that
	// isn't a number puts back the level we had.
	private commit(level: number) {
		if (isNaN(level)) {
			this.setInputValue(this.getSourceValue());
			return;
		}
		this.show(Math.min(this.max, Math.max(1, Math.round(level))));
		this.inputChanged(TypedEvent.nextEventID());
	}

	private show(level: number) {
		this.slider.value = String(level);
		this.field.value = String(level);
		this.down.disabled = level <= 1;
		this.up.disabled = level >= this.max;
		// The part of the track left of the thumb is filled, see .level-picker-slider.
		this.slider.style.setProperty('--fill', `${((level - 1) / (this.max - 1)) * 100}%`);
	}
}

// The player's level, from 1 to 60.
export class LevelPicker extends LevelSlider<Player<Spec>> {
	constructor(parent: HTMLElement, player: Player<Spec>) {
		super(parent, player, {
			id: 'level',
			label: 'Level',
			labelTooltip:
				'Character level. Base stats, ability ranks, the crit each point of Agility gives, the talent points and the items offered in the gear picker follow it.',
			max: MAX_CHARACTER_LEVEL,
			changedEvent: player => player.miscOptionsChangeEmitter,
			getValue: player => player.getLevel(),
			setValue: (eventID, player, newValue) => player.setLevel(eventID, newValue),
		});
	}
}

// The Skyborne are new in Forever and wowhead has no race icon for them, so they get the
// icon of their Walk on Air racial.
const raceIcons: Record<Race, string> = {
	[Race.RaceUnknown]: 'inv_misc_questionmark',
	[Race.RaceDwarf]: 'race_dwarf_male',
	[Race.RaceGnome]: 'race_gnome_male',
	[Race.RaceHuman]: 'race_human_male',
	[Race.RaceNightElf]: 'race_nightelf_male',
	[Race.RaceOrc]: 'race_orc_male',
	[Race.RaceTauren]: 'race_tauren_male',
	[Race.RaceTroll]: 'race_troll_male',
	[Race.RaceUndead]: 'race_scourge_male',
	[Race.RaceSkyborneHighOrder]: 'inv_elemental_primal_air',
	[Race.RaceSkyborneWindshaper]: 'inv_elemental_primal_air',
};

// An enum input as a row of icons, for the values that have an icon. The selected one is lit
// and the others are dimmed, and the label says which one it is.
//
// A value without an icon, like None for the shaman weapon imbue, gets no icon of its own. We
// pick it by clicking the selected icon again.
export class IconEnumRowPicker<ModObject> extends Input<ModObject, number> {
	private readonly values: Array<EnumValueConfig>;
	private readonly label: HTMLElement;
	private readonly options = new Map<number, HTMLElement>();
	private value = 0;

	constructor(parent: HTMLElement, modObject: ModObject, config: EnumPickerConfig<ModObject>, cssClass = 'icon-enum-row-picker-root') {
		super(parent, cssClass, modObject, config);
		this.rootElem.classList.add('player-icon-picker', 'icon-enum-row-picker');
		this.values = config.values;

		this.label = (<span className="player-icon-picker-value"></span>) as HTMLElement;
		const options = (<div className="player-icon-picker-options"></div>) as HTMLElement;
		this.rootElem.append(this.label, options);

		const none = config.values.find(value => !value.icon);
		config.values.forEach(value => {
			if (!value.icon) return;
			const option = iconOption(value.icon, value.name);
			option.addEventListener(
				'click',
				event => {
					event.preventDefault();
					const newValue = value.value !== this.value ? value.value : none?.value;
					if (newValue === undefined) return;
					this.setInputValue(newValue);
					this.inputChanged(TypedEvent.nextEventID());
				},
				{ signal: this.signal },
			);
			const plain = value.tooltip ? `${value.name}: ${value.tooltip}` : value.name;
			const tooltip = value.richTooltip
				? tippy(option, {
						theme: 'game-spell',
						content: plain,
						onShow: instance => instance.setContent(value.richTooltip!(this.modObject) ?? plain),
				  })
				: tippy(option, { content: plain });
			this.addOnDisposeCallback(() => tooltip.destroy());
			this.options.set(value.value, option);
			options.appendChild(option);
		});

		this.init();
	}

	getInputElem(): HTMLElement {
		return this.rootElem;
	}

	getInputValue(): number {
		return this.value;
	}

	setInputValue(newValue: number) {
		this.value = newValue;
		this.label.textContent = this.nameOf(newValue);
		this.options.forEach((option, value) => option.classList.toggle('active', value === newValue));
	}

	protected formatPresetValue(value: number): string {
		return this.nameOf(value);
	}

	private nameOf(value: number): string {
		return this.values.find(v => v.value === value)?.name ?? '';
	}
}

// The races our class can be.
export class RacePicker extends IconEnumRowPicker<Player<Spec>> {
	constructor(parent: HTMLElement, player: Player<Spec>) {
		super(
			parent,
			player,
			{
				id: 'player-race',
				label: 'Race',
				values: specToEligibleRaces[player.spec].map(race => ({ name: raceNames.get(race)!, value: race, icon: raceIcons[race] })),
				changedEvent: player => player.raceChangeEmitter,
				getValue: player => player.getRace(),
				setValue: (eventID, player, newValue) => player.setRace(eventID, newValue),
			},
			'race-picker-root',
		);
	}
}

const professionIcons: Record<Profession, string> = {
	[Profession.ProfessionUnknown]: 'inv_misc_questionmark',
	[Profession.Alchemy]: 'trade_alchemy',
	[Profession.Blacksmithing]: 'trade_blacksmithing',
	[Profession.Enchanting]: 'trade_engraving',
	[Profession.Engineering]: 'trade_engineering',
	[Profession.Herbalism]: 'trade_herbalism',
	[Profession.Leatherworking]: 'trade_leatherworking',
	[Profession.Mining]: 'trade_mining',
	[Profession.Skinning]: 'inv_misc_pelt_wolf_01',
	[Profession.Tailoring]: 'trade_tailoring',
};

// Engineering comes first, then the other crafting professions, then the gathering ones.
const professionOrder: Array<Profession> = [
	Profession.Engineering,
	Profession.Alchemy,
	Profession.Enchanting,
	Profession.Blacksmithing,
	Profession.Leatherworking,
	Profession.Tailoring,
	Profession.Herbalism,
	Profession.Mining,
	Profession.Skinning,
];

const MAX_PROFESSIONS = 2;

// The professions as icons, where we pick up to two.
//
// When we already have two, the others are dimmed and a click on them does nothing. To switch
// one, we first click one of our two to drop it. The value is sorted, so picking the same two
// in another order doesn't count as a change from the preset.
export class ProfessionsPicker extends Input<Player<Spec>, Array<Profession>> {
	private readonly value: HTMLElement;
	private readonly options = new Map<Profession, HTMLElement>();
	private professions: Array<Profession> = [];

	constructor(parent: HTMLElement, player: Player<Spec>) {
		super(parent, 'professions-picker-root', player, {
			label: 'Professions',
			changedEvent: player => player.professionChangeEmitter,
			getValue: player => [...player.getProfessions()].sort((a, b) => a - b),
			setValue: (eventID, player, newValue) => player.setProfessions(eventID, newValue),
		});
		this.rootElem.classList.add('player-icon-picker');

		this.value = (<span className="player-icon-picker-value"></span>) as HTMLElement;
		const options = (<div className="player-icon-picker-options"></div>) as HTMLElement;
		this.rootElem.append(this.value, options);

		professionOrder.forEach(profession => {
			const option = iconOption(professionIcons[profession], professionNames.get(profession)!);
			option.addEventListener('click', () => this.toggle(profession), { signal: this.signal });
			const tooltip = tippy(option, {
				content: professionNames.get(profession),
				onShow: instance => {
					const full = this.professions.length >= MAX_PROFESSIONS && !this.professions.includes(profession);
					instance.setContent(
						full ? `${professionNames.get(profession)} (drop one of your two professions first)` : professionNames.get(profession)!,
					);
				},
			});
			this.addOnDisposeCallback(() => tooltip.destroy());
			this.options.set(profession, option);
			options.appendChild(option);
		});

		this.init();
	}

	getInputElem(): HTMLElement {
		return this.rootElem;
	}

	getInputValue(): Array<Profession> {
		return [...this.professions];
	}

	setInputValue(newValue: Array<Profession>) {
		this.professions = [...newValue].sort((a, b) => a - b);
		this.value.textContent =
			[...this.professions]
				.sort((a, b) => professionOrder.indexOf(a) - professionOrder.indexOf(b))
				.map(profession => professionNames.get(profession))
				.join(', ') || 'None';
		const full = this.professions.length >= MAX_PROFESSIONS;
		this.options.forEach((option, profession) => {
			const active = this.professions.includes(profession);
			option.classList.toggle('active', active);
			option.classList.toggle('locked', full && !active);
		});
	}

	protected formatPresetValue(value: Array<Profession>): string {
		return value.map(profession => professionNames.get(profession)).join(', ') || 'none';
	}

	private toggle(profession: Profession) {
		if (this.professions.includes(profession)) {
			this.setInputValue(this.professions.filter(p => p !== profession));
		} else if (this.professions.length < MAX_PROFESSIONS) {
			this.setInputValue([...this.professions, profession]);
		} else {
			return;
		}
		this.inputChanged(TypedEvent.nextEventID());
	}
}

function iconOption(icon: string, name: string): HTMLElement {
	return (
		<button type="button" className="player-icon-picker-option" aria-label={name} style={{ backgroundImage: `url('${iconUrl(icon)}')` }}></button>
	) as HTMLElement;
}
